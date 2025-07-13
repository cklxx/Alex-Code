# Alex - High-Performance Universal AI Software Engineering Assistant

[![CI](https://github.com/cklxx/Alex-Code/actions/workflows/ci.yml/badge.svg)](https://github.com/cklxx/Alex-Code/actions/workflows/ci.yml)
[![Deploy to GitHub Pages](https://github.com/cklxx/Alex-Code/actions/workflows/deploy-pages.yml/badge.svg)](https://github.com/cklxx/Alex-Code/actions/workflows/deploy-pages.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/cklxx/Alex-Code)](https://goreportcard.com/report/github.com/cklxx/Alex-Code)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Alex** is a high-performance, universally accessible AI software engineering assistant featuring advanced dual-architecture design with both legacy and modern ReAct (Reasoning and Acting) agent systems. Built in Go for maximum performance and designed for developers at all skill levels, Alex provides an intuitive natural language interface for code analysis, file operations, and development tasks through an intelligent agent architecture with advanced tool calling capabilities, comprehensive security, streaming responses, and **CodeAct integration** for interactive code learning and execution.

🌐 **[Visit our website](https://cklxx.github.io/Alex-Code/)** | 📚 **[Documentation](docs/)** | 🚀 **[Quick Start](#quick-start)**

## Quick Start

```bash
# Build Alex
make build                    # Builds ./alex binary

# Interactive conversational mode (ReAct agent by default)
./alex                        # Auto-detects TTY and enters interactive mode
./alex -i                     # Explicit interactive mode

# Single prompt mode (shows completion time)
./alex "Analyze the current directory structure"
# Output: ✅ Task completed in 1.2s

# With streaming responses (default behavior)
./alex "List all Go files"

# Session management
./alex -r session_id -i       # Resume specific session
./alex session list           # List all sessions
```

## Core Features

**🧠 Dual Agent Architecture**: Advanced ReAct (Reasoning and Acting) agent with fallback to legacy mode for maximum reliability  
**🤖 CodeAct Integration**: Interactive code learning through direct execution and experimentation for enhanced problem solving  
**🛠 Enhanced Tool Ecosystem**: 8+ built-in tools with intelligent recommendations, concurrent execution, and security validation  
**🌐 Multi-Model LLM System**: Factory pattern supporting OpenAI, DeepSeek, OpenRouter with BasicModel and ReasoningModel types  
**🔒 Security-First Design**: Enterprise-grade risk assessment, threat detection, command validation, and path protection  
**⚡ High Performance**: Native Go implementation with concurrent tool execution and sub-30ms response times  
**📝 Session-Aware Todo Management**: Persistent todo lists per session with context-aware task tracking  
**🎯 Universal Access**: Natural language interface designed for developers at all skill levels

## Usage

### Interactive Mode - Your AI Coding Partner
```bash
./alex                        # Auto-detects terminal and enters interactive mode
./alex -i                     # Explicit interactive mode flag
```

### Configuration Management
```bash
./alex config show                   # Show current configuration
```

### Advanced Usage
```bash
# Configure model parameters
./alex --tokens 4000 --temperature 0.8 "Complex analysis task"

# Architecture selection (automatic fallback)
USE_REACT_AGENT=true ./alex -i       # Force ReAct agent
USE_LEGACY_AGENT=true ./alex -i      # Force legacy agent

# Development workflow
make dev                             # Format, vet, build, and test
make dev-safe                        # Safe development workflow
make test-functionality              # Quick functionality test
```

## Enhanced Tool System & CodeAct Integration

**File Operations**: `file_read`, `file_update`, `file_replace`, `file_list`, `directory_create`  
**Shell Execution**: `bash`, `script_runner`, `process_monitor` with security controls  
**Search Tools**: `grep`, `ripgrep`, `find` with flexible pattern matching  
**Session-Aware Todo Management**: `session_todo_read`, `session_todo_update` with persistent storage  
**Web Integration**: `web_search` for information retrieval  
**Reasoning Tools**: `think` for structured problem solving

### 🤖 CodeAct Agent Integration

Alex implements the **CodeAct** paradigm - an advanced approach where AI agents learn to solve coding tasks through **direct code execution** and **interactive experimentation**. This enables Alex to:

**🔬 Interactive Code Learning**
- **Execute and Learn**: Runs code snippets to understand behavior and debug issues
- **Iterative Refinement**: Tests solutions, observes results, and improves code through direct feedback
- **Real-time Problem Solving**: Adapts solutions based on actual execution results rather than theoretical knowledge

**⚡ Enhanced Code Understanding**
- **Dynamic Analysis**: Analyzes code behavior through execution rather than static analysis alone
- **Context-Aware Debugging**: Uses execution results to understand complex codebases and dependencies
- **Live Testing**: Validates solutions in real-time during development

**🛠 Practical Implementation**
- **Safety-First Execution**: Secure sandbox environment for code testing and experimentation
- **Multi-Language Support**: CodeAct capabilities across Go, Python, JavaScript, and other languages
- **Development Workflow Integration**: Seamlessly integrates with existing development tools and processes

**Tool System Features:**
- **Intelligent Recommendations**: Task-aware tool suggestions with confidence scoring
- **Concurrent Execution**: Optimized parallel/sequential execution based on dependencies  
- **Security Validation**: Comprehensive parameter and execution validation
- **Performance Metrics**: Usage statistics, error tracking, execution metrics
- **CodeAct Execution**: Safe code experimentation with learning-based improvements

## Project Architecture

```
alex/
├── cmd/                    # CLI entry points and command handlers
│   ├── main.go            # Primary application entry point
│   └── config.go          # Advanced configuration management commands
├── internal/               # Private application code
│   ├── agent/             # ReAct agent implementation with dual architecture
│   ├── llm/               # Multi-model LLM integration layer
│   ├── tools/             # Enhanced tool system with registry and execution
│   │   ├── registry/      # Tool discovery and management
│   │   ├── builtin/       # Core tool implementations
│   │   └── execution/     # Tool execution engine
│   ├── prompts/           # Centralized prompt templates (markdown-based)
│   ├── config/            # Multi-model configuration management
│   ├── session/           # Persistent session management with todo system
│   └── security/          # Security framework and threat detection
├── pkg/                   # Library code for external use
│   ├── interfaces/        # Public interfaces
│   └── types/             # Public type definitions
├── docs/                  # Comprehensive documentation
├── scripts/               # Development and automation scripts
└── benchmarks/            # Performance testing and benchmarks
```

## Development

```bash
# Development workflow
make dev                   # Format, vet, build, and test functionality
make dev-safe              # Safe development workflow (excludes broken tests)
make dev-robust            # Ultra-robust workflow with dependency management

# Testing options
make test                  # Run all tests
make test-working          # Run only working tests
make test-functionality    # Quick test of core functionality

# Code quality
make fmt                   # Format Go code
make vet                   # Run go vet
make build                 # Build Alex binary

# Testing individual components
go test ./internal/agent/             # Test ReAct agent system
go test ./internal/tools/builtin/     # Test builtin tools
go test ./internal/session/           # Test session management

# Docker development
./scripts/docker.sh dev    # Start development environment
./scripts/docker.sh test   # Run tests in container
```

## 🌐 Website & Documentation

Alex includes a beautiful, modern website that showcases the project features and provides comprehensive documentation.

### Local Development
```bash
# Start local website server
cd docs/
./deploy.sh               # Choose option 1 for local server

# Or use Python directly
python -m http.server 8000
```

### Automated Deployment
The website automatically deploys to GitHub Pages via CI/CD:

- **🔄 Auto-deploy**: Pushes to `main` branch trigger deployment
- **⚡ Fast**: Typically deploys in 2-5 minutes  
- **🔍 Validated**: HTML validation and optimization included
- **📊 Stats**: Auto-generates project statistics

### Setup GitHub Pages
```bash
# One-time setup for GitHub Pages
./scripts/setup-github-pages.sh
```

This script will:
1. ✅ Verify all required files exist
2. 🔧 Configure repository URLs
3. 📤 Commit and push changes
4. 📋 Provide setup instructions

**Manual Setup Steps:**
1. Go to repository **Settings > Pages**
2. Set source to **"GitHub Actions"**
3. Enable **"Read and write permissions"** in **Settings > Actions**

🌐 **Live Website**: [https://cklxx.github.io/Alex-Code/](https://cklxx.github.io/Alex-Code/)

## Configuration

### Initial Setup

1. **Get OpenRouter API Key**: Visit [OpenRouter](https://openrouter.ai/settings/keys) to create a free account and get your API key
2. **First Run**: Alex will create default configuration on first use
3. **Set API Key**: Edit `~/.alex-config.json` and replace `"sk-or-xxx"` with your actual API key

Alex stores configuration in: `~/.alex-config.json`

### Configuration Management
```bash
./alex config show                   # Show current configuration
```

**Default Configuration:**
```json
{
    "api_key": "sk-or-xxx",
    "base_url": "https://openrouter.ai/api/v1", 
    "model": "deepseek/deepseek-chat-v3-0324:free",
    "max_tokens": 4000,
    "temperature": 0.7,
    "max_turns": 25,
    "basic_model": {
        "model": "deepseek/deepseek-chat-v3-0324:free",
        "max_tokens": 4000,
        "temperature": 0.7
    },
    "reasoning_model": {
        "model": "deepseek/deepseek-r1:free",
        "max_tokens": 8000,
        "temperature": 0.3
    }
}
```

### Multi-Model Configuration Explained

- **basic_model**: Used for general tasks and tool calling (lighter, faster)
- **reasoning_model**: Used for complex problem-solving and analysis (more capable)
- Alex automatically selects the appropriate model based on task complexity

### Environment Variables

Configuration precedence: **Environment Variables > Config File > Defaults**

```bash
export OPENAI_API_KEY="your-openrouter-key"  # Overrides config file api_key
export ALLOWED_TOOLS="file_read,bash,grep"   # Restrict available tools 
export USE_REACT_AGENT="true"                # Force ReAct agent mode
export USE_LEGACY_AGENT="true"               # Force legacy agent mode
```

### Common Configuration Tasks

```bash
# View current configuration
./alex config show

# Quick start with environment variable (no config file editing needed)
OPENAI_API_KEY="your-key" ./alex "Hello world"

# Test configuration
./alex "Test my setup"
```

## Why Alex Excels

**🚀 Advanced Architecture & Performance**
- **Dual Agent Design**: ReAct agent with automatic fallback to legacy mode for maximum reliability
- **Zero Dependencies**: Built on Go standard library for maximum stability and performance  
- **Concurrent Execution**: Intelligent parallel tool processing with dependency analysis
- **Memory Efficient**: Automatic session cleanup and smart resource management
- **Lightning Speed**: Sub-30ms response times with 40-100x performance improvement over predecessors

**🛠 Enterprise-Grade Features**
- **Security-First Design**: Multi-layered security with threat detection and risk assessment
- **Session Management**: Persistent conversations with context-aware todo management
- **Multi-Model Support**: Factory pattern supporting different LLM providers and model types
- **Tool Ecosystem**: Enhanced tool system with intelligent recommendations and metrics
- **Industry Standards**: Follows Go project layout, enterprise patterns, and modern AI frameworks

**🎯 Universal Accessibility**
- **Natural Language Interface**: No special syntax required, intuitive for all skill levels
- **Cross-Platform**: Seamless operation on macOS, Linux, and Windows
- **Lightweight Deployment**: Minimal resource usage, suitable for any development environment
- **Extensible Design**: Clean interfaces for custom tool development and integration

## Recent Updates (v1.0 - 2025)

**🔄 Architecture Enhancements:**
- **Unified Prompt System**: All prompts centralized in `internal/prompts` with markdown templates
- **Session-Aware Todo Management**: Persistent todo lists per session with context injection
- **Enhanced Tool System**: Intelligent recommendations, concurrent execution, performance metrics
- **Simplified Context System**: Streamlined ProjectSummary replacing complex ProjectInfo/SystemEnv

**⚡ Performance Optimizations:**
- **ReAct Agent Refinements**: Improved Think-Act-Observe cycle with streaming support
- **Tool Calling Standardization**: OpenAI-compatible format throughout for better reliability
- **Memory Management**: Enhanced session cleanup and resource optimization

**🔧 Developer Experience:**
- **Enhanced Project Detection**: Better virtual environment detection for Python, Node.js, Rust
- **Improved Build System**: Comprehensive Makefile with multiple workflow options
- **Docker Development**: Complete containerized development environment

## Documentation

- **[CLAUDE.md](CLAUDE.md)**: Comprehensive project instructions and architecture overview
- **[Software Engineering Roles Analysis](docs/software-engineering-roles-analysis.md)**: Analysis of roles and responsibilities across software engineering phases

## Contributing

We welcome contributions! Please see our development workflow:

1. **Setup**: `make dev-robust` for complete environment setup
2. **Testing**: `make test-functionality` for quick validation
3. **Quality**: `make fmt && make vet` before submitting
4. **Architecture**: Follow the patterns established in `internal/` packages

## License

MIT License