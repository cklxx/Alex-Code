package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"alex/internal/llm"
	"alex/internal/memory"
	"alex/internal/session"
)

// ContextManager - 真正的上下文管理器
type ContextManager struct {
	llmClient      llm.Client
	memoryManager  *memory.MemoryManager
	tokenEstimator *TokenEstimator

	// 上下文策略
	maxContextTokens int
	compressionRatio float64
	qualityThreshold float64

	// 缓存
	contextCache map[string]*CachedContext
	cacheMutex   sync.RWMutex
	cacheExpiry  time.Duration
}

// CachedContext - 缓存的上下文
type CachedContext struct {
	CompressedMessages []*session.Message
	QualityScore       float64
	TokenCount         int
	LastModified       time.Time
}

// ContextQuality - 上下文质量评估
type ContextQuality struct {
	Score              float64   `json:"score"`
	TokenUtilization   float64   `json:"token_utilization"`
	InformationDensity float64   `json:"information_density"`
	Freshness          float64   `json:"freshness"`
	Coherence          float64   `json:"coherence"`
	Timestamp          time.Time `json:"timestamp"`
}

// NewContextManager - 创建上下文管理器
func NewContextManager(llmClient llm.Client, memoryManager *memory.MemoryManager) *ContextManager {
	return &ContextManager{
		llmClient:      llmClient,
		memoryManager:  memoryManager,
		tokenEstimator: NewTokenEstimator(),

		// 默认策略
		maxContextTokens: 80000, // 80K token限制
		compressionRatio: 0.6,   // 60%压缩比
		qualityThreshold: 0.7,   // 70%质量阈值

		// 缓存
		contextCache: make(map[string]*CachedContext),
		cacheExpiry:  10 * time.Minute,
	}
}

// OptimizeContext - 智能上下文优化
func (cm *ContextManager) OptimizeContext(ctx context.Context, sessionID string, messages []*session.Message) ([]*session.Message, error) {
	// 1. 检查缓存
	if cached := cm.getCachedContext(sessionID, messages); cached != nil {
		log.Printf("[CONTEXT] Cache hit for session %s", sessionID)
		return cached.CompressedMessages, nil
	}

	// 2. 评估当前上下文质量
	quality := cm.evaluateContextQuality(messages)
	log.Printf("[CONTEXT] Quality score: %.2f for session %s", quality.Score, sessionID)

	// 3. 根据质量决定优化策略
	var optimizedMessages []*session.Message
	var err error

	if quality.Score < cm.qualityThreshold {
		// 质量不足，需要智能重构
		optimizedMessages, err = cm.intelligentRestructure(ctx, messages)
	} else {
		// 质量良好，只需轻微调整
		optimizedMessages, err = cm.lightOptimization(messages)
	}

	if err != nil {
		return messages, err // 失败时返回原始消息
	}

	// 4. 缓存结果
	cm.setCachedContext(sessionID, messages, optimizedMessages, quality.Score)

	log.Printf("[CONTEXT] Optimized %d -> %d messages for session %s",
		len(messages), len(optimizedMessages), sessionID)

	return optimizedMessages, nil
}

// intelligentRestructure - 智能重构上下文
func (cm *ContextManager) intelligentRestructure(ctx context.Context, messages []*session.Message) ([]*session.Message, error) {
	if cm.llmClient == nil {
		return cm.lightOptimization(messages)
	}

	// 构建重构请求
	prompt := cm.buildRestructurePrompt(messages)

	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: "You are a context optimization expert. Restructure the conversation to maintain key information while improving coherence and reducing redundancy.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ModelType: llm.BasicModel,
		Config: &llm.Config{
			Temperature: 0.3,
			MaxTokens:   2000,
		},
	}

	response, err := cm.llmClient.Chat(ctx, request)
	if err != nil {
		log.Printf("[CONTEXT] LLM restructure failed: %v", err)
		return cm.lightOptimization(messages)
	}

	if len(response.Choices) == 0 {
		return cm.lightOptimization(messages)
	}

	// 解析LLM的重构结果
	restructuredContent := response.Choices[0].Message.Content
	return cm.parseRestructuredContent(restructuredContent, messages)
}

// lightOptimization - 轻量级优化
func (cm *ContextManager) lightOptimization(messages []*session.Message) ([]*session.Message, error) {
	if len(messages) <= 20 {
		return messages, nil
	}

	// 保留最近的消息
	recentCount := 15
	importantCount := 5

	recent := messages[len(messages)-recentCount:]

	// 从历史消息中选择重要的
	historical := messages[:len(messages)-recentCount]
	important := cm.selectImportantMessages(historical, importantCount)

	// 合并
	result := make([]*session.Message, 0, len(important)+len(recent))
	result = append(result, important...)
	result = append(result, recent...)

	return result, nil
}

// evaluateContextQuality - 评估上下文质量
func (cm *ContextManager) evaluateContextQuality(messages []*session.Message) *ContextQuality {
	if len(messages) == 0 {
		return &ContextQuality{
			Score:     1.0,
			Timestamp: time.Now(),
		}
	}

	// 计算各项指标
	tokenUtil := cm.calculateTokenUtilization(messages)
	infoDensity := cm.calculateInformationDensity(messages)
	freshness := cm.calculateFreshness(messages)
	coherence := cm.calculateCoherence(messages)

	// 加权计算总分
	score := tokenUtil*0.3 + infoDensity*0.3 + freshness*0.2 + coherence*0.2

	return &ContextQuality{
		Score:              score,
		TokenUtilization:   tokenUtil,
		InformationDensity: infoDensity,
		Freshness:          freshness,
		Coherence:          coherence,
		Timestamp:          time.Now(),
	}
}

// calculateTokenUtilization - 计算token利用率
func (cm *ContextManager) calculateTokenUtilization(messages []*session.Message) float64 {
	totalTokens := cm.tokenEstimator.EstimateSessionMessages(messages)
	utilization := float64(totalTokens) / float64(cm.maxContextTokens)

	// 理想范围: 0.6-0.8
	if utilization < 0.6 {
		return utilization / 0.6
	} else if utilization <= 0.8 {
		return 1.0
	} else {
		return 1.0 - (utilization-0.8)/0.2
	}
}

// calculateInformationDensity - 计算信息密度
func (cm *ContextManager) calculateInformationDensity(messages []*session.Message) float64 {
	if len(messages) == 0 {
		return 1.0
	}

	var totalScore float64
	for _, msg := range messages {
		score := cm.scoreMessageInformation(msg)
		totalScore += score
	}

	return totalScore / float64(len(messages))
}

// calculateFreshness - 计算新鲜度
func (cm *ContextManager) calculateFreshness(messages []*session.Message) float64 {
	if len(messages) == 0 {
		return 1.0
	}

	now := time.Now()
	var totalAge time.Duration

	for _, msg := range messages {
		age := now.Sub(msg.Timestamp)
		totalAge += age
	}

	avgAge := totalAge / time.Duration(len(messages))

	// 1小时内新鲜度1.0，24小时后0.1
	if avgAge < time.Hour {
		return 1.0
	} else if avgAge < 24*time.Hour {
		return 1.0 - float64(avgAge-time.Hour)/float64(23*time.Hour)
	}
	return 0.1
}

// calculateCoherence - 计算连贯性
func (cm *ContextManager) calculateCoherence(messages []*session.Message) float64 {
	if len(messages) <= 1 {
		return 1.0
	}

	// 简单的连贯性评估：检查相邻消息的话题相关性
	coherentPairs := 0
	totalPairs := len(messages) - 1

	for i := 0; i < len(messages)-1; i++ {
		if cm.areMessagesCoherent(messages[i], messages[i+1]) {
			coherentPairs++
		}
	}

	return float64(coherentPairs) / float64(totalPairs)
}

// scoreMessageInformation - 评估消息信息价值
func (cm *ContextManager) scoreMessageInformation(msg *session.Message) float64 {
	content := msg.Content
	score := 0.0

	// 代码块高价值
	if strings.Contains(content, "```") {
		score += 0.8
	}

	// 工具调用高价值
	if len(msg.ToolCalls) > 0 {
		score += 0.7
	}

	// 错误信息高价值
	if strings.Contains(strings.ToLower(content), "error") {
		score += 0.6
	}

	// 解决方案高价值
	if strings.Contains(strings.ToLower(content), "solution") ||
		strings.Contains(strings.ToLower(content), "fix") {
		score += 0.7
	}

	// 适中长度加分
	if len(content) > 50 && len(content) < 1000 {
		score += 0.3
	}

	// 避免过高评分
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// areMessagesCoherent - 检查消息连贯性
func (cm *ContextManager) areMessagesCoherent(msg1, msg2 *session.Message) bool {
	// 简单的连贯性检查：共同关键词
	words1 := strings.Fields(strings.ToLower(msg1.Content))
	words2 := strings.Fields(strings.ToLower(msg2.Content))

	// 技术关键词
	techKeywords := []string{"function", "method", "class", "error", "bug", "fix", "code", "implement"}

	common := 0
	for _, word1 := range words1 {
		for _, word2 := range words2 {
			if word1 == word2 && len(word1) > 3 {
				common++
				break
			}
		}
		// 检查技术关键词
		for _, keyword := range techKeywords {
			if strings.Contains(word1, keyword) {
				for _, word2 := range words2 {
					if strings.Contains(word2, keyword) {
						common++
						break
					}
				}
			}
		}
	}

	return common >= 2 // 至少2个共同关键词
}

// selectImportantMessages - 选择重要消息
func (cm *ContextManager) selectImportantMessages(messages []*session.Message, count int) []*session.Message {
	if len(messages) <= count {
		return messages
	}

	// 计算重要性分数
	type scoredMessage struct {
		msg   *session.Message
		score float64
	}

	var scored []scoredMessage
	for _, msg := range messages {
		score := cm.scoreMessageInformation(msg)
		scored = append(scored, scoredMessage{msg, score})
	}

	// 按分数排序
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// 选择前count个
	result := make([]*session.Message, 0, count)
	for i := 0; i < count && i < len(scored); i++ {
		result = append(result, scored[i].msg)
	}

	return result
}

// buildRestructurePrompt - 构建重构提示
func (cm *ContextManager) buildRestructurePrompt(messages []*session.Message) string {
	var parts []string
	parts = append(parts, "Please restructure this conversation to improve coherence and reduce redundancy:")
	parts = append(parts, "")

	for i, msg := range messages {
		if i > 20 { // 限制长度
			break
		}
		parts = append(parts, fmt.Sprintf("[%s]: %s", msg.Role, msg.Content))
	}

	parts = append(parts, "")
	parts = append(parts, "Output a restructured version that:")
	parts = append(parts, "1. Maintains all key information")
	parts = append(parts, "2. Improves logical flow")
	parts = append(parts, "3. Reduces redundancy")
	parts = append(parts, "4. Preserves technical details")

	return strings.Join(parts, "\n")
}

// parseRestructuredContent - 解析重构内容
func (cm *ContextManager) parseRestructuredContent(content string, originalMessages []*session.Message) ([]*session.Message, error) {
	// 简单解析：将重构内容作为系统消息
	restructuredMsg := &session.Message{
		Role:    "system",
		Content: content,
		Metadata: map[string]interface{}{
			"type":           "restructured_context",
			"original_count": len(originalMessages),
			"timestamp":      time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}

	// 保留最近的几条消息
	recentCount := 5
	var result []*session.Message
	result = append(result, restructuredMsg)

	if len(originalMessages) > recentCount {
		result = append(result, originalMessages[len(originalMessages)-recentCount:]...)
	} else {
		result = append(result, originalMessages...)
	}

	return result, nil
}

// 缓存管理方法
func (cm *ContextManager) getCachedContext(sessionID string, messages []*session.Message) *CachedContext {
	cm.cacheMutex.RLock()
	defer cm.cacheMutex.RUnlock()

	key := fmt.Sprintf("%s:%d", sessionID, len(messages))
	if cached, exists := cm.contextCache[key]; exists {
		if time.Since(cached.LastModified) < cm.cacheExpiry {
			return cached
		}
		delete(cm.contextCache, key)
	}

	return nil
}

func (cm *ContextManager) setCachedContext(sessionID string, original, compressed []*session.Message, quality float64) {
	cm.cacheMutex.Lock()
	defer cm.cacheMutex.Unlock()

	key := fmt.Sprintf("%s:%d", sessionID, len(original))
	cm.contextCache[key] = &CachedContext{
		CompressedMessages: compressed,
		QualityScore:       quality,
		TokenCount:         cm.tokenEstimator.EstimateSessionMessages(compressed),
		LastModified:       time.Now(),
	}
}
