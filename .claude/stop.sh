#!/bin/bash

# Claude stop hook - Build tasksh binary
# This script runs when a Claude Code session ends

echo "🔨 Building tasksh binary..."

# Change to the project directory (in case hook is run from elsewhere)
cd "$(dirname "$0")/.." || exit 1

# Build the tasksh binary
if go build -o tasksh ./cmd/tasksh; then
    echo "✅ Successfully built tasksh binary"
    
    # Show binary info
    if [ -f "./tasksh" ]; then
        echo "📦 Binary size: $(ls -lh ./tasksh | awk '{print $5}')"
        echo "📍 Location: $(pwd)/tasksh"
    fi
else
    echo "❌ Failed to build tasksh binary"
    exit 1
fi