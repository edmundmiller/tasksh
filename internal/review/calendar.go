package review

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CalendarModel represents a calendar picker component
type CalendarModel struct {
	// Calendar state
	currentDate    time.Time // The month/year being displayed
	selectedDate   time.Time // The currently selected date
	today          time.Time // Today's date for highlighting
	
	// Navigation
	focused        bool
	width          int
	height         int
	
	// Key bindings
	keys           CalendarKeyMap
}

// CalendarKeyMap defines key bindings for calendar navigation
type CalendarKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	NextMonth key.Binding
	PrevMonth key.Binding
	Select   key.Binding
	Cancel   key.Binding
	Today    key.Binding
}

// DefaultCalendarKeyMap returns default key bindings for the calendar
func DefaultCalendarKeyMap() CalendarKeyMap {
	return CalendarKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		NextMonth: key.NewBinding(
			key.WithKeys("n", ">"),
			key.WithHelp("n/>", "next month"),
		),
		PrevMonth: key.NewBinding(
			key.WithKeys("p", "<"),
			key.WithHelp("p/<", "prev month"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "cancel"),
		),
		Today: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "today"),
		),
	}
}

// NewCalendarModel creates a new calendar model
func NewCalendarModel() CalendarModel {
	today := time.Now()
	return CalendarModel{
		currentDate:   today,
		selectedDate:  today,
		today:         today,
		focused:       true,
		keys:          DefaultCalendarKeyMap(),
	}
}

// Init initializes the calendar model
func (m CalendarModel) Init() tea.Cmd {
	return nil
}

// Update handles calendar input and navigation
func (m CalendarModel) Update(msg tea.Msg) (CalendarModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			m.selectedDate = m.selectedDate.AddDate(0, 0, -7)
			if m.selectedDate.Month() != m.currentDate.Month() {
				m.currentDate = m.selectedDate
			}

		case key.Matches(msg, m.keys.Down):
			m.selectedDate = m.selectedDate.AddDate(0, 0, 7)
			if m.selectedDate.Month() != m.currentDate.Month() {
				m.currentDate = m.selectedDate
			}

		case key.Matches(msg, m.keys.Left):
			m.selectedDate = m.selectedDate.AddDate(0, 0, -1)
			if m.selectedDate.Month() != m.currentDate.Month() {
				m.currentDate = m.selectedDate
			}

		case key.Matches(msg, m.keys.Right):
			m.selectedDate = m.selectedDate.AddDate(0, 0, 1)
			if m.selectedDate.Month() != m.currentDate.Month() {
				m.currentDate = m.selectedDate
			}

		case key.Matches(msg, m.keys.NextMonth):
			m.currentDate = m.currentDate.AddDate(0, 1, 0)
			// Try to keep the same day of month, but adjust if not valid
			targetDay := m.selectedDate.Day()
			m.selectedDate = time.Date(m.currentDate.Year(), m.currentDate.Month(), 1, 0, 0, 0, 0, time.Local)
			daysInMonth := time.Date(m.currentDate.Year(), m.currentDate.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()
			if targetDay > daysInMonth {
				targetDay = daysInMonth
			}
			m.selectedDate = m.selectedDate.AddDate(0, 0, targetDay-1)

		case key.Matches(msg, m.keys.PrevMonth):
			m.currentDate = m.currentDate.AddDate(0, -1, 0)
			// Try to keep the same day of month, but adjust if not valid
			targetDay := m.selectedDate.Day()
			m.selectedDate = time.Date(m.currentDate.Year(), m.currentDate.Month(), 1, 0, 0, 0, 0, time.Local)
			daysInMonth := time.Date(m.currentDate.Year(), m.currentDate.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()
			if targetDay > daysInMonth {
				targetDay = daysInMonth
			}
			m.selectedDate = m.selectedDate.AddDate(0, 0, targetDay-1)

		case key.Matches(msg, m.keys.Today):
			m.selectedDate = m.today
			m.currentDate = m.today
		}
	}

	return m, nil
}

// View renders the calendar
func (m CalendarModel) View() string {
	if !m.focused {
		return ""
	}

	var b strings.Builder

	// Header with month/year and navigation hints
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")). // ANSI cyan
		Align(lipgloss.Center).
		Width(21) // 3 chars per day * 7 days

	monthYear := m.currentDate.Format("January 2006")
	b.WriteString(headerStyle.Render(monthYear))
	b.WriteString("\n\n")

	// Day headers
	dayHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")). // ANSI bright black (gray)
		Bold(true)

	dayHeaders := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	for i, header := range dayHeaders {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(dayHeaderStyle.Render(header))
	}
	b.WriteString("\n")

	// Calendar grid
	firstDay := time.Date(m.currentDate.Year(), m.currentDate.Month(), 1, 0, 0, 0, 0, time.Local)
	startOfWeek := firstDay.AddDate(0, 0, -int(firstDay.Weekday()))
	
	for week := 0; week < 6; week++ {
		for day := 0; day < 7; day++ {
			if day > 0 {
				b.WriteString(" ")
			}
			
			currentDay := startOfWeek.AddDate(0, 0, week*7+day)
			dayStr := fmt.Sprintf("%2d", currentDay.Day())
			
			// Style the day based on various states
			var style lipgloss.Style
			
			// Check if this day is in the current month
			if currentDay.Month() != m.currentDate.Month() {
				// Dim days from other months
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // ANSI bright black
			} else if isSameDay(currentDay, m.selectedDate) {
				// Highlight selected date
				style = lipgloss.NewStyle().
					Background(lipgloss.Color("4")). // ANSI blue
					Foreground(lipgloss.Color("15")). // ANSI bright white
					Bold(true)
			} else if isSameDay(currentDay, m.today) {
				// Highlight today
				style = lipgloss.NewStyle().
					Foreground(lipgloss.Color("2")). // ANSI green
					Bold(true)
			} else {
				// Normal day
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // ANSI white
			}
			
			b.WriteString(style.Render(dayStr))
		}
		b.WriteString("\n")
		
		// Stop if we've gone past the current month and we're in the last week
		if week >= 4 {
			nextWeekStart := startOfWeek.AddDate(0, 0, (week+1)*7)
			if nextWeekStart.Month() != m.currentDate.Month() {
				break
			}
		}
	}

	// Add navigation help
	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")). // ANSI bright black (gray)
		Margin(1, 0)
	
	help := "←→/h,l: day  ↑↓/k,j: week  p,</>: month  t: today  enter: select  esc: cancel"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

// GetSelectedDate returns the currently selected date
func (m CalendarModel) GetSelectedDate() time.Time {
	return m.selectedDate
}

// GetSelectedDateString returns the selected date as a string suitable for Taskwarrior
func (m CalendarModel) GetSelectedDateString() string {
	return m.selectedDate.Format("2006-01-02")
}

// SetFocused sets the focus state of the calendar
func (m *CalendarModel) SetFocused(focused bool) {
	m.focused = focused
}

// isSameDay checks if two times represent the same day
func isSameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

// ParseDateInput attempts to parse various date input formats
func ParseDateInput(input string) (time.Time, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	now := time.Now()
	
	// Handle relative dates
	switch input {
	case "today":
		return now, nil
	case "tomorrow":
		return now.AddDate(0, 0, 1), nil
	case "yesterday":
		return now.AddDate(0, 0, -1), nil
	case "next week":
		return now.AddDate(0, 0, 7), nil
	case "next month":
		return now.AddDate(0, 1, 0), nil
	}
	
	// Handle day names (next occurrence)
	dayNames := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
		"sun":       time.Sunday,
		"mon":       time.Monday,
		"tue":       time.Tuesday,
		"wed":       time.Wednesday,
		"thu":       time.Thursday,
		"fri":       time.Friday,
		"sat":       time.Saturday,
	}
	
	if targetDay, exists := dayNames[input]; exists {
		daysUntil := int(targetDay - now.Weekday())
		if daysUntil <= 0 {
			daysUntil += 7 // Next week
		}
		return now.AddDate(0, 0, daysUntil), nil
	}
	
	// Handle "in X days"
	if strings.HasPrefix(input, "in ") && strings.HasSuffix(input, " days") {
		daysStr := strings.TrimPrefix(input, "in ")
		daysStr = strings.TrimSuffix(daysStr, " days")
		if days, err := strconv.Atoi(daysStr); err == nil {
			return now.AddDate(0, 0, days), nil
		}
	}
	
	// Handle "+X" format
	if strings.HasPrefix(input, "+") {
		if days, err := strconv.Atoi(input[1:]); err == nil {
			return now.AddDate(0, 0, days), nil
		}
	}
	
	// Try standard date formats
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"01-02-2006",
		"2006/01/02",
		"Jan 2, 2006",
		"January 2, 2006",
		"2 Jan 2006",
		"2 January 2006",
	}
	
	for _, format := range formats {
		if parsed, err := time.Parse(format, input); err == nil {
			return parsed, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("could not parse date: %s", input)
}