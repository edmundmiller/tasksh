package review

import (
	"testing"
	tea "github.com/charmbracelet/bubbletea"
)

// TestCommandManager tests the command manager
func TestCommandManager(t *testing.T) {
	cm := NewCommandManager()
	
	if cm == nil {
		t.Fatal("NewCommandManager returned nil")
	}
	
	// Check that default commands are registered
	if len(cm.commands) == 0 {
		t.Error("No commands registered")
	}
	
	// Check specific commands exist
	commands := []string{"r", "c", "s", "e", "d", "?", "q"}
	for _, key := range commands {
		if cmd := cm.Get(key); cmd == nil {
			t.Errorf("Command %q not found", key)
		}
	}
}

// TestCommandCategories tests command categorization
func TestCommandCategories(t *testing.T) {
	cm := NewCommandManager()
	
	// Check that categories have commands
	categories := []CommandCategory{
		CategoryNavigation,
		CategoryAction,
		CategoryModify,
		CategorySystem,
	}
	
	for _, cat := range categories {
		cmds := cm.GetByCategory(cat)
		if len(cmds) == 0 {
			t.Errorf("No commands in category %v", cat)
		}
	}
}

// TestGetVisibleCommands tests filtering hidden commands
func TestGetVisibleCommands(t *testing.T) {
	cm := NewCommandManager()
	
	// Add a hidden command
	cm.Register(&ExtendedCommand{
		Name:     "Hidden",
		Category: CategorySystem,
		Keys:     []string{"hidden"},
		Hidden:   true,
		Execute:  func(m *ImprovedModel) (tea.Model, tea.Cmd) { return m, nil },
	})
	
	visible := cm.GetVisibleCommands()
	
	// Check that hidden command is not in visible list
	for _, cmds := range visible {
		for _, cmd := range cmds {
			if cmd.Hidden {
				t.Error("Hidden command found in visible commands")
			}
		}
	}
}

// TestCommandExecution tests safe command execution
func TestCommandExecution(t *testing.T) {
	cm := NewCommandManager()
	m := NewImprovedModel()
	
	// Test executing a valid command
	model, _ := cm.ExecuteCommand("?", m)
	if model == nil {
		t.Error("ExecuteCommand returned nil model")
	}
	
	// The help command should toggle help mode
	// Check the returned model, not the original
	if improvedModel, ok := model.(*ImprovedModel); ok {
		if improvedModel.mode != ModeHelp {
			t.Error("Help mode should have been toggled to ModeHelp")
		}
	} else {
		t.Error("Returned model is not an ImprovedModel")
	}
	
	// Test executing an invalid command
	model, cmd := cm.ExecuteCommand("invalid", m)
	if model == nil {
		t.Error("ExecuteCommand should return model even for invalid command")
	}
	if cmd != nil {
		t.Error("Invalid command should return nil cmd")
	}
}