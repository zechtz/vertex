package models

type Config struct {
	ProjectsDir      string    `json:"projectsDir"`
	JavaHomeOverride string    `json:"javaHomeOverride"`
	Services         []Service `json:"services"`
}

type ConfigService struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

type Configuration struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Services  []ConfigService `json:"services"`
	IsDefault bool            `json:"isDefault,omitempty"`
}

type ServiceConfigRequest struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Dir            string            `json:"dir"`
	JavaOpts       string            `json:"javaOpts"`
	HealthURL      string            `json:"healthUrl"`
	Port           int               `json:"port"`
	Order          int               `json:"order"`
	Description    string            `json:"description"`
	IsEnabled      bool              `json:"isEnabled"`
	BuildSystem    string            `json:"buildSystem"`    // "maven", "gradle", or "auto"
	VerboseLogging bool              `json:"verboseLogging"` // Enable verbose/debug logging for build tools
	EnvVars        map[string]EnvVar `json:"envVars"`
}
