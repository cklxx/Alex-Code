package message

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"alex/internal/llm"
	"alex/internal/session"
)

// MessageConverter handles conversion between different message formats
type MessageConverter struct{}

// NewMessageConverter creates a new MessageConverter
func NewMessageConverter() *MessageConverter {
	return &MessageConverter{}
}

// ConvertSessionToLLM converts session messages to LLM messages
func (mc *MessageConverter) ConvertSessionToLLM(sessionMessages []*session.Message) []llm.Message {
	var llmMessages []llm.Message

	for _, msg := range sessionMessages {
		llmMsg := llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}

		// Convert tool calls if present
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				var args string
				if tc.Args != nil {
					if argsBytes, err := json.Marshal(tc.Args); err == nil {
						args = string(argsBytes)
					}
				}

				llmMsg.ToolCalls = append(llmMsg.ToolCalls, llm.ToolCall{
					ID: tc.ID,
					Function: llm.Function{
						Name:       tc.Name,
						Parameters: args,
					},
				})
			}
		}

		llmMessages = append(llmMessages, llmMsg)
	}

	return llmMessages
}

// ConvertLLMToSession converts LLM messages to session messages
func (mc *MessageConverter) ConvertLLMToSession(llmMessages []llm.Message) []*session.Message {
	var sessionMessages []*session.Message

	for _, msg := range llmMessages {
		sessionMsg := &session.Message{
			Role:      msg.Role,
			Content:   msg.Content,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"source":    "llm_conversion",
				"timestamp": time.Now().Unix(),
			},
		}

		// Convert tool calls if present
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				var args map[string]interface{}
				if tc.Function.Parameters != nil {
					if paramStr, ok := tc.Function.Parameters.(string); ok {
						if err := json.Unmarshal([]byte(paramStr), &args); err != nil {
							// If unmarshal fails, store as raw string
							args = map[string]interface{}{"raw": paramStr}
						}
					}
				}

				sessionMsg.ToolCalls = append(sessionMsg.ToolCalls, session.ToolCall{
					ID:   tc.ID,
					Name: tc.Function.Name,
					Args: args,
				})
			}
		}

		sessionMessages = append(sessionMessages, sessionMsg)
	}

	return sessionMessages
}

// ConvertToDisplayFormat converts messages to a human-readable format
func (mc *MessageConverter) ConvertToDisplayFormat(messages []*session.Message) string {
	var parts []string

	for i, msg := range messages {
		var part strings.Builder

		// Add message header
		part.WriteString(fmt.Sprintf("Message %d [%s]:\n", i+1, msg.Role))

		// Add content
		if msg.Content != "" {
			part.WriteString(fmt.Sprintf("Content: %s\n", msg.Content))
		}

		// Add tool calls if present
		if len(msg.ToolCalls) > 0 {
			part.WriteString("Tool Calls:\n")
			for _, tc := range msg.ToolCalls {
				part.WriteString(fmt.Sprintf("  - %s (%s)\n", tc.Name, tc.ID))
				if tc.Args != nil {
					if argsStr, err := json.MarshalIndent(tc.Args, "    ", "  "); err == nil {
						part.WriteString(fmt.Sprintf("    Args: %s\n", string(argsStr)))
					}
				}
			}
		}

		// Add metadata if present
		if len(msg.Metadata) > 0 {
			part.WriteString("Metadata:\n")
			for key, value := range msg.Metadata {
				part.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
			}
		}

		parts = append(parts, part.String())
	}

	return strings.Join(parts, "\n---\n")
}

// SanitizeMessage removes sensitive information from messages
func (mc *MessageConverter) SanitizeMessage(msg *session.Message) *session.Message {
	sanitized := &session.Message{
		Role:      msg.Role,
		Content:   mc.sanitizeContent(msg.Content),
		Timestamp: msg.Timestamp,
		Metadata:  make(map[string]interface{}),
	}

	// Copy safe metadata
	for key, value := range msg.Metadata {
		if mc.isSafeMetadataKey(key) {
			sanitized.Metadata[key] = value
		}
	}

	// Copy and sanitize tool calls
	for _, tc := range msg.ToolCalls {
		sanitizedTC := session.ToolCall{
			ID:   tc.ID,
			Name: tc.Name,
			Args: mc.sanitizeArgs(tc.Args),
		}
		sanitized.ToolCalls = append(sanitized.ToolCalls, sanitizedTC)
	}

	return sanitized
}

// sanitizeContent removes sensitive patterns from content
func (mc *MessageConverter) sanitizeContent(content string) string {
	// Remove potential API keys, tokens, passwords
	sanitized := content

	// Simple pattern matching for sensitive data
	if strings.Contains(strings.ToLower(sanitized), "api") ||
		strings.Contains(strings.ToLower(sanitized), "token") ||
		strings.Contains(strings.ToLower(sanitized), "password") {
		sanitized = "[REDACTED]"
	}

	return sanitized
}

// sanitizeArgs removes sensitive information from tool arguments
func (mc *MessageConverter) sanitizeArgs(args map[string]interface{}) map[string]interface{} {
	if args == nil {
		return nil
	}

	sanitized := make(map[string]interface{})
	for key, value := range args {
		if mc.isSensitiveKey(key) {
			sanitized[key] = "[REDACTED]"
		} else {
			sanitized[key] = value
		}
	}

	return sanitized
}

// isSafeMetadataKey checks if a metadata key is safe to copy
func (mc *MessageConverter) isSafeMetadataKey(key string) bool {
	unsafeKeys := []string{
		"api_key", "token", "password", "secret",
		"auth", "credential", "private_key",
	}

	lowerKey := strings.ToLower(key)
	for _, unsafe := range unsafeKeys {
		if strings.Contains(lowerKey, unsafe) {
			return false
		}
	}

	return true
}

// isSensitiveKey checks if an argument key contains sensitive information
func (mc *MessageConverter) isSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"api_key", "token", "password", "secret",
		"auth", "credential", "private_key", "key",
	}

	lowerKey := strings.ToLower(key)
	for _, sensitive := range sensitiveKeys {
		if strings.Contains(lowerKey, sensitive) {
			return true
		}
	}

	return false
}
