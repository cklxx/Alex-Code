package builtin

import (
	"context"
	"testing"
)

func TestWebSearchTool(t *testing.T) {
	tool := CreateWebSearchTool()

	// Test tool metadata
	if tool.Name() != "web_search" {
		t.Errorf("expected name 'web_search', got %s", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	// Test parameters schema
	params := tool.Parameters()
	if params == nil {
		t.Error("expected parameters schema")
	}

	// Check required parameters
	properties, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Error("expected properties in parameters schema")
	}

	if _, ok := properties["query"]; !ok {
		t.Error("expected 'query' parameter in schema")
	}

	// Test validation
	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
	}{
		{
			name:        "missing query",
			args:        map[string]interface{}{},
			expectError: true,
		},
		{
			name: "empty query",
			args: map[string]interface{}{
				"query": "",
			},
			expectError: true,
		},
		{
			name: "invalid query type",
			args: map[string]interface{}{
				"query": 123,
			},
			expectError: true,
		},
		{
			name: "valid basic query",
			args: map[string]interface{}{
				"query": "golang programming",
			},
			expectError: false,
		},
		{
			name: "valid query with max_results",
			args: map[string]interface{}{
				"query":       "machine learning",
				"max_results": 10.0,
			},
			expectError: false,
		},
		{
			name: "invalid max_results",
			args: map[string]interface{}{
				"query":       "test query",
				"max_results": 25.0, // exceeds maximum of 20
			},
			expectError: true,
		},
		{
			name: "invalid search_depth",
			args: map[string]interface{}{
				"query":        "test query",
				"search_depth": "invalid",
			},
			expectError: true,
		},
		{
			name: "valid search_depth",
			args: map[string]interface{}{
				"query":        "test query",
				"search_depth": "advanced",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tool.Validate(tt.args)
			if tt.expectError && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no validation error, got %v", err)
			}
		})
	}
}

func TestWebSearchToolExecution(t *testing.T) {
	tool := CreateWebSearchTool()

	// Test execution without API key (should return configuration message)
	args := map[string]interface{}{
		"query": "golang programming tutorial",
	}

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// Should return a configuration message when no API key is set
	if result.Content == "" {
		t.Error("expected content in result")
	}

	// Check that it mentions configuration
	if !contains(result.Content, "configuration") && !contains(result.Content, "API key") {
		t.Error("expected configuration-related message")
	}

	// Check result data
	if result.Data == nil {
		t.Error("expected data in result")
	}

	if configured, ok := result.Data["configured"].(bool); !ok || configured {
		t.Error("expected configured to be false when no API key is set")
	}
}

func TestWebSearchToolWithAPIKey(t *testing.T) {
	tool := CreateWebSearchTool()

	// Set a mock API key
	tool.SetAPIKey("test-api-key")

	// Test that the API key is set (we can't test actual API calls without a real key)
	// but we can verify the tool accepts the key
	if tool.apiKey != "test-api-key" {
		t.Error("expected API key to be set")
	}
}

func TestNewsSearchTool(t *testing.T) {
	tool := CreateNewsSearchTool()

	if tool.Name() != "news_search" {
		t.Errorf("expected name 'news_search', got %s", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	// Test that news search has appropriate defaults
	args := map[string]interface{}{
		"query": "latest technology news",
	}

	// Should not error during validation
	err := tool.Validate(args)
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestAcademicSearchTool(t *testing.T) {
	tool := CreateAcademicSearchTool()

	if tool.Name() != "academic_search" {
		t.Errorf("expected name 'academic_search', got %s", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	// Test that academic search has appropriate defaults
	args := map[string]interface{}{
		"query": "machine learning research papers",
	}

	// Should not error during validation
	err := tool.Validate(args)
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestWebSearchToolFormatResults(t *testing.T) {
	tool := CreateWebSearchTool()

	// Test formatting with mock response
	response := &TavilyResponse{
		Query:  "test query",
		Answer: "This is a test answer",
		Results: []TavilyResult{
			{
				Title:     "Test Result 1",
				URL:       "https://example.com/1",
				Content:   "This is test content 1",
				Score:     0.95,
				Published: "2024-01-01",
			},
			{
				Title:   "Test Result 2",
				URL:     "https://example.com/2",
				Content: "This is test content 2",
				Score:   0.85,
			},
		},
		Images: []TavilyImage{
			{
				URL:         "https://example.com/image1.jpg",
				Description: "Test image 1",
			},
		},
		FollowUpQuestions: []string{
			"What about test question 1?",
			"How about test question 2?",
		},
		ResponseTime: 1.23,
	}

	formatted := tool.formatResults(response)

	// Check that formatted output contains expected elements
	expectedElements := []string{
		"# Web Search Results",
		"test query",
		"## Summary",
		"This is a test answer",
		"## Search Results",
		"Test Result 1",
		"https://example.com/1",
		"This is test content 1",
		"Test Result 2",
		"## Related Images",
		"https://example.com/image1.jpg",
		"## Follow-up Questions",
		"What about test question 1?",
		"1.23 seconds",
	}

	for _, element := range expectedElements {
		if !contains(formatted, element) {
			t.Errorf("expected formatted output to contain '%s'", element)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
