//go:build windows

package services

import (
	"os/exec"
	"syscall"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procTerminateProcess = kernel32.NewProc("TerminateProcess")
	procOpenProcess      = kernel32.NewProc("OpenProcess")
	procCloseHandle      = kernel32.NewProc("CloseHandle")
)

// SetProcessGroup sets the process group for Windows systems
func SetProcessGroup(cmd *exec.Cmd) {
	// On Windows, create a new process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// KillProcess kills a process on Windows systems
func KillProcess(pid int) error {
	return terminateProcess(pid)
}

// ForceKillProcess force kills a process on Windows systems
func ForceKillProcess(pid int) error {
	return terminateProcess(pid)
}

// GetProcessGroup gets the process group ID on Windows systems
// On Windows, we return the PID itself as the "group"
func GetProcessGroup(pid int) (int, error) {
	return pid, nil
}

// KillProcessGroup kills a process group on Windows systems
func KillProcessGroup(pgid int) error {
	return terminateProcess(pgid)
}

// ForceKillProcessGroup force kills a process group on Windows systems
func ForceKillProcessGroup(pgid int) error {
	return terminateProcess(pgid)
}

// terminateProcess terminates a process on Windows
func terminateProcess(pid int) error {
	const PROCESS_TERMINATE = 0x0001

	handle, _, _ := procOpenProcess.Call(
		PROCESS_TERMINATE,
		0,
		uintptr(pid),
	)

	if handle == 0 {
		return syscall.GetLastError()
	}
	defer procCloseHandle.Call(handle)

	ret, _, _ := procTerminateProcess.Call(handle, 1)
	if ret == 0 {
		return syscall.GetLastError()
	}

	return nil
}

// IsProcessRunning checks if a process is running on Windows systems
func IsProcessRunning(pid int) bool {
	const PROCESS_QUERY_INFORMATION = 0x0400

	handle, _, _ := procOpenProcess.Call(
		PROCESS_QUERY_INFORMATION,
		0,
		uintptr(pid),
	)

	if handle == 0 {
		return false
	}
	defer procCloseHandle.Call(handle)

	return true
}
