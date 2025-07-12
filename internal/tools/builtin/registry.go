package builtin

import (
	"alex/internal/config"
	"alex/internal/session"
)

// GetAllBuiltinTools returns a list of all builtin tools
func GetAllBuiltinTools() []Tool {
	return GetAllBuiltinToolsWithConfig(nil)
}

// GetAllBuiltinToolsWithConfig returns a list of all builtin tools with configuration
func GetAllBuiltinToolsWithConfig(configManager *config.Manager) []Tool {
	// Create a session manager for session-aware tools
	sessionManager, _ := session.NewManager()

	// Create web search tool and configure it if config is available
	webSearchTool := CreateWebSearchTool()
	if configManager != nil {
		if apiKey, err := configManager.Get("tavilyApiKey"); err == nil {
			if apiKeyStr, ok := apiKey.(string); ok && apiKeyStr != "" {
				webSearchTool.SetAPIKey(apiKeyStr)
			}
		}
	}

	return []Tool{
		// Thinking and reasoning tools
		NewThinkTool(),

		// Task management tools (session-aware)
		NewSessionTodoUpdateTool(sessionManager),
		NewSessionTodoReadTool(sessionManager),

		// File tools
		CreateFileReadTool(),
		CreateFileUpdateTool(),
		CreateFileReplaceTool(),
		CreateFileListTool(),

		// Search tools
		CreateGrepTool(),
		CreateRipgrepTool(),

		// Web search tools
		webSearchTool,

		// Shell tools
		CreateBashTool(),
		CreateCodeExecutorTool(),
	}
}

// GetToolByName creates a tool instance by name
func GetToolByName(name string) Tool {
	return GetToolByNameWithConfig(name, nil)
}

// GetToolByNameWithConfig creates a tool instance by name with configuration
func GetToolByNameWithConfig(name string, configManager *config.Manager) Tool {
	sessionManager, _ := session.NewManager()

	switch name {
	case "think":
		return NewThinkTool()
	case "todo_update":
		return NewSessionTodoUpdateTool(sessionManager)
	case "todo_read":
		return NewSessionTodoReadTool(sessionManager)
	case "file_read":
		return CreateFileReadTool()
	case "file_update":
		return CreateFileUpdateTool()
	case "file_replace":
		return CreateFileReplaceTool()
	case "file_list":
		return CreateFileListTool()
	case "grep":
		return CreateGrepTool()
	case "ripgrep":
		return CreateRipgrepTool()
	case "web_search":
		webSearchTool := CreateWebSearchTool()
		if configManager != nil {
			if apiKey, err := configManager.Get("tavilyApiKey"); err == nil {
				if apiKeyStr, ok := apiKey.(string); ok && apiKeyStr != "" {
					webSearchTool.SetAPIKey(apiKeyStr)
				}
			}
		}
		return webSearchTool
	case "bash":
		return CreateBashTool()
	case "code_execute":
		return CreateCodeExecutorTool()
	default:
		return nil
	}
}

// GetToolsByCategory returns tools grouped by category
func GetToolsByCategory() map[string][]Tool {
	return GetToolsByCategoryWithConfig(nil)
}

// GetToolsByCategoryWithConfig returns tools grouped by category with configuration
func GetToolsByCategoryWithConfig(configManager *config.Manager) map[string][]Tool {
	sessionManager, _ := session.NewManager()

	// Create web search tools and configure them if config is available
	webSearchTool := CreateWebSearchTool()

	if configManager != nil {
		if apiKey, err := configManager.Get("tavilyApiKey"); err == nil {
			if apiKeyStr, ok := apiKey.(string); ok && apiKeyStr != "" {
				webSearchTool.SetAPIKey(apiKeyStr)
			}
		}
	}

	return map[string][]Tool{
		"reasoning": {
			NewThinkTool(),
		},
		"task_management": {
			NewSessionTodoUpdateTool(sessionManager),
			NewSessionTodoReadTool(sessionManager),
		},
		"file": {
			CreateFileReadTool(),
			CreateFileUpdateTool(),
			CreateFileReplaceTool(),
			CreateFileListTool(),
		},
		"search": {
			CreateGrepTool(),
			CreateRipgrepTool(),
		},
		"web": {
			webSearchTool,
		},
		"execution": {
			CreateBashTool(),
			CreateCodeExecutorTool(),
		},
	}
}
