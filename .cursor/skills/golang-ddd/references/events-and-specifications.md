# Domain Events and Specifications

## Domain Events

Domain Events represent significant occurrences within the domain. They are immutable records of things that happened.

### Key Principles

- **Immutability** - Events are facts; they cannot change after creation
- **Past tense naming** - `OrderPlaced`, `PaymentReceived` (not `PlaceOrder`)
- **Self-contained** - Include all data subscribers need
- **Created by aggregates, published by application layer**
- **Support both sync and async handling**

### Event Interface

```go
package shared

import (
    "time"
    "github.com/google/uuid"
)

// Base event interface
type Event interface {
    EventName() string
    OccurredAt() time.Time
    AggregateID() uuid.UUID
}
```

### Concrete Domain Events

```go
package order

import (
    "time"
    "github.com/google/uuid"
)

// OrderPlaced event - immutable
type OrderPlaced struct {
    orderID    uuid.UUID
    customerID uuid.UUID
    total      Money
    occurredAt time.Time
}

func NewOrderPlacedEvent(orderID, customerID uuid.UUID, total Money) OrderPlaced {
    return OrderPlaced{
        orderID:    orderID,
        customerID: customerID,
        total:      total,
        occurredAt: time.Now(),
    }
}

func (e OrderPlaced) EventName() string        { return "order.placed" }
func (e OrderPlaced) OccurredAt() time.Time    { return e.occurredAt }
func (e OrderPlaced) AggregateID() uuid.UUID   { return e.orderID }
func (e OrderPlaced) CustomerID() uuid.UUID    { return e.customerID }
func (e OrderPlaced) Total() Money             { return e.total }

// OrderCancelled event
type OrderCancelled struct {
    orderID    uuid.UUID
    reason     string
    occurredAt time.Time
}

func NewOrderCancelledEvent(orderID uuid.UUID, reason string) OrderCancelled {
    return OrderCancelled{
        orderID:    orderID,
        reason:     reason,
        occurredAt: time.Now(),
    }
}

func (e OrderCancelled) EventName() string      { return "order.cancelled" }
func (e OrderCancelled) OccurredAt() time.Time  { return e.occurredAt }
func (e OrderCancelled) AggregateID() uuid.UUID { return e.orderID }
func (e OrderCancelled) Reason() string         { return e.reason }
```

### Collecting Events in Aggregates

```go
package order

type Order struct {
    id     uuid.UUID
    status Status
    events []shared.Event  // Collected events
}

func (o *Order) Place() error {
    if o.status != StatusDraft {
        return ErrOrderNotDraft
    }

    o.status = StatusPlaced
    // Collect event - don't publish yet
    o.events = append(o.events, NewOrderPlacedEvent(o.id, o.customerID, o.Total()))
    return nil
}

// Pull and clear events
func (o *Order) PullEvents() []shared.Event {
    events := o.events
    o.events = nil
    return events
}
```

### In-Process Event Bus

```go
package eventbus

import (
    "context"
    "sync"
    "myapp/internal/domain/shared"
)

type Handler func(ctx context.Context, event shared.Event) error

type EventBus struct {
    mu       sync.RWMutex
    handlers map[string][]Handler
}

func New() *EventBus {
    return &EventBus{
        handlers: make(map[string][]Handler),
    }
}

func (eb *EventBus) Subscribe(eventName string, handler Handler) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    eb.handlers[eventName] = append(eb.handlers[eventName], handler)
}

func (eb *EventBus) Publish(ctx context.Context, events ...shared.Event) error {
    eb.mu.RLock()
    defer eb.mu.RUnlock()

    for _, event := range events {
        handlers := eb.handlers[event.EventName()]
        for _, handler := range handlers {
            if err := handler(ctx, event); err != nil {
                return err
            }
        }
    }
    return nil
}
```

### Application Layer Publishing

```go
package application

type PlaceOrderHandler struct {
    orderRepo order.Repository
    eventBus  *eventbus.EventBus
}

func (h *PlaceOrderHandler) Handle(ctx context.Context, cmd PlaceOrderCommand) error {
    o, err := h.orderRepo.Find(ctx, cmd.OrderID)
    if err != nil {
        return err
    }

    // Execute domain logic
    if err := o.Place(); err != nil {
        return err
    }

    // Persist aggregate
    if err := h.orderRepo.Save(ctx, o); err != nil {
        return err
    }

    // Publish events after successful persistence
    return h.eventBus.Publish(ctx, o.PullEvents()...)
}
```

### Event Handlers (Subscribers)

```go
package application

// Send notification when order is placed
func NewOrderNotificationHandler(notifier Notifier) eventbus.Handler {
    return func(ctx context.Context, event shared.Event) error {
        e, ok := event.(order.OrderPlaced)
        if !ok {
            return nil
        }
        return notifier.SendOrderConfirmation(ctx, e.CustomerID(), e.AggregateID())
    }
}

// Update read model when order is placed
func NewOrderReadModelHandler(readStore OrderReadStore) eventbus.Handler {
    return func(ctx context.Context, event shared.Event) error {
        e, ok := event.(order.OrderPlaced)
        if !ok {
            return nil
        }
        return readStore.AddOrder(ctx, OrderView{
            ID:         e.AggregateID(),
            CustomerID: e.CustomerID(),
            Total:      e.Total().Amount(),
            PlacedAt:   e.OccurredAt(),
        })
    }
}
```

### Wiring It Together

```go
func main() {
    bus := eventbus.New()

    // Register handlers
    bus.Subscribe("order.placed", NewOrderNotificationHandler(emailNotifier))
    bus.Subscribe("order.placed", NewOrderReadModelHandler(readStore))
    bus.Subscribe("order.cancelled", NewRefundHandler(paymentService))

    // Create application handlers
    placeOrderHandler := &PlaceOrderHandler{
        orderRepo: orderRepo,
        eventBus:  bus,
    }
}
```

---

## Specifications

Specifications encapsulate business rules as composable, reusable objects. They answer the question: "Does this object satisfy a particular criteria?"

### Key Principles

- **Single responsibility** - Each specification tests one rule
- **Composability** - Combine with AND, OR, NOT operators
- **Reusability** - Same spec for validation, filtering, and querying
- **Type safety** - Use Go generics (1.21+)

### Generic Specification Interface

```go
package specification

// Core interface using generics
type Specification[T any] interface {
    IsSatisfiedBy(candidate T) bool
}
```

### Composite Specifications

```go
package specification

// AND - all specs must be satisfied
type andSpec[T any] struct {
    specs []Specification[T]
}

func And[T any](specs ...Specification[T]) Specification[T] {
    return &andSpec[T]{specs: specs}
}

func (s *andSpec[T]) IsSatisfiedBy(candidate T) bool {
    for _, spec := range s.specs {
        if !spec.IsSatisfiedBy(candidate) {
            return false
        }
    }
    return true
}

// OR - at least one spec must be satisfied
type orSpec[T any] struct {
    specs []Specification[T]
}

func Or[T any](specs ...Specification[T]) Specification[T] {
    return &orSpec[T]{specs: specs}
}

func (s *orSpec[T]) IsSatisfiedBy(candidate T) bool {
    for _, spec := range s.specs {
        if spec.IsSatisfiedBy(candidate) {
            return true
        }
    }
    return false
}

// NOT - inverts a spec
type notSpec[T any] struct {
    spec Specification[T]
}

func Not[T any](spec Specification[T]) Specification[T] {
    return &notSpec[T]{spec: spec}
}

func (s *notSpec[T]) IsSatisfiedBy(candidate T) bool {
    return !s.spec.IsSatisfiedBy(candidate)
}
```

### Function-Based Specification

For quick, inline specifications.

```go
package specification

// SpecFunc wraps a function as a Specification
type SpecFunc[T any] func(T) bool

func (f SpecFunc[T]) IsSatisfiedBy(candidate T) bool {
    return f(candidate)
}

// Usage
activeCustomer := specification.SpecFunc[Customer](func(c Customer) bool {
    return c.IsActive()
})
```

### Domain-Specific Specifications

```go
package customer

import "myapp/internal/domain/shared/specification"

// Premium customer specification
type PremiumCustomerSpec struct {
    minSpent        int64
    minMemberMonths int
}

func NewPremiumCustomerSpec(minSpent int64, minMonths int) *PremiumCustomerSpec {
    return &PremiumCustomerSpec{
        minSpent:        minSpent,
        minMemberMonths: minMonths,
    }
}

func (s *PremiumCustomerSpec) IsSatisfiedBy(c *Customer) bool {
    return c.TotalSpent() >= s.minSpent &&
        c.MembershipMonths() >= s.minMemberMonths
}

// Discount eligibility specification
type DiscountEligibleSpec struct {
    minOrders int
}

func NewDiscountEligibleSpec(minOrders int) *DiscountEligibleSpec {
    return &DiscountEligibleSpec{minOrders: minOrders}
}

func (s *DiscountEligibleSpec) IsSatisfiedBy(c *Customer) bool {
    return c.OrderCount() >= s.minOrders && !c.HasOverduePayments()
}

// Compose specifications
func VIPCustomerSpec() specification.Specification[*Customer] {
    return specification.And[*Customer](
        NewPremiumCustomerSpec(10000, 24),
        NewDiscountEligibleSpec(10),
        specification.Not[*Customer](
            &BlacklistedSpec{},
        ),
    )
}
```

### Using Specifications

```go
// Validation
vipSpec := customer.VIPCustomerSpec()
if vipSpec.IsSatisfiedBy(cust) {
    cust.ApplyVIPDiscount()
}

// Filtering a collection
func Filter[T any](items []T, spec specification.Specification[T]) []T {
    var result []T
    for _, item := range items {
        if spec.IsSatisfiedBy(item) {
            result = append(result, item)
        }
    }
    return result
}

eligibleCustomers := Filter(allCustomers, customer.VIPCustomerSpec())
```

### Specifications for Database Queries

Extend specifications to generate query conditions.

```go
package specification

// QuerySpec can generate SQL conditions
type QuerySpec interface {
    ToSQL() (clause string, args []any)
}

// Combined specification + query
type CustomerEmailSpec struct {
    email string
}

func (s *CustomerEmailSpec) IsSatisfiedBy(c *Customer) bool {
    return string(c.Email()) == s.email
}

func (s *CustomerEmailSpec) ToSQL() (string, []any) {
    return "email = ?", []any{s.email}
}

// AND query spec
type AndQuerySpec struct {
    specs []QuerySpec
}

func (s *AndQuerySpec) ToSQL() (string, []any) {
    clauses := make([]string, 0, len(s.specs))
    args := make([]any, 0)

    for _, spec := range s.specs {
        clause, specArgs := spec.ToSQL()
        clauses = append(clauses, clause)
        args = append(args, specArgs...)
    }

    return "(" + strings.Join(clauses, " AND ") + ")", args
}
```

### When to Use Specifications

| Use Case | Example |
|----------|---------|
| **Validation** | Check if customer is eligible for promotion |
| **Filtering** | Select orders matching criteria from a list |
| **Querying** | Build dynamic database queries |
| **Policy enforcement** | Verify business rules before operations |

### When NOT to Use Specifications

- Simple boolean checks that won't be reused
- Checks that are only used in one place
- Performance-critical hot paths (interface dispatch adds overhead)
