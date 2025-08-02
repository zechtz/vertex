package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zechtz/vertex/internal/database"
	"github.com/zechtz/vertex/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db        *database.Database
	jwtSecret []byte
}

func NewAuthService(db *database.Database) *AuthService {
	// Try to get JWT secret from environment variable first
	jwtSecretStr := os.Getenv("JWT_SECRET")
	var secret []byte

	if jwtSecretStr != "" {
		// Use provided secret from environment
		secret = []byte(jwtSecretStr)
		log.Printf("[INFO] Using JWT secret from environment variable")
	} else {
		// Use a consistent fallback secret for development
		// In production, you should always set JWT_SECRET environment variable
		const fallbackSecret = "vertex-manager-development-secret-change-in-production"
		secret = []byte(fallbackSecret)
		log.Printf("[WARN] Using fallback JWT secret. Set JWT_SECRET environment variable for production")
	}

	return &AuthService{
		db:        db,
		jwtSecret: secret,
	}
}

// Register creates a new user account
func (as *AuthService) Register(registration *models.UserRegistration) (*models.User, error) {
	// Check if user already exists
	if exists, err := as.userExists(registration.Email, registration.Username); err != nil {
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	} else if exists {
		return nil, fmt.Errorf("user with this email or username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registration.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate user ID
	userID := generateUserID()

	// Create user
	user := &models.User{
		ID:        userID,
		Username:  registration.Username,
		Email:     registration.Email,
		Password:  string(hashedPassword),
		Role:      "user",
		CreatedAt: time.Now(),
	}

	// Save to database
	if err := as.saveUser(user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// Don't return password hash
	user.Password = ""
	return user, nil
}

// Login authenticates a user and returns JWT token
func (as *AuthService) Login(login *models.UserLogin) (*models.AuthResponse, error) {
	// Get user by email
	user, err := as.getUserByEmail(login.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Update last login
	if err := as.updateLastLogin(user.ID); err != nil {
		log.Printf("Failed to update last login for user %s: %v", user.ID, err)
	}

	// Generate JWT token
	token, err := as.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Don't return password hash
	user.Password = ""
	user.LastLogin = time.Now()

	return &models.AuthResponse{
		User:  *user,
		Token: token,
	}, nil
}

// ValidateToken validates a JWT token and returns user claims
func (as *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Make sure token method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return as.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		// Check if token is expired
		if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
			return nil, fmt.Errorf("token has expired")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetUserByID retrieves a user by ID
func (as *AuthService) GetUserByID(userID string) (*models.User, error) {
	user, err := as.getUserByID(userID)
	if err != nil {
		return nil, err
	}
	// Don't return password hash
	user.Password = ""
	return user, nil
}

// Private helper methods

func (as *AuthService) userExists(email, username string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = ? OR username = ?`
	err := as.db.QueryRow(query, email, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (as *AuthService) saveUser(user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, role, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := as.db.Exec(query, user.ID, user.Username, user.Email, user.Password, user.Role, user.CreatedAt)
	return err
}

func (as *AuthService) getUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password_hash, role, created_at, last_login
		FROM users WHERE email = ?
	`
	var lastLogin sql.NullTime
	err := as.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.CreatedAt, &lastLogin,
	)
	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}

	return user, nil
}

func (as *AuthService) getUserByID(userID string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password_hash, role, created_at, last_login
		FROM users WHERE id = ?
	`
	var lastLogin sql.NullTime
	err := as.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.CreatedAt, &lastLogin,
	)
	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}

	return user, nil
}

func (as *AuthService) updateLastLogin(userID string) error {
	query := `UPDATE users SET last_login = ? WHERE id = ?`
	_, err := as.db.Exec(query, time.Now(), userID)
	return err
}

func (as *AuthService) generateJWT(user *models.User) (string, error) {
	claims := &models.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "vertex-manager",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(as.jwtSecret)
}

func generateUserID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal("Failed to generate user ID")
	}
	return hex.EncodeToString(bytes)
}
