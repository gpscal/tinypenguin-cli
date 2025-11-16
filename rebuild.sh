#!/bin/bash

# Rebuild script for TinyPenguin CLI
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

print_step() {
    echo -e "${BLUE}📦 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Get the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"

echo "🔨 TinyPenguin Rebuild Script"
echo "=============================="
echo ""

# Check if we're in the right directory
if [ ! -f "$PROJECT_ROOT/cli/go.mod" ]; then
    print_error "Cannot find cli/go.mod. Please run this script from the tinypenguin root directory"
    exit 1
fi

# Change to CLI directory
cd "$PROJECT_ROOT/cli"

print_step "Cleaning previous build artifacts..."
if make clean > /dev/null 2>&1; then
    print_success "Cleaned build artifacts"
else
    print_info "No previous build artifacts to clean"
fi

print_step "Generating protobuf files..."
if make proto-gen; then
    print_success "Protobuf files generated"
else
    print_error "Failed to generate protobuf files"
    print_info "Make sure protoc and protoc-gen-go plugins are installed"
    exit 1
fi

print_step "Building CLI tool..."
if make build-cli; then
    print_success "CLI tool built successfully"
else
    print_error "Failed to build CLI tool"
    exit 1
fi

print_step "Building gRPC server..."
if make build-server; then
    print_success "gRPC server built successfully"
else
    print_error "Failed to build gRPC server"
    exit 1
fi

print_step "Verifying binaries..."
if [ -f "bin/tinypenguin-cli" ]; then
    print_success "CLI binary found: bin/tinypenguin-cli"
    # Show binary info
    if command -v file > /dev/null; then
        print_info "Binary type: $(file bin/tinypenguin-cli | cut -d: -f2)"
    fi
    if command -v ls > /dev/null; then
        print_info "Binary size: $(ls -lh bin/tinypenguin-cli | awk '{print $5}')"
    fi
else
    print_error "CLI binary not found"
    exit 1
fi

if [ -f "bin/tinypenguin" ]; then
    print_success "Server binary found: bin/tinypenguin"
else
    print_error "Server binary not found"
    exit 1
fi

echo ""
echo "🎉 Rebuild Complete!"
echo "===================="
echo ""
echo "Binaries location:"
echo "  CLI:    $PROJECT_ROOT/cli/bin/tinypenguin-cli"
echo "  Server: $PROJECT_ROOT/cli/bin/tinypenguin"
echo ""
echo "Quick test:"
echo "  cd cli && ./bin/tinypenguin-cli --help"
echo ""
echo "Usage examples:"
echo "  ./bin/tinypenguin-cli run 'Check current users'"
echo "  ./bin/tinypenguin-cli --debug run 'List processes'"
echo "  ./bin/tinypenguin-cli --tools=false run 'Provide advice'"
echo ""
