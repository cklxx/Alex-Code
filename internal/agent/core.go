package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"alex/internal/llm"
	"alex/pkg/types"
)

// ReactCore - 使用工具调用流程的ReactCore核心实现
type ReactCore struct {
	agent          *ReactAgent
	streamCallback StreamCallback // 当前流回调
}

// NewReactCore - 创建ReAct核心实例
func NewReactCore(agent *ReactAgent) *ReactCore {
	return &ReactCore{agent: agent}
}

// SolveTask - 使用工具调用流程的简化任务解决方法
func (rc *ReactCore) SolveTask(ctx context.Context, task string, streamCallback StreamCallback) (*types.ReactTaskResult, error) {
	// 设置流回调
	rc.streamCallback = streamCallback

	// 生成任务ID
	taskID := generateTaskID()

	// 初始化任务上下文
	taskCtx := types.NewReactTaskContext(taskID, task)

	// 决定是否使用流式处理
	isStreaming := streamCallback != nil
	if isStreaming {
		streamCallback(StreamChunk{Type: "status", Content: "🧠 Starting tool-driven ReAct process...", Metadata: map[string]any{"phase": "initialization"}})
	}

	// 使用简化的系统消息，避免token过多导致API错误
	messages := []llm.Message{
		{Role: "system", Content: rc.buildToolDrivenTaskPrompt()},
		{Role: "system", Content: rc.agent.contextMgr.CompressContext(taskCtx)},
		{Role: "user", Content: task + "\n\n think about the task and break it down into a list of todos and then call the todo_update tool to create the todos"},
	}

	// 执行工具驱动的ReAct循环
	maxIterations := 10 // 减少迭代次数，依赖智能工具调用

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
		tools := rc.buildToolDefinitions()
		toolChoice := "auto"

		request := &llm.ChatRequest{
			Messages:   messages,
			ModelType:  llm.BasicModel,
			Tools:      tools,
			ToolChoice: toolChoice,
			Config:     rc.agent.llmConfig,
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
		if err := rc.validateLLMRequest(request); err != nil {
			log.Printf("[ERROR] ReactCore: Invalid LLM request at iteration %d: %v", iteration, err)
			if isStreaming {
				streamCallback(StreamChunk{Type: "error", Content: fmt.Sprintf("❌ Invalid request: %v", err)})
			}
			return nil, fmt.Errorf("invalid LLM request at iteration %d: %w", iteration, err)
		}

		// 执行LLM调用，带重试机制
		response, err := rc.callLLMWithRetry(ctx, client, request, 3)
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
					Content:  fmt.Sprintf("⚡ Executing %d tool(s): %s", len(toolCalls), rc.formatToolNames(toolCalls)),
					Metadata: map[string]any{"iteration": iteration, "tools": rc.formatToolNames(toolCalls)}})
			}

			// 执行工具调用
			toolResult := rc.agent.executeSerialToolsStream(ctx, toolCalls, streamCallback)

			step.Result = toolResult

			// 将工具结果添加到对话历史
			if toolResult != nil {
				toolMessages := rc.buildToolMessages(toolResult)
				messages = append(messages, toolMessages...)

				step.Observation = rc.generateObservation(toolResult)
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

			return rc.buildFinalResult(taskCtx, finalAnswer, 0.8, true), nil
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

	return rc.buildFinalResult(taskCtx, "Maximum iterations reached without completion", 0.5, false), nil
}

// buildToolDrivenTaskPrompt - 构建工具驱动的任务提示
func (rc *ReactCore) buildToolDrivenTaskPrompt() string {
	// 使用项目内的prompt builder
	if rc.agent.promptBuilder != nil && rc.agent.promptBuilder.promptLoader != nil {
		// 尝试使用React thinking prompt作为基础模板
		template, err := rc.agent.promptBuilder.promptLoader.GetReActThinkingPrompt()
		if err != nil {
			log.Printf("[WARN] ReactCore: Failed to get ReAct thinking prompt, trying fallback: %v", err)
		}
		// 构建增强的任务提示，将特定任务信息与ReAct模板结合
		return template
	}

	// Fallback to hardcoded prompt if prompt builder is not available
	log.Printf("[WARN] ReactCore: Prompt builder not available, using hardcoded prompt")
	return rc.buildHardcodedTaskPrompt()
}

// buildHardcodedTaskPrompt - 构建硬编码的任务提示（fallback）
func (rc *ReactCore) buildHardcodedTaskPrompt() string {

	return fmt.Sprintf(`You are an intelligent agent with access to powerful tools. Your goal is to complete this task efficiently:

**time:** %s


**Approach:**
1. **For complex tasks**: Start with the 'think' tool to analyze and plan
2. **For multi-step tasks**: Use 'todo_update' to create structured task lists
3. **For file operations**: Use appropriate file tools (file_read, file_update, etc.)
4. **For system operations**: Use bash tool when needed
5. **For search/analysis**: Use grep or other search tools

**Think Tool Capabilities:**
- Phase: analyze, plan, reflect, reason, ultra_think
- Depth: shallow, normal, deep, ultra
- Use for strategic thinking and problem breakdown

**Todo Management:**
- todo_update: Create, batch create, update, complete tasks
- todo_read: Read current todos with filtering and statistics

**Guidelines:**
- Use the 'think' tool first for complex problems requiring analysis
- Break down multi-step tasks using todo_update
- Execute tools systematically to achieve the goal
- Provide clear, actionable results

Begin by determining the best approach for this task.`, time.Now().Format(time.RFC3339))
}

// buildToolDefinitions - 构建工具定义列表（包括think工具）
func (rc *ReactCore) buildToolDefinitions() []llm.Tool {
	var tools []llm.Tool

	for _, tool := range rc.agent.tools {
		toolDef := llm.Tool{
			Type: "function",
			Function: llm.Function{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			},
		}

		tools = append(tools, toolDef)
	}

	return tools
}

// buildToolMessages - 构建工具结果消息
func (rc *ReactCore) buildToolMessages(actionResult []*types.ReactToolResult) []llm.Message {
	var toolMessages []llm.Message

	for _, result := range actionResult {
		content := result.Content
		if !result.Success {
			content = result.Error
		}

		// Ensure CallID is not empty - generate one if missing
		callID := result.CallID
		if callID == "" {
			callID = fmt.Sprintf("tool_%d", time.Now().UnixNano())
			log.Printf("[WARN] buildToolMessages: Missing CallID for tool %s, generated: %s", result.ToolName, callID)
		}

		toolMessages = append(toolMessages, llm.Message{
			Role:       "tool",
			Content:    content,
			Name:       result.ToolName,
			ToolCallId: callID,
		})
	}

	return toolMessages
}

// generateObservation - 生成观察结果
func (rc *ReactCore) generateObservation(toolResult []*types.ReactToolResult) string {
	if toolResult == nil {
		return "No tool execution result to observe"
	}

	for _, result := range toolResult {
		if result.Success {
			// 检查是否是特定工具的结果
			if len(result.ToolCalls) > 0 {
				toolName := result.ToolCalls[0].Name
				// 清理工具输出，移除冗余格式信息
				cleanContent := rc.cleanToolOutput(result.Content)
				switch toolName {
				case "think":
					return fmt.Sprintf("🧠 Thinking completed: %s", rc.truncateContent(cleanContent, 100))
				case "todo_update":
					return fmt.Sprintf("📋 Todo management: %s", rc.truncateContent(cleanContent, 100))
				case "file_read":
					return fmt.Sprintf("📖 File read: %s", rc.truncateContent(cleanContent, 100))
				case "bash":
					return fmt.Sprintf("⚡ Command executed: %s", rc.truncateContent(cleanContent, 100))
				default:
					return fmt.Sprintf("✅ %s completed: %s", toolName, rc.truncateContent(cleanContent, 100))
				}
			}
			return fmt.Sprintf("✅ Tool execution successful: %s", rc.truncateContent(rc.cleanToolOutput(toolResult[0].Content), 100))
		} else {
			return fmt.Sprintf("❌ Tool execution failed: %s", result.Error)
		}
	}
	return "No tool execution result to observe"
}

// formatToolNames - 格式化工具名称列表
func (rc *ReactCore) formatToolNames(toolCalls []*types.ReactToolCall) string {
	var names []string
	for _, tc := range toolCalls {
		names = append(names, tc.Name)
	}
	return strings.Join(names, ", ")
}

// truncateContent - 截断内容到指定长度
func (rc *ReactCore) truncateContent(content string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(content) <= maxLen {
		return content
	}
	// 确保不会越界
	if maxLen > len(content) {
		maxLen = len(content)
	}
	return content[:maxLen] + "..."
}

// buildFinalResult - 构建最终结果
func (rc *ReactCore) buildFinalResult(taskCtx *types.ReactTaskContext, answer string, confidence float64, success bool) *types.ReactTaskResult {
	totalDuration := time.Since(taskCtx.StartTime)

	return &types.ReactTaskResult{
		Success:    success,
		Answer:     answer,
		Confidence: confidence,
		Steps:      taskCtx.History,
		Duration:   totalDuration,
		TokensUsed: taskCtx.TokensUsed,
	}
}

// validateLLMRequest - 验证LLM请求参数
func (rc *ReactCore) validateLLMRequest(request *llm.ChatRequest) error {
	if request == nil {
		return fmt.Errorf("request is nil")
	}

	if len(request.Messages) == 0 {
		return fmt.Errorf("no messages in request")
	}

	if request.Config == nil {
		return fmt.Errorf("config is nil")
	}

	return nil
}

// callLLMWithRetry - 带重试机制的流式LLM调用
func (rc *ReactCore) callLLMWithRetry(ctx context.Context, client llm.Client, request *llm.ChatRequest, maxRetries int) (*llm.ChatResponse, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 使用流式调用
		streamChan, err := client.ChatStream(ctx, request)
		if err != nil {
			lastErr = err
			log.Printf("[WARN] ReactCore: Stream initialization failed (attempt %d): %v", attempt, err)

			// 检查是否是500错误，如果是，说明请求格式可能有问题，不要重试
			if strings.Contains(err.Error(), "500") {
				log.Printf("[ERROR] ReactCore: Server error 500, not retrying: %v", err)
				return nil, fmt.Errorf("server error 500 - request format issue: %w", err)
			}

			if attempt < maxRetries {
				backoffDuration := time.Duration(attempt*2) * time.Second
				log.Printf("[WARN] ReactCore: Retrying in %v", backoffDuration)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(backoffDuration):
					continue
				}
			}
			continue
		}

		// 处理流式响应并重构为完整响应
		response, err := rc.collectStreamingResponse(ctx, streamChan)
		if err == nil && response != nil {
			return response, nil
		}

		lastErr = err
		log.Printf("[WARN] ReactCore: Failed to collect streaming response (attempt %d): %v", attempt, err)

		if attempt < maxRetries {
			backoffDuration := time.Duration(attempt*2) * time.Second
			log.Printf("[WARN] ReactCore: Retrying in %v", backoffDuration)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoffDuration):
				continue
			}
		}
	}

	return nil, fmt.Errorf("streaming LLM call failed after %d attempts: %w", maxRetries, lastErr)
}

// collectStreamingResponse - 收集流式响应并重构为完整响应
func (rc *ReactCore) collectStreamingResponse(ctx context.Context, streamChan <-chan llm.StreamDelta) (*llm.ChatResponse, error) {
	var response *llm.ChatResponse
	var contentBuilder strings.Builder
	var toolCalls []llm.ToolCall
	var currentToolCall *llm.ToolCall

	// 检查是否有流回调需要通知
	hasStreamCallback := rc.streamCallback != nil

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case delta, ok := <-streamChan:
			if !ok {
				// 流结束，构建最终响应
				if response == nil {
					return nil, fmt.Errorf("no response received from stream")
				}

				// 设置最终的消息内容
				if len(response.Choices) > 0 {
					response.Choices[0].Message.Content = contentBuilder.String()
					if len(toolCalls) > 0 {
						response.Choices[0].Message.ToolCalls = toolCalls
					}
				}
				return response, nil
			}

			// 初始化响应对象
			if response == nil {
				response = &llm.ChatResponse{
					ID:      delta.ID,
					Object:  delta.Object,
					Created: delta.Created,
					Model:   delta.Model,
					Choices: make([]llm.Choice, 1),
				}
				response.Choices[0] = llm.Choice{
					Index: 0,
					Message: llm.Message{
						Role: "assistant",
					},
				}
			}

			// 处理每个delta中的choice
			if len(delta.Choices) > 0 {
				choice := delta.Choices[0]

				// 处理内容增量
				if choice.Delta.Content != "" {
					contentBuilder.WriteString(choice.Delta.Content)

					// 如果启用流式，实时显示LLM输出内容
					if hasStreamCallback {
						rc.streamCallback(StreamChunk{
							Type:     "llm_content",
							Content:  choice.Delta.Content,
							Metadata: map[string]any{"streaming": true},
						})
					}
				}

				// 处理工具调用增量
				if len(choice.Delta.ToolCalls) > 0 {
					for _, deltaToolCall := range choice.Delta.ToolCalls {
						if deltaToolCall.ID != "" {
							// 新的工具调用
							newToolCall := llm.ToolCall{
								ID:   deltaToolCall.ID,
								Type: deltaToolCall.Type,
								Function: llm.Function{
									Name:      deltaToolCall.Function.Name,
									Arguments: deltaToolCall.Function.Arguments,
								},
							}
							toolCalls = append(toolCalls, newToolCall)
							currentToolCall = &toolCalls[len(toolCalls)-1]
						} else if currentToolCall != nil {
							// 继续现有工具调用
							if deltaToolCall.Function.Name != "" {
								currentToolCall.Function.Name += deltaToolCall.Function.Name
							}
							if deltaToolCall.Function.Arguments != "" {
								currentToolCall.Function.Arguments += deltaToolCall.Function.Arguments
							}
						}
					}
				}

				// 检查完成原因
				if choice.FinishReason != "" {
					response.Choices[0].FinishReason = choice.FinishReason
				}
			}
		}
	}
}

// cleanToolOutput - 清理工具输出，只保留工具调用格式
func (rc *ReactCore) cleanToolOutput(content string) string {
	lines := strings.Split(content, "\n")
	var cleanLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 只保留🔧工具调用格式的行，其他格式的行都移除
		if strings.HasPrefix(trimmedLine, "🔧 ") {
			cleanLines = append(cleanLines, trimmedLine)
		}
	}

	// 如果没有找到工具调用格式，返回简洁的摘要
	if len(cleanLines) == 0 {
		return rc.truncateContent(content, 50)
	}

	return strings.Join(cleanLines, "\n")
}
