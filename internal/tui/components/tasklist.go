package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"

	"github.com/emiller/tasksh/internal/taskwarrior"
	"github.com/emiller/tasksh/internal/tui/theme"
)

// TaskList is a modern, filterable task list component
type TaskList struct {
	*BaseComponent
	
	// Data
	tasks         []taskwarrior.Task
	filteredTasks []taskwarrior.Task
	selected      int
	
	// State
	filter        string
	filterActive  bool
	
	// Styling
	theme         *theme.Theme
	styles        theme.TaskListStyles
	
	// Behavior
	multiSelect   bool
	selectedItems map[int]bool
	
	// Viewport
	offset        int
	visibleLines  int
	
	// Key bindings
	keyMap        TaskListKeyMap
}

// TaskListKeyMap defines key bindings for the task list
type TaskListKeyMap struct {
	Up         key.Binding
	Down       key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	Home       key.Binding
	End        key.Binding
	Select     key.Binding
	Filter     key.Binding
	ClearFilter key.Binding
	Enter      key.Binding
	Escape     key.Binding
}

// DefaultTaskListKeyMap returns default key bindings
func DefaultTaskListKeyMap() TaskListKeyMap {
	return TaskListKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdown", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear filter"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// NewTaskList creates a new task list component
func NewTaskList() *TaskList {
	base := NewBaseComponent()
	t := theme.GetTheme()
	
	return &TaskList{
		BaseComponent: base,
		tasks:         []taskwarrior.Task{},
		filteredTasks: []taskwarrior.Task{},
		selected:      0,
		selectedItems: make(map[int]bool),
		theme:         t,
		styles:        t.Components.TaskList,
		keyMap:        DefaultTaskListKeyMap(),
	}
}

// SetTasks updates the task list
func (tl *TaskList) SetTasks(tasks []taskwarrior.Task) {
	tl.tasks = tasks
	tl.applyFilter()
	tl.adjustSelection()
}

// Tasks returns the current tasks
func (tl *TaskList) Tasks() []taskwarrior.Task {
	return tl.tasks
}

// FilteredTasks returns the filtered tasks
func (tl *TaskList) FilteredTasks() []taskwarrior.Task {
	return tl.filteredTasks
}

// SelectedTask returns the currently selected task
func (tl *TaskList) SelectedTask() *taskwarrior.Task {
	if len(tl.filteredTasks) == 0 || tl.selected < 0 || tl.selected >= len(tl.filteredTasks) {
		return nil
	}
	return &tl.filteredTasks[tl.selected]
}

// SelectedIndex returns the selected index
func (tl *TaskList) SelectedIndex() int {
	return tl.selected
}

// SetSelected sets the selected index
func (tl *TaskList) SetSelected(index int) {
	if index < 0 {
		tl.selected = 0
	} else if index >= len(tl.filteredTasks) {
		tl.selected = len(tl.filteredTasks) - 1
	} else {
		tl.selected = index
	}
	tl.adjustViewport()
}

// SetFilter sets the filter string
func (tl *TaskList) SetFilter(filter string) {
	tl.filter = filter
	tl.applyFilter()
	tl.adjustSelection()
}

// Filter returns the current filter
func (tl *TaskList) Filter() string {
	return tl.filter
}

// ClearFilter clears the current filter
func (tl *TaskList) ClearFilter() {
	tl.filter = ""
	tl.filterActive = false
	tl.applyFilter()
	tl.adjustSelection()
}

// SetMultiSelect enables/disables multi-select mode
func (tl *TaskList) SetMultiSelect(enabled bool) {
	tl.multiSelect = enabled
	if !enabled {
		tl.selectedItems = make(map[int]bool)
	}
}

// ToggleSelection toggles selection of the current item
func (tl *TaskList) ToggleSelection() {
	if !tl.multiSelect || tl.selected < 0 || tl.selected >= len(tl.filteredTasks) {
		return
	}
	
	if tl.selectedItems[tl.selected] {
		delete(tl.selectedItems, tl.selected)
	} else {
		tl.selectedItems[tl.selected] = true
	}
}

// SelectedItems returns the indices of selected items
func (tl *TaskList) SelectedItems() []int {
	var indices []int
	for i := range tl.selectedItems {
		indices = append(indices, i)
	}
	return indices
}

// applyFilter filters tasks based on the current filter string
func (tl *TaskList) applyFilter() {
	if tl.filter == "" {
		tl.filteredTasks = tl.tasks
		return
	}
	
	// Create searchable strings for fuzzy matching
	var searchStrings []string
	for _, task := range tl.tasks {
		searchable := fmt.Sprintf("%s %s %s", 
			task.Description, 
			task.Project, 
			strings.Join(task.Tags, " "))
		searchStrings = append(searchStrings, searchable)
	}
	
	// Perform fuzzy search
	matches := fuzzy.Find(tl.filter, searchStrings)
	
	// Build filtered task list
	tl.filteredTasks = make([]taskwarrior.Task, len(matches))
	for i, match := range matches {
		tl.filteredTasks[i] = tl.tasks[match.Index]
	}
}

// adjustSelection ensures the selection is valid
func (tl *TaskList) adjustSelection() {
	if len(tl.filteredTasks) == 0 {
		tl.selected = 0
		return
	}
	
	if tl.selected >= len(tl.filteredTasks) {
		tl.selected = len(tl.filteredTasks) - 1
	}
	
	if tl.selected < 0 {
		tl.selected = 0
	}
	
	tl.adjustViewport()
}

// adjustViewport adjusts the viewport to show the selected item
func (tl *TaskList) adjustViewport() {
	if len(tl.filteredTasks) == 0 {
		return
	}
	
	// Calculate visible lines (leave space for borders and padding)
	tl.visibleLines = tl.height - 4
	if tl.visibleLines < 1 {
		tl.visibleLines = 1
	}
	
	// Adjust offset to keep selection visible
	if tl.selected < tl.offset {
		tl.offset = tl.selected
	} else if tl.selected >= tl.offset+tl.visibleLines {
		tl.offset = tl.selected - tl.visibleLines + 1
	}
	
	// Ensure offset is valid
	maxOffset := len(tl.filteredTasks) - tl.visibleLines
	if maxOffset < 0 {
		maxOffset = 0
	}
	if tl.offset > maxOffset {
		tl.offset = maxOffset
	}
	if tl.offset < 0 {
		tl.offset = 0
	}
}

// Init implements tea.Model
func (tl *TaskList) Init() tea.Cmd {
	return tl.BaseComponent.Init()
}

// Update implements tea.Model
func (tl *TaskList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Handle base component updates
	_, cmd := tl.BaseComponent.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tl.SetSize(msg.Width, msg.Height)
		tl.adjustViewport()
		
	case tea.KeyMsg:
		if tl.filterActive {
			return tl.handleFilterInput(msg)
		}
		return tl.handleKeyInput(msg)
		
	case ThemeChangeMsg:
		if t, ok := msg.Theme.(*theme.Theme); ok {
			tl.theme = t
			tl.styles = t.Components.TaskList
		}
	}
	
	return tl, tea.Batch(cmds...)
}

// handleKeyInput handles key input when not filtering
func (tl *TaskList) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, tl.keyMap.Up):
		if tl.selected > 0 {
			tl.SetSelected(tl.selected - 1)
		}
		
	case key.Matches(msg, tl.keyMap.Down):
		if tl.selected < len(tl.filteredTasks)-1 {
			tl.SetSelected(tl.selected + 1)
		}
		
	case key.Matches(msg, tl.keyMap.PageUp):
		tl.SetSelected(tl.selected - tl.visibleLines)
		
	case key.Matches(msg, tl.keyMap.PageDown):
		tl.SetSelected(tl.selected + tl.visibleLines)
		
	case key.Matches(msg, tl.keyMap.Home):
		tl.SetSelected(0)
		
	case key.Matches(msg, tl.keyMap.End):
		tl.SetSelected(len(tl.filteredTasks) - 1)
		
	case key.Matches(msg, tl.keyMap.Select):
		if tl.multiSelect {
			tl.ToggleSelection()
		}
		
	case key.Matches(msg, tl.keyMap.Filter):
		tl.filterActive = true
		
	case key.Matches(msg, tl.keyMap.ClearFilter):
		tl.ClearFilter()
		
	case key.Matches(msg, tl.keyMap.Enter):
		if task := tl.SelectedTask(); task != nil {
			return tl, SelectionChangeCmd(tl, tl.selected, *task)
		}
	}
	
	return tl, nil
}

// handleFilterInput handles key input when filtering
func (tl *TaskList) handleFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		tl.filterActive = false
		return tl, FilterChangeCmd(tl, tl.filter, len(tl.filteredTasks))
		
	case "esc":
		tl.filterActive = false
		tl.ClearFilter()
		
	case "backspace":
		if len(tl.filter) > 0 {
			tl.filter = tl.filter[:len(tl.filter)-1]
			tl.applyFilter()
			tl.adjustSelection()
		}
		
	default:
		if len(msg.Runes) > 0 {
			tl.filter += string(msg.Runes)
			tl.applyFilter()
			tl.adjustSelection()
		}
	}
	
	return tl, nil
}

// View implements tea.Model
func (tl *TaskList) View() string {
	if tl.width == 0 || tl.height == 0 {
		return ""
	}
	
	var content strings.Builder
	
	// Header with filter info
	if tl.filter != "" || tl.filterActive {
		filterText := tl.filter
		if tl.filterActive {
			filterText += "█" // cursor
		}
		header := fmt.Sprintf("Filter: %s (%d/%d)", 
			filterText, len(tl.filteredTasks), len(tl.tasks))
		content.WriteString(tl.theme.Styles.Caption.Render(header))
		content.WriteString("\n")
	}
	
	// Task list
	if len(tl.filteredTasks) == 0 {
		emptyMsg := "No tasks"
		if tl.filter != "" {
			emptyMsg = "No tasks match filter"
		}
		content.WriteString(tl.theme.Styles.Caption.Render(emptyMsg))
	} else {
		tl.renderTasks(&content)
	}
	
	// Apply container styling
	containerStyle := tl.styles.Item
	if tl.Focused() {
		containerStyle = containerStyle.Border(lipgloss.RoundedBorder()).
			BorderForeground(tl.theme.Colors.Primary)
	}
	
	return containerStyle.
		Width(tl.width - 2).
		Height(tl.height - 2).
		Render(content.String())
}

// renderTasks renders the visible tasks
func (tl *TaskList) renderTasks(content *strings.Builder) {
	end := tl.offset + tl.visibleLines
	if end > len(tl.filteredTasks) {
		end = len(tl.filteredTasks)
	}
	
	for i := tl.offset; i < end; i++ {
		task := tl.filteredTasks[i]
		line := tl.renderTask(task, i == tl.selected, tl.selectedItems[i])
		content.WriteString(line)
		if i < end-1 {
			content.WriteString("\n")
		}
	}
}

// renderTask renders a single task
func (tl *TaskList) renderTask(task taskwarrior.Task, selected, multiSelected bool) string {
	var parts []string
	
	// Multi-select indicator
	if tl.multiSelect {
		if multiSelected {
			parts = append(parts, "●")
		} else {
			parts = append(parts, "○")
		}
	}
	
	// Priority indicator
	if task.Priority != "" {
		var priorityStyle lipgloss.Style
		switch task.Priority {
		case "H":
			priorityStyle = lipgloss.NewStyle().Foreground(tl.theme.Colors.PriorityHigh)
		case "M":
			priorityStyle = lipgloss.NewStyle().Foreground(tl.theme.Colors.PriorityMedium)
		case "L":
			priorityStyle = lipgloss.NewStyle().Foreground(tl.theme.Colors.PriorityLow)
		}
		parts = append(parts, priorityStyle.Render("●"))
	}
	
	// Status indicator
	statusColor := tl.theme.Colors.TaskPending
	switch task.Status {
	case "completed":
		statusColor = tl.theme.Colors.TaskCompleted
	case "waiting":
		statusColor = tl.theme.Colors.TaskWaiting
	case "deleted":
		statusColor = tl.theme.Colors.TaskDeleted
	case "recurring":
		statusColor = tl.theme.Colors.TaskRecurring
	}
	
	// Description
	desc := tl.styles.Description.Foreground(statusColor).Render(task.Description)
	parts = append(parts, desc)
	
	// Project
	if task.Project != "" {
		project := tl.styles.Project.Render(fmt.Sprintf("(%s)", task.Project))
		parts = append(parts, project)
	}
	
	// Tags
	if len(task.Tags) > 0 {
		tags := tl.styles.Tags.Render(fmt.Sprintf("+%s", strings.Join(task.Tags, " +")))
		parts = append(parts, tags)
	}
	
	// Due date
	if task.Due != "" {
		if due, err := time.Parse("20060102T150405Z", task.Due); err == nil {
			dueStr := due.Format("Jan 2")
			if due.Before(time.Now()) {
				dueStr = tl.theme.Styles.Error.Render(dueStr + " (overdue)")
			} else if due.Before(time.Now().AddDate(0, 0, 7)) {
				dueStr = tl.theme.Styles.Warning.Render(dueStr)
			}
			parts = append(parts, dueStr)
		}
	}
	
	line := strings.Join(parts, " ")
	
	// Apply selection styling
	if selected {
		return tl.styles.ItemSelected.Render(line)
	}
	
	return tl.styles.Item.Render(line)
}

// Implement interfaces
var _ Component = (*TaskList)(nil)
var _ Selectable = (*TaskList)(nil)
var _ Scrollable = (*TaskList)(nil)

// Selectable interface implementation
func (tl *TaskList) Select(index int) tea.Cmd {
	tl.SetSelected(index)
	if task := tl.SelectedTask(); task != nil {
		return SelectionChangeCmd(tl, index, *task)
	}
	return nil
}

func (tl *TaskList) Selected() int {
	return tl.selected
}

func (tl *TaskList) SelectNext() tea.Cmd {
	if tl.selected < len(tl.filteredTasks)-1 {
		return tl.Select(tl.selected + 1)
	}
	return nil
}

func (tl *TaskList) SelectPrev() tea.Cmd {
	if tl.selected > 0 {
		return tl.Select(tl.selected - 1)
	}
	return nil
}

func (tl *TaskList) ItemCount() int {
	return len(tl.filteredTasks)
}

func (tl *TaskList) Item(index int) interface{} {
	if index < 0 || index >= len(tl.filteredTasks) {
		return nil
	}
	return tl.filteredTasks[index]
}

// Filterable interface implementation
func (tl *TaskList) SetFilterText(filter string) tea.Cmd {
	tl.SetFilter(filter)
	return FilterChangeCmd(tl, filter, len(tl.filteredTasks))
}

func (tl *TaskList) ClearFilterText() tea.Cmd {
	tl.ClearFilter()
	return FilterChangeCmd(tl, "", len(tl.filteredTasks))
}

func (tl *TaskList) FilteredCount() int {
	return len(tl.filteredTasks)
}

func (tl *TaskList) TotalCount() int {
	return len(tl.tasks)
}

// Scrollable interface implementation
func (tl *TaskList) ScrollUp(lines int) tea.Cmd {
	tl.SetSelected(tl.selected - lines)
	return nil
}

func (tl *TaskList) ScrollDown(lines int) tea.Cmd {
	tl.SetSelected(tl.selected + lines)
	return nil
}

func (tl *TaskList) ScrollToTop() tea.Cmd {
	tl.SetSelected(0)
	return nil
}

func (tl *TaskList) ScrollToBottom() tea.Cmd {
	tl.SetSelected(len(tl.filteredTasks) - 1)
	return nil
}

func (tl *TaskList) ScrollPosition() (offset, total int) {
	return tl.offset, len(tl.filteredTasks)
}

func (tl *TaskList) CanScrollUp() bool {
	return tl.selected > 0
}

func (tl *TaskList) CanScrollDown() bool {
	return tl.selected < len(tl.filteredTasks)-1
}