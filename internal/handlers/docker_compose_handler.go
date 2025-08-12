package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zechtz/vertex/internal/models"
	"github.com/zechtz/vertex/internal/services"
)

// dockerComposeHandler handles Docker Compose generation requests
type dockerComposeHandler struct {
	dockerComposeService *services.DockerComposeService
	profileService       *services.ProfileService
	authService          *services.AuthService
}

func registerDockerComposeRoutes(h *Handler, r *mux.Router) {
	// Create Docker Compose service instance
	dockerService := services.NewDockerComposeService(h.serviceManager.GetDatabase(), h.serviceManager)
	
	dockerHandler := &dockerComposeHandler{
		dockerComposeService: dockerService,
		profileService:       h.profileService,
		authService:          h.authService,
	}

	// Docker Compose generation routes
	r.HandleFunc("/api/profiles/{id}/docker-compose", dockerHandler.generateDockerComposeHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}/docker-compose/download", dockerHandler.downloadDockerComposeHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}/docker-compose/preview", dockerHandler.previewDockerComposeHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}/docker-compose/override", dockerHandler.generateOverrideHandler).Methods("GET")
	
	// Docker configuration management
	r.HandleFunc("/api/profiles/{id}/docker-config", dockerHandler.getDockerConfigHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}/docker-config", dockerHandler.updateDockerConfigHandler).Methods("PUT")
	r.HandleFunc("/api/profiles/{id}/docker-config", dockerHandler.deleteDockerConfigHandler).Methods("DELETE")
}

// generateDockerComposeHandler generates Docker Compose YAML and returns it as JSON
func (dh *dockerComposeHandler) generateDockerComposeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, dh.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	// Get profile
	profile, err := dh.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get service profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get service profile", http.StatusInternalServerError)
		}
		return
	}

	// Parse query parameters for generation options
	request := &models.DockerComposeRequest{
		Environment:      r.URL.Query().Get("environment"),
		IncludeExternal:  r.URL.Query().Get("includeExternal") == "true",
		GenerateOverride: r.URL.Query().Get("generateOverride") == "true",
		CustomOverrides:  make(map[string]interface{}),
	}

	// Set default environment if not specified
	if request.Environment == "" {
		request.Environment = "development"
	}

	log.Printf("[INFO] Generating Docker Compose for profile '%s' (environment: %s)", profile.Name, request.Environment)

	// Generate Docker Compose
	compose, err := dh.dockerComposeService.GenerateFromProfile(profile, request)
	if err != nil {
		log.Printf("[ERROR] Failed to generate Docker Compose: %v", err)
		http.Error(w, fmt.Sprintf("Failed to generate Docker Compose: %v", err), http.StatusInternalServerError)
		return
	}

	// Validate the generated compose file
	if err := compose.Validate(); err != nil {
		log.Printf("[ERROR] Generated Docker Compose is invalid: %v", err)
		http.Error(w, fmt.Sprintf("Generated Docker Compose is invalid: %v", err), http.StatusInternalServerError)
		return
	}

	// Return as JSON with both YAML content and metadata
	response := map[string]interface{}{
		"profileId":     profile.ID,
		"profileName":   profile.Name,
		"environment":   request.Environment,
		"yaml":          compose.ToYAML(),
		"serviceCount":  len(compose.Services),
		"networkCount":  len(compose.Networks),
		"volumeCount":   len(compose.Volumes),
		"services":      getServiceNames(compose.Services),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// downloadDockerComposeHandler generates and returns Docker Compose file for download
func (dh *dockerComposeHandler) downloadDockerComposeHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := extractClaimsFromRequest(r, dh.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	// Get profile
	profile, err := dh.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get service profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get service profile", http.StatusInternalServerError)
		}
		return
	}

	// Parse query parameters
	request := &models.DockerComposeRequest{
		Environment:     r.URL.Query().Get("environment"),
		IncludeExternal: r.URL.Query().Get("includeExternal") == "true",
		CustomOverrides: make(map[string]interface{}),
	}

	if request.Environment == "" {
		request.Environment = "development"
	}

	// Generate Docker Compose
	compose, err := dh.dockerComposeService.GenerateFromProfile(profile, request)
	if err != nil {
		log.Printf("[ERROR] Failed to generate Docker Compose: %v", err)
		http.Error(w, fmt.Sprintf("Failed to generate Docker Compose: %v", err), http.StatusInternalServerError)
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("docker-compose-%s.yml", sanitizeFilename(profile.Name))
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Write YAML content
	yamlContent := compose.ToYAML()
	if _, err := w.Write([]byte(yamlContent)); err != nil {
		log.Printf("[ERROR] Failed to write YAML content: %v", err)
		return
	}

	log.Printf("[INFO] Docker Compose file downloaded for profile '%s'", profile.Name)
}

// previewDockerComposeHandler returns a preview of the Docker Compose without generating the full file
func (dh *dockerComposeHandler) previewDockerComposeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, dh.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	// Get profile
	profile, err := dh.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get service profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get service profile", http.StatusInternalServerError)
		}
		return
	}

	// Generate preview without external services for faster response
	request := &models.DockerComposeRequest{
		Environment:     r.URL.Query().Get("environment"),
		IncludeExternal: false, // Don't include external services in preview
		CustomOverrides: make(map[string]interface{}),
	}

	if request.Environment == "" {
		request.Environment = "development"
	}

	// Generate Docker Compose
	compose, err := dh.dockerComposeService.GenerateFromProfile(profile, request)
	if err != nil {
		log.Printf("[ERROR] Failed to generate Docker Compose preview: %v", err)
		http.Error(w, fmt.Sprintf("Failed to generate Docker Compose preview: %v", err), http.StatusInternalServerError)
		return
	}

	// Create preview response with summary information
	services := make([]map[string]interface{}, 0, len(compose.Services))
	for name, service := range compose.Services {
		serviceInfo := map[string]interface{}{
			"name":         name,
			"image":        service.Image,
			"ports":        service.Ports,
			"environment":  len(service.Environment),
			"volumes":      len(service.Volumes),
			"dependencies": service.DependsOn,
		}
		if service.Build != nil {
			serviceInfo["buildContext"] = service.Build.Context
		}
		services = append(services, serviceInfo)
	}

	response := map[string]interface{}{
		"profileId":       profile.ID,
		"profileName":     profile.Name,
		"environment":     request.Environment,
		"services":        services,
		"serviceCount":    len(compose.Services),
		"networkCount":    len(compose.Networks),
		"volumeCount":     len(compose.Volumes),
		"hasExternalDeps": dh.hasExternalDependencies(profile),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode preview response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// generateOverrideHandler generates a docker-compose.override.yml file
func (dh *dockerComposeHandler) generateOverrideHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := extractClaimsFromRequest(r, dh.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	// Get profile
	profile, err := dh.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get service profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get service profile", http.StatusInternalServerError)
		}
		return
	}

	// Generate override file
	override, err := dh.dockerComposeService.GenerateOverrideFile(profile)
	if err != nil {
		log.Printf("[ERROR] Failed to generate Docker Compose override: %v", err)
		http.Error(w, fmt.Sprintf("Failed to generate override file: %v", err), http.StatusInternalServerError)
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("docker-compose.override-%s.yml", sanitizeFilename(profile.Name))
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Write YAML content
	yamlContent := override.ToYAML()
	if _, err := w.Write([]byte(yamlContent)); err != nil {
		log.Printf("[ERROR] Failed to write override YAML content: %v", err)
		return
	}

	log.Printf("[INFO] Docker Compose override file generated for profile '%s'", profile.Name)
}

// getDockerConfigHandler retrieves Docker configuration for a profile
func (dh *dockerComposeHandler) getDockerConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, dh.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	// Verify profile access
	_, err := dh.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to verify profile access: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to verify profile access", http.StatusInternalServerError)
		}
		return
	}

	// Get Docker config - we'll need to add a method to access the database through the service
	config, err := dh.dockerComposeService.GetDockerConfig(profileID)
	if err != nil {
		log.Printf("[ERROR] Failed to get Docker config: %v", err)
		http.Error(w, "Failed to get Docker configuration", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(config); err != nil {
		log.Printf("[ERROR] Failed to encode Docker config response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// updateDockerConfigHandler updates Docker configuration for a profile
func (dh *dockerComposeHandler) updateDockerConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, dh.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	// Verify profile access
	_, err := dh.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to verify profile access: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to verify profile access", http.StatusInternalServerError)
		}
		return
	}

	var config models.DockerConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		log.Printf("[ERROR] Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure profile ID matches
	config.ProfileID = profileID

	// Save Docker config
	if err := dh.dockerComposeService.SaveDockerConfig(&config); err != nil {
		log.Printf("[ERROR] Failed to save Docker config: %v", err)
		http.Error(w, "Failed to save Docker configuration", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Docker configuration updated successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// deleteDockerConfigHandler deletes Docker configuration for a profile
func (dh *dockerComposeHandler) deleteDockerConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, dh.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	// Verify profile access
	_, err := dh.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to verify profile access: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to verify profile access", http.StatusInternalServerError)
		}
		return
	}

	// Delete Docker config
	if err := dh.dockerComposeService.DeleteDockerConfig(profileID); err != nil {
		log.Printf("[ERROR] Failed to delete Docker config: %v", err)
		http.Error(w, "Failed to delete Docker configuration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func getServiceNames(services map[string]models.ComposeService) []string {
	names := make([]string, 0, len(services))
	for name := range services {
		names = append(names, name)
	}
	return names
}

func sanitizeFilename(filename string) string {
	// Replace spaces and special characters with hyphens
	sanitized := strings.ReplaceAll(filename, " ", "-")
	sanitized = strings.ToLower(sanitized)
	
	// Remove any characters that aren't alphanumeric, hyphens, or underscores
	var result strings.Builder
	for _, r := range sanitized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

func (dh *dockerComposeHandler) hasExternalDependencies(profile *models.ServiceProfile) bool {
	// Check if any services in the profile have environment variables that suggest external dependencies
	// This is a simple heuristic - in a real implementation you'd want more sophisticated detection
	
	// Check profile environment variables for common external service indicators
	for key := range profile.EnvVars {
		upperKey := strings.ToUpper(key)
		if strings.Contains(upperKey, "DATABASE") || strings.Contains(upperKey, "REDIS") || 
		   strings.Contains(upperKey, "POSTGRES") || strings.Contains(upperKey, "MYSQL") ||
		   strings.Contains(upperKey, "MONGO") || strings.Contains(upperKey, "RABBITMQ") {
			return true
		}
	}
	
	return false
}