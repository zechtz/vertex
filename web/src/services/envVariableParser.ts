export type EnvFormat = 
  | 'dotenv' 
  | 'json' 
  | 'yaml' 
  | 'shell' 
  | 'fish' 
  | 'batch' 
  | 'properties'
  | 'semicolon'
  | 'auto';

export interface ParseResult {
  variables: Record<string, string>;
  errors: string[];
  format: EnvFormat;
}

export class EnvVariableParser {
  /**
   * Parse environment variables from text in various formats
   */
  static parse(text: string, format: EnvFormat = 'auto'): ParseResult {
    const trimmedText = text.trim();
    
    if (!trimmedText) {
      return { variables: {}, errors: [], format: 'dotenv' };
    }

    // Auto-detect format if not specified
    const detectedFormat = format === 'auto' ? this.detectFormat(trimmedText) : format;
    
    switch (detectedFormat) {
      case 'json':
        return this.parseJSON(trimmedText);
      case 'yaml':
        return this.parseYAML(trimmedText);
      case 'shell':
        return this.parseShell(trimmedText);
      case 'fish':
        return this.parseFish(trimmedText);
      case 'batch':
        return this.parseBatch(trimmedText);
      case 'properties':
        return this.parseProperties(trimmedText);
      case 'semicolon':
        return this.parseSemicolon(trimmedText);
      case 'dotenv':
      default:
        return this.parseDotEnv(trimmedText);
    }
  }

  /**
   * Auto-detect the format of the input text
   */
  private static detectFormat(text: string): EnvFormat {
    const trimmed = text.trim();
    
    // JSON detection
    if ((trimmed.startsWith('{') && trimmed.endsWith('}')) || 
        (trimmed.startsWith('[') && trimmed.endsWith(']'))) {
      return 'json';
    }
    
    // YAML detection (basic heuristics)
    if (trimmed.includes(':\n') || trimmed.includes(': ') || 
        /^[a-zA-Z_][a-zA-Z0-9_]*:\s/.test(trimmed)) {
      return 'yaml';
    }
    
    // Shell export detection
    if (trimmed.includes('export ') || /^export\s+[A-Z_][A-Z0-9_]*=/m.test(trimmed)) {
      return 'shell';
    }
    
    // Fish shell detection
    if (trimmed.includes('set -gx ') || /^set\s+-gx\s+[A-Za-z_][A-Za-z0-9_]*\s+/m.test(trimmed)) {
      return 'fish';
    }
    
    // Batch detection
    if (trimmed.includes('set ') || /^set\s+[A-Z_][A-Z0-9_]*=/m.test(trimmed)) {
      return 'batch';
    }
    
    // Properties detection (contains dots in keys)
    if (/^[a-zA-Z][a-zA-Z0-9._]*=/.test(trimmed) && trimmed.includes('.')) {
      return 'properties';
    }
    
    // Semicolon-separated detection (single line with semicolons)
    if (!trimmed.includes('\n') && trimmed.includes(';') && trimmed.includes('=')) {
      return 'semicolon';
    }
    
    // Default to dotenv
    return 'dotenv';
  }

  /**
   * Parse .env / Docker Compose format
   */
  private static parseDotEnv(text: string): ParseResult {
    const variables: Record<string, string> = {};
    const errors: string[] = [];
    const lines = text.split('\n');

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      const lineNum = i + 1;
      
      // Skip empty lines and comments
      if (!line || line.startsWith('#')) {
        continue;
      }

      // Handle line continuations (backslash at end)
      let fullLine = line;
      let j = i;
      while (fullLine.endsWith('\\') && j + 1 < lines.length) {
        fullLine = fullLine.slice(0, -1) + lines[++j].trim();
      }
      i = j; // Update loop counter

      const match = fullLine.match(/^([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$/);
      if (match) {
        const [, key, value] = match;
        variables[key] = this.unquoteValue(value);
      } else {
        errors.push(`Line ${lineNum}: Invalid format - "${line}"`);
      }
    }

    return { variables, errors, format: 'dotenv' };
  }

  /**
   * Parse JSON format
   */
  private static parseJSON(text: string): ParseResult {
    try {
      const data = JSON.parse(text);
      const variables: Record<string, string> = {};
      const errors: string[] = [];

      if (typeof data === 'object' && data !== null && !Array.isArray(data)) {
        for (const [key, value] of Object.entries(data)) {
          if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
            variables[key] = String(value);
          } else {
            errors.push(`Key "${key}": Value must be a string, number, or boolean`);
          }
        }
      } else {
        errors.push('JSON must be an object with key-value pairs');
      }

      return { variables, errors, format: 'json' };
    } catch (error) {
      return { 
        variables: {}, 
        errors: [`Invalid JSON: ${error instanceof Error ? error.message : 'Unknown error'}`], 
        format: 'json' 
      };
    }
  }

  /**
   * Parse YAML format (basic implementation)
   */
  private static parseYAML(text: string): ParseResult {
    const variables: Record<string, string> = {};
    const errors: string[] = [];
    const lines = text.split('\n');

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      const lineNum = i + 1;
      
      // Skip empty lines and comments
      if (!line || line.startsWith('#')) {
        continue;
      }

      const match = line.match(/^([A-Za-z_][A-Za-z0-9_]*)\s*:\s*(.*)$/);
      if (match) {
        const [, key, value] = match;
        variables[key] = this.unquoteValue(value);
      } else {
        errors.push(`Line ${lineNum}: Invalid YAML format - "${line}"`);
      }
    }

    return { variables, errors, format: 'yaml' };
  }

  /**
   * Parse shell export format
   */
  private static parseShell(text: string): ParseResult {
    const variables: Record<string, string> = {};
    const errors: string[] = [];
    const lines = text.split('\n');

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      const lineNum = i + 1;
      
      // Skip empty lines and comments
      if (!line || line.startsWith('#')) {
        continue;
      }

      const match = line.match(/^export\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$/);
      if (match) {
        const [, key, value] = match;
        variables[key] = this.unquoteValue(value);
      } else {
        errors.push(`Line ${lineNum}: Invalid shell export format - "${line}"`);
      }
    }

    return { variables, errors, format: 'shell' };
  }

  /**
   * Parse Fish shell format
   */
  private static parseFish(text: string): ParseResult {
    const variables: Record<string, string> = {};
    const errors: string[] = [];
    const lines = text.split('\n');

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      const lineNum = i + 1;
      
      // Skip empty lines and comments
      if (!line || line.startsWith('#')) {
        continue;
      }

      // Match fish shell format: set -gx VARIABLE_NAME value
      // Also support other set variations: set -g, set -x, set -l
      const match = line.match(/^set\s+(-[gxl]*\s+)?([A-Za-z_][A-Za-z0-9_]*)\s+(.*)$/);
      if (match) {
        const [, , key, value] = match; // Skip flags parameter
        variables[key] = this.unquoteValue(value);
      } else {
        errors.push(`Line ${lineNum}: Invalid fish shell format - "${line}"`);
      }
    }

    return { variables, errors, format: 'fish' };
  }

  /**
   * Parse Windows batch format
   */
  private static parseBatch(text: string): ParseResult {
    const variables: Record<string, string> = {};
    const errors: string[] = [];
    const lines = text.split('\n');

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      const lineNum = i + 1;
      
      // Skip empty lines and comments
      if (!line || line.startsWith('REM ') || line.startsWith('::')) {
        continue;
      }

      const match = line.match(/^set\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$/i);
      if (match) {
        const [, key, value] = match;
        variables[key] = this.unquoteValue(value);
      } else {
        errors.push(`Line ${lineNum}: Invalid batch format - "${line}"`);
      }
    }

    return { variables, errors, format: 'batch' };
  }

  /**
   * Parse Java properties format
   */
  private static parseProperties(text: string): ParseResult {
    const variables: Record<string, string> = {};
    const errors: string[] = [];
    const lines = text.split('\n');

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      const lineNum = i + 1;
      
      // Skip empty lines and comments
      if (!line || line.startsWith('#') || line.startsWith('!')) {
        continue;
      }

      // Handle line continuations (backslash at end)
      let fullLine = line;
      let j = i;
      while (fullLine.endsWith('\\') && j + 1 < lines.length) {
        fullLine = fullLine.slice(0, -1) + lines[++j].trim();
      }
      i = j;

      const match = fullLine.match(/^([A-Za-z_][A-Za-z0-9._]*)\s*[=:]\s*(.*)$/);
      if (match) {
        const [, key, value] = match;
        variables[key] = this.unquoteValue(value);
      } else {
        errors.push(`Line ${lineNum}: Invalid properties format - "${line}"`);
      }
    }

    return { variables, errors, format: 'properties' };
  }

  /**
   * Parse semicolon-separated format (key1=value1;key2=value2;key3=value3)
   */
  private static parseSemicolon(text: string): ParseResult {
    const variables: Record<string, string> = {};
    const errors: string[] = [];
    
    // Split by semicolon and process each pair
    const pairs = text.split(';').map(pair => pair.trim()).filter(pair => pair);
    
    for (let i = 0; i < pairs.length; i++) {
      const pair = pairs[i];
      const pairNum = i + 1;
      
      // Skip empty pairs
      if (!pair) {
        continue;
      }
      
      const match = pair.match(/^([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$/);
      if (match) {
        const [, key, value] = match;
        variables[key] = this.unquoteValue(value);
      } else {
        errors.push(`Pair ${pairNum}: Invalid format - "${pair}"`);
      }
    }

    return { variables, errors, format: 'semicolon' };
  }

  /**
   * Remove quotes from values and handle escape sequences
   */
  private static unquoteValue(value: string): string {
    if (!value) return '';
    
    const trimmed = value.trim();
    
    // Remove surrounding quotes
    if ((trimmed.startsWith('"') && trimmed.endsWith('"')) ||
        (trimmed.startsWith("'") && trimmed.endsWith("'"))) {
      return trimmed.slice(1, -1)
        .replace(/\\"/g, '"')
        .replace(/\\'/g, "'")
        .replace(/\\n/g, '\n')
        .replace(/\\r/g, '\r')
        .replace(/\\t/g, '\t')
        .replace(/\\\\/g, '\\');
    }
    
    return trimmed;
  }

  /**
   * Get format display name
   */
  static getFormatDisplayName(format: EnvFormat): string {
    switch (format) {
      case 'dotenv': return '.env / Docker Compose';
      case 'json': return 'JSON';
      case 'yaml': return 'YAML';
      case 'shell': return 'Shell Export';
      case 'fish': return 'Fish Shell';
      case 'batch': return 'Windows Batch';
      case 'properties': return 'Java Properties';
      case 'semicolon': return 'Semicolon Separated';
      case 'auto': return 'Auto-detect';
      default: return format;
    }
  }

  /**
   * Get example for a format
   */
  static getFormatExample(format: EnvFormat): string {
    switch (format) {
      case 'dotenv':
        return `# .env format
DATABASE_URL=postgres://localhost:5432/db
API_KEY=your-api-key
DEBUG=true`;

      case 'json':
        return `{
  "DATABASE_URL": "postgres://localhost:5432/db",
  "API_KEY": "your-api-key",
  "DEBUG": "true"
}`;

      case 'yaml':
        return `# YAML format
DATABASE_URL: postgres://localhost:5432/db
API_KEY: your-api-key
DEBUG: true`;

      case 'shell':
        return `#!/bin/bash
export DATABASE_URL=postgres://localhost:5432/db
export API_KEY=your-api-key
export DEBUG=true`;

      case 'fish':
        return `#!/usr/bin/fish
set -gx DATABASE_URL postgres://localhost:5432/db
set -gx API_KEY your-api-key
set -gx DEBUG true`;

      case 'batch':
        return `REM Windows batch format
set DATABASE_URL=postgres://localhost:5432/db
set API_KEY=your-api-key
set DEBUG=true`;

      case 'properties':
        return `# Java properties format
database.url=postgres://localhost:5432/db
api.key=your-api-key
debug=true`;

      case 'semicolon':
        return `DATABASE_URL=postgres://localhost:5432/db;API_KEY=your-api-key;DEBUG=true;REDIS_HOST=localhost;REDIS_PORT=6379`;

      default:
        return '';
    }
  }
}