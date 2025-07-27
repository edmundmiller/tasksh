# Tasksh Architecture

This document describes the architecture and design decisions for Tasksh, a modern Go rewrite of the task review shell for Taskwarrior.

## Overview

Tasksh is designed as a modular, maintainable Go application following standard Go project conventions. The architecture prioritizes:

- **Separation of concerns** - Each package has a single responsibility
- **Testability** - All components are testable in isolation
- **AI-friendliness** - Clear structure for AI agents to understand and modify
- **Maintainability** - Clean interfaces and minimal dependencies

## Project Structure

```
tasksh/
├── cmd/tasksh/           # Application entry point
│   └── main.go          # CLI routing and initialization
├── internal/            # Private application packages
│   ├── cli/            # Command handlers
│   │   ├── diagnostics.go
│   │   └── help.go
│   ├── review/         # Task review functionality
│   │   ├── review.go   # Main review logic
│   │   └── bubbletea.go # Interactive UI
│   ├── ai/             # AI integration
│   │   ├── analysis.go # Task analysis logic
│   │   └── mods.go     # Mods integration
│   ├── taskwarrior/    # Taskwarrior integration
│   │   ├── client.go   # Command execution
│   │   └── config.go   # Configuration helpers
│   └── timedb/         # Time tracking database
│       ├── database.go # Database operations
│       └── models.go   # Data models and queries
├── testdata/           # Test utilities and fixtures
├── docs/               # Documentation
├── scripts/            # Build and development scripts
└── .claude/            # Claude Code hooks
```

## Package Responsibilities

### `cmd/tasksh`
- Application entry point
- Command-line argument parsing
- Route commands to appropriate handlers
- Minimal logic, delegates to internal packages

### `internal/cli`
- Command handlers for help, diagnostics, version
- No business logic, purely presentation layer
- Calls other packages for functionality

### `internal/review`
- Core task review functionality
- Bubble Tea interactive interface
- Review workflow orchestration
- Integrates with taskwarrior, ai, and timedb packages

### `internal/taskwarrior`
- All Taskwarrior command execution
- Task data structure definitions
- Configuration management
- Command abstraction layer

### `internal/ai`
- AI-powered task analysis
- Integration with mods command
- Prompt engineering for task suggestions
- Response parsing and validation

### `internal/timedb`
- SQLite database for time tracking
- Task completion history
- Time estimation algorithms
- Database schema management

### `testdata`
- Shared test utilities
- Test fixtures and helpers
- Isolated testing environments
- Mock implementations

## Design Principles

### 1. Dependency Direction

Dependencies flow inward toward the core business logic:

```
cmd/tasksh → internal/review → internal/{taskwarrior,ai,timedb}
          → internal/cli    → internal/{taskwarrior,ai,timedb}
```

- `cmd` packages depend on `internal` packages
- `internal/review` and `internal/cli` are consumers
- `internal/taskwarrior`, `internal/ai`, `internal/timedb` are providers
- No circular dependencies

### 2. Interface Segregation

Each package exposes minimal, focused interfaces:

- `taskwarrior` package: Task CRUD operations, configuration
- `timedb` package: Time tracking, estimation queries
- `ai` package: Task analysis and suggestions
- `review` package: Review workflow orchestration

### 3. Error Handling

- Use standard Go error handling patterns
- Wrap errors with context using `fmt.Errorf`
- Fail fast for configuration issues
- Graceful degradation for optional features (AI, time tracking)

### 4. Testing Strategy

- Unit tests for each package
- Integration tests for Taskwarrior interaction
- Test utilities in `testdata` package
- Isolated test environments to avoid side effects

## Data Flow

### Task Review Flow

1. **Initialization**
   ```go
   review.Run(limit) // Entry point
   ```

2. **Setup**
   ```go
   taskwarrior.EnsureReviewConfig() // Configure Taskwarrior
   taskwarrior.GetTasksForReview()  // Get task list
   ```

3. **Interactive Review**
   ```go
   bubbletea.NewReviewModel() // Create UI model
   // User interactions through Bubble Tea
   ```

4. **Task Operations**
   ```go
   taskwarrior.EditTask()     // Edit operations
   taskwarrior.CompleteTask() // Completion
   timedb.RecordCompletion()  // Time tracking
   ai.AnalyzeTask()          // AI suggestions
   ```

### AI Analysis Flow

1. **Context Gathering**
   ```go
   timedb.GetSimilarTasks()        // Historical data
   timedb.EstimateTimeForTask()    // Time estimates
   ```

2. **Prompt Generation**
   ```go
   ai.buildAnalysisPrompt() // Structured prompt
   ```

3. **AI Execution**
   ```go
   mods command execution // External AI call
   ```

4. **Response Processing**
   ```go
   ai.parseAnalysisResponse() // JSON parsing
   ```

## Configuration Management

### Taskwarrior Configuration

- Automatic UDA setup for `reviewed` field
- Custom report configuration for `_reviewed`
- Non-destructive configuration (checks before setting)

### Application Configuration

- Minimal configuration surface
- Environment-based database location
- Graceful fallbacks for missing dependencies

## Error Handling Patterns

### Command Execution Errors
```go
if err := taskwarrior.CompleteTask(uuid); err != nil {
    return fmt.Errorf("failed to complete task: %w", err)
}
```

### Optional Feature Errors
```go
if err := ai.CheckModsAvailable(); err != nil {
    // Log warning but continue without AI features
    log.Printf("AI features unavailable: %v", err)
}
```

### Database Errors
```go
if db, err := timedb.New(); err != nil {
    // Continue without time tracking
    log.Printf("Time tracking unavailable: %v", err)
}
```

## Performance Considerations

### Database Operations
- SQLite for local time tracking data
- Indexed queries for performance
- Connection pooling not needed (single user)

### External Command Execution
- Minimal `task` command invocations
- Efficient task data retrieval using `_get`
- Batch operations where possible

### Memory Usage
- Streaming task processing
- Minimal data caching
- Bubble Tea's efficient rendering

## Security Considerations

### Command Injection Prevention
- Use `exec.Command` with separate arguments
- No shell interpretation of user input
- Validate UUIDs before command execution

### File System Access
- Database in standard user data directory
- No privileged file access required
- Respect XDG Base Directory specification

## Extensibility Points

### New AI Providers
- Implement `ai.Analyzer` interface
- Add provider-specific configuration
- Maintain common prompt format

### Additional Time Tracking
- Extend `timedb.TimeEntry` model
- Add new query functions
- Maintain backward compatibility

### Custom Review Actions
- Add new key bindings in `bubbletea.go`
- Implement action handlers
- Update help text and documentation

## Migration Strategy

This architecture supports gradual migration from the original C++ implementation:

1. **Phase 1**: Core functionality (completed)
   - Task review workflow
   - Taskwarrior integration
   - Basic UI

2. **Phase 2**: Enhanced features (completed)
   - AI integration
   - Time tracking
   - Improved UI

3. **Phase 3**: Optimization (future)
   - Performance improvements
   - Additional AI providers
   - Enhanced analytics

The modular design allows for incremental improvements without major refactoring.