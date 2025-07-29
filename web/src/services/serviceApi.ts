import { Service, Configuration } from '@/types';

export interface ApiResponse<T = any> {
  data?: T;
  error?: string;
  status: number;
}

/**
 * Service API functions for data operations
 */
export class ServiceApi {

  /**
   * Fetch all services
   */
  static async fetchServices(): Promise<Service[]> {
    const response = await fetch('/api/services');
    if (!response.ok) {
      throw new Error(`Failed to fetch services: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }

  /**
   * Fetch all configurations
   */
  static async fetchConfigurations(): Promise<Configuration[]> {
    const response = await fetch('/api/configurations');
    if (!response.ok) {
      throw new Error(`Failed to fetch configurations: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }

  /**
   * Create a new service
   */
  static async createService(serviceData: Partial<Service>): Promise<Service> {
    const response = await fetch('/api/services', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(serviceData),
    });
    
    if (!response.ok) {
      throw new Error(`Failed to create service: ${response.status} ${response.statusText}`);
    }
    
    return response.json();
  }

  /**
   * Update an existing service
   */
  static async updateService(serviceName: string, serviceData: Partial<Service>): Promise<Service> {
    const response = await fetch(`/api/services/${serviceName}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(serviceData),
    });
    
    if (!response.ok) {
      throw new Error(`Failed to update service: ${response.status} ${response.statusText}`);
    }
    
    return response.json();
  }

  /**
   * Delete a service
   */
  static async deleteService(serviceName: string): Promise<void> {
    const response = await fetch(`/api/services/${serviceName}`, {
      method: 'DELETE',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to delete service: ${response.status} ${response.statusText}`);
    }
  }

  /**
   * Get service logs
   */
  static async getServiceLogs(serviceName: string): Promise<string[]> {
    const response = await fetch(`/api/services/${serviceName}/logs`);
    if (!response.ok) {
      throw new Error(`Failed to fetch logs: ${response.status} ${response.statusText}`);
    }
    const data = await response.json();
    return data.logs || [];
  }

  /**
   * Copy service logs
   */
  static async copyServiceLogs(serviceName: string): Promise<void> {
    const response = await fetch(`/api/services/${serviceName}/logs/copy`, {
      method: 'POST',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to copy logs: ${response.status} ${response.statusText}`);
    }
  }

  /**
   * Clear service logs
   */
  static async clearServiceLogs(serviceName: string): Promise<void> {
    const response = await fetch(`/api/services/${serviceName}/logs/clear`, {
      method: 'POST',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to clear logs: ${response.status} ${response.statusText}`);
    }
  }
}