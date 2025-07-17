package builtin

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
		"required": []string{"file_path"},
	}
}

func (t *FileReadTool) Validate(args map[string]interface{}) error {
	// 检查必需的路径参数（支持 file_path 或 path）
	hasFilePath := false
	if _, ok := args["file_path"]; ok {
		hasFilePath = true
	}
	if _, ok := args["path"]; ok {
		hasFilePath = true
	}

	if !hasFilePath {
		return fmt.Errorf("missing required parameter: file_path or path")
	}

	validator := NewValidationFramework().
		AddOptionalStringField("file_path", "Path to the file to read").
		AddOptionalStringField("path", "Path to the file to read").
		AddOptionalIntField("start_line", "Starting line number (1-based)", 1, 0).
		AddOptionalIntField("end_line", "Ending line number (1-based)", 1, 0)

	return validator.Validate(args)
}

func (t *FileReadTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	// 支持两种参数名称：file_path 和 path
	var filePath string
	if path, ok := args["file_path"]; ok && path != nil {
		filePath = path.(string)
	} else if path, ok := args["path"]; ok && path != nil {
		filePath = path.(string)
	} else {
		return nil, fmt.Errorf("missing required parameter: file_path or path")
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

	// Handle line range if specified
	if startLine, ok := args["start_line"]; ok {
		lines := strings.Split(contentStr, "\n")
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

		selectedLines := lines[start:end]
		contentStr = strings.Join(selectedLines, "\n")
	}

	// Get file info
	fileInfo, _ := os.Stat(resolvedPath)

	return &ToolResult{
		Content: contentStr,
		Data: map[string]interface{}{
			"file_path":     filePath,
			"resolved_path": resolvedPath,
			"file_size":     len(content),
			"lines":         len(strings.Split(string(content), "\n")),
			"modified":      fileInfo.ModTime().Unix(),
		},
	}, nil
}

// FileUpdateTool implements file content updating functionality
type FileUpdateTool struct{}

func CreateFileUpdateTool() *FileUpdateTool {
	return &FileUpdateTool{}
}

func (t *FileUpdateTool) Name() string {
	return "file_edit"
}

func (t *FileUpdateTool) Description() string {
	return "This is a tool for editing files. For moving or renaming files, you should generally use the Bash tool with the 'mv' command instead. For larger edits, use the file_replace tool to overwrite files.\nBefore using this tool:\n1. Use the file_read tool to understand the file's contents and context\n2. Verify the directory path is correct (only applicable when creating new files):\n   - Use the file_list tool to verify the parent directory exists and is the correct location\nTo make a file edit, provide the following:\n1. file_path: The absolute path to the file to modify (must be absolute, not relative)\n2. old_string: The text to replace (must be unique within the file, and must match the file contents exactly, including all whitespace and indentation)\n3. new_string: The edited text to replace the old_string\nThe tool will replace ONE occurrence of old_string with new_string in the specified file.\nCRITICAL REQUIREMENTS FOR USING THIS TOOL:\n1. UNIQUENESS: The old_string MUST uniquely identify the specific instance you want to change. This means:\n   - Include AT LEAST 3-5 lines of context BEFORE the change point\n   - Include AT LEAST 3-5 lines of context AFTER the change point\n   - Include all whitespace, indentation, and surrounding code exactly as it appears in the file\n2. SINGLE INSTANCE: This tool can only change ONE instance at a time. If you need to change multiple instances:\n   - Make separate calls to this tool for each instance\n   - Each call must uniquely identify its specific instance using extensive context\n3. VERIFICATION: Before using this tool:\n   - Check how many instances of the target text exist in the file\n   - If multiple instances exist, gather enough context to uniquely identify each one\n   - Plan separate tool calls for each instance\nWARNING: If you do not follow these requirements:\n   - The tool will fail if old_string matches multiple locations\n   - The tool will fail if old_string doesn't match exactly (including whitespace)\n   - You may change the wrong instance if you don't include enough context\nWhen making edits:\n   - Ensure the edit results in idiomatic, correct code\n   - Do not leave the code in a broken state\n   - Always use absolute file paths (starting with /)\nIf you want to create a new file, use:\n   - A new file path, including dir name if needed\n   - An empty old_string\n   - The new file's contents as new_string\nRemember: when making multiple file edits in a row to the same file, you should prefer to send all edits in a single message with multiple calls to this tool, rather than multiple messages with a single call each."
}

func (t *FileUpdateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "The absolute path to the file to modify (must be absolute, not relative)",
			},
			"old_string": map[string]interface{}{
				"type":        "string",
				"description": "The text to replace (must be unique within the file, and must match the file contents exactly, including all whitespace and indentation)",
			},
			"new_string": map[string]interface{}{
				"type":        "string",
				"description": "The edited text to replace the old_string",
			},
		},
		"required": []string{"file_path", "old_string", "new_string"},
	}
}

func (t *FileUpdateTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("file_path", "The absolute path to the file to modify").
		AddCustomValidator("old_string", "The text to replace (can be empty for new file creation)", true, func(value interface{}) error {
			if _, ok := value.(string); ok {
				return nil // Allow empty string for new file creation
			}
			return fmt.Errorf("old_string must be a string")
		}).
		AddStringField("new_string", "The edited text to replace the old_string")

	return validator.Validate(args)
}

func (t *FileUpdateTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	filePath := args["file_path"].(string)
	oldString := args["old_string"].(string)
	newString := args["new_string"].(string)

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
		return nil, fmt.Errorf("old_string not found in file: %s", oldString)
	}
	if occurrences > 1 {
		return nil, fmt.Errorf("old_string appears %d times in file. Please include more context to make it unique", occurrences)
	}

	// Perform the replacement (only one occurrence)
	newContent := strings.Replace(originalContent, oldString, newString, 1)

	// Calculate diff statistics
	diffStats := calculateDiffStats(originalContent, newContent)

	// Write the modified content
	err = os.WriteFile(resolvedPath, []byte(newContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Get file info after writing
	fileInfo, _ := os.Stat(resolvedPath)

	// Create enhanced output message
	var contentBuilder strings.Builder
	newLineCount := len(strings.Split(newContent, "\n"))
	contentBuilder.WriteString(fmt.Sprintf("Updated %s with %d additions and %d deletions (%d lines total)\n", filePath, diffStats.Additions, diffStats.Deletions, newLineCount))
	
	if diffStats.DiffPreview != "" {
		contentBuilder.WriteString("\nDiff preview:\n")
		contentBuilder.WriteString(diffStats.DiffPreview)
	}

	return &ToolResult{
		Content: contentBuilder.String(),
		Files:   []string{resolvedPath},
		Data: map[string]interface{}{
			"file_path":         filePath,
			"resolved_path":     resolvedPath,
			"operation":         "edited",
			"old_string":        oldString,
			"new_string":        newString,
			"bytes_written":     len(newContent),
			"lines_total":       len(strings.Split(newContent, "\n")),
			"modified":          fileInfo.ModTime().Unix(),
			"replacements_made": 1,
			"additions":         diffStats.Additions,
			"deletions":         diffStats.Deletions,
			"diff_preview":      diffStats.DiffPreview,
		},
	}, nil
}

// FileReplaceTool implements file content replacement functionality
type FileReplaceTool struct{}

func CreateFileReplaceTool() *FileReplaceTool {
	return &FileReplaceTool{}
}

func (t *FileReplaceTool) Name() string {
	return "file_replace"
}

func (t *FileReplaceTool) Description() string {
	return "Write a file to the local filesystem. Overwrites the existing file if there is one.\nBefore using this tool:\n1. Use the file_read tool to understand the file's contents and context\n2. Directory Verification (only applicable when creating new files):\n   - Use the file_list tool to verify the parent directory exists and is the correct location"
}

func (t *FileReplaceTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "The absolute path to the file to write (must be absolute, not relative)",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The content to write to the file",
			},
		},
		"required": []string{"file_path", "content"},
	}
}

func (t *FileReplaceTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("file_path", "The absolute path to the file to write").
		AddCustomValidator("content", "The content to write to the file (can be empty)", true, func(value interface{}) error {
			if _, ok := value.(string); ok {
				return nil // Allow empty string for empty files
			}
			return fmt.Errorf("content must be a string")
		})

	return validator.Validate(args)
}

func (t *FileReplaceTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	filePath := args["file_path"].(string)
	content := args["content"].(string)

	// 解析路径（处理相对路径）
	resolver := GetPathResolverFromContext(ctx)
	resolvedPath := resolver.ResolvePath(filePath)

	// Create parent directories if needed
	dir := filepath.Dir(resolvedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Check if file exists to determine operation type and get original content
	var operation string
	var originalContent string
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		operation = "created"
		originalContent = ""
	} else {
		operation = "overwritten"
		if existingContent, err := os.ReadFile(resolvedPath); err == nil {
			originalContent = string(existingContent)
		}
	}

	// Calculate diff statistics
	diffStats := calculateDiffStats(originalContent, content)

	// Write the content to file (overwrites if exists)
	err := os.WriteFile(resolvedPath, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Get file info after writing
	fileInfo, _ := os.Stat(resolvedPath)

	// Create enhanced output message
	var contentBuilder strings.Builder
	newLineCount := len(strings.Split(content, "\n"))
	if operation == "created" {
		contentBuilder.WriteString(fmt.Sprintf("Created %s with %d lines", filePath, newLineCount))
	} else {
		contentBuilder.WriteString(fmt.Sprintf("Updated %s with %d additions and %d deletions (%d lines total)", filePath, diffStats.Additions, diffStats.Deletions, newLineCount))
	}
	
	if diffStats.DiffPreview != "" && operation == "overwritten" {
		contentBuilder.WriteString("\n\nDiff preview:\n")
		contentBuilder.WriteString(diffStats.DiffPreview)
	}

	return &ToolResult{
		Content: contentBuilder.String(),
		Files:   []string{resolvedPath},
		Data: map[string]interface{}{
			"file_path":     filePath,
			"resolved_path": resolvedPath,
			"operation":     operation,
			"bytes_written": len(content),
			"lines_total":   len(strings.Split(content, "\n")),
			"modified":      fileInfo.ModTime().Unix(),
			"additions":     diffStats.Additions,
			"deletions":     diffStats.Deletions,
			"diff_preview":  diffStats.DiffPreview,
		},
	}, nil
}

// FileListTool implements directory listing functionality
type FileListTool struct{}

func CreateFileListTool() *FileListTool {
	return &FileListTool{}
}

func (t *FileListTool) Name() string {
	return "file_list"
}

func (t *FileListTool) Description() string {
	return "List files and directories in a specified path. Supports glob patterns and recursive listing."
}

func (t *FileListTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to list (directory or glob pattern)",
				"default":     ".",
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"description": "List files recursively",
				"default":     false,
			},
			"show_hidden": map[string]interface{}{
				"type":        "boolean",
				"description": "Include hidden files (starting with .)",
				"default":     false,
			},
		},
	}
}

func (t *FileListTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddOptionalStringField("path", "Path to list (directory or glob pattern)").
		AddBoolField("recursive", "List files recursively", false).
		AddBoolField("show_hidden", "Include hidden files", false)

	return validator.Validate(args)
}

func (t *FileListTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	path := "."
	if pathArg, ok := args["path"]; ok {
		path = pathArg.(string)
	}

	// 解析路径（处理相对路径）
	resolver := GetPathResolverFromContext(ctx)
	resolvedPath := resolver.ResolvePath(path)

	recursive := false
	if recursiveArg, ok := args["recursive"]; ok {
		recursive, _ = recursiveArg.(bool)
	}

	depth := 3
	if depthArg, ok := args["depth"]; ok {
		if d, ok := depthArg.(float64); ok {
			depth = int(d)
		} else if d, ok := depthArg.(int); ok {
			depth = d
		}
		if depth < 1 {
			depth = 1
		} else if depth > 3 {
			depth = 3
		}
	}

	showHidden := false
	if showHiddenArg, ok := args["show_hidden"]; ok {
		showHidden, _ = showHiddenArg.(bool)
	}

	var fileTypes []string
	if fileTypesArg, ok := args["file_types"]; ok {
		if fileTypeSlice, ok := fileTypesArg.([]interface{}); ok {
			for _, ft := range fileTypeSlice {
				fileTypes = append(fileTypes, ft.(string))
			}
		}
	}

	var files []map[string]interface{}
	var totalSize int64

	if recursive {
		basePath, err := filepath.Abs(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}

		err = filepath.WalkDir(resolvedPath, func(currentPath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Calculate current depth relative to base path
			absCurrentPath, err := filepath.Abs(currentPath)
			if err != nil {
				return err
			}
			relPath, err := filepath.Rel(basePath, absCurrentPath)
			if err != nil {
				return err
			}

			currentDepth := 1
			if relPath != "." && relPath != "" && relPath != string(filepath.Separator) {
				// Count path separators to determine depth
				parts := strings.Split(relPath, string(filepath.Separator))
				// Filter out empty parts
				validParts := make([]string, 0, len(parts))
				for _, part := range parts {
					if part != "" {
						validParts = append(validParts, part)
					}
				}
				currentDepth = len(validParts) + 1
			}

			// Skip if beyond specified depth
			if currentDepth > depth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip hidden files if not requested (but not the root directory itself)
			if !showHidden && strings.HasPrefix(d.Name(), ".") && currentPath != resolvedPath {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Filter by file types if specified
			if len(fileTypes) > 0 && !d.IsDir() {
				ext := filepath.Ext(d.Name())
				match := false
				for _, ft := range fileTypes {
					if ext == ft {
						match = true
						break
					}
				}
				if !match {
					return nil
				}
			}

			info, err := d.Info()
			if err != nil {
				return err
			}

			fileInfo := map[string]interface{}{
				"name":     d.Name(),
				"path":     currentPath,
				"rel_path": relPath,
				"is_dir":   d.IsDir(),
				"size":     info.Size(),
				"mode":     info.Mode().String(),
				"modified": info.ModTime().Unix(),
				"depth":    currentDepth,
			}

			files = append(files, fileInfo)
			if !d.IsDir() {
				totalSize += info.Size()
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		entries, err := os.ReadDir(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			// Skip hidden files if not requested
			if !showHidden && strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			// Filter by file types if specified
			if len(fileTypes) > 0 && !entry.IsDir() {
				ext := filepath.Ext(entry.Name())
				match := false
				for _, ft := range fileTypes {
					if ext == ft {
						match = true
						break
					}
				}
				if !match {
					continue
				}
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			fileInfo := map[string]interface{}{
				"name":     entry.Name(),
				"path":     filepath.Join(resolvedPath, entry.Name()),
				"rel_path": entry.Name(),
				"is_dir":   entry.IsDir(),
				"size":     info.Size(),
				"mode":     info.Mode().String(),
				"modified": info.ModTime().Unix(),
				"depth":    1,
			}

			files = append(files, fileInfo)
			if !entry.IsDir() {
				totalSize += info.Size()
			}
		}
	}

	// Generate detailed content and summary
	dirCount := 0
	fileCount := 0
	for _, file := range files {
		if file["is_dir"].(bool) {
			dirCount++
		} else {
			fileCount++
		}
	}

	// Build detailed content with file listing
	var contentBuilder strings.Builder

	// Header with summary
	summary := fmt.Sprintf("Found %d files and %d directories", fileCount, dirCount)
	if totalSize > 0 {
		summary += fmt.Sprintf(" (total size: %s)", formatFileSize(totalSize))
	}
	if recursive {
		summary += fmt.Sprintf(" (depth: %d)", depth)
	}
	contentBuilder.WriteString(summary + "\n\n")

	// Detailed file listing
	if len(files) > 0 {
		contentBuilder.WriteString("Detailed listing:\n")

		// Sort files: directories first, then files, both alphabetically
		var directories []map[string]interface{}
		var regularFiles []map[string]interface{}

		for _, file := range files {
			if file["is_dir"].(bool) {
				directories = append(directories, file)
			} else {
				regularFiles = append(regularFiles, file)
			}
		}

		// List directories first
		if len(directories) > 0 {
			contentBuilder.WriteString("\nDirectories:\n")
			for _, dir := range directories {
				relPath := dir["rel_path"].(string)
				if relPath == "." || relPath == "" {
					relPath = dir["name"].(string)
				}
				mode := dir["mode"].(string)
				contentBuilder.WriteString(fmt.Sprintf("  📁 %s/ [%s]\n", relPath, mode))
			}
		}

		// List files
		if len(regularFiles) > 0 {
			contentBuilder.WriteString("\nFiles:\n")
			for _, file := range regularFiles {
				relPath := file["rel_path"].(string)
				if relPath == "." || relPath == "" {
					relPath = file["name"].(string)
				}
				size := file["size"].(int64)
				mode := file["mode"].(string)
				sizeStr := formatFileSize(size)

				// Get file extension for type indication
				ext := strings.ToLower(filepath.Ext(file["name"].(string)))
				icon := getFileIcon(ext)

				contentBuilder.WriteString(fmt.Sprintf("  %s %s (%s) [%s]\n", icon, relPath, sizeStr, mode))
			}
		}
	} else {
		contentBuilder.WriteString("\nNo files found matching the criteria.\n")
	}

	return &ToolResult{
		Content: contentBuilder.String(),
		Data: map[string]interface{}{
			"path":          path,
			"resolved_path": resolvedPath,
			"files":         files,
			"file_count":    fileCount,
			"dir_count":     dirCount,
			"total_size":    totalSize,
			"recursive":     recursive,
			"depth":         depth,
		},
	}, nil
}

// Helper function to format file sizes
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Helper function to get file icon based on extension
func getFileIcon(ext string) string {
	switch ext {
	case ".go":
		return "🐹"
	case ".ts", ".tsx":
		return "🟦"
	case ".js", ".jsx":
		return "🟨"
	case ".py":
		return "🐍"
	case ".java":
		return "☕"
	case ".cpp", ".c", ".h":
		return "🔧"
	case ".rs":
		return "🦀"
	case ".md":
		return "📝"
	case ".json":
		return "📋"
	case ".yml", ".yaml":
		return "⚙️"
	case ".toml":
		return "🔧"
	case ".xml":
		return "📄"
	case ".html":
		return "🌐"
	case ".css":
		return "🎨"
	case ".sql":
		return "🗃️"
	case ".sh":
		return "🔨"
	case ".bat":
		return "⚙️"
	case ".dockerfile":
		return "🐳"
	case ".txt":
		return "📃"
	case ".log":
		return "📊"
	case ".zip", ".tar", ".gz":
		return "📦"
	case ".png", ".jpg", ".jpeg", ".gif", ".svg":
		return "🖼️"
	case ".pdf":
		return "📕"
	case ".exe":
		return "⚡"
	default:
		return "📄"
	}
}

// DiffStats represents the statistics of a diff operation
type DiffStats struct {
	Additions int
	Deletions int
	DiffPreview string
}

// calculateDiffStats calculates the additions and deletions between old and new content
func calculateDiffStats(oldContent, newContent string) DiffStats {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")
	
	// Simple diff algorithm - count line additions and deletions
	additions := 0
	deletions := 0
	
	// Use a simple LCS approach to calculate diff
	oldSet := make(map[string]bool)
	newSet := make(map[string]bool)
	
	for _, line := range oldLines {
		oldSet[line] = true
	}
	
	for _, line := range newLines {
		newSet[line] = true
	}
	
	// Count additions (lines in new but not in old)
	for _, line := range newLines {
		if !oldSet[line] {
			additions++
		}
	}
	
	// Count deletions (lines in old but not in new)
	for _, line := range oldLines {
		if !newSet[line] {
			deletions++
		}
	}
	
	// Generate diff preview
	diffPreview := generateDiffPreview(oldLines, newLines)
	
	return DiffStats{
		Additions: additions,
		Deletions: deletions,
		DiffPreview: diffPreview,
	}
}

// generateDiffPreview generates a git-style diff preview
func generateDiffPreview(oldLines, newLines []string) string {
	var diffBuilder strings.Builder
	
	// Find the actual changed section
	oldSet := make(map[string]int)
	newSet := make(map[string]int)
	
	for i, line := range oldLines {
		oldSet[line] = i
	}
	
	for i, line := range newLines {
		newSet[line] = i
	}
	
	// Simple diff display - show first few changes
	changeCount := 0
	maxChanges := 5
	
	// Show deletions
	for _, line := range oldLines {
		if _, exists := newSet[line]; !exists && changeCount < maxChanges {
			diffBuilder.WriteString(fmt.Sprintf("-%s\n", line))
			changeCount++
		}
	}
	
	// Show additions
	for _, line := range newLines {
		if _, exists := oldSet[line]; !exists && changeCount < maxChanges {
			diffBuilder.WriteString(fmt.Sprintf("+%s\n", line))
			changeCount++
		}
	}
	
	preview := diffBuilder.String()
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	
	return preview
}
