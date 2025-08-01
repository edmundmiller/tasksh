package timewarrior

import (
	"testing"
	"time"
)

func TestCalculateTotalHours(t *testing.T) {
	entries := []Entry{
		{
			Start: TimeWarriorTime{time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)},
			End:   TimeWarriorTime{time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC)},
		},
		{
			Start: TimeWarriorTime{time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)},
			End:   TimeWarriorTime{time.Date(2024, 1, 1, 16, 15, 0, 0, time.UTC)},
		},
	}
	
	expected := 3.75 // 1.5 + 2.25 hours
	actual := CalculateTotalHours(entries)
	
	if actual != expected {
		t.Errorf("Expected %f hours, got %f", expected, actual)
	}
}

func TestGetTaskDescription(t *testing.T) {
	tests := []struct {
		name     string
		entry    Entry
		expected string
	}{
		{
			name: "with description tag",
			entry: Entry{
				Tags: []string{"task_123", "Fix_bug_in_parser", "project_tasksh"},
			},
			expected: "Fix bug in parser",
		},
		{
			name: "no description tag",
			entry: Entry{
				Tags: []string{"task_456", "project_tasksh"},
			},
			expected: "",
		},
		{
			name: "multiple non-special tags",
			entry: Entry{
				Tags: []string{"task_789", "Review_code", "High_priority", "project_work"},
			},
			expected: "Review code",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := GetTaskDescription(tt.entry)
			if actual != tt.expected {
				t.Errorf("Expected description %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestGetTaskUUID(t *testing.T) {
	tests := []struct {
		name     string
		entry    Entry
		expected string
	}{
		{
			name: "with task UUID",
			entry: Entry{
				Tags: []string{"task_abc123", "Some_task", "project_test"},
			},
			expected: "abc123",
		},
		{
			name: "no task UUID",
			entry: Entry{
				Tags: []string{"Some_task", "project_test"},
			},
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := GetTaskUUID(tt.entry)
			if actual != tt.expected {
				t.Errorf("Expected UUID %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestGetProjectName(t *testing.T) {
	tests := []struct {
		name     string
		entry    Entry
		expected string
	}{
		{
			name: "with project",
			entry: Entry{
				Tags: []string{"task_123", "Fix_bug", "project_tasksh"},
			},
			expected: "tasksh",
		},
		{
			name: "project with underscores",
			entry: Entry{
				Tags: []string{"task_456", "project_my_awesome_project"},
			},
			expected: "my awesome project",
		},
		{
			name: "no project",
			entry: Entry{
				Tags: []string{"task_789", "Some_task"},
			},
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := GetProjectName(tt.entry)
			if actual != tt.expected {
				t.Errorf("Expected project %q, got %q", tt.expected, actual)
			}
		})
	}
}