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
  LayoutGrid,
  List,
} from "lucide-react";
import { UptimeStatistics } from "@/types";
import { UptimeProgressBar } from "./UptimeProgressBar";
import { ServiceDetailModal } from "./ServiceDetailModal";
import { ServiceUptimeCard } from "./ServiceUptimeCard";
import { UptimeFiltersComponent, UptimeFilters } from "./UptimeFilters";

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
  const [viewMode, setViewMode] = useState<'grid' | 'table'>('grid');
  const [selectedService, setSelectedService] = useState<{ id: string; name: string } | null>(null);
  
  // Filter state
  const [filters, setFilters] = useState<UptimeFilters>({
    searchTerm: '',
    statusFilter: 'all',
    uptimeFilter: 'all',
    sortBy: 'name',
    sortOrder: 'asc',
    timeRange: '24h'
  });

  const defaultFilters: UptimeFilters = {
    searchTerm: '',
    statusFilter: 'all',
    uptimeFilter: 'all',
    sortBy: 'name',
    sortOrder: 'asc',
    timeRange: '24h'
  };

  const fetchUptimeStats = async () => {
    try {
      setIsLoading(true);
      setError(null);
      
      // Get auth token for API call
      const token = localStorage.getItem("authToken");
      if (!token) {
        throw new Error("No authentication token");
      }

      const response = await fetch("/api/uptime/statistics", {
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
      });
      
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

  // Filter and sort services
  const getFilteredAndSortedServices = () => {
    if (!uptimeData) return [];
    
    let services = Object.values(uptimeData.statistics);

    // Apply search filter
    if (filters.searchTerm) {
      services = services.filter(service =>
        service.serviceName.toLowerCase().includes(filters.searchTerm.toLowerCase())
      );
    }

    // Apply status filter
    if (filters.statusFilter !== 'all') {
      services = services.filter(service => {
        switch (filters.statusFilter) {
          case 'running':
            return service.status === 'running';
          case 'stopped':
            return service.status === 'stopped';
          case 'unhealthy':
            return service.status === 'running' && service.healthStatus === 'unhealthy';
          default:
            return true;
        }
      });
    }

    // Apply uptime filter
    if (filters.uptimeFilter !== 'all') {
      services = services.filter(service => {
        const uptime = service.stats.uptimePercentage7d;
        switch (filters.uptimeFilter) {
          case 'excellent':
            return uptime >= 99.9;
          case 'good':
            return uptime >= 99 && uptime < 99.9;
          case 'fair':
            return uptime >= 95 && uptime < 99;
          case 'poor':
            return uptime < 95;
          default:
            return true;
        }
      });
    }

    // Apply sorting
    services.sort((a, b) => {
      let aValue: any, bValue: any;
      
      switch (filters.sortBy) {
        case 'name':
          aValue = a.serviceName.toLowerCase();
          bValue = b.serviceName.toLowerCase();
          break;
        case 'uptime24h':
          aValue = a.stats.uptimePercentage24h;
          bValue = b.stats.uptimePercentage24h;
          break;
        case 'uptime7d':
          aValue = a.stats.uptimePercentage7d;
          bValue = b.stats.uptimePercentage7d;
          break;
        case 'restarts':
          aValue = a.stats.totalRestarts;
          bValue = b.stats.totalRestarts;
          break;
        case 'downtime':
          aValue = a.stats.totalDowntime24h;
          bValue = b.stats.totalDowntime24h;
          break;
        default:
          aValue = a.serviceName.toLowerCase();
          bValue = b.serviceName.toLowerCase();
      }

      if (aValue < bValue) return filters.sortOrder === 'asc' ? -1 : 1;
      if (aValue > bValue) return filters.sortOrder === 'asc' ? 1 : -1;
      return 0;
    });

    return services;
  };

  const filteredServices = getFilteredAndSortedServices();

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

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            Service Uptime Statistics
          </h2>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Monitor service availability and reliability metrics ({filteredServices.length} services)
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <div className="flex items-center space-x-1 bg-gray-100 dark:bg-gray-700 rounded-lg p-1">
            <Button
              onClick={() => setViewMode('grid')}
              variant={viewMode === 'grid' ? 'default' : 'ghost'}
              size="sm"
              className="h-8 px-2"
            >
              <LayoutGrid className="w-4 h-4" />
            </Button>
            <Button
              onClick={() => setViewMode('table')}
              variant={viewMode === 'table' ? 'default' : 'ghost'}
              size="sm"
              className="h-8 px-2"
            >
              <List className="w-4 h-4" />
            </Button>
          </div>
          <Button onClick={fetchUptimeStats} variant="outline" size="sm">
            <RefreshCw
              className={`w-4 h-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
            />
            Refresh
          </Button>
        </div>
      </div>

      {/* Filters */}
      <UptimeFiltersComponent
        filters={filters}
        onFiltersChange={setFilters}
        onReset={() => setFilters(defaultFilters)}
      />

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
              {filteredServices.length}
            </div>
            <p className="text-xs text-muted-foreground">
              {filteredServices.length !== uptimeData.summary.totalServices && 
                `${uptimeData.summary.totalServices} total`}
            </p>
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
              {filteredServices.filter(s => s.status === 'running').length}
            </div>
            <p className="text-xs text-muted-foreground">
              Active services
            </p>
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
              {filteredServices.filter(s => s.status === 'running' && s.healthStatus === 'unhealthy').length}
            </div>
            <p className="text-xs text-muted-foreground">
              Need attention
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Service Statistics - Grid View */}
      {viewMode === 'grid' && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredServices.map((service) => (
            <ServiceUptimeCard
              key={service.serviceId}
              service={service}
              onClick={() => setSelectedService({ id: service.serviceId, name: service.serviceName })}
            />
          ))}
        </div>
      )}

      {/* Service Statistics - Table View */}
      {viewMode === 'table' && (
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
                  {filteredServices.map((service) => (
                    <tr
                      key={service.serviceId}
                      className="border-b hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
                      onClick={() => setSelectedService({ id: service.serviceId, name: service.serviceName })}
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
                      <td className="py-3 px-4 text-right">
                        <UptimeProgressBar
                          percentage={service.stats.uptimePercentage24h}
                          size="sm"
                          className="justify-end"
                        />
                      </td>
                      <td className="py-3 px-4 text-right">
                        <UptimeProgressBar
                          percentage={service.stats.uptimePercentage7d}
                          size="sm"
                          className="justify-end"
                        />
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

            {filteredServices.length === 0 && (
              <div className="text-center py-8 text-gray-500">
                No services match the current filters
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Service Detail Modal */}
      {selectedService && (
        <ServiceDetailModal
          serviceId={selectedService.id}
          serviceName={selectedService.name}
          isOpen={!!selectedService}
          onClose={() => setSelectedService(null)}
        />
      )}
    </div>
  );
}
