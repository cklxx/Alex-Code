package main

import (
	"fmt"
	"log"
	
	"alex/internal/llm"
)

func main() {
	// Create HTTP client to check cache stats
	client, err := llm.NewHTTPClient()
	if err != nil {
		log.Fatal("Failed to create HTTP client:", err)
	}
	defer client.Close()
	
	// Get Kimi cache statistics
	stats := client.GetKimiCacheStats()
	
	fmt.Println("Kimi Cache Statistics:")
	fmt.Printf("  Total Caches: %v\n", stats["total_caches"])
	fmt.Printf("  Active Caches: %v\n", stats["active_caches"])
	fmt.Printf("  Total Requests: %v\n", stats["total_requests"])
	fmt.Printf("  Cache Provider: %v\n", stats["cache_provider"])
}