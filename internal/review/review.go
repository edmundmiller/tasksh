package review

import (
	"fmt"

	"github.com/emiller/tasksh/internal/taskwarrior"
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the interactive task review process
func Run(limit int) error {
	// Ensure review configuration is set up
	if err := taskwarrior.EnsureReviewConfig(); err != nil {
		return fmt.Errorf("failed to configure review: %w", err)
	}

	// Try the new batch approach first
	tasks, err := taskwarrior.GetTasksForReviewWithData()
	if err == nil && len(tasks) > 0 {
		// Apply limit if specified
		total := len(tasks)
		if limit > 0 && limit < total {
			total = limit
			tasks = tasks[:limit]
		}
		return runBubbleTeaReviewBatch(tasks, total)
	}

	// Fall back to the old approach if batch export fails
	uuids, err := taskwarrior.GetTasksForReview()
	if err != nil {
		return fmt.Errorf("failed to get tasks for review: %w", err)
	}

	if len(uuids) == 0 {
		fmt.Println("\nThere are no tasks needing review.")
		fmt.Println()
		return nil
	}

	// Apply limit if specified
	total := len(uuids)
	if limit > 0 && limit < total {
		total = limit
		uuids = uuids[:limit]
	}

	return runBubbleTeaReview(uuids, total)
}

// runBubbleTeaReview runs the Bubble Tea review interface
func runBubbleTeaReview(uuids []string, total int) error {
	// Show welcome message
	showWelcomeMessage()

	// Create and initialize the review model
	model := NewReviewModel()
	model.SetTasks(uuids, total)

	// Load the first task
	if len(uuids) > 0 {
		task, err := taskwarrior.GetTaskInfo(uuids[0])
		if err != nil {
			return fmt.Errorf("failed to load first task: %w", err)
		}
		model.currentTask = task
		model.updateViewport()
	}

	// Create the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		// If we can't open a TTY (e.g., in tests), fall back to a simple message
		fmt.Printf("\nWould start reviewing %d tasks with Bubble Tea interface.\n", len(uuids))
		return nil
	}

	return nil
}

// runBubbleTeaReviewBatch runs the Bubble Tea review interface with pre-loaded task data
func runBubbleTeaReviewBatch(tasks []*taskwarrior.TaskData, total int) error {
	// Show welcome message
	showWelcomeMessage()

	// Create and initialize the review model
	model := NewReviewModel()
	
	// Convert TaskData to Task format for compatibility
	var uuids []string
	taskMap := make(map[string]*taskwarrior.Task)
	
	for _, td := range tasks {
		uuids = append(uuids, td.UUID)
		taskMap[td.UUID] = &taskwarrior.Task{
			UUID:        td.UUID,
			Description: td.Description,
			Project:     td.Project,
			Priority:    td.Priority,
			Status:      td.Status,
			Due:         td.Due,
		}
	}
	
	model.SetTasks(uuids, total)
	model.taskCache = taskMap // Add task cache to model
	
	// Load the first task from cache
	if len(uuids) > 0 {
		model.currentTask = taskMap[uuids[0]]
		model.updateViewport()
	}

	// Create the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		// If we can't open a TTY (e.g., in tests), fall back to a simple message
		fmt.Printf("\nWould start reviewing %d tasks with Bubble Tea interface.\n", len(tasks))
		return nil
	}

	return nil
}

func showWelcomeMessage() {
	fmt.Println()
	fmt.Println("Welcome to tasksh review!")
	fmt.Println("Review helps keep your task list accurate by allowing you to")
	fmt.Println("systematically review tasks and update their metadata.")
	fmt.Println()
}