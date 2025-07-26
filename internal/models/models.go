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
