# Every recipe works for any project by name; nothing here lists projects.
# Projects are described by their project.yaml (see tools/repo/manifest.go).

repo := "go run ./tools/repo"

default:
    @just --list

# Install hooks, JS dependencies, and each project's .env from its .env.example
setup:
    lefthook install
    pnpm install
    @for example in $(git ls-files '*/.env.example'); do \
        env="${example%.example}"; \
        [ -f "$env" ] || { cp "$example" "$env"; echo "created $env"; }; \
    done

# Regenerate everything: rpc/generated from rpc/proto/, each project's sqlc.yaml output, and compose.projects.yml, infra/projects.json from project.yaml
gen:
    rpc/generators/generate.sh
    {{ repo }} gen

# Fail if anything `just gen` writes is out of date
gen-check:
    {{ repo }} gen --check

# List projects (optionally of one kind: go-service, svelte-app)
list kind="":
    {{ repo }} list {{ if kind == "" { "" } else { "--kind " + kind } }}

# What a change affects, vs a base commit (default: main)
affected base="origin/main":
    {{ repo }} affected --base {{ base }}

# Print a new Ed25519 key pair for identity tokens
keygen:
    go run ./libraries/go/rpc/cmd/identity-keygen

proto-fmt:
    buf format -w

proto-check:
    buf lint
    buf format --diff --exit-code
    buf breaking --against '.git#branch=main'

# Go tests: everything, or packages, e.g. `just test ./services/auth/...`. Integration tests run when the test database is up (`just test-db`).
test *pkgs="./...":
    TEST_DATABASE_URL="${TEST_DATABASE_URL:-postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable}" go test {{ pkgs }}

# Tests for only what changed vs main
test-affected base="origin/main":
    #!/usr/bin/env bash
    set -euo pipefail
    pkgs=$({{ repo }} affected --base {{ base }} | jq -r '.go_packages | join(" ")')
    [ -n "$pkgs" ] && just test $pkgs || echo "no Go packages affected"
    dirs=$({{ repo }} affected --base {{ base }} | jq -r '.ts_packages[]')
    for d in $dirs; do pnpm --dir "$d" run --if-present typecheck; pnpm --dir "$d" run --if-present test; done

# Start the database integration tests use
test-db:
    docker compose up -d postgres

lint:
    test -z "$(gofmt -l $(git ls-files '*.go'))" || { gofmt -l $(git ls-files '*.go'); exit 1; }
    go vet ./...
    pnpm exec biome check

# Local dev runs only what you name, plus everything it depends on (its
# database and the services in `calls`). Requests for services you didn't
# start fail with "not running locally".

# Start projects, e.g. `just up web` or `just up auth orders`
up +targets:
    docker compose up -d gateway {{ targets }}

# Stop some containers, or everything if none given
down *targets:
    {{ if targets == "" { "docker compose down" } else { "docker compose stop " + targets } }}

# Build dev images (all if none given)
build *targets:
    DOCKER_BUILDKIT=1 docker compose build {{ targets }}

rebuild +targets:
    DOCKER_BUILDKIT=1 docker compose up -d --build {{ targets }}

ps:
    docker compose ps

logs *targets:
    docker compose logs -f {{ targets }}

# Apply a service's migrations in its running dev container (they also run on start)
migrate name:
    docker compose exec {{ name }} go run ./services/{{ name }}/cmd/http migrate

# Build a service's production image locally
image name:
    DOCKER_BUILDKIT=1 docker build --build-arg SERVICE={{ name }} -f docker/go-service.Dockerfile -t {{ name }}:local .

[confirm("Are you sure you want to delete all containers, images, and volumes?")]
clean:
    docker compose down -v --rmi all --remove-orphans
    -docker rm -f $(docker ps -aq)
    -docker rmi -f $(docker images -q)
    docker system prune -af --volumes
    @echo "Cleanup completed successfully."
