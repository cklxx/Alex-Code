package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"deep-coding-agent/internal/config"
	"deep-coding-agent/internal/llm"
)

// handleConfigCommand processes config subcommands
func handleConfigCommand(args []string) error {
	if len(args) == 0 {
		return showConfigUsage()
	}

	configManager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	subcommand := args[0]
	switch subcommand {
	case "show", "get":
		return handleConfigShow(configManager, args[1:])
	case "set":
		return handleConfigSet(configManager, args[1:])
	case "list":
		return handleConfigList(configManager)
	case "validate":
		return handleConfigValidate(configManager)
	case "reset":
		return handleConfigReset(configManager)
	default:
		return fmt.Errorf("unknown config subcommand: %s", subcommand)
	}
}

// showConfigUsage displays config command usage
func showConfigUsage() error {
	fmt.Printf(`Deep Coding Agent Configuration Management

USAGE:
    deep-coding-agent config <subcommand> [options]

SUBCOMMANDS:
    show [key]          Show all configuration or specific key
    get <key>           Get specific configuration value  
    set <key> <value>   Set configuration value
    list                List all available configuration keys
    validate            Validate current configuration
    reset               Reset configuration to defaults

CONFIGURATION KEYS:
    api_key             API key for the language model
    base_url            Base URL for API endpoints  
    model               Model name to use
    max_tokens          Maximum tokens for responses (1-100000)
    temperature         Temperature for response generation (0.0-2.0)
    max_turns           Maximum ReAct agent turns (1-20)
    default_model_type  Default model type (basic|reasoning)

NESTED CONFIGURATION KEYS:
    models.basic.api_key        API key for basic model
    models.basic.base_url       Base URL for basic model
    models.basic.model          Model name for basic model
    models.basic.temperature    Temperature for basic model
    models.basic.max_tokens     Max tokens for basic model
    models.reasoning.api_key    API key for reasoning model
    models.reasoning.base_url   Base URL for reasoning model
    models.reasoning.model      Model name for reasoning model
    models.reasoning.temperature Temperature for reasoning model
    models.reasoning.max_tokens Max tokens for reasoning model

EXAMPLES:
    deep-coding-agent config show
    deep-coding-agent config get api_key
    deep-coding-agent config set temperature 0.8
    deep-coding-agent config set model "deepseek/deepseek-chat-v3-0324:free"
    deep-coding-agent config set models.basic.api_key "sk-..."
    deep-coding-agent config set models.reasoning.model "gpt-4"
    deep-coding-agent config get models.basic.temperature
    deep-coding-agent config validate
    deep-coding-agent config reset

`)
	return nil
}

// handleConfigShow shows configuration values
func handleConfigShow(manager *config.Manager, args []string) error {
	if len(args) == 0 {
		// Show all configuration
		config := manager.GetConfig()
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Show specific key
	key := args[0]
	value, err := manager.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get config key '%s': %w", key, err)
	}

	fmt.Printf("%s: %v\n", key, value)
	return nil
}

// handleConfigSet sets configuration values
func handleConfigSet(manager *config.Manager, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("config set requires key and value arguments")
	}

	key := args[0]
	value := strings.Join(args[1:], " ")

	// Type conversion based on key
	var convertedValue interface{}
	
	// Handle nested keys
	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		if len(parts) >= 3 && parts[0] == "models" {
			field := parts[2]
			switch field {
			case "max_tokens":
				intVal, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("max_tokens must be an integer: %w", err)
				}
				if intVal < 1 || intVal > 100000 {
					return fmt.Errorf("max_tokens must be between 1 and 100000")
				}
				convertedValue = intVal
			case "temperature":
				floatVal, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return fmt.Errorf("temperature must be a number: %w", err)
				}
				if floatVal < 0.0 || floatVal > 2.0 {
					return fmt.Errorf("temperature must be between 0.0 and 2.0")
				}
				convertedValue = floatVal
			case "api_key", "base_url", "model":
				if strings.TrimSpace(value) == "" {
					return fmt.Errorf("%s cannot be empty", field)
				}
				convertedValue = value
			default:
				return fmt.Errorf("unknown model configuration field: %s", field)
			}
		} else {
			return fmt.Errorf("unsupported nested configuration key: %s", key)
		}
	} else {
		// Handle top-level keys
		switch key {
		case "max_tokens":
			intVal, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("max_tokens must be an integer: %w", err)
			}
			if intVal < 1 || intVal > 100000 {
				return fmt.Errorf("max_tokens must be between 1 and 100000")
			}
			convertedValue = intVal
		case "max_turns":
			intVal, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("max_turns must be an integer: %w", err)
			}
			if intVal < 1 || intVal > 20 {
				return fmt.Errorf("max_turns must be between 1 and 20")
			}
			convertedValue = intVal
		case "temperature":
			floatVal, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("temperature must be a number: %w", err)
			}
			if floatVal < 0.0 || floatVal > 2.0 {
				return fmt.Errorf("temperature must be between 0.0 and 2.0")
			}
			convertedValue = floatVal
		case "default_model_type":
			modelType := llm.ModelType(value)
			if modelType != llm.BasicModel && modelType != llm.ReasoningModel {
				return fmt.Errorf("default_model_type must be 'basic' or 'reasoning'")
			}
			convertedValue = modelType
		case "api_key", "base_url", "model":
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("%s cannot be empty", key)
			}
			convertedValue = value
		default:
			return fmt.Errorf("unknown configuration key: %s", key)
		}
	}

	if err := manager.Set(key, convertedValue); err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	fmt.Printf("✅ Set %s = %v\n", key, convertedValue)
	return nil
}

// handleConfigList lists all available configuration keys
func handleConfigList(manager *config.Manager) error {
	fmt.Println("Available configuration keys:")
	fmt.Println()

	keys := []struct {
		key         string
		description string
		example     string
	}{
		{"api_key", "API key for the language model", "sk-..."},
		{"base_url", "Base URL for API endpoints", "https://openrouter.ai/api/v1"},
		{"model", "Model name to use", "deepseek/deepseek-chat-v3-0324:free"},
		{"max_tokens", "Maximum tokens for responses", "2048"},
		{"temperature", "Temperature for response generation", "0.7"},
		{"max_turns", "Maximum ReAct agent turns", "3"},
		{"default_model_type", "Default model type", "basic"},
	}

	for _, k := range keys {
		value, _ := manager.Get(k.key)
		fmt.Printf("  %-18s %s\n", k.key, k.description)
		fmt.Printf("  %-18s Current: %v\n", "", value)
		fmt.Printf("  %-18s Example: %s\n", "", k.example)
		fmt.Println()
	}

	return nil
}

// handleConfigValidate validates the current configuration
func handleConfigValidate(manager *config.Manager) error {
	if err := manager.ValidateConfig(); err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
		return err
	}

	fmt.Println("✅ Configuration is valid")
	return nil
}

// handleConfigReset resets configuration to defaults
func handleConfigReset(manager *config.Manager) error {
	fmt.Print("⚠️  This will reset all configuration to defaults. Continue? (y/N): ")

	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Configuration reset cancelled")
		return nil
	}

	// Verify we can create a new manager
	_, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create new config manager: %w", err)
	}

	// Get the current config file path and remove it
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := homeDir + "/.deep-coding-config.json"
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing config: %w", err)
	}

	// Create new config with defaults
	_, err = config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create default config: %w", err)
	}

	fmt.Println("✅ Configuration reset to defaults")
	return nil
}
