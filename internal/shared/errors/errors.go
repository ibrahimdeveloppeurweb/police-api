package errors

import (
	"errors"
	"net/http"
)

var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired      = errors.New("token expired")
	ErrInvalidToken      = errors.New("invalid token")
	ErrUnauthorized      = errors.New("unauthorized")

	// Validation errors
	ErrInvalidInput     = errors.New("invalid input")
	ErrMissingField     = errors.New("missing required field")
	ErrInvalidFormat    = errors.New("invalid format")

	// Resource errors
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrConflict         = errors.New("resource conflict")

	// Internal errors
	ErrInternal         = errors.New("internal server error")
	ErrDatabaseError    = errors.New("database error")
	ErrServiceError     = errors.New("service error")
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(statusCode int, message string, err error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

// MapErrorToHTTP maps common errors to HTTP status codes
func MapErrorToHTTP(err error) *HTTPError {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return NewHTTPError(http.StatusUnauthorized, "Invalid credentials", err)
	case errors.Is(err, ErrTokenExpired):
		return NewHTTPError(http.StatusUnauthorized, "Token expired", err)
	case errors.Is(err, ErrInvalidToken):
		return NewHTTPError(http.StatusUnauthorized, "Invalid token", err)
	case errors.Is(err, ErrUnauthorized):
		return NewHTTPError(http.StatusUnauthorized, "Unauthorized", err)
	case errors.Is(err, ErrInvalidInput):
		return NewHTTPError(http.StatusBadRequest, "Invalid input", err)
	case errors.Is(err, ErrMissingField):
		return NewHTTPError(http.StatusBadRequest, "Missing required field", err)
	case errors.Is(err, ErrInvalidFormat):
		return NewHTTPError(http.StatusBadRequest, "Invalid format", err)
	case errors.Is(err, ErrNotFound):
		return NewHTTPError(http.StatusNotFound, "Resource not found", err)
	case errors.Is(err, ErrAlreadyExists):
		return NewHTTPError(http.StatusConflict, "Resource already exists", err)
	case errors.Is(err, ErrConflict):
		return NewHTTPError(http.StatusConflict, "Resource conflict", err)
	case errors.Is(err, ErrDatabaseError):
		return NewHTTPError(http.StatusInternalServerError, "Database error", err)
	case errors.Is(err, ErrServiceError):
		return NewHTTPError(http.StatusInternalServerError, "Service error", err)
	case errors.Is(err, ErrInternal):
		return NewHTTPError(http.StatusInternalServerError, "Internal server error", err)
	default:
		return NewHTTPError(http.StatusInternalServerError, "Unknown error", err)
	}
}


