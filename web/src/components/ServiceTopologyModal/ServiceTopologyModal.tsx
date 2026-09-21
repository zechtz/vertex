import { Network } from "lucide-react";
import { Modal } from "@/components/ui/Modal";
import { ServiceTopology } from "@/components/ServiceTopology/ServiceTopology";

interface ServiceTopologyModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function ServiceTopologyModal({
  isOpen,
  onClose,
}: ServiceTopologyModalProps) {
  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      size="7xl"
      title={
        <span className="flex items-center gap-3">
          <span className="p-2 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
            <Network className="h-6 w-6 text-purple-600" />
          </span>
          <span>
            Service Topology
            <span className="block text-sm font-normal text-gray-600 dark:text-gray-400">
              Visualize your microservices architecture and dependencies
            </span>
          </span>
        </span>
      }
    >
      <div className="p-6">
        <ServiceTopology />
      </div>
    </Modal>
  );
}

export default ServiceTopologyModal;
