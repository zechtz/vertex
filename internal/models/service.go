// Package models
package models

import (
	"os/exec"
	"sync"
	"time"
)

type Service struct {
	ID             string              `json:"id"` // UUID - unique identifier for the service
	Name           string              `json:"name"`
	Dir            string              `json:"dir"`
	ExtraEnv       string              `json:"extraEnv"`
	JavaOpts       string              `json:"javaOpts"`
	Status         string              `json:"status"`
	HealthStatus   string              `json:"healthStatus"`
	HealthURL      string              `json:"healthUrl"`
	Port           int                 `json:"port"`
	PID            int                 `json:"pid"`
	Order          int                 `json:"order"`
	LastStarted    time.Time           `json:"lastStarted"`
	Uptime         string              `json:"uptime"`
	Description    string              `json:"description"`
	IsEnabled      bool                `json:"isEnabled"`
	BuildSystem    string              `json:"buildSystem"`    // "maven", "gradle", or "auto"
	VerboseLogging bool                `json:"verboseLogging"` // Enable verbose/debug logging for build tools
	GitBranch      string              `json:"gitBranch"`      // Current git branch (if service is a git repo)
	EnvVars        map[string]EnvVar   `json:"envVars"`
	Cmd            *exec.Cmd           `json:"-"`
	Logs           []LogEntry          `json:"logs"`
	Mutex          sync.RWMutex        `json:"-"`
	CPUPercent     float64             `json:"cpuPercent"`
	MemoryUsage    uint64              `json:"memoryUsage"` // in bytes
	MemoryPercent  float32             `json:"memoryPercent"`
	DiskUsage      uint64              `json:"diskUsage"` // in bytes
	NetworkRx      uint64              `json:"networkRx"` // bytes received
	NetworkTx      uint64              `json:"networkTx"` // bytes transmitted
	Metrics        ServiceMetrics      `json:"metrics"`
	Dependencies   []ServiceDependency `json:"dependencies"`
	DependentOn    []string            `json:"dependentOn"`  // Services that depend on this one
	StartupDelay   time.Duration       `json:"startupDelay"` // Delay before starting after dependencies
}
