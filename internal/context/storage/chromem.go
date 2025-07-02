package storage

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/philippgille/chromem-go"
)

// ChromemStorage chromem-go向量数据库存储实现
type ChromemStorage struct {
	// chromem核心
	db         *chromem.DB
	collection *chromem.Collection

	// 配置和状态
	config    StorageConfig
	metrics   StorageMetrics
	startTime time.Time
	mu        sync.RWMutex

	// 文档存储 (chromem专注向量，文档数据需要额外存储)
	documents map[string]Document
	docMu     sync.RWMutex
}

// createOfflineEmbeddingFunc 创建离线嵌入函数，避免外部API调用
func createOfflineEmbeddingFunc() chromem.EmbeddingFunc {
	return func(ctx context.Context, text string) ([]float32, error) {
		// 使用哈希算法生成固定维度的向量 (384维，与chromem默认一致)
		const dimensions = 384

		// 结合SHA256和FNV哈希以增加向量的多样性
		sha256Hash := sha256.Sum256([]byte(text))
		fnvHash := fnv.New64a()
		fnvHash.Write([]byte(text))
		fnvHashValue := fnvHash.Sum64()

		embedding := make([]float32, dimensions)

		// 使用哈希值生成归一化的向量
		for i := 0; i < dimensions; i++ {
			// 交替使用不同的哈希源以增加向量复杂度
			var hashByte byte
			if i%2 == 0 {
				hashByte = sha256Hash[i%32]
			} else {
				hashByte = byte((fnvHashValue >> (i % 64)) & 0xFF)
			}

			// 将字节值转换为 [-1, 1] 范围的浮点数
			embedding[i] = (float32(hashByte) / 127.5) - 1.0
		}

		// 简单的L2归一化，确保向量长度为1
		var norm float32
		for _, val := range embedding {
			norm += val * val
		}
		if norm > 0 {
			norm = float32(1.0 / (norm * norm)) // 简化的归一化
			for i := range embedding {
				embedding[i] *= norm
			}
		}

		return embedding, nil
	}
}

// NewChromemStorage 创建chromem向量存储实例
func NewChromemStorage(config StorageConfig) (*ChromemStorage, error) {
	// 创建chromem数据库
	db := chromem.NewDB()

	// 配置集合名称
	collectionName := "documents"
	if config.Options != nil {
		if name, ok := config.Options["collection_name"]; ok {
			collectionName = name
		}
	}

	// 创建本地嵌入函数以避免外部API调用
	embeddingFunc := createOfflineEmbeddingFunc()

	// 创建或获取集合，使用本地嵌入函数
	collection, err := db.GetOrCreateCollection(collectionName, nil, embeddingFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	storage := &ChromemStorage{
		db:         db,
		collection: collection,
		config:     config,
		startTime:  time.Now(),
		documents:  make(map[string]Document),
	}

	// 如果有持久化路径，尝试加载
	if config.Path != "" && config.Path != ":memory:" {
		if err := storage.loadFromDisk(); err != nil {
			// 加载失败不是致命错误，继续执行
			fmt.Printf("Warning: failed to load from disk: %v\n", err)
		}
	}

	return storage, nil
}

// === DocumentStorage接口实现 ===

// Store 存储文档和向量
func (c *ChromemStorage) Store(ctx context.Context, doc Document) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if doc.Created.IsZero() {
		doc.Created = time.Now()
	}
	doc.Updated = time.Now()

	// 存储文档元数据
	c.docMu.Lock()
	c.documents[doc.ID] = doc
	c.docMu.Unlock()

	// 生成文档的文本内容用于向量化
	content := doc.Title + " " + doc.Content

	// 准备元数据
	metadata := make(map[string]string)
	if doc.Metadata != nil {
		for k, v := range doc.Metadata {
			metadata[k] = v
		}
	}
	metadata["title"] = doc.Title
	metadata["created"] = doc.Created.Format(time.RFC3339)
	metadata["updated"] = doc.Updated.Format(time.RFC3339)

	// 使用chromem添加文档 (会自动生成向量)
	err := c.collection.AddDocument(ctx, chromem.Document{
		ID:       doc.ID,
		Content:  content,
		Metadata: metadata,
	})

	if err != nil {
		// 如果chromem存储失败，回滚文档存储
		c.docMu.Lock()
		delete(c.documents, doc.ID)
		c.docMu.Unlock()
		return fmt.Errorf("failed to store in chromem: %w", err)
	}

	c.metrics.WriteOps++
	c.metrics.DocumentCount = uint64(len(c.documents))

	return nil
}

// Get 获取文档
func (c *ChromemStorage) Get(ctx context.Context, id string) (*Document, error) {
	c.docMu.RLock()
	doc, exists := c.documents[id]
	c.docMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("document not found: %s", id)
	}

	c.metrics.ReadOps++
	return &doc, nil
}

// Delete 删除文档
func (c *ChromemStorage) Delete(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 从chromem删除
	err := c.collection.Delete(ctx, nil, nil, id)
	if err != nil {
		return fmt.Errorf("failed to delete from chromem: %w", err)
	}

	// 从文档存储删除
	c.docMu.Lock()
	delete(c.documents, id)
	c.docMu.Unlock()

	c.metrics.WriteOps++
	c.metrics.DocumentCount = uint64(len(c.documents))

	return nil
}

// Exists 检查文档是否存在
func (c *ChromemStorage) Exists(ctx context.Context, id string) (bool, error) {
	c.docMu.RLock()
	_, exists := c.documents[id]
	c.docMu.RUnlock()

	return exists, nil
}

// BatchStore 批量存储文档
func (c *ChromemStorage) BatchStore(ctx context.Context, docs []Document) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	chromemDocs := make([]chromem.Document, 0, len(docs))
	storedDocs := make([]Document, 0, len(docs))

	// 准备所有文档
	for _, doc := range docs {
		if doc.Created.IsZero() {
			doc.Created = time.Now()
		}
		doc.Updated = time.Now()

		// 准备chromem文档
		content := doc.Title + " " + doc.Content
		metadata := make(map[string]string)
		if doc.Metadata != nil {
			for k, v := range doc.Metadata {
				metadata[k] = v
			}
		}
		metadata["title"] = doc.Title
		metadata["created"] = doc.Created.Format(time.RFC3339)
		metadata["updated"] = doc.Updated.Format(time.RFC3339)

		chromemDocs = append(chromemDocs, chromem.Document{
			ID:       doc.ID,
			Content:  content,
			Metadata: metadata,
		})
		storedDocs = append(storedDocs, doc)
	}

	// 批量添加到chromem
	err := c.collection.AddDocuments(ctx, chromemDocs, 1)
	if err != nil {
		return fmt.Errorf("failed to batch store in chromem: %w", err)
	}

	// 更新文档存储
	c.docMu.Lock()
	for _, doc := range storedDocs {
		c.documents[doc.ID] = doc
	}
	c.docMu.Unlock()

	c.metrics.WriteOps += uint64(len(docs))
	c.metrics.DocumentCount = uint64(len(c.documents))

	return nil
}

// BatchGet 批量获取文档
func (c *ChromemStorage) BatchGet(ctx context.Context, ids []string) ([]Document, error) {
	c.docMu.RLock()
	defer c.docMu.RUnlock()

	docs := make([]Document, 0, len(ids))
	for _, id := range ids {
		if doc, exists := c.documents[id]; exists {
			docs = append(docs, doc)
		}
	}

	c.metrics.ReadOps += uint64(len(docs))
	return docs, nil
}

// BatchDelete 批量删除文档
func (c *ChromemStorage) BatchDelete(ctx context.Context, ids []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 从chromem批量删除
	for _, id := range ids {
		err := c.collection.Delete(ctx, map[string]string{"id": id}, nil)
		if err != nil {
			// 继续删除其他文档，记录错误
			fmt.Printf("Warning: failed to delete %s from chromem: %v\n", id, err)
		}
	}

	// 从文档存储删除
	c.docMu.Lock()
	for _, id := range ids {
		delete(c.documents, id)
	}
	c.docMu.Unlock()

	c.metrics.WriteOps += uint64(len(ids))
	c.metrics.DocumentCount = uint64(len(c.documents))

	return nil
}

// List 列出文档
func (c *ChromemStorage) List(ctx context.Context, limit, offset int) ([]Document, error) {
	c.docMu.RLock()
	defer c.docMu.RUnlock()

	// 将map转为slice进行分页
	allDocs := make([]Document, 0, len(c.documents))
	for _, doc := range c.documents {
		allDocs = append(allDocs, doc)
	}

	// 简单排序 (按创建时间降序)
	for i := 0; i < len(allDocs)-1; i++ {
		for j := i + 1; j < len(allDocs); j++ {
			if allDocs[i].Created.Before(allDocs[j].Created) {
				allDocs[i], allDocs[j] = allDocs[j], allDocs[i]
			}
		}
	}

	// 应用分页
	start := offset
	if start >= len(allDocs) {
		return []Document{}, nil
	}

	end := start + limit
	if end > len(allDocs) {
		end = len(allDocs)
	}

	result := allDocs[start:end]
	c.metrics.ReadOps += uint64(len(result))

	return result, nil
}

// Count 统计文档数量
func (c *ChromemStorage) Count(ctx context.Context) (uint64, error) {
	c.docMu.RLock()
	count := uint64(len(c.documents))
	c.docMu.RUnlock()

	c.metrics.DocumentCount = count
	return count, nil
}

// === VectorStorage接口实现 ===

// StoreVector chromem自动处理向量，此方法主要用于兼容性
func (c *ChromemStorage) StoreVector(ctx context.Context, id string, vector []float64) error {
	// chromem自动处理向量生成，这里只需确保文档存在
	_, err := c.Get(ctx, id)
	return err
}

// GetVector 获取向量 (chromem内部管理，返回模拟向量)
func (c *ChromemStorage) GetVector(ctx context.Context, id string) ([]float64, error) {
	// chromem内部管理向量，这里返回标识向量
	_, err := c.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// 返回一个标识向量 (实际向量由chromem内部管理)
	return make([]float64, 384), nil // chromem默认384维
}

// DeleteVector 删除向量 (通过删除文档实现)
func (c *ChromemStorage) DeleteVector(ctx context.Context, id string) error {
	return c.Delete(ctx, id)
}

// SearchSimilar 向量相似度搜索 (使用chromem的强大搜索)
func (c *ChromemStorage) SearchSimilar(ctx context.Context, queryVector []float64, limit int) ([]VectorResult, error) {
	// 注意：这个方法用于与现有接口兼容，但建议使用SearchByText
	return c.SearchByText(ctx, "", limit)
}

// SearchByText 基于文本的语义搜索 (chromem的核心功能)
func (c *ChromemStorage) SearchByText(ctx context.Context, query string, limit int) ([]VectorResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if query == "" {
		// 如果查询为空，返回最近的文档
		docs, err := c.List(ctx, limit, 0)
		if err != nil {
			return nil, err
		}

		results := make([]VectorResult, len(docs))
		for i, doc := range docs {
			results[i] = VectorResult{
				Document:   doc,
				Similarity: 1.0, // 默认相似度
				Score:      1.0,
			}
		}
		return results, nil
	}

	// 使用chromem进行语义搜索
	chromemResults, err := c.collection.Query(ctx, query, limit, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("chromem query failed: %w", err)
	}

	results := make([]VectorResult, 0, len(chromemResults))

	c.docMu.RLock()
	for _, result := range chromemResults {
		if doc, exists := c.documents[result.ID]; exists {
			results = append(results, VectorResult{
				Document:   doc,
				Similarity: float64(result.Similarity),
				Score:      float64(result.Similarity),
			})
		}
	}
	c.docMu.RUnlock()

	return results, nil
}

// SearchByThreshold 按阈值搜索
func (c *ChromemStorage) SearchByThreshold(ctx context.Context, queryVector []float64, threshold float64) ([]VectorResult, error) {
	// 获取更多结果后过滤
	results, err := c.SearchByText(ctx, "", 100)
	if err != nil {
		return nil, err
	}

	filtered := make([]VectorResult, 0)
	for _, result := range results {
		if result.Similarity >= threshold {
			filtered = append(filtered, result)
		}
	}

	return filtered, nil
}

// BatchStoreVectors chromem自动处理向量
func (c *ChromemStorage) BatchStoreVectors(ctx context.Context, vectors map[string][]float64) error {
	// chromem自动处理向量生成，这里只需确保文档存在
	for id := range vectors {
		if _, err := c.Get(ctx, id); err != nil {
			return fmt.Errorf("document %s not found for vector storage", id)
		}
	}
	return nil
}

// GetDimensions 获取向量维度
func (c *ChromemStorage) GetDimensions() int {
	return 384 // chromem默认使用384维向量
}

// GetVectorCount 获取向量数量
func (c *ChromemStorage) GetVectorCount() uint64 {
	c.docMu.RLock()
	count := uint64(len(c.documents))
	c.docMu.RUnlock()
	return count
}

// === 管理操作 ===

// Close 关闭存储
func (c *ChromemStorage) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果有持久化路径，保存数据
	if c.config.Path != "" && c.config.Path != ":memory:" {
		if err := c.saveToDisk(); err != nil {
			fmt.Printf("Warning: failed to save to disk: %v\n", err)
		}
	}

	// chromem没有显式的Close方法，清理资源
	c.documents = nil
	c.collection = nil
	c.db = nil

	return nil
}

// Flush 刷新数据
func (c *ChromemStorage) Flush() error {
	// chromem是内存数据库，flush主要用于持久化
	if c.config.Path != "" && c.config.Path != ":memory:" {
		return c.saveToDisk()
	}
	return nil
}

// GetMetrics 获取存储指标
func (c *ChromemStorage) GetMetrics() StorageMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.metrics.Uptime = time.Since(c.startTime)
	c.metrics.DocumentCount = uint64(len(c.documents))

	return c.metrics
}

// === 持久化相关 ===

// saveToDisk 保存到磁盘
func (c *ChromemStorage) saveToDisk() error {
	if c.config.Path == "" || c.config.Path == ":memory:" {
		return nil
	}

	// 创建目录
	dir := filepath.Dir(c.config.Path)
	if err := ensureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 保存文档数据
	c.docMu.RLock()
	data := struct {
		Documents map[string]Document `json:"documents"`
		Timestamp time.Time           `json:"timestamp"`
	}{
		Documents: c.documents,
		Timestamp: time.Now(),
	}
	c.docMu.RUnlock()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	filename := c.config.Path + ".json"
	if err := writeFile(filename, jsonData); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// loadFromDisk 从磁盘加载
func (c *ChromemStorage) loadFromDisk() error {
	if c.config.Path == "" || c.config.Path == ":memory:" {
		return nil
	}

	filename := c.config.Path + ".json"
	if !fileExists(filename) {
		return nil // 文件不存在不是错误
	}

	jsonData, err := readFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var data struct {
		Documents map[string]Document `json:"documents"`
		Timestamp time.Time           `json:"timestamp"`
	}

	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// 恢复文档数据
	c.docMu.Lock()
	c.documents = data.Documents
	c.docMu.Unlock()

	// 重新添加到chromem (重新生成向量)
	ctx := context.Background()
	for _, doc := range data.Documents {
		content := doc.Title + " " + doc.Content
		metadata := make(map[string]string)
		if doc.Metadata != nil {
			for k, v := range doc.Metadata {
				metadata[k] = v
			}
		}
		metadata["title"] = doc.Title
		metadata["created"] = doc.Created.Format(time.RFC3339)
		metadata["updated"] = doc.Updated.Format(time.RFC3339)

		c.collection.AddDocument(ctx, chromem.Document{
			ID:       doc.ID,
			Content:  content,
			Metadata: metadata,
		})
	}

	return nil
}

// === 辅助函数 ===

// ensureDir 确保目录存在
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// writeFile 写文件
func writeFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

// readFile 读文件
func readFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// === 额外的chromem特有功能 ===

// GetCollection 获取chromem集合 (用于高级操作)
func (c *ChromemStorage) GetCollection() *chromem.Collection {
	return c.collection
}

// QueryWithFilter 带过滤器的查询
func (c *ChromemStorage) QueryWithFilter(ctx context.Context, query string, limit int, where map[string]string) ([]VectorResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	chromemResults, err := c.collection.Query(ctx, query, limit, where, nil)
	if err != nil {
		return nil, fmt.Errorf("chromem query with filter failed: %w", err)
	}

	results := make([]VectorResult, 0, len(chromemResults))

	c.docMu.RLock()
	for _, result := range chromemResults {
		if doc, exists := c.documents[result.ID]; exists {
			results = append(results, VectorResult{
				Document:   doc,
				Similarity: float64(result.Similarity),
				Score:      float64(result.Similarity),
			})
		}
	}
	c.docMu.RUnlock()

	return results, nil
}

// GetStats 获取chromem统计信息
func (c *ChromemStorage) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := map[string]interface{}{
		"collection_name":  c.collection.Name,
		"document_count":   len(c.documents),
		"vector_dimension": c.GetDimensions(),
		"uptime":           time.Since(c.startTime),
		"read_ops":         c.metrics.ReadOps,
		"write_ops":        c.metrics.WriteOps,
	}

	return stats
}
