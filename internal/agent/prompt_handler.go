package agent

import (
	"alex/pkg/types"
	"fmt"
	"log"
	"time"
)

// PromptHandler handles prompt generation and management
type PromptHandler struct {
	promptBuilder *LightPromptBuilder
}

// NewPromptHandler creates a new prompt handler
func NewPromptHandler(promptBuilder *LightPromptBuilder) *PromptHandler {
	return &PromptHandler{
		promptBuilder: promptBuilder,
	}
}

// buildToolDrivenTaskPrompt - 构建工具驱动的任务提示
func (h *PromptHandler) buildToolDrivenTaskPrompt(taskCtx *types.ReactTaskContext) string {
	// 使用项目内的prompt builder
	if h.promptBuilder != nil && h.promptBuilder.promptLoader != nil {
		// 尝试使用React thinking prompt作为基础模板
		template, err := h.promptBuilder.promptLoader.GetReActThinkingPrompt(taskCtx)
		if err != nil {
			log.Printf("[WARN] PromptHandler: Failed to get ReAct thinking prompt, trying fallback: %v", err)
		}
		// 构建增强的任务提示，将特定任务信息与ReAct模板结合
		return template
	}

	// Fallback to hardcoded prompt if prompt builder is not available
	log.Printf("[WARN] PromptHandler: Prompt builder not available, using hardcoded prompt")
	return h.buildHardcodedTaskPrompt(taskCtx)
}

// buildHardcodedTaskPrompt - 构建硬编码的任务提示（fallback）
func (h *PromptHandler) buildHardcodedTaskPrompt(taskCtx *types.ReactTaskContext) string {
	return fmt.Sprintf(`You are an intelligent agent with access to powerful tools. Your goal is to complete this task efficiently:

**WorkingDir:** %s

**Goal:** %s

**DirectoryInfo:** %s

**Memory:** %s

**Time:** %s

**Approach:**
1. **For complex tasks**: Start with the 'think' tool to analyze and plan
2. **For multi-step tasks**: Use 'todo_update' to create structured task lists
3. **For file operations**: Use appropriate file tools (file_read, file_update, etc.)
4. **For system operations**: Use bash tool when needed
5. **For search/analysis**: Use grep or other search tools

**Think Tool Capabilities:**
- Phase: analyze, plan, reflect, reason, ultra_think
- Depth: shallow, normal, deep, ultra
- Use for strategic thinking and problem breakdown

**Todo Management:**
- todo_update: Create, batch create, update, complete tasks
- todo_read: Read current todos with filtering and statistics

**Guidelines:**
- Use the 'think' tool first for complex problems requiring analysis
- Break down multi-step tasks using todo_update
- Execute tools systematically to achieve the goal
- Provide clear, actionable results

Begin by determining the best approach for this task.`, taskCtx.WorkingDir, taskCtx.Goal, taskCtx.DirectoryInfo.Description, taskCtx.Memory, time.Now().Format(time.RFC3339))
}
