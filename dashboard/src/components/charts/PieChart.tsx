import React from 'react';
import {
  PieChart as RechartsPieChart,
  Pie,
  Cell,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';
import { formatNumber } from '../../utils/formatters';
import { useTheme } from '../../context/ThemeContext';
import { chartColors } from '../../utils/colors';

interface PieChartProps {
  data: Array<{
    name: string;
    value: number;
    color?: string;
  }>;
  height?: number | string;
  title?: string;
  loading?: boolean;
  colors?: string[];
  donut?: boolean;
  showLegend?: boolean;
  showLabels?: boolean;
}

const PieChart: React.FC<PieChartProps> = ({
  data,
  height = 300,
  title,
  loading = false,
  colors = chartColors,
  donut = false,
  showLegend = true,
  showLabels = false
}) => {
  const { theme } = useTheme();

  const RADIAN = Math.PI / 180;
  const renderCustomizedLabel = ({
    cx,
    cy,
    midAngle,
    innerRadius,
    outerRadius,
    percent,
   
  }: any) => {
    const radius = innerRadius + (outerRadius - innerRadius) * 0.5;
    const x = cx + radius * Math.cos(-midAngle * RADIAN);
    const y = cy + radius * Math.sin(-midAngle * RADIAN);

    return (
      <text
        x={x}
        y={y}
        fill="white"
        textAnchor={x > cx ? 'start' : 'end'}
        dominantBaseline="central"
        fontSize={12}
      >
        {`${(percent * 100).toFixed(0)}%`}
      </text>
    );
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
        <RechartsPieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            labelLine={showLabels}
            label={showLabels ? renderCustomizedLabel : undefined}
            outerRadius={donut ? 80 : 100}
            innerRadius={donut ? 60 : 0}
            fill="#8884d8"
            dataKey="value"
          >
            {data.map((entry, index) => (
              <Cell 
                key={`cell-${index}`} 
                fill={entry.color || colors[index % colors.length]} 
              />
            ))}
          </Pie>
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
        </RechartsPieChart>
      </ResponsiveContainer>
    </div>
  );
};

export default PieChart;