package listing

import "fmt"

// Price is a Value Object representing the monetary amount of a listing.
//
// In Domain-Driven Design (DDD):
// 1. Value Objects have NO independent identity (unlike Entities which have an ID).
// 2. They are defined solely by their attributes and are immutable.
// 3. They enforce their own validation invariants upon instantiation.
// 4. Two Value Objects with identical attributes are considered equal.
type Price struct {
	amount int64
}

// NewPrice creates and validates a new Price Value Object.
// It enforces the invariant that price cannot be negative.
func NewPrice(amount int64) (Price, error) {
	if amount < 0 {
		return Price{}, ErrNegativePrice
	}
	return Price{amount: amount}, nil
}

// Amount returns the numerical value of the price.
func (p Price) Amount() int64 {
	return p.amount
}

// Equals checks whether two Price objects are equal by value.
func (p Price) Equals(other Price) bool {
	return p.amount == other.amount
}

func (p Price) String() string {
	return fmt.Sprintf("%d", p.amount)
}

