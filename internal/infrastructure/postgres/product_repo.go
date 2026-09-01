package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yourorg/ws/internal/domain/product"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Save(ctx context.Context, p *product.Product) error {
	const q = `
		INSERT INTO products (id, name, slug, description, price, stock, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::numeric, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, q,
		p.ID(),
		p.Name(),
		p.Slug(),
		nullString(p.Description()),
		p.Price().Amount(),
		p.Stock(),
		p.CreatedAt(),
		p.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert product: %w", err)
	}
	return nil
}

func (r *ProductRepository) FindByID(ctx context.Context, id string) (*product.Product, error) {
	const q = `
		SELECT id, name, slug, COALESCE(description, ''), price::text, stock, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	return r.scanOne(ctx, q, id)
}

func (r *ProductRepository) FindBySlug(ctx context.Context, slug string) (*product.Product, error) {
	const q = `
		SELECT id, name, slug, COALESCE(description, ''), price::text, stock, created_at, updated_at
		FROM products
		WHERE slug = $1
	`
	return r.scanOne(ctx, q, slug)
}

func (r *ProductRepository) List(ctx context.Context, filter product.ListFilter) ([]*product.Product, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	where := "WHERE 1=1"
	args := []any{}
	argN := 1

	if q := strings.TrimSpace(filter.Query); q != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR slug ILIKE $%d)", argN, argN)
		args = append(args, "%"+q+"%")
		argN++
	}

	query := fmt.Sprintf(`
		SELECT id, name, slug, COALESCE(description, ''), price::text, stock, created_at, updated_at
		FROM products
		%s
		ORDER BY name ASC
		LIMIT $%d
	`, where, argN)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var out []*product.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProductRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1)`, slug).Scan(&exists)
	return exists, err
}

func (r *ProductRepository) scanOne(ctx context.Context, query string, arg any) (*product.Product, error) {
	row := r.db.QueryRowContext(ctx, query, arg)
	p, err := scanProduct(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, product.ErrNotFound
	}
	return p, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProduct(row rowScanner) (*product.Product, error) {
	var (
		id, name, slug, description, price string
		stock                              int
		createdAt, updatedAt               time.Time
	)
	if err := row.Scan(&id, &name, &slug, &description, &price, &stock, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	return product.Reconstruct(id, name, slug, description, price, "PHP", stock, createdAt, updatedAt)
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
