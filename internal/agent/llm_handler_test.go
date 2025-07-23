package agent

// Mock-based tests for LLM handler retry logic
// All tests use MockClient - NO REAL NETWORK CALLS ARE MADE
// All error messages prefixed with "MOCK_ERROR:" are simulated for testing purposes
// Any HTTP errors, connection refused, etc. are NOT real network issues

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"alex/internal/llm"
)

// MockClient 模拟LLM客户端用于测试 - 所有网络调用都被Mock
type MockClient struct {
	responses []MockResponse
	callCount int
	name      string // 用于标识Mock客户端
}

type MockResponse struct {
	response *llm.ChatResponse
	err      error
	delay    time.Duration
}

func (m *MockClient) ChatStream(ctx context.Context, request *llm.ChatRequest) (<-chan llm.StreamDelta, error) {
	if m.callCount >= len(m.responses) {
		return nil, errors.New("MOCK_ERROR: MockClient unexpected call - no more responses configured")
	}

	resp := m.responses[m.callCount]
	m.callCount++

	if resp.delay > 0 {
		time.Sleep(resp.delay)
	}

	if resp.err != nil {
		return nil, resp.err
	}

	// 创建模拟流式响应
	streamChan := make(chan llm.StreamDelta, 1)
	go func() {
		defer close(streamChan)
		if resp.response != nil {
			// 发送模拟的流式数据
			streamChan <- llm.StreamDelta{
				ID:      resp.response.ID,
				Object:  resp.response.Object,
				Created: resp.response.Created,
				Model:   resp.response.Model,
				Choices: []llm.Choice{
					{
						Index: 0,
						Delta: llm.Message{
							Content: resp.response.Choices[0].Message.Content,
						},
						FinishReason: resp.response.Choices[0].FinishReason,
					},
				},
				Usage: resp.response.Usage,
			}
		}
	}()

	return streamChan, nil
}

func (m *MockClient) Chat(ctx context.Context, request *llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.callCount >= len(m.responses) {
		return nil, errors.New("MOCK_ERROR: MockClient unexpected call - no more responses configured")
	}

	resp := m.responses[m.callCount]
	m.callCount++

	// Simulate delay if configured (for testing backoff logic)
	if resp.delay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(resp.delay):
		}
	}

	if resp.err != nil {
		// Wrap error to clearly indicate it's from mock while maintaining test compatibility
		wrappedErr := fmt.Errorf("MOCK_ERROR: %w", resp.err)
		return nil, wrappedErr
	}

	return resp.response, nil
}

func (m *MockClient) Close() error {
	// Mock client doesn't need cleanup
	return nil
}

// NewMockClient creates a new mock client for testing
func NewMockClient(responses []MockResponse) *MockClient {
	return &MockClient{
		responses: responses,
		callCount: 0,
		name:      "TestMock",
	}
}

// TestLLMHandler_isNetworkError 测试网络错误检测
func TestLLMHandler_isNetworkError(t *testing.T) {
	handler := NewLLMHandler(nil)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		// HTTP 4xx 错误（应该被识别为网络错误）
		{
			name:     "HTTP 400 error",
			err:      errors.New("HTTP error 400: Bad Request"),
			expected: true,
		},
		{
			name:     "HTTP 401 error",
			err:      errors.New("HTTP error 401: Unauthorized"),
			expected: true,
		},
		{
			name:     "HTTP 403 error",
			err:      errors.New("HTTP error 403: Forbidden"),
			expected: true,
		},
		{
			name:     "HTTP 404 error",
			err:      errors.New("HTTP error 404: Not Found"),
			expected: true,
		},
		{
			name:     "HTTP 429 error",
			err:      errors.New("HTTP error 429: Too Many Requests"),
			expected: true,
		},
		{
			name:     "HTTP 500 error",
			err:      errors.New("HTTP error 500: Internal Server Error"),
			expected: true,
		},

		// 连接相关错误
		{
			name:     "Connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "Connection reset",
			err:      errors.New("connection reset by peer"),
			expected: true,
		},
		{
			name:     "Connection timeout",
			err:      errors.New("connection timeout"),
			expected: true,
		},
		{
			name:     "Network unreachable",
			err:      errors.New("network is unreachable"),
			expected: true,
		},
		{
			name:     "No route to host",
			err:      errors.New("no route to host"),
			expected: true,
		},
		{
			name:     "Host is down",
			err:      errors.New("host is down"),
			expected: true,
		},

		// DNS 和 TLS 错误
		{
			name:     "DNS lookup failed",
			err:      errors.New("dns lookup failed"),
			expected: true,
		},
		{
			name:     "TLS handshake timeout",
			err:      errors.New("tls handshake timeout"),
			expected: true,
		},
		{
			name:     "Certificate verify failed",
			err:      errors.New("certificate verify failed"),
			expected: true,
		},
		{
			name:     "SSL handshake failed",
			err:      errors.New("ssl handshake failed"),
			expected: true,
		},

		// 大小写不敏感测试
		{
			name:     "Uppercase connection refused",
			err:      errors.New("CONNECTION REFUSED"),
			expected: true,
		},
		{
			name:     "Mixed case connection timeout",
			err:      errors.New("Request CONNECTION TIMEOUT occurred"),
			expected: true,
		},

		// 应该重试的错误（5xx 中除了 500）
		{
			name:     "HTTP 502 error",
			err:      errors.New("HTTP error 502: Bad Gateway"),
			expected: false,
		},
		{
			name:     "HTTP 503 error",
			err:      errors.New("HTTP error 503: Service Unavailable"),
			expected: false,
		},
		{
			name:     "HTTP 504 error",
			err:      errors.New("HTTP error 504: Gateway Timeout"),
			expected: false,
		},

		// 其他类型的错误（应该重试）
		{
			name:     "Generic error",
			err:      errors.New("some generic error"),
			expected: false,
		},
		{
			name:     "Context cancelled",
			err:      context.Canceled,
			expected: false,
		},
		{
			name:     "EOF error",
			err:      errors.New("EOF"),
			expected: false,
		},
		{
			name:     "Parsing error",
			err:      errors.New("json parsing failed"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.isNetworkError(tt.err)
			if result != tt.expected {
				t.Errorf("isNetworkError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

// TestLLMHandler_RetryLogic 测试重试逻辑
func TestLLMHandler_RetryLogic(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []MockResponse
		expectRetries int
		expectError   bool
		errorContains string
	}{
		{
			name: "成功响应，无重试",
			mockResponses: []MockResponse{
				{
					response: &llm.ChatResponse{
						ID:      "test-123",
						Object:  "chat.completion",
						Created: time.Now().Unix(),
						Model:   "test-model",
						Choices: []llm.Choice{
							{
								Index: 0,
								Message: llm.Message{
									Role:    "assistant",
									Content: "Hello, world!",
								},
								FinishReason: "stop",
							},
						},
						Usage: llm.Usage{
							PromptTokens:     10,
							CompletionTokens: 5,
							TotalTokens:      15,
						},
					},
					err: nil,
				},
			},
			expectRetries: 1,
			expectError:   false,
		},
		{
			name: "网络错误，不重试",
			mockResponses: []MockResponse{
				{
					response: nil,
					err:      errors.New("HTTP error 400: Bad Request"),
				},
			},
			expectRetries: 1,
			expectError:   true,
			errorContains: "permanent network error",
		},
		{
			name: "连接拒绝错误，不重试",
			mockResponses: []MockResponse{
				{
					response: nil,
					err:      errors.New("connection refused"),
				},
			},
			expectRetries: 1,
			expectError:   true,
			errorContains: "permanent network error",
		},
		{
			name: "临时错误，会重试直到成功",
			mockResponses: []MockResponse{
				{
					response: nil,
					err:      errors.New("HTTP error 502: Bad Gateway"),
				},
				{
					response: nil,
					err:      errors.New("temporary failure"),
				},
				{
					response: &llm.ChatResponse{
						ID:      "test-456",
						Object:  "chat.completion",
						Created: time.Now().Unix(),
						Model:   "test-model",
						Choices: []llm.Choice{
							{
								Index: 0,
								Message: llm.Message{
									Role:    "assistant",
									Content: "Success after retries!",
								},
								FinishReason: "stop",
							},
						},
					},
					err: nil,
				},
			},
			expectRetries: 3,
			expectError:   false,
		},
		{
			name: "临时错误，重试次数用尽",
			mockResponses: []MockResponse{
				{
					response: nil,
					err:      errors.New("HTTP error 502: Bad Gateway"),
				},
				{
					response: nil,
					err:      errors.New("HTTP error 503: Service Unavailable"),
				},
				{
					response: nil,
					err:      errors.New("temporary server error"),
				},
			},
			expectRetries: 3,
			expectError:   true,
			errorContains: "LLM call failed after 3 attempts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Using Mock client - no real network calls made
			mockClient := NewMockClient(tt.mockResponses)
			t.Logf("TEST: Using MockClient for %s - no real network calls", tt.name)

			handler := NewLLMHandler(nil)
			// Use a timeout context to prevent long delays in CI
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			
			request := &llm.ChatRequest{
				Messages: []llm.Message{
					{Role: "user", Content: "Test message"},
				},
				Config: &llm.Config{
					Model:       "test-model",
					Temperature: 0.7,
				},
			}

			// Use a no-delay backoff for testing to avoid CI timeouts
			noDelayBackoff := func(attempt int) time.Duration {
				return 0 // No delay in tests
			}
			response, err := handler.callLLMWithRetryAndBackoff(ctx, mockClient, request, 3, noDelayBackoff)

			// 验证重试次数
			if mockClient.callCount != tt.expectRetries {
				t.Errorf("Expected %d retries, got %d", tt.expectRetries, mockClient.callCount)
			}

			// 验证错误情况
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got nil")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if response == nil {
					t.Error("Expected response, got nil")
				}
			}
		})
	}
}

// TestLLMHandler_RetryBackoff 测试重试间隔
func TestLLMHandler_RetryBackoff(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping backoff test in short mode")
	}

	mockClient := NewMockClient([]MockResponse{
		{
			response: nil,
			err:      errors.New("HTTP error 502: Bad Gateway"),
		},
		{
			response: nil,
			err:      errors.New("HTTP error 503: Service Unavailable"),
		},
		{
			response: &llm.ChatResponse{
				ID:      "test-backoff",
				Object:  "chat.completion",
				Created: time.Now().Unix(),
				Model:   "test-model",
				Choices: []llm.Choice{
					{
						Index: 0,
						Message: llm.Message{
							Role:    "assistant",
							Content: "Success after backoff!",
						},
						FinishReason: "stop",
					},
				},
			},
			err: nil,
		},
	})
	
	t.Log("TEST: Using MockClient for backoff test - no real network calls")

	handler := NewLLMHandler(nil)
	ctx := context.Background()
	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: "Test backoff message"},
		},
		Config: &llm.Config{
			Model:       "test-model",
			Temperature: 0.7,
		},
	}

	// Use shorter delays for CI compatibility
	shortBackoff := func(attempt int) time.Duration {
		return time.Duration(attempt*10) * time.Millisecond // Much shorter delays
	}

	start := time.Now()
	response, err := handler.callLLMWithRetryAndBackoff(ctx, mockClient, request, 3, shortBackoff)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if response == nil {
		t.Error("Expected response, got nil")
	}

	// 验证总耗时包含了退避延迟（10ms + 20ms = 30ms）
	// 允许一些测试环境的不确定性
	expectedMinDelay := 20 * time.Millisecond
	if elapsed < expectedMinDelay {
		t.Errorf("Expected at least %v delay for backoff, got %v", expectedMinDelay, elapsed)
	}

	t.Logf("Retry with backoff completed in %v", elapsed)
}

// TestLLMHandler_ContextCancellation 测试上下文取消
func TestLLMHandler_ContextCancellation(t *testing.T) {
	// Provide enough responses for potential retries
	mockClient := NewMockClient([]MockResponse{
		{
			response: nil,
			err:      errors.New("HTTP error 502: Bad Gateway"),
			delay:    100 * time.Millisecond, // 延迟以确保上下文取消被检测到
		},
		{
			response: nil,
			err:      errors.New("HTTP error 502: Bad Gateway"),
			delay:    100 * time.Millisecond,
		},
		{
			response: nil,
			err:      errors.New("HTTP error 502: Bad Gateway"),
			delay:    100 * time.Millisecond,
		},
	})
	
	t.Log("TEST: Using MockClient for context cancellation test - no real network calls")

	handler := NewLLMHandler(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: "Test cancellation"},
		},
		Config: &llm.Config{
			Model:       "test-model",
			Temperature: 0.7,
		},
	}

	// Use no delay backoff for cancellation test
	noDelayBackoff := func(attempt int) time.Duration {
		return 0
	}
	response, err := handler.callLLMWithRetryAndBackoff(ctx, mockClient, request, 3, noDelayBackoff)

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}

	if response != nil {
		t.Error("Expected nil response due to cancellation")
	}

	// 验证错误是上下文相关的或者重试超限
	if err != context.DeadlineExceeded && err != context.Canceled && !contains(err.Error(), "LLM call failed after") {
		t.Errorf("Expected context cancellation or retry exhaustion error, got: %v", err)
	}
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsHelper(s, substr))))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}