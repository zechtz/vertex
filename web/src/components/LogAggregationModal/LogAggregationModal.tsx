import { FileText } from "lucide-react";
import { Modal } from "@/components/ui/Modal";
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
  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      size="7xl"
      title={
        <span className="flex items-center gap-3">
          <span className="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
            <FileText className="h-6 w-6 text-blue-600" />
          </span>
          <span>
            Log Aggregation &amp; Search
            <span className="block text-sm font-normal text-gray-600 dark:text-gray-400">
              Search and analyze logs across all services
            </span>
          </span>
        </span>
      }
    >
      <div className="p-6">
        <LogSearch services={services} />
      </div>
    </Modal>
  );
}

export default LogAggregationModal;
