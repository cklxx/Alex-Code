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

	return fmt.Sprintf(`You are a professional project analyst. Please conduct an in-depth analysis of the project "%s" and generate comprehensive project documentation using the provided template.

# Task Workflow:

## Step 1: Comprehensive Project Analysis
Use the following tools to deeply analyze the project:
- file_search, grep_search, codebase_search to analyze project structure
- read_file to examine key files (README, config files, core code)
- Understand project architecture, technology stack, and features
- Analyze code quality, design patterns, and best practices

## Step 2: Use the Provided Template
The following is the documentation template with placeholders in {{VariableName}} format.
You need to fill in the analysis results into the corresponding variables:

---
%s
---

## Step 3: Generate Complete Documentation
Write the filled content to file "%s", ensuring:
- All {{variables}} are replaced with actual content based on your project analysis
- Content is detailed, accurate, and professional
- Format is beautiful and structure is clear
- Include practical usage instructions and development guides
- Replace {{ProjectName}} with "%s"

# Important Requirements:
1. **Take Action Immediately** - Start analysis right away, don't ask any questions
2. **Deep Analysis** - Fully understand the project's tech stack, architecture, and functionality
3. **Use Provided Template** - Use the template above and fill all variables with real content
4. **Generate File** - Write final documentation to the specified file
5. **Rich Content** - Ensure each section has substantial content

Start executing immediately! Working directory: %s`, projectName, templateContent, outputFile, projectName, workDir)
}

// buildFallbackPrompt provides a fallback prompt if template loading fails
func buildFallbackPrompt(projectName, workDir, outputFile string) string {
	return fmt.Sprintf(`You are a professional project analyst. Please conduct an in-depth analysis of the project "%s" and generate comprehensive project documentation.

# Task Workflow:

## Step 1: Comprehensive Project Analysis
Use the following tools to deeply analyze the project:
- file_search, grep_search, codebase_search to analyze project structure
- read_file to examine key files (README, config files, core code)
- Understand project architecture, technology stack, and features
- Analyze code quality, design patterns, and best practices

## Step 2: Generate Documentation
Create a comprehensive project documentation file "%s" with the following sections:
- Project Overview and Goals
- Architecture Design and Core Components
- Technology Stack and Dependencies
- Installation and Usage Instructions
- Development Guide and Best Practices
- API Documentation (if applicable)
- Performance Characteristics
- Current Status and Future Plans

# Important Requirements:
1. **Take Action Immediately** - Start analysis right away, don't ask any questions
2. **Deep Analysis** - Fully understand the project's tech stack, architecture, and functionality
3. **Generate File** - Write final documentation to the specified file
4. **Rich Content** - Ensure each section has substantial content

Start executing immediately! Working directory: %s`, projectName, outputFile, workDir)
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
