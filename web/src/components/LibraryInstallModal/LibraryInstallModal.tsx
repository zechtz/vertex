import React, { useState, useEffect } from 'react';
import { X } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface LibraryInstallation {
  file: string;
  group_id: string;
  artifact_id: string;
  version: string;
  packaging: string;
  command: string;
}

interface EnvironmentLibraries {
  environment: string;
  jobName: string;
  libraries: LibraryInstallation[];
  branches: string[];
}

interface LibraryPreview {
  hasLibraries: boolean;
  serviceName: string;
  serviceId: string;
  environments: EnvironmentLibraries[];
  totalLibraries: number;
  gitlabCIExists: boolean;
  errorMessage?: string;
}

// Note: InstallProgress and EnvironmentProgress interfaces will be used 
// for future WebSocket implementation

interface LibraryInstallModalProps {
  serviceId: string;
  serviceName: string;
  isOpen: boolean;
  onClose: () => void;
}

const LibraryInstallModal: React.FC<LibraryInstallModalProps> = ({
  serviceId,
  serviceName,
  isOpen,
  onClose
}) => {
  const [preview, setPreview] = useState<LibraryPreview | null>(null);
  const [selectedEnvironments, setSelectedEnvironments] = useState<string[]>([]);
  const [showConfirmation, setShowConfirmation] = useState(false);
  const [installing, setInstalling] = useState(false);
  const [installResult, setInstallResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen && serviceId) {
      fetchLibraryPreview();
    }
  }, [isOpen, serviceId]);

  const fetchLibraryPreview = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(`/api/services/${serviceId}/libraries/preview`);
      const data = await response.json();
      setPreview(data);
      
      if (!response.ok) {
        throw new Error(data.message || 'Failed to preview libraries');
      }
    } catch (err: any) {
      setError(err.message);
      console.error('Failed to fetch library preview:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleEnvironmentChange = (environment: string, checked: boolean) => {
    if (checked) {
      setSelectedEnvironments(prev => [...prev, environment]);
    } else {
      setSelectedEnvironments(prev => prev.filter(env => env !== environment));
    }
  };

  const handleConfirm = () => {
    setShowConfirmation(true);
  };

  const handleInstall = async () => {
    setInstalling(true);
    setError(null);
    
    try {
      const response = await fetch(`/api/services/${serviceId}/libraries/install`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          environments: selectedEnvironments,
          confirmed: true,
        }),
      });

      const result = await response.json();
      
      if (!response.ok) {
        throw new Error(result.message || 'Failed to install libraries');
      }

      setInstallResult(result);
      
      // Auto-close after 3 seconds if successful
      setTimeout(() => {
        onClose();
        resetState();
      }, 3000);
      
    } catch (err: any) {
      setError(err.message);
      console.error('Failed to install libraries:', err);
    } finally {
      setInstalling(false);
    }
  };

  const resetState = () => {
    setPreview(null);
    setSelectedEnvironments([]);
    setShowConfirmation(false);
    setInstalling(false);
    setInstallResult(null);
    setError(null);
    setLoading(false);
  };

  const handleClose = () => {
    resetState();
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-screen items-center justify-center p-4">
        {/* Backdrop */}
        <div 
          className="fixed inset-0 bg-black/50 z-40" 
          onClick={handleClose}
        />
        
        {/* Modal */}
        <div className="relative w-full max-w-4xl max-h-[90vh] overflow-y-auto z-50">
          <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl">
            {/* Header */}
            <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                Install Libraries - {serviceName}
              </h2>
              <Button
                variant="ghost"
                size="sm"
                onClick={handleClose}
                className="hover:bg-gray-100 dark:hover:bg-gray-700"
              >
                <X className="h-5 w-5" />
              </Button>
            </div>

            {/* Content */}
            <div className="p-6">
              {loading && (
                <div className="flex items-center justify-center py-8">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                  <span className="ml-3 text-gray-600 dark:text-gray-300">Analyzing libraries...</span>
                </div>
              )}

              {error && (
                <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4 mb-4">
                  <div className="flex">
                    <div className="flex-shrink-0">
                      <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                      </svg>
                    </div>
                    <div className="ml-3">
                      <h3 className="text-sm font-medium text-red-800 dark:text-red-200">Error</h3>
                      <p className="text-sm text-red-700 dark:text-red-300 mt-1">{error}</p>
                    </div>
                  </div>
                </div>
              )}

              {installResult && (
                <div className="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-md p-4 mb-4">
                  <div className="flex">
                    <div className="flex-shrink-0">
                      <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                      </svg>
                    </div>
                    <div className="ml-3">
                      <h3 className="text-sm font-medium text-green-800 dark:text-green-200">Success!</h3>
                      <p className="text-sm text-green-700 dark:text-green-300 mt-1">
                        {installResult.message}
                      </p>
                      <p className="text-sm text-green-600 dark:text-green-400 mt-1">
                        Installed {installResult.librariesInstalled} libraries for {installResult.environments?.join(', ')} environments.
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {preview && !preview.hasLibraries && (
                <div className="text-center py-8">
                  <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-gray-100 dark:bg-gray-700">
                    <svg className="h-6 w-6 text-gray-600 dark:text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                  </div>
                  <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-gray-100">No Custom Libraries</h3>
                  <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    This service has no custom libraries to install.
                  </p>
                  {preview.errorMessage && (
                    <p className="mt-2 text-xs text-gray-400 dark:text-gray-500">
                      {preview.errorMessage}
                    </p>
                  )}
                </div>
              )}

              {preview && preview.hasLibraries && !showConfirmation && !installing && (
                <LibrarySelectionView
                  preview={preview}
                  selectedEnvironments={selectedEnvironments}
                  onEnvironmentChange={handleEnvironmentChange}
                  onConfirm={handleConfirm}
                  onCancel={handleClose}
                />
              )}

              {showConfirmation && !installing && (
                <LibraryConfirmationView
                  preview={preview!}
                  selectedEnvironments={selectedEnvironments}
                  onBack={() => setShowConfirmation(false)}
                  onInstall={handleInstall}
                />
              )}

              {installing && (
                <div className="text-center py-8">
                  <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
                  <h3 className="mt-4 text-lg font-medium text-gray-900 dark:text-gray-100">Installing Libraries...</h3>
                  <p className="mt-2 text-sm text-gray-600 dark:text-gray-300">
                    Installing libraries for {selectedEnvironments.join(', ')} environments
                  </p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

// Library Selection View Component
const LibrarySelectionView: React.FC<{
  preview: LibraryPreview;
  selectedEnvironments: string[];
  onEnvironmentChange: (env: string, checked: boolean) => void;
  onConfirm: () => void;
  onCancel: () => void;
}> = ({ preview, selectedEnvironments, onEnvironmentChange, onConfirm, onCancel }) => {
  return (
    <div>
      <div className="mb-6">
        <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
          Found {preview.totalLibraries} custom libraries across {preview.environments.length} environments
        </h3>
        <p className="text-sm text-gray-600 dark:text-gray-300 mb-4">
          Select which environments to install libraries for:
        </p>
        
        <div className="space-y-4">
          {preview.environments.map((env) => (
            <div key={env.environment} className="border border-gray-200 dark:border-gray-600 rounded-lg p-4">
              <div className="flex items-start">
                <div className="flex items-center h-5">
                  <input
                    id={env.environment}
                    type="checkbox"
                    checked={selectedEnvironments.includes(env.environment)}
                    onChange={(e) => onEnvironmentChange(env.environment, e.target.checked)}
                    className="focus:ring-blue-500 h-4 w-4 text-blue-600 border-gray-300 rounded"
                  />
                </div>
                <div className="ml-3 flex-1">
                  <label htmlFor={env.environment} className="font-medium text-gray-900 dark:text-gray-100 cursor-pointer">
                    {env.environment.toUpperCase()} Environment
                  </label>
                  <p className="text-sm text-gray-500 dark:text-gray-400 mb-2">
                    Job: <code className="bg-gray-100 dark:bg-gray-700 px-1 py-0.5 rounded text-xs">{env.jobName}</code>
                    {env.branches.length > 0 && (
                      <span className="ml-2">
                        | Branches: {env.branches.join(', ')}
                      </span>
                    )}
                    <span className="ml-2 text-blue-600">
                      ({env.libraries.length} libraries)
                    </span>
                  </p>
                  
                  <div className="space-y-1">
                    {env.libraries.map((lib, i) => (
                      <div key={i} className="text-xs font-mono bg-gray-50 dark:bg-gray-700 p-2 rounded border-l-2 border-blue-200 dark:border-blue-400">
                        <span className="text-blue-600 dark:text-blue-400">{lib.group_id}</span>:
                        <span className="text-green-600 dark:text-green-400">{lib.artifact_id}</span>:
                        <span className="text-purple-600 dark:text-purple-400">{lib.version}</span>
                        <div className="text-gray-500 dark:text-gray-400 text-[10px] mt-1 truncate">
                          {lib.file}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
      
      <div className="flex justify-end space-x-3 pt-4 border-t border-gray-200 dark:border-gray-600">
        <button
          onClick={onCancel}
          className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          Cancel
        </button>
        <button 
          onClick={onConfirm}
          disabled={selectedEnvironments.length === 0}
          className="px-4 py-2 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Continue ({selectedEnvironments.length} environment{selectedEnvironments.length !== 1 ? 's' : ''} selected)
        </button>
      </div>
    </div>
  );
};

// Library Confirmation View Component
const LibraryConfirmationView: React.FC<{
  preview: LibraryPreview;
  selectedEnvironments: string[];
  onBack: () => void;
  onInstall: () => void;
}> = ({ preview, selectedEnvironments, onBack, onInstall }) => {
  const selectedEnvData = preview.environments.filter(env => 
    selectedEnvironments.includes(env.environment)
  );
  
  const totalLibraries = selectedEnvData.reduce((sum, env) => sum + env.libraries.length, 0);

  return (
    <div>
      <div className="mb-6">
        <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
          Confirm Library Installation
        </h3>
        
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 mb-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h4 className="text-sm font-medium text-blue-800 dark:text-blue-200">Installation Summary</h4>
              <div className="text-sm text-blue-700 dark:text-blue-300 mt-1">
                <p>Service: <strong>{preview.serviceName}</strong></p>
                <p>Environments: <strong>{selectedEnvironments.join(', ')}</strong></p>
                <p>Total Libraries: <strong>{totalLibraries}</strong></p>
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-3">
          {selectedEnvData.map((env) => (
            <div key={env.environment} className="border border-gray-200 dark:border-gray-600 rounded-lg p-3">
              <h4 className="font-medium text-gray-900 dark:text-gray-100 mb-2">
                {env.environment.toUpperCase()} Environment ({env.libraries.length} libraries)
              </h4>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                {env.libraries.map((lib, i) => (
                  <div key={i} className="text-xs font-mono bg-gray-50 dark:bg-gray-700 p-2 rounded">
                    <span className="text-blue-600 dark:text-blue-400">{lib.group_id}</span>:
                    <span className="text-green-600 dark:text-green-400">{lib.artifact_id}</span>:
                    <span className="text-purple-600 dark:text-purple-400">{lib.version}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
      
      <div className="flex justify-end space-x-3 pt-4 border-t border-gray-200 dark:border-gray-600">
        <button
          onClick={onBack}
          className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          Back
        </button>
        <button 
          onClick={onInstall}
          className="px-4 py-2 bg-green-600 text-white rounded-md text-sm font-medium hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
        >
          Install Libraries
        </button>
      </div>
    </div>
  );
};

export default LibraryInstallModal;