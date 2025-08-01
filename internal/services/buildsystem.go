package services

import (
	"os"
	"path/filepath"
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
