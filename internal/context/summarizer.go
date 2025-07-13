package context

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"alex/internal/llm"
	"alex/internal/session"
)

// MessageSummarizer handles intelligent conversation summarization
type MessageSummarizer struct {
	llmClient llm.Client
	config    *ContextLengthConfig
}

// MessageSummary represents a structured summary of conversation messages
type MessageSummary struct {
	Summary     string            `json:"summary"`
	KeyPoints   []string          `json:"key_points"`
	Topics      []string          `json:"topics"`
	ActionItems []string          `json:"action_items"`
	Decisions   []string          `json:"decisions"`
	CodeChanges []CodeChangeInfo  `json:"code_changes"`
	Context     map[string]string `json:"context"`
	TokensUsed  int               `json:"tokens_used"`
	CreatedAt   time.Time         `json:"created_at"`
}

// CodeChangeInfo represents information about code changes discussed
type CodeChangeInfo struct {
	File        string `json:"file"`
	Description string `json:"description"`
	Type        string `json:"type"` // "created", "modified", "deleted"
}

// NewMessageSummarizer creates a new message summarizer
func NewMessageSummarizer(llmClient llm.Client, config *ContextLengthConfig) *MessageSummarizer {
	return &MessageSummarizer{
		llmClient: llmClient,
		config:    config,
	}
}

// SummarizeMessages creates an intelligent summary of conversation messages
func (ms *MessageSummarizer) SummarizeMessages(ctx context.Context, messages []*session.Message) (*MessageSummary, error) {
	if len(messages) == 0 {
		return &MessageSummary{
			Summary:   "No messages to summarize",
			CreatedAt: time.Now(),
		}, nil
	}

	// Convert messages to conversation text
	conversationText := ms.formatMessagesForSummarization(messages)

	// Create summarization prompt
	prompt := ms.buildSummarizationPrompt(conversationText, len(messages))

	// Call LLM for summarization
	req := &llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: "You are an expert conversation summarizer. Your task is to create structured, comprehensive summaries that preserve all important information while being concise and organized.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ModelType:   llm.ReasoningModel, // Use reasoning model for complex summarization
		Temperature: 0.1,                // Low temperature for consistent summaries
		MaxTokens:   2000,               // Generous token limit for detailed summaries
	}

	response, err := ms.llmClient.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no summary generated")
	}

	// Parse the structured response
	summary, err := ms.parseSummaryResponse(response.Choices[0].Message.Content)
	if err != nil {
		// Fallback to simple summary if parsing fails
		summary = &MessageSummary{
			Summary:   response.Choices[0].Message.Content,
			KeyPoints: []string{},
			Topics:    []string{},
			CreatedAt: time.Now(),
		}
	}

	// Add metadata using compatible method
	usage := response.GetUsage()
	summary.TokensUsed = usage.GetTotalTokens()
	summary.CreatedAt = time.Now()

	return summary, nil
}

// formatMessagesForSummarization converts messages to a clean text format
func (ms *MessageSummarizer) formatMessagesForSummarization(messages []*session.Message) string {
	var parts []string

	for i, msg := range messages {
		// Skip system messages except summaries
		if msg.Role == "system" {
			if metadata, ok := msg.Metadata["type"]; ok && metadata == "context_summary" {
				parts = append(parts, fmt.Sprintf("[PREVIOUS SUMMARY]\n%s\n", msg.Content))
			}
			continue
		}

		// Format user/assistant messages
		timestamp := ""
		if !msg.Timestamp.IsZero() {
			timestamp = fmt.Sprintf(" (%s)", msg.Timestamp.Format("15:04:05"))
		}

		role := strings.ToUpper(msg.Role)
		content := strings.TrimSpace(msg.Content)

		// Add context from tool calls if present
		toolInfo := ""
		if len(msg.ToolCalls) > 0 {
			var toolNames []string
			for _, tc := range msg.ToolCalls {
				toolNames = append(toolNames, tc.Name)
			}
			toolInfo = fmt.Sprintf(" [Tools: %s]", strings.Join(toolNames, ", "))
		}

		parts = append(parts, fmt.Sprintf("%d. %s%s%s:\n%s\n", i+1, role, timestamp, toolInfo, content))
	}

	return strings.Join(parts, "\n")
}

// buildSummarizationPrompt creates a comprehensive prompt for summarization
func (ms *MessageSummarizer) buildSummarizationPrompt(conversationText string, messageCount int) string {
	return fmt.Sprintf(`Please analyze and summarize the following conversation (%d messages) in a structured format.

CONVERSATION:
%s

Please provide a JSON response with the following structure:
{
  "summary": "A comprehensive 2-3 paragraph summary of the main conversation flow and outcomes",
  "key_points": ["List of 5-10 most important points or findings"],
  "topics": ["List of main topics discussed"],
  "action_items": ["List of any tasks, TODOs, or action items mentioned"],
  "decisions": ["List of any decisions made or conclusions reached"],
  "code_changes": [{"file": "filename", "description": "what changed", "type": "created/modified/deleted"}],
  "context": {"important_context_key": "context_value"}
}

Focus on:
1. Preserving technical details and code-related discussions
2. Maintaining the logical flow of problem-solving
3. Including all important decisions and their reasoning
4. Noting any unresolved issues or ongoing work
5. Capturing the current state and next steps

Be comprehensive but concise. This summary will be used to maintain context in an ongoing conversation.`, messageCount, conversationText)
}

// parseSummaryResponse attempts to parse the structured JSON response
func (ms *MessageSummarizer) parseSummaryResponse(content string) (*MessageSummary, error) {
	// Try to extract JSON from the response
	content = strings.TrimSpace(content)

	// Find JSON block (could be wrapped in markdown code blocks)
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonContent := content[jsonStart : jsonEnd+1]

	var summary MessageSummary
	err := json.Unmarshal([]byte(jsonContent), &summary)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate required fields
	if summary.Summary == "" {
		return nil, fmt.Errorf("summary is empty")
	}

	// Initialize arrays if nil
	if summary.KeyPoints == nil {
		summary.KeyPoints = []string{}
	}
	if summary.Topics == nil {
		summary.Topics = []string{}
	}
	if summary.ActionItems == nil {
		summary.ActionItems = []string{}
	}
	if summary.Decisions == nil {
		summary.Decisions = []string{}
	}
	if summary.CodeChanges == nil {
		summary.CodeChanges = []CodeChangeInfo{}
	}
	if summary.Context == nil {
		summary.Context = make(map[string]string)
	}

	return &summary, nil
}

// CompactSummarize creates a shorter summary for space-constrained contexts
func (ms *MessageSummarizer) CompactSummarize(ctx context.Context, messages []*session.Message) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	conversationText := ms.formatMessagesForSummarization(messages)

	prompt := fmt.Sprintf(`Provide a very concise summary of this conversation in 2-3 sentences:

%s

Focus only on the most essential information, main outcomes, and current state.`, conversationText)

	req := &llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: "You are a concise summarizer. Provide only the most essential information in 2-3 sentences.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ModelType:   llm.BasicModel, // Use basic model for simple summarization
		Temperature: 0.1,
		MaxTokens:   150,
	}

	response, err := ms.llmClient.Chat(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate compact summary: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no summary generated")
	}

	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}

// GetSummaryStats returns statistics about the summarization process
func (ms *MessageSummarizer) GetSummaryStats(summary *MessageSummary) map[string]interface{} {
	return map[string]interface{}{
		"summary_length":     len(summary.Summary),
		"key_points_count":   len(summary.KeyPoints),
		"topics_count":       len(summary.Topics),
		"action_items_count": len(summary.ActionItems),
		"decisions_count":    len(summary.Decisions),
		"code_changes_count": len(summary.CodeChanges),
		"context_keys_count": len(summary.Context),
		"tokens_used":        summary.TokensUsed,
		"created_at":         summary.CreatedAt,
	}
}
