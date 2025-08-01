package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/emiller/tasksh/internal/planning"
	"github.com/emiller/tasksh/internal/taskwarrior"
)

func main() {
	// Measure startup time for planning view
	fmt.Println("Planning View Startup Performance Test")
	fmt.Println("=====================================")

	// Time the entire startup process
	totalStart := time.Now()

	// Time session creation
	sessionStart := time.Now()
	session := &planning.PlanningSession{
		Date:       time.Now().AddDate(0, 0, 1),
		TotalHours: 0,
	}

	// Add realistic number of tasks
	for i := 0; i < 50; i++ {
		task := planning.PlannedTask{
			Task: &taskwarrior.Task{
				UUID:        fmt.Sprintf("task-%d", i),
				Description: fmt.Sprintf("Task %d with a moderately long description that represents typical task content", i),
				Priority:    []string{"H", "M", "L"}[i%3],
			},
			EstimatedHours:  float64(1 + i%4),
			EnergyLevel:     planning.EnergyLevel(i % 3),
			IsDue:           i%4 == 0,
			IsScheduled:     i%3 == 0,
			OptimalTimeSlot: []string{"morning", "afternoon", "evening", "anytime"}[i%4],
		}

		if i < 15 {
			task.Category = planning.CategoryCritical
			session.CriticalTasks = append(session.CriticalTasks, task)
		} else if i < 30 {
			task.Category = planning.CategoryImportant
			session.ImportantTasks = append(session.ImportantTasks, task)
		} else {
			task.Category = planning.CategoryFlexible
			session.FlexibleTasks = append(session.FlexibleTasks, task)
		}

		session.TotalHours += task.EstimatedHours
	}
	sessionTime := time.Since(sessionStart)

	// Time model creation
	modelStart := time.Now()
	model := planning.NewPlanningModel(session)
	modelTime := time.Since(modelStart)

	// Time initial window size setup
	setupStart := time.Now()
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	model.Update(msg)
	setupTime := time.Since(setupStart)

	// Time first render
	renderStart := time.Now()
	firstRender := model.View()
	firstRenderTime := time.Since(renderStart)

	// Time subsequent renders
	subsequentStart := time.Now()
	for i := 0; i < 10; i++ {
		_ = model.View()
	}
	subsequentAvg := time.Since(subsequentStart) / 10

	totalTime := time.Since(totalStart)

	// Report results
	fmt.Printf("\nStartup Timing Breakdown:\n")
	fmt.Printf("- Session creation: %v\n", sessionTime)
	fmt.Printf("- Model creation: %v\n", modelTime)
	fmt.Printf("- Window setup: %v\n", setupTime)
	fmt.Printf("- First render: %v\n", firstRenderTime)
	fmt.Printf("- Subsequent renders (avg): %v\n", subsequentAvg)
	fmt.Printf("\nTotal startup time: %v\n", totalTime)
	
	// Count lines in first render to check complexity
	lines := 1
	for _, ch := range firstRender {
		if ch == '\n' {
			lines++
		}
	}
	fmt.Printf("\nFirst render stats:\n")
	fmt.Printf("- Total characters: %d\n", len(firstRender))
	fmt.Printf("- Total lines: %d\n", lines)

	// Check for potential issues
	if firstRenderTime > 50*time.Millisecond {
		fmt.Printf("\n⚠️  WARNING: First render is slow (>50ms)\n")
	}
	if subsequentAvg > 10*time.Millisecond {
		fmt.Printf("⚠️  WARNING: Subsequent renders are slow (>10ms)\n")
	}
}