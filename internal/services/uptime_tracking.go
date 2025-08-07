// Package services - Uptime tracking functionality
package services

import (
	"log"
	"sync"
	"time"

	"github.com/zechtz/vertex/internal/models"
)

type UptimeEvent struct {
	ServiceID string    `json:"serviceId"`
	EventType string    `json:"eventType"` // "start", "stop", "restart"
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // "running", "stopped", "unhealthy"
}

type UptimeTracker struct {
	events map[string][]UptimeEvent // serviceID -> events
	mutex  sync.RWMutex
}

var uptimeTracker *UptimeTracker
var once sync.Once

// GetUptimeTracker returns the singleton uptime tracker instance
func GetUptimeTracker() *UptimeTracker {
	once.Do(func() {
		uptimeTracker = &UptimeTracker{
			events: make(map[string][]UptimeEvent),
		}
	})
	return uptimeTracker
}

// RecordEvent records a service state change event
func (ut *UptimeTracker) RecordEvent(serviceID, eventType, status string) {
	ut.mutex.Lock()
	defer ut.mutex.Unlock()

	event := UptimeEvent{
		ServiceID: serviceID,
		EventType: eventType,
		Timestamp: time.Now(),
		Status:    status,
	}

	if ut.events[serviceID] == nil {
		ut.events[serviceID] = make([]UptimeEvent, 0)
	}

	ut.events[serviceID] = append(ut.events[serviceID], event)

	// Keep only last 1000 events per service to prevent memory issues
	if len(ut.events[serviceID]) > 1000 {
		ut.events[serviceID] = ut.events[serviceID][len(ut.events[serviceID])-1000:]
	}

	log.Printf("[DEBUG] Recorded uptime event for %s: %s -> %s", serviceID, eventType, status)
}

// CalculateUptimeStats calculates uptime statistics for a service
func (ut *UptimeTracker) CalculateUptimeStats(serviceID string, service *models.Service) models.UptimeStatistics {
	ut.mutex.RLock()
	defer ut.mutex.RUnlock()

	events := ut.events[serviceID]
	if len(events) == 0 {
		return models.UptimeStatistics{
			UptimePercentage24h: 100.0,
			UptimePercentage7d:  100.0,
			MTBF:                0,
			TotalRestarts:       0,
		}
	}

	now := time.Now()
	day24h := now.Add(-24 * time.Hour)
	day7d := now.Add(-7 * 24 * time.Hour)

	stats := models.UptimeStatistics{}

	// Count restarts and calculate MTBF
	restarts := 0
	var failures []time.Time
	var lastDowntime time.Time

	for _, event := range events {
		if event.EventType == "restart" || (event.EventType == "start" && event.Status == "running") {
			restarts++
		}
		if event.Status == "stopped" || event.Status == "unhealthy" {
			failures = append(failures, event.Timestamp)
			if event.Timestamp.After(lastDowntime) {
				lastDowntime = event.Timestamp
			}
		}
	}

	stats.TotalRestarts = restarts
	stats.LastDowntime = lastDowntime

	// Calculate MTBF (Mean Time Between Failures)
	if len(failures) > 1 {
		totalTimeBetweenFailures := time.Duration(0)
		for i := 1; i < len(failures); i++ {
			totalTimeBetweenFailures += failures[i].Sub(failures[i-1])
		}
		stats.MTBF = totalTimeBetweenFailures / time.Duration(len(failures)-1)
	}

	// Calculate uptime percentages
	stats.UptimePercentage24h = ut.calculateUptimePercentage(events, day24h, now)
	stats.UptimePercentage7d = ut.calculateUptimePercentage(events, day7d, now)

	// Calculate total downtime
	stats.TotalDowntime24h = ut.calculateDowntime(events, day24h, now)
	stats.TotalDowntime7d = ut.calculateDowntime(events, day7d, now)

	return stats
}

// calculateUptimePercentage calculates uptime percentage within a time range
func (ut *UptimeTracker) calculateUptimePercentage(events []UptimeEvent, start, end time.Time) float64 {
	if len(events) == 0 {
		return 100.0 // Assume 100% if no events recorded
	}

	totalTime := end.Sub(start)
	downTime := time.Duration(0)
	isDown := false
	downStart := time.Time{}

	// Check initial state at start time
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Timestamp.Before(start) {
			if events[i].Status == "stopped" || events[i].Status == "unhealthy" {
				isDown = true
				downStart = start
			}
			break
		}
	}

	// Process events within the time range
	for _, event := range events {
		if event.Timestamp.Before(start) {
			continue
		}
		if event.Timestamp.After(end) {
			break
		}

		if event.Status == "stopped" || event.Status == "unhealthy" {
			if !isDown {
				isDown = true
				downStart = event.Timestamp
			}
		} else if event.Status == "running" || event.Status == "healthy" {
			if isDown {
				downTime += event.Timestamp.Sub(downStart)
				isDown = false
			}
		}
	}

	// If still down at end time
	if isDown {
		downTime += end.Sub(downStart)
	}

	uptime := totalTime - downTime
	if totalTime == 0 {
		return 100.0
	}

	percentage := float64(uptime) / float64(totalTime) * 100
	if percentage < 0 {
		return 0.0
	}
	if percentage > 100 {
		return 100.0
	}

	return percentage
}

// calculateDowntime calculates total downtime within a time range
func (ut *UptimeTracker) calculateDowntime(events []UptimeEvent, start, end time.Time) time.Duration {
	if len(events) == 0 {
		return 0
	}

	downTime := time.Duration(0)
	isDown := false
	downStart := time.Time{}

	// Check initial state at start time
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Timestamp.Before(start) {
			if events[i].Status == "stopped" || events[i].Status == "unhealthy" {
				isDown = true
				downStart = start
			}
			break
		}
	}

	// Process events within the time range
	for _, event := range events {
		if event.Timestamp.Before(start) {
			continue
		}
		if event.Timestamp.After(end) {
			break
		}

		if event.Status == "stopped" || event.Status == "unhealthy" {
			if !isDown {
				isDown = true
				downStart = event.Timestamp
			}
		} else if event.Status == "running" || event.Status == "healthy" {
			if isDown {
				downTime += event.Timestamp.Sub(downStart)
				isDown = false
			}
		}
	}

	// If still down at end time
	if isDown {
		downTime += end.Sub(downStart)
	}

	return downTime
}

// GetAllUptimeStats returns uptime statistics for all services
func (ut *UptimeTracker) GetAllUptimeStats() map[string]models.UptimeStatistics {
	ut.mutex.RLock()
	defer ut.mutex.RUnlock()

	stats := make(map[string]models.UptimeStatistics)
	for serviceID := range ut.events {
		stats[serviceID] = ut.CalculateUptimeStats(serviceID, nil)
	}

	return stats
}
