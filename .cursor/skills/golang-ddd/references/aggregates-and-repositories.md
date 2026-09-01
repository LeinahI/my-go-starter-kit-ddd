# Aggregates, Repositories, and Domain Services

## Aggregates

Aggregates are clusters of entities and value objects that form a consistency boundary. The Aggregate Root is the single entry point for all modifications.

### Key Principles

- **Single entry point** - All changes go through the aggregate root
- **Consistency boundary** - Invariants enforced within a single aggregate are always consistent
- **Small aggregates** - Only group entities that share invariants
- **Reference by ID** - Other aggregates reference by ID, not direct pointers
- **Package per aggregate** - Each aggregate lives in its own Go package

### Full Aggregate Example

```go
package customer

import (
    "errors"
    "time"
    "github.com/google/uuid"
)

var (
    ErrCustomerDeactivated = errors.New("customer is deactivated")
    ErrMaxAddresses        = errors.New("maximum addresses reached")
    ErrAddressNotFound     = errors.New("address not found")
    ErrHasActiveOrders     = errors.New("cannot deactivate customer with active orders")
)

const maxAddresses = 5

// Aggregate Root
type Customer struct {
    id          uuid.UUID
    email       Email
    name        string
    addresses   []Address   // Child value objects
    status      Status
    orderCount  int         // Derived from external data
    events      []Event
    createdAt   time.Time
    updatedAt   time.Time
}

// Factory - only way to create a new Customer
func NewCustomer(email Email, name string) (*Customer, error) {
    if name == "" {
        return nil, errors.New("name is required")
    }

    now := time.Now()
    return &Customer{
        id:        uuid.New(),
        email:     email,
        name:      name,
        addresses: make([]Address, 0),
        status:    StatusActive,
        createdAt: now,
        updatedAt: now,
    }, nil
}

// --- Behavior methods enforce invariants ---

func (c *Customer) ChangeEmail(newEmail Email) error {
    if c.status == StatusDeactivated {
        return ErrCustomerDeactivated
    }

    c.email = newEmail
    c.updatedAt = time.Now()
    c.events = append(c.events, NewEmailChangedEvent(c.id, newEmail))
    return nil
}

func (c *Customer) AddAddress(addr Address) error {
    if c.status == StatusDeactivated {
        return ErrCustomerDeactivated
    }
    if len(c.addresses) >= maxAddresses {
        return ErrMaxAddresses
    }

    c.addresses = append(c.addresses, addr)
    c.updatedAt = time.Now()
    return nil
}

func (c *Customer) RemoveAddress(index int) error {
    if index < 0 || index >= len(c.addresses) {
        return ErrAddressNotFound
    }

    c.addresses = append(c.addresses[:index], c.addresses[index+1:]...)
    c.updatedAt = time.Now()
    return nil
}

func (c *Customer) Deactivate() error {
    if c.status == StatusDeactivated {
        return nil // Idempotent
    }

    c.status = StatusDeactivated
    c.updatedAt = time.Now()
    c.events = append(c.events, NewCustomerDeactivatedEvent(c.id))
    return nil
}

// Getters
func (c *Customer) ID() uuid.UUID        { return c.id }
func (c *Customer) Email() Email          { return c.email }
func (c *Customer) Name() string          { return c.name }
func (c *Customer) Status() Status        { return c.status }
func (c *Customer) IsActive() bool        { return c.status == StatusActive }
func (c *Customer) CreatedAt() time.Time  { return c.createdAt }
func (c *Customer) AddressCount() int     { return len(c.addresses) }

// Return copy to prevent external mutation
func (c *Customer) Addresses() []Address {
    addrs := make([]Address, len(c.addresses))
    copy(addrs, c.addresses)
    return addrs
}

// Event collection
func (c *Customer) PullEvents() []Event {
    events := c.events
    c.events = nil
    return events
}
```

### Aggregate Design Rules

1. **Design small aggregates** - Prefer single-entity aggregates when possible
2. **Protect invariants** - The root must validate every state change
3. **No cross-aggregate references** - Use IDs to reference other aggregates
4. **One transaction per aggregate** - Don't modify multiple aggregates in one transaction
5. **Eventual consistency between aggregates** - Use domain events for cross-aggregate communication

### Cross-Aggregate Reference Example

```go
// WRONG: Direct reference to another aggregate
type Order struct {
    customer *Customer  // Don't do this
}

// CORRECT: Reference by ID
type Order struct {
    customerID uuid.UUID  // Reference by identity
}
```

---

## Repositories

Repositories provide an interface for persisting and retrieving aggregates. The interface is defined in the domain layer; implementations live in infrastructure.

### Key Principles

- **Interface in domain** - Define alongside the aggregate root
- **One per aggregate** - Each aggregate root has its own repository
- **Domain types only** - Accept and return domain objects, not DB models
- **Domain errors** - Return domain-specific errors, not database errors
- **Context propagation** - Always accept `context.Context` as first parameter
- **Update pattern** - Use callback for safe read-modify-write operations

### Repository Interface

```go
package customer

import (
    "context"
    "github.com/google/uuid"
)

var (
    ErrCustomerNotFound = errors.New("customer not found")
    ErrDuplicateEmail   = errors.New("email already exists")
)

// Repository interface - defined in domain layer
type Repository interface {
    // Query
    Find(ctx context.Context, id uuid.UUID) (*Customer, error)
    FindByEmail(ctx context.Context, email Email) (*Customer, error)

    // Command
    Save(ctx context.Context, customer *Customer) error
    Delete(ctx context.Context, id uuid.UUID) error

    // Safe update pattern (read-modify-write in transaction)
    Update(ctx context.Context, id uuid.UUID, fn func(*Customer) error) error
}
```

### Infrastructure Implementation

```go
package postgres

import (
    "context"
    "database/sql"
    "myapp/internal/domain/customer"
    "github.com/google/uuid"
)

type CustomerRepository struct {
    db *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
    return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Find(ctx context.Context, id uuid.UUID) (*customer.Customer, error) {
    row := r.db.QueryRowContext(ctx,
        `SELECT id, email, name, status, created_at, updated_at
         FROM customers WHERE id = $1`, id)

    var data customerRow
    if err := row.Scan(&data.ID, &data.Email, &data.Name, &data.Status, &data.CreatedAt, &data.UpdatedAt); err != nil {
        if err == sql.ErrNoRows {
            return nil, customer.ErrCustomerNotFound // Domain error
        }
        return nil, err
    }

    // Load addresses
    addresses, err := r.loadAddresses(ctx, id)
    if err != nil {
        return nil, err
    }

    // Reconstitute domain object
    return customer.ReconstructCustomer(
        data.ID, customer.Email(data.Email), data.Name,
        addresses, customer.Status(data.Status),
        data.CreatedAt, data.UpdatedAt,
    ), nil
}

func (r *CustomerRepository) Save(ctx context.Context, c *customer.Customer) error {
    _, err := r.db.ExecContext(ctx,
        `INSERT INTO customers (id, email, name, status, created_at, updated_at)
         VALUES ($1, $2, $3, $4, $5, $6)
         ON CONFLICT (id) DO UPDATE SET
            email = $2, name = $3, status = $4, updated_at = $6`,
        c.ID(), c.Email(), c.Name(), c.Status(), c.CreatedAt(), c.UpdatedAt())

    if err != nil {
        if isUniqueViolation(err, "customers_email_key") {
            return customer.ErrDuplicateEmail
        }
        return err
    }

    return r.saveAddresses(ctx, c.ID(), c.Addresses())
}

// Transaction-safe update pattern
func (r *CustomerRepository) Update(ctx context.Context, id uuid.UUID, fn func(*customer.Customer) error) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Load within transaction
    c, err := r.findInTx(ctx, tx, id)
    if err != nil {
        return err
    }

    // Apply domain logic
    if err := fn(c); err != nil {
        return err
    }

    // Persist
    if err := r.saveInTx(ctx, tx, c); err != nil {
        return err
    }

    return tx.Commit()
}

// Internal DB model (not exported)
type customerRow struct {
    ID        uuid.UUID
    Email     string
    Name      string
    Status    string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Update Pattern Usage

```go
// In application layer
func (h *ChangeEmailHandler) Handle(ctx context.Context, cmd ChangeEmailCommand) error {
    return h.customerRepo.Update(ctx, cmd.CustomerID, func(c *customer.Customer) error {
        email, err := customer.NewEmail(cmd.NewEmail)
        if err != nil {
            return err
        }
        return c.ChangeEmail(email)
    })
}
```

---

## Domain Services

Domain Services contain business logic that doesn't naturally belong to a single entity or value object. They are stateless and operate across multiple aggregates.

### Key Principles

- **Stateless** - No mutable instance state; only dependencies
- **Domain logic only** - Business rules, not orchestration or infrastructure concerns
- **Multi-aggregate coordination** - When logic spans aggregate boundaries
- **Named after domain concepts** - Use ubiquitous language

### Domain Service vs Application Service

| Aspect | Domain Service | Application Service |
|--------|----------------|---------------------|
| Contains | Pure business logic | Orchestration, transaction management |
| Dependencies | Domain types, repositories | Domain services, infrastructure |
| Layer | Domain | Application |
| Example | `TransferService.Transfer()` | `TransferHandler.Handle()` |

### Domain Service Example

```go
package transfer

import (
    "context"
    "errors"
    "myapp/internal/domain/account"
)

var (
    ErrSameAccount       = errors.New("cannot transfer to same account")
    ErrInsufficientFunds = errors.New("insufficient funds")
    ErrCurrencyMismatch  = errors.New("accounts must share currency")
)

// Domain Service - pure business logic
type MoneyTransferService struct {
    accountRepo account.Repository
}

func NewMoneyTransferService(repo account.Repository) *MoneyTransferService {
    return &MoneyTransferService{accountRepo: repo}
}

func (s *MoneyTransferService) Transfer(
    ctx context.Context,
    fromID, toID uuid.UUID,
    amount account.Money,
) error {
    if fromID == toID {
        return ErrSameAccount
    }

    from, err := s.accountRepo.Find(ctx, fromID)
    if err != nil {
        return err
    }

    to, err := s.accountRepo.Find(ctx, toID)
    if err != nil {
        return err
    }

    // Domain rule: currencies must match
    if from.Currency() != to.Currency() {
        return ErrCurrencyMismatch
    }

    // Domain rule: sufficient funds
    if err := from.Withdraw(amount); err != nil {
        return err
    }

    if err := to.Deposit(amount); err != nil {
        return err
    }

    // Persist both (application layer handles transaction)
    if err := s.accountRepo.Save(ctx, from); err != nil {
        return err
    }

    return s.accountRepo.Save(ctx, to)
}
```

### Pure Function Alternative

When domain logic is simple and doesn't need dependencies, use a pure function instead of a service struct.

```go
package pricing

// Pure function - no dependencies needed
func CalculateDiscount(orderTotal Money, customerTier Tier) Money {
    rate := discountRate(customerTier)
    return orderTotal.Multiply(rate)
}

func discountRate(tier Tier) float64 {
    switch tier {
    case TierGold:
        return 0.15
    case TierSilver:
        return 0.10
    case TierBronze:
        return 0.05
    default:
        return 0
    }
}
```

### Application Service (for comparison)

The application service orchestrates and delegates to domain services.

```go
package application

// Application Service - orchestration layer
type TransferHandler struct {
    transferService *transfer.MoneyTransferService
    eventBus        *eventbus.EventBus
    txManager       TransactionManager
}

func (h *TransferHandler) Handle(ctx context.Context, cmd TransferCommand) error {
    // Parse command into domain types
    amount, err := account.NewMoney(cmd.Amount, cmd.Currency)
    if err != nil {
        return err
    }

    // Execute within transaction (infrastructure concern)
    return h.txManager.Execute(ctx, func(ctx context.Context) error {
        // Delegate to domain service
        if err := h.transferService.Transfer(ctx, cmd.FromID, cmd.ToID, amount); err != nil {
            return err
        }

        // Publish events (infrastructure concern)
        return h.eventBus.Publish(ctx, NewTransferCompletedEvent(cmd.FromID, cmd.ToID, amount))
    })
}
```
