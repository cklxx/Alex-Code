package llm

import (
	"fmt"
	"log"
	"strings"
)

// CacheStatsDisplay displays cache statistics in a user-friendly format
func (cm *CacheManager) DisplayCacheStats() {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 SESSION CACHE STATISTICS")
	fmt.Println(strings.Repeat("=", 60))

	if len(cm.caches) == 0 {
		fmt.Println("No active sessions in cache")
		return
	}

	totalMessages := 0
	totalTokens := 0
	totalRequests := 0

	fmt.Printf("🔢 Total Sessions in Cache: %d\n", len(cm.caches))
	fmt.Println(strings.Repeat("-", 60))

	for sessionID, cache := range cm.caches {
		cache.LastUsed.Format("15:04:05")
		totalMessages += len(cache.Messages)
		totalTokens += cache.TokensUsed
		totalRequests += cache.RequestCount

		fmt.Printf("📝 Session: %s\n", sessionID)
		fmt.Printf("   └─ Messages: %d | Tokens: %d | Requests: %d\n",
			len(cache.Messages), cache.TokensUsed, cache.RequestCount)
		fmt.Printf("   └─ Last Used: %s | Cache Key: %s\n",
			cache.LastUsed.Format("15:04:05"), cache.CacheKey[:8]+"...")

		// Show message optimization potential
		if len(cache.Messages) > 5 {
			optimizedCount := 1 + 3 + 1 // summary + recent + new
			savedMessages := len(cache.Messages) - optimizedCount
			fmt.Printf("   └─ 🚀 Optimization: %d messages → %d messages (saved: %d)\n",
				len(cache.Messages), optimizedCount, savedMessages)
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("📈 TOTALS:\n")
	fmt.Printf("   • Total Cached Messages: %d\n", totalMessages)
	fmt.Printf("   • Total Tokens Processed: %d\n", totalTokens)
	fmt.Printf("   • Total API Requests: %d\n", totalRequests)

	// Calculate potential savings
	estimatedSavings := 0
	for _, cache := range cm.caches {
		if len(cache.Messages) > 5 {
			estimatedSavings += (len(cache.Messages) - 5) * cache.RequestCount
		}
	}

	if estimatedSavings > 0 {
		fmt.Printf("   • 💰 Estimated Messages Saved: %d\n", estimatedSavings)
		fmt.Printf("   • 📊 Cache Efficiency: %.1f%%\n",
			float64(estimatedSavings)/float64(totalMessages)*100)
	}

	fmt.Println(strings.Repeat("=", 60))
}

// ShowSessionDetails shows detailed information about a specific session's cache
func (cm *CacheManager) ShowSessionDetails(sessionID string) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	cache, exists := cm.caches[sessionID]
	if !exists {
		fmt.Printf("❌ Session '%s' not found in cache\n", sessionID)
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("🔍 SESSION DETAILS: %s\n", sessionID)
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("📊 Basic Info:\n")
	fmt.Printf("   • Messages in Cache: %d\n", len(cache.Messages))
	fmt.Printf("   • Total Tokens Used: %d\n", cache.TokensUsed)
	fmt.Printf("   • API Requests Made: %d\n", cache.RequestCount)
	fmt.Printf("   • Last Used: %s\n", cache.LastUsed.Format("2006-01-02 15:04:05"))
	fmt.Printf("   • Cache Key: %s\n", cache.CacheKey)
	fmt.Println()

	fmt.Printf("💬 Message History:\n")
	for i, msg := range cache.Messages {
		role := msg.Role
		content := msg.Content
		if len(content) > 80 {
			content = content[:80] + "..."
		}

		roleIcon := "👤"
		if role == "assistant" {
			roleIcon = "🤖"
		} else if role == "system" {
			roleIcon = "⚙️"
		}

		fmt.Printf("   %d. %s %s: %s\n", i+1, roleIcon, role, content)

		if len(msg.ToolCalls) > 0 {
			fmt.Printf("      └─ 🔧 Tool Calls: %d\n", len(msg.ToolCalls))
		}
	}

	// Show optimization preview
	if len(cache.Messages) > 5 {
		fmt.Println()
		fmt.Printf("🚀 Optimization Preview:\n")
		optimized := cm.GetOptimizedMessages(sessionID, []Message{
			{Role: "user", Content: "[NEW MESSAGE]"},
		})

		fmt.Printf("   • Original Messages: %d\n", len(cache.Messages))
		fmt.Printf("   • Optimized Messages: %d\n", len(optimized))
		fmt.Printf("   • Messages Saved: %d (%.1f%%)\n",
			len(cache.Messages)+1-len(optimized),
			float64(len(cache.Messages)+1-len(optimized))/float64(len(cache.Messages)+1)*100)

		fmt.Printf("   • Optimized Structure:\n")
		for i, msg := range optimized {
			role := msg.Role
			content := msg.Content
			if len(content) > 60 {
				content = content[:60] + "..."
			}

			roleIcon := "👤"
			if role == "assistant" {
				roleIcon = "🤖"
			} else if role == "system" {
				roleIcon = "⚙️"
			}

			fmt.Printf("      %d. %s %s: %s\n", i+1, roleIcon, role, content)
		}
	}

	fmt.Println(strings.Repeat("=", 80))
}

// DemoMessageOptimization demonstrates the cache optimization in action
func DemoMessageOptimization() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🎯 SESSION CACHE OPTIMIZATION DEMO")
	fmt.Println(strings.Repeat("=", 70))

	cm := NewCacheManager()
	sessionID := "demo_session_12345"

	// Simulate a conversation with many messages
	fmt.Println("📝 Simulating a conversation with 15 messages...")

	cache := cm.GetOrCreateCache(sessionID)

	// Add multiple conversation rounds
	messages := []Message{}
	for i := 1; i <= 7; i++ {
		// User message
		userMsg := Message{
			Role:    "user",
			Content: fmt.Sprintf("User question #%d: Can you help me with task %d?", i, i),
		}
		messages = append(messages, userMsg)

		// Assistant response
		assistantMsg := Message{
			Role:    "assistant",
			Content: fmt.Sprintf("Assistant response #%d: Sure! I'll help you with task %d. Here's what I found...", i, i),
		}
		messages = append(messages, assistantMsg)

		// Update cache in batches
		if i%2 == 0 {
			cm.UpdateCache(sessionID, messages, i*25)
			messages = []Message{} // Reset for next batch
		}
	}

	// Add remaining messages
	if len(messages) > 0 {
		cm.UpdateCache(sessionID, messages, 50)
	}

	fmt.Printf("✅ Created session with %d messages\n\n", len(cache.Messages))

	// Show the cache state
	cm.ShowSessionDetails(sessionID)

	// Demonstrate optimization with a new message
	fmt.Println("\n🔄 Now adding a new message to see optimization...")
	newMessage := []Message{
		{Role: "user", Content: "What's the status of all my previous tasks?"},
	}

	originalCount := len(cache.Messages) + len(newMessage)
	optimized := cm.GetOptimizedMessages(sessionID, newMessage)

	fmt.Printf("\n📈 OPTIMIZATION RESULTS:\n")
	fmt.Printf("   • Without Cache: %d messages would be sent to API\n", originalCount)
	fmt.Printf("   • With Cache: %d messages actually sent to API\n", len(optimized))
	fmt.Printf("   • Messages Saved: %d (%.1f%% reduction)\n",
		originalCount-len(optimized),
		float64(originalCount-len(optimized))/float64(originalCount)*100)

	fmt.Printf("\n📋 Optimized message structure:\n")
	for i, msg := range optimized {
		role := msg.Role
		content := msg.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}

		roleIcon := "👤"
		if role == "assistant" {
			roleIcon = "🤖"
		} else if role == "system" {
			roleIcon = "⚙️"
		}

		fmt.Printf("   %d. %s %s: %s\n", i+1, roleIcon, role, content)
	}

	fmt.Println(strings.Repeat("=", 70))
}

// GetCacheVisualization returns a visual representation of cache usage
func (cm *CacheManager) GetCacheVisualization() string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if len(cm.caches) == 0 {
		return "📭 No sessions in cache"
	}

	var lines []string
	lines = append(lines, "📊 Cache Visualization:")
	lines = append(lines, "")

	for sessionID, cache := range cm.caches {
		// Create a simple bar chart for message count
		messageCount := len(cache.Messages)
		maxWidth := 40
		barWidth := int(float64(messageCount) / float64(cm.maxMessageCount) * float64(maxWidth))
		if barWidth > maxWidth {
			barWidth = maxWidth
		}

		bar := strings.Repeat("█", barWidth) + strings.Repeat("░", maxWidth-barWidth)

		lines = append(lines, fmt.Sprintf("Session: %s", sessionID))
		lines = append(lines, fmt.Sprintf("Messages [%2d]: |%s| %d/%d",
			messageCount, bar, messageCount, cm.maxMessageCount))
		lines = append(lines, fmt.Sprintf("Requests: %d | Tokens: %d",
			cache.RequestCount, cache.TokensUsed))
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// LogCacheOperation logs cache operations with detailed information
func (cm *CacheManager) LogCacheOperation(operation string, sessionID string, details map[string]interface{}) {
	timestamp := "[" + strings.Replace(fmt.Sprintf("%v", details["timestamp"]), " ", "T", 1) + "]"

	switch operation {
	case "optimization":
		original := details["original_count"].(int)
		optimized := details["optimized_count"].(int)
		saved := original - optimized

		log.Printf("🚀 %s CACHE-OPTIMIZE session=%s: %d→%d messages (saved %d, %.1f%%)",
			timestamp, sessionID, original, optimized, saved,
			float64(saved)/float64(original)*100)

	case "update":
		messages := details["message_count"].(int)
		tokens := details["tokens"].(int)
		requests := details["requests"].(int)

		log.Printf("📝 %s CACHE-UPDATE session=%s: messages=%d tokens=%d requests=%d",
			timestamp, sessionID, messages, tokens, requests)

	case "compression":
		oldCount := details["old_count"].(int)
		newCount := details["new_count"].(int)

		log.Printf("🗜️ %s CACHE-COMPRESS session=%s: %d→%d messages",
			timestamp, sessionID, oldCount, newCount)

	case "cleanup":
		removed := details["removed"].(int)
		remaining := details["remaining"].(int)

		log.Printf("🧹 %s CACHE-CLEANUP: removed=%d remaining=%d sessions",
			timestamp, removed, remaining)
	}
}
