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
			apiKeyPreview := modelConfig.APIKey
			// Use rune-based slicing to properly handle UTF-8 characters in API key
			keyRunes := []rune(apiKeyPreview)
			if len(keyRunes) > 15 {
				apiKeyPreview = string(keyRunes[:15]) + "..."
			}
			log.Printf("DEBUG: Using model config - BaseURL: %s, APIKey: %s, Model: %s", modelConfig.BaseURL, apiKeyPreview, modelConfig.Model)
			return modelConfig.BaseURL, modelConfig.APIKey, modelConfig.Model
		}
	}

	// Fallback to single model config
	apiKeyPreview := config.APIKey
	// Use rune-based slicing to properly handle UTF-8 characters in API key
	keyRunes := []rune(apiKeyPreview)
	if len(keyRunes) > 15 {
		apiKeyPreview = string(keyRunes[:15]) + "..."
	}
	log.Printf("DEBUG: Using fallback config - BaseURL: %s, APIKey: %s, Model: %s", config.BaseURL, apiKeyPreview, config.Model)
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
	}

	// Get model configuration for this request
	baseURL, apiKey, model := c.getModelConfig(req)
	// Ensure streaming is disabled for HTTP mode
	req.Stream = false

	// Set defaults
	c.setRequestDefaults(req)

	// Override model if not set in request
	if req.Model == "" {
		req.Model = model
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	log.Printf("DEBUG: Request: %s", string(jsonData))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	c.setHeaders(httpReq, apiKey)

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

	// 先读取原始响应体用于调试
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

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

// ContextKey type for session ID to avoid conflicts
type ContextKey string

const SessionIDKey ContextKey = "session_id"

// ExtractSessionID extracts session ID from context or request (public method)
func (c *HTTPLLMClient) ExtractSessionID(ctx context.Context, req *ChatRequest) string {
	return c.extractSessionID(ctx, req)
}

// extractSessionID extracts session ID from context or request
func (c *HTTPLLMClient) extractSessionID(ctx context.Context, req *ChatRequest) string {
	// Try to get session ID from context using typed key
	if sessionID := ctx.Value(SessionIDKey); sessionID != nil {
		if id, ok := sessionID.(string); ok {
			return id
		}
	}

	// Also try with string key for backward compatibility
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
func (c *HTTPLLMClient) setHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	apiKeyPreview := apiKey
	// Use rune-based slicing to properly handle UTF-8 characters in API key
	keyRunes := []rune(apiKeyPreview)
	if len(keyRunes) > 15 {
		apiKeyPreview = string(keyRunes[:15]) + "..."
	}
	log.Printf("DEBUG: Set Authorization header with key: %s", apiKeyPreview)
}
