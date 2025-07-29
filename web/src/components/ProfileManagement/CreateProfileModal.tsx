import { useState, useEffect } from 'react';
import { X, Plus, Trash2, Server, Settings, Star, Search } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useProfile } from '@/contexts/ProfileContext';
import { CreateProfileRequest, Service } from '@/types';
import { useToast, toast } from '@/components/ui/toast';

interface CreateProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function CreateProfileModal({ isOpen, onClose }: CreateProfileModalProps) {
  const { createProfile, isCreating } = useProfile();
  const { addToast } = useToast();
  const [availableServices, setAvailableServices] = useState<Service[]>([]);
  const [formData, setFormData] = useState<CreateProfileRequest>({
    name: '',
    description: '',
    services: [],
    envVars: {},
    projectsDir: '',
    javaHomeOverride: '',
    isDefault: false,
  });
  const [envVarKey, setEnvVarKey] = useState('');
  const [envVarValue, setEnvVarValue] = useState('');

  // Fetch available services
  useEffect(() => {
    if (isOpen) {
      fetchServices();
    }
  }, [isOpen]);

  const fetchServices = async () => {
    try {
      const response = await fetch('/api/services');
      if (response.ok) {
        const services = await response.json();
        setAvailableServices(services || []);
      }
    } catch (error) {
      console.error('Failed to fetch services:', error);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.name.trim()) {
      addToast(toast.error('Validation Error', 'Profile name is required'));
      return;
    }

    // Services are now optional - profiles can be created without services

    try {
      await createProfile(formData);
      addToast(toast.success('Success', 'Profile created successfully!'));
      handleClose();
    } catch (error) {
      console.error('Failed to create profile:', error);
      addToast(toast.error('Error', 'Failed to create profile. Please try again.'));
    }
  };

  const handleClose = () => {
    setFormData({
      name: '',
      description: '',
      services: [],
      envVars: {},
      projectsDir: '',
      javaHomeOverride: '',
      isDefault: false,
    });
    setEnvVarKey('');
    setEnvVarValue('');
    onClose();
  };

  const handleServiceToggle = (serviceName: string) => {
    setFormData(prev => ({
      ...prev,
      services: prev.services.includes(serviceName)
        ? prev.services.filter(s => s !== serviceName)
        : [...prev.services, serviceName]
    }));
  };

  const handleAddEnvVar = () => {
    if (envVarKey.trim() && envVarValue.trim()) {
      setFormData(prev => ({
        ...prev,
        envVars: {
          ...prev.envVars,
          [envVarKey.trim()]: envVarValue.trim()
        }
      }));
      setEnvVarKey('');
      setEnvVarValue('');
    }
  };

  const handleRemoveEnvVar = (key: string) => {
    setFormData(prev => {
      const newEnvVars = { ...prev.envVars };
      delete newEnvVars[key];
      return {
        ...prev,
        envVars: newEnvVars
      };
    });
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-600 flex items-center justify-between flex-shrink-0">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
            Create New Profile
          </h2>
          <Button variant="ghost" onClick={handleClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="flex flex-col flex-1 min-h-0">
          <div className="flex-1 overflow-y-auto p-6 space-y-6">
            {/* Basic Information */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">
                Basic Information
              </h3>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Profile Name *
                </label>
                <Input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                  placeholder="e.g., Nest Development, Production Environment"
                  className="w-full"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Description
                </label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                  placeholder="Describe this profile and when to use it..."
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-gray-100"
                  rows={3}
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Projects Directory
                  </label>
                  <Input
                    type="text"
                    value={formData.projectsDir}
                    onChange={(e) => setFormData(prev => ({ ...prev, projectsDir: e.target.value }))}
                    placeholder="/path/to/your/projects"
                    className="w-full"
                  />
                  <p className="text-xs text-gray-500 mt-1">Directory where your project files are located</p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Java Home Override (Optional)
                  </label>
                  <Input
                    type="text"
                    value={formData.javaHomeOverride}
                    onChange={(e) => setFormData(prev => ({ ...prev, javaHomeOverride: e.target.value }))}
                    placeholder="/path/to/java"
                    className="w-full"
                  />
                  <p className="text-xs text-gray-500 mt-1">Override default Java installation path</p>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="isDefault"
                  checked={formData.isDefault}
                  onChange={(e) => setFormData(prev => ({ ...prev, isDefault: e.target.checked }))}
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <label htmlFor="isDefault" className="flex items-center gap-2 text-sm font-medium text-gray-700 dark:text-gray-300">
                  <Star className="h-4 w-4" />
                  Set as default profile
                </label>
              </div>
            </div>

            {/* Service Selection */}
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">
                  Services ({formData.services.length} selected)
                </h3>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  Optional - You can add services later
                </span>
              </div>
              
              {availableServices.length === 0 ? (
                <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                  <Server className="h-12 w-12 mx-auto mb-2 opacity-50" />
                  <p>No services available</p>
                </div>
              ) : (
                <div>
                  <div className="p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg mb-4">
                    <div className="space-y-2">
                      <p className="text-sm text-blue-800 dark:text-blue-200">
                        💡 <strong>Flexible Service Management:</strong>
                      </p>
                      <ul className="text-xs text-blue-700 dark:text-blue-300 space-y-1 ml-4">
                        <li>• Create profiles without any services and add them later</li>
                        <li>• Use Auto-Discovery to find Maven/Gradle projects in your workspace</li>
                        <li>• Create custom services for non-NeST projects</li>
                        <li>• Mix default microservices with your own custom services</li>
                      </ul>
                    </div>
                  </div>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 max-h-60 overflow-y-auto">
                    {availableServices.map((service) => (
                      <label
                        key={service.name}
                        className={`flex items-center gap-3 p-3 border rounded-lg cursor-pointer transition-colors ${
                          formData.services.includes(service.name)
                            ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                            : 'border-gray-200 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700'
                        }`}
                      >
                        <input
                          type="checkbox"
                          checked={formData.services.includes(service.name)}
                          onChange={() => handleServiceToggle(service.name)}
                          className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                        />
                        <div className="flex-1">
                          <div className="font-medium text-gray-900 dark:text-gray-100">
                            {service.name}
                          </div>
                          {service.description && (
                            <div className="text-sm text-gray-600 dark:text-gray-400">
                              {service.description}
                            </div>
                          )}
                          <div className="text-xs text-gray-500 dark:text-gray-400">
                            Port: {service.port} • Status: {service.status}
                          </div>
                        </div>
                      </label>
                    ))}
                  </div>
                  
                  {/* Service Discovery Actions */}
                  <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-600">
                    <div className="flex flex-wrap gap-2">
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          // This would open the auto-discovery modal
                          // We'll implement this functionality
                          alert('Auto-discovery feature: Coming soon! This will scan your workspace for Maven/Gradle projects.');
                        }}
                        className="flex items-center gap-2"
                      >
                        <Search className="h-4 w-4" />
                        Auto-Discover Services
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          // This would open the service creation modal
                          alert('Custom service creation: Available through the main services interface after creating the profile.');
                        }}
                        className="flex items-center gap-2"
                      >
                        <Plus className="h-4 w-4" />
                        Create Custom Service
                      </Button>
                    </div>
                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
                      These actions are also available after creating the profile
                    </p>
                  </div>
                </div>
              )}
            </div>

            {/* Environment Variables */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">
                Environment Variables
              </h3>
              
              {/* Add Environment Variable */}
              <div className="flex gap-2">
                <Input
                  type="text"
                  value={envVarKey}
                  onChange={(e) => setEnvVarKey(e.target.value)}
                  placeholder="Variable name"
                  className="flex-1"
                />
                <Input
                  type="text"
                  value={envVarValue}
                  onChange={(e) => setEnvVarValue(e.target.value)}
                  placeholder="Variable value"
                  className="flex-1"
                />
                <Button
                  type="button"
                  onClick={handleAddEnvVar}
                  disabled={!envVarKey.trim() || !envVarValue.trim()}
                  className="flex items-center gap-1"
                >
                  <Plus className="h-4 w-4" />
                  Add
                </Button>
              </div>

              {/* Environment Variables List */}
              {Object.keys(formData.envVars).length > 0 && (
                <div className="space-y-2 max-h-40 overflow-y-auto">
                  {Object.entries(formData.envVars).map(([key, value]) => (
                    <div
                      key={key}
                      className="flex items-center gap-2 p-2 bg-gray-50 dark:bg-gray-700 rounded border"
                    >
                      <Settings className="h-4 w-4 text-gray-400" />
                      <span className="font-mono text-sm text-gray-900 dark:text-gray-100">
                        {key}
                      </span>
                      <span className="text-gray-500">=</span>
                      <span className="font-mono text-sm text-gray-600 dark:text-gray-300 flex-1 truncate">
                        {value}
                      </span>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRemoveEnvVar(key)}
                        className="text-red-600 hover:text-red-700"
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Footer */}
          <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 flex items-center justify-end gap-3 flex-shrink-0">
            <Button type="button" variant="outline" onClick={handleClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isCreating || !formData.name.trim()}
              className="flex items-center gap-2"
            >
              {isCreating ? (
                <div className="h-4 w-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              ) : (
                <Plus className="h-4 w-4" />
              )}
              Create Profile
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}