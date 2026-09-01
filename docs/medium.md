A Practical, Production-Ready Go Project Structure
==================================================

![captionless image](https://miro.medium.com/v2/resize:fit:1400/format:webp/1*PDaEc4LtMKYSzDyflyQYIw.png)

If you’re learning Go and confused about **folder structure**, **project layout**, or **why everyone has strong opinions about** `**internal**`, congrats — you’ve just opened one of the cleanest and easiest-to-follow guides on the topic.

### This tutorial will walk you through:

*   How to structure real Go projects like a pro
*   Why, as a Go developer, I love this simple folder layout even though very highly debatable and opinionated.
*   Other architectures: Clean Architecture, Hexagonal, Repository Pattern
*   Beautiful folder diagrams
*   And of course: why Go itself is such a fantastic language

### Why Choose Go in the First Place?

Before we talk folders, let’s appreciate the language itself.
Go gives you:

### 1. Fast runtime

Compiled straight to machine code. No JVM. No Python interpreter. Just: build → run → fly. Perfect for APIs, microservices, and CLIs.

2. Statically typed + readable
-------------------------------

Go basically said:

> _“What if we made a language as simple as Python,
> but as safe and fast as C?”_

And then they actually did it.

3. Compiled (but really fast)
------------------------------

Compile times are so fast that sometimes you barely notice the build happened.

4. Built-in concurrency
------------------------

Goroutines + channels = absolute magic.

```
go doSomething()
```

Boom. You’re doing async tasks like a wizard.

5. Automatic garbage collection
--------------------------------

You get memory safety without babysitting your RAM manually.

6. Easy to ship
----------------

A Go binary is just one file. One. Single. File. Move it anywhere. Send it to your friend. Deploy it to a server. This is why Go is used by major companies/brands and now you😎.

### Go Project Structure — [Github repo](https://github.com/Kibaru/go-level-structure.git)

```
go-level-structure/
│
├── cmd/
│   └── main.go                  # Application entrypoint (bootstrap, start server)
│
├── internal/
│   ├── handlers/                # HTTP handlers for incoming REST requests
│   │   └── todo.go              # Todo HTTP handler (Gin logic, JSON binding)
│   │
│   ├── store/                   # Data persistence layer (Repository pattern)
│   │   └── memory.go            # In-memory store (useful for dev/testing)
│   │
│   ├── routes/                  # App routing setup
│   │   └── routes.go            # Registers endpoints + middleware binding
│   │
│   ├── services/                # Business logic layer (recommended addition)
│   │   └── todo_service.go      # Todo business logic using stores + external clients
│   │
│   ├── clients/                 # External REST API clients (HTTP calls to other services)
│   │   └── weather_client.go    # Example: calling external weather API
│   │
│   ├── dto/                     # Request/response DTOs (optional)
│   │   └── todo_dto.go
│   │
│   └── config/                  # Internal config helpers (if not using pkg/config)
│       └── loader.go
│
├── pkg/                         # Reusable helper libraries (no business logic)
│   ├── config/
│   │   └── config.go            # Load env vars, config structs, Viper, etc.
│   │
│   ├── logger/
│   │   └── logger.go            # Project-wide logger (zerolog/logrus/zap wrapper)
│   │
│   ├── middleware/
│   │   ├── cors.go              # CORS setup
│   │   └── requestid.go         # Injects X-Request-ID for tracing
│   │
│   ├── response/
│   │   └── json.go              # Unified JSON success/error response helpers
│   │
│   ├── utils/
│   │   ├── strings.go           # String utilities (slugify, split, trim, etc.)
│   │   └── conv.go              # Type conversion utilities (Atoi, float, etc.)
│   │
│   └── security/
│       ├── hash.go              # Password hashing (bcrypt/argon2)
│       └── jwt.go               # JWT sign/verify utilities
│
├── .air.toml                    # Live reload config
├── .env                         # Local environment variables
├── go.mod
└── Makefile                     # Build, run, lint, test automation
```

Let’s break each part down _professionally AND in a fun way_.

1. The `cmd/` Folder
---------------------

The entry point of your application `**cmd/main.go**` = the place where your entire application starts. In a big Go application you may have:

```
cmd/
  api/
    main.go
  worker/
    main.go
  migrate/
    main.go
```

2. The `internal/` Folder
--------------------------

This is where the fun happens.`internal/` is a folder protected by Go itself. If another module tries to import your code inside `internal/`, Go will say: _“Nope. That’s illegal. Move along.”_ This gives you a powerful way to hide your implementation details.
Inside `internal/` you have:

```
internal/
├── handlers/
├── routes/
└── store/
```

Let’s break each one down.

internal/handlers
-----------------

These are your HTTP handlers — the functions that respond to incoming API calls.

```
func (h *UserHandler) CreateUser(c *gin.Context) { ... }
```

This is where request → response logic lives.

Why not put it in `routes/` or `store/`?

*   Handlers **should not** know database logic
*   Handlers **should not** define routes
*   Handlers **should only** deal with HTTP and business logic

You got this exactly right.

internal/store
--------------

This is where your data access lives. The store talks to your database:

```
func (s *Store) CreateUser(ctx context.Context, user User) error
```

Handlers call the store. The store calls the DB. Clean separation.

internal/routes
---------------

This is where your endpoints get registered.

```
func RegisterUserRoutes(rg *gin.RouterGroup, h *user.Handler) {
    users := rg.Group("/users")
    users.POST("/", h.CreateUser)
}
```

This is beautiful because:

*   Handlers don’t need to know routing paths
*   The API server doesn’t need to know handler internals
*   Routes are the bridge between the server and your handlers

3. The `pkg/` Folder
---------------------

This folder is public. Anything in here can be imported by external modules. Example content might be:

```
pkg/
  validator/
  jwt/
  logger/
  config/
  utils/
  middleware/
```

If it’s reusable or generic → put it here. If it’s business-specific → keep it in `internal`.

4. Makefile
------------

A Makefile automates useful commands:

```
build:
 @go build -o bin/level cmd/main.go
test:
 @go test -v ./...
 
run: build
 @./bin/level
migration:
 @migrate create -ext sql -dir cmd/migrate/migrations $(filter-out $@,$(MAKECMDGOALS))
migrate-up:
 @go run cmd/migrate/main.go up
migrate-down:
 @go run cmd/migrate/main.go down
```

It gives you short commands like:

```
make run
make test
make build
make migrate-up
```

I really love this because it speeds up workflows.

5. .air.toml (Your live reloader)
----------------------------------

`air` is a Go hot-reload tool. While you code, it automatically rebuilds and reloads your server. It duplicates some Makefile roles, but:

### ✔ Air = local development convenience

### ✔ Makefile = automation + CI + DevOps + production

Both are useful.

> **_I’ve also included a bonus simple Todo List project based on the folder structure above. It’s ready to clone and run directly from the_** [**_GitHub repo here_**](https://github.com/Kibaru/go-todo.git)**_._**

Overall — This Structure is Pro Level
-------------------------------------

It’s clean, standard, maintainable, and perfectly Go-coded. This layout will scale when you project grows:

```
internal/
├── users/
│   ├── handler.go
│   ├── store.go
│   ├── routes.go
│   └── model.go
├── products/
│   ├── handler.go
│   ├── store.go
│   ├── routes.go
│   └── model.go
└── auth/
    ├── handler.go
    ├── store.go
    ├── routes.go
    └── model.go
```

Alternatives & When to Use Them
-------------------------------

1. The Clean Architecture
--------------------------

Great for huge teams. Slower for small projects. Included layers:

```
project-name/
│
├── cmd/
│   └── app/
│       └── main.go
│
├── internal/
│   │
│   ├── domain/               ← Entities + Domain logic
│   │   └── todo/
│   │       ├── entity.go
│   │       ├── value_objects.go
│   │       └── repository.go
│   │
│   ├── usecase/              ← Application services / business rules
│   │   └── todo/
│   │       ├── create.go
│   │       ├── list.go
│   │       ├── mark_done.go
│   │       └── delete.go
│   │
│   ├── interface/            ← Adapters (REST handlers, gRPC, CLI)
│   │   └── rest/
│   │       └── todo_handler.go
│   │
│   ├── infrastructure/       ← Frameworks & drivers
│   │   ├── persistence/
│   │   │   └── todo_memory.go
│   │   ├── http/
│   │   │   └── router.go
│   │   ├── config/
│   │   └── logger/
│   │
│   └── bootstrap/            ← Wiring, DI, app startup
│       └── initialize.go
│
├── pkg/                      ← Reusable libraries/utilities
│   ├── middleware/
│   ├── security/
│   ├── response/
│   ├── utils/
│   └── logger/
│
├── go.mod
├── Makefile
└── .env
```

Hexagonal Architecture (Ports & Adapters)
-----------------------------------------

Focuses on separating business logic from external systems. Hexagonal Architecture ≠ Clean Architecture.But they overlap a lot. This architecture is used in many modern Go microservices because it is:

*   framework-independent
*   highly testable
*   extremely easy to extend
*   clear separation of domain vs infrastructure

```
project/
│
├── cmd/
│   └── app/
│       └── main.go
│
├── internal/
│   │
│   ├── domain/                 ← Core business logic
│   │   └── todo/
│   │       ├── entity.go
│   │       ├── service.go
│   │       └── repository.go   (port)
│   │
│   ├── application/            ← Use cases (input ports)
│   │   └── todo/
│   │       ├── create.go
│   │       ├── list.go
│   │       ├── mark_done.go
│   │       └── delete.go
│   │
│   ├── adapters/               ← OUTBOUND adapters (output ports implementations)
│   │   ├── persistence/
│   │   │   ├── todo_memory.go
│   │   │   └── todo_postgres.go
│   │   ├── publisher/
│   │   │   └── todo_event.go
│   │   └── external/
│   │       └── notify_service.go
│   │
│   ├── ports/                  ← INBOUND & OUTBOUND ports
│   │   ├── inbound/
│   │   │   └── todo_port.go    (handler interface)
│   │   └── outbound/
│   │       └── todo_repository.go (repository interface)
│   │
│   ├── transport/              ← INBOUND adapters (REST, gRPC, CLI)
│   │   ├── rest/
│   │   │   ├── todo_handler.go
│   │   │   └── http_router.go
│   │   └── grpc/
│   │       └── todo_service.go
│   │
│   └── bootstrap/              ← Wiring everything together
│       └── wire.go             (or manual DI)
│
├── pkg/
│   ├── logger/
│   ├── utils/
│   ├── middleware/
│   └── response/
│
├── go.mod
├── Makefile
└── .env
```

3. Repository Pattern
----------------------

This structure is simpler than Clean Architecture or Hexagonal — but still very scalable, very testable, and extremely common in real Go projects. Focuses on how to organize your project when you want to isolate persistence logic behind repository interfaces

```
project/
│
├── cmd/
│   └── app/
│       └── main.go
│
├── internal/
│   ├── handlers/              ← HTTP handlers (Gin, Fiber, Chi, etc.)
│   │   └── todo_handler.go
│   │
│   ├── repositories/          ← Interfaces + implementations
│   │   ├── todo_repository.go         (interface)
│   │   ├── todo_memory.go            (in-memory impl)
│   │   ├── todo_postgres.go          (postgres impl)
│   │   └── mocks/                    (mock repos for testing)
│   │       └── todo_repository_mock.go
│   │
│   ├── services/              ← Business logic using repos
│   │   └── todo_service.go
│   │
│   ├── models/                ← Domain/entities
│   │   └── todo.go
│   │
│   ├── routes/
│   │   └── routes.go
│   │
│   └── config/
│       └── config.go
│
├── pkg/                       ← Shared utilities
│   ├── response/
│   ├── logger/
│   └── utils/
│
├── go.mod
└── Makefile
```

You now understand:

. Go’s philosophy
. Each folder’s purpose `internal/` vs `pkg/`
. How handlers, stores, and routes connect
. How to scale with multiple domains
. Alternative architectures and when to choose them.

What’s Next?
------------

[Pointers scares many not to say the least, but after the next tutorial, pointers will fear YOU](https://medium.com/@gitesky14/a-complete-guide-to-go-pointers-b7e020d0566b) 😎.