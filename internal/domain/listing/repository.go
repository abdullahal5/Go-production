package listing

import "context"

// Repository defines the outbound port (interface) for Listing persistence.
//
// In Domain-Driven Design (DDD):
// 1. The Repository interface belongs to the DOMAIN layer, defining what the domain needs.
// 2. Concrete implementations (PostgreSQL, in-memory, Mongo, etc.) belong to the INFRASTRUCTURE layer.
// 3. This implements the Dependency Inversion Principle: high-level domain logic does
//    not depend on low-level database drivers; both depend on this abstraction.
type Repository interface {
	Create(ctx context.Context, item *Listing) error
	GetByID(ctx context.Context, id string) (*Listing, error)
	List(ctx context.Context) ([]*Listing, error)
	Delete(ctx context.Context, id string) error
}

