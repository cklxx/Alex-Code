package context

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"alex/internal/llm"
	"alex/internal/session"
)

// ContextProcessor provides unified context processing and compression
type ContextProcessor struct {
	tokenEstimator   *TokenEstimator
	messageAnalyzer  *MessageAnalyzer
	compressionEngine *CompressionEngine
	config           *ContextProcessorConfig
}

// ContextProcessorConfig defines configuration for context processing
type ContextProcessorConfig struct {
	MaxTokens              int     `json:"max_tokens"`
	MaxMessages            int     `json:"max_messages"`
	RecentKeepCount        int     `json:"recent_keep_count"`
	SummarizationThreshold int     `json:"summarization_threshold"`
	CompressionRatio       float64 `json:"compression_ratio"`
	PreserveSystemMessages bool    `json:"preserve_system_messages"`
}

// TokenEstimator provides unified token estimation
type TokenEstimator struct {
	charsPerToken int
	overhead      int
}

// MessageAnalyzer analyzes message importance and content
type MessageAnalyzer struct{}

// CompressionEngine handles message compression and summarization
type CompressionEngine struct {
	llmClient llm.Client
}

// NewContextProcessor creates a new unified context processor
func NewContextProcessor(llmClient llm.Client, config *ContextProcessorConfig) *ContextProcessor {
	if config == nil {
		config = &ContextProcessorConfig{
			MaxTokens:              60000,
			MaxMessages:            20,
			RecentKeepCount:        6,
			SummarizationThreshold: 45000, // 75% of max tokens
			CompressionRatio:       0.3,
			PreserveSystemMessages: true,
		}
	}

	return &ContextProcessor{
		tokenEstimator:   &TokenEstimator{charsPerToken: 3, overhead: 100},
		messageAnalyzer:  &MessageAnalyzer{},
		compressionEngine: &CompressionEngine{llmClient: llmClient},
		config:           config,
	}
}

// ProcessContext processes session context with intelligent compression
func (cp *ContextProcessor) ProcessContext(ctx context.Context, sessionMessages []*session.Message) (*ContextProcessingResult, error) {
	// Check if compression is needed
	analysis := cp.AnalyzeContext(sessionMessages)
	
	if !analysis.RequiresTrimming {
		return &ContextProcessingResult{
			Action:         "no_action",
			OriginalCount:  len(sessionMessages),
			ProcessedCount: len(sessionMessages),
		}, nil
	}

	// Apply intelligent compression
	processedMessages, err := cp.ApplyCompression(ctx, sessionMessages)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	return &ContextProcessingResult{
		Action:         "compressed",
		OriginalCount:  len(sessionMessages),
		ProcessedCount: len(processedMessages),
	}, nil
}

// AnalyzeContext analyzes context requirements
func (cp *ContextProcessor) AnalyzeContext(messages []*session.Message) *ContextAnalysis {
	estimatedTokens := cp.tokenEstimator.EstimateSessionMessages(messages)
	
	return &ContextAnalysis{
		TotalMessages:     len(messages),
		EstimatedTokens:   estimatedTokens,
		RequiresTrimming:  estimatedTokens > cp.config.MaxTokens || len(messages) > cp.config.MaxMessages,
		ShouldSummarize:   estimatedTokens > cp.config.SummarizationThreshold,
		CompressionNeeded: estimatedTokens > int(float64(cp.config.MaxTokens)*0.9),
	}
}

// ApplyCompression applies intelligent compression to messages
func (cp *ContextProcessor) ApplyCompression(ctx context.Context, messages []*session.Message) ([]*session.Message, error) {
	// Separate messages by importance
	recentMessages, importantMessages, regularMessages := cp.categorizeMessages(messages)

	var compressedMessages []*session.Message

	// Add important messages
	compressedMessages = append(compressedMessages, importantMessages...)

	// Compress regular messages if needed
	if len(regularMessages) > 10 {
		summaryMsg, err := cp.compressionEngine.CreateSummary(ctx, regularMessages)
		if err != nil {
			log.Printf("[WARN] Failed to create summary: %v", err)
			// Fallback to keeping last few messages
			keepCount := 3
			if len(regularMessages) > keepCount {
				compressedMessages = append(compressedMessages, regularMessages[len(regularMessages)-keepCount:]...)
			} else {
				compressedMessages = append(compressedMessages, regularMessages...)
			}
		} else {
			compressedMessages = append(compressedMessages, summaryMsg)
			// Keep a few recent regular messages
			keepCount := 3
			if len(regularMessages) > keepCount {
				compressedMessages = append(compressedMessages, regularMessages[len(regularMessages)-keepCount:]...)
			}
		}
	} else {
		compressedMessages = append(compressedMessages, regularMessages...)
	}

	// Add recent messages
	compressedMessages = append(compressedMessages, recentMessages...)

	return compressedMessages, nil
}

// categorizeMessages categorizes messages by importance
func (cp *ContextProcessor) categorizeMessages(messages []*session.Message) (recent, important, regular []*session.Message) {
	// Identify recent messages
	recentStart := len(messages) - cp.config.RecentKeepCount
	if recentStart < 0 {
		recentStart = 0
	}
	recent = messages[recentStart:]

	// Analyze messages before recent ones
	for i := 0; i < recentStart; i++ {
		msg := messages[i]
		if cp.messageAnalyzer.IsImportant(msg) {
			important = append(important, msg)
		} else {
			regular = append(regular, msg)
		}
	}

	return recent, important, regular
}

// EstimateSessionMessages estimates token count for session messages
func (te *TokenEstimator) EstimateSessionMessages(messages []*session.Message) int {
	totalChars := 0
	
	for _, msg := range messages {
		totalChars += len(msg.Content) + te.overhead
		
		// Add tokens for tool calls
		for _, tc := range msg.ToolCalls {
			totalChars += len(tc.Name) + len(tc.ID) + 50
		}
	}
	
	return totalChars / te.charsPerToken
}

// IsImportant determines if a message is important
func (ma *MessageAnalyzer) IsImportant(msg *session.Message) bool {
	// Memory messages are always important
	if msgType, ok := msg.Metadata["type"].(string); ok {
		if msgType == "memory_context" || strings.Contains(msgType, "memory") {
			return true
		}
	}

	// Memory-related content is important
	if strings.Contains(msg.Content, "## Relevant Context from Memory") ||
		strings.Contains(msg.Content, "### CodeContext") ||
		strings.Contains(msg.Content, "### TaskHistory") ||
		strings.Contains(msg.Content, "### Solutions") {
		return true
	}

	// Tool calls are important
	if len(msg.ToolCalls) > 0 {
		return true
	}

	// Error messages are important
	content := strings.ToLower(msg.Content)
	errorKeywords := []string{"error", "failed", "exception", "panic", "bug", "issue"}
	for _, keyword := range errorKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}

	// Long messages might be important
	if len(msg.Content) > 200 {
		return true
	}

	// Code blocks are important
	if strings.Contains(msg.Content, "```") {
		return true
	}

	// Summaries are important
	if strings.Contains(msg.Content, "Conversation Summary") ||
		strings.Contains(msg.Content, "## Summary") {
		return true
	}

	return false
}

// CreateSummary creates a summary of messages using LLM
func (ce *CompressionEngine) CreateSummary(ctx context.Context, messages []*session.Message) (*session.Message, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages to summarize")
	}

	// Build conversation text
	conversationText := ce.buildConversationText(messages)
	
	// Create summary prompt
	prompt := fmt.Sprintf(`Please create a concise summary of the following conversation (%d messages).

Requirements:
1. Extract key decisions, actions, and outcomes
2. Preserve important technical details
3. Highlight tool usage and results
4. Maintain chronological flow
5. Keep under 400 words

Conversation:
%s

Summary:`, len(messages), conversationText)

	// Call LLM for summarization
	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You are an expert at summarizing conversations concisely."},
			{Role: "user", Content: prompt},
		},
		ModelType: llm.BasicModel,
		MaxTokens: 800,
		Config: &llm.Config{
			Temperature: 0.3,
			MaxTokens:   800,
		},
	}

	response, err := ce.llmClient.Chat(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("LLM summarization failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	summary := strings.TrimSpace(response.Choices[0].Message.Content)
	if len(summary) == 0 {
		return nil, fmt.Errorf("empty summary from LLM")
	}

	return &session.Message{
		Role:    "system",
		Content: summary,
		Metadata: map[string]interface{}{
			"type":               "llm_summary",
			"original_count":     len(messages),
			"summary_timestamp":  time.Now().Unix(),
			"compression_method": "llm",
		},
		Timestamp: time.Now(),
	}, nil
}

// buildConversationText builds text representation of conversation
func (ce *CompressionEngine) buildConversationText(messages []*session.Message) string {
	var parts []string
	
	for i, msg := range messages {
		// Skip system messages except summaries
		if msg.Role == "system" {
			if msgType, ok := msg.Metadata["type"].(string); ok {
				if !strings.Contains(msgType, "summary") {
					continue
				}
			}
		}

		// Limit message length
		content := msg.Content
		if len(content) > 500 {
			content = content[:500] + "...[truncated]"
		}

		// Format message
		roleName := strings.ToUpper(msg.Role[:1]) + msg.Role[1:]
		if msg.Role == "tool" {
			if toolName, ok := msg.Metadata["tool_name"].(string); ok {
				roleName = fmt.Sprintf("Tool(%s)", toolName)
			}
		}

		// Add tool call info
		if len(msg.ToolCalls) > 0 {
			var tools []string
			for _, tc := range msg.ToolCalls {
				tools = append(tools, tc.Name)
			}
			content += fmt.Sprintf(" [Tools: %s]", strings.Join(tools, ", "))
		}

		parts = append(parts, fmt.Sprintf("[%d] %s: %s", i+1, roleName, content))
	}

	return strings.Join(parts, "\n")
}

