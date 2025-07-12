package builtin

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"alex/internal/session"
	"alex/pkg/types"
)

// SessionTodoUpdateTool handles session-specific todo management operations
type SessionTodoUpdateTool struct {
	sessionManager *session.Manager
}

// NewSessionTodoUpdateTool creates a new session-aware todo update tool
func NewSessionTodoUpdateTool(sessionManager *session.Manager) *SessionTodoUpdateTool {
	return &SessionTodoUpdateTool{
		sessionManager: sessionManager,
	}
}

func (t *SessionTodoUpdateTool) Name() string {
	return "todo_update"
}

func (t *SessionTodoUpdateTool) Description() string {
	return "Create, update, complete, and delete todo tasks in the current session. Supports batch operations and task management."
}

func (t *SessionTodoUpdateTool) Parameters() map[string]interface{} {
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

func (t *SessionTodoUpdateTool) Validate(args map[string]interface{}) error {
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

func (t *SessionTodoUpdateTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	// Try to get session from context first
	sessionID := GetSessionFromContext(ctx)

	// If not found in context, try to get from working directory context
	if sessionID == "" {
		if workingDir := GetWorkingDirFromContext(ctx); workingDir != "" {
			// Create a temporary session ID based on working directory
			sessionID = fmt.Sprintf("wd_%s_%d", strings.ReplaceAll(workingDir, "/", "_"), time.Now().Unix()/3600)
		}
	}

	// If still no session ID, create one
	if sessionID == "" {
		sessionID = fmt.Sprintf("temp_session_%d", time.Now().Unix())
	}

	// Get or create session
	session, err := t.sessionManager.RestoreSession(sessionID)
	if err != nil {
		// If session doesn't exist, create it
		session, err = t.sessionManager.StartSession(sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
	}

	action := args["action"].(string)

	switch action {
	case "create":
		return t.createTodo(session, args)
	case "create_batch":
		return t.createBatchTodos(session, args)
	case "update":
		return t.updateTodo(session, args)
	case "complete":
		return t.completeTodo(session, args)
	case "delete":
		return t.deleteTodo(session, args)
	case "set_progress":
		return t.setProgress(session, args)
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

func (t *SessionTodoUpdateTool) createTodo(session *session.Session, args map[string]interface{}) (*ToolResult, error) {
	content := args["content"].(string)
	order := 1
	if o, ok := args["order"]; ok {
		if orderFloat, ok := o.(float64); ok {
			order = int(orderFloat)
		}
	}

	// Get current todos from session
	todos := t.getTodosFromSession(session)

	// Auto-assign order if not provided or 0
	if order <= 0 {
		order = t.getNextOrder(todos)
	}

	// Generate simple unique ID
	id := fmt.Sprintf("todo_%d", len(todos)+1)

	// Create new todo
	newTodo := types.TodoItem{
		ID:        id,
		Content:   content,
		Status:    "pending",
		Order:     order,
		CreatedAt: time.Now(),
	}

	// Add to todos
	todos = append(todos, newTodo)

	// Save todos to session
	if err := t.saveTodosToSession(session, todos); err != nil {
		return nil, fmt.Errorf("failed to save todos to session: %w", err)
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

func (t *SessionTodoUpdateTool) createBatchTodos(session *session.Session, args map[string]interface{}) (*ToolResult, error) {
	tasks := args["tasks"].([]interface{})

	// Get current todos from session
	todos := t.getTodosFromSession(session)

	var createdTodos []types.TodoItem
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("✅ Created %d todos:\n", len(tasks)))

	baseOrder := t.getNextOrder(todos)

	for i, task := range tasks {
		taskMap := task.(map[string]interface{})
		content := taskMap["content"].(string)

		order := baseOrder + i
		if o, ok := taskMap["order"]; ok {
			if orderFloat, ok := o.(float64); ok {
				order = int(orderFloat)
			}
		}

		// Generate simple unique ID
		id := fmt.Sprintf("todo_%d", len(todos)+i+1)

		// Create new todo
		newTodo := types.TodoItem{
			ID:        id,
			Content:   content,
			Status:    "pending",
			Order:     order,
			CreatedAt: time.Now(),
		}

		// Add to todos
		todos = append(todos, newTodo)
		createdTodos = append(createdTodos, newTodo)

		// Add to summary (safely truncate ID to max 8 chars)
		displayID := id
		if len(id) > 8 {
			displayID = id[:8]
		}
		summary.WriteString(fmt.Sprintf("  %d. [%s] %s\n", order, displayID, content))
	}

	// Save todos to session
	if err := t.saveTodosToSession(session, todos); err != nil {
		return nil, fmt.Errorf("failed to save todos to session: %w", err)
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

func (t *SessionTodoUpdateTool) updateTodo(session *session.Session, args map[string]interface{}) (*ToolResult, error) {
	id := args["id"].(string)

	// Get current todos from session
	todos := t.getTodosFromSession(session)

	// Find todo
	todoIndex := -1
	for i, todo := range todos {
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
		todos[todoIndex].Content = content
	}
	if order, ok := args["order"]; ok {
		if orderFloat, ok := order.(float64); ok {
			todos[todoIndex].Order = int(orderFloat)
		}
	}
	if status, ok := args["status"].(string); ok {
		// Validate status transition
		if err := t.validateStatusTransition(todos, todoIndex, status); err != nil {
			return nil, err
		}
		todos[todoIndex].Status = status
	}

	// Save todos to session
	if err := t.saveTodosToSession(session, todos); err != nil {
		return nil, fmt.Errorf("failed to save todos to session: %w", err)
	}

	todo := todos[todoIndex]
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

func (t *SessionTodoUpdateTool) completeTodo(session *session.Session, args map[string]interface{}) (*ToolResult, error) {
	id := args["id"].(string)

	// Get current todos from session
	todos := t.getTodosFromSession(session)

	// Find todo
	todoIndex := -1
	for i, todo := range todos {
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
	todos[todoIndex].Status = "completed"
	todos[todoIndex].CompletedAt = &now

	// Save todos to session
	if err := t.saveTodosToSession(session, todos); err != nil {
		return nil, fmt.Errorf("failed to save todos to session: %w", err)
	}

	todo := todos[todoIndex]
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

func (t *SessionTodoUpdateTool) deleteTodo(session *session.Session, args map[string]interface{}) (*ToolResult, error) {
	id := args["id"].(string)

	// Get current todos from session
	todos := t.getTodosFromSession(session)

	// Find todo
	todoIndex := -1
	var todoContent string
	for i, todo := range todos {
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
	todos = append(todos[:todoIndex], todos[todoIndex+1:]...)

	// Save todos to session
	if err := t.saveTodosToSession(session, todos); err != nil {
		return nil, fmt.Errorf("failed to save todos to session: %w", err)
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

func (t *SessionTodoUpdateTool) setProgress(session *session.Session, args map[string]interface{}) (*ToolResult, error) {
	id := args["id"].(string)

	// Get current todos from session
	todos := t.getTodosFromSession(session)

	// Find todo
	todoIndex := -1
	for i, todo := range todos {
		if todo.ID == id {
			todoIndex = i
			break
		}
	}

	if todoIndex == -1 {
		return nil, fmt.Errorf("todo not found: %s", id)
	}

	// Check if another task is already in progress
	for i, todo := range todos {
		if i != todoIndex && todo.Status == "in_progress" {
			return nil, fmt.Errorf("another task is already in progress: %s. Only one task can be in progress at a time", todo.Content)
		}
	}

	// Set as in progress
	todos[todoIndex].Status = "in_progress"

	// Save todos to session
	if err := t.saveTodosToSession(session, todos); err != nil {
		return nil, fmt.Errorf("failed to save todos to session: %w", err)
	}

	todo := todos[todoIndex]
	return &ToolResult{
		Content: fmt.Sprintf("🚀 Started working on: %s", todo.Content),
		Data: map[string]interface{}{
			"id":      todo.ID,
			"content": todo.Content,
			"status":  "in_progress",
		},
	}, nil
}

func (t *SessionTodoUpdateTool) validateStatusTransition(todos []types.TodoItem, todoIndex int, newStatus string) error {
	// Allow all transitions for now, but enforce single in_progress rule
	if newStatus == "in_progress" {
		for i, todo := range todos {
			if i != todoIndex && todo.Status == "in_progress" {
				return fmt.Errorf("another task is already in progress: %s. Only one task can be in progress at a time", todo.Content)
			}
		}
	}

	return nil
}

func (t *SessionTodoUpdateTool) getTodosFromSession(session *session.Session) []types.TodoItem {
	todosInterface, exists := session.GetConfig("todos")
	if !exists {
		return []types.TodoItem{}
	}

	// Convert interface{} to []types.TodoItem
	if todosList, ok := todosInterface.([]interface{}); ok {
		var todos []types.TodoItem
		for _, todoInterface := range todosList {
			if todoMap, ok := todoInterface.(map[string]interface{}); ok {
				todo := types.TodoItem{
					ID:      getString(todoMap, "id"),
					Content: getString(todoMap, "content"),
					Status:  getString(todoMap, "status"),
					Order:   getInt(todoMap, "order"),
				}

				// Parse timestamps
				if createdAtStr := getString(todoMap, "created_at"); createdAtStr != "" {
					if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
						todo.CreatedAt = createdAt
					}
				}

				if completedAtStr := getString(todoMap, "completed_at"); completedAtStr != "" {
					if completedAt, err := time.Parse(time.RFC3339, completedAtStr); err == nil {
						todo.CompletedAt = &completedAt
					}
				}

				todos = append(todos, todo)
			}
		}
		return todos
	}

	return []types.TodoItem{}
}

func (t *SessionTodoUpdateTool) saveTodosToSession(session *session.Session, todos []types.TodoItem) error {
	// Convert todos to interface{} format for session storage
	var todosInterface []interface{}
	for _, todo := range todos {
		todoMap := map[string]interface{}{
			"id":         todo.ID,
			"content":    todo.Content,
			"status":     todo.Status,
			"order":      todo.Order,
			"created_at": todo.CreatedAt.Format(time.RFC3339),
		}
		if todo.CompletedAt != nil {
			todoMap["completed_at"] = todo.CompletedAt.Format(time.RFC3339)
		}
		todosInterface = append(todosInterface, todoMap)
	}

	session.SetConfig("todos", todosInterface)
	return t.sessionManager.SaveSession(session)
}

func (t *SessionTodoUpdateTool) getNextOrder(todos []types.TodoItem) int {
	maxOrder := 0
	for _, todo := range todos {
		if todo.Order > maxOrder {
			maxOrder = todo.Order
		}
	}
	return maxOrder + 1
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	return 0
}

// SessionTodoReadTool provides session-specific read-only access to todos
type SessionTodoReadTool struct {
	sessionManager *session.Manager
}

func NewSessionTodoReadTool(sessionManager *session.Manager) *SessionTodoReadTool {
	return &SessionTodoReadTool{
		sessionManager: sessionManager,
	}
}

func (t *SessionTodoReadTool) Name() string {
	return "todo_read"
}

func (t *SessionTodoReadTool) Description() string {
	return "Read and list todo items from the current session with filtering by status and priority. Shows task progress and statistics."
}

func (t *SessionTodoReadTool) Parameters() map[string]interface{} {
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

func (t *SessionTodoReadTool) Validate(args map[string]interface{}) error {
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

func (t *SessionTodoReadTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	// Try to get session from context first
	sessionID := GetSessionFromContext(ctx)

	// If not found in context, try to get from working directory context
	if sessionID == "" {
		if workingDir := GetWorkingDirFromContext(ctx); workingDir != "" {
			// Create a temporary session ID based on working directory
			sessionID = fmt.Sprintf("wd_%s_%d", strings.ReplaceAll(workingDir, "/", "_"), time.Now().Unix()/3600)
		}
	}

	// If still no session ID, create one
	if sessionID == "" {
		sessionID = fmt.Sprintf("temp_session_%d", time.Now().Unix())
	}

	// Get or create session
	session, err := t.sessionManager.RestoreSession(sessionID)
	if err != nil {
		// If session doesn't exist, create it
		session, err = t.sessionManager.StartSession(sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
	}

	statusFilter := "all"
	if s, ok := args["status"].(string); ok {
		statusFilter = s
	}

	limit := 50
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	// Get todos from session
	todos := t.getTodosFromSession(session)

	// Filter and sort todos by order
	var filteredTodos []types.TodoItem
	for _, todo := range todos {
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
	summary.WriteString(fmt.Sprintf("📋 Session Todo List (%d items", len(filteredTodos)))
	if statusFilter != "all" {
		summary.WriteString(fmt.Sprintf(", status: %s", statusFilter))
	}
	summary.WriteString("):\n\n")

	if len(filteredTodos) == 0 {
		summary.WriteString("No todos found in the current session matching the criteria.")
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
		"total":       len(todos),
		"pending":     0,
		"in_progress": 0,
		"completed":   0,
	}

	for _, todo := range todos {
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

func (t *SessionTodoReadTool) getTodosFromSession(session *session.Session) []types.TodoItem {
	todosInterface, exists := session.GetConfig("todos")
	if !exists {
		return []types.TodoItem{}
	}

	// Convert interface{} to []types.TodoItem
	if todosList, ok := todosInterface.([]interface{}); ok {
		var todos []types.TodoItem
		for _, todoInterface := range todosList {
			if todoMap, ok := todoInterface.(map[string]interface{}); ok {
				todo := types.TodoItem{
					ID:      getString(todoMap, "id"),
					Content: getString(todoMap, "content"),
					Status:  getString(todoMap, "status"),
					Order:   getInt(todoMap, "order"),
				}

				// Parse timestamps
				if createdAtStr := getString(todoMap, "created_at"); createdAtStr != "" {
					if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
						todo.CreatedAt = createdAt
					}
				}

				if completedAtStr := getString(todoMap, "completed_at"); completedAtStr != "" {
					if completedAt, err := time.Parse(time.RFC3339, completedAtStr); err == nil {
						todo.CompletedAt = &completedAt
					}
				}

				todos = append(todos, todo)
			}
		}
		return todos
	}

	return []types.TodoItem{}
}

const SessionIDKey ContextKey = "sessionID"

// GetSessionFromContext retrieves the session ID from the context
func GetSessionFromContext(ctx context.Context) string {
	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok {
		return sessionID
	}
	return ""
}

// GetWorkingDirFromContext retrieves the working directory from the context
func GetWorkingDirFromContext(ctx context.Context) string {
	if workingDir, ok := ctx.Value(WorkingDirKey).(string); ok {
		return workingDir
	}
	return ""
}
