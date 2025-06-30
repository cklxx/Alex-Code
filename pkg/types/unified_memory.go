package types

import (
	"time"
)

// UnifiedMemory consolidates all memory-related types into a single, simplified structure
// This replaces the complex hierarchy spread across multiple memory_*.go files
type UnifiedMemory struct {
	ID          string                 `json:"id"`
	Type        MemoryType             `json:"type"`
	Content     string                 `json:"content"`
	Summary     string                 `json:"summary,omitempty"`
	Keywords    []string               `json:"keywords,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Category    string                 `json:"category,omitempty"`
	ProjectID   string                 `json:"projectId,omitempty"`
	Confidence  float64                `json:"confidence"` // 0.0-1.0
	Relevance   float64                `json:"relevance"`  // 0.0-1.0
	AccessCount int                    `json:"accessCount"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	AccessedAt  time.Time              `json:"accessedAt"`
	ExpiresAt   *time.Time             `json:"expiresAt,omitempty"`
}

// MemoryType represents the unified type of memory
type MemoryType string

const (
	MemoryTypeCode       MemoryType = "code"
	MemoryTypePattern    MemoryType = "pattern"
	MemoryTypeDecision   MemoryType = "decision"
	MemoryTypeExperience MemoryType = "experience"
	MemoryTypeError      MemoryType = "error"
	MemoryTypeSolution   MemoryType = "solution"
	MemoryTypeLesson     MemoryType = "lesson"
	MemoryTypeInsight    MemoryType = "insight"
	MemoryTypeContext    MemoryType = "context"
)

// SimplifiedMemoryConfig consolidates all memory configuration into one structure
type SimplifiedMemoryConfig struct {
	MaxItems           int                    `json:"maxItems"`
	RetentionDays      int                    `json:"retentionDays"`
	StorageType        string                 `json:"storageType"` // file, memory, database
	StoragePath        string                 `json:"storagePath"`
	CompressionEnabled bool                   `json:"compressionEnabled"`
	BackupEnabled      bool                   `json:"backupEnabled"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// MemoryQuery represents a simplified query for memory retrieval
type MemoryQuery struct {
	Text         string            `json:"text"`
	Type         MemoryType        `json:"type,omitempty"`
	ProjectID    string            `json:"projectId,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	MaxResults   int               `json:"maxResults"`
	MinRelevance float64           `json:"minRelevance"`
	Filters      map[string]string `json:"filters,omitempty"`
}

// MemorySearchResult represents search results
type MemorySearchResult struct {
	Memories   []UnifiedMemory `json:"memories"`
	TotalCount int             `json:"totalCount"`
	SearchTime time.Duration   `json:"searchTime"`
	Query      MemoryQuery     `json:"query"`
}

// UnifiedMemoryMetrics represents simplified memory metrics
type UnifiedMemoryMetrics struct {
	TotalMemories   int                    `json:"totalMemories"`
	MemoriesByType  map[MemoryType]int     `json:"memoriesByType"`
	RecentAccess    int                    `json:"recentAccess"` // last 24h
	StorageSize     int64                  `json:"storageSize"`  // bytes
	LastCleanup     time.Time              `json:"lastCleanup"`
	PerformanceData map[string]interface{} `json:"performanceData,omitempty"`
}
