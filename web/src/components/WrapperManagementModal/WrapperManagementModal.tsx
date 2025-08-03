import React, { useState, useEffect } from 'react';
import { X, Wrench, Package, AlertTriangle, CheckCircle, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface WrapperValidation {
  serviceId: string;
  serviceName: string;
  buildSystem: string;
  isValid: boolean;
  hasWrapper: boolean;
  wrapperFiles: string[];
  error?: string;
}

interface WrapperManagementModalProps {
  serviceId: string;
  serviceName: string;
  isOpen: boolean;
  onClose: () => void;
  onValidateWrapper: () => Promise<any>;
  onGenerateWrapper: () => Promise<void>;
  onRepairWrapper: () => Promise<void>;
  isValidating?: boolean;
  isGenerating?: boolean;
  isRepairing?: boolean;
}

const WrapperManagementModal: React.FC<WrapperManagementModalProps> = ({
  serviceId,
  serviceName,
  isOpen,
  onClose,
  onValidateWrapper,
  onGenerateWrapper,
  onRepairWrapper,
  isValidating = false,
  isGenerating = false,
  isRepairing = false
}) => {
  const [validation, setValidation] = useState<WrapperValidation | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen && serviceId) {
      handleValidateWrapper();
    }
  }, [isOpen, serviceId]);

  const handleValidateWrapper = async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await onValidateWrapper();
      if (result.success && result.data) {
        setValidation(result.data);
      } else {
        setError(result.error || 'Failed to validate wrapper');
      }
    } catch (err: any) {
      setError(err.message || 'Failed to validate wrapper');
      console.error('Failed to validate wrapper:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleGenerateWrapper = async () => {
    try {
      await onGenerateWrapper();
      // Refresh validation after generation
      setTimeout(() => {
        handleValidateWrapper();
      }, 1000);
    } catch (err: any) {
      setError(err.message || 'Failed to generate wrapper');
    }
  };

  const handleRepairWrapper = async () => {
    try {
      await onRepairWrapper();
      // Refresh validation after repair
      setTimeout(() => {
        handleValidateWrapper();
      }, 1000);
    } catch (err: any) {
      setError(err.message || 'Failed to repair wrapper');
    }
  };

  const renderWrapperStatus = () => {
    if (!validation) return null;

    const statusIcon = validation.isValid ? (
      <CheckCircle className="w-5 h-5 text-green-500" />
    ) : (
      <AlertTriangle className="w-5 h-5 text-yellow-500" />
    );

    const statusText = validation.isValid ? 'Valid' : 'Invalid/Missing';
    const statusColor = validation.isValid ? 'text-green-600 dark:text-green-400' : 'text-yellow-600 dark:text-yellow-400';

    return (
      <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 space-y-3">
        <div className="flex items-center justify-between">
          <h4 className="font-medium text-gray-900 dark:text-gray-100">Wrapper Status</h4>
          <div className="flex items-center gap-2">
            {statusIcon}
            <span className={`font-medium ${statusColor}`}>{statusText}</span>
          </div>
        </div>
        
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-gray-600 dark:text-gray-400">Build System:</span>
            <span className="font-medium text-gray-900 dark:text-gray-100 capitalize">{validation.buildSystem}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-gray-600 dark:text-gray-400">Wrapper Files:</span>
            <span className="font-medium text-gray-900 dark:text-gray-100">
              {validation.hasWrapper ? validation.wrapperFiles.length : 0} files
            </span>
          </div>
        </div>

        {validation.wrapperFiles.length > 0 && (
          <div className="mt-3">
            <div className="text-sm text-gray-600 dark:text-gray-400 mb-2">Files:</div>
            <div className="space-y-1">
              {validation.wrapperFiles.map((file, index) => (
                <div key={index} className="text-xs font-mono bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded">
                  {file}
                </div>
              ))}
            </div>
          </div>
        )}

        {validation.error && (
          <div className="mt-3 text-sm text-red-600 dark:text-red-400">
            Error: {validation.error}
          </div>
        )}
      </div>
    );
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-3">
            <Package className="w-6 h-6 text-blue-500" />
            <div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                Wrapper Management
              </h2>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Manage build wrapper for {serviceName}
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-blue-500" />
              <span className="ml-2 text-gray-600 dark:text-gray-400">Validating wrapper...</span>
            </div>
          ) : error ? (
            <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
              <div className="flex items-center gap-2">
                <AlertTriangle className="w-5 h-5 text-red-500" />
                <span className="font-medium text-red-700 dark:text-red-400">Error</span>
              </div>
              <p className="mt-2 text-sm text-red-600 dark:text-red-300">{error}</p>
            </div>
          ) : (
            <>
              {renderWrapperStatus()}

              <div className="space-y-4">
                <h4 className="font-medium text-gray-900 dark:text-gray-100">Available Actions</h4>
                
                <div className="grid gap-3">
                  <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                    <div>
                      <h5 className="font-medium text-gray-900 dark:text-gray-100">Validate Wrapper</h5>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        Check if wrapper files exist and are functional
                      </p>
                    </div>
                    <Button
                      onClick={handleValidateWrapper}
                      disabled={isValidating}
                      variant="outline"
                      size="sm"
                    >
                      {isValidating ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                      ) : (
                        <CheckCircle className="w-4 h-4" />
                      )}
                      <span className="ml-2">Validate</span>
                    </Button>
                  </div>

                  <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                    <div>
                      <h5 className="font-medium text-gray-900 dark:text-gray-100">Generate Wrapper</h5>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        Create new wrapper files for {validation?.buildSystem || 'the detected'} build system
                      </p>
                    </div>
                    <Button
                      onClick={handleGenerateWrapper}
                      disabled={isGenerating}
                      variant="default"
                      size="sm"
                    >
                      {isGenerating ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                      ) : (
                        <Package className="w-4 h-4" />
                      )}
                      <span className="ml-2">Generate</span>
                    </Button>
                  </div>

                  <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                    <div>
                      <h5 className="font-medium text-gray-900 dark:text-gray-100">Repair Wrapper</h5>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        Fix corrupted or missing wrapper files
                      </p>
                    </div>
                    <Button
                      onClick={handleRepairWrapper}
                      disabled={isRepairing}
                      variant="outline"
                      size="sm"
                    >
                      {isRepairing ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                      ) : (
                        <Wrench className="w-4 h-4" />
                      )}
                      <span className="ml-2">Repair</span>
                    </Button>
                  </div>
                </div>
              </div>
            </>
          )}
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-3 p-6 border-t border-gray-200 dark:border-gray-700">
          <Button onClick={onClose} variant="outline">
            Close
          </Button>
        </div>
      </div>
    </div>
  );
};

export default WrapperManagementModal;