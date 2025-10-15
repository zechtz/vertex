import { useState, useEffect } from "react";
import {
  X,
  Plus,
  Trash2,
  Save,
  RefreshCw,
  FileText,
  Upload,
  Settings,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox";
import { useToast, toast } from "@/components/ui/toast";
import { ButtonSpinner } from "@/components/ui/spinner";
import { ErrorBoundarySection } from "@/components/ui/error-boundary";
import { BulkImportModal } from "@/components/EnvironmentVariables/BulkImportModal";

interface ServiceEnvVar {
  name: string;
  value: string;
  description?: string;
  isRequired?: boolean;
}

interface ServiceEnvModalProps {
  isOpen: boolean;
  onClose: () => void;
  serviceName: string;
  serviceId: string;
}

export function ServiceEnvModal({
  isOpen,
  onClose,
  serviceName,
  serviceId,
}: ServiceEnvModalProps) {
  const [envVars, setEnvVars] = useState<ServiceEnvVar[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [bulkImportText, setBulkImportText] = useState("");
  const [showBulkImport, setShowBulkImport] = useState(false);
  const [isAdvancedImportOpen, setIsAdvancedImportOpen] = useState(false);

  const { addToast } = useToast();

  useEffect(() => {
    if (isOpen && serviceName && serviceId) {
      fetchEnvVars();
    }
  }, [isOpen, serviceName, serviceId]);

  const fetchEnvVars = async () => {
    try {
      setIsLoading(true);
      const response = await fetch(`/api/services/${serviceId}/env-vars`);
      if (!response.ok) {
        throw new Error(
          `Failed to fetch environment variables: ${response.status} ${response.statusText}`,
        );
      }
      const data = await response.json();
      // Convert object to array format
      const vars = Object.entries(data.envVars || {}).map(
        ([key, envVar]: [string, any]) => ({
          name: key,
          value: envVar.value || "",
          description: envVar.description || "",
          isRequired: envVar.isRequired || false,
        }),
      );
      setEnvVars(vars);
    } catch (error) {
      console.error("Failed to fetch service env vars:", error);
      addToast(
        toast.error(
          "Failed to load environment variables",
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
      setIsLoading(true);
      const envVarsObject = envVars.reduce(
        (acc, envVar) => {
          if (envVar.name.trim()) {
            acc[envVar.name] = {
              name: envVar.name,
              value: envVar.value,
              description: envVar.description || "",
              isRequired: envVar.isRequired || false,
            };
          }
          return acc;
        },
        {} as Record<string, ServiceEnvVar>,
      );

      const response = await fetch(`/api/services/${serviceId}/env-vars`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ envVars: envVarsObject }),
      });

      if (!response.ok) {
        throw new Error(
          `Failed to save environment variables: ${response.status} ${response.statusText}`,
        );
      }

      // Refresh the data to show the saved changes
      await fetchEnvVars();
      addToast(
        toast.success(
          "Environment variables saved",
          `Successfully updated ${Object.keys(envVarsObject).length} environment variables for ${serviceName}`,
        ),
      );
      // Close the modal after successful save
      onClose();
    } catch (error) {
      console.error("Failed to save service env vars:", error);
      addToast(
        toast.error(
          "Failed to save environment variables",
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
        ),
      );
    } finally {
      setIsLoading(false);
    }
  };

  const addEnvVar = () => {
    setEnvVars([
      ...envVars,
      { name: "", value: "", description: "", isRequired: false },
    ]);
  };

  const removeEnvVar = (index: number) => {
    setEnvVars(envVars.filter((_, i) => i !== index));
  };

  const updateEnvVar = (
    index: number,
    field: keyof ServiceEnvVar,
    value: string | boolean,
  ) => {
    const updated = [...envVars];
    updated[index] = { ...updated[index], [field]: value };
    setEnvVars(updated);
  };

  const parseBulkImport = (text: string): ServiceEnvVar[] => {
    const variables: ServiceEnvVar[] = [];
    const pairs = text.split(";").filter((pair) => pair.trim());

    pairs.forEach((pair) => {
      const [name, ...valueParts] = pair.split("=");
      if (name && valueParts.length > 0) {
        const value = valueParts.join("="); // Handle values that contain '='
        variables.push({
          name: name.trim(),
          value: value.trim(),
          description: "",
          isRequired: false,
        });
      }
    });

    return variables;
  };

  const handleBulkImport = () => {
    if (!bulkImportText.trim()) {
      addToast(
        toast.warning(
          "No data to import",
          "Please enter environment variables to import",
        ),
      );
      return;
    }

    try {
      const newVars = parseBulkImport(bulkImportText);
      if (newVars.length === 0) {
        addToast(
          toast.error(
            "No valid variables found",
            "Please check the format and try again",
          ),
        );
        return;
      }

      // Merge with existing variables, with imported ones taking precedence
      const mergedVars = [...envVars];

      newVars.forEach((newVar) => {
        const existingIndex = mergedVars.findIndex(
          (v) => v.name === newVar.name,
        );
        if (existingIndex >= 0) {
          // Update existing variable
          mergedVars[existingIndex] = {
            ...mergedVars[existingIndex],
            ...newVar,
          };
        } else {
          // Add new variable
          mergedVars.push(newVar);
        }
      });

      setEnvVars(mergedVars);
      setBulkImportText("");
      setShowBulkImport(false);
      addToast(
        toast.success(
          "Variables imported",
          `Successfully imported ${newVars.length} environment variables`,
        ),
      );
    } catch (error) {
      console.error("Error parsing bulk import:", error);
      addToast(
        toast.error(
          "Import failed",
          "Error parsing environment variables. Please check the format.",
        ),
      );
    }
  };

  const handleAdvancedImport = (variables: Record<string, string>) => {
    // Convert to ServiceEnvVar format
    const newVars: ServiceEnvVar[] = Object.entries(variables).map(
      ([name, value]) => ({
        name,
        value,
        description: "",
        isRequired: false,
      }),
    );

    // Merge with existing variables
    const mergedVars = [...envVars];
    let updatedCount = 0;
    let addedCount = 0;

    newVars.forEach((newVar) => {
      const existingIndex = mergedVars.findIndex((v) => v.name === newVar.name);
      if (existingIndex >= 0) {
        // Update existing variable
        mergedVars[existingIndex] = {
          ...mergedVars[existingIndex],
          value: newVar.value,
        };
        updatedCount++;
      } else {
        // Add new variable
        mergedVars.push(newVar);
        addedCount++;
      }
    });

    setEnvVars(mergedVars);
    addToast(
      toast.success(
        "Variables imported",
        `Added ${addedCount} new variables${updatedCount > 0 ? `, updated ${updatedCount} existing variables` : ""}`,
      ),
    );

    setIsAdvancedImportOpen(false);
  };

  const exportToText = async () => {
    const validVars = envVars.filter((v) => v.name.trim());
    const exportText = validVars.map((v) => `${v.name}=${v.value}`).join(";");

    try {
      await navigator.clipboard.writeText(exportText);
      addToast(
        toast.success(
          "Variables exported",
          `Copied ${validVars.length} environment variables to clipboard`,
        ),
      );
    } catch (error) {
      // Fallback for older browsers
      try {
        const textArea = document.createElement("textarea");
        textArea.value = exportText;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand("copy");
        document.body.removeChild(textArea);
        addToast(
          toast.success(
            "Variables exported",
            `Copied ${validVars.length} environment variables to clipboard`,
          ),
        );
      } catch (fallbackError) {
        addToast(
          toast.error(
            "Export failed",
            "Unable to copy to clipboard. Please try again.",
          ),
        );
      }
    }
  };

  // Filter variables based on search
  const filteredEnvVars = envVars.filter(
    (envVar) =>
      envVar.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      envVar.value.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (envVar.description &&
        envVar.description.toLowerCase().includes(searchTerm.toLowerCase())),
  );

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        <ErrorBoundarySection
          title="Environment Variables Error"
          description="Failed to load the environment variables form."
        >
          <div className="flex items-center justify-between p-6 border-b">
            <h2 className="text-xl font-semibold">
              {serviceName} - Environment Variables
            </h2>
            <Button variant="ghost" size="sm" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          </div>

          <div className="p-6">
            {/* Header with actions */}
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="font-medium text-gray-900 dark:text-gray-100">
                  Service Environment Variables ({envVars.length} total)
                </h3>
                <p className="text-sm text-muted-foreground">
                  Manage environment variables specific to the {serviceName}{" "}
                  service
                </p>
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={fetchEnvVars}
                  disabled={isLoading}
                >
                  <RefreshCw className="h-4 w-4 mr-1" />
                  Refresh
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowBulkImport(!showBulkImport)}
                >
                  <Upload className="h-4 w-4 mr-1" />
                  Quick Import
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setIsAdvancedImportOpen(true)}
                >
                  <Settings className="h-4 w-4 mr-1" />
                  Advanced Import
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={exportToText}
                  disabled={envVars.length === 0}
                >
                  <FileText className="h-4 w-4 mr-1" />
                  Export
                </Button>
                <Button variant="outline" size="sm" onClick={addEnvVar}>
                  <Plus className="h-4 w-4 mr-1" />
                  Add Variable
                </Button>
              </div>
            </div>

            {/* Search */}
            <div className="mb-4">
              <Label htmlFor="search">Search Variables</Label>
              <Input
                id="search"
                placeholder="Search by name, value, or description..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>

            {/* Quick Import Section (Legacy semicolon format) */}
            {showBulkImport && (
              <div className="mb-4 p-4 border rounded-lg bg-blue-50 dark:bg-blue-900/20">
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="bulk-import">
                      Quick Import - Semicolon Format
                    </Label>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setShowBulkImport(false)}
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                  <p className="text-sm text-blue-700 dark:text-blue-300">
                    💡 <strong>Quick Format:</strong> Paste semicolon-separated
                    environment variables (e.g.,
                    VAR1=value1;VAR2=value2;VAR3=value3)
                  </p>
                  <p className="text-xs text-blue-600 dark:text-blue-400">
                    For multiple formats (.env, JSON, YAML, etc.), use the
                    "Advanced Import" button above.
                  </p>
                  <Textarea
                    id="bulk-import"
                    placeholder="ACTIVE_PROFILE=dev;APM_ENABLED=false;CONFIG_PASSWORD=1kzwjz2nzegt3app@ppra.go.tza1q@BmM0Oo;CONFIG_SERVER=app-config-server;CONFIG_USERNAME=app;cors.allowed.methods=*;cors.allowed.origins=*;SERVICE_PORT=8802"
                    value={bulkImportText}
                    onChange={(e) => setBulkImportText(e.target.value)}
                    rows={4}
                    className="font-mono text-sm"
                  />
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      onClick={handleBulkImport}
                      disabled={!bulkImportText.trim()}
                    >
                      <Upload className="h-4 w-4 mr-1" />
                      Quick Import
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setBulkImportText("")}
                    >
                      Clear
                    </Button>
                  </div>
                </div>
              </div>
            )}

            {isLoading ? (
              <p className="text-center text-muted-foreground py-8">
                Loading...
              </p>
            ) : (
              <div className="space-y-4 max-h-96 overflow-y-auto">
                {filteredEnvVars.map((envVar) => {
                  const actualIndex = envVars.findIndex((v) => v === envVar);
                  return (
                    <div
                      key={actualIndex}
                      className="p-4 border rounded-lg space-y-3"
                    >
                      <div className="flex items-center gap-3">
                        <div className="flex-1">
                          <Label htmlFor={`name-${actualIndex}`}>
                            Variable Name
                          </Label>
                          <Input
                            id={`name-${actualIndex}`}
                            value={envVar.name}
                            onChange={(e) =>
                              updateEnvVar(actualIndex, "name", e.target.value)
                            }
                            placeholder="VARIABLE_NAME"
                          />
                        </div>
                        <div className="flex-1">
                          <Label htmlFor={`value-${actualIndex}`}>Value</Label>
                          <Textarea
                            id={`value-${actualIndex}`}
                            value={envVar.value}
                            onChange={(e) =>
                              updateEnvVar(actualIndex, "value", e.target.value)
                            }
                            placeholder="Variable value"
                            rows={1}
                            className="resize-none"
                          />
                        </div>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => removeEnvVar(actualIndex)}
                          className="mt-6"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                      <div className="flex items-center gap-3">
                        <div className="flex-1">
                          <Label htmlFor={`desc-${actualIndex}`}>
                            Description
                          </Label>
                          <Input
                            id={`desc-${actualIndex}`}
                            value={envVar.description || ""}
                            onChange={(e) =>
                              updateEnvVar(
                                actualIndex,
                                "description",
                                e.target.value,
                              )
                            }
                            placeholder="Variable description"
                          />
                        </div>
                        <div className="flex items-center space-x-2 mt-6">
                          <Checkbox
                            id={`required-${actualIndex}`}
                            checked={envVar.isRequired || false}
                            onCheckedChange={(checked) =>
                              updateEnvVar(actualIndex, "isRequired", checked)
                            }
                          />
                          <Label htmlFor={`required-${actualIndex}`}>
                            Required
                          </Label>
                        </div>
                      </div>
                    </div>
                  );
                })}
                {filteredEnvVars.length === 0 && searchTerm && (
                  <p className="text-muted-foreground text-center py-8">
                    No variables match your search criteria
                  </p>
                )}
                {envVars.length === 0 && (
                  <p className="text-muted-foreground text-center py-8">
                    No environment variables defined for this service
                  </p>
                )}
              </div>
            )}
          </div>

          <div className="flex justify-end gap-3 p-6 border-t">
            <Button variant="outline" onClick={onClose} disabled={isLoading}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={isLoading}>
              <ButtonSpinner isLoading={isLoading} loadingText="Saving...">
                <Save className="h-4 w-4 mr-1" />
                Save Changes
              </ButtonSpinner>
            </Button>
          </div>
        </ErrorBoundarySection>
      </div>

      {/* Advanced Import Modal */}
      <BulkImportModal
        isOpen={isAdvancedImportOpen}
        onClose={() => setIsAdvancedImportOpen(false)}
        onImport={handleAdvancedImport}
      />
    </div>
  );
}
