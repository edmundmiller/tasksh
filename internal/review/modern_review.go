package review

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/tui/components"
	"github.com/emiller/tasksh/internal/tui/theme"
)

// RunModern runs the modern component-based review interface
func RunModern(limit int) error {
	// Load tasks
	fmt.Print("Loading tasks for review...")
	
	uuids, err := taskwarrior.GetTasksForReview()
	if err != nil {
		return fmt.Errorf("failed to get tasks for review: %w", err)
	}
	
	if len(uuids) == 0 {
		fmt.Println("\rNo tasks need review.")
		return nil
	}
	
	// Apply limit if specified
	if limit > 0 && limit < len(uuids) {
		uuids = uuids[:limit]
	}
	
	// Load task data with progress
	var tasks []taskwarrior.Task
	for i, uuid := range uuids {
		fmt.Printf("\rLoading tasks for review... %d/%d", i+1, len(uuids))
		
		task, err := taskwarrior.GetTaskInfo(uuid)
		if err != nil {
			fmt.Printf("\rWarning: Failed to load task %s: %v\n", uuid, err)
			continue
		}
		
		tasks = append(tasks, *task)
	}
	
	fmt.Print("\r                                        \r")
	
	if len(tasks) == 0 {
		fmt.Println("No valid tasks found for review.")
		return nil
	}
	
	// Create and run the modern review interface
	model := NewModernReviewModel()
	model.SetTasks(tasks)
	
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	
	return err
}

// ModernReviewModel is the new component-based review interface
type ModernReviewModel struct {
	*components.BaseComponent
	
	// Core data
	tasks         []taskwarrior.Task
	currentIndex  int
	reviewedCount int
	
	// Components
	taskList      *components.TaskList
	progressBar   *components.ProgressBar
	helpSystem    *components.HelpSystem
	loadingIndicator *components.LoadingIndicator
	
	// State
	mode          ModernReviewMode
	loading       bool
	message       string
	error         error
	showTaskList  bool
	
	// Theme
	theme         *theme.Theme
}

// ModernReviewMode represents the current review mode
type ModernReviewMode int

const (
	ModernReviewModeTask ModernReviewMode = iota
	ModernReviewModeList
	ModernReviewModeHelp
	ModernReviewModeLoading
)

// NewModernReviewModel creates a new modern review model
func NewModernReviewModel() *ModernReviewModel {
	base := components.NewBaseComponent()
	t := theme.GetTheme()
	
	// Create components
	taskList := components.NewTaskList()
	progressBar := components.NewProgressBar()
	helpSystem := components.CreateContextualHelp("review")
	loadingIndicator := components.NewLoadingIndicator()
	
	// Configure progress bar
	progressBar.SetLabel("Review Progress")
	progressBar.SetShowPercent(true)
	progressBar.SetShowNumbers(true)
	progressBar.SetAnimated(true)
	
	// Configure loading indicator
	loadingIndicator.SetMessage("Loading tasks...")
	loadingIndicator.SetMode(components.LoadingModeSpinner)
	
	return &ModernReviewModel{
		BaseComponent:    base,
		tasks:            []taskwarrior.Task{},
		currentIndex:     0,
		reviewedCount:    0,
		taskList:         taskList,
		progressBar:      progressBar,
		helpSystem:       helpSystem,
		loadingIndicator: loadingIndicator,
		mode:             ModernReviewModeTask,
		loading:          false,
		showTaskList:     false,
		theme:            t,
	}
}

// SetTasks updates the task list
func (mrm *ModernReviewModel) SetTasks(tasks []taskwarrior.Task) {
	mrm.tasks = tasks
	mrm.taskList.SetTasks(tasks)
	mrm.updateProgress()
	mrm.loading = false
}

// updateProgress updates the progress bar
func (mrm *ModernReviewModel) updateProgress() {
	total := float64(len(mrm.tasks))
	current := float64(mrm.reviewedCount)
	mrm.progressBar.SetProgress(current, total)
}

// getCurrentTask returns the current task
func (mrm *ModernReviewModel) getCurrentTask() *taskwarrior.Task {
	if mrm.currentIndex >= 0 && mrm.currentIndex < len(mrm.tasks) {
		return &mrm.tasks[mrm.currentIndex]
	}
	return nil
}

// nextTask moves to the next task
func (mrm *ModernReviewModel) nextTask() {
	if mrm.currentIndex < len(mrm.tasks)-1 {
		mrm.currentIndex++
		mrm.taskList.SetSelected(mrm.currentIndex)
	}
}

// prevTask moves to the previous task
func (mrm *ModernReviewModel) prevTask() {
	if mrm.currentIndex > 0 {
		mrm.currentIndex--
		mrm.taskList.SetSelected(mrm.currentIndex)
	}
}

// Init implements tea.Model
func (mrm *ModernReviewModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	
	// Initialize base component
	if cmd := mrm.BaseComponent.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	// Initialize components
	if cmd := mrm.taskList.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	if cmd := mrm.progressBar.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	if cmd := mrm.helpSystem.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	return tea.Batch(cmds...)
}

// Update implements tea.Model
func (mrm *ModernReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Handle base component updates
	_, cmd := mrm.BaseComponent.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mrm.SetSize(msg.Width, msg.Height)
		mrm.taskList.SetSize(msg.Width-4, msg.Height-8)
		mrm.progressBar.SetSize(msg.Width-4, 3)
		mrm.helpSystem.SetSize(msg.Width, msg.Height)
		
	case tea.KeyMsg:
		if mrm.mode == ModernReviewModeHelp {
			_, cmd := mrm.helpSystem.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			if !mrm.helpSystem.IsVisible() {
				mrm.mode = ModernReviewModeTask
			}
		} else {
			cmds = append(cmds, mrm.handleKeyInput(msg)...)
		}
		
	case components.SelectionChangeMsg:
		if msg.Component == mrm.taskList {
			mrm.currentIndex = msg.Index
		}
		
	case components.ErrorMsg:
		mrm.error = msg.Error
		mrm.message = msg.Error.Error()
		
	case TaskActionMsg:
		mrm.handleTaskAction(msg)
	}
	
	// Update components
	if mrm.showTaskList || mrm.mode == ModernReviewModeList {
		_, cmd = mrm.taskList.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	
	_, cmd = mrm.progressBar.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	if mrm.helpSystem.IsVisible() {
		_, cmd = mrm.helpSystem.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	
	return mrm, tea.Batch(cmds...)
}

// handleKeyInput handles key input
func (mrm *ModernReviewModel) handleKeyInput(msg tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd
	
	switch msg.String() {
	case "q", "ctrl+c":
		cmds = append(cmds, tea.Quit)
		
	case "?":
		mrm.helpSystem.Toggle()
		if mrm.helpSystem.IsVisible() {
			mrm.mode = ModernReviewModeHelp
		}
		
	case "tab":
		mrm.showTaskList = !mrm.showTaskList
		if mrm.showTaskList {
			mrm.mode = ModernReviewModeList
		} else {
			mrm.mode = ModernReviewModeTask
		}
		
	case "j", "down":
		if mrm.showTaskList {
			mrm.taskList.SelectNext()
		} else {
			mrm.nextTask()
		}
		
	case "k", "up":
		if mrm.showTaskList {
			mrm.taskList.SelectPrev()
		} else {
			mrm.prevTask()
		}
		
	case "d":
		if cmd := mrm.markTaskDone(); cmd != nil {
			cmds = append(cmds, cmd)
		}
		
	case "x":
		if cmd := mrm.deleteTask(); cmd != nil {
			cmds = append(cmds, cmd)
		}
		
	case "s":
		mrm.nextTask()
		mrm.message = "Task skipped"
		
	case "enter":
		if mrm.showTaskList {
			if task := mrm.taskList.SelectedTask(); task != nil {
				// Find the task index in our list
				for i, t := range mrm.tasks {
					if t.UUID == task.UUID {
						mrm.currentIndex = i
						break
					}
				}
				mrm.showTaskList = false
				mrm.mode = ModernReviewModeTask
			}
		}
	}
	
	return cmds
}

// markTaskDone marks the current task as done
func (mrm *ModernReviewModel) markTaskDone() tea.Cmd {
	task := mrm.getCurrentTask()
	if task == nil {
		return nil
	}
	
	return func() tea.Msg {
		err := taskwarrior.CompleteTask(task.UUID)
		if err != nil {
			return components.ErrorMsg{
				Component: mrm,
				Error:     fmt.Errorf("failed to mark task done: %w", err),
			}
		}
		
		return TaskActionMsg{
			Action: "done",
			TaskID: task.UUID,
		}
	}
}

// deleteTask deletes the current task
func (mrm *ModernReviewModel) deleteTask() tea.Cmd {
	task := mrm.getCurrentTask()
	if task == nil {
		return nil
	}
	
	return func() tea.Msg {
		err := taskwarrior.DeleteTask(task.UUID)
		if err != nil {
			return components.ErrorMsg{
				Component: mrm,
				Error:     fmt.Errorf("failed to delete task: %w", err),
			}
		}
		
		return TaskActionMsg{
			Action: "delete",
			TaskID: task.UUID,
		}
	}
}

// TaskActionMsg represents a task action result
type TaskActionMsg struct {
	Action string
	TaskID string
	Error  error
}

// handleTaskAction handles the result of a task action
func (mrm *ModernReviewModel) handleTaskAction(msg TaskActionMsg) {
	if msg.Error != nil {
		mrm.error = msg.Error
		mrm.message = msg.Error.Error()
		return
	}
	
	// Remove the task from the list
	for i, task := range mrm.tasks {
		if task.UUID == msg.TaskID {
			mrm.tasks = append(mrm.tasks[:i], mrm.tasks[i+1:]...)
			break
		}
	}
	
	// Update task list
	mrm.taskList.SetTasks(mrm.tasks)
	
	// Update progress
	mrm.reviewedCount++
	mrm.updateProgress()
	
	// Adjust current index
	if mrm.currentIndex >= len(mrm.tasks) && len(mrm.tasks) > 0 {
		mrm.currentIndex = len(mrm.tasks) - 1
		mrm.taskList.SetSelected(mrm.currentIndex)
	}
	
	// Check if we're done
	if len(mrm.tasks) == 0 {
		mrm.message = "All tasks reviewed! 🎉"
	} else {
		mrm.message = fmt.Sprintf("Task %s: %s", msg.Action, msg.TaskID[:8])
	}
}

// View implements tea.Model
func (mrm *ModernReviewModel) View() string {
	width, height := mrm.Size()
	if width == 0 || height == 0 {
		return ""
	}
	
	switch mrm.mode {
	case ModernReviewModeLoading:
		return mrm.renderLoading()
	case ModernReviewModeHelp:
		return mrm.renderWithHelp()
	default:
		return mrm.renderMain()
	}
}

// renderLoading renders the loading state
func (mrm *ModernReviewModel) renderLoading() string {
	loading := mrm.loadingIndicator.View()
	
	width, height := mrm.Size()
	style := mrm.theme.Styles.Container.
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center)
		
	return style.Render(loading)
}

// renderMain renders the main interface
func (mrm *ModernReviewModel) renderMain() string {
	if len(mrm.tasks) == 0 {
		return mrm.theme.Styles.Success.Render("All tasks reviewed! 🎉\n\nPress q to quit.")
	}
	
	var sections []string
	
	// Progress bar
	progress := mrm.progressBar.View()
	sections = append(sections, progress)
	
	if mrm.showTaskList {
		// Task list view
		taskListView := mrm.taskList.View()
		sections = append(sections, taskListView)
	} else {
		// Single task view
		task := mrm.getCurrentTask()
		if task != nil {
			taskView := mrm.renderCurrentTask(*task)
			sections = append(sections, taskView)
		}
	}
	
	// Status message
	if mrm.message != "" {
		style := mrm.theme.Styles.Info
		if mrm.error != nil {
			style = mrm.theme.Styles.Error
		}
		sections = append(sections, style.Render(mrm.message))
	}
	
	// Footer
	footer := mrm.renderFooter()
	sections = append(sections, footer)
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderCurrentTask renders the current task details
func (mrm *ModernReviewModel) renderCurrentTask(task taskwarrior.Task) string {
	var sections []string
	
	// Task counter
	counter := fmt.Sprintf("Task %d of %d", mrm.currentIndex+1, len(mrm.tasks))
	sections = append(sections, mrm.theme.Styles.Caption.Render(counter))
	
	// Task description
	desc := mrm.theme.Styles.Title.Render(task.Description)
	sections = append(sections, desc)
	
	// Task details
	var details []string
	
	if task.Project != "" {
		project := fmt.Sprintf("Project: %s", task.Project)
		details = append(details, mrm.theme.Styles.Body.Render(project))
	}
	
	if task.Priority != "" {
		var priorityStyle lipgloss.Style
		switch task.Priority {
		case "H":
			priorityStyle = lipgloss.NewStyle().Foreground(mrm.theme.Colors.PriorityHigh)
		case "M":
			priorityStyle = lipgloss.NewStyle().Foreground(mrm.theme.Colors.PriorityMedium)
		case "L":
			priorityStyle = lipgloss.NewStyle().Foreground(mrm.theme.Colors.PriorityLow)
		}
		priority := fmt.Sprintf("Priority: %s", task.Priority)
		details = append(details, priorityStyle.Render(priority))
	}
	
	if task.Due != "" {
		due := fmt.Sprintf("Due: %s", task.Due)
		details = append(details, mrm.theme.Styles.Warning.Render(due))
	}
	
	if len(task.Tags) > 0 {
		tags := fmt.Sprintf("Tags: +%s", strings.Join(task.Tags, " +"))
		details = append(details, mrm.theme.Styles.Caption.Render(tags))
	}
	
	if len(details) > 0 {
		sections = append(sections, lipgloss.JoinVertical(lipgloss.Left, details...))
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderFooter renders the footer with key bindings
func (mrm *ModernReviewModel) renderFooter() string {
	var bindings []string
	
	if mrm.showTaskList {
		bindings = []string{
			"↑/k: up",
			"↓/j: down", 
			"enter: select",
			"tab: task view",
			"?: help",
			"q: quit",
		}
	} else {
		bindings = []string{
			"d: done",
			"x: delete",
			"s: skip",
			"↑/k: prev",
			"↓/j: next",
			"tab: list view",
			"?: help",
			"q: quit",
		}
	}
	
	footer := strings.Join(bindings, " • ")
	return mrm.theme.Styles.Caption.Render(footer)
}

// renderWithHelp renders the interface with help overlay
func (mrm *ModernReviewModel) renderWithHelp() string {
	main := mrm.renderMain()
	help := mrm.helpSystem.View()
	
	// Simple overlay - help takes precedence
	if help != "" {
		return help
	}
	
	return main
}

// Implement Component interface
var _ components.Component = (*ModernReviewModel)(nil)