import * as React from "react"
import { cn } from "@/lib/utils"
import { X, CheckCircle, AlertCircle, Info, AlertTriangle } from "lucide-react"

const TOAST_LIMIT = 5

export type ToastVariant = "default" | "success" | "error" | "warning" | "info"

export interface Toast {
  id: string
  title?: string
  description?: string
  variant?: ToastVariant
  duration?: number
  action?: React.ReactNode
}

interface ToastContextType {
  toasts: Toast[]
  addToast: (toast: Omit<Toast, "id">) => void
  removeToast: (id: string) => void
  removeAllToasts: () => void
}

const ToastContext = React.createContext<ToastContextType | undefined>(undefined)

export function useToast() {
  const context = React.useContext(ToastContext)
  if (!context) {
    throw new Error("useToast must be used within a ToastProvider")
  }
  return context
}

let toastCount = 0

function generateId() {
  toastCount = (toastCount + 1) % Number.MAX_SAFE_INTEGER
  return toastCount.toString()
}

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = React.useState<Toast[]>([])

  const addToast = React.useCallback((toast: Omit<Toast, "id">) => {
    const id = generateId()
    const newToast: Toast = {
      ...toast,
      id,
      duration: toast.duration ?? 5000,
    }

    setToasts((currentToasts) => {
      const updatedToasts = [...currentToasts, newToast]
      if (updatedToasts.length > TOAST_LIMIT) {
        return updatedToasts.slice(-TOAST_LIMIT)
      }
      return updatedToasts
    })

    if (newToast.duration && newToast.duration > 0) {
      setTimeout(() => {
        removeToast(id)
      }, newToast.duration)
    }
  }, [])

  const removeToast = React.useCallback((id: string) => {
    setToasts((currentToasts) => 
      currentToasts.filter((toast) => toast.id !== id)
    )
  }, [])

  const removeAllToasts = React.useCallback(() => {
    setToasts([])
  }, [])

  return (
    <ToastContext.Provider
      value={{
        toasts,
        addToast,
        removeToast,
        removeAllToasts,
      }}
    >
      {children}
    </ToastContext.Provider>
  )
}

const toastVariants = {
  default: "bg-white border-gray-200 text-gray-900",
  success: "bg-green-50 border-green-200 text-green-900",
  error: "bg-red-50 border-red-200 text-red-900",
  warning: "bg-yellow-50 border-yellow-200 text-yellow-900",
  info: "bg-blue-50 border-blue-200 text-blue-900",
}

const iconVariants = {
  default: Info,
  success: CheckCircle,
  error: AlertCircle,
  warning: AlertTriangle,
  info: Info,
}

const iconColors = {
  default: "text-gray-500",
  success: "text-green-500",
  error: "text-red-500",
  warning: "text-yellow-500",
  info: "text-blue-500",
}

export function ToastContainer() {
  const { toasts, removeToast } = useToast()

  return (
    <div className="fixed top-4 right-4 z-50 flex flex-col gap-2 max-w-sm w-full">
      {toasts.map((toast) => {
        const Icon = iconVariants[toast.variant || "default"]
        return (
          <div
            key={toast.id}
            className={cn(
              "flex items-start gap-3 p-4 rounded-lg border shadow-lg animate-in slide-in-from-right-full",
              toastVariants[toast.variant || "default"]
            )}
          >
            <Icon className={cn("h-5 w-5 mt-0.5 flex-shrink-0", iconColors[toast.variant || "default"])} />
            <div className="flex-1 min-w-0">
              {toast.title && (
                <div className="font-medium text-sm mb-1">{toast.title}</div>
              )}
              {toast.description && (
                <div className="text-sm opacity-90">{toast.description}</div>
              )}
              {toast.action && (
                <div className="mt-2">{toast.action}</div>
              )}
            </div>
            <button
              onClick={() => removeToast(toast.id)}
              className="flex-shrink-0 text-gray-400 hover:text-gray-600 transition-colors"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        )
      })}
    </div>
  )
}

// Convenience functions
export const toast = {
  success: (title: string, description?: string) => ({
    title,
    description,
    variant: "success" as const,
  }),
  error: (title: string, description?: string) => ({
    title,
    description,
    variant: "error" as const,
  }),
  warning: (title: string, description?: string) => ({
    title,
    description,
    variant: "warning" as const,
  }),
  info: (title: string, description?: string) => ({
    title,
    description,
    variant: "info" as const,
  }),
}