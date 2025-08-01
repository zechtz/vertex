// Package database - Log storage functionality
package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/zechtz/vertex/internal/models"
)

// InitializeLogTables creates the log-related database tables
func (db *Database) InitializeLogTables() error {
	// Create logs table for persistent log storage
	createLogsTable := `
		CREATE TABLE IF NOT EXISTS service_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			service_id TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(service_id) REFERENCES services(id)
		);
	`

	// Create indexes for better search performance
	createIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_service_logs_service_id ON service_logs(service_id);`,
		`CREATE INDEX IF NOT EXISTS idx_service_logs_timestamp ON service_logs(timestamp);`,
		`CREATE INDEX IF NOT EXISTS idx_service_logs_level ON service_logs(level);`,
		`CREATE INDEX IF NOT EXISTS idx_service_logs_message_fts ON service_logs(message);`,
		`CREATE INDEX IF NOT EXISTS idx_service_logs_created_at ON service_logs(created_at);`,
	}

	// Create log retention settings table
	createRetentionTable := `
		CREATE TABLE IF NOT EXISTS log_retention_settings (
			id INTEGER PRIMARY KEY,
			max_logs_per_service INTEGER DEFAULT 10000,
			retention_days INTEGER DEFAULT 7,
			auto_cleanup_enabled BOOLEAN DEFAULT TRUE,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	// Insert default retention settings
	insertDefaultRetention := `
		INSERT OR IGNORE INTO log_retention_settings (id, max_logs_per_service, retention_days, auto_cleanup_enabled)
		VALUES (1, 10000, 7, TRUE);
	`

	// Execute table creation
	if _, err := db.DB.Exec(createLogsTable); err != nil {
		return fmt.Errorf("failed to create service_logs table: %w", err)
	}

	// Create indexes
	for _, indexSQL := range createIndexes {
		if _, err := db.DB.Exec(indexSQL); err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}

	// Create retention settings table
	if _, err := db.DB.Exec(createRetentionTable); err != nil {
		return fmt.Errorf("failed to create log_retention_settings table: %w", err)
	}

	// Insert default settings
	if _, err := db.DB.Exec(insertDefaultRetention); err != nil {
		return fmt.Errorf("failed to insert default retention settings: %w", err)
	}

	log.Printf("[INFO] Log storage tables initialized successfully")
	return nil
}

// StoreLogEntry stores a single log entry in the database
func (db *Database) StoreLogEntry(serviceID string, logEntry models.LogEntry) error {
	query := `
		INSERT INTO service_logs (service_id, timestamp, level, message)
		VALUES (?, ?, ?, ?)
	`

	// Parse timestamp from log entry
	timestamp, err := time.Parse(time.RFC3339Nano, logEntry.Timestamp)
	if err != nil {
		log.Printf("[WARN] Failed to parse log timestamp %s: %v", logEntry.Timestamp, err)
		timestamp = time.Now()
	}

	_, err = db.DB.Exec(query, serviceID, timestamp, logEntry.Level, logEntry.Message)
	if err != nil {
		return fmt.Errorf("failed to store log entry for service %s: %w", serviceID, err)
	}

	return nil
}

// StoreLogEntries stores multiple log entries in a single transaction
func (db *Database) StoreLogEntries(serviceID string, logEntries []models.LogEntry) error {
	if len(logEntries) == 0 {
		return nil
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO service_logs (service_id, timestamp, level, message)
		VALUES (?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, logEntry := range logEntries {
		timestamp, err := time.Parse(time.RFC3339Nano, logEntry.Timestamp)
		if err != nil {
			log.Printf("[WARN] Failed to parse log timestamp %s: %v", logEntry.Timestamp, err)
			timestamp = time.Now()
		}

		_, err = stmt.Exec(serviceID, timestamp, logEntry.Level, logEntry.Message)
		if err != nil {
			return fmt.Errorf("failed to execute log insert for service %s: %w", serviceID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit log entries transaction: %w", err)
	}

	return nil
}

// LogSearchCriteria defines search parameters for log queries
type LogSearchCriteria struct {
	ServiceIDs   []string  `json:"serviceIds"`
	Levels       []string  `json:"levels"`
	SearchText   string    `json:"searchText"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	Limit        int       `json:"limit"`
	Offset       int       `json:"offset"`
}

// LogSearchResult represents a log entry with additional metadata
type LogSearchResult struct {
	ID        int64     `json:"id"`
	ServiceID string    `json:"serviceId"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

// SearchLogs performs advanced search across service logs
func (db *Database) SearchLogs(criteria LogSearchCriteria) ([]LogSearchResult, int, error) {
	// Build base query
	baseQuery := `
		FROM service_logs 
		WHERE 1=1
	`

	countQuery := "SELECT COUNT(*) " + baseQuery
	selectQuery := `
		SELECT id, service_id, timestamp, level, message, created_at 
	` + baseQuery

	var args []interface{}
	var conditions []string

	// Add service ID filter
	if len(criteria.ServiceIDs) > 0 {
		placeholders := make([]string, len(criteria.ServiceIDs))
		for i, serviceID := range criteria.ServiceIDs {
			placeholders[i] = "?"
			args = append(args, serviceID)
		}
		serviceInClause := "service_id IN (" + strings.Join(placeholders, ", ") + ")"
		conditions = append(conditions, serviceInClause)
	}

	// Add level filter
	if len(criteria.Levels) > 0 {
		placeholders := make([]string, len(criteria.Levels))
		for i, level := range criteria.Levels {
			placeholders[i] = "?"
			args = append(args, level)
		}
		levelInClause := "level IN (" + strings.Join(placeholders, ", ") + ")"
		conditions = append(conditions, levelInClause)
	}

	// Add text search filter
	if criteria.SearchText != "" {
		conditions = append(conditions, "message LIKE ?")
		args = append(args, "%"+criteria.SearchText+"%")
	}

	// Add time range filters
	if !criteria.StartTime.IsZero() {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, criteria.StartTime)
	}

	if !criteria.EndTime.IsZero() {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, criteria.EndTime)
	}

	// Combine conditions
	if len(conditions) > 0 {
		whereClause := " AND " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
		baseQuery += whereClause
		countQuery = "SELECT COUNT(*) " + baseQuery
		selectQuery = `
			SELECT id, service_id, timestamp, level, message, created_at 
		` + baseQuery
	}

	// Get total count
	var totalCount int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	err := db.DB.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get log count: %w", err)
	}

	// Add ordering and pagination
	selectQuery += " ORDER BY timestamp DESC"

	if criteria.Limit > 0 {
		selectQuery += " LIMIT ?"
		args = append(args, criteria.Limit)
	}

	if criteria.Offset > 0 {
		selectQuery += " OFFSET ?"
		args = append(args, criteria.Offset)
	}

	// Execute search query
	rows, err := db.DB.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute log search query: %w", err)
	}
	defer rows.Close()

	var results []LogSearchResult
	for rows.Next() {
		var result LogSearchResult
		err := rows.Scan(
			&result.ID,
			&result.ServiceID,
			&result.Timestamp,
			&result.Level,
			&result.Message,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan log search result: %w", err)
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating log search results: %w", err)
	}

	return results, totalCount, nil
}

// GetRecentLogs retrieves the most recent logs for a service
func (db *Database) GetRecentLogs(serviceID string, limit int) ([]models.LogEntry, error) {
	query := `
		SELECT timestamp, level, message
		FROM service_logs
		WHERE service_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := db.DB.Query(query, serviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve recent logs for service %s: %w", serviceID, err)
	}
	defer rows.Close()

	var logs []models.LogEntry
	for rows.Next() {
		var logEntry models.LogEntry
		var timestamp time.Time

		err := rows.Scan(&timestamp, &logEntry.Level, &logEntry.Message)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}

		logEntry.Timestamp = timestamp.Format(time.RFC3339Nano)
		logs = append(logs, logEntry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recent logs: %w", err)
	}

	// Reverse order to get chronological order
	for i := len(logs)/2 - 1; i >= 0; i-- {
		opp := len(logs) - 1 - i
		logs[i], logs[opp] = logs[opp], logs[i]
	}

	return logs, nil
}

// CleanupOldLogs removes logs older than the retention period
func (db *Database) CleanupOldLogs() error {
	// Get retention settings
	var retentionDays int
	var autoCleanupEnabled bool

	query := `
		SELECT retention_days, auto_cleanup_enabled 
		FROM log_retention_settings 
		WHERE id = 1
	`

	err := db.DB.QueryRow(query).Scan(&retentionDays, &autoCleanupEnabled)
	if err != nil {
		return fmt.Errorf("failed to get retention settings: %w", err)
	}

	if !autoCleanupEnabled {
		return nil // Cleanup disabled
	}

	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Delete old logs
	deleteQuery := `
		DELETE FROM service_logs
		WHERE created_at < ?
	`

	result, err := db.DB.Exec(deleteQuery, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to cleanup old logs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("[INFO] Cleaned up %d old log entries (older than %d days)", rowsAffected, retentionDays)
	}

	return nil
}

// GetLogStatistics returns statistics about stored logs
func (db *Database) GetLogStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total log count
	var totalLogs int64
	err := db.DB.QueryRow("SELECT COUNT(*) FROM service_logs").Scan(&totalLogs)
	if err != nil {
		return nil, fmt.Errorf("failed to get total log count: %w", err)
	}
	stats["totalLogs"] = totalLogs

	// Logs per service
	serviceLogsQuery := `
		SELECT service_id, COUNT(*) as log_count
		FROM service_logs
		GROUP BY service_id
		ORDER BY log_count DESC
	`

	rows, err := db.DB.Query(serviceLogsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs per service: %w", err)
	}
	defer rows.Close()

	serviceStats := make(map[string]int64)
	for rows.Next() {
		var serviceID string
		var logCount int64
		if err := rows.Scan(&serviceID, &logCount); err != nil {
			return nil, fmt.Errorf("failed to scan service log stats: %w", err)
		}
		serviceStats[serviceID] = logCount
	}
	stats["logsByService"] = serviceStats

	// Logs by level
	levelLogsQuery := `
		SELECT level, COUNT(*) as log_count
		FROM service_logs
		GROUP BY level
		ORDER BY log_count DESC
	`

	rows, err = db.DB.Query(levelLogsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs by level: %w", err)
	}
	defer rows.Close()

	levelStats := make(map[string]int64)
	for rows.Next() {
		var level string
		var logCount int64
		if err := rows.Scan(&level, &logCount); err != nil {
			return nil, fmt.Errorf("failed to scan level log stats: %w", err)
		}
		levelStats[level] = logCount
	}
	stats["logsByLevel"] = levelStats

	// Date range
	var oldestLog, newestLog sql.NullTime
	dateRangeQuery := `
		SELECT MIN(timestamp) as oldest, MAX(timestamp) as newest
		FROM service_logs
	`

	err = db.DB.QueryRow(dateRangeQuery).Scan(&oldestLog, &newestLog)
	if err != nil {
		return nil, fmt.Errorf("failed to get log date range: %w", err)
	}

	if oldestLog.Valid && newestLog.Valid {
		stats["oldestLog"] = oldestLog.Time
		stats["newestLog"] = newestLog.Time
	}

	return stats, nil
}

// ClearServiceLogs deletes all logs for a specific service from the database
func (db *Database) ClearServiceLogs(serviceID string) error {
	query := `DELETE FROM service_logs WHERE service_id = ?`
	
	result, err := db.DB.Exec(query, serviceID)
	if err != nil {
		return fmt.Errorf("failed to clear logs for service %s: %w", serviceID, err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	log.Printf("[INFO] Cleared %d log entries for service %s", rowsAffected, serviceID)
	
	return nil
}

// ClearAllServiceLogs deletes logs for multiple services from the database
func (db *Database) ClearAllServiceLogs(serviceIDs []string) (map[string]error, error) {
	if len(serviceIDs) == 0 {
		// Clear all logs
		result, err := db.DB.Exec("DELETE FROM service_logs")
		if err != nil {
			return nil, fmt.Errorf("failed to clear all logs: %w", err)
		}
		
		rowsAffected, _ := result.RowsAffected()
		log.Printf("[INFO] Cleared all %d log entries from database", rowsAffected)
		return nil, nil
	}
	
	// Clear logs for specific services
	results := make(map[string]error)
	
	for _, serviceID := range serviceIDs {
		err := db.ClearServiceLogs(serviceID)
		results[serviceID] = err
	}
	
	return results, nil
}
