package cli

import "fmt"

func ShowHelp() {
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
	fmt.Println("  - Modify task (with smart completion for projects/tags/priorities)")
	fmt.Println("  - AI Analysis (get OpenAI-powered suggestions for improvements)")
	fmt.Println("  - Complete task (with optional time tracking)")
	fmt.Println("  - Delete task")
	fmt.Println("  - Wait task (set waiting status with date and reason)")
	fmt.Println("  - Due date (set or modify task due date)")
	fmt.Println("  - Skip task (will need review again later)")
	fmt.Println("  - Mark as reviewed")
	fmt.Println("  - Quit review session")
}