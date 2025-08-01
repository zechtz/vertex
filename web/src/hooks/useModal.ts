import { useState, useCallback } from "react";

export interface ModalState {
  isOpen: boolean;
  data?: any;
}

export function useModal<T = any>(initialState: boolean = false) {
  const [state, setState] = useState<ModalState>({
    isOpen: initialState,
    data: undefined,
  });

  const open = useCallback((data?: T) => {
    setState({
      isOpen: true,
      data,
    });
  }, []);

  const close = useCallback(() => {
    setState({
      isOpen: false,
      data: undefined,
    });
  }, []);

  const toggle = useCallback((data?: T) => {
    setState((prev) => ({
      isOpen: !prev.isOpen,
      data: prev.isOpen ? undefined : data,
    }));
  }, []);

  return {
    isOpen: state.isOpen,
    data: state.data as T,
    open,
    close,
    toggle,
  };
}