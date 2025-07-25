import { useState, useEffect } from "react";
import { X, File, Edit, Save, RefreshCw, Download } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { useToast, toast } from "@/components/ui/toast";
import { useConfirm } from "@/components/ui/confirm-dialog";
import { ButtonSpinner } from "@/components/ui/spinner";
import { ErrorBoundarySection } from "@/components/ui/error-boundary";

interface ServiceFile {
  name: string;
  path: string;
  content: string;
  type: 'properties' | 'yml' | 'yaml';
  lastModified: string;
}

interface ServiceFilesModalProps {
  serviceName: string;
  serviceDir: string;
  isOpen: boolean;
  onClose: () => void;
}

export function ServiceFilesModal({
  serviceName,
  serviceDir,
  isOpen,
  onClose,
}: ServiceFilesModalProps) {
  const [files, setFiles] = useState<ServiceFile[]>([]);
  const [selectedFile, setSelectedFile] = useState<ServiceFile | null>(null);
  const [editedContent, setEditedContent] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  
  const { addToast } = useToast();
  const { showConfirm } = useConfirm();

  useEffect(() => {
    if (isOpen && serviceName) {
      fetchServiceFiles();
    }
  }, [isOpen, serviceName]);

  const fetchServiceFiles = async () => {
    try {
      setIsLoading(true);
      const response = await fetch(`/api/services/${serviceName}/files`);
      if (!response.ok) {
        throw new Error(`Failed to fetch service files: ${response.status} ${response.statusText}`);
      }
      const data = await response.json();
      setFiles(data.files || []);
      if (data.files && data.files.length > 0) {
        setSelectedFile(data.files[0]);
        setEditedContent(data.files[0].content);
      }
    } catch (error) {
      console.error("Failed to fetch service files:", error);
      addToast(toast.error(
        "Failed to load service files",
        error instanceof Error ? error.message : "An unexpected error occurred"
      ));
    } finally {
      setIsLoading(false);
    }
  };

  const handleFileSelect = async (file: ServiceFile) => {
    if (isEditing) {
      const confirmed = await showConfirm({
        title: "Unsaved Changes",
        description: "You have unsaved changes. Do you want to discard them?",
        confirmText: "Discard Changes",
        variant: "warning"
      });
      if (!confirmed) return;
    }
    
    setSelectedFile(file);
    setEditedContent(file.content);
    setIsEditing(false);
  };

  const handleEdit = () => {
    setIsEditing(true);
  };

  const handleSave = async () => {
    if (!selectedFile) return;

    try {
      setIsSaving(true);
      const response = await fetch(`/api/services/${serviceName}/files/${encodeURIComponent(selectedFile.name)}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          content: editedContent,
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to save file: ${response.status} ${response.statusText}`);
      }

      // Update the file in state
      const updatedFiles = files.map(f => 
        f.name === selectedFile.name 
          ? { ...f, content: editedContent, lastModified: new Date().toISOString() }
          : f
      );
      setFiles(updatedFiles);
      setSelectedFile({ ...selectedFile, content: editedContent });
      setIsEditing(false);
      addToast(toast.success(
        "File saved",
        `Successfully saved ${selectedFile.name}`
      ));
    } catch (error) {
      console.error("Failed to save file:", error);
      addToast(toast.error(
        "Failed to save file",
        error instanceof Error ? error.message : "An unexpected error occurred"
      ));
    } finally {
      setIsSaving(false);
    }
  };

  const handleCancel = () => {
    if (selectedFile) {
      setEditedContent(selectedFile.content);
    }
    setIsEditing(false);
  };

  const handleDownload = () => {
    if (!selectedFile) return;

    // Create a blob with the file content
    const blob = new Blob([selectedFile.content], { type: 'text/plain' });
    const url = window.URL.createObjectURL(blob);
    
    // Create a temporary download link
    const link = document.createElement('a');
    link.href = url;
    link.download = selectedFile.name;
    document.body.appendChild(link);
    link.click();
    
    // Cleanup
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  };

  const downloadAllFiles = () => {
    if (files.length === 0) return;

    files.forEach((file, index) => {
      setTimeout(() => {
        const blob = new Blob([file.content], { type: 'text/plain' });
        const url = window.URL.createObjectURL(blob);
        
        const link = document.createElement('a');
        link.href = url;
        link.download = `${serviceName}-${file.name}`;
        document.body.appendChild(link);
        link.click();
        
        document.body.removeChild(link);
        window.URL.revokeObjectURL(url);
      }, index * 100); // Small delay between downloads
    });
  };

  const getFileTypeColor = (type: string) => {
    switch (type) {
      case 'properties': return 'bg-green-100 text-green-800';
      case 'yml': case 'yaml': return 'bg-blue-100 text-blue-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-6xl max-h-[90vh] overflow-hidden">
        <ErrorBoundarySection title="Service Files Error" description="Failed to load the service files editor.">
        <div className="flex items-center justify-between p-6 border-b">
          <div>
            <h2 className="text-xl font-semibold">Service Configuration Files</h2>
            <p className="text-sm text-muted-foreground">
              {serviceName} - {serviceDir}
            </p>
          </div>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        <div className="flex h-[70vh]">
          {/* File List Sidebar */}
          <div className="w-80 border-r bg-gray-50 dark:bg-gray-900 overflow-y-auto flex flex-col">
            <div className="p-4 border-b">
              <div className="flex items-center justify-between">
                <h3 className="font-medium">Configuration Files</h3>
                <div className="flex gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={downloadAllFiles}
                    disabled={files.length === 0}
                    title="Download all files"
                  >
                    <Download className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={fetchServiceFiles}
                    disabled={isLoading}
                    title="Refresh files"
                  >
                    <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
                  </Button>
                </div>
              </div>
            </div>
            
            <div className="p-2 flex-1 overflow-y-auto">
              {isLoading ? (
                <p className="text-center text-muted-foreground py-4">Loading files...</p>
              ) : files.length === 0 ? (
                <p className="text-center text-muted-foreground py-4">
                  No configuration files found
                </p>
              ) : (
                <div className="space-y-1">
                  {files.map((file) => (
                    <button
                      key={file.path}
                      onClick={() => handleFileSelect(file)}
                      className={`w-full text-left p-3 rounded-lg border transition-colors ${
                        selectedFile?.name === file.name
                          ? "border-blue-500 bg-blue-50 dark:bg-blue-900/20"
                          : "border-gray-200 hover:border-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
                      }`}
                    >
                      <div className="flex items-center gap-3">
                        <File className="h-4 w-4 text-gray-500" />
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="font-medium truncate">{file.name}</span>
                            <Badge className={`text-xs ${getFileTypeColor(file.type)}`}>
                              {file.type}
                            </Badge>
                          </div>
                          <p className="text-xs text-muted-foreground truncate">
                            {file.path}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            Modified: {new Date(file.lastModified).toLocaleDateString()}
                          </p>
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* File Content Editor */}
          <div className="flex-1 flex flex-col">
            {selectedFile ? (
              <>
                <div className="p-4 border-b bg-white dark:bg-gray-800">
                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium">{selectedFile.name}</h4>
                      <p className="text-sm text-muted-foreground">{selectedFile.path}</p>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={handleDownload}
                        title="Download file"
                      >
                        <Download className="h-4 w-4" />
                      </Button>
                      {!isEditing ? (
                        <Button size="sm" onClick={handleEdit}>
                          <Edit className="h-4 w-4 mr-2" />
                          Edit
                        </Button>
                      ) : (
                        <>
                          <Button variant="outline" size="sm" onClick={handleCancel}>
                            Cancel
                          </Button>
                          <Button size="sm" onClick={handleSave} disabled={isSaving}>
                            <ButtonSpinner isLoading={isSaving} loadingText="Saving...">
                              <Save className="h-4 w-4 mr-2" />
                              Save
                            </ButtonSpinner>
                          </Button>
                        </>
                      )}
                    </div>
                  </div>
                </div>
                
                <div className="flex-1 p-4 min-h-0">
                  {isEditing ? (
                    <Textarea
                      value={editedContent}
                      onChange={(e) => setEditedContent(e.target.value)}
                      className="w-full h-full font-mono text-sm resize-none overflow-auto"
                      placeholder="File content..."
                    />
                  ) : (
                    <div className="w-full h-full overflow-auto">
                      <pre className="font-mono text-sm bg-gray-50 dark:bg-gray-900 p-4 rounded border whitespace-pre-wrap min-h-full">
                        {selectedFile.content}
                      </pre>
                    </div>
                  )}
                </div>
              </>
            ) : (
              <div className="flex-1 flex items-center justify-center">
                <div className="text-center">
                  <File className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                  <p className="text-muted-foreground">
                    Select a file to view its contents
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>

        <div className="flex justify-end gap-3 p-6 border-t">
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
        </div>
        </ErrorBoundarySection>
      </div>
    </div>
  );
}