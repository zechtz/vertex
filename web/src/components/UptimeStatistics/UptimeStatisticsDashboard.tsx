import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  RefreshCw,
  Clock,
  AlertTriangle,
  TrendingUp,
  Activity,
} from "lucide-react";
import { UptimeStatistics } from "@/types";

interface ServiceUptimeStats {
  serviceName: string;
  serviceId: string;
  port: number;
  status: string;
  healthStatus: string;
  stats: UptimeStatistics;
}

interface UptimeResponse {
  statistics: Record<string, ServiceUptimeStats>;
  summary: {
    totalServices: number;
    runningServices: number;
    unhealthyServices: number;
  };
}

export function UptimeStatisticsDashboard() {
  const [uptimeData, setUptimeData] = useState<UptimeResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchUptimeStats = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await fetch("/api/uptime/statistics");
      if (!response.ok) {
        throw new Error(
          `Failed to fetch uptime statistics: ${response.statusText}`,
        );
      }
      const data = await response.json();
      setUptimeData(data);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to fetch uptime statistics",
      );
      console.error("Error fetching uptime statistics:", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchUptimeStats();
    // Refresh every 30 seconds
    const interval = setInterval(fetchUptimeStats, 30000);
    return () => clearInterval(interval);
  }, []);

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

  const formatMTBF = (nanoseconds: number): string => {
    if (nanoseconds === 0) return "N/A";
    return formatDuration(nanoseconds);
  };

  const getUptimeColor = (percentage: number): string => {
    if (percentage >= 99) return "text-green-600";
    if (percentage >= 95) return "text-yellow-600";
    return "text-red-600";
  };

  const getStatusBadgeVariant = (status: string, healthStatus: string) => {
    if (status === "running" && healthStatus === "healthy") return "default";
    if (status === "running" && healthStatus === "unhealthy")
      return "destructive";
    if (status === "running") return "secondary";
    return "outline";
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <RefreshCw className="w-6 h-6 animate-spin mr-2" />
        <span>Loading uptime statistics...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <AlertTriangle className="w-12 h-12 text-red-500 mx-auto mb-4" />
        <h3 className="text-lg font-semibold text-red-700 mb-2">
          Error Loading Uptime Statistics
        </h3>
        <p className="text-red-600 mb-4">{error}</p>
        <Button onClick={fetchUptimeStats} variant="outline">
          <RefreshCw className="w-4 h-4 mr-2" />
          Retry
        </Button>
      </div>
    );
  }

  if (!uptimeData) {
    return <div className="p-8 text-center">No uptime data available</div>;
  }

  const services = Object.values(uptimeData.statistics);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            Service Uptime Statistics
          </h2>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Monitor service availability and reliability metrics
          </p>
        </div>
        <Button onClick={fetchUptimeStats} variant="outline" size="sm">
          <RefreshCw
            className={`w-4 h-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
          />
          Refresh
        </Button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Total Services
            </CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {uptimeData.summary.totalServices}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Running Services
            </CardTitle>
            <TrendingUp className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {uptimeData.summary.runningServices}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Unhealthy Services
            </CardTitle>
            <AlertTriangle className="h-4 w-4 text-red-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">
              {uptimeData.summary.unhealthyServices}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Service Statistics Table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Clock className="w-5 h-5 mr-2" />
            Service Uptime Details
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full table-auto">
              <thead>
                <tr className="border-b">
                  <th className="text-left py-3 px-4 font-semibold">Service</th>
                  <th className="text-left py-3 px-4 font-semibold">Status</th>
                  <th className="text-right py-3 px-4 font-semibold">
                    24h Uptime
                  </th>
                  <th className="text-right py-3 px-4 font-semibold">
                    7d Uptime
                  </th>
                  <th className="text-right py-3 px-4 font-semibold">
                    Restarts
                  </th>
                  <th className="text-right py-3 px-4 font-semibold">MTBF</th>
                  <th className="text-right py-3 px-4 font-semibold">
                    24h Downtime
                  </th>
                </tr>
              </thead>
              <tbody>
                {services.map((service) => (
                  <tr
                    key={service.serviceId}
                    className="border-b hover:bg-gray-50 dark:hover:bg-gray-800"
                  >
                    <td className="py-3 px-4">
                      <div>
                        <div className="font-medium">{service.serviceName}</div>
                        <div className="text-sm text-gray-500">
                          Port {service.port}
                        </div>
                      </div>
                    </td>
                    <td className="py-3 px-4">
                      <Badge
                        variant={getStatusBadgeVariant(
                          service.status,
                          service.healthStatus,
                        )}
                        className="text-xs"
                      >
                        {service.status === "running"
                          ? service.healthStatus
                          : service.status}
                      </Badge>
                    </td>
                    <td
                      className={`py-3 px-4 text-right font-medium ${getUptimeColor(service.stats.uptimePercentage24h)}`}
                    >
                      {service.stats.uptimePercentage24h.toFixed(2)}%
                    </td>
                    <td
                      className={`py-3 px-4 text-right font-medium ${getUptimeColor(service.stats.uptimePercentage7d)}`}
                    >
                      {service.stats.uptimePercentage7d.toFixed(2)}%
                    </td>
                    <td className="py-3 px-4 text-right">
                      {service.stats.totalRestarts}
                    </td>
                    <td className="py-3 px-4 text-right text-sm">
                      {formatMTBF(service.stats.mtbf)}
                    </td>
                    <td className="py-3 px-4 text-right text-sm">
                      {formatDuration(service.stats.totalDowntime24h)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {services.length === 0 && (
            <div className="text-center py-8 text-gray-500">
              No services available
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
