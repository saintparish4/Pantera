package api

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
	"github.com/google/uuid"
	"github.com/saintparish4/harmonia/config"
	"github.com/saintparish4/harmonia/database"
	"github.com/saintparish4/harmonia/internal/domain"
	"github.com/saintparish4/harmonia/internal/handlers"
	"github.com/saintparish4/harmonia/internal/middleware"
	"github.com/saintparish4/harmonia/internal/repository"
	"github.com/saintparish4/harmonia/internal/service"
)

const banner = `
╦ ╦╔═╗╦═╗╔╦╗╔═╗╔╗╔╦╔═╗
║ ║╠═╣╠╦╝║║║║ ║║║║║╠═╣
╩═╝╩ ╩╩╚═╩ ╩╚═╝╝╚╝╩╩ ╩
Dynamic Pricing API v1.0.0

Enterprise-grade pricing for indie devs`

// Dependencies holds all application dependencies
type Dependencies struct {
	// Domain Repositories
	DomainAPIKeyRepo         domain.APIKeyRepository
	DomainUserRepo           domain.UserRepository
	DomainPricingRuleRepo    domain.PricingRuleRepository
	DomainProductRepo        domain.ProductRepository
	DomainCalculationLogRepo domain.CalculationLogRepository

	// Service
	PricingEngine *service.PricingEngine
}

// Server represents the HTTP server
type Server struct {
	router *gin.Engine
	config *config.Config
	deps   *Dependencies
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config, deps *Dependencies) *Server {
	return &Server{
		config: cfg,
		deps:   deps,
	}
}

// Setup configures the router with all middleware and routes
func (s *Server) Setup() {
	// Set Gin mode
	if s.config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	s.router = gin.New()

	// Apply global middleware
	s.setupMiddleware()

	// Setup routes
	s.setupRoutes()
}

// setupMiddleware configures global middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware (must be first)
	s.router.Use(middleware.Recovery())

	// Request logging
	s.router.Use(middleware.RequestLogger())

	// CORS
	corsConfig := s.buildCORSConfig()
	s.router.Use(middleware.CORS(corsConfig))

	// Rate limiting
	rateLimiter := middleware.NewRateLimiter(
		s.config.API.RateLimitRequests,
		time.Duration(s.config.API.RateLimitWindowSeconds)*time.Second,
		s.config.API.RateLimitRequests+50, // burst capacity
	)
	s.router.Use(rateLimiter.Middleware())
}

// buildCORSConfig creates CORS configuration from config
func (s *Server) buildCORSConfig() *middleware.CORSConfig {
	return &middleware.CORSConfig{
		AllowedOrigins:   []string{s.config.Security.CORSOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           3600,
	}
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Create handler-specific adapters
	authMiddleware := middleware.NewAuthMiddleware(&APIKeyValidatorAdapter{repo: s.deps.DomainAPIKeyRepo})
	healthHandler := handlers.NewHealthHandler(&DBHealthChecker{db: database.DB})

	// Create handler-layer adapters
	keysAPIKeyRepo := &HandlerAPIKeyRepo{domainRepo: s.deps.DomainAPIKeyRepo}
	keysUserRepo := &HandlerUserRepo{domainRepo: s.deps.DomainUserRepo}
	keyGenerator := &APIKeyGeneratorImpl{}

	rulesRepo := &HandlerRulesRepo{domainRepo: s.deps.DomainPricingRuleRepo}
	productsRepo := &HandlerProductsRepo{domainRepo: s.deps.DomainProductRepo}
	logsRepo := &HandlerLogsRepo{domainRepo: s.deps.DomainCalculationLogRepo}

	pricingEngineHandler := &HandlerPricingEngine{engine: s.deps.PricingEngine}
	calculationLogger := &HandlerCalculationLogger{domainRepo: s.deps.DomainCalculationLogRepo}

	// Initialize handlers
	keysHandler := handlers.NewKeysHandler(keysAPIKeyRepo, keysUserRepo, keyGenerator)
	pricingHandler := handlers.NewPricingHandler(pricingEngineHandler, calculationLogger)
	rulesHandler := handlers.NewRulesHandler(rulesRepo)
	productsHandler := handlers.NewProductsHandler(productsRepo)
	logsHandler := handlers.NewLogsHandler(logsRepo)

	// Health check (public)
	s.router.GET("/health", healthHandler.Check)

	// Root endpoint
	s.router.GET("/", rootHandler)

	// API v1 routes
	v1 := s.router.Group("/v1")
	{
		// User creation (public - for demo/testing)
		// In production, this might be behind admin auth or part of signup flow
		v1.POST("/users", keysHandler.CreateUser)

		// Auth routes (protected by API key)
		auth := v1.Group("/auth")
		auth.Use(authMiddleware.Authenticate())
		{
			auth.POST("/keys", keysHandler.Create)
			auth.GET("/keys", keysHandler.List)
			auth.DELETE("/keys/:id", keysHandler.Revoke)
		}

		// Pricing routes
		pricing := v1.Group("/pricing")
		{
			// List strategies (public or optionally authenticated)
			pricing.GET("/strategies", authMiddleware.OptionalAuth(), pricingHandler.ListStrategies)

			// Protected pricing endpoints
			pricingAuth := pricing.Group("")
			pricingAuth.Use(authMiddleware.Authenticate())
			{
				// Price calculation
				pricingAuth.POST("/calculate", pricingHandler.Calculate)

				// Pricing rules CRUD
				pricingAuth.GET("/rules", rulesHandler.List)
				pricingAuth.POST("/rules", rulesHandler.Create)
				pricingAuth.GET("/rules/:id", rulesHandler.Get)
				pricingAuth.PUT("/rules/:id", rulesHandler.Update)
				pricingAuth.DELETE("/rules/:id", rulesHandler.Delete)
			}
		}

		// Products routes (protected)
		products := v1.Group("/products")
		products.Use(authMiddleware.Authenticate())
		{
			products.GET("", productsHandler.List)
			products.POST("", productsHandler.Create)
			products.GET("/:id", productsHandler.Get)
			products.PUT("/:id", productsHandler.Update)
			products.DELETE("/:id", productsHandler.Delete)
		}

		// Logs routes (protected)
		logs := v1.Group("/logs")
		logs.Use(authMiddleware.Authenticate())
		{
			logs.GET("", logsHandler.List)
		}
	}

	// 404 handler
	s.router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "The requested resource was not found",
			"code":    "NOT_FOUND",
		})
	})
}

// Start starts the HTTP server with graceful shutdown
func (s *Server) Start() error {
	srv := &http.Server{
		Addr:           ":" + s.config.Server.Port,
		Handler:        s.router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Channel to listen for errors from the server
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server listening on :%s", s.config.Server.Port)
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or error
	select {
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-shutdown:
		log.Printf("Received signal: %v. Starting graceful shutdown...", sig)

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
			// Force close if graceful shutdown fails
			if closeErr := srv.Close(); closeErr != nil {
				return fmt.Errorf("server forced close error: %w", closeErr)
			}
			return fmt.Errorf("server shutdown error: %w", err)
		}

		log.Println("✓ Server stopped gracefully")
	}

	return nil
}

// StartServer is the main entry point for the application
func StartServer() {
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

	// Initialize dependencies
	deps := initializeDependencies()

	// Create and setup server
	server := NewServer(cfg, deps)
	server.Setup()

	// Start server
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// initializeDependencies creates all repositories and services
func initializeDependencies() *Dependencies {
	// Initialize domain repositories
	domainAPIKeyRepo := repository.NewAPIKeyRepository(database.DB)
	domainUserRepo := repository.NewUserRepository(database.DB)
	domainPricingRuleRepo := repository.NewPricingRuleRepository(database.DB)
	domainProductRepo := repository.NewProductRepository(database.DB)
	domainCalculationLogRepo := repository.NewCalculationLogRepository(database.DB)

	// Initialize services
	pricingEngine := service.NewPricingEngine()

	return &Dependencies{
		DomainAPIKeyRepo:         domainAPIKeyRepo,
		DomainUserRepo:           domainUserRepo,
		DomainPricingRuleRepo:    domainPricingRuleRepo,
		DomainProductRepo:        domainProductRepo,
		DomainCalculationLogRepo: domainCalculationLogRepo,
		PricingEngine:            pricingEngine,
	}
}

// Adapter implementations to bridge domain and handler layers

// APIKeyValidatorAdapter adapts domain.APIKeyRepository for middleware authentication
type APIKeyValidatorAdapter struct {
	repo domain.APIKeyRepository
}

func (a *APIKeyValidatorAdapter) ValidateKey(ctx context.Context, keyHash string) (uuid.UUID, bool, error) {
	key, err := a.repo.GetByHash(ctx, repository.HashAPIKey(keyHash))
	if err != nil {
		return uuid.Nil, false, err
	}
	if !key.IsActive {
		return uuid.Nil, false, nil
	}
	// Update last used timestamp (fire and forget)
	go a.repo.UpdateLastUsed(context.Background(), key.ID)
	return key.UserID, true, nil
}

// DBHealthChecker wraps database.DB to implement handlers.HealthChecker
type DBHealthChecker struct {
	db interface {
		PingContext(ctx context.Context) error
	}
}

func (h *DBHealthChecker) Ping(ctx context.Context) error {
	return h.db.PingContext(ctx)
}

// APIKeyGeneratorImpl implements handlers.KeyGenerator
type APIKeyGeneratorImpl struct{}

func (g *APIKeyGeneratorImpl) GenerateAPIKey(isLive bool) (rawKey string, keyHash string, keyPrefix string, err error) {
	rawKey, err = repository.GenerateAPIKey(isLive)
	if err != nil {
		return "", "", "", err
	}
	keyHash = repository.HashAPIKey(rawKey)
	keyPrefix = repository.GetKeyPrefix(rawKey)
	return rawKey, keyHash, keyPrefix, nil
}

// Handler layer repository adapters

// HandlerAPIKeyRepo adapts domain.APIKeyRepository to handlers.APIKeyRepository
type HandlerAPIKeyRepo struct {
	domainRepo domain.APIKeyRepository
}

func (r *HandlerAPIKeyRepo) Create(ctx context.Context, key *handlers.APIKey) error {
	domainKey := &domain.APIKey{
		ID:        key.ID,
		UserID:    key.UserID,
		KeyHash:   key.KeyHash,
		KeyPrefix: key.KeyPrefix,
		Name:      key.Name,
		IsActive:  true, // New keys are always active
	}
	return r.domainRepo.Create(ctx, domainKey)
}

func (r *HandlerAPIKeyRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*handlers.APIKey, error) {
	domainKeys, err := r.domainRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	handlerKeys := make([]*handlers.APIKey, len(domainKeys))
	for i, dk := range domainKeys {
		handlerKeys[i] = &handlers.APIKey{
			ID:        dk.ID,
			UserID:    dk.UserID,
			KeyHash:   dk.KeyHash,
			KeyPrefix: dk.KeyPrefix,
			Name:      dk.Name,
			CreatedAt: dk.CreatedAt.Format(time.RFC3339),
		}
	}
	return handlerKeys, nil
}

func (r *HandlerAPIKeyRepo) Revoke(ctx context.Context, keyID uuid.UUID) error {
	return r.domainRepo.Revoke(ctx, keyID)
}

// HandlerUserRepo adapts domain.UserRepository to handlers.UserRepository
type HandlerUserRepo struct {
	domainRepo domain.UserRepository
}

func (r *HandlerUserRepo) Create(ctx context.Context, user *handlers.User) error {
	domainUser := &domain.User{
		ID:    user.ID,
		Email: user.Email,
	}
	return r.domainRepo.Create(ctx, domainUser)
}

func (r *HandlerUserRepo) GetByEmail(ctx context.Context, email string) (*handlers.User, error) {
	domainUser, err := r.domainRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &handlers.User{
		ID:    domainUser.ID,
		Email: domainUser.Email,
	}, nil
}

// HandlerRulesRepo adapts domain.PricingRuleRepository to handlers.PricingRuleRepository
type HandlerRulesRepo struct {
	domainRepo domain.PricingRuleRepository
}

func (r *HandlerRulesRepo) Create(ctx context.Context, rule *handlers.PricingRule) error {
	domainRule := &domain.PricingRule{
		ID:           rule.ID,
		UserID:       rule.UserID,
		Name:         rule.Name,
		StrategyType: rule.StrategyType,
		Config:       rule.Config,
	}
	return r.domainRepo.Create(ctx, domainRule)
}

func (r *HandlerRulesRepo) GetByID(ctx context.Context, id uuid.UUID) (*handlers.PricingRule, error) {
	domainRule, err := r.domainRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &handlers.PricingRule{
		ID:           domainRule.ID,
		UserID:       domainRule.UserID,
		Name:         domainRule.Name,
		StrategyType: domainRule.StrategyType,
		Config:       domainRule.Config,
	}, nil
}

func (r *HandlerRulesRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*handlers.PricingRule, error) {
	domainRules, err := r.domainRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	handlerRules := make([]*handlers.PricingRule, len(domainRules))
	for i, dr := range domainRules {
		handlerRules[i] = &handlers.PricingRule{
			ID:           dr.ID,
			UserID:       dr.UserID,
			Name:         dr.Name,
			StrategyType: dr.StrategyType,
			Config:       dr.Config,
		}
	}
	return handlerRules, nil
}

func (r *HandlerRulesRepo) Update(ctx context.Context, rule *handlers.PricingRule) error {
	domainRule := &domain.PricingRule{
		ID:           rule.ID,
		UserID:       rule.UserID,
		Name:         rule.Name,
		StrategyType: rule.StrategyType,
		Config:       rule.Config,
	}
	return r.domainRepo.Update(ctx, domainRule)
}

func (r *HandlerRulesRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.domainRepo.Delete(ctx, id)
}

// HandlerProductsRepo adapts domain.ProductRepository to handlers.ProductRepository
type HandlerProductsRepo struct {
	domainRepo domain.ProductRepository
}

func (r *HandlerProductsRepo) Create(ctx context.Context, product *handlers.Product) error {
	domainProduct := &domain.Product{
		ID:          product.ID,
		UserID:      product.UserID,
		SKU:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		BaseCost:    0, // Handler Product doesn't have base cost
	}
	return r.domainRepo.Create(ctx, domainProduct)
}

func (r *HandlerProductsRepo) GetByID(ctx context.Context, id uuid.UUID) (*handlers.Product, error) {
	domainProduct, err := r.domainRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &handlers.Product{
		ID:          domainProduct.ID,
		UserID:      domainProduct.UserID,
		SKU:         domainProduct.SKU,
		Name:        domainProduct.Name,
		Description: domainProduct.Description,
	}, nil
}

func (r *HandlerProductsRepo) GetBySKU(ctx context.Context, userID uuid.UUID, sku string) (*handlers.Product, error) {
	domainProduct, err := r.domainRepo.GetBySKU(ctx, userID, sku)
	if err != nil {
		return nil, err
	}
	return &handlers.Product{
		ID:          domainProduct.ID,
		UserID:      domainProduct.UserID,
		SKU:         domainProduct.SKU,
		Name:        domainProduct.Name,
		Description: domainProduct.Description,
	}, nil
}

func (r *HandlerProductsRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*handlers.Product, error) {
	domainProducts, err := r.domainRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	handlerProducts := make([]*handlers.Product, len(domainProducts))
	for i, dp := range domainProducts {
		handlerProducts[i] = &handlers.Product{
			ID:          dp.ID,
			UserID:      dp.UserID,
			SKU:         dp.SKU,
			Name:        dp.Name,
			Description: dp.Description,
		}
	}
	return handlerProducts, nil
}

func (r *HandlerProductsRepo) Update(ctx context.Context, product *handlers.Product) error {
	domainProduct := &domain.Product{
		ID:          product.ID,
		UserID:      product.UserID,
		SKU:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		BaseCost:    0, // Handler Product doesn't have base cost
	}
	return r.domainRepo.Update(ctx, domainProduct)
}

func (r *HandlerProductsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.domainRepo.Delete(ctx, id)
}

// HandlerLogsRepo adapts domain.CalculationLogRepository to handlers.CalculationLogRepository
type HandlerLogsRepo struct {
	domainRepo domain.CalculationLogRepository
}

func (r *HandlerLogsRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, strategyType string, startDate, endDate *time.Time) ([]*handlers.CalculationLog, error) {
	filter := domain.CalculationLogFilter{
		UserID:       userID,
		StrategyType: strategyType,
		Limit:        limit,
		Offset:       offset,
	}
	if startDate != nil {
		filter.From = *startDate
	}
	if endDate != nil {
		filter.To = *endDate
	}

	domainLogs, err := r.domainRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	handlerLogs := make([]*handlers.CalculationLog, len(domainLogs))
	for i, dl := range domainLogs {
		handlerLogs[i] = &handlers.CalculationLog{
			ID:           dl.ID,
			UserID:       dl.UserID,
			StrategyType: dl.StrategyType,
			Input:        dl.InputData,
			Output:       dl.OutputData,
			CreatedAt:    dl.CreatedAt,
		}
	}
	return handlerLogs, nil
}

func (r *HandlerLogsRepo) Count(ctx context.Context, userID uuid.UUID, strategyType string, startDate, endDate *time.Time) (int, error) {
	filter := domain.CalculationLogFilter{
		UserID:       userID,
		StrategyType: strategyType,
		Limit:        -1,
	}
	if startDate != nil {
		filter.From = *startDate
	}
	if endDate != nil {
		filter.To = *endDate
	}

	domainLogs, err := r.domainRepo.List(ctx, filter)
	if err != nil {
		return 0, err
	}
	return len(domainLogs), nil
}

// HandlerPricingEngine adapts service.PricingEngine to handlers.PricingEngine
type HandlerPricingEngine struct {
	engine *service.PricingEngine
}

func (e *HandlerPricingEngine) Calculate(ctx context.Context, req *handlers.PricingRequest) (*handlers.PricingResult, error) {
	domainReq := &domain.PricingRequest{
		Strategy: req.StrategyType,
		Inputs:   req.Context, // Map handler Context to domain Inputs
	}

	response, err := e.engine.Calculate(domainReq, req.Context)
	if err != nil {
		return nil, err
	}

	return &handlers.PricingResult{
		FinalPrice: response.FinalPrice,
		Breakdown: map[string]interface{}{
			"strategy":      response.Strategy,
			"calculated_at": response.CalculatedAt,
			"breakdown":     response.Breakdown,
		},
	}, nil
}

func (e *HandlerPricingEngine) GetAvailableStrategies() []string {
	return e.engine.ListStrategies()
}

// HandlerCalculationLogger adapts domain.CalculationLogRepository to handlers.CalculationLogger
type HandlerCalculationLogger struct {
	domainRepo domain.CalculationLogRepository
}

func (l *HandlerCalculationLogger) Log(ctx context.Context, userID uuid.UUID, strategyType string, input, output map[string]interface{}) error {
	domainLog := &domain.CalculationLog{
		ID:           uuid.New(),
		UserID:       userID,
		StrategyType: strategyType,
		InputData:    input,
		OutputData:   output,
		CreatedAt:    time.Now(),
	}
	return l.domainRepo.Create(ctx, domainLog)
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
        .put { background: #f59e0b; }
        .delete { background: #ef4444; }
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
            <span class="method post">POST</span> /v1/users - Create user account
        </div>
        <div class="endpoint">
            <span class="method post">POST</span> /v1/pricing/calculate - Calculate price
        </div>
        <div class="endpoint">
            <span class="method get">GET</span> /v1/pricing/rules - List pricing rules
        </div>
        <div class="endpoint">
            <span class="method get">GET</span> /v1/products - List products
        </div>
        
        <h2>Authentication</h2>
        <p>Include your API key in requests using the <code>X-API-Key</code> header or <code>Authorization: Bearer</code> header.</p>
        
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
			"users":     "POST /v1/users",
			"auth":      "POST /v1/auth/keys",
			"calculate": "POST /v1/pricing/calculate",
			"rules":     "GET /v1/pricing/rules",
			"products":  "GET /v1/products",
			"logs":      "GET /v1/logs",
			"docs":      "https://github.com/saintparish4/harmonia",
		},
	})
}
