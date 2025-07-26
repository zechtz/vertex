export interface EnvVar {
  name: string;
  value: string;
  description: string;
  isRequired: boolean;
}

export interface ResponseTime {
  timestamp: string;
  duration: number; // in nanoseconds
  status: number;
}

export interface ServiceMetrics {
  responseTimes: ResponseTime[];
  errorRate: number;
  requestCount: number;
  lastChecked: string;
}

export interface Service {
  name: string;
  dir: string;
  extraEnv: string;
  javaOpts: string;
  status: string;
  healthStatus: string;
  healthUrl: string;
  port: number;
  pid: number;
  order: number;
  lastStarted: string;
  uptime: string;
  description: string;
  isEnabled: boolean;
  envVars: { [key: string]: EnvVar };
  logs: LogEntry[];
  // Resource monitoring fields
  cpuPercent?: number;
  memoryUsage?: number;   // in bytes
  memoryPercent?: number;
  diskUsage?: number;     // in bytes
  networkRx?: number;     // bytes received
  networkTx?: number;     // bytes transmitted
  metrics?: ServiceMetrics;
}

export interface LogEntry {
  timestamp: string;
  level: string;
  message: string;
}

export interface ServiceConfigRequest {
  name: string;
  dir: string;
  javaOpts: string;
  healthUrl: string;
  port: number;
  order: number;
  description: string;
  isEnabled: boolean;
  envVars: Record<string, EnvVar>;
}

export interface Configuration {
  id: string;
  name: string;
  services: Array<{
    name: string;
    order: number;
  }>;
  isDefault?: boolean;
}