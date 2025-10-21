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
		api.POST("/calculate", validatePriceRequest(), calculatePrice)

		// Pricing rules management
		api.GET("/rules", getRules)
		api.GET("/rules/:id", validateIDParam(), getRule)
		api.POST("/rules", validateCreateRule(), createRule)
		api.PUT("/rules/:id", validateIDParam(), validateCreateRule(), updateRule)
		api.DELETE("/rules/:id", validateIDParam(), deleteRule)

		// Analytics/audit
		api.GET("/calculations", getCalculations)
	}
}

// ============================================
// Validation Middleware
// ============================================

// validatePriceRequest validates the price calculation request
func validatePriceRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.PriceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate rule_id is provided
		if req.RuleID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "rule_id is required and must be greater than 0",
			})
			c.Abort()
			return
		}

		// Validate demand_level (if provided)
		if req.DemandLevel < 0 || req.DemandLevel > 2 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "demand_level must be between 0.0 and 2.0",
				"hint":  "0.0 = very low demand, 1.0 = normal, 2.0 = extreme demand",
			})
			c.Abort()
			return
		}

		// Validate quantity (if provided)
		if req.Quantity < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "quantity cannot be negative",
			})
			c.Abort()
			return
		}

		// Validate competitor_price (if provided)
		if req.CompetitorPrice < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "competitor_price cannot be negative",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// validateCreateRule validates pricing rule creation/update
func validateCreateRule() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateRuleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate strategy
		validStrategies := map[string]bool{
			"cost_plus":    true,
			"demand_based": true,
			"competitive":  true,
		}
		if !validStrategies[req.Strategy] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":            "Invalid pricing strategy",
				"valid_strategies": []string{"cost_plus", "demand_based", "competitive"},
			})
			c.Abort()
			return
		}

		// Validate base_price
		if req.BasePrice <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "base_price must be greater than 0",
			})
			c.Abort()
			return
		}

		// Validate min/max price relationship
		if req.MinPrice > 0 && req.MaxPrice > 0 && req.MinPrice > req.MaxPrice {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "min_price cannot be greater than max_price",
			})
			c.Abort()
			return
		}

		// Validate markup_percentage for cost_plus strategy
		if req.Strategy == "cost_plus" && req.MarkupPercentage == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "markup_percentage is required for cost_plus strategy",
			})
			c.Abort()
			return
		}

		// Validate name is not empty
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "name cannot be empty",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// validateIDParam validates URL parameter ID
func validateIDParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)

		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID parameter",
				"hint":  "ID must be a positive integer",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ============================================
// Route Handlers
// ============================================

// POST /api/v1/calculate - Calculate price based on rule
func calculatePrice(c *gin.Context) {
	var req models.PriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to parse request body",
			"details": err.Error(),
		})
		return
	}

	engine := services.NewPricingEngine()
	response, err := engine.CalculatePrice(req)
	if err != nil {
		// Check for specific error types
		if err.Error() == "rule not found: sql: no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Pricing rule not found",
				"rule_id": req.RuleID,
				"hint":    "Check that the rule_id exists using GET /api/v1/rules",
			})
			return
		}
		if err.Error() == "pricing rule is not active" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Pricing rule is not active",
				"rule_id": req.RuleID,
				"hint":    "This rule has been deactivated and cannot be used for calculations",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /api/v1/rules - Get all pricing rules
func getRules(c *gin.Context) {
	// Optional filter by active status
	activeOnly := c.Query("active") == "true"

	query := `
		SELECT id, name, strategy, base_price, markup_percentage,
		       min_price, max_price, demand_multiplier, active, created_at, updated_at
		FROM pricing_rules
	`

	if activeOnly {
		query += " WHERE active = true"
	}

	query += " ORDER BY created_at DESC"

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch pricing rules",
			"details": err.Error(),
		})
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to parse pricing rule",
				"details": err.Error(),
			})
			return
		}
		rules = append(rules, rule)
	}

	// Handle empty results
	if rules == nil {
		rules = []models.PricingRule{}
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(rules),
		"rules": rules,
	})
}

// GET /api/v1/rules/:id - Get single pricing rule
func getRule(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	engine := services.NewPricingEngine()
	rule, err := engine.GetRule(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Pricing rule not found",
			"rule_id": id,
		})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// POST /api/v1/rules - Create new pricing rule
func createRule(c *gin.Context) {
	var req models.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to parse request body",
			"details": err.Error(),
		})
		return
	}

	// Set defaults
	if req.DemandMultiplier == 0 {
		req.DemandMultiplier = 1.0
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create pricing rule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pricing rule created successfully",
		"rule":    rule,
	})
}

// PUT /api/v1/rules/:id - Update pricing rule
func updateRule(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req models.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to parse request body",
			"details": err.Error(),
		})
		return
	}

	// Check if rule exists first
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM pricing_rules WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Pricing rule not found",
			"rule_id": id,
		})
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
	err = database.DB.QueryRow(
		query,
		req.Name, req.Strategy, req.BasePrice, req.MarkupPercentage,
		req.MinPrice, req.MaxPrice, req.DemandMultiplier, id,
	).Scan(
		&rule.ID, &rule.Name, &rule.Strategy, &rule.BasePrice,
		&rule.MarkupPercentage, &rule.MinPrice, &rule.MaxPrice,
		&rule.DemandMultiplier, &rule.Active, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update pricing rule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pricing rule updated successfully",
		"rule":    rule,
	})
}

// DELETE /api/v1/rules/:id - Delete (deactivate) pricing rule
func deleteRule(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	// Check if rule exists and is already inactive
	var active bool
	err := database.DB.QueryRow("SELECT active FROM pricing_rules WHERE id = $1", id).Scan(&active)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Pricing rule not found",
			"rule_id": id,
		})
		return
	}

	if !active {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Pricing rule is already deactivated",
			"rule_id": id,
		})
		return
	}

	query := `UPDATE pricing_rules SET active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	result, err := database.DB.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to deactivate pricing rule",
			"details": err.Error(),
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Pricing rule not found",
			"rule_id": id,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pricing rule deactivated successfully",
		"rule_id": id,
		"note":    "This rule can no longer be used for price calculations",
	})
}

// GET /api/v1/calculations - Get price calculation history
func getCalculations(c *gin.Context) {
	// Parse query parameters
	limit := c.DefaultQuery("limit", "50")
	ruleID := c.Query("rule_id")

	// Validate limit
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 || limitInt > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid limit parameter",
			"hint":  "Limit must be between 1 and 100",
		})
		return
	}

	query := `
		SELECT id, rule_id, input_data, calculated_price, strategy_used, created_at
		FROM price_calculations
	`

	args := []interface{}{}

	// Filter by rule_id if provided
	if ruleID != "" {
		ruleIDInt, err := strconv.Atoi(ruleID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid rule_id parameter",
			})
			return
		}
		query += " WHERE rule_id = $1 ORDER BY created_at DESC LIMIT $2"
		args = append(args, ruleIDInt, limitInt)
	} else {
		query += " ORDER BY created_at DESC LIMIT $1"
		args = append(args, limitInt)
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch calculations",
			"details": err.Error(),
		})
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to parse calculation",
				"details": err.Error(),
			})
			return
		}
		calculations = append(calculations, calc)
	}

	// Handle empty results
	if calculations == nil {
		calculations = []models.PriceCalculation{}
	}

	c.JSON(http.StatusOK, gin.H{
		"count":        len(calculations),
		"limit":        limitInt,
		"calculations": calculations,
	})
}
