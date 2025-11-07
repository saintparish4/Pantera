package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	API      APIConfig
	Security SecurityConfig
	Logging  LoggingConfig
}

type ServerConfig struct {
	Port        string
	Environment string // development, staging, production
}

type DatabaseConfig struct {
	URL             string
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type APIConfig struct {
	Version               string
	RateLimitRequests     int
	RateLimitWindowSeconds int
}

type SecurityConfig struct {
	JWTSecret    string
	AdminAPIKey  string
	CORSOrigins  string
	CORSMethods  string
	CORSHeaders  string
}

type LoggingConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error in production)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8080"),
			Environment: getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", ""),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "harmonia"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		API: APIConfig{
			Version:               getEnv("API_VERSION", "v1"),
			RateLimitRequests:     getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindowSeconds: getEnvAsInt("RATE_LIMIT_WINDOW", 3600),
		},
		Security: SecurityConfig{
			JWTSecret:   getEnv("JWT_SECRET", ""),
			AdminAPIKey: getEnv("ADMIN_API_KEY", ""),
			CORSOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
			CORSMethods: getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
			CORSHeaders: getEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	// Database connection required
	if c.Database.URL == "" && c.Database.Host == "" {
		return fmt.Errorf("database configuration missing: set DATABASE_URL or DB_HOST")
	}

	// Security keys required in production
	if c.Server.Environment == "production" {
		if c.Security.JWTSecret == "" {
			return fmt.Errorf("JWT_SECRET is required in production")
		}
		if c.Security.AdminAPIKey == "" {
			return fmt.Errorf("ADMIN_API_KEY is required in production")
		}
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	if c.Database.URL != "" {
		return c.Database.URL
	}

	// Build DSN from individual components
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}