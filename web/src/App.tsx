import { useState, useEffect } from "react";
import { Service, Configuration } from "@/types";
import { Sidebar } from "@/components/Sidebar/Sidebar";
import { ServicesGrid } from "@/components/ServicesGrid/ServicesGrid";
import { LogsDrawer } from "@/components/LogsDrawer/LogsDrawer";
import { ServiceConfigModal } from "@/components/ServiceConfigModal/ServiceConfigModal";
import { ServiceFilesModal } from "@/components/ServiceConfigModal/ServiceFilesModal";
import { ServiceEnvModal } from "@/components/ServiceEnvModal/ServiceEnvModal";
import { GlobalEnvModal } from "@/components/GlobalEnvModal/GlobalEnvModal";
import { GlobalConfigModal } from "@/components/GlobalConfigModal/GlobalConfigModal";
import { ConfigurationManager } from "@/components/ConfigurationManager/ConfigurationManager";
import { SystemMetricsModal } from "@/components/SystemMetricsModal/SystemMetricsModal";
import { LogAggregationModal } from "@/components/LogAggregationModal/LogAggregationModal";
import { ServiceTopologyModal } from "@/components/ServiceTopologyModal/ServiceTopologyModal";
import { DependencyConfigModal } from "@/components/DependencyConfigModal/DependencyConfigModal";
import { AutoDiscoveryModal } from "@/components/AutoDiscoveryModal/AutoDiscoveryModal";
import {
  ToastProvider,
  ToastContainer,
  useToast,
  toast,
} from "@/components/ui/toast";
import {
  ConfirmDialogProvider,
  useConfirm,
  confirmDialogs,
} from "@/components/ui/confirm-dialog";
import { ErrorBoundary } from "@/components/ui/error-boundary";

function AppContent() {
  const [services, setServices] = useState<Service[]>([]);
  const [selectedService, setSelectedService] = useState<Service | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [activeSection, setActiveSection] = useState("services");
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const [copied, setCopied] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [isStartingAll, setIsStartingAll] = useState(false);
  const [isStoppingAll, setIsStoppingAll] = useState(false);
  const [isFixingLombok, setIsFixingLombok] = useState(false);
  const [isSyncingEnvironment, setIsSyncingEnvironment] = useState(false);
  const [isCopyingLogs, setIsCopyingLogs] = useState(false);
  const [isClearingLogs, setIsClearingLogs] = useState(false);
  const [isSavingService, setIsSavingService] = useState(false);

  // Individual service operation states
  const [serviceLoadingStates, setServiceLoadingStates] = useState<
    Record<
      string,
      {
        starting?: boolean;
        stopping?: boolean;
        restarting?: boolean;
        checkingHealth?: boolean;
        installingLibraries?: boolean;
      }
    >
  >({});

  // Get toast and confirm hooks
  const { addToast } = useToast();
  const { showConfirm } = useConfirm();

  // Modal state
  const [showServiceConfig, setShowServiceConfig] = useState(false);
  const [showServiceFiles, setShowServiceFiles] = useState(false);
  const [showServiceEnv, setShowServiceEnv] = useState(false);
  const [editingService, setEditingService] = useState<Service | null>(null);
  const [viewingFilesService, setViewingFilesService] =
    useState<Service | null>(null);
  const [envEditingService, setEnvEditingService] = useState<Service | null>(
    null,
  );

  // Configuration state
  const [configurations, setConfigurations] = useState<Configuration[]>([]);

  useEffect(() => {
    fetchServices();
    fetchConfigurations();

    // WebSocket connection for real-time updates
    const ws = new WebSocket(`ws://${window.location.host}/ws`);

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);

      if (message.type === "service_update") {
        const updatedService: Service = message.payload;
        setServices((prev) =>
          prev.map((service) =>
            service.name === updatedService.name ? updatedService : service,
          ),
        );

        // Clear loading states when service status changes
        setServiceLoadingStates((prev) => ({
          ...prev,
          [updatedService.name]: {
            starting:
              updatedService.status === "running"
                ? false
                : prev[updatedService.name]?.starting || false,
            stopping:
              updatedService.status === "stopped"
                ? false
                : prev[updatedService.name]?.stopping || false,
            restarting: false,
            checkingHealth: false,
          },
        }));

        if (selectedService && selectedService.name === updatedService.name) {
          setSelectedService(updatedService);
        }
      } else if (message.type === "log_entry") {
        const { serviceName, logEntry } = message.payload;
        setServices((prev) =>
          prev.map((service) =>
            service.name === serviceName
              ? { ...service, logs: [...service.logs, logEntry] }
              : service,
          ),
        );
        if (selectedService && selectedService.name === serviceName) {
          setSelectedService((prev) =>
            prev ? { ...prev, logs: [...prev.logs, logEntry] } : null,
          );
        }
      }
    };

    return () => ws.close();
  }, [selectedService]);

  const fetchServices = async () => {
    try {
      const response = await fetch("/api/services");
      if (!response.ok) {
        throw new Error(
          `Failed to fetch services: ${response.status} ${response.statusText}`,
        );
      }
      const data = await response.json();
      const sortedServices = data.sort(
        (a: Service, b: Service) => a.order - b.order,
      );
      setServices(sortedServices);
    } catch (error) {
      console.error("Failed to fetch services:", error);
      addToast(
        toast.error(
          "Failed to load services",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    } finally {
      setIsLoading(false);
    }
  };

  const fetchConfigurations = async () => {
    try {
      const response = await fetch("/api/configurations");
      if (!response.ok) {
        throw new Error(
          `Failed to fetch configurations: ${response.status} ${response.statusText}`,
        );
      }
      const data = await response.json();
      setConfigurations(data);
    } catch (error) {
      console.error("Failed to fetch configurations:", error);
      addToast(
        toast.error(
          "Failed to load configurations",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    }
  };

  // Service operations (same as before)
  const startService = async (serviceName: string) => {
    try {
      setServiceLoadingStates((prev) => ({
        ...prev,
        [serviceName]: { ...prev[serviceName], starting: true },
      }));

      const response = await fetch(`/api/services/${serviceName}/start`, {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to start service: ${response.status} ${response.statusText}`,
        );
      }
      addToast(
        toast.success("Service starting", `${serviceName} is now starting up`),
      );
    } catch (error) {
      console.error("Failed to start service:", error);
      addToast(
        toast.error(
          "Failed to start service",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
      setServiceLoadingStates((prev) => ({
        ...prev,
        [serviceName]: { ...prev[serviceName], starting: false },
      }));
    }
  };

  const stopService = async (serviceName: string) => {
    try {
      setServiceLoadingStates((prev) => ({
        ...prev,
        [serviceName]: { ...prev[serviceName], stopping: true },
      }));

      const response = await fetch(`/api/services/${serviceName}/stop`, {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to stop service: ${response.status} ${response.statusText}`,
        );
      }
      addToast(
        toast.success("Service stopping", `${serviceName} is shutting down`),
      );
    } catch (error) {
      console.error("Failed to stop service:", error);
      addToast(
        toast.error(
          "Failed to stop service",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
      setServiceLoadingStates((prev) => ({
        ...prev,
        [serviceName]: { ...prev[serviceName], stopping: false },
      }));
    }
  };

  const restartService = async (serviceName: string) => {
    try {
      setServiceLoadingStates((prev) => ({
        ...prev,
        [serviceName]: { ...prev[serviceName], restarting: true },
      }));

      const response = await fetch(`/api/services/${serviceName}/restart`, {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to restart service: ${response.status} ${response.statusText}`,
        );
      }
      addToast(
        toast.success(
          "Service restarting",
          `${serviceName} is being restarted`,
        ),
      );
    } catch (error) {
      console.error("Failed to restart service:", error);
      addToast(
        toast.error(
          "Failed to restart service",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
      setServiceLoadingStates((prev) => ({
        ...prev,
        [serviceName]: { ...prev[serviceName], restarting: false },
      }));
    }
  };

  const checkServiceHealth = async (serviceName: string) => {
    try {
      setServiceLoadingStates((prev) => ({
        ...prev,
        [serviceName]: { ...prev[serviceName], checkingHealth: true },
      }));

      const response = await fetch(`/api/services/${serviceName}/health`, {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to check service health: ${response.status} ${response.statusText}`,
        );
      }
      addToast(
        toast.info(
          "Health check initiated",
          `Checking ${serviceName} health status`,
        ),
      );

      setTimeout(() => {
        setServiceLoadingStates((prev) => ({
          ...prev,
          [serviceName]: { ...prev[serviceName], checkingHealth: false },
        }));
      }, 1000);
    } catch (error) {
      console.error("Failed to check service health:", error);
      addToast(
        toast.error(
          "Failed to check service health",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
      setServiceLoadingStates((prev) => ({
        ...prev,
        [serviceName]: { ...prev[serviceName], checkingHealth: false },
      }));
    }
  };

  const installLibraries = async (service: Service) => {
    try {
      setServiceLoadingStates((prev) => ({
        ...prev,
        [service.name]: { ...prev[service.name], installingLibraries: true },
      }));

      const response = await fetch(
        `/api/services/${service.name}/install-libraries`,
        {
          method: "POST",
        },
      );

      if (!response.ok) {
        throw new Error(
          `Failed to install libraries: ${response.status} ${response.statusText}`,
        );
      }

      await response.json();
      addToast(
        toast.success(
          "Library installation started",
          `Installing Maven libraries for ${service.name}. Check logs for progress.`,
        ),
      );

      setTimeout(() => {
        setServiceLoadingStates((prev) => ({
          ...prev,
          [service.name]: { ...prev[service.name], installingLibraries: false },
        }));
      }, 2000);
    } catch (error) {
      console.error("Failed to install libraries:", error);
      addToast(
        toast.error(
          "Failed to install libraries",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
      setServiceLoadingStates((prev) => ({
        ...prev,
        [service.name]: { ...prev[service.name], installingLibraries: false },
      }));
    }
  };

  const startAllServices = async () => {
    try {
      setIsStartingAll(true);
      const response = await fetch("/api/services/start-all", {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to start all services: ${response.status} ${response.statusText}`,
        );
      }
      addToast(
        toast.success(
          "Starting all services",
          "All enabled services are being started",
        ),
      );
    } catch (error) {
      console.error("Failed to start all services:", error);
      addToast(
        toast.error(
          "Failed to start all services",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    } finally {
      setIsStartingAll(false);
    }
  };

  const stopAllServices = async () => {
    const confirmed = await showConfirm(confirmDialogs.stopAllServices());
    if (!confirmed) return;

    try {
      setIsStoppingAll(true);
      const response = await fetch("/api/services/stop-all", {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to stop all services: ${response.status} ${response.statusText}`,
        );
      }
      addToast(
        toast.success(
          "Stopping all services",
          "All services are being stopped",
        ),
      );
    } catch (error) {
      console.error("Failed to stop all services:", error);
      addToast(
        toast.error(
          "Failed to stop all services",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    } finally {
      setIsStoppingAll(false);
    }
  };

  const fixLombok = async () => {
    try {
      setIsFixingLombok(true);
      addToast(
        toast.info(
          "Fixing Lombok",
          "Checking and fixing Lombok compatibility for all services...",
        ),
      );

      const response = await fetch("/api/services/fix-lombok", {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error(
          `Failed to fix Lombok: ${response.status} ${response.statusText}`,
        );
      }

      const result = await response.json();
      const successCount = Object.values(result.results).filter(
        (r: any) => r === "Success",
      ).length;
      const errorCount = Object.values(result.results).filter(
        (r: any) => r !== "Success",
      ).length;

      addToast(
        errorCount > 0
          ? toast.warning(
              "Lombok Fix Complete",
              `Successfully processed ${successCount} services. ${errorCount} services had errors.`,
            )
          : toast.success(
              "Lombok Fix Complete",
              `Successfully processed ${successCount} services.`,
            ),
      );
    } catch (error) {
      console.error("Error fixing Lombok:", error);
      addToast(
        toast.error(
          "Error",
          error instanceof Error ? error.message : "Failed to fix Lombok",
        ),
      );
    } finally {
      setIsFixingLombok(false);
    }
  };

  const syncEnvironment = async () => {
    try {
      setIsSyncingEnvironment(true);
      addToast(
        toast.info(
          "Syncing Environment",
          "Synchronizing environment variables from configuration files...",
        ),
      );

      const response = await fetch("/api/environment/sync", {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error(
          `Failed to sync environment: ${response.status} ${response.statusText}`,
        );
      }

      const result = await response.json();
      addToast(
        result.success
          ? toast.success(
              "Environment Sync Complete",
              `Successfully loaded ${result.variablesSet} environment variables.${result.errors?.length > 0 ? ` ${result.errors.length} warnings occurred.` : ""}`,
            )
          : toast.warning(
              "Environment Sync Complete",
              `Partially loaded ${result.variablesSet} environment variables. ${result.errors?.length || 0} warnings occurred.`,
            ),
      );
    } catch (error) {
      console.error("Error syncing environment:", error);
      addToast(
        toast.error(
          "Error",
          error instanceof Error ? error.message : "Failed to sync environment",
        ),
      );
    } finally {
      setIsSyncingEnvironment(false);
    }
  };

  // Service management handlers
  const openCreateService = () => {
    setEditingService({
      name: "",
      dir: "",
      extraEnv: "",
      javaOpts: "",
      status: "stopped",
      healthStatus: "unknown",
      healthUrl: "",
      port: 8080,
      pid: 0,
      order: services.length + 1,
      lastStarted: "",
      description: "",
      isEnabled: true,
      envVars: {},
      logs: [],
      uptime: "",
    });
    setShowServiceConfig(true);
  };

  const openEditService = (service: Service) => {
    setEditingService(service);
    setShowServiceConfig(true);
  };

  const openViewFiles = (service: Service) => {
    setViewingFilesService(service);
    setShowServiceFiles(true);
  };

  const openEditEnv = (service: Service) => {
    setEnvEditingService(service);
    setShowServiceEnv(true);
  };

  const deleteService = async (serviceName: string) => {
    const confirmed = await showConfirm(
      confirmDialogs.deleteService(serviceName),
    );
    if (!confirmed) return;

    try {
      const response = await fetch(`/api/services/${serviceName}`, {
        method: "DELETE",
      });
      if (!response.ok) {
        throw new Error(
          `Failed to delete service: ${response.status} ${response.statusText}`,
        );
      }
      addToast(
        toast.success(
          "Service deleted",
          `${serviceName} has been deleted successfully`,
        ),
      );
      fetchServices();
    } catch (error) {
      console.error("Failed to delete service:", error);
      addToast(
        toast.error(
          "Failed to delete service",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    }
  };

  // Logs operations
  const copyLogsToClipboard = async (
    selectedLevels: string[] = ["INFO", "WARN", "ERROR"],
  ) => {
    if (!selectedService || selectedService.logs.length === 0) return;

    try {
      setIsCopyingLogs(true);
      const filteredLogs = selectedService.logs.filter((log) =>
        selectedLevels.includes(log.level),
      );
      const logsText = filteredLogs
        .map(
          (log) =>
            `${new Date(log.timestamp).toLocaleString()} [${log.level}] ${log.message}`,
        )
        .join("\n");

      await navigator.clipboard.writeText(logsText);
      setCopied(true);
      addToast(
        toast.success(
          "Logs copied",
          `Copied ${filteredLogs.length} log entries to clipboard`,
        ),
      );
      setTimeout(() => setCopied(false), 2000);
    } catch (error) {
      console.error("Failed to copy logs:", error);
      addToast(
        toast.error(
          "Failed to copy logs",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    } finally {
      setIsCopyingLogs(false);
    }
  };

  const clearLogs = async () => {
    if (!selectedService) return;

    const confirmed = await showConfirm(
      confirmDialogs.clearLogs(selectedService.name),
    );
    if (!confirmed) return;

    try {
      setIsClearingLogs(true);
      const response = await fetch(
        `/api/services/${selectedService.name}/logs`,
        {
          method: "DELETE",
        },
      );
      if (!response.ok) {
        throw new Error(
          `Failed to clear logs: ${response.status} ${response.statusText}`,
        );
      }
      setServices((prev) =>
        prev.map((service) =>
          service.name === selectedService.name
            ? { ...service, logs: [] }
            : service,
        ),
      );
      setSelectedService((prev) => (prev ? { ...prev, logs: [] } : null));
      addToast(
        toast.success(
          "Logs cleared",
          `Logs for ${selectedService.name} have been cleared`,
        ),
      );
    } catch (error) {
      console.error("Failed to clear logs:", error);
      addToast(
        toast.error(
          "Failed to clear logs",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    } finally {
      setIsClearingLogs(false);
    }
  };

  const clearSearch = () => {
    setSearchTerm("");
  };

  // Render different sections based on activeSection
  const renderMainContent = () => {
    switch (activeSection) {
      case "services":
        return (
          <ServicesGrid
            services={services}
            isLoading={isLoading}
            isStartingAll={isStartingAll}
            isStoppingAll={isStoppingAll}
            isFixingLombok={isFixingLombok}
            isSyncingEnvironment={isSyncingEnvironment}
            serviceLoadingStates={serviceLoadingStates}
            onStartAll={startAllServices}
            onStopAll={stopAllServices}
            onFixLombok={fixLombok}
            onSyncEnvironment={syncEnvironment}
            onCreateService={openCreateService}
            onStartService={startService}
            onStopService={stopService}
            onRestartService={restartService}
            onCheckHealth={checkServiceHealth}
            onViewLogs={setSelectedService}
            onEditService={openEditService}
            onDeleteService={deleteService}
            onViewFiles={openViewFiles}
            onEditEnv={openEditEnv}
            onInstallLibraries={installLibraries}
          />
        );
      case "metrics":
        return (
          <SystemMetricsModal
            isOpen={true}
            onClose={() => setActiveSection("services")}
          />
        );
      case "logs":
        return (
          <LogAggregationModal
            isOpen={true}
            onClose={() => setActiveSection("services")}
            services={services}
          />
        );
      case "topology":
        return (
          <ServiceTopologyModal
            isOpen={true}
            onClose={() => setActiveSection("services")}
          />
        );
      case "dependencies":
        return (
          <DependencyConfigModal
            isOpen={true}
            onClose={() => setActiveSection("services")}
            services={services}
          />
        );
      case "auto-discovery":
        return (
          <AutoDiscoveryModal
            isOpen={true}
            onClose={() => setActiveSection("services")}
            onServiceImported={() => {
              fetchServices();
              setActiveSection("services");
            }}
          />
        );
      case "configurations":
        return (
          <ConfigurationManager
            isOpen={true}
            onClose={() => setActiveSection("services")}
            configurations={configurations}
            services={services}
            onConfigurationSaved={() => {
              fetchConfigurations();
              fetchServices();
            }}
          />
        );
      case "environment":
        return (
          <GlobalEnvModal
            isOpen={true}
            onClose={() => setActiveSection("services")}
          />
        );
      case "settings":
        return (
          <GlobalConfigModal
            isOpen={true}
            onClose={() => setActiveSection("services")}
            onConfigUpdated={fetchServices}
          />
        );
      default:
        return <div>Section not found</div>;
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Sidebar */}
      <Sidebar
        activeSection={activeSection}
        onSectionChange={setActiveSection}
        onCollapsedChange={setIsSidebarCollapsed}
      />

      {/* Main Content */}
      <div
        className={`transition-all duration-300 ${isSidebarCollapsed ? "ml-16" : "ml-64"}`}
      >
        <div className="p-8 pb-20">{renderMainContent()}</div>
      </div>

      {/* Logs Drawer */}
      <LogsDrawer
        selectedService={selectedService}
        searchTerm={searchTerm}
        copied={copied}
        isCopyingLogs={isCopyingLogs}
        isClearingLogs={isClearingLogs}
        onSearchChange={setSearchTerm}
        onClearSearch={clearSearch}
        onCopyLogs={copyLogsToClipboard}
        onClearLogs={clearLogs}
        onClose={() => setSelectedService(null)}
        onOpenAdvancedSearch={() => setActiveSection("logs")}
      />

      {/* Modal Components */}
      <ServiceConfigModal
        service={editingService}
        isOpen={showServiceConfig}
        isSaving={isSavingService}
        onClose={() => {
          setShowServiceConfig(false);
          setEditingService(null);
        }}
        onSave={async (service) => {
          try {
            setIsSavingService(true);
            const response = await fetch(`/api/services/${service.name}`, {
              method: "PUT",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify(service),
            });

            if (!response.ok) {
              throw new Error(
                `Failed to save service: ${response.status} ${response.statusText}`,
              );
            }

            addToast(
              toast.success(
                "Service saved",
                `${service.name} configuration has been updated`,
              ),
            );
            fetchServices();
            setShowServiceConfig(false);
            setEditingService(null);
          } catch (error) {
            console.error("Failed to save service:", error);
            addToast(
              toast.error(
                "Failed to save service",
                error instanceof Error
                  ? error.message
                  : "An unexpected error occurred",
              ),
            );
          } finally {
            setIsSavingService(false);
          }
        }}
      />

      <ServiceFilesModal
        serviceName={viewingFilesService?.name || ""}
        serviceDir={viewingFilesService?.dir || ""}
        isOpen={showServiceFiles}
        onClose={() => {
          setShowServiceFiles(false);
          setViewingFilesService(null);
        }}
      />

      <ServiceEnvModal
        serviceName={envEditingService?.name || ""}
        isOpen={showServiceEnv}
        onClose={() => {
          setShowServiceEnv(false);
          setEnvEditingService(null);
        }}
      />
    </div>
  );
}

function App() {
  return (
    <ErrorBoundary>
      <ToastProvider>
        <ConfirmDialogProvider>
          <AppContent />
          <ToastContainer />
        </ConfirmDialogProvider>
      </ToastProvider>
    </ErrorBoundary>
  );
}

export default App;
