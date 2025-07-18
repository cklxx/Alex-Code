package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"alex/internal/llm"
	"alex/pkg/types"
)

// MCPConfig represents MCP configuration (imported from mcp package)
type MCPConfig struct {
	Enabled         bool                    `json:"enabled"`
	Servers         map[string]*ServerConfig `json:"servers"`
	GlobalTimeout   time.Duration           `json:"global_timeout"`
	AutoRefresh     bool                    `json:"auto_refresh"`
	RefreshInterval time.Duration           `json:"refresh_interval"`
	Security        *SecurityConfig         `json:"security,omitempty"`
	Logging         *LoggingConfig          `json:"logging,omitempty"`
}

// ServerConfig represents MCP server configuration
type ServerConfig struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env"`
	WorkDir     string            `json:"workDir"`
	AutoStart   bool              `json:"autoStart"`
	AutoRestart bool              `json:"autoRestart"`
	Timeout     time.Duration     `json:"timeout"`
	Enabled     bool              `json:"enabled"`
}

// SecurityConfig represents MCP security configuration
type SecurityConfig struct {
	AllowedCommands      []string          `json:"allowed_commands"`
	BlockedCommands      []string          `json:"blocked_commands"`
	AllowedPackages      []string          `json:"allowed_packages"`
	BlockedPackages      []string          `json:"blocked_packages"`
	RequireConfirmation  bool              `json:"require_confirmation"`
	SandboxMode          bool              `json:"sandbox_mode"`
	MaxProcesses         int               `json:"max_processes"`
	MaxMemoryMB          int               `json:"max_memory_mb"`
	AllowedEnvironment   map[string]string `json:"allowed_environment"`
	RestrictedPaths      []string          `json:"restricted_paths"`
}

// LoggingConfig represents MCP logging configuration
type LoggingConfig struct {
	Level       string `json:"level"`
	LogRequests bool   `json:"log_requests"`
	LogResponses bool   `json:"log_responses"`
	LogFile     string `json:"log_file,omitempty"`
}

// Config holds application configuration with multi-model support
type Config struct {
	// Legacy single model config (for backward compatibility)
	APIKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`

	// ReAct agent configuration
	MaxTurns int `json:"max_turns"`

	// Multi-model configurations
	Models map[llm.ModelType]*llm.ModelConfig `json:"models,omitempty"`

	// Default model type to use when none specified
	DefaultModelType llm.ModelType `json:"default_model_type,omitempty"`

	// Tool configuration
	TavilyAPIKey string `json:"tavilyApiKey,omitempty"`

	// MCP configuration
	MCP *MCPConfig `json:"mcp,omitempty"`
}

// Manager handles configuration persistence and retrieval
type Manager struct {
	configPath string
	config     *Config
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".alex-config.json")
	manager := &Manager{
		configPath: configPath,
		config:     getDefaultConfig(),
	}

	// Load existing config if it exists
	if err := manager.load(); err != nil {
		// If config doesn't exist, create default
		if err := manager.save(); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}
	return manager, nil
}

// Get retrieves a configuration value by key
func (m *Manager) Get(key string) (interface{}, error) {
	// Handle nested keys like "models.basic.api_key"
	if strings.Contains(key, ".") {
		return m.getNestedValue(key)
	}

	switch key {
	// Core fields
	case "api_key":
		return m.config.APIKey, nil
	case "base_url":
		return m.config.BaseURL, nil
	case "model":
		return m.config.Model, nil
	case "max_tokens":
		return m.config.MaxTokens, nil
	case "temperature":
		return m.config.Temperature, nil
	case "max_turns":
		return m.config.MaxTurns, nil
	case "default_model_type":
		return m.config.DefaultModelType, nil
	case "models":
		return m.config.Models, nil
	case "tavilyApiKey":
		return m.config.TavilyAPIKey, nil
	case "mcp":
		return m.config.MCP, nil
	default:
		return nil, fmt.Errorf("unknown config key: %s", key)
	}
}

// getNestedValue handles nested key access like "models.basic.api_key"
func (m *Manager) getNestedValue(key string) (interface{}, error) {
	parts := strings.Split(key, ".")

	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid nested key format: %s", key)
	}

	switch parts[0] {
	case "models":
		if len(parts) < 3 {
			return nil, fmt.Errorf("models key requires model type and field: %s", key)
		}

		modelTypeStr := parts[1]
		field := parts[2]
		modelType := llm.ModelType(modelTypeStr)

		if m.config.Models == nil {
			return nil, fmt.Errorf("models configuration not found")
		}

		modelConfig, exists := m.config.Models[modelType]
		if !exists {
			return nil, fmt.Errorf("model type '%s' not found in configuration", modelTypeStr)
		}

		switch field {
		case "api_key":
			return modelConfig.APIKey, nil
		case "base_url":
			return modelConfig.BaseURL, nil
		case "model":
			return modelConfig.Model, nil
		case "temperature":
			return modelConfig.Temperature, nil
		case "max_tokens":
			return modelConfig.MaxTokens, nil
		default:
			return nil, fmt.Errorf("unknown model config field: %s", field)
		}
	default:
		return nil, fmt.Errorf("unknown nested config key: %s", key)
	}
}

// Set updates a configuration value
func (m *Manager) Set(key string, value interface{}) error {
	// Handle nested keys like "models.basic.api_key"
	if strings.Contains(key, ".") {
		return m.setNestedValue(key, value)
	}

	switch key {
	// Core fields
	case "api_key":
		if str, ok := value.(string); ok {
			m.config.APIKey = str
		}
	case "base_url":
		if str, ok := value.(string); ok {
			m.config.BaseURL = str
		}
	case "model":
		if str, ok := value.(string); ok {
			m.config.Model = str
		}
	case "max_tokens":
		if num, ok := value.(int); ok {
			m.config.MaxTokens = num
		}
	case "temperature":
		if temp, ok := value.(float64); ok {
			m.config.Temperature = temp
		}
	case "max_turns":
		if num, ok := value.(int); ok {
			m.config.MaxTurns = num
		}
	case "default_model_type":
		if modelType, ok := value.(llm.ModelType); ok {
			m.config.DefaultModelType = modelType
		}
	case "models":
		if models, ok := value.(map[llm.ModelType]*llm.ModelConfig); ok {
			m.config.Models = models
		}
	case "mcp":
		if mcp, ok := value.(*MCPConfig); ok {
			m.config.MCP = mcp
		}
	case "stream_response", "confidence_threshold", "allowed_tools", "max_concurrency", "tool_timeout", "restricted_paths", "session_timeout", "max_messages_per_session":
		// Legacy fields - ignore for simplified config
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return m.save()
}

// setNestedValue handles nested key setting like "models.basic.api_key"
func (m *Manager) setNestedValue(key string, value interface{}) error {
	parts := strings.Split(key, ".")

	if len(parts) < 2 {
		return fmt.Errorf("invalid nested key format: %s", key)
	}

	switch parts[0] {
	case "models":
		if len(parts) < 3 {
			return fmt.Errorf("models key requires model type and field: %s", key)
		}

		modelTypeStr := parts[1]
		field := parts[2]
		modelType := llm.ModelType(modelTypeStr)

		// Initialize models map if it doesn't exist
		if m.config.Models == nil {
			m.config.Models = make(map[llm.ModelType]*llm.ModelConfig)
		}

		// Initialize model config if it doesn't exist
		if m.config.Models[modelType] == nil {
			m.config.Models[modelType] = &llm.ModelConfig{}
		}

		switch field {
		case "api_key":
			if str, ok := value.(string); ok {
				m.config.Models[modelType].APIKey = str
			} else {
				return fmt.Errorf("api_key must be a string")
			}
		case "base_url":
			if str, ok := value.(string); ok {
				m.config.Models[modelType].BaseURL = str
			} else {
				return fmt.Errorf("base_url must be a string")
			}
		case "model":
			if str, ok := value.(string); ok {
				m.config.Models[modelType].Model = str
			} else {
				return fmt.Errorf("model must be a string")
			}
		case "temperature":
			if temp, ok := value.(float64); ok {
				m.config.Models[modelType].Temperature = temp
			} else {
				return fmt.Errorf("temperature must be a float64")
			}
		case "max_tokens":
			if num, ok := value.(int); ok {
				m.config.Models[modelType].MaxTokens = num
			} else {
				return fmt.Errorf("max_tokens must be an integer")
			}
		default:
			return fmt.Errorf("unknown model config field: %s", field)
		}

		return m.save()
	default:
		return fmt.Errorf("unknown nested config key: %s", key)
	}
}

// GetString returns a string configuration value
func (m *Manager) GetString(key string) string {
	value, err := m.Get(key)
	if err != nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// GetInt returns an integer configuration value
func (m *Manager) GetInt(key string) int {
	value, err := m.Get(key)
	if err != nil {
		return 0
	}
	if num, ok := value.(int); ok {
		return num
	}
	return 0
}

// GetFloat64 returns a float64 configuration value
func (m *Manager) GetFloat64(key string) float64 {
	value, err := m.Get(key)
	if err != nil {
		return 0.0
	}
	if f, ok := value.(float64); ok {
		return f
	}
	return 0.0
}

// GetModelConfig returns the configuration for a specific model type
func (m *Manager) GetModelConfig(modelType llm.ModelType) *llm.ModelConfig {
	// First check multi-model configurations
	if m.config.Models != nil {
		if modelConfig, exists := m.config.Models[modelType]; exists {
			return modelConfig
		}
	}

	// Fallback to default single model config
	return &llm.ModelConfig{
		BaseURL:     m.config.BaseURL,
		Model:       m.config.Model,
		APIKey:      m.config.APIKey,
		Temperature: m.config.Temperature,
		MaxTokens:   m.config.MaxTokens,
	}
}

// GetEffectiveModelType returns the model type to use, defaulting if necessary
func (m *Manager) GetEffectiveModelType(requested llm.ModelType) llm.ModelType {
	if requested != "" {
		return requested
	}
	if m.config.DefaultModelType != "" {
		return m.config.DefaultModelType
	}
	return llm.BasicModel
}

// GetLLMConfig converts the config to LLM package format
func (m *Manager) GetLLMConfig() *llm.Config {
	return &llm.Config{
		// Legacy single model config
		APIKey:      m.config.APIKey,
		BaseURL:     m.config.BaseURL,
		Model:       m.config.Model,
		Temperature: m.config.Temperature,
		MaxTokens:   m.config.MaxTokens,
		Timeout:     5 * time.Minute,

		// Multi-model configurations
		Models:           m.config.Models,
		DefaultModelType: m.config.DefaultModelType,
	}
}

// SetModelConfig sets configuration for a specific model type
func (m *Manager) SetModelConfig(modelType llm.ModelType, config *llm.ModelConfig) error {
	if m.config.Models == nil {
		m.config.Models = make(map[llm.ModelType]*llm.ModelConfig)
	}
	m.config.Models[modelType] = config
	return m.save()
}

// GetMCPConfig returns the MCP configuration
func (m *Manager) GetMCPConfig() *MCPConfig {
	if m.config.MCP == nil {
		m.config.MCP = getDefaultMCPConfig()
	}
	return m.config.MCP
}

// AddServerConfig adds a new MCP server configuration
func (c *MCPConfig) AddServerConfig(serverConfig *ServerConfig) error {
	if c.Servers == nil {
		c.Servers = make(map[string]*ServerConfig)
	}
	c.Servers[serverConfig.ID] = serverConfig
	return nil
}

// ListServerConfigs returns all MCP server configurations
func (c *MCPConfig) ListServerConfigs() []*ServerConfig {
	if c.Servers == nil {
		return nil
	}
	
	configs := make([]*ServerConfig, 0, len(c.Servers))
	for _, config := range c.Servers {
		configs = append(configs, config)
	}
	return configs
}

// ToMCPConfig converts config.MCPConfig to mcp.MCPConfig
func (c *MCPConfig) ToMCPConfig() interface{} {
	// This is a placeholder - we need to import mcp package to do the conversion
	// For now, return self and handle conversion in the calling code
	return c
}

// SetMCPConfig sets the MCP configuration
func (m *Manager) SetMCPConfig(config *MCPConfig) error {
	m.config.MCP = config
	return m.save()
}

// UpdateMCPServerConfig updates or adds an MCP server configuration
func (m *Manager) UpdateMCPServerConfig(serverConfig *ServerConfig) error {
	if m.config.MCP == nil {
		m.config.MCP = getDefaultMCPConfig()
	}
	
	if m.config.MCP.Servers == nil {
		m.config.MCP.Servers = make(map[string]*ServerConfig)
	}
	
	m.config.MCP.Servers[serverConfig.ID] = serverConfig
	return m.save()
}

// RemoveMCPServerConfig removes an MCP server configuration
func (m *Manager) RemoveMCPServerConfig(serverID string) error {
	if m.config.MCP == nil || m.config.MCP.Servers == nil {
		return nil
	}
	
	delete(m.config.MCP.Servers, serverID)
	return m.save()
}

// GetMCPServerConfig returns a specific MCP server configuration
func (m *Manager) GetMCPServerConfig(serverID string) (*ServerConfig, bool) {
	if m.config.MCP == nil || m.config.MCP.Servers == nil {
		return nil, false
	}
	
	config, exists := m.config.MCP.Servers[serverID]
	return config, exists
}

// ListMCPServerConfigs returns all MCP server configurations
func (m *Manager) ListMCPServerConfigs() []*ServerConfig {
	if m.config.MCP == nil || m.config.MCP.Servers == nil {
		return nil
	}
	
	configs := make([]*ServerConfig, 0, len(m.config.MCP.Servers))
	for _, config := range m.config.MCP.Servers {
		configs = append(configs, config)
	}
	return configs
}

// GetConfig returns the complete configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// GetLegacyConfig returns configuration in types.Config format for backward compatibility
func (m *Manager) GetLegacyConfig() (*types.Config, error) {
	return &types.Config{
		DefaultLanguage:       "go",
		OutputFormat:          "text",
		AnalysisDepth:         3,
		MaxTokens:             m.config.MaxTokens,
		Temperature:           m.config.Temperature,
		StreamResponse:        true, // Default value
		SessionTimeout:        30,   // Default value
		RestrictedTools:       []string{},
		AllowedTools:          []string{"file_read", "file_list", "file_update", "bash", "directory_create", "grep", "todo_read", "todo_update"},
		MaxConcurrentTools:    5,    // Default value
		ToolExecutionTimeout:  30,   // Default value
		MaxMessagesPerSession: 1000, // Default value
		// Map simplified fields to legacy format
		APIKey:      m.config.APIKey,
		BaseURL:     m.config.BaseURL,
		Model:       m.config.Model,
		LastUpdated: time.Now(),
	}, nil
}

// load loads configuration from file
func (m *Manager) load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, m.config)
}

// save saves configuration to file
func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.config, "", "    ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// Save is an alias for save for backward compatibility
func (m *Manager) Save() error {
	return m.save()
}

// getDefaultConfig returns the default configuration with DeepSeek models
func getDefaultConfig() *Config {
	return &Config{
		// Legacy single model config (for backward compatibility)
		APIKey:      "sk-replace-with-your-actual-api-key-here-xxxxxxxxxxxxxxx",
		BaseURL:     "https://openrouter.ai/api/v1",
		Model:       "deepseek/deepseek-chat-v3-0324:free",
		MaxTokens:   2048,
		Temperature: 0.7,

		// ReAct agent configuration
		MaxTurns: 25, // 统一设置为25次迭代限制

		// Multi-model configurations - 统一使用相同的配置避免配置不一致
		DefaultModelType: llm.BasicModel,
		Models: map[llm.ModelType]*llm.ModelConfig{
			llm.BasicModel: {
				BaseURL:     "https://openrouter.ai/api/v1",
				Model:       "deepseek/deepseek-chat-v3-0324:free",
				APIKey:      "sk-replace-with-your-actual-api-key-here-xxxxxxxxxxxxxxx",
				Temperature: 0.7,
				MaxTokens:   4000, // 增加token限制以支持复杂任务
			},
			llm.ReasoningModel: {
				BaseURL:     "https://openrouter.ai/api/v1",
				Model:       "deepseek/deepseek-chat-v3-0324:free",
				APIKey:      "sk-replace-with-your-actual-api-key-here-xxxxxxxxxxxxxxx",
				Temperature: 0.3,
				MaxTokens:   4000, // 统一token限制
			},
		},

		// MCP configuration
		MCP: getDefaultMCPConfig(),
	}
}

// getDefaultMCPConfig returns the default MCP configuration
func getDefaultMCPConfig() *MCPConfig {
	return &MCPConfig{
		Enabled:         true,
		Servers:         make(map[string]*ServerConfig),
		GlobalTimeout:   30 * time.Second,
		AutoRefresh:     true,
		RefreshInterval: 5 * time.Minute,
		Security: &SecurityConfig{
			AllowedCommands: []string{
				"npx",
				"node",
				"python",
				"python3",
			},
			BlockedCommands: []string{
				"rm",
				"rmdir",
				"del",
				"format",
				"sudo",
				"su",
			},
			AllowedPackages: []string{
				"@modelcontextprotocol/server-*",
				"mcp-*",
			},
			RequireConfirmation: false,
			SandboxMode:         true,
			MaxProcesses:        10,
			MaxMemoryMB:         512,
			AllowedEnvironment: map[string]string{
				"NODE_ENV": "production",
				"PATH":     "",
			},
			RestrictedPaths: []string{
				"/etc",
				"/var",
				"/usr",
				"/bin",
				"/sbin",
				"/root",
				"/home",
				"/tmp",
				"/System",
				"/Library",
				"/Applications",
				"/Volumes",
			},
		},
		Logging: &LoggingConfig{
			Level:        "info",
			LogRequests:  true,
			LogResponses: false,
			LogFile:      "",
		},
	}
}

// Legacy aliases for backward compatibility

// NewUnifiedConfigManager creates a manager (alias for NewManager)
func NewUnifiedConfigManager() (*Manager, error) {
	return NewManager()
}

// UnifiedConfigManager is an alias for Manager
type UnifiedConfigManager = Manager

// ValidateConfig validates the configuration values
func (m *Manager) ValidateConfig() error {
	config := m.config

	// Validate required fields
	if config.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if config.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if config.Model == "" {
		return fmt.Errorf("model is required")
	}
	if config.MaxTokens < 1 || config.MaxTokens > 1000000 {
		return fmt.Errorf("max_tokens must be between 1 and 100000")
	}

	return nil
}
