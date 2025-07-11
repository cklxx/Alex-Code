# CLI 底部输入框设计与实现

> 实现输入框保持在CLI界面最下方，支持打断和插入信息功能的技术方案

## 核心需求分析

1. **固定底部输入框**：输入框始终在屏幕最底部
2. **滚动内容区域**：上方内容可以滚动，不影响输入框
3. **打断功能**：能够中断正在进行的操作
4. **信息插入**：在不影响输入框的情况下插入新信息

## 技术方案对比

### 方案一：Bubble Tea + 布局管理 ⭐⭐⭐⭐⭐

**最推荐的现代化方案**

```go
package main

import (
    "context"
    "fmt"
    "strings"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type model struct {
    width          int
    height         int
    content        []string
    input          string
    cursor         int
    scrollOffset   int
    streaming      bool
    interrupted    bool
    ctx            context.Context
    cancel         context.CancelFunc
}

// 消息类型
type streamMsg string
type interruptMsg struct{}
type insertMsg string
type windowSizeMsg tea.WindowSizeMsg

// 样式定义
var (
    contentStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("62")).
        Padding(1, 2)

    inputStyle = lipgloss.NewStyle().
        Border(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("205")).
        Padding(0, 1)

    statusStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("62")).
        Foreground(lipgloss.Color("230")).
        Padding(0, 1)
)

func initialModel() model {
    ctx, cancel := context.WithCancel(context.Background())
    return model{
        content:    []string{"欢迎使用智能CLI助手", "输入您的问题..."},
        input:      "",
        cursor:     0,
        streaming:  false,
        ctx:        ctx,
        cancel:     cancel,
    }
}

func (m model) Init() tea.Cmd {
    return tea.Batch(
        tea.EnterAltScreen,
        startStreamingData(),
    )
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        return m, nil

    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            // 如果正在流式输出，先中断
            if m.streaming {
                m.interrupted = true
                m.streaming = false
                m.cancel() // 取消流式操作
                m.content = append(m.content, "⚠️ 操作已中断")
                
                // 重新创建context
                ctx, cancel := context.WithCancel(context.Background())
                m.ctx = ctx
                m.cancel = cancel
                return m, nil
            }
            return m, tea.Quit

        case "enter":
            if m.input != "" {
                // 添加用户输入到内容
                m.content = append(m.content, fmt.Sprintf("👤 %s", m.input))
                
                // 开始处理
                query := m.input
                m.input = ""
                m.streaming = true
                m.interrupted = false
                
                // 重新创建context用于新的操作
                ctx, cancel := context.WithCancel(context.Background())
                m.ctx = ctx
                m.cancel = cancel
                
                return m, processQuery(query, ctx)
            }

        case "backspace":
            if len(m.input) > 0 {
                m.input = m.input[:len(m.input)-1]
            }

        case "up":
            if m.scrollOffset > 0 {
                m.scrollOffset--
            }

        case "down":
            maxScroll := len(m.content) - (m.height - 4)
            if maxScroll > 0 && m.scrollOffset < maxScroll {
                m.scrollOffset++
            }

        default:
            m.input += msg.String()
        }

    case streamMsg:
        if !m.interrupted {
            m.content = append(m.content, fmt.Sprintf("🤖 %s", string(msg)))
            // 自动滚动到最新内容
            maxScroll := len(m.content) - (m.height - 4)
            if maxScroll > 0 {
                m.scrollOffset = maxScroll
            }
            return m, waitForNextStream(m.ctx)
        }

    case insertMsg:
        // 插入系统消息
        m.content = append(m.content, fmt.Sprintf("📢 %s", string(msg)))
        maxScroll := len(m.content) - (m.height - 4)
        if maxScroll > 0 {
            m.scrollOffset = maxScroll
        }

    case interruptMsg:
        m.streaming = false
        m.content = append(m.content, "⚠️ 流式输出已停止")
    }

    return m, tea.Batch(cmds...)
}

func (m model) View() string {
    if m.height == 0 {
        return "初始化中..."
    }

    // 计算各部分高度
    contentHeight := m.height - 4 // 为输入框和边框留出空间
    
    // 渲染内容区域
    visibleContent := m.getVisibleContent(contentHeight)
    contentArea := contentStyle.
        Width(m.width - 4).
        Height(contentHeight).
        Render(strings.Join(visibleContent, "\n"))

    // 渲染状态栏
    status := ""
    if m.streaming {
        status = "🔄 正在处理... (Ctrl+C 中断)"
    } else {
        status = "💬 输入您的问题 (↑↓ 滚动, Ctrl+C 退出)"
    }
    statusBar := statusStyle.
        Width(m.width).
        Render(status)

    // 渲染输入框
    inputPrompt := "➤ "
    inputArea := inputStyle.
        Width(m.width - len(inputPrompt) - 2).
        Render(m.input + "█") // 简单的光标效果

    // 组合布局 - 固定底部输入框
    return lipgloss.JoinVertical(
        lipgloss.Left,
        contentArea,
        statusBar,
        inputPrompt+inputArea,
    )
}

// 获取可见内容（支持滚动）
func (m model) getVisibleContent(maxLines int) []string {
    if len(m.content) <= maxLines {
        return m.content
    }

    start := m.scrollOffset
    end := start + maxLines

    if end > len(m.content) {
        end = len(m.content)
        start = end - maxLines
        if start < 0 {
            start = 0
        }
    }

    return m.content[start:end]
}

// 模拟流式数据处理
func processQuery(query string, ctx context.Context) tea.Cmd {
    return func() tea.Msg {
        select {
        case <-ctx.Done():
            return interruptMsg{}
        case <-time.After(500 * time.Millisecond):
            return streamMsg(fmt.Sprintf("正在分析: %s", query))
        }
    }
}

func waitForNextStream(ctx context.Context) tea.Cmd {
    return func() tea.Msg {
        select {
        case <-ctx.Done():
            return interruptMsg{}
        case <-time.After(1 * time.Second):
            responses := []string{
                "找到相关信息...",
                "正在生成回答...",
                "回答生成完成！",
            }
            return streamMsg(responses[time.Now().Second()%len(responses)])
        }
    }
}

func startStreamingData() tea.Cmd {
    return func() tea.Msg {
        time.Sleep(2 * time.Second)
        return insertMsg("系统已就绪")
    }
}

func main() {
    p := tea.NewProgram(
        initialModel(),
        tea.WithAltScreen(),
        tea.WithMouseCellMotion(),
    )

    if _, err := p.Run(); err != nil {
        fmt.Printf("错误: %v", err)
    }
}
```

### 方案二：tview 实现 ⭐⭐⭐⭐

**成熟稳定的组件化方案**

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/gdamore/tcell/v2"
    "github.com/rivo/tview"
)

type ChatApp struct {
    app         *tview.Application
    grid        *tview.Grid
    textView    *tview.TextView
    inputField  *tview.InputField
    statusBar   *tview.TextView
    streaming   bool
    ctx         context.Context
    cancel      context.CancelFunc
}

func NewChatApp() *ChatApp {
    ctx, cancel := context.WithCancel(context.Background())
    
    app := &ChatApp{
        app:    tview.NewApplication(),
        ctx:    ctx,
        cancel: cancel,
    }
    
    app.setupUI()
    return app
}

func (c *ChatApp) setupUI() {
    // 创建组件
    c.textView = tview.NewTextView().
        SetDynamicColors(true).
        SetScrollable(true).
        SetWrap(true)
    
    c.textView.SetBorder(true).
        SetTitle(" 对话记录 ").
        SetTitleAlign(tview.AlignLeft)

    c.inputField = tview.NewInputField().
        SetLabel("➤ ").
        SetFieldBackgroundColor(tcell.ColorBlack).
        SetFieldTextColor(tcell.ColorWhite)

    c.statusBar = tview.NewTextView().
        SetTextAlign(tview.AlignCenter).
        SetDynamicColors(true)
    
    c.updateStatus("准备就绪 - 输入您的问题")

    // 设置输入框回调
    c.inputField.SetDoneFunc(func(key tcell.Key) {
        if key == tcell.KeyEnter {
            c.handleInput()
        }
    })

    // 设置全局按键处理
    c.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        switch event.Key() {
        case tcell.KeyCtrlC:
            if c.streaming {
                c.interruptStreaming()
                return nil
            }
            c.app.Stop()
            return nil
        }
        return event
    })

    // 创建网格布局
    c.grid = tview.NewGrid().
        SetRows(0, 1, 1). // 内容区域、状态栏、输入框
        SetColumns(0).
        AddItem(c.textView, 0, 0, 1, 1, 0, 0, false).
        AddItem(c.statusBar, 1, 0, 1, 1, 0, 0, false).
        AddItem(c.inputField, 2, 0, 1, 1, 0, 0, true)

    c.app.SetRoot(c.grid, true).SetFocus(c.inputField)
}

func (c *ChatApp) handleInput() {
    text := c.inputField.GetText()
    if text == "" {
        return
    }

    // 显示用户输入
    c.addMessage(fmt.Sprintf("[blue]👤 用户:[white] %s", text))
    
    // 清空输入框
    c.inputField.SetText("")
    
    // 开始流式处理
    c.streaming = true
    c.updateStatus("[yellow]🔄 正在处理... (Ctrl+C 中断)")
    
    // 重新创建 context
    c.cancel()
    c.ctx, c.cancel = context.WithCancel(context.Background())
    
    go c.processStreamingResponse(text)
}

func (c *ChatApp) processStreamingResponse(query string) {
    responses := []string{
        "正在分析您的问题...",
        "搜索相关信息...",
        "生成回答中...",
        fmt.Sprintf("针对 '%s' 的回答已生成完成", query),
    }

    for i, response := range responses {
        select {
        case <-c.ctx.Done():
            c.app.QueueUpdateDraw(func() {
                c.addMessage("[red]⚠️ 操作已中断")
                c.streaming = false
                c.updateStatus("就绪 - 输入您的问题")
            })
            return
        case <-time.After(1 * time.Second):
            c.app.QueueUpdateDraw(func() {
                c.addMessage(fmt.Sprintf("[green]🤖 助手:[white] %s", response))
                
                if i == len(responses)-1 {
                    c.streaming = false
                    c.updateStatus("完成 - 输入下一个问题")
                }
            })
        }
    }
}

func (c *ChatApp) interruptStreaming() {
    if c.streaming {
        c.cancel()
        c.streaming = false
        c.addMessage("[red]⚠️ 流式输出已中断")
        c.updateStatus("已中断 - 输入新问题")
        
        // 重新创建 context
        c.ctx, c.cancel = context.WithCancel(context.Background())
    }
}

func (c *ChatApp) addMessage(message string) {
    fmt.Fprintf(c.textView, "%s\n", message)
    c.textView.ScrollToEnd()
}

func (c *ChatApp) updateStatus(status string) {
    c.statusBar.SetText(status)
}

func (c *ChatApp) insertSystemMessage(message string) {
    c.app.QueueUpdateDraw(func() {
        c.addMessage(fmt.Sprintf("[yellow]📢 系统:[white] %s", message))
    })
}

func (c *ChatApp) Run() error {
    // 启动后台任务示例
    go func() {
        time.Sleep(3 * time.Second)
        c.insertSystemMessage("系统监控已启动")
    }()

    return c.app.Run()
}

func main() {
    app := NewChatApp()
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

### 方案三：原生终端控制 ⭐⭐⭐

**轻量级但复杂的方案**

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    "golang.org/x/term"
)

type TerminalChat struct {
    width       int
    height      int
    content     []string
    input       string
    scrollPos   int
    streaming   bool
    ctx         context.Context
    cancel      context.CancelFunc
    oldState    *term.State
}

func NewTerminalChat() (*TerminalChat, error) {
    // 获取终端尺寸
    width, height, err := term.GetSize(int(os.Stdin.Fd()))
    if err != nil {
        return nil, err
    }

    // 设置原始模式
    oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        return nil, err
    }

    ctx, cancel := context.WithCancel(context.Background())

    return &TerminalChat{
        width:    width,
        height:   height,
        content:  []string{"欢迎使用终端聊天", "输入您的问题..."},
        oldState: oldState,
        ctx:      ctx,
        cancel:   cancel,
    }, nil
}

func (t *TerminalChat) cleanup() {
    term.Restore(int(os.Stdin.Fd()), t.oldState)
    fmt.Print("\033[?25h") // 显示光标
    fmt.Print("\033[2J\033[H") // 清屏
}

func (t *TerminalChat) draw() {
    // 清屏并移动到左上角
    fmt.Print("\033[2J\033[H")

    // 计算内容区域高度
    contentHeight := t.height - 3 // 留出输入框和状态栏空间

    // 显示内容
    startIdx := 0
    if len(t.content) > contentHeight {
        startIdx = len(t.content) - contentHeight + t.scrollPos
        if startIdx < 0 {
            startIdx = 0
        }
    }

    for i := 0; i < contentHeight; i++ {
        idx := startIdx + i
        if idx < len(t.content) {
            fmt.Printf("│ %-*s │\n", t.width-4, truncate(t.content[idx], t.width-4))
        } else {
            fmt.Printf("│ %-*s │\n", t.width-4, "")
        }
    }

    // 分隔线
    fmt.Printf("├%s┤\n", strings.Repeat("─", t.width-2))

    // 状态栏
    status := "准备就绪"
    if t.streaming {
        status = "🔄 处理中... (Ctrl+C 中断)"
    }
    fmt.Printf("│ %-*s │\n", t.width-4, status)

    // 输入行
    fmt.Printf("➤ %s█", t.input)
}

func (t *TerminalChat) addContent(message string) {
    t.content = append(t.content, message)
    if len(t.content) > 1000 { // 限制历史记录
        t.content = t.content[100:]
    }
}

func (t *TerminalChat) processInput() {
    if t.input == "" {
        return
    }

    query := t.input
    t.addContent(fmt.Sprintf("👤 %s", query))
    t.input = ""
    t.streaming = true

    // 重新创建 context
    t.cancel()
    t.ctx, t.cancel = context.WithCancel(context.Background())

    go t.streamResponse(query)
}

func (t *TerminalChat) streamResponse(query string) {
    responses := []string{
        "正在分析问题...",
        "搜索相关信息...",
        fmt.Sprintf("关于 '%s' 的回答:", query),
        "回答生成完成",
    }

    for _, response := range responses {
        select {
        case <-t.ctx.Done():
            t.addContent("⚠️ 操作已中断")
            t.streaming = false
            return
        case <-time.After(1 * time.Second):
            t.addContent(fmt.Sprintf("🤖 %s", response))
            t.draw()
        }
    }

    t.streaming = false
}

func (t *TerminalChat) Run() error {
    defer t.cleanup()

    // 设置信号处理
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

    // 启动输入处理
    inputCh := make(chan byte, 1)
    go func() {
        reader := bufio.NewReader(os.Stdin)
        for {
            b, err := reader.ReadByte()
            if err != nil {
                return
            }
            inputCh <- b
        }
    }()

    t.draw()

    for {
        select {
        case <-sigCh:
            if t.streaming {
                t.cancel()
                t.streaming = false
                t.addContent("⚠️ 操作已中断")
                t.draw()
                continue
            }
            return nil

        case b := <-inputCh:
            switch b {
            case 3: // Ctrl+C
                if t.streaming {
                    t.cancel()
                    t.streaming = false
                    t.addContent("⚠️ 操作已中断")
                } else {
                    return nil
                }
            case 13: // Enter
                t.processInput()
            case 127: // Backspace
                if len(t.input) > 0 {
                    t.input = t.input[:len(t.input)-1]
                }
            default:
                if b >= 32 && b <= 126 { // 可打印字符
                    t.input += string(b)
                }
            }
            t.draw()
        }
    }
}

func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}

func main() {
    chat, err := NewTerminalChat()
    if err != nil {
        fmt.Printf("初始化失败: %v\n", err)
        return
    }

    if err := chat.Run(); err != nil {
        fmt.Printf("运行错误: %v\n", err)
    }
}
```

## 集成到现有项目

将上述方案集成到您的 Deep Coding Agent 项目：

```go
// 在 cmd/main.go 中修改 runInteractive 函数
func runInteractive(agentInstance *agent.ReactAgent, configManager *config.Manager, cliConfig *CLIConfig, verbose, debug bool) {
    // 使用 Bubble Tea 替代原有的 bufio.Scanner 方式
    
    model := ChatModel{
        agent:         agentInstance,
        config:        configManager,
        cliConfig:     cliConfig,
        verbose:       verbose,
        debug:        debug,
        content:      []string{"🤖 Deep Coding Agent " + version, "📂 " + getCurrentDir()},
        input:        "",
        streaming:    false,
    }
    
    p := tea.NewProgram(
        model,
        tea.WithAltScreen(),
        tea.WithMouseCellMotion(),
    )
    
    if _, err := p.Run(); err != nil {
        fmt.Printf("TUI 错误: %v\n", err)
        os.Exit(1)
    }
}
```

## 推荐方案

**强烈推荐使用方案一 (Bubble Tea)**，原因：

1. **现代化设计**：事件驱动，符合您项目的 ReAct 架构
2. **易于维护**：清晰的代码结构
3. **功能完整**：天然支持打断、插入、滚动
4. **性能优秀**：与您项目的性能目标一致
5. **生态丰富**：可以配合 Huh、Lip Gloss 等库

这样可以实现类似您图片中展示的效果：输入框固定在底部，内容在上方滚动，支持实时打断和信息插入。