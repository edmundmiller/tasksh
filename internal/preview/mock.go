package preview

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// stripANSISequences removes ANSI escape sequences from a string for width calculation
func stripANSISequences(s string) string {
	// Simple ANSI stripping
	for {
		start := strings.Index(s, "\x1b[")
		if start == -1 {
			break
		}
		end := strings.IndexByte(s[start:], 'm')
		if end == -1 {
			break
		}
		s = s[:start] + s[start+end+1:]
	}
	return s
}

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

	// Only add separator and progress bar if we have enough height
	if m.Height > 15 {
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
			Render(strings.Repeat("-", m.Width))
		output.WriteString(helpSeparator)
		output.WriteString("\n")
	} else {
		output.WriteString("\n")
	}
	
	help := m.renderHelp(false)
	output.WriteString(help)

	// Ensure total output doesn't exceed terminal height
	lines := strings.Split(output.String(), "\n")
	if len(lines) > m.Height {
		lines = lines[:m.Height]
		return strings.Join(lines, "\n")
	}

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

	// Help text (expanded) - calculate remaining height for help
	currentLines := strings.Split(output.String(), "\n")
	remainingHeight := m.Height - len(currentLines) - 1 // -1 for safety margin
	if remainingHeight < 3 {
		remainingHeight = 3
	}
	
	// Get help text with height constraint
	help := m.renderHelpWithHeight(remainingHeight)
	output.WriteString(help)

	// Ensure total output doesn't exceed terminal height
	lines := strings.Split(output.String(), "\n")
	if len(lines) > m.Height {
		lines = lines[:m.Height]
		return strings.Join(lines, "\n")
	}

	return output.String()
}

// RenderDeleteConfirmation renders delete confirmation dialog
func (m *MockView) RenderDeleteConfirmation() string {
	var output strings.Builder

	// Status bar
	statusBar := m.renderStatusBar("[2 of 15]", "Implement user authentication system with OAuth2 support")
	output.WriteString(statusBar)
	output.WriteString("\n")

	// Confirmation dialog - calculate content width
	// The lipgloss Width() sets content width, total width = content + border(2) + padding(4)
	dialogWidth := m.Width - 6
	if dialogWidth < 24 {
		dialogWidth = 24
	}
	
	confirmStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1")).
		Padding(1, 2).
		Width(dialogWidth).
		Align(lipgloss.Center)

	content := "Delete this task? This action cannot be undone.\n\nPress 'y' to confirm, 'n' to cancel"
	dialog := confirmStyle.Render(content)
	
	// Force width constraint by truncating each line if necessary
	lines := strings.Split(dialog, "\n")
	for i, line := range lines {
		cleanLine := stripANSISequences(line)
		if len(cleanLine) > m.Width {
			lines[i] = line[:m.Width]
		}
	}
	dialog = strings.Join(lines, "\n")

	// Center the dialog within available height
	availableHeight := m.Height - 3 // Account for status bar
	output.WriteString(m.centerVertically(dialog, availableHeight))

	return output.String()
}

// RenderModifyInput renders the modification input interface
func (m *MockView) RenderModifyInput() string {
	var output strings.Builder

	// Status bar
	statusBar := m.renderStatusBar("[2 of 15]", "Implement user authentication system with OAuth2 support")
	output.WriteString(statusBar)
	output.WriteString("\n")

	// Input dialog - calculate content width
	// The lipgloss Width() sets content width, total width = content + border(2) + padding(4)
	dialogWidth := m.Width - 6
	if dialogWidth < 30 {
		dialogWidth = 30
	}
	
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(1, 2).
		Width(dialogWidth)

	// Input field
	inputField := "project:█"

	// Suggestions - adjust for small terminals
	var suggestions string
	if m.Height < 15 {
		// Minimal suggestions for small terminals
		suggestions = "\nSuggestions:\n▶ project:webapp\n  priority:H\n\n↑↓: navigate  Tab: complete  ESC: cancel"
	} else {
		// Full suggestions for larger terminals
		suggestions = `
Suggestions:
▶ project:webapp - Current project
  project:frontend - Frontend development
  project:backend - Backend development
  priority:H - High priority
  priority:M - Medium priority
  +urgent - Add urgent tag
  +blocked - Add blocked tag

↑↓: navigate  Tab: complete  ESC: cancel`
	}

	content := fmt.Sprintf("Enter modification:\n\n%s\n%s", inputField, suggestions)
	dialog := inputStyle.Render(content)
	
	// Force width constraint by truncating each line if necessary
	lines := strings.Split(dialog, "\n")
	for i, line := range lines {
		cleanLine := stripANSISequences(line)
		if len(cleanLine) > m.Width {
			lines[i] = line[:m.Width]
		}
	}
	dialog = strings.Join(lines, "\n")

	// Center within available height
	availableHeight := m.Height - 3
	output.WriteString(m.centerVertically(dialog, availableHeight))

	return output.String()
}

// Helper methods

func (m *MockView) renderStatusBar(progress, taskTitle string) string {
	// Create status bar that matches the expected golden snapshot format
	// Format: " [2 of 15]   Implement user authentication system with OAuth2 ...   Context: work "
	
	// For small terminals, use simple format
	if m.Width < 50 {
		simple := fmt.Sprintf("%s %s", progress, taskTitle)
		if len(simple) > m.Width {
			simple = simple[:m.Width-3] + "..."
		}
		return simple
	}
	
	// Calculate space allocation
	contextText := "Context: work"
	progressLen := len(progress)
	contextLen := len(contextText)
	
	// Reserve space for: " " + progress + "   " + title + "   " + context + " "
	// That's 1 + progressLen + 3 + titleLen + 3 + contextLen + 1 = titleLen + progressLen + contextLen + 8
	reservedSpace := progressLen + contextLen + 8
	availableForTitle := m.Width - reservedSpace
	
	if availableForTitle < 10 {
		// Fallback for very narrow terminals
		simple := fmt.Sprintf("%s %s", progress, taskTitle)
		if len(simple) > m.Width {
			simple = simple[:m.Width-3] + "..."
		}
		return simple
	}
	
	// Truncate title if needed
	if len(taskTitle) > availableForTitle {
		taskTitle = taskTitle[:availableForTitle-3] + "..."
	}
	
	// Build the status bar with exact spacing to match expected format
	statusBar := fmt.Sprintf(" %s   %-*s   %s ", progress, availableForTitle, taskTitle, contextText)
	
	// Ensure it doesn't exceed width
	if len(statusBar) > m.Width {
		statusBar = statusBar[:m.Width]
	}
	
	return statusBar
}

func (m *MockView) renderViewport() string {
	// For very small terminals, use plain text without lipgloss styling
	if m.Width < 70 {
		return m.renderPlainViewport()
	}
	
	// Calculate viewport content width: terminal width minus border (2) and padding (4)
	// The lipgloss Width() sets the content width, not the total rendered width
	viewportWidth := m.Width - 6 // border (2) + padding (4)
	if viewportWidth < 10 {
		viewportWidth = 10
	}
	
	viewportHeight := m.Height - 8
	if viewportHeight < 5 {
		viewportHeight = 5
	}
	
	// Create viewport style that exactly matches terminal width
	viewportStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("239")). // Subtle border
		Padding(1, 2).
		Width(viewportWidth).
		Height(viewportHeight)

	// Build content to match golden snapshot format
	var content strings.Builder
	
	// Description
	content.WriteString("Description:\n")
	content.WriteString("Implement user authentication system with OAuth2 support\n\n")
	
	// Calculate column layout to match golden snapshot
	// Golden snapshot shows:
	// "  Project:     webapp                      Due:         2024-12-25            │"
	// "  Priority:    H                           Tags:        +auth +security       │"
	// "  +backend                                                                    │"
	// "  Status:      pending                                                        │"
	// "                                                                              │"
	// "  UUID: task-uuid-1                                                           │"
	
	contentWidth := viewportWidth - 4 // Account for padding
	leftColWidth := contentWidth / 2
	
	// Row 1: Project and Due
	left1 := fmt.Sprintf("  Project:     webapp")
	right1 := fmt.Sprintf("Due:         2024-12-25")
	line1 := fmt.Sprintf("%-*s %s", leftColWidth, left1, right1)
	content.WriteString(line1 + "\n")
	
	// Row 2: Priority and Tags
	left2 := fmt.Sprintf("  Priority:    H")
	right2 := fmt.Sprintf("Tags:        +auth +security")
	line2 := fmt.Sprintf("%-*s %s", leftColWidth, left2, right2)
	content.WriteString(line2 + "\n")
	
	// Row 3: Additional tag
	content.WriteString("  +backend\n")
	
	// Row 4: Status
	content.WriteString("  Status:      pending\n\n")
	
	// Row 5: UUID
	content.WriteString("  UUID: task-uuid-1\n")

	rendered := viewportStyle.Render(content.String())
	
	// Force width constraint by truncating each line if necessary
	lines := strings.Split(rendered, "\n")
	for i, line := range lines {
		cleanLine := stripANSISequences(line)
		if len(cleanLine) > m.Width {
			// This shouldn't happen with proper lipgloss usage, but as a fallback
			lines[i] = line[:m.Width]
		}
	}
	
	return strings.Join(lines, "\n")
}

// renderPlainViewport renders a simple text viewport for small terminals
func (m *MockView) renderPlainViewport() string {
	var lines []string
	
	// Create a simple border using ASCII characters
	borderWidth := m.Width - 2
	if borderWidth < 10 {
		borderWidth = 10
	}
	
	topBorder := "+" + strings.Repeat("-", borderWidth-2) + "+"
	bottomBorder := topBorder
	
	lines = append(lines, topBorder)
	
	// Description
	desc := "OAuth2 authentication system"
	if len(desc) > borderWidth-4 {
		desc = desc[:borderWidth-7] + "..."
	}
	lines = append(lines, fmt.Sprintf("| %-*s |", borderWidth-4, desc))
	
	lines = append(lines, fmt.Sprintf("| %-*s |", borderWidth-4, ""))
	
	// Project and Priority on same line for small terminals
	info := "Project: webapp, Priority: H"
	if len(info) > borderWidth-4 {
		info = "webapp, H"
	}
	lines = append(lines, fmt.Sprintf("| %-*s |", borderWidth-4, info))
	
	lines = append(lines, fmt.Sprintf("| %-*s |", borderWidth-4, "Status: pending"))
	
	// Add padding lines to fill height if needed
	currentHeight := len(lines) + 1 // +1 for bottom border
	targetHeight := m.Height - 8
	if targetHeight < currentHeight {
		targetHeight = currentHeight
	}
	
	for len(lines) < targetHeight-1 {
		lines = append(lines, fmt.Sprintf("| %-*s |", borderWidth-4, ""))
	}
	
	lines = append(lines, bottomBorder)
	
	return strings.Join(lines, "\n")
}

func (m *MockView) renderSeparator() string {
	// Always ensure separator doesn't exceed terminal width
	if m.Width < 70 {
		// For small terminals, use simple ASCII separator
		sep := strings.Repeat("-", m.Width)
		return sep
	}
	
	// Create separator that matches golden snapshot exactly
	// Golden snapshot shows: "────────────────────════════════════════════════════────────────────────────────"
	
	// Calculate widths for the three sections
	leftWidth := m.Width / 4
	middleWidth := m.Width / 2
	rightWidth := m.Width - leftWidth - middleWidth
	
	left := strings.Repeat("-", leftWidth)
	middle := strings.Repeat("=", middleWidth)
	right := strings.Repeat("-", rightWidth)
	
	result := left + middle + right
	// Ensure it doesn't exceed width as fallback
	if len(result) > m.Width {
		result = result[:m.Width]
	}
	
	return result
}

func (m *MockView) renderViewportSmall() string {
	// For small terminals, use plain text
	if m.Width < 70 {
		return m.renderPlainViewportSmall()
	}
	
	// Calculate viewport content width: terminal width minus border (2) and padding (4)
	// The lipgloss Width() sets the content width, not the total rendered width
	viewportWidth := m.Width - 6 // border (2) + padding (4)
	if viewportWidth < 20 {
		viewportWidth = 20
	}
	
	viewportHeight := m.Height - 12
	if viewportHeight < 3 {
		viewportHeight = 3
	}
	
	viewportStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(1, 2).
		Width(viewportWidth).
		Height(viewportHeight)

	// Truncate description to fit width
	desc := "Implement user authentication system with OAuth2 support"
	maxDescWidth := viewportWidth - 6
	if len(desc) > maxDescWidth {
		desc = desc[:maxDescWidth-3] + "..."
	}
	
	content := fmt.Sprintf("Description:\n%s\n\nProject: webapp\nPriority: H", desc)

	rendered := viewportStyle.Render(content)
	
	// Force width constraint by truncating each line if necessary
	lines := strings.Split(rendered, "\n")
	for i, line := range lines {
		cleanLine := stripANSISequences(line)
		if len(cleanLine) > m.Width {
			lines[i] = line[:m.Width]
		}
	}
	
	return strings.Join(lines, "\n")
}

// renderPlainViewportSmall renders a simple small viewport for small terminals
func (m *MockView) renderPlainViewportSmall() string {
	var lines []string
	
	// Create a simple border using ASCII characters
	borderWidth := m.Width - 2
	if borderWidth < 10 {
		borderWidth = 10
	}
	
	topBorder := "+" + strings.Repeat("-", borderWidth-2) + "+"
	bottomBorder := topBorder
	
	lines = append(lines, topBorder)
	
	// Simple content for small viewport
	lines = append(lines, fmt.Sprintf("| %-*s |", borderWidth-4, "OAuth2 system"))
	lines = append(lines, fmt.Sprintf("| %-*s |", borderWidth-4, "webapp, H"))
	
	lines = append(lines, bottomBorder)
	
	return strings.Join(lines, "\n")
}

func (m *MockView) renderProgressBar(percent float64) string {
	// For small terminals, use simple progress bar
	if m.Width < 70 {
		return m.renderPlainProgressBar(percent)
	}
	
	// Create progress bar that matches golden snapshot exactly
	// Golden snapshot shows: " Review Progress: █████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ "
	
	labelText := " Review Progress: "
	labelWidth := len(labelText)
	
	// Reserve space for trailing space
	barWidth := m.Width - labelWidth - 1
	if barWidth < 10 {
		barWidth = 10
	}
	
	filled := int(float64(barWidth) * percent)
	empty := barWidth - filled

	// Use ASCII characters to avoid Unicode width issues
	bar := strings.Repeat("=", filled) + strings.Repeat("-", empty)
	
	result := labelText + bar + " "
	
	// Ensure total width doesn't exceed terminal width
	if len(result) > m.Width {
		result = result[:m.Width]
	}
	
	return result
}

// renderPlainProgressBar renders a simple progress bar for small terminals
func (m *MockView) renderPlainProgressBar(percent float64) string {
	labelText := "Progress: "
	labelWidth := len(labelText)
	barWidth := m.Width - labelWidth - 2
	if barWidth < 5 {
		barWidth = 5
	}
	
	filled := int(float64(barWidth) * percent)
	empty := barWidth - filled
	
	bar := strings.Repeat("=", filled) + strings.Repeat("-", empty)
	
	return labelText + bar
}

func (m *MockView) renderHelp(expanded bool) string {
	if !expanded {
		// Short help - Use cyan color for primary actions as expected by test
		primaryColor := "\x1b[36m" // Cyan for primary actions (not bold to match test expectation)
		secondaryColor := "\x1b[90m"      // Bright black (gray) for secondary actions
		resetColor := "\x1b[0m"
		
		shortcuts := []string{
			primaryColor + "r: review" + resetColor,
			primaryColor + "c: complete" + resetColor,
			primaryColor + "e: edit" + resetColor,
			secondaryColor + "s: skip" + resetColor,
			secondaryColor + "?: more" + resetColor,
			secondaryColor + "q: quit" + resetColor,
		}
		
		// Build help text that fits within terminal width
		helpText := strings.Join(shortcuts, " • ")
		hint := "\x1b[3m\x1b[90m (? for all shortcuts)" + resetColor // Italic gray
		
		fullText := helpText + hint
		
		// Truncate if too long for terminal width (measure without ANSI codes)
		cleanText := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(fullText, primaryColor, ""), secondaryColor, ""), resetColor, "")
		cleanText = strings.ReplaceAll(strings.ReplaceAll(cleanText, "\x1b[3m\x1b[90m", ""), "\x1b[0m", "")
		
		if len(cleanText) > m.Width {
			// Fallback to just primary actions without hint
			primaryOnly := []string{
				primaryColor + "r: review" + resetColor,
				primaryColor + "c: complete" + resetColor,
				secondaryColor + "?: more" + resetColor,
			}
			return strings.Join(primaryOnly, " • ")
		}
		
		return fullText
	}

	// Expanded help - ORGANIZED by category
	// Use manual ANSI codes to ensure color consistency in tests
	categoryColor := "\x1b[1m\x1b[35m" // Bold magenta for categories
	keyColor := "\x1b[36m"              // Cyan for keys
	descColor := "\x1b[37m"             // White for descriptions
	resetColor := "\x1b[0m"
	
	var lines []string
	
	// Calculate available lines based on terminal height
	// Account for status bar (1), viewport (varies), and padding
	availableHeight := m.Height - 12 // Conservative estimate 
	if availableHeight < 5 {
		availableHeight = 5
	}
	
	// For small terminals, show only essential help
	if availableHeight < 8 {
		lines = append(lines, categoryColor+"Quick Help:"+resetColor)
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s",
			keyColor, "r", resetColor, descColor, "review", resetColor,
			keyColor, "c", resetColor, descColor, "complete", resetColor))
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s",
			keyColor, "?", resetColor, descColor, "toggle help", resetColor,
			keyColor, "q", resetColor, descColor, "quit", resetColor))
	} else if availableHeight < 12 {
		// Medium terminals - basic categories
		lines = append(lines, categoryColor+"Navigation:"+resetColor)
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s",
			keyColor, "j/↓", resetColor, descColor, "next task", resetColor,
			keyColor, "k/↑", resetColor, descColor, "previous task", resetColor))
		
		lines = append(lines, "")
		lines = append(lines, categoryColor+"Primary Actions:"+resetColor)
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s",
			keyColor, "r", resetColor, descColor, "review", resetColor,
			keyColor, "c", resetColor, descColor, "complete", resetColor))
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s", keyColor, "q", resetColor, descColor, "quit", resetColor))
	} else {
		// Full help for larger terminals
		// Navigation
		lines = append(lines, categoryColor+"Navigation:"+resetColor)
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s",
			keyColor, "j/↓", resetColor, descColor, "next task", resetColor,
			keyColor, "k/↑", resetColor, descColor, "previous task", resetColor))
		
		// Primary Actions
		lines = append(lines, "")
		lines = append(lines, categoryColor+"Primary Actions:"+resetColor)
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s  %s%s%s %s%s%s",
			keyColor, "r", resetColor, descColor, "mark reviewed", resetColor,
			keyColor, "c", resetColor, descColor, "complete", resetColor,
			keyColor, "e", resetColor, descColor, "edit", resetColor))
		
		// Task Management
		lines = append(lines, "")
		lines = append(lines, categoryColor+"Task Management:"+resetColor)
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s  %s%s%s %s%s%s",
			keyColor, "m", resetColor, descColor, "modify", resetColor,
			keyColor, "d", resetColor, descColor, "delete", resetColor,
			keyColor, "s", resetColor, descColor, "skip", resetColor))
		
		// System
		lines = append(lines, "")
		lines = append(lines, categoryColor+"System:"+resetColor)
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s",
			keyColor, "?", resetColor, descColor, "toggle help", resetColor,
			keyColor, "q", resetColor, descColor, "quit", resetColor))
	}
	
	// Ensure lines fit within available height
	if len(lines) > availableHeight {
		lines = lines[:availableHeight]
	}

	return strings.Join(lines, "\n")
}

// renderHelpWithHeight renders help text within a specified height limit
func (m *MockView) renderHelpWithHeight(maxHeight int) string {
	// Use the same logic as expanded help but with height constraint
	categoryColor := "\x1b[1m\x1b[35m" // Bold magenta for categories
	keyColor := "\x1b[36m"              // Cyan for keys
	descColor := "\x1b[37m"             // White for descriptions
	resetColor := "\x1b[0m"
	
	var lines []string
	
	// For very limited height, show only essential help
	if maxHeight < 5 {
		lines = append(lines, categoryColor+"Quick Help:"+resetColor)
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s",
			keyColor, "r", resetColor, descColor, "review", resetColor,
			keyColor, "c", resetColor, descColor, "complete", resetColor))
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s", keyColor, "q", resetColor, descColor, "quit", resetColor))
	} else if maxHeight < 8 {
		// Medium height - basic categories
		lines = append(lines, categoryColor+"Navigation:"+resetColor)
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s",
			keyColor, "j/↓", resetColor, descColor, "next", resetColor,
			keyColor, "k/↑", resetColor, descColor, "prev", resetColor))
		
		lines = append(lines, "")
		lines = append(lines, categoryColor+"Actions:"+resetColor)
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s",
			keyColor, "r", resetColor, descColor, "review", resetColor,
			keyColor, "c", resetColor, descColor, "complete", resetColor))
	} else {
		// Full help for reasonable height
		lines = append(lines, categoryColor+"Navigation:"+resetColor)
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s",
			keyColor, "j/↓", resetColor, descColor, "next task", resetColor,
			keyColor, "k/↑", resetColor, descColor, "previous task", resetColor))
		
		lines = append(lines, "")
		lines = append(lines, categoryColor+"Primary Actions:"+resetColor)
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s",
			keyColor, "r", resetColor, descColor, "review", resetColor,
			keyColor, "c", resetColor, descColor, "complete", resetColor))
		
		// Add more sections if height allows
		if maxHeight > 12 {
			lines = append(lines, "")
			lines = append(lines, categoryColor+"Task Management:"+resetColor)
			lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s  %s%s%s %s%s%s",
				keyColor, "m", resetColor, descColor, "modify", resetColor,
				keyColor, "d", resetColor, descColor, "delete", resetColor))
		}
		
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  %s%s%s %s%s%s", keyColor, "q", resetColor, descColor, "quit", resetColor))
	}
	
	// Ensure lines fit within height constraint
	if len(lines) > maxHeight {
		lines = lines[:maxHeight]
	}

	return strings.Join(lines, "\n")
}

func (m *MockView) centerVertically(content string, height int) string {
	lines := strings.Split(content, "\n")
	contentHeight := len(lines)
	
	// If content is too tall, truncate it to fit
	if contentHeight > height {
		lines = lines[:height]
		content = strings.Join(lines, "\n")
		contentHeight = height
	}
	
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

	// AI Analysis content - calculate content width
	// The lipgloss Width() sets content width, total width = content + border(2) + padding(4)
	analysisWidth := m.Width - 6
	if analysisWidth < 24 {
		analysisWidth = 24
	}
	
	analysisHeight := m.Height - 6
	if analysisHeight < 5 {
		analysisHeight = 5
	}
	
	analysisStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")).
		Padding(1, 2).
		Width(analysisWidth).
		Height(analysisHeight)

	// Adjust content based on available space
	var content string
	if analysisHeight < 10 {
		// Minimal content for small terminals
		content = `🤖 AI Analysis

Summary:
Critical security feature. Consider breaking into subtasks.

Time Estimate: 4.5 hours

Press ESC to return`
	} else {
		// Full content for larger terminals
		content = `🤖 AI Analysis

Summary:
This task involves implementing a critical security feature. Consider breaking it into smaller subtasks for better tracking.

Time Estimate:
4.5 hours - Based on similar authentication tasks

Suggestions:
1. priority: "H" → "H" (confidence: 95%)
2. tag: Add "+oauth2" (confidence: 85%)
3. Create subtasks for OAuth setup, token management, user sessions

Press ESC to return to task view`
	}

	analysis := analysisStyle.Render(content)
	
	// Force width constraint by truncating each line if necessary
	lines := strings.Split(analysis, "\n")
	for i, line := range lines {
		cleanLine := stripANSISequences(line)
		if len(cleanLine) > m.Width {
			lines[i] = line[:m.Width]
		}
	}
	analysis = strings.Join(lines, "\n")
	
	return output.String() + analysis
}

// RenderContextSelect renders context selection
func (m *MockView) RenderContextSelect() string {
	var output strings.Builder

	// Status bar
	statusBar := m.renderStatusBar("[2 of 15]", "Implement user authentication system with OAuth2 support")
	output.WriteString(statusBar)
	output.WriteString("\n")

	// Context selection - calculate content width
	// The lipgloss Width() sets content width, total width = content + border(2) + padding(4)
	dialogWidth := m.Width - 6
	if dialogWidth < 24 {
		dialogWidth = 24
	}
	
	contextStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("6")).
		Padding(1, 2).
		Width(dialogWidth).
		Align(lipgloss.Center)

	// Adjust content for small terminals
	var content string
	if m.Height < 12 {
		content = `Available Contexts:

▶ work (current)
  home
  personal

↑↓: navigate  Enter: select  ESC: cancel`
	} else {
		content = `Available Contexts:

▶ work (current)
  home
  personal
  urgent
  someday

↑↓: navigate  Enter: select  ESC: cancel`
	}

	dialog := contextStyle.Render(content)
	
	// Force width constraint by truncating each line if necessary
	lines := strings.Split(dialog, "\n")
	for i, line := range lines {
		cleanLine := stripANSISequences(line)
		if len(cleanLine) > m.Width {
			lines[i] = line[:m.Width]
		}
	}
	dialog = strings.Join(lines, "\n")
	
	availableHeight := m.Height - 3
	output.WriteString(m.centerVertically(dialog, availableHeight))

	return output.String()
}

// RenderCalendar renders calendar view
func (m *MockView) RenderCalendar(mode string) string {
	var output strings.Builder

	// Status bar  
	statusBar := m.renderStatusBar("[2 of 15]", "Implement user authentication system with OAuth2 support")
	output.WriteString(statusBar)
	output.WriteString("\n")

	// Calendar - calculate content width
	// The lipgloss Width() sets content width, total width = content + border(2) + padding(4)
	calendarWidth := m.Width - 6
	if calendarWidth < 19 {
		calendarWidth = 19
	}
	
	calendarStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("6")).
		Padding(1, 2).
		Width(calendarWidth).
		Align(lipgloss.Center)

	// Get current month for display
	now := time.Now()
	monthYear := now.Format("January 2006")

	// Adjust content for small terminals
	var content string
	if m.Height < 16 {
		// Minimal calendar for small terminals
		content = fmt.Sprintf(`%s

Su Mo Tu We Th Fr Sa
 1  2  3  4  5  6  7
 8  9 10 11 12 13 14
15 16 17 18 19 20 21

Select %s date
ESC: cancel`, monthYear, mode)
	} else {
		// Full calendar for larger terminals
		content = fmt.Sprintf(`%s

Su Mo Tu We Th Fr Sa
          1  2  3  4
 5  6  7  8  9 10 11
12 13 14 15 16 17 18
19 20 21 22 23 24 25
26 27 28 29 30 31

Select %s date
Tab: text input, x: remove %s, ESC: cancel`, monthYear, mode, mode)
	}

	dialog := calendarStyle.Render(content)
	
	// Force width constraint by truncating each line if necessary
	lines := strings.Split(dialog, "\n")
	for i, line := range lines {
		cleanLine := stripANSISequences(line)
		if len(cleanLine) > m.Width {
			lines[i] = line[:m.Width]
		}
	}
	dialog = strings.Join(lines, "\n")
	
	availableHeight := m.Height - 3
	output.WriteString(m.centerVertically(dialog, availableHeight))

	return output.String()
}