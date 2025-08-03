# CRUSH.md - Development Guide for Tasksh

## Build & Test Commands
- **Build**: `go build -o tasksh ./cmd/tasksh` or `make build`
- **Test all**: `go test ./...` or `make test`
- **Test single package**: `go test ./internal/review -v`
- **Test single function**: `go test ./internal/review -run TestReviewWithNoTasks -v`
- **Test with coverage**: `go test -cover ./internal/review`
- **Benchmarks**: `make bench` (standard) or `make bench-quick` (key metrics)
- **Run binary**: `./tasksh help` or `./tasksh diagnostics`

## Code Style & Conventions
- **Package structure**: Follow `internal/` layout with clear separation (cli, review, taskwarrior, ai, etc.)
- **Imports**: Standard library first, then external deps, then internal packages with blank line separation
- **Error handling**: Use `fmt.Errorf("context: %w", err)` for wrapping, check errors immediately
- **Types**: Use struct embedding for UI models, prefer composition over inheritance
- **Naming**: Use Go conventions - `TaskSuggestion`, `executeTask()`, `lazyLoadThreshold`
- **Comments**: Minimal - only for exported functions/types and complex logic
- **Testing**: Use `testdata.IsTaskwarriorAvailable()` for integration tests, `t.Skip()` when deps unavailable

## Architecture Patterns
- **UI**: Bubble Tea models with embedded components (viewport, help, spinner, etc.)
- **External deps**: Wrap in client packages (taskwarrior, timewarrior) with error handling
- **Performance**: Use batch loading, lazy loading for large datasets, progress indicators
- **Config**: Environment variables with sensible defaults (`TASKSH_LAZY_LOAD_THRESHOLD=100`)

## TUI Design Principles (inspired by Charm Crush)
- **Component composition**: Build reusable UI components with clear interfaces
- **Consistent theming**: Centralized color/style management with lipgloss
- **Smooth interactions**: Use animations, progress indicators, loading states
- **Context awareness**: Pass context.Context for cancellation, use pub/sub for communication
- **Professional UX**: Keyboard shortcuts, help system, fuzzy search, error handling

## Modern UI Features
- **Enable modern UI**: Set `TASKSH_MODERN_UI=true` environment variable
- **Component-based architecture**: Reusable TaskList, ProgressBar, HelpSystem components
- **Advanced theming**: Auto-detect dark/light mode, support NO_COLOR, custom themes via `TASKSH_THEME`
- **Smooth animations**: Progress bars with gradients, animated spinners, smooth transitions
- **Enhanced navigation**: Vim-style keys (j/k), fuzzy search (/), contextual help (?)
- **Better feedback**: Toast notifications, loading indicators, professional error handling
- **Layout system**: Flexible layouts (vertical, horizontal, border) for complex UIs

## Key Dependencies
- Bubble Tea for TUI, modernc.org/sqlite for database, openai-go for AI features
- External: `task` command (Taskwarrior), optional `mods` for AI
- Modern UI: sahilm/fuzzy for search, muesli/termenv for terminal detection