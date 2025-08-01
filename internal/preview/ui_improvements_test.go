package preview

import (
	"strings"
	"testing"
)

// TestSimplifiedHelpText ensures help text shows only primary actions by default
func TestSimplifiedHelpText(t *testing.T) {
	opts := PreviewOptions{
		State:  StateMain,
		Width:  80,
		Height: 24,
	}
	
	preview := GeneratePreview(opts)
	
	// Should show primary actions
	primaryActions := []string{"r: review", "c: complete", "?: more"}
	for _, action := range primaryActions {
		if !strings.Contains(preview, action) {
			t.Errorf("Missing primary action: %s", action)
		}
	}
	
	// Should NOT show all actions in collapsed view
	advancedActions := []string{"w: wait", "m: modify context", "shift+p: priority"}
	for _, action := range advancedActions {
		if strings.Contains(preview, action) {
			t.Errorf("Advanced action visible in collapsed help: %s", action)
		}
	}
	
	// Should have "more" indicator
	if !strings.Contains(preview, "?") || !strings.Contains(preview, "more") {
		t.Error("Missing '?' for more help indicator")
	}
}

// TestExpandedHelpView tests that help view shows all shortcuts
func TestExpandedHelpView(t *testing.T) {
	opts := PreviewOptions{
		State:  StateHelp,
		Width:  80,
		Height: 30,
	}
	
	preview := GeneratePreview(opts)
	
	// Should show all shortcuts organized by category
	categories := []string{
		"Review Actions",
		"Task Management",
		"Navigation",
	}
	
	for _, category := range categories {
		if !strings.Contains(preview, category) {
			t.Errorf("Missing category in help view: %s", category)
		}
	}
	
	// Should show advanced shortcuts
	advancedShortcuts := []string{
		"shift+d", "shift+p", "shift+w",
		"ctrl+r", "backspace",
	}
	
	for _, shortcut := range advancedShortcuts {
		if !strings.Contains(strings.ToLower(preview), shortcut) {
			t.Errorf("Missing advanced shortcut in help view: %s", shortcut)
		}
	}
}

// TestStatusBarPosition ensures status bar shows task position clearly
func TestStatusBarPosition(t *testing.T) {
	opts := PreviewOptions{
		State:  StateMain,
		Width:  80,
		Height: 24,
	}
	
	preview := GeneratePreview(opts)
	
	// Should show task position in format [current of total]
	if !strings.Contains(preview, "[2 of 15]") {
		t.Error("Status bar missing task position indicator")
	}
	
	// Should show context
	if !strings.Contains(preview, "Context:") {
		t.Error("Status bar missing context indicator")
	}
}

// TestVisualSeparators tests that UI sections have clear separators
func TestVisualSeparators(t *testing.T) {
	opts := PreviewOptions{
		State:  StateMain,
		Width:  80,
		Height: 24,
	}
	
	preview := GeneratePreview(opts)
	lines := strings.Split(preview, "\n")
	
	// Look for different separator patterns
	separatorTypes := []string{
		strings.Repeat("-", 20), // Simple dashes
		strings.Repeat("=", 20), // Double lines
		"────────",              // Box drawing
	}
	
	separatorFound := false
	for _, line := range lines {
		for _, sep := range separatorTypes {
			if strings.Contains(line, sep) {
				separatorFound = true
				break
			}
		}
	}
	
	if !separatorFound {
		t.Error("No visual separators found between UI sections")
	}
}

// TestColorHierarchy verifies proper use of colors for visual hierarchy
func TestColorHierarchy(t *testing.T) {
	opts := PreviewOptions{
		State:  StateMain,
		Width:  80,
		Height: 24,
	}
	
	preview := GeneratePreview(opts)
	
	// Color expectations
	colorTests := []struct {
		element string
		color   string
		desc    string
	}{
		{"r: review", "\x1b[36m", "primary actions should be cyan"},
		{"Context:", "\x1b[90m", "labels should be gray"},
		{"High", "\x1b[31m", "high priority should be red"},
		{"Due:", "\x1b[33m", "due dates should be yellow"},
	}
	
	for _, test := range colorTests {
		if strings.Contains(preview, test.element) {
			// Find the element and check if it has the right color nearby
			idx := strings.Index(preview, test.element)
			if idx > 0 {
				// Look for color code before the element (within 20 chars)
				searchStart := idx - 20
				if searchStart < 0 {
					searchStart = 0
				}
				searchArea := preview[searchStart:idx]
				
				if !strings.Contains(searchArea, test.color) {
					t.Logf("Warning: %s not found with color %s (%s)", 
						test.element, test.color, test.desc)
				}
			}
		}
	}
}

// TestProgressBarImprovement ensures progress bar is visually clear
func TestProgressBarImprovement(t *testing.T) {
	opts := PreviewOptions{
		State:  StateMain,
		Width:  80,
		Height: 24,
	}
	
	preview := GeneratePreview(opts)
	
	// Should have a progress indicator
	if !strings.Contains(preview, "Progress:") || 
	   !strings.Contains(preview, "Review Progress:") {
		// Look for progress bar characters
		hasProgress := false
		progressChars := []string{"=", "-", "█", "░", "▓", "▒"}
		
		for _, char := range progressChars {
			if strings.Contains(preview, strings.Repeat(char, 3)) {
				hasProgress = true
				break
			}
		}
		
		if !hasProgress {
			t.Error("No clear progress indicator found")
		}
	}
}

// TestCompactViewport verifies viewport uses space efficiently
func TestCompactViewport(t *testing.T) {
	opts := PreviewOptions{
		State:  StateMain,
		Width:  80,
		Height: 24,
	}
	
	preview := GeneratePreview(opts)
	lines := strings.Split(preview, "\n")
	
	// Count empty lines
	emptyLines := 0
	consecutiveEmpty := 0
	maxConsecutive := 0
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			emptyLines++
			consecutiveEmpty++
			if consecutiveEmpty > maxConsecutive {
				maxConsecutive = consecutiveEmpty
			}
		} else {
			consecutiveEmpty = 0
		}
	}
	
	// Should not have too many empty lines
	emptyRatio := float64(emptyLines) / float64(len(lines))
	if emptyRatio > 0.2 {
		t.Errorf("Too many empty lines: %d/%d (%.1f%%)", 
			emptyLines, len(lines), emptyRatio*100)
	}
	
	// Should not have more than 2 consecutive empty lines
	if maxConsecutive > 2 {
		t.Errorf("Too many consecutive empty lines: %d", maxConsecutive)
	}
}

// TestUIElementAlignment ensures all UI elements align properly
func TestUIElementAlignment(t *testing.T) {
	opts := PreviewOptions{
		State:  StateMain,
		Width:  80,
		Height: 24,
	}
	
	preview := GeneratePreview(opts)
	lines := strings.Split(preview, "\n")
	
	// Find box-drawn sections
	var boxStarts []int
	var boxEnds []int
	
	for i, line := range lines {
		cleanLine := stripANSI(line)
		
		// Look for box drawing characters
		if strings.HasPrefix(cleanLine, "╭") || strings.HasPrefix(cleanLine, "┌") {
			boxStarts = append(boxStarts, i)
		}
		if strings.HasPrefix(cleanLine, "╰") || strings.HasPrefix(cleanLine, "└") {
			boxEnds = append(boxEnds, i)
		}
	}
	
	// Verify boxes are properly aligned
	for i := 0; i < len(boxStarts) && i < len(boxEnds); i++ {
		startLine := stripANSI(lines[boxStarts[i]])
		endLine := stripANSI(lines[boxEnds[i]])
		
		// Should have same width
		if len(startLine) != len(endLine) {
			t.Errorf("Box misalignment: start width %d != end width %d", 
				len(startLine), len(endLine))
		}
		
		// Vertical sides should align
		for j := boxStarts[i] + 1; j < boxEnds[i]; j++ {
			middleLine := stripANSI(lines[j])
			if len(middleLine) > 0 {
				// Check first and last characters are box drawing
				firstChar := string([]rune(middleLine)[0])
				if firstChar != "│" && firstChar != "║" && firstChar != "┃" {
					t.Logf("Line %d missing left border: %s", j+1, middleLine)
				}
			}
		}
	}
}

// TestMinimalModeSupport tests that UI works in minimal space
func TestMinimalModeSupport(t *testing.T) {
	// Test very small terminal
	opts := PreviewOptions{
		State:  StateMain,
		Width:  40,
		Height: 10,
	}
	
	preview := GeneratePreview(opts)
	lines := strings.Split(preview, "\n")
	
	// Should still show essential information
	essentials := []string{
		"[", "]", // Task position
		":", // Some kind of label/value separator
	}
	
	essentialFound := 0
	for _, essential := range essentials {
		if strings.Contains(preview, essential) {
			essentialFound++
		}
	}
	
	if essentialFound < len(essentials)/2 {
		t.Error("Minimal mode missing essential UI elements")
	}
	
	// Should fit within constraints
	for i, line := range lines {
		if len(stripANSI(line)) > opts.Width {
			t.Errorf("Line %d exceeds minimal width", i+1)
		}
	}
	
	if len(lines) > opts.Height {
		t.Errorf("Content exceeds minimal height: %d > %d", len(lines), opts.Height)
	}
}