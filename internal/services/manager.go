package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zechtz/vertex/internal/database"
	"github.com/zechtz/vertex/internal/models"
)

type Manager struct {
	config            models.Config
	services          map[string]*models.Service
	configurations    map[string]*models.Configuration
	activeConfigID    string
	db                *database.Database
	mutex             sync.RWMutex
	clients           map[*websocket.Conn]bool
	clientsMutex      sync.RWMutex
	dependencyManager *DependencyManager
	Id                int64
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewManager(config models.Config, db *database.Database) (*Manager, error) {
	sm := &Manager{
		config:         config,
		services:       make(map[string]*models.Service),
		configurations: make(map[string]*models.Configuration),
		activeConfigID: "default",
		db:             db,
		clients:        make(map[*websocket.Conn]bool),
	}

	// Initialize dependency manager
	sm.dependencyManager = NewDependencyManager(sm)

	// Load or create services
	if err := sm.loadServices(config); err != nil {
		return nil, fmt.Errorf("failed to load services: %w", err)
	}

	// Load global configuration from database (override defaults)
	if err := sm.loadGlobalConfigFromDB(); err != nil {
		log.Printf("Warning: Could not load global config from database: %v", err)
	}

	// Load configurations from database
	if err := sm.loadConfigurations(); err != nil {
		return nil, fmt.Errorf("failed to load configurations: %w", err)
	}

	// Load environment variables from fish file into database
	if err := sm.loadEnvVarsFromFishFile(); err != nil {
		log.Printf("Warning: Could not load environment variables from fish file: %v", err)
	}

	// Initialize default service dependencies
	if err := sm.dependencyManager.InitializeDefaultDependencies(); err != nil {
		log.Printf("Warning: Could not initialize service dependencies: %v", err)
	}

	// Validate dependencies
	if err := sm.dependencyManager.ValidateDependencies(); err != nil {
		log.Printf("Warning: Dependency validation failed: %v", err)
	}

	// Start health check routine
	go sm.healthCheckRoutine()

	// Start resource metrics collection
	go sm.startMetricsCollection()

	// Start periodic log cleanup (daily)
	go sm.startLogCleanupRoutine()

	return sm, nil
}

func (sm *Manager) AddWebSocketClient(conn *websocket.Conn) {
	sm.clientsMutex.Lock()
	sm.clients[conn] = true
	sm.clientsMutex.Unlock()
}

func (sm *Manager) RemoveWebSocketClient(conn *websocket.Conn) {
	sm.clientsMutex.Lock()
	delete(sm.clients, conn)
	sm.clientsMutex.Unlock()
}

func (sm *Manager) GetServices() []models.Service {
	sm.mutex.RLock()
	services := make([]models.Service, 0, len(sm.services))
	for _, service := range sm.services {
		service.Mutex.RLock()
		services = append(services, *service)
		service.Mutex.RUnlock()
	}
	sm.mutex.RUnlock()

	// Sort by order
	sort.Slice(services, func(i, j int) bool {
		return services[i].Order < services[j].Order
	})

	return services
}

// NormalizeServiceOrders ensures service orders are sequential from 1 to N
func (sm *Manager) NormalizeServiceOrders() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Get all services and sort them by current order
	services := make([]*models.Service, 0, len(sm.services))
	for _, service := range sm.services {
		services = append(services, service)
	}

	// Sort by current order
	sort.Slice(services, func(i, j int) bool {
		return services[i].Order < services[j].Order
	})

	// Normalize orders to 1, 2, 3, ... N
	hasChanges := false
	for i, service := range services {
		newOrder := i + 1
		if service.Order != newOrder {
			service.Order = newOrder
			hasChanges = true

			// Update in database
			if err := sm.UpdateServiceConfigInDB(service); err != nil {
				log.Printf("[ERROR] Failed to update service order for UUID %s: %v", service.ID, err)
				return fmt.Errorf("failed to normalize order for service UUID %s: %w", service.ID, err)
			}
		}
	}

	if hasChanges {
		log.Printf("[INFO] Normalized service orders - services now ordered 1 to %d", len(services))
		// Broadcast updates for all changed services
		for _, service := range services {
			sm.broadcastUpdate(service)
		}
	}

	return nil
}

// GetServiceByUUID retrieves a service by its UUID from the services map.
// It returns the service and a boolean indicating whether the service was found.
// If the UUID is empty, it returns nil and false.
func (sm *Manager) GetServiceByUUID(uuid string) (*models.Service, bool) {
	if uuid == "" {
		log.Printf("[WARN] Empty UUID provided for service lookup")
		return nil, false
	}

	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	service, exists := sm.services[uuid]
	return service, exists
}

func (sm *Manager) GetDatabase() *database.Database {
	return sm.db
}

func (sm *Manager) GetConfigurations() []*models.Configuration {
	sm.mutex.RLock()
	configs := make([]*models.Configuration, 0, len(sm.configurations))
	for _, config := range sm.configurations {
		configs = append(configs, config)
	}
	sm.mutex.RUnlock()
	return configs
}

func (sm *Manager) GetConfig() models.Config {
	return sm.config
}

func (sm *Manager) broadcastUpdate(service *models.Service) {
	sm.clientsMutex.Lock()
	defer sm.clientsMutex.Unlock()

	// Create a list of clients to remove (to avoid concurrent map modification)
	var clientsToRemove []*websocket.Conn

	for client := range sm.clients {
		if err := client.WriteJSON(WebSocketMessage{Type: "service_update", Payload: service}); err != nil {
			// Mark client for removal
			clientsToRemove = append(clientsToRemove, client)
		}
	}

	// Remove failed clients
	for _, client := range clientsToRemove {
		delete(sm.clients, client)
		client.Close()
	}
}

func (sm *Manager) broadcastLogEntry(serviceUUID string, logEntry models.LogEntry) {
	sm.clientsMutex.Lock()
	defer sm.clientsMutex.Unlock()

	_, exists := sm.services[serviceUUID]
	if !exists {
		log.Printf("[WARN] Service UUID %s not found for log broadcast", serviceUUID)
		return
	}

	message := WebSocketMessage{
		Type: "log_entry",
		Payload: struct {
			ServiceUUID string          `json:"serviceUUID"`
			LogEntry    models.LogEntry `json:"logEntry"`
		}{
			ServiceUUID: serviceUUID,
			LogEntry:    logEntry,
		},
	}

	var clientsToRemove []*websocket.Conn
	for client := range sm.clients {
		if err := client.WriteJSON(message); err != nil {
			clientsToRemove = append(clientsToRemove, client)
		}
	}

	for _, client := range clientsToRemove {
		delete(sm.clients, client)
		client.Close()
	}
}

func (sm *Manager) GracefulShutdown() {
	log.Printf("[INFO] %s - Stopping all running services...", time.Now().Format("2006-01-02 15:04:05"))

	// Get all running services
	sm.mutex.RLock()
	runningServices := make([]*models.Service, 0)
	for _, service := range sm.services {
		service.Mutex.RLock()
		if service.Status == "running" {
			runningServices = append(runningServices, service)
		}
		service.Mutex.RUnlock()
	}
	sm.mutex.RUnlock()

	if len(runningServices) == 0 {
		log.Printf("[INFO] %s - No running services to stop", time.Now().Format("2006-01-02 15:04:05"))
		return
	}

	// Sort services in reverse order for shutdown (stop in reverse startup order)
	sort.Slice(runningServices, func(i, j int) bool {
		return runningServices[i].Order > runningServices[j].Order
	})

	// Stop each service
	for _, service := range runningServices {
		log.Printf("[INFO] %s - Stopping service UUID: %s", time.Now().Format("2006-01-02 15:04:05"), service.ID)
		if err := sm.StopService(service.ID); err != nil {
			log.Printf("Failed to stop service UUID %s: %v", service.ID, err)
		} else {
			log.Printf("[INFO] %s - Successfully stopped service UUID: %s", time.Now().Format("2006-01-02 15:04:05"), service.ID)
		}
		// Small delay between stops to allow clean shutdown
		time.Sleep(500 * time.Millisecond)
	}

	// Final cleanup - force kill any remaining processes
	log.Printf("[INFO] %s - Performing final cleanup...", time.Now().Format("2006-01-02 15:04:05"))
	time.Sleep(2 * time.Second)

	for _, service := range runningServices {
		service.Mutex.Lock()
		if service.Cmd != nil && service.Cmd.Process != nil {
			log.Printf("[INFO] %s - Force cleaning up remaining process for UUID %s (PID: %d)", time.Now().Format("2006-01-02 15:04:05"), service.ID, service.PID)
			if pgid, err := GetProcessGroup(service.Cmd.Process.Pid); err == nil {
				ForceKillProcessGroup(pgid)
			} else {
				service.Cmd.Process.Kill()
			}
			service.Status = "stopped"
			service.HealthStatus = "unknown"
			service.PID = 0
			service.Cmd = nil
			service.Uptime = ""
			sm.updateServiceInDB(service)
		}
		service.Mutex.Unlock()
	}

	log.Printf("[INFO] %s - All services stopped successfully", time.Now().Format("2006-01-02 15:04:05"))
}

type GlobalConfigResponse struct {
	ProjectsDir      string `json:"projectsDir"`
	JavaHomeOverride string `json:"javaHomeOverride"`
	LastUpdated      string `json:"lastUpdated"`
}

func (sm *Manager) GetGlobalConfig() GlobalConfigResponse {
	return GlobalConfigResponse{
		ProjectsDir:      sm.config.ProjectsDir,
		JavaHomeOverride: sm.config.JavaHomeOverride,
		LastUpdated:      time.Now().Format(time.RFC3339),
	}
}

func (sm *Manager) UpdateGlobalConfig(projectsDir, javaHomeOverride string) (GlobalConfigResponse, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Update the configuration
	if projectsDir != "" {
		sm.config.ProjectsDir = projectsDir
	}
	sm.config.JavaHomeOverride = javaHomeOverride

	// Persist configuration to database
	if err := sm.saveGlobalConfigToDB(sm.config.ProjectsDir, sm.config.JavaHomeOverride); err != nil {
		return GlobalConfigResponse{}, fmt.Errorf("failed to persist global config: %w", err)
	}

	return GlobalConfigResponse{
		ProjectsDir:      sm.config.ProjectsDir,
		JavaHomeOverride: sm.config.JavaHomeOverride,
		LastUpdated:      time.Now().Format(time.RFC3339),
	}, nil
}

// SaveConfiguration saves a configuration to the database
func (sm *Manager) SaveConfiguration(config *models.Configuration) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Save to database
	if err := sm.saveConfigurationToDB(config); err != nil {
		return fmt.Errorf("failed to save configuration to database: %w", err)
	}

	// Update in-memory cache
	sm.configurations[config.ID] = config

	return nil
}

// UpdateConfiguration updates an existing configuration
func (sm *Manager) UpdateConfiguration(config *models.Configuration) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if configuration exists
	_, exists := sm.configurations[config.ID]
	if !exists {
		return fmt.Errorf("configuration %s not found", config.ID)
	}

	// Update database
	if err := sm.updateConfigurationInDB(config); err != nil {
		return fmt.Errorf("failed to update configuration in database: %w", err)
	}

	// Update in-memory cache
	sm.configurations[config.ID] = config

	return nil
}

// ApplyConfiguration applies a configuration by updating service orders
func (sm *Manager) ApplyConfiguration(configID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	config, exists := sm.configurations[configID]
	if !exists {
		return fmt.Errorf("configuration %s not found", configID)
	}

	// Update service orders based on configuration
	for _, configService := range config.Services {
		if service, exists := sm.services[configService.ID]; exists {
			service.Order = configService.Order
			sm.updateServiceInDB(service)
		}
	}

	// Mark this configuration as default and others as not default
	for _, cfg := range sm.configurations {
		cfg.IsDefault = (cfg.ID == configID)
	}

	// Update database to reflect the new default
	return sm.updateConfigurationDefaults(configID)
}

// UpdateService updates a service configuration
func (sm *Manager) UpdateService(serviceConfig *models.ServiceConfigRequest) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	service, exists := sm.services[serviceConfig.ID]
	if !exists {
		return fmt.Errorf("service UUID %s not found", serviceConfig.ID)
	}

	// Check for directory conflicts if directory is being changed
	if service.Dir != serviceConfig.Dir {
		if err := sm.ValidateServiceUniqueness(serviceConfig.ID, serviceConfig.Dir); err != nil {
			return err
		}
	}

	// Update service fields
	service.Name = serviceConfig.Name
	service.Dir = serviceConfig.Dir
	service.JavaOpts = serviceConfig.JavaOpts
	service.HealthURL = serviceConfig.HealthURL
	service.Port = serviceConfig.Port
	service.Order = serviceConfig.Order
	service.Description = serviceConfig.Description
	service.IsEnabled = serviceConfig.IsEnabled
	service.BuildSystem = serviceConfig.BuildSystem
	service.VerboseLogging = serviceConfig.VerboseLogging
	service.EnvVars = serviceConfig.EnvVars

	// Save to database
	if err := sm.UpdateServiceConfigInDB(service); err != nil {
		return fmt.Errorf("failed to update service in database: %w", err)
	}

	// Broadcast update
	sm.broadcastUpdate(service)

	return nil
}

// RenameService renames an existing service's name (not UUID)
func (sm *Manager) RenameService(serviceUUID, newName string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if service exists
	service, exists := sm.services[serviceUUID]
	if !exists {
		return fmt.Errorf("service UUID %s not found", serviceUUID)
	}

	// Check if new name is already taken within profiles that contain this service
	if err := sm.ValidateServiceNameUniquenessInProfile(serviceUUID, newName); err != nil {
		return err
	}

	// If names are the same, no rename needed
	if service.Name == newName {
		return nil
	}

	// Update the service name
	oldName := service.Name
	service.Name = newName

	// Update in database
	if err := sm.UpdateServiceConfigInDB(service); err != nil {
		// Revert the change if database update fails
		service.Name = oldName
		return fmt.Errorf("failed to rename service in database: %w", err)
	}

	// Broadcast update
	sm.broadcastUpdate(service)

	log.Printf("[INFO] Successfully renamed service UUID %s from %s to %s", serviceUUID, oldName, newName)
	return nil
}

// GetSystemResourceSummary returns overall system resource usage summary
func (sm *Manager) GetSystemResourceSummary() map[string]interface{} {
	return sm.getSystemResourceSummary()
}

// CleanupPort cleans up processes using the specified port
func (sm *Manager) CleanupPort(port int) *PortCleanupResult {
	return KillProcessesOnPort(port)
}

// ValidateServiceNameUniquenessInProfile checks if a service name is unique within the profiles it belongs to
// This replaces the global service name uniqueness validation with profile-scoped validation
func (sm *Manager) ValidateServiceNameUniquenessInProfile(serviceUUID, serviceName string) error {
	// Get all profiles that contain services
	profiles, err := sm.db.GetAllServiceProfiles()
	if err != nil {
		return fmt.Errorf("failed to get profiles for validation: %w", err)
	}

	// Check each profile for name conflicts
	for _, profile := range profiles {
		// Parse services JSON to get list of service UUIDs in this profile
		var serviceUUIDs []string
		if err := json.Unmarshal([]byte(profile.ServicesJSON), &serviceUUIDs); err != nil {
			log.Printf("[WARN] Failed to parse services JSON for profile %s: %v", profile.ID, err)
			continue
		}

		// Check if the service we're validating is in this profile
		serviceInProfile := false
		for _, uuid := range serviceUUIDs {
			if uuid == serviceUUID {
				serviceInProfile = true
				break
			}
		}

		// If service is in this profile, check for name conflicts with other services in the same profile
		if serviceInProfile {
			for _, uuid := range serviceUUIDs {
				if uuid == serviceUUID {
					continue // Skip self
				}
				
				// Get the other service
				if otherService, exists := sm.services[uuid]; exists {
					if otherService.Name == serviceName {
						return fmt.Errorf("service name '%s' already exists in profile '%s' (service UUID: %s)", 
							serviceName, profile.Name, uuid)
					}
				}
			}
		}
	}

	return nil
}

// ValidateServiceUniqueness checks if a service would conflict with existing services
// based on the combination of profile root directory and service directory
// Note: This method assumes the caller already holds the appropriate mutex lock
func (sm *Manager) ValidateServiceUniqueness(serviceUUID, serviceDir string) error {
	// Get the default projects directory (global)
	globalProjectsDir := sm.config.ProjectsDir

	// If global projects directory is empty, use current working directory
	if globalProjectsDir == "" {
		if cwd, err := os.Getwd(); err == nil {
			globalProjectsDir = cwd
		} else {
			globalProjectsDir = "."
		}
	}

	// Calculate the proposed service path using global projects directory
	proposedPath := filepath.Join(globalProjectsDir, serviceDir)
	proposedPath = filepath.Clean(proposedPath)

	// Check against all existing services (using direct map access to avoid mutex deadlock)
	for _, existing := range sm.services {
		// Skip self when updating
		if existing.ID == serviceUUID {
			continue
		}

		// Get the existing service's projects directory
		existingProjectsDir := sm.getServiceProjectsDirectory(existing.ID)
		if existingProjectsDir == "" {
			existingProjectsDir = globalProjectsDir
		}

		// Calculate existing service path
		existingPath := filepath.Join(existingProjectsDir, existing.Dir)
		existingPath = filepath.Clean(existingPath)

		// Check if paths would conflict
		if proposedPath == existingPath {
			return fmt.Errorf("service path conflict: UUID '%s' would use the same directory as existing service UUID '%s' (%s)", serviceUUID, existing.ID, existingPath)
		}
	}

	// Also check against all profiles to ensure no conflicts across profiles
	profiles, err := sm.db.GetAllServiceProfiles()
	if err != nil {
		log.Printf("[WARN] Failed to check profile conflicts: %v", err)
		// Continue without profile validation rather than failing
	} else {
		for _, profile := range profiles {
			// Skip if profile has no custom projects directory
			if profile.ProjectsDir == "" {
				continue
			}

			// Parse profile services (now storing UUIDs)
			var profileServiceUUIDs []string
			if err := json.Unmarshal([]byte(profile.ServicesJSON), &profileServiceUUIDs); err != nil {
				continue
			}

			// Check if any profile service would conflict
			for _, profileServiceUUID := range profileServiceUUIDs {
				if profileServiceUUID == serviceUUID {
					continue // Skip self
				}

				// Find the actual service details
				existingService, exists := sm.services[profileServiceUUID]
				if !exists {
					continue
				}

				// Calculate profile service path
				profileServicePath := filepath.Join(profile.ProjectsDir, existingService.Dir)
				profileServicePath = filepath.Clean(profileServicePath)

				if proposedPath == profileServicePath {
					return fmt.Errorf("service path conflict: UUID '%s' would use the same directory as service UUID '%s' in profile '%s' (%s)", serviceUUID, profileServiceUUID, profile.Name, profileServicePath)
				}
			}
		}
	}

	return nil
}

// getServiceProjectsDirectory returns the projects directory for a specific service
func (sm *Manager) getServiceProjectsDirectory(serviceUUID string) string {
	// Query database to find the profile that contains this service
	query := `SELECT projects_dir FROM service_profiles 
			  WHERE services_json LIKE ? AND projects_dir != '' AND projects_dir IS NOT NULL
			  ORDER BY is_active DESC, is_default DESC, created_at DESC
			  LIMIT 1`

	searchPattern := fmt.Sprintf("%%\"%s\"%%", serviceUUID)

	var projectsDir string
	err := sm.db.QueryRow(query, searchPattern).Scan(&projectsDir)
	if err != nil {
		// No profile found, return empty string to use global default
		return ""
	}

	return projectsDir
}

// AddService adds a new service to the manager
func (sm *Manager) AddService(service *models.Service) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if service with this UUID already exists
	if _, exists := sm.services[service.ID]; exists {
		return fmt.Errorf("service with UUID '%s' already exists", service.ID)
	}

	// Validate system-wide uniqueness based on directory path
	if err := sm.ValidateServiceUniqueness(service.ID, service.Dir); err != nil {
		return err
	}

	// Initialize service fields if not set
	if service.EnvVars == nil {
		service.EnvVars = make(map[string]models.EnvVar)
	}
	if service.Logs == nil {
		service.Logs = []models.LogEntry{}
	}
	if service.Status == "" {
		service.Status = "stopped"
	}
	if service.HealthStatus == "" {
		service.HealthStatus = "unknown"
	}

	// Add service to memory
	sm.services[service.ID] = service

	// Save to database (insert or update)
	if err := sm.upsertServiceInDB(service); err != nil {
		// Remove from memory if database save fails
		delete(sm.services, service.ID)
		return fmt.Errorf("failed to save service to database: %w", err)
	}

	// Broadcast the update
	sm.broadcastUpdate(service)

	log.Printf("[INFO] Successfully added service: %s (UUID: %s)", service.Name, service.ID)

	// Normalize orders to ensure sequential ordering
	go func() {
		if err := sm.NormalizeServiceOrders(); err != nil {
			log.Printf("[WARN] Failed to normalize service orders after adding UUID %s: %v", service.ID, err)
		}
	}()

	return nil
}

// DeleteService removes a service from the manager
func (sm *Manager) deleteService(serviceUUID string) error {
	// First, check if service exists and get its status
	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	isRunning := exists && service.Status == "running"
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service UUID '%s' not found", serviceUUID)
	}

	// Stop the service if it's running (without holding the main lock)
	if isRunning {
		log.Printf("[INFO] Stopping service UUID %s before deletion", serviceUUID)
		if err := sm.StopService(serviceUUID); err != nil {
			log.Printf("[WARN] Failed to stop service UUID %s before deletion: %v", serviceUUID, err)
			// Continue with deletion even if stop fails
		}
	}

	// Now acquire the write lock for deletion
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Double-check the service still exists (in case it was deleted by another goroutine)
	service, exists = sm.services[serviceUUID]
	if !exists {
		return fmt.Errorf("service UUID '%s' not found", serviceUUID)
	}

	// Remove from memory
	delete(sm.services, serviceUUID)

	// Remove from database
	if err := sm.db.DeleteService(serviceUUID); err != nil {
		// Re-add to memory if database deletion fails
		sm.services[serviceUUID] = service
		return fmt.Errorf("failed to delete service from database: %w", err)
	}

	log.Printf("[INFO] Successfully deleted service UUID: %s", serviceUUID)

	// Normalize orders to ensure sequential ordering
	go func() {
		if err := sm.NormalizeServiceOrders(); err != nil {
			log.Printf("[WARN] Failed to normalize service orders after deleting UUID %s: %v", serviceUUID, err)
		}
	}()

	return nil
}

// StartService starts a service by UUID
func (sm *Manager) StartService(serviceUUID string) error {
	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service UUID %s not found", serviceUUID)
	}

	// Check if service is enabled
	service.Mutex.RLock()
	isEnabled := service.IsEnabled
	service.Mutex.RUnlock()

	if !isEnabled {
		return fmt.Errorf("service %s is disabled and cannot be started", service.Name)
	}

	log.Printf("[INFO] Starting service UUID: %s", serviceUUID)

	return sm.startService(service)
}

// StopService stops a service by UUID
func (sm *Manager) StopService(serviceUUID string) error {
	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service UUID %s not found", serviceUUID)
	}

	log.Printf("[INFO] Stopping service UUID: %s", serviceUUID)

	return sm.stopService(service)
}

// RestartService restarts a service by UUID
func (sm *Manager) RestartService(serviceUUID string) error {
	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service UUID %s not found", serviceUUID)
	}

	// Check if service is enabled
	service.Mutex.RLock()
	isEnabled := service.IsEnabled
	service.Mutex.RUnlock()

	if !isEnabled {
		return fmt.Errorf("service %s is disabled and cannot be restarted", service.Name)
	}

	log.Printf("[INFO] Restarting service UUID: %s (port %d)", serviceUUID, service.Port)

	// Stop the service first
	if service.Status == "running" {
		if err := sm.stopService(service); err != nil {
			log.Printf("[WARN] Failed to stop service gracefully: %v", err)
			// Continue anyway - we'll clean up the port
		}
		// Wait a moment for cleanup
		time.Sleep(2 * time.Second)
	}

	// Clean up any processes still using the service's port
	if service.Port > 0 {
		log.Printf("[INFO] Cleaning up port %d before restarting service UUID %s", service.Port, serviceUUID)
		if err := CleanupPortBeforeStart(service.Port); err != nil {
			log.Printf("[WARN] Port cleanup failed: %v", err)
			// Continue anyway - the port might be available by now
		}
	}

	// Start the service
	err := sm.startService(service)

	// Record restart event if successful
	if err == nil {
		uptimeTracker := GetUptimeTracker()
		uptimeTracker.RecordEvent(service.ID, "restart", "running")
	}

	return err
}

// StartServiceWithProjectsDir starts a service using a specific projects directory
func (sm *Manager) StartServiceWithProjectsDir(serviceUUID, projectsDir string) error {
	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service UUID %s not found", serviceUUID)
	}

	// Check if service is enabled
	service.Mutex.RLock()
	isEnabled := service.IsEnabled
	service.Mutex.RUnlock()

	if !isEnabled {
		return fmt.Errorf("service %s is disabled and cannot be started", service.Name)
	}

	log.Printf("[INFO] Starting service UUID %s from projects directory: %s", serviceUUID, projectsDir)

	return sm.startServiceWithProjectsDir(service, projectsDir)
}

// RestartServiceWithProjectsDir restarts a service using a specific projects directory
func (sm *Manager) RestartServiceWithProjectsDir(serviceUUID, projectsDir string) error {
	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service UUID %s not found", serviceUUID)
	}

	// Check if service is enabled
	service.Mutex.RLock()
	isEnabled := service.IsEnabled
	service.Mutex.RUnlock()

	if !isEnabled {
		return fmt.Errorf("service %s is disabled and cannot be restarted", service.Name)
	}

	log.Printf("[INFO] Restarting service UUID %s from projects directory: %s (port %d)", serviceUUID, projectsDir, service.Port)

	// Stop the service first
	if service.Status == "running" {
		if err := sm.stopService(service); err != nil {
			log.Printf("[WARN] Failed to stop service gracefully: %v", err)
			// Continue anyway - we'll clean up the port
		}
		// Wait a moment for cleanup
		time.Sleep(2 * time.Second)
	}

	// Clean up any processes still using the service's port
	if service.Port > 0 {
		log.Printf("[INFO] Cleaning up port %d before restarting service UUID %s", service.Port, serviceUUID)
		if err := CleanupPortBeforeStart(service.Port); err != nil {
			log.Printf("[WARN] Port cleanup failed: %v", err)
			// Continue anyway - the port might be available by now
		}
	}

	// Start the service with custom projects directory
	return sm.startServiceWithProjectsDir(service, projectsDir)
}

// startLogCleanupRoutine starts a background routine that periodically cleans up old logs
func (sm *Manager) startLogCleanupRoutine() {
	// Run cleanup every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Run initial cleanup after 1 hour of startup
	initialDelay := time.NewTimer(1 * time.Hour)

	log.Printf("[INFO] Started periodic log cleanup routine (24-hour interval)")

	for {
		select {
		case <-initialDelay.C:
			// Run initial cleanup
			if err := sm.AutoCleanupLogs(); err != nil {
				log.Printf("[ERROR] Initial log cleanup failed: %v", err)
			}
		case <-ticker.C:
			// Run periodic cleanup
			if err := sm.AutoCleanupLogs(); err != nil {
				log.Printf("[ERROR] Periodic log cleanup failed: %v", err)
			}
		}
	}
}

// updateServiceInDB updates a service's status, health status, PID, last started time, and order in the database
func (sm *Manager) updateServiceInDB(service *models.Service) error {
	_, err := sm.db.Exec(`
		UPDATE services 
		SET status = ?, health_status = ?, pid = ?, last_started = ?, service_order = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		service.Status, service.HealthStatus, service.PID, service.LastStarted, service.Order, service.ID)
	if err != nil {
		return fmt.Errorf("failed to update service UUID %s in database: %w", service.ID, err)
	}
	return nil
}

// Wrapper management methods - delegates to buildsystem.go functions

// DetectBuildSystem detects the build system for a service directory
func (sm *Manager) DetectBuildSystem(serviceDir string) BuildSystemType {
	return DetectBuildSystem(serviceDir)
}

// ValidateWrapperIntegrity validates wrapper files for a service
func (sm *Manager) ValidateWrapperIntegrity(serviceDir string, buildSystem BuildSystemType) (bool, error) {
	return ValidateWrapperIntegrity(serviceDir, buildSystem)
}

// GenerateMavenWrapper generates Maven wrapper files
func (sm *Manager) GenerateMavenWrapper(serviceDir string) error {
	return GenerateMavenWrapper(serviceDir)
}

// GenerateGradleWrapper generates Gradle wrapper files
func (sm *Manager) GenerateGradleWrapper(serviceDir string) error {
	return GenerateGradleWrapper(serviceDir)
}

// RepairWrapper repairs wrapper files for a service
func (sm *Manager) RepairWrapper(serviceDir string) error {
	return RepairWrapper(serviceDir)
}

// HasMavenWrapper checks if Maven wrapper exists
func (sm *Manager) HasMavenWrapper(serviceDir string) bool {
	return HasMavenWrapper(serviceDir)
}

// HasGradleWrapper checks if Gradle wrapper exists
func (sm *Manager) HasGradleWrapper(serviceDir string) bool {
	return HasGradleWrapper(serviceDir)
}

// Git-related methods

// GetGitInfo returns git information for a service
func (sm *Manager) GetGitInfo(serviceUUID string) (*GitInfo, error) {
	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service UUID %s not found", serviceUUID)
	}

	return GetGitInfo(service.Dir)
}

// GetGitBranches returns all branches (local and remote) for a service
func (sm *Manager) GetGitBranches(serviceUUID string) ([]string, error) {
	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service UUID %s not found", serviceUUID)
	}

	if !IsGitRepository(service.Dir) {
		return nil, fmt.Errorf("service is not a git repository")
	}

	// Get local branches
	localBranches, err := GetBranches(service.Dir)
	if err != nil {
		return nil, err
	}

	// Get remote branches
	remoteBranches, err := GetRemoteBranches(service.Dir)
	if err != nil {
		// If remote fetch fails, just return local branches
		return localBranches, nil
	}

	// Combine and deduplicate
	branchMap := make(map[string]bool)
	for _, b := range localBranches {
		branchMap[b] = true
	}
	for _, b := range remoteBranches {
		branchMap[b] = true
	}

	branches := []string{}
	for b := range branchMap {
		branches = append(branches, b)
	}

	return branches, nil
}

// SwitchGitBranch switches a service to a different git branch
func (sm *Manager) SwitchGitBranch(serviceUUID, branch string) error {
	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service UUID %s not found", serviceUUID)
	}

	// Check if service is running
	if service.Status == "running" {
		return fmt.Errorf("cannot switch branches while service is running. Please stop the service first")
	}

	// Switch branch
	if err := SwitchBranch(service.Dir, branch); err != nil {
		return err
	}

	// Update the service's git branch info
	currentBranch, err := GetCurrentBranch(service.Dir)
	if err == nil {
		sm.mutex.Lock()
		service.GitBranch = currentBranch
		sm.mutex.Unlock()

		// Broadcast update
		sm.broadcastUpdate(service)
	}

	log.Printf("[INFO] Successfully switched service %s (UUID: %s) to branch %s", service.Name, serviceUUID, branch)
	return nil
}

// UpdateServiceGitBranch updates the git branch information for a service
func (sm *Manager) UpdateServiceGitBranch(serviceUUID string) error {
	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service UUID %s not found", serviceUUID)
	}

	if !IsGitRepository(service.Dir) {
		return nil
	}

	currentBranch, err := GetCurrentBranch(service.Dir)
	if err != nil {
		return err
	}

	sm.mutex.Lock()
	service.GitBranch = currentBranch
	sm.mutex.Unlock()

	return nil
}
