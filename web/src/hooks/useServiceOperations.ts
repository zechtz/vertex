import { useState, useCallback } from 'react';
import { OperationsApi } from '@/services/operationsApi';
import { SystemApi } from '@/services/systemApi';
import { ServiceApi } from '@/services/serviceApi';
import { useAuth } from '@/contexts/AuthContext';
import { useToast, toast } from '@/components/ui/toast';
import { useConfirm, confirmDialogs } from '@/components/ui/confirm-dialog';
import { useProfile } from '@/contexts/ProfileContext';

export interface ServiceLoadingState {
  starting?: boolean;
  stopping?: boolean;
  restarting?: boolean;
  checkingHealth?: boolean;
  installingLibraries?: boolean;
}

export interface UseServiceOperationsReturn {
  // Loading states
  serviceLoadingStates: Record<string, ServiceLoadingState>;
  isStartingAll: boolean;
  isStoppingAll: boolean;
  isFixingLombok: boolean;
  isSyncingEnvironment: boolean;
  isCopyingLogs: boolean;
  isClearingLogs: boolean;
  
  // Individual service operations
  startService: (serviceName: string) => Promise<void>;
  stopService: (serviceName: string) => Promise<void>;
  restartService: (serviceName: string) => Promise<void>;
  checkServiceHealth: (serviceName: string) => Promise<void>;
  installServiceLibraries: (serviceName: string) => Promise<void>;
  copyServiceLogs: (serviceName: string) => Promise<void>;
  clearServiceLogs: (serviceName: string) => Promise<void>;
  
  // Bulk operations
  startAllServices: () => Promise<void>;
  stopAllServices: () => Promise<void>;
  
  // System operations
  fixLombok: () => Promise<void>;
  syncEnvironment: () => Promise<void>;
  copyAllLogs: () => Promise<void>;
  clearAllLogs: () => Promise<void>;
}

export function useServiceOperations(): UseServiceOperationsReturn {
  const { token } = useAuth();
  const { activeProfile } = useProfile();
  const { addToast } = useToast();
  const { showConfirm } = useConfirm();

  // Loading states
  const [serviceLoadingStates, setServiceLoadingStates] = useState<Record<string, ServiceLoadingState>>({});
  const [isStartingAll, setIsStartingAll] = useState(false);
  const [isStoppingAll, setIsStoppingAll] = useState(false);
  const [isFixingLombok, setIsFixingLombok] = useState(false);
  const [isSyncingEnvironment, setIsSyncingEnvironment] = useState(false);
  const [isCopyingLogs, setIsCopyingLogs] = useState(false);
  const [isClearingLogs, setIsClearingLogs] = useState(false);

  /**
   * Update loading state for a specific service operation
   */
  const updateServiceLoadingState = useCallback((serviceName: string, operation: keyof ServiceLoadingState, loading: boolean) => {
    setServiceLoadingStates(prev => ({
      ...prev,
      [serviceName]: {
        ...prev[serviceName],
        [operation]: loading,
      },
    }));
  }, []);

  /**
   * Start a service
   */
  const startService = useCallback(async (serviceName: string): Promise<void> => {
    updateServiceLoadingState(serviceName, 'starting', true);
    try {
      await OperationsApi.startService(serviceName);
      addToast(toast.success('Service started', `${serviceName} has been started successfully`));
    } catch (error) {
      console.error(`Failed to start service ${serviceName}:`, error);
      addToast(toast.error('Failed to start service', error instanceof Error ? error.message : 'Unknown error'));
      throw error;
    } finally {
      updateServiceLoadingState(serviceName, 'starting', false);
    }
  }, [updateServiceLoadingState, addToast]);

  /**
   * Stop a service
   */
  const stopService = useCallback(async (serviceName: string): Promise<void> => {
    updateServiceLoadingState(serviceName, 'stopping', true);
    try {
      await OperationsApi.stopService(serviceName);
      addToast(toast.success('Service stopped', `${serviceName} has been stopped successfully`));
    } catch (error) {
      console.error(`Failed to stop service ${serviceName}:`, error);
      addToast(toast.error('Failed to stop service', error instanceof Error ? error.message : 'Unknown error'));
      throw error;
    } finally {
      updateServiceLoadingState(serviceName, 'stopping', false);
    }
  }, [updateServiceLoadingState, addToast]);

  /**
   * Restart a service
   */
  const restartService = useCallback(async (serviceName: string): Promise<void> => {
    updateServiceLoadingState(serviceName, 'restarting', true);
    try {
      await OperationsApi.restartService(serviceName);
      addToast(toast.success('Service restarted', `${serviceName} has been restarted successfully`));
    } catch (error) {
      console.error(`Failed to restart service ${serviceName}:`, error);
      addToast(toast.error('Failed to restart service', error instanceof Error ? error.message : 'Unknown error'));
      throw error;
    } finally {
      updateServiceLoadingState(serviceName, 'restarting', false);
    }
  }, [updateServiceLoadingState, addToast]);

  /**
   * Check service health
   */
  const checkServiceHealth = useCallback(async (serviceName: string): Promise<void> => {
    updateServiceLoadingState(serviceName, 'checkingHealth', true);
    try {
      await OperationsApi.checkServiceHealth(serviceName);
      addToast(toast.success('Health check completed', `Health check for ${serviceName} completed`));
    } catch (error) {
      console.error(`Failed to check health for ${serviceName}:`, error);
      addToast(toast.error('Health check failed', error instanceof Error ? error.message : 'Unknown error'));
      throw error;
    } finally {
      updateServiceLoadingState(serviceName, 'checkingHealth', false);
    }
  }, [updateServiceLoadingState, addToast]);

  /**
   * Install service libraries
   */
  const installServiceLibraries = useCallback(async (serviceName: string): Promise<void> => {
    updateServiceLoadingState(serviceName, 'installingLibraries', true);
    try {
      await OperationsApi.installServiceLibraries(serviceName);
      addToast(toast.success('Libraries installed', `Libraries for ${serviceName} installed successfully`));
    } catch (error) {
      console.error(`Failed to install libraries for ${serviceName}:`, error);
      addToast(toast.error('Failed to install libraries', error instanceof Error ? error.message : 'Unknown error'));
      throw error;
    } finally {
      updateServiceLoadingState(serviceName, 'installingLibraries', false);
    }
  }, [updateServiceLoadingState, addToast]);

  /**
   * Copy service logs
   */
  const copyServiceLogs = useCallback(async (serviceName: string): Promise<void> => {
    try {
      await ServiceApi.copyServiceLogs(serviceName);
      addToast(toast.success('Logs copied', `Logs for ${serviceName} copied to clipboard`));
    } catch (error) {
      console.error(`Failed to copy logs for ${serviceName}:`, error);
      addToast(toast.error('Failed to copy logs', error instanceof Error ? error.message : 'Unknown error'));
      throw error;
    }
  }, [addToast]);

  /**
   * Clear service logs
   */
  const clearServiceLogs = useCallback(async (serviceName: string): Promise<void> => {
    try {
      await ServiceApi.clearServiceLogs(serviceName);
      addToast(toast.success('Logs cleared', `Logs for ${serviceName} cleared successfully`));
    } catch (error) {
      console.error(`Failed to clear logs for ${serviceName}:`, error);
      addToast(toast.error('Failed to clear logs', error instanceof Error ? error.message : 'Unknown error'));
      throw error;
    }
  }, [addToast]);

  /**
   * Start all services
   */
  const startAllServices = useCallback(async (): Promise<void> => {
    setIsStartingAll(true);
    try {
      const result = await OperationsApi.startAllServices(token);
      addToast(toast.success('Starting services', result.status || 'Services are being started'));
    } catch (error) {
      console.error('Failed to start all services:', error);
      addToast(toast.error('Failed to start all services', error instanceof Error ? error.message : 'Unknown error'));
      throw error;
    } finally {
      setIsStartingAll(false);
    }
  }, [token, addToast]);

  /**
   * Stop all services
   */
  const stopAllServices = useCallback(async (): Promise<void> => {
    const profileContext = activeProfile ? ` in profile "${activeProfile.name}"` : '';
    const confirmed = await showConfirm({
      ...confirmDialogs.stopAllServices(),
      description: `Are you sure you want to stop all services${profileContext}? This will stop all currently running services${profileContext}.`,
    });

    if (!confirmed) return;

    setIsStoppingAll(true);
    try {
      const result = await OperationsApi.stopAllServices(token);
      addToast(toast.success('Stopping services', result.status || 'Services are being stopped'));
    } catch (error) {
      console.error('Failed to stop all services:', error);
      addToast(toast.error('Failed to stop all services', error instanceof Error ? error.message : 'Unknown error'));
      throw error;
    } finally {
      setIsStoppingAll(false);
    }
  }, [token, activeProfile, showConfirm, addToast]);

  /**
   * Fix Lombok issues
   */
  const fixLombok = useCallback(async (): Promise<void> => {
    setIsFixingLombok(true);
    try {
      await SystemApi.fixLombok();
      addToast(toast.success('Lombok fixed', 'Lombok issues have been resolved'));
    } catch (error) {
      console.error('Failed to fix Lombok:', error);
      addToast(toast.error('Failed to fix Lombok', error instanceof Error ? error.message : 'Unknown error'));
      throw error;
    } finally {
      setIsFixingLombok(false);
    }
  }, [addToast]);

  /**
   * Sync environment
   */
  const syncEnvironment = useCallback(async (): Promise<void> => {
    setIsSyncingEnvironment(true);
    try {
      await SystemApi.syncEnvironment();
      addToast(toast.success('Environment synced', 'Environment setup completed successfully'));
    } catch (error) {
      console.error('Failed to sync environment:', error);
      addToast(toast.error('Failed to sync environment', error instanceof Error ? error.message : 'Unknown error'));
      throw error;
    } finally {
      setIsSyncingEnvironment(false);
    }
  }, [addToast]);

  /**
   * Copy all logs
   */
  const copyAllLogs = useCallback(async (): Promise<void> => {
    setIsCopyingLogs(true);
    try {
      await SystemApi.copyAllLogs();
      addToast(toast.success('All logs copied', 'All service logs copied to clipboard'));
    } catch (error) {
      console.error('Failed to copy all logs:', error);
      addToast(toast.error('Failed to copy logs', error instanceof Error ? error.message : 'Unknown error'));
      throw error;
    } finally {
      setIsCopyingLogs(false);
    }
  }, [addToast]);

  /**
   * Clear all logs
   */
  const clearAllLogs = useCallback(async (): Promise<void> => {
    const confirmed = await showConfirm({
      title: 'Clear All Logs',
      description: 'Are you sure you want to clear all service logs? This action cannot be undone.',
      confirmText: 'Clear All',
      variant: 'warning' as const,
    });

    if (!confirmed) return;

    setIsClearingLogs(true);
    try {
      await SystemApi.clearAllLogs();
      addToast(toast.success('All logs cleared', 'All service logs have been cleared'));
    } catch (error) {
      console.error('Failed to clear all logs:', error);
      addToast(toast.error('Failed to clear logs', error instanceof Error ? error.message : 'Unknown error'));
      throw error;
    } finally {
      setIsClearingLogs(false);
    }
  }, [showConfirm, addToast]);

  return {
    // Loading states
    serviceLoadingStates,
    isStartingAll,
    isStoppingAll,
    isFixingLombok,
    isSyncingEnvironment,
    isCopyingLogs,
    isClearingLogs,

    // Individual service operations
    startService,
    stopService,
    restartService,
    checkServiceHealth,
    installServiceLibraries,
    copyServiceLogs,
    clearServiceLogs,

    // Bulk operations
    startAllServices,
    stopAllServices,

    // System operations
    fixLombok,
    syncEnvironment,
    copyAllLogs,
    clearAllLogs,
  };
}