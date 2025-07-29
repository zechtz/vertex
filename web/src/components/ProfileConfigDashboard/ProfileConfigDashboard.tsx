import { useState, useEffect } from 'react';
import { 
  X, 
  Settings, 
  Server, 
  Database, 
  Globe, 
  GitBranch,
  Clock,
  Info,
  CheckCircle,
  AlertCircle,
  Monitor,
  Code,
  Users,
  Star,
  ExternalLink,
  Zap
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useProfile } from '@/contexts/ProfileContext';
import { ServiceProfile, Service } from '@/types';

interface ProfileConfigDashboardProps {
  isOpen: boolean;
  onClose: () => void;
  profile?: ServiceProfile | null;
}

interface ProfileStats {
  totalServices: number;
  runningServices: number;
  envVarsCount: number;
  lastUpdated: string;
}

export function ProfileConfigDashboard({ isOpen, onClose, profile: initialProfile }: ProfileConfigDashboardProps) {
  const { activeProfile, serviceProfiles, getProfileEnvVars } = useProfile();
  const [selectedProfile, setSelectedProfile] = useState<ServiceProfile | null>(initialProfile || activeProfile);
  const [profileEnvVars, setProfileEnvVars] = useState<Record<string, string>>({});
  const [availableServices, setAvailableServices] = useState<Service[]>([]);
  const [profileStats, setProfileStats] = useState<ProfileStats>({
    totalServices: 0,
    runningServices: 0,
    envVarsCount: 0,
    lastUpdated: 'Never'
  });
  const [isLoading, setIsLoading] = useState(false);
  const [activeTab, setActiveTab] = useState<'overview' | 'services' | 'environment' | 'comparison'>('overview');

  useEffect(() => {
    if (isOpen && selectedProfile) {
      loadProfileData();
    }
  }, [isOpen, selectedProfile?.id]);

  const loadProfileData = async () => {
    if (!selectedProfile) return;
    
    setIsLoading(true);
    try {
      // Load profile environment variables
      const envVars = await getProfileEnvVars(selectedProfile.id);
      setProfileEnvVars(envVars);

      // Load all services to get detailed information
      const servicesResponse = await fetch('/api/services');
      if (servicesResponse.ok) {
        const allServices = await servicesResponse.json();
        setAvailableServices(allServices || []);
        
        // Calculate profile stats
        const profileServices = allServices.filter((service: Service) => 
          selectedProfile.services.includes(service.name)
        );
        const runningServices = profileServices.filter((service: Service) => 
          service.status === 'running'
        );
        
        setProfileStats({
          totalServices: profileServices.length,
          runningServices: runningServices.length,
          envVarsCount: Object.keys(envVars).length,
          lastUpdated: selectedProfile.updatedAt || 'Unknown'
        });
      }
    } catch (error) {
      console.error('Failed to load profile data:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const getProfileServices = () => {
    return availableServices.filter(service => 
      selectedProfile?.services.includes(service.name)
    );
  };

  const getServiceStatusIcon = (status: string) => {
    switch (status) {
      case 'running': return <CheckCircle className="h-4 w-4 text-green-600" />;
      case 'stopped': return <AlertCircle className="h-4 w-4 text-gray-400" />;
      case 'error': return <AlertCircle className="h-4 w-4 text-red-600" />;
      default: return <AlertCircle className="h-4 w-4 text-gray-400" />;
    }
  };

  const getBuildSystemIcon = (buildSystem: string) => {
    switch (buildSystem) {
      case 'maven': return <Code className="h-4 w-4 text-orange-600" />;
      case 'gradle': return <Code className="h-4 w-4 text-blue-600" />;
      default: return <Code className="h-4 w-4 text-gray-600" />;
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-6xl max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-600 flex items-center justify-between bg-gradient-to-r from-blue-50 to-purple-50 dark:from-blue-900/20 dark:to-purple-900/20">
          <div className="flex items-center gap-3">
            <Monitor className="h-6 w-6 text-blue-600" />
            <div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                Profile Configuration Dashboard
              </h2>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Comprehensive view of profile settings and services
              </p>
            </div>
          </div>
          <Button variant="ghost" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Profile Selector */}
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700">
          <div className="flex items-center gap-4">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
              Select Profile:
            </label>
            <select
              value={selectedProfile?.id || ''}
              onChange={(e) => {
                const profile = serviceProfiles.find(p => p.id === e.target.value);
                setSelectedProfile(profile || null);
              }}
              className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            >
              <option value="">Select a profile...</option>
              {serviceProfiles.map(profile => (
                <option key={profile.id} value={profile.id}>
                  {profile.name} {profile.isActive ? '(Active)' : ''} {profile.isDefault ? '(Default)' : ''}
                </option>
              ))}
            </select>
            {selectedProfile?.isActive && (
              <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                <CheckCircle className="h-3 w-3 mr-1" />
                Active
              </Badge>
            )}
            {selectedProfile?.isDefault && (
              <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">
                <Star className="h-3 w-3 mr-1" />
                Default
              </Badge>
            )}
          </div>
        </div>

        {selectedProfile ? (
          <>
            {/* Tab Navigation */}
            <div className="px-6 py-3 border-b border-gray-200 dark:border-gray-600">
              <div className="flex space-x-6">
                {[
                  { id: 'overview', label: 'Overview', icon: Monitor },
                  { id: 'services', label: 'Services', icon: Server },
                  { id: 'environment', label: 'Environment', icon: Globe },
                  { id: 'comparison', label: 'Compare', icon: GitBranch }
                ].map(tab => (
                  <button
                    key={tab.id}
                    onClick={() => setActiveTab(tab.id as any)}
                    className={`flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                      activeTab === tab.id
                        ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
                        : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-700'
                    }`}
                  >
                    <tab.icon className="h-4 w-4" />
                    {tab.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Content Area */}
            <div className="flex-1 overflow-y-auto p-6 min-h-0">
              {isLoading ? (
                <div className="flex items-center justify-center py-12">
                  <div className="text-center">
                    <div className="h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
                    <p className="text-gray-600 dark:text-gray-400">Loading profile data...</p>
                  </div>
                </div>
              ) : (
                <>
                  {activeTab === 'overview' && (
                    <OverviewTab 
                      profile={selectedProfile} 
                      stats={profileStats}
                      services={getProfileServices()}
                    />
                  )}
                  {activeTab === 'services' && (
                    <ServicesTab 
                      profile={selectedProfile}
                      services={getProfileServices()}
                      getServiceStatusIcon={getServiceStatusIcon}
                      getBuildSystemIcon={getBuildSystemIcon}
                    />
                  )}
                  {activeTab === 'environment' && (
                    <EnvironmentTab 
                      profile={selectedProfile}
                      envVars={profileEnvVars}
                    />
                  )}
                  {activeTab === 'comparison' && (
                    <ComparisonTab 
                      selectedProfile={selectedProfile}
                      allProfiles={serviceProfiles}
                    />
                  )}
                </>
              )}
            </div>
          </>
        ) : (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <Monitor className="h-16 w-16 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
                No Profile Selected
              </h3>
              <p className="text-gray-600 dark:text-gray-400">
                Select a profile above to view its configuration dashboard.
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// Overview Tab Component
function OverviewTab({ 
  profile, 
  stats, 
  services 
}: { 
  profile: ServiceProfile; 
  stats: ProfileStats;
  services: Service[];
}) {
  return (
    <div className="space-y-6">
      {/* Profile Info Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Info className="h-5 w-5" />
            Profile Information
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium text-gray-600 dark:text-gray-400">Name</label>
              <p className="text-lg font-semibold text-gray-900 dark:text-gray-100">{profile.name}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-600 dark:text-gray-400">Last Updated</label>
              <p className="flex items-center gap-2 text-gray-900 dark:text-gray-100">
                <Clock className="h-4 w-4" />
                {new Date(stats.lastUpdated).toLocaleString()}
              </p>
            </div>
          </div>
          {profile.description && (
            <div>
              <label className="text-sm font-medium text-gray-600 dark:text-gray-400">Description</label>
              <p className="text-gray-900 dark:text-gray-100 mt-1">{profile.description}</p>
            </div>
          )}
          {(profile.projectsDir || profile.javaHomeOverride) && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {profile.projectsDir && (
                <div>
                  <label className="text-sm font-medium text-gray-600 dark:text-gray-400">Projects Directory</label>
                  <p className="text-gray-900 dark:text-gray-100 font-mono text-sm bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded">
                    {profile.projectsDir}
                  </p>
                </div>
              )}
              {profile.javaHomeOverride && (
                <div>
                  <label className="text-sm font-medium text-gray-600 dark:text-gray-400">Java Home Override</label>
                  <p className="text-gray-900 dark:text-gray-100 font-mono text-sm bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded">
                    {profile.javaHomeOverride}
                  </p>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center gap-4">
              <div className="p-3 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                <Server className="h-6 w-6 text-blue-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.totalServices}</p>
                <p className="text-sm text-gray-600 dark:text-gray-400">Total Services</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center gap-4">
              <div className="p-3 bg-green-100 dark:bg-green-900/30 rounded-lg">
                <Zap className="h-6 w-6 text-green-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.runningServices}</p>
                <p className="text-sm text-gray-600 dark:text-gray-400">Running Services</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center gap-4">
              <div className="p-3 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
                <Database className="h-6 w-6 text-purple-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.envVarsCount}</p>
                <p className="text-sm text-gray-600 dark:text-gray-400">Environment Variables</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Quick Service Status */}
      {services.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-5 w-5" />
              Service Status Overview
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {services.slice(0, 6).map(service => (
                <div key={service.name} className="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                  {service.status === 'running' ? (
                    <CheckCircle className="h-4 w-4 text-green-600" />
                  ) : (
                    <AlertCircle className="h-4 w-4 text-gray-400" />
                  )}
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
                      {service.name}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      Port {service.port} • {service.status}
                    </p>
                  </div>
                </div>
              ))}
            </div>
            {services.length > 6 && (
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-3 text-center">
                And {services.length - 6} more services...
              </p>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}

// Services Tab Component
function ServicesTab({ 
  profile, 
  services, 
  getServiceStatusIcon, 
  getBuildSystemIcon 
}: { 
  profile: ServiceProfile;
  services: Service[];
  getServiceStatusIcon: (status: string) => JSX.Element;
  getBuildSystemIcon: (buildSystem: string) => JSX.Element;
}) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">
          Services in {profile.name} ({services.length})
        </h3>
      </div>

      {services.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center">
            <Server className="h-16 w-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
              No Services Configured
            </h3>
            <p className="text-gray-600 dark:text-gray-400">
              This profile doesn't have any services associated with it yet.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {services.map(service => (
            <Card key={service.name}>
              <CardContent className="p-4">
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    {getServiceStatusIcon(service.status)}
                    <div>
                      <h4 className="font-medium text-gray-900 dark:text-gray-100">
                        {service.name}
                      </h4>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        Port {service.port}
                      </p>
                    </div>
                  </div>
                  <Badge variant={service.status === 'running' ? 'default' : 'secondary'}>
                    {service.status}
                  </Badge>
                </div>

                {service.description && (
                  <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
                    {service.description}
                  </p>
                )}

                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-sm">
                    {getBuildSystemIcon(service.buildSystem)}
                    <span className="text-gray-600 dark:text-gray-400">
                      Build System: {service.buildSystem || 'auto'}
                    </span>
                  </div>
                  
                  {service.healthUrl && (
                    <div className="flex items-center gap-2 text-sm">
                      <ExternalLink className="h-4 w-4 text-gray-400" />
                      <span className="text-gray-600 dark:text-gray-400">
                        Health: {service.healthUrl}
                      </span>
                    </div>
                  )}

                  {service.javaOpts && (
                    <div className="text-sm">
                      <span className="text-gray-600 dark:text-gray-400">Java Options:</span>
                      <p className="font-mono text-xs bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded mt-1">
                        {service.javaOpts}
                      </p>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}

// Environment Tab Component
function EnvironmentTab({ 
  profile, 
  envVars 
}: { 
  profile: ServiceProfile;
  envVars: Record<string, string>;
}) {
  const envVarsArray = Object.entries(envVars);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">
          Environment Variables ({envVarsArray.length})
        </h3>
      </div>

      {envVarsArray.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center">
            <Globe className="h-16 w-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
              No Environment Variables
            </h3>
            <p className="text-gray-600 dark:text-gray-400">
              This profile doesn't have any custom environment variables configured.
            </p>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent className="p-0">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-200 dark:border-gray-600">
                    <th className="text-left p-4 font-medium text-gray-900 dark:text-gray-100">
                      Variable Name
                    </th>
                    <th className="text-left p-4 font-medium text-gray-900 dark:text-gray-100">
                      Value
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {envVarsArray.map(([key, value], index) => (
                    <tr 
                      key={key}
                      className={`${index % 2 === 0 ? 'bg-gray-50 dark:bg-gray-700/50' : ''} hover:bg-gray-100 dark:hover:bg-gray-600/50`}
                    >
                      <td className="p-4 font-mono text-sm text-gray-900 dark:text-gray-100">
                        {key}
                      </td>
                      <td className="p-4 font-mono text-sm text-gray-600 dark:text-gray-400 break-all">
                        {value}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Profile-specific environment variables from the profile object */}
      {profile.envVars && Object.keys(profile.envVars).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Settings className="h-5 w-5" />
              Profile Environment Variables
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-200 dark:border-gray-600">
                    <th className="text-left p-4 font-medium text-gray-900 dark:text-gray-100">
                      Variable Name
                    </th>
                    <th className="text-left p-4 font-medium text-gray-900 dark:text-gray-100">
                      Value
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {Object.entries(profile.envVars).map(([key, value], index) => (
                    <tr 
                      key={key}
                      className={`${index % 2 === 0 ? 'bg-gray-50 dark:bg-gray-700/50' : ''} hover:bg-gray-100 dark:hover:bg-gray-600/50`}
                    >
                      <td className="p-4 font-mono text-sm text-gray-900 dark:text-gray-100">
                        {key}
                      </td>
                      <td className="p-4 font-mono text-sm text-gray-600 dark:text-gray-400 break-all">
                        {value}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

// Comparison Tab Component
function ComparisonTab({ 
  selectedProfile, 
  allProfiles 
}: { 
  selectedProfile: ServiceProfile;
  allProfiles: ServiceProfile[];
}) {
  const [compareProfile, setCompareProfile] = useState<ServiceProfile | null>(null);
  
  const otherProfiles = allProfiles.filter(p => p.id !== selectedProfile.id);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">
          Compare Profiles
        </h3>
        <select
          value={compareProfile?.id || ''}
          onChange={(e) => {
            const profile = allProfiles.find(p => p.id === e.target.value);
            setCompareProfile(profile || null);
          }}
          className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          <option value="">Select profile to compare...</option>
          {otherProfiles.map(profile => (
            <option key={profile.id} value={profile.id}>
              {profile.name}
            </option>
          ))}
        </select>
      </div>

      {compareProfile ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Selected Profile */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Users className="h-5 w-5" />
                {selectedProfile.name}
                {selectedProfile.isActive && (
                  <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                    Active
                  </Badge>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <h4 className="font-medium text-gray-900 dark:text-gray-100 mb-2">Services ({selectedProfile.services.length})</h4>
                <div className="space-y-1">
                  {selectedProfile.services.map(serviceName => (
                    <div key={serviceName} className="text-sm px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-200 rounded">
                      {serviceName}
                    </div>
                  ))}
                </div>
              </div>
              
              <div>
                <h4 className="font-medium text-gray-900 dark:text-gray-100 mb-2">Environment Variables ({Object.keys(selectedProfile.envVars || {}).length})</h4>
                <div className="space-y-1">
                  {Object.keys(selectedProfile.envVars || {}).map(key => (
                    <div key={key} className="text-sm font-mono px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200 rounded">
                      {key}
                    </div>
                  ))}
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Compare Profile */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Users className="h-5 w-5" />
                {compareProfile.name}
                {compareProfile.isActive && (
                  <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                    Active
                  </Badge>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <h4 className="font-medium text-gray-900 dark:text-gray-100 mb-2">Services ({compareProfile.services.length})</h4>
                <div className="space-y-1">
                  {compareProfile.services.map(serviceName => (
                    <div 
                      key={serviceName} 
                      className={`text-sm px-2 py-1 rounded ${
                        selectedProfile.services.includes(serviceName)
                          ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-200'
                          : 'bg-orange-100 dark:bg-orange-900/30 text-orange-800 dark:text-orange-200'
                      }`}
                    >
                      {serviceName}
                      {!selectedProfile.services.includes(serviceName) && ' (unique)'}
                    </div>
                  ))}
                </div>
              </div>
              
              <div>
                <h4 className="font-medium text-gray-900 dark:text-gray-100 mb-2">Environment Variables ({Object.keys(compareProfile.envVars || {}).length})</h4>
                <div className="space-y-1">
                  {Object.keys(compareProfile.envVars || {}).map(key => (
                    <div 
                      key={key} 
                      className={`text-sm font-mono px-2 py-1 rounded ${
                        selectedProfile.envVars?.[key]
                          ? 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200'
                          : 'bg-orange-100 dark:bg-orange-900/30 text-orange-800 dark:text-orange-200'
                      }`}
                    >
                      {key}
                      {!selectedProfile.envVars?.[key] && ' (unique)'}
                    </div>
                  ))}
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      ) : (
        <Card>
          <CardContent className="p-12 text-center">
            <GitBranch className="h-16 w-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
              Select a Profile to Compare
            </h3>
            <p className="text-gray-600 dark:text-gray-400">
              Choose another profile from the dropdown above to see a side-by-side comparison.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}