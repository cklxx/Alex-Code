package types

import "time"

// UnifiedConfig consolidates all configuration into a single structure
type UnifiedConfig struct {
	AI     AIConfig     `json:"ai"`
	Agent  AgentConfig  `json:"agent"`
	Tools  ToolsConfig  `json:"tools"`
	Memory MemoryConfig `json:"memory"`
}

// AIConfig contains AI provider configuration
type AIConfig struct {
	Provider    string       `json:"provider"`
	MaxTokens   int          `json:"maxTokens"`
	Temperature float64      `json:"temperature"`
	OpenAI      OpenAIConfig `json:"openai"`
}

// OpenAIConfig contains OpenAI specific configuration
type OpenAIConfig struct {
	APIKey  string `json:"apiKey"`
	Model   string `json:"model"`
	BaseURL string `json:"baseUrl"`
}

// AgentConfig contains agent configuration
type AgentConfig struct {
	UseReActAgent       bool          `json:"useReActAgent"`
	StreamResponse      bool          `json:"streamResponse"`
	MaxTurns            int           `json:"maxTurns"`
	ConfidenceThreshold float64       `json:"confidenceThreshold"`
	TimeoutPerTurn      time.Duration `json:"timeoutPerTurn"`
	Strategy            string        `json:"strategy"`
	EnableFallback      bool          `json:"enableFallback"`
	LogLevel            string        `json:"logLevel"`
}

// ToolsConfig contains tools configuration
type ToolsConfig struct {
	AllowedTools    []string `json:"allowedTools"`
	MaxConcurrency  int      `json:"maxConcurrency"`
	Timeout         int      `json:"timeout"`
	RestrictedPaths []string `json:"restrictedPaths"`
	SecurityLevel   string   `json:"securityLevel"`
	LogLevel        string   `json:"logLevel"`
}

// MemoryConfig contains memory configuration
type MemoryConfig struct {
	MaxItems           int    `json:"maxItems"`
	RetentionDays      int    `json:"retentionDays"`
	StorageType        string `json:"storageType"`
	StoragePath        string `json:"storagePath"`
	CompressionEnabled bool   `json:"compressionEnabled"`
	BackupEnabled      bool   `json:"backupEnabled"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	Content string `json:"content"`
	Delta   string `json:"delta,omitempty"`
	Done    bool   `json:"done,omitempty"`
}

// AnalysisResult represents the result of code analysis
type AnalysisResult struct {
	FileCount   int            `json:"fileCount"`
	LinesOfCode int            `json:"linesOfCode"`
	Complexity  int            `json:"complexity"`
	Patterns    []Pattern      `json:"patterns"`
	Suggestions []string       `json:"suggestions"`
	Files       []FileAnalysis `json:"files,omitempty"`
}

// FileAnalysis represents analysis of a single file
type FileAnalysis struct {
	Path       string   `json:"path"`
	Size       int64    `json:"size"`
	Lines      int      `json:"lines"`
	Functions  int      `json:"functions"`
	Classes    int      `json:"classes"`
	Imports    []string `json:"imports"`
	Complexity int      `json:"complexity"`
	Language   string   `json:"language"`
}

// Pattern represents a detected code pattern
type Pattern struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Occurrences int    `json:"occurrences"`
}

// RefactorResult represents the result of refactoring
type RefactorResult struct {
	Summary string   `json:"summary"`
	Changes []Change `json:"changes"`
}

// Change represents a single refactoring change
type Change struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	LineNumber  *int   `json:"lineNumber,omitempty"`
	Before      string `json:"before,omitempty"`
	After       string `json:"after,omitempty"`
}

// GenerationResult represents the result of code generation
type GenerationResult struct {
	Code        string `json:"code"`
	Language    string `json:"language"`
	Explanation string `json:"explanation,omitempty"`
}

// AIRequest represents a request to AI provider
type AIRequest struct {
	Prompt      string  `json:"prompt"`
	Context     string  `json:"context,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"maxTokens,omitempty"`
}

// AIResponse represents a response from AI provider
type AIResponse struct {
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	Confidence float64    `json:"confidence,omitempty"`
	Usage      *AIUsage   `json:"usage,omitempty"`
}

// AIUsage represents token usage information
type AIUsage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

// TodoItem represents a single todo task
type TodoItem struct {
	ID          string     `json:"id"`
	Content     string     `json:"content"`
	Status      string     `json:"status"` // pending, in_progress, completed
	Order       int        `json:"order"`  // execution order (1, 2, 3...)
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

// Config represents unified application configuration
type Config struct {
	// Core Application Settings
	DefaultLanguage  string   `yaml:"defaultLanguage" json:"defaultLanguage" mapstructure:"defaultLanguage"`
	OutputFormat     string   `yaml:"outputFormat" json:"outputFormat" mapstructure:"outputFormat"`
	AnalysisDepth    int      `yaml:"analysisDepth" json:"analysisDepth" mapstructure:"analysisDepth"`
	BackupOnRefactor bool     `yaml:"backupOnRefactor" json:"backupOnRefactor" mapstructure:"backupOnRefactor"`
	ExcludePatterns  []string `yaml:"excludePatterns" json:"excludePatterns" mapstructure:"excludePatterns"`

	// API Configuration
	APIKey  string `yaml:"api_key" json:"api_key" mapstructure:"api_key"`
	BaseURL string `yaml:"base_url" json:"base_url" mapstructure:"base_url"`
	Model   string `yaml:"model" json:"model" mapstructure:"model"`

	// Tavily API Configuration
	TavilyAPIKey string `yaml:"tavily_api_key" json:"tavily_api_key" mapstructure:"tavily_api_key"`

	// Agent Configuration (previously AgentConfig)
	AllowedTools   []string `yaml:"allowedTools" json:"allowedTools" mapstructure:"allowedTools"`
	MaxTokens      int      `yaml:"maxTokens" json:"maxTokens" mapstructure:"maxTokens"`
	Temperature    float64  `yaml:"temperature" json:"temperature" mapstructure:"temperature"`
	StreamResponse bool     `yaml:"streamResponse" json:"streamResponse" mapstructure:"streamResponse"`
	SessionTimeout int      `yaml:"sessionTimeout" json:"sessionTimeout" mapstructure:"sessionTimeout"` // minutes

	// CLI Configuration (previously CLIConfig)
	Interactive bool   `yaml:"interactive" json:"interactive" mapstructure:"interactive"`
	SessionID   string `yaml:"sessionId" json:"sessionId" mapstructure:"sessionId"`
	ConfigFile  string `yaml:"configFile" json:"configFile" mapstructure:"configFile"`

	// Session Management
	SessionCleanupInterval int `yaml:"sessionCleanupInterval" json:"sessionCleanupInterval" mapstructure:"sessionCleanupInterval"` // hours
	MaxSessionAge          int `yaml:"maxSessionAge" json:"maxSessionAge" mapstructure:"maxSessionAge"`                            // days
	MaxMessagesPerSession  int `yaml:"maxMessagesPerSession" json:"maxMessagesPerSession" mapstructure:"maxMessagesPerSession"`

	// Security Settings
	EnableSandbox        bool     `yaml:"enableSandbox" json:"enableSandbox" mapstructure:"enableSandbox"`
	RestrictedTools      []string `yaml:"restrictedTools" json:"restrictedTools" mapstructure:"restrictedTools"`
	MaxConcurrentTools   int      `yaml:"maxConcurrentTools" json:"maxConcurrentTools" mapstructure:"maxConcurrentTools"`
	ToolExecutionTimeout int      `yaml:"toolExecutionTimeout" json:"toolExecutionTimeout" mapstructure:"toolExecutionTimeout"` // seconds

	// MCP Configuration
	MCPEnabled           bool     `yaml:"mcpEnabled" json:"mcpEnabled" mapstructure:"mcpEnabled"`
	MCPServers           []string `yaml:"mcpServers" json:"mcpServers" mapstructure:"mcpServers"`
	MCPConnectionTimeout int      `yaml:"mcpConnectionTimeout" json:"mcpConnectionTimeout" mapstructure:"mcpConnectionTimeout"` // seconds
	MCPMaxConnections    int      `yaml:"mcpMaxConnections" json:"mcpMaxConnections" mapstructure:"mcpMaxConnections"`

	// ReAct Agent Configuration (ReAct is the core execution mode)
	ReActMaxIterations   int  `yaml:"reactMaxIterations" json:"reactMaxIterations" mapstructure:"reactMaxIterations"`
	ReActThinkingEnabled bool `yaml:"reactThinkingEnabled" json:"reactThinkingEnabled" mapstructure:"reactThinkingEnabled"`

	// Todo Management
	Todos []TodoItem `yaml:"todos" json:"todos" mapstructure:"todos"`

	CustomSettings map[string]string `yaml:"customSettings" json:"customSettings" mapstructure:"customSettings"`
	LastUpdated    time.Time         `yaml:"lastUpdated" json:"lastUpdated" mapstructure:"lastUpdated"`
}

// AnalyzeOptions represents options for code analysis
type AnalyzeOptions struct {
	Depth      int
	Format     string
	UseAI      bool
	AIType     string
	Language   string
	Concurrent bool
}

// GenerateOptions represents options for code generation
type GenerateOptions struct {
	Language string
	Output   string
	UseAI    bool
	Style    string
	Template string
}

// RefactorOptions represents options for code refactoring
type RefactorOptions struct {
	Pattern        string
	Backup         bool
	UseAI          bool
	DryRun         bool
	PreserveFormat bool
}

// SupportedLanguages contains the list of supported programming languages
var SupportedLanguages = map[string]string{
	".go":    "go",
	".js":    "javascript",
	".ts":    "typescript",
	".jsx":   "jsx",
	".tsx":   "tsx",
	".py":    "python",
	".java":  "java",
	".cpp":   "cpp",
	".c":     "c",
	".cs":    "csharp",
	".php":   "php",
	".rb":    "ruby",
	".rs":    "rust",
	".kt":    "kotlin",
	".swift": "swift",
}

// ChangeTypes contains valid refactoring change types
var ChangeTypes = []string{
	"rename", "extract", "inline", "format",
	"optimize", "modernize", "security",
}

// AIAnalysisTypes contains valid AI analysis types
var AIAnalysisTypes = []string{
	"general", "performance", "security",
	"structure", "suggestions", "quality",
}

// FunctionCall represents a standard OpenAI-style function call
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCall represents a tool call with standard format
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // Always "function"
	Function FunctionCall `json:"function"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ID      string      `json:"id"`
	Success bool        `json:"success"`
	Content string      `json:"content,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ChatMessage represents a chat message in the conversation
type ChatMessage struct {
	Role       string     `json:"role"` // system, user, assistant, tool
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// ToolDefinition represents a tool's schema definition
type ToolDefinition struct {
	Type     string             `json:"type"` // Always "function"
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition represents a function's schema
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}
