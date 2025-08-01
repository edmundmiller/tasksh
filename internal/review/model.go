package review

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/emiller/tasksh/internal/taskwarrior"
)

// ImprovedModel represents the improved review UI state
type ImprovedModel struct {
	// Core data
	tasks       []string
	taskCache   map[string]*taskwarrior.Task
	current     int
	reviewed    int
	
	// UI components
	viewport    viewport.Model
	help        help.Model
	input       textinput.Model
	progress    progress.Model
	spinner     spinner.Model
	
	// UI state
	mode        Mode
	width       int
	height      int
	message     string
	error       error
	
	// Features
	commands       *CommandManager
	renderer       *ViewRenderer
	feedback       *FeedbackManager
	sessionManager *SessionManager
	session        *Session
	shortcuts      KeyMap // from bubbletea.go
	
	// Behavior flags
	quitting    bool
	loading     bool
}

// Mode represents the current UI mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeInput
	ModeConfirm
	ModeHelp
	ModeError
)

// NewImprovedModel creates a new improved review model with sensible defaults
func NewImprovedModel() *ImprovedModel {
	// Create viewport
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))
	
	// Create help
	h := help.New()
	h.ShowAll = false
	
	// Create text input
	ti := textinput.New()
	ti.Placeholder = "Type here..."
	ti.CharLimit = 100
	
	// Create progress bar
	prog := progress.New(progress.WithDefaultGradient())
	
	// Create spinner
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	// Create model
	m := &ImprovedModel{
		tasks:     []string{},
		taskCache: make(map[string]*taskwarrior.Task),
		viewport:  vp,
		help:      h,
		input:     ti,
		progress:  prog,
		spinner:   spin,
		mode:      ModeNormal,
		shortcuts: DefaultKeyMap(),
	}
	
	// Initialize subsystems
	m.commands = NewCommandManager()
	m.renderer = NewViewRenderer()
	m.feedback = NewFeedbackManager()
	
	// Apply theme
	theme := GetTheme()
	m.renderer.ApplyTheme(theme)
	
	// Initialize session manager
	if sm, err := NewSessionManager(); err == nil {
		m.sessionManager = sm
	}
	
	return m
}

// Init initializes the model
func (m *ImprovedModel) Init() tea.Cmd {
	return tea.Batch(
		tea.WindowSize(),
		spinner.Tick,
	)
}

// Update handles all UI updates
func (m *ImprovedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Handle global messages
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		
	case tea.KeyMsg:
		if m.mode == ModeInput {
			return m.handleInput(msg)
		}
		return m.handleKeypress(msg)
		
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		
	case TaskLoadedMsg:
		m.loading = false
		if task, ok := m.taskCache[msg.UUID]; ok {
			m.viewport.SetContent(m.renderer.RenderTask(task))
		}
		
	case ActionCompletedMsg:
		m.message = msg.Message
		m.reviewed++
		m.UpdateSession()
		cmds = append(cmds, m.nextTask())
		
	case ErrorMsg:
		m.error = msg.Error
		m.mode = ModeError
		cmds = append(cmds, m.feedback.ShowToast(ToastError, msg.Error.Error()))
		
	case ShowToastMsg:
		cmds = append(cmds, m.feedback.ShowToast(msg.Type, msg.Message))
		
	case toastExpiredMsg:
		m.feedback.RemoveExpiredToast(msg.startTime)
		
	case ReviewCompleteMsg:
		// Clear session on completion
		if m.sessionManager != nil {
			m.sessionManager.ClearSession()
		}
		m.message = fmt.Sprintf("Review complete! %d of %d tasks reviewed.", msg.Reviewed, msg.Total)
		cmds = append(cmds, ShowSuccess("Review complete!"))
	}
	
	// Update components
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	
	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m *ImprovedModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}
	
	var mainView string
	
	switch m.mode {
	case ModeHelp:
		mainView = m.renderer.RenderHelp(m.shortcuts, m.width, m.height)
	case ModeError:
		mainView = m.renderer.RenderError(m.error, m.width, m.height)
	case ModeInput:
		mainView = m.renderer.RenderInput(m.input, m.message, m.width, m.height)
	default:
		mainView = m.renderer.RenderMain(m)
	}
	
	// Overlay toasts if any
	toasts := m.feedback.RenderToasts(m.width)
	if toasts != "" {
		// Position toasts at top-right
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Right, lipgloss.Top,
			toasts,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.NoColor{}),
		) + "\n" + mainView
	}
	
	return mainView
}

// SetTasks initializes the task list
func (m *ImprovedModel) SetTasks(uuids []string, cache map[string]*taskwarrior.Task) {
	m.tasks = uuids
	m.taskCache = cache
	m.current = 0
	m.reviewed = 0
	
	if len(uuids) > 0 {
		m.loadCurrentTask()
	}
}

// Helper methods

func (m *ImprovedModel) updateLayout() {
	// Adjust viewport size
	headerHeight := 3
	footerHeight := 3
	m.viewport.Width = m.width - 2
	m.viewport.Height = m.height - headerHeight - footerHeight
	
	// Update progress bar width
	m.progress.Width = m.width - 10
}

func (m *ImprovedModel) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Use command manager to handle keypress
	return m.commands.ExecuteCommand(msg.String(), m)
}

func (m *ImprovedModel) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		value := m.input.Value()
		m.input.Reset()
		m.mode = ModeNormal
		// Process input based on context
		return m, m.processInput(value)
		
	case tea.KeyEscape:
		m.input.Reset()
		m.mode = ModeNormal
		m.message = ""
	}
	
	// Update input field
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *ImprovedModel) loadCurrentTask() tea.Cmd {
	if m.current >= len(m.tasks) {
		return nil
	}
	
	m.loading = true
	uuid := m.tasks[m.current]
	
	// Check cache first
	if task, ok := m.taskCache[uuid]; ok {
		return func() tea.Msg {
			return TaskLoadedMsg{UUID: uuid, Task: task}
		}
	}
	
	// Load from taskwarrior
	return func() tea.Msg {
		task, err := taskwarrior.GetTaskInfo(uuid)
		if err != nil {
			return ErrorMsg{Error: err}
		}
		m.taskCache[uuid] = task
		return TaskLoadedMsg{UUID: uuid, Task: task}
	}
}

func (m *ImprovedModel) nextTask() tea.Cmd {
	if m.current < len(m.tasks)-1 {
		m.current++
		return m.loadCurrentTask()
	}
	
	// Review complete
	return func() tea.Msg {
		return ReviewCompleteMsg{
			Total:    len(m.tasks),
			Reviewed: m.reviewed,
		}
	}
}

func (m *ImprovedModel) processInput(value string) tea.Cmd {
	// This would be extended based on the input context
	return nil
}

// UpdateSession updates the session state
func (m *ImprovedModel) UpdateSession() {
	if m.sessionManager != nil && m.session != nil {
		m.session.Current = m.current
		m.session.Reviewed = m.reviewed
		m.session.Tasks = m.tasks
		m.sessionManager.SaveSession(m.session)
	}
}


// Messages

type TaskLoadedMsg struct {
	UUID string
	Task *taskwarrior.Task
}

type ActionCompletedMsg struct {
	Message string
}

type ErrorMsg struct {
	Error error
}

type ReviewCompleteMsg struct {
	Total    int
	Reviewed int
}


