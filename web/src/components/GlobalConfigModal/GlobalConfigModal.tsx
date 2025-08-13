import { useState, useEffect } from "react";
import { X, Save, Folder, Coffee, RefreshCw, Star, HelpCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { useToast, toast } from "@/components/ui/toast";
import { ButtonSpinner } from "@/components/ui/spinner";
import { ErrorBoundarySection } from "@/components/ui/error-boundary";

interface GlobalConfig {
  projectsDir: string;
  javaHomeOverride: string;
  lastUpdated?: string;
}

interface GlobalConfigModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfigUpdated?: () => void;
  onboarding?: any;
}

export function GlobalConfigModal({
  isOpen,
  onClose,
  onConfigUpdated,
  onboarding,
}: GlobalConfigModalProps) {
  const [config, setConfig] = useState<GlobalConfig>({
    projectsDir: "",
    javaHomeOverride: "",
  });
  const [originalConfig, setOriginalConfig] = useState<GlobalConfig>({
    projectsDir: "",
    javaHomeOverride: "",
  });
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);

  const { addToast } = useToast();

  useEffect(() => {
    if (isOpen) {
      fetchGlobalConfig();
    }
  }, [isOpen]);

  useEffect(() => {
    const changed =
      config.projectsDir !== originalConfig.projectsDir ||
      config.javaHomeOverride !== originalConfig.javaHomeOverride;
    setHasChanges(changed);
  }, [config, originalConfig]);

  const fetchGlobalConfig = async () => {
    try {
      setIsLoading(true);
      const response = await fetch("/api/config/global");
      if (!response.ok) {
        throw new Error(
          `Failed to fetch global configuration: ${response.status} ${response.statusText}`,
        );
      }
      const data = await response.json();
      setConfig(data);
      setOriginalConfig(data);
    } catch (error) {
      console.error("Failed to fetch global config:", error);
      addToast(
        toast.error(
          "Failed to load global configuration",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    } finally {
      setIsLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      setIsSaving(true);
      const response = await fetch("/api/config/global", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(config),
      });

      if (!response.ok) {
        throw new Error(
          `Failed to save global configuration: ${response.status} ${response.statusText}`,
        );
      }

      const updated = await response.json();
      setOriginalConfig(config);
      setConfig(updated);
      setHasChanges(false);
      addToast(
        toast.success(
          "Global configuration saved",
          "Successfully updated global application settings",
        ),
      );
      onConfigUpdated?.();
    } catch (error) {
      console.error("Failed to save global config:", error);
      addToast(
        toast.error(
          "Failed to save global configuration",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    } finally {
      setIsSaving(false);
    }
  };

  const handleReset = () => {
    setConfig(originalConfig);
    setHasChanges(false);
  };

  const selectDirectory = async (field: "projectsDir" | "javaHomeOverride") => {
    // For now, show a helpful message
    addToast(
      toast.info(
        "Directory Selection",
        `Please manually enter the path for ${field === "projectsDir" ? "Projects Directory" : "Java Home"} in the input field below`,
      ),
    );
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <ErrorBoundarySection
          title="Global Configuration Error"
          description="Failed to load the global configuration form."
        >
          <div className="flex items-center justify-between p-6 border-b">
            <div>
              <h2 className="text-xl font-semibold">
                Global Application Configuration
              </h2>
              <p className="text-sm text-muted-foreground">
                Configure global settings for the Vertex Service Manager
              </p>
            </div>
            <Button variant="ghost" size="sm" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          </div>

          <div className="p-6">
            {isLoading ? (
              <div className="text-center py-8">
                <RefreshCw className="h-8 w-8 mx-auto mb-4 animate-spin text-muted-foreground" />
                <p className="text-muted-foreground">
                  Loading configuration...
                </p>
              </div>
            ) : (
              <div className="space-y-6">
                {/* Projects Directory */}
                <div className="space-y-2">
                  <Label
                    htmlFor="projectsDir"
                    className="text-base font-medium"
                  >
                    Projects Directory
                  </Label>
                  <p className="text-sm text-muted-foreground">
                    Root directory where all microservice projects are located
                  </p>
                  <div className="flex gap-2">
                    <Input
                      id="projectsDir"
                      value={config.projectsDir}
                      onChange={(e) =>
                        setConfig((prev) => ({
                          ...prev,
                          projectsDir: e.target.value,
                        }))
                      }
                      placeholder="/path/to/projects"
                      className="flex-1"
                    />
                    <Button
                      variant="outline"
                      onClick={() => selectDirectory("projectsDir")}
                    >
                      <Folder className="h-4 w-4" />
                    </Button>
                  </div>
                  <div className="text-xs text-muted-foreground">
                    Current:{" "}
                    <code className="bg-muted px-1 rounded">
                      {config.projectsDir || "Not set"}
                    </code>
                  </div>
                </div>

                {/* Java Home Override */}
                <div className="space-y-2">
                  <Label htmlFor="javaHome" className="text-base font-medium">
                    Java Home Override
                  </Label>
                  <p className="text-sm text-muted-foreground">
                    Custom Java installation path (optional). Leave empty to use
                    system default.
                  </p>
                  <div className="flex gap-2">
                    <Input
                      id="javaHome"
                      value={config.javaHomeOverride}
                      onChange={(e) =>
                        setConfig((prev) => ({
                          ...prev,
                          javaHomeOverride: e.target.value,
                        }))
                      }
                      placeholder="/path/to/java/home"
                      className="flex-1"
                    />
                    <Button
                      variant="outline"
                      onClick={() => selectDirectory("javaHomeOverride")}
                    >
                      <Coffee className="h-4 w-4" />
                    </Button>
                  </div>
                  <div className="text-xs text-muted-foreground">
                    Current:{" "}
                    <code className="bg-muted px-1 rounded">
                      {config.javaHomeOverride || "System default"}
                    </code>
                  </div>
                </div>

                {/* Onboarding Section */}
                {onboarding && (
                  <div className="space-y-2">
                    <Label className="text-base font-medium">Getting Started</Label>
                    <p className="text-sm text-muted-foreground">
                      Need help setting up your workspace? Run the setup wizard again.
                    </p>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        onClick={() => {
                          onboarding.forceShowOnboarding();
                          onClose();
                        }}
                        className="flex items-center gap-2"
                      >
                        <Star className="h-4 w-4" />
                        Run Setup Wizard
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          onboarding.resetOnboarding();
                        }}
                        className="flex items-center gap-2"
                      >
                        <HelpCircle className="h-4 w-4" />
                        Reset Onboarding
                      </Button>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      The setup wizard will help you create profiles, discover services, and configure startup order.
                    </p>
                  </div>
                )}

                {/* Configuration Status */}
                <div className="p-4 bg-muted/50 rounded-lg">
                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium">Configuration Status</h4>
                      <p className="text-sm text-muted-foreground">
                        {hasChanges
                          ? "You have unsaved changes"
                          : "Configuration is up to date"}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      {hasChanges && (
                        <Badge
                          variant="outline"
                          className="bg-yellow-50 text-yellow-700 border-yellow-200"
                        >
                          Modified
                        </Badge>
                      )}
                      {config.lastUpdated && (
                        <Badge variant="outline">
                          Last updated:{" "}
                          {new Date(config.lastUpdated).toLocaleDateString()}
                        </Badge>
                      )}
                    </div>
                  </div>
                </div>

                {/* Service Impact Warning */}
                {hasChanges && (
                  <div className="p-4 bg-orange-50 dark:bg-orange-900/20 border border-orange-200 dark:border-orange-800 rounded-lg">
                    <div className="flex items-start gap-3">
                      <div className="flex-shrink-0">
                        <div className="w-2 h-2 bg-orange-500 rounded-full mt-2"></div>
                      </div>
                      <div>
                        <h4 className="font-medium text-orange-800 dark:text-orange-200">
                          Configuration Changes Impact
                        </h4>
                        <p className="text-sm text-orange-700 dark:text-orange-300 mt-1">
                          Changing these settings may affect how services are
                          started and managed. Consider restarting running
                          services after saving changes.
                        </p>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          <div className="flex justify-end gap-3 p-6 border-t">
            {hasChanges && (
              <Button
                variant="outline"
                onClick={handleReset}
                disabled={isSaving}
              >
                Reset
              </Button>
            )}
            <Button variant="outline" onClick={onClose} disabled={isSaving}>
              Close
            </Button>
            <Button
              onClick={handleSave}
              disabled={!hasChanges || isSaving || isLoading}
            >
              <ButtonSpinner isLoading={isSaving} loadingText="Saving...">
                <Save className="h-4 w-4 mr-2" />
                Save Configuration
              </ButtonSpinner>
            </Button>
          </div>
        </ErrorBoundarySection>
      </div>
    </div>
  );
}
