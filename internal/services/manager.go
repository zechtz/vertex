// Package services
package services

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zechtz/nest-up/internal/database"
	"github.com/zechtz/nest-up/internal/models"
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

func (sm *Manager) GetService(name string) (*models.Service, bool) {
	sm.mutex.RLock()
	service, exists := sm.services[name]
	sm.mutex.RUnlock()
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

func (sm *Manager) broadcastLogEntry(serviceName string, logEntry models.LogEntry) {
	sm.clientsMutex.Lock()
	defer sm.clientsMutex.Unlock()

	message := WebSocketMessage{
		Type: "log_entry",
		Payload: struct {
			ServiceName string           `json:"serviceName"`
			LogEntry    models.LogEntry `json:"logEntry"`
		}{
			ServiceName: serviceName,
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
		log.Printf("[INFO] %s - Stopping service: %s", time.Now().Format("2006-01-02 15:04:05"), service.Name)
		if err := sm.StopService(service.Name); err != nil {
			log.Printf("Failed to stop service %s: %v", service.Name, err)
		} else {
			log.Printf("[INFO] %s - Successfully stopped service: %s", time.Now().Format("2006-01-02 15:04:05"), service.Name)
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
			log.Printf("[INFO] %s - Force cleaning up remaining process for %s (PID: %d)", time.Now().Format("2006-01-02 15:04:05"), service.Name, service.PID)
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
		if service, exists := sm.services[configService.Name]; exists {
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
	
	service, exists := sm.services[serviceConfig.Name]
	if !exists {
		return fmt.Errorf("service %s not found", serviceConfig.Name)
	}
	
	// Update service fields
	service.JavaOpts = serviceConfig.JavaOpts
	service.HealthURL = serviceConfig.HealthURL
	service.Port = serviceConfig.Port
	service.Order = serviceConfig.Order
	service.Description = serviceConfig.Description
	service.IsEnabled = serviceConfig.IsEnabled
	service.EnvVars = serviceConfig.EnvVars
	
	// Save to database
	if err := sm.updateServiceInDB(service); err != nil {
		return fmt.Errorf("failed to update service in database: %w", err)
	}
	
	// Broadcast update
	sm.broadcastUpdate(service)
	
	return nil
}

// GetSystemResourceSummary returns overall system resource usage summary
func (sm *Manager) GetSystemResourceSummary() map[string]interface{} {
	return sm.getSystemResourceSummary()
}
