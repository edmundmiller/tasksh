package timedb

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emiller/tasksh/internal/taskwarrior"
)

func TestNewTimeDB(t *testing.T) {
	// Create temp dir for test database
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create TimeDB: %v", err)
	}
	defer db.Close()

	// Check if database file was created
	dbPath := filepath.Join(tmpDir, ".local", "share", "tasksh", "timedb.sqlite3")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestRecordCompletion(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create TimeDB: %v", err)
	}
	defer db.Close()

	task := &taskwarrior.Task{
		UUID:        "test-uuid-123",
		Description: "Test task",
		Project:     "test-project",
		Priority:    "H",
		Status:      "pending",
	}

	// Record completion
	err = db.RecordCompletion(task, 2.0, 3.5)
	if err != nil {
		t.Errorf("Failed to record completion: %v", err)
	}

	// Try to record again (should update, not error)
	err = db.RecordCompletion(task, 2.5, 4.0)
	if err != nil {
		t.Errorf("Failed to update completion record: %v", err)
	}
}

func TestGetSimilarTasks(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create TimeDB: %v", err)
	}
	defer db.Close()

	// Add some test data
	tasks := []struct {
		task     *taskwarrior.Task
		estHours float64
		actHours float64
	}{
		{
			&taskwarrior.Task{UUID: "1", Description: "Write unit tests", Project: "tasksh", Priority: "H"},
			2.0, 3.0,
		},
		{
			&taskwarrior.Task{UUID: "2", Description: "Write integration tests", Project: "tasksh", Priority: "M"},
			3.0, 4.5,
		},
		{
			&taskwarrior.Task{UUID: "3", Description: "Fix bug in parser", Project: "other", Priority: "H"},
			1.0, 1.5,
		},
	}

	for _, tc := range tasks {
		if err := db.RecordCompletion(tc.task, tc.estHours, tc.actHours); err != nil {
			t.Fatalf("Failed to record test data: %v", err)
		}
	}

	// Test finding similar tasks
	searchTask := &taskwarrior.Task{
		UUID:        "search-1",
		Description: "Write more tests",
		Project:     "tasksh",
	}

	similar, err := db.GetSimilarTasks(searchTask, 10)
	if err != nil {
		t.Errorf("Failed to get similar tasks: %v", err)
	}

	// Should find at least the two tasksh project tasks
	if len(similar) < 2 {
		t.Errorf("Expected at least 2 similar tasks, got %d", len(similar))
	}

	// Verify project matching worked
	for _, entry := range similar {
		if entry.Project != "tasksh" && !strings.Contains(entry.Description, "tests") {
			t.Errorf("Unexpected task in results: %s", entry.Description)
		}
	}
}

func TestGetAverageTimeForProject(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create TimeDB: %v", err)
	}
	defer db.Close()

	// Add test data
	project := "test-project"
	tasks := []*taskwarrior.Task{
		{UUID: "1", Description: "Task 1", Project: project},
		{UUID: "2", Description: "Task 2", Project: project},
		{UUID: "3", Description: "Task 3", Project: project},
	}
	hours := []float64{2.0, 3.0, 4.0}

	for i, task := range tasks {
		if err := db.RecordCompletion(task, hours[i], hours[i]); err != nil {
			t.Fatalf("Failed to record test data: %v", err)
		}
	}

	// Test average calculation
	avgHours, count, err := db.GetAverageTimeForProject(project)
	if err != nil {
		t.Errorf("Failed to get average time: %v", err)
	}

	expectedAvg := 3.0 // (2+3+4)/3
	if avgHours != expectedAvg {
		t.Errorf("Expected average of %f, got %f", expectedAvg, avgHours)
	}

	if count != 3 {
		t.Errorf("Expected count of 3, got %d", count)
	}

	// Test with empty project
	avgHours, count, err = db.GetAverageTimeForProject("")
	if err != nil {
		t.Errorf("Failed with empty project: %v", err)
	}
	if avgHours != 0 || count != 0 {
		t.Error("Expected 0 results for empty project")
	}
}

func TestEstimateTimeForTask(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create TimeDB: %v", err)
	}
	defer db.Close()

	// Add historical data
	project := "web-app"
	historicalTasks := []struct {
		desc     string
		estHours float64
		actHours float64
		daysAgo  int
	}{
		{"Implement user authentication", 4.0, 6.0, 30},
		{"Implement password reset", 2.0, 3.0, 20},
		{"Implement user profile page", 3.0, 4.0, 10},
	}

	for i, ht := range historicalTasks {
		task := &taskwarrior.Task{
			UUID:        string(rune('a' + i)),
			Description: ht.desc,
			Project:     project,
		}
		// Adjust completed time to simulate age
		if err := db.RecordCompletion(task, ht.estHours, ht.actHours); err != nil {
			t.Fatalf("Failed to record historical data: %v", err)
		}
	}

	// Test estimation for similar task
	newTask := &taskwarrior.Task{
		UUID:        "new-1",
		Description: "Implement user settings",
		Project:     project,
	}

	estimate, reason, err := db.EstimateTimeForTask(newTask)
	if err != nil {
		t.Errorf("Failed to estimate time: %v", err)
	}

	// Should get a reasonable estimate based on similar tasks
	if estimate <= 0 {
		t.Error("Expected positive time estimate")
	}

	if reason == "" {
		t.Error("Expected estimation reason")
	}

	// Test task with no historical data
	lonelyTask := &taskwarrior.Task{
		UUID:        "lonely-1",
		Description: "Completely unique task",
		Project:     "new-project",
	}

	estimate, reason, err = db.EstimateTimeForTask(lonelyTask)
	if err != nil {
		t.Errorf("Failed to estimate for new task: %v", err)
	}

	if estimate != 0 {
		t.Error("Expected 0 estimate for task with no history")
	}
}

// parseTimeInput function was removed as it's not used in the refactored code