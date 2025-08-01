import { useState, useEffect, useCallback } from "react";
import { Service, Configuration } from "@/types";
import { ServiceOperations } from "@/services/serviceOperations";
import { useProfile } from "@/contexts/ProfileContext";
import { useToast, toast } from "@/components/ui/toast";

export function useServices() {
  const { activeProfile } = useProfile();
  const { addToast } = useToast();

  const [services, setServices] = useState<Service[]>([]);
  const [allServices, setAllServices] = useState<Service[]>([]);
  const [configurations, setConfigurations] = useState<Configuration[]>([]);
  const [allConfigurations, setAllConfigurations] = useState<Configuration[]>(
    [],
  );
  const [selectedService, setSelectedService] = useState<Service | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const fetchServices = useCallback(async () => {
    try {
      const sortedServices = await ServiceOperations.fetchServices();
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
  }, [activeProfile, addToast]);

  const filterServicesByProfile = useCallback(
    (allServices: Service[], activeProfile: any) => {
      if (!activeProfile) {
        // If no active profile, show all services (global view)
        setServices(allServices);
      } else if (
        !activeProfile.services ||
        activeProfile.services.length === 0
      ) {
        // If profile exists but has no services, show empty list (not all services)
        setServices([]);
      } else {
        // Filter to show only services that are in the active profile
        const profileServiceIds = activeProfile.services.map((s: any) => 
          typeof s === 'string' ? s : s.id
        );
        const filteredServices = allServices.filter((service) =>
          profileServiceIds.includes(service.id),
        );
        setServices(filteredServices);
      }
    },
    [],
  );

  const fetchConfigurations = useCallback(async () => {
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
  }, [activeProfile, addToast]);

  const filterConfigurationsByProfile = useCallback(
    (allConfigs: Configuration[], activeProfile: any) => {
      if (
        !activeProfile ||
        !activeProfile.services ||
        activeProfile.services.length === 0
      ) {
        // If no active profile, show all configurations
        setConfigurations(allConfigs);
      } else {
        // Filter to show only configurations that contain services from the active profile
        const profileServiceIds = activeProfile.services.map((s: any) => 
          typeof s === 'string' ? s : s.id
        );
        const filteredConfigs = allConfigs.filter((config) =>
          config.services.some((configService) =>
            profileServiceIds.includes(configService.id),
          ),
        );
        setConfigurations(filteredConfigs);
      }
    },
    [],
  );

  // Effect to re-filter services when active profile changes
  useEffect(() => {
    if (allServices.length > 0) {
      filterServicesByProfile(allServices, activeProfile);
    }
  }, [activeProfile, allServices, filterServicesByProfile]);

  // Effect to re-filter configurations when active profile changes
  useEffect(() => {
    if (allConfigurations.length > 0) {
      filterConfigurationsByProfile(allConfigurations, activeProfile);
    }
  }, [activeProfile, allConfigurations, filterConfigurationsByProfile]);

  // WebSocket handling
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
            service.id === updatedService.id ? updatedService : service,
          ),
        );

        if (selectedService && selectedService.id === updatedService.id) {
          setSelectedService(updatedService);
        }
      } else if (message.type === "log_entry") {
        const { serviceId, logEntry } = message.payload;
        setServices((prev) =>
          prev.map((service) =>
            service.id === serviceId
              ? { ...service, logs: [...service.logs, logEntry] }
              : service,
          ),
        );
        if (selectedService && selectedService.id === serviceId) {
          setSelectedService((prev) =>
            prev ? { ...prev, logs: [...prev.logs, logEntry] } : null,
          );
        }
      }
    };

    return () => ws.close();
  }, [selectedService, fetchServices, fetchConfigurations]);

  return {
    // State
    services,
    allServices,
    configurations,
    allConfigurations,
    selectedService,
    isLoading,

    // Actions
    setSelectedService,
    fetchServices,
    fetchConfigurations,

    // Utilities
    filterServicesByProfile,
    filterConfigurationsByProfile,
  };
}
