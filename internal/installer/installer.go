// Package installer
package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ServiceInstaller handles cross-platform service installation
type ServiceInstaller struct {
	BinaryPath string
	Port       string
	DataDir    string
	User       string
}

// NewServiceInstaller creates a new service installer
func NewServiceInstaller() *ServiceInstaller {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		execPath = "vertex" // fallback
	}

	// Get current user
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME") // Windows
	}

	// Set default data directory
	dataDir := os.Getenv("VERTEX_DATA_DIR")
	if dataDir == "" {
		homeDir, _ := os.UserHomeDir()
		dataDir = filepath.Join(homeDir, ".vertex")
	}

	return &ServiceInstaller{
		BinaryPath: execPath,
		Port:       "8080",
		DataDir:    dataDir,
		User:       user,
	}
}

// Install performs cross-platform service installation
func (si *ServiceInstaller) Install() error {
	fmt.Printf("📦 Installing Vertex as a user service...\n")

	// Create data directory
	if err := si.createDataDirectory(); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// Install binary to user's local bin
	if err := si.installBinary(); err != nil {
		return fmt.Errorf("failed to install binary: %v", err)
	}

	// Create and install service based on platform
	switch runtime.GOOS {
	case "darwin":
		return si.installMacOSService()
	case "linux":
		return si.installLinuxService()
	case "windows":
		return si.installWindowsService()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// createDataDirectory creates the data directory
func (si *ServiceInstaller) createDataDirectory() error {
	fmt.Printf("📁 Creating data directory: %s\n", si.DataDir)
	return os.MkdirAll(si.DataDir, 0755)
}

// installBinary copies the binary to the user's local bin directory
func (si *ServiceInstaller) installBinary() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	localBinDir := filepath.Join(homeDir, ".local", "bin")
	if err := os.MkdirAll(localBinDir, 0755); err != nil {
		return err
	}

	targetPath := filepath.Join(localBinDir, "vertex")
	if runtime.GOOS == "windows" {
		targetPath += ".exe"
	}

	// Get absolute path of current executable
	currentExe, err := filepath.Abs(si.BinaryPath)
	if err != nil {
		return err
	}

	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}

	// Only copy if source and target are different
	if currentExe != targetAbs {
		fmt.Printf("📋 Installing binary from %s to: %s\n", filepath.Base(si.BinaryPath), targetPath)
		return si.copyFile(si.BinaryPath, targetPath)
	}

	fmt.Printf("✅ Binary already in correct location: %s\n", targetPath)
	return nil
}

// copyFile copies a file and preserves permissions
func (si *ServiceInstaller) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy content
	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return err
	}

	// Copy permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

// installMacOSService creates and loads a LaunchAgent
func (si *ServiceInstaller) installMacOSService() error {
	fmt.Printf("🍎 Installing macOS LaunchAgent...\n")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	launchAgentsDir := filepath.Join(homeDir, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return err
	}

	plistFile := filepath.Join(launchAgentsDir, "com.vertex.manager.plist")
	binaryPath := filepath.Join(homeDir, ".local", "bin", "vertex")

	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.vertex.manager</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>-port</string>
        <string>%s</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>VERTEX_DATA_DIR</key>
        <string>%s</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>%s/vertex.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>%s/vertex.stderr.log</string>
</dict>
</plist>`, binaryPath, si.Port, si.DataDir, si.DataDir, si.DataDir)

	// Write plist file
	if err := os.WriteFile(plistFile, []byte(plistContent), 0644); err != nil {
		return err
	}

	fmt.Printf("📝 Created LaunchAgent: %s\n", plistFile)

	// Load the service
	cmd := exec.Command("launchctl", "load", plistFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to load LaunchAgent: %v", err)
	}

	// Start the service
	cmd = exec.Command("launchctl", "start", "com.vertex.manager")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start service: %v", err)
	}

	fmt.Printf("🚀 Started LaunchAgent service\n")
	return nil
}

// installLinuxService creates and enables a systemd user service
func (si *ServiceInstaller) installLinuxService() error {
	fmt.Printf("🐧 Installing Linux systemd user service...\n")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	systemdDir := filepath.Join(homeDir, ".config", "systemd", "user")
	if err := os.MkdirAll(systemdDir, 0755); err != nil {
		return err
	}

	serviceFile := filepath.Join(systemdDir, "vertex.service")
	binaryPath := filepath.Join(homeDir, ".local", "bin", "vertex")

	serviceContent := fmt.Sprintf(`[Unit]
Description=Vertex Service Manager
After=network.target

[Service]
Type=simple
ExecStart=%s --port %s
Environment=VERTEX_DATA_DIR=%s
Restart=always
RestartSec=5

[Install]
WantedBy=default.target`, binaryPath, si.Port, si.DataDir)

	// Write service file
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return err
	}

	fmt.Printf("📝 Created systemd service: %s\n", serviceFile)

	// Reload systemd
	cmd := exec.Command("systemctl", "--user", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %v", err)
	}

	// Enable the service
	cmd = exec.Command("systemctl", "--user", "enable", "vertex")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable service: %v", err)
	}

	// Start the service
	cmd = exec.Command("systemctl", "--user", "start", "vertex")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start service: %v", err)
	}

	fmt.Printf("🚀 Started systemd user service\n")
	return nil
}

// installWindowsService installs as a Windows service
func (si *ServiceInstaller) installWindowsService() error {
	fmt.Printf("🪟 Installing Windows service...\n")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Install binary to user's local bin
	localBinDir := filepath.Join(homeDir, ".local", "bin")
	if err := os.MkdirAll(localBinDir, 0755); err != nil {
		return err
	}

	binaryPath := filepath.Join(localBinDir, "vertex.exe")

	// Create a simple batch file to start the service
	batchFile := filepath.Join(localBinDir, "vertex-service.bat")
	batchContent := fmt.Sprintf(`@echo off
set VERTEX_DATA_DIR=%s
"%s" --port %s`, si.DataDir, binaryPath, si.Port)

	if err := os.WriteFile(batchFile, []byte(batchContent), 0644); err != nil {
		return err
	}

	fmt.Printf("📝 Created service batch file: %s\n", batchFile)

	// Create a scheduled task to run at startup
	taskName := "VertexServiceManager"

	// Remove existing task if it exists
	exec.Command("schtasks", "/delete", "/tn", taskName, "/f").Run()

	// Create new task
	cmd := exec.Command("schtasks", "/create", "/tn", taskName, "/tr", batchFile, "/sc", "onlogon", "/rl", "limited")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create scheduled task: %v", err)
	}

	// Run the task now
	cmd = exec.Command("schtasks", "/run", "/tn", taskName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start scheduled task: %v", err)
	}

	fmt.Printf("🚀 Started Windows scheduled task\n")
	return nil
}

// Uninstall removes the service
func (si *ServiceInstaller) Uninstall() error {
	fmt.Printf("🗑️ Uninstalling Vertex service...\n")

	switch runtime.GOOS {
	case "darwin":
		return si.uninstallMacOSService()
	case "linux":
		return si.uninstallLinuxService()
	case "windows":
		return si.uninstallWindowsService()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (si *ServiceInstaller) uninstallMacOSService() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	plistFile := filepath.Join(homeDir, "Library", "LaunchAgents", "com.vertex.manager.plist")

	// Stop and unload service
	exec.Command("launchctl", "stop", "com.vertex.manager").Run()
	exec.Command("launchctl", "unload", plistFile).Run()

	// Remove files
	os.Remove(plistFile)
	os.Remove(filepath.Join(homeDir, ".local", "bin", "vertex"))
	os.RemoveAll(si.DataDir)

	return nil
}

func (si *ServiceInstaller) uninstallLinuxService() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	serviceFile := filepath.Join(homeDir, ".config", "systemd", "user", "vertex.service")

	// Stop and disable service
	exec.Command("systemctl", "--user", "stop", "vertex").Run()
	exec.Command("systemctl", "--user", "disable", "vertex").Run()
	exec.Command("systemctl", "--user", "daemon-reload").Run()

	// Remove files
	os.Remove(serviceFile)
	os.Remove(filepath.Join(homeDir, ".local", "bin", "vertex"))
	os.RemoveAll(si.DataDir)

	return nil
}

func (si *ServiceInstaller) uninstallWindowsService() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Remove scheduled task
	exec.Command("schtasks", "/delete", "/tn", "VertexServiceManager", "/f").Run()

	// Remove files
	localBinDir := filepath.Join(homeDir, ".local", "bin")
	os.Remove(filepath.Join(localBinDir, "vertex.exe"))
	os.Remove(filepath.Join(localBinDir, "vertex-service.bat"))
	os.RemoveAll(si.DataDir)

	return nil
}
