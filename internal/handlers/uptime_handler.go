// Package handlers - Uptime statistics handler
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zechtz/vertex/internal/models"
	"github.com/zechtz/vertex/internal/services"
)

func registerUptimeRoutes(h *Handler, r *mux.Router) {
	r.HandleFunc("/api/uptime/statistics", h.getUptimeStatisticsHandler).Methods("GET")
	r.HandleFunc("/api/uptime/statistics/{id}", h.getServiceUptimeStatisticsHandler).Methods("GET")
}

// getUptimeStatisticsHandler returns uptime statistics for services in the current active profile
func (h *Handler) getUptimeStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get user from JWT token to determine active profile
	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	uptimeTracker := services.GetUptimeTracker()
	allStats := uptimeTracker.GetAllUptimeStats()

	// Get services from active profile, fallback to all services if no active profile
	var services []models.Service
	activeProfile, err := h.profileService.GetActiveProfile(claims.UserID)
	if err != nil || activeProfile == nil {
		// No active profile, show all services (fallback for backward compatibility)
		services = h.serviceManager.GetServices()
	} else {
		// Filter services by active profile
		allServices := h.serviceManager.GetServices()
		for _, service := range allServices {
			// Check if service is in the active profile
			for _, serviceUUID := range activeProfile.Services {
				if service.ID == serviceUUID {
					services = append(services, service)
					break
				}
			}
		}
	}

	// Enhance with service names
	serviceStats := make(map[string]interface{})

	for _, service := range services {
		if stats, exists := allStats[service.ID]; exists {
			serviceStats[service.ID] = map[string]interface{}{
				"serviceName":  service.Name,
				"serviceId":    service.ID,
				"port":         service.Port,
				"status":       service.Status,
				"healthStatus": service.HealthStatus,
				"stats":        stats,
			}
		} else {
			// Create default stats for services without events
			serviceStats[service.ID] = map[string]interface{}{
				"serviceName":  service.Name,
				"serviceId":    service.ID,
				"port":         service.Port,
				"status":       service.Status,
				"healthStatus": service.HealthStatus,
				"stats": map[string]interface{}{
					"totalRestarts":       0,
					"uptimePercentage24h": 100.0,
					"uptimePercentage7d":  100.0,
					"mtbf":                0,
					"lastDowntime":        nil,
					"totalDowntime24h":    0,
					"totalDowntime7d":     0,
				},
			}
		}
	}

	response := map[string]interface{}{
		"statistics": serviceStats,
		"summary": map[string]interface{}{
			"totalServices":     len(services),
			"runningServices":   countRunningServices(services),
			"unhealthyServices": countUnhealthyServices(services),
		},
	}

	json.NewEncoder(w).Encode(response)
}

// getServiceUptimeStatisticsHandler returns uptime statistics for a specific service
func (h *Handler) getServiceUptimeStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	serviceID := vars["id"]

	service, exists := h.serviceManager.GetServiceByUUID(serviceID)
	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	uptimeTracker := services.GetUptimeTracker()
	stats := uptimeTracker.CalculateUptimeStats(serviceID, service)

	response := map[string]interface{}{
		"serviceName":  service.Name,
		"serviceId":    service.ID,
		"port":         service.Port,
		"status":       service.Status,
		"healthStatus": service.HealthStatus,
		"stats":        stats,
	}

	json.NewEncoder(w).Encode(response)
}

// Helper functions
func countRunningServices(services []models.Service) int {
	count := 0
	for _, service := range services {
		if service.Status == "running" {
			count++
		}
	}
	return count
}

func countUnhealthyServices(services []models.Service) int {
	count := 0
	for _, service := range services {
		if service.Status == "running" && service.HealthStatus == "unhealthy" {
			count++
		}
	}
	return count
}

func getRunningServices(services []*models.Service) []*models.Service {
	var running []*models.Service
	for _, service := range services {
		if service.Status == "running" {
			running = append(running, service)
		}
	}
	return running
}

func getUnhealthyServices(services []*models.Service) []*models.Service {
	var unhealthy []*models.Service
	for _, service := range services {
		if service.Status == "running" && service.HealthStatus == "unhealthy" {
			unhealthy = append(unhealthy, service)
		}
	}
	return unhealthy
}
