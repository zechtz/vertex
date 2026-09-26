package database

import (
	"context"
	"fmt"
	"log"
)

// SQLite's auto_vacuum modes, as PRAGMA auto_vacuum reports them.
const (
	autoVacuumNone        = 0
	autoVacuumIncremental = 2
)

// enableIncrementalVacuum makes deleted rows give their space back to the disk.
//
// With auto_vacuum off, which is SQLite's default, a DELETE only marks pages
// free for reuse and the file never shrinks. Logs are written and pruned
// continuously, so databases grew without bound: one real install reached
// 3.6 GB of which 99.97% was free pages. In incremental mode the free pages
// are tracked so ReleaseFreePages can hand them back.
//
// Switching an existing database takes effect only after a VACUUM, which
// rewrites the file with just its live data. That happens once, on the first
// start after upgrading; a database already in incremental mode is left alone.
// VACUUM needs free disk roughly the size of the live data, not of the file.
func (db *Database) enableIncrementalVacuum() error {
	var mode int
	if err := db.QueryRow("PRAGMA auto_vacuum").Scan(&mode); err != nil {
		return fmt.Errorf("failed to read auto_vacuum mode: %w", err)
	}
	if mode == autoVacuumIncremental {
		return nil
	}

	sizeBefore, _ := db.fileSize()
	log.Printf("[INFO] Compacting the database and enabling incremental vacuum (one-time, %s)", formatBytes(sizeBefore))

	// The new mode is held by the connection that sets it, and applied only by
	// a VACUUM on that same connection. From the pool the two statements could
	// run on different connections, and the switch would silently not happen.
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a database connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "PRAGMA auto_vacuum = INCREMENTAL"); err != nil {
		return fmt.Errorf("failed to set auto_vacuum: %w", err)
	}
	if _, err := conn.ExecContext(ctx, "VACUUM"); err != nil {
		return fmt.Errorf("failed to vacuum: %w", err)
	}

	sizeAfter, _ := db.fileSize()
	log.Printf("[INFO] Database compacted: %s -> %s", formatBytes(sizeBefore), formatBytes(sizeAfter))
	return nil
}

// ReleaseFreePages returns the pages that deleted rows left free to the disk.
// It is cheap when there are none, so callers run it after every delete of
// logs rather than deciding whether it is worth it.
//
// The pragma frees one page per step of its statement, so it is read to the
// end: executed as a plain Exec, the driver would step it once and free a
// single page.
func (db *Database) ReleaseFreePages() error {
	rows, err := db.Query("PRAGMA incremental_vacuum")
	if err != nil {
		return fmt.Errorf("failed to release free pages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to release free pages: %w", err)
	}
	return nil
}

// fileSize is the database's size in bytes, as SQLite accounts for it.
func (db *Database) fileSize() (int64, error) {
	var pageCount, pageSize int64
	if err := db.QueryRow("PRAGMA page_count").Scan(&pageCount); err != nil {
		return 0, err
	}
	if err := db.QueryRow("PRAGMA page_size").Scan(&pageSize); err != nil {
		return 0, err
	}
	return pageCount * pageSize, nil
}

func formatBytes(bytes int64) string {
	const mb = 1024 * 1024
	if bytes >= mb {
		return fmt.Sprintf("%.1f MB", float64(bytes)/mb)
	}
	return fmt.Sprintf("%d KB", bytes/1024)
}
