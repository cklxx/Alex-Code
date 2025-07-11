package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"alex/internal/llm"
	"alex/internal/tools/builtin"
	"alex/pkg/types"
)

// ToolExecutor - 工具执行器
type ToolExecutor struct {
	agent *ReactAgent
}

// NewToolExecutor - 创建工具执行器
func NewToolExecutor(agent *ReactAgent) *ToolExecutor {
	return &ToolExecutor{agent: agent}
}

// parseToolCalls - 解析 OpenAI 标准工具调用格式和文本格式工具调用
func (te *ToolExecutor) parseToolCalls(message *llm.Message) []*types.ReactToolCall {
	var toolCalls []*types.ReactToolCall

	// 首先尝试解析标准 tool_calls 格式
	for _, tc := range message.ToolCalls {
		var args map[string]interface{}
		if tc.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				log.Printf("[ERROR] Failed to parse tool arguments: %v", err)
				continue
			}
		}

		// Ensure CallID is not empty - generate one if missing
		callID := tc.ID
		if callID == "" {
			callID = fmt.Sprintf("call_%d", time.Now().UnixNano())
			log.Printf("[WARN] parseToolCalls: Missing ID for tool %s, generated: %s", tc.Function.Name, callID)
		}

		toolCall := &types.ReactToolCall{
			Name:      tc.Function.Name,
			Arguments: args,
			CallID:    callID,
		}
		toolCalls = append(toolCalls, toolCall)
	}

	// 如果没有标准工具调用，尝试解析文本格式的工具调用
	if len(toolCalls) == 0 && message.Content != "" {
		textToolCalls := te.parseTextToolCalls(message.Content)
		toolCalls = append(toolCalls, textToolCalls...)
	}

	return toolCalls
}

// parseTextToolCalls - 解析文本格式的工具调用
func (te *ToolExecutor) parseTextToolCalls(content string) []*types.ReactToolCall {
	var toolCalls []*types.ReactToolCall

	// 处理 <｜tool▁calls▁begin｜> 格式
	if strings.Contains(content, "<｜tool▁calls▁begin｜>") {
		// 提取工具调用部分
		startIdx := strings.Index(content, "<｜tool▁calls▁begin｜>")
		endIdx := strings.Index(content, "<｜tool▁calls▁end｜>")

		if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
			toolSection := content[startIdx : endIdx+len("<｜tool▁calls▁end｜>")]

			// 解析每个工具调用
			calls := strings.Split(toolSection, "<｜tool▁call▁begin｜>")
			for i, call := range calls {
				if i == 0 || call == "" {
					continue // 跳过第一个空部分
				}

				// 查找工具调用结束标记
				endCallIdx := strings.Index(call, "<｜tool▁call▁end｜>")
				if endCallIdx == -1 {
					continue
				}

				callContent := call[:endCallIdx]
				if toolCall := te.parseIndividualTextToolCall(callContent); toolCall != nil {
					toolCalls = append(toolCalls, toolCall)
				}
			}
		}
	}

	return toolCalls
}

// parseIndividualTextToolCall - 解析单个文本工具调用
func (te *ToolExecutor) parseIndividualTextToolCall(callContent string) *types.ReactToolCall {
	// 格式: function<｜tool▁sep｜>tool_name\n```json\n{args}\n```
	parts := strings.Split(callContent, "<｜tool▁sep｜>")
	if len(parts) < 2 {
		return nil
	}

	if strings.TrimSpace(parts[0]) != "function" {
		return nil
	}

	remainder := parts[1]
	lines := strings.Split(remainder, "\n")
	if len(lines) == 0 {
		return nil
	}

	toolName := strings.TrimSpace(lines[0])
	if toolName == "" {
		return nil
	}

	// 寻找JSON参数
	jsonStart, jsonEnd := -1, -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "```json" {
			jsonStart = i + 1
		} else if trimmed == "```" && jsonStart != -1 {
			jsonEnd = i
			break
		}
	}

	var args map[string]interface{}
	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonContent := strings.Join(lines[jsonStart:jsonEnd], "\n")
		if err := json.Unmarshal([]byte(jsonContent), &args); err != nil {
			log.Printf("[WARN] Failed to parse JSON args for tool %s: %v", toolName, err)
			// 继续执行，使用空参数
			args = make(map[string]interface{})
		}
	} else {
		args = make(map[string]interface{})
	}

	return &types.ReactToolCall{
		Name:      toolName,
		Arguments: args,
		CallID:    fmt.Sprintf("text_%d", time.Now().UnixNano()),
	}
}

// executeSerialToolsStream - 串行执行工具调用（流式版本）
func (te *ToolExecutor) executeSerialToolsStream(ctx context.Context, toolCalls []*types.ReactToolCall, callback StreamCallback) []*types.ReactToolResult {
	if len(toolCalls) == 0 {
		return []*types.ReactToolResult{
			{
				Success: false,
				Error:   "no tool calls provided",
			},
		}
	}

	// 串行执行工具调用，按顺序一个接一个执行
	combinedResult := []*types.ReactToolResult{}

	for _, toolCall := range toolCalls {
		// 发送工具开始信号
		toolCallStr := te.formatToolCallForDisplay(toolCall.Name, toolCall.Arguments)
		callback(StreamChunk{Type: "tool_start", Content: toolCallStr})

		// 执行工具
		result, err := te.executeTool(ctx, toolCall.Name, toolCall.Arguments, toolCall.CallID)

		// 处理执行结果
		if err != nil {
			callback(StreamChunk{Type: "tool_error", Content: fmt.Sprintf("%s: %v", toolCall.Name, err)})
			// 继续执行下一个工具，不因为一个工具失败而中断整个流程
			combinedResult = append(combinedResult, &types.ReactToolResult{
				Success:  false,
				Error:    err.Error(),
				ToolName: toolCall.Name,
				ToolArgs: toolCall.Arguments,
				CallID:   toolCall.CallID,
			})
		} else if result != nil {
			// 发送工具结果信号
			var contentStr string
			if len(result.Content) > 100 {
				contentStr = result.Content[:100] + "..."
			} else {
				contentStr = result.Content
			}
			callback(StreamChunk{Type: "tool_result", Content: contentStr})

			if result.Success {
				// Ensure ToolName is preserved
				if result.ToolName == "" {
					result.ToolName = toolCall.Name
				}
				combinedResult = append(combinedResult, result)
			} else {
				combinedResult = append(combinedResult, &types.ReactToolResult{
					Success:  false,
					Error:    result.Error,
					ToolName: toolCall.Name,
					ToolArgs: toolCall.Arguments,
					CallID:   toolCall.CallID,
				})
				callback(StreamChunk{Type: "tool_error", Content: fmt.Sprintf("%s: %s", toolCall.Name, result.Error)})
			}
		}
	}

	return combinedResult
}

// formatToolCallForDisplay - 格式化工具调用显示
func (te *ToolExecutor) formatToolCallForDisplay(toolName string, args map[string]interface{}) string {
	// Green color for the dot
	greenDot := "\033[32m⏺\033[0m"

	if len(args) == 0 {
		return fmt.Sprintf("%s %s()", greenDot, toolName)
	}

	// Build arguments string
	var argParts []string
	for key, value := range args {
		var valueStr string
		switch v := value.(type) {
		case string:
			// Truncate long strings and add quotes
			if len(v) > 50 {
				valueStr = fmt.Sprintf(`"%s..."`, v[:47])
			} else {
				valueStr = fmt.Sprintf(`"%s"`, v)
			}
		case int, int64, float64, bool:
			valueStr = fmt.Sprintf("%v", v)
		default:
			// For complex types, convert to string and truncate
			str := fmt.Sprintf("%v", v)
			if len(str) > 30 {
				valueStr = str[:27] + "..."
			} else {
				valueStr = str
			}
		}
		argParts = append(argParts, fmt.Sprintf("%s=%s", key, valueStr))
	}

	argsStr := strings.Join(argParts, ", ")
	if len(argsStr) > 100 {
		argsStr = argsStr[:97] + "..."
	}

	return fmt.Sprintf("%s %s(%s)", greenDot, toolName, argsStr)
}

// executeTool - 执行工具
func (te *ToolExecutor) executeTool(ctx context.Context, toolName string, args map[string]interface{}, callId string) (*types.ReactToolResult, error) {
	tool, exists := te.agent.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", toolName)
	}

	// 注入工作目录上下文给文件相关工具
	contextWithWorkingDir := te.injectWorkingDirContext(ctx)

	start := time.Now()
	result, err := tool.Execute(contextWithWorkingDir, args)
	duration := time.Since(start)

	if err != nil {
		log.Printf("[ERROR] ToolExecutor: Tool %s execution failed: %v", toolName, err)
		return &types.ReactToolResult{
			Success:  false,
			Error:    err.Error(),
			Duration: duration,
			ToolName: toolName,
			ToolArgs: args,
			CallID:   callId,
		}, nil
	}

	return &types.ReactToolResult{
		Success:  true,
		Content:  result.Content,
		Data:     result.Data,
		Duration: duration,
		ToolName: toolName,
		ToolArgs: args,
		CallID:   callId,
	}, nil
}

// injectWorkingDirContext - 注入工作目录上下文和会话ID
func (te *ToolExecutor) injectWorkingDirContext(ctx context.Context) context.Context {
	// 尝试从当前会话获取工作目录
	te.agent.mu.RLock()
	currentSession := te.agent.currentSession
	te.agent.mu.RUnlock()

	var workingDir string

	// 如果有当前会话，从会话的WorkingDir字段获取工作目录
	if currentSession != nil && currentSession.WorkingDir != "" {
		workingDir = currentSession.WorkingDir
	}

	// 如果没有找到工作目录，使用当前工作目录
	if workingDir == "" {
		if wd, err := os.Getwd(); err == nil {
			workingDir = wd
		}
	}

	// 将工作目录注入到context中
	ctx = builtin.WithWorkingDir(ctx, workingDir)

	// 将会话ID注入到context中，供session-aware tools使用
	if currentSession != nil {
		ctx = context.WithValue(ctx, SessionIDKey, currentSession.ID)
	}

	return ctx
}
