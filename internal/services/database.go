package services

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zechtz/vertex/internal/models"
)

func (sm *Manager) loadServices(config models.Config) error {
	// First, reset all services to stopped status on application startup
	_, err := sm.db.Exec(`
		UPDATE services 
		SET status = 'stopped', health_status = 'unknown', pid = 0, updated_at = CURRENT_TIMESTAMP`)
	if err != nil {
		log.Printf("Warning: Failed to reset service statuses on startup: %v", err)
	} else {
		log.Printf("[INFO] Reset all service statuses to 'stopped' on application startup")
	}

	// Check if initial service configuration sync has been completed
	const syncType = "initial_service_config"
	initialSyncCompleted, err := sm.db.IsSyncCompleted(syncType)
	if err != nil {
		log.Printf("[WARN] Failed to check initial sync status: %v", err)
		// Continue with safe default behavior
		initialSyncCompleted = true
	}

	for i := range config.Services {
		service := &config.Services[i]

		// Ensure service has a UUID
		if service.ID == "" {
			service.ID = uuid.New().String()
		}

		// Try to load existing service from database
		var dbService models.Service
		row := sm.db.QueryRow(`
			SELECT id, name, dir, extra_env, java_opts, status, health_status, health_url, port, pid, service_order, last_started, description, is_enabled, build_system, verbose_logging
			FROM services WHERE id = ?`, service.ID)

		var description sql.NullString
		var isEnabled sql.NullBool
		var buildSystem sql.NullString
		var verboseLogging sql.NullBool
		err := row.Scan(&dbService.ID, &dbService.Name, &dbService.Dir, &dbService.ExtraEnv, &dbService.JavaOpts,
			&dbService.Status, &dbService.HealthStatus, &dbService.HealthURL, &dbService.Port,
			&dbService.PID, &dbService.Order, &dbService.LastStarted, &description, &isEnabled, &buildSystem, &verboseLogging)

		if err == sql.ErrNoRows {
			// Service doesn't exist in DB, insert it
			_, err = sm.db.Exec(`
				INSERT INTO services (id, name, dir, extra_env, java_opts, status, health_status, health_url, port, service_order, description, is_enabled, build_system, verbose_logging, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
				service.ID, service.Name, service.Dir, service.ExtraEnv, service.JavaOpts, service.Status,
				service.HealthStatus, service.HealthURL, service.Port, service.Order, "", true, "auto", false)
			if err != nil {
				return fmt.Errorf("failed to insert service UUID %s: %w", service.ID, err)
			}
			service.EnvVars = make(map[string]models.EnvVar)
			service.Logs = []models.LogEntry{}
			service.BuildSystem = "auto"
			service.VerboseLogging = false
			sm.services[service.ID] = service
		} else if err != nil {
			return fmt.Errorf("failed to query service UUID %s: %w", service.ID, err)
		} else {
			// Service exists, use DB data but update only directory path (which is config-based)
			dbService.Dir = service.Dir
			dbService.ExtraEnv = service.ExtraEnv

			// Only sync default values if this is the first time loading from config files
			if !initialSyncCompleted {
				// Initial sync: only use defaults if DB values are empty
				if dbService.JavaOpts == "" {
					dbService.JavaOpts = service.JavaOpts
				}
				if dbService.HealthURL == "" {
					dbService.HealthURL = service.HealthURL
				}
				if dbService.Port == 0 {
					dbService.Port = service.Port
				}
			}
			// After initial sync, completely preserve database values for user-configurable fields
			dbService.Logs = []models.LogEntry{}

			if description.Valid {
				dbService.Description = description.String
			}
			if isEnabled.Valid {
				dbService.IsEnabled = isEnabled.Bool
			} else {
				dbService.IsEnabled = true
			}
			if buildSystem.Valid {
				dbService.BuildSystem = buildSystem.String
			} else {
				dbService.BuildSystem = "auto"
			}
			if verboseLogging.Valid {
				dbService.VerboseLogging = verboseLogging.Bool
			} else {
				dbService.VerboseLogging = false
			}

			// Load environment variables for this service
			dbService.EnvVars = make(map[string]models.EnvVar)
			envRows, err := sm.db.Query(`
				SELECT var_name, var_value, description, is_required 
				FROM service_env_vars 
				WHERE service_id = ?`, dbService.ID)
			if err == nil {
				defer envRows.Close()
				for envRows.Next() {
					var envVar models.EnvVar
					var envDesc sql.NullString
					err := envRows.Scan(&envVar.Name, &envVar.Value, &envDesc, &envVar.IsRequired)
					if err == nil {
						if envDesc.Valid {
							envVar.Description = envDesc.String
						}
						dbService.EnvVars[envVar.Name] = envVar
					}
				}
			}

			sm.services[dbService.ID] = &dbService
		}
	}

	// Load additional services that exist in database but not in config file
	if err := sm.loadDynamicServices(); err != nil {
		log.Printf("[WARN] Failed to load dynamic services: %v", err)
	}

	// Mark initial service configuration sync as completed (only if it wasn't already)
	if !initialSyncCompleted {
		if err := sm.db.MarkSyncCompleted(syncType); err != nil {
			log.Printf("[WARN] Failed to mark initial service config sync as completed: %v", err)
		} else {
			log.Printf("[INFO] Initial service configuration sync completed - future restarts will preserve user configurations")
		}
	}

	return nil
}

func (sm *Manager) loadConfigurations() error {
	rows, err := sm.db.Query("SELECT id, name, services_json, is_default FROM configurations")
	if err != nil {
		return fmt.Errorf("failed to query configurations: %w", err)
	}
	defer rows.Close()

	hasConfigs := false
	for rows.Next() {
		hasConfigs = true
		var config models.Configuration
		var servicesJSON string

		err := rows.Scan(&config.ID, &config.Name, &servicesJSON, &config.IsDefault)
		if err != nil {
			return fmt.Errorf("failed to scan configuration: %w", err)
		}

		err = json.Unmarshal([]byte(servicesJSON), &config.Services)
		if err != nil {
			return fmt.Errorf("failed to unmarshal services JSON: %w", err)
		}

		sm.configurations[config.ID] = &config

		if config.IsDefault {
			sm.activeConfigID = config.ID
		}
	}

	// Create default configuration if none exist
	if !hasConfigs {
		defaultConfig := &models.Configuration{
			ID:        "default",
			Name:      "Default Configuration",
			IsDefault: true,
			Services:  make([]models.ConfigService, 0, len(sm.services)),
		}

		for _, service := range sm.services {
			defaultConfig.Services = append(defaultConfig.Services, models.ConfigService{
				ID:    service.ID,
				Order: service.Order,
			})
		}

		sm.configurations["default"] = defaultConfig
		if err := sm.saveConfigurationToDB(defaultConfig); err != nil {
			log.Printf("[WARN] Failed to save default configuration: %v", err)
		}
	}

	return nil
}

func (sm *Manager) saveConfigurationToDB(config *models.Configuration) error {
	servicesJSON, err := json.Marshal(config.Services)
	if err != nil {
		return fmt.Errorf("failed to marshal services: %w", err)
	}

	_, err = sm.db.Exec(`
		INSERT OR REPLACE INTO configurations (id, name, services_json, is_default, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		config.ID, config.Name, string(servicesJSON), config.IsDefault)

	return err
}

func (sm *Manager) updateConfigurationInDB(config *models.Configuration) error {
	servicesJSON, err := json.Marshal(config.Services)
	if err != nil {
		return fmt.Errorf("failed to marshal services: %w", err)
	}

	_, err = sm.db.Exec(`
		UPDATE configurations 
		SET name = ?, services_json = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		config.Name, string(servicesJSON), config.ID)

	return err
}

func (sm *Manager) updateConfigurationDefaults(defaultConfigID string) error {
	tx, err := sm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Set all configurations to not default
	_, err = tx.Exec("UPDATE configurations SET is_default = 0")
	if err != nil {
		return fmt.Errorf("failed to clear default flags: %w", err)
	}

	// Set the specified configuration as default
	_, err = tx.Exec("UPDATE configurations SET is_default = 1 WHERE id = ?", defaultConfigID)
	if err != nil {
		return fmt.Errorf("failed to set default configuration: %w", err)
	}

	return tx.Commit()
}

func (sm *Manager) saveGlobalConfigToDB(projectsDir, javaHomeOverride string) error {
	// First, clear existing configuration
	_, err := sm.db.Exec("DELETE FROM global_config")
	if err != nil {
		return fmt.Errorf("failed to clear existing global config: %w", err)
	}

	// Insert new configuration
	_, err = sm.db.Exec(`
		INSERT INTO global_config (projects_dir, java_home_override, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)`,
		projectsDir, javaHomeOverride)
	if err != nil {
		return fmt.Errorf("failed to save global config: %w", err)
	}

	return nil
}

func (sm *Manager) loadGlobalConfigFromDB() error {
	var projectsDir, javaHomeOverride string
	err := sm.db.QueryRow("SELECT projects_dir, java_home_override FROM global_config ORDER BY id DESC LIMIT 1").
		Scan(&projectsDir, &javaHomeOverride)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// No global config in database, use defaults
			return nil
		}
		return fmt.Errorf("failed to load global config from database: %w", err)
	}

	// Update the configuration
	sm.config.ProjectsDir = projectsDir
	sm.config.JavaHomeOverride = javaHomeOverride

	return nil
}

func (sm *Manager) loadDynamicServices() error {
	// Query all services from database
	rows, err := sm.db.Query(`
		SELECT id, name, dir, extra_env, java_opts, status, health_status, health_url, port, pid, service_order, last_started, description, is_enabled, build_system, verbose_logging
		FROM services`)
	if err != nil {
		return fmt.Errorf("failed to query dynamic services: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dbService models.Service
		var description sql.NullString
		var isEnabled sql.NullBool
		var buildSystem sql.NullString
		var verboseLogging sql.NullBool

		err := rows.Scan(&dbService.ID, &dbService.Name, &dbService.Dir, &dbService.ExtraEnv, &dbService.JavaOpts,
			&dbService.Status, &dbService.HealthStatus, &dbService.HealthURL, &dbService.Port,
			&dbService.PID, &dbService.Order, &dbService.LastStarted, &description, &isEnabled, &buildSystem, &verboseLogging)
		if err != nil {
			log.Printf("[WARN] Failed to scan dynamic service: %v", err)
			continue
		}

		// Skip if service is already loaded from config
		if _, exists := sm.services[dbService.ID]; exists {
			continue
		}

		// Handle nullable fields
		if description.Valid {
			dbService.Description = description.String
		}
		if isEnabled.Valid {
			dbService.IsEnabled = isEnabled.Bool
		} else {
			dbService.IsEnabled = true
		}
		if buildSystem.Valid {
			dbService.BuildSystem = buildSystem.String
		} else {
			dbService.BuildSystem = "auto"
		}
		if verboseLogging.Valid {
			dbService.VerboseLogging = verboseLogging.Bool
		} else {
			dbService.VerboseLogging = false
		}

		// Initialize required fields
		dbService.EnvVars = make(map[string]models.EnvVar)
		dbService.Logs = []models.LogEntry{}

		// Load environment variables for this service
		envRows, err := sm.db.Query("SELECT var_name, var_value, description, is_required FROM service_env_vars WHERE service_id = ?", dbService.ID)
		if err == nil {
			for envRows.Next() {
				var envVar models.EnvVar
				var envDesc sql.NullString
				err := envRows.Scan(&envVar.Name, &envVar.Value, &envDesc, &envVar.IsRequired)
				if err == nil {
					if envDesc.Valid {
						envVar.Description = envDesc.String
					}
					dbService.EnvVars[envVar.Name] = envVar
				}
			}
			envRows.Close()
		}

		// Add to services map
		sm.services[dbService.ID] = &dbService
		log.Printf("[INFO] Loaded dynamic service from database: UUID %s (Name: %s)", dbService.ID, dbService.Name)
	}

	return rows.Err()
}

func (sm *Manager) insertServiceInDB(service *models.Service) error {
	_, err := sm.db.Exec(`
		INSERT INTO services (id, name, dir, extra_env, java_opts, status, health_status, health_url, port, service_order, description, is_enabled, build_system, verbose_logging, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		service.ID, service.Name, service.Dir, service.ExtraEnv, service.JavaOpts, service.Status,
		service.HealthStatus, service.HealthURL, service.Port, service.Order,
		service.Description, service.IsEnabled, service.BuildSystem, service.VerboseLogging)

	return err
}

func (sm *Manager) upsertServiceInDB(service *models.Service) error {
	// Try to update first
	result, err := sm.db.Exec(`
		UPDATE services 
		SET name = ?, dir = ?, extra_env = ?, java_opts = ?, status = ?, health_status = ?, health_url = ?, 
		    port = ?, service_order = ?, description = ?, is_enabled = ?, build_system = ?, 
		    pid = ?, last_started = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		service.Name, service.Dir, service.ExtraEnv, service.JavaOpts, service.Status, service.HealthStatus,
		service.HealthURL, service.Port, service.Order, service.Description, service.IsEnabled,
		service.BuildSystem, service.PID, service.LastStarted, service.ID)
	if err != nil {
		return err
	}

	// Check if any row was affected by the update
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were affected, the service doesn't exist, so insert it
	if rowsAffected == 0 {
		return sm.insertServiceInDB(service)
	}

	return nil
}

func (sm *Manager) UpdateServiceInDB(service *models.Service) error {
	_, err := sm.db.Exec(`
		UPDATE services 
		SET status = ?, health_status = ?, pid = ?, last_started = ?, service_order = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		service.Status, service.HealthStatus, service.PID, service.LastStarted, service.Order, service.ID)

	return err
}

func (sm *Manager) UpdateServiceConfigInDB(service *models.Service) error {
	_, err := sm.db.Exec(`
		UPDATE services
		SET name = ?, java_opts = ?, health_url = ?, port = ?, service_order = ?, description = ?,
		    is_enabled = ?, build_system = ?, verbose_logging = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		service.Name, service.JavaOpts, service.HealthURL, service.Port, service.Order,
		service.Description, service.IsEnabled, service.BuildSystem, service.VerboseLogging, service.ID)

	return err
}

func (sm *Manager) DeleteService(serviceUUID string) error {
	_, err := sm.db.Exec("DELETE FROM services WHERE id = ?", serviceUUID)
	if err != nil {
		return fmt.Errorf("failed to delete service UUID %s: %w", serviceUUID, err)
	}
	return nil
}

func (sm *Manager) GetServiceEnvVars(serviceUUID string) (map[string]models.EnvVar, error) {
	rows, err := sm.db.Query(`
		SELECT var_name, var_value, description, is_required 
		FROM service_env_vars 
		WHERE service_id = ?`, serviceUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query service env vars: %w", err)
	}
	defer rows.Close()

	envVars := make(map[string]models.EnvVar)
	for rows.Next() {
		var envVar models.EnvVar
		var description sql.NullString

		err := rows.Scan(&envVar.Name, &envVar.Value, &description, &envVar.IsRequired)
		if err != nil {
			return nil, fmt.Errorf("failed to scan env var: %w", err)
		}

		if description.Valid {
			envVar.Description = description.String
		}

		envVars[envVar.Name] = envVar
	}

	return envVars, nil
}

func (sm *Manager) UpdateServiceEnvVars(serviceUUID string, envVars map[string]models.EnvVar) error {
	// Start a transaction to ensure atomicity
	tx, err := sm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing environment variables for this service
	_, err = tx.Exec("DELETE FROM service_env_vars WHERE service_id = ?", serviceUUID)
	if err != nil {
		return fmt.Errorf("failed to clear existing service env vars: %w", err)
	}

	// Insert new environment variables
	for _, envVar := range envVars {
		if envVar.Name == "" {
			continue // Skip empty names
		}

		_, err = tx.Exec(`
			INSERT INTO service_env_vars (service_id, var_name, var_value, description, is_required, updated_at)
			VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
			serviceUUID, envVar.Name, envVar.Value, envVar.Description, envVar.IsRequired)
		if err != nil {
			return fmt.Errorf("failed to insert service env var %s: %w", envVar.Name, err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update the in-memory service with new environment variables
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if service, exists := sm.services[serviceUUID]; exists {
		service.Mutex.Lock()
		service.EnvVars = envVars
		service.Mutex.Unlock()
	}

	return nil
}

func (sm *Manager) loadEnvVarsFromFishFile() error {
	fishFile := "env_vars.fish"
	file, err := os.Open(fishFile)
	if err != nil {
		return fmt.Errorf("could not open %s: %w", fishFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	envVarRegex := regexp.MustCompile(`^set -gx (\w+) (.+)$`)

	// Clear existing global environment variables
	_, err = sm.db.Exec("DELETE FROM global_env_vars")
	if err != nil {
		log.Printf("Failed to clear existing global env vars: %v", err)
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		matches := envVarRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := matches[1]
			value := strings.Trim(matches[2], `"`)

			// Store in database
			_, err = sm.db.Exec(`
				INSERT OR REPLACE INTO global_env_vars (var_name, var_value, description, updated_at)
				VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
				key, value, "Common environment variable")
			if err != nil {
				log.Printf("Failed to store global env var %s: %v", key, err)
			}
		}
	}

	return scanner.Err()
}

func (sm *Manager) CleanupAndReloadGlobalEnvVars() error {
	log.Printf("Cleaning up global environment variables...")
	return sm.loadEnvVarsFromFishFile()
}

func (sm *Manager) ReloadEnvVarsFromFishFile() error {
	return sm.loadEnvVarsFromFishFile()
}

func (sm *Manager) GetGlobalEnvVars() (map[string]string, error) {
	rows, err := sm.db.Query("SELECT var_name, var_value FROM global_env_vars")
	if err != nil {
		return nil, fmt.Errorf("failed to query global env vars: %w", err)
	}
	defer rows.Close()

	envVars := make(map[string]string)
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return nil, fmt.Errorf("failed to scan env var: %w", err)
		}
		envVars[name] = value
	}

	return envVars, nil
}

func (sm *Manager) UpdateGlobalEnvVars(envVars map[string]string) error {
	// Start a transaction to ensure atomicity
	tx, err := sm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing global environment variables
	_, err = tx.Exec("DELETE FROM global_env_vars")
	if err != nil {
		return fmt.Errorf("failed to clear existing env vars: %w", err)
	}

	// Insert new environment variables
	for name, value := range envVars {
		if name == "" {
			continue // Skip empty names
		}

		_, err = tx.Exec(`
			INSERT INTO global_env_vars (var_name, var_value, description, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
			name, value, "Updated via web interface")
		if err != nil {
			return fmt.Errorf("failed to insert env var %s: %w", name, err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (sm *Manager) CleanupOldLogs(maxDays int, maxLogsPerService int) error {
	log.Printf("[INFO] Starting log cleanup - keeping logs from last %d days and max %d logs per service", maxDays, maxLogsPerService)

	// Start a transaction for consistency
	tx, err := sm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Count logs before cleanup
	var totalLogsBefore int
	err = tx.QueryRow("SELECT COUNT(*) FROM service_logs").Scan(&totalLogsBefore)
	if err != nil {
		return fmt.Errorf("failed to count logs before cleanup: %w", err)
	}

	// Delete logs older than maxDays
	cutoffDate := time.Now().AddDate(0, 0, -maxDays)
	result, err := tx.Exec(`
		DELETE FROM service_logs 
		WHERE created_at < ?`, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to delete old logs: %w", err)
	}

	deletedOld, _ := result.RowsAffected()

	// For each service, keep only the most recent maxLogsPerService logs
	services := sm.GetServices()
	var deletedPerService int64 = 0

	for _, service := range services {
		// Delete excess logs for this service, keeping only the most recent ones
		result, err := tx.Exec(`
			DELETE FROM service_logs 
			WHERE service_id = ? 
			AND id NOT IN (
				SELECT id FROM service_logs 
				WHERE service_id = ? 
				ORDER BY created_at DESC 
				LIMIT ?
			)`, service.ID, service.ID, maxLogsPerService)
		if err != nil {
			log.Printf("[WARN] Failed to cleanup logs for service UUID %s: %v", service.ID, err)
			continue
		}

		deleted, _ := result.RowsAffected()
		deletedPerService += deleted
	}

	// Count logs after cleanup
	var totalLogsAfter int
	err = tx.QueryRow("SELECT COUNT(*) FROM service_logs").Scan(&totalLogsAfter)
	if err != nil {
		return fmt.Errorf("failed to count logs after cleanup: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit log cleanup transaction: %w", err)
	}

	totalDeleted := deletedOld + deletedPerService
	log.Printf("[INFO] Log cleanup completed - deleted %d logs (%d old, %d excess per service). Logs: %d -> %d",
		totalDeleted, deletedOld, deletedPerService, totalLogsBefore, totalLogsAfter)

	return nil
}

func (sm *Manager) AutoCleanupLogs() error {
	// Default: keep logs from last 7 days and max 1000 logs per service
	return sm.CleanupOldLogs(7, 1000)
}
