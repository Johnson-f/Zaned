package database

import (
	"fmt"
	"net/url"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes the database connection using GORM
func InitDB() error {
	// Get database connection string from environment
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	// Parse URL to ensure proper encoding
	parsedURL, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}

	// Get existing query parameters
	query := parsedURL.Query()
	
	// Add/update connection parameters for better compatibility
	query.Set("sslmode", "require")
	query.Set("connect_timeout", "10")
	
	// Reconstruct URL with updated query parameters
	parsedURL.RawQuery = query.Encode()
	dsn = parsedURL.String()

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		// Provide more helpful error message
		return fmt.Errorf("failed to connect to database: %w\n\nTroubleshooting:\n1. Verify DATABASE_URL is correct in .env file\n2. Check if Supabase project is active (not paused)\n3. Verify network connectivity\n4. Check Supabase IP allowlist settings if enabled", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	return nil
}

// Migrate runs database migrations for the given models
func Migrate(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	return DB.AutoMigrate(models...)
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

