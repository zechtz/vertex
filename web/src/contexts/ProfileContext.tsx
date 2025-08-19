import {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react";
import {
  ServiceProfile,
  UserProfile,
  CreateProfileRequest,
  UpdateProfileRequest,
  UserProfileUpdateRequest,
  ProfileContext as ProfileContextData,
} from "@/types";
import { useAuth } from "./AuthContext";
import { useTheme } from "./ThemeContext";

interface ProfileContextType {
  // User Profile
  userProfile: UserProfile | null;
  updateUserProfile: (req: UserProfileUpdateRequest) => Promise<void>;

  // Service Profiles
  serviceProfiles: ServiceProfile[];
  activeProfile: ServiceProfile | null;
  createProfile: (req: CreateProfileRequest) => Promise<ServiceProfile>;
  updateProfile: (
    id: string,
    req: UpdateProfileRequest,
  ) => Promise<ServiceProfile>;
  deleteProfile: (id: string) => Promise<void>;
  applyProfile: (id: string) => Promise<void>;
  setActiveProfile: (profile: ServiceProfile | null) => void;

  // Profile-scoped configuration
  activateProfile: (id: string) => Promise<void>;
  getProfileContext: (id: string) => Promise<ProfileContextData>;
  getProfileEnvVars: (id: string) => Promise<Record<string, string>>;
  setProfileEnvVar: (
    id: string,
    name: string,
    value: string,
    description?: string,
    isRequired?: boolean,
  ) => Promise<void>;
  deleteProfileEnvVar: (id: string, name: string) => Promise<void>;
  removeServiceFromProfile: (
    profileId: string,
    serviceName: string,
  ) => Promise<void>;

  // Loading states
  isLoading: boolean;
  isCreating: boolean;
  isUpdating: boolean;
  isDeleting: boolean;
  isApplying: boolean;

  // Refresh data
  refreshProfiles: () => Promise<void>;
  refreshUserProfile: () => Promise<void>;
}

const ProfileContext = createContext<ProfileContextType | undefined>(undefined);

export function useProfile() {
  const context = useContext(ProfileContext);
  if (context === undefined) {
    throw new Error("useProfile must be used within a ProfileProvider");
  }
  return context;
}

interface ProfileProviderProps {
  children: ReactNode;
}

export function ProfileProvider({ children }: ProfileProviderProps) {
  const { user, isAuthenticated } = useAuth();
  const { syncWithUserProfile } = useTheme();
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null);
  const [serviceProfiles, setServiceProfiles] = useState<ServiceProfile[]>([]);
  const [activeProfile, setActiveProfileState] =
    useState<ServiceProfile | null>(null);

  // Loading states
  const [isLoading, setIsLoading] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [isApplying, setIsApplying] = useState(false);

  // Get auth token for API calls
  const getAuthToken = () => {
    return localStorage.getItem("authToken");
  };

  // API Helper function
  const apiCall = async (url: string, options: RequestInit = {}) => {
    const token = getAuthToken();
    if (!token) {
      throw new Error("No authentication token");
    }

    const response = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
        ...options.headers,
      },
    });

    if (!response.ok) {
      // Handle authentication errors specifically
      if (response.status === 401) {
        console.warn("Authentication failed, token may be invalid");
        // You could trigger a logout here if needed
        // logout();
        throw new Error("Authentication failed. Please log in again.");
      }

      const errorText = await response.text();
      throw new Error(`API Error: ${response.status} - ${errorText}`);
    }

    if (response.status === 204) {
      return null; // No content
    }

    return response.json();
  };

  // User Profile Operations
  const fetchUserProfile = async () => {
    try {
      const profile = await apiCall("/api/user/profile");
      setUserProfile(profile);

      // Sync theme with user profile preferences
      if (profile?.preferences?.theme) {
        syncWithUserProfile(profile.preferences.theme);
      }
    } catch (error) {
      console.error("Failed to fetch user profile:", error);
      throw error;
    }
  };

  const updateUserProfile = async (req: UserProfileUpdateRequest) => {
    try {
      setIsUpdating(true);
      console.log("Updating user profile with request:", req);

      const updatedProfile = await apiCall("/api/user/profile", {
        method: "PUT",
        body: JSON.stringify(req),
      });

      console.log("User profile updated successfully:", updatedProfile);
      setUserProfile(updatedProfile);

      // Sync theme with updated profile preferences
      if (updatedProfile?.preferences?.theme) {
        syncWithUserProfile(updatedProfile.preferences.theme);
      }
    } catch (error) {
      console.error("Failed to update user profile:", error);
      throw error;
    } finally {
      setIsUpdating(false);
    }
  };

  // Set the active profile based on the isActive flag from backend
  const setActiveProfileFromProfiles = (profiles: ServiceProfile[]) => {
    console.log(
      "Setting active profile from profiles:",
      profiles.map((p: ServiceProfile) => ({
        name: p.name,
        isActive: p.isActive,
      })),
    );

    // Find the profile marked as active in the backend data
    const activeProfile = profiles.find((p) => p.isActive);
    if (activeProfile) {
      console.log("Found active profile from backend:", activeProfile.name);
      setActiveProfileState(activeProfile);
      return;
    }

    console.log(
      "No active profile found in backend data, clearing active profile state",
    );
    setActiveProfileState(null);
  };

  // Service Profile Operations
  const fetchServiceProfiles = async () => {
    try {
      setIsLoading(true);
      const profiles = await apiCall("/api/profiles");
      console.log(
        "Fetched profiles:",
        profiles?.map((p: ServiceProfile) => ({
          name: p.name,
          isActive: p.isActive,
        })),
      );
      setServiceProfiles(profiles || []);

      // Set the active profile based on the isActive flag from backend
      setActiveProfileFromProfiles(profiles || []);
    } catch (error) {
      console.error("Failed to fetch service profiles:", error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const createProfile = async (
    req: CreateProfileRequest,
  ): Promise<ServiceProfile> => {
    try {
      setIsCreating(true);
      const newProfile = await apiCall("/api/profiles", {
        method: "POST",
        body: JSON.stringify(req),
      });
      setServiceProfiles((prev) => [...prev, newProfile]);

      // Set as active if it's the default profile
      if (req.isDefault) {
        setActiveProfileState(newProfile);
      }

      return newProfile;
    } catch (error) {
      console.error("Failed to create profile:", error);
      throw error;
    } finally {
      setIsCreating(false);
    }
  };

  const updateProfile = async (
    id: string,
    req: UpdateProfileRequest,
  ): Promise<ServiceProfile> => {
    try {
      setIsUpdating(true);
      const updatedProfile = await apiCall(`/api/profiles/${id}`, {
        method: "PUT",
        body: JSON.stringify(req),
      });

      setServiceProfiles((prev) =>
        prev.map((profile) => (profile.id === id ? updatedProfile : profile)),
      );

      // Update active profile if it's the one being updated
      if (activeProfile?.id === id) {
        setActiveProfileState(updatedProfile);
      }

      // Set as active if it's now the default profile
      if (req.isDefault) {
        setActiveProfileState(updatedProfile);
      }

      return updatedProfile;
    } catch (error) {
      console.error("Failed to update profile:", error);
      throw error;
    } finally {
      setIsUpdating(false);
    }
  };

  const deleteProfile = async (id: string) => {
    try {
      setIsDeleting(true);
      await apiCall(`/api/profiles/${id}`, {
        method: "DELETE",
      });

      setServiceProfiles((prev) => prev.filter((profile) => profile.id !== id));

      // Clear active profile if it's the one being deleted
      if (activeProfile?.id === id) {
        const remaining = serviceProfiles.filter((p) => p.id !== id);
        const defaultProfile = remaining.find((p) => p.isDefault);
        setActiveProfileState(defaultProfile || remaining[0] || null);
      }
    } catch (error) {
      console.error("Failed to delete profile:", error);
      throw error;
    } finally {
      setIsDeleting(false);
    }
  };

  const applyProfile = async (id: string) => {
    try {
      setIsApplying(true);
      await apiCall(`/api/profiles/${id}/apply`, {
        method: "POST",
      });

      // Find and set the applied profile as active
      const appliedProfile = serviceProfiles.find((p) => p.id === id);
      if (appliedProfile) {
        setActiveProfileState(appliedProfile);
      }
    } catch (error) {
      console.error("Failed to apply profile:", error);
      throw error;
    } finally {
      setIsApplying(false);
    }
  };

  const setActiveProfile = (profile: ServiceProfile | null) => {
    setActiveProfileState(profile);
  };

  const refreshProfiles = async () => {
    await fetchServiceProfiles();
  };

  const refreshUserProfile = async () => {
    await fetchUserProfile();
  };

  // Profile-scoped configuration methods

  const activateProfile = async (id: string) => {
    try {
      console.log("Activating profile:", id);
      await apiCall(`/api/profiles/${id}/activate`, {
        method: "POST",
      });
      console.log("Profile activation API call successful");

      // Re-fetch profiles to get the updated isActive flags from backend
      // This ensures consistency between frontend and backend state
      await fetchServiceProfiles();
      console.log("Profiles re-fetched after activation");
    } catch (error) {
      console.error("Failed to activate profile:", error);
      throw error;
    }
  };

  const getProfileContext = async (id: string): Promise<ProfileContextData> => {
    try {
      return await apiCall(`/api/profiles/${id}/context`);
    } catch (error) {
      console.error("Failed to get profile context:", error);
      throw error;
    }
  };

  const getProfileEnvVars = async (
    id: string,
  ): Promise<Record<string, string>> => {
    try {
      return await apiCall(`/api/profiles/${id}/env-vars`);
    } catch (error) {
      console.error("Failed to get profile env vars:", error);
      throw error;
    }
  };

  const setProfileEnvVar = async (
    id: string,
    name: string,
    value: string,
    description: string = "",
    isRequired: boolean = false,
  ) => {
    try {
      await apiCall(`/api/profiles/${id}/env-vars`, {
        method: "POST",
        body: JSON.stringify({
          name,
          value,
          description,
          isRequired,
        }),
      });
    } catch (error) {
      console.error("Failed to set profile env var:", error);
      throw error;
    }
  };

  const deleteProfileEnvVar = async (id: string, name: string) => {
    try {
      await apiCall(
        `/api/profiles/${id}/env-vars/${encodeURIComponent(name)}`,
        {
          method: "DELETE",
        },
      );
    } catch (error) {
      console.error("Failed to delete profile env var:", error);
      throw error;
    }
  };

  const removeServiceFromProfile = async (
    profileId: string,
    serviceName: string,
  ) => {
    try {
      await apiCall(
        `/api/profiles/${profileId}/services/${encodeURIComponent(serviceName)}`,
        {
          method: "DELETE",
        },
      );
    } catch (error) {
      console.error("Failed to remove service from profile:", error);
      throw error;
    }
  };

  // Load data when user is authenticated
  useEffect(() => {
    if (isAuthenticated && user) {
      fetchUserProfile().catch(console.error);
      fetchServiceProfiles().catch(console.error);
    } else {
      // Clear data when not authenticated
      setUserProfile(null);
      setServiceProfiles([]);
      setActiveProfileState(null);
    }
  }, [isAuthenticated, user]);

  const value: ProfileContextType = {
    // User Profile
    userProfile,
    updateUserProfile,

    // Service Profiles
    serviceProfiles,
    activeProfile,
    createProfile,
    updateProfile,
    deleteProfile,
    applyProfile,
    setActiveProfile,

    // Profile-scoped configuration
    activateProfile,
    getProfileContext,
    getProfileEnvVars,
    setProfileEnvVar,
    deleteProfileEnvVar,
    removeServiceFromProfile,

    // Loading states
    isLoading,
    isCreating,
    isUpdating,
    isDeleting,
    isApplying,

    // Refresh data
    refreshProfiles,
    refreshUserProfile,
  };

  return (
    <ProfileContext.Provider value={value}>{children}</ProfileContext.Provider>
  );
}
