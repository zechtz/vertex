import { useState, useEffect } from "react";
import React from "react";
import {
  Download,
  Eye,
  Settings,
  Container,
  AlertCircle,
  CheckCircle,
  Loader2,
  Copy,
  Check,
  RefreshCw,
  Play,
  X,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Modal } from "@/components/ui/Modal";
import { useToast } from "@/components/ui/toast";
import { ServiceProfile } from "@/types";
import {
  dockerComposeApi,
  DockerComposePreview,
  DockerComposeGeneration,
  DockerComposeRequest,
} from "@/services/dockerComposeApi";

interface DockerComposeModalProps {
  isOpen: boolean;
  onClose: () => void;
  profile: ServiceProfile | null;
}

class ErrorBoundary extends React.Component<
  { children: React.ReactNode; onError?: () => void },
  { hasError: boolean }
> {
  constructor(props: { children: React.ReactNode; onError?: () => void }) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(): { hasError: boolean } {
    return { hasError: true };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error("Docker Compose Modal Error:", error, errorInfo);
    this.props.onError?.();
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex items-center justify-center py-16">
          <div className="text-center">
            <AlertCircle className="h-12 w-12 text-red-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              Something went wrong
            </h3>
            <p className="text-gray-500 mb-4">
              There was an error loading the Docker Compose interface.
            </p>
            <Button
              onClick={() => {
                this.setState({ hasError: false });
                window.location.reload();
              }}
              variant="outline"
            >
              Reload Page
            </Button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export function DockerComposeModal({
  isOpen,
  onClose,
  profile,
}: DockerComposeModalProps) {
  const { addToast } = useToast();
  const [activeTab, setActiveTab] = useState<"preview" | "generate" | "config">(
    "preview",
  );
  const [environment, setEnvironment] = useState("development");
  const [includeExternal, setIncludeExternal] = useState(true);
  const [generateOverride, setGenerateOverride] = useState(false);

  const [isLoading, setIsLoading] = useState(false);
  const [preview, setPreview] = useState<DockerComposePreview | null>(null);
  const [generated, setGenerated] = useState<DockerComposeGeneration | null>(
    null,
  );
  const [yamlCopied, setYamlCopied] = useState(false);

  // Reset state when modal opens
  useEffect(() => {
    if (isOpen && profile) {
      setPreview(null);
      setGenerated(null);
      setActiveTab("preview");
      fetchPreview();
    }
  }, [isOpen, profile]);

  // Fetch preview when environment changes
  useEffect(() => {
    if (isOpen && profile && activeTab === "preview") {
      fetchPreview();
    }
  }, [environment]);

  const fetchPreview = async () => {
    if (!profile) return;

    setIsLoading(true);
    try {
      const data = await dockerComposeApi.getPreview(profile.id, environment);
      setPreview(data);
    } catch (error) {
      console.error("Failed to fetch Docker Compose preview:", error);
      addToast({
        variant: "error",
        title: "Preview Error",
        description: "Failed to fetch Docker Compose preview",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const generateCompose = async () => {
    if (!profile) return;

    setIsLoading(true);
    try {
      const request: DockerComposeRequest = {
        environment,
        includeExternal,
        generateOverride,
      };

      const data = await dockerComposeApi.generate(profile.id, request);
      setGenerated(data);
      setActiveTab("generate");

      addToast({
        variant: "success",
        title: "Generated Successfully",
        description: "Docker Compose file generated successfully!",
      });
    } catch (error) {
      console.error("Failed to generate Docker Compose:", error);
      addToast({
        variant: "error",
        title: "Generation Error",
        description:
          error instanceof Error
            ? error.message
            : "Failed to generate Docker Compose",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const downloadCompose = async (includeOverride = false) => {
    if (!profile) return;

    try {
      const blob = includeOverride
        ? await dockerComposeApi.downloadOverride(profile.id)
        : await dockerComposeApi.download(profile.id, {
            environment,
            includeExternal,
          });

      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.style.display = "none";
      a.href = url;

      const filename = includeOverride
        ? `docker-compose.override-${profile.name.toLowerCase().replace(/\s+/g, "-")}.yml`
        : `docker-compose-${profile.name.toLowerCase().replace(/\s+/g, "-")}.yml`;

      a.download = filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      addToast({
        variant: "success",
        title: "File Downloaded",
        description: `${includeOverride ? "Override" : "Docker Compose"} file downloaded successfully!`,
      });
    } catch (error) {
      console.error("Failed to download file:", error);
      addToast({
        variant: "error",
        title: "Download Error",
        description:
          error instanceof Error ? error.message : "Failed to download file",
      });
    }
  };

  const copyYamlToClipboard = async () => {
    if (!generated?.yaml) return;

    try {
      await navigator.clipboard.writeText(generated.yaml);
      setYamlCopied(true);
      setTimeout(() => setYamlCopied(false), 2000);

      addToast({
        variant: "success",
        title: "Copied",
        description: "YAML copied to clipboard!",
      });
    } catch (error) {
      console.error("Failed to copy YAML:", error);
      addToast({
        variant: "error",
        title: "Error",
        description: "Failed to copy YAML to clipboard",
      });
    }
  };

  if (!isOpen || !profile) return null;

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="2xl">
      <ErrorBoundary onError={onClose}>
        <div className="p-6 space-y-8">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <Container className="h-7 w-7 text-blue-600" />
              <h1 className="text-2xl font-bold text-gray-900">
                Docker Manager
              </h1>
            </div>

            <button
              onClick={onClose}
              className="flex items-center justify-center w-8 h-8 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-full transition-colors"
            >
              <X className="h-5 w-5" />
            </button>
          </div>

          {preview && (
            <div className="flex items-center space-x-6 text-sm text-gray-600 bg-gray-50 rounded-lg px-4 py-3">
              <div className="flex items-center space-x-2">
                <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
                <span className="font-medium">
                  {preview.serviceCount || 0} services
                </span>
              </div>
              <div className="flex items-center space-x-2">
                <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                <span className="font-medium">
                  {preview.networkCount || 0} networks
                </span>
              </div>
              <div className="flex items-center space-x-2">
                <div className="w-2 h-2 bg-purple-500 rounded-full"></div>
                <span className="font-medium">
                  {preview.volumeCount || 0} volumes
                </span>
              </div>
            </div>
          )}

          {/* Profile Info Bar */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg px-6 py-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="font-semibold text-blue-900">
                  Active Profile: {profile.name}
                </h3>
                <p className="text-sm text-blue-700 mt-1">
                  {profile.services?.length || 0} services configured •
                  Projects: {profile.projectsDir || "Not set"}
                </p>
              </div>
              <div className="flex space-x-2">
                <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  Services
                </span>
                <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                  Env Vars
                </span>
                <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                  Docker
                </span>
              </div>
            </div>
          </div>

          {/* Quick Configuration */}
          <div className="flex items-center justify-between py-4 px-6 bg-gray-50 border border-gray-200 rounded-lg">
            <div className="flex items-center space-x-6">
              <div className="flex items-center space-x-2">
                <label className="text-sm font-medium text-gray-700">
                  Environment:
                </label>
                <select
                  value={environment}
                  onChange={(e) => setEnvironment(e.target.value)}
                  disabled={isLoading}
                  className="px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-100"
                >
                  <option value="development">Development</option>
                  <option value="staging">Staging</option>
                  <option value="production">Production</option>
                </select>
              </div>

              <div className="flex items-center space-x-4">
                <label className="flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={includeExternal}
                    onChange={(e) => setIncludeExternal(e.target.checked)}
                    disabled={isLoading}
                    className="h-3.5 w-3.5 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                  <span className="ml-1.5 text-sm text-gray-700">
                    External services
                  </span>
                </label>

                <label className="flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={generateOverride}
                    onChange={(e) => setGenerateOverride(e.target.checked)}
                    disabled={isLoading}
                    className="h-3.5 w-3.5 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                  <span className="ml-1.5 text-sm text-gray-700">
                    Override file
                  </span>
                </label>
              </div>
            </div>

            <Button
              onClick={fetchPreview}
              variant="outline"
              size="sm"
              disabled={isLoading}
              className="text-xs"
            >
              <RefreshCw
                className={`h-3 w-3 mr-1.5 ${isLoading ? "animate-spin" : ""}`}
              />
              Refresh
            </Button>
          </div>

          {/* Tab Navigation */}
          <div className="border-b border-gray-200">
            <nav className="-mb-px flex">
              {[
                { id: "preview", label: "Preview", icon: Eye },
                { id: "generate", label: "Generate", icon: Play },
                { id: "config", label: "Settings", icon: Settings },
              ].map(({ id, label, icon: Icon }) => (
                <button
                  key={id}
                  onClick={() => setActiveTab(id as any)}
                  className={`flex items-center space-x-2 py-2.5 px-4 border-b-2 font-medium text-sm transition-all duration-200 ${
                    activeTab === id
                      ? "border-blue-500 text-blue-600 bg-blue-50/50"
                      : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 hover:bg-gray-50"
                  }`}
                >
                  <Icon className="h-4 w-4" />
                  <span>{label}</span>
                </button>
              ))}
            </nav>
          </div>

          {/* Tab Content */}
          <div className="min-h-[400px] px-2 py-8">
            {activeTab === "preview" && (
              <div className="space-y-6">
                {isLoading ? (
                  <div className="flex items-center justify-center py-20">
                    <div className="text-center">
                      <Loader2 className="h-8 w-8 animate-spin text-blue-600 mx-auto mb-3" />
                      <p className="text-gray-600 font-medium">
                        Analyzing services...
                      </p>
                      <p className="text-sm text-gray-500 mt-1">
                        Loading Docker Compose preview
                      </p>
                    </div>
                  </div>
                ) : preview ? (
                  <div className="space-y-6">
                    {/* Services Grid */}
                    {preview &&
                    preview.services &&
                    Array.isArray(preview.services) &&
                    preview.services.length > 0 ? (
                      <div className="grid gap-3">
                        {preview.services
                          .map((service, index) => {
                            // Ensure service is not null and has required properties
                            if (!service) return null;

                            const serviceName =
                              service.name || `Service ${index + 1}`;
                            const servicePorts =
                              service.ports && Array.isArray(service.ports)
                                ? service.ports
                                : [];
                            const serviceDeps =
                              service.dependencies &&
                              Array.isArray(service.dependencies)
                                ? service.dependencies
                                : [];
                            const serviceEnvCount =
                              typeof service.environment === "number"
                                ? service.environment
                                : 0;
                            const serviceVolCount =
                              typeof service.volumes === "number"
                                ? service.volumes
                                : 0;

                            return (
                              <div
                                key={`${serviceName}-${index}`}
                                className="flex items-center justify-between p-4 border border-gray-200 rounded-lg hover:border-blue-200 hover:bg-blue-50/30 transition-all"
                              >
                                <div className="flex items-center space-x-3">
                                  <div className="flex items-center justify-center w-10 h-10 bg-gradient-to-br from-blue-100 to-blue-200 rounded-lg">
                                    <Container className="h-5 w-5 text-blue-700" />
                                  </div>
                                  <div>
                                    <h4 className="font-semibold text-gray-900">
                                      {serviceName}
                                    </h4>
                                    <p className="text-sm text-gray-500">
                                      {service.image
                                        ? `Image: ${service.image}`
                                        : service.buildContext
                                          ? `Build: ${service.buildContext}`
                                          : "Custom configuration"}
                                    </p>
                                  </div>
                                </div>

                                <div className="flex items-center space-x-6">
                                  {servicePorts.length > 0 && (
                                    <div className="text-center">
                                      <div className="text-lg font-bold text-blue-600">
                                        {servicePorts.length}
                                      </div>
                                      <div className="text-xs text-gray-500">
                                        ports
                                      </div>
                                    </div>
                                  )}
                                  {serviceEnvCount > 0 && (
                                    <div className="text-center">
                                      <div className="text-lg font-bold text-green-600">
                                        {serviceEnvCount}
                                      </div>
                                      <div className="text-xs text-gray-500">
                                        env vars
                                      </div>
                                    </div>
                                  )}
                                  {serviceVolCount > 0 && (
                                    <div className="text-center">
                                      <div className="text-lg font-bold text-purple-600">
                                        {serviceVolCount}
                                      </div>
                                      <div className="text-xs text-gray-500">
                                        volumes
                                      </div>
                                    </div>
                                  )}
                                  {serviceDeps.length > 0 && (
                                    <div className="text-center">
                                      <div className="text-lg font-bold text-orange-600">
                                        {serviceDeps.length}
                                      </div>
                                      <div className="text-xs text-gray-500">
                                        deps
                                      </div>
                                    </div>
                                  )}
                                </div>
                              </div>
                            );
                          })
                          .filter(Boolean)}
                      </div>
                    ) : (
                      <div className="text-center py-12">
                        <Container className="h-12 w-12 text-gray-400 mx-auto mb-3" />
                        <p className="text-gray-500">
                          No services found in this profile
                        </p>
                        {!preview && (
                          <p className="text-xs text-gray-400 mt-2">
                            Click "Load Preview" to analyze your services
                          </p>
                        )}
                      </div>
                    )}

                    {/* External Services Notice */}
                    {preview.hasExternalDeps && includeExternal && (
                      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                        <div className="flex items-center space-x-3">
                          <div className="flex-shrink-0">
                            <CheckCircle className="h-5 w-5 text-blue-600" />
                          </div>
                          <div>
                            <h4 className="font-medium text-blue-900">
                              External services detected
                            </h4>
                            <p className="text-blue-700 text-sm mt-0.5">
                              Dependencies like databases and message queues
                              will be included automatically.
                            </p>
                          </div>
                        </div>
                      </div>
                    )}

                    {/* Generate Button */}
                    <div className="flex justify-center pt-6">
                      <Button
                        onClick={generateCompose}
                        disabled={isLoading}
                        size="lg"
                        className="bg-blue-600 hover:bg-blue-700 text-white px-8 py-3 text-base font-medium"
                      >
                        <Play className="h-5 w-5 mr-2" />
                        Generate Docker Compose
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="flex items-center justify-center py-20">
                    <div className="text-center max-w-md">
                      <div className="mx-auto w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mb-4">
                        <Eye className="h-8 w-8 text-blue-600" />
                      </div>
                      <h3 className="text-xl font-semibold text-gray-900 mb-2">
                        Preview Your Configuration
                      </h3>
                      <p className="text-gray-500 mb-6">
                        Get an overview of services, networks, and volumes
                        before generating your Docker Compose file.
                      </p>
                      <Button
                        onClick={fetchPreview}
                        disabled={isLoading}
                        size="lg"
                        className="bg-blue-600 hover:bg-blue-700 text-white px-6"
                      >
                        <Eye className="h-4 w-4 mr-2" />
                        Load Preview
                      </Button>
                    </div>
                  </div>
                )}
              </div>
            )}

            {activeTab === "generate" && (
              <div>
                {generated ? (
                  <div className="space-y-6">
                    {/* Success & Actions */}
                    <div className="text-center">
                      <div className="inline-flex items-center justify-center w-16 h-16 bg-green-100 rounded-full mb-4">
                        <CheckCircle className="h-8 w-8 text-green-600" />
                      </div>
                      <h3 className="text-xl font-semibold text-gray-900 mb-2">
                        Files Generated Successfully!
                      </h3>
                      <p className="text-gray-600 mb-6 max-w-lg mx-auto">
                        Your Docker Compose configuration includes{" "}
                        {generated.serviceCount} services,{" "}
                        {generated.networkCount} networks, and{" "}
                        {generated.volumeCount} volumes.
                      </p>

                      <div className="flex flex-wrap gap-3 justify-center">
                        <Button
                          onClick={() => downloadCompose(false)}
                          size="lg"
                          className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-3"
                        >
                          <Download className="h-4 w-4 mr-2" />
                          Download docker-compose.yml
                        </Button>
                        {generateOverride && (
                          <Button
                            onClick={() => downloadCompose(true)}
                            variant="outline"
                            size="lg"
                            className="px-6 py-3"
                          >
                            <Download className="h-4 w-4 mr-2" />
                            Download Override
                          </Button>
                        )}
                        <Button
                          onClick={copyYamlToClipboard}
                          variant="outline"
                          size="lg"
                          className="px-6 py-3"
                        >
                          {yamlCopied ? (
                            <>
                              <Check className="h-4 w-4 mr-2 text-green-600" />
                              Copied!
                            </>
                          ) : (
                            <>
                              <Copy className="h-4 w-4 mr-2" />
                              Copy YAML
                            </>
                          )}
                        </Button>
                      </div>
                    </div>

                    {/* YAML Preview - Collapsible */}
                    <div className="border border-gray-200 rounded-lg overflow-hidden">
                      <div className="bg-gray-50 px-4 py-3 border-b border-gray-200">
                        <div className="flex items-center justify-between">
                          <h4 className="font-medium text-gray-900">
                            docker-compose.yml
                          </h4>
                          <div className="flex items-center space-x-3 text-sm text-gray-500">
                            <span>
                              {generated.services?.length || 0} services
                            </span>
                            <span>•</span>
                            <span>
                              {generated.yaml
                                ? generated.yaml.split("\n").length
                                : 0}{" "}
                              lines
                            </span>
                          </div>
                        </div>
                      </div>
                      <div className="bg-gray-900">
                        <div className="p-4 max-h-96 overflow-auto">
                          <pre className="text-sm text-gray-100 whitespace-pre-wrap font-mono leading-relaxed">
                            {generated.yaml}
                          </pre>
                        </div>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="flex items-center justify-center py-20">
                    <div className="text-center max-w-md">
                      <div className="mx-auto w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mb-4">
                        <Play className="h-8 w-8 text-blue-600" />
                      </div>
                      <h3 className="text-xl font-semibold text-gray-900 mb-2">
                        Generate Configuration Files
                      </h3>
                      <p className="text-gray-500 mb-6">
                        Create complete Docker Compose configuration files for
                        your microservices, including all services, networks,
                        and dependencies.
                      </p>
                      <Button
                        onClick={generateCompose}
                        disabled={isLoading}
                        size="lg"
                        className="bg-blue-600 hover:bg-blue-700 text-white px-6"
                      >
                        {isLoading ? (
                          <>
                            <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                            Generating...
                          </>
                        ) : (
                          <>
                            <Play className="h-4 w-4 mr-2" />
                            Generate Files
                          </>
                        )}
                      </Button>
                    </div>
                  </div>
                )}
              </div>
            )}

            {activeTab === "config" && (
              <div className="flex items-center justify-center py-20">
                <div className="text-center max-w-md">
                  <div className="mx-auto w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
                    <Settings className="h-8 w-8 text-gray-400" />
                  </div>
                  <h3 className="text-xl font-semibold text-gray-900 mb-2">
                    Advanced Configuration
                  </h3>
                  <p className="text-gray-500 mb-4">
                    Configure custom base images, volume mappings, and resource
                    limits for fine-tuned Docker Compose generation.
                  </p>
                  <div className="inline-flex items-center px-4 py-2 rounded-full bg-gray-100 text-gray-600 text-sm font-medium">
                    Coming Soon
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="flex justify-center items-center pt-6 border-t border-gray-200">
            <div className="flex items-center space-x-2 text-sm text-gray-500">
              <Container className="h-4 w-4" />
              <span>Docker Compose v3.8 • Generated by Vertex</span>
            </div>
          </div>
        </div>
      </ErrorBoundary>
    </Modal>
  );
}
