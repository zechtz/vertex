import { useState, useEffect, useMemo } from "react";
import { GitBranch, Check, Loader2, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Modal } from "@/components/ui/Modal";
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
  const [isLoading, setIsLoading] = useState(false);
  const [isSwitching, setIsSwitching] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const { addToast } = useToast();

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

  const handleSwitchBranch = async (branch: string) => {
    if (branch === currentBranch) {
      setIsModalOpen(false);
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
        body: JSON.stringify({ branch }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Failed to switch branch`);
      }

      await response.json();
      addToast(
        toast.success(
          "Branch switched",
          `${serviceName} is now on branch '${branch}'`,
        ),
      );

      setIsModalOpen(false);
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
    } finally {
      setIsSwitching(false);
    }
  };

  // Filter branches based on search query
  const filteredBranches = useMemo(() => {
    if (!searchQuery.trim()) {
      return branches;
    }
    const query = searchQuery.toLowerCase();
    return branches.filter((branch) =>
      branch.toLowerCase().includes(query)
    );
  }, [branches, searchQuery]);

  // Don't render if no branches or not a git repo
  if (branches.length === 0 && !isLoading) {
    return null;
  }

  return (
    <>
      <Button
        onClick={() => setIsModalOpen(true)}
        disabled={isLoading || isServiceRunning}
        variant="outline"
        size="sm"
        className="h-7 w-auto min-w-[140px] px-2 border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-xs font-medium hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
      >
        <div className="flex items-center gap-1.5">
          <GitBranch className="h-3.5 w-3.5 text-gray-500 dark:text-gray-400" />
          {isLoading ? (
            <span className="flex items-center gap-1.5">
              <Loader2 className="h-3 w-3 animate-spin" />
              <span className="text-gray-600 dark:text-gray-300">Loading...</span>
            </span>
          ) : (
            <span className="text-gray-700 dark:text-gray-200 truncate max-w-[120px]">
              {currentBranch || "Select branch"}
            </span>
          )}
        </div>
      </Button>

      <Modal
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
          setSearchQuery("");
        }}
        title={`Switch Git Branch - ${serviceName}`}
        size="lg"
        contentClassName="p-0"
      >
        <div className="flex flex-col h-full">
          {/* Search Input */}
          <div className="p-4 border-b border-gray-200 dark:border-gray-700">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <Input
                type="text"
                placeholder="Search branches..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
                autoFocus
              />
            </div>
            {searchQuery && (
              <div className="mt-2 text-xs text-gray-500 dark:text-gray-400">
                Found {filteredBranches.length} of {branches.length} branches
              </div>
            )}
          </div>

          {/* Branch List */}
          <div className="flex-1 overflow-y-auto max-h-[60vh]">
            {filteredBranches.length === 0 ? (
              <div className="p-8 text-center text-gray-500 dark:text-gray-400">
                {searchQuery ? (
                  <>
                    No branches found matching "<strong>{searchQuery}</strong>"
                  </>
                ) : (
                  "No branches available"
                )}
              </div>
            ) : (
              <div className="divide-y divide-gray-200 dark:divide-gray-700">
                {filteredBranches.map((branch) => {
                  const isCurrent = branch === currentBranch;

                  return (
                    <button
                      key={branch}
                      onClick={() => handleSwitchBranch(branch)}
                      disabled={isSwitching}
                      className={`w-full px-6 py-3 flex items-center gap-3 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors text-left ${
                        isCurrent
                          ? "bg-green-50 dark:bg-green-900/20"
                          : ""
                      } ${isSwitching ? "opacity-50 cursor-not-allowed" : ""}`}
                    >
                      {/* Checkmark for current branch */}
                      <div className="flex items-center justify-center w-5 flex-shrink-0">
                        {isCurrent && (
                          <Check className="h-4 w-4 text-green-600 dark:text-green-500" />
                        )}
                      </div>

                      {/* Branch name */}
                      <div className="flex-1 min-w-0">
                        <div
                          className={`font-medium text-sm break-all ${
                            isCurrent
                              ? "text-green-700 dark:text-green-400"
                              : "text-gray-900 dark:text-gray-100"
                          }`}
                        >
                          {branch}
                        </div>
                        {isCurrent && (
                          <div className="text-xs text-green-600 dark:text-green-500 font-medium mt-0.5">
                            Current branch
                          </div>
                        )}
                      </div>

                      {/* Switch indicator for non-current branches */}
                      {!isCurrent && (
                        <div className="text-xs text-gray-400 dark:text-gray-500 flex-shrink-0">
                          Click to switch
                        </div>
                      )}
                    </button>
                  );
                })}
              </div>
            )}
          </div>

          {/* Footer with actions */}
          {isServiceRunning && (
            <div className="p-4 border-t border-gray-200 dark:border-gray-700 bg-yellow-50 dark:bg-yellow-900/20">
              <div className="flex items-start gap-2">
                <div className="text-yellow-600 dark:text-yellow-500 mt-0.5">
                  ⚠️
                </div>
                <div className="text-xs text-yellow-700 dark:text-yellow-400">
                  <strong>Service is running</strong>
                  <br />
                  Please stop {serviceName} before switching branches
                </div>
              </div>
            </div>
          )}
        </div>
      </Modal>
    </>
  );
}
