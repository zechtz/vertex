package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zechtz/vertex/internal/models"
)

func registerProfileRoutes(h *Handler, r *mux.Router) {
	r.HandleFunc("/api/profiles", h.getServiceProfilesHandler).Methods("GET")
	r.HandleFunc("/api/profiles", h.createServiceProfileHandler).Methods("POST")
	r.HandleFunc("/api/profiles/{id}", h.getServiceProfileHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}", h.updateServiceProfileHandler).Methods("PUT")
	r.HandleFunc("/api/profiles/{id}", h.deleteServiceProfileHandler).Methods("DELETE")
	r.HandleFunc("/api/profiles/{id}/apply", h.applyServiceProfileHandler).Methods("POST")
	r.HandleFunc("/api/profiles/{id}/activate", h.setActiveProfileHandler).Methods("POST")
	r.HandleFunc("/api/profiles/active", h.getActiveProfileHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}/context", h.getProfileContextHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}/env-vars", h.getProfileEnvVarsHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}/env-vars", h.setProfileEnvVarHandler).Methods("POST")
	r.HandleFunc("/api/profiles/{id}/env-vars/{name}", h.deleteProfileEnvVarHandler).Methods("DELETE")
	r.HandleFunc("/api/profiles/{id}/service-configs/{service}", h.getProfileServiceConfigHandler).Methods("GET")
	r.HandleFunc("/api/profiles/{id}/service-configs/{service}", h.setProfileServiceConfigHandler).Methods("POST")
	r.HandleFunc("/api/profiles/{id}/service-configs/{service}/{key}", h.deleteProfileServiceConfigHandler).Methods("DELETE")
	r.HandleFunc("/api/profiles/{id}/services", h.addServiceToProfileHandler).Methods("POST")
	r.HandleFunc("/api/profiles/{id}/services/{service}", h.removeServiceFromProfileHandler).Methods("DELETE")
}

func (h *Handler) getServiceProfilesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profiles, err := h.profileService.GetServiceProfiles(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get service profiles: %v", err)
		http.Error(w, "Failed to get service profiles", http.StatusInternalServerError)
		return
	}

	// Create response with enriched services
	response := make([]interface{}, len(profiles))
	for i, profile := range profiles {
		enrichedServices := make([]map[string]string, 0, len(profile.Services))
		for _, serviceUUID := range profile.Services {
			service, exists := h.serviceManager.GetServiceByUUID(serviceUUID)
			if exists {
				enrichedServices = append(enrichedServices, map[string]string{
					"id":   serviceUUID,
					"name": service.Name,
				})
			} else {
				enrichedServices = append(enrichedServices, map[string]string{
					"id":   serviceUUID,
					"name": "Unknown Service",
				})
			}
		}
		
		response[i] = map[string]interface{}{
			"id":               profile.ID,
			"userId":           profile.UserID,
			"name":             profile.Name,
			"description":      profile.Description,
			"services":         enrichedServices,
			"envVars":          profile.EnvVars,
			"projectsDir":      profile.ProjectsDir,
			"javaHomeOverride": profile.JavaHomeOverride,
			"isDefault":        profile.IsDefault,
			"isActive":         profile.IsActive,
			"createdAt":        profile.CreatedAt,
			"updatedAt":        profile.UpdatedAt,
		}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode profiles response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) getServiceProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
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

	profile, err := h.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get service profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get service profile", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Printf("[ERROR] Failed to encode profile response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) createServiceProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[DEBUG] Create profile request: %+v", req)

	profile, err := h.profileService.CreateServiceProfile(claims.UserID, &req)
	if err != nil {
		log.Printf("[ERROR] Failed to create service profile: %v", err)
		if strings.Contains(err.Error(), "invalid services") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to create service profile", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Printf("[ERROR] Failed to encode profile response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) updateServiceProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
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

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[DEBUG] Update profile request for ID %s: %+v", profileID, req)

	profile, err := h.profileService.UpdateServiceProfile(profileID, claims.UserID, &req)
	if err != nil {
		log.Printf("[ERROR] Failed to update service profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "invalid services") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to update service profile", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Printf("[ERROR] Failed to encode profile response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) deleteServiceProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
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

	err := h.profileService.DeleteServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to delete service profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete service profile", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) applyServiceProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
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

	err := h.profileService.ApplyProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to apply service profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to apply service profile", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{
		"message": "Profile applied successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) setActiveProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
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

	err := h.profileService.SetActiveProfile(claims.UserID, profileID)
	if err != nil {
		log.Printf("[ERROR] Failed to set active profile: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to set active profile", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{
		"message": "Active profile set successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) getActiveProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := h.profileService.GetActiveProfile(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get active profile: %v", err)
		http.Error(w, "Failed to get active profile", http.StatusInternalServerError)
		return
	}

	if profile == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) getProfileContextHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
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

	context, err := h.profileService.GetProfileContext(claims.UserID, profileID)
	if err != nil {
		log.Printf("[ERROR] Failed to get profile context: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get profile context", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(context); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) getProfileEnvVarsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
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

	envVars, err := h.profileService.GetProfileEnvVars(claims.UserID, profileID)
	if err != nil {
		log.Printf("[ERROR] Failed to get profile env vars: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get profile env vars", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(envVars); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) setProfileEnvVarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
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

	var request struct {
		Name        string `json:"name"`
		Value       string `json:"value"`
		Description string `json:"description"`
		IsRequired  bool   `json:"isRequired"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Name == "" {
		http.Error(w, "Environment variable name is required", http.StatusBadRequest)
		return
	}

	err := h.profileService.SetProfileEnvVar(claims.UserID, profileID, request.Name, request.Value, request.Description, request.IsRequired)
	if err != nil {
		log.Printf("[ERROR] Failed to set profile env var: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to set profile env var", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{
		"message": "Environment variable set successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) deleteProfileEnvVarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	name := vars["name"]

	if profileID == "" || name == "" {
		http.Error(w, "Profile ID and variable name are required", http.StatusBadRequest)
		return
	}

	err := h.profileService.DeleteProfileEnvVar(claims.UserID, profileID, name)
	if err != nil {
		log.Printf("[ERROR] Failed to delete profile env var: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile or variable not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete profile env var", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getProfileServiceConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	serviceName := vars["service"]

	if profileID == "" || serviceName == "" {
		http.Error(w, "Profile ID and service name are required", http.StatusBadRequest)
		return
	}

	_, err := h.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to verify profile access: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to verify profile access", http.StatusInternalServerError)
		}
		return
	}

	config, err := h.profileService.GetDatabase().GetProfileServiceConfig(profileID, serviceName)
	if err != nil {
		log.Printf("[ERROR] Failed to get profile service config: %v", err)
		http.Error(w, "Failed to get service config", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(config); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) setProfileServiceConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	serviceName := vars["service"]

	if profileID == "" || serviceName == "" {
		http.Error(w, "Profile ID and service name are required", http.StatusBadRequest)
		return
	}

	var request struct {
		Key         string `json:"key"`
		Value       string `json:"value"`
		ConfigType  string `json:"configType"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Key == "" {
		http.Error(w, "Configuration key is required", http.StatusBadRequest)
		return
	}

	if request.ConfigType == "" {
		request.ConfigType = "string"
	}

	_, err := h.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to verify profile access: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to verify profile access", http.StatusInternalServerError)
		}
		return
	}

	err = h.profileService.GetDatabase().SetProfileServiceConfig(profileID, serviceName, request.Key, request.Value, request.ConfigType, request.Description)
	if err != nil {
		log.Printf("[ERROR] Failed to set profile service config: %v", err)
		http.Error(w, "Failed to set service config", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Service configuration set successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) deleteProfileServiceConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	serviceName := vars["service"]
	key := vars["key"]

	if profileID == "" || serviceName == "" || key == "" {
		http.Error(w, "Profile ID, service name, and config key are required", http.StatusBadRequest)
		return
	}

	_, err := h.profileService.GetServiceProfile(profileID, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to verify profile access: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to verify profile access", http.StatusInternalServerError)
		}
		return
	}

	err = h.profileService.GetDatabase().DeleteProfileServiceConfig(profileID, serviceName, key)
	if err != nil {
		log.Printf("[ERROR] Failed to delete profile service config: %v", err)
		http.Error(w, "Failed to delete service config", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) addServiceToProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
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

	var request struct {
		ServiceName string `json:"serviceName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.ServiceName == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Adding service '%s' to profile '%s' for user '%s'", request.ServiceName, profileID, claims.UserID)

	// Convert service name to UUID
	services := h.serviceManager.GetServices()
	var serviceUUID string
	for _, service := range services {
		if service.Name == request.ServiceName {
			serviceUUID = service.ID
			break
		}
	}

	if serviceUUID == "" {
		log.Printf("[ERROR] Service '%s' not found", request.ServiceName)
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	err := h.profileService.AddServiceToProfile(claims.UserID, profileID, serviceUUID)
	if err != nil {
		log.Printf("[ERROR] Failed to add service to profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile or service not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "already exists") {
			http.Error(w, "Service already exists in profile", http.StatusConflict)
		} else {
			http.Error(w, "Failed to add service to profile", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{
		"message": "Service added to profile successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) removeServiceFromProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID := vars["id"]
	serviceName := vars["service"]

	if profileID == "" || serviceName == "" {
		http.Error(w, "Profile ID and service name are required", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Removing service '%s' from profile '%s' for user '%s'", serviceName, profileID, claims.UserID)

	err := h.profileService.RemoveServiceFromProfile(claims.UserID, profileID, serviceName)
	if err != nil {
		log.Printf("[ERROR] Failed to remove service from profile: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Profile or service not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to remove service from profile", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Service '%s' removed from profile successfully", serviceName),
	})
}
