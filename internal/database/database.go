// Package database
package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	*sql.DB
}

func NewDatabase() (*Database, error) {
	db, err := sql.Open("sqlite3", "nest_manager.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{DB: db}
	if err := database.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize database tables: %w", err)
	}

	// Initialize log storage tables
	if err := database.InitializeLogTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize log tables: %w", err)
	}

	return database, nil
}

func (db *Database) initTables() error {
	// Create services table
	createServicesTable := `
	CREATE TABLE IF NOT EXISTS services (
		name TEXT PRIMARY KEY,
		dir TEXT NOT NULL,
		extra_env TEXT,
		java_opts TEXT,
		status TEXT DEFAULT 'stopped',
		health_status TEXT DEFAULT 'unknown',
		health_url TEXT,
		port INTEGER,
		pid INTEGER DEFAULT 0,
		service_order INTEGER,
		last_started DATETIME,
		description TEXT,
		is_enabled BOOLEAN DEFAULT TRUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Add missing columns to existing services table if they don't exist
	alterTableQueries := []string{
		`ALTER TABLE services ADD COLUMN description TEXT;`,
		`ALTER TABLE services ADD COLUMN is_enabled BOOLEAN DEFAULT TRUE;`,
		`ALTER TABLE services ADD COLUMN created_at DATETIME DEFAULT CURRENT_TIMESTAMP;`,
		`ALTER TABLE services ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP;`,
	}

	// Create environment variables table
	createEnvVarsTable := `
	CREATE TABLE IF NOT EXISTS service_env_vars (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_name TEXT NOT NULL,
		var_name TEXT NOT NULL,
		var_value TEXT NOT NULL,
		description TEXT,
		is_required BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (service_name) REFERENCES services(name) ON DELETE CASCADE,
		UNIQUE(service_name, var_name)
	);`

	// Create global environment variables table
	createGlobalEnvTable := `
	CREATE TABLE IF NOT EXISTS global_env_vars (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		var_name TEXT UNIQUE NOT NULL,
		var_value TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Create configurations table
	createConfigsTable := `
	CREATE TABLE IF NOT EXISTS configurations (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		services_json TEXT NOT NULL,
		is_default BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Create global configuration table
	createGlobalConfigTable := `
	CREATE TABLE IF NOT EXISTS global_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		projects_dir TEXT NOT NULL,
		java_home_override TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	tables := []string{createServicesTable, createEnvVarsTable, createGlobalEnvTable, createConfigsTable, createGlobalConfigTable}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Apply schema migrations for existing tables
	for _, query := range alterTableQueries {
		// Ignore errors for columns that already exist
		db.Exec(query)
	}

	return nil
}

// GetGlobalEnvVars retrieves all global environment variables
func (db *Database) GetGlobalEnvVars() (map[string]string, error) {
	query := `SELECT var_name, var_value FROM global_env_vars ORDER BY var_name`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query global env vars: %w", err)
	}
	defer rows.Close()

	envVars := make(map[string]string)
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return nil, fmt.Errorf("failed to scan global env var: %w", err)
		}
		envVars[name] = value
	}

	return envVars, rows.Err()
}

// SetGlobalEnvVar sets a global environment variable
func (db *Database) SetGlobalEnvVar(name, value string) error {
	query := `
		INSERT INTO global_env_vars (var_name, var_value, updated_at) 
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(var_name) DO UPDATE SET 
			var_value = excluded.var_value,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := db.Exec(query, name, value)
	if err != nil {
		return fmt.Errorf("failed to set global env var %s: %w", name, err)
	}
	return nil
}

// DeleteGlobalEnvVar deletes a global environment variable
func (db *Database) DeleteGlobalEnvVar(name string) error {
	query := `DELETE FROM global_env_vars WHERE var_name = ?`
	_, err := db.Exec(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete global env var %s: %w", name, err)
	}
	return nil
}
