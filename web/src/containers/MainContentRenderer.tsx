import { ServicesGrid } from "@/components/ServicesGrid/ServicesGrid";
import { ProfileManagement } from "@/components/ProfileManagement";
import { ProfileConfigDashboard } from "@/components/ProfileConfigDashboard/ProfileConfigDashboard";
import { SystemMetricsModal } from "@/components/SystemMetricsModal/SystemMetricsModal";
import { LogAggregationModal } from "@/components/LogAggregationModal/LogAggregationModal";
import { ServiceTopologyModal } from "@/components/ServiceTopologyModal/ServiceTopologyModal";
import { DependencyConfigModal } from "@/components/DependencyConfigModal/DependencyConfigModal";
import { AutoDiscoveryModal } from "@/components/AutoDiscoveryModal/AutoDiscoveryModal";
import { ConfigurationManager } from "@/components/ConfigurationManager/ConfigurationManager";
import { GlobalEnvModal } from "@/components/GlobalEnvModal/GlobalEnvModal";
import { GlobalConfigModal } from "@/components/GlobalConfigModal/GlobalConfigModal";
import { Service } from "@/types";
import { useProfile } from "@/contexts/ProfileContext";

interface MainContentRendererProps {
  activeSection: string;
  onSectionChange: (section: string) => void;
  servicesData: any;
  serviceOps: any;
  serviceManagement: any;
  activeProfile: any;
}

export function MainContentRenderer({
  activeSection,
  onSectionChange,
  servicesData,
  serviceOps,
  serviceManagement,
}: MainContentRendererProps) {
  const { refreshProfiles } = useProfile();
  const renderMainContent = () => {
    switch (activeSection) {
      case "services":
        return (
          <ServicesGrid
            services={servicesData.services}
            isLoading={servicesData.isLoading}
            isStartingAll={serviceOps.isStartingAll}
            isStoppingAll={serviceOps.isStoppingAll}
            isFixingLombok={serviceOps.isFixingLombok}
            isSyncingEnvironment={serviceOps.isSyncingEnvironment}
            serviceLoadingStates={serviceOps.serviceLoadingStates}
            onStartAll={serviceOps.startAllServices}
            onStopAll={serviceOps.stopAllServices}
            onFixLombok={serviceOps.fixLombok}
            onSyncEnvironment={serviceOps.syncEnvironment}
            onCreateService={serviceManagement.openCreateService}
            onStartService={serviceOps.startService}
            onStopService={serviceOps.stopService}
            onRestartService={serviceOps.restartService}
            onCheckHealth={serviceOps.checkServiceHealth}
            onViewLogs={servicesData.setSelectedService}
            onEditService={serviceManagement.openEditService}
            onDeleteService={(service: Service) =>
              serviceManagement.deleteService(
                service.name,
                servicesData.services,
              )
            }
            onViewFiles={serviceManagement.openViewFiles}
            onEditEnv={serviceManagement.openEditEnv}
            onManageWrappers={(service: Service) => {
              console.log('Manage wrappers for service:', service.name);
              // TODO: Implement wrapper management modal integration
            }}
          />
        );
      case "profiles":
        return (
          <ProfileManagement
            isOpen={true}
            onClose={() => onSectionChange("services")}
            onProfileUpdated={servicesData.fetchServices}
          />
        );
      case "dashboard":
        return (
          <ProfileConfigDashboard
            isOpen={true}
            onClose={() => onSectionChange("services")}
          />
        );
      case "metrics":
        return (
          <SystemMetricsModal
            isOpen={true}
            onClose={() => onSectionChange("services")}
            services={servicesData.services}
          />
        );
      case "logs":
        return (
          <LogAggregationModal
            isOpen={true}
            onClose={() => onSectionChange("services")}
            services={servicesData.services}
          />
        );
      case "topology":
        return (
          <ServiceTopologyModal
            isOpen={true}
            onClose={() => onSectionChange("services")}
          />
        );
      case "dependencies":
        return (
          <DependencyConfigModal
            isOpen={true}
            onClose={() => onSectionChange("services")}
            services={servicesData.services}
          />
        );
      case "auto-discovery":
        return (
          <AutoDiscoveryModal
            isOpen={true}
            onClose={() => onSectionChange("services")}
            onServiceImported={async () => {
              // Refresh both services and profile data
              await Promise.all([
                servicesData.fetchServices(),
                refreshProfiles(),
              ]);
              onSectionChange("services");
            }}
          />
        );
      case "configurations":
        return (
          <ConfigurationManager
            isOpen={true}
            onClose={() => onSectionChange("services")}
            configurations={servicesData.configurations}
            services={servicesData.services}
            onConfigurationSaved={async () => {
              // Refresh configurations, services, and profile data
              await Promise.all([
                servicesData.fetchConfigurations(),
                servicesData.fetchServices(),
                refreshProfiles(),
              ]);
            }}
          />
        );
      case "environment":
        return (
          <GlobalEnvModal
            isOpen={true}
            onClose={() => onSectionChange("services")}
          />
        );
      case "settings":
        return (
          <GlobalConfigModal
            isOpen={true}
            onClose={() => onSectionChange("services")}
            onConfigUpdated={servicesData.fetchServices}
          />
        );
      default:
        return <div>Section not found</div>;
    }
  };

  return renderMainContent();
}
