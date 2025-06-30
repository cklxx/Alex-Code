package testutils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"deep-coding-agent/pkg/types"
)

// TestContext provides test environment context
type TestContext struct {
	TempDir    string
	ConfigPath string
	Timeout    time.Duration
	Cleanup    func()
}

// NewTestContext creates a new test context
func NewTestContext(t *testing.T) *TestContext {
	tempDir, err := os.MkdirTemp("", "deep-coding-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	configPath := filepath.Join(tempDir, "test-config.json")

	ctx := &TestContext{
		TempDir:    tempDir,
		ConfigPath: configPath,
		Timeout:    30 * time.Second,
		Cleanup: func() {
			os.RemoveAll(tempDir)
		},
	}

	t.Cleanup(ctx.Cleanup)
	return ctx
}

// CreateTestFile creates a test file with given content
func (tc *TestContext) CreateTestFile(t *testing.T, name, content string) string {
	filePath := filepath.Join(tc.TempDir, name)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return filePath
}

// CreateTestDir creates a test directory
func (tc *TestContext) CreateTestDir(t *testing.T, name string) string {
	dirPath := filepath.Join(tc.TempDir, name)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	return dirPath
}

// SetupTestMemoryConfig creates a test memory manager config
func SetupTestMemoryConfig() *types.MemoryManagerConfig {
	return &types.MemoryManagerConfig{
		MetricsEnabled:     true,
		CacheEnabled:       true,
		BackupEnabled:      false,
		CompressionEnabled: false,
	}
}

// GenerateTestKnowledge creates test knowledge data
func GenerateTestKnowledge(count int, prefix string) []*types.Knowledge {
	var knowledge []*types.Knowledge

	for i := 0; i < count; i++ {
		k := &types.Knowledge{
			ID:         fmt.Sprintf("%s_knowledge_%d", prefix, i),
			Type:       types.KnowledgeTypeExperience,
			Title:      fmt.Sprintf("Test Knowledge %d", i),
			Content:    fmt.Sprintf("This is test knowledge content %d", i),
			Summary:    fmt.Sprintf("Summary for knowledge %d", i),
			Keywords:   []string{fmt.Sprintf("keyword%d", i), "test", "knowledge"},
			Tags:       []string{fmt.Sprintf("tag%d", i), "test"},
			Category:   "test_category",
			Source:     "test_source",
			Confidence: 0.8 + float64(i%3)*0.1,
			Relevance:  0.7 + float64(i%4)*0.05,
			Quality:    0.9 - float64(i%5)*0.02,
			ProjectID:  "test_project",
			Verified:   i%2 == 0,
			Metadata: map[string]interface{}{
				"test_id": i,
				"batch":   prefix,
			},
			CreatedAt:    time.Now().Add(-time.Duration(i) * time.Hour),
			UpdatedAt:    time.Now().Add(-time.Duration(i) * time.Minute),
			LastUpdated:  time.Now().Add(-time.Duration(i) * time.Minute),
			AccessedAt:   time.Now(),
			LastAccessed: time.Now(),
		}
		knowledge = append(knowledge, k)
	}

	return knowledge
}

// GenerateTestPatterns creates test pattern data
func GenerateTestPatterns(count int, prefix string) []*types.CodePattern {
	var patterns []*types.CodePattern

	patternTypes := []types.CodePatternType{
		types.CodePatternTypeStructural,
		types.CodePatternTypeBehavioral,
		types.CodePatternTypeCreational,
	}

	for i := 0; i < count; i++ {
		p := &types.CodePattern{
			ID:          fmt.Sprintf("%s_pattern_%d", prefix, i),
			Name:        fmt.Sprintf("Test Pattern %d", i),
			Description: fmt.Sprintf("This is test pattern %d", i),
			Type:        patternTypes[i%len(patternTypes)],
			Language:    "go",
			Template:    fmt.Sprintf("func TestPattern%d() {\n\t// pattern code\n}", i),
			Category:    "test_patterns",
			Tags:        []string{fmt.Sprintf("pattern%d", i), "test"},
			Context:     fmt.Sprintf("Test context for pattern %d", i),
			Intent:      fmt.Sprintf("Intent for pattern %d", i),
			Structure:   fmt.Sprintf("Structure: pattern %d", i),
			Usage: &types.PatternUsage{
				Occurrences:  i + 1,
				Projects:     []string{"test_project"},
				Languages:    map[string]int{"go": i + 1},
				LastUsed:     time.Now(),
				Popularity:   0.5 + float64(i%5)*0.1,
				SuccessRate:  0.8 + float64(i%3)*0.05,
				AdoptionRate: 0.6 + float64(i%4)*0.1,
			},
			Quality: &types.PatternQuality{
				Overall:         0.8 + float64(i%5)*0.04,
				Readability:     0.9 - float64(i%3)*0.05,
				Maintainability: 0.8 + float64(i%4)*0.03,
				Reusability:     0.7 + float64(i%6)*0.05,
				Performance:     0.8,
				Security:        0.9,
				Complexity:      i%5 + 1,
				Testability:     0.8,
				Validated:       i%2 == 0,
				LastUpdated:     time.Now(),
			},
			CreatedAt:   time.Now().Add(-time.Duration(i) * time.Hour),
			UpdatedAt:   time.Now().Add(-time.Duration(i) * time.Minute),
			LastUpdated: time.Now().Add(-time.Duration(i) * time.Minute),
		}
		patterns = append(patterns, p)
	}

	return patterns
}

// AssertKnowledgeEqual compares two knowledge items for testing
func AssertKnowledgeEqual(t *testing.T, expected, actual *types.Knowledge) {
	if expected.ID != actual.ID {
		t.Errorf("ID mismatch: expected %s, got %s", expected.ID, actual.ID)
	}
	if expected.Type != actual.Type {
		t.Errorf("Type mismatch: expected %v, got %v", expected.Type, actual.Type)
	}
	if expected.Title != actual.Title {
		t.Errorf("Title mismatch: expected %s, got %s", expected.Title, actual.Title)
	}
	if expected.Content != actual.Content {
		t.Errorf("Content mismatch: expected %s, got %s", expected.Content, actual.Content)
	}
	if expected.Confidence != actual.Confidence {
		t.Errorf("Confidence mismatch: expected %f, got %f", expected.Confidence, actual.Confidence)
	}
	if expected.ProjectID != actual.ProjectID {
		t.Errorf("ProjectID mismatch: expected %s, got %s", expected.ProjectID, actual.ProjectID)
	}
}

// AssertPatternEqual compares two pattern items for testing
func AssertPatternEqual(t *testing.T, expected, actual *types.CodePattern) {
	if expected.ID != actual.ID {
		t.Errorf("ID mismatch: expected %s, got %s", expected.ID, actual.ID)
	}
	if expected.Name != actual.Name {
		t.Errorf("Name mismatch: expected %s, got %s", expected.Name, actual.Name)
	}
	if expected.Type != actual.Type {
		t.Errorf("Type mismatch: expected %v, got %v", expected.Type, actual.Type)
	}
	if expected.Language != actual.Language {
		t.Errorf("Language mismatch: expected %s, got %s", expected.Language, actual.Language)
	}
	if expected.Template != actual.Template {
		t.Errorf("Template mismatch: expected %s, got %s", expected.Template, actual.Template)
	}
}

// CreateContextWithTimeout creates a context with timeout for tests
func CreateContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// CreateTestProjectMemory creates test project memory data
func CreateTestProjectMemory(projectID string) *types.ProjectMemory {
	return &types.ProjectMemory{
		ID:          fmt.Sprintf("memory_%s", projectID),
		ProjectID:   projectID,
		ProjectPath: fmt.Sprintf("/test/projects/%s", projectID),
		Architecture: &types.ProjectArchitecture{
			ID:          fmt.Sprintf("arch_%s", projectID),
			ProjectID:   projectID,
			Name:        fmt.Sprintf("%s Architecture", projectID),
			Description: "Test project architecture",
			Version:     "1.0.0",
			Components: []types.ArchitecturalComponent{
				{
					ID:               "comp1",
					Name:             "Main Component",
					Type:             "service",
					Description:      "Main service component",
					Responsibilities: []string{"main logic", "api handling"},
					Location:         "/main",
					Status:           "active",
				},
			},
			Technologies: []string{"Go", "Docker"},
			Principles:   []string{"Clean Architecture", "SOLID"},
			CreatedAt:    time.Now(),
			LastUpdated:  time.Now(),
		},
		Knowledge: func() []types.Knowledge {
			ptrs := GenerateTestKnowledge(5, projectID)
			result := make([]types.Knowledge, len(ptrs))
			for i, ptr := range ptrs {
				result[i] = *ptr
			}
			return result
		}(),
		Patterns: func() []types.CodePattern {
			ptrs := GenerateTestPatterns(3, projectID)
			result := make([]types.CodePattern, len(ptrs))
			for i, ptr := range ptrs {
				result[i] = *ptr
			}
			return result
		}(),
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
	}
}
