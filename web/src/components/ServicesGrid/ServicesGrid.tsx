import { useState } from "react";
import {
  Plus,
  Play,
  Square,
  RefreshCw,
  Search,
  Settings2,
  Zap,
  X,
  Server,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ServiceCard } from "@/components/ServiceCard/ServiceCard";
import { Service } from "@/types";

interface ServicesGridProps {
  services: Service[];
  isLoading: boolean;
  isStartingAll: boolean;
  isStoppingAll: boolean;
  isFixingLombok: boolean;
  isSyncingEnvironment: boolean;
  serviceLoadingStates: Record<
    string,
    {
      starting?: boolean;
      stopping?: boolean;
      restarting?: boolean;
      checkingHealth?: boolean;
      installingLibraries?: boolean;
      validatingWrapper?: boolean;
      generatingWrapper?: boolean;
      repairingWrapper?: boolean;
    }
  >;
  onStartAll: () => void;
  onStopAll: () => void;
  onFixLombok: () => void;
  onSyncEnvironment: () => void;
  onCreateService: () => void;
  onStartService: (service: Service) => void;
  onStopService: (service: Service) => void;
  onRestartService: (service: Service) => void;
  onCheckHealth: (service: Service) => void;
  onViewLogs: (service: Service) => void;
  onEditService: (service: Service) => void;
  onDeleteService: (service: Service) => void;
  onViewFiles: (service: Service) => void;
  onEditEnv: (service: Service) => void;
  onInstallLibraries: (service: Service) => void;
  onManageWrappers: (service: Service) => void;
}

export function ServicesGrid({
  services,
  isLoading,
  isStartingAll,
  isStoppingAll,
  isFixingLombok,
  isSyncingEnvironment,
  serviceLoadingStates,
  onStartAll,
  onStopAll,
  onFixLombok,
  onSyncEnvironment,
  onCreateService,
  onStartService,
  onStopService,
  onRestartService,
  onCheckHealth,
  onViewLogs,
  onEditService,
  onDeleteService,
  onViewFiles,
  onEditEnv,
  onInstallLibraries,
  onManageWrappers,
}: ServicesGridProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<
    "all" | "running" | "stopped"
  >("all");

  // Filter services based on search and status
  const filteredServices = services.filter((service) => {
    const matchesSearch =
      service.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      service.description?.toLowerCase().includes(searchTerm.toLowerCase());

    const matchesStatus =
      statusFilter === "all" ||
      (statusFilter === "running" && service.status === "running") ||
      (statusFilter === "stopped" && service.status !== "running");

    return matchesSearch && matchesStatus;
  });

  // Service statistics
  const runningServices = services.filter((s) => s.status === "running").length;
  const healthyServices = services.filter(
    (s) => s.status === "running" && s.healthStatus === "healthy",
  ).length;
  const totalServices = services.length;

  if (isLoading) {
    return (
      <div className="h-96 flex items-center justify-center">
        <div className="text-center">
          <RefreshCw className="h-8 w-8 animate-spin mx-auto mb-4 text-blue-600" />
          <p className="text-gray-600 dark:text-gray-400">
            Loading services...
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header Section */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            Services
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Manage your microservices ecosystem
          </p>
        </div>
        <Button onClick={onCreateService} className="shrink-0">
          <Plus className="w-4 h-4 mr-2" />
          Add Service
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-100 rounded-lg">
                <Server className="w-5 h-5 text-blue-600" />
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Total Services
                </p>
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                  {totalServices}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-green-100 rounded-lg">
                <Play className="w-5 h-5 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Running
                </p>
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                  {runningServices}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-emerald-100 rounded-lg">
                <Zap className="w-5 h-5 text-emerald-600" />
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Healthy
                </p>
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                  {healthyServices}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-red-100 rounded-lg">
                <Square className="w-5 h-5 text-red-600" />
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Stopped
                </p>
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                  {totalServices - runningServices}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Enhanced Search & Controls Section */}
      <Card className="bg-white dark:bg-gray-800 shadow-sm border border-gray-200 dark:border-gray-700">
        <CardContent className="p-6">
          <div className="flex flex-col lg:flex-row gap-6 items-start lg:items-center justify-between">
            {/* Search Bar */}
            <div className="flex-1 max-w-2xl">
              <div className="relative">
                <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400 dark:text-gray-500" />
                <Input
                  placeholder="Search services by name or description..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-12 pr-12 h-12 text-base border-gray-300 focus:border-blue-500 focus:ring-blue-500 rounded-lg shadow-sm"
                />
                {searchTerm && (
                  <button
                    onClick={() => setSearchTerm("")}
                    className="absolute right-3 top-1/2 transform -translate-y-1/2 p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-full transition-colors"
                  >
                    <X className="h-4 w-4 text-gray-400 dark:text-gray-500" />
                  </button>
                )}
              </div>
            </div>

            {/* Filters */}
            <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-center">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                Filter:
              </span>
              <div className="flex gap-2">
                <Button
                  variant={statusFilter === "all" ? "default" : "outline"}
                  size="sm"
                  onClick={() => setStatusFilter("all")}
                  className="h-9"
                >
                  All
                </Button>
                <Button
                  variant={statusFilter === "running" ? "default" : "outline"}
                  size="sm"
                  onClick={() => setStatusFilter("running")}
                  className={`h-9 ${statusFilter === "running" ? "bg-green-500 hover:bg-green-600 border-green-500 text-white" : ""}`}
                >
                  Running
                </Button>
                <Button
                  variant={statusFilter === "stopped" ? "default" : "outline"}
                  size="sm"
                  onClick={() => setStatusFilter("stopped")}
                  className="h-9"
                >
                  Stopped
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Results Count */}
      <div className="flex justify-end">
        <div className="text-sm text-gray-600 dark:text-gray-400">
          Showing {filteredServices.length} of {totalServices} services
        </div>
      </div>

      {/* Bulk Actions */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Settings2 className="w-4 h-4" />
            Bulk Operations
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-3">
            <Button
              onClick={onStartAll}
              disabled={isStartingAll}
              variant="outline"
              size="sm"
            >
              {isStartingAll ? (
                <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Play className="w-4 h-4 mr-2" />
              )}
              Start All
            </Button>

            <Button
              onClick={onStopAll}
              disabled={isStoppingAll}
              variant="outline"
              size="sm"
            >
              {isStoppingAll ? (
                <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Square className="w-4 h-4 mr-2" />
              )}
              Stop All
            </Button>

            <Button
              onClick={onFixLombok}
              disabled={isFixingLombok}
              variant="outline"
              size="sm"
            >
              {isFixingLombok ? (
                <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Settings2 className="w-4 h-4 mr-2" />
              )}
              Fix Lombok
            </Button>

            <Button
              onClick={onSyncEnvironment}
              disabled={isSyncingEnvironment}
              variant="outline"
              size="sm"
            >
              {isSyncingEnvironment ? (
                <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <RefreshCw className="w-4 h-4 mr-2" />
              )}
              Sync Environment
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Services Display */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
            Services ({filteredServices.length})
          </h2>
          {searchTerm && (
            <Badge variant="outline">Filtered by: "{searchTerm}"</Badge>
          )}
        </div>

        {filteredServices.length === 0 ? (
          <Card>
            <CardContent className="p-12 text-center">
              <div className="mx-auto w-24 h-24 bg-gray-100 dark:bg-gray-700 rounded-full flex items-center justify-center mb-4">
                <Server className="w-8 h-8 text-gray-400 dark:text-gray-500" />
              </div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">
                {searchTerm ? "No services found" : "No services configured"}
              </h3>
              <p className="text-gray-600 dark:text-gray-400 mb-4">
                {searchTerm
                  ? `No services match "${searchTerm}". Try adjusting your search.`
                  : "Get started by adding your first service."}
              </p>
              {!searchTerm && (
                <Button onClick={onCreateService}>
                  <Plus className="w-4 h-4 mr-2" />
                  Add Service
                </Button>
              )}
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-6">
            {filteredServices.map((service) => (
              <ServiceCard
                key={service.id}
                service={service}
                loadingStates={serviceLoadingStates[service.id] || {}}
                onStart={() => onStartService(service)}
                onStop={() => onStopService(service)}
                onRestart={() => onRestartService(service)}
                onCheckHealth={() => onCheckHealth(service)}
                onViewLogs={() => onViewLogs(service)}
                onEdit={() => onEditService(service)}
                onDelete={() => onDeleteService(service)}
                onViewFiles={() => onViewFiles(service)}
                onEditEnv={() => onEditEnv(service)}
                onInstallLibraries={() => onInstallLibraries(service)}
                onManageWrappers={() => onManageWrappers(service)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default ServicesGrid;
