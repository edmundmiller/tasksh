package preview

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/emiller/tasksh/internal/taskwarrior"
)

// PreviewState represents different UI states to preview
type PreviewState string

const (
	StateMain           PreviewState = "main"
	StateHelp           PreviewState = "help"
	StateDelete         PreviewState = "delete"
	StateModify         PreviewState = "modify"
	StateWaitCalendar   PreviewState = "wait_calendar"
	StateDueCalendar    PreviewState = "due_calendar"
	StateContextSelect  PreviewState = "context_select"
	StateAIAnalysis     PreviewState = "ai_analysis"
	StateAILoading      PreviewState = "ai_loading"
	StatePromptAgent    PreviewState = "prompt_agent"
	StatePromptPreview  PreviewState = "prompt_preview"
	StateCelebration    PreviewState = "celebration"
	StateError          PreviewState = "error"
)

// PreviewOptions configures the preview generation
type PreviewOptions struct {
	State  PreviewState
	Width  int
	Height int
	Task   *taskwarrior.Task
}

// GeneratePreview creates a preview of the specified UI state
func GeneratePreview(opts PreviewOptions) string {
	// Set defaults
	if opts.Width == 0 {
		opts.Width = 80
	}
	if opts.Height == 0 {
		opts.Height = 24
	}

	// Create mock view
	mock := &MockView{
		Width:  opts.Width,
		Height: opts.Height,
		State:  opts.State,
	}

	// Generate appropriate state
	var output string
	switch opts.State {
	case StateMain:
		output = mock.RenderMainView()
	case StateHelp:
		output = mock.RenderHelpView()
	case StateDelete:
		output = mock.RenderDeleteConfirmation()
	case StateModify:
		output = mock.RenderModifyInput()
	case StateWaitCalendar:
		output = mock.RenderCalendar("wait")
	case StateDueCalendar:
		output = mock.RenderCalendar("due")
	case StateContextSelect:
		output = mock.RenderContextSelect()
	case StateAIAnalysis:
		output = mock.RenderAIAnalysis()
	case StateAILoading:
		output = renderAILoading(opts)
	case StatePromptAgent:
		output = renderPromptAgent(opts)
	case StatePromptPreview:
		output = renderPromptPreview(opts)
	case StateCelebration:
		output = renderCelebration(opts)
	case StateError:
		output = renderError(opts)
	default:
		output = mock.RenderMainView()
	}

	// Add preview header (ensure it doesn't exceed terminal width)
	headerText := fmt.Sprintf("=== PREVIEW: %s (%dx%d) ===", getStateDescription(opts.State), opts.Width, opts.Height)
	if len(headerText) > opts.Width {
		// Truncate the header to fit
		headerText = headerText[:opts.Width-3] + "..."
	}
	header := headerText + "\n"
	return header + output
}

// renderAILoading shows AI loading state
func renderAILoading(opts PreviewOptions) string {
	var output strings.Builder

	// Loading animation representation
	output.WriteString(strings.Repeat(" ", opts.Width/2-10))
	output.WriteString("⠋ Analyzing task with AI...\n\n")
	output.WriteString(strings.Repeat(" ", opts.Width/2-15))
	output.WriteString("Task: Implement user authentication system\n\n")
	output.WriteString(strings.Repeat(" ", opts.Width/2-12))
	output.WriteString("This may take a few seconds...\n\n")
	output.WriteString(strings.Repeat(" ", opts.Width/2-8))
	output.WriteString("Press ESC to cancel\n")

	return output.String()
}

// renderPromptAgent shows prompt agent input
func renderPromptAgent(opts PreviewOptions) string {
	var output strings.Builder

	// Mock prompt input
	output.WriteString("Prompt Agent - Tell me what to do:\n\n")
	output.WriteString("┌─────────────────────────────────────────────────────────────┐\n")
	output.WriteString("│ complete this task and log 2 hours█                        │\n")
	output.WriteString("└─────────────────────────────────────────────────────────────┘\n\n")
	output.WriteString("Tell me what you want to do with this task (e.g., 'complete this task and log 2 hours', 'start timing this task')\n")
	output.WriteString("\nPress ESC to cancel")

	return output.String()
}

// renderPromptPreview shows command preview
func renderPromptPreview(opts PreviewOptions) string {
	var output strings.Builder

	output.WriteString("🤖 Prompt Agent - Command Preview\n\n")
	output.WriteString("Commands to execute:\n")
	output.WriteString("1. Complete task: \"Implement user authentication system\"\n")
	output.WriteString("2. Log 2 hours of work time\n\n")
	output.WriteString("This will mark the task as completed and record the time spent.\n\n")
	output.WriteString("Execute these commands? (y) / (n)")

	return output.String()
}

// renderCelebration shows celebration
func renderCelebration(opts PreviewOptions) string {
	var output strings.Builder
	
	output.WriteString("\n")
	output.WriteString("    🎊  🎉  🎊  🎉  🎊\n")
	output.WriteString("  🎉      🎊      🎉\n")
	output.WriteString("      🎊      🎉\n")
	output.WriteString("\n")
	output.WriteString("   🎉 Review complete! 5 of 5 tasks reviewed. 🎉\n")
	output.WriteString("\n")
	output.WriteString("  🎉      🎊      🎉\n")
	output.WriteString("      🎊      🎉\n")
	output.WriteString("    🎊  🎉  🎊  🎉  🎊\n")

	return output.String()
}

// renderError shows error state
func renderError(opts PreviewOptions) string {
	var output strings.Builder

	output.WriteString("\n❌ Error: failed to connect to Taskwarrior: command not found\n\n")
	output.WriteString("Please ensure Taskwarrior is installed and available in your PATH.\n")
	output.WriteString("Run 'tasksh diagnostics' for more information.\n")

	return output.String()
}

// SavePreview writes preview to a file
func SavePreview(preview string, filename string) error {
	return os.WriteFile(filename, []byte(preview), 0644)
}

// GenerateAllPreviews creates previews for all UI states
func GenerateAllPreviews(outputDir string, width, height int) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Define all states to preview
	states := []struct {
		state PreviewState
		name  string
	}{
		{StateMain, "main"},
		{StateHelp, "help"},
		{StateDelete, "delete"},
		{StateModify, "modify"},
		{StateWaitCalendar, "wait_calendar"},
		{StateDueCalendar, "due_calendar"},
		{StateContextSelect, "context_select"},
		{StateAIAnalysis, "ai_analysis"},
		{StateAILoading, "ai_loading"},
		{StatePromptAgent, "prompt_agent"},
		{StatePromptPreview, "prompt_preview"},
		{StateCelebration, "celebration"},
		{StateError, "error"},
	}

	// Generate preview for each state
	for _, s := range states {
		opts := PreviewOptions{
			State:  s.state,
			Width:  width,
			Height: height,
		}
		
		preview := GeneratePreview(opts)
		filename := fmt.Sprintf("%s/%s.txt", outputDir, s.name)
		
		if err := SavePreview(preview, filename); err != nil {
			return fmt.Errorf("failed to save %s preview: %w", s.name, err)
		}
		
		fmt.Printf("Generated preview: %s\n", filename)
	}

	// Generate index file
	indexContent := generateIndex(outputDir, states)
	if err := SavePreview(indexContent, fmt.Sprintf("%s/index.txt", outputDir)); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	return nil
}

// generateIndex creates an index of all previews
func generateIndex(outputDir string, states []struct{ state PreviewState; name string }) string {
	var output strings.Builder
	
	output.WriteString("=== Tasksh UI Preview Index ===\n")
	output.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	output.WriteString("Available previews:\n\n")
	
	for _, s := range states {
		output.WriteString(fmt.Sprintf("- %s.txt: %s\n", s.name, getStateDescription(s.state)))
	}
	
	output.WriteString("\nUsage:\n")
	output.WriteString("  tasksh preview --state=main\n")
	output.WriteString("  tasksh preview --all --output-dir=./previews\n")
	
	return output.String()
}

// getStateDescription returns a human-readable description of a state
func getStateDescription(state PreviewState) string {
	descriptions := map[PreviewState]string{
		StateMain:          "Main task review interface",
		StateHelp:          "Expanded help view with all shortcuts",
		StateDelete:        "Delete task confirmation dialog",
		StateModify:        "Task modification input with completions",
		StateWaitCalendar:  "Calendar picker for wait date",
		StateDueCalendar:   "Calendar picker for due date",
		StateContextSelect: "Context selection interface",
		StateAIAnalysis:    "AI task analysis results",
		StateAILoading:     "AI analysis loading state",
		StatePromptAgent:   "Prompt agent input interface",
		StatePromptPreview: "Command preview from prompt agent",
		StateCelebration:   "Review completion celebration",
		StateError:         "Error state display",
	}
	
	if desc, ok := descriptions[state]; ok {
		return desc
	}
	return string(state)
}