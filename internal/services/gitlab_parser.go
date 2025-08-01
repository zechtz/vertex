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
	// Use global projects directory for backward compatibility
	return sm.ParseGitLabCIWithProjectsDir(serviceUUID, sm.config.ProjectsDir)
}

// ParseGitLabCIWithProjectsDir parses .gitlab-ci.yml files in service directories to extract Maven library installations with a custom projects directory
func (sm *Manager) ParseGitLabCIWithProjectsDir(serviceUUID, projectsDir string) (*models.GitLabCIConfig, error) {
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

	serviceDir := filepath.Join(projectsDir, service.Dir)
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

		// Handle YAML list items (lines starting with "- ")
		if strings.HasPrefix(line, "- ") {
			line = strings.TrimSpace(line[2:]) // Remove "- " prefix
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
	// Use global projects directory for backward compatibility
	return sm.InstallLibrariesWithProjectsDir(serviceUUID, libraries, sm.config.ProjectsDir)
}

// InstallLibrariesWithProjectsDir runs the Maven library installation commands for a specific service with a custom projects directory
func (sm *Manager) InstallLibrariesWithProjectsDir(serviceUUID string, libraries []models.LibraryInstallation, projectsDir string) error {
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

	serviceDir := filepath.Join(projectsDir, service.Dir)

	// If libraries are provided, use them; otherwise, parse .gitlab-ci.yml
	var libsToInstall []models.LibraryInstallation
	if len(libraries) > 0 {
		libsToInstall = libraries
	} else {
		config, err := sm.ParseGitLabCIWithProjectsDir(serviceUUID, projectsDir)
		if err != nil {
			return fmt.Errorf("failed to parse GitLab CI config: %w", err)
		}
		if !config.HasLibraries {
			return fmt.Errorf("no libraries found to install for service UUID %s", serviceUUID)
		}
		libsToInstall = config.Libraries
	}

	log.Printf("[INFO] Installing %d libraries for service UUID %s in directory %s", len(libsToInstall), serviceUUID, serviceDir)

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

// PreviewLibraryInstallation analyzes .gitlab-ci.yml and returns a preview of libraries grouped by environment
func (sm *Manager) PreviewLibraryInstallation(serviceUUID, projectsDir string) (*models.LibraryPreview, error) {
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

	serviceDir := filepath.Join(projectsDir, service.Dir)
	gitlabCIPath := filepath.Join(serviceDir, ".gitlab-ci.yml")

	log.Printf("[DEBUG] PreviewLibraryInstallation - projectsDir: %s, service.Dir: %s, serviceDir: %s, gitlabCIPath: %s", 
		projectsDir, service.Dir, serviceDir, gitlabCIPath)

	// Check if .gitlab-ci.yml exists
	if _, err := os.Stat(gitlabCIPath); os.IsNotExist(err) {
		log.Printf("[DEBUG] .gitlab-ci.yml file does not exist at: %s", gitlabCIPath)
		return &models.LibraryPreview{
			HasLibraries:   false,
			ServiceName:    service.Name,
			ServiceID:      serviceUUID,
			GitlabCIExists: false,
			ErrorMessage:   "No .gitlab-ci.yml file found in service directory",
		}, nil
	}

	// Parse the file and extract environment-based library installations
	environments, err := sm.parseEnvironmentLibraries(gitlabCIPath)
	if err != nil {
		return &models.LibraryPreview{
			HasLibraries:   false,
			ServiceName:    service.Name,
			ServiceID:      serviceUUID,
			GitlabCIExists: true,
			ErrorMessage:   fmt.Sprintf("Failed to parse .gitlab-ci.yml: %v", err),
		}, nil
	}

	if len(environments) == 0 {
		log.Printf("[DEBUG] No environments with libraries found for service %s at path: %s", serviceUUID, gitlabCIPath)
		return &models.LibraryPreview{
			HasLibraries:   false,
			ServiceName:    service.Name,
			ServiceID:      serviceUUID,
			GitlabCIExists: true,
			ErrorMessage:   "No library installation commands found in any environment",
		}, nil
	}

	totalLibraries := 0
	for _, env := range environments {
		totalLibraries += len(env.Libraries)
	}

	return &models.LibraryPreview{
		HasLibraries:   true,
		ServiceName:    service.Name,
		ServiceID:      serviceUUID,
		Environments:   environments,
		TotalLibraries: totalLibraries,
		GitlabCIExists: true,
	}, nil
}

// parseEnvironmentLibraries parses .gitlab-ci.yml and groups library installations by environment
func (sm *Manager) parseEnvironmentLibraries(gitlabCIPath string) ([]models.EnvironmentLibraries, error) {
	file, err := os.Open(gitlabCIPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read .gitlab-ci.yml: %w", err)
	}
	defer file.Close()

	// Regular expressions for parsing
	jobNameRegex := regexp.MustCompile(`^([a-zA-Z0-9_-]+):$`)
	mvnInstallRegex := regexp.MustCompile(`mvn\s+install:install-file\s+(.+)`)
	onlyRegex := regexp.MustCompile(`^\s*only:\s*$`)
	branchRegex := regexp.MustCompile(`^\s*-\s*(.+)$`)

	var environments []models.EnvironmentLibraries
	scanner := bufio.NewScanner(file)
	
	var currentJob string
	var currentLibraries []models.LibraryInstallation
	var currentBranches []string
	var inOnlySection bool
	var inScriptSection bool

	log.Printf("[DEBUG] Starting to parse .gitlab-ci.yml file: %s", gitlabCIPath)

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		log.Printf("[DEBUG] Processing line: '%s' (job: %s, inScript: %v, inOnly: %v)", trimmedLine, currentJob, inScriptSection, inOnlySection)

		// Check for job definition (only jobs that contain "maven-build" or similar patterns)
		if matches := jobNameRegex.FindStringSubmatch(trimmedLine); matches != nil {
			jobName := matches[1]
			
			// Only treat as a job if it looks like a CI job (contains common job patterns)
			if strings.Contains(strings.ToLower(jobName), "maven-build") || 
			   strings.Contains(strings.ToLower(jobName), "build") ||
			   strings.HasSuffix(strings.ToLower(jobName), "-dev") ||
			   strings.HasSuffix(strings.ToLower(jobName), "-staging") ||
			   strings.HasSuffix(strings.ToLower(jobName), "-live") ||
			   strings.HasSuffix(strings.ToLower(jobName), "-prod") ||
			   strings.HasSuffix(strings.ToLower(jobName), "-training") {
				
				// Save previous job if it had libraries
				if currentJob != "" && len(currentLibraries) > 0 {
					env := sm.extractEnvironmentFromJobName(currentJob)
					environments = append(environments, models.EnvironmentLibraries{
						Environment: env,
						JobName:     currentJob,
						Libraries:   currentLibraries,
						Branches:    currentBranches,
					})
					log.Printf("[DEBUG] Saved job '%s' with %d libraries", currentJob, len(currentLibraries))
				}

				// Start new job
				currentJob = jobName
				currentLibraries = []models.LibraryInstallation{}
				currentBranches = []string{}
				inOnlySection = false
				inScriptSection = false
				log.Printf("[DEBUG] Starting new CI job: %s", currentJob)
				continue
			}
		}

		// Only process sections if we're in a CI job
		if currentJob != "" {
			// Check for "only:" section
			if onlyRegex.MatchString(trimmedLine) {
				inOnlySection = true
				inScriptSection = false
				log.Printf("[DEBUG] Entering 'only' section for job: %s", currentJob)
				continue
			}

			// Check for script section
			if strings.HasPrefix(trimmedLine, "script:") {
				inScriptSection = true
				inOnlySection = false
				log.Printf("[DEBUG] Entering 'script' section for job: %s", currentJob)
				continue
			}
		}

		// Parse branches in "only" section (only if we're in a CI job)
		if currentJob != "" && inOnlySection {
			if matches := branchRegex.FindStringSubmatch(trimmedLine); matches != nil {
				currentBranches = append(currentBranches, matches[1])
				log.Printf("[DEBUG] Added branch '%s' to job '%s'", matches[1], currentJob)
			}
		}

		// Parse library installations in script section (only if we're in a CI job)
		if currentJob != "" && inScriptSection {
			// Handle YAML list items (lines starting with "- ")
			originalLine := trimmedLine
			if strings.HasPrefix(trimmedLine, "- ") {
				trimmedLine = strings.TrimSpace(trimmedLine[2:])
				log.Printf("[DEBUG] Stripped YAML list prefix: '%s' -> '%s'", originalLine, trimmedLine)
			}

			// Look for mvn install:install-file commands
			if matches := mvnInstallRegex.FindStringSubmatch(trimmedLine); matches != nil {
				fullCommand := strings.TrimSpace(matches[1])
				log.Printf("[DEBUG] Found Maven install command: %s", fullCommand)
				
				// Extract library parameters
				library := sm.parseLibraryFromCommand(fullCommand, trimmedLine)
				log.Printf("[DEBUG] Parsed library: %+v", library)
				if library.GroupID != "" && library.ArtifactID != "" && library.Version != "" {
					currentLibraries = append(currentLibraries, library)
					log.Printf("[DEBUG] Added library to current job '%s': %s:%s:%s", currentJob, library.GroupID, library.ArtifactID, library.Version)
				} else {
					log.Printf("[DEBUG] Skipped incomplete library: %+v", library)
				}
			}
		}
	}

	// Don't forget the last job
	if currentJob != "" && len(currentLibraries) > 0 {
		env := sm.extractEnvironmentFromJobName(currentJob)
		environments = append(environments, models.EnvironmentLibraries{
			Environment: env,
			JobName:     currentJob,
			Libraries:   currentLibraries,
			Branches:    currentBranches,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .gitlab-ci.yml: %w", err)
	}

	return environments, nil
}

// extractEnvironmentFromJobName extracts environment name from job names like "maven-build-dev", "maven-build-staging"
func (sm *Manager) extractEnvironmentFromJobName(jobName string) string {
	// Common patterns for environment detection
	envPatterns := map[string][]string{
		"development": {"dev", "develop", "development"},
		"staging":     {"staging", "stage", "test"},
		"production":  {"prod", "production", "live"},
		"training":    {"training", "train"},
		"dr":          {"dr", "disaster", "recovery"},
		"newlive":     {"newlive", "new-live"},
	}

	jobLower := strings.ToLower(jobName)
	
	for env, patterns := range envPatterns {
		for _, pattern := range patterns {
			if strings.Contains(jobLower, pattern) {
				return env
			}
		}
	}

	// If no pattern matches, try to extract the last part after the last dash
	parts := strings.Split(jobName, "-")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}

	return "unknown"
}

// parseLibraryFromCommand parses Maven install command and extracts library information
func (sm *Manager) parseLibraryFromCommand(fullCommand, originalLine string) models.LibraryInstallation {
	fileRegex := regexp.MustCompile(`-Dfile=([^\s]+)`)
	groupIdRegex := regexp.MustCompile(`-DgroupId=([^\s]+)`)
	artifactIdRegex := regexp.MustCompile(`-DartifactId=([^\s]+)`)
	versionRegex := regexp.MustCompile(`-Dversion=([^\s]+)`)
	packagingRegex := regexp.MustCompile(`-Dpackaging=([^\s]+)`)

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

	return models.LibraryInstallation{
		File:       file,
		GroupID:    groupId,
		ArtifactID: artifactId,
		Version:    version,
		Packaging:  packaging,
		Command:    "mvn install:install-file " + fullCommand,
	}
}
