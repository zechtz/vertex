package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zechtz/vertex/internal/models"
)

func registerConfigRoutes(h *Handler, r *mux.Router) {
	r.HandleFunc("/api/configurations", h.getConfigurationsHandler).Methods("GET")
	r.HandleFunc("/api/configurations", h.saveConfigurationHandler).Methods("POST")
	r.HandleFunc("/api/configurations/{id}", h.updateConfigurationHandler).Methods("PUT")
	r.HandleFunc("/api/configurations/{id}/apply", h.applyConfigurationHandler).Methods("POST")
	r.HandleFunc("/api/config/global", h.getGlobalConfigHandler).Methods("GET")
	r.HandleFunc("/api/config/global", h.updateGlobalConfigHandler).Methods("PUT")
}

func (h *Handler) getConfigurationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	configs := h.serviceManager.GetConfigurations()
	json.NewEncoder(w).Encode(configs)
}

func (h *Handler) saveConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var config models.Configuration
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if config.ID == "" {
		config.ID = fmt.Sprintf("config_%s", uuid.New().String())
	}

	if err := h.serviceManager.SaveConfiguration(&config); err != nil {
		log.Printf("Failed to save configuration: %v", err)
		http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"id":     config.ID,
	})
}

func (h *Handler) updateConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	configID := vars["id"]

	if configID == "" {
		http.Error(w, "Configuration ID is required", http.StatusBadRequest)
		return
	}

	var config models.Configuration
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure the ID matches the URL parameter
	config.ID = configID

	if err := h.serviceManager.UpdateConfiguration(&config); err != nil {
		log.Printf("Failed to update configuration: %v", err)
		http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"id":     config.ID,
	})
}

func (h *Handler) applyConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	configID := vars["id"]

	if configID == "" {
		http.Error(w, "Configuration ID is required", http.StatusBadRequest)
		return
	}

	if err := h.serviceManager.ApplyConfiguration(configID); err != nil {
		log.Printf("Failed to apply configuration: %v", err)
		http.Error(w, "Failed to apply configuration", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) getGlobalConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	config := h.serviceManager.GetGlobalConfig()
	json.NewEncoder(w).Encode(config)
}

func (h *Handler) updateGlobalConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var request struct {
		ProjectsDir      string `json:"projectsDir"`
		JavaHomeOverride string `json:"javaHomeOverride"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	config, err := h.serviceManager.UpdateGlobalConfig(request.ProjectsDir, request.JavaHomeOverride)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(config)
}
