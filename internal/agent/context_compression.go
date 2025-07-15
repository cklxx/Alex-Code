package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"alex/internal/llm"
	"alex/internal/session"
)

// ContextCompression 负责上下文压缩相关功能
type ContextCompression struct {
	contextHandler *ContextHandler
	tokenEstimator *TokenEstimator
}

// NewContextCompression 创建上下文压缩器
func NewContextCompression(contextHandler *ContextHandler) *ContextCompression {
	return &ContextCompression{
		contextHandler: contextHandler,
		tokenEstimator: NewTokenEstimator(),
	}
}

// intelligentContextCompression - 智能上下文压缩
func (cc *ContextCompression) intelligentContextCompression(sessionMessages []*session.Message) []*session.Message {
	// 配置参数
	const (
		MaxMessages = 20    // 最大消息数量
		MaxTokens   = 60000 // 估算最大token数
		RecentKeep  = 6     // 最近保留的消息数量
	)

	// 如果消息数量不超过限制，直接返回
	if len(sessionMessages) <= MaxMessages {
		estimatedTokens := cc.tokenEstimator.EstimateSessionMessages(sessionMessages)
		if estimatedTokens <= MaxTokens {
			return sessionMessages
		}
	}

	log.Printf("[INFO] Context compression triggered: %d messages, estimated %d tokens",
		len(sessionMessages), cc.tokenEstimator.EstimateSessionMessages(sessionMessages))

	// 分离不同类型的消息
	var (
		recentMessages    []*session.Message // 最近的消息（保留）
		importantMessages []*session.Message // 重要的消息（保留）
		regularMessages   []*session.Message // 普通消息（可压缩）
	)

	// 保留最近的消息
	recentStart := len(sessionMessages) - RecentKeep
	if recentStart < 0 {
		recentStart = 0
	}
	recentMessages = sessionMessages[recentStart:]

	// 分析之前的消息
	for i := 0; i < recentStart; i++ {
		msg := sessionMessages[i]
		if cc.isImportantMessage(msg) {
			importantMessages = append(importantMessages, msg)
		} else {
			regularMessages = append(regularMessages, msg)
		}
	}

	// 构建压缩后的消息列表
	var compressedMessages []*session.Message

	// 添加重要消息
	compressedMessages = append(compressedMessages, importantMessages...)

	// 如果普通消息太多，进行进一步压缩
	if len(regularMessages) > 10 {
		// 创建压缩摘要
		summaryMsg := cc.createCompressionSummary(regularMessages)
		if summaryMsg != nil {
			compressedMessages = append(compressedMessages, summaryMsg)
		}

		// 只保留最后几条普通消息
		keepCount := 3
		if len(regularMessages) > keepCount {
			compressedMessages = append(compressedMessages, regularMessages[len(regularMessages)-keepCount:]...)
		} else {
			compressedMessages = append(compressedMessages, regularMessages...)
		}
	} else {
		// 普通消息不多，全部保留
		compressedMessages = append(compressedMessages, regularMessages...)
	}

	// 添加最近的消息
	compressedMessages = append(compressedMessages, recentMessages...)

	log.Printf("[INFO] Context compressed: %d -> %d messages",
		len(sessionMessages), len(compressedMessages))

	return compressedMessages
}

// isImportantMessage - 判断消息是否重要
func (cc *ContextCompression) isImportantMessage(msg *session.Message) bool {
	// Memory消息非常重要，永远保留
	if msgType, ok := msg.Metadata["type"].(string); ok {
		if msgType == "memory_context" || strings.Contains(msgType, "memory") {
			return true
		}
	}

	// Memory相关内容重要
	if strings.Contains(msg.Content, "## Relevant Context from Memory") ||
		strings.Contains(msg.Content, "### CodeContext") ||
		strings.Contains(msg.Content, "### TaskHistory") ||
		strings.Contains(msg.Content, "### Solutions") {
		return true
	}

	// 工具调用消息重要
	if len(msg.ToolCalls) > 0 {
		return true
	}

	// 包含错误信息的消息重要
	content := strings.ToLower(msg.Content)
	errorKeywords := []string{"error", "failed", "exception", "panic", "bug", "issue"}
	for _, keyword := range errorKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}

	// 长消息可能重要
	if len(msg.Content) > 200 {
		return true
	}

	// 包含代码块的消息重要
	if strings.Contains(msg.Content, "```") {
		return true
	}

	// 压缩摘要消息重要
	if strings.Contains(msg.Content, "Conversation Summary") ||
		strings.Contains(msg.Content, "## Summary") {
		return true
	}

	return false
}

// createCompressionSummary - 使用大模型创建智能压缩摘要
func (cc *ContextCompression) createCompressionSummary(messages []*session.Message) *session.Message {
	if len(messages) == 0 {
		return nil
	}

	// 首先尝试使用LLM进行智能压缩
	if llmSummary := cc.createLLMCompressionSummary(messages); llmSummary != nil {
		return llmSummary
	}

	// LLM失败时回退到统计性摘要
	log.Printf("[WARN] LLM compression failed, using fallback statistical summary")
	return cc.createStatisticalSummary(messages)
}

// createLLMCompressionSummary - 使用LLM进行智能压缩
func (cc *ContextCompression) createLLMCompressionSummary(messages []*session.Message) *session.Message {
	// 获取LLM客户端
	llmClient, err := llm.GetLLMInstance(llm.BasicModel)
	if err != nil {
		log.Printf("[WARN] Failed to get LLM instance for compression: %v", err)
		return nil
	}

	// 构建压缩输入文本
	conversationText := cc.buildCompressionInput(messages)
	if len(conversationText) == 0 {
		return nil
	}

	// 构建压缩prompt
	compressionPrompt := cc.buildCompressionPrompt(conversationText, len(messages))

	// 调用LLM进行压缩
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: "You are an expert at summarizing conversations. Create concise, informative summaries that preserve key information while reducing length.",
			},
			{
				Role:    "user",
				Content: compressionPrompt,
			},
		},
		ModelType: llm.BasicModel,
		Config: &llm.Config{
			Temperature: 0.8, // 确保输出稳定
			MaxTokens:   10000,
		},
	}

	response, err := llmClient.Chat(ctx, request)
	if err != nil {
		log.Printf("[WARN] LLM compression call failed: %v", err)
		return nil
	}

	if len(response.Choices) == 0 {
		log.Printf("[WARN] LLM compression returned no choices")
		return nil
	}

	summaryContent := strings.TrimSpace(response.Choices[0].Message.Content)
	if len(summaryContent) == 0 {
		return nil
	}

	// 获取token使用信息
	usage := response.GetUsage()
	tokensUsed := usage.GetTotalTokens()

	log.Printf("[INFO] LLM compression successful: %d messages -> %d tokens", len(messages), tokensUsed)

	return &session.Message{
		Role:    "system",
		Content: summaryContent,
		Metadata: map[string]interface{}{
			"type":               "llm_compression_summary",
			"original_count":     len(messages),
			"summary_timestamp":  time.Now().Unix(),
			"compression_method": "llm",
			"tokens_used":        tokensUsed,
			"prompt_tokens":      usage.GetPromptTokens(),
			"completion_tokens":  usage.GetCompletionTokens(),
		},
		Timestamp: time.Now(),
	}
}

// buildCompressionInput - 构建压缩输入文本
func (cc *ContextCompression) buildCompressionInput(messages []*session.Message) string {
	var inputParts []string

	for i, msg := range messages {
		// 跳过系统消息和已压缩的摘要
		if msg.Role == "system" {
			if msgType, ok := msg.Metadata["type"].(string); ok {
				if strings.Contains(msgType, "summary") || strings.Contains(msgType, "compression") {
					continue
				}
			}
		}

		// 限制单条消息长度，避免输入过长
		content := msg.Content
		if len(content) > 500 {
			content = content[:500] + "...[truncated]"
		}

		// 格式化消息
		var roleName string
		switch msg.Role {
		case "user":
			roleName = "User"
		case "assistant":
			roleName = "Assistant"
		case "tool":
			roleName = "Tool"
			if toolName, ok := msg.Metadata["tool_name"].(string); ok {
				roleName = fmt.Sprintf("Tool(%s)", toolName)
			}
		default:
			roleName = strings.ToUpper(msg.Role[:1]) + msg.Role[1:]
		}

		// 添加工具调用信息
		if len(msg.ToolCalls) > 0 {
			var tools []string
			for _, tc := range msg.ToolCalls {
				tools = append(tools, tc.Name)
			}
			content += fmt.Sprintf(" [Tools: %s]", strings.Join(tools, ", "))
		}

		inputParts = append(inputParts, fmt.Sprintf("[%d] %s: %s", i+1, roleName, content))
	}

	return strings.Join(inputParts, "\n")
}

// buildCompressionPrompt - 构建压缩提示
func (cc *ContextCompression) buildCompressionPrompt(conversationText string, messageCount int) string {
	return fmt.Sprintf(`Please create a concise and informative summary of the following conversation (%d messages).

Requirements:
1. Extract key decisions, actions, and outcomes
2. Preserve important technical details and context
3. Highlight successful tool usage and any failures
4. Maintain chronological flow of important events
5. Keep the summary under 400 words
6. Use structured format with bullet points or sections

Conversation to summarize:
%s

Please provide a comprehensive summary that maintains the essential context while significantly reducing the length:`, messageCount, conversationText)
}

// createStatisticalSummary - 创建统计性摘要（原有逻辑作为回退）
func (cc *ContextCompression) createStatisticalSummary(messages []*session.Message) *session.Message {
	// 简单的摘要策略：统计工具使用和主要活动
	var (
		userActions []string
		toolUsages  []string
		keyTopics   []string
	)

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			// 提取用户请求的关键词
			content := msg.Content
			if len(content) > 50 {
				content = content[:50] + "..."
			}
			userActions = append(userActions, content)

		case "assistant":
			// 提取工具调用
			for _, tc := range msg.ToolCalls {
				toolUsages = append(toolUsages, tc.Name)
			}

		case "tool":
			// 工具结果中提取关键信息
			if toolName, ok := msg.Metadata["tool_name"].(string); ok {
				success := "✓"
				if toolSuccess, ok := msg.Metadata["tool_success"].(bool); ok && !toolSuccess {
					success = "✗"
				}
				keyTopics = append(keyTopics, fmt.Sprintf("%s%s", success, toolName))
			}
		}
	}

	// 构建摘要内容
	var summaryParts []string
	summaryParts = append(summaryParts, fmt.Sprintf("## Conversation Summary (%d messages compressed)", len(messages)))

	if len(userActions) > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("**User Requests**: %s", strings.Join(userActions, "; ")))
	}

	if len(toolUsages) > 0 {
		// 统计工具使用频率
		toolCount := make(map[string]int)
		for _, tool := range toolUsages {
			toolCount[tool]++
		}
		var toolSummary []string
		for tool, count := range toolCount {
			if count > 1 {
				toolSummary = append(toolSummary, fmt.Sprintf("%s(%d)", tool, count))
			} else {
				toolSummary = append(toolSummary, tool)
			}
		}
		summaryParts = append(summaryParts, fmt.Sprintf("**Tools Used**: %s", strings.Join(toolSummary, ", ")))
	}

	if len(keyTopics) > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("**Key Activities**: %s", strings.Join(keyTopics, ", ")))
	}

	return &session.Message{
		Role:    "system",
		Content: strings.Join(summaryParts, "\n"),
		Metadata: map[string]interface{}{
			"type":               "statistical_summary",
			"original_count":     len(messages),
			"summary_timestamp":  time.Now().Unix(),
			"compression_method": "statistical",
		},
		Timestamp: time.Now(),
	}
}
