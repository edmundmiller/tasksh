package timedb

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
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
	db, err := sql.Open("sqlite3", dbPath)
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
		created_at DATETIME NOT NULL,
		UNIQUE(uuid)
	);
	
	CREATE INDEX IF NOT EXISTS idx_project ON time_entries(project);
	CREATE INDEX IF NOT EXISTS idx_priority ON time_entries(priority);
	CREATE INDEX IF NOT EXISTS idx_completed_at ON time_entries(completed_at);
	`
	
	_, err := tdb.db.Exec(schema)
	return err
}