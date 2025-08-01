package preview

import (
	"testing"
)

// BenchmarkMockViewRender benchmarks different mock view renders
func BenchmarkMockViewRender(b *testing.B) {
	testCases := []struct {
		name   string
		width  int
		height int
		render func(*MockView) string
	}{
		{"MainView_80x24", 80, 24, (*MockView).RenderMainView},
		{"MainView_120x40", 120, 40, (*MockView).RenderMainView},
		{"MainView_60x20", 60, 20, (*MockView).RenderMainView},
		{"HelpView_80x24", 80, 24, (*MockView).RenderHelpView},
		{"DeleteConfirm_80x24", 80, 24, (*MockView).RenderDeleteConfirmation},
		{"ModifyInput_80x24", 80, 24, (*MockView).RenderModifyInput},
		{"AIAnalysis_80x24", 80, 24, (*MockView).RenderAIAnalysis},
		{"ContextSelect_80x24", 80, 24, (*MockView).RenderContextSelect},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			mock := &MockView{
				Width:  tc.width,
				Height: tc.height,
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tc.render(mock)
			}
		})
	}
}

// BenchmarkRenderStatusBar benchmarks status bar rendering
func BenchmarkRenderStatusBar(b *testing.B) {
	testCases := []struct {
		name      string
		width     int
		progress  string
		taskTitle string
	}{
		{"Short_80", 80, "[2 of 15]", "Short task"},
		{"Long_80", 80, "[2 of 15]", "Implement user authentication system with OAuth2 support and multi-factor authentication"},
		{"Short_120", 120, "[2 of 15]", "Short task"},
		{"Long_120", 120, "[2 of 15]", "Implement user authentication system with OAuth2 support and multi-factor authentication"},
		{"Narrow_40", 40, "[2 of 15]", "Task in narrow terminal"},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			mock := &MockView{Width: tc.width}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = mock.renderStatusBar(tc.progress, tc.taskTitle)
			}
		})
	}
}

// BenchmarkRenderHelp benchmarks help text rendering
func BenchmarkRenderHelp(b *testing.B) {
	testCases := []struct {
		name     string
		expanded bool
		width    int
		height   int
	}{
		{"Collapsed_80x24", false, 80, 24},
		{"Expanded_80x24", true, 80, 24},
		{"Collapsed_120x40", false, 120, 40},
		{"Expanded_120x40", true, 120, 40},
		{"Collapsed_40x20", false, 40, 20},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			mock := &MockView{
				Width:  tc.width,
				Height: tc.height,
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = mock.renderHelp(tc.expanded)
			}
		})
	}
}

// BenchmarkStripANSISequences benchmarks ANSI stripping in mock views
func BenchmarkStripANSISequences(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"NoANSI", "Plain text without any ANSI codes"},
		{"SingleColor", "\x1b[36mCyan colored text\x1b[0m"},
		{"MultipleColors", "\x1b[36mCyan\x1b[0m \x1b[1m\x1b[35mBold Magenta\x1b[0m \x1b[90mGray\x1b[0m"},
		{"ComplexFormatting", "\x1b[1m\x1b[36m\x1b[4mBold Cyan Underlined\x1b[0m\x1b[0m\x1b[0m"},
		{"LongWithANSI", "\x1b[36m" + "This is a very long string with ANSI codes repeated many times " + "\x1b[0m"},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = stripANSISequences(tc.input)
			}
		})
	}
}

// BenchmarkGeneratePreview benchmarks the main preview generation
func BenchmarkGeneratePreview(b *testing.B) {
	states := []PreviewState{
		StateMain,
		StateHelp,
		StateDelete,
		StateModify,
		StateAIAnalysis,
		StateContextSelect,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts := PreviewOptions{
			State:  states[i%len(states)],
			Width:  80,
			Height: 24,
		}
		_ = GeneratePreview(opts)
	}
}

// BenchmarkRenderCalendar benchmarks calendar rendering
func BenchmarkRenderCalendar(b *testing.B) {
	modes := []string{"due", "wait", "scheduled"}
	widths := []int{80, 120, 60}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock := &MockView{
			Width:  widths[i%len(widths)],
			Height: 24,
		}
		_ = mock.RenderCalendar(modes[i%len(modes)])
	}
}