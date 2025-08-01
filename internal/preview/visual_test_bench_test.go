package preview

import (
	"strings"
	"testing"
)

// BenchmarkSnapshotComparison benchmarks snapshot comparison logic
func BenchmarkSnapshotComparison(b *testing.B) {
	testCases := []struct {
		name     string
		actual   string
		expected string
	}{
		{
			name:     "Identical",
			actual:   strings.Repeat("This is a line of text\n", 20),
			expected: strings.Repeat("This is a line of text\n", 20),
		},
		{
			name:     "SmallDifference",
			actual:   strings.Repeat("This is a line of text\n", 20),
			expected: strings.Repeat("This is a line of text\n", 19) + "This is a different line\n",
		},
		{
			name:     "LargeDifference",
			actual:   strings.Repeat("Actual content\n", 50),
			expected: strings.Repeat("Expected content\n", 50),
		},
		{
			name:     "ANSICodes",
			actual:   strings.Repeat("\x1b[36mColored text\x1b[0m\n", 30),
			expected: strings.Repeat("\x1b[35mDifferent color\x1b[0m\n", 30),
		},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Simulate comparison logic
				if tc.actual != tc.expected {
					// Would show diff
				}
			}
		})
	}
}

// BenchmarkOutputNormalization benchmarks output normalization logic
func BenchmarkOutputNormalization(b *testing.B) {
	testCases := []struct {
		name   string
		output string
	}{
		{"Simple", "Simple text without special characters"},
		{"WithANSI", "\x1b[36mColored\x1b[0m \x1b[1mBold\x1b[0m text"},
		{"MultiLine", strings.Repeat("Line of text\n", 20)},
		{"TrailingSpaces", "Text with trailing spaces    \nAnother line    "},
		{"MixedLineEndings", "Unix\nWindows\r\nMixed\r\n"},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Simulate normalization
				normalized := strings.ReplaceAll(tc.output, "\r\n", "\n")
				normalized = strings.TrimRight(normalized, " \n")
				_ = normalized
			}
		})
	}
}

// BenchmarkResponsiveTest benchmarks responsive layout testing
func BenchmarkResponsiveTest(b *testing.B) {
	sizes := []struct {
		width  int
		height int
	}{
		{40, 20},
		{80, 24},
		{120, 40},
		{160, 50},
	}
	
	states := []PreviewState{
		StateMain,
		StateHelp,
		StateModify,
		StateAIAnalysis,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		size := sizes[i%len(sizes)]
		state := states[i%len(states)]
		
		opts := PreviewOptions{
			State:  state,
			Width:  size.width,
			Height: size.height,
		}
		
		output := GeneratePreview(opts)
		lines := strings.Split(output, "\n")
		
		// Simulate validation checks
		for _, line := range lines {
			if len(stripANSISequences(line)) > size.width {
				// Line too long
			}
		}
		if len(lines) > size.height {
			// Too many lines
		}
	}
}

// BenchmarkColorConsistency benchmarks color validation
func BenchmarkColorConsistency(b *testing.B) {
	// Generate sample outputs with various ANSI codes
	outputs := []string{
		strings.Repeat("\x1b[36mCyan\x1b[0m ", 50),
		strings.Repeat("\x1b[90mGray\x1b[0m \x1b[36mCyan\x1b[0m ", 30),
		strings.Repeat("\x1b[1m\x1b[35mBold Magenta\x1b[0m\n", 20),
		GeneratePreview(PreviewOptions{State: StateMain, Width: 80, Height: 24}),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output := outputs[i%len(outputs)]
		
		// Simulate color extraction
		primaryCount := strings.Count(output, "\x1b[36m")
		secondaryCount := strings.Count(output, "\x1b[90m")
		
		if primaryCount == 0 && secondaryCount == 0 {
			// No colors found
		}
	}
}

// BenchmarkGenerateAllPreviews benchmarks generating all preview states
func BenchmarkGenerateAllPreviews(b *testing.B) {
	states := []PreviewState{
		StateMain, StateHelp, StateDelete, StateModify,
		StateWaitCalendar, StateDueCalendar, StateContextSelect,
		StateAIAnalysis, StateAILoading, StatePromptAgent,
		StatePromptPreview, StateCelebration, StateError,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, state := range states {
			opts := PreviewOptions{
				State:  state,
				Width:  80,
				Height: 24,
			}
			_ = GeneratePreview(opts)
		}
	}
}