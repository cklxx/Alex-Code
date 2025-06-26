import { promises as fs } from 'fs';
import path from 'path';
import os from 'os';

export interface ConfigData {
  apiKey?: string;
  defaultLanguage?: string;
  outputFormat?: 'json' | 'text';
  analysisDepth?: number;
  backupOnRefactor?: boolean;
  excludePatterns?: string[];
}

export class Config {
  private configPath: string;
  private config: ConfigData;

  constructor() {
    this.configPath = path.join(os.homedir(), '.deep-coding-config.json');
    this.config = {};
  }

  async load(): Promise<void> {
    try {
      const content = await fs.readFile(this.configPath, 'utf-8');
      this.config = JSON.parse(content);
    } catch (error) {
      this.config = this.getDefaultConfig();
      await this.save();
    }
  }

  async save(): Promise<void> {
    const content = JSON.stringify(this.config, null, 2);
    await fs.writeFile(this.configPath, content);
  }

  async get(key: keyof ConfigData): Promise<any> {
    await this.load();
    return this.config[key];
  }

  async set(key: keyof ConfigData, value: any): Promise<void> {
    await this.load();
    this.config[key] = value;
    await this.save();
  }

  async getAll(): Promise<ConfigData> {
    await this.load();
    return { ...this.config };
  }

  async reset(): Promise<void> {
    this.config = this.getDefaultConfig();
    await this.save();
  }

  private getDefaultConfig(): ConfigData {
    return {
      defaultLanguage: 'typescript',
      outputFormat: 'text',
      analysisDepth: 3,
      backupOnRefactor: true,
      excludePatterns: [
        'node_modules/**',
        'dist/**',
        'build/**',
        '.git/**',
        'coverage/**'
      ]
    };
  }
}