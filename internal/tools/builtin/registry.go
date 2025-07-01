package builtin

import "deep-coding-agent/internal/config"

// GetAllBuiltinTools returns a list of all builtin tools
func GetAllBuiltinTools() []Tool {
	// Create a config manager for tools that need it
	configManager, _ := config.NewManager()

	return []Tool{
		// Thinking and reasoning tools
		NewThinkTool(configManager),

		// Task management tools  
		NewTodoUpdateTool(configManager),
		NewTodoReadTool(configManager),

		// File tools
		CreateFileReadTool(),
		CreateFileWriteTool(),
		CreateFileUpdateTool(),
		CreateFileReplaceTool(),
		CreateFileListTool(),
		CreateDirectoryCreateTool(),

		// Search tools
		CreateGrepTool(),
		CreateRipgrepTool(),
		CreateFindTool(),

		// Web search tools
		CreateWebSearchTool(),
		CreateNewsSearchTool(),
		CreateAcademicSearchTool(),

		// Shell tools
		CreateBashTool(),
		CreateScriptRunnerTool(),
		CreateProcessMonitorTool(),
	}
}

// GetToolByName creates a tool instance by name
func GetToolByName(name string) Tool {
	configManager, _ := config.NewManager()

	switch name {
	case "think":
		return NewThinkTool(configManager)
	case "todo_update":
		return NewTodoUpdateTool(configManager)
	case "todo_read":
		return NewTodoReadTool(configManager)
	case "file_read":
		return CreateFileReadTool()
	case "file_write":
		return CreateFileWriteTool()
	case "file_update":
		return CreateFileUpdateTool()
	case "file_replace":
		return CreateFileReplaceTool()
	case "file_list":
		return CreateFileListTool()
	case "directory_create":
		return CreateDirectoryCreateTool()
	case "grep":
		return CreateGrepTool()
	case "ripgrep":
		return CreateRipgrepTool()
	case "find":
		return CreateFindTool()
	case "web_search":
		return CreateWebSearchTool()
	case "news_search":
		return CreateNewsSearchTool()
	case "academic_search":
		return CreateAcademicSearchTool()
	case "bash":
		return CreateBashTool()
	case "script_runner":
		return CreateScriptRunnerTool()
	case "process_monitor":
		return CreateProcessMonitorTool()
	default:
		return nil
	}
}

// GetToolsByCategory returns tools grouped by category
func GetToolsByCategory() map[string][]Tool {
	configManager, _ := config.NewManager()

	return map[string][]Tool{
		"reasoning": {
			NewThinkTool(configManager),
		},
		"task_management": {
			NewTodoUpdateTool(configManager),
			NewTodoReadTool(configManager),
		},
		"file": {
			CreateFileReadTool(),
			CreateFileWriteTool(),
			CreateFileUpdateTool(),
			CreateFileReplaceTool(),
			CreateFileListTool(),
			CreateDirectoryCreateTool(),
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
			CreateScriptRunnerTool(),
			CreateProcessMonitorTool(),
		},
	}
}
