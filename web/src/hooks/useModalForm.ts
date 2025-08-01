import { useState, useCallback, useEffect } from "react";
import { useModal } from "./useModal";

export interface ModalFormConfig<T> {
  initialData?: T;
  resetOnClose?: boolean;
  resetOnOpen?: boolean;
}

export function useModalForm<T>(config: ModalFormConfig<T> = {}) {
  const modal = useModal<T>();
  const [formData, setFormData] = useState<T | undefined>(config.initialData);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Reset form when modal opens/closes
  useEffect(() => {
    if (modal.isOpen && config.resetOnOpen) {
      setFormData(modal.data || config.initialData);
      setErrors({});
    }
  }, [modal.isOpen, modal.data, config.resetOnOpen, config.initialData]);

  const openModal = useCallback((data?: T) => {
    if (config.resetOnOpen) {
      setFormData(data || config.initialData);
      setErrors({});
    }
    modal.open(data);
  }, [modal, config.resetOnOpen, config.initialData]);

  const closeModal = useCallback(() => {
    if (config.resetOnClose) {
      setFormData(config.initialData);
      setErrors({});
    }
    setIsSubmitting(false);
    modal.close();
  }, [modal, config.resetOnClose, config.initialData]);

  const handleSubmit = useCallback(async <R,>(
    submitFn: (data: T) => Promise<R>,
    onSuccess?: (result: R) => void,
    onError?: (error: Error) => void
  ) => {
    if (!formData) return;

    try {
      setIsSubmitting(true);
      setErrors({});
      
      const result = await submitFn(formData);
      
      if (onSuccess) {
        onSuccess(result);
      }
      
      closeModal();
    } catch (error) {
      console.error("Form submission error:", error);
      
      if (error instanceof Error) {
        if (onError) {
          onError(error);
        } else {
          // Default error handling - you could customize this
          setErrors({ general: error.message });
        }
      }
    } finally {
      setIsSubmitting(false);
    }
  }, [formData, closeModal]);

  const setFieldValue = useCallback(<K extends keyof T>(
    field: K,
    value: T[K]
  ) => {
    setFormData(prev => prev ? { ...prev, [field]: value } : prev);
    
    // Clear field error when user starts typing
    if (errors[field as string]) {
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[field as string];
        return newErrors;
      });
    }
  }, [errors]);

  const setFieldError = useCallback((field: string, error: string) => {
    setErrors(prev => ({ ...prev, [field]: error }));
  }, []);

  const clearErrors = useCallback(() => {
    setErrors({});
  }, []);

  return {
    // Modal state
    isOpen: modal.isOpen,
    modalData: modal.data,
    
    // Form state
    formData,
    setFormData,
    isSubmitting,
    errors,
    
    // Actions
    openModal,
    closeModal,
    handleSubmit,
    setFieldValue,
    setFieldError,
    clearErrors,
    
    // Utilities
    hasErrors: Object.keys(errors).length > 0,
    isValid: formData !== undefined && Object.keys(errors).length === 0,
  };
}