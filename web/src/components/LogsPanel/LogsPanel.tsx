import { 
  Terminal, 
  Copy, 
  Check, 
  Trash2, 
  Search, 
  X, 
  Shield, 
  ShieldAlert, 
  Info,
  Filter
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Service } from "@/types";
import { useState } from "react";
import { LogsPanelSkeleton } from "@/components/ui/skeleton";
import { ErrorBoundarySection } from "@/components/ui/error-boundary";
import { ButtonSpinner } from "@/components/ui/spinner";

interface LogsPanelProps {
  selectedService: Service | null;
  searchTerm: string;
  copied: boolean;
  isLoading?: boolean;
  isCopyingLogs?: boolean;
  isClearingLogs?: boolean;
  onSearchChange: (term: string) => void;
  onClearSearch: () => void;
  onCopyLogs: (selectedLevels: string[]) => void;
  onClearLogs: () => void;
  onClose: () => void;
  onOpenAdvancedSearch?: () => void;
}

export function LogsPanel({
  selectedService,
  searchTerm,
  copied,
  isLoading = false,
  isCopyingLogs = false,
  isClearingLogs = false,
  onSearchChange,
  onClearSearch,
  onCopyLogs,
  onClearLogs,
  onClose,
  onOpenAdvancedSearch,
}: LogsPanelProps) {
  const [logLevels, setLogLevels] = useState<string[]>(["INFO", "WARN", "ERROR"]);
  
  if (isLoading) {
    return <LogsPanelSkeleton />;
  }
  if (!selectedService) {
    return (
      <Card className="h-[700px] border-0 shadow-xl bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
        <CardContent className="h-full flex items-center justify-center">
          <div className="text-center">
            <div className="p-4 bg-blue-100 dark:bg-blue-900/30 rounded-full w-20 h-20 flex items-center justify-center mx-auto mb-4">
              <Terminal className="h-10 w-10 text-blue-600 dark:text-blue-400" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">
              Service Logs
            </h3>
            <p className="text-muted-foreground">
              Select a service to view its real-time logs and output
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  const toggleLogLevel = (level: string) => {
    setLogLevels((prev) =>
      prev.includes(level) ? prev.filter((l) => l !== level) : [...prev, level]
    );
  };

  const filteredLogs = selectedService.logs.filter(
    (log) =>
      logLevels.includes(log.level) &&
      log.message.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const getLogLevelClass = (level: string) => {
    switch (level) {
      case "ERROR":
        return "text-red-500";
      case "WARN":
        return "text-yellow-500";
      default:
        return "text-gray-500";
    }
  };

  return (
    <ErrorBoundarySection title="Logs Panel Error" description="Failed to load the logs panel.">
      <Card className="h-[700px] flex flex-col border-0 shadow-xl bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
        <CardHeader className="pb-4 border-b bg-white/50 dark:bg-slate-900/50 backdrop-blur-sm">
          <CardTitle className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                <Terminal className="h-5 w-5 text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <span className="text-lg font-semibold">
                  {selectedService.name} Logs
                </span>
                <p className="text-sm text-muted-foreground font-normal">
                  Real-time service output
                </p>
              </div>
            </div>
          <div className="flex gap-2">
            {onOpenAdvancedSearch && (
              <Button
                variant="outline"
                size="sm"
                onClick={onOpenAdvancedSearch}
                className="hover:bg-purple-50 hover:text-purple-600 hover:border-purple-200"
              >
                <Filter className="h-4 w-4 mr-2" />
                Advanced Search
              </Button>
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={() => onCopyLogs(logLevels)}
              disabled={selectedService.logs.length === 0 || isCopyingLogs}
            >
              <ButtonSpinner isLoading={isCopyingLogs} loadingText="Copying...">
                {copied ? (
                  <Check className="h-4 w-4 mr-2" />
                ) : (
                  <Copy className="h-4 w-4 mr-2" />
                )}
                {copied ? "Copied!" : "Copy"}
              </ButtonSpinner>
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={onClearLogs}
              disabled={selectedService.logs.length === 0 || isClearingLogs}
            >
              <ButtonSpinner isLoading={isClearingLogs} loadingText="Clearing...">
                <Trash2 className="h-4 w-4 mr-2" />
                Clear
              </ButtonSpinner>
            </Button>
            <Button variant="outline" size="sm" onClick={onClose}>
              Close
            </Button>
          </div>
        </CardTitle>
        <div className="flex justify-between items-center mt-4">
          <div className="relative flex-grow">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search logs..."
              value={searchTerm}
              onChange={(e) => onSearchChange(e.target.value)}
              className="pl-10 pr-10"
            />
            {searchTerm && (
              <Button
                variant="ghost"
                size="sm"
                onClick={onClearSearch}
                className="absolute right-1 top-1/2 transform -translate-y-1/2 h-8 w-8 p-0"
              >
                <X className="h-4 w-4" />
              </Button>
            )}
          </div>
          <div className="flex gap-2 ml-4">
            <Button
              variant={logLevels.includes("INFO") ? "default" : "outline"}
              size="sm"
              onClick={() => toggleLogLevel("INFO")}
            >
              <Info className="h-4 w-4 mr-2" />
              Info
            </Button>
            <Button
              variant={logLevels.includes("WARN") ? "default" : "outline"}
              size="sm"
              onClick={() => toggleLogLevel("WARN")}
            >
              <ShieldAlert className="h-4 w-4 mr-2" />
              Warn
            </Button>
            <Button
              variant={logLevels.includes("ERROR") ? "default" : "outline"}
              size="sm"
              onClick={() => toggleLogLevel("ERROR")}
            >
              <Shield className="h-4 w-4 mr-2" />
              Error
            </Button>
          </div>
        </div>
        {searchTerm && (
          <div className="text-sm text-muted-foreground mt-2">
            {filteredLogs.length} of {selectedService.logs.length} logs match "
            {searchTerm}"
          </div>
        )}
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto">
        <div className="font-mono text-sm space-y-1">
          {selectedService.logs.length === 0 ? (
            <p className="text-muted-foreground">No logs available</p>
          ) : filteredLogs.length === 0 ? (
            <p className="text-muted-foreground">No logs match your search or filter</p>
          ) : (
            filteredLogs.map((log, index) => {
              const highlightedLog = searchTerm
                ? log.message.replace(
                    new RegExp(
                      `(${searchTerm.replace(/[.*+?^${}()|[\]]/g, "\\$&")})`,
                      "gi"
                    ),
                    '<mark class="bg-yellow-200 text-black">$1</mark>'
                  )
                : log.message;

              return (
                <div key={index} className={`text-xs leading-relaxed ${getLogLevelClass(log.level)}`}>
                  <span className="mr-2">{new Date(log.timestamp).toLocaleTimeString()}</span>
                  <span dangerouslySetInnerHTML={{ __html: highlightedLog }} />
                </div>
              );
            })
          )}
        </div>
      </CardContent>
      </Card>
    </ErrorBoundarySection>
  );
}