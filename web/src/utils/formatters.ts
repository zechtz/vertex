/**
 * Format bytes into human readable format
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

/**
 * Format percentage values
 */
export function formatPercentage(value: number): string {
  if (isNaN(value) || value === 0) return '0.0%';
  return value.toFixed(1) + '%';
}

/**
 * Format duration in milliseconds to human readable format
 */
export function formatDuration(ms: number): string {
  if (ms < 1000) return ms + 'ms';
  if (ms < 60000) return (ms / 1000).toFixed(1) + 's';
  if (ms < 3600000) return (ms / 60000).toFixed(1) + 'm';
  return (ms / 3600000).toFixed(1) + 'h';
}

/**
 * Format numbers with thousand separators
 */
export function formatNumber(num: number): string {
  return num.toLocaleString();
}

/**
 * Format uptime string to be more readable
 */
export function formatUptime(uptime: string): string {
  if (!uptime) return 'N/A';
  return uptime;
}

/**
 * Get color class based on health status
 */
export function getHealthStatusColor(status: string): string {
  switch (status.toLowerCase()) {
    case 'healthy':
      return 'text-green-600 bg-green-100';
    case 'unhealthy':
      return 'text-red-600 bg-red-100';
    case 'running':
      return 'text-blue-600 bg-blue-100';
    case 'unknown':
    default:
      return 'text-gray-600 bg-gray-100';
  }
}

/**
 * Get color class based on service status
 */
export function getServiceStatusColor(status: string): string {
  switch (status.toLowerCase()) {
    case 'running':
      return 'text-green-600 bg-green-100';
    case 'stopped':
      return 'text-red-600 bg-red-100';
    case 'starting':
      return 'text-yellow-600 bg-yellow-100';
    default:
      return 'text-gray-600 bg-gray-100';
  }
}