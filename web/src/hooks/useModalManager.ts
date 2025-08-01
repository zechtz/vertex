import { useState, useCallback } from "react";

export interface ModalConfig<T = any> {
  isOpen: boolean;
  data?: T;
}

export type ModalState = Record<string, ModalConfig>;

export function useModalManager(initialModals: string[] = []) {
  const [modals, setModals] = useState<ModalState>(() => {
    return initialModals.reduce((acc, modalName) => {
      acc[modalName] = { isOpen: false, data: undefined };
      return acc;
    }, {} as ModalState);
  });

  const openModal = useCallback(<T,>(modalName: string, data?: T) => {
    setModals((prev) => ({
      ...prev,
      [modalName]: {
        isOpen: true,
        data,
      },
    }));
  }, []);

  const closeModal = useCallback((modalName: string) => {
    setModals((prev) => ({
      ...prev,
      [modalName]: {
        isOpen: false,
        data: undefined,
      },
    }));
  }, []);

  const toggleModal = useCallback(<T,>(modalName: string, data?: T) => {
    setModals((prev) => {
      const currentModal = prev[modalName] || { isOpen: false, data: undefined };
      return {
        ...prev,
        [modalName]: {
          isOpen: !currentModal.isOpen,
          data: currentModal.isOpen ? undefined : data,
        },
      };
    });
  }, []);

  const closeAllModals = useCallback(() => {
    setModals((prev) => {
      const updated = { ...prev };
      Object.keys(updated).forEach((key) => {
        updated[key] = { isOpen: false, data: undefined };
      });
      return updated;
    });
  }, []);

  const isModalOpen = useCallback((modalName: string): boolean => {
    return modals[modalName]?.isOpen || false;
  }, [modals]);

  const getModalData = useCallback(<T,>(modalName: string): T | undefined => {
    return modals[modalName]?.data as T;
  }, [modals]);

  const getModal = useCallback((modalName: string) => {
    return modals[modalName] || { isOpen: false, data: undefined };
  }, [modals]);

  return {
    modals,
    openModal,
    closeModal,
    toggleModal,
    closeAllModals,
    isModalOpen,
    getModalData,
    getModal,
  };
}