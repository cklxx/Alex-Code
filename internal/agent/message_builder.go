package agent

import (
	"context"

	"alex/internal/llm"
	"alex/internal/session"
)

// MessageBuilder 负责构建各种类型的消息列表
type MessageBuilder struct {
	contextHandler *ContextHandler
	converter      *MessageConverter
}

// NewMessageBuilder 创建消息构建器
func NewMessageBuilder(contextHandler *ContextHandler) *MessageBuilder {
	return &MessageBuilder{
		contextHandler: contextHandler,
		converter:      NewMessageConverter(),
	}
}

// buildMessagesFromSession - 基于会话历史构建消息列表，支持多轮对话
func (mb *MessageBuilder) buildMessagesFromSession(sess *session.Session, currentTask string, systemPrompt string) []llm.Message {
	var messages []llm.Message

	// 添加系统提示
	messages = append(messages, llm.Message{
		Role:    "system",
		Content: systemPrompt,
	})

	// 如果有会话历史，添加完整的对话历史（用于多轮处理）
	if sess != nil {
		sessionMessages := sess.GetMessages()
		isFirstIteration := len(sessionMessages) == 1 // 只有初始用户消息

		// 使用统一转换器处理会话消息，跳过系统消息
		convertedMessages := mb.converter.ConvertSessionToLLMWithFilter(sessionMessages, true)
		messages = append(messages, convertedMessages...)

		// 如果是第一次迭代，添加任务处理指令
		if isFirstIteration {
			messages = mb.converter.AddTaskInstructions(messages, true)
		}
	} else {
		// 没有会话历史，添加当前任务作为初始用户消息
		messages = append(messages, llm.Message{
			Role:    "user",
			Content: currentTask + "\n\nthink about the task and break it down into a list of todos and then call the todo_update tool to create the todos",
		})
	}

	return messages
}

// updateMessagesWithSessionContent - 基于现有messages更新为包含最新session内容，支持智能压缩（memory统一处理）
func (mb *MessageBuilder) updateMessagesWithSessionContent(ctx context.Context, sess *session.Session, baseMessages []llm.Message) []llm.Message {
	if sess == nil {
		return baseMessages
	}

	// 提取系统消息
	var systemMsg llm.Message
	if len(baseMessages) > 0 && baseMessages[0].Role == "system" {
		systemMsg = baseMessages[0]
	}

	// 获取session消息
	sessionMessages := sess.GetMessages()

	// 将memory信息融入到session消息中进行统一处理
	allMessages := mb.integrateMemoryIntoMessages(ctx, sessionMessages)

	// 对整合后的消息进行智能压缩
	compression := NewContextCompression(mb.contextHandler)
	processedMessages := compression.intelligentContextCompression(allMessages)

	// 构建最终的消息列表
	var messages []llm.Message

	// 添加系统消息
	if systemMsg.Role == "system" {
		messages = append(messages, systemMsg)
	}

	// 使用统一转换器处理压缩后的消息，跳过系统消息
	convertedMessages := mb.converter.ConvertSessionToLLMWithFilter(processedMessages, true)
	messages = append(messages, convertedMessages...)

	return messages
}

// integrateMemoryIntoMessages - 将memory信息整合到session消息中
func (mb *MessageBuilder) integrateMemoryIntoMessages(ctx context.Context, sessionMessages []*session.Message) []*session.Message {
	memoryIntegration := NewMemoryIntegration(mb.contextHandler)
	return memoryIntegration.integrateMemoryIntoMessages(ctx, sessionMessages)
}
