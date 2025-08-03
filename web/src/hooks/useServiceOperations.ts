import { useState, useCallback } from "react";
import { Service } from "@/types";
import {
  ServiceOperations,
  ServiceLoadingStates,
} from "@/services/serviceOperations";
import { useToast, toast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";

export function useServiceOperations() {
  const { token } = useAuth();
  const { addToast } = useToast();

  // Global loading states
  const [isStartingAll, setIsStartingAll] = useState(false);
  const [isStoppingAll, setIsStoppingAll] = useState(false);
  const [isFixingLombok, setIsFixingLombok] = useState(false);
  const [isSyncingEnvironment, setIsSyncingEnvironment] = useState(false);

  // Individual service loading states
  const [serviceLoadingStates, setServiceLoadingStates] =
    useState<ServiceLoadingStates>({});

  const startService = useCallback(
    async (service: Service) => {
      try {
        setServiceLoadingStates((prev) => ({
          ...prev,
          [service.id]: { ...prev[service.id], starting: true },
        }));

        const result = await ServiceOperations.startService(service.id);

        if (result.success) {
          addToast(toast.success("Service starting", result.message!));
        } else {
          addToast(toast.error("Failed to start service", result.error!));
        }
      } catch (error) {
        console.error('Service start operation failed:', error);
        addToast(toast.error("Failed to start service", error instanceof Error ? error.message : "Unknown error"));
      } finally {
        // Always clear loading state after operation completes
        setTimeout(() => {
          setServiceLoadingStates((prev) => ({
            ...prev,
            [service.id]: { ...prev[service.id], starting: false },
          }));
        }, 1000); // Brief delay to show feedback
      }
    },
    [addToast],
  );

  const stopService = useCallback(
    async (service: Service) => {
      try {
        setServiceLoadingStates((prev) => ({
          ...prev,
          [service.id]: { ...prev[service.id], stopping: true },
        }));

        const result = await ServiceOperations.stopService(service.id);

        if (result.success) {
          addToast(toast.success("Service stopping", result.message!));
        } else {
          addToast(toast.error("Failed to stop service", result.error!));
        }
      } catch (error) {
        console.error('Service stop operation failed:', error);
        addToast(toast.error("Failed to stop service", error instanceof Error ? error.message : "Unknown error"));
      } finally {
        // Always clear loading state after operation completes
        setTimeout(() => {
          setServiceLoadingStates((prev) => ({
            ...prev,
            [service.id]: { ...prev[service.id], stopping: false },
          }));
        }, 1000); // Brief delay to show feedback
      }
    },
    [addToast],
  );

  const restartService = useCallback(
    async (service: Service) => {
      try {
        setServiceLoadingStates((prev) => ({
          ...prev,
          [service.id]: { ...prev[service.id], restarting: true },
        }));

        const result = await ServiceOperations.restartService(service.id);

        if (result.success) {
          addToast(toast.success("Service restarting", result.message!));
        } else {
          addToast(toast.error("Failed to restart service", result.error!));
        }
      } catch (error) {
        console.error('Service restart operation failed:', error);
        addToast(toast.error("Failed to restart service", error instanceof Error ? error.message : "Unknown error"));
      } finally {
        // Always clear loading state after operation completes
        setTimeout(() => {
          setServiceLoadingStates((prev) => ({
            ...prev,
            [service.id]: { ...prev[service.id], restarting: false },
          }));
        }, 1000); // Brief delay to show feedback
      }
    },
    [addToast],
  );

  const checkServiceHealth = useCallback(
    async (service: Service) => {
      setServiceLoadingStates((prev) => ({
        ...prev,
        [service.id]: { ...prev[service.id], checkingHealth: true },
      }));

      const result = await ServiceOperations.checkServiceHealth(service.id);

      if (result.success) {
        addToast(toast.info("Health check initiated", result.message!));
      } else {
        addToast(toast.error("Failed to check service health", result.error!));
      }

      // Clear health check loading state after a short delay
      setTimeout(() => {
        setServiceLoadingStates((prev) => ({
          ...prev,
          [service.id]: { ...prev[service.id], checkingHealth: false },
        }));
      }, 1000);
    },
    [addToast],
  );

  const installLibraries = useCallback(
    async (service: Service) => {
      setServiceLoadingStates((prev) => ({
        ...prev,
        [service.id]: { ...prev[service.id], installingLibraries: true },
      }));

      const result = await ServiceOperations.installLibraries(service);

      if (result.success) {
        addToast(
          toast.success("Library installation started", result.message!),
        );
      } else {
        addToast(toast.error("Failed to install libraries", result.error!));
      }

      // Clear installation loading state after a delay
      setTimeout(() => {
        setServiceLoadingStates((prev) => ({
          ...prev,
          [service.id]: { ...prev[service.id], installingLibraries: false },
        }));
      }, 2000);
    },
    [addToast],
  );

  const startAllServices = useCallback(async () => {
    setIsStartingAll(true);

    const result = await ServiceOperations.startAllServices(token || undefined);

    if (result.success) {
      addToast(toast.success("Starting services", result.message!));
    } else {
      addToast(toast.error("Failed to start all services", result.error!));
    }

    setIsStartingAll(false);
  }, [token, addToast]);

  const stopAllServices = useCallback(async () => {
    setIsStoppingAll(true);

    const result = await ServiceOperations.stopAllServices(token || undefined);

    if (result.success) {
      addToast(toast.success("Stopping services", result.message!));
    } else {
      addToast(toast.error("Failed to stop all services", result.error!));
    }

    setIsStoppingAll(false);
  }, [token, addToast]);

  const fixLombok = useCallback(async () => {
    setIsFixingLombok(true);
    addToast(
      toast.info(
        "Fixing Lombok",
        "Checking and fixing Lombok compatibility for all services...",
      ),
    );

    const result = await ServiceOperations.fixLombok();

    if (result.success) {
      addToast(toast.success("Lombok Fix Complete", result.message!));
    } else {
      addToast(toast.error("Error", result.error!));
    }

    setIsFixingLombok(false);
  }, [addToast]);

  const syncEnvironment = useCallback(async () => {
    setIsSyncingEnvironment(true);
    addToast(
      toast.info(
        "Syncing Environment",
        "Synchronizing environment variables from configuration files...",
      ),
    );

    const result = await ServiceOperations.syncEnvironment();

    if (result.success) {
      addToast(toast.success("Environment Sync Complete", result.message!));
    } else {
      addToast(toast.warning("Environment Sync Complete", result.message!));
    }

    setIsSyncingEnvironment(false);
  }, [addToast]);

  const updateServiceLoadingState = useCallback((updatedService: Service) => {
    setServiceLoadingStates((prev) => ({
      ...prev,
      [updatedService.id]: {
        starting:
          updatedService.status === "running"
            ? false
            : prev[updatedService.id]?.starting || false,
        stopping:
          updatedService.status === "stopped"
            ? false
            : prev[updatedService.id]?.stopping || false,
        restarting: false,
        checkingHealth: false,
        installingLibraries: prev[updatedService.id]?.installingLibraries || false,
        validatingWrapper: prev[updatedService.id]?.validatingWrapper || false,
        generatingWrapper: prev[updatedService.id]?.generatingWrapper || false,
        repairingWrapper: prev[updatedService.id]?.repairingWrapper || false,
      },
    }));
  }, []);

  const clearServiceLoadingState = useCallback((serviceId: string) => {
    setServiceLoadingStates((prev) => ({
      ...prev,
      [serviceId]: {
        starting: false,
        stopping: false,
        restarting: false,
        checkingHealth: false,
        installingLibraries: false,
        validatingWrapper: false,
        generatingWrapper: false,
        repairingWrapper: false,
      },
    }));
  }, []);

  const clearAllLoadingStates = useCallback(() => {
    setServiceLoadingStates({});
  }, []);

  // Wrapper Management Operations

  const validateWrapper = useCallback(
    async (service: Service) => {
      setServiceLoadingStates((prev) => ({
        ...prev,
        [service.id]: { ...prev[service.id], validatingWrapper: true },
      }));

      const result = await ServiceOperations.validateWrapper(service.id);

      if (result.success) {
        addToast(toast.info("Wrapper validation", result.message!));
      } else {
        addToast(toast.error("Failed to validate wrapper", result.error!));
      }

      setTimeout(() => {
        setServiceLoadingStates((prev) => ({
          ...prev,
          [service.id]: { ...prev[service.id], validatingWrapper: false },
        }));
      }, 1000);

      return result;
    },
    [addToast],
  );

  const generateWrapper = useCallback(
    async (service: Service) => {
      setServiceLoadingStates((prev) => ({
        ...prev,
        [service.id]: { ...prev[service.id], generatingWrapper: true },
      }));

      const result = await ServiceOperations.generateWrapper(service.id);

      if (result.success) {
        addToast(toast.success("Wrapper generated", result.message!));
      } else {
        addToast(toast.error("Failed to generate wrapper", result.error!));
      }

      setTimeout(() => {
        setServiceLoadingStates((prev) => ({
          ...prev,
          [service.id]: { ...prev[service.id], generatingWrapper: false },
        }));
      }, 2000);
    },
    [addToast],
  );

  const repairWrapper = useCallback(
    async (service: Service) => {
      setServiceLoadingStates((prev) => ({
        ...prev,
        [service.id]: { ...prev[service.id], repairingWrapper: true },
      }));

      const result = await ServiceOperations.repairWrapper(service.id);

      if (result.success) {
        addToast(toast.success("Wrapper repaired", result.message!));
      } else {
        addToast(toast.error("Failed to repair wrapper", result.error!));
      }

      setTimeout(() => {
        setServiceLoadingStates((prev) => ({
          ...prev,
          [service.id]: { ...prev[service.id], repairingWrapper: false },
        }));
      }, 2000);
    },
    [addToast],
  );

  return {
    // Loading states
    isStartingAll,
    isStoppingAll,
    isFixingLombok,
    isSyncingEnvironment,
    serviceLoadingStates,

    // Service operations
    startService,
    stopService,
    restartService,
    checkServiceHealth,
    installLibraries,
    startAllServices,
    stopAllServices,
    fixLombok,
    syncEnvironment,
    updateServiceLoadingState,
    clearServiceLoadingState,
    clearAllLoadingStates,

    // Wrapper management operations
    validateWrapper,
    generateWrapper,
    repairWrapper,
  };
}
