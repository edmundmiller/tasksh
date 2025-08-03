package components

import (
	"context"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Component represents a reusable UI component
type Component interface {
	tea.Model
	
	// SetSize updates the component's dimensions
	SetSize(width, height int)
	
	// Focus sets the component's focus state
	Focus() tea.Cmd
	Blur()
	Focused() bool
	
	// Theme support
	SetTheme(theme interface{})
}

// ViewModel extends Component with view-specific methods
type ViewModel interface {
	Component
	
	// View rendering with context
	ViewWithContext(ctx context.Context) string
	
	// Layout information
	MinWidth() int
	MinHeight() int
	PreferredSize() (width, height int)
}

// Container represents a component that can contain other components
type Container interface {
	Component
	
	// Child management
	AddChild(child Component) error
	RemoveChild(child Component) error
	Children() []Component
	
	// Layout
	SetLayout(layout LayoutManager)
	Layout() LayoutManager
}

// Layout defines how components are arranged
type Layout interface {
	// Arrange positions and sizes child components
	Arrange(container Container, width, height int) []ComponentBounds
}

// ComponentBounds defines a component's position and size
type ComponentBounds struct {
	Component Component
	X, Y      int
	Width     int
	Height    int
}

// Focusable represents components that can receive focus
type Focusable interface {
	Component
	
	// Focus management
	CanFocus() bool
	Focus() tea.Cmd
	Blur()
	Focused() bool
	
	// Navigation
	NextFocusable() Focusable
	PrevFocusable() Focusable
}

// Selectable represents components with selectable items
type Selectable interface {
	Component
	
	// Selection
	Select(index int) tea.Cmd
	Selected() int
	SelectNext() tea.Cmd
	SelectPrev() tea.Cmd
	
	// Items
	ItemCount() int
	Item(index int) interface{}
}

// Scrollable represents components that can scroll
type Scrollable interface {
	Component
	
	// Scrolling
	ScrollUp(lines int) tea.Cmd
	ScrollDown(lines int) tea.Cmd
	ScrollToTop() tea.Cmd
	ScrollToBottom() tea.Cmd
	
	// Position
	ScrollPosition() (offset, total int)
	CanScrollUp() bool
	CanScrollDown() bool
}

// Filterable represents components that support filtering
type Filterable interface {
	Component
	
	// Filtering
	SetFilter(filter string) tea.Cmd
	Filter() string
	ClearFilter() tea.Cmd
	
	// Results
	FilteredCount() int
	TotalCount() int
}

// Animatable represents components that support animations
type Animatable interface {
	Component
	
	// Animation control
	StartAnimation(name string) tea.Cmd
	StopAnimation(name string) tea.Cmd
	IsAnimating() bool
	
	// Animation properties
	SetAnimationSpeed(speed float64)
	AnimationSpeed() float64
}

// BaseComponent provides common functionality for all components
type BaseComponent struct {
	width   int
	height  int
	focused bool
	theme   interface{}
	
	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// NewBaseComponent creates a new base component
func NewBaseComponent() *BaseComponent {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseComponent{
		ctx:    ctx,
		cancel: cancel,
	}
}

// SetSize implements Component
func (c *BaseComponent) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// Size returns the component's dimensions
func (c *BaseComponent) Size() (width, height int) {
	return c.width, c.height
}

// Focus implements Component
func (c *BaseComponent) Focus() tea.Cmd {
	c.focused = true
	return nil
}

// Blur implements Component
func (c *BaseComponent) Blur() {
	c.focused = false
}

// Focused implements Component
func (c *BaseComponent) Focused() bool {
	return c.focused
}

// SetTheme implements Component
func (c *BaseComponent) SetTheme(theme interface{}) {
	c.theme = theme
}

// Theme returns the current theme
func (c *BaseComponent) Theme() interface{} {
	return c.theme
}

// Context returns the component's context
func (c *BaseComponent) Context() context.Context {
	return c.ctx
}

// Close cancels the component's context
func (c *BaseComponent) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

// Init implements tea.Model
func (c *BaseComponent) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (c *BaseComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			c.Close()
			return c, tea.Quit
		}
	}
	return c, nil
}

// View implements tea.Model
func (c *BaseComponent) View() string {
	return ""
}

// Styles for common UI patterns
type CommonStyles struct {
	Border      lipgloss.Style
	BorderFocus lipgloss.Style
	Title       lipgloss.Style
	Content     lipgloss.Style
	Footer      lipgloss.Style
}

// NewCommonStyles creates default common styles
func NewCommonStyles() CommonStyles {
	return CommonStyles{
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")),
			
		BorderFocus: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("69")),
			
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("69")).
			MarginBottom(1),
			
		Content: lipgloss.NewStyle().
			Padding(1),
			
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1),
	}
}

// Message types for component communication
type (
	// FocusMsg requests focus change
	FocusMsg struct {
		Component Component
	}
	
	// BlurMsg requests blur
	BlurMsg struct {
		Component Component
	}
	
	// ResizeMsg requests resize
	ResizeMsg struct {
		Component Component
		Width     int
		Height    int
	}
	
	// ThemeChangeMsg notifies of theme change
	ThemeChangeMsg struct {
		Theme interface{}
	}
	
	// SelectionChangeMsg notifies of selection change
	SelectionChangeMsg struct {
		Component Component
		Index     int
		Item      interface{}
	}
	
	// FilterChangeMsg notifies of filter change
	FilterChangeMsg struct {
		Component Component
		Filter    string
		Results   int
	}
	
	// AnimationMsg for animation updates
	AnimationMsg struct {
		Component Component
		Name      string
		Frame     int
	}
	
	// ErrorMsg for component errors
	ErrorMsg struct {
		Component Component
		Error     error
	}
)

// Command helpers
func FocusCmd(component Component) tea.Cmd {
	return func() tea.Msg {
		return FocusMsg{Component: component}
	}
}

func BlurCmd(component Component) tea.Cmd {
	return func() tea.Msg {
		return BlurMsg{Component: component}
	}
}

func ResizeCmd(component Component, width, height int) tea.Cmd {
	return func() tea.Msg {
		return ResizeMsg{
			Component: component,
			Width:     width,
			Height:    height,
		}
	}
}

func ThemeChangeCmd(theme interface{}) tea.Cmd {
	return func() tea.Msg {
		return ThemeChangeMsg{Theme: theme}
	}
}

func SelectionChangeCmd(component Component, index int, item interface{}) tea.Cmd {
	return func() tea.Msg {
		return SelectionChangeMsg{
			Component: component,
			Index:     index,
			Item:      item,
		}
	}
}

func FilterChangeCmd(component Component, filter string, results int) tea.Cmd {
	return func() tea.Msg {
		return FilterChangeMsg{
			Component: component,
			Filter:    filter,
			Results:   results,
		}
	}
}

func AnimationCmd(component Component, name string, frame int) tea.Cmd {
	return func() tea.Msg {
		return AnimationMsg{
			Component: component,
			Name:      name,
			Frame:     frame,
		}
	}
}

func ErrorCmd(component Component, err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{
			Component: component,
			Error:     err,
		}
	}
}