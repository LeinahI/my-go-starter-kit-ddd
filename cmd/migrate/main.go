package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const migrationsDir = "internal/database/migrations"

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	command := os.Args[1]

	switch command {
	case "create":
		if len(os.Args) < 3 {
			log.Fatal("usage: go run ./cmd/migrate create <name>")
		}
		if err := createMigration(migrationsDir, os.Args[2]); err != nil {
			log.Fatalf("create migration: %v", err)
		}
		return
	}

	db, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	provider, err := newProvider(db)
	if err != nil {
		log.Fatal(err)
	}

	switch command {
	case "up":
		if _, err := provider.Up(ctx); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
	case "down":
		if _, err := provider.Down(ctx); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
	case "status":
		if err := printStatus(ctx, provider); err != nil {
			log.Fatalf("migrate status: %v", err)
		}
	case "version":
		version, err := provider.GetDBVersion(ctx)
		if err != nil {
			log.Fatalf("migrate version: %v", err)
		}
		fmt.Println(version)
	case "fresh":
		if _, err := provider.DownTo(ctx, 0); err != nil {
			log.Fatalf("migrate fresh (reset): %v", err)
		}
		if _, err := provider.Up(ctx); err != nil {
			log.Fatalf("migrate fresh (up): %v", err)
		}
		log.Println("migrate fresh completed")
	case "rollback":
		steps := 1
		if len(os.Args) >= 3 {
			var parseErr error
			steps, parseErr = strconv.Atoi(os.Args[2])
			if parseErr != nil || steps < 1 {
				log.Fatal("usage: go run ./cmd/migrate rollback <steps>")
			}
		}
		for i := 0; i < steps; i++ {
			if _, err := provider.Down(ctx); err != nil {
				log.Fatalf("migrate rollback step %d/%d: %v", i+1, steps, err)
			}
		}
		log.Printf("rolled back %d migration(s)", steps)
	default:
		usage()
	}
}

func newProvider(db *sql.DB) (*goose.Provider, error) {
	return goose.NewProvider(
		goose.DialectPostgres,
		db,
		migrationDirFS{dir: migrationsDir},
	)
}

func printStatus(ctx context.Context, provider *goose.Provider) error {
	statuses, err := provider.Status(ctx)
	if err != nil {
		return err
	}
	log.Println("    Applied At                  Migration")
	log.Println("    =======================================")
	for _, status := range statuses {
		appliedAt := "Pending"
		if status.State == goose.StateApplied && !status.AppliedAt.IsZero() {
			appliedAt = status.AppliedAt.Local().Format(time.ANSIC)
		}
		log.Printf("    %-24s -- %s", appliedAt, status.Source.Path)
	}
	return nil
}

func openDB() (*sql.DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return db, nil
}

func usage() {
	fmt.Println(`goose migration runner

usage:
  go run ./cmd/migrate up                 apply all pending migrations
  go run ./cmd/migrate down               roll back one migration
  go run ./cmd/migrate fresh              reset all migrations, then up
  go run ./cmd/migrate rollback [steps]   roll back N migrations (default: 1)
  go run ./cmd/migrate status             show migration status
  go run ./cmd/migrate version            print current DB version
  go run ./cmd/migrate create <name>      stub YYYYMMDD_HHMMSS_<name>.sql migration file`)
	os.Exit(1)
}
