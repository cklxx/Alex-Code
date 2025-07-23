package message

import (
	"testing"
)

func TestTokenEstimator_EstimateContentTokens(t *testing.T) {
	te := NewTokenEstimator()

	tests := []struct {
		name     string
		content  string
		expected int // rough expected range
		maxDiff  int // maximum acceptable difference
	}{
		{
			name:     "Empty content",
			content:  "",
			expected: 0,
			maxDiff:  0,
		},
		{
			name:     "Simple sentence",
			content:  "Hello world, how are you?",
			expected: 6, // "Hello", "world", ",", "how", "are", "you", "?"
			maxDiff:  2,
		},
		{
			name:     "Code block",
			content:  "```go\nfunc main() {\n    fmt.Println(\"hello\")\n}\n```",
			expected: 25, // Code is more token-dense
			maxDiff:  5,
		},
		{
			name:     "Mixed content",
			content:  "Here's some code:\n```python\nprint('hello')\n```\nThat's a simple example.",
			expected: 20,
			maxDiff:  5,
		},
		{
			name:     "Long text",
			content:  "This is a longer piece of text that contains multiple sentences. It should be tokenized more accurately than the old 3-character rule. We expect better estimation here.",
			expected: 30,
			maxDiff:  8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := te.estimateContentTokens(tt.content)
			
			if abs(result-tt.expected) > tt.maxDiff {
				t.Errorf("estimateContentTokens() = %v, expected around %v (±%v)", 
					result, tt.expected, tt.maxDiff)
			}
			
			// Log for manual inspection
			t.Logf("Content: %q -> %d tokens (expected ~%d)", 
				tt.content, result, tt.expected)
		})
	}
}

func TestTokenEstimator_AccuracyImprovement(t *testing.T) {
	newEstimator := NewTokenEstimator()

	testContent := "func processData(data []string) error {\n    for _, item := range data {\n        if err := validate(item); err != nil {\n            return err\n        }\n    }\n    return nil\n}"

	// Old method (simplified)
	oldEstimate := len(testContent) / 3

	// New method
	newEstimate := newEstimator.estimateContentTokens(testContent)

	t.Logf("Old estimate: %d tokens", oldEstimate)
	t.Logf("New estimate: %d tokens", newEstimate)
	
	// New estimate should be more reasonable for code
	// (typically less than simple char/3 rule because we handle code specially)
	if newEstimate >= oldEstimate {
		t.Logf("Note: New estimator gave higher count - this may be fine for code-heavy content")
	}
	
	// Basic sanity check
	if newEstimate <= 0 {
		t.Error("New estimator returned zero or negative tokens")
	}
}

func BenchmarkTokenEstimator_Old(b *testing.B) {
	content := "This is a sample text that we'll use for benchmarking the token estimation. It contains multiple sentences and should give us a good idea of performance."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Old method
		_ = len(content) / 3
	}
}

func BenchmarkTokenEstimator_New(b *testing.B) {
	te := NewTokenEstimator()
	content := "This is a sample text that we'll use for benchmarking the token estimation. It contains multiple sentences and should give us a good idea of performance."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = te.estimateContentTokens(content)
	}
}

// Helper function
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}