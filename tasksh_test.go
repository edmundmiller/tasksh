package main

import (
	"regexp"
	"testing"
)

// TestVersion verifies that the version command returns valid output
func TestVersion(t *testing.T) {
	// Create test instance
	tasksh := NewTestTasksh(t)
	defer tasksh.Cleanup()

	// Test --version flag doesn't exist in our current implementation
	// Instead test that the binary runs without arguments
	output, err := tasksh.Run()
	if err != nil {
		t.Fatalf("tasksh command failed: %v", err)
	}

	// Should contain help output
	if !regexp.MustCompile(`tasksh - Interactive task review shell`).MatchString(output) {
		t.Errorf("Expected help output, got: %s", output)
	}
}

// TestHelp verifies the help command works
func TestHelp(t *testing.T) {
	tasksh := NewTestTasksh(t)
	defer tasksh.Cleanup()

	output, err := tasksh.Run("help")
	if err != nil {
		t.Fatalf("tasksh help failed: %v", err)
	}

	// Check for expected help content
	expectedPatterns := []string{
		`tasksh - Interactive task review shell`,
		`Commands:`,
		`review \[N\]`,
		`help`,
		`diagnostics`,
	}

	for _, pattern := range expectedPatterns {
		if !regexp.MustCompile(pattern).MatchString(output) {
			t.Errorf("Help output missing pattern '%s', got: %s", pattern, output)
		}
	}
}

// TestDiagnostics verifies the diagnostics command works
func TestDiagnostics(t *testing.T) {
	tasksh := NewTestTasksh(t)
	defer tasksh.Cleanup()

	output, err := tasksh.Run("diagnostics")
	if err != nil {
		t.Fatalf("tasksh diagnostics failed: %v", err)
	}

	// Check for expected diagnostics content
	expectedPatterns := []string{
		`tasksh diagnostics`,
		`Version: 2\.0\.0-go`,
		`Built with: Go`,
		`Taskwarrior:`,
	}

	for _, pattern := range expectedPatterns {
		if !regexp.MustCompile(pattern).MatchString(output) {
			t.Errorf("Diagnostics output missing pattern '%s', got: %s", pattern, output)
		}
	}
}

// TestInvalidCommand verifies error handling for invalid commands
func TestInvalidCommand(t *testing.T) {
	tasksh := NewTestTasksh(t)
	defer tasksh.Cleanup()

	_, err := tasksh.Run("invalid-command")
	if err == nil {
		t.Error("Expected error for invalid command, but got none")
	}

	// Check that it's the right kind of error
	if exitErr, ok := err.(*ExitError); !ok || exitErr.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got: %v", err)
	}
}

// TestTaskwarriorIntegration verifies basic Taskwarrior integration
func TestTaskwarriorIntegration(t *testing.T) {
	// Skip if Taskwarrior is not available
	if !IsTaskwarriorAvailable() {
		t.Skip("Taskwarrior not available, skipping integration test")
	}

	tasksh := NewTestTasksh(t)
	defer tasksh.Cleanup()

	// Test diagnostics to see if Taskwarrior is detected
	output, err := tasksh.Run("diagnostics")
	if err != nil {
		t.Fatalf("tasksh diagnostics failed: %v", err)
	}

	if !regexp.MustCompile(`Taskwarrior: Available`).MatchString(output) {
		t.Errorf("Expected Taskwarrior to be available, got: %s", output)
	}
}

// Benchmark for basic command execution
func BenchmarkHelpCommand(b *testing.B) {
	tasksh := NewTestTasksh(nil)
	defer tasksh.Cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tasksh.Run("help")
		if err != nil {
			b.Fatalf("help command failed: %v", err)
		}
	}
}