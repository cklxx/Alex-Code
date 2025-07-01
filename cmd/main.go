package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"deep-coding-agent/internal/agent"
	"deep-coding-agent/internal/config"
)

const version = "v1.0"

// Logger instance for CLI operations
var cliLogger *log.Logger
var debugMode bool

// Initialize logging
func init() {
	if os.Getenv("DEBUG") == "true" {
		cliLogger = log.New(os.Stdout, "[CLI] ", log.LstdFlags|log.Lshortfile)
	} else {
		cliLogger = log.New(os.Stdout, "[CLI] ", log.LstdFlags)
	}
}

// CLIConfig contains only essential configuration
type CLIConfig struct {
	Interactive  bool     `json:"interactive"`
	SessionID    string   `json:"sessionId"`
	AllowedTools []string `json:"allowedTools"`
	MaxTokens    int      `json:"maxTokens"`
	Temperature  float64  `json:"temperature"`
}

func main() {
	// Check for config subcommand first
	if len(os.Args) > 1 && os.Args[1] == "config" {
		if err := handleConfigCommand(os.Args[2:]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Parse essential flags only
	var cliConfig CLIConfig
	var resumeSession string
	var listSessions bool
	var verbose bool
	var debug bool

	flag.BoolVar(&cliConfig.Interactive, "i", false, "Interactive mode")
	flag.StringVar(&resumeSession, "r", "", "Resume session")
	flag.BoolVar(&listSessions, "ls", false, "List sessions")
	flag.IntVar(&cliConfig.MaxTokens, "tokens", 2000, "Max tokens")
	flag.Float64Var(&cliConfig.Temperature, "temp", 0.7, "Temperature")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.BoolVar(&debug, "debug", false, "Debug mode with detailed logging")

	flag.Usage = func() {
		fmt.Printf(`Deep Coding Agent %s - AI Coding Assistant

USAGE:
    deep-coding-agent [flags] [prompt]
    deep-coding-agent config <subcommand> [args]

FLAGS:
    -i               Interactive mode
    -r <session>     Resume session
    -ls              List sessions
    -tokens <int>    Max tokens (default: 2000)
    -temp <float>    Temperature (default: 0.7)
    -v               Verbose output
    -debug           Debug mode with detailed logging

CONFIG COMMANDS:
    config show      Show current configuration
    config set       Set configuration values
    config list      List all configuration keys
    config validate  Validate configuration
    config reset     Reset to defaults

EXAMPLES:
    deep-coding-agent -i                           # Interactive mode
    deep-coding-agent -i -v                       # Interactive mode with verbose logging
    deep-coding-agent -i -debug                   # Interactive mode with debug logging
    deep-coding-agent "List files in current dir"  # Single prompt
    deep-coding-agent -v "Analyze this project"   # Single prompt with verbose output
    deep-coding-agent -r session_123 -i           # Resume session
    deep-coding-agent config show                  # Show config
    deep-coding-agent config set api_key sk-...   # Set API key

`, version)
	}

	flag.Parse()

	// Set logging mode based on flags
	if debug || debugMode {
		debugMode = true
		cliLogger = log.New(os.Stdout, "[CLI] ", log.LstdFlags|log.Lshortfile)
	} else if verbose {
		cliLogger = log.New(os.Stdout, "[CLI] ", log.LstdFlags)
	}

	// Create configuration manager
	if debugMode || verbose {
		cliLogger.Println("🔧 Initializing configuration manager...")
	}
	configManager, err := config.NewManager()
	if err != nil {
		if debugMode || verbose {
			cliLogger.Printf("❌ Error creating config manager: %v\n", err)
		}
		fmt.Printf("Error creating config manager: %v\n", err)
		os.Exit(1)
	}
	if debugMode || verbose {
		cliLogger.Println("✅ Configuration manager initialized")
	}

	// Create agent
	if debugMode || verbose {
		cliLogger.Println("🤖 Initializing ReAct agent...")
	}
	agentInstance, err := agent.NewReactAgent(configManager)
	if err != nil {
		if debugMode || verbose {
			cliLogger.Printf("❌ Error creating agent: %v\n", err)
		}
		fmt.Printf("Error creating agent: %v\n", err)
		os.Exit(1)
	}
	if debugMode || verbose {
		cliLogger.Println("✅ ReAct agent initialized")
	}

	// Handle session commands
	if listSessions {
		listAvailableSessions()
		return
	}

	if resumeSession != "" {
		if debugMode || verbose {
			cliLogger.Printf("📁 Attempting to resume session: %s", resumeSession)
		}
		_, err := agentInstance.RestoreSession(resumeSession)
		if err != nil {
			if debugMode || verbose {
				cliLogger.Printf("❌ Error resuming session %s: %v", resumeSession, err)
			}
			fmt.Printf("Error resuming session: %v\n", err)
			os.Exit(1)
		}
		if debugMode || verbose {
			cliLogger.Printf("✅ Successfully resumed session: %s", resumeSession)
		}
		fmt.Printf("📁 Resumed session: %s\n", resumeSession)
	} else {
		if debugMode || verbose {
			cliLogger.Println("🆕 Starting new session...")
		}
		session, err := agentInstance.StartSession("")
		if err != nil {
			if debugMode || verbose {
				cliLogger.Printf("❌ Error starting new session: %v", err)
			}
			fmt.Printf("Error starting session: %v\n", err)
			os.Exit(1)
		}
		if debugMode || verbose {
			cliLogger.Printf("✅ New session started: %s", session.ID)
		}
	}

	if cliConfig.Interactive {
		if debugMode || verbose {
			cliLogger.Println("🔄 Entering interactive mode")
		}
		runInteractive(agentInstance, configManager, &cliConfig, verbose, debugMode)
	} else {
		// Single prompt mode
		prompt := strings.Join(flag.Args(), " ")
		if prompt == "" {
			flag.Usage()
			return
		}
		if debugMode || verbose {
			cliLogger.Printf("⚡ Running single prompt mode: %q", prompt)
		}
		runSinglePrompt(agentInstance, configManager, &cliConfig, prompt, verbose, debugMode)
	}
}

// runInteractive runs interactive mode
func runInteractive(agentInstance *agent.ReactAgent, configManager *config.Manager, cliConfig *CLIConfig, verbose, debug bool) {
	// Get current working directory for display
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "unknown"
	}

	fmt.Printf("🤖 Deep Coding Agent %s\n", version)
	fmt.Printf("📂 Working Directory: %s\n", currentDir)
	fmt.Println("Type your questions or 'exit' to quit.")
	fmt.Println()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n👋 Goodbye!")
		cancel()
		os.Exit(0)
	}()

	// Interactive loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		// Show current directory in prompt
		if dir, err := os.Getwd(); err == nil {
			fmt.Printf("\n📂 %s > ", filepath.Base(dir))
		} else {
			fmt.Print("\n> ")
		}

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			break
		}

		if debug {
			cliLogger.Printf("📝 Processing user input: %q", input)
		}

		// Update config with CLI settings
		if debug {
			cliLogger.Printf("⚙️ Updating config - tokens: %d, temp: %.2f, streaming: enabled",
				cliConfig.MaxTokens, cliConfig.Temperature)
		}
		configManager.Set("max_tokens", cliConfig.MaxTokens)
		configManager.Set("temperature", cliConfig.Temperature)
		configManager.Set("stream_response", true)

		// Ensure max_turns is set from config (will use default if not set)
		if maxTurns, err := configManager.Get("max_turns"); err == nil && maxTurns.(int) > 0 {
			configManager.Set("max_turns", maxTurns)
			if debug {
				cliLogger.Printf("⚙️ Using max_turns from config: %d", maxTurns)
			}
		}

		// Get config for agent
		config := configManager.GetConfig()
		if debug {
			cliLogger.Printf("🔧 Agent config prepared, starting streaming processing...")
		}

		// Always use streaming mode
		if debug {
			cliLogger.Println("🌊 Starting streaming response...")
		}
		err := agentInstance.ProcessMessageStream(ctx, input, config, func(chunk agent.StreamChunk) {
			if debug {
				cliLogger.Printf("🌊 Stream chunk: type=%s, content_len=%d", chunk.Type, len(chunk.Content))
			}
			switch chunk.Type {
			case "status":
				fmt.Printf("\n%s\n", chunk.Content)
			case "thinking_start":
				fmt.Printf("\n%s\n", chunk.Content)
			case "thinking_result":
				fmt.Printf("💭 %s\n", chunk.Content)
				if verbose || debug {
					if metadata := chunk.Metadata; metadata != nil {
						if confidence, ok := metadata["confidence"].(float64); ok {
							fmt.Printf("   (Confidence: %.2f)\n", confidence)
						}
					}
				}
			case "action_start":
				fmt.Printf("\n%s\n", chunk.Content)
			case "reasoning_only":
				fmt.Printf("🧠 %s\n", chunk.Content)
			case "observation_start":
				fmt.Printf("\n%s\n", chunk.Content)
			case "observation_result":
				fmt.Printf("📊 %s\n", chunk.Content)
				if verbose || debug {
					if metadata := chunk.Metadata; metadata != nil {
						if duration, ok := metadata["duration"].(string); ok {
							fmt.Printf("   (Duration: %s)\n", duration)
						}
					}
				}
			case "final_answer":
				fmt.Printf("\n🎯 Final Answer:\n%s\n", chunk.Content)
				if metadata := chunk.Metadata; metadata != nil {
					if confidence, ok := metadata["confidence"].(float64); ok {
						fmt.Printf("   (Confidence: %.2f)\n", confidence)
					}
				}
			case "task_complete":
				fmt.Printf("\n🏁 Task Summary:\n%s\n", chunk.Content)
				if verbose || debug {
					if metadata := chunk.Metadata; metadata != nil {
						if iterations, ok := metadata["iterations"].(int); ok {
							fmt.Printf("   (Completed in %d iterations)\n", iterations)
						}
					}
				}
			case "max_iterations":
				fmt.Printf("\n%s\n", chunk.Content)
			case "content":
				fmt.Print(chunk.Content)
			case "llm_content":
				// Real-time LLM response content streaming
				fmt.Print(chunk.Content)
			case "tool_start":
				fmt.Printf("\n🔧 %s\n", chunk.Content)
				if verbose || debug {
					cliLogger.Printf("🔧 Tool execution started: %s", chunk.Content)
				}
			case "tool_result":
				fmt.Printf("  ⎿  %s\n", chunk.Content)
				if verbose || debug {
					cliLogger.Printf("✅ Tool execution completed: %s", chunk.Content)
				}
			case "tool_error":
				fmt.Printf("❌ %s\n", chunk.Content)
				if verbose || debug {
					cliLogger.Printf("❌ Tool execution failed: %s", chunk.Content)
				}
			case "error":
				fmt.Printf("❌ %s\n", chunk.Content)
			case "complete":
				fmt.Printf("\n%s\n", chunk.Content)
				if debug {
					cliLogger.Println("✅ Streaming response completed")
				}
			default:
				// Handle unknown chunk types gracefully
				fmt.Printf("ℹ️ %s\n", chunk.Content)
			}
		})

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			if debug {
				cliLogger.Printf("❌ Streaming error: %v", err)
			}
		}
	}

	fmt.Println("\n👋 Goodbye!")
}

// runSinglePrompt runs single prompt mode
func runSinglePrompt(agentInstance *agent.ReactAgent, configManager *config.Manager, cliConfig *CLIConfig, prompt string, verbose, debug bool) {
	ctx := context.Background()

	if debug {
		cliLogger.Printf("📄 Single prompt mode - processing: %q", prompt)
	}

	// Update config with CLI settings
	if debug {
		cliLogger.Printf("⚙️ Setting config - tokens: %d, temp: %.2f, streaming: enabled", cliConfig.MaxTokens, cliConfig.Temperature)
	}
	configManager.Set("max_tokens", cliConfig.MaxTokens)
	configManager.Set("temperature", cliConfig.Temperature)
	configManager.Set("stream_response", true) // Always use streaming

	// Ensure max_turns is set from config (will use default if not set)
	if maxTurns, err := configManager.Get("max_turns"); err == nil && maxTurns.(int) > 0 {
		configManager.Set("max_turns", maxTurns)
		if debug {
			cliLogger.Printf("⚙️ Using max_turns from config: %d", maxTurns)
		}
	}

	// Get config from manager
	config := configManager.GetConfig()
	if debug {
		cliLogger.Printf("🔧 Config prepared, starting streaming for single prompt...")
	}

	// Always use streaming mode for single prompt
	if debug {
		cliLogger.Println("🌊 Starting streaming response for single prompt...")
	}
	err := agentInstance.ProcessMessageStream(ctx, prompt, config, func(chunk agent.StreamChunk) {
		if debug {
			cliLogger.Printf("🌊 Stream chunk: type=%s, content_len=%d", chunk.Type, len(chunk.Content))
		}
		switch chunk.Type {
		case "status":
			fmt.Printf("\n%s\n", chunk.Content)
		case "thinking_start":
			fmt.Printf("\n%s\n", chunk.Content)
		case "thinking_result":
			fmt.Printf("💭 %s\n", chunk.Content)
			if verbose || debug {
				if metadata := chunk.Metadata; metadata != nil {
					if confidence, ok := metadata["confidence"].(float64); ok {
						fmt.Printf("   (Confidence: %.2f)\n", confidence)
					}
				}
			}
		case "action_start":
			fmt.Printf("\n%s\n", chunk.Content)
		case "reasoning_only":
			fmt.Printf("🧠 %s\n", chunk.Content)
		case "observation_start":
			fmt.Printf("\n%s\n", chunk.Content)
		case "observation_result":
			fmt.Printf("📊 %s\n", chunk.Content)
			if verbose || debug {
				if metadata := chunk.Metadata; metadata != nil {
					if duration, ok := metadata["duration"].(string); ok {
						fmt.Printf("   (Duration: %s)\n", duration)
					}
				}
			}
		case "final_answer":
			fmt.Printf("\n🎯 Final Answer:\n%s\n", chunk.Content)
			if metadata := chunk.Metadata; metadata != nil {
				if confidence, ok := metadata["confidence"].(float64); ok {
					fmt.Printf("   (Confidence: %.2f)\n", confidence)
				}
			}
		case "task_complete":
			fmt.Printf("\n🏁 Task Summary:\n%s\n", chunk.Content)
			if verbose || debug {
				if metadata := chunk.Metadata; metadata != nil {
					if iterations, ok := metadata["iterations"].(int); ok {
						fmt.Printf("   (Completed in %d iterations)\n", iterations)
					}
				}
			}
		case "max_iterations":
			fmt.Printf("\n%s\n", chunk.Content)
		case "content":
			fmt.Print(chunk.Content)
		case "llm_content":
			// Real-time LLM response content streaming
			fmt.Print(chunk.Content)
		case "tool_start":
			fmt.Printf("\n🔧 %s\n", chunk.Content)
			if verbose || debug {
				cliLogger.Printf("🔧 Tool execution started: %s", chunk.Content)
			}
		case "tool_result":
			fmt.Printf("  ⎿  %s\n", chunk.Content)
			if verbose || debug {
				cliLogger.Printf("✅ Tool execution completed: %s", chunk.Content)
			}
		case "tool_error":
			fmt.Printf("❌ %s\n", chunk.Content)
			if verbose || debug {
				cliLogger.Printf("❌ Tool execution failed: %s", chunk.Content)
			}
		case "error":
			fmt.Printf("❌ %s\n", chunk.Content)
		case "complete":
			fmt.Printf("\n%s\n", chunk.Content)
			if debug {
				cliLogger.Println("✅ Streaming response completed")
			}
		default:
			// Handle unknown chunk types gracefully
			fmt.Printf("ℹ️ %s\n", chunk.Content)
		}
	})

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		if debug {
			cliLogger.Printf("❌ Streaming error: %v", err)
		}
		os.Exit(1)
	}
}

// listAvailableSessions lists available sessions (simplified)
func listAvailableSessions() {
	fmt.Println("📁 Available Sessions:")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return
	}

	sessionsDir := homeDir + "/.deep-coding-sessions"
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		fmt.Println("  No sessions found")
		return
	}

	if len(entries) == 0 {
		fmt.Println("  No sessions found")
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			sessionID := strings.TrimSuffix(entry.Name(), ".json")
			fmt.Printf("  %s\n", sessionID)
		}
	}
}
