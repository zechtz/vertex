// Package installer
package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// NginxInstaller handles nginx configuration for Vertex
type NginxInstaller struct {
	Domain     string
	Port       string
	ConfigPath string
	SitesPath  string
	HTTPSEnabled bool
}

// NewNginxInstaller creates a new nginx installer
func NewNginxInstaller(domain, port string) *NginxInstaller {
	ni := &NginxInstaller{
		Domain:       domain,
		Port:         port,
		HTTPSEnabled: false,
	}

	// Set platform-specific paths
	switch runtime.GOOS {
	case "darwin":
		// Detect homebrew installation path by checking where nginx is installed
		if nginxPath, err := exec.LookPath("nginx"); err == nil {
			if strings.Contains(nginxPath, "/opt/homebrew/") {
				// Apple Silicon homebrew path
				ni.ConfigPath = "/opt/homebrew/etc/nginx/nginx.conf"
				ni.SitesPath = "/opt/homebrew/etc/nginx/servers"
			} else {
				// Intel homebrew path
				ni.ConfigPath = "/usr/local/etc/nginx/nginx.conf"
				ni.SitesPath = "/usr/local/etc/nginx/servers"
			}
		} else {
			// Default to Apple Silicon if nginx not found yet (will be installed)
			ni.ConfigPath = "/opt/homebrew/etc/nginx/nginx.conf"
			ni.SitesPath = "/opt/homebrew/etc/nginx/servers"
		}
	case "linux":
		// Standard Linux nginx paths
		ni.ConfigPath = "/etc/nginx/nginx.conf"
		ni.SitesPath = "/etc/nginx/sites-available"
	default:
		// Default paths
		ni.ConfigPath = "/etc/nginx/nginx.conf"
		ni.SitesPath = "/etc/nginx/conf.d"
	}

	return ni
}

// IsNginxInstalled checks if nginx is installed
func (ni *NginxInstaller) IsNginxInstalled() bool {
	_, err := exec.LookPath("nginx")
	return err == nil
}

// installNginx installs nginx on the current platform
func (ni *NginxInstaller) installNginx() error {
	switch runtime.GOOS {
	case "darwin":
		return ni.installNginxMacOS()
	case "linux":
		return ni.installNginxLinux()
	case "windows":
		return ni.installNginxWindows()
	default:
		return fmt.Errorf("automatic nginx installation not supported on %s", runtime.GOOS)
	}
}

// installNginxMacOS installs nginx on macOS using homebrew
func (ni *NginxInstaller) installNginxMacOS() error {
	// Check if homebrew is installed
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("homebrew is required to install nginx on macOS. Please install homebrew first:\n" +
			"/bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"")
	}

	fmt.Printf("🍺 Installing nginx via homebrew...\n")
	cmd := exec.Command("brew", "install", "nginx")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install nginx: %s", string(output))
	}

	fmt.Printf("✅ Nginx installed successfully\n")
	return nil
}

// installNginxLinux installs nginx on Linux using the system package manager
func (ni *NginxInstaller) installNginxLinux() error {
	// Try different package managers
	packageManagers := []struct {
		command string
		args    []string
		name    string
	}{
		{"apt", []string{"update", "&&", "apt", "install", "-y", "nginx"}, "apt"},
		{"yum", []string{"install", "-y", "nginx"}, "yum"},
		{"dnf", []string{"install", "-y", "nginx"}, "dnf"},
		{"pacman", []string{"-S", "--noconfirm", "nginx"}, "pacman"},
		{"zypper", []string{"install", "-y", "nginx"}, "zypper"},
	}

	for _, pm := range packageManagers {
		if _, err := exec.LookPath(pm.command); err == nil {
			fmt.Printf("🐧 Installing nginx using %s...\n", pm.name)
			
			var cmd *exec.Cmd
			if pm.name == "apt" {
				// Handle apt update && apt install specially
				cmd = exec.Command("sh", "-c", "sudo apt update && sudo apt install -y nginx")
			} else {
				args := append([]string{pm.command}, pm.args...)
				cmd = exec.Command("sudo", args...)
			}
			
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("⚠️  Failed with %s: %s\n", pm.name, string(output))
				continue
			}
			
			fmt.Printf("✅ Nginx installed successfully using %s\n", pm.name)
			return nil
		}
	}

	return fmt.Errorf("no supported package manager found. Please install nginx manually:\n" +
		"  Ubuntu/Debian: sudo apt install nginx\n" +
		"  CentOS/RHEL: sudo yum install nginx\n" +
		"  Fedora: sudo dnf install nginx\n" +
		"  Arch: sudo pacman -S nginx\n" +
		"  openSUSE: sudo zypper install nginx")
}

// installNginxWindows installs nginx on Windows
func (ni *NginxInstaller) installNginxWindows() error {
	// Check if chocolatey is available
	if _, err := exec.LookPath("choco"); err == nil {
		fmt.Printf("🍫 Installing nginx via chocolatey...\n")
		cmd := exec.Command("choco", "install", "nginx", "-y")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install nginx via chocolatey: %s", string(output))
		}
		fmt.Printf("✅ Nginx installed successfully via chocolatey\n")
		return nil
	}

	// Check if winget is available
	if _, err := exec.LookPath("winget"); err == nil {
		fmt.Printf("📦 Installing nginx via winget...\n")
		cmd := exec.Command("winget", "install", "nginx")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install nginx via winget: %s", string(output))
		}
		fmt.Printf("✅ Nginx installed successfully via winget\n")
		return nil
	}

	// If no package manager is available, provide manual instructions
	return fmt.Errorf("no package manager found. Please install nginx manually:\n" +
		"  1. Download nginx from: https://nginx.org/en/download.html\n" +
		"  2. Extract to C:\\nginx\n" +
		"  3. Add C:\\nginx to your PATH\n" +
		"  Or install a package manager:\n" +
		"  - Chocolatey: https://chocolatey.org/install\n" +
		"  - Winget: included with Windows 10/11")
}

// createSitesDirectory creates the nginx sites directory with proper permissions
func (ni *NginxInstaller) createSitesDirectory() error {
	// Check if directory already exists
	if _, err := os.Stat(ni.SitesPath); err == nil {
		return nil
	}

	// Try to create directory without sudo first
	if err := os.MkdirAll(ni.SitesPath, 0755); err == nil {
		return nil
	}

	// If that fails, try with sudo
	cmd := exec.Command("sudo", "mkdir", "-p", ni.SitesPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", ni.SitesPath, err)
	}

	// Set proper permissions
	cmd = exec.Command("sudo", "chmod", "755", ni.SitesPath)
	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️  Could not set permissions on %s: %v\n", ni.SitesPath, err)
	}

	return nil
}

// createNginxDirectories creates nginx log and run directories with proper permissions
func (ni *NginxInstaller) createNginxDirectories() error {
	var directories []string
	var logFiles []string
	var pidFiles []string
	
	switch runtime.GOOS {
	case "darwin":
		// Detect homebrew path and create corresponding directories
		if strings.Contains(ni.ConfigPath, "/opt/homebrew/") {
			directories = []string{
				"/opt/homebrew/var/log/nginx",
				"/opt/homebrew/var/run",
			}
			logFiles = []string{
				"/opt/homebrew/var/log/nginx/access.log",
				"/opt/homebrew/var/log/nginx/error.log",
			}
			pidFiles = []string{
				"/opt/homebrew/var/run/nginx.pid",
			}
		} else {
			directories = []string{
				"/usr/local/var/log/nginx",
				"/usr/local/var/run",
			}
			logFiles = []string{
				"/usr/local/var/log/nginx/access.log",
				"/usr/local/var/log/nginx/error.log",
			}
			pidFiles = []string{
				"/usr/local/var/run/nginx.pid",
			}
		}
	case "linux":
		directories = []string{
			"/var/log/nginx",
			"/var/run/nginx",
		}
		logFiles = []string{
			"/var/log/nginx/access.log",
			"/var/log/nginx/error.log",
		}
		pidFiles = []string{
			"/var/run/nginx/nginx.pid",
		}
	default:
		// Skip directory creation for unsupported platforms
		return nil
	}

	fmt.Printf("📁 Creating nginx directories...\n")
	
	for _, dir := range directories {
		// Try to create directory without sudo first
		if err := os.MkdirAll(dir, 0755); err != nil {
			// If that fails, use sudo
			cmd := exec.Command("sudo", "mkdir", "-p", dir)
			if err := cmd.Run(); err != nil {
				fmt.Printf("⚠️  Failed to create %s: %v\n", dir, err)
				continue
			}
		}

		// Set proper ownership (current user for homebrew, nginx user for system)
		if runtime.GOOS == "darwin" {
			// On macOS with homebrew, use current user
			currentUser := os.Getenv("USER")
			if currentUser != "" {
				cmd := exec.Command("sudo", "chown", "-R", currentUser, dir)
				if err := cmd.Run(); err != nil {
					fmt.Printf("⚠️  Could not set ownership on %s: %v\n", dir, err)
				}
			}
		} else {
			// On Linux, try to use nginx user if it exists
			cmd := exec.Command("sudo", "chown", "-R", "nginx:nginx", dir)
			if err := cmd.Run(); err != nil {
				// Fallback to www-data if nginx user doesn't exist
				cmd = exec.Command("sudo", "chown", "-R", "www-data:www-data", dir)
				cmd.Run() // Ignore error as this is best-effort
			}
		}

		// Set proper permissions
		cmd := exec.Command("sudo", "chmod", "755", dir)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Could not set permissions on %s: %v\n", dir, err)
		}
	}

	// Create and fix log files with proper ownership
	for _, logFile := range logFiles {
		// Create log file if it doesn't exist
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			// Create empty log file
			cmd := exec.Command("sudo", "touch", logFile)
			if err := cmd.Run(); err != nil {
				fmt.Printf("⚠️  Could not create log file %s: %v\n", logFile, err)
				continue
			}
		}

		// Fix ownership of existing log files
		if runtime.GOOS == "darwin" {
			currentUser := os.Getenv("USER")
			if currentUser != "" {
				cmd := exec.Command("sudo", "chown", currentUser+":admin", logFile)
				if err := cmd.Run(); err != nil {
					fmt.Printf("⚠️  Could not fix ownership of %s: %v\n", logFile, err)
				} else {
					fmt.Printf("✅ Fixed ownership of %s\n", logFile)
				}
			}
		}

		// Set proper permissions on log file
		cmd := exec.Command("sudo", "chmod", "644", logFile)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Could not set permissions on %s: %v\n", logFile, err)
		}
	}

	// Clean up and fix PID files and directories
	for _, pidFile := range pidFiles {
		// Ensure the parent directory has correct ownership first
		pidDir := filepath.Dir(pidFile)
		if runtime.GOOS == "darwin" {
			currentUser := os.Getenv("USER")
			if currentUser != "" {
				cmd := exec.Command("sudo", "chown", currentUser+":admin", pidDir)
				if err := cmd.Run(); err != nil {
					fmt.Printf("⚠️  Could not fix ownership of %s: %v\n", pidDir, err)
				} else {
					fmt.Printf("✅ Fixed ownership of %s\n", pidDir)
				}
			}
		}

		// Set proper permissions on PID directory
		cmd := exec.Command("sudo", "chmod", "755", pidDir)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Could not set permissions on %s: %v\n", pidDir, err)
		}

		// Remove existing PID file if it exists and is owned by root
		if stat, err := os.Stat(pidFile); err == nil {
			// Check if file is owned by root (UID 0)
			if sys := stat.Sys(); sys != nil {
				if stat, ok := sys.(*syscall.Stat_t); ok && stat.Uid == 0 {
					cmd := exec.Command("sudo", "rm", "-f", pidFile)
					if err := cmd.Run(); err == nil {
						fmt.Printf("✅ Removed root-owned PID file %s\n", pidFile)
					}
				}
			}
		}
	}

	fmt.Printf("✅ Nginx directories created\n")
	return nil
}

// InstallNginxConfig creates nginx configuration for Vertex
func (ni *NginxInstaller) InstallNginxConfig() error {
	if !ni.IsNginxInstalled() {
		fmt.Printf("📦 Nginx not found, installing automatically...\n")
		if err := ni.installNginx(); err != nil {
			return fmt.Errorf("failed to install nginx: %v", err)
		}
	}

	// Setup HTTPS certificates if enabled
	if ni.HTTPSEnabled {
		if err := ni.setupHTTPS(); err != nil {
			return fmt.Errorf("failed to setup HTTPS: %v", err)
		}
	}

	fmt.Printf("🌐 Configuring nginx for %s...\n", ni.Domain)

	// Create sites directory if it doesn't exist
	if err := ni.createSitesDirectory(); err != nil {
		return fmt.Errorf("failed to create nginx sites directory: %v", err)
	}

	// Create Vertex nginx configuration
	configFile := filepath.Join(ni.SitesPath, "vertex.conf")
	if err := ni.createNginxConfig(configFile); err != nil {
		return fmt.Errorf("failed to create nginx config: %v", err)
	}

	// Add domain to /etc/hosts if not already present
	if err := ni.addToHosts(); err != nil {
		fmt.Printf("⚠️  Could not update /etc/hosts automatically: %v\n", err)
		fmt.Printf("Please add this line to /etc/hosts manually:\n")
		fmt.Printf("127.0.0.1 %s\n", ni.Domain)
	}

	// Enable site (Linux specific)
	if runtime.GOOS == "linux" {
		if err := ni.enableSite(); err != nil {
			fmt.Printf("⚠️  Could not enable site automatically: %v\n", err)
			fmt.Printf("Please run: sudo ln -s %s /etc/nginx/sites-enabled/\n", configFile)
		}
	}

	// Create nginx log directories with proper permissions
	if err := ni.createNginxDirectories(); err != nil {
		fmt.Printf("⚠️  Could not create nginx directories: %v\n", err)
	}

	// Test nginx configuration
	if err := ni.testNginxConfig(); err != nil {
		return fmt.Errorf("nginx configuration test failed: %v", err)
	}

	// Start/restart nginx service
	if err := ni.startNginxService(); err != nil {
		fmt.Printf("⚠️  Could not start nginx automatically: %v\n", err)
		fmt.Printf("Please start nginx manually:\n")
		if runtime.GOOS == "darwin" {
			fmt.Printf("  brew services start nginx\n")
		} else {
			fmt.Printf("  sudo systemctl start nginx\n")
		}
	}

	fmt.Printf("✅ Nginx configured successfully!\n")
	protocol := "http"
	if ni.HTTPSEnabled {
		protocol = "https"
	}
	fmt.Printf("🌐 Vertex is now available at: %s://%s\n", protocol, ni.Domain)
	return nil
}

// createNginxConfig creates the nginx configuration file
func (ni *NginxInstaller) createNginxConfig(configFile string) error {
	var config string
	
	if ni.HTTPSEnabled {
		// HTTPS configuration with SSL certificates
		sslDir := filepath.Join(os.Getenv("HOME"), ".vertex", "ssl")
		certFile := filepath.Join(sslDir, ni.Domain+".pem")
		keyFile := filepath.Join(sslDir, ni.Domain+"-key.pem")
		
		config = fmt.Sprintf(`# Vertex Service Manager Configuration (HTTPS)
# HTTP to HTTPS redirect
server {
    listen 80;
    server_name %s;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name %s;

    # SSL Configuration
    ssl_certificate %s;
    ssl_certificate_key %s;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;

    # Modern configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=63072000" always;

    # Proxy settings
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # Main application
    location / {
        proxy_pass http://127.0.0.1:%s;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 86400;
    }

    # WebSocket support for real-time features
    location /ws {
        proxy_pass http://127.0.0.1:%s;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # API endpoints
    location /api/ {
        proxy_pass http://127.0.0.1:%s;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Static assets with caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        proxy_pass http://127.0.0.1:%s;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;
}`, ni.Domain, ni.Domain, certFile, keyFile, ni.Port, ni.Port, ni.Port, ni.Port)
	} else {
		// HTTP configuration
		config = fmt.Sprintf(`# Vertex Service Manager Configuration
server {
    listen 80;
    server_name %s;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";

    # Proxy settings
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # Main application
    location / {
        proxy_pass http://127.0.0.1:%s;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 86400;
    }

    # WebSocket support for real-time features
    location /ws {
        proxy_pass http://127.0.0.1:%s;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # API endpoints
    location /api/ {
        proxy_pass http://127.0.0.1:%s;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Static assets with caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        proxy_pass http://127.0.0.1:%s;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;
}`, ni.Domain, ni.Port, ni.Port, ni.Port, ni.Port)
	}

	// Try to write file normally first
	if err := os.WriteFile(configFile, []byte(config), 0644); err == nil {
		return nil
	}

	// If that fails, use sudo to write the file
	tempFile := "/tmp/vertex-nginx.conf"
	if err := os.WriteFile(tempFile, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to create temporary config: %v", err)
	}

	cmd := exec.Command("sudo", "mv", tempFile, configFile)
	if err := cmd.Run(); err != nil {
		os.Remove(tempFile) // cleanup
		return fmt.Errorf("failed to move config to %s: %v", configFile, err)
	}

	// Set proper permissions
	cmd = exec.Command("sudo", "chmod", "644", configFile)
	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️  Could not set permissions on %s: %v\n", configFile, err)
	}

	return nil
}

// addToHosts adds the domain to /etc/hosts
func (ni *NginxInstaller) addToHosts() error {
	hostsFile := "/etc/hosts"
	hostEntry := fmt.Sprintf("127.0.0.1 %s", ni.Domain)

	// Check if entry already exists
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return err
	}

	if strings.Contains(string(content), hostEntry) {
		fmt.Printf("✅ Domain %s already in /etc/hosts\n", ni.Domain)
		return nil
	}

	// Try to append to hosts file (requires sudo)
	cmd := exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> %s", hostEntry, hostsFile))
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Printf("✅ Added %s to /etc/hosts\n", ni.Domain)
	return nil
}

// enableSite creates symlink in sites-enabled (Linux)
func (ni *NginxInstaller) enableSite() error {
	if runtime.GOOS != "linux" {
		return nil
	}

	sourcePath := filepath.Join(ni.SitesPath, "vertex.conf")
	targetPath := "/etc/nginx/sites-enabled/vertex.conf"

	// Remove existing symlink if it exists
	os.Remove(targetPath)

	// Create symlink
	cmd := exec.Command("sudo", "ln", "-s", sourcePath, targetPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Printf("✅ Enabled nginx site\n")
	return nil
}

// testNginxConfig tests the nginx configuration
func (ni *NginxInstaller) testNginxConfig() error {
	cmd := exec.Command("sudo", "nginx", "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("configuration test failed: %s", string(output))
	}

	fmt.Printf("✅ Nginx configuration test passed\n")
	return nil
}

// startNginxService starts the nginx service
func (ni *NginxInstaller) startNginxService() error {
	switch runtime.GOOS {
	case "darwin":
		// Stop nginx first to ensure clean restart
		stopCmd := exec.Command("brew", "services", "stop", "nginx")
		stopCmd.Run() // Ignore errors if service wasn't running
		
		// Clean up any root-owned PID files before starting
		pidFile := "/opt/homebrew/var/run/nginx.pid"
		if strings.Contains(ni.ConfigPath, "/usr/local/") {
			pidFile = "/usr/local/var/run/nginx.pid"
		}
		
		if stat, err := os.Stat(pidFile); err == nil {
			if sys := stat.Sys(); sys != nil {
				if stat, ok := sys.(*syscall.Stat_t); ok && stat.Uid == 0 {
					cmd := exec.Command("sudo", "rm", "-f", pidFile)
					if err := cmd.Run(); err == nil {
						fmt.Printf("✅ Cleaned root-owned PID file before start\n")
					}
				}
			}
		}
		
		// Start nginx service
		cmd := exec.Command("brew", "services", "start", "nginx")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to start nginx: %s", string(output))
		}
		
		// Wait a moment for service to fully start
		time.Sleep(2 * time.Second)
		
		// Verify nginx is actually running by checking the service status
		statusCmd := exec.Command("brew", "services", "list")
		statusOutput, err := statusCmd.Output()
		if err == nil && strings.Contains(string(statusOutput), "nginx") && strings.Contains(string(statusOutput), "started") {
			fmt.Printf("✅ Nginx service started\n")
			return nil
		}
		
		return fmt.Errorf("nginx service appears to have failed to start properly")
		
	case "linux":
		// Use systemctl on Linux
		cmd := exec.Command("sudo", "systemctl", "enable", "nginx")
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Could not enable nginx service: %v\n", err)
		}
		cmd = exec.Command("sudo", "systemctl", "start", "nginx")
		if err := cmd.Run(); err != nil {
			return err
		}
		fmt.Printf("✅ Nginx service started\n")
		return nil
	default:
		return fmt.Errorf("automatic nginx service management not supported on %s", runtime.GOOS)
	}
}

// reloadNginx reloads the nginx configuration
func (ni *NginxInstaller) reloadNginx() error {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("brew", "services", "reload", "nginx")
		if err := cmd.Run(); err != nil {
			return err
		}
	default:
		cmd := exec.Command("sudo", "nginx", "-s", "reload")
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	fmt.Printf("✅ Nginx reloaded\n")
	return nil
}

// UninstallNginxConfig removes the nginx configuration
func (ni *NginxInstaller) UninstallNginxConfig() error {
	fmt.Printf("🗑️ Removing nginx configuration...\n")

	// Remove config file
	configFile := filepath.Join(ni.SitesPath, "vertex.conf")
	os.Remove(configFile)

	// Remove symlink (Linux)
	if runtime.GOOS == "linux" {
		os.Remove("/etc/nginx/sites-enabled/vertex.conf")
	}

	// Remove from hosts file
	ni.removeFromHosts()

	// Reload nginx
	ni.reloadNginx()

	fmt.Printf("✅ Nginx configuration removed\n")
	return nil
}

// removeFromHosts removes the domain from /etc/hosts
func (ni *NginxInstaller) removeFromHosts() error {
	hostsFile := "/etc/hosts"
	hostEntry := fmt.Sprintf("127.0.0.1 %s", ni.Domain)

	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	for _, line := range lines {
		if !strings.Contains(line, hostEntry) {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n")
	cmd := exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' > %s", newContent, hostsFile))
	return cmd.Run()
}

// EnableHTTPS enables HTTPS configuration
func (ni *NginxInstaller) EnableHTTPS(enable bool) {
	ni.HTTPSEnabled = enable
}

// isMkcertInstalled checks if mkcert is installed
func (ni *NginxInstaller) isMkcertInstalled() bool {
	_, err := exec.LookPath("mkcert")
	return err == nil
}

// installMkcert installs mkcert on the current platform
func (ni *NginxInstaller) installMkcert() error {
	switch runtime.GOOS {
	case "darwin":
		return ni.installMkcertMacOS()
	case "linux":
		return ni.installMkcertLinux()
	case "windows":
		return ni.installMkcertWindows()
	default:
		return fmt.Errorf("automatic mkcert installation not supported on %s", runtime.GOOS)
	}
}

// installMkcertMacOS installs mkcert on macOS using homebrew
func (ni *NginxInstaller) installMkcertMacOS() error {
	// Check if homebrew is installed
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("homebrew is required to install mkcert on macOS")
	}

	fmt.Printf("🔒 Installing mkcert via homebrew...\n")
	cmd := exec.Command("brew", "install", "mkcert")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install mkcert: %s", string(output))
	}

	fmt.Printf("✅ mkcert installed successfully\n")
	return nil
}

// installMkcertLinux installs mkcert on Linux
func (ni *NginxInstaller) installMkcertLinux() error {
	// Try different package managers
	packageManagers := []struct {
		command string
		args    []string
		name    string
	}{
		{"apt", []string{"install", "-y", "mkcert"}, "apt"},
		{"yum", []string{"install", "-y", "mkcert"}, "yum"},
		{"dnf", []string{"install", "-y", "mkcert"}, "dnf"},
		{"pacman", []string{"-S", "--noconfirm", "mkcert"}, "pacman"},
	}

	for _, pm := range packageManagers {
		if _, err := exec.LookPath(pm.command); err == nil {
			fmt.Printf("🔒 Installing mkcert using %s...\n", pm.name)
			
			args := append([]string{pm.command}, pm.args...)
			cmd := exec.Command("sudo", args...)
			
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("⚠️  Failed with %s: %s\n", pm.name, string(output))
				continue
			}
			
			fmt.Printf("✅ mkcert installed successfully using %s\n", pm.name)
			return nil
		}
	}

	return fmt.Errorf("no supported package manager found. Please install mkcert manually")
}

// installMkcertWindows installs mkcert on Windows
func (ni *NginxInstaller) installMkcertWindows() error {
	// Check if chocolatey is available
	if _, err := exec.LookPath("choco"); err == nil {
		fmt.Printf("🔒 Installing mkcert via chocolatey...\n")
		cmd := exec.Command("choco", "install", "mkcert", "-y")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install mkcert via chocolatey: %s", string(output))
		}
		fmt.Printf("✅ mkcert installed successfully via chocolatey\n")
		return nil
	}

	// Check if winget is available
	if _, err := exec.LookPath("winget"); err == nil {
		fmt.Printf("🔒 Installing mkcert via winget...\n")
		cmd := exec.Command("winget", "install", "mkcert")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install mkcert via winget: %s", string(output))
		}
		fmt.Printf("✅ mkcert installed successfully via winget\n")
		return nil
	}

	return fmt.Errorf("no package manager found. Please install mkcert manually")
}

// setupHTTPS sets up HTTPS certificates using mkcert
func (ni *NginxInstaller) setupHTTPS() error {
	if !ni.isMkcertInstalled() {
		fmt.Printf("🔒 mkcert not found, installing automatically...\n")
		if err := ni.installMkcert(); err != nil {
			return fmt.Errorf("failed to install mkcert: %v", err)
		}
	}

	fmt.Printf("🔒 Setting up HTTPS certificates for %s...\n", ni.Domain)

	// Install local CA
	fmt.Printf("🔐 Installing local certificate authority...\n")
	cmd := exec.Command("mkcert", "-install")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install local CA: %s", string(output))
	}

	// Create SSL directory
	sslDir := filepath.Join(os.Getenv("HOME"), ".vertex", "ssl")
	if err := os.MkdirAll(sslDir, 0755); err != nil {
		return fmt.Errorf("failed to create SSL directory: %v", err)
	}

	// Generate certificate for domain
	certFile := filepath.Join(sslDir, ni.Domain+".pem")
	keyFile := filepath.Join(sslDir, ni.Domain+"-key.pem")

	fmt.Printf("🔐 Generating trusted certificate for %s...\n", ni.Domain)
	cmd = exec.Command("mkcert", "-cert-file", certFile, "-key-file", keyFile, ni.Domain)
	cmd.Dir = sslDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %s", string(output))
	}

	fmt.Printf("✅ HTTPS certificates generated successfully\n")
	return nil
}