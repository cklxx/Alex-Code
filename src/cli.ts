#!/usr/bin/env node

import { Command } from 'commander';
import chalk from 'chalk';
import { DeepCodingAgent } from './core/agent';
import { version } from '../package.json';

const program = new Command();
const agent = new DeepCodingAgent();

program
  .name('deep-coding')
  .description('Deep coding agent for analysis and generation')
  .version(version);

program
  .command('analyze')
  .description('Analyze code structure and patterns')
  .argument('<path>', 'Path to analyze')
  .option('-d, --depth <number>', 'Analysis depth', '3')
  .option('-f, --format <format>', 'Output format (json|text)', 'text')
  .action(async (path: string, options: any) => {
    console.log(chalk.blue('🔍 Analyzing code...'));
    try {
      const result = await agent.analyze(path, options);
      console.log(result);
    } catch (error) {
      console.error(chalk.red('Error:'), error instanceof Error ? error.message : error);
      process.exit(1);
    }
  });

program
  .command('generate')
  .description('Generate code based on specifications')
  .argument('<spec>', 'Code specification or description')
  .option('-l, --language <lang>', 'Target language', 'typescript')
  .option('-o, --output <path>', 'Output file path')
  .action(async (spec: string, options: any) => {
    console.log(chalk.green('🚀 Generating code...'));
    try {
      const result = await agent.generate(spec, options);
      console.log(result);
    } catch (error) {
      console.error(chalk.red('Error:'), error instanceof Error ? error.message : error);
      process.exit(1);
    }
  });

program
  .command('refactor')
  .description('Refactor existing code')
  .argument('<path>', 'Path to refactor')
  .option('-p, --pattern <pattern>', 'Refactor pattern to apply')
  .option('-b, --backup', 'Create backup before refactoring')
  .action(async (path: string, options: any) => {
    console.log(chalk.yellow('🔧 Refactoring code...'));
    try {
      const result = await agent.refactor(path, options);
      console.log(result);
    } catch (error) {
      console.error(chalk.red('Error:'), error instanceof Error ? error.message : error);
      process.exit(1);
    }
  });

program
  .command('config')
  .description('Manage configuration settings')
  .option('-s, --set <key=value>', 'Set configuration value')
  .option('-g, --get <key>', 'Get configuration value')
  .option('-l, --list', 'List all configuration')
  .action(async (options: any) => {
    try {
      const result = await agent.config(options);
      console.log(result);
    } catch (error) {
      console.error(chalk.red('Error:'), error instanceof Error ? error.message : error);
      process.exit(1);
    }
  });

program.parse();