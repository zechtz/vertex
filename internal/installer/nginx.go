// Package installer
package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// NginxInstaller handles nginx configuration for Vertex
type NginxInstaller struct {
	Domain     string
	Port       string
	ConfigPath string
	SitesPath  string
}

// NewNginxInstaller creates a new nginx installer
func NewNginxInstaller(domain, port string) *NginxInstaller {
	ni := &NginxInstaller{
		Domain: domain,
		Port:   port,
	}

	// Set platform-specific paths
	switch runtime.GOOS {
	case "darwin":
		// Detect homebrew installation path
		if _, err := os.Stat("/opt/homebrew/etc/nginx"); err == nil {
			// Apple Silicon homebrew path
			ni.ConfigPath = "/opt/homebrew/etc/nginx/nginx.conf"
			ni.SitesPath = "/opt/homebrew/etc/nginx/servers"
		} else {
			// Intel homebrew path
			ni.ConfigPath = "/usr/local/etc/nginx/nginx.conf"
			ni.SitesPath = "/usr/local/etc/nginx/servers"
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
	
	switch runtime.GOOS {
	case "darwin":
		// Detect homebrew path and create corresponding directories
		if strings.Contains(ni.ConfigPath, "/opt/homebrew/") {
			directories = []string{
				"/opt/homebrew/var/log/nginx",
				"/opt/homebrew/var/run",
			}
		} else {
			directories = []string{
				"/usr/local/var/log/nginx",
				"/usr/local/var/run",
			}
		}
	case "linux":
		directories = []string{
			"/var/log/nginx",
			"/var/run/nginx",
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
	fmt.Printf("🌐 Vertex is now available at: http://%s\n", ni.Domain)
	return nil
}

// createNginxConfig creates the nginx configuration file
func (ni *NginxInstaller) createNginxConfig(configFile string) error {
	config := fmt.Sprintf(`# Vertex Service Manager Configuration
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
		// Use brew services on macOS
		cmd := exec.Command("brew", "services", "restart", "nginx")
		if err := cmd.Run(); err != nil {
			return err
		}
		fmt.Printf("✅ Nginx service started\n")
		return nil
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