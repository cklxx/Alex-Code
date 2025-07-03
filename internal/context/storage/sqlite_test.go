package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteStorage_Basic(t *testing.T) {
	// 使用临时文件
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	config := StorageConfig{
		Type: "sqlite",
		Path: dbPath,
	}

	storage, err := NewSQLiteStorage(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer func() { if err := storage.Close(); err != nil { t.Logf("Error closing storage: %v", err) } }()

	ctx := context.Background()

	// 测试文档存储
	doc := Document{
		ID:      "test-doc-1",
		Title:   "测试文档",
		Content: "这是一个测试文档的内容",
		Metadata: map[string]string{
			"author": "test-user",
			"type":   "test",
		},
		Created: time.Now(),
	}

	// 存储文档
	err = storage.Store(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to store document: %v", err)
	}

	// 获取文档
	retrievedDoc, err := storage.Get(ctx, "test-doc-1")
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}

	// 验证文档内容
	if retrievedDoc.ID != doc.ID {
		t.Errorf("Expected ID %s, got %s", doc.ID, retrievedDoc.ID)
	}
	if retrievedDoc.Title != doc.Title {
		t.Errorf("Expected title %s, got %s", doc.Title, retrievedDoc.Title)
	}
	if retrievedDoc.Content != doc.Content {
		t.Errorf("Expected content %s, got %s", doc.Content, retrievedDoc.Content)
	}
	if retrievedDoc.Metadata["author"] != doc.Metadata["author"] {
		t.Errorf("Expected author %s, got %s", doc.Metadata["author"], retrievedDoc.Metadata["author"])
	}

	// 检查文档存在性
	exists, err := storage.Exists(ctx, "test-doc-1")
	if err != nil {
		t.Fatalf("Failed to check document existence: %v", err)
	}
	if !exists {
		t.Error("Document should exist")
	}

	// 统计文档数量
	count, err := storage.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count documents: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// 删除文档
	err = storage.Delete(ctx, "test-doc-1")
	if err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}

	// 验证文档已删除
	exists, err = storage.Exists(ctx, "test-doc-1")
	if err != nil {
		t.Fatalf("Failed to check document existence: %v", err)
	}
	if exists {
		t.Error("Document should not exist after deletion")
	}
}

func TestSQLiteStorage_BatchOperations(t *testing.T) {
	config := StorageConfig{
		Type: "sqlite",
		Path: ":memory:", // 内存数据库
	}

	storage, err := NewSQLiteStorage(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer func() { if err := storage.Close(); err != nil { t.Logf("Error closing storage: %v", err) } }()

	ctx := context.Background()

	// 准备测试文档
	docs := []Document{
		{
			ID:       "batch-doc-1",
			Title:    "批量文档1",
			Content:  "批量测试内容1",
			Metadata: map[string]string{"batch": "1"},
			Created:  time.Now(),
		},
		{
			ID:       "batch-doc-2",
			Title:    "批量文档2",
			Content:  "批量测试内容2",
			Metadata: map[string]string{"batch": "2"},
			Created:  time.Now(),
		},
		{
			ID:       "batch-doc-3",
			Title:    "批量文档3",
			Content:  "批量测试内容3",
			Metadata: map[string]string{"batch": "3"},
			Created:  time.Now(),
		},
	}

	// 批量存储
	err = storage.BatchStore(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to batch store documents: %v", err)
	}

	// 批量获取
	ids := []string{"batch-doc-1", "batch-doc-2", "batch-doc-3"}
	retrievedDocs, err := storage.BatchGet(ctx, ids)
	if err != nil {
		t.Fatalf("Failed to batch get documents: %v", err)
	}

	if len(retrievedDocs) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(retrievedDocs))
	}

	// 列出文档
	listedDocs, err := storage.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}

	if len(listedDocs) != 3 {
		t.Errorf("Expected 3 documents in list, got %d", len(listedDocs))
	}

	// 批量删除
	err = storage.BatchDelete(ctx, ids)
	if err != nil {
		t.Fatalf("Failed to batch delete documents: %v", err)
	}

	// 验证删除结果
	count, err := storage.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count documents after batch delete: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0 after batch delete, got %d", count)
	}
}

func TestSQLiteStorage_Vectors(t *testing.T) {
	config := StorageConfig{
		Type: "sqlite",
		Path: ":memory:",
	}

	storage, err := NewSQLiteStorage(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer func() { if err := storage.Close(); err != nil { t.Logf("Error closing storage: %v", err) } }()

	ctx := context.Background()

	// 先添加一个文档（向量需要关联到文档）
	doc := Document{
		ID:      "vector-doc-1",
		Title:   "向量测试文档",
		Content: "用于测试向量存储的文档",
		Created: time.Now(),
	}

	err = storage.Store(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to store document: %v", err)
	}

	// 测试向量存储
	vector := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	err = storage.StoreVector(ctx, "vector-doc-1", vector)
	if err != nil {
		t.Fatalf("Failed to store vector: %v", err)
	}

	// 获取向量
	retrievedVector, err := storage.GetVector(ctx, "vector-doc-1")
	if err != nil {
		t.Fatalf("Failed to get vector: %v", err)
	}

	if len(retrievedVector) != len(vector) {
		t.Errorf("Expected vector length %d, got %d", len(vector), len(retrievedVector))
	}

	for i, v := range vector {
		if retrievedVector[i] != v {
			t.Errorf("Expected vector[%d] = %f, got %f", i, v, retrievedVector[i])
		}
	}

	// 测试向量搜索
	queryVector := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	results, err := storage.SearchSimilar(ctx, queryVector, 5)
	if err != nil {
		t.Fatalf("Failed to search similar vectors: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one search result")
	}

	// 验证相似度
	if len(results) > 0 && results[0].Similarity <= 0 {
		t.Error("Expected positive similarity score")
	}

	// 测试阈值搜索
	thresholdResults, err := storage.SearchByThreshold(ctx, queryVector, 0.5)
	if err != nil {
		t.Fatalf("Failed to search by threshold: %v", err)
	}

	for _, result := range thresholdResults {
		if result.Similarity < 0.5 {
			t.Errorf("Expected similarity >= 0.5, got %f", result.Similarity)
		}
	}

	// 批量存储向量
	vectors := map[string][]float64{
		"vector-doc-1": {0.1, 0.2, 0.3, 0.4, 0.5},
	}

	err = storage.BatchStoreVectors(ctx, vectors)
	if err != nil {
		t.Fatalf("Failed to batch store vectors: %v", err)
	}

	// 获取向量统计
	dimensions := storage.GetDimensions()
	if dimensions != 5 {
		t.Errorf("Expected dimensions 5, got %d", dimensions)
	}

	vectorCount := storage.GetVectorCount()
	if vectorCount == 0 {
		t.Error("Expected non-zero vector count")
	}

	// 删除向量
	err = storage.DeleteVector(ctx, "vector-doc-1")
	if err != nil {
		t.Fatalf("Failed to delete vector: %v", err)
	}
}

func TestSQLiteStorage_Performance(t *testing.T) {
	config := StorageConfig{
		Type: "sqlite",
		Path: ":memory:",
	}

	storage, err := NewSQLiteStorage(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer func() { if err := storage.Close(); err != nil { t.Logf("Error closing storage: %v", err) } }()

	ctx := context.Background()

	// 性能测试：批量插入文档
	docCount := 1000
	docs := make([]Document, docCount)
	for i := 0; i < docCount; i++ {
		docs[i] = Document{
			ID:      fmt.Sprintf("perf-doc-%d", i),
			Title:   fmt.Sprintf("性能测试文档 %d", i),
			Content: fmt.Sprintf("这是第 %d 个性能测试文档的内容", i),
			Metadata: map[string]string{
				"index": fmt.Sprintf("%d", i),
				"type":  "performance",
			},
			Created: time.Now(),
		}
	}

	// 测试批量插入性能
	start := time.Now()
	err = storage.BatchStore(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to batch store documents: %v", err)
	}
	insertDuration := time.Since(start)

	t.Logf("Batch insert %d documents took: %v", docCount, insertDuration)

	// 测试搜索性能
	start = time.Now()
	listedDocs, err := storage.List(ctx, 100, 0)
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}
	searchDuration := time.Since(start)

	t.Logf("List 100 documents took: %v", searchDuration)
	t.Logf("Retrieved %d documents", len(listedDocs))

	// 验证性能要求
	if insertDuration > 5*time.Second {
		t.Errorf("Batch insert too slow: %v > 5s", insertDuration)
	}

	if searchDuration > 100*time.Millisecond {
		t.Errorf("Search too slow: %v > 100ms", searchDuration)
	}

	// 获取存储指标
	metrics := storage.GetMetrics()
	t.Logf("Storage metrics: %+v", metrics)

	if metrics.DocumentCount == 0 {
		t.Error("Expected non-zero document count in metrics")
	}

	if metrics.WriteOps == 0 {
		t.Error("Expected non-zero write operations in metrics")
	}

	if metrics.ReadOps == 0 {
		t.Error("Expected non-zero read operations in metrics")
	}
}

func TestSQLiteStorage_ErrorHandling(t *testing.T) {
	config := StorageConfig{
		Type: "sqlite",
		Path: ":memory:",
	}

	storage, err := NewSQLiteStorage(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer func() { if err := storage.Close(); err != nil { t.Logf("Error closing storage: %v", err) } }()

	ctx := context.Background()

	// 测试获取不存在的文档
	_, err = storage.Get(ctx, "non-existent-doc")
	if err == nil {
		t.Error("Expected error when getting non-existent document")
	}

	// 测试删除不存在的文档
	err = storage.Delete(ctx, "non-existent-doc")
	if err == nil {
		t.Error("Expected error when deleting non-existent document")
	}

	// 测试获取不存在的向量
	_, err = storage.GetVector(ctx, "non-existent-vector")
	if err == nil {
		t.Error("Expected error when getting non-existent vector")
	}

	// 测试检查不存在文档的存在性
	exists, err := storage.Exists(ctx, "non-existent-doc")
	if err != nil {
		t.Fatalf("Unexpected error when checking non-existent document: %v", err)
	}
	if exists {
		t.Error("Non-existent document should not exist")
	}
}

func TestSQLiteStorage_ConcurrentAccess(t *testing.T) {
	config := StorageConfig{
		Type: "sqlite",
		Path: ":memory:",
	}

	storage, err := NewSQLiteStorage(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer func() { if err := storage.Close(); err != nil { t.Logf("Error closing storage: %v", err) } }()

	ctx := context.Background()

	// 并发写入测试
	const goroutines = 10
	const docsPerGoroutine = 10

	errChan := make(chan error, goroutines)

	for g := 0; g < goroutines; g++ {
		go func(goroutineID int) {
			for i := 0; i < docsPerGoroutine; i++ {
				doc := Document{
					ID:      fmt.Sprintf("concurrent-doc-%d-%d", goroutineID, i),
					Title:   fmt.Sprintf("并发文档 G%d-D%d", goroutineID, i),
					Content: fmt.Sprintf("并发测试内容 G%d-D%d", goroutineID, i),
					Created: time.Now(),
				}

				if err := storage.Store(ctx, doc); err != nil {
					errChan <- err
					return
				}
			}
			errChan <- nil
		}(g)
	}

	// 等待所有goroutine完成
	for i := 0; i < goroutines; i++ {
		if err := <-errChan; err != nil {
			t.Fatalf("Concurrent write failed: %v", err)
		}
	}

	// 验证所有文档都已存储
	expectedCount := uint64(goroutines * docsPerGoroutine)
	count, err := storage.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count documents: %v", err)
	}

	if count != expectedCount {
		t.Errorf("Expected count %d, got %d", expectedCount, count)
	}

	t.Logf("Successfully stored %d documents concurrently", count)
}
