// Package handlers
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/zechtz/nest-up/internal/database"
	"github.com/zechtz/nest-up/internal/models"
	"github.com/zechtz/nest-up/internal/services"
)

type Handler struct {
	serviceManager       *services.Manager
	topologyService      *services.TopologyService
	autoDiscoveryService *services.AutoDiscoveryService
	authService          *services.AuthService
	profileService       *services.ProfileService
	upgrader             websocket.Upgrader
}

func NewHandler(sm *services.Manager) *Handler {
	return &Handler{
		serviceManager:       sm,
		topologyService:      services.NewTopologyService(sm),
		autoDiscoveryService: services.NewAutoDiscoveryService(sm),
		authService:          services.NewAuthService(sm.GetDatabase()),
		profileService:       services.NewProfileService(sm.GetDatabase(), sm),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// getServiceProjectsDir determines the appropriate projects directory for a service
// If the service belongs to an active profile with a custom projectsDir, use that
// Otherwise, use the global default
func (h *Handler) getServiceProjectsDir(serviceName string) string {
	// Get global config as fallback
	globalConfig := h.serviceManager.GetConfig()
	defaultProjectsDir := globalConfig.ProjectsDir
	
	// Query database to find all profiles that contain this service
	query := `SELECT projects_dir FROM service_profiles 
			  WHERE services_json LIKE ? AND projects_dir != '' AND projects_dir IS NOT NULL
			  ORDER BY is_active DESC, is_default DESC, created_at DESC
			  LIMIT 1`
	
	// Use LIKE to search for the service name in the JSON array
	searchPattern := fmt.Sprintf("%%\"%s\"%%", serviceName)
	
	var projectsDir string
	err := h.serviceManager.GetDatabase().QueryRow(query, searchPattern).Scan(&projectsDir)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("[WARN] Failed to query profile projects directory for service %s: %v", serviceName, err)
		}
		// No profile contains this service or query failed, use global default
		return defaultProjectsDir
	}
	
	if projectsDir == "" {
		// Profile exists but has empty projectsDir, use global default
		return defaultProjectsDir
	}
	
	log.Printf("[INFO] Using profile projects directory for service %s: %s", serviceName, projectsDir)
	return projectsDir
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	// Authentication routes (public)
	r.HandleFunc("/api/auth/register", h.registerHandler).Methods("POST")
	r.HandleFunc("/api/auth/login", h.loginHandler).Methods("POST")
	r.HandleFunc("/api/auth/user", h.getCurrentUserHandler).Methods("GET")
	
	// Profile Management routes (protected)
	r.HandleFunc("/api/user/profile", h.getUserProfileHandler).Methods("GET")
	r.HandleFunc("/api/user/profile", h.updateUserProfileHandler).Methods("PUT")
	r.HandleFunc("/api/profiles", h.getServiceProfilesHandler).Methods("GET")
	r.HandleFunc("/api/profiles", h.createServiceProfileHandler).Methods("POST")
	r.HandleFunc("/api/profiles/{id}", h.getServiceProfileHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}", h.updateServiceProfileHandler).Methods("PUT")
	r.HandleFunc("/api/profiles/{id}", h.deleteServiceProfileHandler).Methods("DELETE")
	r.HandleFunc("/api/profiles/{id}/apply", h.applyServiceProfileHandler).Methods("POST")
	
	// Profile-scoped configuration routes (protected)
	r.HandleFunc("/api/profiles/{id}/activate", h.setActiveProfileHandler).Methods("POST")
	r.HandleFunc("/api/profiles/active", h.getActiveProfileHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}/context", h.getProfileContextHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}/env-vars", h.getProfileEnvVarsHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}/env-vars", h.setProfileEnvVarHandler).Methods("POST")
	r.HandleFunc("/api/profiles/{id}/env-vars/{name}", h.deleteProfileEnvVarHandler).Methods("DELETE")
	r.HandleFunc("/api/profiles/{id}/service-configs/{service}", h.getProfileServiceConfigHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}/service-configs/{service}", h.setProfileServiceConfigHandler).Methods("POST")
	r.HandleFunc("/api/profiles/{id}/service-configs/{service}/{key}", h.deleteProfileServiceConfigHandler).Methods("DELETE")
	r.HandleFunc("/api/profiles/{id}/services", h.addServiceToProfileHandler).Methods("POST")
	r.HandleFunc("/api/profiles/{id}/services/{service}", h.removeServiceFromProfileHandler).Methods("DELETE")
	
	// Service routes (will be protected later)
	r.HandleFunc("/api/services", h.getServicesHandler).Methods("GET")
	r.HandleFunc("/api/services/{name}/start", h.startServiceHandler).Methods("POST")
	r.HandleFunc("/api/services/{name}/stop", h.stopServiceHandler).Methods("POST")
	r.HandleFunc("/api/services/{name}/restart", h.restartServiceHandler).Methods("POST")
	r.HandleFunc("/api/services/{name}/health", h.checkHealthHandler).Methods("POST")
	r.HandleFunc("/api/services/{name}/logs", h.getLogsHandler).Methods("GET")
	r.HandleFunc("/api/services/{name}/logs", h.clearLogsHandler).Methods("DELETE")
	r.HandleFunc("/api/services/{name}/port-cleanup", h.portCleanupHandler).Methods("POST")
	r.HandleFunc("/api/services/{name}/metrics", h.getServiceMetricsHandler).Methods("GET")
	r.HandleFunc("/api/logs/search", h.searchLogsHandler).Methods("POST")
	r.HandleFunc("/api/logs/statistics", h.getLogStatisticsHandler).Methods("GET")
	r.HandleFunc("/api/logs/export", h.exportLogsHandler).Methods("POST")
	r.HandleFunc("/api/services/start-all", h.startAllHandler).Methods("POST")
	r.HandleFunc("/api/services/start-all-profile", h.startAllProfileHandler).Methods("POST")
	r.HandleFunc("/api/services/stop-all", h.stopAllHandler).Methods("POST")
	r.HandleFunc("/api/services/stop-all-profile", h.stopAllProfileHandler).Methods("POST")
	r.HandleFunc("/api/system/metrics", h.getSystemMetricsHandler).Methods("GET")
	r.HandleFunc("/api/system/logs/cleanup", h.cleanupLogsHandler).Methods("POST")
	r.HandleFunc("/api/services/fix-lombok", h.fixLombokHandler).Methods("POST")
	r.HandleFunc("/api/environment/setup", h.setupEnvironmentHandler).Methods("POST")
	r.HandleFunc("/api/environment/sync", h.syncEnvironmentHandler).Methods("POST")
	r.HandleFunc("/api/environment/status", h.getEnvironmentStatusHandler).Methods("GET")
	r.HandleFunc("/api/configurations", h.getConfigurationsHandler).Methods("GET")
	r.HandleFunc("/api/configurations", h.saveConfigurationHandler).Methods("POST")
	r.HandleFunc("/api/configurations/{id}", h.updateConfigurationHandler).Methods("PUT")
	r.HandleFunc("/api/configurations/{id}/apply", h.applyConfigurationHandler).Methods("POST")
	r.HandleFunc("/api/services", h.createServiceHandler).Methods("POST")
	r.HandleFunc("/api/services/{name}", h.updateServiceHandler).Methods("PUT")
	r.HandleFunc("/api/services/{name}", h.deleteServiceHandler).Methods("DELETE")
	r.HandleFunc("/api/services/{name}/env-vars", h.getServiceEnvVarsHandler).Methods("GET")
	r.HandleFunc("/api/services/{name}/env-vars", h.updateServiceEnvVarsHandler).Methods("PUT")
	r.HandleFunc("/api/env-vars/global", h.getGlobalEnvVarsHandler).Methods("GET")
	r.HandleFunc("/api/env-vars/global", h.updateGlobalEnvVarsHandler).Methods("PUT")
	r.HandleFunc("/api/env-vars/reload", h.reloadEnvVarsHandler).Methods("POST")
	r.HandleFunc("/api/env-vars/cleanup", h.cleanupGlobalEnvVarsHandler).Methods("POST")
	r.HandleFunc("/api/services/{name}/files", h.getServiceFilesHandler).Methods("GET")
	r.HandleFunc("/api/services/{name}/files/{filename}", h.updateServiceFileHandler).Methods("PUT")
	r.HandleFunc("/api/config/global", h.getGlobalConfigHandler).Methods("GET")
	r.HandleFunc("/api/config/global", h.updateGlobalConfigHandler).Methods("PUT")
	r.HandleFunc("/api/topology", h.getTopologyHandler).Methods("GET")
	r.HandleFunc("/api/topology/debug", h.getTopologyDebugHandler).Methods("GET")
	r.HandleFunc("/api/dependencies", h.getDependenciesHandler).Methods("GET")
	r.HandleFunc("/api/dependencies", h.saveDependenciesHandler).Methods("POST")
	r.HandleFunc("/api/dependencies/graph", h.getDependencyGraphHandler).Methods("GET")
	r.HandleFunc("/api/dependencies/validate", h.validateDependenciesHandler).Methods("GET")
	r.HandleFunc("/api/dependencies/startup-order", h.getStartupOrderHandler).Methods("POST")
	r.HandleFunc("/api/eureka/services", h.getEurekaServicesHandler).Methods("GET")
	r.HandleFunc("/api/eureka/debug", h.debugEurekaHandler).Methods("GET")
	r.HandleFunc("/api/services/{name}/gitlab-ci", h.getGitLabCIHandler).Methods("GET")
	r.HandleFunc("/api/services/{name}/install-libraries", h.installLibrariesHandler).Methods("POST")
	r.HandleFunc("/api/services/gitlab-ci/all", h.getAllGitLabCIHandler).Methods("GET")
	r.HandleFunc("/api/auto-discovery/scan", h.scanAutoDiscoveryHandler).Methods("POST")
	r.HandleFunc("/api/auto-discovery/services", h.getDiscoveredServicesHandler).Methods("GET")
	r.HandleFunc("/api/auto-discovery/import", h.importDiscoveredServiceHandler).Methods("POST")
	r.HandleFunc("/ws", h.websocketHandler)
}

func (h *Handler) getServicesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	services := h.serviceManager.GetServices()
	json.NewEncoder(w).Encode(services)
}

func (h *Handler) startServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get the appropriate projects directory for this service
	projectsDir := h.getServiceProjectsDir(serviceName)
	
	// Use profile-aware service startup if a custom projects directory is found
	globalConfig := h.serviceManager.GetConfig()
	if projectsDir != globalConfig.ProjectsDir {
		// Service belongs to a profile with custom projects directory
		if err := h.serviceManager.StartServiceWithProjectsDir(serviceName, projectsDir); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Use normal startup (global projects directory)
		if err := h.serviceManager.StartService(serviceName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (h *Handler) stopServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := h.serviceManager.StopService(serviceName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (h *Handler) startAllHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := h.serviceManager.StartAllServices(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "starting all services"})
}

func (h *Handler) startAllProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get claims from JWT token to identify user
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's active profile
	profile, err := h.profileService.GetActiveProfile(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get active profile for start all: %v", err)
		// Fall back to global start all if no active profile
		if err := h.serviceManager.StartAllServices(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "starting all services (global)"})
		return
	}

	// Use profile's projects directory if available, otherwise use global
	projectsDir := h.serviceManager.GetConfig().ProjectsDir
	if profile.ProjectsDir != "" {
		projectsDir = profile.ProjectsDir
	}

	// Convert profile services to JSON string
	servicesJSON, err := json.Marshal(profile.Services)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to serialize profile services: %v", err), http.StatusInternalServerError)
		return
	}

	// Start only services in the active profile
	if err := h.serviceManager.StartAllServicesForProfile(string(servicesJSON), projectsDir); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": fmt.Sprintf("starting all services in profile '%s'", profile.Name),
		"profile": profile.Name,
	})
}

func (h *Handler) stopAllHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := h.serviceManager.StopAllServices(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "stopping all services"})
}

func (h *Handler) stopAllProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get claims from JWT token to identify user
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's active profile
	profile, err := h.profileService.GetActiveProfile(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get active profile for stop all: %v", err)
		// Fall back to global stop all if no active profile
		if err := h.serviceManager.StopAllServices(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "stopping all services (global)"})
		return
	}

	// Convert profile services to JSON string
	servicesJSON, err := json.Marshal(profile.Services)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to serialize profile services: %v", err), http.StatusInternalServerError)
		return
	}

	// Stop only services in the active profile
	if err := h.serviceManager.StopAllServicesForProfile(string(servicesJSON)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": fmt.Sprintf("stopping all services in profile '%s'", profile.Name),
		"profile": profile.Name,
	})
}

func (h *Handler) restartServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get the appropriate projects directory for this service
	projectsDir := h.getServiceProjectsDir(serviceName)
	
	// Use profile-aware service restart if a custom projects directory is found
	globalConfig := h.serviceManager.GetConfig()
	if projectsDir != globalConfig.ProjectsDir {
		// Service belongs to a profile with custom projects directory
		if err := h.serviceManager.RestartServiceWithProjectsDir(serviceName, projectsDir); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Use normal restart (global projects directory)
		if err := h.serviceManager.RestartService(serviceName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "restarted"})
}

func (h *Handler) portCleanupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get the service to find its port
	service, exists := h.serviceManager.GetService(serviceName)
	if !exists {
		http.Error(w, fmt.Sprintf("Service '%s' not found", serviceName), http.StatusNotFound)
		return
	}

	if service.Port <= 0 {
		http.Error(w, "Service does not have a valid port configured", http.StatusBadRequest)
		return
	}

	// Perform port cleanup
	result := h.serviceManager.CleanupPort(service.Port)
	
	// Return detailed cleanup result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "port cleanup completed",
		"port": result.Port,
		"processesFound": result.ProcessesFound,
		"processesKilled": result.ProcessesKilled,
		"pids": result.PIDs,
		"errors": result.Errors,
	})
}

func (h *Handler) checkHealthHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := h.serviceManager.CheckServiceHealth(serviceName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "health check triggered"})
}

func (h *Handler) getLogsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	service, exists := h.serviceManager.GetService(serviceName)
	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	service.Mutex.RLock()
	logs := service.Logs
	service.Mutex.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{"logs": logs})
}

func (h *Handler) clearLogsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := h.serviceManager.ClearLogs(serviceName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "logs cleared"})
}

func (h *Handler) getConfigurationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	configs := h.serviceManager.GetConfigurations()
	json.NewEncoder(w).Encode(configs)
}

func (h *Handler) saveConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var config models.Configuration
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if config.ID == "" {
		config.ID = fmt.Sprintf("config_%d", time.Now().Unix())
	}

	if err := h.serviceManager.SaveConfiguration(&config); err != nil {
		log.Printf("Failed to save configuration: %v", err)
		http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"id":     config.ID,
	})
}

func (h *Handler) updateConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	configID := vars["id"]

	if configID == "" {
		http.Error(w, "Configuration ID is required", http.StatusBadRequest)
		return
	}

	var config models.Configuration
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure the ID matches the URL parameter
	config.ID = configID

	if err := h.serviceManager.UpdateConfiguration(&config); err != nil {
		log.Printf("Failed to update configuration: %v", err)
		http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"id":     config.ID,
	})
}

func (h *Handler) applyConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	configID := vars["id"]

	if configID == "" {
		http.Error(w, "Configuration ID is required", http.StatusBadRequest)
		return
	}

	if err := h.serviceManager.ApplyConfiguration(configID); err != nil {
		log.Printf("Failed to apply configuration: %v", err)
		http.Error(w, "Failed to apply configuration", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) createServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var service models.Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if service.Name == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	if service.Dir == "" {
		http.Error(w, "Service directory is required", http.StatusBadRequest)
		return
	}

	if service.Port == 0 {
		service.Port = 8080 // Default port
	}

	// Set default values
	if service.BuildSystem == "" {
		service.BuildSystem = "auto"
	}
	if service.Status == "" {
		service.Status = "stopped"
	}
	if service.HealthStatus == "" {
		service.HealthStatus = "unknown"
	}
	
	// Initialize time fields properly
	service.LastStarted = time.Time{} // Zero time for new services
	service.Uptime = ""
	service.PID = 0
	
	// Initialize maps if nil
	if service.EnvVars == nil {
		service.EnvVars = make(map[string]models.EnvVar)
	}

	log.Printf("[INFO] Creating new service: %s", service.Name)

	if err := h.serviceManager.AddService(&service); err != nil {
		log.Printf("Failed to create service: %v", err)
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, "Service with this name already exists", http.StatusConflict)
		} else {
			http.Error(w, "Failed to create service", http.StatusInternalServerError)
		}
		return
	}

	// Return the created service
	if err := json.NewEncoder(w).Encode(service); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) updateServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	serviceName := vars["name"]

	if serviceName == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	var serviceConfig models.ServiceConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&serviceConfig); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure the service name matches the URL parameter
	serviceConfig.Name = serviceName

	if err := h.serviceManager.UpdateService(&serviceConfig); err != nil {
		log.Printf("Failed to update service: %v", err)
		http.Error(w, "Failed to update service", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) deleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	if serviceName == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Delete service request for: %s", serviceName)

	// Delete the service
	if err := h.serviceManager.DeleteService(serviceName); err != nil {
		log.Printf("[ERROR] Failed to delete service %s: %v", serviceName, err)
		http.Error(w, fmt.Sprintf("Failed to delete service: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Service '%s' deleted successfully", serviceName),
	})
}

func (h *Handler) getServiceEnvVarsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	envVars, err := h.serviceManager.GetServiceEnvVars(serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"envVars": envVars})
}

func (h *Handler) updateServiceEnvVarsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var request struct {
		EnvVars map[string]models.EnvVar `json:"envVars"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.serviceManager.UpdateServiceEnvVars(serviceName, request.EnvVars); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *Handler) getGlobalEnvVarsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	envVars, err := h.serviceManager.GetGlobalEnvVars()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"envVars": envVars})
}

func (h *Handler) updateGlobalEnvVarsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var request struct {
		EnvVars map[string]string `json:"envVars"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.serviceManager.UpdateGlobalEnvVars(request.EnvVars); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *Handler) reloadEnvVarsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Reload environment variables from fish file
	if err := h.serviceManager.ReloadEnvVarsFromFishFile(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "reloaded"})
}

func (h *Handler) cleanupGlobalEnvVarsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Clean up and reload environment variables with only simplified ones
	if err := h.serviceManager.CleanupAndReloadGlobalEnvVars(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "cleaned up and reloaded"})
}

func (h *Handler) getServiceFilesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get the appropriate projects directory for this service (profile-aware)
	projectsDir := h.getServiceProjectsDir(serviceName)

	files, err := h.serviceManager.GetServiceFilesWithProjectsDir(serviceName, projectsDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"files": files})
}

func (h *Handler) updateServiceFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]
	filename := vars["filename"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var request struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.serviceManager.UpdateServiceFile(serviceName, filename, request.Content); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *Handler) getGlobalConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	config := h.serviceManager.GetGlobalConfig()
	json.NewEncoder(w).Encode(config)
}

func (h *Handler) updateGlobalConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var request struct {
		ProjectsDir      string `json:"projectsDir"`
		JavaHomeOverride string `json:"javaHomeOverride"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	config, err := h.serviceManager.UpdateGlobalConfig(request.ProjectsDir, request.JavaHomeOverride)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(config)
}

func (h *Handler) fixLombokHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("[INFO] Starting Lombok compatibility check for all services")
	
	// Get all services
	services := h.serviceManager.GetServices()
	results := make(map[string]string)
	
	for _, service := range services {
		serviceDir := fmt.Sprintf("%s/%s", h.serviceManager.GetConfig().ProjectsDir, service.Dir)
		
		if err := h.serviceManager.CheckAndFixLombokCompatibility(serviceDir, service.Name); err != nil {
			results[service.Name] = fmt.Sprintf("Error: %v", err)
			log.Printf("[ERROR] Lombok fix failed for service %s: %v", service.Name, err)
		} else {
			results[service.Name] = "Success"
			log.Printf("[INFO] Lombok compatibility checked for service %s", service.Name)
		}
	}
	
	response := map[string]interface{}{
		"message": "Lombok compatibility check completed",
		"results": results,
	}
	
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) setupEnvironmentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("[INFO] Setting up environment variables")

	// Get working directory
	workingDir, err := os.Getwd()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get working directory: %v", err), http.StatusInternalServerError)
		return
	}

	// Create environment setup instance with database
	envSetup := services.NewEnvironmentSetup(workingDir, h.serviceManager.GetDatabase())
	
	// Setup environment
	result := envSetup.SetupEnvironment()
	
	if !result.Success {
		w.WriteHeader(http.StatusPartialContent)
	}
	
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) syncEnvironmentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("[INFO] Syncing environment variables from database")

	// Get working directory
	workingDir, err := os.Getwd()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get working directory: %v", err), http.StatusInternalServerError)
		return
	}

	// Create environment setup instance with database
	envSetup := services.NewEnvironmentSetup(workingDir, h.serviceManager.GetDatabase())
	
	// Setup environment (now loads from database)
	result := envSetup.SetupEnvironment()
	
	if !result.Success {
		w.WriteHeader(http.StatusPartialContent)
	}
	
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) getEnvironmentStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get working directory
	workingDir, err := os.Getwd()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get working directory: %v", err), http.StatusInternalServerError)
		return
	}

	// Create environment setup instance with database
	envSetup := services.NewEnvironmentSetup(workingDir, h.serviceManager.GetDatabase())
	
	// Get environment status
	status := envSetup.CheckEnvironmentStatus()
	currentEnv := envSetup.GetCurrentEnvironment()
	
	response := map[string]interface{}{
		"status":      status,
		"environment": currentEnv,
	}
	
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) getServiceMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	serviceName := vars["name"]

	service, exists := h.serviceManager.GetService(serviceName)
	if !exists {
		http.Error(w, fmt.Sprintf("Service '%s' not found", serviceName), http.StatusNotFound)
		return
	}

	service.Mutex.RLock()
	metrics := map[string]interface{}{
		"serviceName":    service.Name,
		"cpuPercent":     service.CPUPercent,
		"memoryUsage":    service.MemoryUsage,
		"memoryPercent":  service.MemoryPercent,
		"diskUsage":      service.DiskUsage,
		"networkRx":      service.NetworkRx,
		"networkTx":      service.NetworkTx,
		"metrics":        service.Metrics,
		"status":         service.Status,
		"healthStatus":   service.HealthStatus,
		"pid":            service.PID,
		"uptime":         service.Uptime,
		"lastStarted":    service.LastStarted,
		"timestamp":      time.Now(),
	}
	service.Mutex.RUnlock()

	json.NewEncoder(w).Encode(metrics)
}

func (h *Handler) getSystemMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get system resource summary
	summary := h.serviceManager.GetSystemResourceSummary()
	
	// Add individual service metrics
	services := h.serviceManager.GetServices()
	serviceMetrics := make([]map[string]interface{}, 0)
	
	for _, service := range services {
		if service.Status == "running" {
			serviceMetric := map[string]interface{}{
				"name":           service.Name,
				"cpuPercent":     service.CPUPercent,
				"memoryUsage":    service.MemoryUsage,
				"memoryPercent":  service.MemoryPercent,
				"diskUsage":      service.DiskUsage,
				"networkRx":      service.NetworkRx,
				"networkTx":      service.NetworkTx,
				"status":         service.Status,
				"healthStatus":   service.HealthStatus,
				"uptime":         service.Uptime,
				"errorRate":      service.Metrics.ErrorRate,
				"requestCount":   service.Metrics.RequestCount,
			}
			serviceMetrics = append(serviceMetrics, serviceMetric)
		}
	}
	
	response := map[string]interface{}{
		"summary":  summary,
		"services": serviceMetrics,
	}
	
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) cleanupLogsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Parse optional parameters
	var request struct {
		MaxDays           int `json:"maxDays"`
		MaxLogsPerService int `json:"maxLogsPerService"`
	}

	// Set defaults
	request.MaxDays = 7
	request.MaxLogsPerService = 1000

	// Try to parse request body for custom parameters
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Printf("[WARN] Failed to parse log cleanup request: %v", err)
			// Continue with defaults
		}
	}

	// Validate parameters
	if request.MaxDays < 1 || request.MaxDays > 365 {
		request.MaxDays = 7
	}
	if request.MaxLogsPerService < 100 || request.MaxLogsPerService > 10000 {
		request.MaxLogsPerService = 1000
	}

	// Perform cleanup
	err := h.serviceManager.CleanupOldLogs(request.MaxDays, request.MaxLogsPerService)
	if err != nil {
		log.Printf("[ERROR] Log cleanup failed: %v", err)
		http.Error(w, fmt.Sprintf("Log cleanup failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":           true,
		"message":           fmt.Sprintf("Log cleanup completed - kept logs from last %d days with max %d logs per service", request.MaxDays, request.MaxLogsPerService),
		"maxDays":           request.MaxDays,
		"maxLogsPerService": request.MaxLogsPerService,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) searchLogsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var criteria struct {
		ServiceNames []string `json:"serviceNames"`
		Levels       []string `json:"levels"`
		SearchText   string   `json:"searchText"`
		StartTime    string   `json:"startTime"`
		EndTime      string   `json:"endTime"`
		Limit        int      `json:"limit"`
		Offset       int      `json:"offset"`
	}

	if err := json.NewDecoder(r.Body).Decode(&criteria); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Parse time strings
	var startTime, endTime time.Time
	var err error

	if criteria.StartTime != "" {
		startTime, err = time.Parse(time.RFC3339, criteria.StartTime)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid start time format: %v", err), http.StatusBadRequest)
			return
		}
	}

	if criteria.EndTime != "" {
		endTime, err = time.Parse(time.RFC3339, criteria.EndTime)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid end time format: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Set default limit if not provided
	if criteria.Limit <= 0 {
		criteria.Limit = 100
	}

	// Create database search criteria
	searchCriteria := database.LogSearchCriteria{
		ServiceNames: criteria.ServiceNames,
		Levels:       criteria.Levels,
		SearchText:   criteria.SearchText,
		StartTime:    startTime,
		EndTime:      endTime,
		Limit:        criteria.Limit,
		Offset:       criteria.Offset,
	}

	// Perform search
	results, totalCount, err := h.serviceManager.GetDatabase().SearchLogs(searchCriteria)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to search logs: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"results":    results,
		"totalCount": totalCount,
		"limit":      criteria.Limit,
		"offset":     criteria.Offset,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) getLogStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	stats, err := h.serviceManager.GetDatabase().GetLogStatistics()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get log statistics: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) exportLogsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var exportRequest struct {
		ServiceNames []string `json:"serviceNames"`
		Levels       []string `json:"levels"`
		SearchText   string   `json:"searchText"`
		StartTime    string   `json:"startTime"`
		EndTime      string   `json:"endTime"`
		Format       string   `json:"format"` // "json", "csv", "txt"
	}

	if err := json.NewDecoder(r.Body).Decode(&exportRequest); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Parse time strings
	var startTime, endTime time.Time
	var err error

	if exportRequest.StartTime != "" {
		startTime, err = time.Parse(time.RFC3339, exportRequest.StartTime)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid start time format: %v", err), http.StatusBadRequest)
			return
		}
	}

	if exportRequest.EndTime != "" {
		endTime, err = time.Parse(time.RFC3339, exportRequest.EndTime)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid end time format: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Create search criteria for export (no limit)
	searchCriteria := database.LogSearchCriteria{
		ServiceNames: exportRequest.ServiceNames,
		Levels:       exportRequest.Levels,
		SearchText:   exportRequest.SearchText,
		StartTime:    startTime,
		EndTime:      endTime,
		Limit:        0, // No limit for export
		Offset:       0,
	}

	// Get logs for export
	results, _, err := h.serviceManager.GetDatabase().SearchLogs(searchCriteria)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to search logs for export: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("nest_logs_%s", timestamp)

	// Handle different export formats
	switch exportRequest.Format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.json\"", filename))
		json.NewEncoder(w).Encode(results)

	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", filename))
		
		w.Write([]byte("Timestamp,Service,Level,Message\n"))
		for _, result := range results {
			// Escape CSV values
			message := strings.ReplaceAll(result.Message, "\"", "\"\"")
			if strings.Contains(message, ",") || strings.Contains(message, "\n") {
				message = "\"" + message + "\""
			}
			
			line := fmt.Sprintf("%s,%s,%s,%s\n",
				result.Timestamp.Format(time.RFC3339),
				result.ServiceName,
				result.Level,
				message,
			)
			w.Write([]byte(line))
		}

	case "txt":
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.txt\"", filename))
		
		for _, result := range results {
			line := fmt.Sprintf("[%s] [%s] [%s] %s\n",
				result.Timestamp.Format("2006-01-02 15:04:05"),
				result.ServiceName,
				result.Level,
				result.Message,
			)
			w.Write([]byte(line))
		}

	default:
		http.Error(w, "Invalid export format. Supported formats: json, csv, txt", http.StatusBadRequest)
		return
	}
}

func (h *Handler) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to websocket", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	h.serviceManager.AddWebSocketClient(conn)
	defer h.serviceManager.RemoveWebSocketClient(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// getTopologyHandler returns the service topology visualization data
func (h *Handler) getTopologyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get claims from JWT token to identify user
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		// For backward compatibility, fall back to global topology if not authenticated
		topology, err := h.topologyService.GenerateTopology()
		if err != nil {
			log.Printf("Failed to generate topology: %v", err)
			http.Error(w, "Failed to generate topology", http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(topology); err != nil {
			log.Printf("Failed to encode topology: %v", err)
			http.Error(w, "Failed to encode topology", http.StatusInternalServerError)
			return
		}
		return
	}

	// Get user's active profile
	profile, err := h.profileService.GetActiveProfile(claims.UserID)
	if err != nil {
		log.Printf("[INFO] No active profile found for topology, using global view: %v", err)
		// Fall back to global topology if no active profile
		topology, err := h.topologyService.GenerateTopology()
		if err != nil {
			log.Printf("Failed to generate topology: %v", err)
			http.Error(w, "Failed to generate topology", http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(topology); err != nil {
			log.Printf("Failed to encode topology: %v", err)
			http.Error(w, "Failed to encode topology", http.StatusInternalServerError)
			return
		}
		return
	}

	// Convert profile services to JSON string
	servicesJSON, err := json.Marshal(profile.Services)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to serialize profile services: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate topology for the active profile
	topology, err := h.topologyService.GenerateTopologyForProfile(string(servicesJSON))
	if err != nil {
		log.Printf("Failed to generate profile topology: %v", err)
		http.Error(w, "Failed to generate topology", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(topology); err != nil {
		log.Printf("Failed to encode topology: %v", err)
		http.Error(w, "Failed to encode topology", http.StatusInternalServerError)
		return
	}
}

// getTopologyDebugHandler returns debug information about topology generation
func (h *Handler) getTopologyDebugHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	debugInfo := make(map[string]interface{})
	debugInfo["authHeader"] = r.Header.Get("Authorization")

	// Try to get claims
	claims, ok := extractClaimsFromRequest(r, h.authService)
	debugInfo["hasValidToken"] = ok
	
	if ok {
		debugInfo["userID"] = claims.UserID
		
		// Try to get active profile
		profile, err := h.profileService.GetActiveProfile(claims.UserID)
		debugInfo["profileError"] = err
		
		if err == nil {
			debugInfo["profileName"] = profile.Name
			debugInfo["profileServices"] = profile.Services
			debugInfo["profileServicesCount"] = len(profile.Services)
		}
	}

	json.NewEncoder(w).Encode(debugInfo)
}

// getDependenciesHandler returns service dependencies information
func (h *Handler) getDependenciesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	services := h.serviceManager.GetServices()
	dependencies := make(map[string]interface{})

	// Load dependencies from database
	db := h.serviceManager.GetDatabase()
	allDependencies, err := db.GetAllServiceDependencies()
	if err != nil {
		log.Printf("Failed to load dependencies from database: %v", err)
		http.Error(w, "Failed to load dependencies", http.StatusInternalServerError)
		return
	}

	for _, service := range services {
		serviceData := map[string]interface{}{
			"order": service.Order,
		}

		// Add dependencies from database if they exist
		if dbDeps, exists := allDependencies[service.Name]; exists {
			serviceData["dependencies"] = dbDeps
		} else {
			serviceData["dependencies"] = []interface{}{}
		}

		// Add dependent on info (reverse dependencies)
		serviceData["dependentOn"] = service.DependentOn
		serviceData["startupDelay"] = service.StartupDelay.String()

		dependencies[service.Name] = serviceData
	}

	if err := json.NewEncoder(w).Encode(dependencies); err != nil {
		log.Printf("Failed to encode dependencies: %v", err)
		http.Error(w, "Failed to encode dependencies", http.StatusInternalServerError)
		return
	}
}

// saveDependenciesHandler saves service dependencies configuration
func (h *Handler) saveDependenciesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var configData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&configData); err != nil {
		log.Printf("Failed to decode dependencies config: %v", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Update service dependencies and orders
	services := h.serviceManager.GetServices()
	serviceMap := make(map[string]*models.Service)
	for i := range services {
		serviceMap[services[i].Name] = &services[i]
	}

	// Process each service's configuration
	for serviceName, config := range configData {
		if configMap, ok := config.(map[string]interface{}); ok {
			service := serviceMap[serviceName]
			if service == nil {
				log.Printf("Service %s not found, skipping", serviceName)
				continue
			}

			// Update service order
			if order, exists := configMap["order"]; exists {
				if orderFloat, ok := order.(float64); ok {
					service.Order = int(orderFloat)
					log.Printf("Updated order for %s to %d", serviceName, service.Order)
				}
			}

			// Update dependencies in database
			if dependencies, exists := configMap["dependencies"]; exists {
				if depsList, ok := dependencies.([]interface{}); ok {
					// Save dependencies to database
					db := h.serviceManager.GetDatabase()
					if err := db.SaveServiceDependencies(serviceName, depsList); err != nil {
						log.Printf("Failed to save dependencies for %s: %v", serviceName, err)
						http.Error(w, fmt.Sprintf("Failed to save dependencies for %s", serviceName), http.StatusInternalServerError)
						return
					}
					log.Printf("Saved %d dependencies for %s", len(depsList), serviceName)
				}
			}

			// Update the service in the service manager
			if err := h.serviceManager.UpdateServiceInDB(service); err != nil {
				log.Printf("Failed to update service %s in database: %v", serviceName, err)
			}
		}
	}

	log.Printf("Dependencies configuration saved successfully")

	response := map[string]interface{}{
		"status":  "success",
		"message": "Dependencies configuration saved successfully",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getDependencyGraphHandler returns the complete dependency graph
func (h *Handler) getDependencyGraphHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Access dependency manager through service manager
	// Note: We need to add a getter method to access the dependency manager
	services := h.serviceManager.GetServices()
	graph := make(map[string]interface{})

	for _, service := range services {
		if len(service.Dependencies) > 0 {
			graph[service.Name] = service.Dependencies
		}
	}

	result := map[string]interface{}{
		"dependencies": graph,
		"generated":    time.Now(),
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode dependency graph: %v", err)
		http.Error(w, "Failed to encode dependency graph", http.StatusInternalServerError)
		return
	}
}

// validateDependenciesHandler validates the dependency configuration
func (h *Handler) validateDependenciesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// For now, we'll do basic validation
	services := h.serviceManager.GetServices()
	serviceNames := make(map[string]bool)
	for _, service := range services {
		serviceNames[service.Name] = true
	}

	errors := []string{}
	warnings := []string{}

	// Check for missing dependencies
	for _, service := range services {
		for _, dep := range service.Dependencies {
			if !serviceNames[dep.ServiceName] {
				errors = append(errors, fmt.Sprintf("Service %s depends on non-existent service %s", service.Name, dep.ServiceName))
			}
		}
	}

	// Basic cycle detection (simplified)
	// In a real implementation, this would use the dependency manager's validation

	result := map[string]interface{}{
		"valid":    len(errors) == 0,
		"errors":   errors,
		"warnings": warnings,
		"checked":  time.Now(),
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode validation result: %v", err)
		http.Error(w, "Failed to encode validation result", http.StatusInternalServerError)
		return
	}
}

// getStartupOrderHandler returns the optimal startup order for given services
func (h *Handler) getStartupOrderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var request struct {
		Services []string `json:"services"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// If no services specified, use all services
	if len(request.Services) == 0 {
		services := h.serviceManager.GetServices()
		for _, service := range services {
			request.Services = append(request.Services, service.Name)
		}
	}

	// Simple ordering based on service Order field for now
	// In a real implementation, this would use the dependency manager
	services := h.serviceManager.GetServices()
	serviceMap := make(map[string]models.Service)
	for _, service := range services {
		serviceMap[service.Name] = service
	}

	var orderedServices []models.Service
	for _, name := range request.Services {
		if service, exists := serviceMap[name]; exists {
			orderedServices = append(orderedServices, service)
		}
	}

	sort.Slice(orderedServices, func(i, j int) bool {
		return orderedServices[i].Order < orderedServices[j].Order
	})

	var orderedNames []string
	for _, service := range orderedServices {
		orderedNames = append(orderedNames, service.Name)
	}

	result := map[string]interface{}{
		"startupOrder": orderedNames,
		"services":     len(orderedNames),
		"generated":    time.Now(),
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode startup order: %v", err)
		http.Error(w, "Failed to encode startup order", http.StatusInternalServerError)
		return
	}
}

// getEurekaServicesHandler returns the status of services from Eureka registry
func (h *Handler) getEurekaServicesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get service status from Eureka
	// We need to add a method to access the service manager's Eureka functionality
	services := h.serviceManager.GetServices()
	result := make(map[string]interface{})

	// For each service, check its registration status with Eureka
	for _, service := range services {
		// Skip Eureka itself
		serviceName := strings.ToUpper(service.Name)
		if serviceName == "EUREKA" || serviceName == "NEST-REGISTRY-SERVER" {
			result[service.Name] = map[string]interface{}{
				"status":      service.HealthStatus,
				"source":      "direct",
				"port":        service.Port,
				"description": "Registry server (not self-registered)",
			}
			continue
		}

		// For other services, we'll show their current health status
		// In a future implementation, this could query Eureka directly
		result[service.Name] = map[string]interface{}{
			"status":      service.HealthStatus,
			"source":      "eureka-aware",
			"port":        service.Port,
			"description": "Health checked via Eureka registry",
		}
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode Eureka services: %v", err)
		http.Error(w, "Failed to encode Eureka services", http.StatusInternalServerError)
		return
	}
}

// debugEurekaHandler provides debug information about Eureka integration
func (h *Handler) debugEurekaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Try to fetch raw Eureka data
	eurekaURL := "http://localhost:8800/eureka/apps"
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", eurekaURL, nil)
	if err != nil {
		result := map[string]interface{}{
			"error": "Failed to create Eureka request",
			"details": err.Error(),
		}
		json.NewEncoder(w).Encode(result)
		return
	}

	req.Header.Set("Accept", "application/xml") // Request XML since that's what your Eureka returns
	resp, err := client.Do(req)
	if err != nil {
		result := map[string]interface{}{
			"error": "Failed to query Eureka",
			"details": err.Error(),
			"eureka_url": eurekaURL,
		}
		json.NewEncoder(w).Encode(result)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		result := map[string]interface{}{
			"error": "Eureka returned non-200 status",
			"status_code": resp.StatusCode,
			"eureka_url": eurekaURL,
		}
		json.NewEncoder(w).Encode(result)
		return
	}

	// Parse the response
	var eurekaData interface{}
	if err := json.NewDecoder(resp.Body).Decode(&eurekaData); err != nil {
		result := map[string]interface{}{
			"error": "Failed to decode Eureka response",
			"details": err.Error(),
		}
		json.NewEncoder(w).Encode(result)
		return
	}

	// Get our services for comparison
	services := h.serviceManager.GetServices()
	serviceNames := make([]string, len(services))
	for i, service := range services {
		serviceNames[i] = service.Name
	}

	result := map[string]interface{}{
		"success": true,
		"eureka_url": eurekaURL,
		"eureka_status": resp.StatusCode,
		"eureka_data": eurekaData,
		"local_services": serviceNames,
		"message": "Raw Eureka data and local services for debugging",
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode debug response: %v", err)
		http.Error(w, "Failed to encode debug response", http.StatusInternalServerError)
		return
	}
}

// getGitLabCIHandler returns GitLab CI configuration for a specific service
func (h *Handler) getGitLabCIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	config, err := h.serviceManager.ParseGitLabCI(serviceName)
	if err != nil {
		log.Printf("Failed to parse GitLab CI for service %s: %v", serviceName, err)
		http.Error(w, fmt.Sprintf("Failed to parse GitLab CI: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(config); err != nil {
		log.Printf("Failed to encode GitLab CI config: %v", err)
		http.Error(w, "Failed to encode GitLab CI config", http.StatusInternalServerError)
		return
	}
}

// getAllGitLabCIHandler returns GitLab CI configurations for all services
func (h *Handler) getAllGitLabCIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	configs := h.serviceManager.GetAllGitLabCIConfigs()

	if err := json.NewEncoder(w).Encode(configs); err != nil {
		log.Printf("Failed to encode GitLab CI configs: %v", err)
		http.Error(w, "Failed to encode GitLab CI configs", http.StatusInternalServerError)
		return
	}
}

// installLibrariesHandler installs Maven libraries for a specific service
func (h *Handler) installLibrariesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("[INFO] Installing libraries for service: %s", serviceName)

	// Install libraries in a goroutine to avoid timeout
	go func() {
		if err := h.serviceManager.InstallLibraries(serviceName); err != nil {
			log.Printf("[ERROR] Failed to install libraries for service %s: %v", serviceName, err)
		} else {
			log.Printf("[INFO] Successfully installed libraries for service %s", serviceName)
		}
	}()

	// Return immediate response
	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Library installation started for service %s. Check logs for progress.", serviceName),
		"service": serviceName,
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode install libraries response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// scanAutoDiscoveryHandler triggers a scan of the project directory for services
func (h *Handler) scanAutoDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's active profile to determine scan directory
	profile, err := h.profileService.GetActiveProfile(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get active profile: %v", err)
		// Fall back to global directory if no active profile
		log.Printf("[INFO] Starting auto-discovery scan (global directory)")
		discoveredServices, err := h.autoDiscoveryService.ScanProjectDirectory()
		if err != nil {
			log.Printf("[ERROR] Auto-discovery scan failed: %v", err)
			http.Error(w, fmt.Sprintf("Failed to scan project directory: %v", err), http.StatusInternalServerError)
			return
		}
		h.sendAutoDiscoveryResponse(w, discoveredServices)
		return
	}

	// Use profile-specific directory if available, otherwise use global
	scanDir := h.serviceManager.GetConfig().ProjectsDir
	if profile.ProjectsDir != "" {
		scanDir = profile.ProjectsDir
	}

	log.Printf("[INFO] Starting auto-discovery scan in profile directory: %s", scanDir)

	discoveredServices, err := h.autoDiscoveryService.ScanDirectory(scanDir)
	if err != nil {
		log.Printf("[ERROR] Auto-discovery scan failed: %v", err)
		http.Error(w, fmt.Sprintf("Failed to scan project directory: %v", err), http.StatusInternalServerError)
		return
	}

	h.sendAutoDiscoveryResponse(w, discoveredServices)
}

// sendAutoDiscoveryResponse sends the auto-discovery scan results
func (h *Handler) sendAutoDiscoveryResponse(w http.ResponseWriter, discoveredServices []services.DiscoveredService) {
	result := map[string]interface{}{
		"success":           true,
		"message":           fmt.Sprintf("Found %d potential services", len(discoveredServices)),
		"discoveredServices": discoveredServices,
		"totalFound":        len(discoveredServices),
	}

	log.Printf("[INFO] Auto-discovery scan completed. Found %d services", len(discoveredServices))

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode auto-discovery response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getDiscoveredServicesHandler returns the last discovered services (if any)
func (h *Handler) getDiscoveredServicesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// For now, we'll trigger a fresh scan each time
	// In a more advanced implementation, we could cache results
	discoveredServices, err := h.autoDiscoveryService.ScanProjectDirectory()
	if err != nil {
		log.Printf("[ERROR] Failed to get discovered services: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get discovered services: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(discoveredServices); err != nil {
		log.Printf("Failed to encode discovered services: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// importDiscoveredServiceHandler imports a discovered service into the system
func (h *Handler) importDiscoveredServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var discoveredService services.DiscoveredService
	if err := json.NewDecoder(r.Body).Decode(&discoveredService); err != nil {
		log.Printf("[ERROR] Failed to decode discovered service: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Importing discovered service: %s from %s", discoveredService.Name, discoveredService.Path)

	service, err := h.autoDiscoveryService.CreateServiceFromDiscovered(discoveredService)
	if err != nil {
		log.Printf("[ERROR] Failed to import discovered service: %v", err)
		http.Error(w, fmt.Sprintf("Failed to import service: %v", err), http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully imported service '%s'", service.Name),
		"service": service,
	}

	log.Printf("[INFO] Successfully imported service: %s", service.Name)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode import service response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Authentication handlers

// registerHandler handles user registration
func (h *Handler) registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var registration models.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&registration); err != nil {
		log.Printf("[ERROR] Failed to decode registration request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if registration.Email == "" || registration.Username == "" || registration.Password == "" {
		http.Error(w, "Email, username, and password are required", http.StatusBadRequest)
		return
	}

	if len(registration.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters long", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(&registration)
	if err != nil {
		log.Printf("[ERROR] Failed to register user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] User registered successfully: %s (%s)", user.Username, user.Email)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User registered successfully",
		"user":    user,
	}); err != nil {
		log.Printf("Failed to encode registration response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// loginHandler handles user login
func (h *Handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var login models.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		log.Printf("[ERROR] Failed to decode login request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if login.Email == "" || login.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	authResponse, err := h.authService.Login(&login)
	if err != nil {
		log.Printf("[ERROR] Failed to login user: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	log.Printf("[INFO] User logged in successfully: %s", authResponse.User.Username)

	if err := json.NewEncoder(w).Encode(authResponse); err != nil {
		log.Printf("Failed to encode login response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getCurrentUserHandler returns the current user info based on JWT token
func (h *Handler) getCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Extract token (format: "Bearer <token>")
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Validate token
	claims, err := h.authService.ValidateToken(tokenParts[1])
	if err != nil {
		log.Printf("[ERROR] Failed to validate token: %v", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Get user details
	user, err := h.authService.GetUserByID(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get user: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Failed to encode user response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Helper function to extract JWT claims from request
func extractClaimsFromRequest(r *http.Request, authService *services.AuthService) (*models.JWTClaims, bool) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, false
	}

	// Extract token (format: "Bearer <token>")
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return nil, false
	}

	// Validate token
	claims, err := authService.ValidateToken(tokenParts[1])
	if err != nil {
		return nil, false
	}

	return claims, true
}

// Profile Management Handlers

// getUserProfileHandler retrieves the current user's profile
func (h *Handler) getUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := h.profileService.GetUserProfile(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get user profile: %v", err)
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Printf("Failed to encode profile response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// updateUserProfileHandler updates the current user's profile
func (h *Handler) updateUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.UserProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	profile, err := h.profileService.UpdateUserProfile(claims.UserID, &req)
	if err != nil {
		log.Printf("[ERROR] Failed to update user profile: %v", err)
		http.Error(w, "Failed to update user profile", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Printf("Failed to encode profile response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getServiceProfilesHandler retrieves all service profiles for the current user
func (h *Handler) getServiceProfilesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profiles, err := h.profileService.GetServiceProfiles(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get service profiles: %v", err)
		http.Error(w, "Failed to get service profiles", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(profiles); err != nil {
		log.Printf("Failed to encode profiles response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getServiceProfileHandler retrieves a specific service profile
func (h *Handler) getServiceProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	profile, err := h.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get service profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get service profile", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Printf("Failed to encode profile response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// createServiceProfileHandler creates a new service profile
func (h *Handler) createServiceProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[DEBUG] Create profile request: %+v", req)

	profile, err := h.profileService.CreateServiceProfile(claims.UserID, &req)
	if err != nil {
		log.Printf("[ERROR] Failed to create service profile: %v", err)
		if strings.Contains(err.Error(), "invalid services") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to create service profile", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Printf("Failed to encode profile response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// updateServiceProfileHandler updates an existing service profile
func (h *Handler) updateServiceProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[DEBUG] Update profile request for ID %s: %+v", profileID, req)

	profile, err := h.profileService.UpdateServiceProfile(profileID, claims.UserID, &req)
	if err != nil {
		log.Printf("[ERROR] Failed to update service profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "invalid services") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to update service profile", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Printf("Failed to encode profile response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// deleteServiceProfileHandler deletes a service profile
func (h *Handler) deleteServiceProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	err := h.profileService.DeleteServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to delete service profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete service profile", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// applyServiceProfileHandler applies a service profile
func (h *Handler) applyServiceProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	err := h.profileService.ApplyProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to apply service profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to apply service profile", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{
		"message": "Profile applied successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Profile-scoped configuration handlers

// setActiveProfileHandler sets a profile as the active profile for the user
func (h *Handler) setActiveProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	err := h.profileService.SetActiveProfile(claims.UserID, profileID)
	if err != nil {
		log.Printf("[ERROR] Failed to set active profile: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to set active profile", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{
		"message": "Active profile set successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getActiveProfileHandler gets the active profile for the user
func (h *Handler) getActiveProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := h.profileService.GetActiveProfile(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get active profile: %v", err)
		http.Error(w, "Failed to get active profile", http.StatusInternalServerError)
		return
	}

	if profile == nil {
		// No active profile - return empty response
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getProfileContextHandler gets the complete configuration context for a profile
func (h *Handler) getProfileContextHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	context, err := h.profileService.GetProfileContext(claims.UserID, profileID)
	if err != nil {
		log.Printf("[ERROR] Failed to get profile context: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get profile context", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(context); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getProfileEnvVarsHandler gets all environment variables for a profile
func (h *Handler) getProfileEnvVarsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	envVars, err := h.profileService.GetProfileEnvVars(claims.UserID, profileID)
	if err != nil {
		log.Printf("[ERROR] Failed to get profile env vars: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get profile env vars", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(envVars); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// setProfileEnvVarHandler sets an environment variable for a profile
func (h *Handler) setProfileEnvVarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	var request struct {
		Name        string `json:"name"`
		Value       string `json:"value"`
		Description string `json:"description"`
		IsRequired  bool   `json:"isRequired"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Name == "" {
		http.Error(w, "Environment variable name is required", http.StatusBadRequest)
		return
	}

	err := h.profileService.SetProfileEnvVar(claims.UserID, profileID, request.Name, request.Value, request.Description, request.IsRequired)
	if err != nil {
		log.Printf("[ERROR] Failed to set profile env var: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to set profile env var", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{
		"message": "Environment variable set successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// deleteProfileEnvVarHandler deletes an environment variable for a profile
func (h *Handler) deleteProfileEnvVarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	name := vars["name"]
	
	if profileID == "" || name == "" {
		http.Error(w, "Profile ID and variable name are required", http.StatusBadRequest)
		return
	}

	err := h.profileService.DeleteProfileEnvVar(claims.UserID, profileID, name)
	if err != nil {
		log.Printf("[ERROR] Failed to delete profile env var: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile or variable not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete profile env var", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// addServiceToProfileHandler adds a service to a profile
func (h *Handler) addServiceToProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	var request struct {
		ServiceName string `json:"serviceName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.ServiceName == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Adding service '%s' to profile '%s' for user '%s'", request.ServiceName, profileID, claims.UserID)

	err := h.profileService.AddServiceToProfile(claims.UserID, profileID, request.ServiceName)
	if err != nil {
		log.Printf("[ERROR] Failed to add service to profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile or service not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "already exists") {
			http.Error(w, "Service already exists in profile", http.StatusConflict)
		} else {
			http.Error(w, "Failed to add service to profile", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{
		"message": "Service added to profile successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// removeServiceFromProfileHandler removes a service from a profile (doesn't delete globally)
func (h *Handler) removeServiceFromProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	serviceName := vars["service"]
	
	if profileID == "" || serviceName == "" {
		http.Error(w, "Profile ID and service name are required", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Removing service '%s' from profile '%s' for user '%s'", serviceName, profileID, claims.UserID)

	err := h.profileService.RemoveServiceFromProfile(claims.UserID, profileID, serviceName)
	if err != nil {
		log.Printf("[ERROR] Failed to remove service from profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile or service not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to remove service from profile", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Service '%s' removed from profile successfully", serviceName),
	})
}

// getProfileServiceConfigHandler gets service configuration for a profile
func (h *Handler) getProfileServiceConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	serviceName := vars["service"]
	
	if profileID == "" || serviceName == "" {
		http.Error(w, "Profile ID and service name are required", http.StatusBadRequest)
		return
	}

	// Verify profile exists and user has access
	_, err := h.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to verify profile access: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to verify profile access", http.StatusInternalServerError)
		}
		return
	}

	config, err := h.profileService.GetDatabase().GetProfileServiceConfig(profileID, serviceName)
	if err != nil {
		log.Printf("[ERROR] Failed to get profile service config: %v", err)
		http.Error(w, "Failed to get service config", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(config); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// setProfileServiceConfigHandler sets service configuration for a profile
func (h *Handler) setProfileServiceConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	serviceName := vars["service"]
	
	if profileID == "" || serviceName == "" {
		http.Error(w, "Profile ID and service name are required", http.StatusBadRequest)
		return
	}

	var request struct {
		Key         string `json:"key"`
		Value       string `json:"value"`
		ConfigType  string `json:"configType"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Key == "" {
		http.Error(w, "Configuration key is required", http.StatusBadRequest)
		return
	}

	if request.ConfigType == "" {
		request.ConfigType = "string"
	}

	// Verify profile exists and user has access
	_, err := h.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to verify profile access: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to verify profile access", http.StatusInternalServerError)
		}
		return
	}

	err = h.profileService.GetDatabase().SetProfileServiceConfig(profileID, serviceName, request.Key, request.Value, request.ConfigType, request.Description)
	if err != nil {
		log.Printf("[ERROR] Failed to set profile service config: %v", err)
		http.Error(w, "Failed to set service config", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Service configuration set successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// deleteProfileServiceConfigHandler deletes service configuration for a profile
func (h *Handler) deleteProfileServiceConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Extract user from JWT token
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	serviceName := vars["service"]
	key := vars["key"]
	
	if profileID == "" || serviceName == "" || key == "" {
		http.Error(w, "Profile ID, service name, and config key are required", http.StatusBadRequest)
		return
	}

	// Verify profile exists and user has access
	_, err := h.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to verify profile access: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to verify profile access", http.StatusInternalServerError)
		}
		return
	}

	err = h.profileService.GetDatabase().DeleteProfileServiceConfig(profileID, serviceName, key)
	if err != nil {
		log.Printf("[ERROR] Failed to delete profile service config: %v", err)
		http.Error(w, "Failed to delete service config", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
