package builtin

import "alex/internal/config"

// GetAllBuiltinTools returns a list of all builtin tools
func GetAllBuiltinTools() []Tool {
	// Create a config manager for tools that need it
	configManager, _ := config.NewManager()

	return []Tool{
		// Thinking and reasoning tools
		NewThinkTool(),

		// Task management tools
		NewTodoUpdateTool(configManager),
		NewTodoReadTool(configManager),

		// File tools
		CreateFileReadTool(),
		CreateFileUpdateTool(),
		CreateFileReplaceTool(),
		CreateFileListTool(),

		// Search tools
		CreateGrepTool(),

		// Web search tools
		CreateWebSearchTool(),

		// Shell tools
		CreateBashTool(),
		CreateCodeExecutorTool(),
	}
}

// GetToolByName creates a tool instance by name
func GetToolByName(name string) Tool {
	configManager, _ := config.NewManager()

	switch name {
	case "think":
		return NewThinkTool()
	case "todo_update":
		return NewTodoUpdateTool(configManager)
	case "todo_read":
		return NewTodoReadTool(configManager)
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
	case "web_search":
		return CreateWebSearchTool()
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
	configManager, _ := config.NewManager()

	return map[string][]Tool{
		"reasoning": {
			NewThinkTool(),
		},
		"task_management": {
			NewTodoUpdateTool(configManager),
			NewTodoReadTool(configManager),
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
			CreateFindTool(),
		},
		"web": {
			CreateWebSearchTool(),
			CreateNewsSearchTool(),
			CreateAcademicSearchTool(),
		},
		"execution": {
			CreateBashTool(),
			CreateCodeExecutorTool(),
		},
	}
}
