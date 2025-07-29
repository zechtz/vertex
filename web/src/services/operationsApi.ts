/**
 * Service Operations API functions
 */
export class OperationsApi {
  private static getAuthHeaders(token?: string | null): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    return headers;
  }

  /**
   * Start a service
   */
  static async startService(serviceName: string): Promise<void> {
    const response = await fetch(`/api/services/${serviceName}/start`, {
      method: 'POST',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to start service: ${response.status} ${response.statusText}`);
    }
  }

  /**
   * Stop a service
   */
  static async stopService(serviceName: string): Promise<void> {
    const response = await fetch(`/api/services/${serviceName}/stop`, {
      method: 'POST',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to stop service: ${response.status} ${response.statusText}`);
    }
  }

  /**
   * Restart a service
   */
  static async restartService(serviceName: string): Promise<void> {
    const response = await fetch(`/api/services/${serviceName}/restart`, {
      method: 'POST',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to restart service: ${response.status} ${response.statusText}`);
    }
  }

  /**
   * Check service health
   */
  static async checkServiceHealth(serviceName: string): Promise<void> {
    const response = await fetch(`/api/services/${serviceName}/health`, {
      method: 'POST',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to check health: ${response.status} ${response.statusText}`);
    }
  }

  /**
   * Install service libraries
   */
  static async installServiceLibraries(serviceName: string): Promise<void> {
    const response = await fetch(`/api/services/${serviceName}/install-libraries`, {
      method: 'POST',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to install libraries: ${response.status} ${response.statusText}`);
    }
  }

  /**
   * Start all services (profile-aware)
   */
  static async startAllServices(token?: string | null): Promise<{ status: string; profile?: string }> {
    const response = await fetch('/api/services/start-all-profile', {
      method: 'POST',
      headers: this.getAuthHeaders(token),
    });
    
    if (!response.ok) {
      throw new Error(`Failed to start all services: ${response.status} ${response.statusText}`);
    }
    
    return response.json();
  }

  /**
   * Stop all services (profile-aware)
   */
  static async stopAllServices(token?: string | null): Promise<{ status: string; profile?: string }> {
    const response = await fetch('/api/services/stop-all-profile', {
      method: 'POST',
      headers: this.getAuthHeaders(token),
    });
    
    if (!response.ok) {
      throw new Error(`Failed to stop all services: ${response.status} ${response.statusText}`);
    }
    
    return response.json();
  }
}