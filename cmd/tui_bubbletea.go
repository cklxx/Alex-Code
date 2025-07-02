package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// 简洁的TUI模型，完全匹配图片样式
type BubbleTeaModel struct {
	textarea   textarea.Model
	viewport   viewport.Model
	messages   []string
	streaming  bool
	startTime  time.Time
	tokensUsed int
	width      int
	height     int
	ready      bool
}

var (
	// 极简样式 - 完全透明无边框
	inputStyle = lipgloss.NewStyle()
	
	promptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))
	
	streamingStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))
	
	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Italic(true)
)

func initialBubbleTeaModel() BubbleTeaModel {
	ta := textarea.New()
	ta.Placeholder = ""
	ta.Focus()
	ta.CharLimit = 2000
	ta.SetWidth(80)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	
	// 完全透明的样式，去除所有视觉元素
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
	ta.FocusedStyle.Text = lipgloss.NewStyle()
	ta.BlurredStyle.Text = lipgloss.NewStyle()
	ta.FocusedStyle.Prompt = lipgloss.NewStyle()
	ta.BlurredStyle.Prompt = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLineNumber = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLineNumber = lipgloss.NewStyle()
	ta.FocusedStyle.EndOfBuffer = lipgloss.NewStyle()
	ta.BlurredStyle.EndOfBuffer = lipgloss.NewStyle()
	ta.FocusedStyle.LineNumber = lipgloss.NewStyle()
	ta.BlurredStyle.LineNumber = lipgloss.NewStyle()
	
	// 设置光标模式 - 使用静态光标而不完全隐藏
	ta.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	
	// 启用多行输入支持
	ta.KeyMap.InsertNewline.SetEnabled(true)

	vp := viewport.New(80, 20)
	vp.SetContent("Deep Coding Agent\n\nReady to help with your coding tasks.\n")

	return BubbleTeaModel{
		textarea:   ta,
		viewport:   vp,
		messages:   []string{},
		streaming:  false,
		startTime:  time.Now(),
		tokensUsed: 0,
		width:      80,
		height:     30,
		ready:      false,
	}
}

func (m BubbleTeaModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m BubbleTeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		taCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			// 初始化时使用合理的高度
			initialHeight := msg.Height - 5  // 为底部区域预留空间
			if initialHeight < 5 {
				initialHeight = 5
			}
			m.viewport = viewport.New(msg.Width, initialHeight)
			m.viewport.SetContent("Deep Coding Agent\n\nReady to help with your coding tasks.\n")
			m.ready = true
		} else {
			// 窗口大小变化时，宽度立即更新，高度在View()中动态计算
			m.viewport.Width = msg.Width
		}

		m.textarea.SetWidth(msg.Width - 2)
		return m, nil

	case tea.KeyMsg:
		switch {
		case msg.String() == "ctrl+c" || msg.String() == "esc":
			if m.streaming {
				m.streaming = false
				m.addMessage("Interrupted")
			} else {
				return m, tea.Quit
			}

		case msg.String() == "enter" && !msg.Alt:
			// Enter发送消息
			if !m.streaming && strings.TrimSpace(m.textarea.Value()) != "" {
				m.processInput()
				return m, m.startStreaming()
			}

		case msg.String() == "shift+enter" || msg.String() == "alt+enter" || msg.String() == "ctrl+enter":
			// 多行输入 - 支持多种组合键以提高兼容性
			if !m.streaming {
				m.textarea.InsertString("\n")
				m.adjustTextareaHeight()
			}

		default:
			if !m.streaming {
				m.textarea, taCmd = m.textarea.Update(msg)
			}
		}

	case streamingMsg:
		if m.streaming {
			m.tokensUsed += msg.tokens
			if msg.content != "" {
				m.addMessage(msg.content)
			}
			if msg.final {
				m.streaming = false
				// 完成工作后重新聚焦输入框
				return m, m.textarea.Focus()
			}
		}
		return m, m.continueStreaming()
	}

	m.viewport, vpCmd = m.viewport.Update(msg)
	return m, tea.Batch(taCmd, vpCmd)
}

func (m BubbleTeaModel) View() string {
	if !m.ready {
		return "\nStarting..."
	}

	// 计算各部分高度
	statusHeight := 0
	if m.streaming {
		statusHeight = 1
	}
	
	inputHeight := m.textarea.Height()
	if !m.streaming {
		inputHeight += 0 // 为 prompt 预留空间
	}
	
	separatorHeight := 1
	bottomAreaHeight := statusHeight + inputHeight + separatorHeight
	
	// 内容区域 - 动态高度，可滚动
	contentHeight := m.height - bottomAreaHeight - 2 // 留出一些边距
	if contentHeight < 5 {
		contentHeight = 5
	}
	
	// 调整viewport大小以适应可用空间
	if m.viewport.Height != contentHeight {
		m.viewport.Height = contentHeight
	}
	
	content := m.viewport.View()

	// 分隔线
	separator := strings.Repeat("─", m.width)

	// 状态栏 - 在输入框正上方
	var statusLine string
	if m.streaming {
		elapsed := time.Since(m.startTime)
		seconds := int(elapsed.Seconds())
		
		var timeStr string
		if seconds < 60 {
			timeStr = fmt.Sprintf("%ds", seconds)
		} else {
			minutes := seconds / 60
			secs := seconds % 60
			timeStr = fmt.Sprintf("%dm%ds", minutes, secs)
		}
		
		statusLine = statusStyle.Render(fmt.Sprintf("✻ Sussing… (%s · ↑ %d tokens · esc to interrupt)", timeStr, m.tokensUsed))
	}

	// 底部输入区域
	var inputArea string
	if !m.streaming {
		// 只在非工作状态显示输入框
		prompt := promptStyle.Render("> ")
		input := m.textarea.View()
		inputArea = prompt + input
	}

	// 垂直布局：内容 + 分隔线 + [状态栏] + [输入框]
	var bottomArea string
	if m.streaming {
		bottomArea = statusLine
	} else {
		bottomArea = inputArea
	}

	return content + "\n" + separator + "\n" + bottomArea
}

func (m *BubbleTeaModel) addMessage(message string) {
	m.messages = append(m.messages, message)

	// 更新viewport内容
	allContent := "Deep Coding Agent\n\n"
	for _, msg := range m.messages {
		allContent += msg + "\n\n"
	}

	m.viewport.SetContent(allContent)
	m.viewport.GotoBottom()
}

func (m *BubbleTeaModel) adjustTextareaHeight() {
	lines := strings.Split(m.textarea.Value(), "\n")
	newHeight := len(lines)
	
	// 设置合理的高度限制
	const minHeight = 1
	const maxHeight = 8
	
	if newHeight < minHeight {
		newHeight = minHeight
	}
	if newHeight > maxHeight {
		newHeight = maxHeight
	}
	
	m.textarea.SetHeight(newHeight)
}

func (m *BubbleTeaModel) processInput() {
	input := strings.TrimSpace(m.textarea.Value())

	// 显示用户输入
	m.addMessage("You: " + input)

	// 清空输入框并重置高度
	m.textarea.SetValue("")
	m.textarea.SetHeight(1)
	
	// 工作时让输入框失去焦点，隐藏光标
	m.textarea.Blur()

	// 开始处理
	m.streaming = true
	m.startTime = time.Now()
	m.tokensUsed = 0
}

// 流式消息类型
type streamingMsg struct {
	content string
	tokens  int
	final   bool
}

func (m BubbleTeaModel) startStreaming() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		responses := []struct {
			content string
			tokens  int
		}{
			{"Assistant: I understand your request.", 25},
			{"Let me analyze this for you...", 15},
			{"Processing your input now.", 12},
			{"Here's my response to your question.", 20},
		}

		responseIndex := int(t.UnixNano()/int64(time.Millisecond*500)) % len(responses)
		response := responses[responseIndex]
		
		return streamingMsg{
			content: response.content,
			tokens:  response.tokens,
			final:   responseIndex == len(responses)-1,
		}
	})
}

func (m BubbleTeaModel) continueStreaming() tea.Cmd {
	if m.streaming {
		return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return streamingMsg{content: "", tokens: 0, final: true}
		})
	}
	return nil
}

// 运行简洁TUI演示
func runBubbleTeaTuiDemo() {
	if !isTerminal() {
		fmt.Println("❌ TUI demo requires a proper terminal environment")
		return
	}

	fmt.Println("🚀 Starting Clean TUI Demo")
	fmt.Println()

	p := tea.NewProgram(
		initialBubbleTeaModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("TUI error: %v\n", err)
		return
	}

	fmt.Println("\n👋 Demo finished!")
}

// isTerminal checks if running in a proper terminal
func isTerminal() bool {
	if !isFileTerminal(os.Stdin) || !isFileTerminal(os.Stdout) || !isFileTerminal(os.Stderr) {
		return false
	}
	return true
}

// isFileTerminal checks if a file is a terminal
func isFileTerminal(f *os.File) bool {
	if f == nil {
		return false
	}

	fd := int(f.Fd())

	if stat, err := f.Stat(); err == nil {
		mode := stat.Mode()
		return mode&os.ModeCharDevice != 0
	}

	return fd >= 0
}