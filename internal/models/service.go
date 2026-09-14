// Package models
package models

import (
	"os/exec"
	"sync"
	"time"
)

type Service struct {
	ID           string    `json:"id"` // UUID - unique identifier for the service
	Name         string    `json:"name"`
	Dir          string    `json:"dir"`
	ExtraEnv     string    `json:"extraEnv"`
	JavaOpts     string    `json:"javaOpts"`
	Status       string    `json:"status"`
	HealthStatus string    `json:"healthStatus"`
	HealthURL    string    `json:"healthUrl"`
	Port         int       `json:"port"`
	PID          int       `json:"pid"`
	Order        int       `json:"order"`
	LastStarted  time.Time `json:"lastStarted"`
	Uptime       string    `json:"uptime"`
	// FailureReason explains why the most recent start failed. The process exit
	// code alone is always "exit status 1", so the cause is recovered from the
	// service's own output instead. Nil once a start succeeds.
	FailureReason     *FailureReason `json:"failureReason,omitempty"`
	Description       string         `json:"description"`
	IsEnabled         bool           `json:"isEnabled"`
	BuildSystem       string         `json:"buildSystem"`       // "maven", "gradle", or "auto"
	VerboseLogging    bool           `json:"verboseLogging"`    // Enable verbose/debug logging for build tools
	GitBranch         string         `json:"gitBranch"`         // Current git branch (if service is a git repo)
	GitHasUncommitted bool           `json:"gitHasUncommitted"` // Has uncommitted changes
	GitCommitsAhead   int            `json:"gitCommitsAhead"`   // Commits ahead of remote
	GitCommitsBehind  int            `json:"gitCommitsBehind"`  // Commits behind remote
	GitIsClean        bool           `json:"gitIsClean"`        // No uncommitted changes and in sync
	// Eureka instance overrides injected as environment variables at startup.
	// Both are unset by default, leaving registration to the service's own
	// configuration; see the injection in operations.go startService.
	EurekaPreferIPAddress *bool               `json:"eurekaPreferIpAddress,omitempty"` // nil = do not override
	EurekaHostname        string              `json:"eurekaHostname,omitempty"`        // "" = do not override
	EnvVars               map[string]EnvVar   `json:"envVars"`
	Cmd                   *exec.Cmd           `json:"-"`
	Logs                  []LogEntry          `json:"logs"`
	Mutex                 sync.RWMutex        `json:"-"`
	CPUPercent            float64             `json:"cpuPercent"`
	MemoryUsage           uint64              `json:"memoryUsage"` // in bytes
	MemoryPercent         float32             `json:"memoryPercent"`
	DiskUsage             uint64              `json:"diskUsage"` // in bytes
	NetworkRx             uint64              `json:"networkRx"` // bytes received
	NetworkTx             uint64              `json:"networkTx"` // bytes transmitted
	Metrics               ServiceMetrics      `json:"metrics"`
	Dependencies          []ServiceDependency `json:"dependencies"`
	DependentOn           []string            `json:"dependentOn"`  // Services that depend on this one
	StartupDelay          time.Duration       `json:"startupDelay"` // Delay before starting after dependencies
}

// FailureReason is a classified explanation of why a service failed to start,
// recovered from its build and startup output.
type FailureReason struct {
	// Code identifies the kind of failure, for the UI to branch on.
	Code string `json:"code"`
	// Summary is a single line naming the cause, suitable for a service card.
	Summary string `json:"summary"`
	// Detail is the output line the diagnosis was drawn from.
	Detail string `json:"detail"`
	// Suggestion is the action that resolves it.
	Suggestion string    `json:"suggestion"`
	DetectedAt time.Time `json:"detectedAt"`
}
