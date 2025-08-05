import {
  Play,
  Square,
  RotateCcw,
  Activity,
  FileText,
  Settings,
  Database,
  Folder,
  Trash2,
  Clock,
  Cpu,
  MemoryStick,
  Shield,
  Server,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Loader,
  Package,
  MoreVertical,
  Wrench,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Service } from "@/types";
import { useState, useRef, useEffect } from "react";

interface ServiceCardProps {
  service: Service;
  loadingStates: {
    starting?: boolean;
    stopping?: boolean;
    restarting?: boolean;
    checkingHealth?: boolean;
    installingLibraries?: boolean;
    validatingWrapper?: boolean;
    generatingWrapper?: boolean;
    repairingWrapper?: boolean;
  };
  onStart: () => void;
  onStop: () => void;
  onRestart: () => void;
  onCheckHealth: () => void;
  onViewLogs: () => void;
  onEdit: () => void;
  onDelete: () => void;
  onViewFiles: () => void;
  onEditEnv: () => void;
  onInstallLibraries: () => void;
  onManageWrappers: () => void;
}

export function ServiceCard({
  service,
  loadingStates,
  onStart,
  onStop,
  onRestart,
  onCheckHealth,
  onViewLogs,
  onEdit,
  onDelete,
  onViewFiles,
  onEditEnv,
  onInstallLibraries,
  onManageWrappers,
}: ServiceCardProps) {
  const [showDropdown, setShowDropdown] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setShowDropdown(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  const getStatusColor = () => {
    if (service.status === "running") {
      switch (service.healthStatus) {
        case "healthy":
          return "bg-green-500";
        case "unhealthy":
          return "bg-red-500";
        case "starting":
          return "bg-yellow-500";
        default:
          return "bg-blue-500";
      }
    }
    return "bg-gray-400";
  };

  const getStatusText = () => {
    if (service.status === "running") {
      switch (service.healthStatus) {
        case "healthy":
          return "Healthy";
        case "unhealthy":
          return "Unhealthy";
        case "starting":
          return "Starting";
        default:
          return "Running";
      }
    }
    return "Stopped";
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
  };

  const formatUptime = (uptime: string) => {
    if (!uptime || uptime === "0s") return "Not running";
    return uptime;
  };

  const isLoading = Object.values(loadingStates).some((state) => state);

  const getStatusIcon = () => {
    if (service.status === "running") {
      switch (service.healthStatus) {
        case "healthy":
          return <CheckCircle className="w-4 h-4 text-green-500" />;
        case "unhealthy":
          return <XCircle className="w-4 h-4 text-red-500" />;
        case "starting":
          return <Loader className="w-4 h-4 text-yellow-500 animate-spin" />;
        default:
          return <Activity className="w-4 h-4 text-blue-500" />;
      }
    }
    return <Square className="w-4 h-4 text-gray-400" />;
  };

  const getCardBorderColor = () => {
    if (service.status === "running") {
      switch (service.healthStatus) {
        case "healthy":
          return "border-l-green-500";
        case "unhealthy":
          return "border-l-red-500";
        case "starting":
          return "border-l-yellow-500";
        default:
          return "border-l-blue-500";
      }
    }
    return "border-l-gray-300";
  };

  return (
    <Card
      className={`h-full hover:shadow-lg transition-all duration-200 group border-l-2 ${getCardBorderColor()} bg-white dark:bg-gray-800 hover:bg-gray-50/30 dark:hover:bg-gray-700/30 relative overflow-hidden z-10 hover:z-20`}
    >
      <CardContent className="p-0 relative">
        {/* Header Section */}
        <div className="p-5 pb-4">
          <div className="flex items-start justify-between mb-3">
            <div className="flex items-center gap-3 flex-1 min-w-0">
              {/* Service Icon */}
              <div className="relative flex-shrink-0">
                <div
                  className={`p-3 rounded-xl ${service.status === "running" ? "bg-blue-500" : "bg-gray-400"} shadow-sm`}
                >
                  <Server className="w-6 h-6 text-white" />
                </div>
                <div
                  className={`absolute -bottom-1 -right-1 w-4 h-4 ${getStatusColor()} rounded-full border-2 border-white flex items-center justify-center`}
                >
                  {service.status === "running" &&
                    service.healthStatus === "starting" && (
                      <div className="w-1.5 h-1.5 bg-white rounded-full animate-pulse" />
                    )}
                </div>
              </div>

              {/* Service Info */}
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 truncate">
                    {service.name}
                  </h3>
                  {service.port && (
                    <Badge
                      variant="outline"
                      className="text-xs px-1.5 py-0.5 bg-gray-50 dark:bg-gray-700 text-gray-600 dark:text-gray-300 border-gray-200 dark:border-gray-600 flex-shrink-0"
                    >
                      :{service.port}
                    </Badge>
                  )}
                </div>

                <div className="flex items-center gap-2 mb-2">
                  {getStatusIcon()}
                  <span
                    className={`text-sm font-medium ${
                      service.status === "running"
                        ? service.healthStatus === "healthy"
                          ? "text-green-600"
                          : service.healthStatus === "unhealthy"
                            ? "text-red-600"
                            : service.healthStatus === "starting"
                              ? "text-yellow-600"
                              : "text-blue-600"
                        : "text-gray-500"
                    }`}
                  >
                    {getStatusText()}
                  </span>
                </div>

                {service.description && (
                  <p className="text-sm text-gray-600 dark:text-gray-400 line-clamp-2 leading-relaxed">
                    {service.description}
                  </p>
                )}
              </div>
            </div>

            {/* Actions Menu */}
            <div className="relative flex-shrink-0 ml-2" ref={dropdownRef}>
              <Button
                onClick={() => setShowDropdown(!showDropdown)}
                variant="ghost"
                size="sm"
                className="h-8 w-8 p-0 opacity-0 group-hover:opacity-100 transition-opacity duration-200 hover:bg-gray-100 dark:hover:bg-gray-700"
              >
                <MoreVertical className="w-4 h-4" />
                <span className="sr-only">More actions</span>
              </Button>

              {showDropdown && (
                <div className="absolute top-full mt-1 right-0 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-50 py-1 min-w-[160px]">
                  <button
                    onClick={() => {
                      onViewFiles();
                      setShowDropdown(false);
                    }}
                    className="w-full px-3 py-2 text-left text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-2"
                  >
                    <Folder className="w-3 h-3" />
                    View Files
                  </button>

                  <button
                    onClick={() => {
                      onEditEnv();
                      setShowDropdown(false);
                    }}
                    className="w-full px-3 py-2 text-left text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-2"
                  >
                    <Database className="w-3 h-3" />
                    Environment Variables
                  </button>

                  <button
                    onClick={() => {
                      onInstallLibraries();
                      setShowDropdown(false);
                    }}
                    className="w-full px-3 py-2 text-left text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-2"
                  >
                    <Package className="w-3 h-3" />
                    Install Libraries
                  </button>

                  <button
                    onClick={() => {
                      onManageWrappers();
                      setShowDropdown(false);
                    }}
                    className="w-full px-3 py-2 text-left text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-2"
                  >
                    <Wrench className="w-3 h-3" />
                    Manage Wrappers
                  </button>

                  <hr className="my-1 border-gray-100 dark:border-gray-700" />

                  <button
                    onClick={() => {
                      onEdit();
                      setShowDropdown(false);
                    }}
                    className="w-full px-3 py-2 text-left text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-2"
                  >
                    <Settings className="w-3 h-3" />
                    Edit Service
                  </button>

                  <button
                    onClick={() => {
                      onDelete();
                      setShowDropdown(false);
                    }}
                    className="w-full px-3 py-2 text-left text-xs text-red-600 hover:bg-red-50 flex items-center gap-2"
                  >
                    <Trash2 className="w-3 h-3" />
                    Delete Service
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Metrics Section - Progressive Disclosure */}
        {service.status === "running" ? (
          <div className="px-5 py-3 bg-gray-50/30 dark:bg-gray-800/30 border-y border-gray-100 dark:border-gray-700">
            <div className="grid grid-cols-2 gap-3">
              {/* CPU & Memory */}
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 bg-blue-100 rounded-md">
                    <Cpu className="w-3 h-3 text-blue-600" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="text-xs text-gray-500 dark:text-gray-400">
                      CPU
                    </div>
                    <div className="text-sm font-semibold text-gray-900 dark:text-gray-100">
                      {service.cpuPercent?.toFixed(1)}%
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <div className="p-1.5 bg-green-100 rounded-md">
                    <MemoryStick className="w-3 h-3 text-green-600" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="text-xs text-gray-500 dark:text-gray-400">
                      Memory
                    </div>
                    <div className="text-sm font-semibold text-gray-900 dark:text-gray-100">
                      {formatBytes(service.memoryUsage || 0)}
                    </div>
                  </div>
                </div>
              </div>

              {/* Uptime & Logs */}
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 bg-purple-100 rounded-md">
                    <Clock className="w-3 h-3 text-purple-600" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="text-xs text-gray-500 dark:text-gray-400">
                      Uptime
                    </div>
                    <div className="text-xs font-semibold text-gray-900 dark:text-gray-100 truncate">
                      {formatUptime(service.uptime)}
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <div className="p-1.5 bg-orange-100 rounded-md">
                    <FileText className="w-3 h-3 text-orange-600" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="text-xs text-gray-500 dark:text-gray-400">
                      Logs
                    </div>
                    <div className="text-sm font-semibold text-gray-900 dark:text-gray-100">
                      {service.logs?.length || 0}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        ) : (
          <div className="px-5 py-3 border-b border-gray-100 dark:border-gray-700">
            <div className="text-center text-sm text-gray-500 dark:text-gray-400">
              <Clock className="w-4 h-4 mx-auto mb-1 text-gray-400" />
              Service stopped • No metrics available
            </div>
          </div>
        )}

        {/* Actions Section */}
        <div className="p-5 pt-4 space-y-3">
          {/* Primary Actions */}
          <div className="flex gap-2">
            {service.status === "running" ? (
              <>
                <Button
                  onClick={onStop}
                  disabled={isLoading}
                  variant="outline"
                  className="flex-1 h-10 border-red-200 text-red-600 hover:bg-red-50 hover:border-red-300"
                >
                  {loadingStates.stopping ? (
                    <Loader className="w-4 h-4 animate-spin" />
                  ) : (
                    <Square className="w-4 h-4" />
                  )}
                  <span className="ml-2">
                    {loadingStates.stopping ? "Stopping..." : "Stop"}
                  </span>
                </Button>
                <Button
                  onClick={onRestart}
                  disabled={isLoading}
                  variant="outline"
                  className="flex-1 h-10 border-blue-200 text-blue-600 hover:bg-blue-50 hover:border-blue-300"
                >
                  {loadingStates.restarting ? (
                    <Loader className="w-4 h-4 animate-spin" />
                  ) : (
                    <RotateCcw className="w-4 h-4" />
                  )}
                  <span className="ml-2">
                    {loadingStates.restarting ? "Restarting..." : "Restart"}
                  </span>
                </Button>
              </>
            ) : (
              <Button
                onClick={onStart}
                disabled={isLoading || !service.isEnabled}
                className="flex-1 h-10 bg-green-500 hover:bg-green-600 font-medium"
              >
                {loadingStates.starting ? (
                  <Loader className="w-4 h-4 animate-spin text-white" />
                ) : (
                  <Play className="w-4 h-4" />
                )}
                <span className="ml-2">
                  {loadingStates.starting ? "Starting..." : "Start Service"}
                </span>
              </Button>
            )}
          </div>

          {/* Secondary Actions */}
          <div className="grid grid-cols-2 gap-2">
            <Button
              onClick={onCheckHealth}
              disabled={isLoading || service.status !== "running"}
              variant="outline"
              size="sm"
              className="h-9 text-xs hover:bg-green-50 hover:border-green-300 hover:text-green-700"
            >
              {loadingStates.checkingHealth ? (
                <Loader className="w-3 h-3 animate-spin" />
              ) : (
                <Shield className="w-3 h-3" />
              )}
              <span className="ml-1">Health</span>
            </Button>
            <Button
              onClick={onViewLogs}
              variant="outline"
              size="sm"
              className="h-9 text-xs hover:bg-orange-50 hover:border-orange-300 hover:text-orange-700"
            >
              <FileText className="w-3 h-3" />
              <span className="ml-1">Logs</span>
            </Button>
          </div>
        </div>

        {/* Disabled Status Banner */}
        {!service.isEnabled && (
          <div className="mx-5 mb-5 -mt-2 p-2 bg-yellow-50 border border-yellow-200 rounded-lg">
            <div className="flex items-center justify-center gap-2">
              <AlertTriangle className="w-3 h-3 text-yellow-600" />
              <p className="text-xs font-medium text-yellow-800">
                Service disabled
              </p>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default ServiceCard;
