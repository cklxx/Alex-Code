package agent

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	contextmgr "alex/internal/context"
	"alex/internal/llm"
	"alex/internal/memory"
	"alex/internal/session"
)

// MessageProcessor 统一的消息处理器，整合所有消息相关功能
type MessageProcessor struct {
	contextMgr     *contextmgr.ContextManager
	sessionManager *session.Manager
	tokenEstimator *TokenEstimator
}

// NewMessageProcessor 创建统一的消息处理器
func NewMessageProcessor(llmClient llm.Client, sessionManager *session.Manager) *MessageProcessor {
	// 创建上下文管理器
	contextConfig := &contextmgr.ContextLengthConfig{
		MaxTokens:              8000,
		SummarizationThreshold: 6000,
		CompressionRatio:       0.3,
		PreserveSystemMessages: true,
	}

	return &MessageProcessor{
		contextMgr:     contextmgr.NewContextManager(llmClient, contextConfig),
		sessionManager: sessionManager,
		tokenEstimator: NewTokenEstimator(),
	}
}

// ========== 消息转换 ==========

// ConvertSessionToLLM 将 session 消息转换为 LLM 格式
func (mp *MessageProcessor) ConvertSessionToLLM(sessionMessages []*session.Message) []llm.Message {
	messages := make([]llm.Message, 0, len(sessionMessages))

	for _, msg := range sessionMessages {
		llmMsg := llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}

		// 处理工具调用
		if len(msg.ToolCalls) > 0 {
			llmMsg.ToolCalls = make([]llm.ToolCall, 0, len(msg.ToolCalls))
			for _, tc := range msg.ToolCalls {
				llmMsg.ToolCalls = append(llmMsg.ToolCalls, llm.ToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: llm.Function{
						Name: tc.Name,
					},
				})
			}
		}

		// 处理工具调用 ID
		if msg.Role == "tool" {
			if callID, ok := msg.Metadata["tool_call_id"].(string); ok {
				llmMsg.ToolCallId = callID
			}
		}

		messages = append(messages, llmMsg)
	}

	return messages
}

// ConvertLLMToSession 将 LLM 消息转换为 session 格式
func (mp *MessageProcessor) ConvertLLMToSession(llmMessages []llm.Message) []*session.Message {
	messages := make([]*session.Message, 0, len(llmMessages))

	for _, msg := range llmMessages {
		sessionMsg := &session.Message{
			Role:      msg.Role,
			Content:   msg.Content,
			Metadata:  make(map[string]interface{}),
			Timestamp: time.Now(),
		}

		// 处理工具调用
		if len(msg.ToolCalls) > 0 {
			sessionMsg.ToolCalls = make([]session.ToolCall, 0, len(msg.ToolCalls))
			for _, tc := range msg.ToolCalls {
				sessionMsg.ToolCalls = append(sessionMsg.ToolCalls, session.ToolCall{
					ID:   tc.ID,
					Name: tc.Function.Name,
				})
			}
		}

		// 处理工具调用 ID
		if msg.Role == "tool" && msg.ToolCallId != "" {
			sessionMsg.Metadata["tool_call_id"] = msg.ToolCallId
		}

		messages = append(messages, sessionMsg)
	}

	return messages
}

// formatMemoryContent 格式化内存内容
func (mp *MessageProcessor) formatMemoryContent(memories []*memory.MemoryItem) string {
	if len(memories) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, "## Relevant Context from Memory\n")

	// 按类别分组
	categoryGroups := make(map[memory.MemoryCategory][]*memory.MemoryItem)
	for _, mem := range memories {
		categoryGroups[mem.Category] = append(categoryGroups[mem.Category], mem)
	}

	// 格式化每个类别
	for category, items := range categoryGroups {
		if len(items) == 0 {
			continue
		}
		categoryName := strings.ToUpper(string(category)[:1]) + string(category)[1:]
		parts = append(parts, fmt.Sprintf("### %s", categoryName))
		for _, item := range items {
			content := item.Content
			if len(content) > 150 {
				content = content[:150] + "..."
			}
			parts = append(parts, fmt.Sprintf("- %s", content))
		}
		parts = append(parts, "")
	}

	return strings.Join(parts, "\n")
}

// ========== 消息压缩 ==========

// compressMessages 智能压缩消息
func (mp *MessageProcessor) compressMessages(sessionMessages []*session.Message) []*session.Message {
	const (
		MaxMessages = 100
		MaxTokens   = 600000
		RecentKeep  = 6
	)

	// 检查是否需要压缩
	if len(sessionMessages) <= MaxMessages {
		estimatedTokens := mp.tokenEstimator.EstimateSessionMessages(sessionMessages)
		if estimatedTokens <= MaxTokens {
			return sessionMessages
		}
	}

	log.Printf("[INFO] Message compression triggered: %d messages, estimated %d tokens",
		len(sessionMessages), mp.tokenEstimator.EstimateSessionMessages(sessionMessages))
	// 构建压缩后的消息
	var compressedMessages []*session.Message
	summaryMsg := mp.createMessageSummary(sessionMessages)
	if summaryMsg != nil {
		compressedMessages = append(compressedMessages, summaryMsg)
	}

	log.Printf("[INFO] Message compression completed: %d -> %d messages",
		len(sessionMessages), len(compressedMessages))

	return compressedMessages
}

// createMessageSummary 创建消息摘要
func (mp *MessageProcessor) createMessageSummary(messages []*session.Message) *session.Message {
	if len(messages) == 0 {
		return nil
	}

	// 尝试 LLM 摘要
	if summary := mp.createLLMSummary(messages); summary != nil {
		return summary
	}

	// 回退到统计摘要
	return mp.createStatisticalSummary(messages)
}

// createLLMSummary 使用 LLM 创建摘要
func (mp *MessageProcessor) createLLMSummary(messages []*session.Message) *session.Message {
	llmClient, err := llm.GetLLMInstance(llm.BasicModel)
	if err != nil {
		log.Printf("[WARN] Failed to get LLM instance for summary: %v", err)
		return nil
	}

	// 构建压缩输入
	conversationText := mp.buildSummaryInput(messages)
	if len(conversationText) == 0 {
		return nil
	}

	// 构建摘要请求
	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: "Create concise summaries that preserve key information while reducing length.",
			},
			{
				Role:    "user",
				Content: mp.buildSummaryPrompt(conversationText, len(messages)),
			},
		},
		ModelType: llm.BasicModel,
		Config: &llm.Config{
			Temperature: 0.8,
			MaxTokens:   10000,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := llmClient.Chat(ctx, request)
	if err != nil || len(response.Choices) == 0 {
		log.Printf("[WARN] LLM summary failed: %v", err)
		return nil
	}

	summaryContent := strings.TrimSpace(response.Choices[0].Message.Content)
	if len(summaryContent) == 0 {
		return nil
	}

	return &session.Message{
		Role:    "system",
		Content: summaryContent,
		Metadata: map[string]interface{}{
			"type":               "llm_summary",
			"original_count":     len(messages),
			"summary_timestamp":  time.Now().Unix(),
			"compression_method": "llm",
		},
		Timestamp: time.Now(),
	}
}

// buildSummaryInput 构建摘要输入
func (mp *MessageProcessor) buildSummaryInput(messages []*session.Message) string {
	var parts []string

	for i, msg := range messages {
		if msg.Role == "system" {
			if msgType, ok := msg.Metadata["type"].(string); ok {
				if strings.Contains(msgType, "summary") {
					continue
				}
			}
		}

		content := msg.Content
		if len(content) > 500 {
			content = content[:500] + "...[truncated]"
		}

		roleName := strings.ToUpper(msg.Role[:1]) + msg.Role[1:]
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

// buildSummaryPrompt 构建摘要提示
func (mp *MessageProcessor) buildSummaryPrompt(conversationText string, messageCount int) string {
	return fmt.Sprintf(`Please create a concise summary of the following conversation (%d messages).

Requirements:
1. Extract key decisions, actions, and outcomes
2. Preserve important technical details
3. Highlight tool usage and results
4. Maintain chronological flow
5. Keep under 400 words
6. Use structured format

Conversation:
%s

Summary:`, messageCount, conversationText)
}

// createStatisticalSummary 创建统计摘要
func (mp *MessageProcessor) createStatisticalSummary(messages []*session.Message) *session.Message {
	var userActions, toolUsages, keyTopics []string

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			content := msg.Content
			if len(content) > 50 {
				content = content[:50] + "..."
			}
			userActions = append(userActions, content)
		case "assistant":
			for _, tc := range msg.ToolCalls {
				toolUsages = append(toolUsages, tc.Name)
			}
		case "tool":
			if toolName, ok := msg.Metadata["tool_name"].(string); ok {
				success := "✓"
				if toolSuccess, ok := msg.Metadata["tool_success"].(bool); ok && !toolSuccess {
					success = "✗"
				}
				keyTopics = append(keyTopics, fmt.Sprintf("%s%s", success, toolName))
			}
		}
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("## Conversation Summary (%d messages)", len(messages)))

	if len(userActions) > 0 {
		parts = append(parts, fmt.Sprintf("**User Requests**: %s", strings.Join(userActions, "; ")))
	}

	if len(toolUsages) > 0 {
		toolCount := make(map[string]int)
		for _, tool := range toolUsages {
			toolCount[tool]++
		}
		var toolSummary []string
		for tool, count := range toolCount {
			if count > 1 {
				toolSummary = append(toolSummary, fmt.Sprintf("%s(%d)", tool, count))
			} else {
				toolSummary = append(toolSummary, tool)
			}
		}
		parts = append(parts, fmt.Sprintf("**Tools Used**: %s", strings.Join(toolSummary, ", ")))
	}

	if len(keyTopics) > 0 {
		parts = append(parts, fmt.Sprintf("**Key Activities**: %s", strings.Join(keyTopics, ", ")))
	}

	return &session.Message{
		Role:    "system",
		Content: strings.Join(parts, "\n"),
		Metadata: map[string]interface{}{
			"type":               "statistical_summary",
			"original_count":     len(messages),
			"summary_timestamp":  time.Now().Unix(),
			"compression_method": "statistical",
		},
		Timestamp: time.Now(),
	}
}

// ========== 会话管理 ==========

// GetCurrentSession 获取当前会话
func (mp *MessageProcessor) GetCurrentSession(ctx context.Context, agent *ReactAgent) *session.Session {
	if agent.currentSession != nil {
		return agent.currentSession
	}

	// 尝试从context中获取session ID
	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok && sessionID != "" {
		sess, err := mp.sessionManager.RestoreSession(sessionID)
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

// GetContextStats 获取上下文统计信息
func (mp *MessageProcessor) GetContextStats(sess *session.Session) *contextmgr.ContextStats {
	if mp.contextMgr == nil || sess == nil {
		return &contextmgr.ContextStats{
			TotalMessages:   0,
			EstimatedTokens: 0,
		}
	}
	return mp.contextMgr.GetContextStats(sess)
}

// RestoreFullContext 恢复完整上下文
func (mp *MessageProcessor) RestoreFullContext(sess *session.Session, backupID string) error {
	if mp.contextMgr == nil {
		return fmt.Errorf("context manager not available")
	}
	return mp.contextMgr.RestoreFullContext(sess, backupID)
}

// addTaskInstructions 添加任务指令
func (mp *MessageProcessor) addTaskInstructions(messages []llm.Message) {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			taskInstruction := "\n\nthink about the task and break it down into a list of todos and then call the todo_update tool to create the todos"
			if !strings.Contains(messages[i].Content, "think about the task") {
				messages[i].Content += taskInstruction
			}
			break
		}
	}
}

// ========== 随机消息生成 ==========

var processingMessages = []string{
	"Processing", "Thinking", "Learning", "Exploring", "Discovering",
	"Analyzing", "Computing", "Reasoning", "Planning", "Executing",
	"Optimizing", "Searching", "Understanding", "Crafting", "Creating",
	"Parsing", "Generating", "Evaluating", "Calculating", "Investigating",
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// GetRandomProcessingMessage 获取随机处理消息
func GetRandomProcessingMessage() string {
	return "👾 " + processingMessages[rng.Intn(len(processingMessages))] + "..."
}

// GetRandomProcessingMessageWithEmoji 获取带emoji的随机处理消息
func GetRandomProcessingMessageWithEmoji() string {
	return "⚡ " + GetRandomProcessingMessage() + " please wait"
}
