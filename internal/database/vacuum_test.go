package database

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

func pragmaInt(t *testing.T, db *sql.DB, pragma string) int64 {
	t.Helper()

	var value int64
	if err := db.QueryRow("PRAGMA " + pragma).Scan(&value); err != nil {
		t.Fatalf("PRAGMA %s: %v", pragma, err)
	}
	return value
}

// fillAndEmpty writes about 4 MB into a table and deletes it again, leaving
// the pages it used free.
func fillAndEmpty(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS filler (payload TEXT)"); err != nil {
		t.Fatal(err)
	}
	payload := strings.Repeat("x", 4000)
	for i := 0; i < 1000; i++ {
		if _, err := db.Exec("INSERT INTO filler (payload) VALUES (?)", payload); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec("DELETE FROM filler"); err != nil {
		t.Fatal(err)
	}
}

// An install from before this change has auto_vacuum off and a file full of
// free pages. Opening it must switch the mode and give the space back, with
// nothing for the user to run.
func TestOpeningAnOldDatabaseCompactsIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vertex.db")

	old, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
	if err != nil {
		t.Fatal(err)
	}
	fillAndEmpty(t, old)
	if mode := pragmaInt(t, old, "auto_vacuum"); mode != autoVacuumNone {
		t.Fatalf("setup: auto_vacuum = %d, want the old default %d", mode, autoVacuumNone)
	}
	freeBefore := pragmaInt(t, old, "freelist_count")
	if freeBefore < 900 {
		t.Fatalf("setup: freelist_count = %d, want the deleted rows' pages free", freeBefore)
	}
	old.Close()

	db, err := NewDatabaseWithPath(path)
	if err != nil {
		t.Fatalf("NewDatabaseWithPath: %v", err)
	}
	defer db.Close()

	if mode := pragmaInt(t, db.DB, "auto_vacuum"); mode != autoVacuumIncremental {
		t.Errorf("auto_vacuum = %d after opening, want %d (incremental)", mode, autoVacuumIncremental)
	}
	if free := pragmaInt(t, db.DB, "freelist_count"); free != 0 {
		t.Errorf("freelist_count = %d after opening, want 0 (was %d)", free, freeBefore)
	}
}

// After the switch, deleting logs followed by ReleaseFreePages shrinks the file
// by everything the delete freed, not by a single page.
func TestReleaseFreePagesReturnsAllOfThem(t *testing.T) {
	db, err := NewDatabaseWithPath(filepath.Join(t.TempDir(), "vertex.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	fillAndEmpty(t, db.DB)
	if free := pragmaInt(t, db.DB, "freelist_count"); free < 900 {
		t.Fatalf("setup: freelist_count = %d, want the deleted rows' pages free", free)
	}

	if err := db.ReleaseFreePages(); err != nil {
		t.Fatalf("ReleaseFreePages: %v", err)
	}
	if free := pragmaInt(t, db.DB, "freelist_count"); free != 0 {
		t.Errorf("freelist_count = %d after ReleaseFreePages, want 0", free)
	}
}
