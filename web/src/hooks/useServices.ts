import { useState, useEffect, useCallback, useRef } from "react";
import { Service, Configuration, LogEntry } from "@/types";
import { ServiceOperations } from "@/services/serviceOperations";
import { useProfile } from "@/contexts/ProfileContext";
import { useToast, toast } from "@/components/ui/toast";

import { apiFetch } from "@/services/apiFetch";

/**
 * The realtime socket is the only thing telling the dashboard a service
 * started, stopped or died. When it drops - a sleep, a server restart, a
 * network blip - the UI keeps rendering the last states it heard, which look
 * exactly like current ones. So the socket reconnects on its own, and says
 * out loud when it is not connected.
 */
const RECONNECT_BASE_DELAY_MS = 1000;
const RECONNECT_MAX_DELAY_MS = 30 * 1000;

/**
 * Log lines kept per service. Logs stream in for as long as the tab is open,
 * so without a ceiling a chatty service grows this array until the tab slows
 * down. The oldest lines go first - the tail is what anyone is reading.
 */
const MAX_LOG_ENTRIES = 1000;

function appendLogEntry(logs: LogEntry[], entry: LogEntry): LogEntry[] {
  const next = [...logs, entry];
  return next.length > MAX_LOG_ENTRIES
    ? next.slice(next.length - MAX_LOG_ENTRIES)
    : next;
}

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
  // Optimistic: the socket opens within milliseconds of mount, and starting
  // at false makes every page load flash the reconnecting pill. A connection
  // that genuinely fails closes right away, which corrects this.
  const [isRealtimeConnected, setIsRealtimeConnected] = useState(true);

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
          typeof s === "string" ? s : s.id,
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
      const response = await apiFetch("/api/configurations");
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
          typeof s === "string" ? s : s.id,
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

  useEffect(() => {
    fetchServices();
    fetchConfigurations();
  }, [fetchServices, fetchConfigurations]);

  // Reached from inside the socket, which is mounted once and so cannot close
  // over a callback that changes with the active profile.
  const fetchServicesRef = useRef(fetchServices);
  useEffect(() => {
    fetchServicesRef.current = fetchServices;
  }, [fetchServices]);

  // Realtime updates. Mounted once for the life of the hook: reconnecting is
  // handled here rather than by re-running the effect, so selecting a service
  // or switching profile no longer tears the connection down and rebuilds it.
  useEffect(() => {
    let socket: WebSocket | null = null;
    let retryTimer: number | undefined;
    let attempt = 0;
    let unmounted = false;

    const handleMessage = (raw: string) => {
      let message: { type?: string; payload?: any };

      try {
        message = JSON.parse(raw);
      } catch {
        // One unreadable frame is not a reason to lose the stream.
        console.error("Ignoring unreadable realtime message");
        return;
      }

      if (message.type === "service_update") {
        const updatedService: Service = message.payload;

        setServices((prev) =>
          prev.map((service) =>
            service.id === updatedService.id ? updatedService : service,
          ),
        );
        setSelectedService((prev) =>
          prev && prev.id === updatedService.id ? updatedService : prev,
        );
      } else if (message.type === "log_entry") {
        const { serviceId, logEntry } = message.payload;

        setServices((prev) =>
          prev.map((service) =>
            service.id === serviceId
              ? { ...service, logs: appendLogEntry(service.logs, logEntry) }
              : service,
          ),
        );
        setSelectedService((prev) =>
          prev && prev.id === serviceId
            ? { ...prev, logs: appendLogEntry(prev.logs, logEntry) }
            : prev,
        );
      }
    };

    const connect = () => {
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      socket = new WebSocket(`${protocol}//${window.location.host}/ws`);

      socket.onopen = () => {
        setIsRealtimeConnected(true);

        // Whatever happened while the socket was down was never delivered, so
        // what is on screen is stale. Reconnecting is only half the recovery.
        if (attempt > 0) fetchServicesRef.current();
        attempt = 0;
      };

      socket.onmessage = (event) => handleMessage(event.data);

      // An error is always followed by a close, so the retry lives there only.
      socket.onerror = () => socket?.close();

      socket.onclose = () => {
        if (unmounted) return;

        setIsRealtimeConnected(false);

        // Backoff, capped: a server that is down should not be hammered, but a
        // machine waking from sleep should be back within a second or two.
        const delay = Math.min(
          RECONNECT_BASE_DELAY_MS * 2 ** attempt,
          RECONNECT_MAX_DELAY_MS,
        );
        attempt += 1;
        retryTimer = window.setTimeout(connect, delay);
      };
    };

    connect();

    return () => {
      unmounted = true;
      window.clearTimeout(retryTimer);
      socket?.close();
    };
  }, []);
  return {
    // State
    services,
    allServices,
    configurations,
    allConfigurations,
    selectedService,
    isLoading,
    isRealtimeConnected,

    // Actions
    setSelectedService,
    fetchServices,
    fetchConfigurations,

    // Utilities
    filterServicesByProfile,
    filterConfigurationsByProfile,
  };
}
