# Deep Coding Agent API Reference

## Overview

This document provides a comprehensive reference for the Deep Coding Agent API, including tool interfaces, configuration options, and usage examples.

## Core Interfaces

### UnifiedAgent Interface

```go
type UnifiedAgent interface {
    ProcessTask(ctx context.Context, task *Task) (*AgentResponse, error)
    GetAvailableTools() []string
    GetCapabilities() []string
    Configure(config *ReActConfig) error
    GetStatus() AgentStatus
}
```

### Tool Interface

```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Validate(args map[string]interface{}) error
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}
```

## Built-in Tools

### File Operations

#### file_read
- **Description**: Read file contents with optional line range support
- **Parameters**:
  - `file_path` (string, required): Path to the file
  - `start_line` (integer, optional): Starting line number (1-based)
  - `end_line` (integer, optional): Ending line number (1-based)

#### file_write
- **Description**: Write content to a file
- **Parameters**:
  - `file_path` (string, required): Path to the file
  - `content` (string, required): Content to write
  - `create_dirs` (boolean, optional): Create parent directories

#### file_list
- **Description**: List files and directories
- **Parameters**:
  - `path` (string, optional): Directory path (default: ".")
  - `recursive` (boolean, optional): Recursive listing
  - `depth` (integer, optional): Maximum depth for recursion
  - `show_hidden` (boolean, optional): Include hidden files
  - `file_types` (array, optional): Filter by file extensions

#### file_search
- **Description**: Search for patterns in file contents
- **Parameters**:
  - `pattern` (string, required): Search pattern (regex supported)
  - `path` (string, optional): Search directory
  - `recursive` (boolean, optional): Search recursively
  - `case_sensitive` (boolean, optional): Case sensitive search
  - `max_results` (integer, optional): Maximum results to return
  - `context_lines` (integer, optional): Context lines around matches

### Command Execution

#### bash
- **Description**: Execute shell commands securely
- **Parameters**:
  - `command` (string, required): Command to execute
  - `working_dir` (string, optional): Working directory
  - `timeout` (integer, optional): Timeout in seconds
  - `capture_output` (boolean, optional): Capture command output

#### script_runner
- **Description**: Execute script files with various interpreters
- **Parameters**:
  - `script_path` (string, required): Path to script file
  - `interpreter` (string, optional): Script interpreter (auto-detected)
  - `args` (array, optional): Script arguments
  - `timeout` (integer, optional): Execution timeout
  - `env_vars` (object, optional): Environment variables

#### process_monitor
- **Description**: Monitor and manage system processes
- **Parameters**:
  - `action` (string, required): Action to perform (list/details/kill/monitor/search)
  - `pid` (integer, optional): Process ID (for details/kill actions)
  - `filter` (string, optional): Process name filter
  - `signal` (string, optional): Signal for kill action
  - `duration` (integer, optional): Monitoring duration

## Configuration

### ReAct Agent Configuration

```json
{
  "maxTurns": 10,
  "maxThinkingTime": "30s",
  "maxExecutionTime": "300s",
  "strategy": "standard",
  "parallelExecution": false,
  "enableFallback": true,
  "confidenceThreshold": 0.7,
  "autoRetry": true,
  "maxRetries": 3,
  "loggingLevel": "info"
}
```

### Tool System Configuration

```json
{
  "maxConcurrentExecutions": 5,
  "defaultTimeout": 30000,
  "cacheConfig": {
    "enabled": true,
    "maxSize": 104857600,
    "defaultTTL": 3600
  },
  "securityConfig": {
    "enableSandbox": false,
    "maxMemoryUsage": 536870912,
    "maxExecutionTime": 60000,
    "allowedTools": ["file_read", "file_list", "file_write", "bash"],
    "restrictedTools": ["rm", "format", "dd"]
  }
}
```

### Security Configuration

```json
{
  "policies": [
    {
      "id": "system_protection",
      "name": "System Protection",
      "enabled": true,
      "rules": [
        {
          "type": "deny",
          "target": "path",
          "pattern": "^/(etc|boot|sys|proc|dev)/.*"
        }
      ]
    }
  ],
  "auditEnabled": true,
  "maxEvents": 1000,
  "eventRetention": "720h"
}
```

## Error Handling

### Error Types

- `ValidationError`: Parameter validation failures
- `SecurityError`: Security policy violations
- `ExecutionError`: Tool execution failures
- `TimeoutError`: Operation timeouts
- `ConfigurationError`: Configuration issues

### Error Response Format

```json
{
  "error": {
    "type": "ValidationError",
    "message": "file_path is required",
    "code": "MISSING_PARAMETER",
    "details": {
      "parameter": "file_path",
      "tool": "file_read"
    }
  }
}
```

## Usage Examples

### Basic File Operations

```go
// Read a file
result, err := agent.ExecuteTool(ctx, "file_read", map[string]interface{}{
    "file_path": "/path/to/file.go",
    "start_line": 1,
    "end_line": 50,
})

// Search in files
result, err := agent.ExecuteTool(ctx, "file_search", map[string]interface{}{
    "pattern": "func .*\\(",
    "path": "./src",
    "recursive": true,
    "file_types": []string{".go"},
})
```

### Process Management

```go
// List processes
result, err := agent.ExecuteTool(ctx, "process_monitor", map[string]interface{}{
    "action": "list",
    "filter": "python",
})

// Monitor processes
result, err := agent.ExecuteTool(ctx, "process_monitor", map[string]interface{}{
    "action": "monitor",
    "duration": 30,
    "filter": "node",
})
```

### Command Execution

```go
// Execute shell command
result, err := agent.ExecuteTool(ctx, "bash", map[string]interface{}{
    "command": "git status",
    "working_dir": "/path/to/repo",
    "timeout": 30,
})

// Run script
result, err := agent.ExecuteTool(ctx, "script_runner", map[string]interface{}{
    "script_path": "./scripts/deploy.sh",
    "args": []string{"production"},
    "timeout": 300,
})
```

## Security Considerations

### Tool Validation

All tool executions are subject to security validation:

1. **Parameter Validation**: Input parameters are validated against schemas
2. **Security Policies**: Tool execution is checked against configured policies
3. **Path Restrictions**: Access to sensitive system paths is restricted
4. **Command Filtering**: Dangerous commands are blocked or restricted
5. **Resource Limits**: Memory and execution time limits are enforced

### Best Practices

1. Always validate user inputs before passing to tools
2. Use least-privilege principle for tool permissions
3. Implement proper error handling for security violations
4. Monitor and log tool executions for auditing
5. Regularly review and update security policies
6. Use timeouts to prevent resource exhaustion
7. Sanitize file paths and command arguments

## Performance Considerations

### Optimization Tips

1. **Caching**: Enable tool result caching for repeated operations
2. **Batching**: Group related file operations when possible
3. **Concurrency**: Use parallel execution for independent operations
4. **Resource Management**: Monitor memory and CPU usage
5. **Timeouts**: Set appropriate timeouts for long-running operations

### Monitoring

The agent provides built-in monitoring capabilities:

- Tool execution metrics
- Performance benchmarks
- Resource usage tracking
- Error rate monitoring
- Security event logging

## MCP Integration

### MCP Tool Discovery

The agent supports Model Context Protocol (MCP) for external tool integration:

```json
{
  "mcpConfig": {
    "enabled": true,
    "connectionTimeout": 30,
    "maxConnections": 10,
    "autoDiscovery": true,
    "servers": [
      {
        "name": "external-tools",
        "command": "npx @modelcontextprotocol/server-tools",
        "args": ["--port", "3000"]
      }
    ]
  }
}
```

### Custom Tool Registration

External tools can be registered through MCP:

1. Implement MCP server protocol
2. Define tool schemas and capabilities
3. Handle tool execution requests
4. Provide proper error handling and validation

---

*For more detailed information, see the implementation files in `internal/tools/` and `pkg/types/`.*