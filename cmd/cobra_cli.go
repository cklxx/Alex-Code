package main

import (
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
	"golang.org/x/term"

	"alex/internal/agent"
	"alex/internal/config"
)

const cobraVersion = "v2.0"

// isTTY checks if the current environment has a TTY available
func isTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

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

func DeepCodingReasoning(msg string) string {
	return cyan("🧠 " + msg)
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
	contentBuffer    strings.Builder // Buffer for accumulating streaming content (using strings.Builder for better performance)
	processing       bool            // Whether currently processing
	currentMessage   string          // Current working message
	inputQueue       chan string     // Queue for pending inputs during processing
}

// NewRootCommand creates the root cobra command
func NewRootCommand() *cobra.Command {
	cli := &CLI{
		inputQueue: make(chan string, 10), // Buffer for 10 pending inputs
	}

	// Pre-allocate contentBuffer for better streaming performance
	cli.contentBuffer.Grow(4096) // Pre-allocate 4KB buffer

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
		Args:    cobra.ArbitraryArgs, // Allow arbitrary arguments for single prompt mode
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// Single prompt mode - initialize first
				if err := cli.initialize(cmd); err != nil {
					return err
				}
				prompt := strings.Join(args, " ")
				return cli.runSinglePrompt(prompt)
			}
			// Check if we have a TTY before starting interactive mode
			if !isTTY() {
				// No TTY available (CI environment), show help instead
				return cmd.Help()
			}
			// Initialize for interactive mode
			if err := cli.initialize(cmd); err != nil {
				return err
			}
			// Use Bubble Tea TUI for interactive mode
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
	rootCmd.AddCommand(newBatchCommand())

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
			// Initialize config manager for config commands
			if err := cli.initializeConfigOnly(); err != nil {
				return err
			}
			cli.showConfig()
			return nil
		},
	})

	return cmd
}

// initializeConfigOnly sets up only the configuration manager
func (cli *CLI) initializeConfigOnly() error {
	// Create configuration manager if not already created
	if cli.config == nil {
		configManager, err := config.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create config manager: %w", err)
		}
		cli.config = configManager
	}
	return nil
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

// updateWorkingIndicatorMessage updates the working indicator message without restarting timer
func (cli *CLI) updateWorkingIndicatorMessage(message string) {
	cli.currentMessage = message
	// Immediately update display
	if cli.currentTermCtrl != nil && cli.processing {
		indicator := cli.formatWorkingIndicator(message, cli.currentStartTime, 0)
		cli.currentTermCtrl.UpdateWorkingIndicator(indicator)
	}
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
	case "reasoning":
		// Handle OpenAI reasoning tokens
		content = DeepCodingReasoning("Reasoning: " + chunk.Content)
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
	case "reasoning_summary":
		// Handle OpenAI reasoning summary
		content = DeepCodingReasoning("Summary: " + chunk.Content)
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
	case "think":
		// Handle <think> tags from model responses
		content = DeepCodingThinking("Model thinking: " + chunk.Content)
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
	case "iteration":
		// Handle ReAct iteration chunks - these represent steps in the think-act-observe cycle
		if cli.debug {
			content = DeepCodingReasoning("ReAct iteration: "+chunk.Content) + "\n"
		}
	case "llm_content", "content":
		// Accumulate streaming content for better markdown processing
		cli.contentBuffer.WriteString(chunk.Content)
		// For immediate display, show raw content without markdown processing
		content = chunk.Content
	case "error":
		content = DeepCodingError(chunk.Content) + "\n"
	case "complete":
		// Process accumulated content for markdown rendering
		if cli.contentBuffer.Len() > 0 {
			bufferedContent := cli.contentBuffer.String()
			if ShouldRenderAsMarkdown(bufferedContent) {
				renderedContent := RenderMarkdown(bufferedContent)
				if cli.currentTermCtrl != nil {
					cli.currentTermCtrl.PrintInScrollRegion("\n--- Output ---\n" + renderedContent)
				} else {
					fmt.Print("\n--- Output ---\n" + renderedContent)
				}
			}
			cli.contentBuffer.Reset()
			cli.contentBuffer.Grow(4096) // Re-allocate buffer after reset for next use
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

	// Output the content if it's not empty
	if content != "" && chunk.Type != "complete" {
		if cli.currentTermCtrl != nil {
			cli.currentTermCtrl.PrintInScrollRegion(content)
		} else {
			fmt.Print(content)
		}
	}
}

// runSinglePrompt handles single prompt execution
func (cli *CLI) runSinglePrompt(prompt string) error {
	// Record start time
	startTime := time.Now()
	
	if cli.verbose {
		fmt.Printf("%s Processing: %s\n", blue("⚡"), prompt)
	}

	ctx := context.Background()
	err := cli.agent.ProcessMessageStream(ctx, prompt, cli.config.GetConfig(), cli.deepCodingStreamCallback)
	
	// Calculate and display completion time
	duration := time.Since(startTime)
	
	// Format duration nicely
	var durationStr string
	if duration < time.Second {
		durationStr = fmt.Sprintf("%.0fms", duration.Seconds()*1000)
	} else if duration < time.Minute {
		durationStr = fmt.Sprintf("%.1fs", duration.Seconds())
	} else {
		durationStr = fmt.Sprintf("%.1fm", duration.Minutes())
	}
	
	// Display completion message with time
	if err != nil {
		fmt.Printf("\n%s Task failed after %s\n", red("❌"), durationStr)
	} else {
		fmt.Printf("\n%s Task completed in %s\n", green("✅"), durationStr)
	}
	
	return err
}

func (cli *CLI) showConfig() {
	cfg := cli.config.GetConfig()
	config := fmt.Sprintf("\n%s Current Configuration:\n", bold("⚙️"))

	// Display legacy config (for compatibility)
	config += fmt.Sprintf("  %s: %s\n", bold("Model"), blue(cfg.Model))
	config += fmt.Sprintf("  %s: %s\n", bold("Max Tokens"), blue(fmt.Sprintf("%d", cfg.MaxTokens)))
	config += fmt.Sprintf("  %s: %s\n", bold("Temperature"), blue(fmt.Sprintf("%.1f", cfg.Temperature)))
	config += fmt.Sprintf("  %s: %s\n", bold("Base URL"), blue(cfg.BaseURL))
	config += fmt.Sprintf("  %s: %s\n", bold("Max Turns"), blue(fmt.Sprintf("%d", cfg.MaxTurns)))

	// Display tool configuration
	if cfg.TavilyAPIKey != "" {
		config += fmt.Sprintf("\n%s Tool Configuration:\n", bold("🛠️"))
		maskedKey := cfg.TavilyAPIKey[:8] + "..." + cfg.TavilyAPIKey[len(cfg.TavilyAPIKey)-8:]
		config += fmt.Sprintf("  %s: %s\n", bold("Tavily API Key"), blue(maskedKey))
	}

	// Display multi-model configurations if available
	if len(cfg.Models) > 0 {
		config += fmt.Sprintf("\n%s Multi-Model Configurations:\n", bold("🤖"))
		config += fmt.Sprintf("  %s: %s\n", bold("Default Model Type"), blue(string(cfg.DefaultModelType)))

		for modelType, modelConfig := range cfg.Models {
			config += fmt.Sprintf("\n  %s %s:\n", bold("📋"), bold(string(modelType)))
			config += fmt.Sprintf("    %s: %s\n", "Model", blue(modelConfig.Model))
			config += fmt.Sprintf("    %s: %s\n", "Max Tokens", blue(fmt.Sprintf("%d", modelConfig.MaxTokens)))
			config += fmt.Sprintf("    %s: %s\n", "Temperature", blue(fmt.Sprintf("%.1f", modelConfig.Temperature)))
			config += fmt.Sprintf("    %s: %s\n", "Base URL", blue(modelConfig.BaseURL))
			// Mask API key for security
			if modelConfig.APIKey != "" {
				maskedKey := modelConfig.APIKey[:8] + "..." + modelConfig.APIKey[len(modelConfig.APIKey)-8:]
				config += fmt.Sprintf("    %s: %s\n", "API Key", blue(maskedKey))
			}
		}
	}

	if cli.currentTermCtrl != nil {
		cli.currentTermCtrl.PrintInScrollRegion(config)
	} else {
		fmt.Print(config)
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
