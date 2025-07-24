package llm

import (
	"context"
	"fmt"
	"testing"
	"time"
)


// TestCacheManager_GetOrCreateCache tests cache creation and retrieval
func TestCacheManager_GetOrCreateCache(t *testing.T) {
	cm := NewCacheManager()
	sessionID := "test_session_1"

	// Test cache creation
	cache1 := cm.GetOrCreateCache(sessionID)
	if cache1 == nil {
		t.Fatal("Expected cache to be created, got nil")
	}
	if cache1.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, cache1.SessionID)
	}

	// Test cache retrieval (should get same instance)
	cache2 := cm.GetOrCreateCache(sessionID)
	if cache1 != cache2 {
		t.Error("Expected same cache instance, got different instances")
	}

	// Verify initial state
	if len(cache1.Messages) != 0 {
		t.Errorf("Expected empty messages, got %d messages", len(cache1.Messages))
	}
	if cache1.TokensUsed != 0 {
		t.Errorf("Expected 0 tokens used, got %d", cache1.TokensUsed)
	}
	if cache1.RequestCount != 0 {
		t.Errorf("Expected 0 request count, got %d", cache1.RequestCount)
	}
}

// TestCacheManager_UpdateCache tests cache updates
func TestCacheManager_UpdateCache(t *testing.T) {
	cm := NewCacheManager()
	sessionID := "test_session_2"

	// Create cache
	cache := cm.GetOrCreateCache(sessionID)

	// Test update with messages
	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}
	tokensUsed := 50

	cm.UpdateCache(sessionID, messages, tokensUsed)

	// Verify updates
	if len(cache.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(cache.Messages))
	}
	if cache.TokensUsed != tokensUsed {
		t.Errorf("Expected %d tokens used, got %d", tokensUsed, cache.TokensUsed)
	}
	if cache.RequestCount != 1 {
		t.Errorf("Expected 1 request count, got %d", cache.RequestCount)
	}
	if cache.CacheKey == "" {
		t.Error("Expected cache key to be generated")
	}

	// Test multiple updates
	moreMessages := []Message{
		{Role: "user", Content: "How are you?"},
		{Role: "assistant", Content: "I'm doing well!"},
	}
	cm.UpdateCache(sessionID, moreMessages, 30)

	if len(cache.Messages) != 4 {
		t.Errorf("Expected 4 messages after second update, got %d", len(cache.Messages))
	}
	if cache.TokensUsed != 80 {
		t.Errorf("Expected 80 tokens used after second update, got %d", cache.TokensUsed)
	}
	if cache.RequestCount != 2 {
		t.Errorf("Expected 2 request count after second update, got %d", cache.RequestCount)
	}
}

// TestCacheManager_GetOptimizedMessages tests message optimization
func TestCacheManager_GetOptimizedMessages(t *testing.T) {
	cm := NewCacheManager()
	sessionID := "test_session_3"

	// Test with no existing cache
	newMessages := []Message{
		{Role: "user", Content: "First message"},
	}
	optimized := cm.GetOptimizedMessages(sessionID, newMessages)
	if len(optimized) != 1 {
		t.Errorf("Expected 1 message with no cache, got %d", len(optimized))
	}

	// Create cache with several messages
	cm.GetOrCreateCache(sessionID)
	existingMessages := make([]Message, 10)
	for i := 0; i < 10; i++ {
		existingMessages[i] = Message{
			Role:    "user",
			Content: fmt.Sprintf("Message %d", i+1),
		}
	}
	cm.UpdateCache(sessionID, existingMessages, 100)

	// Test optimization with cached messages
	newMessages = []Message{
		{Role: "user", Content: "New message"},
	}
	optimized = cm.GetOptimizedMessages(sessionID, newMessages)

	// Should have: summary + recent messages + new messages
	// With 10 cached messages, should get summary + last 3 + new message
	expectedMin := 4 // summary + 3 recent + 1 new
	if len(optimized) < expectedMin {
		t.Errorf("Expected at least %d optimized messages, got %d", expectedMin, len(optimized))
	}

	// Verify that first message is summary
	if len(optimized) > 0 && optimized[0].Role != "system" {
		t.Error("Expected first optimized message to be system summary")
	}
}

// TestCacheManager_CompressMessages tests message compression
func TestCacheManager_CompressMessages(t *testing.T) {
	cm := NewCacheManager()
	cm.maxMessageCount = 10 // Set small limit for testing

	cache := &SessionCache{
		SessionID: "test_compression",
		Messages:  make([]Message, 15), // More than max
	}

	// Fill with test messages
	for i := 0; i < 15; i++ {
		cache.Messages[i] = Message{
			Role:    "user",
			Content: fmt.Sprintf("Message %d", i+1),
		}
	}

	cm.compressMessages(cache)

	// Should be compressed to maxMessageCount/2 + 1 (for summary)
	expectedMax := cm.maxMessageCount/2 + 1
	if len(cache.Messages) > expectedMax {
		t.Errorf("Expected at most %d messages after compression, got %d", expectedMax, len(cache.Messages))
	}

	// First message should be summary
	if len(cache.Messages) > 0 && cache.Messages[0].Role != "system" {
		t.Error("Expected first message after compression to be system summary")
	}
}

// TestCacheManager_CleanupExpired tests cleanup of expired caches
func TestCacheManager_CleanupExpired(t *testing.T) {
	cm := NewCacheManager()
	cm.cacheExpiry = 100 * time.Millisecond // Very short expiry for testing

	// Create some caches
	cm.GetOrCreateCache("session1")
	cache2 := cm.GetOrCreateCache("session2")

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Access one cache to keep it alive
	cache2.LastUsed = time.Now()

	// Create a new cache (this should trigger cleanup)
	cm.maxCacheSize = 1 // Force cleanup
	cm.GetOrCreateCache("session3")

	// Check that expired cache was removed
	cm.mutex.RLock()
	_, exists1 := cm.caches["session1"]
	_, exists3 := cm.caches["session3"]
	cm.mutex.RUnlock()

	if exists1 {
		t.Error("Expected expired cache to be removed")
	}
	if !exists3 {
		t.Error("Expected new cache to exist")
	}
	// Note: session2 might or might not exist depending on LRU cleanup
}

// TestCacheManager_GetCacheStats tests statistics collection
func TestCacheManager_GetCacheStats(t *testing.T) {
	cm := NewCacheManager()

	// Initially empty
	stats := cm.GetCacheStats()
	if stats["total_sessions"].(int) != 0 {
		t.Error("Expected 0 sessions initially")
	}

	// Add some caches with data
	sessionID1 := "stats_test_1"
	sessionID2 := "stats_test_2"

	cm.GetOrCreateCache(sessionID1)
	cm.UpdateCache(sessionID1, []Message{{Role: "user", Content: "Test"}}, 25)

	cm.GetOrCreateCache(sessionID2)
	cm.UpdateCache(sessionID2, []Message{{Role: "user", Content: "Test2"}}, 30)

	stats = cm.GetCacheStats()
	if stats["total_sessions"].(int) != 2 {
		t.Errorf("Expected 2 sessions, got %v", stats["total_sessions"])
	}
	if stats["total_cached_messages"].(int) != 2 {
		t.Errorf("Expected 2 messages, got %v", stats["total_cached_messages"])
	}
	if stats["total_tokens_saved"].(int) != 55 {
		t.Errorf("Expected 55 tokens, got %v", stats["total_tokens_saved"])
	}
}

// TestCacheManager_ClearCache tests cache clearing
func TestCacheManager_ClearCache(t *testing.T) {
	cm := NewCacheManager()
	sessionID := "clear_test"

	// Create and populate cache
	cm.GetOrCreateCache(sessionID)
	cm.UpdateCache(sessionID, []Message{{Role: "user", Content: "Test"}}, 10)

	// Verify cache exists
	cm.mutex.RLock()
	_, exists := cm.caches[sessionID]
	cm.mutex.RUnlock()
	if !exists {
		t.Fatal("Cache should exist before clearing")
	}

	// Clear cache
	cm.ClearCache(sessionID)

	// Verify cache is gone
	cm.mutex.RLock()
	_, exists = cm.caches[sessionID]
	cm.mutex.RUnlock()
	if exists {
		t.Error("Cache should not exist after clearing")
	}
}

// TestCacheManager_GenerateCacheKey tests cache key generation
func TestCacheManager_GenerateCacheKey(t *testing.T) {
	cm := NewCacheManager()

	messages1 := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
	}
	messages2 := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
	}
	messages3 := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hello"}, // Different content
	}

	key1 := cm.generateCacheKey(messages1)
	key2 := cm.generateCacheKey(messages2)
	key3 := cm.generateCacheKey(messages3)

	// Same messages should generate same key
	if key1 != key2 {
		t.Error("Same messages should generate same cache key")
	}

	// Different messages should generate different keys
	if key1 == key3 {
		t.Error("Different messages should generate different cache keys")
	}

	// Keys should be non-empty hex strings
	if len(key1) != 32 { // MD5 hex string length
		t.Errorf("Expected cache key length 32, got %d", len(key1))
	}
}

// TestHTTPLLMClient_ExtractSessionID tests session ID extraction
func TestHTTPLLMClient_ExtractSessionID(t *testing.T) {
	client, err := NewHTTPClient()
	if err != nil {
		t.Fatalf("Failed to create HTTP client: %v", err)
	}

	// Test extraction from context
	ctx := context.WithValue(context.Background(), ContextKeyType("sessionID"), "ctx_session_123")
	req := &ChatRequest{Messages: []Message{}}
	sessionID := client.extractSessionID(ctx, req)
	if sessionID != "ctx_session_123" {
		t.Errorf("Expected session ID from context 'ctx_session_123', got '%s'", sessionID)
	}

	// Test extraction from system message
	ctx = context.Background()
	req = &ChatRequest{
		Messages: []Message{
			{Role: "system", Content: "System prompt with session_id: msg_session_456 and other text"},
			{Role: "user", Content: "Hello"},
		},
	}
	sessionID = client.extractSessionID(ctx, req)
	if sessionID != "msg_session_456" {
		t.Errorf("Expected session ID from message 'msg_session_456', got '%s'", sessionID)
	}

	// Test no session ID found
	ctx = context.Background()
	req = &ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}
	sessionID = client.extractSessionID(ctx, req)
	if sessionID != "" {
		t.Errorf("Expected empty session ID, got '%s'", sessionID)
	}
}

// TestGlobalCacheManager tests singleton behavior
func TestGlobalCacheManager(t *testing.T) {
	cm1 := GetGlobalCacheManager()
	cm2 := GetGlobalCacheManager()

	if cm1 != cm2 {
		t.Error("Global cache manager should be singleton")
	}

	// Test that it works
	sessionID := "global_test"
	cache := cm1.GetOrCreateCache(sessionID)
	if cache == nil {
		t.Error("Expected cache to be created")
	}

	// Should be accessible from second reference
	cache2 := cm2.GetOrCreateCache(sessionID)
	if cache != cache2 {
		t.Error("Should get same cache instance from global manager")
	}
}

// Benchmark tests for performance
func BenchmarkCacheManager_GetOrCreateCache(b *testing.B) {
	cm := NewCacheManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessionID := fmt.Sprintf("bench_session_%d", i%100) // Reuse some sessions
		cm.GetOrCreateCache(sessionID)
	}
}

func BenchmarkCacheManager_UpdateCache(b *testing.B) {
	cm := NewCacheManager()
	sessionID := "bench_update"
	_ = cm.GetOrCreateCache(sessionID)

	messages := []Message{
		{Role: "user", Content: "Benchmark message"},
		{Role: "assistant", Content: "Benchmark response"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.UpdateCache(sessionID, messages, 50)
	}
}

func BenchmarkCacheManager_GetOptimizedMessages(b *testing.B) {
	cm := NewCacheManager()
	sessionID := "bench_optimize"

	// Setup cache with many messages
	cm.GetOrCreateCache(sessionID)
	largeMessageSet := make([]Message, 100)
	for i := 0; i < 100; i++ {
		largeMessageSet[i] = Message{
			Role:    "user",
			Content: fmt.Sprintf("Message %d with some content", i),
		}
	}
	cm.UpdateCache(sessionID, largeMessageSet, 1000)

	newMessages := []Message{
		{Role: "user", Content: "New benchmark message"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.GetOptimizedMessages(sessionID, newMessages)
	}
}
