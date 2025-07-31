package preview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestUISnapshots tests that UI previews match expected snapshots
func TestUISnapshots(t *testing.T) {
	// Define test cases for each UI state
	testCases := []struct {
		name   string
		state  PreviewState
		width  int
		height int
	}{
		{"main view", StateMain, 80, 24},
		{"help view", StateHelp, 80, 24},
		{"delete confirmation", StateDelete, 80, 24},
		{"modify input", StateModify, 80, 24},
		{"context selection", StateContextSelect, 80, 24},
		{"error state", StateError, 80, 24},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate preview
			opts := PreviewOptions{
				State:  tc.state,
				Width:  tc.width,
				Height: tc.height,
			}
			actual := GeneratePreview(opts)

			// Load or create golden file
			goldenPath := filepath.Join("testdata", "snapshots", string(tc.state)+".golden")
			
			if os.Getenv("UPDATE_SNAPSHOTS") == "true" {
				// Update mode: write actual output as new golden
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("Failed to create snapshot directory: %v", err)
				}
				if err := os.WriteFile(goldenPath, []byte(actual), 0644); err != nil {
					t.Fatalf("Failed to write golden file: %v", err)
				}
				t.Logf("Updated snapshot: %s", goldenPath)
				return
			}

			// Read golden file
			golden, err := os.ReadFile(goldenPath)
			if err != nil {
				if os.IsNotExist(err) {
					t.Fatalf("Golden file does not exist: %s\nRun with UPDATE_SNAPSHOTS=true to create", goldenPath)
				}
				t.Fatalf("Failed to read golden file: %v", err)
			}

			// Compare actual vs golden
			if actual != string(golden) {
				t.Errorf("UI output does not match snapshot")
				t.Logf("Differences found in %s", goldenPath)
				
				// Show diff
				actualLines := strings.Split(actual, "\n")
				goldenLines := strings.Split(string(golden), "\n")
				
				maxLines := len(actualLines)
				if len(goldenLines) > maxLines {
					maxLines = len(goldenLines)
				}
				
				for i := 0; i < maxLines; i++ {
					var actualLine, goldenLine string
					if i < len(actualLines) {
						actualLine = actualLines[i]
					}
					if i < len(goldenLines) {
						goldenLine = goldenLines[i]
					}
					
					if actualLine != goldenLine {
						t.Logf("Line %d differs:", i+1)
						t.Logf("  Expected: %q", goldenLine)
						t.Logf("  Actual:   %q", actualLine)
					}
				}
				
				t.Logf("\nRun with UPDATE_SNAPSHOTS=true to update the snapshot")
			}
		})
	}
}

// TestResponsiveLayouts tests UI at different terminal sizes
func TestResponsiveLayouts(t *testing.T) {
	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"small terminal", 60, 20},
		{"standard terminal", 80, 24},
		{"large terminal", 120, 40},
		{"ultra-wide terminal", 160, 30},
	}

	states := []PreviewState{
		StateMain,
		StateHelp,
		StateModify,
	}

	for _, size := range sizes {
		for _, state := range states {
			t.Run(size.name+" - "+string(state), func(t *testing.T) {
				opts := PreviewOptions{
					State:  state,
					Width:  size.width,
					Height: size.height,
				}
				
				preview := GeneratePreview(opts)
				
				// Basic validation
				lines := strings.Split(preview, "\n")
				
				// Check no line exceeds terminal width
				for i, line := range lines {
					// Remove ANSI color codes for accurate length
					cleanLine := stripANSI(line)
					if len(cleanLine) > size.width {
						t.Errorf("Line %d exceeds terminal width (%d > %d): %s", 
							i+1, len(cleanLine), size.width, line)
					}
				}
				
				// Ensure content fits within height
				if len(lines) > size.height {
					t.Errorf("Content exceeds terminal height (%d > %d)", 
						len(lines), size.height)
				}
			})
		}
	}
}

// stripANSI removes ANSI escape sequences from a string
func stripANSI(s string) string {
	// Simple ANSI stripping - in production use a proper library
	for {
		start := strings.Index(s, "\x1b[")
		if start == -1 {
			break
		}
		end := strings.IndexByte(s[start:], 'm')
		if end == -1 {
			break
		}
		s = s[:start] + s[start+end+1:]
	}
	return s
}

// TestColorConsistency ensures colors are used consistently
func TestColorConsistency(t *testing.T) {
	opts := PreviewOptions{
		State:  StateMain,
		Width:  80,
		Height: 24,
	}
	
	preview := GeneratePreview(opts)
	
	// Define expected color patterns
	colorPatterns := map[string]string{
		"primary actions": "\x1b[36m", // Cyan
		"secondary actions": "\x1b[90m", // Bright black (gray)
		"important values": "\x1b[33m", // Yellow
		"error states": "\x1b[31m", // Red
	}
	
	// Check that primary actions use consistent colors
	if strings.Contains(preview, "r: review") {
		if !strings.Contains(preview, colorPatterns["primary actions"]+"r: review") &&
		   !strings.Contains(preview, "\x1b[1m\x1b[36mr: review") { // With bold
			t.Error("Primary action 'review' not using expected cyan color")
		}
	}
}

// BenchmarkPreviewGeneration measures preview generation performance
func BenchmarkPreviewGeneration(b *testing.B) {
	states := []PreviewState{
		StateMain,
		StateHelp,
		StateModify,
		StateAIAnalysis,
	}
	
	for _, state := range states {
		b.Run(string(state), func(b *testing.B) {
			opts := PreviewOptions{
				State:  state,
				Width:  80,
				Height: 24,
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = GeneratePreview(opts)
			}
		})
	}
}