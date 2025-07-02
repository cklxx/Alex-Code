package storage

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// MemoryStorageEngine 内存存储引擎
type MemoryStorageEngine struct {
	config StorageConfig

	// 存储组件
	docStorage    *MemoryDocumentStorage
	vectorStorage *MemoryVectorStorage

	// 生命周期
	startTime time.Time
	closed    bool
	mu        sync.RWMutex
}

// NewMemoryStorageEngine 创建内存存储引擎
func NewMemoryStorageEngine() *MemoryStorageEngine {
	return &MemoryStorageEngine{
		docStorage:    NewMemoryDocumentStorage(),
		vectorStorage: NewMemoryVectorStorage(),
		startTime:     time.Now(),
	}
}

// Initialize 初始化存储引擎
func (mse *MemoryStorageEngine) Initialize(config StorageConfig) error {
	mse.mu.Lock()
	defer mse.mu.Unlock()

	mse.config = config
	return nil
}

// DocumentStorage 获取文档存储
func (mse *MemoryStorageEngine) DocumentStorage() DocumentStorage {
	return mse.docStorage
}

// VectorStorage 获取向量存储
func (mse *MemoryStorageEngine) VectorStorage() VectorStorage {
	return mse.vectorStorage
}

// IndexStorage 获取索引存储 (简化实现，返回nil)
func (mse *MemoryStorageEngine) IndexStorage() IndexStorage {
	return &DummyIndexStorage{}
}

// Close 关闭存储引擎
func (mse *MemoryStorageEngine) Close() error {
	mse.mu.Lock()
	defer mse.mu.Unlock()

	if mse.closed {
		return nil
	}

	mse.closed = true

	// 关闭各个存储组件
	if err := mse.docStorage.Close(); err != nil {
		return err
	}
	if err := mse.vectorStorage.Close(); err != nil {
		return err
	}
	// 索引存储已简化，无需关闭

	return nil
}

// Health 健康检查
func (mse *MemoryStorageEngine) Health() error {
	mse.mu.RLock()
	defer mse.mu.RUnlock()

	if mse.closed {
		return fmt.Errorf("storage engine is closed")
	}

	return nil
}

// GetMetrics 获取存储指标
func (mse *MemoryStorageEngine) GetMetrics() StorageMetrics {
	docMetrics := mse.docStorage.GetMetrics()

	return StorageMetrics{
		DocumentCount: docMetrics.DocumentCount,
		StorageSize:   docMetrics.StorageSize,
		CacheHits:     docMetrics.CacheHits,
		CacheMisses:   docMetrics.CacheMisses,
		ReadOps:       docMetrics.ReadOps,
		WriteOps:      docMetrics.WriteOps,
		LastSync:      docMetrics.LastSync,
		Uptime:        time.Since(mse.startTime),
	}
}

// MemoryDocumentStorage 内存文档存储实现
type MemoryDocumentStorage struct {
	documents map[string]Document
	metrics   StorageMetrics
	mu        sync.RWMutex
}

// NewMemoryDocumentStorage 创建内存文档存储
func NewMemoryDocumentStorage() *MemoryDocumentStorage {
	return &MemoryDocumentStorage{
		documents: make(map[string]Document),
		metrics:   StorageMetrics{LastSync: time.Now()},
	}
}

// Store 存储文档
func (mds *MemoryDocumentStorage) Store(ctx context.Context, doc Document) error {
	mds.mu.Lock()
	defer mds.mu.Unlock()

	doc.Updated = time.Now()
	if doc.Created.IsZero() {
		doc.Created = doc.Updated
	}

	// 检查是否是新文档
	isNew := true
	if _, exists := mds.documents[doc.ID]; exists {
		isNew = false
	}

	mds.documents[doc.ID] = doc

	// 更新指标
	mds.metrics.WriteOps++
	if isNew {
		mds.metrics.DocumentCount++
	}

	// 简单的大小估算
	mds.updateStorageSize()

	return nil
}

// Get 获取文档
func (mds *MemoryDocumentStorage) Get(ctx context.Context, id string) (*Document, error) {
	mds.mu.RLock()
	defer mds.mu.RUnlock()

	doc, exists := mds.documents[id]
	mds.metrics.ReadOps++

	if !exists {
		mds.metrics.CacheMisses++
		return nil, fmt.Errorf("document not found: %s", id)
	}

	mds.metrics.CacheHits++
	return &doc, nil
}

// Delete 删除文档
func (mds *MemoryDocumentStorage) Delete(ctx context.Context, id string) error {
	mds.mu.Lock()
	defer mds.mu.Unlock()

	if _, exists := mds.documents[id]; !exists {
		return fmt.Errorf("document not found: %s", id)
	}

	delete(mds.documents, id)

	// 更新指标
	mds.metrics.WriteOps++
	mds.metrics.DocumentCount--
	mds.updateStorageSize()

	return nil
}

// Exists 检查文档是否存在
func (mds *MemoryDocumentStorage) Exists(ctx context.Context, id string) (bool, error) {
	mds.mu.RLock()
	defer mds.mu.RUnlock()

	_, exists := mds.documents[id]
	mds.metrics.ReadOps++

	return exists, nil
}

// BatchStore 批量存储文档
func (mds *MemoryDocumentStorage) BatchStore(ctx context.Context, docs []Document) error {
	for _, doc := range docs {
		if err := mds.Store(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

// BatchGet 批量获取文档
func (mds *MemoryDocumentStorage) BatchGet(ctx context.Context, ids []string) ([]Document, error) {
	var docs []Document

	for _, id := range ids {
		if doc, err := mds.Get(ctx, id); err == nil {
			docs = append(docs, *doc)
		}
	}

	return docs, nil
}

// BatchDelete 批量删除文档
func (mds *MemoryDocumentStorage) BatchDelete(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := mds.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// List 列出文档
func (mds *MemoryDocumentStorage) List(ctx context.Context, limit, offset int) ([]Document, error) {
	mds.mu.RLock()
	defer mds.mu.RUnlock()

	var docs []Document
	i := 0

	for _, doc := range mds.documents {
		if i < offset {
			i++
			continue
		}

		if len(docs) >= limit {
			break
		}

		docs = append(docs, doc)
		i++
	}

	return docs, nil
}

// Count 获取文档数量
func (mds *MemoryDocumentStorage) Count(ctx context.Context) (uint64, error) {
	mds.mu.RLock()
	defer mds.mu.RUnlock()

	return uint64(len(mds.documents)), nil
}

// Close 关闭存储
func (mds *MemoryDocumentStorage) Close() error {
	mds.mu.Lock()
	defer mds.mu.Unlock()

	mds.documents = nil
	return nil
}

// Flush 刷新数据
func (mds *MemoryDocumentStorage) Flush() error {
	mds.mu.Lock()
	defer mds.mu.Unlock()

	mds.metrics.LastSync = time.Now()
	return nil
}

// GetMetrics 获取指标
func (mds *MemoryDocumentStorage) GetMetrics() StorageMetrics {
	mds.mu.RLock()
	defer mds.mu.RUnlock()

	return mds.metrics
}

// updateStorageSize 更新存储大小估算
func (mds *MemoryDocumentStorage) updateStorageSize() {
	size := uint64(0)
	for _, doc := range mds.documents {
		size += uint64(len(doc.ID) + len(doc.Title) + len(doc.Content))
		for k, v := range doc.Metadata {
			size += uint64(len(k) + len(v))
		}
	}
	mds.metrics.StorageSize = size
}

// MemoryVectorStorage 内存向量存储实现
type MemoryVectorStorage struct {
	vectors    map[string][]float64
	dimensions int
	mu         sync.RWMutex
}

// NewMemoryVectorStorage 创建内存向量存储
func NewMemoryVectorStorage() *MemoryVectorStorage {
	return &MemoryVectorStorage{
		vectors: make(map[string][]float64),
	}
}

// StoreVector 存储向量
func (mvs *MemoryVectorStorage) StoreVector(ctx context.Context, id string, vector []float64) error {
	mvs.mu.Lock()
	defer mvs.mu.Unlock()

	// 设置维度
	if mvs.dimensions == 0 && len(vector) > 0 {
		mvs.dimensions = len(vector)
	}

	// 验证维度
	if len(vector) != mvs.dimensions {
		return fmt.Errorf("vector dimension mismatch: expected %d, got %d", mvs.dimensions, len(vector))
	}

	// 复制向量以避免外部修改
	vectorCopy := make([]float64, len(vector))
	copy(vectorCopy, vector)

	mvs.vectors[id] = vectorCopy
	return nil
}

// GetVector 获取向量
func (mvs *MemoryVectorStorage) GetVector(ctx context.Context, id string) ([]float64, error) {
	mvs.mu.RLock()
	defer mvs.mu.RUnlock()

	vector, exists := mvs.vectors[id]
	if !exists {
		return nil, fmt.Errorf("vector not found: %s", id)
	}

	// 返回副本
	result := make([]float64, len(vector))
	copy(result, vector)

	return result, nil
}

// DeleteVector 删除向量
func (mvs *MemoryVectorStorage) DeleteVector(ctx context.Context, id string) error {
	mvs.mu.Lock()
	defer mvs.mu.Unlock()

	if _, exists := mvs.vectors[id]; !exists {
		return fmt.Errorf("vector not found: %s", id)
	}

	delete(mvs.vectors, id)
	return nil
}

// SearchSimilar 搜索相似向量
func (mvs *MemoryVectorStorage) SearchSimilar(ctx context.Context, vector []float64, limit int) ([]VectorResult, error) {
	mvs.mu.RLock()
	defer mvs.mu.RUnlock()

	if len(vector) != mvs.dimensions {
		return nil, fmt.Errorf("vector dimension mismatch: expected %d, got %d", mvs.dimensions, len(vector))
	}

	var results []VectorResult

	for id, storedVector := range mvs.vectors {
		similarity := cosineSimilarity(vector, storedVector)

		results = append(results, VectorResult{
			Document:   Document{ID: id},
			Similarity: similarity,
			Score:      similarity,
		})
	}

	// 按相似度排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// 限制结果数量
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// SearchByThreshold 按阈值搜索
func (mvs *MemoryVectorStorage) SearchByThreshold(ctx context.Context, vector []float64, threshold float64) ([]VectorResult, error) {
	mvs.mu.RLock()
	defer mvs.mu.RUnlock()

	if len(vector) != mvs.dimensions {
		return nil, fmt.Errorf("vector dimension mismatch: expected %d, got %d", mvs.dimensions, len(vector))
	}

	var results []VectorResult

	for id, storedVector := range mvs.vectors {
		similarity := cosineSimilarity(vector, storedVector)

		if similarity >= threshold {
			results = append(results, VectorResult{
				Document:   Document{ID: id},
				Similarity: similarity,
				Score:      similarity,
			})
		}
	}

	// 按相似度排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	return results, nil
}

// BatchStoreVectors 批量存储向量
func (mvs *MemoryVectorStorage) BatchStoreVectors(ctx context.Context, vectors map[string][]float64) error {
	for id, vector := range vectors {
		if err := mvs.StoreVector(ctx, id, vector); err != nil {
			return err
		}
	}
	return nil
}

// GetDimensions 获取向量维度
func (mvs *MemoryVectorStorage) GetDimensions() int {
	mvs.mu.RLock()
	defer mvs.mu.RUnlock()

	return mvs.dimensions
}

// GetVectorCount 获取向量数量
func (mvs *MemoryVectorStorage) GetVectorCount() uint64 {
	mvs.mu.RLock()
	defer mvs.mu.RUnlock()

	return uint64(len(mvs.vectors))
}

// Close 关闭向量存储
func (mvs *MemoryVectorStorage) Close() error {
	mvs.mu.Lock()
	defer mvs.mu.Unlock()

	mvs.vectors = nil
	return nil
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	return dotProduct / (normA*normB + 1e-10)
}

// DummyIndexStorage 简化的索引存储实现（主引擎已有索引功能）
type DummyIndexStorage struct{}

func (dis *DummyIndexStorage) AddDocument(ctx context.Context, doc Document) error {
	return nil // 主引擎处理索引
}

func (dis *DummyIndexStorage) RemoveDocument(ctx context.Context, id string) error {
	return nil // 主引擎处理索引
}

func (dis *DummyIndexStorage) UpdateDocument(ctx context.Context, doc Document) error {
	return nil // 主引擎处理索引
}

func (dis *DummyIndexStorage) Search(ctx context.Context, query string, limit int) ([]string, error) {
	return []string{}, nil // 主引擎处理搜索
}

func (dis *DummyIndexStorage) SearchTerms(ctx context.Context, terms []string, limit int) ([]string, error) {
	return []string{}, nil // 主引擎处理搜索
}

func (dis *DummyIndexStorage) GetTermFrequency(ctx context.Context, term, docID string) (uint32, error) {
	return 0, nil
}

func (dis *DummyIndexStorage) GetDocumentFrequency(ctx context.Context, term string) (uint32, error) {
	return 0, nil
}

func (dis *DummyIndexStorage) GetTermCount() uint64 {
	return 0
}

func (dis *DummyIndexStorage) GetDocumentCount() uint64 {
	return 0
}

func (dis *DummyIndexStorage) Optimize() error {
	return nil
}

func (dis *DummyIndexStorage) Close() error {
	return nil
}
