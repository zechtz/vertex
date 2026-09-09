package services

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/zechtz/vertex/internal/database"
	"github.com/zechtz/vertex/internal/models"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()

	db, err := database.NewDatabaseWithPath(filepath.Join(t.TempDir(), "vertex.db"))
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return &Manager{
		db:             db,
		services:       make(map[string]*models.Service),
		configurations: make(map[string]*models.Configuration),
		clients:        make(map[*websocket.Conn]*sync.Mutex),
	}
}

// The database must open in WAL mode, otherwise every write blocks every read.
func TestDatabaseOpensInWALMode(t *testing.T) {
	sm := newTestManager(t)

	var journalMode string
	if err := sm.db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("failed to read journal_mode: %v", err)
	}

	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want %q", journalMode, "wal")
	}
}

func seedLogs(t *testing.T, sm *Manager, serviceID string, count int, age time.Duration) {
	t.Helper()

	entries := make([]models.LogEntry, 0, count)
	for i := range count {
		entries = append(entries, models.LogEntry{
			Timestamp: time.Now().Format(time.RFC3339Nano),
			Level:     "INFO",
			Message:   fmt.Sprintf("line %d", i),
		})
	}

	if err := sm.db.StoreLogEntries(serviceID, entries); err != nil {
		t.Fatalf("failed to seed logs for %s: %v", serviceID, err)
	}

	if age > 0 {
		if _, err := sm.db.Exec(
			"UPDATE service_logs SET created_at = ? WHERE service_id = ?",
			time.Now().Add(-age), serviceID); err != nil {
			t.Fatalf("failed to age logs for %s: %v", serviceID, err)
		}
	}
}

func countLogs(t *testing.T, sm *Manager, serviceID string) int {
	t.Helper()

	var count int
	if err := sm.db.QueryRow(
		"SELECT COUNT(*) FROM service_logs WHERE service_id = ?", serviceID).Scan(&count); err != nil {
		t.Fatalf("failed to count logs for %s: %v", serviceID, err)
	}
	return count
}

func TestCleanupOldLogs(t *testing.T) {
	sm := newTestManager(t)

	// "keep" exceeds the per-service cap and spans more than one delete batch,
	// "under" is below the cap, and "stale" is entirely older than the cutoff.
	seedLogs(t, sm, "keep", logCleanupBatchSize*2+250, 0)
	seedLogs(t, sm, "under", 300, 0)
	seedLogs(t, sm, "stale", 400, 30*24*time.Hour)

	for _, id := range []string{"keep", "under", "stale"} {
		sm.services[id] = &models.Service{ID: id, Name: id}
	}

	if err := sm.CleanupOldLogs(7, 1000); err != nil {
		t.Fatalf("CleanupOldLogs returned error: %v", err)
	}

	if got := countLogs(t, sm, "keep"); got != 1000 {
		t.Errorf("keep: got %d logs, want 1000", got)
	}
	if got := countLogs(t, sm, "under"); got != 300 {
		t.Errorf("under: got %d logs, want 300 (below cap, nothing to trim)", got)
	}
	if got := countLogs(t, sm, "stale"); got != 0 {
		t.Errorf("stale: got %d logs, want 0 (all past retention)", got)
	}

	// The most recent entries are the ones that survive.
	var newest string
	if err := sm.db.QueryRow(
		"SELECT message FROM service_logs WHERE service_id = 'keep' ORDER BY id DESC LIMIT 1").Scan(&newest); err != nil {
		t.Fatalf("failed to read newest surviving log: %v", err)
	}
	if want := fmt.Sprintf("line %d", logCleanupBatchSize*2+249); newest != want {
		t.Errorf("newest surviving log = %q, want %q", newest, want)
	}
}

// A service lock held by slow work (a health check mid-HTTP-request) must not
// block reads of the manager. GetServices used to hold the manager lock while
// waiting on a service lock; because sync.RWMutex is writer-preferring, a single
// queued writer then parked every later reader behind that one slow service,
// which is what made configuration fetches hang.
func TestSlowServiceDoesNotBlockConfigurationReads(t *testing.T) {
	sm := newTestManager(t)

	slow := &models.Service{ID: "slow", Name: "slow"}
	sm.services[slow.ID] = slow
	sm.configurations["cfg"] = &models.Configuration{ID: "cfg", Name: "default"}

	// Stand in for a health check holding the service lock across slow I/O.
	slow.Mutex.Lock()
	defer slow.Mutex.Unlock()

	reading := make(chan struct{})
	go func() {
		close(reading)
		sm.GetServices() // blocks on the service lock until this test returns
	}()
	<-reading

	// Queue a writer on the manager lock. Under the old ordering this is what
	// converted one stuck service into a manager-wide stall.
	writing := make(chan struct{})
	go func() {
		defer close(writing)
		if err := sm.SaveConfiguration(&models.Configuration{ID: "written", Name: "written"}); err != nil {
			t.Errorf("SaveConfiguration failed: %v", err)
		}
	}()

	// Let both goroutines reach their locks.
	time.Sleep(100 * time.Millisecond)

	read := make(chan int, 1)
	go func() { read <- len(sm.GetConfigurations()) }()

	select {
	case count := <-read:
		if count < 1 {
			t.Errorf("GetConfigurations returned %d configurations, want at least 1", count)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("GetConfigurations blocked behind a held service lock")
	}

	select {
	case <-writing:
	case <-time.After(5 * time.Second):
		t.Fatal("SaveConfiguration blocked behind a held service lock")
	}
}
