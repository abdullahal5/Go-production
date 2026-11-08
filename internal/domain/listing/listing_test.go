package listing_test

import (
	"strings"
	"testing"
	"time"

	"go-prod-project/internal/domain/listing"
)

func TestNewPrice(t *testing.T) {
	t.Run("valid non-negative price", func(t *testing.T) {
		price, err := listing.NewPrice(50000)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if price.Amount() != 50000 {
			t.Errorf("expected amount 50000, got %d", price.Amount())
		}
	})

	t.Run("zero price is valid", func(t *testing.T) {
		price, err := listing.NewPrice(0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if price.Amount() != 0 {
			t.Errorf("expected amount 0, got %d", price.Amount())
		}
	})

	t.Run("negative price is rejected", func(t *testing.T) {
		_, err := listing.NewPrice(-100)
		if err != listing.ErrNegativePrice {
			t.Errorf("expected ErrNegativePrice, got %v", err)
		}
	})

	t.Run("value object equality", func(t *testing.T) {
		p1, _ := listing.NewPrice(2500)
		p2, _ := listing.NewPrice(2500)
		p3, _ := listing.NewPrice(3000)

		if !p1.Equals(p2) {
			t.Errorf("expected p1 to equal p2")
		}
		if p1.Equals(p3) {
			t.Errorf("expected p1 to NOT equal p3")
		}
	})
}

func TestNewListing(t *testing.T) {
	price, _ := listing.NewPrice(25000)

	t.Run("creates valid listing", func(t *testing.T) {
		l, err := listing.NewListing("Modern Apartment", "Spacious 2 bed", price, "Dhaka")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if l.Title() != "Modern Apartment" {
			t.Errorf("expected title 'Modern Apartment', got '%s'", l.Title())
		}
		if l.City() != "Dhaka" {
			t.Errorf("expected city 'Dhaka', got '%s'", l.City())
		}
		if !l.Price().Equals(price) {
			t.Errorf("expected price %v, got %v", price, l.Price())
		}
		if l.CreatedAt().IsZero() {
			t.Errorf("expected non-zero created_at")
		}
	})

	t.Run("rejects empty title", func(t *testing.T) {
		_, err := listing.NewListing("   ", "Description", price, "Dhaka")
		if err != listing.ErrEmptyTitle {
			t.Errorf("expected ErrEmptyTitle, got %v", err)
		}
	})

	t.Run("rejects title exceeding 100 characters", func(t *testing.T) {
		longTitle := strings.Repeat("a", 101)
		_, err := listing.NewListing(longTitle, "Description", price, "Dhaka")
		if err != listing.ErrTitleTooLong {
			t.Errorf("expected ErrTitleTooLong, got %v", err)
		}
	})

	t.Run("rejects empty city", func(t *testing.T) {
		_, err := listing.NewListing("Title", "Description", price, "   ")
		if err != listing.ErrEmptyCity {
			t.Errorf("expected ErrEmptyCity, got %v", err)
		}
	})
}

func TestReconstitute(t *testing.T) {
	price, _ := listing.NewPrice(12000)
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	l, err := listing.Reconstitute("uuid-123", "Studio Flat", "Cozy", price, "Sylhet", fixedTime)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if l.ID() != "uuid-123" {
		t.Errorf("expected ID 'uuid-123', got '%s'", l.ID())
	}
	if !l.CreatedAt().Equal(fixedTime) {
		t.Errorf("expected timestamp %v, got %v", fixedTime, l.CreatedAt())
	}
}

