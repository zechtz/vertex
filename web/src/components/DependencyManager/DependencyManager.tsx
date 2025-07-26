import { useState, useEffect } from 'react';
import { 
  GitBranch, 
  Clock, 
  Shield, 
  AlertTriangle, 
  CheckCircle, 
  XCircle,
  Edit,
  RefreshCw,
  Activity
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Service } from '@/types';

interface ServiceDependency {
  serviceName: string;
  type: 'hard' | 'soft' | 'optional';
  healthCheck: boolean;
  timeout: number; // in seconds
  retryInterval: number; // in seconds
  required: boolean;
  description: string;
}

interface DependencyInfo {
  dependencies: ServiceDependency[];
  dependentOn: string[];
  startupDelay: string;
}

interface ValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
  checked: string;
}

interface StartupOrder {
  startupOrder: string[];
  services: number;
  generated: string;
}

interface DependencyManagerProps {
  services: Service[];
  className?: string;
}

export function DependencyManager({ services, className = '' }: DependencyManagerProps) {
  const [dependencies, setDependencies] = useState<Record<string, DependencyInfo>>({});
  const [validation, setValidation] = useState<ValidationResult | null>(null);
  const [startupOrder, setStartupOrder] = useState<StartupOrder | null>(null);
  const [selectedService, setSelectedService] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchDependencies = async () => {
    try {
      setIsLoading(true);
      setError(null);
      
      const response = await fetch('/api/dependencies');
      if (!response.ok) {
        throw new Error(`Failed to fetch dependencies: ${response.status}`);
      }
      
      const data = await response.json();
      setDependencies(data);
    } catch (error) {
      console.error('Failed to fetch dependencies:', error);
      setError(error instanceof Error ? error.message : 'Failed to fetch dependencies');
    } finally {
      setIsLoading(false);
    }
  };

  const fetchValidation = async () => {
    try {
      const response = await fetch('/api/dependencies/validate');
      if (!response.ok) return;
      
      const data = await response.json();
      setValidation(data);
    } catch (error) {
      console.error('Failed to fetch validation:', error);
    }
  };

  const fetchStartupOrder = async () => {
    try {
      const response = await fetch('/api/dependencies/startup-order', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ services: [] })
      });
      if (!response.ok) return;
      
      const data = await response.json();
      setStartupOrder(data);
    } catch (error) {
      console.error('Failed to fetch startup order:', error);
    }
  };

  useEffect(() => {
    fetchDependencies();
    fetchValidation();
    fetchStartupOrder();
  }, []);

  const getDependencyTypeColor = (type: string) => {
    switch (type) {
      case 'hard':
        return 'bg-red-100 text-red-800';
      case 'soft':
        return 'bg-yellow-100 text-yellow-800';
      case 'optional':
        return 'bg-blue-100 text-blue-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const formatTimeout = (nanoseconds: number) => {
    const seconds = nanoseconds / 1000000000;
    if (seconds >= 60) {
      return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
    }
    return `${seconds}s`;
  };

  const handleRefresh = () => {
    fetchDependencies();
    fetchValidation();
    fetchStartupOrder();
  };

  const renderDependencyList = () => {
    if (Object.keys(dependencies).length === 0) {
      return (
        <div className="text-center py-8 text-gray-500">
          <GitBranch className="h-12 w-12 mx-auto mb-4 text-gray-300" />
          <p className="text-lg font-medium">No dependencies configured</p>
          <p>Services will start in default order</p>
        </div>
      );
    }

    return (
      <div className="space-y-4">
        {Object.entries(dependencies).map(([serviceName, info]) => (
          <Card 
            key={serviceName}
            className={`transition-all duration-200 hover:shadow-md ${
              selectedService === serviceName ? 'ring-2 ring-blue-500' : ''
            }`}
          >
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-blue-100 rounded-lg">
                    <GitBranch className="h-5 w-5 text-blue-600" />
                  </div>
                  <div>
                    <CardTitle className="text-lg">{serviceName}</CardTitle>
                    <p className="text-sm text-gray-600">
                      {info.dependencies.length} dependencies • 
                      Startup delay: {formatTimeout(parseInt(info.startupDelay) || 0)}
                    </p>
                  </div>
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setSelectedService(
                      selectedService === serviceName ? null : serviceName
                    )}
                  >
                    {selectedService === serviceName ? 'Hide' : 'Details'}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      // TODO: Open edit modal
                    }}
                  >
                    <Edit className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </CardHeader>
            
            {selectedService === serviceName && (
              <CardContent>
                <div className="space-y-4">
                  {/* Dependencies */}
                  <div>
                    <h4 className="font-medium mb-3 flex items-center gap-2">
                      <Shield className="h-4 w-4" />
                      Dependencies ({info.dependencies.length})
                    </h4>
                    <div className="space-y-2">
                      {info.dependencies.map((dep, index) => (
                        <div 
                          key={index}
                          className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
                        >
                          <div className="flex items-center gap-3">
                            <Badge className={getDependencyTypeColor(dep.type)}>
                              {dep.type}
                            </Badge>
                            <div>
                              <p className="font-medium">{dep.serviceName}</p>
                              <p className="text-sm text-gray-600">{dep.description}</p>
                            </div>
                          </div>
                          <div className="flex items-center gap-4 text-sm text-gray-600">
                            {dep.healthCheck && (
                              <div className="flex items-center gap-1">
                                <Activity className="h-3 w-3" />
                                Health Check
                              </div>
                            )}
                            <div className="flex items-center gap-1">
                              <Clock className="h-3 w-3" />
                              {formatTimeout(dep.timeout)}
                            </div>
                            {dep.required && (
                              <Badge variant="destructive" className="text-xs">
                                Required
                              </Badge>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Dependents */}
                  {info.dependentOn && info.dependentOn.length > 0 && (
                    <div>
                      <h4 className="font-medium mb-2">Services depending on this:</h4>
                      <div className="flex flex-wrap gap-2">
                        {info.dependentOn.map((dependent) => (
                          <Badge key={dependent} variant="outline">
                            {dependent}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </CardContent>
            )}
          </Card>
        ))}
      </div>
    );
  };

  const renderValidationStatus = () => {
    if (!validation) return null;

    return (
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            {validation.valid ? (
              <CheckCircle className="h-5 w-5 text-green-600" />
            ) : (
              <XCircle className="h-5 w-5 text-red-600" />
            )}
            Dependency Validation
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Status</span>
              <Badge variant={validation.valid ? 'success' : 'destructive'}>
                {validation.valid ? 'Valid' : 'Invalid'}
              </Badge>
            </div>
            
            {validation.errors.length > 0 && (
              <div>
                <h4 className="font-medium text-red-600 mb-2">Errors:</h4>
                <ul className="list-disc list-inside space-y-1 text-sm text-red-600">
                  {validation.errors.map((error, index) => (
                    <li key={index}>{error}</li>
                  ))}
                </ul>
              </div>
            )}
            
            {validation.warnings.length > 0 && (
              <div>
                <h4 className="font-medium text-yellow-600 mb-2">Warnings:</h4>
                <ul className="list-disc list-inside space-y-1 text-sm text-yellow-600">
                  {validation.warnings.map((warning, index) => (
                    <li key={index}>{warning}</li>
                  ))}
                </ul>
              </div>
            )}
            
            <div className="text-xs text-gray-500">
              Last checked: {new Date(validation.checked).toLocaleString()}
            </div>
          </div>
        </CardContent>
      </Card>
    );
  };

  const renderStartupOrder = () => {
    if (!startupOrder) return null;

    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5 text-blue-600" />
            Startup Order
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-center justify-between text-sm text-gray-600">
              <span>{startupOrder.services} services</span>
              <span>Generated: {new Date(startupOrder.generated).toLocaleString()}</span>
            </div>
            
            <div className="flex flex-wrap gap-2">
              {startupOrder.startupOrder.map((service, index) => (
                <div key={service} className="flex items-center gap-2">
                  <Badge 
                    variant="outline" 
                    className="px-3 py-1"
                  >
                    {index + 1}. {service}
                  </Badge>
                  {index < startupOrder.startupOrder.length - 1 && (
                    <span className="text-gray-400">→</span>
                  )}
                </div>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>
    );
  };

  if (isLoading) {
    return (
      <div className={`${className}`}>
        <Card className="h-96">
          <CardContent className="h-full flex items-center justify-center">
            <div className="text-center">
              <RefreshCw className="h-8 w-8 animate-spin mx-auto mb-4 text-blue-600" />
              <p className="text-gray-600">Loading dependencies...</p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`${className}`}>
        <Card className="h-96">
          <CardContent className="h-full flex items-center justify-center">
            <div className="text-center">
              <AlertTriangle className="h-8 w-8 mx-auto mb-4 text-red-600" />
              <p className="text-red-600 font-semibold mb-2">Failed to load dependencies</p>
              <p className="text-sm text-gray-600 mb-4">{error}</p>
              <Button onClick={fetchDependencies} variant="outline">
                <RefreshCw className="h-4 w-4 mr-2" />
                Retry
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Dependency Management</h2>
          <p className="text-gray-600">
            Configure and visualize service dependencies for optimal startup ordering
          </p>
        </div>
        <Button onClick={handleRefresh} variant="outline">
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </Button>
      </div>

      {/* Validation Status */}
      {renderValidationStatus()}

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Dependencies List */}
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span>Service Dependencies</span>
                <Badge variant="outline">
                  {Object.keys(dependencies).length} services configured
                </Badge>
              </CardTitle>
            </CardHeader>
            <CardContent>
              {renderDependencyList()}
            </CardContent>
          </Card>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Startup Order */}
          {renderStartupOrder()}
          
          {/* Quick Stats */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Quick Stats</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">Total Services</span>
                <Badge variant="outline">{services.length}</Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">With Dependencies</span>
                <Badge variant="outline">{Object.keys(dependencies).length}</Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">Hard Dependencies</span>
                <Badge variant="destructive">
                  {Object.values(dependencies).reduce(
                    (count, info) => count + info.dependencies.filter(dep => dep.type === 'hard').length,
                    0
                  )}
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">Soft Dependencies</span>
                <Badge className="bg-yellow-100 text-yellow-800">
                  {Object.values(dependencies).reduce(
                    (count, info) => count + info.dependencies.filter(dep => dep.type === 'soft').length,
                    0
                  )}
                </Badge>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

export default DependencyManager;