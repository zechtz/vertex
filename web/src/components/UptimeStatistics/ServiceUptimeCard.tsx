import React from 'react';
import { Clock, Zap, Activity, TrendingUp, Server } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { UptimeProgressBar } from './UptimeProgressBar';
import { UptimeStatistics } from '@/types';

interface ServiceUptimeCardProps {
  service: {
    serviceName: string;
    serviceId: string;
    port: number;
    status: string;
    healthStatus: string;
    stats: UptimeStatistics;
  };
  onClick?: () => void;
}

export const ServiceUptimeCard: React.FC<ServiceUptimeCardProps> = ({
  service,
  onClick
}) => {
  const formatDuration = (nanoseconds: number): string => {
    if (nanoseconds === 0) return "None";

    const totalSeconds = nanoseconds / 1_000_000_000;
    const days = Math.floor(totalSeconds / 86400);
    const hours = Math.floor((totalSeconds % 86400) / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);

    if (days > 0) {
      return `${days}d ${hours}h`;
    } else if (hours > 0) {
      return `${hours}h ${minutes}m`;
    } else if (minutes > 0) {
      return `${minutes}m`;
    } else {
      return `${Math.floor(totalSeconds)}s`;
    }
  };

  const getStatusBadgeVariant = (status: string, healthStatus: string) => {
    if (status === "running" && healthStatus === "healthy") return "default";
    if (status === "running" && healthStatus === "unhealthy") return "destructive";
    if (status === "running") return "secondary";
    return "outline";
  };

  const getAvailabilityRating = (uptime: number) => {
    if (uptime >= 99.9) return { rating: "Excellent", color: "text-green-600" };
    if (uptime >= 99) return { rating: "Good", color: "text-blue-600" };
    if (uptime >= 95) return { rating: "Fair", color: "text-yellow-600" };
    return { rating: "Poor", color: "text-red-600" };
  };

  const availabilityRating = getAvailabilityRating(service.stats.uptimePercentage7d);

  return (
    <Card 
      className={`transition-all duration-200 hover:shadow-lg ${onClick ? 'cursor-pointer hover:border-blue-300 dark:hover:border-blue-600' : ''}`}
      onClick={onClick}
    >
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Server className="w-5 h-5 text-blue-500" />
            <span className="font-semibold text-lg">{service.serviceName}</span>
          </div>
          <Badge
            variant={getStatusBadgeVariant(service.status, service.healthStatus)}
            className="text-xs"
          >
            {service.status === "running" ? service.healthStatus : service.status}
          </Badge>
        </CardTitle>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Port {service.port} • {availabilityRating.rating}
        </p>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Uptime Progress Bars */}
        <div className="space-y-3">
          <div>
            <div className="flex justify-between items-center mb-1">
              <span className="text-sm font-medium flex items-center">
                <Clock className="w-3 h-3 mr-1" />
                24h Uptime
              </span>
              <span className="text-xs text-gray-500">
                {service.stats.uptimePercentage24h.toFixed(1)}%
              </span>
            </div>
            <UptimeProgressBar 
              percentage={service.stats.uptimePercentage24h}
              size="sm"
              showPercentage={false}
            />
          </div>

          <div>
            <div className="flex justify-between items-center mb-1">
              <span className="text-sm font-medium flex items-center">
                <TrendingUp className="w-3 h-3 mr-1" />
                7d Uptime
              </span>
              <span className="text-xs text-gray-500">
                {service.stats.uptimePercentage7d.toFixed(1)}%
              </span>
            </div>
            <UptimeProgressBar 
              percentage={service.stats.uptimePercentage7d}
              size="sm"
              showPercentage={false}
            />
          </div>
        </div>

        {/* Key Metrics */}
        <div className="grid grid-cols-2 gap-4 pt-3 border-t border-gray-100 dark:border-gray-700">
          <div className="text-center">
            <div className="flex items-center justify-center mb-1">
              <Zap className="w-4 h-4 text-orange-500 mr-1" />
              <span className="text-xs font-medium text-gray-600 dark:text-gray-400">Restarts</span>
            </div>
            <div className="text-lg font-bold text-gray-900 dark:text-gray-100">
              {service.stats.totalRestarts}
            </div>
          </div>

          <div className="text-center">
            <div className="flex items-center justify-center mb-1">
              <Activity className="w-4 h-4 text-purple-500 mr-1" />
              <span className="text-xs font-medium text-gray-600 dark:text-gray-400">Downtime</span>
            </div>
            <div className="text-lg font-bold text-gray-900 dark:text-gray-100">
              {formatDuration(service.stats.totalDowntime24h)}
            </div>
          </div>
        </div>

        {/* MTBF */}
        <div className="pt-2 border-t border-gray-100 dark:border-gray-700">
          <div className="flex justify-between items-center">
            <span className="text-xs font-medium text-gray-600 dark:text-gray-400">
              Mean Time Between Failures
            </span>
            <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
              {formatDuration(service.stats.mtbf)}
            </span>
          </div>
        </div>

        {/* Availability Rating */}
        <div className="flex justify-center">
          <div className={`text-sm font-medium ${availabilityRating.color}`}>
            {availabilityRating.rating} Availability
          </div>
        </div>
      </CardContent>
    </Card>
  );
};