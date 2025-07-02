package context

import (
	"context"
	"fmt"
	"time"

	"alex/internal/llm"
	"alex/internal/session"
)

// ContextManager handles intelligent context management for long conversations
type ContextManager struct {
	llmClient        llm.Client
	maxContextTokens int
	summarizer       *MessageSummarizer
	preservationMgr  *ContextPreservationManager
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
			MaxTokens:              8000,  // Conservative default
			SummarizationThreshold: 6000,  // Start summarizing at 75% of max
			CompressionRatio:       0.3,   // Compress to 30% of original
			PreserveSystemMessages: true,
		}
	}

	return &ContextManager{
		llmClient:        llmClient,
		maxContextTokens: config.MaxTokens,
		summarizer:       NewMessageSummarizer(llmClient, config),
		preservationMgr:  NewContextPreservationManager(),
	}
}

// CheckContextLength analyzes if the session context is approaching limits
func (cm *ContextManager) CheckContextLength(sess *session.Session) (*ContextAnalysis, error) {
	messages := sess.GetMessages()
	if len(messages) == 0 {
		return &ContextAnalysis{
			TotalMessages:    0,
			EstimatedTokens:  0,
			RequiresTrimming: false,
		}, nil
	}

	// Estimate token usage
	totalTokens := cm.estimateTokenUsage(messages)
	
	analysis := &ContextAnalysis{
		TotalMessages:    len(messages),
		EstimatedTokens:  totalTokens,
		RequiresTrimming: totalTokens > cm.maxContextTokens,
		ShouldSummarize:  totalTokens > int(float64(cm.maxContextTokens)*0.75), // 75% threshold
		CompressionNeeded: totalTokens > int(float64(cm.maxContextTokens)*0.9), // 90% threshold
	}

	return analysis, nil
}

// ProcessContextOverflow handles context overflow by summarizing messages
func (cm *ContextManager) ProcessContextOverflow(ctx context.Context, sess *session.Session) (*ContextProcessingResult, error) {
	analysis, err := cm.CheckContextLength(sess)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze context: %w", err)
	}

	if !analysis.RequiresTrimming {
		return &ContextProcessingResult{
			Action:         "no_action",
			OriginalCount:  analysis.TotalMessages,
			ProcessedCount: analysis.TotalMessages,
		}, nil
	}

	messages := sess.GetMessages()
	
	// Separate system messages and conversation messages
	systemMessages, conversationMessages := cm.separateMessages(messages)
	
	// Preserve context for restoration
	contextBackup := cm.preservationMgr.CreateBackup(sess)
	
	// Summarize conversation messages (excluding recent ones)
	recentCount := cm.calculateRecentMessageCount(len(conversationMessages))
	messagesToSummarize := conversationMessages[:len(conversationMessages)-recentCount]
	recentMessages := conversationMessages[len(conversationMessages)-recentCount:]
	
	summary, err := cm.summarizer.SummarizeMessages(ctx, messagesToSummarize)
	if err != nil {
		return nil, fmt.Errorf("failed to summarize messages: %w", err)
	}
	
	// Create new message list with summary + recent messages
	newMessages := make([]*session.Message, 0)
	
	// Add system messages back
	newMessages = append(newMessages, systemMessages...)
	
	// Add summary as a system message
	summaryMsg := &session.Message{
		Role:    "system",
		Content: fmt.Sprintf("## Conversation Summary\n\n%s\n\n---\n\nThe above is a summary of the previous conversation. Continue from here with full context awareness.", summary.Summary),
		Metadata: map[string]interface{}{
			"type":               "context_summary",
			"original_count":     len(messagesToSummarize),
			"summary_timestamp":  time.Now().Unix(),
			"backup_id":         contextBackup.ID,
			"key_points":        summary.KeyPoints,
			"topics":            summary.Topics,
		},
		Timestamp: time.Now(),
	}
	newMessages = append(newMessages, summaryMsg)
	
	// Add recent messages
	newMessages = append(newMessages, recentMessages...)
	
	// Update session with new message list
	sess.ClearMessages()
	for _, msg := range newMessages {
		sess.AddMessage(msg)
	}
	
	return &ContextProcessingResult{
		Action:         "summarized",
		OriginalCount:  len(messages),
		ProcessedCount: len(newMessages),
		Summary:        summary,
		BackupID:       contextBackup.ID,
	}, nil
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
		EstimatedTokens:   cm.estimateTokenUsage(messages),
		MaxTokens:         cm.maxContextTokens,
	}
}

// Private helper methods

func (cm *ContextManager) estimateTokenUsage(messages []*session.Message) int {
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Content)
		totalChars += 50 // Estimated overhead per message (higher for more realistic estimation)
	}
	// More conservative estimation: 3 characters per token on average
	return totalChars / 3
}

func (cm *ContextManager) separateMessages(messages []*session.Message) ([]*session.Message, []*session.Message) {
	var systemMessages []*session.Message
	var conversationMessages []*session.Message
	
	for _, msg := range messages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg)
		} else {
			conversationMessages = append(conversationMessages, msg)
		}
	}
	
	return systemMessages, conversationMessages
}

func (cm *ContextManager) calculateRecentMessageCount(totalConversation int) int {
	// Keep at least 5 recent messages, or 20% of conversation, whichever is higher
	minRecent := 5
	if totalConversation < minRecent {
		return totalConversation // Can't keep more than we have
	}
	
	ratioRecent := int(float64(totalConversation) * 0.2)
	
	if ratioRecent > minRecent {
		return ratioRecent
	}
	return minRecent
}

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
	Action         string             `json:"action"`
	OriginalCount  int                `json:"original_count"`
	ProcessedCount int                `json:"processed_count"`
	Summary        *MessageSummary    `json:"summary,omitempty"`
	BackupID       string             `json:"backup_id,omitempty"`
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