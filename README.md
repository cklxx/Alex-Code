# Alex - 高性能普惠的软件工程助手

**Alex** is a high-performance, universally accessible AI software engineering assistant built with advanced ReAct architecture. Designed for developers of all levels, Alex provides intelligent code analysis, automated development tasks, and seamless integration with modern development workflows.

## Quick Start

```bash
# Build Alex
make build

# Interactive mode - Start coding conversation
./alex -i

# Single command - Instant analysis
./alex "Analyze current directory structure"
```

## Core Features

**🧠 Intelligent Conversation**: Advanced Think-Act-Observe reasoning with streaming responses and persistent sessions  
**🛠 Rich Tool Ecosystem**: 20+ built-in tools for file operations, search, web integration, and development tasks  
**🌐 Multi-Model Support**: Seamless integration with OpenAI, DeepSeek, and other leading LLM providers  
**🔒 Security-First Design**: Enterprise-grade risk assessment, command detection, and path protection  
**⚡ High Performance**: Native Go implementation with concurrent execution and sub-30ms response times  
**🎯 Universal Access**: Designed for developers at all skill levels - from beginners to experts

## Usage

### Interactive Mode - Your AI Coding Partner
```bash
./alex -i
```

### Configuration Management
```bash
./alex config set api_key your-key    # Set API key
./alex config show                    # View current settings
./alex config validate               # Validate configuration
```

### Session Management - Persistent Conversations
```bash
./alex -r session_id -i              # Resume previous session
./alex -ls                           # List all sessions
./alex -stream "Complex analysis"    # Enable streaming responses
```

## Available Tools

**File Operations**: `file_read`, `file_write`, `file_list`, `directory_create`  
**Shell Execution**: `bash`, `script_runner`, `process_monitor`  
**Search Tools**: `grep`, `ripgrep`, `find`  
**Task Management**: `todo_read`, `todo_update`  
**Web Integration**: `web_search`

## Project Architecture

```
alex/
├── cmd/                    # CLI entry points and command handlers
├── internal/
│   ├── agent/             # Advanced ReAct agent system
│   ├── llm/               # Multi-model LLM integration layer
│   ├── tools/             # Comprehensive tool ecosystem
│   ├── prompts/           # AI prompt templates and management
│   ├── config/            # Configuration and settings management
│   └── session/           # Persistent session management
├── pkg/types/             # Core type definitions and interfaces
├── docs/                  # Comprehensive documentation
├── scripts/               # Development and automation scripts
└── benchmarks/            # Performance testing and benchmarks
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

Alex stores configuration in: `~/.alex-config.json`

```json
{
    "api_key": "sk-or-xxx",
    "base_url": "https://openrouter.ai/api/v1", 
    "model": "deepseek/deepseek-chat-v3-0324:free",
    "max_tokens": 4000,
    "temperature": 0.7,
    "max_turns": 25
}
```

Environment variables:
```bash
export OPENAI_API_KEY="your-key"
export ALLOWED_TOOLS="file_read,bash"
```

## Why Alex Excels

**🚀 Blazing Fast Performance**
- **Zero Dependencies**: Built on Go standard library for maximum reliability
- **Concurrent Execution**: Intelligent parallel processing for complex tasks
- **Memory Efficient**: Automatic session cleanup and resource management
- **Lightning Speed**: Most operations complete in <30ms
- **Proven Performance**: 40-100x faster than comparable implementations

**🎯 Designed for Universal Access**
- **Beginner Friendly**: Natural language interface requires no special syntax
- **Expert Powerful**: Advanced features for complex development workflows
- **Cross-Platform**: Works seamlessly on macOS, Linux, and Windows
- **Lightweight**: Minimal resource usage, runs on any modern machine

## Documentation

- **[Software Engineering Roles Analysis](docs/software-engineering-roles-analysis.md)**: Comprehensive analysis of roles and responsibilities across software engineering phases (2024)

## License

MIT License