package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TodoListTool implements todo listing functionality
type TodoListTool struct{}

func CreateTodoListTool() *TodoListTool {
	return &TodoListTool{}
}

func (t *TodoListTool) Name() string {
	return "todo_list"
}

func (t *TodoListTool) Description() string {
	return "List all todo items in the current session."
}

func (t *TodoListTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Filter by status",
				"enum":        []string{"pending", "in_progress", "completed", "cancelled"},
			},
			"priority": map[string]interface{}{
				"type":        "string",
				"description": "Filter by priority",
				"enum":        []string{"low", "medium", "high"},
			},
			"show_completed": map[string]interface{}{
				"type":        "boolean",
				"description": "Include completed todos",
				"default":     false,
			},
		},
	}
}

func (t *TodoListTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddOptionalStringField("status", "Status filter").
		AddOptionalStringField("priority", "Priority filter").
		AddOptionalBooleanField("show_completed", "Show completed todos")

	return validator.Validate(args)
}

func (t *TodoListTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	// Get session directory from context
	sessionDir := getSessionDirectoryFromContext()
	if sessionDir == "" {
		return nil, fmt.Errorf("session directory not found in context")
	}

	todoFile := filepath.Join(sessionDir, "todos.md")

	// Check if todo file exists
	if _, err := os.Stat(todoFile); os.IsNotExist(err) {
		return &ToolResult{
			Content: "No todos found - todo file doesn't exist",
			Data: map[string]interface{}{
				"todos": []map[string]interface{}{},
				"count": 0,
			},
		}, nil
	}

	// Read todo file
	content, err := os.ReadFile(todoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read todo file: %w", err)
	}

	todos := parseTodos(string(content))

	// Apply filters
	statusFilter := ""
	if s, ok := args["status"]; ok {
		statusFilter = s.(string)
	}

	priorityFilter := ""
	if p, ok := args["priority"]; ok {
		priorityFilter = p.(string)
	}

	showCompleted := false
	if sc, ok := args["show_completed"]; ok {
		if scBool, ok := sc.(bool); ok {
			showCompleted = scBool
		}
	}

	var filteredTodos []map[string]interface{}
	for _, todo := range todos {
		// Skip completed todos unless explicitly requested
		if !showCompleted && todo["status"] == "completed" {
			continue
		}

		// Apply status filter
		if statusFilter != "" && todo["status"] != statusFilter {
			continue
		}

		// Apply priority filter
		if priorityFilter != "" && todo["priority"] != priorityFilter {
			continue
		}

		filteredTodos = append(filteredTodos, todo)
	}

	// Build content
	var contentBuilder strings.Builder
	contentBuilder.WriteString(fmt.Sprintf("Found %d todo items:\n\n", len(filteredTodos)))

	for i, todo := range filteredTodos {
		contentBuilder.WriteString(fmt.Sprintf("%d. **%s** (ID: %s)\n", i+1, todo["title"], todo["id"]))
		contentBuilder.WriteString(fmt.Sprintf("   Status: %s | Priority: %s\n", todo["status"], todo["priority"]))

		if description, ok := todo["description"].(string); ok && description != "" {
			contentBuilder.WriteString(fmt.Sprintf("   Description: %s\n", description))
		}

		if tags, ok := todo["tags"].([]string); ok && len(tags) > 0 {
			contentBuilder.WriteString(fmt.Sprintf("   Tags: %s\n", strings.Join(tags, ", ")))
		}

		contentBuilder.WriteString("\n")
	}

	return &ToolResult{
		Content: contentBuilder.String(),
		Data: map[string]interface{}{
			"todos": filteredTodos,
			"count": len(filteredTodos),
			"total": len(todos),
		},
	}, nil
}

func parseTodos(content string) []map[string]interface{} {
	var todos []map[string]interface{}

	// Split by todo sections (## headers)
	sections := strings.Split(content, "## ")

	for _, section := range sections {
		if section == "" {
			continue
		}

		todo := parseTodoSection(section)
		if todo != nil {
			todos = append(todos, todo)
		}
	}

	return todos
}

func parseTodoSection(section string) map[string]interface{} {
	lines := strings.Split(section, "\n")
	if len(lines) == 0 {
		return nil
	}

	todo := make(map[string]interface{})

	// First line is the title
	todo["title"] = strings.TrimSpace(lines[0])

	// Parse fields
	idRegex := regexp.MustCompile(`\*\*ID:\*\*\s*(.+)`)
	priorityRegex := regexp.MustCompile(`\*\*Priority:\*\*\s*(.+)`)
	statusRegex := regexp.MustCompile(`\*\*Status:\*\*\s*(.+)`)
	descriptionRegex := regexp.MustCompile(`\*\*Description:\*\*\s*\n(.+)`)
	tagsRegex := regexp.MustCompile(`\*\*Tags:\*\*\s*(.+)`)

	for _, line := range lines {
		if matches := idRegex.FindStringSubmatch(line); matches != nil {
			todo["id"] = strings.TrimSpace(matches[1])
		}
		if matches := priorityRegex.FindStringSubmatch(line); matches != nil {
			todo["priority"] = strings.TrimSpace(matches[1])
		}
		if matches := statusRegex.FindStringSubmatch(line); matches != nil {
			todo["status"] = strings.TrimSpace(matches[1])
		}
		if matches := tagsRegex.FindStringSubmatch(line); matches != nil {
			tagsStr := strings.TrimSpace(matches[1])
			tags := strings.Split(tagsStr, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
			todo["tags"] = tags
		}
	}

	// Extract description if present
	if matches := descriptionRegex.FindStringSubmatch(section); matches != nil {
		todo["description"] = strings.TrimSpace(matches[1])
	}

	return todo
}
