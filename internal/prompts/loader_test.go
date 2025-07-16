package prompts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"alex/pkg/types"
)

// TestLoadProjectMemory 测试项目记忆加载功能
func TestLoadProjectMemory(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test_alex_memory")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	}()

	// 创建 PromptLoader 实例
	loader, err := NewPromptLoader()
	if err != nil {
		t.Fatalf("Failed to create prompt loader: %v", err)
	}

	// Test 1: 没有 ALEX.md 文件时应该返回默认值
	t.Run("No ALEX.md file", func(t *testing.T) {
		memory := loader.loadProjectMemory(tempDir)
		expected := "You are a helpful assistant that can help the user with their tasks."
		if memory != expected {
			t.Errorf("Expected default memory, got: %s", memory)
		}
	})

	// Test 2: 有 ALEX.md 文件时应该返回文件内容
	t.Run("With ALEX.md file", func(t *testing.T) {
		alexContent := "# Test Project\n\nThis is a test project with specific context."
		alexPath := filepath.Join(tempDir, "ALEX.md")

		err := os.WriteFile(alexPath, []byte(alexContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write ALEX.md: %v", err)
		}

		memory := loader.loadProjectMemory(tempDir)
		if memory != alexContent {
			t.Errorf("Expected file content, got: %s", memory)
		}
	})

	// Test 3: 空的 ALEX.md 文件时应该返回默认值
	t.Run("Empty ALEX.md file", func(t *testing.T) {
		alexPath := filepath.Join(tempDir, "ALEX.md")

		err := os.WriteFile(alexPath, []byte("   \n  \t  \n"), 0644)
		if err != nil {
			t.Fatalf("Failed to write empty ALEX.md: %v", err)
		}

		memory := loader.loadProjectMemory(tempDir)
		expected := "You are a helpful assistant that can help the user with their tasks."
		if memory != expected {
			t.Errorf("Expected default memory for empty file, got: %s", memory)
		}
	})

	// Test 4: 空工作目录时应该返回默认值
	t.Run("Empty working directory", func(t *testing.T) {
		memory := loader.loadProjectMemory("")
		expected := "You are a helpful assistant that can help the user with their tasks."
		if memory != expected {
			t.Errorf("Expected default memory for empty workdir, got: %s", memory)
		}
	})
}

// TestGetReActThinkingPromptWithALEX 测试集成的 ReAct 提示生成
func TestGetReActThinkingPromptWithALEX(t *testing.T) {
	// 创建临时目录和 ALEX.md 文件
	tempDir, err := os.MkdirTemp("", "test_react_alex")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	}()

	alexContent := "# My Project\n\nThis project implements advanced AI features for code analysis."
	alexPath := filepath.Join(tempDir, "ALEX.md")
	err = os.WriteFile(alexPath, []byte(alexContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write ALEX.md: %v", err)
	}

	// 创建 PromptLoader 和任务上下文
	loader, err := NewPromptLoader()
	if err != nil {
		t.Fatalf("Failed to create prompt loader: %v", err)
	}

	taskCtx := &types.ReactTaskContext{
		WorkingDir: tempDir,
		Goal:       "Test task",
		LastUpdate: time.Now(),
	}

	// Create a DirectoryInfo with minimal fields needed for testing
	taskCtx.DirectoryInfo = &types.DirectoryContextInfo{
		Description: "Test directory",
	}

	// 生成提示
	prompt, err := loader.GetReActThinkingPrompt(taskCtx)
	if err != nil {
		t.Fatalf("Failed to get ReAct thinking prompt: %v", err)
	}

	// 验证提示包含 ALEX.md 的内容
	if !strings.Contains(prompt, alexContent) {
		t.Errorf("Prompt should contain ALEX.md content, but doesn't")
	}

	// 验证不包含默认记忆内容
	defaultMemory := "You are a helpful assistant that can help the user with their tasks."
	if strings.Contains(prompt, defaultMemory) {
		t.Errorf("Prompt should not contain default memory when ALEX.md exists")
	}
}
