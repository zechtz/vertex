package services

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/zechtz/vertex/internal/models"
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

func (ads *AutoDiscoveryService) ScanDirectory(scanDir string) ([]DiscoveredService, error) {
	log.Printf("[INFO] Starting auto-discovery scan in directory: %s", scanDir)

	var discoveredServices []DiscoveredService

	err := filepath.Walk(scanDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if there are permission errors
		}

		// Look for Maven projects (pom.xml files)
		if info.Name() == "pom.xml" {
			service := ads.analyzeMavenProjectWithScanDir(path, scanDir)
			if service != nil {
				discoveredServices = append(discoveredServices, *service)
			}
		}

		// Look for Gradle projects (build.gradle files)
		if info.Name() == "build.gradle" || info.Name() == "build.gradle.kts" {
			service := ads.analyzeGradleProjectWithScanDir(path, scanDir)
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
	ads.checkExistingServices(discoveredServices)

	log.Printf("[INFO] Auto-discovery completed. Found %d potential services", len(discoveredServices))
	return discoveredServices, nil
}

func (ads *AutoDiscoveryService) analyzeMavenProjectWithScanDir(pomPath, scanDir string) *DiscoveredService {
	projectDir := filepath.Dir(pomPath)
	relativePath, _ := filepath.Rel(scanDir, projectDir)

	// Handle the case where project is in the scan directory itself
	if relativePath == "." {
		relativePath = filepath.Base(projectDir)
	}

	log.Printf("[DEBUG] Maven project analysis: scanDir=%s, projectDir=%s, relativePath=%s", scanDir, projectDir, relativePath)

	// Read pom.xml
	pomContent, err := ioutil.ReadFile(pomPath)
	if err != nil {
		log.Printf("[WARN] Failed to read pom.xml at %s: %v", pomPath, err)
		return nil
	}

	var pom MavenPOM
	if err := xml.Unmarshal(pomContent, &pom); err != nil {
		log.Printf("[WARN] Failed to parse pom.xml at %s: %v", pomPath, err)
		return nil
	}

	// Check if this is a Spring Boot project
	isSpringBoot := ads.isSpringBootProject(pom)
	if !isSpringBoot {
		log.Printf("[DEBUG] Project at %s is not a Spring Boot project, skipping", projectDir)
		return nil
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
	service.Port = ads.extractPortFromProject(projectDir)

	// Determine service type based on dependencies and naming
	service.Type = ads.determineServiceType(pom, service.Name)

	log.Printf("[INFO] Discovered Spring Boot service: %s at %s (port: %d)", service.Name, service.Path, service.Port)
	return service
}

func (ads *AutoDiscoveryService) analyzeGradleProject(buildPath string) *DiscoveredService {
	return ads.analyzeGradleProjectWithScanDir(buildPath, ads.projectDir)
}

func (ads *AutoDiscoveryService) analyzeGradleProjectWithScanDir(buildPath, scanDir string) *DiscoveredService {
	projectDir := filepath.Dir(buildPath)
	relativePath, _ := filepath.Rel(scanDir, projectDir)

	// Handle the case where project is in the scan directory itself
	if relativePath == "." {
		relativePath = filepath.Base(projectDir)
	}

	log.Printf("[DEBUG] Gradle project analysis: scanDir=%s, projectDir=%s, relativePath=%s", scanDir, projectDir, relativePath)

	// Read build.gradle
	buildContent, err := ioutil.ReadFile(buildPath)
	if err != nil {
		log.Printf("[WARN] Failed to read build.gradle at %s: %v", buildPath, err)
		return nil
	}

	content := string(buildContent)

	// Check if this is a Spring Boot project
	if !strings.Contains(content, "spring-boot") {
		log.Printf("[DEBUG] Project at %s is not a Spring Boot project, skipping", projectDir)
		return nil
	}

	// Extract project name from directory or settings.gradle
	projectName := filepath.Base(projectDir)
	if settingsPath := filepath.Join(projectDir, "settings.gradle"); fileExists(settingsPath) {
		if settingsContent, err := ioutil.ReadFile(settingsPath); err == nil {
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
	service.Port = ads.extractPortFromProject(projectDir)
	service.Type = ads.determineServiceType(MavenPOM{ArtifactID: projectName}, service.Name)

	log.Printf("[INFO] Discovered Spring Boot (Gradle) service: %s at %s (port: %d)", service.Name, service.Path, service.Port)
	return service
}

func (ads *AutoDiscoveryService) isSpringBootProject(pom MavenPOM) bool {
	// Check dependencies for Spring Boot starter
	for _, dep := range pom.Dependencies.Dependency {
		if dep.GroupID == "org.springframework.boot" && strings.HasPrefix(dep.ArtifactID, "spring-boot-starter") {
			return true
		}
	}

	// Check for Spring Boot parent
	// Note: This is a simplified check, real implementation might need parent parsing
	return false
}

func (ads *AutoDiscoveryService) extractPortFromProject(projectDir string) int {
	// Check application.properties
	if port := ads.extractPortFromProperties(filepath.Join(projectDir, "src/main/resources/application.properties")); port > 0 {
		return port
	}

	// Check bootstrap.properties
	if port := ads.extractPortFromProperties(filepath.Join(projectDir, "src/main/resources/bootstrap.properties")); port > 0 {
		return port
	}

	// Check application.yml
	if port := ads.extractPortFromYaml(filepath.Join(projectDir, "src/main/resources/application.yml")); port > 0 {
		return port
	}

	// Check application-dev.properties
	if port := ads.extractPortFromProperties(filepath.Join(projectDir, "src/main/resources/application-dev.properties")); port > 0 {
		return port
	}

	// Default port for Spring Boot
	return 8080
}

func (ads *AutoDiscoveryService) extractPortFromProperties(filePath string) int {
	if !fileExists(filePath) {
		return 0
	}

	content, err := ioutil.ReadFile(filePath)
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

func (ads *AutoDiscoveryService) extractPortFromYaml(filePath string) int {
	if !fileExists(filePath) {
		return 0
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return 0
	}

	// Simple YAML port extraction (simplified)
	lines := strings.Split(string(content), "\n")
	var inServerSection bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "server:" {
			inServerSection = true
			continue
		}

		if inServerSection && strings.HasPrefix(line, "port:") {
			portStr := strings.TrimSpace(strings.TrimPrefix(line, "port:"))
			if port, err := strconv.Atoi(portStr); err == nil {
				return port
			}
		}

		// Reset if we hit another top-level section
		if inServerSection && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && line != "" && !strings.HasPrefix(line, "#") {
			inServerSection = false
		}
	}

	return 0
}

func (ads *AutoDiscoveryService) determineServiceType(pom MavenPOM, serviceName string) string {
	name := strings.ToLower(serviceName)
	artifactId := strings.ToLower(pom.ArtifactID)

	// Check for common service types
	if strings.Contains(name, "registry") || strings.Contains(name, "eureka") || strings.Contains(artifactId, "registry") {
		return "registry"
	}
	if strings.Contains(name, "config") || strings.Contains(artifactId, "config") {
		return "config-server"
	}
	if strings.Contains(name, "gateway") || strings.Contains(artifactId, "gateway") {
		return "api-gateway"
	}
	if strings.Contains(name, "auth") || strings.Contains(name, "uaa") || strings.Contains(artifactId, "auth") {
		return "authentication"
	}
	if strings.Contains(name, "cache") || strings.Contains(artifactId, "cache") {
		return "cache"
	}

	return "microservice"
}

func (ads *AutoDiscoveryService) generateServiceName(artifactId, relativePath string) string {
	// Use artifact ID if available, otherwise derive from path
	if artifactId != "" {
		return artifactId
	}

	// Use last directory name from path
	parts := strings.Split(relativePath, string(os.PathSeparator))
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "unknown-service"
}

func (ads *AutoDiscoveryService) checkExistingServices(discoveredServices []DiscoveredService) {
	for i := range discoveredServices {
		// Reset exists flag to ensure fresh check each time
		discoveredServices[i].Exists = false

		// Check for name conflicts
		existingServices := ads.manager.GetServices()
		for _, existing := range existingServices {
			if existing.Name == discoveredServices[i].Name {
				discoveredServices[i].Exists = true
				break
			}
		}

		// If not already flagged as existing, check for path conflicts using system-wide validation
		if !discoveredServices[i].Exists {
			if err := ads.manager.ValidateServiceUniqueness(discoveredServices[i].Name, discoveredServices[i].Path); err != nil {
				discoveredServices[i].Exists = true
			}
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
