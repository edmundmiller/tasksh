package planning

import (
	"fmt"
	"strings"
	"testing"
)

// TestStripANSIComplex tests ANSI stripping with complex sequences
func TestStripANSIComplex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"Multiple sequences",
			"\x1b[1m\x1b[36mBold Cyan\x1b[0m \x1b[31mRed\x1b[0m",
			"Bold Cyan Red",
		},
		{
			"256 color mode",
			"\x1b[38;5;214mOrange\x1b[0m text",
			"Orange text",
		},
		{
			"RGB color",
			"\x1b[38;2;255;128;0mRGB Orange\x1b[0m",
			"RGB Orange",
		},
		{
			"Background colors",
			"\x1b[41mRed Background\x1b[0m",
			"Red Background",
		},
		{
			"Cursor movement",
			"Text\x1b[2A\x1b[3DMore",
			"Text\x1b[2A\x1b[3DMore", // These aren't color codes, shouldn't strip
		},
		{
			"Nested sequences",
			"\x1b[1m\x1b[31m\x1b[4mBold Red Underline\x1b[0m",
			"Bold Red Underline",
		},
		{
			"Incomplete sequence",
			"Text\x1b[31",
			"Text\x1b[31", // Incomplete, shouldn't strip
		},
		{
			"Empty sequence",
			"\x1b[mEmpty\x1b[0m",
			"Empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSI(tt.input)
			if result != tt.expected {
				t.Errorf("stripANSI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestVisualWidthEdgeCases tests visual width with edge cases
func TestVisualWidthEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		// Empty and whitespace
		{"Empty string", "", 0},
		{"Spaces only", "   ", 3},
		{"Tabs", "\t\t", 2},
		{"Mixed whitespace", " \t \n ", 5},
		
		// Unicode edge cases
		{"Emoji", "👍 Hello 🎉", 9},
		{"Chinese characters", "你好世界", 4},
		{"Arabic", "مرحبا", 5},
		{"Zero-width joiner", "👨‍👩‍👧‍👦", 1}, // Family emoji
		{"Combining marks", "é", 1}, // e + combining acute
		
		// Box drawing variations
		{"Single line box", "┌─┐│└┘", 6},
		{"Double line box", "╔═╗║╚╝", 6},
		{"Mixed box", "╭─╮│╰╯", 6},
		{"Thick box", "┏━┓┃┗┛", 6},
		
		// ANSI in middle of Unicode
		{"Colored emoji", "\x1b[31m🔴\x1b[0m Red", 6},
		{"Colored Chinese", "\x1b[36m你好\x1b[0m", 2},
		
		// Very long strings
		{"Long ASCII", strings.Repeat("a", 1000), 1000},
		{"Long Unicode", strings.Repeat("你", 500), 500},
		{"Long with ANSI", strings.Repeat("\x1b[31ma\x1b[0m", 100), 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := visualWidth(tt.input)
			if result != tt.expected {
				t.Errorf("visualWidth(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestPadToWidth tests padding functionality
func TestPadToWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		padChar  string
		expected string
	}{
		{"Simple padding", "Hello", 10, " ", "Hello     "},
		{"Dot padding", "Test", 10, ".", "Test......"},
		{"Already wide enough", "LongString", 5, " ", "LongString"},
		{"Unicode padding", "你好", 6, "·", "你好····"},
		{"ANSI in input", "\x1b[31mRed\x1b[0m", 5, " ", "\x1b[31mRed\x1b[0m  "},
		{"Zero width", "Text", 0, " ", "Text"},
		{"Unicode pad char", "Hi", 5, "→", "Hi→→→"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padToWidth(tt.input, tt.width, tt.padChar)
			if result != tt.expected {
				t.Errorf("padToWidth(%q, %d, %q) = %q, want %q", 
					tt.input, tt.width, tt.padChar, result, tt.expected)
			}
			
			// Verify visual width
			resultWidth := visualWidth(result)
			expectedWidth := tt.width
			if tt.width > 0 && resultWidth < expectedWidth && visualWidth(tt.input) < tt.width {
				t.Errorf("Result width %d is less than requested %d", resultWidth, expectedWidth)
			}
		})
	}
}

// TestTruncateToWidth tests truncation functionality
func TestTruncateToWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		expected string
	}{
		{"Simple truncate", "Hello World", 8, "Hello..."},
		{"Already fits", "Short", 10, "Short"},
		{"Unicode truncate", "你好世界你好", 5, "你好..."},
		{"ANSI truncate", "\x1b[31mRed Text Here\x1b[0m", 7, "Red ..."},
		{"Exact fit", "12345", 5, "12345"},
		{"Very short", "Hello World", 4, "H..."},
		{"Zero width", "Text", 0, "Text"}, // Should not truncate
		{"Width 3", "Text", 3, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateToWidth(tt.input, tt.maxWidth)
			if result != tt.expected {
				t.Errorf("truncateToWidth(%q, %d) = %q, want %q", 
					tt.input, tt.maxWidth, result, tt.expected)
			}
			
			// Verify result width doesn't exceed max
			if tt.maxWidth > 0 {
				resultWidth := visualWidth(result)
				if resultWidth > tt.maxWidth {
					t.Errorf("Result width %d exceeds max %d", resultWidth, tt.maxWidth)
				}
			}
		})
	}
}

// TestAlignmentInPractice tests real-world alignment scenarios
func TestAlignmentInPractice(t *testing.T) {
	// Test the actual planning view alignment logic
	tests := []struct {
		name          string
		contentWidth  int
		title         string
		hours         float64
		expectedWidth int
	}{
		{"Standard width", 80, "CRITICAL TASKS", 9.0, 80},
		{"Wide terminal", 120, "IMPORTANT TASKS", 15.5, 120},
		{"Narrow terminal", 60, "FLEXIBLE", 3.25, 60},
		{"Very long title", 80, "VERY LONG SECTION TITLE THAT MIGHT OVERFLOW", 99.99, 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the header creation from renderSection
			headerText := fmt.Sprintf("┏━ %s ", tt.title)
			hoursText := fmt.Sprintf(" %.1fh ━━━┓", tt.hours)
			
			headerWidth := visualWidth(headerText)
			hoursWidth := visualWidth(hoursText)
			padding := tt.contentWidth - headerWidth - hoursWidth
			
			if padding < 1 {
				padding = 1
			}
			
			header := headerText + strings.Repeat("━", padding) + hoursText
			actualWidth := visualWidth(header)
			
			if actualWidth != tt.expectedWidth {
				t.Errorf("Header width = %d, want %d", actualWidth, tt.expectedWidth)
				t.Logf("Header: %q", header)
				t.Logf("Parts: header=%d, hours=%d, padding=%d", 
					headerWidth, hoursWidth, padding)
			}
		})
	}
}

// TestBoxDrawingAlignment tests alignment with various box drawing characters
func TestBoxDrawingAlignment(t *testing.T) {
	width := 40
	
	patterns := []struct {
		name   string
		top    string
		middle string
		bottom string
		chars  struct{ tl, tr, bl, br, h, v string }
	}{
		{
			"Single line",
			"", "", "",
			struct{ tl, tr, bl, br, h, v string }{"┌", "┐", "└", "┘", "─", "│"},
		},
		{
			"Double line",
			"", "", "",
			struct{ tl, tr, bl, br, h, v string }{"╔", "╗", "╚", "╝", "═", "║"},
		},
		{
			"Rounded",
			"", "", "",
			struct{ tl, tr, bl, br, h, v string }{"╭", "╮", "╰", "╯", "─", "│"},
		},
	}

	for _, p := range patterns {
		t.Run(p.name, func(t *testing.T) {
			c := p.chars
			
			// Create box lines
			top := c.tl + strings.Repeat(c.h, width-2) + c.tr
			middle := c.v + strings.Repeat(" ", width-2) + c.v
			bottom := c.bl + strings.Repeat(c.h, width-2) + c.br
			
			// Check all lines have same visual width
			topWidth := visualWidth(top)
			middleWidth := visualWidth(middle)
			bottomWidth := visualWidth(bottom)
			
			if topWidth != width || middleWidth != width || bottomWidth != width {
				t.Errorf("Box lines have inconsistent widths: top=%d, middle=%d, bottom=%d (want %d)",
					topWidth, middleWidth, bottomWidth, width)
			}
		})
	}
}

// BenchmarkVisualWidth benchmarks the visual width calculation
func BenchmarkVisualWidth(b *testing.B) {
	inputs := []struct {
		name string
		text string
	}{
		{"ASCII", "Hello World"},
		{"Unicode", "你好世界 Hello こんにちは"},
		{"ANSI", "\x1b[1m\x1b[36mColored Text\x1b[0m"},
		{"Mixed", "\x1b[31m错误\x1b[0m: \x1b[33m警告\x1b[0m message"},
		{"Long", strings.Repeat("a", 1000)},
		{"Long Unicode", strings.Repeat("你", 500)},
		{"Long ANSI", strings.Repeat("\x1b[31ma\x1b[0m", 100)},
	}

	for _, input := range inputs {
		b.Run(input.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = visualWidth(input.text)
			}
		})
	}
}

// BenchmarkStripANSI benchmarks ANSI stripping
func BenchmarkStripANSI(b *testing.B) {
	text := "\x1b[1m\x1b[36mBold Cyan\x1b[0m \x1b[31mRed\x1b[0m \x1b[38;5;214mOrange\x1b[0m"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stripANSI(text)
	}
}