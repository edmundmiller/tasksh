package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("tasksh - Interactive task review shell")
		fmt.Println("Usage:")
		fmt.Println("  tasksh review [limit]  - Start task review")
		fmt.Println("  tasksh help            - Show help")
		fmt.Println("  tasksh diagnostics     - Show diagnostics")
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
		if err := cmdReview(limit); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help":
		cmdHelp()
	case "diagnostics":
		cmdDiagnostics()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
		os.Exit(1)
	}
}

func cmdHelp() {
	fmt.Println("tasksh - Interactive task review shell")
	fmt.Println()
	fmt.Println("The review process helps keep your task list accurate by allowing you to")
	fmt.Println("systematically review tasks and update their metadata.")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  review [N]     Review tasks (optionally limit to N tasks)")
	fmt.Println("  help           Show this help")
	fmt.Println("  diagnostics    Show system diagnostics")
	fmt.Println()
	fmt.Println("During review, you can:")
	fmt.Println("  - Edit task (opens task editor)")
	fmt.Println("  - Modify task (enter modification arguments)")
	fmt.Println("  - Complete task")
	fmt.Println("  - Delete task")
	fmt.Println("  - Skip task (will need review again later)")
	fmt.Println("  - Mark as reviewed")
	fmt.Println("  - Quit review session")
}

func cmdDiagnostics() {
	fmt.Println("tasksh diagnostics")
	fmt.Println()
	fmt.Printf("Version: %s\n", "2.0.0-go")
	fmt.Printf("Built with: Go\n")
	fmt.Println()
	
	// Check if task command is available
	if err := checkTaskwarrior(); err != nil {
		fmt.Printf("Taskwarrior: NOT FOUND - %v\n", err)
	} else {
		fmt.Println("Taskwarrior: Available")
	}
}