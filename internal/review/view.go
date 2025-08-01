package review

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/emiller/tasksh/internal/taskwarrior"
)

// ViewRenderer handles all view rendering
type ViewRenderer struct {
	styles *Styles
}

// Styles contains all the lipgloss styles
type Styles struct {
	// Layout
	Container    lipgloss.Style
	StatusBar    lipgloss.Style
	Content      lipgloss.Style
	Footer       lipgloss.Style
	
	// Elements
	Title        lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	Muted        lipgloss.Style
	Success      lipgloss.Style
	Warning      lipgloss.Style
	Error        lipgloss.Style
	
	// Interactive
	Input        lipgloss.Style
	Button       lipgloss.Style
	Selected     lipgloss.Style
}

// NewViewRenderer creates a new view renderer
func NewViewRenderer() *ViewRenderer {
	return &ViewRenderer{
		styles: DefaultStyles(),
	}
}

// DefaultStyles returns the default style set
func DefaultStyles() *Styles {
	return &Styles{
		Container: lipgloss.NewStyle().
			Padding(0).
			Margin(0),
			
		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),
			
		Content: lipgloss.NewStyle().
			Padding(1, 2),
			
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1),
			
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")),
			
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")).
			Bold(true),
			
		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
			
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
			
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")),
			
		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
			
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),
			
		Input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(1, 2),
			
		Button: lipgloss.NewStyle().
			Background(lipgloss.Color("39")).
			Foreground(lipgloss.Color("231")).
			Padding(0, 2),
			
		Selected: lipgloss.NewStyle().
			Background(lipgloss.Color("39")).
			Foreground(lipgloss.Color("231")),
	}
}

// RenderMain renders the main review interface
func (r *ViewRenderer) RenderMain(m *ImprovedModel) string {
	sections := []string{
		r.renderStatusBar(m),
		r.renderContent(m),
		r.renderFooter(m),
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderStatusBar renders the top status bar
func (r *ViewRenderer) renderStatusBar(m *ImprovedModel) string {
	// Progress indicator
	progress := fmt.Sprintf("[%d/%d]", m.current+1, len(m.tasks))
	
	// Task description (truncated)
	var desc string
	if m.current < len(m.tasks) && m.current >= 0 {
		if task, ok := m.taskCache[m.tasks[m.current]]; ok {
			desc = truncate(task.Description, 50)
		}
	}
	
	// Loading indicator
	if m.loading {
		desc = m.spinner.View() + " Loading..."
	}
	
	// Combine elements
	left := fmt.Sprintf("%s %s", progress, desc)
	
	// Right side info
	right := fmt.Sprintf("Reviewed: %d", m.reviewed)
	
	// Calculate padding
	totalWidth := m.width
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	padding := totalWidth - leftWidth - rightWidth - 2
	
	if padding < 0 {
		padding = 1
	}
	
	bar := left + strings.Repeat(" ", padding) + right
	
	return r.styles.StatusBar.Width(m.width).Render(bar)
}

// renderContent renders the main content area
func (r *ViewRenderer) renderContent(m *ImprovedModel) string {
	// Show message if present
	if m.message != "" {
		msgStyle := r.styles.Success
		if m.error != nil {
			msgStyle = r.styles.Error
		}
		messageBox := msgStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(msgStyle.GetForeground()).
			Padding(1, 2).
			Margin(1, 2).
			Render(m.message)
		
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.viewport.View(),
			messageBox,
		)
	}
	
	return m.viewport.View()
}

// renderFooter renders the bottom help/progress area
func (r *ViewRenderer) renderFooter(m *ImprovedModel) string {
	// Progress bar
	progressBar := m.progress.ViewAs(float64(m.reviewed) / float64(len(m.tasks)))
	
	// Help text
	helpText := "r: review • c: complete • s: skip • ?: help • q: quit"
	
	footer := lipgloss.JoinVertical(
		lipgloss.Left,
		progressBar,
		r.styles.Footer.Render(helpText),
	)
	
	return footer
}

// RenderTask renders a task with nice formatting
func (r *ViewRenderer) RenderTask(task *taskwarrior.Task) string {
	if task == nil {
		return r.styles.Muted.Render("No task selected")
	}
	
	var sections []string
	
	// Title with priority
	title := task.Description
	if task.Priority != "" {
		priorityBadge := r.getPriorityStyle(task.Priority).
			Padding(0, 1).
			Render(task.Priority)
		title = priorityBadge + " " + title
	}
	sections = append(sections, r.styles.Title.Render(title))
	sections = append(sections, "")
	
	// Metadata grid
	metadata := r.renderMetadata(task)
	if metadata != "" {
		sections = append(sections, metadata)
		sections = append(sections, "")
	}
	
	// Tags - check if task has tags field
	// Note: taskwarrior.Task might not have Tags field exposed
	
	// UUID (muted)
	sections = append(sections, r.styles.Muted.Render("ID: "+task.UUID[:8]))
	
	return strings.Join(sections, "\n")
}

// renderMetadata renders task metadata in a grid
func (r *ViewRenderer) renderMetadata(task *taskwarrior.Task) string {
	var items []string
	
	// Project
	if task.Project != "" {
		items = append(items, r.renderField("Project", task.Project))
	}
	
	// Status
	if task.Status != "" {
		statusStyle := r.styles.Value
		if task.Status == "completed" {
			statusStyle = r.styles.Success
		}
		items = append(items, r.renderFieldStyled("Status", task.Status, statusStyle))
	}
	
	// Due date
	if task.Due != "" {
		dueStyle := r.styles.Value
		// Could add logic to color based on urgency
		items = append(items, r.renderFieldStyled("Due", task.Due, dueStyle))
	}
	
	// Wait field may not exist in taskwarrior.Task
	// Comment out for now since we don't have access to Wait field
	// if task.Wait != "" {
	// 	items = append(items, r.renderFieldStyled("Wait", task.Wait, r.styles.Warning))
	// }
	
	// Create two-column layout
	if len(items) == 0 {
		return ""
	}
	
	var rows []string
	for i := 0; i < len(items); i += 2 {
		if i+1 < len(items) {
			row := items[i] + "    " + items[i+1]
			rows = append(rows, row)
		} else {
			rows = append(rows, items[i])
		}
	}
	
	return strings.Join(rows, "\n")
}

// renderField renders a label-value pair
func (r *ViewRenderer) renderField(label, value string) string {
	return r.renderFieldStyled(label, value, r.styles.Value)
}

// renderFieldStyled renders a label-value pair with custom value style
func (r *ViewRenderer) renderFieldStyled(label, value string, valueStyle lipgloss.Style) string {
	l := r.styles.Label.Render(label + ":")
	v := valueStyle.Render(value)
	return l + " " + v
}

// renderTags renders tags as badges
// func (r *ViewRenderer) renderTags(tags []string) string {
// 	var badges []string
// 	
// 	tagStyle := lipgloss.NewStyle().
// 		Background(lipgloss.Color("235")).
// 		Foreground(lipgloss.Color("183")).
// 		Padding(0, 1)
// 	
// 	for _, tag := range tags {
// 		badges = append(badges, tagStyle.Render(tag))
// 	}
// 	
// 	return r.styles.Label.Render("Tags: ") + strings.Join(badges, " ")
// }

// RenderHelp renders the help screen
func (r *ViewRenderer) RenderHelp(keys KeyMap, width, height int) string {
	helpStyle := lipgloss.NewStyle().
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("241"))
	
	content := r.styles.Title.Render("Keyboard Shortcuts") + "\n\n"
	
	// Navigation
	content += r.styles.Label.Render("Navigation") + "\n"
	content += r.renderHelpItem("j/↓", "Next task") + "\n"
	content += r.renderHelpItem("k/↑", "Previous task") + "\n\n"
	
	// Actions
	content += r.styles.Label.Render("Actions") + "\n"
	content += r.renderHelpItem("r", "Review task") + "\n"
	content += r.renderHelpItem("c", "Complete task") + "\n"
	content += r.renderHelpItem("e", "Edit task") + "\n"
	content += r.renderHelpItem("d", "Delete task") + "\n"
	content += r.renderHelpItem("s", "Skip task") + "\n\n"
	
	// System
	content += r.styles.Label.Render("System") + "\n"
	content += r.renderHelpItem("?", "Toggle help") + "\n"
	content += r.renderHelpItem("q", "Quit") + "\n"
	
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
		helpStyle.Render(content))
}

// renderHelpItem renders a help item
func (r *ViewRenderer) renderHelpItem(key, desc string) string {
	k := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Width(10).
		Render(key)
	d := r.styles.Value.Render(desc)
	return "  " + k + d
}

// RenderError renders an error screen
func (r *ViewRenderer) RenderError(err error, width, height int) string {
	if err == nil {
		return ""
	}
	
	errorBox := r.styles.Error.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(2, 4).
		Render("Error: " + err.Error())
	
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, errorBox)
}

// RenderInput renders an input dialog
func (r *ViewRenderer) RenderInput(input textinput.Model, prompt string, width, height int) string {
	inputBox := r.styles.Input.
		Width(60).
		Render(prompt + "\n\n" + input.View() + "\n\nEnter: confirm • Esc: cancel")
	
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, inputBox)
}

// Helper functions

// getPriorityStyle returns style for priority
func (r *ViewRenderer) getPriorityStyle(priority string) lipgloss.Style {
	switch priority {
	case "H":
		return lipgloss.NewStyle().
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("231")).
			Bold(true)
	case "M":
		return lipgloss.NewStyle().
			Background(lipgloss.Color("214")).
			Foreground(lipgloss.Color("231"))
	case "L":
		return lipgloss.NewStyle().
			Background(lipgloss.Color("82")).
			Foreground(lipgloss.Color("231"))
	default:
		return r.styles.Value
	}
}

// truncate truncates a string to the given length
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}