package utils

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"alex/internal/llm"
	"alex/internal/session"
)

// ContentAnalyzer provides unified content analysis
type ContentAnalyzer struct{}

// NewContentAnalyzer creates a new content analyzer
func NewContentAnalyzer() *ContentAnalyzer {
	return &ContentAnalyzer{}
}

// ExtractCodeBlocks extracts code blocks from text content
func (ca *ContentAnalyzer) ExtractCodeBlocks(content string) []string {
	var blocks []string
	lines := strings.Split(content, "\n")
	var currentBlock []string
	inBlock := false

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inBlock {
				if len(currentBlock) > 0 {
					blocks = append(blocks, strings.Join(currentBlock, "\n"))
				}
				currentBlock = nil
				inBlock = false
			} else {
				inBlock = true
			}
		} else if inBlock {
			currentBlock = append(currentBlock, line)
		}
	}

	// Handle unclosed code blocks
	if inBlock && len(currentBlock) > 0 {
		blocks = append(blocks, strings.Join(currentBlock, "\n"))
	}

	return blocks
}

// HasCodeContent checks if content contains code
func (ca *ContentAnalyzer) HasCodeContent(content string) bool {
	return strings.Contains(content, "```") ||
		strings.Contains(content, "func ") ||
		strings.Contains(content, "class ") ||
		strings.Contains(content, "def ") ||
		strings.Contains(content, "import ")
}

// ContainsError checks if content contains error indicators
func (ca *ContentAnalyzer) ContainsError(content string) bool {
	errorKeywords := []string{"error", "exception", "failed", "panic", "fatal"}
	lower := strings.ToLower(content)

	for _, keyword := range errorKeywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

// ExtractErrorInfo extracts error-related information from content
func (ca *ContentAnalyzer) ExtractErrorInfo(content string) string {
	lines := strings.Split(content, "\n")
	var errorLines []string

	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") ||
			strings.Contains(lower, "exception") ||
			strings.Contains(lower, "failed") {
			errorLines = append(errorLines, strings.TrimSpace(line))
		}
	}

	return strings.Join(errorLines, "\n")
}

// ConversationSummarizer provides unified conversation summarization
type ConversationSummarizer struct {
	llmClient llm.Client
}

// NewConversationSummarizer creates a new conversation summarizer
func NewConversationSummarizer(llmClient llm.Client) *ConversationSummarizer {
	return &ConversationSummarizer{
		llmClient: llmClient,
	}
}

// SummarizeMessages creates a structured summary of conversation messages
func (cs *ConversationSummarizer) SummarizeMessages(ctx context.Context, messages []*session.Message) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	conversationText := cs.formatMessages(messages)
	prompt := cs.buildSummarizationPrompt(conversationText, len(messages))

	req := &llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: "You are an expert conversation summarizer. Create concise but comprehensive summaries that preserve all important information.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ModelType:   llm.ReasoningModel,
		Temperature: 0.1,
		MaxTokens:   1500,
	}

	response, err := cs.llmClient.Chat(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no summary generated")
	}

	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}

// CreateCompactSummary creates a shorter summary for space-constrained contexts
func (cs *ConversationSummarizer) CreateCompactSummary(ctx context.Context, messages []*session.Message) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	conversationText := cs.formatMessages(messages)
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
		ModelType:   llm.BasicModel,
		Temperature: 0.1,
		MaxTokens:   150,
	}

	response, err := cs.llmClient.Chat(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate compact summary: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no summary generated")
	}

	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}

// Private helper methods

func (cs *ConversationSummarizer) formatMessages(messages []*session.Message) string {
	var parts []string

	for i, msg := range messages {
		// Skip system messages except summaries
		if msg.Role == "system" {
			if metadata, ok := msg.Metadata["type"]; ok &&
				!strings.Contains(fmt.Sprintf("%v", metadata), "summary") {
				continue
			}
		}

		timestamp := ""
		if !msg.Timestamp.IsZero() {
			timestamp = fmt.Sprintf(" (%s)", msg.Timestamp.Format("15:04:05"))
		}

		role := strings.ToUpper(msg.Role)
		content := strings.TrimSpace(msg.Content)

		// Limit message length
		if len(content) > 500 {
			content = content[:500] + "...[truncated]"
		}

		// Add tool call info
		toolInfo := ""
		if len(msg.ToolCalls) > 0 {
			var toolNames []string
			for _, tc := range msg.ToolCalls {
				toolNames = append(toolNames, tc.Name)
			}
			toolInfo = fmt.Sprintf(" [Tools: %s]", strings.Join(toolNames, ", "))
		}

		parts = append(parts, fmt.Sprintf("%d. %s%s%s:\n%s\n",
			i+1, role, timestamp, toolInfo, content))
	}

	return strings.Join(parts, "\n")
}

func (cs *ConversationSummarizer) buildSummarizationPrompt(conversationText string, messageCount int) string {
	return fmt.Sprintf(`Please analyze and summarize the following conversation (%d messages) in a structured format.

CONVERSATION:
%s

Provide a comprehensive summary that includes:
1. Main conversation flow and outcomes
2. Key technical details and code-related discussions  
3. Important decisions made or conclusions reached
4. Any unresolved issues or ongoing work
5. Current state and next steps

Focus on preserving essential information for conversation continuation. Be comprehensive but concise.`,
		messageCount, conversationText)
}

// FormatToolArgs formats tool arguments for display
func FormatToolArgs(args map[string]interface{}) string {
	if len(args) == 0 {
		return "{}"
	}

	data, err := json.Marshal(args)
	if err != nil {
		return "{...}"
	}

	// Limit length
	result := string(data)
	if len(result) > 200 {
		result = result[:200] + "..."
	}

	return result
}

// UnifiedContextConfig represents unified configuration for context management
type UnifiedContextConfig struct {
	// Token limits
	MaxTokens              int     `json:"max_tokens"`
	SummarizationThreshold int     `json:"summarization_threshold"`
	CompressionRatio       float64 `json:"compression_ratio"`

	// Message limits
	MaxMessages     int `json:"max_messages"`
	RecentKeepCount int `json:"recent_keep_count"`
	PreserveRecent  int `json:"preserve_recent"`

	// Importance thresholds
	MinImportance float64 `json:"min_importance"`

	// Feature flags
	PreserveSystemMessages bool `json:"preserve_system_messages"`
	EnableLLMCompress      bool `json:"enable_llm_compress"`
}

// NewUnifiedContextConfig creates a new unified context configuration with defaults
func NewUnifiedContextConfig() *UnifiedContextConfig {
	return &UnifiedContextConfig{
		MaxTokens:              60000,
		SummarizationThreshold: 45000,
		CompressionRatio:       0.3,
		MaxMessages:            20,
		RecentKeepCount:        6,
		PreserveRecent:         5,
		MinImportance:          0.5,
		PreserveSystemMessages: true,
		EnableLLMCompress:      true,
	}
}

// GenerateProjectID 基于当前工作目录生成项目ID
func GenerateProjectID() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(workingDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 使用MD5哈希生成项目ID
	hash := md5.Sum([]byte(absPath))
	projectID := fmt.Sprintf("project_%x", hash[:8]) // 使用前8个字节

	return projectID, nil
}

// GetProjectDisplayName 获取项目显示名称
func GetProjectDisplayName() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// 返回目录名称作为显示名称
	return filepath.Base(workingDir), nil
}
