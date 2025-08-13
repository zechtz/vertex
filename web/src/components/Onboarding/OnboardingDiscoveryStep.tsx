import { useState, useEffect } from 'react';
import {
  Search,
  Server,
  RefreshCw,
  Loader2,
  Download
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { useAuth } from '@/contexts/AuthContext';
import { useToast, toast } from '@/components/ui/toast';
import { Service } from '@/types';

interface DiscoveredService {
  name: string;
  path: string;
  port: number;
  type: string;
  framework: string;
  description: string;
  properties: Record<string, string>;
  isValid: boolean;
  exists: boolean;
}

interface OnboardingDiscoveryStepProps {
  profile: any;
  onServicesDiscovered: (services: Service[]) => void;
  isProcessing: boolean;
  setIsProcessing: (processing: boolean) => void;
}

export function OnboardingDiscoveryStep({
  profile,
  onServicesDiscovered,
  isProcessing,
  setIsProcessing
}: OnboardingDiscoveryStepProps) {
  const { token } = useAuth();
  const { addToast } = useToast();

  const [discoveredServices, setDiscoveredServices] = useState<DiscoveredService[]>([]);
  const [selectedServices, setSelectedServices] = useState<string[]>([]);
  const [isScanning, setIsScanning] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const [hasScanned, setHasScanned] = useState(false);

  // Auto-start discovery when step loads
  useEffect(() => {
    if (profile && !hasScanned) {
      scanForServices();
    }
  }, [profile]);

  // Use the same scanForServices function as AutoDiscoveryModal
  const scanForServices = async () => {
    setIsScanning(true);
    setHasScanned(true);

    try {
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };

      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const response = await fetch('/api/auto-discovery/scan', {
        method: 'POST',
        headers,
      });

      if (!response.ok) {
        throw new Error(`Discovery failed: ${response.status} ${response.statusText}`);
      }

      const data = await response.json();
      const services = data.discoveredServices || [];
      
      setDiscoveredServices(services);
      
      // Auto-select all valid services
      const validServiceNames = services
        .filter((service: DiscoveredService) => service.isValid && !service.exists)
        .map((service: DiscoveredService) => service.name);
      
      setSelectedServices(validServiceNames);

      if (services.length === 0) {
        addToast(
          toast.info(
            'No Services Found',
            'No microservices were discovered in the specified directory. You can add services manually later.'
          )
        );
      } else {
        addToast(
          toast.success(
            'Services Discovered!',
            `Found ${services.length} potential services. Review and select which ones to import.`
          )
        );
      }
    } catch (error) {
      console.error('Auto-discovery failed:', error);
      addToast(
        toast.error(
          'Discovery Failed',
          error instanceof Error ? error.message : 'Failed to discover services'
        )
      );
    } finally {
      setIsScanning(false);
    }
  };


  const handleServiceToggle = (serviceName: string, checked: boolean) => {
    setSelectedServices(prev => 
      checked 
        ? [...prev, serviceName]
        : prev.filter(name => name !== serviceName)
    );
  };

  const handleImportServices = async () => {
    if (selectedServices.length === 0) {
      // Allow proceeding with no services
      onServicesDiscovered([]);
      return;
    }

    setIsProcessing(true);
    setIsImporting(true);

    try {
      const servicesToImport = discoveredServices
        .filter(service => selectedServices.includes(service.name))
        .map(service => ({
          name: service.name,
          path: service.path,
          port: service.port,
          type: service.type,
          framework: service.framework,
          description: service.description,
          properties: service.properties,
          profileId: profile.id,
        }));

      const response = await fetch('/api/auto-discovery/import-bulk', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          services: servicesToImport,
        }),
      });

      if (!response.ok) {
        throw new Error(`Import failed: ${response.status} ${response.statusText}`);
      }

      const result = await response.json();
      const importedServices = result.importedServices || [];

      addToast(
        toast.success(
          'Services Imported!',
          `Successfully imported ${importedServices.length} services to "${profile.name}" profile.`
        )
      );

      onServicesDiscovered(importedServices);
    } catch (error) {
      console.error('Failed to import services:', error);
      addToast(
        toast.error(
          'Import Failed',
          error instanceof Error ? error.message : 'Failed to import services'
        )
      );
    } finally {
      setIsProcessing(false);
      setIsImporting(false);
    }
  };

  const handleSkipDiscovery = () => {
    addToast(
      toast.info(
        'Discovery Skipped',
        'You can add services manually later from the main dashboard.'
      )
    );
    onServicesDiscovered([]);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center">
        <div className="mx-auto w-16 h-16 bg-green-100 dark:bg-green-900/20 rounded-full flex items-center justify-center mb-4">
          <Search className="w-8 h-8 text-green-600 dark:text-green-400" />
        </div>
        <h3 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-2">
          Discover Your Services
        </h3>
        <p className="text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
          Let's automatically discover microservices in your "{profile?.name}" profile directory.
          We'll scan for Spring Boot, Node.js, and other common frameworks.
        </p>
      </div>

      {/* Scan Status */}
      {isScanning && (
        <Card className="max-w-2xl mx-auto">
          <CardContent className="p-6">
            <div className="flex items-center gap-3 text-blue-600 dark:text-blue-400">
              <Loader2 className="w-6 h-6 animate-spin" />
              <div>
                <div className="font-medium">Scanning for services...</div>
                <div className="text-sm opacity-75">
                  Discovering microservices in "{profile?.name}" profile directory
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Discovered Services */}
      {hasScanned && (
        <Card className="max-w-4xl mx-auto">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Server className="w-5 h-5 text-green-500" />
              Discovered Services ({discoveredServices.length})
            </CardTitle>
            {discoveredServices.length > 0 && (
              <div className="flex items-center gap-2 text-sm text-gray-500">
                <Checkbox
                  checked={selectedServices.length === discoveredServices.filter(s => s.isValid && !s.exists).length}
                  onCheckedChange={(checked) => {
                    const validServices = discoveredServices
                      .filter(s => s.isValid && !s.exists)
                      .map(s => s.name);
                    setSelectedServices(checked ? validServices : []);
                  }}
                />
                Select All
              </div>
            )}
          </CardHeader>
          <CardContent>
            {discoveredServices.length === 0 ? (
              <div className="text-center py-12">
                <Server className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                <h4 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
                  No Services Found
                </h4>
                <p className="text-gray-600 dark:text-gray-400 mb-4">
                  No microservices were discovered in the specified directory.
                </p>
                <div className="flex gap-2 justify-center">
                  <Button onClick={scanForServices} variant="outline">
                    <RefreshCw className="w-4 h-4 mr-2" />
                    Try Again
                  </Button>
                  <Button onClick={handleSkipDiscovery} variant="outline">
                    Skip & Continue
                  </Button>
                </div>
              </div>
            ) : (
              <div className="space-y-3">
                {discoveredServices.map((service) => (
                  <div
                    key={service.name}
                    className={`p-4 border rounded-lg flex items-start gap-3 ${
                      service.exists
                        ? 'bg-yellow-50 border-yellow-200 dark:bg-yellow-900/10 dark:border-yellow-800'
                        : service.isValid
                          ? 'bg-white border-gray-200 dark:bg-gray-800 dark:border-gray-700'
                          : 'bg-red-50 border-red-200 dark:bg-red-900/10 dark:border-red-800'
                    }`}
                  >
                    <Checkbox
                      checked={selectedServices.includes(service.name)}
                      onCheckedChange={(checked) => handleServiceToggle(service.name, checked === true)}
                      disabled={!service.isValid || service.exists}
                    />
                    
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <h4 className="font-medium text-gray-900 dark:text-gray-100">
                          {service.name}
                        </h4>
                        <Badge variant={service.isValid ? 'default' : 'destructive'}>
                          {service.framework}
                        </Badge>
                        {service.exists && (
                          <Badge variant="secondary">Already Exists</Badge>
                        )}
                      </div>
                      
                      <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                        {service.description}
                      </p>
                      
                      <div className="flex items-center gap-4 text-xs text-gray-500">
                        <span>Port: {service.port}</span>
                        <span>Type: {service.type}</span>
                        <span>Path: {service.path}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Actions */}
      {hasScanned && (
        <div className="flex justify-center gap-3 max-w-2xl mx-auto">
          <Button
            onClick={handleSkipDiscovery}
            variant="outline"
            disabled={isProcessing}
          >
            Skip for Now
          </Button>
          <Button
            onClick={handleImportServices}
            disabled={isProcessing}
            className="min-w-[140px]"
          >
            {isImporting ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Importing...
              </>
            ) : (
              <>
                <Download className="w-4 h-4 mr-2" />
                Import Services ({selectedServices.length})
              </>
            )}
          </Button>
        </div>
      )}

      {/* Tips */}
      <Card className="max-w-2xl mx-auto bg-gray-50 dark:bg-gray-800">
        <CardContent className="p-4">
          <h4 className="font-medium text-gray-900 dark:text-gray-100 mb-2">
            🔍 Discovery Tips
          </h4>
          <ul className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
            <li>• Auto-discovery looks for common project files (pom.xml, package.json, etc.)</li>
            <li>• Services already in Vertex will be marked as "Already Exists"</li>
            <li>• You can always add more services manually after onboarding</li>
            <li>• Selected services will be added to your "{profile?.name}" profile</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}

export default OnboardingDiscoveryStep;