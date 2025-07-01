# ReactAgent Design Document
## Deep Coding Agent v1.0 - ReAct Architecture Implementation

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Core Components](#core-components)
4. [ReAct Pattern Implementation](#react-pattern-implementation)
5. [Tool System Architecture](#tool-system-architecture)
6. [LLM Integration & Multi-Model Support](#llm-integration--multi-model-support)
7. [Streaming & Session Management](#streaming--session-management)
8. [Configuration Management](#configuration-management)
9. [Error Handling & Recovery](#error-handling--recovery)
10. [Performance & Optimization](#performance--optimization)
11. [Security Framework](#security-framework)
12. [Extensibility & Maintenance](#extensibility--maintenance)
13. [Implementation Details](#implementation-details)
14. [Future Enhancements](#future-enhancements)

---

## Executive Summary

The **ReactAgent** is the core component of the Deep Coding Agent v1.0, implementing a sophisticated **ReAct (Reasoning and Acting)** pattern for conversational AI coding assistance. Built in Go for maximum performance, it provides a production-ready foundation for complex reasoning tasks with advanced tool orchestration, streaming responses, and comprehensive session management.

### Key Design Principles

- **Think-Act-Observe Cycle**: Structured three-phase processing for complex problem solving
- **Multi-Model Intelligence**: Dynamic model selection based on task complexity
- **Streaming-First Design**: Real-time feedback with efficient resource utilization
- **Production-Ready Reliability**: Comprehensive error handling, validation, and recovery
- **Extensible Architecture**: Plugin-style tool system with clean interfaces

---

## Architecture Overview

### High-Level Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                          ReactAgent                             │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────────────┐ │
│  │   Think     │  │     Act      │  │       Observe           │ │
│  │  (Reason)   │→ │  (Execute)   │→ │      (Analyze)          │ │
│  │             │  │              │  │                         │ │
│  └─────────────┘  └──────────────┘  └─────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                     Component Integration                       │
├─────────────────────────────────────────────────────────────────┤
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│ │   LLM       │ │   Tools     │ │   Session   │ │   Config    │ │
│ │ Multi-Model │ │  Registry   │ │ Management  │ │  Manager    │ │
│ │   Client    │ │             │ │             │ │             │ │
│ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### Core Architecture Components

```go
type ReactAgent struct {
    llm            llm.Client               // Multi-model LLM abstraction
    configManager  *config.Manager          // Unified configuration management
    sessionManager *session.Manager         // Persistent session storage
    toolRegistry   *SimpleToolRegistry      // Tool discovery and execution
    currentSession *session.Session         // Active conversation state
    promptLoader   *prompts.PromptLoader    // Template-based prompt system
}
```

---

## Core Components

### 1. ReactAgent Structure

**Primary Responsibilities:**
- Orchestrate the ReAct thinking cycle
- Manage conversation state and session persistence
- Coordinate tool execution and result processing
- Handle streaming responses and real-time feedback

**Key Methods:**
```go
// Core processing methods
ProcessMessage(ctx, message, config) (*Response, error)
ProcessMessageStream(ctx, message, config, callback) error

// Session management
StartSession(sessionID) (*Session, error)
RestoreSession(sessionID) (*Session, error)

// ReAct cycle implementation
executeReActCycle(ctx, userMessage, config) (*ReActResult, error)
executeReActCycleStream(ctx, userMessage, config, callback) (*ReActResult, error)
```

### 2. Data Structures

**ThoughtResult** - Reasoning Phase Output:
```go
type ThoughtResult struct {
    Analysis       string            `json:"analysis"`        // Reasoning analysis
    Content        string            `json:"content"`         // Response content
    ShouldComplete bool              `json:"should_complete"` // Completion decision
    Confidence     float64           `json:"confidence"`      // Confidence score (0-1)
    PlannedActions []PlannedAction   `json:"planned_actions"` // Tool execution plan
}
```

**PlannedAction** - Tool Execution Plan:
```go
type PlannedAction struct {
    ToolName  string                 `json:"tool_name"`  // Tool identifier
    Arguments map[string]interface{} `json:"arguments"`  // Tool parameters
    Reasoning string                 `json:"reasoning"`  // Execution rationale
}
```

**ObservationResult** - Analysis Phase Output:
```go
type ObservationResult struct {
    Summary      string   `json:"summary"`       // Result summary
    TaskComplete bool     `json:"task_complete"` // Completion status
    Confidence   float64  `json:"confidence"`    // Analysis confidence
    Insights     []string `json:"insights"`      // Extracted insights
}
```

---

## ReAct Pattern Implementation

### Three-Phase Processing Cycle

#### Phase 1: Think (Reasoning)
**Purpose:** Analyze the user request and plan appropriate actions

**Implementation:**
```go
func (r *ReactAgent) think(ctx context.Context, userMessage string, 
                          conversationHistory []*session.Message) (*ThoughtResult, error)
```

**Key Features:**
- Embedded Markdown prompt templates for structured reasoning
- Context-aware conversation history integration
- JSON-structured response parsing with fallback handling
- Model selection (BasicModel for quick thinking)

**Prompt Template Structure:**
```markdown
## Critical Guidelines
### ✅ Set `should_complete: true` for:
- Simple greetings, general coding questions, explanations

### ❌ Set `should_complete: false` for:
- File operations, code generation, system commands

Response format: {"analysis": "...", "content": "...", "should_complete": bool, "confidence": 0.8, "planned_actions": [...]}
```

#### Phase 2: Act (Tool Execution)
**Purpose:** Execute planned tools and gather results

**Implementation:**
```go
func (r *ReactAgent) act(ctx context.Context, actions []PlannedAction, 
                        config *config.Config) ([]ToolResult, error)
```

**Execution Flow:**
1. **Tool Lookup**: Retrieve tool from registry by name
2. **Parameter Validation**: JSON Schema-based validation
3. **Security Check**: Command safety and path restrictions
4. **Execution**: Context-aware tool execution with timeout
5. **Result Processing**: Structured result wrapping

**Tool Integration:**
```go
func (r *ReactAgent) executeTool(ctx context.Context, toolName string, 
                                arguments map[string]interface{}) (ToolResult, error)
```

#### Phase 3: Observe (Result Analysis)
**Purpose:** Analyze tool results and determine task completion

**Implementation:**
```go
func (r *ReactAgent) observe(ctx context.Context, thought *ThoughtResult, 
                            toolResults []ToolResult) (*ObservationResult, error)
```

**Analysis Process:**
- Aggregate tool execution results
- Evaluate task completion against original request
- Extract insights and learning opportunities
- Use ReasoningModel for deeper analysis

### Flow Control Mechanisms

**Iterative Processing:**
```go
maxTurns := 3  // Configurable via config.MaxTurns
for turn := 1; turn <= maxTurns; turn++ {
    // Execute Think-Act-Observe cycle
    if thought.ShouldComplete && thought.Confidence >= 0.7 {
        break  // Early termination on high confidence
    }
}
```

**Confidence-Based Completion:**
- Confidence threshold: 0.7 (configurable)
- Early termination prevents unnecessary iterations
- Maximum iteration safety net (default: 3 turns)

---

## Tool System Architecture

### SimpleToolRegistry Design

```go
type SimpleToolRegistry struct {
    tools map[string]builtin.Tool
}
```

**Tool Interface:**
```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{} // JSON Schema
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
    Validate(args map[string]interface{}) error
}
```

### Available Tool Categories

**File Operations:**
- `file_read`: Read file contents with encoding detection
- `file_update`: Modify existing files with backup
- `file_replace`: Replace file contents atomically
- `file_list`: Directory listing with filtering
- `directory_create`: Create directories with permissions

**Search Tools:**
- `grep`: Pattern searching with regex support
- `ripgrep`: High-performance text search
- `find`: File system search with complex criteria

**Execution Tools:**
- `bash`: Shell command execution with security controls
- `script_runner`: Script execution with environment isolation
- `process_monitor`: Process status and resource monitoring

**Web Tools:**
- `web_search`: General web search integration
- `news_search`: News-specific search
- `academic_search`: Academic paper search

### Tool Execution Pipeline

1. **Discovery**: Tool lookup from registry
2. **Validation**: Parameter validation via JSON Schema
3. **Security**: Command safety and path restrictions
4. **Execution**: Context-aware execution with timeout
5. **Result Processing**: Structured response formatting

**Security Framework:**
```go
type ValidationFramework struct {
    rules []ValidationRule
}
```

- Parameter type checking and range validation
- Dangerous command pattern detection
- System path restrictions
- Command length limits and injection prevention

---

## LLM Integration & Multi-Model Support

### Multi-Model Architecture

**Model Types:**
```go
const (
    BasicModel     ModelType = "basic"     // Fast, general tasks
    ReasoningModel ModelType = "reasoning" // Complex analysis
)
```

**Configuration:**
```go
type Config struct {
    Models map[ModelType]*ModelConfig `json:"models,omitempty"`
    DefaultModelType ModelType `json:"default_model_type,omitempty"`
}
```

### LLM Client Interface

**Dual-Mode Support:**
```go
type Client interface {
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamDelta, error)
    Close() error
}
```

**Request Structure:**
```go
type ChatRequest struct {
    Messages    []Message `json:"messages"`
    ModelType   ModelType `json:"model_type,omitempty"`  // Dynamic selection
    Temperature float64   `json:"temperature,omitempty"`
    MaxTokens   int       `json:"max_tokens,omitempty"`
    Stream      bool      `json:"stream,omitempty"`
}
```

### Provider Support

**Supported Providers:**
- **OpenAI**: GPT-4, GPT-3.5-turbo with proper authentication
- **Anthropic**: Claude models with custom headers
- **Azure OpenAI**: Enterprise deployment support
- **OpenRouter**: Multi-provider access with DeepSeek defaults

**Default Configuration:**
```go
Models: map[ModelType]*ModelConfig{
    BasicModel: {
        BaseURL:     "https://openrouter.ai/api/v1",
        Model:       "deepseek/deepseek-chat-v3-0324:free",
        APIKey:      "sk-or-v1-...",
        Temperature: 0.7,
        MaxTokens:   2048,
    },
    ReasoningModel: {
        BaseURL:     "https://openrouter.ai/api/v1",
        Model:       "deepseek/deepseek-chat-v3-0324:free",
        APIKey:      "sk-or-v1-...",
        Temperature: 0.3,
        MaxTokens:   4096,
    },
}
```

---

## Streaming & Session Management

### Streaming Architecture

**Dual-Mode Processing:**
```go
// Synchronous mode
ProcessMessage(ctx, userMessage, config) (*Response, error)

// Streaming mode
ProcessMessageStream(ctx, userMessage, config, callback StreamCallback) error
```

**StreamChunk Structure:**
```go
type StreamChunk struct {
    Type     string `json:"type"`     // Event type
    Content  string `json:"content"`  // Content payload
    Complete bool   `json:"complete,omitempty"`
}
```

**Streaming Events:**
- `"status"`: Phase transitions and progress updates
- `"content"`: LLM response content chunks
- `"tool_start"`, `"tool_result"`, `"tool_error"`: Tool execution feedback
- `"complete"`: Processing completion signal

### Server-Sent Events Implementation

**Streaming Client:**
```go
func (c *StreamingLLMClient) ChatStream(ctx context.Context, req *ChatRequest) 
                                       (<-chan StreamDelta, error) {
    // Create buffered channel for performance
    deltaChannel := make(chan StreamDelta, 100)
    
    // Goroutine-based stream processing
    go func() {
        defer close(deltaChannel)
        // SSE parsing and event emission
    }()
    
    return deltaChannel, nil
}
```

### Session Management

**Session Structure:**
```go
type Session struct {
    ID         string                    `json:"id"`
    Created    time.Time                 `json:"created"`
    Updated    time.Time                 `json:"updated"`
    Messages   []*Message                `json:"messages"`
    WorkingDir string                    `json:"working_dir,omitempty"`
    Config     map[string]interface{}    `json:"config,omitempty"`
    mutex      sync.RWMutex             // Thread safety
}
```

**Persistence Strategy:**
- **File-based Storage**: `~/.deep-coding-sessions/`
- **JSON Serialization**: Human-readable format
- **Automatic Cleanup**: Configurable retention policies
- **Memory Management**: In-memory caching with cleanup

**Thread Safety:**
```go
func (s *Session) AddMessage(message *Message) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.Messages = append(s.Messages, message)
    s.Updated = time.Now()
}
```

---

## Configuration Management

### Unified Configuration System

**Configuration Structure:**
```go
type Config struct {
    // Legacy compatibility
    APIKey      string  `json:"api_key"`
    BaseURL     string  `json:"base_url"`
    Model       string  `json:"model"`
    MaxTokens   int     `json:"max_tokens"`
    Temperature float64 `json:"temperature"`
    
    // ReAct agent configuration
    MaxTurns    int     `json:"max_turns"`
    
    // Multi-model configurations
    Models map[ModelType]*ModelConfig `json:"models,omitempty"`
    DefaultModelType ModelType `json:"default_model_type,omitempty"`
}
```

**Configuration Management:**
```go
type Manager struct {
    configPath string    // ~/.deep-coding-config.json
    config     *Config   // In-memory configuration
}
```

### Configuration Hierarchy

**Fallback Order:**
1. Multi-model specific configuration
2. Default single model configuration
3. Environment variables
4. Built-in defaults (DeepSeek models)

**Validation Framework:**
```go
func (m *Manager) ValidateConfig() error {
    // Required field validation
    if config.APIKey == "" {
        return fmt.Errorf("api_key is required")
    }
    
    // Range checking
    if config.Temperature < 0.0 || config.Temperature > 2.0 {
        return fmt.Errorf("temperature must be between 0.0 and 2.0")
    }
    
    // Model availability verification
    return nil
}
```

---

## Error Handling & Recovery

### Layered Error Handling

**Context Propagation:**
- All methods accept `context.Context` for cancellation
- Timeout handling and resource cleanup
- Graceful shutdown on interruption

**Error Wrapping:**
```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### Recovery Mechanisms

**JSON Parsing Fallbacks:**
```go
func parseThoughtResult(responseContent string) (*ThoughtResult, error) {
    cleanContent := cleanJSONResponse(responseContent)
    
    var thought ThoughtResult
    if err := json.Unmarshal([]byte(cleanContent), &thought); err != nil {
        // Graceful degradation
        return &ThoughtResult{
            Analysis:       "Failed to parse JSON response",
            Content:        responseContent,
            ShouldComplete: true,
            Confidence:     0.7,
            PlannedActions: []PlannedAction{},
        }, nil
    }
    
    return &thought, nil
}
```

**Tool Execution Safety:**
- Individual tool failures don't abort the cycle
- Error results captured and analyzed
- Timeout protection via context cancellation

**Response Cleaning:**
```go
func cleanJSONResponse(response string) string {
    // Remove markdown code blocks
    re := regexp.MustCompile("```(?:json)?\n?(.*?)\n?```")
    matches := re.FindAllStringSubmatch(response, -1)
    
    if len(matches) > 0 {
        return strings.TrimSpace(matches[0][1])
    }
    
    // Extract JSON from { ... } structure
    jsonStart := strings.Index(response, "{")
    jsonEnd := strings.LastIndex(response, "}")
    
    if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
        return strings.TrimSpace(response[jsonStart : jsonEnd+1])
    }
    
    return strings.TrimSpace(response)
}
```

### Debug and Monitoring

**Comprehensive Logging:**
```go
fmt.Printf("[DEBUG] Starting ReAct cycle with maxTurns=%d for message: %s\n", maxTurns, userMessage)
fmt.Printf("[DEBUG] Think phase completed. ShouldComplete=%t, Confidence=%.2f\n", 
           thought.ShouldComplete, thought.Confidence)
fmt.Printf("[DEBUG] Tool execution: %s -> %v\n", action.ToolName, result.Success)
```

---

## Performance & Optimization

### Performance Optimizations

**Concurrent Design:**
- Goroutine-based streaming with buffered channels
- Non-blocking tool execution patterns
- Efficient resource management

**Memory Management:**
- Session message trimming (configurable limits)
- Automatic cleanup routines
- Reference counting for tool results

**Model Selection Strategy:**
- BasicModel for quick thinking (lower resource usage)
- ReasoningModel for complex analysis (higher quality)
- Dynamic selection based on task complexity

### Resource Management

**Channel Optimization:**
```go
// Buffered channels for streaming performance
deltaChannel := make(chan StreamDelta, 100)
```

**Session Cleanup:**
```go
func (m *Manager) CleanupExpiredSessions(maxAge time.Duration) error {
    // Automatic cleanup of old sessions
    // Memory optimization and disk space management
}
```

**Tool Execution Timeout:**
```go
ctx, cancel := context.WithTimeout(ctx, toolTimeout)
defer cancel()
```

---

## Security Framework

### Multi-Layered Security

**Command Validation:**
```go
type ValidationFramework struct {
    rules []ValidationRule
}

func (v *ValidationFramework) ValidateCommand(command string) error {
    // Dangerous pattern detection
    // Injection prevention
    // Length limits
}
```

**Path Restrictions:**
- System directory protection
- Executable file restrictions
- User home directory scoping

**Tool Security:**
- Parameter sanitization
- Type checking and validation
- Resource limits and timeouts

### Security Policies

**Risk Assessment:**
- Tool complexity scoring
- Path sensitivity analysis
- Command danger detection

**Access Control:**
- Tool-specific permissions
- Configurable allowed/denied lists
- Runtime restriction enforcement

---

## Extensibility & Maintenance

### Plugin Architecture

**Tool System Extensibility:**
```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
    Validate(args map[string]interface{}) error
}
```

**Dynamic Registration:**
```go
func (r *SimpleToolRegistry) RegisterTool(tool builtin.Tool) {
    r.tools[tool.Name()] = tool
}
```

### Configuration Flexibility

**Multi-Model Support:**
- Easy addition of new LLM providers
- Provider-specific configuration
- Runtime model switching

**Prompt Template System:**
- Markdown-based template management
- Variable substitution support
- Embedded filesystem distribution

### Maintenance Features

**Backward Compatibility:**
- Legacy configuration support
- Graceful migration paths
- Version compatibility checks

**Testing Infrastructure:**
- Comprehensive unit tests
- Integration test coverage
- Mock provider support

---

## Implementation Details

### Key Implementation Patterns

**Factory Pattern Usage:**
```go
func NewReactAgent(configManager *config.Manager) (*ReactAgent, error) {
    // LLM client creation
    llmClient, err := llm.CreateClient(configManager.GetConfig())
    
    // Component initialization
    sessionManager, err := session.NewManager()
    toolRegistry := NewSimpleToolRegistry()
    promptLoader, err := prompts.NewPromptLoader()
    
    return &ReactAgent{
        llm:            llmClient,
        configManager:  configManager,
        sessionManager: sessionManager,
        toolRegistry:   toolRegistry,
        promptLoader:   promptLoader,
    }, nil
}
```

**Interface-Driven Design:**
- Clean separation between components
- Testability through dependency injection
- Extensibility through interface compliance

**Resource Management:**
```go
func (r *ReactAgent) Close() error {
    if r.llm != nil {
        return r.llm.Close()
    }
    return nil
}
```

### Code Quality Patterns

**Consistent Error Handling:**
```go
if err != nil {
    return nil, fmt.Errorf("operation failed: %w", err)
}
```

**Structured Logging:**
```go
fmt.Printf("[%s] %s: %v\n", level, component, message)
```

**Thread Safety:**
```go
type SafeComponent struct {
    mutex sync.RWMutex
    data  map[string]interface{}
}
```

---

## Future Enhancements

### Planned Improvements

**Enhanced Tool System:**
- Advanced dependency analysis
- Parallel execution optimization
- Tool recommendation engine
- Performance metrics tracking

**Advanced ReAct Features:**
- Multi-turn conversation planning
- Context window optimization
- Adaptive confidence thresholds
- Learning from execution patterns

**Monitoring & Observability:**
- Structured logging with levels
- Metrics collection and export
- Performance profiling
- Usage analytics

### Extensibility Roadmap

**Additional LLM Providers:**
- Local model support (Ollama)
- Custom provider integration
- Model performance optimization

**Advanced Security:**
- Sandbox execution environments
- Fine-grained permission systems
- Audit trail and compliance

**Developer Experience:**
- IDE integration support
- Visual debugging tools
- Configuration management UI

---

## Conclusion

The ReactAgent implementation represents a sophisticated, production-ready AI coding assistant with:

1. **Advanced ReAct Pattern**: Complete Think-Act-Observe cycle implementation
2. **Multi-Model Intelligence**: Dynamic model selection for optimal performance
3. **Streaming Architecture**: Real-time feedback with efficient resource usage
4. **Comprehensive Tool System**: Extensible plugin architecture with security
5. **Production-Ready Features**: Error handling, validation, and monitoring
6. **Excellent Performance**: Go-based implementation with concurrent processing

The architecture provides a solid foundation for complex AI reasoning tasks while maintaining flexibility, security, and ease of maintenance. The design patterns and implementation choices enable both immediate productivity and long-term extensibility.

---

*Generated by Deep Coding Agent v1.0 ReactAgent Analysis*
*Last Updated: 2024-06-29*