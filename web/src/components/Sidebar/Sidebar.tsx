import {
  Server,
  Database,
  Settings,
  BarChart3,
  FileText,
  Network,
  GitBranch,
  Layers,
  Search,
  Users,
  Monitor,
} from "lucide-react";

interface SidebarProps {
  activeSection: string;
  onSectionChange: (section: string) => void;
  className?: string;
  isCollapsed?: boolean;
}

interface NavigationItem {
  id: string;
  label: string;
  icon: React.ReactNode;
  description?: string;
}

export function Sidebar({
  activeSection,
  onSectionChange,
  className = "",
  isCollapsed = false,
}: SidebarProps) {
  const navigationItems: NavigationItem[] = [
    {
      id: "services",
      label: "Services",
      icon: <Server className="w-5 h-5" />,
      description: "Manage microservices",
    },
    {
      id: "profiles",
      label: "Profiles",
      icon: <Users className="w-5 h-5" />,
      description: "Service profiles",
    },
    {
      id: "dashboard",
      label: "Dashboard",
      icon: <Monitor className="w-5 h-5" />,
      description: "Profile configuration dashboard",
    },
    {
      id: "metrics",
      label: "Metrics",
      icon: <BarChart3 className="w-5 h-5" />,
      description: "System performance",
    },
    {
      id: "logs",
      label: "Logs",
      icon: <FileText className="w-5 h-5" />,
      description: "Log aggregation",
    },
    {
      id: "topology",
      label: "Topology",
      icon: <Network className="w-5 h-5" />,
      description: "Service architecture",
    },
    {
      id: "dependencies",
      label: "Dependencies",
      icon: <GitBranch className="w-5 h-5" />,
      description: "Service dependencies",
    },
    {
      id: "auto-discovery",
      label: "Auto-Discovery",
      icon: <Search className="w-5 h-5" />,
      description: "Discover new services",
    },
    {
      id: "configurations",
      label: "Configurations",
      icon: <Layers className="w-5 h-5" />,
      description: "Service configs",
    },
    {
      id: "environment",
      label: "Environment",
      icon: <Database className="w-5 h-5" />,
      description: "Environment variables",
    },
    {
      id: "settings",
      label: "Settings",
      icon: <Settings className="w-5 h-5" />,
      description: "Global settings",
    },
  ];

  return (
    <div className={`${className}`}>
      {/* Sidebar */}
      <div
        className={`
        fixed left-0 top-16 h-[calc(100vh-4rem)] bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 transition-all duration-300 z-30
        ${isCollapsed ? "w-16" : "w-64"}
      `}
      >
        {/* Navigation */}
        <nav className="p-4 space-y-2">
          {navigationItems.map((item) => (
            <button
              key={item.id}
              onClick={() => onSectionChange(item.id)}
              className={`
                w-full flex items-center gap-3 p-3 rounded-lg transition-all duration-200
                ${
                  activeSection === item.id
                    ? "bg-blue-50 dark:bg-blue-900/50 text-blue-700 dark:text-blue-300 border border-blue-200 dark:border-blue-700"
                    : "text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-white"
                }
                ${isCollapsed ? "justify-center" : "justify-start"}
              `}
              title={isCollapsed ? item.label : undefined}
            >
              <div
                className={`
                flex-shrink-0 transition-colors
                ${activeSection === item.id ? "text-blue-600" : "text-gray-500"}
              `}
              >
                {item.icon}
              </div>
              {!isCollapsed && (
                <div className="flex-1 text-left">
                  <div className="font-medium text-sm">{item.label}</div>
                  {item.description && (
                    <div className="text-xs text-gray-500 mt-0.5">
                      {item.description}
                    </div>
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
          <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
            <div className="text-xs text-gray-500 dark:text-gray-400 text-center">
              <div className="font-medium">NeST Service Manager</div>
              <div>v2.0.0</div>
            </div>
          </div>
        )}
      </div>

      {/* Mobile Overlay */}
      {!isCollapsed && (
        <div
          className="fixed inset-0 bg-black/20 z-20 lg:hidden"
          onClick={() => {
            // Mobile overlay click - could emit an event if needed
          }}
        />
      )}
    </div>
  );
}

export default Sidebar;
