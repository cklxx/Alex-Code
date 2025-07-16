package main

import (
	"strings"
	"testing"
)

// TestNewInitCommand 测试初始化命令创建
func TestNewInitCommand(t *testing.T) {
	rootCmd := NewRootCommand()

	// 查找 init 子命令
	initCmd := findCommand(rootCmd, "init")
	if initCmd == nil {
		t.Fatal("Expected init command to exist")
	}

	if initCmd.Use != "init" {
		t.Errorf("Expected Use = 'init', got %s", initCmd.Use)
	}

	if initCmd.Short == "" {
		t.Error("Expected non-empty Short description")
	}

	if initCmd.Long == "" {
		t.Error("Expected non-empty Long description")
	}
}

// TestInitCommandFlags 测试初始化命令标志
func TestInitCommandFlags(t *testing.T) {
	rootCmd := NewRootCommand()
	initCmd := findCommand(rootCmd, "init")

	if initCmd == nil {
		t.Skip("Init command not found, skipping flag tests")
		return
	}

	// 测试输出文件标志
	outputFlag := initCmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Error("Expected --output flag to exist")
	}

	// 测试项目名称标志
	projectFlag := initCmd.Flags().Lookup("project")
	if projectFlag == nil {
		t.Error("Expected --project flag to exist")
	}

	// 测试强制覆盖标志
	forceFlag := initCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Expected --force flag to exist")
	}
}

// TestBuildProjectAnalysisPrompt 测试项目分析提示构建
func TestBuildProjectAnalysisPrompt(t *testing.T) {
	projectName := "test-project"
	workDir := "/tmp/test"
	outputFile := "TEST.md"

	prompt := buildProjectAnalysisPrompt(projectName, workDir, outputFile)

	// 验证提示包含必要信息
	if !strings.Contains(prompt, projectName) {
		t.Error("Prompt should contain project name")
	}

	if !strings.Contains(prompt, workDir) {
		t.Error("Prompt should contain working directory")
	}

	if !strings.Contains(prompt, outputFile) {
		t.Error("Prompt should contain output file")
	}

	// 验证包含关键指令
	expectedKeywords := []string{
		"project analyst",
		"provided template",
		"template",
		"file_search",
		"codebase_search",
		"variables",
	}

	for _, keyword := range expectedKeywords {
		if !strings.Contains(prompt, keyword) {
			t.Errorf("Prompt should contain keyword: %s", keyword)
		}
	}
}

// TestFileExists 测试文件存在检查函数
func TestFileExists(t *testing.T) {
	// 测试存在的文件
	if !fileExists("cobra_init_test.go") {
		t.Error("Current test file should exist")
	}

	// 测试不存在的文件
	if fileExists("non-existent-file.xyz") {
		t.Error("Non-existent file should not exist")
	}
}

// TestInitCommandHelp 测试初始化命令帮助信息
func TestInitCommandHelp(t *testing.T) {
	rootCmd := NewRootCommand()
	initCmd := findCommand(rootCmd, "init")

	if initCmd == nil {
		t.Skip("Init command not found, skipping help test")
		return
	}

	// 验证帮助信息包含关键内容
	help := initCmd.Long
	expectedContent := []string{
		"分析项目",
		"生成完整的项目文档",
		"alex init",
		"--output",
		"--force",
	}

	for _, content := range expectedContent {
		if !strings.Contains(help, content) {
			t.Errorf("Help should contain: %s", content)
		}
	}
}

// TestPromptLoaderIntegration 测试 prompt loader 集成
func TestPromptLoaderIntegration(t *testing.T) {
	projectName := "test-project"
	workDir := "/tmp/test"
	outputFile := "TEST.md"

	prompt := buildProjectAnalysisPrompt(projectName, workDir, outputFile)

	// 验证 prompt 包含模板内容或 fallback 内容
	templateLoaded := strings.Contains(prompt, "{{ProjectName}}")
	fallbackUsed := strings.Contains(prompt, "Project Overview and Goals")

	if !templateLoaded && !fallbackUsed {
		t.Error("Prompt should either contain template variables or fallback content")
	}

	// 验证基本提示结构存在
	expectedStructure := []string{
		"project analyst",
		"Step 1",
		"Step 2",
		"Step 3",
		"Take Action Immediately",
	}

	for _, structure := range expectedStructure {
		if !strings.Contains(prompt, structure) {
			t.Errorf("Prompt should contain structure element: %s", structure)
		}
	}

	t.Logf("Prompt loader integration test passed. Template loaded: %v, Fallback used: %v",
		templateLoaded, fallbackUsed)
}
