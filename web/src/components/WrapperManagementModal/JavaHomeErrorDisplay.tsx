import React from 'react';
import { AlertTriangle, Coffee, Terminal } from 'lucide-react';
import { CopyButton } from '@/components/ui/copy-button';

interface JavaHomeErrorDisplayProps {
  error: string;
}

export const JavaHomeErrorDisplay: React.FC<JavaHomeErrorDisplayProps> = ({ error }) => {
  // Check if this is a JAVA_HOME related error
  const isJavaHomeError = error.includes('JAVA_HOME environment variable is not set');
  
  if (!isJavaHomeError) {
    return (
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
        <div className="flex items-center gap-2">
          <AlertTriangle className="w-5 h-5 text-red-500" />
          <span className="font-medium text-red-700 dark:text-red-400">Error</span>
        </div>
        <p className="mt-2 text-sm text-red-600 dark:text-red-300">{error}</p>
      </div>
    );
  }

  // Parse the error message for different shell configurations
  const shellConfigs = [
    {
      shell: 'Bash',
      file: '~/.bashrc',
      command: 'export JAVA_HOME=/path/to/java',
      reload: 'source ~/.bashrc'
    },
    {
      shell: 'Zsh',
      file: '~/.zshrc', 
      command: 'export JAVA_HOME=/path/to/java',
      reload: 'source ~/.zshrc'
    },
    {
      shell: 'Fish',
      file: '~/.config/fish/config.fish',
      command: 'set -x JAVA_HOME /path/to/java',
      reload: 'source ~/.config/fish/config.fish'
    }
  ];

  const findJavaCommands = [
    {
      os: 'macOS',
      command: '/usr/libexec/java_home'
    },
    {
      os: 'Linux',
      command: 'which java'
    },
    {
      os: 'Linux (alternative)',
      command: 'whereis java'
    }
  ];

  return (
    <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-4 space-y-4">
      <div className="flex items-center gap-2">
        <Coffee className="w-5 h-5 text-amber-600" />
        <span className="font-medium text-amber-800 dark:text-amber-400">JAVA_HOME Not Set</span>
      </div>
      
      <p className="text-sm text-amber-700 dark:text-amber-300">
        The JAVA_HOME environment variable is required for wrapper validation but is not currently set.
      </p>

      <div className="space-y-4">
        <div>
          <h4 className="font-medium text-amber-800 dark:text-amber-400 mb-2 flex items-center gap-2">
            <Terminal className="w-4 h-4" />
            1. Find your Java installation
          </h4>
          <div className="space-y-2">
            {findJavaCommands.map((cmd, index) => (
              <div key={index} className="flex items-center justify-between bg-amber-100 dark:bg-amber-900/40 p-2 rounded">
                <div>
                  <span className="text-xs font-medium text-amber-700 dark:text-amber-400">{cmd.os}:</span>
                  <code className="ml-2 text-sm font-mono text-amber-800 dark:text-amber-200">{cmd.command}</code>
                </div>
                <CopyButton text={cmd.command} size="sm" variant="ghost" label="Copy" />
              </div>
            ))}
          </div>
        </div>

        <div>
          <h4 className="font-medium text-amber-800 dark:text-amber-400 mb-2">
            2. Set JAVA_HOME in your shell configuration
          </h4>
          <div className="space-y-3">
            {shellConfigs.map((config, index) => (
              <div key={index} className="bg-amber-100 dark:bg-amber-900/40 p-3 rounded">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-amber-800 dark:text-amber-300">
                    {config.shell} ({config.file})
                  </span>
                </div>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <code className="text-sm font-mono text-amber-800 dark:text-amber-200">
                      {config.command}
                    </code>
                    <CopyButton text={config.command} size="sm" variant="ghost" label="Copy" />
                  </div>
                  <div className="flex items-center justify-between">
                    <code className="text-xs font-mono text-amber-700 dark:text-amber-300">
                      {config.reload}
                    </code>
                    <CopyButton text={config.reload} size="sm" variant="ghost" label="Copy" />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-amber-100 dark:bg-amber-900/40 p-3 rounded">
          <h4 className="font-medium text-amber-800 dark:text-amber-400 mb-2">
            3. Restart your terminal or reload configuration
          </h4>
          <p className="text-sm text-amber-700 dark:text-amber-300">
            After setting JAVA_HOME, restart your terminal or run the appropriate source command above.
            Then restart Vertex to pick up the new environment variable.
          </p>
        </div>

        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 p-3 rounded">
          <h4 className="font-medium text-blue-800 dark:text-blue-400 mb-1">💡 Example</h4>
          <p className="text-sm text-blue-700 dark:text-blue-300 mb-2">
            If Java is installed at <code>/usr/lib/jvm/java-17-openjdk</code>:
          </p>
          <div className="flex items-center justify-between bg-blue-100 dark:bg-blue-900/40 p-2 rounded">
            <code className="text-sm font-mono text-blue-800 dark:text-blue-200">
              export JAVA_HOME=/usr/lib/jvm/java-17-openjdk
            </code>
            <CopyButton 
              text="export JAVA_HOME=/usr/lib/jvm/java-17-openjdk" 
              size="sm" 
              variant="ghost" 
              label="Copy" 
            />
          </div>
        </div>
      </div>
    </div>
  );
};