import { useState, useEffect } from 'react';
import {
  Settings,
  GripVertical,
  Server,
  CheckCircle,
  Loader2,
  ArrowUp,
  ArrowDown
} from 'lucide-react';
import {
  DragDropContext,
  Droppable,
  Draggable,
  DropResult,
} from "react-beautiful-dnd";
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import { useToast, toast } from '@/components/ui/toast';
import { Service } from '@/types';

interface OnboardingConfigurationStepProps {
  services: Service[];
  onConfigurationComplete: () => void;
  isProcessing: boolean;
  setIsProcessing: (processing: boolean) => void;
}

interface ConfigService {
  id: string;
  name: string;
  order: number;
  enabled: boolean;
}

export function OnboardingConfigurationStep({
  services,
  onConfigurationComplete,
  isProcessing,
  setIsProcessing
}: OnboardingConfigurationStepProps) {
  const { addToast } = useToast();
  const [configServices, setConfigServices] = useState<ConfigService[]>([]);
  const [configName] = useState("Onboarding Configuration");

  // Initialize service configuration
  useEffect(() => {
    const servicesWithConfig = services
      .map((service, index) => ({
        id: service.id,
        name: service.name,
        order: index + 1,
        enabled: true,
      }))
      .sort((a, b) => a.order - b.order);

    setConfigServices(servicesWithConfig);
  }, [services]);

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

  const handleSaveConfiguration = async () => {
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
      id: `onboarding-config-${Date.now()}`,
      name: configName,
      services: enabledServices,
      isDefault: true, // Make this the default configuration
    };

    try {
      setIsProcessing(true);
      const response = await fetch("/api/configurations", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(configData),
      });

      if (!response.ok) {
        throw new Error(
          `Failed to save configuration: ${response.status} ${response.statusText}`,
        );
      }

      const result = await response.json();

      // Apply the configuration immediately after creating it
      const applyResponse = await fetch(`/api/configurations/${result.id}/apply`, {
        method: "POST",
      });

      if (!applyResponse.ok) {
        throw new Error(
          `Failed to apply configuration: ${applyResponse.status} ${applyResponse.statusText}`,
        );
      }

      addToast(
        toast.success(
          "Configuration Applied!",
          `Successfully created and applied "${configName}" with ${enabledServices.length} services`,
        ),
      );
      
      onConfigurationComplete();
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
      setIsProcessing(false);
    }
  };

  const handleSkipConfiguration = () => {
    addToast(
      toast.info(
        'Configuration Skipped',
        'You can configure service order and settings later from the configurations page.'
      )
    );
    onConfigurationComplete();
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center">
        <div className="mx-auto w-16 h-16 bg-orange-100 dark:bg-orange-900/20 rounded-full flex items-center justify-center mb-4">
          <Settings className="w-8 h-8 text-orange-600 dark:text-orange-400" />
        </div>
        <h3 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-2">
          Configure Service Order
        </h3>
        <p className="text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
          Set the startup order for your services. Drag services to reorder them or use the arrow buttons.
          Services higher in the list will start first.
        </p>
      </div>

      {services.length === 0 ? (
        /* No Services */
        <Card className="max-w-2xl mx-auto">
          <CardContent className="p-12 text-center">
            <Server className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <h4 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
              No Services to Configure
            </h4>
            <p className="text-gray-600 dark:text-gray-400 mb-6">
              No services were imported in the previous step. You can add services manually later.
            </p>
            <Button onClick={handleSkipConfiguration} className="bg-green-600 hover:bg-green-700">
              <CheckCircle className="w-4 h-4 mr-2" />
              Complete Setup
            </Button>
          </CardContent>
        </Card>
      ) : (
        /* Service Configuration */
        <>
          <Card className="max-w-4xl mx-auto">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Settings className="w-5 h-5 text-orange-500" />
                Services Configuration ({configServices.length} services)
              </CardTitle>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Select and reorder services for this configuration. You can drag services or use the arrow buttons to reorder them.
              </p>
            </CardHeader>
            <CardContent>
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
                                    ? "bg-white dark:bg-gray-800"
                                    : "bg-gray-50 dark:bg-gray-700 opacity-60"
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
            </CardContent>
          </Card>

          {/* Actions */}
          <div className="flex justify-center gap-3 max-w-2xl mx-auto">
            <Button
              onClick={handleSkipConfiguration}
              variant="outline"
              disabled={isProcessing}
            >
              Use Default Order
            </Button>
            <Button
              onClick={handleSaveConfiguration}
              disabled={isProcessing}
              className="bg-green-600 hover:bg-green-700 min-w-[140px]"
            >
              {isProcessing ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Applying Configuration...
                </>
              ) : (
                <>
                  <CheckCircle className="w-4 h-4 mr-2" />
                  Create & Apply Configuration
                </>
              )}
            </Button>
          </div>
        </>
      )}

      {/* Tips */}
      <Card className="max-w-2xl mx-auto bg-gray-50 dark:bg-gray-800">
        <CardContent className="p-4">
          <h4 className="font-medium text-gray-900 dark:text-gray-100 mb-2">
            ⚙️ Configuration Tips
          </h4>
          <ul className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
            <li>• <strong>Service Registry (Eureka)</strong> should typically start first</li>
            <li>• <strong>Config Server</strong> should start early, before other services</li>
            <li>• <strong>API Gateway</strong> usually starts after core services are ready</li>
            <li>• You can uncheck services you don't want in this configuration</li>
            <li>• This will create and apply a default configuration for your profile</li>
            <li>• The configuration will become active and control service startup order</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}

export default OnboardingConfigurationStep;