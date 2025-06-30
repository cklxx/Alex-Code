# 如何运行Deep Coding Agent基准测试

## 前置条件检查

1. **构建代理**：
```bash
cd /Users/ckl/code/deep-coding
make build
```

2. **验证环境**：
```bash
# 检查代理可执行文件
ls -la deep-coding-agent

# 检查Python3
python3 --version
```

## 运行方式

### 方式一：快速测试（推荐）

```bash
cd benchmarks

# 编辑配置文件，减少测试数量
cat > config.json << EOF
{
  "agent_path": "../deep-coding-agent",
  "max_problems": 3,
  "output_dir": "results",
  "timeout_seconds": 60,
  "use_react_agent": false
}
EOF

# 运行基准测试
go run framework.go
```

### 方式二：完整测试

```bash
cd benchmarks

# 编辑配置文件，测试所有164个问题
cat > config.json << EOF
{
  "agent_path": "../deep-coding-agent",
  "max_problems": 164,
  "output_dir": "results",
  "timeout_seconds": 120,
  "use_react_agent": false
}
EOF

# 运行完整基准测试（需要较长时间）
nohup go run framework.go > benchmark.log 2>&1 &
```

### 方式三：指定配置文件

```bash
cd benchmarks

# 创建自定义配置
cat > my_config.json << EOF
{
  "agent_path": "../deep-coding-agent",
  "max_problems": 10,
  "output_dir": "my_results",
  "timeout_seconds": 90,
  "use_react_agent": true
}
EOF

# 使用自定义配置运行
go run framework.go my_config.json
```

## 配置参数说明

| 参数 | 说明 | 默认值 | 推荐值 |
|------|------|--------|--------|
| `agent_path` | 代理可执行文件路径 | `../deep-coding-agent` | 保持默认 |
| `max_problems` | 测试问题数量 | 10 | 3-10（测试），164（完整） |
| `output_dir` | 结果输出目录 | `results` | 保持默认 |
| `timeout_seconds` | 单个问题超时时间 | 30 | 60-120 |
| `use_react_agent` | 使用ReAct代理 | true | false（当前推荐） |

## 结果查看

运行完成后，查看结果：

```bash
cd benchmarks/results

# 查看摘要报告
cat report.txt

# 查看详细结果（JSON格式）
cat results.json | jq '.[0]'  # 查看第一个结果

# 统计通过率
jq '[.[] | select(.passed_tests == true)] | length' results.json
```

## 示例输出

成功运行后会看到类似输出：

```
Deep Coding Agent Benchmark Report
==========================================

Dataset: HumanEval
Total Problems: 3
Agent Path: ../deep-coding-agent
Use ReAct Agent: false

Results:
--------
Successfully Generated: 3/3 (100.0%)
Passed Tests: 2/3 (66.7%)
Average Duration: 1.2s
Total Duration: 3.6s

Pass@1 Rate: 0.667
```

## 故障排除

### 常见问题

1. **代理执行超时**：
   - 增加 `timeout_seconds` 值
   - 减少 `max_problems` 数量

2. **Python测试失败**：
   - 确保安装了Python3
   - 检查生成的代码格式

3. **工具验证错误**：
   - 设置 `use_react_agent: false`
   - 使用Legacy模式更稳定

4. **权限错误**：
   ```bash
   chmod +x ../deep-coding-agent
   ```

### 调试模式

```bash
# 启用调试输出
DEBUG=true go run framework.go

# 查看单个问题执行
go run simple_demo.go
```

## 性能期望

基于行业标准，期望性能：

- **GPT-4**: ~67% Pass@1
- **GPT-3.5**: ~48% Pass@1  
- **CodeT5+**: ~30% Pass@1
- **目标**: >50% Pass@1

## 扩展运行

### 批量测试不同配置

```bash
# 测试不同代理模式
for mode in true false; do
  echo "Testing with use_react_agent: $mode"
  jq ".use_react_agent = $mode" config.json > tmp.json && mv tmp.json config.json
  go run framework.go
  mv results results_react_${mode}
done
```

### 长时间运行监控

```bash
# 后台运行并监控进度
nohup go run framework.go > benchmark.log 2>&1 &

# 监控进度
tail -f benchmark.log

# 检查中间结果
watch -n 30 "jq 'length' results/results.json"
```

这样就可以完整地运行和监控基准测试了。