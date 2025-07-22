package agent

import (
	"context"
	"fmt"
	"log"
	"strconv"
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

// isNetworkError checks if an error is network-related and should not be retried
func (h *LLMHandler) isNetworkError(err error) bool {
	errStr := err.Error()

	// Extract HTTP status code if present
	if strings.Contains(errStr, "HTTP error ") {
		// Look for pattern "HTTP error XXX:"
		parts := strings.Split(errStr, "HTTP error ")
		if len(parts) > 1 {
			statusPart := strings.Split(parts[1], ":")
			if len(statusPart) > 0 {
				if statusCode, parseErr := strconv.Atoi(statusPart[0]); parseErr == nil {
					// Network-related HTTP status codes that shouldn't be retried
					switch statusCode {
					case 400, 401, 403, 404, 405, 406, 408, 409, 410, 411, 412, 413, 414, 415, 416, 417, 418, 421, 422, 423, 424, 425, 426, 428, 429, 431, 451:
						// Client errors (4xx) - usually indicate request issues, not transient network problems
						return true
					case 500: // Server error - request format issue
						return true
						// Note: 502, 503, 504 are temporary server issues and should be retried
					}
				}
			}
		}
	}

	// Check for common network error patterns
	networkErrorPatterns := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"dial timeout",
		"read timeout",
		"write timeout",
		"network is unreachable",
		"no route to host",
		"host is down",
		"dns lookup failed",
		"tls handshake timeout",
		"certificate verify failed",
		"ssl handshake failed",
	}

	lowerErr := strings.ToLower(errStr)
	for _, pattern := range networkErrorPatterns {
		if strings.Contains(lowerErr, pattern) {
			return true
		}
	}

	return false
}

// callLLMWithRetry - 带重试机制的非流式LLM调用
func (h *LLMHandler) callLLMWithRetry(ctx context.Context, client llm.Client, request *llm.ChatRequest, maxRetries int) (*llm.ChatResponse, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 使用非流式调用
		response, err := client.Chat(ctx, request)
		if err != nil {
			lastErr = err
			log.Printf("[WARN] LLMHandler: Chat call failed (attempt %d): %v", attempt, err)

			// 检查是否是网络类错误，如果是，不要重试
			if h.isNetworkError(err) {
				log.Printf("[ERROR] LLMHandler: Network error detected, not retrying: %v", err)
				return nil, fmt.Errorf("network error - not retrying: %w", err)
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

		// 直接返回完整响应
		if response != nil {
			// 如果有回调，可以一次性发送完整内容
			if h.streamCallback != nil && len(response.Choices) > 0 {
				h.streamCallback(StreamChunk{
					Type:     "llm_content",
					Content:  response.Choices[0].Message.Content,
					Metadata: map[string]any{"streaming": false},
				})
			}
			return response, nil
		}

		lastErr = fmt.Errorf("received nil response")
		log.Printf("[WARN] LLMHandler: Received nil response (attempt %d)", attempt)

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

	return nil, fmt.Errorf("LLM call failed after %d attempts: %w", maxRetries, lastErr)
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
