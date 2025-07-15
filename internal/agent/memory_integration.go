package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"alex/internal/memory"
	"alex/internal/session"
)

// MemoryIntegration 负责Memory系统与会话的集成
type MemoryIntegration struct {
	contextHandler *ContextHandler
}

// NewMemoryIntegration 创建Memory集成器
func NewMemoryIntegration(contextHandler *ContextHandler) *MemoryIntegration {
	return &MemoryIntegration{
		contextHandler: contextHandler,
	}
}

// integrateMemoryIntoMessages - 将memory信息整合到session消息中
func (mi *MemoryIntegration) integrateMemoryIntoMessages(ctx context.Context, sessionMessages []*session.Message) []*session.Message {
	// 检查是否有memory信息需要整合
	memoriesValue := ctx.Value(MemoriesKey)
	if memoriesValue == nil {
		return sessionMessages
	}

	recallResult, ok := memoriesValue.(*memory.RecallResult)
	if !ok || len(recallResult.Items) == 0 {
		return sessionMessages
	}

	// 创建memory消息
	memoryContent := mi.formatMemoryContent(recallResult.Items)
	memoryMessage := &session.Message{
		Role:    "system",
		Content: memoryContent,
		Metadata: map[string]interface{}{
			"type":             "memory_context",
			"memory_items":     len(recallResult.Items),
			"integration_time": time.Now().Unix(),
			"source":           "memory_recall",
		},
		Timestamp: time.Now(),
	}

	// 将memory消息插入到session消息的开头（在系统消息之后）
	integratedMessages := make([]*session.Message, 0, len(sessionMessages)+1)

	// 先添加系统消息（如果有的话）
	var systemAdded bool
	for _, msg := range sessionMessages {
		if msg.Role == "system" && !systemAdded {
			integratedMessages = append(integratedMessages, msg)
			systemAdded = true
			break
		}
	}

	// 添加memory消息
	integratedMessages = append(integratedMessages, memoryMessage)

	// 添加其余的session消息
	for _, msg := range sessionMessages {
		if msg.Role != "system" || !systemAdded {
			integratedMessages = append(integratedMessages, msg)
			if msg.Role == "system" {
				systemAdded = true
			}
		}
	}

	log.Printf("[DEBUG] Integrated %d memory items into %d session messages", len(recallResult.Items), len(sessionMessages))
	return integratedMessages
}

// formatMemoryContent - 格式化memory内容（简化版本，用于整合）
func (mi *MemoryIntegration) formatMemoryContent(memories []*memory.MemoryItem) string {
	if len(memories) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, "## Relevant Context from Memory\n")

	// 按category分组
	categoryGroups := make(map[memory.MemoryCategory][]*memory.MemoryItem)
	for _, mem := range memories {
		categoryGroups[mem.Category] = append(categoryGroups[mem.Category], mem)
	}

	// 格式化每个category
	for category, items := range categoryGroups {
		if len(items) == 0 {
			continue
		}
		categoryName := strings.ToUpper(string(category)[:1]) + string(category)[1:]
		parts = append(parts, fmt.Sprintf("### %s", categoryName))
		for _, item := range items {
			// 限制每个memory项的长度，避免过长
			content := item.Content
			if len(content) > 150 {
				content = content[:150] + "..."
			}
			parts = append(parts, fmt.Sprintf("- %s", content))
		}
		parts = append(parts, "")
	}

	parts = append(parts, "---\n")
	return strings.Join(parts, "\n")
}
