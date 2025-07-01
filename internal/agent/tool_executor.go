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

// parseToolCalls - 解析 OpenAI 标准工具调用格式
func (te *ToolExecutor) parseToolCalls(message *llm.Message) []*types.ReactToolCall {
	var toolCalls []*types.ReactToolCall

	// 解析 tool_calls 格式
	for _, tc := range message.ToolCalls {
		var args map[string]interface{}
		if tc.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				log.Printf("[ERROR] Failed to parse tool arguments: %v", err)
				continue
			}
		}

		toolCalls = append(toolCalls, &types.ReactToolCall{
			Name:      tc.Function.Name,
			Arguments: args,
			CallID:    tc.ID,
		})
	}

	return toolCalls
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

	resultChan := make(chan toolResult, len(toolCalls))

	// 启动goroutines并行执行
	for _, tc := range toolCalls {
		// 发送工具开始信号
		toolCallStr := te.formatToolCallForDisplay(tc.Name, tc.Arguments)
		callback(StreamChunk{Type: "tool_start", Content: toolCallStr})

		go func(toolCall *types.ReactToolCall) {
			result, err := te.executeTool(ctx, toolCall.Name, toolCall.Arguments)
			resultChan <- toolResult{
				name:   toolCall.Name,
				result: result,
				err:    err,
				call:   toolCall,
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
	if len(args) == 0 {
		return fmt.Sprintf("%s()", toolName)
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

	return fmt.Sprintf("%s(%s)", toolName, argsStr)
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
