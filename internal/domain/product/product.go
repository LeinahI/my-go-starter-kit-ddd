package product

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ws/internal/domain/shared"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Product struct {
	id          string
	name        string
	slug        string
	description string
	price       shared.Money
	stock       int
	createdAt   time.Time
	updatedAt   time.Time
}

func New(name, slug, description string, price shared.Money, stock int) (*Product, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" {
		return nil, ErrInvalidName
	}
	if !slugPattern.MatchString(slug) {
		return nil, ErrInvalidSlug
	}
	if stock < 0 {
		return nil, ErrInvalidStock
	}

	now := time.Now().UTC()
	return &Product{
		id:          uuid.NewString(),
		name:        name,
		slug:        slug,
		description: strings.TrimSpace(description),
		price:       price,
		stock:       stock,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// Reconstruct loads a product from persistence without re-running creation rules.
func Reconstruct(id, name, slug, description, priceAmount, priceCurrency string, stock int, createdAt, updatedAt time.Time) (*Product, error) {
	price, err := shared.NewMoney(priceAmount, priceCurrency)
	if err != nil {
		return nil, err
	}
	return &Product{
		id:          id,
		name:        name,
		slug:        slug,
		description: description,
		price:       price,
		stock:       stock,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

func (p *Product) Reserve(qty int) error {
	if qty < 1 {
		return ErrInvalidStock
	}
	if qty > p.stock {
		return ErrInsufficientStock
	}
	p.stock -= qty
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Product) ID() string              { return p.id }
func (p *Product) Name() string            { return p.name }
func (p *Product) Slug() string            { return p.slug }
func (p *Product) Description() string     { return p.description }
func (p *Product) Price() shared.Money     { return p.price }
func (p *Product) Stock() int              { return p.stock }
func (p *Product) CreatedAt() time.Time    { return p.createdAt }
func (p *Product) UpdatedAt() time.Time    { return p.updatedAt }

type ListFilter struct {
	Query string
	Limit int
}

type Repository interface {
	Save(ctx context.Context, p *Product) error
	FindByID(ctx context.Context, id string) (*Product, error)
	FindBySlug(ctx context.Context, slug string) (*Product, error)
	List(ctx context.Context, filter ListFilter) ([]*Product, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
}
