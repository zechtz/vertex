import { useState, useCallback } from "react";
import { Service } from "@/types";
import { ServiceOperations } from "@/services/serviceOperations";
import { useToast, toast } from "@/components/ui/toast";
import { useConfirm, confirmDialogs } from "@/components/ui/confirm-dialog";

export function useLogsOperations() {
  const { addToast } = useToast();
  const { showConfirm } = useConfirm();

  const [copied, setCopied] = useState(false);
  const [isCopyingLogs, setIsCopyingLogs] = useState(false);
  const [isClearingLogs, setIsClearingLogs] = useState(false);

  const copyLogsToClipboard = useCallback(
    async (
      selectedService: Service | null,
      selectedLevels: string[] = ["INFO", "WARN", "ERROR"],
    ) => {
      if (!selectedService || selectedService.logs.length === 0) return;

      try {
        setIsCopyingLogs(true);
        const filteredLogs = selectedService.logs.filter((log) =>
          selectedLevels.includes(log.level),
        );
        const logsText = filteredLogs
          .map(
            (log) =>
              `${new Date(log.timestamp).toLocaleString()} [${log.level}] ${log.message}`,
          )
          .join("\n");

        await navigator.clipboard.writeText(logsText);
        setCopied(true);
        addToast(
          toast.success(
            "Logs copied",
            `Copied ${filteredLogs.length} log entries to clipboard`,
          ),
        );
        setTimeout(() => setCopied(false), 2000);
      } catch (error) {
        console.error("Failed to copy logs:", error);
        addToast(
          toast.error(
            "Failed to copy logs",
            error instanceof Error
              ? error.message
              : "An unexpected error occurred",
          ),
        );
      } finally {
        setIsCopyingLogs(false);
      }
    },
    [addToast],
  );

  const clearLogs = useCallback(
    async (selectedService: Service | null) => {
      if (!selectedService) return;

      const confirmed = await showConfirm(
        confirmDialogs.clearLogs(selectedService.id),
      );
      if (!confirmed) return;

      try {
        setIsClearingLogs(true);
        const result = await ServiceOperations.clearServiceLogs(
          selectedService.id,
        );

        if (result.success) {
          addToast(toast.success("Logs cleared", result.message!));
        } else {
          addToast(toast.error("Failed to clear logs", result.error!));
        }
      } catch (error) {
        console.error("Failed to clear logs:", error);
        addToast(
          toast.error(
            "Failed to clear logs",
            error instanceof Error
              ? error.message
              : "An unexpected error occurred",
          ),
        );
      } finally {
        setIsClearingLogs(false);
      }
    },
    [addToast, showConfirm],
  );

  return {
    copied,
    isCopyingLogs,
    isClearingLogs,
    copyLogsToClipboard,
    clearLogs,
  };
}
