import { useState, useEffect } from "react";
import { X, Plus, Trash2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useToast, toast } from "@/components/ui/toast";
import { ButtonSpinner } from "@/components/ui/spinner";
import { ErrorBoundarySection } from "@/components/ui/error-boundary";

interface GlobalEnvVar {
  name: string;
  value: string;
  description?: string;
  category?: string;
}

interface GlobalEnvModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function GlobalEnvModal({ isOpen, onClose }: GlobalEnvModalProps) {
  const [envVars, setEnvVars] = useState<GlobalEnvVar[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("all");
  
  const { addToast } = useToast();

  useEffect(() => {
    if (isOpen) {
      fetchEnvVars();
    }
  }, [isOpen]);

  const fetchEnvVars = async () => {
    try {
      setIsLoading(true);
      const response = await fetch("/api/env-vars/global");
      if (!response.ok) {
        throw new Error(`Failed to fetch global environment variables: ${response.status} ${response.statusText}`);
      }
      const data = await response.json();
      // Convert object to array format with categorization
      const vars = Object.entries(data.envVars || {}).map(([name, value]) => {
        let description = "";
        let category = "other";
        
        // Categorize based on variable names
        if (name.includes("DB_") || name.includes("DATABASE")) {
          category = "database";
          description = "Database configuration";
        } else if (name.includes("CONFIG_") || name.includes("SPRING_")) {
          category = "config";
          description = "Application configuration";
        } else if (name.includes("CLIENT_") || name.includes("AUTH")) {
          category = "auth";
          description = "Authentication configuration";
        } else if (name.includes("REDIS") || name.includes("RABBIT")) {
          category = "cache";
          description = "Cache/Message queue configuration";
        } else if (name.includes("PORT") || name.includes("URL") || name.includes("URI")) {
          category = "network";
          description = "Network configuration";
        }
        
        return {
          name,
          value: value as string,
          description,
          category
        };
      });
      setEnvVars(vars);
    } catch (error) {
      console.error("Failed to fetch global env vars:", error);
      addToast(toast.error(
        "Failed to load global environment variables",
        error instanceof Error ? error.message : "An unexpected error occurred"
      ));
    } finally {
      setIsLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      setIsLoading(true);
      const envVarsObject = envVars.reduce((acc, envVar) => {
        if (envVar.name.trim()) {
          acc[envVar.name] = envVar.value;
        }
        return acc;
      }, {} as Record<string, string>);

      const response = await fetch("/api/env-vars/global", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ envVars: envVarsObject }),
      });

      if (!response.ok) {
        throw new Error(`Failed to save global environment variables: ${response.status} ${response.statusText}`);
      }

      // Refresh the data to show the saved changes
      await fetchEnvVars();
      addToast(toast.success(
        "Global environment variables saved",
        `Successfully updated ${Object.keys(envVarsObject).length} global environment variables`
      ));
    } catch (error) {
      console.error("Failed to save global env vars:", error);
      addToast(toast.error(
        "Failed to save global environment variables",
        error instanceof Error ? error.message : "An unexpected error occurred"
      ));
    } finally {
      setIsLoading(false);
    }
  };

  const addEnvVar = () => {
    setEnvVars([...envVars, { name: "", value: "", description: "", category: "other" }]);
  };

  const removeEnvVar = (index: number) => {
    setEnvVars(envVars.filter((_, i) => i !== index));
  };

  const updateEnvVar = (index: number, field: keyof GlobalEnvVar, value: string) => {
    const updated = [...envVars];
    updated[index] = { ...updated[index], [field]: value };
    setEnvVars(updated);
  };

  const reloadFromFishFile = async () => {
    try {
      setIsLoading(true);
      // This would trigger a backend reload of the fish file
      const response = await fetch("/api/env-vars/reload", { method: "POST" });
      if (response.ok) {
        await fetchEnvVars();
      }
    } catch (error) {
      console.error("Failed to reload env vars:", error);
    } finally {
      setIsLoading(false);
    }
  };

  // Filter variables based on search and category
  const filteredEnvVars = envVars.filter(envVar => {
    const matchesSearch = envVar.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         envVar.value.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesCategory = selectedCategory === "all" || envVar.category === selectedCategory;
    return matchesSearch && matchesCategory;
  });

  const categories = [
    { value: "all", label: "All Variables", count: envVars.length },
    { value: "database", label: "Database", count: envVars.filter(v => v.category === "database").length },
    { value: "config", label: "Configuration", count: envVars.filter(v => v.category === "config").length },
    { value: "auth", label: "Authentication", count: envVars.filter(v => v.category === "auth").length },
    { value: "network", label: "Network", count: envVars.filter(v => v.category === "network").length },
    { value: "cache", label: "Cache/Queue", count: envVars.filter(v => v.category === "cache").length },
    { value: "other", label: "Other", count: envVars.filter(v => v.category === "other").length },
  ];

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-3xl max-h-[90vh] overflow-y-auto">
        <ErrorBoundarySection title="Global Environment Variables Error" description="Failed to load the global environment variables form.">
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">Global Environment Variables</h2>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        <div className="p-6">
          {/* Header with actions */}
          <div className="flex items-center justify-between mb-4">
            <div>
              <h3 className="font-medium text-gray-900 dark:text-gray-100">
                Environment Variables ({envVars.length} total)
              </h3>
              <p className="text-sm text-muted-foreground">
                Manage global environment variables for all services
              </p>
            </div>
            <div className="flex gap-2">
              <Button 
                variant="outline" 
                size="sm" 
                onClick={reloadFromFishFile}
                disabled={isLoading}
              >
                <RefreshCw className="h-4 w-4 mr-1" />
                Reload
              </Button>
              <Button variant="outline" size="sm" onClick={addEnvVar}>
                <Plus className="h-4 w-4 mr-1" />
                Add Variable
              </Button>
            </div>
          </div>

          {/* Search and Filter */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div>
              <Label htmlFor="search">Search Variables</Label>
              <Input
                id="search"
                placeholder="Search by name or value..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor="category">Category</Label>
              <select
                id="category"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                value={selectedCategory}
                onChange={(e) => setSelectedCategory(e.target.value)}
              >
                {categories.map(cat => (
                  <option key={cat.value} value={cat.value}>
                    {cat.label} ({cat.count})
                  </option>
                ))}
              </select>
            </div>
          </div>

          {isLoading ? (
            <p className="text-center text-muted-foreground py-8">Loading...</p>
          ) : (
            <div className="space-y-4 max-h-96 overflow-y-auto">
              {filteredEnvVars.map((envVar) => {
                const actualIndex = envVars.findIndex(v => v.name === envVar.name && v.value === envVar.value);
                return (
                <div key={actualIndex} className="p-4 border rounded-lg space-y-3">
                  <div className="flex items-center gap-3">
                    <div className="flex-1">
                      <Label htmlFor={`name-${actualIndex}`}>Variable Name</Label>
                      <Input
                        id={`name-${actualIndex}`}
                        value={envVar.name}
                        onChange={(e) => updateEnvVar(actualIndex, "name", e.target.value)}
                        placeholder="VARIABLE_NAME"
                      />
                    </div>
                    <div className="flex-1">
                      <Label htmlFor={`value-${actualIndex}`}>Value</Label>
                      <Textarea
                        id={`value-${actualIndex}`}
                        value={envVar.value}
                        onChange={(e) => updateEnvVar(actualIndex, "value", e.target.value)}
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
                  <div>
                    <Label htmlFor={`desc-${actualIndex}`}>Description</Label>
                    <Input
                      id={`desc-${actualIndex}`}
                      value={envVar.description || ""}
                      onChange={(e) => updateEnvVar(actualIndex, "description", e.target.value)}
                      placeholder="Variable description"
                    />
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
                  No global environment variables defined
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
              Save Changes
            </ButtonSpinner>
          </Button>
        </div>
        </ErrorBoundarySection>
      </div>
    </div>
  );
}