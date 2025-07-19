package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"alex/internal/llm"
	"alex/internal/memory"
	"alex/internal/session"
	"alex/internal/tools/builtin"
)

// ActiveMemoryManager - 主动记忆管理器
type ActiveMemoryManager struct {
	llmClient     llm.Client
	memoryManager *memory.MemoryManager
	
	// 记忆生成配置
	enableAutoGenerate bool
	generationThreshold int // 消息数量阈值
	
	// 缓存
	memoryCache      map[string]*memory.RecallResult
	memoryCacheMutex sync.RWMutex
	cacheExpiry      time.Duration
}

// MemoryKeyValue - LLM生成的键值对记忆
type MemoryKeyValue struct {
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	Category    string  `json:"category"`
	Importance  float64 `json:"importance"`
	Context     string  `json:"context,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// GeneratedMemories - LLM生成的记忆集合
type GeneratedMemories struct {
	SessionID string            `json:"session_id"`
	Memories  []MemoryKeyValue  `json:"memories"`
	Summary   string            `json:"summary"`
	Timestamp time.Time         `json:"timestamp"`
}

// NewActiveMemoryManager - 创建主动记忆管理器
func NewActiveMemoryManager(llmClient llm.Client, memoryManager *memory.MemoryManager) *ActiveMemoryManager {
	return &ActiveMemoryManager{
		llmClient:     llmClient,
		memoryManager: memoryManager,
		
		enableAutoGenerate: true,
		generationThreshold: 10, // 每10条消息生成一次记忆
		
		memoryCache: make(map[string]*memory.RecallResult),
		cacheExpiry: 10 * time.Minute,
	}
}

// GenerateMemoriesFromConversation - 让LLM从对话中生成结构化记忆
func (amm *ActiveMemoryManager) GenerateMemoriesFromConversation(ctx context.Context, sessionID string, messages []*session.Message) (*GeneratedMemories, error) {
	if amm.llmClient == nil {
		return nil, fmt.Errorf("LLM client not available")
	}
	
	// 构建记忆生成提示
	prompt := amm.buildMemoryGenerationPrompt(messages)
	
	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: amm.getMemoryGenerationSystemPrompt(),
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
	
	response, err := amm.llmClient.Chat(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("memory generation failed: %w", err)
	}
	
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}
	
	// 解析LLM生成的记忆
	generatedMemories, err := amm.parseGeneratedMemories(sessionID, response.Choices[0].Message.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated memories: %w", err)
	}
	
	// 存储生成的记忆
	err = amm.storeGeneratedMemories(generatedMemories)
	if err != nil {
		log.Printf("[WARN] Failed to store generated memories: %v", err)
	}
	
	log.Printf("[MEMORY] Generated %d memories for session %s", len(generatedMemories.Memories), sessionID)
	
	return generatedMemories, nil
}

// getMemoryGenerationSystemPrompt - 获取记忆生成系统提示
func (amm *ActiveMemoryManager) getMemoryGenerationSystemPrompt() string {
	return `You are a memory extraction expert. Your task is to analyze conversations and extract key information as structured key-value memories.

Instructions:
1. Extract important facts, decisions, solutions, and patterns
2. Create meaningful key-value pairs
3. Categorize memories appropriately
4. Rate importance (0.0-1.0)
5. Add relevant tags
6. Provide brief context if needed

Categories:
- code_context: Technical implementation details
- task_history: Completed tasks and decisions
- solutions: Problem-solving approaches
- error_patterns: Common errors and fixes
- user_preferences: User preferences and habits
- knowledge: General knowledge and facts

Output format: JSON with the following structure:
{
  "memories": [
    {
      "key": "descriptive_key",
      "value": "actual_value_or_fact",
      "category": "appropriate_category",
      "importance": 0.8,
      "context": "brief_context_if_needed",
      "tags": ["tag1", "tag2"]
    }
  ],
  "summary": "Brief summary of the conversation"
}`
}

// buildMemoryGenerationPrompt - 构建记忆生成提示
func (amm *ActiveMemoryManager) buildMemoryGenerationPrompt(messages []*session.Message) string {
	var parts []string
	parts = append(parts, "Extract key memories from this conversation:")
	parts = append(parts, "")
	
	// 限制消息数量以避免过长
	maxMessages := 20
	startIdx := 0
	if len(messages) > maxMessages {
		startIdx = len(messages) - maxMessages
	}
	
	for i := startIdx; i < len(messages); i++ {
		msg := messages[i]
		
		// 过滤掉系统消息和过短消息
		if msg.Role == "system" || len(strings.TrimSpace(msg.Content)) < 10 {
			continue
		}
		
		parts = append(parts, fmt.Sprintf("[%s]: %s", msg.Role, msg.Content))
		
		// 如果有工具调用，也包含进来
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				parts = append(parts, fmt.Sprintf("[tool_call]: %s", tc.Name))
			}
		}
	}
	
	parts = append(parts, "")
	parts = append(parts, "Focus on:")
	parts = append(parts, "- Technical decisions and implementations")
	parts = append(parts, "- Problem-solving approaches")
	parts = append(parts, "- Error patterns and fixes")
	parts = append(parts, "- User preferences and habits")
	parts = append(parts, "- Important facts and knowledge")
	
	return strings.Join(parts, "\n")
}

// parseGeneratedMemories - 解析LLM生成的记忆
func (amm *ActiveMemoryManager) parseGeneratedMemories(sessionID, content string) (*GeneratedMemories, error) {
	// 尝试从响应中提取JSON
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	
	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("no valid JSON found in response")
	}
	
	jsonContent := content[jsonStart : jsonEnd+1]
	
	var parsed struct {
		Memories []MemoryKeyValue `json:"memories"`
		Summary  string           `json:"summary"`
	}
	
	err := json.Unmarshal([]byte(jsonContent), &parsed)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	return &GeneratedMemories{
		SessionID: sessionID,
		Memories:  parsed.Memories,
		Summary:   parsed.Summary,
		Timestamp: time.Now(),
	}, nil
}

// Store - 存储记忆项
func (amm *ActiveMemoryManager) Store(item *memory.MemoryItem) error {
	if amm.memoryManager == nil {
		return fmt.Errorf("memory manager not available")
	}
	return amm.memoryManager.Store(item)
}

// storeGeneratedMemories - 存储生成的记忆
func (amm *ActiveMemoryManager) storeGeneratedMemories(generated *GeneratedMemories) error {
	if amm.memoryManager == nil {
		return fmt.Errorf("memory manager not available")
	}
	
	// Get project ID for the current context
	projectID, err := amm.getProjectID()
	if err != nil {
		// Fall back to session-based storage if project ID is not available
		return amm.storeGeneratedMemoriesLegacy(generated)
	}
	
	for _, mem := range generated.Memories {
		// 转换为memory.MemoryItem
		memoryItem := &memory.MemoryItem{
			ID:        fmt.Sprintf("%s_%s_%d", projectID, mem.Key, time.Now().UnixNano()),
			ProjectID: projectID,
			SessionID: generated.SessionID, // Keep for backwards compatibility
			Type:      memory.LongTermMemory,
			Category:  amm.mapCategory(mem.Category),
			Content:   fmt.Sprintf("%s: %s", mem.Key, mem.Value),
			Metadata: map[string]interface{}{
				"key":             mem.Key,
				"value":           mem.Value,
				"original_category": mem.Category,
				"context":         mem.Context,
				"generation_time": generated.Timestamp,
				"auto_generated":  true,
			},
			Importance:  mem.Importance,
			AccessCount: 0,
			CreatedAt:   generated.Timestamp,
			UpdatedAt:   generated.Timestamp,
			LastAccess:  generated.Timestamp,
			Tags:        mem.Tags,
		}
		
		// 存储到memory系统
		err := amm.Store(memoryItem)
		if err != nil {
			log.Printf("[WARN] Failed to store memory item %s: %v", mem.Key, err)
		}
	}
	
	// 创建会话总结记忆
	if generated.Summary != "" {
		summaryItem := &memory.MemoryItem{
			ID:        fmt.Sprintf("%s_summary_%d", projectID, generated.Timestamp.UnixNano()),
			ProjectID: projectID,
			SessionID: generated.SessionID, // Keep for backwards compatibility
			Type:      memory.ShortTermMemory,
			Category:  memory.TaskHistory,
			Content:   generated.Summary,
			Metadata: map[string]interface{}{
				"type":            "conversation_summary",
				"memory_count":    len(generated.Memories),
				"generation_time": generated.Timestamp,
				"auto_generated":  true,
			},
			Importance:  0.7,
			AccessCount: 0,
			CreatedAt:   generated.Timestamp,
			UpdatedAt:   generated.Timestamp,
			LastAccess:  generated.Timestamp,
			Tags:        []string{"summary", "auto_generated"},
		}
		
		err := amm.memoryManager.Store(summaryItem)
		if err != nil {
			log.Printf("[WARN] Failed to store summary memory: %v", err)
		}
	}
	
	return nil
}

// mapCategory - 映射分类
func (amm *ActiveMemoryManager) mapCategory(category string) memory.MemoryCategory {
	switch strings.ToLower(category) {
	case "code_context":
		return memory.CodeContext
	case "task_history":
		return memory.TaskHistory
	case "solutions":
		return memory.Solutions
	case "error_patterns":
		return memory.ErrorPatterns
	case "user_preferences":
		return memory.UserPreferences
	case "knowledge":
		return memory.Knowledge
	default:
		return memory.Knowledge
	}
}

// ShouldGenerateMemories - 判断是否应该生成记忆
func (amm *ActiveMemoryManager) ShouldGenerateMemories(messages []*session.Message) bool {
	if !amm.enableAutoGenerate {
		return false
	}
	
	// 检查消息数量
	if len(messages) < amm.generationThreshold {
		return false
	}
	
	// 检查是否有足够的实质性内容
	substantialCount := 0
	for _, msg := range messages {
		if msg.Role != "system" && len(strings.TrimSpace(msg.Content)) > 20 {
			substantialCount++
		}
	}
	
	return substantialCount >= amm.generationThreshold/2
}

// RecallMemories - 召回相关记忆（带缓存）
func (amm *ActiveMemoryManager) RecallMemories(ctx context.Context, sessionID, query string, categories []memory.MemoryCategory) (*memory.RecallResult, error) {
	if amm.memoryManager == nil {
		return &memory.RecallResult{Items: []*memory.MemoryItem{}}, nil
	}
	
	// 构建缓存键
	cacheKey := fmt.Sprintf("%s:%s:%v", sessionID, query, categories)
	
	// 检查缓存
	if cached := amm.getCachedMemory(cacheKey); cached != nil {
		return cached, nil
	}
	
	// 构建查询
	memoryQuery := &memory.MemoryQuery{
		SessionID:     sessionID,
		Content:       query,
		Categories:    categories,
		MinImportance: 0.3,
		Limit:         10,
		SortBy:        "importance",
	}
	
	// 执行查询
	result := amm.memoryManager.Recall(memoryQuery)
	
	// 缓存结果
	amm.setCachedMemory(cacheKey, result)
	
	return result, nil
}

// 缓存管理
func (amm *ActiveMemoryManager) getCachedMemory(key string) *memory.RecallResult {
	amm.memoryCacheMutex.RLock()
	defer amm.memoryCacheMutex.RUnlock()
	
	if cached, exists := amm.memoryCache[key]; exists {
		return cached
	}
	
	return nil
}

func (amm *ActiveMemoryManager) setCachedMemory(key string, result *memory.RecallResult) {
	amm.memoryCacheMutex.Lock()
	defer amm.memoryCacheMutex.Unlock()
	
	amm.memoryCache[key] = result
	
	// 简单的过期清理
	if len(amm.memoryCache) > 100 {
		// 清理一半
		count := 0
		for k := range amm.memoryCache {
			if count > 50 {
				break
			}
			delete(amm.memoryCache, k)
			count++
		}
	}
}

// GetMemoryTool - 获取记忆生成工具（供工具系统使用）
func (amm *ActiveMemoryManager) GetMemoryTool() *MemoryGenerationTool {
	return &MemoryGenerationTool{
		manager: amm,
	}
}

// MemoryGenerationTool - 记忆生成工具
type MemoryGenerationTool struct {
	manager *ActiveMemoryManager
}

// Name - 工具名称
func (mgt *MemoryGenerationTool) Name() string {
	return "generate_memories"
}

// Description - 工具描述
func (mgt *MemoryGenerationTool) Description() string {
	return "Generate structured key-value memories from the current conversation for future recall"
}

// Execute - 执行工具
func (mgt *MemoryGenerationTool) Execute(ctx context.Context, args map[string]interface{}) (*builtin.ToolResult, error) {
	// 从上下文获取会话ID
	sessionID, ok := ctx.Value(SessionIDKey).(string)
	if !ok {
		return &builtin.ToolResult{
			Content: "session ID not found in context",
			Data:    map[string]interface{}{"error": "session_id_missing"},
		}, nil
	}
	
	// 这里需要获取当前会话的消息
	// 在实际集成时需要传入消息列表
	messages, ok := args["messages"].([]*session.Message)
	if !ok {
		return &builtin.ToolResult{
			Content: "messages not provided",
			Data:    map[string]interface{}{"error": "messages_missing"},
		}, nil
	}
	
	// 生成记忆
	generated, err := mgt.manager.GenerateMemoriesFromConversation(ctx, sessionID, messages)
	if err != nil {
		return &builtin.ToolResult{
			Content: fmt.Sprintf("memory generation failed: %v", err),
			Data:    map[string]interface{}{"error": err.Error()},
		}, nil
	}
	
	return &builtin.ToolResult{
		Content: fmt.Sprintf("Generated %d memories for session %s", len(generated.Memories), sessionID),
		Data: map[string]interface{}{
			"memories_generated": len(generated.Memories),
			"summary":           generated.Summary,
			"session_id":        generated.SessionID,
			"timestamp":         generated.Timestamp,
		},
	}, nil
}

// Parameters - 工具参数模式
func (mgt *MemoryGenerationTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"reason": map[string]interface{}{
				"type":        "string",
				"description": "Reason for generating memories (optional)",
			},
		},
		"required": []string{},
	}
}

// Validate - 验证参数
func (mgt *MemoryGenerationTool) Validate(args map[string]interface{}) error {
	// 这个工具的参数都是可选的，所以不需要严格验证
	// 只需要检查reason参数如果存在的话是否是字符串
	if reason, exists := args["reason"]; exists {
		if _, ok := reason.(string); !ok {
			return fmt.Errorf("reason must be a string")
		}
	}
	return nil
}

// GetMemoryStats - 获取内存统计信息
func (amm *ActiveMemoryManager) GetMemoryStats() map[string]interface{} {
	if amm.memoryManager == nil {
		return map[string]interface{}{
			"short_term": map[string]interface{}{
				"total_items": 0,
				"total_size":  int64(0),
			},
			"long_term": map[string]interface{}{
				"total_items": 0,
				"total_size":  int64(0),
			},
			"total_items": 0,
			"total_size":  int64(0),
		}
	}
	
	return amm.memoryManager.GetMemoryStats()
}

// getProjectID - 获取当前项目ID
func (amm *ActiveMemoryManager) getProjectID() (string, error) {
	// Import utils to get project ID
	return "", fmt.Errorf("project ID not available") // TODO: implement project ID retrieval
}

// storeGeneratedMemoriesLegacy - 存储生成的记忆（向后兼容）
func (amm *ActiveMemoryManager) storeGeneratedMemoriesLegacy(generated *GeneratedMemories) error {
	for _, mem := range generated.Memories {
		// 转换为memory.MemoryItem
		memoryItem := &memory.MemoryItem{
			ID:        fmt.Sprintf("%s_%s_%d", generated.SessionID, mem.Key, time.Now().UnixNano()),
			SessionID: generated.SessionID,
			Type:      memory.LongTermMemory,
			Category:  amm.mapCategory(mem.Category),
			Content:   fmt.Sprintf("%s: %s", mem.Key, mem.Value),
			Metadata: map[string]interface{}{
				"key":             mem.Key,
				"value":           mem.Value,
				"original_category": mem.Category,
				"context":         mem.Context,
				"generation_time": generated.Timestamp,
				"auto_generated":  true,
			},
			Importance:  mem.Importance,
			AccessCount: 0,
			CreatedAt:   generated.Timestamp,
			UpdatedAt:   generated.Timestamp,
			LastAccess:  generated.Timestamp,
			Tags:        mem.Tags,
		}
		
		// 存储到memory系统
		err := amm.Store(memoryItem)
		if err != nil {
			log.Printf("[WARN] Failed to store memory item %s: %v", mem.Key, err)
		}
	}
	
	// 创建会话总结记忆
	if generated.Summary != "" {
		summaryItem := &memory.MemoryItem{
			ID:        fmt.Sprintf("%s_summary_%d", generated.SessionID, generated.Timestamp.UnixNano()),
			SessionID: generated.SessionID,
			Type:      memory.ShortTermMemory,
			Category:  memory.TaskHistory,
			Content:   generated.Summary,
			Metadata: map[string]interface{}{
				"type":            "conversation_summary",
				"memory_count":    len(generated.Memories),
				"generation_time": generated.Timestamp,
				"auto_generated":  true,
			},
			Importance:  0.7,
			AccessCount: 0,
			CreatedAt:   generated.Timestamp,
			UpdatedAt:   generated.Timestamp,
			LastAccess:  generated.Timestamp,
			Tags:        []string{"summary", "auto_generated"},
		}
		
		err := amm.memoryManager.Store(summaryItem)
		if err != nil {
			log.Printf("[WARN] Failed to store summary memory: %v", err)
		}
	}
	
	return nil
}