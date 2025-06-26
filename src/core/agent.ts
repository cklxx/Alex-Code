import { promises as fs } from 'fs';
import path from 'path';
import chalk from 'chalk';
import { Config } from '../utils/config';
import { CodeAnalyzer } from './analyzer';
import { CodeGenerator } from './generator';
import { CodeRefactor } from './refactor';

export interface AnalyzeOptions {
  depth: string;
  format: 'json' | 'text';
}

export interface GenerateOptions {
  language: string;
  output?: string;
}

export interface RefactorOptions {
  pattern?: string;
  backup?: boolean;
}

export interface ConfigOptions {
  set?: string;
  get?: string;
  list?: boolean;
}

export class DeepCodingAgent {
  private configManager: Config;
  private analyzer: CodeAnalyzer;
  private generator: CodeGenerator;
  private refactorManager: CodeRefactor;

  constructor() {
    this.configManager = new Config();
    this.analyzer = new CodeAnalyzer();
    this.generator = new CodeGenerator();
    this.refactorManager = new CodeRefactor();
  }

  async analyze(targetPath: string, options: AnalyzeOptions): Promise<string> {
    try {
      const depth = parseInt(options.depth);
      const stats = await fs.stat(targetPath);
      
      let result;
      if (stats.isDirectory()) {
        result = await this.analyzer.analyzeDirectory(targetPath, depth);
      } else {
        result = await this.analyzer.analyzeFile(targetPath);
      }

      if (options.format === 'json') {
        return JSON.stringify(result, null, 2);
      }

      return this.formatAnalysisResult(result);
    } catch (error) {
      throw new Error(`Analysis failed: ${error instanceof Error ? error.message : error}`);
    }
  }

  async generate(spec: string, options: GenerateOptions): Promise<string> {
    try {
      const result = await this.generator.generate(spec, options.language);
      
      if (options.output) {
        await fs.writeFile(options.output, result.code);
        return chalk.green(`✅ Code generated and saved to ${options.output}`);
      }

      return result.code;
    } catch (error) {
      throw new Error(`Generation failed: ${error instanceof Error ? error.message : error}`);
    }
  }

  async refactor(targetPath: string, options: RefactorOptions): Promise<string> {
    try {
      if (options.backup) {
        const backupPath = `${targetPath}.backup`;
        await fs.copyFile(targetPath, backupPath);
      }

      const result = await this.refactorManager.refactorFile(targetPath, options.pattern);
      return chalk.green(`✅ Refactoring completed: ${result.summary}`);
    } catch (error) {
      throw new Error(`Refactoring failed: ${error instanceof Error ? error.message : error}`);
    }
  }

  async config(options: ConfigOptions): Promise<string> {
    try {
      if (options.set) {
        const [key, value] = options.set.split('=');
        await this.configManager.set(key as any, value);
        return chalk.green(`✅ Configuration set: ${key} = ${value}`);
      }

      if (options.get) {
        const value = await this.configManager.get(options.get as any);
        return `${options.get}: ${value}`;
      }

      if (options.list) {
        const config = await this.configManager.getAll();
        return JSON.stringify(config, null, 2);
      }

      return 'No configuration action specified';
    } catch (error) {
      throw new Error(`Configuration failed: ${error instanceof Error ? error.message : error}`);
    }
  }

  private formatAnalysisResult(result: any): string {
    const lines = [
      chalk.blue.bold('📊 Code Analysis Results'),
      '',
      `${chalk.cyan('Files analyzed:')} ${result.fileCount || 0}`,
      `${chalk.cyan('Lines of code:')} ${result.linesOfCode || 0}`,
      `${chalk.cyan('Complexity score:')} ${result.complexity || 'N/A'}`,
      '',
    ];

    if (result.patterns && result.patterns.length > 0) {
      lines.push(chalk.yellow.bold('🔍 Detected Patterns:'));
      result.patterns.forEach((pattern: any) => {
        lines.push(`  • ${pattern.name}: ${pattern.description}`);
      });
      lines.push('');
    }

    if (result.suggestions && result.suggestions.length > 0) {
      lines.push(chalk.green.bold('💡 Suggestions:'));
      result.suggestions.forEach((suggestion: string) => {
        lines.push(`  • ${suggestion}`);
      });
    }

    return lines.join('\n');
  }
}