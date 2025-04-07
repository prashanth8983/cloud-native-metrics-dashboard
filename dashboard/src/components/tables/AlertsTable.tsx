import React, { useState } from 'react';
import { ChevronDown, ChevronUp, Search, ExternalLink } from 'lucide-react';
import { Alert } from '../../types/alerts';
import { formatTimestamp } from '../../utils/formatters';
import { getSeverityColor } from '../../utils/colors';

interface AlertsTableProps {
  alerts: Alert[];
  loading?: boolean;
  onAlertClick?: (alert: Alert) => void;
}

const AlertsTable: React.FC<AlertsTableProps> = ({ 
  alerts, 
  loading = false,
  onAlertClick
}) => {
  const [sortField, setSortField] = useState<'name' | 'state' | 'severity' | 'activeAt'>('activeAt');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');
  const [searchTerm, setSearchTerm] = useState('');
  const [expanded, setExpanded] = useState(false);

  // Filter and sort alerts
  const filteredAlerts = alerts
    .filter(alert => 
      alert.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      alert.summary.toLowerCase().includes(searchTerm.toLowerCase())
    )
    .sort((a, b) => {
      if (sortField === 'activeAt') {
        const dateA = new Date(a.activeAt).getTime();
        const dateB = new Date(b.activeAt).getTime();
        return sortDirection === 'asc' ? dateA - dateB : dateB - dateA;
      }

      const aValue = a[sortField as keyof Alert];
      const bValue = b[sortField as keyof Alert];

      if (typeof aValue === 'string' && typeof bValue === 'string') {
        return sortDirection === 'asc'
          ? aValue.localeCompare(bValue)
          : bValue.localeCompare(aValue);
      }

      return 0;
    });

  // Limit display to 10 alerts unless expanded
  const displayedAlerts = expanded ? filteredAlerts : filteredAlerts.slice(0, 10);

  const handleSort = (field: 'name' | 'state' | 'severity' | 'activeAt') => {
    if (field === sortField) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  // Function to get background color class for alert state
  const getStateBgClass = (state: string) => {
    switch (state) {
      case 'firing':
        return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
      case 'pending':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
      default:
        return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
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
        <h3 className="text-md font-medium text-gray-800 dark:text-white">Current Alerts</h3>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
          <input
            type="text"
            placeholder="Search alerts..."
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
                  Alert Name
                  {sortField === 'name' ? (
                    sortDirection === 'asc' ? <ChevronUp className="ml-1 h-4 w-4" /> : <ChevronDown className="ml-1 h-4 w-4" />
                  ) : null}
                </div>
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer"
                onClick={() => handleSort('state')}
              >
                <div className="flex items-center">
                  State
                  {sortField === 'state' ? (
                    sortDirection === 'asc' ? <ChevronUp className="ml-1 h-4 w-4" /> : <ChevronDown className="ml-1 h-4 w-4" />
                  ) : null}
                </div>
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer"
                onClick={() => handleSort('severity')}
              >
                <div className="flex items-center">
                  Severity
                  {sortField === 'severity' ? (
                    sortDirection === 'asc' ? <ChevronUp className="ml-1 h-4 w-4" /> : <ChevronDown className="ml-1 h-4 w-4" />
                  ) : null}
                </div>
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer"
                onClick={() => handleSort('activeAt')}
              >
                <div className="flex items-center">
                  Since
                  {sortField === 'activeAt' ? (
                    sortDirection === 'asc' ? <ChevronUp className="ml-1 h-4 w-4" /> : <ChevronDown className="ml-1 h-4 w-4" />
                  ) : null}
                </div>
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Summary
              </th>
              <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            {displayedAlerts.map((alert) => (
              <tr 
                key={`${alert.name}-${alert.activeAt}`} 
                className={onAlertClick ? "hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer" : ""}
                onClick={() => onAlertClick && onAlertClick(alert)}
              >
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                  {alert.name}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getStateBgClass(alert.state)}`}>
                    {alert.state}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="text-sm" style={{ color: getSeverityColor(alert.severity) }}>
                    {alert.severity}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {formatTimestamp(alert.activeAt)}
                </td>
                <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400 truncate max-w-xs">
                  {alert.summary}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <button
                    className="text-blue-600 dark:text-blue-400 hover:text-blue-900 dark:hover:text-blue-300 mr-3"
                    onClick={(e) => {
                      e.stopPropagation();
                      // Silence action
                    }}
                  >
                    Silence
                  </button>
                  <a
                    href="#"
                    className="text-blue-600 dark:text-blue-400 hover:text-blue-900 dark:hover:text-blue-300 inline-flex items-center"
                    onClick={(e) => e.stopPropagation()}
                  >
                    View
                    <ExternalLink className="ml-1 h-3 w-3" />
                  </a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {filteredAlerts.length > 10 && (
        <div className="p-2 flex justify-center border-t border-gray-200 dark:border-gray-700">
          <button
            className="text-sm text-blue-500 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 flex items-center"
            onClick={() => setExpanded(!expanded)}
          >
            {expanded ? 'Show Less' : `Show All (${filteredAlerts.length})`}
            {expanded ? <ChevronUp className="ml-1 h-4 w-4" /> : <ChevronDown className="ml-1 h-4 w-4" />}
          </button>
        </div>
      )}
    </div>
  );
};

export default AlertsTable;