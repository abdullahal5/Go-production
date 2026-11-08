package listing

import (
	"strings"
	"time"
)

// Listing is an Entity and the Aggregate Root for the Listing bounded context.
//
// In Domain-Driven Design (DDD):
// 1. An Entity is defined by its identity (ID) that persists over time.
// 2. An Aggregate Root acts as the gatekeeper for all changes to the aggregate,
//    ensuring all business rules and invariants remain consistent.
// 3. Entity fields are unexported to prevent unauthorized state mutations
//    from outside the domain layer (avoiding the "Anemic Domain Model" anti-pattern).
type Listing struct {
	id          string
	title       string
	description string
	price       Price
	city        string
	createdAt   time.Time
}

// NewListing is the domain factory function used to create a brand new Listing.
// It verifies all business invariants before constructing the entity.
func NewListing(title, description string, price Price, city string) (*Listing, error) {
	cleanTitle := strings.TrimSpace(title)
	if cleanTitle == "" {
		return nil, ErrEmptyTitle
	}
	if len(cleanTitle) > 100 {
		return nil, ErrTitleTooLong
	}

	cleanCity := strings.TrimSpace(city)
	if cleanCity == "" {
		return nil, ErrEmptyCity
	}

	return &Listing{
		title:       cleanTitle,
		description: strings.TrimSpace(description),
		price:       price,
		city:        cleanCity,
		createdAt:   time.Now().UTC(),
	}, nil
}

// Reconstitute restores an existing Listing entity from persistence (e.g. database).
// In DDD, reconstitution bypasses the rules for new entity creation (like generating
// fresh timestamps) because the entity was already validated and stored previously.
func Reconstitute(id, title, description string, price Price, city string, createdAt time.Time) (*Listing, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, ErrInvalidID
	}

	cleanTitle := strings.TrimSpace(title)
	if cleanTitle == "" {
		return nil, ErrEmptyTitle
	}

	cleanCity := strings.TrimSpace(city)
	if cleanCity == "" {
		return nil, ErrEmptyCity
	}

	return &Listing{
		id:          cleanID,
		title:       cleanTitle,
		description: strings.TrimSpace(description),
		price:       price,
		city:        cleanCity,
		createdAt:   createdAt,
	}, nil
}

// --- Domain Getters (Safe, read-only access) ---

// ID returns the unique identifier of the listing.
func (l *Listing) ID() string {
	return l.id
}

// Title returns the listing's title.
func (l *Listing) Title() string {
	return l.title
}

// Description returns the listing's description.
func (l *Listing) Description() string {
	return l.description
}

// Price returns the Price Value Object of the listing.
func (l *Listing) Price() Price {
	return l.price
}

// City returns the city where the listing is located.
func (l *Listing) City() string {
	return l.city
}

// CreatedAt returns the timestamp when the listing was created.
func (l *Listing) CreatedAt() time.Time {
	return l.createdAt
}

// --- Persistence Helpers ---

// SetID allows the persistence layer to assign a database-generated ID.
func (l *Listing) SetID(id string) {
	l.id = id
}

// SetCreatedAt allows the persistence layer to assign the database-authoritative timestamp.
func (l *Listing) SetCreatedAt(t time.Time) {
	l.createdAt = t
}

// --- Domain Behaviors / Business Operations ---

// UpdateDetails updates the listing's information while guarding invariants.
func (l *Listing) UpdateDetails(title, description string, price Price, city string) error {
	cleanTitle := strings.TrimSpace(title)
	if cleanTitle == "" {
		return ErrEmptyTitle
	}
	if len(cleanTitle) > 100 {
		return ErrTitleTooLong
	}

	cleanCity := strings.TrimSpace(city)
	if cleanCity == "" {
		return ErrEmptyCity
	}

	l.title = cleanTitle
	l.description = strings.TrimSpace(description)
	l.price = price
	l.city = cleanCity
	return nil
}
