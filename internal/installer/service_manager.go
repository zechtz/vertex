// Package installer
package installer

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// ServiceManager handles cross-platform service management
type ServiceManager struct {
	serviceName string
	homeDir     string
}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	homeDir, _ := os.UserHomeDir()
	return &ServiceManager{
		serviceName: "vertex",
		homeDir:     homeDir,
	}
}

// Start starts the Vertex service
func (sm *ServiceManager) Start() error {
	switch runtime.GOOS {
	case "darwin":
		return sm.startMacOSService()
	case "linux":
		return sm.startLinuxService()
	case "windows":
		return sm.startWindowsService()
	default:
		return fmt.Errorf("service management not supported on %s", runtime.GOOS)
	}
}

// Stop stops the Vertex service
func (sm *ServiceManager) Stop() error {
	switch runtime.GOOS {
	case "darwin":
		return sm.stopMacOSService()
	case "linux":
		return sm.stopLinuxService()
	case "windows":
		return sm.stopWindowsService()
	default:
		return fmt.Errorf("service management not supported on %s", runtime.GOOS)
	}
}

// Restart restarts the Vertex service
func (sm *ServiceManager) Restart() error {
	fmt.Printf("🔄 Stopping Vertex service...\n")
	if err := sm.Stop(); err != nil {
		fmt.Printf("⚠️  Warning: Failed to stop service: %v\n", err)
	}
	
	// Wait a moment for the service to fully stop
	time.Sleep(3 * time.Second)
	
	fmt.Printf("🚀 Starting Vertex service...\n")
	return sm.Start()
}

// ShowStatus displays service status
func (sm *ServiceManager) ShowStatus() error {
	switch runtime.GOOS {
	case "darwin":
		return sm.showMacOSStatus()
	case "linux":
		return sm.showLinuxStatus()
	case "windows":
		return sm.showWindowsStatus()
	default:
		return fmt.Errorf("status viewing not supported on %s", runtime.GOOS)
	}
}

// ShowLogs displays service logs
func (sm *ServiceManager) ShowLogs(follow bool) error {
	switch runtime.GOOS {
	case "darwin":
		return sm.showMacOSLogs(follow)
	case "linux":
		return sm.showLinuxLogs(follow)
	case "windows":
		return sm.showWindowsLogs(follow)
	default:
		return fmt.Errorf("log viewing not supported on %s", runtime.GOOS)
	}
}

// macOS service management
func (sm *ServiceManager) startMacOSService() error {
	plistFile := filepath.Join(sm.homeDir, "Library", "LaunchAgents", "com.vertex.manager.plist")
	
	// Check if service file exists
	if _, err := os.Stat(plistFile); os.IsNotExist(err) {
		return fmt.Errorf("service not installed. Run './vertex --install' first")
	}
	
	// Load the service
	cmd := exec.Command("launchctl", "load", plistFile)
	if err := cmd.Run(); err != nil {
		// Service might already be loaded, try to start it
		cmd = exec.Command("launchctl", "start", "com.vertex.manager")
		return cmd.Run()
	}
	
	// Start the service
	cmd = exec.Command("launchctl", "start", "com.vertex.manager")
	return cmd.Run()
}

func (sm *ServiceManager) stopMacOSService() error {
	plistFile := filepath.Join(sm.homeDir, "Library", "LaunchAgents", "com.vertex.manager.plist")
	
	// First try to stop the service
	cmd := exec.Command("launchctl", "stop", "com.vertex.manager")
	cmd.Run() // Ignore errors
	
	// Then unload it to prevent automatic restart
	cmd = exec.Command("launchctl", "unload", plistFile)
	return cmd.Run()
}

func (sm *ServiceManager) showMacOSLogs(follow bool) error {
	logFiles := []string{
		filepath.Join(sm.homeDir, ".vertex", "vertex.stderr.log"),
		filepath.Join(sm.homeDir, ".vertex", "vertex.stdout.log"),
	}
	
	if follow {
		// Follow both stderr and stdout logs
		cmd := exec.Command("tail", "-f", logFiles[0], logFiles[1])
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		// Handle Ctrl+C gracefully
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			cmd.Process.Kill()
		}()
		
		return cmd.Run()
	} else {
		// Show last 50 lines of each log file
		for _, logFile := range logFiles {
			if _, err := os.Stat(logFile); err == nil {
				fmt.Printf("==> %s <==\n", filepath.Base(logFile))
				cmd := exec.Command("tail", "-n", "50", logFile)
				cmd.Stdout = os.Stdout
				cmd.Run()
				fmt.Println()
			}
		}
	}
	
	return nil
}

// Linux service management
func (sm *ServiceManager) startLinuxService() error {
	serviceFile := filepath.Join(sm.homeDir, ".config", "systemd", "user", "vertex.service")
	
	// Check if service file exists
	if _, err := os.Stat(serviceFile); os.IsNotExist(err) {
		return fmt.Errorf("service not installed. Run './vertex --install' first")
	}
	
	// Reload systemd and start service
	cmd := exec.Command("systemctl", "--user", "daemon-reload")
	cmd.Run() // Ignore errors
	
	cmd = exec.Command("systemctl", "--user", "start", "vertex")
	return cmd.Run()
}

func (sm *ServiceManager) stopLinuxService() error {
	cmd := exec.Command("systemctl", "--user", "stop", "vertex")
	return cmd.Run()
}

func (sm *ServiceManager) showLinuxLogs(follow bool) error {
	args := []string{"--user", "-u", "vertex"}
	
	if follow {
		args = append(args, "-f")
	} else {
		args = append(args, "--lines=50")
	}
	
	cmd := exec.Command("journalctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if follow {
		// Handle Ctrl+C gracefully
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			cmd.Process.Kill()
		}()
	}
	
	return cmd.Run()
}

// Windows service management
func (sm *ServiceManager) startWindowsService() error {
	taskName := "VertexServiceManager"
	
	// Check if task exists
	cmd := exec.Command("schtasks", "/query", "/tn", taskName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("service not installed. Run './vertex --install' first")
	}
	
	// Start the scheduled task
	cmd = exec.Command("schtasks", "/run", "/tn", taskName)
	return cmd.Run()
}

func (sm *ServiceManager) stopWindowsService() error {
	taskName := "VertexServiceManager"
	
	// End the scheduled task
	cmd := exec.Command("schtasks", "/end", "/tn", taskName)
	return cmd.Run()
}

func (sm *ServiceManager) showWindowsLogs(follow bool) error {
	logDir := filepath.Join(sm.homeDir, ".vertex")
	logFiles := []string{
		filepath.Join(logDir, "vertex.log"),
		filepath.Join(logDir, "vertex.stdout.log"),
		filepath.Join(logDir, "vertex.stderr.log"),
	}
	
	if follow {
		// Windows doesn't have tail -f, so we'll simulate it
		fmt.Printf("Following logs... Press Ctrl+C to exit\n\n")
		
		// Create a simple log follower
		for {
			for _, logFile := range logFiles {
				if _, err := os.Stat(logFile); err == nil {
					content, err := os.ReadFile(logFile)
					if err == nil && len(content) > 0 {
						fmt.Printf("==> %s <==\n", filepath.Base(logFile))
						// Show last few lines
						lines := strings.Split(string(content), "\n")
						start := len(lines) - 10
						if start < 0 {
							start = 0
						}
						for i := start; i < len(lines); i++ {
							if strings.TrimSpace(lines[i]) != "" {
								fmt.Println(lines[i])
							}
						}
						fmt.Println()
					}
				}
			}
			time.Sleep(2 * time.Second)
		}
	} else {
		// Show contents of log files
		for _, logFile := range logFiles {
			if _, err := os.Stat(logFile); err == nil {
				fmt.Printf("==> %s <==\n", filepath.Base(logFile))
				content, err := os.ReadFile(logFile)
				if err == nil {
					fmt.Println(string(content))
				}
				fmt.Println()
			}
		}
	}
	
	return nil
}
// Status checking functions
func (sm *ServiceManager) showMacOSStatus() error {
	// Check if plist file exists
	plistFile := filepath.Join(sm.homeDir, "Library", "LaunchAgents", "com.vertex.manager.plist")
	if _, err := os.Stat(plistFile); os.IsNotExist(err) {
		fmt.Printf("❌ Service not installed\n")
		return nil
	}
	
	// Check if service is loaded
	cmd := exec.Command("launchctl", "list", "com.vertex.manager")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Service not loaded\n")
		return nil
	}
	
	// Check if service is actually running (test connection)
	if sm.testConnection() {
		fmt.Printf("✅ Service is running\n")
		fmt.Printf("🌐 Available at: http://localhost:54321\n")
		
		// Check nginx proxy status
		httpProxy := sm.checkNginxProxy("http://vertex.dev")
		httpsProxy := sm.checkNginxProxy("https://vertex.dev")
		
		if httpsProxy {
			fmt.Printf("🔒 Available via HTTPS: https://vertex.dev\n")
		}
		if httpProxy {
			fmt.Printf("🌐 Available via HTTP: http://vertex.dev\n")
		}
		
		// Check for other potential domains
		if sm.checkNginxProxy("http://vertex.local") {
			fmt.Printf("🌐 Available via HTTP: http://vertex.local\n")
		}
		if sm.checkNginxProxy("https://vertex.local") {
			fmt.Printf("🔒 Available via HTTPS: https://vertex.local\n")
		}
	} else {
		fmt.Printf("⚠️  Service loaded but not responding\n")
	}
	
	// Show service details
	fmt.Printf("📋 Service details:\n")
	fmt.Printf(string(output))
	
	return nil
}

func (sm *ServiceManager) showLinuxStatus() error {
	// Check if service file exists
	serviceFile := filepath.Join(sm.homeDir, ".config", "systemd", "user", "vertex.service")
	if _, err := os.Stat(serviceFile); os.IsNotExist(err) {
		fmt.Printf("❌ Service not installed\n")
		return nil
	}
	
	// Get systemd status
	cmd := exec.Command("systemctl", "--user", "status", "vertex")
	output, err := cmd.CombinedOutput()
	
	// Check if service is actually running (test connection)
	if sm.testConnection() {
		fmt.Printf("✅ Service is running\n")
		fmt.Printf("🌐 Available at: http://localhost:54321\n")
		
		// Check nginx proxy status
		httpProxy := sm.checkNginxProxy("http://vertex.dev")
		httpsProxy := sm.checkNginxProxy("https://vertex.dev")
		
		if httpsProxy {
			fmt.Printf("🔒 Available via HTTPS: https://vertex.dev\n")
		}
		if httpProxy {
			fmt.Printf("🌐 Available via HTTP: http://vertex.dev\n")
		}
		
		// Check for other potential domains
		if sm.checkNginxProxy("http://vertex.local") {
			fmt.Printf("🌐 Available via HTTP: http://vertex.local\n")
		}
		if sm.checkNginxProxy("https://vertex.local") {
			fmt.Printf("🔒 Available via HTTPS: https://vertex.local\n")
		}
	} else {
		if err != nil {
			fmt.Printf("❌ Service not running\n")
		} else {
			fmt.Printf("⚠️  Service loaded but not responding\n")
		}
	}
	
	// Show systemd status
	fmt.Printf("📋 Service details:\n")
	fmt.Printf(string(output))
	
	return nil
}

func (sm *ServiceManager) showWindowsStatus() error {
	// Check if scheduled task exists
	cmd := exec.Command("schtasks", "/query", "/tn", "VertexServiceManager")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Service not installed\n")
		return nil
	}
	
	// Check if service is actually running (test connection)
	if sm.testConnection() {
		fmt.Printf("✅ Service is running\n")
		fmt.Printf("🌐 Available at: http://localhost:54321\n")
	} else {
		fmt.Printf("❌ Service not running\n")
	}
	
	// Show task details
	fmt.Printf("📋 Service details:\n")
	fmt.Printf(string(output))
	
	return nil
}

// Helper function to test if service is responding
func (sm *ServiceManager) testConnection() bool {
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "http://localhost:54321")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return string(output) == "200"
}

// Helper function to check nginx proxy status
func (sm *ServiceManager) checkNginxProxy(url string) bool {
	// Check if nginx is running and configured for vertex
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return string(output) == "200"
}
