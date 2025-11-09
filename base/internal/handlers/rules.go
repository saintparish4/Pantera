package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saintparish4/harmonia/internal/dto"
)

// PricingRule represents a pricing rule domain model
type PricingRule struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Name         string
	StrategyType string
	Config       map[string]interface{}
	IsActive     bool
}

// PricingRuleRepository defines operations for pricing rule management
type PricingRuleRepository interface {
	Create(ctx context.Context, rule *PricingRule) error
	GetByID(ctx context.Context, id uuid.UUID) (*PricingRule, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*PricingRule, error)
	Update(ctx context.Context, rule *PricingRule) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RulesHandler handles pricing rule endpoints
type RulesHandler struct {
	repo PricingRuleRepository
}

// NewRulesHandler creates a new rules handler
func NewRulesHandler(repo PricingRuleRepository) *RulesHandler {
	return &RulesHandler{repo: repo}
}

// Create handles POST /v1/pricing/rules
func (h *RulesHandler) Create(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	// Bind request
	var req dto.CreatePricingRuleRequest
	if !BindJSON(c, &req) {
		return
	}

	// Validate strategy type
	validStrategies := map[string]bool{
		"cost_plus":  true,
		"geographic": true,
		"time_based": true,
		"rule_based": true,
	}

	if !validStrategies[req.StrategyType] {
		BadRequest(c, "Invalid strategy type")
		return
	}

	ctx := c.Request.Context()

	// Create pricing rule
	rule := &PricingRule{
		ID:           uuid.New(),
		UserID:       userID,
		Name:         req.Name,
		StrategyType: req.StrategyType,
		Config:       req.Config,
		IsActive:     req.IsActive,
	}

	if err := h.repo.Create(ctx, rule); err != nil {
		HandleError(c, err)
		return
	}

	// Return response
	response := dto.PricingRuleResponse{
		ID:           rule.ID,
		UserID:       rule.UserID,
		Name:         rule.Name,
		StrategyType: rule.StrategyType,
		Config:       rule.Config,
		IsActive:     rule.IsActive,
	}

	Created(c, response)
}

// List handles GET /v1/pricing/rules
func (h *RulesHandler) List(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	ctx := c.Request.Context()

	// Get all rules for user
	rules, err := h.repo.GetByUserID(ctx, userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Convert to response DTOs
	response := make([]dto.PricingRuleResponse, len(rules))
	for i, rule := range rules {
		response[i] = dto.PricingRuleResponse{
			ID:           rule.ID,
			UserID:       rule.UserID,
			Name:         rule.Name,
			StrategyType: rule.StrategyType,
			Config:       rule.Config,
			IsActive:     rule.IsActive,
		}
	}

	Success(c, response)
}

// Get handles GET /v1/pricing/rules/:id
func (h *RulesHandler) Get(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	// Validate rule ID
	ruleID, err := ValidateUUID(c, "id")
	if err != nil {
		BadRequest(c, "Invalid rule ID")
		return
	}

	ctx := c.Request.Context()

	// Get rule
	rule, err := h.repo.GetByID(ctx, ruleID)
	if err != nil {
		NotFound(c, "Pricing rule not found")
		return
	}

	// Verify ownership
	if rule.UserID != userID {
		Forbidden(c, "Access denied")
		return
	}

	// Return response
	response := dto.PricingRuleResponse{
		ID:           rule.ID,
		UserID:       rule.UserID,
		Name:         rule.Name,
		StrategyType: rule.StrategyType,
		Config:       rule.Config,
		IsActive:     rule.IsActive,
	}

	Success(c, response)
}

// Update handles PUT /v1/pricing/rules/:id
func (h *RulesHandler) Update(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	// Validate rule ID
	ruleID, err := ValidateUUID(c, "id")
	if err != nil {
		BadRequest(c, "Invalid rule ID")
		return
	}

	// Bind request
	var req dto.UpdatePricingRuleRequest
	if !BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()

	// Get existing rule
	rule, err := h.repo.GetByID(ctx, ruleID)
	if err != nil {
		NotFound(c, "Pricing rule not found")
		return
	}

	// Verify ownership
	if rule.UserID != userID {
		Forbidden(c, "Access denied")
		return
	}

	// Update fields
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.StrategyType != nil {
		rule.StrategyType = *req.StrategyType
	}
	if req.Config != nil {
		rule.Config = req.Config
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}

	// Save updates
	if err := h.repo.Update(ctx, rule); err != nil {
		HandleError(c, err)
		return
	}

	// Return response
	response := dto.PricingRuleResponse{
		ID:           rule.ID,
		UserID:       rule.UserID,
		Name:         rule.Name,
		StrategyType: rule.StrategyType,
		Config:       rule.Config,
		IsActive:     rule.IsActive,
	}

	Success(c, response)
}

// Delete handles DELETE /v1/pricing/rules/:id
func (h *RulesHandler) Delete(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	// Validate rule ID
	ruleID, err := ValidateUUID(c, "id")
	if err != nil {
		BadRequest(c, "Invalid rule ID")
		return
	}

	ctx := c.Request.Context()

	// Get rule to verify ownership
	rule, err := h.repo.GetByID(ctx, ruleID)
	if err != nil {
		NotFound(c, "Pricing rule not found")
		return
	}

	// Verify ownership
	if rule.UserID != userID {
		Forbidden(c, "Access denied")
		return
	}

	// Delete rule
	if err := h.repo.Delete(ctx, ruleID); err != nil {
		HandleError(c, err)
		return
	}

	NoContent(c)
}
