package taskwarrior

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Task represents a Taskwarrior task
type Task struct {
	UUID        string
	Description string
	Project     string
	Priority    string
	Status      string
	Due         string
}

// CheckAvailable verifies that the task command is available
func CheckAvailable() error {
	cmd := exec.Command("task", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("task command not found: %w", err)
	}
	return nil
}

// executeTask runs a task command and returns the output
func executeTask(args ...string) (string, error) {
	cmd := exec.Command("task", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("task command failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// EnsureReviewConfig sets up the required UDA and report for review
func EnsureReviewConfig() error {
	// Check if reviewed UDA exists
	output, err := executeTask("_get", "rc.uda.reviewed.type")
	if err != nil || output != "date" {
		fmt.Println("Setting up 'reviewed' UDA...")
		if _, err := executeTask("rc.confirmation:no", "rc.verbose:nothing", "config", "uda.reviewed.type", "date"); err != nil {
			return fmt.Errorf("failed to set reviewed UDA type: %w", err)
		}
		if _, err := executeTask("rc.confirmation:no", "rc.verbose:nothing", "config", "uda.reviewed.label", "Reviewed"); err != nil {
			return fmt.Errorf("failed to set reviewed UDA label: %w", err)
		}
	}

	// Check if _reviewed report exists
	output, err = executeTask("_get", "rc.report._reviewed.columns")
	if err != nil || output != "uuid" {
		fmt.Println("Setting up '_reviewed' report...")
		reportArgs := [][]string{
			{"rc.confirmation:no", "rc.verbose:nothing", "config", "report._reviewed.description", "Tasksh review report. Adjust the filter to your needs."},
			{"rc.confirmation:no", "rc.verbose:nothing", "config", "report._reviewed.columns", "uuid"},
			{"rc.confirmation:no", "rc.verbose:nothing", "config", "report._reviewed.sort", "reviewed+,modified+"},
			{"rc.confirmation:no", "rc.verbose:nothing", "config", "report._reviewed.filter", "( reviewed.none: or reviewed.before:now-6days ) and ( +PENDING or +WAITING )"},
		}
		
		for _, args := range reportArgs {
			if _, err := executeTask(args...); err != nil {
				return fmt.Errorf("failed to configure _reviewed report: %w", err)
			}
		}
	}

	return nil
}


// GetTasksForReview returns a list of task UUIDs that need review
func GetTasksForReview() ([]string, error) {
	output, err := executeTask(
		"rc.color=off",
		"rc.detection=off",
		"rc._forcecolor=off",
		"rc.verbose=nothing",
		"_reviewed",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for review: %w", err)
	}

	if output == "" {
		return []string{}, nil
	}

	return strings.Split(output, "\n"), nil
}

// GetTaskInfo retrieves detailed information about a task
func GetTaskInfo(uuid string) (*Task, error) {
	// Get task description
	desc, err := executeTask("_get", uuid+".description")
	if err != nil {
		return nil, fmt.Errorf("failed to get task description: %w", err)
	}

	// Get other task details
	project, _ := executeTask("_get", uuid+".project")
	priority, _ := executeTask("_get", uuid+".priority")
	status, _ := executeTask("_get", uuid+".status")
	due, _ := executeTask("_get", uuid+".due")

	return &Task{
		UUID:        uuid,
		Description: strings.TrimSpace(desc),
		Project:     strings.TrimSpace(project),
		Priority:    strings.TrimSpace(priority),
		Status:      strings.TrimSpace(status),
		Due:         strings.TrimSpace(due),
	}, nil
}

// ShowTaskInfo displays detailed task information using the task command
func ShowTaskInfo(uuid string) error {
	cmd := exec.Command("task", uuid, "information")
	cmd.Stdout = nil // Will be handled by the command itself
	cmd.Stderr = nil
	return cmd.Run()
}

// CreateEditCommand creates an exec.Cmd for editing a task
// This is designed to work with tea.ExecProcess for proper terminal handling
func CreateEditCommand(uuid string) *exec.Cmd {
	cmd := exec.Command("task", "rc.confirmation:no", "rc.verbose:nothing", uuid, "edit")
	return cmd
}

// EditTask opens the task for editing (legacy function for non-Bubble Tea contexts)
func EditTask(uuid string) error {
	cmd := exec.Command("task", "rc.confirmation:no", "rc.verbose:nothing", uuid, "edit")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to edit task: %w", err)
	}

	// Mark as reviewed after editing
	return MarkTaskReviewed(uuid)
}

// ModifyTask applies modifications to a task
func ModifyTask(uuid, modifications string) error {
	args := []string{"rc.confirmation:no", "rc.verbose:nothing", uuid, "modify"}
	args = append(args, strings.Fields(modifications)...)
	
	if _, err := executeTask(args...); err != nil {
		return fmt.Errorf("failed to modify task: %w", err)
	}
	return nil
}

// CompleteTask marks a task as completed
func CompleteTask(uuid string) error {
	if _, err := executeTask("rc.confirmation:no", "rc.verbose:nothing", uuid, "done"); err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}
	return nil
}

// DeleteTask deletes a task
func DeleteTask(uuid string) error {
	if _, err := executeTask("rc.confirmation:no", "rc.verbose:nothing", uuid, "delete"); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// MarkTaskReviewed marks a task as reviewed
func MarkTaskReviewed(uuid string) error {
	if _, err := executeTask("rc.confirmation:no", "rc.verbose:nothing", uuid, "modify", "reviewed:now"); err != nil {
		return fmt.Errorf("failed to mark task as reviewed: %w", err)
	}
	return nil
}

// WaitTask sets a task to waiting status with specified date and optional reason
func WaitTask(uuid, waitUntil, reason string) error {
	args := []string{"rc.confirmation:no", "rc.verbose:nothing", uuid, "modify", "wait:" + waitUntil}
	
	// Add reason as annotation if provided
	if reason != "" {
		args = append(args, "+waiting")
		// Add annotation with reason
		if _, err := executeTask("rc.confirmation:no", "rc.verbose:nothing", uuid, "annotate", "Wait reason: "+reason); err != nil {
			// Don't fail if annotation fails, just log
			fmt.Printf("Warning: Could not add wait reason annotation: %v\n", err)
		}
	} else {
		args = append(args, "+waiting")
	}
	
	if _, err := executeTask(args...); err != nil {
		return fmt.Errorf("failed to set task to waiting: %w", err)
	}
	return nil
}

// SetDueDate sets or updates the due date for a task
func SetDueDate(uuid, dueDate string) error {
	if _, err := executeTask("rc.confirmation:no", "rc.verbose:nothing", uuid, "modify", "due:"+dueDate); err != nil {
		return fmt.Errorf("failed to set due date: %w", err)
	}
	return nil
}

// GetContexts returns a list of available contexts
func GetContexts() ([]string, error) {
	output, err := executeTask("context")
	if err != nil {
		return nil, fmt.Errorf("failed to get contexts: %w", err)
	}
	
	var contexts []string
	lines := strings.Split(output, "\n")
	
	// Track unique context names
	contextMap := make(map[string]bool)
	
	// Skip header lines and parse context names
	for i, line := range lines {
		if i < 2 || strings.TrimSpace(line) == "" {
			continue
		}
		
		// Skip the usage line at the bottom
		if strings.HasPrefix(strings.TrimSpace(line), "Use 'task context") {
			continue
		}
		
		// Only process lines that start at column 0 (context names)
		// Lines starting with spaces are continuation lines for definitions
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				contextName := fields[0]
				// Only add unique context names and skip empty names
				if contextName != "" && !contextMap[contextName] {
					contextMap[contextName] = true
					contexts = append(contexts, contextName)
				}
			}
		}
	}
	
	// Add "none" option to clear context
	contexts = append(contexts, "none")
	
	return contexts, nil
}

// SetContext switches to the specified context
func SetContext(contextName string) error {
	if _, err := executeTask("context", contextName); err != nil {
		return fmt.Errorf("failed to set context: %w", err)
	}
	return nil
}

// GetCurrentContext returns the currently active context
func GetCurrentContext() (string, error) {
	output, err := executeTask("context")
	if err != nil {
		return "", fmt.Errorf("failed to get current context: %w", err)
	}
	
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "yes") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[0], nil
			}
		}
	}
	
	return "none", nil
}