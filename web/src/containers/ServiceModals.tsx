import { ServiceConfigModal } from "@/components/ServiceConfigModal/ServiceConfigModal";
import { ServiceFilesModal } from "@/components/ServiceConfigModal/ServiceFilesModal";
import { ServiceEnvModal } from "@/components/ServiceEnvModal/ServiceEnvModal";
import { ServiceActionModal } from "@/components/ServiceActionModal/ServiceActionModal";
import { LibraryInstallModal } from "@/components/LibraryInstallModal";
import WrapperManagementModal from "@/components/WrapperManagementModal/WrapperManagementModal";
import { useProfile } from "@/contexts/ProfileContext";
import { ServiceOperations } from "@/services/serviceOperations";

interface ServiceModalsProps {
  serviceManagement: any;
  onServiceSaved: () => void;
}

export function ServiceModals({ serviceManagement }: ServiceModalsProps) {
  const { activeProfile } = useProfile();

  return (
    <>
      <ServiceConfigModal
        service={serviceManagement.serviceConfigData}
        isOpen={serviceManagement.isServiceConfigOpen}
        isSaving={serviceManagement.isSavingService}
        onClose={serviceManagement.closeServiceConfig}
        isCreateMode={serviceManagement.isCreatingService}
        onSave={serviceManagement.handleSaveService}
      />

      <ServiceFilesModal
        serviceName={serviceManagement.serviceFilesData?.name || ""}
        serviceId={serviceManagement.serviceFilesData?.id || ""}
        serviceDir={serviceManagement.serviceFilesData?.dir || ""}
        isOpen={serviceManagement.isServiceFilesOpen}
        onClose={serviceManagement.closeServiceFiles}
      />

      <ServiceEnvModal
        serviceName={serviceManagement.serviceEnvData?.name || ""}
        serviceId={serviceManagement.serviceEnvData?.id || ""}
        isOpen={serviceManagement.isServiceEnvOpen}
        onClose={serviceManagement.closeServiceEnv}
      />

      <ServiceActionModal
        isOpen={serviceManagement.isServiceActionOpen}
        onClose={serviceManagement.closeServiceActionModal}
        service={serviceManagement.serviceActionData}
        activeProfile={activeProfile}
        onRemoveFromProfile={serviceManagement.handleRemoveFromProfile}
        onDeleteGlobally={serviceManagement.handleDeleteGlobally}
      />

      <LibraryInstallModal
        serviceId={serviceManagement.libraryInstallData?.id || ""}
        serviceName={serviceManagement.libraryInstallData?.name || ""}
        isOpen={serviceManagement.isLibraryInstallOpen}
        onClose={serviceManagement.closeLibraryInstall}
      />

      <WrapperManagementModal
        serviceId={serviceManagement.wrapperManagementData?.id || ""}
        serviceName={serviceManagement.wrapperManagementData?.name || ""}
        isOpen={serviceManagement.isWrapperManagementOpen}
        onClose={serviceManagement.closeWrapperManagement}
        onValidateWrapper={() => ServiceOperations.validateWrapper(serviceManagement.wrapperManagementData?.id || "")}
        onGenerateWrapper={async () => {
          await ServiceOperations.generateWrapper(serviceManagement.wrapperManagementData?.id || "");
        }}
        onRepairWrapper={async () => {
          await ServiceOperations.repairWrapper(serviceManagement.wrapperManagementData?.id || "");
        }}
      />
    </>
  );
}
