package review

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestSessionManager tests session manager creation
func TestSessionManager(t *testing.T) {
	sm, err := NewSessionManager()
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	
	if sm == nil {
		t.Fatal("NewSessionManager returned nil")
	}
	
	// Clean up any existing session
	sm.ClearSession()
}

// TestSessionSaveLoad tests saving and loading sessions
func TestSessionSaveLoad(t *testing.T) {
	sm, err := NewSessionManager()
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	
	// Create a test session
	session := CreateSession([]string{"uuid1", "uuid2"}, "status:pending")
	session.Current = 1
	session.Reviewed = 1
	
	// Save session
	if err := sm.SaveSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}
	
	// Load session
	loaded, err := sm.LoadSession()
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}
	
	if loaded == nil {
		t.Fatal("LoadSession returned nil")
	}
	
	// Verify loaded data
	if len(loaded.Tasks) != 2 {
		t.Errorf("Expected 2 task UUIDs, got %d", len(loaded.Tasks))
	}
	
	if loaded.Current != 1 {
		t.Errorf("Expected Current 1, got %d", loaded.Current)
	}
	
	if loaded.Reviewed != 1 {
		t.Errorf("Expected Reviewed 1, got %d", loaded.Reviewed)
	}
	
	if loaded.Filter != "status:pending" {
		t.Errorf("Expected filter 'status:pending', got %q", loaded.Filter)
	}
	
	// Clean up
	sm.ClearSession()
}

// TestSessionExpiry tests that old sessions are not loaded
func TestSessionExpiry(t *testing.T) {
	sm, err := NewSessionManager()
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	
	// Create an old session
	session := CreateSession([]string{"uuid1"}, "")
	oldTime := time.Now().Add(-25 * time.Hour) // More than 24 hours old
	session.LastActive = oldTime
	
	// Manually save the session to preserve the old LastActive time
	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("Failed to marshal session: %v", err)
	}
	if err := os.WriteFile(sm.sessionPath, data, 0644); err != nil {
		t.Fatalf("Failed to write session: %v", err)
	}
	
	// Try to load - should return nil
	loaded, err := sm.LoadSession()
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}
	
	if loaded != nil {
		t.Error("Expected nil for expired session")
	}
	
	// Session file should be cleaned up
	if _, err := os.Stat(sm.sessionPath); !os.IsNotExist(err) {
		t.Error("Expired session file should have been removed")
	}
}

// TestCreateSession tests session creation
func TestCreateSession(t *testing.T) {
	uuids := []string{"uuid1", "uuid2", "uuid3"}
	filter := "project:test"
	
	session := CreateSession(uuids, filter)
	
	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}
	
	if len(session.Tasks) != 3 {
		t.Errorf("Expected 3 task UUIDs, got %d", len(session.Tasks))
	}
	
	if session.Filter != filter {
		t.Errorf("Expected filter %q, got %q", filter, session.Filter)
	}
	
	if session.Current != 0 {
		t.Errorf("Expected Current 0, got %d", session.Current)
	}
	
	if session.Reviewed != 0 {
		t.Errorf("Expected Reviewed 0, got %d", session.Reviewed)
	}
}

// TestResumePrompt tests the resume prompt generation
func TestResumePrompt(t *testing.T) {
	session := &Session{
		StartTime: time.Now().Add(-30 * time.Minute),
		Tasks:     make([]string, 10),
		Reviewed:  3,
		Filter:    "status:pending",
	}
	
	prompt := ResumePrompt(session)
	
	if prompt == "" {
		t.Error("ResumePrompt should not return empty string")
	}
	
	// Check that key information is present
	if !contains(prompt, "30 minutes") {
		t.Error("Prompt should contain duration")
	}
	
	if !contains(prompt, "3/10") {
		t.Error("Prompt should contain progress")
	}
	
	if !contains(prompt, "30%") {
		t.Error("Prompt should contain percentage")
	}
	
	if !contains(prompt, "status:pending") {
		t.Error("Prompt should contain filter")
	}
}