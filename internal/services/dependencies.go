package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/zechtz/vertex/internal/models"
)

// DependencyManager handles service dependency management and startup ordering
type DependencyManager struct {
	serviceManager *Manager
	mutex          sync.RWMutex
	// Track dependency check status
	dependencyStatus map[string]map[string]bool // service -> dependency -> ready
}

// NewDependencyManager creates a new dependency manager
func NewDependencyManager(serviceManager *Manager) *DependencyManager {
	return &DependencyManager{
		serviceManager:   serviceManager,
		dependencyStatus: make(map[string]map[string]bool),
	}
}

// InitializeDefaultDependencies sets up empty dependencies - users configure their own
func (dm *DependencyManager) InitializeDefaultDependencies() error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// No default dependencies - users will configure through UI/API
	services := dm.serviceManager.GetServices()
	for i := range services {
		service := &services[i]
		// Initialize with empty dependencies and no startup delay
		service.Dependencies = []models.ServiceDependency{}
		service.StartupDelay = 0

		// Update the service in the manager
		dm.updateServiceDependencies(service.Name, service.Dependencies, service.StartupDelay)
	}

	// Calculate dependent services (reverse dependencies)
	dm.calculateReverseDependencies()

	log.Printf("[INFO] Initialized empty dependencies for %d services - users can configure via UI", len(services))
	return nil
}

// updateServiceDependencies updates a service's dependencies in the service manager
func (dm *DependencyManager) updateServiceDependencies(serviceName string, dependencies []models.ServiceDependency, startupDelay time.Duration) {
	// Access the service manager's services map to update dependencies
	dm.serviceManager.mutex.Lock()
	defer dm.serviceManager.mutex.Unlock()

	if service, exists := dm.serviceManager.services[serviceName]; exists {
		service.Dependencies = dependencies
		service.StartupDelay = startupDelay
	}

	// Track this in our dependency status
	if dm.dependencyStatus[serviceName] == nil {
		dm.dependencyStatus[serviceName] = make(map[string]bool)
	}

	for _, dep := range dependencies {
		dm.dependencyStatus[serviceName][dep.ServiceName] = false
	}
}

// calculateReverseDependencies calculates which services depend on each service
func (dm *DependencyManager) calculateReverseDependencies() {
	services := dm.serviceManager.GetServices()
	reverseDeps := make(map[string][]string)

	// Build reverse dependency map
	for _, service := range services {
		for _, dep := range service.Dependencies {
			reverseDeps[dep.ServiceName] = append(reverseDeps[dep.ServiceName], service.Name)
		}
	}

	// Update DependentOn fields (this would need service manager access to update properly)
	for serviceName, dependents := range reverseDeps {
		log.Printf("[DEBUG] Service %s has dependents: %v", serviceName, dependents)
	}
}

// GetStartupOrder returns the optimal startup order based on dependencies
func (dm *DependencyManager) GetStartupOrder(serviceNames []string) ([]string, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	services := dm.serviceManager.GetServices()
	serviceMap := make(map[string]models.Service)
	for _, service := range services {
		serviceMap[service.Name] = service
	}

	// Filter requested services
	var requestedServices []models.Service
	for _, name := range serviceNames {
		if service, exists := serviceMap[name]; exists {
			requestedServices = append(requestedServices, service)
		}
	}

	// Topological sort based on dependencies
	ordered, err := dm.topologicalSort(requestedServices)
	if err != nil {
		return nil, fmt.Errorf("failed to determine startup order: %w", err)
	}

	return ordered, nil
}

// topologicalSort performs topological sorting on services based on dependencies
func (dm *DependencyManager) topologicalSort(services []models.Service) ([]string, error) {
	// Build adjacency list and in-degree count
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	serviceSet := make(map[string]bool)

	// Initialize
	for _, service := range services {
		serviceSet[service.Name] = true
		inDegree[service.Name] = 0
		graph[service.Name] = []string{}
	}

	// Build graph
	for _, service := range services {
		for _, dep := range service.Dependencies {
			if dep.Required && serviceSet[dep.ServiceName] {
				graph[dep.ServiceName] = append(graph[dep.ServiceName], service.Name)
				inDegree[service.Name]++
			}
		}
	}

	// Kahn's algorithm
	var queue []string
	var result []string

	// Find all nodes with no incoming edges
	for serviceName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, serviceName)
		}
	}

	// Process queue
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Remove edges
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycles
	if len(result) != len(services) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

// CheckDependencies checks if all dependencies for a service are ready
func (dm *DependencyManager) CheckDependencies(ctx context.Context, serviceName string) error {
	dm.mutex.RLock()
	services := dm.serviceManager.GetServices()
	dm.mutex.RUnlock()

	var targetService *models.Service
	for i, service := range services {
		if service.Name == serviceName {
			targetService = &services[i]
			break
		}
	}

	if targetService == nil {
		return fmt.Errorf("service %s not found", serviceName)
	}

	// Check each dependency
	for _, dep := range targetService.Dependencies {
		if err := dm.checkSingleDependency(ctx, dep); err != nil {
			if dep.Required {
				return fmt.Errorf("required dependency %s not ready: %w", dep.ServiceName, err)
			}
			log.Printf("[WARN] Optional dependency %s not ready: %v", dep.ServiceName, err)
		}
	}

	return nil
}

// checkSingleDependency checks if a single dependency is ready
func (dm *DependencyManager) checkSingleDependency(ctx context.Context, dep models.ServiceDependency) error {
	// Check if service is running
	services := dm.serviceManager.GetServices()
	var depService *models.Service

	for i, service := range services {
		if service.Name == dep.ServiceName {
			depService = &services[i]
			break
		}
	}

	if depService == nil {
		return fmt.Errorf("dependency service %s not found", dep.ServiceName)
	}

	if depService.Status != "running" {
		return fmt.Errorf("dependency service %s is not running (status: %s)", dep.ServiceName, depService.Status)
	}

	// Perform health check if required
	if dep.HealthCheck && depService.HealthURL != "" {
		if err := dm.performHealthCheck(ctx, depService, dep.Timeout); err != nil {
			return fmt.Errorf("health check failed for %s: %w", dep.ServiceName, err)
		}
	}

	return nil
}

// performHealthCheck performs a health check on a service
func (dm *DependencyManager) performHealthCheck(ctx context.Context, service *models.Service, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", service.HealthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// WaitForDependencies waits for all dependencies to be ready before returning
func (dm *DependencyManager) WaitForDependencies(ctx context.Context, serviceName string) error {
	services := dm.serviceManager.GetServices()
	var targetService *models.Service

	for i, service := range services {
		if service.Name == serviceName {
			targetService = &services[i]
			break
		}
	}

	if targetService == nil {
		return fmt.Errorf("service %s not found", serviceName)
	}

	log.Printf("[INFO] Waiting for dependencies of %s", serviceName)

	// Wait for each required dependency
	for _, dep := range targetService.Dependencies {
		if !dep.Required {
			continue // Skip optional dependencies
		}

		log.Printf("[INFO] Checking dependency %s for %s", dep.ServiceName, serviceName)

		// Wait with retry logic
		ticker := time.NewTicker(dep.RetryInterval)
		defer ticker.Stop()

		timeout := time.After(dep.Timeout)

		for {
			// Check if dependency is ready
			if err := dm.checkSingleDependency(ctx, dep); err == nil {
				log.Printf("[INFO] Dependency %s is ready for %s", dep.ServiceName, serviceName)
				break
			}

			// Wait for next check or timeout
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timeout:
				return fmt.Errorf("timeout waiting for dependency %s (waited %v)", dep.ServiceName, dep.Timeout)
			case <-ticker.C:
				log.Printf("[DEBUG] Retrying dependency check %s for %s", dep.ServiceName, serviceName)
			}
		}
	}

	// Add startup delay
	if targetService.StartupDelay > 0 {
		log.Printf("[INFO] Waiting %v before starting %s", targetService.StartupDelay, serviceName)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(targetService.StartupDelay):
			// Continue
		}
	}

	log.Printf("[INFO] All dependencies ready for %s", serviceName)
	return nil
}

// GetDependencyGraph returns the complete dependency graph
func (dm *DependencyManager) GetDependencyGraph() map[string][]models.ServiceDependency {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	services := dm.serviceManager.GetServices()
	graph := make(map[string][]models.ServiceDependency)

	for _, service := range services {
		if len(service.Dependencies) > 0 {
			graph[service.Name] = service.Dependencies
		}
	}

	return graph
}

// ValidateDependencies checks for circular dependencies and missing services
func (dm *DependencyManager) ValidateDependencies() error {
	services := dm.serviceManager.GetServices()
	serviceNames := make(map[string]bool)

	// Build service name set
	for _, service := range services {
		serviceNames[service.Name] = true
	}

	// Check for missing dependencies
	for _, service := range services {
		for _, dep := range service.Dependencies {
			if !serviceNames[dep.ServiceName] {
				return fmt.Errorf("service %s depends on non-existent service %s", service.Name, dep.ServiceName)
			}
		}
	}

	// Check for circular dependencies using DFS
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for _, service := range services {
		if !visited[service.Name] {
			if dm.hasCycle(service.Name, services, visited, recursionStack) {
				return fmt.Errorf("circular dependency detected involving service %s", service.Name)
			}
		}
	}

	return nil
}

// hasCycle detects cycles in the dependency graph using DFS
func (dm *DependencyManager) hasCycle(serviceName string, services []models.Service, visited, recursionStack map[string]bool) bool {
	visited[serviceName] = true
	recursionStack[serviceName] = true

	// Find the service
	var currentService *models.Service
	for i, service := range services {
		if service.Name == serviceName {
			currentService = &services[i]
			break
		}
	}

	if currentService != nil {
		// Check all dependencies
		for _, dep := range currentService.Dependencies {
			if dep.Required { // Only check required dependencies for cycles
				if !visited[dep.ServiceName] {
					if dm.hasCycle(dep.ServiceName, services, visited, recursionStack) {
						return true
					}
				} else if recursionStack[dep.ServiceName] {
					return true
				}
			}
		}
	}

	recursionStack[serviceName] = false
	return false
}
