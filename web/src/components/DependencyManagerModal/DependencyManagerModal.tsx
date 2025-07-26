import { X, GitBranch } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { DependencyManager } from '@/components/DependencyManager/DependencyManager';
import { Service } from '@/types';

interface DependencyManagerModalProps {
  isOpen: boolean;
  onClose: () => void;
  services: Service[];
}

export function DependencyManagerModal({ 
  isOpen, 
  onClose, 
  services 
}: DependencyManagerModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-screen items-center justify-center p-4">
        {/* Backdrop */}
        <div 
          className="fixed inset-0 bg-black/50" 
          onClick={onClose}
        />
        
        {/* Modal */}
        <div className="relative w-full max-w-7xl max-h-[95vh] overflow-y-auto">
          <div className="relative bg-white rounded-lg shadow-xl">
            {/* Header */}
            <div className="flex items-center justify-between p-6 border-b border-gray-200">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-blue-100 rounded-lg">
                  <GitBranch className="h-6 w-6 text-blue-600" />
                </div>
                <div>
                  <h2 className="text-xl font-semibold text-gray-900">
                    Dependency Management
                  </h2>
                  <p className="text-sm text-gray-600">
                    Configure service dependencies and startup ordering
                  </p>
                </div>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={onClose}
                className="hover:bg-gray-100"
              >
                <X className="h-5 w-5" />
              </Button>
            </div>

            {/* Content */}
            <div className="p-6">
              <DependencyManager services={services} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default DependencyManagerModal;