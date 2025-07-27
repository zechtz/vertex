import { useState } from 'react';
import {
  X,
  Search,
  Download,
  CheckCircle,
  AlertCircle,
  FileCode,
  Server,
  RefreshCw,
  FolderOpen,
  Plus
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';

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

interface AutoDiscoveryModalProps {
  isOpen: boolean;
  onClose: () => void;
  onServiceImported: () => void;
}

export function AutoDiscoveryModal({
  isOpen,
  onClose,
  onServiceImported
}: AutoDiscoveryModalProps) {
  const [discoveredServices, setDiscoveredServices] = useState<DiscoveredService[]>([]);
  const [isScanning, setIsScanning] = useState(false);
  const [isImporting, setIsImporting] = useState<Record<string, boolean>>({});
  const [searchTerm, setSearchTerm] = useState('');
  const [hasScanned, setHasScanned] = useState(false);

  const filteredServices = discoveredServices.filter(service =>
    service.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    service.type.toLowerCase().includes(searchTerm.toLowerCase()) ||
    service.framework.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const scanForServices = async () => {
    setIsScanning(true);
    try {
      const response = await fetch('/api/auto-discovery/scan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' }
      });

      if (!response.ok) {
        throw new Error(`Failed to scan: ${response.status} ${response.statusText}`);
      }

      const result = await response.json();
      setDiscoveredServices(result.discoveredServices || []);
      setHasScanned(true);
    } catch (error) {
      console.error('Failed to scan for services:', error);
      alert('Failed to scan for services: ' + (error instanceof Error ? error.message : 'Unknown error'));
    } finally {
      setIsScanning(false);
    }
  };

  const importService = async (service: DiscoveredService) => {
    setIsImporting(prev => ({ ...prev, [service.name]: true }));
    try {
      const response = await fetch('/api/auto-discovery/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(service)
      });

      if (!response.ok) {
        throw new Error(`Failed to import service: ${response.status} ${response.statusText}`);
      }

      const result = await response.json();
      console.log('Service imported successfully:', result);
      
      // Mark service as existing in our local state
      setDiscoveredServices(prev =>
        prev.map(s => s.name === service.name ? { ...s, exists: true } : s)
      );

      // Notify parent component
      onServiceImported();
    } catch (error) {
      console.error('Failed to import service:', error);
      alert('Failed to import service: ' + (error instanceof Error ? error.message : 'Unknown error'));
    } finally {
      setIsImporting(prev => ({ ...prev, [service.name]: false }));
    }
  };

  const getServiceTypeIcon = (type: string) => {
    switch (type.toLowerCase()) {
      case 'registry':
        return <Server className="w-4 h-4 text-blue-600" />;
      case 'config-server':
        return <FileCode className="w-4 h-4 text-green-600" />;
      case 'api-gateway':
        return <Server className="w-4 h-4 text-purple-600" />;
      case 'authentication':
        return <Server className="w-4 h-4 text-orange-600" />;
      case 'cache':
        return <Server className="w-4 h-4 text-yellow-600" />;
      default:
        return <Server className="w-4 h-4 text-gray-600" />;
    }
  };

  const getServiceTypeBadgeColor = (type: string) => {
    switch (type.toLowerCase()) {
      case 'registry':
        return 'bg-blue-100 text-blue-800';
      case 'config-server':
        return 'bg-green-100 text-green-800';
      case 'api-gateway':
        return 'bg-purple-100 text-purple-800';
      case 'authentication':
        return 'bg-orange-100 text-orange-800';
      case 'cache':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-screen items-center justify-center p-4">
        {/* Backdrop */}
        <div className="fixed inset-0 bg-black/50" onClick={onClose} />
        
        {/* Modal */}
        <div className="relative w-full max-w-6xl max-h-[95vh] overflow-y-auto">
          <div className="relative bg-white rounded-lg shadow-xl">
            {/* Header */}
            <div className="flex items-center justify-between p-6 border-b border-gray-200">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-green-100 rounded-lg">
                  <Search className="h-6 w-6 text-green-600" />
                </div>
                <div>
                  <h2 className="text-xl font-semibold text-gray-900">
                    Auto-Discovery
                  </h2>
                  <p className="text-sm text-gray-600">
                    Automatically detect Spring Boot services in your project directory
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Button
                  onClick={scanForServices}
                  disabled={isScanning}
                  className="bg-green-600 hover:bg-green-700"
                >
                  <Search className={`w-4 h-4 mr-2 ${isScanning ? 'animate-spin' : ''}`} />
                  {isScanning ? 'Scanning...' : 'Scan Directory'}
                </Button>
                <Button variant="ghost" onClick={onClose}>
                  <X className="w-5 h-5" />
                </Button>
              </div>
            </div>

            <div className="p-6">
              {!hasScanned && !isScanning && (
                <div className="text-center py-12">
                  <FolderOpen className="w-16 h-16 text-gray-400 mx-auto mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">
                    Ready to discover services
                  </h3>
                  <p className="text-gray-600 mb-6">
                    Click "Scan Directory" to automatically detect Spring Boot services in your project directory.
                  </p>
                  <Button
                    onClick={scanForServices}
                    className="bg-green-600 hover:bg-green-700"
                  >
                    <Search className="w-4 h-4 mr-2" />
                    Start Scanning
                  </Button>
                </div>
              )}

              {isScanning && (
                <div className="text-center py-12">
                  <RefreshCw className="w-16 h-16 text-blue-600 mx-auto mb-4 animate-spin" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">
                    Scanning for services...
                  </h3>
                  <p className="text-gray-600">
                    Looking for Maven and Gradle projects with Spring Boot dependencies.
                  </p>
                </div>
              )}

              {hasScanned && !isScanning && (
                <>
                  {/* Search and Summary */}
                  <div className="mb-6">
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center gap-4">
                        <Input
                          type="text"
                          placeholder="Search services..."
                          value={searchTerm}
                          onChange={(e) => setSearchTerm(e.target.value)}
                          className="w-64"
                        />
                        <Badge variant="outline" className="text-sm">
                          {filteredServices.length} of {discoveredServices.length} services
                        </Badge>
                      </div>
                      <Button
                        variant="outline"
                        onClick={scanForServices}
                        disabled={isScanning}
                      >
                        <RefreshCw className="w-4 h-4 mr-2" />
                        Rescan
                      </Button>
                    </div>

                    {/* Statistics */}
                    <div className="grid grid-cols-4 gap-4 mb-6">
                      <Card>
                        <CardContent className="p-3">
                          <div className="flex items-center gap-2">
                            <Search className="w-4 h-4 text-blue-600" />
                            <span className="text-sm font-medium">Found</span>
                          </div>
                          <p className="text-2xl font-bold text-blue-600">{discoveredServices.length}</p>
                        </CardContent>
                      </Card>
                      <Card>
                        <CardContent className="p-3">
                          <div className="flex items-center gap-2">
                            <Plus className="w-4 h-4 text-green-600" />
                            <span className="text-sm font-medium">New</span>
                          </div>
                          <p className="text-2xl font-bold text-green-600">
                            {discoveredServices.filter(s => !s.exists).length}
                          </p>
                        </CardContent>
                      </Card>
                      <Card>
                        <CardContent className="p-3">
                          <div className="flex items-center gap-2">
                            <CheckCircle className="w-4 h-4 text-gray-600" />
                            <span className="text-sm font-medium">Existing</span>
                          </div>
                          <p className="text-2xl font-bold text-gray-600">
                            {discoveredServices.filter(s => s.exists).length}
                          </p>
                        </CardContent>
                      </Card>
                      <Card>
                        <CardContent className="p-3">
                          <div className="flex items-center gap-2">
                            <AlertCircle className="w-4 h-4 text-orange-600" />
                            <span className="text-sm font-medium">Invalid</span>
                          </div>
                          <p className="text-2xl font-bold text-orange-600">
                            {discoveredServices.filter(s => !s.isValid).length}
                          </p>
                        </CardContent>
                      </Card>
                    </div>
                  </div>

                  {/* Services List */}
                  {filteredServices.length === 0 && discoveredServices.length > 0 && (
                    <div className="text-center py-8">
                      <Search className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                      <p className="text-gray-600">No services match your search criteria.</p>
                    </div>
                  )}

                  {filteredServices.length === 0 && discoveredServices.length === 0 && (
                    <div className="text-center py-8">
                      <FileCode className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                      <p className="text-gray-600">No Spring Boot services found in the project directory.</p>
                    </div>
                  )}

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {filteredServices.map((service) => (
                      <Card key={service.name} className={`border ${service.exists ? 'border-gray-300 bg-gray-50' : 'border-green-300 bg-green-50'}`}>
                        <CardHeader className="pb-3">
                          <div className="flex items-start justify-between">
                            <div className="flex items-center gap-2">
                              {getServiceTypeIcon(service.type)}
                              <CardTitle className="text-lg">{service.name}</CardTitle>
                            </div>
                            <div className="flex items-center gap-2">
                              <Badge className={getServiceTypeBadgeColor(service.type)}>
                                {service.type}
                              </Badge>
                              {service.exists && (
                                <Badge variant="secondary">
                                  <CheckCircle className="w-3 h-3 mr-1" />
                                  Imported
                                </Badge>
                              )}
                            </div>
                          </div>
                        </CardHeader>
                        <CardContent className="pt-0">
                          <div className="space-y-2 text-sm">
                            <div className="flex items-center gap-2">
                              <span className="font-medium">Framework:</span>
                              <Badge variant="outline">{service.framework}</Badge>
                            </div>
                            <div className="flex items-center gap-2">
                              <span className="font-medium">Path:</span>
                              <span className="text-gray-600 font-mono text-xs">{service.path}</span>
                            </div>
                            <div className="flex items-center gap-2">
                              <span className="font-medium">Port:</span>
                              <Badge variant="outline">{service.port}</Badge>
                            </div>
                            {service.description && (
                              <div>
                                <span className="font-medium">Description:</span>
                                <p className="text-gray-600 text-xs mt-1">{service.description}</p>
                              </div>
                            )}
                          </div>

                          <div className="mt-4 flex justify-end">
                            {service.exists ? (
                              <Button variant="secondary" disabled>
                                <CheckCircle className="w-4 h-4 mr-2" />
                                Already Imported
                              </Button>
                            ) : (
                              <Button
                                onClick={() => importService(service)}
                                disabled={isImporting[service.name]}
                                className="bg-blue-600 hover:bg-blue-700"
                              >
                                <Download className={`w-4 h-4 mr-2 ${isImporting[service.name] ? 'animate-spin' : ''}`} />
                                {isImporting[service.name] ? 'Importing...' : 'Import Service'}
                              </Button>
                            )}
                          </div>
                        </CardContent>
                      </Card>
                    ))}
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default AutoDiscoveryModal;