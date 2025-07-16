package context

import (
	"context"
	"time"
)

// === 核心类型定义 ===

// Document 文档结构
type Document struct {
	ID      string            `json:"id"`
	Title   string            `json:"title"`
	Content string            `json:"content"`
	Created time.Time         `json:"created"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// VectorResult 向量检索结果
type VectorResult struct {
	Document   Document `json:"document"`
	Similarity float64  `json:"similarity"`
}

// Context 上下文结果
type Context struct {
	ID      string    `json:"id"`
	Task    string    `json:"task"`
	Content string    `json:"content"`
	Quality float64   `json:"quality"`
	Created time.Time `json:"created"`
}

// CacheStats 缓存统计
type CacheStats struct {
	HitCount   int64   `json:"hit_count"`
	MissCount  int64   `json:"miss_count"`
	EvictCount int64   `json:"evict_count"`
	ItemCount  int     `json:"item_count"`
	HitRatio   float64 `json:"hit_ratio"`
}

// EngineStats 引擎统计
type EngineStats struct {
	DocumentCount int           `json:"document_count"`
	CacheStats    CacheStats    `json:"cache_stats"`
	VectorLayerOK bool          `json:"vector_layer_ok"`
	IndexLayerOK  bool          `json:"index_layer_ok"`
	MemoryLayerOK bool          `json:"memory_layer_ok"`
	LastQueryTime time.Duration `json:"last_query_time"`
	TotalQueries  int64         `json:"total_queries"`
}

// === 接口定义 ===

// Engine 上下文引擎接口
type Engine interface {
	// 核心功能
	BuildContext(ctx context.Context, task, input string) (*Context, error)
	AddDocument(doc Document) error
	SearchSimilar(query string, limit int) ([]VectorResult, error)

	// 文档管理
	GetDocument(id string) (*Document, error)
	RemoveDocument(id string) error

	// 生命周期
	Close() error
	Stats() EngineStats
}
