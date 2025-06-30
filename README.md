# Deep Coding Agent v1.0

A high-performance conversational AI coding assistant featuring dual-architecture design with both legacy and modern ReAct (Reasoning and Acting) agent systems. Built in Go for maximum performance, it provides a natural language interface for code analysis, file operations, and development tasks through an intelligent agent architecture with advanced tool calling capabilities, comprehensive security, and streaming responses.

## 🚀 Key Features

### 🧠 **Dual Agent Architecture**
- **ReAct Agent (Default)**: Modern Think-Act-Observe cycle with intelligent reasoning
- **Legacy Agent**: Stable fallback system with proven reliability
- **Automatic Selection**: Environment-based agent switching with graceful fallback
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

### 📝 **Template-Based Prompt System**
- **Markdown Templates**: Organized prompt templates in `/internal/prompts/templates/`
- **Variable Substitution**: Dynamic content generation with `{{variable}}` syntax
- **Section-Based Organization**: Modular prompt construction
- **Backward Compatibility**: Maintains legacy prompt support

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
├── cmd/main.go                     # Primary CLI executable
├── internal/                       # Core implementation modules
│   ├── core/                      # Domain-driven core components
│   │   ├── agent/                 # Dual agent architecture
│   │   │   ├── react_agent.go     # ReAct agent implementation
│   │   │   ├── agent.go           # Legacy agent implementation
│   │   │   └── factory.go         # Agent factory and configuration
│   │   ├── reasoning/             # ReAct reasoning engine
│   │   ├── planning/              # Action planning and execution
│   │   └── observation/           # Result analysis and learning
│   ├── tools/                     # Advanced tool system
│   │   ├── registry/              # Tool registry with metrics
│   │   ├── builtin/               # Core tool implementations
│   │   └── execution/             # Tool execution engine
│   ├── security/                  # Multi-layered security system
│   │   └── manager.go             # Security management and policies
│   ├── prompts/                   # Template-based prompt system
│   │   ├── renderer.go            # Prompt rendering engine
│   │   ├── builder.go             # High-level prompt builder
│   │   └── templates/             # Markdown prompt templates
│   ├── config/                    # Unified configuration management
│   ├── session/                   # Session management and persistence
│   ├── memory/                    # Memory and context management
│   ├── ai/                        # AI provider abstraction layer
│   └── cli/                       # CLI interface components
├── pkg/                           # Shared interfaces and types
│   ├── interfaces/                # Clean interface definitions
│   └── types/                     # Comprehensive type system
├── docs/                          # Extensive documentation
│   └── architecture/              # Architecture documentation
├── scripts/                       # Development automation scripts
├── tests/                         # Integration tests
├── benchmarks/                    # Performance benchmarking
└── CLAUDE.md                      # Development guidance for Claude Code
```

## 🔧 **Development**

### Testing
```bash
# Run all tests
go test ./...

# Test specific components
go test ./internal/core/agent/     # Core agent tests
go test ./internal/tools/          # Tool system tests
go test ./internal/security/       # Security tests
go test ./internal/prompts/        # Prompt system tests

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

### Optimizations
- **Zero Dependencies**: Core functionality uses only Go standard library
- **Concurrent Execution**: Parallel tool execution where safe (max 10 workers)
- **Memory Management**: Automatic session cleanup and message trimming
- **Caching**: Tool result caching and context preservation

### Metrics
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
- [ReAct Agent Implementation](docs/architecture/REACT_AGENT_IMPLEMENTATION.md) - Detailed implementation guide
- [System Architecture Analysis](docs/architecture/ARCHITECTURE_ANALYSIS_FINAL.md) - System design overview
- [Prompt System Design](docs/architecture/SYSTEM_PROMPTS_DESIGN.md) - Prompt engineering guide

### User Guides
- [Quick Start Guide](docs/QUICKSTART.md) - Getting started with the agent
- [Tool Development Guide](docs/TOOL_DEVELOPMENT_GUIDE.md) - Creating custom tools
- [API Reference](docs/API_REFERENCE.md) - Complete API documentation

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