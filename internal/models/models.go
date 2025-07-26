// Package models
package models

import (
	"os/exec"
	"sync"
	"time"
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
