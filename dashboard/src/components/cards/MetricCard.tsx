import React from 'react';
import { ArrowUpRight, ArrowDownRight, MoreHorizontal, TrendingUp } from 'lucide-react';
import { formatNumber, formatPercent } from '../../utils/formatters';

interface MetricCardProps {
  title: string;
  value: number | null;
  previousValue?: number | null;
  change?: number;
  changeType?: 'positive' | 'negative' | 'neutral';
  icon?: React.ReactNode;
  onClick?: () => void;
  loading?: boolean;
}

const MetricCard: React.FC<MetricCardProps> = ({ 
  title, 
  value, 
  previousValue, 
  change, 
  changeType = 'neutral', 
  icon, 
  onClick, 
  loading = false 
}) => {
  console.log(previousValue)
  return (
   
    <div 
      className={`bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700 p-4 ${onClick ? 'cursor-pointer hover:shadow-md transition-shadow' : ''}`}
      onClick={onClick}
    >
      <div className="flex justify-between items-start">
        <div className="text-sm font-medium text-gray-500 dark:text-gray-400">
          {title}
        </div>
        {icon ? (
          <div className="p-2 bg-blue-50 dark:bg-blue-900/20 rounded-md text-blue-500 dark:text-blue-400">
            {icon}
          </div>
        ) : (
          <button className="text-gray-400 hover:text-gray-500 dark:text-gray-500 dark:hover:text-gray-400">
            <MoreHorizontal size={16} />
          </button>
        )}
      </div>
      
      <div className="mt-2">
        {loading ? (
          <div className="h-8 w-24 bg-gray-200 dark:bg-gray-700 rounded animate-pulse"></div>
        ) : (
          <div className="text-2xl font-bold text-gray-800 dark:text-white">
            {value !== null ? formatNumber(value) : 'N/A'}
          </div>
        )}
      </div>
      
      {change !== undefined && (
        <div className="mt-2 flex items-center">
          <div 
            className={`flex items-center text-sm 
              ${changeType === 'positive' ? 'text-green-500' : ''} 
              ${changeType === 'negative' ? 'text-red-500' : ''} 
              ${changeType === 'neutral' ? 'text-gray-500 dark:text-gray-400' : ''}`}
          >
            {changeType === 'positive' && <ArrowUpRight size={16} className="mr-1" />}
            {changeType === 'negative' && <ArrowDownRight size={16} className="mr-1" />}
            {changeType === 'neutral' && <TrendingUp size={16} className="mr-1" />}
            <span>{change >= 0 ? '+' : ''}{formatPercent(change / 100)}</span>
          </div>
          <div className="ml-2 text-xs text-gray-500 dark:text-gray-400">
            vs previous
          </div>
        </div>
      )}
    </div>
  );
};

export default MetricCard;