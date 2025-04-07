import React from 'react';
import {
  BarChart as RechartsBarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';
import { formatNumber } from '../../utils/formatters';
import { useTheme } from '../../context/ThemeContext';
import { chartColors } from '../../utils/colors';
import { TimeRange } from '../../types/common';

interface BarChartProps {
  data: Array<Record<string, any>>;
  dataKeys: string[];
  xAxisDataKey: string;
  height?: number | string;
  title?: string;
  loading?: boolean;
  yAxisLabel?: string;
  colors?: string[];
  stacked?: boolean;
  horizontal?: boolean;
  showLegend?: boolean;
  timeRange?: TimeRange
}

const BarChart: React.FC<BarChartProps> = ({
  data,
  dataKeys,
  xAxisDataKey,
  height = 300,
  title,
  loading = false,
  yAxisLabel,
  colors = chartColors,
  stacked = false,
  horizontal = false,
  showLegend = true
}) => {
  const { theme } = useTheme();

  if (loading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700 p-4">
        {title && <h3 className="text-md font-medium text-gray-800 dark:text-white mb-4">{title}</h3>}
        <div className="animate-pulse">
          <div className="h-64 bg-gray-200 dark:bg-gray-700 rounded"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700 p-4">
      {title && <h3 className="text-md font-medium text-gray-800 dark:text-white mb-4">{title}</h3>}
      <ResponsiveContainer width="100%" height={height}>
        <RechartsBarChart
          data={data}
          margin={{ top: 10, right: 30, left: 10, bottom: 20 }}
          layout={horizontal ? 'vertical' : 'horizontal'}
        >
          <CartesianGrid 
            strokeDasharray="3 3" 
            stroke={theme.mode === 'dark' ? '#374151' : '#e5e7eb'} 
            horizontal={!horizontal}
            vertical={horizontal}
          />
          <XAxis 
            dataKey={horizontal ? undefined : xAxisDataKey}
            type={horizontal ? 'number' : 'category'}
            tick={{ fill: theme.mode === 'dark' ? '#9ca3af' : '#6b7280' }}
            stroke={theme.mode === 'dark' ? '#4b5563' : '#d1d5db'}
          />
          <YAxis 
            dataKey={horizontal ? xAxisDataKey : undefined}
            type={horizontal ? 'category' : 'number'}
            tick={{ fill: theme.mode === 'dark' ? '#9ca3af' : '#6b7280' }}
            tickFormatter={horizontal ? undefined : formatNumber}
            stroke={theme.mode === 'dark' ? '#4b5563' : '#d1d5db'}
            label={yAxisLabel ? { 
              value: yAxisLabel, 
              angle: -90, 
              position: 'insideLeft',
              style: { fill: theme.mode === 'dark' ? '#9ca3af' : '#6b7280' }
            } : undefined}
          />
          <Tooltip 
            formatter={(value: number) => [formatNumber(value), '']}
            contentStyle={{ 
              backgroundColor: theme.mode === 'dark' ? '#1f2937' : '#ffffff',
              borderColor: theme.mode === 'dark' ? '#374151' : '#e5e7eb',
              color: theme.mode === 'dark' ? '#f9fafb' : '#111827'
            }}
          />
          {showLegend && (
            <Legend
              wrapperStyle={{ 
                paddingTop: '10px',
                color: theme.mode === 'dark' ? '#f9fafb' : '#111827' 
              }}
            />
          )}
          {dataKeys.map((dataKey, index) => (
            <Bar
              key={dataKey}
              dataKey={dataKey}
              fill={colors[index % colors.length]}
              name={dataKey}
              stackId={stacked ? 'stack' : undefined}
            />
          ))}
        </RechartsBarChart>
      </ResponsiveContainer>
    </div>
  );
};

export default BarChart;