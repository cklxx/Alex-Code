package storage

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestChromemStorage_Basic(t *testing.T) {
	// 创建内存chromem存储
	config := StorageConfig{
		Type: "chromem",
		Path: ":memory:",
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		t.Fatalf("Failed to create Chromem storage: %v", err)
	}
	defer func() { if err := storage.Close(); err != nil { t.Logf("Error closing storage: %v", err) } }()

	ctx := context.Background()

	// 测试文档存储
	doc := Document{
		ID:      "test-doc-1",
		Title:   "人工智能基础",
		Content: "人工智能是一门研究智能机器的科学。机器学习深度学习神经网络算法优化",
		Metadata: map[string]string{
			"category": "AI",
			"level":    "beginner",
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

func TestChromemStorage_BatchOperations(t *testing.T) {
	config := StorageConfig{
		Type: "chromem",
		Path: ":memory:",
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		t.Fatalf("Failed to create Chromem storage: %v", err)
	}
	defer func() { if err := storage.Close(); err != nil { t.Logf("Error closing storage: %v", err) } }()

	ctx := context.Background()

	// 准备测试文档
	docs := []Document{
		{
			ID:       "ai-doc-1",
			Title:    "机器学习基础",
			Content:  "机器学习是人工智能的子领域。监督学习无监督学习强化学习算法模型训练",
			Metadata: map[string]string{"category": "ML"},
			Created:  time.Now(),
		},
		{
			ID:       "ai-doc-2",
			Title:    "深度学习入门",
			Content:  "深度学习基于神经网络。卷积神经网络循环神经网络Transformer架构",
			Metadata: map[string]string{"category": "DL"},
			Created:  time.Now(),
		},
		{
			ID:       "ai-doc-3",
			Title:    "自然语言处理",
			Content:  "自然语言处理研究计算机与人类语言交互。词嵌入语言模型文本分析情感分析",
			Metadata: map[string]string{"category": "NLP"},
			Created:  time.Now(),
		},
	}

	// 批量存储
	err = storage.BatchStore(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to batch store documents: %v", err)
	}

	// 批量获取
	ids := []string{"ai-doc-1", "ai-doc-2", "ai-doc-3"}
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

func TestChromemStorage_SemanticSearch(t *testing.T) {
	config := StorageConfig{
		Type: "chromem",
		Path: ":memory:",
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		t.Fatalf("Failed to create Chromem storage: %v", err)
	}
	defer func() { if err := storage.Close(); err != nil { t.Logf("Error closing storage: %v", err) } }()

	ctx := context.Background()

	// 添加多种类型的技术文档
	docs := []Document{
		{
			ID:       "ai-ml",
			Title:    "机器学习与人工智能",
			Content:  "机器学习是人工智能的核心技术。深度学习神经网络算法模型训练数据分析",
			Metadata: map[string]string{"domain": "AI"},
			Created:  time.Now(),
		},
		{
			ID:       "web-dev",
			Title:    "Web开发技术栈",
			Content:  "前端开发后端开发JavaScript React Vue.js Node.js API设计数据库",
			Metadata: map[string]string{"domain": "Web"},
			Created:  time.Now(),
		},
		{
			ID:       "data-science",
			Title:    "数据科学实践",
			Content:  "数据科学结合统计学机器学习编程。数据清洗特征工程模型评估可视化",
			Metadata: map[string]string{"domain": "DataScience"},
			Created:  time.Now(),
		},
		{
			ID:       "cloud-computing",
			Title:    "云计算架构",
			Content:  "云计算提供弹性计算资源。AWS Azure GCP微服务容器Kubernetes DevOps",
			Metadata: map[string]string{"domain": "Cloud"},
			Created:  time.Now(),
		},
		{
			ID:       "mobile-dev",
			Title:    "移动应用开发",
			Content:  "移动开发包括原生开发跨平台开发。iOS Android React Native Flutter应用发布",
			Metadata: map[string]string{"domain": "Mobile"},
			Created:  time.Now(),
		},
	}

	// 批量存储文档
	err = storage.BatchStore(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to batch store documents: %v", err)
	}

	// 等待chromem处理文档 (可能需要一些时间生成嵌入)
	time.Sleep(100 * time.Millisecond)

	// 测试语义搜索
	testQueries := []struct {
		query    string
		expected string // 期望的最相关文档ID
	}{
		{"人工智能和机器学习", "ai-ml"},
		{"前端和后端开发", "web-dev"},
		{"数据分析和统计", "data-science"},
		{"云服务和容器技术", "cloud-computing"},
		{"手机应用和移动端", "mobile-dev"},
	}

	for _, test := range testQueries {
		t.Run(fmt.Sprintf("Search_%s", test.query), func(t *testing.T) {
			results, err := storage.SearchByText(ctx, test.query, 3)
			if err != nil {
				t.Errorf("Failed to search for '%s': %v", test.query, err)
				return
			}

			if len(results) == 0 {
				t.Errorf("No results found for query: %s", test.query)
				return
			}

			t.Logf("Query: '%s'", test.query)
			for i, result := range results {
				t.Logf("  %d. %s (similarity: %.3f)",
					i+1, result.Document.Title, result.Similarity)
			}

			// 验证相似度分数合理性
			for _, result := range results {
				if result.Similarity < 0 || result.Similarity > 1 {
					t.Errorf("Invalid similarity score: %.3f (should be 0-1)", result.Similarity)
				}
			}

			// 验证结果是按相似度排序的
			for i := 1; i < len(results); i++ {
				if results[i-1].Similarity < results[i].Similarity {
					t.Errorf("Results not sorted by similarity: %.3f < %.3f",
						results[i-1].Similarity, results[i].Similarity)
				}
			}
		})
	}
}

func TestChromemStorage_Performance(t *testing.T) {
	config := StorageConfig{
		Type: "chromem",
		Path: ":memory:",
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		t.Fatalf("Failed to create Chromem storage: %v", err)
	}
	defer func() { if err := storage.Close(); err != nil { t.Logf("Error closing storage: %v", err) } }()

	ctx := context.Background()

	// 性能测试：批量插入文档
	docCount := 100 // 减少数量以适应chromem的处理速度
	docs := make([]Document, docCount)
	for i := 0; i < docCount; i++ {
		docs[i] = Document{
			ID:      fmt.Sprintf("perf-doc-%d", i),
			Title:   fmt.Sprintf("性能测试文档 %d", i),
			Content: fmt.Sprintf("这是第 %d 个性能测试文档，包含机器学习深度学习人工智能算法优化等内容", i),
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

	// 等待chromem处理完成
	time.Sleep(500 * time.Millisecond)

	// 测试搜索性能
	start = time.Now()
	listedDocs, err := storage.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}
	listDuration := time.Since(start)

	t.Logf("List 10 documents took: %v", listDuration)
	t.Logf("Retrieved %d documents", len(listedDocs))

	// 测试语义搜索性能
	start = time.Now()
	searchResults, err := storage.SearchByText(ctx, "机器学习人工智能", 5)
	if err != nil {
		t.Fatalf("Failed to search documents: %v", err)
	}
	searchDuration := time.Since(start)

	t.Logf("Semantic search took: %v", searchDuration)
	t.Logf("Found %d search results", len(searchResults))

	// 获取存储指标
	metrics := storage.GetMetrics()
	t.Logf("Storage metrics: %+v", metrics)

	if metrics.DocumentCount == 0 {
		t.Error("Expected non-zero document count in metrics")
	}

	if metrics.WriteOps == 0 {
		t.Error("Expected non-zero write operations in metrics")
	}

	// chromem的性能要求相对宽松 (处理向量需要时间)
	if insertDuration > 10*time.Second {
		t.Errorf("Batch insert too slow: %v > 10s", insertDuration)
	}

	if searchDuration > 1*time.Second {
		t.Errorf("Semantic search too slow: %v > 1s", searchDuration)
	}
}

func TestChromemStorage_AdvancedFeatures(t *testing.T) {
	config := StorageConfig{
		Type: "chromem",
		Path: ":memory:",
		Options: map[string]string{
			"collection_name": "test_collection",
		},
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		t.Fatalf("Failed to create Chromem storage: %v", err)
	}
	defer func() { if err := storage.Close(); err != nil { t.Logf("Error closing storage: %v", err) } }()

	ctx := context.Background()

	// 测试自定义集合名称
	collection := storage.GetCollection()
	if collection.Name != "test_collection" {
		t.Errorf("Expected collection name 'test_collection', got '%s'", collection.Name)
	}

	// 添加有特定元数据的文档
	docs := []Document{
		{
			ID:      "tech-go",
			Title:   "Go语言编程",
			Content: "Go是Google开发的编程语言。并发编程goroutine channel高性能网络编程",
			Metadata: map[string]string{
				"language": "Go",
				"level":    "intermediate",
				"year":     "2024",
			},
			Created: time.Now(),
		},
		{
			ID:      "tech-python",
			Title:   "Python数据科学",
			Content: "Python在数据科学领域广泛应用。NumPy Pandas机器学习数据可视化",
			Metadata: map[string]string{
				"language": "Python",
				"level":    "beginner",
				"year":     "2024",
			},
			Created: time.Now(),
		},
		{
			ID:      "tech-rust",
			Title:   "Rust系统编程",
			Content: "Rust是系统级编程语言。内存安全并发安全零成本抽象性能优化",
			Metadata: map[string]string{
				"language": "Rust",
				"level":    "advanced",
				"year":     "2023",
			},
			Created: time.Now(),
		},
	}

	err = storage.BatchStore(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to store documents: %v", err)
	}

	// 等待处理
	time.Sleep(200 * time.Millisecond)

	// 测试带过滤器的查询
	results, err := storage.QueryWithFilter(ctx, "编程语言", 5, map[string]string{
		"level": "intermediate",
	})
	if err != nil {
		t.Fatalf("Failed to query with filter: %v", err)
	}

	t.Logf("Filtered query results: %d", len(results))
	for _, result := range results {
		t.Logf("  - %s (level: %s)", result.Document.Title, result.Document.Metadata["level"])
	}

	// 测试统计信息
	stats := storage.GetStats()
	t.Logf("Storage stats: %+v", stats)

	if stats["document_count"] != 3 {
		t.Errorf("Expected document_count 3, got %v", stats["document_count"])
	}

	if stats["collection_name"] != "test_collection" {
		t.Errorf("Expected collection_name 'test_collection', got %v", stats["collection_name"])
	}

	// 测试向量维度
	dimensions := storage.GetDimensions()
	if dimensions <= 0 {
		t.Errorf("Expected positive dimensions, got %d", dimensions)
	}
	t.Logf("Vector dimensions: %d", dimensions)

	// 测试向量数量
	vectorCount := storage.GetVectorCount()
	if vectorCount != 3 {
		t.Errorf("Expected vector count 3, got %d", vectorCount)
	}
}

func TestChromemStorage_ErrorHandling(t *testing.T) {
	config := StorageConfig{
		Type: "chromem",
		Path: ":memory:",
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		t.Fatalf("Failed to create Chromem storage: %v", err)
	}
	defer func() { if err := storage.Close(); err != nil { t.Logf("Error closing storage: %v", err) } }()

	ctx := context.Background()

	// 测试获取不存在的文档
	_, err = storage.Get(ctx, "non-existent-doc")
	if err == nil {
		t.Error("Expected error when getting non-existent document")
	}

	// 测试检查不存在文档的存在性
	exists, err := storage.Exists(ctx, "non-existent-doc")
	if err != nil {
		t.Fatalf("Unexpected error when checking non-existent document: %v", err)
	}
	if exists {
		t.Error("Non-existent document should not exist")
	}

	// 测试空查询
	results, err := storage.SearchByText(ctx, "", 5)
	if err != nil {
		t.Fatalf("Failed to handle empty query: %v", err)
	}
	t.Logf("Empty query returned %d results", len(results))

	// 测试批量获取空列表
	docs, err := storage.BatchGet(ctx, []string{})
	if err != nil {
		t.Fatalf("Failed to handle empty batch get: %v", err)
	}
	if len(docs) != 0 {
		t.Errorf("Expected 0 documents from empty batch get, got %d", len(docs))
	}
}
