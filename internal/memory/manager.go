package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"alex/internal/llm"
	"alex/internal/session"
)

// MemoryManager coordinates short-term and long-term memory systems
type MemoryManager struct {
	shortTerm  *ShortTermMemoryManager
	longTerm   *LongTermMemoryManager
	compressor *ContextCompressor
	controller *MemoryController
	llmClient  llm.Client
}

// NewMemoryManager creates a new unified memory manager
func NewMemoryManager(llmClient llm.Client) (*MemoryManager, error) {
	// Create storage directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	storageDir := filepath.Join(homeDir, ".deep-coding-memory")
	longTermDir := filepath.Join(storageDir, "long-term")

	// Initialize components
	shortTerm := NewShortTermMemoryManager(1000, 24*time.Hour) // 1000 items, 24h TTL
	longTerm, err := NewLongTermMemoryManager(longTermDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create long-term memory: %w", err)
	}

	compressor := NewContextCompressor(llmClient, nil)
	controller := NewMemoryController()

	return &MemoryManager{
		shortTerm:  shortTerm,
		longTerm:   longTerm,
		compressor: compressor,
		controller: controller,
		llmClient:  llmClient,
	}, nil
}

// Store stores a memory item in the appropriate memory system
func (mm *MemoryManager) Store(item *MemoryItem) error {
	// Use controller to determine storage location
	accessPattern := &AccessPattern{
		AccessCount:  item.AccessCount,
		LastAccess:   item.LastAccess,
		RecentAccess: time.Since(item.LastAccess) < 24*time.Hour,
	}
	
	if mm.controller.ShouldPromoteToLongTerm(item, accessPattern) || item.Type == LongTermMemory {
		// Store in long-term memory for important items
		item.Type = LongTermMemory
		return mm.longTerm.Store(item)
	}

	// Store in short-term memory for less important items
	item.Type = ShortTermMemory
	return mm.shortTerm.Store(item)
}

// Recall retrieves memories from both systems and merges results
func (mm *MemoryManager) Recall(query *MemoryQuery) *RecallResult {
	var allItems []*MemoryItem
	var allScores []float64

	// Query short-term memory
	if len(query.Types) == 0 || contains(query.Types, ShortTermMemory) {
		shortResult := mm.shortTerm.Recall(query)
		allItems = append(allItems, shortResult.Items...)
		allScores = append(allScores, shortResult.RelevanceScores...)
	}

	// Query long-term memory
	if len(query.Types) == 0 || contains(query.Types, LongTermMemory) {
		longResult := mm.longTerm.Recall(query)
		allItems = append(allItems, longResult.Items...)
		allScores = append(allScores, longResult.RelevanceScores...)
	}

	// Deduplicate and re-rank
	uniqueItems, uniqueScores := mm.deduplicateResults(allItems, allScores)

	// Apply final limit
	if query.Limit > 0 && len(uniqueItems) > query.Limit {
		uniqueItems = uniqueItems[:query.Limit]
		uniqueScores = uniqueScores[:query.Limit]
	}

	return &RecallResult{
		Items:           uniqueItems,
		TotalFound:      len(uniqueItems),
		RelevanceScores: uniqueScores,
	}
}

// ProcessContextCompression handles context compression for a session
func (mm *MemoryManager) ProcessContextCompression(ctx context.Context, sess *session.Session, maxTokens int) (*CompressionResult, error) {
	messages := sess.GetMessages()

	// Check if compression is needed
	if !mm.compressor.NeedsCompression(messages, maxTokens) {
		return &CompressionResult{
			OriginalCount:   len(messages),
			CompressedCount: len(messages),
		}, nil
	}

	// Perform compression
	result, err := mm.compressor.Compress(ctx, sess.ID, messages)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	// Store extracted memories
	for _, memory := range result.MemoryItems {
		if err := mm.Store(memory); err != nil {
			// Log error but don't fail the compression
			fmt.Printf("Warning: failed to store memory item: %v\n", err)
		}
	}

	// Update session with compressed messages
	if result.CompressedSummary != "" {
		// Clear existing messages
		sess.ClearMessages()

		// Add preserved messages
		for _, msg := range result.PreservedItems {
			sess.AddMessage(msg)
		}

		// Add compression summary as a system message
		summaryMsg := &session.Message{
			Role:    "system",
			Content: fmt.Sprintf("## Conversation Summary\n\n%s\n\n---\n\nThe above is a summary of the previous conversation. Continue from here with full context awareness.", result.CompressedSummary),
			Metadata: map[string]interface{}{
				"type":                "compression_summary",
				"original_count":      result.OriginalCount,
				"compression_ratio":   result.CompressionRatio,
				"tokens_saved":        result.TokensSaved,
				"compression_timestamp": time.Now().Unix(),
			},
			Timestamp: time.Now(),
		}
		sess.AddMessage(summaryMsg)
	}

	return result, nil
}

// CreateMemoryFromMessage intelligently creates memory items from a message
func (mm *MemoryManager) CreateMemoryFromMessage(ctx context.Context, sessionID string, msg *session.Message, sessionMessageCount int) ([]*MemoryItem, error) {
	// Get recent memory count for rate limiting
	recentMemoryCount := mm.getRecentMemoryCount(sessionID, time.Hour)
	
	// Check if we should create memory from this message
	if !mm.controller.ShouldCreateMemory(msg, sessionMessageCount, recentMemoryCount) {
		return nil, nil
	}
	
	// Classify the memory
	category, importance, tags := mm.controller.ClassifyMemory(msg)
	
	// Skip if importance is too low
	if importance < 0.3 {
		return nil, nil
	}
	
	// Create memory items based on content analysis
	var memories []*MemoryItem
	
	// Main content memory
	mainMemory := &MemoryItem{
		ID:         fmt.Sprintf("%s_%s_%d", category, sessionID, time.Now().UnixNano()),
		SessionID:  sessionID,
		Category:   category,
		Content:    mm.extractRelevantContent(msg),
		Importance: importance,
		Tags:       tags,
		CreatedAt:  msg.Timestamp,
		UpdatedAt:  msg.Timestamp,
		LastAccess: msg.Timestamp,
		Metadata: map[string]interface{}{
			"original_role": msg.Role,
			"message_length": len(msg.Content),
			"tool_calls_count": len(msg.ToolCalls),
		},
	}
	
	memories = append(memories, mainMemory)
	
	// Create additional specialized memories
	additionalMemories := mm.createSpecializedMemories(sessionID, msg, category)
	memories = append(memories, additionalMemories...)
	
	// Store all memories
	for _, memory := range memories {
		if err := mm.Store(memory); err != nil {
			fmt.Printf("Warning: failed to store memory %s: %v\n", memory.ID, err)
		}
	}
	
	return memories, nil
}

// AutomaticMemoryMaintenance performs periodic memory maintenance
func (mm *MemoryManager) AutomaticMemoryMaintenance(sessionID string) error {
	// Promote important short-term memories to long-term
	if err := mm.PromoteToLongTerm(sessionID, 0.7); err != nil {
		return fmt.Errorf("failed to promote memories: %w", err)
	}
	
	// Clean up old memories
	if err := mm.CleanupMemories(); err != nil {
		return fmt.Errorf("failed to cleanup memories: %w", err)
	}
	
	return nil
}

// MergeMemoriesToMessages retrieves relevant memories and formats them for inclusion in messages
func (mm *MemoryManager) MergeMemoriesToMessages(ctx context.Context, sessionID string, recentMessages []*session.Message, maxMemories int) ([]*session.Message, error) {
	// Analyze recent messages to determine what memories to recall
	query := mm.buildMemoryQuery(sessionID, recentMessages, maxMemories)

	// Recall relevant memories
	recallResult := mm.Recall(query)

	if len(recallResult.Items) == 0 {
		return recentMessages, nil
	}

	// Format memories as context message
	contextMsg := mm.formatMemoriesAsMessage(recallResult.Items)

	// Merge with recent messages
	var mergedMessages []*session.Message

	// Add context message at the beginning (after system messages)
	systemMsgCount := 0
	for _, msg := range recentMessages {
		if msg.Role == "system" {
			systemMsgCount++
		} else {
			break
		}
	}

	// Insert context after system messages
	mergedMessages = append(mergedMessages, recentMessages[:systemMsgCount]...)
	mergedMessages = append(mergedMessages, contextMsg)
	mergedMessages = append(mergedMessages, recentMessages[systemMsgCount:]...)

	return mergedMessages, nil
}

// GetMemoryStats returns comprehensive memory statistics
func (mm *MemoryManager) GetMemoryStats() map[string]interface{} {
	shortStats := mm.shortTerm.GetStats()
	longStats := mm.longTerm.GetStats()

	return map[string]interface{}{
		"short_term": shortStats,
		"long_term":  longStats,
		"total_items": shortStats.TotalItems + longStats.TotalItems,
		"total_size":  shortStats.TotalSize + longStats.TotalSize,
	}
}

// CleanupMemories performs maintenance on both memory systems
func (mm *MemoryManager) CleanupMemories() error {
	// Cleanup expired short-term memories (automatic in short-term manager)
	
	// Vacuum long-term memories if they exceed limits
	longStats := mm.longTerm.GetStats()
	if longStats.TotalItems > 5000 {
		if err := mm.longTerm.Vacuum(4000, 0.3); err != nil {
			return fmt.Errorf("failed to vacuum long-term memory: %w", err)
		}
	}

	return nil
}

// PromoteToLongTerm moves important short-term memories to long-term storage
func (mm *MemoryManager) PromoteToLongTerm(sessionID string, minImportance float64) error {
	memories := mm.shortTerm.GetSessionMemories(sessionID)

	for _, memory := range memories {
		if memory.Importance >= minImportance {
			// Update for long-term storage
			memory.Type = LongTermMemory
			memory.ExpiresAt = nil // Remove expiration

			// Store in long-term
			if err := mm.longTerm.Store(memory); err != nil {
				continue // Skip errors but continue processing
			}

			// Remove from short-term
			_ = mm.shortTerm.Delete(memory.ID)
		}
	}

	return nil
}

// Private helper methods

func (mm *MemoryManager) buildMemoryQuery(sessionID string, recentMessages []*session.Message, maxMemories int) *MemoryQuery {
	// Extract keywords and topics from recent messages
	var content []string
	var tags []string

	for _, msg := range recentMessages {
		content = append(content, msg.Content)
		
		// Extract potential tags from tool calls
		for _, toolCall := range msg.ToolCalls {
			tags = append(tags, toolCall.Name)
		}
	}

	// Combine content for search
	searchContent := strings.Join(content, " ")

	// Build query focusing on relevant categories
	return &MemoryQuery{
		SessionID: sessionID,
		Categories: []MemoryCategory{
			CodeContext,
			TaskHistory,
			Knowledge,
			Solutions,
		},
		Tags:          tags,
		Content:       mm.extractKeywords(searchContent),
		MinImportance: 0.5,
		Limit:         maxMemories,
		SortBy:        "importance",
	}
}

func (mm *MemoryManager) extractKeywords(content string) string {
	// Simple keyword extraction (could be enhanced with NLP)
	words := strings.Fields(strings.ToLower(content))
	
	var keywords []string
	for _, word := range words {
		if len(word) > 4 && !mm.isCommonWord(word) {
			keywords = append(keywords, word)
		}
	}

	// Return first few keywords as search term
	limit := 10
	if len(keywords) < limit {
		limit = len(keywords)
	}

	return strings.Join(keywords[:limit], " ")
}

func (mm *MemoryManager) isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"please": true, "could": true, "would": true, "should": true,
		"there": true, "where": true, "which": true, "while": true,
		"about": true, "after": true, "again": true, "against": true,
		"before": true, "being": true, "below": true, "between": true,
		"during": true, "further": true, "having": true, "through": true,
	}
	return commonWords[word]
}

func (mm *MemoryManager) formatMemoriesAsMessage(memories []*MemoryItem) *session.Message {
	var parts []string
	parts = append(parts, "## Relevant Context from Memory\n")

	// Group memories by category
	categoryGroups := make(map[MemoryCategory][]*MemoryItem)
	for _, memory := range memories {
		categoryGroups[memory.Category] = append(categoryGroups[memory.Category], memory)
	}

	// Format each category
	for category, items := range categoryGroups {
		parts = append(parts, fmt.Sprintf("### %s", strings.ToTitle(string(category))))
		
		for _, item := range items {
			parts = append(parts, fmt.Sprintf("- %s", item.Content))
		}
		parts = append(parts, "")
	}

	parts = append(parts, "---\n")

	return &session.Message{
		Role:    "system",
		Content: strings.Join(parts, "\n"),
		Metadata: map[string]interface{}{
			"type":           "memory_context",
			"memory_count":   len(memories),
			"categories":     len(categoryGroups),
			"recall_timestamp": time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}
}

func (mm *MemoryManager) deduplicateResults(items []*MemoryItem, scores []float64) ([]*MemoryItem, []float64) {
	seen := make(map[string]bool)
	var uniqueItems []*MemoryItem
	var uniqueScores []float64

	for i, item := range items {
		if !seen[item.ID] {
			seen[item.ID] = true
			uniqueItems = append(uniqueItems, item)
			if i < len(scores) {
				uniqueScores = append(uniqueScores, scores[i])
			}
		}
	}

	return uniqueItems, uniqueScores
}

func contains(slice []MemoryType, item MemoryType) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (mm *MemoryManager) getRecentMemoryCount(sessionID string, duration time.Duration) int {
	cutoff := time.Now().Add(-duration)
	
	memories := mm.shortTerm.GetSessionMemories(sessionID)
	count := 0
	
	for _, memory := range memories {
		if memory.CreatedAt.After(cutoff) {
			count++
		}
	}
	
	return count
}

func (mm *MemoryManager) extractRelevantContent(msg *session.Message) string {
	content := msg.Content
	
	// Limit content length for memory efficiency
	maxLength := 500
	if len(content) > maxLength {
		content = content[:maxLength] + "..."
	}
	
	// Add tool call information if present
	if len(msg.ToolCalls) > 0 {
		var toolInfo []string
		for _, tc := range msg.ToolCalls {
			toolInfo = append(toolInfo, fmt.Sprintf("%s()", tc.Name))
		}
		content += fmt.Sprintf(" [Tools: %s]", strings.Join(toolInfo, ", "))
	}
	
	return content
}

func (mm *MemoryManager) createSpecializedMemories(sessionID string, msg *session.Message, category MemoryCategory) []*MemoryItem {
	var memories []*MemoryItem
	
	// Create tool-specific memories
	for _, toolCall := range msg.ToolCalls {
		memory := &MemoryItem{
			ID:         fmt.Sprintf("tool_%s_%s_%d", toolCall.Name, sessionID, time.Now().UnixNano()),
			SessionID:  sessionID,
			Category:   TaskHistory,
			Content:    fmt.Sprintf("Used tool: %s with args: %s", toolCall.Name, mm.formatToolArgs(toolCall.Args)),
			Importance: 0.6,
			Tags:       []string{"tool", "execution", toolCall.Name},
			CreatedAt:  msg.Timestamp,
			UpdatedAt:  msg.Timestamp,
			LastAccess: msg.Timestamp,
			Metadata: map[string]interface{}{
				"tool_name": toolCall.Name,
				"tool_id":   toolCall.ID,
			},
		}
		memories = append(memories, memory)
	}
	
	// Create code-specific memories for code blocks
	if category == CodeContext {
		codeBlocks := mm.extractCodeBlocks(msg.Content)
		if len(codeBlocks) > 0 {
			memory := &MemoryItem{
				ID:         fmt.Sprintf("code_%s_%d", sessionID, time.Now().UnixNano()),
				SessionID:  sessionID,
				Category:   CodeContext,
				Content:    strings.Join(codeBlocks, "\n---\n"),
				Importance: 0.8,
				Tags:       []string{"code", "implementation"},
				CreatedAt:  msg.Timestamp,
				UpdatedAt:  msg.Timestamp,
				LastAccess: msg.Timestamp,
				Metadata: map[string]interface{}{
					"code_blocks_count": len(codeBlocks),
				},
			}
			memories = append(memories, memory)
		}
	}
	
	return memories
}

func (mm *MemoryManager) formatToolArgs(args map[string]interface{}) string {
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

func (mm *MemoryManager) extractCodeBlocks(content string) []string {
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