// Package services
package services

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zechtz/nest-up/internal/models"
)

func (sm *Manager) CheckServiceHealth(serviceName string) error {
	sm.mutex.RLock()
	service, exists := sm.services[serviceName]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	go sm.checkServiceHealth(service)
	return nil
}

func (sm *Manager) healthCheckRoutine() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.performHealthChecks()
		}
	}
}

func (sm *Manager) performHealthChecks() {
	sm.mutex.RLock()
	services := make([]*models.Service, 0, len(sm.services))
	for _, service := range sm.services {
		services = append(services, service)
	}
	sm.mutex.RUnlock()

	for _, service := range services {
		go sm.checkServiceHealth(service)
	}
}

func (sm *Manager) checkServiceHealth(service *models.Service) {
	service.Mutex.Lock()
	defer service.Mutex.Unlock()

	// Check if process is still running
	if service.Status == "running" && service.PID > 0 {
		// Check if process still exists
		if !sm.isProcessRunning(service.PID) {
			log.Printf("Process %d for service %s is no longer running", service.PID, service.Name)
			service.Status = "stopped"
			service.HealthStatus = "unknown"
			service.PID = 0
			service.Cmd = nil
			service.Uptime = ""
			sm.updateServiceInDB(service)
			sm.broadcastUpdate(service)
			return
		}
	}

	if service.Status != "running" {
		service.HealthStatus = "unknown"
		sm.updateServiceInDB(service)
		return
	}

	// Calculate uptime
	if !service.LastStarted.IsZero() {
		uptime := time.Since(service.LastStarted)
		service.Uptime = formatDuration(uptime)
	}

	// Perform HTTP health check with authentication
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", service.HealthURL, nil)
	if err != nil {
		service.HealthStatus = "unhealthy"
		sm.updateServiceInDB(service)
		sm.broadcastUpdate(service)
		return
	}

	// Add basic auth for Spring Boot services
	if strings.Contains(service.HealthURL, "actuator/health") {
		// Get credentials from environment variables
		username := os.Getenv("CONFIG_USERNAME")
		password := os.Getenv("CONFIG_PASSWORD")
		if username == "" {
			username = "nest"
		}
		if password == "" {
			password = "1kzwjz2nzegt3nest@ppra.go.tza1q@BmM0Oo"
		}
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[DEBUG] Health check failed for %s: %v", service.Name, err)
		
		// If health endpoint fails, try a simple connectivity test to the service port
		simpleURL := fmt.Sprintf("http://localhost:%d/", service.Port)
		simpleReq, err := http.NewRequest("GET", simpleURL, nil)
		if err == nil {
			simpleResp, err := client.Do(simpleReq)
			if err == nil {
				defer simpleResp.Body.Close()
				log.Printf("[DEBUG] Service %s is responsive on port %d (HTTP %d) but health endpoint failed", 
					service.Name, service.Port, simpleResp.StatusCode)
				service.HealthStatus = "running" // Service is running but health endpoint misconfigured
			} else {
				log.Printf("[DEBUG] Service %s is not responsive: %v", service.Name, err)
				service.HealthStatus = "unhealthy"
			}
		} else {
			service.HealthStatus = "unhealthy"
		}
	} else {
		defer resp.Body.Close()
		log.Printf("[DEBUG] Health check for %s returned status: %d", service.Name, resp.StatusCode)

		if resp.StatusCode == 200 {
			// For Spring Boot actuator, also check response body
			if strings.Contains(service.HealthURL, "actuator/health") {
				body := make([]byte, 1000) // Read first 1000 bytes
				n, _ := resp.Body.Read(body)
				bodyStr := string(body[:n])
				log.Printf("[DEBUG] Health check response for %s: %s", service.Name, bodyStr)

				if n > 0 && strings.Contains(bodyStr, `"status":"UP"`) {
					service.HealthStatus = "healthy"
				} else {
					service.HealthStatus = "unhealthy"
				}
			} else {
				service.HealthStatus = "healthy"
			}
		} else if resp.StatusCode == 404 && strings.Contains(service.HealthURL, "actuator/health") {
			// Actuator endpoint not found, but service is responding - check if it's a gateway
			if strings.ToUpper(service.Name) == "GATEWAY" {
				// For gateway services, a 404 with JSON response means it's running but actuator not exposed
				body := make([]byte, 200)
				n, _ := resp.Body.Read(body)
				bodyStr := string(body[:n])

				// If we get a JSON 404 response, the service is healthy but endpoint not configured
				if strings.Contains(bodyStr, `"error":"Not Found"`) && strings.Contains(bodyStr, "timestamp") {
					log.Printf("[DEBUG] Gateway %s is healthy - responding with structured 404", service.Name)
					service.HealthStatus = "healthy"
				} else {
					service.HealthStatus = "unhealthy"
				}
			} else {
				service.HealthStatus = "unhealthy"
			}
		} else if resp.StatusCode == 401 {
			// Unauthorized - auth issue, but service is running and responding
			log.Printf("[DEBUG] Health check for %s returned 401 - service is running but requires different auth", service.Name)
			// Try without auth for services that might not need it
			reqNoAuth, err := http.NewRequest("GET", service.HealthURL, nil)
			if err == nil {
				respNoAuth, err := client.Do(reqNoAuth)
				if err == nil {
					defer respNoAuth.Body.Close()
					if respNoAuth.StatusCode == 200 {
						log.Printf("[DEBUG] Health check for %s succeeded without auth", service.Name)
						service.HealthStatus = "healthy"
					} else {
						// Service is running but health endpoint needs different config
						log.Printf("[DEBUG] Service %s is running (responds to HTTP) but health endpoint misconfigured", service.Name)
						service.HealthStatus = "running" // Mark as running instead of unhealthy
					}
				} else {
					// Service is running but health endpoint needs different config
					service.HealthStatus = "running"
				}
			} else {
				service.HealthStatus = "running"
			}
		} else {
			service.HealthStatus = "unhealthy"
		}
	}

	// Update database and broadcast
	sm.updateServiceInDB(service)
	sm.broadcastUpdate(service)
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	} else {
		days := int(d.Hours()) / 24
		hours := int(d.Hours()) % 24
		return fmt.Sprintf("%dd %dh", days, hours)
	}
}

func (sm *Manager) isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	// Use platform-specific function to check if process exists
	return IsProcessRunning(pid)
}
