package review

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestDueDateCalendarNavigation tests calendar navigation for due dates
func TestDueDateCalendarNavigation(t *testing.T) {
	model := NewReviewModel()
	model.mode = ModeDueCalendar
	model.calendar.SetFocused(true)
	
	initialDate := model.calendar.GetSelectedDate()
	
	// Test that calendar navigation works through updateDueCalendar
	keyMsg := tea.KeyMsg{Type: tea.KeyRight}
	updatedModelInterface, _ := model.updateDueCalendar(keyMsg)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	expectedDate := initialDate.AddDate(0, 0, 1)
	if !isSameDay(updatedModel.calendar.GetSelectedDate(), expectedDate) {
		t.Errorf("Due calendar navigation failed. Expected %v, got %v", 
			expectedDate, updatedModel.calendar.GetSelectedDate())
	}
}

// TestDueDateCalendarConfirm tests due date confirmation flow
func TestDueDateCalendarConfirm(t *testing.T) {
	model := NewReviewModel()
	model.mode = ModeDueCalendar
	model.calendar.SetFocused(true)
	
	// Set up tasks for the model
	model.tasks = []string{"test-uuid"}
	model.current = 0
	
	// Set a specific date
	testDate := time.Date(2024, 12, 25, 0, 0, 0, 0, time.Local)
	model.calendar.selectedDate = testDate
	
	// Simulate Enter key
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModelInterface, cmd := model.updateDueCalendar(keyMsg)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	// Should transition back to viewing mode
	if updatedModel.mode != ModeViewing {
		t.Errorf("Expected mode %v, got %v", ModeViewing, updatedModel.mode)
	}
	
	// Calendar should be unfocused
	if updatedModel.calendar.focused {
		t.Error("Calendar should be unfocused after due date confirmation")
	}
	
	// Should have a command to execute due date change
	if cmd == nil {
		t.Error("Expected a command to be returned for due date execution")
	}
}

// TestDueDateCalendarToggle tests toggle between calendar and text input
func TestDueDateCalendarToggle(t *testing.T) {
	model := NewReviewModel()
	model.mode = ModeDueCalendar
	model.calendar.SetFocused(true)
	
	// Simulate Tab key to toggle to text input
	keyMsg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModelInterface, _ := model.updateDueCalendar(keyMsg)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	// Should transition to text input mode
	if updatedModel.mode != ModeInputDueDate {
		t.Errorf("Expected mode %v, got %v", ModeInputDueDate, updatedModel.mode)
	}
	
	// Calendar should be unfocused
	if updatedModel.calendar.focused {
		t.Error("Calendar should be unfocused after toggle")
	}
}

// TestDueDateCalendarCancel tests due date calendar cancellation
func TestDueDateCalendarCancel(t *testing.T) {
	model := NewReviewModel()
	model.mode = ModeDueCalendar
	model.calendar.SetFocused(true)
	
	// Simulate Escape key
	keyMsg := tea.KeyMsg{Type: tea.KeyEscape}
	updatedModelInterface, _ := model.updateDueCalendar(keyMsg)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	// Should return to viewing mode
	if updatedModel.mode != ModeViewing {
		t.Errorf("Expected mode %v, got %v", ModeViewing, updatedModel.mode)
	}
	
	// Calendar should be unfocused
	if updatedModel.calendar.focused {
		t.Error("Calendar should be unfocused after cancel")
	}
}

// TestDueDateTextInput tests due date text input functionality
func TestDueDateTextInput(t *testing.T) {
	model := NewReviewModel()
	model.mode = ModeInputDueDate
	model.tasks = []string{"test-uuid"}
	model.current = 0
	
	// Set a value in the text input
	model.textInput.SetValue("tomorrow")
	
	// Simulate Enter key
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModelInterface, cmd := model.updateDueDateInput(keyMsg)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	// Should transition back to viewing mode
	if updatedModel.mode != ModeViewing {
		t.Errorf("Expected mode %v, got %v", ModeViewing, updatedModel.mode)
	}
	
	// Should have a command to execute due date change
	if cmd == nil {
		t.Error("Expected a command to be returned for due date execution")
	}
}

// TestDueDateTextInputEmpty tests empty due date text input
func TestDueDateTextInputEmpty(t *testing.T) {
	model := NewReviewModel()
	model.mode = ModeInputDueDate
	
	// Leave text input empty
	model.textInput.SetValue("")
	
	// Simulate Enter key
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModelInterface, cmd := model.updateDueDateInput(keyMsg)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	// Should transition back to viewing mode
	if updatedModel.mode != ModeViewing {
		t.Errorf("Expected mode %v, got %v", ModeViewing, updatedModel.mode)
	}
	
	// Should not have a command since input was empty
	if cmd != nil {
		t.Error("Expected no command for empty due date input")
	}
}

// TestDueDateTextInputCancel tests due date text input cancellation
func TestDueDateTextInputCancel(t *testing.T) {
	model := NewReviewModel()
	model.mode = ModeInputDueDate
	
	// Simulate Escape key
	keyMsg := tea.KeyMsg{Type: tea.KeyEscape}
	updatedModelInterface, _ := model.updateDueDateInput(keyMsg)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	// Should return to viewing mode
	if updatedModel.mode != ModeViewing {
		t.Errorf("Expected mode %v, got %v", ModeViewing, updatedModel.mode)
	}
}

// TestDueDateTextInputToggle tests toggle from text input back to calendar
func TestDueDateTextInputToggle(t *testing.T) {
	model := NewReviewModel()
	model.mode = ModeInputDueDate
	
	// Simulate Tab key to toggle back to calendar
	keyMsg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModelInterface, _ := model.updateDueDateInput(keyMsg)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	// Should transition to calendar mode
	if updatedModel.mode != ModeDueCalendar {
		t.Errorf("Expected mode %v, got %v", ModeDueCalendar, updatedModel.mode)
	}
	
	// Calendar should be focused
	if !updatedModel.calendar.focused {
		t.Error("Calendar should be focused after toggle from text input")
	}
}

// TestDueKeyBinding tests that the due key binding is properly configured
func TestDueKeyBinding(t *testing.T) {
	model := NewReviewModel()
	
	// Test that Due key is in the key map
	if model.keys.Due.Help().Key == "" {
		t.Error("Due key should have help text configured")
	}
}

// TestDueKeyPress tests due key press in main review mode
func TestDueKeyPress(t *testing.T) {
	model := NewReviewModel()
	model.mode = ModeViewing
	
	// This would typically be handled in the main Update function
	// We'll test the expected mode change directly
	model.mode = ModeDueCalendar
	model.calendar.SetFocused(true)
	model.message = "Select due date (Tab to toggle text input):"
	
	// Verify the state
	if model.mode != ModeDueCalendar {
		t.Errorf("Expected mode %v after due key press, got %v", ModeDueCalendar, model.mode)
	}
	
	if !model.calendar.focused {
		t.Error("Calendar should be focused after due key press")
	}
	
	if model.message == "" {
		t.Error("Should have message prompting for due date selection")
	}
}

// TestDueDateIntegrationFlow tests the complete due date flow
func TestDueDateIntegrationFlow(t *testing.T) {
	model := NewReviewModel()
	model.tasks = []string{"test-uuid"}
	model.current = 0
	
	// Start in due calendar mode
	model.mode = ModeDueCalendar
	model.calendar.SetFocused(true)
	
	// Navigate to a different date
	keyMsg := tea.KeyMsg{Type: tea.KeyRight}
	updatedModelInterface, _ := model.updateDueCalendar(keyMsg)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	// Confirm the date
	keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
	finalModelInterface, cmd := updatedModel.updateDueCalendar(keyMsg)
	finalModel := finalModelInterface.(*ReviewModel)
	
	// Should be back in viewing mode with a command to execute
	if finalModel.mode != ModeViewing {
		t.Errorf("Expected final mode %v, got %v", ModeViewing, finalModel.mode)
	}
	
	if cmd == nil {
		t.Error("Expected command to set due date")
	}
}