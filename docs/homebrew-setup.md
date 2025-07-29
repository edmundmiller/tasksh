# Homebrew Tap Setup Instructions

This document explains how to set up and maintain the Homebrew tap for taskagent.

## Prerequisites

1. Create a GitHub Personal Access Token:
   - Go to GitHub Settings > Developer settings > Personal access tokens
   - Create a new token with `repo` scope
   - Save it as `HOMEBREW_TAP_GITHUB_TOKEN` in your repository secrets

2. Create the tap repository:
   ```bash
   gh repo create homebrew-taskagent --public \
     --description "Homebrew formulae for taskagent"
   ```

## Release Process

1. Tag a new version:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. GitHub Actions will automatically:
   - Build binaries for macOS and Linux
   - Create a GitHub release
   - Update the Homebrew formula in your tap repository

## Testing Locally

1. Test the GoReleaser configuration:
   ```bash
   ./scripts/test-goreleaser.sh
   ```

2. Test the formula locally:
   ```bash
   # After running GoReleaser in snapshot mode
   brew install --build-from-source ./dist/homebrew/Formula/taskagent.rb
   ```

## Naming Decision

The project is being renamed from `tasksh` to `taskagent` to better reflect its AI-powered capabilities. The binary will be called `taskagent`.

## Formula Details

The formula will:
- Install the binary as `taskagent`
- Support both Intel and ARM macOS
- Include proper architecture detection
- Generate SHA256 checksums automatically

## Troubleshooting

1. **Authentication errors**: Ensure your `HOMEBREW_TAP_GITHUB_TOKEN` has proper permissions
2. **Formula conflicts**: The name `taskagent` should be unique in the Homebrew ecosystem
3. **Build failures**: Check that CGO is properly configured for SQLite support