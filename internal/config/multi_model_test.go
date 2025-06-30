package config

import (
	"testing"

	"deep-coding-agent/internal/llm"
)

func TestMultiModelConfiguration(t *testing.T) {
	// Create a config manager
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}

	// Test getting LLM config
	llmConfig := manager.GetLLMConfig()
	if llmConfig == nil {
		t.Error("expected LLM config but got nil")
	}

	// Test that DeepSeek configuration is loaded by default
	if len(llmConfig.Models) != 2 {
		t.Errorf("expected 2 model configurations, got %d", len(llmConfig.Models))
	}

	// Test basic model config
	basicConfig := manager.GetModelConfig(llm.BasicModel)
	if basicConfig == nil {
		t.Error("expected basic model config but got nil")
	}
	if basicConfig.BaseURL != "https://openrouter.ai/api/v1" {
		t.Errorf("expected OpenRouter base URL, got %s", basicConfig.BaseURL)
	}
	if basicConfig.Model != "deepseek/deepseek-chat-v3-0324:free" {
		t.Errorf("expected DeepSeek model, got %s", basicConfig.Model)
	}
	if basicConfig.Temperature != 0.7 {
		t.Errorf("expected temperature 0.7 for basic model, got %f", basicConfig.Temperature)
	}

	// Test reasoning model config
	reasoningConfig := manager.GetModelConfig(llm.ReasoningModel)
	if reasoningConfig == nil {
		t.Error("expected reasoning model config but got nil")
	}
	if reasoningConfig.BaseURL != "https://openrouter.ai/api/v1" {
		t.Errorf("expected OpenRouter base URL, got %s", reasoningConfig.BaseURL)
	}
	if reasoningConfig.Model != "deepseek/deepseek-chat-v3-0324:free" {
		t.Errorf("expected DeepSeek model, got %s", reasoningConfig.Model)
	}
	if reasoningConfig.Temperature != 0.3 {
		t.Errorf("expected temperature 0.3 for reasoning model, got %f", reasoningConfig.Temperature)
	}

	// Test effective model type
	effectiveType := manager.GetEffectiveModelType("")
	if effectiveType != llm.BasicModel {
		t.Errorf("expected default model type to be basic, got %s", effectiveType)
	}

	effectiveType = manager.GetEffectiveModelType(llm.ReasoningModel)
	if effectiveType != llm.ReasoningModel {
		t.Errorf("expected reasoning model type, got %s", effectiveType)
	}
}

func TestSetModelConfig(t *testing.T) {
	// Create a config manager
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}

	// Create a custom model config
	customConfig := &llm.ModelConfig{
		BaseURL:     "https://api.custom.com/v1",
		Model:       "custom-model",
		APIKey:      "custom-key",
		Temperature: 0.5,
		MaxTokens:   3000,
	}

	// Set custom config for basic model
	err = manager.SetModelConfig(llm.BasicModel, customConfig)
	if err != nil {
		t.Fatalf("failed to set model config: %v", err)
	}

	// Verify the config was set
	retrievedConfig := manager.GetModelConfig(llm.BasicModel)
	if retrievedConfig.BaseURL != customConfig.BaseURL {
		t.Errorf("expected base URL %s, got %s", customConfig.BaseURL, retrievedConfig.BaseURL)
	}
	if retrievedConfig.Model != customConfig.Model {
		t.Errorf("expected model %s, got %s", customConfig.Model, retrievedConfig.Model)
	}
	if retrievedConfig.APIKey != customConfig.APIKey {
		t.Errorf("expected API key %s, got %s", customConfig.APIKey, retrievedConfig.APIKey)
	}
	if retrievedConfig.Temperature != customConfig.Temperature {
		t.Errorf("expected temperature %f, got %f", customConfig.Temperature, retrievedConfig.Temperature)
	}
}

func TestLegacyConfigCompatibility(t *testing.T) {
	// Create a config manager
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}

	// Test that legacy config methods still work
	legacyConfig, err := manager.GetLegacyConfig()
	if err != nil {
		t.Fatalf("failed to get legacy config: %v", err)
	}

	if legacyConfig == nil {
		t.Error("expected legacy config but got nil")
	}

	// Test that single model config is used as fallback
	manager.config.Models = nil // Clear multi-model configs

	fallbackConfig := manager.GetModelConfig(llm.BasicModel)
	if fallbackConfig.BaseURL != manager.config.BaseURL {
		t.Errorf("expected fallback to single config base URL %s, got %s", manager.config.BaseURL, fallbackConfig.BaseURL)
	}
	if fallbackConfig.Model != manager.config.Model {
		t.Errorf("expected fallback to single config model %s, got %s", manager.config.Model, fallbackConfig.Model)
	}
}
