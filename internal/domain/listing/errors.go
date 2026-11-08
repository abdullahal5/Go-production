package listing

import "errors"

// Domain errors represent violations of core business rules and invariants.
// In Domain-Driven Design (DDD), domain errors are completely decoupled from
// delivery mechanisms like HTTP status codes or database error codes.
var (
	ErrNotFound      = errors.New("listing not found")
	ErrEmptyTitle    = errors.New("listing title cannot be empty")
	ErrTitleTooLong  = errors.New("listing title cannot exceed 100 characters")
	ErrEmptyCity     = errors.New("listing city cannot be empty")
	ErrNegativePrice = errors.New("listing price cannot be negative")
	ErrInvalidID     = errors.New("listing ID cannot be empty")
)

