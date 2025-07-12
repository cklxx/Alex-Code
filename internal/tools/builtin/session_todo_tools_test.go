package builtin

import (
	"context"
	"testing"

	"alex/internal/session"
	"alex/pkg/types"
)

func TestSessionTodoIDIncrement(t *testing.T) {
	// Create a temporary session manager
	sessionManager, err := session.NewManager()
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	// Create todo tool
	tool := NewSessionTodoUpdateTool(sessionManager)

	// Start a test session
	_, err = sessionManager.StartSession("test-session")
	if err != nil {
		t.Fatalf("Failed to start session: %v", err)
	}

	// Test 1: Create first todo - should get ID "1"
	args1 := map[string]interface{}{
		"action":  "create",
		"content": "First todo",
	}

	result1, err := tool.Execute(context.Background(), args1)
	if err != nil {
		t.Fatalf("Failed to create first todo: %v", err)
	}

	firstID := result1.Data["id"].(string)
	if firstID != "1" {
		t.Errorf("Expected first todo ID to be '1', got '%s'", firstID)
	}

	// Test 2: Create second todo - should get ID "2"
	args2 := map[string]interface{}{
		"action":  "create",
		"content": "Second todo",
	}

	result2, err := tool.Execute(context.Background(), args2)
	if err != nil {
		t.Fatalf("Failed to create second todo: %v", err)
	}

	secondID := result2.Data["id"].(string)
	if secondID != "2" {
		t.Errorf("Expected second todo ID to be '2', got '%s'", secondID)
	}

	// Test 3: Create third todo - should get ID "3"
	args3 := map[string]interface{}{
		"action":  "create",
		"content": "Third todo",
	}

	result3, err := tool.Execute(context.Background(), args3)
	if err != nil {
		t.Fatalf("Failed to create third todo: %v", err)
	}

	thirdID := result3.Data["id"].(string)
	if thirdID != "3" {
		t.Errorf("Expected third todo ID to be '3', got '%s'", thirdID)
	}

	// Test 4: Delete second todo
	deleteArgs := map[string]interface{}{
		"action": "delete",
		"id":     "2",
	}

	_, err = tool.Execute(context.Background(), deleteArgs)
	if err != nil {
		t.Fatalf("Failed to delete todo: %v", err)
	}

	// Test 5: Create fourth todo - should get ID "4" (not "2")
	args4 := map[string]interface{}{
		"action":  "create",
		"content": "Fourth todo",
	}

	result4, err := tool.Execute(context.Background(), args4)
	if err != nil {
		t.Fatalf("Failed to create fourth todo: %v", err)
	}

	fourthID := result4.Data["id"].(string)
	if fourthID != "4" {
		t.Errorf("Expected fourth todo ID to be '4', got '%s'", fourthID)
	}

	// Test 6: Batch creation - should get IDs "5" and "6"
	batchArgs := map[string]interface{}{
		"action": "create_batch",
		"tasks": []interface{}{
			map[string]interface{}{"content": "Batch todo 1"},
			map[string]interface{}{"content": "Batch todo 2"},
		},
	}

	batchResult, err := tool.Execute(context.Background(), batchArgs)
	if err != nil {
		t.Fatalf("Failed to create batch todos: %v", err)
	}

	batchTodos := batchResult.Data["todos"].([]types.TodoItem)
	if len(batchTodos) != 2 {
		t.Fatalf("Expected 2 batch todos, got %d", len(batchTodos))
	}

	if batchTodos[0].ID != "5" {
		t.Errorf("Expected first batch todo ID to be '5', got '%s'", batchTodos[0].ID)
	}

	if batchTodos[1].ID != "6" {
		t.Errorf("Expected second batch todo ID to be '6', got '%s'", batchTodos[1].ID)
	}
}

func TestSessionTodoGetNextID(t *testing.T) {
	sessionManager, err := session.NewManager()
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	tool := NewSessionTodoUpdateTool(sessionManager)

	// Test empty todos
	emptyTodos := []types.TodoItem{}
	nextID := tool.getNextID(emptyTodos)
	if nextID != 1 {
		t.Errorf("Expected next ID for empty todos to be 1, got %d", nextID)
	}

	// Test todos with IDs 1, 3, 5 (gaps)
	todosWithGaps := []types.TodoItem{
		{ID: "1", Content: "Todo 1"},
		{ID: "3", Content: "Todo 3"},
		{ID: "5", Content: "Todo 5"},
	}
	nextID = tool.getNextID(todosWithGaps)
	if nextID != 6 {
		t.Errorf("Expected next ID after todos 1,3,5 to be 6, got %d", nextID)
	}

	// Test todos with non-numeric IDs (should ignore them)
	mixedTodos := []types.TodoItem{
		{ID: "1", Content: "Todo 1"},
		{ID: "abc", Content: "Todo abc"},
		{ID: "3", Content: "Todo 3"},
	}
	nextID = tool.getNextID(mixedTodos)
	if nextID != 4 {
		t.Errorf("Expected next ID after todos 1,abc,3 to be 4, got %d", nextID)
	}
}