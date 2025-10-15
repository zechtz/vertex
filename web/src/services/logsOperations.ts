import { Service } from "@/types";

export class LogsOperations {
  static async copyLogsToClipboard(
    selectedService: Service,
    selectedLevels: string[] = ["INFO", "WARN", "ERROR"]
  ): Promise<{ success: boolean; count: number; error?: string }> {
    try {
      if (!selectedService || selectedService.logs.length === 0) {
        return { success: false, count: 0, error: "No logs to copy" };
      }

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
      return { success: true, count: filteredLogs.length };
    } catch (error) {
      return {
        success: false,
        count: 0,
        error: error instanceof Error ? error.message : "Failed to copy logs",
      };
    }
  }

  static async clearServiceLogs(serviceName: string): Promise<{ success: boolean; error?: string }> {
    try {
      const token = localStorage.getItem("authToken");
      if (!token) {
        throw new Error("No authentication token");
      }

      const response = await fetch(`/api/services/${serviceName}/logs`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
      });
      
      if (!response.ok) {
        throw new Error(
          `Failed to clear logs: ${response.status} ${response.statusText}`,
        );
      }
      
      return { success: true };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "Failed to clear logs",
      };
    }
  }
}