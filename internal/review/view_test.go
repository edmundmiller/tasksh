package review

import (
	"testing"
	"github.com/emiller/tasksh/internal/taskwarrior"
)

// TestViewRenderer tests the view renderer
func TestViewRenderer(t *testing.T) {
	r := NewViewRenderer()
	
	if r == nil {
		t.Fatal("NewViewRenderer returned nil")
	}
	
	if r.styles == nil {
		t.Error("Styles not initialized")
	}
}

// TestRenderTask tests task rendering
func TestRenderTask(t *testing.T) {
	r := NewViewRenderer()
	
	// Test nil task
	result := r.RenderTask(nil)
	if result == "" {
		t.Error("Expected non-empty result for nil task")
	}
	
	// Test normal task
	task := &taskwarrior.Task{
		UUID:        "test-uuid-1234",
		Description: "Test task description",
		Project:     "TestProject",
		Priority:    "H",
		Status:      "pending",
		// Tags:        []string{"tag1", "tag2"},
	}
	
	result = r.RenderTask(task)
	
	// Check that key elements are present
	if !contains(result, "Test task description") {
		t.Error("Task description not found in rendered output")
	}
	
	if !contains(result, "TestProject") {
		t.Error("Project not found in rendered output")
	}
	
	// if !contains(result, "tag1") {
	// 	t.Error("Tags not found in rendered output")
	// }
}

// TestTruncate tests the truncate function
func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"short", 10, "short"},
		{"very long string", 10, "very lo..."},
		{"exact", 5, "exact"},
		{"", 5, ""},
	}
	
	for _, tt := range tests {
		result := truncate(tt.input, tt.length)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", 
				tt.input, tt.length, result, tt.expected)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		(s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}