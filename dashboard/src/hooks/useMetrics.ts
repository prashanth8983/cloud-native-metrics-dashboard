import { useState, useEffect, useCallback } from 'react';
import metricsService from '../services/metricsService';
import { Metric, TopMetric, MetricHealth } from '../types/metrics';

export function useMetrics() {
  const [metrics, setMetrics] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const fetchMetrics = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await metricsService.getMetrics();
      setMetrics(data);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch metrics'));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchMetrics();
  }, [fetchMetrics]);

  const getMetricSummary = useCallback(async (name: string): Promise<Metric | null> => {
    try {
      return await metricsService.getMetricSummary(name);
    } catch (err) {
      setError(err instanceof Error ? err : new Error(`Failed to fetch metric summary for ${name}`));
      return null;
    }
  }, []);

  const getTopMetrics = useCallback(async (limit: number = 10): Promise<TopMetric[]> => {
    try {
      return await metricsService.getTopMetrics(limit);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch top metrics'));
      return [];
    }
  }, []);

  const getMetricHealth = useCallback(async (name: string): Promise<MetricHealth | null> => {
    try {
      return await metricsService.getMetricHealth(name);
    } catch (err) {
      setError(err instanceof Error ? err : new Error(`Failed to fetch health for ${name}`));
      return null;
    }
  }, []);

  return {
    metrics,
    loading,
    error,
    fetchMetrics,
    getMetricSummary,
    getTopMetrics,
    getMetricHealth
  };
}
