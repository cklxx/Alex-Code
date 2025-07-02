package context

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"alex/internal/context/algorithms"
)

// === 核心类型定义 ===

// UnifiedEngine 统一的上下文引擎
type UnifiedEngine struct {
	// 存储
	documents     map[string]Document
	vectors       map[string][]float64
	invertedIndex map[string][]string

	// 缓存
	queryCache   map[string][]VectorResult
	contextCache map[string]*Context

	// 配置
	config *EngineConfig

	// 统计
	stats     EngineStats
	startTime time.Time

	// 并发控制
	mu     sync.RWMutex
	closed bool
}

// EngineConfig 引擎配置
type EngineConfig struct {
	CacheSize          int                            `json:"cache_size"`
	MaxContextLen      int                            `json:"max_context_len"`
	CompressRatio      float64                        `json:"compress_ratio"`
	EmbeddingConfig    *algorithms.EmbeddingConfig    `json:"embedding_config"`
	RelevanceThreshold *algorithms.RelevanceThreshold `json:"relevance_threshold"`
}

// DefaultEngineConfig 默认引擎配置
func DefaultEngineConfig() *EngineConfig {
	return &EngineConfig{
		CacheSize:          1000,
		MaxContextLen:      1000,
		CompressRatio:      0.7,
		EmbeddingConfig:    algorithms.DefaultEmbeddingConfig(),
		RelevanceThreshold: algorithms.DefaultRelevanceThreshold(),
	}
}

// NewUnifiedEngine 创建统一引擎
func NewUnifiedEngine(config *EngineConfig) *UnifiedEngine {
	if config == nil {
		config = DefaultEngineConfig()
	}

	return &UnifiedEngine{
		documents:     make(map[string]Document),
		vectors:       make(map[string][]float64),
		invertedIndex: make(map[string][]string),
		queryCache:    make(map[string][]VectorResult),
		contextCache:  make(map[string]*Context),
		config:        config,
		startTime:     time.Now(),
		stats: EngineStats{
			VectorLayerOK: true,
			IndexLayerOK:  true,
			MemoryLayerOK: true,
		},
	}
}

// === 核心接口实现 ===

// BuildContext 构建上下文
func (ue *UnifiedEngine) BuildContext(ctx context.Context, task, input string) (*Context, error) {
	startTime := time.Now()
	defer func() {
		ue.updateStats(time.Since(startTime))
	}()

	ue.mu.Lock()
	defer ue.mu.Unlock()

	if ue.closed {
		return nil, fmt.Errorf("engine is closed")
	}

	// 生成ID
	id := fmt.Sprintf("ctx_%d", time.Now().Unix())

	// 搜索相关信息
	relevantResults, _ := ue.searchSimilarInternal(input, 3)

	// 构建内容
	content := ue.buildContent(task, input, relevantResults)

	// 压缩优化
	if len(content) > ue.config.MaxContextLen {
		content = algorithms.CompressText(content, ue.config.CompressRatio)
	}

	// 计算质量
	qualityMetrics := algorithms.CalculateQuality(content, len(relevantResults))

	// 创建上下文
	result := &Context{
		ID:      id,
		Task:    task,
		Content: content,
		Quality: qualityMetrics.FinalScore,
		Created: time.Now(),
	}

	// 缓存结果
	ue.contextCache[id] = result
	ue.manageCacheSize()

	return result, nil
}

// AddDocument 添加文档
func (ue *UnifiedEngine) AddDocument(doc Document) error {
	ue.mu.Lock()
	defer ue.mu.Unlock()

	if ue.closed {
		return fmt.Errorf("engine is closed")
	}

	// 设置时间戳
	if doc.Created.IsZero() {
		doc.Created = time.Now()
	}

	// 存储文档
	ue.documents[doc.ID] = doc

	// 建立倒排索引
	ue.buildInvertedIndex(doc)

	// 生成向量表示
	text := doc.Content + " " + doc.Title
	ue.vectors[doc.ID] = algorithms.GenerateEmbedding(text, ue.config.EmbeddingConfig)

	// 更新统计
	ue.stats.DocumentCount = len(ue.documents)

	return nil
}

// SearchSimilar 混合搜索
func (ue *UnifiedEngine) SearchSimilar(query string, limit int) ([]VectorResult, error) {
	ue.mu.RLock()
	defer ue.mu.RUnlock()

	if ue.closed {
		return nil, fmt.Errorf("engine is closed")
	}

	return ue.searchSimilarInternal(query, limit)
}

// GetDocument 获取文档
func (ue *UnifiedEngine) GetDocument(id string) (*Document, error) {
	ue.mu.RLock()
	defer ue.mu.RUnlock()

	if ue.closed {
		return nil, fmt.Errorf("engine is closed")
	}

	doc, exists := ue.documents[id]
	if !exists {
		return nil, fmt.Errorf("document not found: %s", id)
	}

	return &doc, nil
}

// RemoveDocument 删除文档
func (ue *UnifiedEngine) RemoveDocument(id string) error {
	ue.mu.Lock()
	defer ue.mu.Unlock()

	if ue.closed {
		return fmt.Errorf("engine is closed")
	}

	// 检查文档是否存在
	if _, exists := ue.documents[id]; !exists {
		return fmt.Errorf("document not found: %s", id)
	}

	// 删除文档
	delete(ue.documents, id)
	delete(ue.vectors, id)

	// 清理倒排索引
	ue.cleanupInvertedIndex(id)

	// 更新统计
	ue.stats.DocumentCount = len(ue.documents)

	return nil
}

// Close 关闭引擎
func (ue *UnifiedEngine) Close() error {
	ue.mu.Lock()
	defer ue.mu.Unlock()

	if ue.closed {
		return nil
	}

	ue.closed = true

	// 清理资源
	ue.documents = nil
	ue.vectors = nil
	ue.invertedIndex = nil
	ue.queryCache = nil
	ue.contextCache = nil

	return nil
}

// Stats 获取统计信息
func (ue *UnifiedEngine) Stats() EngineStats {
	ue.mu.RLock()
	defer ue.mu.RUnlock()

	stats := ue.stats
	stats.DocumentCount = len(ue.documents)

	// 更新缓存统计
	stats.CacheStats.ItemCount = len(ue.queryCache)
	total := stats.CacheStats.HitCount + stats.CacheStats.MissCount
	if total > 0 {
		stats.CacheStats.HitRatio = float64(stats.CacheStats.HitCount) / float64(total)
	}

	return stats
}

// === 内部实现方法 ===

// searchSimilarInternal 内部搜索实现
func (ue *UnifiedEngine) searchSimilarInternal(query string, limit int) ([]VectorResult, error) {
	startTime := time.Now()
	defer func() {
		ue.updateStats(time.Since(startTime))
	}()

	// 检查缓存
	if cached, found := ue.queryCache[query]; found {
		ue.stats.CacheStats.HitCount++
		if len(cached) > limit {
			return cached[:limit], nil
		}
		return cached, nil
	}
	ue.stats.CacheStats.MissCount++

	// 分词
	queryTokens := algorithms.Tokenize(query)
	if len(queryTokens.Words) == 0 {
		return []VectorResult{}, nil
	}

	// 生成查询向量
	queryVector := algorithms.GenerateEmbedding(query, ue.config.EmbeddingConfig)

	// 1. 倒排索引搜索
	indexResults := ue.searchByIndex(queryTokens.Words)

	// 2. 向量相似度搜索
	vectorResults := ue.searchByVector(queryVector)

	// 3. 混合评分
	hybridResults := ue.combineResults(indexResults, vectorResults, query)

	// 4. 排序并限制结果数量
	sort.Slice(hybridResults, func(i, j int) bool {
		return hybridResults[i].Similarity > hybridResults[j].Similarity
	})

	if len(hybridResults) > limit {
		hybridResults = hybridResults[:limit]
	}

	// 缓存结果
	ue.queryCache[query] = hybridResults
	ue.manageCacheSize()

	return hybridResults, nil
}

// buildInvertedIndex 建立倒排索引
func (ue *UnifiedEngine) buildInvertedIndex(doc Document) {
	tokens := algorithms.Tokenize(doc.Content + " " + doc.Title)

	for _, word := range tokens.Words {
		// 检查该词是否已存在该文档ID
		exists := false
		for _, existingID := range ue.invertedIndex[word] {
			if existingID == doc.ID {
				exists = true
				break
			}
		}

		if !exists {
			ue.invertedIndex[word] = append(ue.invertedIndex[word], doc.ID)
		}
	}
}

// cleanupInvertedIndex 清理倒排索引
func (ue *UnifiedEngine) cleanupInvertedIndex(docID string) {
	for word, docIDs := range ue.invertedIndex {
		for i, id := range docIDs {
			if id == docID {
				// 删除该文档ID
				ue.invertedIndex[word] = append(docIDs[:i], docIDs[i+1:]...)
				break
			}
		}

		// 如果该词没有文档了，删除词条
		if len(ue.invertedIndex[word]) == 0 {
			delete(ue.invertedIndex, word)
		}
	}
}

// searchByIndex 基于倒排索引搜索
func (ue *UnifiedEngine) searchByIndex(queryWords []string) map[string]float64 {
	results := make(map[string]float64)

	for _, word := range queryWords {
		if docIDs, exists := ue.invertedIndex[word]; exists {
			// TF-IDF权重
			idf := 1.0 / float64(len(docIDs))
			for _, docID := range docIDs {
				results[docID] += idf
			}
		}
	}

	return results
}

// searchByVector 基于向量相似度搜索
func (ue *UnifiedEngine) searchByVector(queryVector []float64) map[string]float64 {
	results := make(map[string]float64)

	for docID, docVector := range ue.vectors {
		similarity := algorithms.CosineSimilarity(queryVector, docVector)
		if similarity > ue.config.RelevanceThreshold.MinScore {
			results[docID] = similarity
		}
	}

	return results
}

// combineResults 混合评分
func (ue *UnifiedEngine) combineResults(indexResults, vectorResults map[string]float64, query string) []VectorResult {
	var results []VectorResult

	// 收集所有候选文档
	candidates := make(map[string]bool)
	for docID := range indexResults {
		candidates[docID] = true
	}
	for docID := range vectorResults {
		candidates[docID] = true
	}

	for docID := range candidates {
		doc := ue.documents[docID]

		indexScore := indexResults[docID]
		vectorScore := vectorResults[docID]
		textScore := algorithms.CalculateTextMatch(query, doc.Content)

		// 使用算法库计算混合评分
		scoreResult := algorithms.CalculateHybridScore(indexScore, vectorScore, textScore)

		if scoreResult.FinalScore > ue.config.RelevanceThreshold.MinScore {
			results = append(results, VectorResult{
				Document:   doc,
				Similarity: scoreResult.FinalScore,
			})
		}
	}

	return results
}

// buildContent 构建上下文内容
func (ue *UnifiedEngine) buildContent(task, input string, relevantInfo []VectorResult) string {
	var parts []string

	if task != "" {
		parts = append(parts, fmt.Sprintf("Task: %s", task))
	}

	if input != "" {
		parts = append(parts, fmt.Sprintf("Input: %s", input))
	}

	if len(relevantInfo) > 0 {
		var contextItems []string
		for _, result := range relevantInfo {
			contextItems = append(contextItems, fmt.Sprintf("%s (%.2f): %s",
				result.Document.Title, result.Similarity, result.Document.Content))
		}
		parts = append(parts, fmt.Sprintf("Context: %s", strings.Join(contextItems, "; ")))
	}

	return strings.Join(parts, "\n")
}

// manageCacheSize 管理缓存大小
func (ue *UnifiedEngine) manageCacheSize() {
	// 简单的LRU策略：当缓存超过限制时删除最旧的一半
	if len(ue.queryCache) > ue.config.CacheSize {
		// 删除一半缓存
		count := 0
		targetCount := ue.config.CacheSize / 2
		for key := range ue.queryCache {
			delete(ue.queryCache, key)
			count++
			if count >= targetCount {
				break
			}
		}
		ue.stats.CacheStats.EvictCount += int64(count)
	}

	// 管理上下文缓存
	if len(ue.contextCache) > ue.config.CacheSize {
		count := 0
		targetCount := ue.config.CacheSize / 2
		for key := range ue.contextCache {
			delete(ue.contextCache, key)
			count++
			if count >= targetCount {
				break
			}
		}
	}
}

// updateStats 更新统计信息
func (ue *UnifiedEngine) updateStats(queryTime time.Duration) {
	ue.stats.LastQueryTime = queryTime
	ue.stats.TotalQueries++
}

// === 向后兼容接口 ===

// NewEngine 创建引擎（向后兼容）
func NewEngine() *UnifiedEngine {
	return NewUnifiedEngine(DefaultEngineConfig())
}
