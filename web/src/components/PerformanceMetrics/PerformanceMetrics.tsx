import { formatDuration, formatNumber, formatPercentage } from '../../utils/formatters';

interface ResponseTime {
  timestamp: string;
  duration: number; // in nanoseconds
  status: number;
}

interface ServiceMetrics {
  responseTimes: ResponseTime[];
  errorRate: number;
  requestCount: number;
  lastChecked: string;
}

interface PerformanceMetricsProps {
  metrics: ServiceMetrics;
  className?: string;
}

export function PerformanceMetrics({ metrics, className = '' }: PerformanceMetricsProps) {
  // Calculate average response time from recent requests
  const getAverageResponseTime = () => {
    if (!metrics.responseTimes || metrics.responseTimes.length === 0) return 0;
    
    const recentResponses = metrics.responseTimes.slice(-10); // Last 10 requests
    const totalTime = recentResponses.reduce((sum, rt) => sum + (rt.duration / 1000000), 0); // Convert to milliseconds
    return totalTime / recentResponses.length;
  };

  // Get latest response time
  const getLatestResponseTime = () => {
    if (!metrics.responseTimes || metrics.responseTimes.length === 0) return 0;
    const latest = metrics.responseTimes[metrics.responseTimes.length - 1];
    return latest.duration / 1000000; // Convert to milliseconds
  };

  // Get response time trend (better/worse compared to average)
  const getResponseTimeTrend = () => {
    const avg = getAverageResponseTime();
    const latest = getLatestResponseTime();
    
    if (avg === 0 || latest === 0) return 'stable';
    
    const difference = ((latest - avg) / avg) * 100;
    if (difference > 20) return 'worse';
    if (difference < -20) return 'better';
    return 'stable';
  };

  const avgResponseTime = getAverageResponseTime();
  const latestResponseTime = getLatestResponseTime();
  const trend = getResponseTimeTrend();

  const getErrorRateColor = (rate: number) => {
    if (rate > 10) return 'text-red-600';
    if (rate > 5) return 'text-yellow-600';
    return 'text-green-600';
  };

  const getResponseTimeColor = (time: number) => {
    if (time > 1000) return 'text-red-600';
    if (time > 500) return 'text-yellow-600';
    return 'text-green-600';
  };

  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'better':
        return <span className="text-green-500">↗</span>;
      case 'worse':
        return <span className="text-red-500">↘</span>;
      default:
        return <span className="text-gray-500">→</span>;
    }
  };

  return (
    <div className={`bg-gray-50 rounded-lg p-3 space-y-2 ${className}`}>
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-gray-700">Performance Metrics</h4>
        <div className="text-xs text-gray-500">
          {metrics.lastChecked && new Date(metrics.lastChecked).toLocaleTimeString()}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 text-xs">
        {/* Response Time */}
        <div className="space-y-1">
          <div className="flex items-center space-x-1">
            <svg className="w-3 h-3 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
            </svg>
            <span className="text-gray-600">Response Time</span>
            {getTrendIcon(trend)}
          </div>
          <div className="space-y-0.5">
            <div className="flex justify-between">
              <span className="text-gray-500">Latest:</span>
              <span className={`font-medium ${getResponseTimeColor(latestResponseTime)}`}>
                {formatDuration(latestResponseTime)}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-500">Avg:</span>
              <span className={`font-medium ${getResponseTimeColor(avgResponseTime)}`}>
                {formatDuration(avgResponseTime)}
              </span>
            </div>
          </div>
        </div>

        {/* Error Rate & Requests */}
        <div className="space-y-1">
          <div className="flex items-center space-x-1">
            <svg className="w-3 h-3 text-orange-500" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M3 10a7 7 0 019.307-6.611 1 1 0 00.658-1.889 9 9 0 105.98 7.501 1 1 0 00-1.988.22A7 7 0 113 10z" clipRule="evenodd" />
              <path fillRule="evenodd" d="M13.5 3a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0z" clipRule="evenodd" />
            </svg>
            <span className="text-gray-600">Quality</span>
          </div>
          <div className="space-y-0.5">
            <div className="flex justify-between">
              <span className="text-gray-500">Error Rate:</span>
              <span className={`font-medium ${getErrorRateColor(metrics.errorRate)}`}>
                {formatPercentage(metrics.errorRate)}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-500">Requests:</span>
              <span className="font-medium text-gray-700">
                {formatNumber(metrics.requestCount)}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Response Time Sparkline (simplified) */}
      {metrics.responseTimes && metrics.responseTimes.length > 0 && (
        <div className="mt-2">
          <div className="flex items-center justify-between text-xs text-gray-500 mb-1">
            <span>Recent Response Times</span>
            <span>{metrics.responseTimes.length} samples</span>
          </div>
          <div className="flex items-end space-x-0.5 h-8">
            {metrics.responseTimes.slice(-20).map((rt, index) => {
              const height = Math.max(2, Math.min(32, (rt.duration / 1000000) / 10)); // Scale to fit
              const isError = rt.status >= 400 || rt.status === 0;
              return (
                <div
                  key={index}
                  className={`w-1 rounded-t ${isError ? 'bg-red-400' : 'bg-blue-400'}`}
                  style={{ height: `${height}px` }}
                  title={`${formatDuration(rt.duration / 1000000)} - Status: ${rt.status}`}
                />
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

export default PerformanceMetrics;