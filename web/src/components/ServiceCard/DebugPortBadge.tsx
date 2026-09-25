import { useState } from "react";
import { Bug, Check } from "lucide-react";

interface DebugPortBadgeProps {
  port: number;
}

/**
 * Shows where a debugger can attach to a running service. Clicking copies the
 * address, ready to paste into an IDE's remote debug configuration.
 */
export function DebugPortBadge({ port }: DebugPortBadgeProps) {
  const [copied, setCopied] = useState(false);
  const address = `localhost:${port}`;

  const copyAddress = async () => {
    try {
      await navigator.clipboard.writeText(address);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy debugger address:", err);
    }
  };

  return (
    <button
      type="button"
      onClick={copyAddress}
      title={`Debugger listening on ${address}. Click to copy.`}
      aria-label={`Copy debugger address ${address}`}
      className="inline-flex items-center gap-1 rounded-md border border-purple-200 dark:border-purple-800 bg-purple-50 dark:bg-purple-900/30 px-1.5 py-0.5 text-xs font-medium text-purple-700 dark:text-purple-300 hover:bg-purple-100 dark:hover:bg-purple-900/50 flex-shrink-0"
    >
      {copied ? <Check className="h-3 w-3" /> : <Bug className="h-3 w-3" />}
      {copied ? "Copied" : `:${port}`}
    </button>
  );
}
