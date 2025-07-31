package review

import (
	"fmt"
	"testing"
	"github.com/emiller/tasksh/internal/taskwarrior"
)

// TestReviewModelCreation tests that we can create a review model
func TestReviewModelCreation(t *testing.T) {
	model := NewReviewModel()
	if model == nil {
		t.Fatal("Failed to create review model")
	}
	
	// Test setting tasks
	model.SetTasks([]string{"uuid1", "uuid2", "uuid3"}, 3)
	if len(model.tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(model.tasks))
	}
	if model.total != 3 {
		t.Errorf("Expected total of 3, got %d", model.total)
	}
}

// TestLazyLoadingSetup tests lazy loading configuration
func TestLazyLoadingSetup(t *testing.T) {
	model := NewReviewModel()
	
	// Setup for lazy loading
	model.SetTasks(make([]string, 100), 100)
	model.lazyLoadEnabled = true
	model.loadedTasks = 25
	model.totalTasks = 100
	
	if !model.lazyLoadEnabled {
		t.Error("Lazy loading should be enabled")
	}
	if model.loadedTasks != 25 {
		t.Errorf("Expected 25 loaded tasks, got %d", model.loadedTasks)
	}
}

// TestTaskCaching tests the task cache functionality
func TestTaskCaching(t *testing.T) {
	model := NewReviewModel()
	model.SetTasks([]string{"uuid1", "uuid2"}, 2)
	
	// Add tasks to cache
	model.taskCache = map[string]*taskwarrior.Task{
		"uuid1": {
			UUID:        "uuid1",
			Description: "Test task 1",
			Status:      "pending",
		},
		"uuid2": {
			UUID:        "uuid2",
			Description: "Test task 2",
			Status:      "pending",
		},
	}
	
	// Test cache retrieval
	if task, ok := model.taskCache["uuid1"]; !ok {
		t.Error("Failed to retrieve task from cache")
	} else if task.Description != "Test task 1" {
		t.Errorf("Wrong task description: %s", task.Description)
	}
}

// TestProgressCalculationLogic tests progress calculation
func TestProgressCalculationLogic(t *testing.T) {
	model := NewReviewModel()
	model.SetTasks([]string{"uuid1", "uuid2", "uuid3", "uuid4"}, 4)
	
	// Test different progress states
	testCases := []struct {
		reviewed int
		expected float64
	}{
		{0, 0.0},
		{1, 0.25},
		{2, 0.5},
		{3, 0.75},
		{4, 1.0},
	}
	
	for _, tc := range testCases {
		model.reviewed = tc.reviewed
		progress := float64(model.reviewed) / float64(len(model.tasks))
		if progress != tc.expected {
			t.Errorf("For %d reviewed of 4: expected %f, got %f", 
				tc.reviewed, tc.expected, progress)
		}
	}
}

// TestLoadCurrentTaskLogic tests the load current task logic
func TestLoadCurrentTaskLogic(t *testing.T) {
	model := NewReviewModel()
	model.SetTasks([]string{"uuid1", "uuid2", "uuid3"}, 3)
	
	// Test with cache
	model.taskCache = map[string]*taskwarrior.Task{
		"uuid1": {UUID: "uuid1", Description: "Cached task"},
	}
	
	// Test loading from cache
	model.current = 0
	cmd := model.loadCurrentTask()
	msg := cmd()
	
	if taskMsg, ok := msg.(taskLoadedMsg); ok {
		if taskMsg.task.Description != "Cached task" {
			t.Error("Should load task from cache")
		}
	} else {
		t.Error("Expected taskLoadedMsg")
	}
	
	// Test out of range
	model.current = 5
	cmd = model.loadCurrentTask()
	msg = cmd()
	
	if _, ok := msg.(errorMsg); !ok {
		t.Error("Expected error for out of range index")
	}
}

// TestBackgroundLoadingTrigger tests when background loading should trigger
func TestBackgroundLoadingTrigger(t *testing.T) {
	model := NewReviewModel()
	
	// Setup lazy loading scenario
	allUUIDs := make([]string, 50)
	for i := 0; i < 50; i++ {
		allUUIDs[i] = fmt.Sprintf("uuid-%d", i)
	}
	
	model.SetTasks(allUUIDs, 50)
	model.lazyLoadEnabled = true
	model.loadedTasks = 20
	model.totalTasks = 50
	model.loadingMore = false
	
	// Test trigger conditions
	testCases := []struct {
		current      int
		shouldTrigger bool
		description  string
	}{
		{5, false, "Far from end"},
		{9, false, "Just outside trigger range"},
		{10, true, "Exactly at trigger point (20-10)"},
		{11, true, "Within 10 of loaded (20)"},
		{15, true, "Getting closer to end"},
		{19, true, "Almost at end of loaded"},
	}
	
	for _, tc := range testCases {
		model.current = tc.current
		model.loadingMore = false // Reset
		
		// Check if we should trigger loading
		shouldLoad := model.lazyLoadEnabled && 
			!model.loadingMore && 
			model.current >= model.loadedTasks-10 && 
			model.loadedTasks < model.totalTasks
			
		if shouldLoad != tc.shouldTrigger {
			t.Errorf("%s: current=%d, expected trigger=%v, got=%v",
				tc.description, tc.current, tc.shouldTrigger, shouldLoad)
		}
	}
}

// TestRenderProgressBarString tests the progress bar rendering logic
func TestRenderProgressBarString(t *testing.T) {
	model := NewReviewModel()
	model.SetTasks([]string{"uuid1", "uuid2"}, 2)
	model.width = 100
	model.height = 24
	
	// Test normal progress bar
	output := model.renderProgressBar()
	if output == "" {
		t.Error("Progress bar should not be empty")
	}
	
	// Test with lazy loading
	model.lazyLoadEnabled = true
	model.loadedTasks = 10
	model.totalTasks = 50
	model.loadingMore = false
	
	output = model.renderProgressBar()
	if output == "" {
		t.Error("Progress bar with lazy loading should not be empty")
	}
	
	// The output should contain loading status
	// We can't easily test the exact output due to styling,
	// but we can ensure it's not empty
	
	// Test with loading more state
	model.loadingMore = true
	output = model.renderProgressBar()
	if output == "" {
		t.Error("Progress bar with loading more should not be empty")
	}
}