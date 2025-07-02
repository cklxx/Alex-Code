package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"alex/internal/agent"
	"alex/internal/config"
)

const cobraVersion = "v2.0"

// Color definitions for Claude Code style output
var (
	blue   = color.New(color.FgBlue).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	gray   = color.New(color.FgHiBlack).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
)

// Input box template for consistent display
const inputBox = `┌────────────────────────────────────────────────────────────────┐
│                                                                │
└────────────────────────────────────────────────────────────────┘`

// Styling functions for Deep Coding Agent output
func DeepCodingError(msg string) string {
	return red("❌ " + msg)
}

func DeepCodingAction(msg string) string {
	return blue("🔧 " + msg)
}

func DeepCodingThinking(msg string) string {
	return yellow("🤔 " + msg)
}

func DeepCodingResult(msg string) string {
	return green("✅ " + msg)
}

func DeepCodingSuccess(msg string) string {
	return green("🎉 " + msg)
}

func DeepCodingToolExecution(title, content string) string {
	return fmt.Sprintf("%s %s:\n%s\n", cyan("🛠️"), title, content)
}

// CLI holds the command line interface state
type CLI struct {
	agent            *agent.ReactAgent
	config           *config.Manager
	interactive      bool
	verbose          bool
	debug            bool
	useTUI           bool // Whether to use Bubble Tea TUI
	currentTermCtrl  *TerminalController
	currentStartTime time.Time
	contentBuffer    bytes.Buffer       // Buffer for accumulating streaming content
	timerCancel      context.CancelFunc // Cancel function for timer
	processing       bool               // Whether currently processing
	currentMessage   string             // Current working message
	inputQueue       chan string        // Queue for pending inputs during processing
}

// NewRootCommand creates the root cobra command
func NewRootCommand() *cobra.Command {
	cli := &CLI{
		inputQueue: make(chan string, 10), // Buffer for 10 pending inputs
	}

	rootCmd := &cobra.Command{
		Use:   "alex",
		Short: "🤖 AI-powered coding assistant with ReAct intelligence",
		Long: fmt.Sprintf(`%s

%s is an intelligent coding assistant built on ReAct (Reasoning and Acting) architecture.
It provides natural language interface for code analysis, file operations, and development tasks
through streaming responses and advanced tool calling capabilities.

%s
  alex                           # Interactive mode
  alex "analyze this project"    # Single prompt
  alex -r session_123            # Resume session
  alex config show               # Show configuration

%s
  • 🧠 ReAct Intelligence - Think, Act, Observe cycle
  • 🌊 Streaming Responses - Real-time feedback  
  • 🛠️ Advanced Tools - File operations, shell, web search
  • 📁 Session Management - Persistent conversations
  • ⚙️ Smart Configuration - Multi-model support`,
			bold("Deep Coding Agent "+cobraVersion),
			bold("Deep Coding Agent"),
			bold("EXAMPLES:"),
			bold("FEATURES:")),
		Version: cobraVersion,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return cli.initialize(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// Single prompt mode
				prompt := strings.Join(args, " ")
				return cli.runSinglePrompt(prompt)
			}
			// Always use Bubble Tea TUI for interactive mode
			return cli.runTUI()
		},
	}

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&cli.interactive, "interactive", "i", false, "Interactive mode")
	rootCmd.PersistentFlags().BoolVarP(&cli.verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVarP(&cli.debug, "debug", "d", false, "Debug mode")
	rootCmd.PersistentFlags().BoolVar(&cli.useTUI, "tui", false, "Use Bubble Tea TUI (experimental)")
	rootCmd.PersistentFlags().StringP("resume", "r", "", "Resume session by ID")
	rootCmd.PersistentFlags().StringP("model", "m", "", "Specify model")
	rootCmd.PersistentFlags().IntP("tokens", "t", 2000, "Max tokens")
	rootCmd.PersistentFlags().Float64P("temperature", "", 0.7, "Temperature")

	// Add subcommands
	rootCmd.AddCommand(newConfigCommand(cli))
	rootCmd.AddCommand(newSessionCommand(cli))
	rootCmd.AddCommand(createToolsCommands(cli))

	// Configure viper
	viper.SetConfigName("deep-coding-config")
	viper.SetConfigType("json")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(".")

	return rootCmd
}

// newConfigCommand creates the config subcommand
func newConfigCommand(cli *CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "⚙️ Configuration management",
		Long:  "Manage Alex configuration settings",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli.showConfig()
			return nil
		},
	})

	return cmd
}

// initialize sets up the CLI
func (cli *CLI) initialize(cmd *cobra.Command) error {
	// Redirect logs to file to prevent interference with UI
	if !cli.debug {
		logFile, err := os.OpenFile("alex-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(logFile)
		} else {
			// If can't create log file, disable logging
			log.SetOutput(io.Discard)
		}
	}

	// Initialize markdown renderer
	if err := InitMarkdownRenderer(); err != nil {
		if cli.debug {
			fmt.Printf("⚠️  Failed to initialize markdown renderer: %v\n", err)
		}
	}

	// Read config
	if err := viper.ReadInConfig(); err != nil {
		if cli.debug {
			fmt.Printf("⚠️  Config file not found: %v\n", err)
		}
	}

	// Create configuration manager
	configManager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}
	cli.config = configManager

	// Create agent
	agentInstance, err := agent.NewReactAgent(configManager)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}
	cli.agent = agentInstance

	// Handle session resume
	if resumeID, _ := cmd.Flags().GetString("resume"); resumeID != "" {
		if _, err := cli.agent.RestoreSession(resumeID); err != nil {
			return fmt.Errorf("failed to resume session %s: %w", resumeID, err)
		}
		fmt.Printf("%s Resumed session: %s\n", blue("📁"), resumeID)
	} else {
		if _, err := cli.agent.StartSession(""); err != nil {
			return fmt.Errorf("failed to start session: %w", err)
		}
	}

	return nil
}

// processInputAsync handles input processing in background
func (cli *CLI) processInputAsync(input string, termCtrl *TerminalController) {
	// Set processing state
	start := time.Now()
	cli.currentTermCtrl = termCtrl
	cli.currentStartTime = start
	cli.processing = true
	cli.currentMessage = "Processing"

	// Start the working indicator timer (only once)
	cli.startWorkingIndicatorTimer("Processing")

	// Show initial processing indicator (keep input box visible)
	workingIndicator := cli.formatWorkingIndicator("Processing", start, 0)
	termCtrl.ShowFixedBottomInterface(workingIndicator, inputBox)

	// Process the input
	ctx := context.Background()
	err := cli.agent.ProcessMessageStream(ctx, input, cli.config.GetConfig(), cli.deepCodingStreamCallback)
	if err != nil {
		errorMsg := DeepCodingError(fmt.Sprintf("Error: %v", err)) + "\n"
		termCtrl.PrintInScrollRegion(errorMsg)
	}

	// Stop timer and restore input box
	cli.stopWorkingIndicatorTimer()
	termCtrl.ShowFixedBottomInterface("", inputBox)
}

// runTUI starts the modern Bubble Tea TUI interface
func (cli *CLI) runTUI() error {
	return runModernTUI(cli.agent, cli.config)
}

// formatWorkingIndicator formats the working indicator string
func (cli *CLI) formatWorkingIndicator(message string, startTime time.Time, tokens int) string {
	duration := time.Since(startTime)
	if tokens > 0 {
		return color.HiBlackString(fmt.Sprintf("✶ %s… (%.0fs · %d tokens · esc to interrupt)", message, duration.Seconds(), tokens))
	}
	return color.HiBlackString(fmt.Sprintf("✶ %s… (%.0fs · esc to interrupt)", message, duration.Seconds()))
}

// startWorkingIndicatorTimer starts a timer that updates the working indicator every second
func (cli *CLI) startWorkingIndicatorTimer(message string) {
	// Cancel any existing timer
	if cli.timerCancel != nil {
		cli.timerCancel()
	}

	// Create new timer context
	ctx, cancel := context.WithCancel(context.Background())
	cli.timerCancel = cancel
	cli.processing = true

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if cli.processing && cli.currentTermCtrl != nil {
					// Use current message if available, fallback to original message
					currentMsg := cli.currentMessage
					if currentMsg == "" {
						currentMsg = message
					}
					indicator := cli.formatWorkingIndicator(currentMsg, cli.currentStartTime, 0)
					cli.currentTermCtrl.UpdateWorkingIndicator(indicator)
				}
			}
		}
	}()
}

// stopWorkingIndicatorTimer stops the working indicator timer
func (cli *CLI) stopWorkingIndicatorTimer() {
	if cli.timerCancel != nil {
		cli.timerCancel()
		cli.timerCancel = nil
	}
	cli.processing = false
	cli.currentMessage = ""
}

// updateWorkingIndicatorMessage updates the working indicator message without restarting timer
func (cli *CLI) updateWorkingIndicatorMessage(message string) {
	cli.currentMessage = message
	// Immediately update display
	if cli.currentTermCtrl != nil && cli.processing {
		indicator := cli.formatWorkingIndicator(message, cli.currentStartTime, 0)
		cli.currentTermCtrl.UpdateWorkingIndicator(indicator)
	}
}

// handleSpecialCommands processes Claude Code style special commands
func (cli *CLI) handleSpecialCommands(input string) bool {
	switch {
	case input == "?":
		cli.showShortcuts()
		return true
	case strings.HasPrefix(input, "/"):
		return cli.handleSlashCommand(input)
	case strings.HasPrefix(input, "@"):
		return cli.handleAtCommand(input)
	case strings.HasPrefix(input, "!"):
		return cli.handleShellCommand(input)
	}
	return false
}

// showShortcuts displays keyboard shortcuts
func (cli *CLI) showShortcuts() {
	shortcuts := fmt.Sprintf("\n%s Keyboard Shortcuts:\n", bold("⌨️"))
	shortcuts += fmt.Sprintf("  %s  Show shortcuts\n", blue("?"))
	shortcuts += fmt.Sprintf("  %s    Slash commands (help, config, memory, etc.)\n", blue("/command"))
	shortcuts += fmt.Sprintf("  %s    Include file content\n", blue("@file.go"))
	shortcuts += fmt.Sprintf("  %s    Execute shell command\n", blue("!ls -la"))
	shortcuts += fmt.Sprintf("  %s  Exit\n", blue("exit"))

	if cli.currentTermCtrl != nil {
		cli.currentTermCtrl.PrintInScrollRegion(shortcuts)
	} else {
		fmt.Print(shortcuts)
	}
}

// handleSlashCommand processes slash commands
func (cli *CLI) handleSlashCommand(input string) bool {
	parts := strings.SplitN(input[1:], " ", 2)
	command := parts[0]

	switch command {
	case "help":
		cli.showHelp()
	case "config":
		cli.showConfig()
	case "memory":
		cli.showMemory()
	case "stats":
		cli.showStats()
	case "clear":
		if cli.currentTermCtrl != nil {
			// Clear scroll region instead of full screen
			fmt.Printf("\033[%d;%dr", cli.currentTermCtrl.scrollRegionTop, cli.currentTermCtrl.scrollRegionBot)
			fmt.Print("\033[2J")
			fmt.Printf("\033[%d;1H", cli.currentTermCtrl.scrollRegionTop)
		} else {
			fmt.Print("\033[2J\033[H") // Clear screen
		}
	default:
		errorMsg := fmt.Sprintf("%s Unknown command: /%s\n", red("❌"), command)
		errorMsg += fmt.Sprintf("%s Type /help for available commands\n", gray("💡"))
		if cli.currentTermCtrl != nil {
			cli.currentTermCtrl.PrintInScrollRegion(errorMsg)
		} else {
			fmt.Print(errorMsg)
		}
	}
	return true
}

// handleAtCommand includes file content
func (cli *CLI) handleAtCommand(input string) bool {
	filename := strings.TrimSpace(input[1:])
	if filename == "" {
		errorMsg := fmt.Sprintf("%s Usage: @filename.go\n", red("❌"))
		if cli.currentTermCtrl != nil {
			cli.currentTermCtrl.PrintInScrollRegion(errorMsg)
		} else {
			fmt.Print(errorMsg)
		}
		return true
	}

	// TODO: Implement file reading and injection
	message := fmt.Sprintf("%s File injection: %s (feature coming soon)\n", blue("📄"), filename)
	if cli.currentTermCtrl != nil {
		cli.currentTermCtrl.PrintInScrollRegion(message)
	} else {
		fmt.Print(message)
	}
	return true
}

// handleShellCommand executes shell commands
func (cli *CLI) handleShellCommand(input string) bool {
	command := strings.TrimSpace(input[1:])
	if command == "" {
		errorMsg := fmt.Sprintf("%s Usage: !command\n", red("❌"))
		if cli.currentTermCtrl != nil {
			cli.currentTermCtrl.PrintInScrollRegion(errorMsg)
		} else {
			fmt.Print(errorMsg)
		}
		return true
	}

	// TODO: Implement shell execution
	message := fmt.Sprintf("%s Shell command: %s (feature coming soon)\n", blue("🔧"), command)
	if cli.currentTermCtrl != nil {
		cli.currentTermCtrl.PrintInScrollRegion(message)
	} else {
		fmt.Print(message)
	}
	return true
}

// deepCodingStreamCallback handles streaming responses with Deep Coding Agent styling
func (cli *CLI) deepCodingStreamCallback(chunk agent.StreamChunk) {
	var content string

	switch chunk.Type {
	case "status":
		content = DeepCodingAction(chunk.Content) + "\n"
	case "thinking_start":
		content = DeepCodingThinking("Analyzing your request...") + "\n"
		// Update timer message to "Thinking" (don't restart timer)
		if cli.processing {
			cli.updateWorkingIndicatorMessage("Thinking")
		}
	case "thinking_result":
		// Render thinking result as markdown if it contains markdown
		content = DeepCodingResult(chunk.Content)
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
	case "action_start":
		content = DeepCodingAction("Taking action...") + "\n"
		// Update timer message to "Working" (don't restart timer)
		if cli.processing {
			cli.updateWorkingIndicatorMessage("Working")
		}
	case "tool_start":
		content = DeepCodingAction(chunk.Content) + "\n"
	case "tool_result":
		content = DeepCodingToolExecution("Tool Result", chunk.Content)
	case "tool_error":
		content = DeepCodingError(chunk.Content) + "\n"
	case "final_answer":
		content = "\n" + DeepCodingResult(chunk.Content)
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
	case "task_complete":
		content = DeepCodingSuccess("Task completed") + "\n"
	case "llm_content", "content":
		// Accumulate streaming content for better markdown processing
		cli.contentBuffer.WriteString(chunk.Content)
		// For immediate display, show raw content without markdown processing
		content = chunk.Content
	case "error":
		content = DeepCodingError(chunk.Content) + "\n"
	case "complete":
		if cli.debug {
			content = DeepCodingSuccess("Stream completed") + "\n"
		}
		// Process accumulated content for markdown rendering
		if cli.contentBuffer.Len() > 0 {
			bufferedContent := cli.contentBuffer.String()
			if ShouldRenderAsMarkdown(bufferedContent) {
				renderedContent := RenderMarkdown(bufferedContent)
				if cli.currentTermCtrl != nil {
					cli.currentTermCtrl.PrintInScrollRegion("\n--- Formatted Output ---\n" + renderedContent)
				} else {
					fmt.Print("\n--- Formatted Output ---\n" + renderedContent)
				}
			}
			cli.contentBuffer.Reset()
		}
		// Update message to show completion
		if cli.processing {
			cli.updateWorkingIndicatorMessage("Completed")
		}
	default:
		if cli.debug {
			content = fmt.Sprintf("Unknown chunk type: %s\n", chunk.Type)
		}
	}

	// Print content in scroll region if we have terminal controller
	if content != "" {
		if cli.currentTermCtrl != nil {
			cli.currentTermCtrl.PrintInScrollRegion(content)
		} else {
			// Fallback to regular print
			fmt.Print(content)
		}
	}
}

// runSinglePrompt handles single prompt execution
func (cli *CLI) runSinglePrompt(prompt string) error {
	if cli.verbose {
		fmt.Printf("%s Processing: %s\n", blue("⚡"), prompt)
	}

	ctx := context.Background()
	return cli.agent.ProcessMessageStream(ctx, prompt, cli.config.GetConfig(), cli.deepCodingStreamCallback)
}

// Helper methods for slash commands
func (cli *CLI) showHelp() {
	help := fmt.Sprintf("\n%s Available Commands:\n", bold("📖"))
	help += fmt.Sprintf("  %s         Show this help\n", blue("/help"))
	help += fmt.Sprintf("  %s       Show configuration\n", blue("/config"))
	help += fmt.Sprintf("  %s       Show memory status\n", blue("/memory"))
	help += fmt.Sprintf("  %s        Show session statistics\n", blue("/stats"))
	help += fmt.Sprintf("  %s        Clear screen\n", blue("/clear"))

	if cli.currentTermCtrl != nil {
		cli.currentTermCtrl.PrintInScrollRegion(help)
	} else {
		fmt.Print(help)
	}
}

func (cli *CLI) showConfig() {
	config := fmt.Sprintf("\n%s Current Configuration:\n", bold("⚙️"))
	// TODO: Display actual config
	config += fmt.Sprintf("  Model: %s\n", blue("deepseek-chat-v3"))
	config += fmt.Sprintf("  Max Tokens: %s\n", blue("2000"))
	config += fmt.Sprintf("  Temperature: %s\n", blue("0.7"))

	if cli.currentTermCtrl != nil {
		cli.currentTermCtrl.PrintInScrollRegion(config)
	} else {
		fmt.Print(config)
	}
}

func (cli *CLI) showMemory() {
	memory := fmt.Sprintf("\n%s Memory Status:\n", bold("🧠"))
	// TODO: Display actual memory stats
	memory += fmt.Sprintf("  Context Size: %s\n", blue("1.2MB"))
	memory += fmt.Sprintf("  Messages: %s\n", blue("42"))

	if cli.currentTermCtrl != nil {
		cli.currentTermCtrl.PrintInScrollRegion(memory)
	} else {
		fmt.Print(memory)
	}
}

func (cli *CLI) showStats() {
	stats := fmt.Sprintf("\n%s Session Statistics:\n", bold("📊"))
	// TODO: Display actual stats
	stats += fmt.Sprintf("  Duration: %s\n", blue("15m 32s"))
	stats += fmt.Sprintf("  Tokens Used: %s\n", blue("1,247"))
	stats += fmt.Sprintf("  Tools Called: %s\n", blue("8"))

	if cli.currentTermCtrl != nil {
		cli.currentTermCtrl.PrintInScrollRegion(stats)
	} else {
		fmt.Print(stats)
	}
}

// runCobraCLI initializes and runs the new Cobra-driven CLI
func runCobraCLI() {
	rootCmd := NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%s %v\n", red("Error:"), err)
		os.Exit(1)
	}
}
