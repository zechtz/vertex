import { useState, useEffect, useRef } from 'react';
import { 
  Database, 
  Server, 
  Globe, 
  MessageSquare, 
  Zap, 
  Activity,
  Info,
  RefreshCw
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useAuth } from '@/contexts/AuthContext';

interface TopologyNode {
  id: string;
  name: string;
  type: 'service' | 'database' | 'external';
  status: string;
  healthStatus: string;
  port: number;
  position?: {
    x: number;
    y: number;
  };
  metadata: Record<string, any>;
}

interface Connection {
  source: string;
  target: string;
  type: 'http' | 'database' | 'message_queue';
  status: 'active' | 'inactive' | 'error';
  description: string;
}

interface ServiceTopology {
  services: TopologyNode[];
  connections: Connection[];
  generated: string;
}

interface ServiceTopologyProps {
  className?: string;
}

export function ServiceTopology({ className = '' }: ServiceTopologyProps) {
  const { token } = useAuth();
  const [topology, setTopology] = useState<ServiceTopology | null>(null);
  const [selectedNode, setSelectedNode] = useState<TopologyNode | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const svgRef = useRef<SVGSVGElement>(null);

  const fetchTopology = async () => {
    try {
      setIsLoading(true);
      setError(null);
      
      const headers: Record<string, string> = {};
      
      // Use the token from AuthContext
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
      
      const response = await fetch('/api/topology', {
        headers,
      });
      if (!response.ok) {
        throw new Error(`Failed to fetch topology: ${response.status} ${response.statusText}`);
      }
      
      const data: ServiceTopology = await response.json();
      setTopology(data);
    } catch (error) {
      console.error('Failed to fetch topology:', error);
      setError(error instanceof Error ? error.message : 'Failed to fetch topology');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchTopology();
  }, []);

  const getNodeIcon = (node: TopologyNode) => {
    switch (node.type) {
      case 'service':
        return <Server className="w-6 h-6" />;
      case 'database':
        return node.id === 'redis' ? <Zap className="w-6 h-6" /> : <Database className="w-6 h-6" />;
      case 'external':
        return node.id === 'rabbitmq' ? <MessageSquare className="w-6 h-6" /> : <Globe className="w-6 h-6" />;
      default:
        return <Activity className="w-6 h-6" />;
    }
  };

  const getNodeColor = (node: TopologyNode) => {
    if (node.type === 'external' || node.type === 'database') {
      return 'bg-gray-100 dark:bg-gray-700 border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300';
    }
    
    switch (node.status) {
      case 'running':
        return node.healthStatus === 'healthy' 
          ? 'bg-green-100 dark:bg-green-900/30 border-green-300 dark:border-green-600 text-green-700 dark:text-green-300'
          : 'bg-yellow-100 dark:bg-yellow-900/30 border-yellow-300 dark:border-yellow-600 text-yellow-700 dark:text-yellow-300';
      case 'stopped':
        return 'bg-red-100 dark:bg-red-900/30 border-red-300 dark:border-red-600 text-red-700 dark:text-red-300';
      default:
        return 'bg-gray-100 dark:bg-gray-700 border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300';
    }
  };

  const getConnectionColor = (connection: Connection) => {
    switch (connection.status) {
      case 'active':
        return '#10b981'; // green
      case 'inactive':
        return '#9ca3af'; // gray
      case 'error':
        return '#ef4444'; // red
      default:
        return '#9ca3af';
    }
  };

  const getConnectionStyle = (connection: Connection) => {
    switch (connection.type) {
      case 'database':
        return '5,5'; // dashed
      case 'message_queue':
        return '10,5'; // long dash
      default:
        return ''; // solid
    }
  };

  const handleNodeClick = (node: TopologyNode) => {
    setSelectedNode(node);
  };

  const renderConnections = () => {
    if (!topology) return null;

    return topology.connections.map((connection, index) => {
      const sourceNode = topology.services.find(n => n.id === connection.source);
      const targetNode = topology.services.find(n => n.id === connection.target);
      
      if (!sourceNode?.position || !targetNode?.position) return null;

      const x1 = sourceNode.position.x;
      const y1 = sourceNode.position.y;
      const x2 = targetNode.position.x;
      const y2 = targetNode.position.y;

      // Calculate arrow position
      const dx = x2 - x1;
      const dy = y2 - y1;
      const length = Math.sqrt(dx * dx + dy * dy);
      const unitX = dx / length;
      const unitY = dy / length;
      
      // Adjust for node radius (40px)
      const startX = x1 + unitX * 40;
      const startY = y1 + unitY * 40;
      const endX = x2 - unitX * 40;
      const endY = y2 - unitY * 40;

      return (
        <g key={`${connection.source}-${connection.target}-${index}`}>
          <line
            x1={startX}
            y1={startY}
            x2={endX}
            y2={endY}
            stroke={getConnectionColor(connection)}
            strokeWidth="2"
            strokeDasharray={getConnectionStyle(connection)}
            opacity={connection.status === 'active' ? 1 : 0.5}
          />
          {/* Arrow marker */}
          <polygon
            points={`${endX-8},${endY-4} ${endX},${endY} ${endX-8},${endY+4}`}
            fill={getConnectionColor(connection)}
            opacity={connection.status === 'active' ? 1 : 0.5}
          />
        </g>
      );
    });
  };

  const renderNodes = () => {
    if (!topology) return null;

    return topology.services.map((node) => {
      if (!node.position) return null;

      return (
        <g 
          key={node.id} 
          transform={`translate(${node.position.x}, ${node.position.y})`}
          onClick={() => handleNodeClick(node)}
          className="cursor-pointer"
        >
          {/* Node circle */}
          <circle
            r="35"
            className={`${getNodeColor(node)} stroke-2 hover:scale-110 transition-transform`}
            fill="currentColor"
            fillOpacity="0.1"
            stroke="currentColor"
          />
          
          {/* Node icon */}
          <foreignObject x="-12" y="-12" width="24" height="24">
            <div className="flex items-center justify-center">
              {getNodeIcon(node)}
            </div>
          </foreignObject>
          
          {/* Node label */}
          <text
            y="50"
            textAnchor="middle"
            className="text-xs font-medium fill-gray-700"
          >
            {node.name}
          </text>
          
          {/* Status indicator */}
          {node.type === 'service' && (
            <circle
              cx="25"
              cy="-25"
              r="6"
              className={
                node.status === 'running' 
                  ? 'fill-green-500' 
                  : 'fill-red-500'
              }
            />
          )}
        </g>
      );
    });
  };

  if (isLoading) {
    return (
      <div className={`${className}`}>
        <Card className="h-[600px]">
          <CardContent className="h-full flex items-center justify-center">
            <div className="text-center">
              <RefreshCw className="h-8 w-8 animate-spin mx-auto mb-4 text-blue-600" />
              <p className="text-gray-600">Loading service topology...</p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`${className}`}>
        <Card className="h-[600px]">
          <CardContent className="h-full flex items-center justify-center">
            <div className="text-center">
              <div className="text-red-600 mb-4">
                <Info className="h-8 w-8 mx-auto mb-2" />
                <p className="font-semibold">Failed to load topology</p>
                <p className="text-sm text-gray-600">{error}</p>
              </div>
              <Button onClick={fetchTopology} variant="outline">
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
    <div className={`space-y-4 ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold">Service Topology</h3>
          <p className="text-sm text-gray-600">
            Visual representation of service dependencies and connections
          </p>
        </div>
        <Button onClick={fetchTopology} variant="outline" size="sm">
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </Button>
      </div>

      {/* Main topology view */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Topology diagram */}
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Service Architecture</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="relative bg-gray-50 dark:bg-gray-800 rounded-lg overflow-hidden">
                <svg
                  ref={svgRef}
                  width="100%"
                  height="500"
                  viewBox="0 0 800 600"
                  className="border border-gray-200 dark:border-gray-700"
                >
                  {/* Grid background */}
                  <defs>
                    <pattern
                      id="grid-light"
                      width="20"
                      height="20"
                      patternUnits="userSpaceOnUse"
                    >
                      <path
                        d="M 20 0 L 0 0 0 20"
                        fill="none"
                        stroke="#e5e7eb"
                        strokeWidth="1"
                      />
                    </pattern>
                    <pattern
                      id="grid-dark"
                      width="20"
                      height="20"
                      patternUnits="userSpaceOnUse"
                    >
                      <path
                        d="M 20 0 L 0 0 0 20"
                        fill="none"
                        stroke="#4b5563"
                        strokeWidth="1"
                      />
                    </pattern>
                  </defs>
                  <rect width="100%" height="100%" fill="url(#grid-light)" className="dark:hidden" />
                  <rect width="100%" height="100%" fill="url(#grid-dark)" className="hidden dark:block" />
                  
                  {/* Connections */}
                  {renderConnections()}
                  
                  {/* Nodes */}
                  {renderNodes()}
                </svg>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Node details panel */}
        <div>
          <Card className="h-fit sticky top-4">
            <CardHeader>
              <CardTitle className="text-base">
                {selectedNode ? 'Node Details' : 'Service Legend'}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {selectedNode ? (
                <div className="space-y-4">
                  <div className="flex items-center gap-3">
                    <div className={`p-2 rounded-lg ${getNodeColor(selectedNode)}`}>
                      {getNodeIcon(selectedNode)}
                    </div>
                    <div>
                      <h4 className="font-medium">{selectedNode.name}</h4>
                      <p className="text-sm text-gray-600 capitalize">{selectedNode.type}</p>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Status</span>
                      <Badge 
                        variant={selectedNode.status === 'running' ? 'success' : 'secondary'}
                      >
                        {selectedNode.status}
                      </Badge>
                    </div>
                    
                    {selectedNode.port && (
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Port</span>
                        <code className="text-sm bg-gray-100 px-2 py-1 rounded">
                          {selectedNode.port}
                        </code>
                      </div>
                    )}

                    {selectedNode.metadata.description && (
                      <div>
                        <span className="text-sm text-gray-600">Description</span>
                        <p className="text-sm mt-1">{selectedNode.metadata.description}</p>
                      </div>
                    )}

                    {selectedNode.type === 'service' && selectedNode.metadata.cpuPercent !== undefined && (
                      <div className="space-y-2">
                        <span className="text-sm text-gray-600">Resource Usage</span>
                        <div className="text-xs space-y-1">
                          <div>CPU: {selectedNode.metadata.cpuPercent?.toFixed(1)}%</div>
                          <div>Memory: {selectedNode.metadata.memoryPercent?.toFixed(1)}%</div>
                          {selectedNode.metadata.uptime && (
                            <div>Uptime: {selectedNode.metadata.uptime}</div>
                          )}
                        </div>
                      </div>
                    )}
                  </div>

                  <Button 
                    variant="outline" 
                    size="sm" 
                    onClick={() => setSelectedNode(null)}
                    className="w-full"
                  >
                    Close Details
                  </Button>
                </div>
              ) : (
                <div className="space-y-4">
                  <div className="space-y-3">
                    <div className="flex items-center gap-2 text-sm">
                      <div className="w-4 h-4 bg-green-100 border border-green-300 rounded-full"></div>
                      <span>Running Service</span>
                    </div>
                    <div className="flex items-center gap-2 text-sm">
                      <div className="w-4 h-4 bg-red-100 border border-red-300 rounded-full"></div>
                      <span>Stopped Service</span>
                    </div>
                    <div className="flex items-center gap-2 text-sm">
                      <div className="w-4 h-4 bg-gray-100 border border-gray-300 rounded-full"></div>
                      <span>External System</span>
                    </div>
                  </div>

                  <hr />

                  <div className="space-y-3">
                    <div className="flex items-center gap-2 text-sm">
                      <div className="w-6 h-0.5 bg-green-500"></div>
                      <span>Active Connection</span>
                    </div>
                    <div className="flex items-center gap-2 text-sm">
                      <div className="w-6 h-0.5 bg-gray-400"></div>
                      <span>Inactive Connection</span>
                    </div>
                    <div className="flex items-center gap-2 text-sm">
                      <div className="w-6 h-0.5 bg-gray-400 border-dashed border-t"></div>
                      <span>Database Connection</span>
                    </div>
                  </div>

                  <p className="text-xs text-gray-500">
                    Click on any node to view detailed information
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {topology && (
        <div className="text-xs text-gray-500">
          Last updated: {new Date(topology.generated).toLocaleString()}
        </div>
      )}
    </div>
  );
}

export default ServiceTopology;