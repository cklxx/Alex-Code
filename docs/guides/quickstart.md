# Deep Coding Agent - Quick Start Guide

## Overview

Deep Coding Agent is a high-performance AI coding assistant that uses a unified ReAct (Reasoning and Acting) architecture with powerful tool calling capabilities. This guide will get you up and running quickly.

## Installation

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional, for using Makefile commands)

### Quick Install

```bash
# Clone the repository
git clone https://github.com/your-org/deep-coding-agent.git
cd deep-coding-agent

# Build the agent
make build

# Or build manually
go build -o deep-coding-agent ./cmd/main.go
```

### Using Install Script

```bash
# Run the installation script
./scripts/install.sh

# This will:
# - Install dependencies
# - Build the binary
# - Set up configuration
# - Optionally add to PATH
```

## Configuration

### Initial Setup

```bash
# Initialize with default configuration
./deep-coding-agent --init

# Or manually configure
./deep-coding-agent config --init
```

### AI Provider Configuration

Configure your AI provider (OpenAI or compatible):

```bash
# Set OpenAI API key
./deep-coding-agent config --set openaiApiKey=your-api-key-here

# Or set ARK API key (ByteDance alternative)
./deep-coding-agent config --set arkApiKey=your-ark-key-here

# Set custom API endpoint (optional)
./deep-coding-agent config --set apiBaseURL=https://your-custom-endpoint.com

# Set model (optional)
./deep-coding-agent config --set apiModel=gpt-4-turbo
```

### Basic Configuration Options

```bash
# Enable ReAct mode (recommended)
./deep-coding-agent config --set reactMode=true

# Set max iterations for complex tasks
./deep-coding-agent config --set reactMaxIterations=10

# Enable thinking process display
./deep-coding-agent config --set reactThinkingEnabled=true

# Configure allowed tools
./deep-coding-agent config --set allowedTools=file_read,file_write,file_list,bash
```

## Basic Usage

### Interactive Mode

Start an interactive session for conversational coding assistance:

```bash
# Start interactive mode
./deep-coding-agent -i

# In interactive mode, you can:
# - Ask questions about your code
# - Request file analysis
# - Get help with debugging
# - Generate code snippets
```

### Single Command Mode

Execute one-off commands:

```bash
# Analyze current directory
./deep-coding-agent "Analyze the project structure and provide insights"

# Read and explain a specific file
./deep-coding-agent "Read main.go and explain what it does"

# Get help with a specific task
./deep-coding-agent "How can I optimize this Go code for better performance?"
```

### JSON Output

Get structured responses for integration with other tools:

```bash
# JSON format output
./deep-coding-agent --format json "List all Go files in this project"

# Pipe to jq for processing
./deep-coding-agent --format json "Analyze code quality" | jq '.data.metrics'
```

## Common Use Cases

### 1. Project Analysis

```bash
# Analyze entire project structure
./deep-coding-agent -i
> "Analyze this project's architecture and suggest improvements"

# Check for code smells
> "Scan the codebase for potential issues and code smells"

# Review test coverage
> "Check test coverage and suggest areas that need more testing"
```

### 2. Code Generation

```bash
# Generate a new feature
> "Create a REST API handler for user authentication in Go"

# Generate tests
> "Generate unit tests for the user service in internal/user/service.go"

# Create documentation
> "Generate API documentation for the endpoints in main.go"
```

### 3. Debugging and Optimization

```bash
# Debug performance issues
> "Analyze the performance of this Go application and suggest optimizations"

# Find and fix bugs
> "Help me debug why this function is not working as expected"

# Code review
> "Review the code in internal/handlers/ and suggest improvements"
```

### 4. File Operations

```bash
# Read specific files
> "Read the README.md file and summarize the project"

# Search in files
> "Search for all TODO comments in Go files"

# List project files
> "List all files in the src directory recursively"
```

### 5. Git and Repository Management

```bash
# Check git status
> "Show me the current git status and any uncommitted changes"

# Analyze git history
> "Analyze recent commits and suggest a changelog entry"

# Help with git workflow
> "Help me create a proper commit message for these changes"
```

## Advanced Features

### ReAct Mode

ReAct (Reasoning and Acting) mode enables the agent to think through problems step by step:

```bash
# Enable ReAct mode with thinking display
./deep-coding-agent config --set reactMode=true --set reactThinkingEnabled=true

# The agent will show its thinking process:
# 🤔 Thinking: The user wants to analyze the project structure...
# 🛠️ Action: I'll use the file_list tool to see the directory structure
# 📋 Observation: Found 15 Go files and 3 directories...
# 🤔 Thinking: Based on the structure, this appears to be a web service...
```

### Tool Restrictions

Control which tools the agent can use for security:

```bash
# Restrict to read-only operations
ALLOWED_TOOLS="file_read,file_list" ./deep-coding-agent "Analyze the codebase"

# Allow file operations but not command execution
ALLOWED_TOOLS="file_read,file_write,file_list" ./deep-coding-agent "Help me refactor this code"
```

### Session Management

```bash
# Start named session
./deep-coding-agent --session-id myproject "Start working on the authentication feature"

# Resume previous session
./deep-coding-agent --session-id myproject "Continue with the authentication work"

# List active sessions
./deep-coding-agent --list-sessions
```

## Configuration Files

### Main Configuration

Location: `~/.deep-coding-config.json`

```json
{
  "aiProvider": "openai",
  "openaiApiKey": "your-key-here",
  "reactMode": true,
  "reactMaxIterations": 10,
  "allowedTools": ["file_read", "file_write", "file_list", "bash"],
  "maxTokens": 2000,
  "temperature": 0.3
}
```

### Tool Configuration

```json
{
  "toolsConfig": {
    "maxConcurrentExecutions": 5,
    "defaultTimeout": 30000,
    "securityConfig": {
      "enableSandbox": false,
      "allowedTools": ["file_read", "file_list", "file_write", "bash"],
      "restrictedTools": ["rm", "format", "dd"]
    }
  }
}
```

## Development Mode

For development and testing:

```bash
# Start with hot reload
./scripts/run.sh dev

# Run development workflow (format, vet, build, test)
make dev

# Run tests with coverage
make test

# Format code
make fmt
```

## Troubleshooting

### Common Issues

1. **API Key Not Configured**
   ```bash
   Error: OpenAI API key not configured
   Solution: ./deep-coding-agent config --set openaiApiKey=your-key
   ```

2. **Tool Execution Denied**
   ```bash
   Error: Tool 'bash' is restricted by security policy
   Solution: ./deep-coding-agent config --set allowedTools=file_read,file_list,bash
   ```

3. **File Permission Errors**
   ```bash
   Error: Permission denied accessing /restricted/path
   Solution: Check file permissions or use allowed directories
   ```

### Debug Mode

```bash
# Enable debug logging
./deep-coding-agent --debug "Analyze this file"

# Verbose output
./deep-coding-agent --verbose "Run comprehensive analysis"
```

### Check Configuration

```bash
# View current configuration
./deep-coding-agent config --show

# Validate configuration
./deep-coding-agent config --validate

# Reset to defaults
./deep-coding-agent config --reset
```

## Integration Examples

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Code Analysis
  run: |
    ./deep-coding-agent --format json "Analyze code quality and security" > analysis.json
    
- name: Generate Documentation
  run: |
    ./deep-coding-agent "Generate API documentation" > docs/api.md
```

### IDE Integration

```bash
# VS Code task example
{
  "label": "Deep Coding Analysis",
  "type": "shell",
  "command": "./deep-coding-agent",
  "args": ["Analyze current file and suggest improvements"],
  "group": "build"
}
```

### Git Hooks

```bash
# Pre-commit hook
#!/bin/sh
./deep-coding-agent "Review staged changes for potential issues"
```

## Performance Tips

1. **Use specific commands** instead of broad requests for faster responses
2. **Enable caching** for repeated operations: `--enable-cache`
3. **Limit tool scope** to only what you need: `ALLOWED_TOOLS="file_read,file_list"`
4. **Use JSON output** for programmatic processing: `--format json`
5. **Set reasonable timeouts** for long operations: `--timeout 60`

## Next Steps

- Read the [Architecture Documentation](AGENT_ARCHITECTURE.md) for deeper understanding
- Check the [API Reference](API_REFERENCE.md) for detailed tool documentation
- Explore [Tool Development Guide](TOOL_DEVELOPMENT_GUIDE.md) for creating custom tools
- Review [Security Best Practices](../internal/security/README.md) for production use

## Getting Help

```bash
# Built-in help
./deep-coding-agent --help

# Tool-specific help
./deep-coding-agent tools --help

# Configuration help
./deep-coding-agent config --help

# Interactive help
./deep-coding-agent -i
> "How do I configure the agent for my specific use case?"
```

---

*For more examples and detailed documentation, see the complete documentation in the `docs/` directory.*