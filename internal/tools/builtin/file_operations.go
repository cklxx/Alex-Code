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
	return "file_update"
}

func (t *FileUpdateTool) Description() string {
	return "Update a file by appending content, inserting at specific lines, or creating new files. Preserves existing content."
}

func (t *FileUpdateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to update",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to add to the file",
			},
			"mode": map[string]interface{}{
				"type":        "string",
				"description": "Update mode: append, prepend, create",
				"enum":        []string{"append", "prepend", "create"},
				"default":     "append",
			},
			"create_dirs": map[string]interface{}{
				"type":        "boolean",
				"description": "Create parent directories if they don't exist",
				"default":     true,
			},
		},
		"required": []string{"file_path", "content"},
	}
}

func (t *FileUpdateTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("file_path", "Path to the file to update").
		AddStringField("content", "Content to add to the file").
		AddCustomValidator("mode", "Update mode (append, prepend, create)", false, func(value interface{}) error {
			if value == nil {
				return nil // Optional field
			}
			modeStr, ok := value.(string)
			if !ok {
				return fmt.Errorf("mode must be a string")
			}
			validModes := []string{"append", "prepend", "create"}
			for _, vm := range validModes {
				if modeStr == vm {
					return nil
				}
			}
			return fmt.Errorf("mode must be one of: %v", validModes)
		}).
		AddBoolField("create_dirs", "Create parent directories if they don't exist", false)

	return validator.Validate(args)
}

func (t *FileUpdateTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	filePath := args["file_path"].(string)
	content := args["content"].(string)

	// 解析路径（处理相对路径）
	resolver := GetPathResolverFromContext(ctx)
	resolvedPath := resolver.ResolvePath(filePath)

	mode := "append"
	if modeArg, ok := args["mode"]; ok {
		mode, _ = modeArg.(string)
	}

	createDirs := false
	if createDirsArg, ok := args["create_dirs"]; ok {
		createDirs, _ = createDirsArg.(bool)
	}

	backup := false
	if backupArg, ok := args["backup"]; ok {
		backup, _ = backupArg.(bool)
	}

	// Create parent directories if requested
	if createDirs {
		dir := filepath.Dir(resolvedPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directories: %w", err)
		}
	}

	var finalContent string
	var operation string

	switch mode {
	case "create":
		// Check if file already exists
		if _, err := os.Stat(resolvedPath); err == nil {
			return nil, fmt.Errorf("file already exists: %s", filePath)
		}
		finalContent = content
		operation = "created"

	case "append":
		existingContent := ""
		if data, err := os.ReadFile(resolvedPath); err == nil {
			existingContent = string(data)
		}
		finalContent = existingContent + content
		operation = "appended"

	case "prepend":
		existingContent := ""
		if data, err := os.ReadFile(resolvedPath); err == nil {
			existingContent = string(data)
		}
		finalContent = content + existingContent
		operation = "prepended"

	case "insert":
		existingContent := ""
		if data, err := os.ReadFile(resolvedPath); err == nil {
			existingContent = string(data)
		}

		lineNumber := int(args["line_number"].(float64))
		lines := strings.Split(existingContent, "\n")

		// Insert at specified line (1-based)
		insertPos := lineNumber - 1
		if insertPos < 0 {
			insertPos = 0
		}
		if insertPos > len(lines) {
			insertPos = len(lines)
		}

		// Split content into lines for insertion
		newLines := strings.Split(content, "\n")

		// Build final content
		result := make([]string, 0, len(lines)+len(newLines))
		result = append(result, lines[:insertPos]...)
		result = append(result, newLines...)
		if insertPos < len(lines) {
			result = append(result, lines[insertPos:]...)
		}
		finalContent = strings.Join(result, "\n")
		operation = fmt.Sprintf("inserted at line %d", lineNumber)
	}

	// Create backup if requested
	if backup && mode != "create" {
		backupPath := resolvedPath + ".bak"
		if existingData, err := os.ReadFile(resolvedPath); err == nil {
			if err := os.WriteFile(backupPath, existingData, 0644); err != nil {
				return nil, fmt.Errorf("failed to create backup: %w", err)
			}
		}
	}

	// Write the final content
	err := os.WriteFile(resolvedPath, []byte(finalContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to update file: %w", err)
	}

	// Get file info after writing
	fileInfo, _ := os.Stat(resolvedPath)

	return &ToolResult{
		Content: fmt.Sprintf("Successfully %s content to %s (%d bytes)", operation, filePath, len(finalContent)),
		Files:   []string{resolvedPath},
		Data: map[string]interface{}{
			"file_path":     filePath,
			"resolved_path": resolvedPath,
			"operation":     operation,
			"bytes_written": len(finalContent),
			"lines_total":   len(strings.Split(finalContent, "\n")),
			"modified":      fileInfo.ModTime().Unix(),
			"mode":          mode,
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
	return "Replace specific text content in a file. Performs case-sensitive exact text matching."
}

func (t *FileReplaceTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to modify",
			},
			"search": map[string]interface{}{
				"type":        "string",
				"description": "Text to search for",
			},
			"replace": map[string]interface{}{
				"type":        "string",
				"description": "Replacement text",
			},
		},
		"required": []string{"file_path", "search", "replace"},
	}
}

func (t *FileReplaceTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("file_path", "Path to the file to modify").
		AddStringField("search", "Text to search for").
		AddStringField("replace", "Replacement text")

	return validator.Validate(args)
}

func (t *FileReplaceTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	filePath := args["file_path"].(string)
	search := args["search"].(string)
	replace := args["replace"].(string)

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

	originalContent := string(content)
	
	// Simple case-sensitive first occurrence replacement
	newContent := strings.Replace(originalContent, search, replace, 1)
	replacementCount := 0
	if newContent != originalContent {
		replacementCount = 1
	}

	result := &ToolResult{
		Content: fmt.Sprintf("File %s: replaced %d occurrence(s) of '%s' with '%s'", filePath, replacementCount, search, replace),
		Data: map[string]interface{}{
			"file_path":         filePath,
			"resolved_path":     resolvedPath,
			"search_pattern":    search,
			"replacement":       replace,
			"replacements_made": replacementCount,
		},
	}

	// No changes needed
	if replacementCount == 0 {
		return result, nil
	}

	// Write the modified content
	err = os.WriteFile(resolvedPath, []byte(newContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Get file info after writing
	fileInfo, _ := os.Stat(resolvedPath)
	result.Files = []string{resolvedPath}
	result.Data["bytes_written"] = len(newContent)
	result.Data["modified"] = fileInfo.ModTime().Unix()

	return result, nil
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

// DirectoryCreateTool implements directory creation functionality
type DirectoryCreateTool struct{}

func CreateDirectoryCreateTool() *DirectoryCreateTool {
	return &DirectoryCreateTool{}
}

func (t *DirectoryCreateTool) Name() string {
	return "directory_create"
}

func (t *DirectoryCreateTool) Description() string {
	return "Create a directory at the specified path. Creates parent directories if they don't exist."
}

func (t *DirectoryCreateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path for the new directory",
			},
		},
		"required": []string{"path"},
	}
}

func (t *DirectoryCreateTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("path", "Path for the new directory")

	return validator.Validate(args)
}

func (t *DirectoryCreateTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	path := args["path"].(string)

	// 解析路径（处理相对路径）
	resolver := GetPathResolverFromContext(ctx)
	resolvedPath := resolver.ResolvePath(path)

	permissions := os.FileMode(0755) // Default permissions

	if permStr, ok := args["permissions"].(string); ok {
		// Parse octal permissions string (e.g., "755" -> 0755)
		switch permStr {
		case "755":
			permissions = 0755
		case "644":
			permissions = 0644
		case "777":
			permissions = 0777
		}
		// Default to 0755 for any other values
	}

	// Create the directory with all parent directories
	err := os.MkdirAll(resolvedPath, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Get directory info
	dirInfo, err := os.Stat(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat created directory: %w", err)
	}

	return &ToolResult{
		Content: fmt.Sprintf("Successfully created directory: %s", path),
		Data: map[string]interface{}{
			"path":          path,
			"resolved_path": resolvedPath,
			"permissions":   dirInfo.Mode().String(),
			"created":       dirInfo.ModTime().Unix(),
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

