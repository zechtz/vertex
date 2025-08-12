package services

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/zechtz/vertex/internal/database"
	"github.com/zechtz/vertex/internal/models"
)

// DockerComposeService handles Docker Compose generation
type DockerComposeService struct {
	db             *database.Database
	serviceManager *Manager
}

// NewDockerComposeService creates a new DockerComposeService
func NewDockerComposeService(db *database.Database, serviceManager *Manager) *DockerComposeService {
	return &DockerComposeService{
		db:             db,
		serviceManager: serviceManager,
	}
}

// GenerateFromProfile generates a Docker Compose configuration from a Vertex profile
func (dcs *DockerComposeService) GenerateFromProfile(profile *models.ServiceProfile, request *models.DockerComposeRequest) (*models.DockerCompose, error) {
	log.Printf("[INFO] Generating Docker Compose for profile: %s", profile.Name)

	compose := &models.DockerCompose{
		Version:  "3.8",
		Services: make(map[string]models.ComposeService),
		Networks: make(map[string]models.ComposeNetwork),
		Volumes:  make(map[string]models.ComposeVolume),
	}

	// Load Docker config for this profile if exists
	dockerConfig, err := dcs.loadDockerConfig(profile.ID)
	if err != nil {
		log.Printf("[WARN] Could not load Docker config for profile %s: %v", profile.ID, err)
		// Continue with default config
		dockerConfig = &models.DockerConfig{
			ProfileID:      profile.ID,
			BaseImages:     make(map[string]string),
			VolumeMappings: make(map[string][]string),
			ResourceLimits: make(map[string]models.ResourceLimit),
		}
	}

	// Create network for this profile
	networkName := dcs.sanitizeName(fmt.Sprintf("vertex-%s", profile.Name))
	compose.Networks[networkName] = models.ComposeNetwork{
		Driver: "bridge",
	}

	// Load service dependencies from database
	allDependencies, err := dcs.db.GetAllServiceDependencies()
	if err != nil {
		log.Printf("[WARN] Could not load service dependencies: %v", err)
		allDependencies = make(map[string][]map[string]any)
	}

	// External services that might be needed
	externalServices := make(map[string]bool)
	
	// Process each service in the profile
	for _, serviceUUID := range profile.Services {
		service, exists := dcs.serviceManager.GetServiceByUUID(serviceUUID)
		if !exists {
			log.Printf("[WARN] Service %s not found, skipping", serviceUUID)
			continue
		}

		composeService, externals := dcs.convertToComposeService(service, profile, dockerConfig, request, networkName, allDependencies)
		compose.Services[dcs.sanitizeName(service.Name)] = composeService

		// Track external services needed
		for external := range externals {
			externalServices[external] = true
		}
	}

	// Add external services if requested
	if request.IncludeExternal {
		dcs.addExternalServices(compose, externalServices, networkName, request.Environment)
	}

	// Add shared volumes that might be needed
	dcs.addSharedVolumes(compose)

	return compose, nil
}

// convertToComposeService converts a Vertex service to a Docker Compose service
func (dcs *DockerComposeService) convertToComposeService(
	service *models.Service,
	profile *models.ServiceProfile,
	dockerConfig *models.DockerConfig,
	request *models.DockerComposeRequest,
	networkName string,
	allDependencies map[string][]map[string]any,
) (models.ComposeService, map[string]bool) {
	
	composeService := models.ComposeService{
		Networks: []string{networkName},
		Restart:  "unless-stopped",
	}

	externalServices := make(map[string]bool)

	// Determine build context or image
	if baseImage, hasCustomImage := dockerConfig.BaseImages[service.ID]; hasCustomImage {
		composeService.Image = baseImage
	} else {
		// Use build context - assume service directory contains Dockerfile
		projectsDir := profile.ProjectsDir
		if projectsDir == "" {
			projectsDir = dcs.serviceManager.GetConfig().ProjectsDir
		}
		
		buildContext := filepath.Join(".", service.Dir)
		if projectsDir != "" {
			buildContext = filepath.Join(projectsDir, service.Dir)
		}
		
		composeService.Build = &models.ComposeBuild{
			Context: buildContext,
		}
		
		// Auto-detect Dockerfile based on build system
		buildSystem := dcs.detectBuildSystem(service)
		dockerfile := dcs.generateDockerfileForBuildSystem(buildSystem)
		if dockerfile != "" {
			composeService.Build.Dockerfile = dockerfile
		}
	}

	// Port mapping
	if service.Port > 0 {
		// Find an available external port starting from service port
		externalPort := dcs.findAvailablePort(service.Port, profile.Services)
		composeService.Ports = []string{fmt.Sprintf("%d:%d", externalPort, service.Port)}
	}

	// Environment variables
	envVars := dcs.buildEnvironmentVariables(service, profile, externalServices)
	if len(envVars) > 0 {
		composeService.Environment = envVars
	}

	// Dependencies
	if deps, hasDeps := allDependencies[service.ID]; hasDeps {
		dependsOn := dcs.buildDependsOn(deps)
		if len(dependsOn) > 0 {
			composeService.DependsOn = dependsOn
		}
	}

	// Volume mappings
	if volumeMappings, hasVolumes := dockerConfig.VolumeMappings[service.ID]; hasVolumes {
		composeService.Volumes = volumeMappings
	} else {
		// Default volume mappings for development
		if request.Environment == "development" {
			composeService.Volumes = []string{
				fmt.Sprintf("./%s:/app", service.Dir),
			}
		}
	}

	// Working directory
	composeService.WorkingDir = "/app"

	// Health check
	composeService.HealthCheck = dcs.generateHealthCheck(service)

	// Resource limits
	if limits, hasLimits := dockerConfig.ResourceLimits[service.ID]; hasLimits {
		composeService.Deploy = &models.ComposeDeploy{
			Resources: &models.ComposeResources{
				Limits: &models.ComposeResourceLimits{
					CPU:    limits.CPULimit,
					Memory: limits.MemoryLimit,
				},
				Reservations: &models.ComposeResourceLimits{
					CPU:    limits.CPUReserve,
					Memory: limits.MemoryReserve,
				},
			},
		}
	}

	// Labels for identification
	composeService.Labels = map[string]string{
		"vertex.service.id":   service.ID,
		"vertex.service.name": service.Name,
		"vertex.profile.id":   profile.ID,
		"vertex.profile.name": profile.Name,
	}

	return composeService, externalServices
}

// buildEnvironmentVariables builds the environment variables for a service
func (dcs *DockerComposeService) buildEnvironmentVariables(service *models.Service, profile *models.ServiceProfile, externalServices map[string]bool) []string {
	// Use a map to manage env vars with service-specific taking precedence
	envMap := make(map[string]string)

	log.Printf("[DEBUG] Building env vars for service %s. Service has %d env vars, Profile has %d env vars", 
		service.Name, len(service.EnvVars), len(profile.EnvVars))

	// First, add profile/global environment variables as base
	for key, envValue := range profile.EnvVars {
		envMap[key] = envValue
		dcs.detectExternalServiceFromEnvVar(key, envValue, externalServices)
		log.Printf("[DEBUG] Added profile env var: %s=%s", key, envValue)
	}

	// Then add service-specific environment variables (these override global ones)
	for key, envVar := range service.EnvVars {
		envMap[key] = envVar.Value  // This will override any profile-level env var with same key
		dcs.detectExternalServiceFromEnvVar(key, envVar.Value, externalServices)
		log.Printf("[DEBUG] Added service-specific env var: %s=%s (overrides global if exists)", key, envVar.Value)
	}

	// Convert map to slice
	var envVars []string
	for key, value := range envMap {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	log.Printf("[DEBUG] Final env vars for service %s: %d total", service.Name, len(envVars))

	return envVars
}

// detectExternalServiceFromEnvVar detects if an environment variable references external services
func (dcs *DockerComposeService) detectExternalServiceFromEnvVar(key, value string, externalServices map[string]bool) {
	key = strings.ToUpper(key)
	value = strings.ToLower(value)

	// Database detection
	if strings.Contains(key, "DATABASE") || strings.Contains(key, "DB") {
		if strings.Contains(value, "postgres") {
			externalServices["postgres"] = true
		} else if strings.Contains(value, "mysql") {
			externalServices["mysql"] = true
		} else if strings.Contains(value, "mongo") {
			externalServices["mongodb"] = true
		}
	}

	// Redis detection
	if strings.Contains(key, "REDIS") || strings.Contains(key, "CACHE") {
		if strings.Contains(value, "redis") {
			externalServices["redis"] = true
		}
	}

	// Message queue detection
	if strings.Contains(key, "RABBITMQ") || strings.Contains(key, "AMQP") {
		externalServices["rabbitmq"] = true
	}

	if strings.Contains(key, "KAFKA") {
		externalServices["kafka"] = true
	}
}

// buildDependsOn builds the depends_on configuration from service dependencies
func (dcs *DockerComposeService) buildDependsOn(dependencies []map[string]any) []string {
	var dependsOn []string

	for _, dep := range dependencies {
		if serviceName, ok := dep["serviceName"].(string); ok {
			dependsOn = append(dependsOn, dcs.sanitizeName(serviceName))
		}
	}

	return dependsOn
}

// addExternalServices adds commonly needed external services
func (dcs *DockerComposeService) addExternalServices(compose *models.DockerCompose, externalServices map[string]bool, networkName, environment string) {
	// PostgreSQL
	if externalServices["postgres"] {
		compose.Services["postgres"] = models.ComposeService{
			Image:    "postgres:13-alpine",
			Networks: []string{networkName},
			Environment: []string{
				"POSTGRES_DB=vertex_dev",
				"POSTGRES_USER=vertex",
				"POSTGRES_PASSWORD=vertex_password",
			},
			Volumes: []string{"postgres_data:/var/lib/postgresql/data"},
			Restart: "unless-stopped",
		}
		compose.Volumes["postgres_data"] = models.ComposeVolume{}
	}

	// Redis
	if externalServices["redis"] {
		compose.Services["redis"] = models.ComposeService{
			Image:    "redis:7-alpine",
			Networks: []string{networkName},
			Volumes:  []string{"redis_data:/data"},
			Restart:  "unless-stopped",
		}
		compose.Volumes["redis_data"] = models.ComposeVolume{}
	}

	// MySQL
	if externalServices["mysql"] {
		compose.Services["mysql"] = models.ComposeService{
			Image:    "mysql:8.0",
			Networks: []string{networkName},
			Environment: []string{
				"MYSQL_ROOT_PASSWORD=root_password",
				"MYSQL_DATABASE=vertex_dev",
				"MYSQL_USER=vertex",
				"MYSQL_PASSWORD=vertex_password",
			},
			Volumes: []string{"mysql_data:/var/lib/mysql"},
			Restart: "unless-stopped",
		}
		compose.Volumes["mysql_data"] = models.ComposeVolume{}
	}

	// MongoDB
	if externalServices["mongodb"] {
		compose.Services["mongodb"] = models.ComposeService{
			Image:    "mongo:5.0",
			Networks: []string{networkName},
			Environment: []string{
				"MONGO_INITDB_ROOT_USERNAME=vertex",
				"MONGO_INITDB_ROOT_PASSWORD=vertex_password",
			},
			Volumes: []string{"mongodb_data:/data/db"},
			Restart: "unless-stopped",
		}
		compose.Volumes["mongodb_data"] = models.ComposeVolume{}
	}

	// RabbitMQ
	if externalServices["rabbitmq"] {
		compose.Services["rabbitmq"] = models.ComposeService{
			Image:    "rabbitmq:3-management-alpine",
			Networks: []string{networkName},
			Environment: []string{
				"RABBITMQ_DEFAULT_USER=vertex",
				"RABBITMQ_DEFAULT_PASS=vertex_password",
			},
			Ports:   []string{"15672:15672", "5672:5672"},
			Volumes: []string{"rabbitmq_data:/var/lib/rabbitmq"},
			Restart: "unless-stopped",
		}
		compose.Volumes["rabbitmq_data"] = models.ComposeVolume{}
	}
}

// addSharedVolumes adds commonly needed shared volumes
func (dcs *DockerComposeService) addSharedVolumes(compose *models.DockerCompose) {
	// Add logs volume for shared logging
	if _, exists := compose.Volumes["vertex_logs"]; !exists {
		compose.Volumes["vertex_logs"] = models.ComposeVolume{}
	}
}

// generateHealthCheck generates a health check configuration for a service
func (dcs *DockerComposeService) generateHealthCheck(service *models.Service) *models.ComposeHealthCheck {
	if service.Port <= 0 {
		return nil
	}

	// Default HTTP health check
	return &models.ComposeHealthCheck{
		Test:        []string{"CMD", "curl", "-f", fmt.Sprintf("http://localhost:%d/health", service.Port)},
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		Retries:     3,
		StartPeriod: 40 * time.Second,
	}
}

// detectBuildSystem detects the build system for a service
func (dcs *DockerComposeService) detectBuildSystem(service *models.Service) string {
	if service.BuildSystem != "" && service.BuildSystem != "auto" {
		return service.BuildSystem
	}

	// This would typically scan the service directory for build files
	// For now, return the configured build system or default to maven
	return "maven"
}

// generateDockerfileForBuildSystem generates appropriate Dockerfile content based on build system
func (dcs *DockerComposeService) generateDockerfileForBuildSystem(buildSystem string) string {
	switch buildSystem {
	case "maven":
		return "Dockerfile.maven"
	case "gradle":
		return "Dockerfile.gradle"
	case "nodejs":
		return "Dockerfile.nodejs"
	default:
		return ""
	}
}

// findAvailablePort finds an available external port for a service
func (dcs *DockerComposeService) findAvailablePort(preferredPort int, serviceUUIDs []string) int {
	usedPorts := make(map[int]bool)

	// Collect ports already used by other services in the profile
	for _, uuid := range serviceUUIDs {
		if service, exists := dcs.serviceManager.GetServiceByUUID(uuid); exists {
			usedPorts[service.Port] = true
		}
	}

	// Start with preferred port and find next available
	port := preferredPort
	for usedPorts[port] && port < 65535 {
		port++
	}

	return port
}

// sanitizeName sanitizes a name to be Docker Compose compatible
func (dcs *DockerComposeService) sanitizeName(name string) string {
	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9-_]`)
	sanitized := reg.ReplaceAllString(name, "-")
	
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	sanitized = reg.ReplaceAllString(sanitized, "-")
	
	// Trim hyphens from start and end
	sanitized = strings.Trim(sanitized, "-")
	
	// Ensure it starts with a letter or number
	if len(sanitized) > 0 && !regexp.MustCompile(`^[a-zA-Z0-9]`).MatchString(sanitized) {
		sanitized = "service-" + sanitized
	}
	
	return strings.ToLower(sanitized)
}

// loadDockerConfig loads Docker configuration for a profile
func (dcs *DockerComposeService) loadDockerConfig(profileID string) (*models.DockerConfig, error) {
	return dcs.db.GetDockerConfig(profileID)
}

// SaveDockerConfig saves Docker configuration for a profile
func (dcs *DockerComposeService) SaveDockerConfig(config *models.DockerConfig) error {
	log.Printf("[INFO] Saving Docker config for profile: %s", config.ProfileID)
	return dcs.db.SaveDockerConfig(config)
}

// GetDockerConfig retrieves Docker configuration for a profile
func (dcs *DockerComposeService) GetDockerConfig(profileID string) (*models.DockerConfig, error) {
	return dcs.db.GetDockerConfig(profileID)
}

// DeleteDockerConfig deletes Docker configuration for a profile
func (dcs *DockerComposeService) DeleteDockerConfig(profileID string) error {
	return dcs.db.DeleteDockerConfig(profileID)
}

// GenerateOverrideFile generates a docker-compose.override.yml for development
func (dcs *DockerComposeService) GenerateOverrideFile(profile *models.ServiceProfile) (*models.DockerCompose, error) {
	compose := &models.DockerCompose{
		Version:  "3.8",
		Services: make(map[string]models.ComposeService),
	}

	// Add development-specific overrides
	for _, serviceUUID := range profile.Services {
		service, exists := dcs.serviceManager.GetServiceByUUID(serviceUUID)
		if !exists {
			continue
		}

		serviceName := dcs.sanitizeName(service.Name)
		composeService := models.ComposeService{
			Volumes: []string{
				fmt.Sprintf("./%s:/app", service.Dir),
			},
			Environment: []string{
				"NODE_ENV=development",
				"SPRING_PROFILES_ACTIVE=development",
				"LOG_LEVEL=debug",
			},
		}

		compose.Services[serviceName] = composeService
	}

	return compose, nil
}