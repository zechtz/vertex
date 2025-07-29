import { X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import SystemMetrics from '@/components/SystemMetrics/SystemMetrics';
import { Service } from '@/types';

interface SystemMetricsModalProps {
  isOpen: boolean;
  onClose: () => void;
  services: Service[];
}

export function SystemMetricsModal({ isOpen, onClose, services }: SystemMetricsModalProps) {
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
        <div className="relative w-full max-w-6xl max-h-[90vh] overflow-y-auto">
          <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl">
            {/* Header */}
            <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                System Resource Monitoring
              </h2>
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
              <SystemMetrics services={services} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default SystemMetricsModal;