#!/bin/bash

# Claude Code pre-write hook for Go formatting and linting
# This script runs before Claude writes any Go files to ensure code quality

set -e

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to format Go files
format_go_files() {
    echo "Formatting Go files..."
    
    # Format all Go files in the project
    if command_exists gofmt; then
        find . -name "*.go" -not -path "./vendor/*" -exec gofmt -w {} \;
        echo "✓ Go files formatted with gofmt"
    else
        echo "⚠ gofmt not found, skipping formatting"
    fi
    
    # Run goimports if available (better than gofmt for imports)
    if command_exists goimports; then
        find . -name "*.go" -not -path "./vendor/*" -exec goimports -w {} \;
        echo "✓ Go imports organized with goimports"
    fi
}

# Function to run go mod tidy
tidy_modules() {
    echo "Tidying Go modules..."
    if [ -f "go.mod" ]; then
        go mod tidy
        echo "✓ Go modules tidied"
    else
        echo "⚠ No go.mod found, skipping module tidy"
    fi
}

# Function to run basic linting
lint_code() {
    echo "Running basic Go checks..."
    
    # Check for build errors
    if go build ./... >/dev/null 2>&1; then
        echo "✓ Go build successful"
    else
        echo "⚠ Go build failed - there may be compilation errors"
        go build ./...  # Show the errors
        return 1
    fi
    
    # Run go vet for common issues
    if command_exists go && go vet ./... >/dev/null 2>&1; then
        echo "✓ go vet passed"
    else
        echo "⚠ go vet found issues:"
        go vet ./...
    fi
    
    # Run golint if available
    if command_exists golint; then
        if golint ./... | grep -v "should have comment" | grep -q .; then
            echo "⚠ golint found issues:"
            golint ./... | grep -v "should have comment"
        else
            echo "✓ golint passed"
        fi
    fi
    
    # Run staticcheck if available
    if command_exists staticcheck; then
        if staticcheck ./... >/dev/null 2>&1; then
            echo "✓ staticcheck passed"
        else
            echo "⚠ staticcheck found issues:"
            staticcheck ./...
        fi
    fi
}

# Function to update documentation if needed
update_docs() {
    echo "Checking documentation..."
    
    # Ensure README exists
    if [ ! -f "README.md" ] && [ ! -f "docs/README.md" ]; then
        echo "⚠ No README.md found - consider adding project documentation"
    else
        echo "✓ Documentation exists"
    fi
}

# Main execution
main() {
    echo "🔧 Running Claude Code pre-write hook for Go project..."
    echo ""
    
    # Change to project root
    cd "$(dirname "$0")/.." || exit 1
    
    # Run formatting and checks
    format_go_files
    echo ""
    
    tidy_modules
    echo ""
    
    lint_code
    echo ""
    
    update_docs
    echo ""
    
    echo "✅ Pre-write hook completed successfully!"
    echo ""
    
    # Show summary of project structure
    echo "📁 Current project structure:"
    tree -I 'vendor|node_modules|.git' -L 2 2>/dev/null || ls -la
}

# Run the main function
main "$@"