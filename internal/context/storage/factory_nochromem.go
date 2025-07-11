//go:build !chromem
// +build !chromem

package storage

import (
	"fmt"
)

// ChromemStorageProvider Chromem存储提供者 (占位实现)
type ChromemStorageProvider struct{}

// Name 返回提供者名称
func (csp *ChromemStorageProvider) Name() string {
	return string(ChromemStorageT)
}

// Create 创建Chromem存储 (不可用时返回错误)
func (csp *ChromemStorageProvider) Create(config StorageConfig) (StorageEngine, error) {
	return nil, fmt.Errorf("chromem storage not available - build with -tags chromem to enable")
}

// Validate 验证Chromem配置
func (csp *ChromemStorageProvider) Validate(config StorageConfig) error {
	if config.Type != string(ChromemStorageT) {
		return fmt.Errorf("invalid storage type for chromem provider: %s", config.Type)
	}
	return fmt.Errorf("chromem storage not available - build with -tags chromem to enable")
}

// ChromemStorageWrapper Chromem存储包装器 (占位实现)
type ChromemStorageWrapper struct{}

// DocumentStorage 实现StorageEngine接口
func (w *ChromemStorageWrapper) DocumentStorage() DocumentStorage {
	return nil
}

// VectorStorage 实现StorageEngine接口
func (w *ChromemStorageWrapper) VectorStorage() VectorStorage {
	return nil
}

// IndexStorage 实现StorageEngine接口
func (w *ChromemStorageWrapper) IndexStorage() IndexStorage {
	return nil
}

// Initialize 初始化
func (w *ChromemStorageWrapper) Initialize(config StorageConfig) error {
	return fmt.Errorf("chromem storage not available")
}

// Close 关闭存储
func (w *ChromemStorageWrapper) Close() error {
	return nil
}

// Health 健康检查
func (w *ChromemStorageWrapper) Health() error {
	return fmt.Errorf("chromem storage not available")
}

// GetMetrics 获取指标
func (w *ChromemStorageWrapper) GetMetrics() StorageMetrics {
	return StorageMetrics{}
}

// ChromemStorage 占位类型
type ChromemStorage struct{}

// NewChromemDB 创建Chromem数据库的便利函数 (占位实现)
func NewChromemDB(path string) (*ChromemStorage, error) {
	return nil, fmt.Errorf("chromem storage not available - build with -tags chromem to enable")
}
