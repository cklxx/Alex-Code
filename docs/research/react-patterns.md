# ReAct Agent Patterns for Multi-Turn Tool Calling

## Overview

This document provides comprehensive analysis and implementation guidance for ReAct (Reasoning and Acting) agent patterns, specifically focusing on single-task multi-turn tool calling scenarios. Based on industry research and the current Deep Coding Agent implementation, this guide presents practical patterns for sophisticated agent behaviors.

## ReAct Pattern Fundamentals

### Core ReAct Cycle

The ReAct pattern follows an iterative cycle:

```
Thought → Action → Observation → [Repeat] → Answer
```

**Key Characteristics:**
- **Thought**: Chain-of-thought reasoning about the current situation
- **Action**: Tool execution based on reasoning
- **Observation**: Analysis of tool results and environmental feedback
- **Iteration**: Continuation of the cycle until task completion

### ReAct vs Function Calling Agents

| Aspect | ReAct Agent | Function Calling Agent |
|--------|-------------|------------------------|
| **LLM Compatibility** | Works with any LLM | Requires function calling support |
| **Reasoning Transparency** | Explicit reasoning traces | Implicit reasoning |
| **Multi-turn Capability** | Native multi-turn support | Limited multi-turn by design |
| **Error Recovery** | Adaptive error handling | Fixed error handling |
| **Context Preservation** | Maintains reasoning context | Limited context between calls |

## Single Task Multi-Turn Patterns

### 1. **Progressive Refinement Pattern**

**Use Case**: Complex analysis requiring iterative deepening

```
Initial Request: "Analyze the performance bottlenecks in this Go application"

Turn 1: Initial Discovery
Thought: "I need to understand the application structure and identify potential performance issues"
Action: [file_list, file_read go.mod, bash "find . -name '*.go' | head -10"]
Observation: "Found a web server with database connections and multiple goroutines"

Turn 2: Performance Profiling
Thought: "Based on the structure, I should run performance profiling to identify bottlenecks"
Action: [bash "go test -bench=. -benchmem", bash "go tool pprof cpu.prof"]
Observation: "High memory allocation in database query handling, CPU bottleneck in JSON parsing"

Turn 3: Deep Analysis
Thought: "The profiling shows specific issues. Let me examine the problematic code sections"
Action: [file_read "internal/db/queries.go", file_read "internal/api/handlers.go"]
Observation: "Found N+1 query problem and inefficient JSON marshaling"

Turn 4: Solution Synthesis
Thought: "Now I can provide specific recommendations based on the analysis"
Action: [todo_update "Create performance optimization plan"]
Answer: "Identified key bottlenecks: N+1 queries, inefficient JSON processing..."
```

### 2. **Exploratory Search Pattern**

**Use Case**: Bug investigation with unknown root cause

```
Initial Request: "The application is crashing intermittently, help me debug"

Turn 1: Context Gathering
Thought: "I need to understand the application state and gather error information"
Action: [bash "tail -100 /var/log/app.log", bash "ps aux | grep app", file_read "go.mod"]
Observation: "Found panic traces related to nil pointer dereferences in concurrent code"

Turn 2: Code Investigation
Thought: "The panic traces point to concurrency issues. Let me examine the concurrent code paths"
Action: [bash "grep -r 'go func\\|goroutine' --include='*.go'", file_read "internal/worker/pool.go"]
Observation: "Found several goroutines accessing shared resources without proper synchronization"

Turn 3: Race Condition Analysis
Thought: "This looks like race conditions. Let me run the race detector and examine synchronization"
Action: [bash "go test -race ./...", bash "grep -r 'sync\\.' --include='*.go'"]
Observation: "Race detector confirms data races in worker pool and cache access"

Turn 4: Root Cause Identification
Thought: "I've identified the root cause. Let me provide a comprehensive solution"
Action: [file_write "race_conditions_report.md", todo_update "Fix race conditions in worker pool"]
Answer: "Found race conditions in concurrent access to shared state..."
```

### 3. **Incremental Implementation Pattern**

**Use Case**: Building a new feature step-by-step

```
Initial Request: "Implement a caching layer for the API responses"

Turn 1: Architecture Planning
Thought: "I need to understand the current API structure and plan the caching integration"
Action: [file_list "internal/api", file_read "internal/api/server.go", file_read "go.mod"]
Observation: "REST API with middleware support, no existing caching, uses Chi router"

Turn 2: Infrastructure Setup
Thought: "I'll create the caching infrastructure with Redis support and configuration"
Action: [directory_create "internal/cache", file_write "internal/cache/redis.go", file_write "internal/cache/interface.go"]
Observation: "Created cache interface and Redis implementation"

Turn 3: Middleware Implementation
Thought: "Now I'll implement the caching middleware to integrate with the existing API"
Action: [file_write "internal/middleware/cache.go", file_read "internal/api/routes.go"]
Observation: "Created caching middleware with configurable TTL and cache key generation"

Turn 4: Integration and Testing
Thought: "Let me integrate the middleware and create tests to ensure everything works"
Action: [file_write "internal/middleware/cache_test.go", bash "go test ./internal/cache/", bash "go test ./internal/middleware/"]
Observation: "All tests pass, caching middleware successfully integrated"

Turn 5: Configuration and Documentation
Thought: "Finally, I'll add configuration options and document the new caching feature"
Action: [file_write "config/cache.yaml", file_write "docs/CACHING.md", bash "go run main.go -test-cache"]
Answer: "Successfully implemented caching layer with Redis backend..."
```

### 4. **Adaptive Problem-Solving Pattern**

**Use Case**: Complex problems requiring strategy adjustment

```
Initial Request: "Optimize the database queries for better performance"

Turn 1: Initial Assessment
Thought: "I need to identify the current query patterns and performance issues"
Action: [bash "grep -r 'SELECT\\|INSERT\\|UPDATE' --include='*.go'", file_read "internal/db/migrations/"]
Observation: "Found many queries but no obvious N+1 patterns, need deeper analysis"

Turn 2: Query Analysis (Strategy Shift)
Thought: "Simple grep isn't sufficient. Let me analyze actual query execution and patterns"
Action: [bash "go test -v ./internal/db/ -run TestQueries", file_read "internal/db/user.go"]
Observation: "Tests reveal slow queries in user profile loading, complex JOIN operations"

Turn 3: Performance Profiling (Strategy Refinement)
Thought: "I need to profile the actual database interactions to find bottlenecks"
Action: [bash "go test -bench=BenchmarkUserQueries -benchmem", file_read "internal/db/profile.go"]
Observation: "Profiling shows excessive memory allocation and query execution time"

Turn 4: Optimization Implementation (Focused Strategy)
Thought: "Based on profiling, I'll implement specific optimizations: query batching and indexing"
Action: [file_write "internal/db/batch.go", file_write "migrations/add_indexes.sql", bash "go test -bench=. -run=Bench"]
Observation: "40% performance improvement with batching, 60% with new indexes"

Answer: "Implemented query batching and strategic indexing, achieved 75% performance improvement..."
```

## Advanced Multi-Turn Patterns

### 1. **Parallel Investigation with Convergence**

For complex issues requiring multiple investigation paths:

```go
// Pseudo-code for parallel investigation pattern
func (a *Agent) parallelInvestigationPattern(ctx context.Context, issue string) {
    // Turn 1: Launch parallel investigations
    investigations := []Investigation{
        {Name: "performance", Tools: []string{"bash", "file_read"}},
        {Name: "security", Tools: []string{"bash", "file_list"}},
        {Name: "dependencies", Tools: []string{"bash", "file_read"}},
    }
    
    // Execute investigations in parallel
    results := a.executeParallelInvestigations(ctx, investigations)
    
    // Turn 2: Analyze convergent findings
    commonIssues := a.findCommonPatterns(results)
    
    // Turn 3: Deep dive into most critical findings
    criticalPath := a.selectCriticalPath(commonIssues)
    deepAnalysis := a.executeFocusedAnalysis(ctx, criticalPath)
    
    // Turn 4: Synthesize comprehensive solution
    solution := a.synthesizeSolution(commonIssues, deepAnalysis)
}
```

### 2. **Hierarchical Task Decomposition**

For large tasks requiring systematic breakdown:

```go
// Pseudo-code for hierarchical decomposition
func (a *Agent) hierarchicalDecomposition(ctx context.Context, task ComplexTask) {
    // Turn 1: High-level decomposition
    subtasks := a.decomposeTask(task)
    
    for _, subtask := range subtasks {
        // Turn 2: Subtask analysis
        requirements := a.analyzeSubtask(ctx, subtask)
        
        // Turn 3: Implementation planning
        plan := a.createImplementationPlan(requirements)
        
        // Turn 4: Execution with validation
        result := a.executeWithValidation(ctx, plan)
        
        // Turn 5: Integration testing
        a.validateIntegration(ctx, result, task.Context)
    }
    
    // Final turn: System integration and testing
    a.performSystemIntegration(ctx, subtasks)
}
```

### 3. **Adaptive Learning Pattern**

For scenarios where the agent learns from failures and adjusts approach:

```go
// Pseudo-code for adaptive learning
func (a *Agent) adaptiveLearningPattern(ctx context.Context, goal Goal) {
    attempt := 1
    maxAttempts := 5
    
    for attempt <= maxAttempts {
        // Turn N: Attempt implementation
        strategy := a.selectStrategy(goal, attempt)
        result := a.executeStrategy(ctx, strategy)
        
        if result.Success {
            return result
        }
        
        // Turn N+1: Failure analysis and learning
        failureAnalysis := a.analyzeFailure(result)
        a.updateStrategyKnowledge(failureAnalysis)
        
        // Turn N+2: Strategy adjustment
        adjustedGoal := a.adjustGoal(goal, failureAnalysis)
        goal = adjustedGoal
        attempt++
    }
}
```

## Implementation Patterns in Current Agent

### Current Multi-Turn Implementation

The Deep Coding Agent implements basic multi-turn capability:

```go
// From internal/agent/agent.go:234-281
if len(toolResults) > 0 {
    followUpResponse, err := a.generateFollowUpResponse(ctx, toolResults, config)
    
    // Check for additional tool calls in follow-up
    followUpToolCalls, followUpCleanContent := a.parseToolCalls(followUpResponse.Content)
    
    if len(followUpToolCalls) > 0 {
        // Execute additional tool calls (multi-turn)
        followUpResults, err := a.executeToolCalls(ctx, followUpToolCalls, config.AllowedTools)
        
        // Generate final response with all results
        finalResponse, err := a.generateFollowUpResponse(ctx, toolResults, config)
    }
}
```

### Enhanced Multi-Turn Implementation

Proposed enhancements for sophisticated ReAct patterns:

```go
// Enhanced multi-turn with reasoning traces
type ReActTurn struct {
    TurnNumber   int                    `json:"turn_number"`
    Thought      string                 `json:"thought"`
    Action       []ToolCall            `json:"action"`
    Observation  []ToolResult          `json:"observation"`
    NextAction   string                `json:"next_action,omitempty"`
}

type ReActSession struct {
    Goal         string      `json:"goal"`
    Turns        []ReActTurn `json:"turns"`
    Status       string      `json:"status"` // planning, executing, completed, failed
    Strategy     string      `json:"strategy"`
    Context      map[string]interface{} `json:"context"`
}

func (a *Agent) ProcessReActSession(ctx context.Context, goal string, config *types.Config) (*ReActSession, error) {
    session := &ReActSession{
        Goal:    goal,
        Status:  "planning",
        Context: make(map[string]interface{}),
    }
    
    for session.Status != "completed" && len(session.Turns) < config.MaxTurns {
        turn := a.executeTurn(ctx, session, config)
        session.Turns = append(session.Turns, turn)
        
        // Evaluate completion criteria
        session.Status = a.evaluateStatus(session)
        
        // Adapt strategy if needed
        if a.shouldAdaptStrategy(session) {
            session.Strategy = a.adaptStrategy(session)
        }
    }
    
    return session, nil
}
```

## Tool Calling Strategies

### 1. **Intelligent Tool Selection**

```go
type ToolSelector struct {
    contextAnalyzer *ContextAnalyzer
    patterns        map[string][]string
}

func (ts *ToolSelector) SelectTools(context Context, goal string) []string {
    // Pattern matching for tool selection
    if strings.Contains(goal, "analyze") {
        return []string{"file_read", "file_list", "bash"}
    }
    
    if strings.Contains(goal, "implement") {
        return []string{"file_write", "directory_create", "bash"}
    }
    
    if strings.Contains(goal, "debug") {
        return []string{"file_read", "bash", "file_list"}
    }
    
    // Default comprehensive toolset
    return ts.patterns["default"]
}
```

### 2. **Execution Strategy Selection**

```go
func (a *Agent) selectExecutionStrategy(toolCalls []ToolCall) ExecutionStrategy {
    // Check for dependencies
    if a.hasDependencies(toolCalls) {
        return SequentialExecution
    }
    
    // Check for state modification
    if a.hasStateModification(toolCalls) {
        return SequentialExecution
    }
    
    // Check for resource contention
    if a.hasResourceContention(toolCalls) {
        return SequentialExecution
    }
    
    // Default to parallel for efficiency
    return ParallelExecution
}
```

### 3. **Context-Aware Tool Parameter Generation**

```go
func (a *Agent) generateToolParameters(toolName string, context Context) map[string]interface{} {
    switch toolName {
    case "file_read":
        // Prioritize files based on context
        return map[string]interface{}{
            "file_path": a.selectRelevantFile(context),
        }
    
    case "bash":
        // Generate context-appropriate commands
        return map[string]interface{}{
            "command": a.generateContextualCommand(context),
        }
    
    case "file_list":
        // Focus on relevant directories
        return map[string]interface{}{
            "path":      a.selectRelevantDirectory(context),
            "recursive": a.shouldRecurse(context),
        }
    }
    
    return make(map[string]interface{})
}
```

## Error Handling and Recovery Patterns

### 1. **Graceful Degradation**

```go
func (a *Agent) handleToolFailure(toolCall ToolCall, error error) RecoveryAction {
    switch toolCall.Name {
    case "bash":
        if strings.Contains(error.Error(), "command not found") {
            return RecoveryAction{
                Type: "alternative_tool",
                Tool: "file_read", // Read logs instead of running command
            }
        }
    
    case "file_read":
        if strings.Contains(error.Error(), "permission denied") {
            return RecoveryAction{
                Type: "user_guidance",
                Message: "Please check file permissions for: " + toolCall.Args["file_path"].(string),
            }
        }
    }
    
    return RecoveryAction{Type: "retry", MaxAttempts: 3}
}
```

### 2. **Context-Aware Recovery**

```go
func (a *Agent) recoverFromFailure(session *ReActSession, failure ToolFailure) {
    // Analyze failure context
    context := a.analyzeFailureContext(session, failure)
    
    // Generate alternative approach
    alternative := a.generateAlternativeApproach(context)
    
    // Update session strategy
    session.Strategy = alternative.Strategy
    
    // Add recovery turn
    recoveryTurn := ReActTurn{
        TurnNumber:  len(session.Turns) + 1,
        Thought:     alternative.Reasoning,
        Action:      alternative.Actions,
        Observation: []ToolResult{},
    }
    
    session.Turns = append(session.Turns, recoveryTurn)
}
```

## Performance Optimization for Multi-Turn

### 1. **Context Compression**

```go
func (a *Agent) compressContext(session *ReActSession) {
    if len(session.Turns) > a.config.MaxContextTurns {
        // Keep first and last turns, compress middle
        compressed := a.compressMiddleTurns(session.Turns[1:len(session.Turns)-1])
        session.Turns = append(
            session.Turns[:1],
            append(compressed, session.Turns[len(session.Turns)-1:]...)...,
        )
    }
}
```

### 2. **Intelligent Caching**

```go
type ToolResultCache struct {
    cache map[string]ToolResult
    ttl   map[string]time.Time
}

func (trc *ToolResultCache) Get(toolCall ToolCall) (ToolResult, bool) {
    key := trc.generateKey(toolCall)
    if result, exists := trc.cache[key]; exists {
        if time.Now().Before(trc.ttl[key]) {
            return result, true
        }
        delete(trc.cache, key)
        delete(trc.ttl, key)
    }
    return ToolResult{}, false
}
```

### 3. **Predictive Tool Loading**

```go
func (a *Agent) predictNextTools(session *ReActSession) []string {
    // Analyze pattern from previous turns
    pattern := a.analyzeToolPattern(session.Turns)
    
    // Predict likely next tools based on current context
    return a.predictTools(pattern, session.Context)
}
```

## Best Practices and Guidelines

### 1. **Multi-Turn Design Principles**

- **Bounded Iteration**: Set maximum turn limits to prevent infinite loops
- **Progress Validation**: Ensure each turn advances toward the goal
- **Context Preservation**: Maintain relevant context across turns
- **Graceful Termination**: Provide meaningful results even if goal isn't fully achieved

### 2. **Tool Calling Best Practices**

- **Parallel by Default**: Use parallel execution unless dependencies exist
- **Fail Fast**: Identify failures quickly and adapt strategy
- **Resource Efficiency**: Minimize redundant tool calls through caching
- **Security First**: Validate all tool parameters for security implications

### 3. **User Experience Considerations**

- **Progress Transparency**: Show users what the agent is thinking and doing
- **Intermediate Results**: Provide valuable insights even during execution
- **Interruption Handling**: Allow users to modify goals mid-execution
- **Result Explanation**: Explain how conclusions were reached through multi-turn reasoning

## Conclusion

ReAct patterns provide a powerful framework for implementing sophisticated multi-turn tool calling in AI agents. The key to successful implementation lies in:

1. **Structured Reasoning**: Clear thought processes that guide tool selection
2. **Adaptive Execution**: Flexibility to change strategy based on observations
3. **Efficient Tool Orchestration**: Optimal use of parallel and sequential execution
4. **Robust Error Handling**: Graceful recovery from failures and unexpected situations
5. **Context Management**: Intelligent handling of conversation state and memory

The Deep Coding Agent's current implementation provides a solid foundation for these patterns, with opportunities for enhancement in reasoning transparency, adaptive strategy selection, and sophisticated error recovery mechanisms.