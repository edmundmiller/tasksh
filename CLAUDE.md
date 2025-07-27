# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Tasksh is a modern Go application providing an interactive task review interface for Taskwarrior. It features AI-powered analysis, time tracking, and a clean modular architecture.

## Build System

This project uses Go with a standard project layout:

- **Build**: `go build -o tasksh ./cmd/tasksh`
- **Install**: `go install github.com/emiller/tasksh/cmd/tasksh@latest`
- **Dependencies**: Managed with `go mod`
- **Main entry point**: `cmd/tasksh/main.go`

## Project Structure

```
cmd/tasksh/           # Application entry point
internal/
├── cli/             # Command-line interface handlers
├── review/          # Task review functionality  
├── ai/              # AI integration
├── taskwarrior/     # Taskwarrior integration
└── timedb/          # Time tracking database
testdata/            # Test utilities
docs/                # Documentation
.claude/             # Claude Code hooks
```

## Testing

- **Run all tests**: `go test ./...`
- **Test with coverage**: `go test -cover ./...`
- **Test specific package**: `go test ./internal/review`
- **Build and test**: `go build -o tasksh ./cmd/tasksh && ./tasksh help`

### Test Structure

- **internal/cli/tasksh_test.go**: CLI command tests
- **internal/review/review_test.go**: Review workflow tests
- **internal/ai/ai_test.go**: AI analysis tests
- **internal/timedb/timedb_test.go**: Database operation tests
- **testdata/test_utils.go**: Shared test utilities

Tests use isolated environments and skip integration tests when dependencies (like Taskwarrior) are unavailable.

## Architecture

Modern Go application with clean separation of concerns:

- **cmd/tasksh**: Entry point, minimal logic
- **internal/cli**: Command handlers (help, diagnostics)
- **internal/review**: Core review functionality with Bubble Tea UI
- **internal/taskwarrior**: Taskwarrior command abstraction
- **internal/ai**: AI-powered task analysis using mods
- **internal/timedb**: SQLite-based time tracking

## Dependencies

- **Go 1.24.5+**: Programming language
- **github.com/charmbracelet/huh**: Interactive forms
- **github.com/charmbracelet/bubbletea**: Terminal UI framework
- **github.com/mattn/go-sqlite3**: SQLite database driver
- **Taskwarrior**: External dependency (`task` command)
- **mods**: Optional, for AI features

## Key Features

- **Modular Architecture**: Clean package separation following Go conventions
- **Interactive UI**: Modern Bubble Tea interface with keyboard shortcuts
- **AI Integration**: Intelligent task suggestions via mods
- **Time Tracking**: Built-in SQLite database for completion analytics
- **Test Coverage**: Comprehensive test suite with isolated environments
- **Documentation**: Detailed architecture and API documentation

## Development Guidelines

1. **Package Organization**: Follow the established internal package structure
2. **Error Handling**: Use standard Go patterns with context wrapping
3. **Testing**: Write tests for all new functionality
4. **Dependencies**: Minimize external dependencies, prefer standard library
5. **Documentation**: Update docs for architectural changes

## Common Commands

```bash
# Development workflow
go build -o tasksh ./cmd/tasksh
go test ./...
./tasksh diagnostics

# Package-specific testing
go test ./internal/review -v
go test ./internal/ai -cover

# Full build verification
go build -o tasksh ./cmd/tasksh && ./tasksh help
```