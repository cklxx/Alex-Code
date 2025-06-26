import { promises as fs } from 'fs';

export interface RefactorResult {
  summary: string;
  changes: Change[];
}

export interface Change {
  type: 'rename' | 'extract' | 'inline' | 'format';
  description: string;
  lineNumber?: number;
}

export class CodeRefactor {
  async refactorFile(filePath: string, pattern?: string): Promise<RefactorResult> {
    const content = await fs.readFile(filePath, 'utf-8');
    let refactoredContent = content;
    const changes: Change[] = [];

    if (pattern) {
      const result = await this.applyPattern(refactoredContent, pattern);
      refactoredContent = result.content;
      changes.push(...result.changes);
    } else {
      const result = await this.applyDefaultRefactoring(refactoredContent);
      refactoredContent = result.content;
      changes.push(...result.changes);
    }

    if (refactoredContent !== content) {
      await fs.writeFile(filePath, refactoredContent);
    }

    return {
      summary: `Applied ${changes.length} refactoring changes`,
      changes
    };
  }

  private async applyPattern(content: string, pattern: string): Promise<{ content: string; changes: Change[] }> {
    const changes: Change[] = [];
    let refactoredContent = content;

    switch (pattern.toLowerCase()) {
      case 'extract-function':
        const extractResult = this.extractLongFunctions(refactoredContent);
        refactoredContent = extractResult.content;
        changes.push(...extractResult.changes);
        break;

      case 'rename-variables':
        const renameResult = this.improveVariableNames(refactoredContent);
        refactoredContent = renameResult.content;
        changes.push(...renameResult.changes);
        break;

      case 'remove-unused':
        const removeResult = this.removeUnusedCode(refactoredContent);
        refactoredContent = removeResult.content;
        changes.push(...removeResult.changes);
        break;

      default:
        changes.push({
          type: 'format',
          description: `Unknown pattern: ${pattern}. Applied default formatting.`
        });
    }

    return { content: refactoredContent, changes };
  }

  private async applyDefaultRefactoring(content: string): Promise<{ content: string; changes: Change[] }> {
    const changes: Change[] = [];
    let refactoredContent = content;

    const formatResult = this.formatCode(refactoredContent);
    refactoredContent = formatResult.content;
    changes.push(...formatResult.changes);

    const unusedResult = this.removeUnusedCode(refactoredContent);
    refactoredContent = unusedResult.content;
    changes.push(...unusedResult.changes);

    return { content: refactoredContent, changes };
  }

  private extractLongFunctions(content: string): { content: string; changes: Change[] } {
    const changes: Change[] = [];
    const lines = content.split('\n');
    let refactoredContent = content;

    const functionPattern = /function\s+(\w+)\s*\(/;
    let inFunction = false;
    let functionStart = -1;
    let braceCount = 0;

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      
      if (functionPattern.test(line)) {
        inFunction = true;
        functionStart = i;
        braceCount = 0;
      }
      
      if (inFunction) {
        braceCount += (line.match(/\{/g) || []).length;
        braceCount -= (line.match(/\}/g) || []).length;
        
        if (braceCount === 0 && functionStart !== -1) {
          const functionLength = i - functionStart + 1;
          if (functionLength > 20) {
            changes.push({
              type: 'extract',
              description: `Long function detected (${functionLength} lines). Consider extracting smaller functions.`,
              lineNumber: functionStart + 1
            });
          }
          inFunction = false;
          functionStart = -1;
        }
      }
    }

    return { content: refactoredContent, changes };
  }

  private improveVariableNames(content: string): { content: string; changes: Change[] } {
    const changes: Change[] = [];
    let refactoredContent = content;

    const poorNames = [
      { old: /\bdata\b/g, new: 'result' },
      { old: /\btemp\b/g, new: 'temporary' },
      { old: /\bi\b/g, new: 'index' },
      { old: /\bj\b/g, new: 'innerIndex' },
      { old: /\bx\b/g, new: 'value' },
      { old: /\by\b/g, new: 'coordinate' }
    ];

    poorNames.forEach(({ old, new: newName }) => {
      if (old.test(refactoredContent)) {
        refactoredContent = refactoredContent.replace(old, newName);
        changes.push({
          type: 'rename',
          description: `Renamed variable to improve clarity: ${old.source} -> ${newName}`
        });
      }
    });

    return { content: refactoredContent, changes };
  }

  private removeUnusedCode(content: string): { content: string; changes: Change[] } {
    const changes: Change[] = [];
    let refactoredContent = content;

    const lines = refactoredContent.split('\n');
    const filteredLines = lines.filter(line => {
      const trimmed = line.trim();
      
      if (trimmed === '' || trimmed.startsWith('//')) {
        return true;
      }
      
      if (trimmed.includes('console.log') && !trimmed.includes('// keep')) {
        changes.push({
          type: 'inline',
          description: 'Removed console.log statement'
        });
        return false;
      }
      
      if (trimmed.includes('debugger')) {
        changes.push({
          type: 'inline',
          description: 'Removed debugger statement'
        });
        return false;
      }
      
      return true;
    });

    refactoredContent = filteredLines.join('\n');

    return { content: refactoredContent, changes };
  }

  private formatCode(content: string): { content: string; changes: Change[] } {
    const changes: Change[] = [];
    let refactoredContent = content;

    const lines = refactoredContent.split('\n');
    const formattedLines = lines.map(line => {
      let formatted = line;
      
      formatted = formatted.replace(/\s+$/, '');
      
      if (formatted.trim() && !formatted.match(/^\s/)) {
        formatted = formatted.trim();
      }
      
      return formatted;
    });

    refactoredContent = formattedLines.join('\n');

    const originalLineCount = lines.length;
    const newLineCount = formattedLines.length;
    
    if (originalLineCount !== newLineCount) {
      changes.push({
        type: 'format',
        description: `Code formatting applied. Line count changed from ${originalLineCount} to ${newLineCount}`
      });
    }

    return { content: refactoredContent, changes };
  }
}