package message

import (
	"encoding/json"
	"testing"
	"time"
)

// TestMessageCreation tests basic message creation
func TestMessageCreation(t *testing.T) {
	// Test user message
	userMsg := NewUserMessage("Hello")
	if userMsg.GetRole() != "user" {
		t.Errorf("Expected role 'user', got '%s'", userMsg.GetRole())
	}
	if userMsg.GetContent() != "Hello" {
		t.Errorf("Expected content 'Hello', got '%s'", userMsg.GetContent())
	}
	
	// Test assistant message
	assistantMsg := NewAssistantMessage("Hi there!")
	if assistantMsg.GetRole() != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", assistantMsg.GetRole())
	}
	
	// Test tool message
	toolMsg := NewToolMessage("Result", "call_123")
	if toolMsg.GetRole() != "tool" {
		t.Errorf("Expected role 'tool', got '%s'", toolMsg.GetRole())
	}
	if toolMsg.GetToolCallID() != "call_123" {
		t.Errorf("Expected tool call ID 'call_123', got '%s'", toolMsg.GetToolCallID())
	}
}

// TestToolCallCreation tests tool call creation and usage
func TestToolCallCreation(t *testing.T) {
	args := map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	}
	
	toolCall := NewToolCall("call_456", "test_tool", args)
	
	if toolCall.GetID() != "call_456" {
		t.Errorf("Expected ID 'call_456', got '%s'", toolCall.GetID())
	}
	
	if toolCall.GetName() != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", toolCall.GetName())
	}
	
	if value, exists := toolCall.GetArgument("param1"); !exists || value != "value1" {
		t.Errorf("Expected argument 'param1' = 'value1', got exists=%v, value=%v", exists, value)
	}
	
	jsonArgs := toolCall.GetArgumentsJSON()
	if jsonArgs == "" || jsonArgs == "{}" {
		t.Error("Expected non-empty JSON arguments")
	}
}

// TestMessageWithToolCalls tests messages with tool calls
func TestMessageWithToolCalls(t *testing.T) {
	msg := NewAssistantMessage("I'll help you with that.")
	
	// Add tool call
	args := map[string]interface{}{
		"query": "test query",
	}
	msg.AddToolCallFromData("call_789", "search", args)
	
	toolCalls := msg.GetToolCalls()
	if len(toolCalls) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(toolCalls))
	}
	
	if toolCalls[0].GetName() != "search" {
		t.Errorf("Expected tool name 'search', got '%s'", toolCalls[0].GetName())
	}
}

// TestProtocolConversion tests conversion between different formats
func TestProtocolConversion(t *testing.T) {
	// Create original message
	original := NewAssistantMessage("Test message")
	original.AddToolCallFromData("call_001", "test_tool", map[string]interface{}{
		"param": "value",
	})
	
	// Convert to LLM format
	llmMsg := original.ToLLMMessage()
	if llmMsg.Role != original.GetRole() {
		t.Errorf("LLM conversion role mismatch: expected %s, got %s", original.GetRole(), llmMsg.Role)
	}
	if len(llmMsg.ToolCalls) != len(original.GetToolCalls()) {
		t.Errorf("LLM conversion tool calls mismatch: expected %d, got %d", len(original.GetToolCalls()), len(llmMsg.ToolCalls))
	}
	
	// Convert back from LLM format
	converted := FromLLMMessage(llmMsg)
	if converted.GetRole() != original.GetRole() {
		t.Errorf("Round-trip role mismatch: expected %s, got %s", original.GetRole(), converted.GetRole())
	}
	if converted.GetContent() != original.GetContent() {
		t.Errorf("Round-trip content mismatch: expected %s, got %s", original.GetContent(), converted.GetContent())
	}
	
	// Convert to Session format
	sessionMsg := original.ToSessionMessage()
	if sessionMsg.Role != original.GetRole() {
		t.Errorf("Session conversion role mismatch: expected %s, got %s", original.GetRole(), sessionMsg.Role)
	}
	
	// Convert back from Session format
	convertedFromSession := FromSessionMessage(sessionMsg)
	if convertedFromSession.GetRole() != original.GetRole() {
		t.Errorf("Session round-trip role mismatch: expected %s, got %s", original.GetRole(), convertedFromSession.GetRole())
	}
}

// TestSessionStorage tests session storage functionality
func TestSessionStorage(t *testing.T) {
	session := NewSessionStorage("test_session")
	
	// Add messages
	userMsg := NewUserMessage("Question")
	assistantMsg := NewAssistantMessage("Answer")
	session.AddMessages(userMsg, assistantMsg)
	
	if session.Count() != 2 {
		t.Errorf("Expected 2 messages in session, got %d", session.Count())
	}
	
	// Test filtering
	userMessages := session.GetMessagesByRole("user")
	if len(userMessages) != 1 {
		t.Errorf("Expected 1 user message, got %d", len(userMessages))
	}
	
	lastMessage := session.GetLastMessage()
	if lastMessage == nil || lastMessage.GetRole() != "assistant" {
		t.Error("Expected last message to be assistant message")
	}
}

// TestJSONMarshalUnmarshal tests JSON marshaling/unmarshaling
func TestJSONMarshalUnmarshal(t *testing.T) {
	original := NewAssistantMessage("Test JSON")
	original.AddMetadata("test_key", "test_value")
	original.AddToolCallFromData("call_123", "test_tool", map[string]interface{}{
		"param": "value",
	})
	
	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}
	
	// Unmarshal from JSON
	var restored Message
	if err := json.Unmarshal(jsonData, &restored); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}
	
	// Verify data integrity
	if original.GetRole() != restored.GetRole() {
		t.Errorf("Role mismatch after JSON round-trip: expected %s, got %s", original.GetRole(), restored.GetRole())
	}
	
	if original.GetContent() != restored.GetContent() {
		t.Errorf("Content mismatch after JSON round-trip: expected %s, got %s", original.GetContent(), restored.GetContent())
	}
	
	if len(original.GetToolCalls()) != len(restored.GetToolCalls()) {
		t.Errorf("Tool calls count mismatch after JSON round-trip: expected %d, got %d", 
			len(original.GetToolCalls()), len(restored.GetToolCalls()))
	}
}

// TestMessageCollection tests message collection utilities
func TestMessageCollection(t *testing.T) {
	collection := NewMessageCollection()
	
	userMsg := NewUserMessage("Question 1")
	assistantMsg := NewAssistantMessage("Answer 1")
	assistantWithTools := NewAssistantMessage("Let me search")
	assistantWithTools.AddToolCallFromData("call_1", "search", map[string]interface{}{
		"query": "test",
	})
	
	collection.AddMessages(userMsg, assistantMsg, assistantWithTools)
	
	if collection.Count() != 3 {
		t.Errorf("Expected 3 messages in collection, got %d", collection.Count())
	}
	
	userMessages := collection.GetUserMessages()
	if len(userMessages) != 1 {
		t.Errorf("Expected 1 user message in collection, got %d", len(userMessages))
	}
	
	assistantMessages := collection.GetAssistantMessages()
	if len(assistantMessages) != 2 {
		t.Errorf("Expected 2 assistant messages in collection, got %d", len(assistantMessages))
	}
	
	toolMessages := collection.GetMessagesWithToolCalls()
	if len(toolMessages) != 1 {
		t.Errorf("Expected 1 message with tool calls, got %d", len(toolMessages))
	}
}

// TestToolResult tests tool result creation and usage
func TestToolResult(t *testing.T) {
	// Test success result
	successResult := NewSuccessResult("call_123", "test_tool", "Success content", time.Second*2)
	if !successResult.GetSuccess() {
		t.Error("Expected success result to be successful")
	}
	if successResult.GetContent() != "Success content" {
		t.Errorf("Expected content 'Success content', got '%s'", successResult.GetContent())
	}
	if successResult.GetDuration() != time.Second*2 {
		t.Errorf("Expected duration 2s, got %v", successResult.GetDuration())
	}
	
	// Test error result
	errorResult := NewErrorResult("call_456", "test_tool", "Error occurred", time.Second*1)
	if errorResult.GetSuccess() {
		t.Error("Expected error result to be unsuccessful")
	}
	if errorResult.GetError() != "Error occurred" {
		t.Errorf("Expected error 'Error occurred', got '%s'", errorResult.GetError())
	}
	
	// Test adding data
	successResult.AddData("result_count", 5)
	data := successResult.GetData()
	if count, exists := data["result_count"]; !exists || count != 5 {
		t.Errorf("Expected result_count=5, got exists=%v, count=%v", exists, count)
	}
}