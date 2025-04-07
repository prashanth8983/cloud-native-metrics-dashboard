import React from 'react';
import { AlertCircle, Clock, ExternalLink } from 'lucide-react';
import { Alert } from '../../types/alerts';
import { formatTimestamp } from '../../utils/formatters';
import { getSeverityColor } from '../../utils/colors';

interface AlertCardProps {
  alert: Alert;
  onClick?: () => void;
}

const AlertCard: React.FC<AlertCardProps> = ({ alert, onClick }) => {
  // Determine background color based on severity
  const getSeverityClass = () => {
    switch (alert.severity.toLowerCase()) {
      case 'critical':
        return 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800';
      case 'high':
      case 'warning':
        return 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800';
      case 'medium':
        return 'bg-orange-50 dark:bg-orange-900/20 border-orange-200 dark:border-orange-800';
      case 'low':
        return 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800';
      case 'info':
        return 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800';
      default:
        return 'bg-gray-50 dark:bg-gray-900/20 border-gray-200 dark:border-gray-700';
    }
  };

  // Get icon color
  const iconColor = getSeverityColor(alert.severity);

  return (
    <div 
      className={`rounded-lg border shadow-sm p-4 ${getSeverityClass()} ${onClick ? 'cursor-pointer hover:shadow-md transition-shadow' : ''}`}
      onClick={onClick}
    >
      <div className="flex items-start">
        <div className="flex-shrink-0 mr-3">
          <AlertCircle size={20} style={{ color: iconColor }} />
        </div>
        <div className="flex-1">
          <div className="flex justify-between items-start">
            <h3 className="text-sm font-medium text-gray-900 dark:text-white">
              {alert.name}
            </h3>
            <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
              alert.state === 'firing' ? 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-400' : 
              alert.state === 'pending' ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-400' :
              'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-400'
            }`}>
              {alert.state}
            </span>
          </div>
          <p className="mt-1 text-sm text-gray-600 dark:text-gray-300">{alert.summary}</p>
          
          {Object.keys(alert.labels).length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1">
              {Object.entries(alert.labels)
                .filter(([key]) => key !== 'severity' && key !== 'alertname')
                .map(([key, value]) => (
                  <span 
                    key={key} 
                    className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200"
                  >
                    {key}={value}
                  </span>
                ))
              }
            </div>
          )}
          
          <div className="mt-2 flex items-center text-sm text-gray-500 dark:text-gray-400">
            <Clock size={14} className="mr-1" />
            <span>Active since {formatTimestamp(alert.activeAt)}</span>
          </div>
        </div>
      </div>
      
      {/* Optional actions */}
      <div className="mt-3 border-t border-gray-200 dark:border-gray-700 pt-3 flex justify-between">
        <div className="flex space-x-2">
          <button className="text-xs font-medium text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300">
            Silence
          </button>
          <button className="text-xs font-medium text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-300">
            View similar
          </button>
        </div>
        <a 
          href="#" 
          className="text-xs font-medium text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 flex items-center"
        >
          View in Alertmanager
          <ExternalLink size={12} className="ml-1" />
        </a>
      </div>
    </div>
  );
};

export default AlertCard;