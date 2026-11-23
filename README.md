## Layout

```
services/<name>/          Go services (one ECS service each); project.yaml describes it
apps/<name>/              SvelteKit apps (S3 + CloudFront)
libraries/go/<name>/      shared Go: service (the standard main), server, rpc (Connect runtime: auth, clients), db, telemetry, config, logging
libraries/ts/<name>/      shared TypeScript: @laurel/ui, @laurel/rpc-runtime (Connect transport)
rpc/
  proto/<name>/v1/*.proto        each service's API (package <name>.v1)
  generators/                    codegen config behind `just gen`
  generated/{go,typescript}/     generated from rpc/proto/
tools/repo/               manifests, generators, affected-project detection
infra/                    Terraform: shared (ECR, CI roles), modules, envs/{staging,production}
docker/                   the shared Dockerfiles (one prod image recipe, one dev image per language)
```

Go is a single module (`go.mod` at the root), so every service builds against
the same dependency versions. TypeScript dependency versions live once, in the
`catalog:` of `pnpm-workspace.yaml`.

### Projects

Every service and app has a `project.yaml` (see
`tools/repo/manifest.go` for every field):

```yaml
name: orders
kind: go-service           # go-service | svelte-app
database: true             # own Postgres database, migrations embedded in the binary
calls: [inventory]         # local dev starts these too
secrets: [STRIPE_KEY]      # from Secrets Manager when deployed
deploy: { cpu: 256, memory: 512, count: 2 }
```

`just gen` turns these into `compose.projects.yml` (local dev) and
`infra/projects.json` (Terraform). Nothing else lists projects: CI, local dev
and infra all discover them. CI fails if the generated files are stale.
Ownership is in `.github/CODEOWNERS`, edited by hand.

### Adding a service or app

Copy an existing one: for a service, `services/auth` (with a database) or
`services/gateway` (without), plus its proto under `rpc/proto/<name>/v1/`.
Rename it in `project.yaml`, `main.go` and the proto package (apps also need
a new subdomain and port), then run `just gen` and `just up <name>`. Deploying
it needs one `terraform apply` per environment (see `infra/README.md`) to
create its ECS service; after that, merges deploy it.

## Auth

```
browser --cookie--> gateway --X-Identity: <user JWT> + X-Via-Gateway--> <name>:8080
                       |                                              --X-Identity--> other services
                       '-- AuthInternalService/Authenticate (cached 30s) --> auth
```

- The browser only talks to the **gateway**. It routes `/<name>.v1.<Service>/<Method>`
  to `http://<name>:8080`, so adding a service needs no gateway or load balancer change.
- **Users:** the session is an opaque, revocable cookie owned by **auth**. The
  gateway exchanges it for a 5-minute identity token, caching each session for
  30s in a bounded LRU (so a revoked session stops working within 30s). Only
  auth receives cookies.
- **Services** trust each other by network: only the gateway is reachable from
  outside (security groups), and it marks everything it forwards with
  `X-Via-Gateway`. A request without that header came from another service.
- **Access rules live in the schema** and every service enforces them
  (`rpc.Authenticator`):
  - a method needs a signed-in user unless it has `option (options.v1.public) = true`
  - a service with `option (options.v1.internal) = true` rejects calls marked
    `X-Via-Gateway`, so only other services can reach it.

  Read the user with `rpc.IdentityFrom(ctx)`.
- **Calling another service:** `rpc.NewClient(s.Clients, "orders", ordersv1connect.NewOrdersServiceClient)`.
  It forwards the user's identity, sets a default
  10s deadline (propagated downstream), and traces the call.

Identity keys: `just keygen`. `IDENTITY_PRIVATE_KEY` is auth's secret;
`identity_public_keys` (Terraform) goes to every service. To rotate: add the
new public key, switch auth's private key, wait 5 minutes, drop the old public
key.

## Apps

Apps are served at `<subdomain>.<domain>` and the API at `api.<domain>`. They
must share the domain: the session cookie is `SameSite=Lax`, which browsers
don't send to a different site. An app on another domain needs its own backend
on that domain. The gateway allows exactly the environment's apps as CORS
origins (computed by Terraform).

## A Go service

`libraries/go/service` is the whole main: config, JSON logging with request
and trace IDs, OpenTelemetry tracing and metrics, auth, clients, the
database, health checks (`/health`, `/readiness`), and graceful shutdown.

```go
func main() {
	service.Run(service.Options{Name: "orders", Migrations: db.Migrations}, setup)
}

func setup(ctx context.Context, s *service.Service) error {
	pool, err := s.Database(ctx)
	...
	s.Mux.Handle(ordersv1connect.NewOrdersServiceHandler(h, s.HandlerOptions()...))
	return nil
}
```

**Migrations** are embedded in the binary. Deploys run them as a one-off task
(`server migrate`) before rolling out, while the previous release is still
serving, so every migration must work with the previous code: add first,
remove a release later. CI lints new migrations with squawk.

## Setup

```bash
brew install go node pnpm just lefthook buf
just setup
```

`just setup` installs hooks and dependencies, and creates each project's `.env`
from its `.env.example`. Shared dev settings (dev-only keys, URLs) are in
`dev.env`.

## Dev

Run only what you're working on. `just up` starts the named projects plus the
gateway and everything they depend on (their database, the services in
`calls`), nothing else.

```bash
just up web            # web + gateway + auth + postgres
just up orders         # a service, with its database and dependencies
just down orders       # stop one; `just down` stops everything
just logs orders
```

- One dev image serves every Go service, with hot reload; one Postgres holds a
  database per service, migrated on start.
- Calls to services that aren't running fail with `unavailable: Service "x" is
  unavailable (not running locally? ...)`.
- To use a shared environment for some dependency instead of running it, set
  `SERVICE_URL_<NAME>=https://...` in your project's `.env` (the gateway honors
  it too).
- Frontend only, no Docker: point `apps/<app>/.env`'s `API_URL` at a shared
  gateway and run `pnpm dev` in the app.
- Traces: uncomment `OTEL_EXPORTER_OTLP_ENDPOINT` in `dev.env`, `just up jaeger <targets>`,
  open http://localhost:16686.

```bash
just gen          # regenerate rpc/generated, CODEOWNERS, compose.projects.yml, infra/projects.json
just proto-check  # lint, format check, breaking changes vs main
just affected     # what your change affects (packages, services, apps)
```

## Tests

```bash
just test-db               # Postgres for integration tests
just test                  # all Go tests
just test ./services/auth/...
just test-affected         # only what your change affects, Go and TypeScript
```

Integration tests get a fresh migrated database each from
`libraries/go/db/dbtest`. They're skipped locally without the test database,
and fail in CI if it's missing.

## CI

PR checks run only what the change affects: `tools/repo affected` follows the
Go import graph and the pnpm workspace graph, so a change to one service vets,
tests and builds that service. Dependency changes (`go.mod`, `pnpm-lock.yaml`)
check everything in that language; deploys still skip services and apps whose
image or build didn't change. Matrices are sharded (at most 20 jobs). Always: generated
files are current, and proto lint and breaking-change checks run when `rpc/proto/`
changes.

## Deploy

Merges to `main` deploy to **staging**, then **production** (after approval:
add required reviewers to the `production` GitHub environment).

- Each project is compared against the commit last deployed to each
  environment (`deploy/<env>/<name>` tags). A cancelled, superseded or failed
  run never loses changes; the next run deploys everything not yet deployed.
- Images are built once, reproducibly. A service whose image digest didn't
  change, or an app whose build didn't change, is skipped.
- ECS rolls back rollouts that don't become healthy (deployment circuit breaker).
- **Rollback:** run the Rollback workflow with an environment, project and
  commit.
- AWS access is via OIDC roles only (see `infra/README.md` for the GitHub
  variables).
