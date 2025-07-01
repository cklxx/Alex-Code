# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Deep Coding Agent v1.0** is a high-performance conversational AI coding assistant featuring dual-architecture design with both legacy and modern ReAct (Reasoning and Acting) agent systems. Built in Go for maximum performance, it provides a natural language interface for code analysis, file operations, and development tasks through an intelligent agent architecture with advanced tool calling capabilities, comprehensive security, and streaming responses.

## Essential Development Commands

**⚠️ Recent Build Fix (2024-06-29)**: The Makefile has been updated to build the entire `cmd` package instead of just `cmd/main.go`, fixing compilation issues with the `handleConfigCommand` function. The project now builds successfully.

### Building and Testing
```bash
# Build the agent
make build                    # Builds ./deep-coding-agent binary

# Development workflow
make dev                      # Format, vet, build, and test functionality
make test                     # Run all tests
make fmt                      # Format Go code
make vet                      # Run go vet

# Using scripts for advanced workflows
./scripts/dev.sh dev          # Complete development workflow with setup
./scripts/test.sh all         # Comprehensive test suite (unit, integration, performance)
./scripts/run.sh dev          # Hot reload development with Air
```

### Agent Usage
```bash
# Interactive conversational mode (ReAct agent by default)
./deep-coding-agent -i

# Single prompt mode
./deep-coding-agent "Analyze the current directory structure"

# With streaming responses (default)
./deep-coding-agent -stream "List all Go files"

# Configure model parameters
./deep-coding-agent -tokens 4000 -temp 0.8 "Complex analysis task"

# Session management
./deep-coding-agent -r session_id -i       # Resume specific session
./deep-coding-agent -ls                    # List all sessions

# Configuration management
./deep-coding-agent config show            # Show current configuration
./deep-coding-agent config set api_key sk-... # Set API key
./deep-coding-agent config list            # List configuration keys
./deep-coding-agent config validate        # Validate configuration
```

### Testing Individual Components
```bash
# Run tests for specific package
go test ./internal/analyzer/         # Test analyzer package
go test -v ./internal/ai/            # Verbose tests for AI package
go test -run TestAnalyzeFile         # Run specific test
go test ./internal/agent/core/       # Test ReAct agent components
go test ./internal/tools/            # Test tool system
go test ./internal/security/         # Test security components

# Coverage testing
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Integration testing
go test -v ./internal/agent/core/ -count=1    # ReAct integration tests
```

### Docker Development
```bash
./scripts/docker.sh dev       # Start development environment
./scripts/docker.sh test      # Run tests in container
./scripts/docker.sh build     # Build Docker images
```

## Architecture Overview

### Dual Agent Architecture

**Architecture Selection Strategy:**
- **Default Mode**: ReAct agent for intelligent task processing (as of v1.0)
- **Legacy Mode**: Available via `USE_LEGACY_AGENT=true` for stability
- **Explicit ReAct**: Force via `USE_REACT_AGENT=true`
- **Fallback System**: Automatic fallback to legacy on ReAct initialization failure

### ReAct Agent Architecture (Default - `internal/agent/`)

**Modern conversational AI with Reasoning and Acting capabilities:**

**Core Components:**
1. **ReactAgent** (`internal/agent/react_agent.go`)
   - Unified ReAct Think-Act-Observe cycle with streaming support
   - Centralized prompt loading through `internal/prompts`
   - Session-based conversation management
   - Simplified tool orchestration with clear separation of concerns

2. **Prompt Management** (`internal/prompts/`)
   - Embedded markdown-based prompt templates
   - Unified prompt loading and rendering system
   - Template variable substitution
   - Fallback prompt support for reliability

3. **ReAct Core Components**:
   - **ReactCore**: Unified task solving logic (streaming/non-streaming)
   - **ThinkingEngine**: Centralized reasoning and analysis
   - **ToolExecutor**: Simplified tool execution management

4. **Multi-Model LLM System** (`internal/llm/`)
   - Factory pattern for dynamic model creation
   - HTTP and streaming client implementations
   - Model type selection (BasicModel, ReasoningModel)

5. **Session Management** (`internal/session/`)
   - Persistent file-based session storage
   - Message history preservation and restoration

**ReAct Processing Flow:**
```
User Input → Think Phase → Act Phase → Observe Phase → [Loop until complete]
```

**Key Features:**
- **Unified Processing**: Single ReactCore handles both streaming and non-streaming tasks
- **Centralized Prompts**: All prompts loaded from `internal/prompts` markdown files
- **Simplified Architecture**: Clear separation between thinking, acting, and observing
- **Reliable Fallbacks**: Multiple prompt fallback layers ensure system stability
- **Tool Orchestration**: Streamlined tool execution through dedicated ToolExecutor

### Configuration Management (`internal/config/`)

**Advanced multi-model configuration system:**

**Key Features:**
- **Multi-Model Support**: Different configurations for BasicModel and ReasoningModel
- **File-based Storage**: Configuration stored in `~/.deep-coding-config.json`
- **Backward Compatibility**: Legacy single-model configuration support
- **Environment Integration**: Environment variable overrides
- **Validation System**: Configuration validation with sensible defaults

**Default Configuration:**
- **API Provider**: OpenRouter (https://openrouter.ai/api/v1)
- **Default Model**: DeepSeek Chat V3 (`deepseek/deepseek-chat-v3-0324:free`)
- **Model Types**: BasicModel and ReasoningModel with appropriate parameters

### Advanced Tool System (`internal/tools/`)

**Domain-Driven Tool Management Architecture:**
```
internal/tools/
├── registry/ (Tool Registry and Management)
│   ├── registry.go (Enhanced Registry with Metrics & Recommendations)
│   ├── configurator.go (Tool Configuration Management)
│   └── registry_test.go (Comprehensive Testing)
├── builtin/ (Core Tool Implementations)
│   ├── file_operations.go (Read, Update, Replace, List, Directory Creation)
│   ├── shell_tools.go (Bash, Script Runner, Process Monitor)
│   ├── search_tools.go (Grep, Ripgrep, Find with Advanced Search)
│   ├── todo_tools.go (Task Management Tools)
│   ├── web_search_tools.go (Web Search Integration)
│   └── types.go (Tool interfaces and result types)
├── execution/ (Tool Execution Engine)
│   ├── tool_adapter.go (Advanced Tool System Adapter)
│   ├── Tool Recommendation Engine
│   ├── Execution Plan Optimization
│   ├── Performance Metrics Tracking
│   └── Dependency Analysis
└── configurator.go (External Tool Configuration)
```

**Enhanced Tool Capabilities:**
- **Intelligent Recommendations**: Task-aware tool suggestions with confidence scoring
- **Execution Planning**: Dependency analysis and optimization strategies
- **Performance Monitoring**: Usage statistics, error tracking, execution metrics
- **Security Validation**: Comprehensive parameter and execution validation
- **Concurrent Execution**: Optimized parallel/sequential execution based on dependencies

**Built-in Tools:**
- **File Operations**: `file_read`, `file_update`, `file_replace`, `file_list`, `directory_create`
- **Shell Execution**: `bash`, `script_runner`, `process_monitor` with security controls
- **Search Tools**: `grep`, `ripgrep`, `find` with flexible pattern matching
- **Todo Management**: Task creation, reading, status management

**Tool System Features:**
- **Dynamic Registration**: Runtime tool discovery and registration
- **Schema Validation**: JSON Schema-based parameter validation
- **Execution Context**: Sandbox support, timeout controls, permission management
- **Category Organization**: Tools grouped by functionality (file, execution, search, analysis)
- **Wrapper Pattern**: Interface compatibility between different tool implementations

### Security Architecture (`internal/security/`)

**Multi-layered Security System with Threat Detection:**

**Core Security Components:**
- **Security Manager**: Policy-based access control and risk assessment
- **Threat Detection**: Pattern-based threat identification with anomaly detection
- **Behavior Analysis**: User and tool usage profiling for security insights
- **Audit Logging**: Comprehensive security event tracking

**Security Features:**
- **Risk Assessment Engine**: Dynamic risk scoring based on:
  - Tool complexity (bash: 0.8, file_delete: 0.6, file_read: 0.2)
  - Path sensitivity (system paths: 1.0, executables: 0.7)
  - Command danger detection (destructive patterns: 1.0)
- **Path Protection**: Restricted access to system directories
- **Command Safety**: Detection of privilege escalation and destructive commands
- **Parameter Validation**: Input sanitization and validation
- **Tool Restrictions**: Configurable allowed/denied tool lists

### Configuration Management (`internal/config/`)

**Unified Configuration System:**
- **Legacy Manager**: Traditional key-value configuration
- **Unified Manager**: ReAct, tools, memory, and context configuration
- **JSON-based Storage**: `~/.deep-coding-config.json`
- **Environment Integration**: Override support for API keys and tool restrictions
- **Default Fallbacks**: Sensible defaults for all configuration options

**Configuration Scopes:**
- AI Provider settings (OpenAI, ARK API, mock provider)
- ReAct agent parameters (max turns, confidence thresholds)
- Tool system configuration (timeouts, concurrency limits)
- Security policies and restrictions
- Memory and context management settings

### Session Management (`internal/session/`)

**Persistent Session System:**
- **File-based Storage**: `~/.deep-coding-sessions/`
- **JSON Serialization**: Full conversation history preservation
- **Memory Management**: Automatic cleanup and message trimming
- **Context Preservation**: Working directory and session metadata
- **Multi-session Support**: Session listing, resumption, and partial ID matching

**Session Features:**
- Thread-safe message handling
- Configurable message limits (default: 1000)
- Automatic cleanup (configurable retention: 7 days)
- Session restoration with error handling

### CLI Interface (`cmd/main.go` + `cmd/config.go`)

**Rich Command-Line Interface with configuration management:**

**CLI Components:**
- **Main CLI** (`cmd/main.go`): Primary application entry point with interactive and single-prompt modes
- **Configuration CLI** (`cmd/config.go`): Advanced configuration management commands

**CLI Features:**
- **Interactive Mode**: Full conversational interface with slash commands
- **Single Prompt Mode**: One-shot task execution
- **Session Management**: Resume, continue, list sessions
- **Streaming Support**: Real-time response display with enhanced formatting
- **Output Formats**: Text, JSON, streaming

**Advanced CLI Features:**
- Graceful shutdown handling
- Progress indicators for tool execution
- Enhanced tool result formatting
- Session continuity with partial ID matching
- Comprehensive help system

### Development Infrastructure

**Recently Added Infrastructure Components:**

**CI/CD Pipeline** (`.github/workflows/`)
- `ci.yml`: Continuous integration with Go testing and linting
- `release.yml`: Automated release pipeline with multi-platform builds
- `security.yml`: Security scanning and vulnerability checks

**Docker Development** 
- `Dockerfile`: Production container with multi-stage build
- `Dockerfile.dev`: Development container with hot reload
- `docker-compose.yml`: Complete development environment
- `.air.toml`: Hot reload configuration for development

**Development Scripts** (`scripts/`)
- `dev.sh`: Complete development workflow setup
- `test.sh`: Comprehensive testing suite (unit, integration, performance)
- `run.sh`: Hot reload development server
- `docker.sh`: Docker workflow management
- `install.sh` & `release.sh`: Installation and release automation

**Documentation & Testing**
- `docs/`: Comprehensive project documentation
- `benchmarks/`: Performance benchmarking framework
- `tests/`: Integration and end-to-end testing infrastructure

**Changelog Management** (`changelog/`)
- Structured changelog system for tracking project evolution
- Sequential numbering system (001, 002, 003...)
- Each changelog entry includes modification timestamp
- Standardized format for consistent documentation

## Key Technical Patterns

### Interface-Driven Design
- Comprehensive interface definitions in `pkg/interfaces/`
- Clean separation between legacy and unified architectures
- Extensible tool system with common interfaces

### Type-Safe Architecture
- Rich type definitions in `pkg/types/`
- Strong typing for all core concepts (tasks, tools, sessions, security)
- JSON schema validation throughout

### Concurrent Processing
- Worker pool patterns for tool execution
- Semaphore-controlled concurrency (max 10 workers)
- Performance optimizations for large-scale operations

### Error Handling & Resilience
- Comprehensive error propagation
- Retry mechanisms for transient failures
- Graceful degradation and fallback systems

### Security-First Design
- Defense-in-depth security model
- Proactive threat detection
- Configurable security policies

## Performance Characteristics

### Optimizations:
- **Zero Dependencies**: Core functionality uses only Go standard library
- **Concurrent Execution**: Parallel tool execution where safe
- **Memory Management**: Automatic session cleanup and message trimming
- **Caching**: Tool result caching and context preservation

### Metrics:
- Target execution times: sub-30ms for most operations
- 40-100x performance improvement over TypeScript predecessor
- Configurable concurrency limits (default: 5 concurrent tools)
- Memory-efficient session management

## Version Migration Notes

### Unified Prompt System (v1.0.1+)

**ENHANCEMENT**: Centralized prompt management through `internal/prompts` module.

#### What Changed:
- **Centralized**: All prompts now loaded from markdown files in `internal/prompts/`
- **Eliminated**: Hardcoded prompt templates in ReactAgent code
- **Enhanced**: Structured prompt templates with variable substitution
- **Reliable**: Multiple fallback layers for prompt loading failures

#### Prompt Templates Available:
- `react_thinking.md`: Main ReAct reasoning instructions with tool execution strategy and task handling guidelines

#### Benefits of Unified System:
- **Maintainability**: Prompts managed as separate markdown files
- **Consistency**: Single source of truth for all prompt templates
- **Flexibility**: Easy to update prompts without code changes
- **Reliability**: Graceful fallback when prompts fail to load

### Tool Calling Format Standardization (v1.0+)

**STANDARD**: OpenAI-compatible tool calling format used throughout.

#### What Changed:
- **Removed**: Legacy text-based format `Action: tool_name(arg1=value1, arg2=value2)`
- **Standard**: Only OpenAI-compatible `tool_calls` and `function_call` formats supported
- **Enhanced**: Full parallel tool execution with proper error handling

#### Benefits of Standardization:
- **Industry Compatibility**: Follows OpenAI API standards adopted industry-wide
- **Parallel Execution**: Multiple tools can execute simultaneously
- **Better Error Handling**: Structured error reporting with tool call IDs
- **Reduced Maintenance**: Single, well-tested code path
- **Future-Proof**: Compatible with evolving LLM provider APIs

No action required for normal usage - the system automatically handles tool calling through the standard LLM interface.

## Testing Coverage

**Comprehensive Test Suite:**
- **Unit Tests**: All core components with table-driven tests
- **Integration Tests**: End-to-end agent workflows
- **Tool System Tests**: Comprehensive tool execution validation
- **Security Tests**: Security validation and threat detection
- **Performance Tests**: Benchmarking and metrics validation

**Test Organization:**
- `*_test.go` files throughout codebase (29+ test functions)
- Mock providers for testing without API dependencies
- 100% test coverage for tool system components
- ReAct agent integration testing

## Development Workflow

### Agent Architecture Components

**`internal/core/`** - Domain-Driven Core Components
- **`agent/`**: Core agent orchestration and management
  - `agent.go`: Legacy agent with tool calling capabilities
  - `react_agent.go`: Unified ReAct agent with Think-Act-Observe cycle
  - `factory.go`: Agent factory and configuration management
  - `unified_wrapper.go`: Bridge between ReAct and legacy architectures
  - `*_adapter.go`: Context, memory, and tool system adapters
- **`reasoning/`**: ReAct reasoning engine and strategy planning
- **`planning/`**: Action planning and execution strategy
- **`observation/`**: Result analysis and learning extraction

**`internal/tools/`** - Modular Tool Architecture
- **`registry/`**: Tool registry with metrics and recommendations
- **`builtin/`**: Core tool implementations with security validation
- **`execution/`**: Tool execution engine with advanced features
- Tool recommendation engine with confidence scoring
- Execution plan optimization with dependency analysis

**`internal/security/`** - Multi-layered Security System
- `manager.go`: Policy-based security management
- Risk assessment engine with threat detection
- Behavior analysis and audit logging

**`internal/session/`** - Session Management
- `session.go`: Persistent conversation storage and restoration
- Message history tracking with metadata
- Automatic cleanup and memory management

**`internal/cli/`** - CLI Interface Components
- **`commands/`**: Slash command handlers and CLI operations
- **`input/`**: Context-aware input processing and validation

**`pkg/types/`** - Comprehensive Type System
- Core types: `Task`, `ToolCall`, `ToolResult`, `AgentResponse`
- ReAct types: `ThinkingResult`, `ActionPlan`, `ObservationResult`
- Security types: `SecurityPolicy`, `RiskAssessment`, `ThreatDetection`
- Tool types: `ToolSchema`, `ExecutionPlan`, `ToolRecommendation`

### Testing and Development Notes

The project includes comprehensive test coverage with table-driven tests for all core functionality. Use `go test -v ./internal/[package]/` to test individual components during development.

**Test Execution:**
```bash
go test ./internal/core/agent/     # Core agent tests (ReAct and legacy)
go test ./internal/core/reasoning/ # Reasoning engine tests
go test ./internal/core/planning/  # Planning tests
go test ./internal/core/observation/ # Observation tests
go test ./internal/tools/registry/ # Tool registry tests
go test ./internal/tools/execution/ # Tool execution tests
go test ./internal/tools/builtin/  # Built-in tool tests
go test ./internal/security/       # Security tests
go test ./internal/cli/            # CLI tests
go test ./...                      # All tests
```

Performance is critical - all operations target sub-30ms execution times. The codebase was migrated from TypeScript to achieve 40-100x performance improvements through Go's compiled nature and concurrent processing capabilities.

Configuration is managed through `~/.deep-coding-config.json` with the config manager handling persistence automatically. AI provider switching is seamless through the interface abstraction.

## Code Principles

### Core Design Philosophy

**保持简洁清晰，如无需求勿增实体，尤其禁止过度配置**

- **Simplicity First**: Always choose the simplest solution that works
- **Clear Intent**: Code should be self-documenting through clear naming and structure
- **Minimal Configuration**: Avoid configuration options unless absolutely necessary
- **Purposeful Entities**: Only create new types, interfaces, or abstractions when they serve a clear purpose

### Naming Guidelines

**Use clear, descriptive names that express intent:**

#### Good Naming Patterns
- **Functions**: `AnalyzeCode()`, `LoadPrompts()`, `ExecuteTool()`
- **Types**: `ReactAgent`, `PromptLoader`, `ToolExecutor`
- **Variables**: `taskResult`, `userMessage`, `promptTemplate`

#### Avoid Generic Names
- ❌ `Manager`, `Handler`, `Service`, `Processor`
- ❌ `NewThing()`, `ProcessData()`, `HandleRequest()`
- ✅ `ReactAgent`, `ExecuteTool()`, `LoadPrompt()`

### Architectural Principles

1. **Single Responsibility**: Each component has one clear purpose
2. **Minimal Dependencies**: Reduce coupling between components
3. **Clear Interfaces**: Define simple, focused interfaces
4. **Error Handling**: Fail fast with clear error messages
5. **No Over-Engineering**: Don't build for theoretical future needs

### Configuration Policy

- **Default Behavior**: System should work with minimal configuration
- **Essential Only**: Only expose configuration for truly necessary options
- **Sensible Defaults**: All configuration should have reasonable defaults
- **Environment-Driven**: Prefer environment variables over config files when possible

## Current Architecture Status

### Production Ready Components:
- **ReAct Agent System**: Complete implementation with streaming support (`internal/agent/react_agent.go`)
- **Multi-Model LLM System**: Advanced factory pattern with HTTP and streaming clients (`internal/llm/`)
- **Configuration Management**: Multi-model configuration with validation (`internal/config/manager.go`)
- **Session Management**: File-based persistent storage (`internal/session/session.go`)
- **Advanced Tool System**: Comprehensive tool registry and execution engine (`internal/tools/`)
- **CLI Interface**: Rich command-line interface with config management (`cmd/`)

### Architecture Maturity:
- **ReAct Agent**: ✅ Production-ready with streaming and multi-model support
- **LLM Integration**: ✅ Complete with factory pattern and model type selection
- **Tool System**: ✅ Comprehensive with 8+ built-in tools and extensible registry
- **Configuration**: ✅ Advanced multi-model configuration management
- **CLI Interface**: ✅ Feature-complete with both interactive and single-prompt modes
- **Build System**: ✅ Fixed and fully functional with proper package compilation

### Performance Characteristics:
- Go-based implementation for maximum performance
- Concurrent tool execution with dependency management
- Memory-efficient session management
- Sub-30ms execution times for most operations
- 40-100x performance improvement over predecessor implementations

### Recent Major Changes (2024-06-30):
- **Unified Prompt System**: All prompts centralized in `internal/prompts` with markdown templates
- **Simplified Architecture**: Clear separation of concerns with ReactCore, ThinkingEngine, ToolExecutor
- **Eliminated Redundancy**: Removed hardcoded prompt templates from ReactAgent
- **Enhanced Reliability**: Multiple fallback layers for prompt loading failures
- **Code Simplification**: Adherence to "如无需求勿增实体" principle

This represents a mature, production-ready AI coding assistant with simplified architecture, centralized prompt management, and excellent maintainability. The system follows the core principle of keeping code simple and clear, avoiding unnecessary complexity.