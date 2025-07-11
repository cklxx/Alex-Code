//go:build chromem
// +build chromem

package storage

import (
	"fmt"
)

// ChromemStorageProvider Chromem存储提供者
type ChromemStorageProvider struct{}

// Name 返回提供者名称
func (csp *ChromemStorageProvider) Name() string {
	return string(ChromemStorageT)
}

// Create 创建Chromem存储
func (csp *ChromemStorageProvider) Create(config StorageConfig) (StorageEngine, error) {
	storage, err := NewChromemStorage(config)
	if err != nil {
		return nil, err
	}

	// 创建一个包装器来实现StorageEngine接口
	return &ChromemStorageWrapper{storage: storage}, nil
}

// Validate 验证Chromem配置
func (csp *ChromemStorageProvider) Validate(config StorageConfig) error {
	if config.Type != string(ChromemStorageT) {
		return fmt.Errorf("invalid storage type for chromem provider: %s", config.Type)
	}
	return nil
}

// ChromemStorageWrapper Chromem存储包装器
type ChromemStorageWrapper struct {
	storage *ChromemStorage
}

// DocumentStorage 实现StorageEngine接口
func (w *ChromemStorageWrapper) DocumentStorage() DocumentStorage {
	return w.storage
}

// VectorStorage 实现StorageEngine接口
func (w *ChromemStorageWrapper) VectorStorage() VectorStorage {
	return w.storage
}

// IndexStorage 实现StorageEngine接口
func (w *ChromemStorageWrapper) IndexStorage() IndexStorage {
	return nil // Chromem不需要独立的IndexStorage
}

// Initialize 初始化
func (w *ChromemStorageWrapper) Initialize(config StorageConfig) error {
	return nil // Chromem在创建时已初始化
}

// Close 关闭存储
func (w *ChromemStorageWrapper) Close() error {
	return w.storage.Close()
}

// Health 健康检查
func (w *ChromemStorageWrapper) Health() error {
	return nil // 简化实现
}

// GetMetrics 获取指标
func (w *ChromemStorageWrapper) GetMetrics() StorageMetrics {
	return w.storage.GetMetrics()
}

// NewChromemDB 创建Chromem数据库的便利函数
func NewChromemDB(path string) (*ChromemStorage, error) {
	config := StorageConfig{
		Type: string(ChromemStorageT),
		Path: path,
	}
	return NewChromemStorage(config)
}
