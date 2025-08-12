package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zechtz/vertex/internal/database"
	"github.com/zechtz/vertex/internal/models"
)

type ProfileService struct {
	db    *database.Database
	sm    *Manager // Access to service manager for service operations
	mutex sync.RWMutex
}

func NewProfileService(db *database.Database, sm *Manager) *ProfileService {
	return &ProfileService{
		db:    db,
		sm:    sm,
		mutex: sync.RWMutex{},
	}
}

// GetDatabase returns the database instance for external access
func (ps *ProfileService) GetDatabase() *database.Database {
	return ps.db
}

// GetUserProfile retrieves or creates a user profile
func (ps *ProfileService) GetUserProfile(userID string) (*models.UserProfile, error) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	var profile models.UserProfile
	var preferencesJSON string

	query := `SELECT user_id, display_name, avatar, preferences_json, created_at, updated_at 
			  FROM user_profiles WHERE user_id = ?`

	err := ps.db.QueryRow(query, userID).Scan(
		&profile.UserID,
		&profile.DisplayName,
		&profile.Avatar,
		&preferencesJSON,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create default profile
			return ps.createDefaultUserProfile(userID)
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Parse preferences JSON
	if err := json.Unmarshal([]byte(preferencesJSON), &profile.Preferences); err != nil {
		return nil, fmt.Errorf("failed to parse preferences: %w", err)
	}

	return &profile, nil
}

// UpdateUserProfile updates user profile information
func (ps *ProfileService) UpdateUserProfile(userID string, req *models.UserProfileUpdateRequest) (*models.UserProfile, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	preferencesJSON, err := json.Marshal(req.Preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal preferences: %w", err)
	}

	query := `UPDATE user_profiles 
			  SET display_name = ?, avatar = ?, preferences_json = ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE user_id = ?`

	_, err = ps.db.Exec(query, req.DisplayName, req.Avatar, string(preferencesJSON), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	// Use internal method to avoid deadlock (we already hold the lock)
	return ps.getUserProfileInternal(userID)
}

// getUserProfileInternal retrieves a user profile without acquiring locks (for internal use)
func (ps *ProfileService) getUserProfileInternal(userID string) (*models.UserProfile, error) {
	var profile models.UserProfile
	var preferencesJSON string

	query := `SELECT user_id, display_name, avatar, preferences_json, created_at, updated_at 
			  FROM user_profiles WHERE user_id = ?`

	err := ps.db.QueryRow(query, userID).Scan(
		&profile.UserID,
		&profile.DisplayName,
		&profile.Avatar,
		&preferencesJSON,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create default profile (but we need to release lock first)
			return nil, fmt.Errorf("user profile not found")
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Parse preferences JSON
	if err := json.Unmarshal([]byte(preferencesJSON), &profile.Preferences); err != nil {
		return nil, fmt.Errorf("failed to parse preferences: %w", err)
	}

	return &profile, nil
}

// GetServiceProfiles retrieves all service profiles for a user
func (ps *ProfileService) GetServiceProfiles(userID string) ([]models.ServiceProfile, error) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	query := `SELECT id, user_id, name, description, services_json, env_vars_json, projects_dir, java_home_override, is_default, is_active, created_at, updated_at 
			  FROM service_profiles WHERE user_id = ? ORDER BY is_active DESC, is_default DESC, created_at DESC`

	rows, err := ps.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query service profiles: %w", err)
	}
	defer rows.Close()

	var profiles []models.ServiceProfile
	for rows.Next() {
		var profile models.ServiceProfile
		var servicesJSON, envVarsJSON string

		err := rows.Scan(
			&profile.ID,
			&profile.UserID,
			&profile.Name,
			&profile.Description,
			&servicesJSON,
			&envVarsJSON,
			&profile.ProjectsDir,
			&profile.JavaHomeOverride,
			&profile.IsDefault,
			&profile.IsActive,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service profile: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(servicesJSON), &profile.Services); err != nil {
			return nil, fmt.Errorf("failed to parse services JSON: %w", err)
		}
		if err := json.Unmarshal([]byte(envVarsJSON), &profile.EnvVars); err != nil {
			return nil, fmt.Errorf("failed to parse env vars JSON: %w", err)
		}

		profiles = append(profiles, profile)
	}

	return profiles, rows.Err()
}

// GetServiceProfile retrieves a specific service profile
func (ps *ProfileService) GetServiceProfile(profileID, userID string) (*models.ServiceProfile, error) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	return ps.getServiceProfileInternal(profileID, userID)
}

// getServiceProfileInternal retrieves a service profile without acquiring locks (for internal use)
func (ps *ProfileService) getServiceProfileInternal(profileID, userID string) (*models.ServiceProfile, error) {
	var profile models.ServiceProfile
	var servicesJSON, envVarsJSON string

	query := `SELECT id, user_id, name, description, services_json, env_vars_json, projects_dir, java_home_override, is_default, is_active, created_at, updated_at 
			  FROM service_profiles WHERE id = ? AND user_id = ?`

	err := ps.db.QueryRow(query, profileID, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Name,
		&profile.Description,
		&servicesJSON,
		&envVarsJSON,
		&profile.ProjectsDir,
		&profile.JavaHomeOverride,
		&profile.IsDefault,
		&profile.IsActive,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("service profile not found")
		}
		return nil, fmt.Errorf("failed to get service profile: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(servicesJSON), &profile.Services); err != nil {
		return nil, fmt.Errorf("failed to parse services JSON: %w", err)
	}
	if err := json.Unmarshal([]byte(envVarsJSON), &profile.EnvVars); err != nil {
		return nil, fmt.Errorf("failed to parse env vars JSON: %w", err)
	}

	return &profile, nil
}

// CreateServiceProfile creates a new service profile
func (ps *ProfileService) CreateServiceProfile(userID string, req *models.CreateProfileRequest) (*models.ServiceProfile, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	log.Printf("[DEBUG] Creating profile for user %s: %+v", userID, req)

	// Generate unique ID
	profileID := uuid.New().String()

	// Handle default profile logic
	if req.IsDefault {
		if err := ps.clearDefaultProfiles(userID); err != nil {
			return nil, fmt.Errorf("failed to clear existing default profiles: %w", err)
		}
	}

	// Validate services exist (temporarily disabled for debugging)
	log.Printf("[DEBUG] Skipping service validation for debugging purposes")
	// if err := ps.validateServices(req.Services); err != nil {
	// 	return nil, fmt.Errorf("invalid services: %w", err)
	// }

	// Marshal JSON fields
	servicesJSON, err := json.Marshal(req.Services)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal services: %w", err)
	}

	// Handle nil envVars
	envVars := req.EnvVars
	if envVars == nil {
		envVars = make(map[string]string)
	}
	envVarsJSON, err := json.Marshal(envVars)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal env vars: %w", err)
	}

	query := `INSERT INTO service_profiles (id, user_id, name, description, services_json, env_vars_json, projects_dir, java_home_override, is_default, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err = ps.db.Exec(query, profileID, userID, req.Name, req.Description, string(servicesJSON), string(envVarsJSON), req.ProjectsDir, req.JavaHomeOverride, req.IsDefault)
	if err != nil {
		return nil, fmt.Errorf("failed to create service profile: %w", err)
	}

	return ps.getServiceProfileInternal(profileID, userID)
}

// UpdateServiceProfile updates an existing service profile
func (ps *ProfileService) UpdateServiceProfile(profileID, userID string, req *models.UpdateProfileRequest) (*models.ServiceProfile, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	log.Printf("[DEBUG] Updating profile %s for user %s: %+v", profileID, userID, req)

	// Check if profile exists and belongs to user
	log.Printf("[DEBUG] Checking if profile exists...")
	if _, err := ps.getServiceProfileInternal(profileID, userID); err != nil {
		log.Printf("[ERROR] Profile not found: %v", err)
		return nil, err
	}
	log.Printf("[DEBUG] Profile exists, proceeding...")

	// Handle default profile logic
	if req.IsDefault {
		log.Printf("[DEBUG] Clearing existing default profiles...")
		if err := ps.clearDefaultProfiles(userID); err != nil {
			log.Printf("[ERROR] Failed to clear default profiles: %v", err)
			return nil, fmt.Errorf("failed to clear existing default profiles: %w", err)
		}
		log.Printf("[DEBUG] Default profiles cleared")
	}

	// Validate services exist (temporarily disabled for debugging)
	log.Printf("[DEBUG] Skipping service validation for debugging purposes")
	// if err := ps.validateServices(req.Services); err != nil {
	// 	return nil, fmt.Errorf("invalid services: %w", err)
	// }

	// Marshal JSON fields
	log.Printf("[DEBUG] Marshaling services JSON...")
	servicesJSON, err := json.Marshal(req.Services)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal services: %v", err)
		return nil, fmt.Errorf("failed to marshal services: %w", err)
	}
	log.Printf("[DEBUG] Services JSON: %s", string(servicesJSON))

	// Handle nil envVars
	envVars := req.EnvVars
	if envVars == nil {
		envVars = make(map[string]string)
	}

	log.Printf("[DEBUG] Marshaling env vars JSON...")
	envVarsJSON, err := json.Marshal(envVars)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal env vars: %v", err)
		return nil, fmt.Errorf("failed to marshal env vars: %w", err)
	}

	log.Printf("[DEBUG] EnvVars JSON: %s", string(envVarsJSON))

	query := `UPDATE service_profiles 
			  SET name = ?, description = ?, services_json = ?, env_vars_json = ?, projects_dir = ?, java_home_override = ?, is_default = ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = ? AND user_id = ?`

	log.Printf("[DEBUG] Executing database update...")

	_, err = ps.db.Exec(query, req.Name, req.Description, string(servicesJSON), string(envVarsJSON), req.ProjectsDir, req.JavaHomeOverride, req.IsDefault, profileID, userID)
	if err != nil {
		log.Printf("[ERROR] Database update failed: %v", err)
		return nil, fmt.Errorf("failed to update service profile: %w", err)
	}
	log.Printf("[DEBUG] Database update successful")

	log.Printf("[DEBUG] Fetching updated profile...")
	result, err := ps.getServiceProfileInternal(profileID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch updated profile: %v", err)
		return nil, err
	}
	log.Printf("[DEBUG] Update operation completed")
	return result, nil
}

// DeleteServiceProfile deletes a service profile
func (ps *ProfileService) DeleteServiceProfile(profileID, userID string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	// Check if profile exists and belongs to user
	if _, err := ps.GetServiceProfile(profileID, userID); err != nil {
		return err
	}

	query := `DELETE FROM service_profiles WHERE id = ? AND user_id = ?`
	result, err := ps.db.Exec(query, profileID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete service profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no profile deleted")
	}

	return nil
}

// ApplyProfile applies a service profile by configuring and starting the specified services
func (ps *ProfileService) ApplyProfile(profileID, userID string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	log.Printf("[INFO] Applying profile %s for user %s", profileID, userID)

	// Get profile
	profile, err := ps.getServiceProfileInternal(profileID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get profile: %v", err)
		return fmt.Errorf("failed to get profile: %w", err)
	}

	log.Printf("[INFO] Profile '%s' loaded with %d services", profile.Name, len(profile.Services))

	// Validate services exist before applying profile
	if err := ps.validateServicesForProfile(profile); err != nil {
		log.Printf("[ERROR] Service validation failed: %v", err)
		return fmt.Errorf("profile validation failed: %w", err)
	}

	// Apply projects directory if specified
	if profile.ProjectsDir != "" {
		log.Printf("[INFO] Setting projects directory to: %s", profile.ProjectsDir)
		if err := ps.applyProjectsDirectory(profile.ProjectsDir); err != nil {
			log.Printf("[WARN] Failed to apply projects directory: %v", err)
			// Don't fail the entire operation for this
		}
	}

	// Apply Java home override if specified
	if profile.JavaHomeOverride != "" {
		log.Printf("[INFO] Setting Java home override to: %s", profile.JavaHomeOverride)
		if err := ps.applyJavaHomeOverride(profile.JavaHomeOverride); err != nil {
			log.Printf("[WARN] Failed to apply Java home override: %v", err)
			// Don't fail the entire operation for this
		}
	}

	// Apply environment variables if any
	if len(profile.EnvVars) > 0 {
		log.Printf("[INFO] Applying %d environment variables", len(profile.EnvVars))
		if err := ps.applyEnvironmentVariables(profile.EnvVars); err != nil {
			log.Printf("[ERROR] Failed to apply environment variables: %v", err)
			return fmt.Errorf("failed to apply environment variables: %w", err)
		}
	}

	// Stop all services first for clean slate
	if ps.sm != nil {
		log.Printf("[INFO] Stopping all running services")
		if err := ps.sm.StopAllServices(); err != nil {
			log.Printf("[WARN] Failed to stop some services: %v", err)
			// Continue anyway as some services might not be running
		}
	}

	// Start services specified in profile with dependency ordering
	if ps.sm != nil && len(profile.Services) > 0 {
		log.Printf("[INFO] Starting %d services from profile", len(profile.Services))

		// Use dependency-aware startup for better reliability
		if err := ps.startServicesWithDependencies(profile.Services); err != nil {
			log.Printf("[ERROR] Failed to start services: %v", err)
			return fmt.Errorf("failed to start services: %w", err)
		}
	} else if len(profile.Services) == 0 {
		log.Printf("[INFO] Profile '%s' has no services configured - skipping service startup", profile.Name)
	}

	// Mark profile as default if specified
	if profile.IsDefault {
		log.Printf("[INFO] Marking profile '%s' as default", profile.Name)
		if err := ps.clearDefaultProfiles(userID); err != nil {
			log.Printf("[WARN] Failed to clear other default profiles: %v", err)
		}
		// Update this profile to be default (in case it wasn't already)
		ps.db.Exec("UPDATE service_profiles SET is_default = TRUE WHERE id = ? AND user_id = ?", profileID, userID)
	}

	log.Printf("[INFO] Profile '%s' applied successfully", profile.Name)
	return nil
}

// SetActiveProfile sets a profile as the active profile for a user
func (ps *ProfileService) SetActiveProfile(userID, profileID string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	log.Printf("[INFO] Setting active profile %s for user %s", profileID, userID)

	// Verify the profile exists and belongs to the user
	_, err := ps.getServiceProfileInternal(profileID, userID)
	if err != nil {
		return fmt.Errorf("profile validation failed: %w", err)
	}

	// Set the active profile in the database
	if err := ps.db.SetActiveProfile(userID, profileID); err != nil {
		return fmt.Errorf("failed to set active profile: %w", err)
	}

	log.Printf("[INFO] Active profile set successfully")
	return nil
}

// GetActiveProfile gets the active profile for a user
func (ps *ProfileService) GetActiveProfile(userID string) (*models.ServiceProfile, error) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	activeProfileID, err := ps.db.GetActiveProfile(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile ID: %w", err)
	}

	if activeProfileID == "" {
		return nil, nil // No active profile
	}

	return ps.getServiceProfileInternal(activeProfileID, userID)
}

// GetProfileContext retrieves the complete configuration context for a profile
func (ps *ProfileService) GetProfileContext(userID, profileID string) (*models.ProfileContext, error) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	// Get the profile
	profile, err := ps.getServiceProfileInternal(profileID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Get profile-scoped environment variables
	envVars, err := ps.db.GetProfileEnvVars(profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile env vars: %w", err)
	}

	// Get service configurations for all services in the profile
	serviceConfigs := make(map[string]map[string]string)
	for _, serviceName := range profile.Services {
		config, err := ps.db.GetProfileServiceConfig(profileID, serviceName)
		if err != nil {
			log.Printf("[WARN] Failed to get service config for %s: %v", serviceName, err)
			continue
		}
		if len(config) > 0 {
			serviceConfigs[serviceName] = config
		}
	}

	// TODO: Add profile dependencies when dependency management is enhanced

	return &models.ProfileContext{
		Profile:        profile,
		EnvVars:        envVars,
		ServiceConfigs: serviceConfigs,
		Dependencies:   make(map[string][]models.ProfileDependency), // Placeholder
		IsActive:       profile.IsActive,
	}, nil
}

// SetProfileEnvVar sets an environment variable for a specific profile
func (ps *ProfileService) SetProfileEnvVar(userID, profileID, name, value, description string, isRequired bool) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	// Verify the profile exists and belongs to the user
	_, err := ps.getServiceProfileInternal(profileID, userID)
	if err != nil {
		return fmt.Errorf("profile validation failed: %w", err)
	}

	return ps.db.SetProfileEnvVar(profileID, name, value, description, isRequired)
}

// GetProfileEnvVars gets all environment variables for a specific profile
func (ps *ProfileService) GetProfileEnvVars(userID, profileID string) (map[string]string, error) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	// Verify the profile exists and belongs to the user
	_, err := ps.getServiceProfileInternal(profileID, userID)
	if err != nil {
		return nil, fmt.Errorf("profile validation failed: %w", err)
	}

	return ps.db.GetProfileEnvVars(profileID)
}

// DeleteProfileEnvVar deletes an environment variable for a specific profile
func (ps *ProfileService) DeleteProfileEnvVar(userID, profileID, name string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	// Verify the profile exists and belongs to the user
	_, err := ps.getServiceProfileInternal(profileID, userID)
	if err != nil {
		return fmt.Errorf("profile validation failed: %w", err)
	}

	return ps.db.DeleteProfileEnvVar(profileID, name)
}

// AddServiceToProfile adds a service to a profile's services list
func (ps *ProfileService) AddServiceToProfile(userID, profileID, serviceUUID string) error {
	ps.mutex.Lock()

	// Get the current profile
	profile, err := ps.getServiceProfileInternal(profileID, userID)
	if err != nil {
		ps.mutex.Unlock()
		return fmt.Errorf("profile validation failed: %w", err)
	}

	// Check if service is already in the profile
	if slices.Contains(profile.Services, serviceUUID) {
		ps.mutex.Unlock()
		return fmt.Errorf("service '%s' already exists in profile '%s'", serviceUUID, profile.Name)
	}

	// Verify that the service exists globally
	if _, exists := ps.sm.GetServiceByUUID(serviceUUID); !exists {
		ps.mutex.Unlock()
		return fmt.Errorf("service '%s' does not exist", serviceUUID)
	}

	// Add the service to the profile's services list
	updatedServices := append(profile.Services, serviceUUID)

	// Create update request
	updateReq := &models.UpdateProfileRequest{
		Name:             profile.Name,
		Description:      profile.Description,
		ProjectsDir:      profile.ProjectsDir,
		JavaHomeOverride: profile.JavaHomeOverride,
		Services:         updatedServices,
		EnvVars:          profile.EnvVars,
		IsDefault:        profile.IsDefault,
	}

	// Release the mutex before calling UpdateServiceProfile to avoid deadlock
	ps.mutex.Unlock()

	// Update the profile with the new services list
	_, err = ps.UpdateServiceProfile(profileID, userID, updateReq)
	return err
}

func (ps *ProfileService) RemoveServiceFromProfile(userID, profileID, serviceName string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	// Get the current profile
	profile, err := ps.getServiceProfileInternal(profileID, userID)
	if err != nil {
		return fmt.Errorf("profile validation failed: %w", err)
	}

	// Check if service is in the profile
	serviceIndex := -1
	for i, service := range profile.Services {
		if service == serviceName {
			serviceIndex = i
			break
		}
	}

	if serviceIndex == -1 {
		return fmt.Errorf("service '%s' not found in profile '%s'", serviceName, profile.Name)
	}

	// Remove the service from the profile's services list
	updatedServices := make([]string, 0, len(profile.Services)-1)
	updatedServices = append(updatedServices, profile.Services[:serviceIndex]...)
	updatedServices = append(updatedServices, profile.Services[serviceIndex+1:]...)

	// Update the profile with the new services list
	updateReq := &models.UpdateProfileRequest{
		Name:             profile.Name,
		Description:      profile.Description,
		Services:         updatedServices,
		EnvVars:          profile.EnvVars,
		ProjectsDir:      profile.ProjectsDir,
		JavaHomeOverride: profile.JavaHomeOverride,
		IsDefault:        profile.IsDefault,
	}

	// Update the profile - we already have the lock, so unlock temporarily
	ps.mutex.Unlock()
	_, err = ps.UpdateServiceProfile(profileID, userID, updateReq)
	ps.mutex.Lock()
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	log.Printf("[INFO] Removed service '%s' from profile '%s'", serviceName, profile.Name)
	return nil
}

// Helper methods

func (ps *ProfileService) createDefaultUserProfile(userID string) (*models.UserProfile, error) {
	defaultPreferences := models.UserPreferences{
		Theme:    "light",
		Language: "en",
		NotificationSettings: map[string]bool{
			"serviceStatus": true,
			"errors":        true,
			"deployments":   true,
		},
		DashboardLayout: "grid",
		AutoRefresh:     true,
		RefreshInterval: 30,
	}

	preferencesJSON, err := json.Marshal(defaultPreferences)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal default preferences: %w", err)
	}

	query := `INSERT INTO user_profiles (user_id, display_name, avatar, preferences_json, created_at, updated_at)
			  VALUES (?, '', '', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err = ps.db.Exec(query, userID, string(preferencesJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create default user profile: %w", err)
	}

	return ps.GetUserProfile(userID)
}

func (ps *ProfileService) clearDefaultProfiles(userID string) error {
	query := `UPDATE service_profiles SET is_default = FALSE WHERE user_id = ? AND is_default = TRUE`
	_, err := ps.db.Exec(query, userID)
	return err
}

func (ps *ProfileService) validateServices(serviceNames []string) error {
	if ps.sm == nil {
		log.Printf("[WARN] Service manager not available, skipping service validation")
		return nil // Skip validation if service manager not available
	}

	// Allow empty service lists - profiles can be created without services
	if len(serviceNames) == 0 {
		log.Printf("[DEBUG] Profile created with no services - this is allowed")
		return nil
	}

	availableServices := ps.sm.GetServices()
	availableMap := make(map[string]bool)
	for _, service := range availableServices {
		availableMap[service.Name] = true
	}

	log.Printf("[DEBUG] Available services: %v", availableServices)
	log.Printf("[DEBUG] Validating services: %v", serviceNames)

	for _, serviceName := range serviceNames {
		if !availableMap[serviceName] {
			return fmt.Errorf("service '%s' not found", serviceName)
		}
	}

	return nil
}

func (ps *ProfileService) applyEnvironmentVariables(envVars map[string]string) error {
	for name, value := range envVars {
		if err := ps.db.SetGlobalEnvVar(name, value); err != nil {
			return fmt.Errorf("failed to set env var %s: %w", name, err)
		}
	}
	return nil
}

// validateServicesForProfile validates that all services in a profile exist and are available
func (ps *ProfileService) validateServicesForProfile(profile *models.ServiceProfile) error {
	if ps.sm == nil {
		return fmt.Errorf("service manager not available")
	}

	// Allow profiles with no services - they can be configured later
	if len(profile.Services) == 0 {
		log.Printf("[DEBUG] Profile '%s' has no services configured - this is allowed", profile.Name)
		return nil
	}

	availableServices := ps.sm.GetServices()
	availableMap := make(map[string]bool)
	for _, service := range availableServices {
		availableMap[service.Name] = true
	}

	var missingServices []string
	for _, serviceName := range profile.Services {
		if !availableMap[serviceName] {
			missingServices = append(missingServices, serviceName)
		}
	}

	if len(missingServices) > 0 {
		return fmt.Errorf("services not found: %v", missingServices)
	}

	return nil
}

// applyProjectsDirectory sets the projects directory for service operations
func (ps *ProfileService) applyProjectsDirectory(projectsDir string) error {
	// Set the PROJECTS_DIR environment variable
	if err := ps.db.SetGlobalEnvVar("PROJECTS_DIR", projectsDir); err != nil {
		return fmt.Errorf("failed to set PROJECTS_DIR: %w", err)
	}

	// Also update the working directory for the service manager if possible
	if ps.sm != nil {
		// This would be implementation-specific to your service manager
		log.Printf("[INFO] Projects directory set to: %s", projectsDir)
	}

	return nil
}

// applyJavaHomeOverride sets the Java home override for service operations
func (ps *ProfileService) applyJavaHomeOverride(javaHome string) error {
	// Set the JAVA_HOME environment variable
	if err := ps.db.SetGlobalEnvVar("JAVA_HOME", javaHome); err != nil {
		return fmt.Errorf("failed to set JAVA_HOME: %w", err)
	}

	// Also set PATH to include the Java bin directory
	javaPath := javaHome + "/bin"
	if err := ps.db.SetGlobalEnvVar("JAVA_PATH_OVERRIDE", javaPath); err != nil {
		return fmt.Errorf("failed to set JAVA_PATH_OVERRIDE: %w", err)
	}

	log.Printf("[INFO] Java home override set to: %s", javaHome)
	return nil
}

// startServicesWithDependencies starts services in dependency order
func (ps *ProfileService) startServicesWithDependencies(serviceNames []string) error {
	if ps.sm == nil {
		return fmt.Errorf("service manager not available")
	}

	// Get all available services with their dependency information
	allServices := ps.sm.GetServices()
	serviceMap := make(map[string]*models.Service)
	for _, service := range allServices {
		serviceMap[service.Name] = &service
	}

	// Filter to only the services we want to start
	var servicesToStart []*models.Service
	for _, name := range serviceNames {
		if service, exists := serviceMap[name]; exists {
			servicesToStart = append(servicesToStart, service)
		} else {
			log.Printf("[WARN] Service '%s' not found, skipping", name)
		}
	}

	if len(servicesToStart) == 0 {
		return fmt.Errorf("no valid services to start")
	}

	// Sort services by their order field (this provides basic dependency ordering)
	// For more sophisticated dependency management, we'd use the dependency graph
	sortedServices := make([]*models.Service, len(servicesToStart))
	copy(sortedServices, servicesToStart)

	// Simple bubble sort by order field
	for i := 0; i < len(sortedServices)-1; i++ {
		for j := 0; j < len(sortedServices)-i-1; j++ {
			if sortedServices[j].Order > sortedServices[j+1].Order {
				sortedServices[j], sortedServices[j+1] = sortedServices[j+1], sortedServices[j]
			}
		}
	}

	// Start services in dependency order
	for _, service := range sortedServices {
		log.Printf("[INFO] Starting service: %s (order: %d)", service.Name, service.Order)

		if err := ps.sm.StartService(service.Name); err != nil {
			log.Printf("[ERROR] Failed to start service %s: %v", service.Name, err)
			// Continue starting other services rather than failing completely
			continue
		}

		// Brief delay between service starts to allow proper initialization
		time.Sleep(2 * time.Second)
	}

	return nil
}

func (ps *ProfileService) ProfileHasService(profileID, serviceUUID string) (bool, error) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	// Get the profile
	profile, err := ps.getServiceProfileInternal(profileID, "")
	if err != nil {
		return false, fmt.Errorf("failed to get profile: %w", err)
	}

	// Check if the service exists in the profile
	if slices.Contains(profile.Services, serviceUUID) {
		return true, nil
	}

	return false, nil
}
