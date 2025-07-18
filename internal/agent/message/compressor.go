package message

import (
	"context"
	"fmt"
	"log"
	"sort"
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

// CompressMessages compresses a batch of messages
func (mc *MessageCompressor) CompressMessages(messages []*session.Message) []*session.Message {
	if len(messages) <= 10 {
		return messages
	}

	// Estimate total tokens
	totalTokens := mc.estimateTokens(messages)
	
	// Choose compression strategy based on message count and token count
	if totalTokens > 8000 {
		return mc.aggressiveCompress(messages, 5)
	} else if totalTokens > 6000 {
		return mc.moderateCompress(messages, 8)
	} else if len(messages) > 20 {
		return mc.lightCompress(messages, 15)
	}
	
	return messages
}

// findToolAwareSplitPoint finds the split point that respects tool call pairs
func (mc *MessageCompressor) findToolAwareSplitPoint(messages []*session.Message, recentKeep int) int {
	if len(messages) <= recentKeep {
		return 0
	}

	// 从后往前扫描，确保工具调用和响应成对保留
	keptCount := 0
	splitPoint := len(messages)

	for i := len(messages) - 1; i >= 0 && keptCount < recentKeep; i-- {
		msg := messages[i]

		// 如果是工具响应消息，需要确保对应的工具调用也被保留
		if msg.Role == "tool" {
			if toolCallId, ok := msg.Metadata["tool_call_id"].(string); ok && toolCallId != "" {
				// 向前查找对应的工具调用
				foundToolCall := false
				for j := i - 1; j >= 0; j-- {
					if messages[j].Role == "assistant" && len(messages[j].ToolCalls) > 0 {
						// 检查是否包含匹配的工具调用ID
						for _, tc := range messages[j].ToolCalls {
							if tc.ID == toolCallId {
								foundToolCall = true
								break
							}
						}
						if foundToolCall {
							break
						}
					}
				}
				
				// 如果找到了对应的工具调用，并且它在切分点之前，需要调整切分点
				if foundToolCall {
					// 继续向前查找，确保包含完整的工具调用序列
					for j := i - 1; j >= 0; j-- {
						if messages[j].Role == "assistant" && len(messages[j].ToolCalls) > 0 {
							// 检查这个助手消息是否包含当前工具响应的调用
							hasMatchingCall := false
							for _, tc := range messages[j].ToolCalls {
								if tc.ID == toolCallId {
									hasMatchingCall = true
									break
								}
							}
							if hasMatchingCall {
								splitPoint = j
								keptCount = len(messages) - j
								break
							}
						}
					}
				}
			}
		}

		// 如果是助手消息且包含工具调用，需要确保所有对应的工具响应都被保留
		if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			allResponsesIncluded := true
			maxResponseIndex := i

			// 检查所有工具调用是否都有对应的响应在保留范围内
			for _, tc := range msg.ToolCalls {
				responseFound := false
				for j := i + 1; j < len(messages); j++ {
					if messages[j].Role == "tool" {
						if callId, ok := messages[j].Metadata["tool_call_id"].(string); ok && callId == tc.ID {
							responseFound = true
							if j > maxResponseIndex {
								maxResponseIndex = j
							}
							break
						}
					}
				}
				if !responseFound {
					allResponsesIncluded = false
					break
				}
			}

			// 如果所有响应都在范围内，调整切分点以包含完整序列
			if allResponsesIncluded {
				splitPoint = i
				keptCount = len(messages) - i
			}
		}

		// 简单情况：如果还没有达到保留数量限制，继续向前
		if keptCount < recentKeep {
			splitPoint = i
			keptCount++
		}
	}

	return splitPoint
}

// aggressiveCompress applies aggressive compression
func (mc *MessageCompressor) aggressiveCompress(messages []*session.Message, recentKeep int) []*session.Message {
	if len(messages) <= recentKeep {
		return messages
	}

	// Keep recent messages and system messages
	var result []*session.Message
	var toCompress []*session.Message
	
	// Add system messages
	for _, msg := range messages {
		if msg.Role == "system" {
			result = append(result, msg)
		}
	}
	
	// Find tool-aware split point
	recentStart := mc.findToolAwareSplitPoint(messages, recentKeep)
	
	// Messages to compress
	for i := 0; i < recentStart; i++ {
		msg := messages[i]
		if msg.Role != "system" {
			toCompress = append(toCompress, msg)
		}
	}
	
	// Create summary if there are messages to compress
	if len(toCompress) > 0 {
		summary := mc.createLLMSummary(toCompress)
		if summary != nil {
			result = append(result, summary)
		}
	}
	
	// Add recent messages
	for i := recentStart; i < len(messages); i++ {
		msg := messages[i]
		if msg.Role != "system" {
			result = append(result, msg)
		}
	}
	
	return result
}

// moderateCompress applies moderate compression
func (mc *MessageCompressor) moderateCompress(messages []*session.Message, recentKeep int) []*session.Message {
	if len(messages) <= recentKeep {
		return messages
	}

	// Keep important messages and recent messages
	var result []*session.Message
	var toCompress []*session.Message
	
	// Add system messages
	for _, msg := range messages {
		if msg.Role == "system" {
			result = append(result, msg)
		}
	}
	
	// Select important messages to keep
	importantMessages := mc.selectImportantMessages(messages, recentKeep/2)
	
	// Find tool-aware split point
	recentStart := mc.findToolAwareSplitPoint(messages, recentKeep)
	
	// Messages to compress (exclude important and recent)
	importantSet := make(map[*session.Message]bool)
	for _, msg := range importantMessages {
		importantSet[msg] = true
	}
	
	for i := 0; i < recentStart; i++ {
		msg := messages[i]
		if msg.Role != "system" && !importantSet[msg] {
			toCompress = append(toCompress, msg)
		}
	}
	
	// Create summary if there are messages to compress
	if len(toCompress) > 0 {
		summary := mc.createLLMSummary(toCompress)
		if summary != nil {
			result = append(result, summary)
		}
	}
	
	// Add important messages
	result = append(result, importantMessages...)
	
	// Add recent messages
	for i := recentStart; i < len(messages); i++ {
		msg := messages[i]
		if msg.Role != "system" {
			result = append(result, msg)
		}
	}
	
	return result
}

// lightCompress applies light compression
func (mc *MessageCompressor) lightCompress(messages []*session.Message, recentKeep int) []*session.Message {
	if len(messages) <= recentKeep {
		return messages
	}

	// Remove low-value messages
	var result []*session.Message
	
	// Keep all system messages
	for _, msg := range messages {
		if msg.Role == "system" {
			result = append(result, msg)
		}
	}
	
	// Find tool-aware split point
	recentStart := mc.findToolAwareSplitPoint(messages, recentKeep)
	
	// Filter out low-value messages from older messages
	for i := 0; i < recentStart; i++ {
		msg := messages[i]
		if msg.Role != "system" && !mc.isLowValueMessage(msg) {
			result = append(result, msg)
		}
	}
	
	// Add recent messages
	for i := recentStart; i < len(messages); i++ {
		msg := messages[i]
		if msg.Role != "system" {
			result = append(result, msg)
		}
	}
	
	return result
}

// selectImportantMessages selects the most important messages
func (mc *MessageCompressor) selectImportantMessages(messages []*session.Message, maxCount int) []*session.Message {
	// Calculate importance scores
	type messageScore struct {
		message *session.Message
		score   float64
	}
	
	var scores []messageScore
	for _, msg := range messages {
		if msg.Role != "system" {
			score := mc.calculateMessageImportance(msg)
			scores = append(scores, messageScore{message: msg, score: score})
		}
	}
	
	// Sort by importance (descending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// Return top messages
	var result []*session.Message
	for i := 0; i < len(scores) && i < maxCount; i++ {
		result = append(result, scores[i].message)
	}
	
	return result
}

// isLowValueMessage checks if a message has low value
func (mc *MessageCompressor) isLowValueMessage(msg *session.Message) bool {
	content := strings.ToLower(strings.TrimSpace(msg.Content))
	
	// Check for low-value patterns
	lowValuePatterns := []string{
		"ok", "yes", "no", "done", "thanks", "thank you",
		"got it", "understood", "sure", "alright", "okay",
	}
	
	for _, pattern := range lowValuePatterns {
		if content == pattern {
			return true
		}
	}
	
	// Check for very short messages
	if len(content) < 10 {
		return true
	}
	
	return false
}

// calculateMessageImportance calculates the importance score of a message
func (mc *MessageCompressor) calculateMessageImportance(msg *session.Message) float64 {
	score := 0.0
	content := strings.ToLower(msg.Content)
	
	// Length factor
	score += float64(len(msg.Content)) * 0.01
	
	// Role factor
	switch msg.Role {
	case "user":
		score += 10.0
	case "assistant":
		score += 8.0
	case "tool":
		score += 6.0
	}
	
	// Content importance keywords
	importantKeywords := []string{
		"error", "problem", "issue", "bug", "fix", "solution",
		"important", "critical", "urgent", "help", "need",
		"implement", "create", "develop", "build", "design",
	}
	
	for _, keyword := range importantKeywords {
		if strings.Contains(content, keyword) {
			score += 5.0
		}
	}
	
	// Tool calls add importance
	if len(msg.ToolCalls) > 0 {
		score += float64(len(msg.ToolCalls)) * 3.0
	}
	
	// Recency factor (more recent = more important)
	age := time.Since(msg.Timestamp).Hours()
	if age < 1 {
		score += 10.0
	} else if age < 24 {
		score += 5.0
	}
	
	return score
}

// createLLMSummary creates a summary using LLM
func (mc *MessageCompressor) createLLMSummary(messages []*session.Message) *session.Message {
	if mc.llmClient == nil || len(messages) == 0 {
		return mc.createStatisticalSummary(messages)
	}
	
	conversationText := mc.buildSummaryInput(messages)
	prompt := mc.buildOptimizedSummaryPrompt(conversationText, len(messages))
	
	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: mc.buildOptimizedSystemPrompt(),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ModelType: llm.BasicModel,
		Config: &llm.Config{
			Temperature: 0.3,
			MaxTokens:   500,
		},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	response, err := mc.llmClient.Chat(ctx, request)
	if err != nil {
		log.Printf("[WARN] MessageCompressor: LLM summary failed: %v", err)
		return mc.createStatisticalSummary(messages)
	}
	
	if len(response.Choices) == 0 {
		return mc.createStatisticalSummary(messages)
	}
	
	return &session.Message{
		Role:    "system",
		Content: fmt.Sprintf("Previous conversation summary (%d messages): %s", len(messages), response.Choices[0].Message.Content),
		Metadata: map[string]interface{}{
			"type":         "llm_summary",
			"original_count": len(messages),
			"created_at":   time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}
}

// buildOptimizedSystemPrompt builds the system prompt for summarization
func (mc *MessageCompressor) buildOptimizedSystemPrompt() string {
	return `You are an expert at summarizing conversations. Create concise, informative summaries that preserve key information, decisions, and context while being much shorter than the original.

Focus on:
- Key decisions and outcomes
- Important technical details
- User intentions and goals
- Problem-solving steps
- Context that affects future interactions

Avoid:
- Repetitive information
- Greeting/closing pleasantries
- Verbose explanations
- Unnecessary details`
}

// buildOptimizedSummaryPrompt builds the prompt for summarization
func (mc *MessageCompressor) buildOptimizedSummaryPrompt(conversationText string, messageCount int) string {
	return fmt.Sprintf(`Summarize the following conversation (%d messages) in 2-3 sentences. Focus on key decisions, technical details, and important context:

%s

Summary:`, messageCount, conversationText)
}

// buildSummaryInput builds the input text for summarization
func (mc *MessageCompressor) buildSummaryInput(messages []*session.Message) string {
	var parts []string
	
	for _, msg := range messages {
		if msg.Role != "system" && len(strings.TrimSpace(msg.Content)) > 0 {
			parts = append(parts, fmt.Sprintf("[%s]: %s", msg.Role, msg.Content))
		}
	}
	
	text := strings.Join(parts, "\n")
	
	// Truncate if too long
	if len(text) > 4000 {
		text = text[:4000] + "..."
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
		Metadata: map[string]interface{}{
			"type":           "statistical_summary",
			"original_count": len(messages),
			"user_count":     userCount,
			"assistant_count": assistantCount,
			"tool_count":     toolCount,
			"created_at":     time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}
}

// estimateTokens estimates the total tokens in messages
func (mc *MessageCompressor) estimateTokens(messages []*session.Message) int {
	total := 0
	for _, msg := range messages {
		total += mc.tokenEstimator.EstimateTokens(msg.Content)
	}
	return total
}