package routes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/saintparish4/pantera/base/database"
	"github.com/saintparish4/pantera/base/models"
	"github.com/saintparish4/pantera/base/services"
	)

func SetupPricingRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		// Pricing calculation endpoint (main feature!)
		api.POST("/calculate", calculatePrice)

		// Pricing rules management
		api.GET("/rules", getRules)
		api.GET("/rules/:id", getRule)
		api.POST("/rules", createRule)
		api.PUT("/rules/:id", updateRule)
		api.DELETE("/rules/:id", deleteRule)

		// Analytics/audit
		api.GET("/calculations", getCalculations)
	}
}

// POST /api/v1/calculate - Calculate price based on rule
func calculatePrice(c *gin.Context) {
	var req models.PriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	engine := services.NewPricingEngine()
	response, err := engine.CalculatePrice(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /api/v1/rules - Get all pricing rules
func getRules(c *gin.Context) {
	query := `
		SELECT id, name, strategy, base_price, markup_percentage,
		       min_price, max_price, demand_multiplier, active, created_at, updated_at
		FROM pricing_rules
		ORDER BY created_at DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var rules []models.PricingRule
	for rows.Next() {
		var rule models.PricingRule
		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Strategy, &rule.BasePrice,
			&rule.MarkupPercentage, &rule.MinPrice, &rule.MaxPrice,
			&rule.DemandMultiplier, &rule.Active, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			continue
		}
		rules = append(rules, rule)
	}

	c.JSON(http.StatusOK, rules)
}

// GET /api/v1/rules/:id - Get single pricing rule
func getRule(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	
	engine := services.NewPricingEngine()
	rule, err := engine.GetRule(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// POST /api/v1/rules - Create new pricing rule
func createRule(c *gin.Context) {
	var req models.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate strategy
	validStrategies := map[string]bool{
		"cost_plus":     true,
		"demand_based":  true,
		"competitive":   true,
	}
	if !validStrategies[req.Strategy] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid strategy. Must be: cost_plus, demand_based, or competitive"})
		return
	}

	query := `
		INSERT INTO pricing_rules 
		(name, strategy, base_price, markup_percentage, min_price, max_price, demand_multiplier)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	var rule models.PricingRule
	rule.Name = req.Name
	rule.Strategy = req.Strategy
	rule.BasePrice = req.BasePrice
	rule.MarkupPercentage = req.MarkupPercentage
	rule.MinPrice = req.MinPrice
	rule.MaxPrice = req.MaxPrice
	rule.DemandMultiplier = req.DemandMultiplier
	rule.Active = true

	err := database.DB.QueryRow(
		query,
		req.Name, req.Strategy, req.BasePrice, req.MarkupPercentage,
		req.MinPrice, req.MaxPrice, req.DemandMultiplier,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// PUT /api/v1/rules/:id - Update pricing rule
func updateRule(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	
	var req models.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		UPDATE pricing_rules
		SET name = $1, strategy = $2, base_price = $3, markup_percentage = $4,
		    min_price = $5, max_price = $6, demand_multiplier = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8
		RETURNING id, name, strategy, base_price, markup_percentage, min_price, max_price, 
		          demand_multiplier, active, created_at, updated_at
	`

	var rule models.PricingRule
	err := database.DB.QueryRow(
		query,
		req.Name, req.Strategy, req.BasePrice, req.MarkupPercentage,
		req.MinPrice, req.MaxPrice, req.DemandMultiplier, id,
	).Scan(
		&rule.ID, &rule.Name, &rule.Strategy, &rule.BasePrice,
		&rule.MarkupPercentage, &rule.MinPrice, &rule.MaxPrice,
		&rule.DemandMultiplier, &rule.Active, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DELETE /api/v1/rules/:id - Delete (deactivate) pricing rule
func deleteRule(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	query := `UPDATE pricing_rules SET active = false WHERE id = $1`
	result, err := database.DB.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule deactivated successfully"})
}

// GET /api/v1/calculations - Get price calculation history
func getCalculations(c *gin.Context) {
	limit := c.DefaultQuery("limit", "50")

	query := `
		SELECT id, rule_id, input_data, calculated_price, strategy_used, created_at
		FROM price_calculations
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := database.DB.Query(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var calculations []models.PriceCalculation
	for rows.Next() {
		var calc models.PriceCalculation
		err := rows.Scan(
			&calc.ID, &calc.RuleID, &calc.InputData,
			&calc.CalculatedPrice, &calc.StrategyUsed, &calc.CreatedAt,
		)
		if err != nil {
			continue
		}
		calculations = append(calculations, calc)
	}

	c.JSON(http.StatusOK, calculations)
}