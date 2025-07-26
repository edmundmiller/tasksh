package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
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
		fmt.Println("\nThere are no tasks needing review.\n")
		return nil
	}

	// Apply limit if specified
	total := len(uuids)
	if limit > 0 && limit < total {
		total = limit
		uuids = uuids[:limit]
	}

	return runReviewLoop(uuids, total)
}

// runReviewLoop runs the main review loop
func runReviewLoop(uuids []string, total int) error {
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

		// Get user action
		action, shouldContinue, err := getReviewAction()
		if err != nil {
			return fmt.Errorf("review interrupted: %w", err)
		}

		if !shouldContinue {
			break
		}

		// Process the action
		actionResult, err := processAction(action, uuid, task)
		if err != nil {
			fmt.Printf("Error processing action: %v\n", err)
			continue
		}

		// Handle action result
		switch actionResult {
		case "advance":
			current++
			reviewed++
		case "advance_reviewed":
			current++
			reviewed++
		case "skip":
			current++
		case "repeat":
			// Stay on same task
		case "quit":
			goto done
		}

		// Clear screen if desired (can be made configurable)
		fmt.Print("\033[2J\033[0;0H")
	}

done:
	fmt.Printf("\nEnd of review. %d out of %d tasks reviewed.\n\n", reviewed, total)
	return nil
}

// showWelcomeMessage displays the welcome message for review
func showWelcomeMessage() {
	welcome := `The review process is important for keeping your list accurate, so you are working on the right tasks.

For each task you are shown, look at the metadata. Determine whether the task needs to be changed, or whether it is accurate. You may skip a task but a skipped task is not considered reviewed.

You may stop at any time, and resume later right where you left off.`

	fmt.Printf("\n%s\n", welcome)
}

// getReviewAction prompts the user for what to do with the current task
func getReviewAction() (string, bool, error) {
	var action string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What would you like to do with this task? (Press / to filter)").
				Options(
					huh.NewOption("1. Mark as reviewed", "review"),
					huh.NewOption("2. Edit task", "edit"),
					huh.NewOption("3. Modify task", "modify"),
					huh.NewOption("4. Complete task", "complete"),
					huh.NewOption("5. Delete task", "delete"),
					huh.NewOption("6. Skip task", "skip"),
					huh.NewOption("7. Quit review", "quit"),
				).
				Filtering(true).
				Value(&action),
		),
	)

	err := form.Run()
	if err != nil {
		return "", false, err
	}

	return action, action != "quit", nil
}

// processAction handles the selected action
func processAction(action, uuid string, task *Task) (string, error) {
	switch action {
	case "review":
		if err := markTaskReviewed(uuid); err != nil {
			return "", err
		}
		fmt.Println("Marked as reviewed.\n")
		return "advance_reviewed", nil

	case "edit":
		if err := editTask(uuid); err != nil {
			return "", err
		}
		fmt.Println("Task updated.\n")
		return "advance_reviewed", nil

	case "modify":
		modifications, err := getModifications()
		if err != nil {
			return "", err
		}
		if modifications == "" {
			return "repeat", nil
		}
		
		if err := modifyTask(uuid, modifications); err != nil {
			return "", err
		}
		fmt.Println("Task modified.\n")
		return "repeat", nil

	case "complete":
		if err := completeTask(uuid); err != nil {
			return "", err
		}
		fmt.Println("Task completed.\n")
		return "advance", nil

	case "delete":
		confirmed, err := confirmDelete(task)
		if err != nil {
			return "", err
		}
		if !confirmed {
			return "repeat", nil
		}
		
		if err := deleteTask(uuid); err != nil {
			return "", err
		}
		fmt.Println("Task deleted.\n")
		return "advance", nil

	case "skip":
		fmt.Println("Task skipped.\n")
		return "skip", nil

	case "quit":
		return "quit", nil

	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

// getModifications prompts for task modifications
func getModifications() (string, error) {
	var modifications string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter modification arguments").
				Description("Examples: +tag -tag /old/new/ project:newproject priority:H").
				Placeholder("Enter modifications...").
				Value(&modifications),
		),
	)

	err := form.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(modifications), nil
}

// confirmDelete asks for confirmation before deleting a task
func confirmDelete(task *Task) (bool, error) {
	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Delete task: %s", task.Description)).
				Description("This action cannot be undone.").
				Affirmative("Delete").
				Negative("Cancel").
				Value(&confirmed),
		),
	)

	err := form.Run()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}