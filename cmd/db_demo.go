package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"deep-coding-agent/internal/context/algorithms"
	"deep-coding-agent/internal/context/storage"
)

// runDatabaseDemo 运行数据库集成演示
func runDatabaseDemo() {
	fmt.Println("🚀 深度编程代理 - SQLite数据库集成演示")
	fmt.Println("=" + string(make([]byte, 50)))

	// 1. 创建SQLite存储
	config := storage.StorageConfig{
		Type: "sqlite",
		Path: ":memory:", // 使用内存数据库进行演示
	}

	sqliteStorage, err := storage.NewSQLiteStorage(config)
	if err != nil {
		log.Fatalf("❌ 创建SQLite存储失败: %v", err)
	}
	defer sqliteStorage.Close()

	ctx := context.Background()

	fmt.Println("✅ SQLite存储初始化成功")

	// 2. 准备演示数据
	docs := []storage.Document{
		{
			ID:      "ai-ml-guide",
			Title:   "人工智能与机器学习指南",
			Content: "深度学习神经网络机器学习算法人工智能深度学习框架TensorFlow PyTorch模型训练数据预处理特征工程",
			Metadata: map[string]string{
				"category": "AI/ML",
				"level":    "advanced",
				"language": "通用",
			},
			Created: time.Now(),
		},
		{
			ID:      "go-concurrency",
			Title:   "Go语言并发编程最佳实践",
			Content: "goroutine channel 并发并行协程异步编程 select mutex 锁机制通信共享内存模式设计并发安全",
			Metadata: map[string]string{
				"category": "编程语言",
				"level":    "intermediate",
				"language": "Go",
			},
			Created: time.Now(),
		},
		{
			ID:      "db-optimization",
			Title:   "数据库性能优化技术",
			Content: "索引优化查询优化SQL调优数据库设计范式分布式数据库分片复制主从同步事务ACID性质",
			Metadata: map[string]string{
				"category": "数据库",
				"level":    "expert",
				"language": "通用",
			},
			Created: time.Now(),
		},
		{
			ID:      "microservices",
			Title:   "微服务架构设计模式",
			Content: "微服务架构分布式系统服务治理API网关负载均衡服务发现容器化Docker Kubernetes云原生",
			Metadata: map[string]string{
				"category": "系统架构",
				"level":    "expert",
				"language": "通用",
			},
			Created: time.Now(),
		},
		{
			ID:      "frontend-react",
			Title:   "React前端开发进阶",
			Content: "React组件状态管理Redux Hooks JSX虚拟DOM生命周期响应式设计UI组件库前端工程化",
			Metadata: map[string]string{
				"category": "前端开发",
				"level":    "intermediate",
				"language": "JavaScript",
			},
			Created: time.Now(),
		},
	}

	// 3. 批量存储文档
	fmt.Println("\n📚 存储技术文档...")
	start := time.Now()
	err = sqliteStorage.BatchStore(ctx, docs)
	if err != nil {
		log.Fatalf("❌ 存储文档失败: %v", err)
	}
	storeTime := time.Since(start)
	fmt.Printf("✅ 成功存储 %d 个文档，耗时: %v\n", len(docs), storeTime)

	// 4. 生成向量并存储
	fmt.Println("\n🔢 生成文档向量...")
	embeddingConfig := algorithms.DefaultEmbeddingConfig()

	vectors := make(map[string][]float64)
	for _, doc := range docs {
		text := doc.Title + " " + doc.Content
		vector := algorithms.GenerateEmbedding(text, embeddingConfig)
		vectors[doc.ID] = vector
	}

	start = time.Now()
	err = sqliteStorage.BatchStoreVectors(ctx, vectors)
	if err != nil {
		log.Fatalf("❌ 存储向量失败: %v", err)
	}
	vectorTime := time.Since(start)
	fmt.Printf("✅ 成功存储 %d 个向量，耗时: %v\n", len(vectors), vectorTime)

	// 5. 演示查询功能
	fmt.Println("\n🔍 数据库查询演示...")

	// 根据ID获取文档
	doc, err := sqliteStorage.Get(ctx, "go-concurrency")
	if err != nil {
		log.Printf("❌ 获取文档失败: %v", err)
	} else {
		fmt.Printf("📄 ID查询: %s\n", doc.Title)
	}

	// 统计文档数量
	count, err := sqliteStorage.Count(ctx)
	if err != nil {
		log.Printf("❌ 统计失败: %v", err)
	} else {
		fmt.Printf("📊 文档总数: %d\n", count)
	}

	// 分页查询
	pageSize := 3
	pagedDocs, err := sqliteStorage.List(ctx, pageSize, 0)
	if err != nil {
		log.Printf("❌ 分页查询失败: %v", err)
	} else {
		fmt.Printf("📋 分页查询 (前%d个): %d 条结果\n", pageSize, len(pagedDocs))
	}

	// 6. 智能向量搜索演示
	fmt.Println("\n🎯 智能向量搜索演示...")

	searchQueries := []string{
		"机器学习和深度学习",
		"并发编程和goroutine",
		"数据库优化和性能调优",
		"微服务架构设计",
		"React前端开发",
	}

	for _, query := range searchQueries {
		fmt.Printf("\n🔍 搜索查询: \"%s\"\n", query)

		queryVector := algorithms.GenerateEmbedding(query, embeddingConfig)

		start = time.Now()
		results, err := sqliteStorage.SearchSimilar(ctx, queryVector, 3)
		if err != nil {
			log.Printf("❌ 搜索失败: %v", err)
			continue
		}
		searchTime := time.Since(start)

		fmt.Printf("⚡ 搜索耗时: %v\n", searchTime)
		fmt.Printf("📊 找到 %d 个相关文档:\n", len(results))

		for i, result := range results {
			fmt.Printf("  %d. %s (相似度: %.3f)\n",
				i+1, result.Document.Title, result.Similarity)
		}
	}

	// 7. 性能统计
	fmt.Println("\n📈 性能统计报告...")

	vectorCount := sqliteStorage.GetVectorCount()
	dimensions := sqliteStorage.GetDimensions()
	metrics := sqliteStorage.GetMetrics()

	fmt.Printf("📄 文档数量: %d\n", metrics.DocumentCount)
	fmt.Printf("🔢 向量数量: %d\n", vectorCount)
	fmt.Printf("📐 向量维度: %d\n", dimensions)
	fmt.Printf("📖 读操作: %d 次\n", metrics.ReadOps)
	fmt.Printf("✏️  写操作: %d 次\n", metrics.WriteOps)
	fmt.Printf("⏱️  运行时间: %v\n", metrics.Uptime)

	// 8. 阈值搜索演示
	fmt.Println("\n📏 阈值搜索演示...")
	threshold := 0.5
	query := "人工智能机器学习"
	queryVector := algorithms.GenerateEmbedding(query, embeddingConfig)

	thresholdResults, err := sqliteStorage.SearchByThreshold(ctx, queryVector, threshold)
	if err != nil {
		log.Printf("❌ 阈值搜索失败: %v", err)
	} else {
		fmt.Printf("🎚️ 阈值搜索 (>= %.1f): 找到 %d 个高相关性文档\n", threshold, len(thresholdResults))
		for _, result := range thresholdResults {
			fmt.Printf("  - %s (相似度: %.3f)\n", result.Document.Title, result.Similarity)
		}
	}

	// 9. 批量操作演示
	fmt.Println("\n📦 批量操作演示...")

	// 批量获取
	ids := []string{"ai-ml-guide", "go-concurrency", "db-optimization"}
	start = time.Now()
	batchDocs, err := sqliteStorage.BatchGet(ctx, ids)
	if err != nil {
		log.Printf("❌ 批量获取失败: %v", err)
	} else {
		batchTime := time.Since(start)
		fmt.Printf("📥 批量获取 %d 个文档，耗时: %v\n", len(batchDocs), batchTime)
	}

	fmt.Println("\n🎉 SQLite数据库集成演示完成！")
	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Println("✅ 主要功能验证:")
	fmt.Println("  - ✅ 文档存储与检索")
	fmt.Println("  - ✅ 向量生成与存储")
	fmt.Println("  - ✅ 智能相似度搜索")
	fmt.Println("  - ✅ 批量操作支持")
	fmt.Println("  - ✅ 性能监控统计")
	fmt.Println("  - ✅ 阈值过滤搜索")

	fmt.Printf("\n📊 性能总结:\n")
	fmt.Printf("  - 文档存储: %.2f docs/ms\n", float64(len(docs))/float64(storeTime.Nanoseconds())*1000000)
	fmt.Printf("  - 向量存储: %.2f vectors/ms\n", float64(len(vectors))/float64(vectorTime.Nanoseconds())*1000000)
	fmt.Printf("  - 内存使用: SQLite内存数据库\n")
	fmt.Printf("  - 存储引擎: 高效SQLite\n")
}
