package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/saintparish4/harmonia/internal/domain"
)

// APIKeyRepo implements domain.APIKeyRepository
type APIKeyRepo struct {
	db *sql.DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *sql.DB) domain.APIKeyRepository {
	return &APIKeyRepo{db: db}
}

// Create creates a new API key
func (r *APIKeyRepo) Create(ctx context.Context, key *domain.APIKey) error {
	query := `
		INSERT INTO api_keys (
			id, user_id, key_hash, key_prefix, name, is_active, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	// Generate ID if not provided
	if key.ID == uuid.Nil {
		key.ID = uuid.New()
	}

	// Set timestamp
	if key.CreatedAt.IsZero() {
		key.CreatedAt = time.Now()
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		key.ID,
		key.UserID,
		key.KeyHash,
		key.KeyPrefix,
		key.Name,
		key.IsActive,
		key.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	return nil
}

// GetByHash retrieves an API key by its hash
func (r *APIKeyRepo) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, key_prefix, name, last_used_at, 
		       is_active, created_at, revoked_at
		FROM api_keys
		WHERE key_hash = $1 AND is_active = true
	`

	key := &domain.APIKey{}
	var lastUsedAt, revokedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&key.ID,
		&key.UserID,
		&key.KeyHash,
		&key.KeyPrefix,
		&key.Name,
		&lastUsedAt,
		&key.IsActive,
		&key.CreatedAt,
		&revokedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("API key not found or inactive")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	// Handle nullable timestamps
	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}
	if revokedAt.Valid {
		key.RevokedAt = &revokedAt.Time
	}

	return key, nil
}

// GetByUserID retrieves all API keys for a user
func (r *APIKeyRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, key_prefix, name, last_used_at, 
		       is_active, created_at, revoked_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var keys []*domain.APIKey

	for rows.Next() {
		key := &domain.APIKey{}
		var lastUsedAt, revokedAt sql.NullTime

		err := rows.Scan(
			&key.ID,
			&key.UserID,
			&key.KeyHash,
			&key.KeyPrefix,
			&key.Name,
			&lastUsedAt,
			&key.IsActive,
			&key.CreatedAt,
			&revokedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}

		if lastUsedAt.Valid {
			key.LastUsedAt = &lastUsedAt.Time
		}
		if revokedAt.Valid {
			key.RevokedAt = &revokedAt.Time
		}

		keys = append(keys, key)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating API keys: %w", err)
	}

	return keys, nil
}

// UpdateLastUsed updates the last_used_at timestamp for an API key
func (r *APIKeyRepo) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE api_keys
		SET last_used_at = $1
		WHERE id = $2 AND is_active = true
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to update API key last_used_at: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("API key not found or inactive: %s", id)
	}

	return nil
}

// Revoke revokes an API key (soft delete)
func (r *APIKeyRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE api_keys
		SET is_active = false, revoked_at = $1
		WHERE id = $2 AND is_active = true
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("API key not found or already revoked: %s", id)
	}

	return nil
}

// Delete permanently deletes an API key
func (r *APIKeyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM api_keys WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("API key not found: %s", id)
	}

	return nil
}
