package services

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// BuildSystemType represents the type of build system
type BuildSystemType string

const (
	BuildSystemMaven  BuildSystemType = "maven"
	BuildSystemGradle BuildSystemType = "gradle"
	BuildSystemAuto   BuildSystemType = "auto"
)

// BuildSystemCommands holds the commands for each build system
type BuildSystemCommands struct {
	Start         string
	StartWithOpts string
	Clean         string
	Test          string
	Package       string
}

// GetBuildSystemCommands returns the appropriate commands for the build system
func GetBuildSystemCommands(buildSystem BuildSystemType) BuildSystemCommands {
	switch buildSystem {
	case BuildSystemMaven:
		return BuildSystemCommands{
			Start:         "./mvnw spring-boot:run",
			StartWithOpts: "./mvnw spring-boot:run -Dspring-boot.run.jvmArguments=\"%s\"",
			Clean:         "./mvnw clean",
			Test:          "./mvnw test",
			Package:       "./mvnw package",
		}
	case BuildSystemGradle:
		return BuildSystemCommands{
			Start:         "./gradlew bootRun",
			StartWithOpts: "./gradlew bootRun -Dspring-boot.run.jvmArguments=\"%s\"",
			Clean:         "./gradlew clean",
			Test:          "./gradlew test",
			Package:       "./gradlew build",
		}
	default:
		// For auto detection, we'll detect at runtime
		return BuildSystemCommands{}
	}
}

// DetectBuildSystem automatically detects the build system for a given directory
func DetectBuildSystem(serviceDir string) BuildSystemType {
	// Check for Maven files
	mavenFiles := []string{"pom.xml", "mvnw", "mvnw.cmd"}
	for _, file := range mavenFiles {
		if _, err := os.Stat(filepath.Join(serviceDir, file)); err == nil {
			return BuildSystemMaven
		}
	}

	// Check for Gradle files
	gradleFiles := []string{"build.gradle", "build.gradle.kts", "gradlew", "gradlew.bat", "settings.gradle", "settings.gradle.kts"}
	for _, file := range gradleFiles {
		if _, err := os.Stat(filepath.Join(serviceDir, file)); err == nil {
			return BuildSystemGradle
		}
	}

	// Default to Maven if nothing is detected
	return BuildSystemMaven
}

// GetEffectiveBuildSystem returns the actual build system to use
// If buildSystem is "auto", it will detect the build system
func GetEffectiveBuildSystem(serviceDir, buildSystem string) BuildSystemType {
	if buildSystem == "" || buildSystem == string(BuildSystemAuto) {
		return DetectBuildSystem(serviceDir)
	}
	return BuildSystemType(buildSystem)
}

// HasMavenWrapper checks if the service directory has Maven wrapper
func HasMavenWrapper(serviceDir string) bool {
	wrapperFiles := []string{"mvnw", "mvnw.cmd"}
	for _, file := range wrapperFiles {
		if _, err := os.Stat(filepath.Join(serviceDir, file)); err == nil {
			return true
		}
	}
	return false
}

// HasGradleWrapper checks if the service directory has Gradle wrapper
func HasGradleWrapper(serviceDir string) bool {
	wrapperFiles := []string{"gradlew", "gradlew.bat"}
	for _, file := range wrapperFiles {
		if _, err := os.Stat(filepath.Join(serviceDir, file)); err == nil {
			return true
		}
	}
	return false
}

// GetStartCommand returns the appropriate start command for the service
func GetStartCommand(serviceDir, buildSystem string, javaOpts string, extraEnv string) (string, error) {
	effectiveBuildSystem := GetEffectiveBuildSystem(serviceDir, buildSystem)
	commands := GetBuildSystemCommands(effectiveBuildSystem)

	var baseCommand string
	if javaOpts != "" {
		baseCommand = commands.StartWithOpts
		if effectiveBuildSystem == BuildSystemMaven {
			baseCommand = strings.Replace(baseCommand, "%s", javaOpts, 1)
		} else if effectiveBuildSystem == BuildSystemGradle {
			// For Gradle, we use different JVM args format
			baseCommand = strings.Replace(commands.Start, "bootRun", "bootRun --args=\""+javaOpts+"\"", 1)
		}
	} else {
		baseCommand = commands.Start
	}

	// Construct the full command with directory change and environment
	var fullCommand string
	if extraEnv != "" {
		if javaOpts != "" && effectiveBuildSystem == BuildSystemMaven {
			// For Maven, also set MAVEN_OPTS
			fullCommand = "cd " + serviceDir + " && " + extraEnv + " MAVEN_OPTS=\"" + javaOpts + "\" " + baseCommand
		} else if javaOpts != "" && effectiveBuildSystem == BuildSystemGradle {
			// For Gradle, set GRADLE_OPTS
			fullCommand = "cd " + serviceDir + " && " + extraEnv + " GRADLE_OPTS=\"" + javaOpts + "\" " + baseCommand
		} else {
			fullCommand = "cd " + serviceDir + " && " + extraEnv + " " + baseCommand
		}
	} else {
		if javaOpts != "" && effectiveBuildSystem == BuildSystemMaven {
			fullCommand = "cd " + serviceDir + " && MAVEN_OPTS=\"" + javaOpts + "\" " + baseCommand
		} else if javaOpts != "" && effectiveBuildSystem == BuildSystemGradle {
			fullCommand = "cd " + serviceDir + " && GRADLE_OPTS=\"" + javaOpts + "\" " + baseCommand
		} else {
			fullCommand = "cd " + serviceDir + " && " + baseCommand
		}
	}

	return fullCommand, nil
}

// ValidateBuildSystem ensures the detected build system has the required files
func ValidateBuildSystem(serviceDir string, buildSystem BuildSystemType) bool {
	switch buildSystem {
	case BuildSystemMaven:
		// Check for pom.xml
		if _, err := os.Stat(filepath.Join(serviceDir, "pom.xml")); err == nil {
			return true
		}
	case BuildSystemGradle:
		// Check for build.gradle or build.gradle.kts
		gradleFiles := []string{"build.gradle", "build.gradle.kts"}
		for _, file := range gradleFiles {
			if _, err := os.Stat(filepath.Join(serviceDir, file)); err == nil {
				return true
			}
		}
	}
	return false
}

// GenerateMavenWrapper generates Maven wrapper files in the specified service directory
func GenerateMavenWrapper(serviceDir string) error {
	// Validate service directory
	log.Printf("[DEBUG] Generating Maven wrapper in directory: %s", serviceDir)
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return fmt.Errorf("service directory %s does not exist", serviceDir)
	}

	// Validate pom.xml presence
	pomPath := filepath.Join(serviceDir, "pom.xml")
	if _, err := os.Stat(pomPath); os.IsNotExist(err) {
		return fmt.Errorf("no pom.xml found in %s, not a valid Maven project", serviceDir)
	}

	// Determine the Maven executable name based on OS
	mvnExecutable := "mvn"
	if runtime.GOOS == "windows" {
		mvnExecutable = "mvn.cmd" // Windows uses mvn.cmd
	}
	log.Printf("[DEBUG] Looking for Maven executable: %s", mvnExecutable)

	// Attempt to find mvn in PATH (mimics `command -v mvn` or `where mvn`)
	mvnPath, err := exec.LookPath(mvnExecutable)
	if err != nil {
		// Log the PATH for debugging
		currentPath := os.Getenv("PATH")
		log.Printf("[DEBUG] PATH environment variable: %s", currentPath)

		// Check MAVEN_HOME environment variable
		mavenHome := os.Getenv("MAVEN_HOME")
		if mavenHome != "" {
			potentialPath := filepath.Join(mavenHome, "bin", mvnExecutable)
			if _, err := os.Stat(potentialPath); err == nil {
				mvnPath = potentialPath
				log.Printf("[DEBUG] Found %s via MAVEN_HOME at: %s", mvnExecutable, mvnPath)
			} else {
				log.Printf("[DEBUG] MAVEN_HOME set to %s, but %s not found", mavenHome, potentialPath)
			}
		}

		// If still not found, return error
		if mvnPath == "" {
			return fmt.Errorf("%s not found in PATH or MAVEN_HOME; please install Maven or add it to PATH (e.g., /opt/homebrew/bin for Homebrew on macOS) or set MAVEN_HOME", mvnExecutable)
		}
	}
	log.Printf("[DEBUG] Using %s path: %s", mvnExecutable, mvnPath)

	// Verify executable permissions
	if info, err := os.Stat(mvnPath); err != nil || info.Mode().Perm()&0111 == 0 {
		return fmt.Errorf("%s path %s is not executable or inaccessible: %v", mvnExecutable, mvnPath, err)
	}

	// Execute mvn -N io.takari:maven:wrapper
	cmd := exec.Command(mvnPath, "-N", "io.takari:maven:wrapper")

	fmt.Printf("[DEBUG] Executing command: %s\n", strings.Join(cmd.Args, " "))

	cmd.Dir = serviceDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to generate Maven wrapper in %s: %v, output: %s", serviceDir, err, string(output))
	}
	log.Printf("[DEBUG] Maven wrapper generated successfully in %s, output: %s", serviceDir, string(output))
	return nil
}

// GenerateGradleWrapper creates Gradle wrapper files
func GenerateGradleWrapper(serviceDir string) error {
	log.Printf("[INFO] Generating Gradle wrapper in %s", serviceDir)

	// Use gradle wrapper command to generate wrapper files
	cmd := exec.Command("gradle", "wrapper")
	cmd.Dir = serviceDir

	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[WARN] Failed to generate Gradle wrapper: %v - output: %s", err, string(output))
		return fmt.Errorf("failed to generate Gradle wrapper: %w", err)
	}

	// Make gradlew executable on Unix systems
	gradlewPath := filepath.Join(serviceDir, "gradlew")
	if err := os.Chmod(gradlewPath, 0755); err != nil {
		log.Printf("[WARN] Failed to make gradlew executable: %v", err)
	}

	log.Printf("[INFO] Successfully generated Gradle wrapper in %s", serviceDir)
	log.Printf("[DEBUG] Gradle wrapper output: %s", string(output))
	return nil
}

// ValidateWrapperIntegrity checks if wrapper files are valid and not corrupted
func ValidateWrapperIntegrity(serviceDir string, buildSystem BuildSystemType) (bool, error) {
	switch buildSystem {
	case BuildSystemMaven:
		return validateMavenWrapperIntegrity(serviceDir)
	case BuildSystemGradle:
		return validateGradleWrapperIntegrity(serviceDir)
	default:
		return false, fmt.Errorf("unsupported build system: %s", buildSystem)
	}
}

// validateMavenWrapperIntegrity checks Maven wrapper files
func validateMavenWrapperIntegrity(serviceDir string) (bool, error) {
	// Check if JAVA_HOME is set
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		return false, fmt.Errorf("JAVA_HOME environment variable is not set. Please set JAVA_HOME to fix wrapper validation.\n\nTo set JAVA_HOME:\n• For bash (~/.bashrc): export JAVA_HOME=/path/to/java\n• For zsh (~/.zshrc): export JAVA_HOME=/path/to/java\n• For fish (~/.config/fish/config.fish): set -x JAVA_HOME /path/to/java\n• Then restart your terminal or run: source ~/.bashrc (or ~/.zshrc)\n\nTo find Java location:\n• macOS: /usr/libexec/java_home\n• Linux: which java or whereis java")
	}

	requiredFiles := []string{"mvnw", ".mvn/wrapper/maven-wrapper.properties"}

	for _, file := range requiredFiles {
		path := filepath.Join(serviceDir, file)
		if _, err := os.Stat(path); err != nil {
			return false, fmt.Errorf("missing or corrupted wrapper file: %s", file)
		}

		// Check if mvnw is executable and not empty
		if file == "mvnw" {
			info, err := os.Stat(path)
			if err != nil {
				return false, err
			}
			if info.Size() == 0 {
				return false, fmt.Errorf("mvnw file is empty/corrupted")
			}
		}
	}

	// Try to run wrapper to test if it works
	cmd := exec.Command("./mvnw", "--version")
	cmd.Dir = serviceDir
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("wrapper execution test failed: %w", err)
	}

	return true, nil
}

// validateGradleWrapperIntegrity checks Gradle wrapper files
func validateGradleWrapperIntegrity(serviceDir string) (bool, error) {
	// Check if JAVA_HOME is set
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		return false, fmt.Errorf("JAVA_HOME environment variable is not set. Please set JAVA_HOME to fix wrapper validation.\n\nTo set JAVA_HOME:\n• For bash (~/.bashrc): export JAVA_HOME=/path/to/java\n• For zsh (~/.zshrc): export JAVA_HOME=/path/to/java\n• For fish (~/.config/fish/config.fish): set -x JAVA_HOME /path/to/java\n• Then restart your terminal or run: source ~/.bashrc (or ~/.zshrc)\n\nTo find Java location:\n• macOS: /usr/libexec/java_home\n• Linux: which java or whereis java")
	}

	requiredFiles := []string{"gradlew", "gradle/wrapper/gradle-wrapper.properties"}

	for _, file := range requiredFiles {
		path := filepath.Join(serviceDir, file)
		if _, err := os.Stat(path); err != nil {
			return false, fmt.Errorf("missing or corrupted wrapper file: %s", file)
		}

		// Check if gradlew is executable and not empty
		if file == "gradlew" {
			info, err := os.Stat(path)
			if err != nil {
				return false, err
			}
			if info.Size() == 0 {
				return false, fmt.Errorf("gradlew file is empty/corrupted")
			}
		}
	}

	// Try to run wrapper to test if it works
	cmd := exec.Command("./gradlew", "--version")
	cmd.Dir = serviceDir
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("wrapper execution test failed: %w", err)
	}

	return true, nil
}

// RepairWrapper generates/repairs wrapper files for the detected build system
func RepairWrapper(serviceDir string) error {
	buildSystem := DetectBuildSystem(serviceDir)

	switch buildSystem {
	case BuildSystemMaven:
		return GenerateMavenWrapper(serviceDir)
	case BuildSystemGradle:
		return GenerateGradleWrapper(serviceDir)
	default:
		return fmt.Errorf("unable to detect build system for wrapper repair")
	}
}

// EnsureMavenWrapper creates or updates Maven wrapper files if they don't exist or are outdated
// Deprecated: Use GenerateMavenWrapper instead
func EnsureMavenWrapper(serviceDir string) error {
	return GenerateMavenWrapper(serviceDir)
}
