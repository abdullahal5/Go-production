package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	listingapp "go-prod-project/internal/application/listing"
	"go-prod-project/internal/domain/listing"
	"go-prod-project/internal/server"
)

// fakeRepository for testing the HTTP layer end-to-end
type fakeRepository struct {
	mu       sync.Mutex
	listings map[string]*listing.Listing
	counter  int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		listings: make(map[string]*listing.Listing),
	}
}

func (r *fakeRepository) Create(_ context.Context, item *listing.Listing) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counter++
	item.SetID(fmt.Sprintf("test-uuid-%d", r.counter))
	item.SetCreatedAt(time.Now().UTC())
	r.listings[item.ID()] = item
	return nil
}

func (r *fakeRepository) GetByID(_ context.Context, id string) (*listing.Listing, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.listings[id]
	if !ok {
		return nil, listing.ErrNotFound
	}
	return item, nil
}

func (r *fakeRepository) List(_ context.Context) ([]*listing.Listing, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make([]*listing.Listing, 0, len(r.listings))
	for _, item := range r.listings {
		res = append(res, item)
	}
	return res, nil
}

func (r *fakeRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.listings[id]; !ok {
		return listing.ErrNotFound
	}
	delete(r.listings, id)
	return nil
}

func setupTestServer() http.Handler {
	repo := newFakeRepository()
	service := listingapp.NewService(repo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := server.NewListingHandler(service, logger)
	return server.NewRouter(handler)
}

func TestHealthCheck(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "OK" {
		t.Errorf("expected body 'OK', got '%s'", rec.Body.String())
	}
}

func TestCreateAndGetListingHTTP(t *testing.T) {
	router := setupTestServer()

	// 1. POST /listings
	payload := map[string]any{
		"title":       "Penthouse Suite",
		"description": "Top floor view",
		"price":       95000,
		"city":        "Dhaka",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/listings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var createResp listingapp.CreateListingResponse
	if err := json.NewDecoder(rec.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if createResp.ID == "" {
		t.Errorf("expected non-empty ID")
	}
	if createResp.Title != "Penthouse Suite" {
		t.Errorf("expected title 'Penthouse Suite', got '%s'", createResp.Title)
	}

	// 2. GET /listings/{id}
	getReq := httptest.NewRequest(http.MethodGet, "/listings/"+createResp.ID, nil)
	getRec := httptest.NewRecorder()

	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", getRec.Code)
	}

	var output listingapp.ListingOutput
	if err := json.NewDecoder(getRec.Body).Decode(&output); err != nil {
		t.Fatalf("failed to decode listing: %v", err)
	}
	if output.ID != createResp.ID {
		t.Errorf("expected ID %s, got %s", createResp.ID, output.ID)
	}

	// 3. DELETE /listings/{id}
	delReq := httptest.NewRequest(http.MethodDelete, "/listings/"+createResp.ID, nil)
	delRec := httptest.NewRecorder()

	router.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 No Content, got %d", delRec.Code)
	}

	// 4. GET /listings/{id} should now be 404
	getAgainReq := httptest.NewRequest(http.MethodGet, "/listings/"+createResp.ID, nil)
	getAgainRec := httptest.NewRecorder()

	router.ServeHTTP(getAgainRec, getAgainReq)

	if getAgainRec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 Not Found, got %d", getAgainRec.Code)
	}
}

func TestValidationErrorsHTTP(t *testing.T) {
	router := setupTestServer()

	t.Run("empty title returns 400 Bad Request", func(t *testing.T) {
		payload := map[string]any{
			"title": "   ",
			"price": 5000,
			"city":  "Dhaka",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/listings", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("negative price returns 400 Bad Request", func(t *testing.T) {
		payload := map[string]any{
			"title": "Valid title",
			"price": -10,
			"city":  "Dhaka",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/listings", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}

