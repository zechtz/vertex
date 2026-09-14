import type { BuildInfo, EnvVar, InstalledJdk } from '@/types';

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

  /**
   * List the JDKs installed on the machine running Vertex.
   */
  static async getInstalledJdks(): Promise<InstalledJdk[]> {
    const response = await fetch('/api/java/jdks');

    if (!response.ok) {
      throw new Error(`Failed to list JDKs: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }

  /**
   * Pin a service to a specific JDK by setting its JAVA_HOME, which takes
   * priority over the profile and global Java settings. Existing variables are
   * preserved: the endpoint replaces the whole set.
   */
  static async setServiceJavaHome(
    serviceId: string,
    existing: Record<string, EnvVar>,
    javaHome: string,
  ): Promise<void> {
    const envVars: Record<string, EnvVar> = {
      ...existing,
      JAVA_HOME: {
        name: 'JAVA_HOME',
        value: javaHome,
        description: 'JDK used to build and run this service',
        isRequired: false,
      },
    };

    const response = await fetch(`/api/services/${serviceId}/env-vars`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ envVars }),
    });

    if (!response.ok) {
      throw new Error(`Failed to set JAVA_HOME: ${response.status} ${response.statusText}`);
    }
  }

  /**
   * Build information for the running binary.
   */
  static async getBuildInfo(): Promise<BuildInfo> {
    const response = await fetch('/api/version');

    if (!response.ok) {
      throw new Error(`Failed to get version: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }
}
