package review

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
	"github.com/maaslalani/confetty/confetti"

	"github.com/emiller/tasksh/internal/ai"
	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/timedb"
)

// ReviewMode represents the current mode of the review interface
type ReviewMode int

const (
	ModeViewing ReviewMode = iota
	ModeConfirmDelete
	ModeWaiting
	ModeModifying
	ModeInputWaitDate
	ModeInputWaitReason
	ModeInputModification
	ModeWaitCalendar
	ModeDueCalendar
	ModeInputDueDate
	ModeCelebrating
	ModeContextSelect
	ModeAIAnalysis
	ModeAILoading
)

// ReviewModel represents the state of the Bubble Tea review interface
type ReviewModel struct {
	// Task review state
	tasks       []string // UUIDs of tasks to review
	current     int      // Current task index
	currentTask *taskwarrior.Task    // Current task details
	total       int      // Total tasks to review
	reviewed    int      // Number reviewed

	// UI components
	viewport   viewport.Model
	help       help.Model
	textInput  textinput.Model
	calendar   CalendarModel
	completion *CompletionModel
	confetti   tea.Model
	keys       KeyMap

	// Application state
	mode     ReviewMode
	err      error
	quitting bool
	width    int
	height   int
	modeJustChanged bool // Prevents input from processing mode-triggering key

	// For confirmations and input
	pendingAction string
	message       string
	waitDate      string
	waitReason    string
	
	// Completion state
	selectedSuggestion int
	
	// Celebration state
	celebrationStart time.Time
	
	// Context state
	contexts          []string
	selectedContext   int
	currentContext    string
	
	// AI analysis state
	aiAnalyzer      *ai.Analyzer
	currentAnalysis *ai.TaskAnalysis
	aiSpinner       spinner.Model
}

// KeyMap defines the key bindings for the review interface
type KeyMap struct {
	// Navigation
	NextTask key.Binding
	PrevTask key.Binding
	
	// Actions
	Review   key.Binding
	Edit     key.Binding
	Modify   key.Binding
	Complete key.Binding
	Delete   key.Binding
	Wait     key.Binding
	Due      key.Binding
	Skip     key.Binding
	Context  key.Binding
	AIAnalysis key.Binding
	
	// General
	Help key.Binding
	Quit key.Binding
	
	// Calendar
	ToggleCalendar key.Binding
	
	// Confirmations
	Confirm key.Binding
	Cancel  key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		NextTask: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "next task"),
		),
		PrevTask: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "previous task"),
		),
		
		// Actions
		Review: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "mark reviewed"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit task"),
		),
		Modify: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "modify task"),
		),
		Complete: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "complete task"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete task"),
		),
		Wait: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "wait task"),
		),
		Due: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "due date"),
		),
		Skip: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "skip task"),
		),
		Context: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "switch context"),
		),
		AIAnalysis: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "AI analysis"),
		),
		
		// General
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		
		// Calendar
		ToggleCalendar: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "toggle calendar"),
		),
		
		// Confirmations
		Confirm: key.NewBinding(
			key.WithKeys("y", "enter"),
			key.WithHelp("y/enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n", "esc"),
			key.WithHelp("n/esc", "cancel"),
		),
	}
}

// ShortHelp returns the short help text
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Review, k.Edit, k.Modify, k.Complete, k.Delete, k.Wait, k.Due, k.Skip, k.Context, k.AIAnalysis, k.Help, k.Quit}
}

// FullHelp returns the full help text
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextTask, k.PrevTask},
		{k.Review, k.Edit, k.Modify},
		{k.Complete, k.Delete, k.Wait, k.Due, k.Skip},
		{k.Context, k.AIAnalysis, k.Help, k.Quit},
	}
}

// NewReviewModel creates a new review model
func NewReviewModel() *ReviewModel {
	// Create viewport
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).  // ANSI bright black (gray)
		PaddingRight(2)

	// Create help model
	h := help.New()
	h.ShowAll = false
	
	// Style the help for better visibility
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))   // ANSI cyan for keys
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))  // ANSI white for descriptions
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // ANSI bright black for separators
	h.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))    // ANSI cyan for keys
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))   // ANSI white for descriptions
	h.Styles.FullSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))   // ANSI bright black for separators

	// Create text input
	ti := textinput.New()
	ti.Focus()

	// Create calendar
	cal := NewCalendarModel()
	cal.SetFocused(false) // Start unfocused

	// Create completion model
	completion := NewCompletionModel()
	completion.LoadDynamicData() // Load projects and tags

	// Create confetti model
	confettiModel := confetti.InitialModel()

	// Create AI analyzer
	var aiAnalyzer *ai.Analyzer
	if timeDB, err := timedb.New(); err == nil {
		aiAnalyzer = ai.NewAnalyzer(timeDB)
		// Note: TimeDB will be closed when the model is cleaned up
		// For now, we'll leave it open during the review session
	}

	// Create AI loading spinner
	aiSpinner := spinner.New()
	aiSpinner.Spinner = spinner.Dot
	aiSpinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // Magenta for AI

	model := &ReviewModel{
		viewport:   vp,
		help:       h,
		textInput:  ti,
		calendar:   cal,
		completion: completion,
		confetti:   confettiModel,
		keys:       DefaultKeyMap(),
		mode:       ModeViewing,
		aiAnalyzer: aiAnalyzer,
		aiSpinner:  aiSpinner,
	}

	// Initialize current context
	if currentContext, err := taskwarrior.GetCurrentContext(); err == nil {
		model.currentContext = currentContext
	} else {
		model.currentContext = "none"
	}

	return model
}

// Init initializes the review model
func (m *ReviewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *ReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update viewport dimensions
		headerHeight := 3 // Status bar
		footerHeight := 4 // Help text
		verticalMarginHeight := headerHeight + footerHeight
		
		if !m.help.ShowAll {
			footerHeight = 2
			verticalMarginHeight = headerHeight + footerHeight
		}
		
		m.viewport.Width = msg.Width - 2
		m.viewport.Height = msg.Height - verticalMarginHeight
		
		// Forward window size to confetti
		m.confetti, cmd = m.confetti.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		// Handle special input modes
		switch m.mode {
		case ModeConfirmDelete:
			return m.updateConfirmDelete(msg)
		case ModeInputModification:
			return m.updateModificationInput(msg)
		case ModeInputWaitDate:
			return m.updateWaitDateInput(msg)
		case ModeInputWaitReason:
			return m.updateWaitReasonInput(msg)
		case ModeWaitCalendar:
			return m.updateWaitCalendar(msg)
		case ModeDueCalendar:
			return m.updateDueCalendar(msg)
		case ModeInputDueDate:
			return m.updateDueDateInput(msg)
		case ModeContextSelect:
			return m.updateContextSelect(msg)
		case ModeAIAnalysis:
			return m.updateAIAnalysis(msg)
		case ModeAILoading:
			return m.updateAILoading(msg)
		}
		
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, m.keys.NextTask):
			if m.current < len(m.tasks)-1 {
				m.current++
				return m, m.loadCurrentTask()
			}

		case key.Matches(msg, m.keys.PrevTask):
			if m.current > 0 {
				m.current--
				return m, m.loadCurrentTask()
			}

		case key.Matches(msg, m.keys.Review):
			return m, m.reviewCurrentTask()

		case key.Matches(msg, m.keys.Edit):
			return m, m.editCurrentTask()

		case key.Matches(msg, m.keys.Complete):
			return m, m.completeCurrentTask()

		case key.Matches(msg, m.keys.Delete):
			m.mode = ModeConfirmDelete
			m.message = "Delete this task? This action cannot be undone."

		case key.Matches(msg, m.keys.Modify):
			m.mode = ModeInputModification
			m.textInput.Placeholder = "Enter modification (e.g., +tag, project:new, priority:H)"
			m.textInput.SetValue("")
			m.textInput.Focus()
			m.message = "Enter modification:"
			m.modeJustChanged = true
			m.selectedSuggestion = 0
			// Initialize completion suggestions
			m.completion.UpdateSuggestions("", 0)

		case key.Matches(msg, m.keys.Wait):
			m.mode = ModeWaitCalendar
			m.calendar.SetFocused(true)
			m.message = "Select wait date (Tab to toggle text input):"

		case key.Matches(msg, m.keys.Due):
			m.mode = ModeDueCalendar
			m.calendar.SetFocused(true)
			m.message = "Select due date (Tab to toggle text input):"

		case key.Matches(msg, m.keys.Skip):
			return m, m.skipCurrentTask()

		case key.Matches(msg, m.keys.Context):
			return m, m.initContextSelect()
			
		case key.Matches(msg, m.keys.AIAnalysis):
			if m.aiAnalyzer != nil {
				m.mode = ModeAILoading
				m.message = "Analyzing task with AI..."
				// Start the spinner with a tick command
				tickCmd := func() tea.Msg {
					return m.aiSpinner.Tick()
				}
				return m, tea.Batch(tickCmd, m.analyzeCurrentTask())
			} else {
				m.message = "AI analysis not available (OpenAI API key or timedb unavailable)"
			}
		}

	case taskLoadedMsg:
		m.currentTask = msg.task
		m.updateViewport()

	case actionCompletedMsg:
		m.message = msg.message
		m.reviewed++
		
		// Move to next task if not at the end
		if m.current < len(m.tasks)-1 {
			m.current++
			return m, m.loadCurrentTask()
		} else {
			// Review complete - start celebration!
			m.mode = ModeCelebrating
			m.celebrationStart = time.Now()
			m.message = fmt.Sprintf("🎉 Review complete! %d of %d tasks reviewed. 🎉", m.reviewed, len(m.tasks))
			// Initialize confetti with window size and start celebration timer
			confettiCmd := m.confetti.Init()
			sizeCmd := func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
			celebrationCmd := func() tea.Msg {
				time.Sleep(3 * time.Second) // Show confetti for 3 seconds
				return celebrationCompleteMsg{}
			}
			return m, tea.Batch(confettiCmd, tea.Cmd(sizeCmd), tea.Cmd(celebrationCmd))
		}

	case taskSkippedMsg:
		m.message = msg.message
		// Don't increment reviewed count for skipped tasks
		
		// Move to next task if not at the end
		if m.current < len(m.tasks)-1 {
			m.current++
			return m, m.loadCurrentTask()
		} else {
			// Review complete - start celebration!
			m.mode = ModeCelebrating
			m.celebrationStart = time.Now()
			m.message = fmt.Sprintf("🎉 Review complete! %d of %d tasks reviewed. 🎉", m.reviewed, len(m.tasks))
			// Initialize confetti with window size and start celebration timer
			confettiCmd := m.confetti.Init()
			sizeCmd := func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
			celebrationCmd := func() tea.Msg {
				time.Sleep(3 * time.Second) // Show confetti for 3 seconds
				return celebrationCompleteMsg{}
			}
			return m, tea.Batch(confettiCmd, tea.Cmd(sizeCmd), tea.Cmd(celebrationCmd))
		}

	case celebrationCompleteMsg:
		// Celebration finished, now quit gracefully
		m.quitting = true
		return m, tea.Quit

	case errorMsg:
		m.err = msg.error
		m.message = fmt.Sprintf("Error: %v", msg.error)

	case contextsLoadedMsg:
		m.contexts = msg.contexts
		m.mode = ModeContextSelect
		m.selectedContext = 0
		m.message = "Select context (↑↓: navigate, Enter: select, ESC: cancel):"

	case contextChangedMsg:
		m.mode = ModeViewing
		m.currentContext = msg.context
		m.message = fmt.Sprintf("Context switched to: %s", msg.context)
		
	case aiAnalysisCompleteMsg:
		m.currentAnalysis = msg.analysis
		m.mode = ModeAIAnalysis
		m.message = "AI analysis complete (ESC to return to task view)"
	}

	// Update components based on mode
	if m.mode == ModeInputModification || m.mode == ModeInputWaitDate || m.mode == ModeInputWaitReason || m.mode == ModeInputDueDate {
		// Don't process the triggering key if mode just changed
		if !m.modeJustChanged {
			m.textInput, cmd = m.textInput.Update(msg)
			cmds = append(cmds, cmd)
		}
		m.modeJustChanged = false
	} else if m.mode == ModeWaitCalendar || m.mode == ModeDueCalendar {
		m.calendar, cmd = m.calendar.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.mode == ModeCelebrating {
		m.confetti, cmd = m.confetti.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.mode == ModeAILoading {
		m.aiSpinner, cmd = m.aiSpinner.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// updateConfirmDelete handles delete confirmation
func (m *ReviewModel) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		m.mode = ModeViewing
		m.message = ""
		return m, m.deleteCurrentTask()

	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeViewing
		m.message = ""
	}
	
	return m, nil
}

// updateModificationInput handles modification input
func (m *ReviewModel) updateModificationInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg.Type {
	case tea.KeyEnter:
		modification := strings.TrimSpace(m.textInput.Value())
		if modification == "" {
			m.mode = ModeViewing
			m.message = ""
			return m, nil
		}
		m.mode = ModeViewing
		m.message = ""
		return m, m.modifyCurrentTask(modification)
		
	case tea.KeyEscape:
		m.mode = ModeViewing
		m.message = ""
		return m, nil
		
	case tea.KeyTab:
		// Apply selected completion suggestion if available
		suggestions := m.completion.GetSuggestions()
		if len(suggestions) > 0 && m.selectedSuggestion < len(suggestions) {
			newText := m.completion.GetCompletionText(suggestions[m.selectedSuggestion], m.textInput.Value())
			m.textInput.SetValue(newText)
			m.textInput.CursorEnd()
			// Reset selection and update suggestions
			m.selectedSuggestion = 0
			m.completion.UpdateSuggestions(m.textInput.Value(), m.textInput.Position())
		}
		return m, nil
		
	case tea.KeyUp:
		// Navigate up in completion suggestions
		suggestions := m.completion.GetSuggestions()
		if len(suggestions) > 0 {
			m.selectedSuggestion--
			if m.selectedSuggestion < 0 {
				m.selectedSuggestion = len(suggestions) - 1
			}
		}
		return m, nil
		
	case tea.KeyDown:
		// Navigate down in completion suggestions
		suggestions := m.completion.GetSuggestions()
		if len(suggestions) > 0 {
			m.selectedSuggestion++
			if m.selectedSuggestion >= len(suggestions) {
				m.selectedSuggestion = 0
			}
		}
		return m, nil
	}
	
	// Update the text input with the keystroke
	m.textInput, cmd = m.textInput.Update(msg)
	
	// Update completion suggestions based on new input and reset selection
	m.completion.UpdateSuggestions(m.textInput.Value(), m.textInput.Position())
	m.selectedSuggestion = 0
	
	return m, cmd
}

// updateWaitDateInput handles wait date input
func (m *ReviewModel) updateWaitDateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg.Type {
	case tea.KeyEnter:
		waitDate := strings.TrimSpace(m.textInput.Value())
		if waitDate == "" {
			m.mode = ModeViewing
			m.message = ""
			return m, nil
		}
		m.waitDate = waitDate
		m.mode = ModeInputWaitReason
		m.textInput.Placeholder = "Enter wait reason (optional)"
		m.textInput.SetValue("")
		m.message = "Enter wait reason (optional):"
		
	case tea.KeyEscape:
		m.mode = ModeViewing
		m.message = ""
		return m, nil
		
	case tea.KeyTab:
		// Switch back to calendar mode
		m.mode = ModeWaitCalendar
		m.calendar.SetFocused(true)
		m.message = "Select wait date (Tab to toggle text input):"
		return m, nil
	}
	
	// Update the text input with the keystroke
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// updateWaitReasonInput handles wait reason input
func (m *ReviewModel) updateWaitReasonInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg.Type {
	case tea.KeyEnter:
		waitReason := strings.TrimSpace(m.textInput.Value())
		m.mode = ModeViewing
		m.message = ""
		return m, m.waitCurrentTask(m.waitDate, waitReason)
		
	case tea.KeyEscape:
		m.mode = ModeViewing
		m.message = ""
		return m, nil
	}
	
	// Update the text input with the keystroke
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// updateWaitCalendar handles calendar input for wait date selection
func (m *ReviewModel) updateWaitCalendar(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.ToggleCalendar):
		// Switch to text input mode
		m.mode = ModeInputWaitDate
		m.calendar.SetFocused(false)
		m.textInput.Placeholder = "Enter wait date (e.g., tomorrow, next week, 2024-12-25)"
		m.textInput.SetValue("")
		m.message = "Enter wait date (Tab to toggle calendar):"
		return m, nil
		
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeViewing
		m.calendar.SetFocused(false)
		m.message = ""
		return m, nil
		
	case key.Matches(msg, m.keys.Confirm):
		// Select the date from calendar
		waitDate := m.calendar.GetSelectedDateString()
		m.waitDate = waitDate
		m.mode = ModeInputWaitReason
		m.calendar.SetFocused(false)
		m.textInput.Placeholder = "Enter wait reason (optional)"
		m.textInput.SetValue("")
		m.message = "Enter wait reason (optional):"
		return m, nil
	}
	
	// Forward unhandled keys to calendar for navigation
	var cmd tea.Cmd
	m.calendar, cmd = m.calendar.Update(msg)
	return m, cmd
}

// updateDueCalendar handles calendar input for due date selection
func (m *ReviewModel) updateDueCalendar(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.ToggleCalendar):
		// Switch to text input mode
		m.mode = ModeInputDueDate
		m.calendar.SetFocused(false)
		m.textInput.Placeholder = "Enter due date (e.g., tomorrow, next week, 2024-12-25)"
		m.textInput.SetValue("")
		m.message = "Enter due date (Tab to toggle calendar):"
		return m, nil
		
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeViewing
		m.calendar.SetFocused(false)
		m.message = ""
		return m, nil
		
	case key.Matches(msg, m.keys.Confirm):
		// Select the date from calendar
		dueDate := m.calendar.GetSelectedDateString()
		m.mode = ModeViewing
		m.calendar.SetFocused(false)
		m.message = ""
		return m, m.dueCurrentTask(dueDate)
	}
	
	// Forward unhandled keys to calendar for navigation
	var cmd tea.Cmd
	m.calendar, cmd = m.calendar.Update(msg)
	return m, cmd
}

// updateDueDateInput handles due date text input
func (m *ReviewModel) updateDueDateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg.Type {
	case tea.KeyEnter:
		dueDate := strings.TrimSpace(m.textInput.Value())
		if dueDate == "" {
			m.mode = ModeViewing
			m.message = ""
			return m, nil
		}
		m.mode = ModeViewing
		m.message = ""
		return m, m.dueCurrentTask(dueDate)
		
	case tea.KeyEscape:
		m.mode = ModeViewing
		m.message = ""
		return m, nil
		
	case tea.KeyTab:
		// Switch back to calendar mode
		m.mode = ModeDueCalendar
		m.calendar.SetFocused(true)
		m.message = "Select due date (Tab to toggle text input):"
		return m, nil
	}
	
	// Update the text input with the keystroke
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View renders the review interface
func (m *ReviewModel) View() string {
	if m.quitting {
		return fmt.Sprintf("\nEnd of review. %d out of %d tasks reviewed.\n\n", m.reviewed, m.total)
	}

	var sections []string

	// Status bar
	statusBar := m.renderStatusBar()
	sections = append(sections, statusBar)

	// Main content area
	if m.mode == ModeConfirmDelete {
		sections = append(sections, m.renderConfirmation())
	} else if m.mode == ModeInputModification || m.mode == ModeInputWaitDate || m.mode == ModeInputWaitReason || m.mode == ModeInputDueDate {
		sections = append(sections, m.renderInput())
	} else if m.mode == ModeWaitCalendar || m.mode == ModeDueCalendar {
		sections = append(sections, m.renderCalendar())
	} else if m.mode == ModeContextSelect {
		sections = append(sections, m.renderContextSelect())
	} else if m.mode == ModeAIAnalysis {
		sections = append(sections, m.renderAIAnalysis())
	} else if m.mode == ModeAILoading {
		sections = append(sections, m.renderAILoading())
	} else if m.mode == ModeCelebrating {
		sections = append(sections, m.confetti.View())
	} else {
		sections = append(sections, m.viewport.View())
	}

	// Message area
	if m.message != "" {
		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).  // ANSI green
			Bold(true).
			Margin(1, 0)
		sections = append(sections, messageStyle.Render(m.message))
	}

	// Help
	sections = append(sections, m.help.View(m.keys))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderStatusBar renders the top status bar
func (m *ReviewModel) renderStatusBar() string {
	progress := fmt.Sprintf("[%d of %d]", m.current+1, m.total)
	
	var taskTitle string
	if m.currentTask != nil {
		taskTitle = m.currentTask.Description
		if len(taskTitle) > 60 {
			taskTitle = taskTitle[:57] + "..."
		}
	}

	statusStyle := lipgloss.NewStyle().
		Reverse(true).  // Use terminal's reverse video for better compatibility
		Padding(0, 1).
		Width(m.width)

	left := progress + " " + taskTitle
	return statusStyle.Render(left)
}

// renderConfirmation renders delete confirmation dialog
func (m *ReviewModel) renderConfirmation() string {
	confirmStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1")).  // ANSI red for delete warning
		Padding(1, 2).
		Margin(2, 4)

	content := fmt.Sprintf("%s\n\n%s", 
		m.message,
		"Press 'y' to confirm, 'n' to cancel")

	return confirmStyle.Render(content)
}

// renderInput renders input dialog
func (m *ReviewModel) renderInput() string {
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).  // ANSI blue for input dialogs
		Padding(1, 2).
		Margin(2, 4)

	content := ""
	if m.message != "" {
		content = fmt.Sprintf("%s\n\n%s", m.message, m.textInput.View())
	} else {
		content = m.textInput.View()
	}

	// Add completion suggestions for modify mode
	if m.mode == ModeInputModification {
		suggestions := m.completion.GetSuggestions()
		if len(suggestions) > 0 {
			content += "\n\n" + m.renderCompletionSuggestions(suggestions)
		}
		content += "\n\n↑↓: navigate  Tab: complete  ESC: cancel"
	} else {
		content += "\n\nPress ESC to cancel"
	}

	return inputStyle.Render(content)
}

// renderCompletionSuggestions renders the completion suggestions popup
func (m *ReviewModel) renderCompletionSuggestions(suggestions []CompletionItem) string {
	if len(suggestions) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Bold(true).
		Render("Suggestions:"))

	for i, suggestion := range suggestions {
		prefix := "  "
		if i == m.selectedSuggestion {
			prefix = "▶ " // Highlight selected suggestion
		}
		
		// Style based on completion type
		var style lipgloss.Style
		switch suggestion.Type {
		case CompletionAttribute:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
		case CompletionProject:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
		case CompletionTag:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // Magenta
		case CompletionPriority:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // Red
		case CompletionStatus:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("4")) // Blue
		default:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // White
		}
		
		line := fmt.Sprintf("%s%s", prefix, style.Render(suggestion.Text))
		if suggestion.Description != "" {
			line += lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Render(" - " + suggestion.Description)
		}
		
		lines = append(lines, line)
		
		// Limit visible suggestions
		if i >= 7 {
			break
		}
	}

	return strings.Join(lines, "\n")
}

// renderCalendar renders the calendar component
func (m *ReviewModel) renderCalendar() string {
	calendarStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("6")). // ANSI blue for calendar
		Padding(1, 2).
		Margin(2, 4)

	return calendarStyle.Render(m.calendar.View())
}

// updateViewport updates the viewport content with current task info
func (m *ReviewModel) updateViewport() {
	if m.currentTask == nil {
		m.viewport.SetContent("Loading task...")
		return
	}

	var content strings.Builder
	
	// Use minimal styling - let the terminal handle the colors
	labelStyle := lipgloss.NewStyle().Bold(true)
	
	// Task description
	content.WriteString(labelStyle.Render("Description:"))
	content.WriteString("\n")
	content.WriteString(m.currentTask.Description)
	content.WriteString("\n\n")

	// Task details
	if m.currentTask.Project != "" {
		content.WriteString(labelStyle.Render("Project: "))
		content.WriteString(m.currentTask.Project)
		content.WriteString("\n")
	}
	if m.currentTask.Priority != "" {
		content.WriteString(labelStyle.Render("Priority: "))
		content.WriteString(m.currentTask.Priority)
		content.WriteString("\n")
	}
	if m.currentTask.Status != "" {
		content.WriteString(labelStyle.Render("Status: "))
		content.WriteString(m.currentTask.Status)
		content.WriteString("\n")
	}
	if m.currentTask.Due != "" {
		content.WriteString(labelStyle.Render("Due: "))
		content.WriteString(m.currentTask.Due)
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(labelStyle.Render("UUID: "))
	content.WriteString(m.currentTask.UUID)

	m.viewport.SetContent(content.String())
}

// Command messages
type taskLoadedMsg struct {
	task *taskwarrior.Task
}

type actionCompletedMsg struct {
	message string
}

type taskSkippedMsg struct {
	message string
}

type celebrationCompleteMsg struct{}

type errorMsg struct {
	error error
}

type contextsLoadedMsg struct {
	contexts []string
}

type contextChangedMsg struct {
	context string
}

type aiAnalysisCompleteMsg struct {
	analysis *ai.TaskAnalysis
}

// Commands for async operations
func (m *ReviewModel) loadCurrentTask() tea.Cmd {
	return func() tea.Msg {
		if m.current >= len(m.tasks) {
			return errorMsg{fmt.Errorf("task index out of range")}
		}
		
		task, err := taskwarrior.GetTaskInfo(m.tasks[m.current])
		if err != nil {
			return errorMsg{err}
		}
		
		return taskLoadedMsg{task: task}
	}
}

func (m *ReviewModel) reviewCurrentTask() tea.Cmd {
	return func() tea.Msg {
		if err := taskwarrior.MarkTaskReviewed(m.tasks[m.current]); err != nil {
			return errorMsg{err}
		}
		return actionCompletedMsg{message: "Marked as reviewed."}
	}
}

func (m *ReviewModel) editCurrentTask() tea.Cmd {
	return tea.ExecProcess(taskwarrior.CreateEditCommand(m.tasks[m.current]), func(err error) tea.Msg {
		if err != nil {
			return errorMsg{err}
		}
		
		// Mark task as reviewed after successful edit
		if err := taskwarrior.MarkTaskReviewed(m.tasks[m.current]); err != nil {
			return errorMsg{fmt.Errorf("edit succeeded but failed to mark as reviewed: %w", err)}
		}
		
		return actionCompletedMsg{message: "Task updated."}
	})
}

func (m *ReviewModel) completeCurrentTask() tea.Cmd {
	return func() tea.Msg {
		if err := taskwarrior.CompleteTask(m.tasks[m.current]); err != nil {
			return errorMsg{err}
		}
		return actionCompletedMsg{message: "Task completed."}
	}
}

func (m *ReviewModel) deleteCurrentTask() tea.Cmd {
	return func() tea.Msg {
		if err := taskwarrior.DeleteTask(m.tasks[m.current]); err != nil {
			return errorMsg{err}
		}
		return actionCompletedMsg{message: "Task deleted."}
	}
}

func (m *ReviewModel) skipCurrentTask() tea.Cmd {
	return func() tea.Msg {
		// Just move to next task without marking as reviewed
		return taskSkippedMsg{message: "Task skipped."}
	}
}

func (m *ReviewModel) modifyCurrentTask(modification string) tea.Cmd {
	return func() tea.Msg {
		if err := taskwarrior.ModifyTask(m.tasks[m.current], modification); err != nil {
			return errorMsg{err}
		}
		return actionCompletedMsg{message: "Task modified."}
	}
}

func (m *ReviewModel) waitCurrentTask(waitDate, reason string) tea.Cmd {
	return func() tea.Msg {
		if err := taskwarrior.WaitTask(m.tasks[m.current], waitDate, reason); err != nil {
			return errorMsg{err}
		}
		message := fmt.Sprintf("Task set to wait until %s.", waitDate)
		return actionCompletedMsg{message: message}
	}
}

func (m *ReviewModel) dueCurrentTask(dueDate string) tea.Cmd {
	return func() tea.Msg {
		if err := taskwarrior.SetDueDate(m.tasks[m.current], dueDate); err != nil {
			return errorMsg{err}
		}
		message := fmt.Sprintf("Task due date set to %s.", dueDate)
		return actionCompletedMsg{message: message}
	}
}

// SetTasks initializes the review session with tasks
func (m *ReviewModel) SetTasks(tasks []string, total int) {
	m.tasks = tasks
	m.total = total
	m.current = 0
	m.reviewed = 0
}

// updateContextSelect handles context selection input
func (m *ReviewModel) updateContextSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeViewing
		m.message = ""
		return m, nil

	case key.Matches(msg, m.keys.Confirm):
		if len(m.contexts) > 0 && m.selectedContext < len(m.contexts) {
			selectedContext := m.contexts[m.selectedContext]
			m.mode = ModeViewing
			m.message = ""
			return m, m.switchContext(selectedContext)
		}
		m.mode = ModeViewing
		m.message = ""
		return m, nil

	case key.Matches(msg, m.keys.NextTask) || msg.String() == "down":
		if len(m.contexts) > 0 {
			m.selectedContext++
			if m.selectedContext >= len(m.contexts) {
				m.selectedContext = 0
			}
		}
		return m, nil

	case key.Matches(msg, m.keys.PrevTask) || msg.String() == "up":
		if len(m.contexts) > 0 {
			m.selectedContext--
			if m.selectedContext < 0 {
				m.selectedContext = len(m.contexts) - 1
			}
		}
		return m, nil
	}

	return m, nil
}

// renderContextSelect renders the context selection interface
func (m *ReviewModel) renderContextSelect() string {
	contextStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("6")). // ANSI cyan
		Padding(1, 2).
		Margin(2, 4)

	var content strings.Builder
	content.WriteString("Available Contexts:\n\n")

	for i, context := range m.contexts {
		prefix := "  "
		if i == m.selectedContext {
			prefix = "▶ " // Highlight selected context
		}

		line := fmt.Sprintf("%s%s", prefix, context)
		if context == m.currentContext {
			line += " (current)"
		}

		content.WriteString(line)
		content.WriteString("\n")
	}

	content.WriteString("\n↑↓: navigate  Enter: select  ESC: cancel")

	return contextStyle.Render(content.String())
}

// initContextSelect initializes context selection mode
func (m *ReviewModel) initContextSelect() tea.Cmd {
	return func() tea.Msg {
		contexts, err := taskwarrior.GetContexts()
		if err != nil {
			return errorMsg{err}
		}

		return contextsLoadedMsg{contexts: contexts}
	}
}

// switchContext switches to the selected context
func (m *ReviewModel) switchContext(contextName string) tea.Cmd {
	return func() tea.Msg {
		if err := taskwarrior.SetContext(contextName); err != nil {
			return errorMsg{err}
		}
		return contextChangedMsg{context: contextName}
	}
}

// analyzeCurrentTask performs AI analysis on the current task
func (m *ReviewModel) analyzeCurrentTask() tea.Cmd {
	return func() tea.Msg {
		if m.currentTask == nil {
			return errorMsg{fmt.Errorf("no current task to analyze")}
		}
		
		analysis, err := m.aiAnalyzer.AnalyzeTask(m.currentTask)
		if err != nil {
			return errorMsg{fmt.Errorf("AI analysis failed: %w", err)}
		}
		
		return aiAnalysisCompleteMsg{analysis: analysis}
	}
}

// updateAIAnalysis handles input in AI analysis mode
func (m *ReviewModel) updateAIAnalysis(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = ModeViewing
		m.message = ""
		m.currentAnalysis = nil
		return m, nil
		
	case key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, tea.Quit
	}
	
	return m, nil
}

// renderAIAnalysis renders the AI analysis view
func (m *ReviewModel) renderAIAnalysis() string {
	if m.currentAnalysis == nil {
		return "No AI analysis available"
	}

	analysisStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")). // ANSI magenta for AI
		Padding(1, 2).
		Margin(2, 4)

	var content strings.Builder
	
	// Analysis summary
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Bold(true).
		Render("🤖 AI Analysis"))
	content.WriteString("\n\n")
	
	if m.currentAnalysis.Summary != "" {
		content.WriteString(lipgloss.NewStyle().
			Bold(true).
			Render("Summary:"))
		content.WriteString("\n")
		content.WriteString(m.currentAnalysis.Summary)
		content.WriteString("\n\n")
	}
	
	// Time estimate
	if m.currentAnalysis.TimeEstimate.Hours > 0 {
		content.WriteString(lipgloss.NewStyle().
			Bold(true).
			Render("Time Estimate:"))
		content.WriteString("\n")
		content.WriteString(fmt.Sprintf("%.1f hours - %s", 
			m.currentAnalysis.TimeEstimate.Hours,
			m.currentAnalysis.TimeEstimate.Reason))
		content.WriteString("\n\n")
	}
	
	// Suggestions
	if len(m.currentAnalysis.Suggestions) > 0 {
		content.WriteString(lipgloss.NewStyle().
			Bold(true).
			Render("Suggestions:"))
		content.WriteString("\n")
		
		for i, suggestion := range m.currentAnalysis.Suggestions {
			// Type indicator with color
			typeStyle := getTypeStyle(suggestion.Type)
			content.WriteString(fmt.Sprintf("%d. %s: ", i+1, typeStyle.Render(suggestion.Type)))
			
			// Current vs suggested
			if suggestion.CurrentValue != "" {
				content.WriteString(fmt.Sprintf("\"%s\" → \"%s\"", 
					suggestion.CurrentValue, suggestion.SuggestedValue))
			} else {
				content.WriteString(fmt.Sprintf("Add \"%s\"", suggestion.SuggestedValue))
			}
			content.WriteString("\n")
			
			// Reason and confidence
			content.WriteString(fmt.Sprintf("   %s (confidence: %.0f%%)", 
				suggestion.Reason, suggestion.Confidence*100))
			content.WriteString("\n\n")
		}
	}
	
	content.WriteString("\nPress ESC to return to task view")
	
	return analysisStyle.Render(content.String())
}

// getTypeStyle returns appropriate styling for suggestion types
func getTypeStyle(suggestionType string) lipgloss.Style {
	switch suggestionType {
	case "priority":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // Red
	case "due_date":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
	case "project":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
	case "tag":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // Magenta
	case "estimate":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("4")) // Blue
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // White
	}
}

// updateAILoading handles input in AI loading mode
func (m *ReviewModel) updateAILoading(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		// Allow canceling the loading
		m.mode = ModeViewing
		m.message = "AI analysis cancelled"
		return m, nil
		
	case key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, tea.Quit
	}
	
	return m, nil
}

// renderAILoading renders the AI loading view
func (m *ReviewModel) renderAILoading() string {
	loadingStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")). // ANSI magenta for AI
		Padding(2, 4).
		Margin(2, 4).
		Align(lipgloss.Center)

	var content strings.Builder
	
	// Spinner and loading message
	content.WriteString(m.aiSpinner.View())
	content.WriteString(" ")
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("5")).
		Bold(true).
		Render("Analyzing task with AI..."))
	content.WriteString("\n\n")
	
	// Current task info for context
	if m.currentTask != nil {
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render("Task: "))
		content.WriteString(m.currentTask.Description)
		content.WriteString("\n")
	}
	
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("This may take a few seconds..."))
	content.WriteString("\n\n")
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("Press ESC to cancel"))
	
	return loadingStyle.Render(content.String())
}