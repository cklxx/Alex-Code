package message

import (
	"encoding/json"
)

// Adapter provides conversion utilities between different message formats
type Adapter struct{}

// NewAdapter creates a new message adapter
func NewAdapter() *Adapter {
	return &Adapter{}
}

// ConvertLLMMessages converts slice of LLM messages to unified messages
func (a *Adapter) ConvertLLMMessages(llmMessages []LLMMessage) []*Message {
	messages := make([]*Message, len(llmMessages))
	for i, llmMsg := range llmMessages {
		messages[i] = FromLLMMessage(llmMsg)
	}
	return messages
}

// ConvertToLLMMessages converts slice of unified messages to LLM messages
func (a *Adapter) ConvertToLLMMessages(messages []*Message) []LLMMessage {
	llmMessages := make([]LLMMessage, len(messages))
	for i, msg := range messages {
		llmMessages[i] = msg.ToLLMMessage()
	}
	return llmMessages
}

// ConvertSessionMessages converts slice of session messages to unified messages
func (a *Adapter) ConvertSessionMessages(sessionMessages []SessionMessage) []*Message {
	messages := make([]*Message, len(sessionMessages))
	for i, sessionMsg := range sessionMessages {
		messages[i] = FromSessionMessage(sessionMsg)
	}
	return messages
}

// ConvertToSessionMessages converts slice of unified messages to session messages
func (a *Adapter) ConvertToSessionMessages(messages []*Message) []SessionMessage {
	sessionMessages := make([]SessionMessage, len(messages))
	for i, msg := range messages {
		sessionMessages[i] = msg.ToSessionMessage()
	}
	return sessionMessages
}

// BatchConvertToolCalls converts multiple tool calls
func (a *Adapter) BatchConvertToolCalls(toolCalls []*ToolCallImpl) ([]LLMToolCall, []SessionToolCall) {
	llmToolCalls := make([]LLMToolCall, len(toolCalls))
	sessionToolCalls := make([]SessionToolCall, len(toolCalls))

	for i, tc := range toolCalls {
		llmToolCalls[i] = tc.ToLLMToolCall()
		sessionToolCalls[i] = tc.ToSessionToolCall()
	}

	return llmToolCalls, sessionToolCalls
}

// LegacyLLMMessage represents the legacy LLM message format
type LegacyLLMMessage struct {
	Role             string              `json:"role"`
	Content          string              `json:"content,omitempty"`
	ToolCalls        []LegacyLLMToolCall `json:"tool_calls,omitempty"`
	ToolCallId       string              `json:"tool_call_id,omitempty"`
	Name             string              `json:"name,omitempty"`
	Reasoning        string              `json:"reasoning,omitempty"`
	ReasoningSummary string              `json:"reasoning_summary,omitempty"`
	Think            string              `json:"think,omitempty"`
}

// LegacyLLMToolCall represents the legacy LLM tool call format
type LegacyLLMToolCall struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Function LegacyLLMFunction `json:"function"`
}

// LegacyLLMFunction represents the legacy LLM function format
type LegacyLLMFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
	Arguments   string      `json:"arguments,omitempty"`
}

// LegacySessionMessage represents the legacy session message format
type LegacySessionMessage struct {
	Role      string                  `json:"role"`
	Content   string                  `json:"content"`
	ToolCalls []LegacySessionToolCall `json:"tool_calls,omitempty"`
	ToolID    string                  `json:"tool_id,omitempty"`
	Metadata  map[string]interface{}  `json:"metadata,omitempty"`
	Timestamp string                  `json:"timestamp"` // Legacy uses string timestamp
}

// LegacySessionToolCall represents the legacy session tool call format
type LegacySessionToolCall struct {
	ID   string                 `json:"id"`
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// ConvertFromLegacyLLM converts legacy LLM message to unified message
func (a *Adapter) ConvertFromLegacyLLM(legacyMsg LegacyLLMMessage) *Message {
	// Convert to new format first
	newMsg := LLMMessage{
		Role:             legacyMsg.Role,
		Content:          legacyMsg.Content,
		ToolCallID:       legacyMsg.ToolCallId,
		Name:             legacyMsg.Name,
		Reasoning:        legacyMsg.Reasoning,
		ReasoningSummary: legacyMsg.ReasoningSummary,
		Think:            legacyMsg.Think,
	}

	// Convert tool calls
	for _, legacyTC := range legacyMsg.ToolCalls {
		newTC := LLMToolCall{
			ID:   legacyTC.ID,
			Type: legacyTC.Type,
			Function: LLMFunction{
				Name:        legacyTC.Function.Name,
				Description: legacyTC.Function.Description,
				Parameters:  legacyTC.Function.Parameters,
				Arguments:   legacyTC.Function.Arguments,
			},
		}
		newMsg.ToolCalls = append(newMsg.ToolCalls, newTC)
	}

	return FromLLMMessage(newMsg)
}

// ConvertFromLegacySession converts legacy session message to unified message
func (a *Adapter) ConvertFromLegacySession(legacyMsg LegacySessionMessage) *Message {
	// Convert to new format first
	newMsg := SessionMessage{
		Role:     legacyMsg.Role,
		Content:  legacyMsg.Content,
		ToolID:   legacyMsg.ToolID,
		Metadata: legacyMsg.Metadata,
	}

	// Parse timestamp from string if possible
	// This is a simplified conversion - you might want to handle different timestamp formats
	// newMsg.Timestamp = parseTimestamp(legacyMsg.Timestamp)

	// Convert tool calls
	for _, legacyTC := range legacyMsg.ToolCalls {
		newTC := SessionToolCall(legacyTC)
		newMsg.ToolCalls = append(newMsg.ToolCalls, newTC)
	}

	return FromSessionMessage(newMsg)
}

// MessageCollection provides utilities for working with message collections
type MessageCollection struct {
	Messages []*Message `json:"messages"`
}

// NewMessageCollection creates a new message collection
func NewMessageCollection() *MessageCollection {
	return &MessageCollection{
		Messages: make([]*Message, 0),
	}
}

// Add adds a message to the collection
func (mc *MessageCollection) Add(message *Message) {
	mc.Messages = append(mc.Messages, message)
}

// AddMessages adds multiple messages to the collection
func (mc *MessageCollection) AddMessages(messages ...*Message) {
	mc.Messages = append(mc.Messages, messages...)
}

// GetByRole filters messages by role
func (mc *MessageCollection) GetByRole(role string) []*Message {
	var filtered []*Message
	for _, msg := range mc.Messages {
		if msg.GetRole() == role {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// GetUserMessages returns all user messages
func (mc *MessageCollection) GetUserMessages() []*Message {
	return mc.GetByRole(string(MessageTypeUser))
}

// GetAssistantMessages returns all assistant messages
func (mc *MessageCollection) GetAssistantMessages() []*Message {
	return mc.GetByRole(string(MessageTypeAssistant))
}

// GetSystemMessages returns all system messages
func (mc *MessageCollection) GetSystemMessages() []*Message {
	return mc.GetByRole(string(MessageTypeSystem))
}

// GetToolMessages returns all tool messages
func (mc *MessageCollection) GetToolMessages() []*Message {
	return mc.GetByRole(string(MessageTypeTool))
}

// GetMessagesWithToolCalls returns messages that have tool calls
func (mc *MessageCollection) GetMessagesWithToolCalls() []*Message {
	var filtered []*Message
	for _, msg := range mc.Messages {
		if len(msg.GetToolCalls()) > 0 {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// ToLLMMessages converts all messages to LLM format
func (mc *MessageCollection) ToLLMMessages() []LLMMessage {
	adapter := NewAdapter()
	return adapter.ConvertToLLMMessages(mc.Messages)
}

// ToSessionMessages converts all messages to session format
func (mc *MessageCollection) ToSessionMessages() []SessionMessage {
	adapter := NewAdapter()
	return adapter.ConvertToSessionMessages(mc.Messages)
}

// Count returns the number of messages in the collection
func (mc *MessageCollection) Count() int {
	return len(mc.Messages)
}

// Clear removes all messages from the collection
func (mc *MessageCollection) Clear() {
	mc.Messages = make([]*Message, 0)
}

// JSON support for MessageCollection
func (mc *MessageCollection) MarshalJSON() ([]byte, error) {
	type Alias MessageCollection
	return json.Marshal((*Alias)(mc))
}

func (mc *MessageCollection) UnmarshalJSON(data []byte) error {
	type Alias MessageCollection
	aux := (*Alias)(mc)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Ensure messages slice is initialized
	if mc.Messages == nil {
		mc.Messages = make([]*Message, 0)
	}

	return nil
}
