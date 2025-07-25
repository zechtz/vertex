// Package services
package services

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	
	"github.com/zechtz/nest-up/internal/database"
)

// EnvironmentSetup handles environment variable setup and synchronization
type EnvironmentSetup struct {
	workingDir string
	db         *database.Database
}

// NewEnvironmentSetup creates a new environment setup instance
func NewEnvironmentSetup(workingDir string, db *database.Database) *EnvironmentSetup {
	return &EnvironmentSetup{
		workingDir: workingDir,
		db:         db,
	}
}

// DefaultEnvironmentVariables returns the default environment variables for NeST services
func (e *EnvironmentSetup) DefaultEnvironmentVariables() map[string]string {
	return map[string]string{
		// Profile and Feature Flags
		"ACTIVE_PROFILE": "dev",
		"APM_ENABLED":    "false",

		// Config Server Settings
		"CONFIG_USERNAME": "nest",
		"CONFIG_PASSWORD": "1kzwjz2nzegt3nest@ppra.go.tza1q@BmM0Oo",
		"CONFIG_SERVER":   "nest-config-server",

		// Discovery Server (Eureka)
		"DISCOVERY_SERVER": "nest-registry-server",
		"DEFAULT_ZONE":     "http://nest-registry-server:8800/eureka/",

		// Database Settings (PostgreSQL)
		"DB_HOST": "localhost",
		"DB_PORT": "5432",
		"DB_USER": "postgres",
		"DB_PASS": "P057gr35",

		// Gateway
		"GATE_WAY": "nest-gateway",

		// RabbitMQ Settings
		"RABBIT_HOSTNAME": "localhost",
		"RABBIT_PORT":     "5672",
		"RABBIT_USERNAME": "rabbitmq",
		"RABBIT_PASSWORD": "R@bb17mq",

		// Redis Settings
		"REDIS_HOST": "localhost",
		"REDIS_USER": "default",
		"REDIS_PASS": "mypassword",

		// Common Client ID
		"CLIENT_ID": "nest",

		// Service-Specific Database Names
		"DB_NAME_UAA":      "nest_uaa",
		"DB_NAME_APP":      "nest_app",
		"DB_NAME_CONTRACT": "nest_contract",
		"DB_NAME_DSMS":     "nest_dsms",

		// Service Ports
		"SERVICE_PORT_REGISTRY": "8800",
		"SERVICE_PORT_CONFIG":   "8801",
		"SERVICE_PORT_GATEWAY":  "8802",
		"SERVICE_PORT_UAA":      "8803",
		"SERVICE_PORT_APP":      "8805",
		"SERVICE_PORT_CONTRACT": "8818",
		"SERVICE_PORT_DSMS":     "8812",
	}
}

// InitializeDefaultEnvironmentVariables ensures default environment variables are in the database
func (e *EnvironmentSetup) InitializeDefaultEnvironmentVariables() error {
	if e.db == nil {
		return fmt.Errorf("database not available")
	}

	defaultVars := e.DefaultEnvironmentVariables()
	
	// Check if environment variables are already initialized
	existingVars, err := e.db.GetGlobalEnvVars()
	if err != nil {
		log.Printf("[WARN] Failed to check existing environment variables: %v", err)
	}

	// Only add variables that don't exist
	addedCount := 0
	for key, value := range defaultVars {
		if _, exists := existingVars[key]; !exists {
			if err := e.db.SetGlobalEnvVar(key, value); err != nil {
				log.Printf("[WARN] Failed to set environment variable %s: %v", key, err)
			} else {
				addedCount++
			}
		}
	}

	if addedCount > 0 {
		log.Printf("[INFO] Initialized %d default environment variables in database", addedCount)
	} else {
		log.Printf("[INFO] All default environment variables already exist in database")
	}

	return nil
}

// SetupEnvironmentResult represents the result of environment setup
type SetupEnvironmentResult struct {
	Success              bool              `json:"success"`
	Message              string            `json:"message"`
	VariablesSet         int               `json:"variablesSet"`
	ShellProfileUpdated  bool              `json:"shellProfileUpdated"`
	ShellProfile         string            `json:"shellProfile,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVariables"`
	Errors               []string          `json:"errors,omitempty"`
}

// SetupEnvironment sets up the NeST environment variables
func (e *EnvironmentSetup) SetupEnvironment() *SetupEnvironmentResult {
	result := &SetupEnvironmentResult{
		Success:              true,
		Message:              "Environment setup completed successfully",
		EnvironmentVariables: make(map[string]string),
		Errors:               []string{},
	}

	log.Printf("[INFO] Setting up NeST Service Manager environment...")

	// Initialize default variables in database if needed
	if err := e.InitializeDefaultEnvironmentVariables(); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to initialize database variables: %v", err))
	}

	// Load environment variables from database
	if e.db != nil {
		dbVars, err := e.db.GetGlobalEnvVars()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to load variables from database: %v", err))
			// Fallback to default variables
			dbVars = e.DefaultEnvironmentVariables()
		}

		// Set environment variables in the current process
		for key, value := range dbVars {
			if err := os.Setenv(key, value); err != nil {
				errorMsg := fmt.Sprintf("Failed to set environment variable %s: %v", key, err)
				result.Errors = append(result.Errors, errorMsg)
				log.Printf("[WARN] %s", errorMsg)
			} else {
				result.EnvironmentVariables[key] = value
			}
		}
	} else {
		// Fallback to default variables if no database
		defaultVars := e.DefaultEnvironmentVariables()
		for key, value := range defaultVars {
			if err := os.Setenv(key, value); err != nil {
				errorMsg := fmt.Sprintf("Failed to set environment variable %s: %v", key, err)
				result.Errors = append(result.Errors, errorMsg)
				log.Printf("[WARN] %s", errorMsg)
			} else {
				result.EnvironmentVariables[key] = value
			}
		}
	}

	result.VariablesSet = len(result.EnvironmentVariables)

	// Get current environment variables for shell profile and file creation
	currentVars := result.EnvironmentVariables
	if len(currentVars) == 0 {
		// Fallback to default variables if no current variables
		currentVars = e.DefaultEnvironmentVariables()
	}

	// Try to update shell profile for persistence
	if err := e.updateShellProfile(currentVars); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Shell profile update failed: %v", err))
		log.Printf("[WARN] Shell profile update failed: %v", err)
	} else {
		result.ShellProfileUpdated = true
		result.ShellProfile = e.getShellProfile()
	}

	// Create or update common_env_settings.sh file
	if err := e.createCommonEnvFile(currentVars); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to create common_env_settings.sh: %v", err))
		log.Printf("[WARN] Failed to create common_env_settings.sh: %v", err)
	}

	// Update result based on errors
	if len(result.Errors) > 0 {
		result.Success = len(result.Errors) < len(currentVars) // Partial success if some vars were set
		if !result.Success {
			result.Message = "Environment setup failed"
		} else {
			result.Message = fmt.Sprintf("Environment setup completed with %d warnings", len(result.Errors))
		}
	}

	log.Printf("[INFO] Environment setup complete - %d variables set, shell profile updated: %v", 
		result.VariablesSet, result.ShellProfileUpdated)

	return result
}

// SyncEnvironmentFromFile synchronizes environment variables from existing files
func (e *EnvironmentSetup) SyncEnvironmentFromFile() *SetupEnvironmentResult {
	result := &SetupEnvironmentResult{
		Success:              true,
		Message:              "Environment sync completed successfully",
		EnvironmentVariables: make(map[string]string),
		Errors:               []string{},
	}

	log.Printf("[INFO] Syncing environment from existing files...")

	// Try to load from common_env_settings.sh first
	commonEnvFile := filepath.Join(e.workingDir, "common_env_settings.sh")
	if vars, err := e.loadFromShellFile(commonEnvFile); err == nil {
		for key, value := range vars {
			if err := os.Setenv(key, value); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to set %s: %v", key, err))
			} else {
				result.EnvironmentVariables[key] = value
			}
		}
		result.VariablesSet = len(result.EnvironmentVariables)
		log.Printf("[INFO] Loaded %d variables from %s", len(vars), commonEnvFile)
	} else {
		// Fallback to default setup if file doesn't exist
		log.Printf("[INFO] %s not found, using default environment setup", commonEnvFile)
		return e.SetupEnvironment()
	}

	// Try to load from env_vars.fish as well
	fishFile := filepath.Join(e.workingDir, "env_vars.fish")
	if fishVars, err := e.loadFromFishFile(fishFile); err == nil {
		for key, value := range fishVars {
			if _, exists := result.EnvironmentVariables[key]; !exists {
				if err := os.Setenv(key, value); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to set %s from fish file: %v", key, err))
				} else {
					result.EnvironmentVariables[key] = value
					result.VariablesSet++
				}
			}
		}
		log.Printf("[INFO] Loaded additional variables from %s", fishFile)
	}

	if len(result.Errors) > 0 {
		result.Success = result.VariablesSet > 0
		result.Message = fmt.Sprintf("Environment sync completed with %d warnings", len(result.Errors))
	}

	return result
}

// loadFromShellFile loads environment variables from a shell script file
func (e *EnvironmentSetup) loadFromShellFile(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Look for export statements
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
			if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.Trim(strings.TrimSpace(parts[1]), `'"`)
				vars[key] = value
			}
		}
	}

	return vars, scanner.Err()
}

// loadFromFishFile loads environment variables from a fish shell file
func (e *EnvironmentSetup) loadFromFishFile(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Look for set -gx statements
		if strings.HasPrefix(line, "set -gx ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				key := parts[2]
				value := strings.Join(parts[3:], " ")
				value = strings.Trim(value, `'"`)
				vars[key] = value
			}
		}
	}

	return vars, scanner.Err()
}

// createCommonEnvFile creates or updates the common_env_settings.sh file
func (e *EnvironmentSetup) createCommonEnvFile(vars map[string]string) error {
	filePath := filepath.Join(e.workingDir, "common_env_settings.sh")
	
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	_, err = file.WriteString("#!/bin/bash\n")
	if err != nil {
		return err
	}
	_, err = file.WriteString("# Common Environment Variables for NeST Microservices\n")
	if err != nil {
		return err
	}
	_, err = file.WriteString(fmt.Sprintf("# Generated by NeST Service Manager on %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	if err != nil {
		return err
	}

	// Write environment variables in groups
	groups := map[string][]string{
		"Profile and Feature Flags": {"ACTIVE_PROFILE", "APM_ENABLED"},
		"Config Server Settings":     {"CONFIG_USERNAME", "CONFIG_PASSWORD", "CONFIG_SERVER"},
		"Discovery Server":           {"DISCOVERY_SERVER", "DEFAULT_ZONE"},
		"Database Settings":          {"DB_HOST", "DB_PORT", "DB_USER", "DB_PASS"},
		"Gateway":                    {"GATE_WAY"},
		"RabbitMQ Settings":         {"RABBIT_HOSTNAME", "RABBIT_PORT", "RABBIT_USERNAME", "RABBIT_PASSWORD"},
		"Redis Settings":            {"REDIS_HOST", "REDIS_USER", "REDIS_PASS"},
		"Common Client ID":          {"CLIENT_ID"},
		"Service Database Names":    {"DB_NAME_UAA", "DB_NAME_APP", "DB_NAME_CONTRACT", "DB_NAME_DSMS"},
		"Service Ports":             {"SERVICE_PORT_REGISTRY", "SERVICE_PORT_CONFIG", "SERVICE_PORT_GATEWAY", "SERVICE_PORT_UAA", "SERVICE_PORT_APP", "SERVICE_PORT_CONTRACT", "SERVICE_PORT_DSMS"},
	}

	for groupName, keys := range groups {
		_, err = file.WriteString(fmt.Sprintf("# %s\n", groupName))
		if err != nil {
			return err
		}

		for _, key := range keys {
			if value, exists := vars[key]; exists {
				_, err = file.WriteString(fmt.Sprintf("export %s='%s'\n", key, value))
				if err != nil {
					return err
				}
			}
		}
		_, err = file.WriteString("\n")
		if err != nil {
			return err
		}
	}

	// Write footer
	_, err = file.WriteString("echo \"✅ NeST common environment variables loaded successfully\"\n")
	if err != nil {
		return err
	}

	return nil
}

// updateShellProfile updates the user's shell profile to source environment variables
func (e *EnvironmentSetup) updateShellProfile(vars map[string]string) error {
	shellProfile := e.getShellProfile()
	if shellProfile == "" {
		return fmt.Errorf("unable to determine shell profile")
	}

	// Check if file exists
	if _, err := os.Stat(shellProfile); os.IsNotExist(err) {
		// Create the file if it doesn't exist
		if file, err := os.Create(shellProfile); err != nil {
			return fmt.Errorf("failed to create shell profile %s: %w", shellProfile, err)
		} else {
			file.Close()
		}
	}

	// Read current content
	content, err := os.ReadFile(shellProfile)
	if err != nil {
		return fmt.Errorf("failed to read shell profile %s: %w", shellProfile, err)
	}

	contentStr := string(content)
	commonEnvFile := filepath.Join(e.workingDir, "common_env_settings.sh")
	sourceLine := fmt.Sprintf("source \"%s\"", commonEnvFile)

	// Check if already configured
	if strings.Contains(contentStr, "common_env_settings.sh") {
		log.Printf("[INFO] Shell profile %s already configured", shellProfile)
		return nil
	}

	// Append the configuration
	file, err := os.OpenFile(shellProfile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open shell profile %s for writing: %w", shellProfile, err)
	}
	defer file.Close()

	_, err = file.WriteString("\n# NeST Service Manager Environment Variables\n")
	if err != nil {
		return err
	}
	_, err = file.WriteString(sourceLine + "\n")
	if err != nil {
		return err
	}

	log.Printf("[INFO] Added environment loading to %s", shellProfile)
	return nil
}

// getShellProfile determines the appropriate shell profile file
func (e *EnvironmentSetup) getShellProfile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Check for zsh first (most common on macOS)
	if _, err := os.Stat("/bin/zsh"); err == nil {
		return filepath.Join(homeDir, ".zshrc")
	}

	// Check for bash
	if _, err := os.Stat("/bin/bash"); err == nil {
		return filepath.Join(homeDir, ".bashrc")
	}

	// On macOS, bash uses .bash_profile
	if runtime.GOOS == "darwin" {
		return filepath.Join(homeDir, ".bash_profile")
	}

	return filepath.Join(homeDir, ".bashrc")
}

// GetCurrentEnvironment returns the current environment variables
func (e *EnvironmentSetup) GetCurrentEnvironment() map[string]string {
	defaultVars := e.DefaultEnvironmentVariables()
	currentEnv := make(map[string]string)

	for key := range defaultVars {
		if value := os.Getenv(key); value != "" {
			currentEnv[key] = value
		}
	}

	return currentEnv
}

// CheckEnvironmentStatus checks if the environment is properly configured
func (e *EnvironmentSetup) CheckEnvironmentStatus() map[string]interface{} {
	defaultVars := e.DefaultEnvironmentVariables()
	status := map[string]interface{}{
		"configured":     0,
		"missing":        0,
		"total":          len(defaultVars),
		"missingVars":    []string{},
		"configuredVars": []string{},
	}

	for key := range defaultVars {
		if value := os.Getenv(key); value != "" {
			status["configured"] = status["configured"].(int) + 1
			status["configuredVars"] = append(status["configuredVars"].([]string), key)
		} else {
			status["missing"] = status["missing"].(int) + 1
			status["missingVars"] = append(status["missingVars"].([]string), key)
		}
	}

	return status
}