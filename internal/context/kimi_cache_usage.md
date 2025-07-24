# Kimi API Context Caching 实现

## 概述

为 Kimi API (https://api.moonshot.cn/v1) 实现了自动上下文缓存功能，基于官方 Context Caching API，通过缓存 system message 来提高性能并降低 API 成本。

## 工作原理

1. **自动检测**: 当 base URL 为 `https://api.moonshot.cn/v1` 时，自动启用上下文缓存。

2. **缓存创建**: 首次请求时调用 `/v1/caching` API 创建缓存，缓存 system message。

3. **Headers 缓存使用**: 后续请求通过 `X-Msh-Context-Cache` header 引用缓存 ID。

4. **消息保持**: 根据 Kimi API 要求，请求中仍需包含被缓存的消息以通过 Hash 验证。

5. **TTL 刷新**: 每次使用缓存时自动刷新过期时间（3600秒）。

6. **生命周期管理**: 只在 CLI 完全退出时清理缓存，会话期间保持可用。

## 技术实现细节

### 缓存创建流程

```go
// 创建缓存请求
POST /v1/caching
{
  "model": "moonshot-v1",
  "messages": [
    {
      "role": "system", 
      "content": "system message content"
    }
  ],
  "ttl": 3600
}
```

### 缓存使用流程

```go
// 在 chat/completions 请求中使用 Headers
POST /v1/chat/completions
Headers:
  X-Msh-Context-Cache: cache-id-xxxxx
  X-Msh-Context-Cache-Reset-TTL: 3600

// 请求体必须包含与缓存完全一致的消息
{
  "model": "moonshot-v1-8k",
  "messages": [
    {
      "role": "system",
      "content": "system message content"  // 必须与缓存一致
    },
    {
      "role": "user", 
      "content": "new user message"
    }
  ]
}
```

### 核心文件

- `internal/llm/kimi_cache.go` - 核心缓存管理实现
- `internal/llm/http_client.go` - LLM 客户端集成
- `internal/session/session.go` - 会话中的缓存 ID 存储

### 关键方法

1. **KimiCacheManager.CreateCacheIfNeeded()** - 创建缓存（如果不存在）
2. **KimiCacheManager.PrepareRequestWithCache()** - 准备缓存 Headers
3. **HTTPLLMClient.setHeaders()** - 设置缓存 Headers

## 使用方式

缓存功能完全透明，无需用户配置：

```bash
# 当配置使用 Kimi API 时，自动启用缓存
./alex -i  # 交互模式

# 首次请求会创建缓存
# 后续请求自动使用缓存，节省 tokens
```

## 优势特点

- **性能提升**: 减少重复传输 system message，降低延迟
- **成本节省**: 缓存命中时节省 tokens 消费
- **自动管理**: 无需手动配置，自动创建和清理
- **错误处理**: 缓存失败时不影响正常 API 调用
- **TTL 刷新**: 活跃会话自动延长缓存时间

## 调试信息

开启 DEBUG 日志可以看到缓存操作：

```
DEBUG: Created new Kimi cache for session session_xxx with cache ID: cache-xxx
DEBUG: Prepared request for session session_xxx to use cache cache-xxx via Headers
DEBUG: Set cache header X-Msh-Context-Cache: cache-xxx
DEBUG: Set cache header X-Msh-Context-Cache-Reset-TTL: 3600
```

## Headers 缓存的限制条件

### 🔴 **严格要求**

根据 Kimi API 官方文档，使用 Headers 方式缓存有以下限制：

1. **消息前缀完全一致**
   ```
   请求的 messages 前 N 个必须与缓存的所有 messages 完全一致
   (N = 缓存 messages 长度)
   - 消息顺序必须一致
   - 每个字段值必须完全相同
   ```

2. **Tools 完全一致**
   ```
   请求的 tools 必须与缓存的 tools 完全一致
   ```

3. **Hash 校验机制**
   ```
   Kimi API 使用 Hash 校验：
   - messages 前缀与缓存是否一致
   - tools 与缓存是否一致
   校验失败将无法命中缓存
   ```

4. **必须重发缓存消息**
   ```
   即使消息已缓存，请求中仍需包含所有缓存的消息
   不能省略任何缓存的消息
   ```

### ⚠️ **实现保障**

当前实现通过以下机制确保合规：

1. **存储完整缓存内容**: 保存所有缓存的 messages 和 tools
2. **一致性验证**: 请求前验证消息和工具是否匹配
3. **自动缓存更新**: 内容变化时自动删除旧缓存并创建新缓存
4. **失败降级**: 验证失败时自动跳过缓存，正常发送请求

## 其他注意事项

1. **缓存时间**: 默认缓存时间为 1 小时，每次使用时刷新
2. **清理时机**: 只在 CLI 完全退出时清理，不在请求间清理
3. **错误处理**: 缓存操作失败不影响正常 API 调用
4. **性能优化**: 只有符合条件的请求才会启用缓存

## API 规格

实现严格遵循 Kimi 官方 Context Caching API 文档：
- 缓存创建: `POST /v1/caching`
- 缓存删除: `DELETE /v1/caching/{cache-id}`
- Headers 使用: `X-Msh-Context-Cache` 和 `X-Msh-Context-Cache-Reset-TTL`