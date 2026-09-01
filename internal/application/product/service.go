package product

import (
	"context"
	"fmt"

	domain "github.com/yourorg/ws/internal/domain/product"
	"github.com/yourorg/ws/internal/domain/shared"
)

type CreateInput struct {
	Name        string
	Slug        string
	Description string
	Price       string
	Stock       int
}

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.Product, error) {
	price, err := shared.NewMoney(in.Price, "PHP")
	if err != nil {
		return nil, err
	}

	exists, err := s.repo.SlugExists(ctx, in.Slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrDuplicateSlug
	}

	p, err := domain.New(in.Name, in.Slug, in.Description, price, in.Stock)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, p); err != nil {
		return nil, fmt.Errorf("save product: %w", err)
	}
	return p, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Product, error) {
	return s.repo.List(ctx, filter)
}
