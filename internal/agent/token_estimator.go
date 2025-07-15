package agent

import (
	"alex/internal/llm"
	"alex/internal/session"
)

// TokenEstimator provides unified token estimation logic
type TokenEstimator struct {
	// Configuration for token estimation
	charsPerToken int
	overhead      int
}

// NewTokenEstimator creates a new token estimator
func NewTokenEstimator() *TokenEstimator {
	return &TokenEstimator{
		charsPerToken: 3,   // Rough estimate: 3 characters = 1 token
		overhead:      100, // Overhead for role, metadata, etc.
	}
}

// EstimateSessionMessages estimates tokens for session messages
func (te *TokenEstimator) EstimateSessionMessages(messages []*session.Message) int {
	totalChars := 0
	
	for _, msg := range messages {
		totalChars += len(msg.Content) + te.overhead
		
		// Add tokens for tool calls
		for _, tc := range msg.ToolCalls {
			totalChars += len(tc.Name) + len(tc.ID) + 50 // Tool call overhead
		}
	}
	
	return totalChars / te.charsPerToken
}

// EstimateLLMMessages estimates tokens for LLM messages
func (te *TokenEstimator) EstimateLLMMessages(messages []llm.Message) int {
	totalChars := 0
	
	for _, msg := range messages {
		totalChars += len(msg.Content) + te.overhead
		
		// Add tokens for tool calls
		for _, tc := range msg.ToolCalls {
			totalChars += len(tc.Function.Name) + len(tc.ID) + 50 // Tool call overhead
		}
	}
	
	return totalChars / te.charsPerToken
}

// EstimateString estimates tokens for a string
func (te *TokenEstimator) EstimateString(content string) int {
	return (len(content) + te.overhead) / te.charsPerToken
}

// EstimateMessages estimates tokens for mixed message types
func (te *TokenEstimator) EstimateMessages(sessionMessages []*session.Message, llmMessages []llm.Message) int {
	totalTokens := 0
	
	if len(sessionMessages) > 0 {
		totalTokens += te.EstimateSessionMessages(sessionMessages)
	}
	
	if len(llmMessages) > 0 {
		totalTokens += te.EstimateLLMMessages(llmMessages)
	}
	
	return totalTokens
}

// CheckTokenLimit checks if messages exceed token limit
func (te *TokenEstimator) CheckTokenLimit(messages []*session.Message, maxTokens int) bool {
	estimated := te.EstimateSessionMessages(messages)
	return estimated > maxTokens
}

// GetCompressionThreshold calculates compression threshold based on max tokens
func (te *TokenEstimator) GetCompressionThreshold(maxTokens int, threshold float64) int {
	return int(float64(maxTokens) * threshold)
}