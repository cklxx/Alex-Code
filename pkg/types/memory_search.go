package types

import (
	"time"
)

// SearchResult represents a single search result
type SearchResult struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Summary    string                 `json:"summary"`
	Score      float64                `json:"score"`
	Relevance  float64                `json:"relevance"`
	Confidence float64                `json:"confidence"`
	Source     string                 `json:"source"`
	Highlights []string               `json:"highlights"`
	Context    string                 `json:"context"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// MemorySearchQuery represents a search query for memory system
type MemorySearchQuery struct {
	Query     string         `json:"query"`
	Type      KnowledgeType  `json:"type,omitempty"`
	Keywords  []string       `json:"keywords,omitempty"`
	Tags      []string       `json:"tags,omitempty"`
	ProjectID string         `json:"projectId,omitempty"`
	TimeRange *TimeRange     `json:"timeRange,omitempty"`
	Filters   *MemoryFilters `json:"filters,omitempty"`
	Limit     int            `json:"limit,omitempty"`
	SortBy    string         `json:"sortBy,omitempty"`
	SortOrder string         `json:"sortOrder,omitempty"`
}

// TimeRange represents a time range for filtering
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// MemoryFilters represents filters for memory searches
type MemoryFilters struct {
	MinConfidence float64    `json:"minConfidence,omitempty"`
	MinRelevance  float64    `json:"minRelevance,omitempty"`
	Categories    []string   `json:"categories,omitempty"`
	Sources       []string   `json:"sources,omitempty"`
	Authors       []string   `json:"authors,omitempty"`
	DateRange     *DateRange `json:"dateRange,omitempty"`
	ProjectScope  []string   `json:"projectScope,omitempty"`
	Verified      *bool      `json:"verified,omitempty"`
}

// MemorySearchResults represents search results from memory
type MemorySearchResults struct {
	Results     []SearchResult         `json:"results"`
	Total       int                    `json:"total"`
	Page        int                    `json:"page"`
	PageSize    int                    `json:"pageSize"`
	HasMore     bool                   `json:"hasMore"`
	Query       string                 `json:"query"`
	TimeSpent   time.Duration          `json:"timeSpent"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SearchContext represents context for search operations
type SearchContext struct {
	UserID      string                 `json:"userId,omitempty"`
	SessionID   string                 `json:"sessionId,omitempty"`
	ProjectID   string                 `json:"projectId,omitempty"`
	TaskContext string                 `json:"taskContext,omitempty"`
	CurrentFile string                 `json:"currentFile,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Intent      string                 `json:"intent,omitempty"`
	History     []string               `json:"history,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SearchExample represents an example for search
type SearchExample struct {
	Query    string                 `json:"query"`
	Content  string                 `json:"content"`
	Type     string                 `json:"type"`
	Expected []string               `json:"expected,omitempty"`
	Context  *SearchContext         `json:"context,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// QueryOptimizationContext represents context for query optimization
type QueryOptimizationContext struct {
	UserProfile   *UserProfile           `json:"userProfile,omitempty"`
	SearchHistory []string               `json:"searchHistory,omitempty"`
	Domain        string                 `json:"domain,omitempty"`
	Language      string                 `json:"language,omitempty"`
	Context       *SearchContext         `json:"context,omitempty"`
	Constraints   map[string]interface{} `json:"constraints,omitempty"`
}

// UserProfile represents a user profile for search optimization
type UserProfile struct {
	UserID         string            `json:"userId"`
	Preferences    map[string]string `json:"preferences"`
	Expertise      []string          `json:"expertise"`
	Interests      []string          `json:"interests"`
	SearchPatterns []string          `json:"searchPatterns"`
	Language       string            `json:"language"`
}

// IndexBuildConfig represents configuration for building indexes
type IndexBuildConfig struct {
	Incremental      bool                   `json:"incremental"`
	RebuildIndex     bool                   `json:"rebuildIndex"`
	MaxConcurrency   int                    `json:"maxConcurrency"`
	ChunkSize        int                    `json:"chunkSize"`
	IndexType        string                 `json:"indexType"` // full, partial, semantic
	Fields           []string               `json:"fields"`
	Weights          map[string]float64     `json:"weights,omitempty"`
	Filters          *IndexFilters          `json:"filters,omitempty"`
	QualityThreshold float64                `json:"qualityThreshold"`
	KnowledgeItems   []*Knowledge           `json:"knowledgeItems,omitempty"`
	PatternItems     []*CodePattern         `json:"patternItems,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// IndexFilters represents filters for index building
type IndexFilters struct {
	IncludeTypes []string `json:"includeTypes,omitempty"`
	ExcludeTypes []string `json:"excludeTypes,omitempty"`
	MinRelevance float64  `json:"minRelevance,omitempty"`
	MaxAge       string   `json:"maxAge,omitempty"`
	Categories   []string `json:"categories,omitempty"`
	Languages    []string `json:"languages,omitempty"`
}

// IndexUpdate represents an update to an index
type IndexUpdate struct {
	Type       string                 `json:"type"` // add, update, remove
	DocumentID string                 `json:"documentId"`
	Content    string                 `json:"content,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// DocumentType represents the type of document
type DocumentType string

const (
	DocumentTypeText         DocumentType = "text"
	DocumentTypeCode         DocumentType = "code"
	DocumentTypeMarkdown     DocumentType = "markdown"
	DocumentTypeJSON         DocumentType = "json"
	DocumentTypeXML          DocumentType = "xml"
	DocumentTypeHTML         DocumentType = "html"
	DocumentTypePDF          DocumentType = "pdf"
	DocumentTypeImage        DocumentType = "image"
	DocumentTypeAudio        DocumentType = "audio"
	DocumentTypeVideo        DocumentType = "video"
	DocumentTypeArchive      DocumentType = "archive"
	DocumentTypeSpreadsheet  DocumentType = "spreadsheet"
	DocumentTypePresentation DocumentType = "presentation"
)

// DocumentSection represents a section in a document
type DocumentSection struct {
	ID       string                 `json:"id"`
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Level    int                    `json:"level"`
	Parent   string                 `json:"parent,omitempty"`
	Children []string               `json:"children,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentQuality represents quality metrics for a document
type DocumentQuality struct {
	Overall       float64   `json:"overall"`
	Readability   float64   `json:"readability"`
	Completeness  float64   `json:"completeness"`
	Accuracy      float64   `json:"accuracy"`
	Freshness     float64   `json:"freshness"`
	Relevance     float64   `json:"relevance"`
	Structure     float64   `json:"structure"`
	Consistency   float64   `json:"consistency"`
	LastEvaluated time.Time `json:"lastEvaluated"`
}

// DocumentMetadata represents metadata for a document
type DocumentMetadata struct {
	Title      string            `json:"title"`
	Author     string            `json:"author"`
	Version    string            `json:"version"`
	Language   string            `json:"language"`
	Format     string            `json:"format"`
	Tags       []string          `json:"tags"`
	Categories []string          `json:"categories"`
	Keywords   []string          `json:"keywords"`
	Source     string            `json:"source"`
	CreatedAt  time.Time         `json:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt"`
	Properties map[string]string `json:"properties,omitempty"`
}

// DocumentFilters represents filters for document searches
type DocumentFilters struct {
	Types      []DocumentType `json:"types,omitempty"`
	Languages  []string       `json:"languages,omitempty"`
	Formats    []string       `json:"formats,omitempty"`
	Authors    []string       `json:"authors,omitempty"`
	Tags       []string       `json:"tags,omitempty"`
	Categories []string       `json:"categories,omitempty"`
	TimeRange  *TimeRange     `json:"timeRange,omitempty"`
	MinQuality float64        `json:"minQuality,omitempty"`
	MaxSize    int64          `json:"maxSize,omitempty"`
}

// DocumentAnalysis represents analysis of a document
type DocumentAnalysis struct {
	DocumentID       string                 `json:"documentId"`
	ReadabilityScore float64                `json:"readabilityScore"`
	ComplexityScore  float64                `json:"complexityScore"`
	Sentiment        *Sentiment             `json:"sentiment,omitempty"`
	Topics           []string               `json:"topics"`
	Keywords         []string               `json:"keywords"`
	Entities         []Entity               `json:"entities"`
	Structure        *DocumentStructure     `json:"structure"`
	Quality          *DocumentQuality       `json:"quality"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
}

// DocumentStructure represents the structure of a document
type DocumentStructure struct {
	Sections   []DocumentSection `json:"sections"`
	Hierarchy  []string          `json:"hierarchy"`
	WordCount  int               `json:"wordCount"`
	LineCount  int               `json:"lineCount"`
	HasIndex   bool              `json:"hasIndex"`
	HasTOC     bool              `json:"hasToc"`
	References int               `json:"references"`
}

// DocumentValidation represents validation of a document
type DocumentValidation struct {
	Valid             bool     `json:"valid"`
	Errors            []string `json:"errors,omitempty"`
	Warnings          []string `json:"warnings,omitempty"`
	Suggestions       []string `json:"suggestions,omitempty"`
	QualityScore      float64  `json:"qualityScore"`
	CompletenessScore float64  `json:"completenessScore"`
	ConsistencyScore  float64  `json:"consistencyScore"`
}

// DocumentLinkType represents the type of document link
type DocumentLinkType string

const (
	DocumentLinkTypeReferences DocumentLinkType = "references"
	DocumentLinkTypeRelated    DocumentLinkType = "related"
	DocumentLinkTypeDepends    DocumentLinkType = "depends"
	DocumentLinkTypeSupersedes DocumentLinkType = "supersedes"
	DocumentLinkTypeRevises    DocumentLinkType = "revises"
	DocumentLinkTypeTranslates DocumentLinkType = "translates"
)

// DocumentLink represents a link between documents
type DocumentLink struct {
	ID            string           `json:"id"`
	SourceID      string           `json:"sourceId"`
	TargetID      string           `json:"targetId"`
	Type          DocumentLinkType `json:"type"`
	Description   string           `json:"description"`
	Strength      float64          `json:"strength"` // 0.0-1.0
	Bidirectional bool             `json:"bidirectional"`
	CreatedAt     time.Time        `json:"createdAt"`
}

// DocumentManagerConfig represents configuration for document manager
type DocumentManagerConfig struct {
	MaxDocuments      int                      `json:"maxDocuments"`
	AutoIndexing      bool                     `json:"autoIndexing"`
	AutoLinking       bool                     `json:"autoLinking"`
	QualityThreshold  float64                  `json:"qualityThreshold"`
	ValidationEnabled bool                     `json:"validationEnabled"`
	AnalysisEnabled   bool                     `json:"analysisEnabled"`
	ContentExtraction *ContentExtractionConfig `json:"contentExtraction"`
	SearchConfig      *DocumentSearchConfig    `json:"searchConfig"`
}

// ContentExtractionConfig represents configuration for content extraction
type ContentExtractionConfig struct {
	Enabled          bool     `json:"enabled"`
	MaxSize          int64    `json:"maxSize"` // bytes
	SupportedFormats []string `json:"supportedFormats"`
	ExtractMetadata  bool     `json:"extractMetadata"`
	ExtractImages    bool     `json:"extractImages"`
	ExtractTables    bool     `json:"extractTables"`
}

// DocumentSearchConfig represents configuration for document search
type DocumentSearchConfig struct {
	Enabled                   bool    `json:"enabled"`
	FullTextSearch            bool    `json:"fullTextSearch"`
	SemanticSearch            bool    `json:"semanticSearch"`
	FuzzySearch               bool    `json:"fuzzySearch"`
	MaxResults                int     `json:"maxResults"`
	DefaultRelevanceThreshold float64 `json:"defaultRelevanceThreshold"`
}

// SearchEngineConfig represents configuration for the search engine
type SearchEngineConfig struct {
	Enabled               bool `json:"enabled"`
	IndexingEnabled       bool `json:"indexingEnabled"`
	SemanticSearchEnabled bool `json:"semanticSearchEnabled"`
	MaxSearchResults      int  `json:"maxSearchResults"`
}

// Document represents a document in the system
type Document struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Summary    string                 `json:"summary"`
	Type       DocumentType           `json:"type"`
	Format     string                 `json:"format"`
	Language   string                 `json:"language"`
	Author     string                 `json:"author"`
	Version    string                 `json:"version"`
	Tags       []string               `json:"tags"`
	Categories []string               `json:"categories"`
	Keywords   []string               `json:"keywords"`
	Source     string                 `json:"source"`
	Path       string                 `json:"path"`
	Size       int64                  `json:"size"`
	Checksum   string                 `json:"checksum"`
	Sections   []DocumentSection      `json:"sections"`
	References []DocumentReference    `json:"references"`
	Quality    *DocumentQuality       `json:"quality"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
	AccessedAt time.Time              `json:"accessedAt"`
	IndexedAt  *time.Time             `json:"indexedAt,omitempty"`
}

// DocumentReference represents a reference in a document
type DocumentReference struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	URL     string `json:"url,omitempty"`
	Author  string `json:"author,omitempty"`
	Year    string `json:"year,omitempty"`
	Context string `json:"context,omitempty"`
}
