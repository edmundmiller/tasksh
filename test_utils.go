package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestTasksh manages a tasksh instance for testing
type TestTasksh struct {
	t           *testing.T
	binaryPath  string
	workingDir  string
	tempDirs    []string
	env         []string
}

// ExitError represents a command that exited with non-zero status
type ExitError struct {
	ExitCode int
	Output   string
	Stderr   string
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("command exited with code %d: %s", e.ExitCode, e.Stderr)
}

// NewTestTasksh creates a new test instance
func NewTestTasksh(t *testing.T) *TestTasksh {
	binaryPath := findTaskshBinary()
	
	ts := &TestTasksh{
		t:          t,
		binaryPath: binaryPath,
		workingDir: getCurrentDir(),
		tempDirs:   make([]string, 0),
		env:        os.Environ(),
	}
	
	return ts
}

// Run executes tasksh with the given arguments
func (ts *TestTasksh) Run(args ...string) (string, error) {
	return ts.RunWithTimeout(30*time.Second, args...)
}

// RunWithTimeout executes tasksh with a timeout
func (ts *TestTasksh) RunWithTimeout(timeout time.Duration, args ...string) (string, error) {
	cmd := exec.Command(ts.binaryPath, args...)
	cmd.Dir = ts.workingDir
	cmd.Env = ts.env
	
	// Create a channel to signal completion
	done := make(chan error, 1)
	var output []byte
	var stderr []byte
	
	go func() {
		var err error
		output, err = cmd.Output()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				stderr = exitError.Stderr
			}
		}
		done <- err
	}()
	
	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				return string(output), &ExitError{
					ExitCode: exitError.ExitCode(),
					Output:   string(output),
					Stderr:   string(stderr),
				}
			}
			return string(output), err
		}
		return string(output), nil
	case <-time.After(timeout):
		cmd.Process.Kill()
		return "", fmt.Errorf("command timed out after %v", timeout)
	}
}

// RunExpectingError executes tasksh expecting a non-zero exit code
func (ts *TestTasksh) RunExpectingError(args ...string) (string, error) {
	output, err := ts.Run(args...)
	if err == nil {
		return output, fmt.Errorf("expected command to fail, but it succeeded with output: %s", output)
	}
	return output, err
}

// CreateTempTaskwarriorConfig creates a temporary Taskwarrior configuration
func (ts *TestTasksh) CreateTempTaskwarriorConfig() string {
	tempDir, err := os.MkdirTemp("", "tasksh_test_")
	if err != nil {
		if ts.t != nil {
			ts.t.Fatalf("Failed to create temp directory: %v", err)
		}
		panic(err)
	}
	
	ts.tempDirs = append(ts.tempDirs, tempDir)
	
	// Set TASKDATA environment variable to isolate Taskwarrior data
	ts.env = append(ts.env, fmt.Sprintf("TASKDATA=%s", tempDir))
	ts.env = append(ts.env, fmt.Sprintf("TASKRC=%s", filepath.Join(tempDir, ".taskrc")))
	
	// Create basic .taskrc
	taskrc := filepath.Join(tempDir, ".taskrc")
	taskrcContent := `data.location=` + tempDir + `
confirmation=no
verbose=nothing
`
	
	if err := os.WriteFile(taskrc, []byte(taskrcContent), 0644); err != nil {
		if ts.t != nil {
			ts.t.Fatalf("Failed to create .taskrc: %v", err)
		}
		panic(err)
	}
	
	return tempDir
}

// AddTask adds a test task using the task command
func (ts *TestTasksh) AddTask(description string, attributes ...string) error {
	args := append([]string{"add", description}, attributes...)
	cmd := exec.Command("task", args...)
	cmd.Env = ts.env
	cmd.Dir = ts.workingDir
	
	_, err := cmd.Output()
	return err
}

// Cleanup removes temporary directories and resources
func (ts *TestTasksh) Cleanup() {
	for _, dir := range ts.tempDirs {
		os.RemoveAll(dir)
	}
}

// SetEnv sets an environment variable for the test instance
func (ts *TestTasksh) SetEnv(key, value string) {
	// Remove existing env var if it exists
	newEnv := make([]string, 0, len(ts.env))
	prefix := key + "="
	
	for _, env := range ts.env {
		if !strings.HasPrefix(env, prefix) {
			newEnv = append(newEnv, env)
		}
	}
	
	// Add new value
	newEnv = append(newEnv, fmt.Sprintf("%s=%s", key, value))
	ts.env = newEnv
}

// findTaskshBinary locates the tasksh binary for testing
func findTaskshBinary() string {
	// First, try the binary in the current directory (development)
	if _, err := os.Stat("./tasksh"); err == nil {
		abs, _ := filepath.Abs("./tasksh")
		return abs
	}
	
	// Then try building it if it doesn't exist
	if err := exec.Command("go", "build", "-o", "tasksh").Run(); err == nil {
		if abs, err := filepath.Abs("./tasksh"); err == nil {
			return abs
		}
	}
	
	// Finally, try PATH
	if path, err := exec.LookPath("tasksh"); err == nil {
		return path
	}
	
	panic("tasksh binary not found - run 'go build -o tasksh' first")
}

// getCurrentDir returns the current working directory
func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current directory: %v", err))
	}
	return dir
}

// IsTaskwarriorAvailable checks if the task command is available
func IsTaskwarriorAvailable() bool {
	cmd := exec.Command("task", "version")
	return cmd.Run() == nil
}

// Helper assertion functions

// AssertContains checks if the output contains the expected substring
func AssertContains(t *testing.T, output, expected string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain '%s', but got: %s", expected, output)
	}
}

// AssertNotContains checks if the output does not contain the substring
func AssertNotContains(t *testing.T, output, unexpected string) {
	t.Helper()
	if strings.Contains(output, unexpected) {
		t.Errorf("Expected output to NOT contain '%s', but got: %s", unexpected, output)
	}
}

// AssertMatches checks if the output matches the regex pattern
func AssertMatches(t *testing.T, output, pattern string) {
	t.Helper()
	matched, err := regexp.MatchString(pattern, output)
	if err != nil {
		t.Fatalf("Invalid regex pattern '%s': %v", pattern, err)
	}
	if !matched {
		t.Errorf("Expected output to match pattern '%s', but got: %s", pattern, output)
	}
}

// AssertNotMatches checks if the output does not match the regex pattern
func AssertNotMatches(t *testing.T, output, pattern string) {
	t.Helper()
	matched, err := regexp.MatchString(pattern, output)
	if err != nil {
		t.Fatalf("Invalid regex pattern '%s': %v", pattern, err)
	}
	if matched {
		t.Errorf("Expected output to NOT match pattern '%s', but got: %s", pattern, output)
	}
}

// AssertExitCode checks if the error has the expected exit code
func AssertExitCode(t *testing.T, err error, expectedCode int) {
	t.Helper()
	if exitErr, ok := err.(*ExitError); ok {
		if exitErr.ExitCode != expectedCode {
			t.Errorf("Expected exit code %d, got %d", expectedCode, exitErr.ExitCode)
		}
	} else {
		t.Errorf("Expected ExitError with code %d, got: %v", expectedCode, err)
	}
}