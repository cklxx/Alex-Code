package memory

import (
	"time"
)

// MemoryType represents different types of memory storage
type MemoryType string

const (
	ShortTermMemory MemoryType = "short_term"
	LongTermMemory  MemoryType = "long_term"
)

// MemoryCategory represents categorization of memories
type MemoryCategory string

const (
	CodeContext     MemoryCategory = "code_context"
	UserPreferences MemoryCategory = "user_preferences"
	TaskHistory     MemoryCategory = "task_history"
	Knowledge       MemoryCategory = "knowledge"
	ErrorPatterns   MemoryCategory = "error_patterns"
	Solutions       MemoryCategory = "solutions"
)

// MemoryItem represents a single memory entry
type MemoryItem struct {
	ID          string                 `json:"id"`
	SessionID   string                 `json:"session_id"`
	Type        MemoryType             `json:"type"`
	Category    MemoryCategory         `json:"category"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	Importance  float64                `json:"importance"` // 0.0 - 1.0
	AccessCount int                    `json:"access_count"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastAccess  time.Time              `json:"last_access"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Tags        []string               `json:"tags"`
}

// MemoryQuery represents a query for memory retrieval
type MemoryQuery struct {
	SessionID     string           `json:"session_id,omitempty"`
	Types         []MemoryType     `json:"types,omitempty"`
	Categories    []MemoryCategory `json:"categories,omitempty"`
	Tags          []string         `json:"tags,omitempty"`
	Content       string           `json:"content,omitempty"`
	MinImportance float64          `json:"min_importance,omitempty"`
	Limit         int              `json:"limit,omitempty"`
	SortBy        string           `json:"sort_by,omitempty"` // "importance", "recency", "access_count"
}

// CompressionConfig represents configuration for context compression
type CompressionConfig struct {
	Threshold         float64 `json:"threshold"`           // Token usage threshold to trigger compression
	CompressionRatio  float64 `json:"compression_ratio"`   // Target compression ratio
	PreserveRecent    int     `json:"preserve_recent"`     // Number of recent messages to preserve
	MinImportance     float64 `json:"min_importance"`      // Minimum importance score to preserve
	EnableLLMCompress bool    `json:"enable_llm_compress"` // Use LLM for intelligent compression
}

// MemoryStats represents memory system statistics
type MemoryStats struct {
	TotalItems      int                    `json:"total_items"`
	ItemsByType     map[MemoryType]int     `json:"items_by_type"`
	ItemsByCategory map[MemoryCategory]int `json:"items_by_category"`
	TotalSize       int64                  `json:"total_size"`
	LastCompression time.Time              `json:"last_compression"`
	CompressionRate float64                `json:"compression_rate"`
}

// RecallResult represents the result of memory recall
type RecallResult struct {
	Items           []*MemoryItem `json:"items"`
	TotalFound      int           `json:"total_found"`
	RelevanceScores []float64     `json:"relevance_scores"`
	ProcessingTime  time.Duration `json:"processing_time"`
}
