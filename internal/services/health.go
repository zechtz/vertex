// Package services
package services

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zechtz/vertex/internal/models"
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

// healthProbeTarget is the snapshot a probe runs against. Probing performs HTTP
// requests, which must not happen while the service lock is held: GetServices
// waits on that lock while holding the manager lock, so one unresponsive health
// endpoint would otherwise stall every reader of the manager.
type healthProbeTarget struct {
	name        string
	port        int
	healthURL   string
	lastStarted time.Time
}

func (sm *Manager) checkServiceHealth(service *models.Service) {
	service.Mutex.RLock()
	status := service.Status
	healthStatus := service.HealthStatus
	pid := service.PID
	target := healthProbeTarget{
		name:        service.Name,
		port:        service.Port,
		healthURL:   service.HealthURL,
		lastStarted: service.LastStarted,
	}
	service.Mutex.RUnlock()

	// Check if process is still running
	if status == "running" && pid > 0 && !sm.isProcessRunning(pid) {
		log.Printf("Process %d for service %s is no longer running", pid, target.name)

		service.Mutex.Lock()
		service.Status = "stopped"
		service.HealthStatus = "unknown"
		service.PID = 0
		service.Cmd = nil
		service.Uptime = ""
		service.Mutex.Unlock()

		sm.publishServiceState(service)
		return
	}

	if status != "running" {
		sm.applyHealthStatus(service, "unknown", "")
		return
	}

	// Give new services time to initialize their health endpoints.
	// Config services especially need time for actuator endpoints to be ready.
	if !target.lastStarted.IsZero() && time.Since(target.lastStarted) < 30*time.Second {
		// Keep services in "starting" state for first 30 seconds
		if healthStatus != "starting" {
			sm.applyHealthStatus(service, "starting", "")
		}
		return
	}

	// Calculate uptime
	uptime := ""
	if !target.lastStarted.IsZero() {
		uptime = formatDuration(time.Since(target.lastStarted))
	}

	sm.applyHealthStatus(service, sm.probeHealth(target), uptime)
}

// applyHealthStatus records a probe result and publishes it. The service lock is
// held only for the field writes, never across the probe that produced them.
func (sm *Manager) applyHealthStatus(service *models.Service, healthStatus, uptime string) {
	service.Mutex.Lock()
	service.HealthStatus = healthStatus
	if uptime != "" {
		service.Uptime = uptime
	}
	service.Mutex.Unlock()

	sm.publishServiceState(service)
}

// probeHealth determines a service's health status. It performs only network
// I/O and holds no locks, so a slow endpoint delays nothing but itself.
func (sm *Manager) probeHealth(target healthProbeTarget) string {
	// Try Eureka-based health check first (for microservices that register with Eureka)
	if status, registered := sm.checkEurekaHealth(target.name, target.port); registered {
		log.Printf("[DEBUG] Health status for %s updated from Eureka: %s", target.name, status)
		return status
	}

	// Fall back to direct HTTP health check
	log.Printf("[DEBUG] Using direct health check for %s (not found in Eureka or Eureka unavailable)", target.name)
	client := sm.createHealthCheckClient()
	req, err := sm.createHealthCheckRequest(target.healthURL)
	if err != nil {
		return "unhealthy"
	}

	resp, err := client.Do(req)
	if err != nil {
		// For CONFIG services, treat an early failure as still initializing
		if strings.ToUpper(target.name) == "CONFIG" && !target.lastStarted.IsZero() &&
			time.Since(target.lastStarted) < 2*time.Minute {
			log.Printf("[DEBUG] Health check failed for %s (still initializing): %v", target.name, err)
			return "starting"
		}

		log.Printf("[DEBUG] Health check failed for %s: %v", target.name, err)

		// If health endpoint fails, try a simple connectivity test to the service port
		simpleReq, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/", target.port), nil)
		if err != nil {
			return "unhealthy"
		}

		simpleResp, err := client.Do(simpleReq)
		if err != nil {
			log.Printf("[DEBUG] Service %s is not responsive: %v", target.name, err)
			return "unhealthy"
		}
		defer simpleResp.Body.Close()

		log.Printf("[DEBUG] Service %s is responsive on port %d (HTTP %d) but health endpoint failed",
			target.name, target.port, simpleResp.StatusCode)
		return "running" // Service is running but health endpoint misconfigured
	}
	defer resp.Body.Close()

	log.Printf("[DEBUG] Health check for %s returned status: %d", target.name, resp.StatusCode)

	switch {
	case resp.StatusCode == 200:
		// For Spring Boot actuator, also check response body
		if !strings.Contains(target.healthURL, "actuator/health") {
			return "healthy"
		}

		body := make([]byte, 1000) // Read first 1000 bytes
		n, _ := resp.Body.Read(body)
		bodyStr := string(body[:n])
		log.Printf("[DEBUG] Health check response for %s: %s", target.name, bodyStr)

		if n > 0 && strings.Contains(bodyStr, `"status":"UP"`) {
			return "healthy"
		}
		return "unhealthy"

	case resp.StatusCode == 404 && strings.Contains(target.healthURL, "actuator/health"):
		// Actuator endpoint not found, but service is responding - check if it's a gateway
		if strings.ToUpper(target.name) != "GATEWAY" {
			return "unhealthy"
		}

		// For gateway services, a structured JSON 404 means it is running but
		// the actuator endpoint is simply not exposed
		body := make([]byte, 200)
		n, _ := resp.Body.Read(body)
		bodyStr := string(body[:n])

		if strings.Contains(bodyStr, `"error":"Not Found"`) && strings.Contains(bodyStr, "timestamp") {
			log.Printf("[DEBUG] Gateway %s is healthy - responding with structured 404", target.name)
			return "healthy"
		}
		return "unhealthy"

	case resp.StatusCode == 401:
		// Unauthorized - auth issue, but service is running and responding
		log.Printf("[DEBUG] Health check for %s returned 401 - service is running but requires different auth", target.name)

		// Try without auth for services that might not need it
		reqNoAuth, err := http.NewRequest("GET", target.healthURL, nil)
		if err != nil {
			return "running"
		}

		respNoAuth, err := client.Do(reqNoAuth)
		if err != nil {
			return "running"
		}
		defer respNoAuth.Body.Close()

		if respNoAuth.StatusCode == 200 {
			log.Printf("[DEBUG] Health check for %s succeeded without auth", target.name)
			return "healthy"
		}

		// Service is running but health endpoint needs different config
		log.Printf("[DEBUG] Service %s is running (responds to HTTP) but health endpoint misconfigured", target.name)
		return "running"

	default:
		return "unhealthy"
	}
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

// createHealthCheckClient creates an HTTP client for health checks
func (sm *Manager) createHealthCheckClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second} // Increased timeout for Spring Boot services
}

// createHealthCheckRequest creates an HTTP request for health checks with authentication
func (sm *Manager) createHealthCheckRequest(healthURL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", healthURL, nil)
	if err != nil {
		return nil, err
	}

	// Add basic auth for Spring Boot services
	if strings.Contains(healthURL, "actuator/health") {
		// Get credentials from environment variables
		username := os.Getenv("CONFIG_USERNAME")
		password := os.Getenv("CONFIG_PASSWORD")
		req.SetBasicAuth(username, password)
	}

	return req, nil
}
