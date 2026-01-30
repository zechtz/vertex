package services

import (
	"fmt"
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
	gradlePath := filepath.Join(serviceDir, "build.gradle")

	// Check for Maven project (pom.xml)
	if _, err := os.Stat(pomPath); err == nil {
		return sm.handleMavenLombok(pomPath, serviceName)
	}

	// Check for Gradle project (build.gradle)
	if _, err := os.Stat(gradlePath); err == nil {
		return sm.handleGradleLombok(gradlePath, serviceName)
	}

	log.Printf("[INFO] Neither pom.xml nor build.gradle found for service %s, skipping Lombok check", serviceName)
	return nil
}

// handleMavenLombok handles Lombok configuration for Maven projects
func (sm *Manager) handleMavenLombok(pomPath, serviceName string) error {
	// Read pom.xml content
	content, err := os.ReadFile(pomPath)
	if err != nil {
		return fmt.Errorf("failed to read pom.xml for service %s: %w", serviceName, err)
	}

	pomContent := string(content)

	// Check if service uses Lombok
	if !strings.Contains(pomContent, "org.projectlombok") {
		log.Printf("[INFO] Service %s doesn't use Lombok in pom.xml, skipping compatibility check", serviceName)
		return nil
	}

	log.Printf("[INFO] Checking Lombok compatibility for Maven service %s", serviceName)

	// Check if annotationProcessorPaths is already configured
	if strings.Contains(pomContent, "annotationProcessorPaths") {
		log.Printf("[INFO] Service %s already has Lombok annotation processor configured in pom.xml", serviceName)
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

// handleGradleLombok handles Lombok configuration for Gradle projects
func (sm *Manager) handleGradleLombok(gradlePath, serviceName string) error {
	// Read build.gradle content
	content, err := os.ReadFile(gradlePath)
	if err != nil {
		return fmt.Errorf("failed to read build.gradle for service %s: %w", serviceName, err)
	}

	gradleContent := string(content)

	// Check if service uses Lombok
	if !strings.Contains(gradleContent, "org.projectlombok:lombok") && !strings.Contains(gradleContent, "io.freefair.lombok") {
		log.Printf("[INFO] Service %s doesn't use Lombok in build.gradle, skipping compatibility check", serviceName)
		return nil
	}

	log.Printf("[INFO] Checking Lombok compatibility for Gradle service %s", serviceName)

	// Check if annotationProcessor is already configured for Lombok
	if strings.Contains(gradleContent, "annotationProcessor 'org.projectlombok:lombok'") {
		log.Printf("[INFO] Service %s already has Lombok annotation processor configured in build.gradle", serviceName)
		return nil
	}

	// Add Lombok annotation processor
	log.Printf("[INFO] Adding Lombok annotation processor to build.gradle for service %s", serviceName)
	return sm.addLombokAnnotationProcessorToGradle(gradlePath, gradleContent, serviceName)
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

// addLombokAnnotationProcessorToGradle adds Lombok annotation processor to build.gradle
func (sm *Manager) addLombokAnnotationProcessorToGradle(gradlePath, gradleContent, serviceName string) error {
	// 1. Get Lombok version and prepare the line
	lombokVersion := sm.extractLombokVersionFromGradle(gradleContent)
	// Using a simpler string for the check to avoid whitespace issues
	processorDependency := "org.projectlombok:lombok"
	newLine := fmt.Sprintf("\n    annotationProcessor 'org.projectlombok:lombok:%s'", lombokVersion)

	// 2. Find all dependencies blocks
	// This regex captures: [0:1] full match, [2:3] "dependencies {", [4:5] internal content, [6:7] "}"
	depsRegex := regexp.MustCompile(`(?s)(dependencies\s*\{)(.*?)(\})`)
	matches := depsRegex.FindAllStringSubmatchIndex(gradleContent, -1)

	if len(matches) == 0 {
		// No dependencies block at all, append a new one at the end
		gradleContent += fmt.Sprintf("\ndependencies {%s\n}\n", newLine)
	} else {
		targetIndex := -1

		for _, m := range matches {
			// Check the text before the match
			prefix := gradleContent[:m[0]]

			// If open and closed braces are equal, we are at the ROOT level
			// This successfully skips the 'buildscript { dependencies { ... } }' block
			if strings.Count(prefix, "{") == strings.Count(prefix, "}") {
				targetIndex = m[0]

				// Check if Lombok processor already exists in THIS specific block
				blockContent := gradleContent[m[4]:m[5]]
				if strings.Contains(blockContent, processorDependency) && strings.Contains(blockContent, "annotationProcessor") {
					return nil // Already configured, exit early
				}

				// Inject the new line at the start of the block content
				newBlockContent := newLine + blockContent
				gradleContent = gradleContent[:m[4]] + newBlockContent + gradleContent[m[5]:]
				break
			}
		}

		// If we found dependencies but they were all nested (unlikely),
		// targetIndex remains -1 and we fall back to appending
		if targetIndex == -1 {
			gradleContent += fmt.Sprintf("\ndependencies {%s\n}\n", newLine)
		}
	}

	return sm.writeGradleFile(gradlePath, gradleContent, serviceName)
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

// extractLombokVersionFromGradle extracts the Lombok version from build.gradle
func (sm *Manager) extractLombokVersionFromGradle(gradleContent string) string {
	// Try to find explicit version first
	// Matches: compile 'org.projectlombok:lombok:1.18.30' or implementation 'org.projectlombok:lombok:1.18.30'
	versionRegex := regexp.MustCompile(`(compile|implementation)\s*['"]org\.projectlombok:lombok:([^'"]+)['"]`)
	matches := versionRegex.FindStringSubmatch(gradleContent)

	if len(matches) > 2 {
		return matches[2]
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
	if err := os.WriteFile(pomPath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("failed to write updated pom.xml for service %s: %w", serviceName, err)
	}

	log.Printf("[INFO] Successfully updated Lombok configuration in pom.xml for service %s", serviceName)
	return nil
}

// writeGradleFile writes the modified content back to build.gradle with backup
func (sm *Manager) writeGradleFile(gradlePath, newContent, serviceName string) error {
	// Create backup
	backupPath := gradlePath + ".backup"
	if err := sm.createBackup(gradlePath, backupPath); err != nil {
		log.Printf("[WARN] Failed to create backup for %s: %v", serviceName, err)
	}

	// Write new content
	if err := os.WriteFile(gradlePath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("failed to write updated build.gradle for service %s: %w", serviceName, err)
	}

	log.Printf("[INFO] Successfully updated Lombok configuration in build.gradle for service %s", serviceName)
	return nil
}

// createBackup creates a backup of the original file
func (sm *Manager) createBackup(sourcePath, backupPath string) error {
	// Don't create backup if it already exists
	if _, err := os.Stat(backupPath); err == nil {
		return nil
	}

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	return os.WriteFile(backupPath, content, 0o644)
}

// restorePomBackup restores pom.xml from backup if something goes wrong
func (sm *Manager) restorePomBackup(pomPath, serviceName string) error {
	backupPath := pomPath + ".backup"

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found for service %s", serviceName)
	}

	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup for service %s: %w", serviceName, err)
	}

	if err := os.WriteFile(pomPath, content, 0o644); err != nil {
		return fmt.Errorf("failed to restore backup for service %s: %w", serviceName, err)
	}

	log.Printf("[INFO] Restored pom.xml backup for service %s", serviceName)
	return nil
}

// restoreGradleBackup restores build.gradle from backup if something goes wrong
func (sm *Manager) restoreGradleBackup(gradlePath, serviceName string) error {
	backupPath := gradlePath + ".backup"

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found for service %s", serviceName)
	}

	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup for service %s: %w", serviceName, err)
	}

	if err := os.WriteFile(gradlePath, content, 0o644); err != nil {
		return fmt.Errorf("failed to restore backup for service %s: %w", serviceName, err)
	}

	log.Printf("[INFO] Restored build.gradle backup for service %s", serviceName)
	return nil
}
