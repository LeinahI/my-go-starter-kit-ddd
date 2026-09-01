# Common Anti-Patterns and How to Avoid Them

## Entity Anti-Patterns

### Public Fields (Anemic Domain Model)

```go
// WRONG: Public fields allow bypassing invariants
type Order struct {
    ID     uuid.UUID
    Status string
    Items  []Item
}

// Anyone can do: order.Status = "shipped" without validation

// CORRECT: Unexported fields + behavior methods
type Order struct {
    id     uuid.UUID
    status Status
    items  []Item
}

func (o *Order) Ship() error {
    if o.status != StatusConfirmed {
        return ErrCannotShip
    }
    o.status = StatusShipped
    return nil
}
```

### Generic Setters

```go
// WRONG: Setter exposes implementation, bypasses business rules
func (o *Order) SetStatus(s Status) { o.status = s }

// CORRECT: Domain-specific methods that enforce rules
func (o *Order) Place() error   { /* validates draft → placed */ }
func (o *Order) Confirm() error { /* validates placed → confirmed */ }
func (o *Order) Ship() error    { /* validates confirmed → shipped */ }
func (o *Order) Cancel() error  { /* validates not already shipped */ }
```

### Value Receiver for Entities

```go
// WRONG: Value receiver copies the struct, mutations are lost
func (o Order) AddItem(item Item) {
    o.items = append(o.items, item) // Lost!
}

// CORRECT: Pointer receiver for mutable entities
func (o *Order) AddItem(item Item) error {
    o.items = append(o.items, item) // Persisted
    return nil
}
```

---

## Value Object Anti-Patterns

### Mutable Value Objects

```go
// WRONG: Pointer receiver allows mutation
func (m *Money) Add(other Money) {
    m.amount += other.amount // Mutates in place!
}

// CORRECT: Value receiver returns new instance
func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}
```

### Missing Validation

```go
// WRONG: No validation at creation
type Email string // Anyone can do: email := Email("not-an-email")

// CORRECT: Factory function with validation
func NewEmail(s string) (Email, error) {
    if !strings.Contains(s, "@") {
        return "", ErrInvalidEmail
    }
    return Email(strings.ToLower(strings.TrimSpace(s))), nil
}
```

---

## Aggregate Anti-Patterns

### Large Aggregates

```go
// WRONG: Too many entities in one aggregate
type Shop struct {
    id         uuid.UUID
    products   []Product    // Could be thousands
    orders     []Order      // Could be millions
    customers  []Customer   // Separate concern
    reviews    []Review     // Separate concern
}

// CORRECT: Small, focused aggregates
type Shop struct {
    id       uuid.UUID
    name     string
    settings ShopSettings
}
// Products, Orders, Customers are separate aggregates
// referencing Shop by shopID
```

### Cross-Aggregate Object References

```go
// WRONG: Direct pointer to another aggregate
type Order struct {
    customer *Customer  // Tight coupling, can bypass Customer's invariants
}

// CORRECT: Reference by ID
type Order struct {
    customerID uuid.UUID  // Loose coupling
}
```

### Exposing Internal Collections

```go
// WRONG: Returns mutable reference
func (o *Order) Items() []Item {
    return o.items // Caller can modify directly!
}

// CORRECT: Return copy
func (o *Order) Items() []Item {
    items := make([]Item, len(o.items))
    copy(items, o.items)
    return items
}
```

### Modifying Multiple Aggregates in One Transaction

```go
// WRONG: Transaction spans aggregates
func (s *Service) TransferAndNotify(ctx context.Context) error {
    tx := s.db.Begin()
    // Modify order aggregate
    order.Ship()
    // Modify customer aggregate in same tx
    customer.AddLoyaltyPoints(100)
    // Modify inventory aggregate in same tx
    inventory.Reduce(order.Items())
    return tx.Commit() // Three aggregates in one tx!
}

// CORRECT: One aggregate per transaction + events
func (s *Service) ShipOrder(ctx context.Context, orderID uuid.UUID) error {
    err := s.orderRepo.Update(ctx, orderID, func(o *order.Order) error {
        return o.Ship() // Single aggregate
    })
    if err != nil {
        return err
    }

    // Other aggregates updated via domain events
    // OrderShipped → AddLoyaltyPoints (eventual consistency)
    // OrderShipped → ReduceInventory (eventual consistency)
    return s.eventBus.Publish(ctx, order.PullEvents()...)
}
```

---

## Repository Anti-Patterns

### Leaking Database Types

```go
// WRONG: Repository returns DB types
func (r *Repo) Find(ctx context.Context, id uuid.UUID) (*sql.Row, error) {
    return r.db.QueryRowContext(ctx, "SELECT * FROM orders WHERE id = $1", id), nil
}

// CORRECT: Returns domain types
func (r *Repo) Find(ctx context.Context, id uuid.UUID) (*order.Order, error) {
    row := r.db.QueryRowContext(ctx, "SELECT ... FROM orders WHERE id = $1", id)
    // Map to domain object
    return order.ReconstructOrder(/* ... */), nil
}
```

### Generic Repository

```go
// WRONG: One-size-fits-all repository
type Repository[T any] interface {
    Find(ctx context.Context, id uuid.UUID) (T, error)
    Save(ctx context.Context, entity T) error
    Delete(ctx context.Context, id uuid.UUID) error
    FindAll(ctx context.Context) ([]T, error)
    FindBy(ctx context.Context, field string, value any) ([]T, error)
}

// CORRECT: Specific repository per aggregate
type OrderRepository interface {
    Find(ctx context.Context, id uuid.UUID) (*Order, error)
    FindByCustomer(ctx context.Context, customerID uuid.UUID) ([]*Order, error)
    Save(ctx context.Context, order *Order) error
}

type CustomerRepository interface {
    Find(ctx context.Context, id uuid.UUID) (*Customer, error)
    FindByEmail(ctx context.Context, email Email) (*Customer, error)
    Save(ctx context.Context, customer *Customer) error
}
```

### Repository in Wrong Layer

```go
// WRONG: Implementation in domain package
package order

type Repository struct {
    db *sql.DB  // Domain depends on infrastructure!
}

// CORRECT: Interface in domain, implementation in infrastructure
// domain/order/repository.go
package order
type Repository interface { /* ... */ }

// infrastructure/postgres/order_repo.go
package postgres
type OrderRepository struct { db *sql.DB }
```

---

## Domain Service Anti-Patterns

### Stateful Domain Service

```go
// WRONG: Service holds mutable state
type PricingService struct {
    lastPrice Money  // Mutable state!
    cache     map[string]Money
}

// CORRECT: Stateless, only dependencies
type PricingService struct {
    productRepo product.Repository  // Dependency, not state
}
```

### Infrastructure in Domain Service

```go
// WRONG: Domain service calls external API directly
type ShippingService struct {
    httpClient *http.Client  // Infrastructure concern!
}

func (s *ShippingService) CalculateRate() {
    resp, _ := s.httpClient.Get("https://api.shipping.com/rates")
    // ...
}

// CORRECT: Use domain interface, implement in infrastructure
// Domain layer
type ShippingRateProvider interface {
    GetRate(ctx context.Context, weight Weight, destination Address) (Money, error)
}

type ShippingService struct {
    rateProvider ShippingRateProvider  // Domain interface
}

// Infrastructure layer
type APIShippingRateProvider struct {
    httpClient *http.Client
}
```

---

## Domain Event Anti-Patterns

### Publishing from Aggregates

```go
// WRONG: Aggregate publishes events directly
func (o *Order) Place(eventBus *EventBus) error {
    o.status = StatusPlaced
    eventBus.Publish(OrderPlacedEvent{}) // Couples domain to infrastructure
    return nil
}

// CORRECT: Aggregate collects events, application layer publishes
func (o *Order) Place() error {
    o.status = StatusPlaced
    o.events = append(o.events, NewOrderPlacedEvent(o.id))
    return nil
}

// Application layer
func (h *Handler) Handle(ctx context.Context) error {
    order.Place()
    orderRepo.Save(ctx, order)
    eventBus.Publish(ctx, order.PullEvents()...) // Published after persistence
}
```

### Mutable Events

```go
// WRONG: Public fields allow mutation after creation
type OrderPlaced struct {
    OrderID   uuid.UUID
    Total     int64  // Can be changed!
    CreatedAt time.Time
}

// CORRECT: Unexported fields with getters
type OrderPlaced struct {
    orderID   uuid.UUID
    total     Money
    occurredAt time.Time
}

func (e OrderPlaced) OrderID() uuid.UUID  { return e.orderID }
func (e OrderPlaced) Total() Money        { return e.total }
```

---

## Project Structure Anti-Patterns

### Single Models Package

```go
// WRONG: Everything in one package
models/
├── customer.go
├── order.go
├── product.go
└── payment.go  // All 50+ types in one package

// CORRECT: Package per aggregate
domain/
├── customer/
│   ├── customer.go
│   └── repository.go
├── order/
│   ├── order.go
│   ├── item.go
│   └── repository.go
└── product/
    ├── product.go
    └── repository.go
```

### Circular Dependencies

```go
// WRONG: Customer imports Order, Order imports Customer
package customer
import "myapp/domain/order"
func (c *Customer) Orders() []order.Order { /* ... */ }

package order
import "myapp/domain/customer"
func (o *Order) Customer() customer.Customer { /* ... */ }

// CORRECT: Reference by ID, load in application layer
package order
type Order struct {
    customerID uuid.UUID  // ID reference, no import
}

// Application layer loads both and coordinates
```

## Summary Checklist

Before implementing DDD tactical patterns, verify:

- [ ] Entities use unexported fields and pointer receivers
- [ ] Value objects use value receivers and return new instances
- [ ] Aggregate roots enforce all invariants through behavior methods
- [ ] Repositories are interfaces in domain, implementations in infrastructure
- [ ] Domain services are stateless and contain only business logic
- [ ] Domain events are immutable, collected by aggregates, published by application layer
- [ ] Factories validate invariants at creation time
- [ ] Aggregates are small and reference other aggregates by ID
- [ ] No circular dependencies between domain packages
- [ ] `internal/` directory used to prevent external imports
