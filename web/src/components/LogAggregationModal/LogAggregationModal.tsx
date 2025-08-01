import { X, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Service } from "@/types";
import LogSearch from "@/components/LogSearch/LogSearch";

interface LogAggregationModalProps {
  isOpen: boolean;
  onClose: () => void;
  services: Service[];
}

export function LogAggregationModal({
  isOpen,
  onClose,
  services = [],
}: LogAggregationModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-screen items-center justify-center p-4">
        {/* Backdrop */}
        <div className="fixed inset-0 bg-black/50" onClick={onClose} />

        {/* Modal */}
        <div className="relative w-full max-w-7xl max-h-[90vh] overflow-y-auto">
          <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl">
            {/* Header */}
            <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                  <FileText className="h-6 w-6 text-blue-600" />
                </div>
                <div>
                  <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                    Log Aggregation & Search
                  </h2>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Search and analyze logs across all services
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
              <LogSearch services={services} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default LogAggregationModal;
