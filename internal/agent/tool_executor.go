package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"deep-coding-agent/internal/llm"
	"deep-coding-agent/internal/tools/builtin"
	"deep-coding-agent/pkg/types"
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

	log.Printf("[DEBUG] ToolExecutor: Parsing message with %d standard tool calls, content length: %d", 
		len(message.ToolCalls), len(message.Content))

	// 首先尝试解析标准 tool_calls 格式
	for _, tc := range message.ToolCalls {
		log.Printf("[DEBUG] ToolExecutor: Processing standard tool call: %s (ID: %s)", tc.Function.Name, tc.ID)
		var args map[string]interface{}
		if tc.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				log.Printf("[ERROR] Failed to parse tool arguments: %v", err)
				continue
			}
		}

		toolCall := &types.ReactToolCall{
			Name:      tc.Function.Name,
			Arguments: args,
			CallID:    tc.ID,
		}
		log.Printf("[DEBUG] ToolExecutor: Created tool call: %s with %d args", toolCall.Name, len(toolCall.Arguments))
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
			toolSection := content[startIdx:endIdx+len("<｜tool▁calls▁end｜>")]
			
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
	
	log.Printf("[DEBUG] Parsed %d text tool calls from content", len(toolCalls))
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
	var jsonStart, jsonEnd int = -1, -1
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
	
	log.Printf("[DEBUG] Parsed text tool call: %s with %d args", toolName, len(args))
	return &types.ReactToolCall{
		Name:      toolName,
		Arguments: args,
		CallID:    fmt.Sprintf("text_%d", time.Now().UnixNano()),
	}
}

// executeParallelTools - 并行执行工具调用
func (te *ToolExecutor) executeParallelTools(ctx context.Context, toolCalls []*types.ReactToolCall) *types.ReactToolResult {
	if len(toolCalls) == 0 {
		return &types.ReactToolResult{
			Success: false,
			Error:   "no tool calls provided",
		}
	}

	// 并行执行工具调用（统一处理一个或多个）
	type toolResult struct {
		name   string
		result *types.ReactToolResult
		err    error
	}

	resultChan := make(chan toolResult, len(toolCalls))

	// 启动goroutines并行执行
	for _, tc := range toolCalls {
		go func(toolCall *types.ReactToolCall) {
			result, err := te.executeTool(ctx, toolCall.Name, toolCall.Arguments)
			resultChan <- toolResult{
				name:   toolCall.Name,
				result: result,
				err:    err,
			}
		}(tc)
	}

	// 收集结果
	var results []string
	var errors []string
	var allMetadata []map[string]interface{}
	overallSuccess := true

	for i := 0; i < len(toolCalls); i++ {
		res := <-resultChan

		if res.err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", res.name, res.err))
			overallSuccess = false
		} else if res.result != nil {
			if res.result.Success {
				results = append(results, fmt.Sprintf("%s: %s", res.name, res.result.Content))
				if res.result.Metadata != nil {
					allMetadata = append(allMetadata, res.result.Metadata)
				}
			} else {
				errors = append(errors, fmt.Sprintf("%s: %s", res.name, res.result.Error))
				overallSuccess = false
			}
		}
	}

	// 组合结果
	combinedResult := &types.ReactToolResult{
		Success: overallSuccess,
	}

	if len(results) > 0 {
		combinedResult.Content = strings.Join(results, "\n")
	}

	if len(errors) > 0 {
		if combinedResult.Content != "" {
			combinedResult.Content += "\nErrors:\n"
		}
		combinedResult.Error = strings.Join(errors, "\n")
	}

	if len(allMetadata) > 0 {
		combinedResult.Metadata = map[string]interface{}{
			"parallel_execution": true,
			"tool_count":         len(toolCalls),
			"results":            allMetadata,
		}
	}

	// 保存所有工具调用信息
	combinedResult.ToolCalls = toolCalls

	return combinedResult
}

// executeParallelToolsStream - 并行执行工具调用（流式版本）
func (te *ToolExecutor) executeParallelToolsStream(ctx context.Context, toolCalls []*types.ReactToolCall, callback StreamCallback) *types.ReactToolResult {
	if len(toolCalls) == 0 {
		return &types.ReactToolResult{
			Success: false,
			Error:   "no tool calls provided",
		}
	}

	// 并行执行工具调用（统一处理一个或多个）
	type toolResult struct {
		name   string
		result *types.ReactToolResult
		err    error
		call   *types.ReactToolCall
	}

	// 启动goroutines并行执行，但保持结果的有序显示
	type indexedResult struct {
		toolResult
		index int
	}
	
	indexedResultChan := make(chan indexedResult, len(toolCalls))
	
	for i, tc := range toolCalls {
		go func(toolCall *types.ReactToolCall, index int) {
			// 在goroutine内部发送工具开始信号，避免竞态条件
			toolCallStr := te.formatToolCallForDisplay(toolCall.Name, toolCall.Arguments)
			callback(StreamChunk{Type: "tool_start", Content: toolCallStr})
			
			result, err := te.executeTool(ctx, toolCall.Name, toolCall.Arguments)
			indexedResultChan <- indexedResult{
				toolResult: toolResult{
					name:   toolCall.Name,
					result: result,
					err:    err,
					call:   toolCall,
				},
				index: index,
			}
		}(tc, i)
	}

	// 收集结果并按原始顺序处理
	var results []string
	var errors []string
	var allMetadata []map[string]interface{}
	overallSuccess := true
	
	// 使用数组来存储按顺序的结果
	orderedResults := make([]indexedResult, len(toolCalls))
	resultCount := 0

	for resultCount < len(toolCalls) {
		indexedRes := <-indexedResultChan
		orderedResults[indexedRes.index] = indexedRes
		resultCount++
	}
	
	// 按原始顺序处理结果
	for _, indexedRes := range orderedResults {
		res := indexedRes.toolResult
		
		if res.err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", res.name, res.err))
			callback(StreamChunk{Type: "tool_error", Content: fmt.Sprintf("%s: %v", res.name, res.err)})
			overallSuccess = false
		} else if res.result != nil {
			if res.result.Success {
				results = append(results, fmt.Sprintf("%s: %s", res.name, res.result.Content))
				callback(StreamChunk{Type: "tool_result", Content: res.result.Content})
				if res.result.Metadata != nil {
					allMetadata = append(allMetadata, res.result.Metadata)
				}
			} else {
				errors = append(errors, fmt.Sprintf("%s: %s", res.name, res.result.Error))
				callback(StreamChunk{Type: "tool_error", Content: fmt.Sprintf("%s: %s", res.name, res.result.Error)})
				overallSuccess = false
			}
		}
	}

	// 组合结果
	combinedResult := &types.ReactToolResult{
		Success: overallSuccess,
	}

	if len(results) > 0 {
		combinedResult.Content = strings.Join(results, "\n")
	}

	if len(errors) > 0 {
		if combinedResult.Content != "" {
			combinedResult.Content += "\nErrors:\n"
		}
		combinedResult.Error = strings.Join(errors, "\n")
	}

	if len(allMetadata) > 0 {
		combinedResult.Metadata = map[string]interface{}{
			"parallel_execution": true,
			"tool_count":         len(toolCalls),
			"results":            allMetadata,
		}
	}

	// 保存所有工具调用信息
	combinedResult.ToolCalls = toolCalls

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
func (te *ToolExecutor) executeTool(ctx context.Context, toolName string, args map[string]interface{}) (*types.ReactToolResult, error) {
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
		}, nil
	}

	return &types.ReactToolResult{
		Success:  true,
		Content:  result.Content,
		Data:     result.Data,
		Duration: duration,
		ToolName: toolName,
		ToolArgs: args,
	}, nil
}

// injectWorkingDirContext - 注入工作目录上下文
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
	return builtin.WithWorkingDir(ctx, workingDir)
}
