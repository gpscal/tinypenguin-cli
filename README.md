# TinyPenguin - RHCSA Assistant

TinyPenguin is a CLI tool that integrates with tinyllama to provide Red Hat Certified System Administrator (RHCSA) assistance. It helps with Linux system administration tasks including file operations, package management, service control, and more.

## Features

- **AI-Powered Assistance**: Leverages tinyllama for intelligent system administration
- **RHCSA-Focused**: Specialized for Red Hat system administration tasks
- **Safe Command Execution**: Built-in security checks to prevent dangerous operations
- **File Editing**: Edit files using diff-based or direct content replacement
- **Command Execution**: Run shell commands with timeout and approval controls
- **Tool Integration**: Extensible tool system for various system administration tasks

## Prerequisites

- **tinyllama**: Running locally on `http://localhost:11434/v1`
- **Go 1.20+**: For building the CLI tools
- **Protocol Buffers Compiler**: For generating gRPC code
- **Node.js 16+**: For TypeScript components (optional)

## Installation

### 1. Install Protocol Buffers Compiler

**Linux:**
```bash
make install-protoc-linux
```

**macOS (using Homebrew):**
```bash
make install-protoc-macos
```

**Manual Installation:**
Download from [Protocol Buffers releases](https://github.com/protocolbuffers/protobuf/releases) and add `protoc` to your PATH.

### 2. Build the Project

```bash
# Clone the repository
git clone <repository-url>
cd tinypenguin

# Generate protobuf files and build
make build
```

### 3. Install CLI Tool (Optional)

```bash
sudo make install
```

## Setup

### 1. Start tinyllama

Ensure tinyllama is running on the default port:

```bash
# Start tinyllama (example command)
tinyllama serve --port 11434
```

### 2. Verify Installation

```bash
# Test the CLI tool
./bin/tinypenguin-cli --help

# Test tinyllama connection
./bin/tinypenguin-cli run "Test connection"
```

## Usage

### Basic Commands

```bash
# Run a task
tinypenguin-cli run "Create a new user named john"

# Install a package
tinypenguin-cli run "Install nginx web server"

# Create a bash script
tinypenguin-cli run "Create a backup script that compresses /home directory"

# Edit a configuration file
tinypenguin-cli run "Add a new user to /etc/sudoers"

# Check system status
tinypenguin-cli run "Show disk usage and running services"
```

### Advanced Usage

```bash
# Specify custom tinyllama URL
tinypenguin-cli --url http://localhost:11434/v1 run "Your query here"

# Use specific model
tinypenguin-cli --model tinyllama run "Your query here"

# List available tasks
tinypenguin-cli list

# Cancel a running task
tinypenguin-cli cancel --task-id task-123
```

### Server Mode

Start the gRPC server for programmatic access:

```bash
# Start the server
./bin/tinypenguin

# Or with custom port
./bin/tinypenguin -port 50051
```

## RHCSA Task Examples

### User Management
```bash
# Create a new user
tinypenguin-cli run "Create a user named alice with home directory /home/alice"

# Add user to sudo group
tinypenguin-cli run "Add user alice to the wheel group for sudo access"

# Set user password
tinypenguin-cli run "Set password for user alice"
```

### Package Management
```bash
# Install packages
tinypenguin-cli run "Install apache web server and php"

# Update system
tinypenguin-cli run "Update all installed packages"

# Check package information
tinypenguin-cli run "Show information about installed nginx package"
```

### Service Management
```bash
# Start/stop services
tinypenguin-cli run "Start the nginx service and enable it at boot"

# Check service status
tinypenguin-cli run "Show status of all running services"

# Configure firewall
tinypenguin-cli run "Open port 80 and 443 in the firewall"
```

### File System Operations
```bash
# Create directories
tinypenguin-cli run "Create /opt/app directory with proper permissions"

# Edit configuration files
tinypenguin-cli run "Configure /etc/hosts file to add local hostname mappings"

# Set permissions
tinypenguin-cli run "Set proper permissions on /var/www/html directory"
```

## Security Features

### Command Validation
- Blocks dangerous commands like `rm -rf /`
- Prevents unauthorized access to system files
- Validates command syntax before execution

### Approval System
- Requires approval for potentially risky operations
- Provides command preview before execution
- Allows users to deny unsafe operations

### Sandboxing
- Commands run with limited privileges
- Timeout enforcement prevents hanging processes
- Working directory restrictions

## Configuration

### Environment Variables
```bash
export TINYLLAMA_URL=http://localhost:11434/v1
export TINYLLAMA_MODEL=tinyllama
export TASK_TIMEOUT=30
```

### Configuration File
Create `~/.tinypenguin/config.json`:
```json
{
  "tinyllama_url": "http://localhost:11434/v1",
  "model": "tinyllama",
  "timeout": 30,
  "enable_sudo": false,
  "allowed_packages": ["nginx", "apache", "php"],
  "protected_paths": ["/etc", "/usr", "/boot"]
}
```

## API Integration

### TypeScript Client
```typescript
import { ToolExecutor } from './src/core/task/ToolExecutor';

const executor = new ToolExecutor();

// Edit a file
const result = await executor.executeEditFilesTool(
  '/etc/hostname',
  '',
  '--- original\n+++ new\n@@ -1 +1 @@\n-old-hostname\n+new-hostname'
);

// Run a command
const result = await executor.executeRunCommandTool(
  'systemctl status nginx',
  '/root',
  30,
  false
);
```

### Go Client
```go
import "github.com/anthropic-cli/tinypenguin/cli/pkg/cli"

err := cli.RunTask("Create a backup script")
if err != nil {
    log.Fatal(err)
}
```

## Troubleshooting

### Common Issues

1. **tinyllama Connection Failed**
   ```bash
   # Check if tinyllama is running
   curl http://localhost:11434/v1/models
   
   # Verify URL configuration
   tinypenguin-cli --url http://localhost:11434/v1 run "Test"
   ```

2. **Protobuf Generation Errors**
   ```bash
   # Check protoc version
   protoc --version
   
   # Regenerate protobuf files
   make clean
   make proto-gen
   ```

3. **Permission Denied**
   ```bash
   # Check file permissions
   ls -la bin/tinypenguin-cli
   
   # Run with appropriate privileges
   sudo ./bin/tinypenguin-cli run "sudo command"
   ```

### Debug Mode
```bash
# Enable verbose logging
TINYLLAMA_DEBUG=1 tinypenguin-cli run "Your query"
```

## Development

### Building from Source
```bash
# Install dependencies
make

# Run tests
make test

# Format code
make fmt

# Lint code
make lint
```

### Adding New Tools
1. Define tool in `proto/tinypenguin/common.proto`
2. Implement tool handler in TypeScript
3. Update CLI interface
4. Add tests

### Testing
```bash
# Unit tests
make test

# Integration tests
./scripts/test-integration.sh

# Manual testing
./bin/tinypenguin-cli run "Test query"
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- **Documentation**: [Project Wiki](https://github.com/anthropic-cli/tinypenguin/wiki)
- **Issues**: [GitHub Issues](https://github.com/anthropic-cli/tinypenguin/issues)
- **Discussions**: [GitHub Discussions](https://github.com/anthropic-cli/tinypenguin/discussions)

## Acknowledgments

- tinyllama team for the excellent LLM platform
- Red Hat for RHCSA certification standards
- Open source community for valuable tools and libraries