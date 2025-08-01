package review

import (
	"testing"
	"time"
)

// TestFeedbackManager tests the feedback manager
func TestFeedbackManager(t *testing.T) {
	fm := NewFeedbackManager()
	
	if fm == nil {
		t.Fatal("NewFeedbackManager returned nil")
	}
	
	if fm.maxToasts != 3 {
		t.Errorf("Expected maxToasts to be 3, got %d", fm.maxToasts)
	}
	
	if len(fm.toasts) != 0 {
		t.Errorf("Expected empty toasts, got %d", len(fm.toasts))
	}
}

// TestShowToast tests toast creation
func TestShowToast(t *testing.T) {
	fm := NewFeedbackManager()
	
	// Show a toast
	cmd := fm.ShowToast(ToastSuccess, "Test message")
	if cmd == nil {
		t.Error("ShowToast should return a command")
	}
	
	// Check toast was added
	if len(fm.toasts) != 1 {
		t.Errorf("Expected 1 toast, got %d", len(fm.toasts))
	}
	
	toast := fm.toasts[0]
	if toast.Type != ToastSuccess {
		t.Errorf("Expected ToastSuccess, got %v", toast.Type)
	}
	
	if toast.Message != "Test message" {
		t.Errorf("Expected 'Test message', got %q", toast.Message)
	}
}

// TestMaxToasts tests toast limit
func TestMaxToasts(t *testing.T) {
	fm := NewFeedbackManager()
	fm.maxToasts = 2
	
	// Add 3 toasts
	fm.ShowToast(ToastInfo, "Toast 1")
	fm.ShowToast(ToastInfo, "Toast 2")
	fm.ShowToast(ToastInfo, "Toast 3")
	
	// Should only have 2 toasts (most recent)
	if len(fm.toasts) != 2 {
		t.Errorf("Expected 2 toasts, got %d", len(fm.toasts))
	}
	
	// Check that we have the most recent toasts
	if fm.toasts[0].Message != "Toast 2" {
		t.Errorf("Expected 'Toast 2', got %q", fm.toasts[0].Message)
	}
	
	if fm.toasts[1].Message != "Toast 3" {
		t.Errorf("Expected 'Toast 3', got %q", fm.toasts[1].Message)
	}
}

// TestRemoveExpiredToast tests toast removal
func TestRemoveExpiredToast(t *testing.T) {
	fm := NewFeedbackManager()
	
	// Add toasts with specific start times
	now := time.Now()
	fm.toasts = []Toast{
		{StartTime: now.Add(-2 * time.Second), Message: "Old"},
		{StartTime: now, Message: "New"},
	}
	
	// Remove the old toast
	fm.RemoveExpiredToast(now.Add(-2 * time.Second))
	
	// Should only have 1 toast
	if len(fm.toasts) != 1 {
		t.Errorf("Expected 1 toast, got %d", len(fm.toasts))
	}
	
	if fm.toasts[0].Message != "New" {
		t.Errorf("Wrong toast remained: %q", fm.toasts[0].Message)
	}
}

// TestProgressIndicator tests progress rendering
func TestProgressIndicator(t *testing.T) {
	tests := []struct {
		current int
		total   int
		width   int
	}{
		{0, 10, 20},
		{5, 10, 20},
		{10, 10, 20},
		{0, 0, 20}, // Edge case: no total
	}
	
	for _, tt := range tests {
		pi := NewProgressIndicator(tt.current, tt.total, tt.width)
		result := pi.Render()
		
		if tt.total == 0 && result != "" {
			t.Errorf("Expected empty result for zero total")
		}
		
		if tt.total > 0 && result == "" {
			t.Errorf("Expected non-empty result for progress %d/%d", 
				tt.current, tt.total)
		}
	}
}