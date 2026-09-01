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

## Multi-repo

- **ws** exposes HTTP + `api/openapi.yaml`
- **core** and **site** are separate frontend repos — they do not import Go packages from ws

See [solution.md](solution.md) for the full design document.
