package review

import (
	"testing"
	"github.com/emiller/tasksh/internal/taskwarrior"
)

// TestNewImprovedModel tests model creation
func TestNewImprovedModel(t *testing.T) {
	m := NewImprovedModel()
	
	if m == nil {
		t.Fatal("NewModel returned nil")
	}
	
	// Check initial state
	if m.mode != ModeNormal {
		t.Errorf("Expected mode ModeNormal, got %v", m.mode)
	}
	
	if m.current != 0 {
		t.Errorf("Expected current to be 0, got %d", m.current)
	}
	
	if m.reviewed != 0 {
		t.Errorf("Expected reviewed to be 0, got %d", m.reviewed)
	}
	
	// Check subsystems are initialized
	if m.commands == nil {
		t.Error("Commands manager not initialized")
	}
	
	if m.renderer == nil {
		t.Error("Renderer not initialized")
	}
	
	if m.feedback == nil {
		t.Error("Feedback manager not initialized")
	}
}

// TestSetTasks tests task initialization
func TestSetTasks(t *testing.T) {
	m := NewImprovedModel()
	
	uuids := []string{"uuid1", "uuid2", "uuid3"}
	cache := map[string]*taskwarrior.Task{
		"uuid1": {UUID: "uuid1", Description: "Task 1"},
		"uuid2": {UUID: "uuid2", Description: "Task 2"},
		"uuid3": {UUID: "uuid3", Description: "Task 3"},
	}
	
	m.SetTasks(uuids, cache)
	
	if len(m.tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(m.tasks))
	}
	
	if len(m.taskCache) != 3 {
		t.Errorf("Expected 3 cached tasks, got %d", len(m.taskCache))
	}
	
	if m.current != 0 {
		t.Errorf("Expected current to be 0, got %d", m.current)
	}
}