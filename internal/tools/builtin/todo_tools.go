package builtin

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"alex/internal/config"
	"alex/pkg/types"
)

// TodoUpdateTool handles todo management operations (create, update, complete, delete)
type TodoUpdateTool struct {
	configManager *config.Manager
}

func NewTodoUpdateTool(configManager *config.Manager) *TodoUpdateTool {
	return &TodoUpdateTool{
		configManager: configManager,
	}
}

func (t *TodoUpdateTool) Name() string {
	return "todo_update"
}

func (t *TodoUpdateTool) Description() string {
	return "Create, update, complete, and delete todo tasks. Supports batch operations and task management."
}

func (t *TodoUpdateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action to perform: create, create_batch, update, complete, delete, set_progress",
				"enum":        []string{"create", "create_batch", "update", "complete", "delete", "set_progress"},
			},
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Todo item ID (required for update, complete, delete)",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Todo content (required for create, optional for update)",
			},
			"tasks": map[string]interface{}{
				"type":        "array",
				"description": "Array of tasks for batch creation (required for create_batch)",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"content": map[string]interface{}{
							"type":        "string",
							"description": "Task content",
						},
						"order": map[string]interface{}{
							"type":        "integer",
							"description": "Execution order (1, 2, 3...). Lower numbers execute first",
							"minimum":     1,
							"default":     1,
						},
					},
					"required": []string{"content"},
				},
			},
			"order": map[string]interface{}{
				"type":        "integer",
				"description": "Execution order (1, 2, 3...). Lower numbers execute first",
				"minimum":     1,
				"default":     1,
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Task status: pending, in_progress, completed",
				"enum":        []string{"pending", "in_progress", "completed"},
			},
		},
		"required": []string{"action"},
	}
}

func (t *TodoUpdateTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddCustomValidator("action", "Action to perform (create, create_batch, update, complete, delete, set_progress)", true, func(value interface{}) error {
			action, ok := value.(string)
			if !ok {
				return fmt.Errorf("action must be a string")
			}
			validActions := []string{"create", "create_batch", "update", "complete", "delete", "set_progress"}
			for _, va := range validActions {
				if action == va {
					return nil
				}
			}
			return fmt.Errorf("invalid action: %s", action)
		}).
		AddOptionalStringField("id", "Todo item ID (required for update, complete, delete)").
		AddOptionalStringField("content", "Todo content (required for create, optional for update)").
		AddOptionalIntField("order", "Execution order (1, 2, 3...)", 1, 0)

	// First run standard validation
	if err := validator.Validate(args); err != nil {
		return err
	}

	// Get validated action
	action := args["action"].(string)

	// Additional action-specific validation
	switch action {
	case "create":
		if _, ok := args["content"]; !ok {
			return fmt.Errorf("content is required for create action")
		}
	case "create_batch":
		tasks, ok := args["tasks"]
		if !ok {
			return fmt.Errorf("tasks array is required for create_batch action")
		}
		tasksSlice, ok := tasks.([]interface{})
		if !ok {
			return fmt.Errorf("tasks must be an array")
		}
		if len(tasksSlice) == 0 {
			return fmt.Errorf("tasks array cannot be empty")
		}
		for i, task := range tasksSlice {
			taskMap, ok := task.(map[string]interface{})
			if !ok {
				return fmt.Errorf("task %d must be an object", i)
			}
			if _, ok := taskMap["content"].(string); !ok {
				return fmt.Errorf("task %d must have content field", i)
			}
		}
	case "update", "complete", "delete", "set_progress":
		if _, ok := args["id"]; !ok {
			return fmt.Errorf("id is required for %s action", action)
		}
	}

	return nil
}

func (t *TodoUpdateTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	action := args["action"].(string)

	// Get current config
	config, err := t.configManager.GetLegacyConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	switch action {
	case "create":
		return t.createTodo(config, args)
	case "create_batch":
		return t.createBatchTodos(config, args)
	case "update":
		return t.updateTodo(config, args)
	case "complete":
		return t.completeTodo(config, args)
	case "delete":
		return t.deleteTodo(config, args)
	case "set_progress":
		return t.setProgress(config, args)
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

func (t *TodoUpdateTool) createTodo(config *types.Config, args map[string]interface{}) (*ToolResult, error) {
	content := args["content"].(string)
	order := 1
	if o, ok := args["order"]; ok {
		if orderFloat, ok := o.(float64); ok {
			order = int(orderFloat)
		}
	}

	// Auto-assign order if not provided or 0
	if order <= 0 {
		order = t.getNextOrder(config)
	}

	// Generate unique ID
	id := fmt.Sprintf("todo_%d", time.Now().UnixNano())

	// Create new todo
	newTodo := types.TodoItem{
		ID:        id,
		Content:   content,
		Status:    "pending",
		Order:     order,
		CreatedAt: time.Now(),
	}

	// Add to config
	config.Todos = append(config.Todos, newTodo)
	config.LastUpdated = time.Now()

	// Save config
	if err := t.saveConfig(config); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}
	return &ToolResult{
		Content: fmt.Sprintf("✅ Created todo: %s (ID: %s, Order: %d)", content, id, order),
		Data: map[string]interface{}{
			"id":      id,
			"content": content,
			"status":  "pending",
			"order":   order,
			"created": newTodo.CreatedAt.Unix(),
		},
	}, nil
}

func (t *TodoUpdateTool) createBatchTodos(config *types.Config, args map[string]interface{}) (*ToolResult, error) {
	tasks := args["tasks"].([]interface{})

	var createdTodos []types.TodoItem
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("✅ Created %d todos:\n", len(tasks)))

	baseOrder := t.getNextOrder(config)

	for i, task := range tasks {
		taskMap := task.(map[string]interface{})
		content := taskMap["content"].(string)

		order := baseOrder + i
		if o, ok := taskMap["order"]; ok {
			if orderFloat, ok := o.(float64); ok {
				order = int(orderFloat)
			}
		}

		// Generate unique ID
		id := fmt.Sprintf("todo_%d_%d", time.Now().UnixNano(), i)

		// Create new todo
		newTodo := types.TodoItem{
			ID:        id,
			Content:   content,
			Status:    "pending",
			Order:     order,
			CreatedAt: time.Now(),
		}

		// Add to config
		config.Todos = append(config.Todos, newTodo)
		createdTodos = append(createdTodos, newTodo)

		// Add to summary
		summary.WriteString(fmt.Sprintf("  %d. [%s] %s\n", order, id[:8], content))
	}

	config.LastUpdated = time.Now()

	// Save config
	if err := t.saveConfig(config); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}
	return &ToolResult{
		Content: summary.String(),
		Data: map[string]interface{}{
			"created_count": len(createdTodos),
			"todos":         createdTodos,
			"batch_id":      fmt.Sprintf("batch_%d", time.Now().Unix()),
		},
	}, nil
}

func (t *TodoUpdateTool) updateTodo(config *types.Config, args map[string]interface{}) (*ToolResult, error) {
	id := args["id"].(string)

	// Find todo
	todoIndex := -1
	for i, todo := range config.Todos {
		if todo.ID == id {
			todoIndex = i
			break
		}
	}

	if todoIndex == -1 {
		return nil, fmt.Errorf("todo not found: %s", id)
	}

	// Update fields
	if content, ok := args["content"].(string); ok {
		config.Todos[todoIndex].Content = content
	}
	if order, ok := args["order"]; ok {
		if orderFloat, ok := order.(float64); ok {
			config.Todos[todoIndex].Order = int(orderFloat)
		}
	}
	if status, ok := args["status"].(string); ok {
		// Validate status transition
		if err := t.validateStatusTransition(config, todoIndex, status); err != nil {
			return nil, err
		}
		config.Todos[todoIndex].Status = status
	}

	config.LastUpdated = time.Now()

	// Save config
	if err := t.saveConfig(config); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	todo := config.Todos[todoIndex]
	return &ToolResult{
		Content: fmt.Sprintf("📝 Updated todo: %s (Status: %s, Order: %d)", todo.Content, todo.Status, todo.Order),
		Data: map[string]interface{}{
			"id":      todo.ID,
			"content": todo.Content,
			"status":  todo.Status,
			"order":   todo.Order,
		},
	}, nil
}

func (t *TodoUpdateTool) completeTodo(config *types.Config, args map[string]interface{}) (*ToolResult, error) {
	id := args["id"].(string)

	// Find todo
	todoIndex := -1
	for i, todo := range config.Todos {
		if todo.ID == id {
			todoIndex = i
			break
		}
	}

	if todoIndex == -1 {
		return nil, fmt.Errorf("todo not found: %s", id)
	}

	// Mark as completed
	now := time.Now()
	config.Todos[todoIndex].Status = "completed"
	config.Todos[todoIndex].CompletedAt = &now
	config.LastUpdated = time.Now()

	// Save config
	if err := t.saveConfig(config); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	todo := config.Todos[todoIndex]
	return &ToolResult{
		Content: fmt.Sprintf("✅ Completed todo: %s", todo.Content),
		Data: map[string]interface{}{
			"id":           todo.ID,
			"content":      todo.Content,
			"status":       "completed",
			"completed_at": now.Unix(),
		},
	}, nil
}

func (t *TodoUpdateTool) deleteTodo(config *types.Config, args map[string]interface{}) (*ToolResult, error) {
	id := args["id"].(string)

	// Find todo
	todoIndex := -1
	var todoContent string
	for i, todo := range config.Todos {
		if todo.ID == id {
			todoIndex = i
			todoContent = todo.Content
			break
		}
	}

	if todoIndex == -1 {
		return nil, fmt.Errorf("todo not found: %s", id)
	}

	// Remove todo
	config.Todos = append(config.Todos[:todoIndex], config.Todos[todoIndex+1:]...)
	config.LastUpdated = time.Now()

	// Save config
	if err := t.saveConfig(config); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return &ToolResult{
		Content: fmt.Sprintf("🗑️ Deleted todo: %s", todoContent),
		Data: map[string]interface{}{
			"id":      id,
			"content": todoContent,
			"deleted": true,
		},
	}, nil
}

func (t *TodoUpdateTool) setProgress(config *types.Config, args map[string]interface{}) (*ToolResult, error) {
	id := args["id"].(string)

	// Find todo
	todoIndex := -1
	for i, todo := range config.Todos {
		if todo.ID == id {
			todoIndex = i
			break
		}
	}

	if todoIndex == -1 {
		return nil, fmt.Errorf("todo not found: %s", id)
	}

	// Check if another task is already in progress
	for i, todo := range config.Todos {
		if i != todoIndex && todo.Status == "in_progress" {
			return nil, fmt.Errorf("another task is already in progress: %s. Only one task can be in progress at a time", todo.Content)
		}
	}

	// Set as in progress
	config.Todos[todoIndex].Status = "in_progress"
	config.LastUpdated = time.Now()

	// Save config
	if err := t.saveConfig(config); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	todo := config.Todos[todoIndex]
	return &ToolResult{
		Content: fmt.Sprintf("🚀 Started working on: %s", todo.Content),
		Data: map[string]interface{}{
			"id":      todo.ID,
			"content": todo.Content,
			"status":  "in_progress",
		},
	}, nil
}

func (t *TodoUpdateTool) validateStatusTransition(config *types.Config, todoIndex int, newStatus string) error {
	// Allow all transitions for now, but enforce single in_progress rule
	if newStatus == "in_progress" {
		for i, todo := range config.Todos {
			if i != todoIndex && todo.Status == "in_progress" {
				return fmt.Errorf("another task is already in progress: %s. Only one task can be in progress at a time", todo.Content)
			}
		}
	}

	return nil
}

func (t *TodoUpdateTool) saveConfig(config *types.Config) error {
	// The config manager should already have the updated config in memory
	// since we modified config directly. Just call Save to persist it.
	return t.configManager.Save()
}

// TodoReadTool provides read-only access to todos
type TodoReadTool struct {
	configManager *config.Manager
}

func NewTodoReadTool(configManager *config.Manager) *TodoReadTool {
	return &TodoReadTool{
		configManager: configManager,
	}
}

func (t *TodoReadTool) Name() string {
	return "todo_read"
}

func (t *TodoReadTool) Description() string {
	return "Read and list todo items with filtering by status and priority. Shows task progress and statistics."
}

func (t *TodoReadTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Filter by status: pending, in_progress, completed, or 'all' for everything",
				"enum":        []string{"pending", "in_progress", "completed", "all"},
				"default":     "all",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of todos to return",
				"default":     50,
				"minimum":     1,
				"maximum":     100,
			},
		},
	}
}

func (t *TodoReadTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddCustomValidator("status", "Filter by status (pending, in_progress, completed, all)", false, func(value interface{}) error {
			if value == nil {
				return nil // Optional field
			}
			status, ok := value.(string)
			if !ok {
				return fmt.Errorf("status must be a string")
			}
			validStatuses := []string{"pending", "in_progress", "completed", "all"}
			for _, vs := range validStatuses {
				if status == vs {
					return nil
				}
			}
			return fmt.Errorf("status must be one of: %v", validStatuses)
		}).
		AddOptionalIntField("limit", "Maximum number of todos to return", 1, 100)

	return validator.Validate(args)
}

func (t *TodoReadTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	config, err := t.configManager.GetLegacyConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	statusFilter := "all"
	if s, ok := args["status"].(string); ok {
		statusFilter = s
	}

	limit := 50
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	// Filter and sort todos by order
	var filteredTodos []types.TodoItem
	for _, todo := range config.Todos {
		// Status filter
		if statusFilter != "all" && todo.Status != statusFilter {
			continue
		}

		filteredTodos = append(filteredTodos, todo)
	}

	// Sort by order
	for i := 0; i < len(filteredTodos)-1; i++ {
		for j := i + 1; j < len(filteredTodos); j++ {
			if filteredTodos[i].Order > filteredTodos[j].Order {
				filteredTodos[i], filteredTodos[j] = filteredTodos[j], filteredTodos[i]
			}
		}
	}

	// Apply limit
	if len(filteredTodos) > limit {
		filteredTodos = filteredTodos[:limit]
	}

	// Generate summary
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("📋 Todo List (%d items", len(filteredTodos)))
	if statusFilter != "all" {
		summary.WriteString(fmt.Sprintf(", status: %s", statusFilter))
	}
	summary.WriteString("):\n\n")

	if len(filteredTodos) == 0 {
		summary.WriteString("No todos found matching the criteria.")
	} else {
		// Group by status for better readability
		statusGroups := map[string][]types.TodoItem{
			"in_progress": {},
			"pending":     {},
			"completed":   {},
		}

		for _, todo := range filteredTodos {
			statusGroups[todo.Status] = append(statusGroups[todo.Status], todo)
		}

		// Display in order: in_progress, pending, completed
		for _, status := range []string{"in_progress", "pending", "completed"} {
			todos := statusGroups[status]
			if len(todos) == 0 {
				continue
			}

			var statusIcon string
			switch status {
			case "in_progress":
				statusIcon = "🚀"
			case "pending":
				statusIcon = "⏳"
			case "completed":
				statusIcon = "✅"
			}

			summary.WriteString(fmt.Sprintf("%s %s (%d):\n", statusIcon, strings.ToUpper(strings.ReplaceAll(status, "_", " ")), len(todos)))
			for _, todo := range todos {
				summary.WriteString(fmt.Sprintf("  %d. [%s] %s\n", todo.Order, todo.ID[:8], todo.Content))
			}
			summary.WriteString("\n")
		}
	}

	// Statistics
	stats := map[string]int{
		"total":       len(config.Todos),
		"pending":     0,
		"in_progress": 0,
		"completed":   0,
	}

	for _, todo := range config.Todos {
		stats[todo.Status]++
	}

	return &ToolResult{
		Content: summary.String(),
		Data: map[string]interface{}{
			"todos":       filteredTodos,
			"total_count": len(filteredTodos),
			"stats":       stats,
			"filters_applied": map[string]string{
				"status": statusFilter,
				"limit":  strconv.Itoa(limit),
			},
		},
	}, nil
}

// getNextOrder returns the next available order number
func (t *TodoUpdateTool) getNextOrder(config *types.Config) int {
	maxOrder := 0
	for _, todo := range config.Todos {
		if todo.Order > maxOrder {
			maxOrder = todo.Order
		}
	}
	return maxOrder + 1
}
