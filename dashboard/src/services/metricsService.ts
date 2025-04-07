import api from './api';
import { Metric, MetricHealth, TopMetric } from '../types/metrics';

/**
 * Service for interacting with the metrics API
 */
export const metricsService = {
  /**
   * Get a list of all available metrics
   */
  async getMetrics(): Promise<string[]> {
    const response = await api.get('/metrics');
    return response.data.metrics;
  },
  
  /**
   * Get top metrics by cardinality or activity
   * @param limit Maximum number of metrics to return
   */
  async getTopMetrics(limit: number = 10): Promise<TopMetric[]> {
    const response = await api.get('/metrics/top', {
      params: { limit }
    });
    return response.data.metrics;
  },
  
  /**
   * Get detailed information about a specific metric
   * @param name The name of the metric
   */
  async getMetricSummary(name: string): Promise<Metric> {
    const response = await api.get(`/metrics/${name}`);
    return response.data;
  },
  
  /**
   * Get health information about a specific metric
   * @param name The name of the metric
   */
  async getMetricHealth(name: string): Promise<MetricHealth> {
    const response = await api.get(`/metrics/${name}/health`);
    return response.data;
  },
  
  /**
   * Get metrics that match a search query
   * @param query The search term
   * @param limit Maximum number of results
   */
  async searchMetrics(query: string, limit: number = 10): Promise<string[]> {
    const allMetrics = await this.getMetrics();
    
    if (!query) {
      return allMetrics.slice(0, limit);
    }
    
    // Filter metrics by name
    const filteredMetrics = allMetrics.filter(metric => 
      metric.toLowerCase().includes(query.toLowerCase())
    );
    
    return filteredMetrics.slice(0, limit);
  },
  
  /**
   * Get a dashboard-friendly summary of multiple metrics
   * @param metricNames Array of metric names to include
   */
  async getDashboardMetrics(metricNames: string[]): Promise<Metric[]> {
    // Use Promise.all to fetch multiple metrics in parallel
    const metricsPromises = metricNames.map(name => this.getMetricSummary(name));
    return Promise.all(metricsPromises);
  },
  
  /**
   * Get health status for multiple metrics
   * @param metricNames Array of metric names to check
   */
  async getMetricsHealth(metricNames: string[]): Promise<Record<string, MetricHealth>> {
    const healthPromises = metricNames.map(async name => {
      const health = await this.getMetricHealth(name);
      return [name, health];
    });
    
    const healthEntries = await Promise.all(healthPromises);
    return Object.fromEntries(healthEntries);
  }
};

export default metricsService;