# Go Starter Kit (DDD)

Production-oriented starter for the **ws** backend in a multi-repo setup:

| Repo | Role |
|---|---|
| **ws** (this repo) | Go API, domain logic, migrations, seeders |
| **core** | Operations/admin frontend |
| **site** | Public marketing frontend |

Frontends live in separate repos and consume this API over HTTP. See [`api/openapi.yaml`](api/openapi.yaml).

---

## Why this starter exists

This is not another folder-layout reference repo. It is a **working backend template** built for teams — especially developers arriving from **Laravel** — who need a **multi-repository** setup (`ws`, `core`, `site`) without dragging PHP-era habits into Go.

Laravel gives you conventions out of the box: `artisan`, Eloquent, migrations, controllers, service providers. Go gives you something different: a small language, a large standard library, and the freedom to structure code yourself. That freedom is powerful, but it is also where teams get lost. They either copy Java-style layering, reach for the first framework that feels familiar, or end up with a `handlers/` folder that becomes unmaintainable after the tenth domain.

This starter answers a specific question: **how do you get Laravel-like clarity (migrations, seeders, clear boundaries) while writing idiomatic Go?**

The answer here is **Domain-Driven Design (DDD) at the tactical level** — aggregates, value objects, repository interfaces in the domain, use cases in application, Postgres implementations in infrastructure — combined with **Go community layout conventions** (`cmd/`, `internal/`, `test/`, `scripts/`, thin `main.go`). Not Clean Architecture ceremony for its own sake. Not an enterprise framework. Just enough structure that a new teammate can open `internal/domain/product/` and understand where business rules live.

DDD is chosen because it maps naturally to how Laravel developers already think — **models with behavior**, policies, domain events, repositories — while respecting Go's preference for **explicit packages over magic**. Your `Product` aggregate enforces invariants in code, not in a global `Service` class. Your repository interface lives next to the aggregate it serves, not in a generic `repositories/` junk drawer. That is DDD adapted to Go, not DDD copied from Java.

---

## Why standard library first — and why we avoid frameworks and ORMs

Go's culture is often summarized in one line from the community: **before you add a dependency, ask what the standard library cannot do that you actually need.**

That idea shows up repeatedly across the ecosystem — from long-running discussions on r/golang, to Venkatesh Thallam's observation that Go's standard packages are performant enough to ship production services, and that the typical advice from experienced Go developers is still to **start with stdlib**. Daniel Valev puts it plainly: frameworks are not banned — but adding them should be a **conscious decision**, not muscle memory on day one. See [References](#references) for the articles that shaped this approach.

This starter follows that philosophy deliberately.

### HTTP without a web framework

For years, the main argument for Gin, Echo, or Chi was routing ergonomics — especially path parameters and method-based routes. **Go 1.22+ changed that.** The standard `net/http.ServeMux` now supports patterns like `GET /api/v1/products/{id}` and `r.PathValue("id")` natively. As Valev notes, for most internal APIs, microservices, and backends serving separate frontends, **that is sufficient**.

This template uses **`net/http` only** — no Gin, no Echo, no Fiber. Middleware is plain `func(http.Handler) http.Handler`. Handlers are plain functions. You can read the entire request path from router to database without learning a framework's conventions. When your team outgrows stdlib routing, you can adopt a router library with a clear reason — not because the template assumed you needed one on clone.

### Data access without an ORM

ORMs like GORM promise less boilerplate. They also bring implicit queries, magic associations, migration drift, and debugging sessions spent reading generated SQL. Thallam's benchmarks in the HackerNoon article show that alternatives can be faster in isolated tests — but he concludes that **speed alone should not drive the decision**. Maintainability, testability, and long-term dependency health matter more.

Go's answer is `database/sql`: explicit, boring, and well understood. You write SQL. You scan rows into structs. You control exactly what hits the database. This starter uses **`database/sql` with the `pgx` driver** — a driver, not an ORM — and **goose** for versioned SQL migrations. That mirrors Laravel's migration files more honestly than an ORM schema dumper ever will: your schema lives in `internal/database/migrations/` as plain SQL, reviewable in pull requests, identical in every environment.

Repository interfaces sit in the **domain**; Postgres code sits in **infrastructure**. That is DDD's persistence boundary — not GORM's `AutoMigrate`.

### What we still add (and why)

Stdlib-first does not mean zero dependencies. It means **each dependency earns its place**:

| Dependency | Role | Why not stdlib? |
|---|---|---|
| `github.com/jackc/pgx/v5` | Postgres driver | `database/sql` needs a driver; pgx is the community standard |
| `github.com/pressly/goose/v3` | SQL migrations | No migration tool in stdlib; SQL-first matches Laravel's workflow |
| `github.com/swaggo/http-swagger` | API docs UI | OpenAPI generation; contract for `core` and `site` teams |
| `github.com/joho/godotenv` | Local `.env` loading | DX for development; production uses real env vars |

No framework wraps your application. No ORM hides your queries. The startup case for relying on Go's standard library — simplicity, readability, composability, smaller attack surface — applies directly here: fewer moving parts, faster builds, less upgrade churn, and code a Laravel migrant can follow without learning GORM tags *and* Go *and* a framework's idioms at the same time.

### The trade-off, stated honestly

Frameworks and ORMs are not wrong. Large teams with complex routing, heavy middleware chains, or rapid CRUD prototyping may reach for Gin or GORM and ship faster on day one. This starter optimizes for a different outcome: **a codebase that is still understandable on day three hundred**, when the original author is on vacation and a bug is in the order-placement flow.

If you cannot explain in one sentence why a third-party package is needed, pause — that is the stdlib-first rule this repo is built on.

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
cp -r go-starter-kit/ ../ws/
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

### 8. Swagger UI (development)

When `APP_ENV=development`, open:

**http://localhost:8080/swagger/index.html**

Interactive API docs are generated from handler comments via [swaggo/swag](https://github.com/swaggo/swag) and served with [swaggo/http-swagger](https://github.com/swaggo/http-swagger).

Regenerate after changing handler annotations:

```bash
make swagger
```

Generated files live in `docs/` (`docs.go`, `swagger.json`, `swagger.yaml`). Commit them so `go build` works without installing the `swag` CLI.

### 9. Run tests

```bash
make test                  # unit tests (no Docker)
make test-integration      # Postgres via Testcontainers (requires Docker)
```

Unit tests run colocated with source (`internal/domain/product/product_test.go`). Integration tests use [Testcontainers for Go](https://golang.testcontainers.org/) under `test/` — see [Testing](#testing).

---

## Production deployment (VPS)

For a **single VPS** workflow — push to `main`, SSH to the server, deploy — use `scripts/`:

```bash
cd /opt/ws
./scripts/deploy.sh    # git pull → build → migrate → restart systemd
```

Or build only:

```bash
./scripts/build.sh     # git pull origin/main → bin/api + bin/migrate
```

| Tool | When |
|---|---|
| `make …` | Local development on your laptop |
| `scripts/build.sh` | Server: pull latest `main` and compile binaries |
| `scripts/deploy.sh` | Server: full release (build + migrate + restart) |
| `deployments/` | Docker image and compose for containerized deploy |

Production secrets live in `/etc/ws/ws.env` on the server — not in the repo `.env`. Full setup (systemd, first deploy, env vars): [`scripts/README.md`](scripts/README.md).

This mirrors [golang-standards/project-layout `/scripts`](https://github.com/golang-standards/project-layout#scripts) and patterns from [Terraform](https://github.com/hashicorp/terraform/tree/main/scripts), [Helm](https://github.com/kubernetes/helm/tree/master/scripts), and [CockroachDB](https://github.com/cockroachdb/cockroach/tree/master/scripts): **Makefile for dev, scripts for ops.**

---

## Day-to-day commands

| Task | Command |
|---|---|
| Start API | `make run-api` |
| Unit tests | `make test` |
| Integration tests (Testcontainers) | `make test-integration` |
| All tests | `make test-all` |
| Build binary (local) | `make build` |
| Deploy on VPS | `./scripts/deploy.sh` |
| Pull + build on VPS | `./scripts/build.sh` |
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
cmd/              api, migrate, seed entrypoints
internal/
  domain/         aggregates, value objects, repository interfaces
  application/    use cases
  infrastructure/postgres/   repository implementations
  transport/http/            handlers, middleware, router
  database/     connection, migrations, seeders
  config/       env loading
api/              OpenAPI contract for core + site (hand-maintained)
configs/          committed config templates (e.g. seeder.yaml)
deployments/      Dockerfile, docker-compose
scripts/          VPS production ops: build.sh, deploy.sh (see scripts/README.md)
test/             Testcontainers integration tests (see test/README.md)
docs/             swag-generated Swagger + architecture notes
```

Full rationale: [`docs/solution.md`](docs/solution.md)

---

## Testing

This starter uses a **three-tier** testing model. Unit tests and integration tests are intentionally separated.

### Unit tests (colocated, no Docker)

Place `*_test.go` files next to the code they test. The starter includes examples in `internal/domain/product/`.

```bash
make test
go test ./internal/domain/product/... -v
```

Domain logic — aggregates, value objects, use cases with mocked repos — belongs here. Fast, no external dependencies.

### Integration tests (`test/` + Testcontainers)

Use [Testcontainers for Go](https://golang.testcontainers.org/) for tests that need **real Postgres**: repository implementations, SQL constraints, goose migrations.

```
test/
├── integration/
│   └── helpers/
│       └── postgres.go    # SetupPostgres() — disposable DB + migrations
└── e2e/                   # HTTP smoke tests (add later)
```

```bash
make test-integration    # go test -race -tags=integration ./test/integration/...
```

**Prerequisites:** Docker running locally. Integration files use `//go:build integration` so they are excluded from `make test`.

The helper `test/integration/helpers/postgres.go` starts `postgres:16-alpine`, runs goose migrations from `internal/database/migrations/`, and returns `*sql.DB`. Write repository tests against that — not against your dev `docker-compose` database.

Full guide: [`test/README.md`](test/README.md)

### Laravel mapping

| Laravel | Go (this starter) |
|---|---|
| `tests/Unit` | colocated `*_test.go` in `internal/` |
| `tests/Feature` (database) | `test/integration/` with Testcontainers |
| `RefreshDatabase` | fresh container + migrations per test |

### CI recommendation

| Job | Command |
|---|---|
| Unit (every push) | `go test -race ./...` |
| Integration (Docker runner) | `go test -race -tags=integration -timeout=5m ./test/integration/...` |

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
| `tests/Unit` | colocated `*_test.go` in `internal/` |
| `tests/Feature` (DB) | `test/integration/` with Testcontainers |
| Forge “Deploy Now” | `./scripts/deploy.sh` on VPS |
| `php artisan migrate --force` | `bin/migrate up` (in deploy.sh) |

---

## Conventions

- **No `pkg/`** unless you publish a Go library other teams import.
- **No frontends in ws** — use **core** and **site** repos.
- **Package by domain**, not by type (`product/`, not `handlers/`).
- **Unit tests colocated**; **integration tests in `test/`** with Testcontainers.
- **`make` for local dev**; **`scripts/` for VPS deploy** after merge to `main`.
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

### Stdlib-first philosophy

- [I Stopped Reaching for Go Frameworks. The Standard Library Was There All Along](https://towardsdev.com/i-stopped-reaching-for-go-frameworks-the-standard-library-was-there-all-along-f4566b33ed12) — Daniel Valev (Towards Dev)
- [Why Startups Should Rely on Go's Standard Library (and Not Third-Party Bloat)](https://medium.com/@kanishks772/why-startups-should-rely-on-gos-standard-library-and-not-third-party-bloat-bad601a0fd92) — The Latency Gambler (Medium)
- [The Myth about Golang Frameworks and External Libraries](https://medium.com/hackernoon/the-myth-about-golang-frameworks-and-external-libraries-93cb4b7da50f) — Venkatesh Thallam (HackerNoon)
- [Why do Go users avoid frameworks?](https://www.reddit.com/r/golang/comments/1gs1cxq/why_do_go_users_avoid_frameworks/) — r/golang discussion

### Project docs

- [docs/solution.md](docs/solution.md) — structure decisions
- [docs/architecture.md](docs/architecture.md) — layer overview
- [scripts/README.md](scripts/README.md) — VPS deploy workflow
- [docs/medium.md](docs/medium.md) — layered intro
- [Organizing a Go module](https://go.dev/doc/modules/layout) — official Go guidance
