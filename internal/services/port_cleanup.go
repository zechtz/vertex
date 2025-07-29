package services

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// PortCleanupResult represents the result of port cleanup operation
type PortCleanupResult struct {
	Port         int      `json:"port"`
	ProcessesFound int    `json:"processesFound"`
	ProcessesKilled int   `json:"processesKilled"`
	PIDs         []int    `json:"pids"`
	Errors       []string `json:"errors"`
}

// KillProcessesOnPort finds and kills all processes using the specified port
func KillProcessesOnPort(port int) *PortCleanupResult {
	result := &PortCleanupResult{
		Port:            port,
		ProcessesFound:  0,
		ProcessesKilled: 0,
		PIDs:           []int{},
		Errors:         []string{},
	}

	log.Printf("[INFO] Cleaning up processes on port %d", port)

	// Find PIDs using the port
	pids := findProcessesOnPort(port)
	result.ProcessesFound = len(pids)
	result.PIDs = pids

	if len(pids) == 0 {
		log.Printf("[INFO] No processes found using port %d", port)
		return result
	}

	log.Printf("[INFO] Found %d process(es) using port %d: %v", len(pids), port, pids)

	// Kill each process
	for _, pid := range pids {
		if err := killProcessGracefully(pid); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to kill PID %d: %v", pid, err))
			log.Printf("[WARN] Failed to kill process %d: %v", pid, err)
		} else {
			result.ProcessesKilled++
			log.Printf("[INFO] Successfully killed process %d", pid)
		}
	}

	// Wait a moment for processes to die
	time.Sleep(1 * time.Second)

	// Verify cleanup
	remainingPids := findProcessesOnPort(port)
	if len(remainingPids) > 0 {
		log.Printf("[WARN] %d process(es) still using port %d after cleanup: %v", len(remainingPids), port, remainingPids)
		
		// Force kill remaining processes
		for _, pid := range remainingPids {
			if err := killProcessForcefully(pid); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to force kill PID %d: %v", pid, err))
				log.Printf("[ERROR] Failed to force kill process %d: %v", pid, err)
			} else {
				result.ProcessesKilled++
				log.Printf("[INFO] Force killed process %d", pid)
			}
		}
	}

	log.Printf("[INFO] Port cleanup completed: %d processes found, %d processes killed", result.ProcessesFound, result.ProcessesKilled)
	return result
}

// findProcessesOnPort finds all process IDs using the specified port
func findProcessesOnPort(port int) []int {
	var pids []int

	// Try lsof first (most reliable)
	if lsofPids := findProcessesWithLsof(port); len(lsofPids) > 0 {
		pids = append(pids, lsofPids...)
	}

	// Try netstat as fallback
	if len(pids) == 0 {
		if netstatPids := findProcessesWithNetstat(port); len(netstatPids) > 0 {
			pids = append(pids, netstatPids...)
		}
	}

	// Try fuser as another fallback
	if len(pids) == 0 {
		if fuserPids := findProcessesWithFuser(port); len(fuserPids) > 0 {
			pids = append(pids, fuserPids...)
		}
	}

	// Remove duplicates and invalid PIDs
	return deduplicateAndValidatePids(pids)
}

// findProcessesWithLsof uses lsof to find processes using the port
func findProcessesWithLsof(port int) []int {
	cmd := exec.Command("lsof", "-t", "-i", fmt.Sprintf(":%d", port))
	output, err := cmd.Output()
	if err != nil {
		return []int{}
	}

	return parsePidsFromOutput(string(output))
}

// findProcessesWithNetstat uses netstat to find processes using the port
func findProcessesWithNetstat(port int) []int {
	// Try different netstat variations based on OS
	commands := [][]string{
		{"netstat", "-tlnp"}, // Linux style
		{"netstat", "-an", "-p", "tcp"}, // Alternative
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		pids := parseNetstatOutput(string(output), port)
		if len(pids) > 0 {
			return pids
		}
	}

	return []int{}
}

// findProcessesWithFuser uses fuser to find processes using the port
func findProcessesWithFuser(port int) []int {
	cmd := exec.Command("fuser", fmt.Sprintf("%d/tcp", port))
	output, err := cmd.Output()
	if err != nil {
		return []int{}
	}

	return parsePidsFromOutput(string(output))
}

// parseNetstatOutput parses netstat output to extract PIDs for the given port
func parseNetstatOutput(output string, port int) []int {
	var pids []int
	lines := strings.Split(output, "\n")
	portStr := fmt.Sprintf(":%d", port)

	for _, line := range lines {
		if strings.Contains(line, portStr) && strings.Contains(line, "LISTEN") {
			fields := strings.Fields(line)
			// Look for PID in the last field (format: PID/program)
			for _, field := range fields {
				if strings.Contains(field, "/") {
					pidStr := strings.Split(field, "/")[0]
					if pid, err := strconv.Atoi(pidStr); err == nil {
						pids = append(pids, pid)
					}
				}
			}
		}
	}

	return pids
}

// parsePidsFromOutput parses space or newline separated PIDs from command output
func parsePidsFromOutput(output string) []int {
	var pids []int
	output = strings.TrimSpace(output)
	if output == "" {
		return pids
	}

	// Split by both spaces and newlines
	pidStrings := strings.FieldsFunc(output, func(c rune) bool {
		return c == ' ' || c == '\n' || c == '\t'
	})

	for _, pidStr := range pidStrings {
		pidStr = strings.TrimSpace(pidStr)
		if pidStr == "" {
			continue
		}
		
		if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
			pids = append(pids, pid)
		}
	}

	return pids
}

// deduplicateAndValidatePids removes duplicates and validates that PIDs exist
func deduplicateAndValidatePids(pids []int) []int {
	seen := make(map[int]bool)
	var result []int

	for _, pid := range pids {
		if seen[pid] || pid <= 0 {
			continue
		}

		// Check if process actually exists
		if processExists(pid) {
			result = append(result, pid)
			seen[pid] = true
		}
	}

	return result
}

// processExists checks if a process with the given PID exists
func processExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix systems, os.FindProcess always succeeds, so we need to check if we can signal it
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// killProcessGracefully attempts to kill a process gracefully (SIGTERM first)
func killProcessGracefully(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	// Try SIGTERM first
	log.Printf("[INFO] Sending SIGTERM to process %d", pid)
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to process %d: %w", pid, err)
	}

	// Wait a bit for graceful shutdown
	time.Sleep(2 * time.Second)

	// Check if process is still alive
	if processExists(pid) {
		// Still alive, try SIGKILL
		return killProcessForcefully(pid)
	}

	return nil
}

// killProcessForcefully kills a process with SIGKILL
func killProcessForcefully(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	log.Printf("[INFO] Force killing process %d with SIGKILL", pid)
	if err := process.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to send SIGKILL to process %d: %w", pid, err)
	}

	return nil
}

// CleanupPortBeforeStart ensures a port is available before starting a service
func CleanupPortBeforeStart(port int) error {
	result := KillProcessesOnPort(port)
	
	if len(result.Errors) > 0 {
		log.Printf("[WARN] Port cleanup had %d error(s): %v", len(result.Errors), result.Errors)
	}

	// Final verification that port is available
	finalCheck := findProcessesOnPort(port)
	if len(finalCheck) > 0 {
		return fmt.Errorf("port %d is still in use by process(es): %v", port, finalCheck)
	}

	return nil
}