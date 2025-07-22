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
	return "Update the entire session todo with free-form content. Content can include goals, tasks, notes in any markdown format."
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

	// Write content directly to todo.md file
	err := os.WriteFile(todoFile, []byte(content), 0644)
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
