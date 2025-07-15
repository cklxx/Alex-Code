# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Alex - 高性能普惠的软件工程助手 v1.0** is a high-performance AI software engineering assistant built in Go, featuring ReAct (Reasoning and Acting) agent architecture with advanced tool calling capabilities, streaming responses, and comprehensive security.

## Essential Development Commands

### Building and Testing
```bash
# Build Alex
make build                    # Builds ./alex binary

# Development workflow
make dev                      # Format, vet, build, and test functionality
make test                     # Run all tests
make fmt                      # Format Go code
make vet                      # Run go vet
```

### Alex Usage
```bash
# Interactive mode (ReAct agent)
./alex -i

# Single prompt mode
./alex "Analyze the current directory structure"

# Session management
./alex -r session_id -i       # Resume session
./alex session list           # List sessions

# Configuration
./alex config show            # Show configuration
./alex config set api_key sk-... # Set API key
```

## Architecture Overview

### Core Components

1. **ReAct Agent** (`internal/agent/react_agent.go`)
   - Think-Act-Observe cycle with streaming support
   - Centralized prompt management via `internal/prompts`
   - Session-based conversation management

2. **Tool System** (`internal/tools/`)
   - Built-in tools: file operations, shell execution, search, todos, web integration
   - Dynamic tool registry with concurrent execution
   - Security validation and sandboxing

3. **LLM Integration** (`internal/llm/`)
   - Multi-model support with factory pattern
   - HTTP and streaming client implementations
   - OpenAI-compatible tool calling format

4. **Session Management** (`internal/session/`)
   - File-based persistent storage (`~/.alex-sessions/`)
   - Context preservation and message history
   - Session-aware todo management

5. **Configuration** (`internal/config/`)
   - Multi-model configuration system
   - Default: OpenRouter + DeepSeek Chat V3
   - Environment variable overrides

### Built-in Tools
- **File Operations**: `file_read`, `file_update`, `file_replace`, `file_list`
- **Shell Execution**: `bash`, `script_runner` with security controls
- **Search**: `grep`, `ripgrep`, `find`
- **Todo Management**: Session-aware task tracking
- **Web Integration**: `web_search`

### Security Features
- Risk assessment engine with dynamic scoring
- Path protection for system directories
- Command safety detection
- Configurable tool restrictions

## Performance Characteristics

- **Target**: Sub-30ms execution times
- **Concurrency**: Up to 10 parallel tool executions
- **Memory**: <100MB baseline, <500MB peak
- **Storage**: File-based sessions with automatic cleanup

## Code Principles

### Core Design Philosophy

**保持简洁清晰，如无需求勿增实体，尤其禁止过度配置**

- **Simplicity First**: Always choose the simplest solution that works
- **Clear Intent**: Code should be self-documenting through clear naming
- **Minimal Configuration**: Avoid configuration options unless absolutely necessary
- **Purposeful Entities**: Only create new types/interfaces when they serve a clear purpose

### Naming Guidelines
- **Functions**: `AnalyzeCode()`, `LoadPrompts()`, `ExecuteTool()`
- **Types**: `ReactAgent`, `PromptLoader`, `ToolExecutor`
- **Variables**: `taskResult`, `userMessage`, `promptTemplate`

### Architectural Principles
1. **Single Responsibility**: Each component has one clear purpose
2. **Minimal Dependencies**: Reduce coupling between components
3. **Clear Interfaces**: Define simple, focused interfaces
4. **Error Handling**: Fail fast with clear error messages
5. **No Over-Engineering**: Don't build for theoretical future needs

## Current Status

### Production Ready:
- ✅ **ReAct Agent System**: Complete with streaming support
- ✅ **Multi-Model LLM System**: Advanced factory pattern
- ✅ **Tool System**: 8+ built-in tools with extensible registry
- ✅ **Configuration Management**: Multi-model configuration
- ✅ **Session Management**: File-based persistent storage
- ✅ **CLI Interface**: Interactive and single-prompt modes

### Performance:
- Go-based implementation for maximum performance
- 40-100x performance improvement over TypeScript predecessor
- Concurrent tool execution with dependency management
- Memory-efficient session management

### Recent Changes (2025-07):
- Session-aware todo system with context injection
- Enhanced project detection for Python, Node.js, Rust
- Simplified context system with ProjectSummary
- Unified prompt system via `internal/prompts`
- Continued code simplification

## Testing

```bash
# Test specific packages
go test ./internal/agent/             # ReAct agent system
go test ./internal/tools/builtin/     # Built-in tools
go test ./internal/llm/               # LLM integration
go test ./internal/session/           # Session management

# Coverage testing
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

This represents a mature, production-ready AI coding assistant with enterprise-grade architecture while maintaining simplicity and performance focus.