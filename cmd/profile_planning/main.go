package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/emiller/tasksh/internal/planning"
	"github.com/emiller/tasksh/internal/taskwarrior"
)

func main() {
	// Create CPU profile
	f, err := os.Create("cpu.prof")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	
	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()
	
	// Create a mock planning session
	session := &planning.PlanningSession{
		Date: time.Now().AddDate(0, 0, 1),
		CriticalTasks: []planning.PlannedTask{},
		ImportantTasks: []planning.PlannedTask{},
		FlexibleTasks: []planning.PlannedTask{},
		TotalHours: 0,
	}
	
	// Add many tasks to stress test
	for i := 0; i < 50; i++ {
		task := planning.PlannedTask{
			Task: &taskwarrior.Task{
				UUID:        fmt.Sprintf("task-%d", i),
				Description: fmt.Sprintf("Task number %d with a very long description that might need truncation in the UI", i),
				Priority:    "H",
			},
			EstimatedHours:  2.5,
			EnergyLevel:     planning.EnergyHigh,
			IsDue:           i%3 == 0,
			IsScheduled:     i%2 == 0,
			OptimalTimeSlot: "morning",
		}
		
		if i < 20 {
			session.CriticalTasks = append(session.CriticalTasks, task)
		} else if i < 35 {
			session.ImportantTasks = append(session.ImportantTasks, task)
		} else {
			session.FlexibleTasks = append(session.FlexibleTasks, task)
		}
		
		session.TotalHours += task.EstimatedHours
	}
	
	// Create planning model
	model := planning.NewPlanningModel(session)
	
	// Simulate many renders
	start := time.Now()
	iterations := 1000
	
	for i := 0; i < iterations; i++ {
		// Simulate window resize by sending tea.WindowSizeMsg
		if i%100 == 0 {
			width := 80 + (i/100)*10
			height := 24
			msg := tea.WindowSizeMsg{Width: width, Height: height}
			model.Update(msg)
		}
		
		// Render view
		_ = model.View()
	}
	
	elapsed := time.Since(start)
	fmt.Printf("Rendered %d times in %v\n", iterations, elapsed)
	fmt.Printf("Average time per render: %v\n", elapsed/time.Duration(iterations))
	
	// Calculate renders per second
	rendersPerSecond := float64(iterations) / elapsed.Seconds()
	fmt.Printf("Renders per second: %.2f\n", rendersPerSecond)
}