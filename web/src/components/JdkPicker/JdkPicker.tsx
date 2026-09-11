import { useEffect, useState } from "react";
import { Coffee, Loader } from "lucide-react";
import { Button } from "@/components/ui/button";
import { SystemApi } from "@/services/systemApi";
import type { EnvVar, InstalledJdk } from "@/types";

interface JdkPickerProps {
  serviceId: string;
  /** The service's current variables, preserved when JAVA_HOME is written. */
  envVars: Record<string, EnvVar>;
  /** Applied after JAVA_HOME is set, so the fix can be tried immediately. */
  onApplied: () => void;
}

/**
 * Offers the JDKs installed on this machine so a service that failed on a JDK
 * mismatch can be pinned to a different one.
 *
 * A service-level JAVA_HOME takes priority over the profile and global Java
 * settings, so this is enough on its own to change which JDK builds a service.
 */
export function JdkPicker({ serviceId, envVars, onApplied }: JdkPickerProps) {
  const [jdks, setJdks] = useState<InstalledJdk[] | null>(null);
  const [selected, setSelected] = useState("");
  const [applying, setApplying] = useState(false);
  const [error, setError] = useState("");

  const currentJavaHome = envVars?.JAVA_HOME?.value ?? "";

  useEffect(() => {
    let active = true;

    SystemApi.getInstalledJdks()
      .then((found) => {
        if (!active) return;
        setJdks(found);
        // Default to the JDK already pinned, else the oldest available, which
        // is the usual fix for an annotation processor that lags the JDK.
        setSelected(currentJavaHome || found[0]?.path || "");
      })
      .catch((err: Error) => active && setError(err.message));

    return () => {
      active = false;
    };
  }, [currentJavaHome]);

  const apply = async () => {
    if (!selected) return;

    setApplying(true);
    setError("");

    try {
      await SystemApi.setServiceJavaHome(serviceId, envVars ?? {}, selected);
      onApplied();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to set JAVA_HOME");
    } finally {
      setApplying(false);
    }
  };

  if (error) {
    return <p className="mt-2 text-xs text-red-700 dark:text-red-300">{error}</p>;
  }

  if (jdks === null) {
    return (
      <p className="mt-2 flex items-center gap-1.5 text-xs text-red-700 dark:text-red-300">
        <Loader className="h-3 w-3 animate-spin" />
        Looking for installed JDKs...
      </p>
    );
  }

  if (jdks.length === 0) {
    return (
      <p className="mt-2 text-xs text-red-700 dark:text-red-300">
        No other JDKs found on this machine. Install one, then set JAVA_HOME on
        this service.
      </p>
    );
  }

  return (
    <div className="mt-3">
      <label className="flex items-center gap-1.5 text-xs font-medium text-red-800 dark:text-red-200">
        <Coffee className="h-3 w-3" />
        Build this service with
      </label>

      <div className="mt-1.5 flex flex-wrap items-center gap-2">
        <select
          value={selected}
          onChange={(event) => setSelected(event.target.value)}
          className="min-w-0 flex-1 rounded border border-red-300 bg-white px-2 py-1 text-xs text-gray-900 dark:border-red-800 dark:bg-gray-900 dark:text-gray-100"
        >
          {jdks.map((jdk) => (
            <option key={jdk.path} value={jdk.path}>
              Java {jdk.majorVersion} ({jdk.source})
              {jdk.path === currentJavaHome ? " - current" : ""}
            </option>
          ))}
        </select>

        <Button
          onClick={apply}
          disabled={applying || !selected || selected === currentJavaHome}
          size="sm"
          className="h-7 bg-red-600 text-xs text-white hover:bg-red-700"
        >
          {applying ? "Applying..." : "Apply & restart"}
        </Button>
      </div>

      <p className="mt-1 truncate text-[11px] text-red-600 dark:text-red-400" title={selected}>
        {selected}
      </p>
    </div>
  );
}
