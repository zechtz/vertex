import { useState, useEffect } from 'react';
import { Search, Download, Calendar } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Service } from '@/types';

interface LogSearchResult {
  id: number;
  serviceName: string;
  timestamp: string;
  level: string;
  message: string;
  createdAt: string;
}

interface LogSearchResponse {
  results: LogSearchResult[];
  totalCount: number;
  limit: number;
  offset: number;
}

interface LogSearchProps {
  services: Service[];
  className?: string;
}

export function LogSearch({ services, className = '' }: LogSearchProps) {
  const [searchText, setSearchText] = useState('');
  const [selectedServices, setSelectedServices] = useState<string[]>([]);
  const [selectedLevels, setSelectedLevels] = useState<string[]>(['INFO', 'WARN', 'ERROR']);
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [searchResults, setSearchResults] = useState<LogSearchResult[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [isSearching, setIsSearching] = useState(false);
  const [isExporting, setIsExporting] = useState(false);

  const logLevels = ['INFO', 'WARN', 'ERROR', 'DEBUG', 'TRACE'];
  const resultsPerPage = 50;

  // Auto-search when filters change
  useEffect(() => {
    const searchTimeout = setTimeout(() => {
      if (searchText || selectedServices.length > 0 || selectedLevels.length < logLevels.length) {
        performSearch();
      }
    }, 500);

    return () => clearTimeout(searchTimeout);
  }, [searchText, selectedServices, selectedLevels, startDate, endDate, currentPage]);

  const performSearch = async () => {
    try {
      setIsSearching(true);

      const searchCriteria = {
        serviceNames: selectedServices,
        levels: selectedLevels,
        searchText: searchText,
        startTime: startDate ? new Date(startDate).toISOString() : '',
        endTime: endDate ? new Date(endDate).toISOString() : '',
        limit: resultsPerPage,
        offset: (currentPage - 1) * resultsPerPage,
      };

      const response = await fetch('/api/logs/search', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(searchCriteria),
      });

      if (!response.ok) {
        throw new Error(`Search failed: ${response.status} ${response.statusText}`);
      }

      const data: LogSearchResponse = await response.json();
      setSearchResults(data.results);
      setTotalCount(data.totalCount);
    } catch (error) {
      console.error('Log search failed:', error);
      setSearchResults([]);
      setTotalCount(0);
    } finally {
      setIsSearching(false);
    }
  };

  const exportLogs = async (format: 'json' | 'csv' | 'txt') => {
    try {
      setIsExporting(true);

      const exportRequest = {
        serviceNames: selectedServices,
        levels: selectedLevels,
        searchText: searchText,
        startTime: startDate ? new Date(startDate).toISOString() : '',
        endTime: endDate ? new Date(endDate).toISOString() : '',
        format: format,
      };

      const response = await fetch('/api/logs/export', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(exportRequest),
      });

      if (!response.ok) {
        throw new Error(`Export failed: ${response.status} ${response.statusText}`);
      }

      // Download the file
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `nest_logs_${new Date().toISOString().split('T')[0]}.${format}`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (error) {
      console.error('Log export failed:', error);
    } finally {
      setIsExporting(false);
    }
  };

  const toggleService = (serviceName: string) => {
    setSelectedServices(prev => 
      prev.includes(serviceName)
        ? prev.filter(s => s !== serviceName)
        : [...prev, serviceName]
    );
    setCurrentPage(1);
  };

  const toggleLevel = (level: string) => {
    setSelectedLevels(prev => 
      prev.includes(level)
        ? prev.filter(l => l !== level)
        : [...prev, level]
    );
    setCurrentPage(1);
  };

  const getLevelColor = (level: string) => {
    switch (level.toUpperCase()) {
      case 'ERROR':
        return 'bg-red-100 text-red-800';
      case 'WARN':
        return 'bg-yellow-100 text-yellow-800';
      case 'INFO':
        return 'bg-blue-100 text-blue-800';
      case 'DEBUG':
        return 'bg-gray-100 text-gray-800';
      case 'TRACE':
        return 'bg-purple-100 text-purple-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const highlightSearchTerm = (text: string, searchTerm: string) => {
    if (!searchTerm) return text;
    
    const regex = new RegExp(`(${searchTerm})`, 'gi');
    const parts = text.split(regex);
    
    return parts.map((part, index) => 
      regex.test(part) ? (
        <mark key={index} className="bg-yellow-200 px-1 rounded">{part}</mark>
      ) : part
    );
  };

  const totalPages = Math.ceil(totalCount / resultsPerPage);

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Search Controls */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Search className="h-5 w-5" />
            Advanced Log Search
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Text Search */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Search Text
            </label>
            <div className="relative">
              <Search className="absolute left-3 top-3 h-4 w-4 text-gray-400" />
              <input
                type="text"
                value={searchText}
                onChange={(e) => {
                  setSearchText(e.target.value);
                  setCurrentPage(1);
                }}
                placeholder="Search log messages..."
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </div>

          {/* Service Filter */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Services ({selectedServices.length === 0 ? 'All' : selectedServices.length} selected)
            </label>
            <div className="flex flex-wrap gap-2">
              {services.map((service) => (
                <Button
                  key={service.name}
                  variant={selectedServices.includes(service.name) ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => toggleService(service.name)}
                  className="text-xs"
                >
                  {service.name}
                  {service.status === 'running' && (
                    <div className="ml-1 w-2 h-2 bg-green-500 rounded-full"></div>
                  )}
                </Button>
              ))}
            </div>
          </div>

          {/* Level Filter */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Log Levels
            </label>
            <div className="flex flex-wrap gap-2">
              {logLevels.map((level) => (
                <Button
                  key={level}
                  variant={selectedLevels.includes(level) ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => toggleLevel(level)}
                  className={`text-xs ${selectedLevels.includes(level) ? getLevelColor(level) : ''}`}
                >
                  {level}
                </Button>
              ))}
            </div>
          </div>

          {/* Date Range */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Start Date
              </label>
              <div className="relative">
                <Calendar className="absolute left-3 top-3 h-4 w-4 text-gray-400" />
                <input
                  type="datetime-local"
                  value={startDate}
                  onChange={(e) => {
                    setStartDate(e.target.value);
                    setCurrentPage(1);
                  }}
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                End Date
              </label>
              <div className="relative">
                <Calendar className="absolute left-3 top-3 h-4 w-4 text-gray-400" />
                <input
                  type="datetime-local"
                  value={endDate}
                  onChange={(e) => {
                    setEndDate(e.target.value);
                    setCurrentPage(1);
                  }}
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
            </div>
          </div>

          {/* Export Options */}
          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-600">
              {totalCount > 0 && (
                <span>Found {totalCount.toLocaleString()} log entries</span>
              )}
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => exportLogs('json')}
                disabled={isExporting || searchResults.length === 0}
              >
                <Download className="h-4 w-4 mr-1" />
                JSON
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => exportLogs('csv')}
                disabled={isExporting || searchResults.length === 0}
              >
                <Download className="h-4 w-4 mr-1" />
                CSV
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => exportLogs('txt')}
                disabled={isExporting || searchResults.length === 0}
              >
                <Download className="h-4 w-4 mr-1" />
                TXT
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Search Results */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span>Search Results</span>
            {isSearching && (
              <div className="flex items-center text-sm text-gray-500">
                <div className="animate-spin w-4 h-4 border-2 border-gray-300 border-t-blue-600 rounded-full mr-2"></div>
                Searching...
              </div>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {searchResults.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              {isSearching ? 'Searching logs...' : 'No log entries found matching your criteria.'}
            </div>
          ) : (
            <div className="space-y-4">
              {/* Results List */}
              <div className="space-y-2">
                {searchResults.map((result) => (
                  <div
                    key={result.id}
                    className="flex items-start space-x-3 p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                  >
                    <div className="flex-shrink-0 text-xs text-gray-500 w-32">
                      {new Date(result.timestamp).toLocaleString()}
                    </div>
                    <Badge variant="outline" className="flex-shrink-0">
                      {result.serviceName}
                    </Badge>
                    <Badge className={`flex-shrink-0 text-xs ${getLevelColor(result.level)}`}>
                      {result.level}
                    </Badge>
                    <div className="flex-1 text-sm font-mono text-gray-800 break-all">
                      {highlightSearchTerm(result.message, searchText)}
                    </div>
                  </div>
                ))}
              </div>

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-between">
                  <div className="text-sm text-gray-600">
                    Page {currentPage} of {totalPages} 
                    ({((currentPage - 1) * resultsPerPage) + 1}-{Math.min(currentPage * resultsPerPage, totalCount)} 
                    of {totalCount.toLocaleString()} entries)
                  </div>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                      disabled={currentPage === 1 || isSearching}
                    >
                      Previous
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                      disabled={currentPage === totalPages || isSearching}
                    >
                      Next
                    </Button>
                  </div>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default LogSearch;