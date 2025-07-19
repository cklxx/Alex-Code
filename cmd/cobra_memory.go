package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"alex/internal/utils"
	"github.com/spf13/cobra"
)

// newMemoryCommand creates the memory management command
func newMemoryCommand(cli *CLI) *cobra.Command {
	memoryCmd := &cobra.Command{
		Use:     "memory",
		Short:   "🧠 Memory management",
		Long:    "Manage project-based memories and knowledge",
		Aliases: []string{"mem", "m"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.listMemories()
		},
	}

	// memory list
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "List all memories",
		Long:    "Display all stored memories for the current project",
		Aliases: []string{"ls", "l"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.listMemories()
		},
	}

	// memory stats
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show memory statistics",
		Long:  "Display detailed memory usage statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.showMemoryStats()
		},
	}

	// memory clear
	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear project memories",
		Long:  "Clear all memories for the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.clearProjectMemories()
		},
	}

	memoryCmd.AddCommand(listCmd, statsCmd, clearCmd)
	return memoryCmd
}

// listMemories displays all stored memories from both short-term and long-term memory systems
func (cli *CLI) listMemories() error {
	fmt.Printf("\n%s Project Memories:\n", bold("🧠"))
	fmt.Println()

	// Show project information
	if err := cli.showProjectInfo(); err != nil {
		fmt.Printf("%s Project information unavailable: %v\n", yellow("⚠️"), err)
	}
	fmt.Println()

	// Check if we have an agent for runtime stats
	if cli.agent != nil {
		// Get memory statistics from agent
		stats := cli.agent.GetMemoryStats()
		if stats != nil {
			cli.displayMemoryStats(stats)
			fmt.Printf("%s Runtime memory statistics (current session)\n", gray("💡"))
		}
	}

	// Always show disk-based memory statistics
	cli.displayDiskMemoryStats()
	
	fmt.Printf("%s Memories are project-based and persist across sessions\n", gray("💡"))
	return nil
}

// showMemoryStats displays detailed memory statistics
func (cli *CLI) showMemoryStats() error {
	fmt.Printf("\n%s Memory Statistics:\n", bold("📊"))
	fmt.Println()

	// Show project information
	if err := cli.showProjectInfo(); err != nil {
		fmt.Printf("%s Project information unavailable: %v\n", yellow("⚠️"), err)
		return err
	}
	fmt.Println()

	// Get project ID for filtering
	projectID, err := utils.GenerateProjectID()
	if err != nil {
		fmt.Printf("%s Cannot generate project ID: %v\n", yellow("⚠️"), err)
		projectID = "unknown"
	}

	// Show detailed disk statistics
	cli.displayDetailedDiskStats(projectID)

	// Show runtime statistics if available
	if cli.agent != nil {
		stats := cli.agent.GetMemoryStats()
		if stats != nil {
			fmt.Printf("%s Runtime Memory Details:\n", blue("⚡"))
			cli.displayMemoryStats(stats)
		}
	}

	return nil
}

// clearProjectMemories clears all memories for the current project
func (cli *CLI) clearProjectMemories() error {
	fmt.Printf("\n%s Clear Project Memories:\n", bold("🗑️"))
	fmt.Println()

	// Get project information
	projectID, err := utils.GenerateProjectID()
	if err != nil {
		return fmt.Errorf("cannot generate project ID: %w", err)
	}

	projectName, err := utils.GetProjectDisplayName()
	if err != nil {
		return fmt.Errorf("cannot get project name: %w", err)
	}

	fmt.Printf("Project: %s (%s)\n", blue(projectName), blue(projectID))
	fmt.Println()

	// Count memories to be deleted
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot access home directory: %w", err)
	}

	memoryDir := filepath.Join(homeDir, ".deep-coding-memory", "long-term")
	entries, err := os.ReadDir(memoryDir)
	if err != nil {
		fmt.Printf("%s No memories found to clear\n", yellow("⚠️"))
		return nil
	}

	var projectFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			// Check if file belongs to current project by checking filename pattern
			if strings.Contains(entry.Name(), projectID) {
				projectFiles = append(projectFiles, entry.Name())
			}
		}
	}

	if len(projectFiles) == 0 {
		fmt.Printf("%s No memories found for this project\n", yellow("⚠️"))
		return nil
	}

	fmt.Printf("%s Found %d memories for this project\n", blue("📋"), len(projectFiles))
	fmt.Printf("%s This will permanently delete all project memories!\n", red("⚠️"))
	fmt.Print("Are you sure? (yes/no): ")

	var response string
	_, _ = fmt.Scanln(&response)

	if strings.ToLower(response) != "yes" && strings.ToLower(response) != "y" {
		fmt.Printf("%s Operation cancelled\n", yellow("❌"))
		return nil
	}

	// Delete project memory files
	deletedCount := 0
	for _, filename := range projectFiles {
		filePath := filepath.Join(memoryDir, filename)
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("%s Failed to delete %s: %v\n", red("❌"), filename, err)
		} else {
			deletedCount++
		}
	}

	fmt.Printf("%s Successfully deleted %d memories\n", green("✅"), deletedCount)
	return nil
}

// displayDetailedDiskStats shows detailed disk-based memory statistics
func (cli *CLI) displayDetailedDiskStats(projectID string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("%s Cannot access home directory: %v\n", yellow("⚠️"), err)
		return
	}

	memoryDir := filepath.Join(homeDir, ".deep-coding-memory", "long-term")
	entries, err := os.ReadDir(memoryDir)
	if err != nil {
		fmt.Printf("%s No persistent memories found\n", yellow("⚠️"))
		return
	}

	var totalSize, projectSize int64
	totalCount, projectCount := 0, 0

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			totalCount++
			if info, err := entry.Info(); err == nil {
				totalSize += info.Size()
				
				// Check if this memory belongs to current project
				if strings.Contains(entry.Name(), projectID) {
					projectCount++
					projectSize += info.Size()
				}
			}
		}
	}

	fmt.Printf("%s Storage Statistics:\n", blue("💾"))
	fmt.Printf("  Storage Location: %s\n", blue(memoryDir))
	fmt.Printf("  Total Memories: %s\n", blue(fmt.Sprintf("%d", totalCount)))
	fmt.Printf("  Total Size: %s\n", blue(formatFileSize(totalSize)))
	fmt.Println()
	
	fmt.Printf("%s Current Project:\n", blue("🏗️"))
	fmt.Printf("  Project Memories: %s\n", blue(fmt.Sprintf("%d", projectCount)))
	fmt.Printf("  Project Size: %s\n", blue(formatFileSize(projectSize)))
	if totalCount > 0 {
		fmt.Printf("  Percentage: %s\n", blue(fmt.Sprintf("%.1f%%", float64(projectCount)/float64(totalCount)*100)))
	}
	fmt.Println()
}

// formatFileSize formats file size in human-readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// showProjectInfo displays current project information
func (cli *CLI) showProjectInfo() error {
	projectID, err := utils.GenerateProjectID()
	if err != nil {
		return err
	}

	displayName, err := utils.GetProjectDisplayName()
	if err != nil {
		return err
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	fmt.Printf("%s Project Information:\n", blue("🏗️"))
	fmt.Printf("  Name: %s\n", blue(displayName))
	fmt.Printf("  ID: %s\n", blue(projectID))
	fmt.Printf("  Path: %s\n", blue(workingDir))

	return nil
}

// displayMemoryStats displays memory statistics from the agent
func (cli *CLI) displayMemoryStats(stats map[string]interface{}) {
	fmt.Printf("%s Runtime Memory Statistics:\n", blue("📊"))
	if totalItems, ok := stats["total_items"]; ok {
		fmt.Printf("  Total Items: %s\n", blue(fmt.Sprintf("%v", totalItems)))
	}
	if totalSize, ok := stats["total_size"]; ok {
		fmt.Printf("  Total Size: %s\n", blue(formatFileSize(totalSize.(int64))))
	}
	fmt.Println()

	// Display short-term memory
	if shortTerm, ok := stats["short_term"]; ok {
		if shortStats, ok := shortTerm.(map[string]interface{}); ok {
			fmt.Printf("%s Short-Term Memory:\n", green("⏰"))
			if items, ok := shortStats["total_items"]; ok {
				fmt.Printf("  Items: %s\n", blue(fmt.Sprintf("%v", items)))
			}
			if size, ok := shortStats["total_size"]; ok {
				fmt.Printf("  Size: %s\n", blue(formatFileSize(size.(int64))))
			}
			fmt.Println()
		}
	}

	// Display long-term memory
	if longTerm, ok := stats["long_term"]; ok {
		if longStats, ok := longTerm.(map[string]interface{}); ok {
			fmt.Printf("%s Long-Term Memory:\n", purple("🗄️"))
			if items, ok := longStats["total_items"]; ok {
				fmt.Printf("  Items: %s\n", blue(fmt.Sprintf("%v", items)))
			}
			if size, ok := longStats["total_size"]; ok {
				fmt.Printf("  Size: %s\n", blue(formatFileSize(size.(int64))))
			}
			fmt.Println()
		}
	}
}

// displayDiskMemoryStats displays actual stored memory statistics from disk
func (cli *CLI) displayDiskMemoryStats() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("%s Cannot access home directory: %v\n", yellow("⚠️"), err)
		return
	}

	memoryDir := filepath.Join(homeDir, ".deep-coding-memory", "long-term")
	entries, err := os.ReadDir(memoryDir)
	if err != nil {
		fmt.Printf("%s No persistent memories found\n", yellow("⚠️"))
		return
	}

	var totalSize int64
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			count++
			if info, err := entry.Info(); err == nil {
				totalSize += info.Size()
			}
		}
	}

	fmt.Printf("%s Persistent Memory Statistics:\n", blue("💾"))
	fmt.Printf("  Stored Items: %s\n", blue(fmt.Sprintf("%d", count)))
	fmt.Printf("  Total Size: %s\n", blue(formatFileSize(totalSize)))
	fmt.Printf("  Storage: %s\n", blue(memoryDir))
	fmt.Println()
}