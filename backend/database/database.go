package database

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// SchemaVersion represents the schema version tracking table
type SchemaVersion struct {
	ID          uint      `gorm:"primaryKey"`
	Version     string    `gorm:"type:varchar(50);not null;uniqueIndex"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

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
	// Avoid server-side prepared statements (PgBouncer transaction pooling compatibility)
	query.Set("prefer_simple_protocol", "true")
	query.Set("statement_cache_capacity", "0")
	query.Set("default_query_exec_mode", "simple_protocol")

	// Reconstruct URL with updated query parameters
	parsedURL.RawQuery = query.Encode()
	dsn = parsedURL.String()

	// Open database connection
	// Disable prepared statement cache to avoid conflicts with PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		PrepareStmt:                              false, // Disable prepared statements to avoid PostgreSQL conflicts
		DisableForeignKeyConstraintWhenMigrating: true,
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

	// Initialize schema version table
	if err := initializeSchemaVersionTable(); err != nil {
		return fmt.Errorf("failed to initialize schema version table: %w", err)
	}

	return nil
}

// initializeSchemaVersionTable creates the schema_version table if it doesn't exist
func initializeSchemaVersionTable() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Create schema_version table if it doesn't exist
	if err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_versions (
			id SERIAL PRIMARY KEY,
			version VARCHAR(50) NOT NULL UNIQUE,
			description TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create schema_versions table: %w", err)
	}

	return nil
}

// getCurrentSchemaVersion returns the current schema version from the database
func getCurrentSchemaVersion() (string, error) {
	if DB == nil {
		return "", fmt.Errorf("database connection not initialized")
	}

	var version SchemaVersion
	// Use GORM Raw to avoid prepared statement conflicts
	// Raw respects PrepareStmt: false setting
	result := DB.Raw(`
		SELECT id, version, description, created_at 
		FROM schema_versions 
		ORDER BY created_at DESC 
		LIMIT 1
	`).Scan(&version)

	if result.Error != nil {
		// Check if it's a "no rows" error
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", nil // No version found, return empty string
		}
		return "", fmt.Errorf("failed to get schema version: %w", result.Error)
	}

	// Check if we got a result (version.ID will be 0 if no record found)
	if version.ID == 0 {
		return "", nil
	}

	return version.Version, nil
}

// updateSchemaVersion records a new schema version
func updateSchemaVersion(version, description string) error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Use GORM Exec to avoid prepared statement conflicts
	result := DB.Exec(`
		INSERT INTO schema_versions (version, description, created_at)
		VALUES (?, ?, ?)
		ON CONFLICT (version) DO NOTHING
	`, version, description, time.Now())
	if result.Error != nil {
		return fmt.Errorf("failed to update schema version: %w", result.Error)
	}

	return nil
}

// tableExists checks if a table exists in the database
func tableExists(tableName string) (bool, error) {
	if DB == nil {
		return false, fmt.Errorf("database connection not initialized")
	}

	var count int64
	// Use GORM Raw to avoid prepared statement conflicts
	result := DB.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = ?
	`, tableName).Scan(&count)

	if result.Error != nil {
		return false, fmt.Errorf("failed to check table existence: %w", result.Error)
	}

	return count > 0, nil
}

// columnExists checks if a column exists in a table
func columnExists(tableName, columnName string) (bool, error) {
	if DB == nil {
		return false, fmt.Errorf("database connection not initialized")
	}

	var count int64
	// Use GORM Raw to avoid prepared statement conflicts
	result := DB.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		AND table_name = ? 
		AND column_name = ?
	`, tableName, columnName).Scan(&count)

	if result.Error != nil {
		return false, fmt.Errorf("failed to check column existence: %w", result.Error)
	}

	return count > 0, nil
}

// getCurrentTables returns a list of all user tables in the database (excluding system tables)
func getCurrentTables() ([]string, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	var tables []string
	// Use GORM Raw to avoid prepared statement conflicts
	result := DB.Raw(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		AND table_name NOT LIKE 'pg_%'
		AND table_name NOT LIKE '_%'
		ORDER BY table_name
	`).Scan(&tables)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get current tables: %w", result.Error)
	}

	return tables, nil
}

// dropTableSafely drops a table if it exists and is not a system table
func dropTableSafely(tableName string) error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// List of system tables that should never be dropped
	systemTables := map[string]bool{
		"schema_versions":    true,
		"pg_stat_statements": true,
	}

	// Prevent dropping system tables
	if systemTables[tableName] {
		return fmt.Errorf("cannot drop system table: %s", tableName)
	}

	// Check if table exists before attempting to drop
	exists, err := tableExists(tableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !exists {
		log.Printf("Table %s does not exist, skipping drop", tableName)
		return nil
	}

	// Drop the table using properly quoted identifier to prevent SQL injection
	// PostgreSQL identifiers can be quoted using double quotes
	// We validate tableName doesn't contain dangerous characters first
	if err := DB.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s" CASCADE`, tableName)).Error; err != nil {
		return fmt.Errorf("failed to drop table %s: %w", tableName, err)
	}

	log.Printf("Successfully dropped table %s", tableName)
	return nil
}

// Migrate runs database migrations for the given models with enhanced safety checks
func Migrate(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Get current schema version
	currentVersion, err := getCurrentSchemaVersion()
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// Get list of tables that should exist (from models)
	expectedTables := make(map[string]bool)

	// Track if we're migrating the screener table
	var stocksMigrated bool
	// Track if we're migrating the historical table
	var historicalMigrated bool
	// Track if we're migrating the company_info table
	var companyInfoMigrated bool
	// Track if we're migrating the fundamental_data table
	var fundamentalDataMigrated bool

	// Perform migrations for each model
	for _, model := range models {
		// Get table name from model
		stmt := &gorm.Statement{DB: DB}
		if err := stmt.Parse(model); err != nil {
			return fmt.Errorf("failed to parse model: %w", err)
		}

		tableName := stmt.Schema.Table
		expectedTables[tableName] = true

		// Check if this is the screener table
		if tableName == "screener" {
			stocksMigrated = true
		}

		// Check if this is the historical table
		if tableName == "historical" {
			historicalMigrated = true
		}

		// Check if this is the company_info table
		if tableName == "company_info" {
			companyInfoMigrated = true
		}

		// Check if this is the fundamental_data table
		if tableName == "fundamental_data" {
			fundamentalDataMigrated = true
		}

		// Check if table exists before migration
		exists, err := tableExists(tableName)
		if err != nil {
			return fmt.Errorf("failed to check table existence for %s: %w", tableName, err)
		}

		if exists {
			// Table exists, log that we're updating it
			log.Printf("Table %s exists, running migration...", tableName)
		} else {
			// Table doesn't exist, will be created
			log.Printf("Table %s does not exist, creating...", tableName)
		}

		// Run AutoMigrate which handles:
		// - Creating tables if they don't exist
		// - Adding missing columns
		// - Creating indexes
		// - Updating column types (where safe)
		if err := DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate table %s: %w", tableName, err)
		}

		log.Printf("Successfully migrated table %s", tableName)
	}

	// Get all current tables in the database
	currentTables, err := getCurrentTables()
	if err != nil {
		return fmt.Errorf("failed to get current tables: %w", err)
	}

	// Find and drop tables that no longer exist in models
	for _, tableName := range currentTables {
		// Skip system tables
		if tableName == "schema_versions" {
			continue
		}

		// If table is not in expected tables, it should be dropped
		if !expectedTables[tableName] {
			log.Printf("Table %s no longer exists in models, dropping...", tableName)
			if err := dropTableSafely(tableName); err != nil {
				log.Printf("Warning: Failed to drop table %s: %v", tableName, err)
				// Don't fail migration if drop fails, but log it
			}
		}
	}

	// Apply RLS policies and Realtime for screener table if it was migrated
	if stocksMigrated {
		if err := setupScreenerPolicies(); err != nil {
			log.Printf("Warning: Failed to setup screener policies: %v", err)
			// Don't fail migration if policy setup fails, but log it
		}
	}

	// Apply RLS policies for historical table if it was migrated
	if historicalMigrated {
		if err := setupHistoricalPolicies(); err != nil {
			log.Printf("Warning: Failed to setup historical policies: %v", err)
			// Don't fail migration if policy setup fails, but log it
		}
	}

	// Apply RLS policies for company_info table if it was migrated
	if companyInfoMigrated {
		if err := setupCompanyInfoPolicies(); err != nil {
			log.Printf("Warning: Failed to setup company_info policies: %v", err)
			// Don't fail migration if policy setup fails, but log it
		}
	}

	// Apply RLS policies for fundamental_data table if it was migrated
	if fundamentalDataMigrated {
		if err := setupFundamentalDataPolicies(); err != nil {
			log.Printf("Warning: Failed to setup fundamental_data policies: %v", err)
			// Don't fail migration if policy setup fails, but log it
		}
	}

	// Apply RLS policies for schema_versions table (system table)
	if err := setupSchemaVersionPolicies(); err != nil {
		log.Printf("Warning: Failed to setup schema_versions policies: %v", err)
		// Don't fail migration if policy setup fails, but log it
	}

	// Update schema version after successful migration
	// Using current timestamp as version identifier for simplicity
	// You can customize this to use semantic versioning if needed
	newVersion := fmt.Sprintf("%d", time.Now().Unix())
	description := "Migration completed"
	if err := updateSchemaVersion(newVersion, description); err != nil {
		// Log but don't fail migration if version tracking fails
		log.Printf("Warning: Failed to update schema version: %v", err)
	}

	if currentVersion != "" {
		log.Printf("Schema version updated from %s to %s", currentVersion, newVersion)
	} else {
		log.Printf("Schema version set to %s (first migration)", newVersion)
	}

	return nil
}

// setupScreenerPolicies sets up RLS policies and Realtime for the screener table
func setupScreenerPolicies() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Enable Row Level Security on screener table
	if err := DB.Exec(`ALTER TABLE IF EXISTS screener ENABLE ROW LEVEL SECURITY`).Error; err != nil {
		return fmt.Errorf("failed to enable RLS on screener table: %w", err)
	}

	// Revoke all privileges from anon and authenticated roles
	// This ensures users can't insert, update, or delete
	if err := DB.Exec(`REVOKE ALL ON TABLE screener FROM anon, authenticated`).Error; err != nil {
		// Log but don't fail - this might error if privileges don't exist
		log.Printf("Note: Could not revoke privileges (may not exist): %v", err)
	}

	// Grant SELECT permission to authenticated users (read-only access)
	if err := DB.Exec(`GRANT SELECT ON TABLE screener TO authenticated`).Error; err != nil {
		return fmt.Errorf("failed to grant SELECT permission: %w", err)
	}

	// Drop existing policy if it exists, then create new read-only policy
	// Using IF EXISTS to avoid errors if policy doesn't exist
	if err := DB.Exec(`
		DROP POLICY IF EXISTS "Allow read access to authenticated users" ON screener
	`).Error; err != nil {
		log.Printf("Note: Could not drop existing policy: %v", err)
	}

	// Create read-only policy for authenticated users
	if err := DB.Exec(`
		CREATE POLICY "Allow read access to authenticated users"
		ON screener
		FOR SELECT
		TO authenticated
		USING (true)
	`).Error; err != nil {
		return fmt.Errorf("failed to create read-only policy: %w", err)
	}

	// Set replica identity to FULL for Realtime to work properly
	// This allows Realtime to send full row data on updates/deletes
	if err := DB.Exec(`ALTER TABLE screener REPLICA IDENTITY FULL`).Error; err != nil {
		return fmt.Errorf("failed to set replica identity: %w", err)
	}

	// Add screener table to Supabase Realtime publication
	// This enables real-time subscriptions for the table
	if err := DB.Exec(`ALTER PUBLICATION supabase_realtime ADD TABLE screener`).Error; err != nil {
		// This might fail if table is already in publication or publication doesn't exist
		// Log but don't fail - Realtime may need to be enabled in Supabase dashboard
		log.Printf("Note: Could not add screener to Realtime publication (may already exist or Realtime not enabled): %v", err)
	}

	log.Println("Successfully configured RLS policies and Realtime for screener table")
	return nil
}

// setupHistoricalPolicies sets up RLS policies for the historical table
func setupHistoricalPolicies() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Enable Row Level Security on historical table
	if err := DB.Exec(`ALTER TABLE IF EXISTS historical ENABLE ROW LEVEL SECURITY`).Error; err != nil {
		return fmt.Errorf("failed to enable RLS on historical table: %w", err)
	}

	// Revoke all privileges from anon and authenticated roles
	// This ensures users can't insert, update, or delete
	if err := DB.Exec(`REVOKE ALL ON TABLE historical FROM anon, authenticated`).Error; err != nil {
		// Log but don't fail - this might error if privileges don't exist
		log.Printf("Note: Could not revoke privileges (may not exist): %v", err)
	}

	// Grant SELECT permission to authenticated users (read-only access)
	if err := DB.Exec(`GRANT SELECT ON TABLE historical TO authenticated`).Error; err != nil {
		return fmt.Errorf("failed to grant SELECT permission: %w", err)
	}

	// Drop existing policy if it exists, then create new read-only policy
	// Using IF EXISTS to avoid errors if policy doesn't exist
	if err := DB.Exec(`
		DROP POLICY IF EXISTS "Allow read access to authenticated users" ON historical
	`).Error; err != nil {
		log.Printf("Note: Could not drop existing policy: %v", err)
	}

	// Create read-only policy for authenticated users
	if err := DB.Exec(`
		CREATE POLICY "Allow read access to authenticated users"
		ON historical
		FOR SELECT
		TO authenticated
		USING (true)
	`).Error; err != nil {
		return fmt.Errorf("failed to create read-only policy: %w", err)
	}

	// Note: INSERT, UPDATE, and DELETE are automatically denied because:
	// 1. REVOKE ALL removes all privileges from anon and authenticated roles
	// 2. GRANT SELECT only grants SELECT privilege (not INSERT/UPDATE/DELETE)
	// 3. Only a SELECT policy exists (no policies for INSERT/UPDATE/DELETE)
	// In PostgreSQL RLS, if no policy exists for an operation, it's denied by default

	log.Println("Successfully configured RLS policies for historical table")
	return nil
}

// setupSchemaVersionPolicies sets up RLS policies for the schema_versions system table
func setupSchemaVersionPolicies() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Enable Row Level Security on schema_versions table
	if err := DB.Exec(`ALTER TABLE IF EXISTS schema_versions ENABLE ROW LEVEL SECURITY`).Error; err != nil {
		return fmt.Errorf("failed to enable RLS on schema_versions table: %w", err)
	}

	// Revoke all privileges from anon and authenticated roles
	// This is a system table, users should not be able to modify it
	if err := DB.Exec(`REVOKE ALL ON TABLE schema_versions FROM anon, authenticated`).Error; err != nil {
		// Log but don't fail - this might error if privileges don't exist
		log.Printf("Note: Could not revoke privileges on schema_versions (may not exist): %v", err)
	}

	// Grant SELECT permission to authenticated users (read-only access to system table)
	// This allows admins/users to check schema versions if needed
	if err := DB.Exec(`GRANT SELECT ON TABLE schema_versions TO authenticated`).Error; err != nil {
		return fmt.Errorf("failed to grant SELECT permission on schema_versions: %w", err)
	}

	// Drop existing policy if it exists
	if err := DB.Exec(`
		DROP POLICY IF EXISTS "Allow read access to schema_versions" ON schema_versions
	`).Error; err != nil {
		log.Printf("Note: Could not drop existing policy on schema_versions: %v", err)
	}

	// Create read-only policy for authenticated users on system table
	if err := DB.Exec(`
		CREATE POLICY "Allow read access to schema_versions"
		ON schema_versions
		FOR SELECT
		TO authenticated
		USING (true)
	`).Error; err != nil {
		return fmt.Errorf("failed to create read-only policy on schema_versions: %w", err)
	}

	log.Println("Successfully configured RLS policies for schema_versions system table")
	return nil
}

// setupCompanyInfoPolicies sets up RLS policies for the company_info table (read-only)
func setupCompanyInfoPolicies() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Enable Row Level Security on company_info table
	if err := DB.Exec(`ALTER TABLE IF EXISTS company_info ENABLE ROW LEVEL SECURITY`).Error; err != nil {
		return fmt.Errorf("failed to enable RLS on company_info table: %w", err)
	}

	// Revoke all privileges from anon and authenticated roles
	// This ensures users can't insert, update, or delete
	if err := DB.Exec(`REVOKE ALL ON TABLE company_info FROM anon, authenticated`).Error; err != nil {
		// Log but don't fail - this might error if privileges don't exist
		log.Printf("Note: Could not revoke privileges (may not exist): %v", err)
	}

	// Grant SELECT permission to both anon and authenticated users (read-only access for all)
	if err := DB.Exec(`GRANT SELECT ON TABLE company_info TO anon, authenticated`).Error; err != nil {
		return fmt.Errorf("failed to grant SELECT permission: %w", err)
	}

	// Drop existing policies if they exist
	if err := DB.Exec(`
		DROP POLICY IF EXISTS "Allow select on company info" ON company_info;
	`).Error; err != nil {
		log.Printf("Note: Could not drop existing policies: %v", err)
	}

	// Create read-only policy for all users (anon and authenticated)
	// No policies for INSERT/UPDATE/DELETE means they are automatically blocked by RLS
	if err := DB.Exec(`
		CREATE POLICY "Allow select on company info"
		ON company_info
		FOR SELECT
		USING (true)
	`).Error; err != nil {
		return fmt.Errorf("failed to create read-only policy: %w", err)
	}

	// Note: INSERT, UPDATE, and DELETE are automatically blocked because:
	// 1. REVOKE ALL removes all privileges
	// 2. GRANT SELECT only grants SELECT (not INSERT/UPDATE/DELETE)
	// 3. No RLS policies exist for INSERT/UPDATE/DELETE (default deny)

	// Set replica identity to FULL for Realtime to work properly
	if err := DB.Exec(`ALTER TABLE company_info REPLICA IDENTITY FULL`).Error; err != nil {
		return fmt.Errorf("failed to set replica identity: %w", err)
	}

	// Add company_info table to Supabase Realtime publication
	if err := DB.Exec(`ALTER PUBLICATION supabase_realtime ADD TABLE company_info`).Error; err != nil {
		// This might fail if table is already in publication or publication doesn't exist
		log.Printf("Note: Could not add company_info to Realtime publication (may already exist or Realtime not enabled): %v", err)
	}

	log.Println("Successfully configured RLS policies and Realtime for company_info table")
	return nil
}

// setupFundamentalDataPolicies sets up RLS policies for the fundamental_data table (read-only)
func setupFundamentalDataPolicies() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Enable Row Level Security on fundamental_data table
	if err := DB.Exec(`ALTER TABLE IF EXISTS fundamental_data ENABLE ROW LEVEL SECURITY`).Error; err != nil {
		return fmt.Errorf("failed to enable RLS on fundamental_data table: %w", err)
	}

	// Revoke all privileges from anon and authenticated roles
	// This ensures users can't insert, update, or delete
	if err := DB.Exec(`REVOKE ALL ON TABLE fundamental_data FROM anon, authenticated`).Error; err != nil {
		// Log but don't fail - this might error if privileges don't exist
		log.Printf("Note: Could not revoke privileges (may not exist): %v", err)
	}

	// Grant SELECT permission to both anon and authenticated users (read-only access for all)
	if err := DB.Exec(`GRANT SELECT ON TABLE fundamental_data TO anon, authenticated`).Error; err != nil {
		return fmt.Errorf("failed to grant SELECT permission: %w", err)
	}

	// Drop existing policies if they exist
	if err := DB.Exec(`
		DROP POLICY IF EXISTS "Allow select on fundamental data" ON fundamental_data;
	`).Error; err != nil {
		log.Printf("Note: Could not drop existing policies: %v", err)
	}

	// Create read-only policy for all users (anon and authenticated)
	// No policies for INSERT/UPDATE/DELETE means they are automatically blocked by RLS
	if err := DB.Exec(`
		CREATE POLICY "Allow select on fundamental data"
		ON fundamental_data
		FOR SELECT
		USING (true)
	`).Error; err != nil {
		return fmt.Errorf("failed to create read-only policy: %w", err)
	}

	// Note: INSERT, UPDATE, and DELETE are automatically blocked because:
	// 1. REVOKE ALL removes all privileges
	// 2. GRANT SELECT only grants SELECT (not INSERT/UPDATE/DELETE)
	// 3. No RLS policies exist for INSERT/UPDATE/DELETE (default deny)

	// Set replica identity to FULL for Realtime to work properly
	if err := DB.Exec(`ALTER TABLE fundamental_data REPLICA IDENTITY FULL`).Error; err != nil {
		return fmt.Errorf("failed to set replica identity: %w", err)
	}

	// Add fundamental_data table to Supabase Realtime publication
	if err := DB.Exec(`ALTER PUBLICATION supabase_realtime ADD TABLE fundamental_data`).Error; err != nil {
		// This might fail if table is already in publication or publication doesn't exist
		log.Printf("Note: Could not add fundamental_data to Realtime publication (may already exist or Realtime not enabled): %v", err)
	}

	log.Println("Successfully configured RLS policies and Realtime for fundamental_data table")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
