package types

import (
	"time"
)

// MemoryUsageAnalysis represents analysis of memory usage patterns
type MemoryUsageAnalysis struct {
	Period          string                 `json:"period"`
	TotalUsage      int64                  `json:"totalUsage"`
	UsageByType     map[string]int64       `json:"usageByType"`
	AccessPatterns  []AccessPattern        `json:"accessPatterns"`
	GrowthRate      float64                `json:"growthRate"`
	EfficiencyScore float64                `json:"efficiencyScore"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

// AccessPattern represents a pattern in memory access
type AccessPattern struct {
	Type        string    `json:"type"`
	Frequency   int       `json:"frequency"`
	Times       []string  `json:"times"`       // time patterns
	UserGroups  []string  `json:"userGroups"`
	Resources   []string  `json:"resources"`
	Description string    `json:"description"`
}

// MemoryPattern represents a pattern found in memory
type MemoryPattern struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Pattern     string                 `json:"pattern"`
	Occurrences int                    `json:"occurrences"`
	Confidence  float64                `json:"confidence"`
	Significance float64               `json:"significance"`
	Context     []string               `json:"context"`
	Examples    []string               `json:"examples"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	DiscoveredAt time.Time             `json:"discoveredAt"`
}

// DecisionPatternAnalysis represents analysis of decision patterns
type DecisionPatternAnalysis struct {
	ProjectID       string                 `json:"projectId"`
	Patterns        []DecisionPattern      `json:"patterns"`
	CommonFactors   []string               `json:"commonFactors"`
	SuccessFactors  []string               `json:"successFactors"`
	RiskFactors     []string               `json:"riskFactors"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

// DecisionPattern represents a pattern in decision making
type DecisionPattern struct {
	Type        string   `json:"type"`
	Frequency   int      `json:"frequency"`
	SuccessRate float64  `json:"successRate"`
	Context     []string `json:"context"`
	Outcomes    []string `json:"outcomes"`
}

// AnalysisScope represents scope for analysis
type AnalysisScope struct {
	Type        string     `json:"type"`        // project, global, domain
	ProjectIDs  []string   `json:"projectIds,omitempty"`
	Domains     []string   `json:"domains,omitempty"`
	TimeRange   *TimeRange `json:"timeRange,omitempty"`
	Categories  []string   `json:"categories,omitempty"`
	UserGroups  []string   `json:"userGroups,omitempty"`
}

// MemoryQualityAssessment represents assessment of memory quality
type MemoryQualityAssessment struct {
	OverallScore    float64                `json:"overallScore"`
	AccuracyScore   float64                `json:"accuracyScore"`
	CompletenessScore float64              `json:"completenessScore"`
	ConsistencyScore float64               `json:"consistencyScore"`
	FreshnessScore  float64                `json:"freshnessScore"`
	UsabilityScore  float64                `json:"usabilityScore"`
	IssuesByCategory map[string][]string   `json:"issuesByCategory"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

// DuplicateGroup represents a group of duplicate items
type DuplicateGroup struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Items       []string  `json:"items"`        // IDs of duplicate items
	Similarity  float64   `json:"similarity"`   // similarity score
	Primary     string    `json:"primary"`      // suggested primary item
	Action      string    `json:"action"`       // merge, remove, keep
	Confidence  float64   `json:"confidence"`
	DetectedAt  time.Time `json:"detectedAt"`
}

// ConsistencyAnalysis represents analysis of consistency
type ConsistencyAnalysis struct {
	Domain           string                 `json:"domain"`
	ConsistencyScore float64                `json:"consistencyScore"`
	Inconsistencies  []Inconsistency        `json:"inconsistencies"`
	Categories       map[string]float64     `json:"categories"`    // category -> consistency score
	Recommendations  []string               `json:"recommendations"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
}

// Inconsistency represents an inconsistency found
type Inconsistency struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Items       []string `json:"items"`       // conflicting items
	Severity    string   `json:"severity"`    // low, medium, high
	Category    string   `json:"category"`
	Suggestion  string   `json:"suggestion"`
}

// MemoryInsight represents an insight derived from memory analysis
type MemoryInsight struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Impact      string                 `json:"impact"`      // low, medium, high
	Category    string                 `json:"category"`
	Evidence    []string               `json:"evidence"`
	Actions     []string               `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
}

// InsightContext represents context for generating insights
type InsightContext struct {
	Scope       *AnalysisScope         `json:"scope"`
	Focus       []string               `json:"focus,omitempty"`
	Goals       []string               `json:"goals,omitempty"`
	Constraints []string               `json:"constraints,omitempty"`
	UserProfile *UserProfile           `json:"userProfile,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MemoryNeedsPrediction represents prediction of memory needs
type MemoryNeedsPrediction struct {
	ProjectID       string                 `json:"projectId"`
	Predictions     []NeedsPrediction      `json:"predictions"`
	Confidence      float64                `json:"confidence"`
	TimeHorizon     string                 `json:"timeHorizon"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

// NeedsPrediction represents a specific prediction
type NeedsPrediction struct {
	Type        string  `json:"type"`
	Area        string  `json:"area"`
	Probability float64 `json:"probability"`
	Timeline    string  `json:"timeline"`
	Impact      string  `json:"impact"`
	Description string  `json:"description"`
}

// OptimizationResult represents the result of memory optimization
type OptimizationResult struct {
	Success          bool          `json:"success"`
	OptimizationTime time.Duration `json:"optimizationTime"`
	SizeBefore       int64         `json:"sizeBefore"`
	SizeAfter        int64         `json:"sizeAfter"`
	SpaceSaved       int64         `json:"spaceSaved"`
	ItemsRemoved     int           `json:"itemsRemoved"`
	ItemsConsolidated int          `json:"itemsConsolidated"`
	PerformanceGain   float64      `json:"performanceGain"`
	Message          string        `json:"message"`
}

// ConsolidationRules represents rules for memory consolidation
type ConsolidationRules struct {
	SimilarityThreshold   float64              `json:"similarityThreshold"`
	ConsolidationStrategy ConsolidationStrategy `json:"consolidationStrategy"`
	PreserveMetadata      bool                 `json:"preserveMetadata"`
	MergeConflicts        bool                 `json:"mergeConflicts"`
	QualityPreference     QualityPreference    `json:"qualityPreference"`
}

// ConsolidationStrategy represents strategies for memory consolidation
type ConsolidationStrategy string

const (
	ConsolidationStrategyMerge      ConsolidationStrategy = "merge"
	ConsolidationStrategyReplace    ConsolidationStrategy = "replace"
	ConsolidationStrategyAggregate  ConsolidationStrategy = "aggregate"
	ConsolidationStrategyReference  ConsolidationStrategy = "reference"
)

// QualityPreference represents preference for quality during consolidation
type QualityPreference string

const (
	QualityPreferenceHighest QualityPreference = "highest"
	QualityPreferenceNewest  QualityPreference = "newest"
	QualityPreferenceOldest  QualityPreference = "oldest"
	QualityPreferenceMerged  QualityPreference = "merged"
)

// ForgetCriteria represents criteria for forgetting memory
type ForgetCriteria struct {
	Age           *TimeRange         `json:"age,omitempty"`
	Usage         *UsageCriteria     `json:"usage,omitempty"`
	Quality       *QualityCriteria   `json:"quality,omitempty"`
	Relevance     *RelevanceCriteria `json:"relevance,omitempty"`
	Categories    []string           `json:"categories,omitempty"`
	ForceForget   bool               `json:"forceForget"`
}

// UsageCriteria represents usage-based criteria
type UsageCriteria struct {
	MaxAccessCount     int     `json:"maxAccessCount"`
	MinDaysSinceAccess int     `json:"minDaysSinceAccess"`
	MaxUsageFrequency  float64 `json:"maxUsageFrequency"`
}

// QualityCriteria represents quality-based criteria
type QualityCriteria struct {
	MinQualityScore float64 `json:"minQualityScore"`
	MaxErrorRate    float64 `json:"maxErrorRate"`
	MinConfidence   float64 `json:"minConfidence"`
}

// RelevanceCriteria represents relevance-based criteria
type RelevanceCriteria struct {
	MinRelevanceScore float64  `json:"minRelevanceScore"`
	ContextRelevance  bool     `json:"contextRelevance"`
	ProjectScope      []string `json:"projectScope,omitempty"`
}

// MemoryAnalyzerConfig represents configuration for memory analyzer
type MemoryAnalyzerConfig struct {
	Enabled          bool   `json:"enabled"`
	AnalysisInterval string `json:"analysisInterval"`
	PatternDetection bool   `json:"patternDetection"`
	TrendAnalysis    bool   `json:"trendAnalysis"`
	QualityAnalysis  bool   `json:"qualityAnalysis"`
}

// MemoryBackupConfig represents configuration for memory backup
type MemoryBackupConfig struct {
	Enabled         bool   `json:"enabled"`
	Interval        string `json:"interval"`
	RetentionPeriod string `json:"retentionPeriod"`
	BackupPath      string `json:"backupPath"`
	Compression     bool   `json:"compression"`
}
