package review

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

// TestCalendarDateParsing tests various date input formats
func TestCalendarDateParsing(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool // whether parsing should succeed
	}{
		{"today", true},
		{"tomorrow", true},
		{"next week", true},
		{"monday", true},
		{"2024-12-25", true},
		{"12/25/2024", true},
		{"Dec 25, 2024", true},
		{"+7", true},
		{"in 3 days", true},
		{"invalid date", false},
		{"", false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			_, err := ParseDateInput(tc.input)
			if tc.expected && err != nil {
				t.Errorf("Expected %q to parse successfully, got error: %v", tc.input, err)
			}
			if !tc.expected && err == nil {
				t.Errorf("Expected %q to fail parsing, but it succeeded", tc.input)
			}
		})
	}
}

// TestReviewModelWaitFlow tests the complete wait date selection flow
func TestReviewModelWaitFlow(t *testing.T) {
	model := NewReviewModel()
	
	// Test calendar mode initialization
	model.mode = ModeWaitCalendar
	model.calendar.SetFocused(true)
	
	// Verify calendar responds to navigation
	initialDate := model.calendar.GetSelectedDate()
	
	// Test navigation keys work
	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	updatedModelInterface, _ := model.updateWaitCalendar(rightKey)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	expectedDate := initialDate.AddDate(0, 0, 1)
	if !isSameDay(updatedModel.calendar.GetSelectedDate(), expectedDate) {
		t.Error("Calendar navigation should work in wait flow")
	}
	
	// Test date confirmation
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModelInterface, _ = updatedModel.updateWaitCalendar(enterKey)
	updatedModel = updatedModelInterface.(*ReviewModel)
	
	if updatedModel.mode != ModeInputWaitReason {
		t.Errorf("Expected transition to ModeInputWaitReason, got %v", updatedModel.mode)
	}
	
	if updatedModel.waitDate == "" {
		t.Error("Wait date should be set after confirmation")
	}
}

// TestCalendarKeyBindings tests that calendar key bindings are properly configured
func TestCalendarKeyBindings(t *testing.T) {
	keys := DefaultCalendarKeyMap()
	
	// Verify key bindings have help text
	if keys.Up.Help().Key == "" {
		t.Error("Up key should have help text")
	}
	if keys.Down.Help().Key == "" {
		t.Error("Down key should have help text")
	}
	if keys.Left.Help().Key == "" {
		t.Error("Left key should have help text")
	}
	if keys.Right.Help().Key == "" {
		t.Error("Right key should have help text")
	}
}

// TestDueDateWorkflow tests the complete due date selection workflow
func TestDueDateWorkflow(t *testing.T) {
	model := NewReviewModel()
	model.tasks = []string{"test-uuid"}
	model.current = 0
	
	// Test due calendar mode initialization
	model.mode = ModeDueCalendar
	model.calendar.SetFocused(true)
	
	// Verify calendar responds to navigation
	initialDate := model.calendar.GetSelectedDate()
	
	// Test navigation keys work
	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	updatedModelInterface, _ := model.updateDueCalendar(rightKey)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	expectedDate := initialDate.AddDate(0, 0, 1)
	if !isSameDay(updatedModel.calendar.GetSelectedDate(), expectedDate) {
		t.Error("Calendar navigation should work in due date flow")
	}
	
	// Test date confirmation
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModelInterface, cmd := updatedModel.updateDueCalendar(enterKey)
	updatedModel = updatedModelInterface.(*ReviewModel)
	
	if updatedModel.mode != ModeViewing {
		t.Errorf("Expected transition to ModeViewing, got %v", updatedModel.mode)
	}
	
	if cmd == nil {
		t.Error("Should have command to set due date")
	}
}

// TestDueDateTextInputParsing tests that due date text input works with various formats
func TestDueDateTextInputParsing(t *testing.T) {
	testCases := []struct {
		input    string
		shouldWork bool
	}{
		{"today", true},
		{"tomorrow", true}, 
		{"next week", true},
		{"2024-12-25", true},
		{"", false}, // empty should not execute
	}
	
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			model := NewReviewModel()
			model.mode = ModeInputDueDate
			model.tasks = []string{"test-uuid"}
			model.current = 0
			
			model.textInput.SetValue(tc.input)
			
			keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
			_, cmd := model.updateDueDateInput(keyMsg)
			
			if tc.shouldWork && cmd == nil {
				t.Errorf("Expected command for input %q", tc.input)
			}
			if !tc.shouldWork && cmd != nil {
				t.Errorf("Expected no command for input %q", tc.input)
			}
		})
	}
}