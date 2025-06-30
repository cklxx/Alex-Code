package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	// Test the new API that doesn't require config
	client, err := NewHTTPClient()
	if err != nil {
		t.Fatalf("NewHTTPClient() failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewHTTPClient() returned nil client")
	}
}

func TestHTTPClient_Chat(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response
		response := ChatResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "test-model",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "Test response",
					},
					FinishReason: "stop",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewHTTPClient()
	if err != nil {
		t.Fatalf("NewHTTPClient() failed: %v", err)
	}

	// Test with config in request
	config := &Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "test-model",
	}

	req := &ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "test message"},
		},
		ModelType: BasicModel,
		Config:    config,
	}

	ctx := context.Background()
	response, err := client.Chat(ctx, req)
	if err != nil {
		t.Fatalf("Chat() failed: %v", err)
	}

	if response == nil {
		t.Fatal("Chat() returned nil response")
	}

	if len(response.Choices) == 0 {
		t.Fatal("Chat() returned empty choices")
	}

	if response.Choices[0].Message.Content != "Test response" {
		t.Errorf("Expected 'Test response', got '%s'", response.Choices[0].Message.Content)
	}
}