//go:build windows

package services

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// SetProcessGroup is a no-op on Windows.
func SetProcessGroup(cmd *exec.Cmd) {
	// Windows does not support setting a process group ID in the same way as Unix.
}

// KillProcess kills a process on Windows using its PID.
func KillProcess(pid int) error {
	return exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid)).Run()
}

// ForceKillProcess force kills a process and its children on Windows.
func ForceKillProcess(pid int) error {
	return exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid)).Run()
}

// GetProcessGroup is not supported on Windows and will always return an error.
func GetProcessGroup(pid int) (int, error) {
	// Return the pid itself and no error, so the calling code can try to kill it.
	return pid, nil
}

// KillProcessGroup kills a process tree on Windows.
func KillProcessGroup(pgid int) error {
	// On Windows, we can't kill a group by pgid. We kill the process and its children.
	return exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pgid)).Run()
}

// ForceKillProcessGroup kills a process tree on Windows.
func ForceKillProcessGroup(pgid int) error {
	return exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pgid)).Run()
}

// IsProcessRunning checks if a process with the given PID is running on Windows.
func IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	// Use tasklist command to check for the process
	// The "2>nul" part is to redirect stderr to null, so we don't see errors if the process doesn't exist.
	cmd := exec.Command("cmd", "/C", "tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "2>nul")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	// If the output contains the PID, the process is running
	return strings.Contains(string(output), strconv.Itoa(pid))
}
