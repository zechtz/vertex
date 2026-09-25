import { GitBranch } from 'lucide-react';
import { Modal } from '@/components/ui/Modal';
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
  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      size="7xl"
      title={
        <span className="flex items-center gap-3">
          <span className="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
            <GitBranch className="h-6 w-6 text-blue-600" />
          </span>
          <span>
            Dependency Management
            <span className="block text-sm font-normal text-gray-600 dark:text-gray-400">
              Configure service dependencies and startup ordering
            </span>
          </span>
        </span>
      }
    >
      <div className="p-6">
        <DependencyManager services={services} />
      </div>
    </Modal>
  );
}

export default DependencyManagerModal;
