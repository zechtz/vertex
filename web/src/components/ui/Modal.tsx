import React, { useEffect, useId, useRef } from "react";
import { X } from "lucide-react";
import { Button } from "./button";

export interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  children: React.ReactNode;
  size?: "sm" | "md" | "lg" | "xl" | "2xl" | "full";
  showCloseButton?: boolean;
  closeOnOverlayClick?: boolean;
  closeOnEscape?: boolean;
  className?: string;
  headerClassName?: string;
  contentClassName?: string;
}

/**
 * What Tab is allowed to reach. Used both to place focus on open and to keep
 * it from escaping to the page behind, which is inert to the eye but was not
 * inert to the keyboard or to a screen reader.
 */
const FOCUSABLE_SELECTOR = [
  "a[href]",
  "button:not([disabled])",
  "textarea:not([disabled])",
  "input:not([disabled])",
  "select:not([disabled])",
  '[tabindex]:not([tabindex="-1"])',
].join(", ");

const sizeClasses = {
  sm: "max-w-sm",
  md: "max-w-md",
  lg: "max-w-lg",
  xl: "max-w-xl",
  "2xl": "max-w-2xl",
  full: "max-w-full mx-4",
};

export function Modal({
  isOpen,
  onClose,
  title,
  children,
  size = "md",
  showCloseButton = true,
  closeOnOverlayClick = true,
  closeOnEscape = true,
  className = "",
  headerClassName = "",
  contentClassName = "",
}: ModalProps) {
  const titleId = useId();
  const panelRef = useRef<HTMLDivElement>(null);

  // Handle escape key
  useEffect(() => {
    if (!closeOnEscape || !isOpen) return;

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    document.addEventListener("keydown", handleEscape);
    return () => document.removeEventListener("keydown", handleEscape);
  }, [isOpen, onClose, closeOnEscape]);

  // Prevent body scroll when modal is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "unset";
    }

    return () => {
      document.body.style.overflow = "unset";
    };
  }, [isOpen]);

  // Move focus into the dialog when it opens and hand it back to whatever
  // opened it on close, so keyboard users are not dropped at the top of the
  // page with no idea where they are.
  useEffect(() => {
    if (!isOpen) return;

    const previouslyFocused = document.activeElement as HTMLElement | null;

    // The panel itself rather than its first control: the first control is the
    // close button, and "Close, button" is a worse greeting than the dialog's
    // own title, which aria-labelledby reads out from here.
    panelRef.current?.focus();

    return () => previouslyFocused?.focus?.();
  }, [isOpen]);

  if (!isOpen) return null;

  // Tab cycles within the dialog rather than walking out behind the backdrop.
  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (event.key !== "Tab") return;

    const panel = panelRef.current;
    if (!panel) return;

    const focusable = Array.from(
      panel.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR),
    ).filter((element) => element.offsetParent !== null);

    if (focusable.length === 0) {
      event.preventDefault();
      return;
    }

    const first = focusable[0];
    const last = focusable[focusable.length - 1];

    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  };

  const handleOverlayClick = (e: React.MouseEvent) => {
    if (closeOnOverlayClick && e.target === e.currentTarget) {
      onClose();
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div 
        className="flex min-h-screen items-center justify-center p-4"
        onClick={handleOverlayClick}
      >
        {/* Backdrop */}
        <div
          className="fixed inset-0 bg-black/50 transition-opacity"
          aria-hidden="true"
        />
        
        {/* Modal */}
        <div
          ref={panelRef}
          role="dialog"
          aria-modal="true"
          aria-labelledby={title ? titleId : undefined}
          tabIndex={-1}
          onKeyDown={handleKeyDown}
          className={`relative w-full ${sizeClasses[size]} max-h-[95vh] overflow-hidden outline-none ${className}`}
          onClick={(e) => e.stopPropagation()}
        >
          <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl flex flex-col max-h-[95vh]">
            {/* Header */}
            {(title || showCloseButton) && (
              <div className={`flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700 flex-shrink-0 ${headerClassName}`}>
                {title && (
                  <h2
                    id={titleId}
                    className="text-xl font-semibold text-gray-900 dark:text-gray-100"
                  >
                    {title}
                  </h2>
                )}
                {showCloseButton && (
                  <Button variant="ghost" size="sm" onClick={onClose}>
                    <X className="w-5 h-5" />
                  </Button>
                )}
              </div>
            )}

            {/* Content */}
            <div className={`flex-1 overflow-y-auto ${contentClassName}`}>
              {children}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Modal;