package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/emiller/tasksh/internal/cli"
	"github.com/emiller/tasksh/internal/planning"
	"github.com/emiller/tasksh/internal/planning/daily"
	"github.com/emiller/tasksh/internal/planning/weekly"
	"github.com/emiller/tasksh/internal/review"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("tasksh - Interactive task management shell")
		fmt.Println("Usage:")
		fmt.Println("  tasksh review [limit]     - Start task review")
		fmt.Println("  tasksh plan daily         - Guided daily planning (5 steps)")
		fmt.Println("  tasksh plan weekly        - Strategic weekly planning")
		fmt.Println("  tasksh plan today         - Legacy: plan today's tasks")
		fmt.Println("  tasksh plan tomorrow      - Legacy: plan tomorrow's tasks")
		fmt.Println("  tasksh plan week          - Legacy: plan upcoming week")
		fmt.Println("  tasksh plan quick         - Legacy: quick planning")
		fmt.Println("  tasksh preview            - Preview UI states")
		fmt.Println("  tasksh help               - Show help")
		fmt.Println("  tasksh diagnostics        - Show diagnostics")
		os.Exit(0)
	}

	switch args[0] {
	case "review":
		limit := 0
		if len(args) > 1 {
			if l, err := strconv.Atoi(args[1]); err == nil {
				limit = l
			}
		}
		if err := review.Run(limit); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "plan":
		if len(args) < 2 {
			fmt.Println("Usage: tasksh plan <daily|weekly|today|tomorrow|week|quick>")
			fmt.Println("  daily    - New guided daily planning workflow")
			fmt.Println("  weekly   - New strategic weekly planning workflow")
			fmt.Println("  today    - Legacy daily planning")
			fmt.Println("  tomorrow - Legacy planning for tomorrow")
			fmt.Println("  week     - Legacy weekly planning")
			fmt.Println("  quick    - Legacy quick planning")
			os.Exit(1)
		}
		
		switch args[1] {
		case "daily":
			// New guided daily planning for today
			targetDate := time.Now()
			if err := daily.Run(targetDate); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		case "weekly":
			// New strategic weekly planning
			now := time.Now()
			// Find the start of this week (Monday)
			weekday := now.Weekday()
			if weekday == 0 { // Sunday
				weekday = 7
			}
			daysFromMonday := int(weekday) - 1
			weekStart := now.AddDate(0, 0, -daysFromMonday)
			weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
			
			if err := weekly.Run(weekStart); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		
		// Legacy planning modes for backward compatibility
		case "today", "tomorrow", "week", "quick":
			var horizon planning.PlanningHorizon
			switch args[1] {
			case "today":
				horizon = planning.HorizonToday
			case "tomorrow":
				horizon = planning.HorizonTomorrow
			case "week":
				horizon = planning.HorizonWeek
			case "quick":
				horizon = planning.HorizonQuick
			}
			
			if err := planning.Run(horizon); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "Unknown planning mode: %s\n", args[1])
			fmt.Println("Available modes: daily, weekly, today, tomorrow, week, quick")
			os.Exit(1)
		}
	case "preview":
		if err := cli.RunPreview(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help":
		cli.ShowHelp()
	case "diagnostics":
		cli.ShowDiagnostics()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
		os.Exit(1)
	}
}