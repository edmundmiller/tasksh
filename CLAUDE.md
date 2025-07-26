# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build System

This project uses Go for building:

- **Build**: `go build -o tasksh`
- **Install**: `go install` or copy binary to desired location
- **Dependencies**: Managed automatically with `go mod`

## Testing

- **Run tests**: `go test ./...`
- **Build and test**: `go build -o tasksh && ./tasksh help`

## Architecture

Tasksh is a task review interface for Taskwarrior written in Go using the Huh library for interactive forms. Key components:

- **main.go**: Entry point and command parsing (supports `help`, `diagnostics`, `review` commands)
- **taskwarrior.go**: Taskwarrior integration functions that execute task commands and parse results
- **review.go**: Interactive task review functionality using Huh forms and prompts

The application provides a modern terminal interface for reviewing Taskwarrior tasks with options to edit, modify, complete, delete, skip, or mark tasks as reviewed.

## Dependencies

- **Go**: Programming language and build system
- **github.com/charmbracelet/huh**: Interactive forms and prompts library for terminal interfaces
- **Taskwarrior**: External dependency - the `task` command must be available in PATH

## Key Features

- **Interactive forms**: Uses Huh library for modern terminal interface
- **Task review workflow**: Systematic review of tasks needing attention
- **UDA management**: Automatically configures `reviewed` UDA and `_reviewed` report
- **Progress tracking**: Shows current position in review process
- **Flexible actions**: Edit, modify, complete, delete, skip, or mark as reviewed
- **Error handling**: Graceful handling of Taskwarrior command failures

## Project Structure

- Root directory contains all Go source files
- `go.mod` and `go.sum` manage dependencies
- Single binary executable `tasksh`