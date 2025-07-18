package agent

import (
	"testing"
	"time"

	"alex/internal/session"
)

func TestKeepRecentMessagesWithToolPairing(t *testing.T) {
	// 创建测试消息，模拟工具调用和响应序列
	messages := []*session.Message{
		{
			Role:      "user",
			Content:   "请帮我分析一下项目结构",
			Timestamp: time.Now().Add(-10 * time.Minute),
		},
		{
			Role:    "assistant",
			Content: "我来帮您分析项目结构",
			ToolCalls: []session.ToolCall{
				{ID: "call_1", Name: "file_list"},
				{ID: "call_2", Name: "file_read"},
			},
			Timestamp: time.Now().Add(-9 * time.Minute),
		},
		{
			Role:      "tool",
			Content:   "文件列表结果",
			Metadata:  map[string]interface{}{"tool_call_id": "call_1"},
			Timestamp: time.Now().Add(-8 * time.Minute),
		},
		{
			Role:      "tool",
			Content:   "文件读取结果",
			Metadata:  map[string]interface{}{"tool_call_id": "call_2"},
			Timestamp: time.Now().Add(-7 * time.Minute),
		},
		{
			Role:      "user",
			Content:   "这个分析很有帮助",
			Timestamp: time.Now().Add(-6 * time.Minute),
		},
		{
			Role:      "assistant",
			Content:   "很高兴能帮到您",
			Timestamp: time.Now().Add(-5 * time.Minute),
		},
	}

	// 创建MessageProcessor实例
	mp := &MessageProcessor{}

	// 测试工具调用配对逻辑
	originalCount := len(messages)
	if originalCount != 6 {
		t.Errorf("Expected 6 original messages, got %d", originalCount)
	}
	
	// 调用keepRecentMessagesWithToolPairing，保留3条消息
	result := mp.keepRecentMessagesWithToolPairing(messages, 3)
	
	if len(result) < 3 {
		t.Errorf("Expected at least 3 messages after processing, got %d", len(result))
	}
	
	// 验证工具调用和响应是否配对
	if !validateToolCallPairing(result) {
		t.Error("Tool call pairing validation failed")
	}
}

func validateToolCallPairing(messages []*session.Message) bool {
	toolCalls := make(map[string]bool)
	toolResponses := make(map[string]bool)
	
	for _, msg := range messages {
		// 收集工具调用ID
		for _, tc := range msg.ToolCalls {
			toolCalls[tc.ID] = true
		}
		
		// 收集工具响应ID
		if msg.Role == "tool" {
			if callId, ok := msg.Metadata["tool_call_id"].(string); ok {
				toolResponses[callId] = true
			}
		}
	}
	
	// 检查是否所有工具调用都有对应响应
	for callId := range toolCalls {
		if !toolResponses[callId] {
			return false
		}
	}
	
	// 检查是否所有工具响应都有对应调用
	for responseId := range toolResponses {
		if !toolCalls[responseId] {
			return false
		}
	}
	
	return true
}
