package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	var err error
	connStr := os.Getenv("DATABASE_URL")

	if connStr == "" {
		// Fallback to individual env vars
		connStr = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
		)
	}

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Successfully connected to database")
}

func Migrate() {
	// Create pricing_rules table
	pricingRulesTable := `
	CREATE TABLE IF NOT EXISTS pricing_rules (
	id SERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	strategy VARCHAR(50) NOT NULL,
	base_price DECIMAL(10,2),
	markup_percentage DECIMAL(5,2),
	min_price DECIMAL(10,2),
	max_price DECIMAL(10,2),
	demand_multiplier DECIMAL(5,2) DEFAULT 1.0,
	active BOOLEAN DEFAULT true,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create price_calculations table (audit log)
	priceCalculationsTable := `
	CREATE TABLE IF NOT EXISTS price_calculations (
	id SERIAL PRIMARY KEY,
	rule_id INTEGER REFERENCES pricing_rules(id),
	input_data JSONB,
	calculated_price DECIMAL(10,2),
	strategy_used VARCHAR(50),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create index for faster queries
	indexes := `
	CREATE INDEX IF NOT EXISTS idx_pricing_rules_active ON pricing_rules(active);
	CREATE INDEX IF NOT EXISTS idx_price_calculations_rule_id ON price_calculations(rule_id);
	`
	_, err := DB.Exec(pricingRulesTable)
	if err != nil {
		log.Fatalf("Failed to create pricing_rules table: %v", err)
	}

	_, err = DB.Exec(priceCalculationsTable)
	if err != nil {
		log.Fatalf("Failed to create price_calculations table: %v", err)
	}

	_, err = DB.Exec(indexes)
	if err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	log.Println("Database migrated successfully")
}
