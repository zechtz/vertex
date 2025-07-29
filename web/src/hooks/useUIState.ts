import { useState, useCallback } from 'react';
import { Service } from '@/types';

export interface UseUIStateReturn {
  // Navigation and layout
  activeSection: string;
  setActiveSection: (section: string) => void;
  isSidebarCollapsed: boolean;
  toggleSidebar: () => void;
  
  // Search and filtering
  searchTerm: string;
  setSearchTerm: (term: string) => void;
  
  // Selection state
  selectedService: Service | null;
  setSelectedService: (service: Service | null) => void;
  
  // UI feedback states
  copied: boolean;
  setCopied: (copied: boolean) => void;
  showCopyFeedback: () => void;
  
  // Save state
  isSavingService: boolean;
  setIsSavingService: (saving: boolean) => void;
}

export function useUIState(): UseUIStateReturn {
  // Navigation and layout
  const [activeSection, setActiveSection] = useState('services');
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  
  // Search and filtering
  const [searchTerm, setSearchTerm] = useState('');
  
  // Selection state
  const [selectedService, setSelectedService] = useState<Service | null>(null);
  
  // UI feedback states
  const [copied, setCopied] = useState(false);
  const [isSavingService, setIsSavingService] = useState(false);

  /**
   * Toggle sidebar collapsed state
   */
  const toggleSidebar = useCallback(() => {
    setIsSidebarCollapsed(prev => !prev);
  }, []);

  /**
   * Show copy feedback with auto-hide
   */
  const showCopyFeedback = useCallback(() => {
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, []);

  return {
    // Navigation and layout
    activeSection,
    setActiveSection,
    isSidebarCollapsed,
    toggleSidebar,
    
    // Search and filtering
    searchTerm,
    setSearchTerm,
    
    // Selection state
    selectedService,
    setSelectedService,
    
    // UI feedback states
    copied,
    setCopied,
    showCopyFeedback,
    
    // Save state
    isSavingService,
    setIsSavingService,
  };
}