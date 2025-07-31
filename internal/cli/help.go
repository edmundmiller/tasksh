package cli

import (
	"fmt"

	"github.com/emiller/tasksh/internal/ai"
)

func ShowHelp() {
	fmt.Println("tasksh - Interactive task management shell")
	fmt.Println()
	fmt.Println("Daily planning helps you organize and prioritize tasks with capacity")
	fmt.Println("management and realistic time estimates. The review process helps keep")
	fmt.Println("your task list accurate by systematically reviewing and updating metadata.")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  plan tomorrow      Plan tomorrow's tasks with time estimates")
	fmt.Println("  plan week          Plan upcoming week's tasks")
	fmt.Println("  review [N]         Review tasks (optionally limit to N tasks)")
	fmt.Println("  help               Show this help")
	fmt.Println("  diagnostics        Show system diagnostics")
	fmt.Println()
	fmt.Println("Planning Features:")
	fmt.Println("  - Smart task selection based on urgency and due dates")
	fmt.Println("  - Time estimation using historical data")
	fmt.Println("  - Capacity warnings to prevent overcommitment")
	fmt.Println("  - Interactive playlist reordering")
	fmt.Println("  - Time projection showing completion estimates")
	fmt.Println()
	fmt.Println("During review, you can:")
	fmt.Println("  - Edit task (opens task editor)")
	fmt.Println("  - Modify task (with smart completion for projects/tags/priorities)")
	
	// Only show AI features if available
	if ai.CheckOpenAIAvailable() == nil {
		fmt.Println("  - AI Analysis (get OpenAI-powered suggestions for improvements)")
		fmt.Println("  - Prompt Agent (tell AI what to do with natural language)")
	}
	
	fmt.Println("  - Complete task (with optional time tracking)")
	fmt.Println("  - Delete task")
	fmt.Println("  - Wait task (set waiting status with date and reason)")
	fmt.Println("  - Due date (set or modify task due date)")
	fmt.Println("  - Skip task (will need review again later)")
	fmt.Println("  - Mark as reviewed")
	fmt.Println("  - Quit review session")
}