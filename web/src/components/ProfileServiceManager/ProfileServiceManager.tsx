import { useState, useEffect } from 'react';
import { 
  Save, 
  X, 
  Server, 
  Check,
  AlertCircle,
  Info
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useProfile } from '@/contexts/ProfileContext';
import { useToast, toast } from '@/components/ui/toast';
import { ServiceProfile, Service } from '@/types';

interface ProfileServiceManagerProps {
  isOpen: boolean;
  onClose: () => void;
  profile: ServiceProfile | null;
  onProfileUpdated?: () => void;
}

export function ProfileServiceManager({ isOpen, onClose, profile, onProfileUpdated }: ProfileServiceManagerProps) {
  const { updateProfile } = useProfile();
  const { addToast } = useToast();
  
  const [availableServices, setAvailableServices] = useState<Service[]>([]);
  const [profileServices, setProfileServices] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  // Load available services and profile services when profile changes
  useEffect(() => {
    if (profile && isOpen) {
      loadAvailableServices();
      // Extract service IDs from the new object structure
      const serviceIds = profile.services.map(service => 
        typeof service === 'string' ? service : service.id
      );
      setProfileServices([...serviceIds]);
    }
  }, [profile?.id, isOpen]);

  const loadAvailableServices = async () => {
    try {
      setIsLoading(true);
      const response = await fetch('/api/services');
      if (response.ok) {
        const services = await response.json();
        setAvailableServices(services || []);
      }
    } catch (error) {
      console.error('Failed to load available services:', error);
      addToast(toast.error('Error', 'Failed to load available services'));
    } finally {
      setIsLoading(false);
    }
  };

  const handleToggleService = (serviceId: string) => {
    setProfileServices(prev => 
      prev.includes(serviceId)
        ? prev.filter(s => s !== serviceId)
        : [...prev, serviceId]
    );
  };

  const handleSave = async () => {
    if (!profile) return;

    try {
      setIsSaving(true);
      await updateProfile(profile.id, {
        name: profile.name,
        description: profile.description,
        services: profileServices,
        envVars: profile.envVars,
        projectsDir: profile.projectsDir,
        javaHomeOverride: profile.javaHomeOverride,
        isDefault: profile.isDefault,
      });
      
      addToast(toast.success('Success', 'Profile services updated successfully'));
      if (onProfileUpdated) {
        onProfileUpdated();
      }
      onClose();
    } catch (error) {
      console.error('Failed to update profile services:', error);
      addToast(toast.error('Error', 'Failed to update profile services'));
    } finally {
      setIsSaving(false);
    }
  };

  const handleClose = () => {
    if (profile) {
      // Extract service IDs from the new object structure
      const serviceIds = profile.services.map(service => 
        typeof service === 'string' ? service : service.id
      );
      setProfileServices([...serviceIds]);
    }
    onClose();
  };

  if (!isOpen || !profile) return null;

  const originalServiceIds = profile.services.map(service => 
    typeof service === 'string' ? service : service.id
  );
  
  const addedServices = profileServices.filter(serviceId => 
    !originalServiceIds.includes(serviceId)
  );
  const removedServices = originalServiceIds.filter(serviceId => 
    !profileServices.includes(serviceId)
  );
  const hasChanges = addedServices.length > 0 || removedServices.length > 0;

  // Helper function to get service name by ID
  const getServiceNameById = (serviceId: string) => {
    const service = availableServices.find(s => s.id === serviceId);
    return service ? service.name : serviceId;
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-4xl max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-600 flex items-center justify-between flex-shrink-0">
          <div className="flex items-center gap-3">
            <Server className="h-6 w-6 text-blue-600" />
            <div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                Manage Services
              </h2>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Profile: {profile.name}
              </p>
            </div>
          </div>
          <Button variant="ghost" onClick={handleClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {/* Info Banner */}
          <div className="mb-6 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
            <div className="flex items-start gap-3">
              <Info className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
              <div>
                <h3 className="text-sm font-medium text-blue-900 dark:text-blue-100">
                  Service Management
                </h3>
                <p className="text-sm text-blue-700 dark:text-blue-300 mt-1">
                  Select which services should be available when this profile is active. Only selected services will be visible in the main interface.
                </p>
              </div>
            </div>
          </div>

          {/* Service Discovery Guidance */}
          <div className="mb-6 p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
            <div className="flex items-start gap-3">
              <Server className="h-5 w-5 text-green-600 flex-shrink-0 mt-0.5" />
              <div className="flex-1">
                <h3 className="text-sm font-medium text-green-900 dark:text-green-100">
                  Need More Services?
                </h3>
                <div className="text-sm text-green-700 dark:text-green-300 mt-1 space-y-1">
                  <p>• Don't see your project? Use <strong>Auto-Discovery</strong> to scan for Maven/Gradle projects</p>
                  <p>• Need a custom service? Create one using the <strong>+ Create Service</strong> button in the main interface</p>
                  <p>• Services support both Maven and Gradle build systems automatically</p>
                </div>
              </div>
            </div>
          </div>

          {/* Changes Summary */}
          {hasChanges && (
            <div className="mb-6 p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
              <div className="flex items-start gap-3">
                <AlertCircle className="h-5 w-5 text-yellow-600 flex-shrink-0 mt-0.5" />
                <div className="flex-1">
                  <h3 className="text-sm font-medium text-yellow-900 dark:text-yellow-100">
                    Pending Changes
                  </h3>
                  <div className="text-sm text-yellow-700 dark:text-yellow-300 mt-1">
                    {addedServices.length > 0 && (
                      <p>
                        <span className="font-medium">Adding:</span> {addedServices.map(getServiceNameById).join(', ')}
                      </p>
                    )}
                    {removedServices.length > 0 && (
                      <p>
                        <span className="font-medium">Removing:</span> {removedServices.map(getServiceNameById).join(', ')}
                      </p>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Service Selection */}
          <div>
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
              Available Services ({profileServices.length} of {availableServices.length} selected)
            </h3>
            
            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <div className="h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin" />
              </div>
            ) : availableServices.length === 0 ? (
              <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                <Server className="h-12 w-12 mx-auto mb-2 opacity-50" />
                <p>No services available</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                {availableServices.map((service) => {
                  const isSelected = profileServices.includes(service.id);
                  const wasOriginallySelected = originalServiceIds.includes(service.id);
                  const isChanged = isSelected !== wasOriginallySelected;
                  
                  return (
                    <label
                      key={service.name}
                      className={`flex items-center gap-3 p-4 border-2 rounded-lg cursor-pointer transition-all ${
                        isSelected
                          ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                          : 'border-gray-200 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700'
                      } ${
                        isChanged ? 'ring-2 ring-yellow-400 ring-opacity-50' : ''
                      }`}
                    >
                      <div className="relative">
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => handleToggleService(service.id)}
                          className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                        />
                        {isSelected && (
                          <Check className="h-3 w-3 text-blue-600 absolute -top-1 -right-1" />
                        )}
                      </div>
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <span className="font-medium text-gray-900 dark:text-gray-100">
                            {service.name}
                          </span>
                          {isChanged && (
                            <span className={`px-1.5 py-0.5 text-xs rounded ${
                              isSelected 
                                ? 'bg-green-100 text-green-800' 
                                : 'bg-red-100 text-red-800'
                            }`}>
                              {isSelected ? 'Adding' : 'Removing'}
                            </span>
                          )}
                        </div>
                        {service.description && (
                          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                            {service.description}
                          </p>
                        )}
                        <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                          Port: {service.port} • Status: {service.status}
                        </div>
                      </div>
                    </label>
                  );
                })}
              </div>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 flex items-center justify-between flex-shrink-0">
          <div className="text-sm text-gray-600 dark:text-gray-400">
            {profileServices.length} service{profileServices.length !== 1 ? 's' : ''} selected
            {hasChanges && (
              <span className="ml-2 text-yellow-600 dark:text-yellow-400">
                • Unsaved changes
              </span>
            )}
          </div>
          <div className="flex items-center gap-3">
            <Button variant="outline" onClick={handleClose}>
              Cancel
            </Button>
            <Button
              onClick={handleSave}
              disabled={isSaving || !hasChanges}
              className="flex items-center gap-2"
            >
              {isSaving ? (
                <div className="h-4 w-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              ) : (
                <Save className="h-4 w-4" />
              )}
              Save Changes
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}