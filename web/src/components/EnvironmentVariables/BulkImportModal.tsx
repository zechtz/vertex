import { useState, useCallback } from "react";
import { Upload, FileText, AlertCircle, CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import Modal from "@/components/ui/Modal";
import { EnvVariableParser, EnvFormat, ParseResult } from "@/services/envVariableParser";

interface BulkImportModalProps {
  isOpen: boolean;
  onClose: () => void;
  onImport: (variables: Record<string, string>) => void;
}

export function BulkImportModal({ isOpen, onClose, onImport }: BulkImportModalProps) {
  const [selectedFormat, setSelectedFormat] = useState<EnvFormat>('auto');
  const [inputText, setInputText] = useState('');
  const [parseResult, setParseResult] = useState<ParseResult | null>(null);
  const [showPreview, setShowPreview] = useState(false);

  const handleFormatChange = (format: EnvFormat) => {
    setSelectedFormat(format);
    if (inputText.trim()) {
      parseAndPreview(inputText, format);
    }
  };

  const parseAndPreview = useCallback((text: string, format: EnvFormat) => {
    if (!text.trim()) {
      setParseResult(null);
      setShowPreview(false);
      return;
    }

    const result = EnvVariableParser.parse(text, format);
    setParseResult(result);
    setShowPreview(true);
  }, []);

  const handleInputChange = (text: string) => {
    setInputText(text);
    parseAndPreview(text, selectedFormat);
  };

  const handleImport = () => {
    if (parseResult && Object.keys(parseResult.variables).length > 0) {
      onImport(parseResult.variables);
      handleClose();
    }
  };

  const handleClose = () => {
    setInputText('');
    setParseResult(null);
    setShowPreview(false);
    setSelectedFormat('auto');
    onClose();
  };

  const formatOptions: { value: EnvFormat; label: string }[] = [
    { value: 'auto', label: 'Auto-detect' },
    { value: 'dotenv', label: '.env / Docker Compose' },
    { value: 'json', label: 'JSON' },
    { value: 'yaml', label: 'YAML' },
    { value: 'shell', label: 'Shell Export' },
    { value: 'fish', label: 'Fish Shell' },
    { value: 'batch', label: 'Windows Batch' },
    { value: 'properties', label: 'Java Properties' },
  ];

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Bulk Import Environment Variables"
      size="2xl"
      className="max-h-[90vh]"
    >
      <div className="flex flex-col h-full min-h-0">
        {/* Format Selection */}
        <div className="p-6 border-b border-gray-200 dark:border-gray-600">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Format
              </label>
              <select
                value={selectedFormat}
                onChange={(e) => handleFormatChange(e.target.value as EnvFormat)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-gray-100"
              >
                {formatOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Format Example */}
            {selectedFormat !== 'auto' && (
              <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                <div className="flex items-center gap-2 mb-2">
                  <FileText className="h-4 w-4 text-gray-500" />
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                    {EnvVariableParser.getFormatDisplayName(selectedFormat)} Example:
                  </span>
                </div>
                <pre className="text-xs text-gray-600 dark:text-gray-400 bg-white dark:bg-gray-900 p-3 rounded border overflow-x-auto">
                  {EnvVariableParser.getFormatExample(selectedFormat)}
                </pre>
              </div>
            )}
          </div>
        </div>

        {/* Content Area */}
        <div className="flex-1 flex min-h-0">
          {/* Input Section */}
          <div className="flex-1 flex flex-col border-r border-gray-200 dark:border-gray-600">
            <div className="p-4 border-b border-gray-200 dark:border-gray-600">
              <div className="flex items-center gap-2">
                <Upload className="h-4 w-4 text-gray-500" />
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  Paste Environment Variables
                </span>
              </div>
            </div>
            <div className="flex-1 p-4">
              <textarea
                value={inputText}
                onChange={(e) => handleInputChange(e.target.value)}
                placeholder={`Paste your environment variables here...\n\n${selectedFormat !== 'auto' ? EnvVariableParser.getFormatExample(selectedFormat) : 'The format will be auto-detected, or select a specific format above.'}`}
                className="w-full h-full resize-none border border-gray-300 dark:border-gray-600 rounded-md p-3 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-gray-100"
              />
            </div>
          </div>

          {/* Preview Section */}
          <div className="flex-1 flex flex-col">
            <div className="p-4 border-b border-gray-200 dark:border-gray-600">
              <div className="flex items-center gap-2">
                {parseResult && parseResult.errors.length === 0 && Object.keys(parseResult.variables).length > 0 ? (
                  <CheckCircle2 className="h-4 w-4 text-green-500" />
                ) : parseResult && parseResult.errors.length > 0 ? (
                  <AlertCircle className="h-4 w-4 text-red-500" />
                ) : (
                  <FileText className="h-4 w-4 text-gray-500" />
                )}
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  Preview
                  {parseResult && (
                    <span className="ml-2 text-xs text-gray-500">
                      ({Object.keys(parseResult.variables).length} variables
                      {parseResult.format !== 'auto' && ` • ${EnvVariableParser.getFormatDisplayName(parseResult.format)}`})
                    </span>
                  )}
                </span>
              </div>
            </div>

            <div className="flex-1 p-4 overflow-y-auto">
              {!showPreview ? (
                <div className="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">
                  <div className="text-center">
                    <FileText className="h-12 w-12 mx-auto mb-2 opacity-50" />
                    <p className="text-sm">Paste variables to see preview</p>
                  </div>
                </div>
              ) : parseResult ? (
                <div className="space-y-4">
                  {/* Errors */}
                  {parseResult.errors.length > 0 && (
                    <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
                      <div className="flex items-center gap-2 mb-2">
                        <AlertCircle className="h-4 w-4 text-red-500" />
                        <span className="text-sm font-medium text-red-800 dark:text-red-200">
                          Parsing Errors ({parseResult.errors.length})
                        </span>
                      </div>
                      <div className="space-y-1">
                        {parseResult.errors.map((error, index) => (
                          <div key={index} className="text-xs text-red-700 dark:text-red-300 font-mono">
                            {error}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Variables */}
                  {Object.keys(parseResult.variables).length > 0 ? (
                    <div className="space-y-2">
                      <div className="text-sm font-medium text-gray-700 dark:text-gray-300">
                        Variables to Import ({Object.keys(parseResult.variables).length})
                      </div>
                      <div className="space-y-1 max-h-96 overflow-y-auto">
                        {Object.entries(parseResult.variables).map(([key, value]) => (
                          <div
                            key={key}
                            className="flex items-start gap-2 p-2 bg-gray-50 dark:bg-gray-700 rounded border text-xs"
                          >
                            <span className="font-mono font-medium text-blue-600 dark:text-blue-400 break-all">
                              {key}
                            </span>
                            <span className="text-gray-500">=</span>
                            <span className="font-mono text-gray-600 dark:text-gray-300 flex-1 break-all">
                              {value}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  ) : (
                    <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
                      <div className="text-center">
                        <AlertCircle className="h-8 w-8 mx-auto mb-2 opacity-50" />
                        <p className="text-sm">No valid variables found</p>
                      </div>
                    </div>
                  )}
                </div>
              ) : null}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 flex items-center justify-between flex-shrink-0">
          <div className="text-sm text-gray-600 dark:text-gray-400">
            {parseResult && Object.keys(parseResult.variables).length > 0 && (
              <span>
                Ready to import {Object.keys(parseResult.variables).length} variables
                {parseResult.errors.length > 0 && (
                  <span className="text-red-600 dark:text-red-400 ml-2">
                    ({parseResult.errors.length} errors)
                  </span>
                )}
              </span>
            )}
          </div>
          <div className="flex items-center gap-3">
            <Button type="button" variant="outline" onClick={handleClose}>
              Cancel
            </Button>
            <Button
              onClick={handleImport}
              disabled={!parseResult || Object.keys(parseResult.variables).length === 0}
              className="flex items-center gap-2"
            >
              <Upload className="h-4 w-4" />
              Import Variables ({parseResult ? Object.keys(parseResult.variables).length : 0})
            </Button>
          </div>
        </div>
      </div>
    </Modal>
  );
}