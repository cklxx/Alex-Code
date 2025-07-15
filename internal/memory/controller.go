package memory

import (
	"fmt"
	"strings"
	"time"

	"alex/internal/session"
)

// MemoryController controls when and what memories should be written
type MemoryController struct {
	config *MemoryControlConfig
}

// MemoryControlConfig defines rules for memory storage
type MemoryControlConfig struct {
	// 触发条件
	MinMessageCount    int     `json:"min_message_count"`     // 最少消息数才开始记忆
	MinImportanceScore float64 `json:"min_importance_score"`  // 最小重要性分数
	MaxMemoriesPerHour int     `json:"max_memories_per_hour"` // 每小时最大记忆数量

	// 过滤规则
	SkipSystemMessages bool     `json:"skip_system_messages"` // 跳过系统消息
	SkipShortMessages  bool     `json:"skip_short_messages"`  // 跳过短消息
	MinContentLength   int      `json:"min_content_length"`   // 最小内容长度
	ImportantKeywords  []string `json:"important_keywords"`   // 重要关键词
	SkipKeywords       []string `json:"skip_keywords"`        // 跳过关键词

	// 分类规则
	CodeKeywords       []string `json:"code_keywords"`       // 代码相关关键词
	ErrorKeywords      []string `json:"error_keywords"`      // 错误相关关键词
	SolutionKeywords   []string `json:"solution_keywords"`   // 解决方案关键词
	PreferenceKeywords []string `json:"preference_keywords"` // 用户偏好关键词
}

// NewMemoryController creates a new memory controller
func NewMemoryController() *MemoryController {
	config := &MemoryControlConfig{
		MinMessageCount:    3,   // 至少3条消息后开始记忆
		MinImportanceScore: 0.3, // 最小重要性0.3
		MaxMemoriesPerHour: 50,  // 每小时最多50条记忆

		SkipSystemMessages: true,
		SkipShortMessages:  true,
		MinContentLength:   20, // 最少20字符

		ImportantKeywords: []string{
			"error", "solution", "fix", "bug", "implement", "create", "delete",
			"modify", "change", "problem", "issue", "resolve", "configure",
			"install", "deploy", "test", "debug", "optimize", "refactor",
		},

		SkipKeywords: []string{
			"hello", "hi", "thanks", "thank you", "ok", "okay", "yes", "no",
			"sure", "please", "sorry", "excuse me", "got it", "understood",
		},

		CodeKeywords: []string{
			"function", "class", "method", "variable", "import", "export",
			"const", "let", "var", "def", "type", "interface", "struct",
			"package", "module", "library", "framework", "api", "endpoint",
			"database", "query", "sql", "json", "xml", "yaml", "config",
		},

		ErrorKeywords: []string{
			"error", "exception", "panic", "fail", "failed", "crash", "bug",
			"issue", "problem", "broken", "not working", "doesn't work",
			"syntax error", "runtime error", "compilation error",
		},

		SolutionKeywords: []string{
			"solution", "fix", "resolve", "solved", "fixed", "working", "success",
			"implement", "create", "add", "modify", "change", "update", "upgrade",
		},

		PreferenceKeywords: []string{
			"prefer", "like", "dislike", "favorite", "always", "never", "usually",
			"recommend", "suggest", "opinion", "think", "believe", "want",
		},
	}

	return &MemoryController{config: config}
}

// ShouldCreateMemory determines if a message should create memory items
func (mc *MemoryController) ShouldCreateMemory(msg *session.Message, sessionMessageCount int, recentMemoryCount int) bool {
	// 检查消息数量阈值
	if sessionMessageCount < mc.config.MinMessageCount {
		return false
	}

	// 检查每小时记忆数量限制
	if recentMemoryCount >= mc.config.MaxMemoriesPerHour {
		return false
	}

	// 跳过系统消息
	if mc.config.SkipSystemMessages && msg.Role == "system" {
		return false
	}

	// 检查内容长度
	if mc.config.SkipShortMessages && len(msg.Content) < mc.config.MinContentLength {
		return false
	}

	// 检查跳过关键词
	if mc.containsSkipKeywords(msg.Content) {
		return false
	}

	// 检查重要关键词
	if !mc.containsImportantKeywords(msg.Content) {
		return false
	}

	return true
}

// ClassifyMemory determines the category and importance of a memory
func (mc *MemoryController) ClassifyMemory(msg *session.Message) (MemoryCategory, float64, []string) {
	content := strings.ToLower(msg.Content)

	// 确定分类
	category := mc.determineCategory(content, msg)

	// 计算重要性分数
	importance := mc.calculateImportance(content, msg, category)

	// 生成标签
	tags := mc.generateTags(content, msg, category)

	return category, importance, tags
}

// ShouldPromoteToLongTerm determines if a memory should be promoted to long-term storage
func (mc *MemoryController) ShouldPromoteToLongTerm(item *MemoryItem, accessPattern *AccessPattern) bool {
	// 高重要性直接升级
	if item.Importance >= 0.8 {
		return true
	}

	// 访问频繁的记忆
	if accessPattern != nil && accessPattern.AccessCount >= 5 {
		return true
	}

	// 包含解决方案的记忆
	if item.Category == Solutions || item.Category == CodeContext {
		return true
	}

	// 错误模式记忆
	if item.Category == ErrorPatterns && item.Importance >= 0.6 {
		return true
	}

	return false
}

// AccessPattern represents memory access statistics
type AccessPattern struct {
	AccessCount  int
	LastAccess   time.Time
	AccessFreq   float64 // accesses per day
	RecentAccess bool    // accessed in last 24h
}

// FilterMemoriesForRecall filters memories for recall based on context
func (mc *MemoryController) FilterMemoriesForRecall(memories []*MemoryItem, context string, limit int) []*MemoryItem {
	contextLower := strings.ToLower(context)

	// 计算相关性分数
	type memoryScore struct {
		memory *MemoryItem
		score  float64
	}

	var scored []memoryScore
	for _, memory := range memories {
		relevance := mc.calculateRelevanceScore(memory, contextLower)
		if relevance >= 0.3 { // 最小相关性阈值
			scored = append(scored, memoryScore{memory, relevance})
		}
	}

	// 按分数排序
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[i].score < scored[j].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// 返回前N个
	var result []*MemoryItem
	for i := 0; i < len(scored) && i < limit; i++ {
		result = append(result, scored[i].memory)
	}

	return result
}

// Private helper methods

func (mc *MemoryController) containsSkipKeywords(content string) bool {
	contentLower := strings.ToLower(content)
	for _, keyword := range mc.config.SkipKeywords {
		if strings.Contains(contentLower, keyword) {
			return true
		}
	}
	return false
}

func (mc *MemoryController) containsImportantKeywords(content string) bool {
	contentLower := strings.ToLower(content)
	for _, keyword := range mc.config.ImportantKeywords {
		if strings.Contains(contentLower, keyword) {
			return true
		}
	}
	return false
}

func (mc *MemoryController) determineCategory(content string, msg *session.Message) MemoryCategory {
	// 检查错误相关 (优先级更高)
	if mc.containsKeywords(content, mc.config.ErrorKeywords) {
		return ErrorPatterns
	}

	// 检查代码相关
	if mc.containsKeywords(content, mc.config.CodeKeywords) || mc.hasCodeBlocks(msg.Content) {
		return CodeContext
	}

	// 检查解决方案相关
	if mc.containsKeywords(content, mc.config.SolutionKeywords) {
		return Solutions
	}

	// 检查用户偏好
	if mc.containsKeywords(content, mc.config.PreferenceKeywords) {
		return UserPreferences
	}

	// 检查工具使用
	if len(msg.ToolCalls) > 0 {
		return TaskHistory
	}

	// 默认分类
	return Knowledge
}

func (mc *MemoryController) calculateImportance(content string, msg *session.Message, category MemoryCategory) float64 {
	importance := 0.5 // 基础分数

	// 分类加权
	switch category {
	case ErrorPatterns:
		importance += 0.3
	case Solutions:
		importance += 0.3
	case CodeContext:
		importance += 0.2
	case UserPreferences:
		importance += 0.1
	case TaskHistory:
		importance += 0.1
	}

	// 内容长度加权
	if len(msg.Content) > 200 {
		importance += 0.1
	}
	if len(msg.Content) > 500 {
		importance += 0.1
	}

	// 工具调用加权
	if len(msg.ToolCalls) > 0 {
		importance += 0.1
		if len(msg.ToolCalls) > 2 {
			importance += 0.1
		}
	}

	// 代码块加权
	if mc.hasCodeBlocks(msg.Content) {
		importance += 0.2
	}

	// 确保在0-1范围内
	if importance > 1.0 {
		importance = 1.0
	}
	if importance < 0.0 {
		importance = 0.0
	}

	return importance
}

func (mc *MemoryController) generateTags(content string, msg *session.Message, category MemoryCategory) []string {
	var tags []string

	// 添加分类标签
	tags = append(tags, string(category))

	// 添加角色标签
	tags = append(tags, msg.Role)

	// 添加工具标签
	for _, toolCall := range msg.ToolCalls {
		tags = append(tags, "tool:"+toolCall.Name)
	}

	// 添加内容特征标签
	if mc.hasCodeBlocks(msg.Content) {
		tags = append(tags, "code")
	}

	if mc.containsKeywords(content, mc.config.ErrorKeywords) {
		tags = append(tags, "error")
	}

	if mc.containsKeywords(content, mc.config.SolutionKeywords) {
		tags = append(tags, "solution")
	}

	// 添加时间标签
	tags = append(tags, fmt.Sprintf("hour:%d", msg.Timestamp.Hour()))
	tags = append(tags, fmt.Sprintf("day:%s", msg.Timestamp.Weekday().String()))

	return tags
}

func (mc *MemoryController) calculateRelevanceScore(memory *MemoryItem, context string) float64 {
	score := memory.Importance * 0.5 // 基础重要性权重

	// 内容相似性
	contentSimilarity := mc.calculateTextSimilarity(strings.ToLower(memory.Content), context)
	score += contentSimilarity * 0.3

	// 标签匹配
	tagRelevance := mc.calculateTagRelevance(memory.Tags, context)
	score += tagRelevance * 0.2

	// 访问频率加权
	if memory.AccessCount > 0 {
		accessBoost := float64(memory.AccessCount) * 0.05
		if accessBoost > 0.2 {
			accessBoost = 0.2
		}
		score += accessBoost
	}

	// 时间衰减
	daysSinceCreation := time.Since(memory.CreatedAt).Hours() / 24
	if daysSinceCreation > 30 {
		score *= 0.9 // 30天后开始衰减
	}
	if daysSinceCreation > 90 {
		score *= 0.8 // 90天后进一步衰减
	}

	return score
}

func (mc *MemoryController) containsKeywords(content string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

func (mc *MemoryController) hasCodeBlocks(content string) bool {
	return strings.Contains(content, "```") ||
		strings.Contains(content, "func ") ||
		strings.Contains(content, "class ") ||
		strings.Contains(content, "def ") ||
		strings.Contains(content, "import ")
}

func (mc *MemoryController) calculateTextSimilarity(text1, text2 string) float64 {
	// 简单的关键词重叠计算
	words1 := strings.Fields(text1)
	words2 := strings.Fields(text2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	wordSet1 := make(map[string]bool)
	for _, word := range words1 {
		if len(word) > 3 { // 只考虑长度>3的单词
			wordSet1[word] = true
		}
	}

	overlap := 0
	totalWords := 0
	for _, word := range words2 {
		if len(word) > 3 {
			totalWords++
			if wordSet1[word] {
				overlap++
			}
		}
	}

	if totalWords == 0 {
		return 0.0
	}

	return float64(overlap) / float64(totalWords)
}

func (mc *MemoryController) calculateTagRelevance(tags []string, context string) float64 {
	if len(tags) == 0 {
		return 0.0
	}

	matches := 0
	for _, tag := range tags {
		if strings.Contains(context, tag) {
			matches++
		}
	}

	return float64(matches) / float64(len(tags))
}
