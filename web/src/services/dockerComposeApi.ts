
export interface DockerComposePreview {
  profileId: string;
  profileName: string;
  environment: string;
  services: Array<{
    name: string;
    image?: string;
    buildContext?: string;
    ports: string[];
    environment: number;
    volumes: number;
    dependencies: string[];
  }>;
  serviceCount: number;
  networkCount: number;
  volumeCount: number;
  hasExternalDeps: boolean;
}

export interface DockerComposeGeneration {
  profileId: string;
  profileName: string;
  environment: string;
  yaml: string;
  serviceCount: number;
  networkCount: number;
  volumeCount: number;
  services: string[];
}

export interface DockerComposeConfig {
  profileId: string;
  baseImages: Record<string, string>;
  volumeMappings: Record<string, string[]>;
  networkSettings: Record<string, any>;
  resourceLimits: Record<string, {
    cpuLimit?: string;
    memoryLimit?: string;
    cpuReserve?: string;
    memoryReserve?: string;
  }>;
}

export interface DockerComposeRequest {
  environment: string;
  includeExternal: boolean;
  generateOverride: boolean;
}

class DockerComposeApiService {
  private getAuthToken(): string | null {
    return localStorage.getItem('authToken');
  }

  private async apiCall(url: string, options: RequestInit = {}): Promise<Response> {
    const token = this.getAuthToken();
    if (!token) {
      throw new Error('No authentication token available');
    }

    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
        ...options.headers,
      },
    });

    if (!response.ok) {
      if (response.status === 401) {
        throw new Error('Authentication failed. Please log in again.');
      }
      const errorText = await response.text();
      throw new Error(`API Error: ${response.status} - ${errorText}`);
    }

    return response;
  }

  async getPreview(profileId: string, environment: string = 'development'): Promise<DockerComposePreview> {
    const response = await this.apiCall(
      `/api/profiles/${profileId}/docker-compose/preview?environment=${environment}`
    );
    return response.json();
  }

  async generate(profileId: string, request: DockerComposeRequest): Promise<DockerComposeGeneration> {
    const params = new URLSearchParams({
      environment: request.environment,
      includeExternal: request.includeExternal.toString(),
      generateOverride: request.generateOverride.toString(),
    });

    const response = await this.apiCall(
      `/api/profiles/${profileId}/docker-compose?${params}`
    );
    return response.json();
  }

  async download(profileId: string, request: Omit<DockerComposeRequest, 'generateOverride'>): Promise<Blob> {
    const params = new URLSearchParams({
      environment: request.environment,
      includeExternal: request.includeExternal.toString(),
    });

    const response = await this.apiCall(
      `/api/profiles/${profileId}/docker-compose/download?${params}`
    );
    return response.blob();
  }

  async downloadOverride(profileId: string): Promise<Blob> {
    const response = await this.apiCall(
      `/api/profiles/${profileId}/docker-compose/override`
    );
    return response.blob();
  }

  async getConfig(profileId: string): Promise<DockerComposeConfig> {
    const response = await this.apiCall(`/api/profiles/${profileId}/docker-config`);
    return response.json();
  }

  async updateConfig(profileId: string, config: DockerComposeConfig): Promise<void> {
    await this.apiCall(`/api/profiles/${profileId}/docker-config`, {
      method: 'PUT',
      body: JSON.stringify(config),
    });
  }

  async deleteConfig(profileId: string): Promise<void> {
    await this.apiCall(`/api/profiles/${profileId}/docker-config`, {
      method: 'DELETE',
    });
  }
}

export const dockerComposeApi = new DockerComposeApiService();