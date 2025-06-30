package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"deep-coding-agent/internal/llm"
	"deep-coding-agent/pkg/types"
)

// ThinkingEngine - 思考引擎
type ThinkingEngine struct {
	agent *ReactAgent
}

// ActionPlan - 行动规划结果
type ActionPlan struct {
	ToolCalls  []*types.LightToolCall `json:"tool_calls,omitempty"`
	CodeBlock  *CodeBlock             `json:"code_block,omitempty"`
	Reasoning  string                 `json:"reasoning"`
	HasActions bool                   `json:"has_actions"`
}

// NewThinkingEngine - 创建思考引擎
func NewThinkingEngine(agent *ReactAgent) *ThinkingEngine {
	return &ThinkingEngine{agent: agent}
}

// pureThink - 纯思考阶段，不涉及工具调用
func (te *ThinkingEngine) pureThink(ctx context.Context, prompt string, taskCtx *types.LightTaskContext) (string, float64, int, error) {
	// 构建纯思考的prompt，明确指示不要调用工具
	thinkingPrompt := fmt.Sprintf(`%s

IMPORTANT: This is the THINKING phase. You should only analyze and reason about the situation. 
DO NOT call any tools or functions. DO NOT plan specific actions yet.
Just think through the problem step by step and provide your analysis.

Current situation analysis:`, prompt)

	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: thinkingPrompt},
		},
		ModelType:  llm.BasicModel,
		Tools:      nil,    // 明确不提供工具
		ToolChoice: "none", // 禁用工具调用
		Config:     te.agent.llmConfig,
	}

	// 获取LLM实例
	client, err := llm.GetLLMInstance(llm.BasicModel)
	if err != nil {
		return "", 0.0, 0, fmt.Errorf("failed to get LLM instance: %w", err)
	}

	// 重试机制
	var response *llm.ChatResponse
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		response, err = client.Chat(ctx, request)
		if err == nil {
			break
		}

		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * time.Second
			time.Sleep(waitTime)
			continue
		}

		return "", 0.0, 0, fmt.Errorf("LLM call failed after %d attempts: %w", maxRetries, err)
	}

	if len(response.Choices) == 0 {
		return "", 0.0, 0, fmt.Errorf("no response choices received")
	}

	choice := response.Choices[0]
	thought := strings.TrimSpace(choice.Message.Content)
	confidence := te.calculateConfidence(thought, taskCtx)

	// 计算token使用量
	tokensUsed := response.Usage.TotalTokens
	if tokensUsed == 0 {
		tokensUsed = len(strings.Fields(thought)) + len(strings.Fields(prompt))
	}

	return thought, confidence, tokensUsed, nil
}

// planActions - 行动规划阶段，基于思考结果决定需要执行的工具
func (te *ThinkingEngine) planActions(ctx context.Context, thought string, taskCtx *types.LightTaskContext) (*ActionPlan, int, error) {
	// 构建行动规划的prompt
	planningPrompt := fmt.Sprintf(`Based on the previous thinking:
"%s"

IMPORTANT: You are in the ACTION PLANNING phase. You MUST either:
1. Call the appropriate tools immediately to gather information or perform tasks
2. Execute code if needed
3. ONLY if you already have complete information, provide a final answer

DO NOT ask for permission or confirmation. DO NOT say "Would you like me to...". 
Just execute the necessary tools to accomplish the goal.

Goal: %s
Steps completed: %d
Available tools: file_read, file_list, file_update, bash, grep, directory_create, etc.

Execute the next required action immediately:`,
		thought, taskCtx.Goal, len(taskCtx.History))

	// 构建可用工具列表
	tools := te.buildToolDefinitions()

	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: planningPrompt},
		},
		ModelType:  llm.BasicModel,
		Tools:      tools,
		ToolChoice: "auto", // 让模型自动决定是否调用工具
		Config:     te.agent.llmConfig,
	}

	// 获取LLM实例
	client, err := llm.GetLLMInstance(llm.BasicModel)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get LLM instance: %w", err)
	}

	// 重试机制
	var response *llm.ChatResponse
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		response, err = client.Chat(ctx, request)
		if err == nil {
			break
		}

		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * time.Second
			time.Sleep(waitTime)
			continue
		}

		return nil, 0, fmt.Errorf("LLM call failed after %d attempts: %w", maxRetries, err)
	}

	if len(response.Choices) == 0 {
		return nil, 0, fmt.Errorf("no response choices received")
	}

	choice := response.Choices[0]
	reasoning := strings.TrimSpace(choice.Message.Content)

	// 解析工具调用和代码块
	toolCalls := te.agent.parseToolCalls(&choice.Message)
	codeBlock := te.agent.parseCodeBlock(reasoning)

	plan := &ActionPlan{
		ToolCalls:  toolCalls,
		CodeBlock:  codeBlock,
		Reasoning:  reasoning,
		HasActions: len(toolCalls) > 0 || codeBlock != nil,
	}

	// 计算token使用量
	tokensUsed := response.Usage.TotalTokens
	if tokensUsed == 0 {
		tokensUsed = len(strings.Fields(reasoning)) + len(strings.Fields(planningPrompt))
	}

	return plan, tokensUsed, nil
}

// thinkWithResponse - 思考阶段实现（支持工具调用，返回完整响应）
func (te *ThinkingEngine) thinkWithResponse(ctx context.Context, prompt string, taskCtx *types.LightTaskContext) (*llm.ChatResponse, string, float64, int, error) {
	// 构建可用工具列表
	tools := te.buildToolDefinitions()

	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: prompt},
		},
		ModelType:  llm.BasicModel,
		Tools:      tools,
		ToolChoice: "auto", // 让模型自动决定是否调用工具
		Config:     te.agent.llmConfig,
	}

	// 获取LLM实例
	client, err := llm.GetLLMInstance(llm.BasicModel)
	if err != nil {
		return nil, "", 0.0, 0, fmt.Errorf("failed to get LLM instance: %w", err)
	}
	// 重试机制：最多尝试3次
	var response *llm.ChatResponse
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		response, err = client.Chat(ctx, request)
		if err == nil {
			break
		}

		// 如果不是最后一次尝试，等待一段时间后重试
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * time.Second
			time.Sleep(waitTime)
			continue
		}

		return nil, "", 0.0, 0, fmt.Errorf("LLM call failed after %d attempts: %w", maxRetries, err)
	}

	if len(response.Choices) == 0 {
		return nil, "", 0.0, 0, fmt.Errorf("no response choices received")
	}

	choice := response.Choices[0]
	thought := strings.TrimSpace(choice.Message.Content)
	confidence := te.calculateConfidence(thought, taskCtx)

	// 计算token使用量
	tokensUsed := response.Usage.TotalTokens
	if tokensUsed == 0 {
		tokensUsed = len(strings.Fields(thought)) + len(strings.Fields(prompt)) // 简化计算
	}

	return response, thought, confidence, tokensUsed, nil
}

// thinkWithConversation - 基于对话历史的思考阶段实现（支持工具调用，返回完整响应）
func (te *ThinkingEngine) thinkWithConversation(ctx context.Context, messages []llm.Message, taskCtx *types.LightTaskContext) (*llm.ChatResponse, string, float64, int, error) {
	// 构建可用工具列表
	tools := te.buildToolDefinitions()

	request := &llm.ChatRequest{
		Messages:   messages,
		ModelType:  llm.BasicModel,
		Tools:      tools,
		ToolChoice: "auto", // 让模型自动决定是否调用工具
		Config:     te.agent.llmConfig,
	}
	// 获取LLM实例
	client, err := llm.GetLLMInstance(llm.BasicModel)
	if err != nil {
		return nil, "", 0.0, 0, fmt.Errorf("failed to get LLM instance: %w", err)
	}

	// 重试机制：最多尝试3次
	var response *llm.ChatResponse
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		response, err = client.Chat(ctx, request)
		if err == nil {
			break
		}

		// 如果不是最后一次尝试，等待一段时间后重试
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * time.Second
			time.Sleep(waitTime)
			continue
		}

		return nil, "", 0.0, 0, fmt.Errorf("LLM call failed after %d attempts: %w", maxRetries, err)
	}

	if len(response.Choices) == 0 {
		return nil, "", 0.0, 0, fmt.Errorf("no response choices received")
	}

	choice := response.Choices[0]
	thought := strings.TrimSpace(choice.Message.Content)
	confidence := te.calculateConfidence(thought, taskCtx)

	// 计算token使用量
	tokensUsed := response.Usage.TotalTokens
	if tokensUsed == 0 {
		tokensUsed = len(strings.Fields(thought)) + len(strings.Fields(fmt.Sprintf("%v", messages))) // 简化计算
	}

	return response, thought, confidence, tokensUsed, nil
}

// buildToolDefinitions - 构建工具定义列表
func (te *ThinkingEngine) buildToolDefinitions() []llm.Tool {
	var tools []llm.Tool

	for _, tool := range te.agent.tools {
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

// 辅助方法
func (te *ThinkingEngine) canProvideDirectAnswer(thought string, confidence float64) bool {
	// 更严格的判断：只有明确包含具体答案的内容才认为可以直接提供答案
	if strings.TrimSpace(thought) == "" {
		return false
	}

	thoughtLower := strings.ToLower(thought)

	// 如果包含这些分析词汇，说明还在思考过程中，不是最终答案
	analysisIndicators := []string{
		"need to", "should", "let me", "first", "then", "next",
		"we need", "i need", "requires", "step", "approach",
		"break down", "can be", "operations", "process",
		"multi-step", "broken down", "analysis", "plan",
	}

	for _, indicator := range analysisIndicators {
		if strings.Contains(thoughtLower, indicator) {
			return false // 还在分析阶段，不是答案
		}
	}

	// 检查是否包含明确的最终答案标识
	answerIndicators := []string{
		"the answer is", "the result is", "the solution is",
		"current directory is", "directory is", "path is",
		"final answer:", "conclusion:", "answer:",
	}

	for _, indicator := range answerIndicators {
		if strings.Contains(thoughtLower, indicator) {
			return true
		}
	}

	// 检查是否包含具体的路径或数值结果（对简单查询有效）
	if strings.Contains(thought, "/Users/") ||
		strings.Contains(thought, "/home/") ||
		strings.Contains(thought, "C:\\") {
		// 确实包含路径信息，可能是答案
		return !strings.Contains(thoughtLower, "find") &&
			!strings.Contains(thoughtLower, "search") &&
			!strings.Contains(thoughtLower, "locate")
	}

	return false
}

func (te *ThinkingEngine) isTaskComplete(observation string, confidence float64) bool {
	// 业界标准：主要依赖观察结果内容判断，不过度依赖置信度
	if strings.TrimSpace(observation) == "" {
		return false
	}

	obsLower := strings.ToLower(observation)

	// 明确的完成标识
	completionIndicators := []string{
		"completed", "finished", "done", "solved", "successful",
		"answer:", "result:", "output:", "directory is",
		"current directory", "the answer", "here is",
		"task completed", "execution completed", "successfully executed",
	}

	for _, indicator := range completionIndicators {
		if strings.Contains(obsLower, indicator) {
			return true
		}
	}

	// 检查是否包含具体的有效结果（路径、文件内容等）
	if len(strings.TrimSpace(observation)) > 15 {
		// 路径结果
		if strings.Contains(observation, "/Users/") ||
			strings.Contains(observation, "/home/") ||
			strings.Contains(observation, "C:\\") ||
			(strings.Contains(observation, "/") && strings.Contains(observation, "code")) {
			return true
		}

		// 命令执行成功结果
		if strings.Contains(obsLower, "execution completed") ||
			strings.Contains(obsLower, "executed successfully") ||
			strings.Contains(obsLower, "tool execution successful") {
			return true
		}
	}

	return false
}

func (te *ThinkingEngine) calculateConfidence(thought string, taskCtx *types.LightTaskContext) float64 {
	confidence := 0.6 // 提高基础置信度

	if len(thought) > 50 {
		confidence += 0.1
	}

	if strings.Contains(strings.ToLower(thought), "because") ||
		strings.Contains(strings.ToLower(thought), "since") {
		confidence += 0.1
	}

	if strings.Contains(strings.ToLower(thought), "action:") {
		confidence += 0.1
	}

	// 如果最近的工具调用成功，大幅提升置信度
	if len(taskCtx.History) > 0 {
		lastStep := taskCtx.History[len(taskCtx.History)-1]
		if lastStep.Result != nil && lastStep.Result.Success {
			confidence += 0.3 // 提高成功工具调用的置信度奖励
		}
	}

	// 检查是否包含明确的答案或结果
	thoughtLower := strings.ToLower(thought)
	if strings.Contains(thoughtLower, "answer:") ||
		strings.Contains(thoughtLower, "result:") ||
		strings.Contains(thoughtLower, "solution:") ||
		strings.Contains(thoughtLower, "current directory") ||
		strings.Contains(thoughtLower, "the answer is") {
		confidence += 0.2
	}

	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}
