export interface EnvVar {
  name: string;
  value: string;
  description: string;
  isRequired: boolean;
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