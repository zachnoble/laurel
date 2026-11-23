# Infrastructure

Terraform for AWS. Nothing here lists projects: every stack reads
`projects.json`, which `just gen` writes from the projects' `project.yaml`.

```
shared/                     once per account: ECR repos, GitHub OIDC provider, image build role
envs/{staging,production}/  one per environment: modules/environment with its variables
modules/environment/        cluster, Service Connect namespace, load balancer (gateway only),
                            security groups, deploy role, and every project
modules/go-service/         ECS service + task definition, IAM, logs, secrets, autoscaling
modules/svelte-app/         S3 + CloudFront (SPA fallback), DNS
```

## First time

1. Configure a state backend in each `main.tf` (the `backend "s3"` block).
2. `terraform -chdir=infra/shared apply -var region=... -var github_repository=owner/name`.
   Set the repository variables `AWS_REGION` and `AWS_BUILD_ROLE_ARN` (output `build_role_arn`).
3. For each environment: copy `terraform.tfvars.example` to `terraform.tfvars`, fill
   it in, and `terraform -chdir=infra/envs/staging apply`. In the GitHub environment of
   the same name, set the variables `AWS_REGION` and `AWS_DEPLOY_ROLE_ARN` (output
   `deploy_role_arn`). On `production`, add required reviewers.
4. Fill in the secrets listed by the output `secrets_to_set` (see below).
5. Run the Deploy workflow with **force** to deploy everything once. Until then
   services run a placeholder image and won't start.

## Adding a project

After the new project's `project.yaml` merges, `terraform apply` in `shared` (its ECR repository)
and in each environment (its service or app). Deploys pick it up from there.

## Secrets

Secrets are declared in `project.yaml` (`secrets:`), plus `DATABASE_URL` for
`database: true`. Terraform creates each as an empty Secrets Manager secret,
`laurel/<env>/<service>/<NAME>`, and wires it into the task. Values never
pass through Terraform state; set them directly:

```bash
aws secretsmanager put-secret-value --secret-id laurel/staging/auth/COOKIE_SECRET --secret-string '...'
```


## Databases

Each `database: true` service gets its own database. Creating the Postgres
server, databases and users is outside this Terraform; put each service's
connection string in its `DATABASE_URL` secret. Connections add up
(`MaxConns` × tasks × services, see `libraries/go/db`), so with many services
put RDS Proxy (or PgBouncer) in front of the server and point `DATABASE_URL`
at the proxy.

## How deploys use this

`.github/scripts/deploy-*.sh` read SSM parameters under `/laurel/<env>/`:
`cluster`, `network` (for migration tasks), `api_url`, `apps/<name>/{bucket,distribution}`,
and write `services/<name>/image`, the running image. Terraform registers task
definitions with that image, so an apply never rolls a service back.
