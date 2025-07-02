package builtin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"deep-coding-agent/internal/tools"
)

// BashTool implements shell command execution functionality
type BashTool struct{}

func CreateBashTool() *BashTool {
	return &BashTool{}
}

func (t *BashTool) Name() string {
	return "bash"
}

func (t *BashTool) Description() string {
	return "Execute shell commands in the system. Use with caution as this can modify the system."
}

func (t *BashTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to execute",
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Working directory for the command (optional)",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default: 30)",
				"default":     30,
				"minimum":     1,
				"maximum":     300,
			},
			"capture_output": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to capture command output (default: true)",
				"default":     true,
			},
			"allow_interactive": map[string]interface{}{
				"type":        "boolean",
				"description": "Allow interactive commands (default: false)",
				"default":     false,
			},
		},
		"required": []string{"command"},
	}
}

func (t *BashTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("command", "The shell command to execute").
		AddOptionalStringField("working_dir", "Working directory for the command").
		AddOptionalIntField("timeout", "Timeout in seconds", 1, 300).
		AddBoolField("capture_output", "Whether to capture command output", false).
		AddBoolField("allow_interactive", "Allow interactive commands", false)

	// First run standard validation
	if err := validator.Validate(args); err != nil {
		return err
	}

	// Get validated command
	command := args["command"].(string)

	// Enhanced security validation
	if err := t.validateSecurity(command); err != nil {
		return err
	}

	// Validate working directory if provided
	if workingDir, ok := args["working_dir"]; ok && workingDir != nil {
		if workingDirStr, ok := workingDir.(string); ok && workingDirStr != "" {
			if _, err := os.Stat(workingDirStr); os.IsNotExist(err) {
				return fmt.Errorf("working directory does not exist: %s", workingDirStr)
			}
		}
	}

	return nil
}

func (t *BashTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	command := args["command"].(string)

	// Get optional parameters
	workingDir := ""
	if wd, ok := args["working_dir"]; ok {
		workingDir, _ = wd.(string)
	}

	timeout := 30
	if timeoutArg, ok := args["timeout"]; ok {
		if timeoutFloat, ok := timeoutArg.(float64); ok {
			timeout = int(timeoutFloat)
		}
	}

	captureOutput := true
	if captureArg, ok := args["capture_output"]; ok {
		captureOutput, _ = captureArg.(bool)
	}

	allowInteractive := false
	if interactiveArg, ok := args["allow_interactive"]; ok {
		allowInteractive, _ = interactiveArg.(bool)
	}

	// Create command context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Determine shell command based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(cmdCtx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(cmdCtx, "sh", "-c", command)
	}

	// Set working directory if specified
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Prepare for output capture
	var stdout, stderr strings.Builder
	var exitCode int

	startTime := time.Now()

	if captureOutput {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	} else if !allowInteractive {
		// If not capturing output and not interactive, discard output
		cmd.Stdout = nil
		cmd.Stderr = nil
	} else {
		// Interactive mode - connect to terminal
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// Execute command
	err := cmd.Run()
	duration := time.Since(startTime)

	// Get exit code
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	// Prepare result
	var resultContent string
	success := err == nil

	if captureOutput {
		stdoutStr := stdout.String()
		stderrStr := stderr.String()

		if stdoutStr != "" && stderrStr != "" {
			resultContent = fmt.Sprintf("STDOUT:\n%s\n\nSTDERR:\n%s", stdoutStr, stderrStr)
		} else if stdoutStr != "" {
			resultContent = stdoutStr
		} else if stderrStr != "" {
			resultContent = stderrStr
		} else {
			resultContent = "Command executed successfully (no output)"
		}
	} else {
		if success {
			resultContent = "Command executed successfully"
		} else {
			resultContent = fmt.Sprintf("Command failed: %v", err)
		}
	}

	// Handle context cancellation (timeout)
	if cmdCtx.Err() == context.DeadlineExceeded {
		resultContent += fmt.Sprintf("\n\nCommand timed out after %d seconds", timeout)
		success = false
	}

	return &ToolResult{
		Content: resultContent,
		Data: map[string]interface{}{
			"command":     command,
			"exit_code":   exitCode,
			"success":     success,
			"duration_ms": duration.Milliseconds(),
			"working_dir": workingDir,
			"stdout":      stdout.String(),
			"stderr":      stderr.String(),
		},
	}, nil
}

// validateSecurity performs comprehensive security validation on commands
func (t *BashTool) validateSecurity(command string) error {
	lowerCommand := strings.ToLower(strings.TrimSpace(command))

	// Dangerous commands that could harm the system
	dangerousCommands := []string{
		"rm -rf /", "rm -rf .", "rm -rf *", "rm -rf ~",
		"dd if=", "mkfs", "fdisk", "format", "diskpart",
		"del /s", "rmdir /s", "rd /s",
		"shutdown", "reboot", "halt", "poweroff", "init 0", "init 6",
		"killall", "pkill -9", "kill -9",
		"chmod 777 /", "chown root /",
		"mv / ", "cp -r / ",
		"cat /dev/urandom", "> /dev/sda", "> /dev/null",
		":(){ :|:& };:", // fork bomb
	}

	for _, dangerous := range dangerousCommands {
		if strings.Contains(lowerCommand, dangerous) {
			return fmt.Errorf("dangerous command detected: %s", dangerous)
		}
	}

	// Suspicious patterns that warrant extra scrutiny
	suspiciousPatterns := []string{
		"/etc/passwd", "/etc/shadow", "/etc/sudoers",
		"sudo su", "sudo -i", "sudo bash", "sudo sh",
		"nc -l", "netcat -l", "socat",
		"wget http", "curl http", "curl ftp",
		"python -c", "python3 -c", "perl -e",
		"base64 -d", "echo | sh", "eval ",
		"$(", "`", // command substitution
		"&& rm", "|| rm", "; rm",
		"chmod +x", "chmod 755", "chmod 777",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerCommand, pattern) {
			return fmt.Errorf("potentially dangerous pattern detected: %s", pattern)
		}
	}

	// Check for attempts to access sensitive directories
	restrictedPaths := []string{
		"/etc/", "/root/", "/boot/", "/sys/", "/proc/",
		"/var/log/", "/usr/bin/", "/usr/sbin/", "/sbin/",
		"c:\\windows\\", "c:\\program files\\", "c:\\system32\\",
	}

	for _, path := range restrictedPaths {
		if strings.Contains(lowerCommand, path) {
			return fmt.Errorf("access to restricted path detected: %s", path)
		}
	}

	// Check for networking commands that could be used maliciously
	networkCommands := []string{
		"ssh", "scp", "rsync", "ftp", "sftp",
		"telnet", "nmap", "ping -f", "ping -c 1000",
		"iptables", "ufw", "firewall-cmd",
	}

	for _, netCmd := range networkCommands {
		if strings.Contains(lowerCommand, netCmd) {
			return fmt.Errorf("network command requires explicit permission: %s", netCmd)
		}
	}

	// Check command length to prevent buffer overflow attempts
	if len(command) > 1000 {
		return fmt.Errorf("command too long (max 1000 characters)")
	}

	return nil
}

// ScriptRunnerTool implements script execution functionality
type ScriptRunnerTool struct{}

func CreateScriptRunnerTool() *ScriptRunnerTool {
	return &ScriptRunnerTool{}
}

func (t *ScriptRunnerTool) Name() string {
	return "script_runner"
}

func (t *ScriptRunnerTool) Description() string {
	return "Execute script files with various interpreters. Supports shell scripts, Python, Node.js, and more."
}

func (t *ScriptRunnerTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"script_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the script file to execute",
			},
			"interpreter": map[string]interface{}{
				"type":        "string",
				"description": "Script interpreter (auto-detected if not specified)",
				"enum":        []string{"bash", "sh", "python", "python3", "node", "ruby", "perl", "php", "auto"},
				"default":     "auto",
			},
			"args": map[string]interface{}{
				"type":        "array",
				"description": "Arguments to pass to the script",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Working directory for script execution",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default: 60)",
				"default":     60,
				"minimum":     1,
				"maximum":     600,
			},
		},
		"required": []string{"script_path"},
	}
}

func (t *ScriptRunnerTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddStringField("script_path", "Path to the script file to execute").
		AddCustomValidator("interpreter", "Script interpreter (auto-detected if not specified)", false, func(value interface{}) error {
			if value == nil {
				return nil // Optional field
			}
			interpreter, ok := value.(string)
			if !ok {
				return fmt.Errorf("interpreter must be a string")
			}
			validInterpreters := []string{"bash", "sh", "python", "python3", "node", "ruby", "perl", "php", "auto"}
			for _, vi := range validInterpreters {
				if interpreter == vi {
					return nil
				}
			}
			return fmt.Errorf("invalid interpreter: %s", interpreter)
		}).
		AddOptionalStringField("working_dir", "Working directory for script execution").
		AddOptionalIntField("timeout", "Timeout in seconds", 1, 600)

	// First run standard validation
	if err := validator.Validate(args); err != nil {
		return err
	}

	// Additional validation: check if script file exists
	scriptPath := args["script_path"].(string)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script file does not exist: %s", scriptPath)
	}

	return nil
}

func (t *ScriptRunnerTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	scriptPath := args["script_path"].(string)

	// Get interpreter
	interpreter := "auto"
	if interpreterArg, ok := args["interpreter"]; ok {
		interpreter, _ = interpreterArg.(string)
	}

	// Auto-detect interpreter if needed
	if interpreter == "auto" {
		interpreter = t.detectInterpreter(scriptPath)
	}

	// Get timeout
	timeout := 60
	if timeoutArg, ok := args["timeout"]; ok {
		if timeoutFloat, ok := timeoutArg.(float64); ok {
			timeout = int(timeoutFloat)
		}
	}

	// Get working directory
	workingDir := os.TempDir()
	if wd, ok := args["working_dir"]; ok {
		if wdStr, ok := wd.(string); ok && wdStr != "" {
			workingDir = wdStr
		}
	}

	// Get script arguments
	var scriptArgs []string
	if argsArg, ok := args["args"]; ok {
		if argsSlice, ok := argsArg.([]interface{}); ok {
			for _, arg := range argsSlice {
				if argStr, ok := arg.(string); ok {
					scriptArgs = append(scriptArgs, argStr)
				}
			}
		}
	}

	// Create command
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	switch interpreter {
	case "bash":
		cmdArgs := append([]string{scriptPath}, scriptArgs...)
		cmd = exec.CommandContext(cmdCtx, "bash", cmdArgs...)
	case "sh":
		cmdArgs := append([]string{scriptPath}, scriptArgs...)
		cmd = exec.CommandContext(cmdCtx, "sh", cmdArgs...)
	case "python", "python3":
		cmdArgs := append([]string{scriptPath}, scriptArgs...)
		cmd = exec.CommandContext(cmdCtx, interpreter, cmdArgs...)
	case "node":
		cmdArgs := append([]string{scriptPath}, scriptArgs...)
		cmd = exec.CommandContext(cmdCtx, "node", cmdArgs...)
	case "ruby":
		cmdArgs := append([]string{scriptPath}, scriptArgs...)
		cmd = exec.CommandContext(cmdCtx, "ruby", cmdArgs...)
	case "perl":
		cmdArgs := append([]string{scriptPath}, scriptArgs...)
		cmd = exec.CommandContext(cmdCtx, "perl", cmdArgs...)
	case "php":
		cmdArgs := append([]string{scriptPath}, scriptArgs...)
		cmd = exec.CommandContext(cmdCtx, "php", cmdArgs...)
	default:
		return nil, fmt.Errorf("unsupported interpreter: %s", interpreter)
	}

	cmd.Dir = workingDir

	// Execute script
	startTime := time.Now()
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(startTime)

	// Get exit code
	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	// Prepare result content
	var resultContent string
	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	if stdoutStr != "" && stderrStr != "" {
		resultContent = fmt.Sprintf("STDOUT:\n%s\n\nSTDERR:\n%s", stdoutStr, stderrStr)
	} else if stdoutStr != "" {
		resultContent = stdoutStr
	} else if stderrStr != "" {
		resultContent = stderrStr
	} else {
		resultContent = "Script executed successfully (no output)"
	}

	// Handle timeout
	if cmdCtx.Err() == context.DeadlineExceeded {
		resultContent += fmt.Sprintf("\n\nScript timed out after %d seconds", timeout)
	}

	return &ToolResult{
		Content: resultContent,
		Data: map[string]interface{}{
			"script_path": scriptPath,
			"interpreter": interpreter,
			"exit_code":   exitCode,
			"success":     err == nil,
			"duration_ms": duration.Milliseconds(),
			"working_dir": workingDir,
			"stdout":      stdoutStr,
			"stderr":      stderrStr,
			"args":        scriptArgs,
		},
	}, nil
}

// detectInterpreter attempts to detect the appropriate interpreter for a script
func (t *ScriptRunnerTool) detectInterpreter(scriptPath string) string {
	// First, try file extension
	ext := strings.ToLower(filepath.Ext(scriptPath))
	switch ext {
	case ".sh", ".bash":
		return "bash"
	case ".py":
		return "python3"
	case ".js":
		return "node"
	case ".rb":
		return "ruby"
	case ".pl":
		return "perl"
	case ".php":
		return "php"
	}

	return "sh" // default fallback
}

// ProcessMonitorTool implements process monitoring functionality
type ProcessMonitorTool struct{}

func CreateProcessMonitorTool() *ProcessMonitorTool {
	return &ProcessMonitorTool{}
}

func (t *ProcessMonitorTool) Name() string {
	return "process_monitor"
}

func (t *ProcessMonitorTool) Description() string {
	return "Monitor and manage system processes. List processes, get details, and control process lifecycle."
}

func (t *ProcessMonitorTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action to perform",
				"enum":        []string{"list", "search"},
			},
			"filter": map[string]interface{}{
				"type":        "string",
				"description": "Filter processes by name pattern (for list and search actions)",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of processes to return",
				"default":     50,
				"minimum":     1,
				"maximum":     500,
			},
		},
		"required": []string{"action"},
	}
}

func (t *ProcessMonitorTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddCustomValidator("action", "Action to perform (list, search)", true, func(value interface{}) error {
			action, ok := value.(string)
			if !ok {
				return fmt.Errorf("action must be a string")
			}
			validActions := []string{"list", "search"}
			for _, va := range validActions {
				if action == va {
					return nil
				}
			}
			return fmt.Errorf("invalid action: %s", action)
		}).
		AddOptionalStringField("filter", "Filter processes by name pattern").
		AddOptionalIntField("max_results", "Maximum number of processes to return", 1, 500)

	return validator.Validate(args)
}

func (t *ProcessMonitorTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	action := args["action"].(string)

	switch action {
	case "list":
		return t.listProcesses(args)
	case "search":
		filter := ""
		if f, ok := args["filter"]; ok {
			filter, _ = f.(string)
		}
		return t.searchProcesses(filter, args)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (t *ProcessMonitorTool) listProcesses(args map[string]interface{}) (*ToolResult, error) {
	filter := ""
	if f, ok := args["filter"]; ok {
		filter, _ = f.(string)
	}

	maxResults := 50
	if mr, ok := args["max_results"]; ok {
		if mrFloat, ok := mr.(float64); ok {
			maxResults = int(mrFloat)
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist", "/fo", "csv")
	} else {
		cmd = exec.Command("ps", "aux")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	processes := t.parseProcessList(string(output), filter, maxResults)

	// Create summary
	summary := fmt.Sprintf("Found %d processes", len(processes))
	if filter != "" {
		summary += fmt.Sprintf(" matching filter '%s'", filter)
	}

	// Create detailed output
	var contentBuilder strings.Builder
	contentBuilder.WriteString(summary + "\n\n")

	if runtime.GOOS == "windows" {
		contentBuilder.WriteString("PID\tName\tMemory\n")
		contentBuilder.WriteString("---\t----\t------\n")
	} else {
		contentBuilder.WriteString("PID\tUser\tCPU\tMem\tCommand\n")
		contentBuilder.WriteString("---\t----\t---\t---\t-------\n")
	}

	for _, proc := range processes {
		if runtime.GOOS == "windows" {
			contentBuilder.WriteString(fmt.Sprintf("%d\t%s\t%s\n",
				proc["pid"], proc["name"], proc["memory"]))
		} else {
			contentBuilder.WriteString(fmt.Sprintf("%d\t%s\t%s\t%s\t%s\n",
				proc["pid"], proc["user"], proc["cpu"], proc["memory"], proc["command"]))
		}
	}

	return &ToolResult{
		Content: contentBuilder.String(),
		Data: map[string]interface{}{
			"processes":     processes,
			"process_count": len(processes),
			"filter":        filter,
			"os":            runtime.GOOS,
		},
	}, nil
}

func (t *ProcessMonitorTool) parseProcessList(output, filter string, maxResults int) []map[string]interface{} {
	var processes []map[string]interface{}
	lines := strings.Split(output, "\n")
	count := 0

	if runtime.GOOS == "windows" {
		// Parse Windows tasklist CSV output
		for i, line := range lines {
			if i == 0 || strings.TrimSpace(line) == "" || count >= maxResults {
				continue // skip header and empty lines
			}

			// Simple CSV parsing for tasklist output
			fields := strings.Split(line, ",")
			if len(fields) >= 5 {
				name := strings.Trim(fields[0], "\"")
				pidStr := strings.Trim(fields[1], "\"")
				memory := strings.Trim(fields[4], "\"")

				if filter != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(filter)) {
					continue
				}

				if pid := parsePID(pidStr); pid > 0 {
					processes = append(processes, map[string]interface{}{
						"pid":    pid,
						"name":   name,
						"memory": memory,
					})
					count++
				}
			}
		}
	} else {
		// Parse Unix ps aux output
		for i, line := range lines {
			if i == 0 || strings.TrimSpace(line) == "" || count >= maxResults {
				continue // skip header and empty lines
			}

			fields := strings.Fields(line)
			if len(fields) >= 11 {
				user := fields[0]
				pidStr := fields[1]
				cpu := fields[2]
				memory := fields[3]
				command := strings.Join(fields[10:], " ")

				if filter != "" && !strings.Contains(strings.ToLower(command), strings.ToLower(filter)) {
					continue
				}

				if pid := parsePID(pidStr); pid > 0 {
					processes = append(processes, map[string]interface{}{
						"pid":     pid,
						"user":    user,
						"cpu":     cpu,
						"memory":  memory,
						"command": command,
					})
					count++
				}
			}
		}
	}

	return processes
}

func (t *ProcessMonitorTool) searchProcesses(filter string, args map[string]interface{}) (*ToolResult, error) {
	if filter == "" {
		return nil, fmt.Errorf("filter is required for search action")
	}

	return t.listProcesses(args)
}

// Helper function to parse PID strings
func parsePID(pidStr string) int {
	if pid, err := strconv.Atoi(pidStr); err == nil {
		return pid
	}
	return 0
}

// CodeExecutorTool implements the CodeActExecutor as a tool
type CodeExecutorTool struct {
	executor *tools.CodeActExecutor
}

func CreateCodeExecutorTool() *CodeExecutorTool {
	return &CodeExecutorTool{
		executor: tools.NewCodeActExecutor(),
	}
}

func (t *CodeExecutorTool) Name() string {
	return "code_execute"
}

func (t *CodeExecutorTool) Description() string {
	return "Execute code in supported languages (Python, Go, JavaScript, Bash) in a sandboxed environment."
}

func (t *CodeExecutorTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"language": map[string]interface{}{
				"type":        "string",
				"description": "Programming language to execute",
				"enum":        []string{"python", "go", "javascript", "js", "bash"},
			},
			"code": map[string]interface{}{
				"type":        "string",
				"description": "Source code to execute",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default: 30)",
				"default":     30,
				"minimum":     1,
				"maximum":     300,
			},
		},
		"required": []string{"language", "code"},
	}
}

func (t *CodeExecutorTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddCustomValidator("language", "Programming language to execute", true, func(value interface{}) error {
			language, ok := value.(string)
			if !ok {
				return fmt.Errorf("language must be a string")
			}
			supportedLangs := []string{"python", "go", "javascript", "js", "bash"}
			for _, lang := range supportedLangs {
				if language == lang {
					return nil
				}
			}
			return fmt.Errorf("unsupported language: %s", language)
		}).
		AddStringField("code", "Source code to execute").
		AddOptionalIntField("timeout", "Timeout in seconds", 1, 300)

	return validator.Validate(args)
}

func (t *CodeExecutorTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	language := args["language"].(string)
	code := args["code"].(string)

	// Set timeout if provided
	if timeoutArg, ok := args["timeout"]; ok {
		if timeoutFloat, ok := timeoutArg.(float64); ok {
			timeout := time.Duration(timeoutFloat) * time.Second
			t.executor.SetTimeout(timeout)
		}
	}

	// Execute the code
	result, err := t.executor.ExecuteCode(ctx, language, code)
	if err != nil {
		return nil, fmt.Errorf("failed to execute code: %w", err)
	}

	// Prepare content
	var content string
	if result.Success {
		if result.Output != "" {
			content = fmt.Sprintf("Code executed successfully in %v:\n\n%s", result.ExecutionTime, result.Output)
		} else {
			content = fmt.Sprintf("Code executed successfully in %v (no output)", result.ExecutionTime)
		}
	} else {
		content = fmt.Sprintf("Code execution failed:\n\n%s", result.Error)
	}

	return &ToolResult{
		Content: content,
		Data: map[string]interface{}{
			"success":        result.Success,
			"output":         result.Output,
			"error":          result.Error,
			"exit_code":      result.ExitCode,
			"execution_time": result.ExecutionTime.Milliseconds(),
			"language":       result.Language,
			"code":           result.Code,
		},
	}, nil
}
