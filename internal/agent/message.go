package agent

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	contextmgr "alex/internal/context"
	"alex/internal/llm"
	"alex/internal/session"
)

// CachedCompressionResult - LLM压缩结果缓存
type CachedCompressionResult struct {
	Summary      *session.Message
	Timestamp    time.Time
	InputHash    string
	MessageCount int
}

// MessageProcessor 统一的消息处理器，整合所有消息相关功能
type MessageProcessor struct {
	contextMgr     *contextmgr.ContextManager
	sessionManager *session.Manager
	tokenEstimator *TokenEstimator

	// LLM压缩缓存优化
	compressionCache map[string]*CachedCompressionResult
	compressionMutex sync.RWMutex
	cacheExpiry      time.Duration
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

		// 初始化LLM压缩缓存
		compressionCache: make(map[string]*CachedCompressionResult),
		cacheExpiry:      30 * time.Minute, // 30分钟缓存过期
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

// ========== 消息压缩 ==========

// compressMessages 智能压缩消息
func (mp *MessageProcessor) compressMessages(sessionMessages []*session.Message) []*session.Message {
	const (
		MaxMessages = 80   // 降低消息数量阈值
		MaxTokens   = 6000 // 降低token阈值，预留空间
		RecentKeep  = 10   // 保留更多最近消息
	)

	// 检查是否需要压缩 - 更早触发
	estimatedTokens := mp.tokenEstimator.EstimateSessionMessages(sessionMessages)
	if len(sessionMessages) <= 10 || estimatedTokens <= 3000 {
		return sessionMessages // 消息很少时不压缩
	}

	// 渐进式压缩：根据压力等级选择策略
	if len(sessionMessages) > MaxMessages || estimatedTokens > MaxTokens {
		log.Printf("[INFO] High pressure compression triggered: %d messages, estimated %d tokens",
			len(sessionMessages), estimatedTokens)
		return mp.aggressiveCompress(sessionMessages, RecentKeep)
	}

	// 中等压力：保留重要消息和最近消息
	if len(sessionMessages) > 40 || estimatedTokens > 4000 {
		log.Printf("[INFO] Medium pressure compression triggered: %d messages, estimated %d tokens",
			len(sessionMessages), estimatedTokens)
		return mp.moderateCompress(sessionMessages, RecentKeep)
	}

	// 低压力：只做轻微整理
	if len(sessionMessages) > 25 || estimatedTokens > 3500 {
		log.Printf("[INFO] Light compression triggered: %d messages, estimated %d tokens",
			len(sessionMessages), estimatedTokens)
		return mp.lightCompress(sessionMessages, RecentKeep)
	}

	return sessionMessages
}

// aggressiveCompress 激进压缩：生成摘要
func (mp *MessageProcessor) aggressiveCompress(messages []*session.Message, recentKeep int) []*session.Message {
	if len(messages) <= recentKeep {
		return messages
	}

	// 保留最近的消息
	recentMessages := messages[len(messages)-recentKeep:]

	// 对其余消息创建摘要
	oldMessages := messages[:len(messages)-recentKeep]
	summaryMsg := mp.createMessageSummary(oldMessages)

	var result []*session.Message
	if summaryMsg != nil {
		result = append(result, summaryMsg)
	}
	result = append(result, recentMessages...)

	log.Printf("[INFO] Aggressive compression: %d -> %d messages", len(messages), len(result))
	return result
}

// moderateCompress 中等压缩：选择性保留重要消息
func (mp *MessageProcessor) moderateCompress(messages []*session.Message, recentKeep int) []*session.Message {
	if len(messages) <= recentKeep+5 {
		return messages
	}

	// 保留最近的消息
	recentMessages := messages[len(messages)-recentKeep:]

	// 从旧消息中选择重要的消息
	oldMessages := messages[:len(messages)-recentKeep]
	importantMessages := mp.selectImportantMessages(oldMessages, 5)

	var result []*session.Message
	result = append(result, importantMessages...)
	result = append(result, recentMessages...)

	log.Printf("[INFO] Moderate compression: %d -> %d messages", len(messages), len(result))
	return result
}

// lightCompress 轻度压缩：移除低价值消息
func (mp *MessageProcessor) lightCompress(messages []*session.Message, recentKeep int) []*session.Message {
	var result []*session.Message

	for i, msg := range messages {
		// 总是保留最近的消息
		if i >= len(messages)-recentKeep {
			result = append(result, msg)
			continue
		}

		// 移除低价值消息
		if mp.isLowValueMessage(msg) {
			continue
		}

		result = append(result, msg)
	}

	log.Printf("[INFO] Light compression: %d -> %d messages", len(messages), len(result))
	return result
}

// selectImportantMessages 选择重要消息
func (mp *MessageProcessor) selectImportantMessages(messages []*session.Message, maxCount int) []*session.Message {
	if len(messages) <= maxCount {
		return messages
	}

	// 计算消息重要性分数
	type msgWithScore struct {
		msg   *session.Message
		score float64
	}

	var scored []msgWithScore
	for _, msg := range messages {
		score := mp.calculateMessageImportance(msg)
		scored = append(scored, msgWithScore{msg, score})
	}

	// 按分数排序
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[i].score < scored[j].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// 选择前maxCount个
	var result []*session.Message
	for i := 0; i < maxCount && i < len(scored); i++ {
		result = append(result, scored[i].msg)
	}

	return result
}

// isLowValueMessage 判断是否为低价值消息
func (mp *MessageProcessor) isLowValueMessage(msg *session.Message) bool {
	content := strings.TrimSpace(msg.Content)

	// 空消息不被认为是低价值消息（在别处处理）
	if len(content) == 0 {
		return false
	}

	// 短消息通常价值较低
	if len(content) < 10 {
		return true
	}

	// 纯确认消息
	lowValuePhrases := []string{
		"好的", "OK", "ok", "是的", "收到", "明白", "了解", "谢谢",
		"好", "行", "没问题", "可以", "继续", "next", "yes", "no",
	}

	contentLower := strings.ToLower(content)
	for _, phrase := range lowValuePhrases {
		if contentLower == strings.ToLower(phrase) {
			return true
		}
	}

	return false
}

// calculateMessageImportance 计算消息重要性
func (mp *MessageProcessor) calculateMessageImportance(msg *session.Message) float64 {
	score := 0.0
	content := msg.Content

	// 长度因子
	score += float64(len(content)) * 0.01

	// 代码块加分
	if strings.Contains(content, "```") {
		score += 10.0
	}

	// 工具调用加分
	if len(msg.ToolCalls) > 0 {
		score += float64(len(msg.ToolCalls)) * 5.0
	}

	// 错误信息加分
	if strings.Contains(strings.ToLower(content), "error") ||
		strings.Contains(strings.ToLower(content), "错误") {
		score += 5.0
	}

	// 问题和解决方案加分
	if strings.Contains(content, "?") || strings.Contains(content, "？") {
		score += 5.0
	}

	// 关键词加分
	keywords := []string{"implement", "how", "why", "what", "where", "when", "function", "method", "error", "issue", "problem"}
	contentLower := strings.ToLower(content)
	for _, keyword := range keywords {
		if strings.Contains(contentLower, keyword) {
			score += 2.0
		}
	}

	return score
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

// createLLMSummary 使用 LLM 创建摘要 - 优化版本带缓存
func (mp *MessageProcessor) createLLMSummary(messages []*session.Message) *session.Message {
	// 首先检查缓存
	if cachedSummary := mp.getCachedCompressionResult(messages); cachedSummary != nil {
		return cachedSummary
	}

	// 智能决策：是否使用LLM压缩
	if !mp.shouldUseLLMCompression(messages) {
		log.Printf("[INFO] Using statistical summary instead of LLM for %d messages", len(messages))
		return mp.createStatisticalSummary(messages)
	}

	llmClient, err := llm.GetLLMInstance(llm.BasicModel)
	if err != nil {
		log.Printf("[WARN] Failed to get LLM instance for summary: %v", err)
		return mp.createStatisticalSummary(messages) // 降级到统计摘要
	}

	// 构建压缩输入
	conversationText := mp.buildSummaryInput(messages)
	if len(conversationText) == 0 {
		return nil
	}

	// 构建优化的摘要请求
	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: mp.buildOptimizedSystemPrompt(),
			},
			{
				Role:    "user",
				Content: mp.buildOptimizedSummaryPrompt(conversationText, len(messages)),
			},
		},
		ModelType: llm.BasicModel,
		Config: &llm.Config{
			Temperature: 0.3,  // 降低温度提高一致性
			MaxTokens:   1000, // 减少最大token数
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // 减少超时时间
	defer cancel()

	response, err := llmClient.Chat(ctx, request)
	if err != nil || len(response.Choices) == 0 {
		log.Printf("[WARN] LLM summary failed: %v, falling back to statistical summary", err)
		return mp.createStatisticalSummary(messages) // 降级到统计摘要
	}

	summaryContent := strings.TrimSpace(response.Choices[0].Message.Content)
	if len(summaryContent) == 0 {
		return mp.createStatisticalSummary(messages) // 降级到统计摘要
	}

	summary := &session.Message{
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

	// 缓存压缩结果
	mp.setCachedCompressionResult(messages, summary)

	log.Printf("[INFO] LLM summary created for %d messages, cached for future use", len(messages))
	return summary
}

// ========== LLM压缩优化方法 ==========

// getCachedCompressionResult 获取缓存的压缩结果
func (mp *MessageProcessor) getCachedCompressionResult(messages []*session.Message) *session.Message {
	if len(messages) == 0 {
		return nil
	}

	// 构建缓存键
	inputHash := mp.buildMessageHash(messages)

	mp.compressionMutex.RLock()
	defer mp.compressionMutex.RUnlock()

	if cached, exists := mp.compressionCache[inputHash]; exists {
		// 检查缓存是否过期
		if time.Since(cached.Timestamp) < mp.cacheExpiry {
			log.Printf("[DEBUG] Cache hit for compression of %d messages", len(messages))
			return cached.Summary
		}
		// 缓存过期，删除
		delete(mp.compressionCache, inputHash)
	}

	return nil
}

// setCachedCompressionResult 设置缓存的压缩结果
func (mp *MessageProcessor) setCachedCompressionResult(messages []*session.Message, summary *session.Message) {
	if len(messages) == 0 || summary == nil {
		return
	}

	inputHash := mp.buildMessageHash(messages)

	mp.compressionMutex.Lock()
	defer mp.compressionMutex.Unlock()

	mp.compressionCache[inputHash] = &CachedCompressionResult{
		Summary:      summary,
		Timestamp:    time.Now(),
		InputHash:    inputHash,
		MessageCount: len(messages),
	}
}

// buildMessageHash 为消息列表构建哈希
func (mp *MessageProcessor) buildMessageHash(messages []*session.Message) string {
	var hashInput strings.Builder
	for i, msg := range messages {
		if i > 0 {
			hashInput.WriteString("|")
		}
		hashInput.WriteString(msg.Role)
		hashInput.WriteString(":")
		// 只使用内容的前100个字符来构建哈希，避免过长
		content := msg.Content
		if len(content) > 100 {
			content = content[:100]
		}
		hashInput.WriteString(content)
	}
	return fmt.Sprintf("%x", hashInput.String())
}

// shouldUseLLMCompression 智能决策是否使用LLM压缩
func (mp *MessageProcessor) shouldUseLLMCompression(messages []*session.Message) bool {
	// 消息数量太少不值得LLM压缩
	if len(messages) < 10 {
		return false
	}

	// 估算token数量
	estimatedTokens := mp.tokenEstimator.EstimateSessionMessages(messages)

	// 内容太少不值得LLM压缩
	if estimatedTokens < 2000 {
		return false
	}

	// 检查内容复杂度
	complexity := mp.calculateContentComplexity(messages)

	// 高复杂度内容使用LLM压缩
	if complexity > 0.7 {
		return true
	}

	// 中等复杂度且消息数量多时使用LLM压缩
	if complexity > 0.5 && len(messages) > 20 {
		return true
	}

	// 其他情况使用统计摘要
	return false
}

// calculateContentComplexity 计算内容复杂度
func (mp *MessageProcessor) calculateContentComplexity(messages []*session.Message) float64 {
	var totalScore float64
	var totalLength int

	for _, msg := range messages {
		content := msg.Content
		totalLength += len(content)

		// 代码内容复杂度更高
		if strings.Contains(content, "```") || strings.Contains(content, "function") || strings.Contains(content, "class") {
			totalScore += 1.0
		}

		// 错误消息复杂度高
		if strings.Contains(strings.ToLower(content), "error") || strings.Contains(strings.ToLower(content), "exception") {
			totalScore += 0.8
		}

		// 工具调用复杂度高
		if len(msg.ToolCalls) > 0 {
			totalScore += 0.6
		}

		// 长内容复杂度高
		if len(content) > 500 {
			totalScore += 0.4
		}
	}

	if len(messages) == 0 {
		return 0
	}

	return totalScore / float64(len(messages))
}

// buildOptimizedSystemPrompt 构建优化的系统提示
func (mp *MessageProcessor) buildOptimizedSystemPrompt() string {
	return `Create a concise summary that preserves key information while significantly reducing length.
Focus on:
1. Important decisions and outcomes
2. Key technical details and solutions
3. Error patterns and fixes
4. Critical context for future reference

Output format: Brief paragraph format, no bullet points.
Target: 70-80% length reduction while maintaining essential information.`
}

// buildOptimizedSummaryPrompt 构建优化的摘要提示
func (mp *MessageProcessor) buildOptimizedSummaryPrompt(conversationText string, messageCount int) string {
	return fmt.Sprintf(`Summarize this conversation thread of %d messages:

%s

Create a concise summary that captures the essential information while reducing length by 70-80%%. Focus on key decisions, technical solutions, and important context.`, messageCount, conversationText)
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
