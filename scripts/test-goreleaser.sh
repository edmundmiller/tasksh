#!/bin/bash
# Test GoReleaser configuration locally

set -e

echo "Testing GoReleaser configuration..."

# Check if goreleaser is installed
if ! command -v goreleaser &> /dev/null; then
    echo "GoReleaser not found. Installing..."
    brew install goreleaser
fi

# Run GoReleaser in snapshot mode (doesn't publish)
echo "Running GoReleaser in snapshot mode..."
goreleaser release --snapshot --clean --skip=publish

echo "Build artifacts created in dist/"
echo ""
echo "To test the formula locally:"
echo "1. Create a test formula from the generated archive"
echo "2. Run: brew install --build-from-source ./path/to/formula.rb"