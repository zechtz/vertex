import { X, Network } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ServiceTopology } from '@/components/ServiceTopology/ServiceTopology';

interface ServiceTopologyModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function ServiceTopologyModal({ isOpen, onClose }: ServiceTopologyModalProps) {
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
          <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl">
            {/* Header */}
            <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
                  <Network className="h-6 w-6 text-purple-600" />
                </div>
                <div>
                  <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                    Service Topology
                  </h2>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Visualize your microservices architecture and dependencies
                  </p>
                </div>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={onClose}
                className="hover:bg-gray-100 dark:hover:bg-gray-700"
              >
                <X className="h-5 w-5" />
              </Button>
            </div>

            {/* Content */}
            <div className="p-6">
              <ServiceTopology />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default ServiceTopologyModal;