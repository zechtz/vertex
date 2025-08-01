// Package services
package services

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ServiceFile struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	Content      string `json:"content"`
	Type         string `json:"type"`
	LastModified string `json:"lastModified"`
}

func (sm *Manager) GetServiceFiles(serviceUUID string) ([]ServiceFile, error) {
	return sm.GetServiceFilesWithProjectsDir(serviceUUID, sm.config.ProjectsDir)
}

func (sm *Manager) GetServiceFilesWithProjectsDir(serviceUUID, projectsDir string) ([]ServiceFile, error) {
	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service with UUID %s not found", serviceUUID)
	}

	// Construct the full path to service directory using the provided projects directory
	serviceDir := filepath.Join(projectsDir, service.Dir)

	log.Printf("[DEBUG] GetServiceFilesWithProjectsDir - serviceUUID: %s, projectsDir: '%s', service.Dir: '%s', serviceDir: '%s'", serviceUUID, projectsDir, service.Dir, serviceDir)

	// Check if the service directory exists
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("service directory does not exist: %s (projectsDir: '%s', service.Dir: '%s'). Please configure the projects directory in settings or ensure the service directory exists", serviceDir, projectsDir, service.Dir)
	}

	var files []ServiceFile

	// Look for configuration files in common locations
	searchPaths := []string{
		"src/main/resources",
		"src/main/resources/config",
		"config",
		".", // Current directory
	}

	for _, searchPath := range searchPaths {
		fullSearchPath := filepath.Join(serviceDir, searchPath)
		if _, err := os.Stat(fullSearchPath); os.IsNotExist(err) {
			continue
		}

		foundFiles, err := sm.findConfigFiles(fullSearchPath)
		if err != nil {
			log.Printf("[DEBUG] Could not read directory %s: %v", fullSearchPath, err)
			continue // Skip directories we can't read
		}

		files = append(files, foundFiles...)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no configuration files found in service directory: %s. Searched in: %v", serviceDir, searchPaths)
	}

	return files, nil
}

func (sm *Manager) findConfigFiles(searchDir string) ([]ServiceFile, error) {
	var files []ServiceFile

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if info.IsDir() {
			return nil
		}

		// Check if it's a configuration file
		ext := strings.ToLower(filepath.Ext(info.Name()))
		name := strings.ToLower(info.Name())

		isConfigFile := false
		fileType := "unknown"

		// Check for properties files
		if ext == ".properties" || strings.Contains(name, "application") || strings.Contains(name, "config") {
			isConfigFile = true
			fileType = "properties"
		}

		// Check for YAML files
		if ext == ".yml" || ext == ".yaml" {
			isConfigFile = true
			if ext == ".yml" {
				fileType = "yml"
			} else {
				fileType = "yaml"
			}
		}

		// Check for specific configuration files
		if name == "bootstrap.properties" || name == "application.properties" ||
			name == "bootstrap.yml" || name == "application.yml" ||
			name == "bootstrap.yaml" || name == "application.yaml" ||
			strings.HasPrefix(name, "application-") {
			isConfigFile = true
			if strings.Contains(name, ".properties") {
				fileType = "properties"
			} else {
				fileType = "yml"
			}
		}

		if isConfigFile {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return nil // Skip files we can't read
			}

			relativePath, _ := filepath.Rel(searchDir, path)

			files = append(files, ServiceFile{
				Name:         info.Name(),
				Path:         relativePath,
				Content:      string(content),
				Type:         fileType,
				LastModified: info.ModTime().Format(time.RFC3339),
			})
		}

		return nil
	})

	return files, err
}

func (sm *Manager) UpdateServiceFile(serviceUUID, filename, content string) error {
	return sm.UpdateServiceFileWithProjectsDir(serviceUUID, filename, content, sm.config.ProjectsDir)
}

func (sm *Manager) UpdateServiceFileWithProjectsDir(serviceUUID, filename, content, projectsDir string) error {
	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service with UUID %s not found", serviceUUID)
	}

	log.Printf("[DEBUG] UpdateServiceFileWithProjectsDir - serviceUUID: %s, filename: %s, projectsDir: '%s', service.Dir: '%s'", serviceUUID, filename, projectsDir, service.Dir)

	// Find the file first to get its path
	files, err := sm.GetServiceFilesWithProjectsDir(serviceUUID, projectsDir)
	if err != nil {
		return err
	}

	var targetFile *ServiceFile
	for _, file := range files {
		if file.Name == filename {
			targetFile = &file
			break
		}
	}

	if targetFile == nil {
		return fmt.Errorf("file %s not found in service %s", filename, serviceUUID)
	}

	// Construct full file path using provided projects directory
	serviceDir := filepath.Join(projectsDir, service.Dir)

	// Try to find the file in the search paths
	searchPaths := []string{
		"src/main/resources",
		"src/main/resources/config",
		"config",
		".",
	}

	var fullFilePath string
	for _, searchPath := range searchPaths {
		testPath := filepath.Join(serviceDir, searchPath, targetFile.Path)
		if _, err := os.Stat(testPath); err == nil {
			fullFilePath = testPath
			log.Printf("[DEBUG] Found file for update at: %s", fullFilePath)
			break
		}
	}

	if fullFilePath == "" {
		return fmt.Errorf("could not locate file %s for writing in service directory %s", filename, serviceDir)
	}

	// Write the content to the file
	err = ioutil.WriteFile(fullFilePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	log.Printf("[INFO] Successfully updated file %s for service %s at %s", filename, serviceUUID, fullFilePath)
	return nil
}
