import React, { useState } from 'react';
import { Copy, Check } from 'lucide-react';
import { Button } from './button';

interface CopyButtonProps {
  text: string;
  className?: string;
  size?: 'sm' | 'default' | 'lg';
  variant?: 'outline' | 'ghost' | 'default';
  label?: string;
}

export const CopyButton: React.FC<CopyButtonProps> = ({
  text,
  className = '',
  size = 'sm',
  variant = 'outline',
  label = 'Copy'
}) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy text:', err);
    }
  };

  return (
    <Button
      onClick={handleCopy}
      variant={variant}
      size={size}
      className={`${className} transition-all duration-200`}
    >
      {copied ? (
        <Check className="w-4 h-4 text-green-500" />
      ) : (
        <Copy className="w-4 h-4" />
      )}
      <span className="ml-2">{copied ? 'Copied!' : label}</span>
    </Button>
  );
};