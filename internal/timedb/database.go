package timedb

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// TimeDB handles time estimation database operations
type TimeDB struct {
	db *sql.DB
}

// New creates or opens the time estimation database
func New() (*TimeDB, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	
	dbDir := filepath.Join(homeDir, ".local", "share", "tasksh")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}
	
	dbPath := filepath.Join(dbDir, "timedb.sqlite3")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	timeDB := &TimeDB{db: db}
	if err := timeDB.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}
	
	return timeDB, nil
}

// Close closes the database connection
func (tdb *TimeDB) Close() error {
	return tdb.db.Close()
}

// initSchema creates the necessary tables
func (tdb *TimeDB) initSchema() error {
	// First, check if the table exists with the old schema
	var tableExists bool
	err := tdb.db.QueryRow(`SELECT EXISTS (SELECT name FROM sqlite_master WHERE type='table' AND name='time_entries')`).Scan(&tableExists)
	if err != nil {
		return err
	}
	
	if tableExists {
		// Check if we have the old schema with unique constraint
		var hasUniqueConstraint bool
		err = tdb.db.QueryRow(`
			SELECT COUNT(*) > 0 FROM sqlite_master 
			WHERE type='index' AND sql LIKE '%UNIQUE%' AND tbl_name='time_entries'
		`).Scan(&hasUniqueConstraint)
		if err != nil {
			return err
		}
		
		if hasUniqueConstraint {
			// Need to migrate - drop and recreate the table
			// First backup the data
			_, err = tdb.db.Exec(`ALTER TABLE time_entries RENAME TO time_entries_old`)
			if err != nil {
				return fmt.Errorf("failed to rename old table: %w", err)
			}
		}
	}
	
	// Create the new schema
	schema := `
	CREATE TABLE IF NOT EXISTS time_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid TEXT NOT NULL,
		description TEXT NOT NULL,
		project TEXT DEFAULT '',
		tags TEXT DEFAULT '',
		priority TEXT DEFAULT '',
		estimated_hours REAL DEFAULT 0,
		actual_hours REAL DEFAULT 0,
		completed_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_uuid ON time_entries(uuid);
	CREATE INDEX IF NOT EXISTS idx_project ON time_entries(project);
	CREATE INDEX IF NOT EXISTS idx_priority ON time_entries(priority);
	CREATE INDEX IF NOT EXISTS idx_completed_at ON time_entries(completed_at);
	
	CREATE TABLE IF NOT EXISTS sync_metadata (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`
	
	_, err = tdb.db.Exec(schema)
	if err != nil {
		return err
	}
	
	// If we have an old table, migrate the data
	var oldTableExists bool
	err = tdb.db.QueryRow(`SELECT EXISTS (SELECT name FROM sqlite_master WHERE type='table' AND name='time_entries_old')`).Scan(&oldTableExists)
	if err != nil {
		return err
	}
	
	if oldTableExists {
		// Copy data from old table, keeping only the most recent entry per UUID
		_, err = tdb.db.Exec(`
			INSERT INTO time_entries (uuid, description, project, tags, priority, estimated_hours, actual_hours, completed_at, created_at)
			SELECT uuid, description, project, tags, priority, estimated_hours, actual_hours, completed_at, created_at
			FROM time_entries_old
			WHERE id IN (
				SELECT MAX(id) FROM time_entries_old GROUP BY uuid
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to migrate data: %w", err)
		}
		
		// Drop the old table
		_, err = tdb.db.Exec(`DROP TABLE time_entries_old`)
		if err != nil {
			return fmt.Errorf("failed to drop old table: %w", err)
		}
	}
	
	return nil
}