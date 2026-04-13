import { AlertCircle, ArrowUp, ArrowDown, Check } from "lucide-react";

interface GitStatusBadgeProps {
  hasUncommitted: boolean;
  commitsAhead: number;
  commitsBehind: number;
  isClean: boolean;
}

export function GitStatusBadge({
  hasUncommitted,
  commitsAhead,
  commitsBehind,
  isClean,
}: GitStatusBadgeProps) {
  // Priority order: uncommitted > behind > ahead > clean
  // This ensures we show the most important status first

  if (hasUncommitted) {
    return (
      <div
        className="flex items-center gap-1 px-1.5 py-0.5 rounded-full bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400"
        title="Uncommitted changes"
      >
        <AlertCircle className="h-3 w-3" />
        <span className="text-[10px] font-medium">Uncommitted</span>
      </div>
    );
  }

  if (commitsBehind > 0) {
    return (
      <div
        className="flex items-center gap-1 px-1.5 py-0.5 rounded-full bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-400"
        title={`${commitsBehind} commit${commitsBehind > 1 ? "s" : ""} behind remote`}
      >
        <ArrowDown className="h-3 w-3" />
        <span className="text-[10px] font-medium">
          {commitsBehind} behind
        </span>
      </div>
    );
  }

  if (commitsAhead > 0) {
    return (
      <div
        className="flex items-center gap-1 px-1.5 py-0.5 rounded-full bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400"
        title={`${commitsAhead} commit${commitsAhead > 1 ? "s" : ""} ahead of remote`}
      >
        <ArrowUp className="h-3 w-3" />
        <span className="text-[10px] font-medium">
          {commitsAhead} ahead
        </span>
      </div>
    );
  }

  if (isClean) {
    return (
      <div
        className="flex items-center gap-1 px-1.5 py-0.5 rounded-full bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400"
        title="Clean and synced with remote"
      >
        <Check className="h-3 w-3" />
        <span className="text-[10px] font-medium">Clean</span>
      </div>
    );
  }

  // No status to show (not a git repo or no remote)
  return null;
}
