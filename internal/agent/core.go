package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	contextmgr "alex/internal/context"
	"alex/internal/llm"
	"alex/internal/memory"
	"alex/internal/session"
	"alex/pkg/types"
)

// ReactCore - 使用工具调用流程的ReactCore核心实现
type ReactCore struct {
	agent            *ReactAgent
	streamCallback   StreamCallback
	messageProcessor *MessageProcessor
	llmHandler       *LLMHandler
	toolHandler      *ToolHandler
	promptHandler    *PromptHandler
}

// NewReactCore - 创建ReAct核心实例
func NewReactCore(agent *ReactAgent) *ReactCore {
	llmClient, err := llm.GetLLMInstance(llm.BasicModel)
	if err != nil {
		log.Printf("[ERROR] NewReactCore: Failed to get LLM instance: %v", err)
		llmClient = nil
	}

	return &ReactCore{
		agent:            agent,
		messageProcessor: NewMessageProcessor(llmClient, agent.sessionManager),
		llmHandler:       NewLLMHandler(nil), // Will be set per request
		toolHandler:      NewToolHandler(agent.tools),
		promptHandler:    NewPromptHandler(agent.promptBuilder),
	}
}

// SolveTask - 使用工具调用流程的简化任务解决方法
func (rc *ReactCore) SolveTask(ctx context.Context, task string, streamCallback StreamCallback) (*types.ReactTaskResult, error) {
	// 设置流回调
	rc.streamCallback = streamCallback
	rc.llmHandler.streamCallback = streamCallback

	// 生成任务ID
	taskID := generateTaskID()

	// 初始化任务上下文
	taskCtx := types.NewReactTaskContext(taskID, task)

	// 决定是否使用流式处理
	isStreaming := streamCallback != nil
	if isStreaming {
		streamCallback(StreamChunk{Type: "status", Content: GetRandomProcessingMessage(), Metadata: map[string]any{"phase": "initialization"}})
	}

	// 构建系统提示（只需构建一次）
	systemPrompt := rc.promptHandler.buildToolDrivenTaskPrompt(taskCtx)
	messages := []llm.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

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

		// 第一次迭代更新消息列表，添加最新的会话内容
		if iteration == 1 {
			sess := rc.messageProcessor.GetCurrentSession(ctx, rc.agent)
			// 使用新的上下文管理器优化消息
			sessionMessages := sess.GetMessages()
			optimizedMessages, err := rc.agent.contextManager.OptimizeContext(ctx, sess.ID, sessionMessages)
			if err != nil {
				log.Printf("[WARN] Context optimization failed: %v", err)
				optimizedMessages = sessionMessages
			}
			llmMessages := rc.messageProcessor.ConvertSessionToLLM(optimizedMessages)
			messages = append(messages, llmMessages...)
		} else {
			sessionMessages := rc.messageProcessor.ConvertLLMToSession(messages)
			sessionMessages = rc.messageProcessor.compressMessages(sessionMessages)
			messages = rc.messageProcessor.ConvertSessionToLLM(sessionMessages)
		}
		// 构建可用工具列表 - 每轮都包含工具定义以确保模型能调用工具
		tools := rc.toolHandler.buildToolDefinitions()

		request := &llm.ChatRequest{
			Messages:   messages,
			ModelType:  llm.BasicModel,
			Tools:      tools,
			ToolChoice: "auto",
			Config:     rc.agent.llmConfig,
			MaxTokens:  rc.agent.llmConfig.MaxTokens,
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

		// Extract token usage from response using compatible method
		usage := response.GetUsage()
		tokensUsed := usage.GetTotalTokens()
		promptTokens := usage.GetPromptTokens()
		completionTokens := usage.GetCompletionTokens()

		// Update task context with token usage
		taskCtx.TokensUsed += tokensUsed
		taskCtx.PromptTokens += promptTokens
		taskCtx.CompletionTokens += completionTokens
		step.TokensUsed = tokensUsed

		// Send token usage via stream callback
		if isStreaming && tokensUsed > 0 {
			streamCallback(StreamChunk{
				Type:             "token_usage",
				Content:          fmt.Sprintf("Tokens used: %d (prompt: %d, completion: %d)", tokensUsed, promptTokens, completionTokens),
				TokensUsed:       tokensUsed,
				TotalTokensUsed:  taskCtx.TokensUsed,
				PromptTokens:     promptTokens,
				CompletionTokens: completionTokens,
				Metadata:         map[string]any{"iteration": iteration, "phase": "token_accounting"},
			})
		}

		if len(choice.Message.Content) > 0 && len(choice.Message.ToolCalls) > 0 {
			streamCallback(StreamChunk{
				Type:     "thinking_result",
				Content:  choice.Message.Content,
				Metadata: map[string]any{"iteration": iteration, "phase": "thinking_result"}})
		}
		// 添加assistant消息到对话历史和session
		// 重要修复：即使没有content，也要添加包含工具调用的assistant消息
		if len(choice.Message.Content) > 0 || len(choice.Message.ToolCalls) > 0 {
			log.Printf("[DEBUG] ReactCore: Adding assistant message - Content length: %d, ToolCalls: %d", len(choice.Message.Content), len(choice.Message.ToolCalls))
			messages = append(messages, choice.Message)
			// 同时添加到session以供memory系统学习
			rc.addMessageToSession(ctx, &choice.Message)
		}

		// 解析并执行工具调用
		toolCalls := rc.agent.parseToolCalls(&choice.Message)
		log.Printf("[DEBUG] ReactCore: Parsed %d tool calls", len(toolCalls))

		// 记录所有从LLM接收到的工具调用ID，用于验证响应完整性
		expectedToolCallIDs := make([]string, 0, len(choice.Message.ToolCalls))
		for _, tc := range choice.Message.ToolCalls {
			expectedToolCallIDs = append(expectedToolCallIDs, tc.ID)
			log.Printf("[DEBUG] ReactCore: Expected tool call ID: %s", tc.ID)
		}

		if len(toolCalls) > 0 {
			step.Action = "tool_execution"
			step.ToolCall = toolCalls // 记录所有工具调用

			// 执行工具调用
			toolResult := rc.agent.executeSerialToolsStream(ctx, toolCalls, streamCallback)
			step.Result = toolResult

			log.Printf("[DEBUG] ReactCore: Tool execution returned %d results", len(toolResult))
			for i, result := range toolResult {
				log.Printf("[DEBUG] ReactCore: Tool result %d - Tool: '%s', CallID: '%s', Success: %v", i, result.ToolName, result.CallID, result.Success)
			}

			// 将工具结果添加到对话历史和session
			if toolResult != nil {
				isGemini := strings.Contains(request.Config.BaseURL, "googleapis")
				log.Printf("[DEBUG] ReactCore: Building tool messages, isGemini: %v", isGemini)
				toolMessages := rc.toolHandler.buildToolMessages(toolResult, isGemini)
				log.Printf("[DEBUG] ReactCore: Built %d tool messages", len(toolMessages))

				for i, msg := range toolMessages {
					log.Printf("[DEBUG] ReactCore: Tool message %d - Role: '%s', ToolCallId: '%s'", i, msg.Role, msg.ToolCallId)
				}

				// 验证响应完整性：确保每个期望的工具调用ID都有对应的响应
				receivedIDs := make(map[string]bool)
				for _, msg := range toolMessages {
					if msg.ToolCallId != "" {
						receivedIDs[msg.ToolCallId] = true
					}
				}

				// 检查是否有缺失的响应
				var missingIDs []string
				for _, expectedID := range expectedToolCallIDs {
					if !receivedIDs[expectedID] {
						missingIDs = append(missingIDs, expectedID)
					}
				}

				// 如果有缺失的ID，生成fallback响应 - 加强错误处理
				if len(missingIDs) > 0 {
					log.Printf("[ERROR] ReactCore: Missing responses for tool call IDs: %v", missingIDs)
					log.Printf("[ERROR] ReactCore: Expected IDs: %v, Received IDs: %v", expectedToolCallIDs, func() []string {
						var received []string
						for id := range receivedIDs {
							received = append(received, id)
						}
						return received
					}())

					for _, missingID := range missingIDs {
						// 尝试找到对应的工具名称
						var toolName = "unknown"
						for _, tc := range choice.Message.ToolCalls {
							if tc.ID == missingID {
								toolName = tc.Function.Name
								break
							}
						}

						fallbackMsg := llm.Message{
							Role:       "tool",
							Content:    fmt.Sprintf("Tool execution failed: no response generated for %s", toolName),
							ToolCallId: missingID,
							Name:       toolName,
						}
						toolMessages = append(toolMessages, fallbackMsg)
						log.Printf("[ERROR] ReactCore: Generated fallback response for missing ID: %s (tool: %s)", missingID, toolName)
					}

					// 如果有缺失响应，通过流回调通知用户
					if isStreaming {
						streamCallback(StreamChunk{
							Type:     "tool_error",
							Content:  fmt.Sprintf("Warning: %d tool call(s) failed to generate proper responses", len(missingIDs)),
							Metadata: map[string]any{"missing_tool_calls": missingIDs},
						})
					}
				}

				messages = append(messages, toolMessages...)

				// 将工具消息添加到session供memory系统学习
				rc.addToolMessagesToSession(ctx, toolMessages, toolResult)

				step.Observation = rc.toolHandler.generateObservation(toolResult)
			}
		} else {
			finalAnswer := choice.Message.Content

			step.Action = "direct_answer"
			step.Observation = finalAnswer
			step.Duration = time.Since(step.Timestamp)
			taskCtx.History = append(taskCtx.History, step)

			result := buildFinalResult(taskCtx, finalAnswer, 0.8, true)
			result.TokensUsed = taskCtx.TokensUsed

			if isStreaming {
				streamCallback(StreamChunk{
					Type:     "final_answer",
					Content:  finalAnswer,
					Metadata: map[string]any{"iteration": iteration, "phase": "final_answer"}})
			}
			return result, nil
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
	}

	return buildFinalResult(taskCtx, "Maximum iterations reached without completion", 0.5, false), nil
}

// GetContextStats - 获取上下文统计信息
func (rc *ReactCore) GetContextStats(sess *session.Session) *contextmgr.ContextStats {
	return rc.messageProcessor.GetContextStats(sess)
}

// RestoreFullContext - 恢复完整上下文
func (rc *ReactCore) RestoreFullContext(sess *session.Session, backupID string) error {
	return rc.messageProcessor.RestoreFullContext(sess, backupID)
}

// addMessageToSession - 将LLM消息添加到session中供memory系统学习
func (rc *ReactCore) addMessageToSession(ctx context.Context, llmMsg *llm.Message) {
	// 获取当前会话
	sess := rc.messageProcessor.GetCurrentSession(ctx, rc.agent)
	if sess == nil {
		return // 没有会话则跳过
	}

	// 转换LLM消息为session消息格式
	sessionMsg := &session.Message{
		Role:      llmMsg.Role,
		Content:   llmMsg.Content,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"source":    "llm_response",
			"timestamp": time.Now().Unix(),
		},
	}

	// 转换工具调用信息
	if len(llmMsg.ToolCalls) > 0 {
		for _, tc := range llmMsg.ToolCalls {
			// 将Arguments字符串解析为map[string]interface{}
			var args map[string]interface{}
			if tc.Function.Arguments != "" {
				// 简单处理：如果是JSON字符串尝试解析，否则存为字符串
				args = map[string]interface{}{"raw": tc.Function.Arguments}
			}

			sessionMsg.ToolCalls = append(sessionMsg.ToolCalls, session.ToolCall{
				ID:   tc.ID,
				Name: tc.Function.Name,
				Args: args,
			})
		}
		sessionMsg.Metadata["has_tool_calls"] = true
		sessionMsg.Metadata["tool_count"] = len(llmMsg.ToolCalls)
	}

	// 添加到session
	sess.AddMessage(sessionMsg)
}

// addToolMessagesToSession - 将工具消息添加到session中供memory系统学习
func (rc *ReactCore) addToolMessagesToSession(ctx context.Context, toolMessages []llm.Message, toolResults []*types.ReactToolResult) {
	// 获取当前会话
	sess := rc.messageProcessor.GetCurrentSession(ctx, rc.agent)
	if sess == nil {
		return // 没有会话则跳过
	}

	// 处理每个工具消息
	for _, toolMsg := range toolMessages {
		sessionMsg := &session.Message{
			Role:      toolMsg.Role,
			Content:   toolMsg.Content,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"source":    "tool_result",
				"timestamp": time.Now().Unix(),
			},
		}

		// 保存tool_call_id到metadata中 - 这是关键修复
		if toolMsg.ToolCallId != "" {
			sessionMsg.Metadata["tool_call_id"] = toolMsg.ToolCallId
		}

		// 如果是工具结果消息，添加额外的元数据
		if toolMsg.Role == "tool" && len(toolResults) > 0 {
			// 尝试匹配对应的工具结果
			for _, result := range toolResults {
				if result != nil && toolMsg.ToolCallId == result.CallID {
					sessionMsg.Metadata["tool_name"] = result.ToolName
					sessionMsg.Metadata["tool_success"] = result.Success
					sessionMsg.Metadata["execution_time"] = result.Duration.Milliseconds()
					if result.Error != "" {
						sessionMsg.Metadata["tool_error"] = result.Error
					}
					break
				}
			}
		}

		// 添加到session
		sess.AddMessage(sessionMsg)
	}

	// 异步创建工具使用相关的memory
	if rc.agent.memoryManager != nil && len(toolResults) > 0 {
		go rc.createToolUsageMemory(sess.ID, toolResults)
	}
}

// createToolUsageMemory - 创建工具使用相关的记忆
func (rc *ReactCore) createToolUsageMemory(sessionID string, toolResults []*types.ReactToolResult) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR] Tool usage memory creation panic: %v", r)
		}
	}()

	if rc.agent.memoryManager == nil {
		return
	}

	// 统计工具使用情况
	var successfulTools, failedTools []string
	totalTime := time.Duration(0)

	for _, result := range toolResults {
		if result == nil {
			continue
		}

		if result.Success {
			successfulTools = append(successfulTools, result.ToolName)
		} else {
			failedTools = append(failedTools, result.ToolName)
		}
		totalTime += result.Duration
	}

	// 创建工具使用模式记忆
	if len(successfulTools) > 0 {
		memory := &memory.MemoryItem{
			ID:         fmt.Sprintf("tool_usage_%s_%d", sessionID, time.Now().UnixNano()),
			SessionID:  sessionID,
			Category:   memory.TaskHistory,
			Content:    fmt.Sprintf("Successfully used tools: %s (total time: %v)", strings.Join(successfulTools, ", "), totalTime),
			Importance: 0.7,
			Tags:       append([]string{"tool_usage", "success"}, successfulTools...),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			LastAccess: time.Now(),
			Metadata: map[string]interface{}{
				"successful_tools": successfulTools,
				"execution_time":   totalTime.Milliseconds(),
				"tool_count":       len(successfulTools),
			},
		}

		if err := rc.agent.memoryManager.Store(memory); err != nil {
			log.Printf("[WARN] Failed to store tool usage memory: %v", err)
		}
	}

	// 创建工具失败记忆
	if len(failedTools) > 0 {
		memory := &memory.MemoryItem{
			ID:         fmt.Sprintf("tool_failure_%s_%d", sessionID, time.Now().UnixNano()),
			SessionID:  sessionID,
			Category:   memory.ErrorPatterns,
			Content:    fmt.Sprintf("Failed tools: %s", strings.Join(failedTools, ", ")),
			Importance: 0.8, // 失败记忆更重要，用于避免重复错误
			Tags:       append([]string{"tool_failure", "error"}, failedTools...),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			LastAccess: time.Now(),
			Metadata: map[string]interface{}{
				"failed_tools":  failedTools,
				"failure_count": len(failedTools),
			},
		}

		if err := rc.agent.memoryManager.Store(memory); err != nil {
			log.Printf("[WARN] Failed to store tool failure memory: %v", err)
		}
	}
}
