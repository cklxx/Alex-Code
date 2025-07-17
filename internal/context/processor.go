package context

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

// ContextProcessorConfig defines configuration for context processing
type ContextProcessorConfig struct {
	MaxTokens              int     `json:"max_tokens"`
	MaxMessages            int     `json:"max_messages"`
	RecentKeepCount        int     `json:"recent_keep_count"`
	SummarizationThreshold int     `json:"summarization_threshold"`
	CompressionRatio       float64 `json:"compression_ratio"`
	PreserveSystemMessages bool    `json:"preserve_system_messages"`
}

// ContextProcessor provides unified context processing and compression
type ContextProcessor struct {
	tokenEstimator    *utils.TokenEstimator
	contentAnalyzer   *utils.ContentAnalyzer
	summarizer        *utils.ConversationSummarizer
	config            *ContextProcessorConfig
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
		tokenEstimator:  utils.NewTokenEstimator(),
		contentAnalyzer: utils.NewContentAnalyzer(),
		summarizer:      utils.NewConversationSummarizer(llmClient),
		config:          config,
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
	estimatedTokens := cp.tokenEstimator.EstimateMessages(messages)

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
		summary, err := cp.summarizer.SummarizeMessages(ctx, regularMessages)
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
			summaryMsg := &session.Message{
				Role:    "system",
				Content: summary,
				Metadata: map[string]interface{}{
					"type":               "context_summary",
					"original_count":     len(regularMessages),
					"compression_method": "llm",
				},
				Timestamp: time.Now(),
			}
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
		if cp.isImportant(msg) {
			important = append(important, msg)
		} else {
			regular = append(regular, msg)
		}
	}

	return recent, important, regular
}

// isImportant determines if a message is important using unified logic
func (cp *ContextProcessor) isImportant(msg *session.Message) bool {
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
	if cp.contentAnalyzer.ContainsError(msg.Content) {
		return true
	}

	// Long messages might be important
	if len(msg.Content) > 200 {
		return true
	}

	// Code blocks are important
	if cp.contentAnalyzer.HasCodeContent(msg.Content) {
		return true
	}

	// Summaries are important
	if strings.Contains(msg.Content, "Conversation Summary") ||
		strings.Contains(msg.Content, "## Summary") {
		return true
	}

	return false
}
