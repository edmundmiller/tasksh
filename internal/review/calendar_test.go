package review

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestCalendarNavigation tests basic calendar navigation functionality
func TestCalendarNavigation(t *testing.T) {
	cal := NewCalendarModel()
	cal.SetFocused(true)
	
	// Get initial date
	initialDate := cal.GetSelectedDate()
	
	// Test right arrow key (next day)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	cal, _ = cal.Update(keyMsg)
	
	expectedDate := initialDate.AddDate(0, 0, 1)
	if !isSameDay(cal.GetSelectedDate(), expectedDate) {
		t.Errorf("Right navigation failed. Expected %v, got %v", expectedDate, cal.GetSelectedDate())
	}
	
	// Test left arrow key (previous day)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	cal, _ = cal.Update(keyMsg)
	
	if !isSameDay(cal.GetSelectedDate(), initialDate) {
		t.Errorf("Left navigation failed. Expected %v, got %v", initialDate, cal.GetSelectedDate())
	}
	
	// Test down arrow key (next week)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	cal, _ = cal.Update(keyMsg)
	
	expectedDate = initialDate.AddDate(0, 0, 7)
	if !isSameDay(cal.GetSelectedDate(), expectedDate) {
		t.Errorf("Down navigation failed. Expected %v, got %v", expectedDate, cal.GetSelectedDate())
	}
	
	// Test up arrow key (previous week)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	cal, _ = cal.Update(keyMsg)
	
	if !isSameDay(cal.GetSelectedDate(), initialDate) {
		t.Errorf("Up navigation failed. Expected %v, got %v", initialDate, cal.GetSelectedDate())
	}
}

// TestCalendarArrowKeys tests arrow key navigation
func TestCalendarArrowKeys(t *testing.T) {
	cal := NewCalendarModel()
	cal.SetFocused(true)
	
	initialDate := cal.GetSelectedDate()
	
	testCases := []struct {
		name     string
		key      tea.KeyType
		expected time.Time
	}{
		{"right arrow", tea.KeyRight, initialDate.AddDate(0, 0, 1)},
		{"left arrow", tea.KeyLeft, initialDate},
		{"down arrow", tea.KeyDown, initialDate.AddDate(0, 0, 7)},
		{"up arrow", tea.KeyUp, initialDate},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyMsg := tea.KeyMsg{Type: tc.key}
			cal, _ = cal.Update(keyMsg)
			
			if !isSameDay(cal.GetSelectedDate(), tc.expected) {
				t.Errorf("%s failed. Expected %v, got %v", tc.name, tc.expected, cal.GetSelectedDate())
			}
		})
	}
}

// TestCalendarMonthNavigation tests month navigation
func TestCalendarMonthNavigation(t *testing.T) {
	cal := NewCalendarModel()
	cal.SetFocused(true)
	
	// Set to a specific date to test month boundaries
	testDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.Local)
	cal.selectedDate = testDate
	cal.currentDate = testDate
	
	// Test next month
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	cal, _ = cal.Update(keyMsg)
	
	expectedMonth := time.Month(7)
	if cal.GetSelectedDate().Month() != expectedMonth {
		t.Errorf("Next month navigation failed. Expected month %v, got %v", expectedMonth, cal.GetSelectedDate().Month())
	}
	
	// Test previous month
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	cal, _ = cal.Update(keyMsg)
	
	expectedMonth = time.Month(6)
	if cal.GetSelectedDate().Month() != expectedMonth {
		t.Errorf("Previous month navigation failed. Expected month %v, got %v", expectedMonth, cal.GetSelectedDate().Month())
	}
}

// TestCalendarTodayKey tests the today key functionality
func TestCalendarTodayKey(t *testing.T) {
	cal := NewCalendarModel()
	cal.SetFocused(true)
	
	// Move to a different date
	futureDate := time.Now().AddDate(0, 1, 5)
	cal.selectedDate = futureDate
	cal.currentDate = futureDate
	
	// Press 't' for today
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	cal, _ = cal.Update(keyMsg)
	
	if !isSameDay(cal.GetSelectedDate(), cal.today) {
		t.Errorf("Today key failed. Expected %v, got %v", cal.today, cal.GetSelectedDate())
	}
}

// TestCalendarFocusState tests that unfocused calendar doesn't respond to keys
func TestCalendarFocusState(t *testing.T) {
	cal := NewCalendarModel()
	cal.SetFocused(false)
	
	initialDate := cal.GetSelectedDate()
	
	// Try to navigate while unfocused
	keyMsg := tea.KeyMsg{Type: tea.KeyRight}
	cal, _ = cal.Update(keyMsg)
	
	// Date should not have changed
	if !isSameDay(cal.GetSelectedDate(), initialDate) {
		t.Errorf("Unfocused calendar should not respond to keys. Expected %v, got %v", initialDate, cal.GetSelectedDate())
	}
}

// TestCalendarDateString tests date string formatting
func TestCalendarDateString(t *testing.T) {
	cal := NewCalendarModel()
	
	testDate := time.Date(2024, 12, 25, 0, 0, 0, 0, time.Local)
	cal.selectedDate = testDate
	
	expected := "2024-12-25"
	actual := cal.GetSelectedDateString()
	
	if actual != expected {
		t.Errorf("Date string formatting failed. Expected %s, got %s", expected, actual)
	}
}

// TestCalendarViewRenders tests that the calendar view renders without panic
func TestCalendarViewRenders(t *testing.T) {
	cal := NewCalendarModel()
	cal.SetFocused(true)
	
	// Simulate window size
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	cal, _ = cal.Update(sizeMsg)
	
	view := cal.View()
	if view == "" {
		t.Error("Calendar view should not be empty when focused")
	}
	
	// Unfocused calendar should return empty view
	cal.SetFocused(false)
	view = cal.View()
	if view != "" {
		t.Error("Calendar view should be empty when unfocused")
	}
}

// TestCalendarMonthBoundaries tests navigation across month boundaries
func TestCalendarMonthBoundaries(t *testing.T) {
	cal := NewCalendarModel()
	cal.SetFocused(true)
	
	// Set to last day of month
	testDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.Local)
	cal.selectedDate = testDate
	cal.currentDate = testDate
	
	// Navigate to next day (should go to next month)
	keyMsg := tea.KeyMsg{Type: tea.KeyRight}
	cal, _ = cal.Update(keyMsg)
	
	expectedDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.Local)
	if !isSameDay(cal.GetSelectedDate(), expectedDate) {
		t.Errorf("Month boundary navigation failed. Expected %v, got %v", expectedDate, cal.GetSelectedDate())
	}
	
	// Current month should have updated too
	if cal.currentDate.Month() != time.February {
		t.Errorf("Current month should have updated. Expected February, got %v", cal.currentDate.Month())
	}
}

// TestReviewModelCalendarIntegration tests calendar integration in review model
func TestReviewModelCalendarIntegration(t *testing.T) {
	model := NewReviewModel()
	model.mode = ModeWaitCalendar
	model.calendar.SetFocused(true)
	
	initialDate := model.calendar.GetSelectedDate()
	
	// Test that calendar navigation works through updateWaitCalendar
	keyMsg := tea.KeyMsg{Type: tea.KeyRight}
	updatedModelInterface, _ := model.updateWaitCalendar(keyMsg)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	expectedDate := initialDate.AddDate(0, 0, 1)
	if !isSameDay(updatedModel.calendar.GetSelectedDate(), expectedDate) {
		t.Errorf("Calendar navigation through review model failed. Expected %v, got %v", 
			expectedDate, updatedModel.calendar.GetSelectedDate())
	}
}

// TestReviewModelCalendarConfirm tests calendar confirmation flow
func TestReviewModelCalendarConfirm(t *testing.T) {
	model := NewReviewModel()
	model.mode = ModeWaitCalendar
	model.calendar.SetFocused(true)
	
	// Set a specific date
	testDate := time.Date(2024, 12, 25, 0, 0, 0, 0, time.Local)
	model.calendar.selectedDate = testDate
	
	// Simulate Enter key
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModelInterface, _ := model.updateWaitCalendar(keyMsg)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	// Should transition to wait reason input mode
	if updatedModel.mode != ModeInputWaitReason {
		t.Errorf("Expected mode %v, got %v", ModeInputWaitReason, updatedModel.mode)
	}
	
	// Wait date should be set
	expected := "2024-12-25"
	if updatedModel.waitDate != expected {
		t.Errorf("Expected waitDate %s, got %s", expected, updatedModel.waitDate)
	}
}

// TestReviewModelCalendarToggle tests toggle between calendar and text input
func TestReviewModelCalendarToggle(t *testing.T) {
	model := NewReviewModel()
	model.mode = ModeWaitCalendar
	model.calendar.SetFocused(true)
	
	// Simulate Tab key to toggle to text input
	keyMsg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModelInterface, _ := model.updateWaitCalendar(keyMsg)
	updatedModel := updatedModelInterface.(*ReviewModel)
	
	// Should transition to text input mode
	if updatedModel.mode != ModeInputWaitDate {
		t.Errorf("Expected mode %v, got %v", ModeInputWaitDate, updatedModel.mode)
	}
	
	// Calendar should be unfocused
	if updatedModel.calendar.focused {
		t.Error("Calendar should be unfocused after toggle")
	}
}

// TestReviewModelCalendarCancel tests calendar cancellation
func TestReviewModelCalendarCancel(t *testing.T) {
	model := NewReviewModel()
	model.mode = ModeWaitCalendar
	model.calendar.SetFocused(true)
	
	// Simulate Escape key
	keyMsg := tea.KeyMsg{Type: tea.KeyEscape}
	updatedModelInterface, _ := model.updateWaitCalendar(keyMsg)
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