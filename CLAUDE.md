# CLAUDE.md

## Project Overview

**Alex - 高性能普惠的软件工程助手 v1.0** is a production-ready AI software engineering assistant built in Go with ReAct agent architecture, MCP protocol integration, memory management, and SWE-Bench evaluation.

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
# Interactive mode (auto-detects TTY)
./alex                        # Auto-enters interactive mode
./alex -i                     # Explicit interactive mode

# Single prompt mode
./alex "Analyze the current directory structure"

# Session management
./alex -r session_id -i       # Resume session
./alex session list           # List sessions
./alex memory compress        # Compress session memory

# SWE-Bench evaluation
./alex run-batch --dataset.subset lite --workers 4 --output ./results

# Configuration
./alex config show            # Show configuration
```

## Architecture Overview

### Core Components

1. **ReAct Agent** (`internal/agent/`) - Think-Act-Observe cycle with streaming and memory
2. **MCP Protocol** (`internal/mcp/`) - Model Context Protocol with JSON-RPC 2.0
3. **Memory System** (`internal/memory/`, `internal/context/`) - Dual-layer with vector storage
4. **Tool System** (`internal/tools/`) - 12+ built-in tools with MCP integration
5. **LLM Integration** (`internal/llm/`) - Multi-model support with caching
6. **Session Management** (`internal/session/`) - Persistent storage with compression
7. **SWE-Bench** (`evaluation/swe_bench/`) - Evaluation system with parallel processing
8. **Configuration** (`internal/config/`) - Multi-model config (default: OpenRouter + DeepSeek)

### Built-in Tools
- **File**: `file_read`, `file_update`, `file_replace`, `file_list`
- **Shell**: `bash`, `code_executor` with sandbox controls
- **Search**: `grep`, `ripgrep`, `find`
- **Tasks**: `todo_create`, `todo_update`, `todo_list`
- **Web**: `web_search` with Tavily API
- **Reasoning**: `think`
- **MCP**: Dynamic external tool integration

### Security Features
- Risk assessment and path protection
- Command safety detection
- Configurable restrictions

## Performance
- Sub-30ms execution, 10 parallel tools, <100MB baseline memory
- File-based sessions with automatic cleanup

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

## Status
✅ Production ready with ReAct agent, MCP protocol, memory system, tools, SWE-Bench, caching, terminal UI, and security

## Testing

```bash
# Test packages
go test ./internal/agent/ ./internal/tools/builtin/ ./internal/llm/ ./internal/memory/ ./internal/mcp/ ./internal/session/ ./evaluation/swe_bench/

# Quick tests
make test-functionality
make test-working

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# SWE-Bench evaluation
./alex run-batch --dataset.subset lite --instance-limit 5 --workers 2
```