package main

import (
	"context"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"alex/internal/agent"
	"alex/internal/config"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		PaddingLeft(1).
		PaddingRight(1)

	userStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#06B6D4")).
		Bold(true)

	assistantStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981"))

	systemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	processingStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Bold(true)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444")).
		Bold(true)

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	borderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6B7280")).
		Padding(0, 1)
)

// Message types for Bubble Tea
type (
	errMsg                struct{ error }
	processingStartedMsg  struct{ input string }
	processingUpdateMsg   struct{ content string }
	processingFinishedMsg struct{ response string }
	streamChunkMsg        struct{ chunk agent.StreamChunk }
)

// ChatModel represents the main TUI model
type ChatModel struct {
	viewport    viewport.Model
	textarea    textarea.Model
	messages    []string
	processing  bool
	agent       *agent.ReactAgent
	config      *config.Manager
	width       int
	height      int
	ready       bool
}

// InitialChatModel creates a new chat model
func InitialChatModel(agent *agent.ReactAgent, config *config.Manager) ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Type your message here..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 2000
	ta.SetWidth(50)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false) // Enter submits instead of new line

	vp := viewport.New(80, 20)
	vp.SetContent("")

	// Add welcome message
	welcomeMsg := systemStyle.Render("🤖 Deep Coding Agent v2.0 - Powered by Bubble Tea")
	workingDir := systemStyle.Render("📂 Working Directory: " + getCurrentDir())
	helpMsg := helpStyle.Render("💡 Type your questions and press Enter to submit. Ctrl+C to exit.")

	return ChatModel{
		textarea: ta,
		viewport: vp,
		messages: []string{
			welcomeMsg,
			workingDir,
			"",
			helpMsg,
			"",
		},
		agent:   agent,
		config:  config,
		ready:   false,
	}
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

func (m ChatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			// Initialize viewport and textarea dimensions
			headerHeight := 2
			footerHeight := 4
			verticalMarginHeight := headerHeight + footerHeight

			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.textarea.SetWidth(msg.Width - 4)
			m.ready = true

			// Update viewport content
			m.updateViewportContent()
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 6
			m.textarea.SetWidth(msg.Width - 4)
		}

		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if !m.processing {
				input := strings.TrimSpace(m.textarea.Value())
				if input != "" {
					// Add user message
					m.addMessage(userStyle.Render("👤 You: ") + input)
					m.textarea.Reset()
					
					// Start processing
					return m, m.processInput(input)
				}
			}
		}

	case errMsg:
		m.addMessage(errorStyle.Render("❌ Error: " + msg.error.Error()))
		m.processing = false

	case processingStartedMsg:
		m.processing = true
		m.addMessage(processingStyle.Render("✶ Processing..."))

	case streamChunkMsg:
		return m.handleStreamChunk(msg.chunk)

	case processingFinishedMsg:
		m.processing = false
		if msg.response != "" {
			m.addMessage(assistantStyle.Render("🤖 Alex: ") + msg.response)
		}
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *ChatModel) addMessage(message string) {
	m.messages = append(m.messages, message)
	m.updateViewportContent()
}

func (m *ChatModel) updateViewportContent() {
	content := strings.Join(m.messages, "\n")
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m ChatModel) handleStreamChunk(chunk agent.StreamChunk) (tea.Model, tea.Cmd) {
	var content string

	switch chunk.Type {
	case "status":
		content = processingStyle.Render("🔧 " + chunk.Content)
	case "thinking_start":
		content = processingStyle.Render("🤔 Analyzing your request...")
	case "thinking_result":
		content = assistantStyle.Render("💭 " + chunk.Content)
	case "action_start":
		content = processingStyle.Render("⚡ Taking action...")
	case "tool_start":
		content = processingStyle.Render("🛠️ " + chunk.Content)
	case "tool_result":
		content = systemStyle.Render("📋 Tool Result:\n" + chunk.Content)
	case "tool_error":
		content = errorStyle.Render("🚫 " + chunk.Content)
	case "final_answer":
		content = assistantStyle.Render("🤖 Alex: " + chunk.Content)
	case "task_complete":
		content = systemStyle.Render("✅ " + chunk.Content)
	case "llm_content", "content":
		// For streaming content, append to last message if it's from assistant
		if len(m.messages) > 0 && strings.HasPrefix(m.messages[len(m.messages)-1], assistantStyle.Render("🤖 Alex: ")) {
			m.messages[len(m.messages)-1] += chunk.Content
		} else {
			m.addMessage(assistantStyle.Render("🤖 Alex: ") + chunk.Content)
		}
		m.updateViewportContent()
		return m, nil
	case "error":
		content = errorStyle.Render("❌ " + chunk.Content)
	case "complete":
		m.processing = false
		return m, nil
	default:
		// Ignore unknown chunk types
		return m, nil
	}

	if content != "" {
		m.addMessage(content)
	}

	return m, nil
}

func (m ChatModel) processInput(input string) tea.Cmd {
	return func() tea.Msg {
		// For now, use a simplified non-streaming approach
		// We'll process synchronously and return the result
		
		ctx := context.Background()
		var result strings.Builder
		
		streamCallback := func(chunk agent.StreamChunk) {
			switch chunk.Type {
			case "final_answer", "llm_content", "content":
				result.WriteString(chunk.Content)
			case "complete":
				// Processing finished
			}
		}
		
		err := m.agent.ProcessMessageStream(ctx, input, m.config.GetConfig(), streamCallback)
		if err != nil {
			return errMsg{err}
		}
		
		return processingFinishedMsg{response: result.String()}
	}
}


func (m ChatModel) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Header
	header := titleStyle.Render("🤖 Deep Coding Agent - Interactive Chat")
	
	// Main content area (viewport)
	content := m.viewport.View()
	
	// Footer with input area
	var footer string
	if m.processing {
		footer = borderStyle.Render(processingStyle.Render("✶ Processing... (please wait)"))
	} else {
		footer = borderStyle.Render(m.textarea.View())
	}
	
	// Help text
	help := helpStyle.Render("Enter: Send • Ctrl+C: Exit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
		"",
		footer,
		help,
	)
}

// runBubbleTeaTUI starts the new Bubble Tea TUI
func runBubbleTeaTUI(agent *agent.ReactAgent, config *config.Manager) error {
	model := InitialChatModel(agent, config)
	
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	
	_, err := p.Run()
	return err
}