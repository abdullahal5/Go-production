package httpx

import (
	"encoding/json"
	"net/http"
)

type Code string

const (
	// General errors
	CodeInternalError      Code = "internal_error"
	CodeInvalidRequest     Code = "invalid_request"
	CodeInvalidInput       Code = "invalid_input"
	CodeValidationFailed   Code = "validation_failed"
	CodeNotFound           Code = "not_found"
	CodeMethodNotAllowed   Code = "method_not_allowed"
	CodeUnauthorized      Code = "unauthorized"
	CodeForbidden         Code = "forbidden"
	CodeConflict          Code = "conflict"
	CodeTooManyRequests   Code = "too_many_requests"
	CodeServiceUnavailable Code = "service_unavailable"

	// ID and resource errors
	CodeInvalidID         Code = "invalid_id"
	CodeResourceNotFound  Code = "resource_not_found"
	CodeDuplicateResource Code = "duplicate_resource"

	// Authentication errors
	CodeInvalidCredentials Code = "invalid_credentials"
	CodeInvalidToken       Code = "invalid_token"
	CodeTokenExpired       Code = "token_expired"
	CodeTokenRevoked       Code = "token_revoked"
	CodeAccountDisabled    Code = "account_disabled"

	// Authorization errors
	CodeInsufficientPermissions Code = "insufficient_permissions"

	// Database errors
	CodeDatabaseError Code = "database_error"

	// Request errors
	CodeMissingField Code = "missing_field"
	CodeInvalidField Code = "invalid_field"
	CodePayloadTooLarge Code = "payload_too_large"
	CodeUnsupportedMediaType Code = "unsupported_media_type"
)

type errorEnvelop struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    Code `json:"code"`
	Message string `json:"message"`
}

func Error(w http.ResponseWriter, status int, message string, code Code) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(errorEnvelop{
		Error: errorPayload{
			Code:    code,
			Message: message,
		},
	})
}
