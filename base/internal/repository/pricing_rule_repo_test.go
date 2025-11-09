package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/saintparish4/harmonia/internal/domain"
)

// DB is a test database connection for integration tests.
// It should be initialized in test setup when running integration tests.
var DB *sql.DB

// TODO: Implement test database setup and teardown:
//   - Add TestMain function to initialize database connection for integration tests
//   - Set up test database schema/migrations
//   - Implement cleanup between tests (transaction rollback or table truncation)
//   - Add build tag for integration tests (e.g., //go:build integration)
//   - Consider using testcontainers or a dedicated test database instance

func TestPricingRuleRepository_Create(t *testing.T) {
	// Note: These tests require a test database connection
	// In a real scenario, you'd set up a test database or use a mock
	// For now, these are integration test templates

	t.Skip("Requires database connection - run with integration tag")

	ctx := context.Background()
	repo := NewPricingRuleRepository(DB)

	rule := &domain.PricingRule{
		UserID:       uuid.New(),
		Name:         "Test Rule",
		Description:  "Test pricing rule",
		StrategyType: domain.StrategyTypeCostPlus,
		Config: map[string]interface{}{
			"markup_type":  "percentage",
			"markup_value": 25.0,
		},
		IsActive: true,
	}

	err := repo.Create(ctx, rule)
	if err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	// Verify ID was generated
	if rule.ID == uuid.Nil {
		t.Error("expected ID to be generated")
	}

	// Verify timestamps were set
	if rule.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if rule.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestPricingRuleRepository_GetByID(t *testing.T) {
	t.Skip("Requires database connection - run with integration tag")

	ctx := context.Background()
	repo := NewPricingRuleRepository(DB)

	// Create a rule first
	userID := uuid.New()
	rule := &domain.PricingRule{
		UserID:       userID,
		Name:         "Test Rule",
		Description:  "Test pricing rule",
		StrategyType: domain.StrategyTypeCostPlus,
		Config: map[string]interface{}{
			"markup_type":  "percentage",
			"markup_value": 25.0,
		},
		IsActive: true,
	}

	err := repo.Create(ctx, rule)
	if err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	// Retrieve the rule
	retrieved, err := repo.GetByID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("failed to get rule: %v", err)
	}

	// Verify fields
	if retrieved.ID != rule.ID {
		t.Errorf("expected ID %s, got %s", rule.ID, retrieved.ID)
	}
	if retrieved.Name != rule.Name {
		t.Errorf("expected Name %s, got %s", rule.Name, retrieved.Name)
	}
	if retrieved.StrategyType != rule.StrategyType {
		t.Errorf("expected StrategyType %s, got %s", rule.StrategyType, retrieved.StrategyType)
	}

	// Verify config was preserved
	markupType, ok := retrieved.Config["markup_type"].(string)
	if !ok || markupType != "percentage" {
		t.Error("config not properly preserved")
	}
}

func TestPricingRuleRepository_GetByUserID(t *testing.T) {
	t.Skip("Requires database connection - run with integration tag")

	ctx := context.Background()
	repo := NewPricingRuleRepository(DB)

	userID := uuid.New()

	// Create multiple rules
	rules := []*domain.PricingRule{
		{
			UserID:       userID,
			Name:         "Rule 1",
			StrategyType: domain.StrategyTypeCostPlus,
			Config:       map[string]interface{}{"markup_value": 10.0},
			IsActive:     true,
		},
		{
			UserID:       userID,
			Name:         "Rule 2",
			StrategyType: domain.StrategyTypeGeographic,
			Config:       map[string]interface{}{"regional_multipliers": map[string]interface{}{"US": 1.0}},
			IsActive:     true,
		},
		{
			UserID:       userID,
			Name:         "Rule 3 (Inactive)",
			StrategyType: domain.StrategyTypeCostPlus,
			Config:       map[string]interface{}{"markup_value": 15.0},
			IsActive:     false,
		},
	}

	for _, rule := range rules {
		if err := repo.Create(ctx, rule); err != nil {
			t.Fatalf("failed to create rule: %v", err)
		}
	}

	// Retrieve active rules
	retrieved, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("failed to get rules: %v", err)
	}

	// Should only return active rules (2)
	if len(retrieved) != 2 {
		t.Errorf("expected 2 active rules, got %d", len(retrieved))
	}
}

func TestPricingRuleRepository_Update(t *testing.T) {
	t.Skip("Requires database connection - run with integration tag")

	ctx := context.Background()
	repo := NewPricingRuleRepository(DB)

	// Create a rule
	rule := &domain.PricingRule{
		UserID:       uuid.New(),
		Name:         "Original Name",
		Description:  "Original Description",
		StrategyType: domain.StrategyTypeCostPlus,
		Config:       map[string]interface{}{"markup_value": 10.0},
		IsActive:     true,
	}

	if err := repo.Create(ctx, rule); err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	// Update the rule
	rule.Name = "Updated Name"
	rule.Description = "Updated Description"
	rule.Config["markup_value"] = 25.0

	if err := repo.Update(ctx, rule); err != nil {
		t.Fatalf("failed to update rule: %v", err)
	}

	// Retrieve and verify
	retrieved, err := repo.GetByID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("failed to get rule: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("expected Name 'Updated Name', got %s", retrieved.Name)
	}
	if retrieved.Description != "Updated Description" {
		t.Errorf("expected Description 'Updated Description', got %s", retrieved.Description)
	}

	markupValue, ok := retrieved.Config["markup_value"].(float64)
	if !ok || markupValue != 25.0 {
		t.Error("config not properly updated")
	}
}

func TestPricingRuleRepository_Delete(t *testing.T) {
	t.Skip("Requires database connection - run with integration tag")

	ctx := context.Background()
	repo := NewPricingRuleRepository(DB)

	// Create a rule
	rule := &domain.PricingRule{
		UserID:       uuid.New(),
		Name:         "Test Rule",
		StrategyType: domain.StrategyTypeCostPlus,
		Config:       map[string]interface{}{"markup_value": 10.0},
		IsActive:     true,
	}

	if err := repo.Create(ctx, rule); err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	// Delete the rule
	if err := repo.Delete(ctx, rule.ID); err != nil {
		t.Fatalf("failed to delete rule: %v", err)
	}

	// Verify it's soft deleted (should not be retrievable)
	_, err := repo.GetByID(ctx, rule.ID)
	if err == nil {
		t.Error("expected error when retrieving deleted rule")
	}
}

func TestPricingRuleRepository_List(t *testing.T) {
	t.Skip("Requires database connection - run with integration tag")

	ctx := context.Background()
	repo := NewPricingRuleRepository(DB)

	userID := uuid.New()

	// Create rules with different strategies
	rules := []*domain.PricingRule{
		{
			UserID:       userID,
			Name:         "Cost Plus Rule",
			StrategyType: domain.StrategyTypeCostPlus,
			Config:       map[string]interface{}{"markup_value": 10.0},
			IsActive:     true,
		},
		{
			UserID:       userID,
			Name:         "Geographic Rule",
			StrategyType: domain.StrategyTypeGeographic,
			Config:       map[string]interface{}{"regional_multipliers": map[string]interface{}{"US": 1.0}},
			IsActive:     true,
		},
	}

	for _, rule := range rules {
		if err := repo.Create(ctx, rule); err != nil {
			t.Fatalf("failed to create rule: %v", err)
		}
	}

	// List with strategy filter
	filter := domain.PricingRuleFilter{
		UserID:       userID,
		StrategyType: domain.StrategyTypeCostPlus,
		Limit:        10,
	}

	retrieved, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("failed to list rules: %v", err)
	}

	// Should only return cost_plus rules
	if len(retrieved) != 1 {
		t.Errorf("expected 1 cost_plus rule, got %d", len(retrieved))
	}

	if retrieved[0].StrategyType != domain.StrategyTypeCostPlus {
		t.Errorf("expected cost_plus strategy, got %s", retrieved[0].StrategyType)
	}
}

// Mock test for JSONB conversion
func TestJSONB_Conversion(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
	}{
		{
			name: "simple map",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"key1": "value1",
				"nested": map[string]interface{}{
					"inner": "value",
				},
			},
		},
		{
			name: "with array",
			input: map[string]interface{}{
				"key1":  "value1",
				"array": []interface{}{1, 2, 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to JSONB
			jsonb := FromMap(tt.input)

			// Convert back to map
			result := jsonb.ToMap()

			// Verify keys exist
			for key := range tt.input {
				if _, ok := result[key]; !ok {
					t.Errorf("expected key %s in result", key)
				}
			}
		})
	}
}
