import React, { useState, useEffect } from 'react';
import { X, Clock, Activity, AlertTriangle, TrendingUp, Server, Zap } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { UptimeProgressBar } from './UptimeProgressBar';

interface ServiceDetailModalProps {
  serviceId: string;
  serviceName: string;
  isOpen: boolean;
  onClose: () => void;
}

interface ServiceDetailData {
  serviceName: string;
  serviceId: string;
  port: number;
  status: string;
  healthStatus: string;
  stats: {
    totalRestarts: number;
    uptimePercentage24h: number;
    uptimePercentage7d: number;
    mtbf: number;
    lastDowntime: string | null;
    totalDowntime24h: number;
    totalDowntime7d: number;
  };
}

export const ServiceDetailModal: React.FC<ServiceDetailModalProps> = ({
  serviceId,
  serviceName,
  isOpen,
  onClose
}) => {
  const [serviceDetail, setServiceDetail] = useState<ServiceDetailData | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchServiceDetail = async () => {
    try {
      setIsLoading(true);
      setError(null);
      
      const token = localStorage.getItem("authToken");
      if (!token) {
        throw new Error("No authentication token");
      }

      const response = await fetch(`/api/uptime/statistics/${serviceId}`, {
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
      });
      
      if (!response.ok) {
        throw new Error(`Failed to fetch service details: ${response.statusText}`);
      }
      
      const data = await response.json();
      setServiceDetail(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to fetch service details");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen && serviceId) {
      fetchServiceDetail();
    }
  }, [isOpen, serviceId]);

  const formatDuration = (nanoseconds: number): string => {
    if (nanoseconds === 0) return "None";

    const totalSeconds = nanoseconds / 1_000_000_000;
    const days = Math.floor(totalSeconds / 86400);
    const hours = Math.floor((totalSeconds % 86400) / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);

    if (days > 0) {
      return `${days}d ${hours}h ${minutes}m`;
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

  const calculateAvailability = (uptime: number) => {
    if (uptime >= 99.9) return "Excellent";
    if (uptime >= 99) return "Good";
    if (uptime >= 95) return "Fair";
    return "Poor";
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-3">
            <Server className="w-6 h-6 text-blue-500" />
            <div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                Service Details
              </h2>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                {serviceName}
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <Activity className="w-6 h-6 animate-spin mr-2" />
              <span>Loading service details...</span>
            </div>
          ) : error ? (
            <div className="text-center py-8">
              <AlertTriangle className="w-12 h-12 text-red-500 mx-auto mb-4" />
              <h3 className="text-lg font-semibold text-red-700 mb-2">
                Error Loading Service Details
              </h3>
              <p className="text-red-600 mb-4">{error}</p>
              <Button onClick={fetchServiceDetail} variant="outline">
                <Activity className="w-4 h-4 mr-2" />
                Retry
              </Button>
            </div>
          ) : serviceDetail ? (
            <div className="space-y-6">
              {/* Service Overview */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Status</CardTitle>
                    <Activity className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-center space-x-2">
                      <Badge
                        variant={getStatusBadgeVariant(serviceDetail.status, serviceDetail.healthStatus)}
                        className="text-xs"
                      >
                        {serviceDetail.status === "running" ? serviceDetail.healthStatus : serviceDetail.status}
                      </Badge>
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">
                      Port {serviceDetail.port}
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Total Restarts</CardTitle>
                    <Zap className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{serviceDetail.stats.totalRestarts}</div>
                    <p className="text-xs text-muted-foreground">
                      All time
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">MTBF</CardTitle>
                    <Clock className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold text-sm">
                      {formatDuration(serviceDetail.stats.mtbf)}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Mean time between failures
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Availability</CardTitle>
                    <TrendingUp className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {calculateAvailability(serviceDetail.stats.uptimePercentage7d)}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      7-day rating
                    </p>
                  </CardContent>
                </Card>
              </div>

              {/* Uptime Statistics */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center">
                    <Clock className="w-5 h-5 mr-2" />
                    Uptime Statistics
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div>
                    <div className="flex justify-between items-center mb-2">
                      <span className="text-sm font-medium">24 Hour Uptime</span>
                      <span className="text-sm text-gray-500">
                        {serviceDetail.stats.uptimePercentage24h.toFixed(2)}%
                      </span>
                    </div>
                    <UptimeProgressBar 
                      percentage={serviceDetail.stats.uptimePercentage24h}
                      size="lg"
                      showPercentage={false}
                    />
                  </div>

                  <div>
                    <div className="flex justify-between items-center mb-2">
                      <span className="text-sm font-medium">7 Day Uptime</span>
                      <span className="text-sm text-gray-500">
                        {serviceDetail.stats.uptimePercentage7d.toFixed(2)}%
                      </span>
                    </div>
                    <UptimeProgressBar 
                      percentage={serviceDetail.stats.uptimePercentage7d}
                      size="lg"
                      showPercentage={false}
                    />
                  </div>
                </CardContent>
              </Card>

              {/* Downtime Analysis */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center">
                    <AlertTriangle className="w-5 h-5 mr-2" />
                    Downtime Analysis
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                      <h4 className="text-sm font-medium mb-2">24 Hour Downtime</h4>
                      <div className="text-lg font-bold text-red-600">
                        {formatDuration(serviceDetail.stats.totalDowntime24h)}
                      </div>
                    </div>
                    <div>
                      <h4 className="text-sm font-medium mb-2">7 Day Downtime</h4>
                      <div className="text-lg font-bold text-red-600">
                        {formatDuration(serviceDetail.stats.totalDowntime7d)}
                      </div>
                    </div>
                  </div>

                  {serviceDetail.stats.lastDowntime && (
                    <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
                      <h4 className="text-sm font-medium mb-2">Last Downtime</h4>
                      <div className="text-sm text-gray-600 dark:text-gray-400">
                        {new Date(serviceDetail.stats.lastDowntime).toLocaleString()}
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              No service details available
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-3 p-6 border-t border-gray-200 dark:border-gray-700">
          <Button onClick={onClose} variant="outline">
            Close
          </Button>
        </div>
      </div>
    </div>
  );
};