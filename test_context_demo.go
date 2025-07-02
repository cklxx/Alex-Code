package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"alex/internal/agent"
	"alex/internal/config"
	"alex/internal/session"
)

func main() {
	fmt.Println("🧠 Alex - 上下文管理系统演示")
	fmt.Println("==============================")

	// 创建配置管理器
	configMgr, err := config.NewManager()
	if err != nil {
		log.Fatalf("Failed to create config manager: %v", err)
	}

	// 创建ReactAgent
	agent, err := agent.NewReactAgent(configMgr)
	if err != nil {
		log.Fatalf("Failed to create ReactAgent: %v", err)
	}

	// 启动新会话
	sess, err := agent.StartSession("context-demo-session")
	if err != nil {
		log.Fatalf("Failed to start session: %v", err)
	}

	fmt.Printf("✅ 会话启动成功: %s\n", sess.ID)

	// 模拟长对话，添加多条消息来测试上下文管理
	messages := []string{
		"你好，我想了解Go语言的并发编程",
		"请详细解释goroutine的工作原理",
		"channel的使用场景有哪些？",
		"如何避免goroutine泄漏？",
		"sync包中有哪些常用的同步原语？",
		"context包的使用方法是什么？",
		"WaitGroup的使用注意事项有哪些？",
		"Mutex和RWMutex的区别是什么？",
		"如何进行性能调优？",
		"Go的内存模型是怎样的？",
		"请帮我分析这个复杂的并发问题，涉及多个goroutine之间的协调",
	}

	ctx := context.Background()

	for i, msg := range messages {
		fmt.Printf("\n📝 消息 %d: %s\n", i+1, msg)

		// 获取当前会话的上下文统计
		if agent.GetReactCore() != nil {
			stats := agent.GetReactCore().GetContextStats(sess)
			fmt.Printf("📊 当前上下文: %d 消息, 约 %d tokens\n", 
				stats.TotalMessages, stats.EstimatedTokens)

			// 如果上下文接近限制，显示警告
			if stats.EstimatedTokens > 6000 {
				fmt.Printf("⚠️ 上下文接近限制 (%d/%d tokens)\n", 
					stats.EstimatedTokens, stats.MaxTokens)
			}
		}

		// 模拟处理消息 (这里简化为添加到会话)
		userMsg := &session.Message{
			Role:    "user",
			Content: msg,
			Metadata: map[string]interface{}{
				"timestamp": time.Now().Unix(),
			},
			Timestamp: time.Now(),
		}
		sess.AddMessage(userMsg)

		// 模拟AI回复
		aiResponse := fmt.Sprintf("关于 \"%s\" 的详细回答...", msg)
		assistantMsg := &session.Message{
			Role:    "assistant",
			Content: aiResponse,
			Metadata: map[string]interface{}{
				"timestamp": time.Now().Unix(),
			},
			Timestamp: time.Now(),
		}
		sess.AddMessage(assistantMsg)

		// 每隔几条消息检查是否需要进行上下文管理
		if i%3 == 2 && agent.GetReactCore() != nil {
			stats := agent.GetReactCore().GetContextStats(sess)
			if stats.EstimatedTokens > 4000 {
				fmt.Printf("🔄 执行上下文管理检查...\n")
				
				result, err := agent.GetReactCore().ForceContextSummarization(ctx, sess)
				if err != nil {
					fmt.Printf("❌ 上下文总结失败: %v\n", err)
				} else {
					fmt.Printf("✅ 上下文已总结: %d → %d 消息 (备份: %s)\n", 
						result.OriginalCount, result.ProcessedCount, result.BackupID)
				}
			}
		}

		// 短暂延迟模拟真实对话
		time.Sleep(100 * time.Millisecond)
	}

	// 最终统计
	fmt.Printf("\n📈 最终统计:\n")
	if agent.GetReactCore() != nil {
		finalStats := agent.GetReactCore().GetContextStats(sess)
		fmt.Printf("• 总消息数: %d\n", finalStats.TotalMessages)
		fmt.Printf("• 系统消息: %d\n", finalStats.SystemMessages)
		fmt.Printf("• 用户消息: %d\n", finalStats.UserMessages)
		fmt.Printf("• 助手消息: %d\n", finalStats.AssistantMessages)
		fmt.Printf("• 总结消息: %d\n", finalStats.SummaryMessages)
		fmt.Printf("• 估计tokens: %d\n", finalStats.EstimatedTokens)
	}

	// 保存会话
	err = agent.GetSessionManager().SaveSession(sess)
	if err != nil {
		fmt.Printf("⚠️ 保存会话失败: %v\n", err)
	} else {
		fmt.Printf("💾 会话已保存\n")
	}

	fmt.Println("\n🎉 上下文管理演示完成！")
	fmt.Println("功能亮点:")
	fmt.Println("✅ 智能上下文长度检测")
	fmt.Println("✅ 自动消息总结和压缩")
	fmt.Println("✅ 完整对话历史备份")
	fmt.Println("✅ 无缝会话连续性")
	fmt.Println("✅ 实时上下文统计")
}