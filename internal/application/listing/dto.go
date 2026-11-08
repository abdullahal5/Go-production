package listingapp

import "time"

// CreateListingInput is the Data Transfer Object (DTO) containing the payload
// required to create a new listing.
type CreateListingInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	City        string `json:"city"`
}

// ListingOutput is the DTO returned by application use cases to callers
// (such as HTTP handlers or CLI commands).
type ListingOutput struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	City        string    `json:"city"`
	CreatedAt   time.Time `json:"createdAt"`
}

// CreateListingResponse maintains compatibility with the existing API response format.
type CreateListingResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"createdAt"`
}

