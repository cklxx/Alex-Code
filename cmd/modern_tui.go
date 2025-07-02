package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"alex/internal/agent"
	"alex/internal/config"
)

// Modern TUI with clean, professional interface
var (
	// Color scheme
	primaryColor   = lipgloss.Color("#7C3AED")
	successColor   = lipgloss.Color("#10B981") 
	warningColor   = lipgloss.Color("#F59E0B")
	errorColor     = lipgloss.Color("#EF4444")
	mutedColor     = lipgloss.Color("#6B7280")
	backgroundColor = lipgloss.Color("#1F2937")
	
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
		BorderForeground(primaryColor).
		Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true)
)

// Message types
type (
	streamResponseMsg struct{ content string }
	processingDoneMsg struct{}
	errorOccurredMsg  struct{ err error }
)

// ModernChatModel represents the clean TUI model
type ModernChatModel struct {
	viewport     viewport.Model
	textarea     textarea.Model
	messages     []ChatMessage
	processing   bool
	agent        *agent.ReactAgent
	config       *config.Manager
	width        int
	height       int
	ready        bool
	currentInput string
}

// ChatMessage represents a chat message with type and content
type ChatMessage struct {
	Type    string // "user", "assistant", "system", "processing", "error"
	Content string
	Time    time.Time
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

	// Configure viewport
	vp := viewport.New(80, 20)

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
		textarea: ta,
		viewport: vp,
		messages: initialMessages,
		agent:    agent,
		config:   config,
		ready:    false,
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

func (m ModernChatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m ModernChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			// Initialize dimensions
			m.viewport = viewport.New(msg.Width, msg.Height-8) // Reserve space for header and input
			m.textarea.SetWidth(msg.Width - 6)
			m.ready = true
			m.updateViewport()
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 8
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
				
				// Start processing
				m.processing = true
				m.addMessage(ChatMessage{
					Type:    "processing",
					Content: "Processing your request...",
					Time:    time.Now(),
				})
				
				return m, m.processUserInput(input)
			}
		}

	case streamResponseMsg:
		// Remove last processing message and add response
		if len(m.messages) > 0 && m.messages[len(m.messages)-1].Type == "processing" {
			m.messages = m.messages[:len(m.messages)-1]
		}
		
		m.addMessage(ChatMessage{
			Type:    "assistant",
			Content: msg.content,
			Time:    time.Now(),
		})
		
		return m, func() tea.Msg { return processingDoneMsg{} }

	case processingDoneMsg:
		m.processing = false

	case errorOccurredMsg:
		// Remove processing message
		if len(m.messages) > 0 && m.messages[len(m.messages)-1].Type == "processing" {
			m.messages = m.messages[:len(m.messages)-1]
		}
		
		m.addMessage(ChatMessage{
			Type:    "error",
			Content: fmt.Sprintf("Error: %v", msg.err),
			Time:    time.Now(),
		})
		m.processing = false
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *ModernChatModel) addMessage(msg ChatMessage) {
	m.messages = append(m.messages, msg)
	m.updateViewport()
}

func (m *ModernChatModel) updateViewport() {
	var content strings.Builder
	
	for i, msg := range m.messages {
		if i > 0 {
			content.WriteString("\n") // Single line between messages
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
		
		content.WriteString(styledContent)
	}
	
	m.viewport.SetContent(content.String())
	m.viewport.GotoBottom()
}

func (m ModernChatModel) processUserInput(input string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var responseBuilder strings.Builder
		
		// Collect all response content
		streamCallback := func(chunk agent.StreamChunk) {
			switch chunk.Type {
			case "final_answer", "llm_content", "content":
				responseBuilder.WriteString(chunk.Content)
			case "tool_result":
				if chunk.Content != "" {
					responseBuilder.WriteString("\n\n📋 " + chunk.Content)
				}
			case "error":
				// Error will be handled separately
			}
		}
		
		err := m.agent.ProcessMessageStream(ctx, input, m.config.GetConfig(), streamCallback)
		if err != nil {
			return errorOccurredMsg{err: err}
		}
		
		response := strings.TrimSpace(responseBuilder.String())
		if response == "" {
			response = "I processed your request, but didn't generate a visible response."
		}
		
		return streamResponseMsg{content: response}
	}
}

func (m ModernChatModel) View() string {
	if !m.ready {
		return "Initializing Deep Coding Agent..."
	}

	// Header
	header := headerStyle.Render("🤖 Deep Coding Agent - AI-Powered Coding Assistant")
	
	// Main content
	content := m.viewport.View()
	
	// Input area
	var inputArea string
	if m.processing {
		inputArea = inputStyle.Render(processingStyle.Render("⚡ Processing... please wait"))
	} else {
		inputArea = inputStyle.Render(m.textarea.View())
	}
	
	// Footer
	footer := footerStyle.Render("Enter: Send message • Ctrl+C: Exit")
	
	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		"", // Single spacer
		inputArea,
		footer,
	)
}

// Run the modern TUI
func runModernTUI(agent *agent.ReactAgent, config *config.Manager) error {
	model := NewModernChatModel(agent, config)
	
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	
	_, err := program.Run()
	return err
}