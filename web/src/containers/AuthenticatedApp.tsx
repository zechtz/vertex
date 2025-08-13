import { useState } from "react";
import { Sidebar } from "@/components/Sidebar/Sidebar";
import { Toolbar } from "@/components/Toolbar/Toolbar";
import { LogsDrawer } from "@/components/LogsDrawer/LogsDrawer";
import { useAuth } from "@/contexts/AuthContext";
import { useProfile } from "@/contexts/ProfileContext";
import { useServices } from "@/hooks/useServices";
import { useServiceOperations } from "@/hooks/useServiceOperations";
import { useServiceManagement } from "@/hooks/useServiceManagement";
import { useLogsOperations } from "@/hooks/useLogsOperations";
import { useOnboarding } from "@/hooks/useOnboarding";
import { MainContentRenderer } from "./MainContentRenderer";
import { ServiceModals } from "./ServiceModals";
import { OnboardingStepper } from "@/components/Onboarding/OnboardingStepper";

export function AuthenticatedApp() {
  const { user, logout } = useAuth();
  const { activeProfile } = useProfile();

  // Custom hooks
  const servicesData = useServices();
  const serviceOps = useServiceOperations();
  const serviceManagement = useServiceManagement(servicesData.fetchServices);
  const logsOps = useLogsOperations();
  const onboarding = useOnboarding();

  // UI State
  const [activeSection, setActiveSection] = useState("services");
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");

  const clearSearch = () => {
    setSearchTerm("");
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Toolbar */}
      <Toolbar
        user={
          user
            ? {
                id: user.id,
                username: user.username,
                email: user.email,
                role: user.role,
                lastLogin: user.lastLogin,
              }
            : null
        }
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
        <div className="p-8 pb-20">
          <MainContentRenderer
            activeSection={activeSection}
            onSectionChange={setActiveSection}
            servicesData={servicesData}
            serviceOps={serviceOps}
            serviceManagement={serviceManagement}
            activeProfile={activeProfile}
            onboarding={onboarding}
          />
        </div>
      </div>

      {/* Logs Drawer */}
      <LogsDrawer
        selectedService={servicesData.selectedService}
        searchTerm={searchTerm}
        copied={logsOps.copied}
        isCopyingLogs={logsOps.isCopyingLogs}
        isClearingLogs={logsOps.isClearingLogs}
        onSearchChange={setSearchTerm}
        onClearSearch={clearSearch}
        onCopyLogs={(selectedLevels) =>
          logsOps.copyLogsToClipboard(
            servicesData.selectedService,
            selectedLevels,
          )
        }
        onClearLogs={() => logsOps.clearLogs(servicesData.selectedService)}
        onClose={() => servicesData.setSelectedService(null)}
        onOpenAdvancedSearch={() => setActiveSection("logs")}
      />

      {/* Service Modals */}
      <ServiceModals
        serviceManagement={serviceManagement}
        onServiceSaved={servicesData.fetchServices}
      />

      {/* Onboarding Stepper */}
      <OnboardingStepper
        isOpen={onboarding.isOnboardingOpen}
        onClose={onboarding.closeOnboarding}
        onComplete={onboarding.completeOnboarding}
      />
    </div>
  );
}
