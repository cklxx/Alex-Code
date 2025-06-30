//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"deep-coding-agent/internal/memory"
	"deep-coding-agent/pkg/types"
	"deep-coding-agent/tests/testutils"

	"github.com/stretchr/testify/suite"
)

// MemoryIntegrationTestSuite tests memory system integration
type MemoryIntegrationTestSuite struct {
	suite.Suite
	testCtx *testutils.TestContext
	manager *memory.UnifiedMemoryManager
}

func (suite *MemoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		suite.T().Skip("Skipping integration tests in short mode")
	}
}

func (suite *MemoryIntegrationTestSuite) SetupTest() {
	suite.testCtx = testutils.NewTestContext(suite.T())

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
			ConnectionString: suite.testCtx.TempDir,
			MaxSize:          1024 * 1024 * 100, // 100MB
			RetentionPeriod:  "30d",
			BackupConfig: &types.BackupConfig{
				Enabled:     true,
				Interval:    "1h",
				MaxBackups:  5,
				Destination: suite.testCtx.TempDir + "/backups",
			},
		},
		MetricsEnabled:     true,
		CacheEnabled:       true,
		BackupEnabled:      true,
		CompressionEnabled: false,
	}

	suite.manager = memory.NewUnifiedMemoryManager(config)
}

func (suite *MemoryIntegrationTestSuite) TearDownTest() {
	if suite.manager != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		suite.manager.Shutdown(ctx)
	}
}

func (suite *MemoryIntegrationTestSuite) TestMemorySystem_FullWorkflow() {
	ctx, cancel := testutils.CreateContextWithTimeout(30 * time.Second)
	defer cancel()

	// 1. Store various types of knowledge
	knowledgeItems := testutils.GenerateTestKnowledge(20, "integration_test")

	for _, k := range knowledgeItems {
		err := suite.manager.Store(ctx, k)
		suite.Require().NoError(err)
	}

	// 2. Store code patterns
	patterns := testutils.GenerateTestPatterns(10, "integration_test")

	for _, p := range patterns {
		err := suite.manager.StorePattern(ctx, p)
		suite.Require().NoError(err)
	}

	// 3. Create project memory
	projectMemory := testutils.CreateTestProjectMemory("integration_test_project")
	err := suite.manager.UpdateProjectMemory(ctx, projectMemory.ProjectID, projectMemory)
	suite.Require().NoError(err)

	// 4. Test search functionality
	searchResults, err := suite.manager.SearchKnowledge(ctx, "test knowledge", nil)
	suite.Require().NoError(err)
	suite.Assert().GreaterOrEqual(len(searchResults), 10)

	// 5. Test pattern search
	patternResults, err := suite.manager.SearchPatterns(ctx, map[string]interface{}{
		"language": "go",
	})
	suite.Require().NoError(err)
	suite.Assert().Len(patternResults, 10)

	// 6. Test memory metrics
	stats, err := suite.manager.GetMemoryStats(ctx)
	suite.Require().NoError(err)
	suite.Assert().Equal(20, stats["knowledge_items"])
	suite.Assert().Equal(10, stats["patterns"])
	suite.Assert().Equal(1, stats["project_memories"])

	// 7. Test memory optimization
	result, err := suite.manager.OptimizeMemory(ctx)
	suite.Require().NoError(err)
	suite.Assert().True(result.Success)
}

func (suite *MemoryIntegrationTestSuite) TestMemorySystem_ConcurrentOperations() {
	ctx, cancel := testutils.CreateContextWithTimeout(60 * time.Second)
	defer cancel()

	const numGoroutines = 10
	const itemsPerGoroutine = 50

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
				if err := suite.manager.Store(ctx, k); err != nil {
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

			patterns := testutils.GenerateTestPatterns(itemsPerGoroutine/5,
				fmt.Sprintf("concurrent_pattern_worker_%d", workerID))

			for _, p := range patterns {
				if err := suite.manager.StorePattern(ctx, p); err != nil {
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
			suite.Fail("Concurrent operation failed", err.Error())
		case <-ctx.Done():
			suite.Fail("Test timed out")
		}
	}

	// Verify final state
	stats, err := suite.manager.GetMemoryStats(ctx)
	suite.Require().NoError(err)

	expectedKnowledge := numGoroutines * itemsPerGoroutine
	expectedPatterns := numGoroutines * (itemsPerGoroutine / 5)

	suite.Assert().Equal(expectedKnowledge, stats["knowledge_items"])
	suite.Assert().Equal(expectedPatterns, stats["patterns"])
}

func (suite *MemoryIntegrationTestSuite) TestMemorySystem_LargeDataset() {
	ctx, cancel := testutils.CreateContextWithTimeout(120 * time.Second)
	defer cancel()

	const largeDatasetSize = 1000

	// Store large dataset
	suite.T().Log("Storing large dataset...")
	start := time.Now()

	knowledge := testutils.GenerateTestKnowledge(largeDatasetSize, "large_dataset")
	for i, k := range knowledge {
		err := suite.manager.Store(ctx, k)
		suite.Require().NoError(err)

		if i%100 == 0 {
			suite.T().Logf("Stored %d/%d items", i+1, largeDatasetSize)
		}
	}

	storeTime := time.Since(start)
	suite.T().Logf("Stored %d items in %v (avg: %v per item)",
		largeDatasetSize, storeTime, storeTime/time.Duration(largeDatasetSize))

	// Perform searches on large dataset
	suite.T().Log("Performing searches on large dataset...")
	start = time.Now()

	searchQueries := []string{
		"test knowledge",
		"content",
		"large_dataset",
		"experience",
	}

	for _, query := range searchQueries {
		results, err := suite.manager.SearchKnowledge(ctx, query, nil)
		suite.Require().NoError(err)
		suite.Assert().Greater(len(results), 0)
	}

	searchTime := time.Since(start)
	suite.T().Logf("Completed %d searches in %v (avg: %v per search)",
		len(searchQueries), searchTime, searchTime/time.Duration(len(searchQueries)))

	// Test memory usage
	stats, err := suite.manager.GetMemoryStats(ctx)
	suite.Require().NoError(err)
	suite.Assert().Equal(largeDatasetSize, stats["knowledge_items"])

	// Performance assertions
	avgStoreTime := storeTime / time.Duration(largeDatasetSize)
	avgSearchTime := searchTime / time.Duration(len(searchQueries))

	suite.Assert().Less(avgStoreTime, 50*time.Millisecond,
		"Average store time should be less than 50ms")
	suite.Assert().Less(avgSearchTime, 100*time.Millisecond,
		"Average search time should be less than 100ms")
}

func (suite *MemoryIntegrationTestSuite) TestMemorySystem_PersistenceAndRecovery() {
	ctx, cancel := testutils.CreateContextWithTimeout(30 * time.Second)
	defer cancel()

	// Store test data
	originalKnowledge := testutils.GenerateTestKnowledge(50, "persistence_test")
	for _, k := range originalKnowledge {
		err := suite.manager.Store(ctx, k)
		suite.Require().NoError(err)
	}

	// Get initial stats
	stats, err := suite.manager.GetMemoryStats(ctx)
	suite.Require().NoError(err)
	originalCount := stats["knowledge_items"].(int)

	// Shutdown manager (simulate restart)
	err = suite.manager.Shutdown(ctx)
	suite.Require().NoError(err)

	// Create new manager with same config (simulate recovery)
	config := suite.manager.GetConfiguration()
	newManager := memory.NewUnifiedMemoryManager(config)

	// Verify data persistence (if implemented)
	// Note: This depends on the actual persistence implementation
	recoveredStats, err := newManager.GetMemoryStats(ctx)
	suite.Require().NoError(err)

	// For in-memory implementation, this might be 0
	// For persistent implementation, should match original
	suite.T().Logf("Original count: %d, Recovered count: %d",
		originalCount, recoveredStats["knowledge_items"])

	// Clean up new manager
	err = newManager.Shutdown(ctx)
	suite.Require().NoError(err)
}

func (suite *MemoryIntegrationTestSuite) TestMemorySystem_MemoryLeaks() {
	ctx, cancel := testutils.CreateContextWithTimeout(60 * time.Second)
	defer cancel()

	// Measure initial memory
	initialStats, err := suite.manager.GetMemoryStats(ctx)
	suite.Require().NoError(err)

	// Perform memory-intensive operations
	for round := 0; round < 10; round++ {
		suite.T().Logf("Memory test round %d/10", round+1)

		// Store and delete data repeatedly
		knowledge := testutils.GenerateTestKnowledge(100, fmt.Sprintf("leak_test_round_%d", round))

		// Store
		for _, k := range knowledge {
			err := suite.manager.Store(ctx, k)
			suite.Require().NoError(err)
		}

		// Delete
		for _, k := range knowledge {
			err := suite.manager.Delete(ctx, k.ID)
			suite.Require().NoError(err)
		}

		// Trigger cleanup
		err = suite.manager.CleanupMemory(ctx)
		suite.Require().NoError(err)
	}

	// Check final memory state
	finalStats, err := suite.manager.GetMemoryStats(ctx)
	suite.Require().NoError(err)

	// Should be back to initial state
	suite.Assert().Equal(initialStats["knowledge_items"], finalStats["knowledge_items"])
	suite.T().Logf("Memory leak test completed successfully")
}

func TestMemoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(MemoryIntegrationTestSuite))
}
