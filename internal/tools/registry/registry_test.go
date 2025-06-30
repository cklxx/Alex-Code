package registry

import (
	"context"
	"testing"
)

func TestRegistryUnified(t *testing.T) {
	// Test that registry creates and registers builtin tools
	registry := NewRegistry()
	
	// Check that core tools are registered
	tools := registry.ListTools()
	if len(tools) == 0 {
		t.Fatal("Expected builtin tools to be registered, got none")
	}
	
	// Test that we have the expected tools
	expectedTools := []string{
		"file_read", "file_update", "file_replace", "file_list", "directory_create",
		"grep", "bash",
	}
	
	registeredTools := make(map[string]bool)
	for _, tool := range tools {
		registeredTools[tool] = true
	}
	
	for _, expected := range expectedTools {
		if !registeredTools[expected] {
			t.Errorf("Expected tool '%s' to be registered, but it wasn't", expected)
		}
	}
	
	// Test tool execution
	result, err := registry.ExecuteTool(context.Background(), "file_read", map[string]interface{}{
		"file_path": "nonexistent.txt",
	})
	
	// Should get an error for nonexistent file, but not a registry error
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	
	if result != nil {
		t.Error("Expected nil result for failed execution")
	}
}

func TestToolValidation(t *testing.T) {
	registry := NewRegistry()
	
	// Test validation
	err := registry.ValidateToolArgs("file_read", map[string]interface{}{
		"file_path": "/etc/passwd",
	})
	
	if err != nil {
		t.Errorf("Expected valid args to pass validation, got: %v", err)
	}
	
	// Test invalid args
	err = registry.ValidateToolArgs("file_read", map[string]interface{}{
		// missing file_path
	})
	
	if err == nil {
		t.Error("Expected validation error for missing file_path")
	}
}

func TestToolMetadata(t *testing.T) {
	registry := NewRegistry()
	
	metadata := registry.GetToolMetadata("file_read")
	if metadata == nil {
		t.Fatal("Expected metadata for file_read tool")
	}
	
	if metadata.Name != "file_read" {
		t.Errorf("Expected name 'file_read', got '%s'", metadata.Name)
	}
	
	if metadata.Category == "" {
		t.Error("Expected category to be set")
	}
}