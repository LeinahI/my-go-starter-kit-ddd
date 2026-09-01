# Go Starter Kit (DDD)

Production-oriented starter for the **ws** backend in a multi-repo setup:

| Repo | Role |
|---|---|
| **ws** (this repo) | Go API, domain logic, migrations, seeders |
| **core** | Operations/admin frontend |
| **site** | Public marketing frontend |

Frontends live in separate repos and consume this API over HTTP. See [`api/openapi.yaml`](api/openapi.yaml).

---

## Initialize

Use this when cloning the template for a new project.

### Prerequisites

| Tool | Version | Check |
|---|---|---|
| [Go](https://go.dev/dl/) | 1.27 | `go version` |
| [Docker](https://www.docker.com/) | any recent | `docker --version` |
| Make | optional but recommended | `make --version` |
| Git | any recent | `git --version` |

### 1. Get the template

**Option A — use as a new repo**

```bash
git clone <this-repo-url> ws
cd ws
```

**Option B — copy into an existing folder**

```bash
cp -r project-layout/ ../ws/
cd ../ws
```

### 2. Set your Go module path

Replace the placeholder module name with your real path (e.g. `github.com/yourorg/ws`).

**`go.mod`** — change the first line:

```go
module github.com/yourorg/ws
```

**Imports** — update every file that imports `github.com/yourorg/ws`:

```bash
# Linux / macOS / Git Bash
find . -name '*.go' -exec sed -i 's|github.com/yourorg/ws|github.com/YOUR-ORG/ws|g' {} +

# Or use your editor's find-and-replace across the project
```

Then sync dependencies:

```bash
go mod tidy
```

### 3. Configure environment

```bash
cp .env.example .env
```

Edit `.env` if needed. Defaults match the bundled Docker Postgres:

| Variable | Default | Purpose |
|---|---|---|
| `HTTP_PORT` | `8080` | API listen port |
| `DATABASE_URL` | `postgres://postgres:root@127.0.0.1:5433/ws?sslmode=disable` | Postgres connection |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173` | Allowed frontend origins |

**Using your own Postgres (no Docker)?** Point `DATABASE_URL` at your instance and create an empty database named `ws`.

### 4. Start the database

```bash
make docker-db-up
```

Wait until Postgres is healthy:

```bash
docker compose -f deployments/docker-compose.yml ps
```

Stop later with `make docker-db-down`.

### 5. Run migrations and seeders

```bash
make migrate-up    # apply SQL migrations (Laravel: php artisan migrate)
make seed          # load demo data from configs/seeder.yaml
```

Reset everything and re-seed (Laravel: `migrate:fresh --seed`):

```bash
make migrate-fresh-seed
```

### 6. Start the API

```bash
make run-api
```

Or build a binary:

```bash
make build
./bin/api        # Windows: bin\api.exe
```

### 7. Verify

```bash
curl http://localhost:8080/up
curl http://localhost:8080/api/v1/products
```

Expected: health returns `{"data":{"status":"ok"}}`; products returns the seeded list.

### 9. Swagger UI (development)

When `APP_ENV=development`, open:

**http://localhost:8080/swagger/index.html**

Interactive API docs are generated from handler comments via [swaggo/swag](https://github.com/swaggo/swag) and served with [swaggo/http-swagger](https://github.com/swaggo/http-swagger).

Regenerate after changing handler annotations:

```bash
make swagger
```

Generated files live in `docs/` (`docs.go`, `swagger.json`, `swagger.yaml`). Commit them so `go build` works without installing the `swag` CLI.

### 10. Run tests

```bash
make test
```

---

## Day-to-day commands

| Task | Command |
|---|---|
| Start API | `make run-api` |
| Run tests | `make test` |
| Build binary | `make build` |
| Migrate up | `make migrate-up` |
| Migrate down (one step) | `make migrate-down` |
| Migration status | `make migrate-status` |
| Fresh DB + seed | `make migrate-fresh-seed` |
| Roll back N migrations | `make migrate-rollback step=1` |
| New migration file | `make migrate-create name=add_users_table` |
| Seed only | `make seed` |
| Start / stop Docker DB | `make docker-db-up` / `make docker-db-down` |
| Regenerate Swagger docs | `make swagger` |
| Swagger UI (dev only) | http://localhost:8080/swagger/index.html |

Migration files are created under `internal/database/migrations/`.

---

## Project layout

```
cmd/           api, migrate, seed entrypoints
internal/
  domain/      aggregates, value objects, repository interfaces
  application/ use cases
  infrastructure/postgres/  repository implementations
  transport/http/           handlers, middleware, router
  database/    connection, migrations, seeders
  config/      env loading
api/           OpenAPI contract for core + site (hand-maintained)
configs/       committed config templates (e.g. seeder.yaml)
deployments/   Docker, compose
docs/          swag-generated Swagger + architecture notes
```

Full rationale: [`docs/solution.md`](docs/solution.md)

---

## Laravel mapping

| Laravel | ws |
|---|---|
| `artisan serve` | `make run-api` |
| `artisan migrate` | `make migrate-up` |
| `artisan db:seed` | `make seed` |
| `artisan migrate:fresh --seed` | `make migrate-fresh-seed` |
| `routes/api.php` | `internal/transport/http/router.go` |
| `app/Http/Controllers` | `internal/transport/http/*_handler.go` |
| `app/Models` (rich) | `internal/domain/{aggregate}/` |
| `app/Services` | `internal/application/{aggregate}/` |
| `database/migrations` | `internal/database/migrations/` |

---

## Conventions

- **No `pkg/`** unless you publish a Go library other teams import.
- **No frontends in ws** — use **core** and **site** repos.
- **Package by domain**, not by type (`product/`, not `handlers/`).
- Run commands from the **repo root** (where `go.mod` lives).

---

## Troubleshooting

**`DATABASE_URL is required`**
- Copy `.env.example` to `.env`, or export `DATABASE_URL` before `make migrate-up`.

**`connection refused` on port 5433**
- Run `make docker-db-up` and confirm the container is healthy.

**`migrate up` fails on existing tables**
- Use `make migrate-status` to inspect state, or `make migrate-fresh-seed` on a dev machine only.

**`swagger: command not found`**
- Install the CLI: `go install github.com/swaggo/swag/cmd/swag@latest`, then run `make swagger`.

---

## References

- [docs/solution.md](docs/solution.md) — structure decisions
- [docs/medium.md](docs/medium.md) — layered intro
- [Organizing a Go module](https://go.dev/doc/modules/layout)
