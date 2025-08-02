package weekly

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PlanningModel represents the Bubble Tea model for weekly planning
type PlanningModel struct {
	session      *WeeklyPlanningSession
	viewport     viewport.Model
	textInput    textinput.Model
	textArea     textarea.Model
	help         help.Model
	keys         KeyMap
	
	// UI state
	selectedIndex     int
	message          string
	err              error
	quitting         bool
	width            int
	height           int
	
	// Step-specific state
	editingObjective bool
	editingWorkStream bool
	showingHelp      bool
}

// KeyMap defines key bindings for the weekly planning interface
type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Add       key.Binding
	Edit      key.Binding
	Delete    key.Binding
	Next      key.Binding
	Previous  key.Binding
	Help      key.Binding
	Quit      key.Binding
	
	// Text input specific
	Confirm   key.Binding
	Cancel    key.Binding
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
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
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
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
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
	return []key.Binding{k.Up, k.Down, k.Add, k.Next, k.Previous, k.Help, k.Quit}
}

// FullHelp returns the full help text
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Add, k.Edit, k.Delete},
		{k.Next, k.Previous, k.Confirm, k.Cancel},
		{k.Help, k.Quit},
	}
}

// NewPlanningModel creates a new weekly planning model
func NewPlanningModel(session *WeeklyPlanningSession) *PlanningModel {
	// Create viewport
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle()

	// Create text input for objectives and work streams
	ti := textinput.New()
	ti.Placeholder = "Enter title..."
	ti.CharLimit = 100
	ti.Width = 70

	// Create text area for journaling
	ta := textarea.New()
	ta.Placeholder = "Share your strategic thoughts for the week..."
	ta.SetWidth(70)
	ta.SetHeight(8)

	// Create help
	h := help.New()
	h.ShowAll = false
	
	// Style the help for better visibility (matching review interface)
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))   // ANSI cyan for keys
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))  // ANSI white for descriptions
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // ANSI bright black for separators
	h.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))    // ANSI cyan for keys in full help
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))   // ANSI white for descriptions in full help
	h.Styles.FullSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // ANSI bright black for separators

	model := &PlanningModel{
		session:   session,
		viewport:  vp,
		textInput: ti,
		textArea:  ta,
		help:      h,
		keys:      DefaultKeyMap(),
		width:     80,
		height:    24,
	}

	return model
}

// Init initializes the model
func (m *PlanningModel) Init() tea.Cmd {
	return tea.Batch(
		tea.WindowSize(),
		textinput.Blink,
	)
}

// Update handles messages and updates the model
func (m *PlanningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		headerHeight := 4
		footerHeight := 4
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight - footerHeight
		
		m.updateViewport()

	case tea.KeyMsg:
		// Handle text input modes first
		if m.editingObjective && m.textInput.Focused() {
			switch {
			case key.Matches(msg, m.keys.Confirm):
				m.handleObjectiveInput()
				return m, nil
			case key.Matches(msg, m.keys.Cancel):
				m.editingObjective = false
				m.textInput.Blur()
				m.textInput.SetValue("")
				return m, nil
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}

		if m.session.CurrentStep == StepJournaling && m.textArea.Focused() {
			switch {
			case key.Matches(msg, m.keys.Confirm):
				m.session.SetJournalEntry(m.textArea.Value())
				m.textArea.Blur()
				m.message = "Journal entry saved"
				m.updateViewport()
				return m, nil
			case key.Matches(msg, m.keys.Cancel):
				m.textArea.Blur()
				return m, nil
			default:
				m.textArea, cmd = m.textArea.Update(msg)
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

		case key.Matches(msg, m.keys.Add):
			m.handleAdd()

		case key.Matches(msg, m.keys.Edit):
			m.handleEdit()

		case key.Matches(msg, m.keys.Delete):
			m.handleDelete()

		case key.Matches(msg, m.keys.Next):
			m.handleNextStep()

		case key.Matches(msg, m.keys.Previous):
			m.handlePreviousStep()
		}
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the weekly planning interface
func (m *PlanningModel) View() string {
	if m.quitting {
		return "\nWeekly planning session ended.\n\n"
	}

	if m.width == 0 || m.height == 0 {
		return ""
	}

	var sections []string

	// Header
	sections = append(sections, m.renderHeader())

	// Main content
	if m.editingObjective {
		sections = append(sections, m.renderObjectiveInput())
	} else if m.session.CurrentStep == StepJournaling && m.textArea.Focused() {
		sections = append(sections, m.renderJournalingInput())
	} else {
		sections = append(sections, m.viewport.View())
	}

	// Message area
	if m.message != "" {
		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
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
	
	title := fmt.Sprintf("Weekly Planning - Step %d/%d: %s", current+1, total, stepName)
	dateRange := fmt.Sprintf("%s - %s", 
		m.session.WeekStart.Format("Monday, Jan 2"), 
		m.session.WeekEnd.Format("Friday, Jan 6, 2006"))
	
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("5")).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width).
		Padding(0, 1)

	header := headerStyle.Render(title)
	dateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Align(lipgloss.Center).
		Width(m.width)
	dateHeader := dateStyle.Render(dateRange)
	
	separator := strings.Repeat("━", m.width)
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	return header + "\n" + dateHeader + "\n" + sepStyle.Render(separator)
}

// renderFooter renders the help and navigation footer
func (m *PlanningModel) renderFooter() string {
	helpSepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // Bright black for separator (matches help separator style)
	helpSep := helpSepStyle.Render(strings.Repeat("━", m.width))
	
	return helpSep + "\n" + m.help.View(m.keys)
}

// renderObjectiveInput renders the objective input interface
func (m *PlanningModel) renderObjectiveInput() string {
	var content strings.Builder
	
	content.WriteString("\n🎯 Add Weekly Objective\n\n")
	content.WriteString("Create a clear, achievable outcome for this week.\n")
	content.WriteString("Focus on results rather than activities.\n\n")
	content.WriteString("Objective title:\n")
	content.WriteString(m.textInput.View())
	content.WriteString("\n\nPress Enter to add, Esc to cancel")
	
	return content.String()
}

// renderJournalingInput renders the journaling text area
func (m *PlanningModel) renderJournalingInput() string {
	var content strings.Builder
	
	content.WriteString("\n📝 Strategic Journaling\n\n")
	content.WriteString("Take a moment to think strategically about the week ahead:\n\n")
	
	prompts := m.session.GetJournalingPrompts()
	for i, prompt := range prompts[:3] { // Show first 3 prompts
		content.WriteString(fmt.Sprintf("• %s\n", prompt))
		if i < 2 {
			content.WriteString("  \n")
		}
	}
	
	content.WriteString("\n")
	content.WriteString(m.textArea.View())
	content.WriteString("\n\nPress Enter to save, Esc to cancel")
	
	return content.String()
}

// updateViewport updates the viewport content based on current step
func (m *PlanningModel) updateViewport() {
	var content strings.Builder
	content.WriteString("\n")

	switch m.session.CurrentStep {
	case StepWeeklyReflection:
		content.WriteString(m.renderWeeklyReflectionStep())
	case StepObjectiveSetting:
		content.WriteString(m.renderObjectiveSettingStep())
	case StepJournaling:
		content.WriteString(m.renderJournalingStep())
	case StepWorkStreamPlanning:
		content.WriteString(m.renderWorkStreamPlanningStep())
	case StepWeeklySummary:
		content.WriteString(m.renderWeeklySummaryStep())
	default:
		content.WriteString("Unknown step")
	}

	m.viewport.SetContent(content.String())
}

// renderWeeklyReflectionStep renders the weekly reflection step
func (m *PlanningModel) renderWeeklyReflectionStep() string {
	var content strings.Builder
	
	content.WriteString("🔄 Weekly Reflection\n\n")
	content.WriteString("Strategic planning begins with reflection. Take time to review\n")
	content.WriteString("last week's outcomes and extract insights for better planning.\n\n")
	
	if m.session.WeeklyReflection != nil {
		content.WriteString("✓ Weekly reflection completed\n\n")
		reflection := m.session.WeeklyReflection
		
		if len(reflection.KeyAccomplishments) > 0 {
			content.WriteString("Key accomplishments:\n")
			for _, acc := range reflection.KeyAccomplishments {
				content.WriteString(fmt.Sprintf("  • %s\n", acc))
			}
			content.WriteString("\n")
		}
		
		if len(reflection.LessonsLearned) > 0 {
			content.WriteString("Lessons learned:\n")
			for _, lesson := range reflection.LessonsLearned {
				content.WriteString(fmt.Sprintf("  • %s\n", lesson))
			}
			content.WriteString("\n")
		}
		
		content.WriteString("Press 'n' to continue to objective setting")
	} else {
		content.WriteString("Reflection prompts:\n\n")
		prompts := m.session.GetReflectionPrompts()
		for i, prompt := range prompts {
			content.WriteString(fmt.Sprintf("%d. %s\n", i+1, prompt))
		}
		content.WriteString("\nPress 'a' to add reflection notes, or 'n' to skip")
	}
	
	return content.String()
}

// renderObjectiveSettingStep renders the objective setting step
func (m *PlanningModel) renderObjectiveSettingStep() string {
	var content strings.Builder
	
	content.WriteString("🎯 Set Weekly Objectives\n\n")
	content.WriteString("Define 2-3 key outcomes that would make this week successful.\n")
	content.WriteString("Focus on results and impact, not just activities.\n\n")
	
	if len(m.session.Objectives) > 0 {
		content.WriteString("WEEKLY OBJECTIVES:\n")
		for i, obj := range m.session.Objectives {
			prefix := "  "
			if i == m.selectedIndex {
				prefix = "▶ "
			}
			content.WriteString(fmt.Sprintf("%s%d. %s\n", prefix, i+1, obj.Title))
			if obj.Description != "" {
				content.WriteString(fmt.Sprintf("     %s\n", obj.Description))
			}
		}
		content.WriteString("\n")
	}
	
	// Show suggestions
	content.WriteString("SUGGESTIONS:\n")
	suggestions := m.session.GetObjectiveSuggestions()
	for _, suggestion := range suggestions[:5] { // Show first 5 suggestions
		content.WriteString(fmt.Sprintf("  • %s\n", suggestion))
	}
	
	content.WriteString("\nPress 'a' to add objective, 'e' to edit, 'd' to delete, 'n' to continue")
	
	return content.String()
}

// renderJournalingStep renders the journaling step
func (m *PlanningModel) renderJournalingStep() string {
	var content strings.Builder
	
	content.WriteString("📝 Strategic Journaling\n\n")
	content.WriteString("Step back and think about the bigger picture. This journaling\n")
	content.WriteString("helps align your week with your values and long-term goals.\n\n")
	
	if m.session.JournalEntry != "" {
		content.WriteString("STRATEGIC REFLECTION:\n")
		content.WriteString("─────────────────────\n")
		content.WriteString(m.session.JournalEntry)
		content.WriteString("\n\n✓ Journal entry completed\n")
		content.WriteString("Press 'e' to edit, or 'n' to continue")
	} else {
		content.WriteString("REFLECTION PROMPTS:\n")
		prompts := m.session.GetJournalingPrompts()
		for i, prompt := range prompts {
			content.WriteString(fmt.Sprintf("  %d. %s\n", i+1, prompt))
		}
		content.WriteString("\nPress 'a' to start journaling, or 'n' to skip")
	}
	
	return content.String()
}

// renderWorkStreamPlanningStep renders the work stream planning step
func (m *PlanningModel) renderWorkStreamPlanningStep() string {
	var content strings.Builder
	
	content.WriteString("🚀 Plan Work Streams\n\n")
	content.WriteString("Organize your work into themed streams that support your objectives.\n")
	content.WriteString("This helps maintain focus and track progress.\n\n")
	
	if len(m.session.WorkStreams) > 0 {
		content.WriteString("WORK STREAMS:\n")
		for i, ws := range m.session.WorkStreams {
			prefix := "  "
			if i == m.selectedIndex {
				prefix = "▶ "
			}
			content.WriteString(fmt.Sprintf("%s%d. %s (%.1fh)\n", prefix, i+1, ws.Title, ws.EstimatedHours))
			if ws.Description != "" {
				content.WriteString(fmt.Sprintf("     %s\n", ws.Description))
			}
			if len(ws.Tasks) > 0 {
				content.WriteString(fmt.Sprintf("     %d tasks assigned\n", len(ws.Tasks)))
			}
		}
		content.WriteString("\n")
	} else {
		content.WriteString("No work streams created yet.\n\n")
	}
	
	// Show available tasks
	if len(m.session.AvailableTasks) > 0 {
		content.WriteString(fmt.Sprintf("Available tasks for planning: %d\n", len(m.session.AvailableTasks)))
	}
	
	content.WriteString("Press 'a' to create work stream, 'e' to edit, 'n' to continue")
	
	return content.String()
}

// renderWeeklySummaryStep renders the final summary
func (m *PlanningModel) renderWeeklySummaryStep() string {
	var content strings.Builder
	
	content.WriteString("📊 Weekly Plan Summary\n\n")
	content.WriteString(m.session.GetWeeklySummary())
	content.WriteString("\n\nPress 'n' to complete weekly planning")
	
	return content.String()
}

// Helper methods

func (m *PlanningModel) getMaxSelectableIndex() int {
	switch m.session.CurrentStep {
	case StepObjectiveSetting:
		if len(m.session.Objectives) > 0 {
			return len(m.session.Objectives) - 1
		}
	case StepWorkStreamPlanning:
		if len(m.session.WorkStreams) > 0 {
			return len(m.session.WorkStreams) - 1
		}
	}
	return 0
}

func (m *PlanningModel) handleAdd() {
	switch m.session.CurrentStep {
	case StepWeeklyReflection:
		// Create a simple reflection (could be enhanced with proper input)
		reflection := &WeeklyReflectionData{
			PreviousWeekStart: m.session.WeekStart.AddDate(0, 0, -7),
			PreviousWeekEnd:   m.session.WeekStart.AddDate(0, 0, -1),
			KeyAccomplishments: []string{"Completed key tasks from last week"},
			LessonsLearned:    []string{"Focus on realistic planning"},
			OverallSatisfaction: 4,
		}
		m.session.SetWeeklyReflection(reflection)
		m.message = "Basic reflection added"
		m.updateViewport()
		
	case StepObjectiveSetting:
		m.editingObjective = true
		m.textInput.Focus()
		m.textInput.SetValue("")
		
	case StepJournaling:
		m.textArea.Focus()
		
	case StepWorkStreamPlanning:
		// Simple work stream creation (could be enhanced)
		title := fmt.Sprintf("Work Stream %d", len(m.session.WorkStreams)+1)
		m.session.CreateWorkStream(title, "Focus area for the week", []string{}, 8.0)
		m.message = "Work stream created"
		m.updateViewport()
	}
}

func (m *PlanningModel) handleEdit() {
	switch m.session.CurrentStep {
	case StepJournaling:
		m.textArea.Focus()
		if m.session.JournalEntry != "" {
			m.textArea.SetValue(m.session.JournalEntry)
		}
	}
}

func (m *PlanningModel) handleDelete() {
	switch m.session.CurrentStep {
	case StepObjectiveSetting:
		if m.selectedIndex >= 0 && m.selectedIndex < len(m.session.Objectives) {
			err := m.session.RemoveObjective(m.selectedIndex)
			if err != nil {
				m.message = fmt.Sprintf("Error: %v", err)
			} else {
				m.message = "Objective removed"
				if m.selectedIndex >= len(m.session.Objectives) && len(m.session.Objectives) > 0 {
					m.selectedIndex = len(m.session.Objectives) - 1
				}
			}
			m.updateViewport()
		}
	}
}

func (m *PlanningModel) handleObjectiveInput() {
	title := strings.TrimSpace(m.textInput.Value())
	if title != "" {
		m.session.AddObjective(title, "", len(m.session.Objectives)+1)
		m.message = "Objective added"
		m.updateViewport()
	}
	
	m.editingObjective = false
	m.textInput.Blur()
	m.textInput.SetValue("")
}

func (m *PlanningModel) handleNextStep() {
	err := m.session.NextStep()
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
	} else {
		m.selectedIndex = 0
		m.message = ""
		m.updateViewport()
	}
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

// Run starts the weekly planning interface
func Run(weekStart time.Time) error {
	session, err := NewWeeklyPlanningSession(weekStart)
	if err != nil {
		return fmt.Errorf("failed to create weekly planning session: %w", err)
	}
	defer session.Close()

	model := NewPlanningModel(session)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run weekly planning interface: %w", err)
	}

	return nil
}