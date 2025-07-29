import { useState } from 'react';
import { ChevronDown, Star, Play, Users, Check, Settings } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useProfile } from '@/contexts/ProfileContext';
import { useToast, toast } from '@/components/ui/toast';

export function ProfileSwitcher() {
  const { serviceProfiles, activeProfile, activateProfile, applyProfile, isApplying } = useProfile();
  const { addToast } = useToast();
  const [showDropdown, setShowDropdown] = useState(false);
  const [activatingId, setActivatingId] = useState<string | null>(null);

  const handleProfileActivate = async (profileId: string) => {
    try {
      setActivatingId(profileId);
      await activateProfile(profileId);
      addToast(toast.success('Success', 'Profile activated successfully!'));
      setShowDropdown(false);
    } catch (error) {
      console.error('Failed to activate profile:', error);
      addToast(toast.error('Error', 'Failed to activate profile. Please try again.'));
    } finally {
      setActivatingId(null);
    }
  };

  const handleProfileApply = async (profileId: string) => {
    try {
      await applyProfile(profileId);
      addToast(toast.success('Success', 'Profile applied and services started!'));
      setShowDropdown(false);
    } catch (error) {
      console.error('Failed to apply profile:', error);
      addToast(toast.error('Error', 'Failed to apply profile. Please try again.'));
    }
  };

  if (serviceProfiles.length === 0) {
    return null; // Don't show if no profiles exist
  }

  return (
    <div className="relative">
      <Button
        variant="outline"
        size="sm"
        onClick={() => setShowDropdown(!showDropdown)}
        className="flex items-center gap-2 min-w-[140px] justify-between"
        disabled={isApplying}
      >
        <div className="flex items-center gap-2">
          <Users className="h-4 w-4" />
          <span className="truncate">
            {activeProfile ? activeProfile.name : 'No Active Profile'}
          </span>
          {activeProfile?.isActive && (
            <div className="h-2 w-2 bg-green-500 rounded-full flex-shrink-0" title="Active Profile" />
          )}
        </div>
        <ChevronDown className={`h-4 w-4 transition-transform ${showDropdown ? 'rotate-180' : ''}`} />
      </Button>

      {showDropdown && (
        <>
          {/* Backdrop */}
          <div 
            className="fixed inset-0 z-40" 
            onClick={() => setShowDropdown(false)}
          />
          
          {/* Dropdown */}
          <div className="absolute right-0 top-full mt-2 w-64 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-600 py-2 z-50">
            <div className="px-3 py-2 border-b border-gray-100 dark:border-gray-600">
              <h3 className="font-medium text-gray-900 dark:text-gray-100">Profile Management</h3>
              <p className="text-xs text-gray-600 dark:text-gray-400">
                Activate: Set as current context • Apply: Start services
              </p>
            </div>
            
            <div className="max-h-64 overflow-y-auto">
              {serviceProfiles.map((profile) => (
                <div
                  key={profile.id}
                  className={`px-3 py-2 transition-colors ${
                    profile.isActive ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-gray-900 dark:text-gray-100 truncate">
                          {profile.name}
                        </span>
                        {profile.isDefault && (
                          <Star className="h-3 w-3 text-yellow-500 fill-current flex-shrink-0" />
                        )}
                        {profile.isActive && (
                          <div className="flex items-center gap-1">
                            <Check className="h-3 w-3 text-green-600 flex-shrink-0" />
                            <span className="px-1.5 py-0.5 bg-green-100 text-green-800 text-xs rounded flex-shrink-0">
                              Active
                            </span>
                          </div>
                        )}
                      </div>
                      <p className="text-xs text-gray-600 dark:text-gray-400 truncate">
                        {profile.description || 'No description'}
                      </p>
                      <p className="text-xs text-gray-500 dark:text-gray-500">
                        {profile.services.length} service{profile.services.length !== 1 ? 's' : ''}
                      </p>
                    </div>
                  </div>
                  
                  {/* Action buttons */}
                  <div className="flex items-center gap-2 mt-2">
                    {!profile.isActive && (
                      <button
                        onClick={() => handleProfileActivate(profile.id)}
                        disabled={activatingId === profile.id}
                        className="flex items-center gap-1 px-2 py-1 text-xs bg-blue-100 text-blue-700 hover:bg-blue-200 rounded transition-colors"
                      >
                        {activatingId === profile.id ? (
                          <div className="h-3 w-3 border-2 border-blue-600 border-t-transparent rounded-full animate-spin" />
                        ) : (
                          <Settings className="h-3 w-3" />
                        )}
                        Activate
                      </button>
                    )}
                    <button
                      onClick={() => handleProfileApply(profile.id)}
                      disabled={isApplying || activatingId === profile.id}
                      className="flex items-center gap-1 px-2 py-1 text-xs bg-green-100 text-green-700 hover:bg-green-200 rounded transition-colors"
                    >
                      {isApplying ? (
                        <div className="h-3 w-3 border-2 border-green-600 border-t-transparent rounded-full animate-spin" />
                      ) : (
                        <Play className="h-3 w-3" />
                      )}
                      Apply & Start
                    </button>
                  </div>
                </div>
              ))}
            </div>
            
            {serviceProfiles.length === 0 && (
              <div className="px-3 py-4 text-center text-gray-500 dark:text-gray-400">
                <Users className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p className="text-sm">No profiles available</p>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}