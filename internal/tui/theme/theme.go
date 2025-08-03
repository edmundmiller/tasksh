package theme

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Theme represents a complete UI theme
type Theme struct {
	Name        string
	Colors      Colors
	Styles      Styles
	Components  Components
}

// Colors defines the color palette
type Colors struct {
	// Base colors
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Accent      lipgloss.Color
	
	// Semantic colors
	Success     lipgloss.Color
	Warning     lipgloss.Color
	Error       lipgloss.Color
	Info        lipgloss.Color
	
	// UI colors
	Background  lipgloss.Color
	Surface     lipgloss.Color
	Border      lipgloss.Color
	Text        lipgloss.Color
	TextMuted   lipgloss.Color
	TextInverse lipgloss.Color
	
	// Task-specific colors
	TaskPending    lipgloss.Color
	TaskCompleted  lipgloss.Color
	TaskWaiting    lipgloss.Color
	TaskDeleted    lipgloss.Color
	TaskRecurring  lipgloss.Color
	
	// Priority colors
	PriorityHigh   lipgloss.Color
	PriorityMedium lipgloss.Color
	PriorityLow    lipgloss.Color
}

// Styles defines common styling patterns
type Styles struct {
	// Layout
	Container   lipgloss.Style
	Header      lipgloss.Style
	Content     lipgloss.Style
	Footer      lipgloss.Style
	Sidebar     lipgloss.Style
	
	// Text
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Body        lipgloss.Style
	Caption     lipgloss.Style
	Code        lipgloss.Style
	
	// Interactive
	Button      lipgloss.Style
	ButtonFocus lipgloss.Style
	Input       lipgloss.Style
	InputFocus  lipgloss.Style
	
	// Status
	StatusBar   lipgloss.Style
	Badge       lipgloss.Style
	Tag         lipgloss.Style
	
	// Feedback
	Success     lipgloss.Style
	Warning     lipgloss.Style
	Error       lipgloss.Style
	Info        lipgloss.Style
}

// Components defines component-specific styles
type Components struct {
	TaskList    TaskListStyles
	Progress    ProgressStyles
	Calendar    CalendarStyles
	Help        HelpStyles
	Spinner     SpinnerStyles
}

type TaskListStyles struct {
	Item         lipgloss.Style
	ItemSelected lipgloss.Style
	ItemFocused  lipgloss.Style
	Description  lipgloss.Style
	Project      lipgloss.Style
	Tags         lipgloss.Style
	Due          lipgloss.Style
	Priority     lipgloss.Style
}

type ProgressStyles struct {
	Track    lipgloss.Style
	Fill     lipgloss.Style
	Text     lipgloss.Style
}

type CalendarStyles struct {
	Container    lipgloss.Style
	Header       lipgloss.Style
	Day          lipgloss.Style
	DaySelected  lipgloss.Style
	DayToday     lipgloss.Style
	DayOther     lipgloss.Style
}

type HelpStyles struct {
	Container    lipgloss.Style
	Key          lipgloss.Style
	Description  lipgloss.Style
	Separator    lipgloss.Style
}

type SpinnerStyles struct {
	Spinner lipgloss.Style
	Text    lipgloss.Style
}

var currentTheme *Theme

// GetTheme returns the current theme
func GetTheme() *Theme {
	if currentTheme == nil {
		currentTheme = detectTheme()
	}
	return currentTheme
}

// SetTheme sets the current theme
func SetTheme(theme *Theme) {
	currentTheme = theme
}

// detectTheme automatically detects the appropriate theme
func detectTheme() *Theme {
	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return NoColorTheme()
	}
	
	// Check explicit theme preference
	if themeName := os.Getenv("TASKSH_THEME"); themeName != "" {
		switch strings.ToLower(themeName) {
		case "light":
			return LightTheme()
		case "dark":
			return DarkTheme()
		case "auto":
			return autoDetectTheme()
		}
	}
	
	return autoDetectTheme()
}

// autoDetectTheme detects theme based on terminal capabilities
func autoDetectTheme() *Theme {
	profile := termenv.ColorProfile()
	
	switch profile {
	case termenv.TrueColor, termenv.ANSI256:
		// Modern terminal - check for dark mode indicators
		if isDarkMode() {
			return DarkTheme()
		}
		return LightTheme()
	default:
		// Limited color support
		return BasicTheme()
	}
}

// isDarkMode attempts to detect if terminal is in dark mode
func isDarkMode() bool {
	// Check common dark mode indicators
	indicators := []string{
		"DARK_MODE",
		"TERM_PROGRAM",
		"COLORTERM",
	}
	
	for _, env := range indicators {
		value := strings.ToLower(os.Getenv(env))
		if strings.Contains(value, "dark") || 
		   value == "iterm.app" || 
		   value == "vscode" ||
		   value == "truecolor" {
			return true
		}
	}
	
	// Default to dark for modern terminals
	return true
}

// DarkTheme returns a modern dark theme
func DarkTheme() *Theme {
	colors := Colors{
		Primary:        lipgloss.Color("#7C3AED"), // Purple
		Secondary:      lipgloss.Color("#06B6D4"), // Cyan
		Accent:         lipgloss.Color("#F59E0B"), // Amber
		
		Success:        lipgloss.Color("#10B981"), // Emerald
		Warning:        lipgloss.Color("#F59E0B"), // Amber
		Error:          lipgloss.Color("#EF4444"), // Red
		Info:           lipgloss.Color("#3B82F6"), // Blue
		
		Background:     lipgloss.Color("#0F172A"), // Slate 900
		Surface:        lipgloss.Color("#1E293B"), // Slate 800
		Border:         lipgloss.Color("#334155"), // Slate 700
		Text:           lipgloss.Color("#F1F5F9"), // Slate 100
		TextMuted:      lipgloss.Color("#94A3B8"), // Slate 400
		TextInverse:    lipgloss.Color("#0F172A"), // Slate 900
		
		TaskPending:    lipgloss.Color("#3B82F6"), // Blue
		TaskCompleted:  lipgloss.Color("#10B981"), // Emerald
		TaskWaiting:    lipgloss.Color("#F59E0B"), // Amber
		TaskDeleted:    lipgloss.Color("#6B7280"), // Gray
		TaskRecurring:  lipgloss.Color("#8B5CF6"), // Violet
		
		PriorityHigh:   lipgloss.Color("#EF4444"), // Red
		PriorityMedium: lipgloss.Color("#F59E0B"), // Amber
		PriorityLow:    lipgloss.Color("#6B7280"), // Gray
	}
	
	return buildTheme("dark", colors)
}

// LightTheme returns a modern light theme
func LightTheme() *Theme {
	colors := Colors{
		Primary:        lipgloss.Color("#7C3AED"), // Purple
		Secondary:      lipgloss.Color("#0891B2"), // Cyan
		Accent:         lipgloss.Color("#D97706"), // Amber
		
		Success:        lipgloss.Color("#059669"), // Emerald
		Warning:        lipgloss.Color("#D97706"), // Amber
		Error:          lipgloss.Color("#DC2626"), // Red
		Info:           lipgloss.Color("#2563EB"), // Blue
		
		Background:     lipgloss.Color("#FFFFFF"), // White
		Surface:        lipgloss.Color("#F8FAFC"), // Slate 50
		Border:         lipgloss.Color("#E2E8F0"), // Slate 200
		Text:           lipgloss.Color("#0F172A"), // Slate 900
		TextMuted:      lipgloss.Color("#64748B"), // Slate 500
		TextInverse:    lipgloss.Color("#FFFFFF"), // White
		
		TaskPending:    lipgloss.Color("#2563EB"), // Blue
		TaskCompleted:  lipgloss.Color("#059669"), // Emerald
		TaskWaiting:    lipgloss.Color("#D97706"), // Amber
		TaskDeleted:    lipgloss.Color("#6B7280"), // Gray
		TaskRecurring:  lipgloss.Color("#7C3AED"), // Violet
		
		PriorityHigh:   lipgloss.Color("#DC2626"), // Red
		PriorityMedium: lipgloss.Color("#D97706"), // Amber
		PriorityLow:    lipgloss.Color("#6B7280"), // Gray
	}
	
	return buildTheme("light", colors)
}

// BasicTheme returns a theme for terminals with limited color support
func BasicTheme() *Theme {
	colors := Colors{
		Primary:        lipgloss.Color("12"), // Bright Blue
		Secondary:      lipgloss.Color("14"), // Bright Cyan
		Accent:         lipgloss.Color("11"), // Bright Yellow
		
		Success:        lipgloss.Color("10"), // Bright Green
		Warning:        lipgloss.Color("11"), // Bright Yellow
		Error:          lipgloss.Color("9"),  // Bright Red
		Info:           lipgloss.Color("12"), // Bright Blue
		
		Background:     lipgloss.Color("0"),  // Black
		Surface:        lipgloss.Color("8"),  // Bright Black
		Border:         lipgloss.Color("7"),  // White
		Text:           lipgloss.Color("15"), // Bright White
		TextMuted:      lipgloss.Color("7"),  // White
		TextInverse:    lipgloss.Color("0"),  // Black
		
		TaskPending:    lipgloss.Color("12"), // Bright Blue
		TaskCompleted:  lipgloss.Color("10"), // Bright Green
		TaskWaiting:    lipgloss.Color("11"), // Bright Yellow
		TaskDeleted:    lipgloss.Color("8"),  // Bright Black
		TaskRecurring:  lipgloss.Color("13"), // Bright Magenta
		
		PriorityHigh:   lipgloss.Color("9"),  // Bright Red
		PriorityMedium: lipgloss.Color("11"), // Bright Yellow
		PriorityLow:    lipgloss.Color("8"),  // Bright Black
	}
	
	return buildTheme("basic", colors)
}

// NoColorTheme returns a theme with no colors
func NoColorTheme() *Theme {
	noColor := lipgloss.Color("")
	
	colors := Colors{
		Primary:        noColor,
		Secondary:      noColor,
		Accent:         noColor,
		Success:        noColor,
		Warning:        noColor,
		Error:          noColor,
		Info:           noColor,
		Background:     noColor,
		Surface:        noColor,
		Border:         noColor,
		Text:           noColor,
		TextMuted:      noColor,
		TextInverse:    noColor,
		TaskPending:    noColor,
		TaskCompleted:  noColor,
		TaskWaiting:    noColor,
		TaskDeleted:    noColor,
		TaskRecurring:  noColor,
		PriorityHigh:   noColor,
		PriorityMedium: noColor,
		PriorityLow:    noColor,
	}
	
	return buildTheme("no-color", colors)
}

// buildTheme constructs a complete theme from colors
func buildTheme(name string, colors Colors) *Theme {
	styles := buildStyles(colors)
	components := buildComponents(colors, styles)
	
	return &Theme{
		Name:       name,
		Colors:     colors,
		Styles:     styles,
		Components: components,
	}
}

// buildStyles creates styles from colors
func buildStyles(c Colors) Styles {
	return Styles{
		// Layout
		Container: lipgloss.NewStyle().
			Background(c.Background).
			Foreground(c.Text),
			
		Header: lipgloss.NewStyle().
			Background(c.Surface).
			Foreground(c.Text).
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(c.Border),
			
		Content: lipgloss.NewStyle().
			Padding(1, 2).
			Background(c.Background).
			Foreground(c.Text),
			
		Footer: lipgloss.NewStyle().
			Background(c.Surface).
			Foreground(c.TextMuted).
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(c.Border),
			
		Sidebar: lipgloss.NewStyle().
			Background(c.Surface).
			Foreground(c.Text).
			Padding(1).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(c.Border),
			
		// Text
		Title: lipgloss.NewStyle().
			Foreground(c.Primary).
			Bold(true).
			MarginBottom(1),
			
		Subtitle: lipgloss.NewStyle().
			Foreground(c.Secondary).
			Bold(true),
			
		Body: lipgloss.NewStyle().
			Foreground(c.Text),
			
		Caption: lipgloss.NewStyle().
			Foreground(c.TextMuted).
			Italic(true),
			
		Code: lipgloss.NewStyle().
			Foreground(c.Accent).
			Background(c.Surface).
			Padding(0, 1),
			
		// Interactive
		Button: lipgloss.NewStyle().
			Background(c.Primary).
			Foreground(c.TextInverse).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c.Primary),
			
		ButtonFocus: lipgloss.NewStyle().
			Background(c.Secondary).
			Foreground(c.TextInverse).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c.Secondary),
			
		Input: lipgloss.NewStyle().
			Background(c.Surface).
			Foreground(c.Text).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c.Border),
			
		InputFocus: lipgloss.NewStyle().
			Background(c.Surface).
			Foreground(c.Text).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c.Primary),
			
		// Status
		StatusBar: lipgloss.NewStyle().
			Background(c.Primary).
			Foreground(c.TextInverse).
			Padding(0, 1),
			
		Badge: lipgloss.NewStyle().
			Background(c.Accent).
			Foreground(c.TextInverse).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()),
			
		Tag: lipgloss.NewStyle().
			Background(c.Surface).
			Foreground(c.Text).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c.Border),
			
		// Feedback
		Success: lipgloss.NewStyle().
			Foreground(c.Success).
			Bold(true),
			
		Warning: lipgloss.NewStyle().
			Foreground(c.Warning).
			Bold(true),
			
		Error: lipgloss.NewStyle().
			Foreground(c.Error).
			Bold(true),
			
		Info: lipgloss.NewStyle().
			Foreground(c.Info).
			Bold(true),
	}
}

// buildComponents creates component-specific styles
func buildComponents(c Colors, s Styles) Components {
	return Components{
		TaskList: TaskListStyles{
			Item: lipgloss.NewStyle().
				Padding(0, 1).
				MarginBottom(0),
				
			ItemSelected: lipgloss.NewStyle().
				Background(c.Primary).
				Foreground(c.TextInverse).
				Padding(0, 1).
				MarginBottom(0),
				
			ItemFocused: lipgloss.NewStyle().
				Background(c.Surface).
				Foreground(c.Text).
				Padding(0, 1).
				MarginBottom(0).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(c.Primary),
				
			Description: lipgloss.NewStyle().
				Foreground(c.Text),
				
			Project: lipgloss.NewStyle().
				Foreground(c.Secondary).
				Bold(true),
				
			Tags: lipgloss.NewStyle().
				Foreground(c.Accent),
				
			Due: lipgloss.NewStyle().
				Foreground(c.Warning),
				
			Priority: lipgloss.NewStyle().
				Bold(true),
		},
		
		Progress: ProgressStyles{
			Track: lipgloss.NewStyle().
				Background(c.Surface).
				Foreground(c.Border),
				
			Fill: lipgloss.NewStyle().
				Background(c.Primary).
				Foreground(c.TextInverse),
				
			Text: lipgloss.NewStyle().
				Foreground(c.Text),
		},
		
		Calendar: CalendarStyles{
			Container: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(c.Border).
				Padding(1),
				
			Header: lipgloss.NewStyle().
				Foreground(c.Primary).
				Bold(true).
				Align(lipgloss.Center),
				
			Day: lipgloss.NewStyle().
				Foreground(c.Text).
				Padding(0, 1),
				
			DaySelected: lipgloss.NewStyle().
				Background(c.Primary).
				Foreground(c.TextInverse).
				Padding(0, 1),
				
			DayToday: lipgloss.NewStyle().
				Background(c.Accent).
				Foreground(c.TextInverse).
				Padding(0, 1).
				Bold(true),
				
			DayOther: lipgloss.NewStyle().
				Foreground(c.TextMuted).
				Padding(0, 1),
		},
		
		Help: HelpStyles{
			Container: lipgloss.NewStyle().
				Background(c.Surface).
				Foreground(c.Text).
				Padding(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(c.Border),
				
			Key: lipgloss.NewStyle().
				Foreground(c.Primary).
				Bold(true),
				
			Description: lipgloss.NewStyle().
				Foreground(c.Text),
				
			Separator: lipgloss.NewStyle().
				Foreground(c.TextMuted),
		},
		
		Spinner: SpinnerStyles{
			Spinner: lipgloss.NewStyle().
				Foreground(c.Primary),
				
			Text: lipgloss.NewStyle().
				Foreground(c.Text).
				MarginLeft(1),
		},
	}
}