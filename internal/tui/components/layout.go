package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LayoutManager defines how components are arranged within a container
type LayoutManager interface {
	// Arrange positions and sizes child components within the given bounds
	Arrange(container Container, width, height int) []ComponentBounds
	
	// MinSize returns the minimum size required for this layout
	MinSize(container Container) (width, height int)
	
	// PreferredSize returns the preferred size for this layout
	PreferredSize(container Container) (width, height int)
}

// VerticalLayout arranges components vertically
type VerticalLayout struct {
	// Spacing between components
	Spacing int
	
	// Alignment for components
	Alignment lipgloss.Position
	
	// Whether to distribute extra space evenly
	Distribute bool
}

// NewVerticalLayout creates a new vertical layout
func NewVerticalLayout() *VerticalLayout {
	return &VerticalLayout{
		Spacing:   0,
		Alignment: lipgloss.Left,
		Distribute: false,
	}
}

// SetSpacing sets the spacing between components
func (vl *VerticalLayout) SetSpacing(spacing int) *VerticalLayout {
	vl.Spacing = spacing
	return vl
}

// SetAlignment sets the alignment for components
func (vl *VerticalLayout) SetAlignment(alignment lipgloss.Position) *VerticalLayout {
	vl.Alignment = alignment
	return vl
}

// SetDistribute sets whether to distribute extra space
func (vl *VerticalLayout) SetDistribute(distribute bool) *VerticalLayout {
	vl.Distribute = distribute
	return vl
}

// Arrange implements Layout
func (vl *VerticalLayout) Arrange(container Container, width, height int) []ComponentBounds {
	children := container.Children()
	if len(children) == 0 {
		return []ComponentBounds{}
	}
	
	var bounds []ComponentBounds
	
	// Calculate total preferred height and find flexible components
	totalPreferredHeight := 0
	flexibleComponents := 0
	
	for _, child := range children {
		if vm, ok := child.(ViewModel); ok {
			_, preferredHeight := vm.PreferredSize()
			totalPreferredHeight += preferredHeight
		} else {
			totalPreferredHeight += 3 // Default height
		}
		
		// Check if component can grow
		if _, ok := child.(Scrollable); ok {
			flexibleComponents++
		}
	}
	
	// Add spacing
	totalSpacing := (len(children) - 1) * vl.Spacing
	totalPreferredHeight += totalSpacing
	
	// Calculate available extra space
	extraSpace := height - totalPreferredHeight
	extraPerComponent := 0
	
	if vl.Distribute && flexibleComponents > 0 && extraSpace > 0 {
		extraPerComponent = extraSpace / flexibleComponents
	}
	
	// Position components
	currentY := 0
	
	for _, child := range children {
		var componentWidth, componentHeight int
		
		// Determine component size
		if vm, ok := child.(ViewModel); ok {
			preferredWidth, preferredHeight := vm.PreferredSize()
			componentWidth = min(preferredWidth, width)
			componentHeight = preferredHeight
		} else {
			componentWidth = width
			componentHeight = 3
		}
		
		// Add extra space if distributing and component is flexible
		if vl.Distribute && extraPerComponent > 0 {
			if _, ok := child.(Scrollable); ok {
				componentHeight += extraPerComponent
			}
		}
		
		// Calculate X position based on alignment
		var componentX int
		switch vl.Alignment {
		case lipgloss.Left:
			componentX = 0
		case lipgloss.Center:
			componentX = (width - componentWidth) / 2
		case lipgloss.Right:
			componentX = width - componentWidth
		}
		
		// Ensure component fits
		if currentY + componentHeight > height {
			componentHeight = height - currentY
		}
		
		if componentHeight > 0 {
			bounds = append(bounds, ComponentBounds{
				Component: child,
				X:         componentX,
				Y:         currentY,
				Width:     componentWidth,
				Height:    componentHeight,
			})
		}
		
		currentY += componentHeight + vl.Spacing
		
		// Stop if we've run out of space
		if currentY >= height {
			break
		}
	}
	
	return bounds
}

// MinSize implements Layout
func (vl *VerticalLayout) MinSize(container Container) (width, height int) {
	children := container.Children()
	if len(children) == 0 {
		return 0, 0
	}
	
	maxWidth := 0
	totalHeight := 0
	
	for i, child := range children {
		var minWidth, minHeight int
		
		if vm, ok := child.(ViewModel); ok {
			minWidth = vm.MinWidth()
			minHeight = vm.MinHeight()
		} else {
			minWidth = 10
			minHeight = 1
		}
		
		if minWidth > maxWidth {
			maxWidth = minWidth
		}
		
		totalHeight += minHeight
		
		// Add spacing
		if i > 0 {
			totalHeight += vl.Spacing
		}
	}
	
	return maxWidth, totalHeight
}

// PreferredSize implements Layout
func (vl *VerticalLayout) PreferredSize(container Container) (width, height int) {
	children := container.Children()
	if len(children) == 0 {
		return 0, 0
	}
	
	maxWidth := 0
	totalHeight := 0
	
	for i, child := range children {
		var preferredWidth, preferredHeight int
		
		if vm, ok := child.(ViewModel); ok {
			preferredWidth, preferredHeight = vm.PreferredSize()
		} else {
			preferredWidth = 40
			preferredHeight = 3
		}
		
		if preferredWidth > maxWidth {
			maxWidth = preferredWidth
		}
		
		totalHeight += preferredHeight
		
		// Add spacing
		if i > 0 {
			totalHeight += vl.Spacing
		}
	}
	
	return maxWidth, totalHeight
}

// HorizontalLayout arranges components horizontally
type HorizontalLayout struct {
	// Spacing between components
	Spacing int
	
	// Alignment for components
	Alignment lipgloss.Position
	
	// Whether to distribute extra space evenly
	Distribute bool
}

// NewHorizontalLayout creates a new horizontal layout
func NewHorizontalLayout() *HorizontalLayout {
	return &HorizontalLayout{
		Spacing:   1,
		Alignment: lipgloss.Top,
		Distribute: false,
	}
}

// SetSpacing sets the spacing between components
func (hl *HorizontalLayout) SetSpacing(spacing int) *HorizontalLayout {
	hl.Spacing = spacing
	return hl
}

// SetAlignment sets the alignment for components
func (hl *HorizontalLayout) SetAlignment(alignment lipgloss.Position) *HorizontalLayout {
	hl.Alignment = alignment
	return hl
}

// SetDistribute sets whether to distribute extra space
func (hl *HorizontalLayout) SetDistribute(distribute bool) *HorizontalLayout {
	hl.Distribute = distribute
	return hl
}

// Arrange implements Layout
func (hl *HorizontalLayout) Arrange(container Container, width, height int) []ComponentBounds {
	children := container.Children()
	if len(children) == 0 {
		return []ComponentBounds{}
	}
	
	var bounds []ComponentBounds
	
	// Calculate total preferred width and find flexible components
	totalPreferredWidth := 0
	flexibleComponents := 0
	
	for _, child := range children {
		if vm, ok := child.(ViewModel); ok {
			preferredWidth, _ := vm.PreferredSize()
			totalPreferredWidth += preferredWidth
		} else {
			totalPreferredWidth += 20 // Default width
		}
		
		// Check if component can grow
		if _, ok := child.(Scrollable); ok {
			flexibleComponents++
		}
	}
	
	// Add spacing
	totalSpacing := (len(children) - 1) * hl.Spacing
	totalPreferredWidth += totalSpacing
	
	// Calculate available extra space
	extraSpace := width - totalPreferredWidth
	extraPerComponent := 0
	
	if hl.Distribute && flexibleComponents > 0 && extraSpace > 0 {
		extraPerComponent = extraSpace / flexibleComponents
	}
	
	// Position components
	currentX := 0
	
	for _, child := range children {
		var componentWidth, componentHeight int
		
		// Determine component size
		if vm, ok := child.(ViewModel); ok {
			preferredWidth, preferredHeight := vm.PreferredSize()
			componentWidth = preferredWidth
			componentHeight = min(preferredHeight, height)
		} else {
			componentWidth = 20
			componentHeight = height
		}
		
		// Add extra space if distributing and component is flexible
		if hl.Distribute && extraPerComponent > 0 {
			if _, ok := child.(Scrollable); ok {
				componentWidth += extraPerComponent
			}
		}
		
		// Calculate Y position based on alignment
		var componentY int
		switch hl.Alignment {
		case lipgloss.Top:
			componentY = 0
		case lipgloss.Center:
			componentY = (height - componentHeight) / 2
		case lipgloss.Bottom:
			componentY = height - componentHeight
		}
		
		// Ensure component fits
		if currentX + componentWidth > width {
			componentWidth = width - currentX
		}
		
		if componentWidth > 0 {
			bounds = append(bounds, ComponentBounds{
				Component: child,
				X:         currentX,
				Y:         componentY,
				Width:     componentWidth,
				Height:    componentHeight,
			})
		}
		
		currentX += componentWidth + hl.Spacing
		
		// Stop if we've run out of space
		if currentX >= width {
			break
		}
	}
	
	return bounds
}

// MinSize implements Layout
func (hl *HorizontalLayout) MinSize(container Container) (width, height int) {
	children := container.Children()
	if len(children) == 0 {
		return 0, 0
	}
	
	totalWidth := 0
	maxHeight := 0
	
	for i, child := range children {
		var minWidth, minHeight int
		
		if vm, ok := child.(ViewModel); ok {
			minWidth = vm.MinWidth()
			minHeight = vm.MinHeight()
		} else {
			minWidth = 10
			minHeight = 1
		}
		
		totalWidth += minWidth
		
		if minHeight > maxHeight {
			maxHeight = minHeight
		}
		
		// Add spacing
		if i > 0 {
			totalWidth += hl.Spacing
		}
	}
	
	return totalWidth, maxHeight
}

// PreferredSize implements Layout
func (hl *HorizontalLayout) PreferredSize(container Container) (width, height int) {
	children := container.Children()
	if len(children) == 0 {
		return 0, 0
	}
	
	totalWidth := 0
	maxHeight := 0
	
	for i, child := range children {
		var preferredWidth, preferredHeight int
		
		if vm, ok := child.(ViewModel); ok {
			preferredWidth, preferredHeight = vm.PreferredSize()
		} else {
			preferredWidth = 20
			preferredHeight = 3
		}
		
		totalWidth += preferredWidth
		
		if preferredHeight > maxHeight {
			maxHeight = preferredHeight
		}
		
		// Add spacing
		if i > 0 {
			totalWidth += hl.Spacing
		}
	}
	
	return totalWidth, maxHeight
}

// BorderLayout arranges components in border positions (top, bottom, left, right, center)
type BorderLayout struct {
	// Spacing around components
	Spacing int
}

// BorderPosition defines where a component should be placed
type BorderPosition int

const (
	BorderTop BorderPosition = iota
	BorderBottom
	BorderLeft
	BorderRight
	BorderCenter
)

// BorderComponent associates a component with a border position
type BorderComponent struct {
	*BaseComponent
	Component Component
	Position  BorderPosition
	Size      int // Fixed size for top/bottom (height) or left/right (width)
}

// NewBorderLayout creates a new border layout
func NewBorderLayout() *BorderLayout {
	return &BorderLayout{
		Spacing: 0,
	}
}

// SetSpacing sets the spacing around components
func (bl *BorderLayout) SetSpacing(spacing int) *BorderLayout {
	bl.Spacing = spacing
	return bl
}

// Arrange implements Layout
func (bl *BorderLayout) Arrange(container Container, width, height int) []ComponentBounds {
	children := container.Children()
	if len(children) == 0 {
		return []ComponentBounds{}
	}
	
	var bounds []ComponentBounds
	
	// Separate components by position
	var top, bottom, left, right, center []Component
	componentSizes := make(map[Component]int)
	
	for _, child := range children {
		if bc, ok := child.(*BorderComponent); ok {
			switch bc.Position {
			case BorderTop:
				top = append(top, bc.Component)
				componentSizes[bc.Component] = bc.Size
			case BorderBottom:
				bottom = append(bottom, bc.Component)
				componentSizes[bc.Component] = bc.Size
			case BorderLeft:
				left = append(left, bc.Component)
				componentSizes[bc.Component] = bc.Size
			case BorderRight:
				right = append(right, bc.Component)
				componentSizes[bc.Component] = bc.Size
			case BorderCenter:
				center = append(center, bc.Component)
			}
		} else {
			// Default to center
			center = append(center, child)
		}
	}
	
	// Calculate available space
	availableWidth := width
	availableHeight := height
	currentY := 0
	currentX := 0
	
	// Place top components
	for _, comp := range top {
		size := componentSizes[comp]
		if size == 0 {
			if vm, ok := comp.(ViewModel); ok {
				_, size = vm.PreferredSize()
			} else {
				size = 3
			}
		}
		
		bounds = append(bounds, ComponentBounds{
			Component: comp,
			X:         currentX,
			Y:         currentY,
			Width:     availableWidth,
			Height:    size,
		})
		
		currentY += size + bl.Spacing
		availableHeight -= size + bl.Spacing
	}
	
	// Place bottom components (from bottom up)
	bottomY := height
	for i := len(bottom) - 1; i >= 0; i-- {
		comp := bottom[i]
		size := componentSizes[comp]
		if size == 0 {
			if vm, ok := comp.(ViewModel); ok {
				_, size = vm.PreferredSize()
			} else {
				size = 3
			}
		}
		
		bottomY -= size
		bounds = append(bounds, ComponentBounds{
			Component: comp,
			X:         currentX,
			Y:         bottomY,
			Width:     availableWidth,
			Height:    size,
		})
		
		bottomY -= bl.Spacing
		availableHeight -= size + bl.Spacing
	}
	
	// Place left components
	for _, comp := range left {
		size := componentSizes[comp]
		if size == 0 {
			if vm, ok := comp.(ViewModel); ok {
				size, _ = vm.PreferredSize()
			} else {
				size = 20
			}
		}
		
		bounds = append(bounds, ComponentBounds{
			Component: comp,
			X:         currentX,
			Y:         currentY,
			Width:     size,
			Height:    availableHeight,
		})
		
		currentX += size + bl.Spacing
		availableWidth -= size + bl.Spacing
	}
	
	// Place right components (from right)
	rightX := width
	for i := len(right) - 1; i >= 0; i-- {
		comp := right[i]
		size := componentSizes[comp]
		if size == 0 {
			if vm, ok := comp.(ViewModel); ok {
				size, _ = vm.PreferredSize()
			} else {
				size = 20
			}
		}
		
		rightX -= size
		bounds = append(bounds, ComponentBounds{
			Component: comp,
			X:         rightX,
			Y:         currentY,
			Width:     size,
			Height:    availableHeight,
		})
		
		rightX -= bl.Spacing
		availableWidth -= size + bl.Spacing
	}
	
	// Place center components
	if len(center) > 0 {
		// Use vertical layout for center components
		centerLayout := NewVerticalLayout().SetDistribute(true)
		centerBounds := centerLayout.Arrange(&simpleContainer{children: center}, availableWidth, availableHeight)
		
		// Adjust positions to account for border components
		for _, bound := range centerBounds {
			bound.X += currentX
			bound.Y += currentY
			bounds = append(bounds, bound)
		}
	}
	
	return bounds
}

// MinSize implements Layout
func (bl *BorderLayout) MinSize(container Container) (width, height int) {
	// For border layout, minimum size is the sum of all border components
	// plus the minimum size of the center
	return 40, 10 // Reasonable minimum
}

// PreferredSize implements Layout
func (bl *BorderLayout) PreferredSize(container Container) (width, height int) {
	// For border layout, preferred size accommodates all components
	return 80, 24 // Reasonable default
}

// simpleContainer is a basic container implementation for layout calculations
type simpleContainer struct {
	children []Component
}

func (sc *simpleContainer) Children() []Component {
	return sc.children
}

func (sc *simpleContainer) AddChild(child Component) error {
	sc.children = append(sc.children, child)
	return nil
}

func (sc *simpleContainer) RemoveChild(child Component) error {
	for i, c := range sc.children {
		if c == child {
			sc.children = append(sc.children[:i], sc.children[i+1:]...)
			return nil
		}
	}
	return nil
}

// SetLayout sets the layout manager
func (sc *simpleContainer) SetLayout(layout LayoutManager) {}
func (sc *simpleContainer) Layout() LayoutManager { return nil }

// Implement required methods for Component interface
func (sc *simpleContainer) Init() tea.Cmd { return nil }
func (sc *simpleContainer) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return sc, nil }
func (sc *simpleContainer) View() string { return "" }
func (sc *simpleContainer) SetSize(width, height int) {}
func (sc *simpleContainer) Focus() tea.Cmd { return nil }
func (sc *simpleContainer) Blur() {}
func (sc *simpleContainer) Focused() bool { return false }
func (sc *simpleContainer) SetTheme(theme interface{}) {}

var _ Container = (*simpleContainer)(nil)