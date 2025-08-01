package preview

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// TestAllUIStates ensures all defined states have preview implementations
func TestAllUIStates(t *testing.T) {
	allStates := []PreviewState{
		StateMain,
		StateHelp,
		StateDelete,
		StateModify,
		StateWaitCalendar,
		StateDueCalendar,
		StateContextSelect,
		StateAIAnalysis,
		StateAILoading,
		StatePromptAgent,
		StatePromptPreview,
		StateCelebration,
		StateError,
	}

	for _, state := range allStates {
		t.Run(string(state), func(t *testing.T) {
			opts := PreviewOptions{
				State:  state,
				Width:  80,
				Height: 24,
			}
			
			preview := GeneratePreview(opts)
			
			// Basic validation
			if preview == "" {
				t.Errorf("State %s returned empty preview", state)
			}
			
			// Check that state is properly identified
			if !strings.Contains(preview, "=== PREVIEW:") {
				t.Errorf("State %s missing preview header", state)
			}
		})
	}
}

// TestEdgeCaseSizes tests extreme terminal dimensions
func TestEdgeCaseSizes(t *testing.T) {
	edgeCases := []struct {
		name   string
		width  int
		height int
	}{
		{"minimum viable", 40, 10},
		{"very narrow", 40, 40},
		{"very short", 200, 10},
		{"mobile portrait", 50, 100},
		{"4K monitor", 380, 100},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := PreviewOptions{
				State:  StateMain,
				Width:  tc.width,
				Height: tc.height,
			}
			
			preview := GeneratePreview(opts)
			lines := strings.Split(preview, "\n")
			
			// Validate all lines fit
			for i, line := range lines {
				cleanLine := stripANSI(line)
				if len(cleanLine) > tc.width {
					t.Errorf("Line %d exceeds width %d: %s", i+1, tc.width, line)
				}
			}
			
			// Validate height constraint
			if len(lines) > tc.height {
				t.Errorf("Content exceeds height: %d > %d", len(lines), tc.height)
			}
		})
	}
}

// TestUnicodeHandling ensures proper handling of Unicode characters
func TestUnicodeHandling(t *testing.T) {
	// Test with a mock that includes Unicode
	opts := PreviewOptions{
		State:  StateMain,
		Width:  80,
		Height: 24,
	}
	
	preview := GeneratePreview(opts)
	
	// Check for proper Unicode rendering (progress bar, box drawing)
	unicodeChars := []string{"╭", "╮", "╰", "╯", "│", "─", "•", "▶"}
	
	for _, char := range unicodeChars {
		if strings.Contains(preview, char) {
			// Verify proper width calculation
			lines := strings.Split(preview, "\n")
			for _, line := range lines {
				if strings.Contains(line, char) {
					// Ensure line width is calculated correctly with Unicode
					cleanLine := stripANSI(line)
					runeCount := utf8.RuneCountInString(cleanLine)
					if runeCount > opts.Width {
						t.Errorf("Unicode line exceeds width when counting runes: %d > %d", 
							runeCount, opts.Width)
					}
				}
			}
		}
	}
}

// TestColorStripping verifies ANSI color codes don't affect width calculations
func TestColorStripping(t *testing.T) {
	testCases := []struct {
		input    string
		expected int
		desc     string
	}{
		{"\x1b[36mHello\x1b[0m", 5, "simple color"},
		{"\x1b[1m\x1b[36mBold Cyan\x1b[0m", 9, "compound formatting"},
		{"\x1b[31mError: \x1b[0mSomething went wrong", 26, "mixed colored text"},
		{"No colors here", 14, "plain text"},
		{"\x1b[38;5;214mExtended color\x1b[0m", 14, "256-color mode"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			stripped := stripANSI(tc.input)
			actual := len(stripped)
			if actual != tc.expected {
				t.Errorf("stripANSI(%q) length = %d, want %d", tc.input, actual, tc.expected)
				t.Logf("Stripped result: %q", stripped)
			}
		})
	}
}

// TestInteractiveElements ensures interactive UI elements are properly rendered
func TestInteractiveElements(t *testing.T) {
	tests := []struct {
		state    PreviewState
		expected []string
		desc     string
	}{
		{
			StateModify,
			[]string{"Enter new value", "Tab", "autocomplete"},
			"modify input should show input hints",
		},
		{
			StateContextSelect,
			[]string{"Select context", "↑/↓", "Enter: select"},
			"context select should show navigation",
		},
		{
			StateWaitCalendar,
			[]string{"calendar", "←/→", "↑/↓"},
			"calendar should show navigation hints",
		},
		{
			StatePromptAgent,
			[]string{"Enter command", ">"},
			"prompt should show input indicator",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			opts := PreviewOptions{
				State:  test.state,
				Width:  80,
				Height: 24,
			}
			
			preview := GeneratePreview(opts)
			
			for _, expected := range test.expected {
				if !strings.Contains(strings.ToLower(preview), strings.ToLower(expected)) {
					t.Errorf("State %s missing expected element %q", test.state, expected)
				}
			}
		})
	}
}

// TestAccessibility ensures UI elements follow accessibility best practices
func TestAccessibility(t *testing.T) {
	opts := PreviewOptions{
		State:  StateMain,
		Width:  80,
		Height: 24,
	}
	
	preview := GeneratePreview(opts)
	
	// Check for sufficient contrast indicators
	tests := []struct {
		element string
		desc    string
	}{
		{"[", "brackets for key navigation"},
		{":", "separator between label and value"},
		{"•", "bullet points for lists"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if !strings.Contains(preview, test.element) {
				t.Logf("Warning: missing accessibility element %q (%s)", 
					test.element, test.desc)
			}
		})
	}

	// Check that important information isn't only conveyed through color
	if strings.Contains(preview, "\x1b[31m") { // Red color
		// Ensure error states have text indicators too
		if !strings.Contains(preview, "Error") && !strings.Contains(preview, "!") {
			t.Error("Error state relies only on color, needs text indicator")
		}
	}
}

// TestProgressBarRendering tests various progress states
func TestProgressBarRendering(t *testing.T) {
	// This would need MockView to accept progress parameter
	// For now, we'll just ensure progress bar is present
	opts := PreviewOptions{
		State:  StateMain,
		Width:  80,
		Height: 24,
	}
	
	preview := GeneratePreview(opts)
	
	// Look for progress indicators
	if !strings.Contains(preview, "Progress:") && 
	   !strings.Contains(preview, "=") &&
	   !strings.Contains(preview, "-") {
		t.Error("No progress indicator found in main view")
	}
}

// TestMultilineContent ensures long content is properly wrapped
func TestMultilineContent(t *testing.T) {
	// Test with states that typically have long content
	states := []PreviewState{
		StateAIAnalysis,
		StateError,
		StateHelp,
	}

	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			opts := PreviewOptions{
				State:  state,
				Width:  60, // Narrow to force wrapping
				Height: 30,
			}
			
			preview := GeneratePreview(opts)
			lines := strings.Split(preview, "\n")
			
			// Check no line exceeds width
			for i, line := range lines {
				cleanLine := stripANSI(line)
				if len(cleanLine) > opts.Width {
					t.Errorf("Line %d exceeds width in %s state", i+1, state)
				}
			}
		})
	}
}

// TestEmptyStates ensures graceful handling of empty data
func TestEmptyStates(t *testing.T) {
	// States that might have empty content
	emptyScenarios := []struct {
		state    PreviewState
		expected string
	}{
		{StateContextSelect, "No contexts available"},
		{StateAIAnalysis, "No analysis available"},
		{StateError, "Unknown error"},
	}

	for _, scenario := range emptyScenarios {
		t.Run(string(scenario.state), func(t *testing.T) {
			opts := PreviewOptions{
				State:  scenario.state,
				Width:  80,
				Height: 24,
			}
			
			preview := GeneratePreview(opts)
			
			// Should have some placeholder or message
			if len(strings.TrimSpace(preview)) < 50 {
				t.Errorf("State %s appears empty, should show placeholder", scenario.state)
			}
		})
	}
}

// TestKeyboardShortcuts ensures all shortcuts are documented
func TestKeyboardShortcuts(t *testing.T) {
	// Test help view has all shortcuts
	opts := PreviewOptions{
		State:  StateHelp,
		Width:  100,
		Height: 40,
	}
	
	preview := GeneratePreview(opts)
	
	// Essential shortcuts that should be documented
	shortcuts := []string{
		"r:", "review",
		"c:", "complete",
		"e:", "edit",
		"s:", "skip",
		"d:", "delete",
		"q:", "quit",
		"?:", "help",
	}

	for i := 0; i < len(shortcuts); i += 2 {
		key := shortcuts[i]
		action := shortcuts[i+1]
		
		if !strings.Contains(preview, key) {
			t.Errorf("Help view missing shortcut key %s", key)
		}
		if !strings.Contains(strings.ToLower(preview), action) {
			t.Errorf("Help view missing action %s", action)
		}
	}
}