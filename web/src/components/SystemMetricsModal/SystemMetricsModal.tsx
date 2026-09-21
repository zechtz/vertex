import { Modal } from '@/components/ui/Modal';
import SystemMetrics from '@/components/SystemMetrics/SystemMetrics';
import { Service } from '@/types';

interface SystemMetricsModalProps {
  isOpen: boolean;
  onClose: () => void;
  services: Service[];
}

export function SystemMetricsModal({ isOpen, onClose, services }: SystemMetricsModalProps) {
  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="System Resource Monitoring"
      size="6xl"
    >
      <div className="p-6">
        <SystemMetrics services={services} />
      </div>
    </Modal>
  );
}

export default SystemMetricsModal;
