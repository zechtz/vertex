// Package handlers
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/zechtz/nest-up/internal/database"
	"github.com/zechtz/nest-up/internal/models"
	"github.com/zechtz/nest-up/internal/services"
)

type Handler struct {
	serviceManager *services.Manager
	upgrader       websocket.Upgrader
}

func NewHandler(sm *services.Manager) *Handler {
	return &Handler{
		serviceManager: sm,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/services", h.getServicesHandler).Methods("GET")
	r.HandleFunc("/api/services/{name}/start", h.startServiceHandler).Methods("POST")
	r.HandleFunc("/api/services/{name}/stop", h.stopServiceHandler).Methods("POST")
	r.HandleFunc("/api/services/{name}/restart", h.restartServiceHandler).Methods("POST")
	r.HandleFunc("/api/services/{name}/health", h.checkHealthHandler).Methods("POST")
	r.HandleFunc("/api/services/{name}/logs", h.getLogsHandler).Methods("GET")
	r.HandleFunc("/api/services/{name}/logs", h.clearLogsHandler).Methods("DELETE")
	r.HandleFunc("/api/services/{name}/metrics", h.getServiceMetricsHandler).Methods("GET")
	r.HandleFunc("/api/logs/search", h.searchLogsHandler).Methods("POST")
	r.HandleFunc("/api/logs/statistics", h.getLogStatisticsHandler).Methods("GET")
	r.HandleFunc("/api/logs/export", h.exportLogsHandler).Methods("POST")
	r.HandleFunc("/api/services/start-all", h.startAllHandler).Methods("POST")
	r.HandleFunc("/api/services/stop-all", h.stopAllHandler).Methods("POST")
	r.HandleFunc("/api/system/metrics", h.getSystemMetricsHandler).Methods("GET")
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

	if err := h.serviceManager.StartService(serviceName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

func (h *Handler) stopAllHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := h.serviceManager.StopAllServices(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "stopping all services"})
}

func (h *Handler) restartServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := h.serviceManager.RestartService(serviceName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "restarted"})
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
	// Placeholder - implement service creation
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{"status": "not implemented"})
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
	// Placeholder - implement service deletion
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{"status": "not implemented"})
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

	files, err := h.serviceManager.GetServiceFiles(serviceName)
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
