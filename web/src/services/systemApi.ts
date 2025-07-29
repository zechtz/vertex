/**
 * System-level API functions
 */
export class SystemApi {
  /**
   * Fix Lombok issues
   */
  static async fixLombok(): Promise<void> {
    const response = await fetch('/api/services/fix-lombok', {
      method: 'POST',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to fix Lombok: ${response.status} ${response.statusText}`);
    }
  }

  /**
   * Sync environment setup
   */
  static async syncEnvironment(): Promise<void> {
    const response = await fetch('/api/environment/setup', {
      method: 'POST',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to sync environment: ${response.status} ${response.statusText}`);
    }
  }

  /**
   * Copy all logs
   */
  static async copyAllLogs(): Promise<void> {
    const response = await fetch('/api/logs/copy-all', {
      method: 'POST',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to copy logs: ${response.status} ${response.statusText}`);
    }
  }

  /**
   * Clear all logs
   */
  static async clearAllLogs(): Promise<void> {
    const response = await fetch('/api/system/logs/cleanup', {
      method: 'POST',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to clear logs: ${response.status} ${response.statusText}`);
    }
  }
}