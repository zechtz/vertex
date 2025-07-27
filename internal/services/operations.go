// Package services
package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/zechtz/nest-up/internal/models"
)

var logLevelRegex = regexp.MustCompile(`(?i)(INFO|WARN|ERROR|DEBUG|TRACE)`)

// WaitForServiceReady waits for a service to be fully running and healthy
func (sm *Manager) WaitForServiceReady(serviceName string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second) // Check every second for faster detection
	defer ticker.Stop()

	log.Printf("[INFO] Waiting for service %s to be ready...", serviceName)

	// Add a small initial delay to let the service start up
	time.Sleep(2 * time.Second)

	for {
		sm.mutex.RLock()
		service, exists := sm.services[serviceName]
		sm.mutex.RUnlock()

		if !exists {
			return fmt.Errorf("service %s not found", serviceName)
		}

		service.Mutex.RLock()
		status := service.Status
		healthStatus := service.HealthStatus
		service.Mutex.RUnlock()

		// Optimized readiness criteria for faster startup
		// Focus on basic availability rather than full health for quicker startup
		serviceNameUpper := strings.ToUpper(serviceName)
		
		if status == "running" {
			// For critical infrastructure services, wait for at least "starting" health status
			// This means the service is up and initializing, which is sufficient for dependents to start
			if serviceNameUpper == "EUREKA" || serviceNameUpper == "CONFIG" || serviceNameUpper == "NEST-REGISTRY-SERVER" || serviceNameUpper == "NEST-CONFIG-SERVER" {
				if healthStatus == "healthy" || healthStatus == "starting" {
					log.Printf("[INFO] Critical service %s is ready (status: %s, health: %s)", serviceName, status, healthStatus)
					return nil
				} else if healthStatus == "running" {
					// For very fast startup, accept "running" health status after a short delay
					// This allows dependent services to start while this service is still initializing
					log.Printf("[INFO] Critical service %s is running and responsive, proceeding (status: %s, health: %s)", serviceName, status, healthStatus)
					return nil
				} else {
					log.Printf("[DEBUG] Critical service %s is running but not yet initialized (health: %s), continuing to wait...", serviceName, healthStatus)
				}
			} else {
				// Non-critical services just need to be running
				log.Printf("[INFO] Service %s is ready (status: %s, health: %s)", serviceName, status, healthStatus)
				return nil
			}
		}

		// Check if service failed to start
		if status == "stopped" {
			return fmt.Errorf("service %s failed to start or stopped unexpectedly", serviceName)
		}

		log.Printf("[DEBUG] Service %s not ready yet (status: %s, health: %s), waiting...", serviceName, status, healthStatus)

		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for service %s to be ready (status: %s, health: %s)", serviceName, status, healthStatus)
		case <-ticker.C:
			// Continue polling
		}
	}
}

func (sm *Manager) StartService(serviceName string) error {
	sm.mutex.RLock()
	service, exists := sm.services[serviceName]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	log.Printf("[INFO] Starting service %s (without dependency checks)", serviceName)

	return sm.startService(service)
}

func (sm *Manager) StopService(serviceName string) error {
	sm.mutex.RLock()
	service, exists := sm.services[serviceName]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	return sm.stopService(service)
}

func (sm *Manager) RestartService(serviceName string) error {
	sm.mutex.RLock()
	service, exists := sm.services[serviceName]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	// Stop the service first
	if service.Status == "running" {
		if err := sm.stopService(service); err != nil {
			return fmt.Errorf("failed to stop service for restart: %w", err)
		}
		// Wait a moment for cleanup
		time.Sleep(2 * time.Second)
	}

	// Start the service
	return sm.startService(service)
}

func (sm *Manager) StartAllServices() error {
	// Get all services and sort by order field
	services := make([]*models.Service, 0, len(sm.services))
	sm.mutex.RLock()
	for _, service := range sm.services {
		services = append(services, service)
	}
	sm.mutex.RUnlock()
	
	// Sort by the service order field
	sort.Slice(services, func(i, j int) bool {
		return services[i].Order < services[j].Order
	})
	
	// Extract service names in order
	orderedNames := make([]string, 0, len(services))
	for _, service := range services {
		orderedNames = append(orderedNames, service.Name)
	}

	log.Printf("[INFO] Starting services in configured order: %v", orderedNames)

	go func() {
		for _, serviceName := range orderedNames {
			sm.mutex.RLock()
			service, exists := sm.services[serviceName]
			sm.mutex.RUnlock()

			if !exists {
				log.Printf("[WARN] Service %s not found, skipping", serviceName)
				continue
			}

			service.Mutex.RLock()
			status := service.Status
			isEnabled := service.IsEnabled
			service.Mutex.RUnlock()

			if status != "running" && isEnabled {
				log.Printf("[INFO] Starting service %s (order %d) and waiting for it to be ready...", serviceName, service.Order)
				
				// Start the service
				if err := sm.StartService(serviceName); err != nil {
					log.Printf("[ERROR] Failed to start service %s: %v", serviceName, err)
					continue
				}

				// Wait for the service to be ready before starting the next one
				// Optimized timeout based on service type for faster startup
				var timeout time.Duration
				switch strings.ToUpper(serviceName) {
				case "EUREKA", "NEST-REGISTRY-SERVER":
					timeout = 90 * time.Second // Registry services typically start fastest
				case "CONFIG", "NEST-CONFIG-SERVER":
					timeout = 2 * time.Minute // Config services need more time to load configurations
				case "CACHE", "NEST-CACHE":
					timeout = 60 * time.Second // Cache services are usually quick
				case "GATEWAY", "NEST-GATEWAY":
					timeout = 90 * time.Second // Gateways need time to discover services
				default:
					timeout = 2 * time.Minute // Default for other services
				}
				if err := sm.WaitForServiceReady(serviceName, timeout); err != nil {
					log.Printf("[ERROR] Service %s did not become ready within timeout: %v", serviceName, err)
					log.Printf("[WARN] Continuing with next service despite %s not being ready", serviceName)
					continue
				}

				log.Printf("[INFO] Service %s is ready, proceeding to next service", serviceName)
			} else if status == "running" {
				log.Printf("[INFO] Service %s (order %d) is already running, skipping", serviceName, service.Order)
			} else {
				log.Printf("[INFO] Service %s (order %d) is disabled, skipping", serviceName, service.Order)
			}
		}
		log.Printf("[INFO] Completed sequential service startup in configured order")
	}()

	return nil
}

func (sm *Manager) StopAllServices() error {
	// Get all services and sort by reverse order (stop in reverse)
	sm.mutex.RLock()
	services := make([]*models.Service, 0, len(sm.services))
	for _, service := range sm.services {
		services = append(services, service)
	}
	sm.mutex.RUnlock()

	sort.Slice(services, func(i, j int) bool {
		return services[i].Order > services[j].Order
	})

	go func() {
		for _, service := range services {
			service.Mutex.RLock()
			status := service.Status
			service.Mutex.RUnlock()

			if status == "running" {
				if err := sm.stopService(service); err != nil {
					log.Printf("Failed to stop service %s: %v", service.Name, err)
					continue
				}
				time.Sleep(1 * time.Second) // Brief wait between stops
			}
		}
	}()

	return nil
}

func (sm *Manager) startService(service *models.Service) error {
	service.Mutex.Lock()
	defer service.Mutex.Unlock()

	if service.Status == "running" {
		return fmt.Errorf("service %s is already running", service.Name)
	}

	serviceDir := filepath.Join(sm.config.ProjectsDir, service.Dir)
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return fmt.Errorf("service directory does not exist: %s", serviceDir)
	}

	// Check and fix Lombok compatibility before starting the service
	if err := sm.checkAndFixLombokCompatibility(serviceDir, service.Name); err != nil {
		log.Printf("[WARN] Lombok compatibility check failed for service %s: %v", service.Name, err)
		// Continue with startup - the error might not be critical
	}

	// Get global environment variables
	globalEnvVars, err := sm.GetGlobalEnvVars()
	if err != nil {
		log.Printf("Warning: Failed to load global environment variables for service %s: %v", service.Name, err)
		globalEnvVars = make(map[string]string)
	}

	// Build Maven command using Maven wrapper
	var cmd *exec.Cmd
	
	// Prepare JVM arguments for Spring Boot application
	jvmArgs := ""
	if service.JavaOpts != "" {
		jvmArgs = fmt.Sprintf(" -Dspring-boot.run.jvmArguments=\"%s\"", service.JavaOpts)
	}
	
	if service.ExtraEnv != "" {
		cmdStr := fmt.Sprintf("cd %s && %s ./mvnw spring-boot:run -e -X%s", serviceDir, service.ExtraEnv, jvmArgs)
		if service.JavaOpts != "" {
			// Also set MAVEN_OPTS for Maven JVM in case it's needed
			cmdStr = fmt.Sprintf("cd %s && %s MAVEN_OPTS=\"%s\" ./mvnw spring-boot:run -e -X%s", serviceDir, service.ExtraEnv, service.JavaOpts, jvmArgs)
		}
		cmd = exec.Command("bash", "-c", cmdStr)
	} else {
		if service.JavaOpts != "" {
			// Set both MAVEN_OPTS and spring-boot.run.jvmArguments
			cmd = exec.Command("bash", "-c", fmt.Sprintf("cd %s && MAVEN_OPTS=\"%s\" ./mvnw spring-boot:run%s", serviceDir, service.JavaOpts, jvmArgs))
		} else {
			cmd = exec.Command("bash", "-c", fmt.Sprintf("cd %s && ./mvnw spring-boot:run", serviceDir))
		}
	}

	// log the cmd
	// fmt.Printf("The command to run is: %s", cmd)

	// Set process group for proper cleanup
	SetProcessGroup(cmd)

	// Set environment variables for the process
	cmd.Env = os.Environ() // Start with current environment

	// Apply Java Home override if set
	if sm.config.JavaHomeOverride != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("JAVA_HOME=%s", sm.config.JavaHomeOverride))
		cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s/bin:%s", sm.config.JavaHomeOverride, os.Getenv("PATH")))
	}

	// Add global environment variables
	for key, value := range globalEnvVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Add service-specific environment variables
	for key, envVar := range service.EnvVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, envVar.Value))
	}

	// Add service-specific configuration (simplified)
	switch service.Name {
	case "nest-uaa":
		if dbName, exists := globalEnvVars["DB_NAME_UAA"]; exists {
			cmd.Env = append(cmd.Env, fmt.Sprintf("DB_NAME=%s", dbName))
		}
		if port, exists := globalEnvVars["SERVICE_PORT_UAA"]; exists {
			cmd.Env = append(cmd.Env, fmt.Sprintf("SERVICE_PORT=%s", port))
		}
	case "nest-app":
		if dbName, exists := globalEnvVars["DB_NAME_APP"]; exists {
			cmd.Env = append(cmd.Env, fmt.Sprintf("DB_NAME=%s", dbName))
		}
		if port, exists := globalEnvVars["SERVICE_PORT_APP"]; exists {
			cmd.Env = append(cmd.Env, fmt.Sprintf("SERVICE_PORT=%s", port))
		}
	case "nest-contract":
		if dbName, exists := globalEnvVars["DB_NAME_CONTRACT"]; exists {
			cmd.Env = append(cmd.Env, fmt.Sprintf("DB_NAME=%s", dbName))
		}
		if port, exists := globalEnvVars["SERVICE_PORT_CONTRACT"]; exists {
			cmd.Env = append(cmd.Env, fmt.Sprintf("SERVICE_PORT=%s", port))
		}
	case "nest-dsms":
		if dbName, exists := globalEnvVars["DB_NAME_DSMS"]; exists {
			cmd.Env = append(cmd.Env, fmt.Sprintf("DB_NAME=%s", dbName))
		}
		if port, exists := globalEnvVars["SERVICE_PORT_DSMS"]; exists {
			cmd.Env = append(cmd.Env, fmt.Sprintf("SERVICE_PORT=%s", port))
		}
	case "nest-gateway":
		if port, exists := globalEnvVars["SERVICE_PORT_GATEWAY"]; exists {
			cmd.Env = append(cmd.Env, fmt.Sprintf("SERVICE_PORT=%s", port))
		}
	case "nest-config-server":
		if port, exists := globalEnvVars["SERVICE_PORT_CONFIG"]; exists {
			cmd.Env = append(cmd.Env, fmt.Sprintf("SERVICE_PORT=%s", port))
		}
	case "nest-registry-server":
		if port, exists := globalEnvVars["SERVICE_PORT_REGISTRY"]; exists {
			cmd.Env = append(cmd.Env, fmt.Sprintf("SERVICE_PORT=%s", port))
		}
	}

	// Ensure ACTIVE_PROFILE is set for all services
	if activeProfile, exists := globalEnvVars["ACTIVE_PROFILE"]; exists {
		cmd.Env = append(cmd.Env, fmt.Sprintf("ACTIVE_PROFILE=%s", activeProfile))
		cmd.Env = append(cmd.Env, fmt.Sprintf("SPRING_PROFILES_ACTIVE=%s", activeProfile))
	}

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	// fmt.Printf("RUNNING THE COMMAND:\n%s\n", cmd)
	fmt.Printf("THE_CURRENT_ENVs:\n%s\n", cmd)

	service.Status = "running"
	service.HealthStatus = "starting"
	service.PID = cmd.Process.Pid
	service.Cmd = cmd
	service.LastStarted = time.Now()
	service.Logs = []models.LogEntry{}

	// Start reading logs
	go sm.readLogs(service, stdout)
	go sm.readLogs(service, stderr)

	// Monitor process completion
	go func() {
		err := cmd.Wait()
		service.Mutex.Lock()
		defer service.Mutex.Unlock()

		if err != nil {
			log.Printf("Service %s exited with error: %v", service.Name, err)
			
			// Check if it's a compilation error that might be related to Lombok
			if strings.Contains(err.Error(), "compilation") || strings.Contains(err.Error(), "cannot find symbol") {
				log.Printf("[INFO] Compilation error detected for service %s, attempting pom.xml backup restoration", service.Name)
				pomPath := filepath.Join(serviceDir, "pom.xml")
				if restoreErr := sm.restorePomBackup(pomPath, service.Name); restoreErr != nil {
					log.Printf("[WARN] Failed to restore backup for service %s: %v", service.Name, restoreErr)
				}
			}
		} else {
			log.Printf("Service %s exited successfully", service.Name)
		}

		service.Status = "stopped"
		service.HealthStatus = "unknown"
		service.PID = 0
		service.Cmd = nil
		service.Uptime = ""
		sm.updateServiceInDB(service)
		sm.broadcastUpdate(service)
	}()

	// Update database and broadcast
	sm.updateServiceInDB(service)
	sm.broadcastUpdate(service)

	log.Printf("Started service %s with PID %d", service.Name, service.PID)
	return nil
}

func (sm *Manager) stopService(service *models.Service) error {
	service.Mutex.Lock()
	defer service.Mutex.Unlock()

	if service.Status != "running" || service.Cmd == nil {
		return fmt.Errorf("service %s is not running", service.Name)
	}

	log.Printf("Stopping service %s (PID: %d)", service.Name, service.PID)

	// Get the process group ID and kill the entire group
	if pgid, err := GetProcessGroup(service.Cmd.Process.Pid); err != nil {
		log.Printf("Failed to get process group for %s: %v", service.Name, err)
		// Fallback to killing just the main process
		if err := service.Cmd.Process.Kill(); err != nil {
			return err
		}
	} else {
		// Kill the entire process group
		if err := KillProcessGroup(pgid); err != nil {
			log.Printf("Failed to terminate process group for %s: %v", service.Name, err)
			// Try force kill if regular kill fails
			if err := ForceKillProcessGroup(pgid); err != nil {
				log.Printf("Failed to force kill process group for %s: %v", service.Name, err)
				// Fallback to killing just the main process
				if err := service.Cmd.Process.Kill(); err != nil {
					return err
				}
			}
		}
	}

	service.Status = "stopped"
	service.HealthStatus = "unknown"
	service.PID = 0
	service.Cmd = nil
	service.Uptime = ""

	// Update database
	sm.updateServiceInDB(service)
	sm.broadcastUpdate(service)
	return nil
}

func (sm *Manager) readLogs(service *models.Service, pipe io.Reader) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()

		logEntry := parseLogLine(line)

		service.Mutex.Lock()
		// Keep in-memory logs for immediate access (last 1000 entries)
		service.Logs = append(service.Logs, logEntry)
		if len(service.Logs) > 1000 {
			service.Logs = service.Logs[len(service.Logs)-1000:]
		}
		service.Mutex.Unlock()

		// Store log entry in database for persistent storage
		if err := sm.db.StoreLogEntry(service.Name, logEntry); err != nil {
			log.Printf("Failed to store log entry for service %s: %v", service.Name, err)
		}

		// Broadcast the new log entry
		sm.broadcastLogEntry(service.Name, logEntry)
	}
}

func parseLogLine(line string) models.LogEntry {
	match := logLevelRegex.FindStringSubmatch(line)
	level := "INFO" // Default level
	if len(match) > 1 {
		level = strings.ToUpper(match[1])
	}

	return models.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Level:     level,
		Message:   line,
	}
}

func (sm *Manager) ClearLogs(serviceName string) error {
	sm.mutex.RLock()
	service, exists := sm.services[serviceName]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	service.Mutex.Lock()
	service.Logs = []models.LogEntry{}
	service.Mutex.Unlock()

	sm.broadcastUpdate(service)
	return nil
}
