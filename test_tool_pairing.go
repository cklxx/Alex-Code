package main

import (
	"fmt"
	"time"

	"alex/internal/agent"
	"alex/internal/session"
)

func main() {
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
	mp := &agent.MessageProcessor{}

	// 测试工具调用配对逻辑
	fmt.Println("原始消息数量:", len(messages))
	
	// 调用keepRecentMessagesWithToolPairing，保留3条消息
	result := mp.KeepRecentMessagesWithToolPairing(messages, 3)
	
	fmt.Println("保留后消息数量:", len(result))
	
	// 验证结果
	for i, msg := range result {
		fmt.Printf("消息 %d: Role=%s, Content=%s", i, msg.Role, msg.Content[:min(20, len(msg.Content))])
		if len(msg.ToolCalls) > 0 {
			fmt.Printf(", ToolCalls=%d", len(msg.ToolCalls))
		}
		if toolCallId, ok := msg.Metadata["tool_call_id"].(string); ok {
			fmt.Printf(", ToolCallId=%s", toolCallId)
		}
		fmt.Println()
	}
	
	// 验证工具调用和响应是否配对
	validateToolCallPairing(result)
}

func validateToolCallPairing(messages []*session.Message) {
	fmt.Println("\n验证工具调用配对:")
	
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
			fmt.Printf("❌ 工具调用 %s 没有对应响应\n", callId)
		} else {
			fmt.Printf("✅ 工具调用 %s 有对应响应\n", callId)
		}
	}
	
	// 检查是否所有工具响应都有对应调用
	for responseId := range toolResponses {
		if !toolCalls[responseId] {
			fmt.Printf("❌ 工具响应 %s 没有对应调用\n", responseId)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}