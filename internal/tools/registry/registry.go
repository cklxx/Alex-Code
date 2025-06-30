package registry

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"deep-coding-agent/internal/config"
	"deep-coding-agent/internal/tools/builtin"
)

// Tool represents a tool that can be executed by the agent
type Tool interface {
	// Name returns the unique name of the tool
	Name() string
	
	// Description returns a human-readable description of what the tool does
	Description() string
	
	// Parameters returns the JSON schema for the tool's parameters
	Parameters() map[string]interface{}
	
	// Execute runs the tool with the given arguments
	Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
	
	// Validate checks if the provided arguments are valid for this tool
	Validate(args map[string]interface{}) error
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Content string                 `json:"content"`
	Data    interface{}           `json:"data,omitempty"`
	Files   []string              `json:"files,omitempty"`    // Files that were modified/created
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToolMetadata holds metadata about a tool
type ToolMetadata struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Category    string                 `json:"category"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author,omitempty"`
	
	// Execution settings
	Timeout     int  `json:"timeout,omitempty"`      // seconds
	RequiresSudo bool `json:"requires_sudo,omitempty"`
	IsDangerous bool `json:"is_dangerous,omitempty"`  // Requires explicit confirmation
}

// Registry manages available tools and their execution
type Registry struct {
	tools    map[string]Tool
	metadata map[string]*ToolMetadata
	mutex    sync.RWMutex
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	registry := &Registry{
		tools:    make(map[string]Tool),
		metadata: make(map[string]*ToolMetadata),
	}
	
	// Register core tools by default
	registry.registerCoreTools()
	
	return registry
}

// NewRegistryWithConfig creates a new tool registry with configuration support
func NewRegistryWithConfig(configManager interface{}) *Registry {
	registry := NewRegistry()
	
	// Configure tools if config manager is provided
	if configManager != nil {
		// Try to cast to the expected config manager type
		if cm, ok := configManager.(*config.Manager); ok {
			configurator := NewToolConfigurator(cm)
			configurator.ConfigureAllTools(registry)
		}
	}
	
	return registry
}

// RegisterTool registers a new tool
func (r *Registry) RegisterTool(tool Tool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool '%s' is already registered", name)
	}

	r.tools[name] = tool
	
	// Create metadata
	r.metadata[name] = &ToolMetadata{
		Name:        name,
		Description: tool.Description(),
		Parameters:  tool.Parameters(),
		Category:    r.inferCategory(name),
		Version:     "1.0.0",
	}

	return nil
}

// UnregisterTool removes a tool from the registry
func (r *Registry) UnregisterTool(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.tools[name]; !exists {
		return fmt.Errorf("tool '%s' is not registered", name)
	}

	delete(r.tools, name)
	delete(r.metadata, name)
	return nil
}

// GetTool returns a tool by name
func (r *Registry) GetTool(name string) Tool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	return r.tools[name]
}

// GetToolMetadata returns metadata for a tool
func (r *Registry) GetToolMetadata(name string) *ToolMetadata {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	if metadata, exists := r.metadata[name]; exists {
		// Return a copy to prevent external modification
		metadataCopy := *metadata
		return &metadataCopy
	}
	return nil
}

// ListTools returns a list of all registered tool names
func (r *Registry) ListTools() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tools := make([]string, 0, len(r.tools))
	for name := range r.tools {
		tools = append(tools, name)
	}
	return tools
}

// ListToolsByCategory returns tools grouped by category
func (r *Registry) ListToolsByCategory() map[string][]string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	categories := make(map[string][]string)
	for name, metadata := range r.metadata {
		category := metadata.Category
		if category == "" {
			category = "other"
		}
		categories[category] = append(categories[category], name)
	}
	return categories
}

// ExecuteTool executes a tool with the given arguments
func (r *Registry) ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	r.mutex.RLock()
	tool, exists := r.tools[name]
	metadata := r.metadata[name]
	r.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}

	// Validate arguments
	if err := tool.Validate(args); err != nil {
		return nil, fmt.Errorf("invalid arguments for tool '%s': %w", name, err)
	}

	// Check for dangerous tools
	if metadata != nil && metadata.IsDangerous {
		// In a production system, this would prompt for confirmation
		// For now, we'll just add a warning to the result
	}

	// Apply timeout if specified
	if metadata != nil && metadata.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(metadata.Timeout)*time.Second)
		defer cancel()
	}

	// Execute the tool
	result, err := tool.Execute(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed for '%s': %w", name, err)
	}

	// Add execution metadata
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["tool_name"] = name
	result.Metadata["executed_at"] = time.Now().Unix()

	return result, nil
}

// ValidateToolArgs validates arguments for a specific tool
func (r *Registry) ValidateToolArgs(name string, args map[string]interface{}) error {
	r.mutex.RLock()
	tool, exists := r.tools[name]
	r.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("tool '%s' not found", name)
	}

	return tool.Validate(args)
}

// GetToolSchema returns the JSON schema for a tool's parameters
func (r *Registry) GetToolSchema(name string) (map[string]interface{}, error) {
	r.mutex.RLock()
	tool, exists := r.tools[name]
	r.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}

	return tool.Parameters(), nil
}

// registerCoreTools registers the core tools
func (r *Registry) registerCoreTools() {
	// Register builtin tools directly
	r.registerBuiltinTools()
}

// registerBuiltinTools registers all builtin tools
func (r *Registry) registerBuiltinTools() {
	builtinTools := builtin.GetAllBuiltinTools()
	for _, builtinTool := range builtinTools {
		// Create a wrapper to convert builtin.Tool to tools.Tool
		wrapper := &BuiltinToolWrapper{tool: builtinTool}
		if err := r.RegisterTool(wrapper); err != nil {
			fmt.Printf("Warning: failed to register builtin tool '%s': %v\n", builtinTool.Name(), err)
		}
	}
}

// BuiltinToolWrapper wraps a builtin tool to implement the tools.Tool interface
type BuiltinToolWrapper struct {
	tool builtin.Tool
}

func (w *BuiltinToolWrapper) Name() string {
	return w.tool.Name()
}

func (w *BuiltinToolWrapper) Description() string {
	return w.tool.Description()
}

func (w *BuiltinToolWrapper) Parameters() map[string]interface{} {
	return w.tool.Parameters()
}

func (w *BuiltinToolWrapper) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	builtinResult, err := w.tool.Execute(ctx, args)
	if err != nil {
		return nil, err
	}
	
	// Convert builtin.ToolResult to tools.ToolResult
	return &ToolResult{
		Content:  builtinResult.Content,
		Data:     builtinResult.Data,
		Files:    builtinResult.Files,
		Metadata: builtinResult.Metadata,
	}, nil
}

func (w *BuiltinToolWrapper) Validate(args map[string]interface{}) error {
	return w.tool.Validate(args)
}

// inferCategory attempts to categorize a tool based on its name
func (r *Registry) inferCategory(name string) string {
	switch {
	case contains(name, []string{"todo", "task"}):
		return "productivity"
	case contains(name, []string{"file", "read", "write", "list", "dir", "update", "replace"}):
		return "file"
	case contains(name, []string{"bash", "shell", "cmd", "exec", "script", "process"}):
		return "execution"
	case contains(name, []string{"grep", "find", "search", "ripgrep"}):
		return "search"
	case contains(name, []string{"analyze", "analysis", "lint", "check"}):
		return "analysis"
	case contains(name, []string{"generate", "create", "build"}):
		return "generation"
	case contains(name, []string{"refactor", "optimize", "transform"}):
		return "refactoring"
	case contains(name, []string{"git", "commit", "branch", "merge"}):
		return "git"
	case contains(name, []string{"test", "testing", "spec"}):
		return "testing"
	default:
		return "other"
	}
}

// Helper function to check if a string contains any of the given substrings
func contains(s string, substrings []string) bool {
	s = strings.ToLower(s)
	for _, sub := range substrings {
		if strings.Contains(s, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

