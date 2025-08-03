package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/emiller/tasksh/internal/tui/theme"
)

// HelpSystem is a modern, contextual help component
type HelpSystem struct {
	*BaseComponent
	
	// State
	visible     bool
	context     string
	
	// Content
	keyBindings []KeyBinding
	sections    []HelpSection
	
	// Styling
	theme       *theme.Theme
	styles      theme.HelpStyles
	
	// Layout
	columns     int
	maxWidth    int
}

// KeyBinding represents a key binding with help text
type KeyBinding struct {
	Key         string
	Description string
	Context     string
	Group       string
}

// HelpSection represents a section of help content
type HelpSection struct {
	Title       string
	Content     string
	KeyBindings []KeyBinding
}

// NewHelpSystem creates a new help system
func NewHelpSystem() *HelpSystem {
	base := NewBaseComponent()
	t := theme.GetTheme()
	
	return &HelpSystem{
		BaseComponent: base,
		visible:       false,
		context:       "default",
		keyBindings:   []KeyBinding{},
		sections:      []HelpSection{},
		theme:         t,
		styles:        t.Components.Help,
		columns:       2,
		maxWidth:      80,
	}
}

// Show displays the help system
func (hs *HelpSystem) Show() {
	hs.visible = true
}

// Hide hides the help system
func (hs *HelpSystem) Hide() {
	hs.visible = false
}

// Toggle toggles help visibility
func (hs *HelpSystem) Toggle() {
	hs.visible = !hs.visible
}

// IsVisible returns true if help is visible
func (hs *HelpSystem) IsVisible() bool {
	return hs.visible
}

// SetContext sets the current help context
func (hs *HelpSystem) SetContext(context string) {
	hs.context = context
}

// Context returns the current context
func (hs *HelpSystem) Context() string {
	return hs.context
}

// AddKeyBinding adds a key binding to the help system
func (hs *HelpSystem) AddKeyBinding(binding KeyBinding) {
	hs.keyBindings = append(hs.keyBindings, binding)
}

// AddKeyBindings adds multiple key bindings
func (hs *HelpSystem) AddKeyBindings(bindings []KeyBinding) {
	hs.keyBindings = append(hs.keyBindings, bindings...)
}

// SetKeyBindings replaces all key bindings
func (hs *HelpSystem) SetKeyBindings(bindings []KeyBinding) {
	hs.keyBindings = bindings
}

// AddSection adds a help section
func (hs *HelpSystem) AddSection(section HelpSection) {
	hs.sections = append(hs.sections, section)
}

// SetColumns sets the number of columns for key bindings
func (hs *HelpSystem) SetColumns(columns int) {
	if columns < 1 {
		columns = 1
	}
	if columns > 4 {
		columns = 4
	}
	hs.columns = columns
}

// SetMaxWidth sets the maximum width
func (hs *HelpSystem) SetMaxWidth(width int) {
	if width < 40 {
		width = 40
	}
	hs.maxWidth = width
}

// GetContextualBindings returns key bindings for the current context
func (hs *HelpSystem) GetContextualBindings() []KeyBinding {
	var contextual []KeyBinding
	
	for _, binding := range hs.keyBindings {
		if binding.Context == "" || binding.Context == hs.context {
			contextual = append(contextual, binding)
		}
	}
	
	return contextual
}

// Init implements tea.Model
func (hs *HelpSystem) Init() tea.Cmd {
	return hs.BaseComponent.Init()
}

// Update implements tea.Model
func (hs *HelpSystem) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Handle base component updates
	_, cmd := hs.BaseComponent.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "?", "h":
			hs.Toggle()
		case "esc":
			if hs.visible {
				hs.Hide()
			}
		}
		
	case ThemeChangeMsg:
		if t, ok := msg.Theme.(*theme.Theme); ok {
			hs.theme = t
			hs.styles = t.Components.Help
		}
	}
	
	return hs, tea.Batch(cmds...)
}

// View implements tea.Model
func (hs *HelpSystem) View() string {
	if !hs.visible {
		return ""
	}
	
	var content strings.Builder
	
	// Title
	title := fmt.Sprintf("Help - %s", strings.Title(hs.context))
	content.WriteString(hs.styles.Container.Render(
		hs.theme.Styles.Title.Render(title),
	))
	content.WriteString("\n\n")
	
	// Key bindings
	if len(hs.keyBindings) > 0 {
		content.WriteString(hs.renderKeyBindings())
		content.WriteString("\n")
	}
	
	// Sections
	for i, section := range hs.sections {
		if i > 0 {
			content.WriteString("\n")
		}
		content.WriteString(hs.renderSection(section))
	}
	
	// Footer
	footer := "Press ? to toggle help, esc to close"
	content.WriteString("\n")
	content.WriteString(hs.theme.Styles.Caption.Render(footer))
	
	// Apply container styling
	containerStyle := hs.styles.Container.
		Width(min(hs.maxWidth, hs.width-4)).
		MaxHeight(hs.height - 4)
	
	return containerStyle.Render(content.String())
}

// renderKeyBindings renders the key bindings in columns
func (hs *HelpSystem) renderKeyBindings() string {
	bindings := hs.GetContextualBindings()
	if len(bindings) == 0 {
		return ""
	}
	
	// Group bindings
	groups := make(map[string][]KeyBinding)
	var groupOrder []string
	
	for _, binding := range bindings {
		group := binding.Group
		if group == "" {
			group = "General"
		}
		
		if _, exists := groups[group]; !exists {
			groupOrder = append(groupOrder, group)
		}
		groups[group] = append(groups[group], binding)
	}
	
	var content strings.Builder
	
	for i, groupName := range groupOrder {
		if i > 0 {
			content.WriteString("\n")
		}
		
		// Group title
		content.WriteString(hs.theme.Styles.Subtitle.Render(groupName))
		content.WriteString("\n")
		
		// Group bindings
		groupBindings := groups[groupName]
		content.WriteString(hs.renderBindingGroup(groupBindings))
	}
	
	return content.String()
}

// renderBindingGroup renders a group of key bindings in columns
func (hs *HelpSystem) renderBindingGroup(bindings []KeyBinding) string {
	if len(bindings) == 0 {
		return ""
	}
	
	// Calculate column width
	maxKeyWidth := 0
	maxDescWidth := 0
	
	for _, binding := range bindings {
		if len(binding.Key) > maxKeyWidth {
			maxKeyWidth = len(binding.Key)
		}
		if len(binding.Description) > maxDescWidth {
			maxDescWidth = len(binding.Description)
		}
	}
	
	keyWidth := maxKeyWidth + 2
	descWidth := maxDescWidth + 2
	columnWidth := keyWidth + descWidth + 4
	
	// Calculate how many columns fit
	availableWidth := hs.maxWidth - 8 // Account for padding
	actualColumns := min(hs.columns, availableWidth/columnWidth)
	if actualColumns < 1 {
		actualColumns = 1
	}
	
	// Create columns
	var columns [][]string
	for i := 0; i < actualColumns; i++ {
		columns = append(columns, []string{})
	}
	
	// Distribute bindings across columns
	for i, binding := range bindings {
		col := i % actualColumns
		
		key := hs.styles.Key.Width(keyWidth).Render(binding.Key)
		desc := hs.styles.Description.Width(descWidth).Render(binding.Description)
		line := lipgloss.JoinHorizontal(lipgloss.Left, key, desc)
		
		columns[col] = append(columns[col], line)
	}
	
	// Balance column heights
	maxHeight := 0
	for _, col := range columns {
		if len(col) > maxHeight {
			maxHeight = len(col)
		}
	}
	
	for i := range columns {
		for len(columns[i]) < maxHeight {
			columns[i] = append(columns[i], "")
		}
	}
	
	// Join columns
	var rows []string
	for row := 0; row < maxHeight; row++ {
		var rowParts []string
		for col := 0; col < actualColumns; col++ {
			if row < len(columns[col]) {
				rowParts = append(rowParts, columns[col][row])
			}
		}
		rows = append(rows, strings.Join(rowParts, "  "))
	}
	
	return strings.Join(rows, "\n")
}

// renderSection renders a help section
func (hs *HelpSystem) renderSection(section HelpSection) string {
	var content strings.Builder
	
	// Section title
	if section.Title != "" {
		content.WriteString(hs.theme.Styles.Subtitle.Render(section.Title))
		content.WriteString("\n")
	}
	
	// Section content
	if section.Content != "" {
		content.WriteString(hs.theme.Styles.Body.Render(section.Content))
		content.WriteString("\n")
	}
	
	// Section key bindings
	if len(section.KeyBindings) > 0 {
		content.WriteString(hs.renderBindingGroup(section.KeyBindings))
	}
	
	return content.String()
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Predefined key bindings for common contexts
func DefaultTaskListBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "↑/k", Description: "Move up", Context: "tasklist", Group: "Navigation"},
		{Key: "↓/j", Description: "Move down", Context: "tasklist", Group: "Navigation"},
		{Key: "g", Description: "Go to top", Context: "tasklist", Group: "Navigation"},
		{Key: "G", Description: "Go to bottom", Context: "tasklist", Group: "Navigation"},
		{Key: "pgup", Description: "Page up", Context: "tasklist", Group: "Navigation"},
		{Key: "pgdn", Description: "Page down", Context: "tasklist", Group: "Navigation"},
		{Key: "/", Description: "Filter tasks", Context: "tasklist", Group: "Search"},
		{Key: "ctrl+l", Description: "Clear filter", Context: "tasklist", Group: "Search"},
		{Key: "enter", Description: "Select task", Context: "tasklist", Group: "Actions"},
		{Key: "space", Description: "Toggle selection", Context: "tasklist", Group: "Actions"},
		{Key: "d", Description: "Mark done", Context: "tasklist", Group: "Task Actions"},
		{Key: "x", Description: "Delete task", Context: "tasklist", Group: "Task Actions"},
		{Key: "m", Description: "Modify task", Context: "tasklist", Group: "Task Actions"},
		{Key: "w", Description: "Set waiting", Context: "tasklist", Group: "Task Actions"},
		{Key: "a", Description: "AI analysis", Context: "tasklist", Group: "AI"},
		{Key: "q", Description: "Quit", Context: "tasklist", Group: "General"},
		{Key: "?", Description: "Toggle help", Context: "tasklist", Group: "General"},
	}
}

func DefaultReviewBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "d", Description: "Mark done", Context: "review", Group: "Actions"},
		{Key: "x", Description: "Delete task", Context: "review", Group: "Actions"},
		{Key: "m", Description: "Modify task", Context: "review", Group: "Actions"},
		{Key: "w", Description: "Set waiting", Context: "review", Group: "Actions"},
		{Key: "s", Description: "Skip task", Context: "review", Group: "Actions"},
		{Key: "u", Description: "Undo last action", Context: "review", Group: "Actions"},
		{Key: "a", Description: "AI analysis", Context: "review", Group: "AI"},
		{Key: "c", Description: "Change context", Context: "review", Group: "Context"},
		{Key: "p", Description: "Set priority", Context: "review", Group: "Task Properties"},
		{Key: "t", Description: "Add/edit tags", Context: "review", Group: "Task Properties"},
		{Key: "due", Description: "Set due date", Context: "review", Group: "Task Properties"},
		{Key: "q", Description: "Quit review", Context: "review", Group: "General"},
		{Key: "?", Description: "Toggle help", Context: "review", Group: "General"},
	}
}

func DefaultPlanningBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "↑/k", Description: "Move up", Context: "planning", Group: "Navigation"},
		{Key: "↓/j", Description: "Move down", Context: "planning", Group: "Navigation"},
		{Key: "enter", Description: "Select/confirm", Context: "planning", Group: "Actions"},
		{Key: "space", Description: "Toggle selection", Context: "planning", Group: "Actions"},
		{Key: "n", Description: "Next step", Context: "planning", Group: "Planning"},
		{Key: "p", Description: "Previous step", Context: "planning", Group: "Planning"},
		{Key: "r", Description: "Refresh tasks", Context: "planning", Group: "Planning"},
		{Key: "s", Description: "Save plan", Context: "planning", Group: "Planning"},
		{Key: "q", Description: "Quit planning", Context: "planning", Group: "General"},
		{Key: "?", Description: "Toggle help", Context: "planning", Group: "General"},
	}
}

// CreateContextualHelp creates a help system with predefined bindings
func CreateContextualHelp(context string) *HelpSystem {
	help := NewHelpSystem()
	help.SetContext(context)
	
	switch context {
	case "tasklist":
		help.SetKeyBindings(DefaultTaskListBindings())
		help.AddSection(HelpSection{
			Title: "Task List",
			Content: "Navigate and manage your tasks. Use filters to find specific tasks quickly.",
		})
		
	case "review":
		help.SetKeyBindings(DefaultReviewBindings())
		help.AddSection(HelpSection{
			Title: "Task Review",
			Content: "Review tasks one by one. Make decisions about each task to keep your list organized.",
		})
		
	case "planning":
		help.SetKeyBindings(DefaultPlanningBindings())
		help.AddSection(HelpSection{
			Title: "Planning",
			Content: "Plan your day or week by selecting and organizing tasks based on your capacity and priorities.",
		})
		
	default:
		help.SetKeyBindings([]KeyBinding{
			{Key: "?", Description: "Toggle help", Group: "General"},
			{Key: "q", Description: "Quit", Group: "General"},
		})
	}
	
	return help
}

// Implement Component interface
var _ Component = (*HelpSystem)(nil)