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
