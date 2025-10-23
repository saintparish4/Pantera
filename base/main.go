package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/saintparish4/pantera/database"
	"github.com/saintparish4/pantera/routes"
)

func main() {
	// Load environment variables from parent directory
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, trying current directory...")
		godotenv.Load() // Try current directory as fallback
	}

	// Initialize database
	database.Connect()
	database.Migrate()

	// Setup Gin router
	router := gin.Default()

	// CORS middleware for Frontend
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":        "Pantera Dynamic Pricing API",
			"version":     "1.0.0",
			"status":      "live",
			"description": "Multi-strategy pricing engine",
			"endpoints": gin.H{
				"health":       "/health",
				"rules":        "/api/v1/rules",
				"calculate":    "POST /api/v1/calculate",
				"calculations": "/api/v1/calculations",
			},
			"strategies": []string{
				"cost_plus",
				"geographic",
				"gemstone",
			},
		})
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		if err := database.DB.Ping(); err != nil {
			c.JSON(500, gin.H{
				"status":   "unhealthy",
				"database": "disconnected",
				"error":    err.Error(),
			})
			return
		}
		c.JSON(200, gin.H{
			"service":  "base-api",
			"status":   "healthy",
			"version":  "1.0.0",
			"database": "connected",
		})
	})

	// API routes
	routes.SetupPricingRoutes(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf(" BASE API running on port %s", port)
	router.Run(":" + port)
}
