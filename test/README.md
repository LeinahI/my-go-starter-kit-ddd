# Integration and E2E tests

This directory holds **black-box** tests that need real infrastructure. Integration tests use [Testcontainers for Go](https://golang.testcontainers.org/) to spin up disposable Postgres containers — the same approach used in production CI pipelines without pointing tests at your local dev database.

## Where tests live

| Type | Location | Command | Needs Docker |
|---|---|---|---|
| **Unit** | Colocated: `internal/domain/product/product_test.go` | `make test` | No |
| **Integration** | `test/integration/` | `make test-integration` | Yes |
| **E2E / smoke** | `test/e2e/` | `make test-e2e` (add when ready) | Depends |

```bash
make test                 # fast — unit tests only
make test-integration     # Postgres via testcontainers + goose migrations
make test-all             # both
```

## Why Testcontainers?

Mocks do not catch real SQL mistakes — wrong constraints, type coercion, migration drift. [Testcontainers for Go](https://golang.testcontainers.org/) starts a real `postgres:16-alpine` container per test (or per package via `TestMain`), runs your code against it, and tears it down automatically.

This starter already includes:

```
test/
├── README.md
├── integration/
│   └── helpers/
│       └── postgres.go     # SetupPostgres() — container + migrations
└── e2e/                    # HTTP smoke tests (add later)
```

### `helpers.SetupPostgres`

```go
//go:build integration

package product_test

import (
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/yourorg/ws/internal/infrastructure/postgres"
    "github.com/yourorg/ws/test/integration/helpers"
)

func TestProductRepository_SaveAndFind(t *testing.T) {
    db := helpers.SetupPostgres(t)
    repo := postgres.NewProductRepository(db)

    // test real SQL against a migrated schema
}
```

`SetupPostgres` (in `test/integration/helpers/postgres.go`):

1. Starts a Postgres container via `testcontainers-go/modules/postgres`
2. Opens a `database/sql` connection with the `pgx` driver
3. Runs goose migrations from `internal/database/migrations/`
4. Registers cleanup with `testcontainers.CleanupContainer` and `t.Cleanup`

All integration code is gated with `//go:build integration` so `make test` never requires Docker.

## Writing a new integration test

1. Create `test/integration/{name}_test.go` with the build tag:

```go
//go:build integration

package integration_test
```

2. Call `helpers.SetupPostgres(t)` — do not reuse `deployments/docker-compose.yml` or your `.env` `DATABASE_URL`. Tests must be isolated from dev data.

3. Test infrastructure code (`internal/infrastructure/postgres/`) or full use-case flows — not domain unit logic (that stays colocated).

4. Run from the **repo root** (migrations path is relative):

```bash
make test-integration
# or
go test -race -tags=integration -timeout=5m ./test/integration/... -v
```

## Prerequisites

| Requirement | Check |
|---|---|
| Docker running | `docker info` |
| Go modules installed | `go mod tidy` |
| testcontainers-go | already in `go.mod` |

Dependencies (already added to this starter):

```bash
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

## Build tags

Integration tests use `//go:build integration`. Without `-tags=integration`, Go skips these files entirely — including the `helpers` package. That keeps `go test ./...` fast and Docker-free.

For CI, run integration tests in a separate job on a Docker-enabled runner (e.g. `ubuntu-latest` on GitHub Actions). See the [Testcontainers CI docs](https://golang.testcontainers.org/system_requirements/ci/).

## Performance tips

| Pattern | When |
|---|---|
| One container per test (`SetupPostgres(t)`) | Few tests, maximum isolation |
| Shared container via `TestMain` | Many tests in one package — see `.cursor/skills/golang-testcontainers/` |
| `postgres:16-alpine` | Lightweight image, fast startup |

## Fixtures

| Scope | Location |
|---|---|
| Single package | `internal/domain/product/testdata/` |
| Shared across integration tests | `test/testdata/` |

Go ignores directories named `testdata` during builds.

## Laravel mapping

| Laravel | This starter |
|---|---|
| `tests/Unit` | `internal/**/**/*_test.go` |
| `tests/Feature` (needs DB) | `test/integration/` with Testcontainers |
| `RefreshDatabase` trait | Fresh container + goose migrations per test (or `TestMain` snapshot) |

## Further reading

- [Testcontainers for Go — quickstart](https://golang.testcontainers.org/quickstart/)
- [Postgres module](https://golang.testcontainers.org/modules/postgres/)
- Project skill: `.cursor/skills/golang-testcontainers/SKILL.md`
