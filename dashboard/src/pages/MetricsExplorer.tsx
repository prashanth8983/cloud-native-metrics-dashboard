import React, { useState, useEffect } from 'react';
import { Search, RefreshCw, Filter, AlertCircle, Info } from 'lucide-react';
import { DashboardLayout } from '../components/layout';
import { TimeSeriesChart } from '../components/charts';
import { Button,  Tabs, Loader } from '../components/ui';
import { useTimeRangeContext } from '../context/TimeRangeContext';
import { useQueries } from '../hooks/useQueries';
import { useMetrics } from '../hooks/useMetrics';
import { Metric, MetricHealth } from '../types/metrics';
import { getStepSize } from '../utils/time';
import { formatNumber, formatTimestamp } from '../utils/formatters';

const MetricsExplorer: React.FC = () => {
  const { timeRange } = useTimeRangeContext();
  const { executeRangeQuery, getQuerySuggestions } = useQueries();
  const {  getMetricSummary, getMetricHealth } = useMetrics();
  
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedMetric, setSelectedMetric] = useState<string>('');
  const [metricSummary, setMetricSummary] = useState<Metric | null>(null);
  const [metricHealth, setMetricHealth] = useState<MetricHealth | null>(null);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [customQuery, setCustomQuery] = useState('');
  const [chartData, setChartData] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [_, setActiveTab] = useState('visualization');
  
  // Fetch suggestions when search term changes
  useEffect(() => {
    if (searchTerm.length >= 2) {
      const fetchSuggestions = async () => {
        try {
          const results = await getQuerySuggestions(searchTerm);
          setSuggestions(results);
        } catch (error) {
          console.error('Error fetching suggestions:', error);
        }
      };
      
      fetchSuggestions();
    } else {
      setSuggestions([]);
    }
  }, [searchTerm, getQuerySuggestions]);
  
  // Fetch metric data when selection changes
  useEffect(() => {
    if (selectedMetric) {
      fetchMetricData(selectedMetric);
    }
  }, [selectedMetric, timeRange]);
  
  const fetchMetricData = async (metricName: string) => {
    setLoading(true);
    try {
      // Fetch metric summary
      const summary = await getMetricSummary(metricName);
      setMetricSummary(summary);
      
      // Fetch metric health
      const health = await getMetricHealth(metricName);
      setMetricHealth(health);
      
      // Generate query based on the metric
      const query = customQuery || metricName;
      
      // Calculate appropriate step size
      const step = getStepSize(timeRange.start, timeRange.end);
      
      // Execute range query
      const result = await executeRangeQuery(query, timeRange.start, timeRange.end, step);
      setChartData(result);
      
    } catch (error) {
      console.error('Error fetching metric data:', error);
    } finally {
      setLoading(false);
    }
  };
  
  const handleMetricSelect = (metric: string) => {
    setSelectedMetric(metric);
    setCustomQuery('');
    setSearchTerm('');
    setSuggestions([]);
  };
  
  const handleCustomQueryChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setCustomQuery(e.target.value);
  };
  
  const handleSearchKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && customQuery) {
      fetchMetricData(selectedMetric || 'custom');
    }
  };
  
  const handleRunQuery = () => {
    fetchMetricData(selectedMetric || 'custom');
  };
  
  const renderMetricDetails = () => {
    if (!metricSummary) {
      return (
        <div className="p-4 bg-white dark:bg-gray-800 rounded-lg shadow-sm">
          <p className="text-gray-500 dark:text-gray-400">
            Select a metric to view details
          </p>
        </div>
      );
    }
    
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h3 className="text-lg font-medium text-gray-900 dark:text-white">
            {metricSummary.name}
          </h3>
        </div>
        
        <div className="px-6 py-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Cardinality
              </p>
              <p className="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">
                {formatNumber(metricSummary.cardinality)}
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Last Updated
              </p>
              <p className="mt-1 text-sm text-gray-900 dark:text-white">
                {formatTimestamp(metricSummary.lastUpdated)}
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Health Status
              </p>
              <div className="mt-1 flex items-center">
                {metricHealth ? (
                  <>
                    <span className={`inline-block h-3 w-3 rounded-full mr-2 ${
                      metricHealth.isStale || metricHealth.hasGaps
                        ? 'bg-yellow-500'
                        : 'bg-green-500'
                    }`}></span>
                    <span className="text-sm text-gray-900 dark:text-white">
                      {metricHealth.isStale || metricHealth.hasGaps ? 'Warning' : 'Healthy'}
                    </span>
                  </>
                ) : (
                  <span className="text-sm text-gray-500 dark:text-gray-400">Unknown</span>
                )}
              </div>
            </div>
          </div>
          
          <div className="mb-6">
            <h4 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">
              Labels
            </h4>
            <div className="flex flex-wrap gap-2">
              {metricSummary.labels.map(label => (
                <span 
                  key={label}
                  className="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-xs text-gray-800 dark:text-gray-200"
                >
                  {label}
                </span>
              ))}
            </div>
          </div>
          
          <div className="mb-6">
            <h4 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">
              Statistics
            </h4>
            <div className="grid grid-cols-3 gap-4">
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400">Min</p>
                <p className="font-medium text-gray-900 dark:text-white">
                  {formatNumber(metricSummary.stats.min)}
                </p>
              </div>
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400">Max</p>
                <p className="font-medium text-gray-900 dark:text-white">
                  {formatNumber(metricSummary.stats.max)}
                </p>
              </div>
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400">Avg</p>
                <p className="font-medium text-gray-900 dark:text-white">
                  {formatNumber(metricSummary.stats.avg)}
                </p>
              </div>
            </div>
          </div>
          
          {metricHealth && (
            <div>
              <h4 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">
                Health Diagnostics
              </h4>
              <div className="space-y-2">
                <div className="flex items-center">
                  <span className={`inline-block h-2 w-2 rounded-full mr-2 ${
                    metricHealth.isStale ? 'bg-yellow-500' : 'bg-green-500'
                  }`}></span>
                  <span className="text-sm text-gray-800 dark:text-gray-200">
                    {metricHealth.isStale 
                      ? 'Metric is stale (not recently updated)' 
                      : 'Metric is fresh (recently updated)'}
                  </span>
                </div>
                <div className="flex items-center">
                  <span className={`inline-block h-2 w-2 rounded-full mr-2 ${
                    metricHealth.hasGaps ? 'bg-yellow-500' : 'bg-green-500'
                  }`}></span>
                  <span className="text-sm text-gray-800 dark:text-gray-200">
                    {metricHealth.hasGaps 
                      ? 'Metric has gaps in data collection' 
                      : 'Metric has consistent data collection'}
                  </span>
                </div>
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
                  Last scraped: {formatTimestamp(metricHealth.lastScraped)}
                </p>
              </div>
            </div>
          )}
        </div>
      </div>
    );
  };
  
  const renderSamples = () => {
    if (!metricSummary || !metricSummary.samples || metricSummary.samples.length === 0) {
      return (
        <div className="p-4 bg-white dark:bg-gray-800 rounded-lg shadow-sm">
          <p className="text-gray-500 dark:text-gray-400">
            No samples available
          </p>
        </div>
      );
    }
    
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h3 className="text-lg font-medium text-gray-900 dark:text-white">
            Sample Values
          </h3>
        </div>
        
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead className="bg-gray-50 dark:bg-gray-900">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Value
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Timestamp
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Labels
                </th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              {metricSummary.samples.map((sample, index) => (
                <tr key={index}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                    {formatNumber(sample.value)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                    {formatTimestamp(sample.timestamp)}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                    <div className="flex flex-wrap gap-1">
                      {Object.entries(sample.labels).map(([key, value]) => (
                        <span 
                          key={key}
                          className="px-2 py-0.5 bg-gray-100 dark:bg-gray-700 rounded text-xs text-gray-800 dark:text-gray-200"
                        >
                          {key}={value}
                        </span>
                      ))}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  };
  
  return (
    <DashboardLayout>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Metrics Explorer</h1>
        
        <Button
          variant="outline"
          size="sm"
          leftIcon={<RefreshCw size={14} />}
          onClick={() => selectedMetric && fetchMetricData(selectedMetric)}
          disabled={!selectedMetric && !customQuery}
        >
          Refresh
        </Button>
      </div>
      
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left sidebar */}
        <div className="lg:col-span-1 space-y-6">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4">
            <h2 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
              Metrics
            </h2>
            
            <div className="relative mb-4">
              <input
                type="text"
                className="w-full px-4 py-2 pl-10 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200"
                placeholder="Search metrics..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            </div>
            
            {/* Suggestions dropdown */}
            {suggestions.length > 0 && (
              <div className="mb-4 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md shadow-sm max-h-60 overflow-y-auto">
                {suggestions.map((suggestion) => (
                  <button
                    key={suggestion}
                    className="w-full text-left px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-600 text-sm text-gray-700 dark:text-gray-200"
                    onClick={() => handleMetricSelect(suggestion)}
                  >
                    {suggestion}
                  </button>
                ))}
              </div>
            )}
            
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Custom Query
              </label>
              <div className="flex">
                <input
                  type="text"
                  className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-l-md bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200"
                  placeholder="Enter PromQL query..."
                  value={customQuery}
                  onChange={handleCustomQueryChange}
                  onKeyDown={handleSearchKeyDown}
                />
                <Button
                  onClick={handleRunQuery}
                  className="rounded-l-none"
                >
                  Run
                </Button>
              </div>
            </div>
            
            {selectedMetric && (
              <div className="flex items-center justify-between">
                <div>
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                    Selected Metric:
                  </span>
                  <span className="ml-2 text-sm text-blue-600 dark:text-blue-400">
                    {selectedMetric}
                  </span>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setSelectedMetric('')}
                >
                  Clear
                </Button>
              </div>
            )}
          </div>
          
          {renderMetricDetails()}
        </div>
        
        {/* Main content */}
        <div className="lg:col-span-2">
          <Tabs
            tabs={[
              {
                id: 'visualization',
                label: 'Visualization',
                content: (
                  <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4">
                    {loading ? (
                      <div className="flex justify-center items-center h-64">
                        <Loader size="lg" text="Loading chart data..." />
                      </div>
                    ) : chartData ? (
                      <TimeSeriesChart
                        data={chartData}
                        timeRange={timeRange}
                        height={400}
                        title={customQuery || selectedMetric}
                      />
                    ) : (
                      <div className="flex flex-col items-center justify-center h-64 text-gray-500 dark:text-gray-400">
                        <Info size={48} className="mb-4 opacity-50" />
                        <p className="text-lg">Select a metric or enter a custom query</p>
                        <p className="text-sm mt-2">
                          Visualization will appear here
                        </p>
                      </div>
                    )}
                  </div>
                ),
                icon: <Filter size={16} />
              },
              {
                id: 'samples',
                label: 'Samples',
                content: renderSamples(),
                icon: <AlertCircle size={16} />
              }
            ]}
            defaultTab="visualization"
            onChange={setActiveTab}
            className="h-full"
          />
        </div>
      </div>
    </DashboardLayout>
  );
};

export default MetricsExplorer;