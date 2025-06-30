package types

import (
	"time"
)

// ReActSession represents a ReAct reasoning session
type ReActSession struct {
	ID           string           `json:"id"`
	TaskID       string           `json:"taskId"`
	StartTime    time.Time        `json:"startTime"`
	EndTime      *time.Time       `json:"endTime,omitempty"`
	Status       SessionStatus    `json:"status"`
	Turns        []ReActTurn      `json:"turns"`
	FinalResult  *AgentResponse   `json:"finalResult,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	MaxTurns     int              `json:"maxTurns"`     // Maximum allowed turns
	CurrentTurn  int              `json:"currentTurn"`  // Current turn number
}

// SessionStatus represents the status of a ReAct session
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusFailed    SessionStatus = "failed"
	SessionStatusTimeout   SessionStatus = "timeout"
	SessionStatusAborted   SessionStatus = "aborted"
)

// ReActTurn represents a single turn in the ReAct process
type ReActTurn struct {
	ID          string           `json:"id"`
	TurnNumber  int              `json:"turnNumber"`
	Thought     *ThoughtProcess  `json:"thought,omitempty"`
	Action      *ActionPlan      `json:"action,omitempty"`
	Observation *ObservationResult `json:"observation,omitempty"`
	Status      TurnStatus       `json:"status"`
	StartTime   time.Time        `json:"startTime"`
	EndTime     *time.Time       `json:"endTime,omitempty"`
	Duration    time.Duration    `json:"duration"`
	Error       *AgentError      `json:"error,omitempty"`
}

// TurnStatus represents the status of a ReAct turn
type TurnStatus string

const (
	TurnStatusThinking   TurnStatus = "thinking"
	TurnStatusPlanning   TurnStatus = "planning"
	TurnStatusExecuting  TurnStatus = "executing"
	TurnStatusObserving  TurnStatus = "observing"
	TurnStatusCompleted  TurnStatus = "completed"
	TurnStatusFailed     TurnStatus = "failed"
)

// ThoughtProcess represents the thinking phase of ReAct
type ThoughtProcess struct {
	ID             string            `json:"id"`
	SessionID      string            `json:"sessionId"`      // Session this thought belongs to
	TaskID         string            `json:"taskId"`         // Task this thought is for
	Content        string            `json:"content"`        // The actual thought
	Analysis       string            `json:"analysis"`       // Analysis of the situation
	Strategy       string            `json:"strategy"`       // Chosen strategy
	Reasoning      []string          `json:"reasoning"`      // Step-by-step reasoning
	Confidence     float64           `json:"confidence"`     // Confidence in the thought (0.0-1.0)
	Alternatives   []string          `json:"alternatives"`   // Alternative approaches considered
	Context        map[string]interface{} `json:"context"`   // Contextual information used
	Duration       time.Duration     `json:"duration"`
	TokensUsed     int               `json:"tokensUsed"`
}

// ActionPlan represents the action planning phase of ReAct
type ActionPlan struct {
	ID           string              `json:"id"`
	Strategy     ExecutionStrategy   `json:"strategy"`
	Tools        []PlannedToolCall   `json:"tools"`
	Steps        []ActionStep        `json:"steps"`
	ExpectedResult string            `json:"expectedResult"`
	Rationale    string              `json:"rationale"`
	Confidence   float64             `json:"confidence"`     // Confidence in the plan (0.0-1.0)
	Priority     int                 `json:"priority"`       // 1-10, higher = more important
	EstimatedTime time.Duration      `json:"estimatedTime"`
	Dependencies []string            `json:"dependencies"`   // Dependencies between steps
	Alternatives []ActionPlan        `json:"alternatives"`   // Alternative plans
}

// PlannedToolCall represents a tool call in the action plan
type PlannedToolCall struct {
	ToolName    string                 `json:"toolName"`
	Arguments   map[string]interface{} `json:"arguments"`
	Rationale   string                 `json:"rationale"`
	Priority    int                    `json:"priority"`
	Parallel    bool                   `json:"parallel"`    // Can be executed in parallel
	Optional    bool                   `json:"optional"`    // Optional tool call
	Fallback    *PlannedToolCall       `json:"fallback,omitempty"` // Fallback if this fails
}

// ActionStep represents a single step in the action plan
type ActionStep struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Type        StepType  `json:"type"`
	Status      StepStatus `json:"status"`
	Dependencies []string `json:"dependencies"` // IDs of steps this depends on
	Tools       []string  `json:"tools"`        // Tools required for this step
	EstimatedTime time.Duration `json:"estimatedTime"`
	ActualTime    time.Duration `json:"actualTime"`
	Result        interface{}   `json:"result,omitempty"`
	Error         *AgentError   `json:"error,omitempty"`
}

// StepType represents the type of action step
type StepType string

const (
	StepTypeToolCall    StepType = "tool_call"
	StepTypeAnalysis    StepType = "analysis"
	StepTypeValidation  StepType = "validation"
	StepTypeTransform   StepType = "transform"
	StepTypeOutput      StepType = "output"
)

// StepStatus represents the status of an action step
type StepStatus string

const (
	StepStatusPending    StepStatus = "pending"
	StepStatusExecuting  StepStatus = "executing"
	StepStatusCompleted  StepStatus = "completed"
	StepStatusFailed     StepStatus = "failed"
	StepStatusSkipped    StepStatus = "skipped"
)

// ObservationResult represents the observation phase of ReAct
type ObservationResult struct {
	ID             string              `json:"id"`
	Summary        string              `json:"summary"`
	ToolResults    []ToolResult        `json:"toolResults"`
	Analysis       string              `json:"analysis"`
	Success        bool                `json:"success"`
	NextAction     string              `json:"nextAction"`     // What to do next
	Confidence     float64             `json:"confidence"`     // Confidence in observation (0.0-1.0)
	Quality        *QualityScore       `json:"quality,omitempty"`
	Insights       []string            `json:"insights"`       // Key insights discovered
	Problems       []string            `json:"problems"`       // Problems encountered
	Recommendations []string           `json:"recommendations"` // Recommendations for next steps
	Metrics        *ObservationMetrics `json:"metrics,omitempty"`
}

// ObservationMetrics represents metrics about the observation
type ObservationMetrics struct {
	DataProcessed   int64         `json:"dataProcessed"`   // Bytes of data processed
	TimeElapsed     time.Duration `json:"timeElapsed"`
	ToolsExecuted   int           `json:"toolsExecuted"`
	ErrorsEncountered int         `json:"errorsEncountered"`
	SuccessRate     float64       `json:"successRate"`
}

// ReActStrategy represents different ReAct execution strategies
type ReActStrategy string

const (
	ReActStrategyStandard    ReActStrategy = "standard"    // Standard ReAct loop
	ReActStrategyOptimized   ReActStrategy = "optimized"   // Optimized for performance
	ReActStrategyConservative ReActStrategy = "conservative" // Conservative, fewer risks
	ReActStrategyAggressive  ReActStrategy = "aggressive"  // Aggressive, more parallel
	ReActStrategyDebug       ReActStrategy = "debug"       // Debug mode with extra logging
)

// ReActConfig represents configuration for ReAct processing
type ReActConfig struct {
	MaxTurns          int             `json:"maxTurns"`
	MaxThinkingTime   time.Duration   `json:"maxThinkingTime"`
	MaxExecutionTime  time.Duration   `json:"maxExecutionTime"`
	Strategy          ReActStrategy   `json:"strategy"`
	ParallelExecution bool            `json:"parallelExecution"`
	EnableFallback    bool            `json:"enableFallback"`
	ConfidenceThreshold float64       `json:"confidenceThreshold"` // Minimum confidence to proceed
	AutoRetry         bool            `json:"autoRetry"`
	MaxRetries        int             `json:"maxRetries"`
	LoggingLevel      string          `json:"loggingLevel"`
	Temperature       float64         `json:"temperature"`         // AI temperature setting
	MaxTokens         int             `json:"maxTokens"`           // Maximum tokens for AI responses
}

// ThinkingPrompt represents prompts used in the thinking phase
type ThinkingPrompt struct {
	TaskType    TaskType `json:"taskType"`
	Template    string   `json:"template"`
	Variables   map[string]string `json:"variables"`
	Instructions []string `json:"instructions"`
	Examples    []string `json:"examples"`
}

// ReActMetrics represents performance metrics for ReAct sessions
type ReActMetrics struct {
	SessionID         string        `json:"sessionId"`
	TotalTurns        int           `json:"totalTurns"`
	SuccessfulTurns   int           `json:"successfulTurns"`
	TotalDuration     time.Duration `json:"totalDuration"`
	AverageThinkTime  time.Duration `json:"averageThinkTime"`
	AverageActionTime time.Duration `json:"averageActionTime"`
	AverageObserveTime time.Duration `json:"averageObserveTime"`
	ToolCallsTotal    int           `json:"toolCallsTotal"`
	ToolCallsSuccessful int         `json:"toolCallsSuccessful"`
	TokensUsed        int           `json:"tokensUsed"`
	FinalSuccess      bool          `json:"finalSuccess"`
	QualityScore      *QualityScore `json:"qualityScore,omitempty"`
}

// UnifiedAgentConfig represents configuration for the unified agent
type UnifiedAgentConfig struct {
	ReActConfig       *ReActConfig              `json:"reactConfig"`
	ToolConfig        *ToolSystemConfig         `json:"toolConfig"`
	SecurityConfig    *SecurityManagerConfig    `json:"securityConfig"`
	ContextConfig     *ContextManagerConfig     `json:"contextConfig"`
	MemoryConfig      *MemoryManagerConfig      `json:"memoryConfig"`
	MaxConcurrency    int                       `json:"maxConcurrency"`
	Timeout           time.Duration             `json:"timeout"`
	DebugMode         bool                      `json:"debugMode"`
	MetricsEnabled    bool                      `json:"metricsEnabled"`
}

// AgentMetrics represents overall agent performance metrics
type AgentMetrics struct {
	AgentID           string        `json:"agentId"`
	Uptime            time.Duration `json:"uptime"`
	TasksProcessed    int           `json:"tasksProcessed"`
	TasksSuccessful   int           `json:"tasksSuccessful"`
	AverageTaskTime   time.Duration `json:"averageTaskTime"`
	MemoryUsage       int64         `json:"memoryUsage"`       // bytes
	CPUUsage          float64       `json:"cpuUsage"`          // percentage
	ActiveSessions    int           `json:"activeSessions"`
	TotalSessions     int           `json:"totalSessions"`
	ErrorCount        int           `json:"errorCount"`
	LastActivity      time.Time     `json:"lastActivity"`
	ReasoningMetrics  *ReasoningMetrics `json:"reasoningMetrics"`
}

// ReasoningMetrics represents metrics specific to reasoning processes
type ReasoningMetrics struct {
	TotalThoughts     int           `json:"totalThoughts"`
	AverageThinkTime  time.Duration `json:"averageThinkTime"`
	AverageConfidence float64       `json:"averageConfidence"`
	SuccessfulPlans   int           `json:"successfulPlans"`
	FailedPlans       int           `json:"failedPlans"`
	InsightsGenerated int           `json:"insightsGenerated"`
	PatternRecognition float64      `json:"patternRecognition"` // 0.0-1.0
	LearningRate      float64       `json:"learningRate"`       // 0.0-1.0
}

// Insight represents a piece of insight discovered during processing
type Insight struct {
	ID          string                 `json:"id"`
	Type        InsightType            `json:"type"`
	Content     string                 `json:"content"`
	Confidence  float64                `json:"confidence"`  // 0.0-1.0
	Relevance   float64                `json:"relevance"`   // 0.0-1.0
	Source      string                 `json:"source"`      // Where this insight came from
	Context     map[string]interface{} `json:"context"`     // Additional context
	Tags        []string               `json:"tags"`
	CreatedAt   time.Time              `json:"createdAt"`
	UsageCount  int                    `json:"usageCount"`  // How often this insight has been used
}

// InsightType represents the type of insight
type InsightType string

const (
	InsightTypePattern      InsightType = "pattern"      // Pattern recognition
	InsightTypeOptimization InsightType = "optimization" // Performance optimization
	InsightTypeSecurity     InsightType = "security"     // Security-related
	InsightTypeArchitecture InsightType = "architecture" // Architectural insight
	InsightTypeBugFix       InsightType = "bugfix"       // Bug fixing insight
	InsightTypeRefactoring  InsightType = "refactoring"  // Refactoring suggestion
	InsightTypeGeneral      InsightType = "general"      // General insight
)

// SecurityManagerConfig represents configuration for the security manager
type SecurityManagerConfig struct {
	ThreatDetection     *ThreatDetectionConfig     `json:"threatDetection"`
	RiskAssessment      *RiskAssessmentConfig      `json:"riskAssessment"`
	AccessControl       *AccessControlConfig       `json:"accessControl"`
	SecurityPolicies    []SecurityPolicy           `json:"securityPolicies"`
	AuditLogging        bool                       `json:"auditLogging"`
	EncryptionEnabled   bool                       `json:"encryptionEnabled"`
	SandboxEnabled      bool                       `json:"sandboxEnabled"`
	MaxRiskLevel        string                     `json:"maxRiskLevel"`        // low, medium, high, critical
	SecurityValidation  bool                       `json:"securityValidation"`
	ThreatIntelligence  bool                       `json:"threatIntelligence"`
}

// ThreatDetectionConfig represents threat detection configuration
type ThreatDetectionConfig struct {
	Enabled            bool     `json:"enabled"`
	ScanInterval       string   `json:"scanInterval"`
	ThreatSources      []string `json:"threatSources"`
	RealTimeMonitoring bool     `json:"realTimeMonitoring"`
	AlertThreshold     float64  `json:"alertThreshold"`
}

// RiskAssessmentConfig represents risk assessment configuration
type RiskAssessmentConfig struct {
	Enabled              bool    `json:"enabled"`
	RiskModel            string  `json:"riskModel"`
	MinRiskScore         float64 `json:"minRiskScore"`
	MaxRiskScore         float64 `json:"maxRiskScore"`
	RiskFactors          []string `json:"riskFactors"`
	AssessmentFrequency  string  `json:"assessmentFrequency"`
}

// AccessControlConfig represents access control configuration
type AccessControlConfig struct {
	Enabled            bool              `json:"enabled"`
	DefaultPermissions []string          `json:"defaultPermissions"`
	RoleBasedAccess    bool              `json:"roleBasedAccess"`
	UserPermissions    map[string][]string `json:"userPermissions"`
	SessionTimeout     time.Duration     `json:"sessionTimeout"`
}

// SecurityPolicy represents a security policy
type SecurityPolicy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Rules       []SecurityRule         `json:"rules"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityRule represents a security rule
type SecurityRule struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`      // validation, restriction, monitoring
	Condition string   `json:"condition"`
	Action    string   `json:"action"`    // allow, deny, warn, require_approval
	Severity  string   `json:"severity"`  // low, medium, high, critical
	Message   string   `json:"message"`
	Enabled   bool     `json:"enabled"`
}