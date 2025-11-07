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

// JavaEnvironment holds information about the detected Java installation
type JavaEnvironment struct {
	JavaHome    string
	JavaPath    string
	Version     string
	Available   bool
	ErrorMsg    string
}

// DetectJavaEnvironment finds and validates Java installation
func DetectJavaEnvironment() *JavaEnvironment {
	env := &JavaEnvironment{}
	
	log.Println("[INFO] Detecting Java environment...")
	
	// Method 1: Check JAVA_HOME environment variable
	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		// Check if JAVA_HOME points to asdf shims directory and resolve it
		if strings.Contains(javaHome, "/.asdf/shims") {
			log.Printf("[DEBUG] JAVA_HOME points to asdf shims, attempting to resolve: %s", javaHome)
			// Try to resolve using asdf
			cmd := exec.Command("asdf", "which", "java")
			if output, err := cmd.Output(); err == nil && len(output) > 0 {
				realJavaPath := strings.TrimSpace(string(output))
				javaHome = inferJavaHome(realJavaPath)
				log.Printf("[DEBUG] Resolved JAVA_HOME from asdf to: %s", javaHome)
			}
		}

		javaPath := filepath.Join(javaHome, "bin", getJavaExecutable())
		if isExecutable(javaPath) && isWorkingJava(javaPath) {
			env.JavaHome = javaHome
			env.JavaPath = javaPath
			env.Available = true
			log.Printf("[INFO] Found working Java via JAVA_HOME: %s", javaHome)
		}
	}
	
	// Method 2: Check PATH for java executable - but validate it actually works
	if !env.Available {
		if javaPath, err := exec.LookPath("java"); err == nil {
			// Test if this Java actually works by running -version
			if isWorkingJava(javaPath) {
				env.JavaPath = javaPath
				env.JavaHome = inferJavaHome(javaPath)
				env.Available = true
				log.Printf("[INFO] Found working Java in PATH: %s", javaPath)
			} else {
				log.Printf("[WARN] Java found in PATH but not working: %s", javaPath)
			}
		}
	}
	
	// Method 3: Platform-specific default locations
	if !env.Available {
		defaultPaths := getDefaultJavaPaths()
		for _, path := range defaultPaths {
			javaPath := filepath.Join(path, "bin", getJavaExecutable())
			if isExecutable(javaPath) && isWorkingJava(javaPath) {
				env.JavaHome = path
				env.JavaPath = javaPath
				env.Available = true
				log.Printf("[INFO] Found working Java at default location: %s", path)
				break
			}
		}
	}
	
	// Get Java version if found
	if env.Available {
		env.Version = getJavaVersion(env.JavaPath)
		log.Printf("[INFO] Java version: %s", env.Version)
	} else {
		env.ErrorMsg = "Java runtime not found. Please install Java 17 or later."
		log.Printf("[WARN] %s", env.ErrorMsg)
	}
	
	return env
}

// SetupJavaEnvironment configures the current process environment for Java
func (j *JavaEnvironment) SetupJavaEnvironment() error {
	if !j.Available {
		return fmt.Errorf("Java not available: %s", j.ErrorMsg)
	}
	
	// Set JAVA_HOME if not already set
	if os.Getenv("JAVA_HOME") == "" && j.JavaHome != "" {
		os.Setenv("JAVA_HOME", j.JavaHome)
		log.Printf("[INFO] Set JAVA_HOME=%s", j.JavaHome)
	}
	
	// Ensure Java bin directory is in PATH
	currentPath := os.Getenv("PATH")
	javaBinDir := filepath.Dir(j.JavaPath)
	
	if !strings.Contains(currentPath, javaBinDir) {
		newPath := javaBinDir + string(os.PathListSeparator) + currentPath
		os.Setenv("PATH", newPath)
		log.Printf("[INFO] Added Java bin directory to PATH: %s", javaBinDir)
	}
	
	return nil
}

// GetDiagnostics returns detailed information for troubleshooting
func (j *JavaEnvironment) GetDiagnostics() map[string]interface{} {
	diag := map[string]interface{}{
		"available":    j.Available,
		"java_home":    j.JavaHome,
		"java_path":    j.JavaPath,
		"version":      j.Version,
		"error":        j.ErrorMsg,
		"current_path": os.Getenv("PATH"),
		"platform":     runtime.GOOS,
	}
	
	// Check various Java paths
	defaultPaths := getDefaultJavaPaths()
	pathChecks := make(map[string]bool)
	for _, path := range defaultPaths {
		javaPath := filepath.Join(path, "bin", getJavaExecutable())
		pathChecks[javaPath] = isExecutable(javaPath) && isWorkingJava(javaPath)
	}
	diag["path_checks"] = pathChecks
	
	return diag
}

// EnsureVertexUserProjectAccess is no longer needed since we run as the current user
// Kept for backward compatibility but does nothing
func EnsureVertexUserProjectAccess(projectsDir string) error {
	// Running as current user - no permission adjustments needed
	log.Printf("[DEBUG] Running as current user - no permission setup needed for: %s", projectsDir)
	return nil
}

// ensureServiceBuildDirectory creates build directories for the specific service
func ensureServiceBuildDirectory(serviceDir string) error {
	log.Printf("[DEBUG] Checking build directory for service at: %s", serviceDir)
	
	// Check if this is a Maven or Gradle project and create appropriate build directories
	buildDirs := []string{}
	
	if _, err := os.Stat(filepath.Join(serviceDir, "pom.xml")); err == nil {
		// Maven project - ensure target directory is writable
		targetDir := filepath.Join(serviceDir, "target")
		if err := os.MkdirAll(targetDir, 0755); err == nil {
			buildDirs = append(buildDirs, filepath.Join(targetDir, "classes"))
		}
	}
	
	if _, err := os.Stat(filepath.Join(serviceDir, "build.gradle")); err == nil {
		// Gradle project - ensure build directory is writable
		buildDir := filepath.Join(serviceDir, "build")
		if err := os.MkdirAll(buildDir, 0755); err == nil {
			buildDirs = append(buildDirs, filepath.Join(buildDir, "classes"))
		}
	}
	
	// Create build output directories
	for _, buildDir := range buildDirs {
		if err := os.MkdirAll(buildDir, 0755); err == nil {
			log.Printf("[DEBUG] Ensured build directory exists: %s", buildDir)
		}
	}
	
	return nil
}

// Helper functions

func getJavaExecutable() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	
	if runtime.GOOS == "windows" {
		return !info.IsDir()
	}
	
	// Unix-like systems: check execute permission
	return !info.IsDir() && (info.Mode().Perm()&0111) != 0
}

func isWorkingJava(javaPath string) bool {
	// Test if Java actually works by running -version
	cmd := exec.Command(javaPath, "-version")
	return cmd.Run() == nil
}

func inferJavaHome(javaPath string) string {
	// Check if this is an asdf shim and resolve it to the real Java path
	if strings.Contains(javaPath, "/.asdf/shims/") {
		log.Printf("[DEBUG] Detected asdf shim, attempting to resolve actual Java path: %s", javaPath)

		// Try to use 'asdf which java' to get the real path
		cmd := exec.Command("asdf", "which", "java")
		if output, err := cmd.Output(); err == nil && len(output) > 0 {
			realJavaPath := strings.TrimSpace(string(output))
			log.Printf("[DEBUG] Resolved asdf shim to: %s", realJavaPath)
			// Use the real Java path for inference
			binDir := filepath.Dir(realJavaPath)
			if filepath.Base(binDir) == "bin" {
				resolvedHome := filepath.Dir(binDir)
				log.Printf("[DEBUG] Inferred JAVA_HOME from resolved path: %s", resolvedHome)
				return resolvedHome
			}
		} else {
			log.Printf("[WARN] Failed to resolve asdf shim: %v", err)
		}
	}

	// Check if this is an SDKMAN installation and use the current symlink
	if strings.Contains(javaPath, "/.sdkman/candidates/java/") && strings.Contains(javaPath, "/current/") {
		log.Printf("[DEBUG] Detected SDKMAN Java installation: %s", javaPath)
		// For SDKMAN, we can use the path as-is since 'current' is already resolved
		binDir := filepath.Dir(javaPath)
		if filepath.Base(binDir) == "bin" {
			resolvedHome := filepath.Dir(binDir)
			log.Printf("[DEBUG] Using SDKMAN Java home: %s", resolvedHome)
			return resolvedHome
		}
	}

	// Standard inference: Remove /bin/java to get JAVA_HOME
	binDir := filepath.Dir(javaPath)
	if filepath.Base(binDir) == "bin" {
		return filepath.Dir(binDir)
	}
	return binDir
}

func getDefaultJavaPaths() []string {
	paths := []string{}
	
	// Add user-specific paths first (higher priority for current user setup)
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		homeDir := os.Getenv("HOME")
		if homeDir != "" {
			userPaths := []string{
				homeDir + "/.asdf/installs/java/openjdk-21",
				homeDir + "/.asdf/installs/java/openjdk-17",
				homeDir + "/.asdf/installs/java/openjdk-11",
				homeDir + "/.sdkman/candidates/java/current",
			}
			paths = append(paths, userPaths...)
		}
	}
	
	// Add system-wide paths as fallback
	switch runtime.GOOS {
	case "darwin": // macOS
		systemPaths := []string{
			"/opt/homebrew/opt/openjdk/libexec/openjdk.jdk/Contents/Home",
			"/usr/local/opt/openjdk/libexec/openjdk.jdk/Contents/Home",
			"/Library/Java/JavaVirtualMachines/openjdk.jdk/Contents/Home",
		}
		paths = append(paths, systemPaths...)
	case "linux":
		systemPaths := []string{
			"/usr/lib/jvm/default-java",
			"/usr/lib/jvm/java-17-openjdk",
			"/usr/lib/jvm/java-11-openjdk",
			"/opt/java/openjdk",
		}
		paths = append(paths, systemPaths...)
	case "windows":
		systemPaths := []string{
			`C:\Program Files\Java\jdk-17`,
			`C:\Program Files\Java\jdk-11`,
			`C:\Program Files (x86)\Java\jdk-17`,
		}
		paths = append(paths, systemPaths...)
	}
	
	return paths
}

func getJavaVersion(javaPath string) string {
	cmd := exec.Command(javaPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "unknown"
	}
	
	// Parse version from output like: openjdk version "17.0.1" 2021-10-19
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		line := lines[0]
		if strings.Contains(line, "version") {
			// Extract version string between quotes
			start := strings.Index(line, `"`)
			if start != -1 {
				end := strings.Index(line[start+1:], `"`)
				if end != -1 {
					return line[start+1 : start+1+end]
				}
			}
		}
	}
	
	return "unknown"
}