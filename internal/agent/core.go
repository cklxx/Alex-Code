package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	contextmgr "alex/internal/context"
	"alex/internal/llm"
	"alex/internal/session"
	"alex/pkg/types"
)

// ReactCore - 使用工具调用流程的ReactCore核心实现
type ReactCore struct {
	agent          *ReactAgent
	streamCallback StreamCallback
	contextHandler *ContextHandler
	llmHandler     *LLMHandler
	toolHandler    *ToolHandler
	promptHandler  *PromptHandler
}

// NewReactCore - 创建ReAct核心实例
func NewReactCore(agent *ReactAgent) *ReactCore {
	llmClient, err := llm.GetLLMInstance(llm.BasicModel)
	if err != nil {
		log.Printf("[ERROR] NewReactCore: Failed to get LLM instance: %v", err)
		llmClient = nil
	}

	return &ReactCore{
		agent:          agent,
		contextHandler: NewContextHandler(llmClient, agent.sessionManager),
		llmHandler:     NewLLMHandler(nil), // Will be set per request
		toolHandler:    NewToolHandler(agent.tools),
		promptHandler:  NewPromptHandler(agent.promptBuilder),
	}
}

// SolveTask - 使用工具调用流程的简化任务解决方法
func (rc *ReactCore) SolveTask(ctx context.Context, task string, streamCallback StreamCallback) (*types.ReactTaskResult, error) {
	// 设置流回调
	rc.streamCallback = streamCallback
	rc.llmHandler.streamCallback = streamCallback

	// 获取当前会话
	sess := rc.contextHandler.getCurrentSession(ctx, rc.agent)
	if sess != nil {
		// 检查并处理上下文溢出
		if err := rc.contextHandler.handleContextOverflow(ctx, sess, streamCallback); err != nil {
			log.Printf("[WARNING] Context overflow handling failed: %v", err)
		}
	}

	// 生成任务ID
	taskID := generateTaskID()

	// 初始化任务上下文
	taskCtx := types.NewReactTaskContext(taskID, task)

	// 决定是否使用流式处理
	isStreaming := streamCallback != nil
	if isStreaming {
		streamCallback(StreamChunk{Type: "status", Content: "🧠 Starting tool-driven ReAct process...", Metadata: map[string]any{"phase": "initialization"}})
	}

	// 构建消息列表，基于会话历史
	systemPrompt := rc.promptHandler.buildToolDrivenTaskPrompt()
	messages := rc.contextHandler.buildMessagesFromSession(sess, task, systemPrompt)

	// 执行工具驱动的ReAct循环
	maxIterations := 25 // 减少迭代次数，依赖智能工具调用

	for iteration := 1; iteration <= maxIterations; iteration++ {
		step := types.ReactExecutionStep{
			Number:    iteration,
			Timestamp: time.Now(),
		}

		if isStreaming {
			streamCallback(StreamChunk{
				Type:     "iteration",
				Content:  fmt.Sprintf("🔄 Iteration %d: Processing with tool-driven approach...", iteration),
				Metadata: map[string]any{"iteration": iteration, "phase": "tool_driven_processing"}})
		}

		// 构建可用工具列表 - 每轮都包含工具定义以确保模型能调用工具
		tools := rc.toolHandler.buildToolDefinitions()

		request := &llm.ChatRequest{
			Messages:   messages,
			ModelType:  llm.BasicModel,
			Tools:      tools,
			ToolChoice: "auto",
			Config:     rc.agent.llmConfig,
			MaxTokens:  12000,
		}

		// 获取LLM实例
		client, err := llm.GetLLMInstance(llm.BasicModel)
		if err != nil {
			log.Printf("[ERROR] ReactCore: Failed to get LLM instance at iteration %d: %v", iteration, err)
			if isStreaming {
				streamCallback(StreamChunk{Type: "error", Content: fmt.Sprintf("❌ LLM initialization failed: %v", err)})
			}
			return nil, fmt.Errorf("LLM initialization failed at iteration %d: %w", iteration, err)
		}

		// 添加请求参数验证
		if err := rc.llmHandler.validateLLMRequest(request); err != nil {
			log.Printf("[ERROR] ReactCore: Invalid LLM request at iteration %d: %v", iteration, err)
			if isStreaming {
				streamCallback(StreamChunk{Type: "error", Content: fmt.Sprintf("❌ Invalid request: %v", err)})
			}
			return nil, fmt.Errorf("invalid LLM request at iteration %d: %w", iteration, err)
		}

		// 执行LLM调用，带重试机制
		response, err := rc.llmHandler.callLLMWithRetry(ctx, client, request, 3)
		if err != nil {
			log.Printf("[ERROR] ReactCore: LLM call failed at iteration %d after retries: %v", iteration, err)
			if isStreaming {
				streamCallback(StreamChunk{Type: "error", Content: fmt.Sprintf("❌ LLM call failed: %v", err)})
			}
			return nil, fmt.Errorf("LLM call failed at iteration %d: %w", iteration, err)
		}

		// 增强的响应验证
		if response == nil {
			err := fmt.Errorf("received nil response from LLM at iteration %d", iteration)
			log.Printf("[ERROR] ReactCore: %v", err)
			if isStreaming {
				streamCallback(StreamChunk{Type: "error", Content: "❌ Received empty response from LLM"})
			}
			return nil, err
		}

		if len(response.Choices) == 0 {
			log.Printf("[ERROR] ReactCore: No response choices at iteration %d", iteration)
			if isStreaming {
				streamCallback(StreamChunk{Type: "error", Content: "❌ No response choices from LLM - API response format issue"})
			}
			return nil, fmt.Errorf("no response choices received at iteration %d - API response format issue", iteration)
		}

		log.Printf("DEBUG: Response: %+v", response)
		choice := response.Choices[0]
		step.Thought = strings.TrimSpace(choice.Message.Content)
		// 添加assistant消息到对话历史
		if len(choice.Message.Content) > 0 {
			messages = append(messages, choice.Message)
		}
		// 解析并执行工具调用
		toolCalls := rc.agent.parseToolCalls(&choice.Message)
		if len(toolCalls) > 0 {
			step.Action = "tool_execution"
			step.ToolCall = toolCalls[0] // 记录第一个工具调用

			if isStreaming {
				streamCallback(StreamChunk{
					Type:     "tool_start",
					Content:  fmt.Sprintf("⚡ Executing %d tool(s): %s", len(toolCalls), rc.toolHandler.formatToolNames(toolCalls)),
					Metadata: map[string]any{"iteration": iteration, "tools": rc.toolHandler.formatToolNames(toolCalls)}})
			}

			// 执行工具调用
			toolResult := rc.agent.executeSerialToolsStream(ctx, toolCalls, streamCallback)

			step.Result = toolResult

			// 将工具结果添加到对话历史
			if toolResult != nil {
				toolMessages := rc.toolHandler.buildToolMessages(toolResult)
				messages = append(messages, toolMessages...)

				step.Observation = rc.toolHandler.generateObservation(toolResult)
			}
		} else {
			finalAnswer := choice.Message.Content

			if isStreaming {
				streamCallback(StreamChunk{
					Type:     "final_answer",
					Content:  finalAnswer,
					Metadata: map[string]any{"iteration": iteration}})
				streamCallback(StreamChunk{Type: "complete", Content: "✅ Task completed"})
			}

			step.Action = "direct_answer"
			step.Observation = finalAnswer
			step.Duration = time.Since(step.Timestamp)
			taskCtx.History = append(taskCtx.History, step)

			return buildFinalResult(taskCtx, finalAnswer, 0.8, true), nil
		}

		step.Duration = time.Since(step.Timestamp)
		taskCtx.History = append(taskCtx.History, step)
		taskCtx.LastUpdate = time.Now()
	}

	// 达到最大迭代次数
	log.Printf("[WARN] ReactCore: Maximum iterations (%d) reached", maxIterations)
	if isStreaming {
		streamCallback(StreamChunk{
			Type:     "max_iterations",
			Content:  fmt.Sprintf("⚠️ Reached maximum iterations (%d)", maxIterations),
			Metadata: map[string]any{"max_iterations": maxIterations}})
		streamCallback(StreamChunk{Type: "complete", Content: "⚠️ Maximum iterations reached"})
	}

	return buildFinalResult(taskCtx, "Maximum iterations reached without completion", 0.5, false), nil
}

// GetContextStats - 获取上下文统计信息
func (rc *ReactCore) GetContextStats(sess *session.Session) *contextmgr.ContextStats {
	return rc.contextHandler.GetContextStats(sess)
}

// ForceContextSummarization - 强制进行上下文总结
func (rc *ReactCore) ForceContextSummarization(ctx context.Context, sess *session.Session) (*contextmgr.ContextProcessingResult, error) {
	return rc.contextHandler.ForceContextSummarization(ctx, sess)
}

// RestoreFullContext - 恢复完整上下文
func (rc *ReactCore) RestoreFullContext(sess *session.Session, backupID string) error {
	return rc.contextHandler.RestoreFullContext(sess, backupID)
}
