package services

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/zechtz/nest-up/internal/models"
)

// TopologyService manages service topology analysis and visualization
type TopologyService struct {
	serviceManager *Manager
}

// NewTopologyService creates a new topology service
func NewTopologyService(serviceManager *Manager) *TopologyService {
	return &TopologyService{
		serviceManager: serviceManager,
	}
}

// GenerateTopology analyzes services and generates topology visualization data
func (ts *TopologyService) GenerateTopology() (*models.ServiceTopology, error) {
	servicesSlice := ts.serviceManager.GetServices()
	services := make([]*models.Service, 0, len(servicesSlice))
	for i := range servicesSlice {
		services = append(services, &servicesSlice[i])
	}
	return ts.generateTopologyForServices(services)
}

// GenerateTopologyForProfile generates topology for services in a specific profile
func (ts *TopologyService) GenerateTopologyForProfile(profileServicesJson string) (*models.ServiceTopology, error) {
	// Parse the profile services JSON to get the list of service names
	var profileServiceNames []string
	if err := json.Unmarshal([]byte(profileServicesJson), &profileServiceNames); err != nil {
		return nil, fmt.Errorf("failed to parse profile services: %v", err)
	}

	// Create a map for quick lookup of profile services
	profileServicesMap := make(map[string]bool)
	for _, serviceName := range profileServiceNames {
		profileServicesMap[serviceName] = true
	}

	// Get all services and filter only those in the profile
	allServices := ts.serviceManager.GetServices()
	var profileServices []*models.Service
	for i := range allServices {
		if profileServicesMap[allServices[i].Name] {
			profileServices = append(profileServices, &allServices[i])
		}
	}

	return ts.generateTopologyForServices(profileServices)
}

// generateTopologyForServices is the core topology generation logic
func (ts *TopologyService) generateTopologyForServices(services []*models.Service) (*models.ServiceTopology, error) {
	
	// Create topology nodes for services
	nodes := make([]models.TopologyNode, 0)
	
	// Add service nodes
	for _, service := range services {
		node := models.TopologyNode{
			ID:           service.Name,
			Name:         service.Name,
			Type:         "service",
			Status:       service.Status,
			HealthStatus: service.HealthStatus,
			Port:         service.Port,
			Metadata: map[string]interface{}{
				"description":    service.Description,
				"cpuPercent":     service.CPUPercent,
				"memoryUsage":    service.MemoryUsage,
				"memoryPercent":  service.MemoryPercent,
				"uptime":         service.Uptime,
				"lastStarted":    service.LastStarted,
			},
		}
		nodes = append(nodes, node)
	}
	
	// Add infrastructure nodes
	infraNodes := ts.getInfrastructureNodes()
	nodes = append(nodes, infraNodes...)
	
	// Analyze service dependencies from logs
	connections := ts.analyzeServiceDependencies(services)
	
	// Calculate positions using force-directed layout
	ts.calculateNodePositions(nodes, connections)
	
	topology := &models.ServiceTopology{
		Services:    nodes,
		Connections: connections,
		Generated:   time.Now(),
	}
	
	return topology, nil
}

// getInfrastructureNodes returns nodes for databases, message queues, etc.
func (ts *TopologyService) getInfrastructureNodes() []models.TopologyNode {
	nodes := make([]models.TopologyNode, 0)
	
	// Database node
	dbNode := models.TopologyNode{
		ID:           "postgresql",
		Name:         "PostgreSQL",
		Type:         "database",
		Status:       "external",
		HealthStatus: "unknown",
		Port:         5432,
		Metadata: map[string]interface{}{
			"host": "localhost",
			"type": "database",
		},
	}
	nodes = append(nodes, dbNode)
	
	// Redis node
	redisNode := models.TopologyNode{
		ID:           "redis",
		Name:         "Redis Cache",
		Type:         "database",
		Status:       "external",
		HealthStatus: "unknown",
		Port:         6379,
		Metadata: map[string]interface{}{
			"host": "localhost",
			"type": "cache",
		},
	}
	nodes = append(nodes, redisNode)
	
	// RabbitMQ node
	rabbitNode := models.TopologyNode{
		ID:           "rabbitmq",
		Name:         "RabbitMQ",
		Type:         "external",
		Status:       "external",
		HealthStatus: "unknown",
		Port:         5672,
		Metadata: map[string]interface{}{
			"host": "localhost",
			"type": "message_queue",
		},
	}
	nodes = append(nodes, rabbitNode)
	
	return nodes
}

// analyzeServiceDependencies analyzes service-to-service dependencies from logs
func (ts *TopologyService) analyzeServiceDependencies(services []*models.Service) []models.Connection {
	connections := make([]models.Connection, 0)
	
	for _, service := range services {
		deps := ts.analyzeServiceLogs(service)
		for _, dep := range deps {
			connectionType := "http"
			description := fmt.Sprintf("%s communicates with %s", service.Name, dep)
			
			// Determine connection type
			if dep == "postgresql" || dep == "redis" {
				connectionType = "database"
				description = fmt.Sprintf("%s stores data in %s", service.Name, dep)
			} else if dep == "rabbitmq" {
				connectionType = "message_queue"
				description = fmt.Sprintf("%s sends messages via %s", service.Name, dep)
			}
			
			// Determine connection status
			status := "inactive"
			if service.Status == "running" {
				status = "active"
			}
			
			connection := models.Connection{
				Source:      service.Name,
				Target:      dep,
				Type:        connectionType,
				Status:      status,
				Description: description,
			}
			connections = append(connections, connection)
		}
	}
	
	return connections
}

// calculateNodePositions uses a simple force-directed layout algorithm
func (ts *TopologyService) calculateNodePositions(nodes []models.TopologyNode, connections []models.Connection) {
	if len(nodes) == 0 {
		return
	}
	
	// Initialize positions
	centerX, centerY := 400.0, 300.0
	radius := 200.0
	
	for i := range nodes {
		angle := 2 * math.Pi * float64(i) / float64(len(nodes))
		
		// Special positioning for different node types
		switch nodes[i].Type {
		case "service":
			// Services in a circle
			nodes[i].Position = &models.NodePosition{
				X: centerX + radius*math.Cos(angle),
				Y: centerY + radius*math.Sin(angle),
			}
		case "database":
			// Databases on the right side
			nodes[i].Position = &models.NodePosition{
				X: centerX + radius*1.5,
				Y: centerY + float64(i-len(nodes)/2)*60,
			}
		case "external":
			// External services on the left side
			nodes[i].Position = &models.NodePosition{
				X: centerX - radius*1.5,
				Y: centerY + float64(i-len(nodes)/2)*60,
			}
		}
	}
	
	// Adjust positions based on service hierarchy
	ts.adjustHierarchicalPositions(nodes)
}

// adjustHierarchicalPositions positions services based on their architectural layers
func (ts *TopologyService) adjustHierarchicalPositions(nodes []models.TopologyNode) {
	centerX, centerY := 400.0, 300.0
	
	// Define service layers
	layers := map[string]int{
		"EUREKA":     1, // Infrastructure layer
		"CONFIG":     1,
		"CACHE":      1,
		"GATEWAY":    2, // Gateway layer
		"UAA":        3, // Security layer
		"APP":        4, // Application layer
		"CONTRACT":   4,
		"DSMS":       4,
		"postgresql": 0, // Data layer
		"redis":      0,
		"rabbitmq":   0,
	}
	
	layerY := map[int]float64{
		0: centerY + 200, // Data layer at bottom
		1: centerY - 100, // Infrastructure at top
		2: centerY - 50,  // Gateway
		3: centerY,       // Security
		4: centerY + 100, // Applications
	}
	
	// Count services per layer
	layerCounts := make(map[int]int)
	for _, node := range nodes {
		if layer, exists := layers[node.ID]; exists {
			layerCounts[layer]++
		}
	}
	
	// Position nodes within their layers
	layerOffsets := make(map[int]int)
	for i := range nodes {
		nodeID := nodes[i].ID
		if layer, exists := layers[nodeID]; exists {
			count := layerCounts[layer]
			offset := layerOffsets[layer]
			
			// Calculate X position within the layer
			spacing := 150.0
			startX := centerX - (float64(count-1)*spacing)/2
			
			nodes[i].Position = &models.NodePosition{
				X: startX + float64(offset)*spacing,
				Y: layerY[layer],
			}
			
			layerOffsets[layer]++
		}
	}
}

// analyzeServiceLogs analyzes service logs for dependency hints
func (ts *TopologyService) analyzeServiceLogs(service *models.Service) []string {
	dependencies := make([]string, 0)
	
	// Patterns to look for in logs
	patterns := map[string]*regexp.Regexp{
		"postgresql": regexp.MustCompile(`(?i)(postgresql|postgres|jdbc:postgresql)`),
		"redis":      regexp.MustCompile(`(?i)(redis|jedis)`),
		"rabbitmq":   regexp.MustCompile(`(?i)(rabbitmq|amqp)`),
		"eureka":     regexp.MustCompile(`(?i)(eureka|discovery)`),
	}
	
	// Check recent logs
	for _, log := range service.Logs {
		message := strings.ToLower(log.Message)
		for dep, pattern := range patterns {
			if pattern.MatchString(message) {
				// Avoid duplicates
				found := false
				for _, existing := range dependencies {
					if existing == dep {
						found = true
						break
					}
				}
				if !found {
					dependencies = append(dependencies, dep)
				}
			}
		}
	}
	
	return dependencies
}