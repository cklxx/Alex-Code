import { promises as fs } from 'fs';
import path from 'path';

export interface AnalysisResult {
  fileCount: number;
  linesOfCode: number;
  complexity: number;
  patterns: Pattern[];
  suggestions: string[];
  files?: FileAnalysis[];
}

export interface FileAnalysis {
  path: string;
  size: number;
  lines: number;
  functions: number;
  classes: number;
  imports: string[];
  complexity: number;
}

export interface Pattern {
  name: string;
  description: string;
  occurrences: number;
}

export class CodeAnalyzer {
  private supportedExtensions = ['.ts', '.js', '.tsx', '.jsx', '.py', '.java', '.cpp', '.c', '.cs'];

  async analyzeDirectory(dirPath: string, depth: number = 3): Promise<AnalysisResult> {
    const files = await this.getCodeFiles(dirPath, depth);
    const fileAnalyses: FileAnalysis[] = [];
    
    for (const file of files) {
      try {
        const analysis = await this.analyzeFile(file);
        fileAnalyses.push(analysis);
      } catch (error) {
        console.warn(`Warning: Could not analyze ${file}: ${error}`);
      }
    }

    return this.aggregateResults(fileAnalyses);
  }

  async analyzeFile(filePath: string): Promise<FileAnalysis> {
    const content = await fs.readFile(filePath, 'utf-8');
    const stats = await fs.stat(filePath);
    
    const lines = content.split('\n');
    const nonEmptyLines = lines.filter(line => line.trim().length > 0);
    
    return {
      path: filePath,
      size: stats.size,
      lines: nonEmptyLines.length,
      functions: this.countFunctions(content),
      classes: this.countClasses(content),
      imports: this.extractImports(content),
      complexity: this.calculateComplexity(content)
    };
  }

  private async getCodeFiles(dirPath: string, maxDepth: number, currentDepth: number = 0): Promise<string[]> {
    if (currentDepth >= maxDepth) return [];
    
    const files: string[] = [];
    const entries = await fs.readdir(dirPath, { withFileTypes: true });
    
    for (const entry of entries) {
      const fullPath = path.join(dirPath, entry.name);
      
      if (entry.isDirectory() && !this.shouldIgnoreDirectory(entry.name)) {
        const subFiles = await this.getCodeFiles(fullPath, maxDepth, currentDepth + 1);
        files.push(...subFiles);
      } else if (entry.isFile() && this.isCodeFile(entry.name)) {
        files.push(fullPath);
      }
    }
    
    return files;
  }

  private isCodeFile(filename: string): boolean {
    const ext = path.extname(filename);
    return this.supportedExtensions.includes(ext);
  }

  private shouldIgnoreDirectory(dirname: string): boolean {
    const ignoreDirs = ['node_modules', '.git', 'dist', 'build', 'coverage', '.next'];
    return ignoreDirs.includes(dirname);
  }

  private countFunctions(content: string): number {
    const functionPatterns = [
      /function\s+\w+/g,
      /const\s+\w+\s*=\s*\(/g,
      /\w+\s*:\s*\([^)]*\)\s*=>/g,
      /def\s+\w+/g,
      /public\s+\w+\s+\w+\s*\(/g
    ];
    
    let count = 0;
    for (const pattern of functionPatterns) {
      const matches = content.match(pattern);
      if (matches) count += matches.length;
    }
    
    return count;
  }

  private countClasses(content: string): number {
    const classPattern = /class\s+\w+/g;
    const matches = content.match(classPattern);
    return matches ? matches.length : 0;
  }

  private extractImports(content: string): string[] {
    const importPatterns = [
      /import\s+.*?from\s+['"`]([^'"`]+)['"`]/g,
      /require\s*\(\s*['"`]([^'"`]+)['"`]\s*\)/g,
      /from\s+['"`]([^'"`]+)['"`]/g
    ];
    
    const imports: string[] = [];
    for (const pattern of importPatterns) {
      let match;
      while ((match = pattern.exec(content)) !== null) {
        imports.push(match[1]);
      }
    }
    
    return [...new Set(imports)];
  }

  private calculateComplexity(content: string): number {
    const complexityPatterns = [
      /if\s*\(/g,
      /else\s+if\s*\(/g,
      /while\s*\(/g,
      /for\s*\(/g,
      /switch\s*\(/g,
      /case\s+/g,
      /catch\s*\(/g,
      /&&/g,
      /\|\|/g
    ];
    
    let complexity = 1;
    for (const pattern of complexityPatterns) {
      const matches = content.match(pattern);
      if (matches) complexity += matches.length;
    }
    
    return complexity;
  }

  private aggregateResults(fileAnalyses: FileAnalysis[]): AnalysisResult {
    const totalLines = fileAnalyses.reduce((sum, file) => sum + file.lines, 0);
    const totalComplexity = fileAnalyses.reduce((sum, file) => sum + file.complexity, 0);
    const avgComplexity = fileAnalyses.length > 0 ? Math.round(totalComplexity / fileAnalyses.length) : 0;
    
    const patterns = this.detectPatterns(fileAnalyses);
    const suggestions = this.generateSuggestions(fileAnalyses, avgComplexity);
    
    return {
      fileCount: fileAnalyses.length,
      linesOfCode: totalLines,
      complexity: avgComplexity,
      patterns,
      suggestions,
      files: fileAnalyses
    };
  }

  private detectPatterns(fileAnalyses: FileAnalysis[]): Pattern[] {
    const patterns: Pattern[] = [];
    
    const hasReactFiles = fileAnalyses.some(f => f.imports.some(imp => imp.includes('react')));
    if (hasReactFiles) {
      patterns.push({
        name: 'React Application',
        description: 'React-based frontend application detected',
        occurrences: fileAnalyses.filter(f => f.imports.some(imp => imp.includes('react'))).length
      });
    }
    
    const hasTestFiles = fileAnalyses.some(f => f.path.includes('.test.') || f.path.includes('.spec.'));
    if (hasTestFiles) {
      patterns.push({
        name: 'Testing Framework',
        description: 'Test files detected in the codebase',
        occurrences: fileAnalyses.filter(f => f.path.includes('.test.') || f.path.includes('.spec.')).length
      });
    }
    
    return patterns;
  }

  private generateSuggestions(fileAnalyses: FileAnalysis[], avgComplexity: number): string[] {
    const suggestions: string[] = [];
    
    if (avgComplexity > 10) {
      suggestions.push('Consider refactoring complex functions to improve maintainability');
    }
    
    const largeFiles = fileAnalyses.filter(f => f.lines > 500);
    if (largeFiles.length > 0) {
      suggestions.push(`${largeFiles.length} files have more than 500 lines - consider splitting them`);
    }
    
    const filesWithoutTests = fileAnalyses.filter(f => !f.path.includes('.test.') && !f.path.includes('.spec.'));
    const hasAnyTests = fileAnalyses.some(f => f.path.includes('.test.') || f.path.includes('.spec.'));
    
    if (!hasAnyTests && filesWithoutTests.length > 0) {
      suggestions.push('Consider adding unit tests to improve code reliability');
    }
    
    return suggestions;
  }
}