import React, { createContext, useContext } from "react";
import { useModalManager } from "@/hooks/useModalManager";

interface ModalContextType {
  modalManager: ReturnType<typeof useModalManager>;
}

const ModalContext = createContext<ModalContextType | undefined>(undefined);

export function useModalContext() {
  const context = useContext(ModalContext);
  if (context === undefined) {
    throw new Error("useModalContext must be used within a ModalProvider");
  }
  return context;
}

interface ModalProviderProps {
  children: React.ReactNode;
  modals?: string[];
}

export function ModalProvider({ children, modals = [] }: ModalProviderProps) {
  const modalManager = useModalManager(modals);

  return (
    <ModalContext.Provider value={{ modalManager }}>
      {children}
    </ModalContext.Provider>
  );
}

// Higher-order component for components that need modal management
export function withModalManager<P extends object>(
  Component: React.ComponentType<P>,
  requiredModals: string[] = [],
) {
  return function WrappedComponent(props: P) {
    return (
      <ModalProvider modals={requiredModals}>
        <Component {...props} />
      </ModalProvider>
    );
  };
}
