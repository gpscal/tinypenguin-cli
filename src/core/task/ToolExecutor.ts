import { ToolUseResult } from './types';
import * as fs from 'fs';
import * as path from 'path';
import * as child_process from 'child_process';
import { promisify } from 'util';

const exec = promisify(child_process.exec);

export interface FileOperation {
  path: string;
  content?: string;
  diff?: string;
}

export interface CommandOperation {
  command: string;
  workingDirectory?: string;
  timeoutSeconds?: number;
  requiresApproval?: boolean;
}

export class ToolExecutor {
  private readonly tinyllamaAPI: string;
  private readonly model: string;

  constructor(tinyllamaAPI: string = 'http://localhost:11434/v1', model: string = 'qwen2.5-coder:3b') {
    this.tinyllamaAPI = tinyllamaAPI;
    this.model = model;
  }

  async executeEditFilesTool(
    path: string, 
    content: string,
    diff?: string
  ): Promise<ToolUseResult> {
    try {
      console.log(`📝 Editing file: ${path}`);
      
      // Validate inputs
      if (!path) {
        return {
          status: 'error',
          message: 'File path is required',
          toolName: 'edit_files',
          errorDetails: 'Path parameter is missing'
        };
      }

      // Ensure directory exists
      const dir = path.dirname(path);
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
        console.log(`📁 Created directory: ${dir}`);
      }

      if (diff) {
        // Apply diff-based editing
        return await this.applyDiffToFile(path, diff);
      } else if (content) {
        // Direct content replacement
        return await this.writeContentToFile(path, content);
      } else {
        return {
          status: 'error',
          message: 'Either content or diff is required',
          toolName: 'edit_files',
          errorDetails: 'No content or diff provided'
        };
      }
    } catch (error) {
      console.error(`❌ Error editing file: ${error}`);
      return {
        status: 'error',
        message: `Failed to edit file: ${error.message}`,
        toolName: 'edit_files',
        errorDetails: error.stack || error.toString()
      };
    }
  }

  private async applyDiffToFile(path: string, diff: string): Promise<ToolUseResult> {
    try {
      let currentContent = '';
      
      // Read existing file if it exists
      if (fs.existsSync(path)) {
        currentContent = fs.readFileSync(path, 'utf8');
        console.log(`📄 Current file content read (${currentContent.length} chars)`);
      }

      // Apply diff (simplified implementation)
      const newContent = this.applyDiff(currentContent, diff);
      
      // Write the new content
      fs.writeFileSync(path, newContent, 'utf8');
      
      console.log(`✅ Successfully applied diff to ${path}`);
      return {
        status: 'success',
        message: `Applied diff to file: ${path}`,
        toolName: 'edit_files',
        toolOutput: `File updated with ${diff.split('\n').filter(line => line.startsWith('+')).length} additions and ${diff.split('\n').filter(line => line.startsWith('-')).length} deletions`
      };
    } catch (error) {
      throw new Error(`Failed to apply diff: ${error.message}`);
    }
  }

  private applyDiff(original: string, diff: string): string {
    // Simplified diff application - in a real implementation, you'd use a proper diff library
    const lines = original.split('\n');
    const diffLines = diff.split('\n');
    
    let result: string[] = [...lines];
    let lineIndex = 0;
    
    for (const line of diffLines) {
      if (line.startsWith(' ')) {
        // Unchanged line
        lineIndex++;
      } else if (line.startsWith('+')) {
        // Added line
        result.splice(lineIndex, 0, line.substring(1));
        lineIndex++;
      } else if (line.startsWith('-')) {
        // Removed line
        result.splice(lineIndex, 1);
      }
    }
    
    return result.join('\n');
  }

  private async writeContentToFile(path: string, content: string): Promise<ToolUseResult> {
    try {
      fs.writeFileSync(path, content, 'utf8');
      
      console.log(`✅ Successfully wrote ${content.length} chars to ${path}`);
      return {
        status: 'success',
        message: `Wrote content to file: ${path}`,
        toolName: 'edit_files',
        toolOutput: `File created/updated with ${content.length} characters`
      };
    } catch (error) {
      throw new Error(`Failed to write file: ${error.message}`);
    }
  }

  async executeRunCommandTool(
    command: string, 
    workingDirectory?: string,
    timeoutSeconds: number = 30,
    requiresApproval: boolean = false
  ): Promise<ToolUseResult> {
    try {
      console.log(`💻 Running command: ${command}`);
      console.log(`📂 Working directory: ${workingDirectory || process.cwd()}`);
      console.log(`⏱️  Timeout: ${timeoutSeconds}s`);
      console.log(`🔒 Requires approval: ${requiresApproval}`);

      // Validate command
      if (!command) {
        return {
          status: 'error',
          message: 'Command is required',
          toolName: 'run_commands',
          errorDetails: 'Command parameter is missing'
        };
      }

      // Security check for dangerous commands
      if (this.isDangerousCommand(command)) {
        return {
          status: 'denied',
          message: 'Command was denied for safety reasons',
          toolName: 'run_commands',
          errorDetails: 'Command contains potentially dangerous operations'
        };
      }

      // Prepare execution options
      const options: child_process.ExecOptions = {
        cwd: workingDirectory || process.cwd(),
        timeout: timeoutSeconds * 1000,
        maxBuffer: 10 * 1024 * 1024, // 10MB buffer
      };

      // Execute the command
      console.log(`🚀 Executing command...`);
      const { stdout, stderr } = await exec(command, options);
      
      const output = stdout + (stderr ? `\nSTDERR: ${stderr}` : '');
      
      console.log(`✅ Command completed successfully`);
      console.log(`📤 Output: ${output.substring(0, 200)}${output.length > 200 ? '...' : ''}`);
      
      return {
        status: 'success',
        message: `Command executed successfully`,
        toolName: 'run_commands',
        toolOutput: output
      };
    } catch (error) {
      console.error(`❌ Command failed: ${error}`);
      
      if (error.message.includes('timed out')) {
        return {
          status: 'error',
          message: 'Command timed out',
          toolName: 'run_commands',
          errorDetails: `Command exceeded ${timeoutSeconds}s timeout`
        };
      }
      
      return {
        status: 'error',
        message: `Command failed: ${error.message}`,
        toolName: 'run_commands',
        errorDetails: error.stack || error.toString()
      };
    }
  }

  private isDangerousCommand(command: string): boolean {
    const dangerousPatterns = [
      /^rm\s+-rf\s+\/$/,
      /^rm\s+-rf\s+\/usr/,
      /^rm\s+-rf\s+\/bin/,
      /^dd\s+if=/,
      /^mkfs/,
      /^fdisk/,
      /^shred/,
      /^cryptsetup/,
      /^chmod\s+777/,
      /^chown\s+.*\:.*\/etc/,
    ];

    const lowerCommand = command.toLowerCase();
    
    return dangerousPatterns.some(pattern => pattern.test(lowerCommand)) ||
           (lowerCommand.includes('sudo') && lowerCommand.includes('rm -rf /'));
  }

  async checkFileExists(path: string): Promise<boolean> {
    try {
      await fs.promises.access(path);
      return true;
    } catch {
      return false;
    }
  }

  async readFile(path: string): Promise<string> {
    try {
      return await fs.promises.readFile(path, 'utf8');
    } catch (error) {
      throw new Error(`Failed to read file ${path}: ${error.message}`);
    }
  }

  // Test connectivity to tinyllama
  async testTinyllamaConnection(): Promise<boolean> {
    try {
      const response = await fetch(`${this.tinyllamaAPI}/models`);
      return response.ok;
    } catch (error) {
      console.error(`❌ Failed to connect to tinyllama: ${error}`);
      return false;
    }
  }

  // Get available models from tinyllama
  async getAvailableModels(): Promise<string[]> {
    try {
      const response = await fetch(`${this.tinyllamaAPI}/models`);
      const data = await response.json();
      return data.models?.map(m => m.name) || [];
    } catch (error) {
      console.error(`❌ Failed to get models: ${error}`);
      return [];
    }
  }
}