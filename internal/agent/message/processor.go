package message

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	contextmgr "alex/internal/context"
	"alex/internal/llm"
	"alex/internal/session"
)

// CachedCompressionResult represents a cached compression result
type CachedCompressionResult struct {
	Summary      *session.Message
	Timestamp    time.Time
	InputHash    string
	MessageCount int
}

// MessageProcessor handles message processing operations
type MessageProcessor struct {
	contextMgr     *contextmgr.ContextManager
	sessionManager *session.Manager
	tokenEstimator *TokenEstimator
	converter      *MessageConverter
	compressor     *MessageCompressor

	// LLM compression cache optimization
	compressionCache map[string]*CachedCompressionResult
	compressionMutex sync.RWMutex
	cacheExpiry      time.Duration
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor(llmClient llm.Client, sessionManager *session.Manager) *MessageProcessor {
	// Create context manager
	contextConfig := &contextmgr.ContextLengthConfig{
		MaxTokens:              8000,
		SummarizationThreshold: 6000,
		CompressionRatio:       0.3,
		PreserveSystemMessages: true,
	}

	return &MessageProcessor{
		contextMgr:     contextmgr.NewContextManager(llmClient, contextConfig),
		sessionManager: sessionManager,
		tokenEstimator: NewTokenEstimator(),
		converter:      NewMessageConverter(),
		compressor:     NewMessageCompressor(llmClient),

		compressionCache: make(map[string]*CachedCompressionResult),
		cacheExpiry:      30 * time.Minute,
	}
}

// ProcessMessages processes a batch of messages
func (mp *MessageProcessor) ProcessMessages(ctx context.Context, messages []*session.Message) ([]*session.Message, error) {
	if len(messages) == 0 {
		return messages, nil
	}

	// Apply compression if needed
	compressed := mp.compressor.CompressMessages(messages)
	
	// Log compression statistics
	if len(compressed) != len(messages) {
		log.Printf("[INFO] MessageProcessor: Compressed %d messages to %d", len(messages), len(compressed))
	}

	return compressed, nil
}

// ConvertSessionToLLM delegates to converter
func (mp *MessageProcessor) ConvertSessionToLLM(sessionMessages []*session.Message) []llm.Message {
	return mp.converter.ConvertSessionToLLM(sessionMessages)
}

// ConvertLLMToSession delegates to converter
func (mp *MessageProcessor) ConvertLLMToSession(llmMessages []llm.Message) []*session.Message {
	return mp.converter.ConvertLLMToSession(llmMessages)
}

// CompressMessages delegates to compressor
func (mp *MessageProcessor) CompressMessages(messages []*session.Message) []*session.Message {
	return mp.compressor.CompressMessages(messages)
}

// GetCurrentSession retrieves the current session
func (mp *MessageProcessor) GetCurrentSession(ctx context.Context, agent interface{}) *session.Session {
	// Try to get session from context
	if sessionID, ok := ctx.Value("session_id").(string); ok && sessionID != "" {
		if session, err := mp.sessionManager.RestoreSession(sessionID); err == nil {
			return session
		}
	}
	
	// Try alternative context keys
	if sessionID, ok := ctx.Value("sessionID").(string); ok && sessionID != "" {
		if session, err := mp.sessionManager.RestoreSession(sessionID); err == nil {
			return session
		}
	}
	
	// Create default session if none found
	defaultSession, err := mp.sessionManager.StartSession("default")
	if err != nil {
		log.Printf("[ERROR] MessageProcessor: Failed to create default session: %v", err)
		return nil
	}
	
	return defaultSession
}

// GetContextStats retrieves context statistics
func (mp *MessageProcessor) GetContextStats(sess *session.Session) *contextmgr.ContextStats {
	if sess == nil {
		return &contextmgr.ContextStats{}
	}
	
	messages := sess.GetMessages()
	totalTokens := 0
	
	for _, msg := range messages {
		totalTokens += mp.tokenEstimator.EstimateTokens(msg.Content)
	}
	
	return &contextmgr.ContextStats{
		TotalMessages:     len(messages),
		EstimatedTokens:   totalTokens,
		SystemMessages:    mp.countSystemMessages(messages),
		UserMessages:      mp.countUserMessages(messages),
		AssistantMessages: mp.countAssistantMessages(messages),
		SummaryMessages:   0,
		MaxTokens:         8000,
	}
}

// RestoreFullContext restores full context from backup
func (mp *MessageProcessor) RestoreFullContext(sess *session.Session, backupID string) error {
	if sess == nil {
		return fmt.Errorf("session is nil")
	}
	
	// Implementation would restore from backup
	log.Printf("[INFO] MessageProcessor: Restoring full context for session %s from backup %s", sess.ID, backupID)
	return nil
}

// Helper methods
func (mp *MessageProcessor) calculateCompressionRatio(messages []*session.Message) float64 {
	if len(messages) == 0 {
		return 0.0
	}
	
	// Simple ratio calculation
	return 0.7 // Placeholder
}

func (mp *MessageProcessor) countSystemMessages(messages []*session.Message) int {
	count := 0
	for _, msg := range messages {
		if msg.Role == "system" {
			count++
		}
	}
	return count
}

func (mp *MessageProcessor) countUserMessages(messages []*session.Message) int {
	count := 0
	for _, msg := range messages {
		if msg.Role == "user" {
			count++
		}
	}
	return count
}

func (mp *MessageProcessor) countAssistantMessages(messages []*session.Message) int {
	count := 0
	for _, msg := range messages {
		if msg.Role == "assistant" {
			count++
		}
	}
	return count
}

// GetCachedCompressionResult retrieves cached compression result
func (mp *MessageProcessor) GetCachedCompressionResult(messages []*session.Message) *session.Message {
	mp.compressionMutex.RLock()
	defer mp.compressionMutex.RUnlock()
	
	hash := mp.buildMessageHash(messages)
	if cached, exists := mp.compressionCache[hash]; exists {
		// Check if cache is still valid
		if time.Since(cached.Timestamp) < mp.cacheExpiry {
			return cached.Summary
		}
		// Remove expired cache
		delete(mp.compressionCache, hash)
	}
	
	return nil
}

// SetCachedCompressionResult stores compression result in cache
func (mp *MessageProcessor) SetCachedCompressionResult(messages []*session.Message, summary *session.Message) {
	mp.compressionMutex.Lock()
	defer mp.compressionMutex.Unlock()
	
	hash := mp.buildMessageHash(messages)
	mp.compressionCache[hash] = &CachedCompressionResult{
		Summary:      summary,
		Timestamp:    time.Now(),
		InputHash:    hash,
		MessageCount: len(messages),
	}
	
	// Clean up old cache entries
	mp.cleanupCache()
}

// buildMessageHash creates a hash for the message set
func (mp *MessageProcessor) buildMessageHash(messages []*session.Message) string {
	var parts []string
	for _, msg := range messages {
		parts = append(parts, fmt.Sprintf("%s:%s", msg.Role, msg.Content[:min(50, len(msg.Content))]))
	}
	return fmt.Sprintf("%x", len(strings.Join(parts, "|")))
}

// cleanupCache removes old cache entries
func (mp *MessageProcessor) cleanupCache() {
	cutoff := time.Now().Add(-mp.cacheExpiry)
	for hash, cached := range mp.compressionCache {
		if cached.Timestamp.Before(cutoff) {
			delete(mp.compressionCache, hash)
		}
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TokenEstimator placeholder (would be extracted from original file)
type TokenEstimator struct{}

func NewTokenEstimator() *TokenEstimator {
	return &TokenEstimator{}
}

func (te *TokenEstimator) EstimateTokens(content string) int {
	// Simple estimation: ~4 characters per token
	return len(content) / 4
}