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

	// Create sync metadata table to track one-time synchronizations
	createSyncMetadataTable := `
	CREATE TABLE IF NOT EXISTS sync_metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sync_type TEXT UNIQUE NOT NULL,
		is_completed BOOLEAN DEFAULT FALSE,
		completed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Create service dependencies table
	createServiceDependenciesTable := `
	CREATE TABLE IF NOT EXISTS service_dependencies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_name TEXT NOT NULL,
		dependency_service_name TEXT NOT NULL,
		dependency_type TEXT NOT NULL DEFAULT 'hard',
		health_check BOOLEAN DEFAULT TRUE,
		timeout_seconds INTEGER DEFAULT 120,
		retry_interval_seconds INTEGER DEFAULT 5,
		is_required BOOLEAN DEFAULT TRUE,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (service_name) REFERENCES services(name) ON DELETE CASCADE,
		UNIQUE(service_name, dependency_service_name)
	);`

	tables := []string{createServicesTable, createEnvVarsTable, createGlobalEnvTable, createConfigsTable, createGlobalConfigTable, createSyncMetadataTable, createServiceDependenciesTable}

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

// IsSyncCompleted checks if a specific sync type has been completed
func (db *Database) IsSyncCompleted(syncType string) (bool, error) {
	var isCompleted bool
	query := `SELECT is_completed FROM sync_metadata WHERE sync_type = ?`
	err := db.QueryRow(query, syncType).Scan(&isCompleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Not found means not completed
		}
		return false, fmt.Errorf("failed to check sync status for %s: %w", syncType, err)
	}
	return isCompleted, nil
}

// MarkSyncCompleted marks a specific sync type as completed
func (db *Database) MarkSyncCompleted(syncType string) error {
	query := `
		INSERT INTO sync_metadata (sync_type, is_completed, completed_at, updated_at) 
		VALUES (?, TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(sync_type) DO UPDATE SET 
			is_completed = TRUE,
			completed_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := db.Exec(query, syncType)
	if err != nil {
		return fmt.Errorf("failed to mark sync completed for %s: %w", syncType, err)
	}
	return nil
}

// SaveServiceDependencies saves service dependencies to the database
func (db *Database) SaveServiceDependencies(serviceName string, dependencies []interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing dependencies for this service
	_, err = tx.Exec("DELETE FROM service_dependencies WHERE service_name = ?", serviceName)
	if err != nil {
		return fmt.Errorf("failed to clear existing dependencies: %w", err)
	}

	// Insert new dependencies
	for _, dep := range dependencies {
		if depMap, ok := dep.(map[string]interface{}); ok {
			dependencyServiceName, _ := depMap["serviceName"].(string)
			dependencyType, _ := depMap["type"].(string)
			healthCheck, _ := depMap["healthCheck"].(bool)
			timeoutSeconds := 120 // default
			retryIntervalSeconds := 5 // default
			isRequired, _ := depMap["required"].(bool)
			description, _ := depMap["description"].(string)

			if timeoutSecondsFloat, ok := depMap["timeoutSeconds"].(float64); ok {
				timeoutSeconds = int(timeoutSecondsFloat)
			}
			if retrySecondsFloat, ok := depMap["retryIntervalSeconds"].(float64); ok {
				retryIntervalSeconds = int(retrySecondsFloat)
			}

			if dependencyType == "" {
				dependencyType = "hard"
			}

			_, err = tx.Exec(`
				INSERT INTO service_dependencies (
					service_name, dependency_service_name, dependency_type, 
					health_check, timeout_seconds, retry_interval_seconds, 
					is_required, description, updated_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
				serviceName, dependencyServiceName, dependencyType,
				healthCheck, timeoutSeconds, retryIntervalSeconds,
				isRequired, description)
			if err != nil {
				return fmt.Errorf("failed to insert dependency %s -> %s: %w", serviceName, dependencyServiceName, err)
			}
		}
	}

	return tx.Commit()
}

// LoadServiceDependencies loads service dependencies from the database
func (db *Database) LoadServiceDependencies(serviceName string) ([]map[string]interface{}, error) {
	rows, err := db.Query(`
		SELECT dependency_service_name, dependency_type, health_check, 
		       timeout_seconds, retry_interval_seconds, is_required, description
		FROM service_dependencies 
		WHERE service_name = ?
		ORDER BY dependency_service_name`, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer rows.Close()

	var dependencies []map[string]interface{}
	for rows.Next() {
		var dependencyServiceName, dependencyType, description string
		var healthCheck, isRequired bool
		var timeoutSeconds, retryIntervalSeconds int

		err := rows.Scan(&dependencyServiceName, &dependencyType, &healthCheck,
			&timeoutSeconds, &retryIntervalSeconds, &isRequired, &description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}

		dependencies = append(dependencies, map[string]interface{}{
			"serviceName":            dependencyServiceName,
			"type":                   dependencyType,
			"healthCheck":            healthCheck,
			"timeoutSeconds":         timeoutSeconds,
			"retryIntervalSeconds":   retryIntervalSeconds,
			"required":               isRequired,
			"description":            description,
		})
	}

	return dependencies, rows.Err()
}

// GetAllServiceDependencies returns all service dependencies
func (db *Database) GetAllServiceDependencies() (map[string][]map[string]interface{}, error) {
	rows, err := db.Query(`
		SELECT service_name, dependency_service_name, dependency_type, health_check, 
		       timeout_seconds, retry_interval_seconds, is_required, description
		FROM service_dependencies 
		ORDER BY service_name, dependency_service_name`)
	if err != nil {
		return nil, fmt.Errorf("failed to query all dependencies: %w", err)
	}
	defer rows.Close()

	allDependencies := make(map[string][]map[string]interface{})
	for rows.Next() {
		var serviceName, dependencyServiceName, dependencyType, description string
		var healthCheck, isRequired bool
		var timeoutSeconds, retryIntervalSeconds int

		err := rows.Scan(&serviceName, &dependencyServiceName, &dependencyType, &healthCheck,
			&timeoutSeconds, &retryIntervalSeconds, &isRequired, &description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}

		if allDependencies[serviceName] == nil {
			allDependencies[serviceName] = []map[string]interface{}{}
		}

		allDependencies[serviceName] = append(allDependencies[serviceName], map[string]interface{}{
			"serviceName":            dependencyServiceName,
			"type":                   dependencyType,
			"healthCheck":            healthCheck,
			"timeoutSeconds":         timeoutSeconds,
			"retryIntervalSeconds":   retryIntervalSeconds,
			"required":               isRequired,
			"description":            description,
		})
	}

	return allDependencies, rows.Err()
}
