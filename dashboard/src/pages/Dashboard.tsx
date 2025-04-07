import React, { useState, useEffect } from 'react';
import { 
  BarChartIcon, 
  ClockIcon, 
  ServerCrashIcon, 
  ActivityIcon,
  RefreshCw
} from 'lucide-react';

import { MetricCard, StatCard, AlertCard } from '../components/cards';
import { LineChart, BarChart, PieChart } from '../components/charts';
import { MetricsTable, AlertsTable } from '../components/tables';
import { Button, Tabs } from '../components/ui';
import { useTimeRangeContext } from '../context/TimeRangeContext';
import { useQueries } from '../hooks/useQueries';
import { useAlerts } from '../hooks/useAlerts';
import { useMetrics } from '../hooks/useMetrics';
import { getStepSize } from '../utils/time';
import { getSeverityColor } from '../utils/colors';
import { DashboardLayout } from '../components/layout';


const Dashboard: React.FC = () => {
  const { timeRange } = useTimeRangeContext();
  const { executeRangeQuery } = useQueries();
  const { alerts, summary: alertSummary, loading: alertsLoading } = useAlerts();
  const { getTopMetrics } = useMetrics();
  
  const [cpuData, setCpuData] = useState<any[]>([]);
  const [memoryData, setMemoryData] = useState<any[]>([]);
  const [topMetrics, setTopMetrics] = useState<any[]>([]);
  const [requestsData, setRequestsData] = useState<any[]>([]);
  const [latencyData, setLatencyData] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const fetchDashboardData = async () => {
    setRefreshing(true);
    try {
      // Calculate step size based on time range
      const step = getStepSize(timeRange.start, timeRange.end);
      
      // Fetch CPU data
      const cpuResult = await executeRangeQuery(
        'avg by(instance) (rate(node_cpu_seconds_total{mode="user"}[5m]) * 100)',
        timeRange.start,
        timeRange.end,
        step
      );
      
      if (cpuResult && cpuResult.series.length > 0) {
        // Transform data for chart
        const transformedData = cpuResult.series[0].dataPoints.map(point => ({
          timestamp: point.timestamp,
          cpu: parseFloat(point.value.toFixed(2))
        }));
        setCpuData(transformedData);
      }
      
      // Fetch Memory data
      const memoryResult = await executeRangeQuery(
        'node_memory_MemFree_bytes / node_memory_MemTotal_bytes * 100',
        timeRange.start,
        timeRange.end,
        step
      );
      
      if (memoryResult && memoryResult.series.length > 0) {
        // Transform data for chart
        const transformedData = memoryResult.series[0].dataPoints.map(point => ({
          timestamp: point.timestamp,
          memory: 100 - parseFloat(point.value.toFixed(2)) // Convert to used memory percentage
        }));
        setMemoryData(transformedData);
      }
      
      // Fetch HTTP request rate
      const requestsResult = await executeRangeQuery(
        'sum(rate(http_requests_total[5m])) by (code)',
        timeRange.start,
        timeRange.end,
        step
      );
      
      if (requestsResult && requestsResult.series.length > 0) {
        // Create a map of timestamp to data point
        const dataByTimestamp = new Map();
        
        // Process each series (one per status code)
        requestsResult.series.forEach(series => {
          const code = series.labels.code || 'unknown';
          
          series.dataPoints.forEach(point => {
            const timestamp = point.timestamp;
            if (!dataByTimestamp.has(timestamp)) {
              dataByTimestamp.set(timestamp, { timestamp });
            }
            
            const dataPoint = dataByTimestamp.get(timestamp);
            dataPoint[`${code}`] = parseFloat(point.value.toFixed(2));
          });
        });
        
        setRequestsData(Array.from(dataByTimestamp.values()));
      }
      
      // Fetch latency data
      const latencyResult = await executeRangeQuery(
        'histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))',
        timeRange.start,
        timeRange.end,
        step
      );
      
      if (latencyResult && latencyResult.series.length > 0) {
        // Transform data for chart
        const transformedData = latencyResult.series[0].dataPoints.map(point => ({
          timestamp: point.timestamp,
          latency: parseFloat((point.value * 1000).toFixed(2)) // Convert seconds to ms
        }));
        setLatencyData(transformedData);
      }
      
      // Fetch top metrics
      const metrics = await getTopMetrics(10);
      setTopMetrics(metrics);
      
    } catch (error) {
      console.error('Error fetching dashboard data:', error);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    let isMounted = true;
  
    const fetch = async () => {
      await fetchDashboardData();
    };
  
    fetch();
  
    const intervalId = setInterval(() => {
      if (isMounted) fetchDashboardData();
    }, 60000);
  
    return () => {
      isMounted = false;
      clearInterval(intervalId);
    };
  }, [timeRange]);
  

  const handleRefresh = () => {
    fetchDashboardData();
  };

  // Create alert summary data for pie chart
  const alertSummaryData = alertSummary ? [
    { name: 'Firing', value: alertSummary.firingCount, color: getSeverityColor('critical') },
    { name: 'Pending', value: alertSummary.pendingCount, color: getSeverityColor('warning') },
    { name: 'Resolved', value: alertSummary.resolvedCount, color: getSeverityColor('low') }
  ] : [];

  // Get status code counts for request distribution
  const getStatusCodeCounts = () => {
    const counts: Record<string, number> = {};
    
    if (requestsData.length > 0) {
      const latestData = requestsData[requestsData.length - 1];
      
      Object.entries(latestData).forEach(([key, value]) => {
        if (key !== 'timestamp') {
          counts[key] = value as number;
        }
      });
    }
    
    return Object.entries(counts).map(([code, count]) => {
      let color = '#4285F4'; // Default blue
      
      if (code.startsWith('2')) {
        color = '#0F9D58'; // Green for 2xx
      } else if (code.startsWith('4')) {
        color = '#F4B400'; // Yellow for 4xx
      } else if (code.startsWith('5')) {
        color = '#DB4437'; // Red for 5xx
      }
      
      return {
        name: code,
        value: count,
        color
      };
    });
  };

  return (
    <DashboardLayout>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Dashboard</h1>
        <Button
          variant="outline"
          size="sm"
          leftIcon={<RefreshCw size={14} />}
          onClick={handleRefresh}
          isLoading={refreshing}
        >
          Refresh
        </Button>
      </div>

      {/* Key metrics section */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <MetricCard
          title="CPU Usage"
          value={cpuData.length > 0 ? cpuData[cpuData.length - 1].cpu : null}
          change={1.8}
          changeType="positive"
          icon={<ServerCrashIcon size={16} />}
          loading={loading}
        />
        <MetricCard
          title="Memory Usage"
          value={memoryData.length > 0 ? memoryData[memoryData.length - 1].memory : null}
          change={-0.5}
          changeType="negative"
          icon={<ActivityIcon size={16} />}
          loading={loading}
        />
        <MetricCard
          title="Requests/sec"
          value={
            requestsData.length > 0 
              ? Object.entries(requestsData[requestsData.length - 1])
                  .filter(([key]) => key !== 'timestamp')
                  .reduce((sum, [_, value]) => sum + (value as number), 0)
              : null
          }
          change={2.4}
          changeType="positive"
          icon={<BarChartIcon size={16} />}
          loading={loading}
        />
        <MetricCard
          title="Avg. Latency"
          value={latencyData.length > 0 ? latencyData[latencyData.length - 1].latency : null}
          change={0.3}
          changeType="neutral"
          icon={<ClockIcon size={16} />}
          loading={loading}
        />
      </div>

      {/* Alerts summary */}
      {alertSummary && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <StatCard
            title="Firing Alerts"
            value={alertSummary.firingCount}
            subtitle="Critical issues requiring attention"
            accentColor="red"
            onClick={() => {}}
          />
          <StatCard
            title="Pending Alerts"
            value={alertSummary.pendingCount}
            subtitle="Issues that may require attention soon"
            accentColor="yellow"
            onClick={() => {}}
          />
          <StatCard
            title="Total Alerts"
            value={alertSummary.totalCount}
            subtitle="All active and resolved alerts"
            accentColor="blue"
            onClick={() => {}}
          />
        </div>
      )}

      {/* Charts section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <LineChart
          data={cpuData}
          dataKeys={['cpu']}
          xAxisDataKey="timestamp"
          timeRange={timeRange}
          title="CPU Usage (%)"
          loading={loading}
          yAxisLabel="Usage %"
        />
        <LineChart
          data={memoryData}
          dataKeys={['memory']}
          xAxisDataKey="timestamp"
          timeRange={timeRange}
          title="Memory Usage (%)"
          loading={loading}
          yAxisLabel="Usage %"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <LineChart
          data={latencyData}
          dataKeys={['latency']}
          xAxisDataKey="timestamp"
          timeRange={timeRange}
          title="Request Latency (ms)"
          loading={loading}
          yAxisLabel="Latency (ms)"
        />
        <BarChart
          data={requestsData}
          dataKeys={['200', '201', '400', '403', '404', '500', '503'].filter(code => 
            requestsData.length > 0 && requestsData[0][code] !== undefined
          )}
          xAxisDataKey="timestamp"
          timeRange={timeRange}
          title="HTTP Requests by Status Code"
          loading={loading}
          yAxisLabel="Requests/sec"
          stacked={true}
        />
      </div>

      {/* Tables and additional widgets */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
        <div className="lg:col-span-2">
          <Tabs
            tabs={[
              {
                id: 'top-metrics',
                label: 'Top Metrics',
                content: <MetricsTable metrics={topMetrics} loading={loading} />
              },
              {
                id: 'alerts',
                label: 'Alerts',
                content: <AlertsTable alerts={alerts.slice(0, 10)} loading={alertsLoading} />
              }
            ]}
          />
        </div>
        <div className="space-y-6">
          <PieChart
            data={getStatusCodeCounts()}
            title="HTTP Status Distribution"
            loading={loading}
            donut={true}
          />
          
          <PieChart
            data={alertSummaryData}
            title="Alert Status"
            loading={alertsLoading}
            donut={true}
          />
          
          {alerts.length > 0 && (
            <div>
              <h3 className="text-md font-medium text-gray-800 dark:text-white mb-4">Critical Alerts</h3>
              <div className="space-y-4">
                {alerts
                  .filter(alert => alert.severity === 'critical' && alert.state === 'firing')
                  .slice(0, 3)
                  .map(alert => (
                    <AlertCard key={alert.name + alert.activeAt} alert={alert} />
                  ))}
                {alerts.filter(alert => alert.severity === 'critical' && alert.state === 'firing').length === 0 && (
                  <p className="text-sm text-gray-500 dark:text-gray-400">No critical alerts</p>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </DashboardLayout>
  );
};

export default Dashboard;