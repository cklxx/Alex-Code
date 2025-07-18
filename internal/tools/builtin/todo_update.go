package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TodoUpdateTool implements todo update functionality
type TodoUpdateTool struct{}

func CreateTodoUpdateTool() *TodoUpdateTool {
	return &TodoUpdateTool{}
}

func (t *TodoUpdateTool) Name() string {
	return "todo_update"
}

func (t *TodoUpdateTool) Description() string {
	return "Update the status or details of an existing todo item."
}

func (t *TodoUpdateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the todo item to update",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "New status for the todo item",
				"enum":        []string{"pending", "in_progress", "completed", "cancelled"},
			},
			"priority": map[string]interface{}{
				"type":        "string",
				"description": "New priority level",
				"enum":        []string{"low", "medium", "high"},
			},
			"add_note": map[string]interface{}{
				"type":        "string",
				"description": "Add a note to the todo item",
			},
		},
		"required": []string{"id"},
	}
}

func (t *TodoUpdateTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("id", "Todo ID").
		AddOptionalStringField("status", "New status").
		AddOptionalStringField("priority", "New priority").
		AddOptionalStringField("add_note", "Note to add")

	return validator.Validate(args)
}

func (t *TodoUpdateTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	todoID := args["id"].(string)
	
	// Get session directory from context
	sessionDir := getSessionDirectoryFromContext(ctx)
	if sessionDir == "" {
		return nil, fmt.Errorf("session directory not found in context")
	}
	
	todoFile := filepath.Join(sessionDir, "todos.md")
	
	// Read current todo file
	content, err := os.ReadFile(todoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read todo file: %w", err)
	}
	
	todoContent := string(content)
	
	// Find the todo item by ID
	todoIndex := strings.Index(todoContent, fmt.Sprintf("**ID:** %s", todoID))
	if todoIndex == -1 {
		return nil, fmt.Errorf("todo item with ID %s not found", todoID)
	}
	
	// Find the section boundaries
	sectionStart := strings.LastIndex(todoContent[:todoIndex], "## ")
	if sectionStart == -1 {
		return nil, fmt.Errorf("malformed todo file")
	}
	
	sectionEnd := strings.Index(todoContent[todoIndex:], "---")
	if sectionEnd == -1 {
		sectionEnd = len(todoContent) - todoIndex
	} else {
		sectionEnd = todoIndex + sectionEnd + 3 // Include the "---"
	}
	
	todoSection := todoContent[sectionStart:sectionEnd]
	updatedSection := todoSection
	
	// Update status if provided
	if status, ok := args["status"]; ok {
		statusStr := status.(string)
		updatedSection = updateTodoField(updatedSection, "Status", statusStr)
	}
	
	// Update priority if provided
	if priority, ok := args["priority"]; ok {
		priorityStr := priority.(string)
		updatedSection = updateTodoField(updatedSection, "Priority", priorityStr)
	}
	
	// Add note if provided
	if note, ok := args["add_note"]; ok {
		noteStr := note.(string)
		noteSection := fmt.Sprintf("\n**Note (%s):** %s\n", getCurrentTimestampString(), noteStr)
		// Insert before the final "---"
		if strings.HasSuffix(updatedSection, "---\n\n") {
			updatedSection = strings.TrimSuffix(updatedSection, "---\n\n") + noteSection + "---\n\n"
		} else {
			updatedSection += noteSection
		}
	}
	
	// Replace the section in the file
	newContent := todoContent[:sectionStart] + updatedSection + todoContent[sectionEnd:]
	
	// Write back to file
	err = os.WriteFile(todoFile, []byte(newContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write updated todo file: %w", err)
	}
	
	return &ToolResult{
		Content: fmt.Sprintf("Updated todo item: %s", todoID),
		Data: map[string]interface{}{
			"id":      todoID,
			"updated": getCurrentTimestamp(),
		},
	}, nil
}

func updateTodoField(section, fieldName, newValue string) string {
	pattern := fmt.Sprintf("**%s:** ", fieldName)
	startIdx := strings.Index(section, pattern)
	if startIdx == -1 {
		// Field doesn't exist, add it
		return section + fmt.Sprintf("**%s:** %s\n", fieldName, newValue)
	}
	
	startIdx += len(pattern)
	endIdx := strings.Index(section[startIdx:], "\n")
	if endIdx == -1 {
		endIdx = len(section) - startIdx
	}
	
	return section[:startIdx] + newValue + section[startIdx+endIdx:]
}

func getCurrentTimestampString() string {
	return "2024-01-01 12:00:00" // Placeholder
}