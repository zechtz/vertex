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

export interface UptimeStatistics {
  totalRestarts: number;
  uptimePercentage24h: number;
  uptimePercentage7d: number;
  mtbf: number; // Mean Time Between Failures in nanoseconds
  lastDowntime: string | null;
  totalDowntime24h: number; // Duration in nanoseconds
  totalDowntime7d: number; // Duration in nanoseconds
}

export interface ServiceMetrics {
  responseTimes: ResponseTime[];
  errorRate: number;
  requestCount: number;
  lastChecked: string;
  uptimeStats: UptimeStatistics;
}

export interface ServiceDependency {
  serviceName: string;
  type: string; // "hard", "soft", "optional"
  healthCheck: boolean;
  timeout: number; // Duration in nanoseconds
  retryInterval: number; // Duration in nanoseconds
  required: boolean;
  description: string;
}

export interface Service {
  id: string; // UUID - unique identifier for the service
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
  buildSystem: string; // "maven", "gradle", or "auto"
  verboseLogging: boolean; // Enable verbose/debug logging for build tools
  envVars: { [key: string]: EnvVar };
  logs: LogEntry[];
  // Resource monitoring fields
  cpuPercent: number;
  memoryUsage: number; // in bytes
  memoryPercent: number;
  diskUsage: number; // in bytes
  networkRx: number; // bytes received
  networkTx: number; // bytes transmitted
  metrics: ServiceMetrics;
  // Service dependencies
  dependencies: ServiceDependency[] | null;
  dependentOn: string[] | null;
  startupDelay: number;
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
  buildSystem: string;
  verboseLogging: boolean;
  envVars: Record<string, EnvVar>;
}

export interface Configuration {
  id: string;
  name: string;
  services: Array<{
    id: string;
    name: string;
    order: number;
  }>;
  isDefault?: boolean;
}

// Profile Management Types
export interface UserPreferences {
  theme: string;
  language: string;
  notificationSettings: Record<string, boolean>;
  dashboardLayout: string;
  autoRefresh: boolean;
  refreshInterval: number; // seconds
}

// Profile-scoped configuration types

export interface ProfileEnvVar {
  id: number;
  profileId: string;
  varName: string;
  varValue: string;
  description: string;
  isRequired: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ProfileServiceConfig {
  id: number;
  profileId: string;
  serviceName: string;
  configKey: string;
  configValue: string;
  configType: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

export interface ProfileDependency {
  id: number;
  profileId: string;
  serviceName: string;
  dependencyServiceName: string;
  dependencyType: string;
  healthCheck: boolean;
  timeoutSeconds: number;
  retryIntervalSeconds: number;
  isRequired: boolean;
  description: string;
  createdAt: string;
  updatedAt: string;
}

export interface ProfileContext {
  profile: ServiceProfile;
  envVars: Record<string, string>;
  serviceConfigs: Record<string, Record<string, string>>;
  dependencies: Record<string, ProfileDependency[]>;
  isActive: boolean;
}

export interface UserProfile {
  userId: string;
  displayName: string;
  avatar: string;
  preferences: UserPreferences;
  createdAt: string;
  updatedAt: string;
}

export type ServiceType = {
  id: string;
  name: string;
};

export interface ServiceProfile {
  id: string;
  userId: string;
  name: string;
  description: string;
  services: ServiceType[];
  envVars: Record<string, string>;
  projectsDir: string;
  javaHomeOverride: string;
  isDefault: boolean;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateProfileRequest {
  name: string;
  description: string;
  services: string[];
  envVars: Record<string, string>;
  projectsDir: string;
  javaHomeOverride: string;
  isDefault: boolean;
  isActive?: boolean;
}

export interface UpdateProfileRequest {
  name: string;
  description: string;
  services: string[];
  envVars: Record<string, string>;
  projectsDir: string;
  javaHomeOverride: string;
  isDefault: boolean;
}

export interface UserProfileUpdateRequest {
  displayName: string;
  avatar: string;
  preferences: UserPreferences;
}
