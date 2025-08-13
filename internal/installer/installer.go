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
	BinaryPath   string
	Port         string
	DataDir      string
	User         string
	Domain       string
	EnableNginx  bool
	HTTPSEnabled bool
}

// NewServiceInstaller creates a new service installer
func NewServiceInstaller() *ServiceInstaller {
	execPath, err := os.Executable()
	if err != nil {
		execPath = "vertex"
	}
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	dataDir := os.Getenv("VERTEX_DATA_DIR")
	if dataDir == "" {
		homeDir, _ := os.UserHomeDir()
		dataDir = filepath.Join(homeDir, ".vertex")
	}
	return &ServiceInstaller{
		BinaryPath:   execPath,
		Port:         "54321",
		DataDir:      dataDir,
		User:         user,
		Domain:       "vertex.dev",
		EnableNginx:  false,
		HTTPSEnabled: false,
	}
}

// Install performs cross-platform service installation
func (si *ServiceInstaller) Install() error {
	fmt.Printf("📦 Installing Vertex as a user service...\n")
	if err := si.createDataDirectory(); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}
	if err := si.installBinary(); err != nil {
		return fmt.Errorf("failed to install binary: %v", err)
	}
	var serviceErr error
	switch runtime.GOOS {
	case "darwin":
		serviceErr = si.installMacOSService()
	case "linux":
		serviceErr = si.installLinuxService()
	case "windows":
		serviceErr = si.installWindowsService()
	default:
		serviceErr = fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	if serviceErr != nil {
		return serviceErr
	}
	if si.EnableNginx {
		if err := si.installNginxConfig(); err != nil {
			fmt.Printf("⚠️  Nginx configuration failed: %v\n", err)
			fmt.Printf("Service is still accessible at http://localhost:%s\n", si.Port)
		}
	}
	return nil
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
	currentExe, err := filepath.Abs(si.BinaryPath)
	if err != nil {
		return err
	}
	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}
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
	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return err
	}
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
	// Use current process's PATH
	envPath := os.Getenv("PATH")
	if envPath == "" {
		envPath = "/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
	}
	// Include MAVEN_HOME if set
	mavenHome := os.Getenv("MAVEN_HOME")
	envVars := map[string]string{
		"VERTEX_DATA_DIR": si.DataDir,
		"PATH":            envPath,
	}
	if mavenHome != "" {
		envVars["MAVEN_HOME"] = mavenHome
	}
	envVarsXML := ""
	for key, value := range envVars {
		envVarsXML += fmt.Sprintf("        <key>%s</key>\n        <string>%s</string>\n", key, value)
	}
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.vertex.manager</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>--port</string>
        <string>%s</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
%s
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
</plist>`, binaryPath, si.Port, envVarsXML, si.DataDir, si.DataDir)
	if err := os.WriteFile(plistFile, []byte(plistContent), 0644); err != nil {
		return err
	}
	fmt.Printf("📝 Created LaunchAgent: %s\n", plistFile)
	cmd := exec.Command("launchctl", "load", plistFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to load LaunchAgent: %v", err)
	}
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
	envPath := os.Getenv("PATH")
	if envPath == "" {
		envPath = "/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
	}
	mavenHome := os.Getenv("MAVEN_HOME")
	envVars := []string{
		fmt.Sprintf("Environment=VERTEX_DATA_DIR=%s", si.DataDir),
		fmt.Sprintf("Environment=PATH=%s", envPath),
	}
	if mavenHome != "" {
		envVars = append(envVars, fmt.Sprintf("Environment=MAVEN_HOME=%s", mavenHome))
	}
	envVarsStr := ""
	for _, env := range envVars {
		envVarsStr += env + "\n"
	}
	serviceContent := fmt.Sprintf(`[Unit]
Description=Vertex Service Manager
After=network.target

[Service]
Type=simple
ExecStart=%s --port %s
%sRestart=always
RestartSec=5

[Install]
WantedBy=default.target`, binaryPath, si.Port, envVarsStr)
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return err
	}
	fmt.Printf("📝 Created systemd service: %s\n", serviceFile)
	cmd := exec.Command("systemctl", "--user", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %v", err)
	}
	cmd = exec.Command("systemctl", "--user", "enable", "vertex")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable service: %v", err)
	}
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
	localBinDir := filepath.Join(homeDir, ".local", "bin")
	if err := os.MkdirAll(localBinDir, 0755); err != nil {
		return err
	}
	binaryPath := filepath.Join(localBinDir, "vertex.exe")
	envPath := os.Getenv("PATH")
	if envPath == "" {
		envPath = `C:\Windows\System32;C:\Windows;C:\Windows\System32\wbem`
	}
	mavenHome := os.Getenv("MAVEN_HOME")
	batchFile := filepath.Join(localBinDir, "vertex-service.bat")
	batchContent := fmt.Sprintf(`@echo off
set VERTEX_DATA_DIR=%s
set PATH=%s
`, si.DataDir, envPath)
	if mavenHome != "" {
		batchContent += fmt.Sprintf("set MAVEN_HOME=%s\n", mavenHome)
	}
	batchContent += fmt.Sprintf(`"%s" --port %s`, binaryPath, si.Port)
	if err := os.WriteFile(batchFile, []byte(batchContent), 0644); err != nil {
		return err
	}
	fmt.Printf("📝 Created service batch file: %s\n", batchFile)
	taskName := "VertexServiceManager"
	exec.Command("schtasks", "/delete", "/tn", taskName, "/f").Run()
	cmd := exec.Command("schtasks", "/create", "/tn", taskName, "/tr", batchFile, "/sc", "onlogon", "/rl", "limited")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create scheduled task: %v", err)
	}
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
	exec.Command("launchctl", "stop", "com.vertex.manager").Run()
	exec.Command("launchctl", "unload", plistFile).Run()
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
	exec.Command("systemctl", "--user", "stop", "vertex").Run()
	exec.Command("systemctl", "--user", "disable", "vertex").Run()
	exec.Command("systemctl", "--user", "daemon-reload").Run()
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
	exec.Command("schtasks", "/delete", "/tn", "VertexServiceManager", "/f").Run()
	localBinDir := filepath.Join(homeDir, ".local", "bin")
	os.Remove(filepath.Join(localBinDir, "vertex.exe"))
	os.Remove(filepath.Join(localBinDir, "vertex-service.bat"))
	os.RemoveAll(si.DataDir)
	return nil
}

// installNginxConfig installs nginx configuration for domain access
func (si *ServiceInstaller) installNginxConfig() error {
	nginxInstaller := NewNginxInstaller(si.Domain, si.Port)
	nginxInstaller.EnableHTTPS(si.HTTPSEnabled)
	return nginxInstaller.InstallNginxConfig()
}

// SetDomain sets the domain for nginx configuration
func (si *ServiceInstaller) SetDomain(domain string) {
	si.Domain = domain
}

// EnableNginxProxy enables nginx proxy configuration
func (si *ServiceInstaller) EnableNginxProxy(enable bool) {
	si.EnableNginx = enable
}

// EnableHTTPS enables HTTPS configuration
func (si *ServiceInstaller) EnableHTTPS(enable bool) {
	si.HTTPSEnabled = enable
}
