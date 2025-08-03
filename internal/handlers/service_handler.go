package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zechtz/vertex/internal/models"
)

func registerServiceRoutes(h *Handler, r *mux.Router) {
	// Service CRUD operations (RESTful with UUIDs)
	r.HandleFunc("/api/services", h.getServicesHandler).Methods("GET")
	r.HandleFunc("/api/services", h.createServiceHandler).Methods("POST")
	r.HandleFunc("/api/services/{id}", h.getServiceHandler).Methods("GET")
	r.HandleFunc("/api/services/{id}", h.updateServiceHandler).Methods("PUT")
	r.HandleFunc("/api/services/{id}", h.deleteServiceHandler).Methods("DELETE")

	// Service operations (by UUID)
	r.HandleFunc("/api/services/{id}/start", h.startServiceHandler).Methods("POST")
	r.HandleFunc("/api/services/{id}/stop", h.stopServiceHandler).Methods("POST")
	r.HandleFunc("/api/services/{id}/restart", h.restartServiceHandler).Methods("POST")
	r.HandleFunc("/api/services/{id}/health", h.checkHealthHandler).Methods("POST")
	r.HandleFunc("/api/services/{id}/env-vars", h.getServiceEnvVarsHandler).Methods("GET")
	r.HandleFunc("/api/services/{id}/env-vars", h.updateServiceEnvVarsHandler).Methods("PUT")
	r.HandleFunc("/api/services/{id}/install-libraries", h.installLibrariesHandler).Methods("POST")
	r.HandleFunc("/api/services/{id}/libraries/preview", h.previewLibrariesHandler).Methods("GET")
	r.HandleFunc("/api/services/{id}/libraries/install", h.installSelectedLibrariesHandler).Methods("POST")
	r.HandleFunc("/api/services/{id}/files", h.getServiceFilesHandler).Methods("GET")
	r.HandleFunc("/api/services/{id}/files/{filename}", h.updateServiceFileHandler).Methods("PUT")

	r.HandleFunc("/api/services/start-all", h.startAllHandler).Methods("POST")
	r.HandleFunc("/api/services/stop-all", h.stopAllHandler).Methods("POST")
	r.HandleFunc("/api/services/{id}/port-cleanup", h.portCleanupHandler).Methods("POST")
	r.HandleFunc("/api/services/{id}/logs", h.getLogsHandler).Methods("GET")
	r.HandleFunc("/api/services/{id}/logs", h.clearLogsHandler).Methods("DELETE")
	r.HandleFunc("/api/services/logs/clear", h.clearAllLogsHandler).Methods("DELETE")
	r.HandleFunc("/api/services/{id}/metrics", h.getServiceMetricsHandler).Methods("GET")

	// Wrapper management endpoints
	r.HandleFunc("/api/services/{id}/wrapper/validate", h.validateWrapperHandler).Methods("GET")
	r.HandleFunc("/api/services/{id}/wrapper/generate", h.generateWrapperHandler).Methods("POST")
	r.HandleFunc("/api/services/{id}/wrapper/repair", h.repairWrapperHandler).Methods("POST")

	// Utility endpoints
	r.HandleFunc("/api/services/available-for-profile", h.getAvailableServicesForProfileHandler).Methods("GET")
	r.HandleFunc("/api/services/normalize-order", h.normalizeServiceOrderHandler).Methods("POST")
}

func (h *Handler) getServicesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	services := h.serviceManager.GetServices()
	json.NewEncoder(w).Encode(services)
}

func (h *Handler) getAvailableServicesForProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok || claims == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	excludeProfileID := r.URL.Query().Get("excludeProfile")

	allServices := h.serviceManager.GetServices()

	userProfiles, err := h.profileService.GetServiceProfiles(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get user profiles: %v", err)
		http.Error(w, "Failed to get user profiles", http.StatusInternalServerError)
		return
	}

	assignedServices := make(map[string]bool)
	for _, profile := range userProfiles {
		if excludeProfileID != "" && profile.ID == excludeProfileID {
			continue
		}
		for _, serviceUUID := range profile.Services {
			assignedServices[serviceUUID] = true
		}
	}

	var availableServices []models.Service
	for _, service := range allServices {
		if !assignedServices[service.ID] {
			availableServices = append(availableServices, service)
		}
	}

	log.Printf("[DEBUG] Total services: %d, Assigned services (excluding profile %s): %d, Available services: %d",
		len(allServices), excludeProfileID, len(assignedServices), len(availableServices))

	json.NewEncoder(w).Encode(availableServices)
}

func (h *Handler) startServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	_, exists := h.serviceManager.GetServiceByUUID(serviceUUID)
	if !exists {
		http.Error(w, fmt.Sprintf("Service with UUID %s not found", serviceUUID), http.StatusNotFound)
		return
	}

	projectsDir := h.getServiceProjectsDir(serviceUUID)
	globalConfig := h.serviceManager.GetConfig()
	if projectsDir != globalConfig.ProjectsDir {
		if err := h.serviceManager.StartServiceWithProjectsDir(serviceUUID, projectsDir); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.serviceManager.StartService(serviceUUID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (h *Handler) stopServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := h.serviceManager.StopService(serviceUUID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (h *Handler) restartServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	projectsDir := h.getServiceProjectsDir(serviceUUID)
	globalConfig := h.serviceManager.GetConfig()
	if projectsDir != globalConfig.ProjectsDir {
		if err := h.serviceManager.RestartServiceWithProjectsDir(serviceUUID, projectsDir); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.serviceManager.RestartService(serviceUUID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "restarted"})
}

func (h *Handler) checkHealthHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := h.serviceManager.CheckServiceHealth(serviceUUID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "health check triggered"})
}

func (h *Handler) createServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var service models.Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		log.Printf("[ERROR] Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if service.Name == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	if service.Dir == "" {
		http.Error(w, "Service directory is required", http.StatusBadRequest)
		return
	}

	// Generate UUID if not provided
	if service.ID == "" {
		service.ID = uuid.New().String()
	}

	if service.Port == 0 {
		service.Port = 8080
	}

	if service.BuildSystem == "" {
		service.BuildSystem = "auto"
	}
	if service.Status == "" {
		service.Status = "stopped"
	}
	if service.HealthStatus == "" {
		service.HealthStatus = "unknown"
	}

	if service.EnvVars == nil {
		service.EnvVars = make(map[string]models.EnvVar)
	}

	log.Printf("[INFO] Creating new service: %s (UUID: %s)", service.Name, service.ID)

	if err := h.serviceManager.AddService(&service); err != nil {
		log.Printf("[ERROR] Failed to create service: %v", err)
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, "Service with this UUID or path already exists", http.StatusConflict)
		} else {
			http.Error(w, "Failed to create service", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(service); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) getServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	service, exists := h.serviceManager.GetServiceByUUID(serviceUUID)
	if !exists {
		log.Printf("[ERROR] Service with UUID %s not found", serviceUUID)
		http.Error(w, fmt.Sprintf("Service with UUID %s not found", serviceUUID), http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(service); err != nil {
		log.Printf("[ERROR] Failed to encode service response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) updateServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	var serviceConfig models.ServiceConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&serviceConfig); err != nil {
		log.Printf("[ERROR] Failed to decode service config: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[DEBUG] Received service config for UUID %s: %+v", serviceUUID, serviceConfig)

	if serviceConfig.ID != "" && serviceConfig.ID != serviceUUID {
		log.Printf("[INFO] Renaming service UUID %s to %s", serviceUUID, serviceConfig.ID)
		if err := h.serviceManager.RenameService(serviceUUID, serviceConfig.ID); err != nil {
			log.Printf("[ERROR] Failed to rename service UUID %s to %s: %v", serviceUUID, serviceConfig.ID, err)
			http.Error(w, fmt.Sprintf("Failed to rename service: %v", err), http.StatusInternalServerError)
			return
		}
		serviceUUID = serviceConfig.ID
	} else {
		serviceConfig.ID = serviceUUID
	}

	if err := h.serviceManager.UpdateService(&serviceConfig); err != nil {
		log.Printf("[ERROR] Failed to update service UUID %s: %v", serviceUUID, err)
		http.Error(w, fmt.Sprintf("Failed to update service: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) deleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Delete service request for UUID: %s", serviceUUID)

	if err := h.serviceManager.DeleteService(serviceUUID); err != nil {
		log.Printf("[ERROR] Failed to delete service UUID %s: %v", serviceUUID, err)
		http.Error(w, fmt.Sprintf("Failed to delete service: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Service with UUID '%s' deleted successfully", serviceUUID),
	})
}

func (h *Handler) normalizeServiceOrderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("[INFO] Normalizing service orders")

	if err := h.serviceManager.NormalizeServiceOrders(); err != nil {
		log.Printf("[ERROR] Failed to normalize service orders: %v", err)
		http.Error(w, fmt.Sprintf("Failed to normalize service orders: %v", err), http.StatusInternalServerError)
		return
	}

	services := h.serviceManager.GetServices()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"message":  fmt.Sprintf("Successfully normalized orders for %d services", len(services)),
		"services": services,
	})
}

func (h *Handler) getServiceEnvVarsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	envVars, err := h.serviceManager.GetServiceEnvVars(serviceUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"envVars": envVars})
}

func (h *Handler) updateServiceEnvVarsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var request struct {
		EnvVars map[string]models.EnvVar `json:"envVars"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.serviceManager.UpdateServiceEnvVars(serviceUUID, request.EnvVars); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *Handler) installLibrariesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check if service exists
	_, exists := h.serviceManager.GetServiceByUUID(serviceUUID)
	if !exists {
		http.Error(w, fmt.Sprintf("Service with UUID %s not found", serviceUUID), http.StatusNotFound)
		return
	}

	// Get the correct projects directory using profile-aware logic
	projectsDir := h.getServiceProjectsDir(serviceUUID)
	
	log.Printf("[INFO] Installing libraries for service %s (auto-discovery from .gitlab-ci.yml) using projects dir: %s", serviceUUID, projectsDir)

	// Call InstallLibrariesWithProjectsDir to use the correct directory
	if err := h.serviceManager.InstallLibrariesWithProjectsDir(serviceUUID, []models.LibraryInstallation{}, projectsDir); err != nil {
		log.Printf("[ERROR] Failed to install libraries for service UUID %s: %v", serviceUUID, err)
		http.Error(w, fmt.Sprintf("Failed to install libraries: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Successfully installed libraries for service %s", serviceUUID),
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) startAllHandler(w http.ResponseWriter, r *http.Request) {
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
		"status":  fmt.Sprintf("starting all services in profile '%s'", profile.Name),
		"profile": profile.Name,
	})
}

func (h *Handler) stopAllHandler(w http.ResponseWriter, r *http.Request) {
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
		"status":  fmt.Sprintf("stopping all services in profile '%s'", profile.Name),
		"profile": profile.Name,
	})
}

func (h *Handler) portCleanupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get the service to find its port
	service, exists := h.serviceManager.GetServiceByUUID(serviceUUID)
	if !exists {
		http.Error(w, fmt.Sprintf("Service '%s' not found", serviceUUID), http.StatusNotFound)
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
		"status":          "port cleanup completed",
		"port":            result.Port,
		"processesFound":  result.ProcessesFound,
		"processesKilled": result.ProcessesKilled,
		"pids":            result.PIDs,
		"errors":          result.Errors,
	})
}

func (h *Handler) getLogsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	service, exists := h.serviceManager.GetServiceByUUID(serviceUUID)
	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	service.Mutex.RLock()
	logs := service.Logs
	service.Mutex.RUnlock()

	json.NewEncoder(w).Encode(map[string]any{"logs": logs})
}

func (h *Handler) clearLogsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	service, exists := h.serviceManager.GetServiceByUUID(serviceUUID)
	if !exists {
		http.Error(w, fmt.Sprintf("Service '%s' not found", serviceUUID), http.StatusNotFound)
		return
	}

	if err := h.serviceManager.ClearLogs(serviceUUID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "logs cleared",
		"service": map[string]string{"id": serviceUUID, "name": service.Name},
	})
}

func (h *Handler) clearAllLogsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var request struct {
		ServiceNames []string `json:"serviceNames,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// If no body or invalid JSON, clear all logs
		request.ServiceNames = []string{}
	}

	results := h.serviceManager.ClearAllLogs(request.ServiceNames)
	
	successCount := 0
	errorCount := 0
	for _, result := range results {
		if result == "Success" {
			successCount++
		} else {
			errorCount++
		}
	}

	response := map[string]interface{}{
		"status":       "completed",
		"results":      results,
		"successCount": successCount,
		"errorCount":   errorCount,
	}

	if len(request.ServiceNames) == 0 {
		response["message"] = fmt.Sprintf("Cleared logs for all %d services", successCount)
	} else {
		response["message"] = fmt.Sprintf("Cleared logs for %d of %d specified services", successCount, len(request.ServiceNames))
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) getServiceMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	service, exists := h.serviceManager.GetServiceByUUID(serviceUUID)
	if !exists {
		http.Error(w, fmt.Sprintf("Service '%s' not found", serviceUUID), http.StatusNotFound)
		return
	}

	service.Mutex.RLock()
	metrics := map[string]interface{}{
		"serviceName":   service.Name,
		"cpuPercent":    service.CPUPercent,
		"memoryUsage":   service.MemoryUsage,
		"memoryPercent": service.MemoryPercent,
		"diskUsage":     service.DiskUsage,
		"networkRx":     service.NetworkRx,
		"networkTx":     service.NetworkTx,
		"metrics":       service.Metrics,
		"status":        service.Status,
		"healthStatus":  service.HealthStatus,
		"pid":           service.PID,
		"uptime":        service.Uptime,
		"lastStarted":   service.LastStarted,
		"timestamp":     time.Now(),
	}
	service.Mutex.RUnlock()

	json.NewEncoder(w).Encode(metrics)
}

func (h *Handler) getServiceFilesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Try to extract user from JWT token to get user-specific profile (optional)
	claims, ok := extractClaimsFromRequest(r, h.authService)
	var projectsDir string
	if ok && claims != nil {
		// User is authenticated, use profile-aware directory lookup
		projectsDir = h.getServiceProjectsDirForUser(serviceUUID, claims.UserID)
		log.Printf("[INFO] Loading files for service %s from projects directory: %s (user: %s)", serviceUUID, projectsDir, claims.UserID)
	} else {
		// User not authenticated or no valid token, use global logic
		projectsDir = h.getServiceProjectsDir(serviceUUID)
		log.Printf("[INFO] Loading files for service %s from projects directory: %s (no auth)", serviceUUID, projectsDir)
	}

	files, err := h.serviceManager.GetServiceFilesWithProjectsDir(serviceUUID, projectsDir)
	if err != nil {
		log.Printf("[ERROR] Failed to get service files for %s: %v", serviceUUID, err)
		// Return a JSON error response instead of plain text
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"error":      err.Error(),
			"service":    serviceUUID,
			"searchPath": projectsDir,
		})
		return
	}

	log.Printf("[INFO] Found %d files for service %s", len(files), serviceUUID)
	json.NewEncoder(w).Encode(map[string]any{"files": files})
}

func (h *Handler) updateServiceFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]
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

	// Use the same profile-aware directory lookup as getServiceFilesHandler
	claims, ok := extractClaimsFromRequest(r, h.authService)
	var projectsDir string
	if ok && claims != nil {
		// User is authenticated, use profile-aware directory lookup
		projectsDir = h.getServiceProjectsDirForUser(serviceUUID, claims.UserID)
		log.Printf("[INFO] Updating file for service %s from projects directory: %s (user: %s)", serviceUUID, projectsDir, claims.UserID)
	} else {
		// User not authenticated or no valid token, use global logic
		projectsDir = h.getServiceProjectsDir(serviceUUID)
		log.Printf("[INFO] Updating file for service %s from projects directory: %s (no auth)", serviceUUID, projectsDir)
	}

	if err := h.serviceManager.UpdateServiceFileWithProjectsDir(serviceUUID, filename, request.Content, projectsDir); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// previewLibrariesHandler returns a preview of libraries that can be installed for a service
func (h *Handler) previewLibrariesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check if service exists
	_, exists := h.serviceManager.GetServiceByUUID(serviceUUID)
	if !exists {
		http.Error(w, fmt.Sprintf("Service with UUID %s not found", serviceUUID), http.StatusNotFound)
		return
	}

	// Get the correct projects directory using profile-aware logic
	projectsDir := h.getServiceProjectsDir(serviceUUID)
	
	log.Printf("[INFO] Previewing libraries for service %s using projects dir: %s", serviceUUID, projectsDir)

	// Get library preview
	preview, err := h.serviceManager.PreviewLibraryInstallation(serviceUUID, projectsDir)
	if err != nil {
		log.Printf("[ERROR] Failed to preview libraries for service UUID %s: %v", serviceUUID, err)
		http.Error(w, fmt.Sprintf("Failed to preview libraries: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(preview)
}

// installSelectedLibrariesHandler installs libraries for selected environments
func (h *Handler) installSelectedLibrariesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check if service exists
	service, exists := h.serviceManager.GetServiceByUUID(serviceUUID)
	if !exists {
		http.Error(w, fmt.Sprintf("Service with UUID %s not found", serviceUUID), http.StatusNotFound)
		return
	}

	var request models.LibraryInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("[ERROR] Failed to decode request body for service %s: %v", serviceUUID, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !request.Confirmed {
		http.Error(w, "Installation must be confirmed", http.StatusBadRequest)
		return
	}

	if len(request.Environments) == 0 {
		http.Error(w, "At least one environment must be selected", http.StatusBadRequest)
		return
	}

	// Get the correct projects directory using profile-aware logic
	projectsDir := h.getServiceProjectsDir(serviceUUID)
	
	log.Printf("[INFO] Installing libraries for service %s in environments %v using projects dir: %s", 
		serviceUUID, request.Environments, projectsDir)

	// Get library preview to understand what needs to be installed
	preview, err := h.serviceManager.PreviewLibraryInstallation(serviceUUID, projectsDir)
	if err != nil {
		log.Printf("[ERROR] Failed to preview libraries for service UUID %s: %v", serviceUUID, err)
		http.Error(w, fmt.Sprintf("Failed to preview libraries: %v", err), http.StatusInternalServerError)
		return
	}

	if !preview.HasLibraries {
		http.Error(w, "No libraries found to install", http.StatusBadRequest)
		return
	}

	// Filter libraries by selected environments
	var librariesToInstall []models.LibraryInstallation
	for _, env := range preview.Environments {
		for _, selectedEnv := range request.Environments {
			if env.Environment == selectedEnv {
				librariesToInstall = append(librariesToInstall, env.Libraries...)
				break
			}
		}
	}

	if len(librariesToInstall) == 0 {
		http.Error(w, "No libraries found for selected environments", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Installing %d libraries for service %s from %d environments", 
		len(librariesToInstall), serviceUUID, len(request.Environments))

	// Install the selected libraries
	if err := h.serviceManager.InstallLibrariesWithProjectsDir(serviceUUID, librariesToInstall, projectsDir); err != nil {
		log.Printf("[ERROR] Failed to install libraries for service UUID %s: %v", serviceUUID, err)
		http.Error(w, fmt.Sprintf("Failed to install libraries: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":      "success",
		"message":     fmt.Sprintf("Successfully installed libraries for service %s", service.Name),
		"serviceName": service.Name,
		"serviceId":   serviceUUID,
		"environments": request.Environments,
		"librariesInstalled": len(librariesToInstall),
	}

	json.NewEncoder(w).Encode(response)
}

// validateWrapperHandler validates the integrity of wrapper files for a service
func (h *Handler) validateWrapperHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check if service exists
	service, exists := h.serviceManager.GetServiceByUUID(serviceUUID)
	if !exists {
		http.Error(w, fmt.Sprintf("Service with UUID %s not found", serviceUUID), http.StatusNotFound)
		return
	}

	// Get the service directory using profile-aware logic
	projectsDir := h.getServiceProjectsDir(serviceUUID)
	serviceDir := fmt.Sprintf("%s/%s", projectsDir, service.Dir)

	log.Printf("[INFO] Validating wrapper for service %s in directory: %s", service.Name, serviceDir)

	// Import the services package to access build system functions
	buildSystem := h.serviceManager.DetectBuildSystem(serviceDir)
	isValid, err := h.serviceManager.ValidateWrapperIntegrity(serviceDir, buildSystem)

	response := map[string]interface{}{
		"serviceId":     serviceUUID,
		"serviceName":   service.Name,
		"buildSystem":   string(buildSystem),
		"isValid":       isValid,
		"hasWrapper":    false,
		"wrapperFiles":  []string{},
	}

	if err != nil {
		response["error"] = err.Error()
		response["isValid"] = false
		log.Printf("[WARN] Wrapper validation failed for service %s: %v", service.Name, err)
	} else {
		log.Printf("[INFO] Wrapper validation successful for service %s", service.Name)
	}

	// Check which wrapper files exist
	switch buildSystem {
	case "maven":
		if h.serviceManager.HasMavenWrapper(serviceDir) {
			response["hasWrapper"] = true
			response["wrapperFiles"] = []string{"mvnw", ".mvn/wrapper/maven-wrapper.properties"}
		}
	case "gradle":
		if h.serviceManager.HasGradleWrapper(serviceDir) {
			response["hasWrapper"] = true
			response["wrapperFiles"] = []string{"gradlew", "gradle/wrapper/gradle-wrapper.properties"}
		}
	}

	json.NewEncoder(w).Encode(response)
}

// generateWrapperHandler generates wrapper files for a service
func (h *Handler) generateWrapperHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check if service exists
	service, exists := h.serviceManager.GetServiceByUUID(serviceUUID)
	if !exists {
		http.Error(w, fmt.Sprintf("Service with UUID %s not found", serviceUUID), http.StatusNotFound)
		return
	}

	// Get the service directory using profile-aware logic
	projectsDir := h.getServiceProjectsDir(serviceUUID)
	serviceDir := fmt.Sprintf("%s/%s", projectsDir, service.Dir)

	log.Printf("[INFO] Generating wrapper for service %s in directory: %s", service.Name, serviceDir)

	// Detect build system and generate appropriate wrapper
	buildSystem := h.serviceManager.DetectBuildSystem(serviceDir)
	var err error

	switch buildSystem {
	case "maven":
		err = h.serviceManager.GenerateMavenWrapper(serviceDir)
	case "gradle":
		err = h.serviceManager.GenerateGradleWrapper(serviceDir)
	default:
		err = fmt.Errorf("unsupported build system: %s", buildSystem)
	}

	if err != nil {
		log.Printf("[ERROR] Failed to generate wrapper for service %s: %v", service.Name, err)
		response := map[string]interface{}{
			"status":      "error",
			"message":     fmt.Sprintf("Failed to generate wrapper: %v", err),
			"serviceId":   serviceUUID,
			"serviceName": service.Name,
			"buildSystem": string(buildSystem),
		}
		http.Error(w, fmt.Sprintf("Failed to generate wrapper: %v", err), http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("[INFO] Successfully generated %s wrapper for service %s", buildSystem, service.Name)

	response := map[string]interface{}{
		"status":      "success",
		"message":     fmt.Sprintf("Successfully generated %s wrapper for service %s", buildSystem, service.Name),
		"serviceId":   serviceUUID,
		"serviceName": service.Name,
		"buildSystem": string(buildSystem),
	}

	json.NewEncoder(w).Encode(response)
}

// repairWrapperHandler repairs corrupted wrapper files for a service
func (h *Handler) repairWrapperHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check if service exists
	service, exists := h.serviceManager.GetServiceByUUID(serviceUUID)
	if !exists {
		http.Error(w, fmt.Sprintf("Service with UUID %s not found", serviceUUID), http.StatusNotFound)
		return
	}

	// Get the service directory using profile-aware logic
	projectsDir := h.getServiceProjectsDir(serviceUUID)
	serviceDir := fmt.Sprintf("%s/%s", projectsDir, service.Dir)

	log.Printf("[INFO] Repairing wrapper for service %s in directory: %s", service.Name, serviceDir)

	// Use the RepairWrapper function which detects build system and repairs accordingly
	err := h.serviceManager.RepairWrapper(serviceDir)
	if err != nil {
		log.Printf("[ERROR] Failed to repair wrapper for service %s: %v", service.Name, err)
		response := map[string]interface{}{
			"status":      "error",
			"message":     fmt.Sprintf("Failed to repair wrapper: %v", err),
			"serviceId":   serviceUUID,
			"serviceName": service.Name,
		}
		http.Error(w, fmt.Sprintf("Failed to repair wrapper: %v", err), http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	buildSystem := h.serviceManager.DetectBuildSystem(serviceDir)
	log.Printf("[INFO] Successfully repaired %s wrapper for service %s", buildSystem, service.Name)

	response := map[string]interface{}{
		"status":      "success",
		"message":     fmt.Sprintf("Successfully repaired %s wrapper for service %s", buildSystem, service.Name),
		"serviceId":   serviceUUID,
		"serviceName": service.Name,
		"buildSystem": string(buildSystem),
	}

	json.NewEncoder(w).Encode(response)
}
