package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"deep-coding-agent/internal/llm"
	"deep-coding-agent/pkg/types"
)

// ReactCore - ReAct循环的核心逻辑
type ReactCore struct {
	agent *ReactAgent
}

// NewReactCore - 创建ReAct核心实例
func NewReactCore(agent *ReactAgent) *ReactCore {
	return &ReactCore{agent: agent}
}

// SolveTask - 统一的任务解决方法，支持流式和非流式
func (rc *ReactCore) SolveTask(ctx context.Context, task string, streamCallback StreamCallback) (*types.LightTaskResult, error) {
	// 生成任务ID
	taskID := generateTaskID()
	log.Printf("[DEBUG] ReactCore: Generated task ID: %s", taskID)

	// 初始化任务上下文
	taskCtx := types.NewLightTaskContext(taskID, task)
	log.Printf("[DEBUG] ReactCore: Initialized task context for task: %s", task[:min(len(task), 100)])

	// 决定是否使用流式处理
	isStreaming := streamCallback != nil
	if isStreaming {
		streamCallback(StreamChunk{Type: "status", Content: "🧠 Starting analysis...", Metadata: map[string]interface{}{"phase": "initialization"}})
	}

	// 构建初始conversation messages (用于非流式) 或 prompt (用于流式)
	var messages []llm.Message
	var prompt string
	
	if isStreaming {
		prompt = rc.agent.promptBuilder.BuildTaskPrompt(task, taskCtx)
		log.Printf("[DEBUG] ReactCore: Built initial streaming prompt, length: %d characters", len(prompt))
	} else {
		messages = []llm.Message{
			{Role: "user", Content: task},
		}
		log.Printf("[DEBUG] ReactCore: Built initial conversation with %d messages", len(messages))
	}

	// 执行ReAct循环 - 限制25次迭代
	maxIterations := 25
	if isStreaming {
		maxIterations = rc.agent.config.MaxIterations
	}
	
	log.Printf("[DEBUG] ReactCore: Starting ReAct loop with max iterations: %d, streaming: %t", maxIterations, isStreaming)
	
	for iteration := 1; iteration <= maxIterations; iteration++ {
		log.Printf("[DEBUG] ReactCore: === Iteration %d/%d ===", iteration, maxIterations)
		step := types.LightExecutionStep{
			Number:    iteration,
			Timestamp: time.Now(),
		}
		log.Printf("[DEBUG] ReactCore: Step %d initialized at %v", iteration, step.Timestamp)

		// 1. Think Phase (纯思考阶段 - 不涉及工具调用)
		log.Printf("[DEBUG] ReactCore: Starting Think phase for iteration %d", iteration)
		if isStreaming {
			streamCallback(StreamChunk{
				Type: "thinking_start", 
				Content: fmt.Sprintf("🤔 Step %d: Pure thinking and analysis...", iteration),
				Metadata: map[string]interface{}{"iteration": iteration, "phase": "thinking"}})
		}
		
		var thought string
		var confidence float64
		var thinkTokens int
		var err error

		// 纯思考阶段：只分析，不调用工具
		if isStreaming {
			thought, confidence, thinkTokens, err = rc.agent.thinkingEngine.pureThink(ctx, prompt, taskCtx)
		} else {
			// 对于非流式模式，我们也使用纯思考，但需要基于消息历史
			conversationPrompt := rc.buildConversationPrompt(messages, taskCtx)
			thought, confidence, thinkTokens, err = rc.agent.thinkingEngine.pureThink(ctx, conversationPrompt, taskCtx)
		}

		if err != nil {
			log.Printf("[ERROR] ReactCore: Think phase failed at iteration %d: %v", iteration, err)
			if isStreaming {
				streamCallback(StreamChunk{Type: "error", Content: fmt.Sprintf("❌ Thinking failed: %v", err)})
			}
			return nil, fmt.Errorf("thinking failed at iteration %d: %w", iteration, err)
		}
		
		log.Printf("[DEBUG] ReactCore: Think phase completed - Thought length: %d, Confidence: %.2f, Tokens: %d",
			len(thought), confidence, thinkTokens)
		
		// Stream the thinking output
		if isStreaming && len(strings.TrimSpace(thought)) > 0 {
			streamCallback(StreamChunk{
				Type: "thinking_result", 
				Content: thought,
				Metadata: map[string]interface{}{
					"iteration": iteration, 
					"confidence": confidence,
					"tokens_used": thinkTokens,
					"phase": "pure_reasoning"}})
		}

		step.Thought = thought
		step.Confidence = confidence
		step.TokensUsed = thinkTokens
		taskCtx.TokensUsed += thinkTokens

		// 检查是否可以直接提供答案（在规划之前）
		if rc.agent.canProvideDirectAnswer(thought, confidence) {
			log.Printf("[DEBUG] ReactCore: Direct answer detected after thinking, completing task")
			step.Action = "provide_answer"
			step.Observation = thought
			step.Duration = time.Since(step.Timestamp)
			taskCtx.History = append(taskCtx.History, step)
			
			if isStreaming {
				streamCallback(StreamChunk{
					Type: "final_answer", 
					Content: thought,
					Metadata: map[string]interface{}{"confidence": confidence, "direct_answer": true}})
				streamCallback(StreamChunk{Type: "complete", Content: "✅ Task completed"})
			}
			
			log.Printf("[DEBUG] ReactCore: Task completed with direct answer in %d iterations", iteration)
			return rc.buildFinalResult(taskCtx, thought, confidence, true), nil
		}

		// 2. Plan Phase (规划阶段 - 基于思考结果决定行动)
		log.Printf("[DEBUG] ReactCore: Starting Plan phase for iteration %d", iteration)
		if isStreaming {
			streamCallback(StreamChunk{
				Type: "planning_start", 
				Content: fmt.Sprintf("📋 Step %d: Planning actions...", iteration),
				Metadata: map[string]interface{}{"iteration": iteration, "phase": "planning"}})
		}

		actionPlan, planTokens, err := rc.agent.thinkingEngine.planActions(ctx, thought, taskCtx)
		if err != nil {
			log.Printf("[ERROR] ReactCore: Plan phase failed at iteration %d: %v", iteration, err)
			if isStreaming {
				streamCallback(StreamChunk{Type: "error", Content: fmt.Sprintf("❌ Planning failed: %v", err)})
			}
			return nil, fmt.Errorf("planning failed at iteration %d: %w", iteration, err)
		}

		step.TokensUsed += planTokens
		taskCtx.TokensUsed += planTokens

		log.Printf("[DEBUG] ReactCore: Plan phase completed - HasActions: %t, ToolCalls: %d, CodeBlock: %t, Tokens: %d",
			actionPlan.HasActions, len(actionPlan.ToolCalls), actionPlan.CodeBlock != nil, planTokens)

		// Stream the planning result
		if isStreaming && len(strings.TrimSpace(actionPlan.Reasoning)) > 0 {
			streamCallback(StreamChunk{
				Type: "planning_result", 
				Content: actionPlan.Reasoning,
				Metadata: map[string]interface{}{
					"iteration": iteration,
					"has_actions": actionPlan.HasActions,
					"tool_calls_count": len(actionPlan.ToolCalls),
					"has_code": actionPlan.CodeBlock != nil,
					"tokens_used": planTokens,
					"phase": "action_planning"}})
		}

		// 如果没有规划的行动，说明任务可能已经完成
		if !actionPlan.HasActions {
			log.Printf("[DEBUG] ReactCore: No actions planned, treating as completion")
			step.Action = "no_action_needed"
			step.Observation = actionPlan.Reasoning
			step.Duration = time.Since(step.Timestamp)
			taskCtx.History = append(taskCtx.History, step)
			
			if isStreaming {
				streamCallback(StreamChunk{
					Type: "final_answer", 
					Content: actionPlan.Reasoning,
					Metadata: map[string]interface{}{"confidence": confidence, "no_actions": true}})
				streamCallback(StreamChunk{Type: "complete", Content: "✅ Task completed"})
			}
			
			log.Printf("[DEBUG] ReactCore: Task completed without actions in %d iterations", iteration)
			return rc.buildFinalResult(taskCtx, actionPlan.Reasoning, confidence, true), nil
		}

		// 3. Act Phase (执行阶段) - 执行规划的行动
		log.Printf("[DEBUG] ReactCore: Starting Act phase for iteration %d", iteration)
		if isStreaming {
			streamCallback(StreamChunk{
				Type: "action_start", 
				Content: fmt.Sprintf("⚡ Step %d: Executing planned actions...", iteration),
				Metadata: map[string]interface{}{"iteration": iteration, "phase": "executing"}})
		}
		var actionResult *types.LightToolResult

		// 执行规划的工具调用
		if len(actionPlan.ToolCalls) > 0 {
			log.Printf("[DEBUG] ReactCore: Executing %d planned tool calls", len(actionPlan.ToolCalls))
			step.Action = "tool_calls"
			// 保存第一个工具调用信息用于显示（如果有多个，只显示第一个）
			if len(actionPlan.ToolCalls) > 0 {
				step.ToolCall = actionPlan.ToolCalls[0]
			}
			
			if isStreaming {
				actionResult = rc.agent.executeParallelToolsStream(ctx, actionPlan.ToolCalls, streamCallback)
			} else {
				actionResult = rc.agent.executeParallelTools(ctx, actionPlan.ToolCalls)
			}
		} else if actionPlan.CodeBlock != nil {
			// 执行规划的代码
			log.Printf("[DEBUG] ReactCore: Executing planned code block - Language: %s, Code length: %d",
				actionPlan.CodeBlock.Language, len(actionPlan.CodeBlock.Code))
			step.Action = "code_execution"
			
			if isStreaming {
				streamCallback(StreamChunk{Type: "tool_start", Content: fmt.Sprintf("Executing %s code", actionPlan.CodeBlock.Language)})
			}
			
			actionResult = rc.executeCodeBlock(ctx, actionPlan.CodeBlock, isStreaming, streamCallback)
		} else {
			log.Printf("[DEBUG] ReactCore: No specific actions to execute, using plan reasoning")
			step.Action = "reasoning"
			if isStreaming {
				streamCallback(StreamChunk{
					Type: "reasoning_only", 
					Content: actionPlan.Reasoning,
					Metadata: map[string]interface{}{"iteration": iteration, "plan_only": true}})
			}
		}

		step.Result = actionResult

		// 如果执行了工具，将工具结果添加到conversation (非流式模式)
		if !isStreaming && actionResult != nil && (actionResult.Success || actionResult.Error != "") {
			toolMessages := rc.buildToolMessages(actionResult)
			messages = append(messages, toolMessages...)
			log.Printf("[DEBUG] ReactCore: Added %d tool result messages to conversation", len(toolMessages))
		}

		// 4. Observe Phase (观察阶段)
		log.Printf("[DEBUG] ReactCore: Starting Observe phase for iteration %d", iteration)
		if isStreaming {
			streamCallback(StreamChunk{
				Type: "observation_start", 
				Content: fmt.Sprintf("👁️ Step %d: Observing results...", iteration),
				Metadata: map[string]interface{}{"iteration": iteration, "phase": "observing"}})
		}
		var observation string
		
		if !isStreaming && actionResult != nil {
			// 非流式模式：让模型观察和分析工具结果
			observeResponse, observeThought, _, observeTokens, err := rc.agent.thinkWithConversation(ctx, messages, taskCtx)
			if err != nil {
				observation = rc.agent.observe(actionResult, confidence) // 回退到简单观察
			} else {
				observation = observeThought
				taskCtx.TokensUsed += observeTokens
				// 将观察结果也添加到conversation
				if observeResponse != nil && len(observeResponse.Choices) > 0 {
					messages = append(messages, llm.Message{
						Role:    "assistant",
						Content: observeResponse.Choices[0].Message.Content,
					})
				}
			}
		} else {
			observation = rc.agent.observe(actionResult, confidence)
		}

		step.Observation = observation
		step.Duration = time.Since(step.Timestamp)
		log.Printf("[DEBUG] ReactCore: Observe phase completed - Duration: %v, Observation length: %d",
			step.Duration, len(observation))
		
		// Stream the observation/analysis
		if isStreaming && len(strings.TrimSpace(observation)) > 0 {
			streamCallback(StreamChunk{
				Type: "observation_result", 
				Content: observation,
				Metadata: map[string]interface{}{
					"iteration": iteration, 
					"duration": step.Duration.String(),
					"phase": "analysis"}})
		}

		// 添加到历史记录
		taskCtx.History = append(taskCtx.History, step)
		taskCtx.LastUpdate = time.Now()

		// 检查完成条件 - 基于最新的观察和分析
		log.Printf("[DEBUG] ReactCore: Checking task completion (confidence: %.2f >= threshold: %.2f)",
			confidence, rc.agent.config.ConfidenceThreshold)
		
		if rc.agent.isTaskComplete(observation, confidence) {
			log.Printf("[DEBUG] ReactCore: Task completed successfully after %d iterations", iteration)
			if isStreaming {
				streamCallback(StreamChunk{
					Type: "task_complete", 
					Content: observation,
					Metadata: map[string]interface{}{
						"iterations": iteration, 
						"confidence": confidence,
						"success": true}})
				streamCallback(StreamChunk{Type: "complete", Content: "✅ Task completed successfully"})
			}
			return rc.buildFinalResult(taskCtx, observation, confidence, true), nil
		}

		// 准备下一轮迭代 (流式模式需要重新构建prompt)
		if isStreaming {
			if rc.agent.config.ContextCompression {
				prompt = rc.agent.contextMgr.CompressContext(taskCtx)
			} else {
				prompt = rc.agent.promptBuilder.BuildTaskPrompt(task, taskCtx)
			}
		}

		log.Printf("[DEBUG] ReactCore: Iteration %d completed, continuing to next iteration", iteration)
	}

	// 达到最大迭代次数
	log.Printf("[WARN] ReactCore: Maximum iterations (%d) reached without completion", maxIterations)
	if isStreaming {
		streamCallback(StreamChunk{
			Type: "max_iterations", 
			Content: fmt.Sprintf("⚠️ Reached maximum iterations (%d) without full completion", maxIterations),
			Metadata: map[string]interface{}{
				"max_iterations": maxIterations,
				"partial_completion": true}})
		streamCallback(StreamChunk{Type: "complete", Content: "⚠️ Maximum iterations reached"})
	}
	return rc.buildFinalResult(taskCtx, "Maximum iterations reached without completion", 0.5, false), nil
}

// executeCodeBlock - 执行代码块，支持流式和非流式
func (rc *ReactCore) executeCodeBlock(ctx context.Context, codeBlock *CodeBlock, isStreaming bool, streamCallback StreamCallback) *types.LightToolResult {
	log.Printf("[DEBUG] ReactCore: Executing %s code", codeBlock.Language)
	codeResult, err := rc.agent.codeExecutor.ExecuteCode(ctx, codeBlock.Language, codeBlock.Code)
	
	if err != nil {
		actionResult := &types.LightToolResult{
			Success: false,
			Error:   err.Error(),
		}
		if isStreaming {
			streamCallback(StreamChunk{Type: "tool_error", Content: err.Error()})
		}
		return actionResult
	}

	actionResult := &types.LightToolResult{
		Success: codeResult.Success,
		Content: codeResult.Output,
		Error:   codeResult.Error,
		Metadata: map[string]interface{}{
			"language":       codeResult.Language,
			"execution_time": codeResult.ExecutionTime,
			"exit_code":      codeResult.ExitCode,
		},
	}

	if isStreaming {
		if codeResult.Success {
			streamCallback(StreamChunk{Type: "tool_result", Content: codeResult.Output})
		} else {
			streamCallback(StreamChunk{Type: "tool_error", Content: codeResult.Error})
		}
	}

	return actionResult
}

// buildToolMessages - 构建工具结果消息
func (rc *ReactCore) buildToolMessages(actionResult *types.LightToolResult) []llm.Message {
	var toolMessages []llm.Message
	
	if len(actionResult.ToolCalls) > 0 {
		// 处理多个工具调用的结果
		for _, toolCall := range actionResult.ToolCalls {
			toolMessage := llm.Message{
				Role:       "tool",
				Content:    actionResult.Content,
				ToolCallId: toolCall.CallID,
			}
			if !actionResult.Success {
				toolMessage.Content = actionResult.Error
			}
			toolMessages = append(toolMessages, toolMessage)
		}
	} else {
		// 处理单个工具或代码执行结果
		content := actionResult.Content
		if !actionResult.Success {
			content = actionResult.Error
		}
		toolMessages = append(toolMessages, llm.Message{
			Role:    "tool",
			Content: content,
		})
	}
	
	return toolMessages
}

// buildConversationPrompt - 基于消息历史构建对话prompt
func (rc *ReactCore) buildConversationPrompt(messages []llm.Message, taskCtx *types.LightTaskContext) string {
	var parts []string
	
	// 添加任务目标
	parts = append(parts, fmt.Sprintf("Task Goal: %s", taskCtx.Goal))
	
	// 添加消息历史的摘要
	for _, msg := range messages {
		if msg.Role == "user" {
			parts = append(parts, fmt.Sprintf("User: %s", msg.Content))
		} else if msg.Role == "assistant" {
			parts = append(parts, fmt.Sprintf("Assistant: %s", msg.Content))
		}
	}
	
	// 添加执行历史摘要
	if len(taskCtx.History) > 0 {
		parts = append(parts, "Previous steps:")
		for i, step := range taskCtx.History {
			parts = append(parts, fmt.Sprintf("Step %d: %s", i+1, step.Thought))
		}
	}
	
	return strings.Join(parts, "\n")
}

// buildFinalResult - 构建最终结果
func (rc *ReactCore) buildFinalResult(taskCtx *types.LightTaskContext, answer string, confidence float64, success bool) *types.LightTaskResult {
	totalDuration := time.Since(taskCtx.StartTime)
	log.Printf("[DEBUG] ReactCore: Final stats - Duration: %v, Tokens: %d, Steps: %d, Success: %t",
		totalDuration, taskCtx.TokensUsed, len(taskCtx.History), success)
	
	return &types.LightTaskResult{
		Success:    success,
		Answer:     answer,
		Confidence: confidence,
		Steps:      taskCtx.History,
		Duration:   totalDuration,
		TokensUsed: taskCtx.TokensUsed,
	}
}