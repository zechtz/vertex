import { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useProfile } from '@/contexts/ProfileContext';

export function useOnboarding() {
  const { user, isAuthenticated } = useAuth();
  const { serviceProfiles, isLoading: isLoadingProfiles } = useProfile();
  const [shouldShowOnboarding, setShouldShowOnboarding] = useState(false);
  const [isOnboardingOpen, setIsOnboardingOpen] = useState(false);
  const [hasCheckedOnboarding, setHasCheckedOnboarding] = useState(false);

  // Check if user needs onboarding
  useEffect(() => {
    if (!isAuthenticated || !user || isLoadingProfiles || hasCheckedOnboarding) {
      return;
    }

    // Check if this is a first-time user (no profiles)
    const isFirstTimeUser = serviceProfiles.length === 0;
    
    // Check if user has dismissed onboarding before
    const hasSkippedOnboarding = localStorage.getItem(`onboarding-skipped-${user.id}`) === 'true';
    
    // Check if user has completed onboarding before
    const hasCompletedOnboarding = localStorage.getItem(`onboarding-completed-${user.id}`) === 'true';

    // Show onboarding if:
    // 1. User is a first-time user (no profiles)
    // 2. User hasn't explicitly skipped onboarding
    // 3. User hasn't completed onboarding before
    const shouldShow = isFirstTimeUser && !hasSkippedOnboarding && !hasCompletedOnboarding;
    
    setShouldShowOnboarding(shouldShow);
    setHasCheckedOnboarding(true);

    // Auto-open onboarding for first-time users
    if (shouldShow) {
      // Add a small delay to ensure UI is ready
      setTimeout(() => {
        setIsOnboardingOpen(true);
      }, 1000);
    }
  }, [isAuthenticated, user, serviceProfiles, isLoadingProfiles, hasCheckedOnboarding]);

  const openOnboarding = () => {
    setIsOnboardingOpen(true);
  };

  const closeOnboarding = () => {
    setIsOnboardingOpen(false);
    
    // Mark as skipped if user closes without completing
    if (user) {
      localStorage.setItem(`onboarding-skipped-${user.id}`, 'true');
    }
  };

  const completeOnboarding = () => {
    setIsOnboardingOpen(false);
    setShouldShowOnboarding(false);
    
    // Mark as completed
    if (user) {
      localStorage.setItem(`onboarding-completed-${user.id}`, 'true');
      // Remove skipped flag if it exists
      localStorage.removeItem(`onboarding-skipped-${user.id}`);
    }
  };

  const resetOnboarding = () => {
    if (user) {
      localStorage.removeItem(`onboarding-skipped-${user.id}`);
      localStorage.removeItem(`onboarding-completed-${user.id}`);
      setHasCheckedOnboarding(false);
      setShouldShowOnboarding(false);
    }
  };

  const forceShowOnboarding = () => {
    setShouldShowOnboarding(true);
    setIsOnboardingOpen(true);
  };

  return {
    shouldShowOnboarding,
    isOnboardingOpen,
    openOnboarding,
    closeOnboarding,
    completeOnboarding,
    resetOnboarding,
    forceShowOnboarding,
    hasCheckedOnboarding,
  };
}

export default useOnboarding;