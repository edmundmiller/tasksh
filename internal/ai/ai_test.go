package ai

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/timedb"
)

func TestBuildAnalysisPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	db, err := timedb.New()
	if err != nil {
		t.Fatalf("Failed to create TimeDB: %v", err)
	}
	defer db.Close()

	analyzer := NewAnalyzer(db)

	task := &taskwarrior.Task{
		UUID:        "test-123",
		Description: "Implement new feature",
		Project:     "myproject",
		Priority:    "H",
		Due:         "2024-12-31",
		Status:      "pending",
	}

	similar := []timedb.TimeEntry{
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

	if !strings.Contains(prompt, "Analysis Request") {
		t.Error("Prompt missing analysis request section")
	}

	if !strings.Contains(prompt, "json") {
		t.Error("Prompt missing JSON format request")
	}
}

func TestParseAnalysisResponse(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	db, err := timedb.New()
	if err != nil {
		t.Fatalf("Failed to create TimeDB: %v", err)
	}
	defer db.Close()

	analyzer := NewAnalyzer(db)

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



func TestCheckOpenAIAvailable(t *testing.T) {
	// This test will pass/fail based on whether OpenAI API key is available
	err := CheckOpenAIAvailable()
	if err != nil {
		t.Logf("OpenAI API not available (expected if no API key): %v", err)
	} else {
		t.Log("OpenAI API is available")
	}
}

func TestGetValueOrEmpty(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "none"},
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

// formatHours function was removed as it's not used in the refactored code