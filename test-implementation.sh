#!/bin/bash

# Test script for TinyPenguin implementation
set -e

echo "🚀 Testing TinyPenguin Implementation"
echo "======================================"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test directory
TEST_DIR="/tmp/tinypenguin-test"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
        exit 1
    fi
}

# Function to print info
print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

echo ""
echo "📋 Pre-requisites Check"
echo "----------------------"

# Check if tinyllama is running
print_info "Checking tinyllama connection..."
if curl -s http://localhost:11434/v1/models > /dev/null; then
    print_result 0 "tinyllama is accessible"
else
    print_result 1 "tinyllama is not running or accessible"
fi

# Check Go installation
print_info "Checking Go installation..."
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | cut -d' ' -f3)
    print_result 0 "Go is installed ($GO_VERSION)"
else
    print_result 1 "Go is not installed"
fi

# Check protoc installation
print_info "Checking protoc installation..."
if command -v protoc &> /dev/null; then
    PROTOC_VERSION=$(protoc --version)
    print_result 0 "protoc is installed ($PROTOC_VERSION)"
else
    print_result 1 "protoc is not installed"
fi

echo ""
echo "🏗️  Building Project"
echo "-------------------"

# Change to tinypenguin directory
cd /home/cal/tinyllama_sh_scripting/tinypenguin/cli

# Clean previous builds
print_info "Cleaning previous builds..."
make clean || true

# Generate protobuf files
print_info "Generating protobuf files..."
if make proto-gen; then
    print_result 0 "Protobuf generation successful"
else
    print_result 1 "Protobuf generation failed"
fi

# Build the project
print_info "Building the project..."
if make build; then
    print_result 0 "Build successful"
else
    print_result 1 "Build failed"
fi

echo ""
echo "🧪 Testing CLI Tool"
echo "-------------------"

# Check if binaries were created
if [ -f "bin/tinypenguin-cli" ]; then
    print_result 0 "CLI binary created"
else
    print_result 1 "CLI binary not found"
fi

if [ -f "bin/tinypenguin" ]; then
    print_result 0 "Server binary created"
else
    print_result 1 "Server binary not found"
fi

# Test CLI help
print_info "Testing CLI help command..."
if ./bin/tinypenguin-cli --help > /dev/null; then
    print_result 0 "CLI help command works"
else
    print_result 1 "CLI help command failed"
fi

echo ""
echo "🔧 Functional Tests"
echo "------------------"

# Test command execution (non-destructive)
print_info "Testing command execution..."
if timeout 10s ./bin/tinypenguin-cli run "Show current date and time" > test_output.txt 2>&1; then
    if grep -q "success\|completed" test_output.txt || [ -s test_output.txt ]; then
        print_result 0 "Command execution test passed"
    else
        print_info "Command execution output:"
        cat test_output.txt
        print_result 1 "Command execution test failed"
    fi
else
    print_info "Command execution test output:"
    cat test_output.txt 2>/dev/null || echo "No output file"
    print_result 1 "Command execution test timed out or failed"
fi

# Test file creation
print_info "Testing file operations..."
TEST_FILE="$TEST_DIR/test-file.txt"
if timeout 10s ./bin/tinypenguin-cli run "Create a test file at $TEST_FILE with content 'Hello, TinyPenguin!'" > /dev/null 2>&1; then
    if [ -f "$TEST_FILE" ]; then
        print_result 0 "File creation test passed"
        rm -f "$TEST_FILE"
    else
        print_result 1 "File creation test failed - file not created"
    fi
else
    print_result 1 "File creation test timed out or failed"
fi

echo ""
echo "🔒 Security Tests"
echo "-----------------"

# Test dangerous command blocking
print_info "Testing dangerous command blocking..."
if timeout 10s ./bin/tinypenguin-cli run "Delete all files with rm -rf /" > security_test.txt 2>&1; then
    if grep -q "denied\|blocked\|unsafe\|dangerous" security_test.txt; then
        print_result 0 "Dangerous command blocking works"
    else
        print_info "Security test output:"
        cat security_test.txt
        print_result 1 "Dangerous command not properly blocked"
    fi
else
    print_info "Security test output:"
    cat security_test.txt 2>/dev/null || echo "No output file"
    print_result 1 "Security test timed out"
fi

echo ""
echo "🌐 Network Tests"
echo "----------------"

# Test server startup (in background)
print_info "Testing server startup..."
./bin/tinypenguin -port 50052 > server.log 2>&1 &
SERVER_PID=$!
sleep 2

if kill -0 $SERVER_PID 2>/dev/null; then
    print_result 0 "Server started successfully"
    kill $SERVER_PID
    wait $SERVER_PID 2>/dev/null || true
else
    print_result 1 "Server failed to start"
    cat server.log
fi

echo ""
echo "📊 Test Summary"
echo "--------------"

# Count test files
TEST_FILES=$(find /tmp/tinypenguin-test -name "*.txt" -o -name "*.log" | wc -l)
echo "Created $TEST_FILES test files in $TEST_DIR"

# Clean up
print_info "Cleaning up test files..."
rm -rf "$TEST_DIR"

echo ""
echo "🎉 TinyPenguin Implementation Test Complete!"
echo "============================================"
echo ""
echo "Next Steps:"
echo "1. Start tinyllama if not already running"
echo "2. Run: ./bin/tinypenguin-cli run 'Your query here'"
echo "3. For server mode: ./bin/tinypenguin"
echo "4. See README.md for detailed usage instructions"
echo ""
echo "Example commands to try:"
echo "  ./bin/tinypenguin-cli run 'Create a user named testuser'"
echo "  ./bin/tinypenguin-cli run 'Install nginx package'"
echo "  ./bin/tinypenguin-cli run 'Create a backup script'"