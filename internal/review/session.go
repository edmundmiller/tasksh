package review

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Session represents a review session that can be saved/loaded
type Session struct {
	ID           string    `json:"id"`
	StartTime    time.Time `json:"start_time"`
	LastActive   time.Time `json:"last_active"`
	Tasks        []string  `json:"tasks"`        // Changed from TaskUUIDs
	Current      int       `json:"current"`       // Changed from CurrentIndex
	Reviewed     int       `json:"reviewed"`      // Changed from ReviewedCount
	Filter       string    `json:"filter,omitempty"`
}

// SessionManager handles session persistence
type SessionManager struct {
	sessionPath string
}

// NewSessionManager creates a new session manager
func NewSessionManager() (*SessionManager, error) {
	// Get cache directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	
	// Create tasksh directory
	taskshDir := filepath.Join(cacheDir, "tasksh")
	if err := os.MkdirAll(taskshDir, 0755); err != nil {
		return nil, err
	}
	
	return &SessionManager{
		sessionPath: filepath.Join(taskshDir, "review_session.json"),
	}, nil
}

// SaveSession saves the current session
func (sm *SessionManager) SaveSession(session *Session) error {
	session.LastActive = time.Now()
	
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	
	// Write atomically
	tmpPath := sm.sessionPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session: %w", err)
	}
	
	if err := os.Rename(tmpPath, sm.sessionPath); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	
	return nil
}

// LoadSession loads a saved session
func (sm *SessionManager) LoadSession() (*Session, error) {
	data, err := os.ReadFile(sm.sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No session exists
		}
		return nil, fmt.Errorf("failed to read session: %w", err)
	}
	
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}
	
	// Check if session is too old (24 hours)
	if time.Since(session.LastActive) > 24*time.Hour {
		sm.ClearSession()
		return nil, nil
	}
	
	return &session, nil
}

// ClearSession removes the saved session
func (sm *SessionManager) ClearSession() error {
	err := os.Remove(sm.sessionPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear session: %w", err)
	}
	return nil
}

// CreateSession creates a new session
func CreateSession(taskUUIDs []string, filter string) *Session {
	return &Session{
		ID:         fmt.Sprintf("session_%d", time.Now().Unix()),
		StartTime:  time.Now(),
		LastActive: time.Now(),
		Tasks:      taskUUIDs,
		Current:    0,
		Reviewed:   0,
		Filter:     filter,
	}
}

// ResumePrompt creates a prompt for resuming a session
func ResumePrompt(session *Session) string {
	duration := time.Since(session.StartTime).Round(time.Minute)
	progress := float64(session.Reviewed) / float64(len(session.Tasks)) * 100
	
	return fmt.Sprintf(
		"Resume previous session?\n"+
		"Started: %s ago\n"+
		"Progress: %d/%d tasks (%.0f%%)\n"+
		"Filter: %s",
		formatDuration(duration),
		session.Reviewed,
		len(session.Tasks),
		progress,
		session.Filter,
	)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%d hours %d minutes", hours, minutes)
		}
		return fmt.Sprintf("%d hours", hours)
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%d days", days)
}


// Add these fields to the Model struct:
// sessionManager *SessionManager
// session        *Session

// Add to Model initialization:
// sessionManager, _ := NewSessionManager()
// m.sessionManager = sessionManager

// Check for existing session on startup:
// if session, err := sessionManager.LoadSession(); err == nil && session != nil {
//     // Prompt user to resume
// }