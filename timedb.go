package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TimeEntry represents a task completion time record
type TimeEntry struct {
	UUID           string
	Description    string
	Project        string
	Tags           string
	Priority       string
	EstimatedHours float64
	ActualHours    float64
	CompletedAt    time.Time
	CreatedAt      time.Time
}

// TimeDB handles time estimation database operations
type TimeDB struct {
	db *sql.DB
}

// NewTimeDB creates or opens the time estimation database
func NewTimeDB() (*TimeDB, error) {
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

// RecordCompletion records a task completion with timing data
func (tdb *TimeDB) RecordCompletion(task *Task, estimatedHours, actualHours float64) error {
	query := `
	INSERT OR REPLACE INTO time_entries 
	(uuid, description, project, tags, priority, estimated_hours, actual_hours, completed_at, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	_, err := tdb.db.Exec(query,
		task.UUID,
		task.Description,
		task.Project,
		"", // We'll need to get tags from task if needed
		task.Priority,
		estimatedHours,
		actualHours,
		now,
		now,
	)
	
	return err
}

// GetSimilarTasks finds similar tasks based on description, project, and tags
func (tdb *TimeDB) GetSimilarTasks(task *Task, limit int) ([]TimeEntry, error) {
	// Simple similarity based on project match and description keywords
	query := `
	SELECT uuid, description, project, tags, priority, estimated_hours, actual_hours, completed_at, created_at
	FROM time_entries 
	WHERE (project = ? AND project != '') 
	   OR description LIKE ?
	ORDER BY completed_at DESC
	LIMIT ?
	`
	
	// Extract key words from description for matching
	descPattern := "%" + strings.Join(strings.Fields(task.Description), "%") + "%"
	
	rows, err := tdb.db.Query(query, task.Project, descPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var entries []TimeEntry
	for rows.Next() {
		var entry TimeEntry
		err := rows.Scan(
			&entry.UUID,
			&entry.Description,
			&entry.Project,
			&entry.Tags,
			&entry.Priority,
			&entry.EstimatedHours,
			&entry.ActualHours,
			&entry.CompletedAt,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	
	return entries, rows.Err()
}

// GetAverageTimeForProject calculates average completion time for a project
func (tdb *TimeDB) GetAverageTimeForProject(project string) (float64, int, error) {
	if project == "" {
		return 0, 0, nil
	}
	
	query := `
	SELECT AVG(actual_hours), COUNT(*)
	FROM time_entries 
	WHERE project = ? AND actual_hours > 0
	`
	
	var avgHours sql.NullFloat64
	var count int
	err := tdb.db.QueryRow(query, project).Scan(&avgHours, &count)
	if err != nil {
		return 0, 0, err
	}
	
	if avgHours.Valid {
		return avgHours.Float64, count, nil
	}
	return 0, count, nil
}

// GetEstimationAccuracy returns statistics about estimation accuracy
func (tdb *TimeDB) GetEstimationAccuracy() (map[string]interface{}, error) {
	query := `
	SELECT 
		COUNT(*) as total_tasks,
		AVG(actual_hours) as avg_actual,
		AVG(estimated_hours) as avg_estimated,
		AVG(ABS(actual_hours - estimated_hours) / actual_hours * 100) as avg_error_pct
	FROM time_entries 
	WHERE actual_hours > 0 AND estimated_hours > 0
	`
	
	var stats map[string]interface{} = make(map[string]interface{})
	var totalTasks int
	var avgActual, avgEstimated, avgErrorPct sql.NullFloat64
	
	err := tdb.db.QueryRow(query).Scan(&totalTasks, &avgActual, &avgEstimated, &avgErrorPct)
	if err != nil {
		return stats, err
	}
	
	stats["total_tasks"] = totalTasks
	if avgActual.Valid {
		stats["avg_actual_hours"] = avgActual.Float64
	} else {
		stats["avg_actual_hours"] = 0.0
	}
	if avgEstimated.Valid {
		stats["avg_estimated_hours"] = avgEstimated.Float64
	} else {
		stats["avg_estimated_hours"] = 0.0
	}
	if avgErrorPct.Valid {
		stats["avg_error_percent"] = avgErrorPct.Float64
	} else {
		stats["avg_error_percent"] = 0.0
	}
	
	return stats, nil
}

// EstimateTimeForTask provides a time estimate based on historical data
func (tdb *TimeDB) EstimateTimeForTask(task *Task) (float64, string, error) {
	// Try to find similar tasks
	similar, err := tdb.GetSimilarTasks(task, 10)
	if err != nil {
		return 0, "", err
	}
	
	if len(similar) == 0 {
		// Try project average
		if task.Project != "" {
			avgHours, count, err := tdb.GetAverageTimeForProject(task.Project)
			if err != nil {
				return 0, "", err
			}
			if count > 0 {
				return avgHours, fmt.Sprintf("Based on %d similar tasks in project '%s'", count, task.Project), nil
			}
		}
		return 0, "No historical data available", nil
	}
	
	// Calculate weighted average (more recent tasks weighted higher)
	totalWeight := 0.0
	weightedSum := 0.0
	
	for i, entry := range similar {
		// Weight decreases with age and position in results
		weight := 1.0 / float64(i+1)
		weightedSum += entry.ActualHours * weight
		totalWeight += weight
	}
	
	estimate := weightedSum / totalWeight
	return estimate, fmt.Sprintf("Based on %d similar tasks", len(similar)), nil
}