package agent

import (
	"fmt"
	"log"
	"strings"
	"time"

	"alex/internal/llm"
	"alex/internal/tools/builtin"
	"alex/pkg/types"
)

// ToolHandler handles tool-related operations
type ToolHandler struct {
	tools map[string]builtin.Tool
}

// NewToolHandler creates a new tool handler
func NewToolHandler(tools map[string]builtin.Tool) *ToolHandler {
	return &ToolHandler{
		tools: tools,
	}
}

// buildToolDefinitions - 构建工具定义列表（包括think工具）
func (h *ToolHandler) buildToolDefinitions() []llm.Tool {
	var tools []llm.Tool

	for _, tool := range h.tools {
		toolDef := llm.Tool{
			Type: "function",
			Function: llm.Function{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			},
		}

		tools = append(tools, toolDef)
	}

	return tools
}

// buildToolMessages - 构建工具结果消息
func (h *ToolHandler) buildToolMessages(actionResult []*types.ReactToolResult, isGemini bool) []llm.Message {
	var toolMessages []llm.Message

	log.Printf("[DEBUG] buildToolMessages: Processing %d tool results", len(actionResult))

	for i, result := range actionResult {
		log.Printf("[DEBUG] buildToolMessages: Result %d - Tool: '%s', CallID: '%s', Success: %v", i, result.ToolName, result.CallID, result.Success)

		content := result.Content
		if !result.Success {
			content = result.Error
		}

		// 确保CallID不为空，这是关键的修复
		callID := result.CallID
		if callID == "" {
			log.Printf("[ERROR] buildToolMessages: Missing CallID for tool %s, generating fallback ID", result.ToolName)
			log.Printf("[ERROR] buildToolMessages: Full result object: %+v", result)
			// 生成一个fallback ID，确保不跳过任何工具结果
			// 这样可以确保每个工具调用都有对应的响应消息
			callID = fmt.Sprintf("fallback_%s_%d", result.ToolName, time.Now().UnixNano())
			log.Printf("[ERROR] buildToolMessages: Generated fallback CallID: %s", callID)
		}

		// Ensure ToolName is not empty and properly formatted for Gemini API
		toolName := result.ToolName

		// Debug logging for Gemini API compatibility
		log.Printf("[DEBUG] buildToolMessages: Creating tool message - Name: '%s', CallID: '%s'", toolName, callID)

		// Gemini API compatibility: ensure tool response format is correct
		// 兼容所有类型的api
		role := "tool"
		if isGemini {
			content = toolName + " executed result: " + content
			role = "user"
		}

		toolMessage := llm.Message{
			Role:       role,
			Content:    content,
			Name:       toolName,
			ToolCallId: callID,
		}

		log.Printf("[DEBUG] buildToolMessages: Created tool message - Role: '%s', ToolCallId: '%s'", toolMessage.Role, toolMessage.ToolCallId)
		toolMessages = append(toolMessages, toolMessage)
	}

	log.Printf("[DEBUG] buildToolMessages: Generated %d tool messages", len(toolMessages))
	return toolMessages
}

// generateObservation - 生成观察结果
func (h *ToolHandler) generateObservation(toolResult []*types.ReactToolResult) string {
	if toolResult == nil {
		return "No tool execution result to observe"
	}

	for _, result := range toolResult {
		if result.Success {
			// 检查是否是特定工具的结果
			if len(result.ToolCalls) > 0 {
				toolName := result.ToolCalls[0].Name
				// 清理工具输出，移除冗余格式信息
				cleanContent := h.cleanToolOutput(result.Content)
				switch toolName {
				case "think":
					return fmt.Sprintf("🧠 Thinking completed: %s", h.truncateContent(cleanContent, 100))
				case "todo_update":
					return fmt.Sprintf("📋 Todo management: %s", h.truncateContent(cleanContent, 100))
				case "file_read":
					return fmt.Sprintf("📖 File read: %s", h.truncateContent(cleanContent, 100))
				case "bash":
					return fmt.Sprintf("⚡ Command executed: %s", h.truncateContent(cleanContent, 100))
				default:
					return fmt.Sprintf("✅ %s completed: %s", toolName, h.truncateContent(cleanContent, 100))
				}
			}
			return fmt.Sprintf("✅ Tool execution successful: %s", h.truncateContent(h.cleanToolOutput(toolResult[0].Content), 100))
		} else {
			return fmt.Sprintf("❌ Tool execution failed: %s", result.Error)
		}
	}
	return "No tool execution result to observe"
}

// cleanToolOutput - 清理工具输出，只保留工具调用格式
func (h *ToolHandler) cleanToolOutput(content string) string {
	lines := strings.Split(content, "\n")
	var cleanLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 只保留🔧工具调用格式的行，其他格式的行都移除
		if strings.HasPrefix(trimmedLine, "🔧 ") {
			cleanLines = append(cleanLines, trimmedLine)
		}
	}

	// 如果没有找到工具调用格式，返回简洁的摘要
	if len(cleanLines) == 0 {
		return h.truncateContent(content, 50)
	}

	return strings.Join(cleanLines, "\n")
}

// truncateContent - 截断内容到指定长度
func (h *ToolHandler) truncateContent(content string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}

	// Use rune-based slicing to properly handle UTF-8 characters like Chinese text
	runes := []rune(content)
	if len(runes) <= maxLen {
		return content
	}

	return string(runes[:maxLen]) + "..."
}
