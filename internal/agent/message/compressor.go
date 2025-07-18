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
	
	// Keep recent messages
	recentStart := len(messages) - recentKeep
	if recentStart < 0 {
		recentStart = 0
	}
	
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
	
	// Keep recent messages
	recentStart := len(messages) - recentKeep
	if recentStart < 0 {
		recentStart = 0
	}
	
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
	
	// Keep recent messages
	recentStart := len(messages) - recentKeep
	if recentStart < 0 {
		recentStart = 0
	}
	
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