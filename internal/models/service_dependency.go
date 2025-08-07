package models

import (
	"time"
)

type ServiceMetrics struct {
	ResponseTimes []ResponseTime `json:"responseTimes"`
	ErrorRate     float64        `json:"errorRate"`
	RequestCount  uint64         `json:"requestCount"`
	LastChecked   time.Time      `json:"lastChecked"`
	// Uptime Statistics
	UptimeStats   UptimeStatistics `json:"uptimeStats"`
}

type UptimeStatistics struct {
	TotalRestarts      int           `json:"totalRestarts"`
	UptimePercentage24h float64      `json:"uptimePercentage24h"`
	UptimePercentage7d  float64      `json:"uptimePercentage7d"`
	MTBF               time.Duration `json:"mtbf"` // Mean Time Between Failures
	LastDowntime       time.Time     `json:"lastDowntime"`
	TotalDowntime24h   time.Duration `json:"totalDowntime24h"`
	TotalDowntime7d    time.Duration `json:"totalDowntime7d"`
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
