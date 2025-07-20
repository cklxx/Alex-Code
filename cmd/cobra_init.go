package main

import (
	"alex/internal/prompts"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// newInitCommand 创建初始化命令
func newInitCommand(cli *CLI) *cobra.Command {
	var (
		outputFile  string
		projectName string
		force       bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "🚀 初始化项目文档",
		Long: `使用 AI 分析当前项目并生成完整的项目文档。

该命令会：
1. 分析项目结构和代码
2. 生成项目概述、架构说明、使用指南等
3. 将结果写入 ALEX.md 文件

示例:
  alex init                          # 分析当前目录项目
  alex init --output README.md       # 输出到指定文件
  alex init --project MyProject      # 指定项目名称
  alex init --force                  # 强制覆盖已存在的文件`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 初始化CLI
			if err := cli.initialize(cmd); err != nil {
				return fmt.Errorf("failed to initialize CLI: %w", err)
			}

			// 获取当前工作目录
			workDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			// 如果没有指定项目名称，使用目录名
			if projectName == "" {
				projectName = filepath.Base(workDir)
			}

			// 检查输出文件是否存在
			if !force && fileExists(outputFile) {
				return fmt.Errorf("file %s already exists, use --force to overwrite", outputFile)
			}

			fmt.Printf("%s Analyzing project: %s\n", blue("🔍"), projectName)
			fmt.Printf("%s Working directory: %s\n", gray("📁"), workDir)
			fmt.Printf("%s Output file: %s\n", gray("📄"), outputFile)
			fmt.Println()

			// 构建分析提示
			prompt := buildProjectAnalysisPrompt(projectName, workDir, outputFile)

			// 调用 react agent 进行分析
			ctx := context.Background()
			err = cli.agent.ProcessMessageStream(ctx, prompt, cli.config.GetConfig(), cli.deepCodingStreamCallback)
			if err != nil {
				return fmt.Errorf("project analysis failed: %w", err)
			}

			fmt.Printf("\n%s Project documentation generated: %s\n", green("✅"), outputFile)
			return nil
		},
	}

	// 添加标志
	cmd.Flags().StringVarP(&outputFile, "output", "o", "ALEX.md", "输出文件路径")
	cmd.Flags().StringVarP(&projectName, "project", "p", "", "项目名称 (默认使用目录名)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "强制覆盖已存在的文件")

	return cmd
}

// buildProjectAnalysisPrompt builds the project analysis prompt with template content
func buildProjectAnalysisPrompt(projectName, workDir, outputFile string) string {
	// Load the template content using prompt loader
	promptLoader, err := prompts.NewPromptLoader()
	if err != nil {
		// Fallback if prompt loader fails
		return buildFallbackPrompt(projectName, workDir, outputFile)
	}

	template, err := promptLoader.GetPrompt("initial")
	if err != nil {
		// Fallback if template not found
		return buildFallbackPrompt(projectName, workDir, outputFile)
	}

	templateContent := template.Content

	return fmt.Sprintf(`You are a professional project analyst. Your task is to analyze the project "%s" and generate a comprehensive ALEX.md documentation file.

# CRITICAL INSTRUCTIONS:
1. **THIS IS NOT ABOUT CREATING CONVERSATION MEMORY** - You are creating project documentation
2. **OUTPUT MUST BE A MARKDOWN FILE** - Generate actual ALEX.md file content
3. **DO NOT CREATE SHORT-TERM MEMORY** - This is a documentation generation task

# Task Workflow:

## Step 1: Deep Project Analysis
Use the following tools to comprehensively analyze the project:
- file_list to explore project structure  
- file_read to examine key files (README, main.go, config files, core modules)
- grep to search for patterns, features, and technologies used
- Understand the project's purpose, architecture, and key features
- Identify build system, testing approach, and usage patterns
- Analyze the codebase to understand design principles and architecture

## Step 2: Generate ALEX.md Documentation
Using the provided template, replace ALL {{variables}} with actual content:

---
%s
---

### Variable Mapping Instructions:
- {{ProjectName}} → "%s" 
- {{ProjectDescription}} → Brief description of what this project does
- {{BuildCommands}} → Actual build/test commands from Makefile or build scripts
- {{UsageCommands}} → How to run and use the project
- {{CoreComponents}} → List and describe major modules/packages
- {{ToolsSection}} → Title for tools section (e.g., "Built-in Tools")
- {{BuiltinTools}} → List of available tools/features
- {{SecurityFeatures}} → Security measures and protections
- {{PerformanceMetrics}} → Performance characteristics
- {{DesignPhilosophy}} → Core design principles
- {{NamingGuidelines}} → Code naming conventions
- {{ArchitecturalPrinciples}} → Key architectural decisions
- {{CurrentStatus}} → Current development status
- {{TestingInstructions}} → How to run tests

## Step 3: Write the ALEX.md File
Use file_update or file_write to create the file "%s" with:
- Complete markdown content with all variables replaced
- Professional documentation quality
- Clear structure and formatting
- Practical usage examples
- Comprehensive project insights

# CRITICAL REQUIREMENTS:
1. **GENERATE ACTUAL FILE** - Must create "%s" file with documentation content
2. **NO CONVERSATION MEMORY** - This is pure documentation generation
3. **REPLACE ALL VARIABLES** - Every {{variable}} must be filled with real content
4. **PROFESSIONAL QUALITY** - Documentation should be comprehensive and useful

Start analysis and file generation immediately! Working directory: %s`, projectName, templateContent, projectName, outputFile, outputFile, workDir)
}

// buildFallbackPrompt provides a fallback prompt if template loading fails
func buildFallbackPrompt(projectName, workDir, outputFile string) string {
	return fmt.Sprintf(`You are a professional project analyst. Your task is to analyze the project "%s" and generate comprehensive ALEX.md documentation.

# CRITICAL INSTRUCTIONS:
1. **THIS IS NOT ABOUT CREATING CONVERSATION MEMORY** - You are creating project documentation
2. **OUTPUT MUST BE A MARKDOWN FILE** - Generate actual ALEX.md file content
3. **DO NOT CREATE SHORT-TERM MEMORY** - This is a documentation generation task

# Task Workflow:

## Step 1: Comprehensive Project Analysis
Use the following tools to deeply analyze the project:
- file_list to explore project structure
- file_read to examine key files (README, main.go, config files, core modules)
- grep to search for patterns, features, and technologies used
- Understand project architecture, technology stack, and features
- Analyze code quality, design patterns, and best practices

## Step 2: Generate ALEX.md File
Create a comprehensive documentation file "%s" with complete sections including:
- Project Overview with description of %s
- Essential Development Commands (build, test, usage)
- Architecture Overview with core components
- Built-in tools and features
- Security features and protections
- Performance characteristics
- Code principles and design philosophy
- Current status and testing instructions

# CRITICAL REQUIREMENTS:
1. **GENERATE ACTUAL FILE** - Must create "%s" file with documentation content
2. **NO CONVERSATION MEMORY** - This is pure documentation generation
3. **ANALYZE FIRST** - Thoroughly examine the codebase before writing
4. **PROFESSIONAL QUALITY** - Documentation should be comprehensive and useful

Start analysis and file generation immediately! Working directory: %s`, projectName, outputFile, projectName, outputFile, workDir)
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
