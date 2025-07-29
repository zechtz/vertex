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
		`ALTER TABLE services ADD COLUMN build_system TEXT DEFAULT 'auto';`,
		`ALTER TABLE service_profiles ADD COLUMN projects_dir TEXT DEFAULT '';`,
		`ALTER TABLE service_profiles ADD COLUMN java_home_override TEXT DEFAULT '';`,
		`ALTER TABLE service_profiles ADD COLUMN is_active BOOLEAN DEFAULT FALSE;`,
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

	// Create users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT DEFAULT 'user',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_login DATETIME,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Create user profiles table
	createUserProfilesTable := `
	CREATE TABLE IF NOT EXISTS user_profiles (
		user_id TEXT PRIMARY KEY,
		display_name TEXT,
		avatar TEXT,
		preferences_json TEXT DEFAULT '{}',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	// Create service profiles table
	createServiceProfilesTable := `
	CREATE TABLE IF NOT EXISTS service_profiles (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		services_json TEXT NOT NULL DEFAULT '[]',
		env_vars_json TEXT DEFAULT '{}',
		projects_dir TEXT DEFAULT '',
		java_home_override TEXT DEFAULT '',
		is_default BOOLEAN DEFAULT FALSE,
		is_active BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(user_id, name)
	);`

	// Create profile-scoped global environment variables table
	createProfileEnvVarsTable := `
	CREATE TABLE IF NOT EXISTS profile_env_vars (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_id TEXT NOT NULL,
		var_name TEXT NOT NULL,
		var_value TEXT NOT NULL,
		description TEXT,
		is_required BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (profile_id) REFERENCES service_profiles(id) ON DELETE CASCADE,
		UNIQUE(profile_id, var_name)
	);`

	// Create profile-scoped service configuration overrides table
	createProfileServiceConfigsTable := `
	CREATE TABLE IF NOT EXISTS profile_service_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_id TEXT NOT NULL,
		service_name TEXT NOT NULL,
		config_key TEXT NOT NULL,
		config_value TEXT NOT NULL,
		config_type TEXT DEFAULT 'string',
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (profile_id) REFERENCES service_profiles(id) ON DELETE CASCADE,
		UNIQUE(profile_id, service_name, config_key)
	);`

	// Create profile-scoped dependencies table
	createProfileDependenciesTable := `
	CREATE TABLE IF NOT EXISTS profile_dependencies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_id TEXT NOT NULL,
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
		FOREIGN KEY (profile_id) REFERENCES service_profiles(id) ON DELETE CASCADE,
		UNIQUE(profile_id, service_name, dependency_service_name)
	);`

	tables := []string{createServicesTable, createEnvVarsTable, createGlobalEnvTable, createConfigsTable, createGlobalConfigTable, createSyncMetadataTable, createServiceDependenciesTable, createUsersTable, createUserProfilesTable, createServiceProfilesTable, createProfileEnvVarsTable, createProfileServiceConfigsTable, createProfileDependenciesTable}

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

// Profile-scoped environment variable methods

// GetProfileEnvVars retrieves all environment variables for a specific profile
func (db *Database) GetProfileEnvVars(profileID string) (map[string]string, error) {
	query := `SELECT var_name, var_value FROM profile_env_vars WHERE profile_id = ? ORDER BY var_name`
	rows, err := db.Query(query, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query profile env vars: %w", err)
	}
	defer rows.Close()

	envVars := make(map[string]string)
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return nil, fmt.Errorf("failed to scan profile env var: %w", err)
		}
		envVars[name] = value
	}

	return envVars, rows.Err()
}

// SetProfileEnvVar sets an environment variable for a specific profile
func (db *Database) SetProfileEnvVar(profileID, name, value, description string, isRequired bool) error {
	query := `INSERT OR REPLACE INTO profile_env_vars (profile_id, var_name, var_value, description, is_required, updated_at) 
			  VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`
	_, err := db.Exec(query, profileID, name, value, description, isRequired)
	if err != nil {
		return fmt.Errorf("failed to set profile env var %s: %w", name, err)
	}
	return nil
}

// DeleteProfileEnvVar deletes an environment variable for a specific profile
func (db *Database) DeleteProfileEnvVar(profileID, name string) error {
	query := `DELETE FROM profile_env_vars WHERE profile_id = ? AND var_name = ?`
	_, err := db.Exec(query, profileID, name)
	if err != nil {
		return fmt.Errorf("failed to delete profile env var %s: %w", name, err)
	}
	return nil
}

// Profile-scoped service configuration methods

// GetProfileServiceConfig retrieves service configuration overrides for a specific profile
func (db *Database) GetProfileServiceConfig(profileID, serviceName string) (map[string]string, error) {
	query := `SELECT config_key, config_value FROM profile_service_configs 
			  WHERE profile_id = ? AND service_name = ? ORDER BY config_key`
	rows, err := db.Query(query, profileID, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to query profile service config: %w", err)
	}
	defer rows.Close()

	config := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan profile service config: %w", err)
		}
		config[key] = value
	}

	return config, rows.Err()
}

// SetProfileServiceConfig sets a service configuration override for a specific profile
func (db *Database) SetProfileServiceConfig(profileID, serviceName, key, value, configType, description string) error {
	query := `INSERT OR REPLACE INTO profile_service_configs 
			  (profile_id, service_name, config_key, config_value, config_type, description, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`
	_, err := db.Exec(query, profileID, serviceName, key, value, configType, description)
	if err != nil {
		return fmt.Errorf("failed to set profile service config %s.%s: %w", serviceName, key, err)
	}
	return nil
}

// DeleteProfileServiceConfig deletes a service configuration override for a specific profile
func (db *Database) DeleteProfileServiceConfig(profileID, serviceName, key string) error {
	query := `DELETE FROM profile_service_configs WHERE profile_id = ? AND service_name = ? AND config_key = ?`
	_, err := db.Exec(query, profileID, serviceName, key)
	if err != nil {
		return fmt.Errorf("failed to delete profile service config %s.%s: %w", serviceName, key, err)
	}
	return nil
}

// Active profile management

// GetActiveProfile retrieves the active profile for a user
func (db *Database) GetActiveProfile(userID string) (string, error) {
	query := `SELECT id FROM service_profiles WHERE user_id = ? AND is_active = TRUE LIMIT 1`
	var profileID string
	err := db.QueryRow(query, userID).Scan(&profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // No active profile
		}
		return "", fmt.Errorf("failed to get active profile: %w", err)
	}
	return profileID, nil
}

// SetActiveProfile sets the active profile for a user
func (db *Database) SetActiveProfile(userID, profileID string) error {
	// First, clear any existing active profile
	clearQuery := `UPDATE service_profiles SET is_active = FALSE WHERE user_id = ?`
	if _, err := db.Exec(clearQuery, userID); err != nil {
		return fmt.Errorf("failed to clear active profiles: %w", err)
	}

	// Set the new active profile
	setQuery := `UPDATE service_profiles SET is_active = TRUE WHERE id = ? AND user_id = ?`
	result, err := db.Exec(setQuery, profileID, userID)
	if err != nil {
		return fmt.Errorf("failed to set active profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("profile not found or access denied")
	}

	return nil
}

// DeleteService removes a service from the database
func (db *Database) DeleteService(serviceName string) error {
	// Delete the service - this will cascade delete all related records
	// due to foreign key constraints (env vars, dependencies, etc.)
	query := `DELETE FROM services WHERE name = ?`
	result, err := db.Exec(query, serviceName)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	return nil
}
