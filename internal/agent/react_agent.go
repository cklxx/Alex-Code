package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"alex/internal/config"
	"alex/internal/llm"
	"alex/internal/prompts"
	"alex/internal/session"
	"alex/internal/tools/builtin"
	"alex/internal/tools/mcp"
	"alex/pkg/types"
)

// ContextKey 用于在context中存储值，避免类型冲突
type ContextKey string

const (
	SessionIDKey ContextKey = "sessionID"
)

// GeneratedMemories - 空记忆结构体 (Memory module removed)
type GeneratedMemories struct {
	SessionID string    `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
}

// ReactCoreInterface - ReAct核心接口
type ReactCoreInterface interface {
	SolveTask(ctx context.Context, task string, streamCallback StreamCallback) (*types.ReactTaskResult, error)
}

// ReactAgent - 简化的ReAct引擎
type ReactAgent struct {
	// 核心组件
	llm            llm.Client
	configManager  *config.Manager
	sessionManager *session.Manager
	tools          map[string]builtin.Tool
	config         *types.ReactConfig
	llmConfig      *llm.Config
	currentSession *session.Session

	// 核心组件
	reactCore     ReactCoreInterface
	toolExecutor  *ToolExecutor
	promptBuilder *LightPromptBuilder

	// 简单的同步控制
	mu sync.RWMutex
}

// Response - 响应格式
type Response struct {
	Message     *session.Message        `json:"message"`
	ToolResults []types.ReactToolResult `json:"toolResults"`
	SessionID   string                  `json:"sessionId"`
	Complete    bool                    `json:"complete"`
}

// StreamChunk - 流式响应
type StreamChunk struct {
	Type             string                 `json:"type"`
	Content          string                 `json:"content"`
	Complete         bool                   `json:"complete,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	TokensUsed       int                    `json:"tokens_used,omitempty"`
	TotalTokensUsed  int                    `json:"total_tokens_used,omitempty"`
	PromptTokens     int                    `json:"prompt_tokens,omitempty"`
	CompletionTokens int                    `json:"completion_tokens,omitempty"`
}

// StreamCallback - 流式回调函数
type StreamCallback func(StreamChunk)

// LightPromptBuilder - 轻量化prompt构建器
type LightPromptBuilder struct {
	promptLoader *prompts.PromptLoader
}

// NewReactAgent - 创建简化的ReactAgent
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

	// 集成MCP工具
	allTools := integrateWithMCPTools(configManager, builtinTools)

	for _, tool := range allTools {
		tools[tool.Name()] = tool
	}

	agent := &ReactAgent{
		llm:            llmClient,
		configManager:  configManager,
		sessionManager: sessionManager,
		tools:          tools,
		config:         types.NewReactConfig(),
		llmConfig:      llmConfig,

		promptBuilder: NewLightPromptBuilder(),
	}

	// 初始化核心组件
	agent.reactCore = NewReactCore(agent)
	agent.toolExecutor = NewToolExecutor(agent)

	// Memory tools removed

	return agent, nil
}

// ========== 会话管理 ==========

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

// ProcessMessage - 处理消息（简化版）
func (r *ReactAgent) ProcessMessage(ctx context.Context, userMessage string, config *config.Config) (*Response, error) {
	r.mu.RLock()
	currentSession := r.currentSession
	r.mu.RUnlock()

	// If no active session, create one automatically
	if currentSession == nil {
		log.Printf("[DEBUG] No active session found, creating new session automatically")
		sessionID := fmt.Sprintf("auto_%d", time.Now().UnixNano())
		newSession, err := r.StartSession(sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to create session automatically: %w", err)
		}
		currentSession = newSession
		log.Printf("[DEBUG] Auto-created session: %s", currentSession.ID)
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

	// 将会话ID注入context - 使用类型安全的key
	ctxWithSession := context.WithValue(ctx, SessionIDKey, currentSession.ID)
	log.Printf("[DEBUG] 🔧 Context set with session ID: %s", currentSession.ID)

	// 执行ReAct循环
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

	// Memory generation removed

	// 转换结果
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

// ProcessMessageStream - 流式处理消息（简化版）
func (r *ReactAgent) ProcessMessageStream(ctx context.Context, userMessage string, config *config.Config, callback StreamCallback) error {
	log.Printf("[DEBUG] ====== ProcessMessageStream called with message: %s", userMessage)
	
	r.mu.RLock()
	currentSession := r.currentSession
	r.mu.RUnlock()

	// If no active session, create one automatically
	if currentSession == nil {
		log.Printf("[DEBUG] No active session found, creating new session automatically")
		sessionID := fmt.Sprintf("auto_%d", time.Now().UnixNano())
		newSession, err := r.StartSession(sessionID)
		if err != nil {
			return fmt.Errorf("failed to create session automatically: %w", err)
		}
		currentSession = newSession
		log.Printf("[DEBUG] Auto-created session: %s", currentSession.ID)
	} else {
		if currentSession.ID == "" {
			log.Printf("[DEBUG] ⚠️ Session exists but has empty ID!")
		} else {
			log.Printf("[DEBUG] Using existing session: %s", currentSession.ID)
		}
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

	// 设置上下文 - 使用类型安全的key
	ctxWithSession := context.WithValue(ctx, SessionIDKey, currentSession.ID)
	log.Printf("[DEBUG] 🔧 Context set with session ID: %s", currentSession.ID)

	// 执行流式ReAct循环
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

	// Memory generation removed

	// 发送完成信号
	if callback != nil {
		callback(StreamChunk{
			Type:             "complete",
			Content:          "Task completed",
			Complete:         true,
			TotalTokensUsed:  result.TokensUsed,
			PromptTokens:     result.PromptTokens,
			CompletionTokens: result.CompletionTokens,
		})
	}

	return nil
}

// ========== 公共接口 ==========

// GetAvailableTools - 获取可用工具列表
func (r *ReactAgent) GetAvailableTools() []string {
	tools := make([]string, 0, len(r.tools))
	for name := range r.tools {
		tools = append(tools, name)
	}
	return tools
}

// GetSessionHistory - 获取会话历史
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

// GenerateMemories - 手动生成记忆 (Memory module removed)
func (r *ReactAgent) GenerateMemories(ctx context.Context, sessionID string) (*GeneratedMemories, error) {
	return &GeneratedMemories{}, nil
}

// ========== 代理方法 ==========

// parseToolCalls - 委托给ToolExecutor
func (r *ReactAgent) parseToolCalls(message *llm.Message) []*types.ReactToolCall {
	return r.toolExecutor.parseToolCalls(message)
}

// executeSerialToolsStream - 委托给ToolExecutor
func (r *ReactAgent) executeSerialToolsStream(ctx context.Context, toolCalls []*types.ReactToolCall, callback StreamCallback) []*types.ReactToolResult {
	return r.toolExecutor.executeSerialToolsStream(ctx, toolCalls, callback)
}

// ========== 组件创建函数 ==========

// NewLightPromptBuilder - 创建轻量化提示构建器
func NewLightPromptBuilder() *LightPromptBuilder {
	promptLoader, err := prompts.NewPromptLoader()
	if err != nil {
		log.Printf("[ERROR] LightPromptBuilder: Failed to create prompt loader: %v", err)
		return &LightPromptBuilder{promptLoader: nil}
	}

	return &LightPromptBuilder{
		promptLoader: promptLoader,
	}
}

// GetMemoryStats - 获取内存统计信息 (Memory module removed)
func (r *ReactAgent) GetMemoryStats() map[string]interface{} {
	return map[string]interface{}{
		"memory_disabled": true,
	}
}

// integrateWithMCPTools - 集成MCP工具
func integrateWithMCPTools(configManager *config.Manager, builtinTools []builtin.Tool) []builtin.Tool {
	// 获取MCP配置
	configMCP := configManager.GetMCPConfig()
	if !configMCP.Enabled {
		log.Printf("[INFO] MCP integration is disabled")
		return builtinTools
	}

	// 转换配置格式
	mcpConfig := convertConfigToMCP(configMCP)

	// 创建MCP管理器
	mcpManager := mcp.NewManager(mcpConfig)

	// 启动MCP管理器
	ctx := context.Background()
	if err := mcpManager.Start(ctx); err != nil {
		log.Printf("[WARN] Failed to start MCP manager: %v", err)
		return builtinTools
	}

	// 集成工具
	allTools := mcpManager.IntegrateWithBuiltinTools(builtinTools)
	log.Printf("[INFO] Integrated %d MCP tools with %d builtin tools", len(allTools)-len(builtinTools), len(builtinTools))

	return allTools
}

// convertConfigToMCP - 转换配置格式从config包到mcp包
func convertConfigToMCP(configMCP *config.MCPConfig) *mcp.MCPConfig {
	mcpConfig := &mcp.MCPConfig{
		Enabled:         configMCP.Enabled,
		Servers:         make(map[string]*mcp.ServerConfig),
		GlobalTimeout:   configMCP.GlobalTimeout,
		AutoRefresh:     configMCP.AutoRefresh,
		RefreshInterval: configMCP.RefreshInterval,
	}

	// 转换服务器配置
	for id, configServer := range configMCP.Servers {
		mcpServer := &mcp.ServerConfig{
			ID:          configServer.ID,
			Name:        configServer.Name,
			Type:        mcp.SpawnerType(configServer.Type),
			Command:     configServer.Command,
			Args:        configServer.Args,
			Env:         configServer.Env,
			WorkDir:     configServer.WorkDir,
			AutoStart:   configServer.AutoStart,
			AutoRestart: configServer.AutoRestart,
			Timeout:     configServer.Timeout,
			Enabled:     configServer.Enabled,
		}
		mcpConfig.Servers[id] = mcpServer
	}

	// 转换安全配置
	if configMCP.Security != nil {
		mcpConfig.Security = &mcp.SecurityConfig{
			AllowedCommands:     configMCP.Security.AllowedCommands,
			BlockedCommands:     configMCP.Security.BlockedCommands,
			AllowedPackages:     configMCP.Security.AllowedPackages,
			BlockedPackages:     configMCP.Security.BlockedPackages,
			RequireConfirmation: configMCP.Security.RequireConfirmation,
			SandboxMode:         configMCP.Security.SandboxMode,
			MaxProcesses:        configMCP.Security.MaxProcesses,
			MaxMemoryMB:         configMCP.Security.MaxMemoryMB,
			AllowedEnvironment:  configMCP.Security.AllowedEnvironment,
			RestrictedPaths:     configMCP.Security.RestrictedPaths,
		}
	}

	// 转换日志配置
	if configMCP.Logging != nil {
		mcpConfig.Logging = &mcp.LoggingConfig{
			Level:        configMCP.Logging.Level,
			LogRequests:  configMCP.Logging.LogRequests,
			LogResponses: configMCP.Logging.LogResponses,
			LogFile:      configMCP.Logging.LogFile,
		}
	}

	return mcpConfig
}
