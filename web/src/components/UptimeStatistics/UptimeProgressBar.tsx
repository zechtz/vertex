import React from 'react';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';

interface UptimeProgressBarProps {
  percentage: number;
  size?: 'sm' | 'md' | 'lg';
  showPercentage?: boolean;
  trend?: number; // Previous period comparison
  className?: string;
}

export const UptimeProgressBar: React.FC<UptimeProgressBarProps> = ({
  percentage,
  size = 'md',
  showPercentage = true,
  trend,
  className = ''
}) => {
  const getColorClass = (pct: number) => {
    if (pct >= 99) return 'bg-green-500';
    if (pct >= 95) return 'bg-yellow-500';
    return 'bg-red-500';
  };

  const getBackgroundColorClass = (pct: number) => {
    if (pct >= 99) return 'bg-green-100 dark:bg-green-900/20';
    if (pct >= 95) return 'bg-yellow-100 dark:bg-yellow-900/20';
    return 'bg-red-100 dark:bg-red-900/20';
  };

  const getTextColorClass = (pct: number) => {
    if (pct >= 99) return 'text-green-700 dark:text-green-300';
    if (pct >= 95) return 'text-yellow-700 dark:text-yellow-300';
    return 'text-red-700 dark:text-red-300';
  };

  const sizeClasses = {
    sm: 'h-1.5',
    md: 'h-2',
    lg: 'h-3'
  };

  const TrendIcon = ({ trend }: { trend: number }) => {
    if (Math.abs(trend) < 0.1) return <Minus className="w-3 h-3" />;
    return trend > 0 ? 
      <TrendingUp className="w-3 h-3 text-green-600" /> : 
      <TrendingDown className="w-3 h-3 text-red-600" />;
  };

  return (
    <div className={`flex items-center space-x-3 ${className}`}>
      <div className="flex-1">
        <div className={`w-full ${getBackgroundColorClass(percentage)} rounded-full ${sizeClasses[size]} overflow-hidden`}>
          <div 
            className={`${sizeClasses[size]} rounded-full ${getColorClass(percentage)} transition-all duration-500 ease-out`}
            style={{ width: `${Math.min(percentage, 100)}%` }}
          />
        </div>
      </div>
      
      {showPercentage && (
        <div className="flex items-center space-x-1 min-w-0">
          <span className={`text-sm font-medium ${getTextColorClass(percentage)}`}>
            {percentage.toFixed(2)}%
          </span>
          {trend !== undefined && (
            <div className="flex items-center">
              <TrendIcon trend={trend} />
              {Math.abs(trend) >= 0.1 && (
                <span className={`text-xs ml-1 ${trend > 0 ? 'text-green-600' : 'text-red-600'}`}>
                  {Math.abs(trend).toFixed(1)}%
                </span>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
};