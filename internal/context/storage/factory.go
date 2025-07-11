package storage

import (
	"fmt"
	"strings"
)

// StorageType 存储类型
type StorageType string

const (
	MemoryStorage   StorageType = "memory"
	SQLiteStorageT  StorageType = "sqlite"
	ChromemStorageT StorageType = "chromem"
)

// MemoryStorageProvider 内存存储提供者
type MemoryStorageProvider struct{}

// Name 返回提供者名称
func (msp *MemoryStorageProvider) Name() string {
	return string(MemoryStorage)
}

// Create 创建存储引擎
func (msp *MemoryStorageProvider) Create(config StorageConfig) (StorageEngine, error) {
	engine := NewMemoryStorageEngine()
	if err := engine.Initialize(config); err != nil {
		return nil, err
	}
	return engine, nil
}

// Validate 验证配置
func (msp *MemoryStorageProvider) Validate(config StorageConfig) error {
	if config.Type != string(MemoryStorage) {
		return fmt.Errorf("invalid storage type for memory provider: %s", config.Type)
	}
	return nil
}

// SQLiteStorageProvider SQLite存储提供者
type SQLiteStorageProvider struct{}

// Name 返回提供者名称
func (ssp *SQLiteStorageProvider) Name() string {
	return string(SQLiteStorageT)
}

// Create 创建SQLite存储
func (ssp *SQLiteStorageProvider) Create(config StorageConfig) (StorageEngine, error) {
	storage, err := NewSQLiteStorage(config)
	if err != nil {
		return nil, err
	}

	// 创建一个包装器来实现StorageEngine接口
	return &SQLiteStorageWrapper{storage: storage}, nil
}

// Validate 验证SQLite配置
func (ssp *SQLiteStorageProvider) Validate(config StorageConfig) error {
	if config.Type != string(SQLiteStorageT) {
		return fmt.Errorf("invalid storage type for sqlite provider: %s", config.Type)
	}
	return nil
}

// StorageFactory 存储工厂
type StorageFactory struct {
	providers map[string]StorageProvider
}

// NewStorageFactory 创建存储工厂
func NewStorageFactory() *StorageFactory {
	factory := &StorageFactory{
		providers: make(map[string]StorageProvider),
	}

	// 注册所有可用的存储提供者
	factory.RegisterProvider(&MemoryStorageProvider{})
	factory.RegisterProvider(&SQLiteStorageProvider{})
	factory.RegisterProvider(&ChromemStorageProvider{})

	return factory
}

// RegisterProvider 注册存储提供者
func (sf *StorageFactory) RegisterProvider(provider StorageProvider) {
	sf.providers[provider.Name()] = provider
}

// CreateStorage 创建存储引擎
func (sf *StorageFactory) CreateStorage(config StorageConfig) (StorageEngine, error) {
	storageType := strings.ToLower(config.Type)
	if storageType == "" {
		storageType = string(MemoryStorage) // 默认使用内存存储
	}

	provider, exists := sf.providers[storageType]
	if !exists {
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}

	if err := provider.Validate(config); err != nil {
		return nil, fmt.Errorf("invalid config for %s storage: %w", storageType, err)
	}

	return provider.Create(config)
}

// GetAvailableTypes 获取可用的存储类型
func (sf *StorageFactory) GetAvailableTypes() []string {
	var types []string
	for name := range sf.providers {
		types = append(types, name)
	}
	return types
}

// DefaultStorageConfig 默认存储配置
func DefaultStorageConfig() StorageConfig {
	return StorageConfig{
		Type:      string(MemoryStorage), // 默认使用内存存储
		Path:      "",
		MaxSize:   1024 * 1024 * 1024, // 1GB
		CacheSize: 1000,
		Options:   make(map[string]string),
	}
}

// === 存储包装器实现 ===

// SQLiteStorageWrapper SQLite存储包装器
type SQLiteStorageWrapper struct {
	storage *SQLiteStorage
}

// DocumentStorage 实现StorageEngine接口
func (w *SQLiteStorageWrapper) DocumentStorage() DocumentStorage {
	return w.storage
}

// VectorStorage 实现StorageEngine接口
func (w *SQLiteStorageWrapper) VectorStorage() VectorStorage {
	return w.storage
}

// IndexStorage 实现StorageEngine接口
func (w *SQLiteStorageWrapper) IndexStorage() IndexStorage {
	return nil // SQLite不实现IndexStorage
}

// Initialize 初始化
func (w *SQLiteStorageWrapper) Initialize(config StorageConfig) error {
	return nil // SQLite在创建时已初始化
}

// Close 关闭存储
func (w *SQLiteStorageWrapper) Close() error {
	return w.storage.Close()
}

// Health 健康检查
func (w *SQLiteStorageWrapper) Health() error {
	return nil // 简化实现
}

// GetMetrics 获取指标
func (w *SQLiteStorageWrapper) GetMetrics() StorageMetrics {
	return w.storage.GetMetrics()
}

// === 便利函数 ===

// NewSQLiteDB 创建SQLite数据库的便利函数
func NewSQLiteDB(path string) (*SQLiteStorage, error) {
	config := StorageConfig{
		Type: string(SQLiteStorageT),
		Path: path,
	}
	return NewSQLiteStorage(config)
}
