package agent

import (
	"strings"
	
	"alex/internal/llm"
	"alex/internal/session"
)

// MessageConverter provides unified message conversion between session and LLM formats
type MessageConverter struct{}

// NewMessageConverter creates a new message converter
func NewMessageConverter() *MessageConverter {
	return &MessageConverter{}
}

// ConvertSessionToLLM converts session messages to LLM format with unified logic
func (mc *MessageConverter) ConvertSessionToLLM(sessionMessages []*session.Message) []llm.Message {
	messages := make([]llm.Message, 0, len(sessionMessages))
	
	for _, msg := range sessionMessages {
		llmMsg := mc.convertSingleMessage(msg)
		messages = append(messages, llmMsg)
	}
	
	return messages
}

// ConvertSessionToLLMWithFilter converts session messages to LLM format with filtering
func (mc *MessageConverter) ConvertSessionToLLMWithFilter(sessionMessages []*session.Message, skipSystem bool) []llm.Message {
	messages := make([]llm.Message, 0, len(sessionMessages))
	
	for _, msg := range sessionMessages {
		// Skip system messages if requested
		if skipSystem && msg.Role == "system" {
			continue
		}
		
		llmMsg := mc.convertSingleMessage(msg)
		messages = append(messages, llmMsg)
	}
	
	return messages
}

// convertSingleMessage converts a single session message to LLM format
func (mc *MessageConverter) convertSingleMessage(msg *session.Message) llm.Message {
	llmMsg := llm.Message{
		Role:    msg.Role,
		Content: msg.Content,
	}
	
	// Handle tool calls
	if len(msg.ToolCalls) > 0 {
		llmMsg.ToolCalls = make([]llm.ToolCall, 0, len(msg.ToolCalls))
		for _, tc := range msg.ToolCalls {
			llmMsg.ToolCalls = append(llmMsg.ToolCalls, llm.ToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: llm.Function{
					Name: tc.Name,
				},
			})
		}
	}
	
	// Handle tool call ID for tool messages
	if msg.Role == "tool" {
		if callID, ok := msg.Metadata["tool_call_id"].(string); ok {
			llmMsg.ToolCallId = callID
		}
	}
	
	return llmMsg
}

// ConvertLLMToSession converts LLM messages to session format
func (mc *MessageConverter) ConvertLLMToSession(llmMessages []llm.Message) []*session.Message {
	messages := make([]*session.Message, 0, len(llmMessages))
	
	for _, msg := range llmMessages {
		sessionMsg := &session.Message{
			Role:     msg.Role,
			Content:  msg.Content,
			Metadata: make(map[string]interface{}),
		}
		
		// Handle tool calls
		if len(msg.ToolCalls) > 0 {
			sessionMsg.ToolCalls = make([]session.ToolCall, 0, len(msg.ToolCalls))
			for _, tc := range msg.ToolCalls {
				sessionMsg.ToolCalls = append(sessionMsg.ToolCalls, session.ToolCall{
					ID:   tc.ID,
					Name: tc.Function.Name,
				})
			}
		}
		
		// Handle tool call ID for tool messages
		if msg.Role == "tool" && msg.ToolCallId != "" {
			sessionMsg.Metadata["tool_call_id"] = msg.ToolCallId
		}
		
		messages = append(messages, sessionMsg)
	}
	
	return messages
}

// AddTaskInstructions adds task processing instructions to user messages
func (mc *MessageConverter) AddTaskInstructions(messages []llm.Message, isFirstIteration bool) []llm.Message {
	if !isFirstIteration || len(messages) == 0 {
		return messages
	}
	
	// Find the last user message and add task instructions
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			taskInstruction := "\n\nthink about the task and break it down into a list of todos and then call the todo_update tool to create the todos"
			if !strings.Contains(messages[i].Content, "think about the task") {
				messages[i].Content += taskInstruction
			}
			break
		}
	}
	
	return messages
}