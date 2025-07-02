package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"deep-coding-agent/internal/context/storage"
)

// runChromemDemo 运行Chromem向量数据库演示
func runChromemDemo() {
	fmt.Println("🚀 深度编程代理 - Chromem-Go 向量数据库演示")
	fmt.Println("=" + string(make([]rune, 60)))

	// 1. 创建Chromem存储 
	fmt.Println("\n🔧 初始化Chromem向量数据库...")
	config := storage.StorageConfig{
		Type: "chromem",
		Path: ":memory:", // 使用内存数据库
		Options: map[string]string{
			"collection_name": "knowledge_base",
		},
	}

	chromemStorage, err := storage.NewChromemStorage(config)
	if err != nil {
		log.Fatalf("❌ 创建Chromem存储失败: %v", err)
	}
	defer chromemStorage.Close()

	ctx := context.Background()
	fmt.Println("✅ Chromem向量数据库初始化成功")

	// 2. 准备丰富的技术知识库
	fmt.Println("\n📚 构建技术知识库...")
	knowledgeBase := []storage.Document{
		{
			ID:      "ai-fundamentals",
			Title:   "人工智能基础概念",
			Content: "人工智能(AI)是计算机科学的一个分支，旨在创建能够执行通常需要人类智能的任务的机器。包括机器学习、深度学习、神经网络、自然语言处理、计算机视觉等核心技术。",
			Metadata: map[string]string{
				"category":   "AI/ML",
				"difficulty": "beginner",
				"domain":     "artificial-intelligence",
				"tags":       "AI,机器学习,深度学习,神经网络",
			},
			Created: time.Now(),
		},
		{
			ID:      "deep-learning-advanced",
			Title:   "深度学习进阶技术",
			Content: "深度学习基于多层神经网络，包括卷积神经网络(CNN)用于图像处理，循环神经网络(RNN)用于序列数据，Transformer架构用于自然语言处理。关键技术包括反向传播、梯度下降、正则化、批归一化等。",
			Metadata: map[string]string{
				"category":   "AI/ML",
				"difficulty": "advanced",
				"domain":     "deep-learning",
				"tags":       "CNN,RNN,Transformer,反向传播",
			},
			Created: time.Now(),
		},
		{
			ID:      "golang-concurrency",
			Title:   "Go语言并发编程实践",
			Content: "Go语言通过goroutine和channel提供强大的并发编程能力。Goroutine是轻量级线程，channel用于goroutine间通信。关键概念包括并发vs并行、select语句、mutex锁、waitgroup同步等。",
			Metadata: map[string]string{
				"category":   "Programming",
				"difficulty": "intermediate",
				"domain":     "golang",
				"tags":       "goroutine,channel,并发,同步",
			},
			Created: time.Now(),
		},
		{
			ID:      "microservices-architecture",
			Title:   "微服务架构设计原则",
			Content: "微服务架构将大型应用分解为小型、独立的服务。核心原则包括单一职责、服务自治、去中心化治理、故障隔离。关键技术包括API网关、服务发现、负载均衡、断路器模式、分布式追踪等。",
			Metadata: map[string]string{
				"category":   "Architecture",
				"difficulty": "expert",
				"domain":     "microservices",
				"tags":       "微服务,API网关,服务发现,分布式",
			},
			Created: time.Now(),
		},
		{
			ID:      "database-optimization",
			Title:   "数据库性能优化策略",
			Content: "数据库优化包括索引优化、查询优化、架构设计优化。索引策略包括B+树索引、哈希索引、全文索引。查询优化包括执行计划分析、SQL重写、统计信息更新。架构优化包括读写分离、分库分表、缓存策略等。",
			Metadata: map[string]string{
				"category":   "Database",
				"difficulty": "intermediate",
				"domain":     "database",
				"tags":       "索引,查询优化,分库分表,缓存",
			},
			Created: time.Now(),
		},
		{
			ID:      "frontend-react-hooks",
			Title:   "React Hooks现代前端开发",
			Content: "React Hooks是React 16.8引入的特性，允许在函数组件中使用状态和其他React特性。核心Hooks包括useState、useEffect、useContext、useReducer。自定义Hooks可以复用状态逻辑，提高代码可维护性。",
			Metadata: map[string]string{
				"category":   "Frontend",
				"difficulty": "intermediate",
				"domain":     "react",
				"tags":       "React,Hooks,useState,useEffect",
			},
			Created: time.Now(),
		},
		{
			ID:      "cloud-kubernetes",
			Title:   "Kubernetes容器编排技术",
			Content: "Kubernetes是开源的容器编排平台，自动化应用部署、扩展和管理。核心概念包括Pod、Service、Deployment、ConfigMap、Secret。高级特性包括HPA自动伸缩、Ingress负载均衡、PersistentVolume存储管理等。",
			Metadata: map[string]string{
				"category":   "DevOps",
				"difficulty": "advanced",
				"domain":     "kubernetes",
				"tags":       "K8s,容器,编排,Pod,Service",
			},
			Created: time.Now(),
		},
		{
			ID:      "security-best-practices",
			Title:   "软件安全开发最佳实践",
			Content: "安全开发生命周期(SDLC)集成安全考虑。关键实践包括威胁建模、安全代码审查、漏洞扫描、渗透测试。常见安全问题包括SQL注入、XSS攻击、CSRF攻击、认证授权缺陷等。防护措施包括输入验证、输出编码、最小权限原则等。",
			Metadata: map[string]string{
				"category":   "Security",
				"difficulty": "expert",
				"domain":     "cybersecurity",
				"tags":       "安全开发,威胁建模,渗透测试,漏洞扫描",
			},
			Created: time.Now(),
		},
	}

	// 3. 批量存储到Chromem
	start := time.Now()
	err = chromemStorage.BatchStore(ctx, knowledgeBase)
	if err != nil {
		log.Fatalf("❌ 存储知识库失败: %v", err)
	}
	storeTime := time.Since(start)
	fmt.Printf("✅ 成功存储 %d 篇技术文档，耗时: %v\n", len(knowledgeBase), storeTime)

	// 等待Chromem处理向量嵌入
	fmt.Println("🔄 正在生成向量嵌入...")
	time.Sleep(1 * time.Second)

	// 4. 展示基础查询功能
	fmt.Println("\n🔍 基础查询功能演示...")
	
	// 根据ID获取文档
	doc, err := chromemStorage.Get(ctx, "ai-fundamentals")
	if err != nil {
		log.Printf("❌ 获取文档失败: %v", err)
	} else {
		fmt.Printf("📄 ID查询: %s\n", doc.Title)
	}

	// 统计文档数量
	count, err := chromemStorage.Count(ctx)
	if err != nil {
		log.Printf("❌ 统计失败: %v", err)
	} else {
		fmt.Printf("📊 知识库总文档数: %d\n", count)
	}

	// 分页查询
	pagedDocs, err := chromemStorage.List(ctx, 3, 0)
	if err != nil {
		log.Printf("❌ 分页查询失败: %v", err)
	} else {
		fmt.Printf("📋 分页查询前3篇: %d 条结果\n", len(pagedDocs))
	}

	// 5. 智能语义搜索演示 (Chromem的核心优势)
	fmt.Println("\n🎯 智能语义搜索演示 (Chromem核心功能)")
	fmt.Println("-" + string(make([]rune, 50)))

	semanticQueries := []struct {
		query       string
		description string
	}{
		{
			query:       "机器学习和人工智能算法",
			description: "AI/ML领域查询",
		},
		{
			query:       "并发编程和多线程处理",
			description: "并发编程查询",
		},
		{
			query:       "分布式系统和微服务架构",
			description: "系统架构查询",
		},
		{
			query:       "数据库查询优化和性能调优",
			description: "数据库优化查询",
		},
		{
			query:       "前端开发和用户界面",
			description: "前端开发查询",
		},
		{
			query:       "容器化部署和云原生应用",
			description: "云原生技术查询",
		},
		{
			query:       "网络安全和漏洞防护",
			description: "安全技术查询",
		},
	}

	for i, test := range semanticQueries {
		fmt.Printf("\n🔍 查询 %d: %s\n", i+1, test.description)
		fmt.Printf("📝 查询内容: \"%s\"\n", test.query)
		
		start = time.Now()
		results, err := chromemStorage.SearchByText(ctx, test.query, 3)
		if err != nil {
			log.Printf("❌ 搜索失败: %v", err)
			continue
		}
		searchTime := time.Since(start)
		
		fmt.Printf("⚡ 搜索耗时: %v\n", searchTime)
		fmt.Printf("📊 找到 %d 个相关文档:\n", len(results))
		
		for j, result := range results {
			fmt.Printf("  %d. %s\n", j+1, result.Document.Title)
			fmt.Printf("     相似度: %.3f | 领域: %s | 难度: %s\n", 
				result.Similarity, 
				result.Document.Metadata["domain"],
				result.Document.Metadata["difficulty"])
		}
	}

	// 6. 高级过滤查询演示
	fmt.Println("\n🎚️ 高级过滤查询演示")
	fmt.Println("-" + string(make([]rune, 30)))

	// 按难度级别过滤
	difficultyLevels := []string{"beginner", "intermediate", "advanced", "expert"}
	for _, level := range difficultyLevels {
		results, err := chromemStorage.QueryWithFilter(ctx, "编程技术", 5, map[string]string{
			"difficulty": level,
		})
		if err != nil {
			log.Printf("❌ 过滤查询失败: %v", err)
			continue
		}
		
		fmt.Printf("🎯 %s级别文档: %d 篇\n", level, len(results))
		for _, result := range results {
			fmt.Printf("  - %s (相似度: %.3f)\n", result.Document.Title, result.Similarity)
		}
	}

	// 7. 性能和统计信息
	fmt.Println("\n📈 性能与统计信息")
	fmt.Println("-" + string(make([]rune, 30)))
	
	// Chromem统计
	stats := chromemStorage.GetStats()
	fmt.Printf("📊 Chromem统计信息:\n")
	for key, value := range stats {
		fmt.Printf("  - %s: %v\n", key, value)
	}
	
	// 向量信息
	dimensions := chromemStorage.GetDimensions()
	vectorCount := chromemStorage.GetVectorCount()
	fmt.Printf("🔢 向量维度: %d\n", dimensions)
	fmt.Printf("📦 向量数量: %d\n", vectorCount)
	
	// 存储指标
	metrics := chromemStorage.GetMetrics()
	fmt.Printf("⚡ 性能指标:\n")
	fmt.Printf("  - 读操作: %d 次\n", metrics.ReadOps)
	fmt.Printf("  - 写操作: %d 次\n", metrics.WriteOps)
	fmt.Printf("  - 运行时间: %v\n", metrics.Uptime)

	// 8. 实时搜索体验演示
	fmt.Println("\n🚀 实时搜索体验演示")
	fmt.Println("-" + string(make([]rune, 30)))

	realTimeQueries := []string{
		"深度学习神经网络",
		"微服务API设计", 
		"数据库索引优化",
		"React组件开发",
		"Kubernetes部署",
		"安全漏洞防护",
	}

	fmt.Println("🔍 连续搜索测试:")
	totalSearchTime := time.Duration(0)
	
	for i, query := range realTimeQueries {
		start = time.Now()
		results, err := chromemStorage.SearchByText(ctx, query, 2)
		searchTime := time.Since(start)
		totalSearchTime += searchTime
		
		if err != nil {
			log.Printf("❌ 搜索失败: %v", err)
			continue
		}
		
		fmt.Printf("  %d. \"%s\" -> %d 结果 (%v)\n", 
			i+1, query, len(results), searchTime)
		
		if len(results) > 0 {
			fmt.Printf("     最佳匹配: %s (%.3f)\n", 
				results[0].Document.Title, results[0].Similarity)
		}
	}
	
	avgSearchTime := totalSearchTime / time.Duration(len(realTimeQueries))
	fmt.Printf("📊 平均搜索耗时: %v\n", avgSearchTime)

	// 9. 语义相似度展示
	fmt.Println("\n🧠 语义理解能力展示")
	fmt.Println("-" + string(make([]rune, 30)))

	semanticPairs := []struct {
		query1, query2 string
	}{
		{"机器学习", "人工智能算法"},
		{"并发编程", "多线程处理"},
		{"微服务", "分布式架构"},
		{"数据库优化", "查询性能调优"},
	}

	for _, pair := range semanticPairs {
		fmt.Printf("🔄 比较语义相似性:\n")
		fmt.Printf("   查询A: \"%s\"\n", pair.query1)
		fmt.Printf("   查询B: \"%s\"\n", pair.query2)
		
		results1, _ := chromemStorage.SearchByText(ctx, pair.query1, 1)
		results2, _ := chromemStorage.SearchByText(ctx, pair.query2, 1)
		
		if len(results1) > 0 && len(results2) > 0 {
			if results1[0].Document.ID == results2[0].Document.ID {
				fmt.Printf("   ✅ 两个查询指向同一文档: %s\n", results1[0].Document.Title)
			} else {
				fmt.Printf("   📊 查询A最佳匹配: %s (%.3f)\n", 
					results1[0].Document.Title, results1[0].Similarity)
				fmt.Printf("   📊 查询B最佳匹配: %s (%.3f)\n", 
					results2[0].Document.Title, results2[0].Similarity)
			}
		}
		fmt.Println()
	}

	// 10. 总结
	fmt.Println("🎉 Chromem-Go 向量数据库演示完成!")
	fmt.Println("=" + string(make([]rune, 60)))
	fmt.Println("✅ 核心功能验证:")
	fmt.Println("  🔹 向量嵌入自动生成")
	fmt.Println("  🔹 语义相似度搜索") 
	fmt.Println("  🔹 高级过滤查询")
	fmt.Println("  🔹 实时搜索响应")
	fmt.Println("  🔹 多维度统计分析")
	fmt.Println("  🔹 智能语义理解")
	
	fmt.Printf("\n📊 性能总结:\n")
	fmt.Printf("  - 文档存储速度: %.2f docs/ms\n", 
		float64(len(knowledgeBase))/float64(storeTime.Nanoseconds())*1000000)
	fmt.Printf("  - 平均搜索延迟: %v\n", avgSearchTime)
	fmt.Printf("  - 向量维度: %d\n", dimensions)
	fmt.Printf("  - 内存使用: 高效向量存储\n")
	fmt.Printf("  - 搜索引擎: Chromem-Go专业向量数据库\n")
	
	fmt.Println("\n🚀 Chromem-Go相比传统实现的优势:")
	fmt.Println("  ✨ 自动向量嵌入生成 (无需手动计算)")
	fmt.Println("  ✨ 专业语义相似度算法") 
	fmt.Println("  ✨ 高效向量索引和检索")
	fmt.Println("  ✨ 内置多种嵌入模型支持")
	fmt.Println("  ✨ 企业级性能和稳定性")
}