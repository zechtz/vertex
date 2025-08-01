import { useModalManager } from "./useModalManager";
import { ServiceProfile } from "@/types";

export function useProfileModals() {
  const modalManager = useModalManager([
    "createProfile",
    "editProfile",
    "userProfile",
    "profileConfig",
  ]);

  const openCreateProfile = () => {
    modalManager.openModal("createProfile");
  };

  const openEditProfile = (profile: ServiceProfile) => {
    modalManager.openModal("editProfile", profile);
  };

  const openUserProfile = () => {
    modalManager.openModal("userProfile");
  };

  const openProfileConfig = (profile: ServiceProfile) => {
    modalManager.openModal("profileConfig", profile);
  };

  return {
    // Modal states
    isCreateProfileOpen: modalManager.isModalOpen("createProfile"),
    isEditProfileOpen: modalManager.isModalOpen("editProfile"),
    isUserProfileOpen: modalManager.isModalOpen("userProfile"),
    isProfileConfigOpen: modalManager.isModalOpen("profileConfig"),

    // Modal data
    editProfileData: modalManager.getModalData<ServiceProfile>("editProfile"),
    profileConfigData:
      modalManager.getModalData<ServiceProfile>("profileConfig"),

    // Actions
    openCreateProfile,
    openEditProfile,
    openUserProfile,
    openProfileConfig,

    // Close handlers
    closeCreateProfile: () => modalManager.closeModal("createProfile"),
    closeEditProfile: () => modalManager.closeModal("editProfile"),
    closeUserProfile: () => modalManager.closeModal("userProfile"),
    closeProfileConfig: () => modalManager.closeModal("profileConfig"),

    // Utility
    closeAllProfileModals: modalManager.closeAllModals,
    modalManager,
  };
}
