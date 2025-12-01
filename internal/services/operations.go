// Package services
package services

import (
	"bufio"
	"context"
	"encoding/json"
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

	"github.com/zechtz/vertex/internal/models"
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

		// Generic readiness criteria - all services treated equally
		if status == "running" {
			// Accept running services with healthy, starting, or running health status
			if healthStatus == "healthy" || healthStatus == "starting" || healthStatus == "running" {
				log.Printf("[INFO] Service %s is ready (status: %s, health: %s)", serviceName, status, healthStatus)
				return nil
			} else {
				log.Printf("[DEBUG] Service %s is running but not yet ready (health: %s), continuing to wait...", serviceName, healthStatus)
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
				case "EUREKA":
					timeout = 90 * time.Second // Registry services typically start fastest
				case "CONFIG":
					timeout = 2 * time.Minute // Config services need more time to load configurations
				case "CACHE":
					timeout = 60 * time.Second // Cache services are usually quick
				case "GATEWAY":
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

// StopAllServicesForProfile stops all services that belong to a specific profile
func (sm *Manager) StopAllServicesForProfile(profileServicesJson string) error {
	// Parse the profile services JSON to get the list of service UUIDs
	var profileServiceUUIDs []string
	if err := json.Unmarshal([]byte(profileServicesJson), &profileServiceUUIDs); err != nil {
		return fmt.Errorf("failed to parse profile services: %v", err)
	}

	log.Printf("[INFO] Stopping services for profile: %v", profileServiceUUIDs)

	// Create a map for quick lookup of profile services
	profileServicesMap := make(map[string]bool)
	for _, serviceUUID := range profileServiceUUIDs {
		profileServicesMap[serviceUUID] = true
	}

	// Get all services and filter only those in the profile
	sm.mutex.RLock()
	var profileServices []*models.Service
	for _, service := range sm.services {
		if profileServicesMap[service.ID] {
			profileServices = append(profileServices, service)
		}
	}
	sm.mutex.RUnlock()

	log.Printf("[INFO] Found %d services in profile to stop", len(profileServices))

	// Sort by reverse order (stop in reverse dependency order)
	sort.Slice(profileServices, func(i, j int) bool {
		return profileServices[i].Order > profileServices[j].Order
	})

	go func() {
		for _, service := range profileServices {
			service.Mutex.RLock()
			status := service.Status
			service.Mutex.RUnlock()

			if status == "running" {
				if err := sm.stopService(service); err != nil {
					log.Printf("Failed to stop service %s (profile): %v", service.Name, err)
					continue
				}
				time.Sleep(1 * time.Second) // Brief wait between stops
			}
		}
	}()

	return nil
}

// StartAllServicesForProfile starts all services that belong to a specific profile
func (sm *Manager) StartAllServicesForProfile(profileServicesJSON string, projectsDir string) error {
	// Parse the profile services JSON to get the list of service UUIDs
	var profileServiceUUIDs []string
	if err := json.Unmarshal([]byte(profileServicesJSON), &profileServiceUUIDs); err != nil {
		return fmt.Errorf("failed to parse profile services: %v", err)
	}

	log.Printf("[INFO] Starting services for profile: %v", profileServiceUUIDs)

	// Create a map for quick lookup of profile services
	profileServicesMap := make(map[string]bool)
	for _, serviceUUID := range profileServiceUUIDs {
		profileServicesMap[serviceUUID] = true
	}

	// Get all services and filter only those in the profile
	sm.mutex.RLock()
	var profileServices []*models.Service
	for _, service := range sm.services {
		if profileServicesMap[service.ID] {
			profileServices = append(profileServices, service)
		}
	}
	sm.mutex.RUnlock()

	log.Printf("[INFO] Found %d services in profile to start", len(profileServices))

	// Sort by order (start in dependency order)
	sort.Slice(profileServices, func(i, j int) bool {
		return profileServices[i].Order < profileServices[j].Order
	})

	go func() {
		for _, service := range profileServices {
			service.Mutex.RLock()
			status := service.Status
			isEnabled := service.IsEnabled
			service.Mutex.RUnlock()

			if status != "running" && isEnabled {
				log.Printf("[INFO] Starting service %s (order %d) in profile", service.Name, service.Order)
				// Use profile-aware starting if projectsDir is different from global
				globalConfig := sm.GetConfig()
				if projectsDir != globalConfig.ProjectsDir {
					if err := sm.startServiceWithProjectsDir(service, projectsDir); err != nil {
						log.Printf("Failed to start service %s (profile): %v", service.Name, err)
						continue
					}
				} else {
					if err := sm.startService(service); err != nil {
						log.Printf("Failed to start service %s (profile): %v", service.Name, err)
						continue
					}
				}
				time.Sleep(2 * time.Second) // Brief wait between starts
			} else if status == "running" {
				log.Printf("[INFO] Service %s (order %d) is already running, skipping", service.Name, service.Order)
			} else {
				log.Printf("[INFO] Service %s (order %d) is disabled, skipping", service.Name, service.Order)
			}
		}
	}()

	return nil
}

func (sm *Manager) startServiceWithProjectsDir(service *models.Service, projectsDir string) error {
	service.Mutex.Lock()
	defer service.Mutex.Unlock()

	if service.Status == "running" {
		return fmt.Errorf("service %s is already running", service.Name)
	}

	serviceDir := filepath.Join(projectsDir, service.Dir)
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return fmt.Errorf("service directory does not exist: %s", serviceDir)
	}

	log.Printf("[INFO] Starting service %s from directory: %s", service.Name, serviceDir)

	// Ensure vertex user has access to the profile's project directory
	if err := EnsureVertexUserProjectAccess(projectsDir); err != nil {
		log.Printf("[WARN] Failed to setup project directory access for service %s: %v", service.Name, err)
		// Continue with startup - this shouldn't block service startup
	}

	// Also ensure the specific service's build directory exists with proper permissions
	if err := ensureServiceBuildDirectory(serviceDir); err != nil {
		log.Printf("[WARN] Failed to setup build directory for service %s: %v", service.Name, err)
	}

	// Check and fix Lombok compatibility before starting the service
	if err := sm.checkAndFixLombokCompatibility(serviceDir, service.Name); err != nil {
		log.Printf("[WARN] Lombok compatibility check failed for service %s: %v", service.Name, err)
		// Continue with startup
	}

	// Get global environment variables
	globalEnvVars, err := sm.GetGlobalEnvVars()
	if err != nil {
		log.Printf("Warning: Failed to load global environment variables for service %s: %v", service.Name, err)
		globalEnvVars = make(map[string]string)
	}

	// Auto-detect build system
	effectiveBuildSystem := GetEffectiveBuildSystem(serviceDir, service.BuildSystem)
	log.Printf("[INFO] Using build system '%s' for service %s", effectiveBuildSystem, service.Name)

	// Ensure Maven wrapper exists for Maven projects
	if effectiveBuildSystem == BuildSystemMaven {
		if err := EnsureMavenWrapper(serviceDir); err != nil {
			log.Printf("[WARN] Failed to ensure Maven wrapper for service %s: %v", service.Name, err)
			// Continue with startup - this is not a critical failure
		}
	}

	// Get start command
	cmdString, err := GetStartCommand(serviceDir, string(effectiveBuildSystem), service.JavaOpts, service.ExtraEnv)
	if err != nil {
		return fmt.Errorf("failed to construct start command: %w", err)
	}

	// Clean up port
	if service.Port > 0 {
		log.Printf("[INFO] Checking port %d for conflicts before starting service %s", service.Port, service.Name)
		if err := CleanupPortBeforeStart(service.Port); err != nil {
			log.Printf("[WARN] Port cleanup failed for service %s: %v", service.Name, err)
		}
	}

	cmd := exec.Command("bash", "-c", cmdString)
	cmd.Dir = serviceDir
	SetProcessGroup(cmd)

	// Start with the current environment
	cmd.Env = os.Environ()

	// Build environment variables with proper precedence
	// Priority: Service-specific env vars > Profile Java Home override > Global env vars

	// Create a map to track which variables are set by service
	serviceEnvKeys := make(map[string]bool)
	for key := range service.EnvVars {
		serviceEnvKeys[key] = true
	}

	// Apply Java Home override if set (only if not overridden by service)
	if sm.config.JavaHomeOverride != "" && !serviceEnvKeys["JAVA_HOME"] {
		cmd.Env = append(cmd.Env, fmt.Sprintf("JAVA_HOME=%s", sm.config.JavaHomeOverride))
		cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s/bin:%s", sm.config.JavaHomeOverride, os.Getenv("PATH")))
	}

	// Add global environment variables (only if not overridden by service)
	for key, value := range globalEnvVars {
		if !serviceEnvKeys[key] {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Add service-specific environment variables (these take precedence)
	for key, envVar := range service.EnvVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, envVar.Value))
		// If service sets JAVA_HOME, also update PATH to use that Java
		if key == "JAVA_HOME" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s/bin:%s", envVar.Value, os.Getenv("PATH")))
		}
		// Also set SPRING_PROFILES_ACTIVE for Spring Boot if ACTIVE_PROFILE is set
		if key == "ACTIVE_PROFILE" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("SPRING_PROFILES_ACTIVE=%s", envVar.Value))
		}
	}

	// Ensure ACTIVE_PROFILE and SPRING_PROFILES_ACTIVE are set if not already set by service
	if activeProfile, exists := globalEnvVars["ACTIVE_PROFILE"]; exists {
		if !serviceEnvKeys["ACTIVE_PROFILE"] {
			cmd.Env = append(cmd.Env, fmt.Sprintf("ACTIVE_PROFILE=%s", activeProfile))
			cmd.Env = append(cmd.Env, fmt.Sprintf("SPRING_PROFILES_ACTIVE=%s", activeProfile))
		}
	}

	// Detect and log Java version being used
	logJavaVersion(cmd.Env, service.Name)

	// Log the final command and environment variables for profile services
	log.Printf("[DEBUG] Starting profile service %s with command: %s", service.Name, cmdString)
	log.Printf("[DEBUG] Working directory: %s", serviceDir)
	log.Printf("[DEBUG] Environment variables for %s:", service.Name)
	for _, env := range cmd.Env {
		if strings.Contains(env, "ACTIVE_PROFILE") || strings.Contains(env, "SPRING_PROFILES") || strings.Contains(env, "SERVICE_PORT") || strings.Contains(env, "CONFIG_") || strings.Contains(env, "JAVA_HOME") {
			log.Printf("[DEBUG]   %s", env)
		}
	}

	// Create stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	service.Status = "running"
	service.HealthStatus = "starting"
	service.LastStarted = time.Now()
	service.PID = cmd.Process.Pid
	service.Cmd = cmd
	service.Uptime = ""
	service.Logs = []models.LogEntry{}

	// Save and broadcast
	sm.updateServiceInDB(service)
	sm.broadcastUpdate(service)

	go sm.readLogs(service, stdout)
	go sm.readLogs(service, stderr)

	go func() {
		err := cmd.Wait()
		service.Mutex.Lock()
		defer service.Mutex.Unlock()

		if err != nil {
			log.Printf("Service %s exited with error: %v", service.Name, err)
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

	log.Printf("[INFO] Service %s started successfully with PID %d", service.Name, service.PID)
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

	// Ensure vertex user has access to the profile's project directory
	if err := EnsureVertexUserProjectAccess(sm.config.ProjectsDir); err != nil {
		log.Printf("[WARN] Failed to setup project directory access for service %s: %v", service.Name, err)
		// Continue with startup - this shouldn't block service startup
	}

	// Also ensure the specific service's build directory exists with proper permissions
	if err := ensureServiceBuildDirectory(serviceDir); err != nil {
		log.Printf("[WARN] Failed to setup build directory for service %s: %v", service.Name, err)
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

	// Auto-detect build system if needed and get appropriate command
	effectiveBuildSystem := GetEffectiveBuildSystem(serviceDir, service.BuildSystem)
	log.Printf("[INFO] Using build system '%s' for service %s", effectiveBuildSystem, service.Name)

	// Ensure Maven wrapper exists for Maven projects
	if effectiveBuildSystem == BuildSystemMaven {
		if err := EnsureMavenWrapper(serviceDir); err != nil {
			log.Printf("[WARN] Failed to ensure Maven wrapper for service %s: %v", service.Name, err)
			// Continue with startup - this is not a critical failure
		}
	}

	// Get the start command for the detected build system
	cmdString, err := GetStartCommand(serviceDir, string(effectiveBuildSystem), service.JavaOpts, service.ExtraEnv)
	if err != nil {
		return fmt.Errorf("failed to construct start command: %w", err)
	}

	// Clean up any processes using the service's port before starting
	if service.Port > 0 {
		log.Printf("[INFO] Checking port %d for conflicts before starting service %s", service.Port, service.Name)
		if err := CleanupPortBeforeStart(service.Port); err != nil {
			log.Printf("[WARN] Port cleanup failed for service %s: %v", service.Name, err)
			// Continue anyway - the service might still be able to start
		}
	}

	log.Printf("[INFO] Starting service %s with command: %s", service.Name, cmdString)
	cmd := exec.Command("bash", "-c", cmdString)

	// log the cmd
	// fmt.Printf("The command to run is: %s", cmd)

	// Set process group for proper cleanup
	SetProcessGroup(cmd)

	// Set environment variables for the process
	cmd.Env = os.Environ() // Start with current environment

	// Build environment variables with proper precedence
	// Priority: Service-specific env vars > Profile Java Home override > Global env vars

	// Create a map to track which variables are set by service
	serviceEnvKeys := make(map[string]bool)
	for key := range service.EnvVars {
		serviceEnvKeys[key] = true
	}

	// Apply Java Home override if set (only if not overridden by service)
	if sm.config.JavaHomeOverride != "" && !serviceEnvKeys["JAVA_HOME"] {
		cmd.Env = append(cmd.Env, fmt.Sprintf("JAVA_HOME=%s", sm.config.JavaHomeOverride))
		cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s/bin:%s", sm.config.JavaHomeOverride, os.Getenv("PATH")))
	}

	// Add global environment variables (only if not overridden by service)
	for key, value := range globalEnvVars {
		if !serviceEnvKeys[key] {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Add service-specific environment variables (these take precedence)
	for key, envVar := range service.EnvVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, envVar.Value))
		// If service sets JAVA_HOME, also update PATH to use that Java
		if key == "JAVA_HOME" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s/bin:%s", envVar.Value, os.Getenv("PATH")))
		}
		// Also set SPRING_PROFILES_ACTIVE for Spring Boot if ACTIVE_PROFILE is set
		if key == "ACTIVE_PROFILE" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("SPRING_PROFILES_ACTIVE=%s", envVar.Value))
		}
	}

	// Ensure ACTIVE_PROFILE and SPRING_PROFILES_ACTIVE are set if not already set by service
	if activeProfile, exists := globalEnvVars["ACTIVE_PROFILE"]; exists {
		if !serviceEnvKeys["ACTIVE_PROFILE"] {
			cmd.Env = append(cmd.Env, fmt.Sprintf("ACTIVE_PROFILE=%s", activeProfile))
			cmd.Env = append(cmd.Env, fmt.Sprintf("SPRING_PROFILES_ACTIVE=%s", activeProfile))
		}
	}

	// Detect and log Java version being used
	logJavaVersion(cmd.Env, service.Name)

	// Log the final command and environment variables
	// log.Printf("[DEBUG] Starting service %s with command: %s", service.Name, cmdString)
	// log.Printf("[DEBUG] Working directory: %s", serviceDir)
	// log.Printf("[DEBUG] Environment variables for %s:", service.Name)
	for _, env := range cmd.Env {
		if strings.Contains(env, "ACTIVE_PROFILE") || strings.Contains(env, "SPRING_PROFILES") || strings.Contains(env, "SERVICE_PORT") || strings.Contains(env, "CONFIG_") || strings.Contains(env, "JAVA_HOME") {
			log.Printf("[DEBUG]   %s", env)
		}
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
	// fmt.Printf("THE_CURRENT_CMD_WITH_OPTS:\n%s\n", cmd)

	service.Status = "running"
	service.HealthStatus = "starting"
	service.PID = cmd.Process.Pid
	service.Cmd = cmd
	service.LastStarted = time.Now()
	service.Logs = []models.LogEntry{}

	// Record uptime event
	uptimeTracker := GetUptimeTracker()
	uptimeTracker.RecordEvent(service.ID, "start", "running")

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

		// Record uptime event
		uptimeTracker := GetUptimeTracker()
		uptimeTracker.RecordEvent(service.ID, "stop", "stopped")

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
		if err := sm.db.StoreLogEntry(service.ID, logEntry); err != nil {
			log.Printf("Failed to store log entry for service %s: %v", service.ID, err)
		}

		// Broadcast the new log entry
		sm.broadcastLogEntry(service.ID, logEntry)
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

func (sm *Manager) ClearLogs(serviceID string) error {
	service, exists := sm.GetServiceByUUID(serviceID)
	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	// Clear logs from database first
	if err := sm.db.ClearServiceLogs(serviceID); err != nil {
		return fmt.Errorf("failed to clear logs from database: %w", err)
	}

	// Clear in-memory logs
	service.Mutex.Lock()
	service.Logs = []models.LogEntry{}
	service.Mutex.Unlock()

	sm.broadcastUpdate(service)
	return nil
}

func (sm *Manager) ClearAllLogs(serviceNames []string) map[string]string {
	results := make(map[string]string)

	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// If no specific services provided, clear all services
	if len(serviceNames) == 0 {
		// Clear all logs from database
		if _, err := sm.db.ClearAllServiceLogs([]string{}); err != nil {
			// If database clear fails, still continue with in-memory clear but log the error
			log.Printf("[ERROR] Failed to clear all logs from database: %v", err)
		}

		// Clear in-memory logs for all services
		for _, service := range sm.services {
			service.Mutex.Lock()
			service.Logs = []models.LogEntry{}
			service.Mutex.Unlock()

			sm.broadcastUpdate(service)
			results[service.Name] = "Success"
		}
	} else {
		// Clear logs for specific services by name
		var serviceIDs []string
		serviceNameToID := make(map[string]string)

		// First, collect service IDs for database operation
		for _, serviceName := range serviceNames {
			found := false
			for _, service := range sm.services {
				if service.Name == serviceName {
					serviceIDs = append(serviceIDs, service.ID)
					serviceNameToID[serviceName] = service.ID
					found = true
					break
				}
			}
			if !found {
				results[serviceName] = fmt.Sprintf("Service '%s' not found", serviceName)
			}
		}

		// Clear from database
		if len(serviceIDs) > 0 {
			dbResults, err := sm.db.ClearAllServiceLogs(serviceIDs)
			if err != nil {
				log.Printf("[ERROR] Failed to clear logs from database: %v", err)
			} else if dbResults != nil {
				// Log any database errors but don't fail the entire operation
				for serviceID, dbErr := range dbResults {
					if dbErr != nil {
						log.Printf("[ERROR] Failed to clear logs from database for service %s: %v", serviceID, dbErr)
					}
				}
			}
		}

		// Clear in-memory logs
		for _, serviceName := range serviceNames {
			if _, exists := results[serviceName]; exists {
				continue // Already marked as not found
			}

			found := false
			for _, service := range sm.services {
				if service.Name == serviceName {
					service.Mutex.Lock()
					service.Logs = []models.LogEntry{}
					service.Mutex.Unlock()

					sm.broadcastUpdate(service)
					results[serviceName] = "Success"
					found = true
					break
				}
			}
			if !found {
				results[serviceName] = fmt.Sprintf("Service '%s' not found", serviceName)
			}
		}
	}

	return results
}

// isPortEnvironmentVariable checks if an environment variable name represents a port configuration
func isPortEnvironmentVariable(key string) bool {
	portVarNames := []string{
		"ACTIVE_PORT", "SERVER_PORT", "PORT", "SERVICE_PORT", "APP_PORT",
		"HTTP_PORT", "HTTPS_PORT", "WEB_PORT", "API_PORT", "TOMCAT_PORT",
		"SPRING_SERVER_PORT", "MICROSERVICE_PORT", "APPLICATION_PORT",
	}

	keyUpper := strings.ToUpper(key)
	for _, portVar := range portVarNames {
		if keyUpper == portVar || strings.HasSuffix(keyUpper, "_"+portVar) {
			return true
		}
	}
	return false
}

// logJavaVersion detects and logs the Java version being used for a service
func logJavaVersion(env []string, serviceName string) {
	// Extract JAVA_HOME from environment
	var javaHome string
	for _, envVar := range env {
		if strings.HasPrefix(envVar, "JAVA_HOME=") {
			javaHome = strings.TrimPrefix(envVar, "JAVA_HOME=")
			break
		}
	}

	if javaHome == "" {
		log.Printf("[INFO] Service %s: No JAVA_HOME set, using system default Java", serviceName)
		return
	}

	// Try to get Java version
	javaPath := filepath.Join(javaHome, "bin", "java")
	cmd := exec.Command(javaPath, "-version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("[WARN] Service %s: JAVA_HOME=%s but failed to detect version: %v", serviceName, javaHome, err)
		return
	}

	// Parse version from output
	version := "unknown"
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		line := lines[0]
		if strings.Contains(line, "version") {
			// Extract version string between quotes
			start := strings.Index(line, `"`)
			if start != -1 {
				end := strings.Index(line[start+1:], `"`)
				if end != -1 {
					version = line[start+1 : start+1+end]
				}
			}
		}
	}

	// Determine source of JAVA_HOME
	source := "system"
	if strings.Contains(javaHome, "/.asdf/installs/java/") {
		source = "asdf"
	} else if strings.Contains(javaHome, "/.sdkman/candidates/java/") {
		source = "sdkman"
	}

	log.Printf("[INFO] Service %s: Using Java %s from %s (%s)", serviceName, version, source, javaHome)
}
