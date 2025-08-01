package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zechtz/vertex/internal/models"
)

func registerUserRoutes(h *Handler, r *mux.Router) {
	r.HandleFunc("/api/auth/register", h.registerHandler).Methods("POST")
	r.HandleFunc("/api/auth/login", h.loginHandler).Methods("POST")
	r.HandleFunc("/api/auth/user", h.getCurrentUserHandler).Methods("GET")
	r.HandleFunc("/api/user/profile", h.getUserProfileHandler).Methods("GET")
	r.HandleFunc("/api/user/profile", h.updateUserProfileHandler).Methods("PUT")
}

// registerHandler handles user registration
func (h *Handler) registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var registration models.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&registration); err != nil {
		log.Printf("[ERROR] Failed to decode registration request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := registration.Validate(); err != nil {
		log.Printf("[ERROR] Validation failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(&registration)
	if err != nil {
		log.Printf("[ERROR] Failed to register user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] User registered successfully: %s (%s)", user.Username, user.Email)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User registered successfully",
		"user":    user,
	}); err != nil {
		log.Printf("[ERROR] Failed to encode registration response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// loginHandler handles user login
func (h *Handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var login models.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		log.Printf("[ERROR] Failed to decode login request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if login.Email == "" || login.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	authResponse, err := h.authService.Login(&login)
	if err != nil {
		log.Printf("[ERROR] Failed to login user: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	log.Printf("[INFO] User logged in successfully: %s", authResponse.User.Username)

	if err := json.NewEncoder(w).Encode(authResponse); err != nil {
		log.Printf("[ERROR] Failed to encode login response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getCurrentUserHandler returns the current user info based on JWT token
func (h *Handler) getCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	claims, err := h.authService.ValidateToken(tokenParts[1])
	if err != nil {
		log.Printf("[ERROR] Failed to validate token: %v", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	user, err := h.authService.GetUserByID(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get user: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("[ERROR] Failed to encode user response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getUserProfileHandler retrieves the current user's profile
func (h *Handler) getUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := h.profileService.GetUserProfile(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get user profile: %v", err)
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Printf("[ERROR] Failed to encode profile response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// updateUserProfileHandler updates the current user's profile
func (h *Handler) updateUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := extractClaimsFromRequest(r, h.authService)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.UserProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	profile, err := h.profileService.UpdateUserProfile(claims.UserID, &req)
	if err != nil {
		log.Printf("[ERROR] Failed to update user profile: %v", err)
		http.Error(w, "Failed to update user profile", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Printf("[ERROR] Failed to encode profile response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
