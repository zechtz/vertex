// Package handlers
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/zechtz/vertex/internal/models"
	"github.com/zechtz/vertex/internal/services"
)

type Handler struct {
	serviceManager       *services.Manager
	topologyService      *services.TopologyService
	autoDiscoveryService *services.AutoDiscoveryService
	authService          *services.AuthService
	profileService       *services.ProfileService
	upgrader             websocket.Upgrader
}

func NewHandler(sm *services.Manager) *Handler {
	return &Handler{
		serviceManager:       sm,
		topologyService:      services.NewTopologyService(sm),
		autoDiscoveryService: services.NewAutoDiscoveryService(sm),
		authService:          services.NewAuthService(sm.GetDatabase()),
		profileService:       services.NewProfileService(sm.GetDatabase(), sm),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// getServiceProjectsDir determines the appropriate projects directory for a service
// If the service belongs to an active profile with a custom projectsDir, use that
// Otherwise, use the global default
func (h *Handler) getServiceProjectsDir(serviceUUID string) string {
	// Get global config as fallback
	globalConfig := h.serviceManager.GetConfig()
	defaultProjectsDir := globalConfig.ProjectsDir
	
	log.Printf("[DEBUG] getServiceProjectsDir - serviceUUID: %s, defaultProjectsDir: '%s'", serviceUUID, defaultProjectsDir)

	// Query database to find all profiles that contain this service
	query := `SELECT projects_dir FROM service_profiles 
			  WHERE services_json LIKE ? AND projects_dir != '' AND projects_dir IS NOT NULL
			  ORDER BY is_active DESC, is_default DESC, created_at DESC
			  LIMIT 1`

	// Use LIKE to search for the service UUID in the JSON array
	searchPattern := fmt.Sprintf("%%\"%s\"%%", serviceUUID)
	log.Printf("[DEBUG] Searching for service in profiles with pattern: %s", searchPattern)

	var projectsDir string
	err := h.serviceManager.GetDatabase().QueryRow(query, searchPattern).Scan(&projectsDir)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("[WARN] Failed to query profile projects directory for service UUID %s: %v", serviceUUID, err)
		} else {
			log.Printf("[DEBUG] No profile found containing service UUID %s", serviceUUID)
		}
		// No profile contains this service or query failed, use global default
		// If global default is empty, try current working directory
		if defaultProjectsDir == "" {
			if cwd, err := os.Getwd(); err == nil {
				log.Printf("[INFO] No projects directory configured, using current directory: %s", cwd)
				return cwd
			}
		}
	} else {
		log.Printf("[DEBUG] Found profile projects directory for service: %s", projectsDir)
	}

	if projectsDir != "" {
		log.Printf("[DEBUG] Returning profile-specific projects directory: %s", projectsDir)
		return projectsDir
	}

	log.Printf("[DEBUG] Returning default projects directory: %s", defaultProjectsDir)
	return defaultProjectsDir
}

// getServiceProjectsDirForUser determines the appropriate projects directory for a service for a specific user
// Checks the user's active profile first, then falls back to global logic
func (h *Handler) getServiceProjectsDirForUser(serviceUUID, userID string) string {
	log.Printf("[DEBUG] getServiceProjectsDirForUser called - serviceUUID: %s, userID: %s", serviceUUID, userID)
	
	// First, try to get the user's active profile
	activeProfile, err := h.profileService.GetActiveProfile(userID)
	if err != nil {
		log.Printf("[WARN] Failed to get active profile for user %s: %v", userID, err)
		fallbackDir := h.getServiceProjectsDir(serviceUUID)
		log.Printf("[DEBUG] Using fallback directory: %s", fallbackDir)
		return fallbackDir
	}

	log.Printf("[DEBUG] Active profile found: %+v", activeProfile)

	// Check if the active profile contains this service and has a custom projects directory
	if activeProfile != nil && activeProfile.ProjectsDir != "" {
		log.Printf("[DEBUG] Active profile has ProjectsDir: %s", activeProfile.ProjectsDir)
		log.Printf("[DEBUG] Active profile services: %v", activeProfile.Services)
		
		// Check if the service is in this profile
		if slices.Contains(activeProfile.Services, serviceUUID) {
			log.Printf("[INFO] Using active profile projects directory for service UUID %s: %s", serviceUUID, activeProfile.ProjectsDir)
			return activeProfile.ProjectsDir
		} else {
			log.Printf("[DEBUG] Service %s not found in active profile services", serviceUUID)
		}
	} else {
		if activeProfile == nil {
			log.Printf("[DEBUG] Active profile is nil")
		} else {
			log.Printf("[DEBUG] Active profile ProjectsDir is empty: '%s'", activeProfile.ProjectsDir)
		}
	}

	// Service not in active profile or no active profile, use global logic
	fallbackDir := h.getServiceProjectsDir(serviceUUID)
	log.Printf("[DEBUG] Using global fallback directory: %s", fallbackDir)
	return fallbackDir
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	registerUtilityRoutes(h, r)
	// Authentication routes (public)
	registerUserRoutes(h, r)

	// Profile Management routes (protected)

	// Profile-scoped configuration routes (protected)
	registerProfileRoutes(h, r)
	registerCIRoutes(h, r)
	registerConfigRoutes(h, r)
	registerServiceRoutes(h, r)
	registerUptimeRoutes(h, r)

	// Service routes (will be protected later)
	registerTopologyRoutes(h, r)
}

// sendAutoDiscoveryResponse sends the auto-discovery scan results
func (h *Handler) sendAutoDiscoveryResponse(w http.ResponseWriter, discoveredServices []services.DiscoveredService) {
	result := map[string]any{
		"success":            true,
		"message":            fmt.Sprintf("Found %d potential services", len(discoveredServices)),
		"discoveredServices": discoveredServices,
		"totalFound":         len(discoveredServices),
	}

	log.Printf("[INFO] Auto-discovery scan completed. Found %d services", len(discoveredServices))

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode auto-discovery response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Helper function to extract JWT claims from request
func extractClaimsFromRequest(r *http.Request, authService *services.AuthService) (*models.JWTClaims, bool) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, false
	}

	// Extract token (format: "Bearer <token>")
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return nil, false
	}

	// Validate token
	claims, err := authService.ValidateToken(tokenParts[1])
	if err != nil {
		return nil, false
	}

	return claims, true
}
