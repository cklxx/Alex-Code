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
	log.Printf("[DEBUG] parseToolCalls: Processing %d tool calls from LLM", len(message.ToolCalls))
	for i, tc := range message.ToolCalls {
		log.Printf("[DEBUG] parseToolCalls: Tool call %d - ID: '%s', Name: '%s'", i, tc.ID, tc.Function.Name)

		var args map[string]interface{}
		if tc.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				log.Printf("[ERROR] Failed to parse tool arguments: %v", err)
				continue
			}
		}

		// 确保CallID不为空 - 如果缺少则生成一个，但要确保一致性
		callID := tc.ID
		if callID == "" {
			callID = fmt.Sprintf("call_%d", time.Now().UnixNano())
			log.Printf("[WARN] parseToolCalls: Missing ID for tool %s, generated: %s", tc.Function.Name, callID)
			// 重要：更新原始工具调用的ID以保持一致性
			tc.ID = callID
		}

		toolCall := &types.ReactToolCall{
			Name:      tc.Function.Name,
			Arguments: args,
			CallID:    callID,
		}

		log.Printf("[DEBUG] parseToolCalls: Created ReactToolCall - Name: '%s', CallID: '%s'", toolCall.Name, toolCall.CallID)
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

	log.Printf("[DEBUG] executeSerialToolsStream: Starting execution of %d tool calls", len(toolCalls))
	for i, tc := range toolCalls {
		log.Printf("[DEBUG] executeSerialToolsStream: Tool call %d - Name: '%s', CallID: '%s'", i, tc.Name, tc.CallID)
	}

	// 串行执行工具调用，按顺序一个接一个执行
	// 确保为每个输入的工具调用都产生一个对应的结果
	combinedResult := make([]*types.ReactToolResult, 0, len(toolCalls))

	for i, toolCall := range toolCalls {
		log.Printf("[DEBUG] executeSerialToolsStream: Processing tool call %d/%d - Name: '%s', CallID: '%s'", i+1, len(toolCalls), toolCall.Name, toolCall.CallID)

		// 发送工具开始信号
		toolCallStr := te.formatToolCallForDisplay(toolCall.Name, toolCall.Arguments)
		callback(StreamChunk{Type: "tool_start", Content: toolCallStr})

		// 执行工具
		result, err := te.executeTool(ctx, toolCall.Name, toolCall.Arguments, toolCall.CallID)

		// 确保每个工具调用都产生一个结果，无论什么情况
		var finalResult *types.ReactToolResult

		if err != nil {
			log.Printf("[DEBUG] executeSerialToolsStream: Tool call %d failed with error: %v", i+1, err)
			callback(StreamChunk{Type: "tool_error", Content: fmt.Sprintf("%s: %v", toolCall.Name, err)})
			finalResult = &types.ReactToolResult{
				Success:  false,
				Error:    err.Error(),
				ToolName: toolCall.Name,
				ToolArgs: toolCall.Arguments,
				CallID:   toolCall.CallID,
			}
		} else if result != nil {
			log.Printf("[DEBUG] executeSerialToolsStream: Tool call %d succeeded", i+1)
			// 发送工具结果信号
			var contentStr string
			// Use rune-based slicing to properly handle UTF-8 characters like Chinese text
			runes := []rune(result.Content)
			if len(runes) > 200 {
				contentStr = string(runes[:200]) + "..."
			} else {
				contentStr = result.Content
			}
			callback(StreamChunk{Type: "tool_result", Content: contentStr})

			// 确保关键字段都正确设置
			if result.ToolName == "" {
				result.ToolName = toolCall.Name
			}
			if result.CallID == "" {
				result.CallID = toolCall.CallID
			}
			// 确保工具参数也被保存
			if result.ToolArgs == nil {
				result.ToolArgs = toolCall.Arguments
			}

			finalResult = result

			if !result.Success {
				callback(StreamChunk{Type: "tool_error", Content: fmt.Sprintf("%s: %s", toolCall.Name, result.Error)})
			}
		} else {
			// 这种情况不应该发生：err == nil 但 result == nil
			log.Printf("[ERROR] executeSerialToolsStream: Tool call %d returned nil result without error", i+1)
			finalResult = &types.ReactToolResult{
				Success:  false,
				Error:    "tool execution returned nil result",
				ToolName: toolCall.Name,
				ToolArgs: toolCall.Arguments,
				CallID:   toolCall.CallID,
			}
			callback(StreamChunk{Type: "tool_error", Content: fmt.Sprintf("%s: nil result", toolCall.Name)})
		}

		// 确保每个工具调用都有对应的结果
		combinedResult = append(combinedResult, finalResult)
		log.Printf("[DEBUG] executeSerialToolsStream: Added result for tool call %d - CallID: '%s', Success: %v", i+1, finalResult.CallID, finalResult.Success)
	}

	log.Printf("[DEBUG] executeSerialToolsStream: Completed execution - Input: %d tool calls, Output: %d results", len(toolCalls), len(combinedResult))

	// 验证结果数量与输入匹配
	if len(combinedResult) != len(toolCalls) {
		log.Printf("[ERROR] executeSerialToolsStream: Mismatch! Expected %d results, got %d", len(toolCalls), len(combinedResult))
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
			// Use rune-based slicing to properly handle UTF-8 characters like Chinese text
			runes := []rune(v)
			if len(runes) > 50 {
				valueStr = fmt.Sprintf(`"%s..."`, string(runes[:47]))
			} else {
				valueStr = fmt.Sprintf(`"%s"`, v)
			}
		case int, int64, float64, bool:
			valueStr = fmt.Sprintf("%v", v)
		default:
			// For complex types, convert to string and truncate
			str := fmt.Sprintf("%v", v)
			// Use rune-based slicing to properly handle UTF-8 characters like Chinese text
			runes := []rune(str)
			if len(runes) > 30 {
				valueStr = string(runes[:27]) + "..."
			} else {
				valueStr = str
			}
		}
		argParts = append(argParts, fmt.Sprintf("%s=%s", key, valueStr))
	}

	argsStr := strings.Join(argParts, ", ")
	// Use rune-based slicing to properly handle UTF-8 characters like Chinese text
	runes := []rune(argsStr)
	if len(runes) > 100 {
		argsStr = string(runes[:97]) + "..."
	}

	return fmt.Sprintf("%s %s(%s)", greenDot, toolName, argsStr)
}

// executeTool - 执行工具
func (te *ToolExecutor) executeTool(ctx context.Context, toolName string, args map[string]interface{}, callId string) (*types.ReactToolResult, error) {
	log.Printf("[DEBUG] executeTool: Starting execution - Tool: '%s', CallID: '%s'", toolName, callId)

	tool, exists := te.agent.tools[toolName]
	if !exists {
		log.Printf("[ERROR] executeTool: Tool %s not found", toolName)
		return nil, fmt.Errorf("tool %s not found", toolName)
	}

	// 注入工作目录上下文给文件相关工具
	contextWithWorkingDir := te.injectWorkingDirContext(ctx)

	start := time.Now()
	result, err := tool.Execute(contextWithWorkingDir, args)
	duration := time.Since(start)

	if err != nil {
		log.Printf("[ERROR] ToolExecutor: Tool %s execution failed: %v", toolName, err)
		resultObj := &types.ReactToolResult{
			Success:  false,
			Error:    err.Error(),
			Duration: duration,
			ToolName: toolName,
			ToolArgs: args,
			CallID:   callId,
		}
		log.Printf("[DEBUG] executeTool: Error result - CallID: '%s', Success: %v", resultObj.CallID, resultObj.Success)
		return resultObj, nil
	}

	resultObj := &types.ReactToolResult{
		Success:  true,
		Content:  result.Content,
		Data:     result.Data,
		Duration: duration,
		ToolName: toolName,
		ToolArgs: args,
		CallID:   callId,
	}

	log.Printf("[DEBUG] executeTool: Success result - CallID: '%s', Success: %v", resultObj.CallID, resultObj.Success)
	return resultObj, nil
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
