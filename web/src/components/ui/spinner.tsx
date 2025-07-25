import * as React from "react"
import { cn } from "@/lib/utils"

export interface SpinnerProps extends React.HTMLAttributes<HTMLDivElement> {
  size?: "sm" | "md" | "lg"
  variant?: "default" | "primary" | "white"
}

const sizeVariants = {
  sm: "w-4 h-4 border-2",
  md: "w-6 h-6 border-2",
  lg: "w-8 h-8 border-3",
}

const variantStyles = {
  default: "border-gray-300 border-t-gray-600",
  primary: "border-blue-200 border-t-blue-600",
  white: "border-white/30 border-t-white",
}

export function Spinner({ 
  className, 
  size = "md", 
  variant = "default",
  ...props 
}: SpinnerProps) {
  return (
    <div
      className={cn(
        "rounded-full animate-spin",
        sizeVariants[size],
        variantStyles[variant],
        className
      )}
      {...props}
    />
  )
}

export interface LoadingProps {
  text?: string
  size?: "sm" | "md" | "lg"
  variant?: "default" | "primary" | "white"
  className?: string
}

export function Loading({ 
  text = "Loading...", 
  size = "md", 
  variant = "default",
  className 
}: LoadingProps) {
  return (
    <div className={cn("flex items-center gap-3", className)}>
      <Spinner size={size} variant={variant} />
      <span className={cn(
        "text-sm",
        variant === "white" ? "text-white" : "text-gray-600"
      )}>
        {text}
      </span>
    </div>
  )
}

export interface LoadingOverlayProps {
  isLoading: boolean
  text?: string
  children: React.ReactNode
  className?: string
}

export function LoadingOverlay({ 
  isLoading, 
  text = "Loading...", 
  children, 
  className 
}: LoadingOverlayProps) {
  return (
    <div className={cn("relative", className)}>
      {children}
      {isLoading && (
        <div className="absolute inset-0 bg-white/80 backdrop-blur-sm flex items-center justify-center z-10 rounded-md">
          <Loading text={text} variant="primary" />
        </div>
      )}
    </div>
  )
}

export interface ButtonSpinnerProps {
  isLoading: boolean
  children: React.ReactNode
  loadingText?: string
}

export function ButtonSpinner({ 
  isLoading, 
  children, 
  loadingText 
}: ButtonSpinnerProps) {
  if (isLoading) {
    return (
      <div className="flex items-center gap-2">
        <Spinner size="sm" variant="white" />
        {loadingText || "Loading..."}
      </div>
    )
  }
  
  return <>{children}</>
}