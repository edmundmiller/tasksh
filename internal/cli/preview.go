package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/emiller/tasksh/internal/preview"
)

// RunPreview handles the preview command
func RunPreview(args []string) error {
	// Create flag set for preview command
	flags := flag.NewFlagSet("preview", flag.ExitOnError)
	
	// Define flags
	state := flags.String("state", "main", "UI state to preview")
	all := flags.Bool("all", false, "Generate all UI state previews")
	outputDir := flags.String("output-dir", "./previews", "Directory to save previews")
	width := flags.Int("width", 80, "Terminal width")
	height := flags.Int("height", 24, "Terminal height")
	output := flags.String("output", "", "Output file (stdout if not specified)")
	list := flags.Bool("list", false, "List available states")
	
	// Parse flags
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	
	// Handle list flag
	if *list {
		listStates()
		return nil
	}
	
	// Handle all flag
	if *all {
		if err := preview.GenerateAllPreviews(*outputDir, *width, *height); err != nil {
			return fmt.Errorf("failed to generate all previews: %w", err)
		}
		fmt.Printf("\nAll previews generated in %s/\n", *outputDir)
		return nil
	}
	
	// Generate single preview
	opts := preview.PreviewOptions{
		State:  preview.PreviewState(*state),
		Width:  *width,
		Height: *height,
	}
	
	previewOutput := preview.GeneratePreview(opts)
	
	// Output to file or stdout
	if *output != "" {
		if err := os.WriteFile(*output, []byte(previewOutput), 0644); err != nil {
			return fmt.Errorf("failed to save preview: %w", err)
		}
		fmt.Printf("Preview saved to %s\n", *output)
	} else {
		fmt.Print(previewOutput)
	}
	
	return nil
}

// listStates prints all available preview states
func listStates() {
	fmt.Println("Available preview states:")
	fmt.Println()
	
	states := []struct {
		name        string
		description string
	}{
		{"main", "Main task review interface"},
		{"help", "Expanded help view with all shortcuts"},
		{"delete", "Delete task confirmation dialog"},
		{"modify", "Task modification input with completions"},
		{"wait_calendar", "Calendar picker for wait date"},
		{"due_calendar", "Calendar picker for due date"},
		{"context_select", "Context selection interface"},
		{"ai_analysis", "AI task analysis results"},
		{"ai_loading", "AI analysis loading state"},
		{"prompt_agent", "Prompt agent input interface"},
		{"prompt_preview", "Command preview from prompt agent"},
		{"celebration", "Review completion celebration"},
		{"error", "Error state display"},
	}
	
	for _, s := range states {
		fmt.Printf("  %-15s - %s\n", s.name, s.description)
	}
	
	fmt.Println()
	fmt.Println("Usage examples:")
	fmt.Println("  tasksh preview --state=main")
	fmt.Println("  tasksh preview --state=help --width=120 --height=30")
	fmt.Println("  tasksh preview --all --output-dir=./ui-previews")
}

// PreviewHelp returns help text for the preview command
func PreviewHelp() string {
	return `tasksh preview - Generate UI state previews

Usage:
  tasksh preview [options]

Options:
  --state=STATE      UI state to preview (default: main)
  --all              Generate all UI state previews
  --output-dir=DIR   Directory to save previews (default: ./previews)
  --width=N          Terminal width (default: 80)
  --height=N         Terminal height (default: 24)
  --output=FILE      Output file (stdout if not specified)
  --list             List available states

Examples:
  # Preview main interface
  tasksh preview

  # Preview help screen with custom size
  tasksh preview --state=help --width=120 --height=30

  # Generate all previews
  tasksh preview --all

  # Save preview to file
  tasksh preview --state=modify --output=modify-preview.txt

This command helps visualize UI changes without running the full application.
Use it to iterate on design improvements and document UI states.`
}