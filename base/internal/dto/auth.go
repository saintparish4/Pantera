package dto

import (
	"time"

	"github.com/google/uuid"
)

// --- User DTOs ---

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// UserResponse represents a user
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// --- API Key DTOs ---

// CreateKeyRequest represents a request to create an API key
type CreateKeyRequest struct {
	Name   string `json:"name" binding:"required"`
	IsLive bool   `json:"is_live"`
}

// CreateKeyResponse represents the response when creating an API key
type CreateKeyResponse struct {
	ID        uuid.UUID `json:"id"`
	Key       string    `json:"key"` // Raw key - only shown once
	KeyPrefix string    `json:"key_prefix"`
	Name      string    `json:"name"`
	IsLive    bool      `json:"is_live"`
	CreatedAt string    `json:"created_at"`
	Warning   string    `json:"warning"`
}

// APIKeyResponse represents an API key (without the raw key)
type APIKeyResponse struct {
	ID        uuid.UUID  `json:"id"`
	KeyPrefix string     `json:"key_prefix"`
	Name      string     `json:"name"`
	IsLive    bool       `json:"is_live"`
	IsRevoked bool       `json:"is_revoked"`
	LastUsed  *time.Time `json:"last_used,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
