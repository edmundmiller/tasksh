package review

import (
	"github.com/charmbracelet/bubbles/key"
)

// ImprovedKeyMap provides a simplified key map with primary actions
type ImprovedKeyMap struct {
	KeyMap
	aiAvailable bool
}

// PrimaryActions returns only the most important actions for short help
func (i ImprovedKeyMap) PrimaryActions() []key.Binding {
	// Show only the 6 most important actions
	return []key.Binding{
		i.Review,
		i.Complete,
		i.Edit,
		i.Skip,
		i.Help,
		i.Quit,
	}
}

// ShortHelp returns simplified help text
func (i ImprovedKeyMap) ShortHelp() []key.Binding {
	return i.PrimaryActions()
}

// FullHelp returns complete help organized by category
func (i ImprovedKeyMap) FullHelp() [][]key.Binding {
	// Organize by action categories for better UX
	navigation := []key.Binding{i.NextTask, i.PrevTask}
	
	primary := []key.Binding{i.Review, i.Complete, i.Edit}
	
	taskManagement := []key.Binding{i.Modify, i.Delete, i.Wait, i.Due, i.Skip}
	
	advanced := []key.Binding{i.Context, i.Undo}
	if i.aiAvailable {
		advanced = append(advanced, i.AIAnalysis, i.PromptAgent)
	}
	
	system := []key.Binding{i.Help, i.Quit}
	
	return [][]key.Binding{
		navigation,
		primary,
		taskManagement,
		advanced,
		system,
	}
}

// GetHelpCategories returns help bindings with category labels
func (i ImprovedKeyMap) GetHelpCategories() []struct {
	Label    string
	Bindings []key.Binding
} {
	categories := []struct {
		Label    string
		Bindings []key.Binding
	}{
		{"Navigation", []key.Binding{i.NextTask, i.PrevTask}},
		{"Primary Actions", []key.Binding{i.Review, i.Complete, i.Edit}},
		{"Task Management", []key.Binding{i.Modify, i.Delete, i.Wait, i.Due, i.Skip}},
	}
	
	// Advanced features (conditional)
	advanced := []key.Binding{i.Context, i.Undo}
	if i.aiAvailable {
		advanced = append(advanced, i.AIAnalysis, i.PromptAgent)
	}
	categories = append(categories, struct {
		Label    string
		Bindings []key.Binding
	}{"Advanced", advanced})
	
	categories = append(categories, struct {
		Label    string
		Bindings []key.Binding
	}{"System", []key.Binding{i.Help, i.Quit}})
	
	return categories
}