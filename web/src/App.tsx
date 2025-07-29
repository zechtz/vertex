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
import { AuthContainer } from "@/components/Auth/AuthContainer";
import { AuthProvider, useAuth } from "@/contexts/AuthContext";
import { ThemeProvider } from "@/contexts/ThemeContext";
import { ProfileProvider, useProfile } from "@/contexts/ProfileContext";
import { Toolbar } from "@/components/Toolbar/Toolbar";
import { ProfileManagement } from "@/components/ProfileManagement";
import { ProfileConfigDashboard } from "@/components/ProfileConfigDashboard/ProfileConfigDashboard";
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
import { ServiceActionModal } from "@/components/ServiceActionModal/ServiceActionModal";

function AuthenticatedApp() {
  const { user, logout, token } = useAuth();
  const { activeProfile, removeServiceFromProfile } = useProfile();
  const [services, setServices] = useState<Service[]>([]);
  const [allServices, setAllServices] = useState<Service[]>([]);
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
  const [isCreatingService, setIsCreatingService] = useState(false);
  const [showServiceFiles, setShowServiceFiles] = useState(false);
  const [showServiceEnv, setShowServiceEnv] = useState(false);
  const [editingService, setEditingService] = useState<Service | null>(null);
  const [viewingFilesService, setViewingFilesService] =
    useState<Service | null>(null);
  const [envEditingService, setEnvEditingService] = useState<Service | null>(
    null,
  );
  
  // Service action modal state
  const [showServiceActionModal, setShowServiceActionModal] = useState(false);
  const [serviceToAction, setServiceToAction] = useState<Service | null>(null);

  // Configuration state
  const [configurations, setConfigurations] = useState<Configuration[]>([]);
  const [allConfigurations, setAllConfigurations] = useState<Configuration[]>([]);

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
      setAllServices(sortedServices);
      
      // Filter services based on active profile
      filterServicesByProfile(sortedServices, activeProfile);
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

  const filterServicesByProfile = (allServices: Service[], activeProfile: any) => {
    if (!activeProfile || !activeProfile.services || activeProfile.services.length === 0) {
      // If no active profile or no services in profile, show all services
      setServices(allServices);
    } else {
      // Filter to show only services that are in the active profile
      const filteredServices = allServices.filter(service => 
        activeProfile.services.includes(service.name)
      );
      setServices(filteredServices);
    }
  };

  const filterConfigurationsByProfile = (allConfigs: Configuration[], activeProfile: any) => {
    if (!activeProfile || !activeProfile.services || activeProfile.services.length === 0) {
      // If no active profile, show all configurations
      setConfigurations(allConfigs);
    } else {
      // Filter to show only configurations that contain services from the active profile
      const filteredConfigs = allConfigs.filter(config => 
        config.services.some(configService => 
          activeProfile.services.includes(configService.name)
        )
      );
      setConfigurations(filteredConfigs);
    }
  };

  // Effect to re-filter services when active profile changes
  useEffect(() => {
    if (allServices.length > 0) {
      filterServicesByProfile(allServices, activeProfile);
    }
  }, [activeProfile, allServices]);

  // Effect to re-filter configurations when active profile changes
  useEffect(() => {
    if (allConfigurations.length > 0) {
      filterConfigurationsByProfile(allConfigurations, activeProfile);
    }
  }, [activeProfile, allConfigurations]);

  const fetchConfigurations = async () => {
    try {
      const response = await fetch("/api/configurations");
      if (!response.ok) {
        throw new Error(
          `Failed to fetch configurations: ${response.status} ${response.statusText}`,
        );
      }
      const data = await response.json();
      setAllConfigurations(data);
      
      // Filter configurations based on active profile
      filterConfigurationsByProfile(data, activeProfile);
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
      const headers: Record<string, string> = {};
      
      // Only include Authorization header if token exists
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
      
      const response = await fetch("/api/services/start-all-profile", {
        method: "POST",
        headers,
      });
      if (!response.ok) {
        throw new Error(
          `Failed to start all services: ${response.status} ${response.statusText}`,
        );
      }
      
      const result = await response.json();
      addToast(
        toast.success(
          "Starting services",
          result.status || "Services are being started",
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
    const profileContext = activeProfile ? ` in profile "${activeProfile.name}"` : "";
    const confirmed = await showConfirm({
      ...confirmDialogs.stopAllServices(),
      description: `Are you sure you want to stop all services${profileContext}? This will stop all currently running services${profileContext}.`,
    });
    if (!confirmed) return;

    try {
      setIsStoppingAll(true);
      const headers: Record<string, string> = {};
      
      // Only include Authorization header if token exists
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
      
      const response = await fetch("/api/services/stop-all-profile", {
        method: "POST",
        headers,
      });
      if (!response.ok) {
        throw new Error(
          `Failed to stop all services: ${response.status} ${response.statusText}`,
        );
      }
      
      const result = await response.json();
      addToast(
        toast.success(
          "Stopping services",
          result.status || "Services are being stopped",
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
      lastStarted: new Date().toISOString(), // Use proper ISO string format
      description: "",
      isEnabled: true,
      buildSystem: "auto",
      envVars: {},
      logs: [],
      uptime: "",
      cpuPercent: 0,
      memoryUsage: 0,
      memoryPercent: 0,
      diskUsage: 0,
      networkRx: 0,
      networkTx: 0,
      metrics: {
        responseTimes: [],
        errorRate: 0,
        requestCount: 0,
        lastChecked: new Date().toISOString(),
      },
      dependencies: null,
      dependentOn: null,
      startupDelay: 0,
    });
    setIsCreatingService(true);
    setShowServiceConfig(true);
  };

  const openEditService = (service: Service) => {
    setEditingService(service);
    setIsCreatingService(false);
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
    const service = services.find(s => s.name === serviceName);
    if (!service) return;
    
    setServiceToAction(service);
    setShowServiceActionModal(true);
  };

  const handleRemoveFromProfile = async (serviceName: string) => {
    if (!activeProfile) return;
    
    try {
      await removeServiceFromProfile(activeProfile.id, serviceName);
      addToast(
        toast.success(
          "Service removed from profile",
          `${serviceName} has been removed from ${activeProfile.name}`,
        ),
      );
      fetchServices();
    } catch (error) {
      console.error("Failed to remove service from profile:", error);
      addToast(
        toast.error(
          "Failed to remove service from profile",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    }
  };

  const handleDeleteGlobally = async (serviceName: string) => {
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
          `${serviceName} has been deleted permanently`,
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
      case "profiles":
        return (
          <ProfileManagement
            isOpen={true}
            onClose={() => setActiveSection("services")}
            onProfileUpdated={fetchServices}
          />
        );
      case "dashboard":
        return (
          <ProfileConfigDashboard
            isOpen={true}
            onClose={() => setActiveSection("services")}
          />
        );
      case "metrics":
        return (
          <SystemMetricsModal
            isOpen={true}
            onClose={() => setActiveSection("services")}
            services={services}
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
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Toolbar */}
      <Toolbar
        user={user ? {
          id: user.id,
          username: user.username,
          email: user.email,
          role: user.role,
          lastLogin: user.lastLogin
        } : null}
        onLogout={logout}
        onToggleSidebar={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
        isSidebarCollapsed={isSidebarCollapsed}
      />

      {/* Sidebar */}
      <Sidebar
        activeSection={activeSection}
        onSectionChange={setActiveSection}
        isCollapsed={isSidebarCollapsed}
      />

      {/* Main Content */}
      <div
        className={`transition-all duration-300 pt-16 ${isSidebarCollapsed ? "ml-16" : "ml-64"}`}
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
          setIsCreatingService(false);
        }}
        isCreateMode={isCreatingService}
        onSave={async (service, profileId) => {
          try {
            setIsSavingService(true);
            const isCreate = isCreatingService;
            
            const url = isCreate ? "/api/services" : `/api/services/${service.name}`;
            const method = isCreate ? "POST" : "PUT";
            
            console.log("Saving service:", { method, url, service });
            const response = await fetch(url, {
              method: method,
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify(service),
            });

            if (!response.ok) {
              throw new Error(
                `Failed to ${isCreate ? 'create' : 'save'} service: ${response.status} ${response.statusText}`,
              );
            }

            // If creating a new service and a profile was selected, add it to the profile
            if (isCreate && profileId) {
              try {
                const addToProfileResponse = await fetch(`/api/profiles/${profileId}/services`, {
                  method: 'POST',
                  headers: { 'Content-Type': 'application/json' },
                  body: JSON.stringify({ serviceName: service.name })
                });
                if (!addToProfileResponse.ok) {
                  console.warn('Failed to add service to profile, but service was created successfully');
                }
              } catch (error) {
                console.warn('Failed to add service to profile:', error);
              }
            }

            addToast(
              toast.success(
                isCreate ? "Service created" : "Service saved",
                `${service.name} has been ${isCreate ? 'created' : 'updated'} successfully${profileId && isCreate ? ' and added to profile' : ''}`,
              ),
            );
            fetchServices();
            setShowServiceConfig(false);
            setEditingService(null);
            setIsCreatingService(false);
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

      <ServiceActionModal
        isOpen={showServiceActionModal}
        onClose={() => {
          setShowServiceActionModal(false);
          setServiceToAction(null);
        }}
        service={serviceToAction}
        activeProfile={activeProfile}
        onRemoveFromProfile={handleRemoveFromProfile}
        onDeleteGlobally={handleDeleteGlobally}
      />
    </div>
  );
}

function AppContent() {
  const { isAuthenticated, isLoading, login } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-purple-50">
        <div className="text-center">
          <div className="h-12 w-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-gray-600">Loading NeST Manager...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <AuthContainer
        onAuthSuccess={(user, token) => {
          login(user, token);
        }}
      />
    );
  }

  return <AuthenticatedApp />;
}

function App() {
  return (
    <ErrorBoundary>
      <ThemeProvider>
        <AuthProvider>
          <ProfileProvider>
            <ToastProvider>
              <ConfirmDialogProvider>
                <AppContent />
                <ToastContainer />
              </ConfirmDialogProvider>
            </ToastProvider>
          </ProfileProvider>
        </AuthProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
}

export default App;
