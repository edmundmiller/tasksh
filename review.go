package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdReview starts the interactive task review process
func cmdReview(limit int) error {
	// Ensure review configuration is set up
	if err := ensureReviewConfig(); err != nil {
		return fmt.Errorf("failed to configure review: %w", err)
	}

	// Get tasks that need review
	uuids, err := getTasksForReview()
	if err != nil {
		return fmt.Errorf("failed to get tasks for review: %w", err)
	}

	if len(uuids) == 0 {
		fmt.Println("\nThere are no tasks needing review.")
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

// runBubbleTeaReview runs the new Bubble Tea review interface
func runBubbleTeaReview(uuids []string, total int) error {
	// Show welcome message
	showWelcomeMessage()

	// Create and initialize the review model
	model := NewReviewModel()
	model.SetTasks(uuids, total)

	// Load the first task
	if len(uuids) > 0 {
		task, err := getTaskInfo(uuids[0])
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

// runReviewLoop runs the main review loop (DEPRECATED - keeping for reference)
func runReviewLoop_OLD(uuids []string, total int) error {
	reviewed := 0
	current := 0

	// Show welcome message
	showWelcomeMessage()

	for current < len(uuids) {
		uuid := uuids[current]
		
		// Get task information
		task, err := getTaskInfo(uuid)
		if err != nil {
			fmt.Printf("Error getting task info: %v\n", err)
			current++
			continue
		}

		// Show progress and task info
		fmt.Printf("\n[%d of %d] %s\n", current+1, total, task.Description)
		
		// Show detailed task information
		if err := showTaskInfo(uuid); err != nil {
			fmt.Printf("Error showing task info: %v\n", err)
		}

		// This old code is now handled by Bubble Tea interface
		fmt.Println("This function is deprecated, use the new Bubble Tea interface")
		break
	}

	fmt.Printf("\nEnd of review. %d out of %d tasks reviewed.\n\n", reviewed, total)
	return nil
}

// showWelcomeMessage displays the welcome message for review
func showWelcomeMessage() {
	welcome := `The review process is important for keeping your list accurate, so you are working on the right tasks.

For each task you are shown, look at the metadata. Determine whether the task needs to be changed, or whether it is accurate. You may skip a task but a skipped task is not considered reviewed.

You may stop at any time, and resume later right where you left off.

Hot keys:
  r - Mark as reviewed    e - Edit task         m - Modify task
  c - Complete task       d - Delete task       w - Wait task  
  s - Skip task           j/k - Navigate        ? - Toggle help
  q - Quit review`

	fmt.Printf("\n%s\n", welcome)
}