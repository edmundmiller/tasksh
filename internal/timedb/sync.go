package timedb

import (
	"database/sql"
	"fmt"
	"time"
	
	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/timewarrior"
)

// SyncResult contains information about the sync operation
type SyncResult struct {
	NewEntries     int
	UpdatedEntries int
	TotalProcessed int
}

// SyncFromTimewarrior syncs time entries from timewarrior to the database
func (tdb *TimeDB) SyncFromTimewarrior(tw *timewarrior.Client, since time.Time) (*SyncResult, error) {
	// Get entries from timewarrior since the given date
	entries, err := tw.GetEntriesForDateRange(since, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get timewarrior entries: %w", err)
	}
	
	result := &SyncResult{TotalProcessed: len(entries)}
	
	for _, entry := range entries {
		// Skip entries without end time (still running)
		if entry.End.IsZero() {
			continue
		}
		
		// Extract task UUID and project from tags
		taskUUID := timewarrior.GetTaskUUID(entry)
		if taskUUID == "" {
			// Skip entries without task UUID
			continue
		}
		
		projectName := timewarrior.GetProjectName(entry)
		description := timewarrior.GetTaskDescription(entry)
		
		// Calculate duration in hours
		duration := entry.End.Sub(entry.Start.Time).Hours()
		
		// Check if entry already exists
		exists, err := tdb.entryExists(taskUUID, entry.Start.Time)
		if err != nil {
			return nil, err
		}
		
		if exists {
			// Update existing entry
			err = tdb.updateTimeEntry(taskUUID, entry.Start.Time, duration)
			if err != nil {
				return nil, err
			}
			result.UpdatedEntries++
		} else {
			// Create new entry
			err = tdb.createTimeEntry(taskUUID, description, projectName, "", duration, entry.End.Time)
			if err != nil {
				return nil, err
			}
			result.NewEntries++
		}
	}
	
	// Update last sync time
	if err := tdb.UpdateLastSyncTime(time.Now()); err != nil {
		// Log but don't fail the sync
		fmt.Printf("Warning: failed to update sync timestamp: %v\n", err)
	}
	
	return result, nil
}

// SyncTaskCompletion syncs a completed task with its timewarrior entries
func (tdb *TimeDB) SyncTaskCompletion(tw *timewarrior.Client, task *taskwarrior.Task) error {
	// Get all time entries for this task
	entries, err := tw.GetEntriesForTask(task.UUID)
	if err != nil {
		return fmt.Errorf("failed to get timewarrior entries for task: %w", err)
	}
	
	// Calculate total actual hours
	totalHours := timewarrior.CalculateTotalHours(entries)
	
	if totalHours > 0 {
		// Get estimated hours (from task estimate UDA or fallback)
		estimatedHours := 0.0 // This would come from task UDA or estimation
		
		// Record the completion with actual hours from timewarrior
		err = tdb.RecordCompletion(task, estimatedHours, totalHours)
		if err != nil {
			return fmt.Errorf("failed to record completion: %w", err)
		}
	}
	
	return nil
}

// entryExists checks if a time entry already exists
func (tdb *TimeDB) entryExists(taskUUID string, startTime time.Time) (bool, error) {
	query := `
	SELECT COUNT(*) FROM time_entries 
	WHERE uuid = ? AND completed_at BETWEEN ? AND ?
	`
	
	// Check within a 1-hour window to account for slight time differences
	startWindow := startTime.Add(-30 * time.Minute)
	endWindow := startTime.Add(30 * time.Minute)
	
	var count int
	err := tdb.db.QueryRow(query, taskUUID, startWindow, endWindow).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// updateTimeEntry updates the actual hours for an existing entry
func (tdb *TimeDB) updateTimeEntry(taskUUID string, startTime time.Time, actualHours float64) error {
	query := `
	UPDATE time_entries 
	SET actual_hours = ?
	WHERE uuid = ? AND completed_at BETWEEN ? AND ?
	`
	
	startWindow := startTime.Add(-30 * time.Minute)
	endWindow := startTime.Add(30 * time.Minute)
	
	_, err := tdb.db.Exec(query, actualHours, taskUUID, startWindow, endWindow)
	return err
}

// createTimeEntry creates a new time entry from timewarrior data
func (tdb *TimeDB) createTimeEntry(taskUUID, description, project, priority string, actualHours float64, completedAt time.Time) error {
	query := `
	INSERT INTO time_entries 
	(uuid, description, project, tags, priority, estimated_hours, actual_hours, completed_at, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	_, err := tdb.db.Exec(query,
		taskUUID,
		description,
		project,
		"", // tags
		priority,
		0.0, // estimated_hours - will be updated later
		actualHours,
		completedAt,
		now,
	)
	
	return err
}

// GetTimewarriorStats returns statistics about timewarrior-tracked time
func (tdb *TimeDB) GetTimewarriorStats(projectFilter string) (map[string]interface{}, error) {
	query := `
	SELECT 
		COUNT(DISTINCT uuid) as unique_tasks,
		COUNT(*) as total_entries,
		SUM(actual_hours) as total_hours,
		AVG(actual_hours) as avg_hours_per_entry,
		MIN(completed_at) as first_entry,
		MAX(completed_at) as last_entry
	FROM time_entries 
	WHERE actual_hours > 0
	`
	
	args := []interface{}{}
	if projectFilter != "" {
		query += " AND project = ?"
		args = append(args, projectFilter)
	}
	
	var stats map[string]interface{} = make(map[string]interface{})
	var uniqueTasks, totalEntries int
	var totalHours, avgHours float64
	var firstEntry, lastEntry time.Time
	
	err := tdb.db.QueryRow(query, args...).Scan(
		&uniqueTasks,
		&totalEntries, 
		&totalHours,
		&avgHours,
		&firstEntry,
		&lastEntry,
	)
	if err != nil {
		return stats, err
	}
	
	stats["unique_tasks"] = uniqueTasks
	stats["total_entries"] = totalEntries
	stats["total_hours"] = totalHours
	stats["avg_hours_per_entry"] = avgHours
	stats["first_entry"] = firstEntry
	stats["last_entry"] = lastEntry
	
	return stats, nil
}

// GetProjectsWithTime returns a list of projects that have time tracked
func (tdb *TimeDB) GetProjectsWithTime() ([]string, error) {
	query := `
	SELECT DISTINCT project 
	FROM time_entries 
	WHERE project != '' AND actual_hours > 0
	ORDER BY project
	`
	
	rows, err := tdb.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var projects []string
	for rows.Next() {
		var project string
		if err := rows.Scan(&project); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	
	return projects, rows.Err()
}

// GetLastSyncTime retrieves the last sync timestamp from metadata
func (tdb *TimeDB) GetLastSyncTime() (time.Time, error) {
	var lastSyncStr string
	err := tdb.db.QueryRow(`
		SELECT value FROM sync_metadata WHERE key = 'last_timewarrior_sync'
	`).Scan(&lastSyncStr)
	
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	
	return time.Parse(time.RFC3339, lastSyncStr)
}

// UpdateLastSyncTime updates the last sync timestamp in metadata
func (tdb *TimeDB) UpdateLastSyncTime(syncTime time.Time) error {
	_, err := tdb.db.Exec(`
		INSERT OR REPLACE INTO sync_metadata (key, value, updated_at)
		VALUES ('last_timewarrior_sync', ?, ?)
	`, syncTime.Format(time.RFC3339), time.Now())
	return err
}