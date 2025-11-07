package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saintparish4/harmonia/config"
	"github.com/saintparish4/harmonia/database"
)

const banner = `
╦ ╦╔═╗╦═╗╔╦╗╔═╗╔╗╔╦╔═╗
║ ║╠═╣╠╦╝║║║║ ║║║║║╠═╣
╩═╝╩ ╩╩╚═╩ ╩╚═╝╝╚╝╩╩ ╩
Dynamic Pricing API v1.0.0

Enterprise-grade pricing for indie devs`

func main() {
	// Print banner
	fmt.Println(banner)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Harmonia API [%s mode]", cfg.Server.Environment)

	// Connect to database
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.Migrate("./database/migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := setupRouter(cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server listening on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("✓ Server exited gracefully")
}

func setupRouter(cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware(cfg))

	// Health check endpoint (no authentication required)
	router.GET("/health", healthCheckHandler)

	// Root endpoint
	router.GET("/", rootHandler)

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// TODO: Add API routes in Layer 4
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
				"version": "v1",
			})
		})
	}

	return router
}

// corsMiddleware configures CORS
func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.Security.CORSOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Methods", cfg.Security.CORSMethods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", cfg.Security.CORSHeaders)
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// healthCheckHandler returns database health status
func healthCheckHandler(c *gin.Context) {
	if err := database.HealthCheck(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"message": "Database connection failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "All systems operational",
		"uptime":  time.Now().Unix(),
	})
}

// rootHandler returns API information
func rootHandler(c *gin.Context) {
	acceptHeader := c.GetHeader("Accept")

	// Return HTML if Accept header includes text/html
	if acceptHeader != "" && (acceptHeader == "text/html" || acceptHeader == "*/*") {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html>
<head>
    <title>Harmonia API</title>
    <style>
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; 
            max-width: 800px; 
            margin: 50px auto; 
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
        }
        .container {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            padding: 40px;
            border-radius: 15px;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
        }
        h1 { font-size: 2.5em; margin-bottom: 10px; }
        .tagline { font-size: 1.2em; opacity: 0.9; margin-bottom: 30px; }
        .endpoint { 
            background: rgba(255, 255, 255, 0.1); 
            padding: 15px; 
            margin: 10px 0; 
            border-radius: 8px;
            font-family: 'Courier New', monospace;
        }
        .method { 
            display: inline-block;
            padding: 4px 12px;
            border-radius: 4px;
            font-weight: bold;
            margin-right: 10px;
        }
        .get { background: #10b981; }
        .post { background: #3b82f6; }
        a { color: #fbbf24; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎵 Harmonia API</h1>
        <p class="tagline">Enterprise-grade dynamic pricing for indie developers</p>
        
        <h2>Quick Start</h2>
        <div class="endpoint">
            <span class="method get">GET</span> /health - Health check
        </div>
        <div class="endpoint">
            <span class="method post">POST</span> /v1/calculate - Calculate price
        </div>
        <div class="endpoint">
            <span class="method get">GET</span> /v1/rules - List pricing rules
        </div>
        
        <p style="margin-top: 30px;">
            📖 <a href="https://github.com/saintparish4/harmonia">Documentation</a> | 
            🐛 <a href="https://github.com/saintparish4/harmonia/issues">Report Issues</a>
        </p>
    </div>
</body>
</html>
		`)
		return
	}

	// Return JSON for API clients
	c.JSON(http.StatusOK, gin.H{
		"name":    "Harmonia API",
		"version": "1.0.0",
		"tagline": "Enterprise-grade dynamic pricing for indie developers",
		"endpoints": gin.H{
			"health":    "GET /health",
			"calculate": "POST /v1/calculate",
			"rules":     "GET /v1/rules",
			"docs":      "https://github.com/saintparish4/harmonia",
		},
	})
}
