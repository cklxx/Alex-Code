#!/bin/bash

# SWE-Bench 批处理使用示例脚本
# 演示了 Alex SWE-Bench 批处理功能的各种用法

set -e

echo "==================================="
echo "Alex SWE-Bench 批处理使用示例"
echo "==================================="
echo

# 检查 alex 二进制文件是否存在
if [ ! -f "../../alex" ]; then
    echo "错误：找不到 alex 二进制文件"
    echo "请先运行 'make build' 构建项目"
    exit 1
fi

ALEX_BIN="../../alex"

echo "1. 快速测试 - 运行 2 个实例验证功能"
echo "----------------------------------------"
$ALEX_BIN run-batch \
    --dataset.subset lite \
    --dataset.split dev \
    --instance-limit 2 \
    --workers 1 \
    --output ./quick_test_results \
    --quiet

echo "✓ 快速测试完成，结果保存在 ./quick_test_results/"
echo

echo "2. 生成配置文件模板"
echo "---------------------"
if [ ! -f "./batch_config.yaml" ]; then
    cp config.example.yaml batch_config.yaml
    echo "✓ 配置文件模板已生成: ./batch_config.yaml"
else
    echo "✓ 配置文件已存在: ./batch_config.yaml"
fi
echo

echo "3. 使用配置文件运行（限制 5 个实例）"
echo "------------------------------------"

# 创建临时配置文件，限制实例数量
cat > temp_config.yaml << EOF
agent:
  model:
    name: "deepseek/deepseek-chat-v3-0324:free"
    temperature: 0.1
    max_tokens: 4000
  max_turns: 20
  timeout: 300

instances:
  type: "swe_bench"
  subset: "lite"
  split: "dev"
  instance_limit: 5

num_workers: 2
output_path: "./config_test_results"
enable_logging: true
EOF

$ALEX_BIN run-batch --config temp_config.yaml --quiet

echo "✓ 配置文件测试完成，结果保存在 ./config_test_results/"
echo

echo "4. 演示不同参数组合"
echo "--------------------"

echo "4a. 使用不同模型（如果有 API 密钥）"
echo "注意：此示例需要 OpenAI API 密钥"
# $ALEX_BIN run-batch \
#     --model "openai/gpt-4o-mini" \
#     --dataset.subset lite \
#     --instance-limit 1 \
#     --output ./openai_test_results \
#     --quiet

echo "4b. 并行处理示例"
$ALEX_BIN run-batch \
    --dataset.subset lite \
    --dataset.split dev \
    --instance-limit 6 \
    --workers 3 \
    --output ./parallel_test_results \
    --quiet

echo "✓ 并行处理示例完成，结果保存在 ./parallel_test_results/"
echo

echo "5. 结果分析示例"
echo "----------------"

if [ -f "./parallel_test_results/summary.json" ]; then
    echo "批处理摘要："
    if command -v jq >/dev/null 2>&1; then
        cat ./parallel_test_results/summary.json | jq '{
            总任务数: .total_tasks,
            完成任务数: .completed_tasks,
            失败任务数: .failed_tasks,
            成功率: .success_rate,
            总耗时: .duration,
            平均耗时: .avg_duration,
            总成本: .total_cost
        }'
    else
        echo "安装 jq 以获得更好的 JSON 格式化显示"
        head -20 ./parallel_test_results/summary.json
    fi
fi
echo

echo "6. 错误处理和重试示例"
echo "----------------------"

echo "6a. 超时处理示例（设置短超时）"
$ALEX_BIN run-batch \
    --dataset.subset lite \
    --instance-limit 2 \
    --timeout 10 \
    --workers 1 \
    --output ./timeout_test_results \
    --quiet

echo "6b. 重试机制示例"
$ALEX_BIN run-batch \
    --dataset.subset lite \
    --instance-limit 2 \
    --max-retries 2 \
    --workers 1 \
    --output ./retry_test_results \
    --quiet

echo "✓ 错误处理示例完成"
echo

echo "7. 高级功能示例"
echo "----------------"

echo "7a. 实例过滤 - 处理特定范围"
$ALEX_BIN run-batch \
    --dataset.subset lite \
    --dataset.split dev \
    --instance-slice "0,3" \
    --workers 1 \
    --output ./slice_test_results \
    --quiet

echo "7b. 随机排序示例"
$ALEX_BIN run-batch \
    --dataset.subset lite \
    --dataset.split dev \
    --instance-limit 3 \
    --shuffle \
    --workers 1 \
    --output ./shuffle_test_results \
    --quiet

echo "✓ 高级功能示例完成"
echo

echo "8. 输出格式验证"
echo "----------------"

RESULT_DIR="./parallel_test_results"
if [ -d "$RESULT_DIR" ]; then
    echo "检查输出文件："
    ls -la $RESULT_DIR/
    
    echo
    echo "验证 SWE-Bench 格式的 preds.json："
    if [ -f "$RESULT_DIR/preds.json" ]; then
        echo "✓ preds.json 存在"
        if command -v jq >/dev/null 2>&1; then
            echo "预测数量: $(cat $RESULT_DIR/preds.json | jq length)"
            echo "示例预测:"
            cat $RESULT_DIR/preds.json | jq '.[0] | {instance_id, status, duration_seconds}' 2>/dev/null || echo "JSON 格式检查需要 jq"
        fi
    else
        echo "✗ preds.json 不存在"
    fi
    
    echo
    echo "验证详细结果文件："
    for file in "batch_results.json" "summary.json" "detailed_results.json" "config.yaml"; do
        if [ -f "$RESULT_DIR/$file" ]; then
            echo "✓ $file 存在"
        else
            echo "✗ $file 不存在"
        fi
    done
fi
echo

echo "9. 清理示例"
echo "------------"

echo "清理所有测试结果（可选）："
echo "rm -rf ./quick_test_results ./config_test_results ./parallel_test_results"
echo "rm -rf ./timeout_test_results ./retry_test_results ./slice_test_results ./shuffle_test_results"
echo "rm -f temp_config.yaml"

read -p "是否清理测试结果？(y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf ./quick_test_results ./config_test_results ./parallel_test_results
    rm -rf ./timeout_test_results ./retry_test_results ./slice_test_results ./shuffle_test_results
    rm -f temp_config.yaml
    echo "✓ 测试结果已清理"
else
    echo "保留测试结果用于进一步分析"
fi

echo
echo "==================================="
echo "示例演示完成！"
echo "==================================="
echo
echo "要了解更多用法，请查看："
echo "- README.md: 完整文档"
echo "- config.example.yaml: 配置示例"
echo "- 运行 '$ALEX_BIN run-batch --help' 查看所有选项"
echo
echo "快速命令参考："
echo "- 快速测试: $ALEX_BIN run-batch --dataset.subset lite --instance-limit 2 --workers 1"
echo "- 完整测试: $ALEX_BIN run-batch --dataset.subset lite --workers 3"
echo "- 使用配置: $ALEX_BIN run-batch --config batch_config.yaml"
echo "- Makefile: make swe-bench-test, make swe-bench-lite"