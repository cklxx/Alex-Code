package context

import (
	"alex/internal/llm"
	"alex/internal/session"
	"alex/internal/utils"
)

// ContextManager handles intelligent context management for long conversations
type ContextManager struct {
	llmClient        llm.Client
	maxContextTokens int
	summarizer       *MessageSummarizer
	preservationMgr  *ContextPreservationManager
	tokenEstimator   *utils.TokenEstimator
}


// ContextLengthConfig defines configuration for context length management
type ContextLengthConfig struct {
	MaxTokens              int     `json:"max_tokens"`
	SummarizationThreshold int     `json:"summarization_threshold"`
	CompressionRatio       float64 `json:"compression_ratio"`
	PreserveSystemMessages bool    `json:"preserve_system_messages"`
}

// NewContextManager creates a new context manager
func NewContextManager(llmClient llm.Client, config *ContextLengthConfig) *ContextManager {
	if config == nil {
		config = &ContextLengthConfig{
			MaxTokens:              8000, // Conservative default
			SummarizationThreshold: 6000, // Start summarizing at 75% of max
			CompressionRatio:       0.3,  // Compress to 30% of original
			PreserveSystemMessages: true,
		}
	}

	return &ContextManager{
		llmClient:        llmClient,
		maxContextTokens: config.MaxTokens,
		summarizer:       NewMessageSummarizer(llmClient, config),
		preservationMgr:  NewContextPreservationManager(),
		tokenEstimator:   utils.NewTokenEstimator(),
	}
}

// RestoreFullContext restores the complete conversation history
func (cm *ContextManager) RestoreFullContext(sess *session.Session, backupID string) error {
	return cm.preservationMgr.RestoreBackup(sess, backupID)
}

// GetContextStats returns detailed context statistics
func (cm *ContextManager) GetContextStats(sess *session.Session) *ContextStats {
	messages := sess.GetMessages()
	systemMsgs, userMsgs, assistantMsgs := 0, 0, 0
	summaryMsgs := 0

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			if metadata, ok := msg.Metadata["type"]; ok && metadata == "context_summary" {
				summaryMsgs++
			} else {
				systemMsgs++
			}
		case "user":
			userMsgs++
		case "assistant":
			assistantMsgs++
		}
	}

	return &ContextStats{
		TotalMessages:     len(messages),
		SystemMessages:    systemMsgs,
		UserMessages:      userMsgs,
		AssistantMessages: assistantMsgs,
		SummaryMessages:   summaryMsgs,
		EstimatedTokens:   cm.tokenEstimator.EstimateMessages(messages),
		MaxTokens:         cm.maxContextTokens,
	}
}

// Private helper methods

// Data structures

// ContextAnalysis represents the result of context length analysis
type ContextAnalysis struct {
	TotalMessages     int  `json:"total_messages"`
	EstimatedTokens   int  `json:"estimated_tokens"`
	RequiresTrimming  bool `json:"requires_trimming"`
	ShouldSummarize   bool `json:"should_summarize"`
	CompressionNeeded bool `json:"compression_needed"`
}

// ContextProcessingResult represents the result of context processing
type ContextProcessingResult struct {
	Action         string          `json:"action"`
	OriginalCount  int             `json:"original_count"`
	ProcessedCount int             `json:"processed_count"`
	Summary        *MessageSummary `json:"summary,omitempty"`
	BackupID       string          `json:"backup_id,omitempty"`
}

// ContextStats provides detailed statistics about session context
type ContextStats struct {
	TotalMessages     int `json:"total_messages"`
	SystemMessages    int `json:"system_messages"`
	UserMessages      int `json:"user_messages"`
	AssistantMessages int `json:"assistant_messages"`
	SummaryMessages   int `json:"summary_messages"`
	EstimatedTokens   int `json:"estimated_tokens"`
	MaxTokens         int `json:"max_tokens"`
}
