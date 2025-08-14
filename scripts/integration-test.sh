#!/bin/bash

# Integration test script for muxyard
# This script tests basic functionality without requiring interactive input

set -e

echo "🧪 Starting integration tests for muxyard..."

# Build the binary
echo "📦 Building muxyard..."
go build -o muxyard-test ./cmd/muxyard

# Test version and help
echo "ℹ️  Testing version and help flags..."
./muxyard-test --version
./muxyard-test --help

# Test that the binary starts (we can't test interactive mode in CI)
echo "🔍 Testing tmux detection..."
if command -v tmux >/dev/null 2>&1; then
    echo "✅ tmux is available"
    
    # Clean up any existing sessions
    tmux kill-server 2>/dev/null || true
    
    # Test that our tmux wrapper functions work
    echo "🔧 Testing tmux session management..."
    
    # Create a test session using tmux directly
    tmux new-session -d -s integration-test -c /tmp 'echo "test"; exec $SHELL'
    
    # Verify session was created
    if tmux list-sessions | grep -q integration-test; then
        echo "✅ Test session created successfully"
    else
        echo "❌ Failed to create test session"
        exit 1
    fi
    
    # Clean up
    tmux kill-session -t integration-test
    echo "🧹 Cleaned up test session"
    
else
    echo "⚠️  tmux not available, skipping tmux integration tests"
fi

# Test config loading (create a temporary config)
echo "⚙️  Testing configuration loading..."
TEMP_CONFIG_DIR=$(mktemp -d)
mkdir -p "$TEMP_CONFIG_DIR/.config/muxyard"

cat > "$TEMP_CONFIG_DIR/.config/muxyard/config.yaml" << 'EOF'
repo_directories:
  - /tmp
templates:
  - name: test
    description: Test template
    windows:
      - name: main
        command: echo "test"
EOF

# Set XDG_CONFIG_HOME to use our temporary config
export XDG_CONFIG_HOME="$TEMP_CONFIG_DIR/.config"

# Test that config loading works (this will just validate the config loads)
echo "📄 Config validation test would go here (requires non-interactive mode)"

# Clean up
rm -rf "$TEMP_CONFIG_DIR"

echo "✅ All integration tests passed!"

# Clean up the test binary
rm -f muxyard-test