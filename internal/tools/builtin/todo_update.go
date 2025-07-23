package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NewTodoUpdateTool implements todo update functionality (full replacement)
type NewTodoUpdateTool struct{}

func CreateNewTodoUpdateTool() *NewTodoUpdateTool {
	return &NewTodoUpdateTool{}
}

func (t *NewTodoUpdateTool) Name() string {
	return "todo_update"
}

func (t *NewTodoUpdateTool) Description() string {
	return `Update the entire session todo with free-form content. Content can include goals, tasks, notes in any markdown format.

Example:
# Current Sprint Goals
☐ Fix authentication bug in login module  
☐ Implement user profile API
☒ Update documentation

## Notes
- Bug appears only in production
- API needs validation for email format`
}

func (t *NewTodoUpdateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The complete todo content in markdown format. Can include goals, tasks, notes, etc.",
			},
		},
		"required": []string{"content"},
	}
}

func (t *NewTodoUpdateTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("content", "Todo content")

	return validator.Validate(args)
}

func (t *NewTodoUpdateTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	// Extract content parameter
	content := args["content"].(string)

	// Get sessions directory and ensure it exists
	sessionsDir, err := getSessionsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions directory: %w", err)
	}

	// Get session ID or use default
	var todoFile string
	if id, ok := ctx.Value(SessionIDKey).(string); ok && id != "" {
		// Use session-specific todo file
		todoFile = filepath.Join(sessionsDir, id+"_todo.md")
	} else {
		// Use default session todo file
		todoFile = filepath.Join(sessionsDir, "default_todo.md")
	}

	// Write content directly to todo.md file
	err = os.WriteFile(todoFile, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write todo file: %w", err)
	}

	// Count lines for basic statistics
	lines := strings.Split(content, "\n")
	lineCount := len(lines)

	// Count checkboxes if present
	completedCount := 0
	pendingCount := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "☒") {
			completedCount++
		} else if strings.HasPrefix(line, "☐") {
			pendingCount++
		}
	}

	return &ToolResult{
		Content: content,
		Data: map[string]interface{}{
			"content":         content,
			"line_count":      lineCount,
			"pending_count":   pendingCount,
			"completed_count": completedCount,
			"file_path":       todoFile,
		},
	}, nil
}
