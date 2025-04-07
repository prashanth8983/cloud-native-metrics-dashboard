import React from 'react';
import { ExternalLink } from 'lucide-react';

interface StatCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon?: React.ReactNode;
  accentColor?: string;
  onClick?: () => void;
  className?: string;
  link?: string;
}

const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  subtitle,
  icon,
  accentColor = 'blue',
  onClick,
  className = '',
  link
}) => {
  const colors = {
    blue: {
      bg: 'bg-blue-50 dark:bg-blue-900/20',
      text: 'text-blue-600 dark:text-blue-400',
      iconBg: 'bg-blue-500',
      iconText: 'text-white'
    },
    green: {
      bg: 'bg-green-50 dark:bg-green-900/20',
      text: 'text-green-600 dark:text-green-400',
      iconBg: 'bg-green-500',
      iconText: 'text-white'
    },
    red: {
      bg: 'bg-red-50 dark:bg-red-900/20',
      text: 'text-red-600 dark:text-red-400',
      iconBg: 'bg-red-500',
      iconText: 'text-white'
    },
    yellow: {
      bg: 'bg-yellow-50 dark:bg-yellow-900/20',
      text: 'text-yellow-600 dark:text-yellow-400',
      iconBg: 'bg-yellow-500',
      iconText: 'text-white'
    },
    purple: {
      bg: 'bg-purple-50 dark:bg-purple-900/20',
      text: 'text-purple-600 dark:text-purple-400',
      iconBg: 'bg-purple-500',
      iconText: 'text-white'
    },
    gray: {
      bg: 'bg-gray-50 dark:bg-gray-900/20',
      text: 'text-gray-600 dark:text-gray-400',
      iconBg: 'bg-gray-500',
      iconText: 'text-white'
    }
  };

  const color = colors[accentColor as keyof typeof colors] || colors.blue;

  const CardContent = () => (
    <>
      <div className="flex justify-between">
        <div>
          <h3 className="text-lg font-medium text-gray-800 dark:text-white">{title}</h3>
          <div className="mt-1 text-3xl font-semibold text-gray-900 dark:text-white">{value}</div>
          {subtitle && <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">{subtitle}</p>}
        </div>
        
        {icon && (
          <div className={`p-3 rounded-full ${color.iconBg} ${color.iconText}`}>
            {icon}
          </div>
        )}
      </div>
      
      {link && (
        <div className="mt-4 border-t border-gray-200 dark:border-gray-700 pt-2">
          <a 
            href={link} 
            className={`flex items-center text-sm font-medium ${color.text}`}
            target="_blank"
            rel="noopener noreferrer"
          >
            View details
            <ExternalLink size={14} className="ml-1" />
          </a>
        </div>
      )}
    </>
  );

  if (onClick) {
    return (
      <button
        onClick={onClick}
        className={`w-full text-left p-5 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700 ${color.bg} hover:shadow-md transition-shadow ${className}`}
      >
        <CardContent />
      </button>
    );
  }

  return (
    <div className={`p-5 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700 ${color.bg} ${className}`}>
      <CardContent />
    </div>
  );
};

export default StatCard;