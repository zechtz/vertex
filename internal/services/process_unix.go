//go:build !windows

package services

import (
	"os/exec"
	"syscall"
)

// SetProcessGroup sets the process group for Unix systems
func SetProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// KillProcess kills a process on Unix systems
func KillProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}

// ForceKillProcess force kills a process on Unix systems
func ForceKillProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}

// GetProcessGroup gets the process group ID on Unix systems
func GetProcessGroup(pid int) (int, error) {
	return syscall.Getpgid(pid)
}

// KillProcessGroup kills a process group on Unix systems
func KillProcessGroup(pgid int) error {
	return syscall.Kill(-pgid, syscall.SIGTERM)
}

// ForceKillProcessGroup force kills a process group on Unix systems
func ForceKillProcessGroup(pgid int) error {
	return syscall.Kill(-pgid, syscall.SIGKILL)
}

// IsProcessRunning checks if a process is running on Unix systems
func IsProcessRunning(pid int) bool {
	// Use signal 0 to check if process exists
	err := syscall.Kill(pid, 0)
	return err == nil
}
