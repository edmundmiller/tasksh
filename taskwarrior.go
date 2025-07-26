package main

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

// checkTaskwarrior verifies that the task command is available
func checkTaskwarrior() error {
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

// ensureReviewConfig sets up the required UDA and report for review
func ensureReviewConfig() error {
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

// getTasksForReview returns a list of task UUIDs that need review
func getTasksForReview() ([]string, error) {
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

// getTaskInfo retrieves detailed information about a task
func getTaskInfo(uuid string) (*Task, error) {
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

// showTaskInfo displays detailed task information using the task command
func showTaskInfo(uuid string) error {
	cmd := exec.Command("task", uuid, "information")
	cmd.Stdout = nil // Will be handled by the command itself
	cmd.Stderr = nil
	return cmd.Run()
}

// editTask opens the task for editing
func editTask(uuid string) error {
	cmd := exec.Command("task", "rc.confirmation:no", "rc.verbose:nothing", uuid, "edit")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to edit task: %w", err)
	}

	// Mark as reviewed after editing
	return markTaskReviewed(uuid)
}

// modifyTask applies modifications to a task
func modifyTask(uuid, modifications string) error {
	args := []string{"rc.confirmation:no", "rc.verbose:nothing", uuid, "modify"}
	args = append(args, strings.Fields(modifications)...)
	
	if _, err := executeTask(args...); err != nil {
		return fmt.Errorf("failed to modify task: %w", err)
	}
	return nil
}

// completeTask marks a task as completed
func completeTask(uuid string) error {
	if _, err := executeTask("rc.confirmation:no", "rc.verbose:nothing", uuid, "done"); err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}
	return nil
}

// deleteTask deletes a task
func deleteTask(uuid string) error {
	if _, err := executeTask("rc.confirmation:no", "rc.verbose:nothing", uuid, "delete"); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// markTaskReviewed marks a task as reviewed
func markTaskReviewed(uuid string) error {
	if _, err := executeTask("rc.confirmation:no", "rc.verbose:nothing", uuid, "modify", "reviewed:now"); err != nil {
		return fmt.Errorf("failed to mark task as reviewed: %w", err)
	}
	return nil
}

// getProjects returns a list of existing projects
func getProjects() ([]string, error) {
	output, err := executeTask("rc.verbose:nothing", "_projects")
	if err != nil {
		return []string{}, nil // Return empty list if no projects
	}
	
	if output == "" {
		return []string{}, nil
	}
	
	projects := strings.Split(output, "\n")
	// Filter out empty strings
	var filtered []string
	for _, project := range projects {
		if strings.TrimSpace(project) != "" {
			filtered = append(filtered, strings.TrimSpace(project))
		}
	}
	return filtered, nil
}

// getTags returns a list of existing tags
func getTags() ([]string, error) {
	output, err := executeTask("rc.verbose:nothing", "_tags")
	if err != nil {
		return []string{}, nil // Return empty list if no tags
	}
	
	if output == "" {
		return []string{}, nil
	}
	
	tags := strings.Split(output, "\n")
	// Filter out empty strings and add + prefix for suggestions
	var filtered []string
	for _, tag := range tags {
		if strings.TrimSpace(tag) != "" {
			filtered = append(filtered, "+"+strings.TrimSpace(tag))
		}
	}
	return filtered, nil
}

// getPriorities returns available priority levels
func getPriorities() []string {
	return []string{"priority:H", "priority:M", "priority:L", "priority:"}
}

// getCommonModifications returns common modification patterns
func getCommonModifications() []string {
	return []string{
		"due:tomorrow",
		"due:next week", 
		"due:next month",
		"due:",
		"wait:tomorrow",
		"wait:next week",
		"wait:",
		"scheduled:tomorrow",
		"scheduled:next week",
		"scheduled:",
		"depends:",
	}
}

// waitTask sets a task to waiting status with specified date and optional reason
func waitTask(uuid, waitUntil, reason string) error {
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

// getWaitPeriods returns common wait periods for autocompletion
func getWaitPeriods() []string {
	return []string{
		"tomorrow",
		"next week",
		"next month",
		"1week",
		"2weeks", 
		"1month",
		"3months",
		"monday",
		"friday",
		"january",
		"next year",
	}
}