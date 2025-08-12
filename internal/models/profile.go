package models

import (
	"time"
)

type ServiceProfile struct {
	ID               string            `json:"id" db:"id"`
	UserID           string            `json:"userId" db:"user_id"`
	Name             string            `json:"name" db:"name"`
	Description      string            `json:"description" db:"description"`
	Services         []string          `json:"services" db:"services_json"` // service ids (UUID)
	EnvVars          map[string]string `json:"envVars" db:"env_vars_json"`
	ProjectsDir      string            `json:"projectsDir" db:"projects_dir"`
	JavaHomeOverride string            `json:"javaHomeOverride" db:"java_home_override"`
	IsDefault        bool              `json:"isDefault" db:"is_default"`
	IsActive         bool              `json:"isActive" db:"is_active"`
	CreatedAt        time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time         `json:"updatedAt" db:"updated_at"`
}

type ProfileService struct {
	ServiceName string            `json:"serviceName"`
	ServicePath string            `json:"servicePath"`
	IsEnabled   bool              `json:"isEnabled"`
	Order       int               `json:"order"`
	EnvVars     map[string]string `json:"envVars"`
}

type CreateProfileRequest struct {
	Name             string            `json:"name" validate:"required,min=3,max=100"`
	Description      string            `json:"description"`
	Services         []string          `json:"services"`
	EnvVars          map[string]string `json:"envVars"`
	ProjectsDir      string            `json:"projectsDir"`
	JavaHomeOverride string            `json:"javaHomeOverride"`
	IsDefault        bool              `json:"isDefault"`
}

type UpdateProfileRequest struct {
	Name             string            `json:"name" validate:"required,min=3,max=100"`
	Description      string            `json:"description"`
	Services         []string          `json:"services"`
	EnvVars          map[string]string `json:"envVars"`
	ProjectsDir      string            `json:"projectsDir"`
	JavaHomeOverride string            `json:"javaHomeOverride"`
	IsDefault        bool              `json:"isDefault"`
}

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

type ProfileServiceConfig struct {
	ID          int       `json:"id" db:"id"`
	ProfileID   string    `json:"profileId" db:"profile_id"`
	ServiceName string    `json:"serviceName" db:"service_name"`
	ConfigKey   string    `json:"configKey" db:"config_key"`
	ConfigValue string    `json:"configValue" db:"config_value"`
	ConfigType  string    `json:"configType" db:"config_type"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type ProfileDependency struct {
	ID                    int       `json:"id" db:"id"`
	ProfileID             string    `json:"profileId" db:"profile_id"`
	ServiceName           string    `json:"serviceName" db:"service_name"`
	DependencyServiceName string    `json:"dependencyServiceName" db:"dependency_service_name"`
	DependencyType        string    `json:"dependencyType" db:"dependency_type"`
	HealthCheck           bool      `json:"healthCheck" db:"health_check"`
	TimeoutSeconds        int       `json:"timeoutSeconds" db:"timeout_seconds"`
	RetryIntervalSeconds  int       `json:"retryIntervalSeconds" db:"retry_interval_seconds"`
	IsRequired            bool      `json:"isRequired" db:"is_required"`
	Description           string    `json:"description" db:"description"`
	CreatedAt             time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt             time.Time `json:"updatedAt" db:"updated_at"`
}

type ProfileContext struct {
	Profile        *ServiceProfile                `json:"profile"`
	EnvVars        map[string]string              `json:"envVars"`
	ServiceConfigs map[string]map[string]string   `json:"serviceConfigs"` // serviceName -> configKey -> configValue
	Dependencies   map[string][]ProfileDependency `json:"dependencies"`   // serviceName -> dependencies
	IsActive       bool                           `json:"isActive"`
}
