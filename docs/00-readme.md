# Deep Coding Agent Documentation

Welcome to the Deep Coding Agent documentation. This directory contains comprehensive documentation for the high-performance conversational AI coding assistant.

## 📋 Documentation Structure

### Core Documentation (Read in Order)
- **[00-readme.md](00-readme.md)** - This documentation index
- **[01-architecture-overview.md](01-architecture-overview.md)** - Comprehensive architecture analysis and implementation plan
- **[02-react-agent-design.md](02-react-agent-design.md)** - Detailed ReactAgent design and implementation
- **[03-prompt-system.md](03-prompt-system.md)** - Prompt template system architecture
- **[04-prompt-design.md](04-prompt-design.md)** - System prompts design patterns

### Implementation Guides
- **[guides/quickstart.md](guides/quickstart.md)** - Quick start guide for users
- **[guides/tool-development.md](guides/tool-development.md)** - Tool development guide

### CodeAct Integration
- **[codeact/integration-guide.md](codeact/integration-guide.md)** - Complete CodeAct integration guide
- **[codeact/api-reference.md](codeact/api-reference.md)** - CodeAct API reference
- **[codeact/implementation-roadmap.md](codeact/implementation-roadmap.md)** - Implementation roadmap

### API Reference
- **[reference/api-reference.md](reference/api-reference.md)** - General API reference

### Research & Analysis
- **[research/industry-benchmarks.md](research/industry-benchmarks.md)** - Industry research and benchmarking
- **[research/execution-flow-analysis.md](research/execution-flow-analysis.md)** - ReAct execution flow analysis
- **[research/react-architecture-summary.md](research/react-architecture-summary.md)** - ReAct architecture summary
- **[research/react-patterns.md](research/react-patterns.md)** - ReAct implementation patterns
- **[research/react-implementation.md](research/react-implementation.md)** - ReAct implementation details
- **[research/agent-architecture.md](research/agent-architecture.md)** - Agent architecture research

## 🚀 Quick Start

1. **Begin with Architecture**: Start with [01-architecture-overview.md](01-architecture-overview.md) for a complete understanding of the system
2. **Implementation Details**: Read [02-react-agent-design.md](02-react-agent-design.md) for detailed design patterns
3. **Getting Started**: Follow [guides/quickstart.md](guides/quickstart.md) for immediate usage
4. **CodeAct Features**: Explore [codeact/integration-guide.md](codeact/integration-guide.md) for advanced capabilities

## 📖 Key Features Documented

- **Dual-Architecture ReAct Agent**: Think-Act-Observe cycle with streaming support
- **Multi-Model LLM Integration**: Dynamic model selection with factory pattern
- **Advanced Tool System**: 8+ built-in tools with extensible registry
- **Session Management**: Persistent conversation storage and restoration
- **CodeAct Integration**: Executable Python code as action language
- **Security Framework**: Multi-layered security with sandbox execution
- **Performance Optimization**: Go-based implementation with 40-100x improvements

## 🏗️ Architecture Overview

The Deep Coding Agent features a sophisticated dual-architecture design:

```
┌─────────────────────────────────────────────────────────────┐
│                    Deep Coding Agent                       │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────────┐ │
│  │   Think     │  │     Act      │  │       Observe       │ │
│  │  (Reason)   │→ │  (Execute)   │→ │      (Analyze)      │ │
│  └─────────────┘  └──────────────┘  └─────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│ Multi-Model LLM • Tool Registry • Session Management       │
└─────────────────────────────────────────────────────────────┘
```

## 📝 Contributing

When contributing to documentation:

1. Follow the established naming convention (numbered prefixes for core docs)
2. Update this index when adding new documentation
3. Place research and experimental docs in the `research/` directory
4. Use clear, descriptive filenames with hyphens for spacing

## 🔍 Document Categories

- **Core**: Essential system documentation (numbered 00-04)
- **Guides**: User and developer guides
- **CodeAct**: CodeAct-specific documentation
- **Reference**: API and technical reference
- **Research**: Research papers, analysis, and experimental documentation

For support and questions, refer to the main project README or contact the development team.