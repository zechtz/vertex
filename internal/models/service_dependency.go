package models

import (
	"time"
)

type ServiceMetrics struct {
	ResponseTimes []ResponseTime `json:"responseTimes"`
	ErrorRate     float64        `json:"errorRate"`
	RequestCount  uint64         `json:"requestCount"`
	LastChecked   time.Time      `json:"lastChecked"`
}

type ServiceDependency struct {
	ServiceName   string        `json:"serviceName"`   // Name of the dependent service
	Type          string        `json:"type"`          // "hard", "soft", "optional"
	HealthCheck   bool          `json:"healthCheck"`   // Whether to check health before considering ready
	Timeout       time.Duration `json:"timeout"`       // Max time to wait for dependency
	RetryInterval time.Duration `json:"retryInterval"` // Interval between dependency checks
	Required      bool          `json:"required"`      // Whether this dependency is required for startup
	Description   string        `json:"description"`   // Human-readable description
}
