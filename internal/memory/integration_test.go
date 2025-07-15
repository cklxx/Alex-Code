package memory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"alex/internal/llm"
	"alex/internal/session"
)

// MockLLMClient for testing
type MockLLMClient struct{}

func (m *MockLLMClient) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	// Return mock compression summary
	mockSummary := `{
		"summary": "User discussed implementing a memory system with context compression. Code changes were made to handle short-term and long-term memory storage.",
		"key_points": [
			"Implemented memory management system",
			"Added context compression functionality", 
			"Created intelligent memory classification"
		],
		"code_changes": [
			"Created memory package with types, managers, and controllers",
			"Added compression algorithms for context management"
		],
		"decisions": [
			"Use dual memory system with short-term and long-term storage",
			"Implement intelligent classification based on content analysis"
		],
		"next_steps": [
			"Integrate memory system with React agent",
			"Perform comprehensive testing"
		],
		"context": {
			"project": "Alex AI coding assistant",
			"focus": "Memory and context management"
		}
	}`

	return &llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.Message{
					Role:    "assistant",
					Content: mockSummary,
				},
			},
		},
		Usage: llm.Usage{
			PromptTokens:     100,
			CompletionTokens: 150,
			TotalTokens:      250,
		},
	}, nil
}

func (m *MockLLMClient) ChatStream(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamDelta, error) {
	ch := make(chan llm.StreamDelta, 1)
	close(ch)
	return ch, nil
}

func (m *MockLLMClient) Close() error {
	return nil
}

// Test complete memory system integration
func TestMemorySystemIntegration(t *testing.T) {
	// Setup test environment
	testDir := filepath.Join(os.TempDir(), "alex-memory-test", fmt.Sprintf("%d", time.Now().UnixNano()))
	defer func() { _ = os.RemoveAll(testDir) }()

	// Create mock LLM client
	mockLLM := &MockLLMClient{}

	// Create memory manager
	manager, err := createTestMemoryManager(mockLLM, testDir)
	if err != nil {
		t.Fatalf("Failed to create memory manager: %v", err)
	}

	// Test session creation
	sessionID := "test_session_001"
	sess := createTestSession(sessionID)

	t.Run("Memory Creation and Classification", func(t *testing.T) {
		testMemoryCreationAndClassification(t, manager, sess)
	})

	t.Run("Context Compression", func(t *testing.T) {
		testContextCompression(t, manager, sess)
	})

	t.Run("Memory Recall and Filtering", func(t *testing.T) {
		testMemoryRecallAndFiltering(t, manager, sess)
	})

	t.Run("Memory Promotion", func(t *testing.T) {
		testMemoryPromotion(t, manager, sess)
	})

	t.Run("Message Merging", func(t *testing.T) {
		testMessageMerging(t, manager, sess)
	})

	t.Run("Automatic Maintenance", func(t *testing.T) {
		testAutomaticMaintenance(t, manager, sess)
	})
}

func createTestMemoryManager(llmClient llm.Client, testDir string) (*MemoryManager, error) {
	// Create storage directories
	longTermDir := filepath.Join(testDir, "long-term")
	if err := os.MkdirAll(longTermDir, 0755); err != nil {
		return nil, err
	}

	// Create components
	shortTerm := NewShortTermMemoryManager(100, time.Hour)
	longTerm, err := NewLongTermMemoryManager(longTermDir)
	if err != nil {
		return nil, err
	}

	compressor := NewContextCompressor(llmClient, &CompressionConfig{
		Threshold:         0.8,
		CompressionRatio:  0.3,
		PreserveRecent:    3,
		MinImportance:     0.5,
		EnableLLMCompress: true,
	})

	controller := NewMemoryController()

	return &MemoryManager{
		shortTerm:  shortTerm,
		longTerm:   longTerm,
		compressor: compressor,
		controller: controller,
		llmClient:  llmClient,
	}, nil
}

func createTestSession(sessionID string) *session.Session {
	return &session.Session{
		ID:         sessionID,
		Created:    time.Now(),
		Updated:    time.Now(),
		Messages:   []*session.Message{},
		WorkingDir: "/test/dir",
	}
}

func testMemoryCreationAndClassification(t *testing.T, manager *MemoryManager, sess *session.Session) {
	ctx := context.Background()

	// Test various message types
	testMessages := []*session.Message{
		{
			Role:      "user",
			Content:   "Please help me implement a function to parse JSON data",
			Timestamp: time.Now(),
		},
		{
			Role:    "assistant",
			Content: "I'll help you create a JSON parser. Here's the implementation:\n```go\nfunc parseJSON(data []byte) (map[string]interface{}, error) {\n    var result map[string]interface{}\n    err := json.Unmarshal(data, &result)\n    return result, err\n}\n```",
			ToolCalls: []session.ToolCall{
				{ID: "call_1", Name: "file_write", Args: map[string]interface{}{"path": "parser.go", "content": "..."}},
			},
			Timestamp: time.Now(),
		},
		{
			Role:      "user",
			Content:   "Error: undefined variable 'json' in the code",
			Timestamp: time.Now(),
		},
		{
			Role:      "assistant",
			Content:   "I need to add the import statement. Here's the fix:\n```go\nimport \"encoding/json\"\n```",
			Timestamp: time.Now(),
		},
	}

	// Add messages to session and create memories
	for i, msg := range testMessages {
		sess.AddMessage(msg)

		memories, err := manager.CreateMemoryFromMessage(ctx, sess.ID, msg, i+1)
		if err != nil {
			t.Errorf("Failed to create memory from message %d: %v", i, err)
			continue
		}

		if i < 2 { // Skip memory creation for first 2 messages due to min count threshold
			if len(memories) > 0 {
				t.Errorf("Expected no memories for message %d due to threshold, got %d", i, len(memories))
			}
			continue
		}

		// Validate memory creation for later messages
		if len(memories) == 0 {
			t.Errorf("Expected memories for message %d, got none", i)
			continue
		}

		// Check memory classification
		for _, memory := range memories {
			if memory.SessionID != sess.ID {
				t.Errorf("Memory session ID mismatch: expected %s, got %s", sess.ID, memory.SessionID)
			}

			if memory.Importance < 0.3 {
				t.Errorf("Memory importance too low: %f", memory.Importance)
			}

			// Validate category assignment
			switch msg.Content {
			case testMessages[1].Content: // Code message
				if memory.Category != CodeContext && memory.Category != TaskHistory {
					t.Errorf("Expected CodeContext or TaskHistory for code message, got %s", memory.Category)
				}
			case testMessages[2].Content: // Error message
				if memory.Category != ErrorPatterns {
					t.Errorf("Expected ErrorPatterns for error message, got %s", memory.Category)
				}
			}
		}
	}

	// Verify memory storage
	stats := manager.GetMemoryStats()
	totalItems := stats["total_items"].(int)
	if totalItems == 0 {
		t.Error("No memories were stored")
	}

	t.Logf("Created %d memory items", totalItems)
}

func testContextCompression(t *testing.T, manager *MemoryManager, sess *session.Session) {
	ctx := context.Background()

	// Add many messages to trigger compression
	for i := 0; i < 20; i++ {
		msg := &session.Message{
			Role:      "user",
			Content:   fmt.Sprintf("This is message %d with some content to fill up the context", i),
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
		sess.AddMessage(msg)
	}

	// Test compression
	result, err := manager.ProcessContextCompression(ctx, sess, 1000) // Low token limit to force compression
	if err != nil {
		t.Fatalf("Context compression failed: %v", err)
	}

	if result.CompressionRatio >= 1.0 {
		t.Error("Expected compression ratio < 1.0")
	}

	if result.CompressedSummary == "" {
		t.Error("Expected compressed summary")
	}

	if len(result.MemoryItems) == 0 {
		t.Error("Expected memory items from compression")
	}

	// Verify session was updated with compressed content
	messages := sess.GetMessages()
	foundSummary := false
	for _, msg := range messages {
		if msg.Role == "system" && strings.Contains(msg.Content, "Conversation Summary") {
			foundSummary = true
			break
		}
	}

	if !foundSummary {
		t.Error("Expected compression summary in session messages")
	}

	t.Logf("Compression ratio: %.2f, Summary length: %d, Memory items: %d",
		result.CompressionRatio, len(result.CompressedSummary), len(result.MemoryItems))
}

func testMemoryRecallAndFiltering(t *testing.T, manager *MemoryManager, sess *session.Session) {
	// Test memory recall with different queries
	testQueries := []*MemoryQuery{
		{
			SessionID:  sess.ID,
			Categories: []MemoryCategory{CodeContext},
			Limit:      5,
		},
		{
			SessionID:     sess.ID,
			Content:       "json error",
			MinImportance: 0.5,
			Limit:         3,
		},
		{
			SessionID: sess.ID,
			Tags:      []string{"code", "error"},
			SortBy:    "importance",
			Limit:     10,
		},
	}

	for i, query := range testQueries {
		result := manager.Recall(query)

		if result.TotalFound < 0 {
			t.Errorf("Query %d: Invalid total found: %d", i, result.TotalFound)
		}

		// Verify filtering
		for _, item := range result.Items {
			if item.SessionID != sess.ID {
				t.Errorf("Query %d: Session ID mismatch in result", i)
			}

			if item.Importance < query.MinImportance {
				t.Errorf("Query %d: Item importance %f below threshold %f",
					i, item.Importance, query.MinImportance)
			}
		}

		t.Logf("Query %d returned %d items", i, len(result.Items))
	}
}

func testMemoryPromotion(t *testing.T, manager *MemoryManager, sess *session.Session) {
	// Create high-importance memories
	highImportanceMemory := &MemoryItem{
		ID:          "high_importance_test",
		SessionID:   sess.ID,
		Category:    Solutions,
		Content:     "Critical solution for parsing errors",
		Importance:  0.9,
		AccessCount: 5,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		LastAccess:  time.Now(),
		Tags:        []string{"solution", "critical"},
	}

	err := manager.Store(highImportanceMemory)
	if err != nil {
		t.Fatalf("Failed to store high importance memory: %v", err)
	}

	// Check if it was promoted to long-term
	longTermStats := manager.longTerm.GetStats()
	if longTermStats.TotalItems == 0 {
		t.Error("Expected high importance memory to be promoted to long-term storage")
	}

	// Test promotion of frequently accessed memories
	err = manager.PromoteToLongTerm(sess.ID, 0.7)
	if err != nil {
		t.Errorf("Memory promotion failed: %v", err)
	}

	t.Logf("Long-term storage has %d items after promotion", longTermStats.TotalItems)
}

func testMessageMerging(t *testing.T, manager *MemoryManager, sess *session.Session) {
	ctx := context.Background()

	// Create test messages
	recentMessages := []*session.Message{
		{
			Role:      "user",
			Content:   "I need to implement error handling for JSON parsing",
			Timestamp: time.Now(),
		},
		{
			Role:      "assistant",
			Content:   "Let me help you with that",
			Timestamp: time.Now(),
		},
	}

	// Test memory merging
	mergedMessages, err := manager.MergeMemoriesToMessages(ctx, sess.ID, recentMessages, 5)
	if err != nil {
		t.Fatalf("Message merging failed: %v", err)
	}

	// Verify merged messages contain memory context
	foundMemoryContext := false
	for _, msg := range mergedMessages {
		if msg.Role == "system" && strings.Contains(msg.Content, "Relevant Context from Memory") {
			foundMemoryContext = true
			break
		}
	}

	if len(mergedMessages) <= len(recentMessages) {
		t.Log("No memory context was added (may be expected if no relevant memories)")
	} else if !foundMemoryContext {
		t.Error("Expected memory context in merged messages")
	}

	t.Logf("Merged %d recent messages with memory context, total: %d",
		len(recentMessages), len(mergedMessages))
}

func testAutomaticMaintenance(t *testing.T, manager *MemoryManager, sess *session.Session) {
	// Test automatic memory maintenance
	err := manager.AutomaticMemoryMaintenance(sess.ID)
	if err != nil {
		t.Errorf("Automatic maintenance failed: %v", err)
	}

	// Verify cleanup occurred
	shortStats := manager.shortTerm.GetStats()
	longStats := manager.longTerm.GetStats()

	t.Logf("After maintenance - Short-term: %d items, Long-term: %d items",
		shortStats.TotalItems, longStats.TotalItems)

	// Test memory statistics
	overallStats := manager.GetMemoryStats()
	if overallStats["total_items"] == nil {
		t.Error("Expected total_items in memory stats")
	}

	t.Logf("Overall memory stats: %+v", overallStats)
}

// Benchmark memory operations
func BenchmarkMemoryOperations(b *testing.B) {
	testDir := filepath.Join(os.TempDir(), "alex-memory-bench")
	defer func() { _ = os.RemoveAll(testDir) }()

	mockLLM := &MockLLMClient{}
	manager, err := createTestMemoryManager(mockLLM, testDir)
	if err != nil {
		b.Fatalf("Failed to create memory manager: %v", err)
	}

	sessionID := "bench_session"
	_ = createTestSession(sessionID)

	b.Run("Memory Creation", func(b *testing.B) {
		ctx := context.Background()
		for i := 0; i < b.N; i++ {
			msg := &session.Message{
				Role:      "user",
				Content:   fmt.Sprintf("Benchmark message %d with code: func test() { return %d }", i, i),
				Timestamp: time.Now(),
			}

			_, err := manager.CreateMemoryFromMessage(ctx, sessionID, msg, i+10) // +10 to pass threshold
			if err != nil {
				b.Errorf("Memory creation failed: %v", err)
			}
		}
	})

	b.Run("Memory Recall", func(b *testing.B) {
		query := &MemoryQuery{
			SessionID: sessionID,
			Content:   "benchmark test function",
			Limit:     10,
		}

		for i := 0; i < b.N; i++ {
			result := manager.Recall(query)
			if result.TotalFound < 0 {
				b.Errorf("Invalid recall result")
			}
		}
	})
}

// Test memory system edge cases
func TestMemoryEdgeCases(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "alex-memory-edge-test")
	defer func() { _ = os.RemoveAll(testDir) }()

	mockLLM := &MockLLMClient{}
	manager, err := createTestMemoryManager(mockLLM, testDir)
	if err != nil {
		t.Fatalf("Failed to create memory manager: %v", err)
	}

	sessionID := "edge_test_session"

	t.Run("Empty Content Handling", func(t *testing.T) {
		msg := &session.Message{
			Role:      "user",
			Content:   "",
			Timestamp: time.Now(),
		}

		memories, err := manager.CreateMemoryFromMessage(context.Background(), sessionID, msg, 5)
		if err != nil {
			t.Errorf("Failed to handle empty content: %v", err)
		}

		if len(memories) > 0 {
			t.Error("Expected no memories for empty content")
		}
	})

	t.Run("Invalid Session Handling", func(t *testing.T) {
		query := &MemoryQuery{
			SessionID: "nonexistent_session",
			Limit:     5,
		}

		result := manager.Recall(query)
		if result.TotalFound != 0 {
			t.Error("Expected no results for nonexistent session")
		}
	})

	t.Run("Large Content Handling", func(t *testing.T) {
		largeContent := strings.Repeat("This is a very long message content. ", 1000)
		msg := &session.Message{
			Role:      "user",
			Content:   largeContent,
			Timestamp: time.Now(),
		}

		memories, err := manager.CreateMemoryFromMessage(context.Background(), sessionID, msg, 5)
		if err != nil {
			t.Errorf("Failed to handle large content: %v", err)
		}

		for _, memory := range memories {
			if len(memory.Content) > 600 { // Should be truncated
				t.Error("Expected content to be truncated for large messages")
			}
		}
	})
}
