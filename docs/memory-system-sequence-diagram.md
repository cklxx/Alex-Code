# Memory System Sequence Diagram

## ReactAgent Memory System Call Flow

```mermaid
sequenceDiagram
    participant User
    participant ReactAgent as ReactAgent
    participant MemoryManager as MemoryManager
    participant ShortTerm as ShortTermMemory
    participant LongTerm as LongTermMemory
    participant Controller as MemoryController
    participant Compressor as ContextCompressor
    participant LLM as LLM Client
    participant Session as Session
    participant ReactCore as ReactCore
    participant ContextHandler as ContextHandler

    Note over User, ContextHandler: User Message Processing Flow

    User->>ReactAgent: ProcessMessage(userMessage)
    ReactAgent->>Session: AddMessage(userMsg)
    
    Note over ReactAgent, MemoryManager: Memory Recall Phase (50ms timeout)
    ReactAgent->>ReactAgent: enhanceContextWithMemory()
    ReactAgent->>MemoryManager: Recall(query)
    MemoryManager->>Controller: FilterMemoriesForRecall()
    MemoryManager->>ShortTerm: Search(query)
    ShortTerm-->>MemoryManager: shortTermResults
    MemoryManager->>LongTerm: Search(query)
    LongTerm-->>MemoryManager: longTermResults
    MemoryManager-->>ReactAgent: RecallResult{Items: []}
    ReactAgent->>ReactAgent: context.WithValue("memories", result)

    Note over ReactAgent, ContextHandler: Context Management
    ReactAgent->>ReactAgent: manageContextWithMemory()
    ReactAgent->>ReactAgent: estimateTokenUsage()
    alt Token usage > 80% of max
        ReactAgent->>MemoryManager: ProcessContextCompression()
        MemoryManager->>Compressor: Compress(sessionID, messages)
        Compressor->>LLM: Chat(compressionPrompt)
        LLM-->>Compressor: CompressedSummary
        Compressor-->>MemoryManager: CompressionResult
        MemoryManager-->>ReactAgent: CompressionResult
    end

    Note over ReactAgent, ReactCore: Task Execution
    ReactAgent->>ReactCore: SolveTask(ctx, userMessage, callback)
    ReactCore->>ContextHandler: buildMessagesWithMemoryContext()
    ContextHandler->>ContextHandler: formatMemoryAsMessage(memories)
    ContextHandler-->>ReactCore: messages[]
    ReactCore->>LLM: Chat(messages, tools)
    LLM-->>ReactCore: Response
    
    Note over ReactCore: Tool Call Processing & Memory Integration
    alt LLM Response contains tool calls
        ReactCore->>ReactCore: addMessageToSession(assistantMsg)
        ReactCore->>ReactCore: executeSerialToolsStream()
        ReactCore->>ReactCore: addToolMessagesToSession(toolMessages, toolResults)
        ReactCore->>ReactCore: createToolUsageMemory() [async]
    else Direct answer
        ReactCore->>ReactCore: addMessageToSession(assistantMsg)
    end
    
    ReactCore-->>ReactAgent: ReactTaskResult

    Note over ReactAgent, Session: Response Generation
    ReactAgent->>Session: AddMessage(assistantMsg)

    Note over ReactAgent, MemoryManager: Memory Creation Phase (Async)
    ReactAgent->>ReactAgent: createMemoryAsync() [goroutine]
    
    par Memory Creation for User Message
        ReactAgent->>MemoryManager: CreateMemoryFromMessage(userMsg)
        MemoryManager->>Controller: ShouldCreateMemory()
        Controller-->>MemoryManager: decision
        alt Should create memory
            MemoryManager->>Controller: ClassifyMemory()
            Controller-->>MemoryManager: category, importance, tags
            MemoryManager->>MemoryManager: createMemoryItem()
            MemoryManager->>ShortTerm: Store(memoryItem)
            ShortTerm-->>MemoryManager: success
        end
    and Memory Creation for Assistant Message
        ReactAgent->>MemoryManager: CreateMemoryFromMessage(assistantMsg)
        MemoryManager->>Controller: ShouldCreateMemory()
        Controller-->>MemoryManager: decision
        alt Should create memory
            MemoryManager->>Controller: ClassifyMemory()
            Controller-->>MemoryManager: category, importance, tags
            MemoryManager->>MemoryManager: createMemoryItem()
            MemoryManager->>ShortTerm: Store(memoryItem)
            ShortTerm-->>MemoryManager: success
        end
    and Task Execution Memory
        ReactAgent->>ReactAgent: createTaskExecutionMemory()
        ReactAgent->>MemoryManager: Store(toolPatternMemory)
        MemoryManager->>ShortTerm: Store(memoryItem)
    end

    Note over ReactAgent, MemoryManager: Memory Maintenance (Async)
    ReactAgent->>MemoryManager: AutomaticMemoryMaintenance()
    MemoryManager->>ShortTerm: Cleanup()
    MemoryManager->>Controller: ShouldPromoteToLongTerm()
    alt Should promote
        MemoryManager->>LongTerm: Store(memoryItem)
        MemoryManager->>ShortTerm: Remove(memoryItem)
    end

    ReactAgent-->>User: Response{message, toolResults, sessionID}
```

## Key Components Interaction

### 1. Memory Recall Flow (Synchronous - 50ms timeout)

```mermaid
sequenceDiagram
    participant ReactAgent
    participant MemoryManager
    participant Query as MemoryQuery
    participant ShortTerm
    participant LongTerm
    participant Controller

    ReactAgent->>MemoryManager: Recall(query)
    MemoryManager->>Query: Build query with categories
    MemoryManager->>ShortTerm: Search(query.content)
    ShortTerm->>ShortTerm: calculateRelevanceScore()
    ShortTerm-->>MemoryManager: relevantItems[]
    MemoryManager->>LongTerm: Search(query.content)
    LongTerm->>LongTerm: calculateRelevanceScore()
    LongTerm-->>MemoryManager: relevantItems[]
    MemoryManager->>Controller: FilterMemoriesForRecall()
    Controller->>Controller: calculateRelevanceScore()
    Controller-->>MemoryManager: filteredItems[]
    MemoryManager-->>ReactAgent: RecallResult{items, totalFound}
```

### 2. Context Compression Flow (When needed)

```mermaid
sequenceDiagram
    participant ReactAgent
    participant MemoryManager
    participant Compressor
    participant LLM
    participant Session

    ReactAgent->>MemoryManager: ProcessContextCompression()
    MemoryManager->>Session: GetMessages()
    MemoryManager->>Compressor: Compress(sessionID, messages)
    Compressor->>Compressor: identifyImportantMessages()
    Compressor->>LLM: Chat(compressionPrompt)
    Note over LLM: Uses AU2-style compression<br/>with importance scoring
    LLM-->>Compressor: compressedSummary
    Compressor->>Compressor: calculateTokensSaved()
    Compressor-->>MemoryManager: CompressionResult
    MemoryManager->>Session: ReplaceMessages(compressed)
    MemoryManager-->>ReactAgent: CompressionResult
```

### 3. Tool Call Message Processing Flow

```mermaid
sequenceDiagram
    participant ReactCore
    participant Session
    participant MemoryManager
    participant LLM
    participant ToolExecutor

    Note over ReactCore: Process LLM Response with Tool Calls
    ReactCore->>ReactCore: parseToolCalls(response)
    
    alt Has tool calls
        ReactCore->>ReactCore: addMessageToSession(assistantMsg)
        Note over ReactCore: Convert LLM message to session format
        ReactCore->>Session: AddMessage(sessionMsg with tool_calls)
        
        ReactCore->>ToolExecutor: executeSerialToolsStream(toolCalls)
        ToolExecutor-->>ReactCore: toolResults[]
        
        ReactCore->>ReactCore: addToolMessagesToSession(toolMessages, toolResults)
        loop For each tool message
            ReactCore->>ReactCore: Convert to session format
            ReactCore->>ReactCore: Add tool metadata (name, success, timing)
            ReactCore->>Session: AddMessage(toolSessionMsg)
        end
        
        Note over ReactCore: Async Tool Usage Memory Creation
        ReactCore->>ReactCore: createToolUsageMemory() [goroutine]
        
        par Successful Tools Memory
            ReactCore->>MemoryManager: Store(toolUsageMemory)
            Note over MemoryManager: Category: TaskHistory<br/>Importance: 0.7<br/>Tags: [tool_usage, success, toolNames...]
        and Failed Tools Memory
            ReactCore->>MemoryManager: Store(toolFailureMemory)
            Note over MemoryManager: Category: ErrorPatterns<br/>Importance: 0.8<br/>Tags: [tool_failure, error, toolNames...]
        end
    else Direct answer only
        ReactCore->>ReactCore: addMessageToSession(assistantMsg)
        ReactCore->>Session: AddMessage(sessionMsg)
    end
```

### 4. Memory Creation Flow (Asynchronous)

```mermaid
sequenceDiagram
    participant ReactAgent
    participant MemoryManager
    participant Controller
    participant ShortTerm
    participant LongTerm
    participant Message

    ReactAgent->>MemoryManager: CreateMemoryFromMessage()
    MemoryManager->>Controller: ShouldCreateMemory()
    Controller->>Controller: checkMessageCount()
    Controller->>Controller: checkRateLimit()
    Controller->>Controller: checkContentLength()
    Controller-->>MemoryManager: shouldCreate
    
    alt Should create memory
        MemoryManager->>Controller: ClassifyMemory()
        Controller->>Controller: determineCategory()
        Controller->>Controller: calculateImportance()
        Controller->>Controller: generateTags()
        Controller-->>MemoryManager: category, importance, tags
        
        MemoryManager->>MemoryManager: createMemoryItem()
        MemoryManager->>ShortTerm: Store(memoryItem)
        ShortTerm->>ShortTerm: addToLRUCache()
        ShortTerm-->>MemoryManager: stored
        
        Note over MemoryManager: Check promotion to long-term
        MemoryManager->>Controller: ShouldPromoteToLongTerm()
        alt Should promote
            MemoryManager->>LongTerm: Store(memoryItem)
            MemoryManager->>ShortTerm: Remove(memoryItem)
        end
    end
```

## Memory Categories and Processing

### Memory Classification Logic

```mermaid
flowchart TD
    A[Message Content] --> B{Contains Error Keywords?}
    B -->|Yes| C[ErrorPatterns Category]
    B -->|No| D{Contains Code Keywords?}
    D -->|Yes| E[CodeContext Category]
    D -->|No| F{Contains Solution Keywords?}
    F -->|Yes| G[Solutions Category]
    F -->|No| H{Contains Preference Keywords?}
    H -->|Yes| I[UserPreferences Category]
    H -->|No| J{Has Tool Calls?}
    J -->|Yes| K[TaskHistory Category]
    J -->|No| L[Knowledge Category]
    
    C --> M[Importance += 0.3]
    E --> N[Importance += 0.2]
    G --> O[Importance += 0.3]
    I --> P[Importance += 0.1]
    K --> Q[Importance += 0.1]
    L --> R[Base Importance: 0.5]
```

## Performance Considerations

### Memory Recall Timeout Protection

```mermaid
sequenceDiagram
    participant ReactAgent
    participant SafeRecall as safeMemoryRecall
    participant MemoryManager
    participant Timeout as 50ms Timer

    ReactAgent->>SafeRecall: safeMemoryRecall(query, 50ms)
    SafeRecall->>MemoryManager: Recall(query) [goroutine]
    SafeRecall->>Timeout: Start 50ms timer
    
    alt Memory recall completes in time
        MemoryManager-->>SafeRecall: RecallResult
        SafeRecall-->>ReactAgent: RecallResult
    else Timeout occurs
        Timeout-->>SafeRecall: timeout signal
        SafeRecall-->>ReactAgent: EmptyRecallResult
        Note over SafeRecall: Logs warning but continues
    end
```

## Error Handling and Fallbacks

```mermaid
flowchart TD
    A[Memory Operation] --> B{Memory System Available?}
    B -->|No| C[Continue without memory]
    B -->|Yes| D[Execute memory operation]
    D --> E{Operation successful?}
    E -->|No| F[Log warning]
    E -->|Yes| G[Use memory result]
    F --> C
    G --> H[Enhanced processing]
    C --> I[Standard processing]
    H --> J[Response to user]
    I --> J
```

## Enhanced Tool Call Processing

### New Features Added

**Tool Message Session Integration**: All LLM responses and tool execution results are now properly added to the session, enabling the memory system to learn from:

1. **Assistant Messages**: LLM responses with tool call information
2. **Tool Messages**: Tool execution results with metadata (success/failure, execution time, errors)
3. **Tool Usage Patterns**: Automatic memory creation for successful tool chains
4. **Tool Failure Patterns**: High-importance memory creation for failed tools to avoid repeated errors

### Tool Memory Categories

```
Tool Usage Memory (TaskHistory):
- Importance: 0.7
- Tags: [tool_usage, success, <tool_names>]
- Content: "Successfully used tools: file_read, todo_update (total time: 150ms)"

Tool Failure Memory (ErrorPatterns):
- Importance: 0.8 (higher to prevent repeat failures)
- Tags: [tool_failure, error, <tool_names>]
- Content: "Failed tools: invalid_tool"
```

### Session Message Enhancement

Tool-related messages now include rich metadata:
- **Tool Call Tracking**: Conversion of LLM tool calls to session format
- **Execution Metrics**: Duration, success/failure status
- **Error Information**: Detailed error messages for failed tools
- **Tool Identification**: Tool names and arguments for analysis

## Summary

This memory system provides:

1. **Intelligent Context Management**: Automatic memory recall and context compression
2. **Performance Optimization**: Async memory operations with timeout protection
3. **Graceful Degradation**: Continues working even if memory system fails
4. **Smart Classification**: 6-category memory system with importance scoring
5. **Dual Storage**: Short-term LRU cache + long-term persistent storage
6. **Seamless Integration**: Transparent memory enhancement of ReactAgent responses
7. **🆕 Tool Call Learning**: Complete tool usage pattern learning and error prevention
8. **🆕 Rich Tool Metadata**: Comprehensive tool execution tracking and analysis

The system maintains the <30ms response time target by using async memory creation and 50ms timeout for memory recall operations. Tool call processing adds comprehensive learning capabilities without impacting response performance.