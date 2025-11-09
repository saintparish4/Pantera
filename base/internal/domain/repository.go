package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PricingRuleRepository defines operations for pricing rules
type PricingRuleRepository interface {
	// Create creates a new pricing rule
	Create(ctx context.Context, rule *PricingRule) error

	// GetByID retrieves a pricing rule by ID
	GetByID(ctx context.Context, id uuid.UUID) (*PricingRule, error)

	// GetByUserID retrieves all active pricing rules for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*PricingRule, error)

	// Update updates an existing pricing rule
	Update(ctx context.Context, rule *PricingRule) error

	// Delete soft deletes a pricing rule
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves pricing rules with optional filters
	List(ctx context.Context, filter PricingRuleFilter) ([]*PricingRule, error)
}

// ProductRepository defines operations for products
type ProductRepository interface {
	// Create creates a new product
	Create(ctx context.Context, product *Product) error

	// GetByID retrieves a product by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)

	// GetBySKU retrieves a product by SKU for a specific user
	GetBySKU(ctx context.Context, userID uuid.UUID, sku string) (*Product, error)

	// GetByUserID retrieves all active products for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Product, error)

	// Update updates an existing product
	Update(ctx context.Context, product *Product) error

	// Delete soft deletes a product
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves products with optional filters
	List(ctx context.Context, filter ProductFilter) ([]*Product, error)
}

// CalculationLogRepository defines operations for calculation logs
type CalculationLogRepository interface {
	// Create creates a new calculation log entry
	Create(ctx context.Context, log *CalculationLog) error

	// GetByID retrieves a calculation log by ID
	GetByID(ctx context.Context, id uuid.UUID) (*CalculationLog, error)

	// GetByUserID retrieves calculation logs for a user
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*CalculationLog, error)

	// List retrieves calculation logs with filters
	List(ctx context.Context, filter CalculationLogFilter) ([]*CalculationLog, error)

	// GetStats retrieves calculation statistics
	GetStats(ctx context.Context, userID uuid.UUID, from, to time.Time) (*CalculationStats, error)
}

// APIKeyRepository defines operations for API keys
type APIKeyRepository interface {
	// Create creates a new API key
	Create(ctx context.Context, key *APIKey) error

	// GetByHash retrieves an API key by its hash
	GetByHash(ctx context.Context, hash string) (*APIKey, error)

	// GetByUserID retrieves all API keys for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*APIKey, error)

	// UpdateLastUsed updates the last_used_at timestamp
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error

	// Revoke revokes an API key (soft delete)
	Revoke(ctx context.Context, id uuid.UUID) error

	// Delete permanently deletes an API key
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserRepository defines operations for users
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *User) error

	// Delete deletes a user
	Delete(ctx context.Context, id uuid.UUID) error
}

// Filter types for list operations

// PricingRuleFilter defines filters for pricing rules
type PricingRuleFilter struct {
	UserID       uuid.UUID
	StrategyType string
	IsActive     *bool
	Limit        int
	Offset       int
}

// ProductFilter defines filters for products
type ProductFilter struct {
	UserID   uuid.UUID
	IsActive *bool
	Limit    int
	Offset   int
}

// CalculationLogFilter defines filters for calculation logs
type CalculationLogFilter struct {
	UserID       uuid.UUID
	APIKeyID     *uuid.UUID
	RuleID       *uuid.UUID
	StrategyType string
	From         time.Time
	To           time.Time
	Limit        int
	Offset       int
}

// Additional domain models

// User represents an application user
type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIKey represents an API authentication key
type APIKey struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	KeyHash    string     `json:"-"` // Never expose hash
	KeyPrefix  string     `json:"key_prefix"`
	Name       string     `json:"name"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

// CalculationStats represents aggregated calculation statistics
type CalculationStats struct {
	TotalCalculations int            `json:"total_calculations"`
	ByStrategy        map[string]int `json:"by_strategy"`
	AvgExecutionTime  float64        `json:"avg_execution_time_ms"`
	MinPrice          float64        `json:"min_price"`
	MaxPrice          float64        `json:"max_price"`
	AvgPrice          float64        `json:"avg_price"`
	Period            string         `json:"period"`
}
