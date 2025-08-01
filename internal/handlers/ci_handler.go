package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func registerCIRoutes(h *Handler, r *mux.Router) {
	r.HandleFunc("/api/services/{id}/gitlab-ci", h.getGitLabCIHandler).Methods("GET")
	r.HandleFunc("/api/services/gitlab-ci/all", h.getAllGitLabCIHandler).Methods("GET")
}

// getGitLabCIHandler returns GitLab CI configuration for a specific service
func (h *Handler) getGitLabCIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceUUID := vars["id"]

	if serviceUUID == "" {
		http.Error(w, "Service UUID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	config, err := h.serviceManager.ParseGitLabCI(serviceUUID)
	if err != nil {
		log.Printf("Failed to parse GitLab CI for service UUID %s: %v", serviceUUID, err)
		http.Error(w, fmt.Sprintf("Failed to parse GitLab CI: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(config); err != nil {
		log.Printf("Failed to encode GitLab CI config: %v", err)
		http.Error(w, "Failed to encode GitLab CI config", http.StatusInternalServerError)
		return
	}
}

// getAllGitLabCIHandler returns GitLab CI configurations for all services
func (h *Handler) getAllGitLabCIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	configs := h.serviceManager.GetAllGitLabCIConfigs()

	if err := json.NewEncoder(w).Encode(configs); err != nil {
		log.Printf("Failed to encode GitLab CI configs: %v", err)
		http.Error(w, "Failed to encode GitLab CI configs", http.StatusInternalServerError)
		return
	}
}
