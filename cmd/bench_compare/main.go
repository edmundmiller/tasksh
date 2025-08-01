package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/emiller/tasksh/internal/planning"
	"github.com/emiller/tasksh/internal/taskwarrior"
)

func main() {
	// Create a realistic planning session
	session := &planning.PlanningSession{
		Date:       time.Now().AddDate(0, 0, 1),
		TotalHours: 0,
	}

	// Add tasks similar to real usage
	taskDescriptions := []string{
		"Implement user authentication system with OAuth2 support",
		"Write comprehensive unit tests for the API endpoints",
		"Refactor database connection pooling for better performance",
		"Update documentation for the new features",
		"Fix bug in the payment processing module",
		"Optimize image upload functionality",
		"Review and merge pending pull requests",
		"Set up continuous integration pipeline",
		"Conduct security audit of the application",
		"Implement caching layer for frequently accessed data",
	}

	// Create tasks distributed across categories
	for i := 0; i < 30; i++ {
		desc := taskDescriptions[i%len(taskDescriptions)] + fmt.Sprintf(" (Task %d)", i)
		task := planning.PlannedTask{
			Task: &taskwarrior.Task{
				UUID:        fmt.Sprintf("task-%d", i),
				Description: desc,
				Priority:    []string{"H", "M", "L"}[i%3],
			},
			EstimatedHours:  float64(1 + i%4),
			EnergyLevel:     planning.EnergyLevel(i % 3),
			IsDue:           i%4 == 0,
			IsScheduled:     i%3 == 0,
			OptimalTimeSlot: []string{"morning", "afternoon", "evening", "anytime"}[i%4],
		}

		// Distribute tasks across categories
		if i < 10 {
			task.Category = planning.CategoryCritical
			session.CriticalTasks = append(session.CriticalTasks, task)
		} else if i < 20 {
			task.Category = planning.CategoryImportant
			session.ImportantTasks = append(session.ImportantTasks, task)
		} else {
			task.Category = planning.CategoryFlexible
			session.FlexibleTasks = append(session.FlexibleTasks, task)
		}

		session.TotalHours += task.EstimatedHours
	}

	// Create and initialize model
	model := planning.NewPlanningModel(session)
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	model.Update(msg)

	// Warm up
	for i := 0; i < 10; i++ {
		_ = model.View()
	}

	// Measure rendering performance
	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		// Simulate some interaction
		if i%50 == 0 {
			// Change selected task
			model.Update(tea.KeyMsg{Type: tea.KeyDown})
		}
		if i%100 == 0 {
			// Toggle projection view
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
		}

		_ = model.View()
	}

	elapsed := time.Since(start)
	avgTime := elapsed / time.Duration(iterations)
	rendersPerSecond := float64(iterations) / elapsed.Seconds()

	fmt.Printf("Planning View Performance Test\n")
	fmt.Printf("==============================\n")
	fmt.Printf("Tasks: %d (Critical: %d, Important: %d, Flexible: %d)\n",
		len(session.CriticalTasks)+len(session.ImportantTasks)+len(session.FlexibleTasks),
		len(session.CriticalTasks), len(session.ImportantTasks), len(session.FlexibleTasks))
	fmt.Printf("Iterations: %d\n", iterations)
	fmt.Printf("Total time: %v\n", elapsed)
	fmt.Printf("Average time per render: %v\n", avgTime)
	fmt.Printf("Renders per second: %.2f\n", rendersPerSecond)

	// Check if performance is acceptable
	if avgTime > 10*time.Millisecond {
		fmt.Printf("\n⚠️  WARNING: Rendering is slow (>10ms per frame)\n")
	} else if avgTime > 5*time.Millisecond {
		fmt.Printf("\n⚡ Performance is acceptable but could be improved\n")
	} else {
		fmt.Printf("\n✅ Excellent performance!\n")
	}
}