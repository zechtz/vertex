// Package models
package models

import (
	"os/exec"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	Name         string            `json:"name"`
	Dir          string            `json:"dir"`
	ExtraEnv     string            `json:"extraEnv"`
	JavaOpts     string            `json:"javaOpts"`
	Status       string            `json:"status"`
	HealthStatus string            `json:"healthStatus"`
	HealthURL    string            `json:"healthUrl"`
	Port         int               `json:"port"`
	PID          int               `json:"pid"`
	Order        int               `json:"order"`
	LastStarted  time.Time         `json:"lastStarted"`
	Uptime       string            `json:"uptime"`
	Description  string            `json:"description"`
	IsEnabled    bool              `json:"isEnabled"`
	BuildSystem  string            `json:"buildSystem"` // "maven", "gradle", or "auto"
	EnvVars      map[string]EnvVar `json:"envVars"`
	Cmd          *exec.Cmd         `json:"-"`
	Logs         []LogEntry        `json:"logs"`
	Mutex        sync.RWMutex      `json:"-"`
	// Resource monitoring fields
	CPUPercent    float64           `json:"cpuPercent"`
	MemoryUsage   uint64            `json:"memoryUsage"`   // in bytes
	MemoryPercent float32           `json:"memoryPercent"`
	DiskUsage     uint64            `json:"diskUsage"`     // in bytes
	NetworkRx     uint64            `json:"networkRx"`     // bytes received
	NetworkTx     uint64            `json:"networkTx"`     // bytes transmitted
	Metrics       ServiceMetrics    `json:"metrics"`
	// Service dependencies
	Dependencies  []ServiceDependency `json:"dependencies"`
	DependentOn   []string            `json:"dependentOn"`   // Services that depend on this one
	StartupDelay  time.Duration       `json:"startupDelay"`  // Delay before starting after dependencies
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type ServiceMetrics struct {
	ResponseTimes []ResponseTime `json:"responseTimes"`
	ErrorRate     float64        `json:"errorRate"`
	RequestCount  uint64         `json:"requestCount"`
	LastChecked   time.Time      `json:"lastChecked"`
}

// Topology models for service visualization
type ServiceTopology struct {
	Services    []TopologyNode `json:"services"`
	Connections []Connection   `json:"connections"`
	Generated   time.Time      `json:"generated"`
}

type TopologyNode struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Type         string  `json:"type"` // "service", "database", "external"
	Status       string  `json:"status"`
	HealthStatus string  `json:"healthStatus"`
	Port         int     `json:"port"`
	Position     *NodePosition `json:"position,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type Connection struct {
	Source      string `json:"source"`
	Target      string `json:"target"`
	Type        string `json:"type"` // "http", "database", "message_queue"
	Status      string `json:"status"` // "active", "inactive", "error"
	Description string `json:"description"`
}

type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ServiceDependency represents a dependency relationship between services
type ServiceDependency struct {
	ServiceName     string        `json:"serviceName"`     // Name of the dependent service
	Type            string        `json:"type"`            // "hard", "soft", "optional"
	HealthCheck     bool          `json:"healthCheck"`     // Whether to check health before considering ready
	Timeout         time.Duration `json:"timeout"`         // Max time to wait for dependency
	RetryInterval   time.Duration `json:"retryInterval"`   // Interval between dependency checks
	Required        bool          `json:"required"`        // Whether this dependency is required for startup
	Description     string        `json:"description"`     // Human-readable description
}

type ResponseTime struct {
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Status    int           `json:"status"`
}

type EnvVar struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	IsRequired  bool   `json:"isRequired"`
}

type ServiceConfigRequest struct {
	Name        string            `json:"name"`
	Dir         string            `json:"dir"`
	JavaOpts    string            `json:"javaOpts"`
	HealthURL   string            `json:"healthUrl"`
	Port        int               `json:"port"`
	Order       int               `json:"order"`
	Description string            `json:"description"`
	IsEnabled   bool              `json:"isEnabled"`
	BuildSystem string            `json:"buildSystem"` // "maven", "gradle", or "auto"
	EnvVars     map[string]EnvVar `json:"envVars"`
}

type Config struct {
	ProjectsDir      string    `json:"projectsDir"`
	JavaHomeOverride string    `json:"javaHomeOverride"`
	Services         []Service `json:"services"`
}

type ConfigService struct {
	Name  string `json:"name"`
	Order int    `json:"order"`
}

type Configuration struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Services  []ConfigService `json:"services"`
	IsDefault bool            `json:"isDefault,omitempty"`
}

// LibraryInstallation represents a Maven library installation command
type LibraryInstallation struct {
	File       string `json:"file"`       // Path to the JAR file
	GroupID    string `json:"groupId"`    // Maven group ID
	ArtifactID string `json:"artifactId"` // Maven artifact ID
	Version    string `json:"version"`    // Version
	Packaging  string `json:"packaging"`  // Usually "jar"
	Command    string `json:"command"`    // Full maven command
}

// GitLabCIConfig represents the structure we care about from .gitlab-ci.yml
type GitLabCIConfig struct {
	ServiceName    string                 `json:"serviceName"`
	Libraries      []LibraryInstallation  `json:"libraries"`
	HasLibraries   bool                   `json:"hasLibraries"`
	ErrorMessage   string                 `json:"errorMessage,omitempty"`
}

// User represents a user account
type User struct {
	ID        string    `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	LastLogin time.Time `json:"lastLogin" db:"last_login"`
}

// UserRegistration represents user registration request
type UserRegistration struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UserLogin represents user login request
type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Ensure JWTClaims implements jwt.Claims interface
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

// UserProfile represents extended user profile information
type UserProfile struct {
	UserID      string          `json:"userId" db:"user_id"`
	DisplayName string          `json:"displayName" db:"display_name"`
	Avatar      string          `json:"avatar" db:"avatar"`
	Preferences UserPreferences `json:"preferences" db:"preferences_json"`
	CreatedAt   time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time       `json:"updatedAt" db:"updated_at"`
}

// UserPreferences represents user application preferences
type UserPreferences struct {
	Theme                string            `json:"theme"`
	Language             string            `json:"language"`
	NotificationSettings map[string]bool   `json:"notificationSettings"`
	DashboardLayout      string            `json:"dashboardLayout"`
	AutoRefresh          bool              `json:"autoRefresh"`
	RefreshInterval      int               `json:"refreshInterval"` // seconds
}

// ServiceProfile represents a collection of services and their configurations
type ServiceProfile struct {
	ID               string            `json:"id" db:"id"`
	UserID           string            `json:"userId" db:"user_id"`
	Name             string            `json:"name" db:"name"`
	Description      string            `json:"description" db:"description"`
	Services         []string          `json:"services" db:"services_json"`
	EnvVars          map[string]string `json:"envVars" db:"env_vars_json"`
	ProjectsDir      string            `json:"projectsDir" db:"projects_dir"`
	JavaHomeOverride string            `json:"javaHomeOverride" db:"java_home_override"`
	IsDefault        bool              `json:"isDefault" db:"is_default"`
	IsActive         bool              `json:"isActive" db:"is_active"`
	CreatedAt        time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time         `json:"updatedAt" db:"updated_at"`
}

// ProfileService represents a service within a profile with its configuration
type ProfileService struct {
	ServiceName string            `json:"serviceName"`
	IsEnabled   bool              `json:"isEnabled"`
	Order       int               `json:"order"`
	EnvVars     map[string]string `json:"envVars"`
}

// CreateProfileRequest represents a request to create a new service profile
type CreateProfileRequest struct {
	Name             string            `json:"name" validate:"required,min=3,max=100"`
	Description      string            `json:"description"`
	Services         []string          `json:"services"`
	EnvVars          map[string]string `json:"envVars"`
	ProjectsDir      string            `json:"projectsDir"`
	JavaHomeOverride string            `json:"javaHomeOverride"`
	IsDefault        bool              `json:"isDefault"`
}

// UpdateProfileRequest represents a request to update an existing service profile
type UpdateProfileRequest struct {
	Name             string            `json:"name" validate:"required,min=3,max=100"`
	Description      string            `json:"description"`
	Services         []string          `json:"services"`
	EnvVars          map[string]string `json:"envVars"`
	ProjectsDir      string            `json:"projectsDir"`
	JavaHomeOverride string            `json:"javaHomeOverride"`
	IsDefault        bool              `json:"isDefault"`
}

// UserProfileUpdateRequest represents a request to update user profile
type UserProfileUpdateRequest struct {
	DisplayName string          `json:"displayName"`
	Avatar      string          `json:"avatar"`
	Preferences UserPreferences `json:"preferences"`
}

// ProfileEnvVar represents a profile-scoped environment variable
type ProfileEnvVar struct {
	ID          int       `json:"id" db:"id"`
	ProfileID   string    `json:"profileId" db:"profile_id"`
	VarName     string    `json:"varName" db:"var_name"`
	VarValue    string    `json:"varValue" db:"var_value"`
	Description string    `json:"description" db:"description"`
	IsRequired  bool      `json:"isRequired" db:"is_required"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// ProfileServiceConfig represents a profile-scoped service configuration override
type ProfileServiceConfig struct {
	ID           int       `json:"id" db:"id"`
	ProfileID    string    `json:"profileId" db:"profile_id"`
	ServiceName  string    `json:"serviceName" db:"service_name"`
	ConfigKey    string    `json:"configKey" db:"config_key"`
	ConfigValue  string    `json:"configValue" db:"config_value"`
	ConfigType   string    `json:"configType" db:"config_type"`
	Description  string    `json:"description" db:"description"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// ProfileDependency represents a profile-scoped service dependency
type ProfileDependency struct {
	ID                    int           `json:"id" db:"id"`
	ProfileID             string        `json:"profileId" db:"profile_id"`
	ServiceName           string        `json:"serviceName" db:"service_name"`
	DependencyServiceName string        `json:"dependencyServiceName" db:"dependency_service_name"`
	DependencyType        string        `json:"dependencyType" db:"dependency_type"`
	HealthCheck           bool          `json:"healthCheck" db:"health_check"`
	TimeoutSeconds        int           `json:"timeoutSeconds" db:"timeout_seconds"`
	RetryIntervalSeconds  int           `json:"retryIntervalSeconds" db:"retry_interval_seconds"`
	IsRequired            bool          `json:"isRequired" db:"is_required"`
	Description           string        `json:"description" db:"description"`
	CreatedAt             time.Time     `json:"createdAt" db:"created_at"`
	UpdatedAt             time.Time     `json:"updatedAt" db:"updated_at"`
}

// ProfileContext represents the complete configuration context for a profile
type ProfileContext struct {
	Profile           *ServiceProfile                    `json:"profile"`
	EnvVars           map[string]string                  `json:"envVars"`
	ServiceConfigs    map[string]map[string]string       `json:"serviceConfigs"`    // serviceName -> configKey -> configValue
	Dependencies      map[string][]ProfileDependency     `json:"dependencies"`      // serviceName -> dependencies
	IsActive          bool                               `json:"isActive"`
}
