# Entities, Value Objects, and Factories

## Entities

Entities are domain objects with unique identity that persists throughout their lifecycle. Two entities with identical attributes but different IDs are considered distinct.

### Key Principles

- **Identity over attributes** - Equality based on ID, not field values
- **Mutable state** - Use pointer receivers for state-changing methods
- **Unexported fields** - Prevent external invariant violations
- **Factory creation** - Enforce invariants at construction time
- **Temporal tracking** - Include `createdAt`/`updatedAt` timestamps

### Full Example

```go
package order

import (
    "errors"
    "time"
    "github.com/google/uuid"
)

type Status string

const (
    StatusDraft     Status = "draft"
    StatusPlaced    Status = "placed"
    StatusConfirmed Status = "confirmed"
    StatusCancelled Status = "cancelled"
)

var (
    ErrOrderNotDraft   = errors.New("order is not in draft status")
    ErrEmptyOrder      = errors.New("order must have at least one item")
    ErrAlreadyCancelled = errors.New("order is already cancelled")
)

// Entity with unique identity
type Order struct {
    id         uuid.UUID
    customerID uuid.UUID
    status     Status
    items      []Item
    events     []Event
    createdAt  time.Time
    updatedAt  time.Time
}

// Factory function - enforces invariants
func NewOrder(customerID uuid.UUID) (*Order, error) {
    if customerID == uuid.Nil {
        return nil, errors.New("customer ID is required")
    }

    now := time.Now()
    return &Order{
        id:         uuid.New(),
        customerID: customerID,
        status:     StatusDraft,
        items:      make([]Item, 0),
        createdAt:  now,
        updatedAt:  now,
    }, nil
}

// Reconstitution factory - used by repository to load from DB
// Bypasses business validation since data is already validated
func ReconstructOrder(
    id, customerID uuid.UUID,
    status Status,
    items []Item,
    createdAt, updatedAt time.Time,
) *Order {
    return &Order{
        id:         id,
        customerID: customerID,
        status:     status,
        items:      items,
        createdAt:  createdAt,
        updatedAt:  updatedAt,
    }
}

// Behavior methods (pointer receiver for mutation)
func (o *Order) AddItem(productID uuid.UUID, qty int, price Money) error {
    if o.status != StatusDraft {
        return ErrOrderNotDraft
    }
    if qty <= 0 {
        return errors.New("quantity must be positive")
    }

    item := NewItem(productID, qty, price)
    o.items = append(o.items, item)
    o.updatedAt = time.Now()
    return nil
}

func (o *Order) Place() error {
    if len(o.items) == 0 {
        return ErrEmptyOrder
    }
    if o.status != StatusDraft {
        return ErrOrderNotDraft
    }

    o.status = StatusPlaced
    o.updatedAt = time.Now()
    o.events = append(o.events, NewOrderPlacedEvent(o.id, o.Total()))
    return nil
}

func (o *Order) Cancel(reason string) error {
    if o.status == StatusCancelled {
        return ErrAlreadyCancelled
    }

    o.status = StatusCancelled
    o.updatedAt = time.Now()
    o.events = append(o.events, NewOrderCancelledEvent(o.id, reason))
    return nil
}

// Getters (expose read access)
func (o *Order) ID() uuid.UUID         { return o.id }
func (o *Order) CustomerID() uuid.UUID { return o.customerID }
func (o *Order) Status() Status        { return o.status }
func (o *Order) CreatedAt() time.Time  { return o.createdAt }
func (o *Order) UpdatedAt() time.Time  { return o.updatedAt }

// Return copy of items to prevent external mutation
func (o *Order) Items() []Item {
    items := make([]Item, len(o.items))
    copy(items, o.items)
    return items
}

func (o *Order) ItemCount() int { return len(o.items) }

func (o *Order) Total() Money {
    total := Money{}
    for _, item := range o.items {
        total, _ = total.Add(item.Subtotal())
    }
    return total
}

// Event collection
func (o *Order) PullEvents() []Event {
    events := o.events
    o.events = nil
    return events
}
```

### Child Entity (Item within Order aggregate)

```go
package order

import "github.com/google/uuid"

// Child entity - local identity within aggregate
type Item struct {
    id        uuid.UUID
    productID uuid.UUID
    quantity  int
    unitPrice Money
}

func NewItem(productID uuid.UUID, qty int, price Money) Item {
    return Item{
        id:        uuid.New(),
        productID: productID,
        quantity:  qty,
        unitPrice: price,
    }
}

func (i Item) ID() uuid.UUID        { return i.id }
func (i Item) ProductID() uuid.UUID { return i.productID }
func (i Item) Quantity() int         { return i.quantity }
func (i Item) UnitPrice() Money      { return i.unitPrice }

func (i Item) Subtotal() Money {
    return Money{
        amount:   i.unitPrice.amount * int64(i.quantity),
        currency: i.unitPrice.currency,
    }
}
```

---

## Value Objects

Value Objects represent domain concepts defined entirely by their attributes. They have no identity and are immutable.

### Key Principles

- **Immutability** - Never modify; return new instances
- **Value receivers** - Use `T` not `*T`
- **Self-validating** - Factory function validates on creation
- **Equality by value** - Two value objects with same attributes are equal
- **Small size** - Keep structs small since they are copied by value

### Simple Value Object (Type Alias)

```go
package order

import (
    "errors"
    "strings"
)

type Email string

var ErrInvalidEmail = errors.New("invalid email format")

func NewEmail(s string) (Email, error) {
    s = strings.TrimSpace(s)
    if s == "" || !strings.Contains(s, "@") {
        return "", ErrInvalidEmail
    }
    return Email(strings.ToLower(s)), nil
}

func (e Email) String() string { return string(e) }

func (e Email) Domain() string {
    parts := strings.SplitN(string(e), "@", 2)
    if len(parts) != 2 {
        return ""
    }
    return parts[1]
}

func (e Email) Equals(other Email) bool {
    return e == other
}
```

### Complex Value Object (Struct)

```go
package order

import "errors"

var (
    ErrInvalidCurrency  = errors.New("currency is required")
    ErrNegativeAmount   = errors.New("amount cannot be negative")
    ErrCurrencyMismatch = errors.New("currency mismatch")
)

type Money struct {
    amount   int64  // Store in smallest unit (cents)
    currency string
}

func NewMoney(amount int64, currency string) (Money, error) {
    if currency == "" {
        return Money{}, ErrInvalidCurrency
    }
    if amount < 0 {
        return Money{}, ErrNegativeAmount
    }
    return Money{amount: amount, currency: currency}, nil
}

// Zero value factory
func ZeroMoney(currency string) Money {
    return Money{amount: 0, currency: currency}
}

// Value receivers - immutable operations return new instances
func (m Money) Amount() int64    { return m.amount }
func (m Money) Currency() string { return m.currency }

func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

func (m Money) Subtract(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, ErrCurrencyMismatch
    }
    result := m.amount - other.amount
    if result < 0 {
        return Money{}, ErrNegativeAmount
    }
    return Money{amount: result, currency: m.currency}, nil
}

func (m Money) Multiply(factor int) Money {
    return Money{amount: m.amount * int64(factor), currency: m.currency}
}

func (m Money) IsZero() bool {
    return m.amount == 0
}

func (m Money) GreaterThan(other Money) bool {
    return m.currency == other.currency && m.amount > other.amount
}

func (m Money) Equals(other Money) bool {
    return m.amount == other.amount && m.currency == other.currency
}
```

### Address Value Object

```go
package customer

import "errors"

type Address struct {
    street  string
    city    string
    state   string
    zipCode string
    country string
}

func NewAddress(street, city, state, zipCode, country string) (Address, error) {
    if street == "" || city == "" || country == "" {
        return Address{}, errors.New("street, city, and country are required")
    }
    return Address{
        street:  street,
        city:    city,
        state:   state,
        zipCode: zipCode,
        country: country,
    }, nil
}

func (a Address) Street() string  { return a.street }
func (a Address) City() string    { return a.city }
func (a Address) State() string   { return a.state }
func (a Address) ZipCode() string { return a.zipCode }
func (a Address) Country() string { return a.country }

func (a Address) Equals(other Address) bool {
    return a == other
}

func (a Address) FullAddress() string {
    return a.street + ", " + a.city + ", " + a.state + " " + a.zipCode + ", " + a.country
}
```

---

## Factories

Factories encapsulate complex object creation logic and enforce invariants at construction time.

### Simple Factory Function

The standard Go `NewX()` pattern. Use for straightforward construction.

```go
func NewCustomer(email Email, name string) (*Customer, error) {
    if name == "" {
        return nil, errors.New("name is required")
    }
    return &Customer{
        id:        uuid.New(),
        email:     email,
        name:      name,
        status:    StatusActive,
        createdAt: time.Now(),
    }, nil
}
```

### Functional Options Pattern

Use when objects have many optional parameters.

```go
type OrderOption func(*Order)

func WithDiscount(discount Discount) OrderOption {
    return func(o *Order) { o.discount = &discount }
}

func WithNotes(notes string) OrderOption {
    return func(o *Order) { o.notes = notes }
}

func WithShippingAddress(addr Address) OrderOption {
    return func(o *Order) { o.shippingAddr = addr }
}

func NewOrder(customerID uuid.UUID, opts ...OrderOption) (*Order, error) {
    if customerID == uuid.Nil {
        return nil, errors.New("customer ID required")
    }

    order := &Order{
        id:         uuid.New(),
        customerID: customerID,
        status:     StatusDraft,
        items:      make([]Item, 0),
        createdAt:  time.Now(),
    }

    for _, opt := range opts {
        opt(order)
    }

    return order, nil
}

// Usage
order, err := NewOrder(
    customerID,
    WithDiscount(seasonalDiscount),
    WithShippingAddress(homeAddr),
    WithNotes("Gift wrap please"),
)
```

### Builder Pattern

Use for step-by-step construction with complex validation.

```go
type OrderBuilder struct {
    order Order
    err   error
}

func Build() *OrderBuilder {
    return &OrderBuilder{
        order: Order{id: uuid.New(), items: make([]Item, 0)},
    }
}

func (b *OrderBuilder) ForCustomer(id uuid.UUID) *OrderBuilder {
    if b.err != nil {
        return b
    }
    if id == uuid.Nil {
        b.err = errors.New("customer ID required")
        return b
    }
    b.order.customerID = id
    return b
}

func (b *OrderBuilder) AddItem(item Item) *OrderBuilder {
    if b.err != nil {
        return b
    }
    b.order.items = append(b.order.items, item)
    return b
}

func (b *OrderBuilder) Finish() (*Order, error) {
    if b.err != nil {
        return nil, b.err
    }
    if b.order.customerID == uuid.Nil {
        return nil, errors.New("customer ID required")
    }
    if len(b.order.items) == 0 {
        return nil, errors.New("at least one item required")
    }
    b.order.status = StatusDraft
    b.order.createdAt = time.Now()
    return &b.order, nil
}

// Usage
order, err := Build().
    ForCustomer(customerID).
    AddItem(item1).
    AddItem(item2).
    Finish()
```

### Reconstitution Factory

Separate factory for loading from persistence. Bypasses business validation since data was already validated when originally created.

```go
// Used by repository implementations only
func ReconstructOrder(
    id, customerID uuid.UUID,
    status Status,
    items []Item,
    createdAt, updatedAt time.Time,
) *Order {
    return &Order{
        id:         id,
        customerID: customerID,
        status:     status,
        items:      items,
        createdAt:  createdAt,
        updatedAt:  updatedAt,
    }
}
```

### When to Use Which Factory

| Pattern | Use When |
|---------|----------|
| `NewX()` | Simple creation with few required params |
| Functional Options | Many optional params, extensible API |
| Builder | Step-by-step construction, complex validation |
| `ReconstructX()` | Loading from database/external source |
