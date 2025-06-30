package types

import (
	"time"
)

// 轻量化ReAct agent类型定义
// 避免与现有types包中的类型冲突，使用Light前缀

// LightTaskContext - 轻量化任务上下文
type LightTaskContext struct {
	TaskID     string                 `json:"task_id"`     // 任务ID
	Goal       string                 `json:"goal"`        // 任务目标
	History    []LightExecutionStep   `json:"history"`     // 执行历史
	Memory     map[string]interface{} `json:"memory"`      // 任务内存
	StartTime  time.Time              `json:"start_time"`  // 开始时间
	LastUpdate time.Time              `json:"last_update"` // 最后更新时间
	TokensUsed int                    `json:"tokens_used"` // 已使用token数
	Metadata   map[string]interface{} `json:"metadata"`    // 元数据
}

// LightExecutionStep - 简化的执行步骤
type LightExecutionStep struct {
	Number      int              `json:"number"`              // 步骤编号
	Thought     string           `json:"thought"`             // 思考内容
	Analysis    string           `json:"analysis"`            // 分析结果
	Action      string           `json:"action"`              // 执行动作
	ToolCall    *LightToolCall   `json:"tool_call,omitempty"` // 工具调用
	Result      *LightToolResult `json:"result,omitempty"`    // 执行结果
	Observation string           `json:"observation"`         // 观察结果
	Confidence  float64          `json:"confidence"`          // 置信度 0.0-1.0
	Duration    time.Duration    `json:"duration"`            // 执行时长
	Timestamp   time.Time        `json:"timestamp"`           // 时间戳
	Error       string           `json:"error,omitempty"`     // 错误信息
	TokensUsed  int              `json:"tokens_used"`         // 本步骤使用的token数
}

// LightTaskResult - 任务执行结果
type LightTaskResult struct {
	Success    bool                   `json:"success"`            // 是否成功
	Answer     string                 `json:"answer"`             // 答案内容
	Confidence float64                `json:"confidence"`         // 整体置信度
	Steps      []LightExecutionStep   `json:"steps"`              // 执行步骤
	Duration   time.Duration          `json:"duration"`           // 总耗时
	TokensUsed int                    `json:"tokens_used"`        // 总token使用量
	Metadata   map[string]interface{} `json:"metadata,omitempty"` // 额外元数据
	Error      string                 `json:"error,omitempty"`    // 错误信息
}

// LightToolCall - 轻量级工具调用
type LightToolCall struct {
	Name      string                 `json:"name"`      // 工具名称
	Arguments map[string]interface{} `json:"arguments"` // 调用参数
	CallID    string                 `json:"call_id"`   // 调用ID
}

// LightToolResult - 轻量级工具执行结果
type LightToolResult struct {
	Success   bool                   `json:"success"`              // 是否成功
	Content   string                 `json:"content"`              // 结果内容
	Data      map[string]interface{} `json:"data,omitempty"`       // 结构化数据
	Error     string                 `json:"error,omitempty"`      // 错误信息
	Duration  time.Duration          `json:"duration"`             // 执行时长
	Metadata  map[string]interface{} `json:"metadata,omitempty"`   // 元数据
	ToolName  string                 `json:"tool_name,omitempty"`  // 工具名称
	ToolArgs  map[string]interface{} `json:"tool_args,omitempty"`  // 工具参数
	ToolCalls []*LightToolCall       `json:"tool_calls,omitempty"` // 多个工具调用（并行执行时）
}

// LightTaskType - 任务类型枚举
type LightTaskType string

const (
	LightTaskTypeAnalysis     LightTaskType = "analysis"     // 分析任务
	LightTaskTypeCreation     LightTaskType = "creation"     // 创建任务
	LightTaskTypeModification LightTaskType = "modification" // 修改任务
	LightTaskTypeSearch       LightTaskType = "search"       // 搜索任务
	LightTaskTypeDebugging    LightTaskType = "debugging"    // 调试任务
	LightTaskTypeRefactoring  LightTaskType = "refactoring"  // 重构任务
	LightTaskTypeGeneral      LightTaskType = "general"      // 通用任务
)

// LightConfig - 简化配置
type LightConfig struct {
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

// 轻量化组件类型定义

// LightToolStats - 工具统计
type LightToolStats struct {
	ToolName     string         `json:"tool_name"`     // 工具名称
	CallCount    int64          `json:"call_count"`    // 调用次数
	SuccessCount int64          `json:"success_count"` // 成功次数
	FailureCount int64          `json:"failure_count"` // 失败次数
	TotalTime    time.Duration  `json:"total_time"`    // 总耗时
	AverageTime  time.Duration  `json:"average_time"`  // 平均耗时
	LastUsed     time.Time      `json:"last_used"`     // 最后使用时间
	ErrorTypes   map[string]int `json:"error_types"`   // 错误类型统计
}

// LightCacheEntry - 缓存条目
type LightCacheEntry struct {
	Key       string           `json:"key"`        // 缓存键
	Value     *LightToolResult `json:"value"`      // 缓存值
	CreatedAt time.Time        `json:"created_at"` // 创建时间
	TTL       time.Duration    `json:"ttl"`        // 生存时间
	HitCount  int              `json:"hit_count"`  // 命中次数
}

// CodeActResult - 代码执行结果
type CodeActResult struct {
	Success       bool          `json:"success"`        // 是否成功
	Output        string        `json:"output"`         // 标准输出
	Error         string        `json:"error"`          // 错误输出
	ExitCode      int           `json:"exit_code"`      // 退出码
	ExecutionTime time.Duration `json:"execution_time"` // 执行时间
	Language      string        `json:"language"`       // 编程语言
	Code          string        `json:"code"`           // 执行的代码
}

// 默认配置常量
const (
	DefaultMaxIterations       = 5
	DefaultConfidenceThreshold = 0.7
	DefaultTaskTimeout         = 5 * time.Minute
	DefaultMaxTokens           = 2000
	DefaultTemperature         = 0.7
	DefaultLogLevel            = "info"
	DefaultMaxContextSize      = 10
	DefaultCompressionRatio    = 0.6
	DefaultMemorySlots         = 5
)

// NewLightConfig 创建默认的轻量化配置
func NewLightConfig() *LightConfig {
	return &LightConfig{
		MaxIterations:       DefaultMaxIterations,
		ConfidenceThreshold: DefaultConfidenceThreshold,
		TaskTimeout:         DefaultTaskTimeout,
		EnableAsync:         true,
		ContextCompression:  true,
		StreamingMode:       true,
		LogLevel:            DefaultLogLevel,
		Temperature:         DefaultTemperature,
		MaxTokens:           DefaultMaxTokens,
	}
}

// NewLightTaskContext 创建新的轻量化任务上下文
func NewLightTaskContext(taskID, goal string) *LightTaskContext {
	return &LightTaskContext{
		TaskID:     taskID,
		Goal:       goal,
		History:    make([]LightExecutionStep, 0),
		Memory:     make(map[string]interface{}),
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
		TokensUsed: 0,
		Metadata:   make(map[string]interface{}),
	}
}
