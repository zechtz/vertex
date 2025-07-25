import React, { useState } from "react";
import { X, Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox";
import { Service, EnvVar } from "@/types";
import { ButtonSpinner } from "@/components/ui/spinner";
import { ErrorBoundarySection } from "@/components/ui/error-boundary";

interface ServiceConfigModalProps {
  service: Service | null;
  isOpen: boolean;
  isSaving?: boolean;
  onClose: () => void;
  onSave: (service: Service) => void;
}

export function ServiceConfigModal({
  service,
  isOpen,
  isSaving = false,
  onClose,
  onSave,
}: ServiceConfigModalProps) {
  const [editingService, setEditingService] = useState<Service | null>(service);

  React.useEffect(() => {
    setEditingService(service);
  }, [service]);

  if (!isOpen || !editingService) return null;

  const isCreateMode = !editingService.name || editingService.name === "";

  const handleSave = () => {
    if (editingService) {
      onSave(editingService);
      onClose();
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

  const updateEnvVar = (key: string, field: keyof EnvVar, value: string | boolean) => {
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
        <ErrorBoundarySection title="Service Configuration Error" description="Failed to load the service configuration form.">
        <div className="flex items-center justify-between p-6 border-b">
          <h2 className="text-xl font-semibold">
            {isCreateMode ? "Create New Service" : `Edit Service: ${editingService.name}`}
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
                disabled={!isCreateMode}
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
                value={editingService.port}
                onChange={(e) =>
                  setEditingService({
                    ...editingService,
                    port: parseInt(e.target.value) || 8080,
                  })
                }
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
              {Object.entries(editingService.envVars || {}).map(([key, envVar]) => (
                <div key={key} className="p-3 border rounded-lg space-y-2">
                  <div className="flex items-center justify-between">
                    <Input
                      value={envVar.name}
                      onChange={(e) => updateEnvVar(key, "name", e.target.value)}
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
                    onChange={(e) => updateEnvVar(key, "value", e.target.value)}
                    placeholder="Variable value"
                  />
                  <Input
                    value={envVar.description || ""}
                    onChange={(e) => updateEnvVar(key, "description", e.target.value)}
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
              ))}
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