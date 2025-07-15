package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"alex/internal/llm"
	"alex/internal/session"
)

// ContextCompressor handles intelligent compression of conversation context
type ContextCompressor struct {
	llmClient llm.Client
	config    *CompressionConfig
}

// CompressionResult represents the result of context compression
type CompressionResult struct {
	OriginalCount    int                    `json:"original_count"`
	CompressedCount  int                    `json:"compressed_count"`
	CompressionRatio float64               `json:"compression_ratio"`
	PreservedItems   []*session.Message    `json:"preserved_items"`
	CompressedSummary string               `json:"compressed_summary"`
	MemoryItems      []*MemoryItem         `json:"memory_items"`
	ProcessingTime   time.Duration         `json:"processing_time"`
	TokensSaved      int                   `json:"tokens_saved"`
}

// NewContextCompressor creates a new context compressor
func NewContextCompressor(llmClient llm.Client, config *CompressionConfig) *ContextCompressor {
	if config == nil {
		config = &CompressionConfig{
			Threshold:         0.8,  // 80% token usage
			CompressionRatio:  0.3,  // Compress to 30%
			PreserveRecent:    5,    // Keep 5 recent messages
			MinImportance:     0.5,  // Min importance to preserve
			EnableLLMCompress: true, // Use LLM compression
		}
	}

	return &ContextCompressor{
		llmClient: llmClient,
		config:    config,
	}
}

// NeedsCompression checks if context compression is needed
func (cc *ContextCompressor) NeedsCompression(messages []*session.Message, maxTokens int) bool {
	if len(messages) == 0 {
		return false
	}

	estimatedTokens := cc.estimateTokenUsage(messages)
	usageRatio := float64(estimatedTokens) / float64(maxTokens)
	
	return usageRatio >= cc.config.Threshold
}

// Compress performs intelligent context compression
func (cc *ContextCompressor) Compress(ctx context.Context, sessionID string, messages []*session.Message) (*CompressionResult, error) {
	start := time.Now()

	if len(messages) == 0 {
		return &CompressionResult{
			ProcessingTime: time.Since(start),
		}, nil
	}

	// Separate system and conversation messages
	systemMsgs, conversationMsgs := cc.separateMessages(messages)

	// Calculate how many messages to preserve
	preserveCount := cc.config.PreserveRecent
	if len(conversationMsgs) < preserveCount {
		preserveCount = len(conversationMsgs)
	}

	// Split into messages to compress and messages to preserve
	toCompress := conversationMsgs[:len(conversationMsgs)-preserveCount]
	toPreserve := conversationMsgs[len(conversationMsgs)-preserveCount:]

	var compressedSummary string
	var memoryItems []*MemoryItem
	var err error

	// Perform compression if we have messages to compress
	if len(toCompress) > 0 {
		if cc.config.EnableLLMCompress {
			compressedSummary, memoryItems, err = cc.llmCompress(ctx, sessionID, toCompress)
		} else {
			compressedSummary, memoryItems = cc.simpleCompress(sessionID, toCompress)
		}

		if err != nil {
			return nil, fmt.Errorf("compression failed: %w", err)
		}
	}

	// Build result
	result := &CompressionResult{
		OriginalCount:     len(messages),
		CompressedCount:   len(systemMsgs) + len(toPreserve) + 1, // +1 for summary
		CompressionRatio:  float64(len(systemMsgs)+len(toPreserve)+1) / float64(len(messages)),
		PreservedItems:    append(systemMsgs, toPreserve...),
		CompressedSummary: compressedSummary,
		MemoryItems:       memoryItems,
		ProcessingTime:    time.Since(start),
	}

	// Calculate tokens saved
	originalTokens := cc.estimateTokenUsage(messages)
	compressedTokens := cc.estimateTokenUsage(result.PreservedItems) + len(compressedSummary)/3
	result.TokensSaved = originalTokens - compressedTokens

	return result, nil
}

// ExtractMemories extracts important information as memory items
func (cc *ContextCompressor) ExtractMemories(sessionID string, messages []*session.Message) []*MemoryItem {
	var memories []*MemoryItem

	for _, msg := range messages {
		memories = append(memories, cc.extractMemoriesFromMessage(sessionID, msg)...)
	}

	// Filter and rank memories
	return cc.filterAndRankMemories(memories)
}

// Private helper methods

func (cc *ContextCompressor) llmCompress(ctx context.Context, sessionID string, messages []*session.Message) (string, []*MemoryItem, error) {
	// Format messages for LLM
	conversationText := cc.formatMessages(messages)

	// Create compression prompt
	prompt := cc.buildCompressionPrompt(conversationText, len(messages))

	// Call LLM for compression
	req := &llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: "You are an expert conversation compressor. Create concise but comprehensive summaries that preserve all important information, decisions, code changes, and context needed for continuation.",
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

	response, err := cc.llmClient.Chat(ctx, req)
	if err != nil {
		return "", nil, fmt.Errorf("LLM compression failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", nil, fmt.Errorf("no compression response")
	}

	// Parse structured response
	summary, memories, err := cc.parseCompressionResponse(sessionID, response.Choices[0].Message.Content)
	if err != nil {
		// Fallback to simple summary
		summary = response.Choices[0].Message.Content
		memories = cc.ExtractMemories(sessionID, messages)
	}

	return summary, memories, nil
}

func (cc *ContextCompressor) simpleCompress(sessionID string, messages []*session.Message) (string, []*MemoryItem) {
	var parts []string

	// Count different message types
	userMsgs := 0
	assistantMsgs := 0
	toolCalls := 0

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			userMsgs++
		case "assistant":
			assistantMsgs++
			if len(msg.ToolCalls) > 0 {
				toolCalls += len(msg.ToolCalls)
			}
		}
	}

	parts = append(parts, fmt.Sprintf("Compressed conversation summary (%d messages):", len(messages)))
	parts = append(parts, fmt.Sprintf("- User messages: %d", userMsgs))
	parts = append(parts, fmt.Sprintf("- Assistant messages: %d", assistantMsgs))
	parts = append(parts, fmt.Sprintf("- Tool calls executed: %d", toolCalls))

	// Extract key topics
	topics := cc.extractTopics(messages)
	if len(topics) > 0 {
		parts = append(parts, fmt.Sprintf("- Main topics: %s", strings.Join(topics, ", ")))
	}

	// Get memory items
	memories := cc.ExtractMemories(sessionID, messages)

	return strings.Join(parts, "\n"), memories
}

func (cc *ContextCompressor) extractMemoriesFromMessage(sessionID string, msg *session.Message) []*MemoryItem {
	var memories []*MemoryItem

	// Extract code-related memories
	if strings.Contains(msg.Content, "```") || strings.Contains(msg.Content, "code") {
		memory := &MemoryItem{
			ID:         fmt.Sprintf("code_%s_%d", sessionID, time.Now().UnixNano()),
			SessionID:  sessionID,
			Type:       ShortTermMemory,
			Category:   CodeContext,
			Content:    cc.extractCodeBlocks(msg.Content),
			Importance: 0.8,
			CreatedAt:  msg.Timestamp,
			Tags:       []string{"code", "development"},
		}
		if memory.Content != "" {
			memories = append(memories, memory)
		}
	}

	// Extract tool execution results
	if len(msg.ToolCalls) > 0 {
		for _, toolCall := range msg.ToolCalls {
			memory := &MemoryItem{
				ID:        fmt.Sprintf("tool_%s_%s_%d", sessionID, toolCall.Name, time.Now().UnixNano()),
				SessionID: sessionID,
				Type:      ShortTermMemory,
				Category:  TaskHistory,
				Content:   fmt.Sprintf("Tool: %s\nArgs: %s", toolCall.Name, cc.formatArgs(toolCall.Args)),
				Importance: 0.6,
				CreatedAt:  msg.Timestamp,
				Tags:      []string{"tool", "execution", toolCall.Name},
			}
			memories = append(memories, memory)
		}
	}

	// Extract error patterns
	if cc.containsError(msg.Content) {
		memory := &MemoryItem{
			ID:         fmt.Sprintf("error_%s_%d", sessionID, time.Now().UnixNano()),
			SessionID:  sessionID,
			Type:       LongTermMemory,
			Category:   ErrorPatterns,
			Content:    cc.extractErrorInfo(msg.Content),
			Importance: 0.9,
			CreatedAt:  msg.Timestamp,
			Tags:       []string{"error", "debugging"},
		}
		if memory.Content != "" {
			memories = append(memories, memory)
		}
	}

	return memories
}

func (cc *ContextCompressor) filterAndRankMemories(memories []*MemoryItem) []*MemoryItem {
	// Filter by minimum importance
	var filtered []*MemoryItem
	for _, memory := range memories {
		if memory.Importance >= cc.config.MinImportance {
			filtered = append(filtered, memory)
		}
	}

	// Sort by importance
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Importance > filtered[j].Importance
	})

	// Limit to reasonable number
	maxMemories := 20
	if len(filtered) > maxMemories {
		filtered = filtered[:maxMemories]
	}

	return filtered
}

func (cc *ContextCompressor) estimateTokenUsage(messages []*session.Message) int {
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Content) + 50 // 50 chars overhead
	}
	return totalChars / 3 // Rough estimation: 3 chars per token
}

func (cc *ContextCompressor) separateMessages(messages []*session.Message) ([]*session.Message, []*session.Message) {
	var systemMsgs []*session.Message
	var conversationMsgs []*session.Message

	for _, msg := range messages {
		if msg.Role == "system" {
			systemMsgs = append(systemMsgs, msg)
		} else {
			conversationMsgs = append(conversationMsgs, msg)
		}
	}

	return systemMsgs, conversationMsgs
}

func (cc *ContextCompressor) formatMessages(messages []*session.Message) string {
	var parts []string

	for i, msg := range messages {
		timestamp := ""
		if !msg.Timestamp.IsZero() {
			timestamp = fmt.Sprintf(" [%s]", msg.Timestamp.Format("15:04:05"))
		}

		toolInfo := ""
		if len(msg.ToolCalls) > 0 {
			var tools []string
			for _, tc := range msg.ToolCalls {
				tools = append(tools, tc.Name)
			}
			toolInfo = fmt.Sprintf(" (Tools: %s)", strings.Join(tools, ", "))
		}

		parts = append(parts, fmt.Sprintf("%d. %s%s%s:\n%s\n",
			i+1, strings.ToUpper(msg.Role), timestamp, toolInfo, strings.TrimSpace(msg.Content)))
	}

	return strings.Join(parts, "\n")
}

func (cc *ContextCompressor) buildCompressionPrompt(conversationText string, messageCount int) string {
	return fmt.Sprintf(`Please compress this conversation (%d messages) into a structured summary that preserves all essential information for conversation continuation.

CONVERSATION:
%s

Provide a JSON response with this structure:
{
  "summary": "Comprehensive summary preserving key context, decisions, and current state",
  "key_points": ["Important points and findings"],
  "code_changes": ["Any code modifications or technical details"],
  "decisions": ["Decisions made or conclusions reached"],
  "next_steps": ["Unfinished tasks or next actions"],
  "context": {"key": "important context for continuation"}
}

Focus on:
1. Technical details and code-related discussions
2. Problem-solving progress and solutions
3. User preferences and patterns
4. Unresolved issues requiring follow-up
5. Current state and immediate context

Be comprehensive but concise.`, messageCount, conversationText)
}

func (cc *ContextCompressor) parseCompressionResponse(sessionID string, content string) (string, []*MemoryItem, error) {
	// Extract JSON from response
	content = strings.TrimSpace(content)
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart == -1 || jsonEnd == -1 {
		return content, nil, fmt.Errorf("no JSON found")
	}

	jsonContent := content[jsonStart : jsonEnd+1]

	var response struct {
		Summary     string            `json:"summary"`
		KeyPoints   []string          `json:"key_points"`
		CodeChanges []string          `json:"code_changes"`
		Decisions   []string          `json:"decisions"`
		NextSteps   []string          `json:"next_steps"`
		Context     map[string]string `json:"context"`
	}

	if err := json.Unmarshal([]byte(jsonContent), &response); err != nil {
		return content, nil, err
	}

	// Create memory items from structured data
	var memories []*MemoryItem
	now := time.Now()

	// Key points as knowledge
	for _, point := range response.KeyPoints {
		memories = append(memories, &MemoryItem{
			ID:         fmt.Sprintf("knowledge_%s_%d", sessionID, now.UnixNano()),
			SessionID:  sessionID,
			Type:       LongTermMemory,
			Category:   Knowledge,
			Content:    point,
			Importance: 0.7,
			CreatedAt:  now,
			Tags:       []string{"knowledge", "key_point"},
		})
	}

	// Code changes
	for _, change := range response.CodeChanges {
		memories = append(memories, &MemoryItem{
			ID:         fmt.Sprintf("code_%s_%d", sessionID, now.UnixNano()),
			SessionID:  sessionID,
			Type:       LongTermMemory,
			Category:   CodeContext,
			Content:    change,
			Importance: 0.8,
			CreatedAt:  now,
			Tags:       []string{"code", "change"},
		})
	}

	// Decisions
	for _, decision := range response.Decisions {
		memories = append(memories, &MemoryItem{
			ID:         fmt.Sprintf("decision_%s_%d", sessionID, now.UnixNano()),
			SessionID:  sessionID,
			Type:       LongTermMemory,
			Category:   Solutions,
			Content:    decision,
			Importance: 0.9,
			CreatedAt:  now,
			Tags:       []string{"decision", "solution"},
		})
	}

	return response.Summary, memories, nil
}

func (cc *ContextCompressor) extractCodeBlocks(content string) string {
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

	return strings.Join(blocks, "\n---\n")
}

func (cc *ContextCompressor) extractTopics(messages []*session.Message) []string {
	wordCount := make(map[string]int)
	
	for _, msg := range messages {
		words := strings.Fields(strings.ToLower(msg.Content))
		for _, word := range words {
			if len(word) > 4 && !cc.isStopWord(word) {
				wordCount[word]++
			}
		}
	}

	// Get top topics
	type wordFreq struct {
		word  string
		count int
	}

	var freqs []wordFreq
	for word, count := range wordCount {
		if count > 1 {
			freqs = append(freqs, wordFreq{word, count})
		}
	}

	sort.Slice(freqs, func(i, j int) bool {
		return freqs[i].count > freqs[j].count
	})

	var topics []string
	limit := 5
	if len(freqs) < limit {
		limit = len(freqs)
	}

	for i := 0; i < limit; i++ {
		topics = append(topics, freqs[i].word)
	}

	return topics
}

func (cc *ContextCompressor) isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "and": true, "for": true, "are": true,
		"but": true, "not": true, "you": true, "all": true,
		"can": true, "had": true, "her": true, "was": true,
		"one": true, "our": true, "out": true, "day": true,
		"get": true, "has": true, "him": true, "his": true,
		"how": true, "its": true, "may": true, "new": true,
		"now": true, "old": true, "see": true, "two": true,
		"who": true, "boy": true, "did": true, "man": true,
		"way": true, "too": true, "any": true, "she": true,
	}
	return stopWords[word]
}

func (cc *ContextCompressor) formatArgs(args map[string]interface{}) string {
	data, _ := json.Marshal(args)
	return string(data)
}

func (cc *ContextCompressor) containsError(content string) bool {
	errorKeywords := []string{"error", "exception", "failed", "panic", "fatal"}
	lower := strings.ToLower(content)
	
	for _, keyword := range errorKeywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func (cc *ContextCompressor) extractErrorInfo(content string) string {
	lines := strings.Split(content, "\n")
	var errorLines []string
	
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") || strings.Contains(lower, "exception") || strings.Contains(lower, "failed") {
			errorLines = append(errorLines, strings.TrimSpace(line))
		}
	}
	
	return strings.Join(errorLines, "\n")
}