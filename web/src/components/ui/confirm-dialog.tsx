import * as React from "react"
import { AlertTriangle, Trash2, Power, RotateCcw } from "lucide-react"
import { Button } from "./button"
import { cn } from "@/lib/utils"

export type ConfirmVariant = "default" | "destructive" | "warning"

export interface ConfirmDialogProps {
  isOpen: boolean
  onClose: () => void
  onConfirm: () => void
  title: string
  description: string
  confirmText?: string
  cancelText?: string
  variant?: ConfirmVariant
  isLoading?: boolean
  icon?: React.ReactNode
}

interface ConfirmDialogContextType {
  showConfirm: (options: Omit<ConfirmDialogProps, "isOpen" | "onClose" | "onConfirm">) => Promise<boolean>
}

const ConfirmDialogContext = React.createContext<ConfirmDialogContextType | undefined>(undefined)

export function useConfirm() {
  const context = React.useContext(ConfirmDialogContext)
  if (!context) {
    throw new Error("useConfirm must be used within a ConfirmDialogProvider")
  }
  return context
}

export function ConfirmDialogProvider({ children }: { children: React.ReactNode }) {
  const [dialogState, setDialogState] = React.useState<{
    isOpen: boolean
    resolve?: (value: boolean) => void
    props?: Omit<ConfirmDialogProps, "isOpen" | "onClose" | "onConfirm">
  }>({ isOpen: false })

  const showConfirm = React.useCallback((options: Omit<ConfirmDialogProps, "isOpen" | "onClose" | "onConfirm">) => {
    return new Promise<boolean>((resolve) => {
      setDialogState({
        isOpen: true,
        resolve,
        props: options,
      })
    })
  }, [])

  const handleClose = React.useCallback(() => {
    if (dialogState.resolve) {
      dialogState.resolve(false)
    }
    setDialogState({ isOpen: false })
  }, [dialogState.resolve])

  const handleConfirm = React.useCallback(() => {
    if (dialogState.resolve) {
      dialogState.resolve(true)
    }
    setDialogState({ isOpen: false })
  }, [dialogState.resolve])

  return (
    <ConfirmDialogContext.Provider value={{ showConfirm }}>
      {children}
      {dialogState.isOpen && dialogState.props && (
        <ConfirmDialog
          {...dialogState.props}
          isOpen={dialogState.isOpen}
          onClose={handleClose}
          onConfirm={handleConfirm}
        />
      )}
    </ConfirmDialogContext.Provider>
  )
}

const variantStyles = {
  default: {
    backdrop: "bg-blue-50",
    icon: "text-blue-600",
    confirmButton: "bg-blue-600 hover:bg-blue-700",
  },
  destructive: {
    backdrop: "bg-red-50",
    icon: "text-red-600",
    confirmButton: "bg-red-600 hover:bg-red-700",
  },
  warning: {
    backdrop: "bg-yellow-50",
    icon: "text-yellow-600",
    confirmButton: "bg-yellow-600 hover:bg-yellow-700",
  },
}

export function ConfirmDialog({
  isOpen,
  onClose,
  onConfirm,
  title,
  description,
  confirmText = "Confirm",
  cancelText = "Cancel",
  variant = "default",
  isLoading = false,
  icon,
}: ConfirmDialogProps) {
  const styles = variantStyles[variant]

  if (!isOpen) return null

  const defaultIcon = variant === "destructive" ? <Trash2 className="h-6 w-6" /> : <AlertTriangle className="h-6 w-6" />

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
      />
      
      {/* Dialog */}
      <div className="relative bg-white rounded-xl shadow-2xl max-w-md w-full mx-4 animate-in zoom-in-95 duration-200">
        <div className="p-6">
          {/* Icon and Title */}
          <div className="flex items-start gap-4 mb-4">
            <div className={cn("p-3 rounded-full", styles.backdrop)}>
              <div className={cn(styles.icon)}>
                {icon || defaultIcon}
              </div>
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">
                {title}
              </h3>
              <p className="text-sm text-gray-600 leading-relaxed">
                {description}
              </p>
            </div>
          </div>
          
          {/* Actions */}
          <div className="flex gap-3 justify-end mt-6">
            <Button
              variant="outline"
              onClick={onClose}
              disabled={isLoading}
            >
              {cancelText}
            </Button>
            <Button
              onClick={onConfirm}
              disabled={isLoading}
              className={cn("text-white", styles.confirmButton)}
            >
              {isLoading ? (
                <div className="flex items-center gap-2">
                  <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  Processing...
                </div>
              ) : (
                confirmText
              )}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

// Convenience functions for common confirm dialogs
export const confirmDialogs = {
  delete: (itemName: string) => ({
    title: "Delete Item",
    description: `Are you sure you want to delete "${itemName}"? This action cannot be undone.`,
    confirmText: "Delete",
    variant: "destructive" as const,
    icon: <Trash2 className="h-6 w-6" />,
  }),
  
  deleteService: (serviceName: string) => ({
    title: "Delete Service",
    description: `Are you sure you want to delete the service "${serviceName}"? This will remove all configuration and cannot be undone.`,
    confirmText: "Delete Service",
    variant: "destructive" as const,
    icon: <Trash2 className="h-6 w-6" />,
  }),
  
  stopService: (serviceName: string) => ({
    title: "Stop Service",
    description: `Are you sure you want to stop the service "${serviceName}"?`,
    confirmText: "Stop Service",
    variant: "warning" as const,
    icon: <Power className="h-6 w-6" />,
  }),
  
  restartService: (serviceName: string) => ({
    title: "Restart Service",
    description: `Are you sure you want to restart the service "${serviceName}"? This will temporarily interrupt the service.`,
    confirmText: "Restart Service",
    variant: "warning" as const,
    icon: <RotateCcw className="h-6 w-6" />,
  }),
  
  stopAllServices: () => ({
    title: "Stop All Services",
    description: "Are you sure you want to stop all services? This will stop all currently running services.",
    confirmText: "Stop All",
    variant: "warning" as const,
    icon: <Power className="h-6 w-6" />,
  }),
  
  clearLogs: (serviceName: string) => ({
    title: "Clear Logs",
    description: `Are you sure you want to clear all logs for "${serviceName}"? This action cannot be undone.`,
    confirmText: "Clear Logs",
    variant: "warning" as const,
    icon: <Trash2 className="h-6 w-6" />,
  }),
}