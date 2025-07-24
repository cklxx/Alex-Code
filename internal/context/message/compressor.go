package message

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"alex/internal/llm"
	"alex/internal/session"
)

// MessageCompressor handles message compression operations
type MessageCompressor struct {
	llmClient      llm.Client
	tokenEstimator *TokenEstimator
}

// NewMessageCompressor creates a new message compressor
func NewMessageCompressor(llmClient llm.Client) *MessageCompressor {
	return &MessageCompressor{
		llmClient:      llmClient,
		tokenEstimator: NewTokenEstimator(),
	}
}

// CompressMessages compresses messages using cache-friendly strategy
// Keeps stable prefix for context caching, compresses middle, preserves recent active
func (mc *MessageCompressor) CompressMessages(messages []*session.Message) []*session.Message {
	totalTokens := mc.estimateTokens(messages)
	messageCount := len(messages)

	log.Printf("[DEBUG] Token estimation: %d messages, %d estimated tokens", messageCount, totalTokens)

	// Cache-friendly compression thresholds
	const (
		TokenThreshold      = 115000 // Kimi K2的128K token上限的90%
		MessageThreshold    = 20     // 降低消息数量阈值，更早触发压缩
		CacheablePrefixKeep = 4      // 保留用于缓存的稳定前缀消息数
	)

	// Only compress if we exceed thresholds significantly
	if messageCount > MessageThreshold && totalTokens > TokenThreshold {
		log.Printf("[INFO] Cache-friendly compression triggered: %d messages, %d tokens", messageCount, totalTokens)
		return mc.cacheFriendlyCompress(messages, CacheablePrefixKeep)
	}

	log.Printf("[DEBUG] Compression skipped: %d messages (%d threshold), %d tokens (%d threshold)",
		messageCount, MessageThreshold, totalTokens, TokenThreshold)

	return messages
}

// cacheFriendlyCompress implements cache-friendly compression strategy
// Keeps stable prefix for context caching, compresses the rest
func (mc *MessageCompressor) cacheFriendlyCompress(messages []*session.Message, cacheablePrefixKeep int) []*session.Message {
	if len(messages) <= cacheablePrefixKeep {
		return messages // 消息不够多，不需要压缩
	}

	// Step 1: 分离系统消息和非系统消息
	var systemMessages []*session.Message
	var nonSystemMessages []*session.Message

	for _, msg := range messages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg)
		} else {
			nonSystemMessages = append(nonSystemMessages, msg)
		}
	}

	if len(nonSystemMessages) <= cacheablePrefixKeep {
		return messages // 非系统消息不够多，不需要压缩
	}

	// Step 2: 提取稳定的缓存前缀（考虑工具调用成对）
	cacheablePrefix := mc.findCacheablePrefixWithToolPairing(nonSystemMessages, cacheablePrefixKeep)

	// Step 3: 后续消息全部压缩
	cacheablePrefixEnd := len(cacheablePrefix)
	remainingMessages := nonSystemMessages[cacheablePrefixEnd:]

	log.Printf("[DEBUG] Cache-friendly compression: prefix=%d, remaining=%d",
		len(cacheablePrefix), len(remainingMessages))

	// Step 4: 压缩剩余消息
	compressedRemaining := mc.compressRemainingMessages(remainingMessages)

	// Step 5: 重新组合消息
	result := make([]*session.Message, 0, len(systemMessages)+len(cacheablePrefix)+1)

	// 添加系统消息
	result = append(result, systemMessages...)
	// 添加缓存前缀（稳定，用于 context caching）
	result = append(result, cacheablePrefix...)
	// 添加压缩的剩余部分
	if compressedRemaining != nil {
		result = append(result, compressedRemaining)
	}

	log.Printf("[INFO] Cache-friendly compression completed: %d -> %d messages", len(messages), len(result))
	return result
}

// findCacheablePrefixWithToolPairing finds cacheable prefix while maintaining tool call pairs
func (mc *MessageCompressor) findCacheablePrefixWithToolPairing(messages []*session.Message, targetKeep int) []*session.Message {
	if len(messages) <= targetKeep {
		return messages
	}

	// 构建工具调用配对映射（用于后续的工具调用成对验证）
	_ = mc.buildToolCallPairMap(messages)

	// 从前向后查找，确保工具调用成对
	kept := make([]*session.Message, 0, targetKeep*2)
	mustInclude := make(map[int]bool)

	// 标记必须包含的工具调用对
	for i := 0; i < min(targetKeep*2, len(messages)); i++ {
		msg := messages[i]

		if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			// 查找对应的工具结果
			for j := i + 1; j < len(messages); j++ {
				if messages[j].Role == "tool" {
					if toolCallId, ok := messages[j].Metadata["tool_call_id"].(string); ok {
						// 检查是否匹配当前assistant的某个tool_call
						for _, tc := range msg.ToolCalls {
							if tc.ID == toolCallId {
								mustInclude[i] = true
								mustInclude[j] = true
								break
							}
						}
					}
				}
			}
		}
	}

	// 收集前缀消息，确保工具调用成对
	for i := 0; i < len(messages) && len(kept) < targetKeep*2; i++ {
		if i < targetKeep || mustInclude[i] {
			kept = append(kept, messages[i])
		}
	}

	return kept
}

// compressRemainingMessages compresses remaining messages using AI summarization
func (mc *MessageCompressor) compressRemainingMessages(messages []*session.Message) *session.Message {
	if len(messages) == 0 {
		return nil
	}

	// 使用现有的 AI 摘要方法
	summaryMsg := mc.createComprehensiveAISummary(messages)
	if summaryMsg == nil {
		return nil
	}

	// 标记为缓存友好的压缩
	if summaryMsg.Metadata == nil {
		summaryMsg.Metadata = make(map[string]interface{})
	}
	summaryMsg.Metadata["cache_friendly_compression"] = true
	summaryMsg.Metadata["original_message_count"] = len(messages)

	log.Printf("[DEBUG] Compressed %d remaining messages into summary", len(messages))
	return summaryMsg
}

// buildToolCallPairMap builds a map of tool_call_id -> assistant message index
func (mc *MessageCompressor) buildToolCallPairMap(messages []*session.Message) map[string]int {
	pairs := make(map[string]int)

	for i, msg := range messages {
		if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				pairs[tc.ID] = i
			}
		}
	}

	return pairs
}

// createComprehensiveAISummary creates a comprehensive AI summary preserving important context
func (mc *MessageCompressor) createComprehensiveAISummary(messages []*session.Message) *session.Message {
	if mc.llmClient == nil || len(messages) == 0 {
		return mc.createStatisticalSummary(messages)
	}

	conversationText := mc.buildComprehensiveSummaryInput(messages)
	prompt := mc.buildComprehensiveSummaryPrompt(conversationText, len(messages))

	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: mc.buildComprehensiveSystemPrompt(),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ModelType: llm.BasicModel,
		Config: &llm.Config{
			Temperature: 0.2,  // Lower temperature for more consistent summaries
			MaxTokens:   1000, // More tokens for comprehensive summaries
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	response, err := mc.llmClient.Chat(ctx, request)
	if err != nil {
		log.Printf("[WARN] MessageCompressor: Comprehensive AI summary failed: %v", err)
		return mc.createStatisticalSummary(messages)
	}

	if len(response.Choices) == 0 {
		return mc.createStatisticalSummary(messages)
	}

	return &session.Message{
		Role:    "system",
		Content: fmt.Sprintf("Comprehensive conversation summary (%d messages): %s", len(messages), response.Choices[0].Message.Content),
		Metadata: map[string]any{
			"type":           "comprehensive_ai_summary",
			"original_count": len(messages),
			"created_at":     time.Now().Unix(),
			"summary_method": "ai_comprehensive",
		},
		Timestamp: time.Now(),
	}
}

// buildComprehensiveSystemPrompt builds the system prompt for comprehensive AI summarization
func (mc *MessageCompressor) buildComprehensiveSystemPrompt() string {
	return `You are an expert at creating comprehensive conversation summaries that preserve all important context and information. Your goal is to create detailed summaries that allow future interactions to continue seamlessly as if no compression occurred.

CRITICAL REQUIREMENTS:
- Preserve ALL important technical details, decisions, and context
- Maintain the narrative flow and logical connections between topics
- Include specific examples, code snippets, file names, and technical specifications mentioned
- Preserve user goals, preferences, and stated requirements
- Document any ongoing tasks, problems being solved, or work in progress
- Maintain the contextual relationships between different discussion topics
- Include any established patterns, conventions, or agreed-upon approaches

COMPREHENSIVE COVERAGE:
- Technical implementations and architectures discussed
- Problem-solving approaches and reasoning
- User feedback and iterative improvements
- File structures, code changes, and system modifications
- Testing approaches and validation methods
- Error handling and debugging processes
- Performance considerations and optimizations

The summary should be detailed enough that someone reading it can understand the full context and continue the conversation naturally.`
}

// buildComprehensiveSummaryPrompt builds the prompt for comprehensive AI summarization
func (mc *MessageCompressor) buildComprehensiveSummaryPrompt(conversationText string, messageCount int) string {
	return fmt.Sprintf(`Create a comprehensive summary of the following conversation (%d messages). This summary will replace the original messages in the conversation history, so it must preserve all important context and information needed for future interactions.

CONVERSATION TO SUMMARIZE:
%s

Create a detailed summary that covers:
1. Main topics and themes discussed
2. Technical details, implementations, and decisions made
3. User goals, requirements, and preferences
4. Problem-solving approaches and solutions implemented
5. Any ongoing work, tasks, or issues that need continuation
6. Code examples, file references, and system specifications mentioned
7. Key insights, learnings, and established patterns

COMPREHENSIVE SUMMARY:`, messageCount, conversationText)
}

// buildComprehensiveSummaryInput builds comprehensive input text for AI summarization
func (mc *MessageCompressor) buildComprehensiveSummaryInput(messages []*session.Message) string {
	var parts []string

	for i, msg := range messages {
		if msg.Role != "system" && len(strings.TrimSpace(msg.Content)) > 0 {
			// Include message index for context
			content := msg.Content

			// Include tool call information if present
			if len(msg.ToolCalls) > 0 {
				var toolInfo []string
				for _, tc := range msg.ToolCalls {
					toolInfo = append(toolInfo, fmt.Sprintf("Tool: %s", tc.Name))
				}
				content += fmt.Sprintf(" [Tool calls: %s]", strings.Join(toolInfo, ", "))
			}

			// Include tool response metadata if present
			if msg.Role == "tool" {
				if toolName, ok := msg.Metadata["tool_name"].(string); ok {
					content = fmt.Sprintf("[%s result]: %s", toolName, content)
				}
			}

			parts = append(parts, fmt.Sprintf("[Message %d - %s]: %s", i+1, msg.Role, content))
		}
	}

	text := strings.Join(parts, "\n\n")

	// Allow longer text for comprehensive summaries, but still limit to prevent token overflow
	if len(text) > 8000 {
		text = text[:8000] + "\n\n[... conversation continues with additional technical details ...]"
	}

	return text
}

// createStatisticalSummary creates a summary based on statistics
func (mc *MessageCompressor) createStatisticalSummary(messages []*session.Message) *session.Message {
	userCount := 0
	assistantCount := 0
	toolCount := 0

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			userCount++
		case "assistant":
			assistantCount++
		case "tool":
			toolCount++
		}
	}

	summary := fmt.Sprintf("Previous conversation summary: %d messages (%d user, %d assistant, %d tool)",
		len(messages), userCount, assistantCount, toolCount)

	return &session.Message{
		Role:    "system",
		Content: summary,
		Metadata: map[string]any{
			"type":            "statistical_summary",
			"original_count":  len(messages),
			"user_count":      userCount,
			"assistant_count": assistantCount,
			"tool_count":      toolCount,
			"created_at":      time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}
}

// estimateTokens estimates the total tokens in messages with improved accuracy
func (mc *MessageCompressor) estimateTokens(messages []*session.Message) int {
	total := 0
	for _, msg := range messages {
		// Use more accurate token estimation
		contentTokens := mc.estimateContentTokens(msg.Content)

		// Add overhead for role, metadata, and structure
		roleTokens := 5                          // role field
		metadataTokens := len(msg.Metadata) * 10 // rough metadata overhead

		// Add tokens for tool calls
		toolCallTokens := 0
		for _, tc := range msg.ToolCalls {
			toolCallTokens += len(tc.Name)/3 + len(tc.ID)/3 + 15 // Tool call structure overhead
		}

		messageTotal := contentTokens + roleTokens + metadataTokens + toolCallTokens
		total += messageTotal

		// Debug individual message token count for large messages
		if contentTokens > 1000 {
			log.Printf("[DEBUG] Large message: %d content tokens, %d total tokens (role: %d, metadata: %d, tools: %d)",
				contentTokens, messageTotal, roleTokens, metadataTokens, toolCallTokens)
		}
	}
	return total
}

// estimateContentTokens provides more accurate token estimation for message content
func (mc *MessageCompressor) estimateContentTokens(content string) int {
	if content == "" {
		return 0
	}

	// More sophisticated estimation based on content type
	length := len(content)

	// Different ratios for different content types
	var charsPerToken = 4.0 // Default for regular text

	// Adjust for code-heavy content (more token-dense)
	if strings.Contains(content, "```") || strings.Contains(content, "func ") || strings.Contains(content, "import ") {
		charsPerToken = 2.5 // Code is more token-dense
	}

	// Adjust for JSON/structured data
	if strings.Contains(content, `"role"`) || strings.Contains(content, `{"`) {
		charsPerToken = 3.0 // JSON is moderately token-dense
	}

	// Adjust for very long content (usually less token-dense due to repetition)
	if length > 10000 {
		charsPerToken = 5.0
	}

	estimated := int(float64(length) / charsPerToken)

	// Minimum of 1 token for non-empty content
	if estimated == 0 && length > 0 {
		estimated = 1
	}

	return estimated
}
