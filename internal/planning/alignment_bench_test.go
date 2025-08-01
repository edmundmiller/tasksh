package planning

import (
	"fmt"
	"strings"
	"testing"
	"time"
	
	"github.com/emiller/tasksh/internal/taskwarrior"
)

// BenchmarkVisualWidthPerf benchmarks the visual width calculation performance
func BenchmarkVisualWidthPerf(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"Simple", "Hello World"},
		{"WithANSI", "\x1b[36mHello\x1b[0m World"},
		{"Unicode", "┃ Test ┃"},
		{"Complex", "┃  ▶ 1  Signal Quality Control Plots with very long description that needs truncation"},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = visualWidth(tc.input)
			}
		})
	}
}

// BenchmarkStripANSIPerf benchmarks the ANSI stripping function performance
func BenchmarkStripANSIPerf(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"NoANSI", "Hello World"},
		{"SingleANSI", "\x1b[36mHello\x1b[0m"},
		{"MultipleANSI", "\x1b[36mHello\x1b[0m \x1b[1m\x1b[35mWorld\x1b[0m"},
		{"LongWithANSI", strings.Repeat("\x1b[36mTest\x1b[0m ", 20)},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = stripANSI(tc.input)
			}
		})
	}
}

// BenchmarkRenderSection benchmarks the entire section rendering
func BenchmarkRenderSection(b *testing.B) {
	// Create a mock planning model with many tasks
	session := &PlanningSession{
		Date:       time.Now().AddDate(0, 0, 1),
		TotalHours: 0,
	}
	
	// Add 20 critical tasks
	for i := 0; i < 20; i++ {
		task := PlannedTask{
			Task: &taskwarrior.Task{
				UUID:        fmt.Sprintf("task-%d", i),
				Description: fmt.Sprintf("Task number %d with a very long description that might need truncation in the UI", i),
				Priority:    "H",
			},
			EstimatedHours:  2.5,
			EnergyLevel:     EnergyHigh,
			IsDue:           i%3 == 0,
			IsScheduled:     i%2 == 0,
			OptimalTimeSlot: "morning",
			Category:        CategoryCritical,
		}
		session.CriticalTasks = append(session.CriticalTasks, task)
		session.TotalHours += task.EstimatedHours
	}
	
	model := NewPlanningModel(session)
	model.width = 80
	model.height = 24
	
	// Initialize model
	model.Update(nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

// BenchmarkTruncateToWidth benchmarks the truncation function
func BenchmarkTruncateToWidth(b *testing.B) {
	testCases := []struct {
		name  string
		input string
		width int
	}{
		{"NoTruncation", "Short text", 20},
		{"SimpleTruncation", "This is a very long text that needs to be truncated", 20},
		{"UnicodeTruncation", "This is a text with emojis 🎉🎊 and special chars", 20},
		{"ANSITruncation", "\x1b[36mThis is a colored text that needs truncation\x1b[0m", 20},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = truncateToWidth(tc.input, tc.width)
			}
		})
	}
}