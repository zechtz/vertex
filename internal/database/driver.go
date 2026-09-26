package database

import (
	"database/sql"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

// driverName is the SQLite driver Vertex opens its database with: the stock
// driver plus per-connection settings the DSN cannot carry.
const driverName = "sqlite3_vertex"

// walSizeLimit caps the WAL file's size once it has been checkpointed. SQLite
// reuses the WAL but never shrinks it on its own, so without a limit it stays
// as large as the biggest burst of writes ever made it; one install's sat at
// 27 MB beside a 3 MB database. 4 MB is just above the size at which SQLite
// checkpoints automatically, so everyday writes never have to truncate it.
const walSizeLimit = 4 * 1024 * 1024

func init() {
	sql.Register(driverName, &sqlite3.SQLiteDriver{
		// journal_size_limit is a per-connection setting, and any connection in
		// the pool may be the one that restarts the WAL, so each gets it.
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			_, err := conn.Exec(fmt.Sprintf("PRAGMA journal_size_limit = %d", walSizeLimit), nil)
			return err
		},
	})
}
