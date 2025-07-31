package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/emiller/tasksh/internal/cli"
	"github.com/emiller/tasksh/internal/planning"
	"github.com/emiller/tasksh/internal/review"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("tasksh - Interactive task management shell")
		fmt.Println("Usage:")
		fmt.Println("  tasksh review [limit]     - Start task review")
		fmt.Println("  tasksh plan tomorrow      - Plan tomorrow's tasks")
		fmt.Println("  tasksh plan week          - Plan upcoming week")
		fmt.Println("  tasksh plan quick         - Quick planning (3 critical tasks)")
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
			fmt.Println("Usage: tasksh plan <tomorrow|week>")
			os.Exit(1)
		}
		
		var horizon planning.PlanningHorizon
		switch args[1] {
		case "tomorrow":
			horizon = planning.HorizonTomorrow
		case "week":
			horizon = planning.HorizonWeek
		case "quick":
			horizon = planning.HorizonQuick
		default:
			fmt.Fprintf(os.Stderr, "Unknown planning horizon: %s\n", args[1])
			fmt.Println("Available horizons: tomorrow, week, quick")
			os.Exit(1)
		}
		
		if err := planning.Run(horizon); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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