import { useState, useEffect } from 'react';
import { 
  Plus, 
  Trash2, 
  Save, 
  X, 
  Settings, 
  Eye, 
  EyeOff,
  AlertCircle,
  Info
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useProfile } from '@/contexts/ProfileContext';
import { useToast, toast } from '@/components/ui/toast';
import { ServiceProfile } from '@/types';

interface ProfileEnvManagerProps {
  isOpen: boolean;
  onClose: () => void;
  profile: ServiceProfile | null;
}

interface EnvVar {
  name: string;
  value: string;
  description?: string;
  isRequired?: boolean;
}

export function ProfileEnvManager({ isOpen, onClose, profile }: ProfileEnvManagerProps) {
  const { getProfileEnvVars, setProfileEnvVar, deleteProfileEnvVar } = useProfile();
  const { addToast } = useToast();
  
  const [envVars, setEnvVars] = useState<Record<string, string>>({});
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [newVar, setNewVar] = useState<EnvVar>({
    name: '',
    value: '',
    description: '',
    isRequired: false,
  });
  const [showValues, setShowValues] = useState<Record<string, boolean>>({});
  const [editingVar, setEditingVar] = useState<string | null>(null);
  const [editValues, setEditValues] = useState<EnvVar>({
    name: '',
    value: '',
    description: '',
    isRequired: false,
  });

  // Load environment variables when profile changes
  useEffect(() => {
    if (profile && isOpen) {
      loadEnvVars();
    }
  }, [profile?.id, isOpen]);

  const loadEnvVars = async () => {
    if (!profile) return;
    
    try {
      setIsLoading(true);
      const vars = await getProfileEnvVars(profile.id);
      setEnvVars(vars);
    } catch (error) {
      console.error('Failed to load environment variables:', error);
      addToast(toast.error('Error', 'Failed to load environment variables'));
    } finally {
      setIsLoading(false);
    }
  };

  const handleAddVar = async () => {
    if (!profile || !newVar.name.trim()) {
      addToast(toast.error('Validation Error', 'Variable name is required'));
      return;
    }

    try {
      setIsSaving(true);
      await setProfileEnvVar(
        profile.id, 
        newVar.name.trim(), 
        newVar.value, 
        newVar.description || '', 
        newVar.isRequired || false
      );
      
      // Update local state
      setEnvVars(prev => ({
        ...prev,
        [newVar.name.trim()]: newVar.value
      }));
      
      // Reset form
      setNewVar({
        name: '',
        value: '',
        description: '',
        isRequired: false,
      });
      
      addToast(toast.success('Success', 'Environment variable added'));
    } catch (error) {
      console.error('Failed to add environment variable:', error);
      addToast(toast.error('Error', 'Failed to add environment variable'));
    } finally {
      setIsSaving(false);
    }
  };

  const handleEditVar = (name: string) => {
    setEditingVar(name);
    setEditValues({
      name,
      value: envVars[name] || '',
      description: '',
      isRequired: false,
    });
  };

  const handleSaveEdit = async () => {
    if (!profile || !editValues.name.trim()) return;

    try {
      setIsSaving(true);
      await setProfileEnvVar(
        profile.id,
        editValues.name.trim(),
        editValues.value,
        editValues.description || '',
        editValues.isRequired || false
      );

      // Update local state
      const newEnvVars = { ...envVars };
      if (editingVar && editingVar !== editValues.name.trim()) {
        delete newEnvVars[editingVar];
      }
      newEnvVars[editValues.name.trim()] = editValues.value;
      setEnvVars(newEnvVars);
      
      setEditingVar(null);
      addToast(toast.success('Success', 'Environment variable updated'));
    } catch (error) {
      console.error('Failed to update environment variable:', error);
      addToast(toast.error('Error', 'Failed to update environment variable'));
    } finally {
      setIsSaving(false);
    }
  };

  const handleDeleteVar = async (name: string) => {
    if (!profile) return;

    if (!confirm(`Are you sure you want to delete the environment variable "${name}"?`)) {
      return;
    }

    try {
      await deleteProfileEnvVar(profile.id, name);
      
      // Update local state
      const newEnvVars = { ...envVars };
      delete newEnvVars[name];
      setEnvVars(newEnvVars);
      
      addToast(toast.success('Success', 'Environment variable deleted'));
    } catch (error) {
      console.error('Failed to delete environment variable:', error);
      addToast(toast.error('Error', 'Failed to delete environment variable'));
    }
  };

  const toggleShowValue = (name: string) => {
    setShowValues(prev => ({
      ...prev,
      [name]: !prev[name]
    }));
  };

  if (!isOpen || !profile) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-4xl max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-600 flex items-center justify-between flex-shrink-0">
          <div className="flex items-center gap-3">
            <Settings className="h-6 w-6 text-blue-600" />
            <div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                Environment Variables
              </h2>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Profile: {profile.name}
              </p>
            </div>
          </div>
          <Button variant="ghost" onClick={onClose}>
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
                  Profile-Scoped Environment Variables
                </h3>
                <p className="text-sm text-blue-700 dark:text-blue-300 mt-1">
                  These environment variables are specific to the "{profile.name}" profile and will only be available when this profile is active.
                </p>
              </div>
            </div>
          </div>

          {/* Add New Variable */}
          <div className="mb-6 p-4 border border-gray-200 dark:border-gray-600 rounded-lg">
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
              Add Environment Variable
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Variable Name *
                </label>
                <Input
                  type="text"
                  value={newVar.name}
                  onChange={(e) => setNewVar(prev => ({ ...prev, name: e.target.value }))}
                  placeholder="e.g., DATABASE_URL"
                  className="font-mono"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Variable Value
                </label>
                <Input
                  type="text"
                  value={newVar.value}
                  onChange={(e) => setNewVar(prev => ({ ...prev, value: e.target.value }))}
                  placeholder="e.g., postgresql://localhost:5432/db"
                  className="font-mono"
                />
              </div>
            </div>
            <div className="mt-4">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Description (Optional)
              </label>
              <Input
                type="text"
                value={newVar.description}
                onChange={(e) => setNewVar(prev => ({ ...prev, description: e.target.value }))}
                placeholder="Describe what this variable is used for..."
              />
            </div>
            <div className="flex items-center justify-between mt-4">
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="isRequired"
                  checked={newVar.isRequired}
                  onChange={(e) => setNewVar(prev => ({ ...prev, isRequired: e.target.checked }))}
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <label htmlFor="isRequired" className="text-sm text-gray-700 dark:text-gray-300">
                  Required variable
                </label>
              </div>
              <Button
                onClick={handleAddVar}
                disabled={!newVar.name.trim() || isSaving}
                className="flex items-center gap-2"
              >
                {isSaving ? (
                  <div className="h-4 w-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                ) : (
                  <Plus className="h-4 w-4" />
                )}
                Add Variable
              </Button>
            </div>
          </div>

          {/* Existing Variables */}
          <div>
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
              Current Variables ({Object.keys(envVars).length})
            </h3>
            
            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <div className="h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin" />
              </div>
            ) : Object.keys(envVars).length === 0 ? (
              <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                <AlertCircle className="h-12 w-12 mx-auto mb-2 opacity-50" />
                <p>No environment variables configured for this profile</p>
                <p className="text-sm">Add your first variable above to get started</p>
              </div>
            ) : (
              <div className="space-y-3">
                {Object.entries(envVars).map(([name, value]) => (
                  <div
                    key={name}
                    className="p-4 border border-gray-200 dark:border-gray-600 rounded-lg"
                  >
                    {editingVar === name ? (
                      <div className="space-y-3">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                          <div>
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                              Variable Name
                            </label>
                            <Input
                              type="text"
                              value={editValues.name}
                              onChange={(e) => setEditValues(prev => ({ ...prev, name: e.target.value }))}
                              className="font-mono"
                            />
                          </div>
                          <div>
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                              Variable Value
                            </label>
                            <Input
                              type="text"
                              value={editValues.value}
                              onChange={(e) => setEditValues(prev => ({ ...prev, value: e.target.value }))}
                              className="font-mono"
                            />
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          <Button
                            size="sm"
                            onClick={handleSaveEdit}
                            disabled={isSaving}
                            className="flex items-center gap-1"
                          >
                            <Save className="h-3 w-3" />
                            Save
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => setEditingVar(null)}
                          >
                            <X className="h-3 w-3" />
                            Cancel
                          </Button>
                        </div>
                      </div>
                    ) : (
                      <div className="flex items-center justify-between">
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="font-mono text-sm font-medium text-gray-900 dark:text-gray-100">
                              {name}
                            </span>
                          </div>
                          <div className="flex items-center gap-2 mt-1">
                            <span className="font-mono text-sm text-gray-600 dark:text-gray-400">
                              {showValues[name] ? value : '•'.repeat(Math.min(value.length, 12))}
                            </span>
                            <button
                              onClick={() => toggleShowValue(name)}
                              className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                            >
                              {showValues[name] ? (
                                <EyeOff className="h-3 w-3" />
                              ) : (
                                <Eye className="h-3 w-3" />
                              )}
                            </button>
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleEditVar(name)}
                          >
                            Edit
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleDeleteVar(name)}
                            className="text-red-600 hover:text-red-700"
                          >
                            <Trash2 className="h-3 w-3" />
                          </Button>
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}