package llm

import (
	"context"
	"io"
	"time"
)

// Message represents a chat message
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallId string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`

	// OpenAI reasoning fields (2025 Responses API)
	Reasoning        string `json:"reasoning,omitempty"`
	ReasoningSummary string `json:"reasoning_summary,omitempty"`
	Think            string `json:"think,omitempty"`
}

// ChatRequest represents a request to the LLM
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`

	// Tool calling support
	Tools      []Tool `json:"tools,omitempty"`
	ToolChoice string `json:"tool_choice,omitempty"`
	// Model type selection for multi-model configurations - not serialized to JSON
	ModelType ModelType `json:"-"`

	// Config for dynamic configuration resolution - not serialized to JSON
	Config *Config `json:"-"`
}

// ChatResponse represents a response from the LLM
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage,omitempty"`
}

// Choice represents a choice in the response
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message,omitempty"`
	Delta        Message `json:"delta,omitempty"`
	FinishReason string  `json:"finish_reason,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamDelta represents a streaming response chunk
type StreamDelta struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// ModelType represents different model usage types
type ModelType string

const (
	BasicModel     ModelType = "basic"     // For general tasks, fast responses
	ReasoningModel ModelType = "reasoning" // For complex reasoning, tool calls, high-quality content
)

// ModelConfig represents configuration for a specific model
type ModelConfig struct {
	BaseURL     string  `json:"base_url"`
	Model       string  `json:"model"`
	APIKey      string  `json:"api_key"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

// Config represents LLM client configuration with multi-model support
type Config struct {
	// Default single model config (backward compatibility)
	APIKey      string        `json:"api_key,omitempty"`
	BaseURL     string        `json:"base_url,omitempty"`
	Model       string        `json:"model,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Timeout     time.Duration `json:"timeout,omitempty"`

	// Multi-model configurations
	Models map[ModelType]*ModelConfig `json:"models,omitempty"`

	// Default model type to use when none specified
	DefaultModelType ModelType `json:"default_model_type,omitempty"`
}

// Client interface defines LLM client operations
type Client interface {
	// Chat sends a chat request and returns the response
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ChatStream sends a chat request and returns a streaming response
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamDelta, error)

	// Close closes the client and cleans up resources
	Close() error
}

// StreamReader interface for reading streaming responses
type StreamReader interface {
	io.Reader
	io.Closer
}

// ToolCall represents an OpenAI-standard tool call
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents a function definition or call
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
	Arguments   string      `json:"arguments,omitempty"`
}

// Tool represents a tool definition
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}
