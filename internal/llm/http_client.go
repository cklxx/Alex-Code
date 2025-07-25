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
	httpClient       *http.Client
	cacheManager     *CacheManager
	kimiCacheManager *KimiCacheManager
}

// NewHTTPClient creates a new HTTP-based LLM client
func NewHTTPClient() (*HTTPLLMClient, error) {
	timeout := 200 * time.Second

	httpClient := &http.Client{
		Timeout: timeout,
	}

	client := &HTTPLLMClient{
		httpClient:   httpClient,
		cacheManager: GetGlobalCacheManager(),
	}

	// Initialize Kimi cache manager
	client.kimiCacheManager = NewKimiCacheManager(client)

	return client, nil
}

// getModelConfig returns the model configuration for the request
func (c *HTTPLLMClient) getModelConfig(req *ChatRequest) (string, string, string) {
	config := req.Config
	if config == nil {
		// Fallback to global provider if no config in request
		var err error
		config, err = globalConfigProvider()
		if err != nil {
			// Fallback to some defaults if config fails
			return "https://openrouter.ai/api/v1", "sk-default", "deepseek/deepseek-chat-v3-0324:free"
		}
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
func (c *HTTPLLMClient) Chat(ctx context.Context, req *ChatRequest, sessionID string) (*ChatResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Optimize messages using cache
	originalMessages := req.Messages
	if sessionID != "" {
		req.Messages = c.cacheManager.GetOptimizedMessages(sessionID, req.Messages)
	}

	// Get model configuration for this request
	baseURL, apiKey, model := c.getModelConfig(req)
	// Ensure streaming is disabled for HTTP mode
	req.Stream = false
	log.Printf("[DEBUG] API Provider: %s", baseURL)

	// Debug Kimi cache conditions
	if IsKimiAPI(baseURL) {
		log.Printf("[DEBUG] Kimi API detected: baseURL=%s", baseURL)
		log.Printf("[DEBUG] Session ID: '%s', Messages: %d", sessionID, len(req.Messages))
		if sessionID == "" {
			log.Printf("[DEBUG] ❌ Session ID is empty, cache will not be used")
		}
	}

	// Handle Kimi API context caching
	var cacheHeaders map[string]string
	if IsKimiAPI(baseURL) && sessionID != "" && len(req.Messages) > 0 {
		// 尝试为当前的 messages 和 tools 创建或重用缓存
		if _, err := c.kimiCacheManager.CreateCacheIfNeeded(sessionID, req.Messages, req.Tools, apiKey); err != nil {
			log.Printf("WARNING: Failed to create/reuse Kimi cache: %v", err)
		}
		// Prepare headers for cache usage (verifies message/tool consistency)
		cacheHeaders = c.kimiCacheManager.PrepareRequestWithCache(sessionID, req)
		log.Printf("[DEBUG] Cache created: %s", cacheHeaders)

	}

	// Set defaults
	c.setRequestDefaults(req)

	// Override model if not set in request
	if req.Model == "" {
		req.Model = model
	}

	if req.Model == "qwen/qwen3-coder" {
		req.Provider = map[string]interface{}{
			"only": []string{"alibaba"},
		}
	}

	jsonData, err := json.Marshal(req)

	log.Printf("[DEBUG] Request: %s", string(jsonData))

	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	c.setHeaders(httpReq, apiKey, cacheHeaders)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[ERROR] HTTPLLMClient: HTTP error %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response for cache debugging
	if IsKimiAPI(baseURL) {
		log.Printf("[DEBUG] 📥 Received response: %d bytes", len(body))
	}
	log.Printf("[DEBUG] Response: %s", string(body))

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
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

		// Calculate approximate token usage using compatible method
		usage := chatResp.GetUsage()
		tokensUsed := usage.GetTotalTokens()
		if tokensUsed == 0 {
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
func (c *HTTPLLMClient) ChatStream(ctx context.Context, req *ChatRequest, sessionID string) (<-chan StreamDelta, error) {
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

// GetCacheStats returns cache statistics
func (c *HTTPLLMClient) GetCacheStats() map[string]interface{} {
	return c.cacheManager.GetCacheStats()
}

// ClearSessionCache clears cache for a specific session
func (c *HTTPLLMClient) ClearSessionCache(sessionID string) {
	c.cacheManager.ClearCache(sessionID)

	// Also clear Kimi cache if available
	if c.kimiCacheManager != nil {
		// Get current config to determine if cleanup is needed
		config, err := globalConfigProvider()
		if err == nil && IsKimiAPI(config.BaseURL) {
			if err := c.kimiCacheManager.DeleteCache(sessionID, config.APIKey); err != nil {
				log.Printf("WARNING: Failed to clear Kimi cache for session %s: %v", sessionID, err)
			}
		}
	}
}

// ClearKimiCache clears Kimi cache for a specific session
func (c *HTTPLLMClient) ClearKimiCache(sessionID string, apiKey string) error {
	if c.kimiCacheManager != nil {
		return c.kimiCacheManager.DeleteCache(sessionID, apiKey)
	}
	return nil
}

// GetKimiCacheStats returns Kimi cache statistics
func (c *HTTPLLMClient) GetKimiCacheStats() map[string]interface{} {
	if c.kimiCacheManager != nil {
		return c.kimiCacheManager.GetCacheStats()
	}
	return map[string]interface{}{
		"total_caches":   0,
		"active_caches":  0,
		"total_requests": 0,
		"cache_provider": "none",
	}
}

// Close closes the client and cleans up resources
func (c *HTTPLLMClient) Close() error {
	// Clean up Kimi caches only when truly shutting down
	if c.kimiCacheManager != nil {
		config, err := globalConfigProvider()
		if err == nil && IsKimiAPI(config.BaseURL) {
			c.kimiCacheManager.CleanupExpiredCaches(0, config.APIKey) // Force cleanup all
		}
	}

	// HTTP client doesn't need explicit cleanup
	return nil
}

// CleanupCachesOnExit should be called when the CLI application is exiting
func (c *HTTPLLMClient) CleanupCachesOnExit() error {
	if c.kimiCacheManager != nil {
		config, err := globalConfigProvider()
		if err == nil && IsKimiAPI(config.BaseURL) {
			c.kimiCacheManager.CleanupExpiredCaches(0, config.APIKey)
		}
	}
	return nil
}

// setRequestDefaults sets default values for the request
func (c *HTTPLLMClient) setRequestDefaults(req *ChatRequest) {
	if req.Temperature == 0 {
		req.Temperature = 0.7 // Default temperature
	}

	if req.MaxTokens == 0 {
		req.MaxTokens = 2048 // Default max tokens
	}
}

// setHeaders sets common headers for HTTP requests
func (c *HTTPLLMClient) setHeaders(req *http.Request, apiKey string, cacheHeaders map[string]string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Add Kimi cache headers if available
	if len(cacheHeaders) > 0 {
		log.Printf("[KIMI_CACHE] 📤 Sending request with cache headers:")
		for key, value := range cacheHeaders {
			req.Header.Set(key, value)
			log.Printf("[KIMI_CACHE]   %s: %s", key, value)
		}
	}

}
