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
	"github.com/zechtz/vertex/internal/database"
	"github.com/zechtz/vertex/internal/models"
	"github.com/zechtz/vertex/internal/services"
)

func registerUtilityRoutes(h *Handler, r *mux.Router) {
	r.HandleFunc("/api/system/metrics", h.getSystemMetricsHandler).Methods("GET")
	r.HandleFunc("/api/system/logs/cleanup", h.cleanupLogsHandler).Methods("POST")

	r.HandleFunc("/api/logs/search", h.searchLogsHandler).Methods("POST")
	r.HandleFunc("/api/logs/statistics", h.getLogStatisticsHandler).Methods("GET")
	r.HandleFunc("/api/logs/export", h.exportLogsHandler).Methods("POST")

	r.HandleFunc("/api/services/fix-lombok", h.fixLombokHandler).Methods("POST")
	r.HandleFunc("/api/environment/setup", h.setupEnvironmentHandler).Methods("POST")
	r.HandleFunc("/api/environment/sync", h.syncEnvironmentHandler).Methods("POST")
	r.HandleFunc("/api/environment/status", h.getEnvironmentStatusHandler).Methods("GET")

	r.HandleFunc("/api/env-vars/global", h.getGlobalEnvVarsHandler).Methods("GET")
	r.HandleFunc("/api/env-vars/global", h.updateGlobalEnvVarsHandler).Methods("PUT")
	r.HandleFunc("/api/env-vars/reload", h.reloadEnvVarsHandler).Methods("POST")
	r.HandleFunc("/api/env-vars/cleanup", h.cleanupGlobalEnvVarsHandler).Methods("POST")

	r.HandleFunc("/api/auto-discovery/scan", h.scanAutoDiscoveryHandler).Methods("POST")
	r.HandleFunc("/api/auto-discovery/services", h.getDiscoveredServicesHandler).Methods("GET")
	r.HandleFunc("/api/auto-discovery/import", h.importDiscoveredServiceHandler).Methods("POST")
	r.HandleFunc("/api/auto-discovery/import-bulk", h.importDiscoveredServicesBulkHandler).Methods("POST")

	r.HandleFunc("/ws", h.websocketHandler)
}

func (h *Handler) getSystemMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get system resource summary
	summary := h.serviceManager.GetSystemResourceSummary()

	// Add individual service metrics
	services := h.serviceManager.GetServices()
	serviceMetrics := make([]map[string]any, 0)

	for _, service := range services {
		if service.Status == "running" {
			serviceMetric := map[string]any{
				"name":          service.Name,
				"cpuPercent":    service.CPUPercent,
				"memoryUsage":   service.MemoryUsage,
				"memoryPercent": service.MemoryPercent,
				"diskUsage":     service.DiskUsage,
				"networkRx":     service.NetworkRx,
				"networkTx":     service.NetworkTx,
				"status":        service.Status,
				"healthStatus":  service.HealthStatus,
				"uptime":        service.Uptime,
				"errorRate":     service.Metrics.ErrorRate,
				"requestCount":  service.Metrics.RequestCount,
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
		ServiceIDs []string `json:"serviceIds"`
		Levels     []string `json:"levels"`
		SearchText string   `json:"searchText"`
		StartTime  string   `json:"startTime"`
		EndTime    string   `json:"endTime"`
		Limit      int      `json:"limit"`
		Offset     int      `json:"offset"`
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
		ServiceIDs: criteria.ServiceIDs,
		Levels:     criteria.Levels,
		SearchText: criteria.SearchText,
		StartTime:  startTime,
		EndTime:    endTime,
		Limit:      criteria.Limit,
		Offset:     criteria.Offset,
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
		ServiceIDs []string `json:"serviceIds"`
		Levels     []string `json:"levels"`
		SearchText string   `json:"searchText"`
		StartTime  string   `json:"startTime"`
		EndTime    string   `json:"endTime"`
		Format     string   `json:"format"` // "json", "csv", "txt"
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
		ServiceIDs: exportRequest.ServiceIDs,
		Levels:     exportRequest.Levels,
		SearchText: exportRequest.SearchText,
		StartTime:  startTime,
		EndTime:    endTime,
		Limit:      0, // No limit for export
		Offset:     0,
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
				result.ServiceID,
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
				result.ServiceID,
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

	// Create the service first
	service, err := h.autoDiscoveryService.CreateServiceFromDiscovered(discoveredService)
	if err != nil {
		log.Printf("[ERROR] Failed to import discovered service %s: %v", discoveredService.Name, err)
		http.Error(w, fmt.Sprintf("Failed to import service: %v", err), http.StatusInternalServerError)
		return
	}

	// Try to add the service to the user's active profile (if authenticated)
	if claims, ok := extractClaimsFromRequest(r, h.authService); ok && claims != nil {
		log.Printf("[DEBUG] Single import - User authenticated: %s", claims.UserID)
		if activeProfile, err := h.profileService.GetActiveProfile(claims.UserID); err == nil && activeProfile != nil {
			log.Printf("[DEBUG] Single import - Adding service %s (UUID: %s) to active profile %s (ID: %s)", service.Name, service.ID, activeProfile.Name, activeProfile.ID)

			// Verify service exists in service manager
			if _, exists := h.serviceManager.GetServiceByUUID(service.ID); !exists {
				log.Printf("[ERROR] Service %s (UUID: %s) not found in service manager after creation", service.Name, service.ID)
			} else {
				log.Printf("[DEBUG] Service %s (UUID: %s) confirmed to exist in service manager", service.Name, service.ID)
			}

			// Add service to the active profile using service ID
			if err := h.profileService.AddServiceToProfile(claims.UserID, activeProfile.ID, service.ID); err != nil {
				log.Printf("[WARN] Successfully imported service %s but failed to add to active profile %s: %v", service.Name, activeProfile.Name, err)
			} else {
				log.Printf("[INFO] Successfully imported service %s and added to active profile %s", service.Name, activeProfile.Name)
			}
		} else {
			log.Printf("[WARN] Failed to get active profile for user %s: %v", claims.UserID, err)
		}
	} else {
		log.Printf("[WARN] Single import - No authentication found, service will not be added to profile")
	}

	result := map[string]any{
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

func (h *Handler) importDiscoveredServicesBulkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var request struct {
		Services []services.DiscoveredService `json:"services"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("[ERROR] Failed to decode bulk import request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Bulk importing %d discovered services", len(request.Services))

	// Get user's active profile for adding services (if authenticated)
	var activeProfile *models.ServiceProfile
	var userClaims *models.JWTClaims
	if claims, ok := extractClaimsFromRequest(r, h.authService); ok && claims != nil {
		userClaims = claims
		log.Printf("[DEBUG] Bulk import - User authenticated: %s", claims.UserID)
		if profile, err := h.profileService.GetActiveProfile(claims.UserID); err == nil {
			activeProfile = profile
			log.Printf("[INFO] Will add imported services to active profile: %s (ID: %s)", activeProfile.Name, activeProfile.ID)
		} else {
			log.Printf("[WARN] Failed to get active profile for user %s: %v", claims.UserID, err)
		}
	} else {
		log.Printf("[WARN] Bulk import - No authentication found, services will not be added to profile")
	}

	var importedServices []any
	var errors []string
	var profileErrors []string

	for _, discoveredService := range request.Services {
		log.Printf("[INFO] Importing discovered service: %s from %s", discoveredService.Name, discoveredService.Path)

		// Create the service globally first
		service, err := h.autoDiscoveryService.CreateServiceFromDiscovered(discoveredService)
		if err != nil {
			log.Printf("[ERROR] Failed to import discovered service %s: %v", discoveredService.Name, err)
			errors = append(errors, fmt.Sprintf("Failed to import %s: %v", discoveredService.Name, err))
			continue
		}

		importedServices = append(importedServices, service)
		log.Printf("[INFO] Successfully imported service: %s", service.Name)

		// Add to active profile if available
		if activeProfile != nil && userClaims != nil {
			log.Printf("[DEBUG] Adding service %s (UUID: %s) to profile %s (ID: %s)", service.Name, service.ID, activeProfile.Name, activeProfile.ID)

			// Verify service exists in service manager
			if _, exists := h.serviceManager.GetServiceByUUID(service.ID); !exists {
				profileErrors = append(profileErrors, fmt.Sprintf("Service %s (UUID: %s) not found in service manager", service.Name, service.ID))
				log.Printf("[ERROR] Service %s (UUID: %s) not found in service manager after creation", service.Name, service.ID)
			} else {
				log.Printf("[DEBUG] Service %s (UUID: %s) confirmed to exist in service manager", service.Name, service.ID)
				if err := h.profileService.AddServiceToProfile(userClaims.UserID, activeProfile.ID, service.ID); err != nil {
					profileErrors = append(profileErrors, fmt.Sprintf("Failed to add %s to profile: %v", service.Name, err))
					log.Printf("[WARN] Failed to add service %s to active profile %s: %v", service.Name, activeProfile.Name, err)
				} else {
					log.Printf("[INFO] Successfully added service %s to active profile %s", service.Name, activeProfile.Name)
				}
			}
		}
	}

	// Combine all errors for response
	allErrors := errors
	if len(profileErrors) > 0 {
		allErrors = append(allErrors, profileErrors...)
	}

	result := map[string]any{
		"success":          len(errors) == 0, // Only consider import errors for success status
		"message":          fmt.Sprintf("Bulk import completed. Imported %d/%d services", len(importedServices), len(request.Services)),
		"importedServices": importedServices,
		"errors":           allErrors,
		"profileErrors":    profileErrors,
		"totalRequested":   len(request.Services),
		"totalImported":    len(importedServices),
	}

	log.Printf("[INFO] Bulk import completed: %d/%d services imported successfully", len(importedServices), len(request.Services))

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode bulk import response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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
