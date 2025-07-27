import { useState } from 'react';
import { 
  Server, 
  Database, 
  Settings, 
  BarChart3, 
  FileText, 
  Network, 
  GitBranch,
  Menu,
  X,
  Activity,
  Layers
} from 'lucide-react';

interface SidebarProps {
  activeSection: string;
  onSectionChange: (section: string) => void;
  className?: string;
  onCollapsedChange?: (collapsed: boolean) => void;
}

interface NavigationItem {
  id: string;
  label: string;
  icon: React.ReactNode;
  description?: string;
}

export function Sidebar({ activeSection, onSectionChange, className = '', onCollapsedChange }: SidebarProps) {
  const [isCollapsed, setIsCollapsed] = useState(false);

  const handleToggleCollapse = () => {
    const newCollapsed = !isCollapsed;
    setIsCollapsed(newCollapsed);
    onCollapsedChange?.(newCollapsed);
  };

  const navigationItems: NavigationItem[] = [
    {
      id: 'services',
      label: 'Services',
      icon: <Server className="w-5 h-5" />,
      description: 'Manage microservices'
    },
    {
      id: 'metrics',
      label: 'Metrics',
      icon: <BarChart3 className="w-5 h-5" />,
      description: 'System performance'
    },
    {
      id: 'logs',
      label: 'Logs',
      icon: <FileText className="w-5 h-5" />,
      description: 'Log aggregation'
    },
    {
      id: 'topology',
      label: 'Topology',
      icon: <Network className="w-5 h-5" />,
      description: 'Service architecture'
    },
    {
      id: 'dependencies',
      label: 'Dependencies',
      icon: <GitBranch className="w-5 h-5" />,
      description: 'Service dependencies'
    },
    {
      id: 'configurations',
      label: 'Configurations',
      icon: <Layers className="w-5 h-5" />,
      description: 'Service configs'
    },
    {
      id: 'environment',
      label: 'Environment',
      icon: <Database className="w-5 h-5" />,
      description: 'Environment variables'
    },
    {
      id: 'settings',
      label: 'Settings',
      icon: <Settings className="w-5 h-5" />,
      description: 'Global settings'
    }
  ];

  return (
    <div className={`${className}`}>
      {/* Sidebar */}
      <div className={`
        fixed left-0 top-0 h-full bg-white border-r border-gray-200 transition-all duration-300 z-30
        ${isCollapsed ? 'w-16' : 'w-64'}
      `}>
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200">
          {!isCollapsed && (
            <div className="flex items-center gap-3">
              <div className="p-2 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg">
                <Activity className="h-6 w-6 text-white" />
              </div>
              <div>
                <h1 className="text-lg font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                  NeST Manager
                </h1>
                <p className="text-xs text-gray-500">Service Management</p>
              </div>
            </div>
          )}
          <button
            onClick={handleToggleCollapse}
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          >
            {isCollapsed ? (
              <Menu className="w-5 h-5 text-gray-600" />
            ) : (
              <X className="w-5 h-5 text-gray-600" />
            )}
          </button>
        </div>

        {/* Navigation */}
        <nav className="p-4 space-y-2">
          {navigationItems.map((item) => (
            <button
              key={item.id}
              onClick={() => onSectionChange(item.id)}
              className={`
                w-full flex items-center gap-3 p-3 rounded-lg transition-all duration-200
                ${activeSection === item.id 
                  ? 'bg-blue-50 text-blue-700 border border-blue-200' 
                  : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'
                }
                ${isCollapsed ? 'justify-center' : 'justify-start'}
              `}
              title={isCollapsed ? item.label : undefined}
            >
              <div className={`
                flex-shrink-0 transition-colors
                ${activeSection === item.id ? 'text-blue-600' : 'text-gray-500'}
              `}>
                {item.icon}
              </div>
              {!isCollapsed && (
                <div className="flex-1 text-left">
                  <div className="font-medium text-sm">{item.label}</div>
                  {item.description && (
                    <div className="text-xs text-gray-500 mt-0.5">{item.description}</div>
                  )}
                </div>
              )}
              {!isCollapsed && activeSection === item.id && (
                <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
              )}
            </button>
          ))}
        </nav>

        {/* Footer */}
        {!isCollapsed && (
          <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-gray-200 bg-gray-50">
            <div className="text-xs text-gray-500 text-center">
              <div className="font-medium">NeST Service Manager</div>
              <div>v1.0.0</div>
            </div>
          </div>
        )}
      </div>

      {/* Mobile Overlay */}
      {!isCollapsed && (
        <div 
          className="fixed inset-0 bg-black/20 z-20 lg:hidden"
          onClick={() => {
            setIsCollapsed(true);
            onCollapsedChange?.(true);
          }}
        />
      )}
    </div>
  );
}

export default Sidebar;