#!/bin/bash

# Quick setup script for TinyPenguin
set -e

echo "🐧 TinyPenguin Quick Setup"
echo "=========================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_step() {
    echo -e "${YELLOW}📦 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if we're in the right directory
if [ ! -f "cli/go.mod" ]; then
    print_error "Please run this script from the tinypenguin root directory"
    exit 1
fi

print_step "Installing Protocol Buffers compiler..."

# Install protoc based on OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y protobuf-compiler
    elif command -v yum &> /dev/null; then
        sudo yum install -y protobuf
    elif command -v dnf &> /dev/null; then
        sudo dnf install -y protobuf-devel
    else
        print_error "Package manager not found. Please install protobuf-compiler manually."
        exit 1
    fi
elif [[ "$OSTYPE" == "darwin"* ]]; then
    if command -v brew &> /dev/null; then
        brew install protobuf
    else
        print_error "Homebrew not found. Please install protobuf manually."
        exit 1
    fi
else
    print_error "Unsupported OS. Please install protobuf-compiler manually."
    exit 1
fi

print_success "Protocol Buffers compiler installed"

print_step "Building TinyPenguin..."

cd cli

# Clean previous builds
make clean || true

# Generate protobuf files
if make proto-gen; then
    print_success "Protobuf files generated"
else
    print_error "Failed to generate protobuf files"
    exit 1
fi

# Build the project
if make build; then
    print_success "Build completed successfully"
else
    print_error "Build failed"
    exit 1
fi

print_step "Verifying installation..."

# Check if binaries exist
if [ -f "bin/tinypenguin-cli" ] && [ -f "bin/tinypenguin" ]; then
    print_success "Binaries created successfully"
else
    print_error "Binaries not found"
    exit 1
fi

# Test CLI help
if ./bin/tinypenguin-cli --help > /dev/null; then
    print_success "CLI tool is working"
else
    print_error "CLI tool test failed"
    exit 1
fi

print_step "Running basic functionality test..."

# Test with a simple query (non-destructive)
if timeout 30s ./bin/tinypenguin-cli run "What is the current date and time?" > setup_test.txt 2>&1; then
    if [ -s setup_test.txt ]; then
        print_success "Basic functionality test passed"
    else
        print_error "No output from test command"
    fi
else
    print_error "Test command failed or timed out"
fi

# Clean up test file
rm -f setup_test.txt

echo ""
echo "🎉 Setup Complete!"
echo "================="
echo ""
echo "Usage Examples:"
echo "  ./bin/tinypenguin-cli run 'Create a new user named john'"
echo "  ./bin/tinypenguin-cli run 'Install nginx package'"
echo "  ./bin/tinypenguin-cli run 'Create a backup script'"
echo ""
echo "Server Mode:"
echo "  ./bin/tinypenguin"
echo ""
echo "Prerequisites:"
echo "- tinyllama must be running on http://localhost:11434/v1"
echo "- Use './bin/tinypenguin-cli --help' for all available options"
echo ""
echo "Next Steps:"
echo "1. Start tinyllama if not already running"
echo "2. Try the example commands above"
echo "3. See README.md for detailed documentation"
echo "4. Run './test-implementation.sh' for comprehensive testing"