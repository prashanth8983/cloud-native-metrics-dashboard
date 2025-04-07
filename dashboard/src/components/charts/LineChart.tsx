import React from 'react';
import { 
  LineChart as RechartsLineChart, 
  Line, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  Legend, 
  ResponsiveContainer 
} from 'recharts';
import { formatDateForRange } from '../../utils/time';
import { formatNumber } from '../../utils/formatters';
import { useTheme } from '../../context/ThemeContext';
import { stringToColor } from '../../utils/colors';
import { TimeRange } from '../../types/common';

interface LineChartProps {
  data: Array<Record<string, any>>;
  dataKeys: string[];
  xAxisDataKey: string;
  timeRange?: TimeRange;
  height?: number | string;
  title?: string;
  loading?: boolean;
  yAxisLabel?: string;
  colors?: string[];
  onPointClick?: (point: any) => void;
}

const LineChart: React.FC<LineChartProps> = ({
  data,
  dataKeys,
  xAxisDataKey,
  timeRange,
  height = 300,
  title,
  loading = false,
  yAxisLabel,
  colors,
  onPointClick
}) => {
  const { theme } = useTheme();
  const isTimeAxis = xAxisDataKey === 'timestamp' || xAxisDataKey === 'time' || xAxisDataKey === 'date';
  
  // Generate colors if not provided
  const lineColors = colors || dataKeys.map(key => stringToColor(key));
  
  // Format the x-axis labels
  const formatXAxis = (value: any) => {
    if (isTimeAxis && timeRange) {
      return formatDateForRange(new Date(value), timeRange.start, timeRange.end);
    }
    return value;
  };

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
        <RechartsLineChart
          data={data}
          margin={{ top: 10, right: 30, left: 10, bottom: 20 }}
          onClick={onPointClick}
        >
          <CartesianGrid 
            strokeDasharray="3 3" 
            stroke={theme.mode === 'dark' ? '#374151' : '#e5e7eb'} 
            vertical={false} 
          />
          <XAxis 
            dataKey={xAxisDataKey} 
            tick={{ fill: theme.mode === 'dark' ? '#9ca3af' : '#6b7280' }}
            tickFormatter={formatXAxis}
            stroke={theme.mode === 'dark' ? '#4b5563' : '#d1d5db'}
          />
          <YAxis 
            tick={{ fill: theme.mode === 'dark' ? '#9ca3af' : '#6b7280' }}
            tickFormatter={formatNumber}
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
            labelFormatter={(label) => isTimeAxis ? new Date(label).toLocaleString() : label}
            contentStyle={{ 
              backgroundColor: theme.mode === 'dark' ? '#1f2937' : '#ffffff',
              borderColor: theme.mode === 'dark' ? '#374151' : '#e5e7eb',
              color: theme.mode === 'dark' ? '#f9fafb' : '#111827'
            }}
          />
          <Legend
            wrapperStyle={{ 
              paddingTop: '10px',
              color: theme.mode === 'dark' ? '#f9fafb' : '#111827' 
            }}
          />
          {dataKeys.map((dataKey, index) => (
            <Line
              key={dataKey}
              type="monotone"
              dataKey={dataKey}
              stroke={lineColors[index % lineColors.length]}
              activeDot={{ r: 5, onClick: onPointClick }}
              dot={false}
              name={dataKey}
            />
          ))}
        </RechartsLineChart>
      </ResponsiveContainer>
    </div>
  );
};

export default LineChart;