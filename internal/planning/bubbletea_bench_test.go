package planning

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/emiller/tasksh/internal/taskwarrior"
)

// BenchmarkPlanningModelView benchmarks the View rendering performance
func BenchmarkPlanningModelView(b *testing.B) {
	// Create a realistic planning session
	session := createBenchmarkSession(30)
	model := NewPlanningModel(session)
	
	// Initialize with window size
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

// BenchmarkPlanningModelUpdate benchmarks the Update performance
func BenchmarkPlanningModelUpdate(b *testing.B) {
	session := createBenchmarkSession(30)
	model := NewPlanningModel(session)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	
	// Create different types of messages to benchmark
	messages := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyDown},
		tea.KeyMsg{Type: tea.KeyUp},
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}, // Toggle projection
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}}, // Move up
		tea.WindowSizeMsg{Width: 100, Height: 30},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := messages[i%len(messages)]
		model.Update(msg)
	}
}

// BenchmarkRenderSectionBubbleTea benchmarks section rendering with different task counts
func BenchmarkRenderSectionBubbleTea(b *testing.B) {
	testCases := []struct {
		name      string
		taskCount int
	}{
		{"5Tasks", 5},
		{"10Tasks", 10},
		{"20Tasks", 20},
		{"50Tasks", 50},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			session := createBenchmarkSession(tc.taskCount)
			model := NewPlanningModel(session)
			model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			
			completionTimes := session.GetProjectedCompletionTimes(model.workStartTime)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = model.renderSection("CRITICAL TASKS", session.CriticalTasks, 0, completionTimes, lipgloss.Color("1"))
			}
		})
	}
}

// BenchmarkGetContentWidth benchmarks the content width calculation with caching
func BenchmarkGetContentWidth(b *testing.B) {
	model := &PlanningModel{
		width: 80,
		contentWidthCache: contentWidthCache{},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.getContentWidth()
	}
}

// BenchmarkCachedVisualWidth benchmarks the cached visual width function
func BenchmarkCachedVisualWidth(b *testing.B) {
	testStrings := []string{
		"Simple text",
		"\x1b[36mColored text\x1b[0m",
		"┃  ▶ 1  Very long task description that might need to be cached",
		"High priority • High energy • Due today • Scheduled • Best in morning",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cachedVisualWidth(testStrings[i%len(testStrings)])
	}
}

// BenchmarkUpdateViewport benchmarks viewport updates with different task counts
func BenchmarkUpdateViewport(b *testing.B) {
	testCases := []struct {
		name      string
		taskCount int
	}{
		{"Empty", 0},
		{"10Tasks", 10},
		{"30Tasks", 30},
		{"100Tasks", 100},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			session := createBenchmarkSession(tc.taskCount)
			model := NewPlanningModel(session)
			model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				model.updateViewport()
			}
		})
	}
}

// Helper function to create a benchmark session with specified number of tasks
func createBenchmarkSession(taskCount int) *PlanningSession {
	session := &PlanningSession{
		Horizon:       HorizonTomorrow,
		Date:         time.Now().AddDate(0, 0, 1),
		DailyCapacity: 8.0,
		FocusCapacity: 6.0,
		MaxTasks:      taskCount,
		TotalHours:    0,
	}
	
	// Distribute tasks across categories
	for i := 0; i < taskCount; i++ {
		task := PlannedTask{
			Task: &taskwarrior.Task{
				UUID:        fmt.Sprintf("bench-task-%d", i),
				Description: fmt.Sprintf("Benchmark task %d with a moderately long description for testing", i),
				Priority:    []string{"H", "M", "L"}[i%3],
			},
			EstimatedHours:  float64(1 + i%4),
			EnergyLevel:     EnergyLevel(i % 3),
			IsDue:           i%4 == 0,
			IsScheduled:     i%3 == 0,
			OptimalTimeSlot: []string{"morning", "afternoon", "evening", "anytime"}[i%4],
		}
		
		if i < taskCount/3 {
			task.Category = CategoryCritical
			session.CriticalTasks = append(session.CriticalTasks, task)
		} else if i < 2*taskCount/3 {
			task.Category = CategoryImportant
			session.ImportantTasks = append(session.ImportantTasks, task)
		} else {
			task.Category = CategoryFlexible
			session.FlexibleTasks = append(session.FlexibleTasks, task)
		}
		
		session.TotalHours += task.EstimatedHours
	}
	
	// Build combined task list
	session.Tasks = make([]PlannedTask, 0, len(session.CriticalTasks)+len(session.ImportantTasks)+len(session.FlexibleTasks))
	session.Tasks = append(session.Tasks, session.CriticalTasks...)
	session.Tasks = append(session.Tasks, session.ImportantTasks...)
	session.Tasks = append(session.Tasks, session.FlexibleTasks...)
	
	return session
}