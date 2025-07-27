package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/emiller/tasksh/internal/cli"
	"github.com/emiller/tasksh/internal/review"
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
		if err := review.Run(limit); err != nil {
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