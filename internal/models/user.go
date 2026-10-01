package models

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID        string    `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	LastLogin time.Time `json:"lastLogin" db:"last_login"`
}

type UserRegistration struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// Validate validates the UserRegistration fields
func (ur *UserRegistration) Validate() error {
	if ur.Email == "" {
		return errors.New("email is required")
	}
	if ur.Username == "" {
		return errors.New("username is required")
	}
	return ValidatePassword(ur.Password)
}

// ValidatePassword is the rule every new password must meet, whether it is
// chosen at registration, changed by its owner or reset from the command line.
func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	return nil
}

type PasswordChange struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

type JWTClaims struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type UserProfile struct {
	UserID      string          `json:"userId" db:"user_id"`
	DisplayName string          `json:"displayName" db:"display_name"`
	Avatar      string          `json:"avatar" db:"avatar"`
	Preferences UserPreferences `json:"preferences" db:"preferences_json"`
	CreatedAt   time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time       `json:"updatedAt" db:"updated_at"`
}

type UserPreferences struct {
	Theme                string          `json:"theme"`
	Language             string          `json:"language"`
	NotificationSettings map[string]bool `json:"notificationSettings"`
	DashboardLayout      string          `json:"dashboardLayout"`
	AutoRefresh          bool            `json:"autoRefresh"`
	RefreshInterval      int             `json:"refreshInterval"` // seconds
}

type UserProfileUpdateRequest struct {
	DisplayName string          `json:"displayName"`
	Avatar      string          `json:"avatar"`
	Preferences UserPreferences `json:"preferences"`
}

// GetExpirationTime implements jwt.Claims interface
func (c *JWTClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.ExpiresAt, nil
}

func (c *JWTClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.IssuedAt, nil
}

func (c *JWTClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.NotBefore, nil
}

func (c *JWTClaims) GetIssuer() (string, error) {
	return c.RegisteredClaims.Issuer, nil
}

func (c *JWTClaims) GetSubject() (string, error) {
	return c.RegisteredClaims.Subject, nil
}

func (c *JWTClaims) GetAudience() (jwt.ClaimStrings, error) {
	return c.RegisteredClaims.Audience, nil
}
