import { useState, useEffect } from 'react';
import {
  X,
  Plus,
  Trash2,
  Save,
  RefreshCw,
  GitBranch,
  Settings,
  CheckCircle
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Service } from '@/types';

interface ServiceDependency {
  serviceName: string;
  type: 'hard' | 'soft';
  required: boolean;
}

interface DependencyConfig {
  [serviceName: string]: {
    dependencies: ServiceDependency[];
    order: number;
  };
}

interface DependencyConfigModalProps {
  isOpen: boolean;
  onClose: () => void;
  services: Service[];
}

export function DependencyConfigModal({
  isOpen,
  onClose,
  services
}: DependencyConfigModalProps) {
  const [config, setConfig] = useState<DependencyConfig>({});
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [startupOrder, setStartupOrder] = useState<string[]>([]);

  // Default NeST service dependencies
  const defaultNestServices = {
    'nest-registry-server': { order: 1, dependencies: [] },
    'nest-config-server': { order: 2, dependencies: ['nest-registry-server'] },
    'nest-cache': { order: 3, dependencies: ['nest-registry-server', 'nest-config-server'] },
    'nest-gateway': { order: 4, dependencies: ['nest-registry-server', 'nest-config-server', 'nest-cache'] },
    'nest-uaa': { order: 5, dependencies: ['nest-registry-server', 'nest-config-server', 'nest-cache', 'nest-gateway'] }
  };

  useEffect(() => {
    if (isOpen) {
      loadDependencies();
    }
  }, [isOpen]);

  const loadDependencies = async () => {
    setIsLoading(true);
    try {
      // Load dependencies from API
      const response = await fetch('/api/dependencies', {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
      });

      let apiConfig: any = {};
      
      if (response.ok) {
        apiConfig = await response.json();
        console.log('Loaded dependencies from API:', apiConfig);
      } else {
        console.warn('Failed to load dependencies from API, using defaults');
      }

      // Initialize configuration for all services
      const initialConfig: DependencyConfig = {};
      
      services.forEach(service => {
        // Check if we have saved configuration for this service
        const savedServiceConfig = apiConfig[service.name];
        
        if (savedServiceConfig && savedServiceConfig.dependencies) {
          // Use saved configuration from database
          initialConfig[service.name] = {
            order: savedServiceConfig.order || service.order || 10,
            dependencies: savedServiceConfig.dependencies.map((dep: any) => ({
              serviceName: dep.serviceName,
              type: (dep.type || 'hard') as 'hard' | 'soft',
              required: dep.required !== undefined ? dep.required : true
            }))
          };
        } else {
          // Fall back to default NeST service dependencies
          const defaultService = defaultNestServices[service.name as keyof typeof defaultNestServices];
          
          if (defaultService) {
            initialConfig[service.name] = {
              order: defaultService.order,
              dependencies: defaultService.dependencies.map(dep => ({
                serviceName: dep,
                type: 'hard' as const,
                required: true
              }))
            };
          } else {
            initialConfig[service.name] = {
              order: service.order || 10,
              dependencies: []
            };
          }
        }
      });

      setConfig(initialConfig);
      calculateStartupOrder(initialConfig);
    } catch (error) {
      console.error('Failed to load dependencies:', error);
      
      // Fall back to default configuration on error
      const fallbackConfig: DependencyConfig = {};
      services.forEach(service => {
        const defaultService = defaultNestServices[service.name as keyof typeof defaultNestServices];
        
        if (defaultService) {
          fallbackConfig[service.name] = {
            order: defaultService.order,
            dependencies: defaultService.dependencies.map(dep => ({
              serviceName: dep,
              type: 'hard' as const,
              required: true
            }))
          };
        } else {
          fallbackConfig[service.name] = {
            order: service.order || 10,
            dependencies: []
          };
        }
      });
      
      setConfig(fallbackConfig);
      calculateStartupOrder(fallbackConfig);
    } finally {
      setIsLoading(false);
    }
  };

  const calculateStartupOrder = (currentConfig: DependencyConfig) => {
    const order = Object.entries(currentConfig)
      .sort(([, a], [, b]) => a.order - b.order)
      .map(([name]) => name);
    setStartupOrder(order);
  };

  const addDependency = (serviceName: string) => {
    const availableServices = services
      .filter(s => s.name !== serviceName && !config[serviceName]?.dependencies.some(d => d.serviceName === s.name))
      .map(s => s.name);

    if (availableServices.length === 0) return;

    const newDep: ServiceDependency = {
      serviceName: availableServices[0],
      type: 'hard',
      required: true
    };

    setConfig(prev => ({
      ...prev,
      [serviceName]: {
        ...prev[serviceName],
        dependencies: [...(prev[serviceName]?.dependencies || []), newDep]
      }
    }));
  };

  const removeDependency = (serviceName: string, depIndex: number) => {
    setConfig(prev => ({
      ...prev,
      [serviceName]: {
        ...prev[serviceName],
        dependencies: prev[serviceName]?.dependencies.filter((_, i) => i !== depIndex) || []
      }
    }));
  };

  const updateDependency = (serviceName: string, depIndex: number, updates: Partial<ServiceDependency>) => {
    setConfig(prev => ({
      ...prev,
      [serviceName]: {
        ...prev[serviceName],
        dependencies: prev[serviceName]?.dependencies.map((dep, i) => 
          i === depIndex ? { ...dep, ...updates } : dep
        ) || []
      }
    }));
  };

  const updateServiceOrder = (serviceName: string, newOrder: number) => {
    setConfig(prev => {
      const updated = {
        ...prev,
        [serviceName]: {
          ...prev[serviceName],
          order: newOrder
        }
      };
      calculateStartupOrder(updated);
      return updated;
    });
  };

  const saveDependencies = async () => {
    setIsSaving(true);
    try {
      const response = await fetch('/api/dependencies', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      });

      if (!response.ok) {
        throw new Error(`Failed to save dependencies: ${response.status}`);
      }

      const result = await response.json();
      console.log('Dependencies saved successfully:', result);
      onClose();
    } catch (error) {
      console.error('Failed to save dependencies:', error);
      alert('Failed to save dependencies: ' + (error instanceof Error ? error.message : 'Unknown error'));
    } finally {
      setIsSaving(false);
    }
  };

  const resetToDefaults = () => {
    loadDependencies();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-screen items-center justify-center p-4">
        {/* Backdrop */}
        <div className="fixed inset-0 bg-black/50" onClick={onClose} />
        
        {/* Modal */}
        <div className="relative w-full max-w-7xl max-h-[95vh] overflow-y-auto">
          <div className="relative bg-white rounded-lg shadow-xl">
            {/* Header */}
            <div className="flex items-center justify-between p-6 border-b border-gray-200">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-blue-100 rounded-lg">
                  <GitBranch className="h-6 w-6 text-blue-600" />
                </div>
                <div>
                  <h2 className="text-xl font-semibold text-gray-900">
                    Service Dependencies Configuration
                  </h2>
                  <p className="text-sm text-gray-600">
                    Configure service startup dependencies and order
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  onClick={resetToDefaults}
                  disabled={isLoading}
                >
                  <RefreshCw className={`w-4 h-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
                  Reset to Defaults
                </Button>
                <Button
                  onClick={saveDependencies}
                  disabled={isSaving}
                  className="bg-green-600 hover:bg-green-700"
                >
                  <Save className="w-4 h-4 mr-2" />
                  {isSaving ? 'Saving...' : 'Save Configuration'}
                </Button>
                <Button variant="ghost" onClick={onClose}>
                  <X className="w-5 h-5" />
                </Button>
              </div>
            </div>

            <div className="p-6 grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Service List */}
              <div className="lg:col-span-2">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg flex items-center gap-2">
                      <Settings className="w-5 h-5" />
                      Service Dependencies
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    {services.map(service => (
                      <Card key={service.name} className="border border-gray-200">
                        <CardContent className="p-4">
                          <div className="flex items-center justify-between mb-3">
                            <div className="flex items-center gap-3">
                              <h3 className="font-semibold text-gray-900">{service.name}</h3>
                              <Badge variant="outline" className="text-xs">
                                Order: {config[service.name]?.order || service.order}
                              </Badge>
                            </div>
                            <div className="flex items-center gap-2">
                              <Input
                                type="number"
                                value={config[service.name]?.order || service.order}
                                onChange={(e) => updateServiceOrder(service.name, parseInt(e.target.value) || 1)}
                                className="w-20 h-8"
                                min="1"
                                max="100"
                              />
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => addDependency(service.name)}
                              >
                                <Plus className="w-3 h-3 mr-1" />
                                Add Dependency
                              </Button>
                            </div>
                          </div>

                          {/* Dependencies */}
                          <div className="space-y-2">
                            {config[service.name]?.dependencies.map((dep, index) => (
                              <div key={index} className="flex items-center gap-2 p-2 bg-gray-50 rounded-lg">
                                <select
                                  value={dep.serviceName}
                                  onChange={(e) => updateDependency(service.name, index, { serviceName: e.target.value })}
                                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                                >
                                  {services
                                    .filter(s => s.name !== service.name)
                                    .map(s => (
                                      <option key={s.name} value={s.name}>
                                        {s.name}
                                      </option>
                                    ))}
                                </select>
                                
                                <select
                                  value={dep.type}
                                  onChange={(e) => updateDependency(service.name, index, { type: e.target.value as 'hard' | 'soft' })}
                                  className="w-24 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                                >
                                  <option value="hard">Hard</option>
                                  <option value="soft">Soft</option>
                                </select>

                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => removeDependency(service.name, index)}
                                  className="text-red-600 hover:text-red-700 hover:bg-red-50"
                                >
                                  <Trash2 className="w-3 h-3" />
                                </Button>
                              </div>
                            ))}
                            
                            {(!config[service.name]?.dependencies || config[service.name]?.dependencies.length === 0) && (
                              <p className="text-sm text-gray-500 italic">No dependencies configured</p>
                            )}
                          </div>
                        </CardContent>
                      </Card>
                    ))}
                  </CardContent>
                </Card>
              </div>

              {/* Preview Panel */}
              <div>
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg flex items-center gap-2">
                      <CheckCircle className="w-5 h-5 text-green-600" />
                      Startup Order Preview
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      {startupOrder.map((serviceName, index) => (
                        <div key={serviceName} className="flex items-center gap-3 p-2 bg-gray-50 rounded-lg">
                          <Badge variant="default" className="w-8 h-8 rounded-full flex items-center justify-center">
                            {index + 1}
                          </Badge>
                          <span className="font-medium">{serviceName}</span>
                        </div>
                      ))}
                    </div>

                    <div className="mt-6 p-3 bg-blue-50 rounded-lg">
                      <h4 className="font-medium text-blue-900 mb-2">Default NeST Order:</h4>
                      <ol className="text-sm text-blue-700 space-y-1">
                        <li>1. Eureka (Registry)</li>
                        <li>2. Config Server</li>
                        <li>3. Cache</li>
                        <li>4. Gateway</li>
                        <li>5. UAA</li>
                        <li>6. Other Services</li>
                      </ol>
                    </div>

                    <div className="mt-4 p-3 bg-green-50 rounded-lg">
                      <h4 className="font-medium text-green-900 mb-2">✅ Eureka Integration:</h4>
                      <p className="text-sm text-green-700">
                        Health status is now checked via Eureka registry for better reliability.
                        Services register with Eureka and report their health automatically.
                      </p>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default DependencyConfigModal;