package main

import (
	"context"
	"log"
	"time"

	"github.com/yourorg/ws/internal/config"
	"github.com/yourorg/ws/internal/database"
	seeder "github.com/yourorg/ws/internal/database/seeders"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	seederCfg, err := config.LoadSeederConfig(config.DefaultSeederConfigPath)
	if err != nil {
		log.Fatalf("load seeder config: %v", err)
	}

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := seeder.RunAll(ctx, seeder.Deps{
		DB:        db,
		SeederCfg: seederCfg,
	}); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	log.Println("seed completed")
}
