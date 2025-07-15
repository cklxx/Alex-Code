package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage SQLite数据库存储实现
type SQLiteStorage struct {
	db        *sql.DB
	config    StorageConfig
	metrics   StorageMetrics
	mu        sync.RWMutex
	startTime time.Time
}

// NewSQLiteStorage 创建SQLite存储实例
func NewSQLiteStorage(config StorageConfig) (*SQLiteStorage, error) {
	if config.Path == "" {
		config.Path = ":memory:" // 默认内存数据库
	}

	db, err := sql.Open("sqlite3", config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// 配置连接池以支持并发访问
	if config.Path == ":memory:" {
		// 内存数据库限制连接数避免竞态条件
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
	} else {
		// 文件数据库支持更多并发连接
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
	}
	db.SetConnMaxLifetime(5 * time.Minute)

	storage := &SQLiteStorage{
		db:        db,
		config:    config,
		startTime: time.Now(),
	}

	if err := storage.initialize(); err != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 验证初始化成功
	if _, err := storage.Count(context.Background()); err != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
		return nil, fmt.Errorf("failed to verify database initialization: %w", err)
	}

	return storage, nil
}

// initialize 初始化数据库表结构
func (s *SQLiteStorage) initialize() error {
	queries := []string{
		// SQLite 配置优化
		`PRAGMA foreign_keys = ON`,
		`PRAGMA journal_mode = WAL`,
		`PRAGMA synchronous = NORMAL`,
		`PRAGMA cache_size = 10000`,
		`PRAGMA temp_store = memory`,
		`PRAGMA mmap_size = 268435456`,

		// 文档表
		`CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			metadata TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,

		// 向量表
		`CREATE TABLE IF NOT EXISTS vectors (
			id TEXT PRIMARY KEY,
			vector_data BLOB NOT NULL,
			dimensions INTEGER NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (id) REFERENCES documents(id) ON DELETE CASCADE
		)`,

		// 索引表
		`CREATE TABLE IF NOT EXISTS term_index (
			term TEXT NOT NULL,
			document_id TEXT NOT NULL,
			frequency INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL,
			PRIMARY KEY (term, document_id),
			FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
		)`,

		// 索引优化
		`CREATE INDEX IF NOT EXISTS idx_documents_created ON documents(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_term_index_term ON term_index(term)`,
		`CREATE INDEX IF NOT EXISTS idx_term_index_doc ON term_index(document_id)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	return nil
}

// === DocumentStorage接口实现 ===

// Store 存储文档
func (s *SQLiteStorage) Store(ctx context.Context, doc Document) error {
	metadataJSON, err := json.Marshal(doc.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if doc.Created.IsZero() {
		doc.Created = time.Now()
	}
	doc.Updated = time.Now()

	query := `INSERT OR REPLACE INTO documents 
		(id, title, content, metadata, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err = s.db.ExecContext(ctx, query,
		doc.ID, doc.Title, doc.Content, string(metadataJSON),
		doc.Created, doc.Updated)

	if err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	s.mu.Lock()
	s.metrics.WriteOps++
	s.mu.Unlock()
	return nil
}

// Get 获取文档
func (s *SQLiteStorage) Get(ctx context.Context, id string) (*Document, error) {
	query := `SELECT id, title, content, metadata, created_at, updated_at 
		FROM documents WHERE id = ?`

	row := s.db.QueryRowContext(ctx, query, id)

	var doc Document
	var metadataJSON string
	err := row.Scan(&doc.ID, &doc.Title, &doc.Content, &metadataJSON,
		&doc.Created, &doc.Updated)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &doc.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	s.mu.Lock()
	s.metrics.ReadOps++
	s.mu.Unlock()
	return &doc, nil
}

// Delete 删除文档
func (s *SQLiteStorage) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM documents WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("document not found: %s", id)
	}

	s.mu.Lock()
	s.metrics.WriteOps++
	s.mu.Unlock()
	return nil
}

// Exists 检查文档是否存在
func (s *SQLiteStorage) Exists(ctx context.Context, id string) (bool, error) {
	query := `SELECT 1 FROM documents WHERE id = ? LIMIT 1`
	row := s.db.QueryRowContext(ctx, query, id)

	var exists int
	err := row.Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check document existence: %w", err)
	}

	return true, nil
}

// BatchStore 批量存储文档
func (s *SQLiteStorage) BatchStore(ctx context.Context, docs []Document) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			// Ignore rollback errors as they may be due to successful commit - intentionally empty
			_ = rollbackErr // Suppress staticcheck warning
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `INSERT OR REPLACE INTO documents 
		(id, title, content, metadata, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Printf("Error closing statement: %v", err)
		}
	}()

	for _, doc := range docs {
		metadataJSON, err := json.Marshal(doc.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for doc %s: %w", doc.ID, err)
		}

		if doc.Created.IsZero() {
			doc.Created = time.Now()
		}
		doc.Updated = time.Now()

		_, err = stmt.ExecContext(ctx, doc.ID, doc.Title, doc.Content,
			string(metadataJSON), doc.Created, doc.Updated)
		if err != nil {
			return fmt.Errorf("failed to store document %s: %w", doc.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.mu.Lock()
	s.metrics.WriteOps += uint64(len(docs))
	s.mu.Unlock()
	return nil
}

// BatchGet 批量获取文档
func (s *SQLiteStorage) BatchGet(ctx context.Context, ids []string) ([]Document, error) {
	if len(ids) == 0 {
		return []Document{}, nil
	}

	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1] // 移除最后的逗号

	query := fmt.Sprintf(`SELECT id, title, content, metadata, created_at, updated_at 
		FROM documents WHERE id IN (%s)`, placeholders)

	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var docs []Document
	for rows.Next() {
		var doc Document
		var metadataJSON string
		err := rows.Scan(&doc.ID, &doc.Title, &doc.Content, &metadataJSON,
			&doc.Created, &doc.Updated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}

		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &doc.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		docs = append(docs, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	s.mu.Lock()
	s.metrics.ReadOps += uint64(len(docs))
	s.mu.Unlock()
	return docs, nil
}

// BatchDelete 批量删除文档
func (s *SQLiteStorage) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1] // 移除最后的逗号

	query := fmt.Sprintf(`DELETE FROM documents WHERE id IN (%s)`, placeholders)

	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to batch delete documents: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	s.mu.Lock()
	s.metrics.WriteOps += uint64(rowsAffected)
	s.mu.Unlock()
	return nil
}

// List 列出文档
func (s *SQLiteStorage) List(ctx context.Context, limit, offset int) ([]Document, error) {
	query := `SELECT id, title, content, metadata, created_at, updated_at 
		FROM documents ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var docs []Document
	for rows.Next() {
		var doc Document
		var metadataJSON string
		err := rows.Scan(&doc.ID, &doc.Title, &doc.Content, &metadataJSON,
			&doc.Created, &doc.Updated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}

		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &doc.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		docs = append(docs, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	s.mu.Lock()
	s.metrics.ReadOps += uint64(len(docs))
	s.mu.Unlock()
	return docs, nil
}

// Count 统计文档数量
func (s *SQLiteStorage) Count(ctx context.Context) (uint64, error) {
	query := `SELECT COUNT(*) FROM documents`
	row := s.db.QueryRowContext(ctx, query)

	var count uint64
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	s.mu.Lock()
	s.metrics.DocumentCount = count
	s.mu.Unlock()
	return count, nil
}

// === VectorStorage接口实现 ===

// StoreVector 存储向量
func (s *SQLiteStorage) StoreVector(ctx context.Context, id string, vector []float64) error {
	vectorJSON, err := json.Marshal(vector)
	if err != nil {
		return fmt.Errorf("failed to marshal vector: %w", err)
	}

	query := `INSERT OR REPLACE INTO vectors 
		(id, vector_data, dimensions, created_at) 
		VALUES (?, ?, ?, ?)`

	_, err = s.db.ExecContext(ctx, query, id, vectorJSON, len(vector), time.Now())
	if err != nil {
		return fmt.Errorf("failed to store vector: %w", err)
	}

	return nil
}

// GetVector 获取向量
func (s *SQLiteStorage) GetVector(ctx context.Context, id string) ([]float64, error) {
	query := `SELECT vector_data FROM vectors WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var vectorJSON []byte
	err := row.Scan(&vectorJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("vector not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get vector: %w", err)
	}

	var vector []float64
	if err := json.Unmarshal(vectorJSON, &vector); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vector: %w", err)
	}

	return vector, nil
}

// DeleteVector 删除向量
func (s *SQLiteStorage) DeleteVector(ctx context.Context, id string) error {
	query := `DELETE FROM vectors WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete vector: %w", err)
	}
	return nil
}

// SearchSimilar 相似向量搜索 (简化实现，实际生产应使用专业向量数据库)
func (s *SQLiteStorage) SearchSimilar(ctx context.Context, queryVector []float64, limit int) ([]VectorResult, error) {
	query := `SELECT v.id, v.vector_data, d.title, d.content, d.metadata, d.created_at, d.updated_at
		FROM vectors v 
		JOIN documents d ON v.id = d.id`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query vectors: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var results []VectorResult
	for rows.Next() {
		var doc Document
		var vectorJSON []byte
		var metadataJSON string

		err := rows.Scan(&doc.ID, &vectorJSON, &doc.Title, &doc.Content,
			&metadataJSON, &doc.Created, &doc.Updated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vector result: %w", err)
		}

		var vector []float64
		if err := json.Unmarshal(vectorJSON, &vector); err != nil {
			continue // 跳过无效向量
		}

		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &doc.Metadata); err != nil {
				// Continue with empty metadata if unmarshal fails
				doc.Metadata = make(map[string]string)
			}
		}

		similarity := sqliteCosineSimilarity(queryVector, vector)
		if similarity > 0 { // 只包含有相似性的结果
			results = append(results, VectorResult{
				Document:   doc,
				Similarity: similarity,
				Score:      similarity,
			})
		}
	}

	// 按相似度排序
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Similarity < results[j].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 限制结果数量
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// SearchByThreshold 按阈值搜索向量
func (s *SQLiteStorage) SearchByThreshold(ctx context.Context, queryVector []float64, threshold float64) ([]VectorResult, error) {
	results, err := s.SearchSimilar(ctx, queryVector, 100) // 获取较多结果后筛选
	if err != nil {
		return nil, err
	}

	var filteredResults []VectorResult
	for _, result := range results {
		if result.Similarity >= threshold {
			filteredResults = append(filteredResults, result)
		}
	}

	return filteredResults, nil
}

// BatchStoreVectors 批量存储向量
func (s *SQLiteStorage) BatchStoreVectors(ctx context.Context, vectors map[string][]float64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			// Ignore rollback errors as they may be due to successful commit - intentionally empty
			_ = rollbackErr // Suppress staticcheck warning
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `INSERT OR REPLACE INTO vectors 
		(id, vector_data, dimensions, created_at) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Printf("Error closing statement: %v", err)
		}
	}()

	for id, vector := range vectors {
		vectorJSON, err := json.Marshal(vector)
		if err != nil {
			return fmt.Errorf("failed to marshal vector for %s: %w", id, err)
		}

		_, err = stmt.ExecContext(ctx, id, vectorJSON, len(vector), time.Now())
		if err != nil {
			return fmt.Errorf("failed to store vector %s: %w", id, err)
		}
	}

	return tx.Commit()
}

// GetDimensions 获取向量维度
func (s *SQLiteStorage) GetDimensions() int {
	query := `SELECT dimensions FROM vectors LIMIT 1`
	row := s.db.QueryRow(query)

	var dimensions int
	err := row.Scan(&dimensions)
	if err != nil {
		return 128 // 默认维度
	}

	return dimensions
}

// GetVectorCount 获取向量数量
func (s *SQLiteStorage) GetVectorCount() uint64 {
	query := `SELECT COUNT(*) FROM vectors`
	row := s.db.QueryRow(query)

	var count uint64
	if err := row.Scan(&count); err != nil {
		return 0 // Return 0 if scan fails
	}
	return count
}

// Close 关闭存储
func (s *SQLiteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Flush 刷新数据
func (s *SQLiteStorage) Flush() error {
	_, err := s.db.Exec("PRAGMA optimize")
	return err
}

// GetMetrics 获取存储指标
func (s *SQLiteStorage) GetMetrics() StorageMetrics {
	// 更新文档数量（在获取锁之前）
	if count, err := s.Count(context.Background()); err == nil {
		s.mu.Lock()
		s.metrics.DocumentCount = count
		s.mu.Unlock()
	}

	s.mu.Lock()
	s.metrics.Uptime = time.Since(s.startTime)

	// 返回指标的副本以避免竞态条件
	result := s.metrics
	s.mu.Unlock()

	return result
}

// === 工具函数 ===

// sqliteCosineSimilarity 计算余弦相似度 (SQLite存储专用)
func sqliteCosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
