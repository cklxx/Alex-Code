package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// TodoReadTool implements todo reading functionality
type TodoReadTool struct{}

func CreateTodoReadTool() *TodoReadTool {
	return &TodoReadTool{}
}

func (t *TodoReadTool) Name() string {
	return "todo_read"
}

func (t *TodoReadTool) Description() string {
	return "Read the current session's todo list including the final goal and todo items."
}

func (t *TodoReadTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *TodoReadTool) Validate(args map[string]interface{}) error {
	// No validation needed as there are no parameters
	return nil
}

func (t *TodoReadTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	// Try to get session ID from context for session-based storage
	var todoFile string
	if id, ok := ctx.Value(SessionIDKey).(string); ok && id != "" {
		// Session-based storage
		sessionsDir, err := getSessionsDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get sessions directory: %w", err)
		}
		todoFile = filepath.Join(sessionsDir, id+"_todo.md")
	} else {
		// Fallback to working directory
		resolver := GetPathResolverFromContext(ctx)
		workingDir := resolver.workingDir
		if workingDir == "" {
			var err error
			workingDir, err = os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get current working directory: %w", err)
			}
		}
		todoFile = filepath.Join(workingDir, "todo.md")
	}

	// Check if todo file exists
	if _, err := os.Stat(todoFile); os.IsNotExist(err) {
		return &ToolResult{
			Content: "No todo file found. Use todo_update to create one.",
			Data: map[string]interface{}{
				"has_todos":     false,
				"final_goal":    "",
				"todo_items":    []string{},
				"completed":     []string{},
				"pending":       []string{},
				"total_count":   0,
				"pending_count": 0,
			},
		}, nil
	}

	// Read todo file
	content, err := os.ReadFile(todoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read todo file: %w", err)
	}

	return &ToolResult{
		Content: string(content),
		Data: map[string]interface{}{
			"content": string(content),
		},
	}, nil
}
