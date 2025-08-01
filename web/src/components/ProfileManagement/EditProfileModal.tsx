import { useState, useEffect } from "react";
import { X, Plus, Trash2, Server, Settings, Star, Save, Upload } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useProfile } from "@/contexts/ProfileContext";
import { UpdateProfileRequest, Service, ServiceProfile } from "@/types";
import { useToast, toast } from "@/components/ui/toast";
import { BulkImportModal } from "@/components/EnvironmentVariables/BulkImportModal";

interface EditProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
  profile: ServiceProfile | null;
}

export function EditProfileModal({
  isOpen,
  onClose,
  profile,
}: EditProfileModalProps) {
  const { updateProfile, isUpdating } = useProfile();
  const { addToast } = useToast();
  const [availableServices, setAvailableServices] = useState<Service[]>([]);
  const [formData, setFormData] = useState<UpdateProfileRequest>({
    name: "",
    description: "",
    services: [],
    envVars: {},
    projectsDir: "",
    javaHomeOverride: "",
    isDefault: false,
  });
  const [envVarKey, setEnvVarKey] = useState("");
  const [envVarValue, setEnvVarValue] = useState("");
  const [isBulkImportOpen, setIsBulkImportOpen] = useState(false);

  // Initialize form data when profile changes
  useEffect(() => {
    if (profile && isOpen) {
      // Extract service IDs from the enriched service objects
      const serviceIds = profile.services.map(service => 
        typeof service === 'string' ? service : service.id
      );
      setFormData({
        name: profile.name,
        description: profile.description,
        services: [...serviceIds],
        envVars: { ...profile.envVars },
        projectsDir: profile.projectsDir || "",
        javaHomeOverride: profile.javaHomeOverride || "",
        isDefault: profile.isDefault,
      });
      fetchServices();
    }
  }, [profile, isOpen]);

  const fetchServices = async () => {
    if (!profile) return;

    try {
      const token = localStorage.getItem("authToken");
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };

      if (token) {
        headers["Authorization"] = `Bearer ${token}`;
      }

      // Include excludeProfile parameter to allow services from the current profile being edited
      const response = await fetch(
        `/api/services/available-for-profile?excludeProfile=${profile.id}`,
        {
          method: "GET",
          headers,
        },
      );

      if (response.ok) {
        const services = await response.json();
        setAvailableServices(services || []);
      } else if (response.status === 401) {
        console.error(
          "Authentication required for fetching available services",
        );
        addToast(
          toast.error(
            "Authentication Error",
            "Please log in to view available services",
          ),
        );
      }
    } catch (error) {
      console.error("Failed to fetch services:", error);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!profile) return;

    if (!formData.name.trim()) {
      addToast(toast.error("Validation Error", "Profile name is required"));
      return;
    }

    // Services are now optional - profiles can be created without services

    try {
      await updateProfile(profile.id, formData);
      addToast(toast.success("Success", "Profile updated successfully!"));
      handleClose();
    } catch (error) {
      console.error("Failed to update profile:", error);
      addToast(
        toast.error("Error", "Failed to update profile. Please try again."),
      );
    }
  };

  const handleClose = () => {
    setEnvVarKey("");
    setEnvVarValue("");
    onClose();
  };

  const handleServiceToggle = (serviceId: string) => {
    setFormData((prev) => ({
      ...prev,
      services: prev.services.includes(serviceId)
        ? prev.services.filter((s) => s !== serviceId)
        : [...prev.services, serviceId],
    }));
  };

  const handleAddEnvVar = () => {
    if (envVarKey.trim() && envVarValue.trim()) {
      setFormData((prev) => ({
        ...prev,
        envVars: {
          ...prev.envVars,
          [envVarKey.trim()]: envVarValue.trim(),
        },
      }));
      setEnvVarKey("");
      setEnvVarValue("");
    }
  };

  const handleRemoveEnvVar = (key: string) => {
    setFormData((prev) => {
      const newEnvVars = { ...prev.envVars };
      delete newEnvVars[key];
      return {
        ...prev,
        envVars: newEnvVars,
      };
    });
  };

  const handleEditEnvVar = (
    oldKey: string,
    newKey: string,
    newValue: string,
  ) => {
    if (newKey.trim() && newValue.trim()) {
      setFormData((prev) => {
        const newEnvVars = { ...prev.envVars };
        if (oldKey !== newKey) {
          delete newEnvVars[oldKey];
        }
        newEnvVars[newKey.trim()] = newValue.trim();
        return {
          ...prev,
          envVars: newEnvVars,
        };
      });
    }
  };

  const handleBulkImport = (variables: Record<string, string>) => {
    const existingKeys = new Set(Object.keys(formData.envVars));
    const newVariables = Object.entries(variables)
      .filter(([key]) => !existingKeys.has(key))
      .reduce((acc, [key, value]) => ({ ...acc, [key]: value }), {});
    
    const duplicateCount = Object.keys(variables).length - Object.keys(newVariables).length;
    
    setFormData((prev) => ({
      ...prev,
      envVars: { ...prev.envVars, ...newVariables },
    }));
    
    addToast(toast.success(
      "Variables imported",
      `Imported ${Object.keys(newVariables).length} new variables${duplicateCount > 0 ? `. ${duplicateCount} duplicates skipped.` : ""}`
    ));
    
    setIsBulkImportOpen(false);
  };

  if (!isOpen || !profile) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-600 flex items-center justify-between flex-shrink-0">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
            Edit Profile: {profile.name}
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
                  onChange={(e) =>
                    setFormData((prev) => ({ ...prev, name: e.target.value }))
                  }
                  placeholder="e.g., Development, Production Environment"
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
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      description: e.target.value,
                    }))
                  }
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
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        projectsDir: e.target.value,
                      }))
                    }
                    placeholder="/path/to/your/projects"
                    className="w-full"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Directory where your project files are located
                  </p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Java Home Override (Optional)
                  </label>
                  <Input
                    type="text"
                    value={formData.javaHomeOverride}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        javaHomeOverride: e.target.value,
                      }))
                    }
                    placeholder="/path/to/java"
                    className="w-full"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Override default Java installation path
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="isDefault"
                  checked={formData.isDefault}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      isDefault: e.target.checked,
                    }))
                  }
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <label
                  htmlFor="isDefault"
                  className="flex items-center gap-2 text-sm font-medium text-gray-700 dark:text-gray-300"
                >
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
                  Optional - Services can be added later
                </span>
              </div>

              {availableServices.length === 0 ? (
                <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                  <Server className="h-12 w-12 mx-auto mb-2 opacity-50" />
                  <p>No services available</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 max-h-60 overflow-y-auto">
                  {availableServices.map((service) => (
                    <label
                      key={service.name}
                      className={`flex items-center gap-3 p-3 border rounded-lg cursor-pointer transition-colors ${
                        formData.services.includes(service.id)
                          ? "border-blue-500 bg-blue-50 dark:bg-blue-900/20"
                          : "border-gray-200 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700"
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={formData.services.includes(service.id)}
                        onChange={() => handleServiceToggle(service.id)}
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
              )}
            </div>

            {/* Environment Variables */}
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">
                  Environment Variables
                </h3>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setIsBulkImportOpen(true)}
                  className="flex items-center gap-2"
                >
                  <Upload className="h-4 w-4" />
                  Bulk Import
                </Button>
              </div>

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
                    <EditableEnvVar
                      key={key}
                      originalKey={key}
                      originalValue={value}
                      onUpdate={handleEditEnvVar}
                      onRemove={handleRemoveEnvVar}
                    />
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
              disabled={isUpdating || !formData.name.trim()}
              className="flex items-center gap-2"
            >
              {isUpdating ? (
                <div className="h-4 w-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              ) : (
                <Save className="h-4 w-4" />
              )}
              Save Changes
            </Button>
          </div>
        </form>
      </div>
      
      {/* Bulk Import Modal */}
      <BulkImportModal
        isOpen={isBulkImportOpen}
        onClose={() => setIsBulkImportOpen(false)}
        onImport={handleBulkImport}
      />
    </div>
  );
}

// Editable Environment Variable Component
interface EditableEnvVarProps {
  originalKey: string;
  originalValue: string;
  onUpdate: (oldKey: string, newKey: string, newValue: string) => void;
  onRemove: (key: string) => void;
}

function EditableEnvVar({
  originalKey,
  originalValue,
  onUpdate,
  onRemove,
}: EditableEnvVarProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [key, setKey] = useState(originalKey);
  const [value, setValue] = useState(originalValue);

  const handleSave = () => {
    if (key.trim() && value.trim()) {
      onUpdate(originalKey, key.trim(), value.trim());
      setIsEditing(false);
    }
  };

  const handleCancel = () => {
    setKey(originalKey);
    setValue(originalValue);
    setIsEditing(false);
  };

  return (
    <div className="flex items-center gap-2 p-2 bg-gray-50 dark:bg-gray-700 rounded border">
      <Settings className="h-4 w-4 text-gray-400" />

      {isEditing ? (
        <>
          <Input
            type="text"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            className="font-mono text-sm flex-1"
          />
          <span className="text-gray-500">=</span>
          <Input
            type="text"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            className="font-mono text-sm flex-1"
          />
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleSave}
            className="text-green-600 hover:text-green-700"
          >
            <Save className="h-3 w-3" />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleCancel}
            className="text-gray-600 hover:text-gray-700"
          >
            <X className="h-3 w-3" />
          </Button>
        </>
      ) : (
        <>
          <button
            type="button"
            onClick={() => setIsEditing(true)}
            className="flex-1 text-left hover:bg-gray-100 dark:hover:bg-gray-600 p-1 rounded"
          >
            <span className="font-mono text-sm text-gray-900 dark:text-gray-100">
              {originalKey}
            </span>
            <span className="text-gray-500 mx-2">=</span>
            <span className="font-mono text-sm text-gray-600 dark:text-gray-300 truncate">
              {originalValue}
            </span>
          </button>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => onRemove(originalKey)}
            className="text-red-600 hover:text-red-700"
          >
            <Trash2 className="h-3 w-3" />
          </Button>
        </>
      )}
    </div>
  );
}
