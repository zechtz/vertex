import { useState } from "react";
import {
  X,
  Play,
  Trash2,
  Plus,
  ArrowUp,
  ArrowDown,
  GripVertical,
  Edit,
} from "lucide-react";
import {
  DragDropContext,
  Droppable,
  Draggable,
  DropResult,
} from "react-beautiful-dnd";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Configuration, Service } from "@/types";
import { useToast, toast } from "@/components/ui/toast";
import { useConfirm } from "@/components/ui/confirm-dialog";
import { ButtonSpinner } from "@/components/ui/spinner";
import { ErrorBoundarySection } from "@/components/ui/error-boundary";

interface ConfigurationManagerProps {
  isOpen: boolean;
  onClose: () => void;
  configurations: Configuration[];
  services: Service[];
  onConfigurationSaved?: () => void;
}

export function ConfigurationManager({
  isOpen,
  onClose,
  configurations,
  services,
  onConfigurationSaved,
}: ConfigurationManagerProps) {
  const [newConfigName, setNewConfigName] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const [editingConfig, setEditingConfig] = useState<Configuration | null>(
    null,
  );
  const [configServices, setConfigServices] = useState<
    Array<{ id: string; name: string; order: number; enabled: boolean }>
  >([]);
  const [isSaving, setIsSaving] = useState(false);

  const { addToast } = useToast();
  const { showConfirm } = useConfirm();

  if (!isOpen) return null;

  const startCreating = () => {
    const servicesWithConfig = services
      .map((service) => ({
        id: service.id,
        name: service.name,
        order: service.order,
        enabled: true,
      }))
      .sort((a, b) => a.order - b.order);

    setConfigServices(servicesWithConfig);
    setEditingConfig(null);
    setIsCreating(true);
  };

  const startEditing = (config: Configuration) => {
    // Initialize services from configuration
    const existingServices = config.services.map((s) => ({
      id: s.id,
      name: s.name,
      order: s.order,
      enabled: true,
    }));

    // Add any missing services as disabled
    const missingServices = services
      .filter((service) => !existingServices.find((es) => es.id === service.id))
      .map((service) => ({
        id: service.id,
        name: service.name,
        order: service.order,
        enabled: false,
      }));

    const allServices = [...existingServices, ...missingServices].sort(
      (a, b) => a.order - b.order,
    );

    setConfigServices(allServices);
    setEditingConfig(config);
    setNewConfigName(config.name);
    setIsCreating(true);
  };

  const handleSaveConfig = async () => {
    if (!newConfigName.trim()) {
      addToast(
        toast.warning(
          "Configuration name required",
          "Please enter a name for the configuration",
        ),
      );
      return;
    }

    const enabledServices = configServices
      .filter((s) => s.enabled)
      .map((s) => ({ id: s.id, name: s.name, order: s.order }));

    if (enabledServices.length === 0) {
      addToast(
        toast.warning(
          "No services selected",
          "Please select at least one service for the configuration",
        ),
      );
      return;
    }

    const configData = {
      id: editingConfig?.id || `config-${Date.now()}`,
      name: newConfigName,
      services: enabledServices,
      isDefault: false,
    };

    try {
      setIsSaving(true);
      const method = editingConfig ? "PUT" : "POST";
      const url = editingConfig
        ? `/api/configurations/${editingConfig.id}`
        : "/api/configurations";

      const response = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(configData),
      });

      if (!response.ok) {
        throw new Error(
          `Failed to save configuration: ${response.status} ${response.statusText}`,
        );
      }

      setNewConfigName("");
      setIsCreating(false);
      setEditingConfig(null);
      setConfigServices([]);
      addToast(
        toast.success(
          `Configuration ${editingConfig ? "updated" : "created"}`,
          `Successfully ${editingConfig ? "updated" : "created"} configuration "${newConfigName}" with ${enabledServices.length} services`,
        ),
      );
      onConfigurationSaved?.();
    } catch (error) {
      console.error("Failed to save configuration:", error);
      addToast(
        toast.error(
          "Failed to save configuration",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    } finally {
      setIsSaving(false);
    }
  };

  const moveService = (index: number, direction: "up" | "down") => {
    const newServices = [...configServices];
    const targetIndex = direction === "up" ? index - 1 : index + 1;

    if (targetIndex >= 0 && targetIndex < newServices.length) {
      // Swap the services
      [newServices[index], newServices[targetIndex]] = [
        newServices[targetIndex],
        newServices[index],
      ];

      // Update orders
      newServices.forEach((service, idx) => {
        service.order = idx + 1;
      });

      setConfigServices(newServices);
    }
  };

  const toggleService = (index: number) => {
    const newServices = [...configServices];
    newServices[index].enabled = !newServices[index].enabled;
    setConfigServices(newServices);
  };

  const handleDragEnd = (result: DropResult) => {
    if (!result.destination) {
      return;
    }

    const items = Array.from(configServices);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);

    // Update orders
    items.forEach((service, idx) => {
      service.order = idx + 1;
    });

    setConfigServices(items);
  };

  const cancelEditing = () => {
    setIsCreating(false);
    setEditingConfig(null);
    setNewConfigName("");
    setConfigServices([]);
  };

  const handleApplyConfig = async (config: Configuration) => {
    try {
      const response = await fetch(`/api/configurations/${config.id}/apply`, {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error(
          `Failed to apply configuration: ${response.status} ${response.statusText}`,
        );
      }

      addToast(
        toast.success(
          "Configuration applied",
          `Successfully applied configuration "${config.name}"`,
        ),
      );
      onConfigurationSaved?.(); // Refresh data
    } catch (error) {
      console.error("Failed to apply configuration:", error);
      addToast(
        toast.error(
          "Failed to apply configuration",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    }
  };

  const handleDeleteConfig = async (config: Configuration) => {
    const confirmed = await showConfirm({
      title: "Delete Configuration",
      description: `Are you sure you want to delete configuration "${config.name}"? This action cannot be undone.`,
      confirmText: "Delete",
      variant: "destructive",
      icon: <Trash2 className="h-6 w-6" />,
    });

    if (!confirmed) return;

    try {
      const response = await fetch(`/api/configurations/${config.id}`, {
        method: "DELETE",
      });

      if (!response.ok) {
        throw new Error(
          `Failed to delete configuration: ${response.status} ${response.statusText}`,
        );
      }

      addToast(
        toast.success(
          "Configuration deleted",
          `Successfully deleted configuration "${config.name}"`,
        ),
      );
      onConfigurationSaved?.(); // Refresh data
    } catch (error) {
      console.error("Failed to delete configuration:", error);
      addToast(
        toast.error(
          "Failed to delete configuration",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        <ErrorBoundarySection
          title="Configuration Manager Error"
          description="Failed to load the configuration manager."
        >
          <div className="flex items-center justify-between p-6 border-b">
            <h2 className="text-xl font-semibold">Configuration Manager</h2>
            <Button variant="ghost" size="sm" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          </div>

          <div className="p-6">
            <div className="flex items-center justify-between mb-6">
              <p className="text-muted-foreground">
                Manage service configurations and switch between different
                setups
              </p>
              <Button variant="outline" size="sm" onClick={startCreating}>
                <Plus className="h-4 w-4 mr-1" />
                New Configuration
              </Button>
            </div>

            {isCreating && (
              <div className="mb-6 p-6 border rounded-lg bg-blue-50 dark:bg-blue-900/20">
                <div className="space-y-4">
                  <div>
                    <Label htmlFor="config-name">Configuration Name</Label>
                    <Input
                      id="config-name"
                      value={newConfigName}
                      onChange={(e) => setNewConfigName(e.target.value)}
                      placeholder="Enter configuration name"
                    />
                  </div>

                  <div>
                    <Label className="text-base font-medium">
                      Services Configuration
                    </Label>
                    <p className="text-sm text-muted-foreground mb-3">
                      Select and reorder services for this configuration. You
                      can drag services or use the arrow buttons to reorder
                      them.
                    </p>

                    <div className="max-h-64 overflow-y-auto">
                      <DragDropContext onDragEnd={handleDragEnd}>
                        <Droppable droppableId="services">
                          {(provided) => (
                            <div
                              {...provided.droppableProps}
                              ref={provided.innerRef}
                              className="space-y-2"
                            >
                              {configServices.map((service, index) => (
                                <Draggable
                                  key={service.id}
                                  draggableId={service.id}
                                  index={index}
                                >
                                  {(provided, snapshot) => (
                                    <div
                                      ref={provided.innerRef}
                                      {...provided.draggableProps}
                                      className={`flex items-center gap-3 p-3 border rounded-lg ${
                                        service.enabled
                                          ? "bg-white"
                                          : "bg-gray-50 opacity-60"
                                      } ${
                                        snapshot.isDragging ? "shadow-lg" : ""
                                      }`}
                                    >
                                      <Checkbox
                                        checked={service.enabled}
                                        onCheckedChange={() =>
                                          toggleService(index)
                                        }
                                      />

                                      <div className="flex items-center gap-2">
                                        <div {...provided.dragHandleProps}>
                                          <GripVertical className="h-4 w-4 text-gray-400 cursor-grab" />
                                        </div>
                                        <span className="text-xs bg-blue-100 text-blue-700 px-2 py-1 rounded">
                                          #{service.order}
                                        </span>
                                      </div>

                                      <span className="flex-1 font-medium">
                                        {service.name}
                                      </span>

                                      <div className="flex gap-1">
                                        <Button
                                          variant="ghost"
                                          size="sm"
                                          onClick={() =>
                                            moveService(index, "up")
                                          }
                                          disabled={index === 0}
                                        >
                                          <ArrowUp className="h-3 w-3" />
                                        </Button>
                                        <Button
                                          variant="ghost"
                                          size="sm"
                                          onClick={() =>
                                            moveService(index, "down")
                                          }
                                          disabled={
                                            index === configServices.length - 1
                                          }
                                        >
                                          <ArrowDown className="h-3 w-3" />
                                        </Button>
                                      </div>
                                    </div>
                                  )}
                                </Draggable>
                              ))}
                              {provided.placeholder}
                            </div>
                          )}
                        </Droppable>
                      </DragDropContext>
                    </div>
                  </div>

                  <div className="flex justify-end gap-2">
                    <Button variant="outline" size="sm" onClick={cancelEditing}>
                      Cancel
                    </Button>
                    <Button
                      size="sm"
                      onClick={handleSaveConfig}
                      disabled={isSaving}
                    >
                      <ButtonSpinner
                        isLoading={isSaving}
                        loadingText={
                          editingConfig ? "Updating..." : "Creating..."
                        }
                      >
                        {editingConfig
                          ? "Update Configuration"
                          : "Create Configuration"}
                      </ButtonSpinner>
                    </Button>
                  </div>
                </div>
              </div>
            )}

            <div className="grid gap-4">
              {configurations.map((config) => (
                <div
                  key={config.id}
                  className={`p-4 border rounded-lg transition-colors ${
                    config.isDefault
                      ? "border-blue-500 bg-blue-50 dark:bg-blue-900/20"
                      : "border-gray-200 dark:border-gray-700 hover:border-gray-300"
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <h3 className="font-medium">{config.name}</h3>
                      {config.isDefault && (
                        <Badge variant="default">Active</Badge>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => startEditing(config)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      {!config.isDefault && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleApplyConfig(config)}
                        >
                          <Play className="h-4 w-4 mr-1" />
                          Apply
                        </Button>
                      )}
                      {!config.isDefault && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDeleteConfig(config)}
                          className="text-red-600 hover:text-red-700"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      )}
                    </div>
                  </div>

                  <div className="mt-3">
                    <p className="text-sm text-muted-foreground">
                      Services: {config.services.length} configured
                    </p>
                    <div className="flex flex-wrap gap-1 mt-2">
                      {config.services.slice(0, 5).map((service) => (
                        <Badge
                          key={service.id}
                          variant="outline"
                          className="text-xs"
                        >
                          {service.name}
                        </Badge>
                      ))}
                      {config.services.length > 5 && (
                        <Badge variant="outline" className="text-xs">
                          +{config.services.length - 5} more
                        </Badge>
                      )}
                    </div>
                  </div>
                </div>
              ))}

              {configurations.length === 0 && (
                <p className="text-muted-foreground text-center py-8">
                  No configurations available
                </p>
              )}
            </div>
          </div>

          <div className="flex justify-end gap-3 p-6 border-t">
            <Button variant="outline" onClick={onClose}>
              Close
            </Button>
          </div>
        </ErrorBoundarySection>
      </div>
    </div>
  );
}
