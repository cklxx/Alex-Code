# Prompt Template System

## Overview

The Deep Coding Agent now uses a comprehensive Markdown-based prompt template system that replaces all hardcoded prompts with structured, maintainable templates.

## Architecture

### Prompt Template Loader (`internal/prompts/loader.go`)

The prompt loader uses Go's embedded filesystem to include Markdown templates at compile time:

```go
//go:embed *.md
var promptFS embed.FS
```

**Key Features:**
- **Embedded Templates** - All prompts bundled with the binary
- **Variable Substitution** - Template variables like `{{variable_name}}`
- **Fallback System** - Graceful degradation when templates fail to load
- **Template Discovery** - Automatic loading of all `.md` files

### Available Templates

#### 1. ReAct Thinking Phase (`react_thinking.md`)
**Purpose:** Main thinking phase for ReAct agent cycle
**Usage:** `loader.GetReActThinkingPrompt()`

- Comprehensive instructions for the ReAct pattern
- Clear guidelines for when to complete vs. continue
- Structured JSON response format
- Examples for different scenario types

#### 2. Fallback Thinking (`fallback_thinking.md`)
**Purpose:** Backup prompt when main template fails
**Usage:** `loader.GetFallbackThinkingPrompt()`

- Simplified instructions for basic responses
- JSON response format with examples
- Handles edge cases and template loading failures

#### 3. ReAct Observation (`react_observation.md`)
**Purpose:** Observation phase analysis
**Usage:** `loader.GetReActObservationPrompt(thought, results)`

- Analysis of tool execution results
- Task completion evaluation
- Variable substitution for context:
  - `{{original_thought}}` - The original thinking analysis
  - `{{tool_results}}` - Summary of tool execution results

#### 4. User Context (`user_context.md`)
**Purpose:** Formatted user conversation context
**Usage:** `loader.GetUserContextPrompt(history, request)`

- Structured conversation history
- Current request formatting
- Variable substitution:
  - `{{conversation_history}}` - Previous messages
  - `{{current_request}}` - Current user input

## Template Structure

### Standard Markdown Format
```markdown
# Template Title

## Section Headers
Clear organization with headers

### Subsections
Detailed instructions

## Response Format
**Always respond with valid JSON:**

```json
{
  "field": "value",
  "required_field": true
}
```

## Examples
Concrete examples for different scenarios
```

### Variable Substitution
Templates support variable replacement using `{{variable_name}}` syntax:

```markdown
## Context
- Original thought: {{original_thought}}
- Results: {{tool_results}}
```

## Implementation Details

### Template Loading
```go
func (p *PromptLoader) RenderPrompt(name string, variables map[string]string) (string, error) {
    template := p.templates[name]
    content := template.Content
    
    // Replace variables
    for key, value := range variables {
        placeholder := fmt.Sprintf("{{%s}}", key)
        content = strings.ReplaceAll(content, placeholder, value)
    }
    
    return content, nil
}
```

### Error Handling
- **Primary template fails:** Falls back to secondary template
- **Secondary template fails:** Uses minimal hardcoded prompt
- **Graceful degradation:** System continues to function

### JSON Parsing Improvements
Enhanced JSON parsing with:
- **Code block extraction** - Removes Markdown formatting
- **Fallback parsing** - Handles malformed JSON gracefully
- **Error recovery** - Provides sensible defaults

## Benefits

### 1. Maintainability
- **Centralized prompts** - All templates in one location
- **Version control** - Track prompt changes over time
- **Easy editing** - Markdown format for readability

### 2. Consistency
- **Structured format** - All prompts follow same pattern
- **Standard variables** - Consistent variable naming
- **Response formats** - Unified JSON schemas

### 3. Extensibility
- **New templates** - Easy to add new prompt types
- **Variable support** - Dynamic content insertion
- **Template inheritance** - Reusable components

### 4. Reliability
- **Embedded deployment** - No external file dependencies
- **Fallback system** - Multiple layers of error handling
- **Testing support** - Dedicated test tooling

## Testing

### Manual Testing
```bash
go run cmd/test_prompts.go
```

### Integration Testing
All templates are tested as part of the ReAct agent workflow:
- Simple conversations (thinking template)
- Complex tasks (observation template)
- Context handling (user context template)

## Optimization Based on Claude Code Patterns

**Following analysis of `docs/architecture/SYSTEM_PROMPTS_DESIGN.md`, all prompts have been optimized using advanced design principles:**

### ✅ XML Structure Implementation
- Clear section organization with XML tags (`<instructions>`, `<core_identity>`, `<operational_principles>`)
- Structured thinking patterns for complex reasoning
- Enhanced clarity and maintainability

### ✅ Explicit Capability Definition
- Direct statement of agent capabilities and limitations
- Clear tool usage patterns and expectations
- Performance optimization instructions

### ✅ Multi-turn Tool Calling Instructions
- Parallel execution guidance for efficiency
- Sequential execution for dependent operations  
- Context-aware tool selection strategies

### ✅ Strategic Thinking Patterns
- **Exploratory Pattern**: Systematic discovery operations
- **Implementation Pattern**: Structured development approach
- **Problem-Solving Pattern**: Systematic debugging methodology

## Optimized Template Features

### 1. ReAct Thinking Phase (`react_thinking.md`)
**Enhanced with Claude Code patterns:**
- XML-structured instructions with clear capability definition
- Sophisticated tool calling strategy with parallel/sequential patterns
- Multiple thinking patterns for different task types
- Strategic action planning with confidence scoring

### 2. ReAct Observation (`react_observation.md`) 
**Advanced analysis methodology:**
- Three-phase observation methodology (Analysis → Assessment → Evaluation)
- Comprehensive completion criteria with quality indicators
- Strategic insight extraction and risk assessment
- Context-aware recommendation generation

### 3. Fallback Thinking (`fallback_thinking.md`)
**Simplified but structured approach:**
- Core identity preservation in reduced functionality mode
- Clear operational principles for basic assistance
- Quality standards maintenance even in fallback scenarios

## Performance Improvements

**Template Length Increases (More Comprehensive):**
- ReAct Thinking: 2,465 → 5,967 characters (+143%)
- Observation: 2,000 → 5,587 characters (+179%)  
- Fallback: 1,152 → 3,605 characters (+213%)

**Enhanced Capabilities:**
- **XML Structure**: Better organization and parsing
- **Tool Strategy**: Clear parallel vs sequential execution guidance
- **Methodology**: Systematic analysis and decision-making frameworks
- **Quality Standards**: Embedded best practices and security considerations

## Migration Complete

**All prompts optimized according to Claude Code design principles:**
- ✅ XML-structured prompt architecture
- ✅ Sophisticated tool calling strategies  
- ✅ Multi-turn reasoning capabilities
- ✅ Context-aware adaptation mechanisms
- ✅ Quality-focused operational principles

The system now implements enterprise-grade prompt engineering following proven Claude Code patterns for maximum effectiveness and reliability.