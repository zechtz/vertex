import { useState, useEffect } from 'react';
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
  Filter,
  ChevronUp,
  ChevronDown,
  Maximize2,
  Minimize2
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Service } from "@/types";
import { ButtonSpinner } from "@/components/ui/spinner";

interface LogsDrawerProps {
  selectedService: Service | null;
  searchTerm: string;
  copied: boolean;
  isCopyingLogs?: boolean;
  isClearingLogs?: boolean;
  onSearchChange: (term: string) => void;
  onClearSearch: () => void;
  onCopyLogs: (selectedLevels: string[]) => void;
  onClearLogs: () => void;
  onClose: () => void;
  onOpenAdvancedSearch?: () => void;
}

export function LogsDrawer({
  selectedService,
  searchTerm,
  copied,
  isCopyingLogs = false,
  isClearingLogs = false,
  onSearchChange,
  onClearSearch,
  onCopyLogs,
  onClearLogs,
  onClose,
  onOpenAdvancedSearch,
}: LogsDrawerProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [logLevels, setLogLevels] = useState<string[]>(["INFO", "WARN", "ERROR"]);
  const [isFullscreen, setIsFullscreen] = useState(false);

  // Auto-expand when service is selected
  useEffect(() => {
    if (selectedService) {
      setIsExpanded(true);
    }
  }, [selectedService]);

  const toggleLogLevel = (level: string) => {
    setLogLevels((prev) =>
      prev.includes(level) ? prev.filter((l) => l !== level) : [...prev, level]
    );
  };

  const filteredLogs = selectedService?.logs.filter(
    (log) =>
      logLevels.includes(log.level) &&
      log.message.toLowerCase().includes(searchTerm.toLowerCase())
  ) || [];

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

  if (!selectedService) {
    return null;
  }

  const drawerHeight = isFullscreen ? 'h-screen' : isExpanded ? 'h-96' : 'h-16';
  const zIndex = isFullscreen ? 'z-50' : 'z-40';

  return (
    <>
      {/* Fullscreen Backdrop */}
      {isFullscreen && (
        <div className="fixed inset-0 bg-black/50 z-40" onClick={() => setIsFullscreen(false)} />
      )}

      {/* Logs Drawer */}
      <div className={`
        fixed bottom-0 left-0 right-0 ${zIndex} transition-all duration-300 ease-in-out
        ${isFullscreen ? 'inset-0' : ''}
      `}>
        <div className={`
          bg-white border-t border-gray-200 shadow-2xl ${drawerHeight} flex flex-col
          ${isFullscreen ? 'rounded-none' : 'rounded-t-xl'}
        `}>
          {/* Header */}
          <div className="flex-shrink-0 px-6 py-4 border-b border-gray-200 bg-white">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                {/* Drag Handle */}
                <button
                  onClick={() => setIsExpanded(!isExpanded)}
                  className="p-1 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  {isExpanded ? (
                    <ChevronDown className="w-5 h-5 text-gray-400" />
                  ) : (
                    <ChevronUp className="w-5 h-5 text-gray-400" />
                  )}
                </button>

                <div className="p-2 bg-blue-100 rounded-lg">
                  <Terminal className="h-5 w-5 text-blue-600" />
                </div>
                
                <div>
                  <h3 className="text-lg font-semibold text-gray-900">
                    {selectedService.name} Logs
                  </h3>
                  <p className="text-sm text-gray-500">
                    {filteredLogs.length} of {selectedService.logs.length} entries
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-2">
                {/* Fullscreen Toggle */}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setIsFullscreen(!isFullscreen)}
                  className="hover:bg-gray-100"
                >
                  {isFullscreen ? (
                    <Minimize2 className="h-4 w-4" />
                  ) : (
                    <Maximize2 className="h-4 w-4" />
                  )}
                </Button>

                {/* Advanced Search */}
                {onOpenAdvancedSearch && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={onOpenAdvancedSearch}
                    className="hover:bg-purple-50 hover:text-purple-600 hover:border-purple-200"
                  >
                    <Filter className="h-4 w-4 mr-2" />
                    Advanced
                  </Button>
                )}

                {/* Copy Logs */}
                <Button
                  onClick={() => onCopyLogs(logLevels)}
                  disabled={filteredLogs.length === 0 || isCopyingLogs}
                  variant="outline"
                  size="sm"
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

                {/* Clear Logs */}
                <Button
                  onClick={onClearLogs}
                  disabled={selectedService.logs.length === 0 || isClearingLogs}
                  variant="outline"
                  size="sm"
                >
                  <ButtonSpinner isLoading={isClearingLogs} loadingText="Clearing...">
                    <Trash2 className="h-4 w-4 mr-2" />
                    Clear
                  </ButtonSpinner>
                </Button>

                {/* Close */}
                <Button 
                  variant="ghost" 
                  size="sm" 
                  onClick={onClose}
                  className="hover:bg-gray-100"
                >
                  <X className="h-5 w-5" />
                </Button>
              </div>
            </div>

            {/* Controls - Only show when expanded */}
            {isExpanded && (
              <div className="mt-4 space-y-4">
                {/* Search */}
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
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

                {/* Log Level Filters */}
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-gray-700">Levels:</span>
                  <Button
                    variant={logLevels.includes("INFO") ? "default" : "outline"}
                    size="sm"
                    onClick={() => toggleLogLevel("INFO")}
                    className="h-7 px-3 text-xs"
                  >
                    <Info className="h-3 w-3 mr-1" />
                    Info
                  </Button>
                  <Button
                    variant={logLevels.includes("WARN") ? "default" : "outline"}
                    size="sm"
                    onClick={() => toggleLogLevel("WARN")}
                    className="h-7 px-3 text-xs"
                  >
                    <ShieldAlert className="h-3 w-3 mr-1" />
                    Warn
                  </Button>
                  <Button
                    variant={logLevels.includes("ERROR") ? "default" : "outline"}
                    size="sm"
                    onClick={() => toggleLogLevel("ERROR")}
                    className="h-7 px-3 text-xs"
                  >
                    <Shield className="h-3 w-3 mr-1" />
                    Error
                  </Button>
                </div>

                {/* Search Results Info */}
                {searchTerm && (
                  <div className="text-sm text-gray-600">
                    {filteredLogs.length} of {selectedService.logs.length} logs match "{searchTerm}"
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Logs Content - Only show when expanded */}
          {isExpanded && (
            <div className="flex-1 overflow-hidden">
              <div className="h-full overflow-y-auto p-6">
                <div className="font-mono text-sm space-y-1">
                  {selectedService.logs.length === 0 ? (
                    <div className="text-center py-8 text-gray-500">
                      <Terminal className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                      <p className="text-lg font-medium">No logs available</p>
                      <p>Logs will appear here when the service generates output</p>
                    </div>
                  ) : filteredLogs.length === 0 ? (
                    <div className="text-center py-8 text-gray-500">
                      <Search className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                      <p className="text-lg font-medium">No logs match your criteria</p>
                      <p>Try adjusting your search term or log level filters</p>
                    </div>
                  ) : (
                    filteredLogs.map((log, index) => {
                      const highlightedLog = searchTerm
                        ? log.message.replace(
                            new RegExp(
                              `(${searchTerm.replace(/[.*+?^${}()|[\]]/g, "\\$&")})`,
                              "gi"
                            ),
                            '<mark class="bg-yellow-200 text-black px-1 rounded">$1</mark>'
                          )
                        : log.message;

                      return (
                        <div 
                          key={index} 
                          className={`text-xs leading-relaxed p-2 rounded hover:bg-gray-50 ${getLogLevelClass(log.level)}`}
                        >
                          <div className="flex items-start gap-3">
                            <span className="text-gray-400 w-20 flex-shrink-0">
                              {new Date(log.timestamp).toLocaleTimeString()}
                            </span>
                            <span className={`font-medium w-12 flex-shrink-0 ${getLogLevelClass(log.level)}`}>
                              [{log.level}]
                            </span>
                            <span 
                              className="flex-1 break-all"
                              dangerouslySetInnerHTML={{ __html: highlightedLog }} 
                            />
                          </div>
                        </div>
                      );
                    })
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </>
  );
}

export default LogsDrawer;