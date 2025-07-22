# 统一消息类型系统

## 概述

这个包提供了Alex项目中所有消息类型的统一定义和管理，解决了项目中多个重复的Message和ToolCall类型定义问题。

## 核心特性

- **统一的消息接口**：`BaseMessage`定义了所有消息的核心功能
- **统一的实现**：`Message`提供了完整的消息实现
- **协议适配**：支持LLM、Session、MCP等不同协议的转换
- **工具调用支持**：统一的`ToolCall`和`ToolResult`类型
- **Session集成**：Session可以使用Message作为子类型和元素
- **向后兼容**：提供适配器支持现有代码的渐进式迁移

## 类型层次结构

```
BaseMessage (interface)
├── Message (统一实现)
├── LLMMessage (LLM协议格式)
└── SessionMessage (Session存储格式)

ToolCall (interface)
├── ToolCallImpl (统一实现)
├── LLMToolCall (LLM协议格式)
└── SessionToolCall (Session存储格式)
```

## 基本使用

### 创建消息

```go
import "alex/pkg/types/message"

// 创建用户消息
userMsg := message.NewUserMessage("Hello, how can you help me?")

// 创建助手消息
assistantMsg := message.NewAssistantMessage("I can help you with coding tasks.")

// 创建系统消息
systemMsg := message.NewSystemMessage("You are a helpful assistant.")

// 创建工具响应消息
toolMsg := message.NewToolMessage("File content: ...", "call_123")
```

### 添加工具调用

```go
// 创建带工具调用的消息
msg := message.NewAssistantMessage("I'll read the file for you.")

// 添加工具调用
args := map[string]interface{}{
    "file_path": "/path/to/file.go",
    "start_line": 1,
    "end_line": 50,
}
msg.AddToolCallFromData("call_123", "file_read", args)
```

### 协议转换

```go
// 转换为LLM协议格式
llmMsg := msg.ToLLMMessage()

// 转换为Session存储格式
sessionMsg := msg.ToSessionMessage()

// 从LLM格式创建消息
msg2 := message.FromLLMMessage(llmMsg)

// 从Session格式创建消息
msg3 := message.FromSessionMessage(sessionMsg)
```

## Session集成

### 使用统一Message作为元素

```go
// 创建Session存储
session := message.NewSessionStorage("session_123")

// 直接添加统一Message
userMsg := message.NewUserMessage("Hello")
session.AddMessage(userMsg)

// 批量添加消息
messages := []*message.Message{userMsg, assistantMsg}
session.AddMessages(messages...)

// 查询消息
userMessages := session.GetMessagesByRole("user")
lastMessage := session.GetLastMessage()
toolMessages := session.GetMessagesWithToolCalls()
```

### 使用Message作为子类型

```go
// SessionMessage_Enhanced 展示了如何将Message作为子类型
enhanced := message.NewSessionMessageEnhanced("session_123", 1, userMsg)
enhanced.SessionID = "session_123"
enhanced.Index = 1
enhanced.SetMetadata("custom_field", "value")
```

## 适配器模式

### 批量转换

```go
adapter := message.NewAdapter()

// 转换LLM消息到统一格式
llmMessages := []message.LLMMessage{...}
unifiedMessages := adapter.ConvertLLMMessages(llmMessages)

// 转换统一格式到Session格式
sessionMessages := adapter.ConvertToSessionMessages(unifiedMessages)
```

### 消息集合管理

```go
collection := message.NewMessageCollection()
collection.AddMessages(msg1, msg2, msg3)

// 按角色过滤
userMsgs := collection.GetUserMessages()
assistantMsgs := collection.GetAssistantMessages()

// 查找带工具调用的消息
toolCallMsgs := collection.GetMessagesWithToolCalls()

// 转换整个集合
llmMsgs := collection.ToLLMMessages()
sessionMsgs := collection.ToSessionMessages()
```

## 工具调用管理

### 创建和使用ToolCall

```go
// 创建工具调用
args := map[string]interface{}{
    "query": "search term",
    "limit": 10,
}
toolCall := message.NewToolCall("call_456", "web_search", args)

// 获取参数
query, exists := toolCall.GetArgument("query")
jsonArgs := toolCall.GetArgumentsJSON()

// 转换格式
llmToolCall := toolCall.ToLLMToolCall()
sessionToolCall := toolCall.ToSessionToolCall()
```

### 创建工具结果

```go
// 成功结果
result := message.NewSuccessResult("call_456", "web_search", "Found 5 results", time.Second*2)
result.AddData("result_count", 5)

// 错误结果
errorResult := message.NewErrorResult("call_456", "web_search", "API rate limit exceeded", time.Second*1)
```

## 向后兼容性

### 兼容层

```go
// 为现有Session代码提供兼容层
compat := message.NewSessionCompatibilityLayer("session_123")

// 使用legacy格式添加消息
toolCalls := []message.SessionToolCall{...}
compat.AddLegacyMessage("user", "Hello", toolCalls)

// 获取legacy格式消息
legacyMsgs := compat.GetLegacyMessages()
llmMsgs := compat.GetLLMMessages()
```

## 迁移指南

### 阶段1：引入新类型
1. 在新代码中使用`message.Message`
2. 使用适配器在新旧格式间转换
3. 逐步替换创建消息的代码

### 阶段2：迁移存储层
1. 更新Session使用`message.SessionStorage`
2. 使用兼容层保持现有API不变
3. 逐步迁移数据格式

### 阶段3：清理
1. 移除重复的类型定义
2. 删除不再需要的转换代码
3. 更新所有引用

## 优势

1. **类型统一**：消除了项目中7个重复的Message和ToolCall类型
2. **维护性**：单一的真实来源，修改一处影响全局
3. **可扩展性**：清晰的接口设计，易于添加新功能
4. **向后兼容**：渐进式迁移，不影响现有功能
5. **Session集成**：Message既可作为元素也可作为子类型使用
6. **协议无关**：支持各种协议格式的自动转换

## 注意事项

1. 时间戳统一使用`time.Time`类型
2. 工具调用参数统一使用`map[string]interface{}`
3. JSON序列化/反序列化已经处理
4. 所有nil检查和默认值初始化已处理
5. 元数据字段支持任意扩展