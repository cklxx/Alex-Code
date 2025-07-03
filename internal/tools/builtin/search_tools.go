package builtin

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// GrepTool implements grep-style search functionality
type GrepTool struct{}

func CreateGrepTool() *GrepTool {
	return &GrepTool{}
}

func (t *GrepTool) Name() string {
	return "grep"
}

func (t *GrepTool) Description() string {
	return "Search for patterns in files using grep-style functionality. Supports regex patterns, line numbers, and context."
}

func (t *GrepTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Search pattern (supports regex)",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory path to search in",
				"default":     ".",
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"description": "Search recursively (-r flag)",
				"default":     false,
			},
			"ignore_case": map[string]interface{}{
				"type":        "boolean",
				"description": "Case insensitive search (-i flag)",
				"default":     false,
			},
			"line_numbers": map[string]interface{}{
				"type":        "boolean",
				"description": "Show line numbers (-n flag)",
				"default":     true,
			},
			"context_lines": map[string]interface{}{
				"type":        "integer",
				"description": "Lines of context around matches",
				"default":     0,
				"minimum":     0,
				"maximum":     10,
			},
			"max_matches": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of matches to return",
				"default":     100,
				"minimum":     1,
				"maximum":     1000,
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GrepTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("pattern", "Search pattern (supports regex)").
		AddOptionalStringField("path", "Directory path to search in").
		AddBoolField("recursive", "Search recursively", false).
		AddBoolField("ignore_case", "Case insensitive search", false).
		AddBoolField("line_numbers", "Show line numbers", false).
		AddOptionalIntField("context_lines", "Lines of context around matches", 0, 10).
		AddOptionalIntField("max_matches", "Maximum number of matches", 1, 1000)

	// First run standard validation
	if err := validator.Validate(args); err != nil {
		return err
	}

	// Validate regex pattern
	pattern := args["pattern"].(string)
	if _, err := regexp.Compile(pattern); err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	return nil
}

func (t *GrepTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	pattern := args["pattern"].(string)

	// Get search path
	searchPath := "."
	if pathArg, ok := args["path"]; ok {
		searchPath = pathArg.(string)
	}

	recursive := false
	if recursiveArg, ok := args["recursive"]; ok {
		recursive, _ = recursiveArg.(bool)
	}

	ignoreCase := false
	if ignoreCaseArg, ok := args["ignore_case"]; ok {
		ignoreCase, _ = ignoreCaseArg.(bool)
	}

	lineNumbers := true
	if lineNumbersArg, ok := args["line_numbers"]; ok {
		lineNumbers, _ = lineNumbersArg.(bool)
	}

	contextLines := 0
	if contextArg, ok := args["context_lines"]; ok {
		if c, ok := contextArg.(float64); ok {
			contextLines = int(c)
		}
	}

	maxMatches := 100
	if maxArg, ok := args["max_matches"]; ok {
		if m, ok := maxArg.(float64); ok {
			maxMatches = int(m)
		}
	}

	// Try to use system grep if available (faster for large searches)
	if t.hasSystemGrep() && contextLines == 0 {
		result, err := t.useSystemGrep(pattern, searchPath, recursive, ignoreCase, lineNumbers, maxMatches)
		if err == nil {
			return result, nil
		}
		// Fall back to native implementation if system grep fails
	}

	// Native Go implementation
	return t.nativeGrep(pattern, searchPath, recursive, ignoreCase, lineNumbers, contextLines, maxMatches)
}

func (t *GrepTool) hasSystemGrep() bool {
	if runtime.GOOS == "windows" {
		return false // Windows doesn't have grep by default
	}
	_, err := exec.LookPath("grep")
	return err == nil
}

func (t *GrepTool) useSystemGrep(pattern, searchPath string, recursive, ignoreCase, lineNumbers bool, maxMatches int) (*ToolResult, error) {
	args := []string{}

	if recursive {
		args = append(args, "-r")
	}
	if ignoreCase {
		args = append(args, "-i")
	}
	if lineNumbers {
		args = append(args, "-n")
	}

	args = append(args, pattern, searchPath)

	cmd := exec.Command("grep", args...)
	output, err := cmd.Output()

	// grep returns exit code 1 when no matches found, which is not an error for us
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// No matches found
			return &ToolResult{
				Content: "No matches found",
				Data: map[string]interface{}{
					"pattern":     pattern,
					"matches":     []string{},
					"match_count": 0,
					"method":      "system_grep",
				},
			}, nil
		}
		return nil, fmt.Errorf("grep command failed: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))
	lines := strings.Split(outputStr, "\n")

	// Limit results if needed
	if len(lines) > maxMatches {
		lines = lines[:maxMatches]
	}

	return &ToolResult{
		Content: fmt.Sprintf("Found %d matches:\n%s", len(lines), strings.Join(lines, "\n")),
		Data: map[string]interface{}{
			"pattern":     pattern,
			"matches":     lines,
			"match_count": len(lines),
			"method":      "system_grep",
			"truncated":   len(lines) >= maxMatches,
		},
	}, nil
}

func (t *GrepTool) nativeGrep(pattern, searchPath string, recursive, ignoreCase, lineNumbers bool, contextLines, maxMatches int) (*ToolResult, error) {
	// Compile regex pattern
	var re *regexp.Regexp
	var err error
	if ignoreCase {
		re, err = regexp.Compile("(?i)" + pattern)
	} else {
		re, err = regexp.Compile(pattern)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", err)
	}

	var matches []map[string]interface{}
	matchCount := 0

	searchFunc := func(filePath string, info os.FileInfo, err error) error {
		if err != nil || matchCount >= maxMatches {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil // Skip files that can't be read
		}

		lines := strings.Split(string(content), "\n")

		for i, line := range lines {
			if re.MatchString(line) && matchCount < maxMatches {
				match := map[string]interface{}{
					"file":        filePath,
					"line_number": i + 1,
					"line":        line,
				}

				// Add context if requested
				if contextLines > 0 {
					var contextBefore []string
					var contextAfter []string

					// Before context
					for j := i - contextLines; j < i; j++ {
						if j >= 0 {
							contextBefore = append(contextBefore, lines[j])
						}
					}

					// After context
					for j := i + 1; j <= i+contextLines; j++ {
						if j < len(lines) {
							contextAfter = append(contextAfter, lines[j])
						}
					}

					match["context_before"] = contextBefore
					match["context_after"] = contextAfter
				}

				matches = append(matches, match)
				matchCount++
			}
		}

		return nil
	}

	if recursive {
		err = filepath.Walk(searchPath, searchFunc)
	} else {
		entries, err := os.ReadDir(searchPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				info, err := entry.Info()
				if err == nil {
					filePath := filepath.Join(searchPath, entry.Name())
					if err := searchFunc(filePath, info, nil); err != nil {
						log.Printf("Error in search function: %v", err)
					}
				}
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("search error: %w", err)
	}

	// Generate output
	var contentBuilder strings.Builder
	if len(matches) == 0 {
		contentBuilder.WriteString("No matches found")
	} else {
		contentBuilder.WriteString(fmt.Sprintf("Found %d matches for pattern '%s'\n\n", len(matches), pattern))

		for i, match := range matches {
			if i > 0 {
				contentBuilder.WriteString("\n")
			}

			file := match["file"].(string)
			lineNum := match["line_number"].(int)
			line := match["line"].(string)

			if lineNumbers {
				contentBuilder.WriteString(fmt.Sprintf("%s:%d:%s\n", file, lineNum, line))
			} else {
				contentBuilder.WriteString(fmt.Sprintf("%s:%s\n", file, line))
			}

			// Add context if available
			if contextBefore, ok := match["context_before"].([]string); ok && len(contextBefore) > 0 {
				for j, ctx := range contextBefore {
					ctxLineNum := lineNum - len(contextBefore) + j
					contentBuilder.WriteString(fmt.Sprintf("  %d: %s\n", ctxLineNum, ctx))
				}
			}

			if contextAfter, ok := match["context_after"].([]string); ok && len(contextAfter) > 0 {
				for j, ctx := range contextAfter {
					ctxLineNum := lineNum + j + 1
					contentBuilder.WriteString(fmt.Sprintf("  %d: %s\n", ctxLineNum, ctx))
				}
			}
		}
	}

	return &ToolResult{
		Content: contentBuilder.String(),
		Data: map[string]interface{}{
			"pattern":     pattern,
			"matches":     matches,
			"match_count": len(matches),
			"method":      "native_grep",
			"truncated":   matchCount >= maxMatches,
		},
	}, nil
}

// RipgrepTool implements ripgrep-style search (if available)
type RipgrepTool struct{}

func CreateRipgrepTool() *RipgrepTool {
	return &RipgrepTool{}
}

func (t *RipgrepTool) Name() string {
	return "ripgrep"
}

func (t *RipgrepTool) Description() string {
	return "Fast recursive search using ripgrep (rg). Falls back to native search if ripgrep is not available."
}

func (t *RipgrepTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Search pattern (supports regex)",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory path to search in",
				"default":     ".",
			},
			"ignore_case": map[string]interface{}{
				"type":        "boolean",
				"description": "Case insensitive search",
				"default":     false,
			},
			"type_filter": map[string]interface{}{
				"type":        "string",
				"description": "File type filter (e.g., 'go', 'js', 'py')",
			},
			"glob": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern for files to search",
			},
			"context": map[string]interface{}{
				"type":        "integer",
				"description": "Lines of context around matches",
				"default":     0,
				"minimum":     0,
				"maximum":     10,
			},
			"max_count": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of matches per file",
				"default":     0,
				"minimum":     0,
			},
			"hidden": map[string]interface{}{
				"type":        "boolean",
				"description": "Search hidden files and directories",
				"default":     false,
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *RipgrepTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("pattern", "Search pattern (supports regex)").
		AddOptionalStringField("path", "Directory path to search in").
		AddBoolField("ignore_case", "Case insensitive search", false).
		AddOptionalStringField("type_filter", "File type filter (e.g., 'go', 'js', 'py')").
		AddOptionalStringField("glob", "Glob pattern for files to search").
		AddOptionalIntField("context", "Lines of context around matches", 0, 10).
		AddOptionalIntField("max_count", "Maximum number of matches per file", 0, 0).
		AddBoolField("hidden", "Search hidden files and directories", false)

	return validator.Validate(args)
}

func (t *RipgrepTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	// Check if ripgrep is available
	if !t.hasRipgrep() {
		// Fall back to grep tool
		grepTool := CreateGrepTool()
		grepArgs := map[string]interface{}{
			"pattern":   args["pattern"],
			"recursive": true,
		}

		// Map some common parameters
		if path, ok := args["path"]; ok {
			grepArgs["path"] = path
		}
		if ignoreCase, ok := args["ignore_case"]; ok {
			grepArgs["ignore_case"] = ignoreCase
		}
		if context, ok := args["context"]; ok {
			grepArgs["context_lines"] = context
		}

		result, err := grepTool.Execute(ctx, grepArgs)
		if err != nil {
			return nil, err
		}

		// Update metadata to indicate fallback
		if result.Data == nil {
			result.Data = make(map[string]interface{})
		}
		result.Data["method"] = "fallback_grep"
		result.Data["note"] = "ripgrep not available, used grep fallback"

		return result, nil
	}

	// Use ripgrep
	return t.useRipgrep(args)
}

func (t *RipgrepTool) hasRipgrep() bool {
	_, err := exec.LookPath("rg")
	return err == nil
}

func (t *RipgrepTool) useRipgrep(args map[string]interface{}) (*ToolResult, error) {
	pattern := args["pattern"].(string)

	rgArgs := []string{
		"--line-number",
		"--heading",
	}

	// Add options based on parameters
	if ignoreCase, ok := args["ignore_case"]; ok && ignoreCase.(bool) {
		rgArgs = append(rgArgs, "--ignore-case")
	}

	if typeFilter, ok := args["type_filter"]; ok {
		rgArgs = append(rgArgs, "--type", typeFilter.(string))
	}

	if glob, ok := args["glob"]; ok {
		rgArgs = append(rgArgs, "--glob", glob.(string))
	}

	if context, ok := args["context"]; ok {
		if c, ok := context.(float64); ok && c > 0 {
			rgArgs = append(rgArgs, "--context", fmt.Sprintf("%.0f", c))
		}
	}

	if maxCount, ok := args["max_count"]; ok {
		if mc, ok := maxCount.(float64); ok && mc > 0 {
			rgArgs = append(rgArgs, "--max-count", fmt.Sprintf("%.0f", mc))
		}
	}

	if hidden, ok := args["hidden"]; ok && hidden.(bool) {
		rgArgs = append(rgArgs, "--hidden")
	}

	// Add pattern and path
	rgArgs = append(rgArgs, pattern)

	if path, ok := args["path"]; ok {
		rgArgs = append(rgArgs, path.(string))
	} else {
		rgArgs = append(rgArgs, ".")
	}

	cmd := exec.Command("rg", rgArgs...)
	output, err := cmd.Output()

	// ripgrep returns exit code 1 when no matches found
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return &ToolResult{
				Content: "No matches found",
				Data: map[string]interface{}{
					"pattern":     pattern,
					"matches":     []string{},
					"match_count": 0,
					"method":      "ripgrep",
				},
			}, nil
		}
		return nil, fmt.Errorf("ripgrep command failed: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))
	lines := strings.Split(outputStr, "\n")

	return &ToolResult{
		Content: fmt.Sprintf("Found %d matches:\n%s", len(lines), outputStr),
		Data: map[string]interface{}{
			"pattern":     pattern,
			"matches":     lines,
			"match_count": len(lines),
			"method":      "ripgrep",
		},
	}, nil
}

// FindTool implements find-style file discovery
type FindTool struct{}

func CreateFindTool() *FindTool {
	return &FindTool{}
}

func (t *FindTool) Name() string {
	return "find"
}

func (t *FindTool) Description() string {
	return "Find files and directories by name, type, or other attributes. Similar to Unix find command."
}

func (t *FindTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Starting directory for search",
				"default":     ".",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "File/directory name pattern (supports wildcards)",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Type of entry to find",
				"enum":        []string{"f", "d", "l", "file", "dir", "link"},
			},
			"extension": map[string]interface{}{
				"type":        "string",
				"description": "File extension to search for (e.g., '.go', '.js')",
			},
			"max_depth": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum depth to search",
				"default":     -1,
				"minimum":     1,
				"maximum":     20,
			},
			"case_insensitive": map[string]interface{}{
				"type":        "boolean",
				"description": "Case insensitive name matching",
				"default":     false,
			},
		},
	}
}

func (t *FindTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddOptionalStringField("path", "Starting directory for search").
		AddOptionalStringField("name", "File/directory name pattern (supports wildcards)").
		AddCustomValidator("type", "Type of entry to find", false, func(value interface{}) error {
			if value == nil {
				return nil // Optional field
			}
			typeStr, ok := value.(string)
			if !ok {
				return fmt.Errorf("type must be a string")
			}
			validTypes := []string{"f", "d", "l", "file", "dir", "link"}
			for _, vt := range validTypes {
				if typeStr == vt {
					return nil
				}
			}
			return fmt.Errorf("type must be one of: %v", validTypes)
		}).
		AddOptionalStringField("extension", "File extension to search for").
		AddOptionalIntField("max_depth", "Maximum depth to search", 1, 20).
		AddBoolField("case_insensitive", "Case insensitive name matching", false)

	return validator.Validate(args)
}

func (t *FindTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	searchPath := "."
	if pathArg, ok := args["path"]; ok {
		searchPath = pathArg.(string)
	}

	var namePattern string
	if nameArg, ok := args["name"]; ok {
		namePattern, _ = nameArg.(string)
	}

	var fileType string
	if typeArg, ok := args["type"]; ok {
		fileType, _ = typeArg.(string)
	}

	var extension string
	if extArg, ok := args["extension"]; ok {
		extension, _ = extArg.(string)
	}

	maxDepth := -1
	if maxDepthArg, ok := args["max_depth"]; ok {
		if md, ok := maxDepthArg.(float64); ok {
			maxDepth = int(md)
		}
	}

	caseInsensitive := false
	if ciArg, ok := args["case_insensitive"]; ok {
		caseInsensitive, _ = ciArg.(bool)
	}

	var results []map[string]interface{}

	err := filepath.Walk(searchPath, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate depth
		relPath, err := filepath.Rel(searchPath, currentPath)
		if err != nil {
			return err
		}

		depth := 1
		if relPath != "." {
			depth = len(strings.Split(relPath, string(filepath.Separator))) + 1
		}

		// Check max depth
		if maxDepth > 0 && depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file type
		if fileType != "" {
			switch fileType {
			case "f", "file":
				if info.IsDir() {
					return nil
				}
			case "d", "dir":
				if !info.IsDir() {
					return nil
				}
			case "l", "link":
				if info.Mode()&os.ModeSymlink == 0 {
					return nil
				}
			}
		}

		// Check name pattern
		if namePattern != "" {
			fileName := info.Name()
			if caseInsensitive {
				fileName = strings.ToLower(fileName)
				namePattern = strings.ToLower(namePattern)
			}

			if matched, _ := filepath.Match(namePattern, fileName); !matched {
				return nil
			}
		}

		// Check extension
		if extension != "" && !info.IsDir() {
			if filepath.Ext(info.Name()) != extension {
				return nil
			}
		}

		// Add to results
		result := map[string]interface{}{
			"path":     currentPath,
			"name":     info.Name(),
			"type":     getFileTypeString(info),
			"size":     info.Size(),
			"mode":     info.Mode().String(),
			"modified": info.ModTime().Unix(),
			"depth":    depth,
		}

		results = append(results, result)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("find error: %w", err)
	}

	// Generate content
	var contentBuilder strings.Builder
	contentBuilder.WriteString(fmt.Sprintf("Found %d items\n\n", len(results)))

	for _, result := range results {
		path := result["path"].(string)
		fileType := result["type"].(string)
		size := result["size"].(int64)

		if fileType == "directory" {
			contentBuilder.WriteString(fmt.Sprintf("📁 %s/\n", path))
		} else {
			contentBuilder.WriteString(fmt.Sprintf("📄 %s (%s)\n", path, formatFileSize(size)))
		}
	}

	return &ToolResult{
		Content: contentBuilder.String(),
		Data: map[string]interface{}{
			"search_path": searchPath,
			"results":     results,
			"count":       len(results),
			"filters": map[string]interface{}{
				"name":      namePattern,
				"type":      fileType,
				"extension": extension,
				"max_depth": maxDepth,
			},
		},
	}, nil
}

func getFileTypeString(info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "symlink"
	}
	return "file"
}
