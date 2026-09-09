// Package services - Resource monitoring functionality
package services

import (
	"fmt"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/zechtz/vertex/internal/models"
)

// resourceMetrics is one sample of a process's resource usage.
type resourceMetrics struct {
	cpuPercent    float64
	memoryUsage   uint64
	memoryPercent float32
	diskUsage     uint64
	networkRx     uint64
	networkTx     uint64
}

// collectResourceMetrics samples CPU, memory, and I/O for a running process. It
// takes no locks: gopsutil reads can block on a busy system, and the service
// lock must not be held while they do.
func (sm *Manager) collectResourceMetrics(name string, pid int) (resourceMetrics, error) {
	var metrics resourceMetrics

	// Get process handle
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		log.Printf("[DEBUG] Failed to get process handle for %s (PID %d): %v", name, pid, err)
		return metrics, err
	}

	// Check if process is still running
	isRunning, err := proc.IsRunning()
	if err != nil || !isRunning {
		log.Printf("[DEBUG] Process %d for service %s is no longer running", pid, name)
		return metrics, fmt.Errorf("process no longer running")
	}

	// Collect CPU usage
	cpuPercent, err := proc.CPUPercent()
	if err != nil {
		log.Printf("[DEBUG] Failed to get CPU usage for %s: %v", name, err)
	} else {
		metrics.cpuPercent = cpuPercent
	}

	// Collect memory usage
	memInfo, err := proc.MemoryInfo()
	if err != nil {
		log.Printf("[DEBUG] Failed to get memory info for %s: %v", name, err)
	} else {
		metrics.memoryUsage = memInfo.RSS // Resident Set Size (physical memory)
	}

	// Collect memory percentage
	memPercent, err := proc.MemoryPercent()
	if err != nil {
		log.Printf("[DEBUG] Failed to get memory percentage for %s: %v", name, err)
	} else {
		metrics.memoryPercent = memPercent
	}

	// Collect I/O statistics (disk usage) - not available on every platform
	// (notably macOS), where the zero values above are the right answer
	ioCounters, err := proc.IOCounters()
	if err == nil {
		metrics.diskUsage = ioCounters.ReadBytes + ioCounters.WriteBytes
		// Collect network statistics using I/O counters as a proxy
		metrics.networkRx = ioCounters.ReadCount
		metrics.networkTx = ioCounters.WriteCount
	}

	log.Printf("[DEBUG] Collected metrics for %s - CPU: %.2f%%, Memory: %d bytes (%.2f%%)",
		name, metrics.cpuPercent, metrics.memoryUsage, metrics.memoryPercent)

	return metrics, nil
}

// startMetricsCollection starts periodic resource monitoring for all services
func (sm *Manager) startMetricsCollection() {
	ticker := time.NewTicker(10 * time.Second) // Collect metrics every 10 seconds
	defer ticker.Stop()

	log.Printf("[INFO] Started resource metrics collection (10s interval)")

	for range ticker.C {
		sm.collectAllServiceMetrics()
	}
}

// collectAllServiceMetrics collects metrics for all running services
func (sm *Manager) collectAllServiceMetrics() {
	// Snapshot the tracked services, then release the manager lock before
	// touching any service lock - holding both lets one busy service block
	// every reader of the manager.
	sm.mutex.RLock()
	services := make([]*models.Service, 0, len(sm.services))
	for _, service := range sm.services {
		services = append(services, service)
	}
	sm.mutex.RUnlock()

	for _, service := range services {
		service.Mutex.RLock()
		name, status, pid := service.Name, service.Status, service.PID
		service.Mutex.RUnlock()

		if status != "running" || pid <= 0 {
			continue
		}

		metrics, err := sm.collectResourceMetrics(name, pid)
		if err != nil {
			// If metrics collection fails, the process might have stopped
			if sm.isProcessRunning(pid) {
				continue
			}

			log.Printf("[INFO] Process %d for service %s stopped, updating status", pid, name)

			service.Mutex.Lock()
			service.Status = "stopped"
			service.HealthStatus = "unknown"
			service.PID = 0
			service.Cmd = nil
			service.Uptime = ""
			// Reset metrics
			service.CPUPercent = 0
			service.MemoryUsage = 0
			service.MemoryPercent = 0
			service.DiskUsage = 0
			service.NetworkRx = 0
			service.NetworkTx = 0
			service.Mutex.Unlock()

			// Record uptime event
			GetUptimeTracker().RecordEvent(service.ID, "stop", "stopped")

			sm.publishServiceState(service)
			continue
		}

		uptimeStats := GetUptimeTracker().CalculateUptimeStats(service.ID, service)

		service.Mutex.Lock()
		service.CPUPercent = metrics.cpuPercent
		service.MemoryUsage = metrics.memoryUsage
		service.MemoryPercent = metrics.memoryPercent
		service.DiskUsage = metrics.diskUsage
		service.NetworkRx = metrics.networkRx
		service.NetworkTx = metrics.networkTx
		service.Metrics.UptimeStats = uptimeStats
		service.Mutex.Unlock()

		sm.publishServiceState(service)
	}
}

// collectPerformanceMetrics collects response time and error rate metrics
func (sm *Manager) collectPerformanceMetrics(service *models.Service) error {
	if service.HealthURL == "" || service.Status != "running" {
		return nil
	}

	start := time.Now()

	// Perform HTTP request to health endpoint
	client := sm.createHealthCheckClient()
	req, err := sm.createHealthCheckRequest(service.HealthURL)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	duration := time.Since(start)

	// Record response time
	responseTime := models.ResponseTime{
		Timestamp: time.Now(),
		Duration:  duration,
		Status:    0, // Default for failed requests
	}

	if err == nil {
		responseTime.Status = resp.StatusCode
		resp.Body.Close()
	}

	// Update service metrics
	service.Mutex.Lock()
	defer service.Mutex.Unlock()

	// Initialize metrics if needed
	if service.Metrics.ResponseTimes == nil {
		service.Metrics.ResponseTimes = make([]models.ResponseTime, 0)
	}

	// Add new response time and keep last 100 entries
	service.Metrics.ResponseTimes = append(service.Metrics.ResponseTimes, responseTime)
	if len(service.Metrics.ResponseTimes) > 100 {
		service.Metrics.ResponseTimes = service.Metrics.ResponseTimes[len(service.Metrics.ResponseTimes)-100:]
	}

	// Update request count
	service.Metrics.RequestCount++

	// Calculate error rate (last 10 requests)
	recentRequests := service.Metrics.ResponseTimes
	if len(recentRequests) > 10 {
		recentRequests = recentRequests[len(recentRequests)-10:]
	}

	errorCount := 0
	for _, rt := range recentRequests {
		if rt.Status >= 400 || rt.Status == 0 {
			errorCount++
		}
	}

	if len(recentRequests) > 0 {
		service.Metrics.ErrorRate = float64(errorCount) / float64(len(recentRequests)) * 100
	}

	service.Metrics.LastChecked = time.Now()

	log.Printf("[DEBUG] Performance metrics for %s - Response time: %v, Status: %d, Error rate: %.2f%%",
		service.Name, duration, responseTime.Status, service.Metrics.ErrorRate)

	return nil
}

// getSystemResourceSummary returns overall system resource usage
func (sm *Manager) getSystemResourceSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	// Release the manager lock before taking any service lock
	sm.mutex.RLock()
	tracked := make([]*models.Service, 0, len(sm.services))
	for _, service := range sm.services {
		tracked = append(tracked, service)
	}
	sm.mutex.RUnlock()

	var totalCPU float64
	var totalMemory uint64
	runningServices := 0

	for _, service := range tracked {
		service.Mutex.RLock()
		if service.Status == "running" {
			runningServices++
			totalCPU += service.CPUPercent
			totalMemory += service.MemoryUsage
		}
		service.Mutex.RUnlock()
	}

	summary["runningServices"] = runningServices
	summary["totalServices"] = len(tracked)
	summary["totalCPU"] = totalCPU
	summary["totalMemory"] = totalMemory
	summary["timestamp"] = time.Now()

	return summary
}
