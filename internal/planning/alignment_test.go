package planning

import (
	"strings"
	"testing"
)

func TestVisualWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Simple ASCII", "Hello World", 11},
		{"With ANSI codes", "\x1b[36mHello\x1b[0m World", 11},
		{"Unicode box drawing", "┃ Test ┃", 8},
		{"Mixed content", "┃  ▶ 1  Signal Quality", 22},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := visualWidth(tt.input)
			if got != tt.expected {
				t.Errorf("visualWidth(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLineAlignment(t *testing.T) {
	// Simulate the planning view line construction
	contentWidth := 80
	
	// Test task line alignment
	prefix := "┃  ▶ 1  "
	description := "Signal Quality Control Plots"
	timeInfo := "3.0h  12:00 PM"
	
	prefixWidth := visualWidth(prefix)
	timeInfoWidth := visualWidth(timeInfo)
	rightBorderWidth := 3 // "  ┃"
	
	availableForDesc := contentWidth - prefixWidth - timeInfoWidth - rightBorderWidth - 2
	
	descWidth := visualWidth(description)
	dotsNeeded := availableForDesc - descWidth
	
	line := prefix + description + " " + strings.Repeat(".", dotsNeeded) + " " + timeInfo + "  ┃"
	
	// Check that the line is exactly the content width
	if visualWidth(line) != contentWidth {
		t.Errorf("Line width = %d, want %d", visualWidth(line), contentWidth)
		t.Logf("Line: %q", line)
	}
}

func TestMetadataAlignment(t *testing.T) {
	contentWidth := 80
	
	metaText := "High priority • High energy • Due today • Scheduled • Best in morning"
	metaPrefix := "┃       "
	
	metaPrefixWidth := visualWidth(metaPrefix)
	metaTextWidth := visualWidth(metaText)
	rightBorderWidth := 1
	
	paddingNeeded := contentWidth - metaPrefixWidth - metaTextWidth - rightBorderWidth
	
	metaLine := metaPrefix + metaText + strings.Repeat(" ", paddingNeeded) + "┃"
	
	if visualWidth(metaLine) != contentWidth {
		t.Errorf("Metadata line width = %d, want %d", visualWidth(metaLine), contentWidth)
		t.Logf("Line: %q", metaLine)
	}
}