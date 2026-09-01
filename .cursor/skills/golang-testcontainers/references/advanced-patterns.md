# Advanced Testcontainers Patterns

## PostgreSQL Snapshot/Restore for Test Isolation

Instead of recreating containers per test, use PostgreSQL snapshots for fast resets:

```go
package repository_test

import (
    "context"
    "testing"

    "github.com/jackc/pgx/v5"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupPostgresWithSnapshot(t *testing.T) (*postgres.PostgresContainer, string) {
    t.Helper()
    ctx := context.Background()

    ctr, err := postgres.Run(ctx, "postgres:16-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        postgres.BasicWaitStrategies(),
        postgres.WithSQLDriver("pgx"),
    )
    testcontainers.CleanupContainer(t, ctr)
    require.NoError(t, err)

    // Run migrations
    _, _, err = ctr.Exec(ctx, []string{
        "psql", "-U", "test", "-d", "testdb", "-c",
        "CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL, email TEXT UNIQUE NOT NULL)",
    })
    require.NoError(t, err)

    // Create snapshot after migrations
    err = ctr.Snapshot(ctx)
    require.NoError(t, err)

    dbURL, err := ctr.ConnectionString(ctx)
    require.NoError(t, err)

    return ctr, dbURL
}

func TestWithSnapshots(t *testing.T) {
    ctr, dbURL := setupPostgresWithSnapshot(t)
    ctx := context.Background()

    t.Run("insert user", func(t *testing.T) {
        t.Cleanup(func() {
            err := ctr.Restore(ctx)
            require.NoError(t, err)
        })

        conn, err := pgx.Connect(ctx, dbURL)
        require.NoError(t, err)
        defer conn.Close(ctx)

        _, err = conn.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "Alice", "alice@test.com")
        require.NoError(t, err)

        var count int
        err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
        require.NoError(t, err)
        require.Equal(t, 1, count)
    })

    t.Run("empty after restore", func(t *testing.T) {
        t.Cleanup(func() {
            err := ctr.Restore(ctx)
            require.NoError(t, err)
        })

        conn, err := pgx.Connect(ctx, dbURL)
        require.NoError(t, err)
        defer conn.Close(ctx)

        var count int
        err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
        require.NoError(t, err)
        require.Equal(t, 0, count) // Clean slate after restore
    })
}
```

### Snapshot vs Other Isolation Approaches

| Approach | Speed | Isolation | Trade-off |
|----------|-------|-----------|-----------|
| New container per test | Slow (10-30s) | Complete | Best isolation, worst performance |
| Snapshot/Restore | Fast (~100ms) | Complete | Postgres-specific feature |
| Transaction rollback | Fastest | Partial | Cannot test commit behavior |
| TestMain + truncate | Fast | Good | Manual cleanup required |

## Kafka Integration Tests

```go
package messaging_test

import (
    "context"
    "testing"

    "github.com/IBM/sarama"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/kafka"
)

func setupKafka(t *testing.T) []string {
    t.Helper()
    ctx := context.Background()

    ctr, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0",
        kafka.WithClusterID("test-cluster"),
    )
    testcontainers.CleanupContainer(t, ctr)
    require.NoError(t, err)

    brokers, err := ctr.Brokers(ctx)
    require.NoError(t, err)

    return brokers
}

func TestKafkaProduceConsume(t *testing.T) {
    brokers := setupKafka(t)

    config := sarama.NewConfig()
    config.Producer.Return.Successes = true

    // Create topic
    admin, err := sarama.NewClusterAdmin(brokers, config)
    require.NoError(t, err)
    defer admin.Close()

    err = admin.CreateTopic("test-topic", &sarama.TopicDetail{
        NumPartitions:     1,
        ReplicationFactor: 1,
    }, false)
    require.NoError(t, err)

    // Produce
    producer, err := sarama.NewSyncProducer(brokers, config)
    require.NoError(t, err)
    defer producer.Close()

    _, _, err = producer.SendMessage(&sarama.ProducerMessage{
        Topic: "test-topic",
        Value: sarama.StringEncoder("test-message"),
    })
    require.NoError(t, err)

    // Consume
    consumer, err := sarama.NewConsumer(brokers, config)
    require.NoError(t, err)
    defer consumer.Close()

    partConsumer, err := consumer.ConsumePartition("test-topic", 0, sarama.OffsetOldest)
    require.NoError(t, err)
    defer partConsumer.Close()

    msg := <-partConsumer.Messages()
    require.Equal(t, "test-message", string(msg.Value))
}
```

## Multi-Container Networks

When services need to communicate with each other:

```go
package integration_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
    "github.com/testcontainers/testcontainers-go/network"
)

func setupMultiContainer(t *testing.T) (dbEndpoint, redisEndpoint string) {
    t.Helper()
    ctx := context.Background()

    // Create shared network
    net, err := network.New(ctx)
    require.NoError(t, err)
    t.Cleanup(func() { net.Remove(ctx) })

    networkName := net.Name

    // Start PostgreSQL
    pgCtr, err := postgres.Run(ctx, "postgres:16-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        postgres.BasicWaitStrategies(),
        testcontainers.WithNetwork(networkName),
    )
    testcontainers.CleanupContainer(t, pgCtr)
    require.NoError(t, err)

    // Start Redis
    redisCtr, err := tcredis.Run(ctx, "redis:7",
        testcontainers.WithNetwork(networkName),
    )
    testcontainers.CleanupContainer(t, redisCtr)
    require.NoError(t, err)

    dbEndpoint, err = pgCtr.ConnectionString(ctx)
    require.NoError(t, err)

    redisEndpoint, err = redisCtr.Endpoint(ctx, "")
    require.NoError(t, err)

    return dbEndpoint, redisEndpoint
}

func TestAppWithMultipleServices(t *testing.T) {
    dbEndpoint, redisEndpoint := setupMultiContainer(t)

    app := NewApp(dbEndpoint, redisEndpoint)
    // Test full application stack with real infrastructure
    _ = app
}
```

## Custom Container with Init Scripts

```go
func setupPostgresWithMigrations(t *testing.T) *sql.DB {
    t.Helper()
    ctx := context.Background()

    ctr, err := postgres.Run(ctx, "postgres:16-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        postgres.WithInitScripts(
            filepath.Join("testdata", "001_schema.sql"),
            filepath.Join("testdata", "002_seed.sql"),
        ),
        postgres.BasicWaitStrategies(),
    )
    testcontainers.CleanupContainer(t, ctr)
    require.NoError(t, err)

    connStr, err := ctr.ConnectionString(ctx)
    require.NoError(t, err)

    db, err := sql.Open("pgx", connStr)
    require.NoError(t, err)
    t.Cleanup(func() { db.Close() })

    return db
}
```

## Helper Pattern: Reusable Test Infrastructure

Create a shared `testinfra` package for your project:

```go
// internal/testinfra/containers.go
package testinfra

import (
    "context"
    "database/sql"
    "testing"

    "github.com/redis/go-redis/v9"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// PostgresDB creates a PostgreSQL container and returns a connected *sql.DB.
func PostgresDB(t *testing.T, initScripts ...string) *sql.DB {
    t.Helper()
    ctx := context.Background()

    opts := []testcontainers.ContainerCustomizer{
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        postgres.BasicWaitStrategies(),
    }
    if len(initScripts) > 0 {
        opts = append(opts, postgres.WithInitScripts(initScripts...))
    }

    ctr, err := postgres.Run(ctx, "postgres:16-alpine", opts...)
    testcontainers.CleanupContainer(t, ctr)
    require.NoError(t, err)

    connStr, err := ctr.ConnectionString(ctx)
    require.NoError(t, err)

    db, err := sql.Open("pgx", connStr)
    require.NoError(t, err)
    t.Cleanup(func() { db.Close() })

    return db
}

// RedisClient creates a Redis container and returns a connected client.
func RedisClient(t *testing.T) *redis.Client {
    t.Helper()
    ctx := context.Background()

    ctr, err := tcredis.Run(ctx, "redis:7")
    testcontainers.CleanupContainer(t, ctr)
    require.NoError(t, err)

    endpoint, err := ctr.Endpoint(ctx, "")
    require.NoError(t, err)

    client := redis.NewClient(&redis.Options{Addr: endpoint})
    t.Cleanup(func() { client.Close() })

    return client
}
```

Usage in tests:

```go
package order_test

import (
    "testing"
    "yourapp/internal/testinfra"
)

func TestOrderService(t *testing.T) {
    db := testinfra.PostgresDB(t, "testdata/schema.sql", "testdata/seed.sql")
    cache := testinfra.RedisClient(t)

    svc := NewOrderService(db, cache)
    // Test with real infrastructure
    _ = svc
}
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Integration Tests
on: [push, pull_request]

jobs:
  integration-test:
    runs-on: ubuntu-latest # Docker pre-installed

    steps:
    - uses: actions/checkout@v4

    - uses: actions/setup-go@v5
      with:
        go-version: '1.22'

    - name: Run Integration Tests
      run: go test -race -tags=integration -coverprofile=coverage.out ./...

    - name: Check Coverage
      run: |
        go tool cover -func=coverage.out | grep total | awk '{print $3}' | \
        awk -F'%' '{if ($1 < 80) exit 1}'

    - name: Cleanup Containers
      if: always()
      run: docker container prune -f
```

### Makefile Targets

```makefile
.PHONY: test test-integration test-all

test:
	go test -race ./...

test-integration:
	go test -race -tags=integration -timeout=5m ./...

test-all:
	go test -race -tags=integration -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
```

## Performance Tips

1. **Share containers** - Use `TestMain` or package-level setup for expensive containers
2. **Use snapshots** - PostgreSQL `Snapshot()`/`Restore()` is faster than recreating
3. **Alpine images** - Prefer `postgres:16-alpine`, `redis:7-alpine` for smaller images
4. **Parallel packages** - `go test ./...` runs packages in parallel by default
5. **Build tags** - Separate slow integration tests so unit tests run fast
6. **Docker layer caching** - Images cached locally after first pull
