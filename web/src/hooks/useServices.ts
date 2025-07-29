import { useState, useEffect, useCallback } from 'react';
import { Service, Configuration } from '@/types';
import { ServiceApi } from '@/services/serviceApi';
import { useProfile } from '@/contexts/ProfileContext';

export interface UseServicesReturn {
  // Data
  services: Service[];
  allServices: Service[];
  configurations: Configuration[];
  allConfigurations: Configuration[];
  isLoading: boolean;
  
  // Actions
  fetchServices: () => Promise<void>;
  fetchConfigurations: () => Promise<void>;
  createService: (serviceData: Partial<Service>) => Promise<Service>;
  updateService: (serviceName: string, serviceData: Partial<Service>) => Promise<Service>;
  deleteService: (serviceName: string) => Promise<void>;
  
  // Utility
  getServiceByName: (name: string) => Service | undefined;
  refreshData: () => Promise<void>;
}

export function useServices(): UseServicesReturn {
  const { activeProfile } = useProfile();
  const [services, setServices] = useState<Service[]>([]);
  const [allServices, setAllServices] = useState<Service[]>([]);
  const [configurations, setConfigurations] = useState<Configuration[]>([]);
  const [allConfigurations, setAllConfigurations] = useState<Configuration[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  /**
   * Filter services based on active profile
   */
  const filterServicesByProfile = useCallback((allServices: Service[]): Service[] => {
    if (!activeProfile) return allServices;
    return allServices.filter(service => 
      activeProfile.services.includes(service.name)
    );
  }, [activeProfile]);

  /**
   * Fetch services from API
   */
  const fetchServices = useCallback(async () => {
    try {
      const fetchedServices = await ServiceApi.fetchServices();
      setAllServices(fetchedServices);
      setServices(filterServicesByProfile(fetchedServices));
    } catch (error) {
      console.error('Failed to fetch services:', error);
      throw error;
    }
  }, [filterServicesByProfile]);

  /**
   * Fetch configurations from API
   */
  const fetchConfigurations = useCallback(async () => {
    try {
      const fetchedConfigurations = await ServiceApi.fetchConfigurations();
      setAllConfigurations(fetchedConfigurations);
      setConfigurations(fetchedConfigurations); // TODO: Add profile filtering if needed
    } catch (error) {
      console.error('Failed to fetch configurations:', error);
      throw error;
    }
  }, []);

  /**
   * Create a new service
   */
  const createService = useCallback(async (serviceData: Partial<Service>): Promise<Service> => {
    try {
      const newService = await ServiceApi.createService(serviceData);
      await fetchServices(); // Refresh services list
      return newService;
    } catch (error) {
      console.error('Failed to create service:', error);
      throw error;
    }
  }, [fetchServices]);

  /**
   * Update an existing service
   */
  const updateService = useCallback(async (serviceName: string, serviceData: Partial<Service>): Promise<Service> => {
    try {
      const updatedService = await ServiceApi.updateService(serviceName, serviceData);
      await fetchServices(); // Refresh services list
      return updatedService;
    } catch (error) {
      console.error('Failed to update service:', error);
      throw error;
    }
  }, [fetchServices]);

  /**
   * Delete a service
   */
  const deleteService = useCallback(async (serviceName: string): Promise<void> => {
    try {
      await ServiceApi.deleteService(serviceName);
      await fetchServices(); // Refresh services list
    } catch (error) {
      console.error('Failed to delete service:', error);
      throw error;
    }
  }, [fetchServices]);

  /**
   * Get service by name
   */
  const getServiceByName = useCallback((name: string): Service | undefined => {
    return services.find(service => service.name === name);
  }, [services]);

  /**
   * Refresh all data
   */
  const refreshData = useCallback(async (): Promise<void> => {
    setIsLoading(true);
    try {
      await Promise.all([
        fetchServices(),
        fetchConfigurations(),
      ]);
    } catch (error) {
      console.error('Failed to refresh data:', error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  }, [fetchServices, fetchConfigurations]);

  // Initial data loading
  useEffect(() => {
    refreshData();
  }, [refreshData]);

  // Update filtered services when profile changes
  useEffect(() => {
    setServices(filterServicesByProfile(allServices));
  }, [allServices, filterServicesByProfile]);

  return {
    // Data
    services,
    allServices,
    configurations,
    allConfigurations,
    isLoading,
    
    // Actions
    fetchServices,
    fetchConfigurations,
    createService,
    updateService,
    deleteService,
    
    // Utility
    getServiceByName,
    refreshData,
  };
}