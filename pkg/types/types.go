package types

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	Content string `json:"content"`
	Delta   string `json:"delta,omitempty"`
	Done    bool   `json:"done,omitempty"`
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

// DirectoryContextInfo - 目录上下文信息
type DirectoryContextInfo struct {
	Path         string     `json:"path"`          // 完整路径
	FileCount    int        `json:"file_count"`    // 文件数量
	DirCount     int        `json:"dir_count"`     // 目录数量
	TotalSize    int64      `json:"total_size"`    // 总大小
	LastModified time.Time  `json:"last_modified"` // 最后修改时间
	TopFiles     []FileInfo `json:"top_files"`     // 主要文件列表
	ProjectType  string     `json:"project_type"`  // 项目类型（Go、Python等）
	Description  string     `json:"description"`   // 目录简要描述
}

// FileInfo - 文件信息
type FileInfo struct {
	Name     string    `json:"name"`     // 文件名
	Path     string    `json:"path"`     // 相对路径
	Size     int64     `json:"size"`     // 文件大小
	Modified time.Time `json:"modified"` // 修改时间
	Type     string    `json:"type"`     // 文件类型
	IsDir    bool      `json:"is_dir"`   // 是否为目录
}

// ReactTaskContext - ReAct任务上下文
type ReactTaskContext struct {
	TaskID     string                 `json:"task_id"`     // 任务ID
	Goal       string                 `json:"goal"`        // 任务目标
	History    []ReactExecutionStep   `json:"history"`     // 执行历史
	Memory     map[string]interface{} `json:"memory"`      // 任务内存
	StartTime  time.Time              `json:"start_time"`  // 开始时间
	LastUpdate time.Time              `json:"last_update"` // 最后更新时间
	TokensUsed int                    `json:"tokens_used"` // 已使用token数
	Metadata   map[string]interface{} `json:"metadata"`    // 元数据
	// Directory context information
	WorkingDir    string                `json:"working_dir"`              // 对话发起时的工作目录
	DirectoryInfo *DirectoryContextInfo `json:"directory_info,omitempty"` // 目录信息
}

// ReactExecutionStep - ReAct执行步骤
type ReactExecutionStep struct {
	Number      int                `json:"number"`              // 步骤编号
	Thought     string             `json:"thought"`             // 思考内容
	Analysis    string             `json:"analysis"`            // 分析结果
	Action      string             `json:"action"`              // 执行动作
	ToolCall    *ReactToolCall     `json:"tool_call,omitempty"` // 工具调用
	Result      []*ReactToolResult `json:"result,omitempty"`    // 执行结果
	Observation string             `json:"observation"`         // 观察结果
	Confidence  float64            `json:"confidence"`          // 置信度 0.0-1.0
	Duration    time.Duration      `json:"duration"`            // 执行时长
	Timestamp   time.Time          `json:"timestamp"`           // 时间戳
	Error       string             `json:"error,omitempty"`     // 错误信息
	TokensUsed  int                `json:"tokens_used"`         // 本步骤使用的token数
}

// ReactTaskResult - ReAct任务执行结果
type ReactTaskResult struct {
	Success    bool                   `json:"success"`            // 是否成功
	Answer     string                 `json:"answer"`             // 答案内容
	Confidence float64                `json:"confidence"`         // 整体置信度
	Steps      []ReactExecutionStep   `json:"steps"`              // 执行步骤
	Duration   time.Duration          `json:"duration"`           // 总耗时
	TokensUsed int                    `json:"tokens_used"`        // 总token使用量
	Metadata   map[string]interface{} `json:"metadata,omitempty"` // 额外元数据
	Error      string                 `json:"error,omitempty"`    // 错误信息
}

// ReactToolCall - ReAct工具调用
type ReactToolCall struct {
	Name      string                 `json:"name"`      // 工具名称
	Arguments map[string]interface{} `json:"arguments"` // 调用参数
	CallID    string                 `json:"call_id"`   // 调用ID
}

// ReactToolResult - ReAct工具执行结果
type ReactToolResult struct {
	Success   bool                   `json:"success"`              // 是否成功
	Content   string                 `json:"content"`              // 结果内容
	Data      map[string]interface{} `json:"data,omitempty"`       // 结构化数据
	Error     string                 `json:"error,omitempty"`      // 错误信息
	Duration  time.Duration          `json:"duration"`             // 执行时长
	Metadata  map[string]interface{} `json:"metadata,omitempty"`   // 元数据
	ToolName  string                 `json:"tool_name,omitempty"`  // 工具名称
	ToolArgs  map[string]interface{} `json:"tool_args,omitempty"`  // 工具参数
	ToolCalls []*ReactToolCall       `json:"tool_calls,omitempty"` // 多个工具调用（并行执行时）
	CallID    string                 `json:"call_id,omitempty"`    // 调用ID
}

// ReactConfig - ReAct代理配置
type ReactConfig struct {
	MaxIterations       int           `json:"max_iterations"`       // 最大迭代次数，默认5
	ConfidenceThreshold float64       `json:"confidence_threshold"` // 置信度阈值，默认0.7
	TaskTimeout         time.Duration `json:"task_timeout"`         // 任务超时时间
	EnableAsync         bool          `json:"enable_async"`         // 启用异步执行
	ContextCompression  bool          `json:"context_compression"`  // 启用上下文压缩
	StreamingMode       bool          `json:"streaming_mode"`       // 流式模式
	LogLevel            string        `json:"log_level"`            // 日志级别
	Temperature         float64       `json:"temperature"`          // LLM温度参数
	MaxTokens           int           `json:"max_tokens"`           // 最大token数
}

// ReactConfig默认配置常量
const (
	ReactDefaultMaxIterations       = 5
	ReactDefaultConfidenceThreshold = 0.7
	ReactDefaultTaskTimeout         = 5 * time.Minute
	ReactDefaultMaxTokens           = 2000
	ReactDefaultTemperature         = 0.7
	ReactDefaultLogLevel            = "info"
	ReactDefaultMaxContextSize      = 10
	ReactDefaultCompressionRatio    = 0.6
	ReactDefaultMemorySlots         = 5
)

// NewReactConfig 创建默认的ReAct配置
func NewReactConfig() *ReactConfig {
	return &ReactConfig{
		MaxIterations:       ReactDefaultMaxIterations,
		ConfidenceThreshold: ReactDefaultConfidenceThreshold,
		TaskTimeout:         ReactDefaultTaskTimeout,
		EnableAsync:         true,
		ContextCompression:  true,
		StreamingMode:       true,
		LogLevel:            ReactDefaultLogLevel,
		Temperature:         ReactDefaultTemperature,
		MaxTokens:           ReactDefaultMaxTokens,
	}
}

// NewReactTaskContext 创建新的ReAct任务上下文
func NewReactTaskContext(taskID, goal string) *ReactTaskContext {
	workingDir, _ := getCurrentWorkingDir()
	directoryInfo := gatherDirectoryInfo(workingDir)

	return &ReactTaskContext{
		TaskID:        taskID,
		Goal:          goal,
		History:       make([]ReactExecutionStep, 0),
		Memory:        make(map[string]interface{}),
		StartTime:     time.Now(),
		LastUpdate:    time.Now(),
		TokensUsed:    0,
		Metadata:      make(map[string]interface{}),
		WorkingDir:    workingDir,
		DirectoryInfo: directoryInfo,
	}
}

// getCurrentWorkingDir 获取当前工作目录
func getCurrentWorkingDir() (string, error) {
	return os.Getwd()
}

// gatherDirectoryInfo 收集目录信息
func gatherDirectoryInfo(dirPath string) *DirectoryContextInfo {
	if dirPath == "" {
		return nil
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return &DirectoryContextInfo{
			Path:        dirPath,
			Description: "Unable to read directory",
		}
	}

	var fileCount, dirCount int
	var totalSize int64
	var lastModified time.Time
	var topFiles []FileInfo
	projectType := "Unknown"

	// 分析文件
	for _, entry := range entries {
		// 跳过隐藏文件
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if entry.IsDir() {
			dirCount++
		} else {
			fileCount++
			totalSize += info.Size()

			// 检测项目类型
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			switch ext {
			case ".go":
				if projectType == "Unknown" {
					projectType = "Go"
				}
			case ".py":
				if projectType == "Unknown" || projectType == "Go" {
					projectType = "Python"
				}
			case ".js", ".ts", ".jsx", ".tsx":
				if projectType == "Unknown" {
					projectType = "JavaScript/TypeScript"
				}
			case ".java":
				if projectType == "Unknown" {
					projectType = "Java"
				}
			case ".rs":
				if projectType == "Unknown" {
					projectType = "Rust"
				}
			}
		}

		// 更新最后修改时间
		if info.ModTime().After(lastModified) {
			lastModified = info.ModTime()
		}

		// 收集主要文件（限制数量）
		if len(topFiles) < 10 {
			fileType := "file"
			if entry.IsDir() {
				fileType = "directory"
			} else {
				ext := filepath.Ext(entry.Name())
				if ext != "" {
					fileType = ext[1:] // 去掉点号
				}
			}

			topFiles = append(topFiles, FileInfo{
				Name:     entry.Name(),
				Path:     entry.Name(),
				Size:     info.Size(),
				Modified: info.ModTime(),
				Type:     fileType,
				IsDir:    entry.IsDir(),
			})
		}
	}

	// 生成描述
	description := generateDirectoryDescription(dirPath, fileCount, dirCount, projectType)

	return &DirectoryContextInfo{
		Path:         dirPath,
		FileCount:    fileCount,
		DirCount:     dirCount,
		TotalSize:    totalSize,
		LastModified: lastModified,
		TopFiles:     topFiles,
		ProjectType:  projectType,
		Description:  description,
	}
}

// generateDirectoryDescription 生成目录描述
func generateDirectoryDescription(dirPath string, fileCount, dirCount int, projectType string) string {
	baseName := filepath.Base(dirPath)
	if baseName == "." || baseName == "/" {
		baseName = "current directory"
	}

	var desc strings.Builder
	desc.WriteString("Working in ")
	desc.WriteString(baseName)

	if projectType != "Unknown" {
		desc.WriteString(" (")
		desc.WriteString(projectType)
		desc.WriteString(" project)")
	}

	desc.WriteString(" containing ")
	if fileCount > 0 {
		desc.WriteString(formatCount(fileCount, "file"))
	}
	if dirCount > 0 {
		if fileCount > 0 {
			desc.WriteString(" and ")
		}
		desc.WriteString(formatCount(dirCount, "directory", "directories"))
	}

	return desc.String()
}

// formatCount 格式化计数文本
func formatCount(count int, singular string, plural ...string) string {
	pluralForm := singular + "s"
	if len(plural) > 0 {
		pluralForm = plural[0]
	}

	if count == 1 {
		return "1 " + singular
	}
	return fmt.Sprintf("%d %s", count, pluralForm)
}
