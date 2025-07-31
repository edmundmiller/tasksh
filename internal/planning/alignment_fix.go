package planning

import (
	"strings"
	"unicode/utf8"
)

// stripANSI removes ANSI escape sequences from a string
func stripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Skip ANSI escape sequence
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			if i < len(s) {
				i++ // Skip 'm'
			}
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}

// visualWidth calculates the visual width of a string, accounting for:
// - ANSI escape sequences (zero width)
// - Unicode characters (variable width)
// - Special characters like box drawing
func visualWidth(s string) int {
	// Strip ANSI codes first
	clean := stripANSI(s)
	
	// Count runes for Unicode support
	return utf8.RuneCountInString(clean)
}

// padToWidth pads a string to a specific visual width
func padToWidth(s string, width int, padChar string) string {
	currentWidth := visualWidth(s)
	if currentWidth >= width {
		return s
	}
	
	padCount := width - currentWidth
	return s + strings.Repeat(padChar, padCount)
}

// truncateToWidth truncates a string to fit within a specific visual width
func truncateToWidth(s string, maxWidth int) string {
	if visualWidth(s) <= maxWidth {
		return s
	}
	
	// For simplicity, truncate based on rune count
	runes := []rune(stripANSI(s))
	if len(runes) > maxWidth-3 {
		return string(runes[:maxWidth-3]) + "..."
	}
	return s
}