import {
  Play,
  Square,
  Eye,
  Server,
  RotateCcw,
  Activity,
  Edit,
  Trash2,
  Clock,
  Globe,
  Terminal,
  FileText,
  Settings,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Service } from "@/types";
import { ButtonSpinner } from "@/components/ui/spinner";
import ResourceMetrics from "@/components/ResourceMetrics/ResourceMetrics";
import PerformanceMetrics from "@/components/PerformanceMetrics/PerformanceMetrics";

interface ServiceCardProps {
  service: Service;
  isStarting?: boolean;
  isStopping?: boolean;
  isRestarting?: boolean;
  isCheckingHealth?: boolean;
  onStart: (serviceName: string) => void;
  onStop: (serviceName: string) => void;
  onRestart: (serviceName: string) => void;
  onCheckHealth: (serviceName: string) => void;
  onViewLogs: (service: Service) => void;
  onEdit: (service: Service) => void;
  onDelete: (serviceName: string) => void;
  onViewFiles: (service: Service) => void;
  onEditEnv: (service: Service) => void;
  getStatusBadge: (status: string, healthStatus: string) => JSX.Element;
}

export function ServiceCard({
  service,
  isStarting = false,
  isStopping = false,
  isRestarting = false,
  isCheckingHealth = false,
  onStart,
  onStop,
  onRestart,
  onCheckHealth,
  onViewLogs,
  onEdit,
  onDelete,
  onViewFiles,
  onEditEnv,
  getStatusBadge,
}: ServiceCardProps) {
  return (
    <Card className="group hover:shadow-lg transition-all duration-200 border-l-4 border-l-blue-500 hover:border-l-blue-600">
      <CardHeader className="pb-4">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-4">
            <div className="relative">
              <div
                className={`p-3 rounded-xl transition-colors ${
                  service.status === "running"
                    ? "bg-green-100 text-green-700 dark:bg-green-900/20 dark:text-green-400"
                    : "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400"
                }`}
              >
                <Server className="h-6 w-6" />
              </div>
              <div
                className={`absolute -top-1 -right-1 w-4 h-4 rounded-full border-2 border-background ${
                  service.healthStatus === "healthy"
                    ? "bg-green-500"
                    : service.healthStatus === "unhealthy"
                      ? "bg-red-500"
                      : service.healthStatus === "starting"
                        ? "bg-yellow-500"
                        : "bg-gray-400"
                }`}
              />
            </div>
            <div>
              <div className="flex items-center gap-3 mb-1">
                <CardTitle className="text-xl font-semibold">
                  {service.name}
                </CardTitle>
                <span className="text-xs font-mono bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 px-2 py-1 rounded-full">
                  #{service.order}
                </span>
              </div>
              <div className="flex items-center gap-4 text-sm text-muted-foreground">
                <div className="flex items-center gap-1">
                  <Terminal className="h-3 w-3" />
                  <span>{service.dir}</span>
                </div>
                <div className="flex items-center gap-1">
                  <Globe className="h-3 w-3" />
                  <span>:{service.port}</span>
                </div>
                {service.status === "running" && service.uptime && (
                  <div className="flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    <span>{service.uptime}</span>
                  </div>
                )}
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {getStatusBadge(service.status, service.healthStatus)}
            {service.status === "running" && service.pid > 0 && (
              <Badge variant="outline" className="text-xs">
                PID: {service.pid}
              </Badge>
            )}
          </div>
        </div>

        {service.description && (
          <div className="mt-3 p-3 bg-muted/50 rounded-lg">
            <p className="text-sm text-muted-foreground">
              {service.description}
            </p>
          </div>
        )}

        {/* Resource Metrics */}
        {service.status === "running" && service.pid > 0 && (
          <div className="mt-3 space-y-3">
            <ResourceMetrics
              cpuPercent={service.cpuPercent || 0}
              memoryUsage={service.memoryUsage || 0}
              memoryPercent={service.memoryPercent || 0}
              diskUsage={service.diskUsage || 0}
              networkRx={service.networkRx || 0}
              networkTx={service.networkTx || 0}
            />
            {service.metrics && (
              <PerformanceMetrics metrics={service.metrics} />
            )}
          </div>
        )}
      </CardHeader>

      <CardContent className="pt-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            {service.status === "running" ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onStop(service.name)}
                disabled={isStopping}
                className="hover:bg-red-50 hover:text-red-600 hover:border-red-200"
              >
                <ButtonSpinner isLoading={isStopping} loadingText="Stopping...">
                  <Square className="h-3 w-3 mr-1.5" />
                  Stop
                </ButtonSpinner>
              </Button>
            ) : (
              <Button
                size="sm"
                onClick={() => onStart(service.name)}
                disabled={isStarting}
                className="bg-green-600 hover:bg-green-700 text-white"
              >
                <ButtonSpinner isLoading={isStarting} loadingText="Starting...">
                  <Play className="h-3 w-3 mr-1.5" />
                  Start
                </ButtonSpinner>
              </Button>
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={() => onRestart(service.name)}
              disabled={isRestarting}
              className="hover:bg-blue-50 hover:text-blue-600 hover:border-blue-200"
            >
              <ButtonSpinner
                isLoading={isRestarting}
                loadingText="Restarting..."
              >
                <RotateCcw className="h-3 w-3 mr-1.5" />
                Restart
              </ButtonSpinner>
            </Button>
            {service.status === "running" && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onCheckHealth(service.name)}
                disabled={isCheckingHealth}
                className="hover:bg-purple-50 hover:text-purple-600 hover:border-purple-200"
              >
                <ButtonSpinner
                  isLoading={isCheckingHealth}
                  loadingText="Checking..."
                >
                  <Activity className="h-3 w-3 mr-1.5" />
                  Health
                </ButtonSpinner>
              </Button>
            )}
          </div>

          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onViewLogs(service)}
              className="hover:bg-gray-100 dark:hover:bg-gray-800"
              title="View Logs"
            >
              <Eye className="h-3 w-3" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onViewFiles(service)}
              className="hover:bg-blue-50 hover:text-blue-600 hover:border-blue-200 border"
              title="Edit Configuration Files"
            >
              <FileText className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onEditEnv(service)}
              className="hover:bg-gray-100 dark:hover:bg-gray-800"
              title="Environment Variables"
            >
              <Settings className="h-3 w-3" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onEdit(service)}
              className="hover:bg-gray-100 dark:hover:bg-gray-800"
              title="Edit Service"
            >
              <Edit className="h-3 w-3" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onDelete(service.name)}
              className="hover:bg-red-100 hover:text-red-600 dark:hover:bg-red-900/20"
              title="Delete Service"
            >
              <Trash2 className="h-3 w-3" />
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
