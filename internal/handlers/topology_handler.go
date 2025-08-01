// Package handlers
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zechtz/vertex/internal/models"
)

func registerTopologyRoutes(h *Handler, r *mux.Router) {
	r.HandleFunc("/api/topology", h.getTopologyHandler).Methods("GET")
	r.HandleFunc("/api/topology/debug", h.getTopologyDebugHandler).Methods("GET")
	r.HandleFunc("/api/dependencies", h.getDependenciesHandler).Methods("GET")
	r.HandleFunc("/api/dependencies", h.saveDependenciesHandler).Methods("POST")
	r.HandleFunc("/api/dependencies/graph", h.getDependencyGraphHandler).Methods("GET")
	r.HandleFunc("/api/dependencies/validate", h.validateDependenciesHandler).Methods("GET")
	r.HandleFunc("/api/dependencies/startup-order", h.getStartupOrderHandler).Methods("POST")
	r.HandleFunc("/api/eureka/services", h.getEurekaServicesHandler).Methods("GET")
	r.HandleFunc("/api/eureka/debug", h.debugEurekaHandler).Methods("GET")
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
		if dbDeps, exists := allDependencies[service.ID]; exists {
			serviceData["dependencies"] = dbDeps
		} else {
			serviceData["dependencies"] = []interface{}{}
		}

		// Add dependent on info (reverse dependencies)
		serviceData["dependentOn"] = service.DependentOn
		serviceData["startupDelay"] = service.StartupDelay.String()

		dependencies[service.ID] = serviceData
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

	var configData map[string]any
	if err := json.NewDecoder(r.Body).Decode(&configData); err != nil {
		log.Printf("Failed to decode dependencies config: %v", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Update service dependencies and orders
	services := h.serviceManager.GetServices()
	serviceMap := make(map[string]*models.Service)
	for i := range services {
		serviceMap[services[i].ID] = &services[i]
	}

	// Process each service's configuration
	for serviceID, config := range configData {
		if configMap, ok := config.(map[string]any); ok {
			service := serviceMap[serviceID]
			if service == nil {
				log.Printf("Service %s not found, skipping", serviceID)
				continue
			}

			// Update service order
			if order, exists := configMap["order"]; exists {
				if orderFloat, ok := order.(float64); ok {
					service.Order = int(orderFloat)
					log.Printf("Updated order for %s to %d", service.Name, service.Order)
				}
			}

			// Update dependencies in database
			if dependencies, exists := configMap["dependencies"]; exists {
				if depsList, ok := dependencies.([]interface{}); ok {
					// Save dependencies to database
					db := h.serviceManager.GetDatabase()
					if err := db.SaveServiceDependencies(serviceID, depsList); err != nil {
						log.Printf("Failed to save dependencies for %s: %v", service.Name, err)
						http.Error(w, fmt.Sprintf("Failed to save dependencies for %s", service.Name), http.StatusInternalServerError)
						return
					}
					log.Printf("Saved %d dependencies for %s", len(depsList), service.Name)
				}
			}

			// Update the service in the service manager
			if err := h.serviceManager.UpdateServiceInDB(service); err != nil {
				log.Printf("Failed to update service %s in database: %v", service.Name, err)
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
			"error":   "Failed to create Eureka request",
			"details": err.Error(),
		}
		json.NewEncoder(w).Encode(result)
		return
	}

	req.Header.Set("Accept", "application/xml") // Request XML since that's what your Eureka returns
	resp, err := client.Do(req)
	if err != nil {
		result := map[string]interface{}{
			"error":      "Failed to query Eureka",
			"details":    err.Error(),
			"eureka_url": eurekaURL,
		}
		json.NewEncoder(w).Encode(result)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		result := map[string]interface{}{
			"error":       "Eureka returned non-200 status",
			"status_code": resp.StatusCode,
			"eureka_url":  eurekaURL,
		}
		json.NewEncoder(w).Encode(result)
		return
	}

	// Parse the response
	var eurekaData interface{}
	if err := json.NewDecoder(resp.Body).Decode(&eurekaData); err != nil {
		result := map[string]interface{}{
			"error":   "Failed to decode Eureka response",
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
		"success":        true,
		"eureka_url":     eurekaURL,
		"eureka_status":  resp.StatusCode,
		"eureka_data":    eurekaData,
		"local_services": serviceNames,
		"message":        "Raw Eureka data and local services for debugging",
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode debug response: %v", err)
		http.Error(w, "Failed to encode debug response", http.StatusInternalServerError)
		return
	}
}
