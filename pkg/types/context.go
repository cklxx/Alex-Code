package types

import (
	"time"
)

// Context represents the unified context for agent operations
type Context struct {
	ID           string                 `json:"id"`
	Type         ContextType            `json:"type"`
	Scope        ContextScope           `json:"scope"`
	Input        *ContextInput          `json:"input"`
	State        *ContextState          `json:"state"`
	RelevantInfo []RelevantInfo         `json:"relevantInfo"`
	History      *HistoryContext        `json:"history,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
	ExpiresAt    *time.Time             `json:"expiresAt,omitempty"`
	Size         int64                  `json:"size"`       // Size in bytes
	Compressed   bool                   `json:"compressed"` // Whether context is compressed
	Relevance    float64                `json:"relevance"`  // Relevance score (0.0-1.0)
}

// ContextType represents the type of context
type ContextType string

const (
	ContextTypeFile    ContextType = "file"
	ContextTypeProject ContextType = "project"
	ContextTypeSession ContextType = "session"
	ContextTypeTask    ContextType = "task"
	ContextTypeCode    ContextType = "code"
	ContextTypeError   ContextType = "error"
	ContextTypeUser    ContextType = "user"
	ContextTypeSystem  ContextType = "system"
)

// ContextScope represents the scope of context
type ContextScope string

const (
	ContextScopeLocal   ContextScope = "local"   // Single file/function
	ContextScopeModule  ContextScope = "module"  // Module/package level
	ContextScopeProject ContextScope = "project" // Entire project
	ContextScopeGlobal  ContextScope = "global"  // Cross-project
)

// ContextUpdate represents updates to a context
type ContextUpdate struct {
	RelevantInfo []RelevantInfo         `json:"relevantInfo,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Relevance    *float64               `json:"relevance,omitempty"`
}

// ContextInput represents input for context generation
type ContextInput struct {
	Query       string              `json:"query"`
	Files       []string            `json:"files"`
	Directories []string            `json:"directories"`
	Patterns    []string            `json:"patterns"`
	Filters     *ContextFilters     `json:"filters,omitempty"`
	Options     *ContextOptions     `json:"options,omitempty"`
	Constraints *ContextConstraints `json:"constraints,omitempty"`
}

// ContextFilters represents filters for context generation
type ContextFilters struct {
	FileTypes []string   `json:"fileTypes"`
	Languages []string   `json:"languages"`
	DateRange *DateRange `json:"dateRange,omitempty"`
	SizeRange *SizeRange `json:"sizeRange,omitempty"`
	Authors   []string   `json:"authors,omitempty"`
	Keywords  []string   `json:"keywords,omitempty"`
	Exclude   []string   `json:"exclude"`
	Include   []string   `json:"include"`
}

// DateRange represents a date range filter
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// SizeRange represents a size range filter
type SizeRange struct {
	Min int64 `json:"min"` // bytes
	Max int64 `json:"max"` // bytes
}

// ContextOptions represents options for context generation
type ContextOptions struct {
	MaxDepth       int                `json:"maxDepth"`
	MaxFiles       int                `json:"maxFiles"`
	MaxSize        int64              `json:"maxSize"` // bytes
	IncludeTests   bool               `json:"includeTests"`
	IncludeDocs    bool               `json:"includeDocs"`
	IncludeConfig  bool               `json:"includeConfig"`
	IncludeHidden  bool               `json:"includeHidden"`
	FollowSymlinks bool               `json:"followSymlinks"`
	Compression    bool               `json:"compression"`
	Priorities     map[string]float64 `json:"priorities"` // File type priorities
}

// ContextConstraints represents constraints for context generation
type ContextConstraints struct {
	MaxTokens      int           `json:"maxTokens"`
	MaxTime        time.Duration `json:"maxTime"`
	MinRelevance   float64       `json:"minRelevance"`
	RequiredFiles  []string      `json:"requiredFiles"`
	ForbiddenFiles []string      `json:"forbiddenFiles"`
	RequiredInfo   []string      `json:"requiredInfo"`
}

// ContextState represents the current state of context
type ContextState struct {
	Status          ContextStatus  `json:"status"`
	Progress        float64        `json:"progress"` // 0.0-1.0
	FilesProcessed  int            `json:"filesProcessed"`
	FilesTotal      int            `json:"filesTotal"`
	TokensUsed      int            `json:"tokensUsed"`
	TokensRemaining int            `json:"tokensRemaining"`
	EstimatedTime   time.Duration  `json:"estimatedTime"`
	Warnings        []string       `json:"warnings"`
	Errors          []ContextError `json:"errors"`
}

// ContextStatus represents the status of context generation
type ContextStatus string

const (
	ContextStatusPending    ContextStatus = "pending"
	ContextStatusBuilding   ContextStatus = "building"
	ContextStatusComplete   ContextStatus = "complete"
	ContextStatusFailed     ContextStatus = "failed"
	ContextStatusExpired    ContextStatus = "expired"
	ContextStatusCompressed ContextStatus = "compressed"
)

// ContextError represents an error in context generation
type ContextError struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	File      string    `json:"file,omitempty"`
	Line      int       `json:"line,omitempty"`
	Severity  string    `json:"severity"` // error, warning, info
	Timestamp time.Time `json:"timestamp"`
}

// RelevantInfo represents a piece of relevant information
type RelevantInfo struct {
	ID         string                 `json:"id"`
	Type       InfoType               `json:"type"`
	Source     string                 `json:"source"`
	Content    string                 `json:"content"`
	Summary    string                 `json:"summary"`
	Relevance  float64                `json:"relevance"`  // 0.0-1.0
	Confidence float64                `json:"confidence"` // 0.0-1.0
	Keywords   []string               `json:"keywords"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	References []string               `json:"references"`
	CreatedAt  time.Time              `json:"createdAt"`
}

// InfoType represents the type of relevant information
type InfoType string

const (
	InfoTypeCode          InfoType = "code"
	InfoTypeDocumentation InfoType = "documentation"
	InfoTypeComment       InfoType = "comment"
	InfoTypeFunction      InfoType = "function"
	InfoTypeClass         InfoType = "class"
	InfoTypeInterface     InfoType = "interface"
	InfoTypeType          InfoType = "type"
	InfoTypeVariable      InfoType = "variable"
	InfoTypeImport        InfoType = "import"
	InfoTypeError         InfoType = "error"
	InfoTypePattern       InfoType = "pattern"
	InfoTypeExample       InfoType = "example"
)

// HistoryContext represents historical context information
type HistoryContext struct {
	RecentTasks         []TaskSummary        `json:"recentTasks"`
	UserPatterns        []UserPattern        `json:"userPatterns"`
	ProjectHistory      *ProjectHistory      `json:"projectHistory,omitempty"`
	ConversationContext *ConversationContext `json:"conversationContext,omitempty"`
	DecisionHistory     []Decision           `json:"decisionHistory"`
	LastAccessed        time.Time            `json:"lastAccessed"`
}

// TaskSummary represents a summary of a previous task
type TaskSummary struct {
	ID           string        `json:"id"`
	Type         TaskType      `json:"type"`
	Description  string        `json:"description"`
	Status       TaskStatus    `json:"status"`
	Duration     time.Duration `json:"duration"`
	Success      bool          `json:"success"`
	FilesChanged []string      `json:"filesChanged"`
	ToolsUsed    []string      `json:"toolsUsed"`
	Timestamp    time.Time     `json:"timestamp"`
}

// UserPattern represents a pattern in user behavior
type UserPattern struct {
	Type        PatternType `json:"type"`
	Description string      `json:"description"`
	Frequency   int         `json:"frequency"`
	Confidence  float64     `json:"confidence"`
	LastSeen    time.Time   `json:"lastSeen"`
	Context     []string    `json:"context"`
}

// PatternType represents the type of user pattern
type PatternType string

const (
	PatternTypePreference PatternType = "preference"
	PatternTypeWorkflow   PatternType = "workflow"
	PatternTypeStyle      PatternType = "style"
	PatternTypeUsage      PatternType = "usage"
	PatternTypeError      PatternType = "error"
)

// ProjectHistory represents the history of a project
type ProjectHistory struct {
	CreatedAt     time.Time      `json:"createdAt"`
	LastModified  time.Time      `json:"lastModified"`
	TotalCommits  int            `json:"totalCommits"`
	Contributors  []string       `json:"contributors"`
	Languages     map[string]int `json:"languages"` // Language -> line count
	Architecture  []string       `json:"architecture"`
	Dependencies  []string       `json:"dependencies"`
	RecentChanges []FileChange   `json:"recentChanges"`
	Milestones    []Milestone    `json:"milestones"`
}

// FileChange represents a change to a file
type FileChange struct {
	File         string    `json:"file"`
	Type         string    `json:"type"` // added, modified, deleted
	LinesAdded   int       `json:"linesAdded"`
	LinesRemoved int       `json:"linesRemoved"`
	Author       string    `json:"author"`
	Timestamp    time.Time `json:"timestamp"`
	Commit       string    `json:"commit"`
}

// Milestone represents a project milestone
type Milestone struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Version     string    `json:"version,omitempty"`
	Changes     []string  `json:"changes"`
}

// ConversationContext represents context from previous conversations
type ConversationContext struct {
	SessionID        string            `json:"sessionId"`
	MessageCount     int               `json:"messageCount"`
	TopicEvolution   []Topic           `json:"topicEvolution"`
	KeyDecisions     []Decision        `json:"keyDecisions"`
	UnresolvedIssues []UnresolvedIssue `json:"unresolvedIssues"`
	Preferences      map[string]string `json:"preferences"`
	LastInteraction  time.Time         `json:"lastInteraction"`
}

// Topic represents a topic in the conversation
type Topic struct {
	Name         string    `json:"name"`
	Keywords     []string  `json:"keywords"`
	Relevance    float64   `json:"relevance"`
	FirstMention time.Time `json:"firstMention"`
	LastMention  time.Time `json:"lastMention"`
	Frequency    int       `json:"frequency"`
}

// Decision represents a decision made in the conversation
type Decision struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Rationale   string                 `json:"rationale"`
	Alternative []string               `json:"alternatives"`
	Impact      DecisionImpact         `json:"impact"`
	Confidence  float64                `json:"confidence"`
	Context     map[string]interface{} `json:"context"`
	Timestamp   time.Time              `json:"timestamp"`
	Outcome     *DecisionOutcome       `json:"outcome,omitempty"`
}

// DecisionImpact represents the impact level of a decision
type DecisionImpact string

const (
	DecisionImpactLow      DecisionImpact = "low"
	DecisionImpactMedium   DecisionImpact = "medium"
	DecisionImpactHigh     DecisionImpact = "high"
	DecisionImpactCritical DecisionImpact = "critical"
)

// DecisionOutcome represents the outcome of a decision
type DecisionOutcome struct {
	Success     bool               `json:"success"`
	Description string             `json:"description"`
	Metrics     map[string]float64 `json:"metrics,omitempty"`
	Timestamp   time.Time          `json:"timestamp"`
}

// UnresolvedIssue represents an unresolved issue from the conversation
type UnresolvedIssue struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	Context     []string  `json:"context"`
	FirstSeen   time.Time `json:"firstSeen"`
	LastSeen    time.Time `json:"lastSeen"`
}

// ContextCompression represents configuration for context compression
type ContextCompression struct {
	Algorithm      CompressionAlgorithm `json:"algorithm"`
	Level          int                  `json:"level"`        // 1-9, higher = more compression
	MinSize        int64                `json:"minSize"`      // Only compress if larger than this
	PreserveInfo   []InfoType           `json:"preserveInfo"` // Info types to preserve
	AggressiveMode bool                 `json:"aggressiveMode"`
}

// CompressionAlgorithm represents compression algorithms
type CompressionAlgorithm string

const (
	CompressionAlgorithmGzip   CompressionAlgorithm = "gzip"
	CompressionAlgorithmLZ4    CompressionAlgorithm = "lz4"
	CompressionAlgorithmBrotli CompressionAlgorithm = "brotli"
	CompressionAlgorithmSmart  CompressionAlgorithm = "smart" // Intelligent compression
)

// ContextMetrics represents metrics about context usage
type ContextMetrics struct {
	GenerationTime   time.Duration `json:"generationTime"`
	CompressionRatio float64       `json:"compressionRatio"`
	RelevanceScore   float64       `json:"relevanceScore"`
	UsageCount       int           `json:"usageCount"`
	HitRate          float64       `json:"hitRate"`    // Cache hit rate
	Efficiency       float64       `json:"efficiency"` // Overall efficiency score
	LastUpdated      time.Time     `json:"lastUpdated"`
}

// Configuration Types for Context System

// ContextManagerConfig represents configuration for the context manager
type ContextManagerConfig struct {
	MaxContexts       int                     `json:"maxContexts"`
	DefaultTTL        int64                   `json:"defaultTtl"` // seconds
	CompressionConfig *ContextCompression     `json:"compressionConfig"`
	CacheConfig       *ContextCacheConfig     `json:"cacheConfig"`
	IndexConfig       *IndexConfig            `json:"indexConfig"`
	ValidationConfig  *ContextValidationRules `json:"validationConfig"`
	MetricsEnabled    bool                    `json:"metricsEnabled"`
}

// IndexConfig represents indexing configuration
type IndexConfig struct {
	Enabled          bool   `json:"enabled"`
	RebuildInterval  string `json:"rebuildInterval"`
	OptimizeInterval string `json:"optimizeInterval"`
	MaxIndexSize     int64  `json:"maxIndexSize"` // bytes
}

// ContextValidationRules represents validation rules for contexts
type ContextValidationRules struct {
	MaxSize        int64         `json:"maxSize"` // bytes
	MaxTokens      int           `json:"maxTokens"`
	RequiredFields []string      `json:"requiredFields"`
	AllowedTypes   []ContextType `json:"allowedTypes"`
}

// ContextSearchCriteria represents search criteria for contexts
type ContextSearchCriteria struct {
	Type         ContextType  `json:"type,omitempty"`
	Scope        ContextScope `json:"scope,omitempty"`
	Keywords     []string     `json:"keywords,omitempty"`
	Tags         []string     `json:"tags,omitempty"`
	DateRange    *DateRange   `json:"dateRange,omitempty"`
	SizeRange    *SizeRange   `json:"sizeRange,omitempty"`
	MinRelevance float64      `json:"minRelevance,omitempty"`
	Limit        int          `json:"limit,omitempty"`
}

// ContextSystemMetrics represents system-wide context metrics
type ContextSystemMetrics struct {
	TotalContexts    int       `json:"totalContexts"`
	ActiveContexts   int       `json:"activeContexts"`
	CacheHitRate     float64   `json:"cacheHitRate"`
	AverageSize      int64     `json:"averageSize"`
	CompressionRatio float64   `json:"compressionRatio"`
	LastUpdated      time.Time `json:"lastUpdated"`
}

// AdaptationRules represents rules for adaptive context generation
type AdaptationRules struct {
	DynamicPriority  bool              `json:"dynamicPriority"`
	ContextAwareness bool              `json:"contextAwareness"`
	UserPreferences  map[string]string `json:"userPreferences"`
	TaskType         TaskType          `json:"taskType"`
	QualityThreshold float64           `json:"qualityThreshold"`
	AdaptationLevel  AdaptationLevel   `json:"adaptationLevel"`
}

// AdaptationLevel represents the level of adaptation
type AdaptationLevel string

const (
	AdaptationLevelMinimal    AdaptationLevel = "minimal"
	AdaptationLevelModerate   AdaptationLevel = "moderate"
	AdaptationLevelAggressive AdaptationLevel = "aggressive"
	AdaptationLevelMaximal    AdaptationLevel = "maximal"
)

// ContextEnhancements represents enhancements to apply to a context
type ContextEnhancements struct {
	AddRelevantInfo    []RelevantInfo         `json:"addRelevantInfo,omitempty"`
	EnrichMetadata     map[string]interface{} `json:"enrichMetadata,omitempty"`
	ExpandKeywords     []string               `json:"expandKeywords,omitempty"`
	ImproveQuality     bool                   `json:"improveQuality"`
	AddCrossReferences bool                   `json:"addCrossReferences"`
	EnhanceStructure   bool                   `json:"enhanceStructure"`
}

// MergeStrategy represents strategies for merging contexts
type MergeStrategy string

const (
	MergeStrategyUnion        MergeStrategy = "union"
	MergeStrategyIntersection MergeStrategy = "intersection"
	MergeStrategyPriority     MergeStrategy = "priority"
	MergeStrategyWeighted     MergeStrategy = "weighted"
	MergeStrategySmart        MergeStrategy = "smart"
)

// ContextGeneratorConfig represents configuration for context generation
type ContextGeneratorConfig struct {
	DefaultOptions     *ContextOptions `json:"defaultOptions"`
	MaxConcurrency     int             `json:"maxConcurrency"`
	TimeoutConfig      *TimeoutConfig  `json:"timeoutConfig"`
	QualityThreshold   float64         `json:"qualityThreshold"`
	AdaptiveGeneration bool            `json:"adaptiveGeneration"`
}

// TimeoutConfig represents timeout configuration
type TimeoutConfig struct {
	Default    int64 `json:"default"`    // milliseconds
	Maximum    int64 `json:"maximum"`    // milliseconds
	Generation int64 `json:"generation"` // milliseconds
	Processing int64 `json:"processing"` // milliseconds
}

// CompressionAnalysis represents analysis of compression effectiveness
type CompressionAnalysis struct {
	OriginalSize     int64   `json:"originalSize"`
	CompressedSize   int64   `json:"compressedSize"`
	CompressionRatio float64 `json:"compressionRatio"`
	TimeSpent        string  `json:"timeSpent"`
	SpaceSaved       int64   `json:"spaceSaved"`
	Efficiency       float64 `json:"efficiency"`
}

// CompressionValidation represents validation of compression configuration
type CompressionValidation struct {
	Valid       bool     `json:"valid"`
	Errors      []string `json:"errors,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// CompressionConfig represents configuration for compression
type CompressionConfig struct {
	Enabled        bool                 `json:"enabled"`
	Algorithm      CompressionAlgorithm `json:"algorithm"`
	Level          int                  `json:"level"`
	MinSize        int64                `json:"minSize"`
	PreserveInfo   []InfoType           `json:"preserveInfo"`
	AggressiveMode bool                 `json:"aggressiveMode"`
}

// ContextCacheConfig represents configuration for context cache
type ContextCacheConfig struct {
	MaxSize          int              `json:"maxSize"`
	TTL              int64            `json:"ttl"` // seconds
	EvictionStrategy EvictionStrategy `json:"evictionStrategy"`
}

// Entity represents an entity extracted from context
type Entity struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Value      string                 `json:"value"`
	Confidence float64                `json:"confidence"`
	Location   string                 `json:"location"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ContextPattern represents a pattern found in context
type ContextPattern struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Pattern     string  `json:"pattern"`
	Frequency   int     `json:"frequency"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

// CodeElements represents code elements found in context
type CodeElements struct {
	Functions  []string `json:"functions"`
	Classes    []string `json:"classes"`
	Variables  []string `json:"variables"`
	Imports    []string `json:"imports"`
	Comments   []string `json:"comments"`
	Interfaces []string `json:"interfaces"`
	Types      []string `json:"types"`
}

// Dependency represents a dependency in the project
type Dependency struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Type         string   `json:"type"`
	Description  string   `json:"description"`
	Repository   string   `json:"repository,omitempty"`
	License      string   `json:"license,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// ArchitectureInfo represents architectural information
type ArchitectureInfo struct {
	Style      string   `json:"style"`
	Patterns   []string `json:"patterns"`
	Components []string `json:"components"`
	Layers     []string `json:"layers"`
	Principles []string `json:"principles"`
}

// SentimentAnalysis represents sentiment analysis of context
type SentimentAnalysis struct {
	Score      float64 `json:"score"`      // -1 to 1
	Magnitude  float64 `json:"magnitude"`  // 0 to 1
	Label      string  `json:"label"`      // positive, negative, neutral
	Confidence float64 `json:"confidence"` // 0 to 1
}

// ComplexityAnalysis represents complexity analysis of context
type ComplexityAnalysis struct {
	CyclomaticComplexity int     `json:"cyclomaticComplexity"`
	CognitiveComplexity  int     `json:"cognitiveComplexity"`
	LinesOfCode          int     `json:"linesOfCode"`
	FunctionCount        int     `json:"functionCount"`
	ClassCount           int     `json:"classCount"`
	Score                float64 `json:"score"`
	Level                string  `json:"level"` // low, medium, high
}

// TopicAnalysis represents topic analysis of context
type TopicAnalysis struct {
	Topics     []Topic  `json:"topics"`
	Keywords   []string `json:"keywords"`
	Categories []string `json:"categories"`
	Confidence float64  `json:"confidence"`
	Language   string   `json:"language"`
}

// ExtractionConfig represents configuration for context extraction
type ExtractionConfig struct {
	MaxDepth        int      `json:"maxDepth"`
	IncludeComments bool     `json:"includeComments"`
	IncludeTests    bool     `json:"includeTests"`
	FileTypes       []string `json:"fileTypes"`
	ExcludePatterns []string `json:"excludePatterns"`
	MinRelevance    float64  `json:"minRelevance"`
}

// ContextSearchQuery represents a search query for contexts
type ContextSearchQuery struct {
	Query     string         `json:"query"`
	Type      ContextType    `json:"type,omitempty"`
	Scope     ContextScope   `json:"scope,omitempty"`
	Keywords  []string       `json:"keywords,omitempty"`
	Tags      []string       `json:"tags,omitempty"`
	DateRange *DateRange     `json:"dateRange,omitempty"`
	SizeRange *SizeRange     `json:"sizeRange,omitempty"`
	Filters   *SearchFilters `json:"filters,omitempty"`
	Options   *SearchOptions `json:"options,omitempty"`
	Limit     int            `json:"limit,omitempty"`
	Offset    int            `json:"offset,omitempty"`
	SortBy    string         `json:"sortBy,omitempty"`
	SortOrder string         `json:"sortOrder,omitempty"`
}

// SearchResults represents search results
type SearchResults struct {
	Results     []SearchResult `json:"results"`
	Total       int            `json:"total"`
	Page        int            `json:"page"`
	PageSize    int            `json:"pageSize"`
	HasMore     bool           `json:"hasMore"`
	Query       string         `json:"query"`
	TimeSpent   string         `json:"timeSpent"`
	Suggestions []string       `json:"suggestions,omitempty"`
}

// SearchFilters represents search filters
type SearchFilters struct {
	MinConfidence float64    `json:"minConfidence,omitempty"`
	MinRelevance  float64    `json:"minRelevance,omitempty"`
	Authors       []string   `json:"authors,omitempty"`
	Sources       []string   `json:"sources,omitempty"`
	Categories    []string   `json:"categories,omitempty"`
	DateRange     *DateRange `json:"dateRange,omitempty"`
	SizeRange     *SizeRange `json:"sizeRange,omitempty"`
}

// SearchOptions represents search options
type SearchOptions struct {
	Fuzzy            bool    `json:"fuzzy"`
	CaseSensitive    bool    `json:"caseSensitive"`
	WholeWords       bool    `json:"wholeWords"`
	IncludeContent   bool    `json:"includeContent"`
	IncludeMetadata  bool    `json:"includeMetadata"`
	HighlightMatches bool    `json:"highlightMatches"`
	MaxResultLength  int     `json:"maxResultLength"`
	MinScore         float64 `json:"minScore"`
}

// SemanticSearchOptions represents semantic search options
type SemanticSearchOptions struct {
	Model               string  `json:"model,omitempty"`
	EmbeddingModel      string  `json:"embeddingModel,omitempty"`
	SimilarityThreshold float64 `json:"similarityThreshold"`
	MinSimilarity       float64 `json:"minSimilarity"`
	MaxResults          int     `json:"maxResults"`
	IncludeScore        bool    `json:"includeScore"`
	ContextWindow       int     `json:"contextWindow"`
}

// SearchResult represents a single search result (defined separately from memory.go)
type ContextSearchResult struct {
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

// IndexStats represents indexing statistics
type IndexStats struct {
	TotalDocuments   int       `json:"totalDocuments"`
	IndexedDocuments int       `json:"indexedDocuments"`
	IndexSize        int64     `json:"indexSize"`
	LastIndexed      time.Time `json:"lastIndexed"`
	IndexingTime     string    `json:"indexingTime"`
	ErrorCount       int       `json:"errorCount"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	HitCount      int     `json:"hitCount"`
	MissCount     int     `json:"missCount"`
	HitRate       float64 `json:"hitRate"`
	Size          int64   `json:"size"`
	MaxSize       int64   `json:"maxSize"`
	EntryCount    int     `json:"entryCount"`
	EvictionCount int     `json:"evictionCount"`
}

// EvictionStrategy represents cache eviction strategies
type EvictionStrategy string

const (
	EvictionStrategyLRU    EvictionStrategy = "lru"
	EvictionStrategyFIFO   EvictionStrategy = "fifo"
	EvictionStrategyRandom EvictionStrategy = "random"
	EvictionStrategyLFU    EvictionStrategy = "lfu"
	EvictionStrategyTTL    EvictionStrategy = "ttl"
)

// QualityValidation represents quality validation results
type QualityValidation struct {
	Valid        bool     `json:"valid"`
	Score        float64  `json:"score"`
	Issues       []string `json:"issues,omitempty"`
	Suggestions  []string `json:"suggestions,omitempty"`
	Completeness float64  `json:"completeness"`
	Accuracy     float64  `json:"accuracy"`
	Consistency  float64  `json:"consistency"`
}

// CompletenessRequirements represents completeness requirements
type CompletenessRequirements struct {
	RequiredFields   []string `json:"requiredFields"`
	MinContentLength int      `json:"minContentLength"`
	RequiredSections []string `json:"requiredSections"`
	MinKeywords      int      `json:"minKeywords"`
	RequiredMetadata []string `json:"requiredMetadata"`
}

// ConsistencyValidation represents consistency validation
type ConsistencyValidation struct {
	Valid              bool     `json:"valid"`
	InconsistentFields []string `json:"inconsistentFields,omitempty"`
	Conflicts          []string `json:"conflicts,omitempty"`
	Score              float64  `json:"score"`
	Issues             []string `json:"issues,omitempty"`
}

// SecurityValidation represents security validation
type SecurityValidation struct {
	Valid           bool     `json:"valid"`
	Threats         []string `json:"threats,omitempty"`
	Vulnerabilities []string `json:"vulnerabilities,omitempty"`
	RiskLevel       string   `json:"riskLevel"`
	SecurityScore   float64  `json:"securityScore"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// EntityType represents the type of entity
type EntityType string

const (
	EntityTypePerson       EntityType = "person"
	EntityTypeOrganization EntityType = "organization"
	EntityTypeLocation     EntityType = "location"
	EntityTypeDate         EntityType = "date"
	EntityTypeNumber       EntityType = "number"
	EntityTypeEmail        EntityType = "email"
	EntityTypeURL          EntityType = "url"
	EntityTypeCode         EntityType = "code"
	EntityTypeFunction     EntityType = "function"
	EntityTypeVariable     EntityType = "variable"
	EntityTypeClass        EntityType = "class"
)

// Position represents a position in text
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Offset int `json:"offset"`
}

// FunctionInfo represents information about a function
type FunctionInfo struct {
	Name        string    `json:"name"`
	Signature   string    `json:"signature"`
	Parameters  []string  `json:"parameters"`
	ReturnType  string    `json:"returnType"`
	Description string    `json:"description"`
	StartPos    *Position `json:"startPos,omitempty"`
	EndPos      *Position `json:"endPos,omitempty"`
	Complexity  int       `json:"complexity"`
	LinesOfCode int       `json:"linesOfCode"`
}

// ClassInfo represents information about a class
type ClassInfo struct {
	Name        string         `json:"name"`
	Signature   string         `json:"signature"`
	Methods     []FunctionInfo `json:"methods"`
	Properties  []VariableInfo `json:"properties"`
	Inheritance []string       `json:"inheritance"`
	Description string         `json:"description"`
	StartPos    *Position      `json:"startPos,omitempty"`
	EndPos      *Position      `json:"endPos,omitempty"`
	LinesOfCode int            `json:"linesOfCode"`
}

// VariableInfo represents information about a variable
type VariableInfo struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Value       string    `json:"value,omitempty"`
	Description string    `json:"description"`
	StartPos    *Position `json:"startPos,omitempty"`
	EndPos      *Position `json:"endPos,omitempty"`
	Scope       string    `json:"scope"`
	Mutable     bool      `json:"mutable"`
}

// ImportInfo represents information about an import statement
type ImportInfo struct {
	Module   string    `json:"module"`
	Alias    string    `json:"alias,omitempty"`
	Imported []string  `json:"imported,omitempty"`
	Path     string    `json:"path"`
	StartPos *Position `json:"startPos,omitempty"`
	EndPos   *Position `json:"endPos,omitempty"`
	Type     string    `json:"type"` // default, named, namespace, dynamic
}

// ExportInfo represents information about an export statement
type ExportInfo struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Value       string    `json:"value,omitempty"`
	Description string    `json:"description"`
	StartPos    *Position `json:"startPos,omitempty"`
	EndPos      *Position `json:"endPos,omitempty"`
	Default     bool      `json:"default"`
}

// CommentInfo represents information about a comment
type CommentInfo struct {
	Content    string    `json:"content"`
	Type       string    `json:"type"` // line, block, doc
	StartPos   *Position `json:"startPos,omitempty"`
	EndPos     *Position `json:"endPos,omitempty"`
	Associated string    `json:"associated,omitempty"` // function/class this comment is associated with
}

// AnnotationInfo represents information about annotations/decorators
type AnnotationInfo struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Type      string                 `json:"type"` // decorator, annotation, attribute
	StartPos  *Position              `json:"startPos,omitempty"`
	EndPos    *Position              `json:"endPos,omitempty"`
	Target    string                 `json:"target"` // what this annotation applies to
}

// Parameter represents a function parameter
type Parameter struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Optional     bool   `json:"optional"`
	Description  string `json:"description,omitempty"`
}

// DependencyType represents the type of dependency
type DependencyType string

const (
	DependencyTypeDirect    DependencyType = "direct"
	DependencyTypeIndirect  DependencyType = "indirect"
	DependencyTypeDev       DependencyType = "dev"
	DependencyTypePeer      DependencyType = "peer"
	DependencyTypeOptional  DependencyType = "optional"
	DependencyTypeRuntime   DependencyType = "runtime"
	DependencyTypeBuildTime DependencyType = "build_time"
)

// ArchitectureLayer represents an architectural layer
type ArchitectureLayer struct {
	Name             string   `json:"name"`
	Level            int      `json:"level"`
	Description      string   `json:"description"`
	Components       []string `json:"components"`
	Dependencies     []string `json:"dependencies"`
	Responsibilities []string `json:"responsibilities"`
	Patterns         []string `json:"patterns"`
}

// ArchitectureQuality represents quality metrics for architecture
type ArchitectureQuality struct {
	Maintainability float64 `json:"maintainability"`
	Scalability     float64 `json:"scalability"`
	Performance     float64 `json:"performance"`
	Security        float64 `json:"security"`
	Testability     float64 `json:"testability"`
	Coupling        float64 `json:"coupling"`   // Lower is better
	Cohesion        float64 `json:"cohesion"`   // Higher is better
	Complexity      float64 `json:"complexity"` // Lower is better
	Overall         float64 `json:"overall"`
}

// Sentiment represents sentiment information
type Sentiment struct {
	Score      float64 `json:"score"`      // -1 to 1
	Magnitude  float64 `json:"magnitude"`  // 0 to 1
	Label      string  `json:"label"`      // positive, negative, neutral
	Confidence float64 `json:"confidence"` // 0 to 1
}

// AspectSentiment represents sentiment for specific aspects
type AspectSentiment struct {
	Aspect    string    `json:"aspect"`
	Sentiment Sentiment `json:"sentiment"`
	Keywords  []string  `json:"keywords,omitempty"`
}

// Component represents a component in the system
type Component struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	Description      string                 `json:"description"`
	Responsibilities []string               `json:"responsibilities"`
	Dependencies     []string               `json:"dependencies"`
	Location         string                 `json:"location"`
	Status           string                 `json:"status"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}
