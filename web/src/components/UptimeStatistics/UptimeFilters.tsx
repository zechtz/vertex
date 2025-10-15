import React from 'react';
import { Filter, Search, RotateCcw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export interface UptimeFilters {
  searchTerm: string;
  statusFilter: 'all' | 'running' | 'stopped' | 'unhealthy';
  uptimeFilter: 'all' | 'excellent' | 'good' | 'fair' | 'poor';
  sortBy: 'name' | 'uptime24h' | 'uptime7d' | 'restarts' | 'downtime';
  sortOrder: 'asc' | 'desc';
  timeRange: '1h' | '6h' | '24h' | '7d' | '30d';
}

interface UptimeFiltersProps {
  filters: UptimeFilters;
  onFiltersChange: (filters: UptimeFilters) => void;
  onReset: () => void;
}

export const UptimeFiltersComponent: React.FC<UptimeFiltersProps> = ({
  filters,
  onFiltersChange,
  onReset
}) => {
  const updateFilter = (key: keyof UptimeFilters, value: any) => {
    onFiltersChange({
      ...filters,
      [key]: value
    });
  };

  const timeRangeOptions = [
    { value: '1h', label: '1 Hour' },
    { value: '6h', label: '6 Hours' },
    { value: '24h', label: '24 Hours' },
    { value: '7d', label: '7 Days' },
    { value: '30d', label: '30 Days' }
  ];

  const statusOptions = [
    { value: 'all', label: 'All Status' },
    { value: 'running', label: 'Running' },
    { value: 'stopped', label: 'Stopped' },
    { value: 'unhealthy', label: 'Unhealthy' }
  ];

  const uptimeOptions = [
    { value: 'all', label: 'All Uptime' },
    { value: 'excellent', label: 'Excellent (≥99.9%)' },
    { value: 'good', label: 'Good (≥99%)' },
    { value: 'fair', label: 'Fair (≥95%)' },
    { value: 'poor', label: 'Poor (<95%)' }
  ];

  const sortOptions = [
    { value: 'name', label: 'Service Name' },
    { value: 'uptime24h', label: '24h Uptime' },
    { value: 'uptime7d', label: '7d Uptime' },
    { value: 'restarts', label: 'Restart Count' },
    { value: 'downtime', label: 'Downtime' }
  ];

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4 space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <Filter className="w-4 h-4 text-gray-500" />
          <span className="font-medium text-sm text-gray-700 dark:text-gray-300">Filters & Options</span>
        </div>
        <Button
          onClick={onReset}
          variant="outline"
          size="sm"
          className="text-xs"
        >
          <RotateCcw className="w-3 h-3 mr-1" />
          Reset
        </Button>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
        {/* Search */}
        <div className="space-y-1">
          <label className="text-xs font-medium text-gray-600 dark:text-gray-400">Search</label>
          <div className="relative">
            <Search className="w-3 h-3 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
            <Input
              type="text"
              placeholder="Service name..."
              value={filters.searchTerm}
              onChange={(e) => updateFilter('searchTerm', e.target.value)}
              className="pl-8 h-8 text-sm"
            />
          </div>
        </div>

        {/* Time Range */}
        <div className="space-y-1">
          <label className="text-xs font-medium text-gray-600 dark:text-gray-400">Time Range</label>
          <select
            value={filters.timeRange}
            onChange={(e) => updateFilter('timeRange', e.target.value)}
            className="w-full h-8 px-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          >
            {timeRangeOptions.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        {/* Status Filter */}
        <div className="space-y-1">
          <label className="text-xs font-medium text-gray-600 dark:text-gray-400">Status</label>
          <select
            value={filters.statusFilter}
            onChange={(e) => updateFilter('statusFilter', e.target.value)}
            className="w-full h-8 px-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          >
            {statusOptions.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        {/* Uptime Filter */}
        <div className="space-y-1">
          <label className="text-xs font-medium text-gray-600 dark:text-gray-400">Uptime</label>
          <select
            value={filters.uptimeFilter}
            onChange={(e) => updateFilter('uptimeFilter', e.target.value)}
            className="w-full h-8 px-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          >
            {uptimeOptions.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        {/* Sort By */}
        <div className="space-y-1">
          <label className="text-xs font-medium text-gray-600 dark:text-gray-400">Sort By</label>
          <select
            value={filters.sortBy}
            onChange={(e) => updateFilter('sortBy', e.target.value)}
            className="w-full h-8 px-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          >
            {sortOptions.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        {/* Sort Order */}
        <div className="space-y-1">
          <label className="text-xs font-medium text-gray-600 dark:text-gray-400">Order</label>
          <select
            value={filters.sortOrder}
            onChange={(e) => updateFilter('sortOrder', e.target.value)}
            className="w-full h-8 px-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          >
            <option value="asc">Ascending</option>
            <option value="desc">Descending</option>
          </select>
        </div>
      </div>
    </div>
  );
};