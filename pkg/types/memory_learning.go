package types

import (
	"time"
)

// LearningExperience represents an experience to learn from
type LearningExperience struct {
	ID          string                 `json:"id"`
	Type        LearningType           `json:"type"`
	Context     *LearningContext       `json:"context"`
	Outcome     *Outcome               `json:"outcome"`
	Feedback    *LearningFeedback      `json:"feedback,omitempty"`
	Insights    []string               `json:"insights"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// LearningType represents the type of learning experience
type LearningType string

const (
	LearningTypeSuccess      LearningType = "success"
	LearningTypeFailure      LearningType = "failure"
	LearningTypeOptimization LearningType = "optimization"
	LearningTypeDiscovery    LearningType = "discovery"
	LearningTypeCorrection   LearningType = "correction"
	LearningTypeRefinement   LearningType = "refinement"
)

// LearningContext represents context for learning
type LearningContext struct {
	TaskType    string                 `json:"taskType"`
	ProjectID   string                 `json:"projectId"`
	Environment map[string]string      `json:"environment"`
	Tools       []string               `json:"tools"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Outcome represents the outcome of an experience
type Outcome struct {
	Success     bool                   `json:"success"`
	Description string                 `json:"description"`
	Metrics     map[string]float64     `json:"metrics,omitempty"`
	Impact      string                 `json:"impact"`
	Duration    time.Duration          `json:"duration"`
	Cost        float64                `json:"cost,omitempty"`
	Quality     float64                `json:"quality"`
}

// LearningFeedback represents feedback on performance
type LearningFeedback struct {
	Type        FeedbackType           `json:"type"`
	Source      string                 `json:"source"`
	Content     string                 `json:"content"`
	Rating      float64                `json:"rating"`      // 0.0-1.0
	Aspects     []FeedbackAspect       `json:"aspects"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// FeedbackType represents the type of feedback
type FeedbackType string

const (
	FeedbackTypePositive    FeedbackType = "positive"
	FeedbackTypeNegative    FeedbackType = "negative"
	FeedbackTypeConstructive FeedbackType = "constructive"
	FeedbackTypeSuggestion  FeedbackType = "suggestion"
	FeedbackTypeCorrection  FeedbackType = "correction"
	FeedbackTypeValidation  FeedbackType = "validation"
)

// FeedbackAspect represents a specific aspect of feedback
type FeedbackAspect struct {
	Name        string  `json:"name"`
	Rating      float64 `json:"rating"`      // 0.0-1.0
	Description string  `json:"description"`
	Importance  float64 `json:"importance"`  // 0.0-1.0
}

// LearningTrendAnalysis represents analysis of learning trends
type LearningTrendAnalysis struct {
	ProjectID        string                 `json:"projectId"`
	Period           string                 `json:"period"`
	LearningRate     float64                `json:"learningRate"`
	KnowledgeGrowth  float64                `json:"knowledgeGrowth"`
	PatternEvolution []PatternEvolution     `json:"patternEvolution"`
	SkillDevelopment []SkillProgress        `json:"skillDevelopment"`
	Insights         []string               `json:"insights"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
}

// SkillProgress represents progress in a skill area
type SkillProgress struct {
	Skill       string    `json:"skill"`
	Level       float64   `json:"level"`       // 0.0-1.0
	Growth      float64   `json:"growth"`      // rate of improvement
	Milestones  []string  `json:"milestones"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// Lesson represents a lesson learned
type Lesson struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Context     string                 `json:"context"`
	Learning    string                 `json:"learning"`
	Application []string               `json:"application"`
	Evidence    []string               `json:"evidence"`
	Impact      LessonImpact           `json:"impact"`
	Confidence  float64                `json:"confidence"`
	Verified    bool                   `json:"verified"`
	Applied     bool                   `json:"applied"`
	Effectiveness float64              `json:"effectiveness"`
	Tags        []string               `json:"tags"`
	Related     []string               `json:"related"`     // IDs of related lessons
	Source      string                 `json:"source"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	LastApplied *time.Time             `json:"lastApplied,omitempty"`
}

// LessonImpact represents the impact level of a lesson
type LessonImpact string

const (
	LessonImpactLow      LessonImpact = "low"
	LessonImpactMedium   LessonImpact = "medium"
	LessonImpactHigh     LessonImpact = "high"
	LessonImpactCritical LessonImpact = "critical"
)

// LessonContext represents context for applying lessons
type LessonContext struct {
	ProjectID   string                 `json:"projectId"`
	TaskType    string                 `json:"taskType"`
	Context     string                 `json:"context"`
	Category    string                 `json:"category"`
	Severity    string                 `json:"severity,omitempty"`
	Stakeholders []string              `json:"stakeholders,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LessonApplication represents the application of a lesson
type LessonApplication struct {
	LessonID     string                 `json:"lessonId"`
	Applied      bool                   `json:"applied"`
	Effectiveness float64               `json:"effectiveness"`
	Feedback     string                 `json:"feedback,omitempty"`
	Adaptations  []string               `json:"adaptations,omitempty"`
	Results      map[string]interface{} `json:"results,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// LessonLearnerConfig represents configuration for lesson learner
type LessonLearnerConfig struct {
	AutoLearning        bool    `json:"autoLearning"`
	ConfidenceThreshold float64 `json:"confidenceThreshold"`
	MinEvidence         int     `json:"minEvidence"`
	VerificationRequired bool   `json:"verificationRequired"`
	LearningRate        float64 `json:"learningRate"`
	AdaptationEnabled   bool    `json:"adaptationEnabled"`
	ContextAwareness    bool    `json:"contextAwareness"`
}

// ExperienceFilters represents filters for experience searches
type ExperienceFilters struct {
	Types         []LearningType  `json:"types,omitempty"`
	Outcomes      []bool          `json:"outcomes,omitempty"` // success/failure
	Categories    []string        `json:"categories,omitempty"`
	Projects      []string        `json:"projects,omitempty"`
	TimeRange     *TimeRange      `json:"timeRange,omitempty"`
	MinConfidence float64         `json:"minConfidence,omitempty"`
	HasFeedback   *bool           `json:"hasFeedback,omitempty"`
	Verified      *bool           `json:"verified,omitempty"`
}

// LessonFilters represents filters for lesson searches
type LessonFilters struct {
	Categories     []string       `json:"categories,omitempty"`
	ImpactLevels   []LessonImpact `json:"impactLevels,omitempty"`
	Verified       *bool          `json:"verified,omitempty"`
	Applied        *bool          `json:"applied,omitempty"`
	MinConfidence  float64        `json:"minConfidence,omitempty"`
	MinEffectiveness float64      `json:"minEffectiveness,omitempty"`
	Tags           []string       `json:"tags,omitempty"`
	Sources        []string       `json:"sources,omitempty"`
	TimeRange      *TimeRange     `json:"timeRange,omitempty"`
}
