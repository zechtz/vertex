// Package services - Eureka-based health checking
package services

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/zechtz/vertex/internal/models"
)

// XML structures for Eureka response
type EurekaXMLInstanceInfo struct {
	InstanceID string `xml:"instanceId"`
	App        string `xml:"app"`
	HostName   string `xml:"hostName"`
	IPAddr     string `xml:"ipAddr"`
	Status     string `xml:"status"`
	Port       struct {
		Port    int  `xml:",chardata"`
		Enabled bool `xml:"enabled,attr"`
	} `xml:"port"`
	HomePageURL    string `xml:"homePageUrl"`
	StatusPageURL  string `xml:"statusPageUrl"`
	HealthCheckURL string `xml:"healthCheckUrl"`
}

type EurekaXMLApplication struct {
	Name      string                  `xml:"name"`
	Instances []EurekaXMLInstanceInfo `xml:"instance"`
}

type EurekaXMLApplications struct {
	XMLName      xml.Name               `xml:"applications"`
	Applications []EurekaXMLApplication `xml:"application"`
}

// JSON structures for Eureka response (kept for compatibility)
type EurekaInstanceInfo struct {
	InstanceID string `json:"instanceId"`
	App        string `json:"app"`
	HostName   string `json:"hostName"`
	IPAddr     string `json:"ipAddr"`
	Status     string `json:"status"`
	Port       struct {
		Port    int  `json:"$"`
		Enabled bool `json:"@enabled"`
	} `json:"port"`
	SecurePort struct {
		Port    int  `json:"$"`
		Enabled bool `json:"@enabled"`
	} `json:"securePort"`
	HomePageURL          string `json:"homePageUrl"`
	StatusPageURL        string `json:"statusPageUrl"`
	HealthCheckURL       string `json:"healthCheckUrl"`
	LastUpdatedTimestamp string `json:"lastUpdatedTimestamp"`
	LastDirtyTimestamp   string `json:"lastDirtyTimestamp"`
}

type EurekaApplication struct {
	Name      string               `json:"name"`
	Instances []EurekaInstanceInfo `json:"instance"`
}

type EurekaApplications struct {
	Applications struct {
		VersionsDelta string              `json:"versions__delta"`
		AppsHashcode  string              `json:"apps__hashcode"`
		Applications  []EurekaApplication `json:"application"`
	} `json:"applications"`
}

// checkEurekaHealth checks service health via Eureka registry
func (sm *Manager) checkEurekaHealth(service *models.Service) bool {
	// Only check Eureka for services that should be registered (not Eureka itself)
	serviceName := strings.ToUpper(service.Name)
	if serviceName == "EUREKA" {
		log.Printf("[DEBUG] Skipping Eureka health check for %s (is registry service itself)", service.Name)
		return false // Use direct health check for Eureka itself
	}

	// Get Eureka registry port from environment or use default
	eurekaPort := 8800
	if service.Name == "EUREKA" {
		eurekaPort = service.Port
	}

	// Add small random delay to stagger concurrent requests
	delay := time.Duration(rand.Intn(500)) * time.Millisecond
	time.Sleep(delay)

	// Query Eureka for all applications
	eurekaURL := fmt.Sprintf("http://localhost:%d/eureka/apps", eurekaPort)
	log.Printf("[DEBUG] Checking Eureka health for %s at %s (after %v delay)", service.Name, eurekaURL, delay)

	// Use a client with reasonable timeout and keep-alive disabled to avoid connection pool issues
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	req, err := http.NewRequest("GET", eurekaURL, nil)
	if err != nil {
		log.Printf("[DEBUG] Failed to create Eureka request for %s: %v", service.Name, err)
		return false
	}

	// Request XML (since we know that's what Eureka returns)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Connection", "close")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[DEBUG] Failed to query Eureka for %s: %v", service.Name, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("[DEBUG] Eureka returned status %d for %s", resp.StatusCode, service.Name)
		return false
	}

	// Read the entire response body at once
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[DEBUG] Failed to read Eureka response for %s: %v", service.Name, err)
		return false
	}

	// Try to parse as XML first (since that's what your Eureka returns)
	var xmlResponse EurekaXMLApplications
	if err := xml.Unmarshal(bodyBytes, &xmlResponse); err == nil {
		log.Printf("[DEBUG] Successfully parsed Eureka XML response for %s, found %d applications", service.Name, len(xmlResponse.Applications))

		// Debug: List all applications found in Eureka
		for i, app := range xmlResponse.Applications {
			log.Printf("[DEBUG] Eureka app[%d]: %s with %d instances", i, app.Name, len(app.Instances))
			for j, instance := range app.Instances {
				log.Printf("[DEBUG]   instance[%d]: %s:%d status=%s", j, instance.HostName, instance.Port.Port, instance.Status)
			}
		}

		// Look for the service in Eureka registry using port-based matching
		for _, app := range xmlResponse.Applications {
			for _, instance := range app.Instances {
				log.Printf("[DEBUG] Checking instance %s:%d against service %s:%d", app.Name, instance.Port.Port, service.Name, service.Port)
				// Primary matching: by port (since ports are unique)
				if instance.Port.Port == service.Port {
					log.Printf("[DEBUG] Found %s in Eureka XML by port match - app: %s, status: %s", service.Name, app.Name, instance.Status)

					// Update service health based on Eureka status
					switch strings.ToUpper(instance.Status) {
					case "UP":
						service.HealthStatus = "healthy"
						log.Printf("[DEBUG] Updated %s health status to: healthy (from Eureka)", service.Name)
						return true
					case "DOWN":
						service.HealthStatus = "unhealthy"
						log.Printf("[DEBUG] Updated %s health status to: unhealthy (from Eureka)", service.Name)
						return true
					case "STARTING":
						service.HealthStatus = "starting"
						log.Printf("[DEBUG] Updated %s health status to: starting (from Eureka)", service.Name)
						return true
					case "OUT_OF_SERVICE":
						service.HealthStatus = "unhealthy"
						log.Printf("[DEBUG] Updated %s health status to: unhealthy - out of service (from Eureka)", service.Name)
						return true
					default:
						service.HealthStatus = "unknown"
						log.Printf("[DEBUG] Updated %s health status to: unknown - unknown status '%s' (from Eureka)", service.Name, instance.Status)
						return true
					}
				}
			}
		}
		// Service not found in Eureka XML
		log.Printf("[DEBUG] Service %s (port %d) not found in Eureka XML registry", service.Name, service.Port)
		return false
	} else {
		log.Printf("[DEBUG] Failed to parse Eureka response as XML for %s: %v", service.Name, err)
	}

	// Fallback to JSON parsing
	var jsonResponse EurekaApplications
	if err := json.Unmarshal(bodyBytes, &jsonResponse); err == nil {
		log.Printf("[DEBUG] Successfully parsed Eureka JSON response for %s", service.Name)
		// Look for the service in Eureka registry using port-based matching
		for _, app := range jsonResponse.Applications.Applications {
			for _, instance := range app.Instances {
				log.Printf("[DEBUG] Checking JSON instance %s:%d against service %s:%d", app.Name, instance.Port.Port, service.Name, service.Port)
				// Primary matching: by port (since ports are unique)
				if instance.Port.Port == service.Port {
					log.Printf("[DEBUG] Found %s in Eureka JSON by port match - app: %s, status: %s", service.Name, app.Name, instance.Status)

					// Update service health based on Eureka status
					switch strings.ToUpper(instance.Status) {
					case "UP":
						service.HealthStatus = "healthy"
						log.Printf("[DEBUG] Updated %s health status to: healthy (from Eureka JSON)", service.Name)
						return true
					case "DOWN":
						service.HealthStatus = "unhealthy"
						log.Printf("[DEBUG] Updated %s health status to: unhealthy (from Eureka JSON)", service.Name)
						return true
					case "STARTING":
						service.HealthStatus = "starting"
						log.Printf("[DEBUG] Updated %s health status to: starting (from Eureka JSON)", service.Name)
						return true
					case "OUT_OF_SERVICE":
						service.HealthStatus = "unhealthy"
						log.Printf("[DEBUG] Updated %s health status to: unhealthy - out of service (from Eureka JSON)", service.Name)
						return true
					default:
						service.HealthStatus = "unknown"
						log.Printf("[DEBUG] Updated %s health status to: unknown - unknown status '%s' (from Eureka JSON)", service.Name, instance.Status)
						return true
					}
				}
			}
		}
		// Service not found in Eureka JSON
		log.Printf("[DEBUG] Service %s (port %d) not found in Eureka JSON registry", service.Name, service.Port)
		return false
	}

	// Neither XML nor JSON parsing succeeded
	log.Printf("[DEBUG] Failed to parse Eureka response for %s as either XML or JSON", service.Name)
	return false
}

// checkEurekaServiceRegistration checks if a service is properly registered with Eureka
func (sm *Manager) checkEurekaServiceRegistration(serviceName string) (bool, string) {
	eurekaPort := 8800
	eurekaURL := fmt.Sprintf("http://localhost:%d/eureka/apps", eurekaPort)
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("GET", eurekaURL, nil)
	if err != nil {
		return false, fmt.Sprintf("Failed to create request: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Failed to query Eureka: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, fmt.Sprintf("Eureka returned status: %d", resp.StatusCode)
	}

	var eurekaResponse EurekaApplications
	if err := json.NewDecoder(resp.Body).Decode(&eurekaResponse); err != nil {
		return false, fmt.Sprintf("Failed to decode response: %v", err)
	}

	// Look for the service
	for _, app := range eurekaResponse.Applications.Applications {
		if strings.EqualFold(app.Name, serviceName) ||
			strings.EqualFold(app.Name, strings.ReplaceAll(serviceName, "-", "")) ||
			strings.EqualFold(app.Name, strings.ReplaceAll(serviceName, "vertex-", "")) {

			if len(app.Instances) > 0 {
				return true, fmt.Sprintf("Registered with status: %s", app.Instances[0].Status)
			}
			return true, "Registered but no instances"
		}
	}

	return false, "Not registered in Eureka"
}

// getEurekaServicesStatus returns the status of all services from Eureka
func (sm *Manager) getEurekaServicesStatus() (map[string]string, error) {
	eurekaPort := 8800
	eurekaURL := fmt.Sprintf("http://localhost:%d/eureka/apps", eurekaPort)
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", eurekaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Eureka: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Eureka returned status: %d", resp.StatusCode)
	}

	var eurekaResponse EurekaApplications
	if err := json.NewDecoder(resp.Body).Decode(&eurekaResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	servicesStatus := make(map[string]string)
	for _, app := range eurekaResponse.Applications.Applications {
		if len(app.Instances) > 0 {
			// Use the first instance status (typically there's only one in dev)
			servicesStatus[strings.ToLower(app.Name)] = strings.ToLower(app.Instances[0].Status)
		}
	}

	return servicesStatus, nil
}
