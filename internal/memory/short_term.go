package memory

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// ShortTermMemoryManager manages temporary, session-specific memories
type ShortTermMemoryManager struct {
	memories    map[string]*MemoryItem   // ID -> MemoryItem
	sessionData map[string][]*MemoryItem // SessionID -> Items
	maxItems    int
	ttl         time.Duration
	mutex       sync.RWMutex
}

// NewShortTermMemoryManager creates a new short-term memory manager
func NewShortTermMemoryManager(maxItems int, ttl time.Duration) *ShortTermMemoryManager {
	return &ShortTermMemoryManager{
		memories:    make(map[string]*MemoryItem),
		sessionData: make(map[string][]*MemoryItem),
		maxItems:    maxItems,
		ttl:         ttl,
	}
}

// Store stores a memory item in short-term memory
func (stm *ShortTermMemoryManager) Store(item *MemoryItem) error {
	stm.mutex.Lock()
	defer stm.mutex.Unlock()

	// Set memory type and expiration
	item.Type = ShortTermMemory
	if item.ExpiresAt == nil {
		expiresAt := time.Now().Add(stm.ttl)
		item.ExpiresAt = &expiresAt
	}

	// Update timestamps
	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	item.LastAccess = now

	// Store in main map
	stm.memories[item.ID] = item

	// Add to session index
	if item.SessionID != "" {
		stm.sessionData[item.SessionID] = append(stm.sessionData[item.SessionID], item)
	}

	// Cleanup if needed
	stm.cleanup()

	return nil
}

// Recall retrieves memories based on query
func (stm *ShortTermMemoryManager) Recall(query *MemoryQuery) *RecallResult {
	start := time.Now()
	stm.mutex.RLock()
	defer stm.mutex.RUnlock()

	var candidates []*MemoryItem

	// Filter by session if specified
	if query.SessionID != "" {
		if items, exists := stm.sessionData[query.SessionID]; exists {
			candidates = items
		}
	} else {
		// Collect all items
		for _, item := range stm.memories {
			candidates = append(candidates, item)
		}
	}

	// Apply filters
	var filtered []*MemoryItem
	for _, item := range candidates {
		if stm.matchesQuery(item, query) {
			// Update access info
			item.AccessCount++
			item.LastAccess = time.Now()
			filtered = append(filtered, item)
		}
	}

	// Calculate relevance scores
	scores := make([]float64, len(filtered))
	for i, item := range filtered {
		scores[i] = stm.calculateRelevance(item, query)
	}

	// Sort by relevance
	stm.sortByRelevance(filtered, scores, query.SortBy)

	// Apply limit
	if query.Limit > 0 && len(filtered) > query.Limit {
		filtered = filtered[:query.Limit]
		scores = scores[:query.Limit]
	}

	return &RecallResult{
		Items:           filtered,
		TotalFound:      len(filtered),
		RelevanceScores: scores,
		ProcessingTime:  time.Since(start),
	}
}

// Update updates an existing memory item
func (stm *ShortTermMemoryManager) Update(id string, updater func(*MemoryItem) error) error {
	stm.mutex.Lock()
	defer stm.mutex.Unlock()

	item, exists := stm.memories[id]
	if !exists {
		return fmt.Errorf("memory item %s not found", id)
	}

	if err := updater(item); err != nil {
		return fmt.Errorf("failed to update memory item: %w", err)
	}

	item.UpdatedAt = time.Now()
	return nil
}

// Delete removes a memory item
func (stm *ShortTermMemoryManager) Delete(id string) error {
	stm.mutex.Lock()
	defer stm.mutex.Unlock()

	item, exists := stm.memories[id]
	if !exists {
		return fmt.Errorf("memory item %s not found", id)
	}

	// Remove from main map
	delete(stm.memories, id)

	// Remove from session index
	if item.SessionID != "" {
		if items, exists := stm.sessionData[item.SessionID]; exists {
			for i, sessionItem := range items {
				if sessionItem.ID == id {
					stm.sessionData[item.SessionID] = append(items[:i], items[i+1:]...)
					break
				}
			}
		}
	}

	return nil
}

// GetSessionMemories retrieves all memories for a session
func (stm *ShortTermMemoryManager) GetSessionMemories(sessionID string) []*MemoryItem {
	stm.mutex.RLock()
	defer stm.mutex.RUnlock()

	if items, exists := stm.sessionData[sessionID]; exists {
		// Return copy to prevent external modification
		result := make([]*MemoryItem, len(items))
		copy(result, items)
		return result
	}

	return []*MemoryItem{}
}

// ClearSession removes all memories for a session
func (stm *ShortTermMemoryManager) ClearSession(sessionID string) error {
	stm.mutex.Lock()
	defer stm.mutex.Unlock()

	if items, exists := stm.sessionData[sessionID]; exists {
		for _, item := range items {
			delete(stm.memories, item.ID)
		}
		delete(stm.sessionData, sessionID)
	}

	return nil
}

// GetStats returns statistics about short-term memory
func (stm *ShortTermMemoryManager) GetStats() *MemoryStats {
	stm.mutex.RLock()
	defer stm.mutex.RUnlock()

	stats := &MemoryStats{
		TotalItems:      len(stm.memories),
		ItemsByType:     make(map[MemoryType]int),
		ItemsByCategory: make(map[MemoryCategory]int),
	}

	for _, item := range stm.memories {
		stats.ItemsByType[item.Type]++
		stats.ItemsByCategory[item.Category]++
		stats.TotalSize += int64(len(item.Content))
	}

	return stats
}

// Private helper methods

func (stm *ShortTermMemoryManager) cleanup() {
	now := time.Now()
	var toDelete []string

	// Remove expired items
	for id, item := range stm.memories {
		if item.ExpiresAt != nil && now.After(*item.ExpiresAt) {
			toDelete = append(toDelete, id)
		}
	}

	// Remove expired items
	for _, id := range toDelete {
		_ = stm.Delete(id)
	}

	// Enforce max items limit using LRU strategy
	if len(stm.memories) > stm.maxItems {
		// Collect all items with their last access time
		type itemWithTime struct {
			item       *MemoryItem
			lastAccess time.Time
		}

		var itemsWithTime []itemWithTime
		for _, item := range stm.memories {
			itemsWithTime = append(itemsWithTime, itemWithTime{
				item:       item,
				lastAccess: item.LastAccess,
			})
		}

		// Sort by last access time (oldest first)
		sort.Slice(itemsWithTime, func(i, j int) bool {
			return itemsWithTime[i].lastAccess.Before(itemsWithTime[j].lastAccess)
		})

		// Remove oldest items
		excess := len(stm.memories) - stm.maxItems
		for i := 0; i < excess; i++ {
			_ = stm.Delete(itemsWithTime[i].item.ID)
		}
	}
}

func (stm *ShortTermMemoryManager) matchesQuery(item *MemoryItem, query *MemoryQuery) bool {
	// Check types
	if len(query.Types) > 0 {
		found := false
		for _, t := range query.Types {
			if item.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check categories
	if len(query.Categories) > 0 {
		found := false
		for _, c := range query.Categories {
			if item.Category == c {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check importance
	if query.MinImportance > 0 && item.Importance < query.MinImportance {
		return false
	}

	// Check tags
	if len(query.Tags) > 0 {
		for _, queryTag := range query.Tags {
			found := false
			for _, itemTag := range item.Tags {
				if itemTag == queryTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

func (stm *ShortTermMemoryManager) calculateRelevance(item *MemoryItem, query *MemoryQuery) float64 {
	score := item.Importance

	// Boost score based on access frequency
	accessBoost := float64(item.AccessCount) * 0.1
	if accessBoost > 0.5 {
		accessBoost = 0.5
	}
	score += accessBoost

	// Boost score based on recency
	recencyBoost := 0.0
	if !item.LastAccess.IsZero() {
		hoursSinceAccess := time.Since(item.LastAccess).Hours()
		if hoursSinceAccess < 1 {
			recencyBoost = 0.3
		} else if hoursSinceAccess < 24 {
			recencyBoost = 0.2
		} else if hoursSinceAccess < 168 {
			recencyBoost = 0.1
		}
	}
	score += recencyBoost

	// Ensure score is between 0 and 1
	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (stm *ShortTermMemoryManager) sortByRelevance(items []*MemoryItem, scores []float64, sortBy string) {
	switch sortBy {
	case "importance":
		sort.Slice(items, func(i, j int) bool {
			return items[i].Importance > items[j].Importance
		})
	case "recency":
		sort.Slice(items, func(i, j int) bool {
			return items[i].LastAccess.After(items[j].LastAccess)
		})
	case "access_count":
		sort.Slice(items, func(i, j int) bool {
			return items[i].AccessCount > items[j].AccessCount
		})
	default:
		// Sort by relevance score
		sort.Slice(items, func(i, j int) bool {
			return scores[i] > scores[j]
		})
	}
}
