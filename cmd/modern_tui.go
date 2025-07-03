package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"alex/internal/agent"
	"alex/internal/config"
)

// Modern TUI with clean, professional interface
var (
	// Color scheme
	primaryColor    = lipgloss.Color("#7C3AED")
	successColor    = lipgloss.Color("#10B981")
	warningColor    = lipgloss.Color("#F59E0B")
	errorColor      = lipgloss.Color("#EF4444")
	mutedColor      = lipgloss.Color("#6B7280")
	backgroundColor = lipgloss.Color("#1F2937") //nolint:unused

	// Styles
	headerStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1).
			Margin(0, 0, 1, 0)

	userMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#06B6D4")).
			Bold(true)

	assistantMsgStyle = lipgloss.NewStyle().
				Foreground(successColor)

	systemMsgStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	processingStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	errorMsgStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#E5E7EB")).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	sessionTimeStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true).
				Align(lipgloss.Left)
)

// Message types
type (
	streamResponseMsg struct{ content string }
	streamStartMsg    struct{ input string }
	streamChunkMsg    struct{ content string }
	streamCompleteMsg struct{}
	processingDoneMsg struct{}
	errorOccurredMsg  struct{ err error }
	tickerMsg         struct{}
)

// ModernChatModel represents the clean TUI model
type ModernChatModel struct {
	textarea         textarea.Model
	messages         []ChatMessage
	processing       bool
	agent            *agent.ReactAgent
	config           *config.Manager
	width            int
	height           int
	ready            bool
	currentInput     string
	execTimer        ExecutionTimer
	program          *tea.Program
	currentMessage   *ChatMessage  // Track current streaming message
	sessionStartTime time.Time     // Track session start time
	copyMode         bool          // Track if copy mode is enabled
}

// ChatMessage represents a chat message with type and content
type ChatMessage struct {
	Type    string // "user", "assistant", "system", "processing", "error"
	Content string
	Time    time.Time
}

// ExecutionTimer tracks execution time for processing messages
type ExecutionTimer struct {
	StartTime time.Time
	Duration  time.Duration
	Active    bool
}

// NewModernChatModel creates a clean, modern chat interface
func NewModernChatModel(agent *agent.ReactAgent, config *config.Manager) ModernChatModel {
	// Configure textarea
	ta := textarea.New()
	ta.Placeholder = "Ask me anything about coding..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 2000
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	// Initial messages
	welcomeTime := time.Now()
	initialMessages := []ChatMessage{
		{
			Type:    "system",
			Content: "🤖 Deep Coding Agent v2.0 - Powered by Bubble Tea",
			Time:    welcomeTime,
		},
		{
			Type:    "system",
			Content: fmt.Sprintf("📂 Working in: %s", getCurrentWorkingDir()),
			Time:    welcomeTime,
		},
		{
			Type:    "system",
			Content: "💡 Type your coding questions and press Enter to get help",
			Time:    welcomeTime,
		},
	}

	return ModernChatModel{
		textarea:         ta,
		messages:         initialMessages,
		agent:            agent,
		config:           config,
		ready:            false,
		sessionStartTime: time.Now(), // Initialize session start time
	}
}

func getCurrentWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	// Show only last 2 directories for brevity
	parts := strings.Split(dir, "/")
	if len(parts) > 2 {
		return ".../" + strings.Join(parts[len(parts)-2:], "/")
	}
	return dir
}

// formatSessionRuntime formats the session runtime duration
func (m ModernChatModel) formatSessionRuntime() string {
	// Try to get actual session start time from agent
	var startTime time.Time
	if m.agent != nil {
		sessionManager := m.agent.GetSessionManager()
		if sessionManager != nil {
			// Try to get current session history to find actual start time
			history := m.agent.GetSessionHistory()
			if len(history) > 0 {
				// Use the timestamp of the first message as session start
				startTime = history[0].Timestamp
			}
		}
	}
	
	// Fallback to TUI start time if no session info available
	if startTime.IsZero() {
		startTime = m.sessionStartTime
	}
	
	if startTime.IsZero() {
		return ""
	}
	
	duration := time.Since(startTime)
	
	// Format duration nicely
	if duration < time.Minute {
		return fmt.Sprintf("🕐 Session: %ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % 60
		return fmt.Sprintf("🕐 Session: %dm %ds", minutes, seconds)
	} else {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		return fmt.Sprintf("🕐 Session: %dh %dm", hours, minutes)
	}
}

func (m ModernChatModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.startTicker())
}

func (m ModernChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var tiCmd tea.Cmd

	m.textarea, tiCmd = m.textarea.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			// Initialize dimensions
			m.textarea.SetWidth(msg.Width - 6)
			m.ready = true
		} else {
			m.textarea.SetWidth(msg.Width - 6)
		}
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if !m.processing && m.textarea.Value() != "" {
				input := strings.TrimSpace(m.textarea.Value())
				m.currentInput = input
				m.textarea.Reset()

				// Add user message
				m.addMessage(ChatMessage{
					Type:    "user",
					Content: input,
					Time:    time.Now(),
				})

				// Start processing timer
				m.processing = true
				m.execTimer = ExecutionTimer{
					StartTime: time.Now(),
					Active:    true,
				}

				m.addMessage(ChatMessage{
					Type:    "processing",
					Content: "Processing your request...",
					Time:    time.Now(),
				})

				return m, tea.Batch(m.processUserInput(input), m.startTicker())
			}
		}

	case streamResponseMsg:
		// Remove last processing message and add response
		if len(m.messages) > 0 && m.messages[len(m.messages)-1].Type == "processing" {
			m.messages = m.messages[:len(m.messages)-1]
		}

		// Add execution time to response if available
		content := msg.content
		if m.execTimer.Active || !m.execTimer.StartTime.IsZero() {
			duration := time.Since(m.execTimer.StartTime)
			content += fmt.Sprintf("\n\n⏱️ Execution time: %v", duration.Truncate(10*time.Millisecond))
		}

		m.addMessage(ChatMessage{
			Type:    "assistant",
			Content: content,
			Time:    time.Now(),
		})

		return m, func() tea.Msg { return processingDoneMsg{} }

	case streamStartMsg:
		// Remove processing message and start with empty assistant message
		if len(m.messages) > 0 && m.messages[len(m.messages)-1].Type == "processing" {
			m.messages = m.messages[:len(m.messages)-1]
		}

		// Create initial assistant message for streaming
		assistantMsg := ChatMessage{
			Type:    "assistant",
			Content: "",
			Time:    time.Now(),
		}
		m.addMessage(assistantMsg)
		m.currentMessage = &m.messages[len(m.messages)-1]

		return m, nil

	case streamChunkMsg:
		// Append content to current message
		if m.currentMessage != nil {
			m.currentMessage.Content += msg.content
		}
		return m, nil

	case streamCompleteMsg:
		// Add execution time to final message
		if m.currentMessage != nil && (m.execTimer.Active || !m.execTimer.StartTime.IsZero()) {
			duration := time.Since(m.execTimer.StartTime)
			m.currentMessage.Content += fmt.Sprintf("\n\n⏱️ Execution time: %v", duration.Truncate(10*time.Millisecond))
		}
		m.currentMessage = nil
		return m, func() tea.Msg { return processingDoneMsg{} }

	case tickerMsg:
		if m.execTimer.Active {
			m.execTimer.Duration = time.Since(m.execTimer.StartTime)
			// Update the last processing message with current execution time
			if len(m.messages) > 0 && m.messages[len(m.messages)-1].Type == "processing" {
				elapsed := m.execTimer.Duration.Truncate(time.Second)
				m.messages[len(m.messages)-1].Content = fmt.Sprintf("Processing your request... (%v)", elapsed)
			}
			return m, m.startTicker() // Continue ticking
		} else {
			// Continue ticking for session runtime display even when not processing
			return m, m.startTicker()
		}

	case processingDoneMsg:
		m.processing = false
		m.execTimer.Active = false
		if m.execTimer.StartTime.IsZero() {
			m.execTimer.Duration = 0
		} else {
			m.execTimer.Duration = time.Since(m.execTimer.StartTime)
		}

	case errorOccurredMsg:
		// Remove processing message
		if len(m.messages) > 0 && m.messages[len(m.messages)-1].Type == "processing" {
			m.messages = m.messages[:len(m.messages)-1]
		}

		// Add execution time to error message if available
		errorContent := fmt.Sprintf("Error: %v", msg.err)
		if m.execTimer.Active || !m.execTimer.StartTime.IsZero() {
			duration := time.Since(m.execTimer.StartTime)
			errorContent += fmt.Sprintf("\n⏱️ Execution time: %v", duration.Truncate(10*time.Millisecond))
		}

		m.addMessage(ChatMessage{
			Type:    "error",
			Content: errorContent,
			Time:    time.Now(),
		})
		m.processing = false
		m.execTimer.Active = false
	}

	return m, tiCmd
}

func (m *ModernChatModel) addMessage(msg ChatMessage) {
	m.messages = append(m.messages, msg)
}

func (m *ModernChatModel) startTicker() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickerMsg{}
	})
}

func (m ModernChatModel) processUserInput(input string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		
		// Start processing and send immediate start message
		go func() {
			streamCallback := func(chunk agent.StreamChunk) {
				// Send each chunk immediately as it arrives
				var content string
				switch chunk.Type {
				case "status":
					if chunk.Content != "" {
						content = "📋 " + chunk.Content + "\n"
					}
				case "iteration":
					if chunk.Content != "" {
						content = "🔄 " + chunk.Content + "\n"
					}
				case "tool_start":
					if chunk.Content != "" {
						content = "🛠️ " + chunk.Content + "\n"
					}
				case "tool_result":
					if chunk.Content != "" {
						content = "📋 " + chunk.Content + "\n"
					}
				case "tool_error":
					if chunk.Content != "" {
						content = "❌ " + chunk.Content + "\n"
					}
				case "final_answer":
					if chunk.Content != "" {
						content = "✨ " + chunk.Content + "\n"
					}
				case "llm_content":
					content = chunk.Content
				case "complete":
					if chunk.Content != "" {
						content = "✅ " + chunk.Content + "\n"
					}
				case "max_iterations":
					if chunk.Content != "" {
						content = "⚠️ " + chunk.Content + "\n"
					}
				case "context_management":
					if chunk.Content != "" {
						content = "🧠 " + chunk.Content + "\n"
					}
				case "error":
					// Error will be handled separately
				}
				
				// Send streaming update immediately
				if content != "" {
					m.program.Send(streamChunkMsg{content: content})
				}
			}

			err := m.agent.ProcessMessageStream(ctx, input, m.config.GetConfig(), streamCallback)
			if err != nil {
				m.program.Send(errorOccurredMsg{err: err})
			} else {
				m.program.Send(streamCompleteMsg{})
			}
		}()

		// Return immediately with processing started message
		return streamStartMsg{input: input}
	}
}

func (m ModernChatModel) View() string {
	if !m.ready {
		return "Initializing Deep Coding Agent..."
	}

	var parts []string

	// Header
	header := headerStyle.Render("🤖 Deep Coding Agent - AI-Powered Coding Assistant")
	parts = append(parts, header, "")

	// Session runtime info (displayed at top)
	if !m.sessionStartTime.IsZero() {
		sessionRuntime := m.formatSessionRuntime()
		copyHint := " • Select text with mouse to copy"
		sessionInfo := sessionTimeStyle.Render(sessionRuntime + copyHint)
		parts = append(parts, sessionInfo, "")
	}

	// Messages content (directly rendered, not in viewport)
	for i, msg := range m.messages {
		if i > 0 {
			parts = append(parts, "") // Single line between messages
		}

		var styledContent string
		switch msg.Type {
		case "user":
			styledContent = userMsgStyle.Render("👤 You: ") + msg.Content
		case "assistant":
			styledContent = assistantMsgStyle.Render("🤖 Alex: ") + msg.Content
		case "system":
			styledContent = systemMsgStyle.Render(msg.Content)
		case "processing":
			styledContent = processingStyle.Render("⚡ " + msg.Content)
		case "error":
			styledContent = errorMsgStyle.Render("❌ " + msg.Content)
		default:
			styledContent = msg.Content
		}

		parts = append(parts, styledContent)
	}

	// Add space before input area
	parts = append(parts, "")

	// Input area
	var inputArea string
	if m.processing {
		inputArea = inputStyle.Render(processingStyle.Render("⚡ Processing... please wait"))
	} else {
		inputArea = inputStyle.Render(m.textarea.View())
	}
	parts = append(parts, inputArea)

	// No footer - keep it clean

	// Join all parts and ensure it fits the screen
	result := lipgloss.JoinVertical(lipgloss.Left, parts...)
	
	// If content is too long for screen, only show recent parts
	if m.height > 0 {
		lines := strings.Split(result, "\n")
		if len(lines) > m.height-2 { // Leave some margin
			// Show last messages that fit on screen
			visibleLines := lines[len(lines)-(m.height-2):]
			result = strings.Join(visibleLines, "\n")
		}
	}

	return result
}

// Run the modern TUI
func runModernTUI(agent *agent.ReactAgent, config *config.Manager) error {
	model := NewModernChatModel(agent, config)

	program := tea.NewProgram(
		&model,
		tea.WithAltScreen(),
		// Removed tea.WithMouseCellMotion() to allow text selection
	)

	// Set the program reference for streaming callbacks
	model.program = program

	_, err := program.Run()
	return err
}
