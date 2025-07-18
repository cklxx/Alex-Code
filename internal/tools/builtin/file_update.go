package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileUpdateTool implements file content updating functionality
type FileUpdateTool struct{}

func CreateFileUpdateTool() *FileUpdateTool {
	return &FileUpdateTool{}
}

func (t *FileUpdateTool) Name() string {
	return "file_edit"
}

func (t *FileUpdateTool) Description() string {
	return "Edit files by replacing specific text. For new files, use empty old_string."
}

func (t *FileUpdateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "The absolute path to the file to modify",
			},
			"old_string": map[string]interface{}{
				"type":        "string",
				"description": "The text to replace (empty for new file)",
			},
			"new_string": map[string]interface{}{
				"type":        "string",
				"description": "The text to replace with",
			},
		},
		"required": []string{"file_path", "old_string", "new_string"},
	}
}

func (t *FileUpdateTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("file_path", "Path to the file").
		AddRequiredStringField("old_string", "Text to replace (empty for new file)").
		AddStringField("new_string", "Replacement text")

	return validator.Validate(args)
}

func (t *FileUpdateTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	filePath := args["file_path"].(string)
	newString := args["new_string"].(string)
	
	oldString := ""
	if os, ok := args["old_string"]; ok {
		oldString = os.(string)
	}

	// 解析路径（处理相对路径）
	resolver := GetPathResolverFromContext(ctx)
	resolvedPath := resolver.ResolvePath(filePath)

	// Handle new file creation case (empty old_string)
	if oldString == "" {
		// Create parent directories if needed
		dir := filepath.Dir(resolvedPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directories: %w", err)
		}

		// Check if file already exists
		if _, err := os.Stat(resolvedPath); err == nil {
			return nil, fmt.Errorf("file already exists: %s", filePath)
		}

		// Write new file
		err := os.WriteFile(resolvedPath, []byte(newString), 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %w", err)
		}

		fileInfo, _ := os.Stat(resolvedPath)
		return &ToolResult{
			Content: fmt.Sprintf("Successfully created new file %s (%d bytes)", filePath, len(newString)),
			Files:   []string{resolvedPath},
			Data: map[string]interface{}{
				"file_path":     filePath,
				"resolved_path": resolvedPath,
				"operation":     "created",
				"bytes_written": len(newString),
				"lines_total":   len(strings.Split(newString, "\n")),
				"modified":      fileInfo.ModTime().Unix(),
			},
		}, nil
	}

	// Check if file exists for editing
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read file content
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)

	// Check for uniqueness of old_string
	occurrences := strings.Count(originalContent, oldString)
	if occurrences == 0 {
		return nil, fmt.Errorf("old_string not found in file")
	}
	if occurrences > 1 {
		return nil, fmt.Errorf("old_string appears %d times in file. Please include more context to make it unique", occurrences)
	}

	// Perform the replacement (only one occurrence)
	newContent := strings.Replace(originalContent, oldString, newString, 1)

	// Write the modified content
	err = os.WriteFile(resolvedPath, []byte(newContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Get file info after writing
	fileInfo, _ := os.Stat(resolvedPath)
	newLineCount := len(strings.Split(newContent, "\n"))

	return &ToolResult{
		Content: fmt.Sprintf("Updated %s (%d lines total)", filePath, newLineCount),
		Files:   []string{resolvedPath},
		Data: map[string]interface{}{
			"file_path":         filePath,
			"resolved_path":     resolvedPath,
			"operation":         "edited",
			"bytes_written":     len(newContent),
			"lines_total":       newLineCount,
			"modified":          fileInfo.ModTime().Unix(),
			"replacements_made": 1,
		},
	}, nil
}