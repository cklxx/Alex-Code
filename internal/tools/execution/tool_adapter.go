package execution

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	
	registryPkg "deep-coding-agent/internal/tools/registry"
	"deep-coding-agent/pkg/types"
)

// ToolSystemAdapter adapts tools.Registry to interfaces.ToolOrchestrator
type ToolSystemAdapter struct {
	registry *registryPkg.Registry
	
	// Enhanced metrics and tracking
	metrics *types.ToolSystemMetrics
	execHistory []types.ToolExecutionRecord
	mutex sync.RWMutex
	
	// Configuration
	config *types.ToolSystemConfig
	
	// Tool recommendation engine
	recommender *ToolRecommendationEngine
}

// ToolRecommendationEngine provides intelligent tool recommendations
type ToolRecommendationEngine struct {
	toolUsagePatterns map[string]*types.ToolUsageStats
	taskToolMapping   map[types.TaskType][]string
	mutex            sync.RWMutex
}

// NewToolRecommendationEngine creates a new recommendation engine
func NewToolRecommendationEngine() *ToolRecommendationEngine {
	return &ToolRecommendationEngine{
		toolUsagePatterns: make(map[string]*types.ToolUsageStats),
		taskToolMapping: map[types.TaskType][]string{
			types.TaskTypeAnalysis:      {"file_read", "file_list", "grep", "bash"},
			types.TaskTypeGeneration:    {"file_update", "file_replace", "directory_create"},
			types.TaskTypeRefactor:      {"file_read", "file_replace", "bash"},
			types.TaskTypeTest:          {"bash", "file_read", "file_list"},
			types.TaskTypeExplain:       {"file_read", "file_list", "grep"},
			types.TaskTypeCustom:        {"file_read", "file_list", "bash"},
		},
	}
}

// NewToolSystemAdapter creates a new tool system adapter
func NewToolSystemAdapter(registry *registryPkg.Registry) *ToolSystemAdapter {
	adapter := &ToolSystemAdapter{
		registry:    registry,
		metrics:     initializeMetrics(),
		execHistory: make([]types.ToolExecutionRecord, 0, 1000),
		config:      getDefaultConfig(),
		recommender: NewToolRecommendationEngine(),
	}
	
	return adapter
}

// initializeMetrics creates default metrics
func initializeMetrics() *types.ToolSystemMetrics {
	return &types.ToolSystemMetrics{
		TotalExecutions:      0,
		SuccessfulExecutions: 0,
		FailedExecutions:     0,
		AverageExecutionTime: 0,
		ToolUsageStats:       make(map[string]*types.ToolUsageStats),
		PerformanceStats:     make(map[string]*types.ToolPerformance),
		ErrorStats:           make(map[string]int),
		LastUpdated:          time.Now(),
	}
}

// getDefaultConfig returns default tool system configuration
func getDefaultConfig() *types.ToolSystemConfig {
	return &types.ToolSystemConfig{
		MaxConcurrentExecutions: 5,
		DefaultTimeout:          30000, // 30 seconds in milliseconds
		SecurityConfig: &types.SecurityConfig{
			EnableSandbox:       true,
			MaxMemoryUsage:      1024 * 1024 * 1024, // 1GB
			MaxExecutionTime:    30000,               // 30 seconds
			AllowedTools:        []string{"file_read", "file_update", "file_replace", "file_list", "grep", "bash"},
			RestrictedTools:     []string{},
		},
		MonitoringConfig: &types.MonitoringConfig{
			Enabled:         true,
			MetricsInterval: 60, // 1 minute
			LogLevel:        "info",
		},
	}
}

// GetAvailableTools returns a list of available tool names
func (tsa *ToolSystemAdapter) GetAvailableTools() []string {
	return tsa.registry.ListTools()
}

// GetToolSchema returns the schema for a specific tool
func (tsa *ToolSystemAdapter) GetToolSchema(toolName string) (*types.ToolSchema, error) {
	tool := tsa.registry.GetTool(toolName)
	if tool == nil {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	metadata := tsa.registry.GetToolMetadata(toolName)
	
	schema := &types.ToolSchema{
		Name:        toolName,
		Description: tool.Description(),
		Parameters:  tool.Parameters(),
		Version:     "1.0.0",
		Category:    types.ToolCategoryCustom,
	}
	
	if metadata != nil {
		schema.Category = types.ToolCategory(metadata.Category)
		schema.Version = metadata.Version
	}
	
	return schema, nil
}

// ExecuteTool executes a tool with the given parameters
func (tsa *ToolSystemAdapter) ExecuteTool(ctx context.Context, toolCall *types.ToolCall, execContext *types.ToolExecutionContext) (*types.ToolResult, error) {
	startTime := time.Now()
	toolName := toolCall.Function.Name
	
	// Update metrics - total executions
	tsa.mutex.Lock()
	tsa.metrics.TotalExecutions++
	if tsa.metrics.ToolUsageStats[toolName] == nil {
		tsa.metrics.ToolUsageStats[toolName] = &types.ToolUsageStats{
			TotalCalls:  0,
			AverageTime: 0,
			LastReset:   time.Now(),
		}
	}
	tsa.metrics.ToolUsageStats[toolName].TotalCalls++
	tsa.mutex.Unlock()
	
	// Validate tool call
	validation, err := tsa.ValidateToolCall(toolCall)
	if err != nil {
		tsa.recordFailure(toolName, "validation_failed", time.Since(startTime))
		return nil, fmt.Errorf("tool validation failed: %v", err)
	}
	if validation != nil && !validation.Valid {
		tsa.recordFailure(toolName, "validation_failed", time.Since(startTime))
		return nil, fmt.Errorf("tool validation failed: invalid parameters")
	}
	
	// Apply timeout from context or config
	timeout := time.Duration(tsa.config.DefaultTimeout) * time.Millisecond
	if execContext != nil && execContext.Timeout > 0 {
		timeout = execContext.Timeout
	}
	
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Convert types.ToolCall to tools registry compatible format
	args := toolCall.Function.Arguments
	
	// Execute the tool using the registry
	result, err := tsa.registry.ExecuteTool(execCtx, toolName, args)
	
	duration := time.Since(startTime)
	
	if err != nil {
		tsa.recordFailure(toolName, err.Error(), duration)
		return &types.ToolResult{
			ID:      fmt.Sprintf("exec_%d", time.Now().UnixNano()),
			Success: false,
			Error:   err.Error(),
		}, nil // Return nil error but failed result
	}
	
	// Record successful execution
	tsa.recordSuccess(toolName, duration)
	
	// Convert tools.ToolResult to types.ToolResult
	execResult := &types.ToolResult{
		ID:      fmt.Sprintf("exec_%d", time.Now().UnixNano()),
		Content: result.Content,
		Data:    result.Data,
		Success: true,
	}
	
	// Record execution history
	tsa.recordExecution(toolCall, execResult, duration)
	
	return execResult, nil
}

// recordSuccess updates metrics for successful execution
func (tsa *ToolSystemAdapter) recordSuccess(toolName string, duration time.Duration) {
	tsa.mutex.Lock()
	defer tsa.mutex.Unlock()
	
	tsa.metrics.SuccessfulExecutions++
	
	// Update tool-specific metrics
	stats := tsa.metrics.ToolUsageStats[toolName]
	if stats != nil {
		// Update average time with running average
		stats.AverageTime = time.Duration((int64(stats.AverageTime)*int64(stats.TotalCalls-1) + int64(duration)) / int64(stats.TotalCalls))
	}
	
	// Update performance stats
	if tsa.metrics.PerformanceStats[toolName] == nil {
		tsa.metrics.PerformanceStats[toolName] = &types.ToolPerformance{
			AvgExecutionTime: duration,
			LastBenchmark:    time.Now(),
		}
	} else {
		perf := tsa.metrics.PerformanceStats[toolName]
		perf.AvgExecutionTime = time.Duration((int64(perf.AvgExecutionTime) + int64(duration)) / 2)
		perf.LastBenchmark = time.Now()
	}
	
	// Update system average
	tsa.metrics.AverageExecutionTime = time.Duration(
		(int64(tsa.metrics.AverageExecutionTime)*int64(tsa.metrics.SuccessfulExecutions-1) + int64(duration)) / 
		int64(tsa.metrics.SuccessfulExecutions))
	
	tsa.metrics.LastUpdated = time.Now()
}

// recordFailure updates metrics for failed execution
func (tsa *ToolSystemAdapter) recordFailure(toolName, errorMsg string, duration time.Duration) {
	tsa.mutex.Lock()
	defer tsa.mutex.Unlock()
	
	tsa.metrics.FailedExecutions++
	tsa.metrics.ErrorStats[errorMsg]++
	tsa.metrics.LastUpdated = time.Now()
}

// recordExecution adds execution to history
func (tsa *ToolSystemAdapter) recordExecution(toolCall *types.ToolCall, result *types.ToolResult, duration time.Duration) {
	tsa.mutex.Lock()
	defer tsa.mutex.Unlock()
	
	record := types.ToolExecutionRecord{
		ID:            result.ID,
		ToolName:      toolCall.Function.Name,
		Arguments:     toolCall.Function.Arguments,
		Success:       result.Success,
		ExecutionTime: duration,
		Timestamp:     time.Now(),
		Error:         result.Error,
	}
	
	// Keep only last 1000 records
	if len(tsa.execHistory) >= 1000 {
		tsa.execHistory = tsa.execHistory[1:]
	}
	
	tsa.execHistory = append(tsa.execHistory, record)
}

// ValidateToolCallSimple validates a tool call before execution (helper method)
func (tsa *ToolSystemAdapter) ValidateToolCallSimple(toolCall *types.ToolCall) error {
	return tsa.registry.ValidateToolArgs(toolCall.Function.Name, toolCall.Function.Arguments)
}

// GetToolsByCategoryAll returns tools grouped by category (helper method)
func (tsa *ToolSystemAdapter) GetToolsByCategoryAll() map[string][]string {
	return tsa.registry.ListToolsByCategory()
}

// RegisterTool registers a new tool (if the underlying registry supports it)
func (tsa *ToolSystemAdapter) RegisterTool(toolName string) error {
	// Tool registration not supported through adapter
	// Registry manages its own tools
	return fmt.Errorf("tool registration not supported through adapter")
}

// UnregisterTool removes a tool from the registry
func (tsa *ToolSystemAdapter) UnregisterTool(toolName string) error {
	return tsa.registry.UnregisterTool(toolName)
}

// GetMetrics returns tool system metrics
func (tsa *ToolSystemAdapter) GetMetrics() *types.ToolSystemMetrics {
	return &types.ToolSystemMetrics{
		TotalExecutions:      0, // Would need to track this
		SuccessfulExecutions: 0, // Would need to track this
		FailedExecutions:     0, // Would need to track this
		AverageExecutionTime: 0, // Would need to track this
		ToolUsageStats:       make(map[string]*types.ToolUsageStats),
		PerformanceStats:     make(map[string]*types.ToolPerformance),
		ErrorStats:           make(map[string]int),
		LastUpdated:          time.Now(),
	}
}

// IsToolAvailable checks if a tool is available
func (tsa *ToolSystemAdapter) IsToolAvailable(toolName string) bool {
	tool := tsa.registry.GetTool(toolName)
	return tool != nil
}

// GetToolDescription returns a human-readable description of a tool
func (tsa *ToolSystemAdapter) GetToolDescription(toolName string) string {
	tool := tsa.registry.GetTool(toolName)
	if tool == nil {
		return ""
	}
	return tool.Description()
}

// Additional methods required by interfaces.ToolOrchestrator

// ExecuteTools executes multiple tools
func (tsa *ToolSystemAdapter) ExecuteTools(ctx context.Context, toolCalls []types.ToolCall, execContext *types.ToolExecutionContext) ([]types.ToolResult, error) {
	var results []types.ToolResult
	for _, toolCall := range toolCalls {
		result, err := tsa.ExecuteTool(ctx, &toolCall, execContext)
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}
	return results, nil
}

// ExecutePlan executes an execution plan
func (tsa *ToolSystemAdapter) ExecutePlan(ctx context.Context, plan *types.ExecutionPlan, execContext *types.ToolExecutionContext) ([]types.ToolResult, error) {
	// Placeholder implementation
	return []types.ToolResult{}, nil
}

// GetTool gets a registered tool
func (tsa *ToolSystemAdapter) GetTool(name string) (*types.RegisteredTool, error) {
	tool := tsa.registry.GetTool(name)
	if tool == nil {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	
	metadata := tsa.registry.GetToolMetadata(name)
	
	// Convert tools.ToolMetadata to types.ToolMetadata if available
	var convertedMetadata *types.ToolMetadata
	if metadata != nil {
		convertedMetadata = &types.ToolMetadata{
			Usage: &types.ToolUsageStats{
				TotalCalls:     0,
				AverageTime:    0,
				LastReset:      time.Now(),
			},
			Performance: &types.ToolPerformance{
				AvgExecutionTime: 0,
				LastBenchmark:    time.Now(),
			},
		}
	}
	
	return &types.RegisteredTool{
		Schema: &types.ToolSchema{
			Name:        name,
			Description: tool.Description(),
			Parameters:  tool.Parameters(),
			Version:     "1.0.0",
		},
		Enabled:      true,
		RegisteredAt: time.Now(),
		Metadata:     convertedMetadata,
	}, nil
}

// GetToolsByCategory gets tools by category
func (tsa *ToolSystemAdapter) GetToolsByCategory(category types.ToolCategory) []string {
	allTools := tsa.registry.ListToolsByCategory()
	return allTools[string(category)]
}

// SearchTools searches for tools
func (tsa *ToolSystemAdapter) SearchTools(query string) ([]types.ToolRecommendation, error) {
	// Placeholder implementation
	return []types.ToolRecommendation{}, nil
}

// ValidateToolCall validates a tool call
func (tsa *ToolSystemAdapter) ValidateToolCall(toolCall *types.ToolCall) (*types.ToolCallValidation, error) {
	err := tsa.registry.ValidateToolArgs(toolCall.Function.Name, toolCall.Function.Arguments)
	if err != nil {
		return &types.ToolCallValidation{
			Valid: false,
			Errors: []types.ValidationError{
				{Message: err.Error()},
			},
			Risk: types.ToolRiskLevelMedium,
		}, nil
	}
	
	return &types.ToolCallValidation{
		Valid: true,
		Risk:  types.ToolRiskLevelLow,
	}, nil
}

// ValidateExecutionPlan validates an execution plan
func (tsa *ToolSystemAdapter) ValidateExecutionPlan(plan *types.ExecutionPlan) (*types.ValidationResult, error) {
	// Placeholder implementation
	return &types.ValidationResult{
		IsValid: true,
	}, nil
}

// CheckPermissions checks permissions for tool execution
func (tsa *ToolSystemAdapter) CheckPermissions(toolName string, execContext *types.ToolExecutionContext) error {
	// Placeholder implementation - in real system would check actual permissions
	return nil
}

// RecommendTools recommends tools for a task
func (tsa *ToolSystemAdapter) RecommendTools(ctx context.Context, task *types.Task) ([]types.ToolRecommendation, error) {
	// Always enable recommendations for now
	// if !tsa.config.EnableRecommendations {
	//     return []types.ToolRecommendation{}, nil
	// }
	
	tsa.recommender.mutex.RLock()
	defer tsa.recommender.mutex.RUnlock()
	
	// Get base recommendations for task type
	baseTools := tsa.recommender.taskToolMapping[task.Type]
	var recommendations []types.ToolRecommendation
	
	for _, toolName := range baseTools {
		// Check if tool is available
		if !tsa.IsToolAvailable(toolName) {
			continue
		}
		
		// Calculate confidence based on usage patterns and task context
		confidence := tsa.calculateToolConfidence(toolName, task)
		
		recommendation := types.ToolRecommendation{
			ToolName:    toolName,
			Confidence:  confidence,
			Rationale:   tsa.generateRecommendationReason(toolName, task),
			Risk:        types.ToolRiskLevelLow,
		}
		
		recommendations = append(recommendations, recommendation)
	}
	
	// Sort by confidence (highest first)
	for i := 0; i < len(recommendations)-1; i++ {
		for j := i + 1; j < len(recommendations); j++ {
			if recommendations[i].Confidence < recommendations[j].Confidence {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}
	
	// Return top 5 recommendations
	if len(recommendations) > 5 {
		recommendations = recommendations[:5]
	}
	
	return recommendations, nil
}

// calculateToolConfidence calculates confidence score for tool recommendation
func (tsa *ToolSystemAdapter) calculateToolConfidence(toolName string, task *types.Task) float64 {
	baseConfidence := 0.5 // Default confidence
	
	// Factor in usage statistics
	tsa.mutex.RLock()
	if stats, exists := tsa.metrics.ToolUsageStats[toolName]; exists {
		usageBoost := float64(stats.TotalCalls) / 100.0 // Normalize usage count
		if usageBoost > 0.3 {
			usageBoost = 0.3 // Cap the boost
		}
		baseConfidence += usageBoost
	}
	
	// Factor in performance (faster tools get higher confidence)
	if perf, exists := tsa.metrics.PerformanceStats[toolName]; exists {
		if perf.AvgExecutionTime < 1*time.Second {
			baseConfidence += 0.1
		} else if perf.AvgExecutionTime > 10*time.Second {
			baseConfidence -= 0.1
		}
	}
	tsa.mutex.RUnlock()
	
	// Factor in task context keywords
	taskDescription := strings.ToLower(task.Description)
	switch toolName {
	case "file_read":
		if strings.Contains(taskDescription, "read") || strings.Contains(taskDescription, "show") || strings.Contains(taskDescription, "analyze") {
			baseConfidence += 0.2
		}
	case "file_update", "file_replace":
		if strings.Contains(taskDescription, "write") || strings.Contains(taskDescription, "update") || strings.Contains(taskDescription, "modify") {
			baseConfidence += 0.2
		}
	case "bash":
		if strings.Contains(taskDescription, "run") || strings.Contains(taskDescription, "execute") || strings.Contains(taskDescription, "command") {
			baseConfidence += 0.2
		}
	case "grep":
		if strings.Contains(taskDescription, "search") || strings.Contains(taskDescription, "find") || strings.Contains(taskDescription, "grep") {
			baseConfidence += 0.2
		}
	}
	
	// Cap confidence at 1.0
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}
	
	return baseConfidence
}

// generateRecommendationReason generates explanation for tool recommendation
func (tsa *ToolSystemAdapter) generateRecommendationReason(toolName string, task *types.Task) string {
	switch toolName {
	case "file_read":
		return "Suitable for reading and analyzing file contents"
	case "file_update":
		return "Ideal for adding or modifying file content"
	case "file_replace":
		return "Perfect for replacing specific text or lines in files"
	case "file_list":
		return "Essential for exploring directory structure"
	case "bash":
		return "Powerful for executing system commands and scripts"
	case "grep":
		return "Excellent for searching text patterns in files"
	case "directory_create":
		return "Necessary for creating new directories"
	default:
		return fmt.Sprintf("Recommended for %s tasks", task.Type)
	}
}

// getToolCategory returns the category of a tool
func (tsa *ToolSystemAdapter) getToolCategory(toolName string) types.ToolCategory {
	switch {
	case strings.Contains(toolName, "file"):
		return types.ToolCategoryFile
	case toolName == "bash" || toolName == "script_runner":
		return types.ToolCategoryBash
	case toolName == "grep" || toolName == "find":
		return types.ToolCategorySearch
	default:
		return types.ToolCategoryCustom
	}
}

// CreateExecutionPlan creates an execution plan
func (tsa *ToolSystemAdapter) CreateExecutionPlan(ctx context.Context, toolCalls []types.ToolCall, strategy types.ExecutionStrategy) (*types.ExecutionPlan, error) {
	// Analyze dependencies between tool calls
	dependencies := tsa.analyzeDependencies(toolCalls)
	
	// Convert ToolCalls to ExecutionSteps
	steps := make([]types.ExecutionStep, len(toolCalls))
	for i, call := range toolCalls {
		stepID := fmt.Sprintf("step_%d", i+1)
		stepDeps := dependencies[stepID]
		
		steps[i] = types.ExecutionStep{
			ID:           stepID,
			ToolName:     call.Function.Name,
			Arguments:    call.Function.Arguments,
			Dependencies: stepDeps,
			Status:       types.StepStatusPending,
			Timeout:      time.Duration(tsa.config.DefaultTimeout) * time.Millisecond,
		}
	}
	
	plan := &types.ExecutionPlan{
		ID:             fmt.Sprintf("plan_%d", time.Now().UnixNano()),
		Name:           "Generated Plan",
		Description:    fmt.Sprintf("Execution plan for %d tools", len(toolCalls)),
		Steps:          steps,
		Strategy:       strategy,
		MaxConcurrency: tsa.config.MaxConcurrentExecutions,
		Timeout:        time.Duration(tsa.config.DefaultTimeout) * time.Millisecond,
	}
	
	return plan, nil
}

// analyzeDependencies analyzes dependencies between tool calls (simplified)
func (tsa *ToolSystemAdapter) analyzeDependencies(toolCalls []types.ToolCall) map[string][]string {
	dependencies := make(map[string][]string)
	
	// Simple dependency analysis - tools that modify files should run before tools that read them
	for i, call := range toolCalls {
		stepID := fmt.Sprintf("step_%d", i+1)
		for j, otherCall := range toolCalls {
			if i == j {
				continue
			}
			
			// Check if one tool writes and another reads the same file
			if tsa.hasFileDependency(call, otherCall) {
				otherStepID := fmt.Sprintf("step_%d", j+1)
				dependencies[otherStepID] = append(dependencies[otherStepID], stepID)
			}
		}
	}
	
	return dependencies
}

// hasFileDependency checks if two tools have file dependencies
func (tsa *ToolSystemAdapter) hasFileDependency(call1, call2 types.ToolCall) bool {
	// Check if call1 writes to a file that call2 reads
	if tsa.isFileWriteTool(call1.Function.Name) && tsa.isFileReadTool(call2.Function.Name) {
		file1 := tsa.extractFilePath(call1.Function.Arguments)
		file2 := tsa.extractFilePath(call2.Function.Arguments)
		return file1 != "" && file1 == file2
	}
	return false
}

// isFileWriteTool checks if a tool modifies files
func (tsa *ToolSystemAdapter) isFileWriteTool(toolName string) bool {
	writeTools := []string{"file_update", "file_replace", "directory_create"}
	for _, wt := range writeTools {
		if toolName == wt {
			return true
		}
	}
	return false
}

// isFileReadTool checks if a tool reads files
func (tsa *ToolSystemAdapter) isFileReadTool(toolName string) bool {
	readTools := []string{"file_read", "file_list", "grep"}
	for _, rt := range readTools {
		if toolName == rt {
			return true
		}
	}
	return false
}

// extractFilePath extracts file path from tool arguments
func (tsa *ToolSystemAdapter) extractFilePath(args map[string]interface{}) string {
	if path, ok := args["file_path"].(string); ok {
		return path
	}
	if path, ok := args["path"].(string); ok {
		return path
	}
	return ""
}

// optimizeExecutionOrder optimizes the order of tool execution based on strategy
func (tsa *ToolSystemAdapter) optimizeExecutionOrder(toolCalls []types.ToolCall, strategy types.ExecutionStrategy) ([]types.ToolCall, error) {
	switch strategy {
	case types.ExecutionStrategySequential:
		// Keep original order but ensure dependencies are respected
		return tsa.sortByDependencies(toolCalls), nil
	case types.ExecutionStrategyParallel:
		// Group independent tools together
		return tsa.groupForConcurrency(toolCalls), nil
	case types.ExecutionStrategyOptimized:
		// Balance concurrency with dependencies
		return tsa.optimizeForPerformance(toolCalls), nil
	case types.ExecutionStrategyAdaptive:
		// Use adaptive strategy based on current system load
		return tsa.optimizeForPerformance(toolCalls), nil
	default:
		return toolCalls, nil
	}
}

// sortByDependencies sorts tools to respect dependencies
func (tsa *ToolSystemAdapter) sortByDependencies(toolCalls []types.ToolCall) []types.ToolCall {
	// Simple topological sort - write tools before read tools
	var writeCalls, readCalls, otherCalls []types.ToolCall
	
	for _, call := range toolCalls {
		if tsa.isFileWriteTool(call.Function.Name) {
			writeCalls = append(writeCalls, call)
		} else if tsa.isFileReadTool(call.Function.Name) {
			readCalls = append(readCalls, call)
		} else {
			otherCalls = append(otherCalls, call)
		}
	}
	
	// Combine in order: writes, others, reads
	result := make([]types.ToolCall, 0, len(toolCalls))
	result = append(result, writeCalls...)
	result = append(result, otherCalls...)
	result = append(result, readCalls...)
	
	return result
}

// groupForConcurrency groups tools for concurrent execution
func (tsa *ToolSystemAdapter) groupForConcurrency(toolCalls []types.ToolCall) []types.ToolCall {
	// Group read-only tools together first, then write tools
	var readOnlyCalls, writeCalls []types.ToolCall
	
	for _, call := range toolCalls {
		if tsa.isFileReadTool(call.Function.Name) || call.Function.Name == "grep" {
			readOnlyCalls = append(readOnlyCalls, call)
		} else {
			writeCalls = append(writeCalls, call)
		}
	}
	
	// Combine: read-only first (can run concurrently), then writes (sequential)
	result := make([]types.ToolCall, 0, len(toolCalls))
	result = append(result, readOnlyCalls...)
	result = append(result, writeCalls...)
	
	return result
}

// optimizeForPerformance optimizes for best performance
func (tsa *ToolSystemAdapter) optimizeForPerformance(toolCalls []types.ToolCall) []types.ToolCall {
	// Sort by expected execution time (fastest first for better parallelization)
	sorted := make([]types.ToolCall, len(toolCalls))
	copy(sorted, toolCalls)
	
	// Simple bubble sort by estimated execution time
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			time1 := tsa.getToolExecutionTime(sorted[i].Function.Name)
			time2 := tsa.getToolExecutionTime(sorted[j].Function.Name)
			if time1 > time2 {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	return sorted
}

// getToolExecutionTime estimates execution time for a tool
func (tsa *ToolSystemAdapter) getToolExecutionTime(toolName string) time.Duration {
	tsa.mutex.RLock()
	defer tsa.mutex.RUnlock()
	
	if perf, exists := tsa.metrics.PerformanceStats[toolName]; exists {
		return perf.AvgExecutionTime
	}
	
	// Default estimates based on tool type
	switch toolName {
	case "file_read", "file_list":
		return 100 * time.Millisecond
	case "file_update", "file_replace":
		return 200 * time.Millisecond
	case "bash":
		return 1 * time.Second
	case "grep":
		return 300 * time.Millisecond
	default:
		return 500 * time.Millisecond
	}
}

// estimateExecutionDuration estimates total execution duration
func (tsa *ToolSystemAdapter) estimateExecutionDuration(toolCalls []types.ToolCall) time.Duration {
	var totalDuration time.Duration
	
	for _, call := range toolCalls {
		toolDuration := tsa.getToolExecutionTime(call.Function.Name)
		totalDuration += toolDuration
	}
	
	// Add 20% buffer for overhead
	return time.Duration(float64(totalDuration) * 1.2)
}

// OptimizePlan optimizes an execution plan
func (tsa *ToolSystemAdapter) OptimizePlan(ctx context.Context, plan *types.ExecutionPlan) (*types.ExecutionPlan, error) {
	// Placeholder implementation
	return plan, nil
}

// GetToolMetrics gets tool metrics
func (tsa *ToolSystemAdapter) GetToolMetrics(toolName string) (*types.ToolMetadata, error) {
	metadata := tsa.registry.GetToolMetadata(toolName)
	if metadata == nil {
		return &types.ToolMetadata{}, nil
	}
	
	// Convert tools.ToolMetadata to types.ToolMetadata
	return &types.ToolMetadata{
		Usage: &types.ToolUsageStats{
			TotalCalls:     0,
			AverageTime:    0,
			LastReset:      time.Now(),
		},
		Performance: &types.ToolPerformance{
			AvgExecutionTime: 0,
			LastBenchmark:    time.Now(),
		},
	}, nil
}

// GetSystemMetrics gets system metrics - alias for GetMetrics
func (tsa *ToolSystemAdapter) GetSystemMetrics() (*types.ToolSystemMetrics, error) {
	metrics := tsa.GetMetrics()
	return metrics, nil
}

// GetExecutionHistory gets execution history
func (tsa *ToolSystemAdapter) GetExecutionHistory(limit int) ([]types.ToolExecutionRecord, error) {
	tsa.mutex.RLock()
	defer tsa.mutex.RUnlock()
	
	if limit <= 0 {
		limit = len(tsa.execHistory)
	}
	
	// Return the most recent records up to the limit
	start := 0
	if len(tsa.execHistory) > limit {
		start = len(tsa.execHistory) - limit
	}
	
	result := make([]types.ToolExecutionRecord, len(tsa.execHistory)-start)
	copy(result, tsa.execHistory[start:])
	
	return result, nil
}

// Configure configures the tool system
func (tsa *ToolSystemAdapter) Configure(config *types.ToolSystemConfig) error {
	tsa.mutex.Lock()
	defer tsa.mutex.Unlock()
	
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	tsa.config = config
	return nil
}

// GetConfiguration gets the configuration
func (tsa *ToolSystemAdapter) GetConfiguration() *types.ToolSystemConfig {
	tsa.mutex.RLock()
	defer tsa.mutex.RUnlock()
	
	// Return a copy to prevent external modification
	configCopy := *tsa.config
	return &configCopy
}