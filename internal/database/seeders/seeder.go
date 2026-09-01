package seeder

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/yourorg/ws/internal/config"
	"github.com/yourorg/ws/internal/domain/product"
	"github.com/yourorg/ws/internal/domain/shared"
)

type Seeder interface {
	Name() string
	Run(ctx context.Context) error
}

type Deps struct {
	DB        *sql.DB
	SeederCfg *config.SeederConfig
}

func RunAll(ctx context.Context, deps Deps) error {
	seeders := []Seeder{
		NewProductSeeder(deps),
	}

	for _, s := range seeders {
		if err := s.Run(ctx); err != nil {
			return fmt.Errorf("seeder %s: %w", s.Name(), err)
		}
	}
	return nil
}

type productSeeder struct {
	db  *sql.DB
	cfg *config.SeederConfig
}

func NewProductSeeder(deps Deps) Seeder {
	return &productSeeder{db: deps.DB, cfg: deps.SeederCfg}
}

func (s *productSeeder) Name() string { return "products" }

func (s *productSeeder) Run(ctx context.Context) error {
	for _, item := range s.cfg.DemoProducts {
		var exists bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1)`, item.Slug).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}

		price, err := shared.NewMoney(item.Price, "PHP")
		if err != nil {
			return err
		}

		p, err := product.New(item.Name, item.Slug, item.Description, price, item.Stock)
		if err != nil {
			return err
		}

		const q = `
			INSERT INTO products (id, name, slug, description, price, stock, created_at, updated_at)
			VALUES ($1, $2, $3, NULLIF($4, ''), $5::numeric, $6, $7, $8)
		`
		_, err = s.db.ExecContext(ctx, q,
			p.ID(), p.Name(), p.Slug(), p.Description(), p.Price().Amount(), p.Stock(), p.CreatedAt(), p.UpdatedAt(),
		)
		if err != nil {
			return fmt.Errorf("insert %s: %w", item.Slug, err)
		}
	}
	return nil
}
