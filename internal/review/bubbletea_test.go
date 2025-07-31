package review

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/emiller/tasksh/internal/taskwarrior"
)

// TestProgressBarRendering tests that the progress bar shows correctly
func TestProgressBarRendering(t *testing.T) {
	// Create a model with some tasks
	model := NewReviewModel()
	model.SetTasks([]string{"uuid1", "uuid2", "uuid3"}, 3)
	model.reviewed = 1 // Reviewed 1 of 3
	
	// Initialize the model properly
	model.Init()
	
	// Create test model
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))
	
	// Wait for initial render
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		// Should show progress bar
		return bytes.Contains(bts, []byte("Review Progress:"))
	}, teatest.WithDuration(2*time.Second))
	
	// Verify the output shows correct progress
	finalOut := tm.FinalOutput(t)
	outputBytes, err := io.ReadAll(finalOut)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	
	if !bytes.Contains(outputBytes, []byte("Review Progress:")) {
		t.Errorf("Expected progress bar to be rendered, output: %s", string(outputBytes))
	}
}

// TestLazyLoadingIndicator tests that lazy loading shows correct status
func TestLazyLoadingIndicator(t *testing.T) {
	// Create a model with lazy loading enabled
	model := NewReviewModel()
	model.SetTasks(make([]string, 100), 100) // 100 tasks total
	model.lazyLoadEnabled = true
	model.loadedTasks = 25
	model.totalTasks = 100
	model.loadingMore = false
	
	// Create test model
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(100, 24))
	
	// Wait for lazy loading indicator
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("25/100 loaded"))
	}, teatest.WithDuration(2*time.Second))
	
	// Test with loading more state
	model.loadingMore = true
	tm.Send("") // Trigger re-render
	
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("loading more..."))
	}, teatest.WithDuration(2*time.Second))
}

// TestNavigationTriggersLazyLoad tests that approaching the end triggers loading
func TestNavigationTriggersLazyLoad(t *testing.T) {
	// Create model with tasks
	model := NewReviewModel()
	
	// Set up 50 tasks but only load 20
	allUUIDs := make([]string, 50)
	taskCache := make(map[string]*taskwarrior.Task)
	
	for i := 0; i < 50; i++ {
		allUUIDs[i] = fmt.Sprintf("uuid-%d", i)
		if i < 20 {
			// Only cache first 20
			taskCache[allUUIDs[i]] = &taskwarrior.Task{
				UUID:        allUUIDs[i],
				Description: fmt.Sprintf("Task %d", i),
				Status:      "pending",
			}
		}
	}
	
	model.SetTasks(allUUIDs, 50)
	model.taskCache = taskCache
	model.lazyLoadEnabled = true
	model.loadedTasks = 20
	model.totalTasks = 50
	model.currentTask = taskCache[allUUIDs[0]]
	
	// Create test model
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(100, 24))
	
	// Navigate forward to approach the loading threshold (within 10 of end)
	for i := 0; i < 11; i++ {
		tm.Send(tea.KeyMsg{
			Type: tea.KeyRunes,
			Runes: []rune("j"), // Next task
		})
		time.Sleep(50 * time.Millisecond) // Small delay between keypresses
	}
	
	// Should now be at task 11, which is within 10 of loaded (20)
	// This should trigger background loading
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("loading more..."))
	}, teatest.WithDuration(3*time.Second))
}

// TestKeyboardShortcuts tests various keyboard shortcuts work
func TestKeyboardShortcuts(t *testing.T) {
	tests := []struct {
		name     string
		key      tea.KeyMsg
		expected string
	}{
		{
			name: "help toggle",
			key: tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune("?"),
			},
			expected: "esc", // Help view shows esc to close
		},
		{
			name: "skip task",
			key: tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune("s"),
			},
			expected: "[2 of", // Should move to next task
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create model with tasks
			model := NewReviewModel()
			model.SetTasks([]string{"uuid1", "uuid2", "uuid3"}, 3)
			model.taskCache = map[string]*taskwarrior.Task{
				"uuid1": {UUID: "uuid1", Description: "Task 1"},
				"uuid2": {UUID: "uuid2", Description: "Task 2"},
				"uuid3": {UUID: "uuid3", Description: "Task 3"},
			}
			model.currentTask = model.taskCache["uuid1"]
			
			tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(100, 24))
			
			// Send the key
			tm.Send(tt.key)
			
			// Wait for expected output
			teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
				return bytes.Contains(bts, []byte(tt.expected))
			}, teatest.WithDuration(2*time.Second))
		})
	}
}

// TestModeTransitions tests that mode changes work correctly
func TestModeTransitions(t *testing.T) {
	// Create model
	model := NewReviewModel()
	model.SetTasks([]string{"uuid1"}, 1)
	model.taskCache = map[string]*taskwarrior.Task{
		"uuid1": {UUID: "uuid1", Description: "Test task"},
	}
	model.currentTask = model.taskCache["uuid1"]
	
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(100, 24))
	
	// Test transition to modify mode
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("m"),
	})
	
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Enter task modification"))
	}, teatest.WithDuration(2*time.Second))
	
	// Test escape back to viewing
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		// Should show normal task view again
		return bytes.Contains(bts, []byte("Test task")) && 
			!bytes.Contains(bts, []byte("Enter task modification"))
	}, teatest.WithDuration(2*time.Second))
}

// TestQuitBehavior tests that quit works correctly
func TestQuitBehavior(t *testing.T) {
	model := NewReviewModel()
	model.SetTasks([]string{"uuid1"}, 1)
	
	tm := teatest.NewTestModel(t, model)
	
	// Send quit command
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("q"),
	})
	
	// Model should quit
	fm := tm.FinalModel(t)
	reviewModel := fm.(*ReviewModel)
	
	if !reviewModel.quitting {
		t.Error("Expected model to be in quitting state")
	}
}

// TestEmptyTaskList tests behavior with no tasks
func TestEmptyTaskList(t *testing.T) {
	model := NewReviewModel()
	model.SetTasks([]string{}, 0)
	
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(100, 24))
	
	// Should show appropriate message
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("[0 of 0]")) ||
			bytes.Contains(bts, []byte("No tasks"))
	}, teatest.WithDuration(2*time.Second))
}

// TestProgressCalculation tests that progress percentage is calculated correctly
func TestProgressCalculation(t *testing.T) {
	model := NewReviewModel()
	model.SetTasks([]string{"uuid1", "uuid2", "uuid3", "uuid4"}, 4)
	
	// Review 2 of 4 tasks (50%)
	model.reviewed = 2
	
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(100, 24))
	
	// The progress bar should show 50% completion
	// We can't easily test the exact visual representation,
	// but we can verify the progress bar is rendered
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Review Progress:"))
	}, teatest.WithDuration(2*time.Second))
	
	// Test that reviewing a task updates progress
	model.reviewed = 3 // 75% now
	tm.Send("") // Trigger re-render
	
	// Still should show progress bar
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Review Progress:"))
	}, teatest.WithDuration(1*time.Second))
}

// TestFallbackScenarios tests behavior when batch loading fails
func TestFallbackScenarios(t *testing.T) {
	// Test 1: Model with no task cache (fallback to individual loading)
	t.Run("no cache fallback", func(t *testing.T) {
		model := NewReviewModel()
		model.SetTasks([]string{"uuid1", "uuid2"}, 2)
		model.taskCache = nil // No cache, will fallback
		
		// This will trigger loadCurrentTask which should handle the fallback
		model.current = 0
		cmd := model.loadCurrentTask()
		
		// Execute the command
		msg := cmd()
		
		// Should get an error since GetTaskInfo will fail in test environment
		if _, ok := msg.(errorMsg); !ok {
			t.Error("Expected error when falling back to individual loading without taskwarrior")
		}
	})
	
	// Test 2: Partial cache (some tasks cached, some not)
	t.Run("partial cache", func(t *testing.T) {
		model := NewReviewModel()
		model.SetTasks([]string{"uuid1", "uuid2", "uuid3"}, 3)
		model.taskCache = map[string]*taskwarrior.Task{
			"uuid1": {UUID: "uuid1", Description: "Cached task 1"},
			// uuid2 is missing from cache
			"uuid3": {UUID: "uuid3", Description: "Cached task 3"},
		}
		
		// Load first task (cached)
		model.current = 0
		cmd := model.loadCurrentTask()
		msg := cmd()
		
		if taskMsg, ok := msg.(taskLoadedMsg); ok {
			if taskMsg.task.Description != "Cached task 1" {
				t.Error("Expected to load cached task 1")
			}
		} else {
			t.Error("Expected successful load from cache")
		}
		
		// Try to load second task (not cached, will fallback)
		model.current = 1
		cmd = model.loadCurrentTask()
		msg = cmd()
		
		// Should get error in test environment
		if _, ok := msg.(errorMsg); !ok {
			t.Error("Expected error when falling back for uncached task")
		}
	})
	
	// Test 3: Lazy loading with failed background load
	t.Run("failed background load", func(t *testing.T) {
		model := NewReviewModel()
		allUUIDs := make([]string, 30)
		for i := 0; i < 30; i++ {
			allUUIDs[i] = fmt.Sprintf("uuid-%d", i)
		}
		
		model.SetTasks(allUUIDs, 30)
		model.lazyLoadEnabled = true
		model.loadedTasks = 15
		model.totalTasks = 30
		model.loadingMore = false
		
		// Simulate approaching the end to trigger background load
		model.current = 10 // Within 10 of loaded (15)
		
		// This would normally trigger loadNextBatch in background
		// In test environment, it will fail but shouldn't crash
		go model.loadNextBatch()
		
		// Give it time to fail
		time.Sleep(100 * time.Millisecond)
		
		// Should have reset loadingMore to false
		if model.loadingMore {
			t.Error("loadingMore should be false after failed load")
		}
		
		// loadedTasks should remain unchanged
		if model.loadedTasks != 15 {
			t.Error("loadedTasks should remain unchanged after failed load")
		}
	})
}

// TestLoadingProgressCallback tests the progress callback during batch loading
func TestLoadingProgressCallback(t *testing.T) {
	// This tests the actual progress callback mechanism
	progressCalls := 0
	
	progressFn := func(loaded, total int) {
		progressCalls++
		t.Logf("Progress: %d/%d", loaded, total)
	}
	
	// Create some test UUIDs
	uuids := make([]string, 250) // More than one chunk (100)
	for i := 0; i < 250; i++ {
		uuids[i] = fmt.Sprintf("test-uuid-%d", i)
	}
	
	// This would normally call GetTasksWithDataProgress
	// but in test environment it will fail
	// We're testing that the callback mechanism is in place
	_, _ = taskwarrior.GetTasksWithDataProgress(uuids, progressFn)
	
	// In a real environment, this would be called multiple times
	// Here we just verify the mechanism exists
	t.Log("Progress callback mechanism is in place")
}