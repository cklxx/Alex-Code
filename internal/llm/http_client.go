package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// HTTPLLMClient implements the HTTPClient interface for HTTP-based LLM communication
type HTTPLLMClient struct {
	httpClient *http.Client
}

// NewHTTPClient creates a new HTTP-based LLM client
func NewHTTPClient() (*HTTPLLMClient, error) {
	timeout := 120 * time.Second

	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &HTTPLLMClient{
		httpClient: httpClient,
	}, nil
}

// getModelConfig returns the model configuration for the request
func (c *HTTPLLMClient) getModelConfig(req *ChatRequest) (string, string, string) {
	config := req.Config
	if config == nil {
		// Fallback to some defaults if config fails
		return "https://openrouter.ai/api/v1", "sk-default", "deepseek/deepseek-chat-v3-0324:free"
	}

	modelType := req.ModelType
	if modelType == "" {
		modelType = config.DefaultModelType
		if modelType == "" {
			modelType = BasicModel
		}
	}

	// Try to get specific model config first
	if config.Models != nil {
		if modelConfig, exists := config.Models[modelType]; exists {
			return modelConfig.BaseURL, modelConfig.APIKey, modelConfig.Model
		}
	}
	// Fallback to single model config
	return config.BaseURL, config.APIKey, config.Model
}

// Chat sends a chat request and returns the response
func (c *HTTPLLMClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Get model configuration for this request
	baseURL, apiKey, model := c.getModelConfig(req)
	// Ensure streaming is disabled for HTTP mode
	req.Stream = false

	// Set default model if not specified
	if req.Model == "" {
		req.Model = model
	}

	// Set default temperature if not specified
	if req.Temperature == 0 {
		req.Temperature = 0.7 // Default temperature
	}

	// Set default max tokens if not specified
	if req.MaxTokens == 0 {
		req.MaxTokens = 2048 // Default max tokens
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 调试日志：记录请求数据
	if len(jsonData) < 2000 {
		log.Printf("[DEBUG] HTTPLLMClient: Request JSON: %s", string(jsonData))
	} else {
		log.Printf("[DEBUG] HTTPLLMClient: Request JSON (first 1000 chars): %s...", string(jsonData[:1000]))
	}
	log.Printf("[DEBUG] HTTPLLMClient: Request URL: %s/chat/completions", baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[ERROR] HTTPLLMClient: HTTP error %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// 先读取原始响应体用于调试
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 调试日志：记录原始响应
	if len(body) < 1000 {
		log.Printf("[DEBUG] HTTPLLMClient: Raw API response: %s", string(body))
	} else {
		log.Printf("[DEBUG] HTTPLLMClient: Raw API response (first 500 chars): %s...", string(body[:500]))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 调试日志：记录解析后的结构
	log.Printf("[DEBUG] HTTPLLMClient: Parsed response - Choices count: %d", len(chatResp.Choices))
	if len(chatResp.Choices) > 0 {
		log.Printf("[DEBUG] HTTPLLMClient: First choice - Role: %s, Content length: %d", 
			chatResp.Choices[0].Message.Role, len(chatResp.Choices[0].Message.Content))
	}

	return &chatResp, nil
}

// ChatStream is not supported in HTTP mode
func (c *HTTPLLMClient) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamDelta, error) {
	return nil, fmt.Errorf("streaming not supported in HTTP mode, use streaming client instead")
}

// SetHTTPClient sets a custom HTTP client
func (c *HTTPLLMClient) SetHTTPClient(client *http.Client) {
	if client != nil {
		c.httpClient = client
	}
}

// GetHTTPClient returns the current HTTP client
func (c *HTTPLLMClient) GetHTTPClient() *http.Client {
	return c.httpClient
}

// Close closes the client and cleans up resources
func (c *HTTPLLMClient) Close() error {
	// HTTP client doesn't need explicit cleanup
	return nil
}
