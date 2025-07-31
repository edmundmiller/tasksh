package planning

import (
	"strings"
	"testing"
	"time"

	"github.com/emiller/tasksh/internal/taskwarrior"
	_ "modernc.org/sqlite"
)

func TestNewPlanningSession(t *testing.T) {
	session, err := NewPlanningSession(HorizonTomorrow)
	if err != nil {
		t.Fatalf("Failed to create planning session: %v", err)
	}
	defer session.Close()

	if session.Horizon != HorizonTomorrow {
		t.Errorf("Expected horizon %v, got %v", HorizonTomorrow, session.Horizon)
	}

	if session.DailyCapacity != 8.0 {
		t.Errorf("Expected daily capacity 8.0, got %v", session.DailyCapacity)
	}

	// Check that date is set to tomorrow
	expectedDate := time.Now().AddDate(0, 0, 1)
	if session.Date.Day() != expectedDate.Day() {
		t.Errorf("Expected date to be tomorrow, got %v", session.Date)
	}
}

func TestCalculateUrgency(t *testing.T) {
	session, err := NewPlanningSession(HorizonTomorrow)
	if err != nil {
		t.Fatalf("Failed to create planning session: %v", err)
	}
	defer session.Close()

	tests := []struct {
		name     string
		task     *taskwarrior.Task
		expected float64
	}{
		{
			name: "high priority task",
			task: &taskwarrior.Task{
				Priority: "H",
				Project:  "test",
			},
			expected: 7.0, // 6.0 + 1.0 for project
		},
		{
			name: "medium priority task",
			task: &taskwarrior.Task{
				Priority: "M",
			},
			expected: 1.8,
		},
		{
			name: "no priority task with project",
			task: &taskwarrior.Task{
				Project: "test",
			},
			expected: 1.0,
		},
		{
			name: "basic task",
			task: &taskwarrior.Task{},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urgency := session.calculateUrgency(tt.task)
			if urgency != tt.expected {
				t.Errorf("Expected urgency %v, got %v", tt.expected, urgency)
			}
		})
	}
}

func TestMoveTask(t *testing.T) {
	session, err := NewPlanningSession(HorizonTomorrow)
	if err != nil {
		t.Fatalf("Failed to create planning session: %v", err)
	}
	defer session.Close()

	// Create test tasks
	session.Tasks = []PlannedTask{
		{Task: &taskwarrior.Task{Description: "Task 1"}},
		{Task: &taskwarrior.Task{Description: "Task 2"}},
		{Task: &taskwarrior.Task{Description: "Task 3"}},
	}

	// Move task from position 0 to position 2
	err = session.MoveTask(0, 2)
	if err != nil {
		t.Fatalf("Failed to move task: %v", err)
	}

	// After moving Task 1 from position 0 to position 2:
	// Original: [Task 1, Task 2, Task 3]
	// After:    [Task 2, Task 3, Task 1]
	if session.Tasks[0].Description != "Task 2" {
		t.Errorf("Expected Task 2 at position 0, got %s", session.Tasks[0].Description)
	}
	if session.Tasks[1].Description != "Task 3" {
		t.Errorf("Expected Task 3 at position 1, got %s", session.Tasks[1].Description)
	}
	if session.Tasks[2].Description != "Task 1" {
		t.Errorf("Expected Task 1 at position 2, got %s", session.Tasks[2].Description)
	}
}

func TestRemoveTask(t *testing.T) {
	session, err := NewPlanningSession(HorizonTomorrow)
	if err != nil {
		t.Fatalf("Failed to create planning session: %v", err)
	}
	defer session.Close()

	// Create test tasks with estimates
	session.Tasks = []PlannedTask{
		{Task: &taskwarrior.Task{Description: "Task 1"}, EstimatedHours: 2.0},
		{Task: &taskwarrior.Task{Description: "Task 2"}, EstimatedHours: 3.0},
	}
	session.DailyCapacity = 8.0

	// Initial total should be 5.0
	session.calculateTotals()
	if session.TotalHours != 5.0 {
		t.Errorf("Expected total hours 5.0, got %v", session.TotalHours)
	}

	// Remove first task
	err = session.RemoveTask(0)
	if err != nil {
		t.Fatalf("Failed to remove task: %v", err)
	}

	if len(session.Tasks) != 1 {
		t.Errorf("Expected 1 task remaining, got %d", len(session.Tasks))
	}
	if session.Tasks[0].Description != "Task 2" {
		t.Errorf("Expected Task 2 to remain, got %s", session.Tasks[0].Description)
	}
	if session.TotalHours != 3.0 {
		t.Errorf("Expected total hours 3.0, got %v", session.TotalHours)
	}
}

func TestCalculateTotals(t *testing.T) {
	session, err := NewPlanningSession(HorizonTomorrow)
	if err != nil {
		t.Fatalf("Failed to create planning session: %v", err)
	}
	defer session.Close()

	tests := []struct {
		name            string
		tasks           []PlannedTask
		dailyCapacity   float64
		expectedTotal   float64
		expectedWarning WarningLevel
	}{
		{
			name: "under capacity",
			tasks: []PlannedTask{
				{EstimatedHours: 2.0},
				{EstimatedHours: 3.0},
			},
			dailyCapacity:   8.0,
			expectedTotal:   5.0,
			expectedWarning: WarningNone,
		},
		{
			name: "near capacity",
			tasks: []PlannedTask{
				{EstimatedHours: 3.0},
				{EstimatedHours: 2.5},
			},
			dailyCapacity:   8.0,
			expectedTotal:   5.5,
			expectedWarning: WarningCaution,
		},
		{
			name: "over capacity",
			tasks: []PlannedTask{
				{EstimatedHours: 4.0},
				{EstimatedHours: 5.0},
			},
			dailyCapacity:   8.0,
			expectedTotal:   9.0,
			expectedWarning: WarningOverload,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session.Tasks = tt.tasks
			session.DailyCapacity = tt.dailyCapacity
			session.FocusCapacity = 6.0 // Use the actual focus capacity used for warnings
			session.calculateTotals()

			if session.TotalHours != tt.expectedTotal {
				t.Errorf("Expected total hours %v, got %v", tt.expectedTotal, session.TotalHours)
			}
			if session.WarningLevel != tt.expectedWarning {
				t.Errorf("Expected warning level %v, got %v", tt.expectedWarning, session.WarningLevel)
			}
		})
	}
}

func TestGetCapacityStatus(t *testing.T) {
	session, err := NewPlanningSession(HorizonTomorrow)
	if err != nil {
		t.Fatalf("Failed to create planning session: %v", err)
	}
	defer session.Close()

	tests := []struct {
		name          string
		totalHours    float64
		focusCapacity float64
		warningLevel  WarningLevel
		expectContains string
	}{
		{
			name:          "normal capacity",
			totalHours:    5.0,
			focusCapacity: 6.0,
			warningLevel:  WarningNone,
			expectContains: "5.0h/6.0h (83% focus capacity",
		},
		{
			name:          "near capacity",
			totalHours:    5.5,
			focusCapacity: 6.0,
			warningLevel:  WarningCaution,
			expectContains: "⚠️ Near capacity: 5.5h/6.0h (91%)",
		},
		{
			name:          "over capacity",
			totalHours:    7.0,
			focusCapacity: 6.0,
			warningLevel:  WarningOverload,
			expectContains: "⚠️ Overloaded by 1.0h (116% of focus capacity)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session.TotalHours = tt.totalHours
			session.FocusCapacity = tt.focusCapacity
			session.WarningLevel = tt.warningLevel
			// Set up some dummy tasks for the task count
			session.CriticalTasks = []PlannedTask{}
			session.ImportantTasks = []PlannedTask{}
			session.FlexibleTasks = []PlannedTask{}

			status := session.GetCapacityStatus()
			if !strings.Contains(status, tt.expectContains) {
				t.Errorf("Expected status to contain '%s', got '%s'", tt.expectContains, status)
			}
		})
	}
}

func TestGetProjectedCompletionTimes(t *testing.T) {
	session, err := NewPlanningSession(HorizonTomorrow)
	if err != nil {
		t.Fatalf("Failed to create planning session: %v", err)
	}
	defer session.Close()

	// Create test tasks
	session.Tasks = []PlannedTask{
		{EstimatedHours: 1.0}, // 1 hour
		{EstimatedHours: 2.0}, // 2 hours
		{EstimatedHours: 0.5}, // 30 minutes
	}

	startTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC) // 9:00 AM
	completionTimes := session.GetProjectedCompletionTimes(startTime)

	expectedTimes := []time.Time{
		time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), // 10:00 AM
		time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), // 12:00 PM  
		time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC), // 12:30 PM
	}

	if len(completionTimes) != len(expectedTimes) {
		t.Errorf("Expected %d completion times, got %d", len(expectedTimes), len(completionTimes))
	}

	for i, expected := range expectedTimes {
		if !completionTimes[i].Equal(expected) {
			t.Errorf("Expected completion time %d to be %v, got %v", i, expected, completionTimes[i])
		}
	}
}