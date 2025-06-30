package types

import (
	"time"
)

// Knowledge represents a piece of stored knowledge
type Knowledge struct {
	ID           string                 `json:"id"`
	Type         KnowledgeType          `json:"type"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	Summary      string                 `json:"summary"`
	Keywords     []string               `json:"keywords"`
	Tags         []string               `json:"tags"`
	Category     string                 `json:"category"`
	Source       string                 `json:"source"`
	Confidence   float64                `json:"confidence"` // 0.0-1.0
	Relevance    float64                `json:"relevance"`  // 0.0-1.0
	Quality      float64                `json:"quality"`    // 0.0-1.0
	ProjectID    string                 `json:"projectId"`
	Verified     bool                   `json:"verified"`
	Usage        *KnowledgeUsage        `json:"usage"`
	Relations    []KnowledgeRelation    `json:"relations"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
	LastUpdated  time.Time              `json:"lastUpdated"`
	AccessedAt   time.Time              `json:"accessedAt"`
	LastAccessed time.Time              `json:"lastAccessed"`
	AccessCount  int                    `json:"accessCount"`
	ExpiresAt    *time.Time             `json:"expiresAt,omitempty"`
}

// KnowledgeType represents the type of knowledge
type KnowledgeType string

const (
	KnowledgeTypeCode         KnowledgeType = "code"
	KnowledgeTypePattern      KnowledgeType = "pattern"
	KnowledgeTypeDecision     KnowledgeType = "decision"
	KnowledgeTypeExperience   KnowledgeType = "experience"
	KnowledgeTypeError        KnowledgeType = "error"
	KnowledgeTypeSolution     KnowledgeType = "solution"
	KnowledgeTypeBestPractice KnowledgeType = "best_practice"
	KnowledgeTypeLesson       KnowledgeType = "lesson"
	KnowledgeTypeInsight      KnowledgeType = "insight"
	KnowledgeTypeContext      KnowledgeType = "context"
	KnowledgeTypeArchitecture KnowledgeType = "architecture"
)

// KnowledgeUsage represents usage statistics for knowledge
type KnowledgeUsage struct {
	AccessCount   int           `json:"accessCount"`
	LastAccessed  time.Time     `json:"lastAccessed"`
	AverageRating float64       `json:"averageRating"`
	UsagePattern  string        `json:"usagePattern"`
	Effectiveness float64       `json:"effectiveness"` // How effective this knowledge is
	Staleness     time.Duration `json:"staleness"`     // How old/stale this knowledge is
}

// KnowledgeRelation represents a relationship between knowledge items
type KnowledgeRelation struct {
	Type        RelationType `json:"type"`
	TargetID    string       `json:"targetId"`
	Strength    float64      `json:"strength"` // 0.0-1.0, strength of relationship
	Description string       `json:"description"`
	CreatedAt   time.Time    `json:"createdAt"`
}

// RelationType represents the type of relationship between knowledge items
type RelationType string

const (
	RelationTypeRelated     RelationType = "related"
	RelationTypeDependsOn   RelationType = "depends_on"
	RelationTypeSupersedes  RelationType = "supersedes"
	RelationTypeContradicts RelationType = "contradicts"
	RelationTypeBuildsOn    RelationType = "builds_on"
	RelationTypeImplements  RelationType = "implements"
	RelationTypeExemplifies RelationType = "exemplifies"
	RelationTypeSimilar     RelationType = "similar"
)

// KnowledgeGraph represents a graph of knowledge relationships
type KnowledgeGraph struct {
	Nodes    []KnowledgeNode        `json:"nodes"`
	Edges    []KnowledgeEdge        `json:"edges"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// KnowledgeNode represents a node in the knowledge graph
type KnowledgeNode struct {
	ID         string                 `json:"id"`
	Type       KnowledgeType          `json:"type"`
	Title      string                 `json:"title"`
	Weight     float64                `json:"weight"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// KnowledgeEdge represents an edge in the knowledge graph
type KnowledgeEdge struct {
	ID         string                 `json:"id"`
	Source     string                 `json:"source"`
	Target     string                 `json:"target"`
	Type       RelationType           `json:"type"`
	Weight     float64                `json:"weight"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// KnowledgeFilters represents filters for knowledge searches
type KnowledgeFilters struct {
	Types         []KnowledgeType `json:"types,omitempty"`
	Categories    []string        `json:"categories,omitempty"`
	Tags          []string        `json:"tags,omitempty"`
	MinConfidence float64         `json:"minConfidence,omitempty"`
	MinRelevance  float64         `json:"minRelevance,omitempty"`
	TimeRange     *TimeRange      `json:"timeRange,omitempty"`
	Sources       []string        `json:"sources,omitempty"`
	Authors       []string        `json:"authors,omitempty"`
}

// KnowledgeValidation represents validation of knowledge
type KnowledgeValidation struct {
	Valid        bool     `json:"valid"`
	Errors       []string `json:"errors,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	Suggestions  []string `json:"suggestions,omitempty"`
	Completeness float64  `json:"completeness"`
	Accuracy     float64  `json:"accuracy"`
	Consistency  float64  `json:"consistency"`
	QualityScore float64  `json:"qualityScore"`
}

// KnowledgeScore represents scoring of knowledge quality
type KnowledgeScore struct {
	Overall      float64 `json:"overall"`
	Accuracy     float64 `json:"accuracy"`
	Completeness float64 `json:"completeness"`
	Relevance    float64 `json:"relevance"`
	Freshness    float64 `json:"freshness"`
	Reliability  float64 `json:"reliability"`
	Usefulness   float64 `json:"usefulness"`
}

// KnowledgeBaseConfig represents configuration for knowledge base
type KnowledgeBaseConfig struct {
	MaxKnowledgeItems    int     `json:"maxKnowledgeItems"`
	AutoCategorization   bool    `json:"autoCategorization"`
	AutoTagging          bool    `json:"autoTagging"`
	QualityThreshold     float64 `json:"qualityThreshold"`
	ValidationEnabled    bool    `json:"validationEnabled"`
	GraphEnabled         bool    `json:"graphEnabled"`
	SimilarityThreshold  float64 `json:"similarityThreshold"`
	ConsolidationEnabled bool    `json:"consolidationEnabled"`
}

// KnowledgeGapAnalysis represents analysis of knowledge gaps
type KnowledgeGapAnalysis struct {
	Domain          string                 `json:"domain"`
	IdentifiedGaps  []KnowledgeGap         `json:"identifiedGaps"`
	CoverageScore   float64                `json:"coverageScore"`
	Priority        []GapPriority          `json:"priority"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

// KnowledgeGap represents a gap in knowledge
type KnowledgeGap struct {
	Area        string   `json:"area"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"` // low, medium, high, critical
	Impact      string   `json:"impact"`
	Topics      []string `json:"topics"`
	Sources     []string `json:"sources,omitempty"`
}

// GapPriority represents priority for addressing gaps
type GapPriority struct {
	Gap      string  `json:"gap"`
	Priority float64 `json:"priority"`
	Urgency  string  `json:"urgency"`
	Effort   string  `json:"effort"`
}

// KnowledgeMetrics represents metrics for knowledge analysis
type KnowledgeMetrics struct {
	TotalKnowledge    int            `json:"totalKnowledge"`
	KnowledgeByType   map[string]int `json:"knowledgeByType"`
	AverageQuality    float64        `json:"averageQuality"`
	AverageRelevance  float64        `json:"averageRelevance"`
	AverageConfidence float64        `json:"averageConfidence"`
	LastUpdated       time.Time      `json:"lastUpdated"`
}
