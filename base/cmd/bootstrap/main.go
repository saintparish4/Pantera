package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/saintparish4/harmonia/config"
	"github.com/saintparish4/harmonia/database"
	"github.com/saintparish4/harmonia/internal/domain"
	"github.com/saintparish4/harmonia/internal/repository"
)

func main() {
	email := flag.String("email", "", "User email address (required)")
	name := flag.String("name", "Initial API Key", "API key name")
	isLive := flag.Bool("live", false, "Generate live key (default: test key)")
	flag.Parse()

	if *email == "" {
		log.Fatal("Error: --email is required\n\nUsage: go run cmd/bootstrap/main.go --email your@email.com [--name \"My Key\"] [--live]")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.Migrate("./database/migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	ctx := context.Background()

	// Initialize repositories
	userRepo := repository.NewUserRepository(database.DB)
	apiKeyRepo := repository.NewAPIKeyRepository(database.DB)

	// Check if user exists
	existingUser, _ := userRepo.GetByEmail(ctx, *email)
	var user *domain.User

	if existingUser != nil {
		user = existingUser
		fmt.Printf("✓ User already exists: %s\n", user.Email)
	} else {
		// Create new user
		user = &domain.User{
			ID:        uuid.New(),
			Email:     *email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := userRepo.Create(ctx, user); err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
		fmt.Printf("✓ Created user: %s\n", user.Email)
	}

	// Generate API key
	rawKey, err := repository.GenerateAPIKey(*isLive)
	if err != nil {
		log.Fatalf("Failed to generate API key: %v", err)
	}

	keyHash := repository.HashAPIKey(rawKey)
	keyPrefix := repository.GetKeyPrefix(rawKey)

	// Create API key record
	apiKey := &domain.APIKey{
		ID:        uuid.New(),
		UserID:    user.ID,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		Name:      *name,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if err := apiKeyRepo.Create(ctx, apiKey); err != nil {
		log.Fatalf("Failed to create API key: %v", err)
	}

	// Display results
	separator := strings.Repeat("─", 60)
	fmt.Println("\n" + separator)
	fmt.Println("🎉 Bootstrap Complete!")
	fmt.Println(separator)
	fmt.Printf("User ID:    %s\n", user.ID)
	fmt.Printf("Email:      %s\n", user.Email)
	fmt.Printf("\nAPI Key ID: %s\n", apiKey.ID)
	fmt.Printf("Name:       %s\n", apiKey.Name)
	fmt.Printf("Type:       %s\n", map[bool]string{true: "Live", false: "Test"}[*isLive])
	fmt.Println(separator)
	fmt.Printf("\n🔑 YOUR API KEY (save this securely):\n\n")
	fmt.Printf("    %s\n\n", rawKey)
	fmt.Println("⚠️  This key will not be shown again!")
	fmt.Println(separator)
	fmt.Printf("\nTest your API key:\n\n")
	fmt.Printf("  curl -H \"X-API-Key: %s\" http://localhost:8080/v1/pricing/strategies\n\n", rawKey)
}
