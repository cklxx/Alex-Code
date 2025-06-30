package types

import (
	"time"
)

// CodePattern represents a recognized code pattern
type CodePattern struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Type           CodePatternType        `json:"type"`
	Language       string                 `json:"language"`
	Framework      string                 `json:"framework,omitempty"`
	Template       string                 `json:"template"`
	Category       string                 `json:"category"`
	Tags           []string               `json:"tags"`
	Context        string                 `json:"context"`
	Intent         string                 `json:"intent"`
	Structure      string                 `json:"structure"`
	Participants   []string               `json:"participants"`
	Collaborations []string               `json:"collaborations"`
	Consequences   []string               `json:"consequences"`
	Implementation string                 `json:"implementation"`
	Examples       []PatternExample       `json:"examples"`
	Variations     []PatternVariation     `json:"variations"`
	Usage          *PatternUsage          `json:"usage"`
	Quality        *PatternQuality        `json:"quality"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	LastUpdated    time.Time              `json:"lastUpdated"`
	LastDetected   time.Time              `json:"lastDetected"`
}

// PatternExample represents an example of a code pattern
type PatternExample struct {
	ID          string    `json:"id"`
	PatternID   string    `json:"patternId"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	File        string    `json:"file,omitempty"`
	FilePath    string    `json:"filePath,omitempty"`
	Project     string    `json:"project,omitempty"`
	Quality     float64   `json:"quality"`
	Complexity  int       `json:"complexity"`
	Context     string    `json:"context,omitempty"`
	Explanation string    `json:"explanation,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// PatternVariation represents a variation of a pattern
type PatternVariation struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Template    string                 `json:"template"`
	Differences []string               `json:"differences"`
	Complexity  int                    `json:"complexity"`
	Usage       int                    `json:"usage"`
	Quality     float64                `json:"quality"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PatternUsage represents usage statistics for a pattern
type PatternUsage struct {
	Occurrences     int            `json:"occurrences"`
	Projects        []string       `json:"projects"`
	Languages       map[string]int `json:"languages"`
	Frameworks      map[string]int `json:"frameworks"`
	LastUsed        time.Time      `json:"lastUsed"`
	Popularity      float64        `json:"popularity"`
	SuccessRate     float64        `json:"successRate"`
	PerformanceGain float64        `json:"performanceGain"`
	AdoptionRate    float64        `json:"adoptionRate"`
}

// PatternQuality represents quality metrics for a pattern
type PatternQuality struct {
	Overall         float64   `json:"overall"`
	Readability     float64   `json:"readability"`
	Maintainability float64   `json:"maintainability"`
	Reusability     float64   `json:"reusability"`
	Performance     float64   `json:"performance"`
	Security        float64   `json:"security"`
	Complexity      int       `json:"complexity"`
	Testability     float64   `json:"testability"`
	Validated       bool      `json:"validated"`
	LastUpdated     time.Time `json:"lastUpdated"`
}

// PatternAnalysis represents analysis of a code pattern
type PatternAnalysis struct {
	PatternID       string                 `json:"patternId"`
	Quality         *PatternQuality        `json:"quality"`
	Usage           *PatternUsage          `json:"usage"`
	Complexity      float64                `json:"complexity"`
	Maintainability float64                `json:"maintainability"`
	Performance     float64                `json:"performance"`
	Security        float64                `json:"security"`
	Recommendations []string               `json:"recommendations"`
	Alternatives    []string               `json:"alternatives"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PatternComparison represents comparison between two patterns
type PatternComparison struct {
	Pattern1ID     string                 `json:"pattern1Id"`
	Pattern2ID     string                 `json:"pattern2Id"`
	Similarity     float64                `json:"similarity"`
	Differences    []PatternDifference    `json:"differences"`
	Commonalities  []PatternCommonality   `json:"commonalities"`
	Recommendation string                 `json:"recommendation"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// PatternDifference represents a difference between patterns
type PatternDifference struct {
	Aspect       string  `json:"aspect"`
	Pattern1     string  `json:"pattern1"`
	Pattern2     string  `json:"pattern2"`
	Significance float64 `json:"significance"`
}

// PatternCommonality represents a commonality between patterns
type PatternCommonality struct {
	Aspect      string  `json:"aspect"`
	Description string  `json:"description"`
	Strength    float64 `json:"strength"`
}

// PatternFeedback represents feedback on a pattern
type PatternFeedback struct {
	PatternID   string                 `json:"patternId"`
	UserID      string                 `json:"userId,omitempty"`
	Rating      int                    `json:"rating"` // 1-5
	Comments    string                 `json:"comments,omitempty"`
	Useful      bool                   `json:"useful"`
	Accurate    bool                   `json:"accurate"`
	Complete    bool                   `json:"complete"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// PatternApplication represents the result of applying a pattern
type PatternApplication struct {
	PatternID     string                 `json:"patternId"`
	Success       bool                   `json:"success"`
	GeneratedCode string                 `json:"generatedCode,omitempty"`
	Changes       []CodeChange           `json:"changes"`
	Errors        []string               `json:"errors,omitempty"`
	Warnings      []string               `json:"warnings,omitempty"`
	Suggestions   []string               `json:"suggestions,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// PatternSuggestion represents a suggested pattern
type PatternSuggestion struct {
	PatternID     string                 `json:"patternId"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Confidence    float64                `json:"confidence"`
	Relevance     float64                `json:"relevance"`
	Applicability float64                `json:"applicability"`
	Benefits      []string               `json:"benefits"`
	Requirements  []string               `json:"requirements"`
	Examples      []PatternExample       `json:"examples,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// PatternFilters represents filters for pattern searches
type PatternFilters struct {
	Types      []CodePatternType `json:"types,omitempty"`
	Languages  []string          `json:"languages,omitempty"`
	Frameworks []string          `json:"frameworks,omitempty"`
	Categories []string          `json:"categories,omitempty"`
	MinQuality float64           `json:"minQuality,omitempty"`
	MinUsage   int               `json:"minUsage,omitempty"`
	TimeRange  *TimeRange        `json:"timeRange,omitempty"`
	Complexity []string          `json:"complexity,omitempty"` // low, medium, high
}

// PatternUpdates represents updates to apply to a pattern
type PatternUpdates struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Template    *string                `json:"template,omitempty"`
	Examples    []PatternExample       `json:"examples,omitempty"`
	Variations  []PatternVariation     `json:"variations,omitempty"`
	Quality     *PatternQuality        `json:"quality,omitempty"`
	Usage       *PatternUsage          `json:"usage,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PatternLearnerConfig represents configuration for pattern learner
type PatternLearnerConfig struct {
	AutoLearning          bool    `json:"autoLearning"`
	MinExamples           int     `json:"minExamples"`
	MinQualityThreshold   float64 `json:"minQualityThreshold"`
	MaxVariations         int     `json:"maxVariations"`
	LearningRate          float64 `json:"learningRate"`
	FeedbackWeight        float64 `json:"feedbackWeight"`
	ContextualLearning    bool    `json:"contextualLearning"`
	CrossLanguageLearning bool    `json:"crossLanguageLearning"`
}

// PatternEvolution represents evolution of patterns over time
type PatternEvolution struct {
	PatternID  string      `json:"patternId"`
	Changes    []string    `json:"changes"`
	Quality    []float64   `json:"quality"` // quality over time
	Usage      []int       `json:"usage"`   // usage over time
	Timestamps []time.Time `json:"timestamps"`
}

// PatternIdentificationCriteria represents criteria for pattern identification
type PatternIdentificationCriteria struct {
	MinOccurrences    int        `json:"minOccurrences"`
	MinConfidence     float64    `json:"minConfidence"`
	Languages         []string   `json:"languages,omitempty"`
	Frameworks        []string   `json:"frameworks,omitempty"`
	Complexity        []string   `json:"complexity,omitempty"`
	TimeRange         *TimeRange `json:"timeRange,omitempty"`
	IncludeVariations bool       `json:"includeVariations"`
}

// CodeChange represents a change made to code
type CodeChange struct {
	Type        string `json:"type"` // add, modify, remove
	File        string `json:"file"`
	StartLine   int    `json:"startLine"`
	EndLine     int    `json:"endLine"`
	OldContent  string `json:"oldContent,omitempty"`
	NewContent  string `json:"newContent,omitempty"`
	Description string `json:"description"`
}

// CodeContext represents context for code analysis
type CodeContext struct {
	File         string                 `json:"file"`
	Language     string                 `json:"language"`
	Content      string                 `json:"content"`
	Function     string                 `json:"function,omitempty"`
	Class        string                 `json:"class,omitempty"`
	StartLine    int                    `json:"startLine"`
	EndLine      int                    `json:"endLine"`
	Imports      []string               `json:"imports"`
	Dependencies []string               `json:"dependencies"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// CodePatternType represents the type of code pattern
type CodePatternType string

const (
	CodePatternTypeStructural    CodePatternType = "structural"
	CodePatternTypeBehavioral    CodePatternType = "behavioral"
	CodePatternTypeCreational    CodePatternType = "creational"
	CodePatternTypeArchitectural CodePatternType = "architectural"
	CodePatternTypeFunctional    CodePatternType = "functional"
	CodePatternTypeDesign        CodePatternType = "design"
	CodePatternTypeIdiom         CodePatternType = "idiom"
	CodePatternTypeAntiPattern   CodePatternType = "anti_pattern"
)

// ApplicationContext represents context for pattern application
type ApplicationContext struct {
	ProjectID    string                 `json:"projectId"`
	Language     string                 `json:"language"`
	Framework    string                 `json:"framework,omitempty"`
	TargetFile   string                 `json:"targetFile"`
	CodeContext  string                 `json:"codeContext"`
	Requirements []string               `json:"requirements"`
	Constraints  []string               `json:"constraints"`
	UserIntent   string                 `json:"userIntent"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}
