package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/saintparish4/harmonia/internal/domain"
)

// PricingRuleRepo implements domain.PricingRuleRepository
type PricingRuleRepo struct {
	db *sql.DB
}

// NewPricingRuleRepository creates a new pricing rule repository
func NewPricingRuleRepository(db *sql.DB) domain.PricingRuleRepository {
	return &PricingRuleRepo{db: db}
}

// Create creates a new pricing rule
func (r *PricingRuleRepo) Create(ctx context.Context, rule *domain.PricingRule) error {
	query := `
		INSERT INTO pricing_rules (
			id, user_id, name, description, strategy_type, config, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Generate ID if not provided
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	// Convert config to JSONB
	config := FromMap(rule.Config)

	_, err := r.db.ExecContext(
		ctx,
		query,
		rule.ID,
		rule.UserID,
		rule.Name,
		rule.Description,
		rule.StrategyType,
		config,
		rule.IsActive,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create pricing rule: %w", err)
	}

	return nil
}

// GetByID retrieves a pricing rule by ID
func (r *PricingRuleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PricingRule, error) {
	query := `
		SELECT id, user_id, name, description, strategy_type, config, 
		       is_active, created_at, updated_at, deleted_at
		FROM pricing_rules
		WHERE id = $1 AND deleted_at IS NULL
	`

	rule := &domain.PricingRule{}
	var config JSONB
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID,
		&rule.UserID,
		&rule.Name,
		&rule.Description,
		&rule.StrategyType,
		&config,
		&rule.IsActive,
		&rule.CreatedAt,
		&rule.UpdatedAt,
		&deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("pricing rule not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing rule: %w", err)
	}

	// Convert JSONB to map
	rule.Config = config.ToMap()

	if deletedAt.Valid {
		rule.DeletedAt = &deletedAt.Time
	}

	return rule, nil
}

// GetByUserID retrieves all active pricing rules for a user
func (r *PricingRuleRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.PricingRule, error) {
	query := `
		SELECT id, user_id, name, description, strategy_type, config, 
		       is_active, created_at, updated_at
		FROM pricing_rules
		WHERE user_id = $1 AND is_active = true AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pricing rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.PricingRule

	for rows.Next() {
		rule := &domain.PricingRule{}
		var config JSONB

		err := rows.Scan(
			&rule.ID,
			&rule.UserID,
			&rule.Name,
			&rule.Description,
			&rule.StrategyType,
			&config,
			&rule.IsActive,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pricing rule: %w", err)
		}

		rule.Config = config.ToMap()
		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pricing rules: %w", err)
	}

	return rules, nil
}

// Update updates an existing pricing rule
func (r *PricingRuleRepo) Update(ctx context.Context, rule *domain.PricingRule) error {
	query := `
		UPDATE pricing_rules
		SET name = $1, description = $2, strategy_type = $3, config = $4, 
		    is_active = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`

	rule.UpdatedAt = time.Now()
	config := FromMap(rule.Config)

	result, err := r.db.ExecContext(
		ctx,
		query,
		rule.Name,
		rule.Description,
		rule.StrategyType,
		config,
		rule.IsActive,
		rule.UpdatedAt,
		rule.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update pricing rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pricing rule not found: %s", rule.ID)
	}

	return nil
}

// Delete soft deletes a pricing rule
func (r *PricingRuleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE pricing_rules
		SET deleted_at = $1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete pricing rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pricing rule not found: %s", id)
	}

	return nil
}

// List retrieves pricing rules with optional filters
func (r *PricingRuleRepo) List(ctx context.Context, filter domain.PricingRuleFilter) ([]*domain.PricingRule, error) {
	query := `
		SELECT id, user_id, name, description, strategy_type, config, 
		       is_active, created_at, updated_at
		FROM pricing_rules
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	args := []interface{}{filter.UserID}
	argCount := 1

	// Add strategy type filter if provided
	if filter.StrategyType != "" {
		argCount++
		query += fmt.Sprintf(" AND strategy_type = $%d", argCount)
		args = append(args, filter.StrategyType)
	}

	// Add is_active filter if provided
	if filter.IsActive != nil {
		argCount++
		query += fmt.Sprintf(" AND is_active = $%d", argCount)
		args = append(args, *filter.IsActive)
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list pricing rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.PricingRule

	for rows.Next() {
		rule := &domain.PricingRule{}
		var config JSONB

		err := rows.Scan(
			&rule.ID,
			&rule.UserID,
			&rule.Name,
			&rule.Description,
			&rule.StrategyType,
			&config,
			&rule.IsActive,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pricing rule: %w", err)
		}

		rule.Config = config.ToMap()
		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pricing rules: %w", err)
	}

	return rules, nil
}
