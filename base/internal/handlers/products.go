package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saintparish4/harmonia/internal/dto"
)

// Product represents a product domain model
type Product struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	SKU         string
	Name        string
	Description string
	BaseCost    float64
	IsActive    bool
}

// ProductRepository defines operations for product management
type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	GetBySKU(ctx context.Context, userID uuid.UUID, sku string) (*Product, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Product, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProductsHandler handles product endpoints
type ProductsHandler struct {
	repo ProductRepository
}

// NewProductsHandler creates a new products handler
func NewProductsHandler(repo ProductRepository) *ProductsHandler {
	return &ProductsHandler{repo: repo}
}

// Create handles POST /v1/products
func (h *ProductsHandler) Create(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	// Bind request
	var req dto.CreateProductRequest
	if !BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()

	// Check if SKU already exists for this user
	existing, _ := h.repo.GetBySKU(ctx, userID, req.SKU)
	if existing != nil {
		Conflict(c, "Product with this SKU already exists")
		return
	}

	// Create product
	product := &Product{
		ID:          uuid.New(),
		UserID:      userID,
		SKU:         req.SKU,
		Name:        req.Name,
		Description: req.Description,
		BaseCost:    req.BaseCost,
		IsActive:    true,
	}

	if err := h.repo.Create(ctx, product); err != nil {
		HandleError(c, err)
		return
	}

	// Return response
	response := dto.ProductResponse{
		ID:          product.ID,
		UserID:      product.UserID,
		SKU:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		BaseCost:    product.BaseCost,
		IsActive:    product.IsActive,
	}

	Created(c, response)
}

// List handles GET /v1/products
func (h *ProductsHandler) List(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	ctx := c.Request.Context()

	// Get all products for user
	products, err := h.repo.GetByUserID(ctx, userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Convert to response DTOs
	response := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		response[i] = dto.ProductResponse{
			ID:          product.ID,
			UserID:      product.UserID,
			SKU:         product.SKU,
			Name:        product.Name,
			Description: product.Description,
			BaseCost:    product.BaseCost,
			IsActive:    product.IsActive,
		}
	}

	Success(c, response)
}

// Get handles GET /v1/products/:id
func (h *ProductsHandler) Get(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	// Validate product ID
	productID, err := ValidateUUID(c, "id")
	if err != nil {
		BadRequest(c, "Invalid product ID")
		return
	}

	ctx := c.Request.Context()

	// Get product
	product, err := h.repo.GetByID(ctx, productID)
	if err != nil {
		NotFound(c, "Product not found")
		return
	}

	// Verify ownership
	if product.UserID != userID {
		Forbidden(c, "Access denied")
		return
	}

	// Return response
	response := dto.ProductResponse{
		ID:          product.ID,
		UserID:      product.UserID,
		SKU:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		BaseCost:    product.BaseCost,
		IsActive:    product.IsActive,
	}

	Success(c, response)
}

// Update handles PUT /v1/products/:id
func (h *ProductsHandler) Update(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	// Validate product ID
	productID, err := ValidateUUID(c, "id")
	if err != nil {
		BadRequest(c, "Invalid product ID")
		return
	}

	// Bind request
	var req dto.UpdateProductRequest
	if !BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()

	// Get existing product
	product, err := h.repo.GetByID(ctx, productID)
	if err != nil {
		NotFound(c, "Product not found")
		return
	}

	// Verify ownership
	if product.UserID != userID {
		Forbidden(c, "Access denied")
		return
	}

	// Update fields
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.BaseCost != nil {
		product.BaseCost = *req.BaseCost
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	// Save updates
	if err := h.repo.Update(ctx, product); err != nil {
		HandleError(c, err)
		return
	}

	// Return response
	response := dto.ProductResponse{
		ID:          product.ID,
		UserID:      product.UserID,
		SKU:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		BaseCost:    product.BaseCost,
		IsActive:    product.IsActive,
	}

	Success(c, response)
}

// Delete handles DELETE /v1/products/:id
func (h *ProductsHandler) Delete(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	// Validate product ID
	productID, err := ValidateUUID(c, "id")
	if err != nil {
		BadRequest(c, "Invalid product ID")
		return
	}

	ctx := c.Request.Context()

	// Get product to verify ownership
	product, err := h.repo.GetByID(ctx, productID)
	if err != nil {
		NotFound(c, "Product not found")
		return
	}

	// Verify ownership
	if product.UserID != userID {
		Forbidden(c, "Access denied")
		return
	}

	// Delete product
	if err := h.repo.Delete(ctx, productID); err != nil {
		HandleError(c, err)
		return
	}

	NoContent(c)
}
