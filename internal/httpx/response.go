package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-prod-project/internal/domain/listing"
)

// ErrorResponse represents a standardized JSON error envelope.
type ErrorResponse struct {
	Error string `json:"error"`
}

// JSON serializes the provided data to JSON and sets the HTTP status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	_ = json.NewEncoder(w).Encode(data)
}

// Error formats and writes a standardized JSON error response.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{
		Error: message,
	})
}

// HandleError translates domain-level errors into corresponding HTTP status codes.
// This preserves separation of concerns: domain errors are decoupled from HTTP status codes.
func HandleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, listing.ErrNotFound):
		Error(w, http.StatusNotFound, err.Error())

	case errors.Is(err, listing.ErrEmptyTitle),
		errors.Is(err, listing.ErrTitleTooLong),
		errors.Is(err, listing.ErrEmptyCity),
		errors.Is(err, listing.ErrNegativePrice),
		errors.Is(err, listing.ErrInvalidID):
		Error(w, http.StatusBadRequest, err.Error())

	default:
		Error(w, http.StatusInternalServerError, "internal server error")
	}
}
