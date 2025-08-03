package components

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/emiller/tasksh/internal/tui/theme"
)

// ProgressBar is a modern progress bar component with animations
type ProgressBar struct {
	*BaseComponent
	
	// Progress state
	current     float64
	total       float64
	percentage  float64
	
	// Display
	label       string
	showPercent bool
	showNumbers bool
	
	// Animation
	animated       bool
	animationSpeed float64
	targetPercent  float64
	
	// Styling
	theme          *theme.Theme
	styles         theme.ProgressStyles
	
	// Customization
	width          int
	fillChar       string
	emptyChar      string
	gradient       bool
}

// ProgressTickMsg is sent for animation updates
type ProgressTickMsg struct {
	Time time.Time
}

// NewProgressBar creates a new progress bar
func NewProgressBar() *ProgressBar {
	base := NewBaseComponent()
	t := theme.GetTheme()
	
	return &ProgressBar{
		BaseComponent:  base,
		current:        0,
		total:          100,
		percentage:     0,
		label:          "",
		showPercent:    true,
		showNumbers:    false,
		animated:       true,
		animationSpeed: 0.1,
		targetPercent:  0,
		theme:          t,
		styles:         t.Components.Progress,
		width:          40,
		fillChar:       "█",
		emptyChar:      "░",
		gradient:       true,
	}
}

// SetProgress updates the progress value
func (pb *ProgressBar) SetProgress(current, total float64) {
	pb.current = current
	pb.total = total
	
	if total > 0 {
		pb.targetPercent = (current / total) * 100
	} else {
		pb.targetPercent = 0
	}
	
	if !pb.animated {
		pb.percentage = pb.targetPercent
	}
}

// SetLabel sets the progress label
func (pb *ProgressBar) SetLabel(label string) {
	pb.label = label
}

// SetShowPercent controls percentage display
func (pb *ProgressBar) SetShowPercent(show bool) {
	pb.showPercent = show
}

// SetShowNumbers controls number display
func (pb *ProgressBar) SetShowNumbers(show bool) {
	pb.showNumbers = show
}

// SetAnimated controls animation
func (pb *ProgressBar) SetAnimated(animated bool) {
	pb.animated = animated
	if !animated {
		pb.percentage = pb.targetPercent
	}
}

// SetAnimationSpeed sets animation speed (0.0 to 1.0)
func (pb *ProgressBar) SetAnimationSpeed(speed float64) {
	if speed < 0 {
		speed = 0
	}
	if speed > 1 {
		speed = 1
	}
	pb.animationSpeed = speed
}

// SetWidth sets the progress bar width
func (pb *ProgressBar) SetWidth(width int) {
	if width < 10 {
		width = 10
	}
	pb.width = width
}

// SetFillChar sets the fill character
func (pb *ProgressBar) SetFillChar(char string) {
	if char != "" {
		pb.fillChar = char
	}
}

// SetEmptyChar sets the empty character
func (pb *ProgressBar) SetEmptyChar(char string) {
	if char != "" {
		pb.emptyChar = char
	}
}

// SetGradient controls gradient effect
func (pb *ProgressBar) SetGradient(gradient bool) {
	pb.gradient = gradient
}

// Progress returns current progress values
func (pb *ProgressBar) Progress() (current, total, percentage float64) {
	return pb.current, pb.total, pb.percentage
}

// IsComplete returns true if progress is at 100%
func (pb *ProgressBar) IsComplete() bool {
	return pb.percentage >= 100
}

// Init implements tea.Model
func (pb *ProgressBar) Init() tea.Cmd {
	if pb.animated {
		return pb.tickCmd()
	}
	return pb.BaseComponent.Init()
}

// Update implements tea.Model
func (pb *ProgressBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Handle base component updates
	_, cmd := pb.BaseComponent.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	switch msg := msg.(type) {
	case ProgressTickMsg:
		if pb.animated {
			// Smooth animation towards target
			diff := pb.targetPercent - pb.percentage
			if math.Abs(diff) > 0.1 {
				pb.percentage += diff * pb.animationSpeed
				cmds = append(cmds, pb.tickCmd())
			} else {
				pb.percentage = pb.targetPercent
			}
		}
		
	case ThemeChangeMsg:
		if t, ok := msg.Theme.(*theme.Theme); ok {
			pb.theme = t
			pb.styles = t.Components.Progress
		}
	}
	
	return pb, tea.Batch(cmds...)
}

// tickCmd returns a command for animation ticks
func (pb *ProgressBar) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return ProgressTickMsg{Time: t}
	})
}

// View implements tea.Model
func (pb *ProgressBar) View() string {
	if pb.width == 0 {
		return ""
	}
	
	var parts []string
	
	// Label
	if pb.label != "" {
		labelStyle := pb.styles.Text
		if pb.Focused() {
			labelStyle = labelStyle.Foreground(pb.theme.Colors.Primary)
		}
		parts = append(parts, labelStyle.Render(pb.label))
	}
	
	// Progress bar
	bar := pb.renderBar()
	parts = append(parts, bar)
	
	// Progress text
	if pb.showPercent || pb.showNumbers {
		text := pb.renderProgressText()
		parts = append(parts, pb.styles.Text.Render(text))
	}
	
	return strings.Join(parts, " ")
}

// renderBar renders the progress bar
func (pb *ProgressBar) renderBar() string {
	// Calculate fill width
	fillWidth := int((pb.percentage / 100) * float64(pb.width))
	if fillWidth > pb.width {
		fillWidth = pb.width
	}
	
	var bar strings.Builder
	
	// Build the bar
	for i := 0; i < pb.width; i++ {
		if i < fillWidth {
			if pb.gradient {
				// Gradient effect
				intensity := float64(i) / float64(pb.width)
				color := pb.getGradientColor(intensity)
				style := lipgloss.NewStyle().Foreground(color)
				bar.WriteString(style.Render(pb.fillChar))
			} else {
				bar.WriteString(pb.fillChar)
			}
		} else {
			bar.WriteString(pb.emptyChar)
		}
	}
	
	// Apply styling
	barStyle := pb.styles.Track
	if pb.Focused() {
		barStyle = barStyle.Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(pb.theme.Colors.Primary)
	}
	
	return barStyle.Render(bar.String())
}

// renderProgressText renders the progress text
func (pb *ProgressBar) renderProgressText() string {
	var parts []string
	
	if pb.showNumbers {
		numbers := fmt.Sprintf("%.0f/%.0f", pb.current, pb.total)
		parts = append(parts, numbers)
	}
	
	if pb.showPercent {
		percent := fmt.Sprintf("%.1f%%", pb.percentage)
		parts = append(parts, percent)
	}
	
	return strings.Join(parts, " ")
}

// getGradientColor returns a color for gradient effect
func (pb *ProgressBar) getGradientColor(intensity float64) lipgloss.Color {
	// Simple gradient from secondary to primary
	if intensity < 0.5 {
		return pb.theme.Colors.Secondary
	}
	return pb.theme.Colors.Primary
}

// Spinner is a modern spinner component
type Spinner struct {
	*BaseComponent
	
	// State
	active   bool
	frame    int
	
	// Display
	label    string
	frames   []string
	
	// Styling
	theme    *theme.Theme
	styles   theme.SpinnerStyles
	
	// Animation
	speed    time.Duration
}

// SpinnerTickMsg is sent for spinner animation
type SpinnerTickMsg struct {
	Time time.Time
}

// NewSpinner creates a new spinner
func NewSpinner() *Spinner {
	base := NewBaseComponent()
	t := theme.GetTheme()
	
	return &Spinner{
		BaseComponent: base,
		active:        false,
		frame:         0,
		label:         "",
		frames:        []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		theme:         t,
		styles:        t.Components.Spinner,
		speed:         time.Millisecond * 100,
	}
}

// Start starts the spinner
func (s *Spinner) Start() tea.Cmd {
	s.active = true
	return s.tickCmd()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	s.active = false
}

// SetLabel sets the spinner label
func (s *Spinner) SetLabel(label string) {
	s.label = label
}

// SetFrames sets custom spinner frames
func (s *Spinner) SetFrames(frames []string) {
	if len(frames) > 0 {
		s.frames = frames
		s.frame = 0
	}
}

// SetSpeed sets the animation speed
func (s *Spinner) SetSpeed(speed time.Duration) {
	if speed > 0 {
		s.speed = speed
	}
}

// IsActive returns true if spinner is active
func (s *Spinner) IsActive() bool {
	return s.active
}

// Init implements tea.Model
func (s *Spinner) Init() tea.Cmd {
	if s.active {
		return s.tickCmd()
	}
	return s.BaseComponent.Init()
}

// Update implements tea.Model
func (s *Spinner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Handle base component updates
	_, cmd := s.BaseComponent.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	switch msg := msg.(type) {
	case SpinnerTickMsg:
		if s.active {
			s.frame = (s.frame + 1) % len(s.frames)
			cmds = append(cmds, s.tickCmd())
		}
		
	case ThemeChangeMsg:
		if t, ok := msg.Theme.(*theme.Theme); ok {
			s.theme = t
			s.styles = t.Components.Spinner
		}
	}
	
	return s, tea.Batch(cmds...)
}

// tickCmd returns a command for animation ticks
func (s *Spinner) tickCmd() tea.Cmd {
	return tea.Tick(s.speed, func(t time.Time) tea.Msg {
		return SpinnerTickMsg{Time: t}
	})
}

// View implements tea.Model
func (s *Spinner) View() string {
	if !s.active {
		return ""
	}
	
	var parts []string
	
	// Spinner
	spinnerChar := s.frames[s.frame]
	spinner := s.styles.Spinner.Render(spinnerChar)
	parts = append(parts, spinner)
	
	// Label
	if s.label != "" {
		label := s.styles.Text.Render(s.label)
		parts = append(parts, label)
	}
	
	return strings.Join(parts, " ")
}

// LoadingIndicator combines spinner and progress for complex operations
type LoadingIndicator struct {
	*BaseComponent
	
	// Components
	spinner     *Spinner
	progressBar *ProgressBar
	
	// State
	mode        LoadingMode
	message     string
	
	// Styling
	theme       *theme.Theme
}

// LoadingMode defines the loading indicator mode
type LoadingMode int

const (
	LoadingModeSpinner LoadingMode = iota
	LoadingModeProgress
	LoadingModeBoth
)

// NewLoadingIndicator creates a new loading indicator
func NewLoadingIndicator() *LoadingIndicator {
	base := NewBaseComponent()
	t := theme.GetTheme()
	
	return &LoadingIndicator{
		BaseComponent: base,
		spinner:       NewSpinner(),
		progressBar:   NewProgressBar(),
		mode:          LoadingModeSpinner,
		theme:         t,
	}
}

// SetMode sets the loading mode
func (li *LoadingIndicator) SetMode(mode LoadingMode) {
	li.mode = mode
}

// SetMessage sets the loading message
func (li *LoadingIndicator) SetMessage(message string) {
	li.message = message
	li.spinner.SetLabel(message)
	li.progressBar.SetLabel(message)
}

// Start starts the loading indicator
func (li *LoadingIndicator) Start() tea.Cmd {
	var cmds []tea.Cmd
	
	if li.mode == LoadingModeSpinner || li.mode == LoadingModeBoth {
		cmds = append(cmds, li.spinner.Start())
	}
	
	if li.mode == LoadingModeProgress || li.mode == LoadingModeBoth {
		cmds = append(cmds, li.progressBar.Init())
	}
	
	return tea.Batch(cmds...)
}

// Stop stops the loading indicator
func (li *LoadingIndicator) Stop() {
	li.spinner.Stop()
}

// SetProgress updates progress (for progress mode)
func (li *LoadingIndicator) SetProgress(current, total float64) {
	li.progressBar.SetProgress(current, total)
}

// Init implements tea.Model
func (li *LoadingIndicator) Init() tea.Cmd {
	return li.Start()
}

// Update implements tea.Model
func (li *LoadingIndicator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Handle base component updates
	_, cmd := li.BaseComponent.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	// Update components
	if li.mode == LoadingModeSpinner || li.mode == LoadingModeBoth {
		_, cmd := li.spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	
	if li.mode == LoadingModeProgress || li.mode == LoadingModeBoth {
		_, cmd := li.progressBar.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	
	return li, tea.Batch(cmds...)
}

// View implements tea.Model
func (li *LoadingIndicator) View() string {
	switch li.mode {
	case LoadingModeSpinner:
		return li.spinner.View()
		
	case LoadingModeProgress:
		return li.progressBar.View()
		
	case LoadingModeBoth:
		spinner := li.spinner.View()
		progress := li.progressBar.View()
		return lipgloss.JoinVertical(lipgloss.Left, spinner, progress)
		
	default:
		return ""
	}
}

// Implement interfaces
var _ Component = (*ProgressBar)(nil)
var _ Animatable = (*ProgressBar)(nil)
var _ Component = (*Spinner)(nil)
var _ Animatable = (*Spinner)(nil)
var _ Component = (*LoadingIndicator)(nil)

// Animatable interface implementation for ProgressBar
func (pb *ProgressBar) StartAnimation(name string) tea.Cmd {
	pb.SetAnimated(true)
	return pb.tickCmd()
}

func (pb *ProgressBar) StopAnimation(name string) tea.Cmd {
	pb.SetAnimated(false)
	return nil
}

func (pb *ProgressBar) IsAnimating() bool {
	return pb.animated && math.Abs(pb.targetPercent-pb.percentage) > 0.1
}

func (pb *ProgressBar) AnimationSpeed() float64 {
	return pb.animationSpeed
}

// Animatable interface implementation for Spinner
func (s *Spinner) StartAnimation(name string) tea.Cmd {
	return s.Start()
}

func (s *Spinner) StopAnimation(name string) tea.Cmd {
	s.Stop()
	return nil
}

func (s *Spinner) IsAnimating() bool {
	return s.active
}

func (s *Spinner) SetAnimationSpeed(speed float64) {
	// Convert speed (0.0-1.0) to duration
	duration := time.Duration((1.0 - speed) * 200 * float64(time.Millisecond))
	s.SetSpeed(duration)
}

func (s *Spinner) AnimationSpeed() float64 {
	// Convert duration back to speed
	return 1.0 - (float64(s.speed) / float64(200*time.Millisecond))
}