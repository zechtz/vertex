import { useState } from "react";
import {
  Plus,
  Settings,
  Play,
  Edit,
  Trash2,
  Star,
  Users,
  Server,
  Clock,
  Zap,
  Check,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { useProfile } from "@/contexts/ProfileContext";
import { useToast, toast } from "@/components/ui/toast";
import { ServiceProfile } from "@/types";
import { CreateProfileModal } from "./CreateProfileModal";
import { EditProfileModal } from "./EditProfileModal";
import { UserProfileModal } from "./UserProfileModal";
import { ProfileEnvManager } from "../ProfileEnvManager/ProfileEnvManager";
import { ProfileServiceManager } from "../ProfileServiceManager/ProfileServiceManager";

interface ProfileManagementProps {
  isOpen: boolean;
  onClose: () => void;
  onProfileUpdated?: () => void;
}

export function ProfileManagement({
  isOpen,
  onClose,
  onProfileUpdated,
}: ProfileManagementProps) {
  const {
    serviceProfiles,
    activeProfile,
    isLoading,
    isApplying,
    applyProfile,
    deleteProfile,
    activateProfile,
  } = useProfile();
  const { addToast } = useToast();

  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showUserProfileModal, setShowUserProfileModal] = useState(false);
  const [showEnvManager, setShowEnvManager] = useState(false);
  const [showServiceManager, setShowServiceManager] = useState(false);
  const [editingProfile, setEditingProfile] = useState<ServiceProfile | null>(
    null,
  );
  const [envManagerProfile, setEnvManagerProfile] =
    useState<ServiceProfile | null>(null);
  const [serviceManagerProfile, setServiceManagerProfile] =
    useState<ServiceProfile | null>(null);
  const [deletingProfile, setDeletingProfile] = useState<string | null>(null);
  const [activatingId, setActivatingId] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleEditProfile = (profile: ServiceProfile) => {
    setEditingProfile(profile);
    setShowEditModal(true);
  };

  const handleDeleteProfile = async (profileId: string) => {
    if (
      window.confirm(
        "Are you sure you want to delete this profile? This action cannot be undone.",
      )
    ) {
      try {
        setDeletingProfile(profileId);
        await deleteProfile(profileId);
      } catch (error) {
        console.error("Failed to delete profile:", error);
        alert("Failed to delete profile. Please try again.");
      } finally {
        setDeletingProfile(null);
      }
    }
  };

  const handleApplyProfile = async (profileId: string) => {
    try {
      await applyProfile(profileId);
      addToast(
        toast.success(
          "Success",
          "Profile applied successfully! Services are being started.",
        ),
      );
    } catch (error) {
      console.error("Failed to apply profile:", error);
      addToast(
        toast.error("Error", "Failed to apply profile. Please try again."),
      );
    }
  };

  const handleActivateProfile = async (profileId: string) => {
    try {
      setActivatingId(profileId);
      await activateProfile(profileId);
      addToast(toast.success("Success", "Profile activated successfully!"));
    } catch (error) {
      console.error("Failed to activate profile:", error);
      addToast(
        toast.error("Error", "Failed to activate profile. Please try again."),
      );
    } finally {
      setActivatingId(null);
    }
  };

  const handleManageEnvVars = (profile: ServiceProfile) => {
    setEnvManagerProfile(profile);
    setShowEnvManager(true);
  };

  const handleManageServices = (profile: ServiceProfile) => {
    setServiceManagerProfile(profile);
    setShowServiceManager(true);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getProfileStatusColor = (profile: ServiceProfile) => {
    if (profile.isDefault)
      return "text-yellow-600 dark:text-yellow-400 bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-700";
    if (activeProfile?.id === profile.id)
      return "text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-700";
    return "text-gray-600 dark:text-gray-400 bg-gray-50 dark:bg-gray-800/50 border-gray-200 dark:border-gray-700";
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-6xl h-[90vh] flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-600 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Users className="h-6 w-6 text-blue-600" />
            <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
              Profile Management
            </h2>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowUserProfileModal(true)}
              className="flex items-center gap-2"
            >
              <Settings className="h-4 w-4" />
              User Settings
            </Button>
            <Button
              onClick={() => setShowCreateModal(true)}
              className="flex items-center gap-2"
            >
              <Plus className="h-4 w-4" />
              Create Profile
            </Button>
            <Button variant="ghost" onClick={onClose}>
              ×
            </Button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-hidden">
          {isLoading ? (
            <div className="flex items-center justify-center h-full">
              <div className="text-center">
                <div className="h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
                <p className="text-gray-600 dark:text-gray-400">
                  Loading profiles...
                </p>
              </div>
            </div>
          ) : serviceProfiles.length === 0 ? (
            <div className="flex items-center justify-center h-full">
              <div className="text-center">
                <Users className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
                  No Profiles Yet
                </h3>
                <p className="text-gray-600 dark:text-gray-400 mb-4">
                  Create your first service profile to get started with
                  environment management.
                </p>
                <Button
                  onClick={() => setShowCreateModal(true)}
                  className="flex items-center gap-2"
                >
                  <Plus className="h-4 w-4" />
                  Create Your First Profile
                </Button>
              </div>
            </div>
          ) : (
            <div className="p-6 h-full overflow-y-auto">
              {/* Active Profile Summary */}
              {activeProfile && (
                <div className="mb-6 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-700 rounded-lg">
                  <div className="flex items-center gap-2 mb-3">
                    <Check className="h-5 w-5 text-blue-600" />
                    <h3 className="text-lg font-medium text-blue-900 dark:text-blue-100">
                      Currently Active Profile
                    </h3>
                  </div>
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <p className="font-medium text-blue-800 dark:text-blue-200">
                        {activeProfile.name}
                      </p>
                      <p className="text-sm text-blue-600 dark:text-blue-300">
                        {activeProfile.description}
                      </p>
                      <div className="flex items-center gap-4 mt-2 text-xs text-blue-600 dark:text-blue-400">
                        <span>
                          {activeProfile.services.length} services configured
                        </span>
                        {Object.keys(activeProfile.envVars || {}).length >
                          0 && (
                          <span>
                            {Object.keys(activeProfile.envVars).length} env
                            variables
                          </span>
                        )}
                        {activeProfile.projectsDir && (
                          <span>Projects: {activeProfile.projectsDir}</span>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleManageServices(activeProfile)}
                        className="flex items-center gap-2"
                      >
                        <Server className="h-4 w-4" />
                        Services
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleManageEnvVars(activeProfile)}
                        className="flex items-center gap-2"
                      >
                        <Zap className="h-4 w-4" />
                        Env Vars
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleEditProfile(activeProfile)}
                        className="flex items-center gap-2"
                      >
                        <Edit className="h-4 w-4" />
                        Edit
                      </Button>
                    </div>
                  </div>
                </div>
              )}

              {/* Profile Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {serviceProfiles.map((profile) => (
                  <div
                    key={profile.id}
                    className={`p-4 border rounded-lg transition-all hover:shadow-md ${getProfileStatusColor(profile)}`}
                  >
                    {/* Profile Header */}
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <h3 className="font-medium text-gray-900 dark:text-gray-100">
                            {profile.name}
                          </h3>
                          {profile.isDefault && (
                            <Star className="h-4 w-4 text-yellow-500 fill-current" />
                          )}
                          {activeProfile?.id === profile.id && (
                            <span className="px-2 py-1 bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-200 text-xs rounded-full">
                              Active
                            </span>
                          )}
                        </div>
                        <p className="text-sm text-gray-600 dark:text-gray-400 line-clamp-2">
                          {profile.description || "No description"}
                        </p>
                      </div>
                    </div>

                    {/* Profile Stats */}
                    <div className="flex items-center gap-4 mb-3 text-sm text-gray-600 dark:text-gray-400">
                      <div className="flex items-center gap-1">
                        <Server className="h-4 w-4" />
                        <span>{profile.services.length} services</span>
                      </div>
                      <div className="flex items-center gap-1">
                        <Clock className="h-4 w-4" />
                        <span>{formatDate(profile.updatedAt)}</span>
                      </div>
                    </div>

                    {/* Environment Variables */}
                    {Object.keys(profile.envVars || {}).length > 0 && (
                      <div className="mb-3">
                        <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">
                          Environment Variables:
                        </p>
                        <div className="flex flex-wrap gap-1">
                          {Object.keys(profile.envVars)
                            .slice(0, 3)
                            .map((key) => (
                              <span
                                key={key}
                                className="px-2 py-1 bg-gray-100 dark:bg-gray-700 text-xs rounded text-gray-600 dark:text-gray-300"
                              >
                                {key}
                              </span>
                            ))}
                          {Object.keys(profile.envVars).length > 3 && (
                            <span className="px-2 py-1 bg-gray-100 dark:bg-gray-700 text-xs rounded text-gray-600 dark:text-gray-300">
                              +{Object.keys(profile.envVars).length - 3} more
                            </span>
                          )}
                        </div>
                      </div>
                    )}

                    {/* Services List */}
                    <div className="mb-3">
                      <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">
                        Services:
                      </p>
                      <div className="flex flex-wrap gap-1">
                        {profile.services.slice(0, 3).map((service: any) => (
                          <span
                            key={service.id}
                            className="px-2 py-1 bg-blue-100 dark:bg-blue-900 text-xs rounded text-blue-600 dark:text-blue-300"
                          >
                            {service.name}
                          </span>
                        ))}
                        {profile.services.length > 3 && (
                          <span className="px-2 py-1 bg-blue-100 dark:bg-blue-900 text-xs rounded text-blue-600 dark:text-blue-300">
                            +{profile.services.length - 3} more
                          </span>
                        )}
                      </div>
                    </div>

                    {/* Actions */}
                    <div className="space-y-2">
                      {/* Primary Actions Row */}
                      <div className="flex items-center gap-2">
                        {!profile.isActive && (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleActivateProfile(profile.id)}
                            disabled={activatingId === profile.id}
                            className="flex-1 flex items-center gap-1"
                          >
                            {activatingId === profile.id ? (
                              <div className="h-3 w-3 border-2 border-blue-600 border-t-transparent rounded-full animate-spin" />
                            ) : (
                              <Settings className="h-3 w-3" />
                            )}
                            Activate
                          </Button>
                        )}
                        <Button
                          size="sm"
                          onClick={() => handleApplyProfile(profile.id)}
                          disabled={isApplying}
                          className={`${!profile.isActive ? "flex-1" : "flex-1"} flex items-center gap-1`}
                        >
                          {isApplying ? (
                            <div className="h-3 w-3 border-2 border-white border-t-transparent rounded-full animate-spin" />
                          ) : (
                            <Play className="h-3 w-3" />
                          )}
                          Apply & Start
                        </Button>
                      </div>

                      {/* Secondary Actions Row */}
                      <div className="flex items-center gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleManageServices(profile)}
                          className="flex-1 flex items-center gap-1"
                        >
                          <Server className="h-3 w-3" />
                          Services
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleManageEnvVars(profile)}
                          className="flex-1 flex items-center gap-1"
                        >
                          <Zap className="h-3 w-3" />
                          Env Vars
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleEditProfile(profile)}
                        >
                          <Edit className="h-3 w-3" />
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleDeleteProfile(profile.id)}
                          disabled={deletingProfile === profile.id}
                          className="text-red-600 hover:text-red-700 hover:bg-red-50"
                        >
                          {deletingProfile === profile.id ? (
                            <div className="h-3 w-3 border-2 border-red-600 border-t-transparent rounded-full animate-spin" />
                          ) : (
                            <Trash2 className="h-3 w-3" />
                          )}
                        </Button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 rounded-b-lg">
          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-600 dark:text-gray-400">
              {serviceProfiles.length} profile
              {serviceProfiles.length !== 1 ? "s" : ""} total
              {activeProfile && (
                <span className="ml-2">• Active: {activeProfile.name}</span>
              )}
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" onClick={onClose}>
                Close
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Modals */}
      <CreateProfileModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
      />

      <EditProfileModal
        isOpen={showEditModal}
        onClose={() => {
          setShowEditModal(false);
          setEditingProfile(null);
        }}
        profile={editingProfile}
      />

      <UserProfileModal
        isOpen={showUserProfileModal}
        onClose={() => setShowUserProfileModal(false)}
      />

      <ProfileEnvManager
        isOpen={showEnvManager}
        onClose={() => {
          setShowEnvManager(false);
          setEnvManagerProfile(null);
        }}
        profile={envManagerProfile}
      />

      <ProfileServiceManager
        isOpen={showServiceManager}
        onClose={() => {
          setShowServiceManager(false);
          setServiceManagerProfile(null);
        }}
        profile={serviceManagerProfile}
        onProfileUpdated={onProfileUpdated}
      />
    </div>
  );
}
