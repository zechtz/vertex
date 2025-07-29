import { useState, useCallback } from 'react';
import { Service } from '@/types';

export interface UseModalsReturn {
  // Service Config Modal
  showServiceConfig: boolean;
  isCreatingService: boolean;
  editingService: Service | null;
  openServiceConfig: (service?: Service) => void;
  closeServiceConfig: () => void;
  
  // Service Files Modal
  showServiceFiles: boolean;
  viewingFilesService: Service | null;
  openServiceFiles: (service: Service) => void;
  closeServiceFiles: () => void;
  
  // Service Environment Modal
  showServiceEnv: boolean;
  envEditingService: Service | null;
  openServiceEnv: (service: Service) => void;
  closeServiceEnv: () => void;
  
  // Service Action Modal
  showServiceActionModal: boolean;
  serviceToAction: Service | null;
  openServiceActionModal: (service: Service) => void;
  closeServiceActionModal: () => void;
  
  // Global modals (assuming these are needed based on the imports in App.tsx)
  showGlobalEnvModal: boolean;
  openGlobalEnvModal: () => void;
  closeGlobalEnvModal: () => void;
  
  showGlobalConfigModal: boolean;
  openGlobalConfigModal: () => void;
  closeGlobalConfigModal: () => void;
  
  showSystemMetricsModal: boolean;
  openSystemMetricsModal: () => void;
  closeSystemMetricsModal: () => void;
  
  showLogAggregationModal: boolean;
  openLogAggregationModal: () => void;
  closeLogAggregationModal: () => void;
  
  showServiceTopologyModal: boolean;
  openServiceTopologyModal: () => void;
  closeServiceTopologyModal: () => void;
  
  showDependencyConfigModal: boolean;
  openDependencyConfigModal: () => void;
  closeDependencyConfigModal: () => void;
  
  showAutoDiscoveryModal: boolean;
  openAutoDiscoveryModal: () => void;
  closeAutoDiscoveryModal: () => void;
  
  showProfileManagement: boolean;
  openProfileManagement: () => void;
  closeProfileManagement: () => void;
  
  showProfileConfigDashboard: boolean;
  openProfileConfigDashboard: () => void;
  closeProfileConfigDashboard: () => void;
}

export function useModals(): UseModalsReturn {
  // Service Config Modal
  const [showServiceConfig, setShowServiceConfig] = useState(false);
  const [isCreatingService, setIsCreatingService] = useState(false);
  const [editingService, setEditingService] = useState<Service | null>(null);
  
  // Service Files Modal
  const [showServiceFiles, setShowServiceFiles] = useState(false);
  const [viewingFilesService, setViewingFilesService] = useState<Service | null>(null);
  
  // Service Environment Modal
  const [showServiceEnv, setShowServiceEnv] = useState(false);
  const [envEditingService, setEnvEditingService] = useState<Service | null>(null);
  
  // Service Action Modal
  const [showServiceActionModal, setShowServiceActionModal] = useState(false);
  const [serviceToAction, setServiceToAction] = useState<Service | null>(null);
  
  // Global modals
  const [showGlobalEnvModal, setShowGlobalEnvModal] = useState(false);
  const [showGlobalConfigModal, setShowGlobalConfigModal] = useState(false);
  const [showSystemMetricsModal, setShowSystemMetricsModal] = useState(false);
  const [showLogAggregationModal, setShowLogAggregationModal] = useState(false);
  const [showServiceTopologyModal, setShowServiceTopologyModal] = useState(false);
  const [showDependencyConfigModal, setShowDependencyConfigModal] = useState(false);
  const [showAutoDiscoveryModal, setShowAutoDiscoveryModal] = useState(false);
  const [showProfileManagement, setShowProfileManagement] = useState(false);
  const [showProfileConfigDashboard, setShowProfileConfigDashboard] = useState(false);

  // Service Config Modal handlers
  const openServiceConfig = useCallback((service?: Service) => {
    if (service) {
      setEditingService(service);
      setIsCreatingService(false);
    } else {
      setEditingService(null);
      setIsCreatingService(true);
    }
    setShowServiceConfig(true);
  }, []);

  const closeServiceConfig = useCallback(() => {
    setShowServiceConfig(false);
    setEditingService(null);
    setIsCreatingService(false);
  }, []);

  // Service Files Modal handlers
  const openServiceFiles = useCallback((service: Service) => {
    setViewingFilesService(service);
    setShowServiceFiles(true);
  }, []);

  const closeServiceFiles = useCallback(() => {
    setShowServiceFiles(false);
    setViewingFilesService(null);
  }, []);

  // Service Environment Modal handlers
  const openServiceEnv = useCallback((service: Service) => {
    setEnvEditingService(service);
    setShowServiceEnv(true);
  }, []);

  const closeServiceEnv = useCallback(() => {
    setShowServiceEnv(false);
    setEnvEditingService(null);
  }, []);

  // Service Action Modal handlers
  const openServiceActionModal = useCallback((service: Service) => {
    setServiceToAction(service);
    setShowServiceActionModal(true);
  }, []);

  const closeServiceActionModal = useCallback(() => {
    setShowServiceActionModal(false);
    setServiceToAction(null);
  }, []);

  // Global modal handlers
  const openGlobalEnvModal = useCallback(() => setShowGlobalEnvModal(true), []);
  const closeGlobalEnvModal = useCallback(() => setShowGlobalEnvModal(false), []);
  
  const openGlobalConfigModal = useCallback(() => setShowGlobalConfigModal(true), []);
  const closeGlobalConfigModal = useCallback(() => setShowGlobalConfigModal(false), []);
  
  const openSystemMetricsModal = useCallback(() => setShowSystemMetricsModal(true), []);
  const closeSystemMetricsModal = useCallback(() => setShowSystemMetricsModal(false), []);
  
  const openLogAggregationModal = useCallback(() => setShowLogAggregationModal(true), []);
  const closeLogAggregationModal = useCallback(() => setShowLogAggregationModal(false), []);
  
  const openServiceTopologyModal = useCallback(() => setShowServiceTopologyModal(true), []);
  const closeServiceTopologyModal = useCallback(() => setShowServiceTopologyModal(false), []);
  
  const openDependencyConfigModal = useCallback(() => setShowDependencyConfigModal(true), []);
  const closeDependencyConfigModal = useCallback(() => setShowDependencyConfigModal(false), []);
  
  const openAutoDiscoveryModal = useCallback(() => setShowAutoDiscoveryModal(true), []);
  const closeAutoDiscoveryModal = useCallback(() => setShowAutoDiscoveryModal(false), []);
  
  const openProfileManagement = useCallback(() => setShowProfileManagement(true), []);
  const closeProfileManagement = useCallback(() => setShowProfileManagement(false), []);
  
  const openProfileConfigDashboard = useCallback(() => setShowProfileConfigDashboard(true), []);
  const closeProfileConfigDashboard = useCallback(() => setShowProfileConfigDashboard(false), []);

  return {
    // Service Config Modal
    showServiceConfig,
    isCreatingService,
    editingService,
    openServiceConfig,
    closeServiceConfig,
    
    // Service Files Modal
    showServiceFiles,
    viewingFilesService,
    openServiceFiles,
    closeServiceFiles,
    
    // Service Environment Modal
    showServiceEnv,
    envEditingService,
    openServiceEnv,
    closeServiceEnv,
    
    // Service Action Modal
    showServiceActionModal,
    serviceToAction,
    openServiceActionModal,
    closeServiceActionModal,
    
    // Global modals
    showGlobalEnvModal,
    openGlobalEnvModal,
    closeGlobalEnvModal,
    
    showGlobalConfigModal,
    openGlobalConfigModal,
    closeGlobalConfigModal,
    
    showSystemMetricsModal,
    openSystemMetricsModal,
    closeSystemMetricsModal,
    
    showLogAggregationModal,
    openLogAggregationModal,
    closeLogAggregationModal,
    
    showServiceTopologyModal,
    openServiceTopologyModal,
    closeServiceTopologyModal,
    
    showDependencyConfigModal,
    openDependencyConfigModal,
    closeDependencyConfigModal,
    
    showAutoDiscoveryModal,
    openAutoDiscoveryModal,
    closeAutoDiscoveryModal,
    
    showProfileManagement,
    openProfileManagement,
    closeProfileManagement,
    
    showProfileConfigDashboard,
    openProfileConfigDashboard,
    closeProfileConfigDashboard,
  };
}