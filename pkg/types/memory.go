package types

import (
	"time"
)

// Memory System Core Types
// This file contains the core memory system types and configurations.
// Detailed types are split across multiple files:
// - memory_knowledge.go: Knowledge management types
// - memory_patterns.go: Pattern recognition types
// - memory_search.go: Search and document types
// - memory_learning.go: Learning and feedback types
// - memory_analysis.go: Analysis and optimization types

// MemoryManager represents the unified memory manager
type MemoryManager struct {
	ID           string                 `json:"id"`
	Config       *MemoryManagerConfig   `json:"config"`
	Metrics      *MemoryMetrics         `json:"metrics"`
	Status       MemoryStatus           `json:"status"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
	LastUpdated  time.Time              `json:"lastUpdated"`
}

// Note: MemoryStatus is defined in core.go

// MemoryManagerConfig represents configuration for memory manager
type MemoryManagerConfig struct {
	KnowledgeBase      *KnowledgeBaseConfig   `json:"knowledgeBase"`
	PatternLearner     *PatternLearnerConfig  `json:"patternLearner"`
	LessonLearner      *LessonLearnerConfig   `json:"lessonLearner"`
	DocumentManager    *DocumentManagerConfig `json:"documentManager"`
	ProjectMemory      *ProjectMemoryConfig   `json:"projectMemory"`
	StorageConfig      *StorageConfig         `json:"storageConfig"`
	SearchConfig       *SearchEngineConfig    `json:"searchConfig"`
	SecurityConfig     *MemorySecurityConfig  `json:"securityConfig"`
	OptimizationConfig *OptimizationConfig    `json:"optimizationConfig"`
	MetricsEnabled     bool                   `json:"metricsEnabled"`
	CacheEnabled       bool                   `json:"cacheEnabled"`
	BackupEnabled      bool                   `json:"backupEnabled"`
	CompressionEnabled bool                   `json:"compressionEnabled"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Type             string                 `json:"type"` // file, database, cloud
	ConnectionString string                 `json:"connectionString"`
	MaxSize          int64                  `json:"maxSize"` // bytes
	RetentionPeriod  string                 `json:"retentionPeriod"`
	BackupConfig     *BackupConfig          `json:"backupConfig"`
	Encryption       *EncryptionConfig      `json:"encryption"`
	Compression      *CompressionConfig     `json:"compression"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	Enabled          bool   `json:"enabled"`
	Interval         string `json:"interval"`
	MaxBackups       int    `json:"maxBackups"`
	CompressionLevel int    `json:"compressionLevel"`
	Destination      string `json:"destination"`
	Encrypted        bool   `json:"encrypted"`
}

// EncryptionConfig represents encryption configuration
type EncryptionConfig struct {
	Enabled     bool   `json:"enabled"`
	Algorithm   string `json:"algorithm"`
	KeySize     int    `json:"keySize"`
	KeyRotation bool   `json:"keyRotation"`
}

// MemorySecurityConfig represents security configuration for memory
type MemorySecurityConfig struct {
	AccessControl       bool     `json:"accessControl"`
	EncryptionAtRest    bool     `json:"encryptionAtRest"`
	EncryptionInTransit bool     `json:"encryptionInTransit"`
	AuditLogging        bool     `json:"auditLogging"`
	AllowedUsers        []string `json:"allowedUsers"`
	RestrictedData      []string `json:"restrictedData"`
	DataRetention       string   `json:"dataRetention"`
	Anonymization       bool     `json:"anonymization"`
}

// OptimizationConfig represents optimization configuration
type OptimizationConfig struct {
	AutoOptimization     bool                `json:"autoOptimization"`
	OptimizationInterval string              `json:"optimizationInterval"`
	ConsolidationRules   *ConsolidationRules `json:"consolidationRules"`
	ForgetCriteria       *ForgetCriteria     `json:"forgetCriteria"`
	CompressionThreshold int64               `json:"compressionThreshold"`
	PerformanceThreshold float64             `json:"performanceThreshold"`
}

// MemoryMetrics represents metrics about memory usage
type MemoryMetrics struct {
	TotalKnowledge     int            `json:"totalKnowledge"`
	KnowledgeByType    map[string]int `json:"knowledgeByType"`
	AverageRelevance   float64        `json:"averageRelevance"`
	AverageConfidence  float64        `json:"averageConfidence"`
	StorageUsed        int64          `json:"storageUsed"`
	RetrievalLatency   time.Duration  `json:"retrievalLatency"`
	ConsolidationRate  float64        `json:"consolidationRate"`
	ForgetRate         float64        `json:"forgetRate"`
	LearningEfficiency float64        `json:"learningEfficiency"`
	KnowledgeCount     int            `json:"knowledgeCount"`
	PatternCount       int            `json:"patternCount"`
	LessonCount        int            `json:"lessonCount"`
	DocumentCount      int            `json:"documentCount"`
	LastUpdated        time.Time      `json:"lastUpdated"`
}

// ProjectMemoryConfig represents configuration for project memory
type ProjectMemoryConfig struct {
	AutoSnapshot          bool   `json:"autoSnapshot"`
	SnapshotInterval      string `json:"snapshotInterval"`
	MaxSnapshots          int    `json:"maxSnapshots"`
	CompressionEnabled    bool   `json:"compressionEnabled"`
	ArchitectureTracking  bool   `json:"architectureTracking"`
	DecisionTracking      bool   `json:"decisionTracking"`
	LessonLearning        bool   `json:"lessonLearning"`
	ConfigurationTracking bool   `json:"configurationTracking"`
}

// MemorySnapshot represents a snapshot of memory state
type MemorySnapshot struct {
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"projectId"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Size        int64                  `json:"size"`
	Compressed  bool                   `json:"compressed"`
	Checksum    string                 `json:"checksum"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	ExpiresAt   *time.Time             `json:"expiresAt,omitempty"`
}

// SnapshotComparison represents comparison between memory snapshots
type SnapshotComparison struct {
	Snapshot1ID  string                 `json:"snapshot1Id"`
	Snapshot2ID  string                 `json:"snapshot2Id"`
	Differences  []SnapshotDifference   `json:"differences"`
	Summary      string                 `json:"summary"`
	ChangedItems int                    `json:"changedItems"`
	AddedItems   int                    `json:"addedItems"`
	RemovedItems int                    `json:"removedItems"`
	SizeChange   int64                  `json:"sizeChange"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// SnapshotDifference represents a difference between snapshots
type SnapshotDifference struct {
	Type         string `json:"type"` // added, modified, removed
	Item         string `json:"item"`
	Category     string `json:"category"`
	OldValue     string `json:"oldValue,omitempty"`
	NewValue     string `json:"newValue,omitempty"`
	Significance string `json:"significance"` // low, medium, high
}

// ArchitecturalChange represents a change in project architecture
type ArchitecturalChange struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // component_added, layer_modified, etc.
	Component   string                 `json:"component"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact"` // low, medium, high
	Rationale   string                 `json:"rationale"`
	Before      map[string]interface{} `json:"before,omitempty"`
	After       map[string]interface{} `json:"after,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ConfigurationChange represents a change in project configuration
type ConfigurationChange struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`      // added, modified, removed
	Component string                 `json:"component"` // build, deployment, etc.
	Setting   string                 `json:"setting"`
	OldValue  string                 `json:"oldValue,omitempty"`
	NewValue  string                 `json:"newValue,omitempty"`
	Impact    string                 `json:"impact"` // low, medium, high
	Rationale string                 `json:"rationale"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// DecisionFilters represents filters for decision searches
type DecisionFilters struct {
	Status       []string         `json:"status,omitempty"`
	Impact       []DecisionImpact `json:"impact,omitempty"`
	Tags         []string         `json:"tags,omitempty"`
	Stakeholders []string         `json:"stakeholders,omitempty"`
	TimeRange    *TimeRange       `json:"timeRange,omitempty"`
	Categories   []string         `json:"categories,omitempty"`
}

// ProjectMemory represents memory associated with a specific project
type ProjectMemory struct {
	ID            string                 `json:"id"`
	ProjectID     string                 `json:"projectId"`
	ProjectPath   string                 `json:"projectPath"`
	Architecture  *ProjectArchitecture   `json:"architecture"`
	Decisions     []ProjectDecision      `json:"decisions"`
	Lessons       []ProjectLesson        `json:"lessons"`
	Configuration *ProjectConfiguration  `json:"configuration"`
	Knowledge     []Knowledge            `json:"knowledge"`
	Patterns      []CodePattern          `json:"patterns"`
	Snapshots     []MemorySnapshot       `json:"snapshots"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
	LastUpdated   time.Time              `json:"lastUpdated"`
}

// ProjectArchitecture represents the architecture of a project
type ProjectArchitecture struct {
	ID            string                    `json:"id"`
	ProjectID     string                    `json:"projectId"`
	Name          string                    `json:"name"`
	Description   string                    `json:"description"`
	Version       string                    `json:"version"`
	Components    []ArchitecturalComponent  `json:"components"`
	Layers        []ArchitecturalLayer      `json:"layers"`
	Dependencies  []ArchitecturalDependency `json:"dependencies"`
	Patterns      []string                  `json:"patterns"`
	Technologies  []string                  `json:"technologies"`
	Principles    []string                  `json:"principles"`
	Constraints   []string                  `json:"constraints"`
	Documentation string                    `json:"documentation"`
	Metadata      map[string]interface{}    `json:"metadata,omitempty"`
	CreatedAt     time.Time                 `json:"createdAt"`
	LastUpdated   time.Time                 `json:"lastUpdated"`
}

// ArchitecturalComponent represents a component in the architecture
type ArchitecturalComponent struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	Description      string                 `json:"description"`
	Responsibilities []string               `json:"responsibilities"`
	Interfaces       []ComponentInterface   `json:"interfaces"`
	Dependencies     []string               `json:"dependencies"`
	Location         string                 `json:"location"`
	Status           string                 `json:"status"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ComponentInterface represents an interface of a component
type ComponentInterface struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Methods     []string `json:"methods"`
	Protocol    string   `json:"protocol"`
}

// ArchitecturalLayer represents a layer in the architecture
type ArchitecturalLayer struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Level            int      `json:"level"`
	Components       []string `json:"components"`
	Responsibilities []string `json:"responsibilities"`
}

// ArchitecturalDependency represents a dependency between components
type ArchitecturalDependency struct {
	ID          string `json:"id"`
	From        string `json:"from"`
	To          string `json:"to"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Strength    string `json:"strength"`
}

// ProjectDecision represents a decision made in the project
type ProjectDecision struct {
	ID             string                 `json:"id"`
	ProjectID      string                 `json:"projectId"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Status         DecisionStatus         `json:"status"`
	Impact         DecisionImpact         `json:"impact"`
	Category       string                 `json:"category"`
	Stakeholders   []string               `json:"stakeholders"`
	Alternatives   []DecisionAlternative  `json:"alternatives"`
	SelectedOption string                 `json:"selectedOption"`
	Rationale      string                 `json:"rationale"`
	Consequences   []string               `json:"consequences"`
	ReviewDate     *time.Time             `json:"reviewDate,omitempty"`
	Tags           []string               `json:"tags"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"createdAt"`
	DecidedAt      *time.Time             `json:"decidedAt,omitempty"`
	LastReviewed   *time.Time             `json:"lastReviewed,omitempty"`
}

// DecisionStatus represents the status of a decision
type DecisionStatus string

const (
	DecisionStatusProposed DecisionStatus = "proposed"
	DecisionStatusApproved DecisionStatus = "approved"
	DecisionStatusRejected DecisionStatus = "rejected"
	DecisionStatusDeferred DecisionStatus = "deferred"
	DecisionStatusObsolete DecisionStatus = "obsolete"
)

// DecisionImpact is imported from context.go

// DecisionAlternative represents an alternative for a decision
type DecisionAlternative struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Pros        []string `json:"pros"`
	Cons        []string `json:"cons"`
	Cost        float64  `json:"cost"`
	Risk        string   `json:"risk"`
	Feasibility string   `json:"feasibility"`
}

// ProjectLesson represents a lesson learned in the project
type ProjectLesson struct {
	ID              string                 `json:"id"`
	ProjectID       string                 `json:"projectId"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Category        string                 `json:"category"`
	Impact          LessonImpact           `json:"impact"`
	Source          string                 `json:"source"`
	Context         string                 `json:"context"`
	Lesson          string                 `json:"lesson"`
	Recommendations []string               `json:"recommendations"`
	Tags            []string               `json:"tags"`
	Verified        bool                   `json:"verified"`
	Applied         bool                   `json:"applied"`
	Effectiveness   float64                `json:"effectiveness"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"createdAt"`
	VerifiedAt      *time.Time             `json:"verifiedAt,omitempty"`
	AppliedAt       *time.Time             `json:"appliedAt,omitempty"`
}

// LessonImpact is imported from memory_learning.go

// ProjectConfiguration represents project configuration
type ProjectConfiguration struct {
	ID               string                 `json:"id"`
	ProjectID        string                 `json:"projectId"`
	Name             string                 `json:"name"`
	Version          string                 `json:"version"`
	Environment      string                 `json:"environment"`
	BuildConfig      map[string]interface{} `json:"buildConfig"`
	DeploymentConfig map[string]interface{} `json:"deploymentConfig"`
	Dependencies     []ProjectDependency    `json:"dependencies"`
	Settings         map[string]interface{} `json:"settings"`
	Secrets          []string               `json:"secrets"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"createdAt"`
	LastUpdated      time.Time              `json:"lastUpdated"`
}

// ProjectDependency represents a project dependency
type ProjectDependency struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	Required    bool   `json:"required"`
	Development bool   `json:"development"`
}
