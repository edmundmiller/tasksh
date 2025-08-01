package shared

import (
	"testing"
	
	"github.com/emiller/tasksh/internal/taskwarrior"
)

func TestRealisticTimeEstimates(t *testing.T) {
	analyzer := NewTaskAnalyzer(nil) // No DB, pure keyword estimation
	
	tests := []struct {
		description    string
		expectedHours  float64
	}{
		// Very quick tasks (6 minutes)
		{
			description:    "Brush teeth",
			expectedHours:  0.1,
		},
		{
			description:    "Take medication",
			expectedHours:  0.1,
		},
		{
			description:    "Feed pet",
			expectedHours:  0.1,
		},
		// Quick tasks (15 minutes)
		{
			description:    "Send email to John",
			expectedHours:  0.25,
		},
		{
			description:    "Quick call with team",
			expectedHours:  0.25,
		},
		{
			description:    "Daily standup",
			expectedHours:  0.25,
		},
		// Short tasks (30 minutes)
		{
			description:    "Review PR #123",
			expectedHours:  0.5,
		},
		{
			description:    "Update docs for API",
			expectedHours:  0.5,
		},
		{
			description:    "Workout at gym",
			expectedHours:  0.5,
		},
		{
			description:    "Grocery shopping",
			expectedHours:  0.5,
		},
		// Medium tasks (1.5 hours)
		{
			description:    "Implement feature X",
			expectedHours:  1.5,
		},
		{
			description:    "Team meeting about Q4 goals",
			expectedHours:  1.5,
		},
		// Long tasks (2.5 hours)
		{
			description:    "Research new framework options",
			expectedHours:  2.5,
		},
		{
			description:    "Plan project architecture",
			expectedHours:  2.5,
		},
		// Default cases
		{
			description:    "Fix bug", // Short description
			expectedHours:  0.25, // Very short description
		},
		{
			description:    "Work on the new functionality for processing user input and validation logic", // Long description without keywords
			expectedHours:  1.5, // Long description
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			task := &taskwarrior.Task{
				Description: tt.description,
			}
			
			// Test the fallback directly since we have no DB
			hours := analyzer.fallbackEstimate(task)
			
			// Allow small floating point differences
			if hours < tt.expectedHours-0.01 || hours > tt.expectedHours+0.01 {
				t.Errorf("Expected %.2f hours, got %.2f hours", tt.expectedHours, hours)
			}
		})
	}
}

func TestPriorityDoesNotAffectDuration(t *testing.T) {
	analyzer := NewTaskAnalyzer(nil)
	
	// Same task description should have same duration regardless of priority
	description := "Review pull request changes"
	expectedHours := 0.5 // Based on "review" keyword
	
	priorities := []string{"H", "M", "L", ""}
	
	for _, priority := range priorities {
		t.Run(priority+" priority", func(t *testing.T) {
			task := &taskwarrior.Task{
				Description: description,
				Priority:    priority,
			}
			
			hours := analyzer.fallbackEstimate(task)
			
			if hours < expectedHours-0.01 || hours > expectedHours+0.01 {
				t.Errorf("Priority %s: expected %.2f hours, got %.2f hours (priority should not affect duration)", 
					priority, expectedHours, hours)
			}
		})
	}
}