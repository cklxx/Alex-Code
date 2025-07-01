# CodeAct API Reference

## Overview

This document provides a comprehensive API reference for the CodeAct integration in the Deep Coding Agent. CodeAct extends the ReAct framework by using executable Python code as the primary action language.

## Table of Contents

1. [Core Types](#core-types)
2. [Tool API](#tool-api)
3. [Planning API](#planning-api)
4. [Execution API](#execution-api)
5. [Security API](#security-api)
6. [Configuration API](#configuration-api)
7. [Monitoring API](#monitoring-api)

## Core Types

### CodeActPlan

Represents a plan for executing Python code within the CodeAct framework.

```go
type CodeActPlan struct {
    ID               string                 `json:"id"`
    Code             string                 `json:"code"`
    Language         string                 `json:"language"`
    ExecutionMode    CodeExecutionMode      `json:"execution_mode"`
    Dependencies     []string               `json:"dependencies"`
    ExpectedOutput   string                 `json:"expected_output"`
    SafetyChecks     []SafetyValidation     `json:"safety_checks"`
    Timeout          time.Duration          `json:"timeout"`
    MemoryLimit      int64                 `json:"memory_limit"`
    AllowedModules   []string              `json:"allowed_modules"`
    RestrictedOps    []string              `json:"restricted_ops"`
}
```

**Fields:**
- `ID`: Unique identifier for the plan
- `Code`: Python code to execute
- `Language`: Programming language (currently "python")
- `ExecutionMode`: How the code should be executed
- `Dependencies`: Required modules or dependencies
- `ExpectedOutput`: Description of expected execution result
- `SafetyChecks`: Security validation rules
- `Timeout`: Maximum execution time
- `MemoryLimit`: Maximum memory usage in bytes
- `AllowedModules`: List of permitted Python modules
- `RestrictedOps`: List of forbidden operations

### CodeExecutionMode

Defines how code should be executed.

```go
type CodeExecutionMode string

const (
    CodeExecutionModeInteractive CodeExecutionMode = "interactive"
    CodeExecutionModeBatch       CodeExecutionMode = "batch"
    CodeExecutionModeSandbox     CodeExecutionMode = "sandbox"
)
```

**Values:**
- `interactive`: Interactive Python session with persistent state
- `batch`: One-time script execution
- `sandbox`: Isolated execution environment (recommended)

### CodeExecutionResult

Contains the results of code execution.

```go
type CodeExecutionResult struct {
    ID            string                 `json:"id"`
    Code          string                 `json:"code"`
    Output        string                 `json:"output"`
    Errors        []CodeError            `json:"errors"`
    Success       bool                   `json:"success"`
    ExecutionTime time.Duration          `json:"execution_time"`
    MemoryUsage   int64                 `json:"memory_usage"`
    Variables     map[string]interface{} `json:"variables"`
    FilesCreated  []string              `json:"files_created"`
    FilesModified []string              `json:"files_modified"`
    ReturnValue   interface{}           `json:"return_value"`
    Metadata      map[string]interface{} `json:"metadata"`
}
```

### CodeError

Represents an error that occurred during code execution.

```go
type CodeError struct {
    Line    int    `json:"line"`
    Column  int    `json:"column"`
    Type    string `json:"type"`
    Message string `json:"message"`
    Code    string `json:"code"`
}
```

## Tool API

### PythonInterpreterTool

The core tool for executing Python code in a secure environment.

#### Methods

##### Execute

Executes Python code with specified parameters.

```go
func (pit *PythonInterpreterTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
```

**Parameters:**
```json
{
  "code": "string (required) - Python code to execute",
  "mode": "string (optional) - execution mode: interactive|batch|sandbox",
  "timeout": "integer (optional) - maximum execution time in seconds",
  "preserve_session": "boolean (optional) - preserve variables between executions"
}
```

**Example Usage:**
```go
args := map[string]interface{}{
    "code": `
import pandas as pd
df = pd.read_csv('data.csv')
print(f"Shape: {df.shape}")
result = df.describe()
print(result)
`,
    "mode": "sandbox",
    "timeout": 30,
    "preserve_session": true,
}

result, err := tool.Execute(ctx, args)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Output: %s\n", result.Content)
```

**Response Format:**
```json
{
  "content": "execution output",
  "data": {
    "success": true,
    "output": "Shape: (100, 5)\n...",
    "execution_time": 1250,
    "memory_usage": 45678901,
    "files_created": [],
    "files_modified": ["output.csv"],
    "return_value": null
  }
}
```

##### Name

Returns the tool name.

```go
func (pit *PythonInterpreterTool) Name() string
```

**Returns:** `"python_interpreter"`

##### Description

Returns a human-readable description.

```go
func (pit *PythonInterpreterTool) Description() string
```

##### Parameters

Returns the JSON schema for tool parameters.

```go
func (pit *PythonInterpreterTool) Parameters() map[string]interface{}
```

##### Validate

Validates input parameters before execution.

```go
func (pit *PythonInterpreterTool) Validate(args map[string]interface{}) error
```

## Planning API

### CodeActPlanner

Extends the action planner to generate executable Python code.

#### Constructor

```go
func NewCodeActPlanner(aiProvider ai.Provider, toolSystem interfaces.ToolSystem, config *types.ReActConfig) *CodeActPlanner
```

#### Methods

##### PlanCodeAct

Generates a CodeAct execution plan based on reasoning.

```go
func (cap *CodeActPlanner) PlanCodeAct(ctx context.Context, thought *types.ThoughtProcess, context *types.Context) (*types.CodeActPlan, error)
```

**Parameters:**
- `ctx`: Execution context
- `thought`: Current reasoning state
- `context`: Available context information

**Returns:** CodeActPlan with generated code and execution parameters

**Example:**
```go
planner := NewCodeActPlanner(aiProvider, toolSystem, config)
plan, err := planner.PlanCodeAct(ctx, thought, context)
if err != nil {
    return err
}

fmt.Printf("Generated code:\n%s\n", plan.Code)
fmt.Printf("Expected output: %s\n", plan.ExpectedOutput)
```

## Execution API

### ExecutionSandbox

Provides secure code execution in isolated environments.

#### Methods

##### ExecutePython

Executes Python code in a sandboxed environment.

```go
func (es *ExecutionSandbox) ExecutePython(ctx context.Context, code string, timeout time.Duration) (*types.CodeExecutionResult, error)
```

**Parameters:**
- `ctx`: Execution context with cancellation support
- `code`: Python code to execute
- `timeout`: Maximum execution time

**Example:**
```go
sandbox := NewExecutionSandbox()
result, err := sandbox.ExecutePython(ctx, `
import math
result = math.sqrt(16)
print(f"Square root of 16 is {result}")
`, 30*time.Second)

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Success: %v\n", result.Success)
fmt.Printf("Output: %s\n", result.Output)
fmt.Printf("Execution time: %v\n", result.ExecutionTime)
```

##### CreateEnvironment

Creates a new isolated execution environment.

```go
func (es *ExecutionSandbox) CreateEnvironment(config *SandboxConfig) (*Environment, error)
```

##### DestroyEnvironment

Cleans up an execution environment.

```go
func (es *ExecutionSandbox) DestroyEnvironment(envID string) error
```

## Security API

### CodeSecurityValidator

Validates code for security threats before execution.

#### Constructor

```go
func NewCodeSecurityValidator() *CodeSecurityValidator
```

#### Methods

##### ValidateCode

Performs comprehensive security validation on code.

```go
func (csv *CodeSecurityValidator) ValidateCode(code string) error
```

**Security Checks:**
- Forbidden pattern detection
- Dangerous import validation
- System call prevention
- File system access restriction
- Network operation blocking

**Example:**
```go
validator := NewCodeSecurityValidator()

// Safe code
err := validator.ValidateCode(`
import math
result = math.sqrt(25)
print(result)
`)
// err == nil

// Dangerous code
err = validator.ValidateCode(`
import subprocess
subprocess.call(['rm', '-rf', '/'])
`)
// err != nil (forbidden operation detected)
```

##### AssessRisk

Evaluates the risk level of code execution.

```go
func (csv *CodeSecurityValidator) AssessRisk(code string) (*SecurityRisk, error)
```

**Returns:**
```go
type SecurityRisk struct {
    Level       RiskLevel  // Low, Medium, High, Critical
    Patterns    []string   // Detected risk patterns
    Suggestions []string   // Security improvement suggestions
    Score       float64    // Risk score (0.0-1.0)
}
```

##### UpdatePatterns

Updates security patterns and rules.

```go
func (csv *CodeSecurityValidator) UpdatePatterns(patterns []SecurityPattern) error
```

## Configuration API

### PythonConfig

Configuration for Python code execution.

```go
type PythonConfig struct {
    PythonPath         string        `json:"python_path"`
    WorkingDirectory   string        `json:"working_directory"`
    MaxExecutionTime   time.Duration `json:"max_execution_time"`
    MaxMemoryUsage     int64        `json:"max_memory_usage"`
    AllowedModules     []string     `json:"allowed_modules"`
    ForbiddenPatterns  []string     `json:"forbidden_patterns"`
    SandboxEnabled     bool         `json:"sandbox_enabled"`
}
```

**Default Configuration:**
```go
defaultConfig := &PythonConfig{
    PythonPath:         "python3",
    WorkingDirectory:   "/tmp/codeact_workspace",
    MaxExecutionTime:   30 * time.Second,
    MaxMemoryUsage:     512 * 1024 * 1024, // 512MB
    AllowedModules:     []string{"os", "sys", "json", "math", "datetime", "requests", "pandas", "numpy"},
    ForbiddenPatterns:  []string{"eval(", "exec(", "subprocess", "socket"},
    SandboxEnabled:     true,
}
```

### SandboxConfig

Configuration for execution sandbox.

```go
type SandboxConfig struct {
    ContainerImage    string            `json:"container_image"`
    ResourceLimits    *ResourceLimits   `json:"resource_limits"`
    NetworkEnabled    bool              `json:"network_enabled"`
    MountPaths        []MountPath       `json:"mount_paths"`
    EnvironmentVars   map[string]string `json:"environment_vars"`
    TimeoutSeconds    int               `json:"timeout_seconds"`
}
```

## Monitoring API

### Metrics Collection

#### ExecutionMetrics

Tracks code execution performance and success rates.

```go
type ExecutionMetrics struct {
    TotalExecutions    int           `json:"total_executions"`
    SuccessfulExecutions int         `json:"successful_executions"`
    FailedExecutions   int           `json:"failed_executions"`
    AverageExecutionTime time.Duration `json:"average_execution_time"`
    AverageMemoryUsage int64         `json:"average_memory_usage"`
    ErrorsByType       map[string]int `json:"errors_by_type"`
    LastUpdated        time.Time     `json:"last_updated"`
}
```

#### GetMetrics

Retrieves current execution metrics.

```go
func (pit *PythonInterpreterTool) GetMetrics() *ExecutionMetrics
```

**Example:**
```go
metrics := tool.GetMetrics()
fmt.Printf("Success rate: %.2f%%\n", 
    float64(metrics.SuccessfulExecutions) / float64(metrics.TotalExecutions) * 100)
fmt.Printf("Average execution time: %v\n", metrics.AverageExecutionTime)
```

### Health Checks

#### IsHealthy

Checks if the CodeAct system is operational.

```go
func (cs *CodeActSystem) IsHealthy(ctx context.Context) (bool, error)
```

#### GetStatus

Returns detailed system status.

```go
func (cs *CodeActSystem) GetStatus() *SystemStatus
```

**Returns:**
```go
type SystemStatus struct {
    PythonInterpreter  ComponentStatus `json:"python_interpreter"`
    SecurityValidator  ComponentStatus `json:"security_validator"`
    ExecutionSandbox   ComponentStatus `json:"execution_sandbox"`
    OverallHealth      HealthStatus    `json:"overall_health"`
    LastCheckTime      time.Time       `json:"last_check_time"`
}
```

## Error Handling

### Error Types

#### CodeExecutionError

Errors that occur during code execution.

```go
type CodeExecutionError struct {
    Type        ExecutionErrorType `json:"type"`
    Message     string            `json:"message"`
    Code        string            `json:"code"`
    Line        int               `json:"line"`
    Column      int               `json:"column"`
    Recoverable bool              `json:"recoverable"`
}
```

#### SecurityValidationError

Errors from security validation.

```go
type SecurityValidationError struct {
    Pattern     string       `json:"pattern"`
    Message     string       `json:"message"`
    Severity    RiskLevel    `json:"severity"`
    Suggestion  string       `json:"suggestion"`
}
```

### Error Recovery

#### AutoCorrection

Automatic error correction for common issues.

```go
func (ac *AutoCorrector) CorrectCode(code string, error *CodeExecutionError) (string, error)
```

**Supported Corrections:**
- Syntax error fixes
- Import statement corrections
- Variable naming conflicts
- Indentation issues

## Code Examples

### Basic Code Execution

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "deep-coding-agent/internal/tools/builtin"
)

func main() {
    // Create Python interpreter tool
    tool := builtin.NewPythonInterpreterTool()
    
    // Execute simple calculation
    args := map[string]interface{}{
        "code": `
result = 2 + 2
print(f"2 + 2 = {result}")
`,
        "mode": "sandbox",
    }
    
    result, err := tool.Execute(context.Background(), args)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Output: %s\n", result.Content)
}
```

### Data Analysis Example

```go
// Analyze CSV data
code := `
import pandas as pd
import numpy as np

# Load data
df = pd.read_csv('sales_data.csv')

# Basic analysis
total_sales = df['amount'].sum()
avg_sales = df['amount'].mean()
top_products = df.groupby('product')['amount'].sum().sort_values(ascending=False).head(5)

print(f"Total sales: ${total_sales:,.2f}")
print(f"Average sale: ${avg_sales:.2f}")
print("\nTop 5 products by sales:")
print(top_products)

# Save summary
summary = {
    'total_sales': float(total_sales),
    'average_sale': float(avg_sales),
    'top_products': top_products.to_dict()
}

import json
with open('sales_summary.json', 'w') as f:
    json.dump(summary, f, indent=2)

print("\nSummary saved to sales_summary.json")
`

args := map[string]interface{}{
    "code": code,
    "mode": "sandbox",
    "timeout": 60,
}
```

### Error Handling Example

```go
// Execute code with error handling
result, err := tool.Execute(ctx, args)
if err != nil {
    log.Printf("Execution failed: %v", err)
    return
}

// Check execution result
if data, ok := result.Data.(map[string]interface{}); ok {
    if success, ok := data["success"].(bool); ok && !success {
        fmt.Printf("Code execution failed\n")
        if errors, ok := data["errors"].([]interface{}); ok {
            for _, errData := range errors {
                if errMap, ok := errData.(map[string]interface{}); ok {
                    fmt.Printf("Error: %s at line %v\n", 
                        errMap["message"], errMap["line"])
                }
            }
        }
        return
    }
}

fmt.Printf("Execution successful: %s\n", result.Content)
```

### Security Validation Example

```go
validator := builtin.NewCodeSecurityValidator()

// Validate code before execution
if err := validator.ValidateCode(code); err != nil {
    fmt.Printf("Security validation failed: %v\n", err)
    return
}

// Assess risk level
risk, err := validator.AssessRisk(code)
if err != nil {
    log.Printf("Risk assessment failed: %v", err)
    return
}

fmt.Printf("Risk level: %s (score: %.2f)\n", risk.Level, risk.Score)
if len(risk.Suggestions) > 0 {
    fmt.Println("Security suggestions:")
    for _, suggestion := range risk.Suggestions {
        fmt.Printf("- %s\n", suggestion)
    }
}
```

## Integration Patterns

### Strategy Selection

```go
// Determine execution strategy based on task complexity
func selectStrategy(task *types.Task, context *types.Context) types.ReActStrategy {
    complexity := analyzeComplexity(task)
    
    switch {
    case complexity == "simple":
        return types.ReActStrategyStandard
    case complexity == "moderate":
        return types.ReActStrategyHybrid
    case complexity == "complex":
        return types.ReActStrategyCodeAct
    default:
        return types.ReActStrategyHybrid
    }
}
```

### Hybrid Execution

```go
// Combine traditional tools with CodeAct
func executeHybridTask(ctx context.Context, task *types.Task) error {
    // Use traditional tools for file operations
    fileList, err := toolSystem.ExecuteTool(ctx, "file_list", map[string]interface{}{
        "path": ".",
    })
    if err != nil {
        return err
    }
    
    // Use CodeAct for data processing
    code := generateDataProcessingCode(fileList)
    result, err := pythonTool.Execute(ctx, map[string]interface{}{
        "code": code,
        "mode": "sandbox",
    })
    if err != nil {
        return err
    }
    
    // Use traditional tools for output
    return saveResults(result)
}
```

This API reference provides comprehensive documentation for integrating and using CodeAct within the Deep Coding Agent framework. For additional examples and advanced usage patterns, refer to the main integration guide and test files.