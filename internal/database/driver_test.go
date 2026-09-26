package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A burst of writes grows the WAL; once it has been checkpointed, the next
// write must bring it back under the limit instead of leaving it at its
// largest size for good.
func TestWALShrinksBackAfterABurstOfWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vertex.db")
	db, err := NewDatabaseWithPath(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec("CREATE TABLE filler (payload TEXT)"); err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	payload := strings.Repeat("x", 4000)
	for i := 0; i < 5000; i++ { // about 20 MB in one transaction
		if _, err := tx.Exec("INSERT INTO filler (payload) VALUES (?)", payload); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	walPath := path + "-wal"
	if size := fileSizeOf(t, walPath); size < 3*walSizeLimit {
		t.Fatalf("setup: WAL is %d bytes after the burst, want it well past the %d limit", size, walSizeLimit)
	}

	if _, err := db.Exec("PRAGMA wal_checkpoint(PASSIVE)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO filler (payload) VALUES ('after')"); err != nil {
		t.Fatal(err)
	}

	if size := fileSizeOf(t, walPath); size > walSizeLimit {
		t.Errorf("WAL is %d bytes after checkpoint and write, want at most %d", size, walSizeLimit)
	}
}

func fileSizeOf(t *testing.T, path string) int64 {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Size()
}
