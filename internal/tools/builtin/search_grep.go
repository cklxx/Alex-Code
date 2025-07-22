package builtin

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GrepTool implements grep functionality
type GrepTool struct{}

func CreateGrepTool() *GrepTool {
	return &GrepTool{}
}

func (t *GrepTool) Name() string {
	return "grep"
}

func (t *GrepTool) Description() string {
	return "Search for patterns in files using grep."
}

func (t *GrepTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "The pattern to search for",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to search in",
				"default":     ".",
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"description": "Search recursively",
				"default":     false,
			},
			"ignore_case": map[string]interface{}{
				"type":        "boolean",
				"description": "Ignore case",
				"default":     false,
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GrepTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("pattern", "Pattern to search for").
		AddOptionalStringField("path", "Path to search in").
		AddOptionalBooleanField("recursive", "Search recursively").
		AddOptionalBooleanField("ignore_case", "Ignore case")

	return validator.Validate(args)
}

func (t *GrepTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	// 防御性检查：确保参数存在且有效（通常已通过Validate验证）
	if args == nil {
		return nil, fmt.Errorf("arguments cannot be nil")
	}
	
	patternValue, exists := args["pattern"]
	if !exists {
		return nil, fmt.Errorf("pattern parameter is required")
	}
	
	pattern, ok := patternValue.(string)
	if !ok {
		return nil, fmt.Errorf("pattern must be a string")
	}
	
	if pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}
	
	path := "."
	if p, ok := args["path"]; ok {
		path = p.(string)
	}
	
	recursive := false
	if r, ok := args["recursive"].(bool); ok {
		recursive = r
	}
	
	ignoreCase := false
	if ic, ok := args["ignore_case"].(bool); ok {
		ignoreCase = ic
	}
	
	// Build grep command
	cmdArgs := []string{}
	
	if ignoreCase {
		cmdArgs = append(cmdArgs, "-i")
	}
	
	cmdArgs = append(cmdArgs, "-n") // Always show line numbers
	
	if recursive {
		cmdArgs = append(cmdArgs, "-r")
	}
	
	cmdArgs = append(cmdArgs, pattern, path)
	
	// Execute grep command
	cmd := exec.CommandContext(ctx, "grep", cmdArgs...)
	output, err := cmd.Output()
	
	if err != nil {
		// grep returns exit code 1 when no matches found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return &ToolResult{
				Content: "No matches found",
				Data: map[string]interface{}{
					"pattern":     pattern,
					"path":        path,
					"matches":     0,
					"recursive":   recursive,
					"ignore_case": ignoreCase,
				},
			}, nil
		}
		return nil, fmt.Errorf("grep command failed: %w", err)
	}
	
	// Process output
	lines := strings.Split(string(output), "\n")
	lines = lines[:len(lines)-1] // Remove last empty line
	
	return &ToolResult{
		Content: fmt.Sprintf("Found %d matches:\n%s", len(lines), strings.Join(lines, "\n")),
		Data: map[string]interface{}{
			"pattern":     pattern,
			"path":        path,
			"matches":     len(lines),
			"recursive":   recursive,
			"ignore_case": ignoreCase,
			"results":     lines,
		},
	}, nil
}