import React, { useState } from "react";
import { X, Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Service, EnvVar } from "@/types";
import { useProfile } from "@/contexts/ProfileContext";
import { ButtonSpinner } from "@/components/ui/spinner";
import { ErrorBoundarySection } from "@/components/ui/error-boundary";

interface ServiceConfigModalProps {
  service: Service | null;
  isOpen: boolean;
  isSaving?: boolean;
  onClose: () => void;
  onSave: (service: Service, profileId?: string) => void;
  isCreateMode?: boolean; // Explicitly pass create mode
}

export function ServiceConfigModal({
  service,
  isOpen,
  isSaving = false,
  onClose,
  onSave,
  isCreateMode = false,
}: ServiceConfigModalProps) {
  const [editingService, setEditingService] = useState<Service | null>(service);
  const [selectedProfileId, setSelectedProfileId] = useState<string>("");
  const { serviceProfiles, activeProfile } = useProfile();

  React.useEffect(() => {
    setEditingService(service);
    // Set default profile to active profile when creating new service
    if (isCreateMode && activeProfile) {
      setSelectedProfileId(activeProfile.id);
    }
    console.log(
      "ServiceConfigModal - isCreateMode:",
      isCreateMode,
      "serviceProfiles:",
      serviceProfiles.length,
      "profiles:",
      serviceProfiles,
    );
  }, [service, activeProfile, serviceProfiles, isCreateMode]);

  if (!isOpen || !editingService) return null;

  const handleSave = () => {
    if (editingService) {
      onSave(editingService, selectedProfileId || undefined);
    }
  };

  const addEnvVar = () => {
    if (!editingService) return;

    const newKey = `NEW_VAR_${Object.keys(editingService.envVars || {}).length + 1}`;
    setEditingService({
      ...editingService,
      envVars: {
        ...editingService.envVars,
        [newKey]: {
          name: newKey,
          value: "",
          description: "",
          isRequired: false,
        },
      },
    });
  };

  const removeEnvVar = (key: string) => {
    if (!editingService) return;

    const newEnvVars = { ...editingService.envVars };
    delete newEnvVars[key];

    setEditingService({
      ...editingService,
      envVars: newEnvVars,
    });
  };

  const updateEnvVar = (
    key: string,
    field: keyof EnvVar,
    value: string | boolean,
  ) => {
    if (!editingService) return;

    setEditingService({
      ...editingService,
      envVars: {
        ...editingService.envVars,
        [key]: {
          ...editingService.envVars[key],
          [field]: value,
        },
      },
    });
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <ErrorBoundarySection
          title="Service Configuration Error"
          description="Failed to load the service configuration form."
        >
          <div className="flex items-center justify-between p-6 border-b">
            <h2 className="text-xl font-semibold">
              {isCreateMode
                ? "Create New Service"
                : `Edit Service: ${editingService.name}`}
            </h2>
            <Button variant="ghost" size="sm" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          </div>

          <div className="p-6 space-y-4">
            {/* Basic Information */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="name">Service Name</Label>
                <Input
                  id="name"
                  value={editingService.name}
                  onChange={(e) =>
                    setEditingService({
                      ...editingService,
                      name: e.target.value,
                    })
                  }
                  disabled={false}
                  placeholder="Enter service name"
                />
              </div>
              <div>
                <Label htmlFor="dir">Directory</Label>
                <Input
                  id="dir"
                  value={editingService.dir}
                  onChange={(e) =>
                    setEditingService({
                      ...editingService,
                      dir: e.target.value,
                    })
                  }
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="port">Port</Label>
                <Input
                  id="port"
                  type="number"
                  min="1"
                  max="65535"
                  value={editingService.port}
                  onChange={(e) =>
                    setEditingService({
                      ...editingService,
                      port: parseInt(e.target.value) || 8080,
                    })
                  }
                  placeholder="Port number (1-65535)"
                />
              </div>
              <div>
                <Label htmlFor="order">Order</Label>
                <Input
                  id="order"
                  type="number"
                  value={editingService.order}
                  onChange={(e) =>
                    setEditingService({
                      ...editingService,
                      order: parseInt(e.target.value) || 1,
                    })
                  }
                />
              </div>
            </div>

            {/* Profile Selection for new services */}
            {isCreateMode && (
              <div>
                <Label htmlFor="profile">Add to Profile (Optional)</Label>
                <Select
                  value={selectedProfileId}
                  onValueChange={setSelectedProfileId}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select a profile to add service to (optional)" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">No profile</SelectItem>
                    {serviceProfiles.map((profile) => (
                      <SelectItem key={profile.id} value={profile.id}>
                        {profile.name} {profile.isActive ? "(Active)" : ""}{" "}
                        {profile.isDefault ? "(Default)" : ""}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-sm text-gray-500 mt-1">
                  Choose a profile to automatically add this service to it after
                  creation
                </p>
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="buildSystem">Build System</Label>
                <Select
                  value={editingService.buildSystem || "auto"}
                  onValueChange={(value) =>
                    setEditingService({
                      ...editingService,
                      buildSystem: value,
                    })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select build system" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="auto">Auto-detect</SelectItem>
                    <SelectItem value="maven">Maven</SelectItem>
                    <SelectItem value="gradle">Gradle</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label className="text-sm text-gray-500">
                  Auto-detect will check for pom.xml (Maven) or build.gradle
                  (Gradle)
                </Label>
              </div>
            </div>

            <div>
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={editingService.description}
                onChange={(e) =>
                  setEditingService({
                    ...editingService,
                    description: e.target.value,
                  })
                }
                placeholder="Service description..."
                rows={3}
              />
            </div>

            <div>
              <Label htmlFor="healthUrl">Health URL</Label>
              <Input
                id="healthUrl"
                value={editingService.healthUrl}
                onChange={(e) =>
                  setEditingService({
                    ...editingService,
                    healthUrl: e.target.value,
                  })
                }
                placeholder="http://localhost:8080/actuator/health"
              />
            </div>

            <div>
              <Label htmlFor="javaOpts">Java Options</Label>
              <Textarea
                id="javaOpts"
                value={editingService.javaOpts}
                onChange={(e) =>
                  setEditingService({
                    ...editingService,
                    javaOpts: e.target.value,
                  })
                }
                placeholder="Additional Java options..."
                rows={2}
              />
            </div>

            <div className="flex items-center space-x-2">
              <Checkbox
                id="isEnabled"
                checked={editingService.isEnabled}
                onCheckedChange={(checked) =>
                  setEditingService({
                    ...editingService,
                    isEnabled: checked === true,
                  })
                }
              />
              <Label htmlFor="isEnabled">Enable this service</Label>
            </div>

            <div className="flex items-center space-x-2">
              <Checkbox
                id="verboseLogging"
                checked={editingService.verboseLogging || false}
                onCheckedChange={(checked) =>
                  setEditingService({
                    ...editingService,
                    verboseLogging: checked === true,
                  })
                }
              />
              <Label htmlFor="verboseLogging" className="text-sm">
                Enable verbose logging (Maven: -X, Gradle: -i)
              </Label>
            </div>

            {/* Environment Variables */}
            <div>
              <div className="flex items-center justify-between mb-3">
                <Label>Environment Variables</Label>
                <Button variant="outline" size="sm" onClick={addEnvVar}>
                  <Plus className="h-4 w-4 mr-1" />
                  Add Variable
                </Button>
              </div>

              <div className="space-y-3 max-h-60 overflow-y-auto">
                {Object.entries(editingService.envVars || {}).map(
                  ([key, envVar]) => (
                    <div key={key} className="p-3 border rounded-lg space-y-2">
                      <div className="flex items-center justify-between">
                        <Input
                          value={envVar.name}
                          onChange={(e) =>
                            updateEnvVar(key, "name", e.target.value)
                          }
                          placeholder="Variable name"
                          className="flex-1 mr-2"
                        />
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => removeEnvVar(key)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                      <Input
                        value={envVar.value}
                        onChange={(e) =>
                          updateEnvVar(key, "value", e.target.value)
                        }
                        placeholder="Variable value"
                      />
                      <Input
                        value={envVar.description || ""}
                        onChange={(e) =>
                          updateEnvVar(key, "description", e.target.value)
                        }
                        placeholder="Description (optional)"
                      />
                      <div className="flex items-center space-x-2">
                        <Checkbox
                          checked={envVar.isRequired}
                          onCheckedChange={(checked) =>
                            updateEnvVar(key, "isRequired", checked === true)
                          }
                        />
                        <Label className="text-sm">Required</Label>
                      </div>
                    </div>
                  ),
                )}
                {Object.keys(editingService.envVars || {}).length === 0 && (
                  <p className="text-muted-foreground text-sm text-center py-4">
                    No environment variables defined
                  </p>
                )}
              </div>
            </div>
          </div>

          <div className="flex justify-end gap-3 p-6 border-t">
            <Button variant="outline" onClick={onClose} disabled={isSaving}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={isSaving}>
              <ButtonSpinner
                isLoading={isSaving}
                loadingText={isCreateMode ? "Creating..." : "Saving..."}
              >
                {isCreateMode ? "Create Service" : "Save Changes"}
              </ButtonSpinner>
            </Button>
          </div>
        </ErrorBoundarySection>
      </div>
    </div>
  );
}
