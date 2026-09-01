# Go Project Structure for Laravel Developers (Multi-Repo + DDD)

A balanced layout recommendation that merges [golang-standards/project-layout](../README.md) with the practical patterns from [medium.md](../medium.md), adapted for a **multi-repository** product split and **Domain-Driven Design** in idiomatic Go.

---

## Context

| Repository | Role | Example |
|---|---|---|
| **ws** | Backend API (Go) | REST/JSON API, auth, business rules, database |
| **core** | Operations frontend | Admin panel, internal dashboards |
| **site** | Public frontend | Marketing/public site (e.g. https://www.pdic.gov.ph/) |

Each repo is independent. Only **ws** is Go. **core** and **site** consume ws over HTTP — the same way a Laravel API serves a Vue/React SPA or a separate Next.js site.

This document focuses on **ws** (backend) structure and how **project-layout/** should be redesigned as a template for that repo.

---

## Review: Two Sources, Both Sides

### golang-standards/project-layout (README)

**What to adopt**

| Pattern | Why |
|---|---|
| `cmd/` per binary | Official Go convention; thin `main.go` that only wires dependencies |
| `internal/` | Compiler-enforced privacy — Go's primary architectural boundary |
| `configs/` | Config templates separate from runtime loading code |
| `deployments/` | Docker, compose, K8s — keeps infra out of application code |
| `api/` | OpenAPI/JSON schema contracts shared with core and site teams |
| `docs/` | Architecture decisions, onboarding — not godoc |
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

### Medium Article (go-level-structure)

**What to adopt**

| Pattern | Why |
|---|---|
| Layered flow: transport → application → persistence | Familiar to Laravel developers (Controller → Service → Model/Repository) |
| `cmd/api`, `cmd/migrate`, `cmd/seed` | Maps cleanly to `artisan serve`, `artisan migrate`, `artisan db:seed` |
| Makefile targets | `make run-api`, `make migrate-up`, `make seed` — excellent DX for teams new to Go |
| Scale path: package-by-feature (`internal/users/`, `internal/products/`) | Aligns with Go's "package by responsibility" principle |

**What to avoid or defer**

| Pattern | Why |
|---|---|
| Global `internal/handlers/`, `internal/services/`, `internal/store/` | Works for tutorials; becomes a "junk drawer" at 10+ domains. Files grow, imports tangle, boundaries blur |
| `internal/models/` | Anemic structs with JSON tags are not domain models. Collides with DDD where the model has behavior |
| `internal/dto/` as a top-level package | DTOs belong at the transport boundary (HTTP request/response shapes), not as a parallel domain layer |
| `pkg/utils`, `pkg/response`, `pkg/middleware` for everything | Generic `utils` packages become dependency magnets. Prefer small, purpose-named internal packages |
| Full Hexagonal `ports/` + `adapters/` + `bootstrap/` on day one | Correct for large systems; heavy for a team still learning Go. Introduce when a second transport (gRPC, CLI) or second persistence backend appears |
| `.env` committed to repo | Use `.env.example` only; real secrets stay local or in a secret manager |

**Core philosophy alignment:** The article's best scaling advice is at the end — **package by domain** (`users/`, `products/`) — not its initial type-based layout. Lead with that for DDD.

---

## Review: Current `golang-monorepo/ws`

The existing ws implementation is a solid **layered service** — a good Laravel stepping stone — but not yet **tactical DDD**.

### What ws does well (keep)

```
ws/
├── cmd/api/main.go          # Thin entry — wires config, DB, router
├── cmd/migrate/main.go      # goose migrations, Laravel-like commands (up/down/fresh/rollback)
├── cmd/seed/main.go         # Ordered seeder runner
├── cmd/keygen/main.go       # App key generation (Laravel APP_KEY equivalent)
├── internal/database/
│   ├── migrations/          # Timestamped SQL files
│   └── seeders/             # Interface-based seeders + SQL seed files
├── configs/seeder.yaml      # Seeder configuration
├── deployments/             # docker-compose, Dockerfile
├── Makefile                 # migrate-up, migrate-fresh-seed, seed, run-api
└── .env.example
```

- Migration/seeder workflow is production-ready and familiar to Laravel developers.
- `cmd/` binaries are correctly separated.
- `deployments/` and `configs/` follow project-layout conventions.

### What to evolve (DDD + Go idioms)

| Current | Issue | Target |
|---|---|---|
| `internal/domain/product.go` — public struct fields + JSON tags | Anemic data holder; domain leaks HTTP serialization concerns | Aggregate package with unexported fields, behavior methods, no JSON tags |
| `internal/repository/` — one package, concrete structs | Repository interfaces live far from the aggregate; hard to mock per domain | `Repository` interface in `internal/domain/product/`; Postgres impl in `internal/infrastructure/postgres/` |
| `internal/service/` — depends on concrete `*repository.X` | Violates "accept interfaces, return structs"; hard to test | `internal/application/product/` use cases depend on domain interfaces |
| `internal/handler/routes.go` — 150-line god wiring file | All DI in one place will not scale | `internal/transport/http/router.go` + per-module registration; optional `internal/bootstrap/wire.go` |
| `internal/middleware/`, `internal/auth/` | Fine, but auth is infrastructure | Move to `internal/infrastructure/auth/` or `internal/transport/http/middleware/` |
| Module path `golang-monorepo` | Misleading for a multi-repo world | `github.com/yourorg/ws` — one module per repo |

---

## Recommended Structure

### Design principles (Go-first)

1. **Package by responsibility, not by type.** `product/`, `order/` — not `handlers/`, `repositories/`.
2. **`internal/` is the boundary.** Everything application-specific stays inside; nothing in `pkg/` unless another repo imports it as a library (unlikely for ws).
3. **Dependencies point inward.** `domain` → zero imports from application/infrastructure/transport.
4. **Interfaces where they are used.** Repository interfaces live in the domain aggregate package; Postgres implements them in infrastructure.
5. **Thin `cmd/`.** Only `main()` and fatal error handling.
6. **Start simple, grow layers.** Begin with domain + application + one infrastructure adapter. Add hexagonal folders when a second adapter appears.
7. **No `/src`.** Ever.

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
│   │   │   ├── repository.go       # interface (port) — Find, Save, List
│   │   │   └── errors.go           # ErrNotFound, ErrInvalidPrice, etc.
│   │   ├── order/
│   │   │   ├── order.go
│   │   │   ├── item.go             # child entity (not an aggregate root)
│   │   │   ├── repository.go
│   │   │   └── events.go           # OrderPlaced, OrderCancelled
│   │   ├── user/
│   │   └── shared/
│   │       ├── money.go            # value object
│   │       ├── pagination.go
│   │       └── errors.go
│   │
│   ├── application/                # ← Laravel "Services" / Actions / Commands
│   │   ├── product/
│   │   │   ├── create.go           # CreateProduct use case
│   │   │   ├── list.go
│   │   │   └── update.go
│   │   ├── order/
│   │   │   ├── place.go
│   │   │   └── cancel.go
│   │   └── auth/
│   │       ├── register.go
│   │       └── login.go
│   │
│   ├── infrastructure/             # ← Laravel "Infrastructure" (DB, cache, mail)
│   │   ├── postgres/
│   │   │   ├── product_repo.go     # implements domain/product.Repository
│   │   │   ├── order_repo.go
│   │   │   └── user_repo.go
│   │   ├── auth/
│   │   │   └── jwt.go
│   │   └── crypto/
│   │       └── encrypter.go
│   │
│   ├── transport/                  # ← Laravel "Http/Controllers" + "Routes"
│   │   └── http/
│   │       ├── router.go           # mux setup, middleware chain
│   │       ├── middleware/
│   │       │   ├── auth.go
│   │       │   ├── cors.go
│   │       │   └── logging.go
│   │       ├── product_handler.go  # JSON bind → use case → JSON respond
│   │       ├── order_handler.go
│   │       └── response/
│   │           └── json.go
│   │
│   ├── database/                   # ← Laravel "database/migrations" + "database/seeders"
│   │   ├── postgres.go             # connection pool
│   │   ├── migrations/
│   │   │   └── YYYYMMDD_HHMMSS_*.sql
│   │   └── seeders/
│   │       ├── seeder.go           # Seeder interface + RunAll
│   │       ├── user_seeder.go
│   │       └── sql/
│   │
│   ├── config/
│   │   ├── config.go               # env loading
│   │   └── seeder.go
│   │
│   └── bootstrap/                  # optional; add when wiring grows
│       └── app.go                  # NewApp(cfg) → *http.Handler
│
├── api/                            # OpenAPI spec — contract for core + site teams
│   └── openapi.yaml
│
├── configs/                        # committed config templates
│   └── seeder.yaml
│
├── deployments/
│   ├── docker-compose.yml
│   └── Dockerfile
│
├── docs/
│   └── architecture.md
│
├── scripts/                        # only when Makefile is not enough
│   └── ci-lint.sh
│
├── Makefile
├── go.mod                          # module github.com/yourorg/ws
├── .env.example
└── README.md
```

### Folders intentionally omitted from project-layout template

| Folder | Reason |
|---|---|
| `pkg/` | ws is an application, not a published library |
| `vendor/` | use module proxy unless constrained |
| `web/`, `website/` | frontends are separate repos (core, site) |
| `tools/`, `third_party/`, `githooks/`, `init/`, `assets/`, `examples/` | add on demand |
| `test/` at root | colocate `*_test.go` with packages; use `testdata/` per package |
| `internal/repository/`, `internal/service/`, `internal/handler/` | type-based layers — replaced by domain/application/infrastructure/transport |

---

## Laravel → Go Mapping

| Laravel (ws equivalent) | Go location | Notes |
|---|---|---|
| `routes/api.php` | `internal/transport/http/router.go` | Go 1.22+ `ServeMux` method patterns |
| `app/Http/Controllers/*` | `internal/transport/http/*_handler.go` | Handlers are thin; no business logic |
| `app/Http/Middleware/*` | `internal/transport/http/middleware/` | Auth, CORS, logging |
| `app/Models/*` (behavior-rich) | `internal/domain/{aggregate}/` | Aggregates with methods, not Eloquent |
| `app/Services/*` | `internal/application/{aggregate}/` | One file per use case is fine |
| `app/Repositories/*` or Eloquent | `internal/domain/{aggregate}/repository.go` (interface) + `internal/infrastructure/postgres/` (impl) | Interface in domain, impl in infra |
| `database/migrations/` | `internal/database/migrations/` | Keep existing goose setup |
| `database/seeders/` | `internal/database/seeders/` | Keep existing seeder interface |
| `config/*.php` | `internal/config/` + `configs/` templates | Runtime load vs committed defaults |
| `.env` | `.env.example` + `internal/config` | Never commit secrets |
| `artisan migrate` | `make migrate-up` / `go run ./cmd/migrate up` | |
| `artisan db:seed` | `make seed` / `go run ./cmd/seed` | |
| `artisan migrate:fresh --seed` | `make migrate-fresh-seed` | Already implemented in ws |
| `php artisan key:generate` | `make key-generate` / `go run ./cmd/keygen` | |
| Form Requests / API Resources | Request structs in handler file or `transport/http/dto/` per module | Not a global `dto/` package |
| Events / Listeners | `internal/domain/{aggregate}/events.go` + application publisher | Domain events are immutable structs |
| Policies / Gates | `internal/domain/shared/permission.go` + middleware | RBAC rules in domain; enforcement in middleware |

---

## core and site (frontend repos)

These are **not** Go projects. They do not use this layout.

```
core/                          # admin / operations SPA
├── src/
│   ├── api/                   # typed client generated from ws/api/openapi.yaml
│   ├── features/              # feature folders (products, orders, users)
│   └── ...
├── package.json
└── .env.example               # VITE_API_URL=https://ws.example.com

site/                          # public marketing site
├── src/
├── package.json
└── .env.example               # NEXT_PUBLIC_API_URL=...
```

**Contract between repos:** ws publishes `api/openapi.yaml` (or hosts `/openapi.json`). core and site generate clients from it. No shared Go packages across repos — only HTTP contracts.

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

## Code Sketch: Domain vs Current ws

**Current (anemic — ws today):**

```go
// internal/domain/product.go
type Product struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Price string `json:"price"`
    Stock int    `json:"stock"`
}
```

**Target (DDD aggregate — recommended):**

```go
// internal/domain/product/product.go
package product

type Product struct {
    id    string
    name  string
    price Money
    stock int
}

func New(name string, price Money, stock int) (*Product, error) {
    if stock < 0 {
        return nil, ErrInvalidStock
    }
    return &Product{
        id:    uuid.New().String(),
        name:  name,
        price: price,
        stock: stock,
    }, nil
}

func (p *Product) Reserve(qty int) error {
    if qty > p.stock {
        return ErrInsufficientStock
    }
    p.stock -= qty
    return nil
}

func (p *Product) ID() string   { return p.id }
func (p *Product) Name() string { return p.name }
```

```go
// internal/domain/product/repository.go
package product

type Repository interface {
    Find(ctx context.Context, id string) (*Product, error)
    Save(ctx context.Context, p *Product) error
    List(ctx context.Context, filter ListFilter) ([]*Product, error)
}
```

```go
// internal/infrastructure/postgres/product_repo.go
package postgres

type ProductRepo struct { db *sql.DB }

func (r *ProductRepo) Find(ctx context.Context, id string) (*product.Product, error) {
    // SQL → Reconstruct(id, name, price, stock) — bypasses New() validation
}
```

JSON serialization stays in `transport/http/product_handler.go` — not in domain structs.

---

## Migration Path from Current ws

Do this incrementally. Do not rewrite everything at once.

### Phase 1 — Rename layers (low risk)

| From | To |
|---|---|
| `internal/handler/` | `internal/transport/http/` |
| `internal/service/` | `internal/application/` (split into per-aggregate subpackages over time) |
| `internal/repository/` | `internal/infrastructure/postgres/` |
| Keep `internal/database/` | No change — migrations/seeders are already correct |

### Phase 2 — Enrich domain (medium risk)

1. Pick one aggregate (e.g. `order` — it has real business rules: status transitions, prestige, cart).
2. Create `internal/domain/order/` with behavior methods.
3. Move `Repository` interface into that package.
4. Update application use case to depend on `order.Repository` interface, not `*postgres.OrderRepo`.

### Phase 3 — Extract use cases (medium risk)

Split `internal/application/order/order_service.go` into:

```
application/order/
├── place.go
├── cancel.go
├── confirm_payment.go
└── list_admin.go
```

### Phase 4 — Contract with frontends

1. Add `api/openapi.yaml` describing endpoints ws already exposes.
2. core and site generate TypeScript clients from it.

### Phase 5 — Module rename

Change `go.mod` from `golang-monorepo` to `github.com/yourorg/ws` when the repo is split.

---

## project-layout Template Redesign

The `project-layout/` repo should become a **ws backend starter** — not a copy of every folder from golang-standards/project-layout.

### Keep in template

```
project-layout/
├── cmd/
│   ├── api/
│   ├── migrate/
│   └── seed/
├── internal/
│   ├── domain/
│   │   └── example/            # sample aggregate with README
│   ├── application/
│   │   └── example/
│   ├── infrastructure/
│   │   └── postgres/
│   ├── transport/
│   │   └── http/
│   ├── database/
│   │   ├── migrations/
│   │   └── seeders/
│   └── config/
├── api/
│   └── openapi.yaml
├── configs/
├── deployments/
├── docs/
│   └── solution.md             # this document
├── Makefile
├── go.mod
└── .env.example
```

### Remove or demote to docs-only mention

- `pkg/`, `vendor/`, `web/`, `website/`, `tools/`, `third_party/`, `githooks/`, `init/`, `assets/`, `examples/`, `build/`, `scripts/` (until needed), `test/` at root

### Add to template README

1. Link to this `solution.md`.
2. Laravel mapping table (abbreviated).
3. Note that **core** and **site** are separate repos.
4. Explicit warning: "Do not add `pkg/` unless you publish a Go library."

---

## Makefile (retain ws patterns)

```makefile
DATABASE_URL ?= postgres://postgres:root@localhost:5433/app?sslmode=disable

.PHONY: run-api migrate-up migrate-down migrate-fresh migrate-fresh-seed seed

run-api:
	go run ./cmd/api

migrate-up:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate up

migrate-fresh-seed: migrate-fresh seed

migrate-fresh:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate fresh

seed:
	go run ./cmd/seed
```

This preserves the Laravel muscle memory your team already built in ws.

---

## Decision Summary

| Question | Answer |
|---|---|
| project-layout or medium structure? | **Neither alone.** Skeleton from project-layout (`cmd`, `internal`, `configs`, `deployments`, `api`, `docs`). Internal organization from DDD + medium's package-by-feature scaling path |
| Monorepo or multi-repo? | **Multi-repo** — ws, core, site are separate. One `go.mod` per Go repo |
| Layered or DDD? | **DDD tactical inside `internal/domain/`**, with a thin application layer. Avoid anemic `models/` |
| `pkg/`? | **No** for ws unless you extract a shared Go SDK (unlikely) |
| Type-based or domain-based packages? | **Domain-based** (`product/`, `order/`). Type-based only for the first tutorial sprint |
| Where do migrations/seeders go? | **`internal/database/`** — keep the ws implementation; it maps directly to Laravel |
| Hexagonal/Clean Architecture folders? | **Defer** `ports/`, `adapters/`, `bootstrap/` until you have 2+ inbound/outbound adapters |
| Where do frontends go? | **Not in ws.** core and site repos consume `api/openapi.yaml` |

---

## Further Reading

- [Organizing a Go module](https://go.dev/doc/modules/layout) — official guidance
- [Effective Go — Names](https://go.dev/doc/effective_go#names)
- [Package names (blog)](https://go.dev/blog/package-names)
- [golang-standards/project-layout](../README.md) — community layout reference
- [medium.md](../medium.md) — layered intro and scaling diagram
- DDD tactical patterns — see `.cursor/skills/golang-ddd/SKILL.md` in this repo

---

## Production Readiness Assessment (Appendix)

*Updated after removing deferred folders and implementing the starter template.*

### Was `project-layout/` production-ready after folder cleanup?

**No.** Deleting `pkg/`, `vendor/`, `web/`, `website/`, `tools/`, and other deferred folders was the **correct first step**, but the repo was still only a **directory skeleton** — README placeholders in `cmd/`, `internal/`, `api/`, etc., a placeholder `go.mod`, and no compilable application code. That is documentation of layout ideas, not a deployable backend.

| Criterion | Before cleanup + delete | After this implementation |
|---|---|---|
| Compiles (`go build ./...`) | No Go source | Yes |
| Runnable API | No | `make run-api` |
| Migrations / seeders | No | goose + `make migrate-up` / `make seed` |
| DDD structure | No | `domain` → `application` → `infrastructure` → `transport` |
| Tests | No | `go test ./...` on domain layer |
| OpenAPI contract | No | `api/openapi.yaml` |
| Docker / compose | README only | `deployments/docker-compose.yml` |
| Graceful shutdown | N/A | SIGTERM handling in `cmd/api` |
| Auth / RBAC | N/A | Not in starter — add when porting from `golang-monorepo/ws` |

### Verdict

| State | Meaning |
|---|---|
| **Before** | Folder catalog only — not production-ready |
| **Now** | **Starter production-ready** — safe to clone, rename module, run locally, and grow feature-by-feature |
| **Not yet** | Full production system — still needs auth, CI, observability, integration tests, and your real domain aggregates |

A **starter** is production-ready when it gives you a working, idiomatic foundation. A **product** is production-ready when it meets your SLAs, security, and compliance requirements. This template is the former.

### What was implemented in `project-layout/`

```
project-layout/
├── cmd/api/              # HTTP server with graceful shutdown
├── cmd/migrate/          # goose runner (up/down/fresh/rollback/create)
├── cmd/seed/             # YAML-driven seeder
├── internal/
│   ├── domain/product/   # example DDD aggregate + tests
│   ├── domain/shared/    # Money value object
│   ├── application/product/
│   ├── infrastructure/postgres/
│   ├── transport/http/
│   ├── database/
│   └── config/
├── api/openapi.yaml
├── configs/seeder.yaml
├── deployments/
├── docs/architecture.md
├── Makefile
├── go.mod
└── .env.example
```

### Folders removed (deferred — do not re-add without need)

- `pkg/`, `vendor/`, `web/`, `website/`, `tools/`, `third_party/`, `githooks/`, `init/`, `assets/`, `examples/`
- `build/`, `scripts/`, `test/` at repo root (use Makefile + colocated `*_test.go` instead)

### Next steps when forking this template

1. Rename module: `github.com/yourorg/ws` → your real path in `go.mod` and imports.
2. Copy auth/RBAC patterns from `golang-monorepo/ws` into `internal/infrastructure/auth` and middleware.
3. Add aggregates one at a time under `internal/domain/{name}/`.
4. Publish `api/openapi.yaml` for **core** and **site** teams to generate TypeScript clients.
5. Add CI: `go vet`, `go test -race ./...`, `staticcheck` or `golangci-lint`.

### Clone → run checklist

```bash
cp .env.example .env
make docker-db-up
make migrate-up
make seed
make run-api
curl http://localhost:8080/up
curl http://localhost:8080/api/v1/products
```
