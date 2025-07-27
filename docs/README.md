# Tasksh - Interactive Task Review Shell

A modern, interactive task review interface for Taskwarrior, built with Go and featuring AI-powered task analysis and time tracking.

## Features

- **Interactive Review Interface**: Modern terminal UI built with Bubble Tea
- **AI-Powered Analysis**: Get intelligent suggestions for task improvements using mods
- **Time Tracking**: Built-in SQLite database for tracking task completion times
- **Smart Estimations**: Historical data-driven time estimates
- **Modern Architecture**: Clean Go codebase with proper package structure

## Installation

### Prerequisites

- Go 1.24.5 or later
- [Taskwarrior](https://taskwarrior.org/) installed and configured
- [mods](https://github.com/charmbracelet/mods) (optional, for AI features)

### Building from Source

```bash
git clone https://github.com/emiller/tasksh
cd tasksh
go build -o tasksh ./cmd/tasksh
```

### Installing

```bash
go install github.com/emiller/tasksh/cmd/tasksh@latest
```

## Usage

### Basic Commands

```bash
# Start task review
tasksh review

# Review with limit
tasksh review 10

# Show help
tasksh help

# Show diagnostics
tasksh diagnostics
```

### Review Interface

During review, you can:

- **r** - Mark task as reviewed
- **e** - Edit task (opens in editor)
- **m** - Modify task with quick changes
- **c** - Complete task (with time tracking)
- **d** - Delete task (with confirmation)
- **w** - Set task to waiting status
- **s** - Skip task (will need review later)
- **?** - Toggle help
- **q** - Quit review

### AI Analysis

When mods is installed, tasksh can provide AI-powered suggestions for:

- Task prioritization
- Due date recommendations
- Time estimates based on historical data
- Project and tag suggestions

## Configuration

Tasksh automatically configures the required Taskwarrior UDA (User Defined Attribute) and report:

- **reviewed UDA**: Tracks when tasks were last reviewed
- **_reviewed report**: Shows tasks needing review

The review filter can be customized in your Taskwarrior configuration:

```bash
task config report._reviewed.filter "( reviewed.none: or reviewed.before:now-6days ) and ( +PENDING or +WAITING )"
```

## Time Tracking Database

Tasksh maintains a local SQLite database at `~/.local/share/tasksh/timedb.sqlite3` to track:

- Task completion times
- Estimation accuracy
- Historical patterns for better future estimates

## Development

### Project Structure

```
cmd/tasksh/           # Main application entry point
internal/
├── cli/             # Command-line interface handlers
├── review/          # Task review functionality
├── ai/              # AI integration
├── taskwarrior/     # Taskwarrior integration
└── timedb/          # Time tracking database
testdata/            # Test utilities
docs/                # Documentation
scripts/             # Build and utility scripts
.claude/             # Claude Code hooks
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/review
```

### Building

```bash
# Build binary
go build -o tasksh ./cmd/tasksh

# Build and test
go build -o tasksh ./cmd/tasksh && ./tasksh help
```

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## Troubleshooting

### Common Issues

**Taskwarrior not found**
- Ensure Taskwarrior is installed and the `task` command is in your PATH
- Run `tasksh diagnostics` to check system status

**AI features not working**
- Install mods: `go install github.com/charmbracelet/mods@latest`
- Configure mods with your preferred AI provider

**Interactive review not working**
- Ensure you're running in a proper terminal (not headless)
- Check that TTY is available for interactive input

For more issues, check the [troubleshooting guide](https://github.com/emiller/tasksh/issues) or file a new issue.