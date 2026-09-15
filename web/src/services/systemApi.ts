import type { BuildInfo, EnvVar, InstalledJdk } from '@/types';

import { apiFetch } from "@/services/apiFetch";
/**
 * System-level API functions
 */
export class SystemApi {
  /**
   * Fix Lombok issues
   */
  static async fixLombok(): Promise<void> {
    const response = await apiFetch('/api/services/fix-lombok', {
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
    const response = await apiFetch('/api/environment/setup', {
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
    const response = await apiFetch('/api/logs/copy-all', {
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
    const response = await apiFetch('/api/system/logs/cleanup', {
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
    const response = await apiFetch('/api/java/jdks');

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

    const response = await apiFetch(`/api/services/${serviceId}/env-vars`, {
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
    const response = await apiFetch('/api/version');

    if (!response.ok) {
      throw new Error(`Failed to get version: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }

  /**
   * Extend the current session. Requires a still-valid token; an expired one
   * must go back through login.
   */
  static async refreshSession(): Promise<{ token: string; user: unknown }> {
    // Plain fetch: a failure here means the session is gone, and parking this
    // request for replay after login would be circular.
    const token = localStorage.getItem('authToken');
    const response = await fetch('/api/auth/refresh', {
      method: 'POST',
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });

    if (!response.ok) {
      throw new Error(`Failed to refresh session: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }
}
