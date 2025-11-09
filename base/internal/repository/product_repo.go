package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/saintparish4/harmonia/internal/domain"
)

// ProductRepo implements domain.ProductRepository
type ProductRepo struct {
	db *sql.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *sql.DB) domain.ProductRepository {
	return &ProductRepo{db: db}
}

// Create creates a new product
func (r *ProductRepo) Create(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (
			id, user_id, sku, name, description, base_cost, 
			default_rule_id, metadata, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// Generate ID if not provided
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	// Convert metadata to JSONB
	metadata := FromMap(product.Metadata)

	_, err := r.db.ExecContext(
		ctx,
		query,
		product.ID,
		product.UserID,
		product.SKU,
		product.Name,
		product.Description,
		product.BaseCost,
		product.DefaultRuleID,
		metadata,
		product.IsActive,
		product.CreatedAt,
		product.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID retrieves a product by ID
func (r *ProductRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	query := `
		SELECT id, user_id, sku, name, description, base_cost, 
		       default_rule_id, metadata, is_active, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	product := &domain.Product{}
	var metadata JSONB
	var defaultRuleID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.UserID,
		&product.SKU,
		&product.Name,
		&product.Description,
		&product.BaseCost,
		&defaultRuleID,
		&metadata,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Convert metadata
	product.Metadata = metadata.ToMap()

	// Handle nullable default_rule_id
	if defaultRuleID.Valid {
		ruleID, err := uuid.Parse(defaultRuleID.String)
		if err == nil {
			product.DefaultRuleID = &ruleID
		}
	}

	return product, nil
}

// GetBySKU retrieves a product by SKU for a specific user
func (r *ProductRepo) GetBySKU(ctx context.Context, userID uuid.UUID, sku string) (*domain.Product, error) {
	query := `
		SELECT id, user_id, sku, name, description, base_cost, 
		       default_rule_id, metadata, is_active, created_at, updated_at
		FROM products
		WHERE user_id = $1 AND sku = $2
	`

	product := &domain.Product{}
	var metadata JSONB
	var defaultRuleID sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID, sku).Scan(
		&product.ID,
		&product.UserID,
		&product.SKU,
		&product.Name,
		&product.Description,
		&product.BaseCost,
		&defaultRuleID,
		&metadata,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product not found: %s", sku)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	product.Metadata = metadata.ToMap()

	if defaultRuleID.Valid {
		ruleID, err := uuid.Parse(defaultRuleID.String)
		if err == nil {
			product.DefaultRuleID = &ruleID
		}
	}

	return product, nil
}

// GetByUserID retrieves all active products for a user
func (r *ProductRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Product, error) {
	query := `
		SELECT id, user_id, sku, name, description, base_cost, 
		       default_rule_id, metadata, is_active, created_at, updated_at
		FROM products
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product

	for rows.Next() {
		product := &domain.Product{}
		var metadata JSONB
		var defaultRuleID sql.NullString

		err := rows.Scan(
			&product.ID,
			&product.UserID,
			&product.SKU,
			&product.Name,
			&product.Description,
			&product.BaseCost,
			&defaultRuleID,
			&metadata,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		product.Metadata = metadata.ToMap()

		if defaultRuleID.Valid {
			ruleID, err := uuid.Parse(defaultRuleID.String)
			if err == nil {
				product.DefaultRuleID = &ruleID
			}
		}

		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

// Update updates an existing product
func (r *ProductRepo) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products
		SET sku = $1, name = $2, description = $3, base_cost = $4, 
		    default_rule_id = $5, metadata = $6, is_active = $7, updated_at = $8
		WHERE id = $9
	`

	product.UpdatedAt = time.Now()
	metadata := FromMap(product.Metadata)

	result, err := r.db.ExecContext(
		ctx,
		query,
		product.SKU,
		product.Name,
		product.Description,
		product.BaseCost,
		product.DefaultRuleID,
		metadata,
		product.IsActive,
		product.UpdatedAt,
		product.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found: %s", product.ID)
	}

	return nil
}

// Delete soft deletes a product
func (r *ProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE products
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found: %s", id)
	}

	return nil
}

// List retrieves products with optional filters
func (r *ProductRepo) List(ctx context.Context, filter domain.ProductFilter) ([]*domain.Product, error) {
	query := `
		SELECT id, user_id, sku, name, description, base_cost, 
		       default_rule_id, metadata, is_active, created_at, updated_at
		FROM products
		WHERE user_id = $1
	`

	args := []interface{}{filter.UserID}
	argCount := 1

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
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product

	for rows.Next() {
		product := &domain.Product{}
		var metadata JSONB
		var defaultRuleID sql.NullString

		err := rows.Scan(
			&product.ID,
			&product.UserID,
			&product.SKU,
			&product.Name,
			&product.Description,
			&product.BaseCost,
			&defaultRuleID,
			&metadata,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		product.Metadata = metadata.ToMap()

		if defaultRuleID.Valid {
			ruleID, err := uuid.Parse(defaultRuleID.String)
			if err == nil {
				product.DefaultRuleID = &ruleID
			}
		}

		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}
