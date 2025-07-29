import { useState, useEffect } from 'react';
import { formatBytes, formatPercentage, formatNumber } from '../../utils/formatters';
import { Service } from '@/types';

interface SystemSummary {
  runningServices: number;
  totalServices: number;
  totalCPU: number;
  totalMemory: number;
  timestamp: string;
}

interface ServiceMetric {
  name: string;
  cpuPercent: number;
  memoryUsage: number;
  memoryPercent: number;
  diskUsage: number;
  networkRx: number;
  networkTx: number;
  status: string;
  healthStatus: string;
  uptime: string;
  errorRate: number;
  requestCount: number;
}

interface SystemMetricsData {
  summary: SystemSummary;
  services: ServiceMetric[];
}

interface SystemMetricsProps {
  className?: string;
  services: Service[];
}

export function SystemMetrics({ className = '', services }: SystemMetricsProps) {
  const [metricsData, setMetricsData] = useState<SystemMetricsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const calculateMetrics = () => {
      try {
        // Calculate summary from provided services
        const runningServices = services.filter(service => service.status === 'running').length;
        const totalServices = services.length;
        
        // Calculate totals from running services
        let totalCPU = 0;
        let totalMemory = 0;
        const runningServicesList = services.filter(service => service.status === 'running');
        
        runningServicesList.forEach(service => {
          totalCPU += service.cpuPercent || 0;
          totalMemory += service.memoryUsage || 0;
        });

        const summary: SystemSummary = {
          runningServices,
          totalServices,
          totalCPU,
          totalMemory,
          timestamp: new Date().toISOString()
        };

        // Convert services to metrics format
        const serviceMetrics: ServiceMetric[] = runningServicesList.map(service => ({
          name: service.name,
          cpuPercent: service.cpuPercent || 0,
          memoryUsage: service.memoryUsage || 0,
          memoryPercent: service.memoryPercent || 0,
          diskUsage: service.diskUsage || 0,
          networkRx: service.networkRx || 0,
          networkTx: service.networkTx || 0,
          status: service.status,
          healthStatus: service.healthStatus,
          uptime: service.uptime || '',
          errorRate: service.metrics?.errorRate || 0,
          requestCount: service.metrics?.requestCount || 0,
        }));

        setMetricsData({
          summary,
          services: serviceMetrics
        });
        setError(null);
      } catch (err) {
        console.error('Failed to calculate metrics:', err);
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    // Calculate initial metrics
    calculateMetrics();

    // Update metrics whenever services change
    const interval = setInterval(calculateMetrics, 10000);

    return () => clearInterval(interval);
  }, [services]);

  if (loading) {
    return (
      <div className={`bg-white dark:bg-gray-800 rounded-lg shadow p-6 ${className}`}>
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/4 mb-4"></div>
          <div className="space-y-3">
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded"></div>
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-5/6"></div>
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4"></div>
          </div>
        </div>
      </div>
    );
  }

  if (error || !metricsData) {
    return (
      <div className={`bg-white dark:bg-gray-800 rounded-lg shadow p-6 ${className}`}>
        <div className="flex items-center space-x-2 text-red-600">
          <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
          </svg>
          <span>Failed to load system metrics: {error}</span>
        </div>
      </div>
    );
  }

  const { summary, services: serviceMetrics } = metricsData;

  return (
    <div className={`bg-white dark:bg-gray-800 rounded-lg shadow ${className}`}>
      {/* Header */}
      <div className="p-6 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 dark:text-gray-100">System Overview</h3>
          <div className="text-sm text-gray-500 dark:text-gray-400">
            Last updated: {new Date(summary.timestamp).toLocaleTimeString()}
          </div>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="p-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          {/* Running Services */}
          <div className="bg-green-50 dark:bg-green-900/20 rounded-lg p-4">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-green-800 dark:text-green-200">Running Services</p>
                <p className="text-2xl font-semibold text-green-900 dark:text-green-100">
                  {summary.runningServices}/{summary.totalServices}
                </p>
              </div>
            </div>
          </div>

          {/* Total CPU */}
          <div className="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-4">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <svg className="w-8 h-8 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-blue-800 dark:text-blue-200">Total CPU</p>
                <p className="text-2xl font-semibold text-blue-900 dark:text-blue-100">
                  {formatPercentage(summary.totalCPU)}
                </p>
              </div>
            </div>
          </div>

          {/* Total Memory */}
          <div className="bg-purple-50 dark:bg-purple-900/20 rounded-lg p-4">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <svg className="w-8 h-8 text-purple-600" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M3 7v10a2 2 0 002 2h10a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2 2v0zM9 9a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V9z" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-purple-800 dark:text-purple-200">Total Memory</p>
                <p className="text-2xl font-semibold text-purple-900 dark:text-purple-100">
                  {formatBytes(summary.totalMemory)}
                </p>
              </div>
            </div>
          </div>

          {/* System Health */}
          <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <svg className="w-8 h-8 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-gray-800 dark:text-gray-200">System Health</p>
                <p className="text-2xl font-semibold text-gray-900 dark:text-gray-100">
                  {summary.runningServices === summary.totalServices ? 'Optimal' : 'Warning'}
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Service Details Table */}
        <div>
          <h4 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Service Resource Usage</h4>
          {serviceMetrics.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No running services to display metrics for.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Service
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      CPU
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Memory
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Network
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Performance
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                  {serviceMetrics.map((service) => (
                    <tr key={service.name} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <div className="text-sm font-medium text-gray-900 dark:text-gray-100">{service.name}</div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-gray-900 dark:text-gray-100">{formatPercentage(service.cpuPercent)}</div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-gray-900 dark:text-gray-100">
                          {formatBytes(service.memoryUsage)}
                          <div className="text-xs text-gray-500">
                            ({formatPercentage(service.memoryPercent)})
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-gray-900 dark:text-gray-100">
                          <div className="flex space-x-2">
                            <span className="text-green-600">↓{formatNumber(service.networkRx)}</span>
                            <span className="text-blue-600">↑{formatNumber(service.networkTx)}</span>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-gray-900 dark:text-gray-100">
                          <div>Errors: {formatPercentage(service.errorRate)}</div>
                          <div className="text-xs text-gray-500">
                            Requests: {formatNumber(service.requestCount)}
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex flex-col space-y-1">
                          <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            service.status === 'running' ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-200' : 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-200'
                          }`}>
                            {service.status}
                          </span>
                          <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            service.healthStatus === 'healthy' 
                              ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-200' 
                              : service.healthStatus === 'unhealthy'
                              ? 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-200'
                              : 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200'
                          }`}>
                            {service.healthStatus}
                          </span>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default SystemMetrics;