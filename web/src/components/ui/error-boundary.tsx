import * as React from "react"
import { AlertTriangle, RefreshCcw, Home } from "lucide-react"
import { Button } from "./button"

interface ErrorInfo {
  componentStack: string
}

interface ErrorBoundaryState {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
}

interface ErrorBoundaryProps {
  children: React.ReactNode
  fallback?: React.ComponentType<ErrorFallbackProps>
  onError?: (error: Error, errorInfo: ErrorInfo) => void
}

interface ErrorFallbackProps {
  error: Error | null
  errorInfo: ErrorInfo | null
  onRetry: () => void
  onHome: () => void
}

function DefaultErrorFallback({ error, errorInfo, onRetry, onHome }: ErrorFallbackProps) {
  const isDevelopment = import.meta.env?.DEV ?? false

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-xl shadow-lg p-8 text-center">
        <div className="mb-6">
          <div className="mx-auto w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mb-4">
            <AlertTriangle className="h-8 w-8 text-red-600" />
          </div>
          <h1 className="text-xl font-semibold text-gray-900 mb-2">
            Something went wrong
          </h1>
          <p className="text-gray-600 text-sm">
            We're sorry, but something unexpected happened. Please try refreshing the page.
          </p>
        </div>

        {isDevelopment && error && (
          <div className="mb-6 p-4 bg-red-50 rounded-lg text-left">
            <h3 className="font-medium text-red-800 mb-2">Error Details:</h3>
            <pre className="text-xs text-red-700 whitespace-pre-wrap overflow-auto max-h-32">
              {error.message}
            </pre>
            {errorInfo && (
              <details className="mt-2">
                <summary className="text-xs text-red-600 cursor-pointer">
                  Component Stack
                </summary>
                <pre className="text-xs text-red-600 whitespace-pre-wrap mt-1 overflow-auto max-h-20">
                  {errorInfo.componentStack}
                </pre>
              </details>
            )}
          </div>
        )}

        <div className="flex gap-3 justify-center">
          <Button
            variant="outline"
            onClick={onHome}
            className="flex items-center gap-2"
          >
            <Home className="h-4 w-4" />
            Go Home
          </Button>
          <Button
            onClick={onRetry}
            className="flex items-center gap-2"
          >
            <RefreshCcw className="h-4 w-4" />
            Try Again
          </Button>
        </div>
      </div>
    </div>
  )
}

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    }
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return {
      hasError: true,
      error,
    }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({
      error,
      errorInfo,
    })

    // Log error to external service in production
    if (this.props.onError) {
      this.props.onError(error, errorInfo)
    } else {
      console.error('Error caught by boundary:', error, errorInfo)
    }
  }

  handleRetry = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    })
  }

  handleHome = () => {
    window.location.href = '/'
  }

  render() {
    if (this.state.hasError) {
      const FallbackComponent = this.props.fallback || DefaultErrorFallback
      
      return (
        <FallbackComponent
          error={this.state.error}
          errorInfo={this.state.errorInfo}
          onRetry={this.handleRetry}
          onHome={this.handleHome}
        />
      )
    }

    return this.props.children
  }
}

// Hook for functional components to handle errors
export function useErrorHandler() {
  return React.useCallback((error: Error, errorInfo?: any) => {
    console.error('Error caught by error handler:', error, errorInfo)
    
    // In a real app, you might want to send this to an error reporting service
    if (import.meta.env?.PROD) {
      // Send to error reporting service
      // reportError(error, errorInfo)
    }
  }, [])
}

// Higher-order component for wrapping components with error boundary
export function withErrorBoundary<P extends object>(
  Component: React.ComponentType<P>,
  errorBoundaryProps?: Omit<ErrorBoundaryProps, 'children'>
) {
  const WrappedComponent = (props: P) => (
    <ErrorBoundary {...errorBoundaryProps}>
      <Component {...props} />
    </ErrorBoundary>
  )

  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`
  
  return WrappedComponent
}

// Component-level error boundary for smaller sections
export function ErrorBoundarySection({ 
  children, 
  title = "Section Error",
  description = "This section encountered an error.",
  onRetry,
}: {
  children: React.ReactNode
  title?: string
  description?: string
  onRetry?: () => void
}) {
  return (
    <ErrorBoundary
      fallback={({ error, onRetry: boundaryRetry }) => (
        <div className="p-6 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-center gap-3 mb-3">
            <AlertTriangle className="h-5 w-5 text-red-600" />
            <h3 className="font-medium text-red-800">{title}</h3>
          </div>
          <p className="text-sm text-red-700 mb-4">{description}</p>
          {import.meta.env?.DEV && error && (
            <pre className="text-xs text-red-600 bg-red-100 p-2 rounded mb-4 overflow-auto max-h-20">
              {error.message}
            </pre>
          )}
          <Button
            size="sm"
            variant="outline"
            onClick={onRetry || boundaryRetry}
            className="text-red-600 border-red-200 hover:bg-red-100"
          >
            <RefreshCcw className="h-4 w-4 mr-2" />
            Try Again
          </Button>
        </div>
      )}
    >
      {children}
    </ErrorBoundary>
  )
}