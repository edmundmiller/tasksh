package review

import (
	"fmt"
	"os"
	"strconv"

	"github.com/emiller/tasksh/internal/taskwarrior"
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the interactive task review process
func Run(limit int) error {
	// Ensure review configuration is set up
	if err := taskwarrior.EnsureReviewConfig(); err != nil {
		return fmt.Errorf("failed to configure review: %w", err)
	}

	// Get lazy loading configuration
	lazyLoadThreshold := 100 // default
	if val := os.Getenv("TASKSH_LAZY_LOAD_THRESHOLD"); val != "" {
		if threshold, err := strconv.Atoi(val); err == nil && threshold > 0 {
			lazyLoadThreshold = threshold
		}
	}

	// Check if we should use lazy loading
	uuids, err := taskwarrior.GetTasksForReview()
	if err != nil {
		return fmt.Errorf("failed to get tasks for review: %w", err)
	}

	totalTasks := len(uuids)
	
	// If we have many tasks, use lazy loading
	if totalTasks > lazyLoadThreshold && limit == 0 {
		fmt.Printf("Found %d tasks. Loading first %d for immediate review...\n", totalTasks, lazyLoadThreshold)
		return runBubbleTeaReviewLazy(uuids, lazyLoadThreshold)
	}

	// Otherwise, use regular batch loading
	fmt.Print("Loading tasks for review...")
	
	// Try the new batch approach first with progress
	tasks, err := taskwarrior.GetTasksForReviewWithDataProgress(func(loaded, total int) {
		fmt.Printf("\rLoading tasks for review... %d/%d", loaded, total)
	})
	
	// Clear the loading message
	fmt.Print("\r                                        \r")
	
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
	uuids2, err2 := taskwarrior.GetTasksForReview()
	if err2 != nil {
		return fmt.Errorf("failed to get tasks for review: %w", err2)
	}
	uuids = uuids2

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

// runBubbleTeaReviewLazy runs the review interface with lazy loading
func runBubbleTeaReviewLazy(allUUIDs []string, initialLoad int) error {
	// Show welcome message
	showWelcomeMessage()

	// Load first batch of tasks
	firstBatch := allUUIDs
	if len(allUUIDs) > initialLoad {
		firstBatch = allUUIDs[:initialLoad]
	}

	fmt.Printf("Loading first %d tasks...\n", len(firstBatch))
	
	// Load initial batch with progress
	initialTasks, err := taskwarrior.GetTasksWithDataProgress(firstBatch, func(loaded, total int) {
		fmt.Printf("\rLoading tasks... %d/%d", loaded, total)
	})
	fmt.Print("\r                                        \r")
	
	if err != nil {
		return fmt.Errorf("failed to load initial tasks: %w", err)
	}

	// Create and initialize the review model
	model := NewReviewModel()
	
	// Set up initial tasks
	var uuids []string
	taskMap := make(map[string]*taskwarrior.Task)
	
	for _, td := range initialTasks {
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
	
	// Set all UUIDs but only loaded data for first batch
	model.SetTasks(allUUIDs, len(allUUIDs))
	model.taskCache = taskMap
	model.lazyLoadEnabled = true
	model.loadedTasks = len(firstBatch)
	model.totalTasks = len(allUUIDs)
	
	// Load the first task from cache
	if len(uuids) > 0 {
		model.currentTask = taskMap[uuids[0]]
		model.updateViewport()
	}

	// Create the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run review interface: %w", err)
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