// ws API
//
//	@title			ws API
//	@version		1.0
//	@description	Backend API for core and site frontends.
//
//	@host		localhost:8080
//	@BasePath	/
//
//	@schemes	http
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/yourorg/ws/docs" // swag-generated OpenAPI docs

	"github.com/yourorg/ws/internal/config"
	"github.com/yourorg/ws/internal/database"
	transport "github.com/yourorg/ws/internal/transport/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	router := transport.NewRouter(cfg, db)

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", cfg.Addr())
		if cfg.IsDevelopment() {
			log.Printf("swagger UI: http://%s/swagger/index.html", cfg.Addr())
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}
