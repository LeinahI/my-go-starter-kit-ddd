# Architecture

This starter follows **tactical DDD** inside a standard Go application layout.

## Layers

1. **domain** — business rules, aggregates, value objects, repository interfaces. No imports from outer layers.
2. **application** — use cases that orchestrate domain objects.
3. **infrastructure** — Postgres repositories, external adapters.
4. **transport/http** — HTTP handlers, middleware, JSON responses.
5. **database** — SQL migrations (goose) and seeders.

## Example aggregate

See `internal/domain/product/` for:

- unexported fields and behavior methods (`Reserve`)
- `New()` factory with validation
- `Reconstruct()` for loading from DB
- `Repository` interface colocated with the aggregate

## Dependency flow

```
transport/http  →  application  →  domain  ←  infrastructure/postgres
                                        ↑
                                   database (migrations, pool)
```

`domain` imports nothing from outer layers. `cmd/api/main.go` is the composition root that wires concrete types.

## Multi-repo

- **ws** exposes HTTP + `api/openapi.yaml`
- **core** and **site** are separate frontend repos — they do not import Go packages from ws

## Operations layout

Three folders handle three audiences — do not mix them:

| Folder | Runs on | Purpose |
|---|---|---|
| **`Makefile`** | Developer laptop | `make run-api`, `make test`, `make migrate-up` |
| **`scripts/`** | Production VPS (SSH) | `./scripts/build.sh`, `./scripts/deploy.sh` |
| **`deployments/`** | Docker hosts / CI | `Dockerfile`, `docker-compose.yml` |

Typical VPS release after merge to `main`:

```bash
cd /opt/ws && ./scripts/deploy.sh
```

`deploy.sh` orchestrates: git pull → compile `bin/api` + `bin/migrate` → goose migrate → systemd restart. Secrets live in `/etc/ws/ws.env` on the server, not in git.

This matches [golang-standards/project-layout `/scripts`](https://github.com/golang-standards/project-layout#scripts) and patterns from [Terraform](https://github.com/hashicorp/terraform/tree/main/scripts), [Helm](https://github.com/kubernetes/helm/tree/master/scripts), and [CockroachDB](https://github.com/cockroachdb/cockroach/tree/master/scripts).

Details: [scripts/README.md](../scripts/README.md)

## Testing

Three tiers, three locations:

| Tier | Where | Tooling | Example |
|---|---|---|---|
| **Unit** | Colocated `*_test.go` in `internal/` | `make test` | `internal/domain/product/product_test.go` |
| **Integration** | `test/integration/` | [Testcontainers for Go](https://golang.testcontainers.org/) + goose | `helpers.SetupPostgres()` → real Postgres |
| **E2E** | `test/e2e/` | HTTP client against running API | add when handlers need smoke tests |

Integration tests use `//go:build integration` and run separately:

```bash
make test-integration   # requires Docker
```

See [test/README.md](../test/README.md) and [solution.md](solution.md) for full conventions.
