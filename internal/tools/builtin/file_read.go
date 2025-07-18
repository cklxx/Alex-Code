package builtin

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// FileReadTool implements file reading functionality
type FileReadTool struct{}

func CreateFileReadTool() *FileReadTool {
	return &FileReadTool{}
}

func (t *FileReadTool) Name() string {
	return "file_read"
}

func (t *FileReadTool) Description() string {
	return "Read the contents of a file. Supports reading specific line ranges."
}

func (t *FileReadTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to read",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to read (legacy parameter)",
			},
			"start_line": map[string]interface{}{
				"type":        "integer",
				"description": "Starting line number (1-based, optional)",
				"minimum":     1,
			},
			"end_line": map[string]interface{}{
				"type":        "integer",
				"description": "Ending line number (1-based, optional)",
				"minimum":     1,
			},
		},
	}
}

func (t *FileReadTool) Validate(args map[string]interface{}) error {
	// Check if either file_path or path is provided
	if _, hasFilePath := args["file_path"]; !hasFilePath {
		if _, hasPath := args["path"]; !hasPath {
			return fmt.Errorf("either file_path or path is required")
		}
	}
	
	validator := NewValidationFramework().
		AddOptionalStringField("file_path", "Path to the file to read").
		AddOptionalStringField("path", "Path to the file to read (legacy)").
		AddOptionalIntField("start_line", "Starting line number (1-based)", 1, 0).
		AddOptionalIntField("end_line", "Ending line number (1-based)", 1, 0)

	return validator.Validate(args)
}

func (t *FileReadTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	// Get file path from either parameter
	var filePath string
	if fp, ok := args["file_path"]; ok {
		filePath = fp.(string)
	} else if p, ok := args["path"]; ok {
		filePath = p.(string)
	} else {
		return nil, fmt.Errorf("either file_path or path is required")
	}
	
	// 解析路径（处理相对路径）
	resolver := GetPathResolverFromContext(ctx)
	resolvedPath := resolver.ResolvePath(filePath)
	
	// Check if file exists
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read file content
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")
	var formattedLines []string
	startLineNum := 1
	endLineNum := len(lines)
	
	// Handle line range if specified
	if startLine, ok := args["start_line"]; ok {
		start := int(startLine.(float64)) - 1 // Convert to 0-based
		end := len(lines)

		if endLineArg, ok := args["end_line"]; ok {
			end = int(endLineArg.(float64))
		}

		if start < 0 {
			start = 0
		}
		if start >= len(lines) {
			return &ToolResult{
				Content: "",
				Data: map[string]interface{}{
					"file_path":     filePath,
					"resolved_path": resolvedPath,
					"total_lines":   len(lines),
					"error":         "start_line exceeds file length",
				},
			}, nil
		}

		if end > len(lines) {
			end = len(lines)
		}
		if end <= start {
			end = start + 1
		}

		lines = lines[start:end]
		startLineNum = start + 1
		endLineNum = end
	}

	// Add line numbers to each line
	for i, line := range lines {
		lineNum := startLineNum + i
		formattedLines = append(formattedLines, fmt.Sprintf("%5d %s", lineNum, line))
	}

	contentStr = strings.Join(formattedLines, "\n")

	// Get file info
	fileInfo, _ := os.Stat(resolvedPath)

	return &ToolResult{
		Content: contentStr,
		Data: map[string]interface{}{
			"file_path":       filePath,
			"resolved_path":   resolvedPath,
			"file_size":       len(content),
			"lines":           len(strings.Split(string(content), "\n")),
			"modified":        fileInfo.ModTime().Unix(),
			"start_line":      startLineNum,
			"end_line":        endLineNum,
			"displayed_lines": len(lines),
		},
	}, nil
}