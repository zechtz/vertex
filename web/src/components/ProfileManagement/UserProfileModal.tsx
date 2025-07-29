import { useState, useEffect } from 'react';
import { X, Save, User, Bell, Globe, Monitor, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useProfile } from '@/contexts/ProfileContext';
import { useAuth } from '@/contexts/AuthContext';
import { UserProfileUpdateRequest, UserPreferences } from '@/types';
import { useToast, toast } from '@/components/ui/toast';

interface UserProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function UserProfileModal({ isOpen, onClose }: UserProfileModalProps) {
  const { user } = useAuth();
  const { userProfile, updateUserProfile, isUpdating } = useProfile();
  const { addToast } = useToast();
  const [formData, setFormData] = useState<UserProfileUpdateRequest>({
    displayName: '',
    avatar: '',
    preferences: {
      theme: 'light',
      language: 'en',
      notificationSettings: {
        serviceStatus: true,
        errors: true,
        deployments: true,
      },
      dashboardLayout: 'grid',
      autoRefresh: true,
      refreshInterval: 30,
    },
  });

  // Initialize form data when userProfile changes
  useEffect(() => {
    if (userProfile && isOpen) {
      setFormData({
        displayName: userProfile.displayName || user?.username || '',
        avatar: userProfile.avatar || '',
        preferences: {
          ...userProfile.preferences,
          // Ensure all fields have defaults
          notificationSettings: {
            serviceStatus: true,
            errors: true,
            deployments: true,
            ...userProfile.preferences.notificationSettings,
          },
        },
      });
    } else if (user && isOpen && !userProfile) {
      // Set defaults if no profile exists yet
      setFormData({
        displayName: user.username,
        avatar: '',
        preferences: {
          theme: 'light',
          language: 'en',
          notificationSettings: {
            serviceStatus: true,
            errors: true,
            deployments: true,
          },
          dashboardLayout: 'grid',
          autoRefresh: true,
          refreshInterval: 30,
        },
      });
    }
  }, [userProfile, user, isOpen]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      await updateUserProfile(formData);
      addToast(toast.success('Success', 'Profile updated successfully!'));
      onClose();
    } catch (error) {
      console.error('Failed to update user profile:', error);
      addToast(toast.error('Error', 'Failed to update profile. Please try again.'));
    }
  };

  const handlePreferenceChange = (key: keyof UserPreferences, value: any) => {
    setFormData(prev => ({
      ...prev,
      preferences: {
        ...prev.preferences,
        [key]: value,
      },
    }));
  };

  const handleNotificationChange = (key: string, value: boolean) => {
    setFormData(prev => ({
      ...prev,
      preferences: {
        ...prev.preferences,
        notificationSettings: {
          ...prev.preferences.notificationSettings,
          [key]: value,
        },
      },
    }));
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-600 flex items-center justify-between flex-shrink-0">
          <div className="flex items-center gap-3">
            <User className="h-6 w-6 text-blue-600" />
            <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
              User Profile & Settings
            </h2>
          </div>
          <Button variant="ghost" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="flex flex-col flex-1 min-h-0">
          <div className="flex-1 overflow-y-auto p-6 space-y-6">
            {/* Basic Information */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 flex items-center gap-2">
                <User className="h-5 w-5" />
                Basic Information
              </h3>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Username
                  </label>
                  <Input
                    type="text"
                    value={user?.username || ''}
                    disabled
                    className="bg-gray-100 dark:bg-gray-700"
                  />
                  <p className="text-xs text-gray-500 mt-1">Username cannot be changed</p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Email
                  </label>
                  <Input
                    type="email"
                    value={user?.email || ''}
                    disabled
                    className="bg-gray-100 dark:bg-gray-700"
                  />
                  <p className="text-xs text-gray-500 mt-1">Email cannot be changed</p>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Display Name
                </label>
                <Input
                  type="text"
                  value={formData.displayName}
                  onChange={(e) => setFormData(prev => ({ ...prev, displayName: e.target.value }))}
                  placeholder="How should we display your name?"
                  className="w-full"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Avatar URL
                </label>
                <Input
                  type="url"
                  value={formData.avatar}
                  onChange={(e) => setFormData(prev => ({ ...prev, avatar: e.target.value }))}
                  placeholder="https://example.com/avatar.jpg"
                  className="w-full"
                />
                <p className="text-xs text-gray-500 mt-1">Link to your profile picture</p>
              </div>
            </div>

            {/* Theme & Display */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 flex items-center gap-2">
                <Monitor className="h-5 w-5" />
                Theme & Display
              </h3>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Theme
                  </label>
                  <select
                    value={formData.preferences.theme}
                    onChange={(e) => handlePreferenceChange('theme', e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-gray-100"
                  >
                    <option value="light">Light</option>
                    <option value="dark">Dark</option>
                    <option value="system">System</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Dashboard Layout
                  </label>
                  <select
                    value={formData.preferences.dashboardLayout}
                    onChange={(e) => handlePreferenceChange('dashboardLayout', e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-gray-100"
                  >
                    <option value="grid">Grid</option>
                    <option value="list">List</option>
                    <option value="compact">Compact</option>
                  </select>
                </div>
              </div>
            </div>

            {/* Auto Refresh */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 flex items-center gap-2">
                <RefreshCw className="h-5 w-5" />
                Auto Refresh
              </h3>
              
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="autoRefresh"
                  checked={formData.preferences.autoRefresh}
                  onChange={(e) => handlePreferenceChange('autoRefresh', e.target.checked)}
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <label htmlFor="autoRefresh" className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  Enable auto refresh
                </label>
              </div>

              {formData.preferences.autoRefresh && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Refresh Interval (seconds)
                  </label>
                  <Input
                    type="number"
                    min="5"
                    max="300"
                    value={formData.preferences.refreshInterval}
                    onChange={(e) => handlePreferenceChange('refreshInterval', parseInt(e.target.value))}
                    className="w-32"
                  />
                </div>
              )}
            </div>

            {/* Notifications */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 flex items-center gap-2">
                <Bell className="h-5 w-5" />
                Notification Settings
              </h3>
              
              <div className="space-y-3">
                {Object.entries(formData.preferences.notificationSettings).map(([key, value]) => (
                  <div key={key} className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      id={key}
                      checked={value}
                      onChange={(e) => handleNotificationChange(key, e.target.checked)}
                      className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                    />
                    <label htmlFor={key} className="text-sm font-medium text-gray-700 dark:text-gray-300">
                      {key === 'serviceStatus' && 'Service Status Changes'}
                      {key === 'errors' && 'Error Notifications'}
                      {key === 'deployments' && 'Deployment Updates'}
                    </label>
                  </div>
                ))}
              </div>
            </div>

            {/* Language */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 flex items-center gap-2">
                <Globe className="h-5 w-5" />
                Language & Region
              </h3>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Language
                </label>
                <select
                  value={formData.preferences.language}
                  onChange={(e) => handlePreferenceChange('language', e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-gray-100"
                >
                  <option value="en">English</option>
                  <option value="es">Español</option>
                  <option value="fr">Français</option>
                  <option value="de">Deutsch</option>
                  <option value="zh">中文</option>
                </select>
              </div>
            </div>
          </div>

          {/* Footer */}
          <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 flex items-center justify-end gap-3 flex-shrink-0">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isUpdating}
              className="flex items-center gap-2"
            >
              {isUpdating ? (
                <div className="h-4 w-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              ) : (
                <Save className="h-4 w-4" />
              )}
              Save Changes
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}