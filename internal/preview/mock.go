package preview

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// MockView represents a mock TUI view for preview purposes
type MockView struct {
	Width  int
	Height int
	State  PreviewState
}

// RenderMainView renders the main task review interface
func (m *MockView) RenderMainView() string {
	var output strings.Builder

	// Status bar
	statusBar := m.renderStatusBar("[2 of 15]", "Implement user authentication system with OAuth2 support")
	output.WriteString(statusBar)
	output.WriteString("\n")

	// Task viewport
	viewport := m.renderViewport()
	output.WriteString(viewport)
	output.WriteString("\n")

	// Visual separator
	separator := m.renderSeparator()
	output.WriteString(separator)
	output.WriteString("\n")

	// Progress bar
	progress := m.renderProgressBar(0.13) // 2/15
	output.WriteString(progress)
	output.WriteString("\n")

	// Help text with top border
	helpSeparator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("238")).
		Render(strings.Repeat("─", m.Width))
	output.WriteString(helpSeparator)
	output.WriteString("\n")
	
	help := m.renderHelp(false)
	output.WriteString(help)

	return output.String()
}

// RenderHelpView renders the expanded help view
func (m *MockView) RenderHelpView() string {
	var output strings.Builder

	// Status bar
	statusBar := m.renderStatusBar("[2 of 15]", "Implement user authentication system with OAuth2 support")
	output.WriteString(statusBar)
	output.WriteString("\n")

	// Task viewport (smaller due to expanded help)
	viewport := m.renderViewportSmall()
	output.WriteString(viewport)
	output.WriteString("\n")

	// Help text (expanded)
	help := m.renderHelp(true)
	output.WriteString(help)

	return output.String()
}

// RenderDeleteConfirmation renders delete confirmation dialog
func (m *MockView) RenderDeleteConfirmation() string {
	var output strings.Builder

	// Status bar
	statusBar := m.renderStatusBar("[2 of 15]", "Implement user authentication system with OAuth2 support")
	output.WriteString(statusBar)
	output.WriteString("\n")

	// Confirmation dialog
	confirmStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1")).
		Padding(1, 2).
		Width(60).
		Align(lipgloss.Center)

	content := "Delete this task? This action cannot be undone.\n\nPress 'y' to confirm, 'n' to cancel"
	dialog := confirmStyle.Render(content)

	// Center the dialog
	output.WriteString(m.centerVertically(dialog, 10))

	return output.String()
}

// RenderModifyInput renders the modification input interface
func (m *MockView) RenderModifyInput() string {
	var output strings.Builder

	// Status bar
	statusBar := m.renderStatusBar("[2 of 15]", "Implement user authentication system with OAuth2 support")
	output.WriteString(statusBar)
	output.WriteString("\n")

	// Input dialog
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(1, 2).
		Width(70)

	// Input field
	inputField := "project:█"

	// Suggestions
	suggestions := `
Suggestions:
▶ project:webapp - Current project
  project:frontend - Frontend development
  project:backend - Backend development
  priority:H - High priority
  priority:M - Medium priority
  +urgent - Add urgent tag
  +blocked - Add blocked tag

↑↓: navigate  Tab: complete  ESC: cancel`

	content := fmt.Sprintf("Enter modification:\n\n%s\n%s", inputField, suggestions)
	dialog := inputStyle.Render(content)

	output.WriteString(m.centerVertically(dialog, 12))

	return output.String()
}

// Helper methods

func (m *MockView) renderStatusBar(progress, taskTitle string) string {
	// Improved status bar with better visual hierarchy
	
	// Progress indicator with visual enhancement
	progressStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("4")). // Blue background
		Foreground(lipgloss.Color("15")). // White text
		Bold(true).
		Padding(0, 1)
	
	// Task title with subtle background
	titleStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("235")). // Dark gray background
		Foreground(lipgloss.Color("7")).   // White text
		Padding(0, 1)
	
	// Context indicator
	contextStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")). // Slightly lighter gray
		Foreground(lipgloss.Color("6")).   // Cyan text
		Padding(0, 1)
	
	// Calculate available space
	progressRendered := progressStyle.Render(progress)
	contextText := "Context: work"
	contextRendered := contextStyle.Render(contextText)
	
	progressWidth := lipgloss.Width(progressRendered)
	contextWidth := lipgloss.Width(contextRendered)
	
	// Available space for title
	availableWidth := m.Width - progressWidth - contextWidth - 2
	
	// Truncate title if needed
	if len(taskTitle) > availableWidth {
		taskTitle = taskTitle[:availableWidth-3] + "..."
	}
	
	// Pad title to fill available space
	titlePadded := fmt.Sprintf("%-*s", availableWidth, taskTitle)
	titleRendered := titleStyle.Render(titlePadded)
	
	// Combine all elements
	return progressRendered + " " + titleRendered + " " + contextRendered
}

func (m *MockView) renderViewport() string {
	// Improved viewport with better visual hierarchy
	viewportStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("239")). // Subtle border
		Padding(1, 2).
		Width(m.Width - 2).
		Height(m.Height - 10) // Adjusted for separators

	// Content with improved formatting
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")). // Gray labels
		Width(12)
		
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")) // White values
	
	importantStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("3")). // Yellow for important
		Bold(true)
	
	var content strings.Builder
	
	// Description with emphasis
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Bold(true).
		Render("Description:"))
	content.WriteString("\n")
	content.WriteString(valueStyle.Render("Implement user authentication system with OAuth2 support"))
	content.WriteString("\n\n")
	
	// Metadata in two columns
	leftCol := []string{
		labelStyle.Render("Project:") + " " + valueStyle.Render("webapp"),
		labelStyle.Render("Priority:") + " " + importantStyle.Render("H"),
		labelStyle.Render("Status:") + " " + valueStyle.Render("pending"),
	}
	
	rightCol := []string{
		labelStyle.Render("Due:") + " " + importantStyle.Render("2024-12-25"),
		labelStyle.Render("Tags:") + " " + valueStyle.Render("+auth +security +backend"),
		"",
	}
	
	for i := 0; i < len(leftCol); i++ {
		content.WriteString(fmt.Sprintf("%-40s %s\n", leftCol[i], rightCol[i]))
	}
	
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("238")).
		Render("UUID: " + "task-uuid-1"))

	return viewportStyle.Render(content.String())
}

func (m *MockView) renderSeparator() string {
	// Subtle visual separator
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("236")) // Very dark gray
	
	// Create a gradient-like effect
	left := strings.Repeat("─", m.Width/4)
	middle := strings.Repeat("═", m.Width/2)
	right := strings.Repeat("─", m.Width/4)
	
	return separatorStyle.Render(left + middle + right)
}

func (m *MockView) renderViewportSmall() string {
	viewportStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(1, 2).
		Width(m.Width - 2).
		Height(m.Height - 12)

	content := `Description:
Implement user authentication system with OAuth2 support

Project: webapp
Priority: H`

	return viewportStyle.Render(content)
}

func (m *MockView) renderProgressBar(percent float64) string {
	progressStyle := lipgloss.NewStyle().
		Padding(0, 1)

	barWidth := 40
	filled := int(float64(barWidth) * percent)
	empty := barWidth - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	
	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("Review Progress: ")

	progressBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Render(bar)

	return progressStyle.Render(label + progressBar)
}

func (m *MockView) renderHelp(expanded bool) string {
	if !expanded {
		// Short help - SIMPLIFIED to show only primary actions
		primaryStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).  // Cyan for primary actions
			Bold(true)
		
		secondaryStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))  // Gray for secondary actions
		
		shortcuts := []string{
			primaryStyle.Render("r: review"),
			primaryStyle.Render("c: complete"),
			primaryStyle.Render("e: edit"),
			secondaryStyle.Render("s: skip"),
			secondaryStyle.Render("?: more"),
			secondaryStyle.Render("q: quit"),
		}
		
		// Add subtle hint about more options
		hint := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			Render(" (? for all shortcuts)")
		
		return strings.Join(shortcuts, " • ") + hint
	}

	// Expanded help - ORGANIZED by category
	categoryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("5")).  // Magenta for categories
		Bold(true)
	
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6"))  // Cyan for keys
	
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7"))  // White for descriptions
	
	var lines []string
	
	// Navigation
	lines = append(lines, categoryStyle.Render("Navigation:"))
	lines = append(lines, fmt.Sprintf("  %s %s  %s %s",
		keyStyle.Render("j/↓"), descStyle.Render("next task"),
		keyStyle.Render("k/↑"), descStyle.Render("previous task")))
	
	// Primary Actions
	lines = append(lines, "")
	lines = append(lines, categoryStyle.Render("Primary Actions:"))
	lines = append(lines, fmt.Sprintf("  %s %s  %s %s  %s %s",
		keyStyle.Render("r"), descStyle.Render("mark reviewed"),
		keyStyle.Render("c"), descStyle.Render("complete"),
		keyStyle.Render("e"), descStyle.Render("edit")))
	
	// Task Management
	lines = append(lines, "")
	lines = append(lines, categoryStyle.Render("Task Management:"))
	lines = append(lines, fmt.Sprintf("  %s %s  %s %s  %s %s  %s %s  %s %s",
		keyStyle.Render("m"), descStyle.Render("modify"),
		keyStyle.Render("d"), descStyle.Render("delete"),
		keyStyle.Render("w"), descStyle.Render("wait"),
		keyStyle.Render("u"), descStyle.Render("due date"),
		keyStyle.Render("s"), descStyle.Render("skip")))
	
	// Advanced (with AI if available)
	lines = append(lines, "")
	lines = append(lines, categoryStyle.Render("Advanced:"))
	lines = append(lines, fmt.Sprintf("  %s %s  %s %s  %s %s  %s %s",
		keyStyle.Render("x"), descStyle.Render("context"),
		keyStyle.Render("a"), descStyle.Render("AI analysis"),
		keyStyle.Render("p"), descStyle.Render("prompt agent"),
		keyStyle.Render("z"), descStyle.Render("undo")))
	
	// System
	lines = append(lines, "")
	lines = append(lines, categoryStyle.Render("System:"))
	lines = append(lines, fmt.Sprintf("  %s %s  %s %s",
		keyStyle.Render("?"), descStyle.Render("toggle help"),
		keyStyle.Render("q"), descStyle.Render("quit")))

	return strings.Join(lines, "\n")
}

func (m *MockView) centerVertically(content string, height int) string {
	lines := strings.Split(content, "\n")
	contentHeight := len(lines)
	
	if contentHeight >= height {
		return content
	}

	topPadding := (height - contentHeight) / 2
	bottomPadding := height - contentHeight - topPadding

	var output strings.Builder
	for i := 0; i < topPadding; i++ {
		output.WriteString("\n")
	}
	output.WriteString(content)
	for i := 0; i < bottomPadding; i++ {
		output.WriteString("\n")
	}

	return output.String()
}

// RenderAIAnalysis renders AI analysis view
func (m *MockView) RenderAIAnalysis() string {
	var output strings.Builder

	// Status bar
	statusBar := m.renderStatusBar("[2 of 15]", "Implement user authentication system with OAuth2 support")
	output.WriteString(statusBar)
	output.WriteString("\n")

	// AI Analysis content
	analysisStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")).
		Padding(1, 2).
		Width(m.Width - 2).
		Height(m.Height - 6)

	content := `🤖 AI Analysis

Summary:
This task involves implementing a critical security feature. Consider breaking it into smaller subtasks for better tracking.

Time Estimate:
4.5 hours - Based on similar authentication tasks, accounting for OAuth2 complexity

Suggestions:
1. priority: "H" → "H"
   Security features should maintain high priority (confidence: 95%)

2. tag: Add "+oauth2"
   Specific technology tag for better filtering (confidence: 85%)

3. subtask: Create subtasks for: 1) OAuth provider setup, 2) Token management, 3) User session handling
   Complex task benefits from breakdown (confidence: 90%)

Press ESC to return to task view`

	return output.String() + analysisStyle.Render(content)
}

// RenderContextSelect renders context selection
func (m *MockView) RenderContextSelect() string {
	var output strings.Builder

	// Status bar
	statusBar := m.renderStatusBar("[2 of 15]", "Implement user authentication system with OAuth2 support")
	output.WriteString(statusBar)
	output.WriteString("\n")

	// Context selection
	contextStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("6")).
		Padding(1, 2).
		Width(60).
		Align(lipgloss.Center)

	content := `Available Contexts:

▶ work (current)
  home
  personal
  urgent
  someday

↑↓: navigate  Enter: select  ESC: cancel`

	dialog := contextStyle.Render(content)
	output.WriteString(m.centerVertically(dialog, 12))

	return output.String()
}

// RenderCalendar renders calendar view
func (m *MockView) RenderCalendar(mode string) string {
	var output strings.Builder

	// Status bar  
	statusBar := m.renderStatusBar("[2 of 15]", "Implement user authentication system with OAuth2 support")
	output.WriteString(statusBar)
	output.WriteString("\n")

	// Calendar
	calendarStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("6")).
		Padding(1, 2).
		Width(50).
		Align(lipgloss.Center)

	// Get current month for display
	now := time.Now()
	monthYear := now.Format("January 2006")

	content := fmt.Sprintf(`%s

Su Mo Tu We Th Fr Sa
          1  2  3  4
 5  6  7  8  9 10 11
12 13 14 15 16 17 18
19 20 21 22 23 24 25
26 27 28 29 30 31

Select %s date
Tab: text input, x: remove %s, ESC: cancel`, monthYear, mode, mode)

	dialog := calendarStyle.Render(content)
	output.WriteString(m.centerVertically(dialog, 14))

	return output.String()
}