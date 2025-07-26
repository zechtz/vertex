import { formatBytes, formatPercentage } from '../../utils/formatters';

interface ResourceMetricsProps {
  cpuPercent: number;
  memoryUsage: number;
  memoryPercent: number;
  diskUsage: number;
  networkRx: number;
  networkTx: number;
  className?: string;
}

export function ResourceMetrics({
  cpuPercent,
  memoryUsage,
  memoryPercent,
  diskUsage,
  networkRx,
  networkTx,
  className = '',
}: ResourceMetricsProps) {
  const getCPUColor = (cpu: number) => {
    if (cpu > 80) return 'text-red-600';
    if (cpu > 60) return 'text-yellow-600';
    return 'text-green-600';
  };

  const getMemoryColor = (memory: number) => {
    if (memory > 80) return 'text-red-600';
    if (memory > 60) return 'text-yellow-600';
    return 'text-green-600';
  };

  return (
    <div className={`grid grid-cols-2 gap-2 text-xs ${className}`}>
      {/* CPU Usage */}
      <div className="flex items-center space-x-1">
        <svg className="w-3 h-3 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
          <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
        </svg>
        <span className="text-gray-600">CPU:</span>
        <span className={`font-medium ${getCPUColor(cpuPercent)}`}>
          {formatPercentage(cpuPercent)}
        </span>
      </div>

      {/* Memory Usage */}
      <div className="flex items-center space-x-1">
        <svg className="w-3 h-3 text-purple-500" fill="currentColor" viewBox="0 0 20 20">
          <path d="M3 7v10a2 2 0 002 2h10a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2 2v0zM9 9a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V9z" />
        </svg>
        <span className="text-gray-600">RAM:</span>
        <span className={`font-medium ${getMemoryColor(memoryPercent)}`}>
          {formatBytes(memoryUsage)} ({formatPercentage(memoryPercent)})
        </span>
      </div>

      {/* Disk I/O */}
      <div className="flex items-center space-x-1">
        <svg className="w-3 h-3 text-gray-500" fill="currentColor" viewBox="0 0 20 20">
          <path d="M4 3a2 2 0 100 4h12a2 2 0 100-4H4z" />
          <path fillRule="evenodd" d="M3 8a2 2 0 012-2v9a2 2 0 01-2-2V8zM15 6a2 2 0 012 2v7a2 2 0 01-2 2V6z" clipRule="evenodd" />
        </svg>
        <span className="text-gray-600">Disk:</span>
        <span className="font-medium text-gray-700">
          {formatBytes(diskUsage)}
        </span>
      </div>

      {/* Network I/O */}
      <div className="flex items-center space-x-1">
        <svg className="w-3 h-3 text-green-500" fill="currentColor" viewBox="0 0 20 20">
          <path d="M2 5a2 2 0 012-2h7a2 2 0 012 2v4a2 2 0 01-2 2H9l-3 3v-3H4a2 2 0 01-2-2V5z" />
        </svg>
        <span className="text-gray-600">Net:</span>
        <div className="flex space-x-1">
          <span className="font-medium text-green-600" title="Network RX">
            ↓{networkRx > 1000 ? Math.round(networkRx / 1000) + 'K' : networkRx}
          </span>
          <span className="font-medium text-blue-600" title="Network TX">
            ↑{networkTx > 1000 ? Math.round(networkTx / 1000) + 'K' : networkTx}
          </span>
        </div>
      </div>
    </div>
  );
}

export default ResourceMetrics;