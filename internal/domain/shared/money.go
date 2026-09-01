package shared

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var pricePattern = regexp.MustCompile(`^\d+(?:\.\d{1,2})?$`)

var (
	ErrInvalidPrice    = errors.New("invalid price")
	ErrCurrencyMissing = errors.New("currency is required")
)

// Money is a value object. Use string amounts at the HTTP boundary; convert here.
type Money struct {
	amount   string
	currency string
}

func NewMoney(amount, currency string) (Money, error) {
	amount = strings.TrimSpace(amount)
	currency = strings.TrimSpace(currency)
	if currency == "" {
		return Money{}, ErrCurrencyMissing
	}
	if !pricePattern.MatchString(amount) {
		return Money{}, fmt.Errorf("%w: %q", ErrInvalidPrice, amount)
	}
	return Money{amount: amount, currency: currency}, nil
}

func (m Money) Amount() string   { return m.amount }
func (m Money) Currency() string { return m.currency }

func (m Money) String() string {
	return m.amount + " " + m.currency
}
