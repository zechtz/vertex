package installer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func UpdateService() error {
	switch runtime.GOOS {
	case "darwin":
		return updateMacOS()
	case "linux":
		return updateLinux()
	case "windows":
		return updateWindows()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func updateMacOS() error {
	fmt.Println("Updating Vertex on macOS...")

	// Stop all vertex services
	cmd := exec.Command("launchctl", "stop", "com.vertex.manager")
	if err := cmd.Run(); err != nil {
		fmt.Println("Could not stop launchctl service (might not be running):", err)
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get working directory: %w", err)
	}

	// Source path is vertex binary in current directory
	srcPath := filepath.Join(wd, "vertex")

	// Destination path
	installDir := filepath.Join(os.Getenv("HOME"), ".local", "bin", "vertex")

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(installDir), 0755); err != nil {
		return fmt.Errorf("could not create destination directory: %w", err)
	}

	// Copy binary
	if err := replaceFile(srcPath, installDir); err != nil {
		return fmt.Errorf("could not replace binary: %w", err)
	}

	// Start the service
	cmd = exec.Command("launchctl", "start", "com.vertex.manager")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not start launchctl service: %w", err)
	}

	// Restart nginx
	cmd = exec.Command("brew", "services", "restart", "nginx")
	if err := cmd.Run(); err != nil {
		fmt.Println("Could not restart nginx (might not be installed):", err)
	}

	fmt.Println("Vertex updated successfully!")
	return nil
}

func updateLinux() error {
	fmt.Println("Updating Vertex on Linux...")

	// Stop all vertex services
	cmd := exec.Command("systemctl", "--user", "stop", "vertex")
	if err := cmd.Run(); err != nil {
		fmt.Println("Could not stop systemd service (might not be running):", err)
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get working directory: %w", err)
	}

	// Source path is vertex binary in current directory
	srcPath := filepath.Join(wd, "vertex")

	// Destination path
	installDir := filepath.Join(os.Getenv("HOME"), ".local", "bin", "vertex")

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(installDir), 0755); err != nil {
		return fmt.Errorf("could not create destination directory: %w", err)
	}

	// Copy binary
	if err := replaceFile(srcPath, installDir); err != nil {
		return fmt.Errorf("could not replace binary: %w", err)
	}

	// Start the service
	cmd = exec.Command("systemctl", "--user", "start", "vertex")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not start systemd service: %w", err)
	}

	// Restart nginx
	cmd = exec.Command("sudo", "systemctl", "restart", "nginx")
	if err := cmd.Run(); err != nil {
		fmt.Println("Could not restart nginx (might not be installed):", err)
	}

	fmt.Println("Vertex updated successfully!")
	return nil
}

func updateWindows() error {
	fmt.Println("Updating Vertex on Windows...")

	// Stop all vertex services
	cmd := exec.Command("sc.exe", "stop", "vertex")
	if err := cmd.Run(); err != nil {
		fmt.Println("Could not stop service (might not be running):", err)
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get working directory: %w", err)
	}

	// Source path is vertex binary in current directory
	srcPath := filepath.Join(wd, "vertex.exe")

	// Destination path
	installDir := filepath.Join(os.Getenv("ProgramFiles"), "Vertex", "vertex.exe")

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(installDir), 0755); err != nil {
		return fmt.Errorf("could not create destination directory: %w", err)
	}

	// Copy binary
	if err := replaceFile(srcPath, installDir); err != nil {
		return fmt.Errorf("could not replace binary: %w", err)
	}

	// Start the service
	cmd = exec.Command("sc.exe", "start", "vertex")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not start service: %w", err)
	}

	// Restart nginx
	cmd = exec.Command("nginx.exe", "-s", "reload")
	if err := cmd.Run(); err != nil {
		fmt.Println("Could not restart nginx (might not be in PATH):", err)
	}

	fmt.Println("Vertex updated successfully!")
	return nil
}

// replaceFile copies the src file to the dst file, replacing it if it exists,
// and making it executable.
func replaceFile(src, dst string) error {
	// Check if source file exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %w", err)
	}

	// Remove existing destination file if it exists
	if _, err := os.Stat(dst); err == nil {
		if err := os.Remove(dst); err != nil {
			return fmt.Errorf("failed to remove existing destination file: %w", err)
		}
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Set executable permissions (not needed for Windows, but safe to call)
	return os.Chmod(dst, 0755)
}
