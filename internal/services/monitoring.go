// Package services - Resource monitoring functionality
package services

import (
	"fmt"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/zechtz/nest-up/internal/models"
)

// collectResourceMetrics collects CPU, memory, and network metrics for a service
func (sm *Manager) collectResourceMetrics(service *models.Service) error {
	if service.PID <= 0 {
		// Reset metrics for stopped services
		service.CPUPercent = 0
		service.MemoryUsage = 0
		service.MemoryPercent = 0
		service.DiskUsage = 0
		service.NetworkRx = 0
		service.NetworkTx = 0
		return nil
	}

	// Get process handle
	proc, err := process.NewProcess(int32(service.PID))
	if err != nil {
		log.Printf("[DEBUG] Failed to get process handle for %s (PID %d): %v", service.Name, service.PID, err)
		return err
	}

	// Check if process is still running
	isRunning, err := proc.IsRunning()
	if err != nil || !isRunning {
		log.Printf("[DEBUG] Process %d for service %s is no longer running", service.PID, service.Name)
		// Reset metrics for stopped processes
		service.CPUPercent = 0
		service.MemoryUsage = 0
		service.MemoryPercent = 0
		service.DiskUsage = 0
		service.NetworkRx = 0
		service.NetworkTx = 0
		return fmt.Errorf("process no longer running")
	}

	// Collect CPU usage
	cpuPercent, err := proc.CPUPercent()
	if err != nil {
		log.Printf("[DEBUG] Failed to get CPU usage for %s: %v", service.Name, err)
	} else {
		service.CPUPercent = cpuPercent
	}

	// Collect memory usage
	memInfo, err := proc.MemoryInfo()
	if err != nil {
		log.Printf("[DEBUG] Failed to get memory info for %s: %v", service.Name, err)
	} else {
		service.MemoryUsage = memInfo.RSS // Resident Set Size (physical memory)
	}

	// Collect memory percentage
	memPercent, err := proc.MemoryPercent()
	if err != nil {
		log.Printf("[DEBUG] Failed to get memory percentage for %s: %v", service.Name, err)
	} else {
		service.MemoryPercent = memPercent
	}

	// Collect I/O statistics (disk usage)
	ioCounters, err := proc.IOCounters()
	if err != nil {
		log.Printf("[DEBUG] Failed to get I/O counters for %s: %v", service.Name, err)
	} else {
		service.DiskUsage = ioCounters.ReadBytes + ioCounters.WriteBytes
	}

	// Collect network statistics for child processes
	// Note: Direct network stats per process are complex, so we'll track at system level
	// For now, we'll collect network I/O as part of the process I/O
	if ioCounters != nil {
		service.NetworkRx = ioCounters.ReadCount
		service.NetworkTx = ioCounters.WriteCount
	}

	log.Printf("[DEBUG] Collected metrics for %s - CPU: %.2f%%, Memory: %d bytes (%.2f%%)", 
		service.Name, service.CPUPercent, service.MemoryUsage, service.MemoryPercent)

	return nil
}

// startMetricsCollection starts periodic resource monitoring for all services
func (sm *Manager) startMetricsCollection() {
	ticker := time.NewTicker(10 * time.Second) // Collect metrics every 10 seconds
	defer ticker.Stop()

	log.Printf("[INFO] Started resource metrics collection (10s interval)")

	for {
		select {
		case <-ticker.C:
			sm.collectAllServiceMetrics()
		}
	}
}

// collectAllServiceMetrics collects metrics for all running services
func (sm *Manager) collectAllServiceMetrics() {
	sm.mutex.RLock()
	services := make([]*models.Service, 0, len(sm.services))
	for _, service := range sm.services {
		services = append(services, service)
	}
	sm.mutex.RUnlock()

	for _, service := range services {
		service.Mutex.Lock()
		if service.Status == "running" && service.PID > 0 {
			if err := sm.collectResourceMetrics(service); err != nil {
				// If metrics collection fails, the process might have stopped
				if !sm.isProcessRunning(service.PID) {
					log.Printf("[INFO] Process %d for service %s stopped, updating status", service.PID, service.Name)
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
					sm.updateServiceInDB(service)
					sm.broadcastUpdate(service)
				}
			} else {
				// Successful metrics collection, broadcast update
				sm.broadcastUpdate(service)
			}
		}
		service.Mutex.Unlock()
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
	
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var totalCPU float64
	var totalMemory uint64
	runningServices := 0

	for _, service := range sm.services {
		service.Mutex.RLock()
		if service.Status == "running" {
			runningServices++
			totalCPU += service.CPUPercent
			totalMemory += service.MemoryUsage
		}
		service.Mutex.RUnlock()
	}

	summary["runningServices"] = runningServices
	summary["totalServices"] = len(sm.services)
	summary["totalCPU"] = totalCPU
	summary["totalMemory"] = totalMemory
	summary["timestamp"] = time.Now()

	return summary
}