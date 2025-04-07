import React, { useState } from 'react';
import { ChevronDown, ChevronUp, Search, ArrowUp, ArrowDown } from 'lucide-react';
import { TopMetric } from '../../types/metrics';
import { formatNumber } from '../../utils/formatters';

interface MetricsTableProps {
  metrics: TopMetric[];
  loading?: boolean;
  onMetricClick?: (metric: TopMetric) => void;
}

const MetricsTable: React.FC<MetricsTableProps> = ({ 
  metrics, 
  loading = false,
  onMetricClick
}) => {
  const [sortField, setSortField] = useState<'name' | 'cardinality' | 'sampleRate'>('cardinality');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');
  const [searchTerm, setSearchTerm] = useState('');
  const [expanded, setExpanded] = useState(false);

  // Filter and sort metrics
  const filteredMetrics = metrics
    .filter(metric => metric.name.toLowerCase().includes(searchTerm.toLowerCase()))
    .sort((a, b) => {
      const aValue = a[sortField];
      const bValue = b[sortField];

      if (typeof aValue === 'string' && typeof bValue === 'string') {
        return sortDirection === 'asc'
          ? aValue.localeCompare(bValue)
          : bValue.localeCompare(aValue);
      }

      return sortDirection === 'asc'
        ? (aValue as number) - (bValue as number)
        : (bValue as number) - (aValue as number);
    });

  // Limit display to 10 metrics unless expanded
  const displayedMetrics = expanded ? filteredMetrics : filteredMetrics.slice(0, 10);

  const handleSort = (field: 'name' | 'cardinality' | 'sampleRate') => {
    if (field === sortField) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  if (loading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700 overflow-hidden">
        <div className="animate-pulse">
          <div className="h-12 bg-gray-200 dark:bg-gray-700"></div>
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-12 bg-gray-100 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700"></div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700 overflow-hidden">
      <div className="p-4 flex items-center justify-between border-b border-gray-100 dark:border-gray-700">
        <h3 className="text-md font-medium text-gray-800 dark:text-white">Top Metrics</h3>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
          <input
            type="text"
            placeholder="Search metrics..."
            className="pl-9 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-900">
            <tr>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer"
                onClick={() => handleSort('name')}
              >
                <div className="flex items-center">
                  Metric Name
                  {sortField === 'name' ? (
                    sortDirection === 'asc' ? <ChevronUp className="ml-1 h-4 w-4" /> : <ChevronDown className="ml-1 h-4 w-4" />
                  ) : null}
                </div>
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer"
                onClick={() => handleSort('cardinality')}
              >
                <div className="flex items-center">
                  Cardinality
                  {sortField === 'cardinality' ? (
                    sortDirection === 'asc' ? <ChevronUp className="ml-1 h-4 w-4" /> : <ChevronDown className="ml-1 h-4 w-4" />
                  ) : null}
                </div>
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer"
                onClick={() => handleSort('sampleRate')}
              >
                <div className="flex items-center">
                  Sample Rate
                  {sortField === 'sampleRate' ? (
                    sortDirection === 'asc' ? <ChevronUp className="ml-1 h-4 w-4" /> : <ChevronDown className="ml-1 h-4 w-4" />
                  ) : null}
                </div>
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Trend
              </th>
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            {displayedMetrics.map((metric) => (
              <tr 
                key={metric.name} 
                className={onMetricClick ? "hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer" : ""}
                onClick={() => onMetricClick && onMetricClick(metric)}
              >
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                  {metric.name}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {formatNumber(metric.cardinality)}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {formatNumber(metric.sampleRate)} / sec
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {/* Mock trend indicator - in a real app, this would be calculated */}
                  {Math.random() > 0.5 ? (
                    <ArrowUp className="text-green-500 h-4 w-4 inline" />
                  ) : (
                    <ArrowDown className="text-red-500 h-4 w-4 inline" />
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {filteredMetrics.length > 10 && (
        <div className="p-2 flex justify-center border-t border-gray-200 dark:border-gray-700">
          <button
            className="text-sm text-blue-500 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 flex items-center"
            onClick={() => setExpanded(!expanded)}
          >
            {expanded ? 'Show Less' : `Show All (${filteredMetrics.length})`}
            {expanded ? <ChevronUp className="ml-1 h-4 w-4" /> : <ChevronDown className="ml-1 h-4 w-4" />}
          </button>
        </div>
      )}
    </div>
  );
};

export default MetricsTable;