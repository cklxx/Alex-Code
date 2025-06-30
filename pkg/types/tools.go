package types

import (
	"time"
)

// ToolCategory represents the category of a tool
type ToolCategory string

const (
	ToolCategoryFile     ToolCategory = "file"
	ToolCategoryCode     ToolCategory = "code"
	ToolCategoryBash     ToolCategory = "bash"
	ToolCategoryGit      ToolCategory = "git"
	ToolCategorySearch   ToolCategory = "search"
	ToolCategoryAnalysis ToolCategory = "analysis"
	ToolCategoryFormat   ToolCategory = "format"
	ToolCategoryTest     ToolCategory = "test"
	ToolCategoryDebug    ToolCategory = "debug"
	ToolCategoryMCP      ToolCategory = "mcp"
	ToolCategoryCustom   ToolCategory = "custom"
)

// ToolSchema represents the schema definition for a tool
type ToolSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    ToolCategory           `json:"category"`
	Parameters  map[string]interface{} `json:"parameters"`
	Required    []string               `json:"required"`
	Examples    []ToolExample          `json:"examples,omitempty"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author,omitempty"`
	Permissions []string               `json:"permissions"` // Required permissions
}

// ToolExample represents an example of tool usage
type ToolExample struct {
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
	Output      string                 `json:"output"`
}

// ExecutionPlan represents a plan for executing multiple tools
type ExecutionPlan struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Steps       []ExecutionStep   `json:"steps"`
	Strategy    ExecutionStrategy `json:"strategy"`
	MaxConcurrency int            `json:"maxConcurrency"`
	Timeout     time.Duration     `json:"timeout"`
	RetryPolicy *RetryPolicy      `json:"retryPolicy,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionStep represents a single step in an execution plan
type ExecutionStep struct {
	ID           string                 `json:"id"`
	ToolName     string                 `json:"toolName"`
	Arguments    map[string]interface{} `json:"arguments"`
	Dependencies []string               `json:"dependencies"` // Step IDs this depends on
	Optional     bool                   `json:"optional"`
	Condition    string                 `json:"condition,omitempty"` // Condition for execution
	Timeout      time.Duration          `json:"timeout"`
	RetryCount   int                    `json:"retryCount"`
	Status       StepStatus             `json:"status"`
	Result       *ToolResult            `json:"result,omitempty"`
	StartTime    *time.Time             `json:"startTime,omitempty"`
	EndTime      *time.Time             `json:"endTime,omitempty"`
}

// RetryPolicy represents retry configuration for tool execution
type RetryPolicy struct {
	MaxRetries    int           `json:"maxRetries"`
	InitialDelay  time.Duration `json:"initialDelay"`
	MaxDelay      time.Duration `json:"maxDelay"`
	BackoffFactor float64       `json:"backoffFactor"`
	RetryOnErrors []string      `json:"retryOnErrors"` // Error types to retry on
}

// ToolRegistry represents the registry of available tools
type ToolRegistry struct {
	Tools       map[string]*RegisteredTool `json:"tools"`
	Categories  map[ToolCategory][]string  `json:"categories"`
	LastUpdated time.Time                  `json:"lastUpdated"`
	Version     string                     `json:"version"`
}

// RegisteredTool represents a tool registered in the system
type RegisteredTool struct {
	Schema      *ToolSchema    `json:"schema"`
	Handler     interface{}    `json:"-"` // Function handler (not serialized)
	Config      *ToolConfig    `json:"config"`
	Metadata    *ToolMetadata  `json:"metadata"`
	Enabled     bool           `json:"enabled"`
	Restricted  bool           `json:"restricted"` // Requires special permissions
	RegisteredAt time.Time     `json:"registeredAt"`
}

// ToolConfig represents configuration for a specific tool
type ToolConfig struct {
	Timeout       time.Duration     `json:"timeout"`
	MaxRetries    int               `json:"maxRetries"`
	Environment   map[string]string `json:"environment"`
	WorkingDir    string            `json:"workingDir"`
	Sandbox       bool              `json:"sandbox"`
	LogLevel      string            `json:"logLevel"`
	RateLimit     *RateLimit        `json:"rateLimit,omitempty"`
	Validation    *ValidationConfig `json:"validation,omitempty"`
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	RequestsPerMinute int           `json:"requestsPerMinute"`
	BurstSize         int           `json:"burstSize"`
	WindowSize        time.Duration `json:"windowSize"`
}

// ValidationConfig represents validation configuration for tools
type ValidationConfig struct {
	ValidateInput  bool     `json:"validateInput"`
	ValidateOutput bool     `json:"validateOutput"`
	AllowedPaths   []string `json:"allowedPaths"`
	DeniedPaths    []string `json:"deniedPaths"`
	MaxFileSize    int64    `json:"maxFileSize"`
	MaxDuration    time.Duration `json:"maxDuration"`
}

// ToolMetadata represents metadata about a tool
type ToolMetadata struct {
	Usage        *ToolUsageStats `json:"usage"`
	Performance  *ToolPerformance `json:"performance"`
	LastUsed     *time.Time      `json:"lastUsed,omitempty"`
	ErrorRate    float64         `json:"errorRate"`
	Popularity   int             `json:"popularity"` // Usage ranking
	Reliability  float64         `json:"reliability"` // Success rate
}

// ToolUsageStats represents usage statistics for a tool
type ToolUsageStats struct {
	TotalCalls     int           `json:"totalCalls"`
	SuccessfulCalls int          `json:"successfulCalls"`
	FailedCalls    int           `json:"failedCalls"`
	AverageTime    time.Duration `json:"averageTime"`
	LastReset      time.Time     `json:"lastReset"`
}

// ToolPerformance represents performance metrics for a tool
type ToolPerformance struct {
	MinExecutionTime time.Duration `json:"minExecutionTime"`
	MaxExecutionTime time.Duration `json:"maxExecutionTime"`
	AvgExecutionTime time.Duration `json:"avgExecutionTime"`
	MemoryUsage      int64         `json:"memoryUsage"`
	CPUUsage         float64       `json:"cpuUsage"`
	LastBenchmark    time.Time     `json:"lastBenchmark"`
}

// ToolExecutionContext represents the context for tool execution
type ToolExecutionContext struct {
	SessionID      string            `json:"sessionId"`
	TaskID         string            `json:"taskId"`
	UserID         string            `json:"userId,omitempty"`
	WorkingDir     string            `json:"workingDir"`
	Environment    map[string]string `json:"environment"`
	Permissions    []string          `json:"permissions"`
	Timeout        time.Duration     `json:"timeout"`
	MaxMemory      int64             `json:"maxMemory"`
	Sandbox        bool              `json:"sandbox"`
	LogLevel       string            `json:"logLevel"`
	TraceID        string            `json:"traceId,omitempty"`
}

// ToolCallValidation represents validation result for a tool call
type ToolCallValidation struct {
	Valid        bool                   `json:"valid"`
	Errors       []ValidationError      `json:"errors,omitempty"`
	Warnings     []ValidationWarning    `json:"warnings,omitempty"`
	Suggestions  []string              `json:"suggestions,omitempty"`
	Risk         ToolRiskLevel         `json:"risk"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ToolRiskLevel represents the risk level of a tool call
type ToolRiskLevel string

const (
	ToolRiskLevelLow      ToolRiskLevel = "low"
	ToolRiskLevelMedium   ToolRiskLevel = "medium"
	ToolRiskLevelHigh     ToolRiskLevel = "high"
	ToolRiskLevelCritical ToolRiskLevel = "critical"
)

// ToolSelector represents configuration for tool selection
type ToolSelector struct {
	Strategy    ToolSelectionStrategy `json:"strategy"`
	Criteria    []SelectionCriteria   `json:"criteria"`
	Blacklist   []string              `json:"blacklist"`
	Whitelist   []string              `json:"whitelist"`
	Preferences map[string]float64    `json:"preferences"` // Tool name -> preference score
}

// ToolSelectionStrategy represents strategies for selecting tools
type ToolSelectionStrategy string

const (
	ToolSelectionStrategyBest       ToolSelectionStrategy = "best"       // Best tool for the job
	ToolSelectionStrategyFastest    ToolSelectionStrategy = "fastest"    // Fastest execution
	ToolSelectionStrategyReliable   ToolSelectionStrategy = "reliable"   // Most reliable
	ToolSelectionStrategyBalanced   ToolSelectionStrategy = "balanced"   // Balance of factors
	ToolSelectionStrategyUserPref   ToolSelectionStrategy = "user_pref"  // User preferences
)

// SelectionCriteria represents criteria for tool selection
type SelectionCriteria struct {
	Name      string  `json:"name"`
	Weight    float64 `json:"weight"`    // Weight in selection (0.0-1.0)
	Threshold float64 `json:"threshold"` // Minimum threshold
}

// ToolRecommendation represents a recommendation for tool usage
type ToolRecommendation struct {
	ToolName    string                 `json:"toolName"`
	Confidence  float64                `json:"confidence"`
	Rationale   string                 `json:"rationale"`
	Arguments   map[string]interface{} `json:"arguments"`
	Alternatives []string              `json:"alternatives"`
	Risk        ToolRiskLevel          `json:"risk"`
	Estimated   *EstimatedExecution    `json:"estimated,omitempty"`
}

// EstimatedExecution represents estimated execution metrics
type EstimatedExecution struct {
	Duration    time.Duration `json:"duration"`
	MemoryUsage int64        `json:"memoryUsage"`
	SuccessRate float64      `json:"successRate"`
	Cost        float64      `json:"cost,omitempty"` // If applicable
}

// ToolComposition represents a composition of multiple tools
type ToolComposition struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Tools       []CompositionTool   `json:"tools"`
	Flow        []FlowStep          `json:"flow"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CompositionTool represents a tool in a composition
type CompositionTool struct {
	Alias    string                 `json:"alias"`
	ToolName string                 `json:"toolName"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// FlowStep represents a step in a tool composition flow
type FlowStep struct {
	ID        string            `json:"id"`
	ToolAlias string            `json:"toolAlias"`
	Condition string            `json:"condition,omitempty"`
	OnSuccess string            `json:"onSuccess,omitempty"` // Next step ID
	OnFailure string            `json:"onFailure,omitempty"` // Next step ID on failure
	Parallel  []string          `json:"parallel,omitempty"`  // Parallel step IDs
}

// Configuration Types for Tools

// ToolSystemConfig represents configuration for the tool system
type ToolSystemConfig struct {
	MaxConcurrentExecutions int                 `json:"maxConcurrentExecutions"`
	DefaultTimeout         int64               `json:"defaultTimeout"` // milliseconds
	CacheConfig            *CacheConfig        `json:"cacheConfig"`
	SecurityConfig         *SecurityConfig     `json:"securityConfig"`
	MonitoringConfig       *MonitoringConfig   `json:"monitoringConfig"`
	MCPConfig              *MCPConfig          `json:"mcpConfig"`
}

// CacheConfig represents caching configuration
type CacheConfig struct {
	Enabled    bool  `json:"enabled"`
	MaxSize    int64 `json:"maxSize"`    // bytes
	DefaultTTL int64 `json:"defaultTtl"` // seconds
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	EnableSandbox       bool     `json:"enableSandbox"`
	MaxMemoryUsage      int64    `json:"maxMemoryUsage"`      // bytes
	MaxExecutionTime    int64    `json:"maxExecutionTime"`    // milliseconds
	AllowedTools        []string `json:"allowedTools"`
	RestrictedTools     []string `json:"restrictedTools"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled         bool   `json:"enabled"`
	MetricsInterval int64  `json:"metricsInterval"` // seconds
	LogLevel        string `json:"logLevel"`
}

// MCPConfig represents MCP configuration
type MCPConfig struct {
	Enabled           bool   `json:"enabled"`
	ConnectionTimeout int64  `json:"connectionTimeout"` // seconds
	MaxConnections    int    `json:"maxConnections"`
	AutoDiscovery     bool   `json:"autoDiscovery"`
}

// Additional Tool System Types

// ToolSystemMetrics represents metrics for the tool system
type ToolSystemMetrics struct {
	TotalExecutions    int                      `json:"totalExecutions"`
	SuccessfulExecutions int                    `json:"successfulExecutions"`
	FailedExecutions   int                      `json:"failedExecutions"`
	AverageExecutionTime time.Duration          `json:"averageExecutionTime"`
	ToolUsageStats     map[string]*ToolUsageStats `json:"toolUsageStats"`
	PerformanceStats   map[string]*ToolPerformance `json:"performanceStats"`
	ErrorStats         map[string]int           `json:"errorStats"`
	LastUpdated        time.Time                `json:"lastUpdated"`
}

// ToolExecutionRecord represents a record of tool execution
type ToolExecutionRecord struct {
	ID          string                 `json:"id"`
	ToolName    string                 `json:"toolName"`
	Arguments   map[string]interface{} `json:"arguments"`
	Result      *ToolResult            `json:"result,omitempty"`
	Success     bool                   `json:"success"`
	ExecutionTime time.Duration        `json:"executionTime"`
	Error       string                 `json:"error,omitempty"`
	Context     *ToolExecutionContext  `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	UserID      string                 `json:"userId,omitempty"`
	SessionID   string                 `json:"sessionId,omitempty"`
}

// ToolStatus represents the status of a tool
type ToolStatus string

const (
	ToolStatusActive      ToolStatus = "active"
	ToolStatusInactive    ToolStatus = "inactive"
	ToolStatusMaintenance ToolStatus = "maintenance"
	ToolStatusError       ToolStatus = "error"
	ToolStatusDeprecated  ToolStatus = "deprecated"
)

// ToolSearchCriteria represents criteria for searching tools
type ToolSearchCriteria struct {
	Keywords    []string       `json:"keywords,omitempty"`
	Categories  []ToolCategory `json:"categories,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	Status      []ToolStatus   `json:"status,omitempty"`
	MinRating   float64        `json:"minRating,omitempty"`
	MaxComplexity int          `json:"maxComplexity,omitempty"`
	Permissions []string       `json:"permissions,omitempty"`
	Authors     []string       `json:"authors,omitempty"`
}

// ToolRegistryConfig represents configuration for tool registry
type ToolRegistryConfig struct {
	AutoDiscovery       bool     `json:"autoDiscovery"`
	AllowDynamicLoading bool     `json:"allowDynamicLoading"`
	SecurityValidation  bool     `json:"securityValidation"`
	VersionControl      bool     `json:"versionControl"`
	BackupEnabled       bool     `json:"backupEnabled"`
	MaxToolsPerCategory int      `json:"maxToolsPerCategory"`
	AllowedSources      []string `json:"allowedSources"`
	RestrictedNames     []string `json:"restrictedNames"`
}

// ToolRegistryExport represents export of tool registry
type ToolRegistryExport struct {
	Version     string                    `json:"version"`
	Tools       map[string]*RegisteredTool `json:"tools"`
	Categories  map[ToolCategory][]string  `json:"categories"`
	Metadata    map[string]interface{}     `json:"metadata"`
	ExportedAt  time.Time                  `json:"exportedAt"`
	ExportedBy  string                     `json:"exportedBy,omitempty"`
}

// ToolExecution represents an execution of a tool
type ToolExecution struct {
	ID          string                 `json:"id"`
	ToolName    string                 `json:"toolName"`
	Arguments   map[string]interface{} `json:"arguments"`
	Status      ExecutionStatus        `json:"status"`
	Result      *ToolResult            `json:"result,omitempty"`
	Error       *ExecutionError        `json:"error,omitempty"`
	StartTime   time.Time              `json:"startTime"`
	EndTime     *time.Time             `json:"endTime,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Context     *ToolExecutionContext  `json:"context,omitempty"`
	Retries     int                    `json:"retries"`
	Priority    int                    `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionStatus represents the status of tool execution
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
	ExecutionStatusTimeout   ExecutionStatus = "timeout"
)

// ExecutionError represents an error during tool execution
type ExecutionError struct {
	Type        string    `json:"type"`
	Message     string    `json:"message"`
	Code        string    `json:"code,omitempty"`
	Details     string    `json:"details,omitempty"`
	Recoverable bool      `json:"recoverable"`
	Timestamp   time.Time `json:"timestamp"`
}

// Final Tool Interface Types

// ExecutionResult represents the result of executing multiple tools
type ExecutionResult struct {
	Success     bool                   `json:"success"`
	Results     []*ToolResult          `json:"results"`
	Errors      []ExecutionError       `json:"errors,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ToolExecutorConfig represents configuration for tool executor
type ToolExecutorConfig struct {
	MaxConcurrency    int           `json:"maxConcurrency"`
	DefaultTimeout    time.Duration `json:"defaultTimeout"`
	RetryPolicy       *RetryPolicy  `json:"retryPolicy,omitempty"`
	ErrorHandling     string        `json:"errorHandling"`     // fail_fast, continue, rollback
	LoggingEnabled    bool          `json:"loggingEnabled"`
	MetricsEnabled    bool          `json:"metricsEnabled"`
	ValidationEnabled bool          `json:"validationEnabled"`
	SandboxEnabled    bool          `json:"sandboxEnabled"`
}

// SafetyCheck represents a safety check for tool execution
type SafetyCheck struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`        // pre_execution, post_execution, continuous
	Severity    string    `json:"severity"`    // low, medium, high, critical
	Action      string    `json:"action"`      // warn, block, require_approval
	Pattern     string    `json:"pattern,omitempty"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
}

// DependencyValidation represents validation of tool dependencies
type DependencyValidation struct {
	Valid            bool     `json:"valid"`
	MissingDependencies []string `json:"missingDependencies,omitempty"`
	ConflictingVersions []string `json:"conflictingVersions,omitempty"`
	Warnings         []string `json:"warnings,omitempty"`
	Suggestions      []string `json:"suggestions,omitempty"`
}

// ValidationRules represents validation rules for tools
type ValidationRules struct {
	InputValidation    bool     `json:"inputValidation"`
	OutputValidation   bool     `json:"outputValidation"`
	SecurityValidation bool     `json:"securityValidation"`
	PerformanceValidation bool  `json:"performanceValidation"`
	RequiredFields     []string `json:"requiredFields"`
	ForbiddenPatterns  []string `json:"forbiddenPatterns"`
	MaxExecutionTime   time.Duration `json:"maxExecutionTime"`
	MaxMemoryUsage     int64    `json:"maxMemoryUsage"`
}

// CompositionResult represents the result of executing a tool composition
type CompositionResult struct {
	Success       bool                   `json:"success"`
	StepResults   []*StepResult          `json:"stepResults"`
	FinalResult   interface{}            `json:"finalResult,omitempty"`
	Duration      time.Duration          `json:"duration"`
	FailedStep    string                 `json:"failedStep,omitempty"`
	Error         *ExecutionError        `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// StepResult represents the result of a single step in composition
type StepResult struct {
	StepID      string        `json:"stepId"`
	ToolName    string        `json:"toolName"`
	Success     bool          `json:"success"`
	Result      *ToolResult   `json:"result,omitempty"`
	Error       string        `json:"error,omitempty"`
	Duration    time.Duration `json:"duration"`
	StartTime   time.Time     `json:"startTime"`
	EndTime     time.Time     `json:"endTime"`
}

// FlowResult represents the result of executing a flow
type FlowResult struct {
	Success     bool                   `json:"success"`
	Path        []string               `json:"path"`        // execution path taken
	Results     map[string]*StepResult `json:"results"`     // step ID -> result
	Duration    time.Duration          `json:"duration"`
	Error       *ExecutionError        `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// CompositionExport represents export of a tool composition
type CompositionExport struct {
	Composition *ToolComposition       `json:"composition"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author,omitempty"`
	Description string                 `json:"description,omitempty"`
	Dependencies []string              `json:"dependencies"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ExportedAt  time.Time              `json:"exportedAt"`
}

// MCP and System Types

// MCPServerConfig represents configuration for MCP server
type MCPServerConfig struct {
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Capabilities []string         `json:"capabilities"`
	Auth        *MCPAuth         `json:"auth,omitempty"`
	Timeout     time.Duration    `json:"timeout"`
	RetryPolicy *RetryPolicy     `json:"retryPolicy,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Enabled     bool             `json:"enabled"`
}

// MCPAuth represents authentication for MCP
type MCPAuth struct {
	Type   string `json:"type"`   // bearer, basic, api_key
	Token  string `json:"token,omitempty"`
	Header string `json:"header,omitempty"`
	Value  string `json:"value,omitempty"`
}

// MCPServerStatus represents status of an MCP server
type MCPServerStatus struct {
	Name         string    `json:"name"`
	Connected    bool      `json:"connected"`
	LastPing     time.Time `json:"lastPing"`
	ResponseTime time.Duration `json:"responseTime"`
	ErrorCount   int       `json:"errorCount"`
	Status       string    `json:"status"`    // online, offline, error, timeout
}

// MCPClientConfig represents configuration for MCP client
type MCPClientConfig struct {
	MaxConnections    int           `json:"maxConnections"`
	ConnectionTimeout time.Duration `json:"connectionTimeout"`
	RequestTimeout    time.Duration `json:"requestTimeout"`
	RetryPolicy       *RetryPolicy  `json:"retryPolicy,omitempty"`
	KeepAlive         bool          `json:"keepAlive"`
	CompressionEnabled bool         `json:"compressionEnabled"`
}

// ResourceUsage represents resource usage information
type ResourceUsage struct {
	CPU        float64 `json:"cpu"`        // percentage
	Memory     int64   `json:"memory"`     // bytes
	Disk       int64   `json:"disk"`       // bytes
	Network    int64   `json:"network"`    // bytes
	OpenFiles  int     `json:"openFiles"`
	Goroutines int     `json:"goroutines"`
	Timestamp  time.Time `json:"timestamp"`
}

// SystemResourceUsage represents system-wide resource usage
type SystemResourceUsage struct {
	TotalCPU     float64 `json:"totalCpu"`
	TotalMemory  int64   `json:"totalMemory"`
	AvailableMemory int64 `json:"availableMemory"`
	TotalDisk    int64   `json:"totalDisk"`
	AvailableDisk int64  `json:"availableDisk"`
	LoadAverage  []float64 `json:"loadAverage"`
	ProcessCount int     `json:"processCount"`
	Timestamp    time.Time `json:"timestamp"`
}

// ToolUsageStat represents usage statistics for a tool
type ToolUsageStat struct {
	ToolName      string        `json:"toolName"`
	UsageCount    int           `json:"usageCount"`
	SuccessRate   float64       `json:"successRate"`
	AverageTime   time.Duration `json:"averageTime"`
	ErrorCount    int           `json:"errorCount"`
	LastUsed      time.Time     `json:"lastUsed"`
	ResourceUsage *ResourceUsage `json:"resourceUsage,omitempty"`
}

