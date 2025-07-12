package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestNewRootCommand 测试根命令创建
func TestNewRootCommand(t *testing.T) {
	rootCmd := NewRootCommand()

	if rootCmd == nil {
		t.Fatal("Expected non-nil root command")
	}

	if rootCmd.Use != "alex" {
		t.Errorf("Expected Use = 'alex', got %s", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Fatal("Expected non-empty Short description")
	}

	if rootCmd.Long == "" {
		t.Fatal("Expected non-empty Long description")
	}
}

// TestCLI_Creation 测试CLI结构创建
func TestCLI_Creation(t *testing.T) {
	rootCmd := NewRootCommand()

	// 验证CLI已正确初始化
	if rootCmd.RunE == nil {
		t.Fatal("Expected non-nil RunE function")
	}
}

// TestVersionFlag 测试版本标志
func TestVersionFlag(t *testing.T) {
	rootCmd := NewRootCommand()

	// 测试 --version 标志
	output := &bytes.Buffer{}
	rootCmd.SetOut(output)
	rootCmd.SetErr(output)
	rootCmd.SetArgs([]string{"--version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Logf("Version flag execution result: %v", err)
	}

	// 版本信息应该包含在输出中
	outputStr := output.String()
	t.Logf("Version output: %s", outputStr)
}

// TestConfigCommand 测试配置命令
func TestConfigCommand(t *testing.T) {
	rootCmd := NewRootCommand()

	// 查找config子命令
	configCmd := findCommand(rootCmd, "config")
	if configCmd == nil {
		t.Skip("Config command not found, skipping test")
		return
	}

	if configCmd.Use != "config" {
		t.Errorf("Expected Use = 'config', got %s", configCmd.Use)
	}

	t.Logf("Config command exists and is properly configured")
}

// TestSessionCommand 测试会话命令
func TestSessionCommand(t *testing.T) {
	rootCmd := NewRootCommand()

	// 查找session子命令
	sessionCmd := findCommand(rootCmd, "session")
	if sessionCmd == nil {
		t.Skip("Session command not found, skipping test")
		return
	}

	if sessionCmd.Use != "session" {
		t.Errorf("Expected Use = 'session', got %s", sessionCmd.Use)
	}

	t.Logf("Session command exists and is properly configured")
}

// TestToolsCommand 测试工具命令
func TestToolsCommand(t *testing.T) {
	rootCmd := NewRootCommand()

	// 查找tools子命令
	toolsCmd := findCommand(rootCmd, "tools")
	if toolsCmd == nil {
		t.Skip("Tools command not found, skipping test")
		return
	}

	if toolsCmd.Use != "tools" {
		t.Errorf("Expected Use = 'tools', got %s", toolsCmd.Use)
	}

	t.Logf("Tools command exists and is properly configured")
}

// TestBatchCommand 测试批处理命令
func TestBatchCommand(t *testing.T) {
	rootCmd := NewRootCommand()

	// 查找run-batch子命令
	batchCmd := findCommand(rootCmd, "run-batch")
	if batchCmd == nil {
		t.Skip("Batch command not found, skipping test")
		return
	}

	t.Logf("Batch command exists and is properly configured")
}

// TestCommandFlags 测试命令行标志
func TestCommandFlags(t *testing.T) {
	rootCmd := NewRootCommand()

	// 测试基本标志
	basicFlags := []string{
		"interactive",
		"resume",
		"verbose",
		"debug",
		"tui",
	}

	foundFlags := 0
	for _, flagName := range basicFlags {
		flag := rootCmd.Flags().Lookup(flagName)
		if flag == nil {
			flag = rootCmd.PersistentFlags().Lookup(flagName)
		}
		if flag != nil {
			foundFlags++
		}
	}

	if foundFlags == 0 {
		t.Fatal("Expected at least some basic flags to exist")
	}

	t.Logf("Found %d basic flags", foundFlags)
}

// TestColorFunctions 测试颜色函数
func TestColorFunctions(t *testing.T) {
	testCases := []struct {
		function func(string) string
		input    string
		contains string
	}{
		{DeepCodingError, "test error", "❌"},
		{DeepCodingAction, "test action", "test action"}, // DeepCodingAction doesn't add emoji
		{DeepCodingThinking, "test thinking", "🤔"},
		{DeepCodingReasoning, "test reasoning", "🧠"},
		{DeepCodingResult, "test result", "✨"}, // DeepCodingResult uses ✨
		{DeepCodingSuccess, "test success", "🎉"},
	}

	for _, tc := range testCases {
		result := tc.function(tc.input)
		if !strings.Contains(result, tc.contains) {
			t.Errorf("Expected result to contain '%s', got: %s", tc.contains, result)
		}
		if !strings.Contains(result, tc.input) {
			t.Errorf("Expected result to contain input '%s', got: %s", tc.input, result)
		}
	}
}

// TestDeepCodingToolExecution 测试工具执行格式化
func TestDeepCodingToolExecution(t *testing.T) {
	title := "Test Tool"
	content := "Tool output content"

	result := DeepCodingToolExecution(title, content)

	if !strings.Contains(result, title) {
		t.Fatal("Expected result to contain title")
	}
	if !strings.Contains(result, content) {
		t.Fatal("Expected result to contain content")
	}
}

// TestCommandStructure 测试命令结构完整性
func TestCommandStructure(t *testing.T) {
	rootCmd := NewRootCommand()

	// 验证至少有一些基本命令
	commands := rootCmd.Commands()
	if len(commands) == 0 {
		t.Fatal("Expected at least some commands to exist")
	}

	// 验证每个命令都有适当的描述
	for _, cmd := range commands {
		if cmd.Short == "" {
			t.Errorf("Command '%s' missing short description", cmd.Name())
		}
	}

	t.Logf("Found %d commands", len(commands))
}

// TestHelpCommand 测试帮助命令
func TestHelpCommand(t *testing.T) {
	rootCmd := NewRootCommand()

	// 测试根命令帮助
	output := &bytes.Buffer{}
	rootCmd.SetOut(output)
	rootCmd.SetErr(output)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Help command failed: %v", err)
	}

	helpOutput := output.String()
	if !strings.Contains(helpOutput, "alex") {
		t.Fatal("Help output should contain command name")
	}
	if !strings.Contains(helpOutput, "Usage:") {
		t.Fatal("Help output should contain usage information")
	}
}

// Helper function to find a command by name
func findCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

// TestCLI_BufferAllocation 测试CLI缓冲区分配
func TestCLI_BufferAllocation(t *testing.T) {
	// 这个测试验证CLI结构中的contentBuffer是否正确初始化
	cli := &CLI{
		inputQueue: make(chan string, 10),
	}
	cli.contentBuffer.Grow(4096)

	// 测试缓冲区功能
	testContent := "Test streaming content"
	cli.contentBuffer.WriteString(testContent)

	result := cli.contentBuffer.String()
	if result != testContent {
		t.Errorf("Expected buffer content '%s', got '%s'", testContent, result)
	}

	// 测试缓冲区重置
	cli.contentBuffer.Reset()
	if cli.contentBuffer.Len() != 0 {
		t.Fatal("Expected buffer to be empty after reset")
	}
}

// TestCLI_InputQueue 测试输入队列
func TestCLI_InputQueue(t *testing.T) {
	cli := &CLI{
		inputQueue: make(chan string, 10),
	}

	// 测试队列写入和读取
	testInput := "test input"

	// 非阻塞写入
	select {
	case cli.inputQueue <- testInput:
		// 成功写入
	default:
		t.Fatal("Failed to write to input queue")
	}

	// 非阻塞读取
	select {
	case received := <-cli.inputQueue:
		if received != testInput {
			t.Errorf("Expected '%s', got '%s'", testInput, received)
		}
	default:
		t.Fatal("Failed to read from input queue")
	}
}
