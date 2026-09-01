package product

import "errors"

var (
	ErrNotFound          = errors.New("product not found")
	ErrInvalidName       = errors.New("invalid product name")
	ErrInvalidSlug       = errors.New("invalid product slug")
	ErrInvalidStock      = errors.New("invalid stock")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrDuplicateSlug     = errors.New("product slug already exists")
)
