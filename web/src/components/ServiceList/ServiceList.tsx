import { Play, Square, Plus, Zap, Activity, Wrench, RefreshCw, MoreVertical } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Service } from "@/types";
import { ServiceCard } from "./ServiceCard";
import { ButtonSpinner } from "@/components/ui/spinner";
import { ServiceListSkeleton } from "@/components/ui/skeleton";
import { ErrorBoundarySection } from "@/components/ui/error-boundary";
import { useState, useRef, useEffect } from "react";

interface ServiceListProps {
  services: Service[];
  isStartingAll: boolean;
  isStoppingAll: boolean;
  isFixingLombok?: boolean;
  isSyncingEnvironment?: boolean;
  isLoading?: boolean;
  serviceLoadingStates: Record<string, {
    starting?: boolean;
    stopping?: boolean;
    restarting?: boolean;
    checkingHealth?: boolean;
  }>;
  onStartAll: () => void;
  onStopAll: () => void;
  onFixLombok?: () => void;
  onSyncEnvironment?: () => void;
  onCreateService: () => void;
  onStartService: (serviceName: string) => void;
  onStopService: (serviceName: string) => void;
  onRestartService: (serviceName: string) => void;
  onCheckHealth: (serviceName: string) => void;
  onViewLogs: (service: Service) => void;
  onEditService: (service: Service) => void;
  onDeleteService: (serviceName: string) => void;
  onViewFiles: (service: Service) => void;
  onEditEnv: (service: Service) => void;
  getStatusBadge: (status: string, healthStatus: string) => JSX.Element;
}

export function ServiceList({
  services,
  isStartingAll,
  isStoppingAll,
  isFixingLombok = false,
  isSyncingEnvironment = false,
  isLoading = false,
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
  getStatusBadge,
}: ServiceListProps) {
  const [showOptionsDropdown, setShowOptionsDropdown] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowOptionsDropdown(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  if (isLoading) {
    return <ServiceListSkeleton />;
  }
  return (
    <ErrorBoundarySection title="Services Error" description="Failed to load the services section.">
      <div className="space-y-4">
        <div className="mb-6">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                Services
              </h2>
              <p className="text-sm text-muted-foreground mt-1">
                {services.filter((s) => s.status === "running").length} of{" "}
                {services.length} services running
              </p>
            </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={onCreateService}
              className="hover:bg-green-50 hover:text-green-600 hover:border-green-200"
            >
              <Plus className="h-4 w-4 mr-2" />
              Add Service
            </Button>
            <Button
              onClick={onStartAll}
              disabled={
                isStartingAll || services.every((s) => s.status === "running")
              }
              className="bg-green-600 hover:bg-green-700"
            >
              <ButtonSpinner isLoading={isStartingAll} loadingText="Starting...">
                <Play className="h-4 w-4 mr-2" />
                Start All
              </ButtonSpinner>
            </Button>
            <Button
              variant="destructive"
              onClick={onStopAll}
              disabled={
                isStoppingAll || services.every((s) => s.status === "stopped")
              }
            >
              <ButtonSpinner isLoading={isStoppingAll} loadingText="Stopping...">
                <Square className="h-4 w-4 mr-2" />
                Stop All
              </ButtonSpinner>
            </Button>
            
            {/* More Options Dropdown */}
            <div className="relative" ref={dropdownRef}>
              <Button
                variant="outline"
                onClick={() => setShowOptionsDropdown(!showOptionsDropdown)}
                className="hover:bg-gray-50 hover:text-gray-700 hover:border-gray-300"
              >
                <MoreVertical className="h-4 w-4" />
              </Button>
              
              {showOptionsDropdown && (
                <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 rounded-md shadow-lg border border-gray-200 dark:border-gray-700 z-50">
                  <div className="py-1">
                    {onFixLombok && (
                      <button
                        onClick={() => {
                          onFixLombok();
                          setShowOptionsDropdown(false);
                        }}
                        disabled={isFixingLombok}
                        className="w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-blue-50 dark:hover:bg-blue-900/20 hover:text-blue-600 disabled:opacity-50 disabled:cursor-not-allowed flex items-center"
                      >
                        <Wrench className="h-4 w-4 mr-3" />
                        {isFixingLombok ? "Fixing Lombok..." : "Fix Lombok"}
                      </button>
                    )}
                    {onSyncEnvironment && (
                      <button
                        onClick={() => {
                          onSyncEnvironment();
                          setShowOptionsDropdown(false);
                        }}
                        disabled={isSyncingEnvironment}
                        className="w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-purple-50 dark:hover:bg-purple-900/20 hover:text-purple-600 disabled:opacity-50 disabled:cursor-not-allowed flex items-center"
                      >
                        <RefreshCw className="h-4 w-4 mr-3" />
                        {isSyncingEnvironment ? "Syncing Environment..." : "Sync Environment"}
                      </button>
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Services Status Overview */}
        <div className="grid grid-cols-3 gap-4 mb-4">
          <div className="bg-green-50 dark:bg-green-900/20 p-3 rounded-lg border border-green-200 dark:border-green-800">
            <div className="flex items-center gap-2">
              <Zap className="h-4 w-4 text-green-600" />
              <span className="text-sm font-medium text-green-800 dark:text-green-300">
                Running
              </span>
            </div>
            <p className="text-lg font-bold text-green-900 dark:text-green-100">
              {services.filter((s) => s.status === "running").length}
            </p>
          </div>
          <div className="bg-gray-50 dark:bg-gray-800 p-3 rounded-lg border border-gray-200 dark:border-gray-700">
            <div className="flex items-center gap-2">
              <Square className="h-4 w-4 text-gray-600" />
              <span className="text-sm font-medium text-gray-800 dark:text-gray-300">
                Stopped
              </span>
            </div>
            <p className="text-lg font-bold text-gray-900 dark:text-gray-100">
              {services.filter((s) => s.status === "stopped").length}
            </p>
          </div>
          <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded-lg border border-blue-200 dark:border-blue-800">
            <div className="flex items-center gap-2">
              <Activity className="h-4 w-4 text-blue-600" />
              <span className="text-sm font-medium text-blue-800 dark:text-blue-300">
                Healthy
              </span>
            </div>
            <p className="text-lg font-bold text-blue-900 dark:text-blue-100">
              {services.filter((s) => s.healthStatus === "healthy").length}
            </p>
          </div>
        </div>
      </div>

      {services.map((service) => (
        <ServiceCard
          key={service.name}
          service={service}
          isStarting={serviceLoadingStates[service.name]?.starting || false}
          isStopping={serviceLoadingStates[service.name]?.stopping || false}
          isRestarting={serviceLoadingStates[service.name]?.restarting || false}
          isCheckingHealth={serviceLoadingStates[service.name]?.checkingHealth || false}
          onStart={onStartService}
          onStop={onStopService}
          onRestart={onRestartService}
          onCheckHealth={onCheckHealth}
          onViewLogs={onViewLogs}
          onEdit={onEditService}
          onDelete={onDeleteService}
          onViewFiles={onViewFiles}
          onEditEnv={onEditEnv}
          getStatusBadge={getStatusBadge}
        />
      ))}
      </div>
    </ErrorBoundarySection>
  );
}