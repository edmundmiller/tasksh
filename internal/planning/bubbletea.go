package planning

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PlanningMode represents the current mode of the planning interface
type PlanningMode int

const (
	ModeViewing PlanningMode = iota
	ModeReordering
	ModeEditing
)

// PlanningModel represents the state of the Bubble Tea planning interface
type PlanningModel struct {
	// Planning session
	session *PlanningSession

	// UI components
	viewport viewport.Model
	help     help.Model
	keys     PlanningKeyMap

	// Application state
	mode           PlanningMode
	selectedTask   int
	message        string
	err            error
	quitting       bool
	width          int
	height         int
	showProjection bool // Whether to show time projections

	// Time projection settings
	workStartTime time.Time
}

// PlanningKeyMap defines the key bindings for the planning interface
type PlanningKeyMap struct {
	// Navigation
	Up   key.Binding
	Down key.Binding

	// Section navigation
	JumpSection1 key.Binding // Jump to Critical section
	JumpSection2 key.Binding // Jump to Important section  
	JumpSection3 key.Binding // Jump to Flexible section

	// Actions
	MoveUp        key.Binding
	MoveDown      key.Binding
	Remove        key.Binding
	EditTime      key.Binding
	ToggleView    key.Binding
	Save          key.Binding
	Projection    key.Binding
	PromoteCritical key.Binding // Mark as "must do"
	Defer         key.Binding    // Defer to tomorrow
	BrowseBacklog key.Binding    // Browse backlog tasks
	Filter        key.Binding    // Filter tasks

	// General
	Help key.Binding
	Quit key.Binding
}

// DefaultPlanningKeyMap returns the default key bindings for planning
func DefaultPlanningKeyMap() PlanningKeyMap {
	return PlanningKeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		MoveUp: key.NewBinding(
			key.WithKeys("K", "shift+up"),
			key.WithHelp("K", "move up"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("J", "shift+down"),
			key.WithHelp("J", "move down"),
		),
		Remove: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "remove"),
		),
		EditTime: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit time"),
		),
		ToggleView: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "toggle view"),
		),
		Save: key.NewBinding(
			key.WithKeys("s", "enter"),
			key.WithHelp("s", "save plan"),
		),
		Projection: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "toggle projection"),
		),
		
		// Section navigation
		JumpSection1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "critical section"),
		),
		JumpSection2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "important section"),
		),
		JumpSection3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "flexible section"),
		),
		
		// Enhanced task management
		PromoteCritical: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "promote to critical"),
		),
		Defer: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "defer to tomorrow"),
		),
		BrowseBacklog: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "browse backlog"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter tasks"),
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
func (k PlanningKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.MoveUp, k.MoveDown, k.PromoteCritical, k.Defer, k.Save, k.Help, k.Quit}
}

// FullHelp returns the full help text
func (k PlanningKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.JumpSection1, k.JumpSection2, k.JumpSection3},
		{k.MoveUp, k.MoveDown, k.Remove},
		{k.PromoteCritical, k.Defer, k.BrowseBacklog, k.Filter},
		{k.EditTime, k.Projection, k.ToggleView},
		{k.Save, k.Help, k.Quit},
	}
}

// NewPlanningModel creates a new planning model
func NewPlanningModel(session *PlanningSession) *PlanningModel {
	// Create viewport
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle()

	// Create help model
	h := help.New()
	h.ShowAll = false

	// Style the help
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// Default work start time (9 AM)
	workStart := time.Now()
	workStart = time.Date(workStart.Year(), workStart.Month(), workStart.Day(), 9, 0, 0, 0, workStart.Location())

	model := &PlanningModel{
		session:       session,
		viewport:      vp,
		help:          h,
		keys:          DefaultPlanningKeyMap(),
		mode:          ModeViewing,
		selectedTask:  0,
		showProjection: true,
		workStartTime: workStart,
		width:         80,  // Default width
		height:        24,  // Default height
	}

	return model
}

// Init initializes the planning model
func (m *PlanningModel) Init() tea.Cmd {
	m.updateViewport()
	return nil
}

// Update handles messages and updates the model
func (m *PlanningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update viewport dimensions
		headerHeight := 4 // Title + separator + capacity bar
		footerHeight := 3 // Summary + separator + help text
		verticalMarginHeight := headerHeight + footerHeight

		if !m.help.ShowAll {
			footerHeight = 3
			verticalMarginHeight = headerHeight + footerHeight
		}

		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, m.keys.Up):
			if m.selectedTask > 0 {
				m.selectedTask--
				m.updateViewport()
			}

		case key.Matches(msg, m.keys.Down):
			if m.selectedTask < len(m.session.Tasks)-1 {
				m.selectedTask++
				m.updateViewport()
			}

		case key.Matches(msg, m.keys.MoveUp):
			if m.selectedTask > 0 {
				err := m.session.MoveTask(m.selectedTask, m.selectedTask-1)
				if err == nil {
					m.selectedTask--
					m.updateViewport()
					m.message = "Task moved up"
				} else {
					m.message = fmt.Sprintf("Error: %v", err)
				}
			}

		case key.Matches(msg, m.keys.MoveDown):
			if m.selectedTask < len(m.session.Tasks)-1 {
				err := m.session.MoveTask(m.selectedTask, m.selectedTask+1)
				if err == nil {
					m.selectedTask++
					m.updateViewport()
					m.message = "Task moved down"
				} else {
					m.message = fmt.Sprintf("Error: %v", err)
				}
			}

		case key.Matches(msg, m.keys.Remove):
			if len(m.session.Tasks) > 0 {
				err := m.session.RemoveTask(m.selectedTask)
				if err == nil {
					if m.selectedTask >= len(m.session.Tasks) && len(m.session.Tasks) > 0 {
						m.selectedTask = len(m.session.Tasks) - 1
					}
					m.updateViewport()
					m.message = "Task removed from plan"
				} else {
					m.message = fmt.Sprintf("Error: %v", err)
				}
			}

		case key.Matches(msg, m.keys.Projection):
			m.showProjection = !m.showProjection
			m.updateViewport()
			if m.showProjection {
				m.message = "Time projections enabled"
			} else {
				m.message = "Time projections disabled"
			}

		// Section navigation shortcuts
		case key.Matches(msg, m.keys.JumpSection1):
			if len(m.session.CriticalTasks) > 0 {
				m.selectedTask = 0
				m.updateViewport()
				m.message = "Jumped to Critical section"
			}

		case key.Matches(msg, m.keys.JumpSection2):
			if len(m.session.ImportantTasks) > 0 {
				m.selectedTask = len(m.session.CriticalTasks)
				m.updateViewport()
				m.message = "Jumped to Important section"
			}

		case key.Matches(msg, m.keys.JumpSection3):
			if len(m.session.FlexibleTasks) > 0 {
				m.selectedTask = len(m.session.CriticalTasks) + len(m.session.ImportantTasks)
				m.updateViewport()
				m.message = "Jumped to Flexible section"
			}

		// Enhanced task management
		case key.Matches(msg, m.keys.PromoteCritical):
			if m.selectedTask < len(m.session.Tasks) {
				m.promoteTaskToCritical(m.selectedTask)
			}

		case key.Matches(msg, m.keys.Defer):
			if m.selectedTask < len(m.session.Tasks) {
				m.deferTask(m.selectedTask)
			}

		case key.Matches(msg, m.keys.BrowseBacklog):
			m.showBacklogMessage()

		case key.Matches(msg, m.keys.Filter):
			m.message = "Filter functionality coming soon! (f)"

		case key.Matches(msg, m.keys.Save):
			m.message = "Plan saved! (Note: Implementation needed to persist to taskwarrior)"
			// TODO: Implement saving plan to taskwarrior (add 'planned' UDA)
		}
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the planning interface
func (m *PlanningModel) View() string {
	if m.quitting {
		return fmt.Sprintf("\nPlanning session ended. %d tasks in plan.\n\n", len(m.session.Tasks))
	}

	var sections []string

	// Header with title and date
	header := m.renderHeader()
	sections = append(sections, header)

	// Capacity status bar
	capacityBar := m.renderCapacityBar()
	sections = append(sections, capacityBar)

	// Task list
	sections = append(sections, m.viewport.View())

	// Message area
	if m.message != "" {
		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true).
			Margin(1, 0)
		sections = append(sections, messageStyle.Render(m.message))
	}

	// Help with separator
	helpSepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	helpWidth := m.width - 2
	if helpWidth < 0 {
		helpWidth = 0
	}
	helpSep := helpSepStyle.Render(strings.Repeat("━", helpWidth))
	sections = append(sections, helpSep)
	sections = append(sections, m.help.View(m.keys))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderHeader renders the planning session header
func (m *PlanningModel) renderHeader() string {
	var title string
	switch m.session.Horizon {
	case HorizonTomorrow:
		title = fmt.Sprintf("Tomorrow's Plan (%s)", m.session.Date.Format("Monday, January 2"))
	case HorizonWeek:
		title = "Weekly Plan"
	case HorizonQuick:
		title = "Quick Planning Mode"
	}

	headerWidth := m.width - 2
	if headerWidth < 1 {
		headerWidth = 1
	}
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")). // Bright white
		Background(lipgloss.Color("6")).   // Cyan background
		Bold(true).
		Align(lipgloss.Center).
		Width(headerWidth).
		Padding(0, 1)

	// Add separator line
	sepWidth := m.width - 2
	if sepWidth < 0 {
		sepWidth = 0
	}
	separator := strings.Repeat("━", sepWidth)
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	return headerStyle.Render(title) + "\n" + sepStyle.Render(separator)
}

// renderCapacityBar renders the capacity status bar
func (m *PlanningModel) renderCapacityBar() string {
	status := m.session.GetCapacityStatus()

	var icon string
	var style lipgloss.Style
	switch m.session.WarningLevel {
	case WarningOverload:
		icon = "⚠️ "
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("1")).
			Bold(true)
	case WarningCaution:
		icon = "⚠️ "
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("3")).
			Bold(true)
	default:
		icon = "✓ "
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true)
	}

	// Create capacity bar with icon
	capacityText := icon + status
	barWidth := m.width - 2
	if barWidth < 1 {
		barWidth = 1
	}
	capacityBar := style.Width(barWidth).Align(lipgloss.Center).Padding(0, 1).Render(capacityText)
	
	return capacityBar
}

// updateViewport updates the viewport content with categorized task sections
func (m *PlanningModel) updateViewport() {
	totalTasks := len(m.session.CriticalTasks) + len(m.session.ImportantTasks) + len(m.session.FlexibleTasks)
	if totalTasks == 0 && len(m.session.BacklogTasks) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			Align(lipgloss.Center).
			Margin(2, 0)
		m.viewport.SetContent(emptyStyle.Render("No tasks found for planning."))
		return
	}

	var content strings.Builder

	// Add some spacing
	content.WriteString("\n")

	// Get projected completion times for all tasks if enabled
	var completionTimes []time.Time
	if m.showProjection {
		completionTimes = m.session.GetProjectedCompletionTimes(m.workStartTime)
	}

	currentIndex := 0

	// Always show all three sections for consistency
	content.WriteString(m.renderSection("CRITICAL TASKS", m.session.CriticalTasks, currentIndex, completionTimes, lipgloss.Color("1"))) // Red
	currentIndex += len(m.session.CriticalTasks)
	content.WriteString("\n\n")

	content.WriteString(m.renderSection("IMPORTANT TASKS", m.session.ImportantTasks, currentIndex, completionTimes, lipgloss.Color("3"))) // Yellow
	currentIndex += len(m.session.ImportantTasks)
	content.WriteString("\n\n")

	content.WriteString(m.renderSection("FLEXIBLE TASKS", m.session.FlexibleTasks, currentIndex, completionTimes, lipgloss.Color("2"))) // Green
	currentIndex += len(m.session.FlexibleTasks)
	content.WriteString("\n")

	// Add summary and backlog info
	content.WriteString(m.renderSummary(completionTimes))

	m.viewport.SetContent(content.String())
}

// renderSection renders a categorized section of tasks
func (m *PlanningModel) renderSection(title string, tasks []PlannedTask, startIndex int, completionTimes []time.Time, color lipgloss.Color) string {
	var section strings.Builder

	// Section header with colored border
	headerStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true)
	
	totalHours := 0.0
	for _, task := range tasks {
		totalHours += task.EstimatedHours
	}
	
	// Calculate dynamic width based on terminal size
	contentWidth := m.width - 4 // Leave some margin
	if contentWidth < 80 {
		contentWidth = 80 // Minimum width
	}
	
	// Create header with hours on the right
	headerText := fmt.Sprintf("┏━ %s ", title)
	hoursText := fmt.Sprintf(" %.1fh ━━━┓", totalHours)
	padding := contentWidth - len(headerText) - len(hoursText) + 12 // Account for color codes
	if padding < 1 {
		padding = 1
	}
	header := headerText + strings.Repeat("━", padding) + hoursText
	section.WriteString(headerStyle.Render(header))
	section.WriteString("\n")

	// Render tasks in this section
	if len(tasks) == 0 {
		// Empty section
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)
		emptyText := "(No tasks in this category)"
		emptyStyled := emptyStyle.Render(emptyText)
		emptyPrefix := "┃  "
		
		// Calculate padding
		prefixWidth := visualWidth(emptyPrefix)
		textWidth := visualWidth(emptyText)
		rightBorderWidth := 1
		
		paddingNeeded := contentWidth - prefixWidth - textWidth - rightBorderWidth
		if paddingNeeded < 0 {
			paddingNeeded = 0
		}
		
		emptyLine := emptyPrefix + emptyStyled + strings.Repeat(" ", paddingNeeded) + "┃"
		section.WriteString(emptyLine)
		section.WriteString("\n")
	} else {
		// Add empty line for breathing room
		spaceWidth := contentWidth - 2
		if spaceWidth < 0 {
			spaceWidth = 0
		}
		section.WriteString("┃" + strings.Repeat(" ", spaceWidth) + "┃\n")
		
		for i, task := range tasks {
			taskIndex := startIndex + i
			
			// Selection indicator and task number
			var prefix string
			if taskIndex == m.selectedTask {
				prefix = fmt.Sprintf("┃  ▶ %d  ", taskIndex+1)
			} else {
				prefix = fmt.Sprintf("┃    %d  ", taskIndex+1)
			}

			// Task description
			description := task.Description
			
			// Time and completion info on the right
			timeInfo := fmt.Sprintf("%.1fh", task.EstimatedHours)
			if m.showProjection && taskIndex < len(completionTimes) {
				timeInfo += fmt.Sprintf("  %s", completionTimes[taskIndex].Format("3:04 PM"))
			}
			
			// Calculate available width more accurately
			// Account for: prefix, spacing, timeInfo, right border
			prefixWidth := visualWidth(prefix)
			timeInfoWidth := visualWidth(timeInfo)
			rightBorderWidth := 3 // "  ┃"
			
			availableForDesc := contentWidth - prefixWidth - timeInfoWidth - rightBorderWidth - 2 // 2 for spacing around dots
			
			// Truncate description if needed
			if visualWidth(description) > availableForDesc - 10 {
				description = truncateToWidth(description, availableForDesc - 10)
			}
			
			// Calculate dots needed
			descWidth := visualWidth(description)
			dotsNeeded := availableForDesc - descWidth
			if dotsNeeded < 3 {
				dotsNeeded = 3
			}
			
			// Construct the line with proper alignment
			mainLine := prefix + description + " " + strings.Repeat(".", dotsNeeded) + " " + timeInfo + "  ┃"
			section.WriteString(mainLine)
			section.WriteString("\n")

			// Metadata line
			var metadata []string
			
			// Priority
			if task.Priority == "H" {
				metadata = append(metadata, "High priority")
			} else if task.Priority == "M" {
				metadata = append(metadata, "Medium priority")
			} else if task.Priority == "L" {
				metadata = append(metadata, "Low priority")
			}
			
			// Energy level
			switch task.EnergyLevel {
			case EnergyHigh:
				metadata = append(metadata, "High energy")
			case EnergyMedium:
				metadata = append(metadata, "Medium energy")
			case EnergyLow:
				metadata = append(metadata, "Low energy")
			}
			
			// Due/Scheduled
			if task.IsDue {
				metadata = append(metadata, "Due today")
			}
			if task.IsScheduled {
				metadata = append(metadata, "Scheduled")
			}
			
			// Time slot
			if task.OptimalTimeSlot != "" && task.OptimalTimeSlot != "anytime" {
				metadata = append(metadata, "Best in " + task.OptimalTimeSlot)
			}
			
			if len(metadata) > 0 {
				metaStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("7"))
				metaText := strings.Join(metadata, " • ")
				metaStyled := metaStyle.Render(metaText)
				metaPrefix := "┃       "
				
				// Calculate padding needed
				metaPrefixWidth := visualWidth(metaPrefix)
				metaTextWidth := visualWidth(metaText) // Use unstyle text for width calculation
				rightBorderWidth := 1 // "┃"
				
				paddingNeeded := contentWidth - metaPrefixWidth - metaTextWidth - rightBorderWidth
				if paddingNeeded < 0 {
					paddingNeeded = 0
				}
				
				metaLine := metaPrefix + metaStyled + strings.Repeat(" ", paddingNeeded) + "┃"
				section.WriteString(metaLine)
				section.WriteString("\n")
			}
			
			// Add spacing between tasks
			if i < len(tasks)-1 {
				spaceWidth := contentWidth - 2
				if spaceWidth < 0 {
					spaceWidth = 0
				}
				section.WriteString("┃" + strings.Repeat(" ", spaceWidth) + "┃\n")
			}
		}
		
		// Add empty line at the end
		endSpaceWidth := contentWidth - 2
		if endSpaceWidth < 0 {
			endSpaceWidth = 0
		}
		section.WriteString("┃" + strings.Repeat(" ", endSpaceWidth) + "┃\n")
	}

	// Section footer
	footerWidth := contentWidth - 2
	if footerWidth < 0 {
		footerWidth = 0
	}
	footer := "┗" + strings.Repeat("━", footerWidth) + "┛"
	section.WriteString(headerStyle.Render(footer))

	return section.String()
}

// renderSummary renders the summary section with total hours and backlog info
func (m *PlanningModel) renderSummary(completionTimes []time.Time) string {
	var summary strings.Builder

	// Separator line
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	sepWidth := m.width - 2
	if sepWidth < 0 {
		sepWidth = 0
	}
	summary.WriteString(sepStyle.Render(strings.Repeat("─", sepWidth)))
	summary.WriteString("\n")
	
	// Summary information
	totalTasks := len(m.session.CriticalTasks) + len(m.session.ImportantTasks) + len(m.session.FlexibleTasks)
	
	var summaryParts []string
	if totalTasks > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d tasks planned", totalTasks))
		summaryParts = append(summaryParts, fmt.Sprintf("%.1f hours total", m.session.TotalHours))
		
		if m.showProjection && len(completionTimes) > 0 {
			lastCompletion := completionTimes[len(completionTimes)-1]
			summaryParts = append(summaryParts, fmt.Sprintf("Finish by %s", lastCompletion.Format("3:04 PM")))
		}
	}
	
	if len(m.session.BacklogTasks) > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d tasks in backlog", len(m.session.BacklogTasks)))
	}
	
	sumWidth := m.width - 2
	if sumWidth < 1 {
		sumWidth = 1
	}
	summaryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Align(lipgloss.Center).
		Width(sumWidth)
	
	if len(summaryParts) > 0 {
		summary.WriteString(summaryStyle.Render(strings.Join(summaryParts, " • ")))
	} else {
		summary.WriteString(summaryStyle.Render("No tasks planned"))
	}

	return summary.String()
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// promoteTaskToCritical promotes a task to the critical section
func (m *PlanningModel) promoteTaskToCritical(taskIndex int) {
	if taskIndex >= len(m.session.Tasks) {
		return
	}

	task := m.session.Tasks[taskIndex]
	
	// If already critical, no need to promote
	if task.Category == CategoryCritical {
		m.message = "Task is already in Critical section"
		return
	}

	// Update the task category
	task.Category = CategoryCritical
	
	// Re-organize tasks with the updated category
	allTasks := append(m.session.CriticalTasks, m.session.ImportantTasks...)
	allTasks = append(allTasks, m.session.FlexibleTasks...)
	
	// Update the task in the allTasks slice
	for i := range allTasks {
		if allTasks[i].UUID == task.UUID {
			allTasks[i].Category = CategoryCritical
			break
		}
	}
	
	// Re-organize and update display
	m.session.organizeTasks(allTasks)
	m.updateViewport()
	m.message = "Task promoted to Critical section"
}

// deferTask removes a task from today's plan (moves to backlog)
func (m *PlanningModel) deferTask(taskIndex int) {
	if taskIndex >= len(m.session.Tasks) {
		return
	}

	task := m.session.Tasks[taskIndex]
	
	// Move task to backlog
	m.session.BacklogTasks = append(m.session.BacklogTasks, task)
	
	// Remove from current section
	switch task.Category {
	case CategoryCritical:
		m.session.CriticalTasks = removeTaskFromSlice(m.session.CriticalTasks, task.UUID)
	case CategoryImportant:
		m.session.ImportantTasks = removeTaskFromSlice(m.session.ImportantTasks, task.UUID)
	case CategoryFlexible:
		m.session.FlexibleTasks = removeTaskFromSlice(m.session.FlexibleTasks, task.UUID)
	}
	
	// Rebuild the combined Tasks slice
	m.session.Tasks = make([]PlannedTask, 0, len(m.session.CriticalTasks)+len(m.session.ImportantTasks)+len(m.session.FlexibleTasks))
	m.session.Tasks = append(m.session.Tasks, m.session.CriticalTasks...)
	m.session.Tasks = append(m.session.Tasks, m.session.ImportantTasks...)
	m.session.Tasks = append(m.session.Tasks, m.session.FlexibleTasks...)
	
	// Adjust selected task index
	if m.selectedTask >= len(m.session.Tasks) && len(m.session.Tasks) > 0 {
		m.selectedTask = len(m.session.Tasks) - 1
	}
	
	// Recalculate totals and update display
	m.session.calculateTotals()
	m.updateViewport()
	m.message = fmt.Sprintf("Task deferred to backlog. %d tasks now in backlog", len(m.session.BacklogTasks))
}

// showBacklogMessage shows information about backlog tasks
func (m *PlanningModel) showBacklogMessage() {
	if len(m.session.BacklogTasks) == 0 {
		m.message = "No tasks in backlog"
		return
	}
	
	// Calculate backlog hours
	var backlogHours float64
	for _, task := range m.session.BacklogTasks {
		backlogHours += task.EstimatedHours
	}
	
	m.message = fmt.Sprintf("Backlog: %d tasks, %.1fh total. Use 'tasksh plan tomorrow' to review all tasks", 
		len(m.session.BacklogTasks), backlogHours)
}

// removeTaskFromSlice removes a task with the given UUID from a slice
func removeTaskFromSlice(tasks []PlannedTask, uuid string) []PlannedTask {
	for i, task := range tasks {
		if task.UUID == uuid {
			return append(tasks[:i], tasks[i+1:]...)
		}
	}
	return tasks
}

// Run starts the planning interface
func Run(horizon PlanningHorizon) error {
	// Create planning session
	session, err := NewPlanningSession(horizon)
	if err != nil {
		return fmt.Errorf("failed to create planning session: %w", err)
	}
	defer session.Close()

	// Load tasks
	if err := session.LoadTasks(); err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	if len(session.Tasks) == 0 {
		var horizonName string
		switch horizon {
		case HorizonTomorrow:
			horizonName = "tomorrow"
		case HorizonWeek:
			horizonName = "this week"
		}
		fmt.Printf("\nNo tasks found for %s.\n\n", horizonName)
		return nil
	}

	// Create and run Bubble Tea program
	model := NewPlanningModel(session)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run planning interface: %w", err)
	}

	return nil
}