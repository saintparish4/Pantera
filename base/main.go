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

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "base-api",
			"version": "1.0.0",
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
