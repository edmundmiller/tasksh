package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
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

	fmt.Printf("\nEnd of review. %d out of %d tasks reviewed.\n", reviewed, total)
	fmt.Println()
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
  a - AI Analysis         q - Quit review`

	fmt.Printf("\n%s\n", welcome)
}

// getReviewAction prompts the user for what to do with the current task (DEPRECATED - for huh forms)
func getReviewAction() (string, bool, error) {
	var action string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What would you like to do with this task?").
				Options(
					huh.NewOption("1. Mark as reviewed", "review"),
					huh.NewOption("2. Edit task", "edit"),
					huh.NewOption("3. Modify task", "modify"),
					huh.NewOption("4. AI Analysis", "ai_analyze"),
					huh.NewOption("5. Complete task", "complete"),
					huh.NewOption("6. Delete task", "delete"),
					huh.NewOption("7. Wait task", "wait"),
					huh.NewOption("8. Skip task", "skip"),
					huh.NewOption("9. Quit review", "quit"),
				).
				Value(&action),
		),
	)

	err := form.Run()
	if err != nil {
		return "", false, err
	}

	return action, action != "quit", nil
}

// processAction handles the selected action (DEPRECATED - keeping for AI integration reference)
func processAction(action, uuid string, task *Task) (string, error) {
	switch action {
	case "review":
		if err := markTaskReviewed(uuid); err != nil {
			return "", err
		}
		fmt.Println("Marked as reviewed.")
		return "advance_reviewed", nil

	case "edit":
		if err := editTask(uuid); err != nil {
			return "", err
		}
		fmt.Println("Task updated.")
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
		fmt.Println("Task modified.")
		return "repeat", nil

	case "ai_analyze":
		if err := processAIAnalysis(uuid, task); err != nil {
			fmt.Printf("AI analysis failed: %v\n", err)
		}
		return "repeat", nil

	case "complete":
		if err := completeTaskWithTracking(uuid, task); err != nil {
			return "", err
		}
		fmt.Println("Task completed.")
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
		fmt.Println("Task deleted.")
		return "advance", nil

	case "wait":
		waitUntil, reason, err := getWaitDetails()
		if err != nil {
			return "", err
		}
		if waitUntil == "" {
			return "repeat", nil
		}
		
		if err := waitTask(uuid, waitUntil, reason); err != nil {
			return "", err
		}
		fmt.Printf("Task set to wait until %s.\n", waitUntil)
		return "advance", nil

	case "skip":
		fmt.Println("Task skipped.")
		return "skip", nil

	case "quit":
		return "quit", nil

	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

// ModificationOption represents a suggestion for task modification
type ModificationOption struct {
	Name        string
	Value       string
	Description string
}

// getModifications prompts for task modifications with completion suggestions
func getModifications() (string, error) {
	// Gather completion data
	projects, _ := getProjects()
	tags, _ := getTags()
	priorities := getPriorities()
	common := getCommonModifications()
	
	// Build suggestion options
	var suggestions []ModificationOption
	
	// Add common modifications
	for _, mod := range common {
		suggestions = append(suggestions, ModificationOption{
			Name:        mod,
			Value:       mod,
			Description: "Common modification",
		})
	}
	
	// Add projects
	for _, project := range projects {
		suggestions = append(suggestions, ModificationOption{
			Name:        "project:" + project,
			Value:       "project:" + project,
			Description: "Set project",
		})
	}
	
	// Add tags
	for _, tag := range tags {
		suggestions = append(suggestions, ModificationOption{
			Name:        tag,
			Value:       tag,
			Description: "Add tag",
		})
		// Also add removal option
		if strings.HasPrefix(tag, "+") {
			removeTag := "-" + tag[1:]
			suggestions = append(suggestions, ModificationOption{
				Name:        removeTag,
				Value:       removeTag,
				Description: "Remove tag",
			})
		}
	}
	
	// Add priorities
	for _, priority := range priorities {
		desc := "Set priority"
		if priority == "priority:" {
			desc = "Clear priority"
		}
		suggestions = append(suggestions, ModificationOption{
			Name:        priority,
			Value:       priority,
			Description: desc,
		})
	}
	
	var choice string
	var useCustom bool
	
	// First, ask if they want to use suggestions or enter custom
	form1 := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[bool]().
				Title("How would you like to modify this task?").
				Options(
					huh.NewOption("Choose from suggestions", false),
					huh.NewOption("Enter custom modification", true),
				).
				Value(&useCustom),
		),
	)
	
	if err := form1.Run(); err != nil {
		return "", err
	}
	
	if useCustom {
		// Use custom input
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
	} else {
		// Use suggestions
		if len(suggestions) == 0 {
			return getModificationsCustom()
		}
		
		// Build Huh options from suggestions
		var options []huh.Option[string]
		for _, suggestion := range suggestions {
			options = append(options, huh.NewOption(
				fmt.Sprintf("%s - %s", suggestion.Name, suggestion.Description),
				suggestion.Value,
			))
		}
		
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select a modification:").
					Options(options...).
					Value(&choice),
			),
		)
		
		err := form.Run()
		if err != nil {
			return "", err
		}
		
		return choice, nil
	}
}

// getModificationsCustom is a fallback for custom input
func getModificationsCustom() (string, error) {
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

// getWaitDetails prompts for wait date and optional reason
func getWaitDetails() (string, string, error) {
	waitPeriods := getWaitPeriods()
	
	var useCustomDate bool
	var waitUntil string
	var reason string
	
	// First, ask if they want to use common periods or custom date
	form1 := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[bool]().
				Title("How would you like to specify the wait period?").
				Options(
					huh.NewOption("Choose from common periods", false),
					huh.NewOption("Enter custom date", true),
				).
				Value(&useCustomDate),
		),
	)
	
	if err := form1.Run(); err != nil {
		return "", "", err
	}
	
	if useCustomDate {
		// Custom date input
		form2 := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter wait until date").
					Description("Examples: tomorrow, next week, 2024-12-25, monday").
					Placeholder("Enter date...").
					Value(&waitUntil),
			),
		)
		
		if err := form2.Run(); err != nil {
			return "", "", err
		}
	} else {
		// Choose from common periods
		var options []huh.Option[string]
		for _, period := range waitPeriods {
			options = append(options, huh.NewOption(period, period))
		}
		
		form2 := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select wait period:").
					Options(options...).
					Value(&waitUntil),
			),
		)
		
		if err := form2.Run(); err != nil {
			return "", "", err
		}
	}
	
	// Optional reason
	form3 := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Wait reason (optional)").
				Description("Briefly describe why this task is waiting").
				Placeholder("Enter reason...").
				Value(&reason),
		),
	)
	
	if err := form3.Run(); err != nil {
		return "", "", err
	}
	
	return strings.TrimSpace(waitUntil), strings.TrimSpace(reason), nil
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

// processAIAnalysis handles AI analysis workflow
func processAIAnalysis(uuid string, task *Task) error {
	// Initialize time database and AI analyzer
	timeDB, err := NewTimeDB()
	if err != nil {
		return fmt.Errorf("failed to initialize time database: %w", err)
	}
	defer timeDB.Close()

	analyzer := NewAIAnalyzer(timeDB)

	fmt.Println("\n🤖 Analyzing task with AI...")
	
	// Perform AI analysis
	analysis, err := analyzer.AnalyzeTask(task)
	if err != nil {
		return fmt.Errorf("AI analysis failed: %w", err)
	}

	// Display the analysis
	fmt.Print(analyzer.FormatAnalysis(analysis))

	// Ask if user wants to apply suggestions
	if len(analysis.Suggestions) == 0 {
		fmt.Println("No suggestions to apply.")
		return nil
	}

	var wantSuggestions bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Would you like to see AI modification suggestions?").
				Affirmative("Yes").
				Negative("No").
				Value(&wantSuggestions),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if !wantSuggestions {
		return nil
	}

	// Get AI-generated modification suggestions
	aiSuggestions := analyzer.GetModificationSuggestions(analysis)
	if len(aiSuggestions) == 0 {
		fmt.Println("No actionable modifications available.")
		return nil
	}

	// Present suggestions alongside existing modification options
	return processAISuggestions(uuid, aiSuggestions)
}

// processAISuggestions handles the application of AI suggestions
func processAISuggestions(uuid string, aiSuggestions []ModificationOption) error {
	// Build Huh options from AI suggestions
	var options []huh.Option[string]
	for _, suggestion := range aiSuggestions {
		options = append(options, huh.NewOption(
			fmt.Sprintf("%s - %s", suggestion.Name, suggestion.Description),
			suggestion.Value,
		))
	}
	
	// Add option to skip
	options = append(options, huh.NewOption("Cancel - Don't apply any suggestions", ""))

	var choice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select an AI suggestion to apply:").
				Options(options...).
				Value(&choice),
		),
	)

	err := form.Run()
	if err != nil {
		return err
	}

	if choice == "" {
		return nil
	}

	// Apply the selected modification
	if err := modifyTask(uuid, choice); err != nil {
		return fmt.Errorf("failed to apply AI suggestion: %w", err)
	}

	fmt.Printf("✅ Applied AI suggestion: %s\n", choice)
	fmt.Println()
	return nil
}

// completeTaskWithTracking completes a task and optionally records time tracking data
func completeTaskWithTracking(uuid string, task *Task) error {
	// First complete the task using the existing function
	if err := completeTask(uuid); err != nil {
		return err
	}

	// Ask if user wants to record time tracking data
	var recordTime bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Would you like to record time tracking data for this completed task?").
				Description("This helps improve AI suggestions for similar tasks").
				Affirmative("Yes").
				Negative("No").
				Value(&recordTime),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if !recordTime {
		return nil
	}

	// Get time estimates
	estimatedHours, actualHours, err := getTimeTrackingData(task)
	if err != nil {
		fmt.Printf("Warning: Could not record time data: %v\n", err)
		return nil
	}

	// Record in time database
	timeDB, err := NewTimeDB()
	if err != nil {
		fmt.Printf("Warning: Could not access time database: %v\n", err)
		return nil
	}
	defer timeDB.Close()

	if err := timeDB.RecordCompletion(task, estimatedHours, actualHours); err != nil {
		fmt.Printf("Warning: Could not record completion data: %v\n", err)
	} else {
		fmt.Printf("⏱️  Time tracking data recorded (estimated: %.1fh, actual: %.1fh)\n", estimatedHours, actualHours)
	}

	return nil
}

// getTimeTrackingData prompts user for estimated and actual time spent
func getTimeTrackingData(task *Task) (float64, float64, error) {
	var estimatedStr, actualStr string

	// Get estimated time
	form1 := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("How long did you originally estimate this task would take?").
				Description("Enter time in hours (e.g., 2.5, 0.25 for 15 minutes)").
				Placeholder("2.0").
				Value(&estimatedStr),
		),
	)

	if err := form1.Run(); err != nil {
		return 0, 0, err
	}

	// Get actual time
	form2 := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("How long did this task actually take?").
				Description("Enter time in hours (e.g., 3.2, 0.5 for 30 minutes)").
				Placeholder("2.5").
				Value(&actualStr),
		),
	)

	if err := form2.Run(); err != nil {
		return 0, 0, err
	}

	// Parse the time values
	estimated := parseTimeInput(estimatedStr)
	actual := parseTimeInput(actualStr)

	return estimated, actual, nil
}

// parseTimeInput converts user time input to hours (handles common formats)
func parseTimeInput(input string) float64 {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0
	}

	// Try parsing as float first
	if hours, err := strconv.ParseFloat(input, 64); err == nil {
		return hours
	}

	// Handle common time formats like "2h30m", "90m", "1.5h"
	input = strings.ToLower(input)
	
	// Simple patterns
	if strings.HasSuffix(input, "h") {
		if hours, err := strconv.ParseFloat(strings.TrimSuffix(input, "h"), 64); err == nil {
			return hours
		}
	}
	
	if strings.HasSuffix(input, "m") {
		if minutes, err := strconv.ParseFloat(strings.TrimSuffix(input, "m"), 64); err == nil {
			return minutes / 60.0
		}
	}

	// Default to 0 if can't parse
	return 0
}