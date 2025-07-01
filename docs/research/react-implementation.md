# ReAct Agent Implementation Guide

## Overview

This document provides detailed implementation guidance for transforming the current agent into a sophisticated ReAct (Reasoning and Acting) agent following Claude Code's architecture patterns. The implementation focuses on explicit reasoning traces, multi-turn tool calling, and intelligent strategy adaptation.

## Core ReAct Architecture

### 1. ReAct Agent Structure

```go
// internal/agents/react/agent.go
package react

import (
    "context"
    "fmt"
    "time"
    
    "deep-coding-agent/internal/ai"
    "deep-coding-agent/internal/tools"
    "deep-coding-agent/internal/prompts"
    "deep-coding-agent/internal/memory"
    "deep-coding-agent/pkg/types"
)

type ReActAgent struct {
    id              string
    role            AgentRole
    systemPrompt    *prompts.Template
    toolOrchestrator *tools.Orchestrator
    reasoningEngine *ReasoningEngine
    memoryManager   *memory.Manager
    turnManager     *TurnManager
    
    // Configuration
    config          *ReActConfig
    
    // State
    currentSession  *ReActSession
    metrics         *PerformanceMetrics
}

type ReActConfig struct {
    MaxTurns               int                    `json:"max_turns"`
    ThinkingTimeout        time.Duration          `json:"thinking_timeout"`
    ToolExecutionTimeout   time.Duration          `json:"tool_execution_timeout"`
    StrategyAdaptation     bool                   `json:"strategy_adaptation"`
    ParallelToolExecution  bool                   `json:"parallel_tool_execution"`
    MemoryCompression      bool                   `json:"memory_compression"`
    ReasoningTransparency  bool                   `json:"reasoning_transparency"`
    ContextWindowSize      int                    `json:"context_window_size"`
    AllowedTools           []string               `json:"allowed_tools"`
    SpecializedPrompts     map[string]string      `json:"specialized_prompts"`
}

type AgentRole string
const (
    LeadAgent        AgentRole = "lead"
    AnalysisAgent    AgentRole = "analysis"
    ImplementationAgent AgentRole = "implementation"
    SecurityAgent    AgentRole = "security"
    PerformanceAgent AgentRole = "performance"
    GeneralAgent     AgentRole = "general"
)
```

### 2. ReAct Session Management

```go
// pkg/types/react.go
package types

import (
    "time"
)

type ReActSession struct {
    ID              string                 `json:"id"`
    Goal            string                 `json:"goal"`
    Turns           []ReActTurn           `json:"turns"`
    Status          SessionStatus         `json:"status"`
    Strategy        string                `json:"strategy"`
    Context         map[string]interface{} `json:"context"`
    StartTime       time.Time             `json:"start_time"`
    EndTime         *time.Time            `json:"end_time,omitempty"`
    TotalDuration   time.Duration         `json:"total_duration"`
    AgentRole       AgentRole             `json:"agent_role"`
    ParentSession   string                `json:"parent_session,omitempty"`
    SubSessions     []string              `json:"sub_sessions,omitempty"`
}

type ReActTurn struct {
    TurnNumber      int                    `json:"turn_number"`
    Thought         ThoughtProcess        `json:"thought"`
    Action          ActionPlan            `json:"action"`
    Observation     ObservationResult     `json:"observation"`
    NextAction      string                `json:"next_action,omitempty"`
    Timestamp       time.Time             `json:"timestamp"`
    Duration        time.Duration         `json:"duration"`
    ToolsUsed       []string              `json:"tools_used"`
    SuccessRate     float64               `json:"success_rate"`
}

type ThoughtProcess struct {
    Reasoning       string                `json:"reasoning"`
    Strategy        string                `json:"strategy"`
    Expectations    []string              `json:"expectations"`
    Concerns        []string              `json:"concerns,omitempty"`
    ContextAnalysis string                `json:"context_analysis"`
    DecisionRationale string              `json:"decision_rationale"`
}

type ActionPlan struct {
    ToolCalls       []ToolCall            `json:"tool_calls"`
    ExecutionStrategy string              `json:"execution_strategy"` // parallel, sequential, hybrid
    Priority        []int                 `json:"priority"`
    Dependencies    map[int][]int         `json:"dependencies"`
    ExpectedOutcomes []string             `json:"expected_outcomes"`
    FallbackPlan    *ActionPlan          `json:"fallback_plan,omitempty"`
}

type ObservationResult struct {
    ToolResults     []ToolResult          `json:"tool_results"`
    Analysis        string                `json:"analysis"`
    Insights        []string              `json:"insights"`
    Unexpected      []string              `json:"unexpected,omitempty"`
    NextStepNeeded  bool                  `json:"next_step_needed"`
    CompletionScore float64               `json:"completion_score"`
}

type SessionStatus string
const (
    StatusPlanning   SessionStatus = "planning"
    StatusExecuting  SessionStatus = "executing"
    StatusReflecting SessionStatus = "reflecting"
    StatusCompleted  SessionStatus = "completed"
    StatusFailed     SessionStatus = "failed"
    StatusPaused     SessionStatus = "paused"
)
```

### 3. Reasoning Engine

```go
// internal/agents/react/reasoning.go
package react

import (
    "context"
    "fmt"
    "strings"
    
    "deep-coding-agent/internal/ai"
    "deep-coding-agent/internal/prompts"
    "deep-coding-agent/pkg/types"
)

type ReasoningEngine struct {
    aiProvider      ai.Provider
    promptBuilder   *prompts.Builder
    contextAnalyzer *ContextAnalyzer
    strategySelector *StrategySelector
    
    // Configuration
    maxThinkingTime time.Duration
    transparencyMode bool
}

type ContextAnalyzer struct {
    patterns        map[string]ContextPattern
    historyAnalyzer *HistoryAnalyzer
    goalDecomposer  *GoalDecomposer
}

type StrategySelector struct {
    strategies      map[string]ExecutionStrategy
    performanceData map[string]PerformanceMetrics
    adaptationRules []AdaptationRule
}

func NewReasoningEngine(aiProvider ai.Provider, config *ReActConfig) *ReasoningEngine {
    return &ReasoningEngine{
        aiProvider:      aiProvider,
        promptBuilder:   prompts.NewBuilder(config),
        contextAnalyzer: NewContextAnalyzer(),
        strategySelector: NewStrategySelector(),
        maxThinkingTime: config.ThinkingTimeout,
        transparencyMode: config.ReasoningTransparency,
    }
}

func (re *ReasoningEngine) GenerateThought(ctx context.Context, session *ReActSession, goal string) (*ThoughtProcess, error) {
    // Analyze current context
    contextAnalysis := re.contextAnalyzer.AnalyzeContext(session, goal)
    
    // Build thinking prompt with XML structure
    thinkingPrompt := re.buildThinkingPrompt(session, goal, contextAnalysis)
    
    // Generate reasoning with AI
    aiRequest := &types.AIRequest{
        Prompt:      thinkingPrompt,
        Temperature: 0.7, // Higher temperature for creative thinking
        MaxTokens:   1000,
    }
    
    response, err := re.aiProvider.Generate(aiRequest)
    if err != nil {
        return nil, fmt.Errorf("failed to generate thought: %w", err)
    }
    
    // Parse thought process from response
    thought := re.parseThoughtProcess(response.Content)
    
    return thought, nil
}

func (re *ReasoningEngine) buildThinkingPrompt(session *ReActSession, goal string, context *ContextAnalysis) string {
    return fmt.Sprintf(`<thinking>
<current_goal>%s</current_goal>

<context_analysis>
%s
</context_analysis>

<session_history>
%s
</session_history>

<instructions>
Analyze the current situation and develop a comprehensive thought process:

1. **Reasoning**: What is your understanding of the current situation and goal?
2. **Strategy**: What approach will you take to achieve the goal?
3. **Expectations**: What outcomes do you expect from your planned actions?
4. **Concerns**: What potential issues or challenges do you foresee?
5. **Context Analysis**: How does the current context influence your approach?
6. **Decision Rationale**: Why is this the best approach given the circumstances?

Think step-by-step and be explicit about your reasoning process.
</instructions>
</thinking>`, goal, context.Summary, re.formatSessionHistory(session))
}

func (re *ReasoningEngine) parseThoughtProcess(content string) *ThoughtProcess {
    // Parse structured thought process from AI response
    thought := &ThoughtProcess{}
    
    // Extract reasoning
    if reasoning := re.extractSection(content, "reasoning"); reasoning != "" {
        thought.Reasoning = reasoning
    }
    
    // Extract strategy
    if strategy := re.extractSection(content, "strategy"); strategy != "" {
        thought.Strategy = strategy
    }
    
    // Extract expectations
    thought.Expectations = re.extractList(content, "expectations")
    
    // Extract concerns
    thought.Concerns = re.extractList(content, "concerns")
    
    // Extract context analysis
    if contextAnalysis := re.extractSection(content, "context_analysis"); contextAnalysis != "" {
        thought.ContextAnalysis = contextAnalysis
    }
    
    // Extract decision rationale
    if rationale := re.extractSection(content, "decision_rationale"); rationale != "" {
        thought.DecisionRationale = rationale
    }
    
    return thought
}
```

### 4. Action Planning

```go
// internal/agents/react/action_planner.go
package react

import (
    "context"
    "fmt"
    
    "deep-coding-agent/internal/tools"
    "deep-coding-agent/pkg/types"
)

type ActionPlanner struct {
    toolOrchestrator *tools.Orchestrator
    strategySelector *StrategySelector
    dependencyAnalyzer *DependencyAnalyzer
    
    // Configuration
    parallelEnabled bool
    maxToolCalls   int
}

type DependencyAnalyzer struct {
    dependencies map[string][]string
    conflicts    map[string][]string
}

func (ap *ActionPlanner) PlanAction(ctx context.Context, thought *ThoughtProcess, availableTools []string) (*ActionPlan, error) {
    // Generate tool calls based on thought process
    toolCalls, err := ap.generateToolCalls(thought, availableTools)
    if err != nil {
        return nil, fmt.Errorf("failed to generate tool calls: %w", err)
    }
    
    // Analyze dependencies between tool calls
    dependencies := ap.dependencyAnalyzer.AnalyzeDependencies(toolCalls)
    
    // Select execution strategy
    strategy := ap.selectExecutionStrategy(toolCalls, dependencies)
    
    // Create priority ordering
    priority := ap.calculatePriority(toolCalls, dependencies)
    
    // Generate expected outcomes
    expectedOutcomes := ap.generateExpectedOutcomes(toolCalls, thought)
    
    // Create fallback plan if needed
    fallbackPlan := ap.createFallbackPlan(toolCalls, strategy)
    
    return &ActionPlan{
        ToolCalls:        toolCalls,
        ExecutionStrategy: strategy,
        Priority:         priority,
        Dependencies:     dependencies,
        ExpectedOutcomes: expectedOutcomes,
        FallbackPlan:     fallbackPlan,
    }, nil
}

func (ap *ActionPlanner) generateToolCalls(thought *ThoughtProcess, availableTools []string) ([]types.ToolCall, error) {
    // Build action planning prompt
    actionPrompt := ap.buildActionPrompt(thought, availableTools)
    
    // Use AI to generate tool calls
    response, err := ap.aiProvider.Generate(&types.AIRequest{
        Prompt:      actionPrompt,
        Temperature: 0.3, // Lower temperature for structured output
        MaxTokens:   2000,
    })
    if err != nil {
        return nil, err
    }
    
    // Parse tool calls from response
    toolCalls := ap.parseToolCalls(response.Content)
    
    return toolCalls, nil
}

func (ap *ActionPlanner) buildActionPrompt(thought *ThoughtProcess, availableTools []string) string {
    return fmt.Sprintf(`<action_planning>
<reasoning>%s</reasoning>
<strategy>%s</strategy>

<available_tools>
%s
</available_tools>

<instructions>
Based on your reasoning and strategy, plan the specific actions to take:

1. Select the appropriate tools to use
2. Determine the optimal execution order
3. Consider dependencies between tools
4. Plan for parallel execution where beneficial

Use this exact format for tool calls:
<|FunctionCallBegin|>
[
  {"name": "tool_name", "parameters": {"param": "value"}},
  {"name": "another_tool", "parameters": {"param": "value"}}
]
<|FunctionCallEnd|>

Consider these execution strategies:
- **parallel**: Independent tools that can run simultaneously
- **sequential**: Tools with dependencies that must run in order
- **hybrid**: Mix of parallel and sequential execution
</instructions>
</action_planning>`, thought.Reasoning, thought.Strategy, strings.Join(availableTools, ", "))
}
```

### 5. Observation and Analysis

```go
// internal/agents/react/observer.go
package react

import (
    "context"
    "fmt"
    "strings"
    
    "deep-coding-agent/pkg/types"
)

type Observer struct {
    aiProvider      ai.Provider
    resultAnalyzer  *ResultAnalyzer
    insightExtractor *InsightExtractor
    completionChecker *CompletionChecker
    
    // Configuration
    analysisDepth   int
    insightThreshold float64
}

type ResultAnalyzer struct {
    patterns        map[string]ResultPattern
    anomalyDetector *AnomalyDetector
    qualityAssessor *QualityAssessor
}

type InsightExtractor struct {
    knowledgeBase   *KnowledgeBase
    patternMatcher  *PatternMatcher
    correlationEngine *CorrelationEngine
}

func (o *Observer) ObserveResults(ctx context.Context, action *ActionPlan, toolResults []types.ToolResult, thought *ThoughtProcess) (*ObservationResult, error) {
    // Analyze tool execution results
    analysis := o.resultAnalyzer.AnalyzeResults(toolResults, action.ExpectedOutcomes)
    
    // Extract insights from results
    insights := o.insightExtractor.ExtractInsights(toolResults, thought)
    
    // Identify unexpected outcomes
    unexpected := o.identifyUnexpectedOutcomes(toolResults, action.ExpectedOutcomes)
    
    // Check if next step is needed
    nextStepNeeded := o.completionChecker.CheckCompletion(toolResults, thought)
    
    // Calculate completion score
    completionScore := o.calculateCompletionScore(toolResults, action.ExpectedOutcomes)
    
    // Generate comprehensive observation
    observationPrompt := o.buildObservationPrompt(toolResults, analysis, thought)
    
    response, err := o.aiProvider.Generate(&types.AIRequest{
        Prompt:      observationPrompt,
        Temperature: 0.4,
        MaxTokens:   1500,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to generate observation: %w", err)
    }
    
    // Parse and enhance analysis
    enhancedAnalysis := o.parseObservationAnalysis(response.Content)
    
    return &ObservationResult{
        ToolResults:     toolResults,
        Analysis:        enhancedAnalysis,
        Insights:        insights,
        Unexpected:      unexpected,
        NextStepNeeded:  nextStepNeeded,
        CompletionScore: completionScore,
    }, nil
}

func (o *Observer) buildObservationPrompt(toolResults []types.ToolResult, analysis string, thought *ThoughtProcess) string {
    return fmt.Sprintf(`<observation>
<original_thought>
<reasoning>%s</reasoning>
<expectations>%s</expectations>
</original_thought>

<tool_results>
%s
</tool_results>

<preliminary_analysis>
%s
</preliminary_analysis>

<instructions>
Analyze the tool execution results and provide comprehensive observations:

1. **Analysis**: How do the results compare to your expectations?
2. **Insights**: What new information or patterns do you observe?
3. **Unexpected**: What outcomes were different from what you expected?
4. **Progress Assessment**: How much progress was made toward the goal?
5. **Next Steps**: What should be done next based on these results?

Be thorough in your analysis and identify both successes and areas that need attention.
</instructions>
</observation>`, thought.Reasoning, strings.Join(thought.Expectations, "; "), o.formatToolResults(toolResults), analysis)
}
```

### 6. Turn Management

```go
// internal/agents/react/turn_manager.go
package react

import (
    "context"
    "fmt"
    "time"
    
    "deep-coding-agent/pkg/types"
)

type TurnManager struct {
    maxTurns        int
    currentTurn     int
    progressTracker *ProgressTracker
    completionChecker *CompletionChecker
    adaptationEngine *AdaptationEngine
    
    // State
    turnHistory     []TurnMetrics
    averageDuration time.Duration
    successRate     float64
}

type ProgressTracker struct {
    goalProgress    map[string]float64
    milestones      []Milestone
    completionCriteria []CompletionCriterion
}

type AdaptationEngine struct {
    adaptationRules []AdaptationRule
    strategyHistory []StrategyChange
    performanceMetrics *PerformanceMetrics
}

type TurnMetrics struct {
    TurnNumber      int           `json:"turn_number"`
    Duration        time.Duration `json:"duration"`
    ToolsExecuted   int           `json:"tools_executed"`
    SuccessRate     float64       `json:"success_rate"`
    ProgressMade    float64       `json:"progress_made"`
    Strategy        string        `json:"strategy"`
    Adaptations     int           `json:"adaptations"`
}

func NewTurnManager(config *ReActConfig) *TurnManager {
    return &TurnManager{
        maxTurns:        config.MaxTurns,
        currentTurn:     0,
        progressTracker: NewProgressTracker(),
        completionChecker: NewCompletionChecker(),
        adaptationEngine: NewAdaptationEngine(),
        turnHistory:     make([]TurnMetrics, 0),
    }
}

func (tm *TurnManager) ExecuteTurn(ctx context.Context, session *ReActSession, agent *ReActAgent) (*types.ReActTurn, error) {
    startTime := time.Now()
    tm.currentTurn++
    
    // Check if we've exceeded max turns
    if tm.currentTurn > tm.maxTurns {
        return nil, fmt.Errorf("maximum turns (%d) exceeded", tm.maxTurns)
    }
    
    // Generate thought process
    thought, err := agent.reasoningEngine.GenerateThought(ctx, session, session.Goal)
    if err != nil {
        return nil, fmt.Errorf("failed to generate thought: %w", err)
    }
    
    // Plan actions based on thought
    action, err := agent.actionPlanner.PlanAction(ctx, thought, agent.config.AllowedTools)
    if err != nil {
        return nil, fmt.Errorf("failed to plan action: %w", err)
    }
    
    // Execute tools according to plan
    toolResults, err := agent.toolOrchestrator.ExecutePlan(ctx, action)
    if err != nil {
        return nil, fmt.Errorf("failed to execute tools: %w", err)
    }
    
    // Observe and analyze results
    observation, err := agent.observer.ObserveResults(ctx, action, toolResults, thought)
    if err != nil {
        return nil, fmt.Errorf("failed to observe results: %w", err)
    }
    
    // Create turn record
    turn := &types.ReActTurn{
        TurnNumber:   tm.currentTurn,
        Thought:      *thought,
        Action:       *action,
        Observation:  *observation,
        Timestamp:    startTime,
        Duration:     time.Since(startTime),
        ToolsUsed:    tm.extractToolNames(action.ToolCalls),
        SuccessRate:  tm.calculateTurnSuccessRate(toolResults),
    }
    
    // Determine next action if needed
    if observation.NextStepNeeded {
        turn.NextAction = tm.determineNextAction(observation, session)
    }
    
    // Record turn metrics
    tm.recordTurnMetrics(turn)
    
    // Check for strategy adaptation
    if tm.shouldAdaptStrategy(session, turn) {
        tm.adaptStrategy(session, agent)
    }
    
    return turn, nil
}

func (tm *TurnManager) shouldAdaptStrategy(session *ReActSession, turn *types.ReActTurn) bool {
    // Check if progress is stalling
    if len(session.Turns) >= 3 {
        recentProgress := tm.calculateRecentProgress(session.Turns[len(session.Turns)-3:])
        if recentProgress < 0.1 { // Less than 10% progress in last 3 turns
            return true
        }
    }
    
    // Check if success rate is declining
    if turn.SuccessRate < 0.6 {
        return true
    }
    
    // Check if unexpected outcomes are frequent
    if len(turn.Observation.Unexpected) > 2 {
        return true
    }
    
    return false
}

func (tm *TurnManager) adaptStrategy(session *ReActSession, agent *ReActAgent) {
    // Analyze current strategy effectiveness
    currentStrategy := session.Strategy
    effectiveness := tm.calculateStrategyEffectiveness(session.Turns)
    
    // Select new strategy based on performance
    newStrategy := tm.adaptationEngine.SelectNewStrategy(currentStrategy, effectiveness, session.Context)
    
    // Update session strategy
    session.Strategy = newStrategy
    
    // Update agent configuration if needed
    tm.updateAgentConfiguration(agent, newStrategy)
    
    // Record strategy change
    tm.adaptationEngine.RecordStrategyChange(currentStrategy, newStrategy, effectiveness)
}
```

### 7. Main ReAct Agent Implementation

```go
// internal/agents/react/agent.go (Main implementation)
package react

import (
    "context"
    "fmt"
    "time"
    
    "deep-coding-agent/internal/config"
    "deep-coding-agent/internal/ai"
    "deep-coding-agent/internal/tools"
    "deep-coding-agent/internal/prompts"
    "deep-coding-agent/internal/memory"
    "deep-coding-agent/pkg/types"
)

func NewReActAgent(configManager *config.Manager, role AgentRole) (*ReActAgent, error) {
    config, err := loadReActConfig(configManager, role)
    if err != nil {
        return nil, fmt.Errorf("failed to load ReAct config: %w", err)
    }
    
    // Initialize AI provider
    aiProvider, err := initializeAIProvider(configManager)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize AI provider: %w", err)
    }
    
    // Initialize components
    reasoningEngine := NewReasoningEngine(aiProvider, config)
    toolOrchestrator := tools.NewOrchestrator(configManager)
    memoryManager := memory.NewManager(config.ContextWindowSize)
    turnManager := NewTurnManager(config)
    
    // Create prompt template for role
    promptTemplate, err := prompts.LoadTemplate(role)
    if err != nil {
        return nil, fmt.Errorf("failed to load prompt template: %w", err)
    }
    
    agent := &ReActAgent{
        id:               generateAgentID(role),
        role:             role,
        systemPrompt:     promptTemplate,
        toolOrchestrator: toolOrchestrator,
        reasoningEngine:  reasoningEngine,
        memoryManager:    memoryManager,
        turnManager:      turnManager,
        config:           config,
        metrics:          NewPerformanceMetrics(),
    }
    
    return agent, nil
}

func (ra *ReActAgent) ProcessGoal(ctx context.Context, goal string) (*types.ReActSession, error) {
    // Create new session
    session := &types.ReActSession{
        ID:        generateSessionID(),
        Goal:      goal,
        Status:    types.StatusPlanning,
        Strategy:  ra.selectInitialStrategy(goal),
        Context:   make(map[string]interface{}),
        StartTime: time.Now(),
        AgentRole: ra.role,
    }
    
    ra.currentSession = session
    
    // Initialize session context
    if err := ra.initializeSessionContext(session); err != nil {
        return nil, fmt.Errorf("failed to initialize session context: %w", err)
    }
    
    // Main ReAct loop
    for session.Status != types.StatusCompleted && session.Status != types.StatusFailed {
        // Check for context limitations
        if len(session.Turns) >= ra.config.MaxTurns {
            session.Status = types.StatusFailed
            break
        }
        
        // Execute single turn
        session.Status = types.StatusExecuting
        turn, err := ra.turnManager.ExecuteTurn(ctx, session, ra)
        if err != nil {
            session.Status = types.StatusFailed
            return session, fmt.Errorf("turn execution failed: %w", err)
        }
        
        // Add turn to session
        session.Turns = append(session.Turns, *turn)
        
        // Update session status based on observation
        session.Status = ra.evaluateSessionStatus(turn.Observation)
        
        // Manage memory if needed
        if ra.config.MemoryCompression && len(session.Turns) > ra.config.ContextWindowSize {
            if err := ra.memoryManager.CompressContext(session); err != nil {
                // Log warning but continue
                fmt.Printf("Warning: memory compression failed: %v\n", err)
            }
        }
        
        // Update metrics
        ra.metrics.RecordTurn(turn)
    }
    
    // Finalize session
    ra.finalizeSession(session)
    
    return session, nil
}

func (ra *ReActAgent) evaluateSessionStatus(observation types.ObservationResult) types.SessionStatus {
    // Check completion score
    if observation.CompletionScore >= 0.95 {
        return types.StatusCompleted
    }
    
    // Check if next step is needed
    if !observation.NextStepNeeded && observation.CompletionScore >= 0.8 {
        return types.StatusCompleted
    }
    
    // Check for failure indicators
    if observation.CompletionScore < 0.1 && len(observation.Unexpected) > 3 {
        return types.StatusFailed
    }
    
    // Continue execution
    return types.StatusExecuting
}

func (ra *ReActAgent) selectInitialStrategy(goal string) string {
    // Analyze goal complexity and type
    if strings.Contains(strings.ToLower(goal), "analyze") {
        return "analysis_focused"
    }
    
    if strings.Contains(strings.ToLower(goal), "implement") || strings.Contains(strings.ToLower(goal), "create") {
        return "implementation_focused"
    }
    
    if strings.Contains(strings.ToLower(goal), "debug") || strings.Contains(strings.ToLower(goal), "fix") {
        return "debugging_focused"
    }
    
    if strings.Contains(strings.ToLower(goal), "optimize") || strings.Contains(strings.ToLower(goal), "performance") {
        return "optimization_focused"
    }
    
    // Default strategy
    return "adaptive_general"
}

func (ra *ReActAgent) initializeSessionContext(session *types.ReActSession) error {
    // Get current working directory
    workingDir, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("failed to get working directory: %w", err)
    }
    
    session.Context["working_directory"] = workingDir
    session.Context["agent_role"] = string(ra.role)
    session.Context["strategy"] = session.Strategy
    session.Context["available_tools"] = ra.config.AllowedTools
    session.Context["start_time"] = session.StartTime
    
    return nil
}

func (ra *ReActAgent) finalizeSession(session *types.ReActSession) {
    endTime := time.Now()
    session.EndTime = &endTime
    session.TotalDuration = endTime.Sub(session.StartTime)
    
    // Update metrics
    ra.metrics.RecordSession(session)
    
    // Save session if persistence is enabled
    if ra.config.PersistSessions {
        if err := ra.memoryManager.SaveSession(session); err != nil {
            fmt.Printf("Warning: failed to save session: %v\n", err)
        }
    }
}

// ProcessMessage provides backward compatibility with existing agent interface
func (ra *ReActAgent) ProcessMessage(ctx context.Context, message string, config *types.Config) (*types.Response, error) {
    // Convert message to goal and process with ReAct pattern
    session, err := ra.ProcessGoal(ctx, message)
    if err != nil {
        return nil, err
    }
    
    // Convert ReAct session to legacy response format
    return ra.convertToLegacyResponse(session), nil
}

func (ra *ReActAgent) convertToLegacyResponse(session *types.ReActSession) *types.Response {
    if len(session.Turns) == 0 {
        return &types.Response{
            Message: &types.Message{
                Role:    "assistant",
                Content: "No response generated",
            },
            Complete: true,
        }
    }
    
    lastTurn := session.Turns[len(session.Turns)-1]
    
    // Aggregate tool results from all turns
    var allToolResults []types.ToolResult
    for _, turn := range session.Turns {
        allToolResults = append(allToolResults, turn.Observation.ToolResults...)
    }
    
    // Create response content combining reasoning and observations
    content := ra.formatReActResponse(session)
    
    return &types.Response{
        Message: &types.Message{
            Role:      "assistant",
            Content:   content,
            ToolCalls: ra.convertToolCalls(lastTurn.Action.ToolCalls),
        },
        ToolResults: allToolResults,
        SessionID:   session.ID,
        Complete:    session.Status == types.StatusCompleted,
    }
}

func (ra *ReActAgent) formatReActResponse(session *types.ReActSession) string {
    var response strings.Builder
    
    // Add final analysis if transparent mode is enabled
    if ra.config.ReasoningTransparency && len(session.Turns) > 0 {
        lastTurn := session.Turns[len(session.Turns)-1]
        response.WriteString(fmt.Sprintf("**Analysis**: %s\n\n", lastTurn.Observation.Analysis))
        
        if len(lastTurn.Observation.Insights) > 0 {
            response.WriteString("**Key Insights**:\n")
            for _, insight := range lastTurn.Observation.Insights {
                response.WriteString(fmt.Sprintf("- %s\n", insight))
            }
            response.WriteString("\n")
        }
    }
    
    // Add main response content
    if len(session.Turns) > 0 {
        lastTurn := session.Turns[len(session.Turns)-1]
        response.WriteString(lastTurn.Observation.Analysis)
    }
    
    return response.String()
}
```

### 8. Integration with Existing System

```go
// internal/agents/agent_factory.go
package agents

import (
    "deep-coding-agent/internal/agents/react"
    "deep-coding-agent/internal/agents/legacy"
    "deep-coding-agent/internal/config"
    "deep-coding-agent/pkg/types"
)

type AgentFactory struct {
    configManager *config.Manager
}

func NewAgentFactory(configManager *config.Manager) *AgentFactory {
    return &AgentFactory{
        configManager: configManager,
    }
}

func (af *AgentFactory) CreateAgent() (Agent, error) {
    config, err := af.configManager.GetConfig()
    if err != nil {
        return nil, err
    }
    
    // Check if ReAct mode is enabled
    if config.EnableReActMode {
        return react.NewReActAgent(af.configManager, react.GeneralAgent)
    }
    
    // Use legacy agent for backward compatibility
    return legacy.NewAgent(af.configManager)
}

type Agent interface {
    ProcessMessage(ctx context.Context, message string, config *types.Config) (*types.Response, error)
    GetAvailableTools() []string
    GetSessionHistory() []*types.Message
}
```

## Usage Examples

### 1. Basic ReAct Session

```go
// Example: Code analysis task
agent, err := react.NewReActAgent(configManager, react.AnalysisAgent)
if err != nil {
    log.Fatal(err)
}

session, err := agent.ProcessGoal(ctx, "Analyze the performance bottlenecks in this Go application")
if err != nil {
    log.Fatal(err)
}

// Access reasoning traces
for i, turn := range session.Turns {
    fmt.Printf("Turn %d:\n", i+1)
    fmt.Printf("Thought: %s\n", turn.Thought.Reasoning)
    fmt.Printf("Action: %d tools executed\n", len(turn.Action.ToolCalls))
    fmt.Printf("Observation: %s\n", turn.Observation.Analysis)
    fmt.Printf("Completion: %.2f%%\n\n", turn.Observation.CompletionScore*100)
}
```

### 2. Specialized Agent Usage

```go
// Security-focused analysis
securityAgent, err := react.NewReActAgent(configManager, react.SecurityAgent)
if err != nil {
    log.Fatal(err)
}

session, err := securityAgent.ProcessGoal(ctx, "Perform security audit of the authentication system")
if err != nil {
    log.Fatal(err)
}

// The security agent will use specialized prompts and tools
// for security-focused analysis
```

### 3. Configuration Options

```go
// Custom ReAct configuration
config := &react.ReActConfig{
    MaxTurns:              10,
    ThinkingTimeout:       30 * time.Second,
    ToolExecutionTimeout:  2 * time.Minute,
    StrategyAdaptation:    true,
    ParallelToolExecution: true,
    MemoryCompression:     true,
    ReasoningTransparency: true,
    ContextWindowSize:     50,
    AllowedTools:          []string{"file_read", "bash", "file_list"},
}

agent, err := react.NewReActAgentWithConfig(configManager, react.GeneralAgent, config)
```

## Performance Considerations

### 1. Memory Management
- Implement context compression for long sessions
- Use relevance scoring to prioritize important information
- Automatic session cleanup based on age and size

### 2. Tool Execution Optimization
- Intelligent parallel/sequential execution
- Tool result caching for repeated operations
- Predictive tool loading based on patterns

### 3. Reasoning Efficiency
- Thought process caching for similar contexts
- Strategy adaptation based on performance metrics
- Timeout management for thinking and execution phases

## Testing Strategy

### 1. Unit Tests
- Test each ReAct component independently
- Mock AI provider responses for consistent testing
- Validate reasoning trace parsing and formatting

### 2. Integration Tests
- End-to-end ReAct session testing
- Multi-turn scenario validation
- Performance benchmarking

### 3. Behavioral Tests
- Strategy adaptation testing
- Error recovery scenarios
- Memory management validation

This implementation provides a complete ReAct agent system that maintains backward compatibility while offering advanced reasoning capabilities and multi-turn intelligence.