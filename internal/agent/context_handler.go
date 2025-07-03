package agent

import (
	"context"
	"fmt"
	"log"
	"strings"

	contextmgr "alex/internal/context"
	"alex/internal/llm"
	"alex/internal/session"
)

// ContextHandler handles session context and overflow management
type ContextHandler struct {
	contextMgr     *contextmgr.ContextManager
	sessionManager *session.Manager
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

	return &ContextHandler{
		contextMgr:     ctxMgr,
		sessionManager: sessionManager,
	}
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

// buildMessagesFromSession - 基于会话历史构建消息列表
func (h *ContextHandler) buildMessagesFromSession(sess *session.Session, currentTask string, systemPrompt string) []llm.Message {
	var messages []llm.Message

	// 添加系统提示
	messages = append(messages, llm.Message{
		Role:    "system",
		Content: systemPrompt,
	})

	// 如果有会话历史，添加相关历史消息
	if sess != nil {
		historyMessages := sess.GetMessages()

		// 限制历史消息数量，只包含最近的对话
		maxHistoryMessages := 10
		startIdx := 0
		if len(historyMessages) > maxHistoryMessages {
			startIdx = len(historyMessages) - maxHistoryMessages
		}

		for i := startIdx; i < len(historyMessages); i++ {
			msg := historyMessages[i]

			// 跳过空消息
			if strings.TrimSpace(msg.Content) == "" {
				continue
			}

			llmMsg := llm.Message{
				Role:    msg.Role,
				Content: msg.Content,
			}

			// 添加工具调用信息
			if len(msg.ToolCalls) > 0 {
				var toolCalls []llm.ToolCall
				for _, tc := range msg.ToolCalls {
					toolCalls = append(toolCalls, llm.ToolCall{
						ID:   tc.ID,
						Type: "function",
						Function: llm.Function{
							Name:      tc.Name,
							Arguments: fmt.Sprintf("%v", tc.Args),
						},
					})
				}
				llmMsg.ToolCalls = toolCalls
			}

			messages = append(messages, llmMsg)
		}
	}

	// 添加当前任务
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: currentTask + "\n\n think about the task and break it down into a list of todos and then call the todo_update tool to create the todos",
	})

	return messages
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