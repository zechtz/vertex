import { AlertTriangle, X, Trash2, UserMinus, Info } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Service, ServiceProfile } from "@/types";

interface ServiceActionModalProps {
  isOpen: boolean;
  onClose: () => void;
  service: Service | null;
  activeProfile: ServiceProfile | null;
  onRemoveFromProfile: (serviceName: string) => Promise<void>;
  onDeleteGlobally: (serviceName: string) => Promise<void>;
}

export function ServiceActionModal({
  isOpen,
  onClose,
  service,
  activeProfile,
  onRemoveFromProfile,
  onDeleteGlobally,
}: ServiceActionModalProps) {
  if (!isOpen || !service) return null;

  const handleRemoveFromProfile = async () => {
    await onRemoveFromProfile(service.id);
    onClose();
  };

  const handleDeleteGlobally = async () => {
    await onDeleteGlobally(service.id);
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-md">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-600 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <AlertTriangle className="h-6 w-6 text-orange-600" />
            <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
              Service Action
            </h2>
          </div>
          <Button variant="ghost" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Content */}
        <div className="p-6">
          <div className="mb-6">
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
              What would you like to do with "{service.name}"?
            </h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Choose how you want to handle this service:
            </p>
          </div>

          {/* Options */}
          <div className="space-y-4">
            {/* Remove from Profile Option */}
            {activeProfile && (
              <div className="p-4 border-2 border-blue-200 dark:border-blue-800 rounded-lg bg-blue-50 dark:bg-blue-900/20">
                <div className="flex items-start gap-3">
                  <UserMinus className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
                  <div className="flex-1">
                    <h4 className="font-medium text-blue-900 dark:text-blue-100">
                      Remove from Profile
                    </h4>
                    <p className="text-sm text-blue-700 dark:text-blue-300 mt-1">
                      Remove "{service.name}" from the "{activeProfile.name}"
                      profile only. The service will still exist and can be
                      added to other profiles.
                    </p>
                    <Button
                      onClick={handleRemoveFromProfile}
                      className="mt-3 bg-blue-600 hover:bg-blue-700 text-white"
                      size="sm"
                    >
                      <UserMinus className="h-4 w-4 mr-2" />
                      Remove from Profile
                    </Button>
                  </div>
                </div>
              </div>
            )}

            {/* Delete Globally Option */}
            <div className="p-4 border-2 border-red-200 dark:border-red-800 rounded-lg bg-red-50 dark:bg-red-900/20">
              <div className="flex items-start gap-3">
                <Trash2 className="h-5 w-5 text-red-600 flex-shrink-0 mt-0.5" />
                <div className="flex-1">
                  <h4 className="font-medium text-red-900 dark:text-red-100">
                    Delete Service Completely
                  </h4>
                  <p className="text-sm text-red-700 dark:text-red-300 mt-1">
                    Permanently delete "{service.name}" from the entire system.
                    This will remove it from all profiles and cannot be undone.
                  </p>
                  <Button
                    onClick={handleDeleteGlobally}
                    variant="outline"
                    className="mt-3 border-red-300 text-red-700 hover:bg-red-100 dark:border-red-600 dark:text-red-400 dark:hover:bg-red-900/30"
                    size="sm"
                  >
                    <Trash2 className="h-4 w-4 mr-2" />
                    Delete Permanently
                  </Button>
                </div>
              </div>
            </div>
          </div>

          {/* Info */}
          <div className="mt-6 p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
            <div className="flex items-start gap-2">
              <Info className="h-4 w-4 text-gray-500 flex-shrink-0 mt-0.5" />
              <p className="text-xs text-gray-600 dark:text-gray-400">
                <strong>Tip:</strong> If you're unsure, choose "Remove from
                Profile" first. You can always delete the service completely
                later if needed.
              </p>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 rounded-b-lg">
          <div className="flex justify-end">
            <Button variant="outline" onClick={onClose}>
              Cancel
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
