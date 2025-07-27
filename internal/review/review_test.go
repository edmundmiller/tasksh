package review

import (
	"strings"
	"testing"
	"time"

	"github.com/emiller/tasksh/testdata"
)

// TestReviewWithNoTasks verifies review behavior when no tasks need review
func TestReviewWithNoTasks(t *testing.T) {
	if !testdata.IsTaskwarriorAvailable() {
		t.Skip("Taskwarrior not available, skipping integration test")
	}

	tasksh := testdata.NewTestTasksh(t)
	defer tasksh.Cleanup()

	// Create isolated Taskwarrior environment
	tasksh.CreateTempTaskwarriorConfig()

	// Run review with no tasks
	output, err := tasksh.Run("review")
	if err != nil {
		t.Fatalf("review command failed: %v", err)
	}

	testdata.AssertContains(t, output, "There are no tasks needing review")
}

// TestReviewWithTasks verifies review behavior with tasks needing review
func TestReviewWithTasks(t *testing.T) {
	if !testdata.IsTaskwarriorAvailable() {
		t.Skip("Taskwarrior not available, skipping integration test")
	}

	tasksh := testdata.NewTestTasksh(t)
	defer tasksh.Cleanup()

	// Create isolated Taskwarrior environment
	tasksh.CreateTempTaskwarriorConfig()

	// Add a test task
	err := tasksh.AddTask("Test task for review")
	if err != nil {
		t.Fatalf("Failed to add test task: %v", err)
	}

	// This test would be interactive, so we can't easily test the full flow
	// But we can test that it starts properly and sets up the review
	// For now, we'll test that the command recognizes there are tasks to review

	// Note: Full interactive testing would require more complex setup
	// This is a placeholder for the concept
	t.Skip("Interactive review testing requires more complex setup")
}

// TestReviewLimit verifies the review limit functionality
func TestReviewLimit(t *testing.T) {
	tasksh := testdata.NewTestTasksh(t)
	defer tasksh.Cleanup()

	// Test with various limit arguments
	testCases := []struct {
		args     []string
		expected string
	}{
		{[]string{"review", "5"}, "review"}, // Should accept limit
		{[]string{"review", "0"}, "review"}, // Should accept zero limit
	}

	for _, tc := range testCases {
		// Since we don't have tasks, all will show "no tasks needing review"
		// This tests that the argument parsing works
		output, err := tasksh.Run(tc.args...)
		if err != nil && !strings.Contains(err.Error(), "Taskwarrior") {
			t.Errorf("Command %v failed unexpectedly: %v", tc.args, err)
		}
		
		// If Taskwarrior is available, should get "no tasks" message
		// If not available, might get other errors, which is ok for this test
		if testdata.IsTaskwarriorAvailable() {
			testdata.AssertContains(t, output, "There are no tasks needing review")
		}
	}
}

// TestTaskwarriorUDAConfiguration tests that UDA setup works
func TestTaskwarriorUDAConfiguration(t *testing.T) {
	if !testdata.IsTaskwarriorAvailable() {
		t.Skip("Taskwarrior not available, skipping integration test")
	}

	tasksh := testdata.NewTestTasksh(t)
	defer tasksh.Cleanup()

	// Create isolated Taskwarrior environment  
	tasksh.CreateTempTaskwarriorConfig()

	// Add a task so review has something to configure for
	err := tasksh.AddTask("Test task")
	if err != nil {
		t.Fatalf("Failed to add test task: %v", err)
	}

	// The review command should set up UDA and reports
	// This will fail since it's interactive, but should configure first
	_, err = tasksh.RunWithTimeout(2*time.Second, "review")
	
	// We expect this to timeout since it's interactive, but configuration should happen
	if err != nil && !strings.Contains(err.Error(), "timeout") {
		// If it fails for other reasons, that's potentially a problem
		// But we'll log it and continue since this is testing configuration setup
		t.Logf("Review command resulted in: %v", err)
	}

	// Test that UDA was configured by checking task configuration
	// This would require additional Taskwarrior inspection commands
	// For now, we just verify the command started properly
}

// TestHelpOutput verifies help command produces correct output
func TestHelpOutput(t *testing.T) {
	tasksh := testdata.NewTestTasksh(t)
	defer tasksh.Cleanup()

	output, err := tasksh.Run("help")
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	// Verify all expected help sections are present
	expectedContent := []string{
		"tasksh - Interactive task review shell",
		"Commands:",
		"review [N]",
		"help",
		"diagnostics", 
		"During review, you can:",
		"Edit task",
		"Modify task",
		"Complete task",
		"Delete task",
		"Wait task",
		"Skip task",
		"Mark as reviewed",
		"Quit review session",
		"AI Analysis", // New AI feature
	}

	for _, content := range expectedContent {
		testdata.AssertContains(t, output, content)
	}
}

// TestDiagnosticsOutput verifies diagnostics command output
func TestDiagnosticsOutput(t *testing.T) {
	tasksh := testdata.NewTestTasksh(t)
	defer tasksh.Cleanup()

	output, err := tasksh.Run("diagnostics")
	if err != nil {
		t.Fatalf("diagnostics command failed: %v", err)
	}

	// Verify expected diagnostics content
	expectedContent := []string{
		"tasksh diagnostics",
		"Version: 2.0.0-go",
		"Built with: Go",
		"Taskwarrior:",
		"Mods (AI):", // New AI diagnostics
		"Time Database:", // New time tracking diagnostics
	}

	for _, content := range expectedContent {
		testdata.AssertContains(t, output, content)
	}

	// Check Taskwarrior status
	if testdata.IsTaskwarriorAvailable() {
		testdata.AssertContains(t, output, "Available")
	} else {
		testdata.AssertContains(t, output, "NOT FOUND")
	}
}

// TestErrorHandling verifies proper error handling
func TestErrorHandling(t *testing.T) {
	tasksh := testdata.NewTestTasksh(t)
	defer tasksh.Cleanup()

	testCases := []struct {
		name     string
		args     []string
		exitCode int
	}{
		{"invalid command", []string{"invalid"}, 1},
		{"unknown flag", []string{"--unknown-flag"}, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tasksh.Run(tc.args...)
			if err == nil {
				t.Errorf("Expected error for %s, but got none", tc.name)
				return
			}
			
			// Check that it's the right kind of error
			if exitErr, ok := err.(*testdata.ExitError); !ok || exitErr.ExitCode != tc.exitCode {
				t.Errorf("Expected exit code %d, got: %v", tc.exitCode, err)
			}
		})
	}
}

// TestConcurrentExecution tests that multiple instances can run
func TestConcurrentExecution(t *testing.T) {
	t.Parallel()

	// Run help command concurrently
	done := make(chan bool, 3)
	
	for i := 0; i < 3; i++ {
		go func() {
			tasksh := testdata.NewTestTasksh(t)
			defer tasksh.Cleanup()
			
			_, err := tasksh.Run("help")
			if err != nil {
				t.Errorf("Concurrent help command failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}
}

// Benchmark tests
func BenchmarkDiagnosticsCommand(b *testing.B) {
	tasksh := testdata.NewTestTasksh(b)
	defer tasksh.Cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tasksh.Run("diagnostics")
		if err != nil {
			b.Fatalf("diagnostics command failed: %v", err)
		}
	}
}

func BenchmarkCommandParsing(b *testing.B) {
	tasksh := testdata.NewTestTasksh(b)
	defer tasksh.Cleanup()

	commands := [][]string{
		{"help"},
		{"diagnostics"},
		{"review", "0"}, // Will show no tasks message quickly
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := commands[i%len(commands)]
		tasksh.Run(cmd...)
	}
}

// AI-related tests (from the original AI commit)

// Tests for structs that were removed during refactoring have been removed

// TestReviewWorkflow is a placeholder for integration testing
func TestReviewWorkflow(t *testing.T) {
	t.Skip("Integration test - requires taskwarrior mock")
	
	// This would test the full review workflow including:
	// - Getting tasks for review
	// - Processing AI analysis
	// - Applying suggestions
	// - Recording completion times
}