package types

import (
	"time"
)

// TaskType represents the type of task being processed
type TaskType string

const (
	TaskTypeAnalysis    TaskType = "analysis"
	TaskTypeGeneration  TaskType = "generation"
	TaskTypeRefactor    TaskType = "refactor"
	TaskTypeExplain     TaskType = "explain"
	TaskTypeDebug       TaskType = "debug"
	TaskTypeTest        TaskType = "test"
	TaskTypeSearch      TaskType = "search"
	TaskTypeChat        TaskType = "chat"
	TaskTypeCustom      TaskType = "custom"
)

// TaskCategory represents the category of task complexity
type TaskCategory string

const (
	TaskCategorySimple   TaskCategory = "simple"   // Single-step tasks
	TaskCategoryComplex  TaskCategory = "complex"  // Multi-step tasks
	TaskCategoryCritical TaskCategory = "critical" // High-impact tasks
)

// Task represents a unified task definition
type Task struct {
	ID          string                 `json:"id"`
	Type        TaskType               `json:"type"`
	Category    TaskCategory           `json:"category"`
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
	Context     *TaskContext           `json:"context,omitempty"`
	Priority    int                    `json:"priority"` // 1-10, higher = more urgent
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Status      TaskStatus             `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// TaskContext represents the context in which a task is executed
type TaskContext struct {
	WorkingDirectory string            `json:"workingDirectory"`
	ProjectContext   *ProjectContext   `json:"projectContext,omitempty"`
	SessionContext   *SessionContext   `json:"sessionContext,omitempty"`
	UserPreferences  map[string]string `json:"userPreferences,omitempty"`
}

// ProjectContext represents project-specific context
type ProjectContext struct {
	RootPath     string   `json:"rootPath"`
	ProjectType  string   `json:"projectType"`  // go, nodejs, python, etc.
	Dependencies []string `json:"dependencies"`
	Structure    *ProjectStructure `json:"structure,omitempty"`
}

// ProjectStructure represents the structure of a project
type ProjectStructure struct {
	Directories []string `json:"directories"`
	Files       []string `json:"files"`
	MainFiles   []string `json:"mainFiles"`
	ConfigFiles []string `json:"configFiles"`
}

// SessionContext represents session-specific context
type SessionContext struct {
	SessionID     string            `json:"sessionId"`
	UserID        string            `json:"userId,omitempty"`
	StartTime     time.Time         `json:"startTime"`
	LastActivity  time.Time         `json:"lastActivity"`
	MessageCount  int               `json:"messageCount"`
	ToolUsage     map[string]int    `json:"toolUsage"`
	Preferences   map[string]string `json:"preferences,omitempty"`
}

// ResponseStatus represents the status of an agent response
type ResponseStatus string

const (
	ResponseStatusSuccess     ResponseStatus = "success"
	ResponseStatusPartial     ResponseStatus = "partial"
	ResponseStatusFailed      ResponseStatus = "failed"
	ResponseStatusNeedsInput  ResponseStatus = "needs_input"
	ResponseStatusProcessing  ResponseStatus = "processing"
)

// AgentResponse represents a unified response from the agent
type AgentResponse struct {
	ID          string                 `json:"id"`
	TaskID      string                 `json:"taskId"`
	Status      ResponseStatus         `json:"status"`
	Content     string                 `json:"content"`
	Data        interface{}            `json:"data,omitempty"`
	ToolResults []ToolResult           `json:"toolResults,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	ProcessingTime time.Duration       `json:"processingTime"`
	Error       *AgentError            `json:"error,omitempty"`
}

// AgentError represents an error in agent processing
type AgentError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Type    string `json:"type"` // validation, execution, system, etc.
}

// ExecutionStrategy represents how tasks should be executed
type ExecutionStrategy string

const (
	ExecutionStrategySequential ExecutionStrategy = "sequential"
	ExecutionStrategyParallel   ExecutionStrategy = "parallel"
	ExecutionStrategyOptimized  ExecutionStrategy = "optimized"
	ExecutionStrategyAdaptive   ExecutionStrategy = "adaptive"
)

// AgentMode represents the mode of operation for the agent
type AgentMode string

const (
	AgentModeInteractive AgentMode = "interactive"
	AgentModeBatch       AgentMode = "batch"
	AgentModeStreaming   AgentMode = "streaming"
	AgentModeDebug       AgentMode = "debug"
)

// PerformanceMetrics represents performance tracking data
type PerformanceMetrics struct {
	TasksProcessed   int           `json:"tasksProcessed"`
	AverageTime      time.Duration `json:"averageTime"`
	SuccessRate      float64       `json:"successRate"`
	ToolCallCount    int           `json:"toolCallCount"`
	MemoryUsage      int64         `json:"memoryUsage"`    // bytes
	CPUTime          time.Duration `json:"cpuTime"`
	LastMeasurement  time.Time     `json:"lastMeasurement"`
}

// ValidationResult represents the result of input validation
type ValidationResult struct {
	IsValid bool     `json:"isValid"`
	Errors  []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// QualityScore represents quality metrics for responses
type QualityScore struct {
	Accuracy    float64 `json:"accuracy"`    // 0.0 - 1.0
	Completeness float64 `json:"completeness"` // 0.0 - 1.0
	Relevance   float64 `json:"relevance"`   // 0.0 - 1.0
	Clarity     float64 `json:"clarity"`     // 0.0 - 1.0
	Overall     float64 `json:"overall"`     // computed score
}

// Feedback represents user feedback on agent responses
type Feedback struct {
	ResponseID  string     `json:"responseId"`
	Rating      int        `json:"rating"`      // 1-5 stars
	Comment     string     `json:"comment,omitempty"`
	Helpful     bool       `json:"helpful"`
	Timestamp   time.Time  `json:"timestamp"`
	Category    string     `json:"category,omitempty"` // accuracy, speed, clarity, etc.
}

// AgentStatus represents the current status of an agent
type AgentStatus struct {
	State          AgentState          `json:"state"`
	CurrentTask    *Task               `json:"currentTask,omitempty"`
	ActiveSessions []string            `json:"activeSessions"`
	LoadLevel      float64             `json:"loadLevel"`    // 0.0-1.0
	Memory         *MemoryStatus       `json:"memory"`
	LastActivity   string              `json:"lastActivity"`
	Uptime         string              `json:"uptime"`
	Errors         []AgentError        `json:"errors"`
}

// AgentState represents the state of an agent
type AgentState string

const (
	AgentStateIdle       AgentState = "idle"
	AgentStateThinking   AgentState = "thinking"
	AgentStatePlanning   AgentState = "planning"
	AgentStateExecuting  AgentState = "executing"
	AgentStateObserving  AgentState = "observing"
	AgentStateError      AgentState = "error"
	AgentStateStopped    AgentState = "stopped"
)

// MemoryStatus represents the memory status of an agent
type MemoryStatus struct {
	Used      int64   `json:"used"`      // bytes
	Available int64   `json:"available"` // bytes
	Usage     float64 `json:"usage"`     // percentage
}

// TaskAnalysis represents the result of task analysis
type TaskAnalysis struct {
	TaskID          string                 `json:"taskId"`
	Type            TaskType               `json:"type"`
	Category        TaskCategory           `json:"category"`
	Complexity      int                    `json:"complexity"`    // 1-10
	EstimatedTime   string                 `json:"estimatedTime"`
	RequiredTools   []string               `json:"requiredTools"`
	Dependencies    []string               `json:"dependencies"`
	Risks           []AnalysisRisk         `json:"risks"`
	Recommendations []string               `json:"recommendations"`
	Confidence      float64                `json:"confidence"`    // 0.0-1.0
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AnalysisRisk represents a risk identified during task analysis
type AnalysisRisk struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Probability float64 `json:"probability"` // 0.0-1.0
	Impact      string  `json:"impact"`      // low, medium, high, critical
	Mitigation  string  `json:"mitigation"`
}