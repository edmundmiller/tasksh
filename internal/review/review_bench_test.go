// +build skip

package review

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/emiller/tasksh/internal/taskwarrior"
)

// BenchmarkReviewModelView benchmarks the View rendering performance
func BenchmarkReviewModelView(b *testing.B) {
	tasks := createBenchmarkTasks(20)
	model := createTestModel(tasks)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

// BenchmarkReviewModelUpdate benchmarks Update with different message types
func BenchmarkReviewModelUpdate(b *testing.B) {
	tasks := createBenchmarkTasks(20)
	model := createTestModel(tasks)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	
	messages := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyDown},
		tea.KeyMsg{Type: tea.KeyUp},
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, // Toggle help
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}, // Mark reviewed
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}, // Complete
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.Update(messages[i%len(messages)])
	}
}

// BenchmarkRenderTask benchmarks task rendering with different complexities
func BenchmarkRenderTask(b *testing.B) {
	testCases := []struct {
		name string
		task *taskwarrior.Task
	}{
		{
			name: "Simple",
			task: &taskwarrior.Task{
				UUID:        "simple-1",
				Description: "Simple task",
			},
		},
		{
			name: "WithTags",
			task: &taskwarrior.Task{
				UUID:        "tags-1",
				Description: "Task with tags",
				Tags:        []string{"work", "urgent", "backend", "review"},
			},
		},
		{
			name: "Complex",
			task: &taskwarrior.Task{
				UUID:        "complex-1",
				Description: "Complex task with very long description that includes multiple details and requirements",
				Project:     "big-project",
				Tags:        []string{"work", "urgent", "backend", "review", "testing"},
				Due:         "2024-12-25",
				Priority:    "H",
			},
		},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			model := createTestModel([]*taskwarrior.Task{tc.task})
			model.width = 80
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = model.renderTask(tc.task)
			}
		})
	}
}

// BenchmarkTransitionTo benchmarks state transitions
func BenchmarkTransitionTo(b *testing.B) {
	transitions := []struct {
		name  string
		state reviewState
	}{
		{"ToHelp", stateHelp},
		{"ToModifying", stateModifying},
		{"ToDeleting", stateDeleting},
		{"ToCompleting", stateCompleting},
		{"ToWaiting", stateWaiting},
		{"ToMain", stateMain},
	}
	
	tasks := createBenchmarkTasks(10)
	
	for _, tr := range transitions {
		b.Run(tr.name, func(b *testing.B) {
			model := createTestModel(tasks)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				model.transitionTo(tr.state)
			}
		})
	}
}

// BenchmarkHandleReviewAction benchmarks different review actions
func BenchmarkHandleReviewAction(b *testing.B) {
	actions := []string{"r", "c", "e", "m", "d", "s"}
	tasks := createBenchmarkTasks(20)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model := createTestModel(tasks)
		action := actions[i%len(actions)]
		model.handleReviewAction(action)
	}
}

// BenchmarkRenderHelp benchmarks help rendering in different states
func BenchmarkRenderHelp(b *testing.B) {
	states := []reviewState{
		stateMain,
		stateHelp,
		stateModifying,
		stateCompleting,
	}
	
	tasks := createBenchmarkTasks(5)
	
	for _, state := range states {
		b.Run(string(state), func(b *testing.B) {
			model := createTestModel(tasks)
			model.state = state
			model.width = 80
			model.height = 24
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = model.renderHelp()
			}
		})
	}
}

// Helper functions

func createBenchmarkTasks(count int) []*taskwarrior.Task {
	tasks := make([]*taskwarrior.Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = &taskwarrior.Task{
			UUID:        fmt.Sprintf("bench-%d", i),
			Description: fmt.Sprintf("Benchmark task %d with moderate description length", i),
			Project:     fmt.Sprintf("project-%d", i%3),
			Tags:        []string{"bench", fmt.Sprintf("tag%d", i%5)},
			Priority:    []string{"H", "M", "L", ""}[i%4],
			Due:         "",
			Entry:       time.Now().AddDate(0, 0, -i),
		}
		
		if i%3 == 0 {
			tasks[i].Due = time.Now().AddDate(0, 0, i).Format("2006-01-02")
		}
	}
	return tasks
}

func createTestModel(tasks []*taskwarrior.Task) *ReviewModel {
	return &ReviewModel{
		tasks:        tasks,
		currentIndex: 0,
		state:        stateMain,
		width:        80,
		height:       24,
	}
}