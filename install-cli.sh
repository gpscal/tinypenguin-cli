#!/bin/bash

# Install script for TinyPenguin CLI
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI_BINARY="$SCRIPT_DIR/cli/bin/tinypenguin-cli"
INSTALL_PATH="/usr/local/bin/tinypenguin-cli"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if binary exists
if [ ! -f "$CLI_BINARY" ]; then
    print_error "CLI binary not found at $CLI_BINARY"
    print_info "Please run './rebuild.sh' first to build the CLI"
    exit 1
fi

print_info "Installing tinypenguin-cli..."
print_info "Source: $CLI_BINARY"
print_info "Destination: $INSTALL_PATH"

# Check if we can write to /usr/local/bin
if [ ! -w "/usr/local/bin" ]; then
    print_info "Requires sudo to install to /usr/local/bin"
    sudo cp "$CLI_BINARY" "$INSTALL_PATH"
    sudo chmod +x "$INSTALL_PATH"
else
    cp "$CLI_BINARY" "$INSTALL_PATH"
    chmod +x "$INSTALL_PATH"
fi

if [ -f "$INSTALL_PATH" ]; then
    print_success "tinypenguin-cli installed successfully!"
    echo ""
    echo "You can now use it from anywhere:"
    echo "  tinypenguin-cli --help"
    echo "  tinypenguin-cli --tools=false run 'Your query'"
    echo "  tinypenguin-cli --debug run 'Your query'"
    echo ""
    
    # Verify installation
    if command -v tinypenguin-cli > /dev/null 2>&1; then
        print_success "Installation verified - CLI is in PATH"
        tinypenguin-cli --help | head -15
    else
        print_error "CLI installed but not found in PATH"
        print_info "Make sure /usr/local/bin is in your PATH"
    fi
else
    print_error "Installation failed"
    exit 1
fi
