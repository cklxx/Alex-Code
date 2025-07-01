package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// HTTPLLMClient implements the HTTPClient interface for HTTP-based LLM communication
type HTTPLLMClient struct {
	httpClient   *http.Client
	cacheManager *CacheManager
}

// NewHTTPClient creates a new HTTP-based LLM client
func NewHTTPClient() (*HTTPLLMClient, error) {
	timeout := 120 * time.Second

	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &HTTPLLMClient{
		httpClient:   httpClient,
		cacheManager: GetGlobalCacheManager(),
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

	// Extract session ID from context or request metadata
	sessionID := c.extractSessionID(ctx, req)
	
	// Optimize messages using cache
	originalMessages := req.Messages
	if sessionID != "" {
		req.Messages = c.cacheManager.GetOptimizedMessages(sessionID, req.Messages)
		log.Printf("[DEBUG] HTTPLLMClient: Message optimization - Original: %d, Optimized: %d", 
			len(originalMessages), len(req.Messages))
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

	// Update cache with the new conversation
	if sessionID != "" && len(chatResp.Choices) > 0 {
		// Prepare messages to cache (original user messages + assistant response)
		newMessages := make([]Message, 0, len(originalMessages)+1)
		
		// Add original user messages (the ones that weren't already cached)
		for _, msg := range originalMessages {
			if msg.Role == "user" {
				newMessages = append(newMessages, msg)
			}
		}
		
		// Add assistant response
		newMessages = append(newMessages, chatResp.Choices[0].Message)
		
		// Calculate approximate token usage
		tokensUsed := 0
		if chatResp.Usage.TotalTokens > 0 {
			tokensUsed = chatResp.Usage.TotalTokens
		} else {
			// Rough estimation: ~4 chars per token
			for _, msg := range newMessages {
				tokensUsed += len(msg.Content) / 4
			}
		}
		
		c.cacheManager.UpdateCache(sessionID, newMessages, tokensUsed)
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

// extractSessionID extracts session ID from context or request
func (c *HTTPLLMClient) extractSessionID(ctx context.Context, req *ChatRequest) string {
	// Try to get session ID from context
	if sessionID := ctx.Value("session_id"); sessionID != nil {
		if id, ok := sessionID.(string); ok {
			return id
		}
	}
	
	// Try to get from request metadata
	for _, msg := range req.Messages {
		if msg.Role == "system" && len(msg.Content) > 0 {
			// Look for session ID in system message
			if strings.Contains(msg.Content, "session_id:") {
				parts := strings.Split(msg.Content, "session_id:")
				if len(parts) > 1 {
					sessionPart := strings.TrimSpace(parts[1])
					if idx := strings.Index(sessionPart, " "); idx != -1 {
						return sessionPart[:idx]
					}
					return sessionPart
				}
			}
		}
	}
	
	return ""
}

// GetCacheStats returns cache statistics
func (c *HTTPLLMClient) GetCacheStats() map[string]interface{} {
	return c.cacheManager.GetCacheStats()
}

// ClearSessionCache clears cache for a specific session
func (c *HTTPLLMClient) ClearSessionCache(sessionID string) {
	c.cacheManager.ClearCache(sessionID)
}

// Close closes the client and cleans up resources
func (c *HTTPLLMClient) Close() error {
	// HTTP client doesn't need explicit cleanup
	return nil
}
