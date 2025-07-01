# Deep Coding Agent v1.0

A high-performance conversational AI coding assistant featuring streamlined ReAct (Reasoning and Acting) agent architecture. Built in Go for maximum performance, it provides a natural language interface for code analysis, file operations, and development tasks through an intelligent agent architecture with advanced tool calling capabilities, comprehensive security, and streaming responses.

> **Latest Update (2025-01-01)**: Major architecture simplification - removed over-engineered components, centralized prompt system, and streamlined type definitions for better maintainability and performance.

## 🚀 Key Features

### 🧠 **Streamlined ReAct Architecture**
- **ReAct Agent**: Modern Think-Act-Observe cycle with intelligent reasoning
- **Unified Design**: Simplified architecture focused on essential functionality
- **Multi-Model LLM**: Factory pattern supporting multiple AI providers
- **40-100x Performance**: Significant improvement over TypeScript predecessor

### 🛠️ **Advanced Tool System**
- **29+ Built-in Tools**: File operations, shell execution, search, web integration
- **Tool Orchestration**: Intelligent multi-tool workflow coordination
- **Security Validation**: Comprehensive parameter and execution validation
- **Performance Monitoring**: Usage statistics, error tracking, execution metrics
- **Dynamic Registration**: Runtime tool discovery and registration

### 🔒 **Multi-layered Security**
- **Risk Assessment Engine**: Dynamic risk scoring with threat detection
- **Behavior Analysis**: User and tool usage profiling for security insights
- **Path Protection**: Restricted access to system directories
- **Command Safety**: Detection of privilege escalation and destructive commands
- **Audit Logging**: Comprehensive security event tracking

### 💬 **Interactive CLI Experience**
- **Conversational Interface**: Natural language interaction with streaming responses
- **Session Management**: Persistent conversations with resume capability
- **Tool Integration**: Seamless tool execution within conversations
- **Multiple Modes**: Interactive, single-prompt, and batch processing

### 📝 **Centralized Prompt System**
- **Markdown Templates**: Unified prompt templates in `/internal/prompts/`
- **Centralized Loading**: Single prompt loader with fallback support
- **Template-Based**: ReAct thinking template with embedded instructions
- **Reliability**: Multiple fallback layers for robust operation

## 📋 **Available Tools**

### File Operations
- `file_read` - Read file contents
- `file_write` - Write content to files  
- `file_list` - List files and directories with recursive support
- `file_update` - Update existing files
- `file_replace` - Replace file contents
- `directory_create` - Create directories

### Shell & Execution
- `bash` - Execute shell commands with security controls
- `script_runner` - Run scripts with enhanced features
- `process_monitor` - Monitor running processes

### Search & Analysis
- `grep` - Search for patterns in files
- `ripgrep` - Fast search using ripgrep
- `find` - Find files by name or properties

### Task Management
- `todo_read` - Read task lists and todos
- `todo_update` - Manage tasks and todos

### Web Integration
- `web_search` - Search the web for information

## 🏗️ **Architecture Overview**

### ReAct Agent System (Default)
```
User Input → Think Phase → Act Phase → Observe Phase → [Loop until complete]
```

**Features:**
- **Three-Phase Processing**: Think → Act → Observe cycle
- **Streaming Responses**: Real-time response display with enhanced formatting
- **Confidence-Based Completion**: Intelligent task completion evaluation
- **Memory Integration**: Learning extraction and context preservation
- **Tool Orchestration**: Advanced multi-tool workflow coordination

### Legacy Agent System (Fallback)
- **Session-based Conversation**: Persistent conversation management
- **Tool Calling**: Dynamic tool execution with result processing
- **Concurrent Execution**: Worker pools with semaphore-controlled concurrency
- **Performance Optimized**: Sub-30ms execution times for most operations

## 🚀 **Installation**

### Quick Start
```bash
# Clone the repository
git clone <repository-url>
cd deep-coding

# Build the agent
make build

# Run interactively
./deep-coding-agent -i
```

### Build Options
```bash
# Development workflow
make dev                    # Format, vet, build, and test
make test                   # Run all tests
make fmt                    # Format Go code
make vet                    # Run go vet

# Using scripts for advanced workflows
./scripts/dev.sh dev        # Complete development workflow with setup
./scripts/test.sh all       # Comprehensive test suite
./scripts/run.sh dev        # Hot reload development with Air
```

### Docker Development
```bash
./scripts/docker.sh dev     # Start development environment
./scripts/docker.sh test    # Run tests in container
./scripts/docker.sh build   # Build Docker images
```

## 💻 **Usage**

### Interactive Mode (Recommended)
```bash
# Start interactive conversation with ReAct agent
./deep-coding-agent -i

# Force specific agent mode
USE_LEGACY_AGENT=true ./deep-coding-agent -i
USE_REACT_AGENT=true ./deep-coding-agent -i
```

### Single Prompt Mode
```bash
# Analyze directory structure
./deep-coding-agent "Analyze the current directory structure"

# Generate with specific output format
./deep-coding-agent --format json "List all Go files"

# With tool restrictions
ALLOWED_TOOLS="file_read,file_list" ./deep-coding-agent "Show project structure"
```

### Session Management
```bash
# Resume specific session
./deep-coding-agent --resume session_id

# Continue last session
./deep-coding-agent --continue

# List all sessions
./deep-coding-agent --list-sessions
```

### Example Conversations

**Code Analysis:**
```
You: "Analyze the Go files in this project and identify any potential issues"

Agent: I'll systematically analyze your Go project for potential issues.

<|FunctionCallBegin|>
[
  {"name": "file_list", "parameters": {"path": ".", "recursive": true, "file_types": [".go"]}},
  {"name": "todo_update", "parameters": {"action": "create", "content": "Analyze Go files for potential issues", "priority": "high"}}
]
<|FunctionCallEnd|>

[Analysis results with specific findings and recommendations]
```

**Development Tasks:**
```
You: "Help me implement a new REST API endpoint for user authentication"

Agent: I'll help you implement a user authentication REST API endpoint. Let me break this down into actionable steps.

<|FunctionCallBegin|>
[
  {"name": "todo_update", "parameters": {"action": "create_batch", "tasks": [
    {"content": "Design authentication endpoint structure", "priority": "high"},
    {"content": "Implement handler function with security best practices", "priority": "high"},
    {"content": "Add input validation and error handling", "priority": "medium"},
    {"content": "Write unit tests for the endpoint", "priority": "medium"}
  ]}}
]
<|FunctionCallEnd|>

[Step-by-step implementation guidance]
```

## ⚙️ **Configuration**

### Configuration File
The agent uses `~/.deep-coding-config.json` for persistent configuration:

```json
{
  "aiProvider": "openai",
  "openaiApiKey": "your-api-key",
  "arkApiKey": "alternative-api-key",
  "allowedTools": ["file_read", "file_write", "bash"],
  "maxTokens": 4000,
  "temperature": 0.7,
  "streamResponse": true,
  "sessionTimeout": 30,
  "interactive": true,
  "enableSandbox": true,
  "maxConcurrentTools": 5,
  "toolExecutionTimeout": 30,
  "reactMaxIterations": 10,
  "reactThinkingEnabled": true
}
```

### Environment Variables
```bash
# API Keys
export OPENAI_API_KEY="your-openai-key"
export ARK_API_KEY="your-ark-key"

# Agent Control
export USE_REACT_AGENT=true
export USE_LEGACY_AGENT=false

# Tool Restrictions
export ALLOWED_TOOLS="file_read,file_list,bash"
export RESTRICTED_TOOLS="file_delete,directory_delete"
```

## 🏛️ **Project Structure**

```
deep-coding/
├── cmd/                           # CLI application entry points
│   ├── main.go                    # Primary CLI executable
│   └── config.go                  # Configuration management commands
├── internal/                      # Core implementation modules
│   ├── agent/                     # Simplified agent architecture
│   │   ├── react_agent.go         # ReAct agent implementation
│   │   ├── core.go                # Core agent functionality
│   │   ├── code_executor.go       # Code execution engine
│   │   └── tool_executor.go       # Tool execution management
│   ├── llm/                       # LLM integration layer
│   │   ├── factory.go             # Multi-model LLM factory
│   │   ├── http_client.go         # HTTP-based LLM client
│   │   ├── streaming_client.go    # Streaming response client
│   │   ├── interfaces.go          # LLM interfaces
│   │   └── types.go               # LLM type definitions
│   ├── tools/                     # Advanced tool system
│   │   ├── registry/              # Tool registry with metrics
│   │   │   ├── registry.go        # Core registry implementation
│   │   │   └── configurator.go    # Tool configuration management
│   │   └── builtin/               # Built-in tool implementations
│   │       ├── file_operations.go # File I/O operations
│   │       ├── shell_tools.go     # Shell command execution
│   │       ├── search_tools.go    # Search and grep tools
│   │       ├── todo_tools.go      # Task management tools
│   │       ├── think_tools.go     # Thinking and reasoning tools
│   │       └── web_search_tools.go# Web search integration
│   ├── prompts/                   # Centralized prompt system
│   │   ├── loader.go              # Prompt loading and management
│   │   └── react_thinking.md      # ReAct thinking template
│   ├── config/                    # Configuration management
│   │   └── manager.go             # Unified config manager
│   └── session/                   # Session management
│       └── session.go             # Persistent session storage
├── pkg/                           # Shared interfaces and types
│   └── types/                     # Core type definitions
│       ├── types.go               # Primary type system
│       └── core.go                # Core domain types
├── docs/                          # Comprehensive documentation
│   ├── guides/                    # User guides and tutorials
│   ├── reference/                 # API reference documentation
│   ├── research/                  # Architecture research and analysis
│   └── codeact/                   # CodeAct implementation docs
├── scripts/                       # Development automation scripts
│   ├── dev.sh                     # Development workflow
│   ├── test.sh                    # Testing automation
│   ├── run.sh                     # Hot reload development
│   └── docker.sh                  # Docker workflow management
├── benchmarks/                    # Performance benchmarking framework
│   ├── human-eval/                # HumanEval benchmark integration
│   ├── evalplus/                  # EvalPlus benchmark suite
│   └── google-research/           # Google Research benchmarks
├── changelog/                     # Structured project evolution tracking
│   ├── 001-changelog-system-setup.md    # Changelog system establishment
│   └── 002-prompt-templates-documentation-update.md # Prompt system updates
└── CLAUDE.md                      # Development guidance for Claude Code
```

## 🔧 **Development**

### Testing
```bash
# Run all tests
go test ./...

# Test specific components
go test ./internal/agent/          # Agent system tests
go test ./internal/llm/            # LLM integration tests
go test ./internal/tools/          # Tool system tests
go test ./internal/prompts/        # Prompt system tests
go test ./internal/config/         # Configuration tests

# Coverage testing
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality
```bash
# Format code
make fmt

# Static analysis
make vet

# Comprehensive development workflow
make dev
```

### Performance Benchmarking
```bash
# Run benchmarks
go test -bench=. ./benchmarks/

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof
```

## 📊 **Performance Characteristics**

### Recent Optimizations (2025-01-01)
- **Simplified Architecture**: Removed over-engineered components, focused on essential functionality  
- **Centralized Prompts**: Unified prompt system with markdown templates in `/internal/prompts/`
- **Streamlined Types**: Consolidated type system, removed unused complexity
- **Enhanced LLM Integration**: Multi-model factory pattern with streaming support

### Core Optimizations
- **Zero Dependencies**: Core functionality uses only Go standard library
- **Concurrent Execution**: Parallel tool execution where safe (max 10 workers)
- **Memory Management**: Automatic session cleanup and message trimming
- **Caching**: Tool result caching and context preservation

### Performance Metrics
- **Target Performance**: Sub-30ms execution times for most operations
- **Performance Improvement**: 40-100x faster than TypeScript predecessor
- **Concurrency**: Configurable limits (default: 5 concurrent tools)
- **Memory Efficiency**: Memory-efficient session management with cleanup

## 🔒 **Security Features**

### Risk Assessment Engine
- **Tool Complexity Scoring**: Dynamic risk assessment based on tool capabilities
- **Path Sensitivity Analysis**: System path protection and access control
- **Command Danger Detection**: Identification of potentially destructive operations
- **User Risk History**: Behavioral analysis and risk profiling

### Security Policies
- **System Protection**: Restricted access to critical system directories
- **Command Safety**: Pattern-based detection of dangerous commands
- **Parameter Validation**: Input sanitization and validation
- **Audit Logging**: Comprehensive security event tracking

## 📚 **Documentation**

### Architecture Documentation
- [Architecture Overview](docs/01-architecture-overview.md) - System design overview
- [ReAct Agent Design](docs/02-react-agent-design.md) - ReAct implementation guide
- [Prompt System](docs/03-prompt-system.md) - Prompt management and templates
- [CodeAct Research](docs/deep-research-code-act-best-practices-2025.md) - Latest research on code generation

### User Guides
- [Quick Start Guide](docs/guides/quickstart.md) - Getting started with the agent
- [Tool Development Guide](docs/guides/tool-development.md) - Creating custom tools
- [API Reference](docs/reference/api-reference.md) - Complete API documentation

## 🤝 **Contributing**

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes with tests
4. Run the test suite: `make test`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Submit a pull request

### Development Guidelines
- Follow the existing code style and patterns
- Add tests for new functionality
- Update documentation for API changes
- Ensure security best practices are followed
- Performance considerations for new features

## 📄 **License**

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 **Acknowledgments**

- Built with Go for maximum performance and reliability
- Inspired by the ReAct (Reasoning and Acting) pattern for AI agents
- Designed for developer productivity and AI-assisted coding workflows

---

**Deep Coding Agent v1.0** - *Intelligent. Secure. Performant.*