package swe_bench

import (
	"context"
	"fmt"
	"time"
)

// This file contains a simplified agent implementation for SWE-bench batch processing
// In a full implementation, this would integrate with the actual Alex ReAct agent system

// AgentFactoryImpl implements the AgentFactory interface
type AgentFactoryImpl struct {
}

// NewAgentFactory creates a new agent factory
func NewAgentFactory() *AgentFactoryImpl {
	return &AgentFactoryImpl{}
}

// CreateAgent creates a new agent instance
func (af *AgentFactoryImpl) CreateAgent(ctx context.Context, config *BatchConfig) (Agent, error) {
	// For now, return the simple agent implementation
	return &SimpleAgent{
		config: config,
	}, nil
}

// ValidateConfig validates the agent configuration
func (af *AgentFactoryImpl) ValidateConfig(config *BatchConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Agent.Model.Name == "" {
		return fmt.Errorf("model name is required")
	}

	if config.Agent.Model.Temperature < 0 || config.Agent.Model.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	if config.Agent.MaxTurns <= 0 {
		return fmt.Errorf("max turns must be positive")
	}

	if config.Agent.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// SimpleAgent implements a basic agent for demonstration purposes
// TODO: Replace with actual Alex ReAct agent integration
type SimpleAgent struct {
	config *BatchConfig
}

// ProcessInstance processes a single SWE-Bench instance
func (sa *SimpleAgent) ProcessInstance(ctx context.Context, instance Instance) (*WorkerResult, error) {
	startTime := time.Now()

	// Simulate processing time based on problem complexity
	processingTime := time.Duration(100+len(instance.ProblemStatement)/10) * time.Millisecond
	select {
	case <-time.After(processingTime):
		// Normal completion
	case <-ctx.Done():
		// Context cancelled (timeout)
		return &WorkerResult{
			InstanceID: instance.ID,
			Status:     StatusTimeout,
			StartTime:  startTime,
			EndTime:    time.Now(),
			Duration:   time.Since(startTime),
			Error:      "Processing timed out",
			ErrorType:  "timeout_error",
		}, nil
	}

	// Create a realistic mock solution
	solution := sa.generateSolution(instance)
	explanation := sa.generateExplanation(instance)
	filesChanged := sa.identifyFilesToChange(instance)
	commands := sa.generateTestCommands(instance)
	trace := sa.createProcessingTrace(instance, startTime)

	result := &WorkerResult{
		InstanceID:   instance.ID,
		Status:       StatusCompleted,
		Solution:     solution,
		Explanation:  explanation,
		FilesChanged: filesChanged,
		Commands:     commands,
		StartTime:    startTime,
		EndTime:      time.Now(),
		Duration:     time.Since(startTime),
		TokensUsed:   sa.estimateTokenUsage(instance),
		Cost:         sa.estimateCost(instance),
		Trace:        trace,
	}

	return result, nil
}

// GetConfiguration returns the agent configuration
func (sa *SimpleAgent) GetConfiguration() map[string]interface{} {
	return map[string]interface{}{
		"model_name":  sa.config.Agent.Model.Name,
		"temperature": sa.config.Agent.Model.Temperature,
		"max_tokens":  sa.config.Agent.Model.MaxTokens,
		"max_turns":   sa.config.Agent.MaxTurns,
		"timeout":     sa.config.Agent.Timeout,
		"agent_type":  "alex_react_agent_mock",
		"description": "Mock implementation for SWE-bench batch processing demonstration",
	}
}

// Close releases agent resources
func (sa *SimpleAgent) Close() error {
	// In a real implementation, this would cleanup LLM connections, sessions, etc.
	return nil
}

// Helper methods for generating realistic mock responses

func (sa *SimpleAgent) generateSolution(instance Instance) string {
	solution := fmt.Sprintf("# Solution for %s\n\n", instance.ID)
	solution += "## Problem Analysis\n"
	solution += fmt.Sprintf("The issue described in the problem statement:\n%s\n\n",
		truncateString(instance.ProblemStatement, 300))

	solution += "## Proposed Fix\n"
	solution += "Based on the analysis of the repository and the problem description, "
	solution += "the following changes are needed:\n\n"
	solution += "1. Identify the root cause of the issue\n"
	solution += "2. Implement the necessary code changes\n"
	solution += "3. Add or update tests to verify the fix\n"
	solution += "4. Ensure backward compatibility\n\n"

	solution += "## Implementation Details\n"
	solution += "```python\n# Example fix (this is a mock implementation)\n"
	solution += "def fixed_function():\n"
	solution += "    # Implementation that addresses the issue\n"
	solution += "    pass\n```\n\n"

	if instance.Hints != "" {
		solution += "## Additional Notes\n"
		solution += fmt.Sprintf("Taking into account the provided hints: %s\n",
			truncateString(instance.Hints, 200))
	}

	return solution
}

func (sa *SimpleAgent) generateExplanation(instance Instance) string {
	return fmt.Sprintf("This solution addresses the issue in %s by implementing the necessary "+
		"changes to fix the reported problem. The approach involves analyzing the codebase, "+
		"identifying the root cause, and applying a targeted fix while maintaining "+
		"compatibility with existing functionality.", instance.ID)
}

func (sa *SimpleAgent) identifyFilesToChange(instance Instance) []string {
	// Mock file identification based on common patterns
	files := []string{}

	// Try to extract potential file names from the problem statement
	problemLower := fmt.Sprintf("%s %s", instance.ProblemStatement, instance.Hints)

	if contains(problemLower, "model") || contains(problemLower, "django") {
		files = append(files, "models.py")
	}
	if contains(problemLower, "view") {
		files = append(files, "views.py")
	}
	if contains(problemLower, "test") {
		files = append(files, "tests.py")
	}
	if contains(problemLower, "util") {
		files = append(files, "utils.py")
	}

	// Default fallback
	if len(files) == 0 {
		files = append(files, "main.py")
	}

	return files
}

func (sa *SimpleAgent) generateTestCommands(instance Instance) []string {
	commands := []string{}

	// Common test patterns
	if contains(instance.RepoURL, "django") {
		commands = append(commands, "python manage.py test")
	} else if contains(instance.ProblemStatement, "pytest") {
		commands = append(commands, "python -m pytest tests/")
	} else {
		commands = append(commands, "python -m unittest discover")
	}

	return commands
}

func (sa *SimpleAgent) createProcessingTrace(instance Instance, startTime time.Time) []TraceStep {
	trace := []TraceStep{
		{
			Step:        1,
			Action:      "analyze_repository",
			Observation: fmt.Sprintf("Analyzed repository structure for %s", instance.RepoURL),
			Thought:     "Understanding the codebase structure and identifying relevant files",
			Timestamp:   startTime,
		},
		{
			Step:        2,
			Action:      "read_problem_statement",
			Observation: "Read and analyzed the problem statement",
			Thought:     "Understanding the specific issue that needs to be resolved",
			Timestamp:   startTime.Add(20 * time.Millisecond),
		},
		{
			Step:        3,
			Action:      "identify_root_cause",
			Observation: "Identified potential root cause of the issue",
			Thought:     "Located the specific code section that needs modification",
			Timestamp:   startTime.Add(50 * time.Millisecond),
		},
		{
			Step:        4,
			Action:      "implement_solution",
			Observation: "Implemented the necessary code changes",
			Thought:     "Applied the fix while ensuring compatibility",
			Timestamp:   startTime.Add(80 * time.Millisecond),
		},
	}

	return trace
}

func (sa *SimpleAgent) estimateTokenUsage(instance Instance) int {
	// Simple estimation based on problem statement length
	baseTokens := 200
	problemTokens := len(instance.ProblemStatement) / 4 // rough estimate
	hintsTokens := len(instance.Hints) / 4

	return baseTokens + problemTokens + hintsTokens
}

func (sa *SimpleAgent) estimateCost(instance Instance) float64 {
	tokens := sa.estimateTokenUsage(instance)

	// Mock cost calculation (varies by model)
	var costPerToken float64
	switch {
	case contains(sa.config.Agent.Model.Name, "gpt-4"):
		costPerToken = 0.00003 // $0.03 per 1K tokens
	case contains(sa.config.Agent.Model.Name, "gpt-3.5"):
		costPerToken = 0.000002 // $0.002 per 1K tokens
	case contains(sa.config.Agent.Model.Name, "deepseek"):
		costPerToken = 0.0000005 // Very low cost for free tier
	default:
		costPerToken = 0.000005 // Default estimate
	}

	return float64(tokens) * costPerToken
}

// Helper functions

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					findInString(s, substr))))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
