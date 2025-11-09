package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Code    string            `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Response utilities

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   "Bad Request",
		Message: message,
		Code:    "BAD_REQUEST",
	})
}

// BadRequestWithDetails sends a 400 with validation details
func BadRequestWithDetails(c *gin.Context, message string, details map[string]string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   "Bad Request",
		Message: message,
		Code:    "VALIDATION_ERROR",
		Details: details,
	})
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, ErrorResponse{
		Error:   "Unauthorized",
		Message: message,
		Code:    "UNAUTHORIZED",
	})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, ErrorResponse{
		Error:   "Forbidden",
		Message: message,
		Code:    "FORBIDDEN",
	})
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Error:   "Not Found",
		Message: message,
		Code:    "NOT_FOUND",
	})
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, ErrorResponse{
		Error:   "Conflict",
		Message: message,
		Code:    "CONFLICT",
	})
}

// TooManyRequests sends a 429 Rate Limit response
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, ErrorResponse{
		Error:   "Too Many Requests",
		Message: message,
		Code:    "RATE_LIMIT_EXCEEDED",
	})
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   "Internal Server Error",
		Message: "An unexpected error occurred. Please try again later.",
		Code:    "INTERNAL_ERROR",
	})
}

// HandleError intelligently routes errors to appropriate responses
func HandleError(c *gin.Context, err error) {
	// You can add custom error type handling here
	// For now, default to internal error
	InternalError(c)
}

// Success sends a 200 OK response with data
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessage sends a 200 OK response with message and optional data
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Helper functions for common operations

// BindJSON attempts to bind JSON and returns error response if it fails
func BindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		BadRequest(c, err.Error())
		return false
	}
	return true
}

// ValidateUUID validates a UUID parameter from the URL
func ValidateUUID(c *gin.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// MustGetUserID gets the user ID from context or aborts
func MustGetUserID(c *gin.Context) uuid.UUID {
	value, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, "User ID not found in context")
		return uuid.Nil
	}

	userID, ok := value.(uuid.UUID)
	if !ok {
		InternalError(c)
		return uuid.Nil
	}

	return userID
}
