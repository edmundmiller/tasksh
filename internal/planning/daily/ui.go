package daily

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/emiller/tasksh/internal/planning/shared"
)

// PlanningModel represents the Bubble Tea model for daily planning
type PlanningModel struct {
	session      *DailyPlanningSession
	viewport     viewport.Model
	textInput    textinput.Model
	spinner      spinner.Model
	help         help.Model
	keys         KeyMap
	
	// UI state
	selectedIndex    int
	message         string
	err             error
	quitting        bool
	width           int
	height          int
	isLoading       bool
	loadingMessage  string
	backgroundLoading bool
	tasksPreloaded   bool
	
	// Step-specific state
	reflectionText   string
	focusInput       string
	showingHelp     bool
}

// KeyMap defines key bindings for the daily planning interface
type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Select    key.Binding
	Deselect  key.Binding
	Next      key.Binding
	Previous  key.Binding
	Help      key.Binding
	Quit      key.Binding
	
	// Text input specific
	Confirm   key.Binding
	Clear     key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "select task"),
		),
		Deselect: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "deselect task"),
		),
		Next: key.NewBinding(
			key.WithKeys("n", "enter"),
			key.WithHelp("n/enter", "next step"),
		),
		Previous: key.NewBinding(
			key.WithKeys("p", "backspace"),
			key.WithHelp("p/backspace", "previous step"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Clear: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "clear"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp returns the short help text
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Next, k.Help, k.Quit}
}

// FullHelp returns the full help text
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Select, k.Deselect},
		{k.Next, k.Previous, k.Confirm, k.Clear},
		{k.Help, k.Quit},
	}
}

// NewPlanningModel creates a new daily planning model
func NewPlanningModel(session *DailyPlanningSession) *PlanningModel {
	// Create viewport
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle()

	// Create text input for reflection/focus
	ti := textinput.New()
	ti.Placeholder = "Enter your thoughts..."
	ti.CharLimit = 500
	ti.Width = 70

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Create help
	h := help.New()
	h.ShowAll = false

	model := &PlanningModel{
		session:   session,
		viewport:  vp,
		textInput: ti,
		spinner:   s,
		help:      h,
		keys:      DefaultKeyMap(),
		width:     80,
		height:    24,
	}

	return model
}

// LoadingCompleteMsg is sent when loading is complete
type LoadingCompleteMsg struct {
	err error
}

// TaskLoadingStartedMsg indicates background loading has begun
type TaskLoadingStartedMsg struct{}

// TaskLoadingProgressMsg provides optional progress updates during loading
type TaskLoadingProgressMsg struct {
	message string
	percent float64
}

// Init initializes the model
func (m *PlanningModel) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.WindowSize(),
		textinput.Blink,
		m.spinner.Tick,
	}
	
	// Start background task loading if we're on the reflection step
	if m.session.CurrentStep == StepReflection {
		m.backgroundLoading = true
		cmds = append(cmds, m.session.StartBackgroundTaskLoading())
	}
	
	return tea.Batch(cmds...)
}

// Update handles messages and updates the model
func (m *PlanningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case BackgroundLoadResult:
		// Background loading completed
		m.backgroundLoading = false
		m.tasksPreloaded = msg.Error == nil
		if msg.Error != nil {
			// Store error but don't show it yet - will show when transitioning
			m.err = msg.Error
		}
		return m, nil
		
	case LoadingCompleteMsg:
		m.isLoading = false
		m.loadingMessage = ""
		if msg.err != nil {
			m.message = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.selectedIndex = 0
			m.message = ""
			m.updateViewport()
		}
		return m, nil
		
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		headerHeight := 4
		footerHeight := 4
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight - footerHeight
		
		m.updateViewport()

	case tea.KeyMsg:
		// Handle text input mode first
		if m.needsTextInput() && m.textInput.Focused() {
			switch {
			case key.Matches(msg, m.keys.Confirm):
				m.handleTextInput()
				m.textInput.Blur()
				return m, nil
			case key.Matches(msg, m.keys.Clear):
				m.textInput.SetValue("")
				return m, nil
			case key.Matches(msg, m.keys.Quit):
				m.textInput.Blur()
				return m, nil
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}

		// Handle normal navigation
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			m.updateViewport()

		case key.Matches(msg, m.keys.Up):
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.updateViewport()
			}

		case key.Matches(msg, m.keys.Down):
			maxIndex := m.getMaxSelectableIndex()
			if m.selectedIndex < maxIndex {
				m.selectedIndex++
				m.updateViewport()
			}

		case key.Matches(msg, m.keys.Select):
			m.handleSelection()

		case key.Matches(msg, m.keys.Deselect):
			m.handleDeselection()

		case key.Matches(msg, m.keys.Next):
			if cmd := m.handleNextStep(); cmd != nil {
				return m, cmd
			}

		case key.Matches(msg, m.keys.Previous):
			m.handlePreviousStep()
		}
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	
	// Update spinner if loading
	if m.isLoading {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the daily planning interface
func (m *PlanningModel) View() string {
	if m.quitting {
		return "\nDaily planning session ended.\n\n"
	}

	if m.width == 0 || m.height == 0 {
		return ""
	}

	var sections []string

	// Header
	sections = append(sections, m.renderHeader())

	// Main content
	if m.isLoading {
		// Show loading spinner
		sections = append(sections, m.renderLoadingView())
	} else if m.needsTextInput() && m.textInput.Focused() {
		sections = append(sections, m.renderTextInputView())
	} else {
		sections = append(sections, m.viewport.View())
	}

	// Message area
	if m.message != "" {
		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).  // Bright green for better visibility
			Bold(true).
			Margin(1, 0)
		sections = append(sections, messageStyle.Render(m.message))
	}

	// Help
	sections = append(sections, m.renderFooter())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderHeader renders the header with step progress
func (m *PlanningModel) renderHeader() string {
	current, total := m.session.GetStepProgress()
	stepName := m.session.CurrentStep.String()
	
	title := fmt.Sprintf("Daily Planning - Step %d/%d: %s", current+1, total, stepName)
	date := m.session.Context.Date.Format("Monday, January 2, 2006")
	
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("6")).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width).
		Padding(0, 1)

	header := headerStyle.Render(title)
	dateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).  // Light gray instead of dark gray
		Align(lipgloss.Center).
		Width(m.width)
	dateHeader := dateStyle.Render(date)
	
	separator := strings.Repeat("━", m.width)
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))  // Light gray for better visibility

	return header + "\n" + dateHeader + "\n" + sepStyle.Render(separator)
}

// renderFooter renders the help and navigation footer
func (m *PlanningModel) renderFooter() string {
	helpSepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))  // Light gray for better contrast
	helpSep := helpSepStyle.Render(strings.Repeat("━", m.width))
	
	return helpSep + "\n" + m.help.View(m.keys)
}

// renderTextInputView renders the text input interface
func (m *PlanningModel) renderTextInputView() string {
	var content strings.Builder
	
	content.WriteString("\n")
	
	switch m.session.CurrentStep {
	case StepReflection:
		content.WriteString("Take a moment to reflect on yesterday's work:\n\n")
		prompts := m.session.GetReflectionPrompts()
		for i, prompt := range prompts {
			content.WriteString(fmt.Sprintf("%d. %s\n", i+1, prompt))
		}
		content.WriteString("\nShare your thoughts (press Enter when done):\n")
	case StepFinalization:
		content.WriteString("Set your daily focus/intention:\n\n")
		content.WriteString("What's the main theme or goal for today?\n")
		content.WriteString("This helps maintain clarity throughout your day.\n\n")
	}
	
	content.WriteString(m.textInput.View())
	content.WriteString("\n\nPress Enter to continue, Ctrl+U to clear, or q to skip")
	
	return content.String()
}

// updateViewport updates the viewport content based on current step
func (m *PlanningModel) updateViewport() {
	var content strings.Builder
	content.WriteString("\n")

	switch m.session.CurrentStep {
	case StepReflection:
		content.WriteString(m.renderReflectionStep())
	case StepTaskSelection:
		content.WriteString(m.renderTaskSelectionStep())
	case StepWorkloadAssessment:
		content.WriteString(m.renderWorkloadAssessmentStep())
	case StepFinalization:
		content.WriteString(m.renderFinalizationStep())
	case StepSummary:
		content.WriteString(m.renderSummaryStep())
	default:
		content.WriteString("Unknown step")
	}

	m.viewport.SetContent(content.String())
}

// renderReflectionStep renders the reflection step
func (m *PlanningModel) renderReflectionStep() string {
	var content strings.Builder
	
	content.WriteString("🌅 Starting your daily planning ritual\n\n")
	content.WriteString("Before diving into today's tasks, let's reflect on yesterday's work.\n")
	content.WriteString("This helps you learn from experience and set realistic expectations.\n\n")
	
	if m.session.Reflection != nil {
		completedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
		content.WriteString(completedStyle.Render("✓ Reflection completed") + "\n")
		
		detailStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
		content.WriteString(detailStyle.Render(fmt.Sprintf("Yesterday's energy level: %s", m.session.Reflection.EnergyLevel.String())) + "\n")
		if len(m.session.Reflection.Accomplishments) > 0 {
			content.WriteString(detailStyle.Render(fmt.Sprintf("Key accomplishments: %s", strings.Join(m.session.Reflection.Accomplishments, ", "))) + "\n")
		}
	} else {
		instructionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Italic(true)
		content.WriteString(instructionStyle.Render("Press 'n' to start reflection, or 'enter' to skip to task selection"))
	}
	
	return content.String()
}

// renderTaskSelectionStep renders the task selection step
func (m *PlanningModel) renderTaskSelectionStep() string {
	var content strings.Builder
	
	content.WriteString("📋 Select tasks for today\n\n")
	content.WriteString("Choose tasks that align with your energy and available time.\n")
	content.WriteString("Focus on what you can realistically accomplish.\n\n")
	
	if len(m.session.AvailableTasks) == 0 {
		content.WriteString("No tasks found for today's planning.\n")
		return content.String()
	}
	
	// Calculate layout - split screen into two columns
	leftColumnWidth := 80  // Available tasks column
	
	// Build left column (available tasks)
	var leftColumn strings.Builder
	taskHeaderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	leftColumn.WriteString(taskHeaderStyle.Render("AVAILABLE TASKS:") + "\n")
	
	for i, task := range m.session.AvailableTasks {
		// Determine if this task is selected
		isSelected := false
		for _, selectedTask := range m.session.SelectedTasks {
			if selectedTask.UUID == task.UUID {
				isSelected = true
				break
			}
		}
		
		// Build the line components
		var line strings.Builder
		
		// Selection indicator on the left
		if isSelected {
			checkStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)
			line.WriteString(checkStyle.Render("[✓]"))
		} else {
			line.WriteString("[ ]")
		}
		line.WriteString(" ")
		
		// Current selection arrow
		if i == m.selectedIndex {
			arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
			line.WriteString(arrowStyle.Render("▶"))
		} else {
			line.WriteString(" ")
		}
		line.WriteString(" ")
		
		// Task number
		numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		line.WriteString(numberStyle.Render(fmt.Sprintf("%2d.", i+1)))
		line.WriteString(" ")
		
		// Task description
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
		line.WriteString(descStyle.Render(task.Description))
		line.WriteString(" ")
		
		// Task details
		detailsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
		line.WriteString(detailsStyle.Render(fmt.Sprintf("(%.1fh, %s)", task.EstimatedHours, task.Category.String())))
		
		leftColumn.WriteString(line.String() + "\n")
	}
	
	// Build right column (selected tasks)
	var rightColumn strings.Builder
	if len(m.session.SelectedTasks) > 0 {
		selectedHeaderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
		rightColumn.WriteString(selectedHeaderStyle.Render(fmt.Sprintf("SELECTED TASKS (%d):", len(m.session.SelectedTasks))) + "\n")
		
		totalHours := 0.0
		for i, task := range m.session.SelectedTasks {
			totalHours += task.EstimatedHours
			
			// Format selected task
			numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
			descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
			hoursStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
			
			rightColumn.WriteString(fmt.Sprintf("%s %s %s\n",
				numberStyle.Render(fmt.Sprintf("%d.", i+1)),
				descStyle.Render(truncateString(task.Description, 25)),
				hoursStyle.Render(fmt.Sprintf("%.1fh", task.EstimatedHours))))
		}
		
		// Total hours
		rightColumn.WriteString("\n")
		totalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
		rightColumn.WriteString(totalStyle.Render(fmt.Sprintf("Total: %.1f hours", totalHours)))
	} else {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
		rightColumn.WriteString(emptyStyle.Render("No tasks selected yet"))
	}
	
	// Combine columns side by side
	leftLines := strings.Split(leftColumn.String(), "\n")
	rightLines := strings.Split(rightColumn.String(), "\n")
	
	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}
	
	for i := 0; i < maxLines; i++ {
		leftLine := ""
		if i < len(leftLines) {
			leftLine = leftLines[i]
		}
		// Pad left column to fixed width
		leftLine = padRight(leftLine, leftColumnWidth)
		
		rightLine := ""
		if i < len(rightLines) {
			rightLine = rightLines[i]
		}
		
		content.WriteString(leftLine + "  " + rightLine + "\n")
	}
	
	// Add instruction text
	instructionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Italic(true)
	content.WriteString("\n" + instructionStyle.Render("Use 's' to select/deselect tasks, then 'n' or 'enter' to continue") + "\n")
	
	return content.String()
}

// Helper function to truncate strings
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Helper function to pad string to fixed width
func padRight(s string, width int) string {
	// Account for ANSI color codes when calculating visual width
	visualLen := lipgloss.Width(s)
	if visualLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visualLen)
}

// renderWorkloadAssessmentStep renders the workload assessment step
func (m *PlanningModel) renderWorkloadAssessmentStep() string {
	var content strings.Builder
	
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	content.WriteString(headerStyle.Render("⚖️ Workload Assessment") + "\n\n")
	
	if len(m.session.SelectedTasks) == 0 {
		content.WriteString("No tasks selected. Go back to select tasks.\n")
		return content.String()
	}
	
	// Calculate assessment
	assessment := m.session.CalculateWorkloadAssessment()
	totalHours, breakdown, _ := shared.CalculateWorkloadSummary(m.session.SelectedTasks)
	
	// Format workload summary with better styling
	summaryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	
	content.WriteString(labelStyle.Render("Selected Tasks: ") + summaryStyle.Render(fmt.Sprintf("%d", len(m.session.SelectedTasks))) + "\n")
	content.WriteString(labelStyle.Render("Total Time: ") + summaryStyle.Render(fmt.Sprintf("%.1f hours", totalHours)) + "\n")
	content.WriteString(labelStyle.Render("Available Focus Time: ") + summaryStyle.Render(fmt.Sprintf("%.1f hours", assessment.FocusHours)) + "\n\n")
	
	// Capacity warning
	warning := m.session.GetCapacityWarning()
	content.WriteString(warning + "\n\n")
	
	// Breakdown by category with improved styling
	breakdownHeaderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	content.WriteString(breakdownHeaderStyle.Render("BREAKDOWN BY PRIORITY:") + "\n")
	
	for category, hours := range breakdown {
		if hours > 0 {
			categoryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
			hoursStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
			content.WriteString(fmt.Sprintf("  %s: %s\n", 
				categoryStyle.Render(category.String()), 
				hoursStyle.Render(fmt.Sprintf("%.1fh", hours))))
		}
	}
	
	// Add instruction with styling
	instructionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Italic(true)
	content.WriteString("\n" + instructionStyle.Render("Press 'n' to continue, 'p' to adjust task selection"))
	
	return content.String()
}

// renderFinalizationStep renders the finalization step
func (m *PlanningModel) renderFinalizationStep() string {
	var content strings.Builder
	
	content.WriteString("🎯 Finalize Your Plan\n\n")
	content.WriteString("Set a daily focus to guide your work and decision-making.\n\n")
	
	if m.session.DailyFocus != "" {
		focusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
		content.WriteString(focusStyle.Render(fmt.Sprintf("Daily Focus: %s", m.session.DailyFocus)) + "\n\n")
		
		completedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
		content.WriteString(completedStyle.Render("✓ Daily focus set") + "\n")
	} else {
		instructionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Italic(true)
		content.WriteString(instructionStyle.Render("Press 'n' to set your daily focus"))
	}
	
	return content.String()
}

// renderSummaryStep renders the final summary
func (m *PlanningModel) renderSummaryStep() string {
	var content strings.Builder
	
	content.WriteString("📊 Daily Plan Summary\n\n")
	content.WriteString(m.session.GetDailySummary())
	content.WriteString("\n\nPress 'n' to complete planning")
	
	return content.String()
}

// Helper methods

func (m *PlanningModel) needsTextInput() bool {
	switch m.session.CurrentStep {
	case StepReflection:
		return m.session.Reflection == nil
	case StepFinalization:
		return m.session.DailyFocus == ""
	}
	return false
}

func (m *PlanningModel) getMaxSelectableIndex() int {
	switch m.session.CurrentStep {
	case StepTaskSelection:
		return len(m.session.AvailableTasks) - 1
	}
	return 0
}

func (m *PlanningModel) handleTextInput() {
	value := strings.TrimSpace(m.textInput.Value())
	
	switch m.session.CurrentStep {
	case StepReflection:
		// Create basic reflection data
		reflection := &shared.ReflectionData{
			Date:        m.session.Context.Date.AddDate(0, 0, -1),
			EnergyLevel: shared.EnergyMedium, // Default, could be improved
			Accomplishments: []string{value},
		}
		m.session.SetReflectionData(reflection)
		m.message = "Reflection saved"
	case StepFinalization:
		m.session.SetDailyFocus(value)
		m.message = "Daily focus set"
	}
	
	m.textInput.SetValue("")
	m.updateViewport()
}

func (m *PlanningModel) handleSelection() {
	if m.session.CurrentStep != StepTaskSelection {
		return
	}
	
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.session.AvailableTasks) {
		err := m.session.AddTaskToSelection(m.selectedIndex)
		if err != nil {
			m.message = fmt.Sprintf("Error: %v", err)
		} else {
			m.message = "Task added to plan"
		}
		m.updateViewport()
	}
}

func (m *PlanningModel) handleDeselection() {
	if m.session.CurrentStep != StepTaskSelection {
		return
	}
	
	task := m.session.AvailableTasks[m.selectedIndex]
	for i, selected := range m.session.SelectedTasks {
		if selected.UUID == task.UUID {
			m.session.RemoveTaskFromSelection(i)
			m.message = "Task removed from plan"
			m.updateViewport()
			return
		}
	}
}

func (m *PlanningModel) handleNextStep() tea.Cmd {
	// Special handling for reflection step
	if m.session.CurrentStep == StepReflection {
		// Check if tasks are already preloaded via session
		if m.session.IsTasksLoaded() {
			// Tasks already loaded, proceed immediately
			err := m.session.NextStep()
			if err != nil {
				m.message = fmt.Sprintf("Error: %v", err)
			} else {
				m.selectedIndex = 0
				m.message = ""
				m.updateViewport()
			}
			return nil
		} else if m.backgroundLoading {
			// Still loading in background, show spinner
			m.isLoading = true
			m.loadingMessage = "Finishing task loading..."
			
			// Return a command that waits for loading to complete
			return func() tea.Msg {
				// Wait a bit for loading to complete
				time.Sleep(100 * time.Millisecond)
				err := m.session.NextStep()
				return LoadingCompleteMsg{err: err}
			}
		} else {
			// Not loaded and not loading - start loading now
			m.isLoading = true
			m.loadingMessage = "Loading tasks and calculating estimates..."
			
			return func() tea.Msg {
				err := m.session.NextStep()
				return LoadingCompleteMsg{err: err}
			}
		}
	}
	
	// Save text input if needed
	if m.needsTextInput() && m.textInput.Focused() {
		m.handleTextInput()
		m.textInput.Blur()
		return nil
	}
	
	if m.needsTextInput() {
		m.textInput.Focus()
		return nil
	}
	
	// For other steps, advance normally
	err := m.session.NextStep()
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
	} else {
		m.selectedIndex = 0
		m.message = ""
		m.updateViewport()
	}
	return nil
}

func (m *PlanningModel) handlePreviousStep() {
	err := m.session.PreviousStep()
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
	} else {
		m.selectedIndex = 0
		m.message = ""
		m.updateViewport()
	}
}

// renderLoadingView renders the loading screen
func (m *PlanningModel) renderLoadingView() string {
	// Calculate center position
	contentHeight := 5
	paddingTop := (m.viewport.Height - contentHeight) / 2
	if paddingTop < 0 {
		paddingTop = 0
	}
	
	// Create loading content
	var content strings.Builder
	
	// Add padding
	for i := 0; i < paddingTop; i++ {
		content.WriteString("\n")
	}
	
	// Spinner and message
	spinnerLine := fmt.Sprintf("%s %s", m.spinner.View(), m.loadingMessage)
	spinnerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)
	
	// Center the spinner line
	centeredSpinner := lipgloss.PlaceHorizontal(m.viewport.Width, lipgloss.Center, spinnerStyle.Render(spinnerLine))
	content.WriteString(centeredSpinner + "\n\n")
	
	// Add context
	contextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	
	contextLines := []string{
		"This may include:",
		"• Syncing timewarrior data (if needed)",
		"• Calculating time estimates",
		"• Analyzing task priorities",
	}
	
	for _, line := range contextLines {
		centered := lipgloss.PlaceHorizontal(m.viewport.Width, lipgloss.Center, contextStyle.Render(line))
		content.WriteString(centered + "\n")
	}
	
	return content.String()
}

// Run starts the daily planning interface
func Run(targetDate time.Time) error {
	session, err := NewDailyPlanningSession(targetDate)
	if err != nil {
		return fmt.Errorf("failed to create daily planning session: %w", err)
	}
	defer session.Close()

	model := NewPlanningModel(session)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run daily planning interface: %w", err)
	}

	return nil
}