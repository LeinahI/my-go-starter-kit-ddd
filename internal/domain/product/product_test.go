package product

import (
	"testing"

	"github.com/yourorg/ws/internal/domain/shared"
)

func TestNew_ValidProduct(t *testing.T) {
	price, err := shared.NewMoney("19.99", "PHP")
	if err != nil {
		t.Fatalf("money: %v", err)
	}

	p, err := New("Starter Kit", "starter-kit", "desc", price, 10)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p.Stock() != 10 {
		t.Fatalf("stock = %d, want 10", p.Stock())
	}
}

func TestReserve_InsufficientStock(t *testing.T) {
	price, _ := shared.NewMoney("1.00", "PHP")
	p, err := New("Item", "item", "", price, 2)
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Reserve(3); err != ErrInsufficientStock {
		t.Fatalf("Reserve err = %v, want %v", err, ErrInsufficientStock)
	}
}

func TestReserve_Success(t *testing.T) {
	price, _ := shared.NewMoney("1.00", "PHP")
	p, err := New("Item", "item", "", price, 5)
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Reserve(2); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if p.Stock() != 3 {
		t.Fatalf("stock = %d, want 3", p.Stock())
	}
}
