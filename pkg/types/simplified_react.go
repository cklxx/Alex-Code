package types

import (
	"time"
)

// SimplifiedReActSession consolidates the overly complex ReAct session management
// This replaces the complex SessionStatus, TurnStatus hierarchies with simple states
type SimplifiedReActSession struct {
	ID        string                 `json:"id"`
	TaskID    string                 `json:"taskId"`
	Status    string                 `json:"status"` // "active", "completed", "failed"
	StartTime time.Time              `json:"startTime"`
	EndTime   *time.Time             `json:"endTime,omitempty"`
	Turns     []SimplifiedReActTurn  `json:"turns"`
	MaxTurns  int                    `json:"maxTurns"`
	Result    string                 `json:"result,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SimplifiedReActTurn consolidates the complex turn structure
type SimplifiedReActTurn struct {
	Number      int              `json:"number"`
	Thought     string           `json:"thought"`     // Simplified thinking content
	Action      SimplifiedAction `json:"action"`      // Simplified action plan
	Observation string           `json:"observation"` // Simplified observation
	Confidence  float64          `json:"confidence"`  // 0.0-1.0
	StartTime   time.Time        `json:"startTime"`
	Duration    time.Duration    `json:"duration"`
	Error       string           `json:"error,omitempty"`
}

// SimplifiedAction consolidates the complex ActionPlan and PlannedToolCall structures
type SimplifiedAction struct {
	Description string               `json:"description"`
	Tools       []SimplifiedToolCall `json:"tools"`
	Strategy    string               `json:"strategy"` // "sequential", "parallel"
	Confidence  float64              `json:"confidence"`
}

// SimplifiedToolCall consolidates the complex PlannedToolCall structure
type SimplifiedToolCall struct {
	ToolName  string                 `json:"toolName"`
	Arguments map[string]interface{} `json:"arguments"`
	Parallel  bool                   `json:"parallel,omitempty"`
}

// SimplifiedReActConfig consolidates the complex ReAct configuration
type SimplifiedReActConfig struct {
	MaxTurns            int           `json:"maxTurns"`
	ConfidenceThreshold float64       `json:"confidenceThreshold"`
	TimeoutPerTurn      time.Duration `json:"timeoutPerTurn"`
	Strategy            string        `json:"strategy"` // "standard", "optimized"
	EnableFallback      bool          `json:"enableFallback"`
	LogLevel            string        `json:"logLevel"`
}

// SimplifiedReActStrategy enum - simplified
type SimplifiedReActStrategy string

const (
	SimplifiedReActStrategyStandard  SimplifiedReActStrategy = "standard"
	SimplifiedReActStrategyOptimized SimplifiedReActStrategy = "optimized"
)

// Thinking result - simplified
type ThinkingResult struct {
	Content      string   `json:"content"`
	Analysis     string   `json:"analysis"`
	Reasoning    []string `json:"reasoning"`
	Confidence   float64  `json:"confidence"`
	Alternatives []string `json:"alternatives,omitempty"`
}

// SimplifiedObservationResult represents observation result - simplified
type SimplifiedObservationResult struct {
	Summary     string   `json:"summary"`
	Insights    []string `json:"insights"`
	Confidence  float64  `json:"confidence"`
	NextActions []string `json:"nextActions,omitempty"`
}
