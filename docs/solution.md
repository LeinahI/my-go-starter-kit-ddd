# Go Project Structure for Laravel Developers (Multi-Repo + DDD)

A balanced layout recommendation that merges [golang-standards/project-layout](https://github.com/golang-standards/project-layout) with practical layered patterns, adapted for a **multi-repository** product split and **Domain-Driven Design** in idiomatic Go.

---

## Context

| Repository | Role | Example |
|---|---|---|
| **ws** | Backend API (Go) | REST/JSON API, auth, business rules, database |
| **core** | Operations frontend | Admin panel, internal dashboards |
| **site** | Public frontend | Marketing/public site |

Each repo is independent. Only **ws** is Go. **core** and **site** consume ws over HTTP — the same way a Laravel API serves a Vue/React SPA or a separate Next.js site.

This document focuses on **ws** (backend) structure and how this starter template should be used.

---

## Review: Two Sources, Both Sides

### golang-standards/project-layout

**What to adopt**

| Pattern | Why |
|---|---|
| `cmd/` per binary | Official Go convention; thin `main.go` that only wires dependencies |
| `internal/` | Compiler-enforced privacy — Go's primary architectural boundary |
| `configs/` | Config templates separate from runtime loading code |
| `deployments/` | Docker, compose, K8s — keeps infra out of application code |
| `api/` | OpenAPI/JSON schema contracts shared with core and site teams |
| `docs/` | Architecture decisions, onboarding — not godoc |
| `scripts/` | VPS/production ops: pull, build, migrate, restart — keeps Makefile dev-focused |
| `test/` | Integration/E2E via [Testcontainers for Go](https://golang.testcontainers.org/) — **not** unit tests |
| "Clone and delete what you don't need" | Start minimal; grow deliberately |

**What to avoid or defer**

| Pattern | Why |
|---|---|
| `pkg/` by default | Brad Fitzpatrick and much of the community consider it redundant for application repos. Use `internal/` unless you are publishing a library other teams import |
| `vendor/` | Go module proxy makes this optional; only add when air-gapped or reproducibility demands it |
| `tools/`, `third_party/`, `githooks/`, `init/`, `assets/`, `examples/` | Boilerplate folders that clutter a new project. Add when a real need appears |
| `web/` and `website/` inside ws | In a multi-repo setup, frontends live in **core** and **site**. ws should not host SPA assets unless you are doing server-side rendering in Go |
| `/src` at project root | Java-ism; explicitly discouraged by the README and Go tooling |
| Treating it as "the standard" | It is community convention, not an official Go team spec. The real authority is [Organizing a Go module](https://go.dev/doc/modules/layout) |

**Core philosophy alignment:** The README itself warns that this layout is overkill for PoCs and intentionally avoids prescribing internal package structure. That is correct Go thinking — **folders are not architecture; packages are.**

---

### Layered service patterns (medium-style tutorials)

**What to adopt**

| Pattern | Why |
|---|---|
| Layered flow: transport → application → persistence | Familiar to Laravel developers (Controller → Service → Model/Repository) |
| `cmd/api`, `cmd/migrate`, `cmd/seed` | Maps cleanly to `artisan serve`, `artisan migrate`, `artisan db:seed` |
| Makefile targets | `make run-api`, `make migrate-up`, `make seed` — excellent DX for teams new to Go |
| Scale path: package-by-feature (`internal/domain/product/`, `internal/domain/order/`) | Aligns with Go's "package by responsibility" principle |

**What to avoid or defer**

| Pattern | Why |
|---|---|
| Global `internal/handlers/`, `internal/services/`, `internal/store/` | Works for tutorials; becomes a "junk drawer" at 10+ domains |
| `internal/models/` | Anemic structs with JSON tags are not domain models. Collides with DDD where the model has behavior |
| `internal/dto/` as a top-level package | DTOs belong at the transport boundary (HTTP request/response shapes), not as a parallel domain layer |
| `pkg/utils`, `pkg/response`, `pkg/middleware` for everything | Generic `utils` packages become dependency magnets. Prefer small, purpose-named internal packages |
| Full Hexagonal `ports/` + `adapters/` + `bootstrap/` on day one | Correct for large systems; heavy for a team still learning Go. Introduce when a second transport (gRPC, CLI) or second persistence backend appears |
| `.env` committed to repo | Use `.env.example` only; real secrets stay local or in a secret manager |

**Core philosophy alignment:** The best scaling advice is **package by domain** (`product/`, `order/`) — not type-based layout. Lead with that for DDD.

---

## Recommended Structure

### Design principles (Go-first)

1. **Package by responsibility, not by type.** `product/`, `order/` — not `handlers/`, `repositories/`.
2. **`internal/` is the boundary.** Everything application-specific stays inside; nothing in `pkg/` unless another repo imports it as a library (unlikely for ws).
3. **Dependencies point inward.** `domain` → zero imports from application/infrastructure/transport.
4. **Interfaces where they are used.** Repository interfaces live in the domain aggregate package; Postgres implements them in infrastructure.
5. **Thin `cmd/`.** Only `main()` and fatal error handling.
6. **Two-tier testing.** Unit tests colocated; integration tests in `test/` with Testcontainers.
7. **Makefile for dev, scripts for prod.** `make` on laptop; `scripts/deploy.sh` on VPS after merge to `main`.
8. **Start simple, grow layers.** Begin with domain + application + one infrastructure adapter. Add hexagonal folders when a second adapter appears.
9. **No `/src`.** Ever.

### ws — backend repository

```
ws/
├── cmd/
│   ├── api/
│   │   └── main.go                 # bootstrap: config → DB → router → ListenAndServe
│   ├── migrate/
│   │   └── main.go                 # goose: up | down | fresh | rollback | create
│   └── seed/
│       └── main.go                 # runs seeders in order
│
├── internal/
│   ├── domain/                     # ← Laravel "domain" (not Eloquent models)
│   │   ├── product/
│   │   │   ├── product.go          # aggregate root, behavior, invariants
│   │   │   ├── product_test.go     # unit tests (colocated)
│   │   │   └── errors.go           # ErrNotFound, ErrInvalidPrice, etc.
│   │   ├── order/
│   │   │   ├── order.go
│   │   │   ├── item.go
│   │   │   ├── repository.go
│   │   │   └── events.go
│   │   ├── user/
│   │   └── shared/
│   │       ├── money.go            # value object
│   │       └── pagination.go
│   │
│   ├── application/                # ← Laravel "Services" / Actions / Commands
│   │   ├── product/
│   │   ├── order/
│   │   └── auth/
│   │
│   ├── infrastructure/             # ← Laravel "Infrastructure" (DB, cache, mail)
│   │   ├── postgres/
│   │   │   ├── product_repo.go     # implements domain/product.Repository
│   │   │   └── order_repo.go
│   │   └── auth/
│   │       └── jwt.go
│   │
│   ├── transport/                  # ← Laravel "Http/Controllers" + "Routes"
│   │   └── http/
│   │       ├── router.go
│   │       ├── middleware/
│   │       ├── product_handler.go
│   │       └── response/
│   │           └── json.go
│   │
│   ├── database/                   # ← Laravel "database/migrations" + "database/seeders"
│   │   ├── postgres.go
│   │   ├── migrations/
│   │   └── seeders/
│   │
│   └── config/
│       ├── config.go
│       └── seeder.go
│
├── test/                           # ← Laravel "tests/Feature" (Testcontainers)
│   ├── README.md
│   ├── integration/
│   │   ├── helpers/
│   │   │   └── postgres.go         # SetupPostgres() — container + goose migrations
│   │   └── *_test.go               # repository / use-case integration tests
│   └── e2e/                        # HTTP smoke tests (add later)
│
├── api/                            # OpenAPI spec — contract for core + site teams
│   └── openapi.yaml
│
├── configs/
│   └── seeder.yaml
│
├── deployments/
│   ├── docker-compose.yml
│   └── Dockerfile
│
├── scripts/                        # ← VPS production ops (not local dev)
│   ├── README.md
│   ├── build.sh                    # git pull + compile bin/api, bin/migrate
│   ├── deploy.sh                   # build + migrate + systemd restart
│   └── lib/
│       └── common.sh
│
├── docs/
│   ├── architecture.md
│   └── solution.md                 # this document
│
├── Makefile
├── go.mod
├── .env.example
└── README.md
```

### Folders intentionally omitted from this template

| Folder | Reason |
|---|---|
| `pkg/` | ws is an application, not a published library |
| `vendor/` | use module proxy unless constrained |
| `web/`, `website/` | frontends are separate repos (core, site) |
| `tools/`, `third_party/`, `githooks/`, `init/`, `assets/`, `examples/` | add on demand |
| `build/` | CI packaging configs — add when you outgrow scripts + compose |
| `internal/repository/`, `internal/service/`, `internal/handler/` | type-based layers — replaced by domain/application/infrastructure/transport |

---

## Testing Strategy (Testcontainers)

### Why `test/` exists

Root `test/` is the home for **integration and E2E tests** that need real infrastructure. Unit tests stay colocated in `internal/` — that is non-negotiable Go convention.

This starter wires integration tests to [Testcontainers for Go](https://golang.testcontainers.org/): programmatic Docker containers that start before tests, apply migrations, and are destroyed on cleanup. No shared dev database. No mocks for SQL behavior.

### Three tiers

| Tier | Location | Command | Docker |
|---|---|---|---|
| Unit | `internal/**/**/*_test.go` | `make test` | No |
| Integration | `test/integration/` | `make test-integration` | Yes |
| E2E | `test/e2e/` | add `make test-e2e` later | Depends |

### Why Testcontainers over docker-compose for tests?

| Approach | Problem |
|---|---|
| Point tests at `deployments/docker-compose.yml` | Shared state, flaky tests, dev data pollution |
| Mock `*sql.DB` | Misses constraints, migrations, SQL dialect issues |
| **Testcontainers** | Isolated Postgres per test run, same goose migrations as prod, automatic cleanup |

Dependencies (in `go.mod`):

```bash
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

### Starter helper: `test/integration/helpers/postgres.go`

```go
//go:build integration

func SetupPostgres(t *testing.T) *sql.DB
```

1. Starts `postgres:16-alpine` via `testcontainers-go/modules/postgres`
2. Waits with `BasicWaitStrategies()`
3. Opens `database/sql` with `pgx` driver
4. Runs goose migrations from `internal/database/migrations/`
5. Cleans up via `testcontainers.CleanupContainer(t, ctr)` and `t.Cleanup`

### Build tags

All integration code uses `//go:build integration`. Without `-tags=integration`, Go excludes these files — `make test` never touches Docker.

```bash
make test                                    # unit only
make test-integration                        # Testcontainers
go test -race -tags=integration -timeout=5m ./test/integration/... -v
```

### What goes where

| Test this | Location | Why |
|---|---|---|
| Aggregate `Reserve()`, value objects | `internal/domain/product/product_test.go` | Pure logic, no DB |
| `ProductRepository.Save` SQL | `test/integration/product_repo_test.go` | Real Postgres + migrations |
| `GET /api/v1/products` | `test/e2e/` (later) | Full HTTP stack |

### Laravel mapping

| Laravel | Go (this starter) |
|---|---|
| `tests/Unit` | colocated `*_test.go` |
| `tests/Feature` + `RefreshDatabase` | `test/integration/` + Testcontainers fresh DB |
| PHPUnit `DatabaseMigrations` | goose migrations in `SetupPostgres` |

### CI

Run integration tests in a **separate job** on a Docker-enabled runner:

```yaml
# GitHub Actions example
integration-test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
    - run: go test -race -tags=integration -timeout=5m ./test/integration/...
```

See [Testcontainers CI requirements](https://golang.testcontainers.org/system_requirements/ci/).

### Performance

- **Per-test container** (`SetupPostgres(t)`) — maximum isolation, good for starters
- **Shared container via `TestMain`** — faster for large suites; see `.cursor/skills/golang-testcontainers/`

---

## Production Deployment (`scripts/`)

### Why include `scripts/` in a starter kit?

Earlier drafts deferred `scripts/` (“add when Makefile is not enough”). For a **production-oriented** template aimed at VPS deploys, that was too conservative.

[golang-standards/project-layout `/scripts`](https://github.com/golang-standards/project-layout#scripts) exists for build, install, and release operations that would bloat the Makefile. Large projects use the same split:

| Project | `scripts/` role |
|---|---|
| [HashiCorp Terraform](https://github.com/hashicorp/terraform/tree/main/scripts) | Release, packaging, codegen helpers |
| [Kubernetes Helm](https://github.com/kubernetes/helm/tree/master/scripts) | Build, test, and publish automation |
| [CockroachDB](https://github.com/cockroachdb/cockroach/tree/master/scripts) | CI, release, and cluster tooling |

Your stated workflow — **push to `main`, SSH to VPS, run one command** — is exactly what `scripts/` is for.

### Makefile vs scripts vs deployments

| Layer | Purpose | Example |
|---|---|---|
| **`Makefile`** | Developer ergonomics on a laptop | `make run-api`, `make test`, `make migrate-up` |
| **`scripts/`** | Repeatable production operations on a server | `./scripts/build.sh`, `./scripts/deploy.sh` |
| **`deployments/`** | Container images and compose for Docker-based deploy | `Dockerfile`, `docker-compose.yml` |

**Do not** put VPS deploy logic in the Makefile. Production servers run bash scripts that are explicit, log-friendly, and safe to run over SSH.

### Starter scripts

| Script | What it does |
|---|---|
| `scripts/build.sh` | `git pull origin/main` (optional) → `CGO_ENABLED=0 go build` → `bin/api`, `bin/migrate` |
| `scripts/deploy.sh` | `build.sh` → `bin/migrate up` → `systemctl restart` (if `SYSTEMD_SERVICE` set) |
| `scripts/lib/common.sh` | Shared `log`, `require_env`, `die` |

### VPS release flow

```bash
# On the server (after chmod +x scripts/*.sh)
cd /opt/ws
./scripts/deploy.sh
```

Production env at `/etc/ws/ws.env` (not committed):

```bash
APP_ENV=production
DATABASE_URL=postgres://...
HTTP_PORT=8080
SYSTEMD_SERVICE=ws-api
```

### Laravel / Forge mapping

| Laravel / Forge | Go starter |
|---|---|
| Forge “Deploy Now” | `./scripts/deploy.sh` |
| `git pull` on server | built into `build.sh` |
| `php artisan migrate --force` | `bin/migrate up` in `deploy.sh` |
| `.env` on server | `/etc/ws/ws.env` |
| Supervisor restart | `SYSTEMD_SERVICE=ws-api` |

### Trade-off (stated honestly)

| Approach | When |
|---|---|
| **`scripts/deploy.sh` on VPS** | Single server, small team — **this starter** |
| **Docker on VPS** | Immutable images — use `deployments/Dockerfile` |
| **GitHub Actions → SSH** | Automate `deploy.sh` |
| **Kubernetes / PaaS** | Outgrow on-server builds; ship CI-built artifacts |

Scripts build **on the server** (Go required) and are **not zero-downtime**. That matches Forge-style deploys and is fine until you need blue/green or container orchestration.

---

## Laravel → Go Mapping

| Laravel (ws equivalent) | Go location | Notes |
|---|---|---|
| `routes/api.php` | `internal/transport/http/router.go` | Go 1.22+ `ServeMux` method patterns |
| `app/Http/Controllers/*` | `internal/transport/http/*_handler.go` | Handlers are thin; no business logic |
| `app/Http/Middleware/*` | `internal/transport/http/middleware/` | Auth, CORS, logging |
| `app/Models/*` (behavior-rich) | `internal/domain/{aggregate}/` | Aggregates with methods, not Eloquent |
| `app/Services/*` | `internal/application/{aggregate}/` | One file per use case is fine |
| `app/Repositories/*` or Eloquent | `internal/domain/{aggregate}/` (interface) + `internal/infrastructure/postgres/` (impl) | Interface in domain, impl in infra |
| `database/migrations/` | `internal/database/migrations/` | goose SQL files |
| `database/seeders/` | `internal/database/seeders/` | interface + ordered runner |
| `config/*.php` | `internal/config/` + `configs/` templates | runtime load vs committed defaults |
| `.env` | `.env.example` + `internal/config` | never commit secrets |
| `artisan migrate` | `make migrate-up` | |
| `artisan db:seed` | `make seed` | |
| `artisan migrate:fresh --seed` | `make migrate-fresh-seed` | |
| `tests/Unit` | `internal/**/**/*_test.go` | colocated with source |
| `tests/Feature` (DB) | `test/integration/` | Testcontainers + goose migrations |
| Forge “Deploy Now” | `scripts/deploy.sh` | VPS: pull, build, migrate, restart |
| Server `.env` | `/etc/ws/ws.env` | production secrets outside repo |

---

## core and site (frontend repos)

These are **not** Go projects. They do not use this layout.

```
core/                          # admin / operations SPA
├── src/
│   ├── api/                   # typed client from ws/api/openapi.yaml
│   └── features/
├── package.json
└── .env.example               # VITE_API_URL=https://ws.example.com

site/                          # public marketing site
├── src/
├── package.json
└── .env.example               # NEXT_PUBLIC_API_URL=...
```

**Contract between repos:** ws publishes `api/openapi.yaml`. core and site generate clients from it. No shared Go packages across repos — only HTTP contracts.

---

## Dependency Flow

```
┌─────────────────────────────────────────────────────────┐
│  cmd/api/main.go                                        │
│  (composition root — the only place that knows all types) │
└──────────────────────────┬──────────────────────────────┘
                           │ wires
                           ▼
┌─────────────────────────────────────────────────────────┐
│  transport/http          inbound adapters               │
│  handlers, middleware, router                           │
└──────────────────────────┬──────────────────────────────┘
                           │ calls
                           ▼
┌─────────────────────────────────────────────────────────┐
│  application/            use cases (orchestration)      │
└──────────────────────────┬──────────────────────────────┘
                           │ uses interfaces from
                           ▼
┌─────────────────────────────────────────────────────────┐
│  domain/                 aggregates, value objects,       │
│                          repository interfaces (ports)  │
└──────────────────────────▲──────────────────────────────┘
                           │ implements
┌──────────────────────────┴──────────────────────────────┐
│  infrastructure/         postgres repos, jwt, crypto      │
│  database/               migrations, seeders, pool      │
└─────────────────────────────────────────────────────────┘
```

**Rule:** `domain` imports nothing from `application`, `infrastructure`, or `transport`.

---

## Code Sketch: Domain Aggregate

**Anemic (avoid):**

```go
type Product struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Price string `json:"price"`
    Stock int    `json:"stock"`
}
```

**DDD aggregate (this starter):**

```go
// internal/domain/product/product.go
type Product struct {
    id    string
    name  string
    price shared.Money
    stock int
}

func New(name, slug string, price shared.Money, stock int) (*Product, error) {
    if stock < 0 {
        return nil, ErrInvalidStock
    }
    // ...
}

func (p *Product) Reserve(qty int) error {
    if qty > p.stock {
        return ErrInsufficientStock
    }
    p.stock -= qty
    return nil
}
```

```go
// internal/domain/product/product_test.go — colocated unit test
func TestReserve_InsufficientStock(t *testing.T) { /* ... */ }
```

JSON serialization stays in `transport/http/product_handler.go` — not in domain structs.

---

## Migration Path When Forking This Template

### Phase 1 — Rename and configure

1. Change `go.mod` module path to `github.com/yourorg/ws`.
2. Update imports across the repo.
3. Copy `.env.example` → `.env`, run `make docker-db-up`, `make migrate-fresh-seed`, `make run-api`.

### Phase 2 — Grow domain

1. Add aggregates under `internal/domain/{name}/` with behavior and colocated tests.
2. Add use cases under `internal/application/{name}/`.
3. Add Postgres repos under `internal/infrastructure/postgres/`.

### Phase 3 — Add integration tests (Testcontainers)

1. Create `test/integration/product_repo_test.go` with `//go:build integration`.
2. Use `helpers.SetupPostgres(t)` — already provided in `test/integration/helpers/postgres.go`.
3. Run with `make test-integration` (requires Docker).

Example:

```go
//go:build integration

package integration_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/yourorg/ws/internal/domain/product"
    "github.com/yourorg/ws/internal/domain/shared"
    "github.com/yourorg/ws/internal/infrastructure/postgres"
    "github.com/yourorg/ws/test/integration/helpers"
)

func TestProductRepository_SaveAndFindBySlug(t *testing.T) {
    db := helpers.SetupPostgres(t)
    repo := postgres.NewProductRepository(db)
    ctx := context.Background()

    price, err := shared.NewMoney("19.99", "PHP")
    require.NoError(t, err)

    p, err := product.New("Kit", "starter-kit", "desc", price, 10)
    require.NoError(t, err)
    require.NoError(t, repo.Save(ctx, p))

    found, err := repo.FindBySlug(ctx, "starter-kit")
    require.NoError(t, err)
    require.Equal(t, p.ID(), found.ID())
}
```

### Phase 4 — Contract with frontends

1. Keep `api/openapi.yaml` in sync with handlers.
2. core and site generate TypeScript clients from it.

### Phase 5 — Production VPS

1. Clone repo to `/opt/ws` on the server.
2. Create `/etc/ws/ws.env` with `DATABASE_URL` and `APP_ENV=production`.
3. Add systemd unit for `bin/api` (see `scripts/README.md`).
4. `chmod +x scripts/*.sh` and run `./scripts/deploy.sh` on every release.

---

## Decision Summary

| Question | Answer |
|---|---|
| project-layout or layered tutorial structure? | **Neither alone.** Skeleton from project-layout (`cmd`, `internal`, `configs`, `deployments`, `api`, `docs`, `test`). Internal organization from DDD + package-by-feature |
| Monorepo or multi-repo? | **Multi-repo** — ws, core, site are separate. One `go.mod` per Go repo |
| Layered or DDD? | **DDD tactical inside `internal/domain/`**, with a thin application layer |
| `pkg/`? | **No** for ws unless you extract a shared Go SDK |
| Type-based or domain-based packages? | **Domain-based** (`product/`, `order/`) |
| Where do migrations/seeders go? | **`internal/database/`** |
| Where do unit tests go? | **Colocated** `*_test.go` in `internal/` |
| Where do integration tests go? | **`test/integration/`** with Testcontainers |
| Testcontainers deps? | `testcontainers-go` + `modules/postgres` in `go.mod` |
| `scripts/` for VPS deploy? | **Yes** — `build.sh` + `deploy.sh`; Makefile stays dev-only |
| Hexagonal/Clean Architecture folders? | **Defer** until you have 2+ inbound/outbound adapters |
| Where do frontends go? | **Not in ws.** core and site repos consume `api/openapi.yaml` |

---

## Production Readiness Assessment

| Criterion | Status |
|---|---|
| Compiles (`go build ./...`) | Yes |
| Runnable API | `make run-api` |
| Migrations / seeders | goose + Makefile targets |
| DDD structure | domain → application → infrastructure → transport |
| Unit tests | `internal/domain/product/product_test.go` |
| Integration test helper | `test/integration/helpers/postgres.go` (Testcontainers) |
| Integration test command | `make test-integration` |
| VPS deploy scripts | `scripts/build.sh`, `scripts/deploy.sh` |
| OpenAPI contract | `api/openapi.yaml` |
| Docker / compose | `deployments/docker-compose.yml` |
| Graceful shutdown | SIGTERM handling in `cmd/api` |

A **starter** is production-ready when it gives you a working, idiomatic foundation. A **product** is production-ready when it meets your SLAs, security, and compliance requirements. This template is the former.

### Next steps when forking

1. Rename module path in `go.mod` and imports.
2. Add aggregates one at a time under `internal/domain/{name}/` with colocated tests.
3. Add `test/integration/*_test.go` using `helpers.SetupPostgres(t)`.
4. Publish `api/openapi.yaml` for **core** and **site** teams.
5. Configure VPS: `/etc/ws/ws.env`, systemd, `./scripts/deploy.sh`.
6. Add CI: `go vet`, `go test -race ./...`, `go test -race -tags=integration ./test/integration/...` on Docker runners.

---

## Further Reading

- [Organizing a Go module](https://go.dev/doc/modules/layout) — official guidance
- [Testcontainers for Go](https://golang.testcontainers.org/) — integration test containers
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout) — community layout reference (`/test`, `/scripts`)
- [scripts/README.md](../scripts/README.md) — VPS deploy workflow
- [docs/architecture.md](architecture.md) — layer overview
- [docs/medium.md](medium.md) — layered intro and scaling diagram
- DDD tactical patterns — see `.cursor/skills/golang-ddd/SKILL.md` in this repo
