import { Service } from "@/types";

export interface ServiceLoadingStates {
  [serviceName: string]: {
    starting?: boolean;
    stopping?: boolean;
    restarting?: boolean;
    checkingHealth?: boolean;
    installingLibraries?: boolean;
    validatingWrapper?: boolean;
    generatingWrapper?: boolean;
    repairingWrapper?: boolean;
  };
}

export interface ServiceOperationResult {
  success: boolean;
  message?: string;
  error?: string;
}

export class ServiceOperations {
  static async startService(
    serviceId: string,
  ): Promise<ServiceOperationResult> {
    try {
      const response = await fetch(`/api/services/${serviceId}/start`, {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to start service: ${response.status} ${response.statusText}`,
        );
      }
      const result = await response.json();
      const serviceName = result.service?.name || serviceId;
      return {
        success: true,
        message: `${serviceName} is now starting up`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      };
    }
  }

  static async stopService(serviceId: string): Promise<ServiceOperationResult> {
    try {
      const response = await fetch(`/api/services/${serviceId}/stop`, {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to stop service: ${response.status} ${response.statusText}`,
        );
      }
      const result = await response.json();
      const serviceName = result.service?.name || serviceId;
      return {
        success: true,
        message: `${serviceName} is shutting down`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      };
    }
  }

  static async restartService(
    serviceId: string,
  ): Promise<ServiceOperationResult> {
    try {
      const response = await fetch(`/api/services/${serviceId}/restart`, {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to restart service: ${response.status} ${response.statusText}`,
        );
      }
      const result = await response.json();
      const serviceName = result.service?.name || serviceId;
      return {
        success: true,
        message: `${serviceName} is being restarted`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      };
    }
  }

  static async checkServiceHealth(
    serviceId: string,
  ): Promise<ServiceOperationResult> {
    try {
      const response = await fetch(`/api/services/${serviceId}/health`, {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to check service health: ${response.status} ${response.statusText}`,
        );
      }
      const result = await response.json();
      const serviceName = result.service?.name || serviceId;
      return {
        success: true,
        message: `Checking ${serviceName} health status`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      };
    }
  }

  static async installLibraries(
    service: Service,
  ): Promise<ServiceOperationResult> {
    try {
      const response = await fetch(
        `/api/services/${service.id}/install-libraries`,
        {
          method: "POST",
        },
      );

      if (!response.ok) {
        throw new Error(
          `Failed to install libraries: ${response.status} ${response.statusText}`,
        );
      }

      await response.json();
      return {
        success: true,
        message: `Installing Maven libraries for ${service.name}. Check logs for progress.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      };
    }
  }

  static async startAllServices(
    token?: string,
  ): Promise<ServiceOperationResult> {
    try {
      const headers: Record<string, string> = {};

      // Only include Authorization header if token exists
      if (token) {
        headers["Authorization"] = `Bearer ${token}`;
      }

      const response = await fetch("/api/services/start-all", {
        method: "POST",
        headers,
      });
      if (!response.ok) {
        throw new Error(
          `Failed to start all services: ${response.status} ${response.statusText}`,
        );
      }

      const result = await response.json();
      return {
        success: true,
        message: result.status || "Services are being started",
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      };
    }
  }

  static async stopAllServices(
    token?: string,
  ): Promise<ServiceOperationResult> {
    try {
      const headers: Record<string, string> = {};

      // Only include Authorization header if token exists
      if (token) {
        headers["Authorization"] = `Bearer ${token}`;
      }

      const response = await fetch("/api/services/stop-all", {
        method: "POST",
        headers,
      });
      if (!response.ok) {
        throw new Error(
          `Failed to stop all services: ${response.status} ${response.statusText}`,
        );
      }

      const result = await response.json();
      return {
        success: true,
        message: result.status || "Services are being stopped",
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      };
    }
  }

  static async fixLombok(): Promise<ServiceOperationResult> {
    try {
      const response = await fetch("/api/services/fix-lombok", {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error(
          `Failed to fix Lombok: ${response.status} ${response.statusText}`,
        );
      }

      const result = await response.json();
      const successCount = Object.values(result.results).filter(
        (r: any) => r === "Success",
      ).length;
      const errorCount = Object.values(result.results).filter(
        (r: any) => r !== "Success",
      ).length;

      return {
        success: true,
        message:
          errorCount > 0
            ? `Successfully processed ${successCount} services. ${errorCount} services had errors.`
            : `Successfully processed ${successCount} services.`,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "Failed to fix Lombok",
      };
    }
  }

  static async syncEnvironment(): Promise<ServiceOperationResult> {
    try {
      const response = await fetch("/api/environment/sync", {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error(
          `Failed to sync environment: ${response.status} ${response.statusText}`,
        );
      }

      const result = await response.json();
      return {
        success: result.success,
        message: result.success
          ? `Successfully loaded ${result.variablesSet} environment variables.${result.errors?.length > 0 ? ` ${result.errors.length} warnings occurred.` : ""}`
          : `Partially loaded ${result.variablesSet} environment variables. ${result.errors?.length || 0} warnings occurred.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to sync environment",
      };
    }
  }

  static async deleteService(
    serviceId: string,
  ): Promise<ServiceOperationResult> {
    try {
      const response = await fetch(`/api/services/${serviceId}`, {
        method: "DELETE",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to delete service: ${response.status} ${response.statusText}`,
        );
      }
      const result = await response.json();
      const serviceName = result.service?.name || serviceId;
      return {
        success: true,
        message: `${serviceName} has been deleted permanently`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      };
    }
  }

  static async fetchServices(): Promise<Service[]> {
    try {
      const response = await fetch("/api/services");
      if (!response.ok) {
        throw new Error(
          `Failed to fetch services: ${response.status} ${response.statusText}`,
        );
      }
      const data = await response.json();
      return data.sort((a: Service, b: Service) => a.order - b.order);
    } catch (error) {
      console.error("Failed to fetch services:", error);
      throw error;
    }
  }

  static async clearServiceLogs(
    serviceId: string,
  ): Promise<ServiceOperationResult> {
    try {
      const response = await fetch(`/api/services/${serviceId}/logs`, {
        method: "DELETE",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to clear logs: ${response.status} ${response.statusText}`,
        );
      }
      const result = await response.json();
      const serviceName = result.service?.name || serviceId;
      return {
        success: true,
        message: `Logs for ${serviceName} have been cleared`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      };
    }
  }

  static async clearAllLogs(
    serviceNames?: string[],
  ): Promise<ServiceOperationResult> {
    try {
      const response = await fetch("/api/services/logs/clear", {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          serviceNames: serviceNames || [],
        }),
      });
      if (!response.ok) {
        throw new Error(
          `Failed to clear logs: ${response.status} ${response.statusText}`,
        );
      }
      const result = await response.json();
      return {
        success: true,
        message: result.message || "Logs have been cleared",
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      };
    }
  }

  // Wrapper Management Operations

  static async validateWrapper(
    serviceId: string,
  ): Promise<ServiceOperationResult & { data?: any }> {
    try {
      const response = await fetch(`/api/services/${serviceId}/wrapper/validate`);
      if (!response.ok) {
        throw new Error(
          `Failed to validate wrapper: ${response.status} ${response.statusText}`,
        );
      }
      const result = await response.json();
      return {
        success: true,
        message: result.isValid 
          ? `${result.buildSystem} wrapper is valid for ${result.serviceName}`
          : `${result.buildSystem} wrapper validation failed for ${result.serviceName}`,
        data: result,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred while validating wrapper",
      };
    }
  }

  static async generateWrapper(
    serviceId: string,
  ): Promise<ServiceOperationResult> {
    try {
      const response = await fetch(`/api/services/${serviceId}/wrapper/generate`, {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to generate wrapper: ${response.status} ${response.statusText}`,
        );
      }
      const result = await response.json();
      return {
        success: true,
        message: result.message || `Successfully generated ${result.buildSystem} wrapper`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred while generating wrapper",
      };
    }
  }

  static async repairWrapper(
    serviceId: string,
  ): Promise<ServiceOperationResult> {
    try {
      const response = await fetch(`/api/services/${serviceId}/wrapper/repair`, {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to repair wrapper: ${response.status} ${response.statusText}`,
        );
      }
      const result = await response.json();
      return {
        success: true,
        message: result.message || `Successfully repaired ${result.buildSystem} wrapper`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred while repairing wrapper",
      };
    }
  }
}
