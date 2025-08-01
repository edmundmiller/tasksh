package review

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Toast represents a temporary notification
type Toast struct {
	Type      ToastType
	Message   string
	Duration  time.Duration
	StartTime time.Time
}

// ToastType defines the type of toast
type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// FeedbackManager handles visual feedback
type FeedbackManager struct {
	toasts        []Toast
	maxToasts     int
	defaultDuration time.Duration
}

// NewFeedbackManager creates a new feedback manager
func NewFeedbackManager() *FeedbackManager {
	return &FeedbackManager{
		toasts:          []Toast{},
		maxToasts:       3,
		defaultDuration: 3 * time.Second,
	}
}

// ShowToast displays a toast notification
func (f *FeedbackManager) ShowToast(toastType ToastType, message string) tea.Cmd {
	toast := Toast{
		Type:      toastType,
		Message:   message,
		Duration:  f.defaultDuration,
		StartTime: time.Now(),
	}
	
	f.toasts = append(f.toasts, toast)
	
	// Keep only the most recent toasts
	if len(f.toasts) > f.maxToasts {
		f.toasts = f.toasts[len(f.toasts)-f.maxToasts:]
	}
	
	// Return command to remove toast after duration
	return tea.Tick(toast.Duration, func(t time.Time) tea.Msg {
		return toastExpiredMsg{startTime: toast.StartTime}
	})
}

// RemoveExpiredToast removes a toast that has expired
func (f *FeedbackManager) RemoveExpiredToast(startTime time.Time) {
	for i, toast := range f.toasts {
		if toast.StartTime == startTime {
			f.toasts = append(f.toasts[:i], f.toasts[i+1:]...)
			break
		}
	}
}

// RenderToasts renders all active toasts
func (f *FeedbackManager) RenderToasts(width int) string {
	if len(f.toasts) == 0 {
		return ""
	}
	
	var rendered []string
	for _, toast := range f.toasts {
		rendered = append(rendered, f.renderToast(toast, width))
	}
	
	return strings.Join(rendered, "\n")
}

// renderToast renders a single toast
func (f *FeedbackManager) renderToast(toast Toast, width int) string {
	// Choose style based on type
	var style lipgloss.Style
	var icon string
	
	switch toast.Type {
	case ToastSuccess:
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("82")).
			Foreground(lipgloss.Color("231")).
			Padding(0, 2)
		icon = "✓"
		
	case ToastWarning:
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("214")).
			Foreground(lipgloss.Color("231")).
			Padding(0, 2)
		icon = "⚠"
		
	case ToastError:
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("231")).
			Padding(0, 2)
		icon = "✗"
		
	default:
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("39")).
			Foreground(lipgloss.Color("231")).
			Padding(0, 2)
		icon = "ℹ"
	}
	
	content := fmt.Sprintf("%s %s", icon, toast.Message)
	return style.Render(content)
}

// Confirmation represents a confirmation dialog
type Confirmation struct {
	Title   string
	Message string
	Options []string
}

// RenderConfirmation renders a confirmation dialog
func RenderConfirmation(c Confirmation, selectedIndex int, width, height int) string {
	// Dialog style
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(1, 3)
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214"))
	
	// Build content
	var content strings.Builder
	content.WriteString(titleStyle.Render(c.Title))
	content.WriteString("\n\n")
	content.WriteString(c.Message)
	content.WriteString("\n\n")
	
	// Options
	for i, option := range c.Options {
		if i == selectedIndex {
			selected := lipgloss.NewStyle().
				Background(lipgloss.Color("39")).
				Foreground(lipgloss.Color("231")).
				Padding(0, 2).
				Render(option)
			content.WriteString(selected)
		} else {
			normal := lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Padding(0, 2).
				Render(option)
			content.WriteString(normal)
		}
		
		if i < len(c.Options)-1 {
			content.WriteString("  ")
		}
	}
	
	dialog := dialogStyle.Render(content.String())
	
	// Center on screen
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}

// AnimatedSpinner provides different spinner styles
type AnimatedSpinner struct {
	frames []string
	index  int
}

// NewAnimatedSpinner creates a new animated spinner
func NewAnimatedSpinner(style string) *AnimatedSpinner {
	var frames []string
	
	switch style {
	case "dots":
		frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	case "line":
		frames = []string{"-", "\\", "|", "/"}
	case "circle":
		frames = []string{"◐", "◓", "◑", "◒"}
	default:
		frames = []string{".", "..", "...", "....", ".....", "......"}
	}
	
	return &AnimatedSpinner{
		frames: frames,
		index:  0,
	}
}

// Next advances the spinner and returns the current frame
func (s *AnimatedSpinner) Next() string {
	frame := s.frames[s.index]
	s.index = (s.index + 1) % len(s.frames)
	return frame
}

// ProgressIndicator shows progress with visual style
type ProgressIndicator struct {
	current int
	total   int
	width   int
}

// NewProgressIndicator creates a progress indicator
func NewProgressIndicator(current, total, width int) *ProgressIndicator {
	return &ProgressIndicator{
		current: current,
		total:   total,
		width:   width,
	}
}

// Render renders the progress indicator
func (p *ProgressIndicator) Render() string {
	if p.total == 0 {
		return ""
	}
	
	percentage := float64(p.current) / float64(p.total)
	filled := int(percentage * float64(p.width))
	
	// Build bar
	var bar strings.Builder
	for i := 0; i < p.width; i++ {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}
	
	// Style based on progress
	var style lipgloss.Style
	if percentage < 0.25 {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	} else if percentage < 0.75 {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	} else {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	}
	
	return style.Render(fmt.Sprintf("[%s] %d%%", bar.String(), int(percentage*100)))
}

// Messages

type toastExpiredMsg struct {
	startTime time.Time
}

// Helper functions for quick feedback

// ShowSuccess shows a success toast
func ShowSuccess(message string) tea.Cmd {
	return func() tea.Msg {
		return ShowToastMsg{
			Type:    ToastSuccess,
			Message: message,
		}
	}
}

// ShowError shows an error toast
func ShowError(message string) tea.Cmd {
	return func() tea.Msg {
		return ShowToastMsg{
			Type:    ToastError,
			Message: message,
		}
	}
}

// ShowInfo shows an info toast
func ShowInfo(message string) tea.Cmd {
	return func() tea.Msg {
		return ShowToastMsg{
			Type:    ToastInfo,
			Message: message,
		}
	}
}

// ShowToastMsg is a message to show a toast
type ShowToastMsg struct {
	Type    ToastType
	Message string
}