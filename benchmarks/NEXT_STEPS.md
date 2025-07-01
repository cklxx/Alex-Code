# Deep Coding Agent 基准测试 - 下一步执行计划

## 🎯 立即执行任务

基于当前基准测试框架已成功运行（Pass@1 = 0.333），以下是按优先级排序的下一步行动计划：

---

## 📅 第一周：核心功能增强

### 🔥 任务1：集成真实AI提供者 【最高优先级】
**目标**: 替换mock实现，测试实际代理性能

**具体步骤**:
```bash
# 1. 配置API密钥
export OPENAI_API_KEY="your-openai-key"
# 或
export ARK_API_KEY="your-ark-key"

# 2. 修改代理配置
cd /Users/ckl/code/deep-coding
./deep-coding-agent --config --set aiProvider=openai
```

**修改代码**:
```go
// 在benchmarks/framework.go中修改agent调用
cmd := exec.Command(b.config.AgentPath, 
    "--format", "text", 
    "--temperature", "0.1",
    agentPrompt)
env = append(env, "USE_LEGACY_AGENT=false") // 使用AI代理而非mock
```

**预期结果**: 获得真实的Pass@1性能基线

### 🚀 任务2：优化代理提示词 【高优先级】
**目标**: 提高代码生成质量

**当前问题分析**:
- 代理可能返回解释而非纯代码
- 缺乏明确的格式要求
- 没有强调测试用例重要性

**改进方案**:
```go
// 更精确的提示词模板
agentPrompt := fmt.Sprintf(`You are a Python coding expert. Complete this function implementation.

TASK: %s

REQUIREMENTS:
1. Return ONLY the function body code (indented with 4 spaces)
2. Do NOT include the function signature or docstring
3. Ensure code passes all given test cases
4. Use efficient, readable algorithms
5. Handle edge cases appropriately

EXAMPLE FORMAT:
    # Your implementation here
    result = some_algorithm()
    return result

IMPLEMENTATION:`, problem.Prompt)
```

### 📈 任务3：扩展Mock解决方案覆盖
**目标**: 增加更多HumanEval问题的参考实现

**扩展计划**:
```go
// 新增解决方案到extractFunctionFromPrompt
case "below_zero":
    return `    balance = 0
    for operation in operations:
        balance += operation
        if balance < 0:
            return True
    return False`

case "mean_absolute_deviation":
    return `    mean = sum(numbers) / len(numbers)
    return sum(abs(x - mean) for x in numbers) / len(numbers)`

// ... 继续添加更多问题
```

**目标**: 覆盖前20个HumanEval问题

---

## 📅 第二周：基准扩展

### 🎯 任务4：增强评估指标
**目标**: 提供更详细的性能分析

**新增功能**:
```go
type DetailedMetrics struct {
    PassAtK          map[int]float64    // Pass@1, Pass@5, Pass@10
    CategoryStats    map[string]float64 // 按问题类型统计
    ErrorAnalysis    ErrorBreakdown     // 错误类型分析
    PerformanceStats PerformanceMetrics // 性能统计
}

type ErrorBreakdown struct {
    SyntaxErrors    int `json:"syntax_errors"`
    LogicErrors     int `json:"logic_errors"`
    TimeoutErrors   int `json:"timeout_errors"`
    ImportErrors    int `json:"import_errors"`
}
```

### 📊 任务5：添加MBPP基准支持
**目标**: 支持Google MBPP数据集

**实施步骤**:
```bash
# 1. 下载MBPP数据
cd benchmarks
curl -L https://raw.githubusercontent.com/google-research/google-research/master/mbpp/mbpp.jsonl -o mbpp.jsonl

# 2. 实现MBPP加载器
```

```go
type MBPPProblem struct {
    TaskID      int    `json:"task_id"`
    Text        string `json:"text"`
    Code        string `json:"code"`
    TestSetup   string `json:"test_setup"`
    TestList    []string `json:"test_list"`
    Challenge   bool   `json:"challenge"`
}

func loadMBPPProblems(path string) ([]MBPPProblem, error) {
    // 实现MBPP数据加载
}
```

---

## 📅 第三周：对比分析

### 📋 任务6：创建对比分析报告
**目标**: 生成详细的性能对比报告

**报告结构**:
```go
type BenchmarkReport struct {
    Summary      ReportSummary      `json:"summary"`
    Comparison   IndustryComparison `json:"comparison"`
    Analysis     DetailedAnalysis   `json:"analysis"`
    Recommendations []string        `json:"recommendations"`
}

type IndustryComparison struct {
    HumanEval map[string]float64 `json:"humaneval"` // vs GPT-4, GPT-3.5, CodeT5+
    MBPP      map[string]float64 `json:"mbpp"`
}
```

**报告模板**:
```markdown
# Deep Coding Agent 性能分析报告

## 执行摘要
- HumanEval Pass@1: **XX%** (vs GPT-4: 67%, GPT-3.5: 48%)
- 相对GPT-3.5提升: **+XX%**
- 强项: 数组操作、字符串处理
- 改进空间: 复杂逻辑、数学计算

## 详细分析
### 问题类型表现
- 算法题: XX% (XX/XX)
- 字符串: XX% (XX/XX)
- 数学: XX% (XX/XX)

## 改进建议
1. 优化提示词模板
2. 增强错误处理逻辑
3. 添加代码质量检查
```

---

## 📅 可选扩展任务

### 🔧 SWE-bench集成（如需要）
**目标**: 支持真实GitHub问题修复

**挑战**:
- 多文件上下文理解
- 复杂代码库导航
- 实际bug定位和修复

**实施考虑**:
```bash
# SWE-bench lite版本（更易集成）
git clone https://github.com/princeton-nlp/SWE-bench
cd SWE-bench
python -m swebench.collect --dataset_name princeton-nlp/SWE-bench_Lite
```

---

## 🎯 成功指标与时间线

### 第一周目标
- [ ] 集成真实AI提供者 ✅
- [ ] 优化提示词，Pass@1 > 50% 
- [ ] 扩展到10个HumanEval问题的mock实现

### 第二周目标  
- [ ] 支持MBPP基准测试
- [ ] 实现详细错误分析
- [ ] 建立性能趋势跟踪

### 第三周目标
- [ ] 生成完整对比分析报告
- [ ] Pass@1 > 60% (HumanEval)
- [ ] 识别关键改进领域

## 🚀 立即开始

**下一个行动**:
```bash
# 1. 配置AI提供者
export OPENAI_API_KEY="your-key"

# 2. 修改配置文件
cd benchmarks
jq '.max_problems = 10' config.json > tmp.json && mv tmp.json config.json

# 3. 运行扩展测试
go run framework.go
```

**预期结果**: 
- 获得10个问题的真实性能基线
- 识别当前代理的优势和不足
- 为后续优化提供数据支撑

这个计划专注于核心功能增强和实际性能提升，避免了不必要的CI/CD复杂性，确保在3周内获得有价值的基准测试能力。