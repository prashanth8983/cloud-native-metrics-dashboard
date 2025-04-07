import React, { useMemo } from 'react';
import LineChart from './LineChart';
import { TimeRange } from '../../types/common';
import { RangeQueryResponse, TimeSeries } from '../../types/queries';

interface TimeSeriesChartProps {
  data: RangeQueryResponse;
  height?: number | string;
  title?: string;
  loading?: boolean;
  yAxisLabel?: string;
  colors?: string[];
  timeRange: TimeRange;
  onPointClick?: (point: any) => void;
  limit?: number;
  filter?: (series: TimeSeries) => boolean;
}

const TimeSeriesChart: React.FC<TimeSeriesChartProps> = ({
  data,
  height = 300,
  title,
  loading = false,
  yAxisLabel,
  colors,
  timeRange,
  onPointClick,
  limit = 5,
  filter
}) => {
  // Transform the data for the LineChart component
  const { chartData, seriesNames } = useMemo(() => {
    if (!data || !data.series || !data.series.length) {
      return { chartData: [], seriesNames: [] };
    }
    
    // Filter series if filter function is provided
    let filteredSeries = data.series;
    if (filter) {
      filteredSeries = data.series.filter(filter);
    }
    
    // Limit the number of series if needed
    const limitedSeries = filteredSeries.slice(0, limit);

    // Get unique time points from all series
    const timePoints = new Set<string>();
    limitedSeries.forEach(series => {
      series.dataPoints.forEach(point => {
        timePoints.add(point.timestamp);
      });
    });

    // Sort time points
    const sortedTimePoints = Array.from(timePoints).sort();

    // Create a map of time to values for each series
    const seriesValues = new Map<string, Map<string, number>>();
    limitedSeries.forEach(series => {
      const seriesName = getSeriesName(series);
      const valueMap = new Map<string, number>();
      
      series.dataPoints.forEach(point => {
        valueMap.set(point.timestamp, point.value);
      });
      
      seriesValues.set(seriesName, valueMap);
    });

    // Create chart data
    const chartData = sortedTimePoints.map(timestamp => {
      const point: Record<string, any> = { timestamp };
      
      seriesValues.forEach((valueMap, seriesName) => {
        point[seriesName] = valueMap.get(timestamp) || null;
      });
      
      return point;
    });

    return {
      chartData,
      seriesNames: Array.from(seriesValues.keys())
    };
  }, [data, filter, limit]);

  // Helper to generate a name for each series
  function getSeriesName(series: TimeSeries): string {
    // Use metric name
    let name = series.metricName;
    
    // Add some distinguishing labels if available
    const keyLabels = ['instance', 'job', 'name', 'id', 'method', 'status'];
    const labelParts: string[] = [];
    
    for (const key of keyLabels) {
      if (series.labels[key]) {
        labelParts.push(`${key}="${series.labels[key]}"`);
      }
    }
    
    if (labelParts.length > 0) {
      name += `{${labelParts.join(',')}}`;
    }
    
    return name;
  }

  if (loading || !data) {
    return (
      <LineChart
        data={[]}
        dataKeys={[]}
        xAxisDataKey="timestamp"
        loading={true}
        title={title}
        height={height}
      />
    );
  }

  return (
    <LineChart
      data={chartData}
      dataKeys={seriesNames}
      xAxisDataKey="timestamp"
      timeRange={timeRange}
      height={height}
      title={title}
      loading={loading}
      yAxisLabel={yAxisLabel}
      colors={colors}
      onPointClick={onPointClick}
    />
  );
};

export default TimeSeriesChart;