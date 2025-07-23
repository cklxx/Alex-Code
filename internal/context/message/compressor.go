package message

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"alex/internal/llm"
	"alex/internal/session"
	"alex/internal/utils"
)

// MessageCompressor handles message compression operations
type MessageCompressor struct {
	llmClient      llm.Client
	tokenEstimator *utils.TokenEstimator
}

// NewMessageCompressor creates a new message compressor
func NewMessageCompressor(llmClient llm.Client) *MessageCompressor {
	return &MessageCompressor{
		llmClient:      llmClient,
		tokenEstimator: utils.NewTokenEstimator(),
	}
}

// CompressMessages compresses a batch of messages using AI summarization instead of truncation
func (mc *MessageCompressor) CompressMessages(messages []*session.Message) []*session.Message {
	// Only compress when truly necessary (high thresholds)
	totalTokens := mc.estimateTokens(messages)
	messageCount := len(messages)
	
	// High thresholds aligned with Kimi K2's 128K token context window
	const (
		TokenThreshold = 100000 // 按Kimi K2的128K token上限设置，留20%余量
		MessageThreshold = 50   // Lower threshold to trigger compression earlier
		RecentKeep = 10         // Keep fewer recent messages to preserve more history via summaries
	)
	
	// Only compress if we exceed BOTH thresholds significantly
	if messageCount > MessageThreshold && totalTokens > TokenThreshold {
		log.Printf("[INFO] AI-based compression triggered: %d messages, %d tokens", messageCount, totalTokens)
		return mc.aiBasedCompress(messages, RecentKeep)
	}
	
	// No compression needed
	return messages
}

// findRecentMessagesWithToolPairing finds recent messages while maintaining tool call pairs
func (mc *MessageCompressor) findRecentMessagesWithToolPairing(messages []*session.Message, recentKeep int) []*session.Message {
	if len(messages) <= recentKeep {
		return messages
	}

	// Start from the most recent messages and work backwards
	// But ensure we keep complete tool call sequences
	
	// First, identify all tool call pairs
	toolCallPairs := mc.buildToolCallPairMap(messages)
	
	// Start from the end, keep adding messages while maintaining pairs
	kept := make([]*session.Message, 0, recentKeep*2) // Allow for expansion due to tool pairs
	messageIndices := make(map[*session.Message]int)
	
	// Build index map
	for i, msg := range messages {
		messageIndices[msg] = i
	}
	
	// Track which messages we must include to maintain pairing
	mustInclude := make(map[int]bool)
	
	// Add recent messages from the end
	for i := len(messages) - 1; i >= 0 && len(kept) < recentKeep*2; i-- {
		msg := messages[i]
		
		// If this message is part of a tool call pair, include the whole pair
		if msg.Role == "tool" {
			if toolCallId, ok := msg.Metadata["tool_call_id"].(string); ok {
				if assistantIndex, exists := toolCallPairs[toolCallId]; exists {
					mustInclude[assistantIndex] = true
					mustInclude[i] = true
				}
			}
		} else if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			mustInclude[i] = true
			// Find all corresponding tool responses
			for _, tc := range msg.ToolCalls {
				for j := i + 1; j < len(messages); j++ {
					if messages[j].Role == "tool" {
						if callId, ok := messages[j].Metadata["tool_call_id"].(string); ok && callId == tc.ID {
							mustInclude[j] = true
						}
					}
				}
			}
		} else {
			mustInclude[i] = true
		}
		
		// Stop if we have enough messages (but allow pairs to complete)
		if len(mustInclude) >= recentKeep {
			break
		}
	}
	
	// Find the earliest index we need to include
	minIndex := len(messages)
	for index := range mustInclude {
		if index < minIndex {
			minIndex = index
		}
	}
	
	// Return messages from minIndex to end
	if minIndex < len(messages) {
		return messages[minIndex:]
	}
	
	// Fallback: just return the last recentKeep messages
	if len(messages) > recentKeep {
		return messages[len(messages)-recentKeep:]
	}
	return messages
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

// aiBasedCompress applies AI-based compression using intelligent summarization instead of truncation
func (mc *MessageCompressor) aiBasedCompress(messages []*session.Message, recentKeep int) []*session.Message {
	if len(messages) <= recentKeep {
		return messages
	}

	// Separate system messages (always keep)
	var systemMessages []*session.Message
	var nonSystemMessages []*session.Message
	
	for _, msg := range messages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg)
		} else {
			nonSystemMessages = append(nonSystemMessages, msg)
		}
	}
	
	// If not enough non-system messages, no compression needed
	if len(nonSystemMessages) <= recentKeep {
		return messages
	}
	
	// Find proper split point while maintaining tool call pairs
	recentMessages := mc.findRecentMessagesWithToolPairing(nonSystemMessages, recentKeep)
	recentStart := len(nonSystemMessages) - len(recentMessages)
	
	// Messages to compress (older messages)
	toCompress := nonSystemMessages[:recentStart]
	
	// Build result: system messages + comprehensive summary + recent messages
	var result []*session.Message
	result = append(result, systemMessages...)
	
	// Create comprehensive AI summary of all older messages
	if len(toCompress) > 0 {
		summary := mc.createComprehensiveAISummary(toCompress)
		if summary != nil {
			result = append(result, summary)
		}
	}
	
	// Add recent messages (with tool call pairs intact)
	result = append(result, recentMessages...)
	
	log.Printf("[INFO] AI-based compression: %d -> %d messages", len(messages), len(result))
	return result
}

// comprehensiveCompress applies comprehensive compression - simplified and robust
func (mc *MessageCompressor) comprehensiveCompress(messages []*session.Message, recentKeep int) []*session.Message {
	if len(messages) <= recentKeep {
		return messages
	}

	// Separate system messages (always keep)
	var systemMessages []*session.Message
	var nonSystemMessages []*session.Message
	
	for _, msg := range messages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg)
		} else {
			nonSystemMessages = append(nonSystemMessages, msg)
		}
	}
	
	// If not enough non-system messages, no compression needed
	if len(nonSystemMessages) <= recentKeep {
		return messages
	}
	
	// Find proper split point while maintaining tool call pairs
	recentMessages := mc.findRecentMessagesWithToolPairing(nonSystemMessages, recentKeep)
	recentStart := len(nonSystemMessages) - len(recentMessages)
	
	// Messages to compress (older messages)
	toCompress := nonSystemMessages[:recentStart]
	
	// Build result: system messages + summary + recent messages
	var result []*session.Message
	result = append(result, systemMessages...)
	
	// Create summary if there are messages to compress
	if len(toCompress) > 0 {
		summary := mc.createLLMSummary(toCompress)
		if summary != nil {
			result = append(result, summary)
		}
	}
	
	// Add recent messages (with tool call pairs intact)
	result = append(result, recentMessages...)
	
	log.Printf("[INFO] Comprehensive compression: %d -> %d messages", len(messages), len(result))
	return result
}



// Unused memory-related functions removed

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
			Temperature: 0.2, // Lower temperature for more consistent summaries
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
		Metadata: map[string]interface{}{
			"type":           "comprehensive_ai_summary",
			"original_count": len(messages),
			"created_at":     time.Now().Unix(),
			"summary_method": "ai_comprehensive",
		},
		Timestamp: time.Now(),
	}
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

// buildOptimizedSummaryPrompt builds the prompt for summarization
func (mc *MessageCompressor) buildOptimizedSummaryPrompt(conversationText string, messageCount int) string {
	return fmt.Sprintf(`Summarize the following conversation (%d messages) in 2-3 sentences. Focus on key decisions, technical details, and important context:

%s

Summary:`, messageCount, conversationText)
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
		total += mc.tokenEstimator.EstimateText(msg.Content)
	}
	return total
}