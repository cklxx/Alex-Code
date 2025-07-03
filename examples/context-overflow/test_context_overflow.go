package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"alex/internal/agent"
	"alex/internal/config"
	"alex/internal/session"
)

func main() {
	fmt.Println("🧠 Alex - 上下文溢出处理演示")
	fmt.Println("===============================")

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
	sess, err := agent.StartSession("overflow-demo-session")
	if err != nil {
		log.Fatalf("Failed to start session: %v", err)
	}

	fmt.Printf("✅ 会话启动成功: %s\n", sess.ID)

	// 创建长消息来快速触发上下文限制
	longMessage := strings.Repeat("这是一个很长的消息内容，用来测试上下文管理系统。", 100)
	
	ctx := context.Background()

	// 添加多条长消息直到触发上下文管理
	for i := 1; i <= 50; i++ {
		message := fmt.Sprintf("消息 %d: %s", i, longMessage)
		
		fmt.Printf("\n📝 添加消息 %d (长度: %d)\n", i, len(message))

		// 添加用户消息
		userMsg := &session.Message{
			Role:    "user",
			Content: message,
			Metadata: map[string]interface{}{
				"timestamp": time.Now().Unix(),
				"length":    len(message),
			},
			Timestamp: time.Now(),
		}
		sess.AddMessage(userMsg)

		// 添加AI回复
		aiResponse := fmt.Sprintf("回复消息 %d: 我理解了您的长消息内容，这涉及到复杂的处理逻辑。%s", i, strings.Repeat("详细分析...", 50))
		assistantMsg := &session.Message{
			Role:    "assistant",
			Content: aiResponse,
			Metadata: map[string]interface{}{
				"timestamp": time.Now().Unix(),
				"length":    len(aiResponse),
			},
			Timestamp: time.Now(),
		}
		sess.AddMessage(assistantMsg)

		// 检查当前上下文状态
		if agent.GetReactCore() != nil {
			stats := agent.GetReactCore().GetContextStats(sess)
			fmt.Printf("📊 当前状态: %d 消息, %d tokens (%.1f%% 使用率)\n", 
				stats.TotalMessages, 
				stats.EstimatedTokens,
				float64(stats.EstimatedTokens)/float64(stats.MaxTokens)*100)

			// 当上下文超过阈值时触发处理
			if stats.EstimatedTokens > 6000 {
				fmt.Printf("\n⚠️ 上下文即将溢出，开始处理...\n")
				
				result, err := agent.GetReactCore().ForceContextSummarization(ctx, sess)
				if err != nil {
					fmt.Printf("❌ 上下文总结失败: %v\n", err)
				} else {
					fmt.Printf("✅ 上下文已成功总结:\n")
					fmt.Printf("   • 原始消息: %d\n", result.OriginalCount)
					fmt.Printf("   • 处理后消息: %d\n", result.ProcessedCount)
					fmt.Printf("   • 备份ID: %s\n", result.BackupID)
					
					if result.Summary != nil {
						fmt.Printf("   • 总结要点: %d 个\n", len(result.Summary.KeyPoints))
						fmt.Printf("   • 讨论主题: %d 个\n", len(result.Summary.Topics))
					}
				}
				
				// 显示处理后的新状态
				newStats := agent.GetReactCore().GetContextStats(sess)
				fmt.Printf("📊 处理后状态: %d 消息, %d tokens (%.1f%% 使用率)\n", 
					newStats.TotalMessages, 
					newStats.EstimatedTokens,
					float64(newStats.EstimatedTokens)/float64(newStats.MaxTokens)*100)
				
				fmt.Printf("🔍 消息分布:\n")
				fmt.Printf("   • 系统消息: %d\n", newStats.SystemMessages)
				fmt.Printf("   • 用户消息: %d\n", newStats.UserMessages)
				fmt.Printf("   • 助手消息: %d\n", newStats.AssistantMessages) 
				fmt.Printf("   • 总结消息: %d\n", newStats.SummaryMessages)
				
				break
			}
		}

		// 每10条消息暂停一下
		if i%10 == 0 {
			fmt.Printf("⏸️ 暂停检查...\n")
			time.Sleep(500 * time.Millisecond)
		}
	}

	// 保存会话
	err = agent.GetSessionManager().SaveSession(sess)
	if err != nil {
		fmt.Printf("⚠️ 保存会话失败: %v\n", err)
	} else {
		fmt.Printf("💾 会话已保存\n")
	}

	fmt.Println("\n🎉 上下文溢出处理演示完成！")
	fmt.Println("演示的核心功能:")
	fmt.Println("✅ 实时监控上下文长度")
	fmt.Println("✅ 自动检测溢出阈值")
	fmt.Println("✅ 智能总结压缩机制")
	fmt.Println("✅ 完整历史备份保护")
	fmt.Println("✅ 无缝对话连续性保持")
}