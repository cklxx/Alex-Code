package storage

import (
	"context"
	"time"
)

// Document 文档结构
type Document struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Created  time.Time         `json:"created"`
	Updated  time.Time         `json:"updated"`
}

// VectorResult 向量搜索结果
type VectorResult struct {
	Document   Document `json:"document"`
	Similarity float64  `json:"similarity"`
	Score      float64  `json:"score"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type         string            `json:"type"`          // "memory", "file", "badger"
	Path         string            `json:"path"`          // 存储路径
	MaxSize      int64             `json:"max_size"`      // 最大存储大小
	CacheSize    int               `json:"cache_size"`    // 缓存大小
	SyncInterval time.Duration     `json:"sync_interval"` // 同步间隔
	Options      map[string]string `json:"options"`       // 扩展选项
}

// StorageMetrics 存储指标
type StorageMetrics struct {
	DocumentCount uint64        `json:"document_count"`
	StorageSize   uint64        `json:"storage_size"`
	CacheHits     uint64        `json:"cache_hits"`
	CacheMisses   uint64        `json:"cache_misses"`
	ReadOps       uint64        `json:"read_ops"`
	WriteOps      uint64        `json:"write_ops"`
	LastSync      time.Time     `json:"last_sync"`
	Uptime        time.Duration `json:"uptime"`
}

// DocumentStorage 文档存储接口
type DocumentStorage interface {
	// 基础操作
	Store(ctx context.Context, doc Document) error
	Get(ctx context.Context, id string) (*Document, error)
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)

	// 批量操作
	BatchStore(ctx context.Context, docs []Document) error
	BatchGet(ctx context.Context, ids []string) ([]Document, error)
	BatchDelete(ctx context.Context, ids []string) error

	// 查询操作
	List(ctx context.Context, limit, offset int) ([]Document, error)
	Count(ctx context.Context) (uint64, error)

	// 管理操作
	Close() error
	Flush() error
	GetMetrics() StorageMetrics
}

// VectorStorage 向量存储接口
type VectorStorage interface {
	// 向量操作
	StoreVector(ctx context.Context, id string, vector []float64) error
	GetVector(ctx context.Context, id string) ([]float64, error)
	DeleteVector(ctx context.Context, id string) error

	// 搜索操作
	SearchSimilar(ctx context.Context, vector []float64, limit int) ([]VectorResult, error)
	SearchByThreshold(ctx context.Context, vector []float64, threshold float64) ([]VectorResult, error)

	// 批量操作
	BatchStoreVectors(ctx context.Context, vectors map[string][]float64) error

	// 管理操作
	GetDimensions() int
	GetVectorCount() uint64
	Close() error
}

// IndexStorage 索引存储接口
type IndexStorage interface {
	// 索引操作
	AddDocument(ctx context.Context, doc Document) error
	RemoveDocument(ctx context.Context, id string) error
	UpdateDocument(ctx context.Context, doc Document) error

	// 搜索操作
	Search(ctx context.Context, query string, limit int) ([]string, error)
	SearchTerms(ctx context.Context, terms []string, limit int) ([]string, error)

	// 统计操作
	GetTermFrequency(ctx context.Context, term, docID string) (uint32, error)
	GetDocumentFrequency(ctx context.Context, term string) (uint32, error)
	GetTermCount() uint64
	GetDocumentCount() uint64

	// 管理操作
	Optimize() error
	Close() error
}

// StorageEngine 统一存储引擎接口
type StorageEngine interface {
	// 存储管理
	DocumentStorage() DocumentStorage
	VectorStorage() VectorStorage
	IndexStorage() IndexStorage

	// 生命周期管理
	Initialize(config StorageConfig) error
	Close() error

	// 健康检查
	Health() error
	GetMetrics() StorageMetrics
}

// StorageProvider 存储提供者工厂接口
type StorageProvider interface {
	Name() string
	Create(config StorageConfig) (StorageEngine, error)
	Validate(config StorageConfig) error
}
