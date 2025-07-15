package agent

import (
	"context"
	"fmt"
	"log"

	contextmgr "alex/internal/context"
	"alex/internal/llm"
	"alex/internal/session"
)

// ContextHandler handles session context and overflow management
type ContextHandler struct {
	contextMgr     *contextmgr.ContextManager
	sessionManager *session.Manager

	// 新增：各功能模块
	messageBuilder     *MessageBuilder
	contextCompression *ContextCompression
	memoryIntegration  *MemoryIntegration
}

// NewContextHandler creates a new context handler
func NewContextHandler(llmClient llm.Client, sessionManager *session.Manager) *ContextHandler {
	// 创建上下文管理器
	contextConfig := &contextmgr.ContextLengthConfig{
		MaxTokens:              8000, // 保守的token限制
		SummarizationThreshold: 6000, // 75%时开始总结
		CompressionRatio:       0.3,  // 压缩到30%
		PreserveSystemMessages: true,
	}

	ctxMgr := contextmgr.NewContextManager(llmClient, contextConfig)

	handler := &ContextHandler{
		contextMgr:     ctxMgr,
		sessionManager: sessionManager,
	}

	// 初始化各功能模块
	handler.messageBuilder = NewMessageBuilder(handler)
	handler.contextCompression = NewContextCompression(handler)
	handler.memoryIntegration = NewMemoryIntegration(handler)

	return handler
}

// SessionIDKey is already defined in react_agent.go

// getCurrentSession - 获取当前会话
func (h *ContextHandler) getCurrentSession(ctx context.Context, agent *ReactAgent) *session.Session {
	if agent.currentSession != nil {
		return agent.currentSession
	}

	// 尝试从context中获取session ID
	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok && sessionID != "" {
		sess, err := h.sessionManager.RestoreSession(sessionID)
		if err == nil {
			agent.mu.Lock()
			agent.currentSession = sess
			agent.mu.Unlock()
			return sess
		}
		log.Printf("[WARNING] Failed to restore session %s: %v", sessionID, err)
	}

	return nil
}

// handleContextOverflow - 处理上下文溢出
func (h *ContextHandler) handleContextOverflow(ctx context.Context, sess *session.Session, streamCallback StreamCallback) error {
	// 检查上下文长度
	analysis, err := h.contextMgr.CheckContextLength(sess)
	if err != nil {
		return fmt.Errorf("failed to check context length: %w", err)
	}

	// 如果需要处理上下文溢出
	if analysis.RequiresTrimming {
		if streamCallback != nil {
			streamCallback(StreamChunk{
				Type:     "context_management",
				Content:  fmt.Sprintf("⚠️ Context overflow detected (%d tokens), summarizing conversation...", analysis.EstimatedTokens),
				Metadata: map[string]any{"action": "summarizing", "tokens": analysis.EstimatedTokens},
			})
		}

		result, err := h.contextMgr.ProcessContextOverflow(ctx, sess)
		if err != nil {
			return fmt.Errorf("failed to process context overflow: %w", err)
		}

		if streamCallback != nil {
			streamCallback(StreamChunk{
				Type:     "context_management",
				Content:  fmt.Sprintf("✅ Context summarized: %d → %d messages (backup: %s)", result.OriginalCount, result.ProcessedCount, result.BackupID),
				Metadata: map[string]any{"action": "completed", "backup_id": result.BackupID},
			})
		}

		log.Printf("[INFO] Context summarized: %s, %d → %d messages", result.Action, result.OriginalCount, result.ProcessedCount)
	}

	return nil
}

// GetContextStats - 获取上下文统计信息
func (h *ContextHandler) GetContextStats(sess *session.Session) *contextmgr.ContextStats {
	if h.contextMgr == nil || sess == nil {
		return &contextmgr.ContextStats{
			TotalMessages:   0,
			EstimatedTokens: 0,
		}
	}

	return h.contextMgr.GetContextStats(sess)
}

// ForceContextSummarization - 强制进行上下文总结
func (h *ContextHandler) ForceContextSummarization(ctx context.Context, sess *session.Session) (*contextmgr.ContextProcessingResult, error) {
	if h.contextMgr == nil {
		return nil, fmt.Errorf("context manager not available")
	}

	return h.contextMgr.ProcessContextOverflow(ctx, sess)
}

// RestoreFullContext - 恢复完整上下文
func (h *ContextHandler) RestoreFullContext(sess *session.Session, backupID string) error {
	if h.contextMgr == nil {
		return fmt.Errorf("context manager not available")
	}

	return h.contextMgr.RestoreFullContext(sess, backupID)
}

// ========== 委托方法 - 调用各功能模块 ==========

// buildMessagesFromSession - 委托给MessageBuilder
func (h *ContextHandler) buildMessagesFromSession(sess *session.Session, currentTask string, systemPrompt string) []llm.Message {
	return h.messageBuilder.buildMessagesFromSession(sess, currentTask, systemPrompt)
}

// updateMessagesWithSessionContent - 委托给MessageBuilder（核心功能，保留）
func (h *ContextHandler) updateMessagesWithSessionContent(ctx context.Context, sess *session.Session, baseMessages []llm.Message) []llm.Message {
	return h.messageBuilder.updateMessagesWithSessionContent(ctx, sess, baseMessages)
}

// ========== 模块访问器 ==========

// GetMessageBuilder - 获取消息构建器
func (h *ContextHandler) GetMessageBuilder() *MessageBuilder {
	return h.messageBuilder
}

// GetContextCompression - 获取上下文压缩器
func (h *ContextHandler) GetContextCompression() *ContextCompression {
	return h.contextCompression
}

// GetMemoryIntegration - 获取Memory集成器
func (h *ContextHandler) GetMemoryIntegration() *MemoryIntegration {
	return h.memoryIntegration
}
