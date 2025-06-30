//go:build integration
// +build integration

package integration

import (
	"fmt"
	"testing"
	"time"

	"deep-coding-agent/internal/memory"
	"deep-coding-agent/pkg/types"
	"deep-coding-agent/tests/testutils"
)

func setupMemoryManager(t *testing.T) (*memory.UnifiedMemoryManager, *testutils.TestContext) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	testCtx := testutils.NewTestContext(t)

	config := &types.MemoryManagerConfig{
		KnowledgeBase: &types.KnowledgeBaseConfig{
			MaxKnowledgeItems:    1000,
			AutoCategorization:   true,
			AutoTagging:          true,
			QualityThreshold:     0.7,
			ValidationEnabled:    true,
			GraphEnabled:         true,
			SimilarityThreshold:  0.8,
			ConsolidationEnabled: true,
		},
		PatternLearner: &types.PatternLearnerConfig{
			AutoLearning:          true,
			MinExamples:           3,
			MinQualityThreshold:   0.6,
			MaxVariations:         10,
			LearningRate:          0.1,
			FeedbackWeight:        0.3,
			ContextualLearning:    true,
			CrossLanguageLearning: false,
		},
		ProjectMemory: &types.ProjectMemoryConfig{
			AutoSnapshot:          true,
			SnapshotInterval:      "1h",
			MaxSnapshots:          10,
			CompressionEnabled:    true,
			ArchitectureTracking:  true,
			DecisionTracking:      true,
			LessonLearning:        true,
			ConfigurationTracking: true,
		},
		StorageConfig: &types.StorageConfig{
			Type:             "file",
			ConnectionString: testCtx.TempDir,
			MaxSize:          1024 * 1024 * 100, // 100MB
			RetentionPeriod:  "30d",
			BackupConfig: &types.BackupConfig{
				Enabled:     true,
				Interval:    "1h",
				MaxBackups:  5,
				Destination: testCtx.TempDir + "/backups",
			},
		},
		MetricsEnabled:     true,
		CacheEnabled:       true,
		BackupEnabled:      true,
		CompressionEnabled: false,
	}

	manager := memory.NewUnifiedMemoryManager(config)
	return manager, testCtx
}

func TestMemorySystem_FullWorkflow(t *testing.T) {
	manager, testCtx := setupMemoryManager(t)
	defer testCtx.Cleanup()

	ctx, cancel := testutils.CreateContextWithTimeout(30 * time.Second)
	defer cancel()

	// 1. Store various types of knowledge
	knowledgeItems := testutils.GenerateTestKnowledge(20, "integration_test")

	for _, k := range knowledgeItems {
		err := manager.Store(ctx, k)
		if err != nil {
			t.Fatalf("Failed to store knowledge: %v", err)
		}
	}

	// 2. Store code patterns
	patterns := testutils.GenerateTestPatterns(10, "integration_test")

	for _, p := range patterns {
		err := manager.StorePattern(ctx, p)
		if err != nil {
			t.Fatalf("Failed to store pattern: %v", err)
		}
	}

	// 3. Create project memory
	projectMemory := testutils.CreateTestProjectMemory("integration_test_project")
	err := manager.UpdateProjectMemory(ctx, projectMemory.ProjectID, projectMemory)
	if err != nil {
		t.Fatalf("Failed to update project memory: %v", err)
	}

	// 4. Test search functionality
	searchResults, err := manager.SearchKnowledge(ctx, "test knowledge", nil)
	if err != nil {
		t.Fatalf("Knowledge search failed: %v", err)
	}
	if len(searchResults) < 10 {
		t.Errorf("Expected at least 10 search results, got %d", len(searchResults))
	}

	// 5. Test pattern search
	patternResults, err := manager.SearchPatterns(ctx, map[string]interface{}{
		"language": "go",
	})
	if err != nil {
		t.Fatalf("Pattern search failed: %v", err)
	}
	if len(patternResults) != 10 {
		t.Errorf("Expected 10 pattern results, got %d", len(patternResults))
	}

	// 6. Test memory metrics
	stats, err := manager.GetMemoryStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get memory stats: %v", err)
	}

	knowledgeCount, ok := stats["knowledge_items"].(int)
	if !ok {
		t.Error("Failed to get knowledge_items from stats")
	} else if knowledgeCount != 20 {
		t.Errorf("Expected 20 knowledge items, got %d", knowledgeCount)
	}

	patternCount, ok := stats["patterns"].(int)
	if !ok {
		t.Error("Failed to get patterns from stats")
	} else if patternCount != 10 {
		t.Errorf("Expected 10 patterns, got %d", patternCount)
	}

	projectCount, ok := stats["project_memories"].(int)
	if !ok {
		t.Error("Failed to get project_memories from stats")
	} else if projectCount != 1 {
		t.Errorf("Expected 1 project memory, got %d", projectCount)
	}

	// 7. Test memory optimization
	result, err := manager.OptimizeMemory(ctx)
	if err != nil {
		t.Fatalf("Memory optimization failed: %v", err)
	}
	if !result.Success {
		t.Error("Memory optimization was not successful")
	}
}

func TestMemorySystem_ConcurrentOperations(t *testing.T) {
	manager, testCtx := setupMemoryManager(t)
	defer testCtx.Cleanup()

	ctx, cancel := testutils.CreateContextWithTimeout(60 * time.Second)
	defer cancel()

	const numGoroutines = 5
	const itemsPerGoroutine = 20

	// Channel to collect errors
	errChan := make(chan error, numGoroutines*2)
	done := make(chan bool, numGoroutines*2)

	// Concurrent knowledge storage
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer func() { done <- true }()

			knowledge := testutils.GenerateTestKnowledge(itemsPerGoroutine,
				fmt.Sprintf("concurrent_worker_%d", workerID))

			for _, k := range knowledge {
				if err := manager.Store(ctx, k); err != nil {
					errChan <- err
					return
				}
			}
		}(i)
	}

	// Concurrent pattern storage
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer func() { done <- true }()

			patterns := testutils.GenerateTestPatterns(itemsPerGoroutine/4,
				fmt.Sprintf("concurrent_pattern_worker_%d", workerID))

			for _, p := range patterns {
				if err := manager.StorePattern(ctx, p); err != nil {
					errChan <- err
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines*2; i++ {
		select {
		case <-done:
			// Goroutine completed
		case err := <-errChan:
			t.Fatalf("Concurrent operation failed: %v", err)
		case <-ctx.Done():
			t.Fatal("Test timed out")
		}
	}

	// Verify final state
	stats, err := manager.GetMemoryStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get final stats: %v", err)
	}

	expectedKnowledge := numGoroutines * itemsPerGoroutine
	expectedPatterns := numGoroutines * (itemsPerGoroutine / 4)

	knowledgeCount, ok := stats["knowledge_items"].(int)
	if !ok {
		t.Error("Failed to get knowledge_items from final stats")
	} else if knowledgeCount != expectedKnowledge {
		t.Errorf("Expected %d knowledge items, got %d", expectedKnowledge, knowledgeCount)
	}

	patternCount, ok := stats["patterns"].(int)
	if !ok {
		t.Error("Failed to get patterns from final stats")
	} else if patternCount != expectedPatterns {
		t.Errorf("Expected %d patterns, got %d", expectedPatterns, patternCount)
	}
}

func TestMemorySystem_LargeDataset(t *testing.T) {
	manager, testCtx := setupMemoryManager(t)
	defer testCtx.Cleanup()

	ctx, cancel := testutils.CreateContextWithTimeout(120 * time.Second)
	defer cancel()

	const largeDatasetSize = 500

	// Store large dataset
	t.Log("Storing large dataset...")
	start := time.Now()

	knowledge := testutils.GenerateTestKnowledge(largeDatasetSize, "large_dataset")
	for i, k := range knowledge {
		err := manager.Store(ctx, k)
		if err != nil {
			t.Fatalf("Failed to store knowledge item %d: %v", i, err)
		}

		if i%100 == 0 {
			t.Logf("Stored %d/%d items", i+1, largeDatasetSize)
		}
	}

	storeTime := time.Since(start)
	t.Logf("Stored %d items in %v (avg: %v per item)",
		largeDatasetSize, storeTime, storeTime/time.Duration(largeDatasetSize))

	// Perform searches on large dataset
	t.Log("Performing searches on large dataset...")
	start = time.Now()

	searchQueries := []string{
		"test knowledge",
		"content",
		"large_dataset",
		"experience",
	}

	for _, query := range searchQueries {
		results, err := manager.SearchKnowledge(ctx, query, nil)
		if err != nil {
			t.Fatalf("Search for '%s' failed: %v", query, err)
		}
		if len(results) == 0 {
			t.Errorf("Search for '%s' returned no results", query)
		}
	}

	searchTime := time.Since(start)
	t.Logf("Completed %d searches in %v (avg: %v per search)",
		len(searchQueries), searchTime, searchTime/time.Duration(len(searchQueries)))

	// Test memory usage
	stats, err := manager.GetMemoryStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	knowledgeCount, ok := stats["knowledge_items"].(int)
	if !ok {
		t.Error("Failed to get knowledge_items from stats")
	} else if knowledgeCount != largeDatasetSize {
		t.Errorf("Expected %d knowledge items, got %d", largeDatasetSize, knowledgeCount)
	}

	// Performance assertions
	avgStoreTime := storeTime / time.Duration(largeDatasetSize)
	avgSearchTime := searchTime / time.Duration(len(searchQueries))

	if avgStoreTime > 50*time.Millisecond {
		t.Logf("Warning: Average store time (%v) is high", avgStoreTime)
	}
	if avgSearchTime > 100*time.Millisecond {
		t.Logf("Warning: Average search time (%v) is high", avgSearchTime)
	}
}

func TestMemorySystem_MemoryCleanup(t *testing.T) {
	manager, testCtx := setupMemoryManager(t)
	defer testCtx.Cleanup()

	ctx, cancel := testutils.CreateContextWithTimeout(60 * time.Second)
	defer cancel()

	// Measure initial memory
	initialStats, err := manager.GetMemoryStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get initial stats: %v", err)
	}

	// Perform memory-intensive operations
	for round := 0; round < 5; round++ {
		t.Logf("Memory test round %d/5", round+1)

		// Store and delete data repeatedly
		knowledge := testutils.GenerateTestKnowledge(50, fmt.Sprintf("leak_test_round_%d", round))

		// Store
		for _, k := range knowledge {
			err := manager.Store(ctx, k)
			if err != nil {
				t.Fatalf("Failed to store knowledge in round %d: %v", round, err)
			}
		}

		// Delete
		for _, k := range knowledge {
			err := manager.Delete(ctx, k.ID)
			if err != nil {
				t.Fatalf("Failed to delete knowledge in round %d: %v", round, err)
			}
		}

		// Trigger cleanup
		err = manager.CleanupMemory(ctx)
		if err != nil {
			t.Fatalf("Memory cleanup failed in round %d: %v", round, err)
		}
	}

	// Check final memory state
	finalStats, err := manager.GetMemoryStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get final stats: %v", err)
	}

	// Should be back to initial state
	initialCount, ok1 := initialStats["knowledge_items"].(int)
	finalCount, ok2 := finalStats["knowledge_items"].(int)

	if !ok1 || !ok2 {
		t.Error("Failed to get knowledge_items from stats")
	} else if initialCount != finalCount {
		t.Errorf("Memory leak detected: initial=%d, final=%d", initialCount, finalCount)
	}

	t.Log("Memory cleanup test completed successfully")
}
