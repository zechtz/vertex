import { useState } from 'react';
import {
  X,
  CheckCircle,
  User,
  Search,
  Settings,
  ArrowRight,
  ArrowLeft,
  Loader2,
  Star,
  ChevronRight
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useProfile } from '@/contexts/ProfileContext';
import { useToast, toast } from '@/components/ui/toast';
import { Service } from '@/types';
import { OnboardingProfileStep } from './OnboardingProfileStep';
import { OnboardingDiscoveryStep } from './OnboardingDiscoveryStep';
import { OnboardingConfigurationStep } from './OnboardingConfigurationStep';

interface OnboardingStepperProps {
  isOpen: boolean;
  onClose: () => void;
  onComplete: () => void;
}

export interface OnboardingStep {
  id: number;
  title: string;
  description: string;
  icon: React.ComponentType<{ className?: string }>;
  completed: boolean;
}

export function OnboardingStepper({
  isOpen,
  onClose,
  onComplete
}: OnboardingStepperProps) {
  const { refreshProfiles } = useProfile();
  const { addToast } = useToast();

  const [currentStep, setCurrentStep] = useState(0);
  const [isProcessing, setIsProcessing] = useState(false);
  const [createdProfile, setCreatedProfile] = useState<any>(null);
  const [discoveredServices, setDiscoveredServices] = useState<Service[]>([]);

  const [steps, setSteps] = useState<OnboardingStep[]>([
    {
      id: 0,
      title: 'Create Profile',
      description: 'Set up your workspace profile',
      icon: User,
      completed: false,
    },
    {
      id: 1,
      title: 'Discover Services',
      description: 'Auto-discover your microservices',
      icon: Search,
      completed: false,
    },
    {
      id: 2,
      title: 'Configure Order',
      description: 'Set service startup order',
      icon: Settings,
      completed: false,
    },
  ]);

  const updateStepCompleted = (stepId: number, completed: boolean) => {
    setSteps(prev => prev.map(step => 
      step.id === stepId ? { ...step, completed } : step
    ));
  };

  const handleProfileCreated = (profile: any) => {
    setCreatedProfile(profile);
    updateStepCompleted(0, true);
    setCurrentStep(1);
  };

  const handleServicesDiscovered = (services: Service[]) => {
    setDiscoveredServices(services);
    updateStepCompleted(1, true);
    setCurrentStep(2);
  };

  const handleConfigurationComplete = async () => {
    updateStepCompleted(2, true);
    
    // Refresh profiles to get the latest data
    await refreshProfiles();
    
    // Add success toast
    addToast(
      toast.success(
        'Onboarding Complete!',
        `Welcome! Your profile "${createdProfile?.name}" has been set up with ${discoveredServices.length} services.`
      )
    );

    // Close onboarding
    onComplete();
  };

  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleSkip = () => {
    // Skip to the end and close
    addToast(
      toast.info(
        'Onboarding Skipped',
        'You can always create profiles and discover services later from the sidebar.'
      )
    );
    onComplete();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-900 rounded-xl shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-blue-50 to-purple-50 dark:from-gray-800 dark:to-gray-800">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-500 rounded-lg">
              <Star className="w-6 h-6 text-white" />
            </div>
            <div>
              <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                Welcome to Vertex
              </h2>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Let's get your microservice environment set up in just a few steps
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Button
              onClick={handleSkip}
              variant="ghost"
              size="sm"
              className="text-gray-500 hover:text-gray-700"
            >
              Skip for now
            </Button>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 p-1"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Progress Steps */}
        <div className="px-6 py-4 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            {steps.map((step, index) => (
              <div key={step.id} className="flex items-center flex-1">
                <div className={`flex items-center gap-3 ${
                  index === currentStep 
                    ? 'text-blue-600 dark:text-blue-400' 
                    : step.completed 
                      ? 'text-green-600 dark:text-green-400' 
                      : 'text-gray-400'
                }`}>
                  <div className={`p-2 rounded-full border-2 ${
                    index === currentStep
                      ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                      : step.completed
                        ? 'border-green-500 bg-green-50 dark:bg-green-900/20'
                        : 'border-gray-300 bg-gray-50 dark:bg-gray-700'
                  }`}>
                    {step.completed ? (
                      <CheckCircle className="w-5 h-5" />
                    ) : (
                      <step.icon className="w-5 h-5" />
                    )}
                  </div>
                  <div className="hidden sm:block">
                    <div className="font-medium text-sm">{step.title}</div>
                    <div className="text-xs opacity-75">{step.description}</div>
                  </div>
                </div>
                {index < steps.length - 1 && (
                  <ChevronRight className="w-4 h-4 text-gray-300 mx-4 flex-shrink-0" />
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Step Content */}
        <div className="p-6 min-h-[400px] max-h-[60vh] overflow-y-auto">
          {currentStep === 0 && (
            <OnboardingProfileStep
              onProfileCreated={handleProfileCreated}
              isProcessing={isProcessing}
              setIsProcessing={setIsProcessing}
            />
          )}
          
          {currentStep === 1 && (
            <OnboardingDiscoveryStep
              profile={createdProfile}
              onServicesDiscovered={handleServicesDiscovered}
              isProcessing={isProcessing}
              setIsProcessing={setIsProcessing}
            />
          )}
          
          {currentStep === 2 && (
            <OnboardingConfigurationStep
              services={discoveredServices}
              onConfigurationComplete={handleConfigurationComplete}
              isProcessing={isProcessing}
              setIsProcessing={setIsProcessing}
            />
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between p-6 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
          <div className="text-sm text-gray-500 dark:text-gray-400">
            Step {currentStep + 1} of {steps.length}
          </div>
          
          <div className="flex items-center gap-3">
            <Button
              onClick={handleBack}
              variant="outline"
              disabled={currentStep === 0 || isProcessing}
            >
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back
            </Button>
            
            {currentStep < steps.length - 1 && (
              <Button
                onClick={handleNext}
                disabled={!steps[currentStep].completed || isProcessing}
              >
                {isProcessing ? (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                ) : (
                  <ArrowRight className="w-4 h-4 mr-2" />
                )}
                Next
              </Button>
            )}
            
            {currentStep === steps.length - 1 && steps[currentStep].completed && (
              <Button
                onClick={handleConfigurationComplete}
                disabled={isProcessing}
                className="bg-green-600 hover:bg-green-700"
              >
                {isProcessing ? (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                ) : (
                  <CheckCircle className="w-4 h-4 mr-2" />
                )}
                Complete Setup
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default OnboardingStepper;