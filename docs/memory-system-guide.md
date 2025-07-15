# Alex 记忆系统设计与实现指南

## 概述

基于对Claude Code项目的深度调研，Alex现已实现了完整的上下文压缩和智能记忆系统。该系统能够智能控制记忆写入、分类存储不同类型的记忆，并支持高效的记忆召回和上下文压缩。

## 核心特性

### 1. 双层记忆架构

**短期记忆 (Short-term Memory)**
- 存储时间：24小时TTL
- 容量限制：1000条记忆项
- 适用场景：会话内临时信息，工具执行记录，用户偏好

**长期记忆 (Long-term Memory)**
- 存储时间：永久存储（定期清理）
- 容量管理：智能清理低重要性记忆
- 适用场景：重要解决方案，代码模式，错误处理经验

### 2. 智能记忆控制

**写入控制规则**
```
- 最少消息数：3条消息后开始创建记忆
- 最小重要性：0.3分以上才存储
- 频率限制：每小时最多50条记忆
- 内容过滤：跳过短消息（<20字符）和无意义内容
```

**分类系统**
- **CodeContext**: 代码实现、函数定义、API用法
- **ErrorPatterns**: 错误信息、异常处理、调试经验
- **Solutions**: 解决方案、修复方法、最佳实践
- **TaskHistory**: 工具执行记录、操作历史
- **UserPreferences**: 用户偏好、习惯设置
- **Knowledge**: 通用知识、技术概念

### 3. 上下文压缩系统

**压缩触发条件**
- Token使用率达到80%（可配置）
- 保留最近3-5条消息（可配置）
- 使用LLM进行智能压缩

**压缩算法**
- 基于重要性分数的消息筛选
- 语义聚类保留关键信息
- 结构化输出包含总结、要点、代码变更等

## 使用示例

### 基础集成

```go
// 创建记忆管理器
memoryMgr, err := memory.NewMemoryManager(llmClient)
if err != nil {
    return fmt.Errorf("failed to create memory manager: %w", err)
}

// 在ReactAgent中添加记忆功能
type ReactAgent struct {
    llm         llm.Client
    memory      *memory.MemoryManager
    // ... 其他字段
}
```

### 智能记忆创建

```go
// 处理新消息时自动创建记忆
func (agent *ReactAgent) ProcessMessage(ctx context.Context, msg *session.Message) error {
    // 创建记忆（自动控制是否创建）
    memories, err := agent.memory.CreateMemoryFromMessage(
        ctx, 
        agent.sessionID, 
        msg, 
        len(agent.session.Messages),
    )
    if err != nil {
        return fmt.Errorf("memory creation failed: %w", err)
    }
    
    log.Printf("Created %d memories from message", len(memories))
    return nil
}
```

### 上下文压缩

```go
// 检查是否需要压缩并执行
func (agent *ReactAgent) manageContext(ctx context.Context) error {
    maxTokens := 100000 // 根据模型动态获取
    
    result, err := agent.memory.ProcessContextCompression(
        ctx, 
        agent.session, 
        maxTokens,
    )
    if err != nil {
        return fmt.Errorf("context compression failed: %w", err)
    }
    
    if result.CompressedSummary != "" {
        log.Printf("Compressed %d messages to %d, saved %d tokens", 
            result.OriginalCount, 
            result.CompressedCount, 
            result.TokensSaved)
    }
    
    return nil
}
```

### 记忆召回与合并

```go
// 在发送给LLM前合并相关记忆
func (agent *ReactAgent) prepareMessages(ctx context.Context, userMessage string) ([]*session.Message, error) {
    recentMessages := agent.session.GetMessages()
    
    // 合并相关记忆到消息中
    mergedMessages, err := agent.memory.MergeMemoriesToMessages(
        ctx,
        agent.sessionID,
        recentMessages,
        5, // 最多5条相关记忆
    )
    if err != nil {
        return nil, fmt.Errorf("memory merge failed: %w", err)
    }
    
    return mergedMessages, nil
}
```

### 自动维护

```go
// 定期执行记忆维护
func (agent *ReactAgent) periodicMaintenance() {
    go func() {
        ticker := time.NewTicker(time.Hour)
        defer ticker.Stop()
        
        for range ticker.C {
            if err := agent.memory.AutomaticMemoryMaintenance(agent.sessionID); err != nil {
                log.Printf("Memory maintenance failed: %v", err)
            }
        }
    }()
}
```

## 性能指标

### 基准测试结果

```
Memory Creation: ~93μs/op, 4KB内存/操作
Memory Recall:   ~4ms/op,  2MB内存/操作（包含全索引搜索）
```

### 存储效率

- **压缩比例**: 通常达到70-80%的压缩率
- **内存占用**: 短期记忆<10MB，长期记忆按需加载
- **查询性能**: 索引化搜索，支持分类和标签过滤

## 配置选项

### 记忆控制配置

```go
config := &memory.MemoryControlConfig{
    MinMessageCount:    3,      // 最少消息数
    MinImportanceScore: 0.3,    // 最小重要性
    MaxMemoriesPerHour: 50,     // 每小时限制
    MinContentLength:   20,     // 最短内容长度
    
    // 自定义关键词
    ImportantKeywords: []string{"error", "solution", "implement"},
    SkipKeywords:     []string{"hello", "thanks", "ok"},
}
```

### 压缩配置

```go
compressionConfig := &memory.CompressionConfig{
    Threshold:         0.8,   // 80%使用率触发
    CompressionRatio:  0.3,   // 压缩到30%
    PreserveRecent:    5,     // 保留5条最近消息
    MinImportance:     0.5,   // 最小重要性保留
    EnableLLMCompress: true,  // 启用LLM智能压缩
}
```

## 最佳实践

### 1. 记忆写入控制

✅ **推荐做法**
- 让系统自动判断是否创建记忆
- 重要操作（错误、解决方案）手动标记高重要性
- 定期执行自动维护清理无用记忆

❌ **避免做法**
- 强制为每条消息创建记忆
- 忽略重要性分数设置
- 长期不执行维护操作

### 2. 分类策略

✅ **推荐做法**
- 使用系统自动分类
- 为专业领域添加自定义关键词
- 定期review长期记忆的分类准确性

### 3. 性能优化

✅ **推荐做法**
- 使用分类和标签进行高效查询
- 合理设置查询限制（limit参数）
- 定期执行vacuum操作清理长期记忆

## 故障排除

### 常见问题

**Q: 记忆创建过多导致性能问题**
A: 调整`MaxMemoriesPerHour`和`MinImportanceScore`参数

**Q: 上下文压缩效果不佳**
A: 检查`EnableLLMCompress`是否启用，调整`CompressionRatio`参数

**Q: 记忆召回不准确**
A: 优化查询关键词，调整`MinImportance`过滤条件

### 监控指标

```go
// 获取内存统计
stats := memoryManager.GetMemoryStats()
log.Printf("Total memories: %d, Size: %d bytes", 
    stats["total_items"], stats["total_size"])

// 监控压缩效果
compressionResult, _ := memoryManager.ProcessContextCompression(ctx, session, maxTokens)
log.Printf("Compression ratio: %.2f, Tokens saved: %d", 
    compressionResult.CompressionRatio, compressionResult.TokensSaved)
```

## 技术细节

### 存储结构

```
~/.deep-coding-memory/
├── long-term/           # 长期记忆文件存储
│   ├── item_id1.json
│   └── item_id2.json
└── sessions/            # 会话相关临时文件
```

### 索引机制

- **分类索引**: 按MemoryCategory快速查找
- **标签索引**: 支持多标签组合查询
- **时间索引**: 按创建时间和访问时间排序
- **重要性索引**: 按重要性分数过滤

### 线程安全

所有记忆操作都是线程安全的，使用了读写锁保护：
- 读操作（查询、召回）支持并发
- 写操作（创建、更新、删除）独占访问
- 维护操作在独立goroutine中执行

## 总结

Alex的记忆系统实现了Claude Code级别的智能上下文管理，具备以下核心优势：

1. **智能控制**: 自动判断何时创建记忆，避免信息过载
2. **精准分类**: 6种记忆类型覆盖编程助手的主要使用场景  
3. **高效压缩**: LLM驱动的语义压缩，保持信息完整性
4. **快速召回**: 多重索引支持的高效查询系统
5. **生产就绪**: 完整的测试覆盖和性能优化

该系统为Alex提供了处理长期对话和复杂项目的能力，显著提升了AI助手的实用性和智能化水平。