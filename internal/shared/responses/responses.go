package responses

import (
	"net/http"
	"police-trafic-api-frontend-aligned/internal/shared/errors"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// SuccessResponse represents a successful response
type SuccessResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Success sends a successful JSON response
func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, SuccessResponse{Data: data})
}

// SuccessWithMessage sends a successful JSON response with a message
func SuccessWithMessage(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusOK, SuccessResponse{
		Message: message,
		Data:    data,
	})
}

// Created sends a created JSON response
func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, SuccessResponse{Data: data})
}

// Paginated sends a paginated JSON response
func Paginated(c echo.Context, data interface{}, total, page, pageSize int) error {
	totalPages := (total + pageSize - 1) / pageSize
	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// Error sends an error JSON response
func Error(c echo.Context, err error) error {
	logger := c.Get("logger")
	if logger != nil {
		if zapLogger, ok := logger.(*zap.Logger); ok {
			zapLogger.Error("API Error", zap.Error(err))
		}
	}

	httpErr := errors.MapErrorToHTTP(err)
	
	return c.JSON(httpErr.StatusCode, errors.ErrorResponse{
		Error:   http.StatusText(httpErr.StatusCode),
		Message: httpErr.Message,
		Code:    httpErr.StatusCode,
	})
}

// BadRequest sends a bad request error response
func BadRequest(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, errors.ErrorResponse{
		Error:   "bad_request",
		Message: message,
		Code:    http.StatusBadRequest,
	})
}

// Unauthorized sends an unauthorized error response
func Unauthorized(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, errors.ErrorResponse{
		Error:   "unauthorized",
		Message: message,
		Code:    http.StatusUnauthorized,
	})
}

// Forbidden sends a forbidden error response
func Forbidden(c echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, errors.ErrorResponse{
		Error:   "forbidden",
		Message: message,
		Code:    http.StatusForbidden,
	})
}

// NotFound sends a not found error response
func NotFound(c echo.Context, message string) error {
	return c.JSON(http.StatusNotFound, errors.ErrorResponse{
		Error:   "not_found",
		Message: message,
		Code:    http.StatusNotFound,
	})
}

// Conflict sends a conflict error response
func Conflict(c echo.Context, message string) error {
	return c.JSON(http.StatusConflict, errors.ErrorResponse{
		Error:   "conflict",
		Message: message,
		Code:    http.StatusConflict,
	})
}

// InternalServerError sends an internal server error response
func InternalServerError(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
		Error:   "internal_server_error",
		Message: message,
		Code:    http.StatusInternalServerError,
	})
}


