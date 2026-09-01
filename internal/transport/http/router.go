package http

import (
	"database/sql"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/yourorg/ws/internal/config"
	appproduct "github.com/yourorg/ws/internal/application/product"
	"github.com/yourorg/ws/internal/infrastructure/postgres"
	"github.com/yourorg/ws/internal/transport/http/middleware"
)

func NewRouter(cfg *config.Config, db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	productRepo := postgres.NewProductRepository(db)
	productSvc := appproduct.NewService(productRepo)
	products := NewProductHandler(productSvc)

	mux.HandleFunc("GET /up", Health(db))
	mux.HandleFunc("GET /api/v1/products", products.List)
	mux.HandleFunc("GET /api/v1/products/{id}", products.Get)
	mux.HandleFunc("POST /api/v1/products", products.Create)

	if cfg.IsDevelopment() {
		mux.Handle("/swagger/", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		))
	}

	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.CORS(cfg, h)
	return h
}
