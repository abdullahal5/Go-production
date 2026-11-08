package listingapp_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	listingapp "go-prod-project/internal/application/listing"
	"go-prod-project/internal/domain/listing"
)

// inMemoryRepository is an in-memory fake implementing listing.Repository.
// In DDD, having domain repository ports makes testing application use cases
// incredibly fast and isolated without needing a real database or Docker container.
type inMemoryRepository struct {
	mu       sync.RWMutex
	listings map[string]*listing.Listing
	counter  int
}

func newInMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{
		listings: make(map[string]*listing.Listing),
	}
}

func (r *inMemoryRepository) Create(_ context.Context, item *listing.Listing) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counter++
	item.SetID(fmt.Sprintf("generated-uuid-%d", r.counter))
	item.SetCreatedAt(time.Now().UTC())

	r.listings[item.ID()] = item
	return nil
}

func (r *inMemoryRepository) GetByID(_ context.Context, id string) (*listing.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, exists := r.listings[id]
	if !exists {
		return nil, listing.ErrNotFound
	}
	return item, nil
}

func (r *inMemoryRepository) List(_ context.Context) ([]*listing.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make([]*listing.Listing, 0, len(r.listings))
	for _, item := range r.listings {
		results = append(results, item)
	}
	return results, nil
}

func (r *inMemoryRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.listings[id]; !exists {
		return listing.ErrNotFound
	}

	delete(r.listings, id)
	return nil
}

func TestService_CreateAndGet(t *testing.T) {
	repo := newInMemoryRepository()
	svc := listingapp.NewService(repo)
	ctx := context.Background()

	t.Run("successfully creates listing", func(t *testing.T) {
		out, err := svc.Create(ctx, listingapp.CreateListingInput{
			Title:       "Penthouse with View",
			Description: "Luxurious penthouse",
			Price:       75000,
			City:        "Dhaka",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if out.ID == "" {
			t.Errorf("expected generated ID")
		}
		if out.Title != "Penthouse with View" {
			t.Errorf("expected title 'Penthouse with View', got '%s'", out.Title)
		}
		if out.Price != 75000 {
			t.Errorf("expected price 75000, got %d", out.Price)
		}

		// Verify retrieval
		fetched, err := svc.GetByID(ctx, out.ID)
		if err != nil {
			t.Fatalf("expected fetch to succeed, got %v", err)
		}
		if fetched.ID != out.ID {
			t.Errorf("expected ID %s, got %s", out.ID, fetched.ID)
		}
	})

	t.Run("rejects invalid input via domain invariants", func(t *testing.T) {
		_, err := svc.Create(ctx, listingapp.CreateListingInput{
			Title: "",
			Price: 1000,
			City:  "Dhaka",
		})
		if err != listing.ErrEmptyTitle {
			t.Errorf("expected ErrEmptyTitle, got %v", err)
		}

		_, err = svc.Create(ctx, listingapp.CreateListingInput{
			Title: "Valid Title",
			Price: -50,
			City:  "Dhaka",
		})
		if err != listing.ErrNegativePrice {
			t.Errorf("expected ErrNegativePrice, got %v", err)
		}
	})
}

func TestService_Delete(t *testing.T) {
	repo := newInMemoryRepository()
	svc := listingapp.NewService(repo)
	ctx := context.Background()

	out, _ := svc.Create(ctx, listingapp.CreateListingInput{
		Title: "To Be Deleted",
		Price: 1000,
		City:  "Chittagong",
	})

	if err := svc.Delete(ctx, out.ID); err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}

	_, err := svc.GetByID(ctx, out.ID)
	if err != listing.ErrNotFound {
		t.Errorf("expected ErrNotFound after deletion, got %v", err)
	}
}

