package listingapp

import (
	"context"
	"strings"

	"go-prod-project/internal/domain/listing"
)

// Service implements the application use cases for listings.
//
// In Domain-Driven Design (DDD):
// 1. The Application Service orchestrates domain objects, aggregates, and repository ports.
// 2. It contains NO core business logic (invariants are enforced inside the Domain entities).
// 3. It converts incoming DTOs into domain objects, executes operations, and returns output DTOs.
// 4. It acts as the boundary between external interfaces (HTTP/gRPC/CLI) and the domain model.
type Service struct {
	repo listing.Repository
}

// NewService constructs a new listing application service.
func NewService(repo listing.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// Create executes the use case of creating a new listing.
func (s *Service) Create(ctx context.Context, input CreateListingInput) (*ListingOutput, error) {
	// Step 1: Instantiate Price Value Object (enforces non-negative invariant).
	price, err := listing.NewPrice(input.Price)
	if err != nil {
		return nil, err
	}

	// Step 2: Instantiate Listing Aggregate Root (enforces title/city invariants).
	item, err := listing.NewListing(input.Title, input.Description, price, input.City)
	if err != nil {
		return nil, err
	}

	// Step 3: Persist via the Domain Repository Port.
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	// Step 4: Map domain aggregate to an output DTO.
	return toOutput(item), nil
}

// GetByID executes the use case of finding a listing by its ID.
func (s *Service) GetByID(ctx context.Context, id string) (*ListingOutput, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, listing.ErrInvalidID
	}

	item, err := s.repo.GetByID(ctx, cleanID)
	if err != nil {
		return nil, err
	}

	return toOutput(item), nil
}

// List executes the use case of listing recent listings.
func (s *Service) List(ctx context.Context) ([]*ListingOutput, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]*ListingOutput, 0, len(items))
	for _, item := range items {
		outputs = append(outputs, toOutput(item))
	}

	return outputs, nil
}

// Delete executes the use case of deleting a listing.
func (s *Service) Delete(ctx context.Context, id string) error {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return listing.ErrInvalidID
	}

	return s.repo.Delete(ctx, cleanID)
}

func toOutput(l *listing.Listing) *ListingOutput {
	return &ListingOutput{
		ID:          l.ID(),
		Title:       l.Title(),
		Description: l.Description(),
		Price:       l.Price().Amount(),
		City:        l.City(),
		CreatedAt:   l.CreatedAt(),
	}
}

