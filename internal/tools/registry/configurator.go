package registry

import (
	"deep-coding-agent/internal/config"
	"deep-coding-agent/internal/tools/builtin"
)

// ToolConfigurator handles configuration of tools with external dependencies
type ToolConfigurator struct {
	configManager *config.Manager
}

// NewToolConfigurator creates a new tool configurator
func NewToolConfigurator(configManager *config.Manager) *ToolConfigurator {
	return &ToolConfigurator{
		configManager: configManager,
	}
}

// ConfigureWebSearchTools configures web search tools with API keys from configuration
func (tc *ToolConfigurator) ConfigureWebSearchTools(registry *Registry) error {
	// Get Tavily API key from configuration
	tavilyAPIKey, err := tc.configManager.Get("tavilyApiKey")
	if err != nil {
		return err
	}

	tavilyAPIKeyStr, ok := tavilyAPIKey.(string)
	if !ok {
		tavilyAPIKeyStr = ""
	}

	// Configure web search tools if they are registered
	webSearchToolNames := []string{"web_search", "news_search", "academic_search"}

	for _, toolName := range webSearchToolNames {
		tool := registry.GetTool(toolName)
		if tool != nil {
			// Check if the tool is a BuiltinToolWrapper and unwrap it
			if wrapper, ok := tool.(*BuiltinToolWrapper); ok {
				if webSearchTool, ok := wrapper.tool.(*builtin.WebSearchTool); ok {
					webSearchTool.SetAPIKey(tavilyAPIKeyStr)
				} else if newsSearchTool, ok := wrapper.tool.(*builtin.NewsSearchTool); ok {
					newsSearchTool.WebSearchTool.SetAPIKey(tavilyAPIKeyStr)
				} else if academicSearchTool, ok := wrapper.tool.(*builtin.AcademicSearchTool); ok {
					academicSearchTool.WebSearchTool.SetAPIKey(tavilyAPIKeyStr)
				}
			}
		}
	}

	return nil
}

// ConfigureAllTools configures all tools that require external configuration
func (tc *ToolConfigurator) ConfigureAllTools(registry *Registry) error {
	// Configure web search tools
	if err := tc.ConfigureWebSearchTools(registry); err != nil {
		return err
	}

	// Add other tool configurations here as needed
	// For example:
	// - AI provider tools
	// - Database connection tools
	// - Cloud service tools

	return nil
}
