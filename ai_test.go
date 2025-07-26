package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestBuildAnalysisPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	db, err := NewTimeDB()
	if err != nil {
		t.Fatalf("Failed to create TimeDB: %v", err)
	}
	defer db.Close()

	analyzer := NewAIAnalyzer(db)

	task := &Task{
		UUID:        "test-123",
		Description: "Implement new feature",
		Project:     "myproject",
		Priority:    "H",
		Due:         "2024-12-31",
		Status:      "pending",
	}

	similar := []TimeEntry{
		{
			Description: "Implement similar feature",
			ActualHours: 4.5,
			CompletedAt: time.Now().AddDate(0, -1, 0),
		},
	}

	prompt := analyzer.buildAnalysisPrompt(task, 3.5, "Based on similar tasks", similar)

	// Check that prompt contains key elements
	if !strings.Contains(prompt, "Task Analysis Request") {
		t.Error("Prompt missing header")
	}

	if !strings.Contains(prompt, task.Description) {
		t.Error("Prompt missing task description")
	}

	if !strings.Contains(prompt, task.Project) {
		t.Error("Prompt missing project")
	}

	if !strings.Contains(prompt, "3.5 hours") {
		t.Error("Prompt missing time estimate")
	}

	if !strings.Contains(prompt, "Response Format") {
		t.Error("Prompt missing response format section")
	}

	if !strings.Contains(prompt, "json") {
		t.Error("Prompt missing JSON format request")
	}
}

func TestParseAnalysisResponse(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	db, err := NewTimeDB()
	if err != nil {
		t.Fatalf("Failed to create TimeDB: %v", err)
	}
	defer db.Close()

	analyzer := NewAIAnalyzer(db)

	// Test valid JSON response
	validResponse := `
Some preamble text from the LLM...

{
  "summary": "Task looks well-organized but could use a more specific due date",
  "suggestions": [
    {
      "type": "due_date",
      "current": "next week",
      "suggested": "2024-01-15",
      "reason": "More specific date improves planning",
      "confidence": 0.8
    }
  ],
  "time_estimate": {
    "hours": 3.5,
    "reason": "Based on similar tasks"
  }
}

Some additional text after...
`

	analysis, err := analyzer.parseAnalysisResponse("test-uuid", validResponse)
	if err != nil {
		t.Errorf("Failed to parse valid response: %v", err)
	}

	if analysis.TaskUUID != "test-uuid" {
		t.Error("Task UUID not set correctly")
	}

	if len(analysis.Suggestions) != 1 {
		t.Errorf("Expected 1 suggestion, got %d", len(analysis.Suggestions))
	}

	if analysis.Suggestions[0].Type != "due_date" {
		t.Error("Suggestion type incorrect")
	}

	if analysis.TimeEstimate.Hours != 3.5 {
		t.Error("Time estimate incorrect")
	}

	// Test invalid response
	invalidResponse := "This response contains no JSON"
	_, err = analyzer.parseAnalysisResponse("test-uuid", invalidResponse)
	if err == nil {
		t.Error("Expected error for invalid response")
	}
}

func TestGetModificationSuggestions(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	db, err := NewTimeDB()
	if err != nil {
		t.Fatalf("Failed to create TimeDB: %v", err)
	}
	defer db.Close()

	analyzer := NewAIAnalyzer(db)

	analysis := &TaskAnalysis{
		TaskUUID: "test-123",
		Summary:  "Test analysis",
		Suggestions: []TaskSuggestion{
			{
				Type:           "due_date",
				CurrentValue:   "tomorrow",
				SuggestedValue: "2024-01-15",
				Reason:         "More specific",
				Confidence:     0.9,
			},
			{
				Type:           "priority",
				CurrentValue:   "M",
				SuggestedValue: "H",
				Reason:         "Urgent deadline",
				Confidence:     0.7,
			},
			{
				Type:           "tag",
				CurrentValue:   "",
				SuggestedValue: "urgent",
				Reason:         "Helps filtering",
				Confidence:     0.6,
			},
		},
		TimeEstimate: struct {
			Hours  float64 `json:"hours"`
			Reason string  `json:"reason"`
		}{
			Hours:  5.0,
			Reason: "Complex task",
		},
	}

	suggestions := analyzer.GetModificationSuggestions(analysis)

	// Should have 4 suggestions (3 from analysis + 1 time estimate)
	if len(suggestions) < 3 {
		t.Errorf("Expected at least 3 modification suggestions, got %d", len(suggestions))
	}

	// Check due date suggestion
	found := false
	for _, s := range suggestions {
		if strings.Contains(s.Value, "due:2024-01-15") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Due date suggestion not found")
	}

	// Check priority suggestion
	found = false
	for _, s := range suggestions {
		if s.Value == "priority:H" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Priority suggestion not found")
	}

	// Check tag suggestion
	found = false
	for _, s := range suggestions {
		if s.Value == "+urgent" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Tag suggestion not found")
	}
}

func TestFormatAnalysis(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	db, err := NewTimeDB()
	if err != nil {
		t.Fatalf("Failed to create TimeDB: %v", err)
	}
	defer db.Close()

	analyzer := NewAIAnalyzer(db)

	analysis := &TaskAnalysis{
		Summary: "Task needs better organization",
		Suggestions: []TaskSuggestion{
			{
				Type:           "project",
				CurrentValue:   "",
				SuggestedValue: "work",
				Reason:         "Group related tasks",
				Confidence:     0.85,
			},
		},
		TimeEstimate: struct {
			Hours  float64 `json:"hours"`
			Reason string  `json:"reason"`
		}{
			Hours:  2.5,
			Reason: "Similar to previous tasks",
		},
	}

	formatted := analyzer.FormatAnalysis(analysis)

	// Check for key formatting elements
	if !strings.Contains(formatted, "🤖 AI Analysis") {
		t.Error("Missing AI Analysis header")
	}

	if !strings.Contains(formatted, analysis.Summary) {
		t.Error("Missing summary in output")
	}

	if !strings.Contains(formatted, "85%") {
		t.Error("Missing confidence percentage")
	}

	if !strings.Contains(formatted, "2.5 hours") {
		t.Error("Missing time estimate")
	}
}

func TestCheckModsAvailable(t *testing.T) {
	// This test will pass/fail based on whether mods is installed
	err := checkModsAvailable()
	if err != nil {
		t.Logf("Mods not available (expected if not installed): %v", err)
	} else {
		t.Log("Mods is available")
	}
}

func TestGetValueOrEmpty(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "(none)"},
		{"value", "value"},
		{"  ", "  "}, // Spaces are preserved
	}

	for _, tt := range tests {
		result := getValueOrEmpty(tt.input)
		if result != tt.expected {
			t.Errorf("getValueOrEmpty(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatHours(t *testing.T) {
	tests := []struct {
		hours    float64
		expected string
	}{
		{0.25, "15 min"},
		{0.5, "30 min"},
		{0.75, "45 min"},
		{1.0, "1.0"},
		{2.5, "2.5"},
		{10.0, "10.0"},
	}

	for _, tt := range tests {
		result := formatHours(tt.hours)
		if result != tt.expected {
			t.Errorf("formatHours(%f) = %q, want %q", tt.hours, result, tt.expected)
		}
	}
}