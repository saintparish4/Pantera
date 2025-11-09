package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saintparish4/harmonia/internal/dto"
)

// APIKey represents an API key domain model
type APIKey struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	KeyHash   string
	KeyPrefix string
	Name      string
	IsLive    bool
	IsRevoked bool
	CreatedAt string
}

// User represents a user domain model
type User struct {
	ID    uuid.UUID
	Email string
}

// APIKeyRepository defines operations for API key management
type APIKeyRepository interface {
	Create(ctx context.Context, key *APIKey) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*APIKey, error)
	Revoke(ctx context.Context, keyID uuid.UUID) error
}

// UserRepository defines operations for user management
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
}

// KeyGenerator defines interface for generating API keys
type KeyGenerator interface {
	GenerateAPIKey(isLive bool) (rawKey string, keyHash string, keyPrefix string, err error)
}

// KeysHandler handles API key management endpoints
type KeysHandler struct {
	repo      APIKeyRepository
	userRepo  UserRepository
	generator KeyGenerator
}

// NewKeysHandler creates a new keys handler
func NewKeysHandler(repo APIKeyRepository, userRepo UserRepository, generator KeyGenerator) *KeysHandler {
	return &KeysHandler{
		repo:      repo,
		userRepo:  userRepo,
		generator: generator,
	}
}

// Create handles POST /v1/auth/keys
func (h *KeysHandler) Create(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	// Bind request
	var req dto.CreateKeyRequest
	if !BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()

	// Generate API key
	rawKey, keyHash, keyPrefix, err := h.generator.GenerateAPIKey(req.IsLive)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Create API key record
	apiKey := &APIKey{
		ID:        uuid.New(),
		UserID:    userID,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		Name:      req.Name,
		IsLive:    req.IsLive,
		IsRevoked: false,
	}

	if err := h.repo.Create(ctx, apiKey); err != nil {
		HandleError(c, err)
		return
	}

	// Return response with raw key (only shown once)
	response := dto.CreateKeyResponse{
		ID:        apiKey.ID,
		Key:       rawKey,
		KeyPrefix: keyPrefix,
		Name:      req.Name,
		IsLive:    req.IsLive,
		CreatedAt: apiKey.CreatedAt,
		Warning:   "Save this key securely. You won't be able to see it again!",
	}

	Created(c, response)
}

// List handles GET /v1/auth/keys
func (h *KeysHandler) List(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	ctx := c.Request.Context()

	// Get all keys for user
	keys, err := h.repo.GetByUserID(ctx, userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Convert to response DTOs
	response := make([]dto.APIKeyResponse, len(keys))
	for i, key := range keys {
		response[i] = dto.APIKeyResponse{
			ID:        key.ID,
			KeyPrefix: key.KeyPrefix,
			Name:      key.Name,
			IsLive:    key.IsLive,
			IsRevoked: key.IsRevoked,
			// Note: Not including LastUsed for now - can be added later
		}
	}

	Success(c, response)
}

// Revoke handles DELETE /v1/auth/keys/:id
func (h *KeysHandler) Revoke(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	// Validate key ID
	keyID, err := ValidateUUID(c, "id")
	if err != nil {
		BadRequest(c, "Invalid key ID")
		return
	}

	ctx := c.Request.Context()

	// Get all keys to verify ownership
	keys, err := h.repo.GetByUserID(ctx, userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Verify ownership
	found := false
	for _, key := range keys {
		if key.ID == keyID {
			found = true
			break
		}
	}

	if !found {
		NotFound(c, "API key not found")
		return
	}

	// Revoke key
	if err := h.repo.Revoke(ctx, keyID); err != nil {
		HandleError(c, err)
		return
	}

	SuccessWithMessage(c, "API key revoked successfully", nil)
}

// CreateUser handles POST /v1/users (for demo/testing purposes)
func (h *KeysHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if !BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()

	// Check if user already exists
	existingUser, _ := h.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		BadRequest(c, "User with this email already exists")
		return
	}

	// Create user
	user := &User{
		ID:    uuid.New(),
		Email: req.Email,
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		HandleError(c, err)
		return
	}

	response := dto.UserResponse{
		ID:    user.ID,
		Email: user.Email,
	}

	Created(c, response)
}
