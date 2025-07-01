# Deep Coding Agent

High-performance AI coding assistant built with ReAct architecture, providing natural language interface for code analysis, file operations, and development tasks.

## Quick Start

```bash
# Build
make build

# Interactive mode
./deep-coding-agent -i

# Single command
./deep-coding-agent "Analyze current directory structure"
```

## Core Features

**Intelligent Conversation**: Think-Act-Observe cycle with streaming responses and persistent sessions  
**Tool System**: 20+ built-in tools for file operations, search, and web integration  
**Multi-Model Support**: OpenAI, DeepSeek, and other LLM providers  
**Security Design**: Risk assessment, command detection, and path protection  
**High Performance**: Go implementation with concurrent execution and sub-30ms response times

## Usage

### Interactive Mode
```bash
./deep-coding-agent -i
```

### Configuration
```bash
./deep-coding-agent config set api_key your-key
./deep-coding-agent config show
```

### Session Management
```bash
./deep-coding-agent -r session_id -i  # Resume session
./deep-coding-agent -ls               # List sessions
```

## Available Tools

**File Operations**: `file_read`, `file_write`, `file_list`, `directory_create`  
**Shell Execution**: `bash`, `script_runner`, `process_monitor`  
**Search Tools**: `grep`, `ripgrep`, `find`  
**Task Management**: `todo_read`, `todo_update`  
**Web Integration**: `web_search`

## Project Structure

```
deep-coding/
├── cmd/                    # CLI entry points
├── internal/
│   ├── agent/             # ReAct agent system
│   ├── llm/               # Multi-model LLM integration
│   ├── tools/             # Tool system
│   ├── prompts/           # Prompt templates
│   ├── config/            # Configuration management
│   └── session/           # Session management
├── pkg/types/             # Type definitions
├── docs/                  # Documentation
├── scripts/               # Development scripts
└── benchmarks/            # Performance benchmarks
```

## Development

```bash
# Development workflow
make dev                   # Format, check, build, test

# Testing
go test ./...              # All tests
go test ./internal/agent/  # Specific package

# Hot reload development
./scripts/run.sh dev
```

## Configuration

Default config file: `~/.deep-coding-config.json`

```json
{
    "api_key": "sk-or-xxx",
    "base_url": "https://openrouter.ai/api/v1", 
    "model": "deepseek/deepseek-chat-v3-0324:free",
    "max_tokens": 4000,
    "temperature": 0.7,
    "max_turns": 25,
}
```

Environment variables:
```bash
export OPENAI_API_KEY="your-key"
export ALLOWED_TOOLS="file_read,bash"
```

## Performance

- **Zero Dependencies**: Uses only Go standard library
- **Concurrent Execution**: Intelligent parallel tool execution
- **Memory Management**: Automatic session cleanup
- **Response Speed**: Most operations complete in <30ms
- **Performance Gain**: 40-100x faster than predecessor implementations

## Documentation

- **[Software Engineering Roles Analysis](docs/software-engineering-roles-analysis.md)**: Comprehensive analysis of roles and responsibilities across software engineering phases (2024)

## License

MIT License