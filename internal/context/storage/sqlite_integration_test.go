package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"deep-coding-agent/internal/context/algorithms"
)

// TestSQLiteIntegration 集成测试：SQLite + Context Engine
func TestSQLiteIntegration(t *testing.T) {
	// 创建SQLite存储
	config := StorageConfig{
		Type: "sqlite",
		Path: ":memory:",
	}

	storage, err := NewSQLiteStorage(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	// 测试数据：技术文档
	testDocs := []Document{
		{
			ID:      "go-concurrency",
			Title:   "Go语言并发编程指南",
			Content: "Go语言通过goroutine和channel提供了强大的并发编程能力。goroutine是轻量级的线程，channel用于goroutine之间的通信。",
			Metadata: map[string]string{
				"language":  "go",
				"category":  "concurrency",
				"level":     "intermediate",
				"keywords":  "goroutine,channel,concurrency",
			},
			Created: time.Now(),
		},
		{
			ID:      "python-async",
			Title:   "Python异步编程最佳实践",
			Content: "Python的asyncio库提供了异步编程支持。使用async/await语法可以编写高效的异步代码，避免阻塞操作。",
			Metadata: map[string]string{
				"language":  "python",
				"category":  "async",
				"level":     "advanced",
				"keywords":  "asyncio,async,await,coroutine",
			},
			Created: time.Now(),
		},
		{
			ID:      "rust-ownership",
			Title:   "Rust所有权系统深度解析",
			Content: "Rust的所有权系统是其内存安全的核心。通过所有权、借用和生命周期，Rust在编译时保证内存安全。",
			Metadata: map[string]string{
				"language":  "rust",
				"category":  "memory-safety",
				"level":     "expert",
				"keywords":  "ownership,borrowing,lifetime,memory",
			},
			Created: time.Now(),
		},
		{
			ID:      "js-promises",
			Title:   "JavaScript Promise和异步处理",
			Content: "JavaScript Promise提供了处理异步操作的优雅方式。结合async/await语法，可以写出更清晰的异步代码。",
			Metadata: map[string]string{
				"language":  "javascript",
				"category":  "async",
				"level":     "beginner",
				"keywords":  "promise,async,await,callback",
			},
			Created: time.Now(),
		},
		{
			ID:      "db-optimization",
			Title:   "数据库查询优化技巧",
			Content: "数据库查询优化包括索引设计、查询重写、执行计划分析等。合理的索引可以大幅提升查询性能。",
			Metadata: map[string]string{
				"category":  "database",
				"level":     "intermediate",
				"keywords":  "optimization,index,query,performance",
			},
			Created: time.Now(),
		},
	}

	// 1. 批量存储文档
	t.Log("📚 批量存储技术文档...")
	start := time.Now()
	err = storage.BatchStore(ctx, testDocs)
	if err != nil {
		t.Fatalf("Failed to batch store documents: %v", err)
	}
	storeTime := time.Since(start)
	t.Logf("✅ 存储 %d 个文档耗时: %v", len(testDocs), storeTime)

	// 2. 生成并存储向量
	t.Log("🔢 生成文档向量...")
	embeddingConfig := algorithms.DefaultEmbeddingConfig()
	
	vectors := make(map[string][]float64)
	for _, doc := range testDocs {
		text := doc.Title + " " + doc.Content
		vector := algorithms.GenerateEmbedding(text, embeddingConfig)
		vectors[doc.ID] = vector
	}

	start = time.Now()
	err = storage.BatchStoreVectors(ctx, vectors)
	if err != nil {
		t.Fatalf("Failed to batch store vectors: %v", err)
	}
	vectorTime := time.Since(start)
	t.Logf("✅ 存储 %d 个向量耗时: %v", len(vectors), vectorTime)

	// 3. 测试文档检索
	t.Log("🔍 测试文档检索...")
	
	// 根据ID获取文档
	doc, err := storage.Get(ctx, "go-concurrency")
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}
	t.Logf("📄 获取文档: %s", doc.Title)

	// 列出所有文档
	allDocs, err := storage.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}
	t.Logf("📋 列出文档: %d 个", len(allDocs))

	// 4. 测试向量相似搜索
	t.Log("🎯 测试向量相似搜索...")
	
	// 搜索与"并发编程"相关的文档
	queryText := "并发编程和异步处理"
	queryVector := algorithms.GenerateEmbedding(queryText, embeddingConfig)
	
	start = time.Now()
	results, err := storage.SearchSimilar(ctx, queryVector, 3)
	if err != nil {
		t.Fatalf("Failed to search similar vectors: %v", err)
	}
	searchTime := time.Since(start)
	
	t.Logf("🔍 查询: \"%s\"", queryText)
	t.Logf("⚡ 搜索耗时: %v", searchTime)
	t.Logf("📊 搜索结果: %d 个", len(results))
	
	for i, result := range results {
		t.Logf("  %d. %s (相似度: %.3f)", 
			i+1, result.Document.Title, result.Similarity)
	}

	// 验证搜索结果质量
	if len(results) == 0 {
		t.Error("❌ 搜索应该返回结果")
	}

	// 检查相似度分数合理性
	for _, result := range results {
		if result.Similarity < 0 || result.Similarity > 1 {
			t.Errorf("❌ 相似度分数应该在0-1之间，得到: %.3f", result.Similarity)
		}
	}

	// 5. 测试阈值搜索
	t.Log("📏 测试阈值搜索...")
	thresholdResults, err := storage.SearchByThreshold(ctx, queryVector, 0.3)
	if err != nil {
		t.Fatalf("Failed to search by threshold: %v", err)
	}
	
	t.Logf("🎚️ 阈值搜索 (>= 0.3): %d 个结果", len(thresholdResults))
	for _, result := range thresholdResults {
		if result.Similarity < 0.3 {
			t.Errorf("❌ 阈值搜索结果相似度应该 >= 0.3，得到: %.3f", result.Similarity)
		}
	}

	// 6. 性能基准测试
	t.Log("⚡ 性能基准测试...")
	
	// 批量查询性能
	ids := make([]string, len(testDocs))
	for i, doc := range testDocs {
		ids[i] = doc.ID
	}
	
	start = time.Now()
	batchDocs, err := storage.BatchGet(ctx, ids)
	if err != nil {
		t.Fatalf("Failed to batch get documents: %v", err)
	}
	batchTime := time.Since(start)
	
	t.Logf("📦 批量查询 %d 个文档耗时: %v", len(ids), batchTime)
	
	if len(batchDocs) != len(testDocs) {
		t.Errorf("❌ 批量查询应该返回 %d 个文档，得到 %d 个", len(testDocs), len(batchDocs))
	}

	// 7. 存储统计和指标
	t.Log("📊 存储统计信息...")
	
	count, err := storage.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count documents: %v", err)
	}
	
	vectorCount := storage.GetVectorCount()
	dimensions := storage.GetDimensions()
	metrics := storage.GetMetrics()
	
	t.Logf("📄 文档总数: %d", count)
	t.Logf("🔢 向量总数: %d", vectorCount)
	t.Logf("📐 向量维度: %d", dimensions)
	t.Logf("📈 存储指标:")
	t.Logf("  - 读操作: %d 次", metrics.ReadOps)
	t.Logf("  - 写操作: %d 次", metrics.WriteOps)
	t.Logf("  - 运行时间: %v", metrics.Uptime)

	// 8. 数据清理测试
	t.Log("🧹 测试数据清理...")
	
	// 删除部分文档
	deleteIDs := []string{"js-promises", "db-optimization"}
	err = storage.BatchDelete(ctx, deleteIDs)
	if err != nil {
		t.Fatalf("Failed to batch delete documents: %v", err)
	}
	
	// 验证删除结果
	newCount, err := storage.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count documents after deletion: %v", err)
	}
	
	expectedCount := count - uint64(len(deleteIDs))
	if newCount != expectedCount {
		t.Errorf("❌ 删除后文档数量应该是 %d，得到 %d", expectedCount, newCount)
	}
	
	t.Logf("✅ 成功删除 %d 个文档，剩余 %d 个", len(deleteIDs), newCount)

	// 9. 并发访问测试 (简化版本)
	t.Log("🔄 并发访问测试...")
	
	// 简单的并发读写测试
	errChan := make(chan error, 3)
	
	// 并发读取
	go func() {
		_, err := storage.Get(ctx, "go-concurrency")
		errChan <- err
	}()
	
	// 并发存储
	go func() {
		concurrentDoc := Document{
			ID:      "concurrent-test-doc",
			Title:   "并发测试文档",
			Content: "并发测试内容",
			Created: time.Now(),
		}
		errChan <- storage.Store(ctx, concurrentDoc)
	}()
	
	// 并发查询
	go func() {
		_, err := storage.List(ctx, 5, 0)
		errChan <- err
	}()
	
	// 等待所有操作完成
	for i := 0; i < 3; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("并发操作失败: %v", err)
		}
	}
	
	t.Log("✅ 并发访问测试通过")

	// 10. 最终验证
	t.Log("✅ 集成测试完成！")
	t.Log("==========================================")
	t.Logf("📊 最终统计:")
	t.Logf("  - 文档存储耗时: %v", storeTime)
	t.Logf("  - 向量存储耗时: %v", vectorTime)
	t.Logf("  - 搜索响应时间: %v", searchTime)
	t.Logf("  - 批量查询耗时: %v", batchTime)
	
	// 性能要求验证
	if storeTime > 1*time.Second {
		t.Errorf("❌ 文档存储过慢: %v > 1s", storeTime)
	}
	
	if searchTime > 100*time.Millisecond {
		t.Errorf("❌ 搜索响应过慢: %v > 100ms", searchTime)
	}
	
	t.Log("🎉 SQLite数据库集成测试全部通过！")
}

// testConcurrentOperationsWithSameStorage 使用同一个存储实例测试并发操作
func testConcurrentOperationsWithSameStorage(t *testing.T, storage *SQLiteStorage) {
	ctx := context.Background()
	
	const numGoroutines = 5
	const opsPerGoroutine = 10
	
	errChan := make(chan error, numGoroutines)
	
	// 启动多个goroutine进行并发操作
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < opsPerGoroutine; j++ {
				// 并发存储文档
				doc := Document{
					ID:      fmt.Sprintf("concurrent-%d-%d", id, j),
					Title:   fmt.Sprintf("并发文档 %d-%d", id, j),
					Content: fmt.Sprintf("并发测试内容 %d-%d", id, j),
					Metadata: map[string]string{
						"goroutine": fmt.Sprintf("%d", id),
						"operation": fmt.Sprintf("%d", j),
					},
					Created: time.Now(),
				}
				
				if err := storage.Store(ctx, doc); err != nil {
					errChan <- fmt.Errorf("concurrent store failed: %w", err)
					return
				}
				
				// 并发读取文档
				_, err := storage.Get(ctx, doc.ID)
				if err != nil {
					errChan <- fmt.Errorf("concurrent get failed: %w", err)
					return
				}
			}
			errChan <- nil
		}(i)
	}
	
	// 等待所有操作完成
	for i := 0; i < numGoroutines; i++ {
		if err := <-errChan; err != nil {
			t.Fatalf("Concurrent operation failed: %v", err)
		}
	}
	
	t.Logf("✅ 并发操作测试通过：%d goroutines × %d 操作", numGoroutines, opsPerGoroutine)
}

// TestSQLiteStorageWithRealData 使用真实数据的测试
func TestSQLiteStorageWithRealData(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实数据测试（使用 -short 标志）")
	}

	config := StorageConfig{
		Type: "sqlite",
		Path: ":memory:",
	}

	storage, err := NewSQLiteStorage(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	// 模拟大量真实数据
	t.Log("🗄️ 生成大量测试数据...")
	
	docs := generateLargeDataset(1000) // 1000个文档
	
	// 批量存储
	start := time.Now()
	err = storage.BatchStore(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to store large dataset: %v", err)
	}
	storeTime := time.Since(start)
	
	t.Logf("📊 存储 %d 个文档耗时: %v (%.2f docs/sec)", 
		len(docs), storeTime, float64(len(docs))/storeTime.Seconds())

	// 生成并存储向量
	t.Log("🔢 批量生成向量...")
	embeddingConfig := algorithms.DefaultEmbeddingConfig()
	
	vectors := make(map[string][]float64)
	for _, doc := range docs {
		text := doc.Title + " " + doc.Content
		vector := algorithms.GenerateEmbedding(text, embeddingConfig)
		vectors[doc.ID] = vector
	}
	
	start = time.Now()
	err = storage.BatchStoreVectors(ctx, vectors)
	if err != nil {
		t.Fatalf("Failed to store vectors: %v", err)
	}
	vectorTime := time.Since(start)
	
	t.Logf("🔢 存储 %d 个向量耗时: %v (%.2f vectors/sec)", 
		len(vectors), vectorTime, float64(len(vectors))/vectorTime.Seconds())

	// 测试搜索性能
	testQueries := []string{
		"机器学习算法",
		"数据库优化",
		"网络编程",
		"前端开发",
		"系统架构",
	}

	t.Log("🔍 测试搜索性能...")
	for _, query := range testQueries {
		queryVector := algorithms.GenerateEmbedding(query, embeddingConfig)
		
		start = time.Now()
		results, err := storage.SearchSimilar(ctx, queryVector, 10)
		if err != nil {
			t.Fatalf("Failed to search for '%s': %v", query, err)
		}
		searchTime := time.Since(start)
		
		t.Logf("  查询 '%s': %d 结果, 耗时 %v", query, len(results), searchTime)
		
		// 验证搜索性能
		if searchTime > 200*time.Millisecond {
			t.Errorf("搜索性能不达标: %v > 200ms for query '%s'", searchTime, query)
		}
	}

	// 存储统计
	finalMetrics := storage.GetMetrics()
	t.Logf("📈 最终存储指标:")
	t.Logf("  - 文档数量: %d", finalMetrics.DocumentCount)
	t.Logf("  - 读操作总数: %d", finalMetrics.ReadOps)
	t.Logf("  - 写操作总数: %d", finalMetrics.WriteOps)
	t.Logf("  - 运行时间: %v", finalMetrics.Uptime)
}

// generateLargeDataset 生成大量测试数据
func generateLargeDataset(count int) []Document {
	categories := []string{
		"机器学习", "数据库", "网络编程", "前端开发", "后端开发",
		"系统架构", "云计算", "DevOps", "安全", "移动开发",
	}
	
	languages := []string{
		"Go", "Python", "JavaScript", "Java", "C++", 
		"Rust", "TypeScript", "PHP", "Ruby", "Swift",
	}
	
	docs := make([]Document, count)
	
	for i := 0; i < count; i++ {
		category := categories[i%len(categories)]
		language := languages[i%len(languages)]
		
		docs[i] = Document{
			ID:      fmt.Sprintf("large-doc-%d", i),
			Title:   fmt.Sprintf("%s %s开发指南 %d", language, category, i),
			Content: fmt.Sprintf("这是关于%s语言在%s领域的详细指南第%d篇。包含了最佳实践、性能优化技巧、常见问题解决方案等内容。", language, category, i),
			Metadata: map[string]string{
				"category": category,
				"language": language,
				"index":    fmt.Sprintf("%d", i),
				"level":    []string{"beginner", "intermediate", "advanced"}[i%3],
			},
			Created: time.Now().Add(-time.Duration(i) * time.Hour), // 模拟不同的创建时间
		}
	}
	
	return docs
}