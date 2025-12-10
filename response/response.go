// Package response provides standard API response formats for all services.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"requestId"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
	RequestID  string         `json:"requestId"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"perPage"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// getRequestID gets or generates a request ID from context
func getRequestID(c *gin.Context) string {
	// Try to get request ID from context (if set by middleware)
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	// Generate new request ID if not exists
	return uuid.New().String()
}

// Success sends a successful response with the given data
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// OK sends a 200 OK response
func OK(c *gin.Context, data interface{}) {
	Success(c, http.StatusOK, data)
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	Success(c, http.StatusCreated, data)
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
		RequestID: getRequestID(c),
	})
}

// ErrorWithDetails sends an error response with additional details
func ErrorWithDetails(c *gin.Context, statusCode int, code string, message string, details interface{}) {
	c.JSON(statusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: getRequestID(c),
	})
}

// BadRequest sends a 400 Bad Request error
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized sends a 401 Unauthorized error
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 Forbidden error
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 Not Found error
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict sends a 409 Conflict error
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, "CONFLICT", message)
}

// InternalError sends a 500 Internal Server Error
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, data interface{}, page, perPage int, total int64) {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Data: data,
		Pagination: PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
		RequestID: getRequestID(c),
	})
}

// ValidationError sends a validation error with field details
func ValidationError(c *gin.Context, errors map[string]string) {
	ErrorWithDetails(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", errors)
}
