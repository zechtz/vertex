package models

import (
	"time"
)

type ServiceTopology struct {
	Services    []TopologyNode `json:"services"`
	Connections []Connection   `json:"connections"`
	Generated   time.Time      `json:"generated"`
}

type TopologyNode struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Type         string         `json:"type"` // "service", "database", "external"
	Status       string         `json:"status"`
	HealthStatus string         `json:"healthStatus"`
	Port         int            `json:"port"`
	Position     *NodePosition  `json:"position,omitempty"`
	Metadata     map[string]any `json:"metadata"`
}

type Connection struct {
	Source      string `json:"source"`
	Target      string `json:"target"`
	Type        string `json:"type"`   // "http", "database", "message_queue"
	Status      string `json:"status"` // "active", "inactive", "error"
	Description string `json:"description"`
}

type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
