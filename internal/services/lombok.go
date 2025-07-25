// Package services
package services

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CheckAndFixLombokCompatibility checks if a Java service has proper Lombok configuration
// and automatically fixes it if needed (exported for handlers)
func (sm *Manager) CheckAndFixLombokCompatibility(serviceDir string, serviceName string) error {
	return sm.checkAndFixLombokCompatibility(serviceDir, serviceName)
}

// checkAndFixLombokCompatibility checks if a Java service has proper Lombok configuration
// and automatically fixes it if needed
func (sm *Manager) checkAndFixLombokCompatibility(serviceDir string, serviceName string) error {
	pomPath := filepath.Join(serviceDir, "pom.xml")
	
	// Check if pom.xml exists
	if _, err := os.Stat(pomPath); os.IsNotExist(err) {
		log.Printf("[INFO] No pom.xml found for service %s, skipping Lombok check", serviceName)
		return nil
	}
	
	// Read pom.xml content
	content, err := ioutil.ReadFile(pomPath)
	if err != nil {
		return fmt.Errorf("failed to read pom.xml for service %s: %w", serviceName, err)
	}
	
	pomContent := string(content)
	
	// Check if service uses Lombok
	if !strings.Contains(pomContent, "org.projectlombok") {
		log.Printf("[INFO] Service %s doesn't use Lombok, skipping compatibility check", serviceName)
		return nil
	}
	
	log.Printf("[INFO] Checking Lombok compatibility for service %s", serviceName)
	
	// Check if annotationProcessorPaths is already configured
	if strings.Contains(pomContent, "annotationProcessorPaths") {
		log.Printf("[INFO] Service %s already has Lombok annotation processor configured", serviceName)
		return nil
	}
	
	// Check if maven-compiler-plugin exists
	compilerPluginRegex := regexp.MustCompile(`(?s)<plugin>\s*<groupId>org\.apache\.maven\.plugins</groupId>\s*<artifactId>maven-compiler-plugin</artifactId>.*?</plugin>`)
	match := compilerPluginRegex.FindString(pomContent)
	
	if match != "" {
		// Maven compiler plugin exists, add annotationProcessorPaths to it
		log.Printf("[INFO] Adding Lombok annotation processor to existing maven-compiler-plugin for service %s", serviceName)
		return sm.addAnnotationProcessorToExistingPlugin(pomPath, pomContent, serviceName)
	} else {
		// Maven compiler plugin doesn't exist, add it completely
		log.Printf("[INFO] Adding maven-compiler-plugin with Lombok annotation processor for service %s", serviceName)
		return sm.addCompilerPluginWithLombok(pomPath, pomContent, serviceName)
	}
}

// addAnnotationProcessorToExistingPlugin adds annotationProcessorPaths to existing maven-compiler-plugin
func (sm *Manager) addAnnotationProcessorToExistingPlugin(pomPath, pomContent, serviceName string) error {
	// Find the configuration section in maven-compiler-plugin
	configRegex := regexp.MustCompile(`(?s)(<plugin>\s*<groupId>org\.apache\.maven\.plugins</groupId>\s*<artifactId>maven-compiler-plugin</artifactId>.*?<configuration>.*?)(</configuration>\s*</plugin>)`)
	
	// Get Lombok version from the dependency
	lombokVersion := sm.extractLombokVersion(pomContent)
	
	annotationProcessorConfig := fmt.Sprintf(`
          <annotationProcessorPaths>
            <path>
              <groupId>org.projectlombok</groupId>
              <artifactId>lombok</artifactId>
              <version>%s</version>
            </path>
          </annotationProcessorPaths>`, lombokVersion)
	
	newContent := configRegex.ReplaceAllString(pomContent, "${1}"+annotationProcessorConfig+"\n        ${2}")
	
	if newContent == pomContent {
		return fmt.Errorf("failed to modify maven-compiler-plugin configuration for service %s", serviceName)
	}
	
	return sm.writePomFile(pomPath, newContent, serviceName)
}

// addCompilerPluginWithLombok adds a complete maven-compiler-plugin with Lombok support
func (sm *Manager) addCompilerPluginWithLombok(pomPath, pomContent, serviceName string) error {
	// Find the end of the last plugin in the plugins section
	pluginsRegex := regexp.MustCompile(`(?s)(.*</plugin>\s*)(</plugins>)`)
	
	// Get Lombok version from the dependency
	lombokVersion := sm.extractLombokVersion(pomContent)
	
	compilerPlugin := fmt.Sprintf(`            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-compiler-plugin</artifactId>
                <version>3.11.0</version>
                <configuration>
                    <source>17</source>
                    <target>17</target>
                    <annotationProcessorPaths>
                        <path>
                            <groupId>org.projectlombok</groupId>
                            <artifactId>lombok</artifactId>
                            <version>%s</version>
                        </path>
                    </annotationProcessorPaths>
                </configuration>
            </plugin>
        `, lombokVersion)
	
	newContent := pluginsRegex.ReplaceAllString(pomContent, "${1}"+compilerPlugin+"${2}")
	
	if newContent == pomContent {
		return fmt.Errorf("failed to add maven-compiler-plugin for service %s", serviceName)
	}
	
	return sm.writePomFile(pomPath, newContent, serviceName)
}

// extractLombokVersion extracts the Lombok version from pom.xml
func (sm *Manager) extractLombokVersion(pomContent string) string {
	// Try to find explicit version first
	versionRegex := regexp.MustCompile(`<groupId>org\.projectlombok</groupId>\s*<artifactId>lombok</artifactId>\s*<version>([^<]+)</version>`)
	matches := versionRegex.FindStringSubmatch(pomContent)
	
	if len(matches) > 1 {
		return matches[1]
	}
	
	// Default to a known compatible version
	return "1.18.30"
}

// writePomFile writes the modified content back to pom.xml with backup
func (sm *Manager) writePomFile(pomPath, newContent, serviceName string) error {
	// Create backup
	backupPath := pomPath + ".backup"
	if err := sm.createBackup(pomPath, backupPath); err != nil {
		log.Printf("[WARN] Failed to create backup for %s: %v", serviceName, err)
	}
	
	// Write new content
	if err := ioutil.WriteFile(pomPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated pom.xml for service %s: %w", serviceName, err)
	}
	
	log.Printf("[INFO] Successfully updated Lombok configuration for service %s", serviceName)
	return nil
}

// createBackup creates a backup of the original pom.xml
func (sm *Manager) createBackup(sourcePath, backupPath string) error {
	// Don't create backup if it already exists
	if _, err := os.Stat(backupPath); err == nil {
		return nil
	}
	
	content, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	
	return ioutil.WriteFile(backupPath, content, 0644)
}

// restorePomBackup restores pom.xml from backup if something goes wrong
func (sm *Manager) restorePomBackup(pomPath, serviceName string) error {
	backupPath := pomPath + ".backup"
	
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found for service %s", serviceName)
	}
	
	content, err := ioutil.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup for service %s: %w", serviceName, err)
	}
	
	if err := ioutil.WriteFile(pomPath, content, 0644); err != nil {
		return fmt.Errorf("failed to restore backup for service %s: %w", serviceName, err)
	}
	
	log.Printf("[INFO] Restored pom.xml backup for service %s", serviceName)
	return nil
}