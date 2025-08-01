package review

import (
	"os"
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color theme
type Theme struct {
	Name   string
	Styles *Styles
}

// GetTheme returns the appropriate theme based on environment
func GetTheme() *Theme {
	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return NoColorTheme()
	}
	
	// Check for explicit theme preference
	themeName := os.Getenv("TASKSH_THEME")
	switch themeName {
	case "light":
		return LightTheme()
	case "dark":
		return DarkTheme()
	case "auto":
		return AutoTheme()
	default:
		// Default to auto-detection
		return AutoTheme()
	}
}

// AutoTheme attempts to detect the appropriate theme
func AutoTheme() *Theme {
	// Check common environment variables that indicate dark mode
	colorterm := os.Getenv("COLORTERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	
	// Check for dark mode indicators
	if colorterm == "truecolor" || colorterm == "24bit" {
		// Modern terminal with good color support - use dark theme
		return DarkTheme()
	}
	
	// Check for specific terminal programs
	switch termProgram {
	case "iTerm.app", "vscode", "Hyper":
		return DarkTheme()
	}
	
	// Default to a neutral theme
	return DefaultTheme()
}

// DefaultTheme returns the default (neutral) theme
func DefaultTheme() *Theme {
	return &Theme{
		Name:   "default",
		Styles: DefaultStyles(),
	}
}

// DarkTheme returns a dark theme
func DarkTheme() *Theme {
	styles := &Styles{
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
			Foreground(lipgloss.Color("147")), // Purple
			
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Bold(true),
			
		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
			
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")),
			
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")), // Green
			
		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")), // Orange
			
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")), // Red
			
		Input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")). // Dark blue
			Padding(1, 2),
			
		Button: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("231")).
			Padding(0, 2),
			
		Selected: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("231")),
	}
	
	return &Theme{
		Name:   "dark",
		Styles: styles,
	}
}

// LightTheme returns a light theme
func LightTheme() *Theme {
	styles := &Styles{
		Container: lipgloss.NewStyle().
			Padding(0).
			Margin(0),
			
		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("254")).
			Foreground(lipgloss.Color("235")).
			Padding(0, 1),
			
		Content: lipgloss.NewStyle().
			Padding(1, 2),
			
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Padding(0, 1),
			
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("127")), // Deep purple
			
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Bold(true),
			
		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")),
			
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("247")),
			
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("28")), // Dark green
			
		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("166")), // Dark orange
			
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("124")), // Dark red
			
		Input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("33")). // Blue
			Padding(1, 2),
			
		Button: lipgloss.NewStyle().
			Background(lipgloss.Color("33")).
			Foreground(lipgloss.Color("231")).
			Padding(0, 2),
			
		Selected: lipgloss.NewStyle().
			Background(lipgloss.Color("33")).
			Foreground(lipgloss.Color("231")),
	}
	
	return &Theme{
		Name:   "light",
		Styles: styles,
	}
}

// NoColorTheme returns a theme with no colors
func NoColorTheme() *Theme {
	styles := &Styles{
		Container: lipgloss.NewStyle(),
		StatusBar: lipgloss.NewStyle().
			Reverse(true).
			Padding(0, 1),
		Content: lipgloss.NewStyle().
			Padding(1, 2),
		Footer: lipgloss.NewStyle().
			Padding(0, 1),
		Title: lipgloss.NewStyle().
			Bold(true),
		Label: lipgloss.NewStyle().
			Bold(true),
		Value: lipgloss.NewStyle(),
		Muted: lipgloss.NewStyle().
			Faint(true),
		Success: lipgloss.NewStyle(),
		Warning: lipgloss.NewStyle(),
		Error: lipgloss.NewStyle().
			Bold(true),
		Input: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(1, 2),
		Button: lipgloss.NewStyle().
			Reverse(true).
			Padding(0, 2),
		Selected: lipgloss.NewStyle().
			Reverse(true),
	}
	
	return &Theme{
		Name:   "no-color",
		Styles: styles,
	}
}

// ApplyTheme updates the renderer with a new theme
func (r *ViewRenderer) ApplyTheme(theme *Theme) {
	r.styles = theme.Styles
}

// GetThemeInfo returns information about the current theme
func GetThemeInfo() string {
	theme := GetTheme()
	info := "Theme: " + theme.Name
	
	if os.Getenv("NO_COLOR") != "" {
		info += " (NO_COLOR set)"
	} else if os.Getenv("TASKSH_THEME") != "" {
		info += " (TASKSH_THEME=" + os.Getenv("TASKSH_THEME") + ")"
	} else {
		info += " (auto-detected)"
	}
	
	return info
}