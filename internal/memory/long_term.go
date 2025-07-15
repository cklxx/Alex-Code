package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// LongTermMemoryManager manages persistent, cross-session memories
type LongTermMemoryManager struct {
	storageDir  string
	memories    map[string]*MemoryItem           // ID -> MemoryItem
	categoryIdx map[MemoryCategory][]*MemoryItem // Category -> Items
	tagIdx      map[string][]*MemoryItem         // Tag -> Items
	mutex       sync.RWMutex
}

// NewLongTermMemoryManager creates a new long-term memory manager
func NewLongTermMemoryManager(storageDir string) (*LongTermMemoryManager, error) {
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	ltm := &LongTermMemoryManager{
		storageDir:  storageDir,
		memories:    make(map[string]*MemoryItem),
		categoryIdx: make(map[MemoryCategory][]*MemoryItem),
		tagIdx:      make(map[string][]*MemoryItem),
	}

	// Load existing memories from disk
	if err := ltm.loadFromDisk(); err != nil {
		return nil, fmt.Errorf("failed to load existing memories: %w", err)
	}

	return ltm, nil
}

// Store stores a memory item in long-term memory
func (ltm *LongTermMemoryManager) Store(item *MemoryItem) error {
	ltm.mutex.Lock()
	defer ltm.mutex.Unlock()

	// Set memory type
	item.Type = LongTermMemory

	// Update timestamps
	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	item.LastAccess = now

	// Store in memory
	ltm.memories[item.ID] = item

	// Update indexes
	ltm.updateIndexes(item)

	// Persist to disk
	return ltm.saveToDisk(item)
}

// Recall retrieves memories based on query
func (ltm *LongTermMemoryManager) Recall(query *MemoryQuery) *RecallResult {
	start := time.Now()
	ltm.mutex.RLock()
	defer ltm.mutex.RUnlock()

	var candidates []*MemoryItem

	// Use indexes for efficient retrieval
	if len(query.Categories) > 0 {
		categorySet := make(map[string]bool)
		for _, category := range query.Categories {
			if items, exists := ltm.categoryIdx[category]; exists {
				for _, item := range items {
					if !categorySet[item.ID] {
						candidates = append(candidates, item)
						categorySet[item.ID] = true
					}
				}
			}
		}
	} else if len(query.Tags) > 0 {
		tagSet := make(map[string]bool)
		for _, tag := range query.Tags {
			if items, exists := ltm.tagIdx[tag]; exists {
				for _, item := range items {
					if !tagSet[item.ID] {
						candidates = append(candidates, item)
						tagSet[item.ID] = true
					}
				}
			}
		}
	} else {
		// Full scan
		for _, item := range ltm.memories {
			candidates = append(candidates, item)
		}
	}

	// Apply additional filters and content search
	var filtered []*MemoryItem
	for _, item := range candidates {
		if ltm.matchesQuery(item, query) {
			// Update access info
			item.AccessCount++
			item.LastAccess = time.Now()
			filtered = append(filtered, item)
		}
	}

	// Calculate relevance scores
	scores := make([]float64, len(filtered))
	for i, item := range filtered {
		scores[i] = ltm.calculateRelevance(item, query)
	}

	// Sort by relevance
	ltm.sortByRelevance(filtered, scores, query.SortBy)

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
func (ltm *LongTermMemoryManager) Update(id string, updater func(*MemoryItem) error) error {
	ltm.mutex.Lock()
	defer ltm.mutex.Unlock()

	item, exists := ltm.memories[id]
	if !exists {
		return fmt.Errorf("memory item %s not found", id)
	}

	// Remove from old indexes
	ltm.removeFromIndexes(item)

	if err := updater(item); err != nil {
		// Restore indexes on error
		ltm.updateIndexes(item)
		return fmt.Errorf("failed to update memory item: %w", err)
	}

	item.UpdatedAt = time.Now()

	// Update indexes with new data
	ltm.updateIndexes(item)

	// Persist changes
	return ltm.saveToDisk(item)
}

// Delete removes a memory item
func (ltm *LongTermMemoryManager) Delete(id string) error {
	ltm.mutex.Lock()
	defer ltm.mutex.Unlock()

	item, exists := ltm.memories[id]
	if !exists {
		return fmt.Errorf("memory item %s not found", id)
	}

	// Remove from memory
	delete(ltm.memories, id)

	// Remove from indexes
	ltm.removeFromIndexes(item)

	// Remove from disk
	return ltm.deleteFromDisk(id)
}

// GetByCategory retrieves all memories in a category
func (ltm *LongTermMemoryManager) GetByCategory(category MemoryCategory) []*MemoryItem {
	ltm.mutex.RLock()
	defer ltm.mutex.RUnlock()

	if items, exists := ltm.categoryIdx[category]; exists {
		result := make([]*MemoryItem, len(items))
		copy(result, items)
		return result
	}

	return []*MemoryItem{}
}

// GetByTags retrieves all memories with specified tags
func (ltm *LongTermMemoryManager) GetByTags(tags []string) []*MemoryItem {
	ltm.mutex.RLock()
	defer ltm.mutex.RUnlock()

	itemSet := make(map[string]*MemoryItem)

	for _, tag := range tags {
		if items, exists := ltm.tagIdx[tag]; exists {
			for _, item := range items {
				itemSet[item.ID] = item
			}
		}
	}

	var result []*MemoryItem
	for _, item := range itemSet {
		result = append(result, item)
	}

	return result
}

// GetStats returns statistics about long-term memory
func (ltm *LongTermMemoryManager) GetStats() *MemoryStats {
	ltm.mutex.RLock()
	defer ltm.mutex.RUnlock()

	stats := &MemoryStats{
		TotalItems:      len(ltm.memories),
		ItemsByType:     make(map[MemoryType]int),
		ItemsByCategory: make(map[MemoryCategory]int),
	}

	for _, item := range ltm.memories {
		stats.ItemsByType[item.Type]++
		stats.ItemsByCategory[item.Category]++
		stats.TotalSize += int64(len(item.Content))
	}

	return stats
}

// Vacuum removes old, low-importance memories to free space
func (ltm *LongTermMemoryManager) Vacuum(maxItems int, minImportance float64) error {
	ltm.mutex.Lock()
	defer ltm.mutex.Unlock()

	if len(ltm.memories) <= maxItems {
		return nil
	}

	// Collect items with their scores
	type itemScore struct {
		item  *MemoryItem
		score float64
	}

	var itemScores []itemScore
	for _, item := range ltm.memories {
		score := ltm.calculateRetentionScore(item)
		if score >= minImportance {
			itemScores = append(itemScores, itemScore{item: item, score: score})
		}
	}

	// Sort by retention score (highest first)
	sort.Slice(itemScores, func(i, j int) bool {
		return itemScores[i].score > itemScores[j].score
	})

	// Keep only the top items
	toKeep := maxItems
	if len(itemScores) < maxItems {
		toKeep = len(itemScores)
	}

	keptItems := make(map[string]*MemoryItem)
	for i := 0; i < toKeep; i++ {
		keptItems[itemScores[i].item.ID] = itemScores[i].item
	}

	// Remove items not in the keep list
	for id, item := range ltm.memories {
		if _, keep := keptItems[id]; !keep {
			ltm.removeFromIndexes(item)
			_ = ltm.deleteFromDisk(id)
		}
	}

	// Update memories map
	ltm.memories = keptItems

	return nil
}

// Private helper methods

func (ltm *LongTermMemoryManager) loadFromDisk() error {
	files, err := os.ReadDir(ltm.storageDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			item, err := ltm.loadItemFromDisk(file.Name()[:len(file.Name())-5])
			if err != nil {
				continue // Skip corrupted files
			}
			ltm.memories[item.ID] = item
			ltm.updateIndexes(item)
		}
	}

	return nil
}

func (ltm *LongTermMemoryManager) loadItemFromDisk(id string) (*MemoryItem, error) {
	filePath := filepath.Join(ltm.storageDir, id+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var item MemoryItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}

	return &item, nil
}

func (ltm *LongTermMemoryManager) saveToDisk(item *MemoryItem) error {
	filePath := filepath.Join(ltm.storageDir, item.ID+".json")
	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func (ltm *LongTermMemoryManager) deleteFromDisk(id string) error {
	filePath := filepath.Join(ltm.storageDir, id+".json")
	return os.Remove(filePath)
}

func (ltm *LongTermMemoryManager) updateIndexes(item *MemoryItem) {
	// Update category index
	ltm.categoryIdx[item.Category] = append(ltm.categoryIdx[item.Category], item)

	// Update tag index
	for _, tag := range item.Tags {
		ltm.tagIdx[tag] = append(ltm.tagIdx[tag], item)
	}
}

func (ltm *LongTermMemoryManager) removeFromIndexes(item *MemoryItem) {
	// Remove from category index
	if items, exists := ltm.categoryIdx[item.Category]; exists {
		for i, indexItem := range items {
			if indexItem.ID == item.ID {
				ltm.categoryIdx[item.Category] = append(items[:i], items[i+1:]...)
				break
			}
		}
	}

	// Remove from tag index
	for _, tag := range item.Tags {
		if items, exists := ltm.tagIdx[tag]; exists {
			for i, indexItem := range items {
				if indexItem.ID == item.ID {
					ltm.tagIdx[tag] = append(items[:i], items[i+1:]...)
					break
				}
			}
		}
	}
}

func (ltm *LongTermMemoryManager) matchesQuery(item *MemoryItem, query *MemoryQuery) bool {
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

	// Check importance
	if query.MinImportance > 0 && item.Importance < query.MinImportance {
		return false
	}

	// Check content search
	if query.Content != "" {
		content := strings.ToLower(item.Content)
		searchTerm := strings.ToLower(query.Content)
		if !strings.Contains(content, searchTerm) {
			return false
		}
	}

	return true
}

func (ltm *LongTermMemoryManager) calculateRelevance(item *MemoryItem, query *MemoryQuery) float64 {
	score := item.Importance

	// Content relevance boost
	if query.Content != "" {
		content := strings.ToLower(item.Content)
		searchTerm := strings.ToLower(query.Content)
		if strings.Contains(content, searchTerm) {
			score += 0.2
		}
	}

	// Access frequency boost
	accessBoost := float64(item.AccessCount) * 0.05
	if accessBoost > 0.3 {
		accessBoost = 0.3
	}
	score += accessBoost

	// Recent update boost
	daysSinceUpdate := time.Since(item.UpdatedAt).Hours() / 24
	if daysSinceUpdate < 7 {
		score += 0.1
	}

	// Ensure score is between 0 and 1
	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (ltm *LongTermMemoryManager) calculateRetentionScore(item *MemoryItem) float64 {
	score := item.Importance

	// Access frequency factor
	accessFactor := float64(item.AccessCount) * 0.1
	if accessFactor > 0.5 {
		accessFactor = 0.5
	}
	score += accessFactor

	// Recency factor (penalize very old items)
	daysSinceAccess := time.Since(item.LastAccess).Hours() / 24
	if daysSinceAccess > 365 {
		score -= 0.3
	} else if daysSinceAccess > 90 {
		score -= 0.1
	}

	// Ensure score is not negative
	if score < 0 {
		score = 0
	}

	return score
}

func (ltm *LongTermMemoryManager) sortByRelevance(items []*MemoryItem, scores []float64, sortBy string) {
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
