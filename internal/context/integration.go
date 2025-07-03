package context

import (
	"context"
	"fmt"
	"log"

	"alex/internal/llm"
	"alex/internal/session"
)

// ReactAgentContextIntegration provides context management integration for ReactAgent
type ReactAgentContextIntegration struct {
	contextManager *ContextManager
	enabled        bool
	autoTrimming   bool
}

// IntegrationConfig configures the context management integration
type IntegrationConfig struct {
	Enabled             bool                 `json:"enabled"`
	AutoTrimming        bool                 `json:"auto_trimming"`
	ContextLengthConfig *ContextLengthConfig `json:"context_length_config"`
}

// NewReactAgentContextIntegration creates a new context management integration
func NewReactAgentContextIntegration(llmClient llm.Client, config *IntegrationConfig) *ReactAgentContextIntegration {
	if config == nil {
		config = &IntegrationConfig{
			Enabled:      true,
			AutoTrimming: true,
			ContextLengthConfig: &ContextLengthConfig{
				MaxTokens:              8000,
				SummarizationThreshold: 6000,
				CompressionRatio:       0.3,
				PreserveSystemMessages: true,
			},
		}
	}

	return &ReactAgentContextIntegration{
		contextManager: NewContextManager(llmClient, config.ContextLengthConfig),
		enabled:        config.Enabled,
		autoTrimming:   config.AutoTrimming,
	}
}

// ProcessMessageWithContextManagement processes a message with automatic context management
func (raci *ReactAgentContextIntegration) ProcessMessageWithContextManagement(
	ctx context.Context,
	sess *session.Session,
	userMessage string,
	processFunc func(context.Context, *session.Session, string) error,
) error {
	if !raci.enabled {
		// If context management is disabled, just process normally
		return processFunc(ctx, sess, userMessage)
	}

	// Check if context management is needed before processing
	if raci.autoTrimming {
		if err := raci.checkAndProcessContextOverflow(ctx, sess); err != nil {
			log.Printf("[WARNING] Context management failed: %v", err)
			// Continue processing even if context management fails
		}
	}

	// Process the message normally
	return processFunc(ctx, sess, userMessage)
}

// CheckContextStatus returns the current context status
func (raci *ReactAgentContextIntegration) CheckContextStatus(sess *session.Session) (*ContextAnalysis, error) {
	if !raci.enabled {
		return &ContextAnalysis{
			TotalMessages:    sess.GetMessageCount(),
			EstimatedTokens:  0,
			RequiresTrimming: false,
		}, nil
	}

	return raci.contextManager.CheckContextLength(sess)
}

// ForceContextSummarization forces context summarization regardless of thresholds
func (raci *ReactAgentContextIntegration) ForceContextSummarization(ctx context.Context, sess *session.Session) (*ContextProcessingResult, error) {
	if !raci.enabled {
		return nil, fmt.Errorf("context management is disabled")
	}

	return raci.contextManager.ProcessContextOverflow(ctx, sess)
}

// RestoreFullContext restores the complete conversation history
func (raci *ReactAgentContextIntegration) RestoreFullContext(sess *session.Session, backupID string) error {
	if !raci.enabled {
		return fmt.Errorf("context management is disabled")
	}

	return raci.contextManager.RestoreFullContext(sess, backupID)
}

// GetContextStats returns detailed context statistics
func (raci *ReactAgentContextIntegration) GetContextStats(sess *session.Session) *ContextStats {
	if !raci.enabled {
		return &ContextStats{
			TotalMessages:   sess.GetMessageCount(),
			EstimatedTokens: 0,
		}
	}

	return raci.contextManager.GetContextStats(sess)
}

// EnableContextManagement enables context management
func (raci *ReactAgentContextIntegration) EnableContextManagement() {
	raci.enabled = true
}

// DisableContextManagement disables context management
func (raci *ReactAgentContextIntegration) DisableContextManagement() {
	raci.enabled = false
}

// SetAutoTrimming enables or disables automatic context trimming
func (raci *ReactAgentContextIntegration) SetAutoTrimming(enabled bool) {
	raci.autoTrimming = enabled
}

// IsEnabled returns whether context management is enabled
func (raci *ReactAgentContextIntegration) IsEnabled() bool {
	return raci.enabled
}

// IsAutoTrimmingEnabled returns whether auto-trimming is enabled
func (raci *ReactAgentContextIntegration) IsAutoTrimmingEnabled() bool {
	return raci.autoTrimming
}

// Private helper methods

func (raci *ReactAgentContextIntegration) checkAndProcessContextOverflow(ctx context.Context, sess *session.Session) error {
	analysis, err := raci.contextManager.CheckContextLength(sess)
	if err != nil {
		return fmt.Errorf("failed to analyze context length: %w", err)
	}

	if analysis.RequiresTrimming {
		log.Printf("[INFO] Context overflow detected, processing %d messages", analysis.TotalMessages)

		result, err := raci.contextManager.ProcessContextOverflow(ctx, sess)
		if err != nil {
			return fmt.Errorf("failed to process context overflow: %w", err)
		}

		log.Printf("[INFO] Context processed: %s, %d -> %d messages (backup: %s)",
			result.Action, result.OriginalCount, result.ProcessedCount, result.BackupID)
	}

	return nil
}

// ContextManagementSlashCommands provides slash command support for context management
type ContextManagementSlashCommands struct {
	integration *ReactAgentContextIntegration
}

// NewContextManagementSlashCommands creates slash command support
func NewContextManagementSlashCommands(integration *ReactAgentContextIntegration) *ContextManagementSlashCommands {
	return &ContextManagementSlashCommands{
		integration: integration,
	}
}

// HandleSlashCommand handles context management slash commands
func (cmsc *ContextManagementSlashCommands) HandleSlashCommand(ctx context.Context, sess *session.Session, command string, args []string) (string, error) {
	switch command {
	case "context-status":
		return cmsc.handleContextStatus(sess)
	case "context-summarize":
		return cmsc.handleContextSummarize(ctx, sess)
	case "context-restore":
		if len(args) == 0 {
			return "", fmt.Errorf("backup ID required for context-restore command")
		}
		return cmsc.handleContextRestore(sess, args[0])
	case "context-stats":
		return cmsc.handleContextStats(sess)
	case "context-enable":
		return cmsc.handleContextEnable()
	case "context-disable":
		return cmsc.handleContextDisable()
	default:
		return "", fmt.Errorf("unknown context management command: %s", command)
	}
}

func (cmsc *ContextManagementSlashCommands) handleContextStatus(sess *session.Session) (string, error) {
	analysis, err := cmsc.integration.CheckContextStatus(sess)
	if err != nil {
		return "", err
	}

	status := fmt.Sprintf(`📊 Context Status:
• Total Messages: %d
• Estimated Tokens: %d
• Requires Trimming: %v
• Should Summarize: %v
• Compression Needed: %v`,
		analysis.TotalMessages,
		analysis.EstimatedTokens,
		analysis.RequiresTrimming,
		analysis.ShouldSummarize,
		analysis.CompressionNeeded)

	return status, nil
}

func (cmsc *ContextManagementSlashCommands) handleContextSummarize(ctx context.Context, sess *session.Session) (string, error) {
	result, err := cmsc.integration.ForceContextSummarization(ctx, sess)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`✅ Context summarized:
• Action: %s
• Messages: %d → %d
• Backup ID: %s`,
		result.Action,
		result.OriginalCount,
		result.ProcessedCount,
		result.BackupID), nil
}

func (cmsc *ContextManagementSlashCommands) handleContextRestore(sess *session.Session, backupID string) (string, error) {
	err := cmsc.integration.RestoreFullContext(sess, backupID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("✅ Context restored from backup: %s", backupID), nil
}

func (cmsc *ContextManagementSlashCommands) handleContextStats(sess *session.Session) (string, error) {
	stats := cmsc.integration.GetContextStats(sess)

	return fmt.Sprintf(`📈 Detailed Context Stats:
• Total Messages: %d
• System Messages: %d
• User Messages: %d
• Assistant Messages: %d
• Summary Messages: %d
• Estimated Tokens: %d / %d
• Usage: %.1f%%`,
		stats.TotalMessages,
		stats.SystemMessages,
		stats.UserMessages,
		stats.AssistantMessages,
		stats.SummaryMessages,
		stats.EstimatedTokens,
		stats.MaxTokens,
		float64(stats.EstimatedTokens)/float64(stats.MaxTokens)*100), nil
}

func (cmsc *ContextManagementSlashCommands) handleContextEnable() (string, error) {
	cmsc.integration.EnableContextManagement()
	return "✅ Context management enabled", nil
}

func (cmsc *ContextManagementSlashCommands) handleContextDisable() (string, error) {
	cmsc.integration.DisableContextManagement()
	return "⚠️ Context management disabled", nil
}
