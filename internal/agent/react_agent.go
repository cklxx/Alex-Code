package agent

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"deep-coding-agent/internal/config"
	"deep-coding-agent/internal/llm"
	"deep-coding-agent/internal/prompts"
	"deep-coding-agent/internal/session"
	"deep-coding-agent/internal/tools/builtin"
	"deep-coding-agent/pkg/types"
)

// ReactAgent - 轻量化ReAct引擎
// Think-Act-Observe循环的智能代理实现
type ReactAgent struct {
	llm            llm.Client
	configManager  *config.Manager
	sessionManager *session.Manager
	tools          map[string]builtin.Tool
	config         *types.LightConfig
	llmConfig      *llm.Config
	promptBuilder  *LightPromptBuilder
	contextMgr     *LightContextManager
	codeExecutor   *CodeActExecutor
	currentSession *session.Session
	// 新增组件
	reactCore      *ReactCore
	thinkingEngine *ThinkingEngine
	toolExecutor   *ToolExecutor
	mu             sync.RWMutex
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
	ToolResults []types.LightToolResult `json:"toolResults"`
	SessionID   string                  `json:"sessionId"`
	Complete    bool                    `json:"complete"`
}

// StreamChunk - 兼容原有的流式响应
type StreamChunk struct {
	Type     string                 `json:"type"`
	Content  string                 `json:"content"`
	Complete bool                   `json:"complete,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
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
	builtinTools := builtin.GetAllBuiltinTools()
	for _, tool := range builtinTools {
		tools[tool.Name()] = tool
	}

	// 创建轻量化配置
	lightConfig := types.NewLightConfig()

	// 创建组件
	promptBuilder := NewLightPromptBuilder()
	contextMgr := NewLightContextManager()
	codeExecutor := NewCodeActExecutor()

	agent := &ReactAgent{
		llm:            llmClient,
		configManager:  configManager,
		sessionManager: sessionManager,
		tools:          tools,
		config:         lightConfig,
		llmConfig:      llmConfig,
		promptBuilder:  promptBuilder,
		contextMgr:     contextMgr,
		codeExecutor:   codeExecutor,
	}

	// 初始化组件
	agent.reactCore = NewReactCore(agent)
	agent.thinkingEngine = NewThinkingEngine(agent)
	agent.toolExecutor = NewToolExecutor(agent)

	return agent, nil
}

// StartSession - 开始会话
func (r *ReactAgent) StartSession(sessionID string) (*session.Session, error) {
	session, err := r.sessionManager.StartSession(sessionID)
	if err != nil {
		return nil, err
	}
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

	// 执行统一的ReAct循环（非流式）
	result, err := r.reactCore.SolveTask(ctx, userMessage, nil)
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

	// 转换结果为兼容格式
	toolResults := make([]types.LightToolResult, 0)
	for _, step := range result.Steps {
		if step.Result != nil {
			toolResults = append(toolResults, *step.Result)
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

	// 执行统一的ReAct循环（流式）
	result, err := r.reactCore.SolveTask(ctx, userMessage, callback)
	if err != nil {
		return fmt.Errorf("streaming task solving failed: %w", err)
	}

	// 添加assistant消息
	assistantMsg := &session.Message{
		Role:    "assistant",
		Content: result.Answer,
		Metadata: map[string]interface{}{
			"timestamp":  time.Now().Unix(),
			"streaming":  true,
			"confidence": result.Confidence,
		},
		Timestamp: time.Now(),
	}
	currentSession.AddMessage(assistantMsg)

	callback(StreamChunk{
		Type:     "complete",
		Content:  "Task completed",
		Complete: true,
	})

	return nil
}

func (r *ReactAgent) thinkWithConversation(ctx context.Context, messages []llm.Message, taskCtx *types.LightTaskContext) (*llm.ChatResponse, string, float64, int, error) {
	return r.thinkingEngine.thinkWithConversation(ctx, messages, taskCtx)
}

func (r *ReactAgent) canProvideDirectAnswer(thought string, confidence float64) bool {
	return r.thinkingEngine.canProvideDirectAnswer(thought, confidence)
}

func (r *ReactAgent) isTaskComplete(observation string, confidence float64) bool {
	return r.thinkingEngine.isTaskComplete(observation, confidence)
}

// 代理方法 - 委托给 ToolExecutor
func (r *ReactAgent) parseToolCalls(message *llm.Message) []*types.LightToolCall {
	return r.toolExecutor.parseToolCalls(message)
}

func (r *ReactAgent) executeParallelTools(ctx context.Context, toolCalls []*types.LightToolCall) *types.LightToolResult {
	return r.toolExecutor.executeParallelTools(ctx, toolCalls)
}

func (r *ReactAgent) executeParallelToolsStream(ctx context.Context, toolCalls []*types.LightToolCall, callback StreamCallback) *types.LightToolResult {
	return r.toolExecutor.executeParallelToolsStream(ctx, toolCalls, callback)
}

func (r *ReactAgent) parseCodeBlock(thought string) *CodeBlock {
	// 解析代码块格式：```language\ncode\n```
	re := regexp.MustCompile("```(\\w+)\\n([\\s\\S]*?)\\n```")
	matches := re.FindStringSubmatch(thought)

	if len(matches) >= 3 {
		return &CodeBlock{
			Language: matches[1],
			Code:     matches[2],
		}
	}

	return nil
}

func (r *ReactAgent) observe(result *types.LightToolResult, confidence float64) string {
	if result == nil {
		return "No action executed, continuing with reasoning."
	}

	// Try to use the observation prompt from prompts loader
	if r.promptBuilder.promptLoader != nil {
		toolResults := ""
		if result.Success {
			toolResults = fmt.Sprintf("SUCCESS: %s", result.Content)
		} else {
			toolResults = fmt.Sprintf("FAILED: %s", result.Error)
		}

		observationPrompt, err := r.promptBuilder.promptLoader.GetReActObservationPrompt("", toolResults)
		if err == nil && observationPrompt != "" {
			// Use the structured observation prompt
			return observationPrompt
		}
	}

	// Fallback to simple observation
	if result.Success {
		return fmt.Sprintf("Action execution successful: %s", result.Content)
	} else {
		return fmt.Sprintf("Action execution failed: %s", result.Error)
	}
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

func (pb *LightPromptBuilder) BuildTaskPrompt(task string, context *types.LightTaskContext) string {
	if pb.promptLoader == nil {
		// Fallback to basic template if prompt loader failed
		return fmt.Sprintf("Task: %s\n\nAnalyze and complete this task using available tools.", task)
	}

	// Use react_thinking prompt as the main template
	template, err := pb.promptLoader.GetReActThinkingPrompt()
	if err != nil {
		log.Printf("[ERROR] LightPromptBuilder: Failed to get ReAct thinking prompt, trying fallback: %v", err)

		// Try fallback thinking prompt
		fallbackTemplate, fallbackErr := pb.promptLoader.GetFallbackThinkingPrompt()
		if fallbackErr != nil {
			log.Printf("[ERROR] LightPromptBuilder: Failed to get fallback thinking prompt: %v", fallbackErr)
			// Final fallback to task description
			return fmt.Sprintf("Task: %s\n\nAnalyze and complete this task using available tools.", task)
		}
		template = fallbackTemplate
	}

	// Build context variables
	availableTools := "file_read, file_write, file_list, directory_create, bash, grep"
	compressedContext := fmt.Sprintf("Goal: %s\nSteps completed: %d", context.Goal, len(context.History))

	// The template contains the full ReAct instructions, so we prepend the specific task
	return fmt.Sprintf("Current Task: %s\n\nAvailable Tools: %s\nContext: %s\n\n%s",
		task, availableTools, compressedContext, template)
}

func NewLightContextManager() *LightContextManager {
	return &LightContextManager{
		maxContextSize:   types.DefaultMaxContextSize,
		compressionRatio: types.DefaultCompressionRatio,
		keyStepThreshold: 0.8,
	}
}

func (cm *LightContextManager) CompressContext(context *types.LightTaskContext) string {
	if len(context.History) <= cm.maxContextSize {
		return cm.formatFullContext(context)
	}

	var contextParts []string
	contextParts = append(contextParts, fmt.Sprintf("Goal: %s", context.Goal))

	// 保留最后几个步骤
	for i := max(0, len(context.History)-3); i < len(context.History); i++ {
		step := context.History[i]
		contextParts = append(contextParts,
			fmt.Sprintf("Step %d: %s -> %s", step.Number, step.Thought, step.Observation))
	}

	return strings.Join(contextParts, "\n")
}

func (cm *LightContextManager) formatFullContext(context *types.LightTaskContext) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Goal: %s", context.Goal))

	for _, step := range context.History {
		parts = append(parts,
			fmt.Sprintf("Step %d: %s -> %s", step.Number, step.Thought, step.Observation))
	}

	return strings.Join(parts, "\n")
}

// 辅助函数
func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

func generateCallID() string {
	return fmt.Sprintf("call_%d", time.Now().UnixNano())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
