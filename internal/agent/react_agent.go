package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"alex/internal/config"
	contextmgr "alex/internal/context"
	"alex/internal/llm"
	"alex/internal/memory"
	"alex/internal/prompts"
	"alex/internal/session"
	"alex/internal/tools/builtin"
	"alex/pkg/types"
)

// ContextKey 用于在context中存储值，避免类型冲突
type ContextKey string

const (
	SessionIDKey ContextKey = "sessionID"
	MemoriesKey  ContextKey = "memories"
)

// ReactCoreInterface - ReAct核心接口
type ReactCoreInterface interface {
	SolveTask(ctx context.Context, task string, streamCallback StreamCallback) (*types.ReactTaskResult, error)
	GetContextStats(sess *session.Session) *contextmgr.ContextStats
	ForceContextSummarization(ctx context.Context, sess *session.Session) (*contextmgr.ContextProcessingResult, error)
	RestoreFullContext(sess *session.Session, backupID string) error
}

// ReactAgent - 轻量化ReAct引擎
// Think-Act-Observe循环的智能代理实现
type ReactAgent struct {
	llm            llm.Client
	configManager  *config.Manager
	sessionManager *session.Manager
	memoryManager  *memory.MemoryManager // 新增: Memory系统管理器
	tools          map[string]builtin.Tool
	config         *types.ReactConfig
	llmConfig      *llm.Config
	promptBuilder  *LightPromptBuilder
	contextMgr     *LightContextManager
	currentSession *session.Session
	// 核心组件
	reactCore    ReactCoreInterface
	toolExecutor *ToolExecutor
	mu           sync.RWMutex
}

// LightPromptBuilder - 轻量化prompt构建器
type LightPromptBuilder struct {
	promptLoader *prompts.PromptLoader
}

// LightContextManager - 轻量化上下文管理器
type LightContextManager struct {
	maxContextSize   int
	compressionRatio float64
	keyStepThreshold float64
}

// Response - 响应格式
type Response struct {
	Message     *session.Message        `json:"message"`
	ToolResults []types.ReactToolResult `json:"toolResults"`
	SessionID   string                  `json:"sessionId"`
	Complete    bool                    `json:"complete"`
}

// StreamChunk - 兼容原有的流式响应
type StreamChunk struct {
	Type     string                 `json:"type"`
	Content  string                 `json:"content"`
	Complete bool                   `json:"complete,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	// Token usage information
	TokensUsed       int `json:"tokens_used,omitempty"`       // Current operation token usage
	TotalTokensUsed  int `json:"total_tokens_used,omitempty"` // Cumulative token usage in session
	PromptTokens     int `json:"prompt_tokens,omitempty"`     // Tokens used for prompt
	CompletionTokens int `json:"completion_tokens,omitempty"` // Tokens used for completion
}

// StreamCallback - 流式回调函数
type StreamCallback func(StreamChunk)

// NewReactAgent - 创建新的轻量化ReactAgent
func NewReactAgent(configManager *config.Manager) (*ReactAgent, error) {

	// 设置LLM配置提供函数
	llm.SetConfigProvider(func() (*llm.Config, error) {
		return configManager.GetLLMConfig(), nil
	})

	// 获取LLM配置和客户端
	llmConfig := configManager.GetLLMConfig()
	llmClient, err := llm.GetLLMInstance(llm.BasicModel)
	if err != nil {
		log.Printf("[ERROR] ReactAgent: Failed to get LLM instance: %v", err)
		return nil, fmt.Errorf("failed to get LLM instance: %w", err)
	}

	// 创建session manager
	sessionManager, err := session.NewManager()
	if err != nil {
		log.Printf("[ERROR] ReactAgent: Failed to create session manager: %v", err)
		return nil, fmt.Errorf("failed to create session manager: %w", err)
	}

	// 初始化工具
	tools := make(map[string]builtin.Tool)
	builtinTools := builtin.GetAllBuiltinToolsWithConfig(configManager)
	for _, tool := range builtinTools {
		tools[tool.Name()] = tool
	}

	// 创建轻量化配置
	lightConfig := types.NewReactConfig()

	// 创建组件
	promptBuilder := NewLightPromptBuilder()
	contextMgr := NewLightContextManager()

	// 初始化Memory系统
	var memoryManager *memory.MemoryManager
	if llmClient != nil {
		if mm, err := memory.NewMemoryManager(llmClient); err == nil {
			memoryManager = mm
			log.Printf("[INFO] ReactAgent: Memory system initialized successfully")
		} else {
			log.Printf("[WARN] ReactAgent: Failed to initialize memory system: %v", err)
			// 不失败，继续运行
		}
	}

	agent := &ReactAgent{
		llm:            llmClient,
		configManager:  configManager,
		sessionManager: sessionManager,
		memoryManager:  memoryManager, // 可能为nil，需要处理
		tools:          tools,
		config:         lightConfig,
		llmConfig:      llmConfig,
		promptBuilder:  promptBuilder,
		contextMgr:     contextMgr,
	}

	// 初始化核心组件
	agent.reactCore = NewReactCore(agent)
	agent.toolExecutor = NewToolExecutor(agent)

	return agent, nil
}

// StartSession - 开始会话
func (r *ReactAgent) StartSession(sessionID string) (*session.Session, error) {
	session, err := r.sessionManager.StartSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 会话已经自动记录了工作目录（在session manager中）
	r.mu.Lock()
	r.currentSession = session
	r.mu.Unlock()
	return session, nil
}

// RestoreSession - 恢复会话
func (r *ReactAgent) RestoreSession(sessionID string) (*session.Session, error) {
	session, err := r.sessionManager.RestoreSession(sessionID)
	if err != nil {
		log.Printf("[ERROR] ReactAgent: Failed to restore session %s: %v", sessionID, err)
		return nil, err
	}
	r.mu.Lock()
	r.currentSession = session
	r.mu.Unlock()
	return session, nil
}

// ProcessMessage - 处理消息
func (r *ReactAgent) ProcessMessage(ctx context.Context, userMessage string, config *config.Config) (*Response, error) {
	r.mu.RLock()
	currentSession := r.currentSession
	r.mu.RUnlock()

	if currentSession == nil {
		return nil, fmt.Errorf("no active session")
	}

	// 添加用户消息到会话
	userMsg := &session.Message{
		Role:    "user",
		Content: userMessage,
		Metadata: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}
	currentSession.AddMessage(userMsg)

	// Memory召回: 在处理前获取相关记忆
	ctxWithMemory := r.enhanceContextWithMemory(ctx, currentSession.ID, userMessage)

	// Add session ID to context for caching
	ctxWithSession := context.WithValue(ctxWithMemory, SessionIDKey, currentSession.ID)

	// 检查并处理上下文压缩（集成memory系统）
	if err := r.manageContextWithMemory(ctxWithSession, currentSession); err != nil {
		log.Printf("[WARN] Context management with memory failed: %v", err)
	}

	// 执行统一的ReAct循环（非流式）
	result, err := r.reactCore.SolveTask(ctxWithSession, userMessage, nil)
	if err != nil {
		return nil, fmt.Errorf("task solving failed: %w", err)
	}

	// 创建assistant消息
	assistantMsg := &session.Message{
		Role:    "assistant",
		Content: result.Answer,
		Metadata: map[string]interface{}{
			"timestamp":   time.Now().Unix(),
			"confidence":  result.Confidence,
			"tokens_used": result.TokensUsed,
		},
		Timestamp: time.Now(),
	}
	currentSession.AddMessage(assistantMsg)

	// Memory创建: 异步创建记忆，不阻塞响应
	go r.createMemoryAsync(ctx, currentSession, userMsg, assistantMsg, result)

	// 转换结果为兼容格式
	toolResults := make([]types.ReactToolResult, 0)
	for _, step := range result.Steps {
		if step.Result != nil {
			for _, tr := range step.Result {
				if tr != nil {
					toolResults = append(toolResults, *tr)
				}
			}
		}
	}

	return &Response{
		Message:     assistantMsg,
		ToolResults: toolResults,
		SessionID:   currentSession.ID,
		Complete:    true,
	}, nil
}

// ProcessMessageStream - 流式处理消息
func (r *ReactAgent) ProcessMessageStream(ctx context.Context, userMessage string, config *config.Config, callback StreamCallback) error {
	r.mu.RLock()
	currentSession := r.currentSession
	r.mu.RUnlock()

	if currentSession == nil {
		return fmt.Errorf("no active session")
	}

	// 添加用户消息
	userMsg := &session.Message{
		Role:    "user",
		Content: userMessage,
		Metadata: map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"streaming": true,
		},
		Timestamp: time.Now(),
	}
	currentSession.AddMessage(userMsg)

	// Add session ID to context for caching
	ctxWithSession := context.WithValue(ctx, SessionIDKey, currentSession.ID)

	// 执行统一的ReAct循环（流式）
	result, err := r.reactCore.SolveTask(ctxWithSession, userMessage, callback)
	if err != nil {
		return fmt.Errorf("streaming task solving failed: %w", err)
	}

	// 添加assistant消息
	assistantMsg := &session.Message{
		Role:    "assistant",
		Content: result.Answer,
		Metadata: map[string]interface{}{
			"timestamp":   time.Now().Unix(),
			"streaming":   true,
			"confidence":  result.Confidence,
			"tokens_used": result.TokensUsed,
		},
		Timestamp: time.Now(),
	}
	currentSession.AddMessage(assistantMsg)

	callback(StreamChunk{
		Type:             "complete",
		Content:          "Task completed",
		Complete:         true,
		TotalTokensUsed:  result.TokensUsed,
		PromptTokens:     result.PromptTokens,
		CompletionTokens: result.CompletionTokens,
	})

	return nil
}

// 代理方法 - 委托给 ToolExecutor
func (r *ReactAgent) parseToolCalls(message *llm.Message) []*types.ReactToolCall {
	return r.toolExecutor.parseToolCalls(message)
}

func (r *ReactAgent) executeSerialToolsStream(ctx context.Context, toolCalls []*types.ReactToolCall, callback StreamCallback) []*types.ReactToolResult {
	return r.toolExecutor.executeSerialToolsStream(ctx, toolCalls, callback)
}

// 公共接口方法
func (r *ReactAgent) GetAvailableTools() []string {
	tools := make([]string, 0, len(r.tools))
	for name := range r.tools {
		tools = append(tools, name)
	}
	return tools
}

func (r *ReactAgent) GetSessionHistory() []*session.Message {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.currentSession == nil {
		return nil
	}
	return r.currentSession.Messages
}

// GetReactCore - 获取ReactCore实例
func (r *ReactAgent) GetReactCore() ReactCoreInterface {
	return r.reactCore
}

// GetSessionManager - 获取SessionManager实例
func (r *ReactAgent) GetSessionManager() *session.Manager {
	return r.sessionManager
}

// CodeBlock - 代码块结构
type CodeBlock struct {
	Language string
	Code     string
}

// 组件创建函数

func NewLightPromptBuilder() *LightPromptBuilder {
	promptLoader, err := prompts.NewPromptLoader()
	if err != nil {
		log.Printf("[ERROR] LightPromptBuilder: Failed to create prompt loader: %v", err)
		// Return a builder with nil loader - will cause graceful failures
		return &LightPromptBuilder{promptLoader: nil}
	}

	return &LightPromptBuilder{
		promptLoader: promptLoader,
	}
}

func NewLightContextManager() *LightContextManager {
	return &LightContextManager{
		maxContextSize:   types.ReactDefaultMaxContextSize,
		compressionRatio: types.ReactDefaultCompressionRatio,
		keyStepThreshold: 0.8,
	}
}

func (cm *LightContextManager) CompressContext(context *types.ReactTaskContext) string {
	if len(context.History) <= cm.maxContextSize {
		return cm.formatFullContext(context)
	}

	var contextParts []string

	// 添加目录上下文信息
	contextParts = append(contextParts, cm.formatDirectoryContext(context))
	contextParts = append(contextParts, fmt.Sprintf("Goal: %s", context.Goal))

	// 保留最后几个步骤
	for i := max(0, len(context.History)-3); i < len(context.History); i++ {
		step := context.History[i]
		contextParts = append(contextParts,
			fmt.Sprintf("Step %d: %s -> %s", step.Number, step.Thought, step.Observation))
	}

	return strings.Join(contextParts, "\n")
}

func (cm *LightContextManager) formatFullContext(context *types.ReactTaskContext) string {
	var parts []string

	// 添加目录上下文信息
	parts = append(parts, cm.formatDirectoryContext(context))
	parts = append(parts, fmt.Sprintf("Goal: %s", context.Goal))

	for _, step := range context.History {
		parts = append(parts,
			fmt.Sprintf("Step %d: %s -> %s", step.Number, step.Thought, step.Observation))
	}

	return strings.Join(parts, "\n")
}

// formatDirectoryContext 格式化目录上下文信息
func (cm *LightContextManager) formatDirectoryContext(context *types.ReactTaskContext) string {
	var contextLines []string

	// 当前时间
	contextLines = append(contextLines, fmt.Sprintf("Current Time: %s", time.Now().Format(time.RFC3339)))

	// 工作目录
	if context.WorkingDir != "" {
		contextLines = append(contextLines, fmt.Sprintf("Working Directory: %s", context.WorkingDir))
	}

	// 目录信息
	if context.DirectoryInfo != nil {
		info := context.DirectoryInfo
		contextLines = append(contextLines, fmt.Sprintf("Directory Context: %s", info.Description))

		if len(info.TopFiles) > 0 {
			contextLines = append(contextLines, "Key Files:")
			for _, file := range info.TopFiles[:min(5, len(info.TopFiles))] {
				if file.IsDir {
					contextLines = append(contextLines, fmt.Sprintf("  📁 %s/", file.Name))
				} else {
					contextLines = append(contextLines, fmt.Sprintf("  📄 %s (%s)", file.Name, file.Type))
				}
			}
		}
	}

	return strings.Join(contextLines, "\n")
}

// Memory集成相关方法

// enhanceContextWithMemory 从memory召回相关信息并增强context
func (r *ReactAgent) enhanceContextWithMemory(ctx context.Context, sessionID, userMessage string) context.Context {
	if r.memoryManager == nil {
		return ctx
	}

	// 构建memory查询
	query := &memory.MemoryQuery{
		SessionID: sessionID,
		Content:   userMessage,
		Categories: []memory.MemoryCategory{
			memory.CodeContext,
			memory.TaskHistory,
			memory.Solutions,
			memory.ErrorPatterns,
		},
		MinImportance: 0.5,
		Limit:         5,
		SortBy:        "importance",
	}

	// 快速召回memories（带超时保护）
	recallResult := r.safeMemoryRecall(query, 50*time.Millisecond)

	// 将memories注入context
	return context.WithValue(ctx, MemoriesKey, recallResult)
}

// safeMemoryRecall 安全的memory召回，带超时保护
func (r *ReactAgent) safeMemoryRecall(query *memory.MemoryQuery, timeout time.Duration) *memory.RecallResult {
	resultChan := make(chan *memory.RecallResult, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[ERROR] Memory recall panic: %v", r)
				resultChan <- &memory.RecallResult{Items: []*memory.MemoryItem{}}
			}
		}()

		result := r.memoryManager.Recall(query)
		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		return result
	case <-time.After(timeout):
		log.Printf("[WARN] Memory recall timeout after %v", timeout)
		return &memory.RecallResult{Items: []*memory.MemoryItem{}}
	}
}

// manageContextWithMemory 管理上下文压缩并集成memory系统
func (r *ReactAgent) manageContextWithMemory(ctx context.Context, sess *session.Session) error {
	if r.memoryManager == nil {
		return nil
	}

	// 估算token使用量
	messages := sess.GetMessages()
	estimatedTokens := r.estimateTokenUsage(messages)

	// 假设模型上下文限制为100K tokens（后续可动态获取）
	maxTokens := 100000
	if float64(estimatedTokens)/float64(maxTokens) > 0.8 {
		// 需要压缩时，使用memory系统
		result, err := r.memoryManager.ProcessContextCompression(ctx, sess, maxTokens)
		if err != nil {
			return fmt.Errorf("memory-based context compression failed: %w", err)
		}

		if result.CompressedSummary != "" {
			log.Printf("[INFO] Memory compression: %d->%d messages, saved %d tokens",
				result.OriginalCount, result.CompressedCount, result.TokensSaved)
		}
	}

	return nil
}

// estimateTokenUsage 估算token使用量
func (r *ReactAgent) estimateTokenUsage(messages []*session.Message) int {
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Content) + 50 // 50字符开销
	}
	return totalChars / 3 // 粗略估算：3字符=1token
}

// createMemoryAsync 异步创建记忆
func (r *ReactAgent) createMemoryAsync(ctx context.Context, sess *session.Session, userMsg, assistantMsg *session.Message, result *types.ReactTaskResult) {
	if r.memoryManager == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR] Memory creation panic: %v", r)
		}
	}()

	// 为用户消息创建memory
	if userMemories, err := r.memoryManager.CreateMemoryFromMessage(ctx, sess.ID, userMsg, len(sess.Messages)); err == nil {
		if len(userMemories) > 0 {
			log.Printf("[DEBUG] Created %d memories from user message", len(userMemories))
		}
	} else {
		log.Printf("[WARN] Failed to create memory from user message: %v", err)
	}

	// 为assistant响应创建memory
	if assistantMemories, err := r.memoryManager.CreateMemoryFromMessage(ctx, sess.ID, assistantMsg, len(sess.Messages)); err == nil {
		if len(assistantMemories) > 0 {
			log.Printf("[DEBUG] Created %d memories from assistant message", len(assistantMemories))
		}
	} else {
		log.Printf("[WARN] Failed to create memory from assistant message: %v", err)
	}

	// 创建任务执行相关的memory
	r.createTaskExecutionMemory(ctx, sess.ID, result)

	// 执行定期维护
	if err := r.memoryManager.AutomaticMemoryMaintenance(sess.ID); err != nil {
		log.Printf("[WARN] Memory maintenance failed: %v", err)
	}
}

// createTaskExecutionMemory 创建任务执行相关的memory
func (r *ReactAgent) createTaskExecutionMemory(_ context.Context, sessionID string, result *types.ReactTaskResult) {
	if result == nil || len(result.Steps) == 0 {
		return
	}

	// 提取工具使用模式
	var toolNames []string
	for _, step := range result.Steps {
		if step.ToolCall != nil {
			for _, toolCall := range step.ToolCall {
				toolNames = append(toolNames, toolCall.Name)
			}
		}
	}

	if len(toolNames) > 0 {
		memory := &memory.MemoryItem{
			ID:         fmt.Sprintf("tool_pattern_%s_%d", sessionID, time.Now().UnixNano()),
			SessionID:  sessionID,
			Category:   memory.TaskHistory,
			Content:    fmt.Sprintf("Tool execution pattern: %s", strings.Join(toolNames, " -> ")),
			Importance: 0.6,
			Tags:       append([]string{"tool_pattern", "execution_flow"}, toolNames...),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			LastAccess: time.Now(),
			Metadata: map[string]interface{}{
				"tool_count":   len(toolNames),
				"success_rate": result.Confidence,
				"tokens_used":  result.TokensUsed,
			},
		}

		if err := r.memoryManager.Store(memory); err != nil {
			log.Printf("[WARN] Failed to store tool pattern memory: %v", err)
		}
	}

	// 如果任务失败或有错误，创建错误模式记忆
	if result.Confidence < 0.5 {
		memory := &memory.MemoryItem{
			ID:         fmt.Sprintf("error_pattern_%s_%d", sessionID, time.Now().UnixNano()),
			SessionID:  sessionID,
			Category:   memory.ErrorPatterns,
			Content:    fmt.Sprintf("Task execution had low confidence (%.2f), tools used: %s", result.Confidence, strings.Join(toolNames, ", ")),
			Importance: 0.8,
			Tags:       []string{"error", "low_confidence", "debugging"},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			LastAccess: time.Now(),
			Metadata: map[string]interface{}{
				"confidence":   result.Confidence,
				"failed_tools": toolNames,
			},
		}

		if err := r.memoryManager.Store(memory); err != nil {
			log.Printf("[WARN] Failed to store error pattern memory: %v", err)
		}
	}
}

// 辅助函数 - generateTaskID moved to utils.go

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
