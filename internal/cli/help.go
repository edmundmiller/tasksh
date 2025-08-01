package cli

import (
	"fmt"

	"github.com/emiller/tasksh/internal/ai"
)

func ShowHelp() {
	fmt.Println("tasksh - Interactive task management shell")
	fmt.Println()
	fmt.Println("Tasksh provides two distinct planning experiences:")
	fmt.Println()
	fmt.Println("📅 DAILY PLANNING - Guided 5-step execution planning")
	fmt.Println("  • Reflect on yesterday's work and lessons learned")
	fmt.Println("  • Select tasks based on energy and capacity")
	fmt.Println("  • Assess realistic workload with time projections")
	fmt.Println("  • Set daily focus and intentions")
	fmt.Println("  • Create achievable daily plan")
	fmt.Println()
	fmt.Println("📊 WEEKLY PLANNING - Strategic objective-setting")
	fmt.Println("  • Reflect on previous week's accomplishments")
	fmt.Println("  • Set 2-3 key objectives for the week")
	fmt.Println("  • Strategic journaling and big-picture thinking")
	fmt.Println("  • Organize work into thematic streams")
	fmt.Println("  • Align weekly goals with daily execution")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  plan daily         Guided daily planning (recommended)")
	fmt.Println("  plan weekly        Strategic weekly planning (recommended)")
	fmt.Println("  plan today         Legacy: plan today's tasks")
	fmt.Println("  plan tomorrow      Legacy: plan tomorrow's tasks")
	fmt.Println("  plan week          Legacy: plan upcoming week")
	fmt.Println("  review [N]         Review tasks (optionally limit to N tasks)")
	fmt.Println("  preview            Preview UI states for design iteration")
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