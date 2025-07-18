package agent

import (
	"alex/internal/session"
	"testing"
	"time"
)

func TestMessageCompression_Strategies(t *testing.T) {
	// Test token estimation improvements
	te := NewTokenEstimator()

	// Create test messages with different characteristics
	messages := createTestMessages(100)

	// Test different message counts
	tests := []struct {
		name         string
		messageCount int
	}{
		{"Few messages", 10},
		{"Medium messages", 30},
		{"Many messages", 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testMessages := messages[:tt.messageCount]

			// Measure token estimation
			tokens := te.EstimateSessionMessages(testMessages)

			t.Logf("Messages: %d, Estimated tokens: %d", len(testMessages), tokens)

			// Basic sanity checks
			if tokens <= 0 {
				t.Error("Token estimation should be positive")
			}

			if tokens > len(testMessages)*1000 {
				t.Error("Token estimation seems too high")
			}
		})
	}
}

func TestMessageImportanceScoring(t *testing.T) {
	// Test the importance scoring logic independently
	te := NewTokenEstimator()

	// Create a mock message processor with just the scoring method
	mp := &MessageProcessor{
		tokenEstimator: te,
	}

	tests := []struct {
		name     string
		message  *session.Message
		minScore float64
		maxScore float64
	}{
		{
			name: "High importance - code with error",
			message: &session.Message{
				Content:   "```go\nfunc main() {\n    err := doSomething()\n    if err != nil {\n        return err\n    }\n}\n```\nThis code has an error in the logic.",
				ToolCalls: []session.ToolCall{{Name: "file_read", ID: "1"}},
			},
			minScore: 20.0,
			maxScore: 50.0,
		},
		{
			name: "Low importance - simple acknowledgment",
			message: &session.Message{
				Content: "好的",
			},
			minScore: 0.0,
			maxScore: 5.0,
		},
		{
			name: "Medium importance - question",
			message: &session.Message{
				Content: "How should I implement this feature?",
			},
			minScore: 5.0,
			maxScore: 15.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := mp.calculateMessageImportance(tt.message)

			t.Logf("Message: %q -> Score: %.2f", tt.message.Content, score)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("Expected score between %.2f and %.2f, got %.2f",
					tt.minScore, tt.maxScore, score)
			}
		})
	}
}

// No need for mock LLM client in simplified tests

// Helper to create test messages
func createTestMessages(count int) []*session.Message {
	messages := make([]*session.Message, count)

	for i := 0; i < count; i++ {
		var content string
		var toolCalls []session.ToolCall

		// Create varied message types
		switch i % 5 {
		case 0:
			content = "好的"
		case 1:
			content = "Can you help me implement this function?"
		case 2:
			content = "```go\nfunc example() {\n    // some code\n}\n```"
			toolCalls = []session.ToolCall{{Name: "file_read", ID: "1"}}
		case 3:
			content = "There's an error in the code: undefined variable"
		case 4:
			content = "This is a longer message that contains more detailed information about the implementation approach and considerations."
		}

		messages[i] = &session.Message{
			Role:      "user",
			Content:   content,
			ToolCalls: toolCalls,
			Timestamp: time.Now(),
		}
	}

	return messages
}
