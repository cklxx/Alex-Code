package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"alex/internal/session"
)

// ContextPreservationManager handles backup and restoration of conversation context
type ContextPreservationManager struct {
	backupDir string
}

// ContextBackup represents a complete backup of session context
type ContextBackup struct {
	ID            string              `json:"id"`
	SessionID     string              `json:"session_id"`
	Messages      []*session.Message  `json:"messages"`
	CreatedAt     time.Time           `json:"created_at"`
	Metadata      map[string]interface{} `json:"metadata"`
	OriginalCount int                 `json:"original_count"`
}

// NewContextPreservationManager creates a new context preservation manager
func NewContextPreservationManager() *ContextPreservationManager {
	homeDir, _ := os.UserHomeDir()
	backupDir := filepath.Join(homeDir, ".deep-coding-context-backups")
	os.MkdirAll(backupDir, 0755)
	
	return &ContextPreservationManager{
		backupDir: backupDir,
	}
}

// CreateBackup creates a complete backup of the session context
func (cpm *ContextPreservationManager) CreateBackup(sess *session.Session) *ContextBackup {
	messages := sess.GetMessages()
	backupID := fmt.Sprintf("backup_%s_%d", sess.ID, time.Now().UnixNano())
	
	backup := &ContextBackup{
		ID:            backupID,
		SessionID:     sess.ID,
		Messages:      make([]*session.Message, len(messages)),
		CreatedAt:     time.Now(),
		OriginalCount: len(messages),
		Metadata: map[string]interface{}{
			"context":     sess.GetContext(), 
			"working_dir": sess.WorkingDir,
			"config":      sess.Config,
		},
	}
	
	// Deep copy messages
	copy(backup.Messages, messages)
	
	// Save backup to disk
	if err := cpm.saveBackup(backup); err != nil {
		// Log error but don't fail - backup is still in memory
		fmt.Printf("Warning: failed to save context backup to disk: %v\n", err)
	}
	
	return backup
}

// RestoreBackup restores the complete conversation history from a backup
func (cpm *ContextPreservationManager) RestoreBackup(sess *session.Session, backupID string) error {
	backup, err := cpm.loadBackup(backupID)
	if err != nil {
		return fmt.Errorf("failed to load backup %s: %w", backupID, err)
	}
	
	if backup.SessionID != sess.ID {
		return fmt.Errorf("backup %s belongs to session %s, not %s", backupID, backup.SessionID, sess.ID)
	}
	
	// Clear current messages and restore from backup
	sess.ClearMessages()
	for _, msg := range backup.Messages {
		sess.AddMessage(msg)
	}
	
	// Restore context if available
	if context, ok := backup.Metadata["context"].(string); ok && context != "" {
		sess.SetContext(context)
	}
	
	return nil
}

// ListBackups returns all available backups for a session
func (cpm *ContextPreservationManager) ListBackups(sessionID string) ([]*ContextBackup, error) {
	files, err := os.ReadDir(cpm.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}
	
	var backups []*ContextBackup
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			backup, err := cpm.loadBackup(file.Name()[:len(file.Name())-5])
			if err != nil {
				continue // Skip corrupted backups
			}
			
			if backup.SessionID == sessionID {
				backups = append(backups, backup)
			}
		}
	}
	
	return backups, nil
}

// CleanupOldBackups removes backups older than the specified duration
func (cpm *ContextPreservationManager) CleanupOldBackups(maxAge time.Duration) error {
	files, err := os.ReadDir(cpm.backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}
	
	cutoff := time.Now().Add(-maxAge)
	
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			info, err := file.Info()
			if err != nil {
				continue
			}
			
			if info.ModTime().Before(cutoff) {
				filePath := filepath.Join(cpm.backupDir, file.Name())
				if err := os.Remove(filePath); err != nil {
					fmt.Printf("Warning: failed to remove old backup %s: %v\n", file.Name(), err)
				}
			}
		}
	}
	
	return nil
}

// GetBackupStats returns statistics about stored backups
func (cpm *ContextPreservationManager) GetBackupStats() (map[string]interface{}, error) {
	files, err := os.ReadDir(cpm.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}
	
	totalBackups := 0
	totalSize := int64(0)
	oldestBackup := time.Now()
	newestBackup := time.Time{}
	
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			info, err := file.Info()
			if err != nil {
				continue
			}
			
			totalBackups++
			totalSize += info.Size()
			
			if info.ModTime().Before(oldestBackup) {
				oldestBackup = info.ModTime()
			}
			if info.ModTime().After(newestBackup) {
				newestBackup = info.ModTime()
			}
		}
	}
	
	stats := map[string]interface{}{
		"total_backups": totalBackups,
		"total_size":    totalSize,
		"backup_dir":    cpm.backupDir,
	}
	
	if totalBackups > 0 {
		stats["oldest_backup"] = oldestBackup
		stats["newest_backup"] = newestBackup
	}
	
	return stats, nil
}

// Private helper methods

func (cpm *ContextPreservationManager) saveBackup(backup *ContextBackup) error {
	backupFile := filepath.Join(cpm.backupDir, backup.ID+".json")
	
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup: %w", err)
	}
	
	err = os.WriteFile(backupFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}
	
	return nil
}

func (cpm *ContextPreservationManager) loadBackup(backupID string) (*ContextBackup, error) {
	backupFile := filepath.Join(cpm.backupDir, backupID+".json")
	
	data, err := os.ReadFile(backupFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}
	
	var backup ContextBackup
	err = json.Unmarshal(data, &backup)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal backup: %w", err)
	}
	
	return &backup, nil
}