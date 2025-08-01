import { AuthContainer } from "@/components/Auth/AuthContainer";
import { AuthProvider, useAuth } from "@/contexts/AuthContext";
import { ThemeProvider } from "@/contexts/ThemeContext";
import { ProfileProvider } from "@/contexts/ProfileContext";
import { ToastProvider, ToastContainer } from "@/components/ui/toast";
import { ConfirmDialogProvider } from "@/components/ui/confirm-dialog";
import { ErrorBoundary } from "@/components/ui/error-boundary";
import { AuthenticatedApp } from "@/containers/AuthenticatedApp";

function AppContent() {
  const { isAuthenticated, isLoading, login } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-purple-50">
        <div className="text-center">
          <div className="h-12 w-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-gray-600">Loading Microservice Manager...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <AuthContainer
        onAuthSuccess={(user, token) => {
          login(user, token);
        }}
      />
    );
  }

  return <AuthenticatedApp />;
}

function App() {
  return (
    <ErrorBoundary>
      <ThemeProvider>
        <AuthProvider>
          <ProfileProvider>
            <ToastProvider>
              <ConfirmDialogProvider>
                <AppContent />
                <ToastContainer />
              </ConfirmDialogProvider>
            </ToastProvider>
          </ProfileProvider>
        </AuthProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
}

export default App;
