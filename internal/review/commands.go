package review

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/emiller/tasksh/internal/taskwarrior"
)

// CommandCategory represents a group of related commands
type CommandCategory string

const (
	CategoryNavigation CommandCategory = "Navigation"
	CategoryAction     CommandCategory = "Actions"
	CategoryModify     CommandCategory = "Modify"
	CategorySystem     CommandCategory = "System"
)

// ExtendedCommand is an enhanced command with more metadata
type ExtendedCommand struct {
	Name        string
	Description string
	Category    CommandCategory
	Keys        []string
	Hidden      bool
	RequiresConfirm bool
	Execute     func(*ImprovedModel) (tea.Model, tea.Cmd)
}

// CommandManager manages all available commands
type CommandManager struct {
	commands map[string]*ExtendedCommand
	byCategory map[CommandCategory][]*ExtendedCommand
}

// NewCommandManager creates a new command manager
func NewCommandManager() *CommandManager {
	cm := &CommandManager{
		commands:   make(map[string]*ExtendedCommand),
		byCategory: make(map[CommandCategory][]*ExtendedCommand),
	}
	
	// Register all default commands
	cm.registerDefaultCommands()
	
	return cm
}

// Register adds a command
func (cm *CommandManager) Register(cmd *ExtendedCommand) {
	// Register by each key
	for _, key := range cmd.Keys {
		cm.commands[key] = cmd
	}
	
	// Add to category
	cm.byCategory[cmd.Category] = append(cm.byCategory[cmd.Category], cmd)
}

// Get retrieves a command by key
func (cm *CommandManager) Get(key string) *ExtendedCommand {
	return cm.commands[key]
}

// GetByCategory returns all commands in a category
func (cm *CommandManager) GetByCategory(cat CommandCategory) []*ExtendedCommand {
	return cm.byCategory[cat]
}

// registerDefaultCommands sets up all the default commands
func (cm *CommandManager) registerDefaultCommands() {
	// Navigation commands
	cm.Register(&ExtendedCommand{
		Name:        "Next Task",
		Description: "Move to the next task",
		Category:    CategoryNavigation,
		Keys:        []string{"j", "down"},
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			if m.current < len(m.tasks)-1 {
				m.current++
				return m, m.loadCurrentTask()
			}
			return m, ShowInfo("Already at last task")
		},
	})
	
	cm.Register(&ExtendedCommand{
		Name:        "Previous Task",
		Description: "Move to the previous task",
		Category:    CategoryNavigation,
		Keys:        []string{"k", "up"},
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			if m.current > 0 {
				m.current--
				return m, m.loadCurrentTask()
			}
			return m, ShowInfo("Already at first task")
		},
	})
	
	cm.Register(&ExtendedCommand{
		Name:        "Jump to Task",
		Description: "Jump to task by number",
		Category:    CategoryNavigation,
		Keys:        []string{"g"},
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			m.mode = ModeInput
			m.input.Placeholder = "Enter task number (1-" + fmt.Sprintf("%d", len(m.tasks)) + ")"
			m.input.Reset()
			m.input.Focus()
			m.message = "Jump to task:"
			return m, nil
		},
	})
	
	// Action commands
	cm.Register(&ExtendedCommand{
		Name:        "Review",
		Description: "Mark task as reviewed",
		Category:    CategoryAction,
		Keys:        []string{"r"},
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			if m.current >= len(m.tasks) {
				return m, nil
			}
			uuid := m.tasks[m.current]
			return m, tea.Batch(
				func() tea.Msg {
					if err := taskwarrior.MarkTaskReviewed(uuid); err != nil {
						return ErrorMsg{Error: err}
					}
					return ActionCompletedMsg{Message: "Task marked as reviewed"}
				},
				ShowSuccess("Task reviewed"),
			)
		},
	})
	
	cm.Register(&ExtendedCommand{
		Name:        "Complete",
		Description: "Complete the current task",
		Category:    CategoryAction,
		Keys:        []string{"c"},
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			if m.current >= len(m.tasks) {
				return m, nil
			}
			uuid := m.tasks[m.current]
			return m, tea.Batch(
				func() tea.Msg {
					if err := taskwarrior.CompleteTask(uuid); err != nil {
						return ErrorMsg{Error: err}
					}
					return ActionCompletedMsg{Message: "Task completed"}
				},
				ShowSuccess("Task completed!"),
			)
		},
	})
	
	cm.Register(&ExtendedCommand{
		Name:        "Delete",
		Description: "Delete the current task",
		Category:    CategoryAction,
		Keys:        []string{"d"},
		RequiresConfirm: true,
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			if m.current >= len(m.tasks) {
				return m, nil
			}
			
			// Set up confirmation
			m.mode = ModeConfirm
			m.message = "Delete this task? This cannot be undone."
			
			return m, nil
		},
	})
	
	cm.Register(&ExtendedCommand{
		Name:        "Skip",
		Description: "Skip to next task without reviewing",
		Category:    CategoryAction,
		Keys:        []string{"s"},
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			return m, tea.Batch(
				m.nextTask(),
				ShowInfo("Task skipped"),
			)
		},
	})
	
	// Modify commands
	cm.Register(&ExtendedCommand{
		Name:        "Edit",
		Description: "Edit task in external editor",
		Category:    CategoryModify,
		Keys:        []string{"e"},
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			if m.current >= len(m.tasks) {
				return m, nil
			}
			uuid := m.tasks[m.current]
			
			return m, tea.ExecProcess(
				taskwarrior.CreateEditCommand(uuid),
				func(err error) tea.Msg {
					if err != nil {
						return ErrorMsg{Error: err}
					}
					// Mark as reviewed after edit
					taskwarrior.MarkTaskReviewed(uuid)
					return ActionCompletedMsg{Message: "Task edited and reviewed"}
				},
			)
		},
	})
	
	cm.Register(&ExtendedCommand{
		Name:        "Modify",
		Description: "Modify task attributes",
		Category:    CategoryModify,
		Keys:        []string{"m"},
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			m.mode = ModeInput
			m.input.Placeholder = "Enter modification (e.g., +tag, priority:H)"
			m.input.Reset()
			m.input.Focus()
			m.message = "Modify task:"
			return m, nil
		},
	})
	
	cm.Register(&ExtendedCommand{
		Name:        "Wait",
		Description: "Set task wait date",
		Category:    CategoryModify,
		Keys:        []string{"w"},
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			m.mode = ModeInput
			m.input.Placeholder = "Enter wait date (e.g., tomorrow, 1week, 2024-12-25)"
			m.input.Reset()
			m.input.Focus()
			m.message = "Set wait until:"
			return m, nil
		},
	})
	
	cm.Register(&ExtendedCommand{
		Name:        "Due",
		Description: "Set task due date",
		Category:    CategoryModify,
		Keys:        []string{"u"},
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			m.mode = ModeInput
			m.input.Placeholder = "Enter due date (e.g., tomorrow, 1week, 2024-12-25)"
			m.input.Reset()
			m.input.Focus()
			m.message = "Set due date:"
			return m, nil
		},
	})
	
	// System commands
	cm.Register(&ExtendedCommand{
		Name:        "Help",
		Description: "Show help",
		Category:    CategorySystem,
		Keys:        []string{"?"},
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			if m.mode == ModeHelp {
				m.mode = ModeNormal
			} else {
				m.mode = ModeHelp
			}
			return m, nil
		},
	})
	
	cm.Register(&ExtendedCommand{
		Name:        "Quit",
		Description: "Quit the review",
		Category:    CategorySystem,
		Keys:        []string{"q", "ctrl+c"},
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			m.quitting = true
			return m, tea.Quit
		},
	})
	
	cm.Register(&ExtendedCommand{
		Name:        "Undo",
		Description: "Undo last action",
		Category:    CategorySystem,
		Keys:        []string{"z"},
		Execute: func(m *ImprovedModel) (tea.Model, tea.Cmd) {
			return m, tea.Batch(
				func() tea.Msg {
					if err := taskwarrior.UndoLastAction(); err != nil {
						return ErrorMsg{Error: err}
					}
					return ActionCompletedMsg{Message: "Last action undone"}
				},
				ShowInfo("Last action undone"),
				m.loadCurrentTask(), // Reload current task
			)
		},
	})
}

// GetVisibleCommands returns commands that should be shown in help
func (cm *CommandManager) GetVisibleCommands() map[CommandCategory][]*ExtendedCommand {
	visible := make(map[CommandCategory][]*ExtendedCommand)
	
	for cat, cmds := range cm.byCategory {
		var visibleCmds []*ExtendedCommand
		for _, cmd := range cmds {
			if !cmd.Hidden {
				visibleCmds = append(visibleCmds, cmd)
			}
		}
		if len(visibleCmds) > 0 {
			visible[cat] = visibleCmds
		}
	}
	
	return visible
}

// ExecuteCommand safely executes a command by key
func (cm *CommandManager) ExecuteCommand(key string, model *ImprovedModel) (tea.Model, tea.Cmd) {
	cmd := cm.Get(key)
	if cmd == nil {
		return model, nil
	}
	
	// Check if confirmation is needed
	if cmd.RequiresConfirm && model.mode != ModeConfirm {
		// This should trigger confirmation mode
		return cmd.Execute(model)
	}
	
	return cmd.Execute(model)
}