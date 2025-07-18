package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TodoCreateTool implements todo creation functionality
type TodoCreateTool struct{}

func CreateTodoCreateTool() *TodoCreateTool {
	return &TodoCreateTool{}
}

func (t *TodoCreateTool) Name() string {
	return "todo_create"
}

func (t *TodoCreateTool) Description() string {
	return "Create a new todo item in the session."
}

func (t *TodoCreateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "The title of the todo item",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Detailed description of the todo item",
			},
			"priority": map[string]interface{}{
				"type":        "string",
				"description": "Priority level",
				"enum":        []string{"low", "medium", "high"},
				"default":     "medium",
			},
			"tags": map[string]interface{}{
				"type":        "array",
				"description": "Tags for the todo item",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"title"},
	}
}

func (t *TodoCreateTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("title", "Todo title").
		AddOptionalStringField("description", "Todo description").
		AddOptionalStringField("priority", "Priority level").
		AddOptionalArrayField("tags", "Tags")

	return validator.Validate(args)
}

func (t *TodoCreateTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	title := args["title"].(string)
	
	description := ""
	if d, ok := args["description"]; ok {
		description = d.(string)
	}
	
	priority := "medium"
	if p, ok := args["priority"]; ok {
		priority = p.(string)
	}
	
	var tags []string
	if t, ok := args["tags"]; ok {
		if tagArray, ok := t.([]interface{}); ok {
			for _, tag := range tagArray {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
		}
	}
	
	// Get session directory from context
	sessionDir := getSessionDirectoryFromContext(ctx)
	if sessionDir == "" {
		return nil, fmt.Errorf("session directory not found in context")
	}
	
	// Create todo entry
	todoID := generateTodoID()
	todoEntry := fmt.Sprintf("## %s\n\n**ID:** %s\n**Priority:** %s\n**Status:** pending\n**Created:** %d\n\n", 
		title, todoID, priority, getCurrentTimestamp())
	
	if description != "" {
		todoEntry += fmt.Sprintf("**Description:**\n%s\n\n", description)
	}
	
	if len(tags) > 0 {
		todoEntry += fmt.Sprintf("**Tags:** %s\n\n", strings.Join(tags, ", "))
	}
	
	todoEntry += "---\n\n"
	
	// Append to todo file
	todoFile := filepath.Join(sessionDir, "todos.md")
	file, err := os.OpenFile(todoFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open todo file: %w", err)
	}
	defer file.Close()
	
	_, err = file.WriteString(todoEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to write todo entry: %w", err)
	}
	
	return &ToolResult{
		Content: fmt.Sprintf("Created todo item: %s (ID: %s)", title, todoID),
		Data: map[string]interface{}{
			"id":          todoID,
			"title":       title,
			"description": description,
			"priority":    priority,
			"tags":        tags,
			"status":      "pending",
			"created":     getCurrentTimestamp(),
		},
	}, nil
}

// Helper functions
func getSessionDirectoryFromContext(ctx context.Context) string {
	// This would typically get the session directory from context
	// For now, returning a placeholder
	return "/tmp/alex-session"
}

func generateTodoID() string {
	return fmt.Sprintf("todo_%d", getCurrentTimestamp())
}

func getCurrentTimestamp() int64 {
	return 1642550400 // Placeholder timestamp
}