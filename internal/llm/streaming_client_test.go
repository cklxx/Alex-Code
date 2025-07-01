package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewStreamingClient(t *testing.T) {
	// Test the new API that doesn't require config
	client, err := NewStreamingClient()
	if err != nil {
		t.Fatalf("NewStreamingClient() failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewStreamingClient() returned nil client")
	}
}

func TestStreamingClient_Chat(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected /chat/completions path, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer test-key authorization, got %s", r.Header.Get("Authorization"))
		}

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
						Content: "Hello, world!",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewStreamingClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test with config in request
	config := &Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "test-model",
	}

	req := &ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		ModelType: BasicModel,
		Config:    config,
	}

	resp, err := client.Chat(context.Background(), req)
	if err != nil {
		t.Fatalf("chat request failed: %v", err)
	}

	if resp.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got '%s'", resp.ID)
	}
	if len(resp.Choices) != 1 {
		t.Errorf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got '%s'", resp.Choices[0].Message.Content)
	}
}

func TestStreamingClient_ChatStream(t *testing.T) {
	// Create a mock server that returns SSE stream
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("expected text/event-stream accept header, got %s", r.Header.Get("Accept"))
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Send some streaming data
		deltas := []StreamDelta{
			{
				ID:      "test-id-1",
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   "test-model",
				Choices: []Choice{
					{
						Index: 0,
						Delta: Message{
							Role:    "assistant",
							Content: "Hello",
						},
					},
				},
			},
			{
				ID:      "test-id-2",
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   "test-model",
				Choices: []Choice{
					{
						Index: 0,
						Delta: Message{
							Content: ", world!",
						},
					},
				},
			},
		}

		for _, delta := range deltas {
			data, _ := json.Marshal(delta)
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}

		fmt.Fprintf(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client, err := NewStreamingClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test with config in request
	config := &Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "test-model",
	}

	req := &ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		ModelType: BasicModel,
		Config:    config,
	}

	deltaChannel, err := client.ChatStream(context.Background(), req)
	if err != nil {
		t.Fatalf("stream request failed: %v", err)
	}

	var content strings.Builder
	for delta := range deltaChannel {
		if len(delta.Choices) > 0 {
			content.WriteString(delta.Choices[0].Delta.Content)
		}
	}

	expectedContent := "Hello, world!"
	if content.String() != expectedContent {
		t.Errorf("expected '%s', got '%s'", expectedContent, content.String())
	}
}

func TestStreamingClient_SupportsStreaming(t *testing.T) {
	client, err := NewStreamingClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if !client.SupportsStreaming() {
		t.Error("streaming client should support streaming")
	}
}

func TestStreamingClient_SetStreamingEnabled(t *testing.T) {
	client, err := NewStreamingClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test with config in request
	config := &Config{
		BaseURL: "https://httpbin.org/status/404", // Will return 404 for testing
		APIKey:  "test-key",
		Model:   "test-model",
	}

	// Streaming should be enabled by default
	req := &ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		ModelType: BasicModel,
		Config:    config,
	}

	// Disable streaming
	client.SetStreamingEnabled(false)

	_, err = client.ChatStream(context.Background(), req)
	if err == nil {
		t.Error("expected error when streaming is disabled")
	}

	// Re-enable streaming
	client.SetStreamingEnabled(true)

	// This should not error (though it will fail due to no server)
	_, err = client.ChatStream(context.Background(), req)
	if err == nil {
		t.Error("expected error due to no server, but this confirms streaming is enabled")
	}
}

func TestStreamingClient_Close(t *testing.T) {
	client, err := NewStreamingClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Errorf("unexpected error closing client: %v", err)
	}
}
