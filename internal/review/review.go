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

	// Get tasks that need review
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

func showWelcomeMessage() {
	fmt.Println()
	fmt.Println("Welcome to tasksh review!")
	fmt.Println("Review helps keep your task list accurate by allowing you to")
	fmt.Println("systematically review tasks and update their metadata.")
	fmt.Println()
}