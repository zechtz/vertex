// Package services
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

	"github.com/zechtz/nest-up/internal/models"
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

	for i := range config.Services {
		service := &config.Services[i]

		// Try to load existing service from database
		var dbService models.Service
		row := sm.db.QueryRow(`
			SELECT name, dir, extra_env, java_opts, status, health_status, health_url, port, pid, service_order, last_started, description, is_enabled
			FROM services WHERE name = ?`, service.Name)

		var description sql.NullString
		var isEnabled sql.NullBool
		err := row.Scan(&dbService.Name, &dbService.Dir, &dbService.ExtraEnv, &dbService.JavaOpts,
			&dbService.Status, &dbService.HealthStatus, &dbService.HealthURL, &dbService.Port,
			&dbService.PID, &dbService.Order, &dbService.LastStarted, &description, &isEnabled)

		if err == sql.ErrNoRows {
			// Service doesn't exist in DB, insert it
			_, err = sm.db.Exec(`
				INSERT INTO services (name, dir, extra_env, java_opts, status, health_status, health_url, port, service_order, description, is_enabled)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				service.Name, service.Dir, service.ExtraEnv, service.JavaOpts, service.Status,
				service.HealthStatus, service.HealthURL, service.Port, service.Order, "", true)
			if err != nil {
				return fmt.Errorf("failed to insert service %s: %w", service.Name, err)
			}
			service.EnvVars = make(map[string]models.EnvVar)
			service.Logs = []models.LogEntry{}
			sm.services[service.Name] = service
		} else if err != nil {
			return fmt.Errorf("failed to query service %s: %w", service.Name, err)
		} else {
			// Service exists, use DB data but update only directory path (which is config-based)
			dbService.Dir = service.Dir
			dbService.ExtraEnv = service.ExtraEnv
			// Keep database values for user-configurable fields
			if dbService.JavaOpts == "" {
				dbService.JavaOpts = service.JavaOpts // Use default only if DB is empty
			}
			if dbService.HealthURL == "" {
				dbService.HealthURL = service.HealthURL // Use default only if DB is empty
			}
			if dbService.Port == 0 {
				dbService.Port = service.Port // Use default only if DB is empty
			}
			dbService.Logs = []models.LogEntry{}

			if description.Valid {
				dbService.Description = description.String
			}
			if isEnabled.Valid {
				dbService.IsEnabled = isEnabled.Bool
			} else {
				dbService.IsEnabled = true
			}

			// Load environment variables for this service
			dbService.EnvVars = make(map[string]models.EnvVar)
			envRows, err := sm.db.Query(`
				SELECT var_name, var_value, description, is_required 
				FROM service_env_vars 
				WHERE service_name = ?`, dbService.Name)
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

			sm.services[service.Name] = &dbService
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
				Name:  service.Name,
				Order: service.Order,
			})
		}

		sm.configurations["default"] = defaultConfig
		sm.saveConfigurationToDB(defaultConfig)
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

func (sm *Manager) updateServiceInDB(service *models.Service) error {
	_, err := sm.db.Exec(`
		UPDATE services 
		SET status = ?, health_status = ?, pid = ?, last_started = ?, updated_at = CURRENT_TIMESTAMP
		WHERE name = ?`,
		service.Status, service.HealthStatus, service.PID, service.LastStarted, service.Name)

	return err
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

// Service-specific environment variable methods
func (sm *Manager) GetServiceEnvVars(serviceName string) (map[string]models.EnvVar, error) {
	rows, err := sm.db.Query(`
		SELECT var_name, var_value, description, is_required 
		FROM service_env_vars 
		WHERE service_name = ?`, serviceName)
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

func (sm *Manager) UpdateServiceEnvVars(serviceName string, envVars map[string]models.EnvVar) error {
	// Start a transaction to ensure atomicity
	tx, err := sm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing environment variables for this service
	_, err = tx.Exec("DELETE FROM service_env_vars WHERE service_name = ?", serviceName)
	if err != nil {
		return fmt.Errorf("failed to clear existing service env vars: %w", err)
	}

	// Insert new environment variables
	for _, envVar := range envVars {
		if envVar.Name == "" {
			continue // Skip empty names
		}

		_, err = tx.Exec(`
			INSERT INTO service_env_vars (service_name, var_name, var_value, description, is_required, updated_at)
			VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
			serviceName, envVar.Name, envVar.Value, envVar.Description, envVar.IsRequired)
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

	if service, exists := sm.services[serviceName]; exists {
		service.Mutex.Lock()
		service.EnvVars = envVars
		service.Mutex.Unlock()
	}

	return nil
}

// updateConfigurationDefaults updates which configuration is marked as default
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

// saveGlobalConfigToDB persists global configuration to database
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

// loadGlobalConfigFromDB loads global configuration from database
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
