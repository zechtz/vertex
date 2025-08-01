package services

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/zechtz/vertex/internal/models"
)

// ParseGitLabCI parses .gitlab-ci.yml files in service directories to extract Maven library installations
func (sm *Manager) ParseGitLabCI(serviceUUID string) (*models.GitLabCIConfig, error) {
	// Validate UUID
	if _, err := uuid.Parse(serviceUUID); err != nil {
		return nil, fmt.Errorf("invalid service UUID: %s", serviceUUID)
	}

	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service UUID %s not found", serviceUUID)
	}

	serviceDir := filepath.Join(sm.config.ProjectsDir, service.Dir)
	gitlabCIPath := filepath.Join(serviceDir, ".gitlab-ci.yml")

	config := &models.GitLabCIConfig{
		ServiceID:    serviceUUID,
		Libraries:    []models.LibraryInstallation{},
		HasLibraries: false,
	}

	// Check if .gitlab-ci.yml exists
	if _, err := os.Stat(gitlabCIPath); os.IsNotExist(err) {
		config.ErrorMessage = fmt.Sprintf("No .gitlab-ci.yml found in %s", serviceDir)
		return config, nil
	}

	// Read and parse the file
	file, err := os.Open(gitlabCIPath)
	if err != nil {
		config.ErrorMessage = fmt.Sprintf("Failed to read .gitlab-ci.yml: %v", err)
		return config, nil
	}
	defer file.Close()

	// Regular expression to match mvn install:install-file commands
	mvnInstallRegex := regexp.MustCompile(`mvn\s+install:install-file\s+(.+)`)
	fileRegex := regexp.MustCompile(`-Dfile=([^\s]+)`)
	groupIdRegex := regexp.MustCompile(`-DgroupId=([^\s]+)`)
	artifactIdRegex := regexp.MustCompile(`-DartifactId=([^\s]+)`)
	versionRegex := regexp.MustCompile(`-Dversion=([^\s]+)`)
	packagingRegex := regexp.MustCompile(`-Dpackaging=([^\s]+)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Look for mvn install:install-file commands
		if matches := mvnInstallRegex.FindStringSubmatch(line); matches != nil {
			fullCommand := strings.TrimSpace(matches[1])

			// Extract parameters
			var file, groupId, artifactId, version, packaging string

			if fileMatch := fileRegex.FindStringSubmatch(fullCommand); fileMatch != nil {
				file = fileMatch[1]
			}

			if groupIdMatch := groupIdRegex.FindStringSubmatch(fullCommand); groupIdMatch != nil {
				groupId = groupIdMatch[1]
			}

			if artifactIdMatch := artifactIdRegex.FindStringSubmatch(fullCommand); artifactIdMatch != nil {
				artifactId = artifactIdMatch[1]
			}

			if versionMatch := versionRegex.FindStringSubmatch(fullCommand); versionMatch != nil {
				version = versionMatch[1]
			}

			if packagingMatch := packagingRegex.FindStringSubmatch(fullCommand); packagingMatch != nil {
				packaging = packagingMatch[1]
			} else {
				packaging = "jar" // Default to jar if not specified
			}

			// Only add if we have the essential parameters
			if file != "" && groupId != "" && artifactId != "" && version != "" {
				library := models.LibraryInstallation{
					File:       file,
					GroupID:    groupId,
					ArtifactID: artifactId,
					Version:    version,
					Packaging:  packaging,
					Command:    "mvn install:install-file " + fullCommand,
				}

				config.Libraries = append(config.Libraries, library)
				config.HasLibraries = true

				log.Printf("[INFO] Found library installation for UUID %s: %s:%s:%s", serviceUUID, groupId, artifactId, version)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		config.ErrorMessage = fmt.Sprintf("Error reading .gitlab-ci.yml: %v", err)
		return config, nil
	}

	if len(config.Libraries) == 0 {
		config.ErrorMessage = "No Maven library installation commands found in .gitlab-ci.yml"
	}

	return config, nil
}

// GetAllGitLabCIConfigs returns GitLab CI configurations for all services
func (sm *Manager) GetAllGitLabCIConfigs() map[string]*models.GitLabCIConfig {
	configs := make(map[string]*models.GitLabCIConfig)

	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	for serviceUUID := range sm.services {
		if config, err := sm.ParseGitLabCI(serviceUUID); err == nil {
			configs[serviceUUID] = config
		} else {
			configs[serviceUUID] = &models.GitLabCIConfig{
				ServiceID:    serviceUUID,
				Libraries:    []models.LibraryInstallation{},
				HasLibraries: false,
				ErrorMessage: err.Error(),
			}
		}
	}

	return configs
}

// InstallLibraries runs the Maven library installation commands for a specific service
func (sm *Manager) InstallLibraries(serviceUUID string, libraries []models.LibraryInstallation) error {
	// Validate UUID
	if _, err := uuid.Parse(serviceUUID); err != nil {
		return fmt.Errorf("invalid service UUID: %s", serviceUUID)
	}

	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service UUID %s not found", serviceUUID)
	}

	serviceDir := filepath.Join(sm.config.ProjectsDir, service.Dir)

	// If libraries are provided, use them; otherwise, parse .gitlab-ci.yml
	var libsToInstall []models.LibraryInstallation
	if len(libraries) > 0 {
		libsToInstall = libraries
	} else {
		config, err := sm.ParseGitLabCI(serviceUUID)
		if err != nil {
			return fmt.Errorf("failed to parse GitLab CI config: %w", err)
		}
		if !config.HasLibraries {
			return fmt.Errorf("no libraries found to install for service UUID %s", serviceUUID)
		}
		libsToInstall = config.Libraries
	}

	log.Printf("[INFO] Installing %d libraries for service UUID %s", len(libsToInstall), serviceUUID)

	for i, library := range libsToInstall {
		log.Printf("[INFO] Installing library %d/%d: %s:%s:%s",
			i+1, len(libsToInstall), library.GroupID, library.ArtifactID, library.Version)

		// Check if the library file exists
		libPath := filepath.Join(serviceDir, library.File)
		if _, err := os.Stat(libPath); os.IsNotExist(err) {
			log.Printf("[WARN] Library file not found: %s (continuing anyway)", libPath)
		}

		// Execute the Maven install command
		if err := sm.executeMavenCommand(serviceDir, library.Command); err != nil {
			return fmt.Errorf("failed to install library %s:%s:%s: %w",
				library.GroupID, library.ArtifactID, library.Version, err)
		}

		log.Printf("[INFO] Successfully installed library: %s:%s:%s",
			library.GroupID, library.ArtifactID, library.Version)
	}

	log.Printf("[INFO] Successfully installed all %d libraries for service UUID %s", len(libsToInstall), serviceUUID)
	return nil
}

// executeMavenCommand executes a Maven command in the specified directory
func (sm *Manager) executeMavenCommand(workDir, command string) error {
	// Use Maven wrapper if available, otherwise fall back to mvn
	mvnCommand := "./mvnw"
	if _, err := os.Stat(filepath.Join(workDir, "mvnw")); os.IsNotExist(err) {
		mvnCommand = "mvn"
	}

	// Replace "mvn" in the command with the appropriate Maven executable
	fullCommand := strings.Replace(command, "mvn ", mvnCommand+" ", 1)

	log.Printf("[DEBUG] Executing Maven command in %s: %s", workDir, fullCommand)

	// Use the existing Maven execution pattern from startService
	cmd := fmt.Sprintf("cd %s && %s", workDir, fullCommand)

	// Get global environment variables for Maven execution
	globalEnvVars, err := sm.GetGlobalEnvVars()
	if err != nil {
		log.Printf("[WARN] Failed to load global environment variables: %v", err)
		globalEnvVars = make(map[string]string)
	}

	// Execute the command using the same approach as service startup
	return sm.executeCommand(cmd, globalEnvVars)
}

// executeCommand executes a bash command with environment variables
func (sm *Manager) executeCommand(cmdStr string, envVars map[string]string) error {
	cmd := exec.Command("bash", "-c", cmdStr)

	// Set environment variables for the process
	cmd.Env = os.Environ() // Start with current environment

	// Apply Java Home override if set
	if sm.config.JavaHomeOverride != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("JAVA_HOME=%s", sm.config.JavaHomeOverride))
		cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s/bin:%s", sm.config.JavaHomeOverride, os.Getenv("PATH")))
	}

	// Add environment variables
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Execute the command and wait for completion
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] Command failed: %s", string(output))
		return fmt.Errorf("command execution failed: %w - output: %s", err, string(output))
	}

	log.Printf("[DEBUG] Command output: %s", string(output))
	return nil
}
