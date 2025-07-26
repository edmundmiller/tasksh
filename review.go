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
		fmt.Println("\nThere are no tasks needing review.")
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
				Title("What would you like to do with this task?").
				Options(
					huh.NewOption("1. Mark as reviewed", "review"),
					huh.NewOption("2. Edit task", "edit"),
					huh.NewOption("3. Modify task", "modify"),
					huh.NewOption("4. Complete task", "complete"),
					huh.NewOption("5. Delete task", "delete"),
					huh.NewOption("6. Wait task", "wait"),
					huh.NewOption("7. Skip task", "skip"),
					huh.NewOption("8. Quit review", "quit"),
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

// processAction handles the selected action
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

	case "complete":
		if err := completeTask(uuid); err != nil {
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