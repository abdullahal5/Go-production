package server

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	listingapp "go-prod-project/internal/application/listing"
	"go-prod-project/internal/domain/listing"
	"go-prod-project/internal/httpx"
)

// ListingHandler is the primary (driving) HTTP adapter in the Presentation layer.
//
// In Domain-Driven Design (DDD):
// 1. The Presentation layer is responsible for HTTP request handling, input decoding,
//    and output JSON serialization.
// 2. It delegates all business use cases to the Application Service.
// 3. It translates domain/application errors into HTTP response statuses via httpx.
type ListingHandler struct {
	service *listingapp.Service
	logger  *slog.Logger
}

// NewListingHandler constructs a new HTTP listing handler.
func NewListingHandler(service *listingapp.Service, logger *slog.Logger) *ListingHandler {
	return &ListingHandler{
		service: service,
		logger:  logger,
	}
}

// Create handles POST /listings.
func (h *ListingHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Limit request payload to 1MB to prevent DOS attacks.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req listingapp.CreateListingInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Reject trailing unexpected tokens or multiple JSON objects.
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		httpx.Error(w, http.StatusBadRequest, "request must contain one JSON object")
		return
	}

	item, err := h.service.Create(r.Context(), req)
	if err != nil {
		if !isDomainValidationError(err) {
			h.logger.ErrorContext(
				r.Context(),
				"failed to create listing",
				"error", err,
			)
		}

		httpx.HandleError(w, err)
		return
	}

	response := listingapp.CreateListingResponse{
		ID:        item.ID,
		Title:     item.Title,
		CreatedAt: item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	httpx.JSON(w, http.StatusCreated, response)
}

// List handles GET /listings.
func (h *ListingHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to list listings",
			"error", err,
		)
		httpx.HandleError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, items)
}

// GetByID handles GET /listings/{id}.
func (h *ListingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		httpx.Error(w, http.StatusBadRequest, "listing ID is required")
		return
	}

	item, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if !errors.Is(err, listing.ErrNotFound) {
			h.logger.ErrorContext(
				r.Context(),
				"failed to get listing",
				"error", err,
			)
		}

		httpx.HandleError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, item)
}

// Delete handles DELETE /listings/{id}.
func (h *ListingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		httpx.Error(w, http.StatusBadRequest, "listing ID is required")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if !errors.Is(err, listing.ErrNotFound) && !isDomainValidationError(err) {
			h.logger.ErrorContext(
				r.Context(),
				"failed to delete listing",
				"error", err,
			)
		}

		httpx.HandleError(w, err)
		return
	}

	// 204 No Content indicates successful deletion with no body.
	w.WriteHeader(http.StatusNoContent)
}

func isDomainValidationError(err error) bool {
	return errors.Is(err, listing.ErrEmptyTitle) ||
		errors.Is(err, listing.ErrTitleTooLong) ||
		errors.Is(err, listing.ErrEmptyCity) ||
		errors.Is(err, listing.ErrNegativePrice) ||
		errors.Is(err, listing.ErrInvalidID)
}

