package llm

import (
	"testing"
	"time"
)

func TestNewDefaultClientFactory(t *testing.T) {
	factory := NewDefaultClientFactory()
	if factory == nil {
		t.Error("expected factory but got nil")
	}

	providers := factory.GetSupportedProviders()
	expectedProviders := []string{"openai", "anthropic", "azure", "custom"}

	if len(providers) != len(expectedProviders) {
		t.Errorf("expected %d providers, got %d", len(expectedProviders), len(providers))
	}

	for i, provider := range providers {
		if provider != expectedProviders[i] {
			t.Errorf("expected provider '%s', got '%s'", expectedProviders[i], provider)
		}
	}
}

func TestDefaultClientFactory_CreateHTTPClient(t *testing.T) {
	factory := NewDefaultClientFactory()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "valid config",
			config: &Config{
				BaseURL:     "https://api.test.com",
				APIKey:      "test-key",
				Model:       "test-model",
				Temperature: 0.7,
				MaxTokens:   1000,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := factory.CreateHTTPClient(tt.config)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if client == nil {
				t.Error("expected client but got nil")
			}
		})
	}
}

func TestDefaultClientFactory_CreateStreamingClient(t *testing.T) {
	factory := NewDefaultClientFactory()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "valid config",
			config: &Config{
				BaseURL:     "https://api.test.com",
				APIKey:      "test-key",
				Model:       "test-model",
				Temperature: 0.7,
				MaxTokens:   1000,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := factory.CreateStreamingClient(tt.config)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if client == nil {
				t.Error("expected client but got nil")
			}
		})
	}
}

func TestConfigBuilder(t *testing.T) {
	builder := NewConfigBuilder()
	if builder == nil {
		t.Error("expected builder but got nil")
	}

	config := builder.
		WithAPIKey("test-key").
		WithBaseURL("https://api.test.com").
		WithModel("test-model").
		WithTemperature(0.5).
		WithMaxTokens(2000).
		WithTimeout(60 * time.Second).
		Build()

	if config.APIKey != "test-key" {
		t.Errorf("expected API key 'test-key', got '%s'", config.APIKey)
	}
	if config.BaseURL != "https://api.test.com" {
		t.Errorf("expected base URL 'https://api.test.com', got '%s'", config.BaseURL)
	}
	if config.Model != "test-model" {
		t.Errorf("expected model 'test-model', got '%s'", config.Model)
	}
	if config.Temperature != 0.5 {
		t.Errorf("expected temperature 0.5, got %f", config.Temperature)
	}
	if config.MaxTokens != 2000 {
		t.Errorf("expected max tokens 2000, got %d", config.MaxTokens)
	}
	if config.Timeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", config.Timeout)
	}
}

func TestCreateClientFromProvider(t *testing.T) {
	provider := GetOpenAIConfig()
	apiKey := "test-key"
	model := "gpt-4"

	tests := []struct {
		name        string
		clientType  string
		expectError bool
	}{
		{
			name:        "http client",
			clientType:  "http",
			expectError: false,
		},
		{
			name:        "streaming client",
			clientType:  "streaming",
			expectError: false,
		},
		{
			name:        "unsupported client type",
			clientType:  "websocket",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := CreateClientFromProvider(provider, apiKey, model, tt.clientType)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if client == nil {
				t.Error("expected client but got nil")
			}
		})
	}

	// Test with nil provider
	_, err := CreateClientFromProvider(nil, apiKey, model, "http")
	if err == nil {
		t.Error("expected error with nil provider")
	}
}

func TestGetLLMInstance(t *testing.T) {
	// Set up mock config provider
	mockConfig := &Config{
		BaseURL: "https://api.test.com",
		APIKey:  "test-key",
		Model:   "test-model",
	}
	SetConfigProvider(func() (*Config, error) {
		return mockConfig, nil
	})

	// Clear cache before test
	ClearInstanceCache()

	// Test getting basic model instance
	client, err := GetLLMInstance(BasicModel)
	if err != nil {
		t.Errorf("GetLLMInstance failed: %v", err)
	}
	if client == nil {
		t.Error("expected client but got nil")
	}

	// Test that subsequent calls return cached instance
	client2, err := GetLLMInstance(BasicModel)
	if err != nil {
		t.Errorf("GetLLMInstance failed on second call: %v", err)
	}
	if client != client2 {
		t.Error("expected same cached client instance")
	}

	// Test reasoning model
	reasoningClient, err := GetLLMInstance(ReasoningModel)
	if err != nil {
		t.Errorf("GetLLMInstance failed for reasoning model: %v", err)
	}
	if reasoningClient == nil {
		t.Error("expected reasoning client but got nil")
	}
	if reasoningClient == client {
		t.Error("expected different client instances for different model types")
	}

	// Test error when no config provider is set
	SetConfigProvider(nil)
	_, err = GetLLMInstance(BasicModel)
	if err == nil {
		t.Error("expected error when no config provider is set")
	}
}

func TestGetLLMInstance_ReturnsHTTPClient(t *testing.T) {
	// Set up mock config provider
	mockConfig := &Config{
		BaseURL: "https://api.test.com",
		APIKey:  "test-key",
		Model:   "test-model",
	}
	SetConfigProvider(func() (*Config, error) {
		return mockConfig, nil
	})

	// Clear cache before test
	ClearInstanceCache()

	// Get instance and verify it's an HTTP client
	client, err := GetLLMInstance(BasicModel)
	if err != nil {
		t.Errorf("GetLLMInstance failed: %v", err)
	}

	// Try to cast to HTTPLLMClient to verify it's the right type
	if httpClient, ok := client.(*HTTPLLMClient); !ok {
		t.Error("expected HTTPLLMClient instance")
	} else if httpClient == nil {
		t.Error("expected non-nil HTTPLLMClient")
	}
}
