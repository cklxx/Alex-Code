package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// StreamingLLMClient implements the StreamingClient interface for streaming LLM communication
type StreamingLLMClient struct {
	httpClient    *http.Client
	streamEnabled bool
}

// NewStreamingClient creates a new streaming-capable LLM client
func NewStreamingClient() (*StreamingLLMClient, error) {
	timeout := 180 * time.Second // Timeout for streaming

	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &StreamingLLMClient{
		httpClient:    httpClient,
		streamEnabled: true,
	}, nil
}

// getModelConfig returns the model configuration for the request
func (c *StreamingLLMClient) getModelConfig(req *ChatRequest) (string, string, string) {
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

// Chat sends a chat request and returns the response (non-streaming mode)
func (c *StreamingLLMClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Get model configuration for this request
	baseURL, apiKey, model := c.getModelConfig(req)

	// Force non-streaming for regular chat
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

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	c.setHeaders(httpReq, apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp, nil
}

// ChatStream sends a chat request and returns a streaming response
func (c *StreamingLLMClient) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamDelta, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if !c.streamEnabled {
		return nil, fmt.Errorf("streaming is disabled")
	}

	// Get model configuration for this request
	baseURL, apiKey, model := c.getModelConfig(req)

	// Force streaming mode
	req.Stream = true

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

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	c.setHeaders(httpReq, apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	deltaChannel := make(chan StreamDelta, 100)

	go func() {
		defer close(deltaChannel)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			// Parse SSE format: "data: {...}"
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				// Check for end of stream
				if data == "[DONE]" {
					return
				}

				var delta StreamDelta
				if err := json.Unmarshal([]byte(data), &delta); err != nil {
					// Log error but continue processing
					continue
				}

				select {
				case deltaChannel <- delta:
				case <-ctx.Done():
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			// Could send error through channel, but for now just return
			return
		}
	}()

	return deltaChannel, nil
}

// SupportsStreaming returns true if the client supports streaming
func (c *StreamingLLMClient) SupportsStreaming() bool {
	return true
}

// SetStreamingEnabled enables or disables streaming
func (c *StreamingLLMClient) SetStreamingEnabled(enabled bool) {
	c.streamEnabled = enabled
}

// Close closes the client and cleans up resources
func (c *StreamingLLMClient) Close() error {
	// HTTP client doesn't need explicit cleanup
	return nil
}

// setRequestDefaults sets default values for the request
func (c *StreamingLLMClient) setRequestDefaults(req *ChatRequest) {
	if req.Temperature == 0 {
		req.Temperature = 0.7 // Default temperature
	}

	if req.MaxTokens == 0 {
		req.MaxTokens = 2048 // Default max tokens
	}
}

// setHeaders sets common headers for HTTP requests
func (c *StreamingLLMClient) setHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
}
