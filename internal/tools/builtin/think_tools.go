package builtin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"deep-coding-agent/internal/config"
	"deep-coding-agent/internal/llm"
)

// ThinkTool implements strategic thinking and reasoning phases
type ThinkTool struct {
	configManager *config.Manager
}

func NewThinkTool(configManager *config.Manager) *ThinkTool {
	return &ThinkTool{
		configManager: configManager,
	}
}

func (t *ThinkTool) Name() string {
	return "think"
}

func (t *ThinkTool) Description() string {
	return "Perform strategic thinking, reasoning, and analysis. Use for complex problem solving, planning, and reflection phases."
}

func (t *ThinkTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"phase": map[string]interface{}{
				"type":        "string",
				"description": "Thinking phase: analyze, plan, reflect, reason, ultra_think",
				"enum":        []string{"analyze", "plan", "reflect", "reason", "ultra_think"},
				"default":     "reason",
			},
			"context": map[string]interface{}{
				"type":        "string",
				"description": "Context or situation to think about (required)",
			},
			"goal": map[string]interface{}{
				"type":        "string",
				"description": "Specific goal or objective for the thinking session",
			},
			"constraints": map[string]interface{}{
				"type":        "array",
				"description": "Constraints or limitations to consider",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"depth": map[string]interface{}{
				"type":        "string",
				"description": "Thinking depth: shallow, normal, deep, ultra",
				"enum":        []string{"shallow", "normal", "deep", "ultra"},
				"default":     "normal",
			},
			"focus_areas": map[string]interface{}{
				"type":        "array",
				"description": "Specific areas to focus thinking on",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"context"},
	}
}

func (t *ThinkTool) Validate(args map[string]interface{}) error {
	validator := NewValidationFramework().
		AddCustomValidator("phase", "Thinking phase (analyze, plan, reflect, reason, ultra_think)", false, func(value interface{}) error {
			if value == nil {
				return nil
			}
			phase, ok := value.(string)
			if !ok {
				return fmt.Errorf("phase must be a string")
			}
			validPhases := []string{"analyze", "plan", "reflect", "reason", "ultra_think"}
			for _, vp := range validPhases {
				if phase == vp {
					return nil
				}
			}
			return fmt.Errorf("invalid phase: %s", phase)
		}).
		AddStringField("context", "Context or situation to think about").
		AddOptionalStringField("goal", "Specific goal or objective").
		AddCustomValidator("depth", "Thinking depth (shallow, normal, deep, ultra)", false, func(value interface{}) error {
			if value == nil {
				return nil
			}
			depth, ok := value.(string)
			if !ok {
				return fmt.Errorf("depth must be a string")
			}
			validDepths := []string{"shallow", "normal", "deep", "ultra"}
			for _, vd := range validDepths {
				if depth == vd {
					return nil
				}
			}
			return fmt.Errorf("invalid depth: %s", depth)
		})

	return validator.Validate(args)
}

func (t *ThinkTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	phase := "reason"
	if p, ok := args["phase"].(string); ok {
		phase = p
	}

	context := args["context"].(string)
	
	goal := ""
	if g, ok := args["goal"].(string); ok {
		goal = g
	}

	depth := "normal"
	if d, ok := args["depth"].(string); ok {
		depth = d
	}

	var constraints []string
	if c, ok := args["constraints"].([]interface{}); ok {
		for _, constraint := range c {
			if cs, ok := constraint.(string); ok {
				constraints = append(constraints, cs)
			}
		}
	}

	var focusAreas []string
	if fa, ok := args["focus_areas"].([]interface{}); ok {
		for _, area := range fa {
			if as, ok := area.(string); ok {
				focusAreas = append(focusAreas, as)
			}
		}
	}

	switch phase {
	case "analyze":
		return t.performAnalysis(ctx, context, goal, constraints, focusAreas, depth)
	case "plan":
		return t.performPlanning(ctx, context, goal, constraints, focusAreas, depth)
	case "reflect":
		return t.performReflection(ctx, context, goal, constraints, focusAreas, depth)
	case "reason":
		return t.performReasoning(ctx, context, goal, constraints, focusAreas, depth)
	case "ultra_think":
		return t.performUltraThink(ctx, context, goal, constraints, focusAreas)
	default:
		return nil, fmt.Errorf("unsupported thinking phase: %s", phase)
	}
}

func (t *ThinkTool) performAnalysis(ctx context.Context, context, goal string, constraints, focusAreas []string, depth string) (*ToolResult, error) {
	prompt := t.buildAnalysisPrompt(context, goal, constraints, focusAreas, depth)
	
	result, err := t.executeLLMThinking(ctx, prompt, "analysis", depth)
	if err != nil {
		return nil, err
	}

	return &ToolResult{
		Content: fmt.Sprintf("🔍 **Analysis Complete**\n\n%s", result),
		Data: map[string]interface{}{
			"phase":       "analyze",
			"depth":       depth,
			"context":     context,
			"goal":        goal,
			"constraints": constraints,
			"focus_areas": focusAreas,
			"thinking":    result,
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

func (t *ThinkTool) performPlanning(ctx context.Context, context, goal string, constraints, focusAreas []string, depth string) (*ToolResult, error) {
	prompt := t.buildPlanningPrompt(context, goal, constraints, focusAreas, depth)
	
	result, err := t.executeLLMThinking(ctx, prompt, "planning", depth)
	if err != nil {
		return nil, err
	}

	return &ToolResult{
		Content: fmt.Sprintf("📋 **Strategic Plan**\n\n%s", result),
		Data: map[string]interface{}{
			"phase":       "plan",
			"depth":       depth,
			"context":     context,
			"goal":        goal,
			"constraints": constraints,
			"focus_areas": focusAreas,
			"plan":        result,
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

func (t *ThinkTool) performReflection(ctx context.Context, context, goal string, constraints, focusAreas []string, depth string) (*ToolResult, error) {
	prompt := t.buildReflectionPrompt(context, goal, constraints, focusAreas, depth)
	
	result, err := t.executeLLMThinking(ctx, prompt, "reflection", depth)
	if err != nil {
		return nil, err
	}

	return &ToolResult{
		Content: fmt.Sprintf("🤔 **Reflection & Insights**\n\n%s", result),
		Data: map[string]interface{}{
			"phase":       "reflect",
			"depth":       depth,
			"context":     context,
			"goal":        goal,
			"constraints": constraints,
			"focus_areas": focusAreas,
			"reflection":  result,
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

func (t *ThinkTool) performReasoning(ctx context.Context, context, goal string, constraints, focusAreas []string, depth string) (*ToolResult, error) {
	prompt := t.buildReasoningPrompt(context, goal, constraints, focusAreas, depth)
	
	result, err := t.executeLLMThinking(ctx, prompt, "reasoning", depth)
	if err != nil {
		return nil, err
	}

	return &ToolResult{
		Content: fmt.Sprintf("🧠 **Reasoning Process**\n\n%s", result),
		Data: map[string]interface{}{
			"phase":       "reason",
			"depth":       depth,
			"context":     context,
			"goal":        goal,
			"constraints": constraints,
			"focus_areas": focusAreas,
			"reasoning":   result,
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

func (t *ThinkTool) performUltraThink(ctx context.Context, context, goal string, constraints, focusAreas []string) (*ToolResult, error) {
	prompt := t.buildUltraThinkPrompt(context, goal, constraints, focusAreas)
	
	result, err := t.executeLLMThinking(ctx, prompt, "ultra_think", "ultra")
	if err != nil {
		return nil, err
	}

	return &ToolResult{
		Content: fmt.Sprintf("🚀 **Ultra Think - Deep Cognitive Processing**\n\n%s", result),
		Data: map[string]interface{}{
			"phase":       "ultra_think",
			"depth":       "ultra",
			"context":     context,
			"goal":        goal,
			"constraints": constraints,
			"focus_areas": focusAreas,
			"ultra_think": result,
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

func (t *ThinkTool) buildAnalysisPrompt(context, goal string, constraints, focusAreas []string, depth string) string {
	var prompt strings.Builder
	
	prompt.WriteString("🔍 **ANALYSIS PHASE**\n\n")
	prompt.WriteString(fmt.Sprintf("**Context:** %s\n\n", context))
	
	if goal != "" {
		prompt.WriteString(fmt.Sprintf("**Goal:** %s\n\n", goal))
	}
	
	if len(constraints) > 0 {
		prompt.WriteString("**Constraints:**\n")
		for _, constraint := range constraints {
			prompt.WriteString(fmt.Sprintf("- %s\n", constraint))
		}
		prompt.WriteString("\n")
	}
	
	if len(focusAreas) > 0 {
		prompt.WriteString("**Focus Areas:**\n")
		for _, area := range focusAreas {
			prompt.WriteString(fmt.Sprintf("- %s\n", area))
		}
		prompt.WriteString("\n")
	}

	switch depth {
	case "shallow":
		prompt.WriteString("Provide a brief analysis focusing on the key aspects and immediate implications.")
	case "normal":
		prompt.WriteString("Analyze the situation systematically, identifying key components, relationships, and implications.")
	case "deep":
		prompt.WriteString("Conduct a comprehensive analysis examining all dimensions, underlying factors, potential consequences, and interconnections.")
	case "ultra":
		prompt.WriteString("Perform an exhaustive multi-dimensional analysis considering all possible angles, edge cases, systemic implications, and long-term effects.")
	}

	return prompt.String()
}

func (t *ThinkTool) buildPlanningPrompt(context, goal string, constraints, focusAreas []string, depth string) string {
	var prompt strings.Builder
	
	prompt.WriteString("📋 **PLANNING PHASE**\n\n")
	prompt.WriteString(fmt.Sprintf("**Context:** %s\n\n", context))
	
	if goal != "" {
		prompt.WriteString(fmt.Sprintf("**Goal:** %s\n\n", goal))
	}
	
	if len(constraints) > 0 {
		prompt.WriteString("**Constraints:**\n")
		for _, constraint := range constraints {
			prompt.WriteString(fmt.Sprintf("- %s\n", constraint))
		}
		prompt.WriteString("\n")
	}
	
	if len(focusAreas) > 0 {
		prompt.WriteString("**Focus Areas:**\n")
		for _, area := range focusAreas {
			prompt.WriteString(fmt.Sprintf("- %s\n", area))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("**Planning Instructions:**\n")
	switch depth {
	case "shallow":
		prompt.WriteString("Create a basic plan with the essential steps needed to achieve the goal.")
	case "normal":
		prompt.WriteString("Develop a comprehensive plan with clear steps, priorities, dependencies, and success criteria.")
	case "deep":
		prompt.WriteString("Design a detailed strategic plan including steps, alternatives, risk mitigation, resource requirements, and contingency planning.")
	case "ultra":
		prompt.WriteString("Architect a comprehensive strategic framework with multiple scenarios, detailed execution paths, resource optimization, risk analysis, and adaptive mechanisms.")
	}

	return prompt.String()
}

func (t *ThinkTool) buildReflectionPrompt(context, goal string, constraints, focusAreas []string, depth string) string {
	var prompt strings.Builder
	
	prompt.WriteString("🤔 **REFLECTION PHASE**\n\n")
	prompt.WriteString(fmt.Sprintf("**Context:** %s\n\n", context))
	
	if goal != "" {
		prompt.WriteString(fmt.Sprintf("**Goal:** %s\n\n", goal))
	}
	
	prompt.WriteString("**Reflection Instructions:**\n")
	prompt.WriteString("Reflect on the current situation and provide insights on:\n")
	prompt.WriteString("- What has been learned\n")
	prompt.WriteString("- What could be improved\n")
	prompt.WriteString("- Alternative approaches\n")
	prompt.WriteString("- Lessons for future reference\n\n")

	switch depth {
	case "shallow":
		prompt.WriteString("Provide key insights and immediate takeaways.")
	case "normal":
		prompt.WriteString("Conduct thoughtful reflection on patterns, effectiveness, and potential improvements.")
	case "deep":
		prompt.WriteString("Engage in deep introspection examining underlying assumptions, systemic patterns, and transformative insights.")
	case "ultra":
		prompt.WriteString("Perform comprehensive meta-cognitive reflection on all levels of understanding, mental models, and paradigmatic shifts.")
	}

	return prompt.String()
}

func (t *ThinkTool) buildReasoningPrompt(context, goal string, constraints, focusAreas []string, depth string) string {
	var prompt strings.Builder
	
	prompt.WriteString("🧠 **REASONING PHASE**\n\n")
	prompt.WriteString(fmt.Sprintf("**Context:** %s\n\n", context))
	
	if goal != "" {
		prompt.WriteString(fmt.Sprintf("**Goal:** %s\n\n", goal))
	}
	
	prompt.WriteString("**Reasoning Instructions:**\n")
	prompt.WriteString("Apply logical reasoning to:\n")
	prompt.WriteString("- Analyze cause and effect relationships\n")
	prompt.WriteString("- Evaluate different options\n")
	prompt.WriteString("- Make logical deductions\n")
	prompt.WriteString("- Identify potential solutions\n\n")

	switch depth {
	case "shallow":
		prompt.WriteString("Apply basic logical reasoning to reach conclusions.")
	case "normal":
		prompt.WriteString("Use systematic reasoning with clear logical steps and evidence-based conclusions.")
	case "deep":
		prompt.WriteString("Employ sophisticated reasoning techniques including formal logic, probabilistic reasoning, and multi-criteria analysis.")
	case "ultra":
		prompt.WriteString("Apply advanced cognitive frameworks including systems thinking, dialectical reasoning, and meta-logical analysis.")
	}

	return prompt.String()
}

func (t *ThinkTool) buildUltraThinkPrompt(context, goal string, constraints, focusAreas []string) string {
	var prompt strings.Builder
	
	prompt.WriteString("🚀 **ULTRA THINK - DEEP COGNITIVE PROCESSING**\n\n")
	prompt.WriteString("Engage in the highest level of cognitive processing, combining multiple thinking modes:\n\n")
	
	prompt.WriteString(fmt.Sprintf("**Context:** %s\n\n", context))
	
	if goal != "" {
		prompt.WriteString(fmt.Sprintf("**Ultimate Goal:** %s\n\n", goal))
	}
	
	prompt.WriteString("**Ultra Think Framework:**\n")
	prompt.WriteString("1. **Multi-Dimensional Analysis**: Examine from technical, strategic, ethical, and systemic perspectives\n")
	prompt.WriteString("2. **Pattern Recognition**: Identify deep patterns and meta-patterns\n")
	prompt.WriteString("3. **Synthetic Thinking**: Combine insights from different domains\n")
	prompt.WriteString("4. **Predictive Modeling**: Anticipate consequences and emergent behaviors\n")
	prompt.WriteString("5. **Creative Synthesis**: Generate novel approaches and solutions\n")
	prompt.WriteString("6. **Meta-Cognitive Reflection**: Think about thinking processes themselves\n\n")
	
	if len(constraints) > 0 {
		prompt.WriteString("**Constraints to Navigate:**\n")
		for _, constraint := range constraints {
			prompt.WriteString(fmt.Sprintf("- %s\n", constraint))
		}
		prompt.WriteString("\n")
	}
	
	if len(focusAreas) > 0 {
		prompt.WriteString("**Special Focus Areas:**\n")
		for _, area := range focusAreas {
			prompt.WriteString(fmt.Sprintf("- %s\n", area))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("**Ultra Think Output Requirements:**\n")
	prompt.WriteString("- Provide breakthrough insights and innovative solutions\n")
	prompt.WriteString("- Identify hidden opportunities and potential pitfalls\n")
	prompt.WriteString("- Suggest paradigm shifts or transformative approaches\n")
	prompt.WriteString("- Consider long-term implications and systemic effects\n")
	prompt.WriteString("- Synthesize insights into actionable wisdom\n")

	return prompt.String()
}

func (t *ThinkTool) executeLLMThinking(ctx context.Context, prompt, phase, depth string) (string, error) {
	// Get LLM configuration
	llmConfig := t.configManager.GetLLMConfig()
	
	// Use reasoning model for deep thinking
	modelType := llm.BasicModel
	if depth == "deep" || depth == "ultra" || phase == "ultra_think" {
		modelType = llm.ReasoningModel
	}

	// Get LLM client
	client, err := llm.GetLLMInstance(modelType)
	if err != nil {
		return "", fmt.Errorf("failed to get LLM instance: %w", err)
	}

	// Configure request based on thinking depth
	request := &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: prompt},
		},
		ModelType:  modelType,
		Tools:      nil, // No tools during pure thinking
		ToolChoice: "none",
		Config:     llmConfig,
	}

	// Adjust parameters for thinking depth
	if depth == "ultra" || phase == "ultra_think" {
		// Ultra thinking gets more tokens and higher temperature for creativity
		if request.Config.MaxTokens < 4000 {
			request.Config.MaxTokens = 4000
		}
		if request.Config.Temperature < 0.7 {
			request.Config.Temperature = 0.7
		}
	}

	// Execute thinking
	response, err := client.Chat(ctx, request)
	if err != nil {
		return "", fmt.Errorf("thinking execution failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no thinking response received")
	}

	result := strings.TrimSpace(response.Choices[0].Message.Content)
	return result, nil
}