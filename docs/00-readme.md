# Deep Coding Agent Documentation

Welcome to the Deep Coding Agent documentation. This directory contains comprehensive documentation for the high-performance conversational AI coding assistant.

## 📋 Documentation Structure

### 🏗️ Architecture Documentation
- **[architecture/01-architecture-overview.md](architecture/01-architecture-overview.md)** - Comprehensive architecture analysis and implementation plan
- **[architecture/02-react-agent-design.md](architecture/02-react-agent-design.md)** - Detailed ReactAgent design and implementation
- **[architecture/03-prompt-system.md](architecture/03-prompt-system.md)** - Prompt template system architecture
- **[architecture/04-prompt-design.md](architecture/04-prompt-design.md)** - System prompts design patterns

### 🛠️ Implementation Guides
- **[guides/quickstart.md](guides/quickstart.md)** - Quick start guide for users
- **[guides/tool-development.md](guides/tool-development.md)** - Tool development guide
- **[implementation/chromem-local-embeddings-config.md](implementation/chromem-local-embeddings-config.md)** - Local embeddings configuration
- **[implementation/cli-input-bottom-design.md](implementation/cli-input-bottom-design.md)** - CLI input design patterns

### 🔬 Research & Experiments
- **[research/industry-benchmarks.md](research/industry-benchmarks.md)** - Industry research and benchmarking
- **[research/execution-flow-analysis.md](research/execution-flow-analysis.md)** - ReAct execution flow analysis
- **[research/react-architecture-summary.md](research/react-architecture-summary.md)** - ReAct architecture summary
- **[research/react-patterns.md](research/react-patterns.md)** - ReAct implementation patterns
- **[research/react-implementation.md](research/react-implementation.md)** - ReAct implementation details
- **[research/agent-architecture-old.md](research/agent-architecture-old.md)** - Legacy agent architecture research
- **[research/codeact-research-report.md](research/codeact-research-report.md)** - CodeAct research and analysis

### 📊 Analysis & Reports
- **[analysis/CONTEXT_ENGINEERING_AND_COMPRESSION_RESEARCH.md](analysis/CONTEXT_ENGINEERING_AND_COMPRESSION_RESEARCH.md)** - Context engineering research
- **[analysis/DATABASE_INTEGRATION_ULTRA_ANALYSIS_2025.md](analysis/DATABASE_INTEGRATION_ULTRA_ANALYSIS_2025.md)** - Database integration analysis
- **[analysis/software-engineering-roles-analysis.md](analysis/software-engineering-roles-analysis.md)** - Software engineering roles analysis

### 🧩 CodeAct Integration
- **[codeact/integration-guide.md](codeact/integration-guide.md)** - Complete CodeAct integration guide
- **[codeact/api-reference.md](codeact/api-reference.md)** - CodeAct API reference
- **[codeact/implementation-roadmap.md](codeact/implementation-roadmap.md)** - Implementation roadmap

### 📚 API Reference
- **[reference/api-reference.md](reference/api-reference.md)** - General API reference

## 🚀 Quick Start

1. **Begin with Architecture**: Start with [architecture/01-architecture-overview.md](architecture/01-architecture-overview.md) for a complete understanding of the system
2. **Implementation Details**: Read [architecture/02-react-agent-design.md](architecture/02-react-agent-design.md) for detailed design patterns
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

- **Architecture**: Core system architecture and design documentation (🏗️)
- **Implementation**: Implementation guides and configuration details (🛠️)
- **Research**: Experimental features and research papers (🔬)
- **Analysis**: Data analysis, reports, and engineering studies (📊)
- **CodeAct**: CodeAct-specific integration documentation (🧩)
- **Reference**: API and technical reference materials (📚)
- **Guides**: User and developer guides for getting started

For support and questions, refer to the main project README or contact the development team.