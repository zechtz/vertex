import { useState, useEffect } from "react";
import { GitBranch, Check, Loader2 } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { useToast, toast } from "@/components/ui/toast";

interface GitBranchSwitcherProps {
  serviceId: string;
  serviceName: string;
  currentBranch: string;
  isServiceRunning: boolean;
}

export function GitBranchSwitcher({
  serviceId,
  serviceName,
  currentBranch,
  isServiceRunning,
}: GitBranchSwitcherProps) {
  const [branches, setBranches] = useState<string[]>([]);
  const [selectedBranch, setSelectedBranch] = useState(currentBranch);
  const [isLoading, setIsLoading] = useState(false);
  const [isSwitching, setIsSwitching] = useState(false);
  const { addToast } = useToast();

  useEffect(() => {
    setSelectedBranch(currentBranch);
  }, [currentBranch]);

  useEffect(() => {
    fetchBranches();
  }, [serviceId]);

  const fetchBranches = async () => {
    try {
      setIsLoading(true);
      const token = localStorage.getItem("authToken");
      if (!token) {
        throw new Error("No authentication token");
      }

      const response = await fetch(`/api/services/${serviceId}/git/branches`, {
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        if (response.status === 404 || response.status === 500) {
          // Service is not a git repository, silently skip
          return;
        }
        throw new Error(`Failed to fetch branches: ${response.statusText}`);
      }

      const data = await response.json();
      setBranches(data.branches || []);
    } catch (error) {
      // Silently ignore errors for non-git services
      console.log(`Service ${serviceName} is not a git repository`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSwitchBranch = async () => {
    if (selectedBranch === currentBranch) {
      return;
    }

    if (isServiceRunning) {
      addToast(
        toast.error(
          "Cannot switch branches",
          `Please stop ${serviceName} before switching branches`,
        ),
      );
      return;
    }

    try {
      setIsSwitching(true);
      const token = localStorage.getItem("authToken");
      if (!token) {
        throw new Error("No authentication token");
      }

      const response = await fetch(`/api/services/${serviceId}/git/switch`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ branch: selectedBranch }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Failed to switch branch`);
      }

      await response.json();
      addToast(
        toast.success(
          "Branch switched",
          `${serviceName} is now on branch '${selectedBranch}'`,
        ),
      );

      // Refresh the page to update the service info
      window.location.reload();
    } catch (error) {
      console.error("Failed to switch branch:", error);
      addToast(
        toast.error(
          "Failed to switch branch",
          error instanceof Error ? error.message : "Unknown error",
        ),
      );
      setSelectedBranch(currentBranch); // Reset to current branch
    } finally {
      setIsSwitching(false);
    }
  };

  // Don't render if no branches or not a git repo
  if (branches.length === 0 && !isLoading) {
    return null;
  }

  return (
    <div className="flex items-center space-x-2">
      <GitBranch className="h-4 w-4 text-muted-foreground" />
      <Select
        value={selectedBranch}
        onValueChange={setSelectedBranch}
        disabled={isLoading || isSwitching || isServiceRunning}
      >
        <SelectTrigger className="w-[180px] h-8">
          {isLoading ? (
            <span className="flex items-center">
              <Loader2 className="mr-2 h-3 w-3 animate-spin" />
              Loading...
            </span>
          ) : (
            <SelectValue placeholder="Select branch" />
          )}
        </SelectTrigger>
        <SelectContent>
          {branches.map((branch) => (
            <SelectItem key={branch} value={branch}>
              <div className="flex items-center">
                {branch === currentBranch && (
                  <Check className="mr-2 h-3 w-3 text-green-500" />
                )}
                {branch}
              </div>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {selectedBranch !== currentBranch && (
        <Button
          onClick={handleSwitchBranch}
          disabled={isSwitching || isServiceRunning}
          size="sm"
          variant="outline"
        >
          {isSwitching ? (
            <>
              <Loader2 className="mr-2 h-3 w-3 animate-spin" />
              Switching...
            </>
          ) : (
            <>
              <Check className="mr-2 h-3 w-3" />
              Switch
            </>
          )}
        </Button>
      )}
    </div>
  );
}
