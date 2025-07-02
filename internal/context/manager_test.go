package context

import (
	"context"
	"testing"
	"time"

	"alex/internal/llm"
	"alex/internal/session"
)

// MockLLMClient implements llm.Client for testing
type MockLLMClient struct {
	responses map[string]*llm.ChatResponse
}

func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{
		responses: make(map[string]*llm.ChatResponse),
	}
}

func (m *MockLLMClient) AddResponse(prompt string, response *llm.ChatResponse) {
	m.responses[prompt] = response
}

func (m *MockLLMClient) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	// Simple mock - return a structured summary
	mockSummary := `{
		"summary": "Test conversation involving code analysis and file operations",
		"key_points": ["Analyzed file structure", "Discussed implementation approach", "Identified potential issues"],
		"topics": ["code analysis", "file operations", "testing"],
		"action_items": ["Review implementation", "Add tests"],
		"decisions": ["Use mock client for testing"],
		"code_changes": [{"file": "test.go", "description": "Added test cases", "type": "modified"}],
		"context": {"current_task": "testing context management"}
	}`

	return &llm.ChatResponse{
		ID:      "test-response",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "test-model",
		Choices: []llm.Choice{
			{
				Index: 0,
				Message: llm.Message{
					Role:    "assistant",
					Content: mockSummary,
				},
				FinishReason: "stop",
			},
		},
		Usage: llm.Usage{
			PromptTokens:     100,
			CompletionTokens: 200,
			TotalTokens:      300,
		},
	}, nil
}

func (m *MockLLMClient) ChatStream(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamDelta, error) {
	// Not implemented for this test
	return nil, nil
}

func (m *MockLLMClient) Close() error {
	return nil
}

func createTestSession() *session.Session {
	sessionMgr, _ := session.NewManager()
	sess, _ := sessionMgr.StartSession("test-session")
	
	// Add some test messages
	messages := []*session.Message{
		{Role: "system", Content: "You are a helpful assistant", Timestamp: time.Now()},
		{Role: "user", Content: "Help me analyze this code", Timestamp: time.Now()},
		{Role: "assistant", Content: "I'll help you analyze the code. Please share the code you'd like me to review.", Timestamp: time.Now()},
		{Role: "user", Content: "Here's the code: func main() { fmt.Println(\"Hello\") }", Timestamp: time.Now()},
		{Role: "assistant", Content: "This is a simple Go program that prints 'Hello' to the console. The code looks correct.", Timestamp: time.Now()},
	}
	
	for _, msg := range messages {
		sess.AddMessage(msg)
	}
	
	return sess
}

func TestContextManager_CheckContextLength(t *testing.T) {
	mockClient := NewMockLLMClient()
	config := &ContextLengthConfig{
		MaxTokens:              1000,
		SummarizationThreshold: 750,
		CompressionRatio:       0.3,
		PreserveSystemMessages: true,
	}
	
	cm := NewContextManager(mockClient, config)
	sess := createTestSession()
	
	analysis, err := cm.CheckContextLength(sess)
	if err != nil {
		t.Fatalf("CheckContextLength failed: %v", err)
	}
	
	if analysis.TotalMessages != 5 {
		t.Errorf("Expected 5 messages, got %d", analysis.TotalMessages)
	}
	
	if analysis.EstimatedTokens <= 0 {
		t.Errorf("Expected positive token count, got %d", analysis.EstimatedTokens)
	}
}

func TestContextManager_ProcessContextOverflow(t *testing.T) {
	mockClient := NewMockLLMClient()
	config := &ContextLengthConfig{
		MaxTokens:              50, // Very low to trigger overflow
		SummarizationThreshold: 30,
		CompressionRatio:       0.3,
		PreserveSystemMessages: true,
	}
	
	cm := NewContextManager(mockClient, config)
	sess := createTestSession()
	
	ctx := context.Background()
	result, err := cm.ProcessContextOverflow(ctx, sess)
	if err != nil {
		t.Fatalf("ProcessContextOverflow failed: %v", err)
	}
	
	if result.Action != "summarized" {
		t.Errorf("Expected action 'summarized', got %s", result.Action)
	}
	
	// The processed count might be higher due to system message and summary being added
	if result.OriginalCount == 0 {
		t.Errorf("Expected original count to be greater than 0, got %d", result.OriginalCount)
	}
	
	if result.BackupID == "" {
		t.Errorf("Expected backup ID to be generated")
	}
	
	// Check that the session now has a summary message
	messages := sess.GetMessages()
	hasSummary := false
	for _, msg := range messages {
		if msg.Role == "system" {
			if metadata, ok := msg.Metadata["type"]; ok && metadata == "context_summary" {
				hasSummary = true
				break
			}
		}
	}
	
	if !hasSummary {
		t.Errorf("Expected session to contain a summary message after processing")
	}
}

func TestContextManager_GetContextStats(t *testing.T) {
	mockClient := NewMockLLMClient()
	cm := NewContextManager(mockClient, nil)
	sess := createTestSession()
	
	stats := cm.GetContextStats(sess)
	
	if stats.TotalMessages != 5 {
		t.Errorf("Expected 5 total messages, got %d", stats.TotalMessages)
	}
	
	if stats.SystemMessages != 1 {
		t.Errorf("Expected 1 system message, got %d", stats.SystemMessages)
	}
	
	if stats.UserMessages != 2 {
		t.Errorf("Expected 2 user messages, got %d", stats.UserMessages)
	}
	
	if stats.AssistantMessages != 2 {
		t.Errorf("Expected 2 assistant messages, got %d", stats.AssistantMessages)
	}
}

func TestMessageSummarizer_SummarizeMessages(t *testing.T) {
	mockClient := NewMockLLMClient()
	config := &ContextLengthConfig{}
	summarizer := NewMessageSummarizer(mockClient, config)
	
	sess := createTestSession()
	messages := sess.GetMessages()
	
	ctx := context.Background()
	summary, err := summarizer.SummarizeMessages(ctx, messages)
	if err != nil {
		t.Fatalf("SummarizeMessages failed: %v", err)
	}
	
	if summary.Summary == "" {
		t.Errorf("Expected non-empty summary")
	}
	
	if len(summary.KeyPoints) == 0 {
		t.Errorf("Expected key points to be extracted")
	}
	
	if len(summary.Topics) == 0 {
		t.Errorf("Expected topics to be identified")
	}
	
	if summary.TokensUsed <= 0 {
		t.Errorf("Expected positive token usage, got %d", summary.TokensUsed)
	}
}

func TestContextPreservationManager_CreateBackup(t *testing.T) {
	cpm := NewContextPreservationManager()
	sess := createTestSession()
	
	backup := cpm.CreateBackup(sess)
	
	if backup.ID == "" {
		t.Errorf("Expected backup ID to be generated")
	}
	
	if backup.SessionID != sess.ID {
		t.Errorf("Expected backup session ID to match session ID")
	}
	
	if len(backup.Messages) != sess.GetMessageCount() {
		t.Errorf("Expected backup to contain all messages")
	}
	
	if backup.OriginalCount != sess.GetMessageCount() {
		t.Errorf("Expected original count to match message count")
	}
}

func TestContextPreservationManager_RestoreBackup(t *testing.T) {
	cpm := NewContextPreservationManager()
	sess := createTestSession()
	originalCount := sess.GetMessageCount()
	
	// Create backup
	backup := cpm.CreateBackup(sess)
	
	// Modify session (clear messages)
	sess.ClearMessages()
	if sess.GetMessageCount() != 0 {
		t.Errorf("Expected session to be cleared")
	}
	
	// Restore from backup
	err := cpm.RestoreBackup(sess, backup.ID)
	if err != nil {
		t.Fatalf("RestoreBackup failed: %v", err)
	}
	
	if sess.GetMessageCount() != originalCount {
		t.Errorf("Expected restored session to have %d messages, got %d", 
			originalCount, sess.GetMessageCount())
	}
}

func TestReactAgentContextIntegration_Basic(t *testing.T) {
	mockClient := NewMockLLMClient()
	config := &IntegrationConfig{
		Enabled:      true,
		AutoTrimming: true,
		ContextLengthConfig: &ContextLengthConfig{
			MaxTokens:              1000,
			SummarizationThreshold: 750,
		},
	}
	
	integration := NewReactAgentContextIntegration(mockClient, config)
	sess := createTestSession()
	
	// Test context status check
	analysis, err := integration.CheckContextStatus(sess)
	if err != nil {
		t.Fatalf("CheckContextStatus failed: %v", err)
	}
	
	if analysis.TotalMessages != 5 {
		t.Errorf("Expected 5 messages in analysis, got %d", analysis.TotalMessages)
	}
	
	// Test context stats
	stats := integration.GetContextStats(sess)
	if stats.TotalMessages != 5 {
		t.Errorf("Expected 5 messages in stats, got %d", stats.TotalMessages)
	}
	
	// Test enable/disable
	if !integration.IsEnabled() {
		t.Errorf("Expected integration to be enabled")
	}
	
	integration.DisableContextManagement()
	if integration.IsEnabled() {
		t.Errorf("Expected integration to be disabled")
	}
	
	integration.EnableContextManagement()
	if !integration.IsEnabled() {
		t.Errorf("Expected integration to be enabled again")
	}
}

func TestContextManagementSlashCommands(t *testing.T) {
	mockClient := NewMockLLMClient()
	integration := NewReactAgentContextIntegration(mockClient, nil)
	commands := NewContextManagementSlashCommands(integration)
	sess := createTestSession()
	
	ctx := context.Background()
	
	// Test context-status command
	result, err := commands.HandleSlashCommand(ctx, sess, "context-status", []string{})
	if err != nil {
		t.Fatalf("context-status command failed: %v", err)
	}
	
	if result == "" {
		t.Errorf("Expected non-empty result from context-status")
	}
	
	// Test context-stats command
	result, err = commands.HandleSlashCommand(ctx, sess, "context-stats", []string{})
	if err != nil {
		t.Fatalf("context-stats command failed: %v", err)
	}
	
	if result == "" {
		t.Errorf("Expected non-empty result from context-stats")
	}
	
	// Test context-enable command
	result, err = commands.HandleSlashCommand(ctx, sess, "context-enable", []string{})
	if err != nil {
		t.Fatalf("context-enable command failed: %v", err)
	}
	
	if result == "" {
		t.Errorf("Expected non-empty result from context-enable")
	}
	
	// Test unknown command
	_, err = commands.HandleSlashCommand(ctx, sess, "unknown-command", []string{})
	if err == nil {
		t.Errorf("Expected error for unknown command")
	}
}

func BenchmarkContextManager_CheckContextLength(b *testing.B) {
	mockClient := NewMockLLMClient()
	cm := NewContextManager(mockClient, nil)
	sess := createTestSession()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cm.CheckContextLength(sess)
		if err != nil {
			b.Fatalf("CheckContextLength failed: %v", err)
		}
	}
}

func BenchmarkContextManager_ProcessContextOverflow(b *testing.B) {
	mockClient := NewMockLLMClient()
	config := &ContextLengthConfig{
		MaxTokens: 50, // Very low to always trigger overflow
	}
	cm := NewContextManager(mockClient, config)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sess := createTestSession()
		ctx := context.Background()
		
		_, err := cm.ProcessContextOverflow(ctx, sess)
		if err != nil {
			b.Fatalf("ProcessContextOverflow failed: %v", err)
		}
	}
}