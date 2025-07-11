package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"alex/internal/llm"
)

// LLMHandler handles all LLM-related operations
type LLMHandler struct {
	streamCallback StreamCallback
}

// NewLLMHandler creates a new LLM handler
func NewLLMHandler(streamCallback StreamCallback) *LLMHandler {
	return &LLMHandler{
		streamCallback: streamCallback,
	}
}

// callLLMWithRetry - 带重试机制的流式LLM调用
func (h *LLMHandler) callLLMWithRetry(ctx context.Context, client llm.Client, request *llm.ChatRequest, maxRetries int) (*llm.ChatResponse, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 使用流式调用
		streamChan, err := client.ChatStream(ctx, request)
		if err != nil {
			lastErr = err
			log.Printf("[WARN] LLMHandler: Stream initialization failed (attempt %d): %v", attempt, err)

			// 检查是否是500错误，如果是，说明请求格式可能有问题，不要重试
			if strings.Contains(err.Error(), "500") {
				log.Printf("[ERROR] LLMHandler: Server error 500, not retrying: %v", err)
				return nil, fmt.Errorf("server error 500 - request format issue: %w", err)
			}

			if attempt < maxRetries {
				backoffDuration := time.Duration(attempt*2) * time.Second
				log.Printf("[WARN] LLMHandler: Retrying in %v", backoffDuration)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(backoffDuration):
					continue
				}
			}
			continue
		}

		// 处理流式响应并重构为完整响应
		response, err := h.collectStreamingResponse(ctx, streamChan)
		if err == nil && response != nil {
			return response, nil
		}

		lastErr = err
		log.Printf("[WARN] LLMHandler: Failed to collect streaming response (attempt %d): %v", attempt, err)

		if attempt < maxRetries {
			backoffDuration := time.Duration(attempt*2) * time.Second
			log.Printf("[WARN] LLMHandler: Retrying in %v", backoffDuration)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoffDuration):
				continue
			}
		}
	}

	return nil, fmt.Errorf("streaming LLM call failed after %d attempts: %w", maxRetries, lastErr)
}

// collectStreamingResponse - 收集流式响应并重构为完整响应
func (h *LLMHandler) collectStreamingResponse(ctx context.Context, streamChan <-chan llm.StreamDelta) (*llm.ChatResponse, error) {
	var response *llm.ChatResponse
	var contentBuilder strings.Builder
	contentBuilder.Grow(8192) // Pre-allocate 8KB for better performance
	var toolCalls []llm.ToolCall
	var currentToolCall *llm.ToolCall

	// 检查是否有流回调需要通知
	hasStreamCallback := h.streamCallback != nil

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case delta, ok := <-streamChan:

			if !ok {
				// 流结束，构建最终响应
				if response == nil {
					return nil, fmt.Errorf("no response received from stream")
				}

				// 设置最终的消息内容
				if len(response.Choices) > 0 {
					response.Choices[0].Message.Content = contentBuilder.String()
					if len(toolCalls) > 0 {
						response.Choices[0].Message.ToolCalls = toolCalls
					}
				}
				return response, nil
			}

			// 初始化响应对象
			if response == nil {
				response = &llm.ChatResponse{
					ID:      delta.ID,
					Object:  delta.Object,
					Created: delta.Created,
					Model:   delta.Model,
					Choices: make([]llm.Choice, 1),
				}
				response.Choices[0] = llm.Choice{
					Index: 0,
					Message: llm.Message{
						Role: "assistant",
					},
				}
			}

			// 处理每个delta中的choice
			if len(delta.Choices) > 0 {
				choice := delta.Choices[0]
				// 处理内容增量
				if choice.Delta.Content != "" {
					contentBuilder.WriteString(choice.Delta.Content)

					// 如果启用流式，实时显示LLM输出内容
					if hasStreamCallback {
						h.streamCallback(StreamChunk{
							Type:     "llm_content",
							Content:  choice.Delta.Content,
							Metadata: map[string]any{"streaming": true},
						})
					}
				}

				// 处理 OpenAI reasoning 字段 (如果存在)
				if hasStreamCallback {
					// 处理 reasoning 字段
					if choice.Delta.Reasoning != "" {
						h.streamCallback(StreamChunk{
							Type:     "reasoning",
							Content:  choice.Delta.Reasoning,
							Metadata: map[string]any{"streaming": true, "source": "openai_reasoning"},
						})
					}

					// 处理 reasoning_summary 字段
					if choice.Delta.ReasoningSummary != "" {
						h.streamCallback(StreamChunk{
							Type:     "reasoning_summary",
							Content:  choice.Delta.ReasoningSummary,
							Metadata: map[string]any{"streaming": true, "source": "openai_reasoning_summary"},
						})
					}

					// 处理 think 字段
					if choice.Delta.Think != "" {
						h.streamCallback(StreamChunk{
							Type:     "think",
							Content:  choice.Delta.Think,
							Metadata: map[string]any{"streaming": true, "source": "openai_think"},
						})
					}
				}

				// 处理工具调用增量
				if len(choice.Delta.ToolCalls) > 0 {
					for _, deltaToolCall := range choice.Delta.ToolCalls {
						// 判断是否为新工具调用：有ID、有函数名、或者函数名与当前工具调用不同
						isNewToolCall := deltaToolCall.ID != "" || 
							deltaToolCall.Function.Name != "" ||
							currentToolCall == nil
						
						if isNewToolCall {
							// 新的工具调用
							toolCallID := deltaToolCall.ID
							if toolCallID == "" {
								// 为空ID生成唯一ID
								toolCallID = fmt.Sprintf("tool_call_%d", len(toolCalls))
							}
							newToolCall := llm.ToolCall{
								ID:   toolCallID,
								Type: deltaToolCall.Type,
								Function: llm.Function{
									Name:      deltaToolCall.Function.Name,
									Arguments: deltaToolCall.Function.Arguments,
								},
							}
							toolCalls = append(toolCalls, newToolCall)
							currentToolCall = &toolCalls[len(toolCalls)-1]
						} else if currentToolCall != nil {
							// 继续现有工具调用（仅当只有arguments且没有函数名时）
							if deltaToolCall.Function.Arguments != "" {
								currentToolCall.Function.Arguments += deltaToolCall.Function.Arguments
							}
						}
					}
				}

				// 检查完成原因
				if choice.FinishReason != "" {
					response.Choices[0].FinishReason = choice.FinishReason
				}
			}

		}
	}
}

// validateLLMRequest - 验证LLM请求参数
func (h *LLMHandler) validateLLMRequest(request *llm.ChatRequest) error {
	if request == nil {
		return fmt.Errorf("request is nil")
	}

	if len(request.Messages) == 0 {
		return fmt.Errorf("no messages in request")
	}

	if request.Config == nil {
		return fmt.Errorf("config is nil")
	}

	return nil
}
