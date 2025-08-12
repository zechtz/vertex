// Package services provides functionality for auto-discovery of microservices
// in a project directory.
package services

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/zechtz/vertex/internal/models"
	"gopkg.in/yaml.v3"
)

type AutoDiscoveryService struct {
	manager    *Manager
	projectDir string
}

type DiscoveredService struct {
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Port        int               `json:"port"`
	Type        string            `json:"type"`
	Framework   string            `json:"framework"`
	Description string            `json:"description"`
	Properties  map[string]string `json:"properties"`
	IsValid     bool              `json:"isValid"`
	Exists      bool              `json:"exists"`
}

type MavenPOM struct {
	XMLName      xml.Name `xml:"project"`
	GroupID      string   `xml:"groupId"`
	ArtifactID   string   `xml:"artifactId"`
	Version      string   `xml:"version"`
	Name         string   `xml:"name"`
	Description  string   `xml:"description"`
	Dependencies struct {
		Dependency []struct {
			GroupID    string `xml:"groupId"`
			ArtifactID string `xml:"artifactId"`
			Version    string `xml:"version"`
		} `xml:"dependency"`
	} `xml:"dependencies"`
	Properties struct {
		JavaVersion    string `xml:"java.version"`
		SpringVersion  string `xml:"spring-boot.version"`
		ProjectVersion string `xml:"project.version"`
	} `xml:"properties"`
	Parent struct {
		GroupID    string `xml:"groupId"`
		ArtifactID string `xml:"artifactId"`
		Version    string `xml:"version"`
	} `xml:"parent"`
}

func NewAutoDiscoveryService(manager *Manager) *AutoDiscoveryService {
	return &AutoDiscoveryService{
		manager:    manager,
		projectDir: manager.config.ProjectsDir,
	}
}

func (ads *AutoDiscoveryService) ScanProjectDirectory() ([]DiscoveredService, error) {
	return ads.ScanDirectory(ads.projectDir)
}

// ScanDirectory scans the specified directory for Maven and Gradle projects, returning discovered services.
func (ads *AutoDiscoveryService) ScanDirectory(scanDir string) ([]DiscoveredService, error) {
	if scanDir == "" {
		return nil, fmt.Errorf("scan directory cannot be empty")
	}
	if _, err := os.Stat(scanDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("scan directory does not exist: %s", scanDir)
	}

	log.Printf("[INFO] Starting auto-discovery scan in directory: %s", scanDir)

	var discoveredServices []DiscoveredService

	err := filepath.Walk(scanDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("[WARN] Error accessing path %s: %v", path, err)
			return nil // Continue walking
		}

		// Look for Maven projects (pom.xml files)
		if info.Name() == "pom.xml" {
			service, err := ads.analyzeMavenProjectWithScanDir(path, scanDir)
			if err != nil {
				log.Printf("[WARN] Failed to analyze Maven project at %s: %v", path, err)
			}
			if service != nil {
				discoveredServices = append(discoveredServices, *service)
			}
		}

		// Look for Gradle projects (build.gradle files)
		if info.Name() == "build.gradle" || info.Name() == "build.gradle.kts" {
			service, err := ads.analyzeGradleProjectWithScanDir(path, scanDir)
			if err != nil {
				log.Printf("[WARN] Failed to analyze Gradle project at %s: %v", path, err)
			}
			if service != nil {
				discoveredServices = append(discoveredServices, *service)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking project directory: %w", err)
	}

	// Check if discovered services already exist in the system
	ads.checkExistingServices(&discoveredServices)

	log.Printf("[INFO] Auto-discovery completed. Found %d potential services", len(discoveredServices))
	return discoveredServices, nil
}

// ScanDirectoryForProfile scans for services with profile-aware existence checking
func (ads *AutoDiscoveryService) ScanDirectoryForProfile(scanDir string, profileServices []string) ([]DiscoveredService, error) {
	if scanDir == "" {
		return nil, fmt.Errorf("scan directory cannot be empty")
	}
	if _, err := os.Stat(scanDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("scan directory does not exist: %s", scanDir)
	}

	log.Printf("[INFO] Starting auto-discovery scan for profile in directory: %s", scanDir)

	var discoveredServices []DiscoveredService

	err := filepath.Walk(scanDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("[WARN] Error accessing path %s: %v", path, err)
			return nil
		}

		// Skip hidden directories and files
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Look for Maven projects (pom.xml files)
		if info.Name() == "pom.xml" {
			service, err := ads.analyzeMavenProjectWithScanDir(path, scanDir)
			if err != nil {
				log.Printf("[WARN] Failed to analyze Maven project at %s: %v", path, err)
			}
			if service != nil {
				discoveredServices = append(discoveredServices, *service)
			}
		}

		// Look for Gradle projects (build.gradle files)
		if info.Name() == "build.gradle" || info.Name() == "build.gradle.kts" {
			service, err := ads.analyzeGradleProjectWithScanDir(path, scanDir)
			if err != nil {
				log.Printf("[WARN] Failed to analyze Gradle project at %s: %v", path, err)
			}
			if service != nil {
				discoveredServices = append(discoveredServices, *service)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking project directory: %w", err)
	}

	// Check if discovered services already exist in the specific profile
	ads.checkExistingServicesForProfile(&discoveredServices, profileServices)

	log.Printf("[INFO] Profile-aware auto-discovery completed. Found %d potential services", len(discoveredServices))
	return discoveredServices, nil
}

func (ads *AutoDiscoveryService) analyzeMavenProjectWithScanDir(pomPath, scanDir string) (*DiscoveredService, error) {
	projectDir := filepath.Dir(pomPath)
	relativePath, err := filepath.Rel(scanDir, projectDir)
	if err != nil {
		log.Printf("[WARN] Failed to compute relative path for %s: %v", projectDir, err)
		return nil, fmt.Errorf("failed to compute relative path: %w", err)
	}

	// Handle the case where project is in the scan directory itself
	if relativePath == "." {
		relativePath = filepath.Base(projectDir)
	}

	log.Printf("[DEBUG] Maven project analysis: scanDir=%s, projectDir=%s, relativePath=%s", scanDir, projectDir, relativePath)

	// Read pom.xml
	pomContent, err := os.ReadFile(pomPath)
	if err != nil {
		log.Printf("[WARN] Failed to read pom.xml at %s: %v", pomPath, err)
		return nil, fmt.Errorf("failed to read pom.xml: %w", err)
	}

	var pom MavenPOM
	if err := xml.Unmarshal(pomContent, &pom); err != nil {
		log.Printf("[WARN] Failed to parse pom.xml at %s: %v", pomPath, err)
		return nil, fmt.Errorf("failed to parse pom.xml: %w", err)
	}

	// Check if this is a Spring Boot project
	isSpringBoot := ads.isSpringBootProject(pom)
	if !isSpringBoot {
		log.Printf("[DEBUG] Project at %s is not a Spring Boot project, skipping", projectDir)
		return nil, nil
	}

	service := &DiscoveredService{
		Name:        ads.generateServiceName(pom.ArtifactID, relativePath),
		Path:        relativePath,
		Type:        "microservice",
		Framework:   "Spring Boot",
		Description: pom.Description,
		Properties:  make(map[string]string),
		IsValid:     true,
	}

	// Extract additional information
	service.Properties["groupId"] = pom.GroupID
	service.Properties["artifactId"] = pom.ArtifactID
	service.Properties["version"] = pom.Version
	service.Properties["name"] = pom.Name

	// Try to determine the port
	service.Port, err = ads.extractPortFromProject(projectDir)
	if err != nil {
		log.Printf("[DEBUG] No port found for project at %s, using default: %v", projectDir, err)
		service.Port = 8080 // Fallback to default
	}

	// Determine service type based on dependencies and naming
	service.Type = ads.determineServiceType(pom, service.Name)

	log.Printf("[INFO] Discovered Spring Boot service: %s at %s (port: %d)", service.Name, service.Path, service.Port)
	return service, nil
}

func (ads *AutoDiscoveryService) analyzeGradleProjectWithScanDir(buildPath, scanDir string) (*DiscoveredService, error) {
	projectDir := filepath.Dir(buildPath)
	relativePath, err := filepath.Rel(scanDir, projectDir)
	if err != nil {
		log.Printf("[WARN] Failed to compute relative path for %s: %v", projectDir, err)
		return nil, fmt.Errorf("failed to compute relative path: %w", err)
	}

	// Handle the case where project is in the scan directory itself
	if relativePath == "." {
		relativePath = filepath.Base(projectDir)
	}

	log.Printf("[DEBUG] Gradle project analysis: scanDir=%s, projectDir=%s, relativePath=%s", scanDir, projectDir, relativePath)

	// Read build.gradle
	buildContent, err := os.ReadFile(buildPath)
	if err != nil {
		log.Printf("[WARN] Failed to read build.gradle at %s: %v", buildPath, err)
		return nil, fmt.Errorf("failed to read build.gradle: %w", err)
	}

	content := string(buildContent)

	// Check if this is a Spring Boot project
	if !strings.Contains(content, "spring-boot") {
		log.Printf("[DEBUG] Project at %s is not a Spring Boot project, skipping", projectDir)
		return nil, nil
	}

	// Extract project name from directory or settings.gradle
	projectName := filepath.Base(projectDir)
	if settingsPath := filepath.Join(projectDir, "settings.gradle"); fileExists(settingsPath) {
		settingsContent, err := os.ReadFile(settingsPath)
		if err == nil {
			if name := extractGradleProjectName(string(settingsContent)); name != "" {
				projectName = name
			}
		}
	}

	service := &DiscoveredService{
		Name:       ads.generateServiceName(projectName, relativePath),
		Path:       relativePath,
		Type:       "microservice",
		Framework:  "Spring Boot (Gradle)",
		Properties: make(map[string]string),
		IsValid:    true,
	}

	service.Properties["projectName"] = projectName
	service.Port, err = ads.extractPortFromProject(projectDir)
	if err != nil {
		log.Printf("[DEBUG] No port found for project at %s, using default: %v", projectDir, err)
		service.Port = 8080 // Fallback to default
	}
	service.Type = ads.determineServiceType(MavenPOM{ArtifactID: projectName}, service.Name)

	log.Printf("[INFO] Discovered Spring Boot (Gradle) service: %s at %s (port: %d)", service.Name, service.Path, service.Port)
	return service, nil
}

func (ads *AutoDiscoveryService) isSpringBootProject(pom MavenPOM) bool {
	// Check for Spring Boot parent
	if pom.Parent.GroupID == "org.springframework.boot" && pom.Parent.ArtifactID == "spring-boot-starter-parent" {
		return true
	}
	// Check dependencies for Spring Boot starter
	for _, dep := range pom.Dependencies.Dependency {
		if dep.GroupID == "org.springframework.boot" && strings.HasPrefix(dep.ArtifactID, "spring-boot-starter") {
			return true
		}
	}
	return false
}

func (ads *AutoDiscoveryService) extractPortFromProject(projectDir string) (int, error) {
	// Check application.properties
	if port := ads.extractPortFromProperties(filepath.Join(projectDir, "src/main/resources/application.properties")); port > 0 {
		return port, nil
	}

	// Check bootstrap.properties
	if port := ads.extractPortFromProperties(filepath.Join(projectDir, "src/main/resources/bootstrap.properties")); port > 0 {
		return port, nil
	}

	// Check application.yml
	port, err := ads.extractPortFromYaml(filepath.Join(projectDir, "src/main/resources/application.yml"))
	if err == nil && port > 0 {
		return port, nil
	}

	// Check application-dev.properties
	if port := ads.extractPortFromProperties(filepath.Join(projectDir, "src/main/resources/application-dev.properties")); port > 0 {
		return port, nil
	}

	// No port found
	return 0, fmt.Errorf("no port found in configuration files")
}

func (ads *AutoDiscoveryService) extractPortFromProperties(filePath string) int {
	if !fileExists(filePath) {
		return 0
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0
	}

	// Look for server.port property
	portRegex := regexp.MustCompile(`server\.port\s*=\s*(\d+)`)
	matches := portRegex.FindStringSubmatch(string(content))
	if len(matches) > 1 {
		if port, err := strconv.Atoi(matches[1]); err == nil {
			return port
		}
	}

	return 0
}

func (ads *AutoDiscoveryService) extractPortFromYaml(filePath string) (int, error) {
	if !fileExists(filePath) {
		return 0, nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read YAML file: %w", err)
	}

	var config struct {
		Server struct {
			Port int `yaml:"port"`
		} `yaml:"server"`
	}
	if err := yaml.Unmarshal(content, &config); err != nil {
		return 0, fmt.Errorf("failed to parse YAML: %w", err)
	}
	if config.Server.Port > 0 {
		return config.Server.Port, nil
	}
	return 0, nil
}

func (ads *AutoDiscoveryService) determineServiceType(pom MavenPOM, serviceName string) string {
	name := strings.ToLower(serviceName)
	artifactID := strings.ToLower(pom.ArtifactID)

	// Check for common service types
	if strings.Contains(name, "registry") || strings.Contains(name, "eureka") || strings.Contains(artifactID, "registry") {
		return "registry"
	}
	if strings.Contains(name, "config") || strings.Contains(artifactID, "config") {
		return "config-server"
	}
	if strings.Contains(name, "gateway") || strings.Contains(artifactID, "gateway") {
		return "api-gateway"
	}
	if strings.Contains(name, "auth") || strings.Contains(name, "uaa") || strings.Contains(artifactID, "auth") {
		return "authentication"
	}
	if strings.Contains(name, "cache") || strings.Contains(artifactID, "cache") {
		return "cache"
	}

	return "microservice"
}

func (ads *AutoDiscoveryService) generateServiceName(artifactID, relativePath string) string {
	// Use artifact ID if available, otherwise derive from path
	if artifactID != "" {
		return artifactID
	}

	// Use last directory name from path
	parts := strings.Split(relativePath, string(os.PathSeparator))
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "unknown-service"
}

func (ads *AutoDiscoveryService) checkExistingServices(discoveredServices *[]DiscoveredService) {
	for i := range *discoveredServices {
		// Reset exists flag to ensure fresh check each time
		(*discoveredServices)[i].Exists = false

		// Check for name conflicts
		existingServices := ads.manager.GetServices()
		for _, existing := range existingServices {
			if existing.Name == (*discoveredServices)[i].Name {
				(*discoveredServices)[i].Exists = true
				break
			}
		}

		// If not already flagged as existing, check for path conflicts using system-wide validation
		if !(*discoveredServices)[i].Exists {
			if err := ads.manager.ValidateServiceUniqueness((*discoveredServices)[i].Name, (*discoveredServices)[i].Path); err != nil {
				(*discoveredServices)[i].Exists = true
			}
		}
	}
}

// checkExistingServicesForProfile checks if discovered services exist in a specific profile
func (ads *AutoDiscoveryService) checkExistingServicesForProfile(discoveredServices *[]DiscoveredService, profileServices []string) {
	// Create a map of services in the profile for quick lookup
	profileServiceMap := make(map[string]bool)

	// Get all global services to check paths
	allServices := ads.manager.GetServices()
	globalServicePathMap := make(map[string]models.Service)

	for _, service := range allServices {
		normalizedPath := filepath.ToSlash(filepath.Clean(service.Dir))
		globalServicePathMap[normalizedPath] = service
	}

	// Build profile service map using UUIDs
	for _, serviceUUID := range profileServices {
		profileServiceMap[serviceUUID] = true
	}

	for i := range *discoveredServices {
		(*discoveredServices)[i].Exists = false

		// Check if a service with this path exists globally
		normalizedDiscoveredPath := filepath.ToSlash(filepath.Clean((*discoveredServices)[i].Path))
		var matchingGlobalService *models.Service
		if globalService, exists := globalServicePathMap[normalizedDiscoveredPath]; exists {
			matchingGlobalService = &globalService
		}

		// If service exists globally, check if it's in this profile
		if matchingGlobalService != nil {
			if profileServiceMap[matchingGlobalService.ID] {
				// Service exists in both global and profile
				(*discoveredServices)[i].Exists = true
			}
			// If service exists globally but not in profile, leave Exists = false
			// This allows it to be "re-imported" to the profile
		}
	}
}

func (ads *AutoDiscoveryService) CreateServiceFromDiscovered(discovered DiscoveredService) (*models.Service, error) {
	if discovered.Exists {
		return nil, fmt.Errorf("service %s already exists", discovered.Name)
	}

	// Determine next order
	nextOrder := ads.getNextServiceOrder()

	service := &models.Service{
		ID:           uuid.New().String(),
		Name:         discovered.Name,
		Dir:          discovered.Path,
		Port:         discovered.Port,
		Description:  discovered.Description,
		Order:        nextOrder,
		IsEnabled:    true,
		Status:       "stopped",
		HealthStatus: "unknown",
		EnvVars:      make(map[string]models.EnvVar),
		Logs:         []models.LogEntry{},
	}

	// Set default health URL
	if service.Port > 0 {
		service.HealthURL = fmt.Sprintf("http://localhost:%d/actuator/health", service.Port)
	}

	// Add the service to the manager
	return service, ads.manager.AddService(service)
}

func (ads *AutoDiscoveryService) getNextServiceOrder() int {
	services := ads.manager.GetServices()
	maxOrder := 0

	for _, service := range services {
		if service.Order > maxOrder {
			maxOrder = service.Order
		}
	}

	return maxOrder + 1
}

// Helper functions
func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

func extractGradleProjectName(content string) string {
	// Extract project name from settings.gradle
	nameRegex := regexp.MustCompile(`rootProject\.name\s*=\s*['"]([^'"]+)['"]`)
	matches := nameRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
